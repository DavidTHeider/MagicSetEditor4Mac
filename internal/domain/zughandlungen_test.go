package domain_test

// **Was von selbst geschieht** — Werkstück G9 von Phase 5.5
// (#1592).
//
// # Wogegen hier geprüft wird
//
//   - CR 703.3 — zugbasierte Handlungen laufen **beim Beginn** eines
//     Abschnitts, vor jeder Priorität.
//   - CR 502.3 — der aktive Spieler enttappt seine Permanenten.
//   - CR 504.1 — er zieht eine Karte, CR 121.1 sagt woher.
//   - CR 103.8a — **wer anfängt, überspringt sein erstes Ziehen**, und zwar
//     im Spiel zu zweit.
//   - CR 514.1 — er wirft auf die Handgrenze ab.
//   - CR 514.2 — aller markierte Schaden wird entfernt.
//
// **Und dass ein Abschnitt ohne Priorität nicht wartet** (CR 502.4,
// CR 514.3): er tut, was er zu tun hat, und gibt weiter.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// eineZugpartie gibt eine Partie mit Karten in Bibliothek und Hand.
func eineZugpartie() domain.Partie {
	partie := einePartie()
	partie.Objekte = nil

	for _, sitz := range []string{"sitz-1", "sitz-2"} {
		for stueck := range 3 {
			partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
				Kennung:     sitz + "/k1/" + string(rune('a'+stueck)),
				Art:         domain.ObjektKarte,
				Beherrscher: sitz, Eigner: sitz,
				Zone: domain.ZoneBibliothek, Karte: "k1",
				Ordnung: stueck + 1,
			})
		}
	}

	return partie
}

// **Der aktive Spieler enttappt, der andere nicht** — CR 502.3.
func TestNurDerAktiveSpielerEnttappt(t *testing.T) {
	t.Parallel()

	partie := eineZugpartie()
	partie.Objekte = append(partie.Objekte,
		domain.Spielobjekt{
			Kennung: "sitz-1/w/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneTisch, Getappt: true,
		},
		domain.Spielobjekt{
			Kennung: "sitz-2/w/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-2", Eigner: "sitz-2",
			Zone: domain.ZoneTisch, Getappt: true,
		})

	// **`ZugBeginnen` läuft durch das Enttappen hindurch** — dort bekommt
	// niemand Priorität (CR 502.4), also wartet der Abschnitt nicht.
	gelaufen := partie.ZugBeginnen("sitz-1")

	getappt := map[string]bool{}
	for _, objekt := range gelaufen.Objekte {
		getappt[objekt.Kennung] = objekt.Getappt
	}

	switch {
	case getappt["sitz-1/w/0"]:
		t.Error("der aktive Spieler hat nicht enttappt")
	case !getappt["sitz-2/w/0"]:
		t.Error("der andere Spieler hat mit enttappt")
	}
}

// **Ein Abschnitt ohne Priorität wartet nicht** — nach dem Enttappen steht
// die Partie in der Versorgung, und dort hat der aktive Spieler sie.
func TestDasEnttappenHaeltDenZugNichtAn(t *testing.T) {
	t.Parallel()

	partie := eineZugpartie().ZugBeginnen("sitz-1")

	switch {
	case partie.Zug.Abschnitt.Schritt != domain.SchrittVersorgung:
		t.Errorf("es steht bei %+v", partie.Zug.Abschnitt)
	case partie.Zug.Prioritaet != "sitz-1":
		t.Errorf("Priorität hat %q", partie.Zug.Prioritaet)
	}
}

// **Wer anfängt, zieht in seinem ersten Zug nicht** — CR 103.8a.
//
// **Nachgewiesen an der Kartenzahl und nicht an einer Behauptung**, wie die
// Abnahme es verlangt.
func TestWerAnfaengtZiehtImErstenZugNicht(t *testing.T) {
	t.Parallel()

	partie := eineZugpartie().ZugBeginnen("sitz-1")
	partie.Zug.Starter = "sitz-1"

	starterzug := bisZumZiehen(t, partie)

	if hand := aufDerHand(starterzug, "sitz-1"); hand != 0 {
		t.Errorf("der Starter hat %d Karten gezogen", hand)
	}

	// **Der andere zieht in seinem ersten Zug schon** — CR 103.8a nimmt nur
	// den aus, der anfängt.
	//
	// **Der Starter wird vor dem Ziehschritt gesetzt** und nicht danach: die
	// Regel wird beim Betreten des Abschnitts gefragt, und wer sie erst
	// hinterher stellte, prüfte nichts.
	anderer := eineZugpartie().ZugBeginnen("sitz-2")
	anderer.Zug.Starter = "sitz-1"

	zweiter := bisZumZiehen(t, anderer)

	if hand := aufDerHand(zweiter, "sitz-2"); hand != 1 {
		t.Errorf("der zweite Spieler hat %d Karten auf der Hand", hand)
	}
}

