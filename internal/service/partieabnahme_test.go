package service_test

// **Eine ganze Partie, und niemand setzt eine Regel von Hand durch** — die
// zweite Phasenabnahme der Regelmaschine
// (#1593).
//
// # Warum es diesen Lauf gibt
//
// **Weil neunzehn Werkstücke neunzehn Läufe haben und keiner von ihnen die
// Frage der Phase beantwortet.** Jeder prüft sein Stück: der Stapel
// verrechnet, der Kampf teilt Schaden zu, die zustandsbasierte Prüfung räumt
// ab. **Ob daraus ein Spiel wird**, prüft keiner — und das ist die erste
// Dimension aus §10: löst dieser Beitrag ein, wofür es die Anwendung gibt,
// und **woran ist das zu sehen**?
//
// Hier ist es zu sehen: zwei Sitze spielen von der Starthand — **gestellt und
// zwei Karten groß**, der Grund steht unten — bis zum Spielende, und **jeder
// einzelne Befehl ist einer, den ein Mensch am Fenster stellen kann**. Das wird nicht behauptet, sondern gezählt —
// [abnahmelauf.stelle] hält jede Art fest, und am Ende steht die Liste gegen
// die der [handgriffe].
//
// # Was er hält
//
//   - **Was von selbst geschehen muss, geschieht von selbst**: enttappen,
//     ziehen, die zwölf Abschnitte durchlaufen, Kampfschaden zuteilen, eine
//     tote Kreatur ablegen, eine Partie beenden. Kein Befehl dieses Laufs tut
//     eines davon.
//   - **Jeder Befehl war angeboten, bevor er gestellt wurde** —
//     [service.Partieleitung.VerfuegbareAktionen] hat ihn genannt. Das ist
//     dieselbe Liste, aus der sich die Handlungsleiste des Fensters ergibt
//     (#1623).
//     **Dass es zu jeder Art auch eine Form gibt**, hält
//     `internal/werkzeuge/befehlsformen_test.go` — damit ist die Kette
//     geschlossen: der Kern erlaubt, die Form beschreibt, das Fenster
//     stellt.
//
// # Was er nicht hält
//
//   - **Das Netz.** Er geht über [service.Partieleitung] — dieselbe Naht, die
//     jede der vier Flächen ruft —, nicht über eine Leitung zwischen zwei
//     Rechnern. Dass ein Wunsch über die Leitung geht und als Tatsache
//     zurückkommt, halten die Läufe von `G5`
//     (#1462).
//   - **Den Zufall.** Die Starthand ist **gestellt** und nicht gemischt: ein
//     Lauf, dessen Ausgang an einer Mischung hängt, ist an manchen Tagen rot.
//     Das Mischen hält `G1`.
//   - **Jede Regel.** Er fährt **einen** Weg durch das Regelwerk. Was auf ihm
//     nicht vorkommt, steht deshalb nicht still — es steht in den Läufen der
//     Werkstücke.

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// inHauptphase hält, sobald dieser Sitz in seiner ersten Hauptphase am Zug
// ist und Priorität hat.
func inHauptphase(wer string, zug int) func(domain.Partie) bool {
	return func(partie domain.Partie) bool {
		return partie.Zug.Nummer == zug && partie.Zug.Aktiv == wer &&
			partie.Zug.Abschnitt.Phase == domain.PhaseHauptVorKampf &&
			partie.Zug.Prioritaet == wer
	}
}

// imSchritt hält, sobald dieser Zug in diesem Schritt steht.
func imSchritt(zug int, schritt domain.Schritt) func(domain.Partie) bool {
	return func(partie domain.Partie) bool {
		return partie.Zug.Nummer == zug &&
			partie.Zug.Abschnitt.Schritt == schritt
	}
}

// inZone hält, sobald dieses Objekt in dieser Zone liegt.
func inZone(kennung string, zone domain.Zone) func(domain.Partie) bool {
	return func(partie domain.Partie) bool {
		for _, objekt := range partie.Objekte {
			if objekt.Kennung == kennung {
				return objekt.Zone == zone
			}
		}

		return false
	}
}

