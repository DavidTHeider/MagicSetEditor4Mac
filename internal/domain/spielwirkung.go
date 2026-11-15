package domain

// **Die benannten Wege, über die eine Fähigkeit den Zustand ändert** —
// Werkstück M8 von Phase 5.5
// (#1610).
//
// # Wofür es sie gibt
//
// **Das Issue verlangt „einen Sandkasten mit Grenzen": eine Fähigkeit ändert
// den Zustand nur über benannte Wege, und ein Skript ohne Ende wird
// abgebrochen.** Hier stehen die Wege.
//
// **Ein Skript erzeugt sie und wendet sie nicht an.** Die Auswertung läuft in
// `internal/service` und bekommt Funktionen, die eine [Spielwirkung]
// **aufschreiben**; angewandt wird sie hier. Damit gibt es keinen Weg von
// einem Kartentext zu einem Feld der Partie, der nicht durch diese Liste
// führt — und was die Liste nicht kennt, kann ein Skript nicht tun.
//
// # Und CR 101.2 steht darin
//
// **„Was nicht geschehen kann, wird ausgelassen."** [Partie.Anwenden] gibt
// deshalb ein zweites Ergebnis zurück: eine Wirkung an einem Sitz, den es
// nicht gibt, ändert nichts und sagt es. Sie ist **kein Fehler** — ein
// Kartentext, der ins Leere greift, hält die Partie nicht an (§8).
//
// # Was hier nicht steht
//
// **CR 101.1** („die Karte schlägt die Regel") hat hier keinen Gegenstand:
// eine Wirkung **ist** das, was die Karte sagt, und es gibt keine Regel, die
// sie überschreiben müsste. **CR 101.3** („kann nicht schlägt kann") braucht
// ein Verbot, und Verbote gibt es erst mit `M10`
// (#1612).
// Beides steht als Grenze in der Regelmatrix.

// Wirkungsart sagt, was eine Wirkung tut.
type Wirkungsart string

const (
	// WirkungZiehen — CR 121.1.
	WirkungZiehen Wirkungsart = "ziehen"

	// WirkungLeben — CR 119.3 (gewinnen) und CR 119.4 (verlieren).
	WirkungLeben Wirkungsart = "leben"

	// WirkungSchaden markiert Schaden an einem Permanent — CR 120.3.
	WirkungSchaden Wirkungsart = "schaden"

	// WirkungZerstoeren — CR 701.8a.
	WirkungZerstoeren Wirkungsart = "zerstoeren"

	// WirkungExilieren — CR 701.13a.
	WirkungExilieren Wirkungsart = "exilieren"

	// WirkungMarke legt Marken auf ein Objekt oder nimmt sie — CR 122.1.
	//
	// **Nicht CR 121.1** — das ist das Ziehen einer Karte.
	WirkungMarke Wirkungsart = "marke"

	// WirkungMahlen — CR 701.17a.
	WirkungMahlen Wirkungsart = "mahlen"

	// WirkungTappen — CR 701.26a.
	WirkungTappen Wirkungsart = "tappen"

	// WirkungEnttappen — CR 701.26b.
	//
	// **Nicht CR 701.20a/b** — das ist das Aufdecken einer Karte.
	WirkungEnttappen Wirkungsart = "enttappen"

	// WirkungWerte ändert Stärke und Widerstandskraft bis zum Zugende —
	// CR 613.1g mit CR 613.4c (M10,
	// #1612).
	WirkungWerte Wirkungsart = "werte"

	// WirkungFaehigkeit verleiht eine Schlüsselwortfähigkeit bis zum
	// Zugende — CR 613.1f (M10).
	WirkungFaehigkeit Wirkungsart = "faehigkeit"

	// WirkungStattdessen ersetzt den Friedhof durch das Exil — CR 614.1a
	// (M11, #1613).
	WirkungStattdessen Wirkungsart = "stattdessen"

	// WirkungVerhindern verhindert Schaden — CR 615.1 (M11).
	WirkungVerhindern Wirkungsart = "verhindern"
)

// Wirkungsarten gibt alle dreizehn.
//
// **Exportiert, weil ein Wächter sie zählt** und weil die Bibliothek in
// `internal/service` je Art eine Funktion anbietet — zwei Listen, die
// auseinanderlaufen, wären eine Fähigkeit, die niemand rufen kann (§10).
func Wirkungsarten() []Wirkungsart {
	return []Wirkungsart{
		WirkungZiehen, WirkungLeben, WirkungSchaden, WirkungZerstoeren,
		WirkungExilieren, WirkungMarke, WirkungMahlen, WirkungTappen,
		WirkungEnttappen, WirkungWerte, WirkungFaehigkeit,
		WirkungStattdessen, WirkungVerhindern,
	}
}

