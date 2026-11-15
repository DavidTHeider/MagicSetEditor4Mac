package service_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

const testdaten = "../../testdaten"

// speicherDouble ist ein In-Memory-Double der Schnittstelle (AGENTS.md §3).
type speicherDouble struct {
	set    *domain.Set
	roh    []domain.RawBlock
	fehler error
}

func (s *speicherDouble) SetAnlegen(
	_ context.Context, set *domain.Set, roh []domain.RawBlock,
) (int64, error) {
	if s.fehler != nil {
		return 0, s.fehler
	}

	s.set, s.roh = set, roh

	return 42, nil
}

func TestImportLiestKartenInIhrerReihenfolge(t *testing.T) {
	speicher := &speicherDouble{}

	ergebnis, err := service.NeuerImporteur(speicher).
		Set(t.Context(), filepath.Join(testdaten, "probe.mse-set"))
	if err != nil {
		t.Fatalf("Import: %v", err)
	}

	if ergebnis.SetID != 42 {
		t.Errorf("Kennung %d", ergebnis.SetID)
	}
	if ergebnis.Karten != 3 {
		t.Fatalf("%d Karten, erwartet waren 3", ergebnis.Karten)
	}

	erwartet := []string{"Feuerball", "Eisball", "Doppelt"}
	for i, name := range erwartet {
		text, vorhanden := speicher.set.Cards[i].Value("name")
		if !vorhanden || text != name {
			t.Errorf("Karte %d heißt %q, erwartet war %q", i+1, text, name)
		}
	}
}

func TestImportUebernimmtSetangaben(t *testing.T) {
	speicher := &speicherDouble{}

	if _, err := service.NeuerImporteur(speicher).
		Set(t.Context(), filepath.Join(testdaten, "probe.mse-set")); err != nil {
		t.Fatalf("Import: %v", err)
	}

	if speicher.set.Name != "Mein Probeset" {
		t.Errorf("Setname %q", speicher.set.Name)
	}
	if speicher.set.Game != "probe" {
		t.Errorf("Spiel %q", speicher.set.Game)
	}
	if speicher.set.Style != "probe-stil" {
		t.Errorf("Stil %q", speicher.set.Style)
	}
	if titel, _ := speicher.set.Value("copyright"); titel != "frei" {
		t.Errorf("copyright %q", titel)
	}
}

// Was wir nicht deuten, geht nicht verloren.
//
// Das ist die Rundlaufzusage aus Entscheidung 0002: ein unverstandener Block
// wird roh mitgeführt und beim Schreiben an seiner Stelle wieder ausgegeben.
// Ohne ihn öffnet ein zurückgeschriebenes Set in MSE2 nicht mehr.
func TestEinUnverstandenerBlockWirdRohMitgefuehrt(t *testing.T) {
	speicher := &speicherDouble{}

	if _, err := service.NeuerImporteur(speicher).
		Set(t.Context(), filepath.Join(testdaten, "probe.mse-set")); err != nil {
		t.Fatalf("Import: %v", err)
	}

	// **Einer, seit der Setkopf gedeutet wird** (#1325). Hier standen zwei:
	// `mse version: 2.0.0` war der erste, und er reiste roh mit, weil niemand
	// ihn las. Seit dem 2026-09-03 steht er im Bestand — was hier zählt, ist
	// der **unverstandene** Block darunter.
	if len(speicher.roh) != 1 {
		t.Fatalf("%d Rohblöcke, erwartet war einer", len(speicher.roh))
	}

	// **Wörtlich**, nicht rekonstruiert: Abstände, Einrückung und Zeilenende
	// stehen so da, wie sie in der Datei standen.
	block := speicher.roh[0]
	if erwartet := "unverstandener block:\n\tirgendwas: bleibt liegen\n"; block.Text != erwartet {
		t.Errorf("der Rohblock ist %q, erwartet war %q", block.Text, erwartet)
	}
	if block.Line == 0 {
		t.Error("der Rohblock nennt keine Zeile")
	}
}

