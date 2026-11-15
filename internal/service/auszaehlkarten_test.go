package service_test

// **Welche Karten ein Balken trägt** — UXR-F6
// (#2013).
//
// MSE2 behält zu jedem Element der Auszählung die Karte, aus der es kommt
// (`GraphElement::original_index`, `StatsPanel::showCategory` in
// `src/gui/set/stats_panel.cpp`, Zeilen 477–507, Fassung f9d9356 vom
// 2021-07-06), und die Wahl einer Gruppe gibt diese Karten zurück
// (`StatsFilter`, Zeilen 536–545). Hier ist die Zuordnung derselbe Schritt,
// der die Werte rechnet: ein zweiter Durchlauf über die Skripte hieße, 3.761
// Karten ein zweites Mal 25 Sekunden zu rechnen, bloß um zu erfahren, welche
// im Balken standen.

import (
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// farbwert ist die Dimension der Proben: die Farbe der Karte.
var farbwert = service.Auszaehlangabe{Name: "farbe", Skript: "card.farbe"}

// **Jeder Wert nennt die Karte, aus der er kommt.**
func TestJederWertNenntSeineKarte(t *testing.T) {
	t.Parallel()

	werte, befunde := service.AuszaehlwerteKarten(t.Context(),
		auszaehlkarten("weiss", "rot", "weiss"), auszaehlprobe(""), nil, farbwert)

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	erwartet := []domain.Dimensionswert{
		{Karte: 1, Wert: "weiss"}, {Karte: 2, Wert: "rot"}, {Karte: 3, Wert: "weiss"},
	}

	if !slices.Equal(werte, erwartet) {
		t.Errorf("die Werte lauten %+v", werte)
	}
}

// **Eine Karte ohne Wert hat keinen — und steht in keiner Gruppe.**
func TestEineKarteOhneWertStehtInKeinerGruppe(t *testing.T) {
	t.Parallel()

	werte, _ := service.AuszaehlwerteKarten(t.Context(),
		auszaehlkarten("rot", ""), auszaehlprobe(""), nil, farbwert)

	if !slices.Equal(werte, []domain.Dimensionswert{{Karte: 1, Wert: "rot"}}) {
		t.Errorf("die Werte lauten %+v", werte)
	}
}

// **`split list:` gibt jedem Glied dieselbe Karte** — eine zweifarbige Karte
// steht in beiden Gruppen (`GraphDataPre::splitList` kopiert das Element samt
// `original_index`).
func TestEinGliedDerListeNenntDieKarteDesGanzen(t *testing.T) {
	t.Parallel()

	zerlegt := farbwert
	zerlegt.Zerlegen = true

	werte, _ := service.AuszaehlwerteKarten(t.Context(),
		auszaehlkarten("weiss, blau"), auszaehlprobe(""), nil, zerlegt)

	erwartet := []domain.Dimensionswert{
		{Karte: 1, Wert: "weiss"}, {Karte: 1, Wert: "blau"},
	}

	if !slices.Equal(werte, erwartet) {
		t.Errorf("die Werte lauten %+v", werte)
	}
}

// **Die Werte ohne Karten sind dieselben wie vorher** — die Fälle darüber
// gelten für `Auszaehlwerte` weiter, und dieser hier hält fest, dass es
// dieselbe Rechnung ist.
func TestDieWerteOhneKartenSindDieselbenWieMitKarten(t *testing.T) {
	t.Parallel()

	karten := auszaehlkarten("weiss", "", "rot")

	ohne, _ := service.Auszaehlwerte(t.Context(), karten, auszaehlprobe(""), nil, farbwert)
	mit, _ := service.AuszaehlwerteKarten(t.Context(), karten, auszaehlprobe(""), nil, farbwert)

	if len(ohne) != len(mit) {
		t.Fatalf("%d Werte ohne Karten, %d mit", len(ohne), len(mit))
	}

	for stelle := range ohne {
		if ohne[stelle] != mit[stelle].Wert {
			t.Errorf("Wert %d: %q gegen %q", stelle, ohne[stelle], mit[stelle].Wert)
		}
	}
}

// **Die Auszählung nennt die Karten je Gruppe.**
func TestEineAuszaehlungNenntDieKartenJeGruppe(t *testing.T) {
	t.Parallel()

	auszaehlung, fehler := service.NeuerAuszaehldienst(probedimension()).
		Auszaehlen(t.Context(), 1, "farbe")
	if fehler != nil {
		t.Fatalf("Auszaehlen: %v", fehler)
	}

	if len(auszaehlung.Gruppen) != 2 {
		t.Fatalf("%d Gruppen: %+v", len(auszaehlung.Gruppen), auszaehlung.Gruppen)
	}

	// weiss, rot, rot — die Karten 1, 2 und 3.
	if !slices.Equal(auszaehlung.Gruppen[0].Kartenkennungen, []int64{1}) {
		t.Errorf("weiss trägt %v", auszaehlung.Gruppen[0].Kartenkennungen)
	}

	if !slices.Equal(auszaehlung.Gruppen[1].Kartenkennungen, []int64{2, 3}) {
		t.Errorf("rot trägt %v", auszaehlung.Gruppen[1].Kartenkennungen)
	}
}

// **Die Karte, deren Skript scheitert, steht in keiner Gruppe.**
//
// MSE nimmt sie heraus (`show = false`); die Liste unter dem Balken zeigte
// sonst eine Karte, die der Balken nicht gezählt hat.
func TestEineGescheiterteKarteStehtInKeinerGruppe(t *testing.T) {
	t.Parallel()

	probe := probedimension()
	probe.angaben[0].Skript = `if card.farbe == "rot" then fehlt(card) else card.farbe`

	auszaehlung, fehler := service.NeuerAuszaehldienst(probe).
		Auszaehlen(t.Context(), 1, "farbe")
	if fehler != nil {
		t.Fatalf("Auszaehlen: %v", fehler)
	}

	if len(auszaehlung.Befunde) != 1 {
		t.Fatalf("die Befunde lauten %+v", auszaehlung.Befunde)
	}

	if !slices.Equal(auszaehlung.Gruppen[0].Kartenkennungen, []int64{1}) ||
		len(auszaehlung.Gruppen[1].Kartenkennungen) != 0 {
		t.Errorf("die Gruppen lauten %+v", auszaehlung.Gruppen)
	}
}
