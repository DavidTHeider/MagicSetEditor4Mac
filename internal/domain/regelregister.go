// **Das Regelregister** — PLAY-01
// (#1709):
// „Ein versioniertes Register verbindet CR-Fassung, Regel-ID, eigene
// Kurzbeschreibung, Status, Voraussetzungen, produktiven Aufrufer und
// positive/negative Szenariotests."
//
// # Warum neben der Regelmatrix
//
// **Die Matrix sagt je Bereich einen Grad** — dreizehn Zeilen, die Zusage an
// die Spielenden. **Das Register sagt je Regel, woran sie hängt**: welche
// Funktion sie durchsetzt, welcher Lauf den zulässigen Fall zeigt und welcher
// den unzulässigen abweist. Die Matrix ohne Register wäre ein Grad ohne
// Deckung; `docs/REGELABDECKUNG.md` führt dieselbe Frage je **Kapitel**.
//
// # Was ein Stand heißt
//
//   - **durchgesetzt**: ein unzulässiger Zug wird abgewiesen, oder die Regel
//     wirkt von selbst — und beide Richtungen haben einen Lauf.
//   - **teilweise**: sie gilt in einem benannten Ausschnitt; die Grenze steht
//     dabei.
//   - **manuell**: der Tisch lässt es von Hand zu und weist **nichts** ab —
//     eine fehlende Regel darf keinen zulässigen Zug verbieten (AK-04).
//   - **offen**: nicht gebaut; die Grenze sagt, was fehlt.
//
// # Und eine neue Fassung der CR
//
// **Macht jeden Eintrag ungültig, bis er nachgesehen ist** (AK-04). Die
// Fassung steht je Eintrag und muss der des Grundprofils gleichen; wer
// [Grundprofilquelle] hebt, sieht jede Regel einzeln nach — und nicht einen
// Suchen-und-Ersetzen-Lauf über alle.
//
// **Die Beschreibungen sind eigene Worte** (AGENTS.md §11): sie kommen aus
// den Kommentaren an den durchsetzenden Stellen, die beim Bau gegen die CR
// nachgelesen wurden, und nicht aus dem Wortlaut.

package domain

import (
	"fmt"
	"regexp"
	"strings"
)

// Regelstand sagt, wie weit eine Regel reicht.
type Regelstand string

// Die vier Stände.
const (
	RegelDurchgesetzt Regelstand = "durchgesetzt"
	RegelTeilweise    Regelstand = "teilweise"
	RegelManuell      Regelstand = "manuell"
	RegelOffen        Regelstand = "offen"
)

// Regelkatalog ist eine Gruppe aus PLAY-01.AK-03 — die Reihenfolge, in der
// das Register gefüllt wird.
type Regelkatalog string

// Die sieben Gruppen, in der Ordnung des Issues.
const (
	KatalogSpielaufbau      Regelkatalog = "spielaufbau"
	KatalogZonenwechsel     Regelkatalog = "zonenwechsel"
	KatalogPrioritaetStapel Regelkatalog = "prioritaet-stapel"
	KatalogZieleKosten      Regelkatalog = "ziele-kosten"
	KatalogKampf            Regelkatalog = "kampf"
	KatalogZustandsaktionen Regelkatalog = "zustandsaktionen"
	KatalogEffekte          Regelkatalog = "effekte"
)

// Regelkataloge nennt die sieben Gruppen.
func Regelkataloge() []Regelkatalog {
	return []Regelkatalog{
		KatalogSpielaufbau, KatalogZonenwechsel, KatalogPrioritaetStapel,
		KatalogZieleKosten, KatalogKampf, KatalogZustandsaktionen,
		KatalogEffekte,
	}
}

// Regeleintrag ist eine Regel im Register.
type Regeleintrag struct {
	// Nummer ist die Regel-ID der CR — `508.1a`.
	Nummer string

	// Fassung ist die Fassung der CR, gegen die der Eintrag nachgesehen ist.
	Fassung string

	Katalog Regelkatalog
	Bereich Regelbereich

	// Beschreibung sagt die Regel in eigenen Worten — nie im Wortlaut.
	Beschreibung string

	Stand Regelstand

	// Voraussetzungen sind Regeln, ohne die diese nicht wirken kann.
	Voraussetzungen []string

	// Stelle ist die Funktion, an der die Regel steht und die sie nennt —
	// `domain.Partie.angreiferzuteilung`.
	Stelle string

	// Aufrufer ist der **produktive** Einstieg, der zur Stelle führt —
	// `domain.Partie.Weiter`. Dass er von `service` oder `app` erreicht
	// wird und die Stelle erreicht, prüft
	// `internal/werkzeuge/regelregister_test.go`.
	Aufrufer string

	// Positiv und Negativ sind Testnamen: der zulässige Fall, und der Fall,
	// in dem die Regel abweist oder nicht greift.
	Positiv string
	Negativ string

	// Grenze sagt, was die Regel hier **nicht** tut.
	Grenze string
}

