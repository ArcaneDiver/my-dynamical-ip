package redis

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"my-dynamical-ip/pkg/ip"
	"net/url"
	"strings"
)

type ipRepository struct {
	client *redis.Client
}

var ctx = context.Background()

func createRedisClient(uri string, logger log.Logger) (*redis.Client, error) {
	var (
		password = ""
		parsedUrl *url.URL
	)

	if !strings.Contains(uri, "localhost") {
		parsedUrl, _ = url.Parse(uri)
		password, _ = parsedUrl.User.Password()
		uri = parsedUrl.Host
	}

	logger.Log("repository", "redis", "during", "initialization", "status", "init connection")
	client := redis.NewClient(&redis.Options{
		Addr: uri,
		Password: password,
		DB: 0,
	})


	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

func New(dbUri string, logger log.Logger) (ip.IpRepository, error) {
	r := &ipRepository{}
	client, err := createRedisClient(dbUri, logger)
	if err != nil {
		return nil, errors.Wrap(err, "repository.Redis")
	}

	r.client = client

	logger.Log("repository", "redis", "during", "initialization", "status", "ok")

	return r, nil
}

func (r *ipRepository) Get(ctx context.Context) (string, error) {
	ip, err := r.client.Get(ctx, "ip").Result()
	if err != nil {
		return "", errors.Wrap(err, "repository.Ip.Get")
	}

	return ip, nil
}

func (r *ipRepository) Store(ctx context.Context, ip string) error {
	if err := r.client.Set(ctx, "ip", ip, 0).Err(); err != nil {
		return errors.Wrap(err, "repository.Ip.Store")
	}

	return nil
}