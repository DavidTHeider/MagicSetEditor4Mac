// Der Wächter über `docs/PARITAET.md`.
//
// Die Matrix sagt, was MSE2 kann und was davon hier steht — und vor allem,
// was **unvermessen** ist
// (#1058).
//
// # Was er hält
//
//   - Jede Zeile der Bereichstabellen trägt genau **eine** der sechs Marken.
//   - Keine Einheit steht zweimal in derselben Tabelle.
//   - Die Zählung am Ende stimmt mit dem überein, was in den Tabellen steht.
//   - Jede `new`-Zeile sagt, **wo gesucht wurde** (§15).
//   - Ein **Zahlwort unmittelbar vor einer Marke** im Fließtext stimmt mit
//     der Zählung überein — „die vier `uncertain`".
//
// # Was er nicht hält
//
// **Ob die Marke stimmt.** Dass etwas `ported` ist und nicht `uncertain`, ist
// eine Aussage über Verhalten; das prüft ein Test, kein Wächter über eine
// Tabelle.
//
// **Ob die Matrix vollständig ist.** Dafür müsste er den MSE2-Quelltext lesen,
// und der liegt nicht im Repositorium (§11). Das tut `paritaetquelle_test.go`
// hinter dem Tag `fremdpakete`.
//
// **Ob die C++-Quelle existiert.** Dieselbe Grenze, derselbe Nachbar.
package werkzeuge

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// paritaetsmarken sind die sechs, die die Matrix kennt.
var paritaetsmarken = map[string]bool{
	"ported": true, "adapted": true, "new": true,
	"uncertain": true, "deferred": true, "unsupported": true,
}

// paritaetszeile ist eine Zeile einer Bereichstabelle.
type paritaetszeile struct {
	bereich string
	spalten []string
	marke   string
	nummer  int
}

// paritaetLesen gibt die Zeilen der Bereichstabellen und die Zählung.
func paritaetLesen(t *testing.T) (zeilen []paritaetszeile, zaehlung map[string]int) {
	t.Helper()

	inhalt := paritaetsdatei(t)

	anfang := strings.Index(inhalt, "## Bereich 1")
	ende := strings.Index(inhalt, "## Die Zählung")

	if anfang < 0 || ende < 0 {
		t.Fatal("PARITAET.md hat keine Bereiche oder keine Zählung")
	}

	zaehlung = map[string]int{}

	bereich := ""

	for nummer, zeile := range strings.Split(inhalt[anfang:ende], "\n") {
		if ueberschrift, ist := strings.CutPrefix(zeile, "## "); ist {
			bereich = ueberschrift

			continue
		}

		spalten, gut := paritaetsspalten(zeile)
		if !gut {
			continue
		}

		marke := markeAus(spalten)
		if marke == "" {
			t.Errorf("Zeile %d trägt keine der sechs Marken: %s", nummer, kurz(zeile))

			continue
		}

		zeilen = append(zeilen, paritaetszeile{
			bereich: bereich, spalten: spalten, marke: marke, nummer: nummer,
		})
		zaehlung[marke]++
	}

	return zeilen, zaehlung
}

// paritaetsdatei gibt den Inhalt der Matrix.
func paritaetsdatei(t *testing.T) string {
	t.Helper()

	roh, err := os.ReadFile(filepath.Join("..", "..", "docs", "PARITAET.md"))
	if err != nil {
		t.Fatalf("PARITAET.md lesen: %v", err)
	}

	return string(roh)
}

// paritaetsabschnitt gibt alles ab einer Überschrift — und bricht ab, wenn es
// sie nicht gibt.
//
// **Der Abbruch ist der Punkt.** Ein fehlender Abschnitt hieße, dass jemand
// die Datei umgebaut hat; stillschweigend von Stelle null zu lesen prüfte dann
// die ganze Datei und meldete Unsinn.
func paritaetsabschnitt(t *testing.T, ueberschrift string) string {
	t.Helper()

	inhalt := paritaetsdatei(t)

	stelle := strings.Index(inhalt, ueberschrift)
	if stelle < 0 {
		t.Fatalf("PARITAET.md hat keinen Abschnitt %q", ueberschrift)
	}

	return inhalt[stelle:]
}

