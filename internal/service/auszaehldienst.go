package service

// **Der Dienst, der eine erklärte Dimension über ein Set zählt**
// (#1338,
// Werkstück C von Phase 4.11).
//
// # Warum das etwas anderes ist als `Statistik`
//
// `app.Statistik` zählt über **Feldnamen**: wie viele Karten, wie viele
// Felder, wie viele davon belegt. Das ist eine Auskunft über den Bestand.
// Eine Dimension zählt über einen **gerechneten Wert** — `color category`
// steht in keinem Feld, sondern fällt aus einem Skript, das die Farben der
// Vorderseite liest. Beides nebeneinander stehen zu lassen ist Absicht: die
// erste braucht kein Spielpaket, die zweite ohne eines gar nichts.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Auszaehlauskunft gibt, was eine Auszählung braucht.
//
// **Zwei Rufe und nicht einer**: das Angebot ist eine Frage ans Spielpaket,
// die Karten sind eine an den Bestand. Wer nur das Menü füllt, soll nicht
// dreitausend Karten laden.
type Auszaehlauskunft interface {
	Auszaehldimensionen(ctx context.Context, setID int64) (
		angaben []Auszaehlangabe, vorlage Vorlage, fehler error)

	// VorlageFuer gibt die Vorlage **einer** Karte — sie darf ihren eigenen
	// Stil tragen, und davon hängen Werte ab (siehe [Auszaehlwerte]).
	VorlageFuer(ctx context.Context, karte *domain.CardRow) Vorlage

	Karten(ctx context.Context, setID int64) ([]*domain.CardRow, error)
}

// Auszaehlung ist das Ergebnis: eine Dimension und ihre Gruppen.
type Auszaehlung struct {
	Dimension    string
	Beschreibung string

	// Numerisch sagt der Oberfläche, dass die Gruppen eine Achse bilden und
	// keine Aufzählung.
	Numerisch bool

	Gruppen []domain.Auszaehlgruppe
	Befunde []domain.Befund
}

// Auszaehldienst zählt ein Set nach einer erklärten Dimension aus.
type Auszaehldienst struct{ auskunft Auszaehlauskunft }

// NeuerAuszaehldienst baut ihn.
func NeuerAuszaehldienst(auskunft Auszaehlauskunft) *Auszaehldienst {
	return &Auszaehldienst{auskunft: auskunft}
}

// Angebot nennt die Dimensionen, die das Spiel dieses Sets erklärt.
func (d *Auszaehldienst) Angebot(ctx context.Context, setID int64) (
	[]Auszaehlangabe, error,
) {
	angaben, _, fehler := d.auskunft.Auszaehldimensionen(ctx, setID)
	if fehler != nil {
		return nil, fehler
	}

	if angaben == nil {
		angaben = []Auszaehlangabe{}
	}

	return angaben, nil
}

// Auszaehlen zählt das Set nach der benannten Dimension aus.
func (d *Auszaehldienst) Auszaehlen(ctx context.Context, setID int64, name string) (
	*Auszaehlung, error,
) {
	angaben, vorlage, fehler := d.auskunft.Auszaehldimensionen(ctx, setID)
	if fehler != nil {
		return nil, fehler
	}

	gewaehlt, gefunden := dimensionNachNamen(angaben, name)
	if !gefunden {
		return nil, fmt.Errorf("%w: dieses Spiel erklärt keine Dimension %q",
			domain.ErrNotFound, name)
	}

	karten, kartenfehler := d.auskunft.Karten(ctx, setID)
	if kartenfehler != nil {
		return nil, kartenfehler
	}

	// **Die Werte nennen ihre Karten** (UXR-F6, #2013): der Balken und die
	// Liste dahinter kommen aus demselben Lauf.
	werte, befunde := AuszaehlwerteKarten(ctx, karten, vorlage,
		func(karte *domain.CardRow) Vorlage {
			return d.auskunft.VorlageFuer(ctx, karte)
		}, gewaehlt)

	return &Auszaehlung{
		Dimension:    gewaehlt.Name,
		Beschreibung: gewaehlt.Beschreibung,
		Numerisch:    gewaehlt.Numerisch,
		Gruppen:      domain.AuszaehlenKarten(werte, gewaehlt.Auszaehlart()),
		Befunde:      befunde,
	}, nil
}

// dimensionNachNamen sucht eine Dimension.
func dimensionNachNamen(angaben []Auszaehlangabe, name string) (Auszaehlangabe, bool) {
	for _, angabe := range angaben {
		if angabe.Name == name {
			return angabe, true
		}
	}

	return Auszaehlangabe{}, false
}

// Vorlagenauszaehlung gibt die Dimensionen über den Stilspeicher und die
// Karten über den Bestand.
//
// **Zwei Quellen, weil es zwei Fragen sind** — dieselbe Trennung wie in
// [Auszaehlauskunft]: das Spielpaket erklärt, der Bestand zählt.
type Vorlagenauszaehlung struct {
	stile  *Stilspeicher
	karten Auszaehlkarten
}

// Auszaehlkarten liest die Karten eines Sets.
//
// **Eine eigene Schnittstelle und nicht [Kartenquelle]**: jene führt
// `KarteLesen` und `Setkoepfe` mit, und die braucht hier niemand. Die
// Schnittstelle steht dort, wo sie gebraucht wird (§2).
type Auszaehlkarten interface {
	Karten(ctx context.Context, setID int64) ([]*domain.CardRow, error)
}

// NeueVorlagenauszaehlung baut sie.
func NeueVorlagenauszaehlung(
	stile *Stilspeicher, karten Auszaehlkarten,
) *Vorlagenauszaehlung {
	return &Vorlagenauszaehlung{stile: stile, karten: karten}
}

// Auszaehldimensionen gibt die erklärten Dimensionen des Spielpakets.
//
// **Ohne Vorlage keine Dimensionen, aber auch kein Fehler** — dieselbe Regel
// wie bei den Kartenskripten: ein Set ohne Stil ist eines, dessen Spielpaket
// fehlt, und eine leere Liste ist die richtige Antwort darauf (§8).
func (v *Vorlagenauszaehlung) Auszaehldimensionen(ctx context.Context, setID int64) (
	[]Auszaehlangabe, Vorlage, error,
) {
	// **Die ganze Vorlage und nicht ihre Feldlisten.** Warum, steht bei
	// [Auszaehlwerte]: ohne den Stil scheitert jedes `default:`, und ohne die
	// Vorgaben trägt eine Karte nur, was jemand hineingeschrieben hat.
	vorlage, _, da := v.stile.StilFuerSet(ctx, setID)
	if !da {
		return nil, Vorlage{}, nil
	}

	return vorlage.Auszaehldimensionen, vorlage, nil
}

// VorlageFuer gibt die Vorlage einer Karte — ihren eigenen Stil, wenn sie
// einen trägt.
//
// **Ohne eigenen Stil die des Sets**, und das entscheidet der Stilspeicher
// selbst ([Stilspeicher.StilFuerKarte]); ein zweiter Weg dorthin wäre eine
// zweite Regel.
func (v *Vorlagenauszaehlung) VorlageFuer(
	ctx context.Context, karte *domain.CardRow,
) Vorlage {
	vorlage, _, _ := v.stile.StilFuerKarte(ctx, karte)

	return vorlage
}

// Karten reicht die Karten des Sets durch.
func (v *Vorlagenauszaehlung) Karten(ctx context.Context, setID int64) (
	[]*domain.CardRow, error,
) {
	return v.karten.Karten(ctx, setID)
}
