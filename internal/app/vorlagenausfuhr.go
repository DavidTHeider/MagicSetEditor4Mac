package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Vorlagenquelle nennt die Exportvorlagen eines Sets und führt sie aus.
type Vorlagenquelle interface {
	Vorlagen(ctx context.Context, setID int64) ([]msepaket.Exportvorlage, []domain.Befund, error)

	Fuehre(
		ctx context.Context, mittel service.Ausfuhrmittel,
		vorlage msepaket.Exportvorlage,
		karten []*domain.CardRow, setwerte []domain.Value, wahl map[string]string,
	) (string, []domain.Befund, error)

	// MitSymbolen gibt der Vorlage `symbols_to_html` (#658).
	//
	// **Sie steht als Einzige noch hier**, weil sie an den Paketwurzeln hängt
	// und nicht an einem Ziel: die Anwendung richtet sie einmal ein. Was je
	// Ausfuhr wechselt, reist seit
	// #736
	// als [service.Ausfuhrmittel] mit — ein Wert je Lauf statt vier Feldern an
	// einem geteilten Ding.
	MitSymbolen(schriften service.Symbolschriftquelle,
		rasterer service.Symbolrasterer) *service.Vorlagenausfuehrer
}

// ExportvorlageDTO ist eine Vorlage, wie die Oberfläche sie zeigt.
type ExportvorlageDTO struct {
	Paket string `json:"paket"`
	Name  string `json:"name"`

	// Dateityp ist der Filter des Dateidialogs, wie MSE ihn schreibt:
	// `*.xml|*.xml`. Die Oberfläche zeigt ihn nicht; sie reicht ihn zurück.
	Dateityp string `json:"dateityp"`

	Optionen []ExportoptionDTO `json:"optionen"`
}

// ExportoptionDTO ist eine Wahlmöglichkeit einer Vorlage.
type ExportoptionDTO struct {
	Art          string   `json:"art"`
	Name         string   `json:"name"`
	Beschreibung string   `json:"beschreibung"`
	Vorgabe      string   `json:"vorgabe"`
	Auswahl      []string `json:"auswahl"`

	// Wert ist der **behaltene** Wert — sonst die Vorgabe
	// (#492).
	//
	// **Er steht an der Vorlage und kommt nicht aus einem zweiten Aufruf**:
	// eine Vorlage und ihre Werte gehören zusammen, und zwei Antworten liefen
	// beim nächsten Wechsel auseinander.
	Wert string `json:"wert"`
}

// Optionsablage hält die Optionswerte je Exportvorlage — die Naht zu `storage`.
//
// **Je Vorlage, nicht je Set**, wie MSE sie hält (`exportOptionsFor`,
// `src/data/settings.hpp`, Zeilen 180–186, Fassung f9d9356 vom 2021-07-06).
type Optionsablage interface {
	Fuer(ctx context.Context, vorlage string) (map[string]string, error)
	Setzen(ctx context.Context, vorlage, feld, wert string) error
}

// VorlagenAntwort ist die Liste der Vorlagen eines Sets.
type VorlagenAntwort struct {
	// Vorlagen ist **nie nil** — eine leere Liste ist eine Auskunft, `null`
	// wäre ein Fehler auf der anderen Seite (§3, „zwischen den Prüfebenen").
	Vorlagen []ExportvorlageDTO `json:"vorlagen"`
	Befunde  []string           `json:"befunde"`
	Fehler   *Fehler            `json:"fehler,omitempty"`
}

// VorlagenausfuhrAntwort ist das Ergebnis eines Vorlagenlaufs.
type VorlagenausfuhrAntwort struct {
	Pfad        string   `json:"pfad"`
	Befunde     []string `json:"befunde"`
	Abgebrochen bool     `json:"abgebrochen"`
	Fehler      *Fehler  `json:"fehler,omitempty"`
}

// MitVorlagen setzt die Naht zu den Exportvorlagen.
func (a *App) MitVorlagen(quelle Vorlagenquelle) { a.vorlagen = quelle }

