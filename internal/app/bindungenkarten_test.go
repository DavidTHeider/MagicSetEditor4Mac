package app_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die drei Wege zur Kartenliste gehen durch die Bindungen**
// (#891).
//
// **Eine eigene Datei**, weil `bindungenlesen_test.go` mit ihnen über 400
// Zeilen ginge (§2). Der Schnitt folgt der Frage: dort steht, dass **jede**
// lesende Bindung durchreicht, hier die drei, die dieselbe Fähigkeit in drei
// Mengen geben.
//
// # Was hier nicht geprüft wird
//
// **Ob die Abfrage dahinter stimmt** — das steht in
// `internal/storage/kartenzeilen_test.go`. Hier geht es um den Vertrag: die
// Bindung hängt den Kontext des Fensters an und gibt weiter.
func TestKartenlistenbindungenReichenDurch(t *testing.T) {
	bindungen := bindungsprobe(t).Bindungen

	// **Die drei Wege zur Kartenliste** (#891). Sie reichen durch wie `Karten`
	// — geprüft wird der Vertrag, nicht die Abfrage dahinter.
	t.Run("Kartenzeilen", func(t *testing.T) {
		antwort := bindungen.Kartenzeilen(t.Context(), "1")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if len(antwort.Zeilen) != 1 || antwort.Zeilen[0].Name != "Aus dem Bestand" {
			t.Errorf("%+v", antwort.Zeilen)
		}
	})

	t.Run("Karte", func(t *testing.T) {
		antwort := bindungen.Karte(t.Context(), "7")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if antwort.Karte.ID != "7" {
			t.Errorf("%+v", antwort.Karte)
		}
	})

	t.Run("Kartenfeldwerte", func(t *testing.T) {
		antwort := bindungen.Kartenfeldwerte(t.Context(), "1", "typ")
		if antwort.Fehler != nil {
			t.Fatalf("%+v", antwort.Fehler)
		}

		if len(antwort.Werte) != 1 || antwort.Werte[0].Wert != "Zauber" {
			t.Errorf("%+v", antwort.Werte)
		}
	})
}

// **Und die Listenbilder ebenso** (#1166).
//
// **Ohne Zeichenwerk gibt es keine**, und genau das ist hier zu sehen: die
// Bindungsprobe hat keine Bildquelle, die Antwort ist leer und trotzdem
// fehlerfrei. Eine Liste ohne Bilder ist eine Liste; eine Liste mit einer
// Fehlermeldung wäre etwas anderes.
func TestDieBindungFuerListenbilderReichtDurch(t *testing.T) {
	antwort := bindungsprobe(t).Bindungen.Kartenbilder(t.Context(), "1", []string{"7"})

	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Bilder) != 0 {
		t.Errorf("es kamen Bilder ohne Zeichenwerk: %v", antwort.Bilder)
	}
}

// **Die zwei Wege der Diagnose gehen durch die Bindungen** (#1166).
//
// **Die Dateiwahl wird belegt**, denn ohne sie riefe der Sicherungsweg den
// echten Wails-Dialog — und der weist einen Kontext ab, den es außerhalb des
// Fensters nicht gibt. Eine Vorrichtung, die das umginge, indem sie den Weg
// gar nicht führe, prüfte nichts.
func TestDieDiagnosebindungenReichenDurch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	probe := bindungsprobe(t)
	ziel := filepath.Join(t.TempDir(), "diagnose.txt")

	probe.Anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return ziel, nil
	})

	antwort := probe.Bindungen.Diagnose(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if !strings.Contains(antwort.Text, app.Fassung) {
		t.Errorf("die Diagnose ist leer:\n%s", antwort.Text)
	}

	if geschrieben := probe.Bindungen.DiagnoseSichern(t.Context()); geschrieben.Fehler != nil {
		t.Fatalf("%+v", geschrieben.Fehler)
	}

	if _, err := os.Stat(ziel); err != nil {
		t.Errorf("die Datei fehlt: %v", err)
	}
}
