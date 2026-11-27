package fejkdata

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

var runData = map[string]string{
	"customers": `{ "format": "", "id": "{seq()}", "name": ["Ada", "Bo", "Cy", "Di", "Ed"] }`,
	"orders":    `{ "format": "", "id": "{seq()}", "customer_id": "{/run.customers as c}{c.id}", "customer": "{c.name}" }`,
	"line":      `{ "format": "", "order_id": "{/run.orders as o}{o.id}", "customer": "{o.customer}", "sku": "{/part}" }`,
	"part":      `{ "format": "{/run.products.sku}" }`,
	"products":  `{ "format": "", "sku": "{upper(4)}" }`,
}

// fakeRun runs emits over files and returns each template's CSV lines, and the templates in the order written.
func fakeRun(t *testing.T, files map[string]string, emits ...Emit) (map[string][]string, []string, error) {
	t.Helper()
	lines := map[string][]string{}
	var order []string
	err := FakeRun(emits, func(e Emit, fake func() (*Record, error)) error {
		order = append(order, e.Template)
		for i := 0; i < e.Count; i++ {
			r, err := fake()
			if err != nil {
				return err
			}
			lines[e.Template] = append(lines[e.Template], r.CSVLine())
		}
		return nil
	}, WithDataPath(writeData(t, files)), WithSeed(1), onTestDay)
	return lines, order, err
}

func TestFakeRunWritesParentsFirstAndAChildReadsItsParentsRows(t *testing.T) {
	lines, order, err := fakeRun(t, runData, Emit{"line", 30}, Emit{"orders", 6}, Emit{"products", 3}, Emit{"customers", 2})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(order, " "); got != "customers orders products line" {
		t.Fatalf("order = %q, want each parent ahead of its first child, else as named", got)
	}
	customers, orders, products := map[string]bool{}, map[string]string{}, map[string]bool{}
	for _, l := range lines["customers"] {
		customers[l] = true
	}
	for _, l := range lines["orders"] {
		id, customer, _ := strings.Cut(l, ",")
		if !customers[customer] {
			t.Fatalf("order %q names a customer outside %v", l, lines["customers"])
		}
		_, name, _ := strings.Cut(customer, ",")
		orders[id] = name
	}
	for _, l := range lines["products"] {
		products[l] = true
	}
	for _, l := range lines["line"] {
		cells := strings.Split(l, ",")
		if name, exists := orders[cells[0]]; !exists || name != cells[1] || !products[cells[2]] {
			t.Fatalf("line %q, want an order of %v with its customer, and a sku of %v", l, orders, lines["products"])
		}
	}
}

func TestFakeRunRefusesARun(t *testing.T) {
	cycle := map[string]string{
		"a": `{ "format": "", "id": "{seq()}", "b": "{/run.b.id}" }`,
		"b": `{ "format": "", "id": "{seq()}", "a": "{/helper}" }`,
		// b reads a's rows through a category of its own folder.
		"helper": `{ "format": "{/run.a.id}" }`,
	}
	for _, c := range []struct {
		name  string
		files map[string]string
		emits []Emit
		want  string
	}{
		{"a read template is not named", runData, []Emit{{"orders", 1}}, "orders reads /run.customers"},
		{"a read template is not named, through another category", runData, []Emit{{"line", 1}, {"orders", 1}, {"customers", 1}}, "part reads /run.products.sku"},
		{"a parent has count 0", runData, []Emit{{"customers", 0}, {"orders", 1}}, "orders reads the rows of customers, whose count is 0"},
		{"a cycle", cycle, []Emit{{"a", 1}, {"b", 1}}, "a reads the rows of b, which reads the rows of a"},
		{"a template reads itself", map[string]string{"a": `{ "format": "", "id": "{seq()}", "prev": "{/run.a.id}" }`}, []Emit{{"a", 1}}, "a reads the rows of a"},
		{"a template is named twice", runData, []Emit{{"customers", 1}, {"customers", 1}}, "customers is named twice"},
		{"a category is named run", map[string]string{"run": `{ "format": "x" }`, "a": `{ "format": "", "id": "1" }`}, []Emit{{"a", 1}}, `"run"`},
		{"a folder is named run", map[string]string{"run/x": `{ "format": "x" }`, "a": `{ "format": "", "id": "1" }`}, []Emit{{"a", 1}}, `"run"`},
		{"a template is no category", runData, []Emit{{"nope", 1}}, "nope"},
		{"no template", runData, nil, "no template"},
		{"a count below 0", runData, []Emit{{"customers", -1}}, "-1"},
	} {
		t.Run(c.name, func(t *testing.T) {
			lines, _, err := fakeRun(t, c.files, c.emits...)
			if !errors.Is(err, ErrRun) || !strings.Contains(err.Error(), c.want) || len(lines) != 0 {
				t.Fatalf("FakeRun = %v, after %v, want an ErrRun saying %q before any record", err, lines, c.want)
			}
		})
	}
}

func TestFakeRunKeepsATemplatesRowsWhateverElseTheRunNames(t *testing.T) {
	alone, _, err := fakeRun(t, runData, Emit{"customers", 5}, Emit{"orders", 9})
	if err != nil {
		t.Fatal(err)
	}
	beside, _, err := fakeRun(t, runData, Emit{"products", 4}, Emit{"line", 2}, Emit{"customers", 5}, Emit{"orders", 9})
	if err != nil {
		t.Fatal(err)
	}
	for _, template := range []string{"customers", "orders"} {
		if got, want := strings.Join(beside[template], "\n"), strings.Join(alone[template], "\n"); got != want {
			t.Errorf("%s beside two more templates:\n%s\nwant as without them:\n%s", template, got, want)
		}
	}
}

