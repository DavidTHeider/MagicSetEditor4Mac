package domain

// **Woher eine Regel stammt** — Werkstück G4 von Phase 5.5
// (#1461).
//
// # Warum eine Regel ihre Quelle mitführt
//
// **Weil das Issue es verlangt**: „Jede automatisch durchgesetzte Regel nennt
// aktuelle Regelquelle und eigene positive/negative Grenzfälle." Eine Regel
// ohne Quelle ist eine Behauptung — und sie sieht genauso aus wie eine
// belegte.
//
// **Und weil die Fassung mitgeht.** Die Comprehensive Rules wachsen mit jedem
// Set; wer nur „CR 704" schreibt, sagt nicht, welche 704 gemeint war. §11
// verlangt beides: Nummer **und** Fassung.
//
// # Was hier ausdrücklich nicht steht
//
// **Kein Wortlaut.** Die CR ist urheberrechtlich geschützt und trägt keine
// Weitergabelizenz; eine Regel wird **beschrieben** und mit ihrer Nummer
// belegt, nicht abgeschrieben (§11). Dieser Typ hält deshalb Nummern und
// keine Sätze.

import (
	"fmt"
	"strings"
)

// Regeldokument ist das Werk, aus dem eine Regel stammt.
type Regeldokument string

const (
	// DokumentCR sind die **Comprehensive Rules** — was am Tisch gilt.
	DokumentCR Regeldokument = "CR"

	// DokumentMTR sind die **Magic Tournament Rules** — was ein Turnier
	// daraus macht.
	DokumentMTR Regeldokument = "MTR"
)

// Regeldokumente gibt beide.
func Regeldokumente() []Regeldokument {
	return []Regeldokument{DokumentCR, DokumentMTR}
}

// IstRegeldokument sagt, ob ein Wert eines der beiden ist.
func IstRegeldokument(dokument Regeldokument) bool {
	for _, gilt := range Regeldokumente() {
		if gilt == dokument {
			return true
		}
	}

	return false
}

// Regelquelle nennt Dokument, Fassung und die Nummern.
type Regelquelle struct {
	Dokument Regeldokument

	// Fassung ist das Datum der Ausgabe — `2026-08-19`.
	Fassung string

	// Nummern sind die Abschnitte, auf die sich eine Aussage stützt.
	//
	// **Abschnitte und keine Sätze**: `400` für die Zonen, `704` für die
	// zustandsbasierten Aktionen. Eine feinere Nummer steht dort, wo eine
	// einzelne Regel umgesetzt ist, im Kommentar neben dem Code.
	Nummern []string
}

// Pruefe sagt, ob diese Quelle etwas belegt.
func (q Regelquelle) Pruefe() error {
	if !IstRegeldokument(q.Dokument) {
		return fmt.Errorf("%w: %q ist kein Regeldokument — es gibt CR und MTR",
			ErrValidation, q.Dokument)
	}

	if strings.TrimSpace(q.Fassung) == "" {
		return fmt.Errorf(
			"%w: eine Regelquelle ohne Fassung belegt nichts — die CR wächst "+
				"mit jedem Set", ErrValidation)
	}

	if len(q.Nummern) == 0 {
		return fmt.Errorf(
			"%w: eine Regelquelle ohne Nummer ist eine Behauptung",
			ErrValidation)
	}

	for _, nummer := range q.Nummern {
		if strings.TrimSpace(nummer) == "" {
			return fmt.Errorf("%w: eine leere Regelnummer", ErrValidation)
		}
	}

	return nil
}

// String gibt die Quelle in der Form, die auch im Quelltext steht.
func (q Regelquelle) String() string {
	return fmt.Sprintf("%s %s (Fassung %s)",
		q.Dokument, strings.Join(q.Nummern, ", "), q.Fassung)
}
