package service_test

// **Das Ergebnis kommt aus der Partie und nicht aus einem Eingabefeld** —
// Werkstück M17 der Regelmaschine
// (#1618).
//
// # Wogegen hier geprüft wird
//
//   - **Die dritte Abnahme des Issues**: „Das Ergebnis erreicht die
//     Turnierbindung aus G7, ohne dass jemand es eintippt."
//   - **Eine laufende Partie hat kein Ergebnis** — und wird abgewiesen, statt
//     null zu null zu melden.
//   - **Die Seiten kommen aus der Paarung** (A und B) und nicht aus der
//     Reihenfolge der Sitze: [service.Sitzzuordnung] gibt sie so, wie der
//     Aufrufer sie nennt.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einEntschiedenerMatchtisch gibt eine gebundene Paarung mit Ausgang.
func einEntschiedenerMatchtisch(t *testing.T, sieger string, remis bool) (
	*service.Matchleitung, *matchprobe,
) {
	t.Helper()

	leitung, _, match, _ := einEntschiedenerTischSo(t, sieger, remis)

	return leitung, match
}

// einEntschiedenerTisch gibt ihn samt allen drei Doppeln.
func einEntschiedenerTisch(t *testing.T) (
	*service.Matchleitung, *turnierdoppel, *matchprobe,
	*partieprobeFuerMatch,
) {
	t.Helper()

	return einEntschiedenerTischSo(t, "anna", false)
}

// einEntschiedenerTischSo baut ihn mit einem Ausgang.
func einEntschiedenerTischSo(t *testing.T, sieger string, remis bool) (
	*service.Matchleitung, *turnierdoppel, *matchprobe,
	*partieprobeFuerMatch,
) {
	t.Helper()

	leitung, turniere, match, partien := einMatchtischMitProben(t,
		func(p *domain.Partie) {
			p.Stand = domain.StandBeendet
			p.Ausgang = domain.Partieausgang{
				Sieger: sieger, Unentschieden: remis,
				Ausgeschieden: []domain.Ausgeschieden{
					{Sitz: "bert", Grund: domain.GrundLeben},
				},
			}
		})

	if _, fehler := leitung.Binden(
		context.Background(), 1, 7, "p1", zweiSitze()); fehler != nil {
		t.Fatalf("binden: %v", fehler)
	}

	return leitung, turniere, match, partien
}

// **Wer gewinnt, steht danach in der Meldung** — ohne dass jemand tippt.
func TestDasErgebnisKommtAusDerPartie(t *testing.T) {
	t.Parallel()

	leitung, match := einEntschiedenerMatchtisch(t, "anna", false)

	if fehler := leitung.AusDerPartie(
		context.Background(), 7, "anna"); fehler != nil {
		t.Fatalf("melden: %v", fehler)
	}

	gemeldet := match.meldungen[7]

	switch {
	case len(gemeldet) != 1:
		t.Fatalf("%d Meldungen", len(gemeldet))
	case gemeldet[0].Ergebnis != (domain.Ergebnis{SpieleA: 1}):
		t.Errorf("Ergebnis = %+v", gemeldet[0].Ergebnis)
	case gemeldet[0].Kette != "abdruck-7":
		t.Errorf("Kette = %q", gemeldet[0].Kette)
	case gemeldet[0].Von != "anna":
		t.Errorf("Von = %q", gemeldet[0].Von)
	}
}

// **Die zweite Seite ist die zweite Seite** — der Sieger sitzt bei B.
func TestDerSiegDerZweitenSeiteZaehltFuerB(t *testing.T) {
	t.Parallel()

	leitung, match := einEntschiedenerMatchtisch(t, "bert", false)

	if fehler := leitung.AusDerPartie(
		context.Background(), 7, "bert"); fehler != nil {
		t.Fatalf("melden: %v", fehler)
	}

	if got := match.meldungen[7][0].Ergebnis; got != (domain.Ergebnis{SpieleB: 1}) {
		t.Fatalf("Ergebnis = %+v", got)
	}
}

