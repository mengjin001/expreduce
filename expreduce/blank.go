package expreduce

import (
	"github.com/corywalker/expreduce/expreduce/logging"
	"github.com/corywalker/expreduce/pkg/expreduceapi"
)

func IsBlankTypeOnly(e expreduceapi.Ex) bool {
	asPattern, patternOk := HeadAssertion(e, "System`Pattern")
	if patternOk {
		_, blankOk := HeadAssertion(asPattern.GetParts()[2], "System`Blank")
		_, bsOk := HeadAssertion(asPattern.GetParts()[2], "System`BlankSequence")
		_, bnsOk := HeadAssertion(asPattern.GetParts()[2], "System`BlankNullSequence")
		if blankOk || bsOk || bnsOk {
			return true
		}
	}
	_, blankOk := HeadAssertion(e, "System`Blank")
	_, bsOk := HeadAssertion(e, "System`BlankSequence")
	_, bnsOk := HeadAssertion(e, "System`BlankNullSequence")
	if blankOk || bsOk || bnsOk {
		return true
	}
	return false
}

func IsBlankTypeCapturing(e expreduceapi.Ex, target expreduceapi.Ex, head expreduceapi.Ex, pm *PDManager, cl *logging.CASLogger) (bool, *PDManager) {
	// Similar to IsBlankType, but will capture target into es.patternDefined
	// if there is a valid match.
	asPattern, patternOk := HeadAssertion(e, "System`Pattern")
	if patternOk {
		asBlank, blankOk := HeadAssertion(asPattern.GetParts()[2], "System`Blank")
		asBS, bsOk := HeadAssertion(asPattern.GetParts()[2], "System`BlankSequence")
		asBNS, bnsOk := HeadAssertion(asPattern.GetParts()[2], "System`BlankNullSequence")
		if blankOk || bsOk || bnsOk {
			parts := []expreduceapi.Ex{}
			if blankOk {
				parts = asBlank.GetParts()
			} else if bsOk {
				parts = asBS.GetParts()
			} else if bnsOk {
				parts = asBNS.GetParts()
			}
			//if len(parts) < 2 {
			//return true, pm
			//}
			cl.Debugf("%v %v", parts, len(parts))
			matchesHead := false
			if len(parts) < 2 {
				matchesHead = true
			} else {
				matchesHead = IsSameQ(head, parts[1], cl)
			}
			cl.Debugf("%v", matchesHead)
			if matchesHead {
				sAsSymbol, sAsSymbolOk := asPattern.GetParts()[1].(*Symbol)
				if sAsSymbolOk {
					// TODO: we should handle matches with BlankSequences
					// differently here.
					toMatch, ispd := pm.patternDefined[sAsSymbol.Name]
					if !ispd {
						toMatch = target
						pm.LazyMakeMap()
						pm.patternDefined[sAsSymbol.Name] = target
					}
					if !IsSameQ(toMatch, target, cl) {
						return false, pm
					}
				}
				return true, pm
			}
			return false, pm
		}
	}
	asBlank, blankOk := HeadAssertion(e, "System`Blank")
	asBS, bsOk := HeadAssertion(e, "System`BlankSequence")
	asBNS, bnsOk := HeadAssertion(e, "System`BlankNullSequence")
	if blankOk || bsOk || bnsOk {
		parts := []expreduceapi.Ex{}
		if blankOk {
			parts = asBlank.GetParts()
		} else if bsOk {
			parts = asBS.GetParts()
		} else if bnsOk {
			parts = asBNS.GetParts()
		}
		if len(parts) < 2 {
			return true, pm
		}
		return IsSameQ(head, parts[1], cl), pm
	}
	return false, pm
}

func BlankNullSequenceToBlank(bns expreduceapi.ExpressionInterface) expreduceapi.ExpressionInterface {
	if len(bns.GetParts()) < 2 {
		return NewExpression([]expreduceapi.Ex{NewSymbol("System`Blank")})
	}
	return NewExpression([]expreduceapi.Ex{NewSymbol("System`Blank"), bns.GetParts()[1]})
}

func BlankSequenceToBlank(bs expreduceapi.ExpressionInterface) expreduceapi.ExpressionInterface {
	if len(bs.GetParts()) < 2 {
		return NewExpression([]expreduceapi.Ex{NewSymbol("System`Blank")})
	}
	return NewExpression([]expreduceapi.Ex{NewSymbol("System`Blank"), bs.GetParts()[1]})
}
