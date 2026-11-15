package app

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Verlauf als Liste von Schritten** — Werkstück U6 von Phase 5.1
// (#1417).
// Hier steht nur die Übersetzung: Kennungen als Zeichenketten, der jüngste
// Schritt **zuerst** — die Seite zeigt, was eben geschah, und ⌘Z nimmt genau
// den obersten.

// VerlaufsschrittDTO ist ein Schritt des Verlaufs, wie das Fenster ihn zeigt.
type VerlaufsschrittDTO struct {
	ID        string `json:"id"`
	Zeitpunkt string `json:"zeitpunkt"`
	Handlung  string `json:"handlung"`
	KarteID   string `json:"karteID"`
	Feld      string `json:"feld"`

	Vorher     string `json:"vorher"`
	GabVorher  bool   `json:"gabVorher"`
	Nachher    string `json:"nachher"`
	GabNachher bool   `json:"gabNachher"`

	Zurueckgenommen bool   `json:"zurueckgenommen"`
	Urheber         string `json:"urheber"`
	Herkunft        string `json:"herkunft"`

	// Vorgang ist die Kennung, mit der sich genau dieser Schritt zurücknehmen
	// lässt (#1393) — leer bei einer Einzeländerung, die keine trägt.
	Vorgang string `json:"vorgang"`

	// Umfang ist die Zahl der Karten, die der Schritt berührt hat.
	Umfang int `json:"umfang"`
}

// VerlaufAntwort ist die Liste samt Fehler; `Schritte` ist nie `null`.
type VerlaufAntwort struct {
	Schritte []VerlaufsschrittDTO `json:"schritte"`
	Fehler   *Fehler              `json:"fehler"`
}

// Verlauf gibt die Schritte eines Sets, der jüngste zuerst.
func (a *App) Verlauf(ctx context.Context, setID string) VerlaufAntwort {
	ctx = a.imProfil(ctx)

	leer := VerlaufAntwort{Schritte: []VerlaufsschrittDTO{}}

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		leer.Fehler = befund

		return leer
	}

	schritte, err := (&Verlaufsweg{verlauf: a.verlauf}).Verlauf(ctx, kennung)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	return VerlaufAntwort{Schritte: schritte}
}

// Verlauf gibt die Schritte des Sets über die gemeinsame Naht — für Zeile
// und Dienst dieselbe Übersetzung wie fürs Fenster.
func (w *Verlaufsweg) Verlauf(ctx context.Context, setID int64) ([]VerlaufsschrittDTO, error) {
	eintraege, err := w.verlauf.Verlauf(ctx, setID)
	if err != nil {
		return nil, err
	}

	schritte := domain.Verlaufsschritte(eintraege)
	dtos := make([]VerlaufsschrittDTO, 0, len(schritte))

	// **Der jüngste zuerst**: der Bestand liefert den ältesten zuerst, die
	// Oberfläche will das Gegenteil, und ⌘Z bewegt den obersten.
	for i := len(schritte) - 1; i >= 0; i-- {
		dtos = append(dtos, schrittDTO(schritte[i]))
	}

	return dtos, nil
}

func schrittDTO(schritt domain.Verlaufsschritt) VerlaufsschrittDTO {
	gabVorher, gabNachher := schritt.GabVorher, schritt.GabNachher

	// **Ein Kartenschritt sagt, ob die Karte kam oder ging** (G01,
	// `g01einsteiger`): er trägt keinen Feldweg, sondern zwei Kartenstände,
	// und ohne diese zwei Angaben stand er als »angelegt oder gelöscht« da.
	if schritt.VorherKarte != nil && schritt.NachherKarte != nil {
		gabVorher, gabNachher = schritt.VorherKarte.GabEs, schritt.NachherKarte.GabEs
	}

	return VerlaufsschrittDTO{
		ID:              strconv.FormatInt(schritt.ID, 10),
		Zeitpunkt:       schritt.Zeitpunkt,
		Handlung:        schritt.Handlung,
		KarteID:         strconv.FormatInt(schritt.KarteID, 10),
		Feld:            schritt.Feld,
		Vorher:          schritt.Vorher,
		GabVorher:       gabVorher,
		Nachher:         schritt.Nachher,
		GabNachher:      gabNachher,
		Zurueckgenommen: schritt.Zurueckgenommen,
		Urheber:         string(schritt.Urheber),
		Herkunft:        schritt.Herkunft,
		Vorgang:         schritt.Gruppe,
		Umfang:          schritt.Umfang,
	}
}
