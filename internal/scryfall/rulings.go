package scryfall

import (
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Quelle sagt, wer einen Regelhinweis verfasst hat.
//
// **Zwei Werte, nachgelesen am 2026-08-24** unter
// `https://scryfall.com/docs/api/rulings`:
//
//	source — A computer-readable string indicating which company produced
//	this ruling, either wotc or scryfall.
type Quelle string

const (
	// QuelleWotc ist ein offizieller Hinweis von Wizards of the Coast.
	QuelleWotc Quelle = "wotc"

	// QuelleScryfall ist eine Anmerkung des Scryfall-Teams.
	//
	// **Sie ist nicht offiziell**, und das gehört sichtbar: die Dokumentation
	// sagt, sie erkläre „how the card works in an unofficial format (such as
	// Duel Commander)". Wer beide gleich darstellte, gäbe eine Meinung als
	// Regel aus.
	QuelleScryfall Quelle = "scryfall"
)

// Regelhinweis ist ein einzelner Hinweis zu einer Karte.
type Regelhinweis struct {
	// OrakelID ist die Kennung der Karte, zu der er gehört.
	OrakelID string

	Quelle Quelle

	// Datum bleibt eine Zeichenkette, wie Scryfall es schickt.
	//
	// **Keine Umrechnung.** Ein `time.Time` verlangte eine Zeitzone, und
	// welche gemeint ist, sagt die Dokumentation nicht. Wer sie erfände,
	// verschöbe jedes Datum um bis zu einen Tag — und AGENTS.md §3 nennt
	// genau diesen Fall: „die Wanduhrzeit mit Zone, nicht der UTC-Kalendertag".
	Datum string

	// Text ist der Hinweis selbst.
	Text string
}

// rohhinweis ist die Gestalt, die Scryfall schickt.
type rohhinweis struct {
	Objekt    string `json:"object"`
	OrakelID  string `json:"oracle_id"`
	Quelle    string `json:"source"`
	Datum     string `json:"published_at"`
	Kommentar string `json:"comment"`
}

// rohliste ist der Umschlag, in dem Scryfall Mengen schickt.
//
// **Er heißt bei Scryfall `list`** und trägt die Menge in `data`. Ihn zu
// überspringen und gleich ein Array zu erwarten ginge heute gut und bräche,
// sobald jemand `has_more` auswerten will.
type rohliste[T any] struct {
	Objekt   string `json:"object"`
	Daten    []T    `json:"data"`
	HatMehr  bool   `json:"has_more"`
	Naechste string `json:"next_page"`
}

// LiesRulings liest eine Antwort von `/cards/:id/rulings`.
//
// # Pflichtfelder streng, Zusatzfelder tolerant
//
// **`comment` ist Pflicht**, denn er *ist* der Hinweis; ein leerer wäre eine
// Regelauskunft ohne Inhalt. **`source` ist Pflicht als Feld**, aber ein
// unbekannter Wert ist ein **Befund** und kein Fehler — §16: „Ein unbekannter
// Aufzählungswert wird als Befund geführt und nicht verschluckt."
func LiesRulings(roh []byte) ([]Regelhinweis, []domain.Befund, error) {
	var gelesen rohliste[rohhinweis]

	if fehler := json.Unmarshal(roh, &gelesen); fehler != nil {
		return nil, nil, fmt.Errorf(
			"%w: die Regelhinweise ließen sich nicht lesen: %s",
			domain.ErrValidation, fehler.Error())
	}

	hinweise := make([]Regelhinweis, 0, len(gelesen.Daten))

	var befunde []domain.Befund

	for stelle, einer := range gelesen.Daten {
		if einer.Kommentar == "" {
			return nil, nil, fmt.Errorf(
				"%w: der Regelhinweis an Stelle %d hat keinen Text",
				domain.ErrValidation, stelle)
		}

		quelle, befund := quelleAus(einer.Quelle, stelle)
		if befund != nil {
			befunde = append(befunde, *befund)
		}

		hinweise = append(hinweise, Regelhinweis{
			OrakelID: einer.OrakelID, Quelle: quelle,
			Datum: einer.Datum, Text: einer.Kommentar,
		})
	}

	return hinweise, befunde, nil
}

// quelleAus deutet den Wert von `source`.
//
// **Ein unbekannter Wert reist mit**, statt auf einen der beiden bekannten
// zurückzufallen: wer ihn zu `wotc` machte, gäbe eine unbekannte Anmerkung als
// offizielle Regel aus.
func quelleAus(wert string, stelle int) (Quelle, *domain.Befund) {
	switch Quelle(wert) {
	case QuelleWotc, QuelleScryfall:
		return Quelle(wert), nil
	default:
		return Quelle(wert), &domain.Befund{
			Ort: fmt.Sprintf("Regelhinweis %d", stelle),
			Art: domain.ArtUmgebung,
			Meldung: fmt.Sprintf(
				"Scryfall nennt die unbekannte Quelle %q. Der Hinweis wird angezeigt; "+
					"ob er offiziell ist, lässt sich nicht sagen.", wert),
		}
	}
}
