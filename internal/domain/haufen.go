package domain

// **Ein Haufen ist eine Anordnung und keine Zone** —
// #1756,
// für PLAY-03.AK-03 und AK-04
// (#1711).
//
// # Warum es das gibt
//
// **[Spielobjekt.Haufen] lag seit seiner Entstehung ohne Erzeuger da.**
// Gemessen am 2026-09-21 über den ganzen Go-Bestand: die einzige Zuweisung
// war `objekt.Haufen = ""` beim Zonenwechsel. Kein Dienst, kein Befehl,
// keine Bindung füllte ihn je — und gelesen wurde er auch nicht. Ein Feld,
// das durch fünf Schichten reist und nirgends entsteht, ist ein Befund und
// kein Gestaltungsspielraum.
//
// # Was CR 400.5 erlaubt
//
// In einer Zone, deren Reihenfolge nichts bedeutet, ordnet ein Spieler seine
// Objekte an, wie er will — **das ist der Haufen**. Sieben Wälder
// übereinander sind keine Unterzone des Schlachtfelds, sondern eine Geste am
// Tisch. In einer **geordneten** Zone gibt es ihn deshalb nicht: dort gilt
// schon eine Reihenfolge, und eine zweite darüber wäre keine Anordnung mehr,
// sondern eine Behauptung.
//
// # Und wie AK-04 eingelöst wird
//
// *„Ein zusammengefasster Stapel darf individuelle Ziele oder getappte
// Zustände nicht verschleiern."*
//
// **Nicht durch eine Prüfung hinterher, sondern durch die Bedingung
// davor**: ein Haufen entsteht nur aus Objekten, die in **jedem gezeichneten
// Zustand** gleich sind — gleich getappt, gleich verdeckt, ohne Marken, ohne
// etwas daran Hängendes, in derselben Zone und unter derselben Kontrolle.
// Dann kann er per Konstruktion nichts verschleiern (§10).
//
// **Was das kostet, gehört dazu**: wer eine Karte eines Haufens tappt, nimmt
// sie damit heraus. Das ist die richtige Antwort — eine getappte Karte in
// einem Haufen ungetappter wäre genau die Verschleierung, die AK-04
// ausschließt. Den Griff dazu hat der Aufrufer: [Partie.Enthaeufe].

import "fmt"

// Haeufe legt ein Objekt auf den Haufen eines anderen — CR 400.5.
//
// **Der Haufen trägt die Kennung dessen, auf den gelegt wurde.** Eine eigene
// Kennung wäre eine zweite Größe, die jemand vergeben müsste; so entsteht
// der Name aus dem ersten Griff. Liegt das Ziel schon in einem Haufen,
// kommt der Neue in **denselben** und nicht in einen zweiten.
func (p Partie) Haeufe(wen, worauf string) (Partie, error) {
	if wen == worauf {
		return p, fmt.Errorf(
			"%w: %q lässt sich nicht auf sich selbst legen",
			ErrValidation, wen)
	}

	stelle, fehler := p.objektstelle(wen)
	if fehler != nil {
		return p, fehler
	}

	ziel, fehler := p.objektstelle(worauf)
	if fehler != nil {
		return p, fehler
	}

	if fehler := haufenfaehig(p.Objekte[stelle], p.Objekte[ziel]); fehler != nil {
		return p, fehler
	}

	objekte := make([]Spielobjekt, len(p.Objekte))
	copy(objekte, p.Objekte)

	name := objekte[ziel].Haufen
	if name == "" {
		name = objekte[ziel].Kennung
		objekte[ziel].Haufen = name
	}

	objekte[stelle].Haufen = name
	p.Objekte = objekte

	return p, nil
}

// Enthaeufe nimmt ein Objekt aus seinem Haufen.
//
// **Nur dieses eine.** Wer eine Karte herausnimmt, löst den Haufen nicht auf
// — die anderen liegen weiter beieinander. Ein Haufen aus einer einzigen
// Karte bleibt dabei stehen: er ist kein Fehler, sondern das Ergebnis, und
// der nächste Griff legt wieder etwas darauf.
func (p Partie) Enthaeufe(wen string) (Partie, error) {
	stelle, fehler := p.objektstelle(wen)
	if fehler != nil {
		return p, fehler
	}

	// **Ein stilles Nichts wäre eine Antwort, die aussieht wie ein Erfolg**
	// (§8): wer etwas aus einem Haufen nimmt, in dem es nicht liegt, hat
	// sich vergriffen und soll es lesen.
	if p.Objekte[stelle].Haufen == "" {
		return p, fmt.Errorf(
			"%w: %q liegt in keinem Haufen", ErrValidation, wen)
	}

	objekte := make([]Spielobjekt, len(p.Objekte))
	copy(objekte, p.Objekte)
	objekte[stelle].Haufen = ""
	p.Objekte = objekte

	return p, nil
}

// haufenfaehig sagt, ob zwei Objekte einen Haufen bilden dürfen.
//
// **Jede Bedingung hier ist ein Zustand, den ein Haufen verschleiern
// könnte** — deshalb entsteht er gar nicht erst.
func haufenfaehig(wen, worauf Spielobjekt) error {
	if wen.Zone != worauf.Zone {
		return fmt.Errorf(
			"%w: %q liegt in %q und %q in %q — ein Haufen liegt in einer Zone",
			ErrValidation, wen.Kennung, wen.Zone, worauf.Kennung, worauf.Zone)
	}

	// **CR 400.5**: die freie Anordnung gilt nur, wo die Reihenfolge nichts
	// bedeutet.
	if wen.Zone.Geordnet() {
		return fmt.Errorf(
			"%w: in %q liegt eine Reihenfolge fest, und darüber gibt es "+
				"keinen Haufen (CR 400.5)", ErrValidation, wen.Zone)
	}

	if wen.Beherrscher != worauf.Beherrscher {
		return fmt.Errorf(
			"%w: %q und %q stehen unter verschiedener Kontrolle",
			ErrValidation, wen.Kennung, worauf.Kennung)
	}

	if grund := gleichGezeichnet(wen, worauf); grund != "" {
		return fmt.Errorf(
			"%w: %q und %q sehen verschieden aus — %s. Ein Haufen aus "+
				"beiden verschleierte das (PLAY-03.AK-04)",
			ErrValidation, wen.Kennung, worauf.Kennung, grund)
	}

	return nil
}

// gleichGezeichnet gibt den Grund, warum zwei Objekte verschieden aussehen —
// oder die leere Zeichenkette.
//
// **Ein Grund und kein `bool`**: „geht nicht" ist keine Auskunft, und §8
// verlangt, dass eine Meldung sagt, woran es lag.
func gleichGezeichnet(wen, worauf Spielobjekt) string {
	switch {
	case wen.Getappt != worauf.Getappt:
		return "einer ist getappt und der andere nicht"

	case wen.Verdeckt != worauf.Verdeckt:
		return "einer liegt verdeckt und der andere offen"

	case len(wen.Zaehler) > 0 || len(worauf.Zaehler) > 0:
		return "einer trägt Marken"

	case wen.HaengtAn != "" || worauf.HaengtAn != "":
		return "einer hängt an etwas"
	}

	return ""
}
