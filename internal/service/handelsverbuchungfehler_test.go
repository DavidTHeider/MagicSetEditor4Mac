package service_test

// **Was die Verbuchung meldet, wenn etwas dazwischenkommt** — Werkstück `T6`
// von Phase 5.6
// (#1471).
//
// # Warum das eine eigene Datei ist
//
// **Weil ein Fehlerweg, den niemand fährt, wie ein geprüfter aussieht** (§10).
// Ein Ausweis, der nicht kommt, ein Bestand, der klagt, ein Beleg, der sich
// nicht lesen lässt — an jeder dieser Stellen darf kein halber Stand
// zurückbleiben, und jede muss sagen, was los ist.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// errKeinAusweis ist, was ein Schlüsselbund sagt, der nicht antwortet.
var errKeinAusweis = errors.New("der Ausweis ist nicht zu haben")

// stummerAusweis gibt keinen.
type stummerAusweis struct{}

func (stummerAusweis) Ausweis(context.Context) (peer.Identitaet, error) {
	return peer.Identitaet{}, errKeinAusweis
}

// **Ohne Ausweis wird nicht verbucht** — und auch nicht geschritten.
//
// **Der Vorgang steht dabei da.** Der Ausweis kommt aus dem Schlüsselbund und
// wird erst geholt, wenn es überhaupt etwas zu verbuchen gibt; ein Lauf ohne
// Vorgang scheiterte an dem und nicht am Ausweis — und prüfte damit einen
// anderen Weg als den, den sein Kopf nennt (§10).
func TestOhneAusweisWirdNichtVerbucht(t *testing.T) {
	t.Parallel()

	stand := &handelsprobe{staende: map[string]domain.Handelsangebot{}}
	bestand := &verbuchungsprobe{
		stuecke: map[int64]domain.Exemplar{},
		belege:  map[string]domain.Belegstand{},
		stand:   stand,
	}

	stelleEin(t, stand, bestand, true)

	stelle := service.NeueVerbuchungsstelle(
		stand, bestand, stummerAusweis{}, bestand).MitWache(willigeWache{})

	if _, fehler := stelle.Verbuchen(t.Context(), derVorgang, derFreibrief); !errors.Is(
		fehler, errKeinAusweis) {
		t.Errorf("das Verbuchen gab %v", fehler)
	}

	// **Und das Abgleichen aus dem Zustand, aus dem es führt** — sonst
	// scheiterte es am Übergang und nicht am Ausweis.
	angebot := stand.staende[derVorgang]
	angebot.Zustand = domain.HandelLokalVerbucht
	stand.staende[derVorgang] = angebot

	if _, fehler := stelle.Abgleichen(t.Context(), derVorgang); !errors.Is(
		fehler, errKeinAusweis) {
		t.Errorf("das Abgleichen gab %v", fehler)
	}

	if len(bestand.gebucht) != 0 {
		t.Error("es wurde gebucht")
	}
}

