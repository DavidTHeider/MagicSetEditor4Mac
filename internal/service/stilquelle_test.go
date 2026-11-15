package service_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// koepfequelle gibt die Setköpfe, die ein Test vorgibt.
type koepfequelle struct {
	koepfe      []domain.SetHead
	werte       []domain.Value
	fehler      error
	wertefehler error
	gefragt     int
}

func (k *koepfequelle) Setwerte(context.Context, int64) ([]domain.Value, error) {
	return k.werte, k.wertefehler
}

func (k *koepfequelle) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	k.gefragt++

	return k.koepfe, k.fehler
}

// saatwurzel ist das Verzeichnis, in dem die eigene Saat liegt.
func saatwurzel() string { return filepath.Join("..", "..", "saat") }

// **Der Weg vom Set zum Stil, und zwar ganz.**
//
// Er war bis #206
// nirgends verdrahtet: der Stilplan stand, und niemand fand das Paket dazu.
func TestDerStilspeicherFindetDenStilEinesSets(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{
		{ID: 7, Style: "probe-schlicht", Game: "probe"},
	}}

	speicher := service.NeuerStilspeicher(quelle, saatwurzel())

	vorlage, befunde, da := speicher.StilFuerSet(t.Context(), 7)
	if !da {
		t.Fatalf("kein Stil gefunden, Befunde: %v", befunde)
	}

	if len(vorlage.Stil.Felder) == 0 {
		t.Error("der gefundene Stil führt keine Felder")
	}
}

// **Und das Spielpaket kommt mit** (#195).
//
// Ohne sein `init script:` bleiben die Feldstile unausrechenbar, die seine
// Funktionen rufen — bei `M15-Magic-Pack` sind das Hunderte.
func TestDerStilspeicherBringtDasSpielskriptMit(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{
		{ID: 7, Style: "probe-spielskript", Game: "probe"},
	}}

	vorlage, befunde, da := service.NeuerStilspeicher(quelle, saatwurzel()).
		StilFuerSet(t.Context(), 7)
	if !da {
		t.Fatalf("kein Stil gefunden, Befunde: %v", befunde)
	}

	if vorlage.Spielskript == "" {
		t.Errorf("kein Spielskript; Befunde: %v", befunde)
	}
}

// **Ein fehlendes Spielpaket wird benannt, nicht verschwiegen.**
func TestEinFehlendesSpielpaketWirdBenannt(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{
		{ID: 7, Style: "probe-schlicht", Game: "gibtesnicht"},
	}}

	vorlage, befunde, da := service.NeuerStilspeicher(quelle, saatwurzel()).
		StilFuerSet(t.Context(), 7)
	if !da {
		t.Fatalf("kein Stil gefunden, Befunde: %v", befunde)
	}

	if vorlage.Spielskript != "" || len(befunde) == 0 {
		t.Errorf("Skript %q, Befunde %v", vorlage.Spielskript, befunde)
	}
}

// **Ein Set ohne Spielangabe kostet keinen Befund.**
//
// Es ist keine Störung: ein aus MTG JSON entstandenes Set nennt kein
// Stilpaket, und ein Set ohne Spiel gibt es aus demselben Grund.
func TestOhneSpielangabeKeinBefund(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Style: "probe-schlicht"}}}

	_, befunde, da := service.NeuerStilspeicher(quelle, saatwurzel()).
		StilFuerSet(t.Context(), 7)
	if !da {
		t.Fatalf("kein Stil gefunden, Befunde: %v", befunde)
	}

	if len(befunde) != 0 {
		t.Errorf("%v — ein Set ohne Spielangabe ist kein Befund", befunde)
	}
}

