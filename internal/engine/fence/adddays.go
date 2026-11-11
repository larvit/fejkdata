package fence

import (
	"fmt"
	"math"

	"github.com/larvit/fejkdata/internal/builtinfunc"
	"github.com/larvit/fejkdata/internal/engine/tree"
	"github.com/larvit/fejkdata/internal/grammar"
)

// proveAddDays proves, for each adddays of t whose day count is a read, that the count is a whole
// number within the cap.
func proveAddDays(label string, t *tree.Template) error {
	for _, o := range t.Compiled().Ops() {
		if tok := o.Token(); tok.Kind != grammar.BuiltinCall || tok.Fn != builtinfunc.AddDaysName || len(o.Operands()) == 1 {
			continue
		}
		if err := proveDayCount(o.Operands()[1]); err != nil {
			return fmt.Errorf("%s: token {%s}: %w", label, o.Token().Body, err)
		}
	}
	return nil
}

func proveDayCount(n tree.Arm) error {
	v := (&valueProof{}).proveArm(n)
	switch {
	case v.NotOperand != "":
		return fmt.Errorf("%s is not always a whole number of days: %s; write an integer or read one {int(min,max)}", n.Spelling(), v.NotOperand)
	case !v.Integral:
		return fmt.Errorf("%s is not always a whole number of days: it may print a fraction; read an {int(min,max)}", n.Spelling())
	case math.Max(math.Abs(v.Lo), math.Abs(v.Hi)) > builtinfunc.MaxDays:
		return fmt.Errorf("%s may reach %.0f, which exceeds the maximum of %d days either way; narrow it", n.Spelling(), math.Max(math.Abs(v.Lo), math.Abs(v.Hi)), builtinfunc.MaxDays)
	}
	return nil
}
