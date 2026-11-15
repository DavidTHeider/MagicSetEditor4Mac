package sammelliste_test

// **Eine Kartenliste und ein CSV aus fremder Hand** — Werkstück O11 von
// Phase 5.2 (#1435).
//
// # Warum
//
// Beide Formate kommen von außen: die **Liste** aus einem Deckbauwerkzeug
// oder aus der Zwischenablage, das **CSV** aus einem Marktplatz — oder aus
// diesem Programm selbst, als Sicherung. Für das CSV sagt
// `docs/FAEHIGKEITEN.md` ausdrücklich zu, dass es sich **vollständig**
// zurücklesen lässt; wer eine Sicherung darauf stützt, verlässt sich darauf.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
//   - Panic · Hänger — auch bei einer Datei, die nur aus Trennzeichen besteht
//   - zweimal dieselbe Eingabe, zwei Ergebnisse
//   - **ein Befund ohne Meldung**, und bei einer Zeile: ohne ihren Ort (§8)
//   - **eine angenommene Zeile mit einer Menge unter eins** — sie stünde als
//     Stapel aus null Karten im Bestand
//   - **eine angenommene Listenzeile ohne Namen**: die Liste kennt keine
//     Druckkennung, und ohne Namen ist nichts zu suchen
//   - **ein Rundlauf, der etwas verliert**: was `LiesCSV` annimmt und
//     `SchreibeCSV` schreibt, muss dasselbe wieder ergeben — das ist die
//     Zusage, auf der eine Sicherung ruht
//
// **Was hier nicht steht**: eine leere Druckkennung ist **kein** Fund. Eine
// CSV-Zeile darf sie leer lassen; `service.Sammlungsausfuhr.Einfuegen` setzt
// dann den Alias des getroffenen Drucks ein. Nachgelesen am 2026-09-11, bevor
// die Zusicherung geschrieben wurde.
//
//	go test ./internal/sammelliste -run Fuzz -fuzz FuzzLiesListe -fuzztime 2m
//	go test ./internal/sammelliste -run Fuzz -fuzz FuzzLiesCSV -fuzztime 2m

import (
	"reflect"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sammelliste"
)

// listensaat sind die Gestalten einer Kartenliste.
var listensaat = []string{
	"",
	"4 Lightning Bolt",
	"4x Lightning Bolt",
	"1 Lightning Bolt (M11) 149",
	"Deck\n4 Lightning Bolt\n\nSideboard\n2 Duress",
	"# Kommentar\n// noch einer\n4 Lightning Bolt",
	"0 Lightning Bolt",
	"-3 Lightning Bolt",
	"4",
	"4 ",
	"999999999999999999999 Lightning Bolt",
	"4 Æther Vial (M11) 42",
	"4 Lightning Bolt (M11)",
	"commander\n1 Atraxa",
	"\n\n\n",
}

// csvsaat sind die Gestalten eines Sammlungs-CSV.
var csvsaat = []string{
	"printing,menge\nabc,1",
	"printing,menge,zustand,sprache,foil\nabc,2,played,de,true",
	"printing,menge\nabc,0",
	"printing,menge\nabc,keine",
	"printing,menge\n,3",
	"menge\n1",
	"printing\nabc",
	"printing,menge,notiz\nabc,1,\"mit, Komma\"",
	"printing,menge,unbekannt\nabc,1,x",
	"printing,menge\nabc,1\nabc,2\n",
	"printing,menge\nabc,1,zuviel",
	"\"",
	",",
	"\n",
	"printing,menge\nabc,1e3",
	"printing,menge,erwerbspreis,erwerbswaehrung\nabc,1,4.17,EUR",
	"printing,menge,lagerort,wunsch\nabc,1,Kiste 4,true",
}

// **Der Listenleser bricht nicht ab, und was er annimmt, ist eine Karte.**
func FuzzLiesListe(f *testing.F) {
	for _, saat := range listensaat {
		f.Add(saat)
	}

	f.Fuzz(func(t *testing.T, text string) {
		zeilen, befunde := sammelliste.Lies(text)

		nochmal, zweiteBefunde := sammelliste.Lies(text)
		if !reflect.DeepEqual(zeilen, nochmal) ||
			!reflect.DeepEqual(befunde, zweiteBefunde) {
			t.Fatalf("derselbe Text gab zwei Ergebnisse: %d/%d Zeilen, %d/%d Befunde",
				len(zeilen), len(nochmal), len(befunde), len(zweiteBefunde))
		}

		pruefeBefunde(t, befunde, true)

		for _, zeile := range zeilen {
			if zeile.Menge < 1 {
				t.Errorf("eine angenommene Zeile mit Menge %d: %+v",
					zeile.Menge, zeile)
			}

			if strings.TrimSpace(zeile.Name) == "" {
				t.Errorf("eine angenommene Zeile ohne Namen: %+v", zeile)
			}

			if zeile.Zeilennummer < 1 {
				t.Errorf("eine angenommene Zeile ohne Zeilennummer: %+v", zeile)
			}
		}
	})
}

// **Das CSV bricht nicht ab, und der Rundlauf verliert nichts.**
func FuzzLiesCSV(f *testing.F) {
	for _, saat := range csvsaat {
		f.Add(saat)
	}

	f.Fuzz(func(t *testing.T, text string) {
		zeilen, befunde := sammelliste.LiesCSV(text)

		nochmal, zweiteBefunde := sammelliste.LiesCSV(text)
		if !reflect.DeepEqual(zeilen, nochmal) ||
			!reflect.DeepEqual(befunde, zweiteBefunde) {
			t.Fatalf("dasselbe CSV gab zwei Ergebnisse: %d/%d Zeilen, %d/%d Befunde",
				len(zeilen), len(nochmal), len(befunde), len(zweiteBefunde))
		}

		pruefeBefunde(t, befunde, false)

		for _, zeile := range zeilen {
			if zeile.Menge < 1 {
				t.Errorf("eine angenommene Zeile mit Menge %d: %+v",
					zeile.Menge, zeile)
			}
		}

		if len(zeilen) == 0 {
			return
		}

		// **Der Rundlauf**: geschrieben und wieder gelesen ist dasselbe.
		zurueck, rundbefunde := sammelliste.LiesCSV(sammelliste.SchreibeCSV(zeilen))

		if len(rundbefunde) != 0 {
			t.Errorf("das selbst geschriebene CSV gab %d Befunde: %+v",
				len(rundbefunde), rundbefunde)
		}

		if !reflect.DeepEqual(zeilen, zurueck) {
			t.Errorf("der Rundlauf verlor etwas:\n vorher %+v\nnachher %+v",
				zeilen, zurueck)
		}
	})
}

// pruefeBefunde hält, dass jeder Befund sagt, was auffiel — und wo.
func pruefeBefunde(t *testing.T, befunde []domain.Befund, ortNoetig bool) {
	t.Helper()

	for _, befund := range befunde {
		if strings.TrimSpace(befund.Meldung) == "" {
			t.Errorf("ein Befund ohne Meldung: %+v", befund)
		}

		if ortNoetig && strings.TrimSpace(befund.Ort) == "" {
			t.Errorf("ein Befund ohne Ort: %+v", befund)
		}
	}
}
