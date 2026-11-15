// **Wie eine Kommandozeile zerlegt wird** —
// #1169.
//
// **Eine eigene Datei**, weil `lauf.go` mit der Bestandssicherung über die
// Zeilengrenze aus §2 ging. Der Schnitt folgt der Frage: dort steht,
// **welche** Befehle es gibt, hier, **wie** die Wörter davor auseinandergehen.

package main

import "strings"

// zerlege trennt Befehl und Ausgabeform.
//
// **`--json` darf überall stehen.** Wer es hinten anhängt, meint dasselbe wie
// wer es vorn schreibt; eine Kommandozeile, die darauf besteht, prüft die
// Gewohnheit statt der Absicht.
//
// **Die weiteren Argumente kommen in der Reihenfolge**, in der sie standen —
// eine Kommandozeile, die sie umsortierte, machte `feld a b` und `feld b a` zu
// demselben Aufruf.
// **`--guete=<name>` trägt seinen Wert im selben Wort.** Ein Flag mit
// nachfolgendem Wert bräuchte einen Vorgriff, und der macht aus dieser
// Schleife einen Zustand — für den einzigen Wert, den diese Kommandozeile
// kennt.
// **`--profil=<kennung>` gilt für genau diesen Aufruf** — Werkstück 3 von
// Phase 4.12 (#1351). Der gewählte Stand überlebt sonst den Aufruf, weil er im
// Bestand liegt; ein Skript, das sich davon unabhängig machen will, sagt es
// hier und verstellt dabei **nichts**.
//
// # Warum ein Verbund und keine fünf Rückgabewerte
//
// **Weil der fünfte der war, bei dem es kippte.** Vier namenlose Werte an
// einer Aufrufstelle liest man noch; fünf nicht mehr, und jeder weitere
// Schalter zwänge jeden Rufer zu einer Änderung.
type zeile struct {
	befehl  string
	rest    []string
	alsJSON bool
	guete   string

	// profil ist die Kennung aus `--profil=` — leer heißt: der Zeiger gilt.
	profil string

	// erwartet ist die Erwartung für `felder-setzen` — JSON wie aus
	// `felder-vorschau --json` (U6, #1417); leer heißt: ohne Vergleich.
	erwartet string
}

func zerlege(argumente []string) zeile {
	gelesen := zeile{rest: []string{}}

	for _, argument := range argumente {
		switch {
		case argument == "--json":
			gelesen.alsJSON = true
		case strings.HasPrefix(argument, guetevorsatz):
			gelesen.guete = strings.TrimPrefix(argument, guetevorsatz)
		case strings.HasPrefix(argument, profilvorsatz):
			gelesen.profil = strings.TrimPrefix(argument, profilvorsatz)
		case strings.HasPrefix(argument, erwartetvorsatz):
			gelesen.erwartet = strings.TrimPrefix(argument, erwartetvorsatz)
		case gelesen.befehl == "":
			gelesen.befehl = argument
		default:
			gelesen.rest = append(gelesen.rest, argument)
		}
	}

	return gelesen
}

// guetevorsatz ist der Vorsatz, an dem die Güte erkannt wird.
const guetevorsatz = "--guete="

// profilvorsatz ist der Vorsatz, an dem das Profil erkannt wird.
const (
	profilvorsatz   = "--profil="
	erwartetvorsatz = "--erwartet="
)
