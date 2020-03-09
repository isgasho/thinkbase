package summarize

import (
	"fmt"
	"log"
	"testing"

	"github.com/deepfabric/thinkbase/pkg/algebra/extend"
	"github.com/deepfabric/thinkbase/pkg/algebra/projection"
	"github.com/deepfabric/thinkbase/pkg/algebra/relation"
	"github.com/deepfabric/thinkbase/pkg/algebra/relation/mem"
	"github.com/deepfabric/thinkbase/pkg/algebra/summarize/overload"
	"github.com/deepfabric/thinkbase/pkg/algebra/value"
	"github.com/deepfabric/thinkbase/pkg/context"
)

func TestSummarize(t *testing.T) {
	ct := context.New()
	r := newSummarize(ct)
	{
		fmt.Printf("r:\n%s\n", r)
	}
	attrs := []*projection.Attribute{}
	{
		attrs = append(attrs, &projection.Attribute{E: &extend.Attribute{r.Placeholder(), "A"}})
	}
	{
		attrs = append(attrs, &projection.Attribute{E: &extend.Attribute{r.Placeholder(), "a"}})
	}
	{
		attrs = append(attrs, &projection.Attribute{E: &extend.Attribute{r.Placeholder(), "B"}})
	}
	{
		attrs = append(attrs, &projection.Attribute{E: &extend.Attribute{r.Placeholder(), "C"}})
	}
	p := projection.New(r, ct, attrs)
	pr, err := p.Projection()
	if err != nil {
		log.Fatal(err)
	}
	{
		fmt.Printf("pr:\n%s\n", pr)
	}
}

func newSummarize(c context.Context) relation.Relation {
	r := newTestRelation(c)
	{
		fmt.Printf("r:\n%s\n", r)
	}
	ops := []int{}
	gs := []string{}
	attrs := []*Attribute{}
	{
		gs = append(gs, "b")
	}
	{
		ops = append(ops, overload.Avg)
		attrs = append(attrs, &Attribute{Name: "a", Alias: "A"})
	}
	{
		ops = append(ops, overload.Sum)
		attrs = append(attrs, &Attribute{Name: "a", Alias: "B"})
	}
	{
		ops = append(ops, overload.Max)
		attrs = append(attrs, &Attribute{Name: "b", Alias: "C"})
	}
	{
		ops = append(ops, overload.Count)
		attrs = append(attrs, &Attribute{Name: "b", Alias: "D"})
	}
	s := New(ops, gs, attrs, c, r)
	sr, err := s.Summarize(len(r.Metadata()))
	if err != nil {
		log.Fatal(err)
	}
	return sr
}

func newTestRelation(c context.Context) relation.Relation {
	var attrs []string

	attrs = append(attrs, "a")
	attrs = append(attrs, "b")
	attrs = append(attrs, "c")
	r := mem.New("A", attrs, c)
	{
		var t value.Tuple

		t = append(t, value.NewInt(1))
		t = append(t, value.NewString("a"))
		t = append(t, value.NewString("d"))
		r.AddTuple(t)
	}
	{
		var t value.Tuple

		t = append(t, value.NewInt(3))
		t = append(t, value.NewString("c"))
		t = append(t, value.NewString("c"))
		r.AddTuple(t)
	}
	{
		var t value.Tuple

		t = append(t, value.NewInt(4))
		t = append(t, value.NewString("d"))
		t = append(t, value.NewString("f"))
		r.AddTuple(t)
	}
	{
		var t value.Tuple

		t = append(t, value.NewInt(5))
		t = append(t, value.NewString("d"))
		t = append(t, value.NewString("b"))
		r.AddTuple(t)
	}
	{
		var t value.Tuple

		t = append(t, value.NewInt(6))
		t = append(t, value.NewString("e"))
		t = append(t, value.NewString("f"))
		r.AddTuple(t)
	}
	{
		var t value.Tuple

		t = append(t, value.NewInt(7))
		t = append(t, value.NewString("e"))
		t = append(t, value.NewString("e"))
		r.AddTuple(t)
	}
	{
		var t value.Tuple

		t = append(t, value.NewInt(8))
		t = append(t, value.NewString("e"))
		t = append(t, value.NewString("g"))
		r.AddTuple(t)
	}
	{
		var t value.Tuple

		t = append(t, value.NewInt(9))
		t = append(t, value.NewString("e"))
		t = append(t, value.NewString("f"))
		r.AddTuple(t)
	}
	return r
}
