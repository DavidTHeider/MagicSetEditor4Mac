package app

// **Die Naht zum Marktplatz** — Werkstück A4 von Phase 5.3
// (#1443).
//
// # Was hier geschieht und was nicht
//
// Hier wird der Klient gebaut — Adresse, Fristen, Drossel, Erlaubnis, Zugang
// — und die Gestalt von CardTrader auf die von `domain` gebracht. **Eine
// fachliche Entscheidung fällt hier keine** (§2): welcher Blueprint zu welchem
// Druck gehört, entscheidet `service`.

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/cardtrader"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// CardTraderUmgebung ist die Umgebungsvariable, die die Grundadresse umlenkt.
//
// **Dieselbe Naht wie [MTGJSONPreisumgebung] und aus demselben Grund** (§3):
// ohne sie ginge jeder Flächenlauf gegen api.cardtrader.com.
const CardTraderUmgebung = "MSE_CARDTRADER_ADRESSE"

// cardtraderadresse gibt die umgelenkte Adresse — oder die Vorgabe.
func cardtraderadresse() string {
	if ort := os.Getenv(CardTraderUmgebung); ort != "" {
		return ort
	}

	return cardtrader.Vorgabeadresse
}

// cardtradermarkt bringt den Klienten an [service.Marktquelle].
type cardtradermarkt struct{ klient *cardtrader.Klient }

// Quelle ist die Familie, unter der Preise von hier laufen.
func (c cardtradermarkt) Quelle() domain.Preisquelle { return cardtrader.Familie }

// System ist das Kennungssystem, unter dem CardTrader seine Blueprints führt.
func (c cardtradermarkt) System() domain.Kennungssystem { return domain.SystemCardTrader }

// Fassungen gibt die Kartenfassungen eines Sets.
//
// **Zwei Anfragen und nicht eine**: CardTrader ordnet Blueprints nach seiner
// eigenen `expansion_id`, und die steht in der Erweiterungsliste. Wer den
// Setcode direkt einsetzte, riete eine Nummer.
func (c cardtradermarkt) Fassungen(ctx context.Context, setcode string) (
	[]domain.Fremdfassung, []domain.Befund, error,
) {
	erweiterungen, fehler := c.klient.Erweiterungen(ctx)
	if fehler != nil {
		return nil, nil, fehler
	}

	gefunden := int64(0)

	for _, eine := range erweiterungen {
		// **Nur Magic** (`game_id: 1`): CardTrader führt sechs Spiele, und
		// Setcodes wiederholen sich zwischen ihnen.
		if eine.SpielID == cardtrader.SpielMagic && strings.EqualFold(eine.Code, setcode) {
			gefunden = eine.ID

			break
		}
	}

	if gefunden == 0 {
		return nil, nil, fmt.Errorf(
			"%w: CardTrader führt kein Magic-Set mit dem Code %q",
			domain.ErrNotFound, setcode)
	}

	blueprints, befunde, fehler := c.klient.Blueprints(ctx, gefunden)
	if fehler != nil {
		return nil, nil, fehler
	}

	fassungen := make([]domain.Fremdfassung, 0, len(blueprints))
	for _, blueprint := range blueprints {
		fassungen = append(fassungen, domain.Fremdfassung{
			Kennung:   strconv.FormatInt(blueprint.ID, 10),
			Name:      blueprint.Name,
			Kennungen: blueprint.Kennungen,
		})
	}

	return fassungen, befunde, nil
}

// Angebote gibt die Posten zu einer Fassung.
func (c cardtradermarkt) Angebote(
	ctx context.Context, fassung string, wunsch domain.Angebotswunsch,
) ([]domain.Marktangebot, []domain.Befund, error) {
	nummer, fehler := strconv.ParseInt(fassung, 10, 64)
	if fehler != nil {
		return nil, nil, domain.MitGrund(domain.ErrValidation, fehler,
			"die Fassungsnummer %q lesen", fassung)
	}

	angebote, befunde, holfehler := c.klient.Angebote(ctx, nummer,
		cardtrader.Angebotsfilter{Sprache: wunsch.Sprache, Finish: wunsch.Finish})
	if holfehler != nil {
		return nil, nil, holfehler
	}

	posten := make([]domain.Marktangebot, 0, len(angebote))
	for _, angebot := range angebote {
		posten = append(posten, domain.Marktangebot{
			Fremdkennung: strconv.FormatInt(angebot.ProduktID, 10),
			Preis:        angebot.Preis,
			Waehrung:     angebot.Waehrung,
			Menge:        angebot.Menge,
			Buendel:      angebot.Buendel,
			Finish:       angebot.Finish,
			Zustand:      angebot.Zustand,
			Sprache:      angebot.Sprache,
			Signiert:     angebot.Signiert,
			Veraendert:   angebot.Veraendert,
			Gradiert:     angebot.Gradiert,
			ImUrlaub:     angebot.ImUrlaub,
			Verkaeufer:   angebot.Verkaeufer,
			Land:         angebot.Land,
		})
	}

	return posten, befunde, nil
}

// cardtradermarktUeber baut den Marktplatz — oder sagt, was ihm fehlt.
//
// **Der Schalter erreicht ihn** (§16, Regel 6): `dienstklient` fragt vor jeder
// Anfrage, ob dieser Dienst an ist.
func cardtradermarktUeber(ctx context.Context, wahl Dienstwahl, bund Geheimnisleser) (
	cardtradermarkt, error,
) {
	marke, fehler := geheimnis(ctx, bund,
		domain.ZugangKonto(domain.DienstCardTrader, domain.ZugangMarke))
	if fehler != nil {
		return cardtradermarkt{}, fehler
	}

	if marke == "" {
		return cardtradermarkt{}, domain.ZugangFehlt(
			domain.DienstCardTrader, domain.ZugangMarke)
	}

	adresse := cardtraderadresse()

	return cardtradermarkt{klient: cardtrader.NeuerKlient(adresse,
		dienstklient(wahl, domain.DienstCardTrader),
		cardtrader.Zugang{Marke: marke}).
		MitDrossel(drosselwerk.Fuer(adresse, cardtrader.Abstand))}, nil
}
