package service_test

// **Rotation, Übergangsfrist, letzte Nutzung und Sammelwiderruf** —
// Werkstück O2 von Phase 5.2
// (#1426).

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// rotationsuhr ist eine Uhr, die sich stellen lässt.
type rotationsuhr struct{ jetzt time.Time }

func (u *rotationsuhr) lesen() time.Time { return u.jetzt }

func rotationsverwaltung(ablage *tokendoppel) (*service.Tokenverwaltung, *rotationsuhr) {
	uhr := &rotationsuhr{jetzt: time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)}

	return service.NeueTokenverwaltung(ablage).MitUhr(uhr.lesen), uhr
}

func rotationstoken() domain.Zugriffstoken {
	return domain.Zugriffstoken{
		ProfilID: 1, Bezeichnung: "Nachtlauf", Principal: "Baurechner",
		Scope: domain.ScopeSchreiben,
	}
}

// **Ein rotiertes Token nimmt das neue Geheimnis sofort und das alte bis zur
// Frist — danach nur noch das neue.**
func TestEineRotationHaeltDasAlteGeheimnisEineStundeLang(t *testing.T) {
	t.Parallel()

	ablage := neuesTokendoppel()
	verwaltung, uhr := rotationsverwaltung(ablage)

	angelegt, altes, fehler := verwaltung.Anlegen(context.Background(), rotationstoken())
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	rotiert, neues, rotationsfehler := verwaltung.Rotieren(context.Background(), angelegt.ID)
	if rotationsfehler != nil {
		t.Fatalf("rotieren: %v", rotationsfehler)
	}

	if neues == altes || !strings.HasPrefix(neues, service.Tokenmarke+"_") {
		t.Fatalf("neues Geheimnis %q, altes %q", neues, altes)
	}

	if rotiert.VorgaengerBis != "2026-09-08T13:00:00Z" || rotiert.Fassung != 2 {
		t.Errorf("rotiert: %+v — erwartet Frist 13:00 und Fassung 2", rotiert)
	}

	for name, klartext := range map[string]string{"das neue": neues, "das alte": altes} {
		if _, err := verwaltung.Pruefe(context.Background(), klartext, domain.ScopeLesen); err != nil {
			t.Errorf("%s Geheimnis in der Frist: %v", name, err)
		}
	}

	// **Nach der Frist gilt nur noch das neue.**
	uhr.jetzt = uhr.jetzt.Add(domain.Tokenuebergang + time.Second)

	if _, err := verwaltung.Pruefe(context.Background(), altes, domain.ScopeLesen); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("das alte Geheimnis nach der Frist: %v", err)
	}

	if _, err := verwaltung.Pruefe(context.Background(), neues, domain.ScopeLesen); err != nil {
		t.Errorf("das neue Geheimnis nach der Frist: %v", err)
	}

	// **Eine zweite Rotation wirft das erste Geheimnis sofort weg** — es
	// gibt genau einen Vorgänger.
	_, drittes, fehler := verwaltung.Rotieren(context.Background(), angelegt.ID)
	if fehler != nil {
		t.Fatalf("zweite Rotation: %v", fehler)
	}

	if _, fehler := verwaltung.Pruefe(context.Background(), altes, domain.ScopeLesen); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("das erste Geheimnis nach zwei Rotationen: %v", fehler)
	}

	for name, klartext := range map[string]string{"das zweite": neues, "das dritte": drittes} {
		if _, fehler := verwaltung.Pruefe(context.Background(), klartext, domain.ScopeLesen); fehler != nil {
			t.Errorf("%s Geheimnis nach zwei Rotationen: %v", name, fehler)
		}
	}
}

