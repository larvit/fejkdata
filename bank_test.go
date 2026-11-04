package fejkdata

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func bankgirotMod11Valid(digits string) bool {
	sum := 0
	for i := range digits {
		sum += int(digits[i]-'0') * ((len(digits)-1-i)%10 + 1)
	}
	return sum%11 == 0
}

// TestSwedishBankgiro draws 991-2346 and 5555-5551's two forms: seven digits with the hyphen before
// the last four, eight with it in the middle, the final digit Luhn's check digit as Bankgirot's
// 10-modul sheet has it.
func TestSwedishBankgiro(t *testing.T) {
	if !luhnValid("9912346") || !luhnValid("55555551") {
		t.Fatal("Bankgirot's two examples, 991-2346 and 5555-5551, must pass the Luhn oracle")
	}
	f := shipped(t, WithSeed(1))
	form := regexp.MustCompile(`^([1-9]\d{2}-\d{4}|[1-9]\d{3}-\d{4})$`)
	seven, eight := 0, 0
	for i := 0; i < 2000; i++ {
		v := fake(t, f, "sv_SE.bankgiro")
		if !form.MatchString(v) || !luhnValid(digitsOnly(v)) {
			t.Fatalf("sv_SE.bankgiro %q, want 991-2346 or 5555-5551's form with a Luhn check digit", v)
		}
		if len(v) == 8 {
			seven++
		} else {
			eight++
		}
	}
	if seven < 800 || eight < 800 {
		t.Errorf("%d bankgiro of seven digits and %d of eight in 2000, want about half each", seven, eight)
	}
}

// TestSwedishPlusgiro draws two to eight digits, the hyphen before the check digit and the rest in
// pairs from the right, with a Luhn check digit.
func TestSwedishPlusgiro(t *testing.T) {
	f := shipped(t, WithSeed(1))
	form := regexp.MustCompile(`^[1-9]\d?( \d{2})*-\d$`)
	lengths := map[int]int{}
	for i := 0; i < 3500; i++ {
		v := fake(t, f, "sv_SE.plusgiro")
		d := digitsOnly(v)
		if !form.MatchString(v) || !luhnValid(d) || len(d) < 2 || len(d) > 8 {
			t.Fatalf("sv_SE.plusgiro %q, want 2 to 8 digits, the hyphen before the last, a Luhn check digit", v)
		}
		if strings.Index(v, "-") != len(v)-2 {
			t.Fatalf("sv_SE.plusgiro %q has its hyphen elsewhere than before the last digit", v)
		}
		lengths[len(d)]++
	}
	for n := 2; n <= 8; n++ {
		if lengths[n] < 300 {
			t.Errorf("%d plusgiro of %d digits in 3500, want about 500", lengths[n], n)
		}
	}
}

// clearingRanges are Bankinfrastruktur i Sverige AB's clearing ranges, each with the bank's name,
// the account type and the comment of Bankgirot's "Bankernas kontonummer" of 2024-02-22, which the
// list's column Konto-typ gives as type:comment. The ranges 9565-9569 and 9600-9609 are left out:
// that document does not publish their comment 4.
var clearingRanges = []struct {
	first, last int
	bank        string
	typ, rule   int
}{
	{1100, 1199, "Nordea", 1, 1},
	{1200, 1399, "Danske Bank", 1, 1},
	{1400, 2099, "Nordea", 1, 1},
	{2110, 2119, "Juni Technology AB", 1, 1},
	{2120, 2129, "Steven AB", 1, 1},
	{2130, 2139, "Zimpler AB", 1, 1},
	{2300, 2399, "Ålandsbanken", 1, 2},
	{2400, 2499, "Danske Bank", 1, 1},
	{3000, 3299, "Nordea", 1, 1},
	{3300, 3300, "Nordea", 2, 1},
	{3301, 3399, "Nordea", 1, 1},
	{3400, 3409, "Länsförsäkringar Bank", 1, 1},
	{3410, 3781, "Nordea", 1, 1},
	{3782, 3782, "Nordea", 2, 1},
	{3783, 3999, "Nordea", 1, 1},
	{4000, 4999, "Nordea", 1, 2},
	{5000, 5999, "SEB", 1, 1},
	{6000, 6999, "Handelsbanken", 2, 2},
	{7000, 7999, "Swedbank", 1, 1},
	{8000, 8999, "Swedbank", 2, 3},
	{9020, 9029, "Länsförsäkringar Bank", 1, 2},
	{9040, 9049, "Citibank (filial)", 1, 2},
	{9060, 9069, "Länsförsäkringar Bank", 1, 1},
	{9070, 9079, "Multitude Bank plc.", 1, 1},
	{9100, 9109, "Nordnet Bank", 1, 2},
	{9120, 9124, "SEB", 1, 1},
	{9130, 9149, "SEB", 1, 1},
	{9150, 9169, "Skandiabanken", 1, 2},
	{9170, 9179, "Ikanobanken", 1, 1},
	{9180, 9189, "Danske Bank", 2, 1},
	{9190, 9199, "DNB Bank ASA, filial Sverige", 1, 2},
	{9230, 9239, "Marginalen Bank", 1, 1},
	{9250, 9259, "SBAB Bank", 1, 1},
	{9270, 9279, "ICA Banken", 1, 1},
	{9280, 9289, "Resurs Bank AB", 1, 1},
	{9300, 9349, "Swedbank", 2, 1},
	{9390, 9399, "Landshypotek", 1, 2},
	{9400, 9449, "Forex Bank", 1, 1},
	{9460, 9469, "Santander Consumer Bank", 1, 1},
	{9470, 9479, "BNP Paribas", 1, 2},
	{9500, 9549, "Nordea (Plusgirot)", 2, 3},
	{9550, 9564, "Avanza Bank", 1, 2},
	{9570, 9579, "Sparbanken Syd", 2, 1},
	{9580, 9589, "Aion Bank", 1, 1},
	{9590, 9599, "EP Bank", 1, 2},
	{9630, 9639, "Lån & Spar Bank A/S, filial", 1, 1},
	{9640, 9649, "NOBA Bank Group", 1, 2},
	{9650, 9659, "MedMera Bank AB", 1, 1},
	{9660, 9669, "Svea Bank", 1, 2},
	{9670, 9679, "JAK Medlemsbank", 1, 2},
	{9680, 9689, "Enity Bank Group", 1, 1},
	{9700, 9709, "Ekobanken", 1, 2},
	{9710, 9719, "Lunar Bank", 1, 2},
	{9750, 9759, "Northmill AB", 1, 2},
	{9780, 9789, "Klarna Bank", 1, 2},
	{9960, 9969, "Nordea (Plusgirot)", 2, 3},
}

