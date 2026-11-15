// **Der Weg vom Naht-Fehler zurück zum Go-Fehler** — abgetrennt am
// 2026-09-02, weil `fehler.go` damit über die Zeilengrenze aus §2 ging.
//
// # Der Schnitt folgt der Frage
//
// Dort steht, wie aus einem Go-Fehler die Gestalt wird, die das **Fenster**
// braucht; hier, wie eine Fläche **ohne** Fenster wieder einen Go-Fehler
// daraus macht. Zwei Richtungen, zwei Fragen.

package app

import (
	"context"
	"errors"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// AlsFehler macht aus der Naht-Gestalt wieder einen Go-Fehler.
//
// **Für die Flächen ohne Fenster** — Werkstück
// #1391.
// Die Rechnungen des Symboleditors geben eine `SymbolAntwort` mit einem
// `*Fehler` darin, weil das die Gestalt ist, die das Fenster braucht: Code,
// Schlüssel, Satz. Die Kommandozeile setzt daraus einen Rückgabewert, MCP
// sein `isError`, der Dienst seinen Status — und alle drei fragen mit
// `errors.Is`.
//
// **Der Satz wird nicht neu gebaut, sondern durchgereicht.** [mitSatz] legt
// `fehler.Error()` in die Meldung, und die trägt das Klassenpräfix schon —
// „Validierung fehlgeschlagen: …". Ein `fmt.Errorf("%w: %s", …)` darum gäbe
// es zweimal; gesehen im ersten Anlauf, in der Ausgabe eines Wächters.
//
// **Der Weg zurück ist nicht der Weg hin.** [Uebersetze] verliert die
// ursprüngliche Fehlerkette; was hier entsteht, trägt denselben Satz und
// dieselbe **Klasse**, nicht dieselbe Kette. Mehr braucht keiner der drei.
func (f *Fehler) AlsFehler() error {
	if f == nil {
		return nil
	}

	if klasse, bekannt := arten[f.Code]; bekannt {
		return nahtfehler{satz: f.Meldung, klasse: klasse, genauer: f.genauer()}
	}

	// **Der Rückfall ist ein nackter Fehler und keine Klasse.**
	// [CodeUnbekannt] heißt genau das: wir wissen nicht, was es war. Ihn auf
	// `ErrValidation` abzubilden hieße zu behaupten, der Aufrufer sei schuld.
	return errors.New(f.Meldung)
}

// arten ordnet jedem Code seine Fehlerart zu.
//
// # Warum eine Tabelle und kein Schalter
//
// **Weil der Schalter drei von acht Fällen kannte und niemand es sah.**
// `ErrErlaubnis`, `ErrBelegt`, `ErrSchluesselFehlt` und `ErrNurLesend` fielen
// auf einen nackten Fehler zurück — und damit kam am
// Dienst `500 unbekannt` an, wo `internal/rest/antwort.go` seit Werkstück 11
// eine 403 vorsieht. Gemessen am 2026-09-06 an `PUT /profile/2/freigaben/1`.
//
// **Das ist §3 in Reinform**: zwei grüne Seiten einer Naht sind kein Beleg für
// die Naht. `Uebersetze` hatte seinen Lauf, `antwort.go` hatte seinen, und
// dazwischen fiel die Art heraus.
//
// **`nahtfehlervollstaendig_test.go` hält die Tabelle jetzt vollständig** —
// ein Code ohne Eintrag macht ihn rot.
//
// **[CodeUnbekannt] steht mit Absicht nicht darin.** Er heißt genau das: wir
// wissen nicht, was es war. Ihn auf `ErrValidation` abzubilden hieße zu
// behaupten, der Aufrufer sei schuld.
//
// # Was diese Tabelle **nicht** kann
//
// **Sie sieht nur den Code, nicht den Schlüssel.** `domain.ErrZweiterFaktor`
// kommt als **Feldbefund** herein (`domain.FehlenderFaktor`), und
// [Uebersetze] gibt ihm deshalb [CodeValidierung] mit dem Schlüssel
// `zweiterFaktor`. Am Fenster genügt das — der Katalog übersetzt über den
// Schlüssel. **Ohne Fenster geht die feinere Art verloren**, und ein Aufrufer
// bekommt `ErrValidation`.
//
// **Das ist heute folgenlos und morgen vielleicht nicht**: keine der drei
// Flächen unterscheidet den fehlenden Faktor von einer Abweisung — REST gibt
// beidem 400. Wer das ändert, ändert zuerst diesen Absatz und reicht den
// Schlüssel mit durch.
var arten = map[Fehlercode]error{
	CodeNichtGefunden:   domain.ErrNotFound,
	CodeKonflikt:        domain.ErrConflict,
	CodeValidierung:     domain.ErrValidation,
	CodeBelegt:          domain.ErrBelegt,
	CodeErlaubnis:       domain.ErrErlaubnis,
	CodeSchluesselFehlt: domain.ErrSchluesselFehlt,
	CodeNurLesend:       domain.ErrNurLesend,

	// **Die zwei, die hier fehlten**, bis der Zweitblick über Phase 6.1 sie
	// fand (Befund 5): ein voller Datenträger kam am Dienst als 500 statt
	// 507 an, ein Abbruch als 500 statt 499. `CodeVoll` kam mit #1812 dazu,
	// und diese Tabelle hat niemand angefasst — der Lauf, der sie für
	// vollständig erklärte, führte seine eigene Liste.
	CodeVoll:        domain.ErrVoll,
	CodeAbgebrochen: context.Canceled,

	// **Und der abgeschaltete Bereich** — OPS-10 (#1908).
	CodeAbgeschaltet: domain.ErrAbgeschaltet,

	// **Und der fremde Dienst** (D02, #2233) — ein 429 kommt dabei als
	// Gegenstelle zurück; den Unterschied trägt der Schlüssel.
	CodeNetz: domain.ErrGegenstelle,

	// **Und das eigene Kontingent** (Zweitblick über Phase 1 von 0.9.0,
	// Befund 5) — der Aufrufer liegt über seinem Budget an dieser Anwendung.
	CodeDrossel: domain.ErrKontingent,
}

// nahtfehler trägt einen fertigen Satz und eine Klasse.
//
// **Ein eigener Typ statt `fmt.Errorf`**, weil beides zugleich gebraucht
// wird: `Error()` gibt den Satz, wie er im Katalog steht, und `Unwrap` die
// Klasse, nach der `errors.Is` fragt. Mit `fmt.Errorf("%w: %s", …)` stünde
// das Präfix der Klasse ein zweites Mal im Satz.
//
// **Und, wo die Hinübersetzung mehr braucht, den genauen Fehler** daneben
// (`nahtgenauer.go`).
type nahtfehler struct {
	satz    string
	klasse  error
	genauer error
}

func (n nahtfehler) Error() string { return n.satz }

func (n nahtfehler) Unwrap() []error {
	if n.genauer == nil {
		return []error{n.klasse}
	}

	return []error{n.klasse, n.genauer}
}
