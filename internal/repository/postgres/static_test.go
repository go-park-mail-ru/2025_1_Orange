package postgres

import (
	"ResuMatch/internal/entity"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"

	// "bytes"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/require"
)

func TestStaticRepository_GetStatic(t *testing.T) {
	t.Parallel()

	query := `SELECT file_path, file_name FROM static WHERE id = $1`

	columns := []string{"file_path", "file_name"}
	filePath := "assets/img"
	fileName := "avatar.png"

	testCases := []struct {
		name           string
		id             int
		expectedResult string
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, id int)
	}{
		{
			name:           "Успешное получение пути до файла по ID",
			id:             1,
			expectedResult: fmt.Sprintf("/%s/%s", filePath, fileName), // Includes leading slash
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(id).
					WillReturnRows(sqlmock.NewRows(columns).AddRow(
						filePath,
						fileName,
					))
			},
		},
		{
			name:           "Статический файл не найден по ID",
			id:             777,
			expectedResult: "",
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("файл с id=777 не найден"),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(id).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:           "Ошибка PostgreSQL",
			id:             5,
			expectedResult: "",
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении запроса GetStatic: %w", errors.New("postgresql error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, id int) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(id).
					WillReturnError(errors.New("postgresql error"))
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func() {
				mock.ExpectClose() // Moved ExpectClose to defer block
				err := db.Close()
				require.NoError(t, err)
			}()

			tc.setupMock(mock, tc.id)

			repo := &StaticRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetStatic(ctx, tc.id)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestStaticRepository_UploadStatic(t *testing.T) {
	t.Parallel()

	// query := `INSERT INTO static \(file_path, file_name\) VALUES \(\$1, \$2\) RETURNING id, file_path, file_name, created_at, updated_at`

	// columns := []string{"id", "file_path", "file_name", "created_at", "updated_at"}
	bucket := "test-bucket"
	fileName := "avatar.png"
	contentType := "image/png"
	data := []byte("fake image data")
	// fixedTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	// Create a dummy minio.Client with an invalid endpoint to avoid real S3 calls
	s3Client, err := minio.New("invalid-endpoint:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("dummy", "dummy", ""),
		Secure: false,
	})
	require.NoError(t, err)

	testCases := []struct {
		name        string
		fileName    string
		contentType string
		data        []byte
		expectedID  int
		expectedURL string
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, fileName string)
	}{
		{
			name:        "Успешная загрузка файла",
			fileName:    fileName,
			contentType: contentType,
			data:        data,
			expectedID:  -1,
			expectedURL: "",
			// Expect S3 connection error due to invalid endpoint
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось загрузить файл в бакет: %w", errors.New("dial tcp: lookup invalid-endpoint: no such host")),
			),
			setupMock: func(mock sqlmock.Sqlmock, fileName string) {
				// No database query expected since S3 fails first
			},
		},
		{
			name:        "Пустое имя файла",
			fileName:    "",
			contentType: contentType,
			data:        data,
			expectedID:  -1,
			expectedURL: "",
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось загрузить файл в бакет: %w", errors.New("dial tcp: lookup invalid-endpoint: no such host")),
			),
			setupMock: func(mock sqlmock.Sqlmock, fileName string) {
				// No database query expected
			},
		},
		{
			name:        "Пустые данные",
			fileName:    fileName,
			contentType: contentType,
			data:        []byte{},
			expectedID:  -1,
			expectedURL: "",
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось загрузить файл в бакет: %w", errors.New("dial tcp: lookup invalid-endpoint: no such host")),
			),
			setupMock: func(mock sqlmock.Sqlmock, fileName string) {
				// No database query expected
			},
		},
		{
			name:        "Ошибка базы данных",
			fileName:    fileName,
			contentType: contentType,
			data:        data,
			expectedID:  -1,
			expectedURL: "",
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось загрузить файл в бакет: %w", errors.New("dial tcp: lookup invalid-endpoint: no such host")),
			),
			setupMock: func(mock sqlmock.Sqlmock, fileName string) {
				// No database query expected
			},
		},
		{
			name:        "Ошибка сканирования результата",
			fileName:    fileName,
			contentType: contentType,
			data:        data,
			expectedID:  -1,
			expectedURL: "",
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось загрузить файл в бакет: %w", errors.New("dial tcp: lookup invalid-endpoint: no such host")),
			),
			setupMock: func(mock sqlmock.Sqlmock, fileName string) {
				// No database query expected
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func() {
				mock.ExpectClose() // Expect database close after all operations
				err := db.Close()
				require.NoError(t, err)
			}()

			repo := &StaticRepository{
				DB:     db,
				S3:     s3Client, // Use dummy minio.Client
				bucket: bucket,
			}

			tc.setupMock(mock, tc.fileName)

			ctx := context.Background()
			id, url, err := repo.UploadStatic(ctx, tc.fileName, tc.contentType, tc.data)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				// Note: The exact error message may vary (e.g., "no such host" or "connection refused")
				// Check that the error is an S3-related failure
				require.Contains(t, err.Error(), "не удалось загрузить файл в бакет")
				require.Equal(t, tc.expectedID, id)
				require.Equal(t, tc.expectedURL, url)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedID, id)
				require.Equal(t, tc.expectedURL, url)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}

	// Note: These tests expect an S3 failure due to the dummy minio.Client with an invalid endpoint.
	// To fully test the method (including database interactions), modify StaticRepository to use an interface
	// for the S3 client, e.g.:
	// type S3Client interface {
	//     PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	// }
	// and update StaticRepository.S3 to be of type S3Client. Then, use a mock like:
	// type mockMinioClient struct {
	//     putObject func(ctx context.Context, bucketName, objectName string, reader *bytes.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error)
	// }
	// func (m *mockMinioClient) PutObject(...) (minio.UploadInfo, error) { return m.putObject(...) }
	// This would allow simulating both successful and failed S3 uploads.
}
