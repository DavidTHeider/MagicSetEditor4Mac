package cardtrader

// **Marktangebote** (A4, #1443).
//
// # Die Antwort ist nach Blueprint geordnet
//
// **Nachgelesen am 2026-09-11** (0095): `{"10050": [ … ]}` — nicht ein Array.
// Das fällt bei einem Beispiel mit einem Blueprint nicht auf, solange man nur
// die innere Klammer ansieht, und wäre im Betrieb ein Zerteilfehler.

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Angebot ist, was ein Verkäufer anbietet.
type Angebot struct {
	BlueprintID int64

	// ProduktID ist die Nummer **dieses Angebots** — nicht die der
	// Kartenfassung. Die Trennung steht in [Blueprint].
	ProduktID int64

	Name string

	// Preis ist der Betrag in der kleinsten Einheit der Währung.
	//
	// **CardTrader nennt ihn als ganze Zahl** (`"cents": 217`), und er bleibt
	// eine ganze Zahl: ein Umweg über `float64` ist bei Geld der Fehler, den
	// [0093](../../docs/entscheidungen/0093-eine-preisreihe-hat-sieben-teile.md)
	// vermeidet.
	//
	// **Bei einer Währung ohne Nachkommastellen ist das eine Annahme** — G20
	// aus dem Zweitblick über Phase 6.3 (#1664). Die Referenz sagt „The price
	// in cents, in your currency" (nachgelesen am 2026-10-04) und nennt keine
	// Währung wie den Yen; ob dann Yen oder Hundertstel gemeint sind, steht
	// dort nicht. Gelesen wird die kleinste Einheit der Währung, wie bei
	// jedem anderen Dienst hier.
	Preis    int64
	Waehrung string

	Menge int

	// Buendel ist die Zahl der Karten, die zusammen verkauft werden.
	Buendel int

	Finish domain.Finish

	// Zustand ist die Erhaltung **in der Ausdrucksweise von CardTrader** —
	// `Near Mint`, `Slightly Played`, … Sie wird nicht auf [domain.Zustand]
	// abgebildet: sechs Stufen auf sieben verlangte Gleichsetzungen, die
	// niemand nachprüfen kann (0095).
	Zustand string

	Sprache string

	Signiert   bool
	Veraendert bool
	Gradiert   bool

	// ImUrlaub sagt, dass der Verkäufer gerade nicht versendet.
	ImUrlaub bool

	Verkaeufer string
	Land       string
}

// Angebotsfilter engt ein, wonach gefragt wird.
type Angebotsfilter struct {
	// Sprache ist der zweibuchstabige Code — leer heißt: alle.
	Sprache string

	// Finish engt auf Foil oder Nichtfoil ein — leer heißt: alle.
	//
	// **Leer ist nicht `nonfoil`**: `foil=false` fragt nach Nichtfoils,
	// **kein** `foil` nach allem. Ein Nullwert, der als Filter hinausginge,
	// nähme die Hälfte weg, ohne dass jemand danach gefragt hätte.
	Finish domain.Finish
}

// angebotDTO ist die Gestalt, in der CardTrader ein Angebot schickt.
type angebotDTO struct {
	ID          int64  `json:"id"`
	BlueprintID int64  `json:"blueprint_id"`
	Name        string `json:"name_en"`
	Menge       int    `json:"quantity"`
	Preis       struct {
		Cent     int64  `json:"cents"`
		Waehrung string `json:"currency"`
	} `json:"price"`
	Eigenschaften struct {
		Zustand    string `json:"condition"`
		Signiert   bool   `json:"signed"`
		Foil       bool   `json:"mtg_foil"`
		Sprache    string `json:"mtg_language"`
		Veraendert bool   `json:"altered"`
	} `json:"properties_hash"`
	Verkaeufer struct {
		Name string `json:"username"`
		Land string `json:"country_code"`
	} `json:"user"`
	Gradiert bool `json:"graded"`
	ImUrlaub bool `json:"on_vacation"`
	Buendel  int  `json:"bundle_size"`
}

// Angebote gibt, was zu einer Kartenfassung angeboten wird.
//
// **Ohne Blueprint wird nicht gefragt**: der Marktplatz verlangt eines von
// `blueprint_id` und `expansion_id` (0095). Gefragt wird hier immer je
// Blueprint — eine ganze Erweiterung wären zehntausend Angebote für eine
// Karte, nach der jemand gefragt hat.
func (k *Klient) Angebote(ctx context.Context, blueprint int64, filter Angebotsfilter) (
	[]Angebot, []domain.Befund, error,
) {
	if blueprint <= 0 {
		return nil, nil, fmt.Errorf(
			"%w: ohne Blueprint gibt CardTrader keine Angebote — "+
				"gefragt wurde nach %d", domain.ErrValidation, blueprint)
	}

	nachBlueprint := map[string][]angebotDTO{}

	if fehler := k.hole(ctx, wege.Markt, filter.abfrage(blueprint), &nachBlueprint); fehler != nil {
		return nil, nil, fehler
	}

	var (
		angebote []Angebot
		befunde  []domain.Befund
	)

	for _, gruppe := range nachBlueprint {
		for _, dto := range gruppe {
			angebot, befund := alsAngebot(dto)
			if befund != nil {
				befunde = append(befunde, *befund)

				continue
			}

			angebote = append(angebote, angebot)
		}
	}

	return angebote, befunde, nil
}

