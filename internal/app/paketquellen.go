package app

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"

// **Woher ein Stück Paketinhalt kommt** — die fünf Nähte, die dieselbe Frage
// beantworten.
//
// **Eine eigene Datei**, weil `anwendung.go` mit ihnen über die Zeilengrenze
// aus §2 ginge. Der Schnitt folgt der Frage: dort steht, **was diese
// Anwendung verdrahtet** — hier, **woher der Inhalt eines Pakets kommt**.
//
// **Eingebettet und nicht benannt**: die Felder werden überall als
// `a.symbolzeichnungen` gelesen, und ein Zwischenname änderte hundert
// Fundstellen, ohne etwas zu sagen.
type paketquellen struct {
	// symbolzeichnungen liest die Gestalt einer `.mse-symbol` — für
	// `symbol_variation` in der Ausfuhr
	// (#660).
	symbolzeichnungen service.Symbolquelle

	// setwerte ist die Naht zu den Setangaben — Titel, Setcode, Copyright
	// (#567).
	setwerte Setwertquelle

	// schluesselwoerter ist die Naht zu den Schlüsselwörtern eines Sets
	// (#1062).
	schluesselwoerter Schluesselwortquelle

	// pack zieht einen Booster aus dem Set
	// (#1072).
	pack Packquelle

	// symbolbibliothek führt die Symbole eines Profils, die sich in jedes Set
	// einsetzen lassen (DOM-02,
	// #1697).
	symbolbibliothek Symbolbibliotheksfuehrer

	// sprachpakete führt die eigenen Übersetzungen (LOC-02,
	// #1688).
	sprachpakete *service.Sprachpakete
}
