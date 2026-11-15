package service_test

// **Was an einem Objekt zu tun ist** — P3
// (#2371),
// Befund C-10: die angebotenen Handlungen kannten die Zone nicht. Eine
// Handkarte bot *Tappen*, *Kämpfen gegen …* und *Opfern* an, und welche
// Handlung bei einem Klick primär ist, wusste nur das Fenster.

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einAuskunftsstand legt zum Handlungsstand einen Wald in die Hand von
// `sitz-1` und eine Karte in die Hand von `sitz-2`.
func einAuskunftsstand(t *testing.T) domain.Partie {
	t.Helper()

	partie := einHandlungsstand(t)

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)

	for stelle := range sitze {
		karten := append([]domain.Abzugskarte{}, sitze[stelle].Deck.Karten...)
		karten = append(karten, domain.Abzugskarte{
			Kennung: "wald", Name: "Forest", Menge: 1,
			Seiten: []domain.Abzugsseite{{Felder: map[string]string{
				"super type": "Basic Land", "sub type": "Forest",
			}}},
		})
		sitze[stelle].Deck.Karten = karten
	}

	partie.Sitze = sitze
	partie.Objekte = append(append([]domain.Spielobjekt{}, partie.Objekte...),
		domain.Spielobjekt{
			Kennung: "sitz-1/wald/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneHand, Karte: "wald",
		},
		domain.Spielobjekt{
			Kennung: "sitz-1/wald/1", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneTisch, Karte: "wald",
		},
		domain.Spielobjekt{
			Kennung: "sitz-2/wald/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-2", Eigner: "sitz-2",
			Zone: domain.ZoneHand, Karte: "wald",
		})

	return partie
}

// auskunftZu sucht die Auskunft zu einem Objekt — **über seine
// Sichtkennung** (#2480): unter ihr geht die Auskunft hinaus.
func auskunftZu(
	t *testing.T, partie domain.Partie, auskunft []domain.Objekthandlungen,
	objekt string,
) (domain.Objekthandlungen, bool) {
	t.Helper()

	name := objekt
	for _, gesucht := range partie.Objekte {
		if gesucht.Kennung == objekt {
			name = partie.Sichtkennung(gesucht)
		}
	}

	for _, eine := range auskunft {
		if eine.Objekt == name {
			return eine, true
		}
	}

	return domain.Objekthandlungen{}, false
}

func TestDieHandlungenKennenDieZone(t *testing.T) {
	t.Parallel()

	partie := einAuskunftsstand(t)
	kasten := service.NeuerSandkasten(festeMischung{})
	auskunft := service.Handlungsauskunft(
		partie, "sitz-1", kasten.Erlaubt(partie, "sitz-1"))

	tests := []struct {
		name    string
		objekt  string
		primaer string
		zweit   string
		mit     []string
		ohne    []string
	}{
		{
			name:   "ein Land in der Hand wird gespielt, nicht getappt",
			objekt: "sitz-1/wald/0", primaer: service.BefehlLandspielen,
			mit: []string{service.BefehlLandspielen, service.BefehlSchieben},
			ohne: []string{
				service.BefehlTappen, service.BefehlKaempfen,
				service.BefehlOpfern, service.BefehlWirken,
			},
		},
		{
			name:   "eine Kreatur in der Hand wird gewirkt",
			objekt: "sitz-1/hand/0", primaer: service.BefehlWirken,
			mit:  []string{service.BefehlWirken},
			ohne: []string{service.BefehlTappen, service.BefehlLandspielen},
		},
		{
			name:   "ein Land auf dem Tisch wird getappt und gibt Mana",
			objekt: "sitz-1/wald/1", primaer: service.BefehlTappen,
			zweit: service.BefehlLandmana,
			mit:   []string{service.BefehlTappen, service.BefehlOpfern},
			ohne:  []string{service.BefehlWirken, service.BefehlLandspielen},
		},
		{
			name:   "eine Kreatur auf dem Tisch wird getappt, Mana gibt sie nicht",
			objekt: "sitz-1/baer/0", primaer: service.BefehlTappen,
			mit:  []string{service.BefehlKaempfen},
			ohne: []string{service.BefehlLandmana},
		},
	}

	for _, fall := range tests {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			eine, da := auskunftZu(t, partie, auskunft, fall.objekt)
			if !da {
				t.Fatalf("zu %q steht keine Auskunft", fall.objekt)
			}

			if eine.Primaer != fall.primaer || eine.Zweit != fall.zweit {
				t.Errorf("primär %q, zweit %q", eine.Primaer, eine.Zweit)
			}

			for _, art := range fall.mit {
				if !slices.Contains(eine.Handlungen, art) {
					t.Errorf("%q fehlt in %v", art, eine.Handlungen)
				}
			}

			for _, art := range fall.ohne {
				if slices.Contains(eine.Handlungen, art) {
					t.Errorf("%q steht in %v", art, eine.Handlungen)
				}
			}
		})
	}
}

