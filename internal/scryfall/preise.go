// **Die Preise einer Karte** — Werkstück 3 von Phase 3.4
// (#1165).
//
// # Was die Quelle sagt, nachgelesen am 2026-08-27
//
// [scryfall.com/docs/api/cards](https://scryfall.com/docs/api/cards):
//
//	prices — An object containing daily price information for this card,
//	including usd, usd_foil, usd_etched, eur, eur_foil, eur_etched, and tix
//	prices, as strings.
//
// Und im Fußbereich derselben Seite:
//
//	Card prices … represent daily estimates and/or market values provided by
//	our affiliates. Absolutely no guarantee is made for any price information.
//	See stores for final prices and details.
//
// **Drei Dinge folgen daraus**, und alle drei stehen im Code:
//
//   - **Als Zeichenkette gelesen.** Die Umrechnung in die kleinste Einheit macht
//     [domain.MarktpreisAus], und sie weist ab statt zu runden.
//   - **Ein fehlender Preis ist kein Fehler.** Die Felder sind `null`, wenn es
//     keinen gibt — bei den meisten Karten für `usd_etched` und `tix`.
//   - **Ein unlesbarer Preis ist ein Befund** und bricht nichts (§8, §16). Was
//     sich nicht deuten lässt, wird benannt und weggelassen; die übrigen
//     kommen an.
//
// # Und was die Quelle sagt, ist nicht, was sie schickt
//
// Die Dokumentation nennt **sieben** Felder, `eur_etched` darunter. **Kein
// heute abgerufener Druck hat es.** Nachgemessen am 2026-08-27 an acht
// Drucken — die fünf von »Abundant Harvest« sowie `sta/48`, `khm/276`,
// `znr/385`, darunter drei mit `etched` in `finishes`: jede Antwort trug
// dieselben **sechs** Schlüssel `eur`, `eur_foil`, `tix`, `usd`, `usd_etched`,
// `usd_foil`.
//
// **Die Reichweite dieser Messung sind acht Antworten** (§10) — nicht das
// Scryfall-Korpus und nicht die Zukunft. Deshalb bleibt `eur_etched` gelesen:
// taucht es auf, kommt es an. Was **nicht** stehen bleibt, ist die Behauptung,
// es sei je beobachtet worden. Angestrichen vom Nachlauf über Phase 3.4 als
// Befund 4, und er hat recht mit dem Kern: eine Zusage aus der Dokumentation
// ist eine Zusage der Dokumentation.

package scryfall

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Preis ist ein Preis einer Karte, wie Scryfall ihn führt.
type Preis struct {
	// Waehrung ist der Code, unter dem die Quelle ihn führt — `USD`, `EUR`
	// oder `TIX`.
	Waehrung string

	Finish domain.Finish

	// Markt ist, wo dieser Preis gilt (A3, #1442) — `tix` ist Magic Online,
	// alles andere Papier.
	Markt domain.Preismarkt

	// Betrag ist der Preis in der kleinsten Einheit seiner Währung.
	Betrag int64
}

// AlsStand macht daraus den neutralen Stand, den die Anwendung führt.
//
// **Die Zuordnung steht hier und nicht im Dienst** (§16, dritte Schicht,
// A3/#1442): was Scryfalls Felder **bedeuten**, weiß dieses Paket. Der Dienst
// wusste es bis A3 auch — er schrieb `PreisquelleScryfall` fest in seinen Code
// —, und damit hätte eine zweite Quelle ihn geändert statt nur ergänzt.
//
// **`retail` ohne Alternative**: Scryfall nennt keinen Ankaufspreis. Was es
// nicht gibt, wird nicht erfunden.
//
// **`Beobachtet` bleibt leer**: Scryfall nennt keinen Tag („daily estimates
// … absolutely no guarantee", nachgelesen am 2026-08-27). Einen aus dem
// Abrufzeitpunkt abzuleiten wäre eine Genauigkeit, die die Quelle nicht
// hergibt (§16, Satz 6).
func (p Preis) AlsStand(printing string) domain.Preisstand {
	return domain.Preisstand{
		Printing: printing,
		Quelle:   domain.PreisquelleScryfall,
		Markt:    p.Markt,
		Art:      domain.PreisartVerkauf,
		Waehrung: p.Waehrung,
		Finish:   p.Finish,
		Lage:     domain.PreisGenannt,
		Wert:     p.Betrag,
	}
}

