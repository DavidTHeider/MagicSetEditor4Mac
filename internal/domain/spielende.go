package domain

// **Wie eine Partie endet** — Werkstück M17 der Regelmaschine
// (#1618).
//
// # Warum es dieses Werkstück zuerst gibt
//
// **Weil drei zustandsbasierte Aktionen „verliert die Partie" sagen**
// (CR 104.3b, 104.3c, 104.3d) und `M5` sie alle fünfundzwanzig umsetzt. Ohne
// einen Empfänger wäre das ein Satz ohne Wirkung — und genau die Gestalt aus
// §10, die dieses Projekt schon zweimal teuer bezahlt hat.
//
// # Und warum der Stand allein nicht reicht
//
// **`StandBeendet` gibt es seit G1, und er sagt nicht, warum.** Eine Partie,
// die beendet ist, weil jemand aufgegeben hat, und eine, die beendet ist,
// weil jemand bei null Leben stand, sind für das Turnier aus G7 zwei
// verschiedene Ergebnisse.

// Ausscheidungsgrund sagt, warum jemand die Partie verlassen hat.
//
// **Fünf, und jeder trägt seine Regel.** Ein Grund ohne Nummer wäre eine
// Annahme (§11). **Der fünfte gehört einer Variante** — er greift nur, wo
// das Profil sie führt.
type Ausscheidungsgrund string

const (
	// GrundAufgabe — CR 104.3a: aufgeben geht **jederzeit**, und wer
	// aufgibt, verlässt die Partie sofort. Keine Priorität, kein Warten.
	GrundAufgabe Ausscheidungsgrund = "aufgabe"

	// GrundLeben — CR 104.3b: null oder weniger Leben.
	//
	// **Und zwar beim nächsten Mal, wenn jemand Priorität bekäme** — es ist
	// eine zustandsbasierte Aktion (CR 704), und die baut `M5`.
	GrundLeben Ausscheidungsgrund = "leben"

	// GrundBibliothek — CR 104.3c: aus einer leeren Bibliothek ziehen.
	GrundBibliothek Ausscheidungsgrund = "bibliothek"

	// GrundGift — CR 104.3d: zehn oder mehr Vergiftungsmarken.
	GrundGift Ausscheidungsgrund = "gift"

	// GrundCommander — CR 903.10a: einundzwanzig Kampfschaden desselben
	// Commanders (M19,
	// #1620).
	//
	// **Der fünfte, und er steht nicht in CR 104.3**: die Variante fügt ihn
	// hinzu. CR 903.10 sagt ausdrücklich, dass alle anderen Regeln zum
	// Ende einer Partie weiter gelten.
	GrundCommander Ausscheidungsgrund = "commander"
)

// Ausgeschieden ist ein Sitz, der die Partie verlassen hat.
type Ausgeschieden struct {
	Sitz  string
	Grund Ausscheidungsgrund
}

// Partieausgang ist, wie die Partie ausgegangen ist.
//
// **`Partieausgang` und nicht `Ausgang`**: den Namen trägt schon der Ausgang
// eines geprüften Rufs im Protokoll (`vorgang.go`), und zwei verschiedene
// Dinge unter einem Namen sind ein Fund, den man später macht (§9).
//
// **Leer heißt: sie läuft noch.** Ein Ausgang ohne Sieger und ohne
// Unentschieden ist keine Aussage über den Ausgang, sondern das Fehlen einer.
type Partieausgang struct {
	// Sieger ist die Sitzkennung des Gewinners — leer heißt: keiner.
	Sieger string

	// Unentschieden — CR 104.4a: alle Verbliebenen verlieren gleichzeitig.
	Unentschieden bool

	// Ausgeschieden sind die, die gegangen sind, in der Reihenfolge, in der
	// sie gingen.
	Ausgeschieden []Ausgeschieden
}

// Scheidetaus lässt einen Sitz die Partie verlassen — CR 104.3.
//
// # Was dabei geschieht
//
//  1. **Der Sitz ist draußen**, mit seinem Grund.
//  2. **Seine Objekte gehen mit** — CR 800.4a.
//  3. **Der Ausgang wird geprüft** — CR 104.2a und CR 104.4a.
//
// **Ein zweites Ausscheiden desselben Sitzes ändert nichts.** Eine
// zustandsbasierte Aktion wiederholt sich, bis keine mehr anliegt (CR 704.3);
// wer schon draußen ist, steht nicht zweimal in der Liste.
//
// **Und ein Sitz, den es nicht gibt, scheidet nicht aus.** Er kann nur aus
// einem Bestand kommen, den jemand von Hand geändert hat (§8).
func (p Partie) Scheidetaus(sitz string, grund Ausscheidungsgrund) Partie {
	return p.Scheidenaus([]Ausgeschieden{{Sitz: sitz, Grund: grund}})
}