// regelnummer ist die Gestalt einer Regel-ID der CR.
var regelnummer = regexp.MustCompile(`^\d{3}(\.\d+[a-z]?)?$`)

// Pruefe weist einen Eintrag ab, der keine Deckung hat.
func (e Regeleintrag) Pruefe() error {
	if fehler := e.pruefeKopf(); fehler != nil {
		return fehler
	}

	return e.pruefeDeckung()
}

// pruefeKopf hält Nummer, Fassung, Katalog, Bereich und Beschreibung.
func (e Regeleintrag) pruefeKopf() error {
	switch {
	case !regelnummer.MatchString(e.Nummer):
		return fmt.Errorf("%w: %q ist keine Nummer der CR", ErrValidation, e.Nummer)

	case e.Fassung != Grundprofilquelle:
		return fmt.Errorf("%w: CR %s ist gegen die Fassung %q nachgesehen, "+
			"gilt aber %q — jeder Eintrag wird einzeln nachgesehen (PLAY-01.AK-04)",
			ErrValidation, e.Nummer, e.Fassung, Grundprofilquelle)

	case !istRegelkatalog(e.Katalog):
		return fmt.Errorf("%w: CR %s: %q ist kein Katalog aus PLAY-01.AK-03",
			ErrValidation, e.Nummer, e.Katalog)

	case !IstRegelbereich(e.Bereich):
		return fmt.Errorf("%w: CR %s: %q ist kein Bereich der Regelmatrix",
			ErrValidation, e.Nummer, e.Bereich)

	case strings.TrimSpace(e.Beschreibung) == "":
		return fmt.Errorf("%w: CR %s hat keine Beschreibung", ErrValidation, e.Nummer)
	}

	for _, voraus := range e.Voraussetzungen {
		if !regelnummer.MatchString(voraus) {
			return fmt.Errorf("%w: CR %s: die Voraussetzung %q ist keine Nummer",
				ErrValidation, e.Nummer, voraus)
		}
	}

	return nil
}

// pruefeDeckung hält, was der Stand verlangt.
//
// **Durchgesetzt heißt beide Richtungen**: ein Lauf, der zeigt, dass es geht,
// und einer, der zeigt, dass das Unzulässige nicht geht. Ein Lauf allein
// belegt ein Datenfeld oder eine Funktion — nicht eine Regel (AK-02).
func (e Regeleintrag) pruefeDeckung() error {
	switch e.Stand {
	case RegelDurchgesetzt:
		if e.Stelle == "" || e.Aufrufer == "" || e.Positiv == "" {
			return fmt.Errorf("%w: CR %s ist durchgesetzt und nennt keine "+
				"Stelle, keinen Aufrufer oder keinen Lauf", ErrValidation, e.Nummer)
		}

		if e.Negativ == "" {
			return fmt.Errorf("%w: CR %s ist durchgesetzt und nennt keinen "+
				"Gegenfall", ErrValidation, e.Nummer)
		}

	case RegelTeilweise:
		if e.Stelle == "" || e.Aufrufer == "" || e.Positiv == "" || e.Grenze == "" {
			return fmt.Errorf("%w: CR %s ist teilweise und braucht Stelle, "+
				"Aufrufer, Lauf und Grenze", ErrValidation, e.Nummer)
		}

	// **Manuell hat keine Stelle** — es gibt keine, an der die Regel
	// durchgesetzt würde. Der Aufrufer ist der Weg, der sie **nicht**
	// abweist, und der Lauf belegt, dass er durchgeht.
	case RegelManuell:
		if e.Aufrufer == "" || e.Positiv == "" || e.Grenze == "" {
			return fmt.Errorf("%w: CR %s ist manuell und braucht Aufrufer, Lauf "+
				"und Grenze", ErrValidation, e.Nummer)
		}

	case RegelOffen:
		if e.Grenze == "" {
			return fmt.Errorf("%w: CR %s ist offen und nennt keine Grenze",
				ErrValidation, e.Nummer)
		}

	default:
		return fmt.Errorf("%w: CR %s: %q ist kein Stand", ErrValidation,
			e.Nummer, e.Stand)
	}

	return nil
}

// istRegelkatalog sagt, ob es die Gruppe gibt.
func istRegelkatalog(katalog Regelkatalog) bool {
	for _, bekannt := range Regelkataloge() {
		if katalog == bekannt {
			return true
		}
	}

	return false
}

// Regelregister gibt alle Einträge — in der Ordnung der Kataloge.
func Regelregister() []Regeleintrag {
	return append(regelnDesAblaufs(), regelnDesKampfs()...)
}
