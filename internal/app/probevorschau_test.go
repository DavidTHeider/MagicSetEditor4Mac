package app_test

import (
	"context"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Vorschau, die einen Feldwert einsetzt, ohne ihn zu schreiben**
// (#288).
//
// # Wofür
//
// AGENTS.md §6 verlangt: „Eine Änderung an einem Feld ist in der Vorschau zu
// sehen, ohne dass jemand etwas bestätigt." Bis hierher schrieb der Editor
// beim **Verlassen** des Feldes — und Verlassen ist eine Bestätigung.
//
// **Warum nicht einfach bei jedem Tastendruck schreiben.** Ein Schreiben legt
// einen Schritt im Verlauf an; wer einen Kartennamen tippt, bekäme zwanzig
// davon, und ⌘Z nähme Buchstaben zurück statt Änderungen. MSE2 zeichnet aus
// dem Wert im Arbeitsspeicher und legt den Schritt gröber an
// (`src/gui/value/text.cpp`, `TextValueEditor::onChar` gegen
// `TextValueAction`, Fassung f9d9356 vom 2021-07-06).
//
// Hier heißt das: die Vorschau bekommt den Wert mit, der Bestand nicht.
func TestDieProbevorschauSetztDenWertEin(t *testing.T) {
	karte := karteprobe{karte: &domain.CardRow{
		ID: 1, Values: []domain.Value{{Field: "name", Text: "Feuerball"}},
	}}

	schreiber := &schreibprobe{}

	antwort := editor(schreiber, karte).
		Probevorschau(probekontext(t), "1", "name", "Eisregen", 1)
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if !strings.HasPrefix(antwort.Vorschau.Bild, app.Bildwegpfad) {
		t.Errorf("%.40s …", antwort.Vorschau.Bild)
	}

	// **Und der Bestand hat nichts davon gesehen.**
	if schreiber.feld != "" || schreiber.wert != "" {
		t.Errorf("geschrieben wurde %q = %q — die Probevorschau schreibt nicht",
			schreiber.feld, schreiber.wert)
	}
}

// **Ein Feld, das die Karte noch nicht führt, kommt hinzu.**
//
// Eine Karte aus einem fremden Set trägt nicht jedes Feld des Spiels; wer dort
// zu tippen anfängt, soll es trotzdem sehen.
func TestDieProbevorschauNimmtEinNeuesFeldAuf(t *testing.T) {
	karte := karteprobe{karte: &domain.CardRow{
		ID: 1, Values: []domain.Value{{Field: "name", Text: "Feuerball"}},
	}}

	antwort := editor(&schreibprobe{}, karte).
		Probevorschau(probekontext(t), "1", "rule text", "Fliegend.", 1)
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if antwort.Vorschau == nil {
		t.Fatal("keine Vorschau")
	}
}

// **Eine unlesbare Kennung wird abgewiesen** — wie bei [app.App.Vorschau].
func TestDieProbevorschauWeistEineUnlesbareKennungAb(t *testing.T) {
	befund := editor(&schreibprobe{}, karteprobe{}).
		Probevorschau(probekontext(t), "abc", "name", "x", 1).Fehler
	if befund == nil || befund.Feld != "id" {
		t.Errorf("%+v", befund)
	}
}

// **Ein Lesefehler kommt als Befund heraus.**
func TestDieProbevorschauMeldetEinenLesefehler(t *testing.T) {
	befund := editor(&schreibprobe{}, karteprobe{fehler: domain.ErrNotFound}).
		Probevorschau(probekontext(t), "1", "name", "x", 1).Fehler
	if befund == nil || befund.Code != app.CodeNichtGefunden {
		t.Errorf("%+v", befund)
	}
}

// **Die Vorschau nennt den Stil, mit dem sie gezeichnet ist** (Zweitblick über
// die Folgen vor 0.8.0, Befund 8): das Formular merkt sich die Lage der Felder
// je Stil, und eine Vorschau, die noch den vorigen zeigt — oder einen, der nur
// getippt ist —, legte ihre Flächen sonst unter dem falschen ab.
func TestDieVorschauNenntIhrenStil(t *testing.T) {
	karte := karteprobe{karte: &domain.CardRow{
		ID: 1, Values: []domain.Value{{Field: "name", Text: "Feuerball"}, {Field: "stylesheet", Text: " m15 "}},
	}}

	geschrieben := editor(&schreibprobe{}, karte).Vorschau(probekontext(t), "1", 1)
	if geschrieben.Fehler != nil || geschrieben.Vorschau.Stil != "m15" {
		t.Errorf("die Vorschau nennt %+v", geschrieben)
	}

	getippt := editor(&schreibprobe{}, karte).Probevorschau(probekontext(t), "1", "stylesheet", "m15-planeswalker", 1)
	if getippt.Fehler != nil || getippt.Vorschau.Stil != "m15-planeswalker" {
		t.Errorf("die Probevorschau nennt %+v", getippt)
	}
}

// **Und das Paket, mit dem sie gezeichnet ist** — #2314: eine Karte, die den
// Stil des Sets erbt, schreibt `”`, und welches Paket das ist, ändert ein
// fremder Setstilwechsel. Das Formular ordnet je Paket.
func TestDieVorschauNenntDasGezeichnetePaket(t *testing.T) {
	quelle := &paketstil{paket: "magic-alt.mse-style"}
	anwendung := app.NeueApp("")
	anwendung.MitEditor(&schreibprobe{}, karteprobe{karte: &domain.CardRow{ID: 1}},
		service.NeuerPlanbauer().MitStilen(quelle), render.NeuerRasterer(ohneBilder{}))

	vorher := anwendung.Vorschau(probekontext(t), "1", 1)
	quelle.paket = "magic-neu.mse-style"
	nachher := anwendung.Vorschau(probekontext(t), "1", 1)

	if vorher.Fehler != nil || nachher.Fehler != nil ||
		vorher.Vorschau.Gezeichnet != "magic-alt.mse-style" || nachher.Vorschau.Gezeichnet != "magic-neu.mse-style" ||
		nachher.Vorschau.Stil != "" {
		t.Errorf("vorher %+v, nachher %+v", vorher.Vorschau, nachher.Vorschau)
	}
}

// paketstil ist ein Setstil, den ein anderer Prozess wechseln kann.
type paketstil struct{ paket string }

func (s *paketstil) StilFuerSet(context.Context, int64) (service.Vorlage, []domain.Befund, bool) {
	return service.Vorlage{Stil: msepaket.Stil{Breite: 375, Hoehe: 523, Paket: s.paket}}, nil, true
}

func (s *paketstil) StilFuerKarte(ctx context.Context, karte *domain.CardRow) (service.Vorlage, []domain.Befund, bool) {
	return s.StilFuerSet(ctx, karte.SetID)
}
