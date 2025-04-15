package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/repository/mock"
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"os"
	"testing"
	"time"
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
				expectedStatic := &entity.Static{
					ID:        1,
					FilePath:  "assets",
					FileName:  "generated_filename.png",
					CreatedAt: time.Date(2025, 4, 2, 12, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2025, 4, 2, 12, 0, 0, 0, time.UTC),
				}

				repo.EXPECT().
					UploadStatic(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).
					Return(expectedStatic, nil)
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
