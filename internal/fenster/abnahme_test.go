package fenster

// **Die Sprache der Leitung** — Tasten, Kürzel, Pfade, Befehle. Die Griffe in
// AppKit stehen in `abnahme_darwin.go` und laufen nur im Abnahmebau.

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestEinTastendruckTraegtCodeZeichenUndSchalter(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		taste    string
		mit      []string
		erwartet Tastendruck
	}{
		{"⌘S", "s", []string{"befehl"}, Tastendruck{Code: 0x01, Zeichen: "s", Ohne: "s", Schalter: 1 << 20}},
		{
			"⇧⌘S macht den Buchstaben groß", "s",
			[]string{"umschalt", "befehl"},
			Tastendruck{Code: 0x01, Zeichen: "S", Ohne: "s", Schalter: 1<<17 | 1<<20},
		},
		{"die Rücktaste ist DEL", "Backspace", nil, Tastendruck{Code: 0x33, Zeichen: "\x7f", Ohne: "\x7f"}},
		{
			"ein Pfeil ist eine Funktionstaste", "ArrowDown", nil,
			Tastendruck{Code: 0x7D, Zeichen: "\uf701", Ohne: "\uf701"},
		},
		{"⌃⌥ zusammen", "q", []string{"ctrl", "wahl"}, Tastendruck{Code: 0x0C, Zeichen: "q", Ohne: "q", Schalter: 1<<18 | 1<<19}},
		{"ein Punkt hat seine Taste", ".", nil, Tastendruck{Code: 0x2F, Zeichen: ".", Ohne: "."}},
		{"ein Umlaut kommt als Zeichen", "ä", nil, Tastendruck{Code: 0, Zeichen: "ä", Ohne: "ä"}},
		// **⇧F10 öffnet das Kontextmenü** (A04, #2218) — `kVK_F10` und
		// `NSF10FunctionKey`, nachgelesen im SDK.
		{
			"⇧F10 ist eine Funktionstaste mit Umschalt", "F10",
			[]string{"umschalt"},
			Tastendruck{Code: 0x6D, Zeichen: "\uf70d", Ohne: "\uf70d", Schalter: 1 << 17},
		},
		// **⇞ und ⇟ blättern durch die Karten** (AP-7, #2364) — `kVK_PageUp`,
		// `kVK_PageDown`, `NSPageUpFunctionKey` und `NSPageDownFunctionKey`,
		// nachgelesen im SDK.
		{"⇞ ist eine Funktionstaste", "PageUp", nil, Tastendruck{Code: 0x74, Zeichen: "\uf72c", Ohne: "\uf72c"}},
		{"⇟ ist eine Funktionstaste", "PageDown", nil, Tastendruck{Code: 0x79, Zeichen: "\uf72d", Ohne: "\uf72d"}},
		// **Pos1 und Ende führen den Spaltentrenner an seine Grenzen** (AP-12,
		// #2369) — `kVK_Home`, `kVK_End`, `NSHomeFunctionKey` und
		// `NSEndFunctionKey`, nachgelesen im SDK.
		{"Pos1 ist eine Funktionstaste", "Home", nil, Tastendruck{Code: 0x73, Zeichen: "\uf729", Ohne: "\uf729"}},
		{"Ende ist eine Funktionstaste", "End", nil, Tastendruck{Code: 0x77, Zeichen: "\uf72b", Ohne: "\uf72b"}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			druck, fehler := tasteAus(fall.taste, fall.mit)
			if fehler != nil {
				t.Fatalf("Fehler: %v", fehler)
			}

			if druck != fall.erwartet {
				t.Errorf("%+v, erwartet %+v", druck, fall.erwartet)
			}
		})
	}
}

// **Wiederholung und Code teilen sich ein Stapelwort** — gemessen am Fenster
// (Hüllenprobe `uxrf1`, 2026-10-03): jede Taste kam mit `code: "KeyA"` an,
// auch Escape, weil der Tastencode als achtes Byte gelesen wurde statt als
// drittes. Nach Apples arm64-Aufrufkonvention liegt `isARepeat` (ein Byte) an
// Stelle 0 des ersten Stapelworts und `keyCode` (zwei Bytes, auf zwei
// ausgerichtet) an Stelle 2.
func TestWiederholungUndCodeTeilenSichEinStapelwort(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		wiederholung bool
		code         uint16
		erwartet     uint64
	}{
		"⌘Q ohne Wiederholung":    {false, 0x0C, 0x0C << 16},
		"Escape als Wiederholung": {true, 0x35, 1 | 0x35<<16},
		"A, der Code 0":           {false, 0x00, 0},
		"der größte Code":         {false, 0xFFFF, 0xFFFF << 16},
	} {
		if wort := stapelwort(fall.wiederholung, fall.code); wort != fall.erwartet {
			t.Errorf("%s: %#x, erwartet %#x", name, wort, fall.erwartet)
		}
	}
}

func TestEineUnbekannteTasteWirdGenannt(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		taste string
		mit   []string
		nennt string
	}{
		"Taste":       {"F13", nil, `"F13"`},
		"Umschalter":  {"s", []string{"hyper"}, `"hyper"`},
		"leere Taste": {"", nil, `""`},
	} {
		_, fehler := tasteAus(fall.taste, fall.mit)
		if !errors.Is(fehler, errUnbekannteTaste) || !strings.Contains(fehler.Error(), fall.nennt) {
			t.Errorf("%s: %v", name, fehler)
		}
	}
}

// **So, wie macOS es in der Leiste schreibt** — ⌃⌥⇧⌘ und der Buchstabe groß.
func TestEinKuerzelStehtWieInDerLeiste(t *testing.T) {
	t.Parallel()

	for _, fall := range []struct {
		aequivalent string
		schalter    uint64
		text        string
	}{
		{"s", 1 << 20, "⌘S"},
		{"s", 1<<17 | 1<<20, "⇧⌘S"},
		{"f", 1<<18 | 1<<20, "⌃⌘F"},
		{"h", 1<<19 | 1<<20, "⌥⌘H"},
		{"", 1 << 20, ""},
	} {
		if text := kuerzelText(fall.aequivalent, fall.schalter); text != fall.text {
			t.Errorf("%q/%d: %q, erwartet %q", fall.aequivalent, fall.schalter, text, fall.text)
		}
	}
}

func leiste() []Menuepunkt {
	return []Menuepunkt{
		{Titel: "MagicSetEditor4Mac", Unter: []Menuepunkt{{Titel: "Über MagicSetEditor4Mac"}}},
		{Titel: "Ablage", Unter: []Menuepunkt{
			{Titel: "Neu"},
			{Trenner: true},
			{Titel: "Sichern unter …"},
			{Titel: "Alle sichern", Verborgen: true},
		}},
	}
}

func TestEinPfadFindetSeineStellen(t *testing.T) {
	t.Parallel()

	stellen, fehler := pfadIn(leiste(), []string{"Ablage", "Sichern unter …"})
	if fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if len(stellen) != 2 || stellen[0] != 1 || stellen[1] != 2 {
		t.Errorf("Stellen %v, erwartet [1 2]", stellen)
	}
}

// **Ein falscher Titel nennt die richtigen** — und ein Trenner ist keiner.
func TestEinFalscherPfadNenntWasDortSteht(t *testing.T) {
	t.Parallel()

	_, fehler := pfadIn(leiste(), []string{"Ablage", "Sichern"})
	if !errors.Is(fehler, errKeinEintrag) {
		t.Fatalf("kein Befund: %v", fehler)
	}

	if !strings.Contains(fehler.Error(), `»Neu«, »Sichern unter …«`) ||
		!strings.Contains(fehler.Error(), "Ebene 2") {
		t.Errorf("nennt die Ebene nicht richtig: %v", fehler)
	}

	// **Ein verborgener Eintrag ist keiner** — die Leiste zeigt ihn nicht,
	// also klickt ihn auch niemand.
	if _, verborgen := pfadIn(leiste(), []string{"Ablage", "Alle sichern"}); !errors.Is(verborgen, errKeinEintrag) ||
		strings.Contains(verborgen.Error(), "»Alle sichern«") {
		t.Errorf("ein verborgener Eintrag wurde gefunden oder genannt: %v", verborgen)
	}

	if _, leer := pfadIn(leiste(), nil); !errors.Is(leer, errKeinEintrag) {
		t.Errorf("ein leerer Pfad fand etwas: %v", leer)
	}
}

func TestEinBefehlWirdStrengGelesen(t *testing.T) {
	t.Parallel()

	befehl, fehler := befehlLesen([]byte(`{"was":"klick","pfad":["Ablage","Neu"]}`))
	if fehler != nil || befehl.Was != "klick" || len(befehl.Pfad) != 2 {
		t.Fatalf("%+v, %v", befehl, fehler)
	}

	// **Ein vertipptes Feld ist ein Fehler**, kein stiller Befehl ohne
	// Wirkung.
	if _, vertippt := befehlLesen([]byte(`{"was":"taste","tast":"s"}`)); vertippt == nil {
		t.Error("ein unbekanntes Feld wurde hingenommen")
	}

	if _, kaputt := befehlLesen([]byte(`kein JSON`)); kaputt == nil {
		t.Error("eine kaputte Zeile wurde hingenommen")
	}
}

func TestEineAntwortTraegtDatenOderFehler(t *testing.T) {
	t.Parallel()

	if gut := antwortAus([]int{1}, nil); !gut.OK || gut.Fehler != "" || gut.Daten == nil {
		t.Errorf("Erfolg: %+v", gut)
	}

	if schlecht := antwortAus(nil, errKeinEintrag); schlecht.OK || schlecht.Fehler == "" {
		t.Errorf("Fehler: %+v", schlecht)
	}
}

// **Ohne vorbereitete Antwort geht der Dialog auf** — im Freigabebau immer,
// im Abnahmebau, solange die Probe keine gelegt hat.
func TestOhneVorbereiteteAntwortGehtDerDialogAuf(t *testing.T) {
	t.Parallel()

	if pfad, bereit := abnahmeDialog(); bereit || pfad != "" {
		t.Errorf("bereit %v, Pfad %q", bereit, pfad)
	}
}

// **Ohne Leitung bleibt die Zwischenablage die des Rechners** — hier im
// gewöhnlichen Bau. Ein Abnahmebau ohne `MSE_ABNAHME` startet gar nicht
// ([TestOhneLeitungStartetDerAbnahmebauNicht]).
func TestOhneLeitungGibtEsKeineErsatzablage(t *testing.T) {
	t.Parallel()

	if ablage := abnahmeAblage(); ablage != nil {
		t.Errorf("eine Ersatzablage ohne Leitung: %+v", ablage)
	}
}

func TestDieErsatzablageGibtZurueckWasHineinging(t *testing.T) {
	t.Parallel()

	var ablage ersatzablage
	if text := ablage.lesen(); text != "" {
		t.Errorf("eine frische Ablage trägt %q", text)
	}

	ablage.schreiben("zwei Karten")

	if text := ablage.lesen(); text != "zwei Karten" {
		t.Errorf("gelesen %q", text)
	}
}

// **Ohne Leitung ist ein Abnahmebau kein Werkzeug der Probe**, sondern ein
// Programm mit einem offenen MCP-Server und dem echten Bestand (N1 des
// Nachlaufs über Phase 6.3).
func TestOhneLeitungStartetDerAbnahmebauNicht(t *testing.T) {
	t.Parallel()

	if fehler := abnahmeGrund(""); fehler == nil {
		t.Error("ohne MSE_ABNAHME darf der Abnahmebau starten")
	}

	if fehler := abnahmeGrund("/tmp/huelle/abnahme.sock"); fehler != nil {
		t.Errorf("mit Leitung: %v", fehler)
	}
}

func TestDerSchluesselDesMcpServersKommtAusDerQuelle(t *testing.T) {
	t.Parallel()

	schluessel, fehler := mcpSchluessel(bytes.NewReader(bytes.Repeat([]byte{0xab}, 32)))
	if fehler != nil || schluessel != strings.Repeat("ab", 32) {
		t.Errorf("mcpSchluessel = %q, %v", schluessel, fehler)
	}

	if _, fehler := mcpSchluessel(bytes.NewReader([]byte{1, 2, 3})); fehler == nil {
		t.Error("ein Schlüssel aus drei Bytes wurde angenommen")
	}
}

// **Im gewöhnlichen Bau bereitet nichts etwas vor** — auch ohne
// `MSE_ABNAHME` startet er.
func TestImGewoehnlichenBauStartetEsOhneLeitung(t *testing.T) {
	t.Parallel()

	if fehler := AbnahmeVorbereiten(); fehler != nil {
		t.Errorf("AbnahmeVorbereiten = %v", fehler)
	}

	// **Und es schließt nichts an** — ohne Anwendung und Fenster, die der
	// gewöhnliche Bau gar nicht anfasst.
	abnahmeAnschliessen(nil, nil)
}

// **Ein Zug ist runter, gezogen, hoch** — A04
// (#2218):
// die Abfolge, die AppKit bei einem echten Ziehen liefert, ohne den
// Fensterserver.
func TestEinZugIstRunterGezogenHoch(t *testing.T) {
	t.Parallel()

	schritte, fehler := zugschritte([]Ort{{X: 10, Y: 20}, {X: 30, Y: 40}, {X: 50, Y: 60}}, false)
	if fehler != nil {
		t.Fatal(fehler)
	}

	erwartet := []mausschritt{
		{art: artMausRunter, ort: Ort{X: 10, Y: 20}},
		{art: artMausGezogen, ort: Ort{X: 30, Y: 40}},
		{art: artMausGezogen, ort: Ort{X: 50, Y: 60}},
		{art: artMausHoch, ort: Ort{X: 50, Y: 60}},
	}

	if !slices.Equal(schritte, erwartet) {
		t.Errorf("Schritte %+v, erwartet %+v", schritte, erwartet)
	}
}

// **Ein gehaltener Zug lässt nicht los** — für den Abbruch mit Escape: die
// Taste kommt, während die Maus noch unten ist.
func TestEinGehaltenerZugLaesstNichtLos(t *testing.T) {
	t.Parallel()

	schritte, fehler := zugschritte([]Ort{{X: 1, Y: 1}, {X: 2, Y: 2}}, true)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if letzter := schritte[len(schritte)-1]; letzter.art != artMausGezogen {
		t.Errorf("der gehaltene Zug endet mit %+v", letzter)
	}
}

func TestEinZugBrauchtZweiPunkte(t *testing.T) {
	t.Parallel()

	if _, fehler := zugschritte([]Ort{{X: 1, Y: 1}}, false); fehler == nil {
		t.Error("ein Zug aus einem Punkt wurde hingenommen")
	}
}

func TestEinRechtsklickIstRechtsRunterUndHoch(t *testing.T) {
	t.Parallel()

	schritte := rechtsklickschritte(Ort{X: 7, Y: 8})

	if len(schritte) != 2 || schritte[0].art != artRechtsRunter || schritte[1].art != artRechtsHoch ||
		schritte[0].ort != (Ort{X: 7, Y: 8}) {
		t.Errorf("Rechtsklick %+v", schritte)
	}
}

func TestEinZugbefehlWirdGelesen(t *testing.T) {
	t.Parallel()

	befehl, fehler := befehlLesen([]byte(`{"was":"ziehe","weg":[{"x":1,"y":2},{"x":3,"y":4}],"halten":true}`))
	if fehler != nil || befehl.Was != "ziehe" || len(befehl.Weg) != 2 || !befehl.Halten {
		t.Fatalf("%+v, %v", befehl, fehler)
	}
}

// **Eine Bewegung ohne Taste** — nach einem Loslassen außerhalb prüft sie,
// dass nichts mehr gefasst ist (A04, Zweitblick über Phase A).
func TestEineBewegungIstOhneTaste(t *testing.T) {
	t.Parallel()

	schritte := bewegeschritte(Ort{X: 3, Y: 4})

	if len(schritte) != 1 || schritte[0].art != artMausBewegt || schritte[0].ort != (Ort{X: 3, Y: 4}) {
		t.Errorf("Bewegung %+v", schritte)
	}
}
