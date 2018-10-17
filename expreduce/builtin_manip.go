package expreduce

import "github.com/corywalker/expreduce/pkg/expreduceapi"

func distribute(e expreduceapi.ExpressionInterface, built expreduceapi.ExpressionInterface, res expreduceapi.ExpressionInterface) {
	i := len(built.GetParts())
	if i >= len(e.GetParts()) {
		res.GetParts() = append(res.GetParts(), built)
		return
	}
	shouldDistribute := false
	partAsExpr, partIsExpr := e.GetParts()[i].(expreduceapi.ExpressionInterface)
	if partIsExpr {
		if hashEx(partAsExpr.GetParts()[0]) == hashEx(res.GetParts()[0]) {
			shouldDistribute = true
		}
	}
	if shouldDistribute {
		for _, dPart := range partAsExpr.GetParts()[1:] {
			builtCopy := built.ShallowCopy()
			builtCopy.appendEx(dPart)
			distribute(e, builtCopy, res)
		}
		return
	}
	built.GetParts() = append(built.GetParts(), e.GetParts()[i])
	distribute(e, built, res)
	return
}

func GetManipDefinitions() (defs []Definition) {
	defs = append(defs, Definition{Name: "Together"})
	defs = append(defs, Definition{Name: "Numerator"})
	defs = append(defs, Definition{Name: "Denominator"})
	defs = append(defs, Definition{
		Name: "Apart",
		// Not fully implemented.
		OmitDocumentation: true,
	})
	defs = append(defs, Definition{
		Name: "Distribute",
		legacyEvalFn: func(this expreduceapi.ExpressionInterface, es expreduceapi.EvalStateInterface) expreduceapi.Ex {
			if len(this.GetParts()) != 3 {
				return this
			}

			expr, isExpr := this.GetParts()[1].(expreduceapi.ExpressionInterface)
			if !isExpr {
				return this.GetParts()[1]
			}
			res := NewExpression([]expreduceapi.Ex{this.GetParts()[2]})
			firstBuilt := NewExpression([]expreduceapi.Ex{expr.GetParts()[0]})
			distribute(expr, firstBuilt, res)
			return res
		},
	})
	return
}
