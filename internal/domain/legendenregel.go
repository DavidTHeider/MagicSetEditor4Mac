package domain

// **Die Legendenregel** — Werkstück M12 der Regelmaschine
// (#1614).
//
// # Die einzige der fünfundzwanzig, die wartet
//
// **CR 704.5j** (Fassung 2026-08-19) verlangt eine **Wahl**: beherrscht
// jemand zwei oder mehr gleichnamige legendäre Permanente, sagt er, welches
// bleibt; die übrigen gehen in die Friedhöfe ihrer Eigner. Die anderen
// vierundzwanzig Aktionen aus CR 704.5 geschehen einfach.
//
// **`M5` (#1607)
// hat sie deshalb als Grenze geführt**: „es gibt am Tisch keinen Weg, jemanden
// zu fragen." Das Issue zu `M12` nennt genau das als Auftrag — „wer CR 704.5j
// baut, baut zuerst diesen Weg". Hier ist er.
//
// # Und der Weg ist eine offene Frage im Zustand
//
// **Nicht ein Rückruf und nicht ein Dialog.** Eine Partie wird über vier
// Flächen bedient (§16), und drei davon haben kein Fenster; eine Frage, die
// nur im Fenster gestellt werden kann, wäre an dreien keine. Sie steht
// deshalb **in der Partie** ([Partie.Legendenwahlen]) und wird mit einem
// Befehl beantwortet — dieselbe Gestalt wie der Abwurf im Aufräumschritt
// (CR 514.1).

import "fmt"

// Legendenwahl ist eine offene Frage aus CR 704.5j.
//
// **Sitz und Name genügen.** Welche Objekte gemeint sind, ergibt sich aus
// dem Tisch — und er ändert sich, während die Frage offen ist. Eine Liste
// von Kennungen daneben liefe damit auseinander.
type Legendenwahl struct {
	// Sitz ist, wer wählen soll — der **Beherrscher** (CR 704.5j) und nicht
	// der Eigner.
	Sitz string

	// Name ist der Name, den sie teilen — CR 201.2.
	Name string
}

// legendenpruefen stellt die offenen Fragen aus CR 704.5j — und nimmt die
// zurück, die sich erledigt haben.
//
// **Beides in einer Funktion, und das ist der Punkt.** Eine Frage, deren
// Gegenstand verschwunden ist, wartete sonst für immer: stirbt die zweite
// Legende durch etwas anderes, ist nichts mehr zu wählen. Wer nur stellte
// und nie zurücknähme, hielte die halbe Zusage (§10).
func (p Partie) legendenpruefen() (Partie, bool) {
	faellig := p.legendenfaellig()

	if gleicheWahlen(p.Legendenwahlen, faellig) {
		return p, false
	}

	p.Legendenwahlen = faellig

	return p, true
}

// legendenfaellig gibt die Fragen, die der Tisch gerade stellt.
//
// **Die Reihenfolge ist die der Objekte** und nicht die einer Tabelle: zwei
// Läufe über dieselbe Partie sollen dieselbe Liste geben, sonst wechselte
// die Antwort auf „hat sich etwas geändert?" bei unverändertem Tisch.
func (p Partie) legendenfaellig() []Legendenwahl {
	gezaehlt := map[Legendenwahl]int{}

	var reihenfolge []Legendenwahl

	for _, objekt := range p.Objekte {
		wahl, legendaer := p.legendenname(objekt)
		if !legendaer {
			continue
		}

		gezaehlt[wahl]++

		if gezaehlt[wahl] == 2 {
			reihenfolge = append(reihenfolge, wahl)
		}
	}

	return reihenfolge
}

// legendenname gibt Beherrscher und Namen, wenn dies eine Legende auf dem
// Tisch ist — CR 704.5j nennt alle drei Bedingungen.
//
// **Ein verdecktes Permanent hat keinen Namen** (CR 708.2a), und damit teilt
// es auch keinen: zwei verdeckte Karten sind nicht zwei gleichnamige
// Legenden. Das fällt hier ohne eigenen Zweig heraus, weil
// [VerdeckteEigenschaften] weder Namen noch Supertyp trägt.
func (p Partie) legendenname(objekt Spielobjekt) (Legendenwahl, bool) {
	if objekt.Zone != ZoneTisch {
		return Legendenwahl{}, false
	}

	eigen, da := p.Eigenschaften(objekt)
	if !da || eigen.Name == "" || !eigen.Typzeile.Hat(SuperLegendaer) {
		return Legendenwahl{}, false
	}

	return Legendenwahl{Sitz: objekt.Beherrscher, Name: eigen.Name}, true
}

// gleicheWahlen sagt, ob zwei Listen dieselben Fragen stellen.
func gleicheWahlen(einer, anderer []Legendenwahl) bool {
	if len(einer) != len(anderer) {
		return false
	}

	for stelle := range einer {
		if einer[stelle] != anderer[stelle] {
			return false
		}
	}

	return true
}

// Legende beantwortet eine Frage aus CR 704.5j — das genannte Permanent
// bleibt, die übrigen gehen in die Friedhöfe ihrer Eigner.
//
// **Der Beherrscher wählt** (CR 704.5j), und deshalb steht `wer` daneben:
// wer nicht gefragt ist, antwortet nicht.
func (p Partie) Legende(wer, objekt string) (Partie, error) {
	stelle, fehler := p.objektstelle(objekt)
	if fehler != nil {
		return p, fehler
	}

	wahl, legendaer := p.legendenname(p.Objekte[stelle])

	if !legendaer || wahl.Sitz != wer || !p.wahlSteht(wahl) {
		return p, fmt.Errorf(
			"%w: über %q wählt %q gerade nicht (CR 704.5j)",
			ErrValidation, objekt, wer)
	}

	for andere := range p.Objekte {
		if andere == stelle {
			continue
		}

		if gleiche, ist := p.legendenname(p.Objekte[andere]); ist &&
			gleiche == wahl {
			p = p.inDenFriedhof(andere)
		}
	}

	// **Die beantwortete Frage fällt weg, und keine zweite bleibt liegen.**
	// Gerechnet wird sie neu aus dem Tisch statt aus der Liste gestrichen:
	// wer nur striche, ließe eine Frage stehen, die dieselbe Antwort schon
	// erledigt hat — drei gleichnamige Legenden unter zwei Beherrschern
	// sind zwei Fragen, und eine davon kann durch den Friedhof vergehen.
	p, _ = p.legendenpruefen()

	return p, nil
}

// wahlSteht sagt, ob diese Frage offen ist.
//
// **Gefragt wird die Liste und nicht der Tisch.** Wer über zwei Legenden
// entschiede, bevor die Prüfung sie bemerkt hat, entschiede über eine Frage,
// die niemand gestellt hat — und der Ereignisstrom trüge eine Antwort ohne
// Frage.
func (p Partie) wahlSteht(wahl Legendenwahl) bool {
	for _, offen := range p.Legendenwahlen {
		if offen == wahl {
			return true
		}
	}

	return false
}
