// **Was eine Fläche über ein Set und eine Karte wissen muss.**
//
// Eine eigene Datei, weil es eine eigene Frage ist: `mcpnaht.go` verdrahtet die
// Schreibwege — hier steht, was **alle vier** Flächen lesen müssen, bevor sie
// mit Scryfall reden. Vier Kopien dieser Abfrage liefen an dem Tag auseinander,
// an dem eine von ihnen die Setprüfung lernte.

package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Kartenstand gibt die heutigen Feldwerte einer Karte und das Spiel ihres Sets.
//
// # Warum das hier steht und nicht in jeder Fläche
//
// Weil **alle vier** es brauchen: ein Vorschlag entsteht aus dem, was heute in
// der Karte steht, und aus dem Spiel — ohne das gibt es keine Zuordnung. Vier
// Kopien dieser Abfrage liefen an dem Tag auseinander, an dem eine von ihnen
// die Setprüfung lernte.
//
// **Und die Setprüfung ist der Grund, warum es sie gibt**: die Karte muss zu
// dem genannten Set gehören. Sonst käme die Zuordnung vom falschen Spiel.
func Kartenstand(
	ctx context.Context, db *sql.DB, setID, karteID int64,
) (werte map[string]string, spiel string, fehler error) {
	return kartenstandUeber(ctx, storage.NeuerSetLeser(db), setID, karteID)
}

// Setspiel gibt das Spiel eines Sets — ohne eine Karte zu brauchen (#1032).
//
// **Für den Weg, der noch keine Karte hat**: das Anlegen aus einem
// Scryfall-Namen braucht die Zuordnung, und die hängt am Spiel. [Kartenstand]
// verlangt eine Karte und wäre hier eine Vorbedingung, die es nicht gibt.
func Setspiel(ctx context.Context, db *sql.DB, setID int64) (string, error) {
	return setspielUeber(ctx, storage.NeuerSetLeser(db), setID)
}

// setspielUeber liest über einer gegebenen Quelle.
func setspielUeber(ctx context.Context, leser Kartenquelle, setID int64) (string, error) {
	koepfe, err := leser.Setkoepfe(ctx)
	if err != nil {
		return "", err
	}

	spiel, gibtEsDasSet := spielVon(koepfe, setID)
	if !gibtEsDasSet {
		return "", fmt.Errorf("%w: das Set %d", domain.ErrNotFound, setID)
	}

	return spiel, nil
}

// Kartenquelle ist, was [Kartenstand] liest.
//
// **Als Schnittstelle, damit auch die Fehlerwege einen Test haben** — ein
// Bestand, der nicht antwortet, lässt sich mit einer echten Datenbank nicht
// herstellen (§3: eine unbequeme Verzweigung ist kein Grund für eine Ausnahme).
type Kartenquelle interface {
	KarteLesen(ctx context.Context, karteID int64) (*domain.CardRow, error)
	Setkoepfe(ctx context.Context) ([]domain.SetHead, error)
}

// kartenstandUeber liest über einer gegebenen Quelle.
//
// # Die Reihenfolge: erst das Set, dann die Karte
//
// **Weil sonst ein Zweig übrig bliebe, den niemand betritt.** Wer die Karte
// zuerst liest und ihr Set prüft, weiß danach, dass es das Set gibt — und der
// „Set nicht gefunden"-Ausgang darunter wäre unerreichbar. So herum trägt
// jeder Ausgang einen Fall.
func kartenstandUeber(
	ctx context.Context, leser Kartenquelle, setID, karteID int64,
) (werte map[string]string, spiel string, fehler error) {
	koepfe, err := leser.Setkoepfe(ctx)
	if err != nil {
		return nil, "", err
	}

	gefundenesSpiel, gibtEsDasSet := spielVon(koepfe, setID)
	if !gibtEsDasSet {
		return nil, "", fmt.Errorf("%w: das Set %d", domain.ErrNotFound, setID)
	}

	karte, kartenfehler := leser.KarteLesen(ctx, karteID)
	if kartenfehler != nil {
		return nil, "", kartenfehler
	}

	if karte.SetID != setID {
		return nil, "", fmt.Errorf(
			"%w: die Karte %d gehört nicht zum Set %d", domain.ErrNotFound, karteID, setID)
	}

	jetzige := make(map[string]string, len(karte.Values))
	for _, wert := range karte.Values {
		jetzige[wert.Field] = wert.Text
	}

	return jetzige, gefundenesSpiel, nil
}

// spielVon sucht das Spiel eines Sets in einer Kopfliste.
func spielVon(koepfe []domain.SetHead, setID int64) (spiel string, gefunden bool) {
	for _, kopf := range koepfe {
		if kopf.ID == setID {
			return kopf.Game, true
		}
	}

	return "", false
}

// Kartenlesernaht gibt den Kartenleser über einem Bestand.
//
// **Sie steht hier, weil `storage` hier stehen darf** (§2): `app.go` und diese
// Datei sind die beiden Composition Roots. Ohne sie müsste jede weitere Naht
// den Speicher selbst kennen — und die Aufzählung, die die Kante begrenzt,
// wäre keine mehr.
func Kartenlesernaht(db *sql.DB) Kartenleser { return storage.NeuerSetLeser(db) }
