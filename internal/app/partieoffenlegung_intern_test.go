package app

// **Freigeben, zurücknehmen, vergessen — am Fenster** — Werkstück G5 von
// Phase 5.5 (#1462).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestOhneEinwilligungIstDieListeLeer(t *testing.T) {
	t.Parallel()

	antwort := freigabeanwendung(t).PartieOffenlegungen(t.Context(), "p1")

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %v", antwort.Fehler)
	case antwort.Offenlegungen == nil:
		t.Error("die Liste ist nil statt leer")
	case len(antwort.Offenlegungen) != 0:
		t.Errorf("erwartet nichts, bekam %v", antwort.Offenlegungen)
	}
}

func TestEineEinwilligungKommtInDerAntwortZurueck(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)

	antwort := anwendung.PartieOffenlegen(
		t.Context(), "p1", "anna", []string{"ben"})

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %v", antwort.Fehler)
	case len(antwort.Offenlegungen) != 1:
		t.Fatalf("erwartet eine, bekam %v", antwort.Offenlegungen)
	case antwort.Offenlegungen[0].Von != "anna":
		t.Errorf("der Geber ist %q", antwort.Offenlegungen[0].Von)
	}
}

// **Ein Fehler lässt eine leere Liste zurück und kein `nil`** — dieselbe
// Zusage auf beiden Wegen. Genau hier ist sie in Phase 5.4 zweimal gebrochen.
func TestEinFehlerBeimOffenlegenLaesstEineLeereListe(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		weg   string
		fahre func(*App) PartieoffenlegungAntwort
	}{
		"lesen": {weg: "Freigaben", fahre: func(a *App) PartieoffenlegungAntwort {
			return a.PartieOffenlegungen(t.Context(), "p1")
		}},
		"offenlegen": {weg: "Offenlegen", fahre: func(a *App) PartieoffenlegungAntwort {
			return a.PartieOffenlegen(t.Context(), "p1", "anna", []string{"ben"})
		}},
		"zurücknehmen": {weg: "Zuruecknehmen", fahre: func(a *App) PartieoffenlegungAntwort {
			return a.PartieOffenlegungZuruecknehmen(t.Context(), "p1", "anna")
		}},
		"vergessen": {weg: "Vergessen", fahre: func(a *App) PartieoffenlegungAntwort {
			return a.PartieVergessen(t.Context(), "p1")
		}},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anwendung := freigabeanwendung(t)
			anwendung.MitPartienaht(&klagendePartieleitung{
				weg: fall.weg, klagt: domain.ErrConflict,
			})

			antwort := fall.fahre(anwendung)

			switch {
			case antwort.Fehler == nil:
				t.Fatal("die Klage kam nicht an")
			case antwort.Offenlegungen == nil:
				t.Error("die Liste ist nil statt leer")
			}
		})
	}
}

// **Und jeder Weg fragt nach der Wirkung, die er braucht.**
//
// **Nicht „irgendetwas wurde abgewiesen"** (§10): ohne die zweite Zusicherung
// stünde hier eine Zusage über eine Rolle, die nichts durchsetzt.
func TestJederOffenlegungswegFragtNachSeinerWirkung(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		fahre   func(*App) PartieoffenlegungAntwort
		wirkung domain.Wirkung
	}{
		"lesen": {
			fahre: func(a *App) PartieoffenlegungAntwort {
				return a.PartieOffenlegungen(t.Context(), "p1")
			},
			wirkung: domain.WirkungLesen,
		},
		"offenlegen": {
			fahre: func(a *App) PartieoffenlegungAntwort {
				return a.PartieOffenlegen(t.Context(), "p1", "anna",
					[]string{"ben"})
			},
			wirkung: domain.WirkungSchreiben,
		},
		"zurücknehmen": {
			fahre: func(a *App) PartieoffenlegungAntwort {
				return a.PartieOffenlegungZuruecknehmen(
					t.Context(), "p1", "anna")
			},
			wirkung: domain.WirkungSchreiben,
		},
		"vergessen": {
			fahre: func(a *App) PartieoffenlegungAntwort {
				return a.PartieVergessen(t.Context(), "p1")
			},
			wirkung: domain.WirkungSchreiben,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wache := &abweisendeWache{}
			anwendung := freigabeanwendung(t)
			anwendung.MitWache(wache)

			antwort := fall.fahre(anwendung)

			switch {
			case antwort.Fehler == nil:
				t.Fatal("die Abweisung kam nicht an")
			case antwort.Offenlegungen == nil:
				t.Error("die Liste ist nil statt leer")
			case wache.gefragt != fall.wirkung:
				t.Errorf("gefragt wurde nach %q", wache.gefragt)
			}
		})
	}
}

// **Der Erfolgsweg jedes Schreibens gibt danach den Stand** — und die leere
// Empfängerliste kommt als Liste zurück und nicht als `null`.
func TestJederSchreibwegDerOffenlegungGibtDenStand(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)

	eingetragen := anwendung.PartieOffenlegen(
		t.Context(), "p1", "anna", []string{"ben"})
	if eingetragen.Fehler != nil {
		t.Fatalf("offenlegen: %+v", eingetragen.Fehler)
	}

	zurueck := anwendung.PartieOffenlegungZuruecknehmen(
		t.Context(), "p1", "anna")

	switch {
	case zurueck.Fehler != nil:
		t.Fatalf("zurücknehmen: %+v", zurueck.Fehler)
	case len(zurueck.Offenlegungen) != 0:
		t.Errorf("es steht noch etwas da: %v", zurueck.Offenlegungen)
	}

	vergessen := anwendung.PartieVergessen(t.Context(), "p1")

	switch {
	case vergessen.Fehler != nil:
		t.Fatalf("vergessen: %+v", vergessen.Fehler)
	case vergessen.Offenlegungen == nil:
		t.Error("die Liste ist nil statt leer")
	}
}

// **Eine Einwilligung ohne Empfänger kommt als leere Liste zurück.**
//
// **Und nicht als `null`** — ein Feld, das aus der Antwort verschwindet, lässt
// die Seite ohne Grund leer (§3). Die Go-Seite weist eine solche Gabe zwar ab;
// was aus dem Bestand kommt, ist trotzdem zu übersetzen.
func TestEineEinwilligungOhneEmpfaengerKommtAlsListeZurueck(t *testing.T) {
	t.Parallel()

	anwendung := freigabeanwendung(t)
	leitung := &schweigsameLeitung{
		freigaben: []domain.Offenlegung{{Von: "anna"}},
	}

	anwendung.MitPartienaht(leitung)

	antwort := anwendung.PartieOffenlegungen(t.Context(), "p1")

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %+v", antwort.Fehler)
	case len(antwort.Offenlegungen) != 1:
		t.Fatalf("erwartet eine, bekam %v", antwort.Offenlegungen)
	case antwort.Offenlegungen[0].An == nil:
		t.Error("die Empfängerliste ist nil statt leer")
	}
}