// Spielwirkung ist eine Änderung, die eine Fähigkeit bewirken darf.
//
// **Ein Verbund und kein Rückruf**: sie geht durch den Ereignisstrom, durch
// die Naht und über vier Flächen (§16). Eine Funktion ließe sich nicht
// aufschreiben.
type Spielwirkung struct {
	Art Wirkungsart

	// Wer ist der betroffene Sitz — leer, wo ein Objekt gemeint ist.
	Wer string

	// Objekt ist das betroffene Objekt — leer, wo ein Sitz gemeint ist.
	Objekt string

	// Marke ist die Art der Marke — nur bei [WirkungMarke]. Bei
	// [WirkungFaehigkeit] steht hier das Schlüsselwort, bei [WirkungWerte]
	// die Widerstandskraft als Zahl im Text.
	//
	// **Ein Feld und nicht drei**: eine Wirkung geht durch den
	// Ereignisstrom und über vier Flächen (§16); ein Verbund mit einem
	// Feld je Art trüge an jeder Stelle acht leere.
	Marke string

	// Menge ist, wie viel. **Sie darf negativ sein**: Leben verlieren ist
	// Leben gewinnen mit anderem Vorzeichen (CR 119.3 und CR 119.4 sind
	// dieselbe Rechnung), und eine Marke lässt sich nehmen.
	Menge int
}

// Anwenden führt eine Wirkung aus — und sagt, ob sie geschehen konnte.
//
// **Das zweite Ergebnis ist CR 101.2**: „was nicht geschehen kann, wird
// ausgelassen". Ein `false` ist **kein Fehler**; es ist die Auskunft, dass
// diese Wirkung nichts getan hat.
func (p Partie) Anwenden(wirkung Spielwirkung) (Partie, bool) {
	switch wirkung.Art {
	case WirkungZiehen:
		return p.wirkeZiehen(wirkung)

	case WirkungLeben:
		return p.wirkeLeben(wirkung)

	case WirkungMahlen:
		return p.wirkeMahlen(wirkung)

	case WirkungSchaden, WirkungMarke, WirkungTappen, WirkungEnttappen:
		return p.wirkeAmObjekt(wirkung)

	case WirkungZerstoeren, WirkungExilieren:
		return p.wirkeZonenwechsel(wirkung)

	case WirkungWerte, WirkungFaehigkeit:
		return p.wirkeDauereffekt(wirkung)

	case WirkungStattdessen, WirkungVerhindern:
		return p.wirkeErsetzung(wirkung)
	}

	return p, false
}

// wirkeZiehen zieht Karten — CR 121.1.
func (p Partie) wirkeZiehen(wirkung Spielwirkung) (Partie, bool) {
	if wirkung.Menge <= 0 || p.SitzVon(wirkung.Wer) == nil {
		return p, false
	}

	for range wirkung.Menge {
		p = p.ziehtEine(wirkung.Wer)
	}

	return p, true
}

// wirkeLeben ändert den Lebensstand — CR 119.3, CR 119.4.
func (p Partie) wirkeLeben(wirkung Spielwirkung) (Partie, bool) {
	if wirkung.Menge == 0 || p.SitzVon(wirkung.Wer) == nil {
		return p, false
	}

	sitze := make([]Sitz, len(p.Sitze))
	copy(sitze, p.Sitze)
	lebengeben(sitze, wirkung.Wer, wirkung.Menge)
	p.Sitze = sitze

	return p, true
}

// wirkeMahlen legt Karten von oben in den Friedhof — CR 701.17a.
func (p Partie) wirkeMahlen(wirkung Spielwirkung) (Partie, bool) {
	if wirkung.Menge <= 0 {
		return p, false
	}

	danach, fehler := p.Mahlen(wirkung.Wer, wirkung.Menge)
	if fehler != nil {
		return p, false
	}

	return danach, true
}

// wirkeAmObjekt ändert etwas an einem Objekt, das liegen bleibt.
func (p Partie) wirkeAmObjekt(wirkung Spielwirkung) (Partie, bool) {
	stelle, fehler := p.objektstelle(wirkung.Objekt)
	if fehler != nil {
		return p, false
	}

	objekte := make([]Spielobjekt, len(p.Objekte))
	copy(objekte, p.Objekte)

	switch wirkung.Art {
	case WirkungSchaden:
		if wirkung.Menge <= 0 {
			return p, false
		}

		// **Verhinderter Schaden wird nicht markiert** — CR 615.1 (M11).
		uebrig := p.schadenVerhindern(wirkung.Objekt, wirkung.Menge)
		if uebrig == 0 {
			return p, true
		}

		objekte[stelle].Schaden += uebrig

	case WirkungMarke:
		if wirkung.Menge == 0 || wirkung.Marke == "" {
			return p, false
		}

		// **Eine eigene Abbildung und keine geteilte** — die Partie ist
		// ein Wertetyp, und eine Map geht als Zeiger mit (CR 733 lebt
		// davon, dass ein alter Stand alt bleibt).
		zaehler := make(map[string]int, len(objekte[stelle].Zaehler)+1)
		for art, wie := range objekte[stelle].Zaehler {
			zaehler[art] = wie
		}

		// **Unter null gibt es keine** — dieselbe Regel wie bei jedem
		// Zähler (CR 122.1).
		zaehler[wirkung.Marke] = max(0, zaehler[wirkung.Marke]+wirkung.Menge)
		objekte[stelle].Zaehler = zaehler

	case WirkungTappen:
		objekte[stelle].Getappt = true

	case WirkungEnttappen:
		objekte[stelle].Getappt = false
	}

	p.Objekte = objekte

	return p, true
}

// wirkeZonenwechsel schickt ein Objekt in eine andere Zone.
func (p Partie) wirkeZonenwechsel(wirkung Spielwirkung) (Partie, bool) {
	wechsel := Partie.Zerstoeren
	if wirkung.Art == WirkungExilieren {
		wechsel = Partie.Exilieren
	}

	danach, fehler := wechsel(p, wirkung.Objekt)
	if fehler != nil {
		return p, false
	}

	return danach, true
}

// wirkeDauereffekt legt einen Effekt an, der bis zum Zugende gilt —
// CR 613.1f und CR 613.1g mit CR 611.2b.
//
// **Bis zum Zugende und nicht unbestimmt**: ein Kartentext, der einen
// Dauereffekt **ohne** Frist erzeugt, sagt das ausdrücklich („solange …"), und
// solche Texte gibt es hier nicht (CR 611.2b steht als Grenze). Die Frist ist
// damit die einzige, die ein Skript stellen kann — und die häufigste.
func (p Partie) wirkeDauereffekt(wirkung Spielwirkung) (Partie, bool) {
	if _, fehler := p.objektstelle(wirkung.Objekt); fehler != nil {
		return p, false
	}

	effekt := Dauereffekt{
		Objekt: wirkung.Objekt, Dauer: DauerZugende,
		Schicht: Schicht7c, Staerke: wirkung.Menge,
		Widerstand: wirkung.Menge,
	}

	if wirkung.Art == WirkungFaehigkeit {
		if !istSchluesselwort(wirkung.Marke) {
			return p, false
		}

		effekt = Dauereffekt{
			Objekt: wirkung.Objekt, Dauer: DauerZugende,
			Schicht: Schicht6, Faehigkeit: wirkung.Marke,
		}
	}

	return p.MitDauereffekt(effekt), true
}

// wirkeErsetzung legt einen Ersetzungs- oder Verhinderungseffekt an —
// CR 614.1a und CR 615.1.
//
// **Bis zum Zugende**, aus demselben Grund wie bei [Partie.wirkeDauereffekt]:
// eine andere Frist kann ein Skript nicht nennen.
//
// **Und ohne Objekt gilt er für jedes** — CR 614.1 lässt beides zu, und ein
// Skript, das kein Objekt nennt, meint das allgemeine.
func (p Partie) wirkeErsetzung(wirkung Spielwirkung) (Partie, bool) {
	effekt := Ersetzungseffekt{
		Objekt: wirkung.Objekt, Dauer: DauerZugende,
		Ereignis: EreignisFriedhof, Ersatz: ErsetzungExil,
	}

	if wirkung.Art == WirkungVerhindern {
		effekt.Ereignis = EreignisSchaden
		effekt.Ersatz = ErsetzungVerhindern
		effekt.Menge = max(0, wirkung.Menge)
	}

	return p.MitErsetzung(effekt), true
}
