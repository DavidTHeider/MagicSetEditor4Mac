package domain_test

// **Erstschlag und Doppelschlag** — Werkstück M6 der Regelmaschine
// (#1608).
//
// # Eine eigene Datei, und der Schnitt folgt der Frage
//
// **Nebenan steht, wie ein Kampf abläuft**; hier steht, **wann er zwei
// Schadensschritte hat** (CR 510.4) — und `kampf_test.go` stand mit beidem
// bei 498 Zeilen (§2).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Erstschlag tötet den Blocker, bevor der zurückschlägt** — CR 510.4.
//
// **Das ist die Abnahme des Issues**, wörtlich: „nachgewiesen an zwei
// Schadensschritten."
//
// # Warum zwei von 194 Schlüsselwörtern hier stehen
//
// **Weil CR 510.4 sie namentlich nennt.** Erstschlag (CR 702.7a) und
// Doppelschlag (CR 702.4a) sind **statische** Fähigkeiten, deren ganze
// Bedeutung in CR 510 steht — sie tun nichts, was ein Text ausführen müsste;
// sie **teilen den Schadensschritt**. Sie zu erkennen heißt, ein Wort zu
// lesen, so wie `M2` die Typzeile liest — nicht, einen Text auszuführen. Die
// übrigen 192 bleiben `M14`
// (#1616).
func TestErstschlagToetetVorDemZurueckschlagen(t *testing.T) {
	t.Parallel()

	// **Der erste Schadensschritt ist gelaufen** — nur der Erstschläger
	// teilt aus (CR 510.4).
	erster := mitErstschlag(t)

	switch {
	// **Der Riese ist unversehrt** — der Bär teilt im ersten Schritt nicht
	// aus (CR 510.4).
	case objektMit(t, erster, "sitz-1/riese/0").Schaden != 0:
		t.Errorf("der Riese trägt schon %d Schaden",
			objektMit(t, erster, "sitz-1/riese/0").Schaden)

	// **Und der Bär liegt bereits im Friedhof.** Die Zustandsprüfung läuft
	// vor der Priorität (CR 704.3), und drei Schaden auf zwei
	// Widerstandskraft sind tödlich (CR 704.5g) — er kommt gar nicht mehr
	// dazu, zurückzuschlagen.
	case objektMit(t, erster, "sitz-2/baer/0").Zone != domain.ZoneFriedhof:
		t.Errorf("der Bär liegt in %q",
			objektMit(t, erster, "sitz-2/baer/0").Zone)

	case erster.Zug.Kampf.Schadensschritte != 1:
		t.Errorf("gezählt sind %d Schadensschritte",
			erster.Zug.Kampf.Schadensschritte)
	}

	// **Der zweite Schadensschritt kommt über den Zug** — beide passen, und
	// CR 510.4 setzt denselben Schritt noch einmal ein.
	zweiter, fehler := erster.Passen(erster.Zug.Prioritaet)
	if fehler != nil {
		t.Fatalf("passen: %v", fehler)
	}

	zweiter, fehler = zweiter.Passen(zweiter.Zug.Prioritaet)
	if fehler != nil {
		t.Fatalf("zweites passen: %v", fehler)
	}

	switch {
	case zweiter.Zug.Abschnitt.Schritt != domain.SchrittKampfschaden:
		t.Errorf("der zweite Schadensschritt kam nicht: %+v",
			zweiter.Zug.Abschnitt)

	// **Und der Riese bleibt unversehrt** — der Bär teilt nichts mehr aus.
	case objektMit(t, zweiter, "sitz-1/riese/0").Schaden != 0:
		t.Errorf("der Riese trägt %d Schaden — der Bär hat zurückgeschlagen",
			objektMit(t, zweiter, "sitz-1/riese/0").Schaden)
	}
}

