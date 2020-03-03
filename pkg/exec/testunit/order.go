package testunit

import (
	"github.com/deepfabric/thinkbase/pkg/exec/unit"
	"github.com/deepfabric/thinkbase/pkg/sql/algebra/order"
	"github.com/deepfabric/thinkbase/pkg/sql/algebra/relation"
	"github.com/deepfabric/thinkbase/pkg/sql/algebra/util"
)

func NewOrder(n int, isNub bool, descs []bool, attrs []string, r relation.Relation) ([]unit.Unit, error) {
	rn, err := r.GetTupleCount()
	if err != nil {
		return nil, err
	}
	var us []unit.Unit
	step := rn / n
	if step < 1 {
		step = 1
	}
	for i := 0; i < rn; i += step {
		u := relation.New("", nil, util.DupMetadata(r.Metadata()))
		cnt := step
		if cnt > rn-i {
			cnt = rn - i
		}
		ts, err := r.GetTuples(i, i+cnt)
		if err != nil {
			return nil, err
		}
		u.AddTuples(ts)
		us = append(us, &orderUnit{isNub, descs, attrs, u})
	}
	return us, nil
}

func (u *orderUnit) Result() (relation.Relation, error) {
	return order.New(u.isNub, u.descs, u.attrs, u.r).Order()
}