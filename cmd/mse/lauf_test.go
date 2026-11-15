// Die Kommandozeile, von außen gefahren
// (#886).
//
// **Über `lauf` und nicht über `main`.** Jenes greift nach den Strömen und
// beendet den Prozess; hier wird geprüft, was dabei herauskommt — dieselbe
// Trennung wie in `cmd/mse-mcp`.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// fahre führt einen Aufruf gegen einen frischen Bestand und gibt alles zurück.
func fahre(t *testing.T, argumente ...string) (code int, aus, fehler string) {
	t.Helper()

	var raus, rfehler bytes.Buffer

	code = lauf(probekontext(t), argumente, nil, &raus, &rfehler,
		frischerBestand(t))

	return code, raus.String(), rfehler.String()
}

// **Ohne Argumente sagt sie, was sie kann.**
//
// Eine Kommandozeile, die auf einen nackten Aufruf nichts sagt, ist eine
// Sackgasse — und der einzige Weg hinein ist, die Quelle zu lesen.
func TestOhneArgumenteKommtDieHilfe(t *testing.T) {
	code, _, fehler := fahre(t)

	if code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet war %d", code, CodeBedienung)
	}

	if !strings.Contains(fehler, "sets") {
		t.Errorf("die Hilfe nennt den Befehl `sets` nicht: %q", fehler)
	}
}

// **Ein unbekannter Befehl ist ein Bedienfehler, kein fachlicher.**
//
// Für ein Skript ist das der Unterschied zwischen „mein Aufruf stimmt nicht"
// und „die Antwort ist nein".
func TestEinUnbekannterBefehlIstEinBedienfehler(t *testing.T) {
	code, _, fehler := fahre(t, "gibtsnicht")

	if code != CodeBedienung {
		t.Errorf("Rückgabewert %d, erwartet war %d", code, CodeBedienung)
	}

	if !strings.Contains(fehler, "gibtsnicht") {
		t.Errorf("die Meldung nennt den Befehl nicht: %q", fehler)
	}
}

// **`sets` auf einem leeren Bestand gibt eine leere Liste — und keinen Fehler.**
//
// „Noch keins" ist eine Antwort, kein Fehlschlag.
func TestSetsAufLeeremBestand(t *testing.T) {
	code, aus, fehler := fahre(t, "sets")

	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	if strings.TrimSpace(aus) == "" {
		t.Error("die Ausgabe ist leer — auch „kein Set\" gehört gesagt")
	}
}

