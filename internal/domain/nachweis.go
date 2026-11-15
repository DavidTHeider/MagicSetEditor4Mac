// **Ein Nachweis bindet, was er erlaubt** — Werkstück O1 von Phase 5.2
// (#1425),
// nach [Vermerk 0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md).
//
// # Wozu
//
// Eine Rolle sagt, was jemand grundsätzlich darf. Für das, was draußen etwas
// ändert oder verbindlich ist — ein Listing, ein Tausch, ein Auftrag mit
// Kosten —, reicht das nicht: dort soll ein Mensch **dieses eine** gewollt
// haben, und es soll sich nicht ein zweites Mal verwenden lassen. Das ist ein
// Nachweis: ausgestellt für einen Nutzer, ein Profil, einen Zweck, ein Ziel
// und einen Stand, mit einer Frist und wahlweise einer Kostengrenze — und
// **einmal** einlösbar.
//
// # Warum ein `bestaetigt`-Feld keiner ist
//
// Ein Feld, das der Aufrufer selbst setzt, beweist nur, dass er das Feld
// kennt. Ein Nachweis entsteht dort, wo der Mensch ist — am Fenster —, wird
// im Bestand geführt und beim Einlösen gegen die Bindung gehalten, die der
// Aufruf mitbringt. Ein wiederverwendeter, abgelaufener oder auf ein anderes
// Ziel gebundener Nachweis erlaubt nichts.

package domain

import (
	"fmt"
	"strings"
	"time"
)

// Kostengrenze ist die Obergrenze, bis zu der ein Nachweis Kosten erlaubt.
//
// **In der kleinsten Einheit der Währung** (Cent), damit es keine Rundung
// gibt; ein Betrag von null heißt „keine Grenze".
type Kostengrenze struct {
	Betrag   int64
	Waehrung string
}

// Nachweis ist die vorab erteilte, eng begrenzte Erlaubnis.
type Nachweis struct {
	// Kennung ist zufällig und einmalig — sie ist, was der Aufruf vorzeigt.
	Kennung string

	NutzerID int64
	ProfilID int64

	// Zweck sagt, wofür — `handel.angebot`, `ki.auftrag` —, Ziel, wofür
	// genau: die Kennung des Angebots, der Name des Anbieters.
	Zweck string
	Ziel  string

	// Revision ist der Stand des Ziels, für den der Nachweis gilt. Ändert
	// sich das Ziel danach, gilt er nicht mehr — niemand hat den neuen Stand
	// gesehen.
	Revision int64

	Grenze Kostengrenze

	Ausgestellt time.Time
	Ablauf      time.Time

	// Eingeloest sagt, dass der Nachweis verbraucht ist.
	Eingeloest bool
}

// Nachweisbindung ist, was ein Aufruf vorzeigt, wenn er einen Nachweis
// einlösen will.
type Nachweisbindung struct {
	NutzerID int64
	ProfilID int64
	Zweck    string
	Ziel     string
	Revision int64

	// Kosten und Waehrung sind, was der Aufruf kosten wird — null, wenn
	// nichts.
	Kosten   int64
	Waehrung string
}

// Pruefe weist einen Nachweis ab, der so nicht ausgestellt werden darf.
func (n Nachweis) Pruefe() error {
	switch {
	case n.NutzerID <= 0:
		return NewFieldError("nutzerID", SchluesselValidierung,
			"Ein Nachweis braucht einen Nutzer.")
	case n.ProfilID <= 0:
		return NewFieldError("profilID", SchluesselValidierung,
			"Ein Nachweis braucht ein Profil.")
	case strings.TrimSpace(n.Zweck) == "":
		return NewFieldError("zweck", SchluesselValidierung,
			"Ein Nachweis braucht einen Zweck — sonst gälte er für alles.")
	case strings.TrimSpace(n.Ziel) == "":
		return NewFieldError("ziel", SchluesselValidierung,
			"Ein Nachweis braucht ein Ziel — sonst gälte er für alles.")
	case !n.Ablauf.After(n.Ausgestellt):
		return NewFieldError("ablauf", SchluesselValidierung,
			"Ein Nachweis läuft nach dem Ausstellen ab, nicht davor.")
	case n.Grenze.Betrag < 0:
		return NewFieldError("grenze", SchluesselValidierung,
			"Eine Kostengrenze ist nicht negativ.")
	case n.Grenze.Betrag > 0 && strings.TrimSpace(n.Grenze.Waehrung) == "":
		return NewFieldError("grenze", SchluesselValidierung,
			"Eine Kostengrenze nennt ihre Währung.")
	}

	return nil
}

// Gilt sagt, ob der Nachweis für diese Bindung zu dieser Zeit gilt.
//
// **Die Absage nennt, was nicht passte** — ein Nachweis, der schweigend
// versagt, ließe jemanden am falschen Ende suchen. Und sie ist [ErrErlaubnis]:
// die Eingabe ist richtig, was fehlt, ist die Erlaubnis.
func (n Nachweis) Gilt(bindung Nachweisbindung, jetzt time.Time) error {
	switch {
	case n.Eingeloest:
		return fmt.Errorf("%w: dieser Nachweis ist schon eingelöst", ErrErlaubnis)
	case jetzt.Before(n.Ausgestellt):
		return fmt.Errorf("%w: dieser Nachweis gilt noch nicht", ErrErlaubnis)
	case !jetzt.Before(n.Ablauf):
		return fmt.Errorf("%w: dieser Nachweis ist abgelaufen", ErrErlaubnis)
	case n.NutzerID != bindung.NutzerID:
		return fmt.Errorf("%w: dieser Nachweis gehört einem anderen Nutzer", ErrErlaubnis)
	case n.ProfilID != bindung.ProfilID:
		return fmt.Errorf("%w: dieser Nachweis gehört einem anderen Profil", ErrErlaubnis)
	case n.Zweck != bindung.Zweck:
		return fmt.Errorf("%w: dieser Nachweis gilt für %q, nicht für %q",
			ErrErlaubnis, n.Zweck, bindung.Zweck)
	case n.Ziel != bindung.Ziel:
		return fmt.Errorf("%w: dieser Nachweis gilt für ein anderes Ziel", ErrErlaubnis)
	case n.Revision != bindung.Revision:
		return fmt.Errorf("%w: das Ziel hat sich seit dem Nachweis geändert — "+
			"Stand %d statt %d", ErrErlaubnis, bindung.Revision, n.Revision)
	}

	return n.Grenze.deckt(bindung.Kosten, bindung.Waehrung)
}

// deckt sagt, ob die Grenze diese Kosten trägt.
//
// **Ohne Grenze trägt sie alles, ohne Kosten fragt niemand** — beides ist
// die häufige Lage: die meisten Wirkungen kosten nichts, und eine Grenze
// ist eine Wahl beim Ausstellen.
func (g Kostengrenze) deckt(kosten int64, waehrung string) error {
	if g.Betrag == 0 || kosten == 0 {
		return nil
	}

	if waehrung != g.Waehrung {
		return fmt.Errorf("%w: die Grenze gilt in %s, die Kosten fallen in %s an",
			ErrErlaubnis, g.Waehrung, waehrung)
	}

	if kosten > g.Betrag {
		return fmt.Errorf("%w: %d übersteigt die Grenze von %d %s",
			ErrErlaubnis, kosten, g.Betrag, g.Waehrung)
	}

	return nil
}
