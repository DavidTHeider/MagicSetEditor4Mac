package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Haushalt: Plätze je Ressourcenklasse** — Werkstück O5 von Phase 5.2
// (#1429).

func TestDerHaushaltVergibtPlaetzeJeKlasse(t *testing.T) {
	t.Parallel()

	grenzen := domain.Vorgabegrenzen()
	grenzen.Parallel[domain.RessourceEinfuhr] = 1
	haushalt := service.NeuerHaushalt(grenzen)

	frei, belegt := haushalt.Versuchen(domain.RessourceEinfuhr)
	if !belegt {
		t.Fatal("erster Platz verweigert")
	}

	if _, zweiter := haushalt.Versuchen(domain.RessourceEinfuhr); zweiter {
		t.Error("zweiter Platz bei Grenze 1 vergeben")
	}

	// Eine andere Klasse ist davon unberührt.
	if andere, ok := haushalt.Versuchen(domain.RessourceRendern); !ok {
		t.Error("Rendern durch Einfuhr blockiert")
	} else {
		andere()
	}

	// **Freigeben gibt frei — und ein zweites Freigeben gibt nichts doppelt.**
	frei()
	frei()

	if _, wieder := haushalt.Versuchen(domain.RessourceEinfuhr); !wieder {
		t.Error("nach Freigabe kein Platz")
	}

	if _, zuViel := haushalt.Versuchen(domain.RessourceEinfuhr); zuViel {
		t.Error("das doppelte Freigeben hat einen zweiten Platz erzeugt")
	}

	// Eine Klasse ohne Grenze ist frei.
	if _, ok := haushalt.Versuchen(domain.Ressourcenklasse("unbekannt")); !ok {
		t.Error("unbekannte Klasse gesperrt")
	}
}

func TestDerHaushaltWartetBisEinPlatzFreiWirdOderDerRufEndet(t *testing.T) {
	t.Parallel()

	grenzen := domain.Vorgabegrenzen()
	grenzen.Parallel[domain.RessourceAusfuhr] = 1
	haushalt := service.NeuerHaushalt(grenzen)

	frei, _ := haushalt.Versuchen(domain.RessourceAusfuhr)

	// **Ein Abbruch gibt auf — als Abbruch, nicht als Fehler des Bestands.**
	abgebrochen, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	if _, err := haushalt.Belegen(abgebrochen, domain.RessourceAusfuhr); !errors.Is(err, context.Canceled) {
		t.Errorf("abgebrochener Ruf: %v", err)
	}

	// Wird der Platz frei, kommt der Wartende dran.
	gekommen := make(chan func())

	go func() {
		freigeben, err := haushalt.Belegen(t.Context(), domain.RessourceAusfuhr)
		if err != nil {
			t.Error(err)
		}
		gekommen <- freigeben
	}()

	frei()

	select {
	case freigeben := <-gekommen:
		freigeben()
	case <-time.After(5 * time.Second):
		t.Fatal("der Wartende kam nicht dran")
	}

	// Eine Klasse ohne Grenze wartet nie.
	if freigeben, err := haushalt.Belegen(abgebrochen, domain.Ressourcenklasse("unbekannt")); err != nil {
		t.Errorf("unbekannte Klasse: %v", err)
	} else {
		freigeben()
	}

	if haushalt.Grenzen().Parallel[domain.RessourceAusfuhr] != 1 {
		t.Error("Grenzen nicht auskunftsfähig")
	}
}

// **Eine Klasse ohne Grenze hält niemanden auf.**
//
// # Warum das eine Zusage ist und kein Versehen
//
// `Haushalt.Versuchen` gibt für eine Klasse, die in `Parallel` nicht steht,
// einen Platz **und** eine Freigabe, die nichts tut. Das ist die ehrliche
// Antwort auf „wie viele dürfen gleichzeitig?" — **beliebig viele** —, und
// eine Sperre bei null wäre die falsche: eine unbekannte Klasse ist keine
// verbotene.
//
// **Angestrichen vom Riegel am 2026-09-11**: die leere Freigabe war der eine
// unbetretene Zweig in `internal/service`, und kein Lauf nahm sie — weil jede
// Klasse in `Vorgabegrenzen` eine Grenze hat.
func TestEineKlasseOhneGrenzeHaeltNiemandenAuf(t *testing.T) {
	t.Parallel()

	haushalt := service.NeuerHaushalt(domain.Vorgabegrenzen())

	const unbekannt domain.Ressourcenklasse = "vorlesen"

	frei, belegt := haushalt.Versuchen(unbekannt)
	if !belegt {
		t.Fatal("eine Klasse ohne Grenze hat keinen Platz gegeben")
	}

	// **Und ein zweiter Platz ebenso** — „beliebig viele" heißt beliebig viele.
	zweiter, nochmal := haushalt.Versuchen(unbekannt)
	if !nochmal {
		t.Error("der zweite Platz wurde verweigert")
	}

	// **Die Freigabe tut nichts und scheitert dabei nicht.**
	frei()
	zweiter()

	if _, danach := haushalt.Versuchen(unbekannt); !danach {
		t.Error("nach der Freigabe kein Platz mehr")
	}
}
