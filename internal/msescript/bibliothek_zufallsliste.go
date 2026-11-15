package msescript

// **Die drei Zufallsfunktionen über Listen** — Werkstück A von Phase 4.9
// (#1324).
//
// # Warum sie erst jetzt kommen
//
// `bibliothek_zufall.go` führte sie ausdrücklich als „was hier fehlt, und
// warum": sie stehen in **keiner** Datei des vollen Packs, und sie sind eine
// andere Familie als „gib mir eine Zahl". Das war richtig, solange die
// Bibliothek nur trug, was eine Vorlage wirklich ruft.
//
// **Phase 4.9 hat ein anderes Ziel**: die Sprache **vollständig**. Ein Name,
// den MSE2 kennt und dieses Programm nicht, ist eine Vorlage, die hier nicht
// läuft — gleich ob heute eine davon Gebrauch macht.
//
// Alles aus `src/script/functions/basic.cpp`, Fassung f9d9356 vom 2021-07-06.

import (
	"context"
	"math/rand/v2"
	"strconv"
)

func init() {
	eintragen("random_shuffle", mischeListe)
	eintragen("random_select", waehleEines)
	eintragen("random_select_many", waehleMehrere)
}

// mischeListe ist `random_shuffle(liste)`.
//
//	ScriptValueP it = input->makeIterator();
//	while (ScriptValueP v = it->next()) { ret->value.push_back(v); }
//	shuffle(ret->value.begin(), ret->value.end());
//
// Zeilen 604–615. **Die Namen werden neu vergeben** — `ScriptCustomCollection`
// bekommt die Werte über `push_back`; dieselbe Regel wie bei `filter_list`.
func mischeListe(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("random_shuffle"); err != nil {
		return Nichts, err
	}

	glieder := gliederAus(ruf.Stellung[0])

	rand.Shuffle(len(glieder), func(i, j int) { // #nosec G404 -- siehe bibliothek_zufall.go
		glieder[i], glieder[j] = glieder[j], glieder[i]
	})

	return listeAus(glieder), nil
}

// waehleEines ist `random_select(liste)`.
//
//	if (itemCount == 0) {
//	  throw ScriptError(_("Can not select a random item from an empty collection"));
//	}
//	return input->getIndex( rand() % itemCount );
//
// Zeilen 617–625.
func waehleEines(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("random_select"); err != nil {
		return Nichts, err
	}

	glieder := gliederAus(ruf.Stellung[0])
	if len(glieder) == 0 {
		return Nichts, ruf.fehlerAn(
			"random_select: aus einer leeren Liste lässt sich nichts wählen")
	}

	return glieder[rand.IntN(len(glieder))], nil // #nosec G404 -- siehe oben
}

// waehleMehrere ist `random_select_many(liste, count: n, replace: b)`.
//
//	bool with_replace = replace && replace->type() != SCRIPT_FUNCTION
//	                    && replace->toBool();
//
// Zeilen 627–656. **`replace:` ist eine Wahrheit und keine Funktion** — die
// Prüfung auf `SCRIPT_FUNCTION` steht dort, weil derselbe Name bei `replace`
// eine Funktion trägt; wer sie hier mitgibt, meint offenbar etwas anderes und
// bekommt „ohne Zurücklegen".
func waehleMehrere(_ context.Context, ruf Ruf) (Wert, error) {
	if err := ruf.eineEingabe("random_select_many"); err != nil {
		return Nichts, err
	}

	anzahlwert, da := ruf.Benannter("count")
	if !da {
		return Nichts, ruf.fehlerAn("random_select_many braucht count:")
	}

	zahl, err := anzahlwert.AlsZahl()
	if err != nil {
		return Nichts, ruf.fehlerAn(
			"random_select_many braucht für count: eine Zahl: " + err.Error())
	}

	anzahl := int(zahl)

	// **Eine negative Zahl wird abgewiesen, nicht abgeschnitten** — Befund 2
	// des Zweitblicks über die Phasen 4.0 bis 4.9.
	//
	// Vorher erreichte sie entweder `glieder[:anzahl]` oder ein `make` mit
	// negativer Kapazität; beides ist eine **Panik**, und die beendet den
	// Prozess. §8 sagt, fremde Daten sind der Normalfall: ein Set, von dem
	// 300 Karten lesbar sind und zwei nicht, wird mit 300 Karten geöffnet —
	// ein Skript, das das Fenster schließt, macht daraus nichts.
	//
	// **MSE2 stirbt daran ebenfalls**, am Orakel gemessen am 2026-09-03:
	// `1+1` gibt Rückgabewert 0, dieser Ausdruck gibt 1 und keine Antwort.
	// Der Grund steht im Quelltext — `ret->value.resize(count)` mit
	// negativem `count` (`src/script/functions/basic.cpp`, Zeile 653,
	// Fassung f9d9356 vom 2021-07-06).
	//
	// **Abgewiesen und nicht auf null geklemmt**: eine leere Liste sähe aus
	// wie ein Ergebnis. Wer `count: -1` schreibt, hat sich verrechnet, und
	// das soll er erfahren.
	if anzahl < 0 {
		return Nichts, ruf.fehlerAn("random_select_many: " +
			strconv.Itoa(anzahl) + " Glieder lassen sich nicht wählen — " +
			"count: braucht eine Zahl ab null")
	}

	// **Und nach oben ebenso** — Befund 1 des Nachlaufs über die Phasen 4.0
	// bis 4.9.
	//
	// Der erste Anlauf wies nur negative Zahlen ab. Mit `replace: true` ging
	// der Wert danach ungeprüft als Kapazität an `make`: `count: 10^12` gab
	// `makeslice: cap out of range` oder erschöpfte den Speicher — **dieselbe
	// Klasse**, eine Größenangabe aus einem fremden Skript, die den Prozess
	// umbringt. Ohne Zurücklegen fängt die Schranke `anzahl > len(glieder)`
	// es seit je; mit Zurücklegen ist dort jede Zahl fachlich denkbar.
	//
	// **Die Prüfung steht vor beiden Wegen**, nicht nur vor dem einen: eine
	// Grenze, die nur einen Zweig kennt, ist genau der Fehler noch einmal.
	if anzahl > hoechstensGezogen {
		return Nichts, ruf.fehlerAn("random_select_many: " +
			strconv.Itoa(anzahl) + " Glieder sind zu viele — " +
			"count: nimmt höchstens " + strconv.Itoa(hoechstensGezogen))
	}

	glieder := gliederAus(ruf.Stellung[0])

	if mitZuruecklegen(ruf) {
		return zurueckgelegt(ruf, glieder, anzahl)
	}

	if anzahl > len(glieder) {
		return Nichts, ruf.fehlerAn("random_select_many: " +
			strconv.Itoa(anzahl) + " Glieder verlangt, die Liste hat " +
			strconv.Itoa(len(glieder)))
	}

	rand.Shuffle(len(glieder), func(i, j int) { // #nosec G404 -- siehe oben
		glieder[i], glieder[j] = glieder[j], glieder[i]
	})

	return listeAus(glieder[:anzahl]), nil
}

