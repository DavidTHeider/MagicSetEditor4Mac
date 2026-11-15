package render

import (
	"errors"
	"fmt"
	"image"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/vector"
)

// Glyphenschrift ist eine Schrift, die ihre Zeichen auch zeichnen kann.
//
// # Warum das eine zweite Schnittstelle ist
//
// `Schrift` misst — Breite und Zeilenhöhe —, und dafür genügt eine Tabelle. Der
// Umbruch aus #31 ist an einer solchen Tabelle geprüft (`festeBreite` gibt
// jedem Zeichen die Breite 1), und das soll so bleiben: ein Umbruchtest, der
// eine echte Schriftdatei braucht, prüft die Schrift mit.
//
// Wer zeichnen will, braucht mehr. Eine Schrift ohne Glyphen ist deshalb kein
// Fehler, sondern ein **Befund** — sie setzt, sie zeichnet nicht, und das steht
// dann an der Karte.
type Glyphenschrift interface {
	Schrift

	// Grundlinie ist der Abstand von der Oberkante der Zeile zur Grundlinie,
	// bei Schriftgröße 1.
	Grundlinie() float64

	// Maske gibt die Deckung eines Zeichens bei dieser Größe — **wie Quartz sie
	// in einem Bitmapkontext zeichnet**, samt Schriftglättung (#1806). Die
	// Maske ist für MSEs vierfachen Weg; wer eine ungeglättete braucht, braucht
	// eine andere Frage.
	//
	// `ecke` ist die linke obere Ecke der Maske **relativ zum Ursprung auf der
	// Grundlinie**; `da` ist falsch, wenn die Schrift dieses Zeichen nicht
	// führt. Ein fehlendes Zeichen ist ein Befund, kein Absturz — fremde Texte
	// sind der Normalfall (AGENTS.md §8).
	Maske(zeichen rune, groesse float64) (maske image.Image, ecke image.Point, da bool)
}

// OpenTypeSchrift ist eine gelesene Schriftdatei.
//
// Der Rasterer dahinter ist `golang.org/x/image`
// ([0008](../../docs/entscheidungen/0008-schriftrasterung.md)). Go bringt keinen
// mit, und die einzige Alternative — Core Text über cgo — beendete die
// Kreuzkompilierung aus §12.
type OpenTypeSchrift struct {
	schrift *sfnt.Font
	einheit float64

	// zeilenhoehe und grundlinie werden **einmal** gelesen.
	//
	// Sie ändern sich nicht, und jede Abfrage ginge sonst durch dieselbe
	// Metrikabfrage — samt ihrem Fehlerpfad. Eine Schrift, deren Metrik nicht
	// zu lesen ist, wird beim Laden abgewiesen, nicht bei jedem Zeichen.
	zeilenhoehe, grundlinie float64

	// kern sind die Paare der Schrift — leer, wenn sie keine führt
	// (#752).
	kern kernung

	// liga sind die Ligaturen der Schrift — leer, wenn sie keine führt
	// (#1805).
	liga ligaturen

	// raster ist der Rasterer der Umrisse ([OpenTypeSchrift.maskeDerKennung]) —
	// **einer je Schrift**, wie der Puffer, und unter derselben Sperre.
	raster vector.Rasterizer

	// puffer ist der Arbeitsspeicher von `sfnt` — er gehört **einer** Schrift
	// und darf nicht nebenläufig benutzt werden.
	puffer sfnt.Buffer

	// schloss deckt Puffer und Rasterer.
	//
	// **Der Kommentar über dem Puffer stand vor dem Schloss da** — und die
	// Anwendung stürzte trotzdem ab: eine Vorschau wird abgebrochen, nicht
	// abgewartet (§7), also laufen immer mehrere, und alle fragen dieselbe
	// Schrift. „Darf nicht nebenläufig benutzt werden" ist eine Zusage an den
	// Aufrufer; wer sie nicht durchsetzt, hat sie aufgeschrieben und nicht
	// gehalten.
	schloss sync.Mutex
}

// SchriftAusDatei liest eine Schriftdatei.
//
// # Sammlungen
//
// Eine `.ttc` führt mehrere Schriften in einer Datei — auf macOS liegt der
// halbe Grundbestand so, `Helvetica.ttc` und `HelveticaNeue.ttc` unter ihnen.
// Gelesen wird die **erste**; sie ist die Grundschnittform der Familie.
//
// **Was das nicht kann:** eine bestimmte Schnittform aus der Sammlung wählen.
// `Helvetica Neue Bold` liegt in derselben Datei und ist von hier aus nicht
// erreichbar. Das ist eine benannte Lücke — wer sie schließt, braucht die
// Namenstabelle jeder Schrift in der Sammlung.
func SchriftAusDatei(roh []byte) (*OpenTypeSchrift, error) {
	gelesen, err := lesen(roh)
	if err != nil {
		return nil, err
	}

	schrift := &OpenTypeSchrift{
		schrift: gelesen,
		einheit: float64(gelesen.UnitsPerEm()),
		kern:    kernungAus(roh),
		liga:    ligaturenAus(roh, gelesen.NumGlyphs()),
	}

	masse, err := masseLesen(gelesen, &schrift.puffer, fixed.I(int(schrift.einheit)))
	if err != nil {
		return nil, fmt.Errorf("die Schriftdatei nennt keine brauchbare Metrik: %w", err)
	}

	schrift.zeilenhoehe = float64(masse.Height) / 64 / schrift.einheit
	schrift.grundlinie = float64(masse.Ascent) / 64 / schrift.einheit

	if err := ohneHhea(schrift, roh); err != nil {
		return nil, err
	}

	return schrift, nil
}

