// **Jede schreibende Bindung hat eine Klasse, jeder genannte Lauf ist da** —
// C04 (#2229).
//
// # Was er hält
//
//  1. Jede schreibende Bindung aus `SCHREIBWEGE.md` steht in **genau einer**
//     Zeile von `FLAECHENPARITAET.md`, und keine Zeile nennt eine, die es
//     nicht gibt. Eine neue schreibende Fähigkeit ohne Klasse ist rot: sie
//     hätte keinen Vertrag über die Flächen.
//  2. Jeder dort genannte Lauf existiert in `internal/flaechen`.
//  3. Die Bilanz stimmt — Nenner, Klassen nur am Fenster und die drei Zähler.
//  4. **Die Rolle entfällt genau dort, wo kein Profil Träger ist** — und das
//     entscheidet nicht die Tabelle, sondern die Kataloge: Bindung →
//     Fähigkeit in `FAEHIGKEITEN.md`, Fähigkeit → Träger in
//     `AUTORISIERUNG.md`. Eine Klasse, deren Bindungen nur dem Nutzer oder
//     dem Rechner gehören, hat keine Rolle zu vergleichen; eine mit einem
//     Profilträger darf sich nicht herausnehmen. Der Rollennenner zählt nur
//     die Klassen mit Profilträger.
//
// # Was er nicht hält
//
// **Ob ein Lauf vergleicht, was seine Spalte sagt** — das liest ein Mensch.
// Und ob eine Klasse an einer Fläche fehlen **darf**: das steht mit Grund in
// `FAEHIGKEITEN.md`, und dort hält es sein eigener Wächter.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

var (
	flaechenparitaetszeile = regexp.MustCompile(`(?m)^\| ([^|]+?) \| \d+: ([^|]*) \| ([^|]*) \| ([^|]*) \| ([^|]*) \| ([^|]*) \|$`)
	paritaetsname          = regexp.MustCompile("`([A-Za-z0-9]+)`")
	paritaetsbilanz        = regexp.MustCompile(`(?m)^\| (Ergebnis|Fehlerklasse|Rolle) \| (\d+) von (\d+) \|`)
	paritaetskopf          = regexp.MustCompile(`Von \*\*(\d+)\*\* Klassen gibt es \*\*(\d+)\*\* nur am Fenster; der Nenner\s+für den Vergleich über Flächen ist \*\*(\d+)\*\*`)
)

func TestJedeSchreibendeBindungHatIhreKlasseUndJederLaufIstDa(t *testing.T) {
	t.Parallel()

	paritaet := lies(t, "docs/FLAECHENPARITAET.md")
	schreibend := map[string]bool{}

	for _, treffer := range aufzaehlung.FindAllStringSubmatch(absatzNach(t, lies(t, "docs/SCHREIBWEGE.md"), "schreibenden Bindungen"), -1) {
		schreibend[treffer[1]] = true
	}

	laeufe := testfunktionenIn(t, filepath.Join("..", "flaechen"))
	traeger := bindungstraeger(t)
	gesehen := map[string]string{}
	zaehler := map[string]int{}
	klassen, nurFenster, mitProfil := 0, 0, 0

	for _, zeile := range flaechenparitaetszeile.FindAllStringSubmatch(paritaet, -1) {
		klasse := strings.TrimSpace(zeile[1])
		klassen++

		klassentraeger := map[string]bool{}

		for _, bindung := range paritaetsname.FindAllStringSubmatch(zeile[2], -1) {
			for _, wer := range traeger[bindung[1]] {
				klassentraeger[wer] = true
			}

			switch {
			case !schreibend[bindung[1]]:
				t.Errorf("%s nennt %q — keine schreibende Bindung", klasse, bindung[1])
			case gesehen[bindung[1]] != "":
				t.Errorf("%q steht bei %s und bei %s", bindung[1], gesehen[bindung[1]], klasse)
			}

			gesehen[bindung[1]] = klasse
		}

		pruefeRollenspalte(t, klasse, strings.TrimSpace(zeile[5]), klassentraeger)

		if strings.HasPrefix(strings.TrimSpace(zeile[6]), "**nur am Fenster**") {
			nurFenster++

			continue
		}

		if klassentraeger["Profil"] {
			mitProfil++
		}

		for spalte, gegenstand := range map[int]string{3: "Ergebnis", 4: "Fehlerklasse", 5: "Rolle"} {
			genannt := paritaetsname.FindAllStringSubmatch(zeile[spalte], -1)
			for _, lauf := range genannt {
				if !laeufe[lauf[1]] {
					t.Errorf("%s nennt den Lauf %q — es gibt ihn nicht", klasse, lauf[1])
				}
			}

			if len(genannt) > 0 {
				zaehler[gegenstand]++
			}
		}
	}

	for bindung := range schreibend {
		if gesehen[bindung] == "" {
			t.Errorf("die schreibende Bindung %q hat keine Klasse in FLAECHENPARITAET.md — "+
				"ohne Klasse hat sie keinen Vertrag über die Flächen", bindung)
		}
	}

	pruefeParitaetsbilanz(t, paritaet, klassen, nurFenster, zaehler, map[string]int{
		"Ergebnis": klassen - nurFenster, "Fehlerklasse": klassen - nurFenster, "Rolle": mitProfil,
	})
}

