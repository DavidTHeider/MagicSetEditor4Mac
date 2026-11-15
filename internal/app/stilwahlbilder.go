package app

import (
	"bytes"
	"context"
	"encoding/base64"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Wahlbilder einer Stiloption** —
// #1225.
//
// # Was gefehlt hat
//
// MSE zeigt zu jeder Wahl einer Stiloption ein Bild; dieses Fenster zeigte
// eine Auswahlliste über die **Namen**. Wer zwischen `modern`, `old` und
// `diagonal T` wählen soll, sah drei Wörter — MSE zeigt drei Tapsymbole.
//
// Über `full-magic-pack/data` betrifft das **303 von 376** Stilpaketen mit
// 1.097 `choice images:`-Blöcken (gezählt am 2026-08-28); es ist dieselbe
// Klasse wie A4-01 im Audit: **für ein visuelles Werkzeug ist eine Liste aus
// Textzeilen die falsche Darstellung.**
//
// # Warum je Option und nicht je Stil
//
// Weil sonst jedes Öffnen der Stilseite alles rasterte, was das Paket erklärt.
// Die Oberfläche fragt nach der Option, die sie zeigt, und sie fragt nur, wo
// [StiloptionDTO.HatWahlbilder] es sagt.

// Wahlbildkante ist die Kantenlänge eines gerasterten Wahlbildes in Punkten.
//
// **adapted.** MSE rastert auf 18 × 18 (`thumbnail_size`,
// `src/gui/value/choice.cpp`, Zeile 17, Fassung f9d9356 vom 2021-07-06) — die
// Größe einer Zeile in einer Windows-Auswahlliste von 2005. Hier steht das
// Doppelte einer 24-Punkt-Kachel: das Fenster zeichnet auf einem Bildschirm
// mit doppelter Auflösung, und ein auf 18 gerastertes Bild sähe dort weich
// aus.
//
// **Quadratisch, wie bei MSE**: `generate` bekommt dieselbe Zahl zweimal und
// `ASPECT_BORDER` dazu — das Seitenverhältnis bleibt, der Rest ist Rand
// (`choice.cpp`, Zeile 53).
const Wahlbildkante = 48

// WahlbildDTO ist das Bild einer Wahl einer Stiloption.
type WahlbildDTO struct {
	// Wahl ist der Name der Wahl, wie die Option ihn erklärt.
	//
	// **Die Beschriftung und nicht der gespeicherte Wert** — MSE schlägt das
	// Bild unter `choiceName(i)` nach (`choice.cpp`, Zeile 175). Bei einer
	// Auswahl und einem Schalter sind beide dasselbe; bei einer Farbe heißt
	// die Wahl `schwarz` und der Wert `rgb(0,0,0)`.
	Wahl string `json:"wahl"`

	// Bild ist ein PNG als Data-URI — unmittelbar in ein `src` einsetzbar,
	// dieselbe Gestalt wie [VorschauDTO.Bild].
	Bild string `json:"bild"`
}

// WahlbilderAntwort trägt die Bilder einer Stiloption.
type WahlbilderAntwort struct {
	Bilder []WahlbildDTO `json:"bilder"`

	// Befunde sind die Wahlen, deren Bild nicht entstanden ist.
	//
	// **Neben den Bildern und nicht statt ihrer** (§8): zwei von drei
	// Tapsymbolen sind besser als keines, und die dritte Wahl steht dann mit
	// ihrem Namen da.
	Befunde []string `json:"befunde"`

	Fehler *Fehler `json:"fehler"`
}

// Stilwahlbilder gibt zu einer Stiloption je Wahl ihr gerastertes Bild.
func (a *App) Stilwahlbilder(ctx context.Context, setID, option string) WahlbilderAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return WahlbilderAntwort{
			Bilder: []WahlbildDTO{}, Befunde: []string{}, Fehler: befund,
		}
	}

	baeume, befunde, err := a.planbauer.Optionsbilder(ctx, kennung, option)
	if err != nil {
		return WahlbilderAntwort{
			Bilder: []WahlbildDTO{}, Befunde: []string{}, Fehler: Uebersetze(ctx, err),
		}
	}

	meldungen := make([]string, 0, len(befunde))
	for _, gefunden := range befunde {
		meldungen = append(meldungen, gefunden.String())
	}

	bilder := make([]WahlbildDTO, 0, len(baeume))

	for _, baum := range baeume {
		gepackt, meldung := a.wahlbild(baum)
		if meldung != "" {
			meldungen = append(meldungen, baum.Wahl+": "+meldung)

			continue
		}

		bilder = append(bilder, WahlbildDTO{Wahl: baum.Wahl, Bild: gepackt})
	}

	return WahlbilderAntwort{Bilder: bilder, Befunde: meldungen}
}

// wahlbild rastert einen Baum und packt ihn als Datenadresse.
//
// **Der Befund des Rasterers und der des Packens sind derselbe Rückgabewert**,
// weil der Aufrufer beide gleich behandelt: die Wahl bleibt bei ihrem Namen,
// und daneben steht, warum.
func (a *App) wahlbild(bild service.Optionsbild) (adresse, meldung string) {
	gerastert, befund := a.rasterer.BildAusBaum(bild.Baum, Wahlbildkante, Wahlbildkante)
	if gerastert == nil {
		// **Ein leerer Befund wäre eine Meldung ohne Inhalt.** Der Rasterer
		// gibt ihn zusammen mit dem Bild, nicht statt seiner
		// (`render/bildausbaum.go`); ohne Bild steht hier immer einer.
		return "", befund
	}

	var gepackt bytes.Buffer
	if err := a.packen(&gepackt, gerastert); err != nil {
		return "", err.Error()
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(gepackt.Bytes()), ""
}

// Wahlbildbauer gibt zu einer Stiloption je Wahl ihren Bildbaum.
//
// **Eine eigene Schnittstelle und trotzdem am [Planbauer]**: es ist derselbe
// Gegenstand — was aus einem Paket und einem Skript etwas Zeichenbares macht.
// Eine zweite Naht daneben ließe zu, dass eine Attrappe die eine belegt und
// die andere nicht (§10).
type Wahlbildbauer interface {
	Optionsbilder(ctx context.Context, setID int64, option string) (
		[]service.Optionsbild, []domain.Befund, error,
	)
}
