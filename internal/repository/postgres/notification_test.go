package postgres

import (
	"ResuMatch/internal/entity"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestNotificationRepository_GetApplyNotificationPreview(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
		SELECT 
			n.id,
			n.type,
			n.sender_id,
			n.receiver_id,
			n.object_id,
			n.resume_id,
			n.is_viewed,
			n.created_at,
			a.first_name AS applicant_name,
			e.company_name AS employer_name,
			v.title
		FROM notification n
		LEFT JOIN applicant a ON n.sender_id = a.id
		LEFT JOIN employer e ON n.receiver_id = e.id
		LEFT JOIN vacancy v ON n.object_id = v.id
		WHERE n.id = $1 AND n.type = 'apply'
	`)

	columns := []string{
		"id", "type", "sender_id", "receiver_id", "object_id",
		"resume_id", "is_viewed", "created_at", "applicant_name",
		"employer_name", "title",
	}

	fixedTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		name           string
		notificationID int
		expectedResult *entity.NotificationPreview
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, notificationID int)
	}{
		{
			name:           "Успешное получение превью уведомления о отклике",
			notificationID: 1,
			expectedResult: &entity.NotificationPreview{
				ID:            1,
				Type:          entity.ApplyNotificationType,
				SenderID:      100,
				ReceiverID:    200,
				ObjectID:      300,
				ResumeID:      400,
				IsViewed:      false,
				CreatedAt:     fixedTime,
				ApplicantName: "Иван",
				EmployerName:  "ООО Рога и Копыта",
				Title:         "Backend Developer",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1, "apply", 100, 200, 300, 400, false, fixedTime,
						"Иван", "ООО Рога и Копыта", "Backend Developer",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Успешное получение превью с просмотренным уведомлением",
			notificationID: 2,
			expectedResult: &entity.NotificationPreview{
				ID:            2,
				Type:          entity.ApplyNotificationType,
				SenderID:      101,
				ReceiverID:    201,
				ObjectID:      301,
				ResumeID:      401,
				IsViewed:      true,
				CreatedAt:     fixedTime,
				ApplicantName: "Петр",
				EmployerName:  "ИП Сидоров",
				Title:         "Frontend Developer",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						2, "apply", 101, 201, 301, 401, true, fixedTime,
						"Петр", "ИП Сидоров", "Frontend Developer",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Превью с пустыми значениями в JOIN полях",
			notificationID: 3,
			expectedResult: &entity.NotificationPreview{
				ID:            3,
				Type:          entity.ApplyNotificationType,
				SenderID:      102,
				ReceiverID:    202,
				ObjectID:      302,
				ResumeID:      402,
				IsViewed:      false,
				CreatedAt:     fixedTime,
				ApplicantName: "",
				EmployerName:  "",
				Title:         "",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						3, "apply", 102, 202, 302, 402, false, fixedTime,
						"", "", "",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Превью с частично заполненными JOIN полями",
			notificationID: 4,
			expectedResult: &entity.NotificationPreview{
				ID:            4,
				Type:          entity.ApplyNotificationType,
				SenderID:      103,
				ReceiverID:    203,
				ObjectID:      303,
				ResumeID:      403,
				IsViewed:      true,
				CreatedAt:     fixedTime,
				ApplicantName: "Анна",
				EmployerName:  "",
				Title:         "DevOps Engineer",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						4, "apply", 103, 203, 303, 403, true, fixedTime,
						"Анна", "", "DevOps Engineer",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Уведомление не найдено",
			notificationID: 999,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("ошибка при получении уведомления: %v", sql.ErrNoRows),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:           "Ошибка базы данных при выполнении запроса",
			notificationID: 5,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("ошибка при получении уведомления: %v", errors.New("connection lost")),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnError(errors.New("connection lost"))
			},
		},
		{
			name:           "Ошибка сканирования - неверный тип данных",
			notificationID: 6,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("ошибка при получении уведомления: %v", errors.New(`sql: Scan error on column index 0, name "id": converting driver.Value type string ("invalid_id") to a int: invalid syntax`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						"invalid_id", "apply", 100, 200, 300, 400, false, fixedTime,
						"Иван", "ООО Рога и Копыта", "Backend Developer",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Превью с нулевыми ID",
			notificationID: 7,
			expectedResult: &entity.NotificationPreview{
				ID:            7,
				Type:          entity.ApplyNotificationType,
				SenderID:      0,
				ReceiverID:    0,
				ObjectID:      0,
				ResumeID:      0,
				IsViewed:      false,
				CreatedAt:     fixedTime,
				ApplicantName: "Тест",
				EmployerName:  "Тест Компания",
				Title:         "Тест Позиция",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						7, "apply", 0, 0, 0, 0, false, fixedTime,
						"Тест", "Тест Компания", "Тест Позиция",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Превью с максимальными значениями ID",
			notificationID: 2147483647, // max int32
			expectedResult: &entity.NotificationPreview{
				ID:            2147483647,
				Type:          entity.ApplyNotificationType,
				SenderID:      2147483646,
				ReceiverID:    2147483645,
				ObjectID:      2147483644,
				ResumeID:      2147483643,
				IsViewed:      true,
				CreatedAt:     fixedTime,
				ApplicantName: "Максимальный",
				EmployerName:  "Максимальная Компания",
				Title:         "Максимальная Позиция",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						2147483647, "apply", 2147483646, 2147483645, 2147483644, 2147483643, true, fixedTime,
						"Максимальный", "Максимальная Компания", "Максимальная Позиция",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Превью с длинными строками",
			notificationID: 8,
			expectedResult: &entity.NotificationPreview{
				ID:            8,
				Type:          entity.ApplyNotificationType,
				SenderID:      104,
				ReceiverID:    204,
				ObjectID:      304,
				ResumeID:      404,
				IsViewed:      false,
				CreatedAt:     fixedTime,
				ApplicantName: "Очень Длинное Имя Соискателя Которое Может Быть В Базе Данных",
				EmployerName:  "Очень Длинное Название Компании Работодателя Которое Тоже Может Быть В Базе",
				Title:         "Очень Длинное Название Вакансии Senior Full Stack Developer with Machine Learning Experience",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						8, "apply", 104, 204, 304, 404, false, fixedTime,
						"Очень Длинное Имя Соискателя Которое Может Быть В Базе Данных",
						"Очень Длинное Название Компании Работодателя Которое Тоже Может Быть В Базе",
						"Очень Длинное Название Вакансии Senior Full Stack Developer with Machine Learning Experience",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Превью с отрицательным ID уведомления",
			notificationID: -1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("ошибка при получении уведомления: %v", sql.ErrNoRows),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:           "Ошибка сканирования NULL значений",
			notificationID: 10,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("ошибка при получении уведомления: %v", errors.New(`sql: Scan error on column index 8, name "applicant_name": converting NULL to string is unsupported`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						10, "apply", 102, 202, 302, 402, false, fixedTime,
						nil, "Компания", "Позиция",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}(db, mock)

			tc.setupMock(mock, tc.notificationID)

			repo := &NotificationRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetApplyNotificationPreview(ctx, tc.notificationID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tc.expectedResult.ID, result.ID)
				require.Equal(t, tc.expectedResult.Type, result.Type)
				require.Equal(t, tc.expectedResult.SenderID, result.SenderID)
				require.Equal(t, tc.expectedResult.ReceiverID, result.ReceiverID)
				require.Equal(t, tc.expectedResult.ObjectID, result.ObjectID)
				require.Equal(t, tc.expectedResult.ResumeID, result.ResumeID)
				require.Equal(t, tc.expectedResult.IsViewed, result.IsViewed)
				require.Equal(t, tc.expectedResult.CreatedAt.Unix(), result.CreatedAt.Unix())
				require.Equal(t, tc.expectedResult.ApplicantName, result.ApplicantName)
				require.Equal(t, tc.expectedResult.EmployerName, result.EmployerName)
				require.Equal(t, tc.expectedResult.Title, result.Title)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNotificationRepository_GetDownloadResumeNotificationPreview(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
		SELECT 
			n.id,
			n.type,
			n.sender_id,
			n.receiver_id,
			n.object_id,
			n.resume_id,
			n.is_viewed,
			n.created_at,
			a.first_name AS applicant_name,
			e.company_name AS employer_name,
			r.profession AS title
		FROM notification n
		LEFT JOIN resume r ON r.id = n.object_id
		LEFT JOIN applicant a ON a.id = r.applicant_id
		LEFT JOIN employer e ON e.id = n.sender_id
		WHERE n.id = $1 AND n.type = 'download_resume'
	`)

	columns := []string{
		"id", "type", "sender_id", "receiver_id", "object_id",
		"resume_id", "is_viewed", "created_at", "applicant_name",
		"employer_name", "title",
	}

	fixedTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		name           string
		notificationID int
		expectedResult *entity.NotificationPreview
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, notificationID int)
	}{
		{
			name:           "Успешное получение превью уведомления о скачивании резюме",
			notificationID: 1,
			expectedResult: &entity.NotificationPreview{
				ID:            1,
				Type:          entity.DownloadResumeType,
				SenderID:      100,
				ReceiverID:    200,
				ObjectID:      300,
				ResumeID:      400,
				IsViewed:      false,
				CreatedAt:     fixedTime,
				ApplicantName: "Анна",
				EmployerName:  "ООО ТехКомпани",
				Title:         "Frontend Developer",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1, "download_resume", 100, 200, 300, 400, false, fixedTime,
						"Анна", "ООО ТехКомпани", "Frontend Developer",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Успешное получение превью с просмотренным уведомлением",
			notificationID: 2,
			expectedResult: &entity.NotificationPreview{
				ID:            2,
				Type:          entity.DownloadResumeType,
				SenderID:      101,
				ReceiverID:    201,
				ObjectID:      301,
				ResumeID:      401,
				IsViewed:      true,
				CreatedAt:     fixedTime,
				ApplicantName: "Михаил",
				EmployerName:  "Стартап Инновации",
				Title:         "Data Scientist",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						2, "download_resume", 101, 201, 301, 401, true, fixedTime,
						"Михаил", "Стартап Инновации", "Data Scientist",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Превью с пустыми значениями в JOIN полях",
			notificationID: 3,
			expectedResult: &entity.NotificationPreview{
				ID:            3,
				Type:          entity.DownloadResumeType,
				SenderID:      102,
				ReceiverID:    202,
				ObjectID:      302,
				ResumeID:      402,
				IsViewed:      false,
				CreatedAt:     fixedTime,
				ApplicantName: "",
				EmployerName:  "",
				Title:         "",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						3, "download_resume", 102, 202, 302, 402, false, fixedTime,
						"", "", "",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Превью с частично заполненными JOIN полями",
			notificationID: 4,
			expectedResult: &entity.NotificationPreview{
				ID:            4,
				Type:          entity.DownloadResumeType,
				SenderID:      103,
				ReceiverID:    203,
				ObjectID:      303,
				ResumeID:      403,
				IsViewed:      true,
				CreatedAt:     fixedTime,
				ApplicantName: "Елена",
				EmployerName:  "",
				Title:         "UX/UI Designer",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						4, "download_resume", 103, 203, 303, 403, true, fixedTime,
						"Елена", "", "UX/UI Designer",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Превью с только именем соискателя",
			notificationID: 5,
			expectedResult: &entity.NotificationPreview{
				ID:            5,
				Type:          entity.DownloadResumeType,
				SenderID:      104,
				ReceiverID:    204,
				ObjectID:      304,
				ResumeID:      404,
				IsViewed:      false,
				CreatedAt:     fixedTime,
				ApplicantName: "Дмитрий",
				EmployerName:  "",
				Title:         "",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						5, "download_resume", 104, 204, 304, 404, false, fixedTime,
						"Дмитрий", "", "",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Уведомление не найдено",
			notificationID: 999,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("ошибка при выполнении запроса GetDownloadResumeNotificationPreview: %v", sql.ErrNoRows),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:           "Ошибка базы данных при выполнении запроса",
			notificationID: 6,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("ошибка при выполнении запроса GetDownloadResumeNotificationPreview: %v", errors.New("database connection failed")),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnError(errors.New("database connection failed"))
			},
		},
		{
			name:           "Ошибка сканирования - неверный тип данных для ID",
			notificationID: 7,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("ошибка при выполнении запроса GetDownloadResumeNotificationPreview: %v", errors.New(`sql: Scan error on column index 0, name "id": converting driver.Value type string ("invalid_id") to a int: invalid syntax`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						"invalid_id", "download_resume", 100, 200, 300, 400, false, fixedTime,
						"Тест", "Тест Компания", "Тест Профессия",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка сканирования - неверный тип данных для времени",
			notificationID: 8,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("ошибка при выполнении запроса GetDownloadResumeNotificationPreview: %v", errors.New(`sql: Scan error on column index 7, name "created_at": unsupported Scan, storing driver.Value type string into type *time.Time`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						8, "download_resume", 100, 200, 300, 400, false, "invalid_time",
						"Тест", "Тест Компания", "Тест Профессия",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка сканирования NULL значений",
			notificationID: 9,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("ошибка при выполнении запроса GetDownloadResumeNotificationPreview: %v", errors.New(`sql: Scan error on column index 8, name "applicant_name": converting NULL to string is unsupported`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						9, "download_resume", 100, 200, 300, 400, false, fixedTime,
						nil, "Компания", "Профессия",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Превью с нулевыми ID",
			notificationID: 10,
			expectedResult: &entity.NotificationPreview{
				ID:            10,
				Type:          entity.DownloadResumeType,
				SenderID:      0,
				ReceiverID:    0,
				ObjectID:      0,
				ResumeID:      0,
				IsViewed:      false,
				CreatedAt:     fixedTime,
				ApplicantName: "Нулевой",
				EmployerName:  "Нулевая Компания",
				Title:         "Нулевая Профессия",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						10, "download_resume", 0, 0, 0, 0, false, fixedTime,
						"Нулевой", "Нулевая Компания", "Нулевая Профессия",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Превью с максимальными значениями ID",
			notificationID: 2147483647, // max int32
			expectedResult: &entity.NotificationPreview{
				ID:            2147483647,
				Type:          entity.DownloadResumeType,
				SenderID:      2147483646,
				ReceiverID:    2147483645,
				ObjectID:      2147483644,
				ResumeID:      2147483643,
				IsViewed:      true,
				CreatedAt:     fixedTime,
				ApplicantName: "Максимальный",
				EmployerName:  "Максимальная Компания",
				Title:         "Максимальная Профессия",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						2147483647, "download_resume", 2147483646, 2147483645, 2147483644, 2147483643, true, fixedTime,
						"Максимальный", "Максимальная Компания", "Максимальная Профессия",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Превью с длинными строками",
			notificationID: 11,
			expectedResult: &entity.NotificationPreview{
				ID:            11,
				Type:          entity.DownloadResumeType,
				SenderID:      105,
				ReceiverID:    205,
				ObjectID:      305,
				ResumeID:      405,
				IsViewed:      false,
				CreatedAt:     fixedTime,
				ApplicantName: "Очень Длинное Имя Соискателя Которое Может Быть В Базе Данных Системы",
				EmployerName:  "Очень Длинное Название Компании Работодателя Которое Тоже Может Быть В Базе Данных",
				Title:         "Очень Длинное Название Профессии Senior Full Stack Developer with Machine Learning and AI Experience",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						11, "download_resume", 105, 205, 305, 405, false, fixedTime,
						"Очень Длинное Имя Соискателя Которое Может Быть В Базе Данных Системы",
						"Очень Длинное Название Компании Работодателя Которое Тоже Может Быть В Базе Данных",
						"Очень Длинное Название Профессии Senior Full Stack Developer with Machine Learning and AI Experience",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Превью с отрицательным ID уведомления",
			notificationID: -1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("ошибка при выполнении запроса GetDownloadResumeNotificationPreview: %v", sql.ErrNoRows),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:           "Превью с специальными символами в строках",
			notificationID: 12,
			expectedResult: &entity.NotificationPreview{
				ID:            12,
				Type:          entity.DownloadResumeType,
				SenderID:      106,
				ReceiverID:    206,
				ObjectID:      306,
				ResumeID:      406,
				IsViewed:      true,
				CreatedAt:     fixedTime,
				ApplicantName: "Иван-Петр О'Коннор",
				EmployerName:  "ООО \"Рога & Копыта\"",
				Title:         "C++ Developer (Senior)",
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						12, "download_resume", 106, 206, 306, 406, true, fixedTime,
						"Иван-Петр О'Коннор", "ООО \"Рога & Копыта\"", "C++ Developer (Senior)",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}(db, mock)

			tc.setupMock(mock, tc.notificationID)

			repo := &NotificationRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetDownloadResumeNotificationPreview(ctx, tc.notificationID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tc.expectedResult.ID, result.ID)
				require.Equal(t, tc.expectedResult.Type, result.Type)
				require.Equal(t, tc.expectedResult.SenderID, result.SenderID)
				require.Equal(t, tc.expectedResult.ReceiverID, result.ReceiverID)
				require.Equal(t, tc.expectedResult.ObjectID, result.ObjectID)
				require.Equal(t, tc.expectedResult.ResumeID, result.ResumeID)
				require.Equal(t, tc.expectedResult.IsViewed, result.IsViewed)
				require.Equal(t, tc.expectedResult.CreatedAt.Unix(), result.CreatedAt.Unix())
				require.Equal(t, tc.expectedResult.ApplicantName, result.ApplicantName)
				require.Equal(t, tc.expectedResult.EmployerName, result.EmployerName)
				require.Equal(t, tc.expectedResult.Title, result.Title)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNotificationRepository_GetNotificationByID(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
		SELECT 
			id,
			type,
			sender_id,
			sender_role,
			receiver_id,
			receiver_role,
			object_id,
			resume_id,
			is_viewed,
			created_at
		FROM notification
		WHERE id = $1
	`)

	columns := []string{
		"id", "type", "sender_id", "sender_role", "receiver_id",
		"receiver_role", "object_id", "resume_id", "is_viewed", "created_at",
	}

	fixedTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		name           string
		notificationID int
		expectedResult *entity.Notification
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, notificationID int)
	}{
		{
			name:           "Успешное получение уведомления о отклике",
			notificationID: 1,
			expectedResult: &entity.Notification{
				ID:           1,
				Type:         entity.ApplyNotificationType,
				SenderID:     100,
				SenderRole:   entity.ApplicantRole,
				ReceiverID:   200,
				ReceiverRole: entity.EmployerRole,
				ObjectID:     300,
				ResumeID:     400,
				IsViewed:     false,
				CreatedAt:    fixedTime,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						1, "apply", 100, "applicant", 200, "employer", 300, 400, false, fixedTime,
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Успешное получение уведомления о скачивании резюме",
			notificationID: 2,
			expectedResult: &entity.Notification{
				ID:           2,
				Type:         entity.DownloadResumeType,
				SenderID:     101,
				SenderRole:   entity.EmployerRole,
				ReceiverID:   201,
				ReceiverRole: entity.ApplicantRole,
				ObjectID:     301,
				ResumeID:     401,
				IsViewed:     true,
				CreatedAt:    fixedTime,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						2, "download_resume", 101, "employer", 201, "applicant", 301, 401, true, fixedTime,
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Успешное получение непросмотренного уведомления",
			notificationID: 3,
			expectedResult: &entity.Notification{
				ID:           3,
				Type:         entity.ApplyNotificationType,
				SenderID:     102,
				SenderRole:   entity.ApplicantRole,
				ReceiverID:   202,
				ReceiverRole: entity.EmployerRole,
				ObjectID:     302,
				ResumeID:     402,
				IsViewed:     false,
				CreatedAt:    fixedTime,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						3, "apply", 102, "applicant", 202, "employer", 302, 402, false, fixedTime,
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Успешное получение уведомления с нулевыми ID",
			notificationID: 4,
			expectedResult: &entity.Notification{
				ID:           4,
				Type:         entity.DownloadResumeType,
				SenderID:     0,
				SenderRole:   entity.EmployerRole,
				ReceiverID:   0,
				ReceiverRole: entity.ApplicantRole,
				ObjectID:     0,
				ResumeID:     0,
				IsViewed:     true,
				CreatedAt:    fixedTime,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						4, "download_resume", 0, "employer", 0, "applicant", 0, 0, true, fixedTime,
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Успешное получение уведомления с максимальными ID",
			notificationID: 2147483647, // max int32
			expectedResult: &entity.Notification{
				ID:           2147483647,
				Type:         entity.ApplyNotificationType,
				SenderID:     2147483646,
				SenderRole:   entity.ApplicantRole,
				ReceiverID:   2147483645,
				ReceiverRole: entity.EmployerRole,
				ObjectID:     2147483644,
				ResumeID:     2147483643,
				IsViewed:     false,
				CreatedAt:    fixedTime,
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						2147483647, "apply", 2147483646, "applicant", 2147483645, "employer", 2147483644, 2147483643, false, fixedTime,
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Уведомление не найдено",
			notificationID: 999,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("уведомление не найдено"),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:           "Ошибка базы данных при выполнении запроса",
			notificationID: 5,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении SQL-запроса: %v", errors.New("database connection failed")),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnError(errors.New("database connection failed"))
			},
		},
		{
			name:           "Ошибка сканирования - неверный тип данных для ID",
			notificationID: 6,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении SQL-запроса: %v", errors.New(`sql: Scan error on column index 0, name "id": converting driver.Value type string ("invalid_id") to a int: invalid syntax`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						"invalid_id", "apply", 100, "applicant", 200, "employer", 300, 400, false, fixedTime,
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка сканирования - неверный тип данных для времени",
			notificationID: 7,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении SQL-запроса: %v", errors.New(`sql: Scan error on column index 9, name "created_at": unsupported Scan, storing driver.Value type string into type *time.Time`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						7, "apply", 100, "applicant", 200, "employer", 300, 400, false, "invalid_time",
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка сканирования - неверный тип данных для boolean",
			notificationID: 8,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении SQL-запроса: %v", errors.New(`sql: Scan error on column index 8, name "is_viewed": sql/driver: couldn't convert "invalid_bool" into type bool`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						8, "apply", 100, "applicant", 200, "employer", 300, 400, "invalid_bool", fixedTime,
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка сканирования - неверный тип уведомления",
			notificationID: 9,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении SQL-запроса: %v", errors.New(`sql: Scan error on column index 1, name "type": unsupported Scan, storing driver.Value type int64 into type *entity.NotificationType`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						9, 123, 100, "applicant", 200, "employer", 300, 400, false, fixedTime,
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Ошибка сканирования - неверная роль отправителя",
			notificationID: 10,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении SQL-запроса: %v", errors.New(`sql: Scan error on column index 3, name "sender_role": unsupported Scan, storing driver.Value type int64 into type *entity.UserRole`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				rows := sqlmock.NewRows(columns).
					AddRow(
						10, "apply", 100, 456, 200, "employer", 300, 400, false, fixedTime,
					)
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnRows(rows)
			},
		},
		{
			name:           "Отрицательный ID уведомления",
			notificationID: -1,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("уведомление не найдено"),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:           "Нулевой ID уведомления",
			notificationID: 0,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("уведомление не найдено"),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name:           "Ошибка подключения к базе данных",
			notificationID: 11,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении SQL-запроса: %v", errors.New("connection refused")),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnError(errors.New("connection refused"))
			},
		},
		{
			name:           "Ошибка таймаута базы данных",
			notificationID: 12,
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении SQL-запроса: %v", errors.New("query timeout")),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectQuery(query).
					WithArgs(notificationID).
					WillReturnError(errors.New("query timeout"))
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}(db, mock)

			tc.setupMock(mock, tc.notificationID)

			repo := &NotificationRepository{DB: db}
			ctx := context.Background()

			result, err := repo.GetNotificationByID(ctx, tc.notificationID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tc.expectedResult.ID, result.ID)
				require.Equal(t, tc.expectedResult.Type, result.Type)
				require.Equal(t, tc.expectedResult.SenderID, result.SenderID)
				require.Equal(t, tc.expectedResult.SenderRole, result.SenderRole)
				require.Equal(t, tc.expectedResult.ReceiverID, result.ReceiverID)
				require.Equal(t, tc.expectedResult.ReceiverRole, result.ReceiverRole)
				require.Equal(t, tc.expectedResult.ObjectID, result.ObjectID)
				require.Equal(t, tc.expectedResult.ResumeID, result.ResumeID)
				require.Equal(t, tc.expectedResult.IsViewed, result.IsViewed)
				require.Equal(t, tc.expectedResult.CreatedAt.Unix(), result.CreatedAt.Unix())
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNotificationRepository_CreateNotification(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
		INSERT INTO notification (
			type,
			sender_id,
		    sender_role,
			receiver_id,
		    receiver_role,
			object_id,
		    resume_id,
			is_viewed
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`)

	fixedTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	testCases := []struct {
		name         string
		notification *entity.Notification
		expectedID   int
		expectedErr  error
		setupMock    func(mock sqlmock.Sqlmock, notification *entity.Notification)
	}{
		{
			name: "Успешное создание уведомления о отклике",
			notification: &entity.Notification{
				Type:         entity.ApplyNotificationType,
				SenderID:     100,
				SenderRole:   entity.ApplicantRole,
				ReceiverID:   200,
				ReceiverRole: entity.EmployerRole,
				ObjectID:     300,
				ResumeID:     400,
				IsViewed:     false,
				CreatedAt:    fixedTime,
			},
			expectedID:  1,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notification *entity.Notification) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(query).
					WithArgs(
						notification.Type,
						notification.SenderID,
						notification.SenderRole,
						notification.ReceiverID,
						notification.ReceiverRole,
						notification.ObjectID,
						notification.ResumeID,
						notification.IsViewed,
					).
					WillReturnRows(rows)
			},
		},
		{
			name: "Успешное создание уведомления о скачивании резюме",
			notification: &entity.Notification{
				Type:         entity.DownloadResumeType,
				SenderID:     200,
				SenderRole:   entity.EmployerRole,
				ReceiverID:   100,
				ReceiverRole: entity.ApplicantRole,
				ObjectID:     500,
				ResumeID:     600,
				IsViewed:     false,
				CreatedAt:    fixedTime,
			},
			expectedID:  2,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notification *entity.Notification) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(2)
				mock.ExpectQuery(query).
					WithArgs(
						notification.Type,
						notification.SenderID,
						notification.SenderRole,
						notification.ReceiverID,
						notification.ReceiverRole,
						notification.ObjectID,
						notification.ResumeID,
						notification.IsViewed,
					).
					WillReturnRows(rows)
			},
		},
		{
			name: "Успешное создание просмотренного уведомления",
			notification: &entity.Notification{
				Type:         entity.ApplyNotificationType,
				SenderID:     101,
				SenderRole:   entity.ApplicantRole,
				ReceiverID:   201,
				ReceiverRole: entity.EmployerRole,
				ObjectID:     301,
				ResumeID:     401,
				IsViewed:     true,
				CreatedAt:    fixedTime,
			},
			expectedID:  3,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notification *entity.Notification) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(3)
				mock.ExpectQuery(query).
					WithArgs(
						notification.Type,
						notification.SenderID,
						notification.SenderRole,
						notification.ReceiverID,
						notification.ReceiverRole,
						notification.ObjectID,
						notification.ResumeID,
						notification.IsViewed,
					).
					WillReturnRows(rows)
			},
		},
		{
			name: "Успешное создание уведомления с нулевыми ID",
			notification: &entity.Notification{
				Type:         entity.DownloadResumeType,
				SenderID:     0,
				SenderRole:   entity.EmployerRole,
				ReceiverID:   0,
				ReceiverRole: entity.ApplicantRole,
				ObjectID:     0,
				ResumeID:     0,
				IsViewed:     false,
				CreatedAt:    fixedTime,
			},
			expectedID:  4,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notification *entity.Notification) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(4)
				mock.ExpectQuery(query).
					WithArgs(
						notification.Type,
						notification.SenderID,
						notification.SenderRole,
						notification.ReceiverID,
						notification.ReceiverRole,
						notification.ObjectID,
						notification.ResumeID,
						notification.IsViewed,
					).
					WillReturnRows(rows)
			},
		},
		{
			name: "Успешное создание уведомления с максимальными ID",
			notification: &entity.Notification{
				Type:         entity.ApplyNotificationType,
				SenderID:     2147483646,
				SenderRole:   entity.ApplicantRole,
				ReceiverID:   2147483645,
				ReceiverRole: entity.EmployerRole,
				ObjectID:     2147483644,
				ResumeID:     2147483643,
				IsViewed:     true,
				CreatedAt:    fixedTime,
			},
			expectedID:  2147483647,
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, notification *entity.Notification) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(2147483647)
				mock.ExpectQuery(query).
					WithArgs(
						notification.Type,
						notification.SenderID,
						notification.SenderRole,
						notification.ReceiverID,
						notification.ReceiverRole,
						notification.ObjectID,
						notification.ResumeID,
						notification.IsViewed,
					).
					WillReturnRows(rows)
			},
		},
		{
			name: "Ошибка базы данных при выполнении запроса",
			notification: &entity.Notification{
				Type:         entity.ApplyNotificationType,
				SenderID:     102,
				SenderRole:   entity.ApplicantRole,
				ReceiverID:   202,
				ReceiverRole: entity.EmployerRole,
				ObjectID:     302,
				ResumeID:     402,
				IsViewed:     false,
				CreatedAt:    fixedTime,
			},
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("ошибка при создании уведомления: %v", errors.New("database connection failed")),
			),
			setupMock: func(mock sqlmock.Sqlmock, notification *entity.Notification) {
				mock.ExpectQuery(query).
					WithArgs(
						notification.Type,
						notification.SenderID,
						notification.SenderRole,
						notification.ReceiverID,
						notification.ReceiverRole,
						notification.ObjectID,
						notification.ResumeID,
						notification.IsViewed,
					).
					WillReturnError(errors.New("database connection failed"))
			},
		},
		{
			name: "Ошибка нарушения ограничения уникальности",
			notification: &entity.Notification{
				Type:         entity.ApplyNotificationType,
				SenderID:     103,
				SenderRole:   entity.ApplicantRole,
				ReceiverID:   203,
				ReceiverRole: entity.EmployerRole,
				ObjectID:     303,
				ResumeID:     403,
				IsViewed:     false,
				CreatedAt:    fixedTime,
			},
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("ошибка при создании уведомления: %v", errors.New("duplicate key value violates unique constraint")),
			),
			setupMock: func(mock sqlmock.Sqlmock, notification *entity.Notification) {
				mock.ExpectQuery(query).
					WithArgs(
						notification.Type,
						notification.SenderID,
						notification.SenderRole,
						notification.ReceiverID,
						notification.ReceiverRole,
						notification.ObjectID,
						notification.ResumeID,
						notification.IsViewed,
					).
					WillReturnError(errors.New("duplicate key value violates unique constraint"))
			},
		},
		{
			name: "Ошибка нарушения внешнего ключа",
			notification: &entity.Notification{
				Type:         entity.DownloadResumeType,
				SenderID:     999,
				SenderRole:   entity.EmployerRole,
				ReceiverID:   998,
				ReceiverRole: entity.ApplicantRole,
				ObjectID:     997,
				ResumeID:     996,
				IsViewed:     false,
				CreatedAt:    fixedTime,
			},
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("ошибка при создании уведомления: %v", errors.New("foreign key constraint violation")),
			),
			setupMock: func(mock sqlmock.Sqlmock, notification *entity.Notification) {
				mock.ExpectQuery(query).
					WithArgs(
						notification.Type,
						notification.SenderID,
						notification.SenderRole,
						notification.ReceiverID,
						notification.ReceiverRole,
						notification.ObjectID,
						notification.ResumeID,
						notification.IsViewed,
					).
					WillReturnError(errors.New("foreign key constraint violation"))
			},
		},
		{
			name: "Ошибка сканирования возвращаемого ID",
			notification: &entity.Notification{
				Type:         entity.ApplyNotificationType,
				SenderID:     104,
				SenderRole:   entity.ApplicantRole,
				ReceiverID:   204,
				ReceiverRole: entity.EmployerRole,
				ObjectID:     304,
				ResumeID:     404,
				IsViewed:     false,
				CreatedAt:    fixedTime,
			},
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("ошибка при создании уведомления: %v", errors.New(`sql: Scan error on column index 0, name "id": converting driver.Value type string ("invalid_id") to a int: invalid syntax`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, notification *entity.Notification) {
				rows := sqlmock.NewRows([]string{"id"}).AddRow("invalid_id")
				mock.ExpectQuery(query).
					WithArgs(
						notification.Type,
						notification.SenderID,
						notification.SenderRole,
						notification.ReceiverID,
						notification.ReceiverRole,
						notification.ObjectID,
						notification.ResumeID,
						notification.IsViewed,
					).
					WillReturnRows(rows)
			},
		},
		{
			name: "Ошибка отсутствия возвращаемых строк",
			notification: &entity.Notification{
				Type:         entity.DownloadResumeType,
				SenderID:     105,
				SenderRole:   entity.EmployerRole,
				ReceiverID:   205,
				ReceiverRole: entity.ApplicantRole,
				ObjectID:     305,
				ResumeID:     405,
				IsViewed:     true,
				CreatedAt:    fixedTime,
			},
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("ошибка при создании уведомления: %v", sql.ErrNoRows),
			),
			setupMock: func(mock sqlmock.Sqlmock, notification *entity.Notification) {
				mock.ExpectQuery(query).
					WithArgs(
						notification.Type,
						notification.SenderID,
						notification.SenderRole,
						notification.ReceiverID,
						notification.ReceiverRole,
						notification.ObjectID,
						notification.ResumeID,
						notification.IsViewed,
					).
					WillReturnError(sql.ErrNoRows)
			},
		},
		{
			name: "Ошибка таймаута базы данных",
			notification: &entity.Notification{
				Type:         entity.ApplyNotificationType,
				SenderID:     106,
				SenderRole:   entity.ApplicantRole,
				ReceiverID:   206,
				ReceiverRole: entity.EmployerRole,
				ObjectID:     306,
				ResumeID:     406,
				IsViewed:     false,
				CreatedAt:    fixedTime,
			},
			expectedID: 0,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("ошибка при создании уведомления: %v", errors.New("query timeout")),
			),
			setupMock: func(mock sqlmock.Sqlmock, notification *entity.Notification) {
				mock.ExpectQuery(query).
					WithArgs(
						notification.Type,
						notification.SenderID,
						notification.SenderRole,
						notification.ReceiverID,
						notification.ReceiverRole,
						notification.ObjectID,
						notification.ResumeID,
						notification.IsViewed,
					).
					WillReturnError(errors.New("query timeout"))
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}(db, mock)

			tc.setupMock(mock, tc.notification)

			repo := &NotificationRepository{DB: db}
			ctx := context.Background()

			// Сохраняем исходное состояние для проверки
			originalID := tc.notification.ID

			err = repo.CreateNotification(ctx, tc.notification)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				// При ошибке ID не должен измениться
				require.Equal(t, originalID, tc.notification.ID)
			} else {
				require.NoError(t, err)
				// При успехе ID должен быть установлен
				require.Equal(t, tc.expectedID, tc.notification.ID)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNotificationRepository_ReadNotification(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
		UPDATE notification
		SET is_viewed = true
		WHERE id = $1
	`)

	testCases := []struct {
		name           string
		notificationID int
		expectedErr    error
		setupMock      func(mock sqlmock.Sqlmock, notificationID int)
	}{
		{
			name:           "Успешная пометка уведомления как прочитанного",
			notificationID: 1,
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectExec(query).
					WithArgs(notificationID).
					WillReturnResult(sqlmock.NewResult(0, 1)) // 1 строка затронута
			},
		},
		{
			name:           "Успешная пометка уведомления с большим ID",
			notificationID: 2147483647, // max int32
			expectedErr:    nil,
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectExec(query).
					WithArgs(notificationID).
					WillReturnResult(sqlmock.NewResult(0, 1)) // 1 строка затронута
			},
		},
		{
			name:           "Уведомление не найдено - 0 затронутых строк",
			notificationID: 999,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("уведомление не найдено в ReadNotification"),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectExec(query).
					WithArgs(notificationID).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк затронуто
			},
		},
		{
			name:           "Отрицательный ID уведомления",
			notificationID: -1,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("уведомление не найдено в ReadNotification"),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectExec(query).
					WithArgs(notificationID).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк затронуто
			},
		},
		{
			name:           "Нулевой ID уведомления",
			notificationID: 0,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("уведомление не найдено в ReadNotification"),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectExec(query).
					WithArgs(notificationID).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк затронуто
			},
		},
		{
			name:           "Ошибка базы данных при выполнении UPDATE",
			notificationID: 2,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("ошибка при выполнении ReadNotification: %v", errors.New("database connection failed")),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectExec(query).
					WithArgs(notificationID).
					WillReturnError(errors.New("database connection failed"))
			},
		},
		{
			name:           "Ошибка получения количества затронутых строк",
			notificationID: 3,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("не удалось получить число затронутых строк в ReadNotification: %v", errors.New("rows affected error")),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				result := sqlmock.NewErrorResult(errors.New("rows affected error"))
				mock.ExpectExec(query).
					WithArgs(notificationID).
					WillReturnResult(result)
			},
		},
		{
			name:           "Ошибка таймаута базы данных",
			notificationID: 4,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("ошибка при выполнении ReadNotification: %v", errors.New("query timeout")),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectExec(query).
					WithArgs(notificationID).
					WillReturnError(errors.New("query timeout"))
			},
		},
		{
			name:           "Ошибка нарушения ограничений базы данных",
			notificationID: 5,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("ошибка при выполнении ReadNotification: %v", errors.New("constraint violation")),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectExec(query).
					WithArgs(notificationID).
					WillReturnError(errors.New("constraint violation"))
			},
		},
		{
			name:           "Ошибка блокировки таблицы",
			notificationID: 6,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("ошибка при выполнении ReadNotification: %v", errors.New("table lock timeout")),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectExec(query).
					WithArgs(notificationID).
					WillReturnError(errors.New("table lock timeout"))
			},
		},
		{
			name:           "Ошибка недостатка прав доступа",
			notificationID: 7,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("ошибка при выполнении ReadNotification: %v", errors.New("permission denied")),
			),
			setupMock: func(mock sqlmock.Sqlmock, notificationID int) {
				mock.ExpectExec(query).
					WithArgs(notificationID).
					WillReturnError(errors.New("permission denied"))
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}(db, mock)

			tc.setupMock(mock, tc.notificationID)

			repo := &NotificationRepository{DB: db}
			ctx := context.Background()

			err = repo.ReadNotification(ctx, tc.notificationID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNotificationRepository_ReadAllNotifications(t *testing.T) {
	t.Parallel()

	applicantQuery := regexp.QuoteMeta(`
		UPDATE notification
		SET is_viewed = true
		WHERE receiver_id = $1 AND type = 'download_resume'
	`)

	employerQuery := regexp.QuoteMeta(`
		UPDATE notification
		SET is_viewed = true
		WHERE receiver_id = $1 AND type = 'apply'
	`)

	testCases := []struct {
		name        string
		userID      int
		role        string
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, userID int, role string)
	}{
		{
			name:        "Успешная пометка всех уведомлений для соискателя",
			userID:      100,
			role:        "applicant",
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(applicantQuery).
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 3)) // 3 строки затронуты
			},
		},
		{
			name:        "Успешная пометка всех уведомлений для работодателя",
			userID:      200,
			role:        "employer",
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(employerQuery).
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 5)) // 5 строк затронуты
			},
		},
		{
			name:        "Успешная пометка для соискателя без уведомлений",
			userID:      101,
			role:        "applicant",
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(applicantQuery).
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк затронуто
			},
		},
		{
			name:        "Успешная пометка для работодателя без уведомлений",
			userID:      201,
			role:        "employer",
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(employerQuery).
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк затронуто
			},
		},
		{
			name:        "Успешная пометка для соискателя с нулевым ID",
			userID:      0,
			role:        "applicant",
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(applicantQuery).
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк затронуто
			},
		},
		{
			name:        "Успешная пометка для работодателя с максимальным ID",
			userID:      2147483647, // max int32
			role:        "employer",
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(employerQuery).
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 1)) // 1 строка затронута
			},
		},
		{
			name:   "Неподдерживаемая роль - admin",
			userID: 102,
			role:   "admin",
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("несуществующая роль: %s", "admin"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				// Никаких ожиданий, так как запрос не должен выполняться
			},
		},
		{
			name:   "Неподдерживаемая роль - пустая строка",
			userID: 103,
			role:   "",
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("несуществующая роль: %s", ""),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				// Никаких ожиданий, так как запрос не должен выполняться
			},
		},
		{
			name:   "Неподдерживаемая роль - неверный регистр",
			userID: 104,
			role:   "APPLICANT",
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("несуществующая роль: %s", "APPLICANT"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				// Никаких ожиданий, так как запрос не должен выполняться
			},
		},
		{
			name:   "Неподдерживаемая роль - случайная строка",
			userID: 105,
			role:   "random_role",
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("несуществующая роль: %s", "random_role"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				// Никаких ожиданий, так как запрос не должен выполняться
			},
		},
		{
			name:   "Ошибка базы данных для соискателя",
			userID: 106,
			role:   "applicant",
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении ReadAllNotifications: %v", errors.New("database connection failed")),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(applicantQuery).
					WithArgs(userID).
					WillReturnError(errors.New("database connection failed"))
			},
		},
		{
			name:   "Ошибка базы данных для работодателя",
			userID: 206,
			role:   "employer",
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении ReadAllNotifications: %v", errors.New("database connection failed")),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(employerQuery).
					WithArgs(userID).
					WillReturnError(errors.New("database connection failed"))
			},
		},
		{
			name:   "Ошибка таймаута для соискателя",
			userID: 107,
			role:   "applicant",
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении ReadAllNotifications: %v", errors.New("query timeout")),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(applicantQuery).
					WithArgs(userID).
					WillReturnError(errors.New("query timeout"))
			},
		},
		{
			name:   "Ошибка блокировки таблицы для работодателя",
			userID: 207,
			role:   "employer",
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении ReadAllNotifications: %v", errors.New("table lock timeout")),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(employerQuery).
					WithArgs(userID).
					WillReturnError(errors.New("table lock timeout"))
			},
		},
		{
			name:        "Отрицательный ID пользователя для соискателя",
			userID:      -1,
			role:        "applicant",
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(applicantQuery).
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк затронуто
			},
		},
		{
			name:        "Отрицательный ID пользователя для работодателя",
			userID:      -1,
			role:        "employer",
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(employerQuery).
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк затронуто
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}(db, mock)

			tc.setupMock(mock, tc.userID, tc.role)

			repo := &NotificationRepository{DB: db}
			ctx := context.Background()

			err = repo.ReadAllNotifications(ctx, tc.userID, tc.role)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNotificationRepository_DeleteAllNotifications(t *testing.T) {
	t.Parallel()

	applicantQuery := regexp.QuoteMeta(`
		DELETE FROM notification
		WHERE receiver_id = $1 AND type = 'download_resume'
	`)

	employerQuery := regexp.QuoteMeta(`
		DELETE FROM notification
		WHERE receiver_id = $1 AND type = 'apply'
	`)

	testCases := []struct {
		name        string
		userID      int
		role        string
		expectedErr error
		setupMock   func(mock sqlmock.Sqlmock, userID int, role string)
	}{
		{
			name:        "Успешное удаление всех уведомлений для соискателя",
			userID:      100,
			role:        "applicant",
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(applicantQuery).
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 3)) // 3 строки удалены
			},
		},
		{
			name:        "Успешное удаление всех уведомлений для работодателя",
			userID:      200,
			role:        "employer",
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(employerQuery).
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 5)) // 5 строк удалены
			},
		},
		{
			name:        "Успешное удаление для соискателя без уведомлений",
			userID:      101,
			role:        "applicant",
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(applicantQuery).
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк удалено
			},
		},
		{
			name:        "Успешное удаление для работодателя без уведомлений",
			userID:      201,
			role:        "employer",
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(employerQuery).
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк удалено
			},
		},
		{
			name:        "Успешное удаление для соискателя с нулевым ID",
			userID:      0,
			role:        "applicant",
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(applicantQuery).
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк удалено
			},
		},
		{
			name:        "Успешное удаление для работодателя с максимальным ID",
			userID:      2147483647, // max int32
			role:        "employer",
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(employerQuery).
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 1)) // 1 строка удалена
			},
		},
		{
			name:        "Успешное удаление одного уведомления для соискателя",
			userID:      102,
			role:        "applicant",
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(applicantQuery).
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 1)) // 1 строка удалена
			},
		},
		{
			name:        "Успешное удаление множественных уведомлений для работодателя",
			userID:      202,
			role:        "employer",
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(employerQuery).
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 100)) // 100 строк удалено
			},
		},
		{
			name:   "Неподдерживаемая роль - admin",
			userID: 103,
			role:   "admin",
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("несуществующая роль: %s", "admin"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				// Никаких ожиданий, так как запрос не должен выполняться
			},
		},
		{
			name:   "Неподдерживаемая роль - пустая строка",
			userID: 104,
			role:   "",
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("несуществующая роль: %s", ""),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				// Никаких ожиданий, так как запрос не должен выполняться
			},
		},
		{
			name:   "Неподдерживаемая роль - неверный регистр",
			userID: 105,
			role:   "EMPLOYER",
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("несуществующая роль: %s", "EMPLOYER"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				// Никаких ожиданий, так как запрос не должен выполняться
			},
		},
		{
			name:   "Неподдерживаемая роль - случайная строка",
			userID: 106,
			role:   "random_role",
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("несуществующая роль: %s", "random_role"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				// Никаких ожиданий, так как запрос не должен выполняться
			},
		},
		{
			name:   "Неподдерживаемая роль - числовая строка",
			userID: 107,
			role:   "123",
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("несуществующая роль: %s", "123"),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				// Никаких ожиданий, так как запрос не должен выполняться
			},
		},
		{
			name:   "Ошибка базы данных для соискателя",
			userID: 108,
			role:   "applicant",
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении DeleteAllNotifications: %v", errors.New("database connection failed")),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(applicantQuery).
					WithArgs(userID).
					WillReturnError(errors.New("database connection failed"))
			},
		},
		{
			name:   "Ошибка базы данных для работодателя",
			userID: 208,
			role:   "employer",
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении DeleteAllNotifications: %v", errors.New("database connection failed")),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(employerQuery).
					WithArgs(userID).
					WillReturnError(errors.New("database connection failed"))
			},
		},
		{
			name:   "Ошибка таймаута для соискателя",
			userID: 109,
			role:   "applicant",
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении DeleteAllNotifications: %v", errors.New("query timeout")),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(applicantQuery).
					WithArgs(userID).
					WillReturnError(errors.New("query timeout"))
			},
		},
		{
			name:   "Ошибка блокировки таблицы для работодателя",
			userID: 209,
			role:   "employer",
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении DeleteAllNotifications: %v", errors.New("table lock timeout")),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(employerQuery).
					WithArgs(userID).
					WillReturnError(errors.New("table lock timeout"))
			},
		},
		{
			name:   "Ошибка нарушения внешнего ключа для соискателя",
			userID: 110,
			role:   "applicant",
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении DeleteAllNotifications: %v", errors.New("foreign key constraint violation")),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(applicantQuery).
					WithArgs(userID).
					WillReturnError(errors.New("foreign key constraint violation"))
			},
		},
		{
			name:   "Ошибка недостатка прав доступа для работодателя",
			userID: 210,
			role:   "employer",
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении DeleteAllNotifications: %v", errors.New("permission denied")),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(employerQuery).
					WithArgs(userID).
					WillReturnError(errors.New("permission denied"))
			},
		},
		{
			name:        "Отрицательный ID пользователя для соискателя",
			userID:      -1,
			role:        "applicant",
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(applicantQuery).
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк удалено
			},
		},
		{
			name:        "Отрицательный ID пользователя для работодателя",
			userID:      -1,
			role:        "employer",
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int, role string) {
				mock.ExpectExec(employerQuery).
					WithArgs(userID).
					WillReturnResult(sqlmock.NewResult(0, 0)) // 0 строк удалено
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}(db, mock)

			tc.setupMock(mock, tc.userID, tc.role)

			repo := &NotificationRepository{DB: db}
			ctx := context.Background()

			err = repo.DeleteAllNotifications(ctx, tc.userID, tc.role)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNotificationRepository_GetApplyNotificationsForUser(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
        SELECT 
            n.id,
            n.type,
            n.sender_id,
            n.receiver_id,
            n.object_id,
            n.resume_id,
            n.is_viewed,
            n.created_at,
            a.first_name AS applicant_name,
            e.company_name AS employer_name,
            v.title
        FROM notification n
        LEFT JOIN applicant a ON n.sender_id = a.id
        LEFT JOIN employer e ON n.receiver_id = e.id
        LEFT JOIN vacancy v ON n.object_id = v.id
        WHERE n.type = 'apply' AND n.receiver_id = $1
        ORDER BY n.created_at DESC
    `)

	fixedTime1 := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	fixedTime2 := time.Date(2024, 1, 14, 10, 30, 0, 0, time.UTC)
	fixedTime3 := time.Date(2024, 1, 13, 8, 15, 0, 0, time.UTC)

	testCases := []struct {
		name                  string
		userID                int
		expectedNotifications []*entity.NotificationPreview
		expectedErr           error
		setupMock             func(mock sqlmock.Sqlmock, userID int)
	}{
		{
			name:   "Успешное получение одного уведомления",
			userID: 100,
			expectedNotifications: []*entity.NotificationPreview{
				{
					ID:            1,
					Type:          entity.ApplyNotificationType,
					SenderID:      200,
					ReceiverID:    100,
					ObjectID:      300,
					ResumeID:      400,
					IsViewed:      false,
					CreatedAt:     fixedTime1,
					ApplicantName: "Иван",
					EmployerName:  "ООО Компания",
					Title:         "Разработчик Go",
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				}).AddRow(
					1, "apply", 200, 100, 300, 400,
					false, fixedTime1, "Иван", "ООО Компания", "Разработчик Go",
				)
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:   "Успешное получение множественных уведомлений в правильном порядке",
			userID: 101,
			expectedNotifications: []*entity.NotificationPreview{
				{
					ID:            1,
					Type:          entity.ApplyNotificationType,
					SenderID:      201,
					ReceiverID:    101,
					ObjectID:      301,
					ResumeID:      401,
					IsViewed:      false,
					CreatedAt:     fixedTime1,
					ApplicantName: "Петр",
					EmployerName:  "ООО Рога",
					Title:         "Backend Developer",
				},
				{
					ID:            2,
					Type:          entity.ApplyNotificationType,
					SenderID:      202,
					ReceiverID:    101,
					ObjectID:      302,
					ResumeID:      402,
					IsViewed:      true,
					CreatedAt:     fixedTime2,
					ApplicantName: "Анна",
					EmployerName:  "ООО Копыта",
					Title:         "Frontend Developer",
				},
				{
					ID:            3,
					Type:          entity.ApplyNotificationType,
					SenderID:      203,
					ReceiverID:    101,
					ObjectID:      303,
					ResumeID:      403,
					IsViewed:      false,
					CreatedAt:     fixedTime3,
					ApplicantName: "Мария",
					EmployerName:  "ООО Хвосты",
					Title:         "Fullstack Developer",
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				}).
					AddRow(1, "apply", 201, 101, 301, 401, false, fixedTime1, "Петр", "ООО Рога", "Backend Developer").
					AddRow(2, "apply", 202, 101, 302, 402, true, fixedTime2, "Анна", "ООО Копыта", "Frontend Developer").
					AddRow(3, "apply", 203, 101, 303, 403, false, fixedTime3, "Мария", "ООО Хвосты", "Fullstack Developer")
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:                  "Успешное получение пустого списка уведомлений",
			userID:                102,
			expectedNotifications: nil,
			expectedErr:           nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				})
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:                  "Успешное получение уведомлений с NULL значениями в JOIN полях",
			userID:                103,
			expectedNotifications: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании строки результата запроса GetApplyNotificationsForUser: %v", errors.New(`sql: Scan error on column index 8, name "applicant_name": converting NULL to string is unsupported`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				}).AddRow(
					4, "apply", 204, 103, 304, 404,
					false, fixedTime1, nil, nil, nil,
				)
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:                  "Успешное получение уведомлений с частично NULL значениями",
			userID:                104,
			expectedNotifications: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании строки результата запроса GetApplyNotificationsForUser: %v", errors.New(`sql: Scan error on column index 9, name "employer_name": converting NULL to string is unsupported`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				}).AddRow(
					5, "apply", 205, 104, 305, 405,
					true, fixedTime1, "Сергей", nil, "DevOps Engineer",
				)
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:   "Успешное получение уведомлений с нулевыми ID",
			userID: 0,
			expectedNotifications: []*entity.NotificationPreview{
				{
					ID:            6,
					Type:          entity.ApplyNotificationType,
					SenderID:      0,
					ReceiverID:    0,
					ObjectID:      0,
					ResumeID:      0,
					IsViewed:      false,
					CreatedAt:     fixedTime1,
					ApplicantName: "Тест",
					EmployerName:  "Тест Компания",
					Title:         "Тест Вакансия",
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				}).AddRow(
					6, "apply", 0, 0, 0, 0,
					false, fixedTime1, "Тест", "Тест Компания", "Тест Вакансия",
				)
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:   "Успешное получение уведомлений с максимальными ID",
			userID: 2147483647,
			expectedNotifications: []*entity.NotificationPreview{
				{
					ID:            2147483647,
					Type:          entity.ApplyNotificationType,
					SenderID:      2147483646,
					ReceiverID:    2147483647,
					ObjectID:      2147483645,
					ResumeID:      2147483644,
					IsViewed:      true,
					CreatedAt:     fixedTime1,
					ApplicantName: "Максимум",
					EmployerName:  "Максимум Компания",
					Title:         "Максимум Вакансия",
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				}).AddRow(
					2147483647, "apply", 2147483646, 2147483647, 2147483645, 2147483644,
					true, fixedTime1, "Максимум", "Максимум Компания", "Максимум Вакансия",
				)
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:                  "Ошибка базы данных при выполнении запроса",
			userID:                105,
			expectedNotifications: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении запроса GetApplyNotificationsForUser: %v", errors.New("database connection failed")),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				mock.ExpectQuery(query).WithArgs(userID).WillReturnError(errors.New("database connection failed"))
			},
		},
		{
			name:                  "Ошибка таймаута базы данных",
			userID:                106,
			expectedNotifications: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении запроса GetApplyNotificationsForUser: %v", errors.New("query timeout")),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				mock.ExpectQuery(query).WithArgs(userID).WillReturnError(errors.New("query timeout"))
			},
		},
		{
			name:                  "Ошибка сканирования строки результата",
			userID:                107,
			expectedNotifications: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании строки результата запроса GetApplyNotificationsForUser: %v", errors.New(`sql: Scan error on column index 0, name "id": converting driver.Value type string ("invalid_id") to a int: invalid syntax`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				}).AddRow(
					"invalid_id", "apply", 207, 107, 307, 407,
					false, fixedTime1, "Тест", "Тест Компания", "Тест Вакансия",
				)
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:                  "Ошибка сканирования типа уведомления",
			userID:                108,
			expectedNotifications: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании строки результата запроса GetApplyNotificationsForUser: %v", errors.New(`sql: Scan error on column index 1, name "type": converting NULL to string is unsupported`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				}).AddRow(
					7, nil, 208, 108, 308, 408,
					false, fixedTime1, "Тест", "Тест Компания", "Тест Вакансия",
				)
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:                  "Ошибка сканирования времени создания",
			userID:                109,
			expectedNotifications: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании строки результата запроса GetApplyNotificationsForUser: %v", errors.New(`sql: Scan error on column index 7, name "created_at": unsupported Scan, storing driver.Value type string into type *time.Time`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				}).AddRow(
					8, "apply", 209, 109, 309, 409,
					false, "invalid_time", "Тест", "Тест Компания", "Тест Вакансия",
				)
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:                  "Отрицательный ID пользователя",
			userID:                -1,
			expectedNotifications: nil,
			expectedErr:           nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				})
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}(db, mock)

			tc.setupMock(mock, tc.userID)

			repo := &NotificationRepository{DB: db}
			ctx := context.Background()

			notifications, err := repo.GetApplyNotificationsForUser(ctx, tc.userID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Nil(t, notifications)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedNotifications, notifications)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestNotificationRepository_GetDownloadResumeNotificationsForUser(t *testing.T) {
	t.Parallel()

	query := regexp.QuoteMeta(`
        SELECT 
            n.id,
            n.type,
            n.sender_id,
            n.receiver_id,
            n.object_id,
            n.resume_id,
            n.is_viewed,
            n.created_at,
            a.first_name AS applicant_name,
            e.company_name AS employer_name,
            r.profession AS title
        FROM notification n
        LEFT JOIN resume r ON r.id = n.object_id
        LEFT JOIN applicant a ON a.id = r.applicant_id
        LEFT JOIN employer e ON e.id = n.sender_id
        WHERE n.receiver_id = $1 AND n.type = 'download_resume'
        ORDER BY n.created_at DESC
    `)

	fixedTime1 := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	fixedTime2 := time.Date(2024, 1, 14, 10, 30, 0, 0, time.UTC)
	fixedTime3 := time.Date(2024, 1, 13, 8, 15, 0, 0, time.UTC)

	testCases := []struct {
		name                  string
		userID                int
		expectedNotifications []*entity.NotificationPreview
		expectedErr           error
		setupMock             func(mock sqlmock.Sqlmock, userID int)
	}{
		{
			name:   "Успешное получение одного уведомления",
			userID: 100,
			expectedNotifications: []*entity.NotificationPreview{
				{
					ID:            1,
					Type:          entity.DownloadResumeType,
					SenderID:      200,
					ReceiverID:    100,
					ObjectID:      300,
					ResumeID:      400,
					IsViewed:      false,
					CreatedAt:     fixedTime1,
					ApplicantName: "Иван",
					EmployerName:  "ООО Компания",
					Title:         "Разработчик Go",
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				}).AddRow(
					1, "download_resume", 200, 100, 300, 400,
					false, fixedTime1, "Иван", "ООО Компания", "Разработчик Go",
				)
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:   "Успешное получение множественных уведомлений в правильном порядке",
			userID: 101,
			expectedNotifications: []*entity.NotificationPreview{
				{
					ID:            1,
					Type:          entity.DownloadResumeType,
					SenderID:      201,
					ReceiverID:    101,
					ObjectID:      301,
					ResumeID:      401,
					IsViewed:      false,
					CreatedAt:     fixedTime1,
					ApplicantName: "Петр",
					EmployerName:  "ООО Рога",
					Title:         "Backend Developer",
				},
				{
					ID:            2,
					Type:          entity.DownloadResumeType,
					SenderID:      202,
					ReceiverID:    101,
					ObjectID:      302,
					ResumeID:      402,
					IsViewed:      true,
					CreatedAt:     fixedTime2,
					ApplicantName: "Анна",
					EmployerName:  "ООО Копыта",
					Title:         "Frontend Developer",
				},
				{
					ID:            3,
					Type:          entity.DownloadResumeType,
					SenderID:      203,
					ReceiverID:    101,
					ObjectID:      303,
					ResumeID:      403,
					IsViewed:      false,
					CreatedAt:     fixedTime3,
					ApplicantName: "Мария",
					EmployerName:  "ООО Хвосты",
					Title:         "Fullstack Developer",
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				}).
					AddRow(1, "download_resume", 201, 101, 301, 401, false, fixedTime1, "Петр", "ООО Рога", "Backend Developer").
					AddRow(2, "download_resume", 202, 101, 302, 402, true, fixedTime2, "Анна", "ООО Копыта", "Frontend Developer").
					AddRow(3, "download_resume", 203, 101, 303, 403, false, fixedTime3, "Мария", "ООО Хвосты", "Fullstack Developer")
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:                  "Успешное получение пустого списка уведомлений",
			userID:                102,
			expectedNotifications: nil,
			expectedErr:           nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				})
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:                  "Успешное получение уведомлений с NULL значениями в JOIN полях",
			userID:                103,
			expectedNotifications: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании результата запроса GetDownloadResumeNotificationsForUser: %v", errors.New(`sql: Scan error on column index 8, name "applicant_name": converting NULL to string is unsupported`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				}).AddRow(
					4, "download_resume", 204, 103, 304, 404,
					false, fixedTime1, nil, nil, nil,
				)
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:                  "Успешное получение уведомлений с частично NULL значениями",
			userID:                104,
			expectedNotifications: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании результата запроса GetDownloadResumeNotificationsForUser: %v", errors.New(`sql: Scan error on column index 9, name "employer_name": converting NULL to string is unsupported`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				}).AddRow(
					5, "download_resume", 205, 104, 305, 405,
					true, fixedTime1, "Сергей", nil, "DevOps Engineer",
				)
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:   "Успешное получение уведомлений с нулевыми ID",
			userID: 0,
			expectedNotifications: []*entity.NotificationPreview{
				{
					ID:            6,
					Type:          entity.DownloadResumeType,
					SenderID:      0,
					ReceiverID:    0,
					ObjectID:      0,
					ResumeID:      0,
					IsViewed:      false,
					CreatedAt:     fixedTime1,
					ApplicantName: "Тест",
					EmployerName:  "Тест Компания",
					Title:         "Тест Профессия",
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				}).AddRow(
					6, "download_resume", 0, 0, 0, 0,
					false, fixedTime1, "Тест", "Тест Компания", "Тест Профессия",
				)
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:   "Успешное получение уведомлений с максимальными ID",
			userID: 2147483647,
			expectedNotifications: []*entity.NotificationPreview{
				{
					ID:            2147483647,
					Type:          entity.DownloadResumeType,
					SenderID:      2147483646,
					ReceiverID:    2147483647,
					ObjectID:      2147483645,
					ResumeID:      2147483644,
					IsViewed:      true,
					CreatedAt:     fixedTime1,
					ApplicantName: "Максимум",
					EmployerName:  "Максимум Компания",
					Title:         "Максимум Профессия",
				},
			},
			expectedErr: nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				}).AddRow(
					2147483647, "download_resume", 2147483646, 2147483647, 2147483645, 2147483644,
					true, fixedTime1, "Максимум", "Максимум Компания", "Максимум Профессия",
				)
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:                  "Ошибка базы данных при выполнении запроса",
			userID:                105,
			expectedNotifications: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении запроса GetDownloadResumeNotificationsForUser: %v", errors.New("database connection failed")),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				mock.ExpectQuery(query).WithArgs(userID).WillReturnError(errors.New("database connection failed"))
			},
		},
		{
			name:                  "Ошибка таймаута базы данных",
			userID:                106,
			expectedNotifications: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при выполнении запроса GetDownloadResumeNotificationsForUser: %v", errors.New("query timeout")),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				mock.ExpectQuery(query).WithArgs(userID).WillReturnError(errors.New("query timeout"))
			},
		},
		{
			name:                  "Ошибка сканирования строки результата",
			userID:                107,
			expectedNotifications: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании результата запроса GetDownloadResumeNotificationsForUser: %v", errors.New(`sql: Scan error on column index 0, name "id": converting driver.Value type string ("invalid_id") to a int: invalid syntax`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				}).AddRow(
					"invalid_id", "download_resume", 207, 107, 307, 407,
					false, fixedTime1, "Тест", "Тест Компания", "Тест Профессия",
				)
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:                  "Ошибка сканирования типа уведомления",
			userID:                108,
			expectedNotifications: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании результата запроса GetDownloadResumeNotificationsForUser: %v", errors.New(`sql: Scan error on column index 1, name "type": converting NULL to string is unsupported`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				}).AddRow(
					7, nil, 208, 108, 308, 408,
					false, fixedTime1, "Тест", "Тест Компания", "Тест Профессия",
				)
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:                  "Ошибка сканирования времени создания",
			userID:                109,
			expectedNotifications: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при сканировании результата запроса GetDownloadResumeNotificationsForUser: %v", errors.New(`sql: Scan error on column index 7, name "created_at": unsupported Scan, storing driver.Value type string into type *time.Time`)),
			),
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				}).AddRow(
					8, "download_resume", 209, 109, 309, 409,
					false, "invalid_time", "Тест", "Тест Компания", "Тест Профессия",
				)
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
		{
			name:                  "Отрицательный ID пользователя",
			userID:                -1,
			expectedNotifications: nil,
			expectedErr:           nil,
			setupMock: func(mock sqlmock.Sqlmock, userID int) {
				rows := sqlmock.NewRows([]string{
					"id", "type", "sender_id", "receiver_id", "object_id", "resume_id",
					"is_viewed", "created_at", "applicant_name", "employer_name", "title",
				})
				mock.ExpectQuery(query).WithArgs(userID).WillReturnRows(rows)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer func(db *sql.DB, mock sqlmock.Sqlmock) {
				mock.ExpectClose()
				err := db.Close()
				require.NoError(t, err)
			}(db, mock)

			tc.setupMock(mock, tc.userID)

			repo := &NotificationRepository{DB: db}
			ctx := context.Background()

			notifications, err := repo.GetDownloadResumeNotificationsForUser(ctx, tc.userID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Nil(t, notifications)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedNotifications, notifications)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
