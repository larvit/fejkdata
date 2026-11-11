package fence

import (
	"fmt"
	"strings"
	"testing"

	"github.com/larvit/fejkdata/internal/datatype"
	"github.com/larvit/fejkdata/internal/engine/load"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/invariant"
	"github.com/larvit/fejkdata/internal/jsonvalue"
)

func TestDisagreementRefusesTwoTextItems(t *testing.T) {
	l := tree.NewLoadFinds()
	n, err := load.Compile(jsonvalue.String("a"), l)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := load.ResolveInlineTemplates(n, "template", nil, l); err != nil {
		t.Fatal(err)
	}
	l.Finish()
	text := n.(*tree.Template)
	defer func() {
		if r := recover(); r == nil || !strings.HasPrefix(fmt.Sprint(r), invariant.Broken("")) {
			t.Errorf("recovered %v, want the invariant reported as an internal error", r)
		}
	}()
	_ = disagreement(text, datatype.String, text, datatype.String)
}
