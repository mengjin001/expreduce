package matcher

import (
	"bytes"
	"sort"
	"strings"

	"github.com/corywalker/expreduce/expreduce/atoms"
	"github.com/corywalker/expreduce/pkg/expreduceapi"
)

type PDManager struct {
	patternDefined map[string]expreduceapi.Ex
}

func EmptyPD() *PDManager {
	return &PDManager{nil}
}

func copyPD(orig *PDManager) (dest *PDManager) {
	dest = EmptyPD()
	// We do not care that this iterates in a random order.
	if (*orig).len() > 0 {
		dest.lazyMakeMap()
		for k, v := range (*orig).patternDefined {
			(*dest).patternDefined[k] = v
		}
	}
	return
}

func (this *PDManager) lazyMakeMap() {
	if this.patternDefined == nil {
		this.patternDefined = make(map[string]expreduceapi.Ex)
	}
}

func (this *PDManager) Define(name string, val expreduceapi.Ex) {
	this.lazyMakeMap()
	this.patternDefined[name] = val
}

func (this *PDManager) update(toAdd *PDManager) {
	if (*toAdd).len() > 0 {
		this.lazyMakeMap()
	}
	// We do not care that this iterates in a random order.
	for k, v := range (*toAdd).patternDefined {
		(*this).patternDefined[k] = v
	}
}

func (this *PDManager) len() int {
	if this.patternDefined == nil {
		return 0
	}
	return len(this.patternDefined)
}

func (this *PDManager) string(es expreduceapi.EvalStateInterface) string {
	var buffer bytes.Buffer
	buffer.WriteString("{")
	// We sort the keys here such that converting identical PDManagers always
	// produces the same string.
	keys := []string{}
	for k := range this.patternDefined {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := this.patternDefined[k]
		buffer.WriteString(k)
		buffer.WriteString("_: ")
		buffer.WriteString(v.String(es))
		buffer.WriteString(", ")
	}
	if strings.HasSuffix(buffer.String(), ", ") {
		buffer.Truncate(buffer.Len() - 2)
	}
	buffer.WriteString("}")
	return buffer.String()
}

func (this *PDManager) Expression() expreduceapi.Ex {
	res := atoms.NewExpression([]expreduceapi.Ex{atoms.NewSymbol("System`List")})
	// We sort the keys here such that converting identical PDManagers always
	// produces the same string.
	keys := []string{}
	for k := range this.patternDefined {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := this.patternDefined[k]
		res.AppendEx(atoms.NewExpression([]expreduceapi.Ex{
			atoms.NewSymbol("System`Rule"),
			atoms.NewString(k),
			v,
		}))
	}
	return res
}

func defineSequence(lhs parsedForm, sequence []expreduceapi.Ex, pm *PDManager, sequenceHead string, es expreduceapi.EvalStateInterface) bool {
	var attemptDefine expreduceapi.Ex = nil
	if lhs.hasPat {
		sequenceHeadSym := atoms.NewSymbol(sequenceHead)
		oneIdent := sequenceHeadSym.Attrs(es.GetDefinedMap()).OneIdentity
		if len(sequence) == 1 && (lhs.isBlank || oneIdent || lhs.isOptional) {
			attemptDefine = sequence[0]
		} else if len(sequence) == 0 && lhs.isOptional && lhs.defaultExpr != nil {
			attemptDefine = lhs.defaultExpr
		} else if lhs.isImpliedBs {
			attemptDefine = atoms.NewExpression(append([]expreduceapi.Ex{sequenceHeadSym}, sequence...))
		} else {
			head := atoms.NewSymbol("System`Sequence")
			attemptDefine = atoms.NewExpression(append([]expreduceapi.Ex{head}, sequence...))
		}

		if pm.patternDefined != nil {
			defined, ispd := pm.patternDefined[lhs.patSym.Name]
			if ispd && !atoms.IsSameQ(defined, attemptDefine, es.GetLogger()) {
				es.Debugf("patterns do not match! continuing.")
				return false
			}
		}
		pm.lazyMakeMap()
		pm.patternDefined[lhs.patSym.Name] = attemptDefine
	}
	return true
}