package domain_test

// **Was an nichts hängt** — Werkstück M5 der Regelmaschine
// (#1607).
//
// # Wogegen hier geprüft wird
//
//   - CR 704.5m — eine **Aura** an nichts geht in den Friedhof.
//   - CR 704.5n — eine **Ausrüstung** an nichts **löst sich** und bleibt
//     liegen.
//   - CR 704.5p — alles andere ebenso.
//   - CR 704.3 — und die Runde **wiederholt sich**: stirbt das Ding, an dem
//     eine Aura hängt, fällt sie in derselben Prüfung hinterher.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineAnhaengselpartie gibt einen Tisch mit einem Träger und einem
// Anhängsel.
func eineAnhaengselpartie(t *testing.T, untertyp string) domain.Partie {
	t.Helper()

	partie := eineKartenpartie()
	partie.Sitze[0].Deck.Karten = append(partie.Sitze[0].Deck.Karten,
		domain.Abzugskarte{
			Kennung: "k2", Name: "Anhängsel", Menge: 1,
			Seiten: []domain.Abzugsseite{{
				Name: "Anhängsel",
				Felder: map[string]string{
					"name": "Anhängsel", "casting cost": "[1]",
					"super type": "Enchantment", "sub type": untertyp,
				},
			}},
		})

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "haenger", Art: domain.ObjektKarte, Karte: "k2",
		Beherrscher: "sitz-2", Eigner: "sitz-1",
		Zone: domain.ZoneTisch, HaengtAn: partie.Objekte[0].Kennung,
	})

	return partie
}

// **Eine Aura an nichts geht in den Friedhof ihres Eigners** — CR 704.5m.
func TestEineAuraAnNichtsGehtInDenFriedhof(t *testing.T) {
	t.Parallel()

	partie := eineAnhaengselpartie(t, "Aura")
	partie.Objekte[1].HaengtAn = "gibtesnicht"

	aura := objektMit(t, partie.Zustandspruefung(), "haenger")

	switch {
	case aura.Zone != domain.ZoneFriedhof:
		t.Errorf("die Aura liegt in %q", aura.Zone)
	case aura.Beherrscher != "sitz-1":
		t.Errorf("im Friedhof beherrscht sie %q", aura.Beherrscher)
	case aura.HaengtAn != "":
		t.Errorf("sie hängt noch an %q", aura.HaengtAn)
	}
}

// **Eine Ausrüstung löst sich und bleibt liegen** — CR 704.5n.
func TestEineAusruestungAnNichtsLoestSich(t *testing.T) {
	t.Parallel()

	partie := eineAnhaengselpartie(t, "Equipment")
	partie.Objekte[1].HaengtAn = "gibtesnicht"

	geraet := objektMit(t, partie.Zustandspruefung(), "haenger")

	switch {
	case geraet.Zone != domain.ZoneTisch:
		t.Errorf("die Ausrüstung liegt in %q", geraet.Zone)
	case geraet.HaengtAn != "":
		t.Errorf("sie hängt noch an %q", geraet.HaengtAn)
	}
}

// **Und was gar nicht hängen darf, löst sich auch** — CR 704.5p.
func TestEineKreaturAnEtwasLoestSich(t *testing.T) {
	t.Parallel()

	partie := eineAnhaengselpartie(t, "Bear")
	partie.Objekte[1].HaengtAn = "gibtesnicht"

	kreatur := objektMit(t, partie.Zustandspruefung(), "haenger")

	switch {
	case kreatur.Zone != domain.ZoneTisch:
		t.Errorf("sie liegt in %q", kreatur.Zone)
	case kreatur.HaengtAn != "":
		t.Errorf("sie hängt noch an %q", kreatur.HaengtAn)
	}
}

// **Was an etwas Liegendem hängt, bleibt hängen** — die Gegenprobe.
func TestEinAnhaengselAnEinemPermanentBleibt(t *testing.T) {
	t.Parallel()

	partie := eineAnhaengselpartie(t, "Aura")

	if objektMit(t, partie.Zustandspruefung(), "haenger").HaengtAn == "" {
		t.Fatal("die Aura hat sich gelöst")
	}
}

// **Und an einem Ding im Friedhof hängt nichts** — CR 110.1: nur ein
// Permanent trägt etwas, und ein Permanent liegt auf dem Schlachtfeld.
func TestEineAuraAnEtwasImFriedhofFaellt(t *testing.T) {
	t.Parallel()

	partie := eineAnhaengselpartie(t, "Aura")
	partie.Objekte[0].Zone = domain.ZoneFriedhof

	if objektMit(t, partie.Zustandspruefung(), "haenger").
		Zone != domain.ZoneFriedhof {
		t.Fatal("die Aura hängt noch auf dem Tisch")
	}
}

// **Die Runde wiederholt sich** — CR 704.3.
//
// **Der Fall, für den es die Wiederholung gibt**: die Kreatur stirbt an ihrer
// Widerstandskraft (CR 704.5f), und **erst dadurch** hängt die Aura an
// nichts (CR 704.5m). Eine einzige Runde ließe sie auf dem Tisch stehen.
func TestDieRundeWiederholtSichBisNichtsMehrAnliegt(t *testing.T) {
	t.Parallel()

	partie := eineAnhaengselpartie(t, "Aura")
	partie.Sitze[0].Deck.Karten[0].Seiten[0].Felder = map[string]string{
		"name": "Winzling", "super type": "Creature", "sub type": "Goblin",
		"power": "1", "toughness": "1",
	}
	partie.Objekte[0].Zaehler = map[string]int{domain.MarkeMinusEins: 1}

	danach := partie.Zustandspruefung()

	switch {
	case objektMit(t, danach, partie.Objekte[0].Kennung).Zone != domain.ZoneFriedhof:
		t.Error("die Kreatur steht noch auf dem Tisch")
	case objektMit(t, danach, "haenger").Zone != domain.ZoneFriedhof:
		t.Error("die Aura steht noch auf dem Tisch — die Runde lief nur einmal")
	}
}

// **Ein Spielstein, der an nichts hängt, ist keine Aura** — er hat keine
// Karte und damit keine Typzeile (CR 111).
//
// **Er löst sich und bleibt liegen** (CR 704.5p), statt in den Friedhof zu
// gehen: ohne Typzeile weiß niemand, dass er eine Aura wäre, und eine
// Annahme auf Verdacht ist keine (§10).
func TestEinSpielsteinAnNichtsLoestSichNur(t *testing.T) {
	t.Parallel()

	partie := eineKartenpartie()
	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "stein", Art: domain.ObjektSpielstein, Name: "Klaue",
		Beherrscher: "sitz-1", Eigner: "sitz-1", Zone: domain.ZoneTisch,
		HaengtAn: "gibtesnicht",
	})

	stein := objektMit(t, partie.Zustandspruefung(), "stein")

	switch {
	case stein.Zone != domain.ZoneTisch:
		t.Errorf("der Spielstein liegt in %q", stein.Zone)
	case stein.HaengtAn != "":
		t.Errorf("er hängt noch an %q", stein.HaengtAn)
	}
}