// **Von der Starthand bis zum Spielende, ohne einen einzigen Handgriff.**
//
// # Der Weg, den dieser Lauf fährt
//
//  1. **Zug 1** — Anna legt ein Gebirge, tappt es und wirkt den Koloss. Beide
//     passen, der Zauber verrechnet, die Kreatur steht.
//  2. **Zug 2** — Bert tut dasselbe mit seinem Bären.
//  3. **Zug 3** — Anna greift an, Bert blockt. Der Bär stirbt **von selbst**
//     (CR 704.5g), Bert verliert kein Leben, der Koloss steht weiter.
//  4. **Zug 5** — Anna greift noch einmal an. Niemand blockt, zwanzig Schaden
//     kommen durch, und die Partie endet **von selbst** (CR 704.5a).
func TestEineGanzePartieLaeuftOhneEinenHandgriff(t *testing.T) {
	t.Parallel()

	lauf := einAbnahmelauf(t)

	// **Zug 1** — Anna legt, tappt, wirkt.
	lauf.passeBis("Annas erste Hauptphase", inHauptphase("sitz-1", 1))
	lauf.stelle("sitz-1", service.BefehlLandspielen,
		map[string]any{"objekt": "sitz-1/gebirge/0"})
	lauf.stelle("sitz-1", service.BefehlLandmana,
		map[string]any{"objekt": "sitz-1/gebirge/0"})
	lauf.stelle("sitz-1", service.BefehlWirken,
		map[string]any{"objekt": "sitz-1/wesen/0"})

	lauf.passeBis("den Koloss auf dem Tisch",
		inZone("sitz-1/wesen/0", domain.ZoneTisch))

	// **Zug 2** — Bert legt, tappt, wirkt.
	lauf.passeBis("Berts erste Hauptphase", inHauptphase("sitz-2", 2))
	lauf.stelle("sitz-2", service.BefehlLandspielen,
		map[string]any{"objekt": "sitz-2/gebirge/0"})
	lauf.stelle("sitz-2", service.BefehlLandmana,
		map[string]any{"objekt": "sitz-2/gebirge/0"})
	lauf.stelle("sitz-2", service.BefehlWirken,
		map[string]any{"objekt": "sitz-2/wesen/0"})

	lauf.passeBis("den Bären auf dem Tisch",
		inZone("sitz-2/wesen/0", domain.ZoneTisch))

	// **Zug 3** — Anna greift an, Bert blockt.
	lauf.passeBis("Annas Angriffsschritt",
		imSchritt(3, domain.SchrittAngreifer))
	lauf.stelle("sitz-1", service.BefehlAngreifer, map[string]any{
		"angriffe": []any{map[string]any{
			"kreatur": "sitz-1/wesen/0", "ziel": "sitz-2",
		}},
	})

	lauf.passeBis("Berts Blockschritt", imSchritt(3, domain.SchrittBlocker))
	lauf.stelle("sitz-2", service.BefehlBlocker, map[string]any{
		"bloecke": []any{map[string]any{
			"kreatur": "sitz-2/wesen/0", "angreifer": "sitz-1/wesen/0",
		}},
	})

	lauf.passeBis("den toten Bären",
		inZone("sitz-2/wesen/0", domain.ZoneFriedhof))

	// **Der Block hat gehalten** — Bert steht noch bei seinem Startleben,
	// und der Koloss hat zwei Schaden weggesteckt.
	nachDemBlock := lauf.partie()
	if leben := lebenVon(t, nachDemBlock, "sitz-2"); leben != 20 {
		t.Errorf("Bert steht bei %d Leben — der Block hat nicht gehalten",
			leben)
	}

	// **Zug 5** — niemand blockt, und die Partie endet von selbst.
	lauf.passeBis("Annas zweiten Angriffsschritt",
		imSchritt(5, domain.SchrittAngreifer))

	// **Und der Koloss steht enttappt da** — CR 502.3. Er hat in Zug 3
	// angegriffen und dabei getappt (CR 508.1f); enttappt hat ihn niemand.
	vorDemAngriff := lauf.partie()

	if objektVon(t, vorDemAngriff, "sitz-1/wesen/0").Getappt {
		t.Error("der Koloss ist noch getappt — niemand hat ihn enttappt")
	}

	// **Beide haben zweimal gezogen, ohne dass jemand ziehen ließ** —
	// CR 504.1. Bert in Zug 2 und 4, Anna in Zug 3 und 5; im **ersten** Zug
	// zieht der Anfangende nicht (CR 103.8a), und deshalb stehen beide
	// Bibliotheken gleich hoch.
	//
	// **Gezählt vor dem letzten Angriff und nicht danach**: wer die Partie
	// verliert, nimmt seine Karten mit (CR 800.4a, `M17`) — hinterher wäre
	// Berts Bibliothek leer, und die Zahl sagte etwas anderes aus.
	for _, wer := range []string{"sitz-1", "sitz-2"} {
		if uebrig := inZonezahl(
			vorDemAngriff, wer, domain.ZoneBibliothek); uebrig != 18 {
			t.Errorf("die Bibliothek von %q hat %d Karten statt 18 — "+
				"gezogen wird von selbst", wer, uebrig)
		}
	}
	lauf.stelle("sitz-1", service.BefehlAngreifer, map[string]any{
		"angriffe": []any{map[string]any{
			"kreatur": "sitz-1/wesen/0", "ziel": "sitz-2",
		}},
	})

	lauf.passeBis("das Spielende", func(partie domain.Partie) bool {
		return partie.Stand == domain.StandBeendet
	})

	ende := lauf.partie()

	if ende.Ausgang.Sieger != "sitz-1" {
		t.Errorf("der Ausgang lautet %+v", ende.Ausgang)
	}

	if leben := lebenVon(t, ende, "sitz-2"); leben > 0 {
		t.Errorf("Bert steht bei %d Leben und die Partie ist beendet", leben)
	}

	// **Und Anna hat keinen Schaden genommen** — der Bär hat geblockt und
	// nie angegriffen. Eine Partie, in der beide Leben verlieren, wäre eine
	// andere als die, die hier beschrieben steht.
	if leben := lebenVon(t, ende, "sitz-1"); leben != 20 {
		t.Errorf("Anna steht bei %d Leben statt bei 20", leben)
	}

	// **Fünf Züge sind von selbst gelaufen** — zwölf Abschnitte je Zug, und
	// kein Befehl hat einen davon weitergeschaltet außer dem Passen.
	if ende.Zug.Nummer != 5 {
		t.Errorf("die Partie steht bei Zug %d", ende.Zug.Nummer)
	}

	// **Und kein einziger Handgriff.** Das ist die Aussage des Laufs.
	for _, art := range lauf.benutzt {
		if slices.Contains(handgriffe, art) {
			t.Errorf("%q ist ein Handgriff — gestellt wurde %v",
				art, lauf.benutzt)
		}
	}

	// **Und die Gegenprobe zur Aufzählung**: gestellt wurde genau das, was
	// ein Mensch am Fenster stellt. Stünde hier nur die Verneinung darüber,
	// wäre der Lauf auch dann grün, wenn er gar nichts gestellt hätte (§10).
	gewollt := []string{
		service.BefehlPassen, service.BefehlLandspielen,
		service.BefehlLandmana, service.BefehlWirken,
		service.BefehlAngreifer, service.BefehlBlocker,
	}

	slices.Sort(gewollt)

	gestellt := slices.Clone(lauf.benutzt)
	slices.Sort(gestellt)

	if !slices.Equal(gestellt, gewollt) {
		t.Errorf("gestellt wurde %v, gewollt war %v", gestellt, gewollt)
	}
}

// objektVon gibt ein Objekt der Partie.
func objektVon(
	t *testing.T, partie domain.Partie, kennung string,
) domain.Spielobjekt {
	t.Helper()

	for _, objekt := range partie.Objekte {
		if objekt.Kennung == kennung {
			return objekt
		}
	}

	t.Fatalf("das Objekt %q gibt es nicht", kennung)

	return domain.Spielobjekt{}
}

// inZonezahl zählt die Objekte eines Sitzes in einer Zone.
func inZonezahl(partie domain.Partie, wer string, zone domain.Zone) int {
	anzahl := 0

	for _, objekt := range partie.Objekte {
		if objekt.Eigner == wer && objekt.Zone == zone {
			anzahl++
		}
	}

	return anzahl
}
