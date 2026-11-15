package service_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// musterablage beantwortet die Frage nach einem `match:`-Muster.
type musterablage struct {
	gesehen string
	pakete  []domain.Paket
	fehler  error
}

func (m *musterablage) PaketeNachMuster(
	_ context.Context, muster string,
) ([]domain.Paket, error) {
	m.gesehen = muster

	return m.pakete, m.fehler
}

// **Die Wahl einer `package choice` sind die passenden Pakete**
// (#632).
//
// **ported.** Gespeichert wird der Dateiname samt Endung — `relativeFilename()`
// —, angezeigt der Kurzname mit großem Anfangsbuchstaben —
// `capitalize_sentence(short_name)` (`src/render/value/package_choice.cpp`,
// Zeilen 28–43, `src/util/string.cpp`, Zeilen 115–119, Fassung f9d9356 vom
// 2021-07-06).
func TestEinePaketwahlZeigtDieKurznamenUndSpeichertDieDateinamen(t *testing.T) {
	ablage := &musterablage{pakete: []domain.Paket{
		{Art: "mse-include", Name: "probe-overlay-folie", Kurzname: "folie glänzend"},
		{Art: "mse-include", Name: "probe-overlay-matt"},
	}}

	wahlen, err := service.Paketwahlen(t.Context(), ablage, domain.StyleField{
		Kind: domain.KindPackageChoice, Match: "probe-overlay-*.mse-include", Required: true,
	})
	if err != nil {
		t.Fatalf("wahlen: %v", err)
	}

	if ablage.gesehen != "probe-overlay-*.mse-include" {
		t.Errorf("gefragt wurde nach %q", ablage.gesehen)
	}

	erwartet := []domain.Wahlmoeglichkeit{
		{Wert: "probe-overlay-folie.mse-include", Beschriftung: "Folie glänzend"},
		// **Ohne `short name:` gilt der Dateiname ohne Endung** — MSEs
		// Rückfall in `Packaged::validate`.
		{Wert: "probe-overlay-matt.mse-include", Beschriftung: "Probe-overlay-matt"},
	}

	vergleicheWahlen(t, wahlen, erwartet)
}

// **Ist die Wahl nicht verlangt, steht der leere Eintrag vorn.**
//
// `itemText(0)` gibt `empty_name`, wenn `!required` — und `select(0)` setzt
// den leeren Wert (`src/gui/value/package_choice.cpp`, Zeilen 44–59, dieselbe
// Fassung).
func TestOhnePflichtStehtDerLeereEintragVorn(t *testing.T) {
	ablage := &musterablage{pakete: []domain.Paket{
		{Art: "mse-include", Name: "probe-overlay-folie", Kurzname: "folie"},
	}}

	wahlen, err := service.Paketwahlen(t.Context(), ablage, domain.StyleField{
		Kind: domain.KindPackageChoice, Match: "probe-overlay-*.mse-include",
		Required: false, EmptyName: "keines",
	})
	if err != nil {
		t.Fatalf("wahlen: %v", err)
	}

	vergleicheWahlen(t, wahlen, []domain.Wahlmoeglichkeit{
		{Wert: "", Beschriftung: "Keines"},
		{Wert: "probe-overlay-folie.mse-include", Beschriftung: "Folie"},
	})
}

// **Ein Fehler des Katalogs geht durch** — die Liste erfindet keine Pakete.
func TestEinFehlerDesKatalogsGehtDurch(t *testing.T) {
	ablage := &musterablage{fehler: domain.ErrNotFound}

	_, err := service.Paketwahlen(t.Context(), ablage, domain.StyleField{
		Kind: domain.KindPackageChoice, Match: "probe-*.mse-include", Required: true,
	})
	if err == nil {
		t.Error("kein Fehler")
	}
}

// **Ein leerer Name bleibt leer** — `grossAmAnfang` fasst nichts an, wo nichts
// steht. Das trifft eine `package choice` mit `empty name:` ohne Wert.
func TestEinLeererNameBleibtLeer(t *testing.T) {
	ablage := &musterablage{}

	wahlen, err := service.Paketwahlen(t.Context(), ablage, domain.StyleField{
		Kind: domain.KindPackageChoice, Match: "probe-*.mse-include",
		Required: false, EmptyName: "",
	})
	if err != nil {
		t.Fatalf("wahlen: %v", err)
	}

	vergleicheWahlen(t, wahlen, []domain.Wahlmoeglichkeit{{Wert: "", Beschriftung: ""}})
}

// **Eine Art ohne Auswahl bekommt keine Liste** — und der Katalog wird nicht
// gefragt.
//
// `text`, `info`, `slider` und `image`: für die ist ein Textfeld die Antwort,
// und eine leere Liste sagt dem Fenster genau das.
func TestEineArtOhneAuswahlBekommtKeineListe(t *testing.T) {
	katalog := &ablage{pakete: []domain.Paket{
		{Art: "mse-game", Name: "probe", Pfad: "/p/probe.mse-game"},
		{
			Art: "mse-style", Name: "probe-stil", Version: "2026-08-15",
			Pfad: "/p/probe-stil.mse-style",
		},
	}}

	waehle := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{
			{ID: 7, Game: "probe", Style: "probe-stil"},
		}},
		katalog, &stilablage{}, &feldliste{}, testdatenwurzel(t),
	)

	optionen, err := waehle.Optionen(t.Context(), 7)
	if err != nil {
		t.Fatalf("Optionen: %v", err)
	}

	gesehen := 0

	for _, option := range optionen {
		if option.Kind != domain.KindText {
			continue
		}

		gesehen++

		if len(option.Moeglichkeiten) != 0 {
			t.Errorf("%s: %+v, erwartet keine Möglichkeiten",
				option.Name, option.Moeglichkeiten)
		}
	}

	// **Ohne diese Zeile prüfte der Test nichts**: eine Schleife über keine
	// Textoption ist grün, ohne ihren Fall zu enthalten (§10).
	if gesehen == 0 {
		t.Fatal("keine Textoption im Prüfstück")
	}
}

// **Eine Farbwahl trägt Namen und Farben** — und fragt den Katalog nicht
// (#634).
//
// **ported.** MSE zeigt zu jeder `choice:` ihren Namen mit einer Farbkachel
// (`src/gui/value/color.cpp`, Zeilen 60–72, Fassung f9d9356 vom 2021-07-06);
// gespeichert wird die Farbe in der Form, die `format_color` erzeugt.
func TestEineFarbwahlTraegtNamenUndFarben(t *testing.T) {
	wahlen := service.Farbwahlen(domain.StyleField{
		Kind:       domain.KindColor,
		Choices:    []string{"schwarz", "weiß"},
		Wahlfarben: []string{"rgb(0,0,0)", "rgb(255,255,255)"},
	})

	vergleicheWahlen(t, wahlen, []domain.Wahlmoeglichkeit{
		{Wert: "rgb(0,0,0)", Beschriftung: "schwarz"},
		{Wert: "rgb(255,255,255)", Beschriftung: "weiß"},
	})
}

// **Eine Farbe ohne Wahlen gibt keine Liste** — und das ist kein Mangel: 16
// der 105 Farbstiloptionen des Packs nennen keine, und für die ist der
// Farbwähler das Einzige, was es gibt (#630).
func TestEineFarbeOhneWahlenGibtKeineListe(t *testing.T) {
	if wahlen := service.Farbwahlen(domain.StyleField{Kind: domain.KindColor}); len(wahlen) != 0 {
		t.Errorf("%+v, erwartet keine", wahlen)
	}
}

// **Ein Name ohne Farbe daneben bricht die Liste ab.**
//
// So kommt eine Vorlage an, in der ein `choice:` keine `color:` trägt.
// Weiterzuzählen hieße, den nächsten Namen mit der nächsten Farbe zu paaren —
// und dann stünde „schwarz" auf Weiß.
func TestEinNameOhneFarbeBrichtDieListeAb(t *testing.T) {
	wahlen := service.Farbwahlen(domain.StyleField{
		Kind:       domain.KindColor,
		Choices:    []string{"schwarz", "ohnefarbe"},
		Wahlfarben: []string{"rgb(0,0,0)"},
	})

	vergleicheWahlen(t, wahlen, []domain.Wahlmoeglichkeit{
		{Wert: "rgb(0,0,0)", Beschriftung: "schwarz"},
	})
}

// vergleicheWahlen hält zwei Listen gegeneinander.
func vergleicheWahlen(t *testing.T, gefunden, erwartet []domain.Wahlmoeglichkeit) {
	t.Helper()

	if len(gefunden) != len(erwartet) {
		t.Fatalf("%+v, erwartet %+v", gefunden, erwartet)
	}

	for i := range erwartet {
		if gefunden[i] != erwartet[i] {
			t.Errorf("Stelle %d: %+v, erwartet %+v", i, gefunden[i], erwartet[i])
		}
	}
}

// **Und `Optionen` reicht die Liste durch** — die Naht, nicht nur das Stück.
//
// Ein Test gegen `Paketwahlen` allein sagte, dass die Liste richtig gebaut
// wird; ob sie **ankommt**, entscheidet der Weg durch den Stilwähler (§10).
func TestOptionenTragenDiePaketwahlen(t *testing.T) {
	katalog := &ablage{pakete: []domain.Paket{
		{Art: "mse-game", Name: "probe", Pfad: "/p/probe.mse-game"},
		{
			Art: "mse-style", Name: "probe-stil", Version: "2026-08-15",
			Pfad:    "/p/probe-stil.mse-style",
			Braucht: []domain.Bedarf{{Paket: "probe.mse-game", Version: "2026-08-15"}},
		},
		{Art: "mse-symbol-font", Name: "probe-mana-rund", Kurzname: "runde Symbole", Rang: 2},
		{Art: "mse-symbol-font", Name: "probe-mana-eckig", Kurzname: "eckige Symbole", Rang: 1},
		{Art: "mse-include", Name: "probe-overlay-folie", Kurzname: "folie"},
	}}

	waehle := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{
			{ID: 7, Game: "probe", Style: "probe-stil"},
		}},
		katalog, &stilablage{}, &feldliste{}, testdatenwurzel(t),
	)

	optionen, err := waehle.Optionen(t.Context(), 7)
	if err != nil {
		t.Fatalf("Optionen: %v", err)
	}

	gefunden := map[string][]domain.Wahlmoeglichkeit{}
	for _, option := range optionen {
		gefunden[option.Name] = option.Moeglichkeiten
	}

	// **Rang zuerst** — `eckig` steht vor `rund`, obwohl der Name anders will.
	vergleicheWahlen(t, gefunden["mana symbols"], []domain.Wahlmoeglichkeit{
		{Wert: "probe-mana-eckig.mse-symbol-font", Beschriftung: "Eckige Symbole"},
		{Wert: "probe-mana-rund.mse-symbol-font", Beschriftung: "Runde Symbole"},
	})

	// **`required: false`** — der leere Eintrag steht vorn und heißt `keines`.
	vergleicheWahlen(t, gefunden["overlay"], []domain.Wahlmoeglichkeit{
		{Wert: "", Beschriftung: "Keines"},
		{Wert: "probe-overlay-folie.mse-include", Beschriftung: "Folie"},
	})

	// **Eine Wahl ohne passende Pakete bleibt leer bis auf den leeren
	// Eintrag** — sie steht trotzdem da, statt zu verschwinden.
	vergleicheWahlen(t, gefunden["pt symbols"], []domain.Wahlmoeglichkeit{
		{Wert: "", Beschriftung: "None"},
	})

	// **Und eine Farboption geht denselben Weg** (#634): Name und Farbe, ohne
	// dass der Katalog gefragt würde.
	vergleicheWahlen(t, gefunden["randfarbe"], []domain.Wahlmoeglichkeit{
		{Wert: "rgb(0,0,0)", Beschriftung: "schwarz"},
		{Wert: "rgb(255,255,255)", Beschriftung: "weiß"},
	})

	// **Gefragt wurde nur nach den drei Paketmustern** — nicht nach den
	// Farben, nicht nach den übrigen Arten.
	if len(katalog.musterfragen) != 3 {
		t.Errorf("gefragt wurde nach %v, erwartet drei Muster", katalog.musterfragen)
	}
}

// testdatenwurzel gibt das Prüfstückverzeichnis als Paketwurzel.
func testdatenwurzel(t *testing.T) string {
	t.Helper()

	wurzel, err := filepath.Abs(filepath.Join("..", "..", "testdaten"))
	if err != nil {
		t.Fatalf("Wurzel: %v", err)
	}

	return wurzel
}

// **Und ein Fehler des Katalogs macht `Optionen` zum Fehler**, nicht zu einer
// Liste ohne Wahlen.
//
// Eine `package choice` ohne Möglichkeiten sähe aus wie eine Vorlage, die
// keine anbietet — und die Ursache stünde nirgends.
func TestEinFehlerDesKatalogsMachtOptionenZumFehler(t *testing.T) {
	katalog := &ablage{
		pakete: []domain.Paket{
			{Art: "mse-game", Name: "probe", Pfad: "/p/probe.mse-game"},
			{
				Art: "mse-style", Name: "probe-stil", Version: "2026-08-15",
				Pfad: "/p/probe-stil.mse-style",
			},
		},
		fehler: domain.ErrNotFound,
	}

	waehle := service.NeuerStilwaehler(
		&koepfequelle{koepfe: []domain.SetHead{
			{ID: 7, Game: "probe", Style: "probe-stil"},
		}},
		katalog, &stilablage{}, &feldliste{}, testdatenwurzel(t),
	)

	if _, err := waehle.Optionen(t.Context(), 7); err == nil {
		t.Error("kein Fehler")
	}
}