// **Unentschieden ist ein eigener Ausgang** — CR 104.4a, die zweite Abnahme.
func TestEinUnentschiedenWirdAlsSolchesGemeldet(t *testing.T) {
	t.Parallel()

	leitung, match := einEntschiedenerMatchtisch(t, "", true)

	if fehler := leitung.AusDerPartie(
		context.Background(), 7, "anna"); fehler != nil {
		t.Fatalf("melden: %v", fehler)
	}

	erwartet := domain.Ergebnis{Unentschieden: 1}
	if got := match.meldungen[7][0].Ergebnis; got != erwartet {
		t.Fatalf("Ergebnis = %+v", got)
	}
}

// **Eine laufende Partie hat kein Ergebnis.**
func TestEineLaufendePartieWirdNichtGemeldet(t *testing.T) {
	t.Parallel()

	leitung, _, _ := einMatchtisch(t)

	if _, fehler := leitung.Binden(
		context.Background(), 1, 7, "p1", zweiSitze()); fehler != nil {
		t.Fatalf("binden: %v", fehler)
	}

	fehler := leitung.AusDerPartie(context.Background(), 7, "anna")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("eine laufende Partie wurde gemeldet: %v", fehler)
	}
}

// **Ein Sieger, der nicht gebunden ist, ist ein Befund** (§8).
//
// Er entsteht, wenn jemand die Bindung auf eine andere Partie zeigen lässt —
// dann gewinnt ein Sitz, den diese Paarung nicht kennt.
func TestEinUngebundenerSiegerWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	leitung, _ := einEntschiedenerMatchtisch(t, "zuschauer", false)

	fehler := leitung.AusDerPartie(context.Background(), 7, "anna")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("ein fremder Sieger ging durch: %v", fehler)
	}
}

// **Eine Partie ohne Sieger und ohne Unentschieden ist nicht entschieden.**
//
// Sie entsteht, wenn der Stand von Hand auf „beendet" gesetzt wurde — der
// Ausgang steht dann leer da.
func TestEinLeererAusgangWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	leitung, _ := einEntschiedenerMatchtisch(t, "", false)

	fehler := leitung.AusDerPartie(context.Background(), 7, "anna")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("ein leerer Ausgang ging durch: %v", fehler)
	}
}

// **Jeder Fehlerweg der Meldung wird betreten** — §3: ein Zweig ohne Lauf ist
// eine Zusage ohne Prüfung.
//
// **Über Doppel, weil der echte Weg sie nicht hergibt**: wer eine Bindung
// lesen kann, liest gewöhnlich auch die Partie dazu, ihr Turnier und ihre
// Kette. Die Zweige stehen trotzdem da, und ein von Hand geänderter Bestand
// erreicht sie.
func TestJederFehlerwegDerMeldungWirdBetreten(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*matchprobe, *turnierdoppel, *partieprobeFuerMatch){
		"die Bindung klagt": func(
			m *matchprobe, _ *turnierdoppel, _ *partieprobeFuerMatch,
		) {
			m.klagt = domain.MitGrund(domain.ErrConflict, nil, "die Probe klagt")
		},

		"die Partie gibt es nicht": func(
			m *matchprobe, _ *turnierdoppel, _ *partieprobeFuerMatch,
		) {
			bindung := m.bindungen[7]
			bindung.PartieID = "gibtesnicht"
			m.bindungen[7] = bindung
		},

		"das Turnier klagt": func(
			_ *matchprobe, t *turnierdoppel, _ *partieprobeFuerMatch,
		) {
			t.fehler = domain.MitGrund(domain.ErrConflict, nil, "die Probe klagt")
		},

		"die Paarung gibt es nicht mehr": func(
			_ *matchprobe, t *turnierdoppel, _ *partieprobeFuerMatch,
		) {
			t.turniere[1].Runden[0].Paarungen = nil
		},

		"die Kette klagt": func(
			_ *matchprobe, _ *turnierdoppel, p *partieprobeFuerMatch,
		) {
			p.kettenklage = domain.MitGrund(
				domain.ErrConflict, nil, "die Probe klagt")
		},
	}

	for name, brich := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			leitung, turniere, match, partien := einEntschiedenerTisch(t)

			brich(match, turniere, partien)

			if fehler := leitung.AusDerPartie(
				context.Background(), 7, "anna"); fehler == nil {
				t.Fatal("der Fehler kam nicht an")
			}
		})
	}
}
