// Die Vorrichtung für die Fehlerwege der Sicherungsprüfung (#940).
package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// handgemacht legt eine Datenbank aus rohem SQL an — **ohne Migration**.
//
// **Der Punkt ist, was fehlt.** Eine migrierte Datenbank hat jede Tabelle;
// geprüft werden soll, was geschieht, wenn eine nicht da ist.
func handgemacht(t *testing.T, schema string) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "handgemacht.db")

	db, err := FuerLaeufe("file:" + pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if _, fehler := db.ExecContext(t.Context(), schema); fehler != nil {
		t.Fatalf("Schema anlegen: %v", fehler)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	return pfad
}

// ohneOeffner lässt das Öffnen für die Dauer eines Tests scheitern.
//
// **Über die Naht und nicht über einen erfundenen Treibernamen** — seit
// Werkstück 14 von Phase 4.12 öffnet [oeffner] und nicht `sql.Open`. Der
// Unterschied ist mehr als ein Name: `sql.Open` scheiterte **nur** am
// Treibernamen, diese Naht scheitert an dem, was der Lauf will.
func ohneOeffner(t *testing.T) {
	t.Helper()

	vorher := oeffner
	oeffner = func(string, string) (*sql.DB, error) {
		return nil, errors.New("das Öffnen ist in diesem Lauf abgeschaltet")
	}

	t.Cleanup(func() { oeffner = vorher })
}

// **Eine unlesbare Migrationsquelle wird gemeldet, nicht übergangen.**
//
// Der Zweig ist von außen unerreichbar — das eingebettete Verzeichnis ist
// immer in Ordnung —, und genau deshalb bekommt `lesenUndPruefen` die Quelle
// als Parameter (§3).
func TestLesenUndPruefenMeldetEineUnlesbareQuelle(t *testing.T) {
	pfad := handgemacht(t, `
		CREATE TABLE bestand (schluessel TEXT PRIMARY KEY, wert TEXT NOT NULL);
		INSERT INTO bestand VALUES ('angelegt_am', '2026-08-23T00:00:00.000Z');
		CREATE TABLE migrationen (name TEXT PRIMARY KEY, angewendet TEXT NOT NULL,
		                          pruefsumme TEXT NOT NULL DEFAULT '');`)

	db, err := FuerLaeufe("file:" + pfad + "?mode=ro")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // gelesen ist gelesen

	for name, fall := range map[string]struct{ quelle, basis fstest.MapFS }{
		// Ein Verzeichnis, dessen einzige Datei nicht der Form entspricht.
		"eine Wanderung ohne Nummer": {quelle: fstest.MapFS{"migrationen/keine-nummer.sql": &fstest.MapFile{}}},
		// **Und eine Grundlinie ohne Umfang** (Vermerk 0189).
		"eine Grundlinie ohne Umfang": {
			quelle: fstest.MapFS{"migrationen": &fstest.MapFile{Mode: fs.ModeDir}},
			basis:  fstest.MapFS{"grundlinie/saat.sql": &fstest.MapFile{}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			var basis fs.FS
			if fall.basis != nil {
				basis = fall.basis
			}

			if _, fehler := lesenUndPruefen(t.Context(), db, fs.FS(fall.quelle), basis, pfad); fehler == nil {
				t.Fatal("die unlesbare Quelle wurde übergangen")
			}
		})
	}
}

// migrierteDatei legt eine migrierte Datenbank als Datei an.
func migrierteDatei(t *testing.T) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "sicherung.db")

	db, err := Oeffnen(t.Context(), "file:"+pfad)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if fehler := Migriere(t.Context(), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	return pfad
}

// **Und `PruefeSicherung` gibt den Kontext weiter.**
func TestPruefeSicherungBrichtMitDemKontextAb(t *testing.T) {
	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	if _, err := PruefeSicherung(ctx, handgemacht(t, `CREATE TABLE leer (x TEXT)`)); err == nil {
		t.Fatal("der Abbruch wurde übergangen")
	}
}