// ohneHhea füllt die Metrik, wenn `hhea` sie nicht hergibt — in zwei Stufen.
//
// # Warum zwei und nicht eine
//
// **Die Ränder sind der letzte Rückfall, nicht der erste.** Sie sind der
// **Tintenumriss** und damit größer als der typografische Ascent; bei
// `MATRIX.TTF` geben sie 0,957 em Zeilenhöhe, wo wx mit **0,860** rechnet, und
// 0,779 Grundlinie, wo wx **0,633** nimmt — gemessen über zwanzig Punktgrößen
// (#597).
//
// Dazwischen gehört `OS/2` typo, und genau dort stehen die beiden Werte, die
// wx trifft. Erst wenn auch die Tabelle nichts hergibt, wird geraten.
//
// **Beide Stufen sind `adapted`** (§15): im MSE-Quelltext steht diese Regel
// nicht, denn MSE fragt den Schriftdienst des Betriebssystems. Nachgebaut ist
// **dessen** Rückfallkette.
func ohneHhea(schrift *OpenTypeSchrift, roh []byte) error {
	if schrift.grundlinie > 0 || schrift.zeilenhoehe > 0 {
		return nil
	}

	if metrik, da := typoAus(roh); da {
		schrift.grundlinie = metrik.Grundlinie
		schrift.zeilenhoehe = metrik.Zeilenhoehe

		return nil
	}

	return ausRaendern(schrift)
}

// ausRaendern holt die Metrik aus den Rändern, wenn auch `OS/2` keine nennt.
//
// **Das ist Raten, und es steht hier, damit man es sieht.** Der Tintenumriss
// ist nicht die typografische Metrik; er ist nur besser als eine Zeilenhöhe
// von null, bei der jeder Zug über seinem Kasten stünde.
//
// # Der Befund
//
// `x/image/font/sfnt` gibt für `MATRIX.TTF` aus `Full-Magic-Pack`
// **ascent = 0, descent = 0, height = 0** — und `err = nil`. Betroffen sind
// `MATRIX.TTF` und `MatrixBold.ttf`; `mplantin.ttf`, `beleren-bold` und
// `ModMatrix.ttf` nennen ihre Metrik. Dieselben Dateien liefern brauchbare
// Ränder: `bounds.Min.Y = -49856` bei `em = 1000`, also 0,78 em über der
// Grundlinie.
//
// # Warum das die Karte trifft
//
// Die Grundlinie ist `Zeilenoberkante + Größe × Ascent`. Mit Ascent null fällt
// sie auf die Oberkante, und **jeder Zug in dieser Schrift steht über seinem
// Kasten** — auf der Archenemy-Karte sieben Bildpunkte zu hoch, gemessen gegen
// MSEs eigene Ausgabe
// (#417).
//
// # Warum das **adapted** ist und nicht **ported**
//
// MSE holt seine Metrik vom Schriftdienst des Betriebssystems (`wxFont`), nicht
// aus der Datei; dort stellt sich die Frage nicht. Diese Rechnung hat im
// Altcode keine Entsprechung — sie ersetzt, was uns die Bibliothek schuldig
// bleibt (#418,
// `docs/PORTIERUNGSKARTE.md`).
//
// **Nur wenn beides null ist.** Eine Schrift mit gültiger Metrik wird nicht
// angefasst — die Ränder sind der **Tintenumriss** und größer als der
// typografische Ascent (bei `mplantin` 1,045 gegen 0,774 em).
//
// **Ohne eigenen Vorabtest**: dass hier noch nichts steht, hat [ohneHhea]
// schon geprüft. Ein zweiter Test wäre ein Zweig, den kein Aufruf betritt (§3).
func ausRaendern(schrift *OpenTypeSchrift) error {
	raender, err := raenderLesen(schrift.schrift, &schrift.puffer,
		fixed.I(int(schrift.einheit)))
	if err != nil {
		return fmt.Errorf("die Schriftdatei nennt weder Metrik noch Ränder: %w", err)
	}

	schrift.grundlinie = float64(-raender.Min.Y) / 64 / schrift.einheit
	schrift.zeilenhoehe = float64(raender.Max.Y-raender.Min.Y) / 64 / schrift.einheit

	if schrift.grundlinie <= 0 || schrift.zeilenhoehe <= 0 {
		return errOhneMetrik
	}

	return nil
}

