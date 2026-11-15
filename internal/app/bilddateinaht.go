// **Ein Kartenbild von der Platte einsetzen — ohne Fenster** —
// #1291.
//
// # Der Befund
//
// `docs/FAEHIGKEITEN.md` führte die Zeile so:
//
//	| Ein Kartenbild einsetzen | `BildSetzen`, `BildEinsetzen`, `BildAusDatei` | — | — | — | ✋ |
//
// Die erhobene Hand meint einen Dateidialog, und der steckt allein in
// `Bindungen.BildSetzen`. `App.BildAusDatei` und `App.BildEinsetzen` nehmen
// einen Pfad entgegen — beide gibt es seit
// #246, weil
// das Fenster Dateien auch **ziehen und ablegen** lässt.
//
// # Und der Zuschnitt?
//
// **Der vorgeschlagene**, wie beim Bild von Scryfall
// ([Fremdbild.Einsetzen]): ohne Fenster zieht niemand einen Rahmen. Genommen
// wird der, den MSEs eigene Regel aus Quelle und Feldfläche errechnet
// (`render.Zuschnittflaeche`) — dasselbe, was das Fenster beim Öffnen anbietet.

package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Bilddatei setzt ein Kartenbild aus einer Datei auf der Platte.
type Bilddatei struct{ anwendung *App }

// Bildsetzung sagt, was gesetzt wurde.
//
// **Ohne `service`-Typen**, damit eine Fläche ohne Fenster sie lesen kann,
// ohne `internal/service` zu kennen (§2).
type Bildsetzung struct {
	Feld string

	// QuelleBreite und QuelleHoehe sind die Maße der gelesenen Datei.
	QuelleBreite int
	QuelleHoehe  int

	// ZielBreite und ZielHoehe sind die Maße, die der Stil für dieses Feld
	// vorsieht — auf sie wird zugeschnitten.
	ZielBreite int
	ZielHoehe  int
}

// Bilddateinaht baut den Weg, auf dem ein Bild von der Platte an eine Karte kommt.
//
// **Dieselben Teile wie [Fremdbildnaht]**, nur ohne den Weg zu Scryfall: ein
// Bild, das schon daliegt, wird nicht geholt.
func Bilddateinaht(db *sql.DB, urheber domain.Urheber) (*Bilddatei, error) {
	schreiber, schreibfehler := Schreibnaht(db, urheber)

	werk, werkfehler := Zeichennaht(db)

	bilderort, ortfehler := Bilderverzeichnis()

	if fehler := errors.Join(schreibfehler, werkfehler, ortfehler); fehler != nil {
		return nil, fehler
	}

	anwendung := NeueApp("")
	anwendung.MitEditor(schreiber, Kartenlesernaht(db), werk.Planbauer, werk.Rasterer)
	anwendung.MitBildablage(service.NeueBildablage(bilderort))

	return &Bilddatei{anwendung: anwendung}, nil
}

// Einsetzen liest eine Bilddatei, schneidet sie zu und schreibt das Feld.
func (b *Bilddatei) Einsetzen(
	ctx context.Context, setID, karteID int64, feld, pfad string,
) (Bildsetzung, error) {
	// **Erst lesen, dann schreiben.** `BildAusDatei` gibt die Maße und weist
	// eine Datei ab, die kein Bild ist — bevor irgendetwas an der Karte
	// geschieht.
	vorbereitet := b.anwendung.BildAusDatei(ctx, kennungText(karteID), feld, pfad)
	if vorbereitet.Fehler != nil {
		return Bildsetzung{}, fmt.Errorf("das Bild lesen: %w", AlsFehler(vorbereitet.Fehler))
	}

	gesetzt := b.anwendung.BildEinsetzen(ctx, kennungText(setID), kennungText(karteID),
		feld, pfad, vorbereitet.Auswahl)
	if gesetzt.Fehler != nil {
		return Bildsetzung{}, fmt.Errorf("das Bild einsetzen: %w", AlsFehler(gesetzt.Fehler))
	}

	return Bildsetzung{
		Feld:         feld,
		QuelleBreite: vorbereitet.QuelleBreite,
		QuelleHoehe:  vorbereitet.QuelleHoehe,
		ZielBreite:   vorbereitet.ZielBreite,
		ZielHoehe:    vorbereitet.ZielHoehe,
	}, nil
}
