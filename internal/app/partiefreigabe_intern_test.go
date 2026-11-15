package app

// **Eine Partie für Gäste öffnen** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// **Dass erreichbar sein und mitspielen lassen zwei Entscheidungen sind.** Der
// Horchposten macht diesen Rechner ansprechbar; die Freigabe sagt, **was** ein
// Gegenüber dann darf. Ohne sie läuft jede angenommene Sitzung ins Leere.
//
// **Und dass die Wirkungsklasse `externEmpfangen` ist** — von hier an ändert
// ein fremder Rechner den Stand dieser Partie. Eine Tabelle, die eine Rolle
// nennt, die nichts durchsetzt, verspricht mehr, als ihr Rumpf hält (§10).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// freigabeanwendung baut eine Anwendung mit einer schweigsamen Leitung.
func freigabeanwendung(t *testing.T) *App {
	t.Helper()

	anwendung := NeueApp("")
	anwendung.db = bestandFuerDenRiegel(t)
	anwendung.profil.Store(1)

	return anwendung.MitPartienaht(&schweigsameLeitung{})
}

// **Ohne Freigabe steht nichts offen.**
func TestOhneFreigabeStehtNichtsOffen(t *testing.T) {
	t.Parallel()

	antwort := freigabeanwendung(t).PartieFreigabeStand(t.Context())

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %v", antwort.Fehler)
	case antwort.Laeuft:
		t.Error("es steht etwas offen")
	case antwort.Vergeben == nil:
		t.Error("die Liste ist nil statt leer")
	}
}

// **Eine freigegebene Partie steht offen — mit ihrem Wirt.**
func TestEineFreigegebenePartieStehtOffen(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)

	antwort := anwendung.PartieFreigeben(
		t.Context(), "p1", "anna", []string{"bert"})

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %v", antwort.Fehler)
	case !antwort.Laeuft:
		t.Error("sie steht nicht offen")
	case antwort.Partie != "p1":
		t.Errorf("Partie %q", antwort.Partie)
	case antwort.Wirt != "anna":
		t.Errorf("Wirt %q", antwort.Wirt)
	}
}

// **Und das Fenster erfährt es** — P9c (#2377): wer mit einer Einladung
// anfängt, gibt den Tisch frei, ohne dass sein Freigabefeld schon fragt; es
// zeigte sonst „zu“, bis jemand „Nachsehen“ drückt.
func TestEineFreigabeMeldetSichAmFenster(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)
	gemeldet := make(chan []any, 4)
	anwendung.melden = func(_ context.Context, ereignis string, daten ...any) {
		if ereignis == EreignisPartiefreigabe {
			gemeldet <- daten
		}
	}

	if antwort := anwendung.PartieFreigeben(t.Context(), "p1", "anna", []string{"bert"}); antwort.Fehler != nil {
		t.Fatalf("Fehler: %v", antwort.Fehler)
	}

	select {
	case daten := <-gemeldet:
		if len(daten) != 1 || daten[0] != "p1" {
			t.Errorf("gemeldet: %v", daten)
		}
	default:
		t.Error("das Fenster erfuhr nichts")
	}
}

// **Ohne Gastsitz gibt es nichts freizugeben.**
func TestOhneGastsitzGibtEsNichtsFreizugeben(t *testing.T) {
	t.Parallel()

	antwort := freigabeanwendung(t).PartieFreigeben(
		t.Context(), "p1", "anna", nil)

	if antwort.Fehler == nil || antwort.Laeuft {
		t.Errorf("es kam %+v", antwort)
	}
}

// **Eine Partie, die es nicht gibt, wird nicht freigegeben.**
//
// **Sonst bekäme der erste Gast eine Absage, die wie ein Netzfehler
// aussieht** (§8) — und gesucht würde am falschen Ende.
func TestEineUnbekanntePartieWirdNichtFreigegeben(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)
	anwendung.MitPartienaht(&klagendePartieleitung{
		weg: "Partie", klagt: errors.New("die Partie gibt es nicht"),
	})

	antwort := anwendung.PartieFreigeben(
		t.Context(), "gibt-es-nicht", "anna", []string{"bert"})

	if antwort.Fehler == nil || antwort.Laeuft {
		t.Errorf("es kam %+v", antwort)
	}
}

// **Das Beenden schließt sie wieder.**
func TestDasBeendenSchliesstDieFreigabe(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)
	anwendung.PartieFreigeben(t.Context(), "p1", "anna", []string{"bert"})

	antwort := anwendung.PartieFreigabeBeenden(t.Context())

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %v", antwort.Fehler)
	case antwort.Laeuft:
		t.Error("sie steht noch offen")
	}
}

// **Jeder der drei Wege fragt nach der Rolle — und der erste nach der
// richtigen.**
//
// **`externEmpfangen` und nicht `lokalSchreiben`**: von der Freigabe an ändert
// ein fremder Rechner den Stand dieser Partie. Dieselbe Klasse wie beim
// Horchposten (N7, #1568).
func TestJederFreigabewegFragtNachDerRolle(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		wirkung domain.Wirkung
		fahre   func(*App) PartiefreigabeAntwort
	}{
		"PartieFreigeben": {domain.WirkungExternEmpfangen, func(a *App) PartiefreigabeAntwort {
			return a.PartieFreigeben(t.Context(), "p1", "anna", []string{"bert"})
		}},
		"PartieFreigabeBeenden": {domain.WirkungSchreiben, func(a *App) PartiefreigabeAntwort {
			return a.PartieFreigabeBeenden(t.Context())
		}},
		"PartieFreigabeStand": {domain.WirkungLesen, func(a *App) PartiefreigabeAntwort {
			return a.PartieFreigabeStand(t.Context())
		}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wache := &zaehlendeWache{}
			anwendung := freigabeanwendung(t).MitWache(wache)

			antwort := fall.fahre(anwendung)
			if antwort.Fehler == nil {
				t.Fatal("der Weg kam durch")
			}

			if len(wache.gefragt) == 0 || wache.gefragt[0] != fall.wirkung {
				t.Errorf("gefragt wurde nach %v statt nach %q",
					wache.gefragt, fall.wirkung)
			}
		})
	}
}
