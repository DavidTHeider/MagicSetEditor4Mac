package domain

import "strings"

// Paket ist ein **installiertes** Vorlagenpaket.
//
// # Warum es das hier gibt und nicht nur `msepaket.Paket`
//
// `msepaket.Paket` ist ein **geöffnetes** Paket: es hält einen Dateibaum, einen
// Schließer und den Rohkopf. Was der Katalog führt, ist etwas anderes — der
// Stand auf der Platte, mit Herkunft, ohne offene Datei. Wer den geöffneten Typ
// in die Datenbank reichte, hielte für jede Katalogzeile ein Archiv offen; das
// ist der Fehler, der in Phase 6 zu `too many open files` beim 31. Set geführt
// hat.
//
// **Die Art bleibt eine Zeichenkette.** `domain` importiert nichts aus dem
// Projekt (§2), also auch nicht `msepaket.Art`. Geprüft wird sie hier gegen die
// sieben Endungen aus `docs/FORMATE.md` — eine achte Art wäre ein Befund und
// keine stille Aufnahme.
type Paket struct {
	Art     string
	Name    string
	Version string

	// Kurzname ist `short name:` — was zur Wahl dasteht, wenn eine
	// `package choice` ihre Liste baut
	// (#632).
	//
	// **Leer ist erlaubt und heißt: keiner angegeben.** MSE fällt dann auf den
	// Dateinamen ohne Endung zurück (`Packaged::validate`,
	// `src/util/io/package.cpp`, Zeilen 622–626, Fassung f9d9356 vom
	// 2021-07-06). Der Rückfall steht dort, wo die Liste gebaut wird — eine
	// Spalte, die den Dateinamen wiederholte, liefe beim Umbenennen
	// auseinander.
	Kurzname string

	// Pfad ist der Ort auf der Platte, an dem dieses Paket liegt.
	Pfad string

	// Herkunft sagt, **woher** es kommt — eine Adresse, ein Verzeichnis, eine
	// MSE2-Installation.
	//
	// **Leer ist erlaubt und heißt: von hier.** Die mitgelieferte Saat hat
	// keine fremde Herkunft, und ein Paket, das jemand von Hand ins Verzeichnis
	// gelegt hat, auch nicht. Wer nicht weiß, woher eine Vorlage kommt, kann
	// ihre Rechtslage nicht einschätzen (#56, §11) — deshalb wird sie geführt,
	// wo sie bekannt ist, und nicht erfunden, wo sie es nicht ist.
	Herkunft string

	// Pruefsumme ist, was beim Bezug herauskam — `sha256:…`.
	//
	// **Leer ist erlaubt und heißt: nicht bezogen.** Ein Paket, das jemand von
	// Hand ins Verzeichnis gelegt hat, hat keine; eine zu erfinden wäre eine
	// Falschangabe (#56).
	Pruefsumme string

	// Braucht sind die Pakete, die dieses voraussetzt — `depends on:`.
	Braucht []Bedarf

	// Abdruck ist der Stand der Dateien beim Einbinden oder Beziehen —
	// DOM-01.AK-01 (#1696). **Leer heißt: keiner festgehalten** (Saat, eigene
	// Pakete); dann gibt es nichts, wovon es abweichen könnte.
	Abdruck string

	// Geaendert sagt, dass die Dateien heute nicht mehr dem Abdruck
	// entsprechen — gerechnet beim Lesen der Liste, nicht gespeichert.
	Geaendert bool

	// Gruppe ist `installer group:` — der Pfad, unter dem dieses Paket in der
	// Verwaltung steht, mit `/` als Trenner (#244).
	//
	// **Leer ist erlaubt und heißt: keine Gruppe.** Über `full-magic-pack`
	// tragen 416 von 526 Paketen eine; die übrigen bekommen in der Oberfläche
	// eine eigene Sammelgruppe, so wie MSE2 sie „custom" nennt.
	Gruppe string

	// Rang ist `position hint:` — die Ordnung innerhalb einer Gruppe.
	//
	// **Groß heißt hinten.** Ohne Angabe steht [msepaket.RangOhneHinweis] da,
	// also eine Million; ein Paket ohne Hinweis steht damit hinter jedem mit.
	Rang int
}

// Bedarf ist eine Abhängigkeit eines Pakets.
type Bedarf struct {
	Paket   string
	Version string
}

// paketarten sind die sieben Endungen aus `docs/FORMATE.md`.
//
// **Als Menge und nicht als Präfixprüfung auf `mse-`**: hier wird ein
// **Katalogeintrag** beurteilt, und ein `mse-erfunden` im Bestand wäre eine
// Art, die kein Leser dieses Projekts öffnen kann. `msepaket.istPaketname`
// prüft bewusst lockerer — es sucht Wurzeln und darf eine achte Art nicht
// übersehen. Die beiden Regeln sind verschieden, weil ihre Fragen es sind.
var paketarten = map[string]bool{
	"mse-set":             true,
	"mse-game":            true,
	"mse-style":           true,
	"mse-symbol-font":     true,
	"mse-include":         true,
	"mse-locale":          true,
	"mse-export-template": true,
}

// Validate prüft, was ein Katalogeintrag mindestens tragen muss.
func (p Paket) Validate() error {
	if strings.TrimSpace(p.Art) == "" {
		return NewFieldError("art", SchluesselPflichtfeld, "Ein Paket nennt keine Art.")
	}

	if !paketarten[p.Art] {
		return NewFieldError("art", SchluesselValidierung,
			"%q ist keine der sieben Paketarten.", p.Art)
	}

	if strings.TrimSpace(p.Name) == "" {
		return NewFieldError("name", SchluesselPflichtfeld, "Ein Paket braucht einen Namen.")
	}

	if strings.TrimSpace(p.Pfad) == "" {
		return NewFieldError("pfad", SchluesselPflichtfeld,
			"Ein installiertes Paket braucht einen Ort.")
	}

	for _, bedarf := range p.Braucht {
		if strings.TrimSpace(bedarf.Paket) == "" {
			return NewFieldError("braucht", SchluesselPflichtfeld,
				"Eine Abhängigkeit von %q nennt kein Paket.", p.Name)
		}
	}

	return nil
}
