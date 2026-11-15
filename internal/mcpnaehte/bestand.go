// Pakete, Kennzahlen, Verlauf und Einstellungen für den MCP-Server
// (#1107).

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// bestand führt die acht Werkzeuge an die Nähte dieser Schicht.
type bestand struct{ db *sql.DB }

func (b bestand) Pakete(ctx context.Context) ([]mcp.Paketzeile, []string, error) {
	weg, err := app.PaketnahtUeber(b.db)
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

	umgesetzt := make([]mcp.Paketzeile, 0, len(liste))
	for _, zeile := range liste {
		paket := zeile.Paket
		umgesetzt = append(umgesetzt, mcp.Paketzeile{
			Art: paket.Art, Name: paket.Name, Version: paket.Version,
			Pfad: paket.Pfad, Herkunft: paket.Herkunft,
			Pruefsumme: paket.Pruefsumme, Herkunftsart: zeile.Herkunftsart,
			Geaendert: paket.Geaendert, Braucht: app.BrauchtZeilen(paket),
			GebrauchtVon: append([]string{}, zeile.GebrauchtVon...), Benutzt: zeile.Benutzt,
		})
	}

	return umgesetzt, gemeldet, nil
}

func (b bestand) PaketBeziehen(
	ctx context.Context, adresse, stand, unterverzeichnis, name string,
) (mcp.Paketbezug, error) {
	weg, err := app.PaketnahtUeber(b.db)
	if err != nil {
		return mcp.Paketbezug{}, err
	}

	ergebnis, bezugsfehler := weg.Beziehen(ctx, adresse, stand, unterverzeichnis, name)
	if bezugsfehler != nil {
		return mcp.Paketbezug{}, bezugsfehler
	}

	return mcp.Paketbezug{
		Ort: ergebnis.Ort, Stand: ergebnis.Stand,
		Pruefsumme: ergebnis.Pruefsumme, Uebergangen: ergebnis.Uebergangen,
	}, nil
}

func (b bestand) Statistik(ctx context.Context, setID int64) (mcp.Setzahlen, error) {
	statistik, err := app.ZahlennahtUeber(b.db).Statistik(ctx, setID)
	if err != nil {
		return mcp.Setzahlen{}, err
	}

	belegung := make([]mcp.Zahl, 0, len(statistik.Belegung))
	for _, feld := range statistik.Belegung {
		belegung = append(belegung, mcp.Zahl{Name: feld.Name, Karten: feld.Karten})
	}

	return mcp.Setzahlen{
		Karten: statistik.Karten, Felder: statistik.Felder,
		Leere: statistik.Leere, Belegung: belegung,
	}, nil
}

func (b bestand) Verteilung(ctx context.Context, setID int64, feld string) (
	[]mcp.Wertzahl, error,
) {
	werte, err := app.ZahlennahtUeber(b.db).Verteilung(ctx, setID, feld)
	if err != nil {
		return nil, err
	}

	verteilung := make([]mcp.Wertzahl, 0, len(werte))
	for _, wert := range werte {
		verteilung = append(verteilung, mcp.Wertzahl{
			Name: wert.Name, Karten: wert.Karten, KartenIDs: wert.KartenIDs,
		})
	}

	return verteilung, nil
}

func (b bestand) Zuruecknehmen(
	ctx context.Context, setID int64, vorgang string,
) (mcp.Bewegung, error) {
	return bewegt(app.VerlaufsnahtUeber(b.db).Zuruecknehmen(ctx, setID, vorgang))
}

func (b bestand) Wiederherstellen(
	ctx context.Context, setID int64, vorgang string,
) (mcp.Bewegung, error) {
	return bewegt(app.VerlaufsnahtUeber(b.db).Wiederherstellen(ctx, setID, vorgang))
}

// Verlauf gibt die Schritte des Sets — dieselbe Übersetzung wie am Fenster,
// der jüngste zuerst (U6, #1417).
func (b bestand) Verlauf(ctx context.Context, setID int64) ([]mcp.Verlaufsschritt, error) {
	schritte, err := app.VerlaufsnahtUeber(b.db).Verlauf(ctx, setID)
	if err != nil {
		return nil, err
	}

	uebersetzt := make([]mcp.Verlaufsschritt, 0, len(schritte))
	for _, schritt := range schritte {
		uebersetzt = append(uebersetzt, mcp.Verlaufsschritt{
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

// bewegt übersetzt beide Richtungen — sie unterscheiden sich nur im Aufruf.
func bewegt(antwort app.BewegungAntwort, err error) (mcp.Bewegung, error) {
	if err != nil {
		return mcp.Bewegung{}, err
	}

	return mcp.Bewegung{
		Bewegt: antwort.Bewegt, KarteID: antwort.KarteID, Feld: antwort.Feld,
		Umfang: antwort.Umfang, Urheber: antwort.Urheber,
	}, nil
}

func (b bestand) Einstellungen(ctx context.Context) ([]mcp.Einstellungszeile, error) {
	alle, err := app.EinstellungsnahtUeber(b.db).Alle(ctx)
	if err != nil {
		return nil, err
	}

	umgesetzt := make([]mcp.Einstellungszeile, 0, len(alle))
	for _, einstellung := range alle {
		umgesetzt = append(umgesetzt, mcp.Einstellungszeile{
			Schluessel: einstellung.Schluessel, Wert: einstellung.Wert,
			Werte: einstellung.Werte, Bereich: einstellung.Bereich,
			Wirksam: einstellung.Wirksam, Ebene: einstellung.Ebene,
			Herkunft: einstellung.Herkunft, Vorgabe: einstellung.Vorgabe,
			Wirkung: einstellung.Wirkung,
		})
	}

	return umgesetzt, nil
}

func (b bestand) EinstellungSetzen(ctx context.Context, schluessel, wert string) error {
	return app.EinstellungsnahtUeber(b.db).Setzen(ctx, schluessel, wert)
}

// **Der Austausch der Einstellungen** — Werkstück O3 von Phase 5.2
// (#1427).
//
// **Über denselben Weg wie das Lesen und Setzen**: eine zweite Naht für
// dieselbe Tafel wären zwei Wege zu einer Sache, und der zweite bekäme beim
// nächsten Umbau keinen Riegel.

func (b bestand) EinstellungenAusfuehren(
	ctx context.Context, abschnitte []string,
) (string, error) {
	return app.EinstellungsnahtUeber(b.db).Ausfuhr(ctx, abschnitte)
}

func (b bestand) EinstellungenVorschau(ctx context.Context, roh []byte) (
	mcp.Einstellungssicht, error,
) {
	sicht, fehler := app.EinstellungsnahtUeber(b.db).Einfuhrvorschau(ctx, roh)
	if fehler != nil {
		return mcp.Einstellungssicht{}, fehler
	}

	return mcp.Einstellungssicht{
		Fassung: sicht.Fassung, Revision: sicht.Revision,
		Zeilen: austauschzeilen(sicht.Zeilen), Befunde: austauschbefunde(sicht.Befunde),
		Unveraendert: sicht.Unveraendert,
	}, nil
}

func (b bestand) EinstellungenUebernehmen(
	ctx context.Context, roh []byte, revision string, nur []string,
) (mcp.Einstellungsbericht, error) {
	bericht, fehler := app.EinstellungsnahtUeber(b.db).
		Einfuhruebernahme(ctx, roh, revision, nur)
	if fehler != nil {
		return mcp.Einstellungsbericht{}, fehler
	}

	return mcp.Einstellungsbericht{
		Uebernommen: bericht.Uebernommen, Revision: bericht.Revision,
		Zeilen: austauschzeilen(bericht.Zeilen), Befunde: austauschbefunde(bericht.Befunde),
	}, nil
}

// austauschzeilen bildet den Diff ab.
func austauschzeilen(zeilen []app.EinfuhrzeileDTO) []mcp.Einstellungsdiffzeile {
	umgesetzt := make([]mcp.Einstellungsdiffzeile, 0, len(zeilen))
	for _, zeile := range zeilen {
		umgesetzt = append(umgesetzt, mcp.Einstellungsdiffzeile{
			Ebene: zeile.Ebene, Schluessel: zeile.Schluessel,
			Vorher: zeile.Vorher, Nachher: zeile.Nachher, Herkunft: zeile.Herkunft,
		})
	}

	return umgesetzt
}

// austauschbefunde bildet die Befunde ab.
func austauschbefunde(befunde []app.EinfuhrbefundDTO) []mcp.Einstellungsbefund {
	umgesetzt := make([]mcp.Einstellungsbefund, 0, len(befunde))
	for _, befund := range befunde {
		umgesetzt = append(umgesetzt, mcp.Einstellungsbefund{
			Ebene: befund.Ebene, Schluessel: befund.Schluessel, Grund: befund.Grund,
		})
	}

	return umgesetzt
}
