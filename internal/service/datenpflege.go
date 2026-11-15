package service

// **Analysieren, zeigen, abgleichen, sichern, ausführen, nachsehen** —
// Werkstück O8 von Phase 5.2
// (#1432).
//
// # Die Reihenfolge ist die Zusage
//
//  1. **Die Analyse ist rein lesend.** Wer nachsieht, ändert nichts — und
//     deshalb darf sie jede Fläche, jederzeit.
//  2. **Die Vorschau trägt eine Frist und je Maßnahme eine Revision.** Sie
//     beschreibt einen Augenblick, und beides sagt, wie lange.
//  3. **Abgeglichen wird gegen den Plan, den ein Mensch gesehen hat.** Was
//     sich seither geändert hat, stoppt — die betroffene Maßnahme, nicht die
//     anderen.
//  4. **Gesichert wird vor dem ersten Schreiben.** Nicht danach, und nicht
//     „falls etwas schiefgeht": eine Sicherung, die erst entsteht, wenn schon
//     gelöscht ist, sichert den Schaden.
//  5. **Ausgeführt wird je Maßnahme in ihrer eigenen Transaktion** — mit einer
//     zweiten Revisionsprüfung darin.
//  6. **Nachgesehen wird danach noch einmal.** Die Bilanz nennt beide Stände;
//     ohne den zweiten wäre „erledigt" eine Behauptung.

import (
	"context"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Pflegeleser sieht nach, was im Bestand herumliegt.
//
// **Die Schnittstelle steht hier, weil dieser Anwendungsfall sie braucht**
// (§2) — erfüllt von `storage.Pflegeanalyse`.
type Pflegeleser func(ctx context.Context, jetzt time.Time) ([]domain.Pflegebefund, error)

// Pflegeschreiber räumt die gewählten Stellen weg.
type Pflegeschreiber func(
	ctx context.Context, jetzt time.Time, revisionen map[string]string,
) ([]domain.Pflegezeile, error)

// Pflegesicherung schreibt den Bestand weg und sagt wohin.
//
// **Sie gibt den Ort zurück** — eine Sicherung, von der niemand weiß, wo sie
// liegt, ist keine.
type Pflegesicherung func(ctx context.Context) (string, error)

// Datenpflege führt den Weg.
type Datenpflege struct {
	lesen     Pflegeleser
	schreiben Pflegeschreiber
	sichern   Pflegesicherung
	uhr       func() time.Time
}

// NeueDatenpflege baut sie.
func NeueDatenpflege(
	lesen Pflegeleser, schreiben Pflegeschreiber, sichern Pflegesicherung,
) *Datenpflege {
	return &Datenpflege{lesen: lesen, schreiben: schreiben, sichern: sichern, uhr: time.Now}
}

// MitUhr setzt die Uhr — für Läufe, die die Frist prüfen.
func (d *Datenpflege) MitUhr(uhr func() time.Time) { d.uhr = uhr }

// Vorschau sieht nach — **und ändert nichts**.
func (d *Datenpflege) Vorschau(ctx context.Context) (domain.Massnahmenplan, error) {
	jetzt := d.uhr()

	befunde, fehler := d.lesen(ctx, jetzt)
	if fehler != nil {
		return domain.Massnahmenplan{}, fehler
	}

	return domain.Massnahmenplan{
		Erstellt: jetzt, Gueltigbis: jetzt.Add(domain.Pflegefrist), Befunde: befunde,
	}, nil
}

// Pflegebilanz ist, was aus einer Pflege wurde.
//
// **Beide Stände stehen darin.** „Drei erledigt" ist keine Auskunft über den
// Bestand; „vorher sieben Befunde, nachher vier" ist eine.
type Pflegebilanz struct {
	// Gesichert ist der Ort der Sicherung, die vor dem Schreiben entstand.
	Gesichert string

	Zeilen []domain.Pflegezeile

	// Vorher und Nachher sind die Befunde vor und nach der Pflege.
	Vorher  []domain.Pflegebefund
	Nachher []domain.Pflegebefund
}

// Pflegen führt die gewählten Maßnahmen aus.
//
// **Die Wahl kommt vom Aufrufer, samt der Revision, die er gesehen hat.** Sie
// hier neu zu holen hieße, gegen etwas abzugleichen, das niemand angeschaut
// hat — und `erstellt` sagt, wann er hingesehen hat.
func (d *Datenpflege) Pflegen(
	ctx context.Context, erstellt time.Time, gewaehlt []domain.Pflegewahl,
) (Pflegebilanz, error) {
	jetzt := d.uhr()

	frisch, lesefehler := d.lesen(ctx, jetzt)
	if lesefehler != nil {
		return Pflegebilanz{}, lesefehler
	}

	zeilen, abgleichfehler := domain.Pflegeabgleich(erstellt, frisch, gewaehlt, jetzt)
	if abgleichfehler != nil {
		return Pflegebilanz{}, abgleichfehler
	}

	bilanz := Pflegebilanz{Zeilen: zeilen, Vorher: frisch, Nachher: frisch}

	revisionen := bereite(zeilen, frisch)
	if len(revisionen) == 0 {
		// **Nichts zu tun heißt: nicht sichern.** Eine Sicherung ohne
		// Schreibvorgang wäre eine Datei, die niemand erklärt.
		return bilanz, nil
	}

	ort, sicherungsfehler := d.sichern(ctx)
	if sicherungsfehler != nil {
		return Pflegebilanz{}, fmt.Errorf(
			"vor der Pflege sichern: %w", sicherungsfehler)
	}

	bilanz.Gesichert = ort

	gelaufen, schreibfehler := d.schreiben(ctx, jetzt, revisionen)
	if schreibfehler != nil {
		return Pflegebilanz{}, schreibfehler
	}

	bilanz.Zeilen = zusammenlegen(zeilen, gelaufen)

	// **Nachgesehen wird auf jeden Fall** — auch wenn eine Maßnahme
	// scheiterte. Ohne den zweiten Stand wäre „erledigt" eine Behauptung.
	nachher, nachfehler := d.lesen(ctx, jetzt)
	if nachfehler != nil {
		return bilanz, fmt.Errorf(
			"die Pflege ist gelaufen, aber der Bestand ließ sich danach nicht "+
				"noch einmal durchsehen: %w", nachfehler)
	}

	bilanz.Nachher = nachher

	return bilanz, nil
}

// bereite sammelt die Revisionen der Maßnahmen, die laufen dürfen.
func bereite(zeilen []domain.Pflegezeile, frisch []domain.Pflegebefund) map[string]string {
	revisionen := map[string]string{}

	nachArt := map[string]domain.Pflegebefund{}
	for _, befund := range frisch {
		nachArt[befund.Art] = befund
	}

	for _, zeile := range zeilen {
		if zeile.Ausgang == domain.PflegeBereit {
			revisionen[zeile.Art] = nachArt[zeile.Art].Revision
		}
	}

	return revisionen
}

// zusammenlegen ersetzt die bereiten Zeilen durch ihren wirklichen Ausgang.
//
// **Die anderen bleiben stehen.** Ein Konflikt aus dem Abgleich ist genauso
// ein Ergebnis wie eine erledigte Maßnahme; ihn aus der Bilanz zu nehmen
// hieße, dem Aufrufer die Hälfte zu verschweigen.
func zusammenlegen(abgeglichen, gelaufen []domain.Pflegezeile) []domain.Pflegezeile {
	nachArt := map[string]domain.Pflegezeile{}
	for _, zeile := range gelaufen {
		nachArt[zeile.Art] = zeile
	}

	zusammen := make([]domain.Pflegezeile, 0, len(abgeglichen))

	for _, zeile := range abgeglichen {
		if wirklich, gelaufen := nachArt[zeile.Art]; gelaufen {
			zusammen = append(zusammen, wirklich)

			continue
		}

		zusammen = append(zusammen, zeile)
	}

	return zusammen
}
