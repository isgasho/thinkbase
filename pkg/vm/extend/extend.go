package extend

import (
	"fmt"

	"github.com/deepfabric/thinkbase/pkg/vm/extend/overload"
	"github.com/deepfabric/thinkbase/pkg/vm/util"
	"github.com/deepfabric/thinkbase/pkg/vm/value"
)

func (e *UnaryExtend) IsLogical() bool {
	return overload.IsLogical(e.Op)
}

func (e *UnaryExtend) Attributes() []string {
	return e.E.Attributes()
}

func (e *UnaryExtend) Eval(mp map[string]value.Value) (value.Value, error) {
	v, err := e.E.Eval(mp)
	if err != nil {
		return nil, err
	}
	return overload.UnaryEval(e.Op, v)
}

func (e *UnaryExtend) String() string {
	switch e.Op {
	case overload.Not:
		return fmt.Sprintf("not %s", e.E.String())
	case overload.Abs:
		return fmt.Sprintf("abs(%s)", e.E.String())
	case overload.Ceil:
		return fmt.Sprintf("ceil(%s)", e.E.String())
	case overload.Sign:
		return fmt.Sprintf("sign(%s)", e.E.String())
	case overload.Floor:
		return fmt.Sprintf("floor(%s)", e.E.String())
	case overload.Lower:
		return fmt.Sprintf("lower(%s)", e.E.String())
	case overload.Round:
		return fmt.Sprintf("round(%s)", e.E.String())
	case overload.Upper:
		return fmt.Sprintf("upper(%s)", e.E.String())
	case overload.Length:
		return fmt.Sprintf("length(%s)", e.E.String())
	case overload.Typeof:
		return fmt.Sprintf("typeof(%s)", e.E.String())
	case overload.UnaryMinus:
		return fmt.Sprintf("-%s", e.E.String())
	}
	return ""
}

func (e *BinaryExtend) IsLogical() bool {
	return overload.IsLogical(e.Op)
}

func (e *BinaryExtend) Attributes() []string {
	return util.MergeAttributes(e.Left.Attributes(), e.Right.Attributes())
}

func (e *BinaryExtend) Eval(mp map[string]value.Value) (value.Value, error) {
	l, err := e.Left.Eval(mp)
	if err != nil {
		return nil, err
	}
	r, err := e.Right.Eval(mp)
	if err != nil {
		return nil, err
	}
	return overload.BinaryEval(e.Op, l, r)
}

func (e *BinaryExtend) String() string {
	switch e.Op {
	case overload.EQ:
		return fmt.Sprintf("%s = %s", e.Left.String(), e.Right.String())
	case overload.LT:
		return fmt.Sprintf("%s < %s", e.Left.String(), e.Right.String())
	case overload.GT:
		return fmt.Sprintf("%s > %s", e.Left.String(), e.Right.String())
	case overload.LE:
		return fmt.Sprintf("%s <= %s", e.Left.String(), e.Right.String())
	case overload.GE:
		return fmt.Sprintf("%s >= %s", e.Left.String(), e.Right.String())
	case overload.NE:
		return fmt.Sprintf("%s <> %s", e.Left.String(), e.Right.String())
	case overload.Or:
		return fmt.Sprintf("%s or %s", e.Left.String(), e.Right.String())
	case overload.And:
		return fmt.Sprintf("%s and %s", e.Left.String(), e.Right.String())
	case overload.Div:
		return fmt.Sprintf("%s / %s", e.Left.String(), e.Right.String())
	case overload.Mod:
		return fmt.Sprintf("%s %% %s", e.Left.String(), e.Right.String())
	case overload.Plus:
		return fmt.Sprintf("%s + %s", e.Left.String(), e.Right.String())
	case overload.Mult:
		return fmt.Sprintf("%s * %s", e.Left.String(), e.Right.String())
	case overload.Minus:
		return fmt.Sprintf("%s - %s", e.Left.String(), e.Right.String())
	case overload.Typecast:
		return fmt.Sprintf("cast(%s as %s)", e.Left.String(), e.Right.String())
	}
	return ""
}

func (e *MultiExtend) IsLogical() bool {
	return overload.IsLogical(e.Op)
}

func (e *MultiExtend) Attributes() []string {
	var rs []string

	mp := make(map[string]struct{})
	for _, arg := range e.Args {
		attrs := arg.Attributes()
		for i, j := 0, len(attrs); i < j; i++ {
			if _, ok := mp[attrs[i]]; !ok {
				mp[attrs[i]] = struct{}{}
				rs = append(rs, attrs[i])
			}
		}
	}
	return rs
}

func (e *MultiExtend) Eval(mp map[string]value.Value) (value.Value, error) {
	var args []value.Value

	for _, v := range e.Args {
		arg, err := v.Eval(mp)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
	return overload.MultiEval(e.Op, args)
}

func (e *MultiExtend) String() string {
	switch e.Op {
	case overload.Concat:
		var r string
		for i, arg := range e.Args {
			if i == 0 {
				r += fmt.Sprintf("%s", arg)
			} else {
				r += fmt.Sprintf(" ++ %s", arg)
			}
		}
		return r
	}
	return ""
}

func (a *Attribute) IsLogical() bool {
	return false
}

func (a *Attribute) Attributes() []string {
	return []string{a.Name}
}

func (a *Attribute) Eval(mp map[string]value.Value) (value.Value, error) {
	if v, ok := mp[a.Name]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("attribute '%s' not exist", a.Name)
}

func (a *Attribute) String() string {
	return a.Name
}
