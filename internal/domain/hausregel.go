package domain

// **Was eine Runde für sich vereinbaren darf** — Werkstück G4 von Phase 5.5
// (#1461).
//
// # Die sechs, und warum es genau diese sind
//
// **Das Issue zählt sie auf**: „Hausregeln OpenHands/Startleben/freier
// Mulligan/Custom Cards/Proxy/Verbote nur innerhalb freigegebener Grenzen."
// Es ist eine geschlossene Liste — was nicht darauf steht, ist keine
// Hausregel, sondern eine Änderung am Programm.
//
// # Warum offene Hände keine Invariante verletzen
//
// **Weil sie ändern, was jemand sehen *darf*, und nicht, ob es durchgesetzt
// wird.** [InvarianteSichtbarkeit] sagt: niemand sieht, was er nicht sehen
// darf. Wer offene Hände vereinbart, **darf** sehen — die Durchsetzung bleibt
// dieselbe, nur die Erlaubnis ist eine andere.
//
// **Der Unterschied ist der ganze Punkt.** Eine Hausregel, die die
// Durchsetzung abschaltete, wäre kein Zugeständnis, sondern ein Loch: dann
// sähe man auch, was niemand vereinbart hat.

import (
	"fmt"
	"slices"
	"strings"
)

// Hausregelart ist eine der sechs vereinbarbaren Abweichungen.
type Hausregelart string

const (
	// HausregelOffeneHaende — alle spielen mit offenen Karten.
	HausregelOffeneHaende Hausregelart = "offene-haende"

	// HausregelStartleben — eine andere Zahl als die des Formats.
	HausregelStartleben Hausregelart = "startleben"

	// HausregelFreierMulligan — die erste Hand kostet nichts.
	HausregelFreierMulligan Hausregelart = "freier-mulligan"

	// HausregelEigeneKarten — selbst entworfene Karten sind erlaubt.
	HausregelEigeneKarten Hausregelart = "eigene-karten"

	// HausregelProxy — ein Zettel vertritt eine Karte.
	HausregelProxy Hausregelart = "proxy"

	// HausregelVerbote — zusätzliche Verbote über das Format hinaus.
	HausregelVerbote Hausregelart = "verbote"
)

// Hausregelarten gibt alle sechs.
func Hausregelarten() []Hausregelart {
	return []Hausregelart{
		HausregelOffeneHaende, HausregelStartleben, HausregelFreierMulligan,
		HausregelEigeneKarten, HausregelProxy, HausregelVerbote,
	}
}

// IstHausregelart sagt, ob ein Wert eine der sechs ist.
func IstHausregelart(art Hausregelart) bool {
	return slices.Contains(Hausregelarten(), art)
}

// Hausregel ist eine vereinbarte Abweichung samt ihrem Wert.
type Hausregel struct {
	Art Hausregelart

	// Wert ist, was vereinbart wurde — `25` für das Startleben, leer für
	// eine Abweichung, die nur an oder aus kennt.
	Wert string
}

// Pruefe sagt, ob diese Hausregel überhaupt eine ist.
func (h Hausregel) Pruefe() error {
	if !IstHausregelart(h.Art) {
		return fmt.Errorf(
			"%w: %q ist keine Hausregel — es gibt %s",
			ErrValidation, h.Art, hausregelnamen())
	}

	// **Das Startleben ist die einzige mit einer Zahl.** Eine Abweichung
	// „offene Hände: 7" wäre eine Angabe, die niemand deutet — und still
	// verschluckt wäre sie schlimmer als abgewiesen.
	if h.Art == HausregelStartleben && strings.TrimSpace(h.Wert) == "" {
		return fmt.Errorf(
			"%w: ein abweichendes Startleben ohne Zahl ist keine Abweichung",
			ErrValidation)
	}

	return nil
}

// hausregelnamen zählt die Arten für eine Meldung auf.
func hausregelnamen() string {
	namen := make([]string, 0, len(Hausregelarten()))
	for _, art := range Hausregelarten() {
		namen = append(namen, string(art))
	}

	return strings.Join(namen, ", ")
}
