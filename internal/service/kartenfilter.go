// **Der Schnellfilter an der Kartenliste** — Werkstück U5 von Phase 5.1
// (#1416).
//
// MSE2 hält die gefilterte Liste als eigene Sicht auf dieselben Karten
// (`FilteredCardList::getItems`, `src/gui/control/filtered_card_list.cpp:26–30`,
// Fassung f9d9356 vom 2021-07-06); der Filter selbst steht in der Domäne
// ([domain.SchnellfilterTrifft]). Hier wird er über den Bestand gelegt, und
// heraus kommen **Kennungen** — die Liste, die das Fenster schon hat, bleibt
// stehen, und was nicht passt, wird darin verborgen.
//
// **In der Reihenfolge des Bestands**, nicht der eingestellten Sortierung:
// wer die Kennungen gegen eine Liste hält, braucht die Menge, nicht die Folge.

package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Filtern gibt die Kennungen der Karten eines Sets, die durch den
// Schnellfilter kommen. Eine leere Anfrage lässt alle durch.
//
// **Ein Set, das es nicht gibt, ist ein Fehler und keine leere Liste**: der
// Bestand gibt für eine fremde Kennung null Karten, und null Treffer sähen
// aus wie ein Filter, der nichts findet.
func (k *Kartenlistenbau) Filtern(ctx context.Context, setID int64, anfrage string) ([]int64, error) {
	if err := k.setMussEsGeben(ctx, setID); err != nil {
		return nil, err
	}

	karten, err := k.karten.Karten(ctx, setID)
	if err != nil {
		return nil, err
	}

	teile := domain.ZerlegeSchnellfilter(anfrage)
	kennungen := make([]int64, 0, len(karten))

	// **Ohne Anfrage wird nichts gerechnet** — eine leere lässt alle durch.
	if len(teile) == 0 {
		for _, karte := range karten {
			kennungen = append(kennungen, karte.ID)
		}

		return kennungen, nil
	}

	rechner := k.feldrechnerFuer(ctx, setID)

	for _, karte := range karten {
		treffer := rechner.trifft(ctx, teile, karte)

		// **Ein Abbruch ist kein Nichttreffer** (Befund 3 des Zweitblicks
		// über Phase G): die Rechnung hört auf, und was fehlt, läse sich als
		// „gerechnet und nicht gefunden" — die Suche gäbe nichts, ihre
		// Verneinung alles.
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if treffer {
			kennungen = append(kennungen, karte.ID)
		}
	}

	return kennungen, nil
}

// feldrechner rechnet, was die Liste zeigt, für den Schnellfilter — FP-01
// (#2155).
//
// # Der Vertrag
//
// **Was gespeichert ist, gilt; wo nichts steht, gilt der gerechnete Wert** —
// dieselbe Regel wie in der Liste ([gerechneteZellen]). MSE2 durchsucht in
// `Card::contains` jeden Wert, wie er nach den Skripten dasteht
// (`src/data/card.cpp`, Zeilen 55–61, Fassung f9d9356 vom 2021-07-06);
// `type` aus `magic.mse-game` wird aus Ober- und Untertyp gerechnet, und wer
// „Creature — Beast" sucht, sucht, was er sieht.
//
//   - **Ein freier Begriff** sieht die gerechneten Werte der **sichtbaren
//     Spalten** — das, was die Liste zeigt.
//   - **Ein genanntes Feld** (`type:…`) wird gerechnet, auch wenn es keine
//     Spalte ist; genannt heißt, wie beim Gespeicherten, dass der Name den
//     Bezug enthält.
//   - **Gespeichert schlägt gerechnet**, anders als in MSE2, wo das Skript den
//     Wert immer neu setzt: der Filter soll nicht anderes finden, als dasteht.
//
// # Was es kostet
//
// **Gerechnet wird nur, was die gespeicherten Werte nicht schon
// entscheiden** — gemessen am 2026-10-07 an „Black" (3.788 Karten,
// `magic.mse-game` mit 123 Skriptfeldern): alle Skriptfelder jeder Karte zu
// rechnen kostete 7,5 s je Anfrage, die Liste selbst 0,45 s.
type feldrechner struct {
	spiel        *msescript.Umgebung
	felder       []domain.Field
	skriptfelder []domain.Field
	sichtbar     map[string]bool
}

