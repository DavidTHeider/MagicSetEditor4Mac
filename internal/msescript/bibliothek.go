package msescript

import (
	"context"
	"errors"
	"strconv"
)

// Ruf sind die Argumente eines Aufrufs, so wie sie dastehen.
//
// Getrennt nach Stellung und Namen, weil beides gemischt vorkommt:
// `contains(style, match:"pw")` und `contains(match:"\n", card.rule_text)` sind
// derselbe Aufruf in anderer Reihenfolge.
type Ruf struct {
	Stellung []Wert
	Benannt  []Eintrag

	Zeile  int
	Spalte int

	// Umgebung ist die Umgebung der Rufstelle.
	//
	// **Damit eine eingebaute Funktion eine übergebene rufen kann.** `replace`
	// nimmt für `replace:` eine Funktion und ruft sie je Treffer
	// (`RegexReplacer::apply`, `src/script/functions/regex.cpp`, Zeilen
	// 78–99); ihr Rumpf läuft dabei in der Umgebung des Aufrufers, so wie
	// jeder andere Rumpf auch.
	//
	// **Sie darf nil sein.** Wer eine eingebaute Funktion von Hand ruft — ein
	// Test, `length` über `number_of_items` —, gibt keine mit, und keine der
	// übrigen fragt danach.
	//
	// Fassung f9d9356 vom 2021-07-06.
	Umgebung *Umgebung
}

// RufeFunktion ruft einen Funktionswert aus einer eingebauten Funktion heraus.
//
// **Der Rumpf läuft in der mitgegebenen Umgebung.** Der Aufrufer legt sie an
// und setzt hinein, was die Funktion sehen soll — bei `replace` sind das
// `input` und die Gruppen `_1` bis `_9`.
func RufeFunktion(ctx context.Context, ziel Wert, umgebung *Umgebung) (Wert, error) {
	if ziel.Funktion == nil {
		return Nichts, errors.New("hier wird " + string(ziel.Art) +
			" gerufen, das geht nur bei einer Funktion")
	}

	eingabe := Nichts
	if wert, da := umgebung.Suche("input"); da {
		eingabe = wert
	}

	return ruferufen(ctx, ziel, Ruf{Stellung: []Wert{eingabe}, Umgebung: umgebung}, umgebung)
}

// mitEingabe macht aus einem benannten `input:` das Stellungsargument.
//
// # Warum das eine Umformung und keine Nachsicht ist
//
// **MSE kennt kein Stellungsargument.** `f(x)` setzt `x` als Kontextvariable
// `input`, und jede eingebaute Funktion liest es dort
// (`SCRIPT_PARAM_C(String, input)`). Die beiden Schreibweisen sind dieselbe
// Sache, und dass dieses Programm sie trennt, ist seine eigene Bauform.
//
// **Sichtbar wird der Unterschied an der Verkettung.** `ScriptCompose` gibt
// das Zwischenergebnis als `input` weiter (`src/script/context.cpp`, Zeilen
// 383–400, Fassung f9d9356 vom 2021-07-06); `text_filter` in
// `magic.mse-game` reiht so rund dreißig teilangewandte Funktionen. Ohne diese
// Umformung bekäme jede von ihnen **kein** Argument
// (#345).
//
// **Ein Stellungsargument schlägt das benannte.** Beides zugleich steht in
// keiner gesichteten Vorlage; die Regel steht hier, damit sie nicht von der
// Reihenfolge einer Schleife abhängt.
func (r Ruf) mitEingabe() Ruf {
	if len(r.Stellung) > 0 {
		return r
	}

	eingabe, da := r.Benannter("input")
	if !da {
		return r
	}

	r.Stellung = []Wert{eingabe}

	return r
}

// Eingebaut ist eine Funktion, die in Go steht statt in einer Vorlage.
type Eingebaut func(ctx context.Context, ruf Ruf) (Wert, error)

// Bibliothek ist die Standardbibliothek als Umgebung.
//
// # Erschlossen, nicht erfunden
//
// Hier steht **nur**, was in den gesichteten Paketen wirklich gerufen wird.
// Eine Funktion, die keine Vorlage benutzt, hilft niemandem und behauptet
// zugleich, MSE habe sie — `docs/FORMATE.md` müsste sie als Vermutung führen,
// und niemand könnte sie je belegen.
//
// **Was hier fehlt und trotzdem in den Dateien steht**, ist der interessantere
// Teil: `hide_foil`, `is_link`, `is_jp` und zwei Dutzend weitere werden in
// `yugioh-series10` gerufen und dort nicht definiert. Sie kommen aus
// `yugioh.mse-game`, das als Abhängigkeit geführt wird. **An den Dateien allein
// ist nicht zu erkennen, was MSE mitbringt und was ein Spielpaket beisteuert** —
// nur `rgb` erscheint in drei verschiedenen Spielen und ist damit sicher keins
// von beidem.
//
// Deshalb wird hier zurückhaltend gebaut: was wie ein allgemeines Werkzeug
// aussieht und in einer Vorlage steht, kommt herein; was nach dem Spiel klingt,
// bleibt draußen, bis ein Paket es widerlegt.
func Bibliothek() *Umgebung {
	umgebung := NeueUmgebung(grundnamen)

	for name, funktion := range eingebaute {
		umgebung.Setze(name, Wert{Art: ArtFunktion, Funktion: &Funktionsteil{Eingebaut: funktion}})
	}

	// **Die Regelfassungen danach**, weil sie ihre Grundfunktion brauchen —
	// siehe [regelfassungenIn].
	regelfassungenIn(umgebung)

	return umgebung
}

