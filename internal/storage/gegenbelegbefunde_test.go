package storage_test

// **Ein Gegenbeleg, der nicht passt, wird festgehalten** —
// #2500.
//
// # Was hier gemessen wird
//
// **Dass der Befund am eigenen Beleg steht**, aus demselben Lesen, und dass
// **der jüngste gilt** — ein Gegenüber, das immer neue Belege schickt, füllt
// den Bestand nicht. **Und dass es ohne eigenen Beleg keinen gibt**: wogegen
// wäre geprüft worden?

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// gegenbefundZu ist ein Befund über den Probehandel.
func gegenbefundZu(mangel domain.Gegenbelegmangel, grund string) domain.Gegenbelegbefund {
	return domain.Gegenbelegbefund{
		Vorgang: einHandel, Von: "bert", Abdruck: "a1", Mangel: mangel, Grund: grund,
	}
}

func TestEinAbgewiesenerGegenbelegStehtAmBeleg(t *testing.T) {
	probe := eineAbschlussprobe(t)
	ctx := probekontext(t)

	if fehler := probe.buch.Abschliessen(ctx, einAbschluss(probe.vierStueck(t, true))); fehler != nil {
		t.Fatalf("abschließen: %v", fehler)
	}

	vorher := time.Now().UTC().Add(-time.Second)
	if fehler := probe.buch.Abweisen(ctx, gegenbefundZu(domain.MangelZeilen, "zwei Zeilen statt drei")); fehler != nil {
		t.Fatalf("abweisen: %v", fehler)
	}

	beleg, fehler := probe.buch.Beleg(ctx, einHandel)
	if fehler != nil {
		t.Fatalf("Beleg: %v", fehler)
	}

	befund := beleg.Befund
	if befund == nil {
		t.Fatal("am Beleg steht kein Befund")
	}

	if befund.Von != "bert" || befund.Abdruck != "a1" || befund.Mangel != domain.MangelZeilen ||
		befund.Grund != "zwei Zeilen statt drei" || befund.Vorgang != einHandel {
		t.Errorf("der Befund ist %+v", *befund)
	}

	if befund.Empfangen.Before(vorher) {
		t.Errorf("empfangen %v, erwartet nach %v", befund.Empfangen, vorher)
	}

	// **Abgewiesen ist nicht gegengezeichnet.**
	if beleg.Beidseitig || beleg.Gegenzeichner != "" {
		t.Error("ein abgewiesener Gegenbeleg machte den Beleg beidseitig")
	}
}

func TestDerJuengsteBefundGilt(t *testing.T) {
	probe := eineAbschlussprobe(t)
	ctx := probekontext(t)

	if fehler := probe.buch.Abschliessen(ctx, einAbschluss(probe.vierStueck(t, true))); fehler != nil {
		t.Fatalf("abschließen: %v", fehler)
	}

	for _, befund := range []domain.Gegenbelegbefund{
		gegenbefundZu(domain.MangelZeilen, "erster"),
		gegenbefundZu(domain.MangelUnterschrift, "zweiter"),
	} {
		if fehler := probe.buch.Abweisen(ctx, befund); fehler != nil {
			t.Fatalf("abweisen: %v", fehler)
		}
	}

	beleg, fehler := probe.buch.Beleg(ctx, einHandel)
	if fehler != nil || beleg.Befund == nil {
		t.Fatalf("Beleg: %v, Befund %v", fehler, beleg)
	}

	if beleg.Befund.Grund != "zweiter" || beleg.Befund.Mangel != domain.MangelUnterschrift {
		t.Errorf("es gilt %+v, erwartet der zweite", *beleg.Befund)
	}
}

func TestOhneBelegGibtEsKeinenBefund(t *testing.T) {
	probe := eineAbschlussprobe(t)

	fehler := probe.buch.Abweisen(probekontext(t), domain.Gegenbelegbefund{
		Vorgang: "v-gibtesnicht", Von: "bert", Mangel: domain.MangelVorgang, Grund: "fremd",
	})
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("das Abweisen gab %v statt „gibt es nicht\"", fehler)
	}
}

func TestEinUngueltigerBefundKommtNichtInDenBestand(t *testing.T) {
	probe := eineAbschlussprobe(t)
	ctx := probekontext(t)

	if fehler := probe.buch.Abschliessen(ctx, einAbschluss(probe.vierStueck(t, true))); fehler != nil {
		t.Fatalf("abschließen: %v", fehler)
	}

	if fehler := probe.buch.Abweisen(ctx, gegenbefundZu("farbe", "falsch")); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("ein unbekannter Mangel gab %v", fehler)
	}

	beleg, fehler := probe.buch.Beleg(ctx, einHandel)
	if fehler != nil || beleg.Befund != nil {
		t.Errorf("Beleg %v mit Befund %v — erwartet keiner", fehler, beleg.Befund)
	}
}

// **Ohne Profil kein Befund** — wessen Beleg wäre gemeint?
func TestAbweisenNenntEinProfil(t *testing.T) {
	probe := eineAbschlussprobe(t)

	fehler := probe.buch.Abweisen(t.Context(), gegenbefundZu(domain.MangelZeilen, "ohne Profil"))
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("ein Befund ohne Profil gab %v", fehler)
	}
}
