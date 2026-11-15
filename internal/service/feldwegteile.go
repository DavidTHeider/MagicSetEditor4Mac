package service

// **Die Felder hinter einem Feld** — aus `feldweg.go` herausgelöst, als dort
// das Ergebnis des Schreibens hinzukam (#2535) und die Datei an die
// Zeilengrenze stieß (§2). Sie beantworten eine eigene Frage: welches Feld
// eines Spiels ein Name meint, und woraus ein berechnetes entsteht.

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Teilfelder gibt die Felder, aus denen ein Feld entsteht — oder nichts.
//
// **Wofür die Oberfläche das braucht:** ein Feld mit `save value: false` nimmt
// keine Eingabe entgegen. Sie bietet an seiner Stelle **seine Teile** an — und
// weiß erst dadurch, dass es welche gibt
// (#696).
//
// **Dieselbe Antwort wie beim Schreiben.** Wer die Frage zweimal beantwortete,
// bekäme irgendwann zwei Antworten: ein Feld, das die Oberfläche zerlegt, aber
// der Bestand ablehnt.
func Teilfelder(felder []domain.Field, form domain.Field) []string {
	if form.Gespeichert {
		return nil
	}

	// **Die schwache Frage, nicht die starke.** Ob sich ein Wert wieder
	// **spalten** lässt, hängt am Trenner; woraus er **entsteht**, nicht.
	// `type` und `pt` rechnen ihren Trenner aus und wären sonst Felder, die
	// dastehen und keine Eingabe annehmen.
	namen := msescript.VerbundfelderAus(form.Skript)
	if len(namen) == 0 {
		return nil
	}

	teile := make([]string, 0, len(namen))
	for _, name := range namen {
		teile = append(teile, feldname(felder, name))
	}

	return teile
}

// feldname macht aus `rule_text` das Feld `rule text`.
//
// **Über die Felder des Spiels und nicht über eine Ersetzung.** Ein Unterstrich
// wird bei MSE nicht in ein Leerzeichen übersetzt, sondern beim Vergleich
// übergangen ([msescript.GliedTrifft]) — und ein Feld, das wirklich einen
// Unterstrich im Namen trägt, gäbe es sonst zweimal. Findet sich keines, bleibt
// der Name stehen: dann legt der Bestand ihn so an, und der Wert ist wenigstens
// nicht verloren.
func feldname(felder []domain.Field, gesucht string) string {
	for _, form := range felder {
		if msescript.GliedTrifft(gesucht, form.Name) {
			return form.Name
		}
	}

	return gesucht
}

// form gibt die Feldbeschreibung zu einem Namen.
func form(felder []domain.Field, feld string) (domain.Field, bool) {
	for _, gefunden := range felder {
		if gefunden.Name == feld {
			return gefunden, true
		}
	}

	return domain.Field{}, false
}
