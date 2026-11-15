package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/svg"
)

// Kartenquelle gibt die Karten, aus denen Bilder werden.
type Kartenquelle interface {
	SetLeser

	KarteLesen(ctx context.Context, karteID int64) (*domain.CardRow, error)
}

// Planquelle macht aus einer Karte einen Zeichenplan.
//
// **Derselbe Weg wie die Vorschau** (#34, #206): zwei Wege zu derselben Karte
// laufen auseinander, sobald einer etwas lernt.
type Planquelle interface {
	Kartenplan(ctx context.Context, karte *domain.CardRow) (*render.Plan, []domain.Befund, error)
}

// Rasterquelle zeichnet einen Plan.
type Rasterquelle interface {
	Zeichne(
		ctx context.Context, plan *render.Plan, vergroesserung float64,
		guete render.Guete,
	) (
		*image.NRGBA, []render.Befund, error,
	)
}

// JPGGuete ist die Güte, mit der ein JPG geschrieben wird.
//
// **Neunzig und nicht das Höchstmaß.** Bei 100 speichert `image/jpeg` die
// Rundungsfehler mit und die Datei wird dreimal so groß, ohne dass jemand einen
// Unterschied sieht; darunter zeigen sich an harten Kanten — und eine
// Kartenvorlage besteht aus harten Kanten — sichtbare Ränder.
const JPGGuete = 90

// GrenzeZielbreite ist die größte Zielbreite in Bildpunkten.
//
// **Sie steht vor dem Rastern, nicht danach.** Wer 100 000 tippt — versehentlich
// oder aus Neugier —, bekäme bei einer Karte von 375 × 523 Punkten ein Bild mit
// vierzehn Milliarden Bildpunkten: 55 Gigabyte, angefordert in einem Stück. Der
// Rechner ist danach nicht langsam, sondern weg; der rote Lauf dieses Tests
// hat zwei Minuten lang nichts anderes getan.
//
// Zehntausend Bildpunkte sind bei 300 dpi ein Bogen von 85 Zentimetern Breite —
// mehr, als eine Karte je braucht.
const GrenzeZielbreite = 10_000

// Bildschreibung ist der Ausgang eines geschriebenen Kartenbilds.
type Bildschreibung struct {
	Pfad   string
	Breite int
	Hoehe  int

	// Befunde sind die Stellen, an denen dem Bild etwas fehlt — eine Schrift,
	// ein Bild. **Sie halten das Schreiben nicht auf** (§8): eine Karte mit elf
	// von zwölf Ebenen und einer Meldung ist brauchbar, eine leere ohne Meldung
	// nicht.
	Befunde []domain.Befund

	// Uebersprungen sagt, dass gar nichts geschrieben wurde — `altbehalten`
	// und eine Datei, die schon dalag
	// (#491).
	//
	// **Ein eigener Wert und kein leerer Pfad**: derselbe Grund wie beim
	// Abbrechen eines Dialogs — eine Oberfläche, die „nichts getan" von einem
	// Fehler nicht unterscheiden kann, zeigt eine Fehlermeldung für etwas, das
	// jemand ausdrücklich so wollte.
	Uebersprungen bool
}

