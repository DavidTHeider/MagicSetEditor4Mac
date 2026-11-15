package service_test

// **Eine selbst gebaute Karte mit einer Fähigkeit wirkt im Spiel** — die
// erste Abnahme von `M8`
// (#1610),
// und das Issue nennt sie „den Punkt dieser Phase".
//
// # Warum dieser Lauf über den Sandkasten geht und nicht über den Deuter
//
// **Weil die Naht die Zusage ist.** Dass ein Skript Wirkungen aufschreibt,
// misst `kartenskript_test.go`; dass sie **am Tisch ankommen**, misst nur
// ein Lauf, der wirkt und verrechnet. Genau diese Naht hat in dieser Phase
// schon zweimal gefehlt — der Kampfstand aus `M6` und Leben und Mana aus
// `M3` waren gebaut und erreichten keine Fläche (§10).
//
// # Und warum er nicht `kartenskriptlauf_test.go` heißt
//
// **Weil es die schon gibt**, und sie handelt von `new_card` aus Phase 4.11
// — ein Skript, das eine **Karte** baut. Dies hier ist ein Skript, das im
// **Spiel** wirkt; zwei Gegenstände, zwei Dateien (§2).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Blitz zieht zwei Karten und nimmt drei Leben.**
func TestEineSelbstgebauteKarteWirktImSpiel(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := einSpielskriptstand(t, `draw(amount: 2)
		change_life(player: target, amount: -3)`)

	_, gewirkt, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
		Art: service.BefehlWirken,
		Nutzlast: map[string]any{
			"objekt": "sitz-1/blitz/0", "ziele": []any{"sitz-2"},
		},
	})
	if fehler != nil {
		t.Fatalf("wirken: %v", fehler)
	}

	// **Noch nichts** — CR 608.2c geschieht beim Verrechnen.
	if leben := lebenVon(t, gewirkt, "sitz-2"); leben != 20 {
		t.Errorf("sitz-2 hat schon %d Leben", leben)
	}

	danach := gewirkt.Verrechnen()

	if leben := lebenVon(t, danach, "sitz-2"); leben != 17 {
		t.Errorf("sitz-2 hat %d Leben", leben)
	}

	// **Drei statt einer**: der Blitz hat die Hand verlassen, und zwei
	// sind nachgekommen. Wie viele ohne Skript liegen bleiben, misst
	// [TestEineKarteOhneSkriptWirktWieBisher] — eine.
	if hand := handzahlVon(danach, "sitz-1"); hand != 3 {
		t.Errorf("%d Karten auf der Hand", hand)
	}

	// **Und der Blitz liegt im Friedhof** — CR 608.2n.
	if zone := objektZone(danach, "sitz-1/blitz/0"); zone !=
		domain.ZoneFriedhof {
		t.Errorf("der Blitz liegt in %q", zone)
	}
}

// **Und ein Skript ohne Ende hält die Partie nicht an** — die zweite
// Abnahme.
//
// **Der Befund kommt beim Verrechnen und nicht beim Ankündigen** —
// CR 608.2c (#1625).
// Hier stand er am Ereignis des Wirkens; dort saß die Rechnung, und sie saß
// eine Station zu früh.
func TestEinSkriptOhneEndeHaeltDiePartieNichtAn(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := einSpielskriptstand(t, "a := { a() }\na()")

	ereignisse, gewirkt, fehler := kasten.Auslegen(partie,
		domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art:      service.BefehlWirken,
			Nutzlast: map[string]any{"objekt": "sitz-1/blitz/0"},
		})
	if fehler != nil {
		t.Fatalf("wirken: %v", fehler)
	}

	// **Beim Ankündigen steht noch nichts da** — gelesen wird später.
	if _, da := ereignisse[0].Oeffentlich["befund"]; da {
		t.Errorf("der Befund stand schon beim Wirken: %+v",
			ereignisse[0].Oeffentlich)
	}

	// **Der Zauber liegt trotzdem auf dem Stapel** — was nicht gelesen
	// werden konnte, hat ihn nicht ungültig gemacht (§8).
	if zone := objektZone(gewirkt, "sitz-1/blitz/0"); zone !=
		domain.ZoneStapel {
		t.Errorf("der Blitz liegt in %q", zone)
	}

	// **Und beim Verrechnen trägt ihn der Strom.**
	danach := gewirkt.Verrechnen()

	if befund := danach.Skriptbefund(); !strings.Contains(
		befund, "Kartenskript") {
		t.Errorf("der Befund fehlt: %q", befund)
	}

	// **Der Blitz ist trotzdem verrechnet** — CR 608.2n.
	if zone := objektZone(danach, "sitz-1/blitz/0"); zone !=
		domain.ZoneFriedhof {
		t.Errorf("der Blitz liegt in %q", zone)
	}
}

// **Was zwischen Ankündigen und Verrechnen geschieht, geht in die Rechnung
// ein** — CR 608.2c (Fassung 2026-08-19),
// #1625.
//
// # Wie der Unterschied sichtbar wird
//
// **Über das Skript selbst.** Ein Kartenskript kann den Tisch heute nicht
// lesen — der Wortschatz sind dreizehn **Wirkungen** und keine Abfrage. Was
// sich messen lässt, ist deshalb der **Augenblick**: `Eigenschaften` liest
// die Karte des Sitzes, und wer sie zwischen den beiden Stationen ändert,
// bekommt beim Verrechnen die neue Rechnung.
//
// **Das ist kein künstlicher Fall.** Genau das tut ein Dauereffekt aus `M10`,
// und genau das ist der Unterschied, den CR 608.2c meint.
func TestDerStandBeimVerrechnenGehtInDieRechnungEin(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := einSpielskriptstand(t, "change_life(player: target, amount: -3)")

	_, gewirkt, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
		Art: service.BefehlWirken,
		Nutzlast: map[string]any{
			"objekt": "sitz-1/blitz/0", "ziele": []any{"sitz-2"},
		},
	})
	if fehler != nil {
		t.Fatalf("wirken: %v", fehler)
	}

	// **Zwischen den beiden Stationen ändert sich, was auf der Karte
	// steht.**
	gewirkt = mitAnderemSkript(t, gewirkt,
		"change_life(player: target, amount: -7)")

	danach := gewirkt.Verrechnen()

	// **Sieben und nicht drei**: gerechnet wurde mit dem Stand von jetzt.
	if leben := lebenVon(t, danach, "sitz-2"); leben != 13 {
		t.Errorf("sitz-2 hat %d Leben — gerechnet wurde mit dem Stand von "+
			"vorhin", leben)
	}
}

// mitAnderemSkript schreibt der Karte `blitz` ein anderes Spielskript.
func mitAnderemSkript(
	t *testing.T, partie domain.Partie, skript string,
) domain.Partie {
	t.Helper()

	sitze := append([]domain.Sitz{}, partie.Sitze...)
	karten := append([]domain.Abzugskarte{}, sitze[0].Deck.Karten...)

	for stelle := range karten {
		if karten[stelle].Kennung != "blitz" {
			continue
		}

		seiten := append([]domain.Abzugsseite{}, karten[stelle].Seiten...)
		felder := map[string]string{}

		for name, wert := range seiten[0].Felder {
			felder[name] = wert
		}

		felder["game script"] = skript
		seiten[0].Felder = felder
		karten[stelle].Seiten = seiten
	}

	sitze[0].Deck.Karten = karten
	partie.Sitze = sitze

	return partie
}

// **Und was die Regeln nicht erlauben, lässt den Rest gelten** — CR 101.2,
// die dritte Abnahme.
func TestEineUnmoeglicheAnweisungLaesstDenRestGelten(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := einSpielskriptstand(t, `destroy(object: "gibtesnicht")
		draw(amount: 1)`)

	_, gewirkt, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
		Art:      service.BefehlWirken,
		Nutzlast: map[string]any{"objekt": "sitz-1/blitz/0"},
	})
	if fehler != nil {
		t.Fatalf("wirken: %v", fehler)
	}

	danach := gewirkt.Verrechnen()

	// **Die zweite Anweisung gilt** — CR 101.2 lässt die eine aus und nicht
	// den Rest: eine mehr als die eine, die ohne Skript liegen bleibt.
	if hand := handzahlVon(danach, "sitz-1"); hand != 2 {
		t.Errorf("%d Karten auf der Hand", hand)
	}
}

