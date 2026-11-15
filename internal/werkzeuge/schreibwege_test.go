// Die Liste der Schreibwege bleibt vollständig
// (#590).
//
// # Was dieser Wächter hält
//
// **Die Vollständigkeit, nicht die Richtigkeit.** Jede Bindung, die
// `frontend/src` ruft, steht in `docs/SCHREIBWEGE.md` — als schreibend oder
// als lesend —, und jede, die dort steht, wird hier gerufen. Eine neue Bindung
// erzwingt damit eine **Entscheidung**, statt still in einer der beiden
// Gruppen zu verschwinden.
//
// # Was er nicht hält, und das steht auch in der Datei
//
// **Ob die Einteilung stimmt.** Dass `SetOeffnen` schreibt und
// `MTGJSONVerzeichnisOeffnen` nicht, ist eine fachliche Aussage über das, was
// hinter der Bindung geschieht — kein Namensmuster. Ein neuer Name mit
// `Oeffnen` darin landet nicht von selbst richtig.
//
// **Ob die genannte Gestalt greift.** Dass neben `Bilder.tsx` „K" steht, prüft
// er nicht; einen entfernten `key` sähe er nicht. Diese Spalte liest ein
// Mensch.
//
// Ein Wächter, der die halbe Zusage hält, macht die andere unsichtbar — also
// wird sie ausgesprochen.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// bezeichner ist ein **ganzer** TypeScript-Bezeichner.
//
// # Warum nicht `[A-Za-z]+`
//
// So stand es bis zum 2026-08-23 — und `api().MSE2Vorschlaege()` wurde damit zu
// `MSE`. Die Aufzählung in `docs/SCHREIBWEGE.md` las derselbe Ausdruck, und
// dort stand ebenfalls `MSE`: **zwei übereinstimmende Verstümmelungen, und der
// Wächter war grün**
// (#931).
//
// Das ist schlimmer als eine halbe Zusage — er bestätigte eine **falsche**.
// Angestrichen vom Nachlauf des Zweitblicks über Phase R.
const bezeichner = `[A-Za-z_$][A-Za-z0-9_$]*`

// bindungsruf findet `api().Name` im Fenstercode.
var bindungsruf = regexp.MustCompile(`api\(\)\.(` + bezeichner + `)`)

// andereAufrufform findet ein `api()`, dem **kein** `.Name` folgt.
//
// **Der Wächter kann nur eine Form lesen, also muss sie die einzige sein** —
// aus dem Zweitblick über Phase L (2026-08-22). `const b = api(); b.Bindung()`
// oder eine Destrukturierung liefen an [bindungsruf] vorbei, und die Zusage
// „jede Bindung steht in der Datei" wäre eine Behauptung über die Syntax
// gewesen, die niemand geprüft hat.
//
// **Die Grenze wird damit zur Regel.** Wer eine andere Form braucht, ändert
// diesen Wächter mit — und entscheidet dabei, wie die Liste vollständig
// bleibt.
var andereAufrufform = regexp.MustCompile(`api\(\)(?:[^.A-Za-z]|$)`)

// aufzaehlung findet die zurückgesetzten Namen in einem Absatz der Datei.
//
// **Derselbe Bezeichner wie beim Aufruf** — zwei Ausdrücke wären zwei
// Antworten auf die Frage, was ein Name ist, und genau daran ist #931
// unsichtbar geblieben.
var aufzaehlung = regexp.MustCompile("`(" + bezeichner + ")`")

func TestJedeBindungStehtInDenSchreibwegen(t *testing.T) {
	t.Parallel()

	gerufen := gerufeneBindungen(t)
	genannt := genannteBindungen(t)

	for name := range gerufen {
		if !genannt[name] {
			t.Errorf("`api().%s` wird gerufen, steht aber in keiner der beiden "+
				"Listen von `docs/SCHREIBWEGE.md` — schreibend oder lesend?", name)
		}
	}

	for name := range genannt {
		if !gerufen[name] {
			t.Errorf("`%s` steht in `docs/SCHREIBWEGE.md`, wird aber von "+
				"`frontend/src` nicht gerufen", name)
		}
	}
}

