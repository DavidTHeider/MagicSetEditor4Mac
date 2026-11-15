package app_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Vorschau, die eine Stiloption einsetzt, ohne sie zu schreiben** —
// #590.
//
// # Wofür
//
// AGENTS.md §6: „Eine Änderung an einem Feld ist in der Vorschau zu sehen,
// ohne dass jemand etwas bestätigt." Für ein **Kartenfeld** löst das
// [app.App.Probevorschau] seit #288 ein. Für eine **Stiloption** galt es bis
// hierher nur zur Hälfte: eine Auswahl und ein Kästchen schreiben beim
// Umstellen, ein **Textfeld** erst beim Verlassen — und Verlassen ist eine
// Bestätigung.
//
// Über die 376 Stilpakete des Packs sind **738 Stiloptionen** vom Typ `text`,
// die dritthäufigste Art (`docs/STILOPTIONEN.md`).
//
// # Warum über die Karte und nicht über das Set
//
// Weil MSE die Reihenfolge schon hat: **die Karte schlägt das Set, das Set
// schlägt die Vorgabe des Stils** (`Set::stylingDataFor`, `src/data/set.cpp`,
// Zeilen 99–101, Fassung f9d9356 vom 2021-07-06; hier in
// `service/stilumgebung.go`). Eine Option auf der Probekarte zu überlagern
// ergibt dasselbe Bild, als hätte jemand sie setweit gesetzt — ohne dass ein
// zweiter Weg in den Bestand entsteht.
func TestDieStilprobevorschauSetztDieOptionEin(t *testing.T) {
	t.Parallel()

	bauer := &stilplanprobe{}
	schreiber := &schreibprobe{}

	antwort := stileditor(schreiber, karteprobe{karte: &domain.CardRow{
		ID: 1, Values: []domain.Value{{Field: "name", Text: "Feuerball"}},
	}}, bauer).
		Stilprobevorschau(probekontext(t), "1", "border color", "rgb(255,0,0)", 1)
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	schalter, _ := bauer.karte.Value("has styling")
	if schalter != "true" {
		t.Errorf("has styling = %q — ohne den Schalter liest der Planbau die "+
			"Kartenwerte gar nicht", schalter)
	}

	werte, _ := bauer.karte.Value("styling data")
	if werte != "border color: rgb(255,0,0)" {
		t.Errorf("styling data = %q", werte)
	}

	// **Und der Bestand hat nichts davon gesehen.**
	if schreiber.feld != "" || schreiber.wert != "" {
		t.Errorf("geschrieben wurde %q = %q — die Probevorschau schreibt nicht",
			schreiber.feld, schreiber.wert)
	}
}

// **Was die Karte selbst an Stilwerten führt, bleibt stehen.**
//
// Eine Karte darf eigene Stiloptionen tragen (`has styling: true`). Wer die
// Zeile ersetzte statt sie zu ergänzen, zeigte eine Vorschau, in der die
// übrigen Einstellungen dieser Karte verschwunden sind — und der Unterschied
// im Bild käme aus der falschen Stelle.
func TestDieStilprobevorschauLaesstDieUebrigenWerteStehen(t *testing.T) {
	t.Parallel()

	bauer := &stilplanprobe{}

	antwort := stileditor(&schreibprobe{}, karteprobe{karte: &domain.CardRow{
		ID: 1,
		Values: []domain.Value{
			{Field: "has styling", Text: "true"},
			{Field: "styling data", Text: "text box mana symbols: klein\nborder color: blau"},
		},
	}}, bauer).
		Stilprobevorschau(probekontext(t), "1", "border color", "rot", 1)
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	werte, _ := bauer.karte.Value("styling data")
	if werte != "text box mana symbols: klein\nborder color: rot" {
		t.Errorf("styling data = %q — erwartet dieselbe Reihenfolge mit "+
			"ersetztem Wert", werte)
	}
}

// **Eine Zeile ohne Doppelpunkt bleibt liegen.**
//
// `kartenstilwerte` übergeht sie beim Lesen (`service/kartenstiloptionen.go`);
// wer sie hier fallen ließe, verlöre beim nächsten Schreiben etwas, das die
// Karte trägt — und die Vorschau zeigte einen Stand, den der Bestand nicht hat.
func TestDieStilprobevorschauLaesstEineZeileOhneDoppelpunktLiegen(t *testing.T) {
	t.Parallel()

	bauer := &stilplanprobe{}

	antwort := stileditor(&schreibprobe{}, karteprobe{karte: &domain.CardRow{
		ID: 1,
		Values: []domain.Value{
			{Field: "styling data", Text: "kaputte Zeile ohne alles"},
		},
	}}, bauer).
		Stilprobevorschau(probekontext(t), "1", "border color", "rot", 1)
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	werte, _ := bauer.karte.Value("styling data")
	if werte != "kaputte Zeile ohne alles\nborder color: rot" {
		t.Errorf("styling data = %q", werte)
	}
}

