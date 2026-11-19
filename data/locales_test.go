package data_test

import (
	"encoding/json"
	"io/fs"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/larvit/fejkdata/data"
)

// shippedLocale is what the shared locale tests expect of a locale. A field runs the shared tests
// of its categories for the locale; sharedCategories maps each category to its field.
type shippedLocale struct {
	name string
	// Its word tables by part of speech.
	words    map[string]wordTable
	months   []string
	weekdays []string
	everyday *everyday
	names    *names
	person   *person
	company  *company
	// Go's layout of the locale's date: "2.1.2006" pads neither the day nor the month.
	dateLayout string
	emailFold  *strings.Replacer
	phone      *phone
	// The shared categories it ships whose tests live outside the shared ones, so no field is set for them.
	testedElsewhere []string
}

type wordTable struct {
	// The column holding the word's form, or "" where the table has none.
	form    string
	spelled *regexp.Regexp
}

type everyday struct {
	price    string
	time     string
	url      string
	username string
	colors   []string
}

type names struct {
	// First names borne by one sex alone, each with that sex, and first names borne by both.
	ofOneSex    map[string]string
	ofBothSexes []string
}

type person struct {
	sexWords map[string]string
	titles   []string
}

// A company's legal forms: the one whose names end in no legal form, and those whose names end in theirs.
type company struct {
	name     *regexp.Regexp
	bare     legalForm
	suffixed []legalForm
}

// A legal form a company draws: the code its country's register counts it under, the last step of
// its path under company, the words a name of it ends in, "" for the form that prints none, and
// its units: the source's count, or its even share of the group the source counts it in.
type legalForm struct {
	code   string
	path   string
	suffix string
	units  int
}

type phone struct {
	callingCode string
	// What its national form opens with and its international form drops, or "".
	trunkPrefix string
	fixed       func(t *testing.T) func(national string) bool
	// Its tables of area codes, each with the code in area and the code without its trunk 0 in ndc.
	areaTables []string
}

const (
	slug    = `[a-z0-9]+(-[a-z0-9]+)*`
	urlPath = `(/` + slug + `){0,2}`
)

var (
	lowerWord       = regexp.MustCompile(`^\p{Ll}+$`)
	capitalisedWord = regexp.MustCompile(`^\p{Lu}\p{Ll}+$`)
	asciiWord       = regexp.MustCompile(`^[a-z]+$`)
	definiteForms   = map[string]wordTable{"noun": {"definite", lowerWord}, "verb": {"present", lowerWord}, "adj": {"definite", lowerWord}, "adv": {"", lowerWord}}
	latinFolds      = strings.NewReplacer("æ", "ae", "ø", "o", "å", "a")
	trunkAreaTables = []string{"phone-fixed", "phone-mobile"}
)

