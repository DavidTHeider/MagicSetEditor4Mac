// **Decks über den Dienst** — Werkstück 7 von Phase 4.6
// (#1293).
//
// Einordnung nach §16: **erweiternd** — elf Wege kommen hinzu, keiner ändert
// einen bestehenden.

package rest

import (
	"context"
	"net/http"
)

// Deckwerkstatt ist, was der Dienst für Decks und Formate braucht.
//
// **Gleichlautend mit `mcp.Decks`**, und das ist Absicht: dieselbe Fähigkeit,
// dieselbe Naht (§16). Die Zuordnung besorgt Gos strukturelle Typisierung.
type Deckwerkstatt interface {
	Decks(ctx context.Context, spiel string) (any, error)
	Deck(ctx context.Context, deckID int64) (any, error)
	DeckAnlegen(ctx context.Context, name, spiel, notiz string) (any, error)
	DeckUmbenennen(ctx context.Context, deckID, fassung int64, name, notiz string) (
		any, error)
	DeckLoeschen(ctx context.Context, deckID int64) error
	DeckKopieren(ctx context.Context, deckID int64, name string) (any, error)
	DeckAusSet(ctx context.Context, setID int64, name string, menge int) (any, error)
	DeckFehlkartenWuenschen(ctx context.Context, deckID int64) (any, error)
	DeckplatzLegen(ctx context.Context, deckID, karteID int64,
		printing string, menge int, bereich string) (any, error)
	DeckplatzNehmen(ctx context.Context, deckID, platzID int64) error
	DeckplatzVerschieben(ctx context.Context, deckID, platzID int64, bereich string) (
		any, error)
	DeckFormatWaehlen(ctx context.Context, deckID, formatID int64) (any, error)
	DeckPruefen(ctx context.Context, deckID int64) (any, error)
	DeckAbgleichen(ctx context.Context, deckID int64) (any, error)

	// **Die Zuordnung** — Werkstück U9 von Phase 5.1 (#1420).
	DeckZuordnungen(ctx context.Context, deckID int64) (any, error)
	DeckZuordnen(ctx context.Context, deckID, platzID, exemplarID int64, menge int) (any, error)
	DeckZuordnungLoesen(ctx context.Context, deckID, zuordnungID int64) error
	DeckFehlkarten(ctx context.Context, deckID int64) (any, error)

	// Deckliste gibt ein Deck als Liste (NET-04, #1708).
	Deckliste(ctx context.Context, deckID int64) (any, error)
	ExemplarDecks(ctx context.Context, exemplarID int64) (any, error)

	Bauarten(ctx context.Context) (any, error)
	Formate(ctx context.Context, spiel string) (any, error)
	FormatAusBauart(ctx context.Context, bauart, name, spiel, scryfallname string) (
		any, error)
	FormatSetzen(ctx context.Context, format map[string]any) (any, error)
	FormatLoeschen(ctx context.Context, formatID int64) error
}

// decks nennt die Decks — `?spiel=` schränkt ein.
func (d *Dienst) decks(schreiben http.ResponseWriter, anfrage *http.Request) {
	liste, fehler := d.naehte.Decks.Decks(
		anfrage.Context(), anfrage.URL.Query().Get("spiel"))
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, liste)
}

// deck gibt ein Deck samt seinen Plätzen.
func (d *Dienst) deck(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.zumDeck(schreiben, anfrage, d.naehte.Decks.Deck)
}

// deckPruefung beurteilt ein Deck gegen sein Format.
func (d *Dienst) deckPruefung(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.zumDeck(schreiben, anfrage, d.naehte.Decks.DeckPruefen)
}

// deckBestand sagt, was von einem Deck in der Sammlung liegt.
func (d *Dienst) deckBestand(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.zumDeck(schreiben, anfrage, d.naehte.Decks.DeckAbgleichen)
}

// deckZuordnungen gibt Bedarf und Zuordnungen (U9, #1420).
func (d *Dienst) deckZuordnungen(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.zumDeck(schreiben, anfrage, d.naehte.Decks.DeckZuordnungen)
}

