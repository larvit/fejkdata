package builtinfunc

import (
	"fmt"
	"testing"
)

// DVV's "Personal identity code" page: the control character is the remainder of the nine-digit
// number, the date and the individual number, divided by 31, read in this table.
const dvvControlCharacters = "0123456789ABCDEFHJKLMNPRSTUVWXY"

func TestHetuCheckIsTheRemainderOfTheNineDigitsDividedBy31InDVVsTable(t *testing.T) {
	for remainder := range 31 {
		digits := fmt.Sprintf("%09d", remainder+31*3_000_000)
		if got, want := hetuCheck(digits), dvvControlCharacters[remainder]; got != want {
			t.Errorf("hetuCheck(%s) = %c, want %c for the remainder %d", digits, got, want, remainder)
		}
	}
}

// DVV's example of a personal identity code is 131052-308T.
func TestHetuCheckGivesTheCharacterOfDVVsExample(t *testing.T) {
	if got := hetuCheck("131052308"); got != 'T' {
		t.Errorf("hetuCheck(131052308) = %c, want T", got)
	}
}

// DVV's century signs: + for the 1800s, - for the 1900s and A for the 2000s. The year is the latest
// not after the year of today that bears the year digits.
func TestHetuSignNamesTheCenturyOfTheLatestYearNotAfterToday(t *testing.T) {
	for _, c := range []struct {
		today, yy int
		want      byte
	}{
		{2026, 0, 'A'}, {2026, 26, 'A'}, {2026, 27, '-'}, {2026, 99, '-'},
		{2000, 0, 'A'}, {1999, 0, '-'}, {1990, 90, '-'}, {1990, 91, '+'}, {1990, 99, '+'},
		{1899, 0, '+'}, {1800, 0, '+'}, {1799, 99, 0}, {2099, 0, 'A'}, {2100, 0, 0},
	} {
		if got := hetuSign(c.today, c.yy); got != c.want {
			t.Errorf("hetuSign(%d, %d) = %q, want %q", c.today, c.yy, got, c.want)
		}
	}
}

// DVV: odd individual numbers go to men, even to women, and all issued are 002 to 899.
func TestHetuTailGivesAnIndividualNumberOfTheParityFrom002To899AndItsCheck(t *testing.T) {
	for _, parity := range []byte{'0', '1', '2', '3', '9'} {
		seen := map[string]bool{}
		for pick := range 449 {
			tail := hetuTail("131052", '-', parity, pick)
			if len(tail) != 5 || tail[0] != '-' {
				t.Fatalf("hetuTail(131052, -, %c, %d) = %q, want a sign, three digits and a character", parity, pick, tail)
			}
			number := int(tail[1]-'0')*100 + int(tail[2]-'0')*10 + int(tail[3]-'0')
			if number < 2 || number > 899 || number%2 != int(parity-'0')%2 {
				t.Errorf("hetuTail(131052, -, %c, %d) = %q, want an individual number of that parity from 002 to 899", parity, pick, tail)
			}
			if want := hetuCheck("131052" + tail[1:4]); tail[4] != want {
				t.Errorf("hetuTail(131052, -, %c, %d) = %q, want the check character %c", parity, pick, tail, want)
			}
			seen[tail] = true
		}
		if len(seen) != 449 {
			t.Errorf("449 picks of parity %c gave %d individual numbers, want each of the 449", parity, len(seen))
		}
	}
}