// **Und im zweiten Zug zieht auch der Starter** — CR 103.8a gilt für den
// ersten.
func TestImZweitenZugZiehtAuchDerStarter(t *testing.T) {
	t.Parallel()

	partie := eineZugpartie().ZugBeginnen("sitz-1")
	partie.Zug.Starter = "sitz-1"
	partie.Zug.Nummer = 3

	gezogen := bisZumZiehen(t, partie)

	if hand := aufDerHand(gezogen, "sitz-1"); hand != 1 {
		t.Errorf("im dritten Zug hat der Starter %d Karten", hand)
	}
}

// **Gezogen wird die oberste Karte** — CR 121.1, und oben ist die kleinste
// Ordnung (`platzFuer(oben) → 1`).
func TestGezogenWirdDieObersteKarte(t *testing.T) {
	t.Parallel()

	partie := eineZugpartie().ZugBeginnen("sitz-2")
	partie.Zug.Starter = "sitz-1"

	gezogen := bisZumZiehen(t, partie)

	for _, objekt := range gezogen.Objekte {
		if objekt.Kennung == "sitz-2/k1/a" && objekt.Zone != domain.ZoneHand {
			t.Errorf("die oberste Karte liegt in %q", objekt.Zone)
		}
	}
}

// **Eine leere Bibliothek hält die Partie nicht an.**
//
// **Weil das Verlieren eine zustandsbasierte Aktion ist** (CR 704.5b) und
// die gibt es hier nicht. Wer aus einer leeren Bibliothek zöge, verlöre —
// bis dahin geschieht nichts, und der Zug läuft weiter.
func TestEineLeereBibliothekHaeltNichtAn(t *testing.T) {
	t.Parallel()

	partie := eineZugpartie()
	partie.Objekte = nil
	partie = partie.ZugBeginnen("sitz-2")
	partie.Zug.Starter = "sitz-1"

	gezogen := bisZumZiehen(t, partie)

	if gezogen.Zug.Abschnitt.Schritt != domain.SchrittZiehen {
		t.Errorf("es steht bei %+v", gezogen.Zug.Abschnitt)
	}
}

// bisZumZiehen passt so lange durch, bis der Ziehschritt dasteht.
//
// **Bis zum Ziehen und nicht bis zu einem beliebigen Schritt**: die Läufe
// hier fragen alle danach, und ein Parameter, den nur ein Wert je füllt, ist
// eine Verallgemeinerung ohne zweiten Fall.
func bisZumZiehen(t *testing.T, partie domain.Partie) domain.Partie {
	t.Helper()

	for range 40 {
		if partie.Zug.Abschnitt.Schritt == domain.SchrittZiehen {
			return partie
		}

		gepasst, fehler := partie.Passen(partie.Zug.Prioritaet)
		if fehler != nil {
			t.Fatalf("Passen bei %+v: %v", partie.Zug.Abschnitt, fehler)
		}

		partie = gepasst
	}

	t.Fatal("der Ziehschritt kam nicht")

	return partie
}

// aufDerHand zählt, was dieser Sitz auf der Hand hält.
func aufDerHand(partie domain.Partie, sitz string) int {
	anzahl := 0

	for _, objekt := range partie.Objekte {
		if objekt.Beherrscher == sitz && objekt.Zone == domain.ZoneHand {
			anzahl++
		}
	}

	return anzahl
}

// **Im Aufräumschritt geht aller markierte Schaden weg** — CR 514.2.
//
// **Und zwar von allem**, nicht nur vom aktiven Spieler: die Regel sagt „all
// damage marked on permanents".
func TestDerAufraeumschrittNimmtDenSchadenWeg(t *testing.T) {
	t.Parallel()

	partie := eineZugpartie()
	partie.Objekte = append(partie.Objekte,
		domain.Spielobjekt{
			Kennung: "sitz-1/w/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-1", Eigner: "sitz-1",
			Zone: domain.ZoneTisch, Schaden: 3,
		},
		domain.Spielobjekt{
			Kennung: "sitz-2/w/0", Art: domain.ObjektKarte,
			Beherrscher: "sitz-2", Eigner: "sitz-2",
			Zone: domain.ZoneTisch, Schaden: 2,
		})

	partie = partie.ZugBeginnen("sitz-1")
	partie.Zug.Abschnitt = domain.Abschnitt{
		Phase: domain.PhaseEnde, Schritt: domain.SchrittEnde,
	}
	partie.Zug.Prioritaet = "sitz-1"

	// **Zweimal passen beendet den Endschritt** (CR 117.4) — danach läuft das
	// Aufräumen und gleich der nächste Zug an.
	for _, wer := range []string{"sitz-1", "sitz-2"} {
		gepasst, fehler := partie.Passen(wer)
		if fehler != nil {
			t.Fatalf("Passen (%s): %v", wer, fehler)
		}

		partie = gepasst
	}

	for _, objekt := range partie.Objekte {
		if objekt.Schaden != 0 {
			t.Errorf("auf %q sind noch %d Schaden", objekt.Kennung, objekt.Schaden)
		}
	}

	// **Und der Zug ist weitergegangen** — im Aufräumen hat niemand
	// Priorität (CR 514.3), also wartet es nicht.
	if partie.Zug.Nummer != 2 {
		t.Errorf("es ist Zug %d", partie.Zug.Nummer)
	}
}