// gerufeneBindungen sammelt jede `api().Name` aus dem Fenstercode ohne Tests.
func gerufeneBindungen(t *testing.T) map[string]bool {
	t.Helper()

	aus := map[string]bool{}
	wurzel := filepath.Join("..", "..", "frontend", "src")

	fehler := filepath.Walk(wurzel, func(pfad string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		name := info.Name()
		if strings.Contains(name, ".test.") ||
			(!strings.HasSuffix(name, ".ts") && !strings.HasSuffix(name, ".tsx")) {
			return nil
		}

		roh, err := os.ReadFile(pfad) //nolint:gosec // der Pfad kommt aus dem eigenen Baumlauf
		if err != nil {
			return err
		}

		for _, treffer := range bindungsruf.FindAllStringSubmatch(string(roh), -1) {
			aus[treffer[1]] = true
		}

		// **Und keine andere Form**, sonst sähe der Wächter oben sie nicht.
		//
		// **Außer dort, wo `api()` entsteht.** `fassade.ts` erklärt die
		// Funktion; ein Aufruf ist das nicht.
		if !strings.HasSuffix(pfad, "fassade.ts") && andereAufrufform.MatchString(string(roh)) {
			t.Errorf("%s benutzt `api()` in einer Form, die dieser Wächter nicht "+
				"lesen kann — erlaubt ist `api().Name`, damit `docs/SCHREIBWEGE.md` "+
				"vollständig bleibt", pfad)
		}

		return nil
	})
	if fehler != nil {
		t.Fatalf("Fenstercode lesen: %v", fehler)
	}

	if len(aus) == 0 {
		t.Fatal("keine einzige Bindung gefunden — der Wächter sieht den Baum nicht")
	}

	return aus
}

// genannteBindungen liest die beiden Aufzählungen aus `docs/SCHREIBWEGE.md`.
//
// **Über die Überschriften und nicht über die ganze Datei**: die Tabelle
// darüber nennt dieselben Namen, aber auch `useVorgang` und Dateinamen. Zwei
// benannte Absätze sind die Stelle, an der die Liste gepflegt wird.
func genannteBindungen(t *testing.T) map[string]bool {
	t.Helper()

	roh, err := os.ReadFile(filepath.Join("..", "..", "docs", "SCHREIBWEGE.md"))
	if err != nil {
		t.Fatalf("SCHREIBWEGE.md lesen: %v", err)
	}

	aus := map[string]bool{}

	for _, ueberschrift := range []string{"schreibenden Bindungen", "lesenden Bindungen"} {
		absatz := absatzNach(t, string(roh), ueberschrift)

		for _, treffer := range aufzaehlung.FindAllStringSubmatch(absatz, -1) {
			aus[treffer[1]] = true
		}
	}

	return aus
}

// absatzNach gibt den Aufzählungsabsatz unter einer Überschrift.
//
// **Der letzte Absatz vor der nächsten Überschrift**, nicht der erste: unter
// „lesenden Bindungen" steht zuerst ein erklärender Absatz.
func absatzNach(t *testing.T, text, ueberschrift string) string {
	t.Helper()

	stelle := strings.Index(text, ueberschrift)
	if stelle < 0 {
		t.Fatalf("`docs/SCHREIBWEGE.md` führt keine Überschrift %q mehr — "+
			"der Wächter prüft dann nichts", ueberschrift)
	}

	rest := text[stelle:]

	// **Bis zur nächsten Überschrift oder zur nächsten Trennlinie** — je
	// nachdem, was zuerst kommt. Zwischen den beiden Aufzählungen steht keine
	// Trennlinie; wer nur auf sie schnitte, läse für die erste Überschrift die
	// Liste der zweiten und prüfte die erste gar nicht.
	for _, marke := range []string{"\n## ", "\n---"} {
		if ende := strings.Index(rest, marke); ende > 0 {
			rest = rest[:ende]
		}
	}

	absaetze := strings.Split(strings.TrimSpace(rest), "\n\n")

	return absaetze[len(absaetze)-1]
}