// Schreibe rastert eine Karte und legt sie als Bild an einen Pfad.
//
// # Die Zielbreite statt eines Faktors
//
// Der Plan misst in **Vorlagenpunkten**; `render.Zeichne` nimmt eine
// Vergrößerung. Für eine Vorschau ist das die richtige Größe — sie *ist* ein
// Zoom. Für eine Ausgabe ist sie es nicht: „viermal" sagt nur etwas, wenn man
// die Größe der Vorlage kennt, und die steht in einer Datei, die niemand
// aufmacht.
//
// **Deshalb die Breite in Bildpunkten.** Sie ist nachrechenbar — 750 Punkte auf
// eine 63 mm breite Karte sind 300 dpi —, und die Rechnung gehört dem, der die
// Karte druckt. Eine Zusage in dpi könnten wir nicht halten: dafür müsste die
// Vorlage sagen, wie groß die Karte **in Millimetern** ist, und das sagt sie
// nicht ([0012](../../docs/entscheidungen/0012-aufloesung-als-zielbreite.md)).
//
// `0` heißt: so groß, wie der Plan es sagt.
func (b *Bildausfuhr) Schreibe(
	ctx context.Context, karteID int64, verzeichnis, datei string,
	format Bildformat, zielbreite int, verhalten string, guete render.Guete,
) (*Bildschreibung, error) {
	if err := b.darfSchreiben(ctx); err != nil {
		return nil, err
	}

	if err := PruefeBildziel(datei, format, zielbreite); err != nil {
		return nil, err
	}

	// **Der Namenskonflikt wird vor dem Zeichnen entschieden** (#491).
	//
	// `altbehalten` schreibt gar nicht — dann ist auch das Zeichnen umsonst,
	// und bei 3.700 Karten in Druckauflösung ist „umsonst" eine Zahl.
	//
	// **Die Reihenfolge ist zugleich MSEs**: `resolve_filename_conflicts`
	// steht vor `export_image` (`src/data/format/image.cpp`, Zeilen 87–91,
	// Fassung f9d9356 vom 2021-07-06).
	if err := os.MkdirAll(verzeichnis, 0o750); err != nil {
		return nil, fmt.Errorf("das Zielverzeichnis anlegen: %w", err)
	}

	datei, schreiben := Namenskonflikt(verzeichnis, datei, verhalten, nil)
	if !schreiben {
		// **Übersprungen ist kein Fehler.** MSE macht hier `continue`; die
		// Karte bleibt ungeschrieben, und die vorhandene Datei bleibt stehen —
		// genau das hat sich jemand mit `altbehalten` gewünscht.
		return &Bildschreibung{Uebersprungen: true}, nil
	}

	karte, err := b.karten.KarteLesen(ctx, karteID)
	if err != nil {
		return nil, fmt.Errorf("die Karte lesen: %w", err)
	}

	plan, planbefunde, err := b.bauer.Kartenplan(ctx, karte)
	if err != nil {
		return nil, fmt.Errorf("den Plan bauen: %w", err)
	}

	vergroesserung := 1.0
	if zielbreite > 0 && plan.Breite > 0 {
		vergroesserung = float64(zielbreite) / plan.Breite
	}

	// **Die Befunde des Plans zuerst** (#206): sie nennen den Grund, die des
	// Ausgabewegs die Folge.
	befunde := beiKarte(planbefunde, datei)
	pfad := filepath.Join(verzeichnis, datei)

	befunde = append(befunde, guetenbefund(guete, format, datei)...)

	// **SVG geht einen eigenen Weg** (#852): es entsteht aus dem Zeichenplan
	// und nicht aus Bildpunkten. Alles davor — Name, Konflikt, Grenze, Plan —
	// gilt für beide Wege; nur das Erzeugen der Bytes unterscheidet sich.
	if format == FormatSVG {
		return b.alsSVG(ctx, plan, pfad, datei, befunde, vergroesserung)
	}

	bild, rasterbefunde, err := b.rasterer.Zeichne(ctx, plan, vergroesserung, guete)
	if err != nil {
		return nil, fmt.Errorf("die Karte zeichnen: %w", err)
	}

	befunde = append(befunde, umschlagen(rasterbefunde, datei)...)

	gepackt, err := packe(bild, format)
	if err != nil {
		return nil, err
	}

	// **Erst packen, dann schreiben** — und dann erst an den Platz legen.
	// Ersteres fängt einen Fehler des **Kodierers**, letzteres einen des
	// **Dateisystems** ([schreibeAtomar], #853).
	if err := schreibeAtomar(pfad, gepackt); err != nil {
		return nil, fmt.Errorf("das Bild schreiben: %w", err)
	}

	return &Bildschreibung{
		Pfad:    pfad,
		Breite:  bild.Bounds().Dx(),
		Hoehe:   bild.Bounds().Dy(),
		Befunde: befunde,
	}, nil
}