// **Jeder Weg der Verbuchung reicht seinen Fehler durch.**
func TestJederWegDerVerbuchungReichtSeinenFehlerDurch(t *testing.T) {
	t.Parallel()

	klage := errors.New("der Bestand klagt")

	faelle := map[string]struct {
		verderben func(*handelsprobe, *verbuchungsprobe)
		fahre     func(*service.Verbuchungsstelle) error
	}{
		"den Vorgang gibt es nicht": {
			func(stand *handelsprobe, _ *verbuchungsprobe) {
				delete(stand.staende, derVorgang)
			},
			func(s *service.Verbuchungsstelle) error {
				_, fehler := s.Verbuchen(t.Context(), derVorgang, derFreibrief)

				return fehler
			},
		},
		"der Beleg lässt sich nicht lesen": {
			func(_ *handelsprobe, bestand *verbuchungsprobe) {
				bestand.belegklage = klage
			},
			func(s *service.Verbuchungsstelle) error {
				_, fehler := s.Verbuchen(t.Context(), derVorgang, derFreibrief)

				return fehler
			},
		},
		"die Zusagen lassen sich nicht lesen": {
			func(_ *handelsprobe, bestand *verbuchungsprobe) {
				bestand.zusagenklage = klage
			},
			func(s *service.Verbuchungsstelle) error {
				_, fehler := s.Verbuchen(t.Context(), derVorgang, derFreibrief)

				return fehler
			},
		},
		"die Besitzzeile gibt es nicht": {
			func(_ *handelsprobe, bestand *verbuchungsprobe) {
				delete(bestand.stuecke, 7)
			},
			func(s *service.Verbuchungsstelle) error {
				_, fehler := s.Verbuchen(t.Context(), derVorgang, derFreibrief)

				return fehler
			},
		},
		"der Abschluss scheitert": {
			func(_ *handelsprobe, bestand *verbuchungsprobe) {
				bestand.klagt = klage
			},
			func(s *service.Verbuchungsstelle) error {
				_, fehler := s.Verbuchen(t.Context(), derVorgang, derFreibrief)

				return fehler
			},
		},
		"der Vorgang steht woanders": {
			func(stand *handelsprobe, _ *verbuchungsprobe) {
				angebot := stand.staende[derVorgang]
				angebot.Zustand = domain.HandelEntwurf
				stand.staende[derVorgang] = angebot
			},
			func(s *service.Verbuchungsstelle) error {
				_, fehler := s.Verbuchen(t.Context(), derVorgang, derFreibrief)

				return fehler
			},
		},
		// **Eine Bestätigung von hier aus gibt es nicht** — der Vorgang
		// steht noch bei der Übergabe.
		"die Bestätigung geht von hier nicht": {
			func(*handelsprobe, *verbuchungsprobe) {},
			func(s *service.Verbuchungsstelle) error {
				_, fehler := s.Bestaetigen(
					t.Context(), derVorgang, "bgegenueber", derFreibrief)

				return fehler
			},
		},
		"der Stand lässt sich nicht lesen": {
			func(stand *handelsprobe, _ *verbuchungsprobe) {
				stand.klagt = klage
			},
			func(s *service.Verbuchungsstelle) error {
				_, fehler := s.Abgleichen(t.Context(), derVorgang)

				return fehler
			},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stelle, stand, bestand := eineVerbuchung(t)
			stelleEin(t, stand, bestand, true)
			fall.verderben(stand, bestand)

			if fehler := fall.fahre(stelle); fehler == nil {
				t.Error("der Fehler kam nicht an")
			}
		})
	}
}

// **Und der Beleg wird durchgereicht, wie er dasteht.**
func TestDerBelegKommtDurch(t *testing.T) {
	t.Parallel()

	stelle, stand, bestand := eineVerbuchung(t)
	stelleEin(t, stand, bestand, true)

	// **Vor dem Verbuchen steht keiner da** — und das ist kein Fehler.
	vorher, fehler := stelle.Beleg(t.Context(), derVorgang)
	if fehler != nil {
		t.Fatalf("Beleg: %v", fehler)
	}

	if vorher != nil {
		t.Errorf("es stand schon ein Beleg da: %+v", vorher)
	}

	bestand.belegklage = errors.New("der Beleg klagt")

	if _, fehler := stelle.Beleg(t.Context(), derVorgang); fehler == nil {
		t.Error("der Fehler kam nicht an")
	}
}

// **Ein Schritt, der sich nicht ablegen lässt, gilt nicht.**
//
// Der Bestand bewegt sich beim Abgleichen nicht — der Vorgang schon. Ein
// Schritt, der nur zurückgegeben und nicht abgelegt würde, wäre nach dem
// nächsten Start nicht geschehen, und der Aufrufer hielte ihn für getan.
func TestEinNichtAblegbarerSchrittGiltNicht(t *testing.T) {
	t.Parallel()

	stelle, stand, bestand := eineVerbuchung(t)
	stelleEin(t, stand, bestand, true)

	if _, fehler := stelle.Verbuchen(t.Context(), derVorgang, derFreibrief); fehler != nil {
		t.Fatalf("verbuchen: %v", fehler)
	}

	// **Erst danach klagt das Ablegen** — sonst scheiterte schon der Schritt
	// davor, und dieser Zweig bliebe unbetreten (§10).
	stand.schreibklage = errors.New("der Vorgang lässt sich nicht ablegen")

	if _, fehler := stelle.Abgleichen(t.Context(), derVorgang); fehler == nil {
		t.Error("der Fehler kam nicht an")
	}
}