// **Die Fehlerwege der Prüfung** — jeder betreten, keiner behauptet (§3).
func TestPruefeSicherungFehlerwege(t *testing.T) {
	t.Parallel()

	t.Run("ohne Buchführung", func(t *testing.T) {
		t.Parallel()

		// Ein Bestand ohne `migrationen`: die Form stimmt, die Frage nach der
		// Fassung geht ins Leere.
		pfad := handgemacht(t, `
			CREATE TABLE bestand (schluessel TEXT PRIMARY KEY, wert TEXT NOT NULL);
			INSERT INTO bestand VALUES ('angelegt_am', '2026-08-23T00:00:00.000Z');`)

		if _, err := PruefeSicherung(t.Context(), pfad); err == nil {
			t.Fatal("eine Datei ohne Buchführung wurde angenommen")
		}
	})

	t.Run("ohne eine verbuchte Wanderung", func(t *testing.T) {
		t.Parallel()

		// **Eine Buchführung ohne Zeile ist kein Bestand** (Vermerk 0189):
		// beim Start würde er abgewiesen, und was sich nicht öffnen lässt,
		// wird nicht erst zurückgespielt.
		pfad := handgemacht(t, `
			CREATE TABLE bestand (schluessel TEXT PRIMARY KEY, wert TEXT NOT NULL);
			INSERT INTO bestand VALUES ('angelegt_am', '2026-08-23T00:00:00.000Z');
			CREATE TABLE migrationen (name TEXT PRIMARY KEY, angewendet TEXT NOT NULL,
			                          pruefsumme TEXT NOT NULL DEFAULT '');`)

		_, err := PruefeSicherung(t.Context(), pfad)
		if !errors.Is(err, domain.ErrSicherung) {
			t.Fatalf("eine leere Buchführung wurde nicht abgewiesen: %v", err)
		}

		// **Und nicht als ältere Fassung** — sie stammt aus keiner.
		var aelter *domain.AeltereFassung
		if errors.As(err, &aelter) {
			t.Errorf("eine leere Buchführung gilt als ältere Fassung: %v", err)
		}
	})
}

// **−1 und nicht 0**: „diese Fassung kennt die Tabelle nicht" ist eine andere
// Auskunft als „sie ist leer". Über die Prüfung ist der Fall nicht mehr zu
// erreichen — seit es nur die Grundlinie gibt, verlangt die Mindestform jede
// ihrer Tabellen —, und deshalb wird er hier unmittelbar betreten.
func TestEineUnbekannteTabelleZaehltMinusEins(t *testing.T) {
	t.Parallel()

	db := leererBestand(t)

	if anzahl := zaehlen(t.Context(), db, "sets"); anzahl != -1 {
		t.Errorf("gezählt wurden %d, erwartet war −1", anzahl)
	}
}

// **Ohne Treiber lässt sich nichts prüfen** — und das wird gemeldet.
//
// **Nicht parallel und als eigener Test**: der Treibername ist global, und wer
// ihn während eines parallelen Laufs verstellt, macht dessen Fehlschlag zu
// seinem eigenen. Dieselbe Gestalt wie [TestOeffnenMeldetEinenUnbekanntenTreiber].
func TestPruefeSicherungMeldetEinenUnbekanntenTreiber(t *testing.T) {
	// **Die Datei entsteht vorher.** Sie anzulegen braucht denselben Treiber.
	pfad := migrierteDatei(t)

	ohneOeffner(t)

	if _, err := PruefeSicherung(t.Context(), pfad); err == nil {
		t.Fatal("ohne Treiber wurde die Datei angenommen")
	}
}

// **Ein Dateiname mit `?` oder `#` wird gelesen, nicht zerschnitten.**
//
// Angestrichen vom Zweitblick über Phase S. Der Pfad kommt aus dem
// Dateidialog; macOS lässt beide Zeichen zu. Ungequotet in eine SQLite-URI
// gesetzt, endet der Pfad vor dem Zeichen — `mode=ro` geht verloren, und
// SQLite öffnet oder **legt an**, was davor steht.
//
// **Beleg:** SQLite verlangt für die Umwandlung eines Dateinamens in eine URI
// ausdrücklich `%3f` für `?` und `%23` für `#`
// (https://www.sqlite.org/uri.html#the_uri_path, nachgelesen am 2026-08-23).
func TestPruefeSicherungLiestEinenPfadMitSonderzeichen(t *testing.T) {
	t.Parallel()

	// **Die Vorlage entsteht außerhalb der Unterläufe.** Deren Verzeichnisse
	// tragen den Fallnamen — und damit dasselbe Sonderzeichen; sie darin
	// anzulegen prüfte den Fehler mit dem Fehler.
	roh, err := os.ReadFile(migrierteDatei(t)) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil {
		t.Fatalf("die Vorlage lesen: %v", err)
	}

	ort := t.TempDir()

	for _, name := range []string{"vorher?eins.db", "vorher#eins.db", "vorher%eins.db"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ziel := filepath.Join(ort, name)
			// #nosec G703 -- `ziel` ist aus t.TempDir() und einem festen Namen
			// dieser Datei zusammengesetzt; es gibt keine Eingabe.
			if fehler := os.WriteFile(ziel, roh, 0o600); fehler != nil {
				t.Fatalf("schreiben: %v", fehler)
			}

			if _, fehler := PruefeSicherung(t.Context(), ziel); fehler != nil {
				t.Fatalf("die Datei %q wurde nicht gelesen: %v", name, fehler)
			}
		})
	}
}

// **Ohne Treiber lässt sich die Belegung nicht prüfen** — und das wird gemeldet.
//
// **Nicht parallel**, aus demselben Grund wie bei der Sicherungsprüfung: der
// Treibername ist global.
func TestAlleinAmBestandMeldetEinenUnbekanntenTreiber(t *testing.T) {
	pfad := migrierteDatei(t)

	ohneOeffner(t)

	if err := AlleinAmBestand(t.Context(), pfad); err == nil {
		t.Fatal("ohne Treiber galt die Datei als frei")
	}
}

