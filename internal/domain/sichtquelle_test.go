package domain_test

// **Wessen Kartenkennung eine Sicht hergibt** — P9
// (#2377),
// Befund C-12, Vermerk 0198.
//
// Eine Kartenkennung ist eine Zeile **eines** Bestands. Am Tisch liegen aber
// Decks aus mehreren — dem des Wirts, dem eines Gastes, dem eines Verleihers.
// Eine Kennung, die in einem fremden Bestand nachgeschlagen wird, findet dort
// eine **andere** Karte (gleiche Zahl, anderer Inhalt) und zeichnet deren
// Bild. Deshalb: die Kennung nur, wo sie hingehört — und daneben eine
// undurchsichtige **Bildkennung**, über die jeder dieselbe Karte erkennt,
// ohne eine Zeile eines fremden Bestands zu sehen.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// zweiQuellen gibt einen Tisch mit dem Deck des Wirts an `sitz1` und dem
// Deck eines Gastes (`peer-bo`) an `sitz2` — beide nennen dieselbe Zeile 7
// und dieselbe Platzkennung `p1`, und es sind verschiedene Karten.
func zweiQuellen() domain.Partie {
	deck := func(quelle, name string) domain.Partiedeck {
		return domain.Partiedeck{
			Quelle: quelle, Abdruck: "abdruck-" + name,
			Karten: []domain.Abzugskarte{{
				Kennung: "p1", Name: name, Menge: 1,
				Seiten: []domain.Abzugsseite{{Name: name, KarteID: 7}},
			}},
		}
	}

	return domain.Partie{
		ID: "partie-1", Spiel: "magic", Wirt: "sitz1", Stand: domain.StandLaeuft,
		Sitze: []domain.Sitz{
			{Kennung: "sitz1", Name: "Ada", Deck: deck("", "Waldboden")},
			{Kennung: "sitz2", Name: "Bo", Deck: deck("peer-bo", "Bergpfad")},
		},
		Objekte: []domain.Spielobjekt{
			{Kennung: "o1", Art: domain.ObjektKarte, Eigner: "sitz1", Beherrscher: "sitz1", Zone: domain.ZoneTisch, Karte: "p1"},
			{Kennung: "o2", Art: domain.ObjektKarte, Eigner: "sitz2", Beherrscher: "sitz2", Zone: domain.ZoneTisch, Karte: "p1"},
			{Kennung: "o3", Art: domain.ObjektKarte, Eigner: "sitz2", Beherrscher: "sitz2", Zone: domain.ZoneHand, Karte: "p1"},
			// **Beherrscht von Ada, aber Bos Karte** — die Kennung kommt aus
			// dem Deck des Eigners, nicht des Beherrschers.
			{Kennung: "o4", Art: domain.ObjektKarte, Eigner: "sitz2", Beherrscher: "sitz1", Zone: domain.ZoneTisch, Karte: "p1"},
		},
	}
}

func TestDieKartenkennungNurAusDemEigenenBestand(t *testing.T) {
	t.Parallel()

	partie := zweiQuellen()

	// **Am Wirt**: seine Zeile 7 ist seine, die des Gastes nicht.
	wirt := sichtobjekte(partie, partie.Sicht("sitz1"))
	if wirt["o1"].KarteID != 7 {
		t.Errorf("die eigene Karte des Wirts: KarteID %d", wirt["o1"].KarteID)
	}

	if wirt["o2"].KarteID != 0 || wirt["o4"].KarteID != 0 {
		t.Errorf("die Karten des Gastes nennen am Wirt eine Zeile: %d, %d",
			wirt["o2"].KarteID, wirt["o4"].KarteID)
	}

	// **Für den Gast gerechnet**: umgekehrt.
	gast := sichtobjekte(partie, partie.SichtVon("sitz2", "peer-bo"))
	if gast["o1"].KarteID != 0 {
		t.Errorf("die Karte des Wirts nennt am Gast eine Zeile: %d", gast["o1"].KarteID)
	}

	if gast["o2"].KarteID != 7 || gast["o3"].KarteID != 7 || gast["o4"].KarteID != 7 {
		t.Errorf("die eigenen Karten des Gastes: %d, %d, %d",
			gast["o2"].KarteID, gast["o3"].KarteID, gast["o4"].KarteID)
	}
}

