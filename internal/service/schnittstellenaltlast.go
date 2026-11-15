package service

// **Ein Dienst unter einem früheren Etikett** — Zweitblick über Phase 4 von
// 0.9.0, Befund 7: blieb der Kennungsumzug unvollständig oder lief er nie —
// die Kommandozeile führt ihn nicht aus —, stand ein LaunchAgent unter dem
// Etikett von vor 0.9.0, und weder das Band noch „Alles widerrufen" sah ihn.

import (
	"html"
	"path/filepath"
	"regexp"
)

// Altagent ist ein LaunchAgent dieser Anwendung unter einem früheren
// Etikett — dem von vor 0.9.0 (Vermerk 0191; welches, sagt `internal/app`).
//
// **Eigen ist er, wenn seine plist das Etikett und das Programm nennt** —
// dieselbe Prüfung wie für das heutige Etikett, und dazu das Programm: ein
// Etikett aus der Vorlage von Wails kann auch eine andere Anwendung tragen.
type Altagent struct {
	Datei, Schluessel, Programm string
	Widerrufbefehle             [][]string
}

// etikettfeld und programmfeld lesen das Etikett und das ausgeführte
// Programm **an ihrer Stelle**: unter `Label` und als erstes Argument von
// `ProgramArguments` — so schreiben 0.8.0 und diese Fassung die plist.
var (
	etikettfeld  = regexp.MustCompile(`<key>Label</key>\s*<string>([^<]*)</string>`)
	programmfeld = regexp.MustCompile(`<key>ProgramArguments</key>\s*<array>\s*<string>([^<]*)</string>`)
)

// AltlastEigen sagt, ob die plist eines früheren Etiketts von dieser
// Anwendung stammt.
//
// **Etikett und Programm an ihrer Stelle, und das Programm in einem eigenen
// Bündel** (Nachlauf über Phase 4, Befund 6): zwei Zeichenfolgen irgendwo in
// der Datei genügten — ein Etikett unter einem anderen Schlüssel, `/mse-rest`
// als Ausgabepfad, ein fremdes Programm dieses Namens. Danach würde der
// Dienst entladen und seine plist entfernt.
func AltlastEigen(alt Altagent) bool {
	inhalt, fehler := zielLesen(alt.Datei)
	if fehler != nil {
		return false
	}

	etikett, programm := etikettfeld.FindSubmatch(inhalt), programmfeld.FindSubmatch(inhalt)
	if len(etikett) != 2 || len(programm) != 2 || html.UnescapeString(string(etikett[1])) != alt.Schluessel {
		return false
	}

	pfad := html.UnescapeString(string(programm[1]))

	return filepath.Base(pfad) == alt.Programm && EigenesBuendel(pfad)
}

// eigeneAltlasten sind die Altlasten eines Ziels, die von hier stammen.
func eigeneAltlasten(ziel Schnittstellenziel) []Altagent {
	var eigene []Altagent

	for _, alt := range ziel.Altlasten {
		if AltlastEigen(alt) {
			eigene = append(eigene, alt)
		}
	}

	return eigene
}
