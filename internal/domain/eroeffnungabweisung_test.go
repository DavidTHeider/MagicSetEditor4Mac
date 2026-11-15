package domain_test

// **Was der Beginn einer Partie abweist** — die Kehrseite von
// `eroeffnung_test.go`, CR 103.5 (Fassung 2026-09-25). Abgewiesen wird mit
// [domain.ErrValidation], und die Partie bleibt, wie sie war.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestEineZweiteStarthandWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	einmal, _ := mitBibliotheken(2, 10).ZieheStarthand("sitz-1", umgekehrt(10)) //nolint:errcheck // der Lauf darüber hält es

	if _, fehler := einmal.ZieheStarthand("sitz-1", umgekehrt(3)); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("zweite Starthand: %v", fehler)
	}
}

func TestUnterlegenVerlangtGenauDieGeschuldetenHandkarten(t *testing.T) {
	t.Parallel()

	partie, _ := mitBibliotheken(2, 10).ZieheStarthand("sitz-1", umgekehrt(10)) //nolint:errcheck // dito
	partie, _ = partie.NimmMulligan("sitz-1", umgekehrt(10))                    //nolint:errcheck // dito

	hand := inZone(partie, "sitz-1", domain.ZoneHand)
	bibliothek := inZone(partie, "sitz-1", domain.ZoneBibliothek)

	for _, fall := range []struct {
		name    string
		objekte []string
	}{
		{"zu viele", []string{hand[0], hand[1]}},
		{"keine", []string{}},
		{"keine Handkarte", []string{bibliothek[0]}},
		{"fremde Karte", []string{"sitz-2/b01"}},
		{"unbekannte Karte", []string{"sitz-1/b99"}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := partie.Unterlege("sitz-1", fall.objekte); !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("unterlegen %v: %v", fall.objekte, fehler)
			}
		})
	}

	// **Zweimal dieselbe Karte ist eine Karte** — auch wenn zwei geschuldet
	// sind.
	zweimal, _ := partie.Unterlege("sitz-1", hand[:1])         //nolint:errcheck // der Lauf darüber hält es
	zweimal, _ = zweimal.NimmMulligan("sitz-1", umgekehrt(10)) //nolint:errcheck // dito
	doppelt := inZone(zweimal, "sitz-1", domain.ZoneHand)[0]

	if _, fehler := zweimal.Unterlege("sitz-1", []string{doppelt, doppelt}); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("dieselbe Karte zweimal: %v", fehler)
	}

	// **Ohne Schuld gibt es nichts unterzulegen.**
	vorher, _ := mitBibliotheken(2, 10).ZieheStarthand("sitz-1", umgekehrt(10)) //nolint:errcheck // dito
	if _, fehler := vorher.Unterlege("sitz-1", hand[:1]); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("unterlegen ohne Mulligan: %v", fehler)
	}
}

// **Bis die Starthand null Karten hätte, und dann nicht mehr** — CR 103.5.
func TestNachDerLeerenStarthandGibtEsKeinenMulliganMehr(t *testing.T) {
	t.Parallel()

	partie, _ := mitBibliotheken(2, 10).ZieheStarthand("sitz-1", umgekehrt(10)) //nolint:errcheck // dito

	// **Sieben gezählte Mulligans** geben eine Starthand von null Karten.
	for stelle := range partie.Sitze {
		if partie.Sitze[stelle].Kennung == "sitz-1" {
			partie.Sitze[stelle].Mulligans = domain.Starthand
		}
	}

	if _, fehler := partie.NimmMulligan("sitz-1", umgekehrt(10)); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Mulligan nach der leeren Starthand: %v", fehler)
	}
}

// **Ein unbekannter Sitz wird abgewiesen**, an jedem der vier Befehle.
func TestEinUnbekannterSitzBeginntNichts(t *testing.T) {
	t.Parallel()

	partie := mitBibliotheken(2, 10)

	if _, fehler := partie.ZieheStarthand("fremd", umgekehrt(10)); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Starthand: %v", fehler)
	}

	if _, fehler := partie.NimmMulligan("fremd", umgekehrt(10)); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Mulligan: %v", fehler)
	}

	if _, fehler := partie.Unterlege("fremd", []string{}); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("unterlegen: %v", fehler)
	}

	if _, fehler := partie.Behalte("fremd"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("behalten: %v", fehler)
	}
}

