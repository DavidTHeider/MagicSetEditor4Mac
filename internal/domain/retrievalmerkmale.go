package domain

// **Die strukturierten Merkmale eines Dokuments** — Werkstück A15 von
// Phase 5.3
// (#1514).
//
// # Warum sie neben dem Text stehen und nicht darin
//
// **Weil ein Filter deterministisch sein muss.** Das Issue: „Strukturierte
// Filter deterministisch: Farbe getrennt von Farbidentität, Typ, Manawert,
// Set, Sprache, Format mit Datenstand. Unbekannter Wert ist kein erfundener
// Treffer. Querytexte ersetzen keine geprüfte Filterstruktur." Wer „rote
// Kreatur" aus dem Fließtext liest, rät; wer `Farben` und `Typen` liest,
// weiß.
//
// # Farbe und Farbidentität sind zwei Dinge
//
// **Und sie fallen regelmäßig auseinander.** Die Farbe einer Karte folgt
// ihren Manakosten und ihren Farbanzeigern; die Farbidentität zählt
// zusätzlich jedes Manasymbol im Regeltext (Regel 903.4). Ein
// Rückstoß-Effekt mit `{R}` im Text ist farblos und rot-identisch — wer
// beides zusammenlegte, machte jede Kommandeur-Frage falsch.
//
// # Und warum ein Format seinen Stand mitbringt
//
// **Weil Legalität ein Datum hat.** Eine Karte ist in Modern legal, bis sie
// es nicht mehr ist; eine Antwort ohne Stand behauptet einen aktuellen
// Bannstand, den niemand geprüft hat. Das Issue verlangt „Format mit
// Datenstand".

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// Retrievalmerkmale sind die filterbaren Angaben eines Dokuments.
//
// **Leere Felder heißen „unbekannt" und nicht „keines".** Eine Regelstelle
// hat keine Farbe; ein Filter nach Farbe darf sie deshalb nicht treffen, und
// auch nicht „weil nichts dagegen spricht".
type Retrievalmerkmale struct {
	// Farben sind die Farben der Karte selbst (`W`, `U`, `B`, `R`, `G`).
	Farben []string

	// Farbidentitaet ist die Identität nach Regel 903.4 — sie umfasst die
	// Farben und kann darüber hinausgehen.
	Farbidentitaet []string

	// Typen sind die Typzeilenwörter in Kleinschreibung
	// (`creature`, `instant`, `legendary`, `goblin`).
	Typen []string

	// Manawert ist der konvertierte Manawert — `nil` heißt: keiner.
	//
	// **Ein Zeiger und keine Null.** Ein Land hat den Manawert 0, eine
	// Regelstelle gar keinen; ohne die Unterscheidung fände ein Filter
	// `Manawert = 0` jede Regel des Regelwerks.
	Manawert *float64

	// Set ist der Setcode des Drucks (`neo`), leer bei Eigenem ohne Set.
	Set string

	// Formate ordnet einem Format seinen Stand zu (`commander` → `legal`).
	Formate map[string]string

	// Formatstand ist die Revision, aus der die Legalitäten stammen.
	//
	// **Ohne ihn gilt keine Legalität.** [Retrievalmerkmale.Format] gibt
	// deshalb nichts heraus, solange er leer ist.
	Formatstand string
}

// dieFarben sind die fünf, in ihrer Reihenfolge.
var dieFarben = []string{"W", "U", "B", "R", "G"}

// Pruefe sagt, ob die Merkmale in sich stimmig sind.
//
// **Die Farbidentität umfasst die Farben.** Das ist keine Konvention, sondern
// Regel 903.4: was in den Manakosten steht, steht auch in der Identität. Eine
// Karte mit rotem Manasymbol und grüner Identität gibt es nicht.
func (m Retrievalmerkmale) Pruefe() error {
	for _, farbe := range append(slices.Clone(m.Farben), m.Farbidentitaet...) {
		if !slices.Contains(dieFarben, farbe) {
			return fmt.Errorf("%w: %q ist keine Farbe — erwartet wird eines "+
				"aus %s", ErrValidation, farbe, strings.Join(dieFarben, ", "))
		}
	}

	for _, farbe := range m.Farben {
		if !slices.Contains(m.Farbidentitaet, farbe) {
			return fmt.Errorf(
				"%w: %q steht in den Farben und nicht in der Farbidentität — "+
					"nach Regel 903.4 umfasst die Identität die Farben",
				ErrValidation, farbe)
		}
	}

	if len(m.Formate) > 0 && strings.TrimSpace(m.Formatstand) == "" {
		return fmt.Errorf(
			"%w: %d Formatangaben ohne Datenstand — eine Legalität ohne Stand "+
				"behauptet einen Bannstand, den niemand geprüft hat",
			ErrValidation, len(m.Formate))
	}

	return nil
}

// Format gibt die Legalität in einem Format — und den Stand, aus dem sie stammt.
//
// **Beides oder nichts.** Ohne Stand gibt es keine Auskunft; das ist die
// Zusage aus dem Kopf dieser Datei, an der Stelle eingelöst, an der sie
// gebraucht wird (§10).
func (m Retrievalmerkmale) Format(name string) (lage, stand string, bekannt bool) {
	if strings.TrimSpace(m.Formatstand) == "" {
		return "", "", false
	}

	lage, bekannt = m.Formate[strings.ToLower(name)]
	if !bekannt {
		return "", "", false
	}

	return lage, m.Formatstand, true
}

// Leer sagt, ob gar keine Merkmale vorliegen.
func (m Retrievalmerkmale) Leer() bool {
	return len(m.Farben) == 0 && len(m.Farbidentitaet) == 0 &&
		len(m.Typen) == 0 && m.Manawert == nil && m.Set == "" &&
		len(m.Formate) == 0
}

// Kopie gibt eine eigene Fassung — Slices und Map mitgezählt.
//
// **Damit ein Dokument seine Merkmale nicht teilt.** Wer sie durchreichte,
// gäbe jedem Leser Schreibzugriff auf den Index.
func (m Retrievalmerkmale) Kopie() Retrievalmerkmale {
	kopie := m
	kopie.Farben = slices.Clone(m.Farben)
	kopie.Farbidentitaet = slices.Clone(m.Farbidentitaet)
	kopie.Typen = slices.Clone(m.Typen)

	if m.Manawert != nil {
		wert := *m.Manawert
		kopie.Manawert = &wert
	}

	if m.Formate != nil {
		kopie.Formate = maps.Clone(m.Formate)
	}

	return kopie
}
