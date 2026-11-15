package service

// **`keyword_usage` — welche Schlüsselwörter eine Karte benutzt** —
// #1354.
//
// # Was MSE2 tut
//
// `SCRIPT_FUNCTION(keyword_usage)` (`src/script/functions/basic.cpp`, Zeilen
// 695–717, Fassung `f9d9356` vom 2021-07-06) läuft über `card->keyword_usage`
// und fügt die Namen mit `", "` zusammen. Mit `unique: true` fällt weg, was
// schon einmal dastand — verglichen wird der **Zeiger** auf das
// Schlüsselwort, nicht sein Name.
//
// **Gefüllt wird die Liste als Nebenwirkung von `expand_keywords`**:
// `expand_keyword` trägt `(feld, schlüsselwort)` ein, **nachdem** der
// Erinnerungstext zusammengesetzt ist (`src/data/keyword.cpp` Zeilen 749–751).
// Eine Bedingung, die nicht zutrifft, kehrt vorher um (Zeilen 717–719) und
// trägt **nichts** ein.
//
// **Und vor jedem Feld wird sein alter Eintrag entfernt** — `remove_from_stats`
// (Zeilen 766–773). Ohne das wüchse die Liste bei jedem Neuzeichnen.

import (
	"testing"
)

// TestEineVerwendungWirdVermerkt ist die Grundzusage.
func TestEineVerwendungWirdVermerkt(t *testing.T) {
	t.Parallel()

	vermerk := &Schluesselwortverwendung{}
	vermerk.Vermerke("rule text", "Flying")
	vermerk.Vermerke("rule text", "Trample")

	if gefunden := vermerk.Liste(false); gefunden != "Flying, Trample" {
		t.Errorf("%q", gefunden)
	}
}

// **`unique: true` lässt eine Wiederholung weg** — MSEs innere Schleife.
func TestEineWiederholungFaelltMitUniqueWeg(t *testing.T) {
	t.Parallel()

	vermerk := &Schluesselwortverwendung{}
	vermerk.Vermerke("rule text", "Flying")
	vermerk.Vermerke("rule text", "Flying")
	vermerk.Vermerke("rule text 2", "Trample")

	if gefunden := vermerk.Liste(true); gefunden != "Flying, Trample" {
		t.Errorf("mit unique: %q", gefunden)
	}

	// **Ohne `unique:` steht sie zweimal da** — das ist die Zusage, die den
	// Unterschied zwischen beiden Formen prüft.
	if gefunden := vermerk.Liste(false); gefunden != "Flying, Flying, Trample" {
		t.Errorf("ohne unique: %q", gefunden)
	}
}

// **Die Reihenfolge ist die des Eintragens**, über die Felder hinweg — MSE
// läuft über den Vektor, nicht über eine Menge.
func TestDieReihenfolgeIstDieDesEintragens(t *testing.T) {
	t.Parallel()

	vermerk := &Schluesselwortverwendung{}
	vermerk.Vermerke("b", "Zweites")
	vermerk.Vermerke("a", "Erstes")

	if gefunden := vermerk.Liste(false); gefunden != "Zweites, Erstes" {
		t.Errorf("%q", gefunden)
	}
}

// **Ein Feld wird vor dem Einsetzen geräumt** — `remove_from_stats`.
//
// Ohne das wüchse die Liste bei jedem Zeichnen: dieselbe Karte, zweimal
// gerechnet, gäbe „Flying, Flying".
func TestEinFeldWirdVorDemEinsetzenGeraeumt(t *testing.T) {
	t.Parallel()

	vermerk := &Schluesselwortverwendung{}
	vermerk.Vermerke("rule text", "Flying")
	vermerk.Vermerke("rule text 2", "Trample")

	vermerk.Raeume("rule text")
	vermerk.Vermerke("rule text", "Flying")

	if gefunden := vermerk.Liste(false); gefunden != "Trample, Flying" {
		t.Errorf("%q", gefunden)
	}
}

// **Und nur das genannte Feld** — die anderen bleiben stehen.
func TestNurDasGenannteFeldWirdGeraeumt(t *testing.T) {
	t.Parallel()

	vermerk := &Schluesselwortverwendung{}
	vermerk.Vermerke("a", "Eins")
	vermerk.Vermerke("b", "Zwei")

	vermerk.Raeume("a")

	if gefunden := vermerk.Liste(false); gefunden != "Zwei" {
		t.Errorf("%q", gefunden)
	}
}

// **Ohne Verwendung ist es leer** — und das ist ein Wert, kein Fehler.
func TestOhneVerwendungIstDieListeLeer(t *testing.T) {
	t.Parallel()

	if gefunden := (&Schluesselwortverwendung{}).Liste(true); gefunden != "" {
		t.Errorf("%q, erwartet leer", gefunden)
	}
}

// **Ein Vermerk, den es nicht gibt, antwortet leer.**
//
// Der Nullzeiger entsteht, wo eine Umgebung ohne Karte gebaut wird — der
// Rasterer tut das für die Attrappe. Ein Absturz dort wäre ein Absturz beim
// Öffnen eines Sets.
func TestEinFehlenderVermerkAntwortetLeer(t *testing.T) {
	t.Parallel()

	var keiner *Schluesselwortverwendung

	if gefunden := keiner.Liste(true); gefunden != "" {
		t.Errorf("%q, erwartet leer", gefunden)
	}
}
