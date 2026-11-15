package domain

// **Schichten, Zeitstempel und die Dauer eines Effekts** — Werkstück M10 von
// Phase 5.5
// (#1612).
//
// # Warum es Schichten gibt und nicht eine Reihenfolge
//
// **Weil dieselben zwei Effekte in anderer Reihenfolge andere Werte
// ergäben.** Eine Kreatur, die auf 5/5 gesetzt wird und +1/+1 bekommt, ist
// 6/6 — gleich, welcher Effekt zuerst da war. CR 613.1 legt sieben Schichten
// fest, und **innerhalb** einer Schicht entscheidet der Zeitstempel
// (CR 613.7).
//
// # Welche Schichten hier wirken
//
// **Zwei von sieben**, und das ist keine Willkür: es sind die, die einen
// Gegenstand haben.
//
//   - **Schicht 6** (CR 613.1f): ein Effekt verleiht eine
//     Schlüsselwortfähigkeit. Damit fällt die halbe Grenze aus `M14`
//     (#1616) —
//     „ein Effekt, der ein Schlüsselwort verleiht, wirkt nicht".
//   - **Schicht 7** (CR 613.1g), und zwar `7b` (setzen) und `7c` (ändern) —
//     CR 613.4b und CR 613.4c.
//
// **Die fünf anderen stehen als Konstanten da und wirken nicht.** Schicht 1
// (Kopien, CR 613.1a) hat `M18`, Schicht 2 (Kontrolle, CR 613.1b) hat `M12`
// — beide ändern das Objekt **unmittelbar** statt über einen Dauereffekt, und
// das ist eine benannte Abweichung. Schicht 3 (Text, CR 612), Schicht 4
// (Typen) und Schicht 5 (Farbe) haben keinen Effekt, der sie erzeugte.
//
// # Und was CR 613.8 hier nicht hat
//
// **Einen Fall.** Eine
// Abhängigkeit entsteht **innerhalb derselben Schicht oder Unterschicht**
// (CR 613.8, CR 613.8a): ein Effekt hängt von einem zweiten ab, wenn dessen
// Anwendung ändern würde, ob der erste überhaupt besteht, worauf er wirkt
// oder was er dort tut. Zwischen zwei **verschiedenen** Schichten entsteht
// nie eine — dort entscheidet die Schichtnummer, und das ist die gewöhnliche
// Reihenfolge und keine Abhängigkeit.
//
// **Innerhalb der beiden wirkenden Schichten gibt es keinen**, und zwar aus
// drei Gründen, die alle am Code nachzulesen sind:
//
//   - Ein Effekt nennt sein Objekt bei der Kennung ([Dauereffekt.Objekt]),
//     und [Partie.dauereffekteFuer] sucht danach. Kein zweiter kann
//     verschieben, worauf er wirkt.
//   - Schicht 6 **verleiht** nur: [Partie.HatFaehigkeit] fragt nach dem
//     Wort und nimmt keines weg. Also kann kein Effekt einen anderen
//     aufheben — sie stehen ohnehin als Datensätze an der Partie und nicht
//     als Fähigkeiten auf einem Objekt.
//   - In 7b und 7c rechnet jeder mit **seinen eigenen** Zahlen
//     ([Partie.werteMitEffekten]) und liest den anderen nicht.
//
// **Und die dritte Bedingung aus CR 613.8a hat hier keinen Gegenstand**: sie
// unterscheidet merkmalsdefinierende Fähigkeiten, und deren Unterschicht 7a
// ist leer.
//
// **Ein Abhängigkeitssystem ohne Abhängigkeit wäre ein Zweig, den kein Weg
// betritt** (§10). Es steht als Grenze in der Regelmatrix, mit diesem Grund.

import "slices"

// Schicht sagt, wann ein Dauereffekt angewandt wird — CR 613.1.
type Schicht string

