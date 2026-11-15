package app

// **Ein Kartenbild, das dieser Bestand nicht selbst zeichnet** — P9
// (#2377),
// Befund C-12, Vermerk 0198.
//
// # Zwei Enden
//
// **Am Gast** fragt [App.Fremdkartenbild] den Wirt nach einem Objekt und legt
// die Antwort ins eigene Fach — unter der Bildkennung, nie unter einer Zeile.
// **Am Wirt** zeichnet [eigenmaler] mit der eigenen Zeile — und am Gast
// dessen eigene Karten (P9b); die Zeile verlässt
// ihn nicht.
//
// # Was der Gast prüft
//
// **Dass es ein PNG ist und die Maße stimmen.** Was über die Leitung kommt,
// geht über die Assetleitung in die Webansicht; ein Gegenüber, das etwas
// anderes schickte als das, was es behauptet, käme damit nicht bis dorthin.

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Bildfrist ist, wie lange der Gast auf ein Bild des Wirts wartet.
const Bildfrist = 10 * time.Second

// Fremdkartenbild gibt das Bild eines Objekts, das dieses Fenster nicht
// selbst zeichnen kann — erfragt beim Wirt des fremden Tisches.
//
// **Mit der Partie, aus deren Bild die Kennung stammt** (Zweitblick über
// Phase 7, Befund 4): ein Fenster kann einen eigenen Tisch führen und an
// einem fremden sitzen. Ohne sie ginge ein Wunsch vom eigenen Tisch an den
// fremden Wirt, und dessen Bild läge unter der Kennung des eigenen.
func (a *App) Fremdkartenbild(
	ctx context.Context, partie, objekt, bild, seite, guete string,
) TexturAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungLesen); fehler != nil {
		return leereTextur(ctx, fehler)
	}

	// **Mit der Bildkennung** (Nachlauf über Phase 7, R2): wer zeichnet,
	// prüft, dass sie die des Objekts ist — abgelegt wird darunter.
	wunsch := domain.Bildwunsch{Objekt: objekt, Seite: seite, Guete: guete, Bild: bild}
	if fehler := wunsch.Pruefe(); fehler != nil {
		return leereTextur(ctx, fehler)
	}

	if bild == "" || len(bild) > domain.Bildgrenze {
		return leereTextur(ctx, fmt.Errorf(
			"%w: ohne Bildkennung gibt es kein fremdes Bild", domain.ErrValidation))
	}

	// **Vor dem Fach**: ein Bild, das für einen anderen Tisch dort liegt,
	// gehört nicht an diesen.
	frage, fehler := a.bildweg(ctx, partie)
	if fehler != nil {
		return leereTextur(ctx, fehler)
	}

	schluessel := "fremd-" + partie + "-" + bild + "-" + seite + "-" + guete
	if adresse, breite, hoehe, da := a.spielbilder.Adresse(schluessel); da {
		return TexturAntwort{Textur: TexturDTO{
			Schluessel: schluessel, Adresse: adresse, Guete: guete, Seite: seite,
			Breite: breite, Hoehe: hoehe, Befunde: []string{},
		}}
	}

	frist, fertig := context.WithTimeout(ctx, Bildfrist)
	defer fertig()

	antwort, fehler := frage(frist, wunsch)
	if fehler != nil {
		return leereTextur(ctx, fehler)
	}

	if antwort.Grund != "" {
		return TexturAntwort{Textur: TexturDTO{
			Guete: guete, Seite: seite, Befunde: []string{antwort.Grund},
		}}
	}

	if fehler := pruefePNG(antwort); fehler != nil {
		return leereTextur(ctx, fehler)
	}

	return TexturAntwort{Textur: TexturDTO{
		Schluessel: schluessel,
		Adresse:    a.spielbilder.Hinterlege(schluessel, antwort.Png, antwort.Breite, antwort.Hoehe),
		Guete:      guete, Seite: seite, Breite: antwort.Breite, Hoehe: antwort.Hoehe,
		Befunde: []string{}, Zweiseitig: antwort.Zweiseitig,
	}}
}

// bildweg gibt den, der das Bild zeichnen kann — **am fremden Tisch** den
// Wirt, **am eigenen** den Gast, dessen mitgebrachtes Deck die Karte stellt
// (P9b, #2377) — und nur, wenn `partie` dieser Tisch ist.
func (a *App) bildweg(
	ctx context.Context, partie string,
) (func(context.Context, domain.Bildwunsch) (domain.Bildantwort, error), error) {
	// **Der eigene Tisch zuerst** (Nachlauf über Phase 7, R2): eine Partie,
	// die ein Gast beim Beitreten nur genannt hat, ist noch nicht bestätigt —
	// nannte er die Kennung seines eigenen Tisches, gingen dessen Bildrufe
	// sonst an den fremden Wirt.
	if offen, wirt := a.offeneFreigabe(); offen != nil && offen.Partie() == partie {
		return func(frist context.Context, wunsch domain.Bildwunsch) (domain.Bildantwort, error) {
			return a.partie.runde.BildVomEigner(frist, a.partienaht(), partie, wirt, wunsch)
		}, nil
	}

	// **Eine Sendeseite gibt es nur mit genannter Partie** (`beitrittMerken`,
	// `partieErfahren`) — deshalb genügt der Vergleich danach.
	if gast, _, fehler := a.gastAmTisch(ctx); fehler == nil {
		if _, fremde, _ := a.beitritt(); partie == fremde {
			return gast.Bild, nil
		}
	}

	return nil, fmt.Errorf(
		"%w: an diesem Tisch zeichnet niemand sonst Bilder", domain.ErrConflict)
}

// pruefePNG hält: ein PNG mit den Maßen, die die Antwort nennt.
func pruefePNG(antwort domain.Bildantwort) error {
	gelesen, fehler := png.DecodeConfig(bytes.NewReader(antwort.Png))
	if fehler != nil || gelesen.Width != antwort.Breite || gelesen.Height != antwort.Hoehe {
		return fmt.Errorf("%w: das Gegenüber schickte kein Bild mit den Maßen, "+
			"die es nannte", domain.ErrValidation)
	}

	return nil
}

// eigenmaler zeichnet eine Zeile des eigenen Bestands — am Wirt für seine
// Gäste, am Gast für seinen Wirt (P9b).
type eigenmaler struct{ a *App }

// Kartenbild zeichnet eine Kartenseite und gibt die Bytes aus dem Fach.
func (m eigenmaler) Kartenbild(
	ctx context.Context, karte int64, seite, guete string,
) (service.Kartenbild, error) {
	zeile, fehler := m.a.kartenleser.KarteLesen(ctx, karte)
	if fehler != nil {
		return service.Kartenbild{}, fehler
	}

	textur, fehler := m.a.texturwerk().Textur(ctx, zeile,
		service.Kartenseite(seite), service.Texturguete(guete))
	if fehler != nil {
		return service.Kartenbild{}, fehler
	}

	bild, da := m.a.spielbilder.Hole(textur.Schluessel)
	if !da {
		return service.Kartenbild{}, fmt.Errorf(
			"%w: das gezeichnete Bild liegt nicht mehr im Fach", domain.ErrConflict)
	}

	return service.Kartenbild{
		Png: bild, Breite: textur.Breite, Hoehe: textur.Hoehe,
		Zweiseitig: textur.Zweiseitig,
	}, nil
}
