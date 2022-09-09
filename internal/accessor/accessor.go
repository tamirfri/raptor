/*
Copyright (c) 2022 RaptorML authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package accessor

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcZap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpcCtxTags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpcValidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/raptor-ml/raptor/api"
	"github.com/raptor-ml/raptor/pkg/sdk"
	coreApi "go.buf.build/raptor/api-go/raptor/core/raptor/core/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

type Accessor interface {
	GRPC(addr string) NoLeaderRunnableFunc
	HTTP(addr string, prefix string) NoLeaderRunnableFunc
}

type accessor struct {
	sdkServer coreApi.EngineServiceServer
	server    *grpc.Server
	logger    logr.Logger
}

func New(e api.FeatureManager, logger logr.Logger) Accessor {
	engine, ok := e.(api.Engine)
	if !ok {
		panic("e FeatureManager is not Engine")
	}
	svc := &accessor{
		sdkServer: sdk.NewServiceServer(engine),
		logger:    logger,
	}

	zapUnderlier, ok := svc.logger.GetSink().(zapr.Underlier)
	if !ok {
		panic("logr.LogSync does not implement Underlier interface")
	}
	zapLogger := zapUnderlier.GetUnderlying()

	grpcMetrics := grpcPrometheus.NewServerMetrics()
	metrics.Registry.MustRegister(grpcMetrics)

	svc.server = grpc.NewServer(
		grpc.StreamInterceptor(grpcMiddleware.ChainStreamServer(
			grpcCtxTags.StreamServerInterceptor(),
			grpcMetrics.StreamServerInterceptor(),
			grpcZap.StreamServerInterceptor(zapLogger),
			grpcValidator.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpcMiddleware.ChainUnaryServer(
			grpcCtxTags.UnaryServerInterceptor(),
			grpcMetrics.UnaryServerInterceptor(),
			grpcZap.UnaryServerInterceptor(zapLogger),
			grpcValidator.UnaryServerInterceptor(),
		)),
	)
	coreApi.RegisterEngineServiceServer(svc.server, svc.sdkServer)
	grpcMetrics.InitializeMetrics(svc.server)
	reflection.Register(svc.server)

	return svc
}

func (a *accessor) GRPC(addr string) NoLeaderRunnableFunc {
	return func(ctx context.Context) error {
		l, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to listen: %w", err)
		}

		a.logger.WithValues("kind", "grpc", "addr", l.Addr()).Info("Starting Accessor server")
		go func() {
			<-ctx.Done()
			a.server.Stop()
		}()
		return a.server.Serve(l)
	}
}

func (a *accessor) HTTP(addr string, prefix string) NoLeaderRunnableFunc {
	return func(ctx context.Context) error {
		gwMux := runtime.NewServeMux()
		err := coreApi.RegisterEngineServiceHandlerServer(ctx, gwMux, a.sdkServer)
		if err != nil {
			return fmt.Errorf("failed to register grpc gateway: %w", err)
		}

		if prefix[len(prefix)-1] == '/' {
			prefix += "/"
		}
		mux := http.NewServeMux()
		mux.Handle(prefix[:len(prefix)-1], http.StripPrefix(fmt.Sprintf("%s/", prefix), gwMux))

		a.logger.WithValues("kind", "http", "addr", addr).Info("Starting Accessor server")
		srv := http.Server{Handler: mux, Addr: addr}
		go func() {
			<-ctx.Done()
			_ = srv.Shutdown(ctx)
		}()
		return srv.ListenAndServe()
	}
}