// PruefeBildziel weist ein Bildziel ab, das nie geschrieben würde.
//
// # Eine Prüfung für zwei Wege
//
// **Der Dateiname ist eine Eingabe, auch wenn er von uns stammt.** Er kommt
// aus `Namen` und ist dort gesäubert — aber er reist durch die Oberfläche
// und kommt von dort zurück. Wer den Pfad erst im Aufrufer zusammensetzt,
// verlässt sich darauf, dass jeder Aufrufer das prüft.
//
// **[Bildausfuhr.Schreibe] ruft sie vor dem ersten Plattenzugriff, und der
// Probelauf einer Exportvorlage ruft dieselbe** (UXR-F7,
// #2015):
// ein `write_image_file`, das im Lauf an einem Namen scheiterte, scheitert im
// Probelauf mit demselben Satz.
func PruefeBildziel(datei string, format Bildformat, zielbreite int) error {
	if datei == "" || datei != filepath.Base(datei) || datei == "." || datei == ".." {
		return fmt.Errorf("%w: %q ist kein Dateiname", domain.ErrValidation, datei)
	}

	if format.Endung() == "" {
		return fmt.Errorf("%w: %q ist kein Bildformat", domain.ErrValidation, format)
	}

	if zielbreite > GrenzeZielbreite {
		return fmt.Errorf("%w: %d Bildpunkte Breite sind mehr als die Grenze von %d",
			domain.ErrValidation, zielbreite, GrenzeZielbreite)
	}

	return nil
}

// alsSVG schreibt den Plan als SVG.
//
// **Die Befunde reisen mit**, so wie beim Rasterer. Bis zum 2026-08-22 gab
// `svg.AusPlan` gar keinen Kanal dafür: eine SVG ohne Kartenrahmen entstand
// erfolgreich, und niemand erfuhr es
// (#852).
func (b *Bildausfuhr) alsSVG(
	ctx context.Context, plan *render.Plan, pfad, datei string,
	befunde []domain.Befund, vergroesserung float64,
) (*Bildschreibung, error) {
	quellen := b.vektor
	quellen.Vergroesserung = vergroesserung

	geschrieben, svgbefunde, err := svg.AusPlan(ctx, plan, quellen)
	if err != nil {
		return nil, fmt.Errorf("die Karte als SVG schreiben: %w", err)
	}

	befunde = append(befunde, umschlagen(svgbefunde, datei)...)

	if err := schreibeAtomar(pfad, []byte(geschrieben)); err != nil {
		return nil, fmt.Errorf("das SVG schreiben: %w", err)
	}

	// **Gemeldet wird die bestellte Größe**, nicht die des Plans — sie steht
	// so in der Datei, und die Oberfläche soll nichts anderes sagen.
	return &Bildschreibung{
		Pfad:    pfad,
		Breite:  int(plan.Breite*vergroesserung + 0.5),
		Hoehe:   int(plan.Hoehe*vergroesserung + 0.5),
		Befunde: befunde,
	}, nil
}

// # Warum es keinen Weg zurück gibt
//
// **Ein gerendertes Kartenbild lässt sich nicht wieder zu einer Karte machen**,
// und dasselbe gilt für PDF, SVG und HTML
// (#841). Es
// ist ein Raster von Farbwerten; welcher Bildpunkt zu `name` gehörte und
// welcher zu `rule text`, steht nirgends darin.
//
// Diese Ausfuhr ist damit eine **Einbahnstraße**, und das ist keine Lücke,
// sondern ihre Natur. Wer ein Set weitergeben und zurückbekommen will, gibt
// `.mse-set` oder MTG JSON weiter. Die Aufstellung je Format steht in
// `docs/FORMATE.md`.

// PackeBild macht aus einem Bild die Bytes seines Formats.
//
// **Ausgeführt, damit ein Test sie einzeln messen kann** (#826). Was ein
// Format mit einer durchsichtigen Ecke tut, ist eine Aussage über **diesen**
// Kodierer und nicht über das Format im Allgemeinen — sie gehört gemessen, und
// dafür braucht der Test einen Weg dorthin, der nicht über einen ganzen
// Ausfuhrlauf führt.
func PackeBild(bild image.Image, format Bildformat) ([]byte, error) {
	var (
		gepackt bytes.Buffer
		fehler  error
	)

	// **Ein Fehlerweg und nicht vier.**
	//
	// Jeder Zweig könnte seinen eigenen Rückweg tragen; dann wären es vier
	// Zweige, die kein Test betritt — ein Kodierer, der in einen Puffer
	// schreibt, scheitert an einem gültigen Bild nicht. Nachgemessen an einem
	// Bild ohne Fläche scheitern **PNG und TIFF**, JPEG und BMP nicht; mit
	// einem gemeinsamen Rückweg genügt einer dieser beiden Fälle, um ihn zu
	// betreten. Vier Wege bräuchten vier Fälle, von denen zwei gar nicht
	// herstellbar sind (§3).
	switch format {
	// **JPG und BMP kennen keine Durchsichtigkeit** und bekommen deshalb
	// beide einen weißen Grund. Ohne ihn käme eine durchsichtige Ecke — und
	// eine Kartenvorlage hat vier davon — als **Schwarz** heraus.
	case FormatJPG:
		fehler = jpeg.Encode(&gepackt, aufWeiss(bild), &jpeg.Options{Quality: JPGGuete})

	case FormatBMP:
		fehler = bmp.Encode(&gepackt, aufWeiss(bild))

	// **TIFF kann Alpha und behält es** — nachgemessen an einer
	// zurückgelesenen Ecke, nicht aus der Formatbeschreibung geschlossen.
	//
	// `Options` bleibt `nil`: das gibt die Vorgabe des Kodierers, und eine
	// Kompressionswahl ohne Anlass wäre eine Einstellung, die niemand trifft.
	case FormatTIFF:
		fehler = tiff.Encode(&gepackt, bild, nil)

	case FormatPNG:
		fehler = png.Encode(&gepackt, bild)

	// **SVG entsteht aus dem Plan und nicht aus Bildpunkten.** Der Zweig
	// nennt den Grund, statt „ist kein Bildformat" zu melden — was für ein
	// Format, das die Oberfläche anbietet, ratlos macht.
	case FormatSVG:
		return nil, fmt.Errorf(
			"%w: SVG entsteht aus dem Zeichenplan, nicht aus einem Bild",
			domain.ErrValidation)

	// **Abgewiesen, nicht geraten** — dieselbe Entscheidung wie in `Namen`
	// und `Schreibe`, und aus demselben Grund: unter welcher Endung wäre
	// geschrieben worden?
	//
	// Der Zweig ist von den Aufrufern aus **unerreichbar**; sie prüfen das
	// Format, bevor sie hierherkommen. Erreichbar ist er über den
	// ausgeführten Einstieg, und deshalb steht er hier statt als stiller
	// Rückfall auf PNG. Der erste Anlauf tat genau das — und zwei vorhandene
	// Tests wurden rot, weil sie diese Zusage schon hielten.
	default:
		return nil, fmt.Errorf("%w: %q ist kein Bildformat",
			domain.ErrValidation, format)
	}

	if fehler != nil {
		return nil, fmt.Errorf("das Bild als %s packen: %w", format, fehler)
	}

	return gepackt.Bytes(), nil
}

// packe ist der innere Name für [PackeBild].
func packe(bild *image.NRGBA, format Bildformat) ([]byte, error) {
	return PackeBild(bild, format)
}

// aufWeiss legt ein durchsichtiges Bild auf einen weißen Grund.
//
// **JPG kennt keine Durchsichtigkeit**, und `image/jpeg` schreibt die
// Farbwerte, ohne den Alphakanal zu beachten. Eine durchsichtige Ecke — und
// eine Kartenvorlage hat vier davon — käme als **Schwarz** heraus, weil ein
// unbeschriebenes NRGBA durchsichtiges Schwarz ist.
func aufWeiss(bild image.Image) *image.NRGBA {
	grund := image.NewNRGBA(bild.Bounds())
	draw.Draw(grund, grund.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(grund, grund.Bounds(), bild, bild.Bounds().Min, draw.Over)

	return grund
}

// umschlagen macht aus Rasterbefunden die Befunde der Fachschicht.
func umschlagen(befunde []render.Befund, datei string) []domain.Befund {
	if len(befunde) == 0 {
		return nil
	}

	umgeschlagen := make([]domain.Befund, 0, len(befunde))
	for _, befund := range befunde {
		umgeschlagen = append(umgeschlagen, domain.Befund{
			Datei:   datei,
			Ort:     fmt.Sprintf("Ebene %d", befund.Ebene),
			Meldung: befund.Meldung,
		})
	}

	return umgeschlagen
}
