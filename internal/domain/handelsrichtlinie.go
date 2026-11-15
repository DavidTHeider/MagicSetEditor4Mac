package domain

// **Die Handelsrichtlinie ist eine Schranke** — Werkstück `T1` von Phase 5.6
// (#1466).
//
// # Was „Richtlinie" hier heißt, und warum nicht etwas Neues
//
// **Dieses Programm kennt den Begriff schon** — Werkstück `O3` von Phase 5.2
// (#1427),
// [Vermerk 0079](../../docs/entscheidungen/0079-einstellungsebenen-und-portabler-import.md):
//
// > **Eine Richtlinie ist eine Schranke und keine Vorliebe.** Sie steht über
// > allen Ebenen und lässt sich von unten nicht aufheben; was gilt, ist die
// > **Schnittmenge** — der strengste Wert gewinnt, und die Reihenfolge der
// > Ebenen spielt keine Rolle.
//
// **Der Handel bekommt deshalb keine zweite Gestalt daneben**, sondern
// dieselbe: fünf Schlüssel unter `richtlinie.handel`, gefaltet von
// [Verschaerft], eingetragen in beide Reihen, auf der Ebene des **Menschen**.
// Eine Handelsrichtlinie, die sich anders faltete als `richtlinie.nurlokal`,
// wäre ein zweiter Mechanismus für dieselbe Frage — und der erste, der sie
// verwechselt, hebt eine Schranke auf, die nicht ihm gehört.
//
// # Und was sie ausdrücklich nicht entscheidet
//
// **Drei Fragen, drei Stellen** — die Trennung ist der Grund, weshalb es sie
// gibt:
//
//  1. **Darf dieser Peer überhaupt handeln?** Das entscheidet sein Kontakt:
//     [Peerkontakt.Darf] mit [ZweckHandel] verlangt einen **bestätigten**
//     Kontakt **und** das erteilte Recht. Gesperrt bleibt gesperrt — auch
//     nach einem gemeinsamen Spiel (`N1`,
//     #1454).
//  2. **Was bekommt er zu sehen?** Das entscheidet eine Freigabe je Empfänger
//     und Umfang (`T3`,
//     #1468).
//  3. **Was lässt dieses Profil überhaupt zu?** Das ist die Schranke hier —
//     **unabhängig davon, wer fragt**.
//
// # Und wo sie greift
//
// **In `service`, an jedem Weg, auf dem ein Handel entsteht oder der Einigung
// näher kommt** (`service/handelsschranke.go`, #2435) — dort steht, wann
// gefragt wird; hier, was die Antwort ist. Bis #2435 fragte sie niemand, und
// die Einstellungsseite versprach eine Schranke, die nichts schränkte.
//
// **Ohne die dritte wäre jede Grenze eine Eigenschaft des Gegenübers.** „Ich
// handle grundsätzlich keine Stapel über vier Stück" ist aber keine Aussage
// über einen Partner, sondern über einen selbst; wer sie an jedem Kontakt
// einzeln hinterlegte, hätte sie beim nächsten vergessen.

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Die fünf Schranken des Handels.
const (
	// RichtlinieHandel schaltet den Handel als Ganzes.
	//
	// **Ein eigener Schlüssel und nicht vier Schalter auf einmal** —
	// derselbe Satz wie bei [RichtlinieNurLokal]: „kein Handel" bleibt
	// stehen, wenn eine fünfte Vertragsart dazukommt.
	RichtlinieHandel = "richtlinie.handel"

	// RichtlinieHandelMarkt schaltet die Handlung bei einem fremden
	// Anbieter.
	//
	// **Getrennt vom Rest, weil sie auf Geld greift.** Teilen, Deckleihe und
	// Tausch bleiben unter zwei Menschen; der Markt nicht.
	RichtlinieHandelMarkt = "richtlinie.handel.markt"

	// RichtlinieHandelMenge ist die Höchstmenge **je Position** — leer
	// heißt: keine.
	RichtlinieHandelMenge = "richtlinie.handel.menge"

	// RichtlinieHandelWert ist der Höchstwert **eines Angebots** in der
	// kleinsten Währungseinheit — leer heißt: keine. **Nur bei geteilter
	// Bewertung.**
	RichtlinieHandelWert = "richtlinie.handel.wert"

	// RichtlinieHandelFrist ist die längste Frist einer Zusage **in
	// Minuten** — leer heißt: keine.
	//
	// **In Minuten und nicht als Dauer**: eine Schranke muss sich falten
	// lassen, und `kleinsteGrenze` vergleicht Zahlen. `1h30m` ließe sich
	// nicht der Größe nach ordnen, ohne es vorher zu deuten.
	RichtlinieHandelFrist = "richtlinie.handel.frist"
)

// Die zwei Stellungen einer Handelsschranke.
//
// **Dieselben Wörter wie bei [RichtlinieNurLokal]**, weil ein Mensch sie an
// beiden Stellen liest — und eigene Konstanten, weil es zwei Fragen sind.
const (
	HandelSchrankeAus = "aus"
	HandelSchrankeAn  = "an"
)

// Handelsrichtlinie ist, was dieses Profil für jeden Handel voraussetzt.
//
// **Der Nullwert einer Grenze heißt „keine Grenze"** und nicht „null". Eine
// Schranke, die aus einem leeren Feld ein Verbot machte, verböte beim ersten
// Start alles.
type Handelsrichtlinie struct {
	// Handeln sagt, ob überhaupt gehandelt wird.
	Handeln bool

	// Markt sagt, ob die Handlung bei einem fremden Anbieter zugelassen ist.
	Markt bool

	// Mengengrenze ist die Höchstmenge je Position — 0 heißt: ohne Grenze.
	Mengengrenze int

	// Wertgrenze ist der Höchstwert eines Angebots — 0 heißt: ohne Grenze.
	Wertgrenze int64

	// Fristgrenze ist die längste Frist einer Zusage — 0 heißt: ohne Grenze.
	Fristgrenze time.Duration
}

// Standardhandelsrichtlinie ist, was ohne Einstellung gilt.
//
// **Handeln ja, Markt nein.** Der Markt greift auf einen fremden Anbieter und
// auf Geld; die Phase nennt ihn ein „optionales externes Schreibpaket, kein
// stiller Checkout"
// (#1410).
// Eine Vorgabe, die ihn einschlösse, wäre genau der stille Checkout.
func Standardhandelsrichtlinie() Handelsrichtlinie {
	return Handelsrichtlinie{Handeln: true}
}

// HandelsrichtlinieAus liest die Schranken aus den **gefalteten** Werten.
//
// **Gefaltet heißt: [Verschaerft] ist schon gelaufen.** Diese Stelle
// entscheidet nicht, welche Ebene gewinnt — das tut die Faltung, und sie tut
// es für jede Schranke gleich.
//
// **Was fehlt oder keinen Sinn ergibt, ist die Vorgabe** (§8): eine
// Einstellung kommt aus einer Datei, die ein Mensch ändern kann.
func HandelsrichtlinieAus(werte map[string]string) Handelsrichtlinie {
	richtlinie := Standardhandelsrichtlinie()

	if wert, gesetzt := schranke(werte, RichtlinieHandel); gesetzt {
		richtlinie.Handeln = wert
	}

	if wert, gesetzt := schranke(werte, RichtlinieHandelMarkt); gesetzt {
		richtlinie.Markt = wert
	}

	if menge := grenze(werte, RichtlinieHandelMenge); menge > 0 {
		richtlinie.Mengengrenze = int(menge)
	}

	richtlinie.Wertgrenze = grenze(werte, RichtlinieHandelWert)

	if minuten := grenze(werte, RichtlinieHandelFrist); minuten > 0 {
		// Sättigend (J05, #2269): eine Grenze, die keine Dauer fasst, ist
		// die weiteste und nicht eine umgelaufene von zwölf Stunden.
		richtlinie.Fristgrenze = ZeitspanneHoechstens(minuten, time.Minute)
	}

	return richtlinie
}

// schranke liest einen Schalter — der zweite Wert sagt, ob er gesetzt war.
func schranke(werte map[string]string, schluessel string) (an, gesetzt bool) {
	switch strings.TrimSpace(werte[schluessel]) {
	case HandelSchrankeAn:
		return true, true

	case HandelSchrankeAus:
		return false, true
	}

	return false, false
}

// grenze liest eine Obergrenze — was keine ist, ist null.
func grenze(werte map[string]string, schluessel string) int64 {
	zahl, fehler := strconv.ParseInt(
		strings.TrimSpace(werte[schluessel]), 10, 64)
	if fehler != nil || zahl < 0 {
		return 0
	}

	return zahl
}

// Erlaubt sagt, ob dieses Profil dieses Angebot zulässt.
//
// **Sie wiederholt die Prüfung des Angebots nicht** ([Handelsangebot.Pruefe]):
// ein Angebot, das keines ist, ist kein Fall für die Schranke — sonst stünde
// an zwei Stellen dieselbe Ablehnung mit zwei verschiedenen Sätzen.
//
// **[ErrValidation] und nicht [ErrConflict]**: eine Schranke ist keine Frage
// des Standes. Wer sie verletzt, kommt auch beim zehnten Versuch nicht durch
// — er muss das Angebot ändern oder die Schranke.
func (r Handelsrichtlinie) Erlaubt(angebot Handelsangebot) error {
	if fehler := r.ErlaubtArt(angebot.Art); fehler != nil {
		return fehler
	}

	if fehler := r.pruefeMengen(angebot); fehler != nil {
		return fehler
	}

	return r.pruefeWert(angebot)
}

// ErlaubtArt sagt, ob dieses Profil eine Vertragsart überhaupt zulässt —
// **für Teilen und Deckleihe**, die kein Angebot mit Positionen tragen (#2435).
// Was für sie gilt, ist nur: wird gehandelt, und wenn am Markt, ist er offen.
func (r Handelsrichtlinie) ErlaubtArt(art Handelsart) error {
	if !r.Handeln {
		return fmt.Errorf(
			"%w: dieses Profil handelt nicht — %q steht auf %q",
			ErrValidation, RichtlinieHandel, HandelSchrankeAus)
	}

	if art == HandelMarkt && !r.Markt {
		return fmt.Errorf(
			"%w: dieses Profil handelt nicht am Markt — %q steht auf %q",
			ErrValidation, RichtlinieHandelMarkt, HandelSchrankeAus)
	}

	return nil
}

// pruefeMengen hält jede Position gegen die Mengengrenze.
func (r Handelsrichtlinie) pruefeMengen(angebot Handelsangebot) error {
	if r.Mengengrenze == 0 {
		return nil
	}

	for stelle, position := range angebot.Positionen {
		if position.Menge > r.Mengengrenze {
			return fmt.Errorf(
				"%w: die Position %d nennt %d Stück, dieses Profil handelt "+
					"höchstens %d je Position",
				ErrValidation, stelle+1, position.Menge, r.Mengengrenze)
		}
	}

	return nil
}

// pruefeWert hält das ganze Angebot gegen die Wertgrenze.
//
// **Das ganze und nicht je Zeile**: zwei Positionen zu je sechshundert sind
// zusammen zwölfhundert, und eine Grenze von tausend meint die Summe.
func (r Handelsrichtlinie) pruefeWert(angebot Handelsangebot) error {
	if r.Wertgrenze == 0 || !angebot.BewertungGeteilt {
		return nil
	}

	zusammen := int64(0)
	for _, position := range angebot.Positionen {
		zusammen += position.Bewertung
	}

	if zusammen > r.Wertgrenze {
		return fmt.Errorf(
			"%w: das Angebot steht bei %d, dieses Profil handelt höchstens %d",
			ErrValidation, zusammen, r.Wertgrenze)
	}

	return nil
}

// ErlaubtFrist sagt, ob eine Zusage so lange festhalten darf.
func (r Handelsrichtlinie) ErlaubtFrist(frist time.Duration) error {
	if r.Fristgrenze == 0 || frist <= r.Fristgrenze {
		return nil
	}

	return fmt.Errorf(
		"%w: die Zusage hielte %s fest, dieses Profil hält höchstens %s",
		ErrValidation, frist, r.Fristgrenze)
}

// Zusagefrist gibt die Frist, die eine Zusage tragen darf — #2435.
//
// **Ohne Wunsch die Grenze**: kein Weg nennt heute eine Frist, und eine
// Zusage ohne hielte unbegrenzt fest — länger als jede Grenze. Abweisen hieße,
// dass unter einer Grenze gar nichts festgehalten werden kann; gekürzt hält die
// Zusage so lange, wie die Schranke sagt. **Ein Wunsch über der Grenze wird
// abgewiesen** ([Handelsrichtlinie.ErlaubtFrist]): wer ihn nennt, will ihn so.
func (r Handelsrichtlinie) Zusagefrist(jetzt, gewuenscht time.Time) (time.Time, error) {
	if r.Fristgrenze == 0 {
		return gewuenscht, nil
	}

	if gewuenscht.IsZero() {
		return jetzt.Add(r.Fristgrenze), nil
	}

	if fehler := r.ErlaubtFrist(gewuenscht.Sub(jetzt)); fehler != nil {
		return time.Time{}, fehler
	}

	return gewuenscht, nil
}

// handelsschluessel gibt die fünf Schranken des Handels.
func handelsschluessel() []string {
	return []string{
		RichtlinieHandel, RichtlinieHandelMarkt, RichtlinieHandelMenge,
		RichtlinieHandelWert, RichtlinieHandelFrist,
	}
}

// haengeHandelsschrankenAn trägt sie in die Faltung und in beide Reihen ein.
//
// **In beide, weil die Faltung zwei Ebenen braucht** — derselbe Satz wie bei
// [haengeRichtlinienAn]: eine Schranke, die es nur im Profil gibt, ist keine.
func haengeHandelsschrankenAn() {
	schalter := richtlinie{
		Art:     reihung,
		Strenge: []string{HandelSchrankeAus, HandelSchrankeAn},
	}

	richtlinien[RichtlinieHandel] = schalter
	richtlinien[RichtlinieHandelMarkt] = schalter

	for _, grenze := range []string{
		RichtlinieHandelMenge, RichtlinieHandelWert, RichtlinieHandelFrist,
	} {
		richtlinien[grenze] = richtlinie{Art: obergrenze}
	}

	for _, zeile := range []struct {
		schluessel string
		werte      []string
		vorgabe    string
		pruefung   func(string) error
	}{
		{
			RichtlinieHandel,
			[]string{HandelSchrankeAn, HandelSchrankeAus},
			HandelSchrankeAn, nil,
		},
		{
			RichtlinieHandelMarkt,
			[]string{HandelSchrankeAus, HandelSchrankeAn},
			HandelSchrankeAus, nil,
		},
		{RichtlinieHandelMenge, nil, "", pruefeHandelsgrenze},
		{RichtlinieHandelWert, nil, "", pruefeHandelsgrenze},
		{RichtlinieHandelFrist, nil, "", pruefeHandelsgrenze},
	} {
		einstellungen = append(einstellungen, einstellung{
			Schluessel: zeile.schluessel,
			Wirkung:    WirkungSofort,
			Bereich:    BereichRichtlinien,
			Werte:      zeile.werte,
			Vorgabe:    zeile.vorgabe,
			Pruefung:   zeile.pruefung,
		})

		nutzereinstellungen = append(nutzereinstellungen, nutzereinstellung{
			Schluessel: zeile.schluessel,
			Werte:      zeile.werte,
			Vorgabe:    zeile.vorgabe,
			Pruefung:   zeile.pruefung,
		})
	}
}

// pruefeHandelsgrenze prüft eine Obergrenze — leer heißt: keine.
func pruefeHandelsgrenze(wert string) error {
	sauber := strings.TrimSpace(wert)
	if sauber == "" {
		return nil
	}

	zahl, fehler := strconv.ParseInt(sauber, 10, 64)
	if fehler != nil || zahl < 1 {
		return fmt.Errorf(
			"%w: %q ist keine Obergrenze — eine Schranke ist eine ganze Zahl "+
				"ab eins, und leer heißt „keine\"", ErrValidation, wert)
	}

	return nil
}
