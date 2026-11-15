package service_test

// **Vier Gestalten über den Sandkasten** — Werkstück M12 der Regelmaschine
// (#1614).
//
// # Was hier geprüft wird
//
// **Dass jede über denselben Weg geht** und damit an allen vier Flächen
// zugleich dasteht (§16) — und dass die Nutzlast gelesen wird.

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einGestaltenstand gibt einen Handlungsstand, dessen Bär zwei Seiten hat.
func einGestaltenstand(t *testing.T) domain.Partie {
	t.Helper()

	partie := einHandlungsstand(t)

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)

	for stelle := range sitze {
		sitze[stelle].Deck.Karten[0].Seiten = append(
			sitze[stelle].Deck.Karten[0].Seiten,
			domain.Abzugsseite{Felder: map[string]string{
				"name": "Werbär", "super type": "Creature",
				"power": "4", "toughness": "4",
			}})
	}

	partie.Sitze = sitze

	return partie
}

// **Ein Spielstein entsteht über den Sandkasten** — CR 701.7a.
func TestSpielsteinbefehlErzeugtDenStein(t *testing.T) {
	t.Parallel()

	_, danach, fehler := service.NeuerSandkasten(festeMischung{}).Auslegen(
		einGestaltenstand(t), domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art: service.BefehlSpielstein,
			Nutzlast: map[string]any{
				"kennung": "sitz-1/soldat/0",
				"felder": map[string]any{
					"name": "Soldat", "super type": "Creature",
					"power": "1", "toughness": "1",
				},
			},
		})
	if fehler != nil {
		t.Fatalf("spielstein: %v", fehler)
	}

	stein := objektAus(t, danach, "sitz-1/soldat/0")

	if stein.Art != domain.ObjektSpielstein {
		t.Errorf("der Spielstein ist ein %q", stein.Art)
	}

	eigen, da := danach.Eigenschaften(stein)
	if !da || eigen.Name != "Soldat" {
		t.Errorf("der Spielstein heißt %q (gefunden: %v)", eigen.Name, da)
	}
}

// **Ohne Kennung vergibt der Kern sie** (#2480): das Fenster sieht nur
// Sichtkennungen und kann nicht wissen, welche Kennung frei ist — es schlug
// sonst immer dieselbe vor, und der zweite Spielstein prallte auf den ersten.
func TestOhneKennungVergibtDerKernSie(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	stand := einGestaltenstand(t)

	for runde, art := range []string{service.BefehlSpielstein, service.BefehlSpielstein, service.BefehlEmblem} {
		_, danach, fehler := kasten.Auslegen(stand, domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: fmt.Sprintf("b%d", runde),
			Art: art, Nutzlast: map[string]any{"felder": map[string]any{"name": "Soldat"}},
		})
		if fehler != nil {
			t.Fatalf("%s ohne Kennung: %v", art, fehler)
		}

		stand = danach
	}

	for _, kennung := range []string{"sitz-1/spielstein/0", "sitz-1/spielstein/1", "sitz-1/emblem/0"} {
		objektAus(t, stand, kennung)
	}
}

// **Ein Emblem entsteht über den Sandkasten** — CR 114.1.
func TestEmblembefehlErzeugtDasEmblem(t *testing.T) {
	t.Parallel()

	_, danach, fehler := service.NeuerSandkasten(festeMischung{}).Auslegen(
		einGestaltenstand(t), domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art:      service.BefehlEmblem,
			Nutzlast: map[string]any{"kennung": "sitz-1/emblem/0"},
		})
	if fehler != nil {
		t.Fatalf("emblem: %v", fehler)
	}

	if zone := objektAus(t, danach,
		"sitz-1/emblem/0").Zone; zone != domain.ZoneKommando {
		t.Errorf("das Emblem liegt in %q", zone)
	}
}

// **Die Kontrolle wechselt über den Sandkasten** — CR 110.2, CR 613.1b.
func TestBeherrschenbefehlWechseltDieKontrolle(t *testing.T) {
	t.Parallel()

	_, danach, fehler := service.NeuerSandkasten(festeMischung{}).Auslegen(
		einGestaltenstand(t), domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art: service.BefehlBeherrschen,
			Nutzlast: map[string]any{
				"objekt": "sitz-2/baer/0", "wer": "sitz-1",
			},
		})
	if fehler != nil {
		t.Fatalf("beherrschen: %v", fehler)
	}

	if wer := objektAus(t, danach,
		"sitz-2/baer/0").Beherrscher; wer != "sitz-1" {
		t.Errorf("den Bären beherrscht %q", wer)
	}
}

// **Und ohne die Angabe, wer** — der Befehl nennt sie, und er wird
// abgewiesen, wenn sie fehlt (§8).
func TestBeherrschenbefehlBrauchtDenNeuenBeherrscher(t *testing.T) {
	t.Parallel()

	_, _, fehler := service.NeuerSandkasten(festeMischung{}).Auslegen(
		einGestaltenstand(t), domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art:      service.BefehlBeherrschen,
			Nutzlast: map[string]any{"objekt": "sitz-2/baer/0"},
		})
	if fehler == nil || !strings.Contains(fehler.Error(), "wer") {
		t.Fatalf("ohne \"wer\": %v", fehler)
	}
}

// **Eine Karte wird umgewandelt** — CR 701.27a.
func TestUmwandelnbefehlDrehtDieKarte(t *testing.T) {
	t.Parallel()

	_, danach, fehler := service.NeuerSandkasten(festeMischung{}).Auslegen(
		einGestaltenstand(t), domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art:      service.BefehlUmwandeln,
			Nutzlast: map[string]any{"objekt": "sitz-1/baer/0"},
		})
	if fehler != nil {
		t.Fatalf("umwandeln: %v", fehler)
	}

	eigen, da := danach.Eigenschaften(objektAus(t, danach, "sitz-1/baer/0"))
	if !da || eigen.Name != "Werbär" {
		t.Errorf("oben liegt %q (gefunden: %v)", eigen.Name, da)
	}
}

