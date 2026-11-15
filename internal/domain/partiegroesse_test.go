package domain_test

// **Eine Partie, die keine Datei mehr wäre, fängt nicht an** — #2065.
//
// Aus zwei Decks mit je 147.720 Karten entstand eine Partie, die sich weder
// auflisten noch ausführen und wieder einlesen ließ: die Einfuhr liest
// höchstens [domain.PartieaustauschHoechstens], das Anlegen fragte nicht.
// Die Seite Spielen zeigte danach einen Panic des Treibers, und ⌘Q hing.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEinePartieDieKeineDateiMehrWaereFaengtNichtAn(t *testing.T) {
	t.Parallel()

	gross := mitBibliotheken(2, 20000)
	if laenge := len(domain.FormePartie(gross)); laenge <= domain.PartieaustauschHoechstens {
		t.Fatalf("die Probe ist mit %d Bytes zu klein für ihre Frage", laenge)
	}

	fehler := gross.Pruefe()
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("eine Partie über der Grenze einer Partiedatei fing an: %v", fehler)
	}

	// **Die Meldung nennt beide Zahlen** — wer abgewiesen wird, soll wissen,
	// wie weit er darüber liegt.
	if !strings.Contains(fehler.Error(), "40000") {
		t.Errorf("die Meldung nennt die Zahl der Objekte nicht: %v", fehler)
	}

	// **Dass jemand zuhört** (§3): dieselbe Partie mit hundert Karten je
	// Bibliothek fängt an.
	if fehler := mitBibliotheken(2, 100).Pruefe(); fehler != nil {
		t.Errorf("eine gewöhnliche Partie wurde abgewiesen: %v", fehler)
	}
}
