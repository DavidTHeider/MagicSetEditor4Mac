// **Woher ein Deck und woher ein Exemplar kam, an der Fassade** — Werkstück 12
// von Phase 4.12
// (#1351),
// nach Vermerk [0071](../../docs/entscheidungen/0071-woher-ein-deck-kommt.md).
//
// # Warum es keine eigene Bindung dafür gibt
//
// **Weil die Herkunft mit dem Ding reist, dessen Herkunft sie ist.** Eine
// Bindung `Herkunft(art, kennung)` wäre ein zweiter Ruf je Datensatz — und
// jede Fläche, die ihn vergisst, zeigte ein übernommenes Deck als eigenes.
// Sie steht deshalb in [DeckDTO] und in [ExemplarDTO], und der `LEFT JOIN`
// darunter sorgt dafür, dass sie nichts kostet.

package app

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

// HerkunftDTO sagt, woher ein Datensatz kam.
//
// **Ein leeres `art` heißt: hier entstanden.** Das Fenster prüft darauf und
// zeigt sonst nichts an — eine Zeile „Herkunft: —" wäre eine Auskunft über
// nichts.
type HerkunftDTO struct {
	// Art ist `profil` oder `einfuhr` — leer heißt: keine.
	Art string `json:"art"`

	// Quelle ist die lesbare Bezeichnung, abgeschrieben zum Zeitpunkt der
	// Übernahme.
	Quelle string `json:"quelle"`

	// QuellProfil ist leer, wenn es keines gab — **nicht "0"**: eine Null sähe
	// aus wie eine Kennung.
	QuellProfil string `json:"quellProfil"`

	Uebernommen string `json:"uebernommen"`
}

// alsHerkunft bildet sie ab.
func alsHerkunft(herkunft domain.Herkunft) HerkunftDTO {
	return HerkunftDTO{
		Art:         string(herkunft.Art),
		Quelle:      herkunft.Quelle,
		QuellProfil: kennungtext(herkunft.QuellProfil),
		Uebernommen: herkunft.Uebernommen,
	}
}

// Deckuebernahmeantwort trägt das übernommene Deck samt dem, was nicht mitkam.
//
// **Die Befunde stehen neben dem Deck und nicht in `Fehler`**: das Deck ist
// da. Was nicht mitkonnte, ist eine Auskunft und keine Absage (§8).
//
// **`Deckuebernahmeantwort` und nicht `UebernahmeAntwort`**: den Namen trägt
// seit Phase 4.10 die Paketübernahme aus einer MSE2-Installation.
type Deckuebernahmeantwort struct {
	Deck    *DeckDTO `json:"deck"`
	Befunde []string `json:"befunde"`
	Fehler  *Fehler  `json:"fehler"`
}
