package service_test

// **Die zwei Befehle, die Mana machen** — Werkstück M3 der Regelmaschine
// (#1605).
//
// # Wogegen hier geprüft wird
//
//   - **`spieler.mana` ist Buchführung** und braucht keine Priorität — wie
//     `spieler.leben` daneben.
//   - **`objekt.mana` ist eine Manafähigkeit** und braucht sie (CR 605.3a).
//   - **Über den Sandkasten und nicht über die Domäne unmittelbar**: die
//     Abnahme des Issues lautet „ein Land tappt für Mana, der Vorrat steht im
//     Fenster" — das ist eine Aussage über den **Weg** (§3).

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Mana nachtragen geht über den Sandkasten.**
func TestManaGehtUeberDenSandkasten(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	ereignisse, danach, fehler := kasten.Auslegen(eineZugpartie(t),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art:      service.BefehlMana,
			Nutzlast: map[string]any{"art": "R", "um": float64(2)},
		})
	if fehler != nil {
		t.Fatalf("mana: %v", fehler)
	}

	switch {
	case danach.Sitze[0].Manavorrat.Zahl(domain.ManaRot) != 2:
		t.Errorf("der Vorrat steht bei %v", danach.Sitze[0].Manavorrat)
	case len(ereignisse) != 1:
		t.Fatalf("%d Ereignisse", len(ereignisse))
	case ereignisse[0].Oeffentlich["vorrat"] != "RR":
		t.Errorf("das Ereignis nennt %v", ereignisse[0].Oeffentlich)
	}
}

// **Auch wer nicht dran ist, trägt Mana nach** — wie beim Lebensstand.
func TestManaBrauchtKeinePrioritaet(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	if _, _, fehler := kasten.Auslegen(eineZugpartie(t), domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-2", Kennung: "b1",
		Art:      service.BefehlMana,
		Nutzlast: map[string]any{"art": "G", "um": float64(1)},
	}); fehler != nil {
		t.Errorf("mana ohne Priorität: %v", fehler)
	}
}

// **Eine Manaart, die es nicht gibt, wird abgewiesen** — CR 106.1b.
func TestEineUnbekannteManaartWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	if _, _, fehler := kasten.Auslegen(eineZugpartie(t), domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
		Art:      service.BefehlMana,
		Nutzlast: map[string]any{"art": "Q", "um": float64(1)},
	}); fehler == nil {
		t.Error("die Manaart Q kam durch")
	}
}

// **Ein Grundland tappt über den Sandkasten** — CR 305.6.
//
// **Das ist die Abnahme des Issues**: „Ein Land tappt für Mana, der Vorrat
// steht im Fenster, und er ist am Ende des Abschnitts leer."
func TestEinGrundlandTapptUeberDenSandkasten(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := einLandtischImDienst(t)

	_, danach, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: partie.Zug.Prioritaet, Kennung: "b1",
		Art:      service.BefehlLandmana,
		Nutzlast: map[string]any{"objekt": "sitz-1/land/0"},
	})
	if fehler != nil {
		t.Fatalf("landmana: %v", fehler)
	}

	if danach.Sitze[0].Manavorrat.Zahl(domain.ManaGruen) != 1 {
		t.Fatalf("der Vorrat steht bei %v", danach.Sitze[0].Manavorrat)
	}

	// **Und am Ende des Abschnitts ist er leer** (CR 500.5) — über den Weg,
	// den ein Mensch nimmt: alle passen.
	nachher, fehler := danach.Passen(danach.Zug.Prioritaet)
	if fehler != nil {
		t.Fatalf("passen: %v", fehler)
	}

	nachher, fehler = nachher.Passen(nachher.Zug.Prioritaet)
	if fehler != nil {
		t.Fatalf("zweites passen: %v", fehler)
	}

	if !nachher.Sitze[0].Manavorrat.Leer() {
		t.Errorf("der Vorrat steht nach dem Abschnitt bei %v",
			nachher.Sitze[0].Manavorrat)
	}
}

// **Wer nicht dran ist, tappt kein Land** — CR 605.3a.
func TestOhnePrioritaetTapptNiemandEinLand(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := einLandtischImDienst(t)

	if _, _, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-2", Kennung: "b1",
		Art:      service.BefehlLandmana,
		Nutzlast: map[string]any{"objekt": "sitz-1/land/0"},
	}); fehler == nil {
		t.Error("ein Land wurde ohne Priorität getappt")
	}
}

// einLandtischImDienst gibt eine Partie mit einem Wald auf dem Tisch.
func einLandtischImDienst(t *testing.T) domain.Partie {
	t.Helper()

	partie := eineZugpartie(t)

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)
	sitze[0].Deck = domain.Partiedeck{
		Abdruck: "abcd",
		Karten: []domain.Abzugskarte{{
			Kennung: "land", Name: "Forest", Menge: 1,
			Seiten: []domain.Abzugsseite{{Felder: map[string]string{
				"super type": "Basic Land", "sub type": "Forest",
			}}},
		}},
	}
	partie.Sitze = sitze

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-1/land/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneTisch, Karte: "land",
	})

	return partie
}

// **Jeder Fehlerweg der zwei Befehle sagt, was fehlt** — §8.
//
// **Sie stehen hier und nicht in der Domäne**, weil sie zur **Nutzlast**
// gehören: was der Kern bekommt, ist schon zerlegt; was hier scheitert, ist
// das Zerlegen.
func TestJederFehlerwegDerManabefehleSagtWasFehlt(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		art      string
		nutzlast map[string]any
		wort     string
	}{
		"Mana ohne Art": {
			art: service.BefehlMana, nutzlast: map[string]any{"um": float64(1)},
			wort: `"art"`,
		},
		"Mana ohne Menge": {
			art: service.BefehlMana, nutzlast: map[string]any{"art": "R"},
			wort: "um",
		},
		"Grundland ohne Objekt": {
			art: service.BefehlLandmana, nutzlast: map[string]any{},
			wort: `"objekt"`,
		},
	}

	kasten := service.NeuerSandkasten(festeMischung{})

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			partie := einLandtischImDienst(t)

			_, _, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
				PartieID: "p1", Absender: partie.Zug.Prioritaet,
				Kennung: "b1", Art: fall.art, Nutzlast: fall.nutzlast,
			})

			if fehler == nil {
				t.Fatal("es kam kein Fehler")
			}

			if !strings.Contains(fehler.Error(), fall.wort) {
				t.Errorf("der Fehler lautet %q", fehler)
			}
		})
	}
}