// **Keine verborgene Karteninformation** — eine Auskunft über eine fremde
// Handkarte oder die eigene Bibliothek verriete, ob dort ein Land liegt.
func TestDieAuskunftVerraetKeineVerborgeneKarte(t *testing.T) {
	t.Parallel()

	partie := einAuskunftsstand(t)
	kasten := service.NeuerSandkasten(festeMischung{})
	auskunft := service.Handlungsauskunft(
		partie, "sitz-1", kasten.Erlaubt(partie, "sitz-1"))

	if eine, da := auskunftZu(t, partie, auskunft, "sitz-2/wald/0"); da {
		t.Errorf("zur fremden Handkarte steht %+v", eine)
	}

	bibliothek, da := auskunftZu(t, partie, auskunft, "sitz-1/bib/a")
	if !da {
		t.Fatal("zur eigenen Bibliothek steht keine Auskunft")
	}

	if bibliothek.Primaer != "" ||
		slices.Contains(bibliothek.Handlungen, service.BefehlLandspielen) ||
		slices.Contains(bibliothek.Handlungen, service.BefehlWirken) {
		t.Errorf("die Bibliothek verrät ihre Karte: %+v", bibliothek)
	}
}

// **Nur, was der Sitz gerade darf** — ohne Priorität steht kein Wirken und
// kein Land an der Hand (CR 117.1).
func TestDieAuskunftFolgtDerErlaubnis(t *testing.T) {
	t.Parallel()

	partie := einAuskunftsstand(t)
	auskunft := service.Handlungsauskunft(
		partie, "sitz-1", []string{service.BefehlSchieben})

	eine, da := auskunftZu(t, partie, auskunft, "sitz-1/wald/0")
	if !da {
		t.Fatal("zum Land steht keine Auskunft")
	}

	if eine.Primaer != "" ||
		!slices.Equal(eine.Handlungen, []string{service.BefehlSchieben}) {
		t.Errorf("ohne Erlaubnis steht %+v", eine)
	}
}

// **Ein Commander wird aus der Kommandozone gewirkt** — CR 903.8 (Fassung
// 2026-09-25); ein Klick dort meint das Wirken.
func TestEinCommanderWirdAusDerKommandozoneGewirkt(t *testing.T) {
	t.Parallel()

	partie := einAuskunftsstand(t)
	partie.Objekte = append([]domain.Spielobjekt{}, partie.Objekte...)

	for stelle := range partie.Objekte {
		if partie.Objekte[stelle].Kennung == "sitz-1/hand/0" {
			partie.Objekte[stelle].Zone = domain.ZoneKommando
		}
	}

	kasten := service.NeuerSandkasten(festeMischung{})
	auskunft := service.Handlungsauskunft(
		partie, "sitz-1", kasten.Erlaubt(partie, "sitz-1"))

	eine, da := auskunftZu(t, partie, auskunft, "sitz-1/hand/0")
	if !da || eine.Primaer != service.BefehlWirken {
		t.Errorf("Auskunft %+v (da: %v)", eine, da)
	}
}
