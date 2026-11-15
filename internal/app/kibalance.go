package app

// **Die Balanceanalyse** — Werkstück A13.2b von Phase 5.3
// (#1452).
//
// # Was hier steht und was in `kieinschaetzung.go`
//
// **Dort der gemeinsame Weg** — Kontextpaket, Modell, Belegprüfung. Hier das,
// was **nur** eine Balanceaussage braucht: die Vergleichskarten, ihre
// Unterschiede, und die Angaben, die das **Programm** macht und nicht das
// Modell.
//
// # Warum Strategie, Format und Stand vom Programm kommen
//
// **Weil das Modell sie nicht weiß.** Es bekommt Ausschnitte; **wie** sie
// ausgewählt wurden — welcher Fall, welche Korpora, welcher Modus, wie viele
// Kandidaten — weiß der Dienst, der sie zusammengestellt hat. Ein Modell, das
// seine eigene Suchstrategie beschriebe, beschriebe eine erfundene.
//
// **Das Issue verlangt sie als „nachvollziehbar"**, und nachvollziehbar ist
// eine Angabe nur, wenn die Stelle sie macht, die sie kennt.

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// KIBalanceanweisung steht vor einer Frage nach dem Vergleich.
//
// **Sie verlangt den Unterschied und verbietet die Note.** „Ähnlich" ist
// keine Aussage — das sagt der Rang schon, der die Karte gefunden hat, und er
// bedeutet nichts ([0103](../../docs/entscheidungen/0103-wie-zwei-trefferlisten-zusammenkommen.md)).
const KIBalanceanweisung = "Du vergleichst eine Sammelkarte mit den " +
	"mitgegebenen Karten. Nenne je Vergleich den Beleg in eckigen Klammern " +
	"und mindestens einen KONKRETEN Unterschied — Manawert, Effekt, " +
	"Zusatzfähigkeit. Sag NICHT, eine Karte sei »stärker« oder vergib eine " +
	"Note: dafür gibt es keinen Maßstab. Sag, was anders ist, und was du " +
	"dabei nicht betrachtet hast."

// VergleichDTO ist eine Karte, gegen die verglichen wurde.
type VergleichDTO struct {
	Beleg        string   `json:"beleg"`
	Name         string   `json:"name"`
	Unterschiede []string `json:"unterschiede"`
}

// BalanceDTO trägt, was nur eine Balanceaussage hat.
//
// **Ohne ein Zahlenfeld**, und das ist die Zusage: „kein objektiver
// Balancewert aus Ähnlichkeitsscore" steht nicht als Prüfung da, sondern als
// fehlendes Feld ([domain.Balancefelder]).
type BalanceDTO struct {
	// Strategie sagt, **wie** die Vergleichskarten gefunden wurden.
	Strategie string `json:"strategie"`

	// Format ist, worauf sich der Vergleich bezieht — leer heißt: keines
	// genannt, und dann steht das in den Einschränkungen.
	Format string `json:"format"`

	// Staende sagen, wie alt die Quellen sind.
	Staende []KorpusstandDTO `json:"staende"`

	Vergleiche      []VergleichDTO `json:"vergleiche"`
	NichtBetrachtet []string       `json:"nichtBetrachtet"`
}

// balanceantwort ist die Form, in der das Modell antwortet.
type balanceantwort struct {
	Einschaetzung    string   `json:"einschaetzung"`
	Einschraenkungen []string `json:"einschraenkungen"`
	Vergleiche       []struct {
		Beleg        string   `json:"beleg"`
		Name         string   `json:"name"`
		Unterschiede []string `json:"unterschiede"`
	} `json:"vergleiche"`
	NichtBetrachtet []string `json:"nichtBetrachtet"`
}

// balanceschema zwingt die Antwort in die Form.
func balanceschema() map[string]any {
	text := map[string]any{"type": "string"}
	textliste := map[string]any{"type": "array", "items": text}

	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"einschaetzung":    text,
			"einschraenkungen": textliste,
			"vergleiche": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"beleg": text, "name": text, "unterschiede": textliste,
					},
					"required":             []string{"beleg", "unterschiede"},
					"additionalProperties": false,
				},
			},
			"nichtBetrachtet": textliste,
		},
		"required":             []string{"einschaetzung", "vergleiche"},
		"additionalProperties": false,
	}
}

// alsBalanceangaben bringt die Modellantwort in die Gestalt von `domain`.
func alsBalanceangaben(gelesen balanceantwort) domain.Balanceangaben {
	vergleiche := make([]domain.Vergleichskarte, 0, len(gelesen.Vergleiche))

	for _, roh := range gelesen.Vergleiche {
		vergleiche = append(vergleiche, domain.Vergleichskarte{
			Beleg: domain.Belegkennung(strings.Trim(roh.Beleg, "[]")),
			Name:  roh.Name, Unterschiede: roh.Unterschiede,
		})
	}

	return domain.Balanceangaben{
		Vergleiche: vergleiche, NichtBetrachtet: gelesen.NichtBetrachtet,
	}
}

