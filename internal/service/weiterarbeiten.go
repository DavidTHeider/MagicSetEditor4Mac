package service

// **Weiterarbeiten** — Werkstück U11 von Phase 5.1
// (#1422).
//
// # Was der Dienst tut, was die Ablage nicht tun kann
//
// Die Ablage führt Schritte und Angeheftetes je Profil. Ob es das Objekt
// **noch gibt**, weiß sie nicht — ein gelöschtes Set steht sonst mit seinem
// Namen in der Übersicht und führt beim Klick ins Leere. Der Dienst fragt
// den Bestand, lässt Verschwundenes fallen und vergisst es in der Ablage.
// Das Vergessen ist Bestenfalls-Arbeit (§7, Regel 3): scheitert es, steht
// die Übersicht trotzdem, nur ohne das Verschwundene.
//
// **Kein fremder Profilname leckt**, weil die Ablage am Profil aus dem
// Kontext schneidet — der Dienst sieht nie Zeilen eines anderen Profils.

import (
	"context"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Arbeitsablage ist die Ablage der Schritte und des Angehefteten — je Profil.
type Arbeitsablage interface {
	Vermerken(ctx context.Context, schritt domain.Arbeitsschritt) error
	Schritte(ctx context.Context) ([]domain.Arbeitsschritt, error)
	Anheften(ctx context.Context, objekt domain.Arbeitsobjekt) error
	Loesen(ctx context.Context, art domain.Arbeitsart, kennung string) (bool, error)
	Angeheftete(ctx context.Context) ([]domain.Arbeitsobjekt, error)
	Vergessen(ctx context.Context, art domain.Arbeitsart, kennung string) error
}

// Arbeitsbestand sagt, ob es ein Objekt noch gibt — im Profil des Kontexts.
type Arbeitsbestand interface {
	Gibt(ctx context.Context, objekt domain.Arbeitsobjekt) (bool, error)
}

// Hoechstens ist, wie viele Schritte je Vorgang die Übersicht zeigt.
const Hoechstens = 6

// Weiterarbeiten ist der Dienst hinter der Übersicht.
type Weiterarbeiten struct {
	ablage  Arbeitsablage
	bestand Arbeitsbestand
	uhr     func() time.Time
}

// NeuesWeiterarbeiten baut den Dienst.
func NeuesWeiterarbeiten(ablage Arbeitsablage, bestand Arbeitsbestand, uhr func() time.Time) *Weiterarbeiten {
	return &Weiterarbeiten{ablage: ablage, bestand: bestand, uhr: uhr}
}

// Vermerken hält fest, dass mit einem Objekt etwas geschah — jetzt.
func (w *Weiterarbeiten) Vermerken(ctx context.Context, objekt domain.Arbeitsobjekt, vorgang domain.Arbeitsvorgang) error {
	schritt := domain.Arbeitsschritt{Objekt: objekt, Vorgang: vorgang, Zeitpunkt: w.uhr()}
	if befunde := schritt.Validate(); len(befunde) > 0 {
		return befunde[0]
	}

	return w.ablage.Vermerken(ctx, schritt)
}

// Anheften hält ein Objekt oben — wenn es das Objekt gibt.
func (w *Weiterarbeiten) Anheften(ctx context.Context, objekt domain.Arbeitsobjekt) error {
	if befunde := objekt.Validate(); len(befunde) > 0 {
		return befunde[0]
	}

	gibt, err := w.bestand.Gibt(ctx, objekt)
	if err != nil {
		return err
	}

	if !gibt {
		return fmt.Errorf("%w: %s %s gibt es nicht — angeheftet wird nur, was da ist",
			domain.ErrNotFound, objekt.Art, objekt.Kennung)
	}

	return w.ablage.Anheften(ctx, objekt)
}

// Loesen nimmt ein Objekt von oben — und sagt, ob es dort war.
func (w *Weiterarbeiten) Loesen(ctx context.Context, art domain.Arbeitsart, kennung string) (bool, error) {
	return w.ablage.Loesen(ctx, art, kennung)
}

// Uebersicht gibt die Übersicht — nur mit dem, was es noch gibt.
func (w *Weiterarbeiten) Uebersicht(ctx context.Context) (domain.Weiterarbeiten, error) {
	schritte, err := w.ablage.Schritte(ctx)
	if err != nil {
		return domain.Weiterarbeiten{}, err
	}

	angeheftet, err := w.ablage.Angeheftete(ctx)
	if err != nil {
		return domain.Weiterarbeiten{}, err
	}

	vorhanden := map[string]bool{}
	gesehen := map[string]bool{}

	pruefe := func(objekt domain.Arbeitsobjekt) (bool, error) {
		schluessel := string(objekt.Art) + ":" + objekt.Kennung
		if gesehen[schluessel] {
			return vorhanden[schluessel], nil
		}

		gesehen[schluessel] = true

		gibt, err := w.bestand.Gibt(ctx, objekt)
		if err != nil {
			return false, err
		}

		vorhanden[schluessel] = gibt

		if !gibt {
			// Bestenfalls-Arbeit: ein Fehler hier ändert die Übersicht nicht.
			_ = w.ablage.Vergessen(ctx, objekt.Art, objekt.Kennung) //nolint:errcheck // §7 Regel 3
		}

		return gibt, nil
	}

	bleibende := make([]domain.Arbeitsschritt, 0, len(schritte))

	for _, schritt := range schritte {
		gibt, err := pruefe(schritt.Objekt)
		if err != nil {
			return domain.Weiterarbeiten{}, err
		}

		if gibt {
			bleibende = append(bleibende, schritt)
		}
	}

	bleibendAngeheftet := make([]domain.Arbeitsobjekt, 0, len(angeheftet))

	for _, objekt := range angeheftet {
		gibt, err := pruefe(objekt)
		if err != nil {
			return domain.Weiterarbeiten{}, err
		}

		if gibt {
			bleibendAngeheftet = append(bleibendAngeheftet, objekt)
		}
	}

	return domain.Weiterarbeit(bleibende, bleibendAngeheftet, Hoechstens), nil
}
