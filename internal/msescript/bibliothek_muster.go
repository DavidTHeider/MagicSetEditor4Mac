package msescript

import "context"

// `match` — die meistgerufene Funktion der bezogenen Pakete.
//
// # Erschlossen, nicht erfunden
//
// `magic.mse-game` ruft sie **616-mal**, drei der neunzehn Exportvorlagen
// hängen an ihr, und sie war der erste Fehler, nachdem sich das Initskript
// vollständig zerteilen ließ
// (#214).
//
// # Welche Regex-Sprache
//
// **Die von MSE**, und das steht vollständig bei [Muster]: RE2 für die 1.630
// Muster, die es trägt, und eine Maschine mit Rückverfolgung für die
// fünfzehn, die mehr brauchen.
func init() {
	eintragen("match", passt)
}

// passt sagt, ob ein Muster in einem Text vorkommt.
//
// # Die Form, belegt
//
// Zwei Gestalten, beide aus `magic.mse-game`:
//
//	match(param1.value, match: " land$")     // unmittelbar
//	match@(match: "(?i)…")                   // teilweise, 606-mal
//
// Der Text steht in der Stellung, das Muster unter `match:`. Die zweite Form
// entsteht ohne Zutun: `@` bindet die benannten Argumente und lässt die
// Stellung offen.
func passt(_ context.Context, ruf Ruf) (Wert, error) {
	roh, benannt := ruf.Benannter("match")
	if !benannt {
		return Nichts, ruf.fehlerAn("match braucht ein Argument `match:`")
	}

	if err := ruf.eineEingabe("match"); err != nil {
		return Nichts, err
	}

	// **Das Muster steht in der Meldung.** Ein Lauf über dreißig Module meldete
	// sonst „ein Muster war unbrauchbar", und niemand fände heraus, welches von
	// 616. Was `UebersetzeMuster` alles trägt, steht bei [Muster].
	muster, err := UebersetzeMuster(roh.AlsText())
	if err != nil {
		return Nichts, ruf.fehlerAn("match bekam " + err.Error())
	}

	trifft, err := muster.Passt(ruf.Stellung[0].AlsText())
	if err != nil {
		return Nichts, ruf.fehlerAn("match: " + err.Error())
	}

	return Wahrheit(trifft), nil
}
