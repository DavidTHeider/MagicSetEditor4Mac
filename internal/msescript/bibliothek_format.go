package msescript

import (
	"context"
	"fmt"
	"math"
	"strings"
)

// **`format`, `format_rule` und das `format:` von `to_string`** —
// #843.
//
// # Was MSE tut
//
// `format_input` (`src/script/functions/basic.cpp`, Zeilen 70–87, Fassung
// f9d9356 vom 2021-07-06):
//
//  1. Das Formatwort wird **normalisiert**: jedes `%` gestrichen, eines davor.
//     `"6.2f"` und `"%6.2f"` sind dasselbe.
//  2. Der **Typ** entscheidet sich am ersten gefundenen Buchstaben:
//     `DdIiOoXx` → ganze Zahl, `EeFfGg` → Gleitkomma, `Ss` → Text.
//  3. Sonst ein Fehler. MSE rät nicht, welche Umwandlung gemeint war.
//
// # Der Datums-Zweig entfällt
//
// MSE prüft zuerst auf `SCRIPT_DATETIME` und formatiert dann über
// `wxDateTime::Format`. **Dieses Projekt hat keinen Werttyp für Zeitpunkte**,
// und `domain.CardRow` führt weder `time_created` noch `time_modified`
// (`service/exportdaten.go`, dort benannt). Ein Format wie `"%y-%m-%d"` fällt
// damit auf den Zahl-Zweig — genau das tut MSE für einen Nicht-Datumswert
// ebenfalls.
//
// # Woher die Zahlen kommen
//
// **Vom Orakel** (§15). Der erste Anlauf rief `format("%d", 42)` — und MSE
// antwortet „Can't convert from built-in function 'format' to string": die
// Argumente sind **benannt**, und ohne sie sucht die Funktion die Variable
// `format` und findet sich selbst.

// formatwoerter sind die Buchstabenklassen und ihre Umwandlung.
//
// **Die Reihenfolge ist die von MSE** und keine Geschmacksfrage: `find_first_of`
// wird dreimal nacheinander gefragt, und `"%x"` trifft die erste Klasse.
var formatwoerter = []struct {
	buchstaben string
	art        string
}{
	{"DdIiOoXx", "ganz"},
	{"EeFfGg", "gleit"},
	{"Ss", "text"},
}

// zahlverben sagt, wie ein Ganzzahlbuchstabe in Go heißt und ob C dort
// **vorzeichenlos** rechnet —
// #850.
//
// # Warum eine Übersetzung nötig ist
//
// **Go kennt kein `%i`.** MSE zählt es zur Ganzzahlklasse, und es ist die
// geläufige C-Schreibweise für `%d`; ohne diese Zeile schriebe `fmt.Sprintf`
// die Zeichenfolge `%!i(int64=-42)` mitten in eine Karte.
//
// **Und `%x`, `%X`, `%o` nehmen in C ein `unsigned int`.** Ein negativer Wert
// erscheint deshalb als sein Bitmuster: `%x` von −42 gibt `ffffffd6`, nicht
// `-2a`. Gefragt am Orakel (§15), nicht erschlossen.
var zahlverben = map[rune]struct {
	verb          rune
	vorzeichenlos bool
}{
	'd': {'d', false},
	'i': {'d', false},
	'o': {'o', true},
	'x': {'x', true},
	'X': {'X', true},

	// **`D`, `I` und `O` wählen bei MSE dieselbe Klasse** (`find_first_of`
	// kennt sie), und was die C-Bibliothek daraus macht, ist **Zufall dieser
	// Maschine**: gemessen `%D` → `4294967254`, `%I` → `"I"`, `%O` →
	// `37777777726`. Das erste ist ein nullerweiterter `int`, gelesen als
	// `%ld` — eine Eigenschaft der arm64-ABI und keine Semantik von
	// MSE-Script. **Die Reichweite einer Messung ist die Reichweite ihrer
	// Aussage** (§10): eine Messung auf einem Rechner belegt keine
	// Sprachregel, und deshalb wird sie hier benannt statt nachgebaut
	// (`docs/FORMATE.md`).
	'D': {'d', false},
	'I': {'d', false},
	'O': {'o', true},
}

// formatiere ist `format_input`.
func formatiere(wort string, eingabe Wert) (string, error) {
	// **Erst normalisieren, dann suchen** — aber gesucht wird im **rohen**
	// Wort, so wie MSE es tut (`format.find_first_of`, nicht `fmt`).
	gereinigt := "%" + strings.ReplaceAll(wort, "%", "")

	for _, klasse := range formatwoerter {
		if !strings.ContainsAny(wort, klasse.buchstaben) {
			continue
		}

		switch klasse.art {
		case "ganz":
			return ganzzahl(gereinigt, klasse.buchstaben, eingabe)

		case "gleit":
			zahl, err := eingabe.AlsZahl()
			if err != nil {
				return "", err
			}

			return fmt.Sprintf(gereinigt, zahl), nil

		default:
			return fmt.Sprintf(gereinigt, eingabe.AlsText()), nil
		}
	}

	return "", fmt.Errorf("%q ist kein bekanntes Format", gereinigt)
}