// deckliste gibt ein Deck als Liste (NET-04, #1708).
func (d *Dienst) deckliste(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.zumDeck(schreiben, anfrage, d.naehte.Decks.Deckliste)
}

// deckFehlkarten gibt, was zu besorgen wäre.
func (d *Dienst) deckFehlkarten(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.zumDeck(schreiben, anfrage, d.naehte.Decks.DeckFehlkarten)
}

// exemplarDecks sagt, welche Decks ein Exemplar benutzen — der Rückweg.
func (d *Dienst) exemplarDecks(schreiben http.ResponseWriter, anfrage *http.Request) {
	exemplarID, gut := kennung(schreiben, anfrage, "exemplarID")
	if !gut {
		return
	}

	ergebnis, fehler := d.naehte.Decks.ExemplarDecks(anfrage.Context(), exemplarID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, ergebnis)
}

// deckZuordnen legt Exemplare in einen Platz.
func (d *Dienst) deckZuordnen(schreiben http.ResponseWriter, anfrage *http.Request) {
	deckID, gut := kennung(schreiben, anfrage, "deckID")
	if !gut {
		return
	}

	var wunsch struct {
		PlatzID    int64 `json:"platzID"`
		ExemplarID int64 `json:"exemplarID"`
		Menge      int   `json:"menge"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	ergebnis, fehler := d.naehte.Decks.DeckZuordnen(
		anfrage.Context(), deckID, wunsch.PlatzID, wunsch.ExemplarID, wunsch.Menge)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, ergebnis)
}

// deckZuordnungLoesen nimmt eine Zuordnung heraus.
func (d *Dienst) deckZuordnungLoesen(schreiben http.ResponseWriter, anfrage *http.Request) {
	deckID, gut := kennung(schreiben, anfrage, "deckID")
	if !gut {
		return
	}

	zuordnungID, gutzwei := kennung(schreiben, anfrage, "zuordnungID")
	if !gutzwei {
		return
	}

	if fehler := d.naehte.Decks.DeckZuordnungLoesen(anfrage.Context(), deckID, zuordnungID); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{"zuordnungID": zuordnungID, "geloest": true})
}

// zumDeck bedient die drei Wege, die nur eine Deckkennung nehmen.
func (d *Dienst) zumDeck(
	schreiben http.ResponseWriter, anfrage *http.Request,
	arbeit func(context.Context, int64) (any, error),
) {
	deckID, gut := kennung(schreiben, anfrage, "deckID")
	if !gut {
		return
	}

	ergebnis, fehler := arbeit(anfrage.Context(), deckID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, ergebnis)
}

// deckAnlegen legt ein leeres Deck an.
//
// **`POST` und `201`**: es entsteht etwas Neues, und die Antwort trägt seine
// Kennung.
func (d *Dienst) deckAnlegen(schreiben http.ResponseWriter, anfrage *http.Request) {
	var wunsch struct {
		Name  string `json:"name"`
		Spiel string `json:"spiel"`
		Notiz string `json:"notiz"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	deck, fehler := d.naehte.Decks.DeckAnlegen(
		anfrage.Context(), wunsch.Name, wunsch.Spiel, wunsch.Notiz)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, deck)
}

// deckUmbenennen schreibt Name und Notiz fort.
//
// **`PUT` mit der Fassung im Rumpf** — sie ist die Bedingung (§7, Regel 2).
func (d *Dienst) deckUmbenennen(schreiben http.ResponseWriter, anfrage *http.Request) {
	deckID, gut := kennung(schreiben, anfrage, "deckID")
	if !gut {
		return
	}

	var wunsch struct {
		Fassung int64  `json:"fassung"`
		Name    string `json:"name"`
		Notiz   string `json:"notiz"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	deck, fehler := d.naehte.Decks.DeckUmbenennen(
		anfrage.Context(), deckID, wunsch.Fassung, wunsch.Name, wunsch.Notiz)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, deck)
}

// deckKopieren legt eine Kopie unter neuem Namen an — DOM-03.AK-02
// (#1698).
//
// **`POST` auf die Kopie des Decks**: es entsteht ein neues, und ein zweiter
// Ruf legt ein zweites an — oder scheitert am vergebenen Namen.
func (d *Dienst) deckKopieren(schreiben http.ResponseWriter, anfrage *http.Request) {
	deckID, gut := kennung(schreiben, anfrage, "deckID")
	if !gut {
		return
	}

	var wunsch struct {
		Name string `json:"name"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	kopie, fehler := d.naehte.Decks.DeckKopieren(anfrage.Context(), deckID, wunsch.Name)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, kopie)
}

