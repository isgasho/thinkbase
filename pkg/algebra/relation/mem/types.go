package mem

import (
	"fmt"

	"github.com/deepfabric/thinkbase/pkg/algebra/value"
	"github.com/deepfabric/thinkbase/pkg/context"
)

type relation struct {
	plh   int // placeholder
	name  string
	attrs []string
	tuple []value.Tuple
	mp    map[string]int
	ct    context.Context
}

type tuples struct {
	descs []bool
	attrs []string
	r     *relation
	tuple []value.Tuple
}

func (t tuples) Len() int { return len(t.tuple) }

func (t tuples) Swap(i, j int) {
	t.tuple[i], t.tuple[j] = t.tuple[j], t.tuple[i]
}

func (t tuples) Less(i, j int) bool {
	return t.r.less(t.tuple[i], t.tuple[j], t.attrs, t.descs)
}

func (r *relation) String() string {
	s := r.name + "\n"
	for i, as := range r.attrs {
		if i > 0 {
			s += "\t"
		}
		s += as
	}
	s += "\n"
	for _, t := range r.tuple {
		for i, v := range t {
			if i > 0 {
				s += "\t"
			}
			s += fmt.Sprintf("%s", v)
		}
		s += "\n"
	}
	return s
}