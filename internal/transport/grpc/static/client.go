package static

import (
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/transport/grpc/interceptors"
	staticPROTO "ResuMatch/internal/transport/grpc/static/proto"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Gateway struct {
	staticClient staticPROTO.StaticServiceClient
}

func NewGateway(connectAddr string) (*Gateway, error) {
	grpcConn, err := grpc.NewClient(
		connectAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(interceptors.RequestIDClientInterceptor()),
	)
	if err != nil {
		return nil, err
	}

	staticClient := staticPROTO.NewStaticServiceClient(grpcConn)
	return &Gateway{staticClient: staticClient}, nil
}

func (gw *Gateway) UploadStatic(ctx context.Context, data []byte) (*dto.UploadStaticResponse, error) {
	resp, err := gw.staticClient.UploadStatic(ctx, &staticPROTO.UploadStaticRequest{Data: data})
	if err != nil {
		return nil, err
	}

	staticDTO := &dto.UploadStaticResponse{
		ID:   int(resp.Id),
		Path: resp.Path,
	}
	return staticDTO, nil
}

func (gw *Gateway) GetStatic(ctx context.Context, id int) (string, error) {
	resp, err := gw.staticClient.GetStatic(ctx, &staticPROTO.FileID{Id: uint64(id)})
	if err != nil {
		return "", err
	}
	return resp.Path, err

}

func (gw *Gateway) DeleteStatic(ctx context.Context, id int) error {
	_, err := gw.staticClient.DeleteStatic(ctx, &staticPROTO.FileID{Id: uint64(id)})
	if err != nil {
		return err
	}
	return nil
}
