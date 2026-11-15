package app_test

// **Eine Zeile ohne Ausführung heißt `nonfoil`, auch im Fenster** — C01
// (#2226).
//
// Der Vertragsrundgang fand `finish: ""` in `ListeDeuten`, wo das Fenster
// `nonfoil | foil | etched` erwartet (`antwortenlisteneinfuhr.ts`). Leer
// heißt in der Domäne `nonfoil` (`domain.Exemplar.Ausfuehrung`); die
// Sammlung nennt es so, und die Deutung tat es bis dahin nicht.

import "testing"

func TestEineZeileOhneAusfuehrungIstNonfoil(t *testing.T) {
	deutung, err := einfuhrnaht(t).Deuten(probekontext(t),
		"2 Erste Probe (probe) 1\n1 Erste Probe (probe) 1 *F*\n1 Erste Probe (probe) 1 (etched)")
	if err != nil {
		t.Fatalf("Deuten: %v", err)
	}

	if len(deutung.Zeilen) != 3 {
		t.Fatalf("%d Zeilen, erwartet 3", len(deutung.Zeilen))
	}

	for stelle, will := range []string{"nonfoil", "foil", "etched"} {
		if deutung.Zeilen[stelle].Finish != will {
			t.Errorf("Zeile %d: %q, erwartet %q", stelle+1, deutung.Zeilen[stelle].Finish, will)
		}
	}
}
