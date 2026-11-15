package service_test

import (
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Stil sieht die Funktionen seines Spielpakets.**
//
// # Der Befund, aus dem das entstanden ist
//
// #186 hat dem
// Planbauer die Umgebung des **Stils** gegeben — 10 Bildebenen wurden 113. Was
// danach fehlte, kommt aus dem **Spielpaket**: `magic.mse-game` führt neben
// `game` die Datei `script`, und die Funktionen darin benutzt jeder Stil dieses
// Spiels. `additional_credit_brush_image` allein kam 458-mal vor
// (#195).
//
// # Was der Test führt
//
// Der Stil `probe-spielskript` der eigenen Saat rechnet seine Maße mit
// `randbreite()` und `feldbreite()` aus, und beide stehen im `init script:` des
// **Spiels**, nicht des Stils.
func TestEinStilSiehtDieFunktionenSeinesSpiels(t *testing.T) {
	vorlage := saatvorlage(t, "probe-spielskript.mse-style", "probe.mse-game")

	plan, befunde, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), planprobekarte(), vorlage)
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	if len(plan.Ebenen) < 2 {
		t.Fatalf("keine Feldebene entstanden, Befunde: %v", befunde)
	}

	// `randbreite()` gibt 24, `feldbreite()` 375 − 48 = 327.
	ebene := plan.Ebenen[1]
	if ebene.Rahmen.Links != 24 || ebene.Rahmen.Breite != 327 {
		t.Errorf("Rahmen %+v — erwartet links 24, Breite 327; Befunde: %v",
			ebene.Rahmen, befunde)
	}
}

// **Beide Schichten stehen nebeneinander — das Spiel und der Stil.**
//
// `randbreite()` kommt aus dem Spiel, `randhoehe()` aus dem `init script:` des
// Stils. Ein Plan, der nur eine der beiden sieht, ließe die andere Angabe auf 0
// fallen — und ein Feld bei 0 sieht aus wie eines, das jemand so gemeint hat.
func TestSpielUndStilStehenNebeneinander(t *testing.T) {
	vorlage := saatvorlage(t, "probe-spielskript.mse-style", "probe.mse-game")

	plan, befunde, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), planprobekarte(), vorlage)
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	if len(plan.Ebenen) < 2 {
		t.Fatalf("keine Feldebene, Befunde: %v", befunde)
	}

	if rahmen := plan.Ebenen[1].Rahmen; rahmen.Links != 24 || rahmen.Oben != 24 {
		t.Errorf("Rahmen %+v — links kommt aus dem Spiel, oben aus dem Stil; Befunde: %v",
			rahmen, befunde)
	}
}

// **Ohne Spielpaket bleibt es beim Stil allein — und das ist kein Fehler.**
//
// Ein Set, dessen Spielpaket auf diesem Rechner fehlt, bekommt die Felder, die
// ohne Spielfunktionen auskommen, und einen Befund für die anderen (§8).
func TestOhneSpielskriptWirdBenanntStattAbgebrochen(t *testing.T) {
	vorlage := service.Vorlage{Stil: gelesenerStil(t, "probe-spielskript.mse-style")}

	plan, befunde, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), planprobekarte(), vorlage)
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	if plan == nil {
		t.Fatal("kein Plan")
	}

	if len(befunde) == 0 {
		t.Error("kein Befund — niemand sieht, warum das Feld fehlt")
	}
}

// saatvorlage liest Stil und Spiel der eigenen Saat.
func saatvorlage(t *testing.T, stilpaket, spielpaket string) service.Vorlage {
	t.Helper()

	return service.Vorlage{
		Stil:        gelesenerStil(t, stilpaket),
		Spielskript: gelesenesSpielskript(t, spielpaket),
	}
}

// gelesenerStil öffnet ein Stilpaket der Saat samt seinen Modulen.
func gelesenerStil(t *testing.T, paketname string) msepaket.Stil {
	t.Helper()

	paket := geoeffnet(t, paketname)
	paket.ModuleEinbinden(msepaket.NeuerPaketleser(filepath.Join("..", "..", "saat")))

	stil, befunde := paket.Stilmodell()
	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	return stil
}

// gelesenesSpielskript öffnet ein Spielpaket und gibt sein Initskript.
func gelesenesSpielskript(t *testing.T, paketname string) string {
	t.Helper()

	paket := geoeffnet(t, paketname)
	paket.ModuleEinbinden(msepaket.NeuerPaketleser(filepath.Join("..", "..", "saat")))

	return paket.Initskript()
}

// geoeffnet öffnet ein Paket der Saat und schließt es zum Schluss.
func geoeffnet(t *testing.T, paketname string) *msepaket.Paket {
	t.Helper()

	paket, err := msepaket.Oeffne(filepath.Join("..", "..", "saat", paketname))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	return paket
}