// **Und eine Karte ohne Skript geht ihren gewohnten Weg** — die Gegenprobe
// (§10): der Normalfall ist eine Karte, die keines trägt.
func TestEineKarteOhneSkriptWirktWieBisher(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := einSpielskriptstand(t, "")

	ereignisse, gewirkt, fehler := kasten.Auslegen(partie,
		domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art:      service.BefehlWirken,
			Nutzlast: map[string]any{"objekt": "sitz-1/blitz/0"},
		})
	if fehler != nil {
		t.Fatalf("wirken: %v", fehler)
	}

	if _, da := ereignisse[0].Oeffentlich["befund"]; da {
		t.Errorf("ein Befund steht da: %+v", ereignisse[0].Oeffentlich)
	}

	danach := gewirkt.Verrechnen()

	// **Eine bleibt liegen** — die, die der Aufbau dorthin gelegt hat; der
	// Blitz ist gewirkt. Das ist die Grundlinie, gegen die die Läufe
	// darüber rechnen.
	if hand := handzahlVon(danach, "sitz-1"); hand != 1 {
		t.Errorf("%d Karten auf der Hand", hand)
	}
}

// einSpielskriptstand gibt eine Partie, in der sitz-1 einen Blitz mit diesem
// Skript auf der Hand hält.
func einSpielskriptstand(t *testing.T, skript string) domain.Partie {
	t.Helper()

	partie := einHandlungsstand(t)

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)

	felder := map[string]string{"name": "Blitz", "super type": "Instant"}
	if skript != "" {
		felder["game script"] = skript
	}

	karten := append([]domain.Abzugskarte{}, sitze[0].Deck.Karten...)
	karten = append(karten, domain.Abzugskarte{
		Kennung: "blitz", Name: "Blitz", Menge: 1,
		Seiten: []domain.Abzugsseite{{Felder: felder}},
	})

	sitze[0].Deck.Karten = karten
	partie.Sitze = sitze

	partie.Objekte = append(append([]domain.Spielobjekt{}, partie.Objekte...),
		domain.Spielobjekt{
			Kennung: "sitz-1/blitz/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneHand, Karte: "blitz",
		})

	return partie
}

// lebenVon gibt den Lebensstand eines Sitzes.
func lebenVon(t *testing.T, partie domain.Partie, wer string) int {
	t.Helper()

	sitz := partie.SitzVon(wer)
	if sitz == nil {
		t.Fatalf("den Sitz %q gibt es nicht", wer)
	}

	return sitz.Leben
}

// handzahlVon zählt die Handkarten eines Sitzes.
func handzahlVon(partie domain.Partie, wer string) int {
	anzahl := 0

	for _, objekt := range partie.Objekte {
		if objekt.Zone == domain.ZoneHand && objekt.Beherrscher == wer {
			anzahl++
		}
	}

	return anzahl
}

// objektZone gibt die Zone eines Objekts — leer, wo keines liegt.
func objektZone(partie domain.Partie, kennung string) domain.Zone {
	for _, objekt := range partie.Objekte {
		if objekt.Kennung == kennung {
			return objekt.Zone
		}
	}

	return ""
}

// **Ein zweites Land in einem Zug wird abgewiesen** — CR 305.2 über den
// Sandkasten.
//
// **Die Gegenprobe zum Weg daneben** (§10): ohne sie sagte der Befehl nur,
// dass ein Land **geht** — nicht, dass der Kern ein zweites abweist und dass
// die Abweisung durch den Sandkasten kommt.
func TestEinZweitesLandInEinemZugWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := einSpielskriptstand(t, "")
	partie.Zug.Laender = 1

	_, _, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
		Art:      service.BefehlLandspielen,
		Nutzlast: map[string]any{"objekt": "sitz-1/blitz/0"},
	})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("das zweite Land ging: %v", fehler)
	}
}