// abfrage baut die Abfrageparameter — und nur, was gesetzt ist.
func (f Angebotsfilter) abfrage(blueprint int64) url.Values {
	abfrage := url.Values{"blueprint_id": {strconv.FormatInt(blueprint, 10)}}

	if f.Sprache != "" {
		abfrage.Set("language", f.Sprache)
	}

	switch f.Finish {
	case domain.FinishFoil:
		abfrage.Set("foil", "true")
	case domain.FinishNonfoil:
		abfrage.Set("foil", "false")
	case domain.FinishEtched:
		// **CardTrader kennt kein `etched`** — sein Marktplatz kennt `mtg_foil`
		// als Ja-Nein. Wer danach filterte, bekäme eine leere Liste ohne
		// Begründung; gefragt wird deshalb nach allem.
	}

	return abfrage
}

// alsAngebot macht aus dem DTO die neutrale Gestalt.
//
// **Pflichtfelder streng, Zusatzfelder tolerant** (§16): ohne Währung ist der
// Preis keine Aussage, und ein fehlender Verkäufername ist keine.
//
// # Zwei Pflichtfelder, nicht eines
//
// **Ohne Blueprint gehört ein Angebot zu keiner Karte** — Werkstück A19 von
// Phase 5.3
// (#1550),
// gefunden vom Fuzzer. Bis dahin prüfte diese Stelle allein die Währung, und
// ein Eintrag ohne `blueprint_id` kam als vollwertiges Angebot durch: `{0, 0,
// "", 0, "000", …}`. Er wäre über [Angebot.AlsStand] zu einer Preisreihe
// geworden — für einen Druck, nach dem niemand gefragt hat.
//
// **Und es ist der gefragte Blueprint, der zählt**: der Rumpf ordnet die
// Angebote nach Kennung, und was in einer fremden Gruppe steht, beantwortet
// die gestellte Frage nicht. Diese Prüfung hält die schwächere Hälfte davon —
// dass überhaupt eine dasteht.
func alsAngebot(dto angebotDTO) (Angebot, *domain.Befund) {
	if dto.BlueprintID <= 0 {
		return Angebot{}, &domain.Befund{
			Ort: strconv.FormatInt(dto.ID, 10),
			Art: domain.ArtUmgebung,
			Meldung: fmt.Sprintf(
				"CardTrader nennt zum Angebot %d keinen Blueprint. "+
					"Ein Angebot ohne Karte ist keine Aussage über einen "+
					"Preis und wird übergangen.", dto.ID),
		}
	}

	if dto.Preis.Waehrung == "" {
		return Angebot{}, &domain.Befund{
			Datei: strconv.FormatInt(dto.BlueprintID, 10),
			Ort:   strconv.FormatInt(dto.ID, 10),
			Art:   domain.ArtUmgebung,
			Meldung: fmt.Sprintf(
				"CardTrader nennt zum Angebot %d keine Währung. "+
					"Ein Preis ohne sie ist keine Aussage und wird übergangen.",
				dto.ID),
		}
	}

	finish := domain.FinishNonfoil
	if dto.Eigenschaften.Foil {
		finish = domain.FinishFoil
	}

	return Angebot{
		BlueprintID: dto.BlueprintID, ProduktID: dto.ID, Name: dto.Name,
		Preis: dto.Preis.Cent, Waehrung: dto.Preis.Waehrung,
		Menge: dto.Menge, Buendel: dto.Buendel, Finish: finish,
		Zustand: dto.Eigenschaften.Zustand, Sprache: dto.Eigenschaften.Sprache,
		Signiert:   dto.Eigenschaften.Signiert,
		Veraendert: dto.Eigenschaften.Veraendert,
		Gradiert:   dto.Gradiert, ImUrlaub: dto.ImUrlaub,
		Verkaeufer: dto.Verkaeufer.Name, Land: dto.Verkaeufer.Land,
	}, nil
}

// AlsStand macht aus einem Angebot eine Preisbeobachtung.
//
// **Alle sieben Teile einer Reihe** (0093): ein Marktpreis ist eine
// Beobachtung wie jede andere, und ohne Markt, Preisart, Zustand und Sprache
// wäre seine Reihe nicht eindeutig.
//
// **`retail`, weil ein Angebot eine Forderung ist**: der Verkäufer nennt, was
// er **verlangt**. Was CardTrader zahlt, steht hier nicht — es steht hinter
// den Kontofunktionen, und die sind Sache von T11.
func (a Angebot) AlsStand(druck string) domain.Preisstand {
	return domain.Preisstand{
		Printing: druck,
		Quelle:   Familie,
		Markt:    domain.MarktPapier,
		Art:      domain.PreisartVerkauf,
		Wert:     a.Preis,
		Waehrung: a.Waehrung,
		Finish:   a.Finish,
		Zustand:  a.Zustand,
		Sprache:  a.Sprache,
		Lage:     domain.PreisGenannt,
	}
}
