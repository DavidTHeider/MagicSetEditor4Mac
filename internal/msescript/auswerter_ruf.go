package msescript

import (
	"context"
	"slices"
)

// ruferufen ruft einen Funktionswert mit einem fertigen Ruf.
//
// # Die Umgebung reist mit, und das ist der Punkt
//
// In einer frischen Umgebung, die nur die Standardbibliothek kennt, wäre
// jeder Name der Vorlage unbekannt, und ein verketteter **Block** scheiterte.
//
// `magic.mse-game` verkettet aber genau so: `mana_filter := to_upper + { if
// chosen(set.mana_cost_sorting, …) … }`. Bei MSE wird der rechte Teil im
// laufenden Kontext ausgewertet (`ScriptCompose::eval`,
// `src/script/context.cpp`, Zeilen 383–400, Fassung f9d9356 vom 2021-07-06);
// `set` steht dort. Ohne sie scheiterte `mana_filter` an `set` — 1.330-mal
// über die 190 Stile — und `text_filter` an `inserts_count`, 5.076-mal
// (#345).
//
// **Ein Verweis auf ein anderes Skript desselben Netzes wird benannt
// abgewiesen.** Er hat weder Rumpf noch Go-Funktion; ihn zu rufen ginge nur
// über das Netz, und dieses hier kennt keines. Kein gesichtetes Paket verkettet
// so — die Meldung sagt trotzdem, was los ist, statt „diese Funktion hat keinen
// Rumpf" zu behaupten.
func ruferufen(ctx context.Context, ziel Wert, ruf Ruf, umgebung *Umgebung) (Wert, error) {
	teil := ziel.Funktion

	if teil.Eingebaut != nil {
		// **Die Umgebung reist mit.** Ohne sie kann `replace` seine
		// Ersatzfunktion nicht rufen — und genau diesen Weg nimmt der
		// Kartentext, denn `text_filter` ist eine Kette aus `+`.
		voll := Ruf{
			Stellung: ruf.Stellung, Benannt: verbunden(teil.Gebunden, ruf.Benannt),
			Zeile: ruf.Zeile, Spalte: ruf.Spalte, Umgebung: umgebung,
		}

		return teil.Eingebaut(ctx, voll.mitEingabe())
	}

	if teil.Rumpf == nil {
		return Nichts, NeuerFehler(ruf.Zeile, ruf.Spalte, SchluesselKeinVerweis)
	}

	innen := umgebung.Darueber()

	// **Dieselbe Reihenfolge wie bei jedem anderen Aufruf** — siehe
	// [belegeAufruf]. Ohne das Stellungsargument sähe ein verketteter Block
	// sein eigenes Zwischenergebnis nicht.
	belegeAufruf(innen, teil, ruf)

	a := &auswerter{ctx: ctx}

	return a.werte(teil.Rumpf, innen)
}

// verbunden reiht die gebundenen Argumente vor die des Rufs.
//
// **Ohne Kopie, wo eine Seite leer ist** — und das ist fast immer: gebunden
// wird nur mit `@`. Bis E03
// (#2240)
// kopierte jedes Glied einer Kette die Liste zweimal. Gelesen wird sie nur;
// `slices.Clip` nimmt ihr die freie Kapazität, damit ein späteres `append` sie
// kopiert, statt in die Liste des Aufrufers zu schreiben.
func verbunden(gebunden, benannt []Eintrag) []Eintrag {
	switch {
	case len(gebunden) == 0:
		return slices.Clip(benannt)
	case len(benannt) == 0:
		return slices.Clip(gebunden)
	}

	return slices.Concat(gebunden, benannt)
}
