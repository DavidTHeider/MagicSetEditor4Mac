package main

// **Geschrieben ist geschrieben — an der Kommandozeile** (§7, Regel 3;
// #1960,
// #1970).
//
// Scheitert nach einer geglückten Änderung nur das Nachlesen, endet der
// Befehl mit [CodeGut] und sagt es auf dem Fehlerstrom. **Nicht mit
// [CodeFachlich]**: ein Skript, das daraufhin wiederholte, liefe gegen den
// geschriebenen Stand. Mit `--json` steht der Grund zusätzlich unter
// `nachladefehler`.

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"

// nachladenMelden sagt, dass geschrieben ist und nur das Nachlesen scheiterte
// — und schweigt, wenn es nichts zu sagen gibt.
func nachladenMelden(an *schreiber, nachladen *app.Fehler) {
	if nachladen != nil {
		an.klage("geschrieben — nur das Nachlesen danach scheiterte: %v",
			nachladen.AlsFehler())
	}
}
