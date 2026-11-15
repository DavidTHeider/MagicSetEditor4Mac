package app_test

// **Ein Turnier als Datei und als Bericht, an der Naht** — DOM-04.AK-03
// (#1699),
// Vermerk 0140.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Als Text** — für MCP und REST, die an keiner Platte stehen.
func TestEinTurnierGehtAlsTextHinUndZurueck(t *testing.T) {
	t.Parallel()

	anwendung, naht := turnieranwendung(t)

	if ausgabe := anwendung.TurnierAusgabeText(t.Context(), "7"); ausgabe.Fehler != nil ||
		ausgabe.Kennung != "7" || ausgabe.Inhalt != "ausgabe von 7\n" {
		t.Errorf("die Ausgabe: %+v", ausgabe)
	}

	eingelesen := anwendung.TurnierEinlesenText(t.Context(), "eine Datei")
	if eingelesen.Fehler != nil || eingelesen.Kennung != "8" || string(naht.eingelesen) != "eine Datei" ||
		len(eingelesen.Befunde) != 1 || eingelesen.Befunde[0] != "ohne Format" {
		t.Errorf("das Einlesen: %+v", eingelesen)
	}

	// **Der Bericht steht in der Sprache der Anwendung** — ohne gewählte ist
	// es Deutsch.
	if bericht := anwendung.TurnierBerichtText(t.Context(), "7"); bericht.Fehler != nil ||
		bericht.Inhalt != "# Bericht mit Rangliste\n" || naht.worte.Bye != "Freilos" {
		t.Errorf("der Bericht: %+v, Wörter %+v", bericht, naht.worte)
	}
}

// **An der Platte** — für Fenster und Kommandozeile.
func TestEinTurnierGehtAlsDateiHinUndZurueck(t *testing.T) {
	t.Parallel()

	anwendung, naht := turnieranwendung(t)
	anwendung.MitTurnierdatei(service.NeueTurnieraustauschdatei())

	ordner := t.TempDir()
	ausgabe := filepath.Join(ordner, "freitag.mseturnier")

	if antwort := anwendung.TurnierAusgabeSchreiben(t.Context(), "7", ausgabe); antwort.Fehler != nil ||
		antwort.Pfad != ausgabe || antwort.Inhalt != "" {
		t.Errorf("die Ausgabe: %+v", antwort)
	}

	if inhalt, err := os.ReadFile(ausgabe); err != nil || string(inhalt) != "ausgabe von 7\n" { //nolint:gosec // ein Pfad dieses Laufs
		t.Errorf("in der Datei steht %q (%v)", inhalt, err)
	}

	if antwort := anwendung.TurnierEinlesenLesen(t.Context(), ausgabe); antwort.Fehler != nil ||
		antwort.Kennung != "8" || antwort.Pfad != ausgabe || string(naht.eingelesen) != "ausgabe von 7\n" {
		t.Errorf("das Einlesen: %+v", antwort)
	}

	bericht := filepath.Join(ordner, "freitag.md")
	if antwort := anwendung.TurnierBerichtSchreiben(t.Context(), "7", bericht); antwort.Fehler != nil ||
		antwort.Pfad != bericht {
		t.Errorf("der Bericht: %+v", antwort)
	}

	if inhalt, err := os.ReadFile(bericht); err != nil || string(inhalt) != "# Bericht mit Rangliste\n" { //nolint:gosec // dito
		t.Errorf("im Bericht steht %q (%v)", inhalt, err)
	}
}

// **Jeder Weg hat sein Nein** — eine krumme Kennung am Feld, ein Fehler der
// Leitung, eine Datei, die es nicht gibt, ein Ort, an den nichts geht.
func TestDieTurnierdateiwegeReichenIhreFehlerDurch(t *testing.T) {
	t.Parallel()

	anwendung, naht := turnieranwendung(t)
	anwendung.MitTurnierdatei(service.NeueTurnieraustauschdatei())

	nirgends := filepath.Join(t.TempDir(), "fehlt", "t.mseturnier")

	for name, antwort := range map[string]app.TurnierdateiAntwort{
		"Ausgabe als Text, krumm":     anwendung.TurnierAusgabeText(t.Context(), "sieben"),
		"Ausgabe als Datei, krumm":    anwendung.TurnierAusgabeSchreiben(t.Context(), "sieben", nirgends),
		"Bericht als Text, krumm":     anwendung.TurnierBerichtText(t.Context(), "sieben"),
		"Bericht als Datei, krumm":    anwendung.TurnierBerichtSchreiben(t.Context(), "sieben", nirgends),
		"Ausgabe an einen leeren Ort": anwendung.TurnierAusgabeSchreiben(t.Context(), "7", nirgends),
		"Bericht an einen leeren Ort": anwendung.TurnierBerichtSchreiben(t.Context(), "7", nirgends),
		"Einlesen, die Datei fehlt":   anwendung.TurnierEinlesenLesen(t.Context(), nirgends),
	} {
		if antwort.Fehler == nil || antwort.Befunde == nil {
			t.Errorf("%s: %+v", name, antwort)
		}
	}

	naht.fehler = domain.ErrNotFound

	for name, antwort := range map[string]app.TurnierdateiAntwort{
		"Ausgabe":  anwendung.TurnierAusgabeText(t.Context(), "7"),
		"Einlesen": anwendung.TurnierEinlesenText(t.Context(), "{}"),
		"Bericht":  anwendung.TurnierBerichtText(t.Context(), "7"),
	} {
		if antwort.Fehler == nil || antwort.Fehler.Code != app.CodeNichtGefunden {
			t.Errorf("%s: %+v", name, antwort.Fehler)
		}
	}
}

// **Ohne eingerichtete Datei ist es ein Nein** und kein Absturz — ein Fehler
// des Aufbaus, den ein Lauf ohne Composition Root sieht.
func TestOhneTurnierdateiGehtKeinDateiweg(t *testing.T) {
	t.Parallel()

	anwendung, _ := turnieranwendung(t)

	for name, antwort := range map[string]app.TurnierdateiAntwort{
		"Ausgabe":  anwendung.TurnierAusgabeSchreiben(t.Context(), "7", "/tmp/t.mseturnier"),
		"Einlesen": anwendung.TurnierEinlesenLesen(t.Context(), "/tmp/t.mseturnier"),
	} {
		if antwort.Fehler == nil {
			t.Errorf("%s: %+v", name, antwort)
		}
	}
}