// **Der Zeiger auf die gelesene Karte wird nicht beschrieben.**
//
// Dieselbe Regel wie bei [app.App.Probevorschau]: der Kartenleser darf einen
// Zeiger auf etwas Gemerktes geben, und wer darin schriebe, änderte den
// Bestand über einen Weg, den niemand als Schreiben erkennt.
func TestDieStilprobevorschauLaesstDieGelesenenKarteUnberuehrt(t *testing.T) {
	t.Parallel()

	gelesen := &domain.CardRow{
		ID:     1,
		Values: []domain.Value{{Field: "styling data", Text: "border color: blau"}},
	}

	if antwort := stileditor(&schreibprobe{}, karteprobe{karte: gelesen}, &stilplanprobe{}).
		Stilprobevorschau(probekontext(t), "1", "border color", "rot", 1); antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if werte, _ := gelesen.Value("styling data"); werte != "border color: blau" {
		t.Errorf("die gelesene Karte trägt jetzt %q", werte)
	}
}

// **Eine unlesbare Kennung wird abgewiesen** — wie bei [app.App.Vorschau].
func TestDieStilprobevorschauWeistEineUnlesbareKennungAb(t *testing.T) {
	t.Parallel()

	befund := stileditor(&schreibprobe{}, karteprobe{}, &stilplanprobe{}).
		Stilprobevorschau(probekontext(t), "abc", "x", "y", 1).Fehler
	if befund == nil || befund.Feld != "id" {
		t.Errorf("%+v", befund)
	}
}

// **Ein Lesefehler kommt als Befund heraus.**
func TestDieStilprobevorschauMeldetEinenLesefehler(t *testing.T) {
	t.Parallel()

	befund := stileditor(&schreibprobe{}, karteprobe{fehler: domain.ErrNotFound},
		&stilplanprobe{}).
		Stilprobevorschau(probekontext(t), "1", "x", "y", 1).Fehler
	if befund == nil || befund.Code != app.CodeNichtGefunden {
		t.Errorf("%+v", befund)
	}
}

// **Ein leerer Name ist kein Aufruf.**
//
// Ohne diese Prüfung entstünde die Zeile `: wert` in den Stildaten, und
// `kartenstilwerte` überginge sie stillschweigend — die Vorschau zeigte dann
// dasselbe Bild wie ohne Änderung, und niemand sähe, dass der Aufruf ins Leere
// ging (§8).
func TestDieStilprobevorschauBrauchtEinenNamen(t *testing.T) {
	t.Parallel()

	befund := stileditor(&schreibprobe{}, karteprobe{karte: &domain.CardRow{ID: 1}},
		&stilplanprobe{}).
		Stilprobevorschau(probekontext(t), "1", "  ", "rot", 1).Fehler
	if befund == nil || befund.Feld != "name" {
		t.Errorf("%+v", befund)
	}
}

// stileditor baut eine Anwendung mit einem Planbauer, der seine Karte merkt.
func stileditor(
	schreiber *schreibprobe, karte karteprobe, bauer app.Planbauer,
) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitEditor(schreiber, karte, bauer, render.NeuerRasterer(ohneBilder{}))

	return anwendung
}

// stilplanprobe merkt sich die Karte, mit der der Plan gebaut wurde.
type stilplanprobe struct {
	ohneWortlisten
	ohneWahlbilder

	karte *domain.CardRow
}

func (p *stilplanprobe) Kartenplan(ctx context.Context, karte *domain.CardRow) (
	*render.Plan, []domain.Befund, error,
) {
	p.karte = karte

	return service.NeuerPlanbauer().Kartenplan(ctx, karte)
}

func (p *stilplanprobe) Kartenpalette(
	ctx context.Context, karte *domain.CardRow, feld string,
) ([]service.Paletteneintrag, bool, []domain.Befund, error) {
	return service.NeuerPlanbauer().Kartenpalette(ctx, karte, feld)
}
