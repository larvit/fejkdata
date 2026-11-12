package render

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

func TestDrawFreshPanicsOnAStepItsNodeLacks(t *testing.T) {
	defer func() {
		if r := recover(); r == nil || !strings.Contains(fmt.Sprint(r), `no field "f"`) {
			t.Errorf("drawFresh(plain, f) recovered %v, want a panic naming the missing field", r)
		}
	}()
	plain := tree.NewTemplate(tree.TemplateSyntax{Format: "plain"})
	drawFresh(drawstate.New(1, time.Time{}), plain, []tree.PathStep{{Kind: tree.StepField, Name: "f"}}, &tree.PinSet{})
}

func TestDrawSiblingRefusesATableStep(t *testing.T) {
	defer func() {
		if r := recover(); r == nil || !strings.Contains(fmt.Sprint(r), "sibling") {
			t.Errorf("drawSibling(a table step) recovered %v, want a panic naming a sibling read", r)
		}
	}()
	drawSibling(drawstate.New(1, time.Time{}), &tree.Table{}, []tree.PathStep{{Kind: tree.StepDraw}})
}
