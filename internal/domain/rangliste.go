// **Die Rangliste** — Werkstück 2 von Phase 4.7
// (#1294),
// nach Vermerk [0062](../../docs/entscheidungen/0062-was-ein-turnier-hier-ist.md).
//
// # Warum sie gerechnet und nicht abgelegt wird
//
// Vermerk 0062, Entscheidung 1: eine abgelegte Tabelle und der Verlauf, aus
// dem sie folgt, laufen auseinander, sobald jemand ein Ergebnis berichtigt —
// und das ist am Turniertisch der Normalfall.
//
// **Was es kostet**: bei zehn Runden und vierhundert Spielern sind das
// vierhunderttausend Multiplikationen, und das ist nichts.
//
// # Die Reihenfolge der Tiebreaker
//
// **MTR §3.1, wörtlich und in dieser Folge:**
//
//  1. Match points
//  2. Opponents' match-win percentage
//  3. Game-win percentage
//  4. Opponents' game-win percentage
//
// Fassung vom 2026-02-27, abgerufen am 2026-09-01. Die Reihenfolge ist keine
// Einstellung: eine andere machte jede Rangliste unvergleichbar.

package domain

import "sort"

// Rang ist eine Zeile der Rangliste.
type Rang struct {
	TeilnehmerID int64
	Name         string

	// Platz zählt ab 1.
	Platz int

	Bilanz Bilanz

	// Matchpunkte stehen doppelt — in der Bilanz und hier.
	//
	// **Weil sie der erste Tiebreaker sind** und jede Fläche sie zeigt; ein
	// Aufrufer, der `rang.Bilanz.Matchpunkte()` schreiben müsste, hätte die
	// Zeile ohne Not.
	Matchpunkte int

	Matchquote       float64
	Gegnermatchquote float64
	Spielquote       float64
	Gegnerspielquote float64

	// Ausgestiegen wird mitgegeben, damit die Fläche es zeigen kann.
	//
	// Er steht in der Liste, denn seine Quote geht in die Gegnerquote jedes
	// Gegners ein (Appendix C).
	Ausgestiegen bool
}

// Rangliste rechnet die Rangfolge aus dem Verlauf eines Turniers.
func Rangliste(turnier Turnier) []Rang {
	bilanzen, gegner := auswerten(turnier)

	raenge := make([]Rang, 0, len(turnier.Teilnehmer))

	for _, teilnehmer := range turnier.Teilnehmer {
		bilanz := bilanzen[teilnehmer.ID]

		raenge = append(raenge, Rang{
			TeilnehmerID: teilnehmer.ID,
			Name:         teilnehmer.Name,
			Bilanz:       bilanz,
			Matchpunkte:  bilanz.Matchpunkte(),
			Matchquote:   bilanz.Matchquote(),
			Spielquote:   bilanz.Spielquote(),
			Gegnermatchquote: Gegnerquote(
				bilanzenVon(gegner[teilnehmer.ID], bilanzen), Bilanz.Matchquote),
			Gegnerspielquote: Gegnerquote(
				bilanzenVon(gegner[teilnehmer.ID], bilanzen), Bilanz.Spielquote),
			Ausgestiegen: teilnehmer.Ausgestiegen,
		})
	}

	ordne(raenge, anmeldereihenfolge(turnier))

	for stelle := range raenge {
		raenge[stelle].Platz = stelle + 1
	}

	return raenge
}

