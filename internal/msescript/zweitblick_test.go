package msescript_test

// Die Befunde des Zweitblicks durch Codex über Phase 6, als Prüfungen.
//
// **Eigene Datei**, weil `paketformen_test.go` über die Zeilengrenze aus
// AGENTS.md §2 gewachsen war. Der Schnitt folgt der Frage: dort steht, welche
// Formen echte Pakete schreiben, hier, woran sie trotzdem scheiterten.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// **Eine unabgeschlossene Zeichenkette meldet ihren Anfang.**
//
// # Was hier zuerst stand, und warum es fiel
//
// Der Zweitblick über Phase 6 hatte gefunden, dass `"a\` samt Umbruch den
// Zeilenwechsel verschluckte, und daraus die Regel gemacht: **eine
// Zeichenkette endet an der Zeile.** Die Begründung lautete, `docs/FORMATE.md`
// führe keine mehrzeilige, und ein Zerteiler, der eine erfände, läse eine
// Datei anders als MSE.
//
// **Ausgezählt ist das widerlegt.** Von 429 Skriptdateien in `M15-Magic-Pack`
// führen 24 eine Zeichenkette über die Zeilengrenze —
// `magic.mse-game/statistics_script` und die `keywords_*`-Dateien mit über
// 2.000 offenen Zeilen. Die Regel war aus dem Fehlen eines Belegs gebaut, nicht
// aus einem Beleg (§10).
//
// **Was von dem Befund bleibt, ist die Zeilenzählung**: der Umbruch in einer
// Zeichenkette wird mitgezählt, sonst zeigt jeder Fehler danach auf eine
// falsche Zeile. Das hält
// [TestNachEinerMehrzeiligenZeichenketteStimmtDieZeile].
//
// **Und was übrig bleibt, ist dieser Fall**: eine Zeichenkette, die bis zum
// Dateiende offen bleibt, meldet die Stelle, an der sie **anfing** — nicht das
// Dateiende. Wer den Anfang sucht, sucht sonst rückwärts durch die ganze Datei.
func TestEineOffeneZeichenketteMeldetIhrenAnfang(t *testing.T) {
	_, err := msescript.Zerteile("a := 1\nb := \"x\ny := 2")

	var befund *msescript.Fehler
	if !errors.As(err, &befund) {
		t.Fatalf("Fehler %v, erwartet ein *msescript.Fehler", err)
	}

	if befund.Zeile != 2 || befund.Spalte != 6 {
		t.Errorf("Zeile %d, Spalte %d — erwartet war 2:6, der Anfang der Zeichenkette",
			befund.Zeile, befund.Spalte)
	}
}

func TestJedeZaehlschleifeOhneFortschrittWirdAbgewiesen(t *testing.T) {
	for name, quelle := range map[string]string{
		"Anfang gleich Ende":    "for x from 9007199254740992 to 9007199254740992 do 1",
		"Anfang unter Ende":     "for x from 9007199254740992 to 9007199254740994 do 1",
		"unterwegs am Anschlag": "for x from 9007199254740991 to 9007199254740994 do 1",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := msescript.AuswertenQuelle(
				t.Context(), quelle, msescript.Bibliothek().Darueber()); err == nil {
				t.Error("kein Fehler — die Schleife käme nicht voran")
			}
		})
	}

	// Und eine gewöhnliche läuft weiter.
	wert, err := msescript.AuswertenQuelle(t.Context(),
		"to_text(for x from 1 to 3 do to_text(x))", msescript.Bibliothek().Darueber())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.AlsText() != "123" {
		t.Errorf("%q", wert.AlsText())
	}
}

// **Eine Zeichenkette darf über mehrere Zeilen gehen.**
//
// # Was hier stand und warum es fiel
//
// Bis zum Zweitblick über Phase 6 endete eine Zeichenkette an ihrer Zeile, und
// der Grund dafür stand daneben: `docs/FORMATE.md` führte keine mehrzeilige,
// und einen Zerteiler, der eine erfindet, wollte niemand.
//
// **Ausgezählt widerlegt.** Von 429 Skriptdateien in `M15-Magic-Pack` führen
// **24** eine Zeichenkette über die Zeilengrenze — darunter
// `magic.mse-game/statistics_script` und die `keywords_*`-Dateien mit über
// 2.000 offenen Zeilen. Ohne diese Form zerteilt sich das Skript des
// Spielpakets gar nicht
// (#195).
func TestEineZeichenketteDarfUeberZeilenGehen(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		"\"erste\nzweite\"", msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.Text != "erste\nzweite" {
		t.Errorf("%q", wert.Text)
	}
}

