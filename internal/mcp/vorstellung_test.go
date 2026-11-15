// Der Wächter über die Vorstellung — die **erste** Auskunft, die ein Gegenüber
// bekommt.
//
// # Zwei Befunde, eine Antwort
//
// `initialize` gab bis zum 2026-08-23 zwei Angaben, die nicht stimmten
// (#896):
//
//	"version": "0.0.0",
//
// Die Anwendung ist 0.1.0. `versionen_pruefen` hält vier Orte zusammen —
// `wails.json`, `frontend/package.json`, die Überschrift in `CHANGELOG.md` und
// den Tag. Diesen fünften kannte es nicht.
//
//	"Es gibt kein Werkzeug, das ein ganzes Set ändert oder eine Karte löscht."
//
// `schreiben.karte.loeschen` gibt es seit
// #621. Der
// Satz war richtig, als er geschrieben wurde — und er ist eine **zweite
// Aufzählung** neben der Werkzeugliste, also die, die niemand nachführt (§9).
//
// # Was hier geprüft wird
//
// Nicht der Wortlaut, sondern die **Ableitung**: die Anweisung entsteht aus
// der Liste, und der Test führt sie gegen die Liste. Ein Werkzeug, das
// dazukommt oder wegfällt, verschiebt beide zugleich.
//
// # Was er nicht prüft
//
//   - **Ob die Fassung die richtige ist.** Er prüft, dass die Vorstellung die
//     nennt, die ihr gegeben wurde. Dass die zu `wails.json` passt, hält
//     `internal/werkzeuge/fassung_test.go` — dort, wo die anderen vier Orte
//     schon zusammengehalten werden.
//   - **Ob ein Gegenüber die Anweisung liest.** Sie steht da; was jemand
//     daraus macht, sieht kein Test.
package mcp_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// vorstellung holt die Antwort auf `initialize`.
func vorstellung(t *testing.T, fassung string) map[string]any {
	t.Helper()

	bestand := probenbestand()
	server := mcp.NeuerServer(mcp.Werkzeuge(mcp.Naehte{
		Leser:       bestand,
		Schreiber:   bestand,
		Ausfuhr:     ruhendeAusfuhr{},
		Abgleich:    nil,
		Bilder:      nil,
		Auskunft:    stilleAuskunft{},
		Sammeldaten: stilleSammeldaten{},
		Anlage:      nil,
		Saetze:      nil,
		Konsole:     nil,
		Wortlisten:  nil,
		Woerter:     nil,
		Pack:        nil,
		Anleger:     nil,
		Angaben:     neueAngabenprobe(),
		Stile:       neueStilprobe(),
		Bestand:     neueBestandsprobeWeit(),
	})).MitFassung(fassung)

	var aus strings.Builder

	ein := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n")
	if err := server.Bediene(t.Context(), ein, &aus); err != nil {
		t.Fatalf("bedienen: %v", err)
	}

	zeilen := strings.Split(strings.TrimSpace(aus.String()), "\n")
	if len(zeilen) != 1 {
		t.Fatalf("%d Antworten", len(zeilen))
	}

	return teilAntwort(t, zeilen[0])
}

// **Die Vorstellung nennt die Fassung, die ihr gegeben wurde.**
func TestDieVorstellungNenntDieGegebeneFassung(t *testing.T) {
	angaben := abbildung(t, vorstellung(t, "1.2.3")["serverInfo"])

	if gemeldet := zeichenkette(t, angaben["version"]); gemeldet != "1.2.3" {
		t.Errorf("die Vorstellung meldet %q statt der gegebenen Fassung 1.2.3", gemeldet)
	}
}

// **Und sie meldet nie `0.0.0`.**
//
// Ein eigener Fall, weil er der Anlass ist: eine Vorgabe, die wie eine Fassung
// aussieht, ist schlimmer als gar keine — sie sieht nach einer Auskunft aus.
func TestDieVorstellungMeldetNiemalsNullFassung(t *testing.T) {
	angaben := abbildung(t, vorstellung(t, "")["serverInfo"])

	if gemeldet := zeichenkette(t, angaben["version"]); gemeldet == "0.0.0" {
		t.Error("die Vorstellung meldet 0.0.0 — das ist keine Fassung, sondern eine Lücke mit Ziffern")
	}
}

// **Die Anweisung sagt, was die Werkzeugliste hergibt — und nichts anderes.**
//
// Geprüft wird beides: dass jedes vorhandene schreibende Werkzeug in ihr
// vorkommt, und dass sie nichts verneint, was es gibt.
func TestDieAnweisungFolgtDerWerkzeugliste(t *testing.T) {
	anweisung := zeichenkette(t, vorstellung(t, "1.2.3")["instructions"])

	vorhanden := map[string]bool{}
	for _, werkzeug := range werkzeugliste(t) {
		vorhanden[zeichenkette(t, werkzeug["name"])] = true
	}

	// **Was es gibt, wird nicht verneint.**
	verneinungen := map[string]string{
		"schreiben.karte.loeschen": "eine Karte löscht",
	}

	for werkzeug, verneinung := range verneinungen {
		if vorhanden[werkzeug] && strings.Contains(anweisung, verneinung) {
			t.Errorf("die Anweisung sagt %q — %s gibt es aber", verneinung, werkzeug)
		}
	}

	// **Und was sie aufzählt, gibt es.**
	for _, name := range namenIn(anweisung) {
		if !vorhanden[name] {
			t.Errorf("die Anweisung nennt %s — das Werkzeug gibt es nicht", name)
		}
	}
}

// namenIn liest die Werkzeugnamen aus einem Text.
//
// **Über das Präfix und nicht über eine Liste**: eine Liste hier wäre die
// dritte Aufzählung neben der Werkzeugliste und der Anweisung.
//
// **`lesen.…` ist kein Name, sondern ein Muster.** Die Anweisung erklärt damit
// die Benennung, und ein Wächter, der es für ein Werkzeug hielte, verböte den
// erklärenden Satz.
func namenIn(text string) []string {
	var namen []string

	for _, wort := range strings.Fields(text) {
		wort = strings.Trim(wort, ".,;:»«\"")

		if strings.HasSuffix(wort, "…") {
			continue
		}

		if strings.HasPrefix(wort, "lesen.") || strings.HasPrefix(wort, "schreiben.") {
			namen = append(namen, wort)
		}
	}

	return namen
}

// **Und ohne schreibendes Werkzeug sagt sie das.**
//
// Der Fall kommt im Betrieb nicht vor — `mcp.Werkzeuge` liefert immer beide
// Reihen. Er kommt an der **Schnittstelle** vor: `NeuerServer` nimmt jede
// Liste, und ein Gegenüber, dem jemand einen nur lesenden Server hinstellt,
// soll das lesen können statt „Diese ändern etwas:" gefolgt von nichts.
func TestOhneSchreibendesWerkzeugSagtDieAnweisungEsAuch(t *testing.T) {
	nurLesend := []mcp.Werkzeug{{Name: "lesen.probe", Art: mcp.Lesend, Idempotent: true}}

	var aus strings.Builder

	ein := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n")
	if err := mcp.NeuerServer(nurLesend).MitFassung("1.2.3").Bediene(t.Context(), ein, &aus); err != nil {
		t.Fatalf("bedienen: %v", err)
	}

	anweisung := zeichenkette(t,
		teilAntwort(t, strings.TrimSpace(aus.String()))["instructions"])

	if !strings.Contains(anweisung, "kein schreibendes Werkzeug") {
		t.Errorf("die Anweisung sagt nicht, dass nichts schreibt: %s", anweisung)
	}
}
