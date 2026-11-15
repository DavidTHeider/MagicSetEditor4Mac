package service

// **Der Abzug eines Tisches als PDF** — PLAY-05.AK-03
// (#1713):
// „Screenshot, Bildexport und **PDF** … mit Auflösung, **Seitenformat** und
// Zuschnitt."
//
// # Warum derselbe Dienst und keine zweite Naht
//
// **Weil es dieselbe Frage mit einer anderen Hülle ist**: ein fertiges Bild
// aus dem Fenster geht an einen Ort. Was dazukommt, ist das Blatt — und wohin
// ein Bild darauf gehört, rechnet [domain.PlatzAufSeite].
//
// # Was der Zuschnitt ist
//
// **Der gewählte Ausschnitt**, und der entsteht im Fenster
// (`tisch/tischausschnitt.ts`): sichtbarer Bereich, ganzer Tisch oder eine
// Zone. Hier wird nichts mehr beschnitten — ein zweiter Zuschnitt an dieser
// Stelle wäre einer, den niemand gesehen hat.

import (
	"fmt"
	"image"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/pdf"
)

// Abzugsrand ist der Abstand zum Papierrand in Millimetern.
//
// **Zehn, wie beim Druckbogen** — kein Bürodrucker druckt bis an die Kante,
// und ein Abzug, dessen Rand abgeschnitten wird, sieht aus wie ein Fehler.
const Abzugsrand = 10

// SchreibePDF legt den Abzug als einseitiges PDF an den Pfad.
func (Tischbildausfuhr) SchreibePDF(datenURL, pfad, blatt string) error {
	if pfad == "" {
		return fmt.Errorf("%w: es fehlt der Ort", domain.ErrValidation)
	}

	papier, gibtEs := domain.BlattZu(blatt)
	if !gibtEs {
		return fmt.Errorf("%w: %q ist kein Papierformat",
			domain.ErrValidation, blatt)
	}

	rohes, fehler := domain.DeuteAbzug(datenURL)
	if fehler != nil {
		return fehler
	}

	// **Erst hier wird das Bild wirklich gelesen.** `DeuteAbzug` prüft die
	// acht Bytes der Signatur; was danach kommt, weiß nur der Dekoder (§8).
	bild, _, fehler := msepaket.DekodiereBild(rohes)
	if fehler != nil {
		return fmt.Errorf("den Abzug lesen: %w", fehler)
	}

	if fehler := schreibeAtomar(pfad, seiteMit(bild, papier)); fehler != nil {
		return fmt.Errorf("das PDF schreiben: %w", fehler)
	}

	return nil
}

// seiteMit baut das einseitige Dokument.
func seiteMit(bild image.Image, papier domain.Blatt) []byte {
	masse := bild.Bounds()

	// **Das Blatt dreht sich nach dem Bild** — ein breiter Tisch auf ein
	// stehendes A4 gelegt füllte ein Drittel der Seite.
	gedreht := domain.BlattFuer(masse.Dx(), masse.Dy(), papier)

	platz := domain.PlatzAufSeite(masse.Dx(), masse.Dy(), gedreht, Abzugsrand)

	dokument := pdf.NeuesDokument()
	seite := dokument.NeueSeite(pdf.MM(gedreht.Breite), pdf.MM(gedreht.Hoehe))

	seite.Bild(bild,
		pdf.MM(platz.Links), pdf.MM(platz.Unten),
		pdf.MM(platz.Breite), pdf.MM(platz.Hoehe))

	return dokument.Bytes()
}
