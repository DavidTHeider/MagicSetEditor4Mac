package domain_test

// **Ausgelöste Fähigkeiten** — Werkstück M9 der Regelmaschine
// (#1611).
//
// # Welche überhaupt auslösen können
//
// **Die eingebauten.** CR 603.1 beschreibt eine ausgelöste Fähigkeit als
// Kartentext, und Kartentexte gibt es ohne `M8` nicht. **Drei aber stehen
// im Regelwerk selbst** und haben ausdrücklich **keine Quelle** — die CR
// nennt das eine Ausnahme zu CR 113.8:
//
//   - **CR 725.2**, zwei für den Monarchen: er zieht im Endschritt, und wer
//     ihm Kampfschaden zufügt, wird selbst Monarch.
//   - **CR 728.1**, eine für die Radioaktivitätsmarken: mahlen in der
//     ersten Hauptphase, Leben je Nichtland, und die Marken gehen mit.
//
// **Damit trägt M9 seinen eigenen Gegenstand** — und M20 verliert drei
// Grenzen.
//
// # Und was daran die Regel ist
//
// **CR 603.2**: sie lösen aus und tun nichts. **CR 603.3**: sie kommen auf
// den Stapel, wenn das nächste Mal jemand Priorität bekäme. **CR 603.3b mit
// CR 101.4**: der aktive Spieler zuerst, dann reihum.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Monarch zieht im Endschritt** — CR 725.2 mit CR 603.3.
func TestDerMonarchZiehtImEndschritt(t *testing.T) {
	t.Parallel()

	partie, fehler := eineAusloeserpartie(t).Monarchwerden("sitz-1")
	if fehler != nil {
		t.Fatalf("monarch: %v", fehler)
	}

	vorher := len(handVon(partie, "sitz-1"))

	danach := bisZumEndschritt(t, partie)

	// **Die Fähigkeit liegt auf dem Stapel und hat noch nichts getan** —
	// CR 603.2.
	if len(handVon(danach, "sitz-1")) != vorher {
		t.Errorf("gezogen wurde schon beim Auslösen")
	}

	verrechnet := danach.Verrechnen()

	if gezogen := len(handVon(verrechnet, "sitz-1")); gezogen != vorher+1 {
		t.Errorf("auf der Hand liegen %d statt %d", gezogen, vorher+1)
	}
}

// **Und ohne Monarch löst nichts aus** — CR 725.1: es gibt keinen, bis ein
// Effekt einen macht.
func TestOhneMonarchLoestNichtsAus(t *testing.T) {
	t.Parallel()

	danach := bisZumEndschritt(t, eineAusloeserpartie(t))

	if anzahl := len(danach.Ausgeloest); anzahl != 0 {
		t.Errorf("%d Fähigkeiten warten: %+v", anzahl, danach.Ausgeloest)
	}

	for _, objekt := range danach.Objekte {
		if objekt.Zone == domain.ZoneStapel {
			t.Errorf("auf dem Stapel liegt %+v", objekt)
		}
	}
}

// **Eine verrechnete Fähigkeit hört auf zu existieren** — CR 608.2n mit
// CR 603.3: sie ist keine Karte und geht nirgendwohin.
func TestEineVerrechneteFaehigkeitVerschwindet(t *testing.T) {
	t.Parallel()

	partie, fehler := eineAusloeserpartie(t).Monarchwerden("sitz-1")
	if fehler != nil {
		t.Fatalf("monarch: %v", fehler)
	}

	danach := bisZumEndschritt(t, partie).Verrechnen()

	for _, objekt := range danach.Objekte {
		if objekt.Art == domain.ObjektFaehigkeit {
			t.Errorf("die Fähigkeit liegt noch in %q", objekt.Zone)
		}
	}
}

