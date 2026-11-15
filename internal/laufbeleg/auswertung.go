package laufbeleg

import (
	"fmt"
	"sort"
)

// Fall ist ein erwarteter Lauf — ein Test in einem Paket.
type Fall struct {
	Paket string `json:"paket"`
	Test  string `json:"test"`
}

// Art sagt, was von einem Fall erwartet wird.
type Art string

// Die vier Arten.
const (
	Pflicht  Art = "pflicht"
	Trend    Art = "trend"
	Optional Art = "optional"
	Weitere  Art = "weitere"
)

// Erwartung sagt, welche Fälle ein Lauf haben muss.
//
// **Trend und Optional nennen je Fall einen Grund** — eine Ausnahme ohne
// Grund ist keine; das prüft, wer die Erwartung liest.
type Erwartung struct {
	Faelle   []Fall
	Trend    map[Fall]string
	Optional map[Fall]string
}

// art sagt, was von einem Fall erwartet wird.
func (e Erwartung) art(einer Fall) Art {
	if _, ist := e.Trend[einer]; ist {
		return Trend
	}

	if _, ist := e.Optional[einer]; ist {
		return Optional
	}

	return Pflicht
}

// Fallergebnis ist, was ein Fall gezeigt hat.
type Fallergebnis struct {
	Fall
	Art         Art     `json:"art"`
	Ausfuehrung Status  `json:"ausfuehrung"`
	Ergebnis    Status  `json:"ergebnis"`
	Dauer       float64 `json:"dauer"`
	Grund       string  `json:"grund,omitempty"`
}

// Zaehlung zählt die Fälle, die niemand erwartet hat.
type Zaehlung struct {
	Gelaufen      int `json:"gelaufen"`
	Bestanden     int `json:"bestanden"`
	Gescheitert   int `json:"gescheitert"`
	Uebersprungen int `json:"uebersprungen"`
}

// Auswertung ist, was ein Strom über eine Erwartung sagt.
type Auswertung struct {
	Status     Status
	Grund      string
	Faelle     []Fallergebnis
	Weitere    Zaehlung
	Kennzahlen map[string]string
	Fehler     []string
}

// Auswerten liest einen Strom gegen eine Erwartung.
//
// `rueckgabe` ist der Rückgabewert von `go test`, und er zählt mit: **ein
// Prozess, der mit einem Fehler endet, den kein Fall erklärt, ist ein
// gescheiterter Lauf** — genau der geschluckte Fehler aus QA-01.AK-03.
func Auswerten(ereignisse []Ereignis, erwartung Erwartung, rueckgabe int) Auswertung {
	gesammelt := sammle(ereignisse)
	auswertung := Auswertung{
		Status:     Bestanden,
		Faelle:     []Fallergebnis{},
		Kennzahlen: gesammelt.kennzahlen,
		Fehler:     gesammelt.fehler,
	}

	erwartet := map[Fall]bool{}
	erklaert := false

	for _, einer := range erwartung.Faelle {
		erwartet[einer] = true
		ergebnis := gesammelt.bewerte(einer, erwartung.art(einer))
		auswertung.Faelle = append(auswertung.Faelle, ergebnis)
		auswertung.stufe(beitrag(ergebnis), ergebnis)
		erklaert = erklaert || ergebnis.Ergebnis == Gescheitert
	}

	erklaert = auswertung.weitere(gesammelt, erwartet) || erklaert
	erklaert = auswertung.baufehler(gesammelt, erwartet) || erklaert
	auswertung.paketeOhneGrund(gesammelt)
	auswertung.rueckgabe(rueckgabe, erklaert)

	if len(gesammelt.fehler) > 0 {
		auswertung.stufe(Gescheitert, Fallergebnis{Grund: gesammelt.fehler[0]})
	}

	// **Null Messfälle sind kein grüner Lauf** (AK-02) — auch dann nicht,
	// wenn niemand welche erwartet hat: dann ist die Erwartung das Problem.
	if len(erwartung.Faelle) == 0 {
		auswertung.stufe(NichtGelaufen, Fallergebnis{Grund: "keine Messfälle erwartet"})
	}

	return auswertung
}

// stufe hebt den Stand, wenn der neue schwerer ist — und nimmt seinen Grund.
func (a *Auswertung) stufe(neu Status, ergebnis Fallergebnis) {
	if neu.rang() <= a.Status.rang() {
		return
	}

	a.Status = neu
	a.Grund = ergebnis.Grund

	if ergebnis.Test != "" {
		a.Grund = fmt.Sprintf("%s.%s: %s — %s", ergebnis.Paket, ergebnis.Test,
			neu, ergebnis.Grund)
	}
}

