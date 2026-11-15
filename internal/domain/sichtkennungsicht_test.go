package domain_test

// **Die Sicht trägt keine Kennung** — #2480, Vermerk 0201.
//
// Geprüft wird die Klasse und nicht ein Feld: in der ganzen Sicht, wie sie
// als JSON hinausgeht, steht **keine** echte Kennung — weder am Objekt noch in
// einem Verweis, einem Kampf oder einem Schlüssel.

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func einePartieMitGeheimnissen() domain.Partie {
	partie := domain.Partie{
		ID: "p1", Geheimnis: "geheim-2480",
		Sitze: []domain.Sitz{
			{Kennung: "bo", Commanderschaden: map[string]int{"cy/p0-1/0": 3}},
			{Kennung: "cy"},
		},
		Objekte: []domain.Spielobjekt{
			einSichtobjekt("bo/p0-12/0", domain.ZoneTisch, 0),
			einSichtobjekt("bo/p0-12/1", domain.ZoneHand, 0),
			einSichtobjekt("bo/p0-12/2", domain.ZoneBibliothek, 1),
			einSichtobjekt("bo/p0-13/0", domain.ZoneBibliothek, 2),
			einSichtobjekt("bo/p0-14/0", domain.ZoneTisch, 0),
			einSichtobjekt("cy/p0-1/0", domain.ZoneTisch, 0),
		},
	}

	partie.Objekte[4].HaengtAn = "bo/p0-12/0"
	partie.Objekte[4].Haufen = "bo/p0-12/0"
	partie.Objekte[5].Eigner, partie.Objekte[5].Beherrscher = "cy", "cy"
	partie.Zug.Kampf = domain.Kampfstand{
		Angriffe: []domain.Angriff{{Kreatur: "bo/p0-12/0", Ziel: "cy"}},
		Bloecke:  []domain.Block{{Kreatur: "cy/p0-1/0", Angreifer: "bo/p0-12/0"}},
	}

	return partie
}

func TestDieSichtTraegtKeineKennung(t *testing.T) {
	t.Parallel()

	partie := einePartieMitGeheimnissen()

	for _, fuer := range []string{"bo", "cy", ""} {
		roh, fehler := json.Marshal(partie.SichtVon(fuer, ""))
		if fehler != nil {
			t.Fatalf("Sicht für %q: %v", fuer, fehler)
		}

		for _, objekt := range partie.Objekte {
			if strings.Contains(string(roh), objekt.Kennung) {
				t.Errorf("die Sicht für %q nennt die Kennung %q", fuer, objekt.Kennung)
			}
		}
	}
}

func TestDieSichtNenntVerweiseUnterIhrerSichtkennung(t *testing.T) {
	t.Parallel()

	partie := einePartieMitGeheimnissen()
	sicht := partie.SichtVon("cy", "")
	name := partie.Sichtkennung(partie.Objekte[0])

	anhaengsel := sicht.Objekte[slices.IndexFunc(sicht.Objekte, func(o domain.Sichtobjekt) bool {
		return o.Kennung == partie.Sichtkennung(partie.Objekte[4])
	})]

	switch {
	case anhaengsel.HaengtAn != name || anhaengsel.Haufen != name:
		t.Errorf("das Anhängsel hängt an %q im Haufen %q, erwartet %q", anhaengsel.HaengtAn, anhaengsel.Haufen, name)
	case sicht.Zug.Kampf.Angriffe[0].Kreatur != name || sicht.Zug.Kampf.Angriffe[0].Ziel != "cy":
		t.Errorf("der Angriff: %+v", sicht.Zug.Kampf.Angriffe[0])
	case sicht.Zug.Kampf.Bloecke[0].Angreifer != name:
		t.Errorf("der Block: %+v", sicht.Zug.Kampf.Bloecke[0])
	case sicht.Sitze[0].Commanderschaden[partie.Sichtkennung(partie.Objekte[5])] != 3:
		t.Errorf("der Commanderschaden: %v", sicht.Sitze[0].Commanderschaden)
	}

	// **Die Partie selbst bleibt, was sie ist** — eine Sicht ist eine
	// Projektion und kein Griff in den Zustand.
	if partie.Zug.Kampf.Angriffe[0].Kreatur != "bo/p0-12/0" || partie.Sitze[0].Commanderschaden["cy/p0-1/0"] != 3 {
		t.Error("die Sicht hat den Zustand verändert")
	}
}

// **Geschwister stehen nicht mehr nebeneinander** — die Sicht nannte die
// Objekte in der Reihenfolge des Decks, und wer nur die Kennung ersetzte,
// ließe dieses Leck offen.
func TestVerborgeneObjekteStehenNachIhrerSichtkennung(t *testing.T) {
	t.Parallel()

	partie := einePartieMitGeheimnissen()
	sicht := partie.SichtVon("cy", "")

	var offen, verborgen []string

	for _, objekt := range sicht.Objekte {
		if objekt.Geheim {
			verborgen = append(verborgen, objekt.Kennung)
		} else {
			if len(verborgen) > 0 {
				t.Fatalf("ein offenes Objekt %q steht hinter verborgenen", objekt.Kennung)
			}

			offen = append(offen, objekt.Kennung)
		}
	}

	if len(verborgen) != 3 || !slices.IsSorted(verborgen) {
		t.Errorf("die verborgenen stehen %v", verborgen)
	}

	if len(offen) != 3 || offen[0] != partie.Sichtkennung(partie.Objekte[0]) {
		t.Errorf("die offenen stehen %v", offen)
	}
}
