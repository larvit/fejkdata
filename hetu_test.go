package fejkdata

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/larvit/fejkdata/internal/fejktest"
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
			if tail[4] != fejktest.DVVControlTable[whole%31] {
				t.Errorf("hetu after %q = %q, want the control character %c for %s", pre, tail, fejktest.DVVControlTable[whole%31], six+tail[1:4])
			}
			if individual, _ := strconv.Atoi(tail[1:4]); individual%2 != 1 {
				t.Errorf("hetu after %q = %q, want an odd individual number for the operand 3", pre, tail)
			}
		})
	}
}

func TestBuiltinHetuRendersNothingForAYearOutside1800To2099(t *testing.T) {
	for _, year := range []int{1799, 2160} {
		t.Run(strconv.Itoa(year), func(t *testing.T) {
			f := engineOn(6, time.Date(year, 6, 1, 0, 0, 0, 0, time.UTC))
			if got := mustRender(t, f, `{"format":"131052{hetu(last)}","last":"3"}`); got != "131052" {
				t.Errorf("hetu on %d = %q, want 131052 and nothing after it", year, got)
			}
		})
	}
}

func TestBuiltinHetuGivesEachDigitTheParityOfItsOperand(t *testing.T) {
	f := emptyGenerator(7)
	for digit := range 10 {
		t.Run(strconv.Itoa(digit), func(t *testing.T) {
			got := mustRender(t, f, fmt.Sprintf(`{"format":"131052{hetu(last)}","last":"%d"}`, digit))
			if len(got) != 11 {
				t.Fatalf("hetu with the operand %d = %q, want a date, a sign, three digits and a control character", digit, got)
			}
			if individual, _ := strconv.Atoi(got[7:10]); individual%2 != digit%2 {
				t.Errorf("hetu with the operand %d = %q, want an individual number of that parity", digit, got)
			}
		})
	}
}

func hetuIndividualNumbersOverSeeds(t *testing.T, seeds int, parity int) map[string]int {
	t.Helper()
	seed := uint64(0)
	return fejktest.TallyDraws(seeds, func() string {
		got := mustRender(t, emptyGenerator(seed), fmt.Sprintf(`{"format":"131052{hetu(last)}","last":"%d"}`, parity))
		seed++
		return got[7:10]
	})
}

func TestBuiltinHetuDrawsEveryIndividualNumberFrom900To999(t *testing.T) {
	for parity := range 2 {
		t.Run(strconv.Itoa(parity), func(t *testing.T) {
			if seen := hetuIndividualNumbersOverSeeds(t, 3000, parity); len(seen) != 50 {
				t.Errorf("3000 draws with the operand %d gave %d individual numbers, want each of the 50", parity, len(seen))
			}
		})
	}
}