// paritaetsspalten gibt die Spalten einer Datenzeile — oder nichts.
func paritaetsspalten(zeile string) ([]string, bool) {
	if !strings.HasPrefix(zeile, "|") || strings.HasPrefix(zeile, "| ---") {
		return nil, false
	}

	spalten := make([]string, 0, 6)
	for _, teil := range strings.Split(strings.Trim(zeile, "|"), "|") {
		spalten = append(spalten, strings.TrimSpace(teil))
	}

	// **Die Kopfzeile ist keine Datenzeile.** Sie erkennt sich an ihrer ersten
	// Spalte; ein Bereich mit einer anderen Überschrift bräuchte hier einen
	// Eintrag, und das ist Absicht — sonst rutschte eine Kopfzeile als Zeile
	// ohne Marke durch und der Lauf wäre rot, ohne dass jemand wüsste, warum.
	switch spalten[0] {
	case "Einheit", "Feldart":
		return nil, false
	}

	return spalten, true
}

// markeAus gibt die eine Marke der Zeile — oder nichts, wenn es keine oder
// mehrere sind.
func markeAus(spalten []string) string {
	gefunden := ""

	for _, spalte := range spalten {
		nackt := strings.Trim(spalte, "*`")
		if !paritaetsmarken[nackt] {
			continue
		}

		if gefunden != "" {
			return ""
		}

		gefunden = nackt
	}

	return gefunden
}

func kurz(zeile string) string {
	if len(zeile) > 70 {
		return zeile[:70] + "…"
	}

	return zeile
}

// **Jede Zeile trägt genau eine Marke, und keine Einheit steht zweimal.**
func TestDieParitaetsmatrixIstWohlgeformt(t *testing.T) {
	t.Parallel()

	zeilen, _ := paritaetLesen(t)

	if len(zeilen) < 50 {
		t.Fatalf("%d Zeilen — die Matrix ist zu kurz, um die acht Bereiche zu decken",
			len(zeilen))
	}

	// **Doppelte je Bereich, nicht über die ganze Datei.** `symbol` steht in
	// Bereich 2 (die Feldart) und noch einmal in Bereich 3 (ihre Bedienstelle),
	// „Auszählung" in Bereich 1 (der Bereich im Fenster) und in Bereich 5 (die
	// Klasse dahinter) — jedes Mal wird eine andere Frage beantwortet.
	//
	// Gesehen, nicht überlegt: der erste Anlauf prüfte über die ganze Datei und
	// fand fünf „Doppelte", von denen keine eine war.
	gesehen := map[string]int{}

	for _, zeile := range zeilen {
		schluessel := zeile.bereich + "|" + zeile.spalten[0]
		if vorher, doppelt := gesehen[schluessel]; doppelt {
			t.Errorf("die Einheit %q steht in %q zweimal — Zeile %d und %d",
				zeile.spalten[0], zeile.bereich, vorher, zeile.nummer)
		}

		gesehen[schluessel] = zeile.nummer
	}
}

// **Die Zählung am Ende ist gezählt, nicht geschätzt.**
//
// # Warum es diesen Lauf gibt
//
// Der erste Anlauf schrieb sechs Zahlen von Hand, und **fünf waren falsch**.
// Eine Zahl in Fließtext, die niemand nachzählt, veraltet nicht erst — sie ist
// von Anfang an eine Behauptung.
func TestDieParitaetszaehlungStimmt(t *testing.T) {
	t.Parallel()

	zeilen, zaehlung := paritaetLesen(t)

	abschnitt := paritaetsabschnitt(t, "## Die Zählung")

	genannt := map[string]int{}
	for _, zeile := range strings.Split(abschnitt, "\n") {
		spalten, gut := paritaetsspalten(zeile)
		if !gut || len(spalten) != 2 {
			continue
		}

		genannt[strings.Trim(spalten[0], "`*")] = zahlAus(spalten[1])
	}

	if genannt["Zeilen insgesamt"] != len(zeilen) {
		t.Errorf("die Zählung nennt %d Zeilen, die Tabellen haben %d",
			genannt["Zeilen insgesamt"], len(zeilen))
	}

	for marke := range paritaetsmarken {
		if genannt[marke] != zaehlung[marke] {
			t.Errorf("die Zählung nennt %d × %q, die Tabellen haben %d",
				genannt[marke], marke, zaehlung[marke])
		}
	}
}

