//go:build fremdpakete

// **Ein Set verliert keine Karte**
// (#539).
//
// # Warum dieser Wächter nötig wurde
//
// `Blue.mse-set` führt 3.693 Karten; importiert wurden **1.713**. Die
// fehlenden 1.980 sind 7,9 % des ganzen Kanons — und sie gingen **still**
// verloren: kein Fehler, keine Meldung, nur weniger Karten.
//
// Bemerkt wurde es an einer Nebenrechnung: der Kanonlauf zeichnete 23.168
// Karten, in den Dateien stehen 25.148, und die Differenz war auf den Kopf
// genau Blues Verlust.
//
// **Der Zähler des Laufs sah nichts davon** — er zählte „3.693 von 3.693", weil
// beide Zahlen aus derselben, schon verkürzten Liste kamen. Ein Wächter, der
// eine Vorrichtung gegen sich selbst hält, ist keiner (§10).
package app

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// TestEinSetVerliertKeineKarte hält jedes kanonische Set gegen seine Datei.
func TestEinSetVerliertKeineKarte(t *testing.T) {
	if os.Getenv("MSE_FREMDPAKETE") == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	anwendung := gestartetMit(t, func() string { return os.Getenv("MSE_FREMDPAKETE") })

	eintraege, err := os.ReadDir(kanonverzeichnis())
	if err != nil {
		t.Skip("die kanonischen Sets liegen nicht vor")
	}

	var inDenDateien, importiert int

	for _, eintrag := range eintraege {
		if filepath.Ext(eintrag.Name()) != ".mse-set" {
			continue
		}

		pfad := filepath.Join(kanonverzeichnis(), eintrag.Name())

		erwartet, lesefehler := kartenImArchiv(pfad)
		if lesefehler != nil {
			// **Rot und nicht übersprungen.** Hier stand `if erwartet == 0
			// { continue }`, und `kartenImArchiv` gab bei **jedem** Fehler
			// null — ein unlesbares Archiv, eine fehlende `set`-Datei, ein
			// Lesefehler. Der Wächter ließ damit genau die Fälle durch, in
			// denen er etwas zu sagen hätte (§10). Angestrichen vom
			// Zweitblick über Phase J.
			t.Errorf("%s: die Kartenzahl im Archiv ist nicht zu ermitteln: %v",
				eintrag.Name(), lesefehler)

			continue
		}

		geoeffnet := anwendung.SetOeffnen(t.Context(), pfad)
		if geoeffnet.Fehler != nil {
			t.Errorf("%s öffnen: %+v", eintrag.Name(), geoeffnet.Fehler)

			continue
		}

		gelesen := len(anwendung.Karten(t.Context(), geoeffnet.SetID).Karten)

		inDenDateien += erwartet
		importiert += gelesen

		if gelesen != erwartet {
			t.Errorf("%s: %d von %d Karten importiert — %d gingen still verloren",
				eintrag.Name(), gelesen, erwartet, erwartet-gelesen)
		}
	}

	t.Logf("%d Karten in den Dateien, %d importiert", inDenDateien, importiert)
}

// kartenImArchiv zählt die `card:`-Blöcke der `set`-Datei eines Archivs.
//
// **Am rohen Text und nicht über unseren Leser.** Ein Wächter, der die zu
// prüfende Vorrichtung benutzt, um ihre Zusage zu belegen, prüft nichts —
// genau daran ist der Verlust so lange unbemerkt geblieben.
//
// # Er scheitert laut
//
// Jeder Fehlschlag kommt als `error` heraus, nicht als null. Vorher gab diese
// Funktion bei einem unlesbaren Archiv, einer fehlenden `set`-Datei und einem
// Lesefehler dieselbe Zahl wie bei einem leeren Set — und der Aufrufer sprang
// darüber hinweg. Ein Wächter, der im Fehlerfall grün wird, ist keiner (§10).
//
// # Und was er **nicht** zählt
//
// Ein `card:` **am Zeilenanfang**, mit `\n` oder `\r\n` davor oder als erste
// Zeile der Datei. Eine Karte, die anders geschrieben stünde — eingerückt, in
// einem `include file:` — zählte er nicht. Beides kommt in den zehn
// kanonischen Sets nicht vor; die Zahl ist gegen den Leser gehalten, nicht
// gegen MSEs Zerteiler.
func kartenImArchiv(pfad string) (int, error) {
	archiv, err := zip.OpenReader(pfad) //nolint:gosec // ein Pfad aus dem Kanonverzeichnis (§3)
	if err != nil {
		return 0, fmt.Errorf("das Archiv öffnen: %w", err)
	}

	defer func() { _ = archiv.Close() }() //nolint:errcheck // gelesen ist gelesen

	for _, datei := range archiv.File {
		if datei.Name != "set" {
			continue
		}

		offen, oeffnungsfehler := datei.Open()
		if oeffnungsfehler != nil {
			return 0, fmt.Errorf("die Datei `set` öffnen: %w", oeffnungsfehler)
		}

		inhalt, lesefehler := io.ReadAll(offen)

		_ = offen.Close() //nolint:errcheck // gelesen ist gelesen

		if lesefehler != nil {
			return 0, fmt.Errorf("die Datei `set` lesen: %w", lesefehler)
		}

		return kartenkoepfe(inhalt), nil
	}

	return 0, errors.New("das Archiv führt keine Datei `set`")
}

// kartenkoepfe zählt die Zeilen, die mit `card:` beginnen.
//
// **`\r\n` zählt mit.** MSE schreibt zwar `\n`, aber ein Set, das einmal durch
// ein Werkzeug unter Windows gegangen ist, trüge das andere — und dieser
// Wächter fiele dann mit „0 von 3.693" auf, obwohl nichts verloren wäre.
func kartenkoepfe(inhalt []byte) int {
	anzahl := bytes.Count(inhalt, []byte("\ncard:"))

	if bytes.HasPrefix(inhalt, []byte("card:")) {
		anzahl++
	}

	return anzahl
}
