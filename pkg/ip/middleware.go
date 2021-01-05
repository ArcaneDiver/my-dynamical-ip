package ip

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/opentracing/opentracing-go"
	"time"
)

type MiddlewareService func(service IpService) IpService
type MiddlewareRepository func (service IpRepository) IpRepository

func LoggingMiddlewareService(logger log.Logger) MiddlewareService {
	return func(service IpService) IpService {
		return &loggingMiddlewareService{
			logger: logger,
			next:   service,
		}
	}
}

func LogginMiddlewareEndpoint(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				logger.Log("transport_error", err, "took", time.Since(begin))
			}(time.Now())

			return next(ctx, request)
		}
	}
}

func LoggingMiddlewareRepository(logger log.Logger) MiddlewareRepository {
	return func(repo IpRepository) IpRepository {
		return &logginMiddlewareRepository{
			logger: logger,
			next:   repo,
		}
	}
}

func TracingMiddlwareRepository(tracer opentracing.Tracer) MiddlewareRepository {
	return func(service IpRepository) IpRepository {
		return &tracingMiddlewareRepository{
			tracer: tracer,
			next: service,
		}
	}
}

type loggingMiddlewareService struct {
	logger log.Logger
	next   IpService
}

func (lm *loggingMiddlewareService) Get(ctx context.Context) (ip string, err error) {
	defer func() {
		lm.logger.Log("service", "ip", "method", "Get", "result", ip, "err", err)
	}()

	return lm.next.Get(ctx)
}

func (lm *loggingMiddlewareService) Store(ctx context.Context, ip string) (err error) {
	defer func() {
		lm.logger.Log("service", "ip", "method", "Store", "ip", ip, "err", err)
	}()

	return lm.next.Store(ctx, ip)
}

type logginMiddlewareRepository struct {
	logger log.Logger
	next IpRepository
}

func (lm *logginMiddlewareRepository) Get(ctx context.Context) (ip string, err error) {
	defer func() {
		lm.logger.Log("repository", "ip", "method", "Get", "result", ip, "err", err)
	}()

	return lm.next.Get(ctx)
}

func (lm *logginMiddlewareRepository) Store(ctx context.Context, ip string) (err error) {
	defer func() {
		lm.logger.Log("service", "ip", "method", "Store", "ip", ip, "err", err)
	}()

	return lm.next.Store(ctx, ip)
}

type tracingMiddlewareRepository struct {
	tracer opentracing.Tracer
	next IpRepository
}

func (t *tracingMiddlewareRepository) Get(ctx context.Context) (string, error) {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		span = t.tracer.StartSpan("Get")
	} else {
		span = opentracing.StartSpan("repository", opentracing.ChildOf(span.Context()))
	}

	defer span.Finish()

	return t.next.Get(opentracing.ContextWithSpan(ctx, span))
}

func (t *tracingMiddlewareRepository) Store(ctx context.Context, ip string) error {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		span = t.tracer.StartSpan("Store")
	} else {
		span = opentracing.StartSpan("repository", opentracing.ChildOf(span.Context()))
	}

	defer span.Finish()

	return t.next.Store(opentracing.ContextWithSpan(ctx, span), ip)
}