// deckFehlkartenWuenschen setzt, was fehlt, auf die Wunschliste — DOM-03.AK-02
// (#1698). **`POST` ohne Rumpf**: das Deck sagt, was fehlt.
func (d *Dienst) deckFehlkartenWuenschen(schreiben http.ResponseWriter, anfrage *http.Request) {
	d.zumDeck(schreiben, anfrage, d.naehte.Decks.DeckFehlkartenWuenschen)
}

// deckLoeschen nimmt ein Deck samt seinen Plätzen heraus.
func (d *Dienst) deckLoeschen(schreiben http.ResponseWriter, anfrage *http.Request) {
	deckID, gut := kennung(schreiben, anfrage, "deckID")
	if !gut {
		return
	}

	if fehler := d.naehte.Decks.DeckLoeschen(anfrage.Context(), deckID); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{})
}

// deckplatzLegen legt eine Karte ins Deck.
//
// **`PUT` auf die Sammlung der Plätze**: derselbe Bezug im selben Bereich ist
// eine höhere Menge und kein zweiter Platz — der Ruf ist damit idempotent.
func (d *Dienst) deckplatzLegen(schreiben http.ResponseWriter, anfrage *http.Request) {
	deckID, gut := kennung(schreiben, anfrage, "deckID")
	if !gut {
		return
	}

	var wunsch struct {
		KarteID  int64  `json:"karteID"`
		Printing string `json:"printing"`
		Menge    int    `json:"menge"`
		Bereich  string `json:"bereich"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	platz, fehler := d.naehte.Decks.DeckplatzLegen(anfrage.Context(), deckID,
		wunsch.KarteID, wunsch.Printing, wunsch.Menge, wunsch.Bereich)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, platz)
}

// deckplatzNehmen nimmt einen Platz heraus.
func (d *Dienst) deckplatzNehmen(schreiben http.ResponseWriter, anfrage *http.Request) {
	deckID, gutesDeck := kennung(schreiben, anfrage, "deckID")
	platzID, guterPlatz := kennung(schreiben, anfrage, "platzID")

	if !gutesDeck || !guterPlatz {
		return
	}

	if fehler := d.naehte.Decks.DeckplatzNehmen(
		anfrage.Context(), deckID, platzID); fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, map[string]any{})
}

// deckFormat stellt ein Deck auf ein Format — **0 nimmt es herunter**.
func (d *Dienst) deckFormat(schreiben http.ResponseWriter, anfrage *http.Request) {
	deckID, gut := kennung(schreiben, anfrage, "deckID")
	if !gut {
		return
	}

	var wunsch struct {
		FormatID int64 `json:"formatID"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	deck, fehler := d.naehte.Decks.DeckFormatWaehlen(
		anfrage.Context(), deckID, wunsch.FormatID)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusOK, deck)
}

// deckAusSet legt aus den Karten eines Sets ein Deck an — FMT-04
// (#1693).
//
// **`POST` auf die Sammlung der Decks**: es entsteht ein neues, und ein
// zweiter Ruf legt ein zweites an.
func (d *Dienst) deckAusSet(schreiben http.ResponseWriter, anfrage *http.Request) {
	var wunsch struct {
		SetID int64  `json:"setID"`
		Name  string `json:"name"`
		Menge int    `json:"menge"`
	}

	if !gelesen(schreiben, anfrage, &wunsch) {
		return
	}

	deck, fehler := d.naehte.Decks.DeckAusSet(anfrage.Context(), wunsch.SetID, wunsch.Name, wunsch.Menge)
	if fehler != nil {
		fehlerhaft(schreiben, fehler)

		return
	}

	ja(schreiben, http.StatusCreated, deck)
}
