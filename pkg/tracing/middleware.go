package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a gRPC unary server interceptor with tracing
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		tracer := otel.Tracer("grpc-server")

		// Extract trace context from metadata
		md, _ := metadata.FromIncomingContext(ctx)
		carrier := &metadataCarrier{md: md}
		ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

		// Start span
		ctx, span := tracer.Start(ctx, info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.method", info.FullMethod),
			),
		)
		defer span.End()

		// Call handler
		resp, err := handler(ctx, req)

		// Record error if any
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			st, _ := status.FromError(err)
			span.SetAttributes(
				attribute.String("rpc.grpc.status_code", st.Code().String()),
			)
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return resp, err
	}
}

// metadataCarrier adapts metadata.MD to propagation.TextMapCarrier
type metadataCarrier struct {
	md metadata.MD
}

func (m *metadataCarrier) Get(key string) string {
	values := m.md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (m *metadataCarrier) Set(key, value string) {
	m.md.Set(key, value)
}

func (m *metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(m.md))
	for k := range m.md {
		keys = append(keys, k)
	}
	return keys
}
