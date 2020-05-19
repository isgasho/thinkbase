package rule201

import (
	"github.com/deepfabric/thinkbase/pkg/vm/container/relation"
	"github.com/deepfabric/thinkbase/pkg/vm/context"
	"github.com/deepfabric/thinkbase/pkg/vm/op"
	igroup "github.com/deepfabric/thinkbase/pkg/vm/op/index/group"
	irestrict "github.com/deepfabric/thinkbase/pkg/vm/op/index/restrict"
	"github.com/deepfabric/thinkbase/pkg/vm/op/origin/group"
	"github.com/deepfabric/thinkbase/pkg/vm/op/origin/restrict"
	"github.com/deepfabric/thinkbase/pkg/vm/op/origin/summarize"
	"github.com/deepfabric/thinkbase/pkg/vm/op/origin/summarize/overload"
	Rule "github.com/deepfabric/thinkbase/pkg/vm/opt/rule"
	"github.com/deepfabric/thinkbase/pkg/vm/types"
)

func New(c context.Context) Rule.Rule {
	return &rule{c}
}

func (r *rule) Match(o op.OP, _ map[string]op.OP) bool {
	if o.Operate() != op.Group {
		return false
	}
	_, ok := o.Children()[0].(irestrict.RestrictOP)
	return ok
}

func (r *rule) Rewrite(o op.OP, mp map[string]op.OP, gmp, gmq map[string]int32) (op.OP, bool) {
	var ts []int
	var no op.OP
	var nes []*igroup.Extend

	fl := o.Children()[0].(irestrict.RestrictOP).Filter()
	prev := o.Children()[0].Children()[0].(relation.Relation)
	gs := o.(group.GroupOP).Group()
	for _, g := range gs {
		typ, ok := gmp[g]
		switch {
		case ok:
			if !r.groupCost(typ, prev, g) {
				return o, false
			}
		default:
			if typ, ok = gmq[g]; !ok {
				return o, false
			}
		}
		ts = append(ts, int(typ))
	}
	es := o.(group.GroupOP).Extends()
	for _, e := range es {
		switch {
		case overload.IsIndexAggFunc(e.Op):
		case overload.IsIndexTryAggFunc(e.Op):
			if !r.summarizeCost(int32(e.Typ), prev, e) {
				return o, false
			}
		default:
			return o, false
		}
		nes = append(nes, &igroup.Extend{
			Typ:   e.Typ,
			Name:  e.Name,
			Alias: e.Alias,
			Op:    overload.Convert(e.Op),
		})
	}
	e := o.(group.GroupOP).Extend()
	no = igroup.New(prev, fl, ts, gs, nes, r.c)
	if e != nil {
		no = restrict.New(no, e, r.c)
	}
	if parent, ok := mp[o.String()]; ok {
		ps := parent.String()
		children := parent.Children()
		for i, child := range children {
			if child == o {
				parent.SetChild(no, i)
				break
			}
		}
		mp[no.String()] = parent
		if gparent, ok := mp[ps]; ok {
			mp[parent.String()] = gparent
		}
	} else {
		mp[""] = no
	}
	return no, true
}

func (r *rule) groupCost(typ int32, _ relation.Relation, attr string) bool {
	return false
}

func (r *rule) summarizeCost(typ int32, _ relation.Relation, e *summarize.Extend) bool {
	if typ == types.T_string && overload.IsMax(e.Op) {
		return false
	}
	return true
}