package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Vorlage ist der deutsche Katalog** — eingebettet und lesbar.
func TestDieSprachvorlageIstDerDeutscheKatalog(t *testing.T) {
	t.Parallel()

	vorlage := app.Sprachvorlage()
	if len(vorlage) < 4000 || vorlage["allgemein.anwendungsname"] != "MagicSetEditor4Mac" {
		t.Errorf("%d Sätze, Anwendungsname %q", len(vorlage), vorlage["allgemein.anwendungsname"])
	}
}

// sprachanwendung baut eine Anwendung mit eigenen Übersetzungen.
func sprachanwendung(t *testing.T) *app.App {
	t.Helper()

	anwendung := app.NeueApp("")
	anwendung.MitSprachpaketen(app.NeueSprachpaketeUeber(offeneProbedatenbank(t)))

	return anwendung
}

// **Eine eigene Übersetzung am Fenster** — anlegen, Satz für Satz ändern,
// ausgeben und wieder einlesen (LOC-02,
// #1688).
func TestEineEigeneUebersetzungAmFenster(t *testing.T) {
	anwendung := sprachanwendung(t)
	ctx := probekontext(t)

	angelegt := anwendung.SprachpaketAnlegen(ctx, "pt-BR", "Português do Brasil", "en")
	gesamt := len(app.Sprachvorlage())

	if angelegt.Fehler != nil || angelegt.Paket.Gesamt != gesamt || angelegt.Paket.Fehlend != gesamt ||
		len(angelegt.FehlendePfade) != gesamt {
		t.Fatalf("angelegt: %+v", angelegt.Paket)
	}

	lizenz := "CC0-1.0"
	geaendert := anwendung.SprachpaketAendern(ctx, "pt-BR", app.SprachpaketangabenDTO{Lizenz: &lizenz},
		map[string]string{"allgemein.arbeitsplatz.keinSet": "Nenhum set aberto"})

	if geaendert.Fehler != nil || geaendert.Paket.Uebersetzt != 1 || geaendert.Paket.Lizenz != lizenz ||
		geaendert.Texte["allgemein.arbeitsplatz.keinSet"].Quelle != "Kein Set offen" {
		t.Fatalf("geändert: %+v", geaendert.Paket)
	}

	// **Der Befund steht am Satz** — dort zeigt ihn der Assistent.
	falsch := anwendung.SprachpaketAendern(ctx, "pt-BR", app.SprachpaketangabenDTO{},
		map[string]string{"allgemein.arbeitsplatz.keinSet": "<b>x</b>"})
	if falsch.Fehler == nil || falsch.Fehler.Feld != "allgemein.arbeitsplatz.keinSet" ||
		falsch.Fehler.Schluessel != domain.SchluesselAuszeichnungUnsicher {
		t.Errorf("falsch: %+v", falsch.Fehler)
	}

	liste := anwendung.Sprachpakete(ctx)
	if liste.Fehler != nil || len(liste.Pakete) != 1 || liste.Pakete[0].Uebersetzt != 1 {
		t.Errorf("die Liste: %+v", liste)
	}

	text := anwendung.SprachpaketAusgabeText(ctx, "pt-BR")
	if text.Fehler != nil || text.Name != "pt-BR.sprachpaket.json" ||
		!strings.Contains(text.Inhalt, domain.Sprachpaketformat) {
		t.Fatalf("als Text: %+v", text)
	}

	ziel := filepath.Join(t.TempDir(), "pt.json")
	if geschrieben := anwendung.SprachpaketAusgabeSchreiben(ctx, "pt-BR", ziel); geschrieben.Fehler != nil ||
		geschrieben.Pfad != ziel {
		t.Fatalf("geschrieben: %+v", geschrieben.Fehler)
	}

	if geloescht := anwendung.SprachpaketLoeschen(ctx, "pt-BR"); geloescht.Fehler != nil {
		t.Fatalf("gelöscht: %+v", geloescht.Fehler)
	}

	if gezeigt := anwendung.Sprachpaket(ctx, "pt-BR"); gezeigt.Fehler == nil {
		t.Error("ein gelöschtes Paket ließ sich zeigen")
	}

	gelesen := anwendung.SprachpaketEinlesenLesen(ctx, ziel, false)
	if gelesen.Fehler != nil || gelesen.Paket.Uebersetzt != 1 || gelesen.Pfad != ziel {
		t.Fatalf("eingelesen: %+v", gelesen)
	}

	inhalt, err := os.ReadFile(ziel) //nolint:gosec // die eigene Probedatei
	if err != nil {
		t.Fatal(err)
	}

	if doppelt := anwendung.SprachpaketEinlesenText(ctx, string(inhalt), false); doppelt.Fehler == nil ||
		doppelt.Fehler.Feld != "kennung" {
		t.Errorf("doppelt: %+v", doppelt.Fehler)
	}

	if ersetzt := anwendung.SprachpaketEinlesenText(ctx, string(inhalt), true); ersetzt.Fehler != nil {
		t.Errorf("ersetzt: %+v", ersetzt.Fehler)
	}
}

