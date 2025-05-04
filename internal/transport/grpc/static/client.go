package static

import (
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/metrics"
	"ResuMatch/internal/transport/grpc/interceptors"
	staticPROTO "ResuMatch/internal/transport/grpc/static/proto"
	"context"
	"github.com/prometheus/client_golang/prometheus"
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
	timer := prometheus.NewTimer(metrics.StaticServiceCallDuration.WithLabelValues("UploadStatic"))
	defer timer.ObserveDuration()

	resp, err := gw.staticClient.UploadStatic(ctx, &staticPROTO.UploadStaticRequest{Data: data})
	if err != nil {
		metrics.StaticServiceCallCounter.WithLabelValues("UploadStatic", "500").Inc()
		return nil, err
	}

	staticDTO := &dto.UploadStaticResponse{
		ID:   int(resp.Id),
		Path: resp.Path,
	}

	metrics.StaticServiceCallCounter.WithLabelValues("UploadStatic", "200").Inc()
	return staticDTO, nil
}

func (gw *Gateway) GetStatic(ctx context.Context, id int) (string, error) {
	timer := prometheus.NewTimer(metrics.StaticServiceCallDuration.WithLabelValues("GetStatic"))
	defer timer.ObserveDuration()

	resp, err := gw.staticClient.GetStatic(ctx, &staticPROTO.FileID{Id: uint64(id)})
	if err != nil {
		metrics.StaticServiceCallCounter.WithLabelValues("GetStatic", "500").Inc()
		return "", err
	}

	metrics.StaticServiceCallCounter.WithLabelValues("GetStatic", "200").Inc()
	return resp.Path, err

}

func (gw *Gateway) DeleteStatic(ctx context.Context, id int) error {
	timer := prometheus.NewTimer(metrics.StaticServiceCallDuration.WithLabelValues("DeleteStatic"))
	defer timer.ObserveDuration()

	_, err := gw.staticClient.DeleteStatic(ctx, &staticPROTO.FileID{Id: uint64(id)})
	if err != nil {
		metrics.StaticServiceCallCounter.WithLabelValues("DeleteStatic", "500").Inc()
		return err
	}

	metrics.StaticServiceCallCounter.WithLabelValues("DeleteStatic", "200").Inc()
	return nil
}
