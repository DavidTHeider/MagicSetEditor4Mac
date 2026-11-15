package msepaket_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// **Ein Stil bringt sein `init script:` mit** — als Text, nicht als Baum.
//
// # Warum als Text
//
// Es ist ein **Programm**, kein `schlüssel: wert`-Baum
// ([0014](../../docs/entscheidungen/0014-eine-mse-datei-ist-nicht-notwendig-ein-baum.md)).
// Wer es zerteilt herausgibt, zwingt den Aufrufer, es wieder zusammenzusetzen —
// und `msescript` will ohnehin Quelltext.
//
// Ohne diese Angabe ließ sich von 356 `image:`-Angaben in `M15-Magic-Pack`
// **keine einzige** ausrechnen: sie rufen `module_card()` und ihresgleichen,
// und die stehen hier
// (#186).
func TestEinStilBringtSeinInitskriptMit(t *testing.T) {
	stil := stilmodell(t)

	if !strings.Contains(stil.Initskript, `grundfarbe := "rot"`) {
		t.Errorf("Initskript ohne Zuweisung:\n%s", stil.Initskript)
	}

	// **Ohne Einrückung.** Sie gehört dem Paketformat, nicht dem Skript; wer
	// sie stehen lässt, reicht `msescript` Tabulatoren, die dort nichts
	// bedeuten.
	for zeile := range strings.SplitSeq(stil.Initskript, "\n") {
		if strings.HasPrefix(zeile, "\t") {
			t.Errorf("eingerückte Zeile: %q", zeile)
		}
	}
}

// **`include file:` wird aufgelöst, bevor der Stil gelesen wird.**
//
// Das Modul steht danach im `init script:`, als wäre es dort geschrieben — und
// genau so benutzt der Stil es: `probe_hintergrund` ruft `modul_hintergrund`,
// das erst das Modul anlegt.
func TestEinModulLandetImInitskript(t *testing.T) {
	paket := stilpaket(t)

	if befunde := paket.ModuleEinbinden(testquelle{}); len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	stil, befunde := paket.Stilmodell()
	if len(befunde) > 0 {
		t.Fatalf("Stilbefunde: %v", befunde)
	}

	if !strings.Contains(stil.Initskript, "modul_hintergrund :=") {
		t.Errorf("das Modul fehlt:\n%s", stil.Initskript)
	}

	// Und der Verweis selbst ist fort — sonst stünde er als Anweisung im
	// Skript und `msescript` scheiterte an ihm.
	if strings.Contains(stil.Initskript, "include file:") {
		t.Errorf("der Verweis steht noch da:\n%s", stil.Initskript)
	}
}

// **Ein Modul, das es nicht gibt, ist ein Befund und kein Abbruch** (§8).
func TestEinFehlendesModulIstEinBefund(t *testing.T) {
	paket := stilpaket(t)

	befunde := paket.ModuleEinbinden(leerequelle{})
	if len(befunde) == 0 {
		t.Fatal("kein Befund")
	}

	// Und der Stil ist trotzdem lesbar.
	if stil, _ := paket.Stilmodell(); len(stil.Felder) == 0 {
		t.Error("kein Feldstil mehr")
	}
}

// stilpaket öffnet das Prüfstück frisch — jeder Test bekommt ein eigenes,
// weil `ModuleEinbinden` den Kopf verändert.
func stilpaket(t *testing.T) *msepaket.Paket {
	t.Helper()

	paket, err := msepaket.Oeffne(filepath.Join(testdaten, "probe-stil.mse-style"))
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

// testquelle liest Module aus `testdaten/`.
type testquelle struct{}

func (testquelle) Lies(pfad msepaket.Paketpfad) ([]byte, error) {
	return msepaket.NeuerPaketleser(testdaten).Lies(pfad)
}

// leerequelle kennt kein Paket.
type leerequelle struct{}

func (leerequelle) Lies(msepaket.Paketpfad) ([]byte, error) {
	return nil, msepaket.ErrPaketpfad
}

// **Der Paketleser sucht die Wurzeln der Reihe nach und meldet, was fehlt.**
//
// Dieselbe Zusage wie bei [msepaket.Bildquelle] und aus demselben Grund: die
// Saat liegt im Programm, bezogene Pakete liegen bei den Nutzenden.
func TestDerPaketleserSuchtUndMeldet(t *testing.T) {
	leser := msepaket.NeuerPaketleser("", testdaten)

	inhalt, err := leser.Lies(msepaket.Paketpfad{
		Paket: "probe-modul.mse-include", Datei: "scripts",
	})
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if !strings.Contains(string(inhalt), "modul_hintergrund") {
		t.Errorf("Inhalt %q", inhalt)
	}

	// **Und eine Datei, die es nicht gibt, wird benannt** — mit dem Paketpfad,
	// nicht mit dem Dateisystempfad, den nur diese Quelle kennt.
	_, fehlt := leser.Lies(msepaket.Paketpfad{Paket: "gibtsnicht.mse-include", Datei: "x"})
	if !errors.Is(fehlt, msepaket.ErrKeinPaket) {
		t.Errorf("Fehler %v, erwartet ErrKeinPaket", fehlt)
	}

	if !strings.Contains(fehlt.Error(), "/gibtsnicht.mse-include/x") {
		t.Errorf("die Meldung nennt den Verweis nicht: %v", fehlt)
	}
}

// **Ein Knoten gibt seine Zeile — zusammengesetzt, wenn die rohe fort ist.**
//
// Der Schreiber erkennt an einer fehlenden Rohzeile, dass er neu setzen muss;
// [msepaket.Knoten.Skripttext] muss dieselbe Stelle bedienen, sonst verlöre ein
// geänderter Eintrag im Initskript seine Zeile.
func TestEineZeileWirdZusammengesetztWennDieRoheFehlt(t *testing.T) {
	// **Mit einer Leerzeile**, denn deren Rohzeile ist von Natur aus leer — sie
	// nimmt denselben Weg wie ein Eintrag, dem jemand die seine genommen hat.
	baum, err := msepaket.Zerteile("probe", []byte("init script:\n\ta := 1\n\n\t# dazu\n"))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	knoten := baum.Ersten("init script")

	// Die Rohzeile eines Eintrags fällt weg, sobald jemand den Wert ändert.
	knoten.Kinder[0].Roh = ""
	knoten.Kinder[0].Wert = "2"

	if text := knoten.Skripttext(); text != "a: 2\n\n# dazu" {
		t.Errorf("Skripttext %q", text)
	}
}
