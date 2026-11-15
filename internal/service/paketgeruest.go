package service

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"

// **Include und Exportvorlage aus einem Gerüst** — FMT-02.AK-01
// (#1691).
//
// # Was ein Gerüst tragen muss — aus MSE2
//
//   - **Ein Include ist nur der Paketkopf** (`IMPLEMENT_REFLECTION(IncludePackage)`
//     mit `REFLECT_BASE(Packaged)`, `src/util/io/package.cpp` Zeilen 656–658,
//     Fassung f9d9356 vom 2021-07-06). Was es einbindet, schreibt jemand
//     danach in weitere Dateien daneben.
//   - **Eine Exportvorlage muss `game:` tragen** — `ExportTemplate::validate`
//     wirft sonst „no game specified" — **und hängt von ihrem Spiel ab**
//     (`requireDependency(game)`, `src/data/export_template.cpp` Zeilen
//     31–38). Ihr Skript ist das Einzige, was sie tut; das Gerüst gibt je
//     Karte ihren Namen aus, wie die mitgelieferte `probe-liste` in knapp.
//
// # Der Name
//
// Gezählt über `full-magic-pack` am 2026-09-29: eine Exportvorlage heißt wie
// ein Stil `<spiel>-<name>` (`magic-mws`, `magic-forum`, …); ein Include
// trägt seinen Namen frei. Das Verzeichnis folgt dem ([stilverzeichnis]).

// anlegbar sind die Arten, die sich aus einem Gerüst anlegen lassen. Ein
// Symbolfont und ein Sprachpaket wären leer nichts, und niemand füllt sie von
// Hand (`docs/FAEHIGKEITEN.md`).
var anlegbar = map[string]bool{
	"mse-game": true, "mse-style": true, "mse-include": true, "mse-export-template": true,
}

// brauchtSpiel sagt, ob eine Art zu einem Spiel gehört.
func brauchtSpiel(art string) bool {
	return art == "mse-style" || art == "mse-export-template"
}

// includebaum baut die `include`-Datei eines neuen Includes.
func includebaum(wunsch Neuanlage) *msepaket.Baum {
	return &msepaket.Baum{
		Datei: "include",
		Wurzel: []*msepaket.Knoten{
			saatzeile("mse version", "2.0.0"),
			saatzeile("short name", wunsch.Name),
			saatzeile("full name", wunsch.Vollname),
			saatzeile("version", heute()),
		},
		MitBOM: true,
	}
}

// exportbaum baut die `export-template`-Datei einer neuen Exportvorlage.
func exportbaum(wunsch Neuanlage) *msepaket.Baum {
	return &msepaket.Baum{
		Datei: "export-template",
		Wurzel: []*msepaket.Knoten{
			saatzeile("mse version", "2.0.0"),
			saatzeile("short name", wunsch.Name),
			saatzeile("full name", wunsch.Vollname),
			saatzeile("version", heute()),
			saatzeile("game", wunsch.Spiel),
			mitKindern("depends on",
				saatzeile("package", wunsch.Spiel+".mse-game"),
			),
			saatzeile("file type", "*.txt|*.txt|*.*|*.*"),
			saatzeile("create directory", "false"),
			leerzeile(),
			saatzeile("script", `to_text(for each card in cards do card.name + "\n")`),
		},
		MitBOM: true,
	}
}
