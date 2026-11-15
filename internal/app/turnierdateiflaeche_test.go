package app_test

// **Turnierdatei und Bericht über die Flächen ohne Fenster** — DOM-04.AK-03
// (#1699).
//
// **Gegen einen echten Bestand**, wie `turnierflaeche_test.go`: geprüft wird
// die Verdrahtung — Kommandozeile über die Fläche, MCP und REST über die
// Werkstatt.

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func TestDieTurnierflaecheGibtAusUndLiestEin(t *testing.T) {
	flaeche := app.TurnierflaecheUeber(frischerBestand(t))

	angelegt := flaeche.TurnierAnlegen(probekontext(t), "Freitag", "magic", "", false, "")
	if angelegt.Fehler != nil {
		t.Fatalf("anlegen: %+v", angelegt.Fehler)
	}

	ausgabe := flaeche.TurnierAusgabeText(probekontext(t), angelegt.Turnier.ID)
	if ausgabe.Fehler != nil || !strings.Contains(ausgabe.Inhalt, `"name": "Freitag"`) {
		t.Fatalf("ausgeben: %+v", ausgabe)
	}

	umbenannt := strings.Replace(ausgabe.Inhalt, `"Freitag"`, `"Samstag"`, 1)
	if eingelesen := flaeche.TurnierEinlesenText(probekontext(t), umbenannt); eingelesen.Fehler != nil ||
		eingelesen.Kennung == angelegt.Turnier.ID {
		t.Errorf("einlesen: %+v", eingelesen)
	}

	if bericht := flaeche.TurnierBerichtText(probekontext(t), angelegt.Turnier.ID); bericht.Fehler != nil ||
		!strings.HasPrefix(bericht.Inhalt, "# Freitag\n") {
		t.Errorf("berichten: %+v", bericht)
	}

	// **Und an der Platte** — die Kommandozeile.
	ordner := t.TempDir()
	datei := filepath.Join(ordner, "t.mseturnier")

	if antwort := flaeche.TurnierAusgabeSchreiben(probekontext(t), angelegt.Turnier.ID, datei); antwort.Fehler != nil {
		t.Errorf("in die Datei: %+v", antwort)
	}

	if antwort := flaeche.TurnierEinlesenLesen(probekontext(t), datei); antwort.Fehler == nil {
		t.Errorf("derselbe Name ein zweites Mal: %+v", antwort)
	}

	if antwort := flaeche.TurnierBerichtSchreiben(probekontext(t), angelegt.Turnier.ID,
		filepath.Join(ordner, "t.md")); antwort.Fehler != nil {
		t.Errorf("den Bericht in die Datei: %+v", antwort)
	}
}

func TestDieTurnierwerkstattGibtAusUndLiestEin(t *testing.T) {
	werkstatt := app.NeueTurnierwerkstatt(frischerBestand(t))

	angelegt, err := werkstatt.TurnierAnlegen(probekontext(t), "Freitag", "magic", "", false, 0)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	kennung := kennungAusAntwort(t, angelegt)

	ausgabe, err := werkstatt.TurnierAusgabe(probekontext(t), kennung)
	if err != nil {
		t.Fatalf("ausgeben: %v", err)
	}

	text, gut := ausgabe.(app.Turniertext)
	if !gut {
		t.Fatalf("die Ausgabe hat die Gestalt %T", ausgabe)
	}

	eingang, err := werkstatt.TurnierEinlesen(probekontext(t), strings.Replace(text.Inhalt, `"Freitag"`, `"Samstag"`, 1))
	if angekommen, gut := eingang.(app.Turniereingang); err != nil || !gut || angekommen.Kennung == "" {
		t.Errorf("einlesen: %+v, %v", eingang, err)
	}

	bericht, err := werkstatt.TurnierBericht(probekontext(t), kennung)
	if berichtet, gut := bericht.(app.Turniertext); err != nil || !gut || !strings.HasPrefix(berichtet.Inhalt, "# Freitag") {
		t.Errorf("berichten: %+v, %v", bericht, err)
	}

	// **Jedes Nein kommt als Fehler** — ein Turnier, das es nicht gibt, und
	// eine Datei, die keine ist.
	if _, err := werkstatt.TurnierAusgabe(probekontext(t), 99); err == nil {
		t.Error("die Ausgabe eines fehlenden Turniers gelang")
	}

	if _, err := werkstatt.TurnierBericht(probekontext(t), 99); err == nil {
		t.Error("der Bericht über ein fehlendes Turnier gelang")
	}

	if _, err := werkstatt.TurnierEinlesen(probekontext(t), "{"); err == nil {
		t.Error("eine kaputte Datei wurde eingelesen")
	}
}