// ganzzahl setzt die Ganzzahlklasse — mit C-Buchstaben und C-Breite.
func ganzzahl(gereinigt, buchstaben string, eingabe Wert) (string, error) {
	zahl, err := eingabe.AlsZahl()
	if err != nil {
		return "", err
	}

	stelle := strings.IndexAny(gereinigt, buchstaben)

	// **Der Buchstabe steht immer da.** [formatiere] kommt nur hierher, wenn
	// das rohe Wort einen trägt, und das Normalisieren streicht allein `%`.
	// Ein Zweig für „nicht gefunden" wäre einer, den kein Aufruf betritt (§3).
	buchstabe := rune(gereinigt[stelle])
	uebersetzt := zahlverben[buchstabe]

	mitGoBuchstabe := gereinigt[:stelle] + string(uebersetzt.verb) + gereinigt[stelle+1:]

	verengt := verenge(zahl)
	if uebersetzt.vorzeichenlos {
		// **Der Überlauf ist die portierte Sache selbst.** `%x` nimmt in C ein
		// `unsigned int`, und −42 erscheint deshalb als `ffffffd6`; gosec sieht
		// die Umdeutung und kennt ihren Grund nicht.
		return fmt.Sprintf(mitGoBuchstabe, uint32(verengt)), nil //#nosec G115 -- die Umdeutung im Zweierkomplement ist der Zweck
	}

	return fmt.Sprintf(mitGoBuchstabe, verengt), nil
}

// verenge macht aus einer Zahl den C++-`int`, mit dem MSE rechnet.
//
// # Warum überhaupt verengt wird
//
// `format_input` übergibt `input.toInt()` an `String::Format`
// (`src/script/functions/basic.cpp`, Zeilen 78–79, Fassung f9d9356 vom
// 2021-07-06), und `ScriptInt` hält einen `int`; ein geparster `long` wird
// beim Anlegen darauf verengt (`src/script/to_value.hpp`, Zeilen 390–392).
// Am Orakel gemessen: `2147483648` gibt `-2147483648`, `4294967295` gibt `-1`.
// Ohne diese Zeile stünde dort die ungekürzte Zahl.
//
// # Was sie nicht leisten kann
//
// **Jenseits von 2⁵³ ist die Zahl schon vorher gerundet.** [Wert] führt sie
// als `float64`, MSE als ganze Zahl; `9223372036854775807` kommt hier als
// 9223372036854775808 an und wird zu **0**, während MSE `-1` gibt. Das ist
// eine Eigenschaft des Werttyps und nicht dieser Rechnung — benannt in
// `docs/FORMATE.md`, statt mit einem Sonderfall überdeckt zu werden.
func verenge(zahl float64) int32 {
	// **`math.Mod` und keine Umwandlung nach `int64`.** Eine Gleitkommazahl
	// außerhalb des Zielbereichs nach `int64` zu wandeln ist in Go nicht
	// erklärt; hier bleibt der Rest immer unter 2³², und der Schritt nach
	// `uint32` ist die Umdeutung im Zweierkomplement, die C++ macht.
	rest := math.Mod(math.Trunc(zahl), 1<<32)

	// **Dasselbe hier**: das Abschneiden auf 32 Bit ist die Verengung, die MSE
	// beim Anlegen eines `ScriptInt` macht — nicht ein übersehener Überlauf.
	return int32(uint32(int64(rest))) //#nosec G115 -- die Verengung auf C-`int` ist der Zweck
}

// formatFunktion ist `format(format: …, input: …)`.
func formatFunktion(_ context.Context, ruf Ruf) (Wert, error) {
	wort, da := ruf.Benannter("format")
	if !da {
		return Nichts, ruf.FehlerAn("format braucht »format«")
	}

	// **`input` kommt benannt **oder** an erster Stelle.**
	//
	// In MSE bindet ein Stellungsargument an den Namen `input` — deshalb liest
	// `script_format` es über `SCRIPT_PARAM_C(ScriptValueP, input)` und
	// `to_upper("abc")` funktioniert trotzdem. Ohne diesen Weg scheiterte
	// `format_rule(format: "%05d")(42)`: die Regelfassung bindet `format:` und
	// reicht die 42 an der Stelle nach.
	eingabe, da := ruf.Benannter("input")
	if !da {
		if len(ruf.Stellung) == 0 {
			return Nichts, ruf.FehlerAn("format braucht »input«")
		}

		eingabe = ruf.Stellung[0]
	}

	gesetzt, err := formatiere(wort.AlsText(), eingabe)
	if err != nil {
		return Nichts, ruf.FehlerAn(err.Error())
	}

	return Text(gesetzt), nil
}

func init() {
	eintragen("format", formatFunktion)
}
