package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Kartenliste in ihren drei Gestalten**
// (#891,
// Werkstück 5 von Phase Y).
//
// Gemessen an `Black.mse-set` mit 3.788 Karten: `Karten` gibt 4.761.182 Bytes
// JSON, `Kartenzeilen` 245.641 — 19,4-mal weniger (`BenchmarkKartenliste`,
// 2026-08-25).

// listenprobe belegt die drei Nähte aus dem Arbeitsspeicher.
type listenprobe struct {
	zeilen []domain.CardLine
	karte  *domain.CardRow
	werte  []domain.CardField
	fehler error
}

func (l listenprobe) Kartenzeilen(context.Context, int64) ([]domain.CardLine, error) {
	return l.zeilen, l.fehler
}

func (l listenprobe) KarteLesen(context.Context, int64) (*domain.CardRow, error) {
	return l.karte, l.fehler
}

func (l listenprobe) Kartenfeldwerte(context.Context, int64, string) (
	[]domain.CardField, error,
) {
	return l.werte, l.fehler
}

func mitListe(probe listenprobe) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitKartenzeilen(probe)
	anwendung.MitEinzelkarte(probe)
	anwendung.MitKartenfeldwerten(probe)

	return anwendung
}

func TestKartenzeilenKommenAlsDTO(t *testing.T) {
	anwendung := mitListe(listenprobe{zeilen: []domain.CardLine{
		{ID: 42, Number: 0, Name: "Feuerball", Fields: 3},
		{ID: 43, Number: 1, Name: "", Fields: 0},
	}})

	antwort := anwendung.Kartenzeilen(t.Context(), "7")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	// **Die Kennung ist eine Zeichenkette** — dieselbe Zusage wie bei den Sets:
	// JavaScript kennt ganze Zahlen genau nur bis 2^53.
	if antwort.Zeilen[0].ID != "42" || antwort.Zeilen[0].Name != "Feuerball" {
		t.Errorf("erste Zeile %+v", antwort.Zeilen[0])
	}

	// **Die Zahl der Felder kommt mit.** Sie ist der Unterschied zwischen
	// „heißt nicht" und „ist leer" — die Liste zeigt es an.
	if antwort.Zeilen[0].Felder != 3 || antwort.Zeilen[1].Felder != 0 {
		t.Errorf("Feldzahlen %d, %d", antwort.Zeilen[0].Felder, antwort.Zeilen[1].Felder)
	}
}

// **Eine leere Liste ist `[]` und nicht `null`** (§3). Ein `null` käme im
// Fenster als `undefined` an, und die erste Stelle darüber bräche.
func TestLeereKartenlistenSindListen(t *testing.T) {
	anwendung := mitListe(listenprobe{})

	roh, err := json.Marshal(anwendung.Kartenzeilen(t.Context(), "7"))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if !strings.Contains(string(roh), `"zeilen":[]`) {
		t.Errorf("JSON %s", roh)
	}

	roh, err = json.Marshal(anwendung.Kartenfeldwerte(t.Context(), "7", "name"))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if !strings.Contains(string(roh), `"werte":[]`) {
		t.Errorf("JSON %s", roh)
	}
}

func TestEineEinzelneKarteTraegtIhreWerte(t *testing.T) {
	anwendung := mitListe(listenprobe{karte: &domain.CardRow{
		ID:     42,
		Number: 3,
		Values: []domain.Value{{Field: "name", Text: "Feuerball"}},
	}})

	antwort := anwendung.Karte(t.Context(), "42")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if antwort.Karte.ID != "42" || len(antwort.Karte.Werte) != 1 {
		t.Errorf("Karte %+v", antwort.Karte)
	}
}

func TestFeldwerteKommenAlsDTO(t *testing.T) {
	anwendung := mitListe(listenprobe{werte: []domain.CardField{
		{ID: 42, Text: "Kreatur"},
	}})

	antwort := anwendung.Kartenfeldwerte(t.Context(), "7", "typ")
	if antwort.Fehler != nil {
		t.Fatalf("Befund: %+v", antwort.Fehler)
	}

	if len(antwort.Werte) != 1 ||
		antwort.Werte[0].ID != "42" || antwort.Werte[0].Wert != "Kreatur" {
		t.Errorf("Werte %+v", antwort.Werte)
	}
}

// **Eine unlesbare Kennung ist ein Validierungsfehler** und keine leere Liste.
func TestUnlesbareKennungenInDerKartenliste(t *testing.T) {
	anwendung := mitListe(listenprobe{})

	for name, hole := range map[string]func() *app.Fehler{
		"Kartenzeilen": func() *app.Fehler {
			return anwendung.Kartenzeilen(t.Context(), "keine Zahl").Fehler
		},
		"Karte": func() *app.Fehler {
			return anwendung.Karte(t.Context(), "keine Zahl").Fehler
		},
		"Kartenfeldwerte": func() *app.Fehler {
			return anwendung.Kartenfeldwerte(t.Context(), "keine Zahl", "typ").Fehler
		},
	} {
		t.Run(name, func(t *testing.T) {
			befund := hole()
			if befund == nil || befund.Code != "validierung" {
				t.Errorf("Befund %+v", befund)
			}
		})
	}
}

// **Ein Lesefehler wird übersetzt** und erreicht die Oberfläche nicht roh (§8).
func TestLesefehlerInDerKartenliste(t *testing.T) {
	anwendung := mitListe(listenprobe{fehler: errors.New("Platte weg")})

	for name, hole := range map[string]func() *app.Fehler{
		"Kartenzeilen": func() *app.Fehler {
			return anwendung.Kartenzeilen(t.Context(), "7").Fehler
		},
		"Karte": func() *app.Fehler {
			return anwendung.Karte(t.Context(), "7").Fehler
		},
		"Kartenfeldwerte": func() *app.Fehler {
			return anwendung.Kartenfeldwerte(t.Context(), "7", "typ").Fehler
		},
	} {
		t.Run(name, func(t *testing.T) {
			if befund := hole(); befund == nil {
				t.Error("kein Befund")
			}
		})
	}
}

// **Eine nicht gesetzte Naht meldet sich** — sie stürzt nicht ab.
//
// Das ist die zweite Hälfte von `TestNachDemStartIstJedeNahtGesetzt`: der
// prüft, dass der Composition Root sie setzt, dieser, dass ein Aufruf ohne sie
// eine Antwort gibt statt eines `nil`-Zeigers.
func TestOhneNahtMeldetSichDieKartenliste(t *testing.T) {
	anwendung := app.NeueApp("")

	for name, hole := range map[string]func() *app.Fehler{
		"Kartenzeilen": func() *app.Fehler {
			return anwendung.Kartenzeilen(t.Context(), "7").Fehler
		},
		"Karte": func() *app.Fehler {
			return anwendung.Karte(t.Context(), "7").Fehler
		},
		"Kartenfeldwerte": func() *app.Fehler {
			return anwendung.Kartenfeldwerte(t.Context(), "7", "typ").Fehler
		},
	} {
		t.Run(name, func(t *testing.T) {
			befund := hole()
			if befund == nil || befund.Code != "validierung" {
				t.Errorf("Befund %+v", befund)
			}
		})
	}
}
