package domain

// **Der Beginn einer Partie** — Mischen, Starthand und Mulligan, CR 103.3
// und CR 103.5 (Fassung 2026-09-25, auch für jede weitere Nummer hier),
// #1839.
//
// # Warum als Befehle und nicht im Aufbau
//
// **Weil der Aufbau so entschieden ist**: `Partieaufbau.Aufbauen` legt alles
// in die Bibliothek, ohne Mischen und ohne Hand — „das sind Befehle, und
// Befehle stehen im Strom". Eine Partie, die schon gemischt anfinge, hätte
// einen Zustand, den kein Ereignis erklärt. Hier stehen die Befehle; der
// Zufall kommt von außen, wie beim Mischen einer Zone.
//
// # Was hier nicht geschieht
//
// **Die Reihenfolge der Erklärungen** — CR 103.5 lässt den Startspieler
// zuerst sagen, ob er einen Mulligan nimmt, dann die anderen, und nimmt alle
// zugleich. Hier nimmt jeder Sitz seinen für sich: im Sandkasten sitzt oft
// ein Mensch an allen Plätzen, und über das Netz sieht keiner die Hand des
// anderen, bevor er sich entscheidet. **Und der freie Mulligan als
// Hausregel**: die Partie führt heute keine vereinbarten Hausregeln, nur die
// erlaubten im Profil (`HausregelFreierMulligan`). **Und der freie erste
// Mulligan im Brawl** (CR 103.5c): die Partie kennt ihr Format nicht, nur
// ihre Sitze — frei ist er deshalb nur, wo mehr als zwei sitzen.
//
// **Und der erste Zug wartet auf den Beginn** — CR 103.8. Der Aufbau
// beginnt ihn zwar sofort (`Partie.ZugBeginnen`), aber solange
// [Partie.BeginnLaeuft], nimmt der Sandkasten nur die Schritte des Beginns
// an und, was ihn nicht berührt — **vom Aufbau an**, seit #1851; bis dahin
// erst ab der ersten gezogenen Starthand (Befund 2 des Zweitblicks über
// Phase 6.9).
//
// **Was dabei fehlt, ist CR 103.6**: Handlungen mit Karten aus der
// Starthand — etwa eine Karte, mit der jemand die Partie auf dem Schlachtfeld
// beginnt — haben keinen eigenen Schritt. Wer eine solche hält, legt sie im
// ersten Zug, sobald er Priorität hat.

import (
	"fmt"
	"slices"
)

// Eroeffnungsstand sagt, wo ein Sitz im Beginn der Partie steht.
//
// **Leer heißt: eine Partie von vorher.** Sie ist aufgebaut worden, bevor es
// diese Befehle gab, und läuft vielleicht im fünften Zug — eine Starthand
// dort mischte die Bibliothek und legte sieben Karten dazu. Deshalb setzt
// der Aufbau [EroeffnungAusstehend], und nur dann beginnt etwas.
type Eroeffnungsstand string

const (
	// EroeffnungAusstehend — der Beginn steht bevor, eine Starthand gibt es
	// noch nicht.
	EroeffnungAusstehend Eroeffnungsstand = "ausstehend"

	// EroeffnungOffen — die Starthand liegt da und ist noch nicht behalten.
	EroeffnungOffen Eroeffnungsstand = "offen"

	// EroeffnungBehalten — dieser Sitz nimmt keinen Mulligan mehr (CR 103.5).
	EroeffnungBehalten Eroeffnungsstand = "behalten"
)

// ZieheStarthand mischt die Bibliothek und zieht die Starthand — CR 103.3,
// dann CR 103.5.
//
// **Eine zu kleine Bibliothek gibt, was sie hat, und der Rest ist ein Griff
// ins Leere**: die Starthand wird *gezogen*, Karte für Karte (CR 121.2), und
// wer mehr ziehen muss, als daliegt, verliert, sobald jemand Priorität
// bekäme (CR 104.3c, CR 704.5b). Vermerkt wird es wie jedes andere Ziehen.
func (p Partie) ZieheStarthand(wer string, mischung []int) (Partie, error) {
	stelle, fehler := p.eroeffnungssitz(wer)
	if fehler != nil {
		return p, fehler
	}

	if p.Sitze[stelle].Eroeffnung != EroeffnungAusstehend {
		return p, fmt.Errorf(
			"%w: für %q steht keine Starthand aus (CR 103.5)", ErrValidation, wer)
	}

	danach, fehler := p.mischeBibliothek(wer, mischung)
	if fehler != nil {
		return p, fehler
	}

	return danach.zieheStarthandkarten(wer).mitSitz(stelle, func(sitz *Sitz) {
		sitz.Eroeffnung = EroeffnungOffen
	}), nil
}

// NimmMulligan mischt die Hand zurück und zieht eine neue — CR 103.5.
//
// **Wie viele Karten danach unter die Bibliothek gehen**, zählt
// [gezaehlteMulligans]; im Mehrspielerspiel ist der erste frei (CR 103.5c).
func (p Partie) NimmMulligan(wer string, mischung []int) (Partie, error) {
	stelle, fehler := p.eroeffnungssitz(wer)
	if fehler != nil {
		return p, fehler
	}

	sitz := p.Sitze[stelle]

	if verwehrt := p.mulliganVerwehrt(sitz); verwehrt != nil {
		return p, verwehrt
	}

	danach := p

	for _, ort := range p.handVon(wer) {
		danach = danach.zonenwechsel(ort, ZoneBibliothek)
	}

	danach, fehler = danach.mischeBibliothek(wer, mischung)
	if fehler != nil {
		return p, fehler
	}

	mulligans := sitz.Mulligans + 1
	danach = danach.zieheStarthandkarten(wer)

	// **Mehr, als auf der Hand liegt, geht nicht unter die Bibliothek** —
	// CR 103.5 sagt dazu nichts, CR 101.3 schon: was sich nicht ausführen
	// lässt, entfällt. Ohne die Grenze säße ein Sitz mit kleiner Bibliothek
	// fest, denn behalten darf er erst, wenn nichts mehr geschuldet ist.
	geschuldet := min(danach.gezaehlteMulligans(mulligans),
		len(danach.handVon(wer)))

	return danach.mitSitz(stelle, func(sitz *Sitz) {
		sitz.Mulligans = mulligans
		sitz.Unterlegen = geschuldet
	}), nil
}

// Unterlege legt die genannten Handkarten unter die Bibliothek, in der
// genannten Reihenfolge — CR 103.5.
//
// **Genau so viele, wie geschuldet sind**, und nur aus der eigenen Hand: wer
// eine Karte aus der Bibliothek nennte, legte sie unter sich selbst.
func (p Partie) Unterlege(wer string, objekte []string) (Partie, error) {
	stelle, fehler := p.eroeffnungssitz(wer)
	if fehler != nil {
		return p, fehler
	}

	geschuldet := p.Sitze[stelle].Unterlegen

	if geschuldet == 0 || len(objekte) != geschuldet {
		return p, fmt.Errorf(
			"%w: %q schuldet %d Karten, genannt sind %d (CR 103.5)",
			ErrValidation, wer, geschuldet, len(objekte))
	}

	danach := p

	// **Gefragt wird der Stand nach den vorigen Karten** — deshalb ist eine
	// zweimal genannte Karte beim zweiten Mal keine Handkarte mehr.
	for _, kennung := range objekte {
		ort := danach.stelleVon(kennung)

		if ort < 0 || danach.Objekte[ort].Beherrscher != wer ||
			danach.Objekte[ort].Zone != ZoneHand {
			return p, fmt.Errorf(
				"%w: %q liegt nicht auf der Hand von %q", ErrValidation,
				kennung, wer)
		}

		danach = danach.zonenwechsel(ort, ZoneBibliothek)
		danach.Objekte[ort].Ordnung = danach.untersteStelle(wer) + 1
	}

	return danach.mitSitz(stelle, func(sitz *Sitz) {
		sitz.Unterlegen = 0
	}), nil
}