// **Die Zahl in der Überschrift stimmt mit der Liste darunter überein** —
// #910.
//
// # Warum das dazugehört
//
// Der Wächter oben hält die **Vollständigkeit** der Liste. Die Überschrift
// nennt daneben eine **Zahl**, und die prüfte niemand: beim Nachzählen am
// 2026-08-23 stand über 31 schreibenden Bindungen „Die 30", über 33 lesenden
// „Die 32". Die erste war schon vor Phase R falsch.
//
// Das ist §10, wörtlich: *ein Wächter, der die halbe Zusage hält, macht die
// andere unsichtbar.*
//
// # Was er nicht prüft
//
// **Ob die Einordnung stimmt.** Er zählt, was unter welcher Überschrift steht;
// dass `AutoErsetzen` wirklich nur liest, sagt kein Test — das ist die
// Entscheidung, die der Wächter oben **erzwingt**, nicht die, die er trifft.
func TestDieUeberschriftenZaehlenIhreListe(t *testing.T) {
	roh, err := os.ReadFile(filepath.Join("..", "..", "docs", "SCHREIBWEGE.md"))
	if err != nil {
		t.Fatalf("SCHREIBWEGE.md lesen: %v", err)
	}

	for _, ueberschrift := range []string{"schreibenden Bindungen", "lesenden Bindungen"} {
		t.Run(ueberschrift, func(t *testing.T) {
			gezaehlt := len(aufzaehlung.FindAllStringSubmatch(
				absatzNach(t, string(roh), ueberschrift), -1))

			genannt := zahlVor(t, string(roh), ueberschrift)

			if genannt != gezaehlt {
				t.Errorf("die Überschrift nennt %d, die Liste führt %d", genannt, gezaehlt)
			}
		})
	}
}

// zahlInUeberschrift findet `## Die <Zahl> … Bindungen`.
var zahlInUeberschrift = regexp.MustCompile(`## Die (\d+) `)

// zahlVor gibt die Zahl aus der Überschrift.
func zahlVor(t *testing.T, text, ueberschrift string) int {
	t.Helper()

	stelle := strings.Index(text, ueberschrift)
	if stelle < 0 {
		t.Fatalf("keine Überschrift %q", ueberschrift)
	}

	// **Rückwärts vom Namen bis zum Zeilenanfang** — die Zahl steht davor.
	beginn := strings.LastIndex(text[:stelle], "## ")
	if beginn < 0 {
		t.Fatalf("die Überschrift %q hat kein `## ` davor", ueberschrift)
	}

	treffer := zahlInUeberschrift.FindStringSubmatch(text[beginn:stelle])
	if treffer == nil {
		t.Fatalf("die Überschrift zu %q nennt keine Zahl", ueberschrift)
	}

	zahl, err := strconv.Atoi(treffer[1])
	if err != nil {
		t.Fatalf("die Zahl in der Überschrift zu %q ist keine: %v", ueberschrift, err)
	}

	return zahl
}

// **Ein Bezeichner mit Ziffer oder Unterstrich wird ganz gelesen** —
// #931.
//
// Die Gegenprobe zum Ausdruck selbst. Ohne sie stünde er da wie der alte:
// plausibel, und an der ersten Ziffer zu Ende.
func TestEinBezeichnerMitZifferWirdGanzGelesen(t *testing.T) {
	faelle := map[string]string{
		"api().MSE2Vorschlaege()": "MSE2Vorschlaege",
		"api().Karte_2Setzen()":   "Karte_2Setzen",
		"api().$Sonderbar()":      "$Sonderbar",
		"api().Sets()":            "Sets",
	}

	for quelle, erwartet := range faelle {
		t.Run(quelle, func(t *testing.T) {
			treffer := bindungsruf.FindStringSubmatch(quelle)
			if treffer == nil {
				t.Fatalf("%q wurde nicht als Aufruf gelesen", quelle)
			}

			if treffer[1] != erwartet {
				t.Errorf("gelesen wurde %q, erwartet %q", treffer[1], erwartet)
			}
		})
	}
}

// **Und die Aufzählung liest denselben Bezeichner.**
func TestDieAufzaehlungLiestDenselbenBezeichner(t *testing.T) {
	treffer := aufzaehlung.FindStringSubmatch("`MSE2Vorschlaege`")
	if treffer == nil {
		t.Fatal("`MSE2Vorschlaege` wurde nicht als Eintrag gelesen")
	}

	if treffer[1] != "MSE2Vorschlaege" {
		t.Errorf("gelesen wurde %q", treffer[1])
	}
}

