package service

// **Die Parameterhilfen an einem Set** — Werkstück D von Phase 4.11
// (#1342).
//
// **Eine eigene Datei**, weil `schluesselwortpflege.go` sonst über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort steht, **was**
// ein Schlüsselwort ist und wie es gepflegt wird, hier, **womit** man seine
// Parameter schreibt.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Parameterhilfen sind beide Menüs zu einem Schlüsselwort.
//
// **Beide in einer Antwort**, weil sie zusammen gebraucht werden: wer einen
// Parameter einfügt, will gleich darauf verweisen. Zwei Rufe hießen zwei
// Stände desselben Musters (§7, Regel 2).
type Parameterhilfen struct {
	// Arten sind die Parameterarten des Spiels — für das **Muster**.
	Arten []Parametereinfuegung

	// Verweise sind die Formen für den **Erinnerungstext**, aus dem Muster
	// dieses Schlüsselworts gerechnet.
	Verweise []Parameterverweis

	Befunde []domain.Befund
}

// Parameterhilfen gibt beide Menüs zu einem Schlüsselwort des Sets.
//
// **Zu einem Schlüsselwort und nicht zu einem Muster.** Die Verweise hängen
// daran, was im Muster steht; ein Aufrufer, der das Muster mitschickte, hätte
// einen zweiten Stand davon — und der wäre der ältere.
func (p *Schluesselwortpflege) Parameterhilfen(
	ctx context.Context, setID, wortID int64,
) (Parameterhilfen, error) {
	liste, err := p.Liste(ctx, setID)
	if err != nil {
		return Parameterhilfen{}, err
	}

	wort, gefunden := wortMitKennung(liste.Woerter, wortID)
	if !gefunden {
		return Parameterhilfen{}, fmt.Errorf(
			"%w: dieses Set hat kein Schlüsselwort %d", domain.ErrNotFound, wortID)
	}

	vorlage := p.spielvorlage(ctx, setID)

	verweise, befunde := Parameterverweise(
		ctx, wort.Muster, vorlage.Parameterarten, vorlage.Spielskript)

	return Parameterhilfen{
		Arten:    Parameterangebot(vorlage.Parameterarten),
		Verweise: verweise,
		Befunde:  befunde,
	}, nil
}

// wortMitKennung sucht ein Schlüsselwort des Sets.
//
// **Nur die eigenen.** Ein festes trägt die Kennung 0 und gehört keinem Set;
// seine Parameter zu bearbeiten hieße, ein Spielpaket zu ändern.
func wortMitKennung(woerter []Schluesselwortsicht, wortID int64) (
	Schluesselwortsicht, bool,
) {
	for _, wort := range woerter {
		if wort.ID == wortID && !wort.Fest {
			return wort, true
		}
	}

	return Schluesselwortsicht{}, false
}