func TestEinPaketDasKeinSetIstWirdAbgewiesen(t *testing.T) {
	_, err := service.NeuerImporteur(&speicherDouble{}).
		Set(t.Context(), filepath.Join(testdaten, "probe.mse-game"))

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war eine Validierung", err)
	}
}

func TestEinFehlendesPaketWirdGemeldet(t *testing.T) {
	if _, err := service.NeuerImporteur(&speicherDouble{}).
		Set(t.Context(), filepath.Join(t.TempDir(), "gibtsnicht.mse-set")); err == nil {
		t.Error("ein fehlendes Paket wurde angenommen")
	}
}

func TestEinFehlerBeimAblegenWirdDurchgereicht(t *testing.T) {
	eigener := errors.New("die Platte ist voll")

	_, err := service.NeuerImporteur(&speicherDouble{fehler: eigener}).
		Set(t.Context(), filepath.Join(testdaten, "probe.mse-set"))

	if !errors.Is(err, eigener) {
		t.Errorf("Fehler %v, erwartet war der eigene", err)
	}
}

// Ein Set ohne Titel heißt wie seine Datei.
//
// Ein Set ohne Titel ist keine Ausnahme — MSE legt eines an, sobald jemand auf
// „Neu" klickt. Ohne Rückfall hieße es in der Liste gar nicht.
func TestEinSetOhneTitelHeisstWieSeineDatei(t *testing.T) {
	speicher := &speicherDouble{}

	if _, err := service.NeuerImporteur(speicher).
		Set(t.Context(), filepath.Join(testdaten, "karg.mse-set")); err != nil {
		t.Fatalf("Import: %v", err)
	}

	if speicher.set.Name != "karg.mse-set" {
		t.Errorf("Setname %q, erwartet war der Dateiname", speicher.set.Name)
	}
	if speicher.set.Style != "" {
		t.Errorf("Stil %q — das Set nennt keinen", speicher.set.Style)
	}
}

// Eine Karte ohne einen einzigen Feldwert ist ein Befund, kein Abbruch.
//
// Sie kommt in echten Sets vor: jemand legt eine Karte an und füllt sie nicht.
// Sie zu übernehmen hieße, eine leere Zeile in die Liste zu schreiben; sie
// abzuweisen hieße, das ganze Set nicht zu öffnen.
func TestEineKarteOhneWerteIstEinBefund(t *testing.T) {
	speicher := &speicherDouble{}

	ergebnis, err := service.NeuerImporteur(speicher).
		Set(t.Context(), filepath.Join(testdaten, "karg.mse-set"))
	if err != nil {
		t.Fatalf("Import: %v", err)
	}

	// Die leere Karte bleibt eine Karte — der Befund sagt, dass etwas fehlt,
	// und wirft den Datensatz nicht weg.
	if ergebnis.Karten != 2 {
		t.Errorf("%d Karten, erwartet waren 2 (eine davon leer)", ergebnis.Karten)
	}
	if len(ergebnis.Befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet war einer", len(ergebnis.Befunde))
	}
	if !strings.Contains(ergebnis.Befunde[0].Meldung, "ohne einen einzigen Feldwert") {
		t.Errorf("Befund %q", ergebnis.Befunde[0].Meldung)
	}
}

