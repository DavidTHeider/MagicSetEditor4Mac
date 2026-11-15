package app

// **Ein Modell sieht sich ein Bild an** — Werkstück A13.3 von Phase 5.3
// (#1452).
//
// # Eine wahlfreie Schnittstelle, wie bei der Bildquelle
//
// **Eine Laufzeit erfüllt sie oder nicht** — geprüft wird mit einer
// Typzusicherung und nicht mit einer Liste daneben. Dasselbe Muster wie
// [KIBildquelle] aus A12, und aus demselben Grund: eine Liste liefe mit den
// Nähten auseinander.
//
// # Und warum das Bild als Bytes kommt
//
// **Weil ein Bild dieses Programm als Bytes erreicht oder gar nicht**
// ([0102](../../docs/entscheidungen/0102-wo-ein-erzeugtes-bild-liegt.md)).
// Der Vermerk sagt es für ein **erzeugtes** Bild; für ein gelesenes gilt er
// genauso: eine Adresse, die dieses Programm für ein Modell holte, wäre ein
// Holweg, den niemand gewollt hat.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// KISichtquelle ist, was eine Laufzeit können muss, um ein Bild zu lesen.
type KISichtquelle interface {
	Sehen(ctx context.Context, wunsch KISichtwunsch) (KIRohantwort, error)
}

// KISichtwunsch ist, was die Naht an den Anbieter gibt.
type KISichtwunsch struct {
	Anweisung string
	Alias     string
	Auftrag   string

	// Bild ist das Bild selbst — **nicht** seine Adresse (0102).
	Bild      []byte
	Medientyp string

	Schema map[string]any
}

// sichtquelleAus sagt, ob dieser Anbieter Bilder liest.
//
// **Die Meldung nennt den Anbieter**, damit niemand rät, welcher von vieren
// gemeint ist — und sie sagt, woran es liegt: nicht am Modell, sondern daran,
// dass dieser Weg den Anbieter gar nicht erreicht.
func sichtquelleAus(quelle KILaufzeitquelle, anbieter string) (KISichtquelle, error) {
	sichtquelle, kann := quelle.(KISichtquelle)
	if !kann {
		return nil, fmt.Errorf(
			"%w: %q liest keine Bilder — dafür gibt es einen Anbieter, der "+
				"%q belegt hat",
			domain.ErrValidation, anbieter, string(domain.KIVision))
	}

	return sichtquelle, nil
}
