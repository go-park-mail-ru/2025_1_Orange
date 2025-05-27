package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/repository/mock"
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestStaticService_UploadAvatar(t *testing.T) {
	t.Parallel()

	readFile := func(t *testing.T, path string) []byte {
		t.Helper()
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		return data
	}

	testCases := []struct {
		Name                string
		Input               func(t *testing.T) []byte
		ExpectedErr         error
		SetupStaticRepoMock func(repo *mock.MockStaticRepository)
	}{
		{
			Name: "Валидный файл",
			Input: func(t *testing.T) []byte {
				return readFile(t, "../../../assets/valid_picture.png")
			},
			ExpectedErr: nil,
			SetupStaticRepoMock: func(repo *mock.MockStaticRepository) {
				repo.EXPECT().
					UploadStatic(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).
					Return(
						1,
						"http://example.com/bucket/filename.png",
						nil,
					)
			},
		},
		{
			Name: "Невалидный файл (WebP формат)",
			Input: func(t *testing.T) []byte {
				return readFile(t, "../../../assets/invalid_format.webp")
			},
			ExpectedErr:         entity.NewError(entity.ErrBadRequest, fmt.Errorf("недопустимый формат файла")),
			SetupStaticRepoMock: func(repo *mock.MockStaticRepository) {},
		},
		{
			Name: "Неподходящий размер изображения",
			Input: func(t *testing.T) []byte {
				return readFile(t, "../../../assets/invalid_size.jpg")
			},
			ExpectedErr:         entity.NewError(entity.ErrBadRequest, fmt.Errorf("размер файла превышает 5MB")),
			SetupStaticRepoMock: func(repo *mock.MockStaticRepository) {},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStaticRepo := mock.NewMockStaticRepository(ctrl)
			tc.SetupStaticRepoMock(mockStaticRepo)

			staticService := NewStaticService(mockStaticRepo)

			ctx := context.Background()
			_, err := staticService.UploadStatic(ctx, tc.Input(t))

			if tc.ExpectedErr != nil {
				require.Error(t, err)
				var repoErr entity.Error
				require.ErrorAs(t, err, &repoErr)
				require.Equal(t, tc.ExpectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func validJPEG() []byte {
	m := image.NewRGBA(image.Rect(0, 0, 1, 1))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, m, nil); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func validPNG() []byte {
	m := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	var buf bytes.Buffer
	if err := png.Encode(&buf, m); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func TestStaticService_GetStatic(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name          string
		id            int
		repoPath      string
		repoError     error
		expectedPath  string
		expectedError error
	}{
		{
			name:         "Success",
			id:           1,
			repoPath:     "uploads/1.jpg",
			expectedPath: "uploads/1.jpg",
		},
		{
			name:          "Repo error",
			id:            2,
			repoError:     errors.New("get error"),
			expectedError: errors.New("get error"),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRepo := mock.NewMockStaticRepository(ctrl)
			mockRepo.EXPECT().GetStatic(gomock.Any(), tc.id).Return(tc.repoPath, tc.repoError)
			svc := NewStaticService(mockRepo)
			path, err := svc.GetStatic(context.Background(), tc.id)
			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedPath, path)
			}
		})
	}
}

func TestStaticService_DeleteStatic(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name      string
		id        int
		repoError error
	}{
		{
			name: "Success",
			id:   1,
		},
		{
			name:      "Error",
			id:        2,
			repoError: errors.New("delete error"),
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockRepo := mock.NewMockStaticRepository(ctrl)
			mockRepo.EXPECT().DeleteStatic(gomock.Any(), tc.id).Return(tc.repoError)
			svc := NewStaticService(mockRepo)
			err := svc.DeleteStatic(context.Background(), tc.id)
			if tc.repoError != nil {
				require.Error(t, err)
				require.Equal(t, tc.repoError.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// func TestStaticService_UploadStatic(t *testing.T) {
// 	t.Parallel()

// 	ctx := context.Background()

// 	validJPEG := []byte{
// 		0xFF, 0xD8, 0xFF, 0xE0, // JPEG header
// 		0x00, 0x10, 0x4A, 0x46,
// 		0x49, 0x46, 0x00, 0x01,
// 		0x01, 0x00, 0x00, 0x01,
// 		0x00, 0x01, 0x00, 0x00,
// 	}
// 	// Это JPEG-заголовок + пустое изображение (не валидное, но тип определяется как image/jpeg)

// 	oversizedData := make([]byte, 5<<20+1) // 5MB + 1 байт
// 	invalidContentType := []byte("this is not an image")
// 	invalidImage := []byte{0xFF, 0xD8, 0xFF, 0xFF} // Поврежденное JPEG

// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockRepo := mock.NewMockStaticRepository(ctrl)

// 	service := &StaticService{
// 		staticRepository: mockRepo,
// 	}

// 	tests := []struct {
// 		name       string
// 		data       []byte
// 		setupMock  func()
// 		expectErr  bool
// 		errMessage string
// 	}{
// 		{
// 			name:       "Размер файла превышает 5MB",
// 			data:       oversizedData,
// 			setupMock:  func() {},
// 			expectErr:  true,
// 			errMessage: "размер файла превышает 5MB",
// 		},
// 		{
// 			name:       "Недопустимый формат файла",
// 			data:       invalidContentType,
// 			setupMock:  func() {},
// 			expectErr:  true,
// 			errMessage: "недопустимый формат файла",
// 		},
// 		{
// 			name: "Невалидное изображение (image/jpeg)",
// 			data: invalidImage,
// 			setupMock: func() {
// 				// Здесь не вызывается mockRepo.UploadStatic
// 			},
// 			expectErr:  true,
// 			errMessage: "невалидное изображение",
// 		},
// 		{
// 			name: "Успешная загрузка",
// 			data: validJPEG,
// 			setupMock: func() {
// 				mockRepo.EXPECT().
// 					UploadStatic(gomock.Any(), gomock.Any(), "image/jpeg", validJPEG).
// 					Return(1, "/path/to/image.jpg", nil)
// 			},
// 			expectErr: false,
// 		},
// 		{
// 			name: "Ошибка репозитория при загрузке",
// 			data: validJPEG,
// 			setupMock: func() {
// 				mockRepo.EXPECT().
// 					UploadStatic(gomock.Any(), gomock.Any(), "image/jpeg", validJPEG).
// 					Return(0, "", fmt.Errorf("ошибка репозитория"))
// 			},
// 			expectErr:  true,
// 			errMessage: "ошибка репозитория",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.setupMock()

// 			resp, err := service.UploadStatic(ctx, tt.data)

// 			if tt.expectErr {
// 				require.Error(t, err)
// 				require.Contains(t, err.Error(), tt.errMessage)
// 				require.Nil(t, resp)
// 			} else {
// 				require.NoError(t, err)
// 				require.NotNil(t, resp)
// 				require.NotEmpty(t, resp.Path)
// 				require.NotZero(t, resp.ID)
// 			}
// 		})
// 	}
// }