// Ein Kommentar auf Wurzelebene ist kein unverstandener Block.
//
// Er trägt keinen Schlüssel; ihn als Rohblock mitzuführen hieße, ihn zweimal zu
// schreiben — einmal als Kommentar, einmal als Block.
func TestEinKommentarIstKeinRohblock(t *testing.T) {
	speicher := &speicherDouble{}

	if _, err := service.NeuerImporteur(speicher).
		Set(t.Context(), filepath.Join(testdaten, "karg.mse-set")); err != nil {
		t.Fatalf("Import: %v", err)
	}

	// **Drei Blöcke seit #1325**: der unverstandene, der Kommentar der
	// Wurzelebene und der Kommentar in der Karte. Vier waren es, solange
	// `mse version` roh mitreiste; seit dem 2026-09-03 steht er im Bestand.
	if len(speicher.roh) != 3 {
		t.Fatalf("%d Rohblöcke, erwartet waren 3", len(speicher.roh))
	}

	// Und der Rohblock führt seine Leerzeile und seinen Kommentar mit —
	// sonst käme das Set beim Zurückschreiben verändert heraus.
	inhalt := ""

	for _, block := range speicher.roh {
		if strings.HasPrefix(block.Text, "unverstanden:") {
			inhalt = block.Text
		}
	}
	for _, teil := range []string{"unverstanden:", "\tdrin: wert", "\t# Kommentar im Block", "\tnoch: einer"} {
		if !strings.Contains(inhalt, teil) {
			t.Errorf("der Rohblock führt %q nicht mit:\n%s", teil, inhalt)
		}
	}
}

// Ein Kommentar innerhalb einer Karte wird kein Feldwert.
//
// Er trägt keinen Schlüssel; ihn als Feld zu lesen ergäbe einen Wert namens
// `# ein Kommentar…`, der beim Zurückschreiben als Feld herauskäme.
func TestEinKommentarInDerKarteWirdKeinFeldwert(t *testing.T) {
	speicher := &speicherDouble{}

	if _, err := service.NeuerImporteur(speicher).
		Set(t.Context(), filepath.Join(testdaten, "karg.mse-set")); err != nil {
		t.Fatalf("Import: %v", err)
	}

	karte := speicher.set.Cards[1]
	if len(karte.Values) != 2 {
		t.Fatalf("%d Feldwerte, erwartet waren 2 (name, farbe)", len(karte.Values))
	}
	for _, wert := range karte.Values {
		if strings.HasPrefix(wert.Field, "#") {
			t.Errorf("ein Kommentar wurde als Feld gelesen: %+v", wert)
		}
	}
}

// Ein Set ohne Spiel wird abgewiesen, bevor der Bestand schreibt.
//
// Bis zum Zweitblick über Phase 2 landete es still mit `spiel = ”` in der
// Datenbank: das Schema verlangt nur `NOT NULL`, nicht „nicht leer". Eine
// fachliche Regel, die nur in der Domäne steht und die niemand ruft, ist keine.
func TestEinSetOhneSpielErreichtDenBestandNicht(t *testing.T) {
	speicher := &speicherDouble{}

	_, err := service.NeuerImporteur(speicher).
		Set(t.Context(), filepath.Join(testdaten, "ohnespiel.mse-set"))

	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Fehler %v, erwartet war eine Validierung", err)
	}
	if speicher.set != nil {
		t.Error("das Set wurde trotz Befund abgelegt")
	}
}

// **Die Quelldatei reist mit** ([0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)).
//
// Ohne sie ist die Byte-Zusage aus #48 nicht zu halten: aus `card_values` allein
// kommen Einrückung, Abstände, Kommentare und Leerzeilen nicht zurück.
func TestDerImportFuehrtDieQuelldateiMit(t *testing.T) {
	speicher := &speicherDouble{}

	if _, err := service.NeuerImporteur(speicher).
		Set(t.Context(), filepath.Join(testdaten, "probe.mse-set")); err != nil {
		t.Fatalf("Import: %v", err)
	}

	if speicher.set.Quelle == nil {
		t.Fatal("keine Quelle — die Byte-Zusage ist damit nicht zu halten")
	}

	if speicher.set.Quelle.Datei != "set" {
		t.Errorf("Datei: %q", speicher.set.Quelle.Datei)
	}

	// **Byte für Byte, samt BOM**: `probe.mse-set/set` beginnt mit einer.
	roh, err := os.ReadFile(filepath.Join(testdaten, "probe.mse-set", "set"))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if !bytes.Equal(speicher.set.Quelle.Inhalt, roh) {
		t.Errorf("die Quelle stimmt nicht mit der Datei überein")
	}
}

