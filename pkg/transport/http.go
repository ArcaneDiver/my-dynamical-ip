package transport

import (
	"context"
	"encoding/json"

	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/gorilla/mux"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	"net/http"

	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"

	"my-dynamical-ip/pkg/ip"
)

func NewHttpHandler(
	endpoints ip.IpEndpoint,
	otTracer stdopentracing.Tracer,
	zipkinTracer *stdzipkin.Tracer,
	logger log.Logger,
) http.Handler {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	if zipkinTracer != nil {
		options = append(options, zipkin.HTTPServerTrace(zipkinTracer))
	}

	r := mux.NewRouter()

	r.Methods("GET").Path("/ip").Handler(httptransport.NewServer(
		endpoints.GetEndpoint,
		decodeGetRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "Get", logger)))...,
	))
	r.Methods("POST").Path("/ip").Handler(httptransport.NewServer(
		endpoints.StoreEndpoint,
		decodeStoreRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "Store", logger)))...,
	))

	return r
}

func decodeGetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req ip.GetIpRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeStoreRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req ip.StoreIpRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := json.NewEncoder(w).Encode(response)

	return err
}

func errorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.WriteHeader(err2code(err))
	json.NewEncoder(w).Encode(errorWrapper{Error: err.Error()})
}

func err2code(err error) int {
	switch err {
	default:
		return http.StatusInternalServerError
	}
}

type errorWrapper struct {
	Error string `json:"error"`
}
