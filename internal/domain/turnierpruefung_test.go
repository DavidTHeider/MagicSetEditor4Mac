package domain_test

// **Was ein Turnier und ein Ergebnis abweisen** — Werkstück 2 von Phase 4.7
// (#1294).
//
// **Zum Ergebnis gehört auch das Nein.** Eine Prüfung, von der nur der
// Erfolgsweg belegt ist, hat die Hälfte ausgelassen (§10).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Turnier braucht Namen und Spiel.**
func TestEinTurnierWeistUnvollstaendigesAb(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		turnier    domain.Turnier
		feld       string
		schluessel domain.Schluessel
	}{
		"ohne Namen": {
			turnier:    domain.Turnier{Spiel: "magic"},
			feld:       "name",
			schluessel: domain.SchluesselTurniernameLeer,
		},
		"nur Leerzeichen als Name": {
			turnier:    domain.Turnier{Name: "   ", Spiel: "magic"},
			feld:       "name",
			schluessel: domain.SchluesselTurniernameLeer,
		},
		"ohne Spiel": {
			turnier:    domain.Turnier{Name: "Küchentisch"},
			feld:       "spiel",
			schluessel: domain.SchluesselTurnierspielLeer,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := fall.turnier.Pruefe()
			if err == nil {
				t.Fatal("angenommen, erwartet abgewiesen")
			}

			if !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Fehler ist %v, erwartet eine Validierung", err)
			}

			var feld *domain.FieldError
			if !errors.As(err, &feld) {
				t.Fatalf("Fehler nennt kein Feld: %v", err)
			}

			if feld.Field != fall.feld || feld.Key != fall.schluessel {
				t.Errorf("Feld %q, Schlüssel %q — erwartet %q und %q",
					feld.Field, feld.Key, fall.feld, fall.schluessel)
			}
		})
	}
}

// **Ein vollständiges Turnier geht durch.**
func TestEinVollstaendigesTurnierGehtDurch(t *testing.T) {
	t.Parallel()

	turnier := domain.Turnier{Name: "Küchentisch", Spiel: "magic"}

	if err := turnier.Pruefe(); err != nil {
		t.Errorf("abgewiesen: %v", err)
	}
}

// **Ein Ergebnis ohne ein einziges Spiel ist keines.**
func TestEinErgebnisWeistUnmoeglichesAb(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		ergebnis   domain.Ergebnis
		hoechstens int
	}{
		"kein einziges Spiel":   {domain.Ergebnis{}, 0},
		"negative Spiele für A": {domain.Ergebnis{SpieleA: -1, SpieleB: 2}, 0},
		"negative Spiele für B": {domain.Ergebnis{SpieleA: 2, SpieleB: -1}, 0},
		"negative Unentschiedene": {
			domain.Ergebnis{SpieleA: 2, Unentschieden: -1}, 0,
		},
		"mehr Spiele als erlaubt": {
			domain.Ergebnis{SpieleA: 2, SpieleB: 2}, 3,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := fall.ergebnis.Pruefe(fall.hoechstens); err == nil {
				t.Error("angenommen, erwartet abgewiesen")
			} else if !errors.Is(err, domain.ErrValidation) {
				t.Errorf("Fehler ist %v, erwartet eine Validierung", err)
			}
		})
	}
}

// **Und was gültig ist, geht durch — auch ohne Obergrenze.**
func TestEinGueltigesErgebnisGehtDurch(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		ergebnis   domain.Ergebnis
		hoechstens int
	}{
		"zwei zu null, ohne Grenze": {domain.Ergebnis{SpieleA: 2}, 0},
		"zwei zu eins, Grenze drei": {domain.Ergebnis{SpieleA: 2, SpieleB: 1}, 3},
		"ein Unentschieden":         {domain.Ergebnis{Unentschieden: 1}, 3},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := fall.ergebnis.Pruefe(fall.hoechstens); err != nil {
				t.Errorf("abgewiesen: %v", err)
			}
		})
	}
}

// **Eine Paarung mit einem unbekannten Teilnehmer wird übergangen.**
//
// Sie ist ein Fehler im Bestand. Sie hier stillschweigend zu einem Rang zu
// machen, verdeckte ihn — und die Rangliste hätte plötzlich einen Teilnehmer
// mehr, als angemeldet ist.
func TestEinUnbekannterTeilnehmerBekommtKeinenRang(t *testing.T) {
	t.Parallel()

	turnier := turnierMit(2)

	runde(&turnier, spiel(1, 99, 2, 0))

	rangliste := domain.Rangliste(turnier)

	if len(rangliste) != 2 {
		t.Fatalf("%d Ränge, erwartet 2", len(rangliste))
	}

	// **Und der Bekannte bekommt seinen Sieg trotzdem.** Was gezählt werden
	// kann, wird gezählt (§8).
	if rangliste[0].TeilnehmerID != 1 || rangliste[0].Matchpunkte != 3 {
		t.Errorf("Platz 1 ist %+v", rangliste[0])
	}
}
