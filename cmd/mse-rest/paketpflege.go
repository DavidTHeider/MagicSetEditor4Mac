// **Die Paketpflege des Dienstes** —
// #1291.
//
// **Eine eigene Datei neben `bestandswege.go`**, weil sie eine andere Frage
// beantwortet: dort steht, was im Bestand liegt und wie man es holt; hier,
// was auf die Platte kommt und was von ihr verschwindet (§2).

package main

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// Installieren bindet ein Paket ein, das schon auf der Platte liegt.
func (p pakete) PaketInstallieren(ctx context.Context, quelle string) (rest.Paket, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return rest.Paket{}, err
	}

	paket, fehler := weg.Installieren(ctx, quelle)
	if fehler != nil {
		return rest.Paket{}, fehler
	}

	return alsRestpaket(paket), nil
}

// PaketAnlegen schreibt ein neues Paket aus einem Gerüst (#1350).
func (p pakete) PaketAnlegen(ctx context.Context, art, name, vollname, spiel string) (
	rest.Paket, error,
) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return rest.Paket{}, err
	}

	paket, fehler := weg.Anlegen(ctx, art, name, vollname, spiel)
	if fehler != nil {
		return rest.Paket{}, fehler
	}

	return alsRestpaket(paket), nil
}

// PaketKopieren legt eine eigene Kopie an (FMT-02, #1691).
func (p pakete) PaketKopieren(ctx context.Context, art, name, neu string) (rest.Paket, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return rest.Paket{}, err
	}

	paket, fehler := weg.Kopieren(ctx, art, name, neu)
	if fehler != nil {
		return rest.Paket{}, fehler
	}

	return alsRestpaket(paket), nil
}

// PaketEntfernen nimmt ein Paket von der Platte und aus dem Katalog.
//
// **Erst die Bestätigung**: `pfad` muss der aus dem Katalog sein.
func (p pakete) PaketEntfernen(ctx context.Context, art, name, pfad string) error {
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
			"%w: ohne den Pfad aus »GET /pakete« wird nichts entfernt",
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

// Uebernehmen holt alle Vorlagenpakete aus einem Verzeichnis.
func (p pakete) PaketUebernehmen(ctx context.Context, ort string) (rest.Uebernahme, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return rest.Uebernahme{}, err
	}

	uebernahme, fehler := weg.Uebernehmen(ctx, ort)
	if uebernahme == nil {
		return rest.Uebernahme{}, fehler
	}

	return alsRestuebernahme(uebernahme), fehler
}

// PaketWeitergeben schreibt Pakete als `.mse-installer` (FMT-02.AK-04, #1691).
//
// **Nichts wird überschrieben** — die Datei wird vorher reserviert, wie beim
// Zeichnen und beim Bogen dieses Dienstes.
func (p pakete) PaketWeitergeben(ctx context.Context, namen []string, ziel string) (string, error) {
	pfad, err := app.Installerziel(namen, ziel)
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

	datei, err := weg.Weitergeben(ctx, namen, pfad)
	if err != nil {
		aufraeumen()

		return "", err
	}

	return datei, nil
}

// PaketEinlesen legt die Pakete eines `.mse-installer` ins Verzeichnis —
// **der Teilerfolg geht mit dem Fehler**, wie bei der Übernahme.
func (p pakete) PaketEinlesen(ctx context.Context, datei string) (rest.Uebernahme, error) {
	weg, err := app.PaketnahtUeber(p.db)
	if err != nil {
		return rest.Uebernahme{}, err
	}

	uebernahme, fehler := weg.Einlesen(ctx, datei)
	if uebernahme == nil {
		return rest.Uebernahme{}, fehler
	}

	return alsRestuebernahme(uebernahme), fehler
}

// alsRestuebernahme übersetzt eine Übernahme für den Dienst — **keine Liste
// ist `nil`** (§3).
func alsRestuebernahme(uebernahme *app.Paketuebernahme) rest.Uebernahme {
	uebernommen := make([]rest.Paket, 0, len(uebernahme.Uebernommen))
	for _, paket := range uebernahme.Uebernommen {
		uebernommen = append(uebernommen, alsRestpaket(paket))
	}

	uebergangen := uebernahme.Uebergangen
	if uebergangen == nil {
		uebergangen = []string{}
	}

	return rest.Uebernahme{
		Uebernommen: uebernommen, Uebergangen: uebergangen, Befunde: uebernahme.Befunde,
	}
}

// MSE2Orte nennt die üblichen Orte einer MSE2-Installation.
func (pakete) MSE2Orte() []string { return app.MSE2Orte() }

// kartenbilder setzt ein Bild von der Platte in ein Kartenfeld (#1291).
type kartenbilder struct{ db *sql.DB }

// AusDatei liest eine Bilddatei, schneidet sie zu und schreibt das Feld.
func (k kartenbilder) AusDatei(
	ctx context.Context, setID, karteID int64, feld, pfad string,
) (rest.Bildmasse, error) {
	naht, err := app.Bilddateinaht(k.db, domain.UrheberServer)
	if err != nil {
		return rest.Bildmasse{}, err
	}

	gesetzt, bildfehler := naht.Einsetzen(ctx, setID, karteID, feld, pfad)
	if bildfehler != nil {
		return rest.Bildmasse{}, bildfehler
	}

	return rest.Bildmasse{
		Feld:         gesetzt.Feld,
		QuelleBreite: gesetzt.QuelleBreite, QuelleHoehe: gesetzt.QuelleHoehe,
		ZielBreite: gesetzt.ZielBreite, ZielHoehe: gesetzt.ZielHoehe,
	}, nil
}

// alsRestpaket übersetzt ein Paket der Domäne.
func alsRestpaket(paket domain.Paket) rest.Paket {
	return rest.Paket{
		Art: paket.Art, Name: paket.Name, Version: paket.Version,
		Pfad: paket.Pfad, Herkunft: paket.Herkunft,
	}
}
