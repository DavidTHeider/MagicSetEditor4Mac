package domain_test

// **Der Kampf** — Werkstück M6 der Regelmaschine
// (#1608).
//
// # Wogegen hier geprüft wird
//
//   - **Angreifer erklären** (CR 508.1a/f): ungetappte Kreaturen, die dem
//     aktiven Spieler gehören — und sie werden dabei getappt.
//   - **Blocker erklären** (CR 509.1a): ungetappte Kreaturen des
//     Verteidigers, jede auf einen Angreifer.
//   - **Kampfschaden** (CR 510.1a–d, 510.2): jede teilt ihre Stärke aus, und
//     zwar **gleichzeitig**.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einKampftisch gibt eine Partie im Angreiferschritt mit zwei Kreaturen.
func einKampftisch(t *testing.T) domain.Partie {
	t.Helper()

	partie := einePartie().ZugBeginnen("sitz-1")

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)

	for stelle := range sitze {
		sitze[stelle].Deck = domain.Partiedeck{
			Abdruck: "abcd",
			Karten: []domain.Abzugskarte{{
				Kennung: "grosser", Name: "Riese", Menge: 1,
				Seiten: []domain.Abzugsseite{{Felder: map[string]string{
					"super type": "Creature", "sub type": "Giant",
					"power": "3", "toughness": "3",
				}}},
			}, {
				Kennung: "kleiner", Name: "Bär", Menge: 1,
				Seiten: []domain.Abzugsseite{{Felder: map[string]string{
					"super type": "Creature", "sub type": "Bear",
					"power": "2", "toughness": "2",
				}}},
			}},
		}
	}

	partie.Sitze = sitze

	partie.Objekte = append(partie.Objekte,
		domain.Spielobjekt{
			Kennung: "sitz-1/riese/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneTisch, Karte: "grosser",
		},
		domain.Spielobjekt{
			Kennung: "sitz-2/baer/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-2", Eigner: "sitz-2",
			Zone: domain.ZoneTisch, Karte: "kleiner",
		})

	return bisZumSchritt(t, partie, domain.SchrittAngreifer)
}

// bisZumSchritt passt durch, bis dieser Schritt dasteht.
func bisZumSchritt(
	t *testing.T, partie domain.Partie, schritt domain.Schritt,
) domain.Partie {
	t.Helper()

	for range 24 {
		if partie.Zug.Abschnitt.Schritt == schritt {
			return partie
		}

		danach, fehler := partie.Passen(partie.Zug.Prioritaet)
		if fehler != nil {
			t.Fatalf("passen bei %+v: %v", partie.Zug.Abschnitt, fehler)
		}

		partie = danach
	}

	t.Fatalf("der Schritt %q kam nicht", schritt)

	return partie
}

// **Ein Angreifer wird erklärt und dabei getappt** — CR 508.1a, CR 508.1f.
func TestEinAngreiferWirdErklaertUndGetappt(t *testing.T) {
	t.Parallel()

	danach, fehler := einKampftisch(t).AngreiferErklaeren("sitz-1",
		[]domain.Angriff{{Kreatur: "sitz-1/riese/0", Ziel: "sitz-2"}})
	if fehler != nil {
		t.Fatalf("angreifer erklären: %v", fehler)
	}

	riese := objektMit(t, danach, "sitz-1/riese/0")

	switch {
	case !riese.Getappt:
		t.Error("der Angreifer steht ungetappt da (CR 508.1f)")
	case len(danach.Zug.Kampf.Angriffe) != 1:
		t.Errorf("erklärt sind %v", danach.Zug.Kampf.Angriffe)
	case !danach.Zug.Kampf.Erklaert:
		t.Error("die Erklärung ist nicht vermerkt")
	}
}

// **Keine Angreifer ist auch eine Erklärung** — CR 508.1a („if any").
func TestKeineAngreiferIstAuchEineErklaerung(t *testing.T) {
	t.Parallel()

	danach, fehler := einKampftisch(t).AngreiferErklaeren("sitz-1", nil)
	if fehler != nil {
		t.Fatalf("angreifer erklären: %v", fehler)
	}

	switch {
	case !danach.Zug.Kampf.Erklaert:
		t.Error("die leere Erklärung ist nicht vermerkt")
	case len(danach.Zug.Kampf.Angriffe) != 0:
		t.Errorf("erklärt sind %v", danach.Zug.Kampf.Angriffe)
	}
}

