package app

// **Ein Bildfeld füllen** —
// #246.
//
// # Der Weg, den MSE geht
//
//	String filename = wxFileSelector(_("Open image file"), …);
//	wxImage image(filename);
//	sliceImage(image);
//	  → ImageSliceWindow s(…, image, style().getSize(), mask);
//	  → LocalFileName f = getLocalPackage().newFileName(field().name, "");
//	  → img.SaveFile(getLocalPackage().nameOut(f), wxBITMAP_TYPE_PNG);
//	  → addAction(value_action(valueP(), f));
//
// `ImageValueEditor::onLeftDClick` und `sliceImage`, `src/gui/value/image.cpp`,
// Zeilen 22–53, Fassung f9d9356 vom 2021-07-06: **wählen, zuschneiden,
// ablegen, den Namen in das Feld schreiben.**
//
// Der **Name** steht im Set, nicht der Pfad. Ein Pfad zeigte auf diesen
// Rechner; weitergegeben öffnete das Set mit leerem Kunstfenster.

import (
	"context"
	"fmt"
	"image"
	"os"

	// Für `image.Decode`: die drei Formate, die MSEs Dateidialog anbietet
	// (`src/gui/value/image.cpp`, Zeile 23, Fassung f9d9356 vom 2021-07-06) —
	// BMP kennt die Standardbibliothek nicht, und TIFF auch nicht.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Bildablage ist, was `app` zum Ablegen eines Bildes braucht.
type Bildablage interface {
	Ablegen(setID int64, feld string, bild image.Image) (string, error)

	// Wegnehmen entfernt ein eben abgelegtes Bild wieder
	// (#1163,
	// Befund Z7-02).
	//
	// **Für den Fehlerweg und sonst nichts.** Das Bild liegt, bevor ein Feld
	// auf es zeigt; scheitert das Feld, wäre es eine Datei ohne Verweis.
	Wegnehmen(setID int64, name string) error

	// Verzeichnis gibt den Ort der Bilder eines Sets — leer, wenn es keinen
	// gibt (A12, #1451).
	//
	// **Hier und nicht über `bilderort`**: die Ablage weiß, wie sie ihre
	// Verzeichnisse benennt, und eine zweite Rechnung daneben liefe mit ihr
	// auseinander, sobald jemand eine davon ändert.
	Verzeichnis(setID int64) string
}

// MitBildablage setzt die Naht zur Bildablage.
func (a *App) MitBildablage(ablage Bildablage) { a.bildablage = ablage }

// bildziel prüft Kennung und Feld und gibt die Fläche, auf die zugeschnitten
// wird.
//
// **Eine Stelle für beide Wege**: der Dialog und das Ziehen stellen dieselben
// Fragen, und zwei Fassungen liefen beim ersten Sonderfall auseinander.
func (a *App) bildziel(ctx context.Context, karteID, feld string) (
	kennung int64, karte *domain.CardRow, breite, hoehe int, befund *Fehler,
) {
	kennung, befund = kennungAus(karteID)
	if befund != nil {
		return 0, nil, 0, 0, befund
	}

	if fehler := domain.PruefeFeldname(feld); fehler != nil {
		return 0, nil, 0, 0, Uebersetze(ctx, fehler)
	}

	// **Ohne Bestand und ohne Planbauer gibt es keine Fläche** — und eine
	// Meldung statt eines Absturzes.
	//
	// **Gefunden am 2026-09-12** (A12, #1451): der Bildweg über einen
	// KI-Vorschlag ruft dieselbe Stelle, und ein Lauf ohne Planbauer ließ
	// sie in einen Nullzeiger laufen. Der Zweig fehlte hier, seit es diese
	// Funktion gibt; gesehen hat ihn niemand, weil die beiden anderen Wege
	// nur aus einem fertig verdrahteten Fenster kommen.
	if a.kartenleser == nil || a.planbauer == nil {
		return 0, nil, 0, 0, Uebersetze(ctx, fehlenderBestand())
	}

	karte, err := a.kartenleser.KarteLesen(ctx, kennung)
	if err != nil {
		return 0, nil, 0, 0, Uebersetze(ctx, err)
	}

	breite, hoehe, err = a.feldgroesse(ctx, karte, feld)
	if err != nil {
		return 0, nil, 0, 0, Uebersetze(ctx, err)
	}

	return kennung, karte, breite, hoehe, nil
}

// feldgroesse gibt die Fläche, auf die ein Bild zugeschnitten wird.
//
// **Aus dem Plan und nicht aus dem Feldstil.** MSE nimmt `style().getSize()`
// (`src/gui/value/image.cpp`, Zeile 44, Fassung f9d9356 vom 2021-07-06) — und
// die Größe eines Feldes ist ein
// Ergebnis, kein Wert: `left:` und `width:` sind Skripte, die die Karte lesen.
// Der Plan hat sie ausgerechnet ([render.Plan.Feldflaechen]).
func (a *App) feldgroesse(ctx context.Context, karte *domain.CardRow, feld string) (
	breite, hoehe int, fehler error,
) {
	plan, befunde, err := a.planbauer.Kartenplan(ctx, karte)
	if err != nil {
		return 0, 0, err
	}

	// **Ein Plan ohne jede Feldfläche ist ein Behelfsplan** —
	// #1055.
	//
	// Der Planbauer wirft keinen Fehler, wenn die Vorlage fehlt: er gibt einen
	// Ersatz und **einen Befund**. Unter `_` verworfen, käme „das Feld liegt
	// auf dieser Karte nicht" an — ohne Schlüssel übersetzt zu „Es ist ein
	// unerwarteter Fehler aufgetreten".
	//
	// **Der Unterschied zählt**: „dieses Feld gibt es hier nicht" ist eine
	// Aussage über die Karte, „die Vorlage fehlt" eine über den Rechner. Wer
	// die erste liest, sucht am falschen Ort.
	if len(plan.Feldflaechen) == 0 && len(befunde) > 0 {
		return 0, 0, domain.NewFieldError(feld, domain.SchluesselVorlageFehlt,
			"für diese Karte ist keine Vorlage eingebunden: %s",
			befunde[0].Meldung)
	}

	for i := range plan.Feldflaechen {
		if plan.Feldflaechen[i].Feld != feld {
			continue
		}

		rahmen := plan.Feldflaechen[i].Rahmen
		if !render.Sichtbar(rahmen) {
			break
		}

		return int(rahmen.Breite), int(rahmen.Hoehe), nil
	}

	return 0, 0, fmt.Errorf(
		"%w: das Feld %q liegt auf dieser Karte nicht — es gibt keine Fläche, "+
			"auf die ein Bild zugeschnitten würde", domain.ErrValidation, feld)
}

// bilddateiLesen öffnet eine Bilddatei und entschlüsselt sie.
func bilddateiLesen(pfad string) (image.Image, error) {
	// #nosec G304 -- aus dem Dateidialog oder, über Dienst und Werkzeugserver, hinter dem Pfadvertrag (docs/G304INVENTAR.md)
	datei, err := os.Open(pfad)
	if err != nil {
		return nil, fmt.Errorf("%w: die Bilddatei ließ sich nicht öffnen: %s",
			domain.ErrValidation, err.Error())
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // gelesen ist gelesen

	bild, _, err := msepaket.DekodiereBildAus(datei)
	if err != nil {
		return nil, msepaket.ZuGrossOder(err, fmt.Errorf("%w: die Datei ist kein lesbares Bild: %s",
			domain.ErrValidation, err.Error()))
	}

	return bild, nil
}