var shippedLocales = []shippedLocale{
	{
		name:     "da_DK",
		names:    &names{map[string]string{"Mette": "f", "Freja": "f", "Søren": "m", "Mads": "m"}, []string{"Kim"}},
		words:    definiteForms,
		months:   []string{"januar", "februar", "marts", "april", "maj", "juni", "juli", "august", "september", "oktober", "november", "december"},
		weekdays: []string{"mandag", "tirsdag", "onsdag", "torsdag", "fredag", "lørdag", "søndag"},
		everyday: &everyday{`^\d{1,3}(\.\d{3})?(,\d{2})? kr\.$`, `^([01]\d|2[0-3])\.[0-5]\d$`, `^https://([a-z]+\.)?example\.(com|org|net|dk)` + urlPath + `$`, `^[a-z0-9._]{3,}$`, []string{"Blå", "Grøn", "Rød"}},
		person:   &person{map[string]string{"f": "kvinde", "m": "mand"}, []string{"dr.", "professor"}},
		// Statistics Denmark's table GF15, 2024: 132,403 sole proprietorships, 218,208 private limited companies,
		// 30,750 public limited companies and 11,679 partnerships and limited partnerships, which I/S and K/S split evenly.
		company: &company{regexp.MustCompile(`^\p{Lu}[\p{L}\p{N} &'.-]*[\p{L}\p{N}]( (ApS|A/S|I/S|K/S))?$`), legalForm{"ENK", "enk", "", 132403}, []legalForm{
			{"ApS", "aps", "ApS", 218208}, {"A/S", "as", "A/S", 30750}, {"I/S", "is", "I/S", 5840}, {"K/S", "ks", "K/S", 5839},
		}},
		// Dansk Sprognævn's answer SV00000046 gives 2.10.2003 and advises against 02.10.2003.
		dateLayout: "2.1.2006",
		emailFold:  latinFolds,
		phone:      &phone{"+45", "", danishFixedPhone, nil},
	},
	{
		name:     "de_DE",
		names:    &names{map[string]string{"Emilia": "f", "Marie": "f", "Maximilian": "m", "Felix": "m"}, []string{"Noa"}},
		words:    map[string]wordTable{"noun": {"plural", capitalisedWord}, "verb": {"", lowerWord}, "adj": {"", lowerWord}, "adv": {"", lowerWord}},
		months:   []string{"Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"},
		weekdays: []string{"Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag", "Sonntag"},
		everyday: &everyday{`^\d{1,3}(\.\d{3})?,\d{2} €$`, `^([01]\d|2[0-3]):[0-5]\d$`, `^https://([a-z]+\.)?example\.(com|org|net|de)` + urlPath + `$`, `^[a-z0-9._]{3,}$`, []string{"Blau", "Grün", "Rot"}},
		person:   &person{map[string]string{"f": "Frau", "m": "Mann"}, []string{"Dr.", "Prof.", "Prof. Dr."}},
		// Destatis' Unternehmensregister, 2024: 2,062,752 sole proprietors, 440,951 partnerships and 861,720
		// corporations, each group split evenly over the forms drawn.
		company: &company{regexp.MustCompile(`^\p{Lu}[\p{L}\p{N} &'.-]*[\p{L}\p{N}]( (KG|OHG|GmbH|UG \(haftungsbeschränkt\)|AG))?$`), legalForm{"Einzelunternehmen", "einzel", "", 2062752}, []legalForm{
			{"KG", "kg", "KG", 220476}, {"OHG", "ohg", "OHG", 220475},
			{"GmbH", "gmbh", "GmbH", 287240}, {"UG (haftungsbeschränkt)", "ug", "UG (haftungsbeschränkt)", 287240}, {"AG", "ag", "AG", 287240},
		}},
		// DIN 5008 writes a date as TT.MM.JJJJ.
		dateLayout: "02.01.2006",
		emailFold:  strings.NewReplacer("ä", "a", "ö", "o", "ü", "u", "ß", "ss"),
		phone:      &phone{"+49", "0", germanFixedPhone, trunkAreaTables},
	},
	{
		name:       "en_US",
		names:      &names{map[string]string{"Emily": "f", "Olivia": "f", "Liam": "m", "Noah": "m"}, []string{"Taylor"}},
		words:      map[string]wordTable{"noun": {"", asciiWord}, "verb": {"present", asciiWord}, "adj": {"", asciiWord}, "adv": {"", asciiWord}},
		months:     []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"},
		weekdays:   []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"},
		dateLayout: "01/02/2006",
		// english_test.go covers its person, sex, title and phone, and TestShippedDataCategories of ../data_test.go the rest.
		testedElsewhere: []string{"color", "company", "email", "person", "phone", "price", "sex", "time", "title", "url", "username"},
	},
	{
		name:   "et_EE",
		names:  &names{map[string]string{"Liis": "f", "Kristiina": "f", "Andres": "m", "Toomas": "m"}, nil},
		person: &person{map[string]string{"f": "naine", "m": "mees"}, []string{"dr", "prof"}},
	},
	{
		name:     "fi_FI",
		names:    &names{map[string]string{"Anne": "f", "Aino": "f", "Juha": "m", "Matti": "m"}, []string{"Kim"}},
		words:    map[string]wordTable{"noun": {"", lowerWord}, "verb": {"", lowerWord}, "adj": {"", lowerWord}, "adv": {"", lowerWord}},
		months:   []string{"tammikuu", "helmikuu", "maaliskuu", "huhtikuu", "toukokuu", "kesäkuu", "heinäkuu", "elokuu", "syyskuu", "lokakuu", "marraskuu", "joulukuu"},
		weekdays: []string{"maanantai", "tiistai", "keskiviikko", "torstai", "perjantai", "lauantai", "sunnuntai"},
		everyday: &everyday{`^\d{1,3}( \d{3})?,\d{2} €$`, `^([01]\d|2[0-3])\.[0-5]\d$`, `^https://([a-z]+\.)?example\.(com|org|net|fi)` + urlPath + `$`, `^[a-z0-9._]{3,}$`, []string{"Sininen", "Vihreä", "Punainen"}},
		person:   &person{map[string]string{"f": "nainen", "m": "mies"}, []string{"tohtori", "professori"}},
		company: &company{regexp.MustCompile(`^\p{L}[\p{L}\p{N} &'.-]*[\p{L}\p{N}]( (Oy|Ky|Ay|osk))?$`), legalForm{"Tmi", "tmi", "", 295203}, []legalForm{
			{"Oy", "oy", "Oy", 240524}, {"Ky", "ky", "Ky", 13356}, {"Ay", "ay", "Ay", 4726}, {"osk", "osk", "osk", 2811},
		}},
		dateLayout: "2.1.2006",
		emailFold:  strings.NewReplacer("ä", "a", "ö", "o", "å", "a"),
		phone:      &phone{"+358", "0", finnishFixedPhone, trunkAreaTables},
	},
	{
		name:   "fr_FR",
		names:  &names{map[string]string{"Nathalie": "f", "Isabelle": "f", "Jean": "m", "Michel": "m"}, []string{"Camille", "Dominique"}},
		person: &person{map[string]string{"f": "femme", "m": "homme"}, []string{"Dr", "Pr"}},
	},
	{
		name:   "lv_LV",
		names:  &names{map[string]string{"Anna": "f", "Līga": "f", "Jānis": "m", "Mārtiņš": "m"}, nil},
		person: &person{map[string]string{"f": "sieviete", "m": "vīrietis"}, []string{"Dr.", "prof."}},
	},
	{
		name:     "nb_NO",
		names:    &names{map[string]string{"Kari": "f", "Nora": "f", "Ola": "m", "Jakob": "m"}, []string{"Kim"}},
		words:    definiteForms,
		months:   []string{"januar", "februar", "mars", "april", "mai", "juni", "juli", "august", "september", "oktober", "november", "desember"},
		weekdays: []string{"mandag", "tirsdag", "onsdag", "torsdag", "fredag", "lørdag", "søndag"},
		everyday: &everyday{`^\d{1,3}( \d{3})?(,\d{2})? kr$`, `^([01]\d|2[0-3]):[0-5]\d$`, `^https://([a-z]+\.)?example\.(com|org|net|no)` + urlPath + `$`, `^[a-z0-9._]{3,}$`, []string{"Blå", "Grønn", "Rød"}},
		person:   &person{map[string]string{"f": "kvinne", "m": "mann"}, []string{"dr.", "professor"}},
		// The active units of each form in Enhetsregisteret on 2026-10-22, leaving out the bankrupt, those being
		// wound up and those forcibly dissolved.
		company: &company{regexp.MustCompile(`^\p{Lu}[\p{L}\p{N} &'-]*[\p{L}\p{N}]$`), legalForm{"ENK", "enk", "", 465166}, []legalForm{
			{"AS", "as", "AS", 423370}, {"DA", "da", "DA", 9105}, {"SA", "sa", "SA", 6761},
			{"ANS", "ans", "ANS", 6214}, {"ASA", "asa", "ASA", 205}, {"KS", "ks", "KS", 105},
		}},
		dateLayout: "02.01.2006",
		emailFold:  latinFolds,
		phone:      &phone{"+47", "", norwegianFixedPhone, nil},
	},
	{
		name:       "nn_NO",
		words:      definiteForms,
		months:     []string{"januar", "februar", "mars", "april", "mai", "juni", "juli", "august", "september", "oktober", "november", "desember"},
		weekdays:   []string{"måndag", "tysdag", "onsdag", "torsdag", "fredag", "laurdag", "søndag"},
		everyday:   &everyday{`^\d{1,3}( \d{3})?(,\d{2})? kr$`, `^([01]\d|2[0-3]):[0-5]\d$`, `^https://([a-z]+\.)?example\.(com|org|net|no)` + urlPath + `$`, `^[a-z0-9._]{3,}$`, []string{"Blå", "Grøn", "Raud"}},
		dateLayout: "02.01.2006",
		// nynorsk_test.go covers these, each against Bokmål's.
		testedElsewhere: []string{"company", "email", "person", "phone", "title"},
	},
	{
		name:       "sv_SE",
		names:      &names{map[string]string{"Anna": "f", "Karin": "f", "Erik": "m", "Lars": "m"}, []string{"Kim"}},
		words:      definiteForms,
		months:     []string{"januari", "februari", "mars", "april", "maj", "juni", "juli", "augusti", "september", "oktober", "november", "december"},
		weekdays:   []string{"måndag", "tisdag", "onsdag", "torsdag", "fredag", "lördag", "söndag"},
		everyday:   &everyday{`^\d{1,3}( \d{3})?(,\d{2})? kr$`, `^([01]\d|2[0-3]):[0-5]\d$`, `^https://([a-z]+\.)?example\.(com|org|net|se)` + urlPath + `$`, `^[a-z0-9._]{3,}$`, []string{"Blå", "Grön", "Röd"}},
		person:     &person{map[string]string{"f": "kvinna", "m": "man"}, []string{"dr", "prof"}},
		dateLayout: "2006-01-02",
		// swedish_test.go covers its company and phone, and TestShippedDataCategories of ../data_test.go its email.
		testedElsewhere: []string{"company", "email", "phone"},
	},
}

