package mapper

import (
	"fmt"
	"github.com/JCorpse96/core/data"
	"github.com/JCorpse96/core/data/expression"
	"github.com/JCorpse96/core/data/mapper/config"
	"github.com/JCorpse96/core/data/resolve"
	"github.com/JCorpse96/core/support/log"
)

type ExprMapperFactory struct {
	exprFactory   expression.Factory
	objectFactory expression.Factory
}

func NewFactory(resolver resolve.CompositeResolver) Factory {
	exprFactory := expression.NewFactory(resolver)
	objMapperFactory := NewObjectMapperFactory(exprFactory)
	return &ExprMapperFactory{exprFactory: exprFactory, objectFactory: objMapperFactory}
}

func (mf *ExprMapperFactory) NewMapper(mappings map[string]interface{}) (Mapper, error) {

	if len(mappings) == 0 {
		return nil, nil
	}

	exprMappings := make(map[string]expression.Expr)
	for key, value := range mappings {
		if value != nil {
			switch t := value.(type) {
			case string:
				if len(t) > 0 && t[0] == '=' {
					//it's an expression
					expr, err := mf.exprFactory.NewExpr(t[1:])
					if err != nil {
						return nil, fmt.Errorf("create expression for field [%s] error: %s", key, err.Error())
					}
					exprMappings[key] = expr
				} else {
					exprMappings[key] = expression.NewLiteralExpr(value)
				}
			default:
				if IsConditionalMapping(t) {
					ifElseMapper, err := createConditionalMapper(t, mf.exprFactory)
					if err != nil {
						return nil, fmt.Errorf("create condiitonal mapper for field [%s] error: %s", key, err.Error())
					}
					exprMappings[key] = ifElseMapper
				} else if mapping, ok := GetObjectMapping(t); ok {
					//Object mapping
					objectExpr, err := NewObjectMapperFactory(mf.exprFactory).(*ObjectMapperFactory).NewObjectMapper(mapping)
					if err != nil {
						return nil, fmt.Errorf("create object mapper for field [%s] error: %s", key, err.Error())
					}
					exprMappings[key] = objectExpr
				} else {
					exprMappings[key] = expression.NewLiteralExpr(value)
				}
			}
		}
	}

	if len(exprMappings) == 0 {
		return nil, nil
	}

	return &ExprMapper{mappings: exprMappings}, nil
}

type ExprMapper struct {
	mappings map[string]expression.Expr
}

func (m *ExprMapper) Apply(inputScope data.Scope) (map[string]interface{}, error) {
	output := make(map[string]interface{}, len(m.mappings))
	for key, expr := range m.mappings {
		val, err := expr.Eval(inputScope)
		if err != nil {
			if config.IsMappingIgnoreErrorsOn() {
				log.RootLogger().Warnf("expresson eval error; %s", err.Error())
				//Skip value set.
				continue
			}
			//todo add some context to error (consider adding String() to exprImpl)
			return nil, err
		}
		output[key] = val
	}

	return output, nil
}

func isExpr(value interface{}) bool {
	if strVal, ok := value.(string); ok && len(strVal) > 0 && (strVal[0] == '=') {
		return true
	}
	return false
}
