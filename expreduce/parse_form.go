package expreduce

import (
	"github.com/corywalker/expreduce/expreduce/logging"
	"github.com/corywalker/expreduce/pkg/expreduceapi"
)

type parsedForm struct {
	startI      int
	endI        int
	form        expreduceapi.Ex
	origForm    expreduceapi.Ex
	isBlank     bool
	isImpliedBs bool
	isOptional  bool
	defaultExpr expreduceapi.Ex
	hasPat      bool
	patSym      *Symbol
}

func ParseRepeated(e expreduceapi.ExpressionInterface) (expreduceapi.Ex, int, int, bool) {
	min, max := -1, -1
	if len(e.GetParts()) < 2 {
		return nil, min, max, false
	}
	if len(e.GetParts()) >= 3 {
		list, isList := HeadAssertion(e.GetParts()[2], "System`List")
		if !isList {
			return nil, min, max, false
		}
		if len(list.GetParts()) != 2 {
			return nil, min, max, false
		}
		i, isInt := list.GetParts()[1].(*Integer)
		if !isInt {
			return nil, min, max, false
		}
		ival := i.Val.Int64()
		min = int(ival)
		max = min
	}
	return e.GetParts()[1], min, max, true
}

func ParseForm(lhs_component expreduceapi.Ex, isFlat bool, sequenceHead string, headDefault expreduceapi.Ex, cl *logging.CASLogger) (res parsedForm) {
	// Calculate the min and max elements this component can match.
	toParse := lhs_component
	optional, isOptional := HeadAssertion(toParse, "System`Optional")
	if isOptional {
		toParse = optional.GetParts()[1]
	}
	patTest, isPatTest := HeadAssertion(toParse, "System`PatternTest")
	if isPatTest {
		toParse = patTest.GetParts()[1]
	}
	pat, isPat := HeadAssertion(toParse, "System`Pattern")
	var patSym *Symbol
	if isPat {
		patIsSym := false
		patSym, patIsSym = pat.GetParts()[1].(*Symbol)
		if patIsSym {
			toParse = pat.GetParts()[2]
		} else {
			// Valid patterns must have symbols to define.
			isPat = false
		}
	}
	bns, isBns := HeadAssertion(toParse, "System`BlankNullSequence")
	bs, isBs := HeadAssertion(toParse, "System`BlankSequence")
	blank, isBlank := HeadAssertion(toParse, "System`Blank")
	repeated, isRepeated := HeadAssertion(toParse, "System`Repeated")
	repeatedNull, isRepeatedNull := HeadAssertion(toParse, "System`RepeatedNull")
	isImpliedBs := isBlank && isFlat
	// Ensure isBlank is exclusive from isImpliedBs
	isBlank = isBlank && !isImpliedBs

	form := toParse
	startI := 1 // also includes implied blanksequence
	endI := 1
	var defaultExpr expreduceapi.Ex

	if isOptional {
		defaultToUse := headDefault
		if len(optional.GetParts()) >= 3 {
			defaultToUse = optional.GetParts()[2]
		}
		if len(optional.GetParts()) >= 2 {
			startI = 0
			if defaultToUse == nil {
				startI = 1
			}
			endI = 1
			// I think the !isPatTest part might be a hack.
			if isImpliedBs && !isPatTest {
				endI = MaxInt
			}
			//form = optional.Parts[1]
			defaultExpr = defaultToUse
		}
	} else if isBns {
		form = BlankNullSequenceToBlank(bns)
		startI = 0
		endI = MaxInt
	} else if isImpliedBs {
		form = blank
		endI = MaxInt
		if len(blank.GetParts()) >= 2 {
			sym, isSym := blank.GetParts()[1].(*Symbol)
			if isSym {
				// If we have a pattern like k__Plus
				if sym.Name == sequenceHead {
					form = NewExpression([]expreduceapi.Ex{NewSymbol("System`Blank")})
					startI = 2
				} else {
					endI = 1
				}
			}
		}
	} else if isBlank {
		form = blank
	} else if isRepeated {
		repPat, repMin, repMax, repOk := ParseRepeated(repeated)
		if repOk {
			if repMin != -1 {
				startI = repMin
			}
			if repMax != -1 {
				endI = repMax
			} else {
				// an undefined end can match to the end of the sequence.
				endI = MaxInt
			}
			form = repPat
		}
	} else if isRepeatedNull {
		if len(repeatedNull.GetParts()) == 2 {
			startI = 0
			endI = MaxInt
			form = repeatedNull.GetParts()[1]
		}
	} else if isBs {
		form = BlankSequenceToBlank(bs)
		endI = MaxInt
	}

	if isPatTest {
		form = NewExpression([]expreduceapi.Ex{patTest.GetParts()[0], form, patTest.GetParts()[2]})
	}

	res.startI = startI
	res.endI = endI
	res.form = form
	res.defaultExpr = defaultExpr
	res.origForm = lhs_component
	res.isImpliedBs = isImpliedBs
	res.isOptional = isOptional
	res.isBlank = isBlank
	if isPat {
		res.hasPat = true
		res.patSym = patSym
	}
	return res
}
