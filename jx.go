package jxgen

import "github.com/go-faster/jx"

type Encoder interface {
	EncodeJSON(e *jx.Encoder) error
}

type Decoder interface {
	DecodeJSON(d *jx.Decoder) error
}
