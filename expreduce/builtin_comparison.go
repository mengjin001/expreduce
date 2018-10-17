package expreduce

import (
	"sort"

	"github.com/corywalker/expreduce/pkg/expreduceapi"
)

type extremaFnType int

const (
	MaxFn extremaFnType = iota
	MinFn
)

func extremaFunction(this expreduceapi.ExpressionInterface, fnType extremaFnType, es expreduceapi.EvalStateInterface) expreduceapi.Ex {
	// Flatten nested lists into arguments.
	origHead := this.GetParts()[0]
	this.GetParts()[0] = S("List")
	dst := E(S("List"))
	flattenExpr(this, dst, 999999999, &es.CASLogger)
	// Previously I always set the pointer but it led to an endless
	// eval loop. I think evaluation might use the pointer to make a
	// "same" comparison.
	if !IsSameQ(this, dst, &es.CASLogger) {
		this = dst
		sort.Sort(this)
	}
	this.GetParts()[0] = origHead

	if len(this.GetParts()) == 1 {
		if fnType == MaxFn {
			return E(S("Times"), NewInt(-1), S("Infinity"))
		} else {
			return S("Infinity")
		}
	}
	if len(this.GetParts()) == 2 {
		return this.GetParts()[1]
	}
	var i int
	for i = 1; i < len(this.GetParts()); i++ {
		if !numberQ(this.GetParts()[i]) {
			break
		}
	}
	if fnType == MaxFn {
		i -= 1
		return NewExpression(append([]expreduceapi.Ex{this.GetParts()[0]}, this.GetParts()[i:]...))
	}
	if i == 1 {
		return this
	}
	return NewExpression(append(this.GetParts()[:2], this.GetParts()[i:]...))
}

func getCompSign(e expreduceapi.Ex) int {
	sym, isSym := e.(*Symbol)
	if !isSym {
		return -2
	}
	switch sym.Name {
	case "System`Less":
		return -1
	case "System`LessEqual":
		return -1
	case "System`Equal":
		return 0
	case "System`GreaterEqual":
		return 1
	case "System`Greater":
		return 1
	}
	return -2
}