// errOhneMetrik meldet eine Schrift, die auch keine Ränder hergibt.
var errOhneMetrik = errors.New("die Schriftdatei nennt weder Metrik noch Ränder")

// lesen zerteilt eine Schriftdatei oder eine Sammlung.
func lesen(roh []byte) (*sfnt.Font, error) {
	// `ttcf` ist die Kennung einer Sammlung — vier Bytes am Anfang der Datei.
	if len(roh) >= 4 && string(roh[:4]) == "ttcf" {
		sammlung, err := sfnt.ParseCollection(roh)
		if err != nil {
			return nil, fmt.Errorf("die Schriftsammlung ließ sich nicht lesen: %w", err)
		}

		gelesen, err := sammlung.Font(0)
		if err != nil {
			return nil, fmt.Errorf("die erste Schrift der Sammlung fehlt: %w", err)
		}

		return gelesen, nil
	}

	gelesen, err := sfnt.Parse(roh)
	if err != nil {
		return nil, fmt.Errorf("die Schriftdatei ließ sich nicht lesen: %w", err)
	}

	return gelesen, nil
}

// Die Nähte zur Schriftbibliothek.
//
// **Als Veränderliche und aus demselben Grund wie `oeffnen` in `internal/app`:**
// die Fehlerpfade dieser Aufrufe setzen eine Schriftdatei voraus, die zerteilbar
// ist und trotzdem keine Metrik hat. Mit einer echten Datei ist das nicht
// herzustellen — und ein Zweig, den kein Test betritt, ist eine Behauptung
// darüber, was im Ernstfall geschieht.
var (
	masseLesen = func(schrift *sfnt.Font, puffer *sfnt.Buffer, groesse fixed.Int26_6) (
		font.Metrics, error,
	) {
		return schrift.Metrics(puffer, groesse, font.HintingNone)
	}

	raenderLesen = func(schrift *sfnt.Font, puffer *sfnt.Buffer, groesse fixed.Int26_6) (
		fixed.Rectangle26_6, error,
	) {
		return schrift.Bounds(puffer, groesse, font.HintingNone)
	}

	vorschubLesen = func(schrift *sfnt.Font, puffer *sfnt.Buffer, kennung sfnt.GlyphIndex,
		groesse fixed.Int26_6,
	) (fixed.Int26_6, error) {
		return schrift.GlyphAdvance(puffer, kennung, groesse, font.HintingNone)
	}

	umrissLesen = func(schrift *sfnt.Font, puffer *sfnt.Buffer, kennung sfnt.GlyphIndex,
		massstab fixed.Int26_6,
	) (sfnt.Segments, error) {
		return schrift.LoadGlyph(puffer, kennung, massstab, nil)
	}
)

// Breite gibt die Vorschubbreite eines Zeichens bei Schriftgröße 1.
func (s *OpenTypeSchrift) Breite(zeichen rune) float64 {
	s.schloss.Lock()
	defer s.schloss.Unlock()

	kennung, da := s.kennung(zeichen)
	if !da {
		return 0
	}

	vorschub, err := vorschubLesen(s.schrift, &s.puffer, kennung, fixed.I(int(s.einheit)))
	if err != nil {
		return 0
	}

	return float64(vorschub) / 64 / s.einheit
}

// Zeilenhoehe gibt den Abstand zweier Grundlinien bei Schriftgröße 1.
func (s *OpenTypeSchrift) Zeilenhoehe() float64 { return s.zeilenhoehe }

// Grundlinie gibt den Abstand von der Oberkante zur Grundlinie bei Größe 1.
func (s *OpenTypeSchrift) Grundlinie() float64 { return s.grundlinie }

// Kernung ist die Verschiebung zwischen zwei Zeichen, bei Schriftgröße 1 —
// [Kernschrift].
//
// **In denselben Einheiten wie [OpenTypeSchrift.Breite]**: Schrifteinheiten
// durch `unitsPerEm`. Eine Schrift ohne Paare gibt durchweg null, und der
// Aufrufer merkt keinen Unterschied.
//
// **Unter derselben Sperre wie die Breite**: `GlyphIndex` benutzt den Puffer
// der Schrift, und der gehört einer.
func (s *OpenTypeSchrift) Kernung(vorher, jetzt rune) float64 {
	if len(s.kern.nachschlaege) == 0 {
		return 0
	}

	s.schloss.Lock()
	defer s.schloss.Unlock()

	erste, da := s.kennung(vorher)
	if !da {
		return 0
	}

	zweite, da := s.kennung(jetzt)
	if !da {
		return 0
	}

	//nolint:gosec // Glyphennummern von `sfnt` passen in uint16
	return float64(s.kern.fuer(uint16(erste), uint16(zweite))) / s.einheit
}