// rohpreise ist der `prices`-Verbund.
//
// **Jedes Feld ein Zeiger**: `null` und `"0.00"` sind zwei verschiedene
// Aussagen — „kein Preis bekannt" und „kostet nichts". Ein `string` machte aus
// beidem die leere Zeichenkette.
type rohpreise struct {
	USD       *string `json:"usd"`
	USDFoil   *string `json:"usd_foil"`
	USDEtched *string `json:"usd_etched"`
	EUR       *string `json:"eur"`
	EURFoil   *string `json:"eur_foil"`
	EUREtched *string `json:"eur_etched"`
	Tix       *string `json:"tix"`
}

// preisantwort ist der Teil einer Kartenantwort, der die Preise trägt.
type preisantwort struct {
	Kennung string     `json:"id"`
	Preise  *rohpreise `json:"prices"`
}

// preisfelder ordnet jedes Feld seiner Währung und Ausführung zu.
//
// **Die Zuordnung steht an einer Stelle** (§16, dritte Schicht). Über den Code
// verteilt liefe sie beim ersten neuen Feld auseinander — und genau das ist
// der Fehler, den `internal/mtgjson` schon einmal vermieden hat.
func preisfelder(roh *rohpreise) []struct {
	name     string
	text     *string
	waehrung string
	finish   domain.Finish
	markt    domain.Preismarkt
} {
	return []struct {
		name     string
		text     *string
		waehrung string
		finish   domain.Finish
		markt    domain.Preismarkt
	}{
		{"usd", roh.USD, "USD", domain.FinishNonfoil, domain.MarktPapier},
		{"usd_foil", roh.USDFoil, "USD", domain.FinishFoil, domain.MarktPapier},
		{"usd_etched", roh.USDEtched, "USD", domain.FinishEtched, domain.MarktPapier},
		{"eur", roh.EUR, "EUR", domain.FinishNonfoil, domain.MarktPapier},
		{"eur_foil", roh.EURFoil, "EUR", domain.FinishFoil, domain.MarktPapier},
		{"eur_etched", roh.EUREtched, "EUR", domain.FinishEtched, domain.MarktPapier},

		// **`tix` ist keine Währung eines Landes**, sondern eine Eintrittskarte
		// für Magic Online. Sie steht trotzdem hier: was die Quelle führt, wird
		// geführt — und der Code sagt, was er ist.
		//
		// **Und sie ist der einzige Onlinepreis dieser Quelle** (A3, #1442) —
		// deshalb steht der Markt jetzt in derselben Zeile wie die Währung
		// und nicht in einer Regel daneben.
		{"tix", roh.Tix, "TIX", domain.FinishNonfoil, domain.MarktOnline},
	}
}

// LiesPreise liest die Preise aus einer Kartenantwort.
//
// **Der zweite Rückgabewert ist die Kennung des Drucks.** Sie gehört zu jedem
// Preis, und wer sie getrennt herausfischen müsste, läse dieselbe Antwort
// zweimal.
func LiesPreise(roh []byte) (string, []Preis, []domain.Befund, error) {
	kennung, preise, _, befunde, fehler := liesPreise(roh)

	return kennung, preise, befunde, fehler
}

// liesPreise liest dazu die Reihen, die Scryfall nennt und die sich nicht
// deuten lassen — F04
// (#2246).
// **Sie sind kein Wegfall**: [Klient.PreiseFuerDruck] gibt sie als
// [domain.PreisUngeprueft], und der letzte bekannte Wert bleibt gültig.
func liesPreise(roh []byte) (
	kennung string, preise, unlesbar []Preis, befunde []domain.Befund,
	fehler error,
) {
	var gelesen preisantwort
	if err := json.Unmarshal(roh, &gelesen); err != nil {
		return "", nil, nil, nil, domain.MitGrund(domain.ErrValidation, err,
			"die Antwort von Scryfall ist kein gültiges JSON")
	}

	// **Streng bei der Kennung** (§16): ohne sie ist nicht zu sagen, zu welchem
	// Druck die Preise gehören — und ein Preisstand ohne Druck ist keiner.
	if gelesen.Kennung == "" {
		return "", nil, nil, nil, fmt.Errorf(
			"%w: die Antwort nennt kein Feld id", domain.ErrValidation)
	}

	// **Kein `prices` ist kein Fehler.** Ein Token oder eine Karte ohne
	// Marktdaten hat keines, und das ist eine Auskunft und kein Fehlschlag.
	if gelesen.Preise == nil {
		return gelesen.Kennung, []Preis{}, nil, nil, nil
	}

	preise = []Preis{}

	for _, feld := range preisfelder(gelesen.Preise) {
		if feld.text == nil {
			continue
		}

		betrag, err := domain.MarktpreisAus(*feld.text, feld.waehrung)
		if err != nil {
			befunde = append(befunde, domain.Befund{
				Datei: gelesen.Kennung,
				Ort:   feld.name,
				Art:   domain.ArtUmgebung,
				Meldung: fmt.Sprintf(
					"Scryfall meldet für %s den Preis %q, der nicht zu deuten ist: %v. "+
						"Die übrigen Preise werden übernommen.",
					feld.name, *feld.text, err),
			})
			unlesbar = append(unlesbar, Preis{
				Waehrung: feld.waehrung, Finish: feld.finish, Markt: feld.markt,
			})

			continue
		}

		preise = append(preise, Preis{
			Waehrung: feld.waehrung, Finish: feld.finish,
			Markt: feld.markt, Betrag: betrag,
		})
	}

	return gelesen.Kennung, preise, unlesbar, befunde, nil
}

