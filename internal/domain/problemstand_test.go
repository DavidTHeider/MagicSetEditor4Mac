package domain

// **Ein Prüfstand sagt, wie weit er gilt** — PERF02
// (#2352):
// eine Zahl der Karten mit Problemen ist nur dann eine, wenn die Prüfung
// vollständig war und das Set seitdem dasselbe ist. Sonst ist sie veraltet
// oder eine Untergrenze — und eine abgebrochene Prüfung, die bis dahin nichts
// fand, ist **kein „0 Probleme"**.

import (
	"testing"
	"time"
)

func TestDerProblemstandSagtWieWeitErGilt(t *testing.T) {
	t.Parallel()

	jetzt := time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)
	mitProblem := []Kartenproblem{{KarteID: 7, Befunde: []string{"Ebene 3 (name): die Schrift Beleren Bold ist nicht zu finden"}}}

	for name, fall := range map[string]struct {
		stand   Problemstand
		abdruck string
		urteil  Problemurteil
		zahl    int
		sicher  bool
	}{
		"nie geprüft":               {Problemstand{}, "a", ProblemeUngeprueft, 0, false},
		"vollständig und aktuell":   {Problemstand{SetID: 1, Abdruck: "a", Geprueft: 9, Gesamt: 9, Vollstaendig: true, Beendet: jetzt, Karten: mitProblem}, "a", ProblemeAktuell, 1, true},
		"vollständig, ohne Problem": {Problemstand{SetID: 1, Abdruck: "a", Geprueft: 9, Gesamt: 9, Vollstaendig: true, Beendet: jetzt}, "a", ProblemeAktuell, 0, true},
		"seitdem geändert":          {Problemstand{SetID: 1, Abdruck: "a", Geprueft: 9, Gesamt: 9, Vollstaendig: true, Beendet: jetzt, Karten: mitProblem}, "b", ProblemeVeraltet, 1, false},
		"abgebrochen ohne Fund":     {Problemstand{SetID: 1, Abdruck: "a", Geprueft: 3, Gesamt: 9, Beendet: jetzt}, "a", ProblemeUnvollstaendig, 0, false},
		"abgebrochen mit Fund":      {Problemstand{SetID: 1, Abdruck: "a", Geprueft: 3, Gesamt: 9, Beendet: jetzt, Karten: mitProblem}, "a", ProblemeUnvollstaendig, 1, false},
	} {
		urteil := fall.stand.Urteil(fall.abdruck)
		zahl, sicher := fall.stand.Zahl(fall.abdruck)

		if urteil != fall.urteil || zahl != fall.zahl || sicher != fall.sicher {
			t.Errorf("%s: %s, %d, sicher %v — erwartet %s, %d, sicher %v", name, urteil, zahl, sicher, fall.urteil, fall.zahl, fall.sicher)
		}
	}
}

// **Eine Zahl, die nicht sicher ist, ist nie null und sicher zugleich** —
// das falsche Null aus dem Issue, über jede Gestalt eines Stands.
func TestEinUnsichererStandBehauptetKeinNull(t *testing.T) {
	t.Parallel()

	for _, stand := range []Problemstand{
		{},
		{SetID: 1, Abdruck: "a", Geprueft: 0, Gesamt: 5},
		{SetID: 1, Abdruck: "a", Geprueft: 5, Gesamt: 5, Vollstaendig: true},
	} {
		for _, abdruck := range []string{"a", "b", ""} {
			zahl, sicher := stand.Zahl(abdruck)
			if zahl == 0 && sicher && stand.Urteil(abdruck) != ProblemeAktuell {
				t.Errorf("%+v bei %q: sicher null, obwohl %s", stand, abdruck, stand.Urteil(abdruck))
			}
		}
	}
}
