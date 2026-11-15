package service_test

// **Die Regelwahl an der Freigabe** — PLAY-02.AK-01 und AK-02
// (#1710).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ohne Angabe gilt der Anfängermodus** — die vorsichtige Vorgabe.
func TestOhneAngabeGiltDerAnfaengermodus(t *testing.T) {
	t.Parallel()

	wunsch := einFreigabewunsch()
	wunsch.Wahl = domain.Regelwahl{
		Abgewaehlt: []domain.Regelbereich{domain.BereichCombat},
	}

	freigabe, fehler := service.NeueRegelfreigabe().Pruefe(wunsch)
	if fehler != nil {
		t.Fatalf("prüfen: %v", fehler)
	}

	if freigabe.Erlaubt {
		t.Error("die Abwahl ging ohne Modusangabe durch")
	}

	if !strings.Contains(strings.Join(freigabe.Hindernisse, " "),
		"freien Modus") {
		t.Errorf("das Hindernis nennt den Ausweg nicht: %v",
			freigabe.Hindernisse)
	}
}

// **Im freien Modus geht sie durch — und die Matrix ändert sich mit.**
func TestImFreienModusAendertDieAbwahlDieMatrix(t *testing.T) {
	t.Parallel()

	wunsch := einFreigabewunsch()
	wunsch.Modus = domain.ModusFrei
	wunsch.Wahl = domain.Regelwahl{
		Abgewaehlt: []domain.Regelbereich{domain.BereichZonen},
	}

	freigabe, fehler := service.NeueRegelfreigabe().Pruefe(wunsch)
	if fehler != nil {
		t.Fatalf("prüfen: %v", fehler)
	}

	if !freigabe.Erlaubt {
		t.Fatalf("die Abwahl wurde abgewiesen: %v", freigabe.Hindernisse)
	}

	if freigabe.Matrix.Grad(domain.BereichZonen) !=
		domain.UnterstuetzungManuell {
		t.Errorf("die Zonen stehen auf %q",
			freigabe.Matrix.Grad(domain.BereichZonen))
	}

	// **Und die Lücke steht in der Liste** — wer abwählt, sieht, dass er es
	// selbst macht.
	if !enthaeltBereich(freigabe.Luecken, domain.BereichZonen) {
		t.Errorf("die Zonen fehlen unter den Lücken: %v", freigabe.Luecken)
	}
}

// **Der Baum kommt mit** — an jeder Freigabe, auch an einer ohne Abwahl.
func TestDieFreigabeTraegtDenBaum(t *testing.T) {
	t.Parallel()

	freigabe, fehler := service.NeueRegelfreigabe().Pruefe(einFreigabewunsch())
	if fehler != nil {
		t.Fatalf("prüfen: %v", fehler)
	}

	if len(freigabe.Baum) != 2 {
		t.Fatalf("der Baum hat %d Gruppen", len(freigabe.Baum))
	}

	if len(freigabe.Widersprueche) != 0 {
		t.Errorf("ohne Abwahl kam ein Widerspruch: %+v", freigabe.Widersprueche)
	}
}

// **Eine Abwahl, an der etwas hängt, wird erklärt und nicht abgewiesen.**
func TestEinWiderspruchWirdErklaertUndNichtAbgewiesen(t *testing.T) {
	t.Parallel()

	wunsch := einFreigabewunsch()
	wunsch.Modus = domain.ModusFrei
	wunsch.Wahl = domain.Regelwahl{
		Abgewaehlt: []domain.Regelbereich{domain.BereichPrioritaet},
	}

	freigabe, fehler := service.NeueRegelfreigabe().Pruefe(wunsch)
	if fehler != nil {
		t.Fatalf("prüfen: %v", fehler)
	}

	if len(freigabe.Widersprueche) == 0 {
		t.Error("der Widerspruch fehlt")
	}

	// **Erlaubt bleibt es trotzdem** — AK-02 verlangt eine Erklärung, keine
	// Sperre.
	if !freigabe.Erlaubt {
		t.Errorf("die Erklärung wurde zur Sperre: %v", freigabe.Hindernisse)
	}
}

// **Ein Name, den es nicht gibt, ist ein Fehler und keine Antwort.**
func TestEineUnbekannteAbwahlIstEinFehler(t *testing.T) {
	t.Parallel()

	wunsch := einFreigabewunsch()
	wunsch.Wahl = domain.Regelwahl{Abgewaehlt: []domain.Regelbereich{"zonnen"}}

	if _, fehler := service.NeueRegelfreigabe().Pruefe(
		wunsch); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}
}

// enthaeltBereich sagt, ob ein Bereich in einer Liste steht.
func enthaeltBereich(liste []domain.Regelbereich, gesucht domain.Regelbereich) bool {
	for _, bereich := range liste {
		if bereich == gesucht {
			return true
		}
	}

	return false
}