// **Wer getappt ist, greift nicht an** — CR 508.1a.
//
// **Das ist die Abnahme des Issues**: „Wer getappt ist, kann nicht angreifen,
// und das wird abgewiesen statt still ausgelassen."
func TestWerGetapptIstGreiftNichtAn(t *testing.T) {
	t.Parallel()

	partie := einKampftisch(t)

	objekte := make([]domain.Spielobjekt, len(partie.Objekte))
	copy(objekte, partie.Objekte)

	for stelle := range objekte {
		if objekte[stelle].Kennung == "sitz-1/riese/0" {
			objekte[stelle].Getappt = true
		}
	}

	partie.Objekte = objekte

	_, fehler := partie.AngreiferErklaeren("sitz-1",
		[]domain.Angriff{{Kreatur: "sitz-1/riese/0", Ziel: "sitz-2"}})

	if fehler == nil || !strings.Contains(fehler.Error(), "CR 508.1a") {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}

// **Ein Blocker wird erklärt und bleibt ungetappt** — CR 509.1a.
//
// **Blocken tappt nicht**: CR 508.1f gilt für das Angreifen, und für das
// Blocken steht nichts Entsprechendes da.
func TestEinBlockerWirdErklaertUndBleibtUngetappt(t *testing.T) {
	t.Parallel()

	partie := mitAngriff(t)

	danach, fehler := partie.BlockerErklaeren("sitz-2",
		[]domain.Block{{Kreatur: "sitz-2/baer/0", Angreifer: "sitz-1/riese/0"}})
	if fehler != nil {
		t.Fatalf("blocker erklären: %v", fehler)
	}

	baer := objektMit(t, danach, "sitz-2/baer/0")

	switch {
	case baer.Getappt:
		t.Error("der Blocker wurde getappt")
	case len(danach.Zug.Kampf.Bloecke) != 1:
		t.Errorf("erklärt sind %v", danach.Zug.Kampf.Bloecke)
	}
}

// **Eine 3/3 greift an, eine 2/2 blockt: beide markieren Schaden, die 2/2
// stirbt** — CR 510.1a/c/d, CR 510.2, CR 704.5g.
//
// **Das ist die Abnahme des Issues**, wörtlich.
func TestEineDreiDreiGegenEineZweiZwei(t *testing.T) {
	t.Parallel()

	// **Der Schaden ist schon ausgeteilt** — er ist eine zugbasierte
	// Handlung (CR 703.4m) und geschieht, sobald der Schritt beginnt.
	danach := mitBlock(t)

	riese := objektMit(t, danach, "sitz-1/riese/0")

	switch {
	// **Der Riese trägt zwei Schaden und lebt** — 2 < 3 (CR 704.5g).
	case riese.Schaden != 2:
		t.Errorf("der Riese trägt %d Schaden", riese.Schaden)
	case riese.Zone != domain.ZoneTisch:
		t.Errorf("der Riese liegt in %q", riese.Zone)
	}

	// **Der Bär ist tot** — drei Schaden auf zwei Widerstandskraft, und die
	// zustandsbasierte Aktion aus `M5` räumt ihn ab.
	gepruefft := danach.Zustandspruefung()

	if baer := objektMit(t, gepruefft,
		"sitz-2/baer/0"); baer.Zone != domain.ZoneFriedhof {
		t.Errorf("der Bär liegt in %q", baer.Zone)
	}
}

// **Ein ungeblockter Angreifer trifft den Spieler** — CR 510.1b.
func TestEinUngeblockterAngreiferTrifftDenSpieler(t *testing.T) {
	t.Parallel()

	partie, fehler := mitAngriff(t).BlockerErklaeren("sitz-2", nil)
	if fehler != nil {
		t.Fatalf("blocker erklären: %v", fehler)
	}

	vorher := partie.Sitze[1].Leben

	danach := bisZumSchritt(t, partie, domain.SchrittKampfschaden)

	if danach.Sitze[1].Leben != vorher-3 {
		t.Errorf("sitz-2 steht bei %d statt %d",
			danach.Sitze[1].Leben, vorher-3)
	}
}

// mitAngriff gibt eine Partie im **Blockerschritt**, in der der Riese
// angreift.
//
// **Der Schritt wird über das Passen erreicht** (§3): CR 508.2 gibt nach der
// Erklärung dem aktiven Spieler Priorität, und erst wenn alle passen, geht es
// weiter. Wer den Schritt hinschriebe, prüfte die Vorrichtung statt des Zugs.
func mitAngriff(t *testing.T) domain.Partie {
	t.Helper()

	partie, fehler := einKampftisch(t).AngreiferErklaeren("sitz-1",
		[]domain.Angriff{{Kreatur: "sitz-1/riese/0", Ziel: "sitz-2"}})
	if fehler != nil {
		t.Fatalf("angreifer erklären: %v", fehler)
	}

	return bisZumSchritt(t, partie, domain.SchrittBlocker)
}

// mitBlock gibt eine Partie im **Kampfschadenschritt**, in der der Bär den
// Riesen blockt.
func mitBlock(t *testing.T) domain.Partie {
	t.Helper()

	partie, fehler := mitAngriff(t).BlockerErklaeren("sitz-2",
		[]domain.Block{{Kreatur: "sitz-2/baer/0", Angreifer: "sitz-1/riese/0"}})
	if fehler != nil {
		t.Fatalf("blocker erklären: %v", fehler)
	}

	return bisZumSchritt(t, partie, domain.SchrittKampfschaden)
}

// **Was in diesem Zug gekommen ist, greift nicht an** — CR 302.6, CR 508.1a.
//
// **Die Einsatzverzögerung ist keine Fähigkeit, sondern eine Regel**: eine
// Kreatur muss seit dem Beginn des laufenden Zuges unter derselben Kontrolle
// stehen. Wer sie ausließe, ließe jede frisch gespielte Kreatur sofort
// angreifen — der sichtbarste Regelbruch, den ein Kartenspiel kennt.
func TestWasInDiesemZugKamGreiftNichtAn(t *testing.T) {
	t.Parallel()

	partie := einKampftisch(t)

	objekte := make([]domain.Spielobjekt, len(partie.Objekte))
	copy(objekte, partie.Objekte)

	for stelle := range objekte {
		if objekte[stelle].Kennung == "sitz-1/riese/0" {
			objekte[stelle].Seit = partie.Zug.Nummer
		}
	}

	partie.Objekte = objekte

	_, fehler := partie.AngreiferErklaeren("sitz-1",
		[]domain.Angriff{{Kreatur: "sitz-1/riese/0", Ziel: "sitz-2"}})

	if fehler == nil || !strings.Contains(fehler.Error(), "CR 302.6") {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}

// **Mit Eile geht es trotzdem** — CR 702.10b, den CR 508.1a namentlich nennt.
//
// **Das dritte von 194 Schlüsselwörtern**, und aus demselben Grund wie die
// ersten beiden: die Regel, die es braucht, **nennt es**. Ohne Eile könnte
// keine Karte mit diesem Wort tun, was auf ihr steht.
func TestMitEileGehtEsTrotzdem(t *testing.T) {
	t.Parallel()

	partie := einKampftisch(t)

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)

	karten := make([]domain.Abzugskarte, len(sitze[0].Deck.Karten))
	copy(karten, sitze[0].Deck.Karten)
	karten[0].Seiten = []domain.Abzugsseite{{Felder: map[string]string{
		"super type": "Creature", "sub type": "Giant",
		"power": "3", "toughness": "3", "rule text": "Haste",
	}}}
	sitze[0].Deck.Karten = karten
	partie.Sitze = sitze

	objekte := make([]domain.Spielobjekt, len(partie.Objekte))
	copy(objekte, partie.Objekte)

	for stelle := range objekte {
		if objekte[stelle].Kennung == "sitz-1/riese/0" {
			objekte[stelle].Seit = partie.Zug.Nummer
		}
	}

	partie.Objekte = objekte

	if _, fehler := partie.AngreiferErklaeren("sitz-1",
		[]domain.Angriff{{
			Kreatur: "sitz-1/riese/0", Ziel: "sitz-2",
		}}); fehler != nil {
		t.Errorf("eine Kreatur mit Eile griff nicht an: %v", fehler)
	}
}

// **Ein verrechneter Zauber bekommt seinen Zug aufgeprägt** — CR 302.6.
//
// **Sonst wäre der Lauf darüber eine Behauptung über ein Feld, das niemand
// setzt** (§10): `Seit` von Hand zu setzen prüft die Vorrichtung. Hier kommt
// eine Kreatur über den **Weg** auf den Tisch — gewirkt und verrechnet —, und
// danach steht die Zahl da.
func TestEineVerrechneteKreaturTraegtIhrenZug(t *testing.T) {
	t.Parallel()

	partie := einKampftisch(t)

	partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
		Kennung: "sitz-1/neuer/0", Art: domain.ObjektKarte,
		Beherrscher: "sitz-1", Eigner: "sitz-1",
		Zone: domain.ZoneStapel, Karte: "kleiner",
	})

	danach := partie.Verrechnen()

	neuer := objektMit(t, danach, "sitz-1/neuer/0")

	switch {
	case neuer.Zone != domain.ZoneTisch:
		t.Fatalf("die Kreatur liegt in %q", neuer.Zone)
	case neuer.Seit != danach.Zug.Nummer:
		t.Errorf("sie trägt Zug %d, die Partie steht bei %d",
			neuer.Seit, danach.Zug.Nummer)
	}
}
