// **Eine Sicherung lässt sich in der Anwendung zurückspielen** (#940).
//
// Der Lebenszyklus wird **gefahren**: eine echte Datenbank wird angelegt,
// beschrieben, gesichert, verändert, zurückgespielt — und danach gelesen. Ein
// Test, der nur die Bindung ruft und den Rückgabewert ansieht, sagt nichts
// darüber, ob die Anwendung danach noch läuft.
package app_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// laufendeAnwendung startet eine Anwendung über einer frischen Datei.
func laufendeAnwendung(t *testing.T, bestand string) *app.App {
	t.Helper()

	anwendung := app.NeueApp(bestand)
	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		// **Beenden darf hier scheitern.** Nach einem Zurückspielen hat die
		// Anwendung eine andere Datenbank offen als die, die dieser Test
		// angelegt hat; ein Fehler beim Schließen sagt nichts über den Fall.
		_ = anwendung.Beenden() //nolint:errcheck // siehe darüber
	})

	return anwendung
}

// merkmal liest die Spracheinstellung — die Marke, an der dieser Test den
// Stand erkennt.
//
// **Eine Einstellung und keine Karte**: sie geht durch dieselbe Datenbank,
// braucht aber kein Paket, kein Spiel und keinen Stilbogen — der Test soll das
// Zurückspielen prüfen und nicht den Import.
func merkmal(t *testing.T, anwendung *app.App) string {
	t.Helper()

	antwort := anwendung.Einstellungen(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("Einstellungen lesen: %+v", antwort.Fehler)
	}

	for _, eintrag := range antwort.Einstellungen {
		if eintrag.Schluessel == "sprache" {
			return eintrag.Wert
		}
	}

	return ""
}

// **Der ganze Weg, an einer echten Datei.**
func TestSicherungZurueckspielenStelltDenAltenStandHer(t *testing.T) {
	ort := t.TempDir()
	bestand := filepath.Join(ort, "bestand.db")
	sicherung := filepath.Join(ort, "sicherung.db")

	anwendung := laufendeAnwendung(t, bestand)

	if antwort := anwendung.EinstellungSetzen(t.Context(), "sprache", "de"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	// Gesichert wird über denselben Weg, den auch das Fenster nimmt.
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return sicherung, nil
	})

	if antwort := anwendung.BestandSichern(t.Context(), "", false); antwort.Fehler != nil {
		t.Fatalf("sichern: %+v", antwort.Fehler)
	}

	// **Danach ändert sich der Stand** — das ist es, was zurückgenommen wird.
	if antwort := anwendung.EinstellungSetzen(t.Context(), "sprache", "en"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	if got := merkmal(t, anwendung); got != "en" {
		t.Fatalf("vor dem Zurückspielen steht %q, erwartet war %q", got, "en")
	}

	antwort := anwendung.SicherungZurueckspielen(t.Context(), sicherung, "")
	if antwort.Fehler != nil {
		t.Fatalf("zurückspielen: %+v", antwort.Fehler)
	}

	// **Und jetzt die eigentliche Frage: läuft sie noch, und was steht darin?**
	if got := merkmal(t, anwendung); got != "de" {
		t.Errorf("nach dem Zurückspielen steht %q, erwartet war %q", got, "de")
	}

	// Der ersetzte Stand liegt beiseite und ist nicht fort.
	if _, err := os.Stat(filepath.Join(ort, app.Rueckspielsicherung)); err != nil {
		t.Errorf("der ersetzte Bestand liegt nicht beiseite: %v", err)
	}
}

// **Eine Sicherung aus einer neueren Fassung wird benannt abgewiesen** — und
// zwar **bevor** sie den Bestand ersetzt.
func TestSicherungZurueckspielenWeistEineNeuereFassungAb(t *testing.T) {
	ort := t.TempDir()
	bestand := filepath.Join(ort, "bestand.db")
	fremd := filepath.Join(ort, "fremd.db")

	anwendung := laufendeAnwendung(t, bestand)

	if antwort := anwendung.EinstellungSetzen(t.Context(), "sprache", "de"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return fremd, nil
	})

	if antwort := anwendung.BestandSichern(t.Context(), "", false); antwort.Fehler != nil {
		t.Fatalf("sichern: %+v", antwort.Fehler)
	}

	zukunftImArchiv(t, fremd)

	antwort := anwendung.SicherungZurueckspielen(t.Context(), fremd, "")
	if antwort.Fehler == nil {
		t.Fatal("die Sicherung aus der Zukunft wurde angenommen")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "9999_aus_der_zukunft.sql") {
		t.Errorf("die Meldung nennt die unbekannte Migration nicht: %q", antwort.Fehler.Meldung)
	}

	// **Der Bestand ist unangetastet** — die Prüfung stand davor, nicht danach.
	if got := merkmal(t, anwendung); got != "de" {
		t.Errorf("der laufende Bestand hat sich verändert: %q", got)
	}

	if _, err := os.Stat(filepath.Join(ort, app.Rueckspielsicherung)); !os.IsNotExist(err) {
		t.Error("es wurde etwas beiseitegelegt, obwohl gar nicht getauscht wurde")
	}
}

