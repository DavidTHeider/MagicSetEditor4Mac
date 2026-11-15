package app_test

// **Das Profilbild über den Dialog und über den Flächenweg** — Werkstück 8
// von Phase 4.12
// (#1351).
//
// # Warum jeder Lauf hier sein eigenes Zuhause bekommt
//
// **`NutzerflaecheUeber` baut ihre eigene Anwendung**, und deren `bilderort`
// ist der echte: `app.Bilderverzeichnis()` hängt an `os.UserConfigDir()`, und
// das liest `HOME`. Ein Lauf ohne `t.Setenv` schriebe in das **echte**
// Benutzerverzeichnis dieses Rechners — §3 nennt genau diesen Fall, und er
// wäre grün, weil er wirklich funktioniert hat.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Dialog gibt einen Pfad, und das Bild landet.**
func TestDerBilddialogLegtDasGewaehlteAb(t *testing.T) {
	anwendung := mitBilderort(t)
	pfad := bilddatei(t, 20, 20)

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return pfad, nil
	})

	antwort := anwendung.NutzerbildWaehlen(t.Context(), "1", "1")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Nutzer.Bild == "" {
		t.Error("der Nutzer trägt kein Bild")
	}
}

// **Ein Abbruch ist kein Fehler und keine Änderung.**
//
// **Beides gehört geprüft.** Eine leere Antwort ohne Fehler sagt „nichts
// gewählt"; dass dabei auch nichts geschrieben wurde, sagt erst die zweite
// Zusicherung.
func TestEinAbgebrochenerBilddialogSchreibtNichts(t *testing.T) {
	anwendung := mitBilderort(t)

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return "", nil
	})

	antwort := anwendung.NutzerbildWaehlen(t.Context(), "1", "1")
	if antwort.Fehler != nil || antwort.Nutzer != nil {
		t.Errorf("%+v — %+v", antwort.Fehler, antwort.Nutzer)
	}

	if bilder := anwendung.Nutzerbilder(t.Context()); len(bilder.Bilder) != 0 {
		t.Errorf("es wurde doch etwas abgelegt: %v", bilder.Bilder)
	}
}

// **Und ein gescheiterter Dialog wird gemeldet.**
func TestEinGescheiterterBilddialogWirdGemeldet(t *testing.T) {
	anwendung := mitBilderort(t)

	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		return "", errors.New("kein Fenster")
	})

	if antwort := anwendung.NutzerbildWaehlen(t.Context(), "1", "1"); antwort.Fehler == nil {
		t.Error("der Fehler ging verloren")
	}
}

// **Der Flächenweg legt ab und nimmt weg** — dieselbe Anwendung für die drei
// Flächen ohne Fenster (§16).
func TestDerNutzerflaechenwegSetztUndNimmtDasBild(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	flaeche := app.NutzerflaecheUeber(offeneProbedatenbank(t))

	gesetzt := flaeche.NutzerbildSetzen(probekontext(t), "1", "1", bilddatei(t, 20, 20))
	if gesetzt.Fehler != nil {
		t.Fatalf("%+v", gesetzt.Fehler)
	}

	ort := filepath.Join(zuhause, "Library", "Application Support",
		"MagicSetEditor4Mac", "beiwerk", "nutzer", gesetzt.Nutzer.Bild)

	if _, err := os.Stat(ort); err != nil {
		t.Errorf("die Datei fehlt: %v", err)
	}

	entfernt := flaeche.NutzerbildEntfernen(probekontext(t), "1", gesetzt.Nutzer.Fassung)
	if entfernt.Fehler != nil {
		t.Fatalf("%+v", entfernt.Fehler)
	}

	if _, err := os.Stat(ort); !os.IsNotExist(err) {
		t.Errorf("die Datei liegt noch: %v", err)
	}
}

// **Der Flächenweg reicht die Einstellungen und die Auskunft durch** — §16:
// dieselbe Anwendung für die drei Flächen ohne Fenster.
func TestDerNutzerflaechenwegFuehrtEinstellungenUndAuskunft(t *testing.T) {
	flaeche := app.NutzerflaecheUeber(offeneProbedatenbank(t))

	gesetzt := flaeche.NutzereinstellungSetzen(probekontext(t), "1",
		domain.NutzerKontrast, domain.KontrastHoch)
	if gesetzt.Fehler != nil {
		t.Fatalf("%+v", gesetzt.Fehler)
	}

	gelesen := flaeche.Nutzereinstellungen(probekontext(t), "1")
	if gelesen.Fehler != nil {
		t.Fatalf("%+v", gelesen.Fehler)
	}

	if len(gelesen.Einstellungen) != len(domain.Nutzereinstellungsreihe()) {
		t.Errorf("es kamen %d", len(gelesen.Einstellungen))
	}

	auskunft := flaeche.Personendaten(probekontext(t), "1")
	if auskunft.Fehler != nil {
		t.Fatalf("%+v", auskunft.Fehler)
	}

	if auskunft.Daten.Nutzer.Name == "" {
		t.Error("der Name fehlt")
	}
}

// **Und der Flächenweg reicht auch das Abzeichen durch** — Werkstück 9
// (#1351), Punkt 13.
func TestDerNutzerflaechenwegSetztUndNimmtDasAbzeichen(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	flaeche := app.NutzerflaecheUeber(offeneProbedatenbank(t))

	gesetzt := flaeche.ProfilbadgeSetzen(probekontext(t), "1", "1", bilddatei(t, 20, 20))
	if gesetzt.Fehler != nil {
		t.Fatalf("%+v", gesetzt.Fehler)
	}

	// **Unter dem Namen, den die Antwort nennt** — er trägt seit dem
	// Zweitblick über Phase 6.1 den Inhalt.
	ort := filepath.Join(zuhause, "Library", "Application Support",
		"MagicSetEditor4Mac", "beiwerk", "nutzer", gesetzt.Profil.Badge)

	if _, err := os.Stat(ort); err != nil {
		t.Errorf("die Datei fehlt: %v", err)
	}

	entfernt := flaeche.ProfilbadgeEntfernen(
		probekontext(t), "1", gesetzt.Profil.Fassung)
	if entfernt.Fehler != nil {
		t.Fatalf("%+v", entfernt.Fehler)
	}

	if _, err := os.Stat(ort); !os.IsNotExist(err) {
		t.Errorf("die Datei liegt noch: %v", err)
	}
}