// **Ein zweites Mal fragen kostet kein Paket.**
//
// Eine Vorschau entsteht bei jedem Tastendruck; wer dabei jedes Mal ein Paket
// von der Platte liest und seine Module einbindet, zahlt dieselbe Rechnung, die
// #187 an der
// Bildquelle aufgemacht hat.
//
// **Der Kopf dagegen je Ruf** (#2005): er ist billig und sagt, ob das
// Gemerkte noch zum Set gehört (`stilkopf_intern_test.go`).
func TestDerStilspeicherLiestNurEinmal(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Style: "probe-schlicht"}}}
	speicher := service.NeuerStilspeicher(quelle, saatwurzel())

	erste, _, da := speicher.StilFuerSet(t.Context(), 7)
	if !da {
		t.Fatal("kein Stil gefunden")
	}

	for range 2 {
		weitere, _, da := speicher.StilFuerSet(t.Context(), 7)
		if !da {
			t.Fatal("kein Stil gefunden")
		}

		// **Dieselbe Vorlage**: ihre Felder liegen an derselben Stelle.
		if &weitere.Stil.Felder[0] != &erste.Stil.Felder[0] {
			t.Error("die Vorlage wurde neu gelesen")
		}
	}

	if quelle.gefragt != 3 {
		t.Errorf("%d Abfragen der Setköpfe, erwartet je Ruf eine", quelle.gefragt)
	}
}

// **Ein Stil, den es nicht gibt, ist ein Befund — kein Fehler.**
//
// Die Vorschau fällt auf den Behelfsplan zurück und sagt, warum. Ein Fehler
// ließe die Seite leer, und das ist die schlechtere Antwort (§8).
func TestEinUnbekannterStilWirdBenannt(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Style: "gibtesnicht"}}}
	speicher := service.NeuerStilspeicher(quelle, saatwurzel())

	_, befunde, da := speicher.StilFuerSet(t.Context(), 7)
	if da {
		t.Fatal("es gibt kein Paket dieses Namens - trotzdem kam ein Stil")
	}

	if len(befunde) == 0 {
		t.Error("kein Befund — dann sieht niemand, warum der Behelfsplan kommt")
	}
}

// **Auch der Fehlschlag wird gemerkt.**
//
// Sonst geht jede Vorschau eines Sets ohne Stilpaket wieder über die Platte —
// und das ist der häufige Fall, nicht der seltene.
func TestAuchEinFehlschlagWirdGemerkt(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Style: "gibtesnicht"}}}
	speicher := service.NeuerStilspeicher(quelle, saatwurzel())

	for range 3 {
		if _, _, da := speicher.StilFuerSet(t.Context(), 7); da {
			t.Fatal("unerwartet ein Stil")
		}
	}

	// **Je Ruf ein Kopf**, nicht ein Paket (`stilkopf_intern_test.go`).
	if quelle.gefragt != 3 {
		t.Errorf("%d Abfragen, erwartet je Ruf eine", quelle.gefragt)
	}
}

// **Ein Set ohne Stilangabe wird benannt, nicht gesucht.**
func TestEinSetOhneStilangabe(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Style: ""}}}
	speicher := service.NeuerStilspeicher(quelle, saatwurzel())

	if _, befunde, da := speicher.StilFuerSet(t.Context(), 7); da || len(befunde) == 0 {
		t.Errorf("da=%v Befunde=%v", da, befunde)
	}
}

// **Ein Lesefehler des Bestands wird gemeldet und nicht gemerkt.**
//
// Er sagt nichts über das Stilpaket — ihn zu merken hieße, einen vorübergehenden
// Zustand dauerhaft zu machen.
func TestEinLesefehlerWirdNichtGemerkt(t *testing.T) {
	quelle := &koepfequelle{fehler: errors.New("Bestand nicht lesbar")}
	speicher := service.NeuerStilspeicher(quelle, saatwurzel())

	for range 2 {
		if _, befunde, da := speicher.StilFuerSet(t.Context(), 7); da || len(befunde) == 0 {
			t.Fatalf("da=%v Befunde=%v", da, befunde)
		}
	}

	if quelle.gefragt != 2 {
		t.Errorf("%d Abfragen, erwartet waren 2 — ein Lesefehler gehört nicht gemerkt",
			quelle.gefragt)
	}
}

// **Ein Set, das der Bestand nicht führt, ist ein Befund.**
func TestEinUnbekanntesSet(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 1, Style: "probe-schlicht"}}}
	speicher := service.NeuerStilspeicher(quelle, saatwurzel())

	if _, befunde, da := speicher.StilFuerSet(t.Context(), 99); da || len(befunde) == 0 {
		t.Errorf("da=%v Befunde=%v", da, befunde)
	}
}

