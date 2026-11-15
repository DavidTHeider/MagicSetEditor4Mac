// **Der Lebenszyklus der Sicherungsnaht** —
// #1169.
//
// # Warum ein interner Lauf
//
// `schliessen` und `oeffnen` sind Rückrufe, die der Dienst in einer bestimmten
// **Reihenfolge** aufruft. Ihre Fehlerwege über den ganzen Weg herzustellen
// hieße, ein Zurückspielen so zu präparieren, dass genau der Wiederöffnen-
// Schritt scheitert — und dann prüfte der Lauf die Präparation.
//
// Hier werden sie einzeln gerufen. Was dabei **nicht** geprüft wird, steht
// daneben: dass der Dienst sie in dieser Reihenfolge ruft, prüft
// `bestandssicherungnaht_test.go` über den ganzen Weg.

package app

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Bestand, der sich nicht öffnen lässt, kommt als Fehler zurück.**
//
// **Und nicht als Absturz**: der Aufrufer bekommt seinen Griff zurück, und was
// er darauf tut, meldet sich. Ein `nil` an dieser Stelle brächte `mitBestand`
// in `cmd/mse` zum Absturz — sein `defer` schließt die Verbindung.
func TestEinUnoeffenbarerBestandMeldetSich(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// **Ein Verzeichnis am Ort der Datei.** SQLite kann es nicht öffnen, und
	// der Fehler ist einer des Öffnens und keiner der Migration.
	ort := filepath.Join(t.TempDir(), "ist-ein-ordner")
	if err := os.Mkdir(ort, 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	var db *sql.DB

	naht := BestandsnahtUeber(&db, ort)

	fehler := naht.oeffnen(probekontext(t))
	if fehler == nil {
		t.Fatal("ein Verzeichnis wurde als Bestand geöffnet")
	}

	if !strings.Contains(fehler.Error(), "öffnen") &&
		!strings.Contains(fehler.Error(), "migrieren") {
		t.Errorf("die Meldung nennt den Schritt nicht: %v", fehler)
	}
}

// **Zweimal schließen ist harmlos.**
//
// Der Weg räumt auf jedem Ausgang auf; ein zweiter Aufruf darf kein Fehler
// sein, und er darf erst recht nicht abstürzen.
func TestZweimalSchliessenIstHarmlos(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	ort := filepath.Join(t.TempDir(), "bestand.db")

	db, err := storage.Oeffnen(probekontext(t), ort)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	naht := BestandsnahtUeber(&db, ort)

	if fehler := naht.schliessen(); fehler != nil {
		t.Fatalf("erstes Schließen: %v", fehler)
	}

	if fehler := naht.schliessen(); fehler != nil {
		t.Errorf("zweites Schließen: %v", fehler)
	}
}

// **Und eine Datei, die sich öffnen, aber nicht migrieren lässt.**
//
// # Warum beide Zweige einzeln stehen
//
// Sie sagen Verschiedenes. „Lässt sich nicht öffnen" heißt: es ist keine
// Datenbank. „Lässt sich nicht migrieren" heißt: es ist eine, aber eine
// fremde — und wer das zweite als das erste meldete, schickte jemanden auf
// die Suche nach einer kaputten Datei, die in Ordnung ist.
//
// Hergestellt wird es mit einer Tabelle, die eine Migration selbst anlegen
// will: `CREATE TABLE sets` scheitert an der, die schon da ist.
func TestEinUnmigrierbarerBestandMeldetSich(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	ort := filepath.Join(t.TempDir(), "fremd.db")

	fremd, err := storage.Oeffnen(probekontext(t), ort)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if _, fehler := fremd.ExecContext(probekontext(t),
		`CREATE TABLE sets (irgendwas TEXT)`); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if fehler := fremd.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	var db *sql.DB

	naht := BestandsnahtUeber(&db, ort)

	fehler := naht.oeffnen(probekontext(t))
	if fehler == nil {
		t.Fatal("die fremde Datenbank wurde migriert")
	}

	if !strings.Contains(fehler.Error(), "migrieren") {
		t.Errorf("die Meldung nennt den Schritt nicht: %v", fehler)
	}
}
