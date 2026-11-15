package app

// **Wie eine Exportvorlage in der Oberfläche aussieht** — die Übersetzung von
// [msepaket.Exportvorlage] in ihre DTOs.
//
// Eigene Datei, weil `vorlagenausfuhr.go` mit ihr über 400 Zeilen ginge (§2).
// Der Schnitt folgt der Frage: dort steht, **was** eine Ausfuhr tut, hier,
// **wie** eine Vorlage dem Fenster erscheint.

import (
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// zuVorlagenDTO übersetzt die Vorlagen für die Oberfläche.
func zuVorlagenDTO(vorlagen []msepaket.Exportvorlage) []ExportvorlageDTO {
	umgesetzt := make([]ExportvorlageDTO, 0, len(vorlagen))

	// Über den Index statt über den Wert: eine `Exportvorlage` wiegt 128 Byte,
	// und `gocritic` zählt jede Kopie in einer Schleife nach.
	for i := range vorlagen {
		umgesetzt = append(umgesetzt, ExportvorlageDTO{
			Paket:    vorlagen[i].Paket,
			Name:     namenVon(&vorlagen[i]),
			Dateityp: vorlagen[i].Dateityp,
			Optionen: zuOptionenDTO(vorlagen[i].Optionen),
		})
	}

	return umgesetzt
}

// namenVon gibt den Namen, den die Oberfläche zeigt.
//
// **Der volle Name, sonst der kurze, sonst das Paket.** Alle drei stehen in der
// Datei, und keiner ist zugesichert; ein leerer Eintrag in einer Auswahlliste
// wäre ein Knopf ohne Aufschrift. **Nicht übersetzt** — er kommt aus dem Paket
// (§9).
func namenVon(vorlage *msepaket.Exportvorlage) string {
	for _, name := range []string{vorlage.VollerName, vorlage.KurzerName, vorlage.Paket} {
		if strings.TrimSpace(name) != "" {
			return name
		}
	}

	return vorlage.Paket
}

// zuOptionenDTO übersetzt die Wahlmöglichkeiten.
func zuOptionenDTO(optionen []msepaket.Exportoption) []ExportoptionDTO {
	umgesetzt := make([]ExportoptionDTO, 0, len(optionen))

	for i := range optionen {
		option := &optionen[i]

		auswahl := option.Auswahl
		if auswahl == nil {
			// **Nie nil.** Ein `null` käme im Frontend als fehlende Liste an,
			// und `.map` darauf wirft — der Fehler zwischen den Prüfebenen, vor
			// dem §3 warnt.
			auswahl = []string{}
		}

		umgesetzt = append(umgesetzt, ExportoptionDTO{
			Art:          option.Art,
			Name:         option.Name,
			Beschreibung: option.Beschreibung,
			Vorgabe:      option.Anfangswert(),
			Auswahl:      auswahl,
		})
	}

	return umgesetzt
}

// meldungen macht aus Befunden ihre Sätze.
func meldungen(befunde []domain.Befund) []string {
	umgesetzt := make([]string, 0, len(befunde))
	for _, befund := range befunde {
		umgesetzt = append(umgesetzt, befund.String())
	}

	return umgesetzt
}

// erstesMuster liest das erste Dateimuster aus einem `file type:`.
//
// MSE schreibt ihn als `*.xml|*.xml|*.*|*.*` — abwechselnd Beschriftung und
// Muster, mehrfach. Genommen wird das **erste** Muster; die weiteren sind bei
// jeder der neunzehn Vorlagen `*.*`, und ein Dialog, der „alle Dateien"
// vorschlägt, hilft niemandem.
//
// **Ohne Angabe `*.*`** — das ist kein Vorschlag, sondern das Eingeständnis,
// keinen zu haben. Vier der Vorlagen führen `file type:` gar nicht.
func erstesMuster(dateityp string) string {
	for teil := range strings.SplitSeq(dateityp, "|") {
		sauber := strings.TrimSpace(teil)
		if strings.HasPrefix(sauber, "*.") && sauber != "*.*" {
			return sauber
		}
	}

	return "*.*"
}

// endungVon macht aus einem Muster die Endung eines Vorschlagsnamens.
func endungVon(muster string) string {
	if muster == "*.*" {
		return ""
	}

	return strings.TrimPrefix(muster, "*")
}