// hoechstensGezogen ist die Obergrenze für `count:`.
//
// # Warum es überhaupt eine gibt
//
// **MSE2 hat keine** (`src/script/functions/basic.cpp`, Zeile 627, Fassung
// f9d9356 vom 2021-07-06) und stirbt an einer großen Zahl ebenso. Dieses
// Programm ist an dieser Stelle bewusst strenger — ein Vorlagenpaket ist
// fremder Inhalt (§8), und ein Vertipper darin soll nicht das Fenster
// mitnehmen.
//
// # Warum diese Zahl
//
// **Gemessen am 2026-09-03: im vollen Pack kommt `random_select_many`
// null-mal vor.** Es gibt also keine Vorlage, an der sich eine Grenze messen
// ließe — und das gehört zur Zahl dazu. Gewählt ist eine, die jede denkbare
// Ziehung durchlässt (ein Booster hat fünfzehn Karten, ein Set einige
// hundert) und jeden Vertipper fängt.
//
// **Sie ist keine Speichergrenze.** Hunderttausend Werte sind wenige
// Megabyte; was sie verhindert, ist die Zahl mit zwölf Stellen.
const hoechstensGezogen = 100000

// mitZuruecklegen sagt, ob `replace:` das Zurücklegen meint.
func mitZuruecklegen(ruf Ruf) bool {
	wert, da := ruf.Benannter("replace")
	if !da || wert.Art == ArtFunktion {
		return false
	}

	trifftZu, _ := wert.AlsWahrheit()

	return trifftZu
}

// zurueckgelegt zieht mit Zurücklegen.
func zurueckgelegt(ruf Ruf, glieder []Wert, anzahl int) (Wert, error) {
	if len(glieder) == 0 {
		return Nichts, ruf.fehlerAn("random_select_many: aus einer leeren " +
			"Liste lassen sich keine " + strconv.Itoa(anzahl) + " wählen")
	}

	gezogen := make([]Wert, 0, anzahl)
	for range anzahl {
		gezogen = append(gezogen, glieder[rand.IntN(len(glieder))]) // #nosec G404 -- siehe oben
	}

	return listeAus(gezogen), nil
}

// gliederAus gibt die Glieder eines Werts als Reihe.
func gliederAus(wert Wert) []Wert {
	glieder := make([]Wert, 0, len(wert.Eintraege))
	for i := range wert.Eintraege {
		glieder = append(glieder, wert.Eintraege[i].Wert)
	}

	return glieder
}

// listeAus baut eine Abbildung mit neu vergebenen Nummern.
func listeAus(glieder []Wert) Wert {
	gebaut := Wert{Art: ArtAbbildung, Eintraege: make([]Eintrag, 0, len(glieder))}

	for i := range glieder {
		gebaut.Eintraege = append(gebaut.Eintraege, Eintrag{
			Name: strconv.Itoa(i),
			Wert: glieder[i],
		})
	}

	return gebaut
}
