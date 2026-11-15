package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Auskunftsquelle ist, was die Fassade von der Auskunft braucht.
type Auskunftsquelle interface {
	Regelhinweise(ctx context.Context, name string) (*service.Regelauskunft, error)
	Satz(ctx context.Context, code string) (*service.Satzauskunft, error)
	Symbole(ctx context.Context) (*service.Symbolauskunft, error)
	Manakostenpruefung(ctx context.Context, kosten string) ([]string, error)
}

// MitScryfallauskunft setzt die Quelle — die Naht für Tests.
func (a *App) MitScryfallauskunft(quelle Auskunftsquelle) { a.auskunft = quelle }

// RegelhinweisDTO ist ein Hinweis, wie das Fenster ihn sieht.
type RegelhinweisDTO struct {
	// Quelle ist `wotc` oder `scryfall` — **unübersetzt**, weil sie von außen
	// kommt (§9). Das Fenster entscheidet, wie es sie zeigt.
	Quelle string `json:"quelle"`

	// Offiziell sagt, ob der Hinweis von Wizards of the Coast stammt.
	//
	// **Die Fassade rechnet es aus und nicht das Fenster.** Ein `=== 'wotc'`
	// in TypeScript wäre eine zweite Fassung derselben Regel, und sie liefe
	// auseinander, sobald Scryfall einen dritten Wert einführt.
	Offiziell bool `json:"offiziell"`

	Datum string `json:"datum"`
	Text  string `json:"text"`
}

// RegelauskunftAntwort ist, was das Fenster zu einer Karte bekommt.
type RegelauskunftAntwort struct {
	Kartenname string            `json:"kartenname"`
	Hinweise   []RegelhinweisDTO `json:"hinweise"`
	Adresse    string            `json:"adresse"`

	// Befunde sind die Auffälligkeiten der Antwort — ohne `omitempty`, damit
	// eine leere Liste eine leere Liste bleibt (§3).
	Befunde []string `json:"befunde"`

	AlterStunden float64 `json:"alterStunden"`

	Fehler *Fehler `json:"fehler,omitempty"`
}

// ScryfallRegeln holt die Regelhinweise zu einer Karte.
func (a *App) ScryfallRegeln(ctx context.Context, name string) RegelauskunftAntwort {
	ctx = a.imProfil(ctx)

	if a.auskunft == nil {
		return RegelauskunftAntwort{
			Hinweise: []RegelhinweisDTO{}, Befunde: []string{},
			Fehler: Uebersetze(ctx, fmt.Errorf(
				"%w: der Zugang zu Scryfall ist nicht eingerichtet",
				domain.ErrValidation)),
		}
	}

	geholt, fehler := a.auskunft.Regelhinweise(ctx, name)
	if fehler != nil {
		return RegelauskunftAntwort{
			Hinweise: []RegelhinweisDTO{}, Befunde: []string{},
			Fehler: Uebersetze(ctx, fehler),
		}
	}

	hinweise := make([]RegelhinweisDTO, 0, len(geholt.Hinweise))
	for _, einer := range geholt.Hinweise {
		hinweise = append(hinweise, RegelhinweisDTO{
			Quelle: string(einer.Quelle), Offiziell: einer.Quelle == scryfallOffiziell,
			Datum: einer.Datum, Text: einer.Text,
		})
	}

	return RegelauskunftAntwort{
		Kartenname: geholt.Kartenname, Hinweise: hinweise, Adresse: geholt.Adresse,
		Befunde: alsMeldungen(geholt.Befunde), AlterStunden: geholt.Alter.Hours(),
	}
}

// ManakostenAntwort sagt, welche Symbole eines Kostens unbekannt sind.
type ManakostenAntwort struct {
	// Unbekannt ist leer, wenn der Kosten in Ordnung ist — **ohne
	// `omitempty`** (§3).
	Unbekannt []string `json:"unbekannt"`

	Fehler *Fehler `json:"fehler,omitempty"`
}

// ScryfallManakosten prüft einen Manakosten gegen die Symbologie.
func (a *App) ScryfallManakosten(ctx context.Context, kosten string) ManakostenAntwort {
	ctx = a.imProfil(ctx)

	if a.auskunft == nil {
		return ManakostenAntwort{Unbekannt: []string{}, Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: der Zugang zu Scryfall ist nicht eingerichtet", domain.ErrValidation))}
	}

	unbekannt, fehler := a.auskunft.Manakostenpruefung(ctx, kosten)
	if fehler != nil {
		return ManakostenAntwort{Unbekannt: []string{}, Fehler: Uebersetze(ctx, fehler)}
	}

	if unbekannt == nil {
		unbekannt = []string{}
	}

	return ManakostenAntwort{Unbekannt: nichtNil(unbekannt)}
}

// alsMeldungen macht aus Befunden eine Liste von Sätzen.
//
// **Nie `nil`**: eine leere Liste verschwände sonst aus der Antwort, und im
// Fenster stünde `undefined` statt einer leeren Liste — die Falle aus §3.
func alsMeldungen(befunde []domain.Befund) []string {
	meldungen := make([]string, 0, len(befunde))
	for _, einer := range befunde {
		meldungen = append(meldungen, einer.String())
	}

	return meldungen
}

// scryfallOffiziell ist die Quelle, die von Wizards of the Coast stammt.
const scryfallOffiziell = "wotc"
