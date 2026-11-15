package app

// **Die Stände der bezogenen Bäume ohne Fenster** — G04
// (#2251):
// dieselbe Auskunft und dieselben Schritte für Kommandozeile, MCP und REST
// (§16), in derselben Gestalt wie im Fenster.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// MitAktualisierer setzt die Naht zu den Bezugsständen — für Tests.
func (w *Paketweg) MitAktualisierer(aktualisierer Paketaktualisierer) *Paketweg {
	w.aktualisierer = aktualisierer

	return w
}

// Bezuege nennt die bezogenen Bäume.
func (w *Paketweg) Bezuege(ctx context.Context, fragen bool) ([]PaketbezugDTO, error) {
	staende, err := w.aktualisierer.Staende(ctx, fragen)
	if err != nil {
		return nil, err
	}

	return BezuegeAlsDTO(ctx, staende), nil
}

// BezugVorschau sagt, was eine Aktualisierung täte.
func (w *Paketweg) BezugVorschau(ctx context.Context, name, stand string) (BezugsvorschauAntwort, error) {
	vorschau, err := w.aktualisierer.Vorschau(ctx, name, stand)
	if err != nil {
		return BezugsvorschauAntwort{}, err
	}

	return BezugsvorschauAlsDTO(vorschau), nil
}

// BezugAktualisieren aktualisiert einen Baum und vermerkt, was ankam.
func (w *Paketweg) BezugAktualisieren(ctx context.Context, name, stand string, bestaetigt bool) (*service.Bezugsergebnis, error) {
	ergebnis, err := w.aktualisierer.Aktualisieren(ctx, name, stand, bestaetigt)
	if err != nil {
		return nil, err
	}

	return ergebnis, w.vermerkt(ctx, ergebnis.Adresse, ergebnis.Ort, ergebnis.Stand, ergebnis.Pruefsumme)
}

// BezugZurueck tauscht einen Baum mit seinem aufgehobenen Stand.
func (w *Paketweg) BezugZurueck(ctx context.Context, name string) (*service.Rueckweg, error) {
	rueckweg, err := w.aktualisierer.Zurueck(ctx, name)
	if err != nil {
		return nil, err
	}

	return rueckweg, w.vermerkt(ctx, rueckweg.Adresse, rueckweg.Ort, rueckweg.Stand, rueckweg.Pruefsumme)
}

// vermerkt trägt die Herkunft ein. **Ein Fehler dabei nimmt den Tausch nicht
// zurück** (§7, Regel 3); er wird gemeldet.
func (w *Paketweg) vermerkt(ctx context.Context, adresse, ort, stand, pruefsumme string) error {
	_, err := w.verwalter.Vermerken(ctx, service.Vermerk{
		Ort: ort, Herkunft: herkunftstext(adresse, stand), Pruefsumme: pruefsumme,
	})

	return err
}