func TestFakeRunDrawsTwoTemplatesOfOneShapeApart(t *testing.T) {
	files := map[string]string{"buyers": runData["products"], "sellers": runData["products"]}
	lines, _, err := fakeRun(t, files, Emit{"buyers", 8}, Emit{"sellers", 8})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(lines["buyers"], "\n") == strings.Join(lines["sellers"], "\n") {
		t.Errorf("buyers and sellers both drew %v, want each template its own draws", lines["buyers"])
	}
}

func TestFakeRunBothCountsZeroStillRendersAChildForItsColumns(t *testing.T) {
	var headers []string
	err := FakeRun([]Emit{{"orders", 0}, {"customers", 0}}, func(e Emit, fake func() (*Record, error)) error {
		if e.Template == "customers" {
			return nil
		}
		r, err := fake()
		if err != nil {
			return err
		}
		headers = append(headers, r.CSVHeader())
		return nil
	}, WithDataPath(writeData(t, runData)), WithSeed(1), onTestDay)
	if err != nil || len(headers) != 1 || headers[0] != "id,customer_id,customer" {
		t.Fatalf("FakeRun = %v, headers %q, want the columns of orders", err, headers)
	}
}

func TestFakeRunReturnsWhatWriteReturns(t *testing.T) {
	stop := errors.New("disk full")
	err := FakeRun([]Emit{{"customers", 1}}, func(Emit, func() (*Record, error)) error { return stop }, WithDataPath(writeData(t, runData)), WithSeed(1), onTestDay)
	if !errors.Is(err, stop) {
		t.Errorf("FakeRun = %v, want write's error", err)
	}
}

func TestFakeRunReturnsNewsErrorWhereAnOptionIsMissing(t *testing.T) {
	if err := FakeRun([]Emit{{"customers", 1}}, func(Emit, func() (*Record, error)) error { return nil }); err == nil || errors.Is(err, ErrRun) {
		t.Errorf("FakeRun with no option = %v, want New's error", err)
	}
}

func TestFakeRunFailsBeforeTheFirstRecordWhereAChildDoesNotLoad(t *testing.T) {
	for name, orders := range map[string]string{
		"a column its parent lacks": `{ "format": "", "c": "{/run.customers.nope}" }`,
		"a reference to nothing":    `{ "format": "", "c": "{/run.customers.id}", "d": "{/nope}" }`,
	} {
		t.Run(name, func(t *testing.T) {
			files := map[string]string{"customers": runData["customers"], "orders": orders}
			lines, _, err := fakeRun(t, files, Emit{"customers", 3}, Emit{"orders", 1})
			if !errors.Is(err, ErrLoad) || errors.Is(err, ErrRun) || !strings.Contains(err.Error(), "nope") || len(lines) != 0 {
				t.Errorf("FakeRun = %v, after %v, want an ErrLoad naming nope before any record", err, lines)
			}
		})
	}
}

func TestFakeRunAChildDrawsFromEveryRowOfItsParent(t *testing.T) {
	for name, c := range map[string]struct {
		orders string
		want   []string
	}{
		"distinct 2":    {`{ "format": "", "n": "{/run.customers distinct 2 joined ' '}" }`, []string{"1 2", "1 3", "2 1", "2 3", "3 1", "3 2"}},
		"where not '1'": {`{ "format": "", "n": "{/run.customers where not '1'}" }`, []string{"2", "3"}},
	} {
		t.Run(name, func(t *testing.T) {
			files := map[string]string{"customers": runData["customers"], "orders": c.orders}
			lines, _, err := fakeRun(t, files, Emit{"customers", 3}, Emit{"orders", 6})
			if err != nil || len(lines["orders"]) != 6 {
				t.Fatalf("FakeRun = %v, after %v, want 6 orders", err, lines)
			}
			for _, l := range lines["orders"] {
				if !slices.Contains(c.want, l) {
					t.Errorf("order %q, want one of %q", l, c.want)
				}
			}
		})
	}
}

func TestFakeRunFailsBeforeTheFirstRecordWhereALaterParentRowDoesNotLoadInAChild(t *testing.T) {
	files := map[string]string{
		// ref is an integer in the first customer, and not in the third.
		"customers": `{ "format": "", "ref": { "format": "{calc(2 / k)}", "k": "{seq()}" } }`,
		"orders":    `{ "format": "", "ref": { "format": "{/run.customers.ref}", "datatype": "integer" } }`,
	}
	lines, _, err := fakeRun(t, files, Emit{"customers", 3}, Emit{"orders", 3})
	if !errors.Is(err, ErrLoad) || !strings.Contains(err.Error(), "is not an integer") || len(lines) != 0 {
		t.Errorf("FakeRun = %v, after %v, want an ErrLoad on ref before any record", err, lines)
	}
}

func TestATemplateOfARunLoadedOutsideOneNamesTheRun(t *testing.T) {
	_, err := New(WithDataPath(writeData(t, runData)), WithSeed(1), onTestDay)
	if !errors.Is(err, ErrLoad) {
		t.Fatalf("New = %v, want an ErrLoad", err)
	}
	for _, want := range []string{"{/run.", "--emit", "FakeRun"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("New = %v, want it to name %s", err, want)
		}
	}
}
