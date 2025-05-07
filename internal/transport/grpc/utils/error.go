package utils

import (
	"ResuMatch/internal/entity"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var clientErrorToGRPCCode = map[error]codes.Code{
	entity.ErrBadRequest:    codes.InvalidArgument,
	entity.ErrUnauthorized:  codes.Unauthenticated,
	entity.ErrForbidden:     codes.PermissionDenied,
	entity.ErrNotFound:      codes.NotFound,
	entity.ErrAlreadyExists: codes.AlreadyExists,
	entity.ErrInternal:      codes.Internal,
}

var grpcCodeToClientError = map[codes.Code]error{
	codes.InvalidArgument:  entity.ErrBadRequest,
	codes.Unauthenticated:  entity.ErrUnauthorized,
	codes.PermissionDenied: entity.ErrForbidden,
	codes.NotFound:         entity.ErrNotFound,
	codes.AlreadyExists:    entity.ErrAlreadyExists,
	codes.Internal:         entity.ErrInternal,
}

func ToGRPCError(err error) error {
	var svcErr entity.Error
	if errors.As(err, &svcErr) {
		code := codes.Internal
		if grpcCode, ok := clientErrorToGRPCCode[svcErr.ClientErr()]; ok {
			code = grpcCode
		}
		return status.Error(code, svcErr.InternalErr().Error())
	}
	return status.Error(codes.Internal, "internal server error")
}

func FromGRPCError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return entity.NewError(entity.ErrInternal, err)
	}

	clientErr, ok := grpcCodeToClientError[st.Code()]
	if !ok {
		clientErr = entity.ErrInternal
	}

	return entity.NewError(clientErr, errors.New(st.Message()))
}