// **Eine Mischung, die keine Umordnung ist, wird abgewiesen** — sie verlöre
// Karten oder erfände welche. Sie kommt von außen, und außen kann irren.
func TestEineFalscheMischungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		name     string
		mischung []int
	}{
		{"zu kurz", umgekehrt(9)},
		{"doppelt", []int{0, 0, 1, 2, 3, 4, 5, 6, 7, 8}},
		{"jenseits der Bibliothek", []int{10, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
		{"negativ", []int{-1, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if _, fehler := mitBibliotheken(2, 10).ZieheStarthand("sitz-1", fall.mischung); !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("Starthand: %v", fehler)
			}

			partie, _ := mitBibliotheken(2, 10).ZieheStarthand("sitz-1", umgekehrt(10)) //nolint:errcheck // der Lauf darüber hält es

			if _, fehler := partie.NimmMulligan("sitz-1", fall.mischung); !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("Mulligan: %v", fehler)
			}
		})
	}
}

// **Der Beginn läuft vom Aufbau bis zum letzten Behalten** — CR 103.8
// (Fassung 2026-09-25): der erste Zug kommt nach dem Beginn, auch wenn noch
// niemand angefangen hat (#1851).
// Bis dahin galt er erst ab der ersten gezogenen Starthand (Befund 2 des
// Zweitblicks über Phase 6.9). Nur eine Partie von vorher läuft ohne ihn.
func TestDerBeginnLaeuftVomAufbauBisZumLetztenBehalten(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name    string
		staende []domain.Eroeffnungsstand
		laeuft  bool
	}{
		{"noch niemand hat angefangen", []domain.Eroeffnungsstand{
			domain.EroeffnungAusstehend, domain.EroeffnungAusstehend,
		}, true},
		{"einer hat gezogen, der andere nicht", []domain.Eroeffnungsstand{
			domain.EroeffnungOffen, domain.EroeffnungAusstehend,
		}, true},
		{"einer hat behalten, der andere noch nicht gezogen", []domain.Eroeffnungsstand{
			domain.EroeffnungBehalten, domain.EroeffnungAusstehend,
		}, true},
		{"einer hält noch eine offene Hand", []domain.Eroeffnungsstand{
			domain.EroeffnungBehalten, domain.EroeffnungOffen,
		}, true},
		{"alle haben behalten", []domain.Eroeffnungsstand{
			domain.EroeffnungBehalten, domain.EroeffnungBehalten,
		}, false},
		{"eine Partie von vorher", []domain.Eroeffnungsstand{"", ""}, false},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			partie := mitBibliotheken(len(fall.staende), 1)
			for stelle, stand := range fall.staende {
				partie.Sitze[stelle].Eroeffnung = stand
			}

			if partie.BeginnLaeuft() != fall.laeuft {
				t.Errorf("BeginnLaeuft = %v", !fall.laeuft)
			}
		})
	}
}

// **Wer gegangen ist, hält den Beginn nicht auf** — Befund 2 des Nachlaufs
// über Phase 6.9. Gibt zu dritt einer mitten im Beginn auf, spielen die
// zwei anderen weiter (CR 104.3a); sein ausstehender Beginn zählte trotzdem
// mit, und die Partie nahm danach nie wieder einen gewöhnlichen Befehl an.
func TestWerGegangenIstHaeltDenBeginnNichtAuf(t *testing.T) {
	t.Parallel()

	partie := mitBibliotheken(3, 1)
	partie.Sitze[0].Eroeffnung = domain.EroeffnungBehalten
	partie.Sitze[1].Eroeffnung = domain.EroeffnungBehalten
	partie.Sitze[2].Eroeffnung = domain.EroeffnungAusstehend

	// **Dass jemand zuhört**: solange der dritte dabei ist, läuft der Beginn.
	if !partie.BeginnLaeuft() {
		t.Fatal("der Beginn läuft nicht, obwohl einer noch nicht gezogen hat")
	}

	partie.Ausgang.Ausgeschieden = []domain.Ausgeschieden{
		{Sitz: "sitz-3", Grund: domain.GrundAufgabe},
	}

	if partie.BeginnLaeuft() {
		t.Error("der Beginn läuft weiter, obwohl nur der Gegangene aussteht")
	}
}
