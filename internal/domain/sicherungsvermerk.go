package domain

import (
	"regexp"
	"slices"
	"strings"
	"time"
)

// Sicherungsvermerk hält fest, dass der Bestand gesichert wurde — M01
// (#2345),
// Befunde A-06 und A-12 des Audits 0.9.0.
//
// # Was er festhält und was nicht
//
// **Dass es eine Sicherung gab**, wann, warum, über welche Fläche, und woran
// sie wiederzuerkennen ist — Name, Größe, Prüfsumme. **Nicht, wo sie heute
// liegt**: eine Sicherung ist eine Datei am Ort, den jemand wählt, und die
// Anwendung räumt sie nie weg (`docs/SICHERUNG.md`). Eine Zusage über ihren
// Verbleib könnte niemand halten.
//
// **Der Name und kein Pfad.** Der Bestand wird weitergegeben — darum liegen
// Zugangsdaten im Schlüsselbund und nicht hier —, und ein Pfad trüge den Namen
// des Benutzerkontos mit.
type Sicherungsvermerk struct {
	// Gesichert ist der Augenblick, in dem die Sicherung geschrieben **und**
	// geprüft war.
	Gesichert time.Time

	Anlass  Sicherungsanlass
	Flaeche Flaeche

	// Datei ist der Name der Sicherung, ohne Ordner.
	Datei string

	Groesse int64

	// Pruefsumme ist SHA-256 über die Datei, wie sie geschrieben wurde — klein
	// und hexadezimal.
	Pruefsumme string

	// Verschluesselt sagt, ob die Datei mit einem Kennwort verschlüsselt ist
	// (`age`, #1723); Größe und Prüfsumme gelten dann für die verschlüsselte.
	Verschluesselt bool
}

// Sicherungsanlass sagt, warum gesichert wurde.
type Sicherungsanlass string

// Die Anlässe einer Sicherung.
const (
	// SicherungVonHand hat jemand verlangt — im Fenster, über `mse`, MCP oder
	// REST.
	SicherungVonHand Sicherungsanlass = "hand"

	// SicherungNachPlan hat der eingerichtete Plan geschrieben (OPS01, #2347).
	SicherungNachPlan Sicherungsanlass = "plan"

	// SicherungVorWanderung steht vor einer Wanderung des Schemas
	// (`internal/app/sicherungvormigration.go`, #939).
	SicherungVorWanderung Sicherungsanlass = "wanderung"
)

// sicherungsanlaesse sind alle Anlässe, die es gibt.
var sicherungsanlaesse = []Sicherungsanlass{SicherungVonHand, SicherungNachPlan, SicherungVorWanderung}

// sicherungsflaechen sind die Flächen, über die gesichert werden kann.
var sicherungsflaechen = []Flaeche{FlaecheFenster, FlaecheCLI, FlaecheMCP, FlaecheREST}

// sha256hex ist eine Prüfsumme, wie [Sicherungsvermerk.Pruefsumme] sie trägt.
var sha256hex = regexp.MustCompile(`^[0-9a-f]{64}$`)

// Pruefe urteilt über einen Vermerk.
//
// **Ein Vermerk entsteht nie aus einer Eingabe**, nur aus einer geglückten
// Sicherung dieser Anwendung. Eine Abweisung ist deshalb ein Programmfehler und
// trägt den allgemeinen Schlüssel — ein eigener Satz in acht Sprachen wäre
// einer, den niemand je liest.
func (v Sicherungsvermerk) Pruefe() error {
	switch {
	case v.Gesichert.IsZero():
		return NewFieldError("gesichert", SchluesselValidierung, "Ein Sicherungsvermerk braucht seinen Zeitpunkt.")
	case !slices.Contains(sicherungsanlaesse, v.Anlass):
		return NewFieldError("anlass", SchluesselValidierung, "%q ist kein Anlass einer Sicherung.", v.Anlass)
	case !slices.Contains(sicherungsflaechen, v.Flaeche):
		return NewFieldError("flaeche", SchluesselValidierung, "%q ist keine Fläche, über die gesichert wird.", v.Flaeche)
	case v.Datei == "" || strings.ContainsAny(v.Datei, `/\`):
		return NewFieldError("datei", SchluesselValidierung,
			"Ein Sicherungsvermerk trägt den Namen der Datei, ohne Ordner — nicht %q.", v.Datei)
	case v.Groesse <= 0:
		return NewFieldError("groesse", SchluesselValidierung, "Eine Sicherung ist nie leer.")
	case !sha256hex.MatchString(v.Pruefsumme):
		return NewFieldError("pruefsumme", SchluesselValidierung,
			"Die Prüfsumme einer Sicherung ist SHA-256, klein und hexadezimal.")
	}

	return nil
}
