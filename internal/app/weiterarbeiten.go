package app

// **Weiterarbeiten** — Werkstück U11 von Phase 5.1
// (#1422).
//
// # Vier Bindungen und ein Vermerk an vielen Stellen
//
// `Weiterarbeiten` liest die Übersicht; `Anheften` und `Loesen` halten ein
// Objekt oben; `DeckBenutzt` ist das Gegenstück zu `SetBenutzt` für Decks.
// **Vermerkt** wird dort, wo etwas geschieht — Set aufgeschlagen, Datei
// eingefahren, Set geschrieben, Paket bezogen — und der Vermerk ist
// Bestenfalls-Arbeit (§7, Regel 3): scheitert er, scheitert nicht der Weg,
// den jemand gerade geht; er steht im Protokoll.
//
// # Warum nur am Fenster
//
// Dieselbe Begründung wie bei `SetBenutzt`: was jemand zuletzt in der Hand
// hatte, ist eine Eigenschaft der **Bedienung**, nicht des Bestands. Eine
// Kommandozeile, die ein Set anfasst, hat keine Übersicht, in der es danach
// oben stünde. Steht so in `FAEHIGKEITEN.md`.

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Weiterarbeitsnaht ist die Naht zum Dienst — erfüllt von `service.Weiterarbeiten`.
type Weiterarbeitsnaht interface {
	Vermerken(ctx context.Context, objekt domain.Arbeitsobjekt, vorgang domain.Arbeitsvorgang) error
	Anheften(ctx context.Context, objekt domain.Arbeitsobjekt) error
	Loesen(ctx context.Context, art domain.Arbeitsart, kennung string) (bool, error)
	Uebersicht(ctx context.Context) (domain.Weiterarbeiten, error)
}

// MitWeiterarbeiten setzt die Naht — für Tests und den Composition Root.
func (a *App) MitWeiterarbeiten(naht Weiterarbeitsnaht) { a.weiterarbeiten = naht }

// ArbeitsobjektDTO ist ein Objekt der Übersicht, wie das Fenster es sieht.
type ArbeitsobjektDTO struct {
	Art        string `json:"art"`
	Kennung    string `json:"kennung"`
	Name       string `json:"name"`
	Angeheftet bool   `json:"angeheftet"`
}

// ArbeitsschrittDTO ist ein Schritt der Übersicht.
type ArbeitsschrittDTO struct {
	Objekt  ArbeitsobjektDTO `json:"objekt"`
	Vorgang string           `json:"vorgang"`
	// Zeitpunkt in RFC 3339, wie bei den Preisständen.
	Zeitpunkt string `json:"zeitpunkt"`
}

// WeiterarbeitenAntwort ist die Übersicht samt Umschlag.
type WeiterarbeitenAntwort struct {
	Zuletzt     *ArbeitsschrittDTO  `json:"zuletzt"`
	Geoeffnet   []ArbeitsschrittDTO `json:"geoeffnet"`
	Eingefuehrt []ArbeitsschrittDTO `json:"eingefuehrt"`
	Ausgegeben  []ArbeitsschrittDTO `json:"ausgegeben"`
	Angeheftet  []ArbeitsobjektDTO  `json:"angeheftet"`
	Fehler      *Fehler             `json:"fehler"`
}

// Weiterarbeiten gibt die Übersicht des Profils.
func (a *App) Weiterarbeiten(ctx context.Context) WeiterarbeitenAntwort {
	ctx = a.imProfil(ctx)

	leer := WeiterarbeitenAntwort{
		Geoeffnet: []ArbeitsschrittDTO{}, Eingefuehrt: []ArbeitsschrittDTO{},
		Ausgegeben: []ArbeitsschrittDTO{}, Angeheftet: []ArbeitsobjektDTO{},
	}

	if a.weiterarbeiten == nil {
		leer.Fehler = Uebersetze(ctx, errors.New("weiterarbeiten: die Naht ist nicht eingerichtet"))

		return leer
	}

	uebersicht, err := a.weiterarbeiten.Uebersicht(ctx)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	return alsWeiterarbeiten(uebersicht)
}

