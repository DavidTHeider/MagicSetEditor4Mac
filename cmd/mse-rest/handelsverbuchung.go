package main

// **Die Verbuchung im Dienst** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// **Eine Naht und keine Fachlogik**: welcher Schritt von hier aus geht,
// entscheidet der Vertrag; hier wird übersetzt — Feld für Feld und nicht über
// `Marshal`.

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// verbuchungsweg reicht die Verbuchung an den Dienst.
type verbuchungsweg struct{ db *sql.DB }

func (v verbuchungsweg) Handelsschritt(
	ctx context.Context, vorgang string, wunsch rest.Schrittwunsch,
) (rest.Handelsstand, error) {
	flaeche := app.PeerflaecheUeber(v.db)
	schritt, wer, freigabe := wunsch.Schritt, wunsch.Wer, wunsch.Freigabe

	// **Ein unbekannter Schritt ist eine Beanstandung und kein 500.**
	// Was es gibt, steht in der Zustandsmatrix; wer etwas anderes schickt,
	// hat sich vertan und soll es lesen können. **Vor der Verzweigung**,
	// denn sonst käme „das gibt es nicht" als Konflikt aus der Leitung — und
	// ein Klient wiederholte, bis der Zustand passt.
	if !domain.IstHandelsschritt(domain.Handelsschritt(schritt)) {
		return rest.Handelsstand{}, fmt.Errorf(
			"%w: %q ist kein Schritt — es gibt %v",
			domain.ErrValidation, schritt, domain.Handelsschritte())
	}

	// **Die neun Schritte vor dem Abschluss gehen über die Leitung** (T1,
	// #1466, angeschlossen mit #1637). Sie bewegen den Zustand und weder
	// Bestand noch Beleg — und brauchen deshalb keine Freigabe.
	if !domain.IstAbschlussschritt(domain.Handelsschritt(schritt)) {
		return leitungsschritt(ctx, flaeche, vorgang, schritt, wer, wunsch.Gesehen)
	}

	// **Nach der Einigung gibt es keine neue Revision** (#2496): ein
	// Gegenangebot geht nur aus „angeboten" und „in Verhandlung". Ein
	// Abdruck hier wird abgewiesen statt still übergangen.
	if wunsch.Gesehen != "" {
		return rest.Handelsstand{}, fmt.Errorf(
			"%w: %q kennt kein „gesehen“ — nach der Einigung ändert sich die "+
				"Revision nicht mehr", domain.ErrValidation, schritt)
	}

	var antwort app.VerbuchungAntwort

	switch schritt {
	case "verbuchen":
		antwort = flaeche.Handelsverbuchung(ctx, vorgang, freigabe)

	case "abgleichen":
		antwort = flaeche.Handelsabgleich(ctx, vorgang)

	case "bestaetigen":
		antwort = flaeche.Handelsbestaetigung(ctx, vorgang, wer, freigabe)

	}

	return alsHandelsstand(antwort)
}

// Handelsbeleg gibt Stand und Beleg, ohne einen Schritt zu gehen.
func (v verbuchungsweg) Handelsbeleg(
	ctx context.Context, vorgang string,
) (rest.Handelsstand, error) {
	return alsHandelsstand(app.PeerflaecheUeber(v.db).Handelsbeleg(ctx, vorgang))
}

// alsHandelsstand übersetzt eine Antwort der Verbuchung — nach einem
// Schritt wie ohne.
func alsHandelsstand(antwort app.VerbuchungAntwort) (rest.Handelsstand, error) {
	if antwort.Fehler != nil {
		return rest.Handelsstand{}, antwort.Fehler.AlsFehler()
	}

	stand := rest.Handelsstand{
		Vorgang: antwort.Vorgang, Zustand: antwort.Zustand,
	}

	if antwort.HatBeleg {
		stand.Abdruck = antwort.Beleg.Abdruck
		stand.Unterschrift = antwort.Beleg.Unterschrift
		stand.Beidseitig = antwort.Beleg.Beidseitig
		stand.Gebucht = antwort.Beleg.Gebucht
		stand.Gegenzeichner = antwort.Beleg.Gegenzeichner
		stand.Gegengezeichnet = antwort.Beleg.Gegengezeichnet

		if befund := antwort.Beleg.Befund; befund != nil {
			stand.Befund = &rest.Gegenbelegbefund{
				Von: befund.Von, Abdruck: befund.Abdruck, Mangel: befund.Mangel,
				Grund: befund.Grund, Empfangen: befund.Empfangen,
			}
		}
	}

	return stand, nil
}

func (v verbuchungsweg) Handelsjournal(
	ctx context.Context, vorgang string,
) ([]rest.Journalzeile, error) {
	antwort := app.PeerflaecheUeber(v.db).Handelsjournal(ctx, vorgang)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	zeilen := make([]rest.Journalzeile, 0, len(antwort.Zeilen))

	for _, zeile := range antwort.Zeilen {
		zeilen = append(zeilen, rest.Journalzeile{
			ID: zeile.ID, Art: zeile.Art, Anlass: zeile.Anlass,
			Vorgang: zeile.Vorgang, Bezug: zeile.Bezug, Grund: zeile.Grund,
			Printing: zeile.Printing, Bezeichnung: zeile.Bezeichnung,
			Menge:   zeile.Menge,
			Zustand: zeile.Zustand, Sprache: zeile.Sprache,
			Finish: zeile.Finish, ExemplarID: zeile.ExemplarID,
			Lagerort: zeile.Lagerort, Notiz: zeile.Notiz,
			Gebucht: zeile.Gebucht,
		})
	}

	return zeilen, nil
}

// leitungsschritt geht einen Schritt vor dem Abschluss.
//
// **Ohne Beleg**, und das ist keine Auslassung: einen Beleg gibt es erst,
// wenn verbucht ist. Wer hier einen leeren zurückgäbe, sähe aus wie ein
// Abschluss ohne Unterschrift.
func leitungsschritt(
	ctx context.Context, flaeche *app.App, vorgang, schritt, wer, gesehen string,
) (rest.Handelsstand, error) {
	antwort := flaeche.HandelSchritt(ctx, vorgang, schritt, wer, gesehen)
	if antwort.Fehler != nil {
		return rest.Handelsstand{}, antwort.Fehler.AlsFehler()
	}

	return rest.Handelsstand{
		Vorgang: antwort.Vorgang.Vorgang,
		Zustand: antwort.Vorgang.Zustand,
		Abdruck: antwort.Vorgang.Abdruck,
	}, nil
}
