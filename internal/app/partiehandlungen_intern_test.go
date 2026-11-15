package app

// **Die Handlungsauskunft kommt mit der Partie** — P3
// (#2371):
// je Objekt, was der fragende Sitz daran tun kann, an derselben Naht wie
// `erlaubt` — und ohne ein Wort über eine fremde Hand.

import (
	"testing"
)

func TestDiePartieNenntDieHandlungenJeObjekt(t *testing.T) {
	t.Parallel()

	gebunden, deckID := eineSpielende(t)

	angelegt := gebunden.PartieAnlegen(t.Context(), einEntwurf(deckID))
	if angelegt.Fehler != nil {
		t.Fatalf("PartieAnlegen: %+v", angelegt.Fehler)
	}

	beginnAbschliessen(t, ueberDieBindungen(t.Context(), gebunden), angelegt.Partie)

	gelesen := gebunden.Partie(t.Context(), angelegt.Partie.ID, "sitz-1")
	if gelesen.Fehler != nil {
		t.Fatalf("Partie: %+v", gelesen.Fehler)
	}

	if len(gelesen.Handlungen) == 0 {
		t.Fatal("zur Partie steht keine Handlung")
	}

	beherrscher := map[string]string{}
	zonen := map[string]string{}
	for _, objekt := range gelesen.Partie.Objekte {
		beherrscher[objekt.Kennung] = objekt.Beherrscher
		zonen[objekt.Kennung] = objekt.Zone
	}

	primaer := 0
	for _, eine := range gelesen.Handlungen {
		if beherrscher[eine.Objekt] != "sitz-1" && zonen[eine.Objekt] == "hand" {
			t.Errorf("zur fremden Handkarte %q steht %+v", eine.Objekt, eine)
		}

		if eine.Primaer != "" {
			primaer++
		}
	}

	if primaer == 0 {
		t.Errorf("keine Handlung ist primär: %+v", gelesen.Handlungen)
	}

	// **Wer nicht dran ist, bekommt keine Primärhandlung an der Hand** — die
	// Erlaubnis gilt auch hier (CR 117.1).
	gegenueber := gebunden.Partie(t.Context(), angelegt.Partie.ID, "sitz-2")
	for _, eine := range gegenueber.Handlungen {
		if zonen[eine.Objekt] == "hand" && eine.Primaer != "" {
			t.Errorf("ohne Priorität ist %q primär an %q", eine.Primaer, eine.Objekt)
		}
	}
}
