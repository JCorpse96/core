package function

import (
	"fmt"

	"github.com/JCorpse96/core/data"
	"github.com/JCorpse96/core/data/coerce"
)

type Function interface {
	Name() string
	Sig() (paramTypes []data.Type, isVariadic bool)
	Eval(params ...interface{}) (interface{}, error)
}

func Eval(f Function, params ...interface{}) (interface{}, error) {

	paramTypes, isVariadic := f.Sig()

	paramValues := make([]interface{}, len(params))
	typeIdx := 0

	if !isVariadic && len(params) != len(paramTypes) {
		return 0, fmt.Errorf("%s function should have %d arguments", f.Name(), len(paramTypes))
	}

	for idx, param := range params {
		if !isVariadic {
			typeIdx = idx
		}

		val, err := coerce.ToType(param, paramTypes[typeIdx])
		if err != nil {
			return nil, err
		}
		paramValues[idx] = val
	}

	return f.Eval(paramValues...)
}
