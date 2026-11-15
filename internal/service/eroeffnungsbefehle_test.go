package service_test

// **Der Beginn einer Partie über den Sandkasten** — CR 103.3 und CR 103.5
// (Fassung 2026-09-25),
// #1839.
//
// # Wogegen hier geprüft wird
//
//   - **Der Aufbau lässt den Beginn ausstehen** und nimmt ihn nicht vorweg.
//   - **Die Mischung kommt aus der eingesetzten Quelle**, wie beim Mischen
//     einer Zone — und eine klagende Quelle hinterlässt keinen halben Stand.
//   - **Angeboten wird genau der nächste Schritt**, über `Erlaubt` und damit
//     an allen vier Flächen (§16) — auch dem, der keine Priorität hat.
//   - **Nach einem Schritt des Beginns wird nicht geprüft**: CR 704.3 prüft,
//     wenn jemand Priorität bekäme, und der Griff ins Leere wirkt erst dann
//     (CR 104.3c).

import (
	"errors"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// deckkarten ist, wie viele Karten ein Deck aus [einPartieaufbau] hat — der
// Aufbau, über den [eineAufgebautePartie] den Beginn ausstehen lässt.
const deckkarten = 6

// reihe ist die Mischung, die nichts umordnet — über ein ganzes Deck.
func reihe() []int {
	return ungemischt(deckkarten)
}

// handDes gibt die Kennungen auf der Hand eines Sitzes.
func handDes(partie domain.Partie, wer string) []string {
	hand := []string{}

	for _, objekt := range partie.Objekte {
		if objekt.Beherrscher == wer && objekt.Zone == domain.ZoneHand {
			hand = append(hand, objekt.Kennung)
		}
	}

	return hand
}

// schritt legt einen Befehl des Beginns aus und bricht bei einem Fehler ab.
func schritt(
	t *testing.T, kasten *service.Sandkasten, partie domain.Partie,
	wer, art string, nutzlast map[string]any,
) ([]domain.Ereignis, domain.Partie) {
	t.Helper()

	ereignisse, danach, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: wer, Kennung: art, Art: art,
		Nutzlast: nutzlast,
	})
	if fehler != nil {
		t.Fatalf("%s: %v", art, fehler)
	}

	return ereignisse, danach
}

// angebotenerBeginn gibt die Schritte des Beginns, die `Erlaubt` nennt.
func angebotenerBeginn(
	kasten *service.Sandkasten, partie domain.Partie, wer string,
) []string {
	angeboten := []string{}

	for _, art := range kasten.Erlaubt(partie, wer) {
		if slices.Contains(service.Eroeffnungsbefehle(), art) {
			angeboten = append(angeboten, art)
		}
	}

	return angeboten
}

func TestDerAufbauLaesstDenBeginnAusstehen(t *testing.T) {
	t.Parallel()

	partie := eineAufgebautePartie(t)

	for _, sitz := range partie.Sitze {
		if sitz.Eroeffnung != domain.EroeffnungAusstehend {
			t.Errorf("%q steht bei %q", sitz.Kennung, sitz.Eroeffnung)
		}

		if hand := handDes(partie, sitz.Kennung); len(hand) != 0 {
			t.Errorf("%q hat schon eine Hand: %v", sitz.Kennung, hand)
		}
	}
}

// **Der ganze Beginn eines Sitzes, Schritt für Schritt** — und an jedem
// wird genau der nächste angeboten.
func TestDerBeginnGehtSchrittFuerSchritt(t *testing.T) {
	t.Parallel()

	// **Sechs Karten je Deck**: die Mischung umfasst die ganze Bibliothek,
	// beim Mulligan samt der zurückgemischten Hand.
	kasten := service.NeuerSandkasten(festeMischung{reihe: reihe()})
	partie := eineAufgebautePartie(t)

	// **Sitz 2 hat keine Priorität** und bekommt den Beginn trotzdem — er
	// liegt vor dem ersten Zug (CR 103.8).
	for _, wer := range []string{"sitz-1", "sitz-2"} {
		if angeboten := angebotenerBeginn(kasten, partie, wer); !slices.Equal(
			angeboten, []string{service.BefehlStarthand}) {
			t.Errorf("%s: angeboten %v", wer, angeboten)
		}
	}

	ereignisse, partie := schritt(t, kasten, partie, "sitz-2",
		service.BefehlStarthand, nil)

	if ereignisse[0].Oeffentlich["hand"] != float64(6) ||
		ereignisse[0].Oeffentlich["stand"] != string(domain.EroeffnungOffen) ||
		len(ereignisse[0].Privat) != 0 {
		t.Errorf("Ereignis der Starthand: %+v", ereignisse[0])
	}

	if angeboten := angebotenerBeginn(kasten, partie, "sitz-2"); !slices.Equal(
		angeboten, []string{service.BefehlMulligan, service.BefehlBehalten}) {
		t.Errorf("nach der Starthand angeboten %v", angeboten)
	}

	ereignisse, partie = schritt(t, kasten, partie, "sitz-2",
		service.BefehlMulligan, nil)

	if ereignisse[0].Oeffentlich["mulligans"] != float64(1) ||
		ereignisse[0].Oeffentlich["unterlegen"] != float64(1) {
		t.Errorf("Ereignis des Mulligans: %+v", ereignisse[0])
	}

	if angeboten := angebotenerBeginn(kasten, partie, "sitz-2"); !slices.Equal(
		angeboten, []string{service.BefehlUnterlegen}) {
		t.Errorf("nach dem Mulligan angeboten %v", angeboten)
	}

	unten := handDes(partie, "sitz-2")[0]

	_, partie = schritt(t, kasten, partie, "sitz-2", service.BefehlUnterlegen,
		map[string]any{"objekte": []any{unten}})

	if slices.Contains(handDes(partie, "sitz-2"), unten) {
		t.Errorf("%q liegt noch auf der Hand", unten)
	}

	ereignisse, partie = schritt(t, kasten, partie, "sitz-2",
		service.BefehlBehalten, nil)

	if ereignisse[0].Oeffentlich["hand"] != float64(5) ||
		ereignisse[0].Oeffentlich["stand"] != string(domain.EroeffnungBehalten) {
		t.Errorf("Ereignis des Behaltens: %+v", ereignisse[0])
	}

	if angeboten := angebotenerBeginn(kasten, partie, "sitz-2"); len(angeboten) != 0 {
		t.Errorf("nach dem Behalten angeboten %v", angeboten)
	}
}

// **Während des Beginns wird nicht geprüft — danach schon.** Sechs Karten
// geben keine Starthand von sieben: der siebte Griff geht ins Leere, und
// verloren ist erst, wenn jemand Priorität bekäme (CR 104.3c, CR 704.3).
func TestDerGriffInsLeereWirktErstNachDemBeginn(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{reihe: reihe()})

	_, partie := schritt(t, kasten, eineAufgebautePartie(t), "sitz-1",
		service.BefehlStarthand, nil)

	if partie.Stand != domain.StandLaeuft || !partie.SitzVon("sitz-1").Leergezogen {
		t.Fatalf("nach der Starthand: Stand %q, leergezogen %v",
			partie.Stand, partie.SitzVon("sitz-1").Leergezogen)
	}

	// **Passen geht erst nach dem ganzen Beginn** (CR 103.8) — also
	// behalten beide zuerst.
	_, partie = schritt(t, kasten, partie, "sitz-1", service.BefehlBehalten, nil)
	_, partie = schritt(t, kasten, partie, "sitz-2", service.BefehlStarthand, nil)
	_, partie = schritt(t, kasten, partie, "sitz-2", service.BefehlBehalten, nil)

	if partie.Stand != domain.StandLaeuft {
		t.Fatalf("nach dem Beginn: Stand %q", partie.Stand)
	}

	_, partie = schritt(t, kasten, partie, "sitz-1", service.BefehlPassen, nil)

	if partie.Stand == domain.StandLaeuft {
		t.Error("der Griff ins Leere blieb nach dem Beginn ohne Folge")
	}
}

// **Eine klagende Quelle beginnt nichts** — kein halber Stand, keine Hand.
func TestEineKlagendeQuelleBeginntNichts(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{klagt: errZufallWeg})

	for _, art := range []string{service.BefehlStarthand, service.BefehlMulligan} {
		_, danach, fehler := kasten.Auslegen(eineAufgebautePartie(t), domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1", Art: art,
		})

		if !errors.Is(fehler, errZufallWeg) || len(danach.Objekte) != 0 {
			t.Errorf("%s: %v, %d Objekte danach", art, fehler, len(danach.Objekte))
		}
	}
}

// **Was unterzulegen ist, steht als Liste von Kennungen in der Nutzlast.**
func TestUnterlegenLiestEineListe(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{reihe: reihe()})

	_, partie := schritt(t, kasten, eineAufgebautePartie(t), "sitz-1",
		service.BefehlStarthand, nil)
	_, partie = schritt(t, kasten, partie, "sitz-1", service.BefehlMulligan, nil)

	for _, fall := range []struct {
		name     string
		nutzlast map[string]any
	}{
		{"keine Liste", map[string]any{"objekte": "sitz-1/k1"}},
		{"keine Kennung darin", map[string]any{"objekte": []any{float64(1)}}},
		{"gar nichts", nil},
	} {
		_, _, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art: service.BefehlUnterlegen, Nutzlast: fall.nutzlast,
		})

		if !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%s: %v", fall.name, fehler)
		}
	}
}