// pruefeRollenspalte verlangt „entfällt" genau bei einer Klasse ohne
// Profilträger, und dann mit den Trägern, die die Kataloge nennen.
func pruefeRollenspalte(t *testing.T, klasse, rolle string, traeger map[string]bool) {
	t.Helper()

	entfaellt := strings.HasPrefix(rolle, "entfällt")

	switch {
	case traeger["Profil"] && entfaellt:
		t.Errorf("%s: die Rolle entfällt, aber eine ihrer Bindungen gehört einem Profil", klasse)
	case !traeger["Profil"]:
		namen := make([]string, 0, len(traeger))
		for wer := range traeger {
			namen = append(namen, wer)
		}

		slices.Sort(namen)

		if will := "entfällt — Träger " + strings.Join(namen, ", "); rolle != will {
			t.Errorf("%s: die Rollenspalte nennt %q, die Kataloge sagen %q", klasse, rolle, will)
		}
	}
}

// bindungstraeger liest je Bindung des Fensters, wem ihre Fähigkeit gehört:
// Bindung → Fähigkeit aus `FAEHIGKEITEN.md`, Fähigkeit → Träger aus
// `AUTORISIERUNG.md`.
func bindungstraeger(t *testing.T) map[string][]string {
	t.Helper()

	jeFaehigkeit := map[string]string{}
	for _, zeile := range matrixzeilen(t) {
		jeFaehigkeit[zeile.faehigkeit] = zeile.traeger
	}

	traeger := map[string][]string{}

	for _, zeile := range faehigkeitszeilen(t) {
		spalten := strings.Split(zeile, "|")
		wer := jeFaehigkeit[strings.TrimSpace(spalten[1])]

		for _, bindung := range paritaetsname.FindAllStringSubmatch(spalten[2], -1) {
			traeger[bindung[1]] = append(traeger[bindung[1]], wer)
		}
	}

	return traeger
}

// pruefeBilanz hält Kopf und Zähler der Bilanz gegen die Tabelle.
func pruefeParitaetsbilanz(t *testing.T, paritaet string, klassen, nurFenster int, zaehler, nenner map[string]int) {
	t.Helper()

	kopf := paritaetskopf.FindStringSubmatch(paritaet)
	if kopf == nil {
		t.Fatal("die Bilanz nennt Klassen und Nenner nicht mehr — der Wächter greift daneben")
	}

	for stelle, will := range map[int]int{1: klassen, 2: nurFenster, 3: klassen - nurFenster} {
		if steht := bilanzzahl(t, kopf[stelle]); steht != will {
			t.Errorf("die Bilanz nennt %d, gezählt sind %d (%q)", steht, will, kopf[0])
		}
	}

	for _, zeile := range paritaetsbilanz.FindAllStringSubmatch(paritaet, -1) {
		steht, von := bilanzzahl(t, zeile[2]), bilanzzahl(t, zeile[3])

		if steht != zaehler[zeile[1]] || von != nenner[zeile[1]] {
			t.Errorf("%s: die Bilanz nennt %d von %d, gezählt sind %d von %d",
				zeile[1], steht, von, zaehler[zeile[1]], nenner[zeile[1]])
		}
	}
}

// laeufeIn sammelt die Namen der Testfunktionen eines Verzeichnisses.
func testfunktionenIn(t *testing.T, ordner string) map[string]bool {
	t.Helper()

	eintraege, err := os.ReadDir(ordner)
	if err != nil {
		t.Fatalf("%s lesen: %v", ordner, err)
	}

	laeufe := map[string]bool{}
	testfunktion := regexp.MustCompile(`(?m)^func (Test\w+)\(`)

	for _, eintrag := range eintraege {
		if !strings.HasSuffix(eintrag.Name(), "_test.go") {
			continue
		}

		inhalt, fehler := os.ReadFile(filepath.Join(ordner, eintrag.Name())) //nolint:gosec // ein Verzeichnis dieses Repositoriums
		if fehler != nil {
			t.Fatalf("%s lesen: %v", eintrag.Name(), fehler)
		}

		for _, treffer := range testfunktion.FindAllStringSubmatch(string(inhalt), -1) {
			laeufe[treffer[1]] = true
		}
	}

	return laeufe
}

// bilanzzahl liest eine Zahl der Bilanz — der Ausdruck verlangt Ziffern, und
// eine, die trotzdem nicht passt, ist ein Fehler des Wächters selbst.
func bilanzzahl(t *testing.T, text string) int {
	t.Helper()

	zahl, fehler := strconv.Atoi(text)
	if fehler != nil {
		t.Fatalf("keine Zahl in der Bilanz: %q", text)
	}

	return zahl
}
