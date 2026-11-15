package app_test

// **Was das Fenster über ein Paket erfährt** — DOM-01.AK-01
// (#1696):
// Herkunftsart, lokale Änderung, wer es braucht, wie viele Sets es benutzen.

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestDasFensterSiehtHerkunftAenderungBedarfUndNutzung(t *testing.T) {
	fuehrer := &paketfuehrer{
		pakete: []domain.Paket{{
			Art: "mse-game", Name: "magic", Pfad: "/p/magic.mse-game",
			Pruefsumme: "sha256:1", Geaendert: true,
		}},
		zusatz: service.Paketzeile{
			Herkunftsart: domain.PaketBezogen,
			GebrauchtVon: []string{"magic-m15.mse-style"},
			Benutzt:      3,
		},
	}

	anwendung := &app.App{}
	anwendung.MitPaketen(fuehrer)

	antwort := anwendung.Pakete(t.Context())
	if antwort.Fehler != nil || len(antwort.Pakete) != 1 {
		t.Fatalf("die Pakete: %+v", antwort)
	}

	paket := antwort.Pakete[0]
	if paket.Herkunftsart != "bezogen" || !paket.Geaendert || paket.Benutzt != 3 ||
		!slices.Equal(paket.GebrauchtVon, []string{"magic-m15.mse-style"}) {
		t.Errorf("das Paket: %+v", paket)
	}

	// **Leer bleibt eine Liste** — `null` erreichte das Fenster als etwas,
	// das es nicht durchgehen kann (§3).
	fuehrer.zusatz = service.Paketzeile{}

	if leer := anwendung.Pakete(t.Context()).Pakete[0]; leer.GebrauchtVon == nil {
		t.Error("GebrauchtVon kam als null")
	}
}