// **Ohne Wurzel wird gar nicht erst gesucht.**
//
// So sieht der Start aus, solange niemand ein Paket bezogen hat.
func TestOhneWurzelKeinStil(t *testing.T) {
	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Style: "probe-schlicht"}}}
	speicher := service.NeuerStilspeicher(quelle)

	if _, befunde, da := speicher.StilFuerSet(t.Context(), 7); da || len(befunde) == 0 {
		t.Errorf("da=%v Befunde=%v", da, befunde)
	}
}

func TestEinKaputtesStilpaketWirdAlsUnlesbarBenannt(t *testing.T) {
	wurzel := t.TempDir()
	if err := os.Mkdir(filepath.Join(wurzel, "kaputt.mse-style"), 0o750); err != nil {
		t.Fatalf("Paket anlegen: %v", err)
	}

	quelle := &koepfequelle{koepfe: []domain.SetHead{{ID: 7, Style: "kaputt"}}}
	_, befunde, da := service.NeuerStilspeicher(quelle, wurzel).StilFuerSet(t.Context(), 7)

	if da {
		t.Fatal("ein Paket ohne Hauptdatei ergab einen Stil")
	}
	if len(befunde) == 0 || !strings.Contains(befunde[0].Meldung, "nicht öffnen") {
		t.Errorf("Befunde %v — das vorhandene kaputte Paket gilt nur als fehlend", befunde)
	}
}

func TestEinKaputtesSpielpaketWirdAlsUnlesbarBenannt(t *testing.T) {
	wurzel := t.TempDir()
	stilpaket := filepath.Join(wurzel, "probe-stil.mse-style")
	if err := os.Mkdir(stilpaket, 0o750); err != nil {
		t.Fatalf("Stilpaket anlegen: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stilpaket, "style"), []byte(
		"mse version: 2.0.0\ngame: probe\ncard width: 375\ncard height: 523\n"), 0o600); err != nil {
		t.Fatalf("Stil schreiben: %v", err)
	}
	if err := os.Mkdir(filepath.Join(wurzel, "probe.mse-game"), 0o750); err != nil {
		t.Fatalf("Spielpaket anlegen: %v", err)
	}

	quelle := &koepfequelle{koepfe: []domain.SetHead{
		{ID: 7, Style: "probe-stil", Game: "probe"},
	}}
	_, befunde, da := service.NeuerStilspeicher(quelle, wurzel).StilFuerSet(t.Context(), 7)

	if !da {
		t.Fatalf("der lesbare Stil ging verloren: %v", befunde)
	}
	if len(befunde) == 0 || !strings.Contains(befunde[0].Meldung, "nicht öffnen") {
		t.Errorf("Befunde %v — das vorhandene kaputte Spielpaket gilt nur als fehlend", befunde)
	}
}

// **Der Stilspeicher bringt die Setangaben mit.**
//
// # Warum sie über diesen Weg kommen
//
// Der Planbauer öffnet keine Pakete und liest keinen Bestand (§2); was er
// braucht, steht in der [service.Vorlage]. Die Setangaben sind das Letzte, was
// dort gefehlt hat — `set.symbol` allein wird über `M15-Magic-Pack` 420-mal
// gelesen.
//
// # Warum an der Kopfquelle und nicht an einer zweiten
//
// Es ist dieselbe Auskunft über dasselbe Set aus derselben Tabelle, und der
// Speicher fragt sie zum selben Zeitpunkt. Eine zweite Schnittstelle daneben
// hieße zwei Wege zu einem Ergebnis.
func TestDerStilspeicherBringtDieSetangabenMit(t *testing.T) {
	quelle := &koepfequelle{
		koepfe: []domain.SetHead{{ID: 7, Style: "probe-schlicht", Game: "probe"}},
		werte:  []domain.Value{{Field: "symbol", Text: "eigenes"}},
	}

	vorlage, befunde, da := service.NeuerStilspeicher(quelle, saatwurzel()).
		StilFuerSet(t.Context(), 7)
	if !da {
		t.Fatalf("kein Stil gefunden: %v", befunde)
	}

	if len(vorlage.Setangaben) != 1 || vorlage.Setangaben[0].Text != "eigenes" {
		t.Errorf("die Vorlage trägt %+v, erwartet das Symbol", vorlage.Setangaben)
	}
}

