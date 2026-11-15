package app_test

// **Ein Vorschlag entsteht und ändert nichts** — Werkstück A7 von Phase 5.3
// (#1446).

import (
	"context"
	"database/sql"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// einKoenner baut ein Modell, das strukturiert antworten kann.
func einKoenner(alias string) domain.KIModell {
	modell := einModell(alias)
	modell.Geprueft = append(modell.Geprueft, domain.KIGeprueft{
		Faehigkeit: domain.KIStrukturierteGenerierung, Datum: "2026-09-11",
		Beleg: "Ollama hält über »format« ein JSON-Schema",
	})

	return modell
}

// einSeher ist ein Modell, das **Bilder liest** — A13.3 (#1452).
//
// **Neben `einKoenner` und nicht darin**: eine Fähigkeit, die jedes
// Probemodell hätte, machte den Lauf über „dieser Anbieter sieht nicht"
// unmöglich.
func einSeher(alias string) domain.KIModell {
	modell := einKoenner(alias)
	modell.Geprueft = append(modell.Geprueft, domain.KIGeprueft{
		Faehigkeit: domain.KIVision, Datum: "2026-09-13",
		Beleg: "gemessen an einem Probebild",
	})

	return modell
}

// **Ein Vorschlag entsteht — und die Karte bleibt, wie sie war.**
//
// **Das ist der Satz des Issues**: „Generieren verändert Zielkarte nicht."
func TestEinVorschlagAendertDieKarteNicht(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	vorher := welt.kartenwerte(t)

	antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "Schreibe den Regeltext.", "")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Felder) == 0 || antwort.Kennung == "" {
		t.Fatalf("die Antwort ist %+v", antwort.KIVorschlagDTO)
	}

	if nachher := welt.kartenwerte(t); nachher != vorher {
		t.Errorf("die Karte wurde geändert: %q statt %q", nachher, vorher)
	}
}

// **Erst die Übernahme schreibt.**
func TestErstDieUebernahmeSchreibt(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	vorschlag := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "Schreibe den Regeltext.", "")
	if vorschlag.Fehler != nil {
		t.Fatalf("%+v", vorschlag.Fehler)
	}

	uebernahme := welt.anwendung.KIVorschlagUebernehmen(
		probekontext(t), vorschlag.Kennung, namenDerFelder(vorschlag.Felder))
	if uebernahme.Fehler != nil {
		t.Fatalf("%+v", uebernahme.Fehler)
	}

	if uebernahme.Felder == 0 {
		t.Fatal("es wurde kein Feld geschrieben")
	}

	if werte := welt.kartenwerte(t); !strings.Contains(werte, "Blitzschlag") {
		t.Errorf("die Karte trägt %q", werte)
	}
}

// **Zweimal übernehmen geht nicht.**
//
// **Die Kennung ist konsumierbar** (Issue): ein zweiter Zug legte denselben
// Vorschlag auf einen Stand, der ihn schon trägt.
func TestZweimalUebernehmenGehtNicht(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	vorschlag := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "x", "")
	if vorschlag.Fehler != nil {
		t.Fatalf("%+v", vorschlag.Fehler)
	}

	if erste := welt.anwendung.KIVorschlagUebernehmen(
		probekontext(t), vorschlag.Kennung, namenDerFelder(vorschlag.Felder)); erste.Fehler != nil {
		t.Fatalf("%+v", erste.Fehler)
	}

	if zweite := welt.anwendung.KIVorschlagUebernehmen(
		probekontext(t), vorschlag.Kennung, namenDerFelder(vorschlag.Felder)); zweite.Fehler == nil {
		t.Error("der zweite Zug ging durch")
	}
}

// **Eine abgeschnittene Antwort wird nicht zum Vorschlag.**
//
// **Sie sieht aus wie eine ganze** — das Issue führt sie unter denen, die
// keine Kartenänderung erzeugen dürfen.
func TestEineAbgeschnitteneAntwortWirdKeinVorschlag(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.laufzeit.antwort = app.KIRohantwort{
		Inhalt: `{"name": "Blitz`, Abgeschnitten: true, Modell: "gemma4:8b",
	}

	antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "x", "")

	if antwort.Fehler == nil {
		t.Fatal("die halbe Antwort kam als Vorschlag durch")
	}
}

// **Das Schema geht als JSON-Schema hinaus — und nur mit schreibbaren Feldern.**
func TestDasSchemaGehtMitSchreibbarenFeldernHinaus(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	if antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "x", ""); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(welt.laufzeit.schemen) != 1 {
		t.Fatalf("%d Schemen gesehen", len(welt.laufzeit.schemen))
	}

	eigenschaften, gibts := welt.laufzeit.schemen[0]["properties"].(map[string]any)
	if !gibts {
		t.Fatalf("das Schema ist %v", welt.laufzeit.schemen[0])
	}

	if _, schreibbar := eigenschaften["name"]; !schreibbar {
		t.Errorf("»name« steht nicht im Schema: %v", eigenschaften)
	}

	// **Ein Bildfeld gehört nicht hinein** — ein Sprachmodell schreibt keine
	// Datei ins Paket.
	if _, drin := eigenschaften["image"]; drin {
		t.Errorf("»image« steht im Schema: %v", eigenschaften)
	}
}

