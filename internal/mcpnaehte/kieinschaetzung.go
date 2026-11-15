package mcpnaehte

// **Die Naht zur Einschätzung** — Werkstück A13.2a von Phase 5.3
// (#1452).
//
// **In einem Zug**, weil diese Fläche die Anwendung je Aufruf baut — der
// Grund steht an [app.App.KIVorschlagInEinemZug]. Für eine Einschätzung ist
// das ohnehin der ganze Weg: sie wird nicht übernommen.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// KIEinschaetzung lässt ein Modell eine Karte beurteilen.
func (p *preise) KIEinschaetzung(
	ctx context.Context, bezug mcp.KIEinschaetzungsbezug,
) (mcp.KIEinschaetzungssicht, error) {
	antwort := p.kianwendung().KIEinschaetzen(ctx,
		app.KIEinschaetzungsbezug{
			Art: bezug.Art, KarteID: bezug.KarteID, Anbieter: bezug.Anbieter,
			Alias: bezug.Alias, Frage: bezug.Frage, Format: bezug.Format,
			Freigegeben: bezug.Freigegeben,
		})
	if antwort.Fehler != nil {
		return mcp.KIEinschaetzungssicht{}, antwort.Fehler.AlsFehler()
	}

	return mcp.KIEinschaetzungssicht{
		Art: antwort.Art, Satz: antwort.Satz, Belegt: antwort.Belegt,
		Vorbehalt:        antwort.Vorbehalt,
		Befunde:          alsBelegbefundsichten(antwort.Befunde),
		Einschraenkungen: antwort.Einschraenkungen,
		Belege:           alsBelegsichten(antwort.Belege),
		Korpusstaende:    alsKorpusstandsichten(antwort.Korpusstaende),
		Anbieter:         antwort.Anbieter, Modell: antwort.Modell,
		Balance: alsBalancesicht(antwort.Balance),
	}, nil
}

// alsBalancesicht übersetzt die Balanceangaben — nil bleibt nil.
//
// **`nil` und nicht ein leerer Verbund**: eine Regeleinschätzung hat keine
// Vergleichskarten, und ein leeres Feld daneben sähe aus, als hätte die Suche
// keine gefunden. Das sind zwei Auskünfte.
func alsBalancesicht(angaben *app.BalanceDTO) *mcp.KIBalancesicht {
	if angaben == nil {
		return nil
	}

	vergleiche := make([]mcp.KIVergleichsicht, 0, len(angaben.Vergleiche))
	for _, eine := range angaben.Vergleiche {
		vergleiche = append(vergleiche, mcp.KIVergleichsicht(eine))
	}

	return &mcp.KIBalancesicht{
		Strategie: angaben.Strategie, Format: angaben.Format,
		Staende:         alsKorpusstandsichten(angaben.Staende),
		Vergleiche:      vergleiche,
		NichtBetrachtet: angaben.NichtBetrachtet,
	}
}

// alsBelegbefundsichten übersetzt die Befunde.
func alsBelegbefundsichten(
	befunde []app.EinschaetzungsbefundDTO,
) []mcp.KIBelegbefundsicht {
	sichten := make([]mcp.KIBelegbefundsicht, 0, len(befunde))
	for _, befund := range befunde {
		sichten = append(sichten, mcp.KIBelegbefundsicht(befund))
	}

	return sichten
}

// alsBelegsichten übersetzt die Belege.
//
// **Nicht jedes Feld reist mit.** Ein Sprachmodell braucht die Kennung, die
// Quelle, den Ausschnitt und das Datum; Kennungssystem, Inhaltshash und
// Zeichenbereich sind Angaben für die Oberfläche, und sie kosten hier
// Tokens, ohne eine Frage zu beantworten.
func alsBelegsichten(belege []app.BelegDTO) []mcp.KIBelegsichtQuelle {
	sichten := make([]mcp.KIBelegsichtQuelle, 0, len(belege))
	for _, beleg := range belege {
		sichten = append(sichten, mcp.KIBelegsichtQuelle{
			Kennung: beleg.Kennung, Quelle: beleg.Quelle, Titel: beleg.Titel,
			Ausschnitt: beleg.Ausschnitt, Regelnummer: beleg.Regelnummer,
			Urheber: beleg.Urheber, Quellrevision: beleg.Quellrevision,
			Quelldatum: beleg.Quelldatum, Abgerufen: beleg.Abgerufen,
			Veraltet: beleg.Veraltet,
		})
	}

	return sichten
}

// alsKorpusstandsichten übersetzt die Stände.
func alsKorpusstandsichten(
	staende []app.KorpusstandDTO,
) []mcp.KIKorpusstandsicht {
	sichten := make([]mcp.KIKorpusstandsicht, 0, len(staende))
	for _, stand := range staende {
		sichten = append(sichten, mcp.KIKorpusstandsicht(stand))
	}

	return sichten
}
