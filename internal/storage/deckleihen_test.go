package storage_test

// **Die digitale Deckleihe im Bestand** — Werkstück G6 von Phase 5.5
// (#1463).

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// einLeihstand gibt einen Stand mit den genannten Abweichungen.
func einLeihstand(aendere ...func(*domain.Leihstand)) domain.Leihstand {
	stand := domain.Leihstand{
		Leihe: domain.Deckleihe{
			Kennung:   "l1",
			Verleiher: "12D3KooWAnna",
			Entleiher: "12D3KooWBert",
			Abdruck:   "abdruckdesabzugs",
			Bindung: domain.Leihbindung{
				Art: domain.LeiheFuerPartie, Woran: "p1",
			},
			Rechte:       []domain.Leihrecht{domain.RechtSpielen},
			Ausgestellt:  time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC),
			Bis:          time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC),
			Modus:        domain.LaufZuEnde,
			Unterschrift: "abcdef",
		},
		Abzug: domain.Partiedeck{
			Abdruck: "abdruckdesabzugs",
			Karten: []domain.Abzugskarte{{
				Kennung: "k1", Name: "Wald", Menge: 40,
				Seiten: []domain.Abzugsseite{{
					Name: "Wald", Felder: map[string]string{"name": "Wald"},
				}},
			}},
			Beigaben: []domain.Beigabe{},
		},
	}

	for _, aenderung := range aendere {
		aenderung(&stand)
	}

	return stand
}

func TestEineLeiheKommtZurueckWieSieHineinging(t *testing.T) {
	t.Parallel()

	ablage, ctx := storage.NeueLeihablage(frischeDatenbank(t)), probekontext(t)

	if fehler := ablage.LeiheSchreiben(ctx, einLeihstand()); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	zurueck, fehler := ablage.Leihe(ctx, "l1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	switch {
	case zurueck.Leihe.Unterschrift != "abcdef":
		t.Errorf("die Unterschrift ist %q", zurueck.Leihe.Unterschrift)
	case !zurueck.Leihe.Ausgestellt.Equal(einLeihstand().Leihe.Ausgestellt):
		t.Errorf("die Ausstellung ist %v", zurueck.Leihe.Ausgestellt)
	case len(zurueck.Abzug.Karten) != 1:
		t.Errorf("der Abzug hat %d Karten", len(zurueck.Abzug.Karten))
	case zurueck.Abzug.Karten[0].Seiten[0].Felder["name"] != "Wald":
		t.Errorf("die Feldwerte sind %v",
			zurueck.Abzug.Karten[0].Seiten[0].Felder)
	}
}

// **Ein zweiter Schreibvorgang ändert den Stand und nicht das Dokument.**
func TestEinZweiterSchreibvorgangAendertNurDenStand(t *testing.T) {
	t.Parallel()

	ablage, ctx := storage.NeueLeihablage(frischeDatenbank(t)), probekontext(t)

	if fehler := ablage.LeiheSchreiben(ctx, einLeihstand()); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	geaendert := einLeihstand(func(s *domain.Leihstand) {
		s.Verbraucht, s.Widerrufen = "p1", true
		s.Leihe.Unterschrift = "eineandere"
	})

	if fehler := ablage.LeiheSchreiben(ctx, geaendert); fehler != nil {
		t.Fatalf("fortschreiben: %v", fehler)
	}

	zurueck, fehler := ablage.Leihe(ctx, "l1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	switch {
	case zurueck.Verbraucht != "p1" || !zurueck.Widerrufen:
		t.Errorf("der Stand ist %+v", zurueck)
	case zurueck.Leihe.Unterschrift != "abcdef":
		t.Errorf("das unterschriebene Dokument wurde überschrieben: %q",
			zurueck.Leihe.Unterschrift)
	}
}

// **Eine fremde Leihe gehört einem anderen Profil** — und ist hier keine.
func TestEineLeiheEinesFremdenProfilsIstHierKeine(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueLeihablage(db)

	if fehler := ablage.LeiheSchreiben(
		probekontext(t), einLeihstand()); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	fremd := storage.MitProfil(t.Context(),
		zweitesProfil(t, storage.NeueNutzerablage(db), "Turnier"))

	if _, fehler := ablage.Leihe(fremd, "l1"); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Fatalf("erwartet ErrNotFound, bekam %v", fehler)
	}

	leihen, fehler := ablage.Leihen(fremd)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(leihen) != 0 {
		t.Fatalf("das fremde Profil sieht %d Leihen", len(leihen))
	}
}

func TestOhneLeiheIstDieListeLeerUndKeinNil(t *testing.T) {
	t.Parallel()

	ablage, ctx := storage.NeueLeihablage(frischeDatenbank(t)), probekontext(t)

	leihen, fehler := ablage.Leihen(ctx)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if leihen == nil || len(leihen) != 0 {
		t.Fatalf("die Liste ist %v", leihen)
	}
}

func TestEineUnbekannteLeiheWirdGemeldet(t *testing.T) {
	t.Parallel()

	ablage, ctx := storage.NeueLeihablage(frischeDatenbank(t)), probekontext(t)

	if _, fehler := ablage.Leihe(ctx, "gibtesnicht"); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Fatalf("erwartet ErrNotFound, bekam %v", fehler)
	}
}
