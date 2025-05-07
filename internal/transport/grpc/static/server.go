package static

import (
	staticPROTO "ResuMatch/internal/transport/grpc/static/proto"
	"ResuMatch/internal/transport/grpc/utils"
	"ResuMatch/internal/usecase"
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GRPC struct {
	staticPROTO.UnimplementedStaticServiceServer
	staticUC usecase.Static
}

func NewGRPC(staticUC usecase.Static) *GRPC {
	return &GRPC{
		staticUC: staticUC,
	}
}

func (service *GRPC) UploadStatic(ctx context.Context, req *staticPROTO.UploadStaticRequest) (*staticPROTO.UploadStaticResponse, error) {
	staticDTO, err := service.staticUC.UploadStatic(ctx, req.Data)
	if err != nil {
		return nil, utils.ToGRPCError(err)
	}

	return &staticPROTO.UploadStaticResponse{
		Id:   uint64(staticDTO.ID),
		Path: staticDTO.Path,
	}, nil
}

func (service *GRPC) GetStatic(ctx context.Context, req *staticPROTO.FileID) (*staticPROTO.StaticURL, error) {
	staticPath, err := service.staticUC.GetStatic(ctx, int(req.Id))
	if err != nil {
		return nil, utils.ToGRPCError(err)
	}

	return &staticPROTO.StaticURL{
		Path: staticPath,
	}, nil
}

func (service *GRPC) DeleteStatic(ctx context.Context, req *staticPROTO.FileID) (*emptypb.Empty, error) {
	err := service.staticUC.DeleteStatic(ctx, int(req.Id))
	if err != nil {
		return nil, utils.ToGRPCError(err)
	}

	return &emptypb.Empty{}, nil
}
