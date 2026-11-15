package service

// **Was die zwei Seiten eines Tauschs ungefähr wert sind** — TRADE-02b
// (#1951).
//
// # Ein Anhalt und keine Bewertung
//
// **Die Preise gehen nirgends hin** (AK-03): nicht in den Vertrag, nicht in
// den Abdruck, nicht über die Leitung. Sie werden bei jedem Lesen aus dem
// Preisbestand gerechnet und nie gespeichert — was zwei Parteien vereinbaren,
// ist ihre Sache; die Bewertung, die über die Leitung geht, regelt TRADE-03.
//
// # Dieselbe Regel wie das Portfolio
//
// **Ein Preis zählt hier, wo er dort zählt** — genannt, Papier, Verkauf,
// kein `tix`, veredelt zu veredelt (`jeWaehrungEiner`). Eine zweite Regel
// für dieselbe Frage liefe beim ersten neuen Markt von der ersten weg.
//
// # Und keine Umrechnung
//
// Summen und Differenz stehen **je Währung** (wie beim Portfolio): ein Kurs
// wäre eine Quelle, deren Bedingungen niemand gelesen hat.

import (
	"cmp"
	"context"
	"slices"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Preislese ist, was der Anhalt vom Preisbestand braucht.
type Preislese interface {
	Neueste(ctx context.Context, printing string) ([]domain.Preisstand, error)
}

// Anhaltspreis ist ein Betrag in einer Währung.
type Anhaltspreis struct {
	Waehrung string

	// Wert steht in der kleinsten Einheit der Währung.
	Wert int64

	Quelle domain.Preisquelle

	// Erhoben ist bei einer Summe der **älteste** Bestandteil — eine Summe
	// ist so frisch wie ihr ältester Preis.
	Erhoben time.Time
}

// Anhaltszeile ist der Preis einer Position mal ihrer Menge, je Währung.
//
// **Leer heißt „kein Preis"** und nicht null: eine Karte, deren Preis
// niemand kennt, ist nicht wertlos.
type Anhaltszeile struct {
	Preise []Anhaltspreis
}

// Anhaltsseite ist, was eine Partei gibt.
type Anhaltsseite struct {
	Von    string
	Summen []Anhaltspreis

	// OhnePreis zählt die Zeilen dieser Seite ohne jeden Preis.
	OhnePreis int
}

// Anhaltsdifferenz ist erhalten weniger gegeben, aus Sicht von `ich`.
type Anhaltsdifferenz struct {
	Waehrung string
	Wert     int64

	// Vollstaendig heißt: **jede** Zeile beider Seiten nennt einen Preis in
	// dieser Währung. Sonst ist die Differenz eine über einen Teil.
	Vollstaendig bool
}

// Preisanhalt ist der ganze Anhalt über einen Tausch.
type Preisanhalt struct {
	// Zeilen stehen in der Reihenfolge der Positionen.
	Zeilen []Anhaltszeile

	// Seiten: die eigene zuerst, dann die übrigen in der Reihenfolge ihres
	// ersten Auftretens.
	Seiten []Anhaltsseite

	// Differenzen stehen nach Währung geordnet.
	Differenzen []Anhaltsdifferenz

	// Hauptwaehrung ist die, die eine Oberfläche zuerst zeigt: die Vorgabe,
	// wo eine Zeile sie nennt, sonst die mit den meisten Zeilen.
	Hauptwaehrung string
}

// Handelsanhalt rechnet den Anhalt über Positionen.
//
// `vorgabe` ist die bevorzugte Währung (`vorgabe.waehrung`) — leer heißt:
// keine.
func Handelsanhalt(
	ctx context.Context, lese Preislese, positionen []domain.Handelsposition,
	ich, vorgabe string,
) (Preisanhalt, error) {
	jeDruck := map[string][]domain.Preisstand{}

	for _, position := range positionen {
		if _, gefragt := jeDruck[position.Printing]; gefragt {
			continue
		}

		staende, fehler := lese.Neueste(ctx, position.Printing)
		if fehler != nil {
			return Preisanhalt{}, fehler
		}

		jeDruck[position.Printing] = staende
	}

	anhalt := Preisanhalt{Zeilen: make([]Anhaltszeile, 0, len(positionen))}
	seiten := map[string]*Anhaltsseite{}
	reihe := []string{}
	zeilenJeWaehrung := map[string]int{}

	for _, position := range positionen {
		zeile := anhaltszeile(jeDruck[position.Printing], position)
		anhalt.Zeilen = append(anhalt.Zeilen, zeile)

		seite, bekannt := seiten[position.Von]
		if !bekannt {
			seite = &Anhaltsseite{Von: position.Von}
			seiten[position.Von] = seite
			reihe = append(reihe, position.Von)
		}

		if len(zeile.Preise) == 0 {
			seite.OhnePreis++
		}

		for _, preis := range zeile.Preise {
			seite.Summen = addiert(seite.Summen, preis)
			zeilenJeWaehrung[preis.Waehrung]++
		}
	}

	// **Die eigene Seite zuerst**, die übrigen in ihrer Reihenfolge.
	if eigene, da := seiten[ich]; da {
		anhalt.Seiten = append(anhalt.Seiten, *eigene)
	}

	for _, von := range reihe {
		if von != ich {
			anhalt.Seiten = append(anhalt.Seiten, *seiten[von])
		}
	}

	anhalt.Differenzen = differenzen(anhalt.Seiten, zeilenJeWaehrung, len(positionen), ich)
	anhalt.Hauptwaehrung = hauptwaehrung(zeilenJeWaehrung, vorgabe)

	return anhalt, nil
}

// anhaltszeile nimmt je Währung den Preis, den das Portfolio nähme.
func anhaltszeile(staende []domain.Preisstand, position domain.Handelsposition) Anhaltszeile {
	gedacht := &domain.Exemplar{Printing: position.Printing, Finish: domain.Finish(position.Finish)}
	gewaehlt := jeWaehrungEiner(staende, gedacht)

	slices.SortFunc(gewaehlt, func(links, rechts domain.Preisstand) int {
		return cmp.Compare(links.Waehrung, rechts.Waehrung)
	})

	zeile := Anhaltszeile{Preise: make([]Anhaltspreis, 0, len(gewaehlt))}
	for _, stand := range gewaehlt {
		zeile.Preise = append(zeile.Preise, Anhaltspreis{
			Waehrung: stand.Waehrung, Wert: stand.Wert * int64(position.Menge),
			Quelle: stand.Quelle, Erhoben: stand.Erhoben,
		})
	}

	return zeile
}

// addiert trägt einen Betrag in die Summen seiner Währung ein.
func addiert(summen []Anhaltspreis, preis Anhaltspreis) []Anhaltspreis {
	for stelle := range summen {
		if summen[stelle].Waehrung != preis.Waehrung {
			continue
		}

		summen[stelle].Wert += preis.Wert
		if preis.Erhoben.Before(summen[stelle].Erhoben) {
			summen[stelle].Erhoben = preis.Erhoben
		}

		return summen
	}

	return append(summen, preis)
}

// differenzen rechnet je Währung erhalten weniger gegeben.
func differenzen(
	seiten []Anhaltsseite, zeilenJeWaehrung map[string]int, zeilen int, ich string,
) []Anhaltsdifferenz {
	waehrungen := make([]string, 0, len(zeilenJeWaehrung))
	for waehrung := range zeilenJeWaehrung {
		waehrungen = append(waehrungen, waehrung)
	}

	slices.Sort(waehrungen)

	ergebnis := make([]Anhaltsdifferenz, 0, len(waehrungen))

	for _, waehrung := range waehrungen {
		differenz := Anhaltsdifferenz{
			Waehrung: waehrung, Vollstaendig: zeilenJeWaehrung[waehrung] == zeilen,
		}

		for _, seite := range seiten {
			betrag := summeIn(seite.Summen, waehrung)
			if seite.Von == ich {
				differenz.Wert -= betrag
			} else {
				differenz.Wert += betrag
			}
		}

		ergebnis = append(ergebnis, differenz)
	}

	return ergebnis
}

// summeIn gibt den Betrag einer Währung oder null.
func summeIn(summen []Anhaltspreis, waehrung string) int64 {
	for _, summe := range summen {
		if summe.Waehrung == waehrung {
			return summe.Wert
		}
	}

	return 0
}

// hauptwaehrung wählt, was zuerst gezeigt wird.
func hauptwaehrung(zeilenJeWaehrung map[string]int, vorgabe string) string {
	if zeilenJeWaehrung[vorgabe] > 0 {
		return vorgabe
	}

	beste := ""
	for waehrung, zahl := range zeilenJeWaehrung {
		if zahl > zeilenJeWaehrung[beste] || (zahl == zeilenJeWaehrung[beste] && waehrung < beste) {
			beste = waehrung
		}
	}

	return beste
}
