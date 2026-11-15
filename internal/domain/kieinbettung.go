package domain

// **Der Einbettungs- und Neuordnungsvertrag** — Werkstück A6 von Phase 5.3
// (#1445).
//
// # Was hier geprüft wird und was nicht
//
// **Die Form und niemals der Inhalt.** Das Issue sagt es für das Schema, und
// es gilt hier genauso: „Schema validiert Form/Felder, niemals Wahrheitsgehalt
// oder Spielbalance einer KI-Aussage." Ob ein Vektor die richtige Bedeutung
// trägt, kann niemand prüfen; ob er die zugesagte Länge hat und aus Zahlen
// besteht, schon.
//
// # Warum Frage und Dokument zwei Anweisungen sind
//
// **Weil viele Modelle es so wollen.** Derselbe Text, einmal als Frage und
// einmal als Fundstelle eingeleitet, ergibt zwei verschiedene Vektoren — und
// das ist der Zweck. Wer beides gleich behandelte, bekäme systematisch
// danebenliegende Ergebnisse, **ohne dass etwas scheitert**; das Issue
// verlangt deshalb ausdrücklich „getrennte Query-/Dokument-Embeddingkonfiguration".

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// KIPooling sagt, wie aus vielen Zeichenvektoren einer wird.
type KIPooling string

// Die drei, die Laufzeiten üblicherweise anbieten.
const (
	// KIPoolingMittel mittelt über alle Zeichen.
	KIPoolingMittel KIPooling = "mittel"

	// KIPoolingErster nimmt den ersten — das `[CLS]`-Zeichen.
	KIPoolingErster KIPooling = "erster"

	// KIPoolingLetzter nimmt den letzten.
	KIPoolingLetzter KIPooling = "letzter"
)

var kiPoolings = []KIPooling{KIPoolingMittel, KIPoolingErster, KIPoolingLetzter}

// KIEinbettungsvertrag ist, was ein Einbettungsmodell zusagt.
//
// **Er gehört zum Modell und nicht zum Index**: A15 entscheidet daran, ob neu
// gebaut werden muss, und A6 liefert ihn — eine Rückabhängigkeit gibt es
// nicht.
type KIEinbettungsvertrag struct {
	// Dimension ist die Länge jedes Vektors.
	Dimension int

	Pooling KIPooling

	// Normalisiert sagt, ob die Vektoren auf Länge eins gebracht sind.
	//
	// **Es ändert das Ergebnis jedes Vergleichs** — deshalb steht es im
	// Fingerabdruck und nicht in einer Fußnote.
	Normalisiert bool

	// Zerteiler nennt den Tokenizer samt seiner Herkunft.
	Zerteiler string

	// Frageanweisung und Dokumentanweisung werden dem Text vorangestellt —
	// leer heißt: **keine**, nicht „ein leerer Vorspann".
	Frageanweisung    string
	Dokumentanweisung string
}

// Pruefe sagt, ob der Vertrag vollständig ist.
func (v KIEinbettungsvertrag) Pruefe() error {
	if v.Dimension <= 0 {
		return fmt.Errorf("%w: ein Einbettungsvertrag nennt eine Dimension "+
			"größer null, dieser nennt %d", ErrValidation, v.Dimension)
	}

	if !kennt(kiPoolings, v.Pooling) {
		return fmt.Errorf("%w: %q ist keine bekannte Zusammenfassung",
			ErrValidation, string(v.Pooling))
	}

	if v.Zerteiler == "" {
		return fmt.Errorf("%w: ein Einbettungsvertrag nennt seinen Zerteiler — "+
			"ohne ihn ist nicht zu sagen, was gleich behandelt wird", ErrValidation)
	}

	return nil
}

