package service

// **Kopieren, Umbenennen, Löschen, Einsetzen, Ein- und Ausgeben** — die
// übrigen Wege der Symbolbibliothek (DOM-02,
// #1697).
// Eine eigene Datei neben `symbolbibliothek.go`: dort, was die Bibliothek
// weiß, hier, was sie tut.

import (
	"context"
	"errors"
	"image"
	"image/color"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Kopieren legt ein Symbol unter einem anderen Namen noch einmal an.
func (b *Symbolbibliothek) Kopieren(ctx context.Context, id int64, name string) (domain.Bibliothekssymbol, error) {
	geprueft, err := domain.PruefeSymbolname(name)
	if err != nil {
		return domain.Bibliothekssymbol{}, err
	}

	symbol, err := b.speicher.Symbol(ctx, id)
	if err != nil {
		return domain.Bibliothekssymbol{}, err
	}

	return b.speicher.SymbolAnlegen(ctx, domain.Bibliothekssymbol{
		Name: geprueft, Inhalt: symbol.Inhalt, Abdruck: symbol.Abdruck,
	})
}

// Umbenennen gibt einem Symbol einen anderen Namen.
func (b *Symbolbibliothek) Umbenennen(ctx context.Context, id int64, name string) error {
	geprueft, err := domain.PruefeSymbolname(name)
	if err != nil {
		return err
	}

	return b.speicher.SymbolUmbenennen(ctx, id, geprueft)
}

// Loeschen nimmt ein Symbol aus der Bibliothek und nennt die Sets, die es
// tragen — die Folgen (AK-02).
//
// **Ohne `bestaetigt` geht ein verwendetes Symbol nicht**: die Absage nennt
// die Sets. **Mit Bestätigung behalten sie ihre Kopie** — gelöscht wird nur
// der Eintrag der Bibliothek, und mit ihm der Nachweis, dass diese Sets
// dasselbe Symbol tragen.
func (b *Symbolbibliothek) Loeschen(ctx context.Context, id int64, bestaetigt bool) ([]string, error) {
	symbol, err := b.speicher.Symbol(ctx, id)
	if err != nil {
		return nil, err
	}

	traeger, err := b.traegerJeAbdruck(ctx)
	if err != nil {
		return nil, err
	}

	folgen := traegernamen(traeger[symbol.Abdruck])

	if len(folgen) > 0 && !bestaetigt {
		return folgen, domain.NewFieldError("id", domain.SchluesselSymbolVerwendet,
			"%s tragen dieses Symbol; gelöscht wird nur mit Bestätigung.", strings.Join(folgen, ", ")).
			MitWerten(map[string]string{"sets": strings.Join(folgen, ", ")})
	}

	return folgen, b.speicher.SymbolLoeschen(ctx, id)
}

// Einsetzen legt eine Kopie des Symbols neben das Set und setzt sie als
// dessen Symbol — und gibt den Namen der Datei.
func (b *Symbolbibliothek) Einsetzen(ctx context.Context, id, setID int64) (string, error) {
	if err := b.darfSchreiben(ctx); err != nil {
		return "", err
	}

	symbol, err := b.speicher.Symbol(ctx, id)
	if err != nil {
		return "", err
	}

	if fehlt := SetDaSein(ctx, b.setkoepfe, setID); fehlt != nil {
		return "", fehlt
	}

	zeichnung, _, err := msepaket.SymbolzeichnungAus(symbol.Name, []byte(symbol.Inhalt))
	if err != nil {
		return "", err
	}

	datei, err := b.ablage.Ablegen(setID, zeichnung, "")
	if err != nil {
		return "", err
	}

	// **Erst die Datei, dann der Wert**: ein Wert, der auf eine Datei zeigt,
	// die es nicht gibt, zeichnete das Set ohne Symbol. Eine Datei ohne Wert
	// ist nur eine Datei mehr neben dem Set.
	if _, _, err := b.setwerte.SetwertSetzenMitVorwert(ctx, setID, setsymbolfeld, datei); err != nil {
		return "", err
	}

	return datei, nil
}

// Ausgeben schreibt ein Symbol als `.mse-symbol` an einen Ort — atomar und
// überschreibend: ob dort etwas liegen darf, entscheidet die Fläche.
func (b *Symbolbibliothek) Ausgeben(ctx context.Context, id int64, ziel string) error {
	if err := b.darfSchreiben(ctx); err != nil {
		return err
	}

	symbol, err := b.speicher.Symbol(ctx, id)
	if err != nil {
		return err
	}

	if strings.TrimSpace(ziel) == "" {
		return domain.NewFieldError("ziel", domain.SchluesselPflichtfeld, "Es fehlt der Ort.")
	}

	return schreibeAtomar(ziel, []byte(symbol.Inhalt))
}

// Einlesen liest eine `.mse-symbol` und legt sie als neues Symbol an. Ohne
// Namen heißt es wie die Datei.
func (b *Symbolbibliothek) Einlesen(ctx context.Context, pfad, name string) (domain.Bibliothekssymbol, error) {
	if strings.TrimSpace(name) == "" {
		name = strings.TrimSuffix(filepath.Base(pfad), ".mse-symbol")
	}

	geprueft, err := domain.PruefeSymbolname(name)
	if err != nil {
		return domain.Bibliothekssymbol{}, err
	}

	zeichnung, _, err := msepaket.LiesSymbolzeichnung(pfad)
	if err != nil {
		return domain.Bibliothekssymbol{}, symbollesefehler(pfad, err)
	}

	return b.anlegen(ctx, geprueft, zeichnung)
}

// Symbol gibt ein Symbol samt Inhalt — für die Flächen, die es als Text
// weiterreichen, weil sie an keiner Platte stehen, die ihr Gegenüber sieht.
func (b *Symbolbibliothek) Symbol(ctx context.Context, id int64) (domain.Bibliothekssymbol, error) {
	return b.speicher.Symbol(ctx, id)
}

// EinlesenText legt ein Symbol aus dem Text einer `.mse-symbol` an — mit
// denselben Grenzen wie eine Datei. **Ein Text hat keinen Dateinamen**, der
// Name ist deshalb Pflicht.
func (b *Symbolbibliothek) EinlesenText(ctx context.Context, name string, roh []byte) (domain.Bibliothekssymbol, error) {
	geprueft, err := domain.PruefeSymbolname(name)
	if err != nil {
		return domain.Bibliothekssymbol{}, err
	}

	zeichnung, _, err := msepaket.SymbolzeichnungAus(geprueft, roh)
	if err != nil {
		return domain.Bibliothekssymbol{}, symbollesefehler(geprueft, err)
	}

	return b.anlegen(ctx, geprueft, zeichnung)
}

// symbollesefehler macht aus einem Lesefehler einen, den das Fenster sagen
// kann (AK-04): zu groß, fehlt, oder keine Zeichnung. Ein Befund über den
// Umfang kommt schon mit seinem Schlüssel.
func symbollesefehler(pfad string, err error) error {
	var feldfehler *domain.FieldError

	switch {
	case errors.As(err, &feldfehler):
		return err

	case errors.Is(err, msepaket.ErrZuGross):
		return domain.NewFieldError("pfad", domain.SchluesselSymbolZuGross,
			"%s ist größer, als eine Symboldatei sein darf.", filepath.Base(pfad)).
			MitWerten(map[string]string{"grenze": "1 MiB"})

	case errors.Is(err, fs.ErrNotExist):
		return domain.MitGrund(domain.ErrNotFound, err, "%s gibt es nicht", filepath.Base(pfad))
	}

	return domain.MitGrund(domain.ErrValidation, err, "%s ist keine lesbare Symbolzeichnung", filepath.Base(pfad))
}

// Vorschaurand ist die Randbreite der Vorschau — ein Anteil der Größe, wie
// `border radius:` einer Fassung; 0,05 ist MSEs Vorgabe
// (`SymbolVariation::SymbolVariation`, `src/data/field/symbol.cpp`, Fassung
// f9d9356 vom 2021-07-06).
const Vorschaurand = 0.05

// Symbolvorschau rastert eine Zeichnung in `groesse` Punkten —
// **schwarz mit weißem Rand**, der schlichtesten Färbung. Wie ein Stil sie
// färbt, sagt er selbst in seinen Fassungen.
func Symbolvorschau(zeichnung msepaket.Symbolzeichnung, groesse int) *image.NRGBA {
	return render.FaerbeSymbol(
		render.RastereSymbol(alsZeichenteile(zeichnung.Teile), groesse, groesse, Vorschaurand, false),
		render.Vollfilter{
			Fuellfarbe: color.NRGBA{A: 255},
			Randfarbe:  color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		})
}
