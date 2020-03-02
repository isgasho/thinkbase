package summarize

import (
	"errors"

	"github.com/deepfabric/thinkbase/pkg/sql/algebra/relation"
	"github.com/deepfabric/thinkbase/pkg/sql/algebra/summarize/overload"
	"github.com/deepfabric/thinkbase/pkg/sql/algebra/summarize/overload/avg"
	"github.com/deepfabric/thinkbase/pkg/sql/algebra/summarize/overload/count"
	"github.com/deepfabric/thinkbase/pkg/sql/algebra/summarize/overload/max"
	"github.com/deepfabric/thinkbase/pkg/sql/algebra/summarize/overload/min"
	"github.com/deepfabric/thinkbase/pkg/sql/algebra/summarize/overload/sum"
	"github.com/deepfabric/thinkbase/pkg/sql/algebra/util"
	"github.com/deepfabric/thinkbase/pkg/sql/algebra/value"
)

func New(ops []int, gs []string, as []*Attribute, r relation.Relation) *summarize {
	var is []int
	var aggs []overload.Aggregation

	for _, a := range gs {
		idx, err := r.GetAttributeIndex(a)
		if err != nil {
			return nil
		}
		is = append(is, idx)
	}
	for _, op := range ops {
		if agg, ok := Aggs[op]; !ok {
			return nil
		} else {
			aggs = append(aggs, agg)
		}
	}
	return &summarize{r: r, is: is, gs: gs, as: as, aggs: aggs}
}

func (s *summarize) Summarize() (relation.Relation, error) {
	if len(s.is) > 0 {
		return s.summarizeByGroup()
	}
	return s.summarize()
}

func (s *summarize) summarize() (relation.Relation, error) {
	r, err := s.newRelation()
	if err != nil {
		return nil, err
	}
	var t value.Tuple
	for i, attr := range s.as {
		s.aggs[i].Reset()
		a, err := s.r.GetAttribute(attr.Name)
		if err != nil {
			return nil, err
		}
		if err := s.aggs[i].Fill(a); err != nil {
			return nil, err
		}
		if v, err := s.aggs[i].Eval(); err != nil {
			return nil, err
		} else {
			t = append(t, v)
		}
	}
	r.AddTuple(t)
	return r, nil
}

func (s *summarize) summarizeByGroup() (relation.Relation, error) {
	var r relation.Relation

	mp := make(map[string]int)
	switch {
	case len(s.as) > 0:
		var err error

		for _, attr := range s.as {
			idx, err := s.r.GetAttributeIndex(attr.Name)
			if err != nil {
				return nil, err
			}
			mp[attr.Name] = idx
		}
		r, err = s.newRelationByGroup()
		if err != nil {
			return nil, err
		}
	default:
		r = relation.New("", nil, s.r.Metadata())
	}
	gs, err := s.group()
	if err != nil {
		return nil, err
	}
	if len(s.as) > 0 {
		for _, g := range gs {
			var t value.Tuple
			for i, attr := range s.as {
				s.aggs[i].Reset()
				if err := s.aggs[i].Fill(g.as[mp[attr.Name]]); err != nil {
					return nil, err
				}
				if v, err := s.aggs[i].Eval(); err != nil {
					return nil, err
				} else {
					t = append(t, v)
				}
			}
			g.r = append(g.r, t...)
		}
	}
	for _, g := range gs {
		r.AddTuple(g.r)
	}
	return r, nil
}

func (s *summarize) newRelation() (relation.Relation, error) {
	var as []*relation.AttributeMetadata

	for _, a := range s.as {
		name, err := getAttributeName(a)
		if err != nil {
			return nil, err
		}
		as = append(as, &relation.AttributeMetadata{
			Name:  name,
			Types: make(map[int32]int),
		})
	}
	return relation.New("", nil, as), nil
}

func (s *summarize) newRelationByGroup() (relation.Relation, error) {
	as := s.r.Metadata()
	for _, a := range s.as {
		name, err := getAttributeName(a)
		if err != nil {
			return nil, err
		}
		as = append(as, &relation.AttributeMetadata{
			Name:  name,
			Types: make(map[int32]int),
		})
	}
	return relation.New("", nil, as), nil
}

type group struct {
	r  value.Tuple
	as []value.Attribute
}

func (s *summarize) group() ([]*group, error) {
	ts, err := util.GetTuples(s.r)
	if err != nil {
		return nil, err
	}
	gs := []*group{}
	mp := make(map[string]*group)
	for i, j := 0, len(ts); i < j; i++ {
		t := getTuple(ts[i], s.is)
		k := t.String()
		if _, ok := mp[k]; !ok {
			g := &group{r: ts[i]}
			for _, v := range ts[i] {
				g.as = append(g.as, value.Attribute{v})
			}
			mp[k] = g
			gs = append(gs, g)
		} else {
			g := mp[k]
			for i, v := range ts[i] {
				g.as[i] = append(g.as[i], v)
			}
		}
	}
	return gs, nil
}

func getTuple(t value.Tuple, is []int) value.Tuple {
	var r value.Tuple

	for _, i := range is {
		r = append(r, t[i])
	}
	return r
}

func getAttributeName(a *Attribute) (string, error) {
	if len(a.Alias) == 0 {
		return "", errors.New("need alias")
	}
	return a.Alias, nil
}

var Aggs = map[int]overload.Aggregation{
	overload.Avg:   avg.New(),
	overload.Max:   max.New(),
	overload.Min:   min.New(),
	overload.Sum:   sum.New(),
	overload.Count: count.New(),
}