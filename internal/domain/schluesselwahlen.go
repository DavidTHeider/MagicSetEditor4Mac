package domain

// **Die zwei Handlungen, die eine Wahl verlangen** — Werkstück M7 der
// Regelmaschine
// (#1609).
//
// # Hellsehen und Überwachen
//
// **CR 701.22a**: die obersten N ansehen, beliebig viele nach unten legen,
// den Rest nach oben. **CR 701.25a**: dasselbe, nur gehen die einen in den
// **Friedhof** statt nach unten.
//
// # Warum die Wahl als Nutzlast kommt und nicht als Rückfrage
//
// **Weil es am Tisch keinen Weg gibt, jemanden zu fragen.** Derselbe Grund
// wie bei der Legendenregel (CR 704.5j) und beim Hybridsymbol — nur ist die
// Wahl hier **ansagbar**: wer hellsieht, weiß, was er gesehen hat, und sagt
// es in einem Zug. Eine Rückfrage wäre ein zweiter Befehl, zwischen dem
// jemand anderes handeln könnte (CR 117.1) — und genau das verbieten
// CR 701.22a und CR 701.25a, indem sie eine **Handlung** beschreiben.
//
// **Was das kostet**: wer die Karten nicht kennt, kann nicht wählen. Deshalb
// gibt [Partie.Obenauf] sie zurück — an den, der die Zone ohnehin ansehen
// darf.

import "fmt"

// Obenauf gibt die Kennungen der obersten Karten einer Bibliothek — CR 701.22a.
//
// **Der Blick vor der Wahl.** Wer hellsieht, sieht zuerst; ohne diesen Weg
// müsste er raten, was er wegsortiert.
//
// **Und er ist derselbe Blick wie `zone.blick`** (G1): er erzeugt keine
// Mischpflicht, weil CR 701.22a das Ansehen ausdrücklich zur Handlung zählt.
func (p Partie) Obenauf(wer string, wie int) []string {
	stellen := p.oberstenStellen(wer, wie)

	oben := make([]string, 0, len(stellen))
	for _, stelle := range stellen {
		oben = append(oben, p.Objekte[stelle].Kennung)
	}

	return oben
}

// oberstenStellen gibt die Stellen der obersten Karten — von oben nach unten.
//
// **Stellen und keine Kennungen**: wer danach umsortiert, greift damit direkt
// zu, statt jede Kennung noch einmal zu suchen — und eine Suche, die ins
// Leere liefe, gäbe es gar nicht erst (§10).
func (p Partie) oberstenStellen(wer string, wie int) []int {
	objekte := make([]Spielobjekt, len(p.Objekte))
	copy(objekte, p.Objekte)

	stellen := make([]int, 0, max(wie, 0))

	for range wie {
		stelle := obersteDerBibliothek(objekte, wer)
		if stelle < 0 {
			break
		}

		stellen = append(stellen, stelle)

		// **Aus der Betrachtung nehmen, nicht aus der Partie** — sonst gäbe
		// jeder Durchlauf dieselbe Karte.
		objekte[stelle].Zone = ""
	}

	return stellen
}

// Hellsehen sortiert die obersten Karten neu — CR 701.22a.
//
// **`nachUnten` sind die, die nach unten gehen**, in der angegebenen
// Reihenfolge; alle übrigen der obersten N bleiben oben, in der Reihenfolge,
// in der sie dastanden. CR 701.22a lässt beide Reihenfolgen frei; dieses
// Programm ordnet die obenbleibenden **wie vorher** — eine Grenze, und sie
// steht in der Matrix.
func (p Partie) Hellsehen(wer string, wie int, nachUnten []string) (Partie, error) {
	stellen, fehler := p.gewaehlteObenauf(wer, wie, nachUnten)
	if fehler != nil {
		return p, fehler
	}

	return p.umsortieren(wer, stellen, nachUnten, ZoneBibliothek), nil
}

// Ueberwachen legt die Gewählten in den Friedhof — CR 701.25a.
func (p Partie) Ueberwachen(wer string, wie int, inDenFriedhof []string) (Partie, error) {
	stellen, fehler := p.gewaehlteObenauf(wer, wie, inDenFriedhof)
	if fehler != nil {
		return p, fehler
	}

	return p.umsortieren(wer, stellen, inDenFriedhof, ZoneFriedhof), nil
}

// gewaehlteObenauf prüft, dass jede gewählte Karte wirklich obenauf liegt.
//
// **Sonst wäre die Wahl ein Griff in die ganze Bibliothek** — und die ist
// auch für ihren Eigner verdeckt (G1).
func (p Partie) gewaehlteObenauf(
	wer string, wie int, gewaehlt []string,
) ([]int, error) {
	if wie < 0 {
		return nil, fmt.Errorf("%w: %d Karten ansehen geht nicht",
			ErrValidation, wie)
	}

	if !p.sitzSteht(wer) {
		return nil, fmt.Errorf(
			"%w: den Sitz %q gibt es in dieser Partie nicht", ErrNotFound, wer)
	}

	stellen := p.oberstenStellen(wer, wie)

	for _, eine := range gewaehlt {
		obenauf := false

		for _, stelle := range stellen {
			if p.Objekte[stelle].Kennung == eine {
				obenauf = true
			}
		}

		if !obenauf {
			return nil, fmt.Errorf(
				"%w: %q liegt nicht unter den obersten %d Karten von %q "+
					"(CR 701.22a)", ErrValidation, eine, wie, wer)
		}
	}

	return stellen, nil
}

// umsortieren legt die Gewählten in ihre Zone und den Rest wieder obenauf.
//
// # Warum die Ordnung neu vergeben wird und nicht verschoben
//
// **Weil `Ordnung` in diesem Kern eine Reihenfolge ist und kein Platz.** Die
// obersten N bekommen aufsteigende Zahlen ab null; alles darunter rückt um
// die Zahl der Weggenommenen nach oben. Wer stattdessen Plätze tauschte,
// hätte zwei Karten auf derselben Zahl, sobald eine wegfällt.
func (p Partie) umsortieren(
	wer string, stellen []int, gewaehlt []string, wohin Zone,
) Partie {
	objekte := make([]Spielobjekt, len(p.Objekte))
	copy(objekte, p.Objekte)

	// **Alles rückt hoch** — die obersten N werden gleich neu einsortiert.
	for stelle := range objekte {
		if objekte[stelle].Zone == ZoneBibliothek &&
			objekte[stelle].Beherrscher == wer {
			objekte[stelle].Ordnung += len(stellen)
		}
	}

	bleiben := 0

	for _, stelle := range stellen {
		if dabei(gewaehlt, objekte[stelle].Kennung) {
			objekte[stelle].Zone = wohin
			objekte[stelle].Ordnung = unterstePlus(objekte, wer, wohin)
			objekte[stelle].Beherrscher = objekte[stelle].Eigner

			continue
		}

		objekte[stelle].Ordnung = bleiben
		bleiben++
	}

	p.Objekte = objekte

	return p
}

// unterstePlus gibt die Ordnung **unter** allem, was schon dort liegt.
//
// **Für den Friedhof heißt das ganz unten** (CR 701.25a legt dort ab), für
// die Bibliothek ebenso (CR 701.22a sagt „nach unten").
func unterstePlus(objekte []Spielobjekt, wer string, zone Zone) int {
	unterste := -1

	for stelle := range objekte {
		if objekte[stelle].Zone == zone &&
			objekte[stelle].Eigner == wer &&
			objekte[stelle].Ordnung > unterste {
			unterste = objekte[stelle].Ordnung
		}
	}

	return unterste + 1
}

// dabei sagt, ob eine Kennung in der Liste steht.
func dabei(liste []string, was string) bool {
	for _, eine := range liste {
		if eine == was {
			return true
		}
	}

	return false
}
