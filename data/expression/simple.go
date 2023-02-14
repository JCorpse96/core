package expression

import (
	"github.com/JCorpse96/core/data"
	"github.com/JCorpse96/core/data/resolve"
)

func NewLiteralExpr(val interface{}) Expr {
	return &literalExpr{val: val}
}

type literalExpr struct {
	val interface{}
}

func (e *literalExpr) Eval(scope data.Scope) (interface{}, error) {
	return e.val, nil
}

type resolutionExpr struct {
	resolution resolve.Resolution
}

func (e *resolutionExpr) Eval(scope data.Scope) (interface{}, error) {

	return e.resolution.GetValue(scope)
}