const (
	// Schicht1 — kopierbare Werte (CR 613.1a).
	Schicht1 Schicht = "1"

	// Schicht2 — Kontrollwechsel (CR 613.1b).
	Schicht2 Schicht = "2"

	// Schicht3 — textändernde Effekte (CR 613.1c, CR 612).
	Schicht3 Schicht = "3"

	// Schicht4 — typändernde Effekte (CR 613.1d).
	Schicht4 Schicht = "4"

	// Schicht5 — farbändernde Effekte (CR 613.1e).
	Schicht5 Schicht = "5"

	// Schicht6 — Fähigkeiten geben und nehmen (CR 613.1f).
	Schicht6 Schicht = "6"

	// Schicht7a — Merkmalsdefinierende Werte (CR 613.4a).
	Schicht7a Schicht = "7a"

	// Schicht7b — Werte **setzen** (CR 613.4b).
	Schicht7b Schicht = "7b"

	// Schicht7c — Werte **ändern** (CR 613.4c).
	Schicht7c Schicht = "7c"

	// Schicht7d — tauschen (CR 613.4d).
	Schicht7d Schicht = "7d"
)

// Schichten gibt die zehn in der Reihenfolge, in der sie gelten — CR 613.1.
//
// **Zehn und nicht sieben**: die siebte zerfällt in vier Unterschichten
// (CR 613.4a–d), und sie sind der Grund, warum es diese Ordnung überhaupt
// braucht.
func Schichten() []Schicht {
	return []Schicht{
		Schicht1, Schicht2, Schicht3, Schicht4, Schicht5, Schicht6,
		Schicht7a, Schicht7b, Schicht7c, Schicht7d,
	}
}

// Dauer sagt, wie lange ein Effekt gilt — CR 611.1.
type Dauer string

const (
	// DauerUnbestimmt heißt: bis etwas ihn beendet — der Nullwert.
	DauerUnbestimmt Dauer = ""

	// DauerZugende endet im Aufräumschritt — CR 514.2, zweite Hälfte.
	DauerZugende Dauer = "zugende"
)

// Dauereffekt ist ein Effekt, der eine Weile gilt — CR 611.1.
type Dauereffekt struct {
	// Schicht sagt, wann er angewandt wird — CR 613.1.
	Schicht Schicht

	// Zeitstempel ordnet ihn innerhalb seiner Schicht — CR 613.7.
	//
	// **Er wird vergeben und nicht mitgegeben**: [Partie.MitDauereffekt]
	// zählt hoch. Wer ihn selbst setzte, hätte zwei Effekte mit derselben
	// Zahl, und CR 613.7 hätte nichts zu entscheiden.
	Zeitstempel int

	// Objekt ist, worauf er wirkt.
	//
	// **Ein Objekt und keine Menge**: „alle Kreaturen, die du beherrschst"
	// braucht einen Kartentext, der eine Menge beschreibt, und das ist
	// `M8` in einer Gestalt, die es dort nicht gibt. Das steht als Grenze.
	Objekt string

	// Staerke und Widerstand sind die Zahlen der siebten Schicht.
	Staerke    int
	Widerstand int

	// Faehigkeit ist das Schlüsselwort der sechsten Schicht — leer, wo
	// keines gemeint ist.
	Faehigkeit string

	// Dauer sagt, wann er endet — CR 611.1.
	Dauer Dauer
}

// naechsterZeitstempel ist der Stempel eines neuen Effekts: **einer über dem
// höchsten lebenden** — CR 613.7 und CR 616.1 (Fassung 2026-09-25) ordnen
// nach ihm, und kein
// lebender Effekt darf ihn schon tragen. Hier stand die Länge der Liste plus
// eins; endete ein Effekt im Aufräumschritt, bekam der nächste den Stempel
// eines noch lebenden (gefunden über die Zufallsmutation, B05, Nachlauf über
// Phase B).
func naechsterZeitstempel(lebende []int) int {
	hoechster := 0
	for _, stempel := range lebende {
		hoechster = max(hoechster, stempel)
	}

	return hoechster + 1
}

// MitDauereffekt legt einen Effekt an und gibt ihm seinen Zeitstempel.
//
// **Exportiert, weil `service` ihn aus einem Kartenskript anlegt** — und
// weil ein Lauf ihn stellen können muss, ohne eine Karte zu bauen.
func (p Partie) MitDauereffekt(effekt Dauereffekt) Partie {
	stempel := make([]int, 0, len(p.Dauereffekte))
	for _, lebend := range p.Dauereffekte {
		stempel = append(stempel, lebend.Zeitstempel)
	}

	effekt.Zeitstempel = naechsterZeitstempel(stempel)

	p.Dauereffekte = append(
		append([]Dauereffekt{}, p.Dauereffekte...), effekt)

	return p
}

