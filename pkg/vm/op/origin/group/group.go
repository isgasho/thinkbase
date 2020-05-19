package group

import (
	"errors"
	"fmt"

	"github.com/deepfabric/thinkbase/pkg/vm/container"
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
	"github.com/deepfabric/thinkbase/pkg/vm/types"
	"github.com/deepfabric/thinkbase/pkg/vm/util"
	"github.com/deepfabric/thinkbase/pkg/vm/util/encoding"
	"github.com/deepfabric/thinkbase/pkg/vm/value"
)

func New(prev op.OP, e extend.Extend, gs []string, es []*summarize.Extend, c context.Context) *group {
	return &group{
		c:       c,
		e:       e,
		es:      es,
		gs:      gs,
		prev:    prev,
		isCheck: false,
	}
}

func (n *group) Group() []string {
	return n.gs
}

func (n *group) Extend() extend.Extend {
	return n.e
}

func (n *group) Extends() []*summarize.Extend {
	return n.es
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
			r += fmt.Sprintf("%s(%s, %v) -> %s", overload.AggName[e.Op], e.Name, &types.T{int32(e.Typ)}, e.Alias)
		default:
			r += fmt.Sprintf(", %s(%s, %v) -> %s", overload.AggName[e.Op], e.Name, &types.T{int32(e.Typ)}, e.Alias)
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

func (n *group) GetAttributes(attrs []string, limit int) (map[string]value.Array, error) {
	var as [][]string

	attrs = util.MergeAttributes(attrs, []string{})
	es := subExtend(n.es, attrs, n.e)
	as = append(as, attributeList(es, n.gs))
	as = append(as, aliasList(es, n.gs))
	if !n.isCheck {
		if err := n.check(as[0]); err != nil {
			return nil, err
		}
		if err := util.Contain(attrs, as[1]); err != nil {
			return nil, err
		}
		dv, err := n.c.NewDictVector()
		if err != nil {
			return nil, err
		}
		n.dv = dv
		if err := n.newByAttributes(as[0], limit); err != nil {
			n.dv.Destroy()
			return nil, err
		}
		n.isCheck = true
	}
	size := 0
	rq := make(map[string]value.Array)
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
				return rq, nil
			}
			n.k = k
		}
		ts, err := n.dv.PopsArray(n.k, limit)
		switch {
		case err == container.NotExist || (err == nil && len(ts) == 0):
			var t value.Array
			{
				v, _, err := encoding.DecodeValue([]byte(n.k))
				if err != nil {
					n.dv.Destroy()
					return nil, err
				}
				t = append(t, v.(value.Array)...)
			}
			for _, e := range es {
				if v, err := e.Agg.Eval(); err != nil {
					n.dv.Destroy()
					return nil, err
				} else {
					t = append(t, v)
				}
				e.Agg.Reset()
			}
			mp := util.Tuple2Map(t, as[1])
			if n.e != nil {
				if ok, err := n.e.Eval(mp); err != nil {
					return nil, err
				} else if value.MustBeBool(ok) {
					for _, attr := range attrs {
						if v, ok := mp[attr]; ok {
							rq[attr] = append(rq[attr], v)
						}
					}
					size += t.Size()
				}
			} else {
				for _, attr := range attrs {
					if v, ok := mp[attr]; ok {
						rq[attr] = append(rq[attr], v)
					}
				}
				size += t.Size()
			}
			n.k = ""
			continue
		case err != nil:
			n.dv.Destroy()
			return nil, err
		}
		mp := util.Tuples2Map(ts, as[0])
		for _, e := range es {
			if err := e.Agg.Fill(mp[e.Name]); err != nil {
				n.dv.Destroy()
				return nil, err
			}
		}
	}
	return rq, nil
}

func (n *group) newByAttributes(attrs []string, limit int) error {
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
			switch n.es[i].Op {
			case overload.Avg, overload.AvgI, overload.AvgIt:
				n.es[i].Agg = avg.New(int32(n.es[i].Typ))
			case overload.Max, overload.MaxI, overload.MaxIt:
				n.es[i].Agg = max.New(int32(n.es[i].Typ))
			case overload.Min, overload.MinI, overload.MinIt:
				n.es[i].Agg = min.New(int32(n.es[i].Typ))
			case overload.Sum, overload.SumI, overload.SumIt:
				n.es[i].Agg = sum.New(int32(n.es[i].Typ))
			case overload.Count, overload.CountI, overload.CountIt:
				n.es[i].Agg = count.New(int32(n.es[i].Typ))
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

func subExtend(es []*summarize.Extend, attrs []string, e extend.Extend) []*summarize.Extend {
	var rs []*summarize.Extend

	mp := make(map[string]struct{})
	for i, j := 0, len(attrs); i < j; i++ {
		mp[attrs[i]] = struct{}{}
	}
	if e != nil {
		as := e.Attributes()
		for i, j := 0, len(as); i < j; i++ {
			mp[as[i]] = struct{}{}
		}
	}
	for i, j := 0, len(es); i < j; i++ {
		if _, ok := mp[es[i].Alias]; ok {
			rs = append(rs, es[i])
		}
	}
	return rs
}
