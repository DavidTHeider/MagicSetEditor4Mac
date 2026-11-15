package werkzeuge_test

// **Was als „gegen sein Original gelaufen" gilt, hat einen Lauf** — Werkstück 8
// von Phase 4.8 (#1321),
// seit Werkstück A1 von Phase 5.3
// (#1440)
// gegen den **Katalog in `domain`** statt gegen eine Liste im Fenster.
//
// # Warum es diesen Wächter gibt
//
// `frontend/src/bausteine/Dienstschalter.tsx` führte eine Liste der Dienste,
// die hier gegen ihr Original gelaufen sind. Sie stand von Hand da — und war
// **falsch**: magicthegathering.io fehlte, ausgerechnet der Dienst, dessen
// Lauf beim allerersten Mal die Drossel angestrichen hat (dokumentiert 5.000
// Anfragen je Stunde, gemessen 1.000).
//
// **Gefunden hat es die Sichtprüfung am laufenden Fenster**, nicht ein Test.
// Beide Seiten waren grün: das Fenster zeigte, was in seiner Liste stand, und
// der Netzlauf lief. Zwei grüne Seiten einer Grenze sind kein Nachweis für
// die Grenze (§10).
//
// # Was er hält
//
// Die Dienste mit einem [domain.Livenachweis] im Katalog sind **genau** die
// Pakete mit einer Datei hinter dem Build-Tag `netzlauf` — bis auf die, deren
// Lauf bereitsteht und mangels Zugang nie gelaufen ist ([bereitOhneNachweis],
// seit #2049); deren Lauf muss ohne Zugang BLOCKED melden. Beide Richtungen:
// ein Dienst ohne Lauf darf keinen Nachweis tragen, und einer mit Lauf soll
// einen. Dazu: die Datei, die der Nachweis nennt, liegt da und trägt den Tag.
//
// # Was er nicht hält
//
// **Ob der Lauf etwas taugt, und ob Datum und Fassung stimmen.** Er sieht
// eine Datei, nicht ihren Inhalt — ein `netzlauf_test.go` mit einem leeren
// Test käme durch, und ein Datum, das niemand nachgetragen hat, auch. Was er
// ausschließt, ist die stille Drift zwischen Katalog und Läufen.

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// dienstpakete ordnet einen Einstellungsschlüssel seinem Paket zu.
//
// **Von Hand, und das ist hier richtig**: der Schlüssel ist ein Begriff der
// Fachlichkeit, der Paketname einer der Ablage. Wer einen Dienst aufnimmt,
// trägt beides ein — und der Lauf unten sagt ihm, wenn er es vergisst.
var dienstpakete = map[string]string{
	"dienst.scryfall": "scryfall",
	"dienst.mtgapi":   "mtgapi",
	"dienst.mtgjson":  "mtggraphql",
	"dienst.scrydex":  "scrydex",

	// **CardTrader, Anthropic, OpenAI und xAI** (A4, A8, A9, A10): ein Lauf
	// gegen das Original braucht ein Konto samt Zugang, bei den drei
	// Cloudanbietern eines mit Guthaben. Seit #2049 steht er bereit und ist
	// nie gelaufen — siehe [bereitOhneNachweis]. Die Einträge stehen hier
	// trotzdem: ohne sie prüfte der Wächter die Dienste gar nicht, und genau
	// das ist die Gestalt aus §10, in der eine zu kurze Aufzählung nichts
	// anstreicht.
	"dienst.cardtrader": "cardtrader",
	"dienst.anthropic":  "anthropic",
	"dienst.openai":     "openai",
	"dienst.xai":        "xai",
}

// bereitOhneNachweis sind die Pakete, deren Lauf gegen das Original
// **bereitsteht und nie gelaufen ist** — #2049: ohne den Zugang des
// Auftraggebers endet er mit BLOCKED, nie mit PASS.
//
// **Eine Datei hinter `netzlauf` ist hier kein Nachweis.** Bis #2049 galt das
// Gegenteil, denn ein Lauf ohne Zugang lag gar nicht erst da. Wer einen
// dieser Läufe mit Zugang fährt und den Nachweis in den Katalog trägt,
// streicht das Paket hier — der Wächter sagt es ihm.
var bereitOhneNachweis = map[string]string{
	"mtggraphql": "Marke (Patreon) des Auftraggebers",
	"scrydex":    "Schlüssel und Team des Auftraggebers",
	"cardtrader": "Marke des Auftraggebers",
	"anthropic":  "Schlüssel des Auftraggebers",
	"openai":     "Schlüssel des Auftraggebers",
	"xai":        "Schlüssel des Auftraggebers",
}

// bautag erkennt den Bautag am Zeilenanfang.
var bautag = regexp.MustCompile(`(?m)^//go:build netzlauf`)