// ErstesMuster liest das erste Dateimuster aus einem `file type:`.
//
// **Exportiert für den Test**, weil es die eine Stelle ist, an der aus MSEs
// Schreibweise ein Dialogfilter wird — und die Formen dafür stehen in den
// Paketen, nicht in unserem Kopf.
func ErstesMuster(dateityp string) string { return erstesMuster(dateityp) }

// EndungVon macht aus einem Muster die Endung eines Vorschlagsnamens.
func EndungVon(muster string) string { return endungVon(muster) }

// Exportvorlagen nennt die Vorlagen, die zum Spiel eines Sets gehören.
func (a *App) Exportvorlagen(ctx context.Context, setID string) VorlagenAntwort {
	ctx = a.imProfil(ctx)

	leer := VorlagenAntwort{Vorlagen: []ExportvorlageDTO{}, Befunde: []string{}}

	kennung, befund := kennungAus(setID)
	if befund != nil {
		leer.Fehler = befund

		return leer
	}

	vorlagen, befunde, err := a.vorlagen.Vorlagen(ctx, kennung)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	leer.Vorlagen = zuVorlagenDTO(vorlagen)
	leer.Befunde = meldungen(befunde)

	// **Die behaltenen Werte kommen dazu** (#492). **Ein Lesefehler nimmt die
	// Liste nicht weg** (§8): die Vorlagen sind der Grund, warum jemand hier
	// ist, die Werte eine Bequemlichkeit. Er wird als Befund gemeldet — ein
	// stiller Ausfall sähe aus wie „nichts eingestellt".
	leer.Befunde = append(leer.Befunde, a.behalteneOptionen(ctx, leer.Vorlagen)...)

	return leer
}

// behalteneOptionen trägt die gespeicherten Werte in die Vorlagen ein.
//
// **Ohne Ablage gilt die Vorgabe.** Die Naht ist wie jede andere setzbar; ein
// Fenster ohne sie zeigt, was die Vorlage mitbringt.
func (a *App) behalteneOptionen(ctx context.Context, vorlagen []ExportvorlageDTO) []string {
	var befunde []string

	for i := range vorlagen {
		behalten := map[string]string{}

		if a.exportoptionen != nil {
			gelesen, err := a.exportoptionen.Fuer(ctx, vorlagen[i].Paket)
			if err != nil {
				befunde = append(befunde,
					"die behaltenen Optionen von "+vorlagen[i].Paket+
						" ließen sich nicht lesen: "+err.Error())
			}

			behalten = gelesen
		}

		for j := range vorlagen[i].Optionen {
			option := &vorlagen[i].Optionen[j]

			option.Wert = option.Vorgabe
			if wert, da := behalten[option.Name]; da {
				option.Wert = wert
			}
		}
	}

	return befunde
}

// ExportoptionSetzen behält den Wert einer Option — je Vorlage.
func (a *App) ExportoptionSetzen(ctx context.Context, vorlage, feld, wert string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return SchreibAntwort{Fehler: Uebersetze(ctx, a.exportoptionen.Setzen(ctx, vorlage, feld, wert))}
}

// MitExportoptionen setzt die Naht zur Ablage der Optionswerte.
func (a *App) MitExportoptionen(ablage Optionsablage) { a.exportoptionen = ablage }

// MitSymbolzeichnungen setzt die Quelle, aus der `.mse-symbol` gelesen wird.
//
// **Sie kommt aus derselben Bildquelle wie die Karte** (#660): jene liest
// `.mse-symbol` aus jedem Paket, auch aus dem Set. Der Setter steht hier, weil
// eine Fläche ohne Fenster ihn braucht
// (#1291) —
// bis dahin wurde das Feld nur im Start gesetzt.
func (a *App) MitSymbolzeichnungen(quelle service.Symbolquelle) {
	a.symbolzeichnungen = quelle
}
