// `mse karten` und `mse feld` — lesen und schreiben auf der Kommandozeile
// (#886).
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// mitKarte legt ein Set mit einer Karte an und gibt Ort, Set- und Kartenkennung.
//
// **Und lenkt `HOME` um** (§3, „Ein Test, der nach draußen greift, ohne dass
// jemand es wollte"). `app.Schreibnaht` und jede andere Naht rufen
// `Paketverzeichnis()`, und das rechnet über `os.UserConfigDir()` — also über
// `HOME`. Ohne Umlenkung läsen diese Tests die **bezogenen Pakete dieses
// Rechners**, und was sie prüfen, hinge daran, wer sie fährt. Gefunden am
// 2026-08-25 beim Bau von `mse setangaben`
// (#1107).
//
// **Mit der Saat darin**, denn ohne ein Spielpaket findet der Stilwähler
// nichts — und ein Test gegen ein Set ohne Stil prüfte den Fehlerweg statt der
// Fähigkeit.
func mitKarte(t *testing.T) (ort, setID, karteID string) {
	t.Helper()

	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	saatBereitstellen(t, zuhause)

	ort = frischerBestand(t)

	db, err := storage.Oeffnen(probekontext(t), ort)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	kennung, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "probe", Style: "probe-schlicht",
		Cards: []*domain.Card{{Values: []domain.Value{{Field: "name", Text: "Feuerball"}}}},
	}, nil)
	if fehler != nil {
		t.Fatalf("Set anlegen: %v", fehler)
	}

	karten, fehler := storage.NeuerSetLeser(db).Karten(probekontext(t), kennung)
	if fehler != nil || len(karten) != 1 {
		t.Fatalf("Karten lesen: %v (%d)", fehler, len(karten))
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	return ort, strconv.FormatInt(kennung, 10), strconv.FormatInt(karten[0].ID, 10)
}

// fahreAn führt einen Aufruf gegen einen benannten Bestand.
func fahreAn(t *testing.T, ort string, argumente ...string) (code int, aus, fehler string) {
	t.Helper()

	var raus, rfehler bytes.Buffer

	code = lauf(probekontext(t), argumente, nil, &raus, &rfehler, ort)

	return code, raus.String(), rfehler.String()
}

// **`karten` nennt Kennung, Nummer und Namen.**
func TestKartenZeigtDieKartenEinesSets(t *testing.T) {
	ort, setID, karteID := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "karten", setID)
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if !strings.Contains(aus, karteID) || !strings.Contains(aus, "Feuerball") {
		t.Errorf("die Ausgabe nennt Karte oder Namen nicht: %q", aus)
	}
}

// **Ohne Setkennung ist es ein Bedienfehler.**
//
// Nicht ein fachlicher: der Aufruf ist unvollständig, nicht die Antwort nein.
func TestKartenOhneSetIstEinBedienfehler(t *testing.T) {
	code, _, fehler := fahre(t, "karten")

	if code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet war %d", code, CodeBedienung)
	}

	if !strings.Contains(fehler, "karten") {
		t.Errorf("die Meldung sagt nicht, was fehlt: %q", fehler)
	}
}

// **Ein Set, das es nicht gibt, ist fachlich nein.**
func TestKartenEinesUnbekanntenSets(t *testing.T) {
	code, _, fehler := fahre(t, "karten", "4711")

	if code != CodeFachlich {
		t.Errorf("Rückgabewert %d, erwartet war %d (%s)", code, CodeFachlich, fehler)
	}
}

// **Und eine Kennung, die keine Zahl ist, ist ein Bedienfehler.**
func TestKartenMitUnlesbarerKennung(t *testing.T) {
	code, _, fehler := fahre(t, "karten", "keinezahl")

	if code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet war %d (%s)", code, CodeBedienung, fehler)
	}
}

