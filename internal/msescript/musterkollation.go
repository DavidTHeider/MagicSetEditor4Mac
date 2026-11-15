package msescript

import "strings"

// **Ein POSIX-Kollationselement benennt ein Zeichen — `[[.quotation-mark.]]`.**
//
// MSE übersetzt seine Muster mit `wxRegEx` im **erweiterten POSIX**-Modus, und
// dort steht `[[.name.]]` für das Zeichen mit diesem Namen. Weder Gos `regexp`
// noch `regexp2` kennen die Schreibweise; die erste meldet
// `[n-m] range in reverse order` — **96 Befunde** über die 190 Stile von
// `M15-Magic-Pack`, alle an derselben Zeile von `magic.mse-game`:
//
//	replace(match: "[[.quotation-mark.]]|”", replace: …)
//
// (#270)

// kollationselemente sind die Namen, die im Pack vorkommen.
//
// **Nur die gemessenen.** Eine vollständige Tabelle der POSIX-Namen wäre eine
// Tabelle ohne Prüfung; was hier steht, ist ausgezählt — und was fehlt,
// scheitert am Übersetzer statt still etwas anderes zu treffen (§8).
var kollationselemente = map[string]string{
	"quotation-mark": `"`,
}

// ohneKollationselemente ersetzt `[[.name.]]` durch sein Zeichen.
//
// **Ein unbekannter Name bleibt stehen.** Dann scheitert das Muster am
// Übersetzer, und die Meldung nennt es — eine erfundene Übersetzung machte aus
// einem Muster ein anderes, und niemand sähe es.
func ohneKollationselemente(quelle string) string {
	if !strings.Contains(quelle, "[[.") {
		return quelle
	}

	gebaut := quelle

	for name, zeichen := range kollationselemente {
		gebaut = strings.ReplaceAll(gebaut, "[[."+name+".]]", zeichen)
	}

	return gebaut
}
