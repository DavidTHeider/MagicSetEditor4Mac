package service

import (
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// symbolschriftnamen sammelt die Symbolschriften, die ein Stil brauchen kann.
//
// # Der Befund
//
// Vorgeladen wurde unter dem Namen, wie er im Feldstil **dasteht**; gesucht
// wird unter dem **ausgerechneten**. Beides ist nur bei einem Literal dasselbe.
// `magic-m15.mse-style` schreibt am Regeltextfeld
//
//	symbol font:
//	    name: { styling.text_box_mana_symbols }
//
// und die Stiloption dazu trägt `initial: magic-mana-small.mse-symbol-font`.
//
// **Gezählt über `M15-Magic-Pack`: 326 von 445 `symbol font: name:` sind ein
// Skript**, dazu 160 `initial:`-Werte, die eine Symbolschrift nennen. Die
// Mehrheit wurde nie geladen — und jedes `<sym>` im Regeltext blieb Text
// (#328).
//
// # Warum aus den Stiloptionen und nicht aus dem Skript
//
// Im Skript steht kein Name, sondern ein Verweis: `styling.text_box_mana_symbols`.
// Der Name steht in der **Option**, unter `initial:` und unter jedem `choice:`.
//
// **Auch die Auswahlwerte**, nicht nur der Vorgabewert: eine Option kann jeden
// ihrer Werte annehmen, und wer nur `initial:` lädt, hat die Schrift für jede
// andere Wahl nicht zur Hand.
//
// # Und warum das eine Näherung ist
//
// Ein Skript könnte einen Namen errechnen, der nirgends dasteht — aus zwei
// Teilen zusammengesetzt, aus einem Kartenwert gelesen. Diese Fälle findet
// dieser Weg nicht. **Er ist deshalb an einer Zahl zu messen und nicht an einer
// Behauptung**: wie viele Befunde „die Symbolschrift … liegt nicht vor" über
// die 190 Stile übrig bleiben.
func symbolschriftnamen(stil msepaket.Stil) []string {
	var (
		namen   []string
		gefragt bool
	)

	for i := range stil.Felder {
		roh := strings.TrimSpace(stil.Felder[i].Symbolschrift.Name)
		if roh == "" {
			continue
		}

		gefragt = true

		if _, geklammert := skriptrumpf(roh); !geklammert {
			namen = aufnehmen(namen, roh)
		}
	}

	// **Kein Feld fragt nach einer Symbolschrift, also wird keine gesucht.**
	// Ohne diese Frage lüde ein Stil die Schriften seiner Optionen auch dann,
	// wenn kein Feld sie je verwendet.
	if !gefragt {
		return nil
	}

	for _, option := range stil.Stiloptionen {
		namen = ausOption(namen, option.Initial)

		for _, wahl := range option.Choices {
			namen = ausOption(namen, wahl)
		}
	}

	return namen
}

// ausOption nimmt den Wert einer Stiloption auf, wenn er eine Schrift benennt.
//
// **Die Endung ist die Bedingung.** Ein Wert wie `never` oder
// `short text only` ist keine Schrift; für jede Stiloption ein Paket zu suchen
// hieße, bei 190 Stilen mit im Schnitt zwanzig Optionen tausende Male ins Leere
// zu greifen.
func ausOption(namen []string, roh string) []string {
	wert := strings.TrimSpace(roh)
	if !strings.HasSuffix(wert, "."+string(msepaket.ArtSymbolschrift)) {
		return namen
	}

	return aufnehmen(namen, wert)
}

// aufnehmen legt einen Namen ohne Endung und ohne Doppelung ab.
func aufnehmen(namen []string, roh string) []string {
	name := ohnePaketendung(strings.TrimSpace(roh))
	if name == "" || slices.Contains(namen, name) {
		return namen
	}

	return append(namen, name)
}

// SymbolschriftnamenFuerTest legt [symbolschriftnamen] für den Test frei.
//
// **Weil die Zusage hier steht und nicht am Speicher.** Über
// [Stilspeicher.Symbolschriften] geprüft, prüfte der Test zugleich das Öffnen
// von Paketen — und ein fehlgeschlagenes Öffnen sähe aus wie ein nicht
// gefundener Name.
func SymbolschriftnamenFuerTest(stil msepaket.Stil) []string {
	return symbolschriftnamen(stil)
}
