package transport

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type wrappedServerStream struct {
	stream      grpc.ServerStream
	reqID       string
	modifiedCtx context.Context
}

func (w *wrappedServerStream) SendMsg(m any) error {
	err := w.stream.SendMsg(m)

	if err != nil {
		statusCode := status.Code(err)
		logrus.WithFields(logrus.Fields{
			"requestID":  w.reqID,
			"statusCode": statusCode,
			"error":      err.Error(),
		}).Errorf("failed to send message in gRPC stream")
	}

	return err
}

func (w *wrappedServerStream) RecvMsg(m any) error {
	err := w.stream.RecvMsg(m)

	if err != nil {
		statusCode := status.Code(err)
		logrus.WithFields(logrus.Fields{
			"requestID":  w.reqID,
			"statusCode": statusCode,
			"error":      err.Error(),
		}).Errorf("failed to receive message in gRPC stream")
	}

	return err
}

func (w *wrappedServerStream) SetHeader(md metadata.MD) error {
	return w.stream.SetHeader(md)
}

func (w *wrappedServerStream) SendHeader(md metadata.MD) error {
	return w.stream.SendHeader(md)
}

func (w *wrappedServerStream) SetTrailer(md metadata.MD) {
	w.stream.SetTrailer(md)
}

func (w *wrappedServerStream) Context() context.Context {
	return w.modifiedCtx
}

func RequestIDInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	requestID := uuid.NewString()
	newCtx := context.WithValue(ctx, "requestID", requestID)

	return handler(newCtx, req)
}

func LoggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	reqID := ctx.Value("requestID")
	if reqID == nil {
		reqID = "no-request-id"
	}

	startTime := time.Now()
	logrus.WithField("requestID", reqID).Infof("received gRPC request: %s", info.FullMethod)

	resp, err := handler(ctx, req)

	duration := time.Since(startTime)
	statusCode := status.Code(err)

	logFields := logrus.Fields{
		"requestID":  reqID,
		"took":       duration,
		"method":     info.FullMethod,
		"statusCode": statusCode,
	}

	if err != nil {
		logFields["error"] = err.Error()
		logrus.WithFields(logFields).Errorf("gRPC request failed")
	} else {
		logrus.WithFields(logFields).Infof("gRPC request completed successfully")
	}

	return resp, err
}

func StreamRequestIDInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ss.Context()
	requestID := uuid.NewString()
	newCtx := context.WithValue(ctx, "requestID", requestID)

	wrappedStreamWithNewCtx := &wrappedServerStream{
		stream:      ss,
		reqID:       requestID,
		modifiedCtx: newCtx,
	}

	return handler(srv, wrappedStreamWithNewCtx)
}

func StreamLoggingInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	reqIDStr := "no-request-id"

	if ws, ok := ss.(*wrappedServerStream); ok {
		reqIDStr = ws.reqID
	} else {
		reqIDFromCtx := ss.Context().Value("requestID")
		if reqIDFromCtx != nil {
			reqIDStr = reqIDFromCtx.(string)
		}
	}

	startTime := time.Now()
	logrus.WithField("requestID", reqIDStr).Infof("received gRPC stream: %s", info.FullMethod)

	err := handler(srv, ss)

	duration := time.Since(startTime)
	statusCode := status.Code(err)

	logFields := logrus.Fields{
		"requestID":  reqIDStr,
		"took":       duration,
		"method":     info.FullMethod,
		"statusCode": statusCode,
	}

	if err != nil {
		logFields["error"] = err.Error()
		logrus.WithFields(logFields).Errorf("gRPC stream failed")
	} else {
		logrus.WithFields(logFields).Infof("gRPC stream completed successfully")
	}

	return err
}