// alsBalanceDTO baut die Antwort ans Fenster.
func alsBalanceDTO(
	angaben domain.Balanceangaben, paket service.Kontextpaket,
	fall domain.Recherchefall, format string,
) BalanceDTO {
	vergleiche := make([]VergleichDTO, 0, len(angaben.Vergleiche))

	for _, eine := range angaben.Vergleiche {
		vergleiche = append(vergleiche, VergleichDTO{
			Beleg: string(eine.Beleg), Name: eine.Name,
			Unterschiede: alsListe(eine.Unterschiede),
		})
	}

	return BalanceDTO{
		Strategie:       suchstrategie(paket, fall),
		Format:          format,
		Staende:         alsKorpusstandDTOs(paket.Manifest.Korpusstaende),
		Vergleiche:      vergleiche,
		NichtBetrachtet: alsListe(angaben.NichtBetrachtet),
	}
}

// suchstrategie beschreibt, wie die Vergleichskarten gefunden wurden.
//
// # Warum sie aus dem Ergebnis kommt und nicht aus einer Konstante
//
// **Weil sie sonst behauptete, was gewollt war, statt zu sagen, was geschah.**
// Ein Index ohne Einbettungsmodell sucht **lexikalisch**, und eine Strategie,
// die „hybrid" behauptete, wäre dann falsch — und zwar genau dort, wo es
// darauf ankommt.
func suchstrategie(
	paket service.Kontextpaket, fall domain.Recherchefall,
) string {
	modus := paket.Ergebnis.Modus
	if modus == "" {
		modus = "keiner"
	}

	strategie := "Ähnlichkeitssuche über die Quellen des Falls »" +
		string(fall) + "«, Modus " + modus +
		", höchstens " + strconv.Itoa(KIEinschaetzungstreffer) +
		" Treffer, davon höchstens " + strconv.Itoa(KIEinschaetzungsbelege) +
		" als Beleg"

	// **Und der Reranker steht dabei, weil er fehlt.** A15 hat gemessen, dass
	// er ohne belegten Zusatznutzen „ein Modell mehr im Weg" wäre
	// (`docs/RETRIEVALMESSUNG.md`); wer die Strategie liest, soll das wissen
	// und nicht annehmen, es sei eines im Spiel.
	return strategie + "; kein Reranker"
}

// balanceMit fährt den Balanceweg mit einem fertigen Werkzeugsatz.
//
// **Ein eigener Weg und kein Schalter im gemeinsamen**: die Antwortform ist
// eine andere, das Schema ist ein anderes, und die Anweisung auch. Ein `if`
// mitten im Regelhinweis machte beide Wege schwerer zu lesen als zwei.
func (a *App) balanceMit(
	ctx context.Context, lauf einschaetzungslauf, satz kiwerkzeugsatz,
) KIEinschaetzungAntwort {
	roh, fehler := satz.quelle.Vorschlagen(ctx, KIBalanceanweisung,
		lauf.bezug.Alias, auftragMitBelegen(lauf.paket, balancefrage(lauf.bezug)),
		balanceschema())
	if fehler != nil {
		return mitFehler(ctx, fehler)
	}

	gelesen, fehler := balanceAus(roh)
	if fehler != nil {
		return mitFehler(ctx, fehler)
	}

	angaben := alsBalanceangaben(gelesen)
	belege := lauf.paket.Manifest.Belege

	geprueft := domain.PruefeBalance(gelesen.Einschaetzung, angaben, belege)

	antwort := lauf.antwortAus(geprueft,
		mitFormathinweis(alsListe(gelesen.Einschraenkungen), lauf.bezug.Format),
		roh.Modell)

	balance := alsBalanceDTO(angaben, lauf.paket, fallFuer(lauf.art),
		lauf.bezug.Format)
	antwort.Balance = &balance

	return antwort
}

// balancefrage stellt dem Modell die Frage samt Format.
//
// **Das Format steht in der Frage und nicht in der Anweisung**: es gehört zu
// dieser Frage und nicht zu jeder.
func balancefrage(bezug KIEinschaetzungsbezug) string {
	if strings.TrimSpace(bezug.Format) == "" {
		return bezug.Frage
	}

	return bezug.Frage + "\n\nFormat: " + bezug.Format
}

// mitFormathinweis trägt nach, dass kein Format genannt wurde.
//
// **Eine Einschränkung und kein Befund.** Ein Vergleich ohne Format ist
// möglich — er gilt dann für keines besonders, und **das** ist die Angabe,
// die das Issue verlangt („Format/Stand … angeben").
func mitFormathinweis(einschraenkungen []string, format string) []string {
	if strings.TrimSpace(format) != "" {
		return einschraenkungen
	}

	return append(einschraenkungen,
		"Es wurde kein Format genannt — der Vergleich gilt für keines "+
			"besonders, und was in einem Format erlaubt ist, entscheidet die "+
			"eingebaute Formatprüfung und kein Modell.")
}

// balanceAus liest die Antwort des Modells.
func balanceAus(roh KIRohantwort) (balanceantwort, error) {
	if roh.Abgeschnitten {
		return balanceantwort{}, fmt.Errorf(
			"%w: die Antwort des Modells ist abgeschnitten", domain.ErrValidation)
	}

	var gelesen balanceantwort
	if fehler := json.Unmarshal([]byte(roh.Inhalt), &gelesen); fehler != nil {
		return balanceantwort{}, fmt.Errorf(
			"%w: die Antwort des Modells ist kein JSON: %s",
			domain.ErrValidation, fehler.Error())
	}

	if strings.TrimSpace(gelesen.Einschaetzung) == "" {
		return balanceantwort{}, fmt.Errorf(
			"%w: das Modell hat nichts gesagt", domain.ErrValidation)
	}

	return gelesen, nil
}
