// Die Naht zum Scryfall-Bildweg für die Flächen ohne Fenster
// (#1005).
//
// # Warum diese Fähigkeit überhaupt ohne Fenster geht
//
// „Ein Kartenbild einsetzen" steht im Fähigkeitskatalog als **Dialogfähigkeit**
// — sie braucht eine Dateiwahl und einen gezogenen Zuschnitt, und beides gibt
// es ohne Fenster nicht. **Der Bildweg über Scryfall hat beides nicht**: der
// Kartenname bestimmt das Bild, und der Zuschnitt ist der, den MSEs eigene
// Regel aus Quelle und Feldfläche errechnet.
//
// Damit fällt der Grund weg, aus dem die Fähigkeit am Fenster hing — und §16
// sagt, was dann gilt: „Eine neue Fähigkeit landet nicht an einer Fläche."

package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Bildeinsatz ist, was ein Einsetzen zurückmeldet.
type Bildeinsatz struct {
	Feld      string `json:"feld"`
	Kuenstler string `json:"kuenstler"`

	// Nennung ist der Satz, der neben dem Bild stehen muss.
	Nennung string `json:"nennung"`

	// Adresse ist die Kartenseite bei Scryfall.
	Adresse string `json:"adresse"`

	// IllustratorGesetzt sagt, ob der Künstler in die Karte geschrieben wurde.
	//
	// **Damit eine Fläche es sagen kann.** Ohne Künstlernamen wird nichts
	// gesetzt (ein leeres Feld löschte einen vorhandenen Namen) — und dann
	// muss der Aufrufer wissen, dass die Nennung an ihm hängt.
	IllustratorGesetzt bool `json:"illustratorGesetzt"`

	// Befunde ist, was die Kartenauskunft tolerant hingenommen hat.
	//
	// **Sie kommen mit** (§16: „nichts wird still übernommen"). Der Weg
	// dahinter warf sie weg; der Grund steht bei [service.Bilduebernahme].
	Befunde []string `json:"befunde"`

	AlterStunden float64 `json:"alterStunden"`
}

// Fremdbild ist der Bildweg, wie ihn eine Fläche ohne Fenster benutzt.
type Fremdbild struct{ anwendung *App }

// Fremdbildnaht baut den Bildweg über einem Bestand.
//
// **Sie baut eine kopflose Fassade**, statt den Zuschnitt ein zweites Mal zu
// schreiben. `App` *ist* die Fassade; ihr fehlt ohne Fenster nur der Kontext,
// und den bringt jeder Aufruf mit.
func Fremdbildnaht(db *sql.DB, urheber domain.Urheber) (*Fremdbild, error) {
	// **Alle vier auf einmal, ein Zweig** — dieselbe Überlegung wie in
	// [Zeichennaht]: sie scheitern an derselben Ursache, dem fehlenden
	// Benutzerverzeichnis. Vier `if` darauf wären drei Zweige, die kein Aufruf
	// betreten kann (§3), und alle vier Fehler zu melden ist die bessere
	// Auskunft.
	schreiber, schreibfehler := Schreibnaht(db, urheber)

	werk, werkfehler := Zeichennaht(db)

	bilderort, ortfehler := Bilderverzeichnis()

	bildweg, wegfehler := ScryfallBildnaht(dienstwahlUeber(db))

	if fehler := errors.Join(
		schreibfehler, werkfehler, ortfehler, wegfehler); fehler != nil {
		return nil, fehler
	}

	anwendung := NeueApp("")
	anwendung.MitEditor(schreiber, Kartenlesernaht(db), werk.Planbauer, werk.Rasterer)
	anwendung.MitBildablage(service.NeueBildablage(bilderort))
	anwendung.MitScryfallbildern(bildweg)

	return &Fremdbild{anwendung: anwendung}, nil
}

// Einsetzen holt das Bild und legt es an der Karte ab.
//
// **In einem Schritt, anders als der Feldabgleich.** Der zeigt erst und
// schreibt dann, weil ein Regeltext von Scryfall einen getippten überschreiben
// könnte. Ein Bildfeld, das noch leer ist, hat diesen Fall nicht — und wer ihn
// hat, sieht den vorigen Wert in der Rücknahme.
//
// **Der Zuschnitt ist der vorgeschlagene.** Ohne Fenster zieht niemand einen;
// genommen wird der, den MSEs eigene Regel aus Quelle und Feldfläche errechnet
// (`render.Zuschnittflaeche`).
func (f *Fremdbild) Einsetzen(
	ctx context.Context, setID, karteID int64, feld, name, art string,
) (Bildeinsatz, error) {
	geholt := f.anwendung.ScryfallBild(ctx, kennungText(karteID), feld, name, art)
	if geholt.Fehler != nil {
		return Bildeinsatz{}, fmt.Errorf("das Bild holen: %w", AlsFehler(geholt.Fehler))
	}

	gesetzt := f.anwendung.ScryfallBildEinsetzen(ctx, kennungText(setID), kennungText(karteID),
		feld, geholt.Pfad, geholt.Kuenstler, geholt.Adresse, geholt.Auswahl)
	if gesetzt.Fehler != nil {
		return Bildeinsatz{}, fmt.Errorf("das Bild einsetzen: %w", AlsFehler(gesetzt.Fehler))
	}

	return Bildeinsatz{
		Feld: feld, Kuenstler: geholt.Kuenstler, Nennung: geholt.Nennung,
		Adresse: geholt.Adresse, IllustratorGesetzt: geholt.Kuenstler != "",
		AlterStunden: geholt.AlterStunden, Befunde: geholt.Befunde,
	}, nil
}

// kennungText macht aus einer Kennung das, was die Fassade erwartet.
//
// **Die Fassade nimmt Zeichenketten**, weil das Fenster keine 64-Bit-Zahlen
// kennt (JavaScript rundet ab 2^53). Eine Fläche mit echten Zahlen wandelt
// deshalb hier — und nicht die Fassade, die dann zwei Gestalten hätte.
func kennungText(kennung int64) string { return strconv.FormatInt(kennung, 10) }
