package utils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRequestID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("Set and Get Request ID", func(t *testing.T) {
		t.Parallel()

		// Создаем мок контекста (хотя в данном случае можно использовать реальный)
		ctx := context.Background()
		expectedID := "test-request-id"

		// Устанавливаем request ID
		newCtx := SetRequestID(ctx, expectedID)

		// Получаем request ID обратно
		actualID := GetRequestID(newCtx)
		require.Equal(t, expectedID, actualID)
	})

	t.Run("Get Request ID from empty context", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		requestID := GetRequestID(ctx)
		require.Empty(t, requestID)
	})

	t.Run("Context with wrong value type", func(t *testing.T) {
		t.Parallel()

		// Создаем контекст с неправильным типом значения
		ctx := context.WithValue(context.Background(), ctxKeyRequestID{}, 123)

		requestID := GetRequestID(ctx)
		require.Empty(t, requestID)
	})
}
