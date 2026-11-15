package app

// **Was je Art abgelegt wird** — NET-04 (#1708). Jede Art nimmt den Weg,
// den ihre Ausfuhr schon geht; neu ist nur, **wohin**: in den Ordner der
// Teilung und nicht an einen Ort, den jemand gewählt hat.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Setabzieher schreibt ein Set als Archiv, ohne seinen Arbeitsort zu
// verschieben — [service.Setschreiber.Abzug] —, und nennt seinen Namen.
type Setabzieher interface {
	Abzug(ctx context.Context, setID int64, archiv string) error
	Setname(ctx context.Context, setID int64) (string, error)
}

// teilgutBauen legt die Kopie einer Art ab.
func (a *App) teilgutBauen(
	ctx context.Context, stelle Teilablage, art domain.Teilart,
	bezug, kennung, ordner string,
) (domain.Teilgut, string, error) {
	switch art {
	case domain.TeilartEinladung:
		return a.einladungTeilen(stelle, bezug, kennung)
	case domain.TeilartKartenbild:
		return a.kartenbildTeilen(ctx, bezug, ordner)
	case domain.TeilartSet:
		return a.setTeilen(ctx, bezug, ordner)
	case domain.TeilartDeck:
		return a.deckTeilen(ctx, stelle, bezug, kennung)
	default:
		return a.sammlungTeilen(ctx, stelle, kennung)
	}
}

// einladungTeilen legt den Code als Text ab — geteilt wird der Text selbst.
func (a *App) einladungTeilen(stelle Teilablage, code, kennung string) (domain.Teilgut, string, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return domain.Teilgut{}, "", domain.MitGrund(domain.ErrValidation, nil,
			"es gibt keinen Code zu teilen — erzeuge zuerst eine Einladung")
	}

	return abgelegt(stelle, kennung, domain.TeilartEinladung, "Einladung", ".txt", []byte(code+"\n"), code)
}

// kartenbildTeilen zeichnet eine Karte als PNG in den Ordner der Teilung.
func (a *App) kartenbildTeilen(ctx context.Context, karteID, ordner string) (domain.Teilgut, string, error) {
	kennung, fehler := strconv.ParseInt(karteID, 10, 64)
	if fehler != nil {
		return domain.Teilgut{}, "", domain.MitGrund(domain.ErrValidation, fehler, "%q ist keine Kennung einer Karte", karteID)
	}

	titel := "Karte " + karteID

	geschrieben, fehler := a.bildausfuhr.Schreibe(ctx, kennung, ordner,
		domain.Teildateiname(titel, ""), service.Bildformat("png"), 0,
		domain.BildkonfliktUeberschreiben, render.GueteMittel)
	if fehler != nil {
		return domain.Teilgut{}, "", fehler
	}

	return dateigut(domain.TeilartKartenbild, titel, geschrieben.Pfad)
}

// setTeilen schreibt ein Set als `.mse-set`-Archiv — ohne den Arbeitsort
// zu verschieben.
func (a *App) setTeilen(ctx context.Context, setID, ordner string) (domain.Teilgut, string, error) {
	kennung, fehler := strconv.ParseInt(setID, 10, 64)
	if fehler != nil {
		return domain.Teilgut{}, "", domain.MitGrund(domain.ErrValidation, fehler, "%q ist keine Kennung eines Sets", setID)
	}

	abzieher, kann := a.speicherer.(Setabzieher)
	if !kann {
		return domain.Teilgut{}, "", ohneTeilen()
	}

	titel := "Set " + setID
	if name, namensfehler := abzieher.Setname(ctx, kennung); namensfehler == nil && name != "" {
		titel = name
	}

	archiv := filepath.Join(ordner, domain.Teildateiname(titel, ".mse-set"))
	if fehler := abzieher.Abzug(ctx, kennung, archiv); fehler != nil {
		return domain.Teilgut{}, "", fehler
	}

	return dateigut(domain.TeilartSet, titel, archiv)
}

// deckTeilen legt ein Deck als Liste ab — dieselbe wie [App.DecklisteAusgeben].
func (a *App) deckTeilen(ctx context.Context, stelle Teilablage, deckID, kennung string) (domain.Teilgut, string, error) {
	id, fehler := strconv.ParseInt(deckID, 10, 64)
	if fehler != nil {
		return domain.Teilgut{}, "", domain.MitGrund(domain.ErrValidation, fehler, "%q ist keine Kennung eines Decks", deckID)
	}

	deck, fehler := a.decks.Deck(ctx, id)
	if fehler != nil {
		return domain.Teilgut{}, "", fehler
	}

	titel := deck.Name
	if titel == "" {
		titel = "Deck " + deckID
	}

	return abgelegt(stelle, kennung, domain.TeilartDeck, titel, ".txt", []byte(service.Deckliste(deck)), "")
}

// sammlungTeilen legt die Sammlung als CSV ab — **die verlustfreie Form**.
func (a *App) sammlungTeilen(ctx context.Context, stelle Teilablage, kennung string) (domain.Teilgut, string, error) {
	if a.sammlungsausfuhr == nil {
		return domain.Teilgut{}, "", keineSammlungsausfuhr()
	}

	text, fehler := a.sammlungsausfuhr.Ausgeben(ctx, "csv")
	if fehler != nil {
		return domain.Teilgut{}, "", fehler
	}

	return abgelegt(stelle, kennung, domain.TeilartSammlung, "Sammlung", ".csv", []byte(text), "")
}

// abgelegt schreibt einen Inhalt in die Teilung und beschreibt ihn.
func abgelegt(
	stelle Teilablage, kennung string, art domain.Teilart,
	titel, endung string, inhalt []byte, text string,
) (domain.Teilgut, string, error) {
	name := domain.Teildateiname(titel, endung)

	pfad, groesse, fehler := stelle.Ablegen(kennung, name, inhalt)
	if fehler != nil {
		return domain.Teilgut{}, "", fehler
	}

	return domain.Teilgut{Art: art, Titel: titel, Dateiname: name, Groesse: groesse, Text: text}, pfad, nil
}

// dateigut beschreibt eine Datei, die ein anderer Weg geschrieben hat.
func dateigut(art domain.Teilart, titel, pfad string) (domain.Teilgut, string, error) {
	info, fehler := os.Stat(pfad)
	if fehler != nil {
		return domain.Teilgut{}, "", fmt.Errorf("die geteilte Datei lesen: %w", fehler)
	}

	return domain.Teilgut{Art: art, Titel: titel, Dateiname: filepath.Base(pfad), Groesse: info.Size()}, pfad, nil
}