func TestJederGeprueftDienstHatEinenNetzlauf(t *testing.T) {
	t.Parallel()

	genannt := gepruefteAusDemKatalog(t)
	mitLauf := paketeMitNetzlauf(t)

	for _, schluessel := range genannt {
		paket, kennt := dienstpakete[schluessel]
		if !kennt {
			t.Errorf("der Katalog nennt %q als geprüft — dazu gibt es kein "+
				"Paket in dienstpakete", schluessel)

			continue
		}

		if !slices.Contains(mitLauf, paket) {
			t.Errorf("der Katalog nennt %q als geprüft, aber internal/%s hat "+
				"keine Datei hinter dem Tag netzlauf", schluessel, paket)
		}
	}

	// **Und jeder Dienst, den Go kennt, steht in dienstpakete.** Ohne diese
	// Schleife verschwände ein neuer Dienst still: er stünde in keiner der
	// beiden Listen, und der Wächter bliebe grün — die Gestalt aus §10, in
	// der eine zu kurze Aufzählung nichts anstreicht.
	//
	// **Über `Dienste()` und nicht über das Präfix `dienst.`** — Werkstück 3
	// von A18 (#1533).
	// Seit die Adresse je Dienst einstellbar ist, trägt die Reihe auch
	// `dienst.openai.adresse`, und die ist **kein Dienst**: sie sagt, wo
	// einer erreicht wird. Ein Präfixvergleich verlangte für sie ein
	// Netzlaufpaket, das es nicht geben kann.
	for _, dienst := range domain.Dienste() {
		schluessel := domain.DienstEinstellung(dienst)

		if _, kennt := dienstpakete[schluessel]; !kennt {
			t.Errorf("domain kennt den Dienst %q, dienstpakete nicht — ohne "+
				"Eintrag prüft dieser Wächter ihn gar nicht", schluessel)
		}
	}

	// **Und die andere Richtung**: ein Dienst mit Lauf soll auch als geprüft
	// dastehen. Ohne sie könnte jemand die Liste leeren, und der Wächter
	// bliebe grün.
	for schluessel, paket := range dienstpakete {
		if !slices.Contains(mitLauf, paket) {
			continue
		}

		if _, bereit := bereitOhneNachweis[paket]; bereit {
			pruefeBereit(t, paket, slices.Contains(genannt, schluessel))

			continue
		}

		if !slices.Contains(genannt, schluessel) {
			t.Errorf("internal/%s hat einen Netzlauf, aber der Katalog nennt "+
				"%q nicht als geprüft", paket, schluessel)
		}
	}
}

// pruefeBereit hält einen bereitstehenden Lauf ehrlich: er meldet ohne
// Zugang BLOCKED, und der Katalog nennt ihn nicht als geprüft — sonst gehört
// er nicht mehr in [bereitOhneNachweis].
func pruefeBereit(t *testing.T, paket string, genannt bool) {
	t.Helper()

	if genannt {
		t.Errorf("internal/%s trägt einen Nachweis im Katalog und steht in "+
			"bereitOhneNachweis — den Eintrag dort streichen", paket)
	}

	roh, err := os.ReadFile(filepath.Join("..", "..", "internal", paket, "netzlauf_test.go")) //nolint:gosec // aufgezählt
	if err != nil || !strings.Contains(string(roh), `t.Fatal("BLOCKED: `) {
		t.Errorf("internal/%s steht in bereitOhneNachweis, aber sein "+
			"netzlauf_test.go endet ohne Zugang nicht mit BLOCKED (%v)", paket, err)
	}
}

// gepruefteAusDemKatalog nennt die Dienste, die im Katalog einen Nachweis
// tragen — und hält jeden genannten Lauf gegen die Platte.
func gepruefteAusDemKatalog(t *testing.T) []string {
	t.Helper()

	var genannt []string

	for _, vertrag := range domain.Dienstvertraege() {
		if vertrag.Nachweis == nil {
			continue
		}

		schluessel := domain.DienstEinstellung(vertrag.Dienst)
		if !slices.Contains(genannt, schluessel) {
			genannt = append(genannt, schluessel)
		}

		roh, err := os.ReadFile(filepath.Join("..", "..", vertrag.Nachweis.Lauf))
		if err != nil {
			t.Errorf("der Nachweis von %s/%s nennt %q — die Datei fehlt: %v",
				vertrag.Dienst, vertrag.Frage, vertrag.Nachweis.Lauf, err)

			continue
		}

		if !bautag.Match(roh) {
			t.Errorf("der Nachweis von %s/%s nennt %q — die Datei steht nicht "+
				"hinter dem Bautag netzlauf", vertrag.Dienst, vertrag.Frage, vertrag.Nachweis.Lauf)
		}
	}

	return genannt
}

// paketeMitNetzlauf gibt die Pakete unter `internal/` mit einem Netzlauf.
func paketeMitNetzlauf(t *testing.T) []string {
	t.Helper()

	var mitLauf []string

	wurzel := filepath.Join("..", "..", "internal")

	pakete, err := os.ReadDir(wurzel)
	if err != nil {
		t.Fatalf("internal lesen: %v", err)
	}

	for _, paket := range pakete {
		if !paket.IsDir() {
			continue
		}

		dateien, lesefehler := os.ReadDir(filepath.Join(wurzel, paket.Name()))
		if lesefehler != nil {
			continue
		}

		for _, datei := range dateien {
			if !strings.HasSuffix(datei.Name(), "_test.go") {
				continue
			}

			inhalt, dateifehler := os.ReadFile( //nolint:gosec // unter internal/, aufgezählt
				filepath.Join(wurzel, paket.Name(), datei.Name()))
			if dateifehler != nil {
				continue
			}

			// **Zeilenanfang und nicht irgendwo.** Diese Datei nennt den
			// Bautag selbst — als Zeichenkette, eingerückt und in
			// Anführungszeichen. Ein `strings.Contains` fände sie und
			// meldete `werkzeuge` als Paket mit Netzlauf: ein Instrument,
			// das sich selbst misst (§10).
			if bautag.Match(inhalt) {
				mitLauf = append(mitLauf, paket.Name())

				break
			}
		}
	}

	return mitLauf
}