// **Ohne eingeschaltete Laufzeit gibt es keinen Vorschlag.**
func TestOhneLaufzeitGibtEsKeinenVorschlag(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.aus(t)

	if antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "x", ""); antwort.Fehler == nil {
		t.Error("ohne Laufzeit kam ein Vorschlag")
	}
}

// **Und ein Modell, das die Fähigkeit nicht belegt, wird abgewiesen.**
//
// **Das ist „eine Modellliste ist keine Fähigkeit" am Schreibweg**: die Karte
// bleibt unberührt, und die Meldung nennt das Modell.
func TestEinModellOhneStrukturierteAusgabeWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.laufzeit.modelle = []domain.KIModell{
		einModell("nurtext:8b"),
	}

	antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "nurtext:8b", "x", "")

	if antwort.Fehler == nil {
		t.Fatal("das Modell ohne Beleg kam durch")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "nurtext:8b") {
		t.Errorf("die Meldung ist %q", antwort.Fehler.Meldung)
	}
}

// kiwelt ist eine Anwendung mit Karte, Paketfeldern und Laufzeit.
type kiwelt struct {
	anwendung *app.App
	laufzeit  *laufzeitdoppel
	karte     int64
	karten    *kartendoppel
	db        *sql.DB

	// felder sind die Paketfelder, mit denen sie anfängt.
	felder []domain.Field
}

// kartenwerte gibt die Felder der Karte als einen Text.
func (w kiwelt) kartenwerte(t *testing.T) string {
	t.Helper()

	teile := make([]string, 0, len(w.karten.zeile.Values))
	for _, wert := range w.karten.zeile.Values {
		teile = append(teile, wert.Field+"="+wert.Text)
	}

	slices.Sort(teile)

	return strings.Join(teile, "|")
}

// stile setzt die Felder des Pakets neu.
func (w kiwelt) stile(felder []domain.Field) {
	w.anwendung.MitStilwahl(stildoppel{felder: felder})
}

// dieProbefelder sind die Felder, mit denen eine Welt anfängt.
func dieProbefelder() []domain.Field {
	return []domain.Field{
		{Name: "name", Kind: domain.KindText},
		{Name: "rarity", Kind: domain.KindChoice, Choices: []string{"common", "rare"}},
		{Name: "image", Kind: domain.KindImage},
	}
}

// aus schaltet die Laufzeit wieder ab.
func (w kiwelt) aus(t *testing.T) {
	t.Helper()

	if fehler := storage.NeueEinstellungen(w.db).Setzen(probekontext(t),
		domain.KILaufzeit, domain.KILaufzeitAus); fehler != nil {
		t.Fatalf("abschalten: %v", fehler)
	}
}

// stildoppel gibt die Felder eines Pakets.
//
// **Die übrigen Methoden kommen aus der eingebetteten Schnittstelle** — sie
// werden hier nie gerufen, und eine Attrappe je Methode wäre eine Liste, die
// mit jedem neuen Ruf wächst, ohne etwas zu prüfen.
type stildoppel struct {
	app.Stilwaehler

	felder []domain.Field
	fehler error
}

func (s stildoppel) Kartenfelder(context.Context, int64) ([]domain.Field, error) {
	return s.felder, s.fehler
}

// einekiwelt baut sie.
func einekiwelt(t *testing.T) kiwelt {
	t.Helper()

	db := frischerBestand(t)

	if fehler := storage.NeueEinstellungen(db).Setzen(probekontext(t),
		domain.KILaufzeit, domain.KILaufzeitAn); fehler != nil {
		t.Fatalf("einschalten: %v", fehler)
	}

	karten := &kartendoppel{zeile: domain.CardRow{
		ID: 4, SetID: 2, Values: []domain.Value{{Field: "name", Text: "Alt"}},
	}}

	laufzeit := &laufzeitdoppel{
		fassung: "0.12.3",
		modelle: []domain.KIModell{einKoenner("gemma4:8b")},
		antwort: app.KIRohantwort{
			Inhalt: `{"name": "Blitzschlag", "rarity": "rare"}`,
			Modell: "gemma4:8b",
		},
	}

	anwendung := app.KIAnwendungUeber(db).MitKIZuhause(t.TempDir()).
		MitKILaufzeit(func(string) app.KILaufzeitquelle { return laufzeit })

	anwendung.MitEditor(karten, karten, nil, nil)
	anwendung.MitStilwahl(stildoppel{felder: dieProbefelder()})

	return kiwelt{
		anwendung: anwendung, laufzeit: laufzeit, karte: 4,
		karten: karten, db: db, felder: dieProbefelder(),
	}
}

// namenDerFelder nennt die Felder eines Vorschlags — die ganze Auswahl.
func namenDerFelder(felder []app.KIFeldwertDTO) []string {
	namen := make([]string, 0, len(felder))
	for _, feld := range felder {
		namen = append(namen, feld.Feld)
	}

	return namen
}