func everyLocale(shippedLocale) bool { return true }

func eachLocale(t *testing.T, has func(l shippedLocale) bool, test func(t *testing.T, l shippedLocale)) {
	t.Helper()
	for _, l := range shippedLocales {
		if has(l) {
			t.Run(l.name, func(t *testing.T) { test(t, l) })
		}
	}
}

// Each category the shared tests cover, with the check that a locale's entry runs those tests.
var sharedCategories = map[string]func(l shippedLocale) bool{
	"word": hasWords, "sentence": hasWords, "paragraph": hasWords,
	"month": hasCalendar, "weekday": hasCalendar,
	"price": hasEveryday, "time": hasEveryday, "url": hasEveryday, "username": hasEveryday, "color": hasEveryday,
	"first-name": hasNames, "last-name": hasNames,
	"person": hasPerson, "sex": hasPerson, "title": hasPerson,
	"company": hasCompany,
	"date":    hasDate,
	"email":   hasEmail,
	"phone":   hasPhone,
}

func (l shippedLocale) coveredCategories() []string {
	var covered []string
	for category, has := range sharedCategories {
		if has(l) {
			covered = append(covered, category)
		}
	}
	covered = append(covered, l.testedElsewhere...)
	slices.Sort(covered)
	return covered
}

var localeNamespace = regexp.MustCompile(`^[a-z]{2}_[A-Z]{2}$`)

