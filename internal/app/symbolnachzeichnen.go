// **Ein Bild als Symbol nachzeichnen** —
// #1340,
// Werkstück F von Phase 4.11.
//
// Portiert aus `import_symbol` (`src/data/format/image_to_symbol.cpp`, Zeilen
// 224–237, Fassung `f9d9356` vom 2021-07-06); die Rechnung steht in
// `internal/service`.
//
// # Warum das an alle vier Flächen gehört
//
// Die übrigen Symbolfähigkeiten stehen mit Grund allein am Fenster: eine
// Grundform, ein Punktzug, eine Symmetrie — das sind Handgriffe auf einer
// Zeichenfläche, und ohne sie gäbe es nichts, worauf man zeigt.
//
// **Nachzeichnen ist keiner.** Es ist eine Umrechnung: Bild hinein, Zeichnung
// heraus. Wer hundert Symbole aus PNG-Dateien machen will, tut das an der
// Kommandozeile und nicht mit hundert Klicks (§16).

package app

import (
	"context"
	"fmt"
	"image"
	"os"

	// **Die drei Bildformate, die auch `BildSetzen` liest.** Ein Format, das
	// hier ginge und dort nicht, wäre eine Falle.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// SymbolNachzeichnen macht aus einer Bilddatei die Teile einer Zeichnung.
//
// **Sie legt nichts ab.** Was dabei herauskommt, ist ein Vorschlag: im Fenster
// steht er zum Weiterbearbeiten in der Zeichenfläche, und erst `SymbolSetzen`
// schreibt ihn ins Set. Wer beides in einem Zug will, ruft beide — so tun es
// Kommandozeile, MCP und Dienst.
func (a *App) SymbolNachzeichnen(ctx context.Context, pfad string) SymbolAntwort {
	ctx = a.imProfil(ctx)

	teile, err := Nachgezeichnet(pfad)
	if err != nil {
		return SymbolAntwort{Teile: []SymbolteilDTO{}, Fehler: Uebersetze(ctx, err)}
	}

	return SymbolAntwort{Teile: teile}
}

// Nachgezeichnet ist dasselbe mit einem `error` statt eines [Fehler].
//
// **Zwei Formen derselben Sache, und das ist der Schnitt der Schicht.** Die
// Übersetzung nach [Fehler] gehört an die Naht zum Fenster (§8); wer eine
// Kette weiterbaut — nachzeichnen und ablegen —, braucht den Fehler noch als
// Fehler, um ihn weiterzureichen.
func Nachgezeichnet(pfad string) ([]SymbolteilDTO, error) {
	gemalt, err := bildLesen(pfad)
	if err != nil {
		return nil, err
	}

	teile, err := service.BildNachzeichnen(gemalt)
	if err != nil {
		return nil, err
	}

	return teileAlsDTO(teile), nil
}

// bildLesen holt ein Bild von der Platte.
//
// **Der Fehler nennt den Pfad.** Eine Meldung „das Bild ließ sich nicht lesen"
// ohne Angabe, welches, ist bei einem Aufruf über die Kommandozeile mit einem
// Verzeichnis voller Dateien wertlos (§8).
func bildLesen(pfad string) (image.Image, error) {
	if pfad == "" {
		return nil, fmt.Errorf("%w: es ist keine Bilddatei genannt", domain.ErrValidation)
	}

	roh, err := os.Open(pfad) // #nosec G304 -- ein Pfad, den die Nutzenden selbst nennen
	if err != nil {
		return nil, domain.MitGrund(domain.ErrNotFound, err,
			"die Bilddatei %s ließ sich nicht öffnen", filepath.Base(pfad))
	}

	defer func() { _ = roh.Close() }() //nolint:errcheck // gelesen ist gelesen

	gemalt, _, err := msepaket.DekodiereBildAus(roh)
	if err != nil {
		return nil, msepaket.ZuGrossOder(err, domain.MitGrund(domain.ErrValidation, err,
			"%s ist kein lesbares Bild — erwartet werden PNG, JPEG oder GIF",
			filepath.Base(pfad)))
	}

	return gemalt, nil
}
