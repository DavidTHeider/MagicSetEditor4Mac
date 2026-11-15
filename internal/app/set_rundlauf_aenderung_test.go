package app_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Ein geändertes Set behält, was es nicht versteht** — FMT-06.AK-02
// (#1695).
//
// # Was der Rundlauf ohne Änderung nicht zeigt
//
// `TestEinDurchgereichtesSetIstByteFuerByteDasselbe` gibt ein Set zurück, an
// dem niemand etwas getan hat. Die Frage aus AK-02 ist eine andere: **nach
// einer Änderung** — bleiben der unverstandene Block, die Kommentare, die
// eigenwilligen Abstände der übrigen Zeilen, das BOM und die gemischten
// Zeilenenden stehen? Und was geschieht mit der geänderten Zeile selbst?
//
// # Was mit ihr geschieht, und das ist die Normalisierung
//
// **Sie verliert ihre Rohform** (`service.werteSetzen`): Einzug per Tab, ein
// Leerzeichen hinter dem Doppelpunkt. **Ihr Zeilenende behält sie** —
// `\r\n` bleibt `\r\n`. Alles andere geht Byte für Byte hinaus. Die Regeln
// stehen in `docs/FORMATE.md` unter „Was beim Zurückschreiben normalisiert
// wird".
func TestEinGeaendertesSetBehaeltWasEsNichtVersteht(t *testing.T) {
	anwendung := gestartet(t)

	quelle := kopiereVerzeichnis(t, filepath.Join("..", "..", "testdaten", "eigenwillig.mse-set"))

	geoeffnet := anwendung.SetOeffnen(t.Context(), quelle)
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	zeilen := anwendung.Kartenzeilen(t.Context(), geoeffnet.SetID)
	if zeilen.Fehler != nil {
		t.Fatalf("Kartenzeilen: %+v", zeilen.Fehler)
	}

	kennung := map[string]string{}
	for _, zeile := range zeilen.Zeilen {
		kennung[zeile.Name] = zeile.ID
	}

	// **Zwei Zeilen, zwei Fälle**: `name` der ersten Karte endet auf `\r\n`
	// und trägt vier Leerzeichen, `farbe` der zweiten zwei.
	for _, aenderung := range []struct{ karte, feld, wert string }{
		{"Feuerball", "name", "Feuerwand"},
		{"Eisball", "farbe", "grün"},
	} {
		if gesetzt := anwendung.FeldSetzen(probekontext(t), geoeffnet.SetID,
			kennung[aenderung.karte], aenderung.feld, aenderung.wert); gesetzt.Fehler != nil {
			t.Fatalf("%s.%s setzen: %+v", aenderung.karte, aenderung.feld, gesetzt.Fehler)
		}
	}

	ziel := filepath.Join(t.TempDir(), "geaendert.mse-set")

	if gespeichert := anwendung.SetSpeichern(t.Context(), geoeffnet.SetID, ziel); gespeichert.Fehler != nil {
		t.Fatalf("speichern: %+v", gespeichert.Fehler)
	}

	vorher := lies(t, filepath.Join(quelle, "set"))
	nachher := lies(t, filepath.Join(ziel, "set"))

	erwartet := bytes.Replace(vorher, []byte("\tname:    Feuerball\r\n"), []byte("\tname: Feuerwand\r\n"), 1)
	erwartet = bytes.Replace(erwartet, []byte("\tfarbe:  blau\n"), []byte("\tfarbe: grün\n"), 1)

	if bytes.Equal(erwartet, vorher) {
		t.Fatal("die Erwartung ist die Quelle — die Ersetzung hat ihre Zeile nicht gefunden")
	}

	if !bytes.Equal(nachher, erwartet) {
		t.Errorf("die Datei ist eine andere als erwartet:\n%s", unterschied(erwartet, nachher))
	}

	// **Und wieder gelesen ändert sich nichts mehr.** Ein zweiter Rundlauf
	// über das Geschriebene muss es Byte für Byte zurückgeben — sonst wäre die
	// Normalisierung keine, sondern eine Drift, die mit jedem Speichern
	// weiterwandert.
	zweite := gestartet(t)

	wieder := zweite.SetOeffnen(t.Context(), ziel)
	if wieder.Fehler != nil {
		t.Fatalf("wieder öffnen: %+v", wieder.Fehler)
	}

	nochmal := filepath.Join(t.TempDir(), "nochmal.mse-set")

	if gespeichert := zweite.SetSpeichern(t.Context(), wieder.SetID, nochmal); gespeichert.Fehler != nil {
		t.Fatalf("nochmal speichern: %+v", gespeichert.Fehler)
	}

	if zweimal := lies(t, filepath.Join(nochmal, "set")); !bytes.Equal(zweimal, nachher) {
		t.Errorf("der zweite Rundlauf ändert die Datei:\n%s", unterschied(nachher, zweimal))
	}
}

// **Ein Wert mit führendem Leerraum übersteht das Wiederlesen.**
//
// Hinter dem Doppelpunkt liest MSE2 mit `trim_left` (`src/util/io/reader.cpp`,
// Zeile 223, Fassung f9d9356 vom 2021-07-06), und dieser Zerteiler genauso.
// Ein Wert, der mit Leerraum beginnt, lässt sich deshalb **nicht** hinter den
// Doppelpunkt schreiben — MSE2 schreibt ihn als Block eine Ebene tiefer, wie
// einen mehrzeiligen (`src/util/io/writer.cpp`, Zeilen 69–104, dieselbe
// Fassung).
func TestEinWertMitFuehrendemLeerraumUebersteht(t *testing.T) {
	anwendung := gestartet(t)

	quelle := kopiereVerzeichnis(t, filepath.Join("..", "..", "testdaten", "probe.mse-set"))

	geoeffnet := anwendung.SetOeffnen(t.Context(), quelle)
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	erste := ersteKartenkennung(t, anwendung, geoeffnet.SetID)

	const eingerueckt = "  eingerückt"

	if gesetzt := anwendung.FeldSetzen(probekontext(t), geoeffnet.SetID, erste,
		"name", eingerueckt); gesetzt.Fehler != nil {
		t.Fatalf("setzen: %+v", gesetzt.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "eingerueckt.mse-set")

	if gespeichert := anwendung.SetSpeichern(t.Context(), geoeffnet.SetID, ziel); gespeichert.Fehler != nil {
		t.Fatalf("speichern: %+v", gespeichert.Fehler)
	}

	zweite := gestartet(t)

	wieder := zweite.SetOeffnen(t.Context(), ziel)
	if wieder.Fehler != nil {
		t.Fatalf("wieder öffnen: %+v", wieder.Fehler)
	}

	karte := zweite.Karte(t.Context(), ersteKartenkennung(t, zweite, wieder.SetID))

	for _, wert := range karte.Karte.Werte {
		if wert.Feld == "name" && wert.Wert != eingerueckt {
			t.Errorf("wieder gelesen %q, geschrieben %q", wert.Wert, eingerueckt)
		}
	}
}

// ersteKartenkennung gibt die Kennung der ersten Karte eines Sets.
func ersteKartenkennung(t *testing.T, anwendung *app.App, setID string) string {
	t.Helper()

	zeilen := anwendung.Kartenzeilen(t.Context(), setID)
	if zeilen.Fehler != nil || len(zeilen.Zeilen) == 0 {
		t.Fatalf("Kartenzeilen: %+v", zeilen)
	}

	return zeilen.Zeilen[0].ID
}

// gestartet gibt eine Anwendung über einem frischen Bestand.
func gestartet(t *testing.T) *app.App {
	t.Helper()

	anwendung := app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	return anwendung
}
