package utils

import "context"

type ctxKeyRequestID struct{}

func SetRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ctxKeyRequestID{}, requestID)
}

func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(ctxKeyRequestID{}).(string); ok {
		return requestID
	}
	return ""
}