// **Eine Datei, die keine Sicherung ist**, kostet den laufenden Stand nicht.
func TestSicherungZurueckspielenWeistFremdeDateienAb(t *testing.T) {
	ort := t.TempDir()
	bestand := filepath.Join(ort, "bestand.db")
	fremd := filepath.Join(ort, "notiz.txt")

	anwendung := laufendeAnwendung(t, bestand)

	if antwort := anwendung.EinstellungSetzen(t.Context(), "sprache", "de"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	if err := os.WriteFile(fremd, []byte("kein Bestand"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if antwort := anwendung.SicherungZurueckspielen(t.Context(), fremd, ""); antwort.Fehler == nil {
		t.Fatal("eine Textdatei wurde als Sicherung angenommen")
	}

	if got := merkmal(t, anwendung); got != "de" {
		t.Errorf("der laufende Bestand hat sich verändert: %q", got)
	}
}

// **Die Prüfung ohne das Zurückspielen** — was das Fenster zeigt, bevor jemand
// bestätigt.
func TestSicherungPruefenNenntWasDrinSteht(t *testing.T) {
	ort := t.TempDir()
	sicherung := filepath.Join(ort, "sicherung.db")

	anwendung := laufendeAnwendung(t, filepath.Join(ort, "bestand.db"))
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return sicherung, nil
	})

	if antwort := anwendung.BestandSichern(t.Context(), "", false); antwort.Fehler != nil {
		t.Fatalf("sichern: %+v", antwort.Fehler)
	}

	antwort := anwendung.SicherungWaehlen(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("prüfen: %+v", antwort.Fehler)
	}

	if antwort.Pfad != sicherung {
		t.Errorf("der Pfad ist %q, erwartet war %q", antwort.Pfad, sicherung)
	}

	if antwort.AngelegtAm == "" {
		t.Error("der Befund sagt nicht, wann der Bestand angelegt wurde")
	}

	if antwort.Schemastand == "" {
		t.Error("der Befund sagt nicht, bis wohin das Schema reicht")
	}
}

// **Ein Abbruch im Dateidialog ist kein Fehler.**
func TestSicherungWaehlenNimmtEinenAbbruchHin(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return "", nil
	})

	antwort := anwendung.SicherungWaehlen(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("ein Abbruch wurde als Fehler gemeldet: %+v", antwort.Fehler)
	}

	if !antwort.Abgebrochen {
		t.Error("der Abbruch ist nicht als solcher gemeldet")
	}
}

// **Ein Fehler im Dateidialog ist einer.**
func TestSicherungWaehlenMeldetEinenFehlerDesDialogs(t *testing.T) {
	anwendung := app.NeueApp("")
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return "", errors.New("das Fenster antwortet nicht")
	})

	antwort := anwendung.SicherungWaehlen(t.Context())
	if antwort.Fehler == nil {
		t.Fatal("der Fehler des Dialogs wurde übergangen")
	}

	if antwort.Abgebrochen {
		t.Error("ein Fehler wurde als Abbruch gemeldet")
	}
}

// **Eine Datei, die keine Sicherung ist, wird schon bei der Auswahl gemeldet.**
func TestSicherungWaehlenMeldetEineUnbrauchbareDatei(t *testing.T) {
	fremd := filepath.Join(t.TempDir(), "notiz.txt")
	if err := os.WriteFile(fremd, []byte("kein Bestand"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return fremd, nil
	})

	antwort := anwendung.SicherungWaehlen(t.Context())
	if antwort.Fehler == nil {
		t.Fatal("die fremde Datei wurde angenommen")
	}

	// **Der Pfad kommt trotzdem zurück** — die Meldung soll sagen können,
	// welche Datei gemeint war.
	if antwort.Pfad != fremd {
		t.Errorf("der Pfad ist %q, erwartet war %q", antwort.Pfad, fremd)
	}
}

// **Die beiden Bindungen reichen durch** — mit dem Kontext des Fensters.
//
// Ohne diesen Test blieben sie ungemessen, und ein `b.app().SicherungWaehlen`
// statt `b.app().SicherungZurueckspielen` fiele niemandem auf: beide bekommen
// einen Pfad und geben eine Antwort mit einem Fehler.
func TestDieBindungenSpielenZurueck(t *testing.T) {
	ort := t.TempDir()
	sicherung := filepath.Join(ort, "sicherung.db")

	anwendung := laufendeAnwendung(t, filepath.Join(ort, "bestand.db"))
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return sicherung, nil
	})

	if antwort := anwendung.BestandSichern(t.Context(), "", false); antwort.Fehler != nil {
		t.Fatalf("sichern: %+v", antwort.Fehler)
	}

	bindungen := app.NeueBindungen(anwendung)

	// **Der Startbefund geht über dieselbe Fassade** (#962). Ohne diese Zeile
	// wäre die Bindung angemeldet, gebunden, im Fähigkeitskatalog geführt —
	// und von keinem Lauf betreten. Genau die Gestalt von #852.
	if befund := bindungen.Startbefund(t.Context()); !befund.Offen {
		t.Errorf("der Startbefund sagt %+v, obwohl die Anwendung läuft", befund)
	}

	auswahl := bindungen.SicherungWaehlen(t.Context())
	if auswahl.Fehler != nil {
		t.Fatalf("wählen: %+v", auswahl.Fehler)
	}

	if auswahl.Pfad != sicherung {
		t.Errorf("gewählt wurde %q, erwartet war %q", auswahl.Pfad, sicherung)
	}

	antwort := bindungen.SicherungZurueckspielen(t.Context(), sicherung, "")
	if antwort.Fehler != nil {
		t.Fatalf("zurückspielen: %+v", antwort.Fehler)
	}

	if !antwort.Ersetzt {
		t.Error("die Antwort meldet keinen Tausch")
	}

	if antwort.Beiseite == "" {
		t.Error("die Antwort sagt nicht, wohin der bisherige Bestand kam")
	}
}
