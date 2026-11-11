package fejkdata

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestBuiltinHetuRendersNothingForAnOperandThatIsNotOneDigit(t *testing.T) {
	f := emptyGenerator(3)
	for _, last := range []string{"", "mies", "12", "x7", "-"} {
		t.Run(fmt.Sprintf("operand %q", last), func(t *testing.T) {
			tmpl := `{"format":"{pre}{hetu(last)}","pre":"131052","last":"` + last + `"}`
			if got := mustRender(t, f, tmpl); got != "131052" {
				t.Errorf("hetu with operand %q = %q, want 131052 and nothing after it", last, got)
			}
		})
	}
}

func TestBuiltinHetuReadsTheLastSixDigitsEmitted(t *testing.T) {
	f := emptyGenerator(4)
	for pre, six := range map[string]string{"x-131052": "131052", "99131052": "131052", "a1b3c1d0e5f2g": "131052", "061": "000061", "5": "000005", "x": "000000"} {
		t.Run(pre, func(t *testing.T) {
			got := mustRender(t, f, `{"format":"{pre}{hetu(last)}","pre":"`+pre+`","last":"3"}`)
			tail := got[len(pre):]
			if len(tail) != 5 || tail[0] != '-' && tail[0] != 'A' {
				t.Fatalf("hetu after %q = %q, want a sign, three digits and a control character", pre, tail)
			}
			whole, _ := strconv.Atoi(six + tail[1:4])
			if tail[4] != dvvControlTable[whole%31] {
				t.Errorf("hetu after %q = %q, want the control character %c for %s", pre, tail, dvvControlTable[whole%31], six+tail[1:4])
			}
			if individual, _ := strconv.Atoi(tail[1:4]); individual%2 != 1 {
				t.Errorf("hetu after %q = %q, want an odd individual number for the operand 3", pre, tail)
			}
		})
	}
}

func TestBuiltinHetuRendersNothingForAYearOutside1800To2099(t *testing.T) {
	for _, year := range []int{1799, 2160} {
		f := engineOn(6, time.Date(year, 6, 1, 0, 0, 0, 0, time.UTC))
		if got := mustRender(t, f, `{"format":"131052{hetu(last)}","last":"3"}`); got != "131052" {
			t.Errorf("hetu on %d = %q, want 131052 and nothing after it", year, got)
		}
	}
}
