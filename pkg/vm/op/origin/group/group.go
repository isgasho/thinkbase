package group

import (
	"errors"
	"fmt"

	"github.com/deepfabric/thinkbase/pkg/vm/container/dictVec"
	"github.com/deepfabric/thinkbase/pkg/vm/context"
	"github.com/deepfabric/thinkbase/pkg/vm/extend"
	"github.com/deepfabric/thinkbase/pkg/vm/op"
	"github.com/deepfabric/thinkbase/pkg/vm/op/origin/summarize"
	"github.com/deepfabric/thinkbase/pkg/vm/op/origin/summarize/overload"
	"github.com/deepfabric/thinkbase/pkg/vm/op/origin/summarize/overload/avg"
	"github.com/deepfabric/thinkbase/pkg/vm/op/origin/summarize/overload/count"
	"github.com/deepfabric/thinkbase/pkg/vm/op/origin/summarize/overload/max"
	"github.com/deepfabric/thinkbase/pkg/vm/op/origin/summarize/overload/min"
	"github.com/deepfabric/thinkbase/pkg/vm/op/origin/summarize/overload/sum"
	"github.com/deepfabric/thinkbase/pkg/vm/util"
	"github.com/deepfabric/thinkbase/pkg/vm/util/encoding"
	"github.com/deepfabric/thinkbase/pkg/vm/value"
)

func New(prev op.OP, e extend.Extend, gs []string, es []*summarize.Extend, c context.Context) *group {
	return &group{isCheck: false, prev: prev, e: e, gs: gs, c: c, es: es}
}

func (n *group) Size() float64 {
	var ops []int

	for _, e := range n.es {
		ops = append(ops, e.Op)
	}
	return n.c.GroupSize(n.prev, n.gs, ops)
}

func (n *group) Cost() float64 {
	var ops []int

	for _, e := range n.es {
		ops = append(ops, e.Op)
	}
	return n.c.GroupCost(n.prev, n.gs, ops)
}

func (n *group) Dup() op.OP {
	return &group{
		c:       n.c,
		e:       n.e,
		es:      n.es,
		gs:      n.gs,
		prev:    n.prev,
		isCheck: n.isCheck,
	}
}

func (n *group) SetChild(o op.OP, _ int) { n.prev = o }
func (n *group) Operate() int            { return op.Group }
func (n *group) Children() []op.OP       { return []op.OP{n.prev} }
func (n *group) IsOrdered() bool         { return n.prev.IsOrdered() }

func (n *group) String() string {
	r := fmt.Sprintf("γ([")
	for i, g := range n.gs {
		switch i {
		case 0:
			r += fmt.Sprintf("%s", g)
		default:
			r += fmt.Sprintf(", %s", g)
		}
	}
	r += "], ["
	for i, e := range n.es {
		switch i {
		case 0:
			r += fmt.Sprintf("%s(%s) -> %s", overload.AggName[e.Op], e.Name, e.Alias)
		case 1:
			r += fmt.Sprintf(", %s(%s) -> %s", overload.AggName[e.Op], e.Name, e.Alias)
		}
	}
	r += fmt.Sprintf("], %v, %s)", n.e, n.prev)
	return r
}

func (n *group) Name() (string, error) {
	return n.prev.Name()
}

func (n *group) AttributeList() ([]string, error) {
	return aliasList(n.es, n.gs), nil
}

