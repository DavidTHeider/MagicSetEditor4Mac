package app_test

// **Das Schema einer Karte** — Werkstück A7 von Phase 5.3
// (#1446).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Nur Text- und Wahlfelder kommen ins Schema.**
//
// # Warum das eine Aufzählung ist und kein „alles außer"
//
// **Weil jede Zeile einen Grund hat.** Ein Sprachmodell schreibt keine Datei
// ins Paket (Bild, Symbol), keine Farbe und keine Paketwahl; `info` ist gar
// nicht beschreibbar. Ein Feld, das es gibt und das niemand setzen darf,
// bleibt trotzdem im Schema — der Unterschied zwischen „kenne ich nicht" und
// „darfst du nicht" ist die Meldung, die weiterhilft.
func TestNurTextUndWahlKommenInsSchema(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.stile([]domain.Field{
		{Name: "name", Kind: domain.KindText},
		{Name: "notes", Kind: domain.KindUnknown},
		{Name: "rarity", Kind: domain.KindChoice, Choices: []string{"common", "rare"}},
		{Name: "farben", Kind: domain.KindMultipleChoice, Choices: []string{"W", "U"}},
		{Name: "image", Kind: domain.KindImage},
		{Name: "symbol", Kind: domain.KindSymbol},
		{Name: "farbe", Kind: domain.KindColor},
		{Name: "stil", Kind: domain.KindPackageChoice},
		{Name: "haken", Kind: domain.KindBoolean},
		{Name: "hinweis", Kind: domain.KindInfo},
		{Name: "unbekannt", Kind: domain.FieldKind("etwasneues")},
	})

	if antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "x", ""); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	eigenschaften, gibts := welt.laufzeit.schemen[0]["properties"].(map[string]any)
	if !gibts {
		t.Fatalf("das Schema ist %v", welt.laufzeit.schemen[0])
	}

	for _, name := range []string{"name", "notes", "rarity", "farben"} {
		if _, drin := eigenschaften[name]; !drin {
			t.Errorf("%q fehlt im Schema: %v", name, eigenschaften)
		}
	}

	for _, name := range []string{
		"image", "symbol", "farbe", "stil", "haken", "hinweis", "unbekannt",
	} {
		if _, drin := eigenschaften[name]; drin {
			t.Errorf("%q steht im Schema: %v", name, eigenschaften)
		}
	}

	// **Ein Mehrfachfeld ist eine Liste, ein Wahlfeld eine Aufzählung.**
	mehrfach, istKarte := eigenschaften["farben"].(map[string]any)
	if !istKarte || mehrfach["type"] != "array" {
		t.Errorf("»farben« ist %v", eigenschaften["farben"])
	}

	wahl, istWahl := eigenschaften["rarity"].(map[string]any)
	if !istWahl || wahl["enum"] == nil {
		t.Errorf("»rarity« ist %v", eigenschaften["rarity"])
	}
}

// **Ohne Bestand gibt es kein Schema.**
func TestOhneBestandGibtEsKeinSchema(t *testing.T) {
	t.Parallel()

	if antwort := app.NeueApp("").KIVorschlagen(probekontext(t), 4,
		"ollama", "gemma4:8b", "x", ""); antwort.Fehler == nil {
		t.Fatal("ohne Bestand kam ein Vorschlag")
	}
}

// **Eine Karte, die sich nicht lesen lässt, gibt kein Schema.**
func TestEineUnlesbareKarteGibtKeinSchema(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.karten.fehler = errors.New("die Karte ist nicht zu lesen")

	if antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "x", ""); antwort.Fehler == nil {
		t.Fatal("die unlesbare Karte kam durch")
	}
}

// **Ein Paket, das seine Felder nicht hergibt, ebenso.**
func TestEinUnlesbaresPaketGibtKeinSchema(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.anwendung.MitStilwahl(stildoppel{
		fehler: errors.New("das Paket ist nicht zu lesen"),
	})

	if antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "x", ""); antwort.Fehler == nil {
		t.Fatal("das unlesbare Paket kam durch")
	}
}

// **Ein Paket ohne Felder ist kein leeres Schema, sondern ein Fehler.**
//
// **A6 sagt es scharf**: „Fehlendes/unlesbares Paket erzeugt für KI **keinen
// beliebigen Feldfreibrief**." Ein leeres Schema wiese zwar alles ab — aber
// es sähe aus wie eine Antwort, und die Meldung, die weiterhilft, fehlte.
func TestEinPaketOhneFelderGibtKeinSchema(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.stile(nil)

	antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "x", "")

	if antwort.Fehler == nil {
		t.Fatal("das Paket ohne Felder kam durch")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "Feld") {
		t.Errorf("die Meldung ist %q", antwort.Fehler.Meldung)
	}
}
