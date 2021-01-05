package ip

import (
	"context"
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

type IpService interface {
	Get(ctx context.Context) (string, error)
	Store(ctx context.Context, ip string) error
}

func NewService(logger log.Logger, repo IpRepository) IpService {
	var svc IpService
	{
		svc = &ipService{
			repo,
		}
		svc = LoggingMiddlewareService(logger)(svc)
	}

	return svc
}

type ipService struct {
	ipRepo IpRepository
}

func (s *ipService) Get(ctx context.Context) (string, error) {
	ip, err := s.ipRepo.Get(ctx)
	if err != nil {
		return "", errors.Wrap(err, "service.Ip.Get")
	}

	return ip, nil
}

func (s *ipService) Store(ctx context.Context, ip string) error {
	if err := s.ipRepo.Store(ctx, ip); err != nil {
		return errors.Wrap(err, "service.Ip.Store")
	}

	return nil
}