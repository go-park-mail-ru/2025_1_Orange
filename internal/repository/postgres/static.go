package postgres

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/metrics"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/utils"
	l "ResuMatch/pkg/logger"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

type StaticRepository struct {
	S3        *minio.Client
	DB        *sql.DB
	bucket    string
	cfg       config.MinioConfig
	publicURL string
}

func NewStaticRepository(db *sql.DB, bucket string, cfg config.MinioConfig) (repository.StaticRepository, error) {
	S3, err := minio.New(cfg.InternalEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.RootUser, cfg.RootPassword, ""),
		Secure: cfg.UseSSL,
	})

	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Static Repository", "NewStaticRepository").Inc()
		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось создать minio клиента: %w", err),
		)
	}

	exists, err := S3.BucketExists(context.Background(), bucket)
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Static Repository", "NewStaticRepository").Inc()
		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось проверить существование бакета: %w", err),
		)
	}

	if !exists {
		if err := S3.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{}); err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Static Repository", "NewStaticRepository").Inc()
			return nil, entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось создать бакет: %w", err),
			)
		}
	}
	return &StaticRepository{
		DB:        db,
		S3:        S3,
		bucket:    bucket,
		cfg:       cfg,
		publicURL: fmt.Sprintf("%s://%s", cfg.Scheme, cfg.PublicEndpoint),
	}, nil
}

func (r *StaticRepository) UploadStatic(ctx context.Context, fileName string, contentType string, data []byte) (int, string, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"fileName":  fileName,
	}).Info("выполнение sql-запроса сохранения статики UploadStatic")

	_, err := r.S3.PutObject(ctx, r.bucket, fileName, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Static Repository", "UploadStatic").Inc()
		return -1, "", entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось загрузить файл в бакет: %w", err),
		)
	}

	var static entity.Static
	query := `
        INSERT INTO static (file_path, file_name) 
        VALUES ($1, $2) 
        RETURNING id, file_path, file_name, created_at, updated_at
    `
	err = r.DB.QueryRow(query, r.bucket, fileName).Scan(
		&static.ID,
		&static.FilePath,
		&static.FileName,
		&static.CreatedAt,
		&static.UpdatedAt,
	)

	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Static Repository", "UploadStatic").Inc()
		return -1, "", entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("внутренная ошибка при выполнении sql-запроса UploadStatic: %w", err),
		)
	}

	return static.ID, r.getStaticURL(static.FilePath, static.FileName), nil
}

func (r *StaticRepository) GetStatic(ctx context.Context, id int) (string, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"id":        id,
	}).Info("выполнение sql-запроса получения статики по id GetStatic")

	query := `SELECT file_path, file_name FROM static WHERE id = $1`

	var filePath, fileName string
	err := r.DB.QueryRow(query, id).Scan(&filePath, &fileName)

	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Static Repository", "GetStatic").Inc()
		if errors.Is(err, sql.ErrNoRows) {
			return "", entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("файл с id=%d не найден", id),
			)
		}

		return "", entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("ошибка при выполнении запроса GetStatic: %w", err),
		)
	}

	return r.getStaticURL(filePath, fileName), nil
}

func (r *StaticRepository) DeleteStatic(ctx context.Context, id int) error {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"id":        id,
	}).Info("выполнение sql-запроса удаления статики по id DeleteStatic")

	var bucket, fileName string
	err := r.DB.QueryRowContext(
		ctx,
		`DELETE FROM static WHERE id = $1 RETURNING file_path, file_name`,
		id,
	).Scan(&bucket, &fileName)

	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Static Repository", "DeleteStatic").Inc()
		if errors.Is(err, sql.ErrNoRows) {
			return entity.NewError(entity.ErrNotFound, fmt.Errorf("файл не найден"))
		}
		return entity.NewError(entity.ErrInternal, err)
	}

	if err := r.S3.RemoveObject(ctx, bucket, fileName, minio.RemoveObjectOptions{}); err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Static Repository", "DeleteStatic").Inc()
		return entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось удалить файл из minio: %w", err),
		)
	}

	return nil
}

func (r *StaticRepository) getStaticURL(bucket, fileName string) string {
	return fmt.Sprintf("%s/%s/%s", r.publicURL, bucket, fileName)
}
