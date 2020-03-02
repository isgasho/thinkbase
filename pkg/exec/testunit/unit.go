package testunit

import (
	"github.com/deepfabric/thinkbase/pkg/exec/unit"
	"github.com/deepfabric/thinkbase/pkg/sql/algebra/relation"
)

func New(n, op int, a, b relation.Relation) ([]unit.Unit, error) {
	switch op {
	case unit.Minus:
		return newMinus(n, a, b)
	case unit.Union:
		return newUnion(n, a, b)
	case unit.Intersect:
		return newIntersect(n, a, b)
	}
	return nil, nil
}