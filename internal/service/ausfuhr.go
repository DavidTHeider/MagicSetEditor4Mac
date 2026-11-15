package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjson"
)

// SetLeser gibt her, was für eine Ausfuhr nötig ist.
//
// Die Schnittstelle sagt, was `service` **braucht** — nicht, was `storage`
// anzubieten hat (§2). Für eine Ausfuhr sind das die Setköpfe und die Karten
// eines Sets; Rohblöcke gehören nicht dazu, denn MTG JSON hat für sie keinen
// Platz.
type SetLeser interface {
	Setkoepfe(ctx context.Context) ([]domain.SetHead, error)

	// Setwerte gibt die Angaben des Sets selbst.
	//
	// **Sie gehören dazu, obwohl MTG JSON die meisten nicht kennt.** Der
	// Setcode steht darunter, und ohne ihn greift der Schreiber auf den Namen
	// zurück: aus `PRB` würde `Probeset`. Der Zweitblick über Phase 9 hat es
	// gefunden.
	Setwerte(ctx context.Context, setID int64) ([]domain.Value, error)

	Karten(ctx context.Context, setID int64) ([]*domain.CardRow, error)
}

// Ausfuhr ist der Ausgang eines Exports.
//
// **Befunde und Erfolg schließen einander auch hier nicht aus** (§8): eine Datei
// entsteht, und was dabei nicht abbildbar war, steht daneben.
type Ausfuhr struct {
	Pfad    string
	Karten  int
	Befunde []domain.Befund
}

// Exporteur schreibt Sets aus dem Bestand hinaus.
type Exporteur struct {
	wache
	haushaltswahl

	leser SetLeser
}

// NeuerExporteur baut den Exporteur über einem Leser.
func NeuerExporteur(leser SetLeser) *Exporteur { return &Exporteur{leser: leser} }

// MitPruefer setzt den Riegel, der vor dem Schreiben fragt, ob dieser Aufruf
// es darf (0078).
func (e *Exporteur) MitPruefer(pruefer Pruefer) *Exporteur {
	e.pruefer = pruefer

	return e
}

// MTGJSON schreibt ein Set als MTG JSON an einen Pfad.
//
// **Geschrieben wird erst, wenn alles gelesen ist.** Wer die Datei anlegt und
// dann feststellt, dass das Set fehlt, hinterlässt eine halbe Ausgabe an einem
// Ort, den jemand ausgesucht hat.
func (e *Exporteur) MTGJSON(ctx context.Context, setID int64, spiel, pfad string) (*Ausfuhr, error) {
	if err := e.darfSchreiben(ctx); err != nil {
		return nil, err
	}

	freigeben, platzfehler := e.platz(ctx, domain.RessourceAusfuhr)
	if platzfehler != nil {
		return nil, platzfehler
	}

	defer freigeben()

	zuordnung, bekannt := mtgjson.ZuordnungFuer(spiel)
	if !bekannt {
		return nil, domain.NewFieldError("spiel", domain.SchluesselValidierung,
			"Für das Spiel %q gibt es keine Zuordnung von MTG-JSON-Feldern.", spiel)
	}

	set, err := e.setAus(ctx, setID)
	if err != nil {
		return nil, err
	}

	inhalt, befunde := zuordnung.Schreibe(set)

	// **Atomar**, wie jede andere Ausgabe an ein vom Nutzer gewähltes Ziel:
	// `os.WriteFile` kürzt eine vorhandene Datei beim Öffnen, und ein Fehler
	// danach ließe sie leer oder halb zurück. Angestrichen vom zweiten
	// Zweitblick über die Freigabe 0.3.0 (#1170, Befund 5).
	if err := schreibeAtomar(pfad, inhalt); err != nil {
		return nil, fmt.Errorf("MTG JSON schreiben: %w", err)
	}

	return &Ausfuhr{Pfad: pfad, Karten: len(set.Cards), Befunde: befunde}, nil
}

// setAus holt ein Set samt seinen Karten aus dem Bestand.
func (e *Exporteur) setAus(ctx context.Context, setID int64) (*domain.Set, error) {
	koepfe, err := e.leser.Setkoepfe(ctx)
	if err != nil {
		return nil, fmt.Errorf("Sets lesen: %w", err)
	}

	var kopf *domain.SetHead

	for i := range koepfe {
		if koepfe[i].ID == setID {
			kopf = &koepfe[i]
		}
	}

	if kopf == nil {
		return nil, fmt.Errorf("%w: das Set %d liegt nicht im Bestand", domain.ErrNotFound, setID)
	}

	werte, err := e.leser.Setwerte(ctx, setID)
	if err != nil {
		return nil, fmt.Errorf("Setangaben lesen: %w", err)
	}

	zeilen, err := e.leser.Karten(ctx, setID)
	if err != nil {
		return nil, fmt.Errorf("Karten lesen: %w", err)
	}

	set := &domain.Set{
		Name:   kopf.Name,
		Game:   kopf.Game,
		Style:  kopf.Style,
		Origin: kopf.Origin,
		Values: werte,
	}

	// Der Setcode kommt aus den Setangaben; führt das Set keinen, greift
	// `Schreibe` auf den Namen zurück. MTG JSON verlangt einen, und ein leerer
	// machte die Ausgabe unbrauchbar.
	for _, zeile := range zeilen {
		set.Cards = append(set.Cards, &domain.Card{Values: zeile.Values})
	}

	return set, nil
}
