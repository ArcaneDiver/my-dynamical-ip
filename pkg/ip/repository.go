package ip


type IpRepository interface {
	Get() (string, error)
	Store(ip string) error
}