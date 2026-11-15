package msepaket_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

const testdaten = "../../testdaten"

// Der Fall, an dem eine Map scheitert.
//
// `depends on:` steht in echten Paketen mehrfach — belegt in `docs/FORMATE.md`
// an `magic-m15.mse-style/style`. Ein Zerteiler, der in eine Abbildung schreibt,
// behält den letzten und verliert alle davor; auffallen würde das erst, wenn
// eine Vorlage ihre zweite Abhängigkeit nicht mehr findet.
func TestWiederholteSchluesselGehenNichtVerloren(t *testing.T) {
	baum := zerteileTestdatei(t, "wiederholung.mse-style/style")

	abhaengigkeiten := baum.Alle("depends on")
	if len(abhaengigkeiten) != 2 {
		t.Fatalf("%d Abhängigkeiten gefunden, erwartet waren 2", len(abhaengigkeiten))
	}

	erwartet := []string{"probe.mse-game", "probe-symbole.mse-symbol-font"}
	for i, knoten := range abhaengigkeiten {
		if gefunden := knoten.Ersten("package").Wert; gefunden != erwartet[i] {
			t.Errorf("Abhängigkeit %d ist %q, erwartet war %q", i+1, gefunden, erwartet[i])
		}
	}
}

func TestSchluesselUndWertWerdenGetrennt(t *testing.T) {
	baum := zerteileTestdatei(t, "wiederholung.mse-style/style")

	faelle := map[string]string{
		"mse version": "2.0.0",
		"game":        "probe",
		"short name":  "Wiederholung",
		"card width":  "375",
	}

	for schluessel, wert := range faelle {
		t.Run(schluessel, func(t *testing.T) {
			knoten := baum.Ersten(schluessel)
			if knoten == nil {
				t.Fatalf("der Schlüssel %q fehlt", schluessel)
			}
			if knoten.Wert != wert {
				t.Errorf("Wert %q, erwartet war %q", knoten.Wert, wert)
			}
		})
	}
}

// Die BOM gehört nicht in den ersten Schlüssel.
//
// Echte Pakete beginnen mit ihr — belegt in `docs/FORMATE.md`. Wer sie stehen
// lässt, bekommt einen Schlüssel `\ufeffmse version`, der auf nichts passt, und
// sucht den Fehler danach im Vergleich statt im Einlesen.
func TestDieBOMLandetNichtImErstenSchluessel(t *testing.T) {
	baum := zerteileTestdatei(t, "wiederholung.mse-style/style")

	if baum.Ersten("mse version") == nil {
		t.Error("der erste Schlüssel ist nicht \"mse version\" — steckt die BOM darin?")
	}
}

func TestEinrueckungOeffnetEinenUnterblock(t *testing.T) {
	baum := zerteileTestdatei(t, "tief.mse-style/style")

	karte := baum.Ersten("karte")
	if karte == nil {
		t.Fatal("der Schlüssel \"karte\" fehlt")
	}

	felder := karte.Alle("feld")
	if len(felder) != 2 {
		t.Fatalf("%d Felder, erwartet waren 2", len(felder))
	}

	if name := felder[1].Ersten("name").Wert; name != "kosten" {
		t.Errorf("das zweite Feld heißt %q, erwartet war %q", name, "kosten")
	}
}

// Gemischte Zeilenenden werden gelesen und **gemerkt**.
//
// Das ist die halbe Rundlaufzusage aus `docs/entscheidungen/0002`: was gelesen
// wurde, muss unverändert zurückgeschrieben werden können. Eine Datei mit
// gemischten Enden auf eine Form zu normalisieren wäre eine Änderung, die
// niemand angeordnet hat — und in einem Byte-für-Byte-Vergleich der Beweis
// dafür, dass wir das Set angefasst haben.
func TestGemischteZeilenendenWerdenGemerkt(t *testing.T) {
	baum := zerteileTestdatei(t, "zeilenenden.mse-style/style")

	faelle := map[string]string{
		"mse version": "\r\n",
		"game":        "\n",
		"short name":  "\r\n",
		"card width":  "\n",
	}

	for schluessel, ende := range faelle {
		t.Run(schluessel, func(t *testing.T) {
			knoten := baum.Ersten(schluessel)
			if knoten == nil {
				t.Fatalf("der Schlüssel %q fehlt", schluessel)
			}
			if knoten.Zeilenende != ende {
				t.Errorf("Zeilenende %q, erwartet war %q", knoten.Zeilenende, ende)
			}
		})
	}
}

func TestJederKnotenKenntSeineStelle(t *testing.T) {
	baum := zerteileTestdatei(t, "wiederholung.mse-style/style")

	// Zeile 1 ist `mse version`, Zeile 4 `depends on`, Zeile 5 das eingerückte
	// `package`.
	faelle := []struct {
		schluessel string
		zeile      int
		spalte     int
	}{
		{"mse version", 1, 1},
		{"game", 2, 1},
	}

	for _, fall := range faelle {
		t.Run(fall.schluessel, func(t *testing.T) {
			knoten := baum.Ersten(fall.schluessel)
			if knoten.Stelle.Zeile != fall.zeile || knoten.Stelle.Spalte != fall.spalte {
				t.Errorf("Stelle %d:%d, erwartet war %d:%d",
					knoten.Stelle.Zeile, knoten.Stelle.Spalte, fall.zeile, fall.spalte)
			}
		})
	}

	if tiefer := baum.Alle("depends on")[0].Ersten("package"); tiefer.Stelle.Zeile != 5 {
		t.Errorf("das eingerückte package steht in Zeile %d, erwartet war 5", tiefer.Stelle.Zeile)
	}
}

// Jeder Fehler nennt Datei, Zeile und Spalte.
//
// Fremde Daten sind der Normalfall (AGENTS.md §8): ein Paket kann von 2007
// stammen und halb geschrieben sein. „Zerteilen fehlgeschlagen" hilft dabei
// niemandem — die Stelle schon.
func TestEinFehlerNenntSeineStelle(t *testing.T) {
	faelle := []struct {
		name   string
		inhalt string
		zeile  int
		spalte int
	}{
		// **Der leere Schlüssel ist der letzte Grund, eine Datei abzuweisen.**
		//
		// Hier stand bis #172 die Zeile ohne Doppelpunkt. Sie ist seit
		// [0014](../../docs/entscheidungen/0014-eine-mse-datei-ist-nicht-notwendig-ein-baum.md)
		// Text — eine MSE-Datei ist nicht notwendig ein Baum, und zwölf
		// gesunde Dateien eines bezogenen Packs sind daran gescheitert.
		{
			name:   "ein leerer Schlüssel",
			inhalt: "mse version: 2.0.0\n: ohne schlüssel\n",
			zeile:  2,
			spalte: 1,
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			_, err := msepaket.Zerteile("probe", []byte(fall.inhalt))

			var befund *msepaket.Zerteilfehler
			if !errors.As(err, &befund) {
				t.Fatalf("Fehler %v ist kein *msepaket.Zerteilfehler", err)
			}
			if befund.Stelle.Datei != "probe" {
				t.Errorf("Datei %q, erwartet war %q", befund.Stelle.Datei, "probe")
			}
			if befund.Stelle.Zeile != fall.zeile || befund.Stelle.Spalte != fall.spalte {
				t.Errorf("Stelle %d:%d, erwartet war %d:%d",
					befund.Stelle.Zeile, befund.Stelle.Spalte, fall.zeile, fall.spalte)
			}
			if befund.Error() == "" {
				t.Error("der Fehler trägt keine Meldung")
			}
		})
	}
}

// ---------------------------------------------------------------- Vorrichtung

func zerteileTestdatei(t *testing.T, name string) *msepaket.Baum {
	t.Helper()

	pfad := filepath.Join(testdaten, name)

	inhalt, err := os.ReadFile(pfad) //nolint:gosec // Testdatei aus dem Repository
	if err != nil {
		t.Fatalf("%s nicht lesbar: %v", pfad, err)
	}

	baum, fehler := msepaket.Zerteile(name, inhalt)
	if fehler != nil {
		t.Fatalf("%s nicht zerteilbar: %v", pfad, fehler)
	}

	return baum
}

// Eine Leerzeile schließt keinen Block.
//
// In echten Paketen steht sie zwischen zwei eingerückten Einträgen. Ein
// Zerteiler, der ihre Einrückung (keine) als Ebene nimmt, schließt damit jeden
// offenen Block — und alles danach hängt an der falschen Stelle. Dasselbe gilt
// für eine Kommentarzeile.
func TestLeerzeileUndKommentarSchliessenKeinenBlock(t *testing.T) {
	baum := zerteileTestdatei(t, "tief.mse-style/style")

	karte := baum.Ersten("karte")
	if karte == nil {
		t.Fatal("der Schlüssel \"karte\" fehlt")
	}

	felder := karte.Alle("feld")
	if len(felder) != 2 {
		t.Fatalf("%d Felder, erwartet waren 2 — hat die Leerzeile den Block geschlossen?",
			len(felder))
	}
	if name := felder[1].Ersten("name").Wert; name != "kosten" {
		t.Errorf("das zweite Feld heißt %q, erwartet war %q", name, "kosten")
	}

	// Und der Schlüssel danach hängt wieder an der Wurzel.
	if baum.Ersten("nachher") == nil {
		t.Error("der Schlüssel nach dem Block fehlt in der Wurzel")
	}
}

// Ein fehlender Schlüssel ist in fremden Daten der Normalfall.
//
// Deshalb gibt jeder Zugriff nil zurück, statt zu scheitern — und ein Zugriff
// auf nil ebenfalls. So lässt sich `baum.Ersten("a").Ersten("b")` schreiben,
// ohne jede Zwischenstufe zu prüfen; wer das nicht kann, prüft irgendwann keine.
func TestEinFehlenderSchluesselGibtNichtsZurueck(t *testing.T) {
	baum := zerteileTestdatei(t, "wiederholung.mse-style/style")

	t.Run("im Baum", func(t *testing.T) {
		if baum.Ersten("gibtsNicht") != nil {
			t.Error("ein erfundener Schlüssel wurde gefunden")
		}
		if baum.Alle("gibtsNicht") != nil {
			t.Error("ein erfundener Schlüssel wurde gefunden")
		}
	})

	t.Run("auf einem nil-Knoten", func(t *testing.T) {
		fehlend := baum.Ersten("gibtsNicht")

		if fehlend.Ersten("erst recht nicht") != nil {
			t.Error("ein Zugriff auf nil gab etwas zurück")
		}
		if fehlend.Alle("erst recht nicht") != nil {
			t.Error("ein Zugriff auf nil gab etwas zurück")
		}
	})
}

// Die letzte Zeile darf ohne Zeilenende enden.
//
// Nicht jede Datei endet mit einem Umbruch, und eine, die es nicht tut, ist
// deswegen nicht kaputt. Ihr Zeilenende ist die leere Zeichenkette — genau das,
// was beim Zurückschreiben wieder herauskommen muss.
func TestDieLetzteZeileDarfOhneZeilenendeEnden(t *testing.T) {
	baum, err := msepaket.Zerteile("probe", []byte("game: probe\ncard width: 375"))
	if err != nil {
		t.Fatalf("nicht zerteilbar: %v", err)
	}

	letzter := baum.Ersten("card width")
	if letzter == nil {
		t.Fatal("die letzte Zeile fehlt")
	}
	if letzter.Zeilenende != "" {
		t.Errorf("Zeilenende %q, erwartet war keines", letzter.Zeilenende)
	}
}

func TestEinLeererSchluesselWirdAbgewiesen(t *testing.T) {
	_, err := msepaket.Zerteile("probe", []byte(": ohne Schlüssel\n"))

	var befund *msepaket.Zerteilfehler
	if !errors.As(err, &befund) {
		t.Fatalf("Fehler %v ist kein *msepaket.Zerteilfehler", err)
	}
	if befund.Stelle.Zeile != 1 {
		t.Errorf("Zeile %d, erwartet war 1", befund.Stelle.Zeile)
	}
}

// Der Zerteilfehler schreibt sich wie eine Fundstelle, die man anklicken kann.
func TestDerZerteilfehlerNenntDateiZeileUndSpalte(t *testing.T) {
	_, err := msepaket.Zerteile("style", []byte("a: 1\n: ohne Schlüssel\n"))

	if err == nil {
		t.Fatal("kein Fehler")
	}
	if erwartet := "style:2:1: "; !strings.HasPrefix(err.Error(), erwartet) {
		t.Errorf("Meldung %q beginnt nicht mit %q", err.Error(), erwartet)
	}
}

// **Eine eingerückte Zeile ohne Doppelpunkt ist Fortsetzung, kein Fehler.**
//
// # Woher dieser Test kommt
//
// Aus dem ersten Lauf über ein **bezogenes** Paket
// (#170): von
// siebzehn Stilen in `M15-Magic-Pack` ließen sich fünf öffnen. Die anderen
// zwölf scheiterten an Zeilen wie `	]` — dem Ende einer mehrzeiligen
// Skriptklammer in `init script:`.
//
// Ein Skriptblock ist ein **mehrzeiliger Wert**; seine Zeilen sind kein
// `schlüssel: wert`. `docs/FORMATE.md` führte die Fortsetzungsregel seit Phase 2
// unter „Noch offen" — die Lücke war bekannt, ihre Tragweite nicht: sie machte
// die Mehrheit aller echten Stile unlesbar.
func TestEineZeileOhneDoppelpunktIstFortsetzung(t *testing.T) {
	paket, err := msepaket.Oeffne(filepath.Join("..", "..", "testdaten", "skriptblock.mse-style"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	// **Der Stil dahinter wird trotzdem gelesen.** Ein Zerteiler, der beim
	// Skriptblock stehen bleibt, verliert alles danach.
	// Der Befund über `rand` gehört zum Prüfstück und ist in `stil_test.go`
	// geprüft; hier zählt allein, dass der Stil **hinter** dem Skriptblock
	// überhaupt ankommt.
	stil, _ := paket.Stilmodell()

	if _, gefunden := stil.Feld("name"); !gefunden {
		t.Errorf("der Feldstil hinter dem Skriptblock fehlt: %+v", stil)
	}
}

// **Eine zu tiefe Einrückung landet in der offenen Ebene, statt abzubrechen.**
//
// Echte Pakete rücken so ein: `magic-blends.mse-include/new-blends` setzt einen
// Tabulator vor jeden Eintrag, ohne dass ein Block offen wäre. Wer das abweist,
// kann die Datei nicht lesen (#170).
func TestEineZuTiefeEinrueckungLandetInDerOffenenEbene(t *testing.T) {
	baum, err := msepaket.Zerteile("probe", []byte("a: 1\n\t\tb: zu tief\n"))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	eltern := baum.Ersten("a")
	if eltern == nil {
		t.Fatal("der Eintrag a fehlt")
	}

	if len(eltern.Kinder) != 1 || eltern.Kinder[0].Schluessel != "b" {
		t.Errorf("Kinder von a: %+v", eltern.Kinder)
	}
}
