package app

// **Das sechste Ziel: die Verknüpfungen in `~/.local/bin`** — OPS03
// (#2349),
// Befund A-10 des Audits 0.9.0.
//
// Seit OPS-03 (#1720) legt *Kommandozeile einrichten* zwei Verknüpfungen an —
// `mse` und `mse-mcp` —, und was eine eigene ist, steht in `befehlsweg.go`.
// **Sie standen neben den Zielen**: „Alles widerrufen" ließ sie liegen, und
// wer das Programm danach in den Papierkorb legte, behielt zwei Verweise ins
// Leere. Hier werden sie ein Ziel wie die anderen — in der Liste, in „Alles
// einrichten" und „Alles widerrufen", an der Kommandozeile und in der
// Außerbetriebnahme. **Die Regeln bleiben, wo sie sind**; das Ziel liest
// über sie.

import (
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// befehlsziel ist der Name des sechsten Ziels.
const befehlsziel = "befehle"

// befehleZiel baut das Ziel der Verknüpfungen über Zuhause, Werkzeugort und
// `PATH` der Anmeldeshell.
func befehleZiel(heim, werkzeug, pfad string) service.Schnittstellenziel {
	verzeichnis := filepath.Join(heim, ".local", "bin")
	lies := func() Befehlsvorschau { return befehlsvorschau(verzeichnis, werkzeug, pfad) }

	zeilen := make([]string, 0, len(befehlsnamen))
	for _, name := range befehlsnamen {
		zeilen = append(zeilen, filepath.Join(verzeichnis, name)+" → "+filepath.Join(werkzeug, name))
	}

	return service.Schnittstellenziel{
		Name: befehlsziel, Art: service.ArtVerknuepfung, Datei: verzeichnis, Vorhanden: true,
		Verknuepfung: &service.Verknuepfungsziel{
			Stand: func() service.Schnittstellenstand { return befehlszielstand(lies()) },
			Einrichten: func() error {
				_, fehler := befehleInstallieren(lies())

				return fehler
			},
			Widerrufen: func() error {
				_, fehler := befehleEntfernen(lies())

				return fehler
			},
			Zeilen: zeilen,
		},
	}
}

// befehlszielstand faltet die Stände der Verknüpfungen in einen.
//
// **Gemischt heißt veraltet**: steht eine eigene neben einer fremden oder
// einer fehlenden, repariert Einrichten, was es darf, und Widerrufen nimmt
// die eigene — „fremd" hieße hier, die eigene bliebe liegen. **Fremd** ist
// das Ziel nur, wenn keine eigene da ist und keine fehlt.
func befehlszielstand(v Befehlsvorschau) service.Schnittstellenstand {
	gezaehlt := map[string]int{}
	for _, eintrag := range v.Eintraege {
		gezaehlt[eintrag.Stand]++
	}

	eigene := gezaehlt[befehlEigen] + gezaehlt[befehlVeraltet]

	switch {
	case eigene == 0 && gezaehlt[befehlFrei] == 0:
		return service.SchnittstelleFremd
	case eigene == 0:
		return service.SchnittstelleOffen
	case gezaehlt[befehlEigen] == len(v.Eintraege):
		return service.SchnittstelleEingerichtet
	}

	return service.SchnittstelleVeraltet
}
