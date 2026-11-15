package service_test

// **Was eine Einladung an einem bekannten Gegenüber tut** — NET-01.AK-03
// (#1705).

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

func TestEinladungEinesGesperrten(t *testing.T) {
	t.Parallel()

	anna, _, annaWerk, bertWerk, ablage := einladungsseiten(t)
	ablage.kontakte[string(anna.Kennung())] = domain.Peerkontakt{
		Kennung: string(anna.Kennung()), Stand: domain.KontaktGesperrt,
	}

	code := annasCode(t, annaWerk)

	vorschau, fehler := bertWerk.Pruefen(t.Context(), code)
	if fehler != nil || vorschau.Lage != domain.LageWiderrufen {
		t.Fatalf("Vorschau %+v, %v", vorschau, fehler)
	}

	if _, fehler := bertWerk.Annehmen(t.Context(),
		service.Einladungsannahme{Code: code}); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("Annehmen: %v", fehler)
	}

	if len(ablage.nonces) != 0 {
		t.Fatal("die Nonce ist verbraucht")
	}
}

func TestEinladungMitNeuemSchluessel(t *testing.T) {
	t.Parallel()

	alt, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatal(fehler)
	}

	_, bert, _, _, ablage := einladungsseiten(t)
	ablage.kontakte[string(alt.Kennung())] = domain.Peerkontakt{
		Kennung: string(alt.Kennung()), Stand: domain.KontaktBestaetigt,
		EigenerName: "Anna", Rechte: []domain.Zweck{domain.ZweckHandel},
	}

	neu, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatal(fehler)
	}

	uhr := func() time.Time { return einladungsuhr }
	annaNeu := service.NeuesEinladungswerk(festerAusweis{ich: neu},
		festeRotation{vorher: alt.Kennung()}, neueEinladungsablage(), uhr)
	bertWerk := service.NeuesEinladungswerk(festerAusweis{ich: bert},
		festeRotation{}, ablage, uhr)

	code := annasCode(t, annaNeu)

	vorschau, fehler := bertWerk.Pruefen(t.Context(), code)
	if fehler != nil || vorschau.Vorgaenger == nil ||
		vorschau.Vorgaenger.EigenerName != "Anna" {
		t.Fatalf("Vorschau %+v, %v", vorschau, fehler)
	}

	kontakt, fehler := bertWerk.Annehmen(t.Context(), service.Einladungsannahme{Code: code})
	if fehler != nil {
		t.Fatal(fehler)
	}

	// **Kein Erbe** — AK-03: der neue Schlüssel übernimmt nichts, bis
	// jemand neu verglichen hat.
	if kontakt.Lage() != domain.LageSchluesselGeaendert || len(kontakt.Rechte) != 0 ||
		kontakt.Vorgaenger != string(alt.Kennung()) {
		t.Fatalf("Kontakt: %+v", kontakt)
	}
}

func TestEinladungMitUnbekanntemVorgaenger(t *testing.T) {
	t.Parallel()

	_, _, _, bertWerk, _ := einladungsseiten(t)

	fremd, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatal(fehler)
	}

	neu, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatal(fehler)
	}

	werk := service.NeuesEinladungswerk(festerAusweis{ich: neu},
		festeRotation{vorher: fremd.Kennung()}, neueEinladungsablage(),
		func() time.Time { return einladungsuhr })

	// **Eine Behauptung über jemanden, den ich nicht kenne, sagt nichts** —
	// der Kontakt ist einfach unbestätigt.
	kontakt, fehler := bertWerk.Annehmen(t.Context(),
		service.Einladungsannahme{Code: annasCode(t, werk)})
	if fehler != nil || kontakt.Vorgaenger != "" || kontakt.Lage() != domain.LageUnbestaetigt {
		t.Fatalf("Kontakt %+v, %v", kontakt, fehler)
	}
}

func TestEinladungNurZumKennenlernen(t *testing.T) {
	t.Parallel()

	anna, bert, annaWerk, bertWerk, _ := einladungsseiten(t)

	ausgestellt, fehler := annaWerk.Erstellen(t.Context(), service.Einladungsbitte{
		Name: "Anna", Zweck: domain.ZweckPaarung,
	})
	if fehler != nil {
		t.Fatal(fehler)
	}

	kontakt, fehler := bertWerk.Annehmen(t.Context(), service.Einladungsannahme{
		Code:       ausgestellt.Code,
		Verglichen: domain.Vergleichscode(string(anna.Kennung()), string(bert.Kennung())),
		Freigeben:  true,
	})
	if fehler != nil || kontakt.Stand != domain.KontaktBestaetigt || len(kontakt.Rechte) != 0 {
		t.Fatalf("Kontakt %+v, %v", kontakt, fehler)
	}
}

// kaputterAusweis gibt keine Identität.
type kaputterAusweis struct{}

