package ip


type IpSerializer interface {
	Encode(ip string) ([]byte, error)
	Decode(input []byte) (string, error)
}

