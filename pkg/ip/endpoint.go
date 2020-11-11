package ip

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/zipkin"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/tracing/opentracing"

	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	"github.com/sony/gobreaker"

	"golang.org/x/time/rate"
	"time"
)

type IpEndpoint struct {
	GetEndpoint endpoint.Endpoint
	StoreEndpoint endpoint.Endpoint
}

func NewEndpoints(s IpService, logger log.Logger, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) IpEndpoint {
	var getEndpoint endpoint.Endpoint
	{
		getEndpoint = makeGetEndpoint(s)
		getEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(getEndpoint)
		getEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(getEndpoint)
		getEndpoint = opentracing.TraceServer(otTracer, "Get")(getEndpoint)
		if zipkinTracer != nil {
			getEndpoint = zipkin.TraceEndpoint(zipkinTracer, "Get")(getEndpoint)
		}
	}

	var storeEndpoint endpoint.Endpoint
	{
		storeEndpoint = makeStoreEndpoint(s)
		storeEndpoint = ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 1))(storeEndpoint)
		storeEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(storeEndpoint)
		storeEndpoint = opentracing.TraceServer(otTracer, "Store")(storeEndpoint)
		if zipkinTracer != nil {
			storeEndpoint = zipkin.TraceEndpoint(zipkinTracer, "Store")(storeEndpoint)
		}
	}

	return IpEndpoint{
		GetEndpoint:   getEndpoint,
		StoreEndpoint: storeEndpoint,
	}
}

type GetIpRequest struct {

}
type GetIpResponse struct {
	Ip string
	Err error
}
func makeGetEndpoint(s IpService) endpoint.Endpoint {
	return func (ctx context.Context, request interface{}) (interface{}, error) {
		_ = request.(GetIpRequest)
		ip, err := s.Get()

		return GetIpResponse{Ip: ip, Err: err}, nil
	}
}

type StoreIpRequest struct {
	Ip string
}
type StoreIpResponse struct {
	Err error
}
func makeStoreEndpoint(s IpService) endpoint.Endpoint {
	return func (ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(StoreIpRequest)
		err := s.Store(req.Ip)

		return StoreIpResponse{Err: err}, nil
	}
}