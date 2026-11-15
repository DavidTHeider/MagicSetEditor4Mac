// Die Gestalten, in denen ein Scryfall-Vorschlag das Fenster erreicht.
//
// **Eine eigene Datei, weil `scryfall.go` mit ihr über 400 Zeilen ginge** (§2).
// Der Schnitt folgt der Frage: dort steht, **woher** die Angaben kommen —
// Klient, Drosseln, Zwischenspeicher, Rückfall —, hier, **in welcher Gestalt**
// sie hinausgehen.

package app

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"

// ScryfallvorschlagDTO ist ein Vorschlag, wie das Fenster ihn sieht.
type ScryfallvorschlagDTO struct {
	Aenderungen []FeldvorschlagDTO `json:"aenderungen"`
	Befunde     []string           `json:"befunde"`
	Herkunft    string             `json:"herkunft"`

	// AlterStunden sagt, wie alt die Antwort ist, aus der der Vorschlag stammt.
	//
	// **In Stunden und als Zahl.** Eine `time.Duration` käme als Nanosekunden
	// an, und die Seite müsste rechnen; ein fertiger Satz käme auf Deutsch an
	// und wäre nicht übersetzbar (§9).
	AlterStunden int `json:"alterStunden"`

	// Stand bindet den Vorschlag an die Kartenwerte, gegen die verglichen
	// wurde (#1241).
	//
	// **Das Fenster gibt ihn beim Übernehmen zurück**; gilt er nicht mehr,
	// wird abgewiesen. Ohne ihn übernähme der Knopf **etwas anderes als das
	// Gezeigte**.
	Stand string `json:"stand"`
}

// FeldvorschlagDTO ist eine einzelne Änderung.
type FeldvorschlagDTO struct {
	Feld string `json:"feld"`
	Alt  string `json:"alt"`
	Neu  string `json:"neu"`
}

// alsVorschlagDTO übersetzt einen Vorschlag für das Fenster.
//
// **Die leeren Listen bleiben Listen.** Ein `nil`-Slice würde zu `null`, und
// eine Seite, die darüber schleift, bekäme einen Fehler statt keiner Runde —
// die Falle aus AGENTS.md §3.
func alsVorschlagDTO(vorschlag *service.Vorschlag) ScryfallvorschlagDTO {
	aenderungen := make([]FeldvorschlagDTO, 0, len(vorschlag.Aenderungen))
	for _, eine := range vorschlag.Aenderungen {
		aenderungen = append(aenderungen,
			FeldvorschlagDTO{Feld: eine.Feld, Alt: eine.Alt, Neu: eine.Neu})
	}

	befunde := make([]string, 0, len(vorschlag.Befunde))
	for _, befund := range vorschlag.Befunde {
		befunde = append(befunde, befund.String())
	}

	return ScryfallvorschlagDTO{
		Aenderungen:  aenderungen,
		Befunde:      befunde,
		Herkunft:     vorschlag.Herkunft,
		AlterStunden: int(vorschlag.Alter.Hours()),
		Stand:        vorschlag.Stand,
	}
}
