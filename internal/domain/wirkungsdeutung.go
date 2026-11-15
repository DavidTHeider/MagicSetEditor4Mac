package domain

// **Was auf einer Karte steht, wird beim Verrechnen gelesen** — CR 608.2c
// (Fassung 2026-08-19),
// #1625.
//
// # Warum es das gibt
//
// **Weil die Rechnung bis zum 2026-09-19 eine Station zu früh saß.** `M8`
// (#1610)
// hat das Skript beim **Ankündigen** ausgewertet und das Ergebnis an den
// Zauber gehängt; beim Verrechnen wurde nur noch abgespult. Für alles Feste
// ist das dasselbe — „ziehe zwei Karten" zieht zwei Karten, gleich wann
// gerechnet wurde. Sichtbar wird der Unterschied, sobald sich zwischen den
// beiden Augenblicken etwas ändert, das in die Rechnung eingeht: **zwischen
// Ankündigen und Verrechnen liegt mindestens eine Prioritätsrunde.**
//
// # Warum ein eingesetzter Deuter und nicht eine offene Rechnung
//
// **Weil die Verrechnung mitten in `Passen` geschieht** (CR 117.3b über
// [Partie.nachDemPassen]) und nicht an einer Stelle, an der ein Aufrufer
// dazwischenkäme. Eine `Verrechnen`-Fassung, die eine offene Rechnung
// zurückgibt, müsste sich durch `Passen`, `Weiter` und jeden ihrer Rufer
// ziehen — und jeder von ihnen müsste eine Zwischenlage kennen, die ihn
// nichts angeht.
//
// **Die Domäne bleibt importfrei.** Die Schnittstelle steht hier, erfüllt
// wird sie in `service` — dieselbe Richtung wie `storage` zu `service` (§2).
//
// # Und was ohne Deuter geschieht
//
// **Nichts, und das ist eine Aussage.** Eine Partie ohne eingesetzten Deuter
// führt keine Kartenskripte aus; sie spielt nach dem Regelwerk und nach dem,
// was ein Mensch stellt. Das ist der Stand jeder Partie vor `M8` und der
// Normalfall in einem Lauf der Domäne, die keinen Auswerter hat.

// Wirkungsdeuter liest ein Kartenskript — **zum Zeitpunkt des Lesens**.
type Wirkungsdeuter interface {
	// Wirkungen gibt, was das Skript bewirken will — und einen **Befund**,
	// wenn es nicht durchgelaufen ist.
	//
	// **Ein Befund ist kein Fehler.** Ein gescheitertes Skript hält die
	// Partie nicht an (§8): der Zauber verrechnet, er bewirkt nichts, und
	// der Strom sagt warum.
	Wirkungen(skript string, lage Skriptlage) ([]Spielwirkung, string)
}

// Skriptlage ist, was ein Skript über seine Umstände wissen darf.
type Skriptlage struct {
	// Beherrscher ist, wem die Fähigkeit gehört — CR 113.8.
	Beherrscher string

	// Objekt ist die Quelle — CR 113.7.
	Objekt string

	// Ziele sind die genannten Ziele — CR 115.1.
	Ziele []string
}

// MitDeuter setzt, wer die Kartenskripte dieser Partie liest.
//
// **Er reist nicht mit in den Bestand.** Das Feld ist unexportiert, und
// `encoding/json` übergeht es; wer eine Partie lädt, setzt ihn neu. Das ist
// richtig so: ein Auswerter ist eine **Fähigkeit dieses Programms** und kein
// Teil des Spielstands.
func (p Partie) MitDeuter(deuter Wirkungsdeuter) Partie {
	p.deuter = deuter

	return p
}

// Skriptbefund sagt, was das zuletzt verrechnete Skript zu melden hatte —
// **leer**, wenn es nichts zu melden gab.
//
// **Er steht an der Partie und nicht im Rückgabewert**, weil die Verrechnung
// tief in `Passen` geschieht: ein zweiter Rückgabewert müsste sich durch
// dieselbe Kette ziehen wie die offene Rechnung oben.
func (p Partie) Skriptbefund() string { return p.skriptbefund }

// gedeutet rechnet die Wirkungen des Objekts an `stelle` — **jetzt**.
func (p Partie) gedeutet(stelle int) Partie {
	p.skriptbefund = ""

	if p.deuter == nil {
		return p
	}

	eigen, da := p.Eigenschaften(p.Objekte[stelle])
	if !da || eigen.Spielskript == "" {
		return p
	}

	wirkungen, befund := p.deuter.Wirkungen(eigen.Spielskript, Skriptlage{
		Beherrscher: p.Objekte[stelle].Beherrscher,
		Objekt:      p.Objekte[stelle].Kennung,
		Ziele:       zielkennungenAus(p.Objekte[stelle].Ziele),
	})

	p.skriptbefund = befund

	objekte := make([]Spielobjekt, len(p.Objekte))
	copy(objekte, p.Objekte)
	objekte[stelle].Wirkungen = wirkungen
	p.Objekte = objekte

	return p
}

// zielkennungenAus macht aus den festgehaltenen Zielen ihre Kennungen.
//
// **Ohne Zone**: ein Skript nennt ein Ziel, und wo es liegt, entscheidet
// CR 608.2b vor der Verrechnung — nicht das Skript.
func zielkennungenAus(ziele []Ziel) []string {
	kennungen := make([]string, 0, len(ziele))

	for _, ziel := range ziele {
		kennungen = append(kennungen, ziel.Kennung)
	}

	return kennungen
}
