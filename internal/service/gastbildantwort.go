package service

// **Der Gast zeichnet das Bild seiner eigenen Karte** — P9b
// (#2377),
// Vermerk 0198.
//
// # Warum aus der letzten Sicht
//
// **Weil sie sagt, was dieser Gast sehen darf und was ihm gehört.** Der Wirt
// rechnet sie für seinen Bestand: eine Zeile steht darin nur bei Karten aus
// dem Deck, das er mitgebracht hat. Eine Frage nach einem Objekt, das sie
// nicht nennt oder verdeckt zeigt, bekommt einen Grund — auch vom Wirt, der
// den Abzug ohnehin kennt; gezeigt wird nur, was gesehen werden darf.
//
// # Wer fragt
//
// **Nur ein Wirt, an dessen Tisch dieser Gast ein Deck mitgebracht hat** —
// und mitbringen kann erst diese Fassung. Ein älterer Gast bekommt die Frage
// also nie; er kennte die Kontrollart nicht und ginge.

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// MitMaler setzt, womit dieser Gast die Bilder seiner eigenen Karten zeichnet.
func (g *Partiegast) MitMaler(maler Kartenbildmaler) *Partiegast {
	g.maler = maler

	return g
}

// bildAntworten zeichnet, was der Wirt erfragt — oder sagt, warum nicht.
func (g *Partiegast) bildAntworten(ctx context.Context, roh []byte) error {
	var wunsch domain.Bildwunsch
	if fehler := json.Unmarshal(roh, &wunsch); fehler != nil {
		return fmt.Errorf("%w: der Bildwunsch ist unlesbar: %s",
			domain.ErrValidation, fehler.Error())
	}

	if fehler := wunsch.Pruefe(); fehler != nil {
		return fehler
	}

	inhalt, _ := domain.VerpackeKontrolle( //nolint:errcheck // Bytes, Zahlen und Zeichenketten klagen nicht
		domain.KontrolleBild, g.bildFuer(ctx, wunsch))

	return g.hinaus(ctx, inhalt)
}

// bildFuer gibt das Bild oder den Grund.
func (g *Partiegast) bildFuer(
	ctx context.Context, wunsch domain.Bildwunsch,
) domain.Bildantwort {
	if g.maler == nil {
		return domain.Bildantwort{
			Wunsch: wunsch,
			Grund:  "dieser Gast zeichnet keine Bilder",
		}
	}

	for _, objekt := range g.sicht.Objekte {
		if objekt.Kennung == wunsch.Objekt && !objekt.Geheim && objekt.KarteID != 0 {
			return gemalt(ctx, g.maler, objekt.KarteID, wunsch)
		}
	}

	return domain.Bildantwort{
		Wunsch: wunsch,
		Grund: "dieses Objekt zeichnet dieser Gast nicht — es liegt nicht " +
			"offen vor ihm, oder sein Deck gehört jemand anderem",
	}
}
