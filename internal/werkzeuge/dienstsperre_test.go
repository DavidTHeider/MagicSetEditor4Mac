package werkzeuge_test

// **Wer das Fenster mit eigenem Bestand startet, sperrt den Dienst** — QA01
// (#2382),
// gefunden in der Gesamtabnahme am 2026-10-20.
//
// # Wofür es diesen Wächter gibt
//
// **Die WebKit-Abnahme schrieb die Vorgaben dieses Rechners um.** Sie
// startete das Fenster im Servermodus mit eigenem Bestand und umgelenktem
// `HOME`, und der Kennungsumzug (R02, Vermerk 0191) legte unter der neuen
// Kennung eine Domäne mit den Werten der alten an: `defaults` geht über
// cfprefsd, und cfprefsd folgt keinem umgelenkten Zuhause — auch nicht
// `CFFIXED_USER_HOME`, nachgemessen am selben Tag. Der Wächter für echte
// Orte fand die neue Datei. **Die Sperre dafür gab es**
// (`MSE_DIENST_GESPERRT`, `internal/app/schnittstellenziele.go`); nur setzte
// dieser Läufer sie nicht, und die Startkonfiguration der Sichtprüfung auch
// nicht.
//
// # Was er hält
//
// **Jeder Läufer, der das Fensterprogramm mit eigenem Bestand startet, setzt
// die Sperre** — die Hüllenprobe, die WebKit-Abnahme, die Sichtprüfung und
// jeder, der dazukommt.
//
// # Was er nicht hält
//
// **Ob die Sperre im Kind ankommt** — das sieht erst der Wächter für echte
// Orte am Ende eines Laufs. Und einen Läufer **ohne** eigenen Bestand: der
// arbeitet am echten und darf auch dessen Vorgaben.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// fensterstart erkennt, dass ein Läufer das Fensterprogramm startet.
var fensterstart = regexp.MustCompile(`bauen\.sh (browser|abnahme)|entwickeln\.sh|build/bin/MagicSetEditor4Mac`)

// eigenerBestand erkennt, dass ein Läufer den Bestand **setzt** — als
// Zuweisung in der Shell oder als Eintrag einer Umgebung. Wer ihn nur
// durchreicht (`entwickeln.sh`), arbeitet am Bestand seines Aufrufers.
var eigenerBestand = regexp.MustCompile(`MSE_BESTAND(=|:|['"]\s*:)`)

func TestWerDasFensterMitEigenemBestandStartetSperrtDenDienst(t *testing.T) {
	t.Parallel()

	dateien := []string{filepath.Join("..", "..", ".claude", "launch.json")}

	for _, muster := range []string{"*.sh", "*.mjs"} {
		gefunden, err := filepath.Glob(filepath.Join("..", "..", "scripts", muster))
		if err != nil {
			t.Fatal(err)
		}

		dateien = append(dateien, gefunden...)
	}

	geprueft := 0

	for _, datei := range dateien {
		roh, err := os.ReadFile(datei) //nolint:gosec // eine Datei des Baums
		if err != nil {
			t.Fatalf("%s: %v", datei, err)
		}

		inhalt := string(roh)
		if !eigenerBestand.MatchString(inhalt) || !fensterstart.MatchString(inhalt) {
			continue
		}

		geprueft++

		if !strings.Contains(inhalt, "MSE_DIENST_GESPERRT") {
			t.Errorf("%s startet das Fenster mit eigenem Bestand und sperrt den Dienst nicht — MSE_DIENST_GESPERRT=an setzen",
				filepath.ToSlash(strings.TrimPrefix(datei, filepath.Join("..", "..")+string(filepath.Separator))))
		}
	}

	// **Eine Positivkontrolle**: findet das Muster keinen Läufer, hält der
	// Wächter nichts.
	if geprueft < 2 {
		t.Errorf("nur %d Läufer erkannt — das Muster greift nicht mehr", geprueft)
	}
}
