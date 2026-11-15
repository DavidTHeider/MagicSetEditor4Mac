package cardtrader

// **Das eigene Konto, und zwar nur lesend** — Werkstück `T11` von Phase 5.6
// (#1476).
//
// # Drei Adressen und keine vierte
//
// `GET /info`, `GET /products/export` und `GET /orders`. Sie stehen in
// derselben Positivliste wie die vier Marktwege, und sie sind aus demselben
// Holz: **kein `POST`, kein `PUT`, kein `PATCH`, kein `DELETE`.**
//
// **Was der Anbieter außerdem anbietet, ist bekannt und bleibt ungebaut**:
// `POST /products`, `PUT`/`DELETE /products/:id`, `POST /cart/add`,
// `/cart/remove`, `/cart/purchase`, `PUT /orders/:id/ship` und die drei
// Widerrufswege. Warum keiner davon hier steht, sagt
// [0122](../../docs/entscheidungen/0122-ein-fremdes-konto-wird-gelesen.md) —
// es ist eine Entscheidung und kein Vergessen.
//
// # Und warum das `shared_secret` nicht gelesen wird
//
// **Weil der Anbieter „Do not share this information" daneben schreibt.** Die
// billigste Art, es nicht weiterzugeben, ist, es nicht zu haben: es steht in
// keinem Feld dieses Programms, und damit in keinem Export, keinem
// Diagnosepaket und keinem Umschlag (§11).

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kontoDTO ist die Gestalt, in der CardTrader es schickt.
type kontoDTO struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	NutzerID int64  `json:"user_id"`
}

// eigenangebotDTO ist die Gestalt, in der CardTrader ein eigenes Produkt
// schickt — **nicht dieselbe wie ein fremdes Marktangebot**: dort stehen
// Verkäufer und Land, hier die eigene Notiz.
type eigenangebotDTO struct {
	ID           int64  `json:"id"`
	BlueprintID  int64  `json:"blueprint_id"`
	Name         string `json:"name_en"`
	Menge        int    `json:"quantity"`
	PreisCent    int64  `json:"price_cents"`
	Waehrung     string `json:"price_currency"`
	Eigennotiz   string `json:"user_data_field"`
	Beschreibung string `json:"description"`
}

// bestellungDTO ist die Gestalt, in der CardTrader eine Bestellung schickt.
type bestellungDTO struct {
	ID             int64  `json:"id"`
	Code           string `json:"code"`
	Zustand        string `json:"state"`
	UeberNullkonto bool   `json:"via_cardtrader_zero"`
	Rolle          string `json:"order_as"`
	Stueck         int    `json:"size"`
	Bezahlt        string `json:"paid_at"`
}

// Konto gibt, was der Zugang über sich sagt — und prüft ihn damit.
//
// **Der Anbieter nennt `/info` selbst den Weg, die Anmeldung zu prüfen.**
func (k *Klient) Konto(ctx context.Context) (domain.Marktkonto, error) {
	var roh kontoDTO

	if fehler := k.hole(ctx, wege.Konto, nil, &roh); fehler != nil {
		return domain.Marktkonto{}, fehler
	}

	return domain.Marktkonto{AppID: roh.ID, Name: roh.Name, NutzerID: roh.NutzerID}, nil
}

// EigeneAngebote gibt die Posten, die dieses Konto anbietet.
//
// **`erweiterung` grenzt ein — null heißt alle.** Der Anbieter warnt, dass
// der ganze Abruf bei großen Beständen Sekunden dauert und empfiehlt eine
// Lesefrist von 120 bis 180 Sekunden (nachgelesen am 2026-09-19); wer eine
// Erweiterung nennt, bleibt darunter.
func (k *Klient) EigeneAngebote(ctx context.Context, erweiterung int64) (
	[]domain.Eigenangebot, error,
) {
	abfrage := url.Values{}
	if erweiterung > 0 {
		abfrage.Set("expansion_id", strconv.FormatInt(erweiterung, 10))
	}

	var roh []eigenangebotDTO

	if fehler := k.hole(ctx, wege.EigeneAngebote, abfrage, &roh); fehler != nil {
		return nil, fehler
	}

	angebote := make([]domain.Eigenangebot, 0, len(roh))
	for _, eines := range roh {
		angebote = append(angebote, domain.Eigenangebot{
			ID: eines.ID, BlueprintID: eines.BlueprintID, Name: eines.Name,
			Menge: eines.Menge, PreisCent: eines.PreisCent,
			Waehrung: eines.Waehrung, Eigennotiz: eines.Eigennotiz,
			Beschreibung: eines.Beschreibung,
		})
	}

	return angebote, nil
}

// Bestellungen gibt die Käufe dieses Kontos — samt Befunden.
//
// **`seite` und `menge` sind die Parameter des Anbieters** (nachgelesen am
// 2026-09-19): `page` zählt ab eins, `limit` liegt zwischen 1 und 100 und
// ist ohne Angabe 20. Eine Null heißt hier „nimm die Vorgabe des Anbieters"
// — und nicht „nimm alle": eine erfundene Obergrenze wäre eine Zusage, die
// dieses Programm nicht halten kann.
//
// **Ein unbekannter Zustand ist ein Befund und keine fehlende Zeile** (§8):
// die Bestellung kommt mit, und daneben steht, was an ihr nicht aufging.
func (k *Klient) Bestellungen(ctx context.Context, seite, menge int) (
	[]domain.Marktbestellung, []domain.Befund, error,
) {
	abfrage := url.Values{}
	if seite > 0 {
		abfrage.Set("page", strconv.Itoa(seite))
	}

	if menge > 0 {
		abfrage.Set("limit", strconv.Itoa(menge))
	}

	var roh []bestellungDTO

	if fehler := k.hole(ctx, wege.Bestellungen, abfrage, &roh); fehler != nil {
		return nil, nil, fehler
	}

	bestellungen := make([]domain.Marktbestellung, 0, len(roh))
	befunde := []domain.Befund{}

	for _, eine := range roh {
		zustand := domain.Bestellzustand(eine.Zustand)

		if !domain.IstBestellzustand(zustand) {
			befunde = append(befunde, domain.Befund{
				Ort: fmt.Sprintf("Bestellung %d", eine.ID),
				Meldung: fmt.Sprintf(
					"CardTrader nennt den Zustand %q — dieses Programm kennt "+
						"ihn nicht und bucht daraus nichts", eine.Zustand),
				Art: domain.ArtEingabe,
			})
		}

		bestellungen = append(bestellungen, domain.Marktbestellung{
			ID: eine.ID, Code: eine.Code, Zustand: zustand,
			UeberNullkonto: eine.UeberNullkonto, Rolle: eine.Rolle,
			Stueck: eine.Stueck, Bezahlt: eine.Bezahlt,
		})
	}

	return bestellungen, befunde, nil
}