// **Ohne Erstschlag gibt es nur einen Schritt.**
//
// **Die Gegenprobe** (§10): ohne sie sagte der Lauf darüber nur, dass
// **irgendwann** zweimal ausgeteilt wird — ein Weg, der immer zwei Schritte
// führte, käme genauso durch, und jede Kreatur schlüge doppelt zu.
func TestOhneErstschlagGibtEsNurEinenSchritt(t *testing.T) {
	t.Parallel()

	danach := mitBlock(t)

	// **Und der Schritt kommt kein zweites Mal** — der Kampf geht weiter
	// ins Kampfende (CR 506.1). Gemessen am **Zug** und nicht an einem
	// inneren Feld: ein zweiter Schadensschritt wäre von außen sichtbar.
	nachher, fehler := danach.Passen(danach.Zug.Prioritaet)
	if fehler != nil {
		t.Fatalf("passen: %v", fehler)
	}

	nachher, fehler = nachher.Passen(nachher.Zug.Prioritaet)
	if fehler != nil {
		t.Fatalf("zweites passen: %v", fehler)
	}

	switch {
	case danach.Zug.Kampf.Schadensschritte != 1:
		t.Errorf("gezählt sind %d Schadensschritte",
			danach.Zug.Kampf.Schadensschritte)
	case nachher.Zug.Abschnitt.Schritt == domain.SchrittKampfschaden:
		t.Error("ohne Erstschlag kam ein zweiter Schadensschritt")
	}
}

// mitErstschlag gibt eine Partie, in der der Riese Erstschlag hat.
func mitErstschlag(t *testing.T) domain.Partie {
	t.Helper()

	partie := einKampftisch(t)

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)

	karten := make([]domain.Abzugskarte, len(sitze[0].Deck.Karten))
	copy(karten, sitze[0].Deck.Karten)
	karten[0].Seiten = []domain.Abzugsseite{{Felder: map[string]string{
		"super type": "Creature", "sub type": "Giant",
		"power": "3", "toughness": "3",
		"rule text": "First strike",
	}}}
	sitze[0].Deck.Karten = karten
	partie.Sitze = sitze

	partie, fehler := partie.AngreiferErklaeren("sitz-1",
		[]domain.Angriff{{Kreatur: "sitz-1/riese/0", Ziel: "sitz-2"}})
	if fehler != nil {
		t.Fatalf("angreifer erklären: %v", fehler)
	}

	partie, fehler = bisZumSchritt(t, partie, domain.SchrittBlocker).
		BlockerErklaeren("sitz-2",
			[]domain.Block{{
				Kreatur: "sitz-2/baer/0", Angreifer: "sitz-1/riese/0",
			}})
	if fehler != nil {
		t.Fatalf("blocker erklären: %v", fehler)
	}

	return bisZumSchritt(t, partie, domain.SchrittKampfschaden)
}

// **Auch ein Blocker mit Erstschlag teilt den Schritt** — CR 510.4.
//
// **Der Zweig, den dieser Lauf betritt**: die Blockerschleife in
// `erstschlagImKampf`. Der Lauf darüber gibt dem **Angreifer** Erstschlag —
// ein Weg, der nur Angreifer fragte, käme damit durch, und ein blockender
// Erstschläger stürbe, bevor er zuschlägt (§10).
func TestAuchEinBlockerMitErstschlagTeiltDenSchritt(t *testing.T) {
	t.Parallel()

	partie := einKampftisch(t)

	sitze := make([]domain.Sitz, len(partie.Sitze))
	copy(sitze, partie.Sitze)

	karten := make([]domain.Abzugskarte, len(sitze[1].Deck.Karten))
	copy(karten, sitze[1].Deck.Karten)
	karten[1].Seiten = []domain.Abzugsseite{{Felder: map[string]string{
		"super type": "Creature", "sub type": "Bear",
		"power": "2", "toughness": "2", "rule text": "First strike",
	}}}
	sitze[1].Deck.Karten = karten
	partie.Sitze = sitze

	partie, fehler := partie.AngreiferErklaeren("sitz-1",
		[]domain.Angriff{{Kreatur: "sitz-1/riese/0", Ziel: "sitz-2"}})
	if fehler != nil {
		t.Fatalf("angreifer erklären: %v", fehler)
	}

	partie, fehler = bisZumSchritt(t, partie, domain.SchrittBlocker).
		BlockerErklaeren("sitz-2", []domain.Block{{
			Kreatur: "sitz-2/baer/0", Angreifer: "sitz-1/riese/0",
		}})
	if fehler != nil {
		t.Fatalf("blocker erklären: %v", fehler)
	}

	erster := bisZumSchritt(t, partie, domain.SchrittKampfschaden)

	switch {
	// **Der Riese trägt zwei Schaden und lebt** — der Bär hat zuerst
	// zugeschlagen.
	case objektMit(t, erster, "sitz-1/riese/0").Schaden != 2:
		t.Errorf("der Riese trägt %d Schaden",
			objektMit(t, erster, "sitz-1/riese/0").Schaden)

	// **Und der Bär ist unversehrt** — der Riese teilt im ersten Schritt
	// nicht aus.
	case objektMit(t, erster, "sitz-2/baer/0").Schaden != 0:
		t.Errorf("der Bär trägt %d Schaden",
			objektMit(t, erster, "sitz-2/baer/0").Schaden)

	case erster.Zug.Kampf.Schadensschritte != 1:
		t.Errorf("gezählt sind %d Schadensschritte",
			erster.Zug.Kampf.Schadensschritte)
	}
}

