package domain

// **Was in keinem Modus fällt** — Werkstück G4 von Phase 5.5
// (#1461).
//
// # Die Zusage, für die es diesen Typ gibt
//
// **„Engine-Invariante lässt sich in keinem Modus überschreiben"** — das ist
// der erste Abnahmepunkt des Issues, und er ist absolut. Kein Strengegrad,
// keine Hausregel, kein Turnierparameter hebt eine davon auf.
//
// **G1 hat den Satz schon**, für die Strenge: „Sandkasten lockert keine
// Invariante — keine über Objekte, keine über Sichtbarkeit, keine über
// Integrität und keine über Rechte" (#1458). Er stand dort als Kommentar;
// hier wird er ein Wert, den man vorzeigen kann.
//
// # Warum ein Typ und nicht ein Kommentar
//
// **Weil das Issue verlangt, dass es sich nicht umgehen lässt** — „weder über
// UI noch REST/CLI/MCP". Eine Zusage, die nur im Kommentar steht, lässt sich
// von keiner Fläche prüfen; eine Aufzählung lässt sich gegen jede
// Regeländerung halten.
//
// # Und warum es genau diese vier sind
//
// **Weil sie die vier Dinge sind, die eine Partie zusammenhalten.** Fällt
// eine, ist der Rest wertlos: ohne Sichtbarkeit gibt es kein Geheimnis, ohne
// Integrität keinen Verlauf, ohne Rechte keinen Schutz, und ohne die Ordnung
// der Objekte keinen Zustand. Regeln **des Spiels** stehen nicht darunter —
// die entscheidet die [Regelmatrix], und im Sandkasten gelten sie nicht.

import "slices"

// Invariante ist eine Zusage, die kein Modus aufhebt.
type Invariante string

const (
	// InvarianteSichtbarkeit — niemand sieht, was er nicht sehen darf.
	//
	// **Auch der Wirt nicht** (#1458): seine Projektion ist dieselbe wie die
	// jedes anderen Sitzes, und der Vollzustand hat an keiner Fläche einen
	// Weg.
	InvarianteSichtbarkeit Invariante = "sichtbarkeit"

	// InvarianteIntegritaet — der Ereignisstrom wird angehängt, nie
	// überschrieben; jedes Ereignis trägt den Abdruck seines Vorgängers.
	InvarianteIntegritaet Invariante = "integritaet"

	// InvarianteRechte — was eine Rolle nicht darf, tut sie nicht.
	InvarianteRechte Invariante = "rechte"

	// InvarianteObjekte — ein Objekt liegt an genau einer Stelle, und jede
	// Bewegung ist ein Ereignis.
	InvarianteObjekte Invariante = "objekte"
)

// Invarianten gibt alle vier.
func Invarianten() []Invariante {
	return []Invariante{
		InvarianteSichtbarkeit, InvarianteIntegritaet,
		InvarianteRechte, InvarianteObjekte,
	}
}

// IstInvariante sagt, ob ein Wert eine davon ist.
func IstInvariante(gesetz Invariante) bool {
	return slices.Contains(Invarianten(), gesetz)
}