// Anheften hält ein Objekt in der Übersicht oben.
func (a *App) Anheften(ctx context.Context, art, kennung string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	if a.weiterarbeiten == nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, errors.New("weiterarbeiten: die Naht ist nicht eingerichtet"))}
	}

	objekt := domain.Arbeitsobjekt{Art: domain.Arbeitsart(art), Kennung: kennung}
	if befunde := objekt.Validate(); len(befunde) > 0 {
		return SchreibAntwort{Fehler: Uebersetze(ctx, befunde[0])}
	}

	objekt.Name = a.arbeitsname(ctx, objekt)

	if err := a.weiterarbeiten.Anheften(ctx, objekt); err != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return SchreibAntwort{}
}

// Loesen nimmt ein Objekt in der Übersicht von oben.
func (a *App) Loesen(ctx context.Context, art, kennung string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	if a.weiterarbeiten == nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, errors.New("weiterarbeiten: die Naht ist nicht eingerichtet"))}
	}

	if _, err := a.weiterarbeiten.Loesen(ctx, domain.Arbeitsart(art), kennung); err != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return SchreibAntwort{}
}

// DeckBenutzt merkt sich, dass ein Deck aufgeschlagen wurde — das Gegenstück
// zu [App.SetBenutzt].
func (a *App) DeckBenutzt(ctx context.Context, deckID string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("deckID", deckID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	a.arbeitVermerken(ctx, domain.ArbeitsartDeck, strconv.FormatInt(kennung, 10), domain.VorgangGeoeffnet)

	return SchreibAntwort{}
}

// arbeitVermerken hält einen Schritt fest — Bestenfalls-Arbeit.
func (a *App) arbeitVermerken(ctx context.Context, art domain.Arbeitsart, kennung string, vorgang domain.Arbeitsvorgang) {
	if a.weiterarbeiten == nil {
		return
	}

	objekt := domain.Arbeitsobjekt{Art: art, Kennung: kennung}
	objekt.Name = a.arbeitsname(ctx, objekt)

	if err := a.weiterarbeiten.Vermerken(ctx, objekt, vorgang); err != nil {
		slog.Default().WarnContext(ctx, "ein Schritt ließ sich nicht vermerken",
			"art", string(art), "stelle", kennung, "vorgang", string(vorgang), "grund", err.Error())
	}
}

// arbeitsname holt den Namen eines Objekts — oder gibt die Kennung zurück.
func (a *App) arbeitsname(ctx context.Context, objekt domain.Arbeitsobjekt) string {
	switch objekt.Art {
	case domain.ArbeitsartSet:
		if kopf, gefunden := a.setkopfZu(ctx, objekt.Kennung); gefunden && kopf.Name != "" {
			return kopf.Name
		}
	case domain.ArbeitsartDeck:
		if id, err := strconv.ParseInt(objekt.Kennung, 10, 64); err == nil && a.decks != nil {
			if deck, err := a.decks.Deck(ctx, id); err == nil && deck.Name != "" {
				return deck.Name
			}
		}
	case domain.ArbeitsartPaket:
		return objekt.Kennung
	}

	return objekt.Kennung
}

// setkopfZu findet den Kopf eines Sets im Profil — Bestenfalls-Arbeit.
func (a *App) setkopfZu(ctx context.Context, kennung string) (domain.SetHead, bool) {
	id, err := strconv.ParseInt(kennung, 10, 64)
	if err != nil || a.bestand == nil {
		return domain.SetHead{}, false
	}

	koepfe, err := a.bestand.Setkoepfe(ctx)
	if err != nil {
		return domain.SetHead{}, false
	}

	for _, kopf := range koepfe {
		if kopf.ID == id {
			return kopf, true
		}
	}

	return domain.SetHead{}, false
}

// alsWeiterarbeiten bildet die Übersicht auf den Umschlag ab.
func alsWeiterarbeiten(uebersicht domain.Weiterarbeiten) WeiterarbeitenAntwort {
	objekt := func(o domain.Arbeitsobjekt) ArbeitsobjektDTO {
		return ArbeitsobjektDTO{
			Art: string(o.Art), Kennung: o.Kennung, Name: o.Name,
			Angeheftet: uebersicht.IstAngeheftet(o.Art, o.Kennung),
		}
	}
	schritte := func(alle []domain.Arbeitsschritt) []ArbeitsschrittDTO {
		umgesetzt := make([]ArbeitsschrittDTO, 0, len(alle))
		for _, eines := range alle {
			umgesetzt = append(umgesetzt, ArbeitsschrittDTO{
				Objekt: objekt(eines.Objekt), Vorgang: string(eines.Vorgang),
				Zeitpunkt: eines.Zeitpunkt.UTC().Format(time.RFC3339),
			})
		}

		return umgesetzt
	}

	antwort := WeiterarbeitenAntwort{
		Geoeffnet: schritte(uebersicht.Geoeffnet), Eingefuehrt: schritte(uebersicht.Eingefuehrt),
		Ausgegeben: schritte(uebersicht.Ausgegeben), Angeheftet: []ArbeitsobjektDTO{},
	}

	for _, eines := range uebersicht.Angeheftet {
		antwort.Angeheftet = append(antwort.Angeheftet, objekt(eines))
	}

	if uebersicht.Zuletzt != nil {
		zuletzt := schritte([]domain.Arbeitsschritt{*uebersicht.Zuletzt})[0]
		antwort.Zuletzt = &zuletzt
	}

	return antwort
}

// arbeitsbestand sagt dem Dienst, ob es ein Objekt noch gibt — im Profil.
//
// **Über die Anwendung, nicht über kopierte Nähte**: die Nähte werden in
// verschiedenen Startschritten gesetzt, und ein Bestand, der sie beim Bauen
// kopierte, hielte für immer die leeren von damals.
type arbeitsbestand struct{ app *App }

// Gibt prüft über die Nähte, die das Fenster ohnehin hat.
// kennungAlsZahl liest eine Kennung als Zahl. Eine, die keine ist, meint kein
// Objekt — das ist kein Fehler, sondern die Antwort „gibt es nicht".
func kennungAlsZahl(kennung string) (int64, bool) {
	id, err := strconv.ParseInt(kennung, 10, 64)

	return id, err == nil
}

func (b arbeitsbestand) Gibt(ctx context.Context, objekt domain.Arbeitsobjekt) (bool, error) {
	switch objekt.Art {
	case domain.ArbeitsartSet:
		koepfe, err := b.app.bestand.Setkoepfe(ctx)
		if err != nil {
			return false, err
		}

		id, ok := kennungAlsZahl(objekt.Kennung)
		if !ok {
			return false, nil
		}

		for _, kopf := range koepfe {
			if kopf.ID == id {
				return true, nil
			}
		}

		return false, nil
	case domain.ArbeitsartDeck:
		id, ok := kennungAlsZahl(objekt.Kennung)
		if !ok {
			return false, nil
		}

		return gibtEs(b.app.decks.Deck(ctx, id))
	default:
		pakete, err := b.app.pakete.Liste(ctx)
		if err != nil {
			return false, err
		}

		for _, paket := range pakete {
			if paket.Name == objekt.Kennung {
				return true, nil
			}
		}

		return false, nil
	}
}

// gibtEs macht aus »nicht gefunden« ein Nein und lässt jeden anderen Fehler stehen.
func gibtEs[T any](_ T, err error) (bool, error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return false, nil
	case err != nil:
		return false, err
	}

	return true, nil
}

