package service

// **Der Gast fragt nach einem Kartenbild** — P9
// (#2377),
// Befund C-12, Vermerk 0198.
//
// # Warum er wartet und nicht die Antwort irgendwohin legt
//
// **Weil das Fenster auf ein Bild wartet**, wie bei einer Textur aus dem
// eigenen Bestand. Der Wunsch geht hinaus, die Antwort kommt über dieselbe
// Leitung wie jede Nachricht des Wirts — hier wird sie dem zugestellt, der
// genau sie erfragt hat.
//
// # Und warum eine späte Antwort kein Fehler ist
//
// **Weil sie der Leitung nichts tut.** Wer nicht mehr wartet, hat die Frist
// verstreichen lassen; eine Antwort danach legte sonst den Tisch nieder.

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// MitBildpost setzt die Post, die dieser Gast mit den anderen seiner Sitzung
// teilt.
func (g *Partiegast) MitBildpost(post *Bildpost) *Partiegast {
	g.post = post

	return g
}

// Bild fragt den Wirt nach einem Kartenbild und wartet auf seine Antwort.
func (g *Partiegast) Bild(
	ctx context.Context, wunsch domain.Bildwunsch,
) (domain.Bildantwort, error) {
	if fehler := wunsch.Pruefe(); fehler != nil {
		return domain.Bildantwort{}, fehler
	}

	return g.post.frage(ctx, wunsch, g.hinaus, "der Wirt")
}

// hinaus legt eine Kontrolle in die Schlange und gibt sie hinaus.
func (g *Partiegast) hinaus(ctx context.Context, inhalt []byte) error {
	if fehler := g.sitzung.Sende(ctx, domain.ArtKontrolle, inhalt); fehler != nil {
		return fehler
	}

	return g.sitzung.Leere(ctx)
}

// bildNehmen stellt die Antwort des Wirts zu.
func (g *Partiegast) bildNehmen(roh []byte) error {
	antwort, fehler := liesBildantwort(roh)
	if fehler != nil {
		return fehler
	}

	g.post.liefere(antwort)

	return nil
}

// liesBildantwort liest eine Antwort und hält sie gegen ihre Gestalt.
func liesBildantwort(roh []byte) (domain.Bildantwort, error) {
	var antwort domain.Bildantwort
	if fehler := json.Unmarshal(roh, &antwort); fehler != nil {
		return domain.Bildantwort{}, fmt.Errorf("%w: der Inhalt ist kein Bild: %s",
			domain.ErrValidation, fehler.Error())
	}

	return antwort, antwort.Pruefe()
}
