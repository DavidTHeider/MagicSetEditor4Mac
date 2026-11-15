package mtgjson_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// In einer AllSets-Datei ist der Schlüssel der Setcode.
//
// MTGJSON schreibt ihn zusätzlich in den Rumpf, aber verlassen kann man sich
// darauf nicht: eine Datei von Hand lässt ihn weg (§8). Ohne diesen Rückgriff
// hieße jedes Set dort schlicht nicht.
func TestSetcodeKommtNotfallsAusDemSchluessel(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "ohnecode.json")

	inhalt := `{"data": {"OHN": {"name": "Set ohne Code im Rumpf", "cards": []}}}`
	if err := os.WriteFile(ort, []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	quelle, err := mtgjson.Lies(ort, mtgjson.ZuordnungMagic)
	if err != nil {
		t.Fatalf("Lies: %v", err)
	}

	if len(quelle.Sets) != 1 || quelle.Sets[0].Code != "OHN" {
		t.Errorf("Sets: %+v, erwartet eines mit dem Code »OHN«", quelle.Sets)
	}
}

// Ein leerer Wert legt kein Kartenfeld an.
//
// `"manaCost": ""` steht in MTG JSON für Karten ohne Manakosten. Daraus ein
// leeres Feld zu machen hieße, eine Angabe zu erfinden, die die Datei nicht
// macht — und beim Schreiben stünde sie dann da.
func TestLeererWertLegtKeinFeldAn(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "leer.json")

	inhalt := `{"code":"LEE","cards":[{"name":"Ohne Kosten","manaCost":"","rarity":"common"}]}`
	if err := os.WriteFile(ort, []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	quelle, err := mtgjson.Lies(ort, mtgjson.ZuordnungMagic)
	if err != nil {
		t.Fatalf("Lies: %v", err)
	}

	if _, da := wertVon(quelle.Sets[0].Karten[0].Werte, "casting cost"); da {
		t.Error("»casting cost« ist angelegt, obwohl die Datei es leer führt")
	}
}

// Eine Wandlung überlebt einen Wert, den die Karte nicht führt.
//
// Ein Feldpaar darf mehrere Schlüssel verbrauchen, und eine Karte führt selten
// alle. Eine Wandlung, die dabei abstürzte, machte jede mehrspaltige Zuordnung
// zu einer Falle — über eine Wails-Bindung nicht als Fehler, sondern als
// beendete Anwendung.
func TestWandlungUeberlebtEinenFehlendenWert(t *testing.T) {
	faelle := map[string]mtgjson.Wandlung{
		"Manakosten": mtgjson.Manakosten,
		"Seltenheit": mtgjson.Seltenheit,
		"Wörter":     mtgjson.Woerter,
	}

	for name, wandlung := range faelle {
		t.Run(name, func(t *testing.T) {
			for _, werte := range [][]json.RawMessage{nil, {nil}} {
				text, err := wandlung(werte)
				if err != nil {
					t.Errorf("%v: %v", werte, err)
				}

				if text != "" {
					t.Errorf("%v: %q, erwartet leer", werte, text)
				}
			}
		})
	}
}

// Was sich im Verzeichnis nicht lesen lässt, wird mit **seiner** Datei gemeldet.
//
// Ein Verzeichnis mit 300 Sets und einer kaputten Datei: die Meldung muss
// sagen, welche es ist. „Irgendeine Datei in diesem Verzeichnis" ließe jemanden
// 300 Dateien durchsehen.
func TestKaputteDateiImVerzeichnisNenntIhrenNamen(t *testing.T) {
	verzeichnis := t.TempDir()

	if err := os.WriteFile(filepath.Join(verzeichnis, "gut.json"),
		[]byte(`{"code":"GUT","cards":[]}`), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if err := os.WriteFile(filepath.Join(verzeichnis, "schlecht.json"),
		[]byte("{ kein JSON"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	_, err := mtgjson.Lies(verzeichnis, mtgjson.ZuordnungMagic)
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(err.Error(), "schlecht.json") {
		t.Errorf("Fehler: %q — die kaputte Datei ist nicht benannt", err)
	}
}

// Ein `data`, unter dem weder ein Set noch eine Abbildung von Sets steht.
func TestUmschlagOhneSetsWirdGemeldet(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "seltsam.json")

	if err := os.WriteFile(ort, []byte(`{"data": [1, 2]}`), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	_, err := mtgjson.Lies(ort, mtgjson.ZuordnungMagic)
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(err.Error(), "seltsam.json") {
		t.Errorf("Fehler: %q — die Datei ist nicht benannt", err)
	}
}

// Was sich nicht öffnen lässt, wird gemeldet statt übergangen.
//
// Eine Datei ohne Leserecht liegt da und ist trotzdem nicht zu lesen. Wer den
// Fehler verschluckt, öffnet ein Verzeichnis mit 300 Sets und zeigt 299, ohne
// dass jemand erfährt, welches fehlt.
func TestFehlendesLeserechtWirdGemeldet(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("als root ist jede Datei lesbar")
	}

	verzeichnis := t.TempDir()

	datei := filepath.Join(verzeichnis, "verschlossen.json")
	if err := os.WriteFile(datei, []byte(`{"code":"ZU","cards":[]}`), 0o000); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	zu := filepath.Join(verzeichnis, "zu")
	if err := os.Mkdir(zu, 0o000); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	// Zurückgeben, was `t.TempDir()` sonst nicht aufräumen kann — und den
	// Fehler dabei melden: ein stiller Fehlschlag hier ließe ein Verzeichnis
	// ohne Rechte liegen, und der nächste Lauf scheiterte an etwas ganz anderem.
	//
	// `0700` und nicht `0600`: ein Verzeichnis braucht das Ausführungsrecht,
	// sonst kommt auch das Aufräumen nicht hinein. gosec misst hier mit dem Maß
	// für Dateien.
	t.Cleanup(func() {
		if err := os.Chmod(zu, 0o700); err != nil { // #nosec G302 -- Verzeichnis, siehe darüber
			t.Errorf("Rechte zurückgeben: %v", err)
		}
	})

	for name, ort := range map[string]string{"eine Datei": datei, "ein Verzeichnis": zu} {
		t.Run(name, func(t *testing.T) {
			if _, err := mtgjson.Lies(ort, mtgjson.ZuordnungMagic); err == nil {
				t.Errorf("Lies(%q): kein Fehler", ort)
			}
		})
	}
}
