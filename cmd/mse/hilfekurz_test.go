package main

// **Die Hilfe ist kurz, und ihre Beispiele laufen** — LOC-03.AK-04
// (#1689):
// „CLI-Beschriftungen werden kürzer und handlungsorientiert, mit tatsächlich
// geprüften Beispielen. Maschinelle JSON-Ausgabe bleibt stabil, Diagnosen
// gehen nach stderr, fachliche Ausgaben nach stdout."
//
// # Was beim ersten Lauf dastand
//
// **48 von 288 Einträgen** hatten mehr als zwei Zeilen Beschreibung — der
// längste neun —, und **sieben hatten keine**. Einer davon war
// `ki-vorschlag`: zwischen seine Aufrufzeile und seine Beschreibung waren
// `ki-einschaetzung` und `ki-batch` geraten, und wer `mse hilfe` las, hielt
// seine Beschreibung für die von `ki-batch` (gemessen am 2026-09-22).
//
// # Was dieser Wächter nicht hält
//
// **Ob ein Satz trifft.** Er zählt Zeilen und führt Beispiele aus; dass
// hinter einem Befehl das Richtige steht, sieht nur ein Mensch.

import (
	"bytes"
	"encoding/json"
	"maps"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// hilfeeintrag ist ein Befehl der Hilfe mit seinen Beschreibungszeilen.
type hilfeeintrag struct {
	kopf         string
	beschreibung int
}

// kopfUndRest trennt eine Hilfezeile an der ersten Doppelleerstelle.
//
// **An der Doppelleerstelle und nicht an einer Spalte**: eine lange
// Aufrufzeile schiebt ihre Beschreibung nach rechts, und eine Regel über
// Spalten zählte sie als Teil des Aufrufs.
var kopfUndRest = regexp.MustCompile(`^ +(\S(?:.*?\S)?)(?: {2,}(\S.*))?$`)

// befehlsabschnitt ist die Hilfe zwischen »Befehle:« und dem nächsten
// Abschnitt — dort stehen die Einträge, und nur dort.
func befehlsabschnitt(t *testing.T) []string {
	t.Helper()

	_, danach, gefunden := strings.Cut(hilfe, "\nBefehle:\n")
	if !gefunden {
		t.Fatal("die Hilfe hat keinen Abschnitt »Befehle:«")
	}

	abschnitt, _, _ := strings.Cut(danach, "\n\n")

	return strings.Split(abschnitt, "\n")
}

// hilfeeintraege zerlegt den Befehlsabschnitt in Einträge.
//
// **Eine Fortsetzung zählt zur Beschreibung**, wenn sie ab der
// Beschreibungsspalte steht oder hinter einer Doppelleerstelle Text trägt;
// sonst setzt sie die Aufrufzeile fort — `[erwerbsdatum] …` beim Exemplar.
func hilfeeintraege(t *testing.T) []hilfeeintrag {
	t.Helper()

	var eintraege []hilfeeintrag

	for _, zeile := range befehlsabschnitt(t) {
		teile := kopfUndRest.FindStringSubmatch(zeile)
		if teile == nil {
			continue
		}

		einzug := len(zeile) - len(strings.TrimLeft(zeile, " "))

		switch {
		case einzug == 2:
			eintraege = append(eintraege, hilfeeintrag{kopf: teile[1]})

			if teile[2] != "" {
				eintraege[len(eintraege)-1].beschreibung++
			}

		case len(eintraege) == 0:
			t.Fatalf("eine Fortsetzung vor dem ersten Befehl: %q", zeile)

		case einzug >= 34 || teile[2] != "":
			eintraege[len(eintraege)-1].beschreibung++
		}
	}

	return eintraege
}

// **Jeder Befehl sagt in einer oder zwei Zeilen, was er tut.**
func TestJederBefehlHatEineKurzeBeschreibung(t *testing.T) {
	t.Parallel()

	eintraege := hilfeeintraege(t)

	// **Die Gegenprobe gegen einen leeren Abschnitt** — ohne sie stimmte
	// der Lauf jeder Hilfe zu, die keinen Befehl mehr nennt.
	if len(eintraege) < 200 {
		t.Fatalf("nur %d Einträge gelesen", len(eintraege))
	}

	for _, eintrag := range eintraege {
		switch {
		case eintrag.beschreibung == 0:
			t.Errorf("%q sagt nicht, was er tut", eintrag.kopf)

		case eintrag.beschreibung > 2:
			t.Errorf("%q beschreibt sich in %d Zeilen — höchstens zwei",
				eintrag.kopf, eintrag.beschreibung)
		}
	}
}

// beispiele liest die Aufrufe unter »Beispiele:«.
func beispiele(t *testing.T) [][]string {
	t.Helper()

	_, danach, gefunden := strings.Cut(hilfe, "\nBeispiele:\n")
	if !gefunden {
		t.Fatal("die Hilfe hat keinen Abschnitt »Beispiele:«")
	}

	abschnitt, _, _ := strings.Cut(danach, "\n\n")

	var gelesen [][]string

	for _, zeile := range strings.Split(abschnitt, "\n") {
		// **Was darunter steht, sagt, wozu das Beispiel gut ist** —
		// dieselbe Spalte wie bei den Befehlen.
		if strings.HasPrefix(zeile, strings.Repeat(" ", 34)) {
			continue
		}

		teile := kopfUndRest.FindStringSubmatch(zeile)
		if len(teile) < 2 || !strings.HasPrefix(teile[1], "mse ") {
			t.Fatalf("keine Beispielzeile: %q", zeile)
		}

		gelesen = append(gelesen, strings.Fields(teile[1])[1:])
	}

	return gelesen
}

// gestalt ist, womit ein Skript rechnet: die obersten Schlüssel jedes
// JSON-Beispiels (gemessen am 2026-09-22).
//
// **Ein neues JSON-Beispiel braucht hier seine Zeile** — sonst stünde es
// unter keiner Zusage, und „bleibt stabil" hieße nur „ist gültig".
var gestalt = map[string][]string{
	"einstellungen --json": {"einstellungen"},
	"decks --json":         {"decks", "fehler"},
	"sammlung --json":      {"exemplare"},
}

// **Jedes Beispiel läuft — nacheinander, gegen denselben Bestand.**
//
// **Ergebnis nach stdout, nichts nach stderr**: ein geglückter Aufruf hat
// keine Diagnose, und wer die Ausgabe in eine Datei leitet, soll dort nur das
// Ergebnis finden. **Und `--json` ist JSON** — sonst bräche das erste
// Skript, das die Ausgabe liest.
func TestDieBeispieleDerHilfeLaufen(t *testing.T) {
	t.Parallel()

	ort := frischerBestand(t)
	gelesen := beispiele(t)

	if len(gelesen) < 3 {
		t.Fatalf("nur %d Beispiele", len(gelesen))
	}

	for _, aufruf := range gelesen {
		var aus, fehler bytes.Buffer

		code := lauf(probekontext(t), aufruf, nil, &aus, &fehler, ort)

		if code != CodeGut {
			t.Errorf("mse %s endete mit %d: %s",
				strings.Join(aufruf, " "), code, fehler.String())

			continue
		}

		if fehler.Len() > 0 {
			t.Errorf("mse %s schrieb nach stderr: %q",
				strings.Join(aufruf, " "), fehler.String())
		}

		if aus.Len() == 0 {
			t.Errorf("mse %s gab nichts aus", strings.Join(aufruf, " "))
		}

		if slices.Contains(aufruf, "--json") {
			pruefeGestalt(t, strings.Join(aufruf, " "), aus.Bytes())
		}
	}
}

// pruefeGestalt hält die obersten Schlüssel einer JSON-Ausgabe fest.
func pruefeGestalt(t *testing.T, aufruf string, roh []byte) {
	t.Helper()

	var gelesen map[string]json.RawMessage
	if fehler := json.Unmarshal(roh, &gelesen); fehler != nil {
		t.Errorf("mse %s gab kein JSON-Objekt: %v", aufruf, fehler)

		return
	}

	erwartet, bekannt := gestalt[aufruf]
	if !bekannt {
		t.Errorf("mse %s hat keine Gestalt in `gestalt` — eintragen", aufruf)

		return
	}

	schluessel := slices.Sorted(maps.Keys(gelesen))
	if !slices.Equal(schluessel, erwartet) {
		t.Errorf("mse %s trägt %v, erwartet %v", aufruf, schluessel, erwartet)
	}
}

// **Eine Absage geht nach stderr, und stdout bleibt leer** — wer die
// Ausgabe weiterreicht, reicht keine Fehlermeldung als Ergebnis weiter.
func TestEineAbsageGehtNachStderr(t *testing.T) {
	t.Parallel()

	code, aus, fehler := fahre(t,
		"einstellungen", "setzen", "darstellung", "de_CH")

	if code != CodeFachlich {
		t.Errorf("Rückgabewert %d, erwartet %d", code, CodeFachlich)
	}

	if aus != "" {
		t.Errorf("auf stdout stand %q", aus)
	}

	// **Und die Absage sagt, was zu tun ist** (AK-03): »de_CH« schreibt
	// sich »de-CH«.
	if !strings.Contains(fehler, "de-CH") {
		t.Errorf("die Absage nennt die richtige Schreibweise nicht: %q", fehler)
	}
}