// **Jede angenommene Prüfung vermerkt die Nutzung** — und ein Vermerk, der
// nicht geht, ist ein Fehler, kein stilles Weiter.
func TestEineAngenommenePruefungVermerktDieNutzung(t *testing.T) {
	t.Parallel()

	ablage := neuesTokendoppel()
	verwaltung, uhr := rotationsverwaltung(ablage)

	angelegt, klartext, fehler := verwaltung.Anlegen(context.Background(), rotationstoken())
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	uhr.jetzt = uhr.jetzt.Add(30 * time.Minute)

	if _, fehler := verwaltung.Pruefe(context.Background(), klartext, domain.ScopeLesen); fehler != nil {
		t.Fatalf("prüfen: %v", fehler)
	}

	if ablage.abgelegt[angelegt.ID].LetzteNutzung != "2026-09-08T12:30:00Z" {
		t.Errorf("letzte Nutzung %q", ablage.abgelegt[angelegt.ID].LetzteNutzung)
	}

	// **Ein abgewiesenes Token hinterlässt keinen Vermerk.**
	if _, fehler := verwaltung.Pruefe(context.Background(), klartext+"x", domain.ScopeLesen); fehler == nil {
		t.Fatal("ein falsches Geheimnis kam durch")
	}

	kaputt := errors.New("die Platte klemmt")
	ablage.vermerkfehler = kaputt

	if _, fehler := verwaltung.Pruefe(context.Background(), klartext, domain.ScopeLesen); !errors.Is(fehler, kaputt) {
		t.Errorf("Vermerkfehler: %v, erwartet den Fehler der Ablage", fehler)
	}
}

// **Der Sammelwiderruf zieht alle Token eines Profils zurück und sagt, wie
// viele.**
func TestDerSammelwiderrufSagtWieViele(t *testing.T) {
	t.Parallel()

	ablage := neuesTokendoppel()
	verwaltung, _ := rotationsverwaltung(ablage)

	for _, name := range []string{"Nachtlauf", "Bogenbau"} {
		eines := rotationstoken()
		eines.Bezeichnung = name

		if _, _, fehler := verwaltung.Anlegen(context.Background(), eines); fehler != nil {
			t.Fatalf("anlegen: %v", fehler)
		}
	}

	fremdes := rotationstoken()
	fremdes.ProfilID = 2

	if _, _, fehler := verwaltung.Anlegen(context.Background(), fremdes); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	gezogen, fehler := verwaltung.AlleWiderrufen(context.Background(), 1)
	if fehler != nil || gezogen != 2 {
		t.Errorf("%d gezogen, %v — erwartet 2", gezogen, fehler)
	}

	uebrig, lesefehler := verwaltung.Alle(context.Background(), 2)
	if lesefehler != nil || len(uebrig) != 1 {
		t.Errorf("im zweiten Profil: %d Token, %v", len(uebrig), lesefehler)
	}
}

// **Die Ablaufwahlen kommen aus der Uhr des Dienstes** — und nicht aus vier
// Uhren an vier Flächen.
func TestDieAblaufwahlenKommenAusDerUhr(t *testing.T) {
	t.Parallel()

	verwaltung, _ := rotationsverwaltung(neuesTokendoppel())

	wahlen := verwaltung.Ablaufwahlen()
	if len(wahlen) != 4 || wahlen[0].Name != "tag" || wahlen[0].Ablauf != "2026-09-09T12:00:00Z" {
		t.Errorf("Ablaufwahlen: %+v", wahlen)
	}
}

// **Jeder Fehlerweg wird betreten** (§3).
func TestDieFehlerwegeDerRotation(t *testing.T) {
	t.Parallel()

	ablage := neuesTokendoppel()
	verwaltung, _ := rotationsverwaltung(ablage)

	if _, _, fehler := verwaltung.Rotieren(context.Background(), 404); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Rotation eines Unbekannten: %v", fehler)
	}

	kaputt := errors.New("kaputt")
	ablage.kaputt = kaputt

	if _, _, fehler := verwaltung.Rotieren(context.Background(), 1); !errors.Is(fehler, kaputt) {
		t.Errorf("Rotation bei kaputter Ablage: %v", fehler)
	}

	if _, fehler := verwaltung.AlleWiderrufen(context.Background(), 1); !errors.Is(fehler, kaputt) {
		t.Errorf("Sammelwiderruf bei kaputter Ablage: %v", fehler)
	}
}
