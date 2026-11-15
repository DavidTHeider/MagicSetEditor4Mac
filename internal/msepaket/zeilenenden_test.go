package msepaket_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein einzelnes `\r` beendet eine Zeile, wie bei MSE** — FMT-06.AK-01
// (#1695).
//
// MSE2 liest seine Dateien selbst Zeile für Zeile und bricht an `\n`, an
// `\r\n` **und** an einem `\r`, dem kein `\n` folgt (`read_utf8_line`,
// `src/util/io/reader.cpp`, Zeilen 159–186, Fassung f9d9356 vom 2021-07-06).
// Hier brach der Zerteiler nur an `\n`: eine Datei mit den Zeilenenden des
// alten Mac OS war eine einzige Zeile, und ihr Kopf hieß
// `mse version: 2.0.0\rfull name: …`.
//
// **Und der Rundlauf hält**: jede Zeile schreibt ihr Ende so zurück, wie sie
// es gelesen hat — jetzt in drei Formen statt zwei.
func TestEinEinzelnesWagenruecklaufBeendetEineZeile(t *testing.T) {
	t.Parallel()

	inhalt := "mse version: 2.0.0\rfull name: Alt\r\ngame: magic\n"

	baum, err := msepaket.Zerteile("style", []byte(inhalt))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	for schluessel, erwartet := range map[string]string{
		"mse version": "2.0.0",
		"full name":   "Alt",
		"game":        "magic",
	} {
		knoten := baum.Ersten(schluessel)
		if knoten == nil || knoten.Wert != erwartet {
			t.Errorf("%s: %v, erwartet %q", schluessel, knoten, erwartet)
		}
	}

	if zurueck := string(msepaket.Schreibe(baum)); zurueck != inhalt {
		t.Errorf("Rundlauf: %q, erwartet %q", zurueck, inhalt)
	}
}

// **Ungültiges UTF-8 wird benannt, mit Zeile** — und nicht still gelesen.
//
// MSE2 wirft hier `Invalid UTF-8 sequence on line N` und liest nicht weiter
// (dieselbe Stelle, Zeilen 179–183). Dieses Programm liest nach §8 so weit
// weiter, wie es verantwortbar ist, und nennt die Stelle — `adapted`.
func TestUngueltigesUTF8WirdMitZeileBenannt(t *testing.T) {
	t.Parallel()

	_, err := msepaket.Zerteile("style", []byte("mse version: 2.0.0\nfull name: A\xffB\n"))
	if !errors.Is(err, msepaket.ErrKodierung) {
		t.Fatalf("Fehler %v, erwartet ErrKodierung", err)
	}

	var stelle *msepaket.Zerteilfehler
	if !errors.As(err, &stelle) || stelle.Stelle.Zeile != 2 {
		t.Errorf("Fehler %v — erwartet die Stelle in Zeile 2", err)
	}
}

// **UTF-16 kann weder MSE2 noch dieses Programm** — aber hier wird es gesagt,
// statt Nullbytes als Inhalt zu lesen.
func TestUTF16WirdAlsSolchesBenannt(t *testing.T) {
	t.Parallel()

	for name, roh := range map[string][]byte{
		"UTF-16 LE": {0xFF, 0xFE, 'm', 0, 's', 0, 'e', 0},
		"UTF-16 BE": {0xFE, 0xFF, 0, 'm', 0, 's', 0, 'e'},
	} {
		if _, err := msepaket.Zerteile("style", roh); !errors.Is(err, msepaket.ErrKodierung) {
			t.Errorf("%s: Fehler %v, erwartet ErrKodierung", name, err)
		}
	}
}
