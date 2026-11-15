package app_test

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Dieselben Wege ohne Fenster** — die Fläche für die Kommandozeile und
// die Werkstatt für MCP und REST (LOC-02,
// #1688).
func TestDieUebersetzungenOhneFenster(t *testing.T) {
	db := offeneProbedatenbank(t)
	flaeche := app.SprachflaecheUeber(db)
	ctx := probekontext(t)

	if antwort := flaeche.SprachpaketAnlegen(ctx, "da", "Dansk", "en"); antwort.Fehler != nil {
		t.Fatalf("angelegt: %+v", antwort.Fehler)
	}

	lizenz := "CC0-1.0"
	if antwort := flaeche.SprachpaketAendern(ctx, "da", app.SprachpaketangabenDTO{Lizenz: &lizenz},
		map[string]string{"allgemein.arbeitsplatz.keinSet": "Intet sæt"}); antwort.Fehler != nil {
		t.Fatalf("geändert: %+v", antwort.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "da.json")
	if antwort := flaeche.SprachpaketAusgabeSchreiben(ctx, "da", ziel); antwort.Fehler != nil {
		t.Fatalf("geschrieben: %+v", antwort.Fehler)
	}

	text := flaeche.SprachpaketAusgabeText(ctx, "da")
	if liste := flaeche.Sprachpakete(ctx); len(liste.Pakete) != 1 || text.Fehler != nil {
		t.Fatalf("die Liste %+v, der Text %+v", liste, text.Fehler)
	}

	if antwort := flaeche.SprachpaketLoeschen(ctx, "da"); antwort.Fehler != nil {
		t.Fatalf("gelöscht: %+v", antwort.Fehler)
	}

	if antwort := flaeche.SprachpaketEinlesenLesen(ctx, ziel, false); antwort.Fehler != nil {
		t.Fatalf("aus der Datei: %+v", antwort.Fehler)
	}

	if antwort := flaeche.SprachpaketEinlesenText(ctx, text.Inhalt, true); antwort.Fehler != nil {
		t.Fatalf("aus dem Text: %+v", antwort.Fehler)
	}

	if antwort := flaeche.Sprachpaket(ctx, "da"); antwort.Paket.Uebersetzt != 1 {
		t.Errorf("gezeigt: %+v", antwort.Paket)
	}

	werkstatt := app.NeueSprachwerkstatt(db)

	pakete, err := werkstatt.Sprachpakete(ctx)
	if err != nil {
		t.Fatalf("über die Werkstatt: %v", err)
	}

	roh, err := json.Marshal(pakete)
	if err != nil || !strings.Contains(string(roh), `"kennung":"da"`) {
		t.Errorf("über die Werkstatt: %s, %v", roh, err)
	}

	if _, anlegefehler := werkstatt.SprachpaketAnlegen(ctx, "pt-BR", "Português", "en"); anlegefehler != nil {
		t.Fatalf("über die Werkstatt angelegt: %v", anlegefehler)
	}

	// **Was in der Karte fehlt, bleibt** — nur der Stand ändert sich.
	geaendert, err := werkstatt.SprachpaketAendern(ctx, "pt-BR", map[string]string{"stand": "maschinell"},
		map[string]string{"allgemein.arbeitsplatz.keinSet": "Nenhum set"})
	antwort, istAntwort := geaendert.(app.SprachpaketAntwort)
	if err != nil || !istAntwort || antwort.Paket.Stand != "maschinell" || antwort.Paket.Name != "Português" {
		t.Errorf("über die Werkstatt geändert: %+v, %v", geaendert, err)
	}

	if _, ohneLizenz := werkstatt.SprachpaketAusgabe(ctx, "pt-BR"); !errors.Is(ohneLizenz, domain.ErrValidation) {
		t.Errorf("ohne Lizenz ausgegeben: %v", ohneLizenz)
	}

	datei, err := werkstatt.SprachpaketAusgabe(ctx, "da")
	if err != nil {
		t.Fatalf("über die Werkstatt ausgegeben: %v", err)
	}

	text, istDatei := datei.(app.SprachpaketdateiAntwort)
	if !istDatei {
		t.Fatalf("keine Datei: %T", datei)
	}

	if _, err := werkstatt.SprachpaketEinlesen(ctx, text.Inhalt, true); err != nil {
		t.Errorf("über die Werkstatt eingelesen: %v", err)
	}

	if gezeigt, err := werkstatt.Sprachpaket(ctx, "da"); err != nil || gezeigt == nil {
		t.Errorf("über die Werkstatt gezeigt: %v", err)
	}

	if _, err := werkstatt.SprachpaketLoeschen(ctx, "pt-BR"); err != nil {
		t.Errorf("über die Werkstatt gelöscht: %v", err)
	}

	if _, err := werkstatt.SprachpaketLoeschen(ctx, "pt-BR"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("zweimal gelöscht: %v", err)
	}
}