// **`feld` schreibt — und der neue Wert steht danach da.**
//
// Gefahren, nicht behauptet: geschrieben und wieder gelesen.
func TestFeldSetztEinenWert(t *testing.T) {
	ort, setID, karteID := mitKarte(t)

	code, _, fehler := fahreAn(t, ort, "feld", setID, karteID, "name", "Eisball")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	_, aus, _ := fahreAn(t, ort, "karten", setID, "--json")

	var gelesen kartenAntwort
	if err := json.Unmarshal([]byte(aus), &gelesen); err != nil {
		t.Fatalf("kein JSON: %v", err)
	}

	if len(gelesen.Karten) != 1 || gelesen.Karten[0].Name != "Eisball" {
		t.Errorf("nach dem Schreiben: %+v", gelesen.Karten)
	}
}

// **Die Antwort nennt den vorigen Wert** — damit ein Skript zurückkann.
func TestFeldNenntDenVorigenWert(t *testing.T) {
	ort, setID, karteID := mitKarte(t)

	code, aus, fehler := fahreAn(t, ort, "feld", setID, karteID, "name", "Eisball", "--json")
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var gelesen feldAntwort
	if err := json.Unmarshal([]byte(aus), &gelesen); err != nil {
		t.Fatalf("kein JSON: %v", err)
	}

	if gelesen.Vorher != "Feuerball" {
		t.Errorf("der vorige Wert kam als %q", gelesen.Vorher)
	}

	if !gelesen.Hatte {
		t.Error("die Antwort sagt, es habe das Feld nicht gegeben")
	}
}

// **Zu wenige Argumente sind ein Bedienfehler.**
func TestFeldMitZuWenigArgumenten(t *testing.T) {
	code, _, _ := fahre(t, "feld", "1", "2", "name")

	if code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet war %d", code, CodeBedienung)
	}
}

// **Und die Kommandozeile sagt sich an.**
//
// Der Verlauf trägt `kommandozeile`, nicht `fenster` — eine Rücknahme, die
// nicht sagt, wessen Änderung sie umkehrt, ist im Mehr-Schreiber-Betrieb eine
// Vermutung (#52).
func TestDieKommandozeileSagtSichAn(t *testing.T) {
	ort, setID, karteID := mitKarte(t)

	if code, _, fehler := fahreAn(t, ort, "feld", setID, karteID, "name", "Eisball"); code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	db, err := storage.Oeffnen(probekontext(t), ort)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur zum Aufräumen

	var urheber string
	if fehler := db.QueryRowContext(probekontext(t),
		`SELECT urheber FROM history ORDER BY id DESC LIMIT 1`).Scan(&urheber); fehler != nil {
		t.Fatalf("den Verlauf lesen: %v", fehler)
	}

	if urheber != string(domain.UrheberKommandozeile) {
		t.Errorf("der Verlauf nennt %q, erwartet war %q", urheber, domain.UrheberKommandozeile)
	}
}

// saatBereitstellen legt die mitgelieferte Saat als Paketverzeichnis ab.
//
// **Kopiert und nicht verlinkt**: der Paketleser folgt Verzeichnissen, und ein
// Symlink auf den Arbeitsbaum machte den Test von dessen Rechten abhängig.
func saatBereitstellen(t *testing.T, zuhause string) {
	t.Helper()

	quelle, err := filepath.Abs(filepath.Join("..", "..", "saat"))
	if err != nil {
		t.Fatalf("Saatpfad: %v", err)
	}

	// **Die ganze Saat und nicht zwei Pakete daraus.** Der erste Anlauf nahm
	// `probe.mse-game` und `probe-schlicht.mse-style`; damit hatte `mse stile`
	// genau einen Stil zur Wahl, und ein Wechsel ließ sich nicht prüfen.
	ziel := filepath.Join(zuhause, "Library", "Application Support",
		"MagicSetEditor4Mac", "pakete")

	if kopierfehler := os.CopyFS(ziel, os.DirFS(quelle)); kopierfehler != nil {
		t.Fatalf("Saat kopieren: %v", kopierfehler)
	}
}
