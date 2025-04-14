package postgres

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/utils"
	l "ResuMatch/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"
<<<<<<< HEAD
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
=======
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
>>>>>>> a6396a4 (Fix mistakes)
)

type StaticRepository struct {
	DB *sql.DB
}

func NewStaticRepository(db *sql.DB) (repository.StaticRepository, error) {
	return &StaticRepository{DB: db}, nil
}

func (r *StaticRepository) UploadStatic(ctx context.Context, filePath, fileName string, data []byte) (*entity.Static, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"filePath":  filePath,
		"fileName":  fileName,
	}).Info("выполнение sql-запроса сохранения статики UploadStatic")

	dir := filepath.Dir(fmt.Sprintf("/app/%s", filePath))
	err := os.MkdirAll(dir, 0755)

	if err != nil {
		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("не удалось создать файл: %w", err),
		)
	}

	dest, err := os.Create(filepath.Join(dir, fileName))
	if err != nil {
		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("внутренная ошибка при создании файла: %w", err),
		)
	}

	if _, err := dest.Write(data); err != nil {
		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("внутренная ошибка при записи данных в файл: %w", err),
		)
	}

	if err := dest.Close(); err != nil {
		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("внутренная ошибка при закрытии файла: %w", err),
		)
	}

	var static entity.Static
	query := `
        INSERT INTO static (file_path, file_name) 
        VALUES ($1, $2) 
        RETURNING id, file_path, file_name, created_at, updated_at
    `
	err = r.DB.QueryRow(query, filePath, fileName).Scan(
		&static.ID,
		&static.FilePath,
		&static.FileName,
		&static.CreatedAt,
		&static.UpdatedAt,
	)
	if err != nil {
		return nil, entity.NewError(
			entity.ErrInternal,
			fmt.Errorf("внутренная ошибка при выполнении sql-запроса UploadStatic: %w", err),
		)
	}
	return &static, nil
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
	return fmt.Sprintf("%s/%s", filePath, fileName), nil
}