func (kaputterAusweis) Ausweis(context.Context) (peer.Identitaet, error) {
	return peer.Identitaet{}, errors.New("kein Schlüssel")
}

// kaputteRotation gibt keine Lage.
type kaputteRotation struct{}

func (kaputteRotation) Lage(context.Context) (service.Peerlage, error) {
	return service.Peerlage{}, errors.New("kein Vermerk")
}

// leererZufall hat nichts.
type leererZufall struct{}

func (leererZufall) Read([]byte) (int, error) { return 0, errors.New("leer") }

func TestEinladungFehlerwege(t *testing.T) {
	t.Parallel()

	anna, _, annaWerk, _, _ := einladungsseiten(t)
	code := annasCode(t, annaWerk)
	uhr := func() time.Time { return einladungsuhr }
	bitte := service.Einladungsbitte{Name: "Anna", Zweck: domain.ZweckHandel}

	t.Run("ohne Ausweis wird nichts ausgestellt", func(t *testing.T) {
		t.Parallel()

		werk := service.NeuesEinladungswerk(kaputterAusweis{}, festeRotation{},
			neueEinladungsablage(), uhr)
		if _, fehler := werk.Erstellen(t.Context(), bitte); fehler == nil {
			t.Fatal("ausgestellt ohne Schlüssel")
		}

		if _, fehler := werk.Pruefen(t.Context(), code); fehler == nil {
			t.Fatal("geprüft ohne Schlüssel")
		}
	})

	t.Run("ohne Vermerk wird nichts ausgestellt", func(t *testing.T) {
		t.Parallel()

		werk := service.NeuesEinladungswerk(festerAusweis{ich: anna},
			kaputteRotation{}, neueEinladungsablage(), uhr)
		if _, fehler := werk.Erstellen(t.Context(), bitte); fehler == nil {
			t.Fatal("ausgestellt ohne Vermerk")
		}
	})

	t.Run("ohne Zufall wird nichts ausgestellt", func(t *testing.T) {
		t.Parallel()

		werk := service.NeuesEinladungswerk(festerAusweis{ich: anna},
			festeRotation{}, neueEinladungsablage(), uhr).MitZufall(leererZufall{})
		if _, fehler := werk.Erstellen(t.Context(), bitte); fehler == nil {
			t.Fatal("ausgestellt ohne Nonce")
		}
	})

	t.Run("ein leerer Name wird nicht ausgestellt", func(t *testing.T) {
		t.Parallel()

		if _, fehler := annaWerk.Erstellen(t.Context(), service.Einladungsbitte{
			Name: " ", Zweck: domain.ZweckHandel,
		}); !errors.Is(fehler, domain.ErrValidation) {
			t.Fatalf("Erstellen: %v", fehler)
		}
	})

	t.Run("kein Code ist keine Einladung", func(t *testing.T) {
		t.Parallel()

		if _, fehler := annaWerk.Pruefen(t.Context(), "hallo"); !errors.Is(fehler, domain.ErrValidation) {
			t.Fatalf("Pruefen: %v", fehler)
		}
	})

	t.Run("die Fehler hinter dem ersten Lesen gehen durch", func(t *testing.T) {
		t.Parallel()

		_, bert, _, _, _ := einladungsseiten(t)
		vorher := service.NeuesEinladungswerk(festerAusweis{ich: anna},
			festeRotation{vorher: "12D3KooWVorgaenger"}, neueEinladungsablage(), uhr)
		mitVorgaenger := annasCode(t, vorher)

		for name, kaputt := range map[string]func(*einladungsablage){
			"Vorgänger": func(a *einladungsablage) { a.vorgaengerfehler = errors.New("weg") },
			"Nonce":     func(a *einladungsablage) { a.noncefehler = errors.New("weg") },
			"Aufnahme":  func(a *einladungsablage) { a.aufnahmefehler = errors.New("weg") },
		} {
			ablage := neueEinladungsablage()
			kaputt(ablage)

			werk := service.NeuesEinladungswerk(festerAusweis{ich: bert},
				festeRotation{}, ablage, uhr)
			if _, fehler := werk.Annehmen(t.Context(),
				service.Einladungsannahme{Code: mitVorgaenger}); fehler == nil {
				t.Errorf("%s: angenommen trotz Fehler", name)
			}
		}
	})

	t.Run("ein Lesefehler im Bestand geht durch", func(t *testing.T) {
		t.Parallel()

		_, bert, _, _, _ := einladungsseiten(t)
		ablage := neueEinladungsablage()
		ablage.fehler = errors.New("Platte voll")

		werk := service.NeuesEinladungswerk(festerAusweis{ich: bert},
			festeRotation{}, ablage, uhr)
		if _, fehler := werk.Annehmen(t.Context(),
			service.Einladungsannahme{Code: code}); fehler == nil {
			t.Fatal("angenommen trotz Lesefehler")
		}
	})
}
