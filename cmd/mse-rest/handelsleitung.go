package main

// **Ein Handel im Dienst** —
// #1637.
//
// **Eine Naht und keine Fachlogik**: was ein Schritt bedeutet, entscheidet
// der Vertrag; hier wird übersetzt — Feld für Feld und nicht über `Marshal`.
// Derselbe Satz steht im Kopf von `handelsverbuchung.go`, und er hat
// denselben Grund: die Wails-DTO ist **innen**, diese Fläche **außen**
// ([0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)).

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// leitungsweg reicht die Schritte vor dem Abschluss an den Dienst.
type leitungsweg struct{ db *sql.DB }

func (l leitungsweg) HandelEroeffnen(
	ctx context.Context, wunsch rest.Handelswunsch,
) (rest.Handelsvorgang, error) {
	return alsRestvorgang(app.PeerflaecheUeber(l.db).HandelEroeffnen(ctx,
		app.Handelswunsch{
			Vorgang: wunsch.Vorgang, Art: wunsch.Art,
			Gegenueber: wunsch.Gegenueber,
			Positionen: alsAppPositionen(wunsch.Positionen),

			BewertungGeteilt: wunsch.BewertungGeteilt,
		}))
}

func (l leitungsweg) HandelVorschlagen(
	ctx context.Context, vorgang string, positionen []rest.Handelsposition,
) (rest.Handelsvorgang, error) {
	return alsRestvorgang(app.PeerflaecheUeber(l.db).HandelVorschlagen(
		ctx, vorgang, alsAppPositionen(positionen)))
}

// HandelProtokoll gibt die Zustandswechsel eines Vorgangs (TRADE-03).
func (l leitungsweg) HandelProtokoll(
	ctx context.Context, vorgang string,
) ([]rest.Handelsprotokollzeile, error) {
	antwort := app.PeerflaecheUeber(l.db).HandelProtokoll(ctx, vorgang)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	zeilen := make([]rest.Handelsprotokollzeile, 0, len(antwort.Zeilen))
	for _, zeile := range antwort.Zeilen {
		zeilen = append(zeilen, rest.Handelsprotokollzeile(zeile))
	}

	return zeilen, nil
}

func (l leitungsweg) HandelStand(
	ctx context.Context, vorgang string,
) (rest.Handelsvorgang, error) {
	return alsRestvorgang(
		app.PeerflaecheUeber(l.db).HandelStand(ctx, vorgang))
}

// HandelVorgaenge nennt die Handel mit einem Gegenüber (UX-08.AK-04).
func (l leitungsweg) HandelVorgaenge(
	ctx context.Context, gegenueber string,
) ([]rest.Handelsvorgang, error) {
	antwort := app.PeerflaecheUeber(l.db).HandelVorgaenge(ctx, gegenueber)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	vorgaenge := make([]rest.Handelsvorgang, 0, len(antwort.Vorgaenge))

	for _, vorgang := range antwort.Vorgaenge {
		vorgaenge = append(vorgaenge, restvorgangAus(vorgang))
	}

	return vorgaenge, nil
}

func (l leitungsweg) HandelSchicken(
	ctx context.Context, vorgang, empfaenger string,
) (string, error) {
	antwort := app.PeerflaecheUeber(l.db).HandelSchicken(
		ctx, vorgang, empfaenger)
	if antwort.Fehler != nil {
		return "", antwort.Fehler.AlsFehler()
	}

	return antwort.NachrichtID, nil
}

func (l leitungsweg) HandelPostbestaetigung(
	ctx context.Context, nachrichtID string,
) error {
	antwort := app.PeerflaecheUeber(l.db).HandelPostbestaetigung(
		ctx, nachrichtID)
	if antwort.Fehler != nil {
		return antwort.Fehler.AlsFehler()
	}

	return nil
}

func (l leitungsweg) HandelWiederaufnahme(
	ctx context.Context, vorgang, empfaenger string,
	seiner rest.Handelsfortschritt,
) (rest.Wiederaufnahme, error) {
	antwort := app.PeerflaecheUeber(l.db).HandelWiederaufnahme(
		ctx, vorgang, empfaenger, app.HandelsfortschrittDTO{
			Vorgang: vorgang, Revision: seiner.Revision,
			Abdruck: seiner.Abdruck, Folge: seiner.Folge,
		})
	if antwort.Fehler != nil {
		return rest.Wiederaufnahme{}, antwort.Fehler.AlsFehler()
	}

	offen := make([]rest.Nachlieferung, 0, len(antwort.Nachzuliefern))

	for _, zeile := range antwort.Nachzuliefern {
		offen = append(offen, rest.Nachlieferung{
			NachrichtID: zeile.NachrichtID, Vorgang: zeile.Vorgang,
			Empfaenger: zeile.Empfaenger, Folge: zeile.Folge, Art: zeile.Art,
		})
	}

	return rest.Wiederaufnahme{
		Nachzuliefern: offen, Klaerung: antwort.Klaerung,
	}, nil
}

// alsRestvorgang macht aus der Antwort des Fensters die Gestalt des Dienstes.
func alsRestvorgang(
	antwort app.HandelsvorgangAntwort,
) (rest.Handelsvorgang, error) {
	if antwort.Fehler != nil {
		return rest.Handelsvorgang{}, antwort.Fehler.AlsFehler()
	}

	return restvorgangAus(antwort.Vorgang), nil
}

// restvorgangAus übersetzt eine Revision nach außen.
func restvorgangAus(vorgang app.HandelsvorgangDTO) rest.Handelsvorgang {
	return rest.Handelsvorgang{
		Vorgang: vorgang.Vorgang, Revision: vorgang.Revision,
		Abdruck: vorgang.Abdruck, Eltern: vorgang.Eltern,
		Art: vorgang.Art, Initiator: vorgang.Initiator,
		Parteien: vorgang.Parteien, Ich: vorgang.Ich, Zustand: vorgang.Zustand,
		Positionen:       alsRestpositionen(vorgang.Positionen),
		BewertungGeteilt: vorgang.BewertungGeteilt,
		Zustimmungen:     vorgang.Zustimmungen,
		Ende:             vorgang.Ende,
		Beschluss:        vorgang.Beschluss, Lage: vorgang.Lage,
		Naechster: vorgang.Naechster, Moegliche: vorgang.Moegliche,
		Phase: vorgang.Phase,
	}
}

// alsRestpositionen übersetzt die Zeilen nach außen.
func alsRestpositionen(
	dtos []app.HandelspositionDTO,
) []rest.Handelsposition {
	positionen := make([]rest.Handelsposition, 0, len(dtos))

	for _, dto := range dtos {
		positionen = append(positionen, rest.Handelsposition{
			Von: dto.Von, Printing: dto.Printing, DruckID: dto.DruckID,
			Menge: dto.Menge, Guete: dto.Guete, Sprache: dto.Sprache,
			Finish: dto.Finish, Abbild: dto.Abbild, Bewertung: dto.Bewertung,
		})
	}

	return positionen
}

// alsAppPositionen übersetzt die Zeilen nach innen.
func alsAppPositionen(
	positionen []rest.Handelsposition,
) []app.HandelspositionDTO {
	dtos := make([]app.HandelspositionDTO, 0, len(positionen))

	for _, position := range positionen {
		dtos = append(dtos, app.HandelspositionDTO{
			Von: position.Von, Printing: position.Printing,
			DruckID: position.DruckID, Menge: position.Menge,
			Guete: position.Guete, Sprache: position.Sprache,
			Finish: position.Finish, Abbild: position.Abbild,
			Bewertung: position.Bewertung,
		})
	}

	return dtos
}

func (l leitungsweg) HandelFesthalten(
	ctx context.Context, vorgang string, exemplarID int64,
	menge int, fassung int64,
) (rest.Festgehalten, error) {
	antwort := app.PeerflaecheUeber(l.db).HandelFesthalten(
		ctx, vorgang, exemplarID, menge, fassung)

	zusagen, fehler := alsZusagen(antwort)
	if fehler != nil {
		return rest.Festgehalten{}, fehler
	}

	return rest.Festgehalten{
		Zusagen: zusagen, Nachladefehler: alsNachladebefund(antwort.Nachladefehler),
	}, nil
}

func (l leitungsweg) Handelszusagen(
	ctx context.Context, vorgang string,
) ([]rest.Handelszusage, error) {
	return alsZusagen(app.PeerflaecheUeber(l.db).Handelszusagen(ctx, vorgang))
}

// alsZusagen übersetzt die Antwort der Fläche für den Dienst.
func alsZusagen(antwort app.ZusagenAntwort) ([]rest.Handelszusage, error) {
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	zusagen := make([]rest.Handelszusage, 0, len(antwort.Zusagen))

	for _, zusage := range antwort.Zusagen {
		zusagen = append(zusagen, rest.Handelszusage{
			ExemplarID: zusage.ExemplarID, Menge: zusage.Menge,
			Fassung: zusage.Fassung, Revision: zusage.Revision,
			Frist: zusage.Frist, Uebergeben: zusage.Uebergeben,
		})
	}

	return zusagen, nil
}

// HandelPreisanhalt gibt den Preisanhalt eines Vorgangs (TRADE-02b).
func (l leitungsweg) HandelPreisanhalt(ctx context.Context, vorgang string) (any, error) {
	antwort := app.PeerflaecheUeber(l.db).HandelPreisanhaltVon(ctx, vorgang)
	if antwort.Fehler != nil {
		return nil, antwort.Fehler.AlsFehler()
	}

	return antwort.Anhalt, nil
}