// auswerten zählt den Verlauf zu Bilanzen und Gegnerlisten aus.
//
// **Eine offene Paarung zählt nicht.** Wer noch nicht gespielt hat, hat keine
// Niederlage — eine laufende Runde darf die Rangliste nicht verfälschen.
func auswerten(turnier Turnier) (bilanzen map[int64]Bilanz, gegner map[int64][]int64) {
	bilanzen = make(map[int64]Bilanz, len(turnier.Teilnehmer))
	gegner = make(map[int64][]int64, len(turnier.Teilnehmer))

	for _, teilnehmer := range turnier.Teilnehmer {
		bilanzen[teilnehmer.ID] = Bilanz{}
	}

	for _, runde := range turnier.Runden {
		for _, paarung := range runde.Paarungen {
			if paarung.IstBye() {
				// **Ein Bye ist ein Ergebnis ohne Ergebnis**: er gilt, sobald
				// er gesetzt ist, denn es gibt nichts zu spielen.
				eintragen(bilanzen, paarung.A, func(b *Bilanz) { b.Bye() })

				continue
			}

			if paarung.Ergebnis == nil {
				continue
			}

			stand := *paarung.Ergebnis

			eintragen(bilanzen, paarung.A, func(b *Bilanz) {
				b.Spiele(stand.SpieleA, stand.SpieleB, stand.Unentschieden)
			})
			eintragen(bilanzen, paarung.B, func(b *Bilanz) {
				b.Spiele(stand.SpieleB, stand.SpieleA, stand.Unentschieden)
			})

			gegner[paarung.A] = append(gegner[paarung.A], paarung.B)
			gegner[paarung.B] = append(gegner[paarung.B], paarung.A)
		}
	}

	return bilanzen, gegner
}

// eintragen ändert die Bilanz eines Teilnehmers, den es gibt.
//
// **Ein unbekannter Teilnehmer wird übergangen und nicht angelegt.** Eine
// Paarung, die auf jemanden zeigt, den die Teilnehmerliste nicht führt, ist
// ein Fehler im Bestand; ihn hier stillschweigend zu einem Rang zu machen,
// verdeckte ihn.
func eintragen(bilanzen map[int64]Bilanz, wer int64, tue func(*Bilanz)) {
	bilanz, gibts := bilanzen[wer]
	if !gibts {
		return
	}

	tue(&bilanz)

	bilanzen[wer] = bilanz
}

// bilanzenVon sammelt die Bilanzen einer Gegnerliste.
func bilanzenVon(wer []int64, bilanzen map[int64]Bilanz) []Bilanz {
	gesammelt := make([]Bilanz, 0, len(wer))

	for _, einer := range wer {
		if bilanz, gibts := bilanzen[einer]; gibts {
			gesammelt = append(gesammelt, bilanz)
		}
	}

	return gesammelt
}

// anmeldereihenfolge gibt die Anmeldereihenfolge je Teilnehmer.
//
// **Nicht `stellen`** — den Namen trägt schon `fassung.go` für etwas anderes.
func anmeldereihenfolge(turnier Turnier) map[int64]int {
	stelle := make(map[int64]int, len(turnier.Teilnehmer))

	for nummer, teilnehmer := range turnier.Teilnehmer {
		if teilnehmer.Stelle != 0 {
			stelle[teilnehmer.ID] = teilnehmer.Stelle

			continue
		}

		stelle[teilnehmer.ID] = nummer + 1
	}

	return stelle
}

// ordne sortiert nach den vier Tiebreakern aus MTR §3.1.
//
// **Und danach nach der Anmeldung.** Die MTR sagt nichts über den Fall, in
// dem alle vier gleich sind (Vermerk 0062, Entscheidung 8 sinngemäß); ohne
// eine Regel gäben zwei Aufrufe zwei Ranglisten.
func ordne(raenge []Rang, stelle map[int64]int) {
	sort.SliceStable(raenge, func(a, b int) bool {
		einer, anderer := raenge[a], raenge[b]

		for _, vergleich := range []struct{ links, rechts float64 }{
			{float64(einer.Matchpunkte), float64(anderer.Matchpunkte)},
			{einer.Gegnermatchquote, anderer.Gegnermatchquote},
			{einer.Spielquote, anderer.Spielquote},
			{einer.Gegnerspielquote, anderer.Gegnerspielquote},
		} {
			if vergleich.links != vergleich.rechts {
				return vergleich.links > vergleich.rechts
			}
		}

		return stelle[einer.TeilnehmerID] < stelle[anderer.TeilnehmerID]
	})
}
