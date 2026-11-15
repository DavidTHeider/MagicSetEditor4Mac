package domain

// **Ersetzung und Verhinderung** — Werkstück M11 von Phase 5.5
// (#1613).
//
// # Warum sie nicht auf den Stapel kommen
//
// **Weil sie nichts tun, sondern etwas *ändern*.** CR 614.1: ein
// Ersetzungseffekt wartet auf ein Ereignis und setzt ein anderes an seine
// Stelle; CR 614.6 sagt, dass das ursprüngliche **nie geschieht**. Ein
// Ersetzungseffekt, der über den Stapel ginge, käme zu spät.
//
// **Der Unterschied zum Dauereffekt** (`M10`,
// #1612):
// der ändert, **was ein Objekt ist**; dieser ändert, **was geschieht**.
//
// # Welche zwei Ereignisse hier ersetzbar sind
//
// **Die zwei, die dieses Programm überhaupt kennt und auslöst.**
//
//   - **Etwas geht in den Friedhof** (CR 404.1) — ersetzt durch das Exil.
//     Das ist die erste Abnahme des Issues: „wird eine Kreatur stattdessen
//     ins Exil geschickt, erreicht sie den Friedhof nie".
//   - **Schaden wird zugefügt** (CR 120.3) — verhindert, ganz oder zum Teil
//     (CR 615.10).
//
// **Was fehlt, fehlt benannt**: ein Ereignis, das dieses Programm nicht als
// Ereignis führt, lässt sich nicht ersetzen — Ziehen, Wirken, das
// Hineinkommen eines Permanents. Sie stehen als Grenze in der Regelmatrix.
//
// # Und wie CR 616.1 hier entschieden wird
//
// **Vom Zeitstempel und nicht vom Spieler.** CR 616.1 lässt den Betroffenen
// wählen, welcher von mehreren zuerst angewandt wird; hier gilt der älteste.
// **Das ist eine benannte Abweichung** (§15, Punkt 4) und keine Auslassung:
// die Wahl ließe sich stellen wie die Reihung aus `M9`, und sie ist nicht
// gestellt.
//
// **Und sie hat heute keine sichtbare Folge**, was dazugehört (§10): beide
// Ereignisse kennen genau **eine** Ersetzungsart, und mehrere Verhinderungen
// ergeben in jeder Reihenfolge dieselbe Zahl. Die feste Ordnung steht
// trotzdem da — ohne sie gäbe dieselbe Partie zweimal verschiedene
// Ergebnisse, und das bräche die Hashkette aus `G1`.
//
// **Und CR 614.5 gilt je Ereignis**: ein Effekt bekommt **eine** Gelegenheit
// für ein Ereignis **und die veränderten Ereignisse, die es ersetzen** — und
// beim nächsten Ereignis wieder eine. Wer das an den Zug knüpfte, hätte einen
// Effekt, der nach seiner ersten Kreatur schwiege.
//
// **Die Gelegenheit ist deshalb lokal zum Ereignis**, und die übrigen
// Effekte kommen danach an dasselbe, **veränderte** Ereignis (CR 616.1,
// zweiter Satz).
//
// # Und was das Issue verlangt und hier nicht steht
//
// **„Die Wahl der Reihenfolge steht im Protokoll."** Sie steht im
// **Ergebnis** — eine Kreatur liegt im Exil statt im Friedhof, ein Schaden
// ist markiert oder nicht —, aber nicht als eigener Eintrag im
// Ereignisstrom. **Das ist eine halbe Zusage, und sie steht hier als
// solche** (§10): ein Eintrag bräuchte ein Ereignis je Ersetzung, und die
// gibt es nicht, solange niemand gefragt wird.

// Ereignisart sagt, worauf ein Ersetzungseffekt wartet — CR 614.1.
type Ereignisart string

const (
	// EreignisFriedhof — ein Objekt geht in den Friedhof (CR 404.1).
	EreignisFriedhof Ereignisart = "friedhof"

	// EreignisSchaden — einem Objekt oder Sitz wird Schaden zugefügt
	// (CR 120.3).
	EreignisSchaden Ereignisart = "schaden"
)

// Ersetzungsart sagt, was an die Stelle tritt.
type Ersetzungsart string

const (
	// ErsetzungExil schickt ins Exil statt in den Friedhof — CR 614.1a
	// („stattdessen").
	ErsetzungExil Ersetzungsart = "exil"

	// ErsetzungVerhindern verhindert Schaden — CR 615.1.
	ErsetzungVerhindern Ersetzungsart = "verhindern"
)

// Ersetzungseffekt wartet auf ein Ereignis und setzt ein anderes an seine
// Stelle — CR 614.1.
type Ersetzungseffekt struct {
	// Ereignis ist, worauf er wartet.
	Ereignis Ereignisart

	// Ersatz ist, was an die Stelle tritt.
	Ersatz Ersetzungsart

	// Objekt ist, für welches Objekt er gilt — leer heißt: für jedes.
	Objekt string

	// Menge ist, wie viel verhindert wird — **null heißt: alles**
	// (CR 615.10 nennt eine Zahl, CR 615.1 nicht).
	Menge int

	// Zeitstempel entscheidet, welcher zuerst greift — CR 616.1 in der
	// Gestalt, die dieses Programm ihm gibt.
	Zeitstempel int

	// Dauer sagt, wann er endet — CR 611.1, dieselbe Frist wie beim
	// Dauereffekt.
	Dauer Dauer
}

// MitErsetzung legt einen Ersetzungseffekt an und gibt ihm seinen
// Zeitstempel.
func (p Partie) MitErsetzung(effekt Ersetzungseffekt) Partie {
	stempel := make([]int, 0, len(p.Ersetzungen))
	for _, lebend := range p.Ersetzungen {
		stempel = append(stempel, lebend.Zeitstempel)
	}

	effekt.Zeitstempel = naechsterZeitstempel(stempel)

	p.Ersetzungen = append(
		append([]Ersetzungseffekt{}, p.Ersetzungen...), effekt)

	return p
}

// ersetzungFuer sucht den ältesten Effekt, der auf dieses Ereignis wartet und
// noch nicht daran gearbeitet hat — CR 616.1 mit CR 614.5.
//
// **`schon` ist die Gelegenheit** (CR 614.5): die Zeitstempel der Effekte,
// die an **diesem** Ereignis bereits gegriffen haben. Sie lebt so lange wie
// das Ereignis und nicht länger — ein Effekt, der eine Kreatur ins Exil
// geschickt hat, tut es bei der nächsten wieder.
func (p Partie) ersetzungFuer(
	ereignis Ereignisart, ersatz Ersetzungsart, objekt string,
	schon map[int]bool,
) (int, bool) {
	gefunden := -1

	for stelle, effekt := range p.Ersetzungen {
		if schon[effekt.Zeitstempel] || effekt.Ereignis != ereignis ||
			effekt.Ersatz != ersatz {
			continue
		}

		if effekt.Objekt != "" && effekt.Objekt != objekt {
			continue
		}

		if gefunden < 0 ||
			p.Ersetzungen[gefunden].Zeitstempel > effekt.Zeitstempel {
			gefunden = stelle
		}
	}

	return gefunden, gefunden >= 0
}

// friedhofErsetzen sagt, wohin ein Objekt wirklich geht — CR 614.1a.
//
// **Die Zone und sonst nichts.** Einen dritten Rückgabewert „wurde ersetzt"
// läse kein Aufrufer — ein Wert, den niemand liest, ist keine Zusage (§10).
// Wer wissen will, ob ersetzt wurde, sieht die Zone an.
//
// **Ein Schritt genügt**: das veränderte Ereignis ist „geht ins Exil", und
// darauf wartet kein Effekt, den dieses Programm kennt. Wer einen zweiten
// baut, braucht hier die Schleife von [Partie.schadenVerhindern].
func (p Partie) friedhofErsetzen(kennung string) Zone {
	if _, da := p.ersetzungFuer(
		EreignisFriedhof, ErsetzungExil, kennung, nil); !da {
		return ZoneFriedhof
	}

	return ZoneExil
}

// schadenVerhindern rechnet aus, wie viel Schaden ankommt — CR 615.1.
//
// **Verhinderter Schaden wird nicht markiert** — die dritte Abnahme des
// Issues, und CR 615.1 sagt es: das Ereignis findet nicht statt.
//
// **Und nur so viel, wie der Effekt nennt** (CR 615.10): ein Effekt über
// eins verhindert eins und nicht drei. Null heißt: alles.
func (p Partie) schadenVerhindern(objekt string, menge int) int {
	// **Nach jeder Ersetzung wird das *veränderte* Ereignis noch einmal
	// gegen die übrigen gehalten** — CR 616.1, zweiter Satz. Und **jeder
	// nur einmal**, CR 614.5: `schon` merkt sich, wer gegriffen hat.
	schon := map[int]bool{}

	for menge > 0 {
		stelle, da := p.ersetzungFuer(
			EreignisSchaden, ErsetzungVerhindern, objekt, schon)
		if !da {
			return menge
		}

		schon[p.Ersetzungen[stelle].Zeitstempel] = true

		// **Ohne Zahl ist alles verhindert** — CR 615.1 gegen CR 615.10.
		wie := p.Ersetzungen[stelle].Menge
		if wie <= 0 {
			return 0
		}

		menge = max(0, menge-wie)
	}

	return menge
}

// ohneErsetzungenBisZugende beendet, was mit dem Zug endet — CR 514.2.
//
// **Und sonst nichts**: die Gelegenheit aus CR 614.5 hängt am Ereignis und
// nicht am Zug — sie wird hier weder genommen noch zurückgegeben.
func (p Partie) ohneErsetzungenBisZugende() Partie {
	uebrig := make([]Ersetzungseffekt, 0, len(p.Ersetzungen))

	for _, effekt := range p.Ersetzungen {
		if effekt.Dauer != DauerZugende {
			uebrig = append(uebrig, effekt)
		}
	}

	p.Ersetzungen = uebrig

	return p
}
