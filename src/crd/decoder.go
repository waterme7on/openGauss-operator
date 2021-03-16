package crd

import (
	"encoding/json"
)

// Decoder interface
type OpenGaussDecoderInterface interface {
	Decode() OpenGaussConfiguration
}

type OpenGaussDecoder struct {
}

// DecodeList() return a OpenGaussListConfiguration
func (decoder *OpenGaussDecoder) DecodeList(s *string) OpenGaussListConfiguration {
	res := OpenGaussListConfiguration{}
	json.Unmarshal([]byte(*s), &res)
	return res
}

// Decode() return a OpenGaussConfiguration
func (decoder *OpenGaussDecoder) Decode(s *string) OpenGaussConfiguration {
	res := OpenGaussConfiguration{}
	json.Unmarshal([]byte(*s), &res)
	return res
}