// Scheidenaus lässt mehrere Sitze **gleichzeitig** gehen — CR 704.4.
//
// # Warum es beide Wege gibt
//
// **Weil „gleichzeitig" eine Aussage über das Ergebnis ist.** Zustandsbasierte
// Aktionen laufen alle auf einmal (CR 704.4): stehen zwei Spieler bei null
// Leben, verlieren sie **zusammen**, und CR 104.4a macht daraus ein
// Unentschieden. Nacheinander wäre es etwas anderes — nach dem ersten hätte
// der zweite schon gewonnen (CR 104.2a, „immediately").
//
// **Aufgeben ist der einzelne Fall** (CR 104.3a) und geht über
// [Partie.Scheidetaus].
func (p Partie) Scheidenaus(mehrere []Ausgeschieden) Partie {
	// **Eine beendete Partie endet nicht noch einmal** — CR 104.1: sie endet
	// **sofort**, und was danach käme, geschieht in keinem Spiel mehr.
	if p.Stand == StandBeendet {
		return p
	}

	gegangen := make([]Ausgeschieden, 0, len(p.Ausgang.Ausgeschieden)+len(mehrere))
	gegangen = append(gegangen, p.Ausgang.Ausgeschieden...)

	for _, einer := range mehrere {
		if p.SitzVon(einer.Sitz) == nil || p.istDraussen(einer.Sitz) {
			continue
		}

		gegangen = append(gegangen, einer)
		p.Objekte = ohneObjekteVon(p.Objekte, einer.Sitz)
		p.Ausgang.Ausgeschieden = gegangen
	}

	p.Ausgang.Ausgeschieden = gegangen

	// **Monarch und Initiative rücken nach** — CR 725.4, CR 726.4 (M20,
	// #1621).
	// **Im selben Augenblick**, sagen beide Regeln, und deshalb steht es
	// hier und nicht in einer zustandsbasierten Prüfung danach.
	//
	// **Und danach der Zug** — wer die Priorität hatte oder am Zug war, ist
	// es nicht mehr (CR 800.4a, CR 800.4j, Fassung 2026-09-25;
	// [Partie.zugNachDemGehen]).
	return p.bezeichnungenNachruecken().ausgangPruefen().zugNachDemGehen()
}

// NochImSpiel sagt, wie viele Sitze noch dabei sind.
func (p Partie) NochImSpiel() int {
	uebrig := 0

	for _, sitz := range p.Sitze {
		if !p.istDraussen(sitz.Kennung) {
			uebrig++
		}
	}

	return uebrig
}

// Gegangen sagt, ob dieser Sitz die Partie verlassen hat.
func (p Partie) Gegangen(sitz string) bool { return p.istDraussen(sitz) }

// PruefeAbsender weist einen Befehl von einem Sitz ab, der gegangen ist.
//
// **Mit ihm haben seine Objekte das Spiel verlassen** — CR 800.4a (Fassung
// 2026-09-25) —, und was er danach stellte, veränderte eine Partie, an der er
// nicht mehr teilnimmt: sein Leben, sein Mana, ein Angriff, ein zweites
// Aufgeben (#1979). Ob es den Sitz gibt, fragt der Rufer vorher; hier geht
// es nur darum, ob er noch dabei ist.
func (p Partie) PruefeAbsender(sitz string) error {
	if !p.istDraussen(sitz) {
		return nil
	}

	return NewFieldError("absender", SchluesselSitzGegangen,
		"%q hat die Partie verlassen und stellt nichts mehr", sitz)
}

// istDraussen sagt, ob dieser Sitz die Partie schon verlassen hat.
func (p Partie) istDraussen(sitz string) bool {
	for _, gegangen := range p.Ausgang.Ausgeschieden {
		if gegangen.Sitz == sitz {
			return true
		}
	}

	return false
}

// ausgangPruefen entscheidet, ob die Partie damit vorbei ist.
//
// **CR 104.2a**: wer übrig bleibt, während alle Gegner gegangen sind,
// gewinnt — sofort, und es überstimmt alles, was das verhindern wollte.
//
// **CR 104.4a**: verlieren alle Verbliebenen gleichzeitig, ist es
// unentschieden.
//
// **Und bei einem einzigen Sitz geschieht nichts.** CR 104.2a verlangt
// Gegner, die gegangen sind; wer allein anfängt (Solo Goldfish, #1460), hat
// keine — und ein Aufbau ist kein Spiel (CR 100.1).
func (p Partie) ausgangPruefen() Partie {
	if len(p.Ausgang.Ausgeschieden) == 0 {
		return p
	}

	uebrig := ""
	zahl := 0

	for _, sitz := range p.Sitze {
		if !p.istDraussen(sitz.Kennung) {
			uebrig = sitz.Kennung
			zahl++
		}
	}

	switch zahl {
	case 0:
		p.Ausgang.Unentschieden = true

	case 1:
		p.Ausgang.Sieger = uebrig

	default:
		return p
	}

	p.Stand = StandBeendet

	return p
}

// ohneObjekteVon gibt die Objekte ohne die dieses Eigners — CR 800.4a.
//
// **Am Eigner und nicht am Beherrscher**: wer geht, nimmt mit, was **ihm**
// gehört; eine gestohlene Kreatur bleibt liegen und kehrt zu niemandem
// zurück, weil ihr Eigner noch da ist.
func ohneObjekteVon(objekte []Spielobjekt, sitz string) []Spielobjekt {
	geblieben := make([]Spielobjekt, 0, len(objekte))

	for _, objekt := range objekte {
		if objekt.Eigner != sitz {
			geblieben = append(geblieben, objekt)
		}
	}

	return geblieben
}