func bankAccountParts(t *testing.T, v string) (clearing int, account string) {
	t.Helper()
	m := regexp.MustCompile(`^(\d{4})-(\d+)$`).FindStringSubmatch(v)
	if m == nil {
		t.Fatalf("sv_SE.bankaccount %q, want a four-digit clearing number, a hyphen and an account number", v)
	}
	clearing, _ = strconv.Atoi(m[1])
	return clearing, m[2]
}

// checkBankAccount applies the check Bankgirot gives the clearing number's range.
func checkBankAccount(t *testing.T, v string) {
	t.Helper()
	clearing, account := bankAccountParts(t, v)
	for _, r := range clearingRanges {
		if clearing < r.first || clearing > r.last {
			continue
		}
		c := fmt.Sprintf("%04d", clearing)
		ok := false
		switch {
		case r.typ == 1 && r.rule == 1:
			ok = len(account) == 7 && bankgirotMod11Valid(c[1:]+account)
		case r.typ == 1 && r.rule == 2:
			ok = len(account) == 7 && bankgirotMod11Valid(c+account)
		case r.typ == 2 && r.rule == 2:
			ok = len(account) == 9 && bankgirotMod11Valid(account)
		default:
			ok = len(account) == 10 && luhnValid(account)
		}
		if !ok {
			t.Errorf("sv_SE.bankaccount %q is not valid under %s's type %d, comment %d", v, r.bank, r.typ, r.rule)
		}
		return
	}
	t.Errorf("sv_SE.bankaccount %q has a clearing number that no listed range holds", v)
}

// Bankgirot's own example of a number correct under the 11-modul is 1912763608957.
func TestSwedishBankAccountOracleAcceptsBankgirotsExample(t *testing.T) {
	if !bankgirotMod11Valid("1912763608957") || bankgirotMod11Valid("1912763608956") {
		t.Fatal("the 11-modul oracle must accept Bankgirot's 1912763608957 and refuse a changed check digit")
	}
}

func TestSwedishBankAccountHasOneRowPerClearingNumber(t *testing.T) {
	b, err := os.ReadFile("data/sv_SE/bankaccount.tsv")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
	if lines[0] != "clearing\tbank\tnumber" {
		t.Fatalf("header %q, want clearing, bank and number", lines[0])
	}
	var want []string
	for _, r := range clearingRanges {
		for c := r.first; c <= r.last; c++ {
			want = append(want, fmt.Sprintf("%d\t%s", c, r.bank))
		}
	}
	if len(lines)-1 != len(want) {
		t.Fatalf("%d rows, want %d: one for each clearing number of the listed ranges", len(lines)-1, len(want))
	}
	for i, l := range lines[1:] {
		if !strings.HasPrefix(l, want[i]+"\t") {
			t.Fatalf("row %d is %q, want it to open with %q", i+1, l, want[i])
		}
	}
}

func TestSwedishBankAccountOfEveryClearingNumberChecks(t *testing.T) {
	f := shipped(t, WithSeed(1))
	for _, r := range clearingRanges {
		for c := r.first; c <= r.last; c++ {
			v := fakeTemplate(t, f, fmt.Sprintf("{/sv_SE.bankaccount[%d]}", c))
			if !strings.HasPrefix(v, strconv.Itoa(c)+"-") {
				t.Fatalf("sv_SE.bankaccount[%d] = %q, want it to open with its clearing number", c, v)
			}
			checkBankAccount(t, v)
			if bank := fakeTemplate(t, f, fmt.Sprintf("{/sv_SE.bankaccount[%d].bank}", c)); bank != r.bank {
				t.Fatalf("sv_SE.bankaccount[%d].bank = %q, want %q", c, bank, r.bank)
			}
		}
	}
}

func TestSwedishBankAccount(t *testing.T) {
	f := shipped(t, WithSeed(2))
	banks := map[string]bool{}
	for i := 0; i < 3000; i++ {
		bank, v, _ := strings.Cut(fakeTemplate(t, f, "{/sv_SE.bankaccount as a}{a.bank}={a}"), "=")
		checkBankAccount(t, v)
		banks[bank] = true
	}
	for _, b := range []string{"Nordea", "SEB", "Swedbank", "Handelsbanken", "Danske Bank"} {
		if !banks[b] {
			t.Errorf("3000 sv_SE.bankaccount never drew a clearing number of %s", b)
		}
	}
}
