package service_test

// **Der Abgleich und was ihn schließt** — Werkstück `T6` von Phase 5.6
// (#1471),
// nachgeschärft durch
// #1650.
//
// **Herausgezogen aus `handelsverbuchung_test.go`**, weil die Datei an der
// Zeilengrenze stand (§2). Der Schnitt folgt der Frage: dort steht, was ein
// Abschluss **bewirkt**, hier, wann er **fertig** ist.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Abgleich bleibt offen, bis die Gegenseite ihn bestätigt.**
//
// Das Issue verlangt es wörtlich: „Fehlende Gegenstelle/zweite Signatur meldet
// Abgleich offen; Wiederanlauf konvergiert ohne zweite Abbuchung."
//
// **Und der Beleg bleibt einseitig, auch wenn beide Stimmen da sind**
// (#1650):
// eine Stimme ist eine Stimme und keine Unterschrift. Beidseitig wird er,
// wenn ein **geprüfter** Beleg der Gegenseite ankommt.
func TestDerAbgleichBleibtOffenBisDieGegenseiteBestaetigt(t *testing.T) {
	t.Parallel()

	stelle, stand, bestand := eineVerbuchung(t)
	stelleEin(t, stand, bestand, true)

	if _, fehler := stelle.Verbuchen(t.Context(), derVorgang, derFreibrief); fehler != nil {
		t.Fatalf("verbuchen: %v", fehler)
	}

	offen, fehler := stelle.Abgleichen(t.Context(), derVorgang)
	if fehler != nil {
		t.Fatalf("abgleichen: %v", fehler)
	}

	if offen.Zustand != domain.HandelAbgleichOffen {
		t.Errorf("der Vorgang steht auf %q", offen.Zustand)
	}

	// **Und der Beleg ist immer noch einseitig.**
	beleg, fehler := stelle.Beleg(t.Context(), derVorgang)
	if fehler != nil {
		t.Fatalf("Beleg: %v", fehler)
	}

	if beleg == nil || beleg.Beidseitig {
		t.Errorf("der Beleg ist %+v", beleg)
	}

	// **Eine Stimme genügt nicht.** Solange eine Partei fehlt, bleibt der
	// Vorgang im Abgleich, und der Beleg bleibt einseitig — zwei
	// Unterschriften vorzutäuschen ist genau das, was das Issue verbietet.
	halb, fehler := stelle.Bestaetigen(t.Context(), derVorgang, "bgegenueber", derFreibrief)
	if fehler != nil {
		t.Fatalf("erste Bestätigung: %v", fehler)
	}

	if halb.Zustand != domain.HandelAbgleichOffen {
		t.Errorf("eine Stimme genügte: %q", halb.Zustand)
	}

	if len(bestand.gezeichn) != 0 {
		t.Error("der Beleg wurde nach einer Stimme gegengezeichnet")
	}

	// **Erst die zweite macht ihn beidseitig** — und sie bucht nichts nach.
	fertig, fehler := stelle.Bestaetigen(
		t.Context(), derVorgang, string(derAusweis(t).Kennung()), derFreibrief)
	if fehler != nil {
		t.Fatalf("zweite Bestätigung: %v", fehler)
	}

	if fertig.Zustand != domain.HandelBeidseitigVerbucht {
		t.Errorf("der Vorgang steht auf %q", fertig.Zustand)
	}

	if len(bestand.gebucht) != 1 {
		t.Errorf("es wurde %dmal gebucht", len(bestand.gebucht))
	}

	// **Und gegengezeichnet wird hier nicht**
	// (#1650).
	//
	// **Hier stand `!= 1`**, und das war der Befund: die zweite Stimme kam
	// von einem Menschen an **diesem** Fenster, und daraufhin sagte der
	// Beleg „auch die Gegenseite hat unterschrieben". Gegengezeichnet wird,
	// wenn ein **geprüfter** Beleg ankommt — und das tut die Post.
	if len(bestand.gezeichn) != 0 {
		t.Errorf("es wurde %dmal gegengezeichnet", len(bestand.gezeichn))
	}

	// **Der Beleg bleibt einseitig, bis einer von drüben kommt.**
	nachher, fehler := stelle.Beleg(t.Context(), derVorgang)
	if fehler != nil {
		t.Fatalf("Beleg: %v", fehler)
	}

	if nachher == nil || nachher.Beidseitig {
		t.Errorf("der Beleg ist %+v", nachher)
	}
}
