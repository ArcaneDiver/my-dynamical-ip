package ip

import "context"

type IpRepository interface {
	Get(ctx context.Context) (string, error)
	Store(ctx context.Context, ip string) error
}