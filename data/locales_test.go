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

// shippedLocale is what the tests shared by the locales expect of one. A nil or empty field
// leaves the locale out of that field's tests: its own file then holds what it tests instead.
type shippedLocale struct {
	name string
	// By part of speech, the column of its word table holding the word's form, or "" where it has none.
	wordForms map[string]string
	word      *regexp.Regexp
	noun      *regexp.Regexp
	months    []string
	weekdays  []string
	everyday  *everyday
	person    *person
	company   *company
	// Go's layout of the locale's date: "2.1.2006" pads neither the day nor the month.
	dateLayout string
	emailFold  *strings.Replacer
	phone      *phone
}

type everyday struct {
	price    string
	time     string
	url      string
	username string
	colors   []string
}

type person struct {
	sexWords map[string]string
	titles   []string
}

type company struct {
	name  *regexp.Regexp
	forms []legalForm
}

// A legal form a company draws: the code its country's register counts it under, the last step of
// its path under company, the words a name of it ends in, "" for the form that prints none, and
// how many units the source counts.
type legalForm struct {
	code   string
	path   string
	suffix string
	units  int
}

type phone struct {
	callingCode      string
	dropsTrunkPrefix bool
	fixed            func(t *testing.T) func(national string) bool
}

const (
	slug    = `[a-z0-9]+(-[a-z0-9]+)*`
	urlPath = `(/` + slug + `){0,2}`
)

var (
	lowerWord     = regexp.MustCompile(`^\p{Ll}+$`)
	definiteForms = map[string]string{"noun": "definite", "verb": "present", "adj": "definite", "adv": ""}
	latinFolds    = strings.NewReplacer("æ", "ae", "ø", "o", "å", "a")
)

var shippedLocales = []shippedLocale{
	{
		name:      "da_DK",
		wordForms: definiteForms,
		word:      lowerWord,
		noun:      lowerWord,
		months:    []string{"januar", "februar", "marts", "april", "maj", "juni", "juli", "august", "september", "oktober", "november", "december"},
		weekdays:  []string{"mandag", "tirsdag", "onsdag", "torsdag", "fredag", "lørdag", "søndag"},
		everyday:  &everyday{`^\d{1,3}(\.\d{3})?(,\d{2})? kr\.$`, `^([01]\d|2[0-3])\.[0-5]\d$`, `^https://([a-z]+\.)?example\.(com|org|net|dk)` + urlPath + `$`, `^[a-z0-9._]{3,}$`, []string{"Blå", "Grøn", "Rød"}},
		person:    &person{map[string]string{"f": "kvinde", "m": "mand"}, []string{"dr.", "professor"}},
		// Statistics Denmark's table GF15, 2024: 132,403 sole proprietorships, 218,208 private limited companies,
		// 30,750 public limited companies and 11,679 partnerships and limited partnerships, split evenly.
		company: &company{regexp.MustCompile(`^\p{Lu}[\p{L}\p{N} &'.-]*[\p{L}\p{N}]( (ApS|A/S|I/S|K/S))?$`), []legalForm{
			{"ENK", "enk", "", 132403}, {"ApS", "aps", "ApS", 218208}, {"A/S", "as", "A/S", 30750}, {"I/S", "is", "I/S", 5840}, {"K/S", "ks", "K/S", 5839},
		}},
		// Dansk Sprognævn's answer SV00000046 gives 2.10.2003 and advises against 02.10.2003.
		dateLayout: "2.1.2006",
		emailFold:  latinFolds,
		phone:      &phone{"+45", false, danishFixedPhone},
	},
	{
		name:      "de_DE",
		wordForms: map[string]string{"noun": "plural", "verb": "", "adj": "", "adv": ""},
		word:      lowerWord,
		noun:      regexp.MustCompile(`^\p{Lu}\p{Ll}+$`),
		months:    []string{"Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "August", "September", "Oktober", "November", "Dezember"},
		weekdays:  []string{"Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag", "Sonntag"},
		everyday:  &everyday{`^\d{1,3}(\.\d{3})?,\d{2} €$`, `^([01]\d|2[0-3]):[0-5]\d$`, `^https://([a-z]+\.)?example\.(com|org|net|de)` + urlPath + `$`, `^[a-z0-9._]{3,}$`, []string{"Blau", "Grün", "Rot"}},
		person:    &person{map[string]string{"f": "Frau", "m": "Mann"}, []string{"Dr.", "Prof.", "Prof. Dr."}},
		// Destatis' Unternehmensregister, 2024: 2,062,752 sole proprietors, 440,951 partnerships and 861,720
		// corporations, each group split evenly over the forms drawn.
		company: &company{regexp.MustCompile(`^\p{Lu}[\p{L}\p{N} &'.-]*[\p{L}\p{N}]( (KG|OHG|GmbH|UG \(haftungsbeschränkt\)|AG))?$`), []legalForm{
			{"Einzelunternehmen", "einzel", "", 2062752}, {"KG", "kg", "KG", 220476}, {"OHG", "ohg", "OHG", 220475},
			{"GmbH", "gmbh", "GmbH", 287240}, {"UG (haftungsbeschränkt)", "ug", "UG (haftungsbeschränkt)", 287240}, {"AG", "ag", "AG", 287240},
		}},
		// DIN 5008 writes a date as TT.MM.JJJJ.
		dateLayout: "02.01.2006",
		emailFold:  strings.NewReplacer("ä", "a", "ö", "o", "ü", "u", "ß", "ss"),
		phone:      &phone{"+49", true, germanFixedPhone},
	},
	{
		name:      "en_US",
		wordForms: map[string]string{"noun": "", "verb": "present", "adj": "", "adv": ""},
		word:      regexp.MustCompile(`^[a-z]+$`),
		noun:      regexp.MustCompile(`^[a-z]+$`),
		months:    []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"},
		weekdays:  []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"},
	},
	{
		name:      "fi_FI",
		wordForms: map[string]string{"noun": "", "verb": "", "adj": "", "adv": ""},
		word:      lowerWord,
		noun:      lowerWord,
		months:    []string{"tammikuu", "helmikuu", "maaliskuu", "huhtikuu", "toukokuu", "kesäkuu", "heinäkuu", "elokuu", "syyskuu", "lokakuu", "marraskuu", "joulukuu"},
		weekdays:  []string{"maanantai", "tiistai", "keskiviikko", "torstai", "perjantai", "lauantai", "sunnuntai"},
		everyday:  &everyday{`^\d{1,3}( \d{3})?,\d{2} €$`, `^([01]\d|2[0-3])\.[0-5]\d$`, `^https://([a-z]+\.)?example\.(com|org|net|fi)` + urlPath + `$`, `^[a-z0-9._]{3,}$`, []string{"Sininen", "Vihreä", "Punainen"}},
		person:    &person{map[string]string{"f": "nainen", "m": "mies"}, []string{"tohtori", "professori"}},
		company: &company{regexp.MustCompile(`^\p{L}[\p{L}\p{N} &'.-]*[\p{L}\p{N}]( (Oy|Ky|Ay|osk))?$`), []legalForm{
			{"Tmi", "tmi", "", 295203}, {"Oy", "oy", "Oy", 240524}, {"Ky", "ky", "Ky", 13356}, {"Ay", "ay", "Ay", 4726}, {"osk", "osk", "osk", 2811},
		}},
		dateLayout: "2.1.2006",
		emailFold:  strings.NewReplacer("ä", "a", "ö", "o", "å", "a"),
		phone:      &phone{"+358", true, finnishFixedPhone},
	},
	{
		name:      "nb_NO",
		wordForms: definiteForms,
		word:      lowerWord,
		noun:      lowerWord,
		months:    []string{"januar", "februar", "mars", "april", "mai", "juni", "juli", "august", "september", "oktober", "november", "desember"},
		weekdays:  []string{"mandag", "tirsdag", "onsdag", "torsdag", "fredag", "lørdag", "søndag"},
		everyday:  &everyday{`^\d{1,3}( \d{3})?(,\d{2})? kr$`, `^([01]\d|2[0-3]):[0-5]\d$`, `^https://([a-z]+\.)?example\.(com|org|net|no)` + urlPath + `$`, `^[a-z0-9._]{3,}$`, []string{"Blå", "Grønn", "Rød"}},
		person:    &person{map[string]string{"f": "kvinne", "m": "mann"}, []string{"dr.", "professor"}},
		// The active units of each form in Enhetsregisteret on 2026-10-22, leaving out the bankrupt, those being
		// wound up and those forcibly dissolved.
		company: &company{regexp.MustCompile(`^\p{Lu}[\p{L}\p{N} &'-]*[\p{L}\p{N}]$`), []legalForm{
			{"ENK", "enk", "", 465166}, {"AS", "as", "AS", 423370}, {"DA", "da", "DA", 9105}, {"SA", "sa", "SA", 6761},
			{"ANS", "ans", "ANS", 6214}, {"ASA", "asa", "ASA", 205}, {"KS", "ks", "KS", 105},
		}},
		dateLayout: "02.01.2006",
		emailFold:  latinFolds,
		phone:      &phone{"+47", false, norwegianFixedPhone},
	},
	{
		name:      "nn_NO",
		wordForms: definiteForms,
		word:      lowerWord,
		noun:      lowerWord,
		months:    []string{"januar", "februar", "mars", "april", "mai", "juni", "juli", "august", "september", "oktober", "november", "desember"},
		weekdays:  []string{"måndag", "tysdag", "onsdag", "torsdag", "fredag", "laurdag", "søndag"},
		everyday:  &everyday{`^\d{1,3}( \d{3})?(,\d{2})? kr$`, `^([01]\d|2[0-3]):[0-5]\d$`, `^https://([a-z]+\.)?example\.(com|org|net|no)` + urlPath + `$`, `^[a-z0-9._]{3,}$`, []string{"Blå", "Grøn", "Raud"}},
	},
	{
		name:      "sv_SE",
		wordForms: definiteForms,
		word:      lowerWord,
		noun:      lowerWord,
		months:    []string{"januari", "februari", "mars", "april", "maj", "juni", "juli", "augusti", "september", "oktober", "november", "december"},
		weekdays:  []string{"måndag", "tisdag", "onsdag", "torsdag", "fredag", "lördag", "söndag"},
	},
}

func eachLocale(t *testing.T, has func(l shippedLocale) bool, test func(t *testing.T, l shippedLocale)) {
	t.Helper()
	for _, l := range shippedLocales {
		if has(l) {
			t.Run(l.name, func(t *testing.T) { test(t, l) })
		}
	}
}

var localeNamespace = regexp.MustCompile(`^[a-z]{2}_[A-Z]{2}$`)

func shippedLocaleModules(t *testing.T) []string {
	t.Helper()
	var names []string
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
		for category := range manifest.Index {
			if namespace, _, _ := strings.Cut(category, "."); localeNamespace.MatchString(namespace) && !slices.Contains(names, namespace) {
				names = append(names, namespace)
			}
		}
	}
	slices.Sort(names)
	return names
}

func TestShippedLocalesAreTheLocaleModulesOfModules(t *testing.T) {
	var listed []string
	for _, l := range shippedLocales {
		listed = append(listed, l.name)
	}
	if modules := shippedLocaleModules(t); !slices.Equal(listed, modules) {
		t.Errorf("shippedLocales holds %v, want one entry per locale module of data.Modules, in order: %v", listed, modules)
	}
}