// PreiseFuerDruck holt die Preise eines Drucks über seine Kennung.
//
// # Warum ein eigener Weg und nicht `KarteNachName`
//
// Ein Name führt auf **irgendeinen** Druck; ein Preis gehört zu **einem**.
// Scryfall hat dafür `/cards/:id`, nachgelesen am 2026-08-27.
//
// **Der Zwischenspeicher gilt auch hier**, samt seinem Alter: „Ohne Netz
// arbeitet die Anwendung weiter" (§16, Satz 5), und der letzte gute Stand sagt,
// wie alt er ist.
// **Und er gibt neutrale Stände** (A3, #1442): was Scryfalls Felder bedeuten,
// weiß dieses Paket — der Dienst dahinter soll es nicht wissen müssen.
func (k *Klient) PreiseFuerDruck(ctx context.Context, kennung string) (
	staende []domain.Preisstand, alter time.Duration,
	befunde []domain.Befund, fehler error,
) {
	// **Die Kennung wird geglättet, und die Adresse kommt von dort.**
	//
	// Die Regel gilt für jeden Weg, der `/cards/:id` fragt — sie steht
	// deshalb in [Klient.druckadresse].
	kennung, schluessel, adresse, fehler := k.druckadresse(kennung)
	if fehler != nil {
		return nil, 0, nil, fehler
	}

	roh, alter, fehler := k.speicher.Hole(ctx, schluessel, func() ([]byte, error) {
		return k.hole(ctx, adresse, "diesen Druck")
	})
	if fehler != nil {
		return nil, 0, nil, fehler
	}

	geantwortet, preise, unlesbar, befunde, lesefehler := liesPreise(roh)
	if lesefehler != nil {
		return nil, 0, nil, lesefehler
	}

	// **Die Antwort muss den Druck nennen, nach dem gefragt wurde.**
	//
	// `LiesPreise` verlangt eine Kennung; der erste Anlauf warf sie danach weg
	// und legte die Preise unter der **angefragten** ab. Eine vertauschte oder
	// beschädigte Antwort aus dem Zwischenspeicher wäre damit der Preis eines
	// anderen Drucks geworden — still. Angestrichen vom Zweitblick über
	// Phase 3.4.
	if !strings.EqualFold(geantwortet, kennung) {
		return nil, 0, befunde, fmt.Errorf(
			"%w: gefragt war der Druck %q, geantwortet hat %q",
			domain.ErrConflict, kennung, geantwortet)
	}

	staende = make([]domain.Preisstand, 0, len(preise)+len(unlesbar))
	for _, preis := range preise {
		staende = append(staende, preis.AlsStand(kennung))
	}

	for _, preis := range unlesbar {
		stand := preis.AlsStand(kennung)
		stand.Lage, stand.Wert = domain.PreisUngeprueft, 0
		staende = append(staende, stand)
	}

	return staende, alter, befunde, nil
}

// Name sagt, wer hier antwortet (A3, #1442).
//
// **Damit ein leerer Abruf trotzdem eine Aussage ist**: „Scryfall nennt für
// diesen Druck nichts mehr" lässt sich aus null Ständen nicht ablesen.
func (k *Klient) Name() domain.Preisquelle { return domain.PreisquelleScryfall }
