package service

import (
	"context"
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/pdf"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Bogenmasse beschreibt Papier, Karte und Raster — alles in Millimetern.
//
// # Warum hier dpi steht und bei der Bildausgabe nicht
//
// [0012](../../docs/entscheidungen/0012-aufloesung-als-zielbreite.md) lehnt eine
// dpi-Angabe ab, weil keine Vorlage sagt, wie groß eine Karte in Millimetern
// ist. **Auf einem Druckbogen sagt es das Papier**: eine Karte ist 63 × 88 mm,
// weil sie genau so geschnitten wird. Damit ist dpi rechenbar und die richtige
// Angabe — dieselbe Entscheidung, andere Lage.
type Bogenmasse struct {
	Seitenbreite float64
	Seitenhoehe  float64

	Kartenbreite float64
	Kartenhoehe  float64

	// Anschnitt ist, was das Bild **über** die Schnittkante hinausreicht.
	//
	// Er ist der Grund, warum eine geschnittene Karte keinen weißen Rand hat.
	// Die Karten liegen deshalb mit dem doppelten Anschnitt Abstand: sonst
	// überschriebe der Anschnitt der einen den der anderen, und er wäre nur am
	// Rand des Rasters wirksam.
	Anschnitt float64

	// Rand ist der Mindestabstand zum Papierrand.
	Rand float64

	// DPI ist die Auflösung, mit der eine Karte gerastert wird.
	DPI int

	Schnittmarken bool
}

// StandardBogen ist A4 mit neun Karten, ohne Anschnitt, in Druckauflösung.
func StandardBogen() Bogenmasse {
	return Bogenmasse{
		Seitenbreite: 210, Seitenhoehe: 297,
		Kartenbreite: 63, Kartenhoehe: 88,
		Anschnitt: 0, Rand: 10, DPI: 300,
		Schnittmarken: true,
	}
}

// GrenzeDPI ist die höchste Auflösung eines Bogens.
//
// **Sechshundert und nicht mehr.** Eine Karte bei 600 dpi ist 1488 × 2079
// Bildpunkte, also 9 MB roh; ein Set mit dreihundert Karten liegt damit als
// gepackter Strom im Arbeitsspeicher, bis das PDF geschrieben ist. Darüber
// wächst es schneller, als es jemand nutzen kann — 600 dpi druckt kein
// Bürodrucker mehr aus.
const GrenzeDPI = 600

// grenzeVerzerrung ist die Abweichung im Seitenverhältnis, ab der es auffällt.
//
// **Ein Hundertstel.** Der Behelfsplan misst 375 × 523 und weicht damit um
// 0,15 % von 63 × 88 ab; wer das meldete, meldete es bei jeder echten Vorlage,
// und ein Befund, der immer kommt, wird nicht gelesen.
const grenzeVerzerrung = 0.01

// Bogenlauf ist der Ausgang eines gelegten Bogens.
type Bogenlauf struct {
	Pfad   string
	Seiten int
	Karten int

	Befunde []domain.Befund
}

// Bogenleger setzt Karten maßhaltig auf Druckseiten.
type Bogenleger struct {
	wache
	haushaltswahl

	karten   Kartenquelle
	bauer    Planquelle
	rasterer Rasterquelle
}

// NeuerBogenleger baut den Leger über Bestand, Planbau und Rasterer.
func NeuerBogenleger(karten Kartenquelle, bauer Planquelle, rasterer Rasterquelle) *Bogenleger {
	return &Bogenleger{karten: karten, bauer: bauer, rasterer: rasterer}
}

// MitPruefer setzt den Riegel, der vor dem Schreiben fragt, ob dieser Aufruf
// es darf (0078).
func (b *Bogenleger) MitPruefer(pruefer Pruefer) *Bogenleger {
	b.pruefer = pruefer

	return b
}

// Lege schreibt die Karten eines Sets als Druckbogen an einen Pfad.
//
// # Maßhaltig heißt: die Karte bestimmt die Größe, nicht die Vorlage
//
// Das Bild wird auf 63 × 88 mm gezogen, gleich welches Seitenverhältnis der
// Plan hat. Eine Vorlage, die deutlich daneben liegt, wird dabei verzerrt —
// das steht als Befund daneben (§8). Die Gegenrechnung wäre ein weißer Rand an
// zwei Seiten, und der ist beim Schneiden schlimmer als ein Prozent Verzerrung.
// # Und eine Teilmenge geht auch
//
// **`auswahl` leer heißt: das ganze Set** — MSE2s `EXPORT_SEL_ENTIRE_SET`, und
// es ist der Nullwert ([domain.Kartenauswahl]). Wer eine Auswahl mitgibt,
// bekommt genau diese Karten, in der Reihenfolge des Sets.
//
// **Die Prüfung steht vor dem ersten Schreiben.** Eine Auswahl, die keine
// Karte trifft, gäbe sonst eine PDF-Datei mit null Seiten — und die ist
// verwirrender als ein Fehler.
func (b *Bogenleger) Lege(
	ctx context.Context, setID int64, pfad string, masse Bogenmasse,
	auswahl domain.Kartenauswahl,
) (*Bogenlauf, error) {
	if err := b.darfSchreiben(ctx); err != nil {
		return nil, err
	}

	freigeben, platzfehler := b.platz(ctx, domain.RessourceAusfuhr)
	if platzfehler != nil {
		return nil, platzfehler
	}

	defer freigeben()

	plan, err := b.plane(ctx, setID, masse, auswahl)
	if err != nil {
		return nil, err
	}

	karten, raster, jeSeite := plan.karten, plan.raster, plan.jeSeite
	dokument := pdf.NeuesDokument()

	lauf := &Bogenlauf{Pfad: pfad, Karten: len(karten)}

	for anfang := 0; anfang < len(karten); anfang += jeSeite {
		aufSeite := karten[anfang:min(anfang+jeSeite, len(karten))]

		seite := dokument.NeueSeite(pdf.MM(masse.Seitenbreite), pdf.MM(masse.Seitenhoehe))
		lauf.Seiten++

		for platz, karte := range aufSeite {
			// **Vor jeder Karte die Frage, ob es weitergeht** (FMT-03): der
			// Bogen steht bis zum Schluss nur im Speicher, ein Abbruch hier
			// hinterlässt nichts.
			if fehler := Schritt(ctx, domain.Fortschritt{
				Erledigt: anfang + platz, Gesamt: len(karten), Einheit: domain.EinheitKarte,
			}); fehler != nil {
				return nil, fehler
			}

			bild, befunde, fehler := b.zeichne(ctx, karte, masse)
			if fehler != nil {
				return nil, fehler
			}

			lauf.Befunde = append(lauf.Befunde, befunde...)

			links, unten := raster.stelle(platz, masse)
			seite.Bild(bild,
				links-pdf.MM(masse.Anschnitt), unten-pdf.MM(masse.Anschnitt),
				pdf.MM(masse.Kartenbreite+2*masse.Anschnitt),
				pdf.MM(masse.Kartenhoehe+2*masse.Anschnitt))
		}

		// **Die Marken zuletzt und nur für die belegten Plätze.** Eine Seite mit
		// einer Karte trüge sonst die Marken für neun — und wer sie schneidet,
		// sucht acht Karten, die es nicht gibt.
		if masse.Schnittmarken {
			raster.marken(seite, masse, len(aufSeite))
		}
	}

	if err := Schritt(ctx, domain.Fortschritt{
		Erledigt: len(karten), Gesamt: len(karten), Einheit: domain.EinheitKarte,
	}); err != nil {
		return nil, err
	}

	if err := schreibeDatei(pfad, dokument); err != nil {
		return nil, err
	}

	return lauf, nil
}

// zeichne rastert eine Karte in der Auflösung des Bogens.
func (b *Bogenleger) zeichne(
	ctx context.Context, karte *domain.CardRow, masse Bogenmasse,
) (image.Image, []domain.Befund, error) {
	plan, planbefunde, err := b.bauer.Kartenplan(ctx, karte)
	if err != nil {
		return nil, nil, fmt.Errorf("den Plan bauen: %w", err)
	}

	// **Die Zielbreite folgt aus dem Papier**: so viele Bildpunkte, wie die
	// gedruckte Breite **der Karte** bei dieser Auflösung trägt. Der Anschnitt
	// kommt danach außen herum (`mitAnschnitt`); bis I04 stand er hier mit
	// drin, und die Karte wurde über die Schnittkante hinaus gezogen.
	breite := masse.Kartenbreite * float64(masse.DPI) / 25.4

	vergroesserung := 1.0
	if plan.Breite > 0 {
		vergroesserung = breite / plan.Breite
	}

	// **Noch die Vorgabe.** Die Güte kommt als Wahl der Nutzenden über die
	// Flächen herein (§16); bis dahin rastert dieser Weg wie bisher.
	bild, rasterbefunde, err := b.rasterer.Zeichne(ctx, plan, vergroesserung,
		render.GueteMittel)
	if err != nil {
		return nil, nil, fmt.Errorf("die Karte zeichnen: %w", err)
	}

	// **Die Befunde des Plans zuerst**: sie sagen, *warum* eine Karte behilflich
	// aussieht (#206), und der Grund gehört vor die Folge.
	befunde := append(beiKarte(planbefunde, masse.name(karte)),
		umschlagen(rasterbefunde, filepath.Base(masse.name(karte)))...)

	if verzerrt(plan.Breite, plan.Hoehe, masse) {
		befunde = append(befunde, domain.Befund{
			Ort: masse.name(karte),
			Meldung: fmt.Sprintf(
				"die Vorlage misst %.0f × %.0f und wird auf %g × %g mm gezogen; "+
					"das verzerrt sie sichtbar",
				plan.Breite, plan.Hoehe, masse.Kartenbreite, masse.Kartenhoehe),
		})
	}

	// **Auf Weiß, bevor es ins PDF geht**: ein Druckbogen ist Papier, und
	// `internal/pdf` nimmt keine Durchsichtigkeit entgegen.
	//
	// **Dann auf das Kartenmaß, dann erst der Anschnitt**: der Rasterer hält
	// das Verhältnis der Vorlage, und eine Vorlage im anderen Verhältnis ließe
	// den Anschnitt aus Karteninhalt bestehen (J05).
	gelegt := aufKartenmass(aufWeiss(bild), masse)

	if masse.Anschnitt > 0 {
		gelegt = mitAnschnitt(gelegt,
			int(math.Round(bildbreite(masse))),
			int(math.Round((masse.Kartenhoehe+2*masse.Anschnitt)*float64(masse.DPI)/25.4)))
	}

	return gelegt, befunde, nil
}

// name nennt eine Karte für einen Befund.
func (m Bogenmasse) name(karte *domain.CardRow) string {
	if wert, da := karte.Value("name"); da && wert != "" {
		return wert
	}

	return fmt.Sprintf("Karte %d", karte.Number)
}

// verzerrt sagt, ob das Verhältnis der Vorlage spürbar von der Karte abweicht.
//
// **Ohne Prüfung auf null**: ein Plan ohne Maß kommt hier nicht an, weil
// `render.Zeichne` ihn vorher abweist. Ein Zweig, den kein Aufruf betreten
// kann, wäre eine Behauptung darüber, was im Ernstfall geschieht (§3).
func verzerrt(planbreite, planhoehe float64, masse Bogenmasse) bool {
	vorlage := planbreite / planhoehe
	karte := masse.Kartenbreite / masse.Kartenhoehe

	return math.Abs(vorlage-karte)/karte > grenzeVerzerrung
}

// schreibeDatei legt das Dokument an seinen Ort.
func schreibeDatei(pfad string, dokument *pdf.Dokument) error {
	if err := os.MkdirAll(filepath.Dir(pfad), 0o750); err != nil {
		return fmt.Errorf("das Zielverzeichnis anlegen: %w", err)
	}

	// **Erst bauen, dann atomar schreiben.** Der erste Halbsatz stand hier
	// schon und fängt einen Fehler des Erzeugers; er fängt nicht `O_TRUNC`.
	// Ein PDF, das zur Hälfte dasteht, sieht im Dateibrowser aus wie ein
	// ganzes — und genau dahin führte `os.WriteFile`, wenn das Schreiben nach
	// dem Kürzen scheitert. Angestrichen vom zweiten Zweitblick über die
	// Freigabe 0.3.0 (#1170, Befund 5).
	if err := schreibeAtomar(pfad, dokument.Bytes()); err != nil {
		return fmt.Errorf("den Bogen schreiben: %w", err)
	}

	return nil
}

// beiKarte hängt an Befunde ohne Ort den Namen der Karte.
//
// **Sonst stünde auf einem Bogen mit dreihundert Karten dreihundertmal
// derselbe Satz** ohne die Angabe, welche Karte gemeint ist.
func beiKarte(befunde []domain.Befund, name string) []domain.Befund {
	benannt := make([]domain.Befund, 0, len(befunde))

	for _, befund := range befunde {
		if befund.Ort == "" {
			befund.Ort = name
		}

		benannt = append(benannt, befund)
	}

	return benannt
}

// setDaSein weist ein Set ab, das es nicht gibt.
//
// **Und nicht dasselbe wie ein leeres Set**: jenes gibt seit dem 2026-08-30
// eine Validierung (`ErrValidation`), dieses ein Nicht-Gefunden. Ein Skript
// soll „die Kennung zeigt auf nichts" von „das Set ist leer" unterscheiden.
//
// # Was es kostet, und was daran ungemessen ist
//
// `Setkoepfe` liest **alle** Sets und rechnet je Set eine korrelierte
// Unterabfrage über `cards` — für eine Existenzprüfung ist das mehr, als nötig
// wäre (Befund 9 des Zweitblicks über Phase 4.4). **Gemessen am 2026-08-30**
// gegen einen Bestand mit fünf Sets: `mse sets` — derselbe Weg — braucht nach
// dem ersten Lauf **unter 10 ms**, und ein Bogen über 107 Karten rastert
// Sekunden.
//
// **Was diese Zahl nicht sagt**: was ein Bestand mit Hunderten Sets kostet.
// Die Reichweite einer Messung ist die Reichweite ihrer Aussage (§10). Der
// Ausweg wäre eine Naht `Setkopf(ctx, id)`; sie fehlt, und das steht hier,
// damit es niemand für geprüft hält.
func (b *Bogenleger) setDaSein(ctx context.Context, setID int64) error {
	koepfe, err := b.karten.Setkoepfe(ctx)
	if err != nil {
		return fmt.Errorf("die Sets lesen: %w", err)
	}

	if slices.ContainsFunc(koepfe, func(kopf domain.SetHead) bool { return kopf.ID == setID }) {
		return nil
	}

	return fmt.Errorf("%w: ein Set mit der Kennung %d gibt es nicht", domain.ErrNotFound, setID)
}