// **Eine Sicherung ohne die Tabellen, die sie verbucht hat, wird abgewiesen** —
// #963.
//
// # Was vorher galt
//
// Geprüft wurden die **Form** (`bestand.angelegt_am`) und die **Fassung**
// (unbekannte Migrationen). Nicht geprüft wurde, ob die Fachtabellen da sind:
// `zaehlen` gibt für eine fehlende Tabelle −1, und das galt als Auskunft.
//
// Eine Datei mit gültigem `angelegt_am` und allen bekannten Migrationsnamen
// samt passenden Prüfsummen, **aber ohne `sets` und `cards`**, bestand Auswahl
// und Start. Erst der nächste fachliche Aufruf scheiterte — nachdem die
// Oberfläche „Zurückgespielt" gemeldet hatte.
//
// # Die Mindestform hängt am verbuchten Stand
//
// Nicht an einer festen Liste: verlangt wird, was die verbuchten Schritte
// anlegen. **Wer auf der Grundlinie steht, muss `sets` haben** — sie legt die
// Tabelle an.
func TestEineSicherungOhneIhreTabellenWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	namen, _, err := grundlinienumfang(grundlinie)
	if err != nil {
		t.Fatal(err)
	}

	var schema strings.Builder

	schema.WriteString(`
		CREATE TABLE bestand (schluessel TEXT PRIMARY KEY, wert TEXT NOT NULL);
		INSERT INTO bestand VALUES ('angelegt_am', '2026-08-23T00:00:00.000Z');
		CREATE TABLE migrationen (name TEXT PRIMARY KEY, angewendet TEXT NOT NULL,
		                          pruefsumme TEXT NOT NULL DEFAULT '');`)

	for _, name := range namen {
		fmt.Fprintf(&schema, "INSERT INTO migrationen VALUES ('%s', '2026-08-23', '');\n", name)
	}

	pfad := handgemacht(t, schema.String())

	_, err = PruefeSicherung(t.Context(), pfad)
	if err == nil {
		t.Fatal("eine Sicherung ohne `sets` wurde angenommen")
	}

	if !errors.Is(err, domain.ErrSicherung) {
		t.Errorf("der Fehler ist %v", err)
	}

	// **Die Meldung nennt die Tabelle.** „Die Sicherung ist unbrauchbar" sagt
	// niemandem, was zu tun ist; `sets` fehlt ist eine Auskunft (§8).
	if !strings.Contains(err.Error(), "sets") {
		t.Errorf("die Meldung nennt die Tabelle nicht: %v", err)
	}
}

// **Eine sehr alte Sicherung bekommt denselben Rat wie ihr Bestand** —
// Zweitblick über die Folgen vor 0.8.0, Befund 2: ihrer Buchführung fehlt die
// Prüfsummenspalte, und die Prüfung scheiterte daran, bevor sie das Alter
// sah. „Wähle eine andere" ist dann der falsche Rat; der richtige führt über
// 0.7.0.
func TestEineSicherungOhnePruefsummenspalteIstAelter(t *testing.T) {
	t.Parallel()

	pfad := handgemacht(t, `
		CREATE TABLE bestand (schluessel TEXT PRIMARY KEY, wert TEXT NOT NULL);
		INSERT INTO bestand VALUES ('angelegt_am', '2026-08-23T00:00:00.000Z');
		CREATE TABLE migrationen (name TEXT PRIMARY KEY, angewendet TEXT NOT NULL);
		INSERT INTO migrationen VALUES ('0001_grundlage.sql', '2026-08-23T00:00:00.000Z');`)

	var aelter *domain.AeltereFassung
	if _, err := PruefeSicherung(t.Context(), pfad); !errors.As(err, &aelter) {
		t.Errorf("der Fehler ist %v, erwartet eine ältere Fassung", err)
	}
}

// **Eine Buchführung, die sich nicht lesen lässt, macht die Sicherung
// unbrauchbar** — ohne `NOT NULL` steht ein Name `NULL`, und das Lesen nach
// `string` scheitert daran.
func TestEineSicherungMitUnlesbarerBuchfuehrungTaugtNicht(t *testing.T) {
	t.Parallel()

	pfad := handgemacht(t, `
		CREATE TABLE bestand (schluessel TEXT PRIMARY KEY, wert TEXT NOT NULL);
		INSERT INTO bestand VALUES ('angelegt_am', '2026-08-23T00:00:00.000Z');
		CREATE TABLE migrationen (name TEXT, angewendet TEXT, pruefsumme TEXT);
		INSERT INTO migrationen VALUES (NULL, 'x', 'y');`)

	if _, err := PruefeSicherung(t.Context(), pfad); !errors.Is(err, domain.ErrSicherung) {
		t.Errorf("der Fehler ist %v", err)
	}
}
