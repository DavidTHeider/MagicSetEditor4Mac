package domain_test

// **Was ein Schlüsselwort im Kampf ändert** — Werkstück M14 von Phase 5.5
// (#1616).
//
// # Fünf Regeln, und jede steht in CR 702
//
//   - **Verteidiger** (CR 702.3b): greift nicht an.
//   - **Wachsamkeit** (CR 702.20b): das Angreifen tappt nicht.
//   - **Flugfähigkeit** (CR 702.9b): geblockt nur von Flugfähigkeit oder
//     Reichweite.
//   - **Reichweite** (CR 702.17b): blockt Flugfähiges.
//   - **Bedrohlich** (CR 702.111b): geblockt nur von zweien oder mehr.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Verteidiger greift nicht an** — CR 702.3b.
func TestEinVerteidigerGreiftNichtAn(t *testing.T) {
	t.Parallel()

	partie := einSchluesseltisch(t, map[string]string{
		"grosser": "Defender",
	})

	_, fehler := partie.AngreiferErklaeren("sitz-1", []domain.Angriff{
		{Kreatur: "sitz-1/riese/0", Ziel: "sitz-2"},
	})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Verteidiger griff an: %v", fehler)
	}

	// **Und die Meldung nennt die Regel** (§11).
	if fehler != nil && !contains(fehler.Error(), "702.3b") {
		t.Errorf("die Meldung nennt die Regel nicht: %v", fehler)
	}
}

// **Wachsamkeit tappt nicht** — CR 702.20b.
func TestWachsamkeitTapptNichtBeimAngreifen(t *testing.T) {
	t.Parallel()

	partie := einSchluesseltisch(t, map[string]string{
		"grosser": "Vigilance",
	})

	danach, fehler := partie.AngreiferErklaeren("sitz-1", []domain.Angriff{
		{Kreatur: "sitz-1/riese/0", Ziel: "sitz-2"},
	})
	if fehler != nil {
		t.Fatalf("angreifen: %v", fehler)
	}

	if objektVon(t, danach, "sitz-1/riese/0").Getappt {
		t.Error("die wachsame Kreatur ist getappt")
	}

	// **Die Gegenprobe**: ohne Wachsamkeit tappt sie (CR 508.1f).
	ohne, fehler := einSchluesseltisch(t, nil).AngreiferErklaeren(
		"sitz-1", []domain.Angriff{
			{Kreatur: "sitz-1/riese/0", Ziel: "sitz-2"},
		})
	if fehler != nil {
		t.Fatalf("angreifen: %v", fehler)
	}

	if !objektVon(t, ohne, "sitz-1/riese/0").Getappt {
		t.Error("die gewöhnliche Kreatur ist nicht getappt")
	}
}

// **Flugfähiges blockt nur, wer fliegt oder Reichweite hat** — CR 702.9b.
func TestFlugfaehigesWirdNurVonFlugOderReichweiteGeblockt(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name    string
		blocker string
		geht    bool
	}{
		{name: "ohne alles", blocker: "", geht: false},
		{name: "mit Flugfähigkeit", blocker: "Flying", geht: true},
		{name: "mit Reichweite", blocker: "Reach", geht: true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			partie := imBlockerschritt(t, map[string]string{
				"grosser": "Flying", "kleiner": fall.blocker,
			})

			_, fehler := partie.BlockerErklaeren("sitz-2", []domain.Block{
				{Kreatur: "sitz-2/baer/0", Angreifer: "sitz-1/riese/0"},
			})

			if fall.geht && fehler != nil {
				t.Fatalf("der Block ging nicht: %v", fehler)
			}

			if !fall.geht {
				if !errors.Is(fehler, domain.ErrValidation) {
					t.Fatalf("der Block ging: %v", fehler)
				}

				if !contains(fehler.Error(), "702.9b") {
					t.Errorf("die Meldung nennt die Regel nicht: %v", fehler)
				}
			}
		})
	}
}

// **Bedrohliches blockt nur, wer zu zweit ist** — CR 702.111b.
func TestBedrohlichesBrauchtZweiBlocker(t *testing.T) {
	t.Parallel()

	partie := imBlockerschritt(t, map[string]string{"grosser": "Menace"})

	_, fehler := partie.BlockerErklaeren("sitz-2", []domain.Block{
		{Kreatur: "sitz-2/baer/0", Angreifer: "sitz-1/riese/0"},
	})

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("einer allein blockte: %v", fehler)
	}

	if !contains(fehler.Error(), "702.111b") {
		t.Errorf("die Meldung nennt die Regel nicht: %v", fehler)
	}

	// **Und zu zweit geht es.**
	zweiter := domain.Spielobjekt{
		Kennung: "sitz-2/baer/1", Art: domain.ObjektKarte,
		Beherrscher: "sitz-2", Eigner: "sitz-2",
		Zone: domain.ZoneTisch, Karte: "kleiner",
	}

	partie.Objekte = append(append([]domain.Spielobjekt{}, partie.Objekte...),
		zweiter)

	if _, fehler := partie.BlockerErklaeren("sitz-2", []domain.Block{
		{Kreatur: "sitz-2/baer/0", Angreifer: "sitz-1/riese/0"},
		{Kreatur: "sitz-2/baer/1", Angreifer: "sitz-1/riese/0"},
	}); fehler != nil {
		t.Errorf("zu zweit ging es nicht: %v", fehler)
	}
}

// einSchluesseltisch gibt einen Kampftisch, dessen Karten die genannten
// Schlüsselwörter tragen.
//
// **Als Regeltext und nicht als Feld am Objekt**: die Fähigkeit steht
// gedruckt auf der Karte (CR 702.1), und genau das liest dieses Programm.
func einSchluesseltisch(t *testing.T, texte map[string]string) domain.Partie {
	t.Helper()

	partie := einKampftisch(t)

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)

	for stelle := range sitze {
		karten := make([]domain.Abzugskarte, len(sitze[stelle].Deck.Karten))
		copy(karten, sitze[stelle].Deck.Karten)

		for lauf := range karten {
			text, da := texte[karten[lauf].Kennung]
			if !da {
				continue
			}

			felder := map[string]string{}
			for name, wert := range karten[lauf].Seiten[0].Felder {
				felder[name] = wert
			}

			felder["rule text"] = text
			karten[lauf].Seiten = []domain.Abzugsseite{{Felder: felder}}
		}

		sitze[stelle].Deck.Karten = karten
	}

	partie.Sitze = sitze

	return partie
}

// imBlockerschritt erklärt den Riesen als Angreifer und geht weiter.
func imBlockerschritt(t *testing.T, texte map[string]string) domain.Partie {
	t.Helper()

	partie, fehler := einSchluesseltisch(t, texte).AngreiferErklaeren(
		"sitz-1", []domain.Angriff{
			{Kreatur: "sitz-1/riese/0", Ziel: "sitz-2"},
		})
	if fehler != nil {
		t.Fatalf("angreifen: %v", fehler)
	}

	return bisZumSchritt(t, partie, domain.SchrittBlocker)
}

// objektVon gibt ein Objekt der Partie — oder scheitert laut.
func objektVon(
	t *testing.T, partie domain.Partie, kennung string,
) domain.Spielobjekt {
	t.Helper()

	for _, objekt := range partie.Objekte {
		if objekt.Kennung == kennung {
			return objekt
		}
	}

	t.Fatalf("das Objekt %q liegt nicht da", kennung)

	return domain.Spielobjekt{}
}
