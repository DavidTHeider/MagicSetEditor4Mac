package domain_test

// **Nach außen eine Sichtkennung statt der Kennung** — #2480, Vermerk 0201.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func einSichtobjekt(kennung string, zone domain.Zone, ordnung int) domain.Spielobjekt {
	return domain.Spielobjekt{
		Kennung: kennung, Art: domain.ObjektKarte, Beherrscher: "bo", Eigner: "bo",
		Zone: zone, Ordnung: ordnung, Karte: "p0-12",
	}
}

func TestEineSichtkennungVerraetNichtsVonDerKennung(t *testing.T) {
	t.Parallel()

	partie := domain.Partie{ID: "p1", Geheimnis: "geheim-2480"}
	erstes := einSichtobjekt("bo/p0-12/0", domain.ZoneHand, 0)
	zweites := einSichtobjekt("bo/p0-12/1", domain.ZoneHand, 0)

	a, b := partie.Sichtkennung(erstes), partie.Sichtkennung(zweites)

	switch {
	case a == "" || a == b:
		t.Errorf("zwei Geschwister heißen %q und %q", a, b)
	case strings.Contains(a, "p0-12") || strings.Contains(a, "bo/"):
		t.Errorf("die Sichtkennung %q trägt die Struktur der Kennung", a)
	case a != partie.Sichtkennung(erstes):
		t.Error("dieselbe Kennung im selben Wechsel heißt zweimal anders")
	}

	gewechselt := erstes
	gewechselt.Wechsel = 3

	if partie.Sichtkennung(gewechselt) == a {
		t.Error("nach einem Wechsel heißt das Objekt wie vorher")
	}

	// **Ohne das Geheimnis nicht nachzurechnen** — eine andere Partie mit
	// denselben Kennungen gibt andere Namen.
	fremd := domain.Partie{ID: "p1", Geheimnis: "ein-anderes"}
	if fremd.Sichtkennung(erstes) == a {
		t.Error("ein anderes Geheimnis gibt dieselbe Sichtkennung")
	}

	// **Kein Trennzeichen, das das Fenster benutzt** — `nutzlast.ts` packt
	// Paare mit „|".
	if strings.ContainsAny(a, "|/") {
		t.Errorf("die Sichtkennung %q trägt ein Trennzeichen", a)
	}
}

func TestEinVerborgenesObjektBekommtEinenNeuenWechsel(t *testing.T) {
	t.Parallel()

	vorher := domain.Partie{
		ID: "p1", Geheimnis: "geheim", Sichtfolge: 4,
		Objekte: []domain.Spielobjekt{
			einSichtobjekt("tisch-friedhof", domain.ZoneTisch, 0),
			einSichtobjekt("tisch-hand", domain.ZoneTisch, 0),
			einSichtobjekt("bleibt-hand", domain.ZoneHand, 0),
			einSichtobjekt("gemischt", domain.ZoneBibliothek, 1),
			einSichtobjekt("liegt-oben", domain.ZoneBibliothek, 2),
			einSichtobjekt("wird-verdeckt", domain.ZoneTisch, 0),
		},
	}

	danach := vorher
	danach.Objekte = []domain.Spielobjekt{
		einSichtobjekt("tisch-friedhof", domain.ZoneFriedhof, 0),
		einSichtobjekt("tisch-hand", domain.ZoneHand, 0),
		einSichtobjekt("bleibt-hand", domain.ZoneHand, 0),
		einSichtobjekt("gemischt", domain.ZoneBibliothek, 2),
		einSichtobjekt("liegt-oben", domain.ZoneBibliothek, 2),
		einSichtobjekt("wird-verdeckt", domain.ZoneTisch, 0),
		einSichtobjekt("neu", domain.ZoneHand, 0),
	}
	danach.Objekte[5].Verdeckt = true

	gewechselt := domain.Sichtwechsel(vorher, danach)

	if gewechselt.Sichtfolge != 5 {
		t.Fatalf("die Sichtfolge ist %d, erwartet 5", gewechselt.Sichtfolge)
	}

	will := map[string]int64{
		// **Zwischen offenen Zonen bleibt es dasselbe Ding.**
		"tisch-friedhof": 0,
		// **In eine verborgene Zone: neu.**
		"tisch-hand": 5,
		// **Was liegen bleibt, bleibt.**
		"bleibt-hand": 0,
		// **Eine neue Lage in einer verdeckten Zone: neu** — das Mischen.
		"gemischt":   5,
		"liegt-oben": 0,
		// **Verdeckt: neu.**
		"wird-verdeckt": 5,
		// **Und was es vorher nicht gab**, ebenso.
		"neu": 5,
	}

	for _, objekt := range gewechselt.Objekte {
		if objekt.Wechsel != will[objekt.Kennung] {
			t.Errorf("%s: Wechsel %d, erwartet %d", objekt.Kennung, objekt.Wechsel, will[objekt.Kennung])
		}
	}

	// **Ohne einen Wechsel bleibt die Sichtfolge**, und nichts wird neu.
	if ruhig := domain.Sichtwechsel(danach, danach); ruhig.Sichtfolge != danach.Sichtfolge {
		t.Errorf("ohne Änderung stieg die Sichtfolge auf %d", ruhig.Sichtfolge)
	}
}
