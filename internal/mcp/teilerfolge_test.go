// **Teilerfolge gehen nicht verloren** — Befund 1 des Zweitblicks über Phase
// 4.4 und seine Fortsetzung im Nachlauf
// (#1291).
//
// # Die Klasse
//
// Mehrere Fachwege geben **absichtlich** ein Ergebnis **und** einen Fehler:
// eine AllSets-Datei, die mitten im Lauf abbricht; ein Archiv, das sich nach
// dem Einfahren nicht schließen lässt; eine Übernahme, deren Katalogauffrischen
// scheitert. In allen drei Fällen **steht etwas im Bestand oder auf der
// Platte**, und wer nur den Fehler weitergibt, lässt es zurück, ohne dass
// jemand davon weiß (§7 Regel 3).
//
// **Eine eigene Datei**, seit `restparitaet_test.go` an der Zeilengrenze aus
// §2 stand.
package mcp_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// **Ein Teilerfolg beim Einlesen geht nicht verloren** — Befund 1 des
// Zweitblicks über Phase 4.4
// (#1291).
//
// Eine AllSets-Datei führt Hunderte Sets, jedes in einer eigenen Transaktion.
// Bricht sie mitten im Lauf ab, **stehen die vorherigen im Bestand** — und wer
// nur den Fehler bekommt, lässt sie zurück, ohne von ihnen zu wissen.
func TestEinTeilerfolgBeimEinlesenGehtNichtVerloren(t *testing.T) {
	probe := &pflegeprobe{}

	_, err := mitPflege(t, probe, "schreiben.mtgjson.lesen").
		Ruf(t.Context(), json.RawMessage(`{"pfad":"/tmp/kaputt.json","spiel":"magic"}`))
	if err == nil {
		t.Fatal("kein Fehler")
	}

	// **Die Kennung steht in der Meldung.** Sie ist die einzige Spur, die ein
	// Gegenüber nach einem Abbruch hat.
	if !strings.Contains(err.Error(), "7") {
		t.Errorf("die Meldung nennt die angelegten Sets nicht: %v", err)
	}
}

// **Ein Teilerfolg bei der Übernahme geht auch nicht verloren** — dieselbe
// Klasse wie beim Einlesen, gefunden beim Weiterfragen nach Befund 1
// (#1291).
//
// `Paketweg.Uebernehmen` kopiert die Pakete und frischt **danach** den Katalog
// auf. Scheitert das, **liegen die Pakete schon im Verzeichnis** — und wer nur
// den Fehler bekommt, lässt sie unerkannt zurück.
func TestEinTeilerfolgBeiDerUebernahmeGehtNichtVerloren(t *testing.T) {
	probe := &pflegeprobe{fehler: errors.New("der Katalog ist zu")}

	_, err := mitPflege(t, probe, "schreiben.paket.uebernehmen").
		Ruf(t.Context(), json.RawMessage(`{"ort":"/tmp/mse2/data"}`))
	if err == nil {
		t.Fatal("kein Fehler")
	}

	// **Der Name und nicht die Zahl** (Befund 2 des Nachlaufs): „1 Pakete"
	// sagt nicht, was wegzuräumen ist.
	if !strings.Contains(err.Error(), "mse-game/magic") {
		t.Errorf("die Meldung nennt das übernommene Paket nicht: %v", err)
	}
}

// **Ein Einfahren ohne angelegtes Set nennt keine Kennung** — der andere Zweig
// von Befund 1 des Nachlaufs. Ohne ihn stünde „angelegt wurde Set 0" da.
func TestEinGescheitertesEinfahrenOhneSetNenntKeineKennung(t *testing.T) {
	probe := &ausfuhrprobe{fehler: errors.New("die Datei ist keine")}

	_, err := mitAusfuhr(t, probe, "schreiben.set.einfahren").
		Ruf(t.Context(), json.RawMessage(`{"pfad":"/tmp/keins.mse-set"}`))
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if strings.Contains(err.Error(), "angelegt wurde") {
		t.Errorf("die Meldung nennt ein Set, das es nicht gibt: %v", err)
	}
}

// **Und eine Übernahme ohne ein einziges Paket sagt »keine«.**
//
// Ohne diesen Zweig endete die Meldung auf „übernommen wurden bereits" — ein
// abgeschnittener Satz.
func TestEineUebernahmeOhnePaketSagtKeine(t *testing.T) {
	probe := &pflegeprobe{fehler: errors.New("der Katalog ist zu"), ohnePakete: true}

	_, err := mitPflege(t, probe, "schreiben.paket.uebernehmen").
		Ruf(t.Context(), json.RawMessage(`{"ort":"/tmp/leer"}`))
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(err.Error(), "keine") {
		t.Errorf("die Meldung sagt nicht »keine«: %v", err)
	}
}

// **Ein Archivschließfehler nennt das eingefahrene Set** — Befund 1 des
// Nachlaufs über Phase 4.4
// (#1291).
func TestEinArchivschliessfehlerNenntDasSetUeberMCP(t *testing.T) {
	_, err := mitPflege(t, &pflegeprobe{}, "schreiben.set.einfahren").
		Ruf(t.Context(), json.RawMessage(`{"pfad":"/tmp/halb.mse-set"}`))
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(err.Error(), "Set 7") {
		t.Errorf("die Meldung nennt das eingefahrene Set nicht: %v", err)
	}
}
