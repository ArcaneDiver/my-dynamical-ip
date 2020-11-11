package json

import (
	"encoding/json"
	"github.com/pkg/errors"
)

type ipSerialized struct {
	Ip string `json:"ip"`
}

type Serializer struct {}

func (s *Serializer) Encode(ip string) ([]byte, error)  {
	 payload := &ipSerialized{}
	 payload.Ip = ip

	 encoded, err := json.Marshal(payload)
	 if err != nil {
	 	return nil, errors.Wrap(err, "serializer.Ip.Encode")
	 }

	 return encoded, nil
}

func (s *Serializer) Decode(input []byte) (string, error) {
	decoded := &ipSerialized{}

	if err := json.Unmarshal(input, decoded); err != nil {
		return "", errors.Wrap(err, "serializer.Ip.Decode")
	}

	return decoded.Ip, nil
}