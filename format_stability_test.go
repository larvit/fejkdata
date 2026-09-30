package fejkdata

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSeededOutputIsStable pins the exact bytes a seeded generator emits for each
// piece of the token grammar, so a change to how a format is scanned or rendered
// cannot quietly shift the rng stream that reproducibility depends on.
func TestSeededOutputIsStable(t *testing.T) {
	dir := writeData(t, map[string]string{
		"alt":          `{"format":"{a|b}","a":"A","b":"B"}`,
		"calc":         `{"format":"{net} x {qty} = {calc(net * qty, 2)}","net":["19.99","5.00","100.00"],"qty":["2","3","7"]}`,
		"classes":      `"{digits(2)}-{int(1,9)}{int(1,9)}-{upper(2)}-{lower(2)}"`,
		"escapes":      `{"format":"01Aa#{x}","x":"!"}`,
		"funcs":        `"{hex(6)} {int(10,99)} {float(0,1,3)} {nanoid(5)} {seq()}"`,
		"municipality": `{"format":"{name}","rows":"municipality.tsv","key":"code","name":"name","parent":"region"}`,
		"nested":       `{"format":"{outer}","outer":{"format":"{inner}-{digits(2)}","inner":"i"}}`,
		"paths":        `{"format":"{p.a}{p.b}","p":[{"format":"x","a":"1","b":"2"},{"format":"y","a":"A","b":"B"},{"format":"z","a":"a","b":"b"}]}`,
		"ref":          `"see {/alt}"`,
		"refpath":      `"{/region.municipality.name} in {/region.name}"`,
		"region":       `{"format":"{name}","rows":"region.tsv","key":"code","name":"name","weight":"people"}`,
		"repeat":       `{"format":"{w}","repeat":4,"separator":",","w":["x","y","z"]}`,
		"sums":         `{"format":"9{d}{luhn()} {e}{ean()} {m}{mod11()}","d":"012345678901234","e":"123456789012","m":"12345678"}`,
		"weights":      `[{"format":"big","weight":9},"tiny"]`,
	})
	for name, rows := range map[string]string{
		"municipality": "code\tname\tregion\n0180\tStockholm\t01\n0184\tSolna\t01\n1280\tMalmö\t12\n1281\tLund\t12\n",
		"region":       "code\tname\tpeople\n01\tStockholms län\t3\n12\tSkåne län\t1\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name+".tsv"), []byte(rows), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want := map[string][]string{
		"alt":                            {"A", "A", "B", "A"},
		"calc":                           {"100.00 x 3 = 300.00", "100.00 x 3 = 300.00", "100.00 x 7 = 700.00", "5.00 x 3 = 15.00"},
		"classes":                        {"49-74-VY-gj", "38-68-RP-xo", "47-68-IB-hs", "91-89-LP-gk"},
		"escapes":                        {"01Aa#!", "01Aa#!", "01Aa#!", "01Aa#!"},
		"funcs":                          {"0016a2 33 0.649 k4Kwj 1", "a00c07 84 0.175 6UjGv 2", "f70eb8 12 0.390 p5p6g 3", "9048a3 90 0.531 I3Aqv 4"},
		"nested":                         {"i-73", "i-23", "i-67", "i-85"},
		"paths":                          {"ab", "ab", "AB", "AB"},
		"paths.p.a":                      {"A", "a", "a", "A"},
		"ref":                            {"see A", "see A", "see A", "see B"},
		"refpath":                        {"Solna in Stockholms län", "Malmö in Skåne län", "Stockholm in Stockholms län", "Malmö in Skåne län"},
		"region.municipality[0184].name": {"Solna", "Solna", "Solna", "Solna"},
		"region[12].municipality.name":   {"Malmö", "Malmö", "Lund", "Malmö"},
		"repeat":                         {"z,y,z,y", "z,z,y,y", "z,z,x,z", "x,z,y,y"},
		"sums":                           {"90123456789012348 1234567890124 123456780", "90123456789012348 1234567890124 123456780", "90123456789012348 1234567890124 123456780", "90123456789012348 1234567890124 123456780"},
		"weights":                        {"big", "big", "big", "big"},
	}
	for path := range want {
		f := newGenerator(t, dir, WithSeed(42))
		for i, expect := range want[path] {
			if got := fake(t, f, path); got != expect {
				t.Errorf("%s draw %d = %q, want %q", path, i, got, expect)
			}
		}
	}
}