// **Ohne Naht antwortet jeder Weg mit einem Befund und leeren Listen.**
func TestOhneSprachpaketeAntwortetJederWegMitEinemBefund(t *testing.T) {
	anwendung := app.NeueApp("")
	ctx := probekontext(t)

	if antwort := anwendung.Sprachpakete(ctx); antwort.Fehler == nil || antwort.Pakete == nil {
		t.Errorf("die Liste: %+v", antwort)
	}

	for name, antwort := range map[string]app.SprachpaketAntwort{
		"zeigen":    anwendung.Sprachpaket(ctx, "da"),
		"anlegen":   anwendung.SprachpaketAnlegen(ctx, "da", "Dansk", "en"),
		"ändern":    anwendung.SprachpaketAendern(ctx, "da", app.SprachpaketangabenDTO{}, nil),
		"schreiben": anwendung.SprachpaketAusgabeSchreiben(ctx, "da", "/x"),
		"lesen":     anwendung.SprachpaketEinlesenLesen(ctx, "/x", false),
		"als Text":  anwendung.SprachpaketEinlesenText(ctx, "{}", false),
	} {
		if antwort.Fehler == nil || antwort.Texte == nil || antwort.FehlendePfade == nil ||
			antwort.VeraltetePfade == nil || antwort.UnbekanntePfade == nil {
			t.Errorf("%s: %+v", name, antwort)
		}
	}

	if antwort := anwendung.SprachpaketAusgabeText(ctx, "da"); antwort.Fehler == nil {
		t.Errorf("als Text: %+v", antwort)
	}

	if antwort := anwendung.SprachpaketLoeschen(ctx, "da"); antwort.Fehler == nil {
		t.Errorf("löschen: %+v", antwort)
	}
}

// **Ein Fehler des Dienstes kommt als Befund an** — beim Ausgeben einer
// unbekannten Übersetzung und bei einem Bestand, der nicht mehr antwortet.
func TestEinFehlerDerUebersetzungenKommtAlsBefundAn(t *testing.T) {
	db := offeneProbedatenbank(t)
	anwendung := app.NeueApp("")
	anwendung.MitSprachpaketen(app.NeueSprachpaketeUeber(db))
	ctx := probekontext(t)

	if antwort := anwendung.SprachpaketAusgabeSchreiben(ctx, "xx", filepath.Join(t.TempDir(), "x.json")); antwort.Fehler == nil ||
		antwort.Pfad != "" {
		t.Errorf("geschrieben: %+v", antwort)
	}

	if antwort := anwendung.SprachpaketAusgabeText(ctx, "xx"); antwort.Fehler == nil {
		t.Errorf("als Text: %+v", antwort)
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	if antwort := anwendung.Sprachpakete(ctx); antwort.Fehler == nil || antwort.Pakete == nil {
		t.Errorf("die Liste: %+v", antwort)
	}
}
