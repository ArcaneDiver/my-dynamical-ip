package main

import (
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"

	"my-dynamical-ip/pkg/ip"
	redisRepo "my-dynamical-ip/pkg/repository/redis"
	"my-dynamical-ip/pkg/transport"

	stdopentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"

	"github.com/go-kit/kit/log"
)

func main() {
	fs := flag.NewFlagSet("IpService", flag.ExitOnError)

	var (
		httpAddr       = fs.String("http-addr", ":8081", "HTTP listen address")
		zipkinURL      = fs.String("zipkin-url", "http://localhost:9411/api/v2/spans", "Enable Zipkin tracing via HTTP reporter URL e.g. http://localhost:9411/api/v2/spans")
		zipkinBridge   = fs.Bool("zipkin-ot-bridge", true, "Use Zipkin OpenTracing bridge instead of native implementation")
	)

	fmt.Println(*zipkinURL)

	fs.Usage = usageFor(fs, os.Args[0]+" [flags]")
	fs.Parse(os.Args[1:])


	godotenv.Load()

	var (
		DB_URI = os.Getenv("DB_URI")
	)

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var zipkinTracer *zipkin.Tracer
	{
		if *zipkinURL != "" {
			var (
				err         error
				hostPort    = "localhost:8081"
				serviceName = "ipsvc"
				reporter    = zipkinhttp.NewReporter(*zipkinURL)
			)

			defer reporter.Close()

			zEP, _ := zipkin.NewEndpoint(serviceName, hostPort)
			zipkinTracer, err = zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zEP))
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}

			if !(*zipkinBridge) {
				logger.Log("tracer", "Zipkin", "type", "Native", "URL", *zipkinURL)
			}
		}
	}

	var tracer stdopentracing.Tracer
	{
		if *zipkinBridge && zipkinTracer != nil {
			logger.Log("tracer", "Zipkin", "type", "OpenTracing", "URL", *zipkinURL)
			tracer = zipkinot.Wrap(zipkinTracer)
			zipkinTracer = nil
		} else {
			logger.Log("tracer", "OpenTracing")
			tracer = stdopentracing.GlobalTracer()
		}
	}

	repo, err := redisRepo.New(DB_URI, logger)
	if err != nil {
		logger.Log("repository", "redis", "err", err)
		os.Exit(1)
	}
	repo = ip.LoggingMiddlewareRepository(logger)(repo)
	repo = ip.TracingMiddlwareRepository(tracer)(repo)

	var (
		service = ip.NewService(logger, repo)
		endpoints = ip.NewEndpoints(service, logger, tracer, zipkinTracer)
		httpHandler = transport.NewHttpHandler(endpoints, tracer, zipkinTracer, logger)
	)

	errs := make(chan error)

	go func() {
		signs := make(chan os.Signal)
		signal.Notify(signs, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-signs)
	}()

	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errs <- http.ListenAndServe(*httpAddr, httpHandler)
	}()

	logger.Log("Exit", <-errs)
}


func usageFor(fs *flag.FlagSet, short string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  %s\n", short)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")
		w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		w.Flush()
		fmt.Fprintf(os.Stderr, "\n")
	}
}