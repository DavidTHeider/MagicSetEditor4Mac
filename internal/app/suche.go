package app

import (
	"context"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Sucher ist, was `app` zum Suchen braucht.
type Sucher interface {
	Suche(ctx context.Context, frage domain.SearchQuery) (domain.SearchResult, error)
	Kartenwerte(ctx context.Context, karten []int64) (map[int64][]domain.Value, error)
	Felder(ctx context.Context, setID int64) ([]string, error)

	// Setkoepfe sagt, ob es ein Set gibt
	// (#1336).
	//
	// **Weil `--set=999` sonst „Kein Treffer" gäbe** — dieselbe Antwort wie
	// eine Suche, die nichts findet. Das Issue stellt die Frage ausdrücklich:
	// ist `--set=` ein Filter, der nichts trifft, oder eine Kennung, die keine
	// ist? Sie ist eine Kennung: wer eine Zahl angibt, meint ein bestimmtes
	// Set, und ein Tippfehler darin ist kein leeres Ergebnis.
	Setkoepfe(ctx context.Context) ([]domain.SetHead, error)
}

// TrefferDTO ist eine gefundene Stelle, wie das Frontend sie sieht.
type TrefferDTO struct {
	KarteID string `json:"karteID"`
	SetID   string `json:"setID"`
	Nummer  int    `json:"nummer"`

	// Kartenname ist der Name der Karte — **nicht** der getroffene Wert. Eine
	// Trefferliste, die nur die Fundstelle zeigt, lässt einen raten, zu welcher
	// Karte sie gehört.
	Kartenname string `json:"kartenname"`

	Feld string `json:"feld"`
	Wert string `json:"wert"`
}

// SuchAntwort trägt Treffer oder einen Fehler.
type SuchAntwort struct {
	Treffer []TrefferDTO `json:"treffer"`

	// Gekuerzt sagt, dass es mehr gibt, als hier steht. **Sichtbar**, weil eine
	// stille Kürzung wie ein vollständiges Ergebnis aussieht.
	Gekuerzt bool `json:"gekuerzt"`

	Fehler *Fehler `json:"fehler"`
}

// Suchen findet Kartenstellen über den Volltextindex.
func (a *App) Suchen(ctx context.Context, text, feld, setID string) SuchAntwort {
	ctx = a.imProfil(ctx)

	leer := SuchAntwort{Treffer: []TrefferDTO{}}

	frage := domain.SearchQuery{Text: text, Field: feld}

	if setID != "" {
		kennung, befund := kennungAus(setID)
		if befund != nil {
			leer.Fehler = befund

			return leer
		}

		if fehler := service.SetDaSein(ctx, a.sucher, kennung); fehler != nil {
			leer.Fehler = Uebersetze(ctx, fehler)

			return leer
		}

		frage.SetID = kennung
	}

	gefunden, err := a.sucher.Suche(ctx, frage)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	namen, err := a.kartennamen(ctx, gefunden.Hits)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	treffer := make([]TrefferDTO, 0, len(gefunden.Hits))
	for _, fund := range gefunden.Hits {
		treffer = append(treffer, TrefferDTO{
			KarteID:    strconv.FormatInt(fund.CardID, 10),
			SetID:      strconv.FormatInt(fund.SetID, 10),
			Nummer:     fund.Number,
			Kartenname: namen[fund.CardID],
			Feld:       fund.Field,
			Wert:       fund.Value,
		})
	}

	return SuchAntwort{Treffer: treffer, Gekuerzt: gefunden.Truncated}
}

// kartennamen holt zu jedem Treffer den Kartennamen.
//
// **In einer Abfrage für alle Treffer.** Bei 200 Treffern wären es sonst 200
// Wege zur Datenbank, und die Suche fühlte sich langsamer an als der Durchlauf,
// den sie ersetzt.
func (a *App) kartennamen(ctx context.Context, treffer []domain.SearchHit) (
	map[int64]string, error,
) {
	if len(treffer) == 0 {
		return map[int64]string{}, nil
	}

	gesehen := map[int64]bool{}
	kennungen := make([]int64, 0, len(treffer))

	for _, fund := range treffer {
		if !gesehen[fund.CardID] {
			gesehen[fund.CardID] = true
			kennungen = append(kennungen, fund.CardID)
		}
	}

	werte, err := a.sucher.Kartenwerte(ctx, kennungen)
	if err != nil {
		return nil, err
	}

	namen := make(map[int64]string, len(werte))

	for karte, felder := range werte {
		for _, wert := range felder {
			if wert.Field == "name" {
				namen[karte] = wert.Text

				break
			}
		}
	}

	return namen, nil
}

// FelderAntwort trägt die Feldnamen oder einen Fehler.
type FelderAntwort struct {
	Felder []string `json:"felder"`
	Fehler *Fehler  `json:"fehler"`
}

// Suchfelder gibt die Feldnamen, nach denen sich filtern lässt.
func (a *App) Suchfelder(ctx context.Context, setID string) FelderAntwort {
	ctx = a.imProfil(ctx)

	leer := FelderAntwort{Felder: []string{}}

	var kennung int64

	if setID != "" {
		gelesen, befund := kennungAus(setID)
		if befund != nil {
			leer.Fehler = befund

			return leer
		}

		kennung = gelesen
	}

	felder, err := a.sucher.Felder(ctx, kennung)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	// **Über den vorbereiteten leeren Ausschnitt**, nicht als neuer Verbund: ein
	// `nil` aus dem Speicher wird sonst zu `null` im JSON, und das Frontend
	// bekommt statt einer leeren Liste einen Wert, den es nicht durchlaufen kann
	// (§3, „Zwischen den Prüfebenen").
	leer.Felder = append(leer.Felder, felder...)

	return leer
}

// MitSucher setzt die Naht zur Suche.
func (a *App) MitSucher(sucher Sucher) { a.sucher = sucher }
