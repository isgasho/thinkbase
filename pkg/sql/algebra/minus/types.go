package minus

import "github.com/deepfabric/thinkbase/pkg/sql/algebra/relation"

type Minus interface {
	Minus() relation.Relation
}

type minus struct {
	a, b relation.Relation
}