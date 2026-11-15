// **Die Paketpflege des MCP-Servers** —
// #1291.
//
// **Eine eigene Datei neben `bestand.go`**, weil sie eine andere Frage
// beantwortet: dort steht, was im Bestand liegt; hier, was auf die Platte
// kommt und was von ihr verschwindet (§2).

package mcpnaehte

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// pflege erfüllt `mcp.Paketpflege` über den Nähten von `internal/app`.
type pflege struct{ db *sql.DB }

// PaketInstallieren bindet ein Paket ein, das schon auf der Platte liegt.
func (p pflege) PaketInstallieren(ctx context.Context, quelle string) (mcp.Paketzeile, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return mcp.Paketzeile{}, err
	}

	paket, fehler := weg.Installieren(ctx, quelle)
	if fehler != nil {
		return mcp.Paketzeile{}, fehler
	}

	return alsPaketzeile(paket), nil
}

// PaketAnlegen schreibt ein neues Paket aus einem Gerüst (#1350).
func (p pflege) PaketAnlegen(ctx context.Context, art, name, vollname, spiel string) (
	mcp.Paketzeile, error,
) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return mcp.Paketzeile{}, err
	}

	paket, fehler := weg.Anlegen(ctx, art, name, vollname, spiel)
	if fehler != nil {
		return mcp.Paketzeile{}, fehler
	}

	return alsPaketzeile(paket), nil
}

// PaketKopieren legt eine eigene Kopie an (FMT-02, #1691).
func (p pflege) PaketKopieren(ctx context.Context, art, name, neu string) (mcp.Paketzeile, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return mcp.Paketzeile{}, err
	}

	paket, fehler := weg.Kopieren(ctx, art, name, neu)
	if fehler != nil {
		return mcp.Paketzeile{}, fehler
	}

	return alsPaketzeile(paket), nil
}

// PaketEntfernen nimmt ein Paket von der Platte und aus dem Katalog.
//
// **Erst die Bestätigung**: `pfad` muss der aus dem Katalog sein.
func (p pflege) PaketEntfernen(ctx context.Context, art, name, pfad string) error {
	// **Ohne Bestätigung wird hier nichts entfernt**
	// ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)):
	// ein Modell ruft, während niemand hinsieht. Das ist die Pflicht **dieser
	// Fläche** und keine Aussage über den Katalog — sie braucht keine Lesung.
	//
	// **Und sie steht vor der Naht** (A01, #2215): dahinter wird der Ort der
	// Pakete bestimmt. Bis dahin stand sie danach, und ein Ruf ohne
	// Bestätigung fragte schon das Zuhause ab — der Test, der „bevor
	// irgendetwas gelesen wird" zusagte, lief gegen das echte.
	if pfad == "" {
		return fmt.Errorf(
			"%w: ohne den Pfad aus »lesen.pakete« wird nichts entfernt",
			domain.ErrConflict)
	}

	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return err
	}

	// **Ob er stimmt, entscheidet der Dienst** — hinter derselben Lesung, die
	// auch löscht (#1344).
	// Ein Vergleich hier läse den Katalog ein zweites Mal, und Bestätigung und
	// Wirkung bezögen sich nicht auf denselben Stand.
	return weg.Entfernen(ctx, art, name, pfad)
}

// PaketUebernehmen holt alle Vorlagenpakete aus einem Verzeichnis.
func (p pflege) PaketUebernehmen(ctx context.Context, ort string) (mcp.Paketuebernahme, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return mcp.Paketuebernahme{}, err
	}

	uebernahme, fehler := weg.Uebernehmen(ctx, ort)
	if uebernahme == nil {
		return mcp.Paketuebernahme{}, fehler
	}

	return alsMCPUebernahme(uebernahme), fehler
}

// PaketWeitergeben schreibt Pakete als `.mse-installer` (FMT-02.AK-04, #1691).
//
// **Nichts wird überschrieben** (0044): die Datei, die entstehen wird, wird
// vorher reserviert — wie beim Zeichnen und Sichern. Am Fenster fragt der
// Dialog, an der Kommandozeile gilt `saveAs` wie bei MSE2.
func (p pflege) PaketWeitergeben(ctx context.Context, pakete []string, ziel string) (string, error) {
	pfad, err := app.Installerziel(pakete, ziel)
	if err != nil {
		return "", err
	}

	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return "", err
	}

	aufraeumen, err := reserviereOrt(pfad)
	if err != nil {
		return "", err
	}

	datei, err := weg.Weitergeben(ctx, pakete, pfad)
	if err != nil {
		aufraeumen()

		return "", err
	}

	return datei, nil
}

// PaketEinlesen legt die Pakete eines `.mse-installer` ins Verzeichnis — der
// Teilerfolg geht mit dem Fehler, wie bei der Übernahme.
func (p pflege) PaketEinlesen(ctx context.Context, datei string) (mcp.Paketuebernahme, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return mcp.Paketuebernahme{}, err
	}

	uebernahme, fehler := weg.Einlesen(ctx, datei)
	if uebernahme == nil {
		return mcp.Paketuebernahme{}, fehler
	}

	return alsMCPUebernahme(uebernahme), fehler
}

// alsMCPUebernahme übersetzt eine Übernahme — **keine Liste ist `nil`** (§3).
func alsMCPUebernahme(uebernahme *app.Paketuebernahme) mcp.Paketuebernahme {
	uebergangen := uebernahme.Uebergangen
	if uebergangen == nil {
		uebergangen = []string{}
	}

	return mcp.Paketuebernahme{
		Uebernommen: alsPaketzeilen(uebernahme.Uebernommen),
		Uebergangen: uebergangen,
		Befunde:     uebernahme.Befunde,
	}
}

// MSE2Orte nennt die üblichen Orte einer MSE2-Installation.
func (pflege) MSE2Orte() []string { return app.MSE2Orte() }

// kartenbilder setzt ein Bild von der Platte in ein Kartenfeld (#1291).
type kartenbilder struct{ db *sql.DB }

// AusDatei liest eine Bilddatei, schneidet sie zu und schreibt das Feld.
func (k kartenbilder) AusDatei(
	ctx context.Context, setID, karteID int64, feld, pfad string,
) (mcp.Bildmasse, error) {
	naht, err := app.Bilddateinaht(k.db, domain.UrheberServer)
	if err != nil {
		return mcp.Bildmasse{}, err
	}

	gesetzt, bildfehler := naht.Einsetzen(ctx, setID, karteID, feld, pfad)
	if bildfehler != nil {
		return mcp.Bildmasse{}, bildfehler
	}

	return mcp.Bildmasse{
		Feld:         gesetzt.Feld,
		QuelleBreite: gesetzt.QuelleBreite, QuelleHoehe: gesetzt.QuelleHoehe,
		ZielBreite: gesetzt.ZielBreite, ZielHoehe: gesetzt.ZielHoehe,
	}, nil
}

// alsPaketzeilen übersetzt eine Liste von Paketen.
func alsPaketzeilen(pakete []domain.Paket) []mcp.Paketzeile {
	zeilen := make([]mcp.Paketzeile, 0, len(pakete))
	for _, paket := range pakete {
		zeilen = append(zeilen, alsPaketzeile(paket))
	}

	return zeilen
}

// alsPaketzeile übersetzt ein Paket.
func alsPaketzeile(paket domain.Paket) mcp.Paketzeile {
	return mcp.Paketzeile{
		Art: paket.Art, Name: paket.Name, Version: paket.Version,
		Pfad: paket.Pfad, Herkunft: paket.Herkunft,
	}
}
