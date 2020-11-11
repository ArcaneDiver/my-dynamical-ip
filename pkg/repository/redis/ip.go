package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"my-dynamical-ip/pkg/ip"
)

type ipRepository struct {
	client *redis.Client
}

var ctx = context.Background()

func createRedisClient(uri string) (*redis.Client, error) {
	opt, err := redis.ParseURL(uri)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)


	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

func New(dbUri string) (ip.IpRepository, error) {
	r := &ipRepository{}
	client, err := createRedisClient(dbUri)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	r.client = client

	return r, nil
}

func (r *ipRepository) Get() (string, error) {
	ip, err := r.client.Get(ctx, "ip").Result()
	if err != nil {
		return "", errors.Wrap(err, "repository.Ip.Get")
	}

	return ip, nil
}

func (r *ipRepository) Store(ip string) error {
	if err := r.client.Set(ctx, "ip", ip, 0).Err(); err != nil {
		return errors.Wrap(err, "repository.Ip.Store")
	}

	return nil
}