// **Die Doubles der Aktualisierung**
// (#1100).
//
// **Eine eigene Datei**, seit `vorrichtung_test.go` über die Zeilengrenze aus
// §2 ging — wie `vorrichtungbestand_test.go` und `vorrichtungzaehler_test.go`
// vor ihr. Der Schnitt folgt der Fläche, die sie bedienen.
package rest_test

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// NeuesteFassung und FassungLaden antworten willig.
//
// **Mit einer neueren Fassung und nicht mit derselben**: eine Antwort „nichts
// Neues" sähe aus wie ein gültiger Erfolg, und die vier Klassen prüften dann
// einen Weg, der nichts zu sagen hat.
func (williger) NeuesteFassung(context.Context) (rest.Fassungsstand, error) {
	return rest.Fassungsstand{
		Laufend: "0.1.0", Freigegeben: "0.2.0",
		Seite: "https://beispiel.test/v", Neuer: true,
	}, nil
}

func (williger) FassungLaden(context.Context) (string, error) {
	return "/Users/x/Downloads/MagicSetEditor4Mac-0.2.0.dmg", nil
}

// NeuesteFassung und FassungLaden zählen mit.
func (z *zaehlender) NeuesteFassung(ctx context.Context) (rest.Fassungsstand, error) {
	z.zaehle()

	return z.williger.NeuesteFassung(ctx)
}

func (z *zaehlender) FassungLaden(ctx context.Context) (string, error) {
	z.zaehle()

	return z.williger.FassungLaden(ctx)
}

// **Und die Sicherung des ganzen Bestands** (#1169).
func (williger) BestandSichern(context.Context, string, string, bool) error { return nil }

func (williger) SicherungSchauen(context.Context, string, string) (rest.Sicherungsangaben, error) {
	return rest.Sicherungsangaben{
		AngelegtAm: "2026-08-28T00:00:00.000Z", Schemastand: "0021_listenspalten.sql",
		Sets: 1, Karten: 3, Bilder: "geleert",
	}, nil
}

// **Und der Netzweg** (N5, #1571).
// **Die Regelfreigabe** (G4, #1461) — willig heißt: sie antwortet.
func (williger) Regelfreigabe(
	context.Context, string, string, int, string, []string,
) (
	rest.Regelfreigabeangaben, error,
) {
	return rest.Regelfreigabeangaben{
		Erlaubt: true, Regelwerk: "sandkasten", Fassung: 1,
		Wirksam: "2026-08-19", Hoechste: "begleitet",
		Matrix:      []rest.Regelbereichsangaben{},
		Luecken:     []string{},
		Hindernisse: []string{},
	}, nil
}

func (williger) Netzstand(context.Context) (rest.Netzstandangaben, error) {
	return rest.Netzstandangaben{FristVerbindung: 15, FristKopf: 30, FristStillstand: 60}, nil
}

// **Und der portable Umschlag** (N4, #1457).
func (williger) UmschlagSchnueren(context.Context, string, string, string) (
	rest.Umschlagangaben, error,
) {
	return rest.Umschlagangaben{Objekte: []rest.Umschlagobjekt{}}, nil
}

func (williger) UmschlagAnsehen(context.Context, string) (rest.Umschlagangaben, error) {
	return rest.Umschlagangaben{
		Urheber: "b1", Fingerabdruck: "AB CD",
		Objekte: []rest.Umschlagobjekt{{Art: "set", Bezeichnung: "set.mse-set", Groesse: 7}},
	}, nil
}

func (williger) UmschlagUebernehmen(context.Context, string) (rest.Umschlagannahme, error) {
	return rest.Umschlagannahme{SetID: "9", Karten: 3, Befunde: []string{}}, nil
}

// **Und die Partien** (G1, #1458).
func (williger) PartieAnlegen(
	context.Context, string, string, []rest.Sitzwunsch, []string,
) (rest.Partieangaben, error) {
	return einePartieangabe(), nil
}

// Partien gibt eine Zeile — die Liste ist ein eigener Weg (#1600).
func (williger) Partien(context.Context) ([]rest.Partiezeilenangaben, error) {
	return []rest.Partiezeilenangaben{{
		ID: "p1", Spiel: "magic", Stand: "laeuft", Objekte: 12,
		Sitze: []rest.Zeilensitzangaben{{Kennung: "anna", Name: "Anna"}},
	}}, nil
}

func (williger) Partie(_ context.Context, kennung, fuer string) (
	rest.Partieangaben, error,
) {
	angaben := einePartieangabe()
	angaben.ID, angaben.Fuer = kennung, fuer

	return angaben, nil
}

func (williger) PartieBefehl(context.Context, rest.Spielbefehl) (
	rest.Befehlsangaben, error,
) {
	return rest.Befehlsangaben{
		Fassung: "2",
		Ereignisse: []rest.Partieereignis{
			{Folge: "1", Art: "objekt.tappen", Kette: "abcd"},
		},
	}, nil
}

func (williger) PartieStand(_ context.Context, kennung, fuer, stand string) (
	rest.Partieangaben, error,
) {
	angaben := einePartieangabe()
	angaben.ID, angaben.Fuer, angaben.Stand = kennung, fuer, stand

	return angaben, nil
}

// **Und der Verlauf** (G5, #1462) — schon gefiltert, wie `service` ihn gibt.
func (williger) PartieVerlauf(_ context.Context, _, fuer string) (
	[]rest.Partieereignis, error,
) {
	return []rest.Partieereignis{{
		Folge: "1", Art: "objekt.ziehen",
		Oeffentlich: map[string]any{"sitz": fuer},
		Privat:      map[string]any{"karte": "Wald"},
		Kette:       "abcd",
	}}, nil
}

// einePartieangabe ist die Antwort, die dieses Doppel gibt.
func einePartieangabe() rest.Partieangaben {
	return rest.Partieangaben{
		ID: "p1", Fassung: "1", Spiel: "magic", Stand: "laeuft",
		Fuer: "sitz-1", Wirt: "sitz-1",
		Sitze: []rest.Partiesitz{
			{Kennung: "sitz-1", Name: "A", Karten: 60},
			{Kennung: "sitz-2", Name: "B", Karten: 60},
		},
		Objekte: []rest.Partieobjekt{
			{Kennung: "o1", Beherrscher: "sitz-1", Zone: "bibliothek"},
		},
		Erlaubt: []string{"objekt.tappen"},
	}
}
