// **Eine Naht hängt an jeder Wurzel, die den Dienst baut** — Werkstück von
// Phase 4.10
// (#1376).
//
// # Der Anlass
//
// `internal/app` hat **nicht eine** Composition Root, sondern viele: eine je
// Fläche, dazu die des Fensters (§2). Wer einem Dienst eine `Mit…`-Naht gibt
// und sie an einer Wurzel verdrahtet, hat die Fähigkeit **an einer Fläche**
// gebaut — und die anderen antworten weiter anders (§16).
//
// Zweimal an einem Tag:
//
//   - `Deckbau.MitKartennamen` und `MitDrucknamen` hingen an `deckstart.go`.
//     Der Dienst war grün, das Fenster zeigte Namen — Kommandozeile, MCP und
//     REST zeigten weiter UUIDs.
//   - `Kartenlistenbau.MitMarken` hing an `start.go`. `ListenzeileDTO.Marken`
//     gibt es an jeder Fläche, und über die drei anderen kam das Feld leer.
//
// Beide Male waren **beide Seiten grün**: der Dienst konnte es, die Fläche
// reichte das Feld durch. Nur die Naht dazwischen fehlte.
//
// # Was er prüft
//
// Für jeden Dienst aus `internal/service` mit `Mit…`-Nähten, der in mehr als
// einer Datei unter `internal/app` gebaut wird: **jede** Wurzel ruft dieselben
// Nähte — oder der Dienst steht in [nahtausnahmen] mit einem Grund.
//
// # Was er nicht prüft, und das ist die Hälfte
//
//   - **Nicht, ob eine Naht überhaupt irgendwo hängt.** Ein Dienst, der nur an
//     einer Stelle gebaut wird, kommt hier nie vor; dafür steht der
//     Aufrufer-Wächter daneben (`bindungsaufrufer_test.go`).
//   - **Nicht, ob die verdrahtete Naht dasselbe *tut*.** Zwei Wurzeln dürfen
//     verschiedene Ablagen einsetzen; dass beide `MitMarken` rufen, heißt
//     nicht, dass beide dieselbe Tabelle lesen. Dafür gibt es je Fähigkeit
//     einen Lauf über alle Flächen (§16).
//   - **Nichts über `cmd/`.** Wer dort selbst zusammensetzt, umgeht ihn — was
//     der Schichtenwächter ohnehin verbietet.
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// nahtausnahmen sind Dienste, deren Nähte bewusst nur an einer Wurzel hängen.
//
// **Ein Eintrag ist eine Entscheidung und kein Nebeneffekt** (§16): er nennt
// den Grund, und der ist nachgelesen und nicht plausibel.
var nahtausnahmen = map[string]string{
	"Vorpruefer": "die drei Fragen des Augenblicks — Platz, laufende Arbeit " +
		"und eine **zurückgelesene** Sicherung — gehören zum Vorbereiten " +
		"eines Wechsels, und das tut nur das Fenster (O6, #1430). Es hält " +
		"den Bestand dafür an und schreibt eine Sicherung; eine " +
		"Kommandozeile, die das auf Zuruf täte, nähme ihn jemandem weg, der " +
		"gerade arbeitet. **Die Fassungsfrage selbst hat überall dieselbe " +
		"Antwort** (§16): `Vorpruefer.Urteil` hängt allein am Ledger, und den " +
		"verdrahten beide Wurzeln gleich.",
	"Symbolbibliothek": "MitFassungen speist allein die Vorschau eines " +
		"Symbols in den Fassungen eines Sets, und die gibt es **nur am " +
		"Fenster** (DOM-02.AK-02, #1697): drei PNG je Fassung liest in einer " +
		"Textantwort niemand — die Abwägung steht in docs/FAEHIGKEITEN.md. " +
		"Keine Fläche ruft Fassungsplaene; die Naht dort zu setzen hieße, " +
		"einen Stilwähler zu bauen, den niemand fragt.",
	"Stilwaehler": "MitSpeicher räumt den Stil eines Sets, sobald er " +
		"gewechselt ist — ein Eintrag weniger im Gedächtnis des Fensters. " +
		"**Nötig ist es nicht mehr**: der Stilspeicher liest den Setkopf je " +
		"Ruf und sieht einen Wechsel an jeder Fläche beim nächsten Bild, auch " +
		"einen aus einem anderen Prozess (#2005, " +
		"TestEinStilwechselWirktSofort). Die Flächen ohne Fenster behalten " +
		"ihr Zeichenwerk seither, verdrahten den Wähler aber über " +
		"stilwaehlerMit, wo es keines gibt.",
	"Stilspeicher": "MitPaketwechsel meldet, wenn ein gemerkter Stil an " +
		"einem geänderten Paket verfällt, damit das Zeichenwerk Bildquelle " +
		"und Ebenen räumt (#2005). Auszählung und Kartenskripte bauen ihren " +
		"Stilspeicher je Ruf und halten **keine Bilder** daneben — es gäbe " +
		"dort nichts zu räumen, und ein Speicher, der den Ruf nicht " +
		"überlebt, kann nicht veralten.",
}