// **Und der einleitende Satz zählt dieselben Listen.**
//
// # Der Befund
//
// **Zweitblick über die Phasen 4.0 bis 4.13** (2026-09-07): der Satz „Das
// Fenster ruft **219** verschiedene Bindungen; **108** davon ändern etwas"
// stand über Listen mit 244 und 124 Namen. Vorher stand dort „98 … 45".
// Dreimal derselbe Weg: Fließtext, den niemand mitzieht.
//
// **Und er stand mit Datum da.** Der allgemeine Zahlenwächter
// (`flaechenzahlen_test.go`) nimmt jeden Absatz mit Datum von der Prüfung aus
// — als historische Angabe. Dieser Satz beschrieb aber die **Gegenwart**; ein
// Datum wurde damit zur dauerhaften Ausnahme.
//
// **Deshalb trägt er jetzt keins mehr**, und dieser Wächter rechnet ihn nach.
//
// # Was er nicht prüft
//
// **Die Dateien, die derselbe Absatz nennt.** „Das Fenster ruft aus … " ist
// eine Aussage über Herkunft, keine Zahl — sie bleibt eine Frage ans Review.
func TestDerSatzZaehltSeineListen(t *testing.T) {
	roh, err := os.ReadFile(filepath.Join("..", "..", "docs", "SCHREIBWEGE.md"))
	if err != nil {
		t.Fatalf("SCHREIBWEGE.md lesen: %v", err)
	}

	fund := satzMitDenZahlen.FindStringSubmatch(string(roh))
	if fund == nil {
		t.Fatal("der einleitende Satz nennt keine zwei Zahlen mehr — " +
			"er ist umgeschrieben worden, und dieser Wächter greift daneben")
	}

	for stelle, ueberschrift := range map[int]string{
		1: "", 2: "schreibenden Bindungen",
	} {
		genannt, wandelfehler := strconv.Atoi(fund[stelle])
		if wandelfehler != nil {
			t.Fatalf("%q ist keine Zahl: %v", fund[stelle], wandelfehler)
		}

		// **Gezählt werden verschiedene Namen, nicht Vorkommen** — Nachlauf
		// zum Zweitblick über die Phasen 4.0 bis 4.13 (2026-09-07), Befund 5.
		//
		// Der Satz sagt „**244** **verschiedene** Bindungen". Hier stand ein
		// `len(FindAllStringSubmatch(…))`, und das zählt Vorkommen: ein
		// doppelter Eintrag samt erhöhter Zahl wäre durchgegangen. Der
		// Vollständigkeitswächter oben faltet dieselben Namen in eine Map und
		// hätte es ebenso wenig gesehen — **zwei Wächter, dieselbe Lücke.**
		listen := []string{ueberschrift}
		if ueberschrift == "" {
			// **Alle Bindungen sind beide Listen zusammen.**
			listen = []string{"schreibenden Bindungen", "lesenden Bindungen"}
		}

		verschieden := map[string]bool{}

		var doppelt []string

		for _, teil := range listen {
			for _, treffer := range aufzaehlung.FindAllStringSubmatch(
				absatzNach(t, string(roh), teil), -1) {
				if verschieden[treffer[1]] {
					doppelt = append(doppelt, treffer[1])
				}

				verschieden[treffer[1]] = true
			}
		}

		// **Ein doppelter Eintrag ist für sich ein Befund**, gleich ob die
		// Zahl dazu passt: die Liste sagt, welche Bindung wo steht, und
		// zweimal dieselbe sagt es zweimal.
		if len(doppelt) > 0 {
			sort.Strings(doppelt)
			t.Errorf("diese Bindungen stehen mehrfach in der Liste: %s (%s)",
				strings.Join(doppelt, ", "), ueberschrift)
		}

		if genannt != len(verschieden) {
			t.Errorf("der Satz nennt %d, die Listen führen %d verschiedene (%s)",
				genannt, len(verschieden), ueberschrift)
		}
	}
}

// satzMitDenZahlen findet „ruft **244** … **124** davon ändern etwas".
var satzMitDenZahlen = regexp.MustCompile(
	`ruft \*\*(\d+)\*\*[^.]*?\*\*(\d+)\*\* davon ändern etwas`)