// **Sind die Angaben nicht lesbar, entsteht der Plan trotzdem** — mit einem
// Befund und ohne sie.
//
// Ein Set ohne Symbol ist eine Karte ohne Seltenheitszeichen; ein Set ohne
// Stil ist keine Karte. Die beiden Fehler wiegen nicht gleich, und deshalb
// hält der eine den anderen nicht auf (§8).
func TestUnlesbareSetangabenHaltenDenStilNichtAuf(t *testing.T) {
	quelle := &koepfequelle{
		koepfe:      []domain.SetHead{{ID: 7, Style: "probe-schlicht", Game: "probe"}},
		wertefehler: errors.New("Bestand weg"),
	}

	vorlage, befunde, da := service.NeuerStilspeicher(quelle, saatwurzel()).
		StilFuerSet(t.Context(), 7)
	if !da {
		t.Fatalf("kein Stil gefunden: %v", befunde)
	}

	if len(vorlage.Setangaben) != 0 {
		t.Errorf("die Vorlage trägt %+v, erwartet nichts", vorlage.Setangaben)
	}

	for _, befund := range befunde {
		if strings.Contains(befund.Meldung, "Bestand weg") {
			return
		}
	}

	t.Errorf("kein Befund nennt den Lesefehler: %v", befunde)
}

// **Eine geänderte Setangabe erscheint beim nächsten Mal.**
//
// # Die Klasse hinter dem Fall
//
// Der Speicher merkt sich die Vorlage je Set und verwirft sie nie — für den
// Paketinhalt geht das durch: ein `.mse-style` ändert sich nicht, während die
// Anwendung läuft. **Die Setangaben ändert der Nutzende**, und zwar im selben
// Fenster: wer das Symbol wechselt, sähe es sonst bis zum Neustart nicht.
//
// Deshalb wird der teure Teil gemerkt — das geöffnete Paket — und der billige
// jedes Mal gelesen. Ein `SELECT` je Vorschau gegen ein Paket je Sitzung.
func TestEineGeaenderteSetangabeErscheintBeimNaechstenMal(t *testing.T) {
	quelle := &koepfequelle{
		koepfe: []domain.SetHead{{ID: 7, Style: "probe-schlicht", Game: "probe"}},
		werte:  []domain.Value{{Field: "symbol", Text: "vorher"}},
	}

	speicher := service.NeuerStilspeicher(quelle, saatwurzel())

	vorher, befunde, da := speicher.StilFuerSet(t.Context(), 7)
	if !da {
		t.Fatalf("kein Stil gefunden: %v", befunde)
	}

	quelle.werte = []domain.Value{{Field: "symbol", Text: "nachher"}}

	vorlage, _, _ := speicher.StilFuerSet(t.Context(), 7)
	if len(vorlage.Setangaben) != 1 || vorlage.Setangaben[0].Text != "nachher" {
		t.Errorf("die Vorlage trägt %+v, erwartet den geänderten Wert", vorlage.Setangaben)
	}

	// **Und das Paket nur einmal gelesen** — die Felder liegen, wo sie lagen.
	if &vorlage.Stil.Felder[0] != &vorher.Stil.Felder[0] {
		t.Error("die Vorlage wurde neu gelesen")
	}
}

// **Der Stilspeicher bringt auch die Setfelder mit.**
//
// Sie stehen im Spielpaket neben den Kartenfeldern und gehen denselben Weg:
// erklärt vom Spiel, ausgefüllt vom Set, gelesen von den Skripten des Stils.
func TestDerStilspeicherBringtDieSetfelderMit(t *testing.T) {
	quelle := &koepfequelle{
		koepfe: []domain.SetHead{{ID: 7, Style: "probe-schlicht", Game: "probe"}},
	}

	vorlage, befunde, da := service.NeuerStilspeicher(quelle, saatwurzel()).
		StilFuerSet(t.Context(), 7)
	if !da {
		t.Fatalf("kein Stil gefunden: %v", befunde)
	}

	if len(vorlage.Setfelder) != 2 {
		t.Errorf("die Vorlage führt %d Setfelder, erwartet zwei: %+v",
			len(vorlage.Setfelder), vorlage.Setfelder)
	}
}
