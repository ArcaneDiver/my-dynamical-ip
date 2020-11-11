package ip

import (
	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"
)

type IpService interface {
	Get() (string, error)
	Store(ip string) error
}

func NewService(logger log.Logger) IpService {
	var svc IpService
	{
		svc = &ipService{}
	}

	return svc
}

type ipService struct {
	ipRepo IpRepository
}

func (s *ipService) Get() (string, error) {
	ip, err := s.ipRepo.Get()
	if err != nil {
		return "", errors.Wrap(err, "service.Ip.Get")
	}

	return ip, nil
}

func (s *ipService) Store(ip string) error {
	if err := s.ipRepo.Store(ip); err != nil {
		return errors.Wrap(err, "service.Ip.Store")
	}

	return nil
}