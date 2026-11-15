package domain_test

// **Was ein Gast bei seiner Rückkehr bekommt** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// **Gegen die Verwechslung, die den ersten Entwurf falsch gemacht hätte.**
// Die Kette eines Ereignisses ist für **alle** gleich — sie schließt die
// privaten Ergänzungen ausdrücklich aus. Ein Gast, der mit ihr allein
// zurückkäme, könnte den privaten Strom eines anderen bekommen, ohne dass es
// auffiele. Das Issue verlangt deshalb einen **sichtbezogenen** Hash.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einEreignisMitPrivatem baut eines mit zwei Ergänzungen an zwei Sitze.
func einEreignisMitPrivatem() domain.Ereignis {
	return domain.Ereignis{
		PartieID: "p1", Folge: 7, Art: "karte.ziehen", Kette: "abcdef",
		Privat: []domain.Ergaenzung{
			{Empfaenger: "anna", Inhalt: map[string]any{"karte": "k1"}},
			{Empfaenger: "bert", Inhalt: map[string]any{"karte": "k2"}},
		},
	}
}

// **Zwei Sitze bei derselben Folge haben verschiedene Abdrücke.**
func TestDerSichtabdruckTrenntZweiSitze(t *testing.T) {
	t.Parallel()

	ereignis := einEreignisMitPrivatem()

	annas, fehler := ereignis.Sichtabdruck("anna")
	if fehler != nil {
		t.Fatalf("annas Abdruck: %v", fehler)
	}

	berts, fehler := ereignis.Sichtabdruck("bert")
	if fehler != nil {
		t.Fatalf("berts Abdruck: %v", fehler)
	}

	if annas == berts {
		t.Error("beide Sitze haben denselben Abdruck — die Ergänzung geht " +
			"nicht ein, und ein vertauschter privater Strom fiele nicht auf")
	}

	// **Und die Kette allein reicht nicht**: ein Zuschauer ohne Ergänzung
	// bekommt einen dritten.
	ohne, fehler := ereignis.Sichtabdruck("carla")
	if fehler != nil {
		t.Fatalf("carlas Abdruck: %v", fehler)
	}

	if ohne == annas || ohne == berts {
		t.Error("wer nichts Privates bekam, hat denselben Abdruck wie jemand, " +
			"der etwas bekam")
	}
}

// **Derselbe Sitz bekommt zweimal denselben Abdruck.**
//
// **Sonst wäre er als Nachweis wertlos** — der Gast rechnet ihn bei sich, der
// Wirt bei sich, und beide müssen auf dasselbe kommen.
func TestDerSichtabdruckIstWiederholbar(t *testing.T) {
	t.Parallel()

	ereignis := einEreignisMitPrivatem()

	einmal, _ := ereignis.Sichtabdruck("anna") //nolint:errcheck // gleich geprüft

	nochmal, fehler := ereignis.Sichtabdruck("anna")
	if fehler != nil {
		t.Fatalf("zweiter Abdruck: %v", fehler)
	}

	if einmal != nochmal {
		t.Errorf("zweimal derselbe Sitz, zwei Abdrücke: %q und %q",
			einmal, nochmal)
	}
}

// **Die gekürzte Fassung ergibt denselben Abdruck wie die volle.**
//
// # Warum das die eigentliche Zusage ist
//
// **Weil der Gast nur die gekürzte hat.** [domain.Ereignis.Fuer] nimmt die
// fremden Ergänzungen heraus; käme dabei ein anderer Abdruck heraus, hielte
// der Wirt jeden zurückkehrenden Gast für einen Fremden.
func TestDerSichtabdruckUeberlebtDieKuerzung(t *testing.T) {
	t.Parallel()

	ereignis := einEreignisMitPrivatem()

	voll, fehler := ereignis.Sichtabdruck("anna")
	if fehler != nil {
		t.Fatalf("voller Abdruck: %v", fehler)
	}

	gekuerzt, fehler := ereignis.Fuer("anna").Sichtabdruck("anna")
	if fehler != nil {
		t.Fatalf("gekürzter Abdruck: %v", fehler)
	}

	if voll != gekuerzt {
		t.Errorf("voll %q, gekürzt %q — der Gast käme nie wieder herein",
			voll, gekuerzt)
	}
}

// **Eine Ergänzung, die sich nicht abbilden lässt, wird benannt.**
//
// **Sie ist erreichbar**: `Ergaenzung.Inhalt` ist eine offene Abbildung, und
// was ein Regelausleger hineinlegt, entscheidet er (§8).
func TestEineUnabbildbareErgaenzungWirdBenannt(t *testing.T) {
	t.Parallel()

	ereignis := domain.Ereignis{
		PartieID: "p1", Folge: 1, Kette: "abc",
		Privat: []domain.Ergaenzung{{
			Empfaenger: "anna",
			Inhalt:     map[string]any{"kanal": make(chan int)},
		}},
	}

	if _, fehler := ereignis.Sichtabdruck("anna"); fehler == nil {
		t.Error("eine unabbildbare Ergänzung ging durch")
	}
}
