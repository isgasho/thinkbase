package summarize

import (
	"github.com/deepfabric/thinkbase/pkg/vm/context"
	"github.com/deepfabric/thinkbase/pkg/vm/op"
	"github.com/deepfabric/thinkbase/pkg/vm/op/origin/summarize/overload"
)

type Extend struct {
	Op    int
	Typ   int
	Name  string
	Alias string
	Agg   overload.Aggregation
}

type SummarizeOP interface {
	op.OP
	Extends() []*Extend
}

type summarize struct {
	isUsed  bool
	isCheck bool
	prev    op.OP
	es      []*Extend
	c       context.Context
}