// **Wer dem Monarchen Kampfschaden zufügt, wird Monarch** — CR 725.2.
func TestWerDemMonarchenSchadetWirdMonarch(t *testing.T) {
	t.Parallel()

	partie, fehler := eineAusloeserpartie(t).Monarchwerden("sitz-2")
	if fehler != nil {
		t.Fatalf("monarch: %v", fehler)
	}

	partie.Zug = domain.Zugstand{
		Nummer: 3, Aktiv: "sitz-1", Prioritaet: "sitz-1",
		Abschnitt: domain.Abschnitt{
			Phase: domain.PhaseKampf, Schritt: domain.SchrittBlocker,
		},
		Kampf: domain.Kampfstand{
			Erklaert: true, Geblockt: true,
			Angriffe: []domain.Angriff{
				{Kreatur: "sitz-1/baer/0", Ziel: "sitz-2"},
			},
		},
	}

	danach, fehler := partie.Passen("sitz-1")
	if fehler != nil {
		t.Fatalf("passen: %v", fehler)
	}

	danach, fehler = danach.Passen(danach.Zug.Prioritaet)
	if fehler != nil {
		t.Fatalf("zweites passen: %v", fehler)
	}

	// **Noch nicht** — CR 603.2: die Fähigkeit hat ausgelöst und nichts
	// getan.
	if danach.Monarch != "sitz-2" {
		t.Errorf("Monarch ist schon %q", danach.Monarch)
	}

	verrechnet := danach.Verrechnen()

	if verrechnet.Monarch != "sitz-1" {
		t.Errorf("Monarch ist %q — der Bär gehört sitz-1",
			verrechnet.Monarch)
	}
}

// **Mehrere Fähigkeiten kommen in APNAP-Reihenfolge auf den Stapel** —
// CR 603.3b mit CR 101.4.
//
// **Der aktive Spieler legt zuerst** (CR 603.3b), und weil der Stapel von
// oben abgeräumt wird (CR 405.2), wird seine **zuletzt** verrechnet — die
// des Nichtaktiven liegt obenauf und geht zuerst.
//
// # Warum hier je Sitz eine steht
//
// **Weil zwei desselben Sitzes eine andere Regel sind.** Dann ordnet **er**,
// und der Kern wartet auf seine Antwort (CR 603.3b, zweiter Satz) — das
// steht in `ausloeserwahl_test.go`. Hier geht es allein um die Reihenfolge
// **zwischen** Sitzen, und die entscheidet die Regel.
func TestMehrereFaehigkeitenKommenInAPNAPAufDenStapel(t *testing.T) {
	t.Parallel()

	partie := eineAusloeserpartie(t)
	partie.Zug.Aktiv = "sitz-2"

	// **Von Hand gestellt und nicht erspielt**: zwei eingebaute Fähigkeiten
	// **verschiedener** Sitze zugleich verlangten zwei Monarchen, und davon
	// gibt es nur einen (CR 725.3). Was hier gemessen wird, ist die **Regel**
	// aus CR 603.3b, und die gilt, sobald `M8` beliebig viele bringt.
	partie.Ausgeloest = []domain.Ausloeser{
		{
			Art: domain.AusloeserMonarchzug, Beherrscher: "sitz-1",
			Kennung: "sitz-1/monarch.zug/0",
		},
		{
			Art: domain.AusloeserMonarchzug, Beherrscher: "sitz-2",
			Kennung: "sitz-2/monarch.zug/0",
		},
	}

	danach := partie.Weiter()

	oben := danach.StapelOben()
	if oben == nil {
		t.Fatal("auf dem Stapel liegt nichts")
	}

	// **Oben liegt die des Nichtaktiven** — und das ist die Regel und
	// nicht ihr Gegenteil.
	//
	// CR 603.3b lässt den aktiven Spieler **zuerst** legen; wer zuerst
	// legt, liegt unten (CR 405.2), und der Stapel wird von oben
	// abgeräumt. Die Fähigkeit des Nichtaktiven wird also **zuerst**
	// verrechnet — so herum steht es am Tisch.
	if oben.Beherrscher != "sitz-1" {
		t.Errorf("oben liegt die Fähigkeit von %q", oben.Beherrscher)
	}

	// **Und beide liegen da** — eine Warteschlange, die eine verschluckt,
	// wäre schlimmer als eine falsche Reihenfolge.
	auf := 0

	for _, objekt := range danach.Objekte {
		if objekt.Zone == domain.ZoneStapel {
			auf++
		}
	}

	if auf != 2 {
		t.Errorf("auf dem Stapel liegen %d Fähigkeiten", auf)
	}

	// **Und die Warteschlange ist leer** — CR 603.3: sie sind gelegt.
	if len(danach.Ausgeloest) != 0 {
		t.Errorf("%d warten noch", len(danach.Ausgeloest))
	}

	// **Und niemand ordnet** — bei einer je Sitz gibt es nichts zu wählen.
	if danach.Ausloeserreihung != "" {
		t.Errorf("geordnet wird von %q", danach.Ausloeserreihung)
	}
}

