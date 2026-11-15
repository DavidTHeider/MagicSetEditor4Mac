package service_test

// **Wirken und Landspielen über den Sandkasten** — Werkstück M4 der
// Regelmaschine
// (#1606).
//
// # Wogegen hier geprüft wird
//
// **Dass beide durch denselben Weg gehen** wie alles andere — damit stehen
// sie an allen vier Flächen zugleich (§16). Was sie **tun**, steht in der
// Domäne; hier geht es darum, dass der Befehl ankommt und was er meldet.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Zauber geht über den Sandkasten auf den Stapel.**
func TestEinZauberGehtUeberDenSandkasten(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	partie, fehler := einZaubertisch(t).
		ManaHinzu("sitz-1", domain.ManaRot, 1)
	if fehler != nil {
		t.Fatalf("mana hinzu: %v", fehler)
	}

	ereignisse, danach, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: partie.Zug.Prioritaet, Kennung: "b1",
		Art: service.BefehlWirken,
		Nutzlast: map[string]any{
			"objekt": "sitz-1/blitz/0", "ziele": []any{"sitz-2"},
		},
	})
	if fehler != nil {
		t.Fatalf("wirken: %v", fehler)
	}

	oben := danach.StapelOben()

	switch {
	case oben == nil:
		t.Fatal("der Stapel ist leer")
	case len(oben.Ziele) != 1 || oben.Ziele[0].Kennung != "sitz-2":
		t.Errorf("die Ziele sind %v", oben.Ziele)
	case len(ereignisse) != 1:
		t.Fatalf("%d Ereignisse", len(ereignisse))
	case ereignisse[0].Oeffentlich["objekt"] != "sitz-1/blitz/0":
		t.Errorf("das Ereignis nennt %v", ereignisse[0].Oeffentlich)
	}
}

// **Ohne Ziele geht es auch** — die meisten Zauber haben keines.
func TestEinZauberOhneZieleGehtAuch(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	partie, fehler := einZaubertisch(t).
		ManaHinzu("sitz-1", domain.ManaRot, 1)
	if fehler != nil {
		t.Fatalf("mana hinzu: %v", fehler)
	}

	if _, _, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: partie.Zug.Prioritaet, Kennung: "b1",
		Art:      service.BefehlWirken,
		Nutzlast: map[string]any{"objekt": "sitz-1/blitz/0"},
	}); fehler != nil {
		t.Errorf("wirken ohne Ziele: %v", fehler)
	}
}

// **Ein Ziel, das keine Zeichenkette ist, wird abgewiesen** — §8.
func TestEinZielMussEineKennungSein(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	partie, fehler := einZaubertisch(t).
		ManaHinzu("sitz-1", domain.ManaRot, 1)
	if fehler != nil {
		t.Fatalf("mana hinzu: %v", fehler)
	}

	_, _, fehler = kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: partie.Zug.Prioritaet, Kennung: "b1",
		Art: service.BefehlWirken,
		Nutzlast: map[string]any{
			"objekt": "sitz-1/blitz/0", "ziele": []any{float64(7)},
		},
	})

	if fehler == nil || !strings.Contains(fehler.Error(), "Kennung") {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}

// **Ein Land geht über denselben Weg** — CR 116.2a.
func TestEinLandGehtUeberDenSandkasten(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := einZaubertisch(t)

	_, danach, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: partie.Zug.Prioritaet, Kennung: "b1",
		Art:      service.BefehlLandspielen,
		Nutzlast: map[string]any{"objekt": "sitz-1/wald/0"},
	})
	if fehler != nil {
		t.Fatalf("land spielen: %v", fehler)
	}

	if danach.Zug.Laender != 1 {
		t.Errorf("gezählt sind %d Länder", danach.Zug.Laender)
	}
}

// **Beide brauchen die Kennung des Objekts** — §8.
func TestWirkenUndLandBrauchenEinObjekt(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := einZaubertisch(t)

	for _, art := range []string{
		service.BefehlWirken, service.BefehlLandspielen,
	} {
		t.Run(art, func(t *testing.T) {
			t.Parallel()

			_, _, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
				PartieID: "p1", Absender: partie.Zug.Prioritaet,
				Kennung: "b1", Art: art, Nutzlast: map[string]any{},
			})

			if fehler == nil || !strings.Contains(fehler.Error(), `"objekt"`) {
				t.Errorf("der Fehler lautet %v", fehler)
			}
		})
	}
}

// einZaubertisch gibt eine Partie mit einem Blitz und einem Wald auf der
// Hand, in der ersten Hauptphase.
func einZaubertisch(t *testing.T) domain.Partie {
	t.Helper()

	partie := eineZugpartie(t)

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)
	sitze[0].Deck = domain.Partiedeck{
		Abdruck: "abcd",
		Karten: []domain.Abzugskarte{{
			Kennung: "blitz", Name: "Blitz", Menge: 1,
			Seiten: []domain.Abzugsseite{{Felder: map[string]string{
				"super type": "Instant", "casting cost": "[R]",
			}}},
		}, {
			Kennung: "wald", Name: "Forest", Menge: 1,
			Seiten: []domain.Abzugsseite{{Felder: map[string]string{
				"super type": "Basic Land", "sub type": "Forest",
			}}},
		}},
	}
	partie.Sitze = sitze

	for kennung, karte := range map[string]string{
		"sitz-1/blitz/0": "blitz", "sitz-1/wald/0": "wald",
	} {
		partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
			Kennung: kennung, Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneHand, Karte: karte,
		})
	}

	// **Bis zur ersten Hauptphase, über den Weg** (§3) — CR 116.2a lässt ein
	// Land nur dort zu, und eine gesetzte Phase prüfte die Vorrichtung.
	for range 20 {
		if partie.Zug.Abschnitt.Phase == domain.PhaseHauptVorKampf {
			return partie
		}

		danach, fehler := partie.Passen(partie.Zug.Prioritaet)
		if fehler != nil {
			t.Fatalf("passen: %v", fehler)
		}

		partie = danach
	}

	t.Fatal("die Hauptphase kam nicht")

	return partie
}

// **„ziele" muss eine Liste sein** — §8.
func TestZieleMuessenEineListeSein(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := einZaubertisch(t)

	_, _, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: partie.Zug.Prioritaet, Kennung: "b1",
		Art: service.BefehlWirken,
		Nutzlast: map[string]any{
			"objekt": "sitz-1/blitz/0", "ziele": "sitz-2",
		},
	})

	if fehler == nil || !strings.Contains(fehler.Error(), "keine Liste") {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}

// **Ein gescheitertes Wirken kommt als Fehler zurück und nicht als Ereignis.**
func TestEinGescheitertesWirkenGibtKeinEreignis(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	// **Ohne Mana** — die Kosten sind nicht zu decken (CR 601.2h).
	ereignisse, _, fehler := kasten.Auslegen(einZaubertisch(t),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art:      service.BefehlWirken,
			Nutzlast: map[string]any{"objekt": "sitz-1/blitz/0"},
		})

	switch {
	case fehler == nil:
		t.Fatal("ein Zauber kam ohne Mana durch")
	case len(ereignisse) != 0:
		t.Errorf("%d Ereignisse trotz Fehler", len(ereignisse))
	}
}