// **Wer über der Handgrenze steht, wirft ab — und der Zug wartet darauf.**
//
// CR 514.1: der aktive Spieler wirft auf die Handgrenze ab. **Welche** Karten,
// entscheidet er — deshalb steht eine Pflicht da und keine Handlung.
func TestUeberDerHandgrenzeWirdAbgeworfen(t *testing.T) {
	t.Parallel()

	partie := eineZugpartie()

	for stueck := range 9 {
		partie.Objekte = append(partie.Objekte, domain.Spielobjekt{
			Kennung: "sitz-1/h/" + string(rune('a'+stueck)),
			Art:     domain.ObjektKarte, Karte: "h",
			Beherrscher: "sitz-1", Eigner: "sitz-1", Zone: domain.ZoneHand,
		})
	}

	partie = partie.ZugBeginnen("sitz-1")
	partie.Zug.Abschnitt = domain.Abschnitt{
		Phase: domain.PhaseEnde, Schritt: domain.SchrittEnde,
	}
	partie.Zug.Prioritaet = "sitz-1"

	for _, wer := range []string{"sitz-1", "sitz-2"} {
		gepasst, fehler := partie.Passen(wer)
		if fehler != nil {
			t.Fatalf("Passen (%s): %v", wer, fehler)
		}

		partie = gepasst
	}

	switch {
	case partie.Zug.Abschnitt.Schritt != domain.SchrittAufraeumen:
		t.Fatalf("es steht bei %+v", partie.Zug.Abschnitt)
	case partie.Zug.Abwurf != 2:
		t.Fatalf("abzuwerfen sind %d", partie.Zug.Abwurf)
	case partie.Zug.Nummer != 1:
		t.Fatalf("der Zug ist schon weiter: %d", partie.Zug.Nummer)
	}

	// **Wer nicht am Zug ist, wirft nicht ab.**
	if _, fehler := partie.Abwerfen("sitz-2", "sitz-1/h/a"); fehler == nil {
		t.Error("ein fremder Abwurf kam durch")
	}

	// **Und keine Karte, die nicht auf der Hand liegt.**
	if _, fehler := partie.Abwerfen("sitz-1", "sitz-1/k1/a"); fehler == nil {
		t.Error("eine Karte aus der Bibliothek wurde abgeworfen")
	}

	erste, fehler := partie.Abwerfen("sitz-1", "sitz-1/h/a")
	if fehler != nil {
		t.Fatalf("Abwerfen: %v", fehler)
	}

	if erste.Zug.Abwurf != 1 || erste.Zug.Nummer != 1 {
		t.Fatalf("nach einem Abwurf: offen %d, Zug %d",
			erste.Zug.Abwurf, erste.Zug.Nummer)
	}

	zweite, fehler := erste.Abwerfen("sitz-1", "sitz-1/h/b")
	if fehler != nil {
		t.Fatalf("Abwerfen: %v", fehler)
	}

	switch {
	case zweite.Zug.Abwurf != 0:
		t.Errorf("es sind noch %d offen", zweite.Zug.Abwurf)
	case zweite.Zug.Nummer != 2:
		t.Errorf("der Zug ging nicht weiter: %d", zweite.Zug.Nummer)
	case aufDerHand(zweite, "sitz-1") != domain.Handgrenze:
		t.Errorf("auf der Hand liegen %d", aufDerHand(zweite, "sitz-1"))
	}

	// **Die abgeworfenen Karten liegen im Friedhof** — CR 514.1 wirft ab,
	// und Abwerfen legt in den Friedhof (CR 701.8a).
	imFriedhof := 0

	for _, objekt := range zweite.Objekte {
		if objekt.Zone == domain.ZoneFriedhof {
			imFriedhof++
		}
	}

	if imFriedhof != 2 {
		t.Errorf("im Friedhof liegen %d Karten", imFriedhof)
	}
}

// **Ohne Pflicht wirft niemand ab.**
func TestOhnePflichtWirdNichtAbgeworfen(t *testing.T) {
	t.Parallel()

	partie := eineZugpartie().ZugBeginnen("sitz-1")

	if _, fehler := partie.Abwerfen("sitz-1", "sitz-1/k1/a"); fehler == nil {
		t.Error("ein Abwurf ohne Pflicht kam durch")
	}
}
