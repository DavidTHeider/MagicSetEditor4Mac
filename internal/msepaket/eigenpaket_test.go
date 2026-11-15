package msepaket_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein `include file:` ohne führenden Schrägstrich meint das eigene Paket.**
//
// # Wie das erschlossen wurde
//
// Ausgezählt über die 327 Pakete von `M15-Magic-Pack`: von 161 verschiedenen
// `include file:`-Angaben tragen **134** einen führenden Schrägstrich und
// nennen damit ein fremdes Paket; **27** tun es nicht. Sie stehen in Stilen
// (`magic-m15-invocation.mse-style` → `invocation`), in Einschüben
// (`magic-blends.mse-include` → `blends_scripts`), in einer Exportvorlage und
// im **Spielpaket** (`magic.mse-game` → `script`, `game_script`, `card_fields`).
//
// In jedem nachgesehenen Fall liegt eine Datei dieses Namens im Paket selbst.
//
// # Warum das mehr ist als ein Sonderfall
//
// `magic.mse-game` bindet **alle** seine Module so ein. Ohne diese Auflösung
// gibt es kein einziges Spielskript — und damit fehlen jedem Stil dieses Spiels
// die Funktionen, mit denen er rechnet
// (#195).
func TestEinVerweisOhneSchraegstrichMeintDasEigenePaket(t *testing.T) {
	paket, err := msepaket.Oeffne("../../saat/probe.mse-game")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if befunde := paket.ModuleEinbinden(
		msepaket.NeuerPaketleser("../../saat")); len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	skript := paket.Initskript()
	if !strings.Contains(skript, "randbreite") {
		t.Errorf("das Initskript trägt das Modul nicht: %q", skript)
	}
}

// **Und ein Ausbruch bleibt ein Ausbruch.**
//
// Ein eigener Verweis darf tiefer greifen (`watermarks/watermarks` steht so in
// `magic-m15-extra.mse-style`), aber nicht hinaus. Ohne diese Zusage öffnete
// die Auflösung des eigenen Pakets genau die Tür, die der Zweitblick über
// Phase 3 an `ZerlegePaketpfad` geschlossen hat.
func TestEinEigenerVerweisBrichtNichtAus(t *testing.T) {
	// **Und kein gültiger Text ist kein Weg** — `fs.ValidPath` weist ihn ab;
	// gefunden von `FuzzPaketpfad` (QA-06b).
	for _, verweis := range []string{"../fremd", "unter/../../fremd", "./hier", "\xff"} {
		t.Run(verweis, func(t *testing.T) {
			if _, err := msepaket.ZerlegeEigenpfad("probe.mse-game", verweis); err == nil {
				t.Errorf("%q wurde angenommen", verweis)
			}
		})
	}
}

// **Ein eigener Verweis mit Unterverzeichnis geht durch.**
func TestEinEigenerVerweisDarfTieferGreifen(t *testing.T) {
	pfad, err := msepaket.ZerlegeEigenpfad("magic-m15-extra.mse-style", "watermarks/watermarks")
	if err != nil {
		t.Fatalf("zerlegen: %v", err)
	}

	if pfad.Paket != "magic-m15-extra.mse-style" || pfad.Datei != "watermarks/watermarks" {
		t.Errorf("%+v", pfad)
	}
}

// **Ein `include file:` darf auch der *Wert* einer Zeile sein.**
//
// # Woran das aufgefallen ist
//
// `magic.mse-game` schreibt seinen Einstieg in **einer** Zeile:
//
//	init script: include file: script
//
// Der Einbinder sah bis dahin nur `include file:` als eigenen **Schlüssel** —
// hier ist es der Wert von `init script:`. Nachgemessen: das Initskript des
// Spielpakets war danach 74 Zeichen lang und bestand aus einer Leerzeile und
// einem Kommentar. `additional_credit_brush_image` kam darin nicht vor, obwohl
// die Funktion in `magic.mse-game/script` steht — 458-mal gerufen
// (#195).
func TestEinVerweisDarfDerWertEinerZeileSein(t *testing.T) {
	paket, err := msepaket.Oeffne("../../saat/probe-einzeiler.mse-style")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if befunde := paket.ModuleEinbinden(
		msepaket.NeuerPaketleser("../../saat")); len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if skript := paket.Initskript(); !strings.Contains(skript, "einzeilig") {
		t.Errorf("das Initskript trägt das Modul nicht: %q", skript)
	}
}

// **Ein Paket ohne `init script:` gibt eine leere Zeichenkette.**
//
// Das ist der Normalfall: die meisten Stilpakete führen keines, und ein
// Fehler dafür machte aus einer Abwesenheit eine Störung.
func TestEinPaketOhneInitskript(t *testing.T) {
	paket, err := msepaket.Oeffne("../../saat/probe-schlicht.mse-style")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if skript := paket.Initskript(); skript != "" {
		t.Errorf("%q — dieses Paket führt kein Initskript", skript)
	}
}

// **Ein leerer eigener Verweis ist keiner.**
//
// `include file:` ohne Angabe: ohne diese Abweisung entstünde ein Pfad auf das
// Paketverzeichnis selbst.
func TestEinLeererEigenerVerweis(t *testing.T) {
	if _, err := msepaket.ZerlegeEigenpfad("probe.mse-game", "   "); err == nil {
		t.Error("ein leerer Verweis wurde angenommen")
	}
}