func TestDieBildkennungErkenntDieKarteOhneZeile(t *testing.T) {
	t.Parallel()

	partie := zweiQuellen()
	wirt := sichtobjekte(partie, partie.Sicht("sitz1"))
	gast := sichtobjekte(partie, partie.SichtVon("sitz2", "peer-bo"))

	if wirt["o1"].Bild == "" || wirt["o2"].Bild == "" {
		t.Fatalf("sichtbare Karten ohne Bildkennung: %q, %q", wirt["o1"].Bild, wirt["o2"].Bild)
	}

	// **Zwei verschiedene Karten, zwei Kennungen** — gleiche Zeile und
	// gleiche Platzkennung zählen nicht.
	if wirt["o1"].Bild == wirt["o2"].Bild {
		t.Error("Waldboden und Bergpfad tragen dieselbe Bildkennung")
	}

	// **Dieselbe Karte, dieselbe Kennung** — an jedem Ende und gleich, wer
	// sie beherrscht.
	if gast["o2"].Bild != wirt["o2"].Bild || wirt["o4"].Bild != wirt["o2"].Bild {
		t.Error("dieselbe Karte trägt verschiedene Bildkennungen")
	}

	// **Sie verrät keine Zeile**: die Zahl kommt in ihr nicht vor.
	if wirt["o1"].Bild == "7" || len(wirt["o1"].Bild) < 16 {
		t.Errorf("die Bildkennung %q ist keine undurchsichtige", wirt["o1"].Bild)
	}
}

// **Was geheim ist, trägt auch keine Bildkennung** — sie erkennte die Karte
// über einen Umweg.
func TestEinGeheimesObjektTraegtKeineBildkennung(t *testing.T) {
	t.Parallel()

	wirt := sichtobjekte(zweiQuellen(), zweiQuellen().Sicht("sitz1"))
	if wirt["o3"].Bild != "" || wirt["o3"].KarteID != 0 {
		t.Errorf("Bos Hand am Wirt: Bild %q, KarteID %d", wirt["o3"].Bild, wirt["o3"].KarteID)
	}

	zuschauer := sichtobjekte(zweiQuellen(), zweiQuellen().Sicht(""))
	if zuschauer["o3"].Bild != "" {
		t.Errorf("Bos Hand beim Zuschauer: Bild %q", zuschauer["o3"].Bild)
	}
}

// **Der Vorrat nennt nur Zeilen des eigenen Bestands.**
func TestDerVorratNurAusDemEigenenBestand(t *testing.T) {
	t.Parallel()

	partie := zweiQuellen()

	if vorrat := partie.Sicht("sitz2").Vorrat; len(vorrat) != 0 {
		t.Errorf("am Wirt nennt Bos Vorrat Zeilen: %v", vorrat)
	}

	if vorrat := partie.SichtVon("sitz2", "peer-bo").Vorrat; len(vorrat) != 1 || vorrat[0] != 7 {
		t.Errorf("am Gast fehlt Bos eigener Vorrat: %v", vorrat)
	}

	if vorrat := partie.SichtVon("sitz1", "peer-bo").Vorrat; len(vorrat) != 0 {
		t.Errorf("für den Gast gerechnet nennt Adas Vorrat Zeilen: %v", vorrat)
	}
}

// **Zwei Sitze mit gleich gebautem Deck tragen verschiedene Bildkennungen**
// — Zweitblick über Phase 7, Befund 3. Der Abdruck fasst Platz, Name, Druck
// und Menge, nicht die Gestaltung; zwei Gäste mit gleich benannten Karten
// in gleicher Ordnung teilten sonst beim Wirt einen Zwischenspeicher, und
// das zuerst geladene Bild stünde für beide.
func TestZweiSitzeMitGleichemAbdruckTragenVerschiedeneBilder(t *testing.T) {
	t.Parallel()

	partie := zweiQuellen()
	partie.Sitze[0].Deck.Abdruck = "derselbe"
	partie.Sitze[0].Deck.Quelle = "peer-cy"
	partie.Sitze[1].Deck.Abdruck = "derselbe"

	wirt := sichtobjekte(partie, partie.Sicht("sitz1"))
	if wirt["o1"].Bild == "" || wirt["o1"].Bild == wirt["o2"].Bild {
		t.Errorf("zwei Sitze, eine Bildkennung: %q, %q", wirt["o1"].Bild, wirt["o2"].Bild)
	}
}
