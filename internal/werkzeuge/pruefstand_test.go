package werkzeuge

// **Der Prüfstand im Fenster stimmt mit den Belegen** — OPS-07.AK-02/04
// (#1724):
// „Technische Kennzahlen werden aus einem aktuellen Beleg erzeugt, nicht
// manuell widersprüchlich fortgeschrieben."
//
// # Was er hält
//
//   - **Jeder Beleg liegt da** — eine Datei im Repositorium oder eine
//     Adresse außerhalb. **Seit 1.0.0 kein Issue**: die Issues bis dahin
//     liegen im privaten Entwicklungsarchiv (Vermerk 0205), ein Beleg, den
//     Anwendende nicht öffnen können, belegt ihnen nichts — und die Adresse
//     eines neuen sähe genauso aus.
//   - **Die fremden Dienste stehen wie in `docs/HAERTUNG.md`**: wer dort „nie
//     gegen das Original gelaufen" ist, heißt im Fenster nicht geprüft.
//
// # Was er nicht hält
//
// **Ob ein Beleg trägt.** Er sieht, dass die Datei da ist, nicht dass sie
// das Behauptete belegt — das bleibt eine Frage ans Review (§10).

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestJederPruefpunktHatSeinenBeleg(t *testing.T) {
	t.Parallel()

	staende := []string{app.StandGeprueft, app.StandVorhanden, app.StandBeschrieben, app.StandBlockiert}

	for _, punkt := range app.Pruefstand() {
		if !slices.Contains(staende, punkt.Stand) {
			t.Errorf("%s: Stand %q", punkt.Gegenstand, punkt.Stand)
		}

		if strings.HasPrefix(punkt.Beleg, "https://github.com/DavidTHeider/MagicSetEditor4Mac/issues/") {
			t.Errorf("%s: der Beleg %q ist ein Issue — seit 1.0.0 belegt eine Datei im Repositorium",
				punkt.Gegenstand, punkt.Beleg)

			continue
		}

		if strings.HasPrefix(punkt.Beleg, "https://") {
			continue
		}

		if _, fehler := os.Stat(filepath.Join(wurzel, punkt.Beleg)); fehler != nil {
			t.Errorf("%s: der Beleg %q fehlt", punkt.Gegenstand, punkt.Beleg)
		}
	}
}

func TestDieFremdenDiensteStehenWieInDerHaertung(t *testing.T) {
	t.Parallel()

	tafel := lies(t, "docs/HAERTUNG.md")
	_, abschnitt, _ := strings.Cut(tafel, "## 4 ·")
	abschnitt, _, _ = strings.Cut(abschnitt, "\n## ")

	gesehen := 0

	for _, punkt := range app.Pruefstand() {
		if punkt.Bereich != app.BereichIntegration {
			continue
		}

		zeile := ""
		for _, kandidat := range strings.Split(abschnitt, "\n") {
			zellen := strings.Split(kandidat, "|")
			if len(zellen) > 2 && strings.Contains(strings.ReplaceAll(zellen[1], "**", ""), punkt.Gegenstand) {
				zeile = zellen[2]
			}
		}

		if zeile == "" {
			t.Errorf("%s steht nicht in docs/HAERTUNG.md §4", punkt.Gegenstand)

			continue
		}

		gesehen++

		erwartet := app.StandGeprueft
		if strings.Contains(zeile, "**nie**") {
			erwartet = app.StandVorhanden
		}

		if punkt.Stand != erwartet {
			t.Errorf("%s: im Fenster %q, in der Härtung %q", punkt.Gegenstand, punkt.Stand, erwartet)
		}
	}

	if gesehen < 10 {
		t.Errorf("nur %d Dienste gesehen — liest der Wächter die Tafel?", gesehen)
	}
}

// **„Geprüft" heißt: ein Lauf hat es belegt** — Nachlauf über Phase 6.4,
// N10. Hier genügte jede Datei, auch eine Testquelle; §3 sagt, ein Test
// belegt sich mit seinem Lauf. Zwei Formen gelten:
//
//   - **ein Beleg unter `docs/belege/`, der seinen Commit nennt** — eine
//     Probe, die einen bestimmten Stand gesehen hat,
//   - **ein Test, den jede Pipeline fährt** — veröffentlicht wird nur ein
//     Commit, dessen `make ci` grün war (§4), also ist sein Lauf belegt.
//
// **Was er nicht hält**: ob der Beleg zum **heutigen** Stand passt. Ein
// älterer Commit kann noch gelten; das entscheidet das Review.
func TestEinGepruefterPunktBelegtEinenLauf(t *testing.T) {
	t.Parallel()

	for _, punkt := range app.Pruefstand() {
		if punkt.Stand != app.StandGeprueft || strings.HasPrefix(punkt.Beleg, "https://") || punkt.Beleg == "docs/HAERTUNG.md" {
			continue
		}

		if strings.Contains(punkt.Beleg, ".test.") || strings.HasSuffix(punkt.Beleg, "_test.go") {
			continue
		}

		if !strings.HasPrefix(punkt.Beleg, "docs/belege/") {
			t.Errorf("%s: %q ist weder ein Beleg eines Laufs noch ein Test der Pipeline", punkt.Gegenstand, punkt.Beleg)

			continue
		}

		if inhalt := lies(t, punkt.Beleg); !strings.Contains(inhalt, "Commit `") && !strings.Contains(inhalt, "| Stand | `") {
			t.Errorf("%s: %q nennt keinen Commit", punkt.Gegenstand, punkt.Beleg)
		}
	}
}

// **Jede Sprache des Fensters steht im Prüfstand** — J01
// (#2265).
// Bis zum 2026-10-11 nannte er „Deutsch, English" geprüft und „Weitere
// Sprachen" beschrieben, mit einem Issue als Beleg, das seit LOC-01
// geschlossen ist — sechs Kataloge standen längst da.
//
// **Was er nicht hält**: ob der Stand je Sprache stimmt; er sieht nur, dass
// jede genannt ist.
func TestJedeSpracheDesFenstersStehtImPruefstand(t *testing.T) {
	t.Parallel()

	endonyme := map[string]string{
		"de": "Deutsch", "en": "English", "es": "Español", "fr": "Français",
		"it": "Italiano", "ja": "日本語", "ru": "Русский", "zh": "中文",
	}

	eintraege, fehler := os.ReadDir(filepath.Join(wurzel, "frontend", "src", "sprache"))
	if fehler != nil {
		t.Fatal(fehler)
	}

	genannt := ""
	for _, punkt := range app.Pruefstand() {
		if punkt.Bereich == app.BereichSprache {
			genannt += punkt.Gegenstand + "\n"
		}
	}

	gesehen := 0
	for _, eintrag := range eintraege {
		if !eintrag.IsDir() || len(eintrag.Name()) != 2 {
			continue
		}

		gesehen++
		endonym, bekannt := endonyme[eintrag.Name()]

		switch {
		case !bekannt:
			t.Errorf("der Katalog %q ist neu — trage seinen Namen hier und im Prüfstand ein", eintrag.Name())
		case !strings.Contains(genannt, endonym):
			t.Errorf("%s (%s) hat einen Katalog, steht aber nicht im Prüfstand", endonym, eintrag.Name())
		}
	}

	if gesehen < 2 {
		t.Errorf("nur %d Kataloge gesehen — liest der Wächter das Verzeichnis?", gesehen)
	}
}