// weiterarbeitenUeber baut den Dienst über den Nähten des Fensters.
func (a *App) weiterarbeitenUeber(ablage service.Arbeitsablage) *service.Weiterarbeiten {
	return service.NeuesWeiterarbeiten(ablage, arbeitsbestand{app: a}, a.uhr)
}

// Weiterarbeiten gibt die Übersicht des Profils — die Fensterbindung.
func (b *Bindungen) Weiterarbeiten(ctx context.Context) WeiterarbeitenAntwort {
	return b.app(ctx).Weiterarbeiten(b.ruf(ctx))
}

// Anheften hält ein Objekt oben — die Fensterbindung.
func (b *Bindungen) Anheften(ctx context.Context, art, kennung string) SchreibAntwort {
	return b.app(ctx).Anheften(b.ruf(ctx), art, kennung)
}

// Loesen nimmt ein Objekt von oben — die Fensterbindung.
func (b *Bindungen) Loesen(ctx context.Context, art, kennung string) SchreibAntwort {
	return b.app(ctx).Loesen(b.ruf(ctx), art, kennung)
}

// DeckBenutzt merkt sich ein aufgeschlagenes Deck — die Fensterbindung.
func (b *Bindungen) DeckBenutzt(ctx context.Context, deckID string) SchreibAntwort {
	return b.app(ctx).DeckBenutzt(b.ruf(ctx), deckID)
}
