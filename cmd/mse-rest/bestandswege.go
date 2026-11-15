// Pakete, Kennzahlen, Verlauf und Einstellungen für den Dienst
// (#1107).
package main

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// pakete führt die zwei Paketendpunkte an die Naht dieser Schicht.
type pakete struct{ db *sql.DB }

func (p pakete) Liste(ctx context.Context) ([]rest.Paket, []string, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return nil, nil, err
	}

	liste, befunde, lesefehler := weg.Liste(ctx)

	gemeldet := make([]string, 0, len(befunde))
	for _, befund := range befunde {
		gemeldet = append(gemeldet, befund.String())
	}

	if lesefehler != nil {
		return nil, gemeldet, lesefehler
	}

	umgesetzt := make([]rest.Paket, 0, len(liste))
	for _, zeile := range liste {
		paket := zeile.Paket
		umgesetzt = append(umgesetzt, rest.Paket{
			Art: paket.Art, Name: paket.Name, Version: paket.Version,
			Pfad: paket.Pfad, Herkunft: paket.Herkunft,
			Pruefsumme: paket.Pruefsumme, Herkunftsart: zeile.Herkunftsart,
			Geaendert: paket.Geaendert, Braucht: app.BrauchtZeilen(paket),
			GebrauchtVon: append([]string{}, zeile.GebrauchtVon...), Benutzt: zeile.Benutzt,
		})
	}

	return umgesetzt, gemeldet, nil
}

func (p pakete) Beziehen(
	ctx context.Context, adresse, stand, unterverzeichnis, name string,
) (rest.Bezug, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return rest.Bezug{}, err
	}

	ergebnis, bezugsfehler := weg.Beziehen(ctx, adresse, stand, unterverzeichnis, name)
	if bezugsfehler != nil {
		return rest.Bezug{}, bezugsfehler
	}

	return rest.Bezug{
		Ort: ergebnis.Ort, Stand: ergebnis.Stand,
		Pruefsumme: ergebnis.Pruefsumme, Uebergangen: ergebnis.Uebergangen,
	}, nil
}

// zahlen führt die zwei Auszählendpunkte an die Naht dieser Schicht.
type zahlen struct{ db *sql.DB }

func (z zahlen) Statistik(ctx context.Context, setID int64) (rest.Setzahlen, error) {
	statistik, err := app.ZahlennahtUeber(z.db).Statistik(ctx, setID)
	if err != nil {
		return rest.Setzahlen{}, err
	}

	belegung := make([]rest.Zahl, 0, len(statistik.Belegung))
	for _, feld := range statistik.Belegung {
		belegung = append(belegung, rest.Zahl{Name: feld.Name, Karten: feld.Karten})
	}

	return rest.Setzahlen{
		Karten: statistik.Karten, Felder: statistik.Felder,
		Leere: statistik.Leere, Belegung: belegung,
	}, nil
}

func (z zahlen) Verteilung(ctx context.Context, setID int64, feld string) (
	[]rest.Wertzahl, error,
) {
	werte, err := app.ZahlennahtUeber(z.db).Verteilung(ctx, setID, feld)
	if err != nil {
		return nil, err
	}

	verteilung := make([]rest.Wertzahl, 0, len(werte))
	for _, wert := range werte {
		verteilung = append(verteilung, rest.Wertzahl{
			Name: wert.Name, Karten: wert.Karten, KartenIDs: wert.KartenIDs,
		})
	}

	return verteilung, nil
}

// verlauf führt die zwei Bewegungsendpunkte an die Naht dieser Schicht.
type verlauf struct{ db *sql.DB }

// Verlauf gibt die Schritte des Sets — dieselbe Übersetzung wie am Fenster,
// der jüngste zuerst (U6, #1417).
func (v verlauf) Verlauf(ctx context.Context, setID int64) ([]rest.Verlaufsschritt, error) {
	schritte, err := app.VerlaufsnahtUeber(v.db).Verlauf(ctx, setID)
	if err != nil {
		return nil, err
	}

	uebersetzt := make([]rest.Verlaufsschritt, 0, len(schritte))
	for _, schritt := range schritte {
		uebersetzt = append(uebersetzt, rest.Verlaufsschritt{
			ID: schritt.ID, Zeitpunkt: schritt.Zeitpunkt, Handlung: schritt.Handlung,
			KarteID: schritt.KarteID, Feld: schritt.Feld,
			Vorher: schritt.Vorher, GabVorher: schritt.GabVorher,
			Nachher: schritt.Nachher, GabNachher: schritt.GabNachher,
			Zurueckgenommen: schritt.Zurueckgenommen, Urheber: schritt.Urheber,
			Herkunft: schritt.Herkunft, Vorgang: schritt.Vorgang, Umfang: schritt.Umfang,
		})
	}

	return uebersetzt, nil
}

func (v verlauf) Zuruecknehmen(
	ctx context.Context, setID int64, vorgang string,
) (rest.Bewegung, error) {
	return bewegt(app.VerlaufsnahtUeber(v.db).Zuruecknehmen(ctx, setID, vorgang))
}

func (v verlauf) Wiederherstellen(
	ctx context.Context, setID int64, vorgang string,
) (rest.Bewegung, error) {
	return bewegt(app.VerlaufsnahtUeber(v.db).Wiederherstellen(ctx, setID, vorgang))
}

// bewegt übersetzt beide Richtungen — sie unterscheiden sich nur im Aufruf.
func bewegt(antwort app.BewegungAntwort, err error) (rest.Bewegung, error) {
	if err != nil {
		return rest.Bewegung{}, err
	}

	return rest.Bewegung{
		Bewegt: antwort.Bewegt, KarteID: antwort.KarteID, Feld: antwort.Feld,
		Umfang: antwort.Umfang, Urheber: antwort.Urheber,
	}, nil
}

// einstellungen führt die zwei Einstellungsendpunkte an die Naht dieser Schicht.
type einstellungen struct{ db *sql.DB }

func (e einstellungen) Alle(ctx context.Context) ([]rest.Einstellung, error) {
	alle, err := app.EinstellungsnahtUeber(e.db).Alle(ctx)
	if err != nil {
		return nil, err
	}

	umgesetzt := make([]rest.Einstellung, 0, len(alle))
	for _, einstellung := range alle {
		umgesetzt = append(umgesetzt, rest.Einstellung{
			Schluessel: einstellung.Schluessel, Wert: einstellung.Wert,
			Werte: einstellung.Werte, Bereich: einstellung.Bereich,
			Wirksam: einstellung.Wirksam, Ebene: einstellung.Ebene,
			Herkunft: einstellung.Herkunft, Vorgabe: einstellung.Vorgabe,
			Wirkung: einstellung.Wirkung,
		})
	}

	return umgesetzt, nil
}

func (e einstellungen) EinstellungSetzen(ctx context.Context, schluessel, wert string) error {
	return app.EinstellungsnahtUeber(e.db).Setzen(ctx, schluessel, wert)
}

// **Der Austausch der Einstellungen** — Werkstück O3 von Phase 5.2
// (#1427).
//
// **Über denselben Weg wie das Lesen und Setzen**: eine zweite Naht für
// dieselbe Tafel wären zwei Wege zu einer Sache.

func (e einstellungen) Ausfuhr(ctx context.Context, abschnitte []string) (string, error) {
	return app.EinstellungsnahtUeber(e.db).Ausfuhr(ctx, abschnitte)
}

func (e einstellungen) Vorschau(ctx context.Context, roh []byte) (
	rest.Einstellungssicht, error,
) {
	sicht, fehler := app.EinstellungsnahtUeber(e.db).Einfuhrvorschau(ctx, roh)
	if fehler != nil {
		return rest.Einstellungssicht{}, fehler
	}

	return rest.Einstellungssicht{
		Fassung: sicht.Fassung, Revision: sicht.Revision,
		Zeilen: austauschzeilen(sicht.Zeilen), Befunde: austauschbefunde(sicht.Befunde),
		Unveraendert: sicht.Unveraendert,
	}, nil
}

func (e einstellungen) Uebernehmen(
	ctx context.Context, roh []byte, revision string, nur []string,
) (rest.Einstellungsbericht, error) {
	bericht, fehler := app.EinstellungsnahtUeber(e.db).
		Einfuhruebernahme(ctx, roh, revision, nur)
	if fehler != nil {
		return rest.Einstellungsbericht{}, fehler
	}

	return rest.Einstellungsbericht{
		Uebernommen: bericht.Uebernommen, Revision: bericht.Revision,
		Zeilen: austauschzeilen(bericht.Zeilen), Befunde: austauschbefunde(bericht.Befunde),
	}, nil
}

// austauschzeilen bildet den Diff ab.
func austauschzeilen(zeilen []app.EinfuhrzeileDTO) []rest.Einstellungsdiffzeile {
	umgesetzt := make([]rest.Einstellungsdiffzeile, 0, len(zeilen))
	for _, zeile := range zeilen {
		umgesetzt = append(umgesetzt, rest.Einstellungsdiffzeile{
			Ebene: zeile.Ebene, Schluessel: zeile.Schluessel,
			Vorher: zeile.Vorher, Nachher: zeile.Nachher, Herkunft: zeile.Herkunft,
		})
	}

	return umgesetzt
}

// austauschbefunde bildet die Befunde ab.
func austauschbefunde(befunde []app.EinfuhrbefundDTO) []rest.Einstellungsbefund {
	umgesetzt := make([]rest.Einstellungsbefund, 0, len(befunde))
	for _, befund := range befunde {
		umgesetzt = append(umgesetzt, rest.Einstellungsbefund{
			Ebene: befund.Ebene, Schluessel: befund.Schluessel, Grund: befund.Grund,
		})
	}

	return umgesetzt
}
