// **Rechte, Rollen und Sichtbarkeit über Nutzer hinweg** — Werkstück 11 von
// Phase 4.12
// (#1351).
//
// # Was heute geht und nicht gehen sollte
//
// Ein Profil gehört einem Nutzer, und jede Wurzel gehört einem Profil
// (Wanderung 0035). **Wer in ein fremdes Profil wechseln will, wird bisher nur
// nach dem zweiten Faktor seines Eigners gefragt** — hat der keinen, kommt
// jeder hinein.
//
// # Drei Rollen, und die erste heißt weniger, als das Wort verspricht
//
// **`leser` heißt „sieht, dass es das Profil gibt"** — nicht „sieht, was darin
// steht". Der Grund steht in
// [0070](../../docs/entscheidungen/0070-rechte-rollen-und-sichtbarkeit.md):
// ein Nur-Lese-Modus **innerhalb** eines Profils bräuchte einen lückenlosen
// Schreibschutz, und den gibt es nur im Treiber. **Am 2026-09-06 gezählt**,
// und die Reichweite gehört zur Zahl (§10): eine Zusage, die an 116
// Bindungen einzeln hängt, bricht beim ersten vergessenen Ruf still (§10).

package domain

import "fmt"

// Rolle sagt, was jemand mit einem fremden Profil darf.
type Rolle string

// Die drei Rollen, aufsteigend.
const (
	// RolleLeser sieht das Profil in der Übersicht — Name, Beschreibung,
	// Eigner. **Er kommt nicht hinein.**
	RolleLeser Rolle = "leser"

	// RolleMitarbeiter wechselt hinein und arbeitet darin.
	RolleMitarbeiter Rolle = "mitarbeiter"

	// RolleVerwalter vergibt und zieht Freigaben zurück.
	RolleVerwalter Rolle = "verwalter"
)

// rangfolge ist die Aufstiegsordnung — jede Rolle schließt die darunter ein.
var rangfolge = map[Rolle]int{
	RolleLeser: 1, RolleMitarbeiter: 2, RolleVerwalter: 3,
}

// Rollen gibt alle drei, aufsteigend.
func Rollen() []Rolle {
	return []Rolle{RolleLeser, RolleMitarbeiter, RolleVerwalter}
}

// Reicht sagt, ob diese Rolle mindestens die verlangte ist.
//
// **Aufsteigend und nicht als Menge**: drei Stufen, und jede schließt die
// darunter ein. Eine Aufzählung von Einzelrechten wollte gepflegt sein, und
// ein ungepflegter Eintrag in einer Rechteliste ist gefährlicher als keiner —
// dieselbe Überlegung wie beim Scope eines Zugriffstokens
// ([0068](../../docs/entscheidungen/0068-ein-token-ist-kein-kennwort.md)).
func (r Rolle) Reicht(verlangt Rolle) bool {
	return rangfolge[r] >= rangfolge[verlangt] && rangfolge[verlangt] > 0
}

// Pruefe weist eine Rolle ab, die es nicht gibt.
func (r Rolle) Pruefe() error {
	if rangfolge[r] > 0 {
		return nil
	}

	return NewFieldError("rolle", SchluesselValidierung,
		"Die Rolle %q gibt es nicht.", string(r))
}

// Sichtbarkeit sagt, wer ein Profil überhaupt sieht.
type Sichtbarkeit string

const (
	// SichtbarPrivat: nur der Eigner und wer eine Freigabe hat.
	SichtbarPrivat Sichtbarkeit = "privat"

	// SichtbarImBestand: jeder Nutzer dieses Bestands sieht es — wie ein
	// [RolleLeser].
	//
	// **Und nicht mehr.** „Sichtbar" heißt hier dasselbe wie dort: der Name
	// steht in der Übersicht, der Inhalt nicht.
	SichtbarImBestand Sichtbarkeit = "bestand"
)

// Sichtbarkeiten gibt beide.
func Sichtbarkeiten() []Sichtbarkeit {
	return []Sichtbarkeit{SichtbarPrivat, SichtbarImBestand}
}

// Pruefe weist eine Sichtbarkeit ab, die es nicht gibt.
//
// **Leer gilt als `privat`.** Ein Bestand aus der Zeit vor dieser Wanderung
// trägt sie zwar überall ein (`DEFAULT 'privat'`), aber ein Aufruf, der sie
// wegläßt, meint nicht „egal" — er meint die vorsichtige Antwort.
func (s Sichtbarkeit) Pruefe() error {
	if s == "" {
		return nil
	}

	for _, bekannt := range Sichtbarkeiten() {
		if s == bekannt {
			return nil
		}
	}

	return NewFieldError("sichtbarkeit", SchluesselValidierung,
		"Die Sichtbarkeit %q gibt es nicht.", string(s))
}

// Profilfreigabe ist die Rolle **eines** Nutzers an **einem** Profil.
type Profilfreigabe struct {
	ProfilID int64
	NutzerID int64
	Rolle    Rolle

	// Angelegt ist RFC 3339 — wann die Freigabe entstand.
	Angelegt string
}

// Validate prüft, was ohne den Bestand zu prüfen ist.
func (f Profilfreigabe) Validate() error {
	if f.ProfilID <= 0 {
		return NewFieldError("profilID", SchluesselValidierung,
			"Eine Freigabe braucht ein Profil.")
	}

	if f.NutzerID <= 0 {
		return NewFieldError("nutzerID", SchluesselValidierung,
			"Eine Freigabe braucht einen Nutzer.")
	}

	return f.Rolle.Pruefe()
}

// Freigabefehler ist die Absage an einen Wechsel, der nicht erlaubt ist.
//
// **Er nennt, was verlangt war** — dieselbe Überlegung wie beim Scope eines
// Zugriffstokens: wer bis hierher kommt, ist kein Fremder, und ohne diesen
// Satz suchte er den Fehler bei sich.
func Freigabefehler(verlangt Rolle) error {
	return fmt.Errorf("%w: dafür braucht es mindestens die Rolle %q",
		ErrErlaubnis, string(verlangt))
}