func getComparisonDefinitions() (defs []Definition) {
	defs = append(defs, Definition{
		Name: "Equal",
		toString: func(this expreduceapi.ExpressionInterface, params expreduceapi.ToStringParams) (bool, string) {
			return ToStringInfixAdvanced(this.GetParts()[1:], " == ", "System`Equal", false, "", "", params)
		},
		legacyEvalFn: func(this expreduceapi.ExpressionInterface, es expreduceapi.EvalStateInterface) expreduceapi.Ex {
			if len(this.GetParts()) < 1 {
				return this
			}

			isequal := true
			for i := 2; i < len(this.GetParts()); i++ {
				var equalstr string = this.GetParts()[1].IsEqual(this.GetParts()[i])
				if equalstr == "EQUAL_UNK" {
					return this
				}
				isequal = isequal && (equalstr == "EQUAL_TRUE")
			}
			if isequal {
				return NewSymbol("System`True")
			}
			return NewSymbol("System`False")
		},
	})
	defs = append(defs, Definition{
		Name: "Unequal",
		toString: func(this expreduceapi.ExpressionInterface, params expreduceapi.ToStringParams) (bool, string) {
			return ToStringInfixAdvanced(this.GetParts()[1:], " != ", "System`Unequal", false, "", "", params)
		},
		legacyEvalFn: func(this expreduceapi.ExpressionInterface, es expreduceapi.EvalStateInterface) expreduceapi.Ex {
			if len(this.GetParts()) != 3 {
				return this
			}

			var isequal string = this.GetParts()[1].IsEqual(this.GetParts()[2])
			if isequal == "EQUAL_UNK" {
				return this
			} else if isequal == "EQUAL_TRUE" {
				return NewSymbol("System`False")
			} else if isequal == "EQUAL_FALSE" {
				return NewSymbol("System`True")
			}

			return NewExpression([]expreduceapi.Ex{NewSymbol("System`Error"), NewString("Unexpected equality return value.")})
		},
	})
	defs = append(defs, Definition{
		Name: "SameQ",
		toString: func(this expreduceapi.ExpressionInterface, params expreduceapi.ToStringParams) (bool, string) {
			return ToStringInfixAdvanced(this.GetParts()[1:], " === ", "System`SameQ", false, "", "", params)
		},
		legacyEvalFn: func(this expreduceapi.ExpressionInterface, es expreduceapi.EvalStateInterface) expreduceapi.Ex {
			if len(this.GetParts()) < 1 {
				return this
			}

			issame := true
			for i := 2; i < len(this.GetParts()); i++ {
				issame = issame && IsSameQ(this.GetParts()[1], this.GetParts()[i], &es.CASLogger)
			}
			if issame {
				return NewSymbol("System`True")
			} else {
				return NewSymbol("System`False")
			}
		},
	})
	defs = append(defs, Definition{
		Name: "UnsameQ",
		toString: func(this expreduceapi.ExpressionInterface, params expreduceapi.ToStringParams) (bool, string) {
			return ToStringInfixAdvanced(this.GetParts()[1:], " =!= ", "System`UnsameQ", false, "", "", params)
		},
		legacyEvalFn: func(this expreduceapi.ExpressionInterface, es expreduceapi.EvalStateInterface) expreduceapi.Ex {
			if len(this.GetParts()) < 1 {
				return this
			}

			for i := 1; i < len(this.GetParts()); i++ {
				for j := i + 1; j < len(this.GetParts()); j++ {
					if IsSameQ(this.GetParts()[i], this.GetParts()[j], &es.CASLogger) {
						return NewSymbol("System`False")
					}
				}
			}
			return NewSymbol("System`True")
		},
	})
	defs = append(defs, Definition{
		Name: "AtomQ",
		legacyEvalFn: func(this expreduceapi.ExpressionInterface, es expreduceapi.EvalStateInterface) expreduceapi.Ex {
			if len(this.GetParts()) != 2 {
				return this
			}

			_, IsExpr := this.GetParts()[1].(expreduceapi.ExpressionInterface)
			if IsExpr {
				return NewSymbol("System`False")
			}
			return NewSymbol("System`True")
		},
	})
	defs = append(defs, Definition{
		Name:         "NumberQ",
		legacyEvalFn: singleParamQEval(numberQ),
	})
	defs = append(defs, Definition{
		Name: "NumericQ",
	})
	defs = append(defs, Definition{
		Name: "Less",
		toString: func(this expreduceapi.ExpressionInterface, params expreduceapi.ToStringParams) (bool, string) {
			return ToStringInfixAdvanced(this.GetParts()[1:], " < ", "", true, "", "", params)
		},
		legacyEvalFn: func(this expreduceapi.ExpressionInterface, es expreduceapi.EvalStateInterface) expreduceapi.Ex {
			if len(this.GetParts()) != 3 {
				return this
			}

			a := NewExpression([]expreduceapi.Ex{NewSymbol("System`N"), this.GetParts()[1]}).Eval(es)
			b := NewExpression([]expreduceapi.Ex{NewSymbol("System`N"), this.GetParts()[2]}).Eval(es)

			if !numberQ(a) || !numberQ(b) {
				return this
			}

			// Less
			if ExOrder(a, b) == 1 {
				return NewSymbol("System`True")
			}
			return NewSymbol("System`False")
		},
	})
	defs = append(defs, Definition{
		Name: "Greater",
		toString: func(this expreduceapi.ExpressionInterface, params expreduceapi.ToStringParams) (bool, string) {
			return ToStringInfixAdvanced(this.GetParts()[1:], " > ", "", true, "", "", params)
		},
		legacyEvalFn: func(this expreduceapi.ExpressionInterface, es expreduceapi.EvalStateInterface) expreduceapi.Ex {
			if len(this.GetParts()) != 3 {
				return this
			}

			a := NewExpression([]expreduceapi.Ex{NewSymbol("System`N"), this.GetParts()[1]}).Eval(es)
			b := NewExpression([]expreduceapi.Ex{NewSymbol("System`N"), this.GetParts()[2]}).Eval(es)

			if !numberQ(a) || !numberQ(b) {
				return this
			}
			// Greater
			if ExOrder(a, b) == -1 {
				return NewSymbol("System`True")
			}
			return NewSymbol("System`False")
		},
	})
	defs = append(defs, Definition{
		Name: "LessEqual",
		toString: func(this expreduceapi.ExpressionInterface, params expreduceapi.ToStringParams) (bool, string) {
			return ToStringInfixAdvanced(this.GetParts()[1:], " <= ", "", true, "", "", params)
		},
		legacyEvalFn: func(this expreduceapi.ExpressionInterface, es expreduceapi.EvalStateInterface) expreduceapi.Ex {
			if len(this.GetParts()) != 3 {
				return this
			}

			a := NewExpression([]expreduceapi.Ex{NewSymbol("System`N"), this.GetParts()[1]}).Eval(es)
			b := NewExpression([]expreduceapi.Ex{NewSymbol("System`N"), this.GetParts()[2]}).Eval(es)

			if !numberQ(a) || !numberQ(b) {
				return this
			}
			// Less
			if ExOrder(a, b) == 1 {
				return NewSymbol("System`True")
			}
			// Equal
			if ExOrder(a, b) == 0 {
				return NewSymbol("System`True")
			}
			return NewSymbol("System`False")
		},
	})
	defs = append(defs, Definition{
		Name: "GreaterEqual",
		toString: func(this expreduceapi.ExpressionInterface, params expreduceapi.ToStringParams) (bool, string) {
			return ToStringInfixAdvanced(this.GetParts()[1:], " >= ", "", true, "", "", params)
		},
		legacyEvalFn: func(this expreduceapi.ExpressionInterface, es expreduceapi.EvalStateInterface) expreduceapi.Ex {
			if len(this.GetParts()) != 3 {
				return this
			}

			a := NewExpression([]expreduceapi.Ex{NewSymbol("System`N"), this.GetParts()[1]}).Eval(es)
			b := NewExpression([]expreduceapi.Ex{NewSymbol("System`N"), this.GetParts()[2]}).Eval(es)

			if !numberQ(a) || !numberQ(b) {
				return this
			}
			// Greater
			if ExOrder(a, b) == -1 {
				return NewSymbol("System`True")
			}
			// Equal
			if ExOrder(a, b) == 0 {
				return NewSymbol("System`True")
			}
			return NewSymbol("System`False")
		},
	})
	defs = append(defs, Definition{
		Name: "Positive",
	})
	defs = append(defs, Definition{
		Name: "Negative",
	})
	defs = append(defs, Definition{
		Name: "Max",
		legacyEvalFn: func(this expreduceapi.ExpressionInterface, es expreduceapi.EvalStateInterface) expreduceapi.Ex {
			return extremaFunction(this, MaxFn, es)
		},
	})
	defs = append(defs, Definition{
		Name: "Min",
		legacyEvalFn: func(this expreduceapi.ExpressionInterface, es expreduceapi.EvalStateInterface) expreduceapi.Ex {
			return extremaFunction(this, MinFn, es)
		},
	})
	defs = append(defs, Definition{Name: "PossibleZeroQ"})
	defs = append(defs, Definition{Name: "MinMax"})
	defs = append(defs, Definition{Name: "Element"})
	defs = append(defs, Definition{
		Name: "Inequality",
		legacyEvalFn: func(this expreduceapi.ExpressionInterface, es expreduceapi.EvalStateInterface) expreduceapi.Ex {
			if len(this.GetParts()) == 1 {
				return this
			}
			if len(this.GetParts()) == 2 {
				return S("True")
			}
			if len(this.GetParts())%2 != 0 {
				return this
			}
			firstSign := getCompSign(this.GetParts()[2])
			if firstSign == -2 {
				return this
			}
			if firstSign != 0 {
				for i := 4; i < len(this.GetParts()); i += 2 {
					thisSign := getCompSign(this.GetParts()[i])
					if thisSign == -2 {
						return this
					}
					if thisSign == -firstSign {
						firstIneq := E(S("Inequality"))
						secondIneq := E(S("Inequality"))
						for j := 1; j < len(this.GetParts()); j++ {
							if j < i {
								firstIneq.appendEx(this.GetParts()[j])
							}
							if j > (i - 2) {
								secondIneq.appendEx(this.GetParts()[j])
							}
						}
						return E(S("And"), firstIneq, secondIneq)
					}
				}
			}
			res := E(S("Inequality"))
			for i := 0; i < (len(this.GetParts())-1)/2; i++ {
				lhs := this.GetParts()[2*i+1]
				if len(res.GetParts()) > 1 {
					lhs = res.GetParts()[len(res.GetParts())-1]
				}
				op := this.GetParts()[2*i+2]
				rhs := this.GetParts()[2*i+3]
				for rhsI := 2*i + 3; rhsI < len(this.GetParts()); rhsI += 2 {
					if falseQ(E(op, lhs, this.GetParts()[rhsI]).Eval(es), &es.CASLogger) {
						return S("False")
					}
				}
				evalRes := E(op, lhs, rhs).Eval(es)
				if !trueQ(evalRes, &es.CASLogger) {
					if !IsSameQ(res.GetParts()[len(res.GetParts())-1], lhs, &es.CASLogger) {
						res.appendEx(lhs)
					}
					res.appendEx(op)
					res.appendEx(rhs)
				}
			}
			if len(res.GetParts()) == 1 {
				return S("True")
			}
			if len(res.GetParts()) == 4 {
				return E(res.GetParts()[2], res.GetParts()[1], res.GetParts()[3])
			}
			return res
		},
	})
	return
}