// **Und die Zeilenzählung läuft dabei mit.**
//
// Sonst zeigt jeder Fehler danach auf eine falsche Zeile — genau der Schaden,
// den der Zweitblick an der Fluchtfolge vor dem Zeilenende gefunden hat.
func TestNachEinerMehrzeiligenZeichenketteStimmtDieZeile(t *testing.T) {
	_, err := msescript.AuswertenQuelle(t.Context(),
		"a := \"eins\nzwei\ndrei\";\nb := ?", msescript.Bibliothek().Darueber())
	if err == nil {
		t.Fatal("kein Fehler")
	}

	var fehler *msescript.Fehler
	if !errors.As(err, &fehler) {
		t.Fatalf("Fehler %v ist kein *msescript.Fehler", err)
	}

	if fehler.Zeile != 4 {
		t.Errorf("Zeile %d, erwartet war 4 — %v", fehler.Zeile, err)
	}
}

// **Ein Schlüssel einer Abbildung darf eine Zeichenkette sein.**
//
// # Woran das aufgefallen ist
//
// `magic.mse-game/language` schreibt seine Sprachtabelle so:
//
//	languages := [
//		"Deutsch": [ code: "de", fallback: "English" ]
//	]
//
// Bis dahin galt: „Die Schlüssel sind **Namen, keine Ausdrücke** […] Ein
// Ausdruck als Schlüssel kommt nirgends vor." Das war ein Schluss aus dem
// Fehlen eines Belegs — und die Sprachtabelle des Spielpakets ist der Beleg
// dagegen (#195).
//
// **Weiterhin kein Ausdruck.** Ein Name oder eine Zeichenkette, sonst nichts:
// beides steht **fest** in der Datei, und beides lässt sich lesen, ohne etwas
// auszurechnen.
func TestEinAbbildungsschluesselDarfEineZeichenketteSein(t *testing.T) {
	wert, err := msescript.AuswertenQuelle(t.Context(),
		`["Deutsch": "de", englisch: "en"].Deutsch`, msescript.Bibliothek())
	if err != nil {
		t.Fatalf("auswerten: %v", err)
	}

	if wert.Text != "de" {
		t.Errorf("%q", wert.Text)
	}
}

// **Ein Rückstrich als letztes Zeichen der Datei bleibt stehen.**
//
// Er flieht nichts, weil nichts mehr folgt — und die Zeichenkette ist damit
// offen. Ohne diesen Fall bliebe der Zweig ungeprüft, in dem der Rückstrich
// **nicht** flieht (§3).
func TestEinRueckstrichAmDateiendeFliehtNichts(t *testing.T) {
	if _, err := msescript.Zerteile(`a := "x\`); err == nil {
		t.Error("eine offene Zeichenkette wurde angenommen")
	}
}

// **Ein geflohener Umbruch zählt als Zeile mit.**
//
// `"a\` samt Umbruch: der Rückstrich flieht ihn, er wird Inhalt — und die
// Zeilenzählung muss ihn trotzdem sehen. Ohne diesen Fall bliebe der Zweig
// ungeprüft, und ein Fehler danach zeigte auf die falsche Zeile.
func TestEinGeflohenerUmbruchZaehltAlsZeile(t *testing.T) {
	_, err := msescript.AuswertenQuelle(t.Context(),
		"a := \"x\\\ny\"\nb := ?", msescript.Bibliothek().Darueber())

	var befund *msescript.Fehler
	if !errors.As(err, &befund) {
		t.Fatalf("Fehler %v, erwartet ein *msescript.Fehler", err)
	}

	if befund.Zeile != 3 {
		t.Errorf("Zeile %d, erwartet war 3 — %v", befund.Zeile, err)
	}
}