// **Nach dem zweiten Schadensschritt teilt niemand mehr aus** — CR 510.4.
//
// **Der Zweig, den dieser Lauf betritt**: der frühe Ausstieg in
// `kampfschadenAlsHandlung`. Ohne ihn liefe die Handlung bei jedem Betreten
// des Schritts noch einmal, und ein Kampf teilte endlos Schaden aus.
func TestNachDemZweitenSchrittTeiltNiemandMehrAus(t *testing.T) {
	t.Parallel()

	erster := mitErstschlag(t)

	zweiter, fehler := erster.Passen(erster.Zug.Prioritaet)
	if fehler != nil {
		t.Fatalf("passen: %v", fehler)
	}

	zweiter, fehler = zweiter.Passen(zweiter.Zug.Prioritaet)
	if fehler != nil {
		t.Fatalf("zweites passen: %v", fehler)
	}

	if zweiter.Zug.Kampf.Schadensschritte != 2 {
		t.Fatalf("gezählt sind %d Schadensschritte",
			zweiter.Zug.Kampf.Schadensschritte)
	}

	// **Und noch einmal passen führt ins Kampfende** — nicht in einen
	// dritten Schadensschritt.
	dritter, fehler := zweiter.Passen(zweiter.Zug.Prioritaet)
	if fehler != nil {
		t.Fatalf("drittes passen: %v", fehler)
	}

	dritter, fehler = dritter.Passen(dritter.Zug.Prioritaet)
	if fehler != nil {
		t.Fatalf("viertes passen: %v", fehler)
	}

	switch {
	case dritter.Zug.Abschnitt.Schritt == domain.SchrittKampfschaden:
		t.Error("ein dritter Schadensschritt kam")
	case dritter.Zug.Kampf.Schadensschritte != 2:
		t.Errorf("gezählt sind %d Schadensschritte",
			dritter.Zug.Kampf.Schadensschritte)
	}
}

// **Nach einem Objekt, das es nicht gibt, fragt niemand mit Erfolg.**
//
// **Der leere Weg von `regeltext`** — er entsteht nur aus einem Kampfstand,
// dessen Kreatur inzwischen weg ist (CR 506.4), und darf nicht umfallen (§8).
func TestEinKampfstandOhneObjektFaelltNichtUm(t *testing.T) {
	t.Parallel()

	partie, fehler := mitAngriff(t).BlockerErklaeren("sitz-2",
		[]domain.Block{{
			Kreatur: "sitz-2/baer/0", Angreifer: "sitz-1/riese/0",
		}})
	if fehler != nil {
		t.Fatalf("blocker erklären: %v", fehler)
	}

	// **Die Kreatur verschwindet aus der Objektreihe** — ein Zustand, den
	// nur ein von Hand geänderter Bestand erzeugt.
	behalten := make([]domain.Spielobjekt, 0, len(partie.Objekte))

	for _, objekt := range partie.Objekte {
		if objekt.Kennung != "sitz-2/baer/0" {
			behalten = append(behalten, objekt)
		}
	}

	partie.Objekte = behalten

	danach := bisZumSchritt(t, partie, domain.SchrittKampfschaden)

	switch {
	case danach.Zug.Kampf.Schadensschritte != 1:
		t.Errorf("gezählt sind %d Schadensschritte",
			danach.Zug.Kampf.Schadensschritte)

	// **Der Riese trägt nichts** — der Blocker, den es nicht mehr gibt,
	// teilt nichts aus (CR 510.1d).
	case objektMit(t, danach, "sitz-1/riese/0").Schaden != 0:
		t.Errorf("der Riese trägt %d Schaden",
			objektMit(t, danach, "sitz-1/riese/0").Schaden)
	}
}
