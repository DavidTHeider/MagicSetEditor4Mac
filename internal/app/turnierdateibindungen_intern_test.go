package app

// **Turnierdatei und Bericht an der gebundenen Oberfläche** — DOM-04.AK-03
// (#1699).
//
// **Ein innerer Lauf**, weil er `NeueBindungen` samt Dateidialog braucht —
// dieselbe Naht, die das Fenster benutzt. Die Leitung ist die echte über
// einem Bestand im Arbeitsspeicher: was hier geprüft wird, ist der ganze Weg
// vom Dialog bis zur Zeile.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mitTurnierdatei baut eine Anwendung samt Dialog, der `gewaehlt` antwortet.
func mitTurnierdatei(t *testing.T, gewaehlt *string, wahlfehler error) *Bindungen {
	t.Helper()

	anwendung := TurnierflaecheUeber(probedatenbank(t)).app

	// **Der Kontext des Fensters trägt ein Profil**, wie nach dem Start.
	anwendung.fenster = probekontext(t)
	anwendung.MitDateiwahl(func(context.Context, Dateiauswahl) (string, error) {
		return *gewaehlt, wahlfehler
	})

	return NeueBindungen(anwendung)
}

func TestEinTurnierGehtUeberDenDialogHinUndZurueck(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()
	gewaehlt := filepath.Join(ordner, "freitag.mseturnier")
	bindungen := mitTurnierdatei(t, &gewaehlt, nil)

	angelegt := bindungen.app(t.Context()).TurnierAnlegen(bindungen.ruf(t.Context()), "Freitag", "magic", "", false, "")
	if angelegt.Fehler != nil {
		t.Fatalf("anlegen: %+v", angelegt.Fehler)
	}

	if antwort := bindungen.TurnierAusgabeSpeichern(t.Context(), angelegt.Turnier.ID); antwort.Fehler != nil ||
		antwort.Pfad != gewaehlt {
		t.Fatalf("ausgeben: %+v", antwort)
	}

	// **Derselbe Name ist ein Konflikt** — der Speicher hält ihn je Spiel
	// eindeutig. Mit einem anderen geht es.
	if antwort := bindungen.TurnierEinlesenOeffnen(t.Context()); antwort.Fehler == nil {
		t.Errorf("zweimal derselbe Name: %+v", antwort)
	}

	roh, _ := os.ReadFile(gewaehlt) //nolint:errcheck,gosec // ein Pfad dieses Laufs; geprüft über den Inhalt
	umbenannt := strings.Replace(string(roh), `"Freitag"`, `"Samstag"`, 1)

	if err := os.WriteFile(gewaehlt, []byte(umbenannt), 0o600); err != nil { //nolint:gosec // ein Pfad dieses Laufs
		t.Fatalf("umbenennen: %v", err)
	}

	if antwort := bindungen.TurnierEinlesenOeffnen(t.Context()); antwort.Fehler != nil || antwort.Kennung == "" {
		t.Errorf("einlesen: %+v", antwort)
	}

	gewaehlt = filepath.Join(ordner, "freitag.md")
	if antwort := bindungen.TurnierBerichtSpeichern(t.Context(), angelegt.Turnier.ID); antwort.Fehler != nil {
		t.Errorf("berichten: %+v", antwort)
	}

	if bericht, err := os.ReadFile(gewaehlt); err != nil || !strings.HasPrefix(string(bericht), "# Freitag\n") { //nolint:gosec // dito
		t.Errorf("im Bericht steht %q (%v)", bericht, err)
	}
}

// **Weggeklickt ist ein Abbruch, ein klagender Dialog ein Fehler** — bei
// allen drei Wegen.
func TestDieTurnierdialogeUnterscheidenAbbruchVonFehler(t *testing.T) {
	t.Parallel()

	leer := ""
	weggeklickt := mitTurnierdatei(t, &leer, nil)
	klagend := mitTurnierdatei(t, &leer, errors.New("kein Fenster"))

	for name, antwort := range map[string]TurnierdateiAntwort{
		"Ausgabe":  weggeklickt.TurnierAusgabeSpeichern(t.Context(), "1"),
		"Einlesen": weggeklickt.TurnierEinlesenOeffnen(t.Context()),
		"Bericht":  weggeklickt.TurnierBerichtSpeichern(t.Context(), "1"),
	} {
		if !antwort.Abgebrochen || antwort.Fehler != nil {
			t.Errorf("%s weggeklickt: %+v", name, antwort)
		}
	}

	for name, antwort := range map[string]TurnierdateiAntwort{
		"Ausgabe":  klagend.TurnierAusgabeSpeichern(t.Context(), "1"),
		"Einlesen": klagend.TurnierEinlesenOeffnen(t.Context()),
		"Bericht":  klagend.TurnierBerichtSpeichern(t.Context(), "1"),
	} {
		if antwort.Abgebrochen || antwort.Fehler == nil {
			t.Errorf("%s klagend: %+v", name, antwort)
		}
	}
}
