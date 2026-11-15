package domain_test

// **Welche Schritte nicht nur den Zustand bewegen** —
// #1637.
//
// # Was hier gemessen wird
//
// **Dass die drei Abschlussschritte benannt sind und die übrigen nicht.**
// Die Zustandsmatrix kennt zwölf Schritte; drei davon bewegen mehr als den
// Zustand — den Bestand und den Beleg —, und wer sie über die bloße
// Zustandsmaschine ginge, hätte den Zustand ohne seine Wirkung.
//
// **Und die Gegenprobe steht daneben**: jeder andere Schritt der Matrix ist
// **keiner**. Eine Aufzählung, die nur ihre Mitglieder prüft, wüchse
// unbemerkt.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDieDreiAbschlussschritteStehenFest(t *testing.T) {
	t.Parallel()

	erwartet := []domain.Handelsschritt{
		domain.SchrittVerbuchen, domain.SchrittAbgleichen,
		domain.SchrittBestaetigen,
	}

	gegeben := domain.Abschlussschritte()
	if len(gegeben) != len(erwartet) {
		t.Fatalf("es sind %d Abschlussschritte, nicht %d",
			len(gegeben), len(erwartet))
	}

	for stelle, schritt := range erwartet {
		if gegeben[stelle] != schritt {
			t.Errorf("an Stelle %d steht %q statt %q",
				stelle, gegeben[stelle], schritt)
		}

		if !domain.IstAbschlussschritt(schritt) {
			t.Errorf("%q gilt nicht als Abschlussschritt", schritt)
		}
	}
}

// **Und kein anderer Schritt der Matrix ist einer.**
//
// **Über die Matrix und nicht über eine zweite Liste** — eine Aufzählung
// daneben veraltete beim nächsten Übergang, und sie wäre still zu kurz.
func TestKeinAndererSchrittIstEinAbschlussschritt(t *testing.T) {
	t.Parallel()

	abschluss := map[domain.Handelsschritt]bool{}
	for _, schritt := range domain.Abschlussschritte() {
		abschluss[schritt] = true
	}

	for _, uebergang := range domain.Handelsmatrix() {
		if abschluss[uebergang.Schritt] {
			continue
		}

		if domain.IstAbschlussschritt(uebergang.Schritt) {
			t.Errorf("%q gilt als Abschlussschritt, steht aber nicht dabei",
				uebergang.Schritt)
		}
	}
}

// **Jeder Schritt der Matrix ist einer, und nichts sonst.**
//
// **Über die Matrix erhoben**: eine Liste daneben veraltete beim nächsten
// Übergang. Gebraucht wird sie an den Flächen — ein unbekannter Schritt ist
// dort eine Beanstandung und kein Konflikt, und das lässt sich nur sagen,
// wenn jemand weiß, welche es gibt.
func TestDieHandelsschritteSindDieDerMatrix(t *testing.T) {
	t.Parallel()

	aus := map[domain.Handelsschritt]bool{}
	for _, uebergang := range domain.Handelsmatrix() {
		aus[uebergang.Schritt] = true
	}

	schritte := domain.Handelsschritte()
	if len(schritte) != len(aus) {
		t.Fatalf("es sind %d Schritte, die Matrix kennt %d",
			len(schritte), len(aus))
	}

	for _, schritt := range schritte {
		if !aus[schritt] {
			t.Errorf("%q steht in der Liste, aber in keinem Übergang", schritt)
		}

		if !domain.IstHandelsschritt(schritt) {
			t.Errorf("%q gilt nicht als Schritt", schritt)
		}
	}

	// **Und was keiner ist, ist keiner** — ohne diesen Fall wäre ein
	// `return true` grün.
	if domain.IstHandelsschritt("tanzen") {
		t.Error("»tanzen« gilt als Handelsschritt")
	}
}