// **Der `styling:`-Block wird ein Setwert** —
// #557.
//
//	IndexMap<FieldP, ValueP>& Set::stylingDataFor(const StyleSheet& stylesheet) {
//	  return styling_data.get(stylesheet.name(), stylesheet.styling_fields);
//	}
//
// `src/data/set.cpp`, Zeilen 99–101, Fassung f9d9356 vom 2021-07-06 —
// **geschlüsselt nach dem Namen des Stilpakets**.
//
// **Flach abgelegt, eine Zeile je Option.** Derselbe Weg wie bei der Karte,
// deren `styling_data:` als gewöhnlicher Feldwert liegt; eine eigene Tabelle
// wäre eine zweite Antwort auf dieselbe Frage.
func TestDerSetweiteStilblockWirdEinSetwert(t *testing.T) {
	speicher := &speicherDouble{}

	if _, err := service.NeuerImporteur(speicher).
		Set(t.Context(), filepath.Join(testdaten, "setweit.mse-set")); err != nil {
		t.Fatalf("Import: %v", err)
	}

	gefunden := ""

	for _, wert := range speicher.set.Values {
		if wert.Field == "styling" {
			gefunden = wert.Text
		}
	}

	if gefunden == "" {
		t.Fatalf("kein Setwert `styling`; gespeichert: %+v", speicher.set.Values)
	}

	for _, erwartet := range []string{
		"probe-probe-stil/probe: setweit",
		"probe-probe-stil/andere option: zwei",
		"probe-anders/probe: falsch",
	} {
		if !strings.Contains(gefunden, erwartet) {
			t.Errorf("%q fehlt in\n%s", erwartet, gefunden)
		}
	}

	// **Und die Titelzeile des Sets bleibt, wo sie war** — der neue Zweig
	// nimmt `set info:` nichts weg.
	if speicher.set.Name != "Setweite Stiloptionen" {
		t.Errorf("Setname %q", speicher.set.Name)
	}
}

// TestEinVonMseGeschriebenesSetBehaeltSeinenTitel ist die zweite sichtbare
// Folge des Nachschlags aus #396.
//
// **Jedes von MSE geschriebene Set trägt `set_info:` mit Unterstrich** — in
// allen zehn kanonischen Sets, gemessen am 2026-08-24. `namenAus` schlug
// `set info` nach und fand nichts; ein Set mit Titel hieß nach seiner Datei.
//
// Dass es niemandem auffiel, hat einen Grund, und er begrenzt die Aussage
// dieses Tests: **keines der zehn kanonischen Sets trägt einen Titel**. Der
// Fehler war da und blieb unsichtbar — die Vorrichtung `setweit.mse-set`
// schreibt `set info:` mit Leerzeichen, so wie eine Vorlage es täte, und
// niemand schreibt eine Vorrichtung so, wie das fremde Programm schreibt.
func TestEinVonMseGeschriebenesSetBehaeltSeinenTitel(t *testing.T) {
	t.Parallel()

	verzeichnis := filepath.Join(t.TempDir(), "Datei.mse-set")
	if err := os.MkdirAll(verzeichnis, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	inhalt := "mse_version: 2.1.2\ngame: probe\nset_info:\n\ttitle: Kanonisch Geschrieben\n"
	if err := os.WriteFile(filepath.Join(verzeichnis, "set"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	speicher := &speicherDouble{}
	if _, err := service.NeuerImporteur(speicher).Set(t.Context(), verzeichnis); err != nil {
		t.Fatalf("Import: %v", err)
	}

	if speicher.set.Name != "Kanonisch Geschrieben" {
		t.Errorf("das Set heißt %q, will den Titel aus set_info", speicher.set.Name)
	}
}