// zahlAus liest eine Zahl aus einer Zelle wie `**56**`.
func zahlAus(zelle string) int {
	zahl := 0

	for _, zeichen := range zelle {
		if zeichen < '0' || zeichen > '9' {
			continue
		}

		zahl = zahl*10 + int(zeichen-'0')
	}

	return zahl
}

// **Jede `new`-Zeile sagt, wo gesucht wurde** (§15).
//
// Sie darf keine unterlassene Suche verdecken — und das ist der Satz, für den
// es diese Matrix überhaupt gibt.
func TestJedesNeuSagtWoGesuchtWurde(t *testing.T) {
	t.Parallel()

	zeilen, _ := paritaetLesen(t)

	for _, zeile := range zeilen {
		if zeile.marke != "new" {
			continue
		}

		if !strings.Contains(strings.Join(zeile.spalten, " "), "gesucht") {
			t.Errorf("die Zeile %q trägt »new« und sagt nicht, wo gesucht wurde",
				zeile.spalten[0])
		}
	}
}

// zahlwoerter ordnet ausgeschriebene Zahlen ihrem Wert zu.
//
// **Bis zwölf**, weil Fließtext darüber ohnehin Ziffern schreibt.
var zahlwoerter = map[string]int{
	// **„null" gehört dazu, seit es eine Marke ohne Zeilen gibt.**
	//
	// Werkstück D der Phase 4.9 hat `deferred` geleert
	// (#1324),
	// und ein Text, der das sagen will, braucht das Wort dafür — sonst
	// stünde dort eine Ziffer, wo überall sonst ein Wort steht.
	"null": 0,
	"eine": 1, "ein": 1, "zwei": 2, "drei": 3, "vier": 4, "fünf": 5, "sechs": 6,
	"sieben": 7, "acht": 8, "neun": 9, "zehn": 10, "elf": 11, "zwölf": 12,
}

// TestDieZahlenImFliesstextStimmenAuch schließt die Lücke, die Befund 9 des
// Zweitblicks über Phase W angestrichen hat.
//
// **Der Zähler daneben prüft nur die Tabelle.** Im Absatz darunter stand „die
// **drei** `uncertain`", während die Tabelle vier führte — alle
// Paritätswächter grün, und die Zahl unmittelbar daneben falsch.
//
// **Was dieser Lauf nicht prüft**: jede andere Zahl in dieser Datei. Er
// greift genau dort, wo ein Zahlwort **unmittelbar** vor einer Marke steht —
// „die vier `uncertain`". Eine Zahl in einem anderen Satzbau bleibt
// unbewacht, und das steht hier, damit niemand ihn für mehr hält (§10).
func TestDieZahlenImFliesstextStimmenAuch(t *testing.T) {
	t.Parallel()

	_, zaehlung := paritaetLesen(t)

	roh, err := os.ReadFile(filepath.Join("..", "..", "docs", "PARITAET.md"))
	if err != nil {
		t.Fatalf("die Paritätsmatrix: %v", err)
	}

	geprueft := 0

	for marke, will := range zaehlung {
		muster := regexp.MustCompile(`(?i)\b(\p{L}+)\s+` + "`" + regexp.QuoteMeta(marke) + "`")

		for _, treffer := range muster.FindAllStringSubmatch(string(roh), -1) {
			hat, istZahl := zahlwoerter[strings.ToLower(treffer[1])]
			if !istZahl {
				continue
			}

			geprueft++

			if hat != will {
				t.Errorf("der Fließtext sagt %q vor %q, die Tabellen zählen %d",
					treffer[1], marke, will)
			}
		}
	}

	// **Ohne diese Zeile wäre der Lauf grün, wenn das Muster nie greift** —
	// die Gegenprobe zu sich selbst.
	if geprueft == 0 {
		t.Error("kein einziges Zahlwort vor einer Marke gefunden — das Muster greift nicht mehr")
	}
}
