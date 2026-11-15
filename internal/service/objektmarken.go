package service

// **Marken an jedem Fachobjekt** — UX-12
// (#1684),
// Vermerk 0135, Marke `new`.
//
// # Was dieser Anwendungsfall entscheidet
//
// **Dass jede Marke ihre Normalform bekommt und jedes Objekt dem Profil
// gehört, bevor etwas geschrieben wird** — an jeder der vier Flächen
// derselbe Weg (§16). Eine Marke an einem Objekt, das es im Profil nicht
// gibt, ist ein Tippfehler an der Kommandozeile oder ein Leck zwischen
// Profilen; beides wird hier abgewiesen, nicht im Aufrufer.
//
// # Was er nicht entscheidet
//
// **Ob jemand schreiben darf.** Das hält die Verbindung des Profils: wer
// nur lesen darf, hat eine, die nicht schreibt (0073). Und **keine Prüfung
// dieses Programms liest eine Marke** — eine Marke ist nie ein Recht
// (UX-12.AK-02).

import (
	"context"
	"fmt"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Objektmarkenablage ist, was die Marken vom Bestand brauchen.
type Objektmarkenablage interface {
	Vorhanden(ctx context.Context, art domain.Markenart, kennung string) (bool, error)
	Setzen(ctx context.Context, art domain.Markenart, kennung, marke string) error
	SetzenAlle(ctx context.Context, art domain.Markenart, kennungen []string, marke string) error
	Entfernen(ctx context.Context, art domain.Markenart, kennung, marke string) error
	An(ctx context.Context, art domain.Markenart, kennung string) ([]string, error)
	JeObjekt(ctx context.Context, art domain.Markenart) (map[string][]string, error)
	Katalog(ctx context.Context) ([]domain.Markenposten, error)
	Umbenennen(ctx context.Context, alt, neu string, zusammenfuehren bool) (int, error)
	UeberallEntfernen(ctx context.Context, marke string) (int, error)
}

// Objektmarken setzt, liest, filtert und pflegt Marken jeder Art.
type Objektmarken struct{ ablage Objektmarkenablage }

// NeueObjektmarken baut den Anwendungsfall.
func NeueObjektmarken(ablage Objektmarkenablage) *Objektmarken { return &Objektmarken{ablage: ablage} }

// Setzen hängt eine Marke an ein Objekt.
func (o *Objektmarken) Setzen(ctx context.Context, art, kennung, marke string) error {
	gelesen, kennung, normal, fehler := o.stelle(ctx, art, kennung, marke)
	if fehler != nil {
		return fehler
	}

	return o.ablage.Setzen(ctx, gelesen, kennung, normal)
}

// Entfernen nimmt eine Marke von einem Objekt.
func (o *Objektmarken) Entfernen(ctx context.Context, art, kennung, marke string) error {
	gelesen, kennung, normal, fehler := o.stelle(ctx, art, kennung, marke)
	if fehler != nil {
		return fehler
	}

	return o.ablage.Entfernen(ctx, gelesen, kennung, normal)
}

// An gibt die Marken eines Objekts.
func (o *Objektmarken) An(ctx context.Context, art, kennung string) ([]string, error) {
	gelesen, kennung, fehler := o.objekt(ctx, art, kennung)
	if fehler != nil {
		return nil, fehler
	}

	return o.ablage.An(ctx, gelesen, kennung)
}

// JeObjekt gibt die Marken jedes Objekts einer Art — nach Kennung, für die
// Listen der Sichten.
func (o *Objektmarken) JeObjekt(ctx context.Context, art string) (map[string][]string, error) {
	gelesen, fehler := domain.PruefeMarkenart(art)
	if fehler != nil {
		return nil, fehler
	}

	return o.ablage.JeObjekt(ctx, gelesen)
}

// Vergeben hängt eine Marke an mehrere Objekte einer Art — **ganz oder gar
// nicht** (UX-12.AK-02).
func (o *Objektmarken) Vergeben(ctx context.Context, art string, kennungen []string, marke string) error {
	gelesen, fehler := domain.PruefeMarkenart(art)
	if fehler != nil {
		return fehler
	}

	normal, fehler := domain.PruefeMarke(marke)
	if fehler != nil {
		return fehler
	}

	if len(kennungen) == 0 {
		return domain.NewFieldError("kennungen", domain.SchluesselPflichtfeld,
			"Nenne wenigstens ein Objekt, das die Marke bekommen soll.")
	}

	gelesene := make([]string, 0, len(kennungen))

	for _, roh := range kennungen {
		kennung, fehler := gelesen.Kennung(roh)
		if fehler != nil {
			return fehler
		}

		gelesene = append(gelesene, kennung)
	}

	eindeutig := slices.Compact(slices.Sorted(slices.Values(gelesene)))

	return o.ablage.SetzenAlle(ctx, gelesen, eindeutig, normal)
}

// Filtern gibt die Kennungen der Objekte einer Art, die durch den Filter
// gehen — geordnet.
//
// **Ohne Marke gibt es hier keinen Filter**, sondern eine Ablehnung: „alle
// Decks" beantwortet die Liste der Decks, nicht die Ablage der Marken.
func (o *Objektmarken) Filtern(ctx context.Context, art string, marken []string, verknuepfung string) ([]string, error) {
	gelesen, fehler := domain.PruefeMarkenart(art)
	if fehler != nil {
		return nil, fehler
	}

	if len(marken) == 0 {
		return nil, domain.NewFieldError("marken", domain.SchluesselPflichtfeld,
			"Nenne wenigstens eine Marke, nach der gefiltert wird.")
	}

	filter, fehler := domain.NeuerMarkenfilter(marken, verknuepfung)
	if fehler != nil {
		return nil, fehler
	}

	je, fehler := o.ablage.JeObjekt(ctx, gelesen)
	if fehler != nil {
		return nil, fehler
	}

	getroffen := []string{}

	for kennung, getragen := range je {
		if filter.Trifft(getragen) {
			getroffen = append(getroffen, kennung)
		}
	}

	slices.Sort(getroffen)

	return getroffen, nil
}

// Katalog zählt die Marken des Profils.
func (o *Objektmarken) Katalog(ctx context.Context) ([]domain.Markenposten, error) {
	return o.ablage.Katalog(ctx)
}

// Umbenennen gibt einer Marke einen neuen Namen — auf einen vorhandenen nur
// mit `zusammenfuehren`.
func (o *Objektmarken) Umbenennen(ctx context.Context, alt, neu string, zusammenfuehren bool) (int, error) {
	altNormal, fehler := domain.PruefeMarke(alt)
	if fehler != nil {
		return 0, fehler
	}

	neuNormal, fehler := domain.PruefeMarke(neu)
	if fehler != nil {
		return 0, fehler
	}

	// **Dieselbe Normalform ist kein neuer Name** — „Tausch" auf „tausch"
	// umzubenennen ändert nichts, und das ist keine Ablehnung wert.
	if altNormal == neuNormal {
		return 0, nil
	}

	zahl, fehler := o.ablage.Umbenennen(ctx, altNormal, neuNormal, zusammenfuehren)
	if fehler == nil && zahl == 0 {
		return 0, fmt.Errorf("%w: »%s« trägt in diesem Profil kein Objekt", domain.ErrNotFound, altNormal)
	}

	return zahl, fehler
}

// UeberallEntfernen nimmt eine Marke von jedem Objekt des Profils.
func (o *Objektmarken) UeberallEntfernen(ctx context.Context, marke string) (int, error) {
	normal, fehler := domain.PruefeMarke(marke)
	if fehler != nil {
		return 0, fehler
	}

	zahl, fehler := o.ablage.UeberallEntfernen(ctx, normal)
	if fehler == nil && zahl == 0 {
		return 0, fmt.Errorf("%w: »%s« trägt in diesem Profil kein Objekt", domain.ErrNotFound, normal)
	}

	return zahl, fehler
}

// stelle prüft Art, Objekt und Marke — und gibt die Kennung in ihrer einen
// Form zurück.
func (o *Objektmarken) stelle(ctx context.Context, art, roh, marke string) (gelesen domain.Markenart, kennung, normal string, fehler error) {
	gelesen, kennung, fehler = o.objekt(ctx, art, roh)
	if fehler != nil {
		return "", "", "", fehler
	}

	normal, fehler = domain.PruefeMarke(marke)

	return gelesen, kennung, normal, fehler
}

// objekt sagt, ob es das Objekt im Profil gibt, und bringt seine Kennung in
// die Form, unter der es seine Marken trägt ([domain.Markenart.Kennung]).
func (o *Objektmarken) objekt(ctx context.Context, art, roh string) (domain.Markenart, string, error) {
	gelesen, fehler := domain.PruefeMarkenart(art)
	if fehler != nil {
		return "", "", fehler
	}

	kennung, fehler := gelesen.Kennung(roh)
	if fehler != nil {
		return "", "", fehler
	}

	gibt, fehler := o.ablage.Vorhanden(ctx, gelesen, kennung)
	if fehler != nil {
		return "", "", fehler
	}

	if !gibt {
		return "", "", fmt.Errorf("%w: %s %s gibt es in diesem Profil nicht", domain.ErrNotFound, gelesen, kennung)
	}

	return gelesen, kennung, nil
}