// dauereffekteFuer gibt die Effekte an einem Objekt, in Reihenfolge —
// CR 613.1 und CR 613.7.
//
// **Erst die Schicht, dann der Zeitstempel.** Eine Liste, die nur nach dem
// Zeitstempel ginge, gäbe einer Kreatur andere Werte, je nachdem, wer zuerst
// da war — genau das verhindert CR 613.
func (p Partie) dauereffekteFuer(kennung string) []Dauereffekt {
	var seine []Dauereffekt

	for _, schicht := range Schichten() {
		for _, effekt := range p.Dauereffekte {
			if effekt.Objekt == kennung && effekt.Schicht == schicht {
				seine = append(seine, effekt)
			}
		}
	}

	return seine
}

// HatFaehigkeit sagt, ob ein Objekt eine Schlüsselwortfähigkeit hat —
// gedruckt oder verliehen.
//
// **Exportiert, weil es die Frage ist, die das Fenster stellt**: was kann
// diese Karte? Der gedruckte Text allein beantwortet sie seit `M10` nicht
// mehr.
func (p Partie) HatFaehigkeit(kennung, wort string) bool {
	if p.hatSchluesselwort(kennung, wort) {
		return true
	}

	for _, effekt := range p.dauereffekteFuer(kennung) {
		if effekt.Schicht == Schicht6 && effekt.Faehigkeit == wort {
			return true
		}
	}

	return false
}

// werteMitEffekten legt die siebte Schicht über gedruckte Werte — CR 613.4.
func (p Partie) werteMitEffekten(
	kennung string, staerke, widerstand Kartenwert,
) (nachher, danach Kartenwert) {
	for _, effekt := range p.dauereffekteFuer(kennung) {
		switch effekt.Schicht {
		// **CR 613.4b**: setzen.
		case Schicht7b:
			staerke.Zahl, widerstand.Zahl = effekt.Staerke, effekt.Widerstand

		// **CR 613.4c**: ändern.
		case Schicht7c:
			staerke.Zahl += effekt.Staerke
			widerstand.Zahl += effekt.Widerstand
		}
	}

	return staerke, widerstand
}

// ohneEffekteAn nimmt die Effekte an einem Objekt weg — CR 400.7.
//
// **Weil es ein neues Objekt ist.** Ein Effekt galt dem alten, und der ist
// mit dem Zonenwechsel vergangen — dieselbe Regel, die Marken und Zustand
// wegräumt.
func (p Partie) ohneEffekteAn(kennung string) Partie {
	uebrig := make([]Dauereffekt, 0, len(p.Dauereffekte))

	for _, effekt := range p.Dauereffekte {
		if effekt.Objekt != kennung {
			uebrig = append(uebrig, effekt)
		}
	}

	p.Dauereffekte = uebrig

	return p
}

// ohneEffekteBisZugende beendet, was mit dem Zug endet — CR 514.2.
//
// **Die zweite Hälfte von CR 514.2**, die `G9` ausdrücklich offen gelassen
// hat: im Aufräumschritt geht nicht nur der markierte Schaden weg, sondern
// auch jeder Effekt „bis zum Ende des Zuges".
func (p Partie) ohneEffekteBisZugende() Partie {
	uebrig := make([]Dauereffekt, 0, len(p.Dauereffekte))

	for _, effekt := range p.Dauereffekte {
		if effekt.Dauer != DauerZugende {
			uebrig = append(uebrig, effekt)
		}
	}

	p.Dauereffekte = uebrig

	return p
}

// istSchicht sagt, ob ein Wert eine der zehn ist.
//
// **Für den Bestand** (§8): eine Partie kann eine Schicht tragen, die eine
// spätere Fassung kannte.
func istSchicht(schicht Schicht) bool {
	return slices.Contains(Schichten(), schicht)
}