// eineAusloeserpartie gibt eine Partie mit einem Bären und Karten in der
// Bibliothek.
func eineAusloeserpartie(t *testing.T) domain.Partie {
	t.Helper()

	deck := domain.Partiedeck{
		Abdruck: "abcd",
		Karten: []domain.Abzugskarte{{
			Kennung: "baer", Name: "Bär", Menge: 8,
			Seiten: []domain.Abzugsseite{{Felder: map[string]string{
				"name": "Bär", "super type": "Creature",
				"power": "2", "toughness": "2",
			}}},
		}},
	}

	partie := domain.Partie{
		Spiel:  "magic",
		Profil: domain.Grundprofil("magic"),
		Regelwerk: domain.Regelwerk{
			Kennung: "sandkasten", Fassung: domain.Grundprofilfassung,
			Strenge: domain.StrengeSandkasten,
		},
		Wirt: "sitz-1",
		Sitze: []domain.Sitz{
			{Kennung: "sitz-1", Name: "Anna", Deck: deck},
			{Kennung: "sitz-2", Name: "Bert", Deck: deck},
		},
		Objekte: []domain.Spielobjekt{{
			Kennung: "sitz-1/baer/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneTisch, Karte: "baer",
		}},
	}.MitStartleben()

	for _, wer := range []string{"sitz-1", "sitz-2"} {
		for lauf := range 4 {
			partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
				Kennung:     wer + "/bib/" + string(rune('a'+lauf)),
				Art:         domain.ObjektKarte,
				Beherrscher: wer, Eigner: wer,
				Zone: domain.ZoneBibliothek, Karte: "baer",
				Ordnung: lauf,
			})
		}
	}

	return partie.ZugBeginnen("sitz-1")
}

// handVon gibt die Handkarten eines Sitzes.
//
//nolint:unparam // der Sitz gehört zur Frage: eine Hand ohne Besitzer sagt nichts
func handVon(partie domain.Partie, wer string) []domain.Spielobjekt {
	var hand []domain.Spielobjekt

	for _, objekt := range partie.Objekte {
		if objekt.Zone == domain.ZoneHand && objekt.Beherrscher == wer {
			hand = append(hand, objekt)
		}
	}

	return hand
}

// bisZumEndschritt passt so lange, bis der Endschritt erreicht ist.
//
// **Über den Zug und nicht über einen Griff in den Zustand**: der Endschritt
// ist der elfte von zwölf (CR 500.1), und was unterwegs geschieht, gehört
// dazu.
func bisZumEndschritt(t *testing.T, partie domain.Partie) domain.Partie {
	t.Helper()

	for range 40 {
		if partie.Zug.Abschnitt.Schritt == domain.SchrittEnde {
			return partie
		}

		if partie.Zug.Prioritaet == "" {
			partie = partie.Weiter()

			continue
		}

		danach, fehler := partie.Passen(partie.Zug.Prioritaet)
		if fehler != nil {
			t.Fatalf("passen: %v", fehler)
		}

		partie = danach
	}

	t.Fatal("der Endschritt kam nicht")

	return partie
}