// shippedLocaleCategories reads the manifests of data.Modules for each locale's categories, sorted.
func shippedLocaleCategories(t *testing.T) map[string][]string {
	t.Helper()
	categories := map[string][]string{}
	for _, module := range data.Modules() {
		raw, err := fs.ReadFile(module, ".fejkdata.json")
		if err != nil {
			t.Fatal(err)
		}
		var manifest struct {
			Index map[string]json.RawMessage `json:"index"`
		}
		if err := json.Unmarshal(raw, &manifest); err != nil {
			t.Fatal(err)
		}
		for path := range manifest.Index {
			if namespace, category, _ := strings.Cut(path, "."); localeNamespace.MatchString(namespace) {
				categories[namespace] = append(categories[namespace], category)
			}
		}
	}
	for _, shipped := range categories {
		slices.Sort(shipped)
	}
	return categories
}

func TestShippedLocalesAreTheLocaleModulesOfModules(t *testing.T) {
	var listed, modules []string
	for _, l := range shippedLocales {
		listed = append(listed, l.name)
	}
	for namespace := range shippedLocaleCategories(t) {
		modules = append(modules, namespace)
	}
	slices.Sort(modules)
	if !slices.Equal(listed, modules) {
		t.Errorf("shippedLocales holds %v, want one entry per locale module of data.Modules, in order: %v", listed, modules)
	}
}

func TestShippedLocaleCoversEachSharedCategoryItsModuleShipsOnce(t *testing.T) {
	categories := shippedLocaleCategories(t)
	eachLocale(t, everyLocale, func(t *testing.T, l shippedLocale) {
		heldByOtherTestsAlone := func(category string) bool { return sharedCategories[category] == nil }
		ships := slices.DeleteFunc(categories[l.name], heldByOtherTestsAlone)
		if covered := l.coveredCategories(); !slices.Equal(covered, ships) {
			t.Errorf("fields and testedElsewhere cover %q, want each shared category the module ships, once: %q", covered, ships)
		}
	})
}
