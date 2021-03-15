package crd

import (
	"encoding/json"
	"fmt"

	rest "k8s.io/client-go/rest"
)

// Decoder interface
type OpenGaussDecoderInterface interface {
	Decode() OpenGaussConfiguration
}

type OpenGaussDecoder struct {
	r *rest.Result
}

// DecodeList() return a OpenGaussListConfiguration
func (decoder *OpenGaussDecoder) DecodeList(s *string) OpenGaussListConfiguration {
	res := OpenGaussListConfiguration{}
	json.Unmarshal([]byte(*s), &res)
	fmt.Println(res)
	return res
}

// Decode() return a OpenGaussConfiguration
func (decoder *OpenGaussDecoder) Decode(s *string) OpenGaussConfiguration {
	res := OpenGaussConfiguration{}
	json.Unmarshal([]byte(*s), &res)
	fmt.Println(res)
	return res
}
