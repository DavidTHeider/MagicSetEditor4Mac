package cardtrader

// **Blueprints und ihre Fremdkennungen** (A4, #1443).
//
// # Ein Blueprint ist nicht ein Verkaufsprodukt
//
// Ein **Blueprint** ist die Kartenfassung, die CardTrader kennt — „Beast //
// Plant aus Erweiterung 8". Ein **Produkt** ist das, was ein bestimmter
// Verkäufer davon anbietet, mit eigener Nummer, eigenem Zustand und eigenem
// Preis. Das Issue nennt die Trennung ausdrücklich; sie steht hier in zwei
// Feldern ([Blueprint.ID] und [Angebot.ProduktID]) und nicht in einem
// Kommentar.

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Blueprint ist eine Kartenfassung, wie CardTrader sie führt.
type Blueprint struct {
	ID            int64
	Name          string
	Fassung       string
	SpielID       int64
	ErweiterungID int64

	// Bild ist die Adresse des Kartenbilds — **flüchtig**.
	//
	// Ob CardTrader sie signiert oder ablaufen lässt, sagt die Referenz nicht
	// (0095). Eine Auskunft, die es nicht gibt, wird nicht erfunden: sie reist
	// mit einer Antwort und wird nicht abgelegt.
	Bild string

	// Kennungen sind die Fremdkennungen dieses Blueprints — seine eigene
	// zuerst, dann Scryfall, dann die Anbieterprodukte.
	Kennungen []domain.Druckkennung
}

// blueprintDTO ist die Gestalt, in der CardTrader einen Blueprint schickt.
//
// **`version`, `scryfall_id` und `tcg_player_id` können `null` sein**, und
// `card_market_ids` ist ein **Array** — alles drei am 2026-09-11 nachgelesen
// (0095). Ein `string` statt `*string` läse `null` als leeren Text, und das
// wäre dieselbe Zeichenkette wie eine wirklich leere Angabe.
type blueprintDTO struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Fassung       *string `json:"version"`
	SpielID       int64   `json:"game_id"`
	ErweiterungID int64   `json:"expansion_id"`
	Bild          string  `json:"image_url"`
	Scryfall      *string `json:"scryfall_id"`
	Cardmarket    []int64 `json:"card_market_ids"`
	TCGplayer     *string `json:"tcg_player_id"`
}

// Blueprints gibt die Kartenfassungen einer Erweiterung.
//
// **Ohne Erweiterung wird nicht gefragt**: `expansion_id` ist Pflicht, und
// CardTrader antwortet ohne sie mit 404 (0095). Eine Anfrage, von der das
// feststeht, ist Last für den Anbieter und Wartezeit für nichts.
func (k *Klient) Blueprints(ctx context.Context, erweiterung int64) (
	[]Blueprint, []domain.Befund, error,
) {
	if erweiterung <= 0 {
		return nil, nil, fmt.Errorf(
			"%w: ohne Erweiterung gibt CardTrader keine Blueprints — "+
				"gefragt wurde nach %d", domain.ErrValidation, erweiterung)
	}

	var roh []blueprintDTO

	abfrage := url.Values{"expansion_id": {strconv.FormatInt(erweiterung, 10)}}

	if fehler := k.hole(ctx, wege.Blueprints, abfrage, &roh); fehler != nil {
		return nil, nil, fehler
	}

	var (
		blueprints []Blueprint
		befunde    []domain.Befund
	)

	for _, eines := range roh {
		blueprint, befund := alsBlueprint(eines)
		if befund != nil {
			befunde = append(befunde, *befund)

			continue
		}

		blueprints = append(blueprints, blueprint)
	}

	return blueprints, befunde, nil
}

// alsBlueprint macht aus dem DTO die neutrale Gestalt.
//
// **Ein Blueprint ohne jede Kennung ist ein Befund** (§8): er lässt sich
// keinem Druck zuordnen, und still übergangen sähe die Antwort aus wie „diese
// Erweiterung hat weniger Karten".
func alsBlueprint(dto blueprintDTO) (Blueprint, *domain.Befund) {
	blueprint := Blueprint{
		ID: dto.ID, Name: dto.Name, Fassung: text(dto.Fassung),
		SpielID: dto.SpielID, ErweiterungID: dto.ErweiterungID, Bild: dto.Bild,
		Kennungen: kennungenAus(dto),
	}

	if len(blueprint.Kennungen) == 0 {
		return Blueprint{}, &domain.Befund{
			Datei: dto.Name, Ort: strconv.FormatInt(dto.ID, 10),
			Art: domain.ArtUmgebung,
			Meldung: fmt.Sprintf(
				"CardTrader führt %q ohne jede Kennung — weder eine eigene "+
					"noch eine von Scryfall, Cardmarket oder TCGplayer. "+
					"Diese Kartenfassung lässt sich keinem Druck zuordnen.",
				dto.Name),
		}
	}

	return blueprint, nil
}

// kennungenAus liest die Fremdkennungen eines Blueprints.
//
// # Die Semantik je System, und warum sie nicht überall »druck« ist
//
// **Ein Blueprint und eine Scryfall-ID nennen einen Druck**
// ([domain.SemantikDruck]). **Eine Cardmarket- oder TCGplayer-Nummer nennt ein
// Anbieterprodukt** ([domain.SemantikProdukt]) — und genau deshalb ist
// `card_market_ids` ein Array: dieselbe Kartenfassung kann dort mehrere
// Produkte haben. Wer daraus eine Kennung mit Semantik »druck« machte, hätte
// zwei Drucke, wo einer ist.
func kennungenAus(dto blueprintDTO) []domain.Druckkennung {
	var kennungen []domain.Druckkennung

	if dto.ID > 0 {
		kennungen = append(kennungen, domain.Druckkennung{
			System: domain.SystemCardTrader, Semantik: domain.SemantikDruck,
			Wert: strconv.FormatInt(dto.ID, 10),
		})
	}

	if wert := text(dto.Scryfall); wert != "" {
		kennung := domain.Druckkennung{
			System: domain.SystemScryfall, Semantik: domain.SemantikDruck, Wert: wert,
		}
		kennung.Normalisieren()
		kennungen = append(kennungen, kennung)
	}

	for _, nummer := range dto.Cardmarket {
		kennungen = append(kennungen, domain.Druckkennung{
			System: domain.SystemCardmarket, Semantik: domain.SemantikProdukt,
			Wert: strconv.FormatInt(nummer, 10),
		})
	}

	if wert := text(dto.TCGplayer); wert != "" {
		kennungen = append(kennungen, domain.Druckkennung{
			System: domain.SystemTCGplayer, Semantik: domain.SemantikProdukt, Wert: wert,
		})
	}

	return kennungen
}

// text gibt den Inhalt eines Zeigers — leer, wo `null` stand.
func text(zeiger *string) string {
	if zeiger == nil {
		return ""
	}

	return *zeiger
}
