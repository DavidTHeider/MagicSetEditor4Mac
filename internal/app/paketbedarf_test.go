package app_test

// **Was einem Set an Paketen fehlt** — DOM-01.AK-04
// (#1696):
// am Fenster, am Paketweg der übrigen Flächen, und mit seinen Fehlern.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDerPaketbedarfEinesSetsKommtAlsDTO(t *testing.T) {
	fuehrer := &paketfuehrer{bedarf: []service.Abhaengigkeitsbefund{{
		Paket: "Erstes", Braucht: "probe.mse-game", Art: service.AbhaengigkeitFehlt,
	}}}

	anwendung := &app.App{}
	anwendung.MitPaketen(fuehrer)

	antwort := anwendung.Paketbedarf(t.Context(), "7")
	if antwort.Fehler != nil || fuehrer.gefragtesSet != 7 || len(antwort.Befunde) != 1 ||
		antwort.Befunde[0].Braucht != "probe.mse-game" || antwort.Befunde[0].Art != "fehlt" {
		t.Errorf("der Bedarf: %+v, gefragt %d", antwort, fuehrer.gefragtesSet)
	}

	// **Über die Bindung derselbe Weg.**
	if gebunden := app.NeueBindungen(anwendung).Paketbedarf(t.Context(), "7"); len(gebunden.Befunde) != 1 {
		t.Errorf("die Bindung: %+v", gebunden)
	}

	if antwort := anwendung.Paketbedarf(t.Context(), "keine Zahl"); antwort.Fehler == nil || antwort.Befunde == nil {
		t.Errorf("eine kaputte Kennung: %+v", antwort)
	}

	fuehrer.fehler = errors.New("klemmt")
	if antwort := anwendung.Paketbedarf(t.Context(), "7"); antwort.Fehler == nil || antwort.Befunde == nil {
		t.Errorf("ein Fehler: %+v", antwort)
	}
}

func TestDerPaketwegNenntDenBedarf(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	weg, nahtfehler := app.PaketnahtUeber(frischerBestand(t))
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	weg.MitVerwalter(&paketfuehrer{bedarf: []service.Abhaengigkeitsbefund{{
		Paket: "Erstes", Braucht: "probe.mse-game", Art: service.AbhaengigkeitFehlt,
	}}})

	if befunde, err := weg.Setbedarf(t.Context(), 7); err != nil || len(befunde) != 1 || befunde[0].Art != "fehlt" {
		t.Errorf("der Bedarf: %+v, %v", befunde, err)
	}

	weg.MitVerwalter(&paketfuehrer{fehler: errors.New("klemmt")})

	if _, err := weg.Setbedarf(t.Context(), 7); err == nil {
		t.Error("der Fehler ging verloren")
	}
}
