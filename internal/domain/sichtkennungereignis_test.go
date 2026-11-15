package domain_test

// **Ein Ereignis trägt die Sichtkennungen seines Augenblicks** — #2480,
// Vermerk 0201, Punkt 5.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func einBenanntesEreignis() (domain.Partie, domain.Ereignis) {
	partie := einePartieMitGeheimnissen()

	ereignis := domain.Ereignis{
		PartieID: "p1", Folge: 7, Art: "objekt.schieben",
		Oeffentlich: map[string]any{
			"sitz": "bo", "objekt": "bo/p0-12/1",
			"ziele": []any{"bo/p0-12/0", "cy"},
			"paare": []domain.Angriff{{Kreatur: "bo/p0-14/0", Ziel: "cy"}},
		},
		Privat: []domain.Ergaenzung{{
			Empfaenger: "bo",
			Inhalt: map[string]any{"karten": []any{
				map[string]any{"objekt": "bo/p0-13/0", "karte": "glut"},
			}},
		}},
	}

	return partie, domain.Benennen(partie, partie, []domain.Ereignis{ereignis})[0]
}

func TestEinEreignisNenntNachAussenKeineKennung(t *testing.T) {
	t.Parallel()

	partie, ereignis := einBenanntesEreignis()

	for _, fuer := range []string{"bo", "cy", ""} {
		roh, fehler := json.Marshal(ereignis.Fuer(fuer))
		if fehler != nil {
			t.Fatalf("für %q: %v", fuer, fehler)
		}

		for _, objekt := range partie.Objekte {
			if strings.Contains(string(roh), objekt.Kennung) {
				t.Errorf("das Ereignis für %q nennt %q: %s", fuer, objekt.Kennung, roh)
			}
		}

		if strings.Contains(string(roh), "#sichtnamen") {
			t.Errorf("die Zuordnung ging an %q hinaus: %s", fuer, roh)
		}
	}

	gezeigt := ereignis.Fuer("bo")

	switch {
	case gezeigt.Oeffentlich["objekt"] != partie.Sichtkennung(partie.Objekte[1]):
		t.Errorf("objekt heißt %v", gezeigt.Oeffentlich["objekt"])
	case gezeigt.Oeffentlich["sitz"] != "bo":
		t.Errorf("ein Sitz ist keine Kennung und blieb nicht: %v", gezeigt.Oeffentlich["sitz"])
	case len(gezeigt.Privat) != 1:
		t.Errorf("die eigene Ergänzung: %+v", gezeigt.Privat)
	}
}

// **Die Kette bleibt die der Kennungen**, und der Sichtabdruck ist derselbe,
// ob ihn der Wirt über den gespeicherten oder der Gast über den empfangenen
// rechnet.
func TestKetteUndSichtabdruckBleibenStimmig(t *testing.T) {
	t.Parallel()

	_, benannt := einBenanntesEreignis()

	roh := benannt
	roh.Privat = roh.Privat[:1]

	vorher, _ := roh.Abdruck("")                     //nolint:errcheck // verglichen wird unten
	nachher, _ := benannt.Abdruck("")                //nolint:errcheck // dito
	wirt, _ := benannt.Sichtabdruck("bo")            //nolint:errcheck // dito
	gast, _ := benannt.Fuer("bo").Sichtabdruck("bo") //nolint:errcheck // dito

	if vorher != nachher {
		t.Error("die Zuordnung änderte die Kette")
	}

	if wirt == "" || wirt != gast {
		t.Errorf("Wirt %q und Gast %q rechnen verschiedene Sichtabdrücke", wirt, gast)
	}
}

// **Was keine Kennung ist, bleibt** — nichts, ein Wahrheitswert, eine Zahl;
// und was sich nicht kodieren lässt, geht als nichts hinaus statt als roher
// Speicher.
func TestEinBenanntesEreignisLaesstAndereWerteStehen(t *testing.T) {
	t.Parallel()

	partie := einePartieMitGeheimnissen()

	ereignis := domain.Benennen(partie, partie, []domain.Ereignis{{
		PartieID: "p1", Folge: 8, Art: "probe",
		Oeffentlich: map[string]any{
			"objekt": "bo/p0-12/0", "leer": nil, "getappt": true,
			"menge": 3, "kanal": func() {},
		},
	}})[0].Fuer("cy")

	switch {
	case ereignis.Oeffentlich["leer"] != nil:
		t.Errorf("leer wurde %v", ereignis.Oeffentlich["leer"])
	case ereignis.Oeffentlich["getappt"] != true || ereignis.Oeffentlich["menge"] != 3:
		t.Errorf("getappt %v, menge %v", ereignis.Oeffentlich["getappt"], ereignis.Oeffentlich["menge"])
	case ereignis.Oeffentlich["kanal"] != nil:
		t.Errorf("ein unkodierbarer Wert ging hinaus: %v", ereignis.Oeffentlich["kanal"])
	}
}