// eingebaute wird von den Bibliotheksdateien gefüllt.
var eingebaute = map[string]Eingebaut{}

// eintragen meldet eine eingebaute Funktion an.
func eintragen(name string, funktion Eingebaut) {
	eingebaute[name] = funktion
}

// FehlerAn baut einen Fehler mit der Stelle des Aufrufs.
//
// **Ausgeführt, weil auch der Aufrufer eingebaute Funktionen setzt.**
// `internal/service` gibt `exists_in_package` und `symbol_variation` in die
// Umgebung; ihre Meldungen sollen dieselbe Stelle nennen wie die der
// Bibliothek.
func (r Ruf) FehlerAn(meldung string) *Fehler { return r.fehlerAn(meldung) }

// fehlerAn baut einen Fehler mit der Stelle des Aufrufs.
func (r Ruf) fehlerAn(meldung string) *Fehler {
	// **Durchgereicht und nicht benannt** (#1178). Die 48 Sätze der
	// Standardbibliothek nennen Funktionsnamen und Argumentnamen; ein
	// Katalogeintrag je Funktion wäre eine zweite Abschrift der Bibliothek in
	// zwei Sprachen. Der Grund steht ausführlich bei [SchluesselDurchgereicht].
	return NeuerFehler(r.Zeile, r.Spalte, SchluesselDurchgereicht, meldung)
}

// EineEingabe prüft, dass **mindestens** ein bloßes Argument dasteht.
//
// **Ausgeführt aus demselben Grund wie [Ruf.FehlerAn]**: `internal/service`
// setzt eigene eingebaute Funktionen in die Umgebung, und die sollen dieselbe
// Meldung geben wie die der Bibliothek.
func (r Ruf) EineEingabe(name string) error { return r.eineEingabe(name) }

// eineEingabe prüft, dass **mindestens** ein bloßes Argument dasteht.
//
// # Warum nicht genau eines
//
//	} else {
//	  // implicit "input" argument
//	  arguments.push_back(SCRIPT_VAR_input);
//	  parseOper(input, script, PREC_SEQ);
//	}
//
// `src/script/parser.cpp`, Zeilen 890–894, Fassung f9d9356 vom 2021-07-06:
// **jedes** unbenannte Argument bekommt denselben Namen `input`. Zwei bloße
// Argumente sind bei MSE also kein Fehler, sondern zwei Zuweisungen an
// dieselbe Stelle.
//
// Welche stehen bleibt, entscheidet die Reihenfolge:
//
//	for (unsigned int j = 0 ; j < i.data ; ++j) {
//	  setVariable((Variable)instr[i.data - j - 1].data, stack.back());
//	  stack.pop_back();
//	}
//
// `src/script/context.cpp`, Zeilen 152–155. Die Schleife läuft vom **letzten**
// Argument zum ersten; die letzte Zuweisung bleibt stehen, und das ist das
// **erste** Argument.
//
// # Woran es aufgefallen ist
//
// `magic-m15-mainframe-dfc.mse-style` schreibt
// `to_number(pull_comma_array(styling.back_stripe_coordinate, cell:2), 152)`.
// Die 152 war als Ersatzwert gemeint; MSE wertet sie aus und verwirft sie.
// Bei uns scheiterte der Aufruf, und zwei Ebenen einer Karte fielen aus
// (#438).
//
// **`rgb` und `rgba` sind davon nicht betroffen**: sie sind bei MSE keine
// Funktionen, sondern Parser-Konstrukte (`I_RGB`, `src/script/parser.cpp`,
// Zeilen 610 und 620) und nehmen dort wirklich drei beziehungsweise vier
// Stellungen. Sie benutzen weiter [Ruf.genauStellung].
func (r Ruf) eineEingabe(name string) error {
	if len(r.Stellung) >= 1 {
		return nil
	}

	// **Dieselbe Wortwahl wie [Ruf.genauStellung]** — zwei Meldungen für
	// denselben Mangel läsen sich wie zwei verschiedene Mängel.
	return r.fehlerAn(name + " erwartet " + zahlwort(1) + ", bekam " + zahlwort(0))
}

// genauStellung prüft die Anzahl der Stellungsargumente.
//
// **Nur noch für `rgb` und `rgba`** — der Grund steht bei [Ruf.eineEingabe].
func (r Ruf) genauStellung(name string, anzahl int) error {
	if len(r.Stellung) == anzahl {
		return nil
	}

	return r.fehlerAn(name + " erwartet " + zahlwort(anzahl) + ", bekam " +
		zahlwort(len(r.Stellung)))
}

func zahlwort(anzahl int) string {
	switch anzahl {
	case 0:
		return "kein Argument"

	case 1:
		return "ein Argument"
	}

	return strconv.Itoa(anzahl) + " Argumente"
}

// NeueEingebaute macht aus einer Go-Funktion einen Funktionswert.
//
// **Für Funktionen, die von außen kommen** — solche, die dieses Paket nicht
// haben darf. `write_image_file` schreibt Dateien; `internal/schichten` weist
// jeden Zugriff auf das Dateisystem aus `msescript` ab, und das soll so
// bleiben. Der Aufrufer legt die Funktion in seine Umgebung, statt sie hier
// einzutragen (#216).
func NeueEingebaute(funktion Eingebaut) Wert {
	return Wert{Art: ArtFunktion, Funktion: &Funktionsteil{Eingebaut: funktion}}
}
