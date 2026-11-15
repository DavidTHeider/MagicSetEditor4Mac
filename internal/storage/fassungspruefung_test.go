// **Was ein Bestand aus einer anderen Fassung erlebt** (#795).
//
// Zwei Richtungen, und sie sind nicht symmetrisch: eine **neuere** Datenbank
// wird abgewiesen, weil das Schema unter ihr unbekannt ist. Eine **ältere**
// wurde bis 0.7.0 Schritt für Schritt nachgezogen; seit es nur die Grundlinie
// gibt (Vermerk 0189), wird eine aus der Zeit vor ihr mit Rat abgewiesen —
// das hält `grundlinienstand_intern_test.go`.
package storage

import (
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

// alterStand ist eine erfundene Quelle mit drei Wanderungen.
//
// **Erfunden und nicht die ersten drei dieses Projekts**: die liegen seit
// Vermerk 0189 nicht mehr als Datei vor. Hier geht es um die Maschine — was
// sie mit Namen tut, die sie nicht kennt —, und dafür genügen drei.
func alterStand(t *testing.T) fstest.MapFS {
	t.Helper()

	return fstest.MapFS{
		"migrationen/0001_grundlage.sql":  &fstest.MapFile{Data: []byte("CREATE TABLE eins (id INTEGER)")},
		"migrationen/0002_setbestand.sql": &fstest.MapFile{Data: []byte("CREATE TABLE zwei (id INTEGER)")},
		"migrationen/0003_volltext.sql":   &fstest.MapFile{Data: []byte("CREATE TABLE drei (id INTEGER)")},
	}
}

// **Ein neuerer Bestand wird benannt abgewiesen.**
func TestEinNeuererBestandWirdMitStandAbgewiesen(t *testing.T) {
	db := offeneDatenbank(t)

	quelle := alterStand(t)
	quelle["migrationen/0004_kuenftig.sql"] = &fstest.MapFile{Data: []byte("SELECT 1")}

	if fehler := migriereAus(probekontext(t), db, quelle, nil); fehler != nil {
		t.Fatalf("neuer Stand: %v", fehler)
	}

	fehler := migriereAus(probekontext(t), db, alterStand(t), nil)
	if !errors.Is(fehler, ErrMigrationen) {
		t.Fatalf("Fehler %v, erwartet war ErrMigrationen", fehler)
	}

	for _, teil := range []string{"0004_kuenftig.sql", "0003_volltext.sql", "neuere"} {
		if !strings.Contains(fehler.Error(), teil) {
			t.Errorf("die Meldung nennt %q nicht: %s", teil, fehler)
		}
	}
}

// **Fehlen mehrere, nennt die Meldung alle — und immer gleich.**
//
// `for name := range bereits` lief über eine Map: bei zwei fehlenden
// Migrationen nannte die Meldung eine zufällige, bei jedem Start eine andere.
// Ein Fehlertext, der zwischen zwei Läufen wechselt, taugt weder zur
// Fehlersuche noch als Zusicherung.
func TestDieAbweisungNenntAlleUnbekanntenInFesterFolge(t *testing.T) {
	quelle := alterStand(t)
	quelle["migrationen/0004_eins.sql"] = &fstest.MapFile{Data: []byte("SELECT 1")}
	quelle["migrationen/0005_zwei.sql"] = &fstest.MapFile{Data: []byte("SELECT 2")}

	texte := map[string]int{}

	for lauf := 1; lauf <= 8; lauf++ {
		db := offeneDatenbank(t)

		if fehler := migriereAus(probekontext(t), db, quelle, nil); fehler != nil {
			t.Fatalf("Lauf %d, neuer Stand: %v", lauf, fehler)
		}

		fehler := migriereAus(probekontext(t), db, alterStand(t), nil)
		if fehler == nil {
			t.Fatalf("Lauf %d: kein Fehler", lauf)
		}

		texte[fehler.Error()]++
	}

	if len(texte) != 1 {
		t.Errorf("%d verschiedene Meldungen in acht Läufen, erwartet war eine", len(texte))

		for text := range texte {
			t.Logf("  %s", text)
		}
	}

	for text := range texte {
		for _, teil := range []string{"0004_eins.sql", "0005_zwei.sql"} {
			if !strings.Contains(text, teil) {
				t.Errorf("die Meldung nennt %q nicht: %s", teil, text)
			}
		}
	}
}

// **Eine Fassung ganz ohne Migrationen nennt auch das.**
//
// Der Zweig ist gebaut, weil `namen[len(namen)-1]` sonst in eine leere Liste
// griffe — und er steht hier, weil eine Zeile, die kein Test betritt, eine
// Behauptung darüber ist, was im Ernstfall geschieht (§10). Die
// Anweisungsabdeckung sah ihn nicht: `stand :=` läuft immer, und dass die
// Zuweisung darunter **nicht** greift, zählt sie nicht mit.
//
// **In der ausgelieferten Anwendung kommt der Fall nicht vor, und zwar
// nachweislich nicht:** `//go:embed migrationen/*.sql` verlangt mindestens
// einen Treffer — ohne eine passende Datei scheitert schon die Übersetzung.
// Hier stand „ein Bau, dem sie abhandengekommen sind"; einen solchen Bau gibt
// es nicht, und der Zweitblick über Phase N hat den Satz angestrichen.
//
// Der Zweig bleibt trotzdem, weil `migriereAus` eine austauschbare Quelle
// nimmt — er ist eine Zusage an **diesen** Einstiegspunkt und nicht an
// `Migriere`.
func TestEineFassungOhneJedeMigrationSagtEsAuch(t *testing.T) {
	db := offeneDatenbank(t)

	if fehler := migriereAus(probekontext(t), db, alterStand(t), nil); fehler != nil {
		t.Fatalf("alter Stand: %v", fehler)
	}

	leer := fstest.MapFS{"migrationen": &fstest.MapFile{Mode: fs.ModeDir}}

	fehler := migriereAus(probekontext(t), db, leer, nil)
	if !errors.Is(fehler, ErrMigrationen) {
		t.Fatalf("Fehler %v, erwartet war ErrMigrationen", fehler)
	}

	if !strings.Contains(fehler.Error(), "keine einzige Migration") {
		t.Errorf("die Meldung nennt den leeren Stand nicht: %s", fehler)
	}
}