// **Eingehend wird zurückgerechnet** — Vermerk 0201, Punkt 6.
func TestEineSichtkennungImBefehlMeintIhrObjekt(t *testing.T) {
	t.Parallel()

	partie := einePartieMitGeheimnissen()
	name := partie.Sichtkennung(partie.Objekte[1])

	nutzlast := partie.Echtnamen(map[string]any{
		"objekt": name, "nach": "tisch", "ziele": []any{name, "cy"},
	})

	if nutzlast["objekt"] != "bo/p0-12/1" || nutzlast["nach"] != "tisch" {
		t.Errorf("zurückgerechnet: %v", nutzlast)
	}

	ziele, _ := nutzlast["ziele"].([]any) //nolint:errcheck // geprüft wird die Länge
	if len(ziele) != 2 || ziele[0] != "bo/p0-12/1" || ziele[1] != "cy" {
		t.Errorf("die Ziele: %v", nutzlast["ziele"])
	}
}

// **Ein Gast über die Leitung nennt keine Kennung** — und wenn doch, sagt
// die Partie, welche.
func TestEineKennungImBefehlFaellt(t *testing.T) {
	t.Parallel()

	partie := einePartieMitGeheimnissen()

	if kennung, nennt := partie.NenntKennung(map[string]any{
		"objekt": partie.Sichtkennung(partie.Objekte[0]), "sitz": "bo",
	}); nennt {
		t.Errorf("eine Sichtkennung galt als Kennung %q", kennung)
	}

	if kennung, nennt := partie.NenntKennung(map[string]any{
		"karten": []any{map[string]any{"objekt": "bo/p0-13/0"}},
	}); !nennt || kennung != "bo/p0-13/0" {
		t.Errorf("die Kennung im Befehl: %q %v", kennung, nennt)
	}
}

// **Eine Absage nennt keine Kennung** — Vermerk 0201, Punkt 7: die Sätze der
// Domäne zitieren sie, und der Grund geht über die Leitung.
func TestEineAbsageNenntKeineKennung(t *testing.T) {
	t.Parallel()

	partie := einePartieMitGeheimnissen()

	satz := partie.OhneKennungen(`das Objekt "bo/p0-12/1" liegt nicht auf dem Tisch`)

	if strings.Contains(satz, "bo/p0-12/1") || !strings.Contains(satz, partie.Sichtkennung(partie.Objekte[1])) {
		t.Errorf("der Satz lautet %q", satz)
	}
}

// **Die Zuordnung ist kein Erfahrenes** — die Rücknahme fragt, ob jemand
// etwas nur für sich erfuhr, und die Zuordnung erfährt niemand.
func TestDieZuordnungIstNichtsErfahrenes(t *testing.T) {
	t.Parallel()

	partie := einePartieMitGeheimnissen()

	nurBenannt := domain.Benennen(partie, partie, []domain.Ereignis{{
		Art: "objekt.tappen", Oeffentlich: map[string]any{"objekt": "bo/p0-12/0"},
	}})[0]

	if nurBenannt.Erfahren() {
		t.Error("ein Ereignis nur mit Zuordnung galt als erfahren")
	}

	if _, mitKarten := einBenanntesEreignis(); !mitKarten.Erfahren() {
		t.Error("ein Ereignis mit eigener Ergänzung galt als nicht erfahren")
	}
}

// **Die nächste freie Kennung vergibt der Kern** — das Fenster sieht nur
// Sichtkennungen.
func TestDieFreieKennungUeberspringtVergebene(t *testing.T) {
	t.Parallel()

	partie := domain.Partie{Objekte: []domain.Spielobjekt{
		{Kennung: "bo/spielstein/0"}, {Kennung: "bo/spielstein/2"},
	}}

	if kennung := partie.FreieKennung("bo", "spielstein"); kennung != "bo/spielstein/1" {
		t.Errorf("frei ist %q", kennung)
	}
}
