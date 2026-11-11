package render

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/larvit/fejkdata/internal/drawstate"
	"github.com/larvit/fejkdata/internal/engine/tree"
)

func TestDrawStepsPanicsOnAStepItsNodeLacks(t *testing.T) {
	defer func() {
		if r := recover(); r == nil || !strings.Contains(fmt.Sprint(r), `no field "f"`) {
			t.Errorf("drawSteps(plain, f) recovered %v, want a panic naming the missing field", r)
		}
	}()
	plain := tree.NewTemplate(tree.TemplateSyntax{Format: "plain"})
	drawSteps(drawstate.New(1, time.Time{}), drawFresh, plain, []tree.PathStep{{Kind: tree.StepField, Name: "f"}}, &tree.PinSet{}, nil, nil)
}

func TestDrawStepsRefusesATableStepOnASiblingRead(t *testing.T) {
	defer func() {
		if r := recover(); r == nil || !strings.Contains(fmt.Sprint(r), "sibling") {
			t.Errorf("drawSteps(sibling, a table step) recovered %v, want a panic naming a sibling read", r)
		}
	}()
	drawSteps(drawstate.New(1, time.Time{}), drawSibling, &tree.Table{}, []tree.PathStep{{Kind: tree.StepDraw}}, nil, nil, nil)
}