// Behalte beendet die Mulligans dieses Sitzes — CR 103.5.
func (p Partie) Behalte(wer string) (Partie, error) {
	stelle, fehler := p.eroeffnungssitz(wer)
	if fehler != nil {
		return p, fehler
	}

	if verwehrt := behaltenVerwehrt(p.Sitze[stelle]); verwehrt != nil {
		return p, verwehrt
	}

	return p.mitSitz(stelle, func(sitz *Sitz) {
		sitz.Eroeffnung = EroeffnungBehalten
	}), nil
}

// Eroeffnungsschritte sagt, welche Schritte des Beginns einem Sitz
// offenstehen.
type Eroeffnungsschritte struct {
	Starthand  bool
	Mulligan   bool
	Unterlegen bool
	Behalten   bool
}

// Eroeffnungsschritte gibt, was dieser Sitz gerade tun kann — **dieselben
// Bedingungen, mit denen die Befehle danach prüfen**, und nicht eine zweite
// Fassung von ihnen. Wer keinen Sitz hat, kann nichts.
func (p Partie) Eroeffnungsschritte(wer string) Eroeffnungsschritte {
	stelle, fehler := p.eroeffnungssitz(wer)
	if fehler != nil {
		return Eroeffnungsschritte{}
	}

	sitz := p.Sitze[stelle]

	return Eroeffnungsschritte{
		Starthand:  sitz.Eroeffnung == EroeffnungAusstehend,
		Mulligan:   p.mulliganVerwehrt(sitz) == nil,
		Unterlegen: sitz.Unterlegen > 0,
		Behalten:   behaltenVerwehrt(sitz) == nil,
	}
}

// mulliganVerwehrt sagt, warum dieser Sitz keinen Mulligan nehmen kann —
// oder nil.
func (p Partie) mulliganVerwehrt(sitz Sitz) error {
	if fehler := behaltenVerwehrt(sitz); fehler != nil {
		return fehler
	}

	// **Bis die Starthand null Karten hätte, und dann nicht mehr** —
	// CR 103.5.
	if p.gezaehlteMulligans(sitz.Mulligans) >= Starthand {
		return fmt.Errorf(
			"%w: %q hat keine Karte mehr, die ein Mulligan gäbe (CR 103.5)",
			ErrValidation, sitz.Kennung)
	}

	return nil
}

// behaltenVerwehrt sagt, warum dieser Sitz seine Hand nicht behalten kann —
// oder nil. **Solange unterzulegen ist, gibt es weder Behalten noch
// Mulligan**: die Hand ist erst fertig, wenn die Karten unten liegen.
func behaltenVerwehrt(sitz Sitz) error {
	if sitz.Eroeffnung != EroeffnungOffen {
		return fmt.Errorf("%w: %q hat keine offene Starthand (CR 103.5)",
			ErrValidation, sitz.Kennung)
	}

	if sitz.Unterlegen > 0 {
		return fmt.Errorf(
			"%w: %q muss erst %d Karten unter die Bibliothek legen (CR 103.5)",
			ErrValidation, sitz.Kennung, sitz.Unterlegen)
	}

	return nil
}

// gezaehlteMulligans gibt, wie viele Mulligans für das Unterlegen zählen —
// im Mehrspielerspiel ist der erste frei (CR 103.5c).
func (p Partie) gezaehlteMulligans(mulligans int) int {
	if len(p.Sitze) > 2 && mulligans > 0 {
		return mulligans - 1
	}

	return mulligans
}

// eroeffnungssitz gibt die Stelle eines Sitzes — oder die Abweisung.
func (p Partie) eroeffnungssitz(wer string) (int, error) {
	for stelle, sitz := range p.Sitze {
		if sitz.Kennung == wer {
			return stelle, nil
		}
	}

	return -1, fmt.Errorf("%w: an diesem Tisch sitzt kein %q",
		ErrValidation, wer)
}

// mitSitz gibt die Partie mit einem geänderten Sitz — auf einer Kopie der
// Sitze, damit die Eingabe bleibt, wie sie war.
func (p Partie) mitSitz(stelle int, aendern func(sitz *Sitz)) Partie {
	sitze := make([]Sitz, len(p.Sitze))
	copy(sitze, p.Sitze)
	aendern(&sitze[stelle])
	p.Sitze = sitze

	return p
}

// stelleVon gibt die Stelle eines Objekts — oder -1.
func (p Partie) stelleVon(kennung string) int {
	for stelle, objekt := range p.Objekte {
		if objekt.Kennung == kennung {
			return stelle
		}
	}

	return -1
}

// handVon gibt die Stellen der Hand eines Sitzes.
//
// **Die Stellen bleiben über einen Zonenwechsel gleich** — er ersetzt das
// Objekt an seiner Stelle, statt es zu verschieben.
func (p Partie) handVon(wer string) []int {
	hand := []int{}

	for stelle, objekt := range p.Objekte {
		if objekt.Beherrscher == wer && objekt.Zone == ZoneHand {
			hand = append(hand, stelle)
		}
	}

	return hand
}

// bibliothekVon gibt die Stellen der Bibliothek eines Sitzes, von oben.
func (p Partie) bibliothekVon(wer string) []int {
	stapel := []int{}

	for stelle, objekt := range p.Objekte {
		if objekt.Beherrscher == wer && objekt.Zone == ZoneBibliothek {
			stapel = append(stapel, stelle)
		}
	}

	slices.SortStableFunc(stapel, func(eine, andere int) int {
		return p.Objekte[eine].Ordnung - p.Objekte[andere].Ordnung
	})

	return stapel
}

// untersteStelle gibt die größte Ordnung der Bibliothek — unten liegt, was
// die größte trägt.
func (p Partie) untersteStelle(wer string) int {
	unterste := 0

	for _, stelle := range p.bibliothekVon(wer) {
		unterste = max(unterste, p.Objekte[stelle].Ordnung)
	}

	return unterste
}

// mischeBibliothek ordnet die Bibliothek eines Sitzes nach der Mischung —
// `mischung[i] = j` heißt: an Stelle i+1 liegt danach, was vorher an Stelle
// j+1 lag, wie beim Mischen einer Zone.
func (p Partie) mischeBibliothek(wer string, mischung []int) (Partie, error) {
	stapel := p.bibliothekVon(wer)

	if len(mischung) != len(stapel) {
		return p, fmt.Errorf(
			"%w: eine Mischung über %d Karten für eine Bibliothek aus %d",
			ErrValidation, len(mischung), len(stapel))
	}

	gesehen := make([]bool, len(stapel))

	for _, woher := range mischung {
		if woher < 0 || woher >= len(stapel) || gesehen[woher] {
			return p, fmt.Errorf("%w: die Mischung ist keine Umordnung",
				ErrValidation)
		}

		gesehen[woher] = true
	}

	objekte := make([]Spielobjekt, len(p.Objekte))
	copy(objekte, p.Objekte)

	for wohin, woher := range mischung {
		objekte[stapel[woher]].Ordnung = wohin + 1
	}

	p.Objekte = objekte

	// **Gemischt ist gemischt**: wer vorher hineingesehen hat, schuldet
	// danach nichts mehr — dieselbe Pflicht, die `zone.mischen` löst.
	p.LoeseMischpflicht(wer, ZoneBibliothek)

	return p, nil
}

// zieheStarthandkarten zieht die Starthand, eine Karte nach der anderen —
// CR 121.2, über denselben Weg wie jedes Ziehen.
func (p Partie) zieheStarthandkarten(wer string) Partie {
	for range Starthand {
		p = p.ziehtEine(wer)
	}

	return p
}
