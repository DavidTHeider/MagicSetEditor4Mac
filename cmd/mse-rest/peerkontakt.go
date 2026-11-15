package main

// **Die bekannten Gegenüber im Dienst** — Werkstück N1 von Phase 5.4
// (#1454).
//
// **Eine Naht und keine Fachlogik**: ob eine Bestätigung gilt, entscheidet
// `internal/service`; hier wird übersetzt.

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// kontaktweg reicht die Kontakte an den Dienst.
type kontaktweg struct{ db *sql.DB }

func (k kontaktweg) Peerkontakte(
	ctx context.Context,
) (rest.Peerkontakte, error) {
	return alsRestkontakte(app.PeerflaecheUeber(k.db).Peerkontakte(ctx))
}

func (k kontaktweg) PeerkontaktAufnehmen(
	ctx context.Context, kennung, name string,
) (rest.Peerkontakte, error) {
	return alsRestkontakte(app.PeerflaecheUeber(k.db).
		PeerkontaktAufnehmen(ctx, kennung, name))
}

func (k kontaktweg) PeerkontaktBestaetigen(
	ctx context.Context, kennung, verglichen string,
) (rest.Peerkontakte, error) {
	return alsRestkontakte(app.PeerflaecheUeber(k.db).
		PeerkontaktBestaetigen(ctx, kennung, verglichen))
}

func (k kontaktweg) PeerkontaktSperren(
	ctx context.Context, kennung string,
) (rest.Peerkontakte, error) {
	return alsRestkontakte(app.PeerflaecheUeber(k.db).
		PeerkontaktSperren(ctx, kennung))
}

func (k kontaktweg) PeerkontaktEntsperren(
	ctx context.Context, kennung string,
) (rest.Peerkontakte, error) {
	return alsRestkontakte(app.PeerflaecheUeber(k.db).
		PeerkontaktEntsperren(ctx, kennung))
}

func (k kontaktweg) PeerzweckErlauben(
	ctx context.Context, kennung, zweck string,
) (rest.Peerkontakte, error) {
	return alsRestkontakte(app.PeerflaecheUeber(k.db).
		PeerzweckErlauben(ctx, kennung, zweck))
}

func (k kontaktweg) PeerzweckEntziehen(
	ctx context.Context, kennung, zweck string,
) (rest.Peerkontakte, error) {
	return alsRestkontakte(app.PeerflaecheUeber(k.db).
		PeerzweckEntziehen(ctx, kennung, zweck))
}

func (k kontaktweg) PeerkontaktAdresse(
	ctx context.Context, kennung, adresse string,
) (rest.Peerkontakte, error) {
	return alsRestkontakte(app.PeerflaecheUeber(k.db).
		PeerkontaktAdresse(ctx, kennung, adresse))
}

func (k kontaktweg) PeerkontaktNotiz(
	ctx context.Context, kennung, notiz string,
) (rest.Peerkontakte, error) {
	return alsRestkontakte(app.PeerflaecheUeber(k.db).
		PeerkontaktNotiz(ctx, kennung, notiz))
}

// alsRestkontakte macht aus der Antwort der Anwendung die des Dienstes.
func alsRestkontakte(
	antwort app.PeerkontakteAntwort,
) (rest.Peerkontakte, error) {
	if antwort.Fehler != nil {
		return rest.Peerkontakte{}, antwort.Fehler.AlsFehler()
	}

	kontakte := make([]rest.Peerkontakt, 0, len(antwort.Kontakte))

	for _, kontakt := range antwort.Kontakte {
		kontakte = append(kontakte, rest.Peerkontakt{
			Kennung:              kontakt.Kennung,
			Fingerabdruck:        kontakt.Fingerabdruck,
			SelbstgewaehlterName: kontakt.SelbstgewaehlterName,
			EigenerName:          kontakt.EigenerName,
			Stand:                kontakt.Stand,
			Gesehen:              kontakt.Gesehen,
			Rechte:               kontakt.Rechte,
			Adresse:              kontakt.Adresse,
			Erstmals:             kontakt.Erstmals,
			Notiz:                kontakt.Notiz,
			Lage:                 kontakt.Lage,
			Vergleichscode:       kontakt.Vergleichscode,
		})
	}

	dto := rest.Peerkontakte{Kontakte: kontakte, Zwecke: antwort.Zwecke}

	if gemeldet := antwort.Gemeldet; gemeldet != nil {
		dto.Gemeldet = &rest.Einloesemeldung{
			Stand: gemeldet.Stand, Grund: gemeldet.Grund, Meldung: gemeldet.Meldung,
		}
	}

	dto.Nachladefehler = alsNachladebefund(antwort.Nachladefehler)

	return dto, nil
}