func (n *group) GetTuples(limit int) (value.Array, error) {
	attrs := attributeList(n.es, n.gs)
	if !n.isCheck {
		if err := n.check(attrs); err != nil {
			return nil, err
		}
		if err := n.newByAttributes(attrs); err != nil {
			n.dv.Destroy()
			return nil, err
		}
		n.isCheck = true
	}
	size := 0
	var a value.Array
	for {
		if size >= limit {
			break
		}
		if len(n.k) == 0 {
			k, err := n.dv.PopKey()
			if err != nil {
				n.dv.Destroy()
				return nil, err
			}
			if len(k) == 0 {
				n.dv.Destroy()
				return a, nil
			}
			n.k = k
		}
		ts, err := n.dv.Pops(n.k, -1, n.c.MemSize())
		switch {
		case err == dictVec.NotExist || (err == nil && len(ts) == 0):
			var t value.Array
			{
				v, _, err := encoding.DecodeValue([]byte(n.k))
				if err != nil {
					n.dv.Destroy()
					return nil, err
				}
				size += v.(value.Array).Size()
				t = append(t, v.(value.Array)...)
			}
			for _, e := range n.es {
				if v, err := e.Agg.Eval(); err != nil {
					n.dv.Destroy()
					return nil, err
				} else {
					t = append(t, v)
				}
				e.Agg.Reset()
			}
			if n.e != nil {
				if ok, err := n.e.Eval(util.Tuple2Map(t, aliasList(n.es, n.gs))); err != nil {
					return nil, err
				} else if value.MustBeBool(ok) {
					a = append(a, t)
					size += t.Size()
				}
			} else {
				a = append(a, t)
				size += t.Size()
			}
			n.k = ""
			continue
		case err != nil:
			n.dv.Destroy()
			return nil, err
		}
		mp := util.Tuples2Map(ts, attrs)
		for _, e := range n.es {
			if err := e.Agg.Fill(mp[e.Name]); err != nil {
				n.dv.Destroy()
				return nil, err
			}
		}
	}
	return a, nil
}

func (n *group) GetAttributes(attrs []string, limit int) (map[string]value.Array, error) {
	if !n.isCheck {
		if err := n.check(attributeList(n.es, n.gs)); err != nil {
			return nil, err
		}
		if err := util.Contain(attrs, aliasList(n.es, n.gs)); err != nil {
			return nil, err
		}
		if err := n.newByAttributes(attributeList(n.es, n.gs)); err != nil {
			n.dv.Destroy()
			return nil, err
		}
		n.isCheck = true
	}
	ts, err := n.GetTuples(limit)
	if err != nil {
		return nil, err
	}
	if len(ts) == 0 {
		return nil, nil
	}
	rq := make(map[string]value.Array)
	is := util.Indexs(attrs, aliasList(n.es, n.gs))
	for _, t := range ts {
		a := t.(value.Array)
		for i := range is {
			rq[attrs[i]] = append(rq[attrs[i]], a[is[i]])
		}
	}
	return rq, nil
}

func (n *group) newByAttributes(attrs []string) error {
	limit := n.c.MemSize()
	dv, err := n.c.NewDictVector()
	if err != nil {
		return err
	}
	n.dv = dv
	for {
		mp, err := n.prev.GetAttributes(attrs, limit)
		if err != nil {
			return err
		}
		if len(mp) == 0 || len(mp[attrs[0]]) == 0 {
			return nil
		}
		for i, j := 0, len(mp[attrs[0]]); i < j; i++ {
			k, err := encoding.EncodeValue(util.Map2Tuple(mp, n.gs, i))
			if err != nil {
				return err
			}
			if err := n.dv.Push(string(k), value.Array{util.Map2Tuple(mp, attrs, i)}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (n *group) check(attrs []string) error {
	{
		for i, j := 0, len(n.es); i < j; i++ {
			if len(n.es[i].Name) == 0 {
				return errors.New("need attribute")
			}
			if len(n.es[i].Alias) == 0 {
				return errors.New("need alias")
			}
			switch n.es[i].Op {
			case overload.Avg:
				n.es[i].Agg = avg.New()
			case overload.Max:
				n.es[i].Agg = max.New()
			case overload.Min:
				n.es[i].Agg = min.New()
			case overload.Sum:
				n.es[i].Agg = sum.New()
			case overload.Count:
				n.es[i].Agg = count.New()
			default:
				return fmt.Errorf("unsupport aggreation operator '%v'", n.es[i].Op)
			}
		}
	}
	as, err := n.prev.AttributeList()
	if err != nil {
		return err
	}
	mp := make(map[string]struct{})
	for _, a := range as {
		mp[a] = struct{}{}
	}
	for _, attr := range attrs {
		if _, ok := mp[attr]; !ok {
			return fmt.Errorf("failed to find attribute '%s'", attr)
		}
	}
	return nil
}

func aliasList(es []*summarize.Extend, attrs []string) []string {
	var rs []string

	for _, e := range es {
		rs = append(rs, e.Alias)
	}
	return util.MergeAttributes(attrs, rs)
}

func attributeList(es []*summarize.Extend, attrs []string) []string {
	var rs []string

	for _, e := range es {
		rs = append(rs, e.Name)
	}
	return util.MergeAttributes(attrs, rs)
}