// beitrag sagt, was ein Fall zum Stand des Laufs beiträgt.
//
// **Ein Trend zählt mit seiner Ausführung und nicht mit seinem Ergebnis**:
// er darf rot sein, aber er muss gemessen haben (AK-03). **Ein optionaler
// Fall darf sich abmelden** — rot werden darf er nicht.
func beitrag(ergebnis Fallergebnis) Status {
	switch {
	case ergebnis.Ausfuehrung != Bestanden &&
		(ergebnis.Art != Optional || ergebnis.Ausfuehrung != Uebersprungen):
		return ergebnis.Ausfuehrung
	case ergebnis.Art == Trend:
		return Bestanden
	case ergebnis.Art == Optional && ergebnis.Ausfuehrung == Uebersprungen:
		return Bestanden
	default:
		return ergebnis.Ergebnis
	}
}

// weitere zählt die Fälle, die niemand erwartet hat — und nimmt die roten
// in die Auswertung auf. Sie gibt zurück, ob einer davon rot war.
func (a *Auswertung) weitere(gesammelt *sammlung, erwartet map[Fall]bool) bool {
	rot := []Fallergebnis{}

	for _, einer := range gesammelt.reihe {
		if erwartet[einer] || !gesammelt.tests[einer].gestartet {
			continue
		}

		ergebnis := gesammelt.bewerte(einer, Weitere)
		a.Weitere.Gelaufen++

		switch ergebnis.Ergebnis {
		case Bestanden:
			a.Weitere.Bestanden++
		case Uebersprungen:
			a.Weitere.Uebersprungen++
		default:
			a.Weitere.Gescheitert++

			rot = append(rot, ergebnis)
		}
	}

	sort.Slice(rot, func(i, j int) bool {
		return rot[i].Paket+"."+rot[i].Test < rot[j].Paket+"."+rot[j].Test
	})

	for _, ergebnis := range rot {
		a.Faelle = append(a.Faelle, ergebnis)
		a.stufe(Gescheitert, ergebnis)
	}

	return len(rot) > 0
}

// baufehler nennt ein Paket, das nicht übersetzt wurde und in dem kein
// erwarteter Fall steht — sonst fiele es durch jede Masche.
func (a *Auswertung) baufehler(gesammelt *sammlung, erwartet map[Fall]bool) bool {
	betroffen := map[string]bool{}
	for einer := range erwartet {
		betroffen[einer.Paket] = true
	}

	pfade := make([]string, 0, len(gesammelt.pakete))
	for pfad := range gesammelt.pakete {
		pfade = append(pfade, pfad)
	}

	sort.Strings(pfade)

	gefunden := false

	for _, pfad := range pfade {
		lauf := gesammelt.pakete[pfad]
		if lauf.baufehler == "" {
			continue
		}

		gefunden = true

		if !betroffen[pfad] {
			a.stufe(Gescheitert, Fallergebnis{
				Grund: pfad + ": Übersetzung gescheitert: " + lauf.baufehler,
			})
		}
	}

	return gefunden
}

// paketeOhneGrund nennt ein Paket, das rot endete, ohne dass ein Fall darin
// rot war oder es nicht übersetzt wurde.
//
// **Der Rückgabewert allein reicht dafür nicht**: ein roter Trend in einem
// anderen Paket erklärt ihn schon, und ein Fehler nach den Läufen — `go test`
// meldet `FAIL` ohne roten Fall — ginge dann still durch (AK-03).
func (a *Auswertung) paketeOhneGrund(gesammelt *sammlung) {
	rot := map[string]bool{}

	for _, ergebnis := range a.Faelle {
		if ergebnis.Ergebnis == Gescheitert {
			rot[ergebnis.Paket] = true
		}
	}

	pfade := make([]string, 0, len(gesammelt.pakete))
	for pfad := range gesammelt.pakete {
		pfade = append(pfade, pfad)
	}

	sort.Strings(pfade)

	for _, pfad := range pfade {
		lauf := gesammelt.pakete[pfad]
		if lauf.ende == "fail" && lauf.baufehler == "" && !rot[pfad] {
			a.stufe(Gescheitert, Fallergebnis{
				Grund: pfad + ": das Paket endete rot, und kein Fall darin erklärt es",
			})
		}
	}
}

// rueckgabe liest den Rückgabewert von `go test`.
func (a *Auswertung) rueckgabe(rueckgabe int, erklaert bool) {
	switch {
	case rueckgabe < 0:
		a.stufe(Gescheitert, Fallergebnis{Grund: "der Prozess wurde abgebrochen"})
	case rueckgabe != 0 && !erklaert:
		a.stufe(Gescheitert, Fallergebnis{Grund: fmt.Sprintf(
			"der Prozess endete mit %d, und kein Fall erklärt es", rueckgabe)})
	}
}