func TestJedeWurzelVerdrahtetDieselbenNaehte(t *testing.T) {
	naehte, konstruktoren := dienstnaehte(t)
	gebaut := wurzelnJeDienst(t, konstruktoren, naehte)

	for _, dienst := range sortierteSchluessel(gebaut) {
		wurzeln := gebaut[dienst]
		if len(wurzeln) < 2 {
			continue
		}

		gerufen := map[string][]string{}
		for wurzel, gesetzt := range wurzeln {
			gerufen[strings.Join(gesetzt, " ")] = append(
				gerufen[strings.Join(gesetzt, " ")], wurzel)
		}

		if len(gerufen) == 1 {
			continue
		}

		if grund, befreit := nahtausnahmen[dienst]; befreit {
			t.Logf("%s ist ausgenommen: %s", dienst, grund)

			continue
		}

		t.Errorf("%s wird an %d Wurzeln gebaut, und sie verdrahten Verschiedenes:",
			dienst, len(wurzeln))

		for _, wurzel := range sortierteSchluessel(wurzeln) {
			fehlt := []string{}

			for _, naht := range naehte[dienst] {
				if !enthaelt(wurzeln[wurzel], naht) {
					fehlt = append(fehlt, naht)
				}
			}

			t.Errorf("    %s: es fehlt %v", wurzel, fehlt)
		}
	}
}

// **Und die Ausnahmeliste führt keinen Dienst, den es nicht mehr gibt.**
//
// Ein Eintrag, dessen Dienst umbenannt wurde, sähe aus wie eine Entscheidung
// und wäre ein Rest — und er befreite lautlos nichts mehr (§10).
func TestDieNahtausnahmenNennenNurDiensteDieEsGibt(t *testing.T) {
	naehte, _ := dienstnaehte(t)

	for dienst := range nahtausnahmen {
		if _, da := naehte[dienst]; !da {
			t.Errorf("%s steht in nahtausnahmen und hat keine Mit…-Naht mehr", dienst)
		}
	}
}

// dienstnaehte liest je Diensttyp seine Nähte und seine Konstruktoren.
func dienstnaehte(t *testing.T) (naehte map[string][]string, konstruktoren map[string]string) {
	t.Helper()

	var (
		nahtform        = regexp.MustCompile(`(?m)^func \(\w+ \*?(\w+)\) (Mit[A-ZÄÖÜ]\w*)\(`)
		konstruktorform = regexp.MustCompile(`(?m)^func (Neue[rs]?\w+)\((?:[^{]|\n)*?\)\s*\(?\*(\w+)`)
	)

	naehte = map[string][]string{}
	konstruktoren = map[string]string{}

	for _, datei := range gofDateien(t, filepath.Join("..", "service")) {
		inhalt := quelltext(t, datei)

		for _, treffer := range nahtform.FindAllStringSubmatch(inhalt, -1) {
			if !enthaelt(naehte[treffer[1]], treffer[2]) {
				naehte[treffer[1]] = append(naehte[treffer[1]], treffer[2])
			}
		}

		for _, treffer := range konstruktorform.FindAllStringSubmatch(inhalt, -1) {
			konstruktoren[treffer[1]] = treffer[2]
		}
	}

	if len(naehte) == 0 {
		t.Fatal("kein Dienst mit einer Mit…-Naht gefunden — der Wächter misst nichts")
	}

	for dienst := range naehte {
		sort.Strings(naehte[dienst])
	}

	return naehte, konstruktoren
}

// wurzelnJeDienst sagt je Dienst, welche Datei in `app` welche Nähte ruft.
func wurzelnJeDienst(
	t *testing.T, konstruktoren map[string]string, naehte map[string][]string,
) map[string]map[string][]string {
	t.Helper()

	gebaut := map[string]map[string][]string{}

	for _, datei := range gofDateien(t, filepath.Join("..", "app")) {
		inhalt := quelltext(t, datei)
		name := filepath.Base(datei)

		for konstruktor, dienst := range konstruktoren {
			if _, hatNaehte := naehte[dienst]; !hatNaehte {
				continue
			}

			if !regexp.MustCompile(`service\.` + konstruktor + `\(`).MatchString(inhalt) {
				continue
			}

			gesetzt := []string{}

			for _, naht := range naehte[dienst] {
				if regexp.MustCompile(`\.\s*` + naht + `\(`).MatchString(inhalt) {
					gesetzt = append(gesetzt, naht)
				}
			}

			if gebaut[dienst] == nil {
				gebaut[dienst] = map[string][]string{}
			}

			gebaut[dienst][name] = gesetzt
		}
	}

	return gebaut
}

// gofDateien listet die Go-Dateien eines Verzeichnisses ohne Tests.
func gofDateien(t *testing.T, verzeichnis string) []string {
	t.Helper()

	eintraege, err := os.ReadDir(verzeichnis)
	if err != nil {
		t.Fatalf("%s lesen: %v", verzeichnis, err)
	}

	dateien := []string{}

	for _, eintrag := range eintraege {
		name := eintrag.Name()
		if eintrag.IsDir() || !strings.HasSuffix(name, ".go") ||
			strings.HasSuffix(name, "_test.go") {
			continue
		}

		dateien = append(dateien, filepath.Join(verzeichnis, name))
	}

	return dateien
}

// quelltext gibt den Inhalt einer Datei.
//
// **Nicht [lies]**: das liest relativ zur Wurzel des Repositoriums, hier
// stehen fertige Pfade aus [gofDateien].
func quelltext(t *testing.T, pfad string) string {
	t.Helper()

	inhalt, err := os.ReadFile(pfad) //nolint:gosec // eigener Baum
	if err != nil {
		t.Fatalf("%s lesen: %v", pfad, err)
	}

	return string(inhalt)
}

// sortierteSchluessel gibt die Schlüssel einer Abbildung geordnet.
func sortierteSchluessel[W any](abbildung map[string]W) []string {
	schluessel := make([]string, 0, len(abbildung))
	for eintrag := range abbildung {
		schluessel = append(schluessel, eintrag)
	}

	sort.Strings(schluessel)

	return schluessel
}