// feldrechnerFuer baut ihn — oder nil, wenn es nichts zu rechnen gibt: ohne
// Vorlage oder ohne ein Feld mit Wertskript bleibt es beim Gespeicherten.
func (k *Kartenlistenbau) feldrechnerFuer(ctx context.Context, setID int64) *feldrechner {
	felder, ohneVorlage := k.felderOderKeine(ctx, setID)

	skriptfelder := make([]domain.Field, 0, len(felder))
	for _, feld := range felder {
		if strings.TrimSpace(feld.Skript) != "" {
			skriptfelder = append(skriptfelder, feld)
		}
	}

	if ohneVorlage || len(skriptfelder) == 0 {
		return nil
	}

	// **Die sichtbaren Spalten wie in der Liste** — ohne lesbare Wahl die
	// Vorgaben des Spiels.
	spiel, _ := k.spielVonSet(ctx, setID)     //nolint:errcheck // das Set ist oben schon geprüft
	gewaehlt, _ := k.wahl.Spalten(ctx, spiel) //nolint:errcheck // ohne Wahl gelten die Vorgaben

	sichtbar := map[string]bool{}
	for _, spalte := range spaltenAus(felder, gewaehlt) {
		if spalte.Sichtbar {
			sichtbar[spalte.Feld] = true
		}
	}

	skript, setfelder := "", []domain.Field(nil)
	if k.spielskript != nil {
		//nolint:errcheck // ohne Skript wird gerechnet, was ohne geht — wie in [Kartenlistenbau.Aufbau]
		skript, _ = k.spielskript.Spielskript(ctx, setID)
		setfelder, _ = k.spielskript.Setfelder(ctx, setID) //nolint:errcheck // dito
	}

	return &feldrechner{
		// **Einmal je Anfrage, nicht je Karte** — dieselbe Rechnung wie die Liste.
		spiel: spielumgebung(ctx, skript, setfelder, felder,
			Wortliste{}, &Schluesselwortverwendung{}),
		felder: felder, skriptfelder: skriptfelder, sichtbar: sichtbar,
	}
}

// trifft sagt, ob eine Karte durch den Filter kommt — mit den gerechneten
// Werten, wo die gespeicherten nicht entscheiden.
func (r *feldrechner) trifft(ctx context.Context, teile []domain.Filterteil, karte *domain.CardRow) bool {
	gerechnet := map[string]string{}

	for _, teil := range teile {
		suche := []domain.Filterteil{{Typ: teil.Typ, Wert: teil.Wert, Soll: true}}

		enthalten := domain.SchnellfilterTrifft(suche, *karte) ||
			(r != nil && domain.SchnellfilterTrifft(suche, r.gerechnete(ctx, karte, teil, gerechnet)))

		if enthalten != teil.Soll {
			return false
		}
	}

	return true
}

// gerechnete gibt die gerechneten Werte, die dieser Teil sehen darf — und
// rechnet nur, was noch fehlt.
func (r *feldrechner) gerechnete(
	ctx context.Context, karte *domain.CardRow, teil domain.Filterteil, gerechnet map[string]string,
) domain.CardRow {
	werte := make(map[string]string, len(karte.Values))
	for _, wert := range karte.Values {
		werte[wert.Field] = domain.OhneAuszeichnung(wert.Text)
	}

	offen := make([]domain.Listenspalte, 0, len(r.skriptfelder))
	gesehen := make([]string, 0, len(r.skriptfelder))

	for _, feld := range r.skriptfelder {
		gemeint := r.sichtbar[feld.Name]
		if teil.Typ != "" {
			gemeint = strings.Contains(strings.ToLower(feld.Name), strings.ToLower(teil.Typ))
		}

		if !gemeint || werte[feld.Name] != "" {
			continue
		}

		gesehen = append(gesehen, feld.Name)

		if _, schon := gerechnet[feld.Name]; !schon {
			offen = append(offen, domain.Listenspalte{Feld: feld.Name, Sichtbar: true})
		}
	}

	if len(offen) > 0 {
		gerechneteZellen(ctx, r.spiel, karte, r.felder, offen, werte)

		for _, spalte := range offen {
			gerechnet[spalte.Feld] = werte[spalte.Feld]
		}
	}

	nur := domain.CardRow{ID: karte.ID, SetID: karte.SetID}
	for _, name := range gesehen {
		nur.Values = append(nur.Values, domain.Value{Field: name, Text: gerechnet[name]})
	}

	return nur
}

// setMussEsGeben schlägt die Kennung in den Setköpfen nach.
func (k *Kartenlistenbau) setMussEsGeben(ctx context.Context, setID int64) error {
	koepfe, err := k.koepfe.Setkoepfe(ctx)
	if err != nil {
		return err
	}

	for _, kopf := range koepfe {
		if kopf.ID == setID {
			return nil
		}
	}

	return fmt.Errorf("%w: das Set %d gibt es nicht", domain.ErrNotFound, setID)
}
