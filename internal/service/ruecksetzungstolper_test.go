package service_test

// **Wo eine Zählung scheitert** — Werkstück O4 von Phase 5.2
// (#1428).
//
// **Eine eigene Datei**, weil `ruecksetzung_test.go` mit ihr über die
// Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort steht, was die
// Rücksetzung **tut**, hier, was geschieht, wenn der Bestand **nicht
// antwortet**.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// stolperablage scheitert nur an einer Stelle.
//
// **Eine eigene und nicht das Doppel mit einem Fehlerfeld**: jenes scheitert
// überall gleichzeitig, und dann kommt der Weg nie bis zu der Stelle, um die
// es geht (§10).
type stolperablage struct {
	*ruecksetzungsdoppel

	beiRueckkehren bool
	beiNutzerzahl  bool
	beiUmfang      bool
	beiZugaengen   bool
}

func (s *stolperablage) FremdeZugaenge(ctx context.Context, nutzerID int64) (
	domain.Betroffenes, error,
) {
	if s.beiZugaengen {
		return domain.Betroffenes{}, errors.New("die Zugänge ließen sich nicht zählen")
	}

	return s.ruecksetzungsdoppel.FremdeZugaenge(ctx, nutzerID)
}

func (s *stolperablage) Profilumfang(ctx context.Context, profilID int64) (
	[]domain.Betroffenes, error,
) {
	if s.beiUmfang {
		return nil, errors.New("der Umfang ließ sich nicht zählen")
	}

	return s.ruecksetzungsdoppel.Profilumfang(ctx, profilID)
}

func (s *stolperablage) OffeneRueckkehrmengen(ctx context.Context, profilID int64) (int, error) {
	if s.beiRueckkehren {
		return 0, errors.New("die Rückkehren ließen sich nicht zählen")
	}

	return s.ruecksetzungsdoppel.OffeneRueckkehrmengen(ctx, profilID)
}

func (s *stolperablage) Nutzerzahl(ctx context.Context) (int, error) {
	if s.beiNutzerzahl {
		return 0, errors.New("die Nutzer ließen sich nicht zählen")
	}

	return s.ruecksetzungsdoppel.Nutzerzahl(ctx)
}

// **Jede Zählung der Vorschau reicht ihren Fehler durch** — eine Vorschau
// ohne Hindernis, weil das Zählen scheiterte, wäre die gefährlichste Antwort.
func TestJedeZaehlungDerVorschauReichtIhrenFehlerDurch(t *testing.T) {
	t.Parallel()

	for name, ablage := range map[string]*stolperablage{
		"die offenen Rückkehren": {
			ruecksetzungsdoppel: &ruecksetzungsdoppel{
				nutzer: 2, profile: []domain.Profil{{ID: 3, Name: "Meins"}},
			},
			beiRueckkehren: true,
		},
		"die Nutzerzahl": {
			ruecksetzungsdoppel: &ruecksetzungsdoppel{
				nutzer: 2, profile: []domain.Profil{{ID: 3, Name: "Meins"}},
			},
			beiNutzerzahl: true,
		},
		"der Umfang eines Profils": {
			ruecksetzungsdoppel: &ruecksetzungsdoppel{
				nutzer: 2, profile: []domain.Profil{{ID: 3, Name: "Meins"}},
			},
			beiUmfang: true,
		},
		// **Die fremden Zugänge** (#2287).
		"die Zugänge an fremden Profilen": {
			ruecksetzungsdoppel: &ruecksetzungsdoppel{
				nutzer: 2, profile: []domain.Profil{{ID: 3, Name: "Meins"}},
			},
			beiZugaengen: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			weg := service.NeueRuecksetzung(ablage, ablage.ruecksetzungsdoppel,
				wache(), ablage.neuanfang)

			if _, fehler := weg.Vorschau(t.Context(), domain.RuecksetzungNutzer,
				service.Ruecksetzungsziel{}); fehler == nil {
				t.Error("der Fehler kam nicht durch")
			}
		})
	}
}

// **Ein freies Profil wird gelöscht — mit seinem Namen als Bindung.**
func TestEinFreiesProfilWirdMitSeinemNamenGeloescht(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{profile: []domain.Profil{{ID: 3, Name: "Meins"}}}

	dienst := ruecksetzung(doppel)
	ziel := service.Ruecksetzungsziel{ProfilID: 3, Name: "Meins"}

	bilanz, fehler := dienst.Zuruecksetzen(
		t.Context(), domain.RuecksetzungProfil, ziel, "",
		gesehen(t, dienst, domain.RuecksetzungProfil, ziel))
	if fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	if doppel.geloescht != "profil:Meins" {
		t.Errorf("gelöscht wurde %q", doppel.geloescht)
	}

	if bilanz.Art != domain.RuecksetzungProfil {
		t.Errorf("die Bilanz nennt %q", bilanz.Art)
	}
}