// **Was die vier an Nutzlast brauchen.**
func TestWasDieGestaltbefehleBrauchen(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		art      string
		nutzlast map[string]any
		wort     string
	}{
		"ein Spielstein ohne Felder": {
			art:      service.BefehlSpielstein,
			nutzlast: map[string]any{"kennung": "sitz-1/soldat/0"},
			wort:     "felder",
		},
		"ein Spielstein mit einer Zahl im Feld": {
			art: service.BefehlSpielstein,
			nutzlast: map[string]any{
				"kennung": "sitz-1/soldat/0",
				"felder":  map[string]any{"power": 1.0},
			},
			wort: "power",
		},
		"ein Kontrollwechsel ohne Objekt": {
			art:      service.BefehlBeherrschen,
			nutzlast: map[string]any{"wer": "sitz-1"},
			wort:     "objekt",
		},
		"ein Umwandeln ohne Objekt": {
			art: service.BefehlUmwandeln, nutzlast: map[string]any{},
			wort: "objekt",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, _, fehler := service.NeuerSandkasten(festeMischung{}).Auslegen(
				einGestaltenstand(t), domain.Spielbefehl{
					PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
					Art: fall.art, Nutzlast: fall.nutzlast,
				})
			if fehler == nil ||
				!strings.Contains(fehler.Error(), fall.wort) {
				t.Errorf("%s: %v", name, fehler)
			}
		})
	}
}

// objektAus gibt das Objekt mit dieser Kennung aus einer Partie.
func objektAus(
	t *testing.T, partie domain.Partie, kennung string,
) domain.Spielobjekt {
	t.Helper()

	for _, objekt := range partie.Objekte {
		if objekt.Kennung == kennung {
			return objekt
		}
	}

	t.Fatalf("das Objekt %q gibt es nicht", kennung)

	return domain.Spielobjekt{}
}

// **Ein Kopieren braucht die Vorlage** — CR 707.2 nennt sie.
func TestKopierbefehlBrauchtDieVorlage(t *testing.T) {
	t.Parallel()

	_, _, fehler := service.NeuerSandkasten(festeMischung{}).Auslegen(
		einGestaltenstand(t), domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art:      service.BefehlKopieren,
			Nutzlast: map[string]any{"objekt": "sitz-1/baer/0"},
		})
	if fehler == nil || !strings.Contains(fehler.Error(), "vorlage") {
		t.Fatalf("ohne \"vorlage\": %v", fehler)
	}
}

// **Und mit ihr geht es durch** — CR 707.2.
func TestKopierbefehlMachtDieKopie(t *testing.T) {
	t.Parallel()

	_, danach, fehler := service.NeuerSandkasten(festeMischung{}).Auslegen(
		einGestaltenstand(t), domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art: service.BefehlKopieren,
			Nutzlast: map[string]any{
				"objekt": "sitz-1/baer/0", "vorlage": "sitz-2/baer/0",
			},
		})
	if fehler != nil {
		t.Fatalf("kopieren: %v", fehler)
	}

	if felder := objektAus(t, danach,
		"sitz-1/baer/0").Felder; len(felder) == 0 {
		t.Error("die Kopie trägt keine Felder")
	}
}

// **Die Bezeichnungen gehen über den Sandkasten** — CR 725, CR 726, CR 728.
func TestDieBezeichnungenGehenUeberDenSandkasten(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		art      string
		nutzlast map[string]any
		pruefe   func(*testing.T, domain.Partie)
	}{
		"Monarch": {
			art: service.BefehlMonarch, nutzlast: map[string]any{},
			pruefe: func(t *testing.T, partie domain.Partie) {
				if partie.Monarch != "sitz-1" {
					t.Errorf("Monarch ist %q", partie.Monarch)
				}
			},
		},
		"Initiative": {
			art: service.BefehlInitiative, nutzlast: map[string]any{},
			pruefe: func(t *testing.T, partie domain.Partie) {
				if partie.Initiative != "sitz-1" {
					t.Errorf("die Initiative hat %q", partie.Initiative)
				}
			},
		},
		"Strahlung": {
			art:      service.BefehlStrahlung,
			nutzlast: map[string]any{"um": 2.0},
			pruefe: func(t *testing.T, partie domain.Partie) {
				if sitz := partie.SitzVon(
					"sitz-1"); sitz == nil || sitz.Strahlung != 2 {
					t.Errorf("die Marken stehen bei %+v", sitz)
				}
			},
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, danach, fehler := service.NeuerSandkasten(
				festeMischung{}).Auslegen(einGestaltenstand(t),
				domain.Spielbefehl{
					PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
					Art: fall.art, Nutzlast: fall.nutzlast,
				})
			if fehler != nil {
				t.Fatalf("%s: %v", name, fehler)
			}

			fall.pruefe(t, danach)
		})
	}
}

// **Und die Strahlung braucht ihre Zahl** — §8.
func TestStrahlungsbefehlBrauchtDieZahl(t *testing.T) {
	t.Parallel()

	_, _, fehler := service.NeuerSandkasten(festeMischung{}).Auslegen(
		einGestaltenstand(t), domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art: service.BefehlStrahlung, Nutzlast: map[string]any{},
		})
	if fehler == nil || !strings.Contains(fehler.Error(), "um") {
		t.Fatalf("ohne \"um\": %v", fehler)
	}
}