// PruefeVektor sagt, ob ein Vektor zu diesem Vertrag passt.
//
// **Länge und Zahlen, mehr nicht.** Was der Vektor bedeutet, steht hier nicht
// zur Debatte — und was er **nicht** sein darf, schon: ein `NaN` steckt an.
// Ein einziges davon macht jedes Skalarprodukt damit zu `NaN`, und eine
// Ähnlichkeitssuche gäbe danach **stumm** keine Treffer mehr.
func (v KIEinbettungsvertrag) PruefeVektor(vektor []float32) error {
	if len(vektor) != v.Dimension {
		return fmt.Errorf("%w: der Vektor hat %d Stellen, der Vertrag sagt %d "+
			"zu — ein Vektor anderer Länge kommt aus einem anderen Modell",
			ErrValidation, len(vektor), v.Dimension)
	}

	for stelle, wert := range vektor {
		if math.IsNaN(float64(wert)) || math.IsInf(float64(wert), 0) {
			return fmt.Errorf("%w: an Stelle %d steht keine Zahl (%v) — "+
				"ein solcher Wert steckt jeden Vergleich an",
				ErrValidation, stelle, wert)
		}
	}

	return nil
}

// FuerFrage stellt die Frageanweisung voran.
func (v KIEinbettungsvertrag) FuerFrage(text string) string {
	return v.Frageanweisung + text
}

// FuerDokument stellt die Dokumentanweisung voran.
func (v KIEinbettungsvertrag) FuerDokument(text string) string {
	return v.Dokumentanweisung + text
}

// Fingerabdruck ist die Kennung dieses Vertrags.
//
// **Jeder Teil geht ein**, und zwar weil jeder das Ergebnis verschiebt: eine
// andere Dimension, eine andere Zusammenfassung, eine andere Normalisierung,
// ein anderer Zerteiler, eine andere Anweisung. A15 entscheidet daran über
// eine Neuindizierung — ein Teil, der fehlte, ließe einen Index stehen, der
// nicht mehr zu den Fragen passt.
func (v KIEinbettungsvertrag) Fingerabdruck() string {
	teile := strings.Join([]string{
		strconv.Itoa(v.Dimension),
		string(v.Pooling),
		strconv.FormatBool(v.Normalisiert),
		v.Zerteiler,
		v.Frageanweisung,
		v.Dokumentanweisung,
	}, "\x1f")

	summe := sha256.Sum256([]byte(teile))

	return hex.EncodeToString(summe[:])
}

// KIBewertung ist, was eine Neuordnung zu einem Kandidaten sagt.
//
// **Eine Zahl und keine Reihenfolge**: wer aus Zahlen eine Liste macht,
// entscheidet über Schwellen, und das ist eine Frage des Aufrufers.
type KIBewertung struct {
	Kandidat string
	Wert     float64
}

// PruefeBewertungen sagt, ob eine Neuordnung zu ihren Kandidaten passt.
//
// **Mehr Bewertungen als Kandidaten sind eine erfundene Aussage** — jemand
// hat etwas bewertet, das niemand geschickt hat. Das Issue nennt es als
// Abnahme: „überzählige Vektoren … abweisen".
func PruefeBewertungen(bewertungen []KIBewertung, kandidaten int) error {
	if len(bewertungen) > kandidaten {
		return fmt.Errorf("%w: %d Bewertungen zu %d Kandidaten — eine davon "+
			"gehört zu nichts", ErrValidation, len(bewertungen), kandidaten)
	}

	gesehen := map[string]bool{}

	for _, bewertung := range bewertungen {
		if bewertung.Kandidat == "" {
			return fmt.Errorf("%w: eine Bewertung nennt ihren Kandidaten nicht",
				ErrValidation)
		}

		if gesehen[bewertung.Kandidat] {
			return fmt.Errorf("%w: %q ist zweimal bewertet — welche Zahl gilt, "+
				"entscheidet dieses Programm nicht", ErrValidation, bewertung.Kandidat)
		}

		gesehen[bewertung.Kandidat] = true

		if math.IsNaN(bewertung.Wert) || math.IsInf(bewertung.Wert, 0) {
			return fmt.Errorf("%w: die Bewertung von %q ist keine Zahl (%v)",
				ErrValidation, bewertung.Kandidat, bewertung.Wert)
		}
	}

	return nil
}