// **Und mit `--json` ist sie maschinenlesbar.**
//
// **Eine leere Liste bleibt eine Liste** (§3): `null` statt `[]` ist genau der
// Fehler, den `omitempty` auf einem Slice erzeugt und den kein Test zwischen
// den Prüfebenen sieht.
func TestSetsAlsJSON(t *testing.T) {
	code, aus, fehler := fahre(t, "sets", "--json")

	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, fehler)
	}

	var gelesen struct {
		Sets []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"sets"`
	}

	if err := json.Unmarshal([]byte(aus), &gelesen); err != nil {
		t.Fatalf("die Ausgabe ist kein JSON: %v\n%s", err, aus)
	}

	if gelesen.Sets == nil {
		t.Error("`sets` ist null statt einer leeren Liste")
	}
}

// **Die Hilfe kommt auch auf Verlangen — und dann nicht als Fehler.**
func TestHilfeAufVerlangen(t *testing.T) {
	code, aus, _ := fahre(t, "hilfe")

	if code != CodeGut {
		t.Errorf("Rückgabewert %d, erwartet war %d", code, CodeGut)
	}

	if !strings.Contains(aus, "mse") {
		t.Errorf("die Hilfe nennt den Programmnamen nicht: %q", aus)
	}
}

// Hier stand `TestEinAbbruchWirdGemeldet` — ein Abbruch bei `sets`, geprüft
// gegen `CodeFachlich`.
//
// **Er ist in `abbruch_test.go` aufgegangen**, und zwar aus zwei Gründen. Er
// prüfte **einen** der sechs Befehle, und ein Abbruch, der bei einem greift
// und beim nächsten nicht, ist keiner. Und seine erwartete Zahl ist mit #886
// eine andere geworden: ein Abbruch meldet 130 und nicht 1 (siehe
// [CodeAbgebrochen]).
//
// Die Zusage, die sein Name trug — geordnet enden statt in einer Panik —,
// hält der neue Fall für **jeden** Befehl mit.

// stockenderStrom scheitert nach der ersten Zeile — wie eine geschlossene Röhre.
type stockenderStrom struct{ geschrieben int }

func (s *stockenderStrom) Write(daten []byte) (int, error) {
	s.geschrieben++
	if s.geschrieben > 1 {
		return 0, errors.New("die Röhre ist zu")
	}

	return len(daten), nil
}

// **Eine abgebrochene Ausgabe wird gemeldet, nicht verschwiegen.**
//
// `mse sets --json | head -1` schließt die Röhre, sobald `head` genug hat. Ein
// Programm, das den Schreibfehler wegwirft und trotzdem 0 zurückgibt, sagt
// einem Skript, es habe alles ausgegeben.
func TestEineAbgebrocheneAusgabeWirdGemeldet(t *testing.T) {
	// Zwei Zeilen: die Hilfe geht in einem Zug, `sets` auf leerem Bestand
	// ebenso — deshalb wird ein Strom gebraucht, der beim **zweiten** Mal
	// scheitert, und zwei Schreibvorgänge.
	strom := &stockenderStrom{}

	var rfehler bytes.Buffer

	ort := frischerBestand(t)

	if code := lauf(probekontext(t), []string{"hilfe"}, nil, strom, &rfehler, ort); code != CodeGut {
		t.Fatalf("der erste Aufruf scheiterte schon: %d", code)
	}

	code := lauf(probekontext(t), []string{"hilfe"}, nil, strom, &rfehler, ort)
	if code != CodeFachlich {
		t.Errorf("Rückgabewert %d, erwartet war %d", code, CodeFachlich)
	}

	if !strings.Contains(rfehler.String(), "nicht vollständig") {
		t.Errorf("die Meldung sagt nicht, was fehlte: %q", rfehler.String())
	}
}

// mitSet legt ein Set im Bestand an und gibt dessen Ort.
//
// **Über `storage` und nicht über die Kommandozeile selbst**: geprüft werden
// soll die Ausgabe, nicht der Weg hinein — den gibt es hier noch nicht.
func mitSet(t *testing.T) string {
	t.Helper()

	ort := frischerBestand(t)

	db, err := storage.Oeffnen(probekontext(t), ort)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	if _, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Probeset", Game: "probe", Style: "probe-schlicht"}, nil); fehler != nil {
		t.Fatalf("Set anlegen: %v", fehler)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	return ort
}

// **Ein vorhandenes Set steht mit Kennung, Namen, Spiel und Stil da.**
//
// Vier Spalten, durch Tabulatoren getrennt — das ist die Form, die `cut` und
// `awk` lesen, und sie ist damit eine **Zusage** (§16).
func TestSetsZeigtEinVorhandenesSet(t *testing.T) {
	var raus, rfehler bytes.Buffer

	code := lauf(probekontext(t), []string{"sets"}, nil, &raus, &rfehler, mitSet(t))
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, rfehler.String())
	}

	zeile := strings.TrimSpace(raus.String())
	spalten := strings.Split(zeile, "\t")

	if len(spalten) != 4 {
		t.Fatalf("die Zeile hat %d Spalten, erwartet waren 4: %q", len(spalten), zeile)
	}

	if spalten[1] != "Probeset" || spalten[2] != "probe" || spalten[3] != "probe-schlicht" {
		t.Errorf("die Zeile lautet %q", zeile)
	}
}

// **Und als JSON trägt sie dieselben vier Angaben.**
func TestSetsAlsJSONZeigtEinVorhandenesSet(t *testing.T) {
	var raus, rfehler bytes.Buffer

	code := lauf(probekontext(t), []string{"sets", "--json"}, nil, &raus, &rfehler, mitSet(t))
	if code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, rfehler.String())
	}

	var gelesen setsAntwort
	if err := json.Unmarshal(raus.Bytes(), &gelesen); err != nil {
		t.Fatalf("kein JSON: %v", err)
	}

	if len(gelesen.Sets) != 1 {
		t.Fatalf("%d Sets, erwartet war eines", len(gelesen.Sets))
	}

	if gelesen.Sets[0].Name != "Probeset" || gelesen.Sets[0].ID == "" {
		t.Errorf("das Set kam als %+v", gelesen.Sets[0])
	}
}

// **`--json` darf vor dem Befehl stehen.**
//
// Wer es vorn schreibt, meint dasselbe wie wer es anhängt; eine
// Kommandozeile, die darauf besteht, prüft die Gewohnheit statt der Absicht.
func TestDieAusgabeformDarfVornStehen(t *testing.T) {
	var raus, rfehler bytes.Buffer

	if code := lauf(probekontext(t), []string{"--json", "sets"}, nil, &raus, &rfehler,
		frischerBestand(t)); code != CodeGut {
		t.Fatalf("Rückgabewert %d (%s)", code, rfehler.String())
	}

	if !strings.HasPrefix(strings.TrimSpace(raus.String()), "{") {
		t.Errorf("die Ausgabe ist kein JSON: %q", raus.String())
	}
}

// **Ein Bestand, der sich nicht anlegen lässt, endet fachlich.**
func TestEinUnbrauchbarerBestandsortEndetFachlich(t *testing.T) {
	// Eine Datei am Ort des Verzeichnisses: `MkdirAll` scheitert daran.
	ort := filepath.Join(t.TempDir(), "keinverzeichnis")
	if err := os.WriteFile(ort, []byte("x"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	var raus, rfehler bytes.Buffer

	code := lauf(probekontext(t), []string{"sets"}, nil, &raus, &rfehler, filepath.Join(ort, "bestand.db"))
	if code != CodeFachlich {
		t.Errorf("Rückgabewert %d, erwartet war %d", code, CodeFachlich)
	}

	if rfehler.Len() == 0 {
		t.Error("es kam keine Meldung")
	}
}

// **Ein Bestand, der sich nicht öffnen lässt, endet fachlich.**
//
// Ein Verzeichnis am Ort der Datei: `MkdirAll` gelingt, `Oeffnen` nicht.
func TestEinNichtOeffenbarerBestandEndetFachlich(t *testing.T) {
	// **Ohne Musterbestand**: hier soll am Ort der Datei ein Verzeichnis
	// stehen, und eine kopierte Datei stünde ihm im Weg.
	ort := filepath.Join(t.TempDir(), "bestand.db")
	if err := os.MkdirAll(ort, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	var raus, rfehler bytes.Buffer

	if code := lauf(probekontext(t), []string{"sets"}, nil, &raus, &rfehler, ort); code != CodeFachlich {
		t.Errorf("Rückgabewert %d, erwartet war %d", code, CodeFachlich)
	}

	if rfehler.Len() == 0 {
		t.Error("es kam keine Meldung")
	}
}
