package storage

// **Ein Bestand von vor 0.7.0 bleibt, wie er ist** — Vermerk 0189, und der
// Zweitblick über die Folgen vor 0.8.0, Befund 1: an einer Datei, die 0.5.0
// selbst geschrieben hat, und über den ganzen Weg — öffnen, wandern, schließen.

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// bestandAusNullFuenf legt den Bestand aus der Sicherung von 0.5.0 als Datei ab.
func bestandAusNullFuenf(t *testing.T) string {
	t.Helper()

	archiv, fehler := zip.OpenReader(filepath.Join("..", "..", "testdaten", "fassung-0.5.0", "sicherung.zip"))
	if fehler != nil {
		t.Fatal(fehler)
	}

	defer func() { _ = archiv.Close() }() //nolint:errcheck // gelesen

	ziel := filepath.Join(t.TempDir(), "bestand.db")

	for _, datei := range archiv.File {
		if datei.Name != "bestand.db" {
			continue
		}

		quelle, oeffnungsfehler := datei.Open()
		if oeffnungsfehler != nil {
			t.Fatal(oeffnungsfehler)
		}

		inhalt, lesefehler := io.ReadAll(quelle)
		_ = quelle.Close() //nolint:errcheck // gelesen

		if lesefehler != nil {
			t.Fatal(lesefehler)
		}

		if schreibfehler := os.WriteFile(ziel, inhalt, 0o600); schreibfehler != nil {
			t.Fatal(schreibfehler)
		}

		return ziel
	}

	t.Fatal("die Sicherung aus 0.5.0 trägt keinen Bestand")

	return ""
}

// abdruckDerDatei liest Schema, Zeilen, Folgezähler **und** Buchführung einer
// Datei, ohne sie zu öffnen, wie die Anwendung öffnet.
func abdruckDerDatei(t *testing.T, pfad string) (abdruck string, buchfuehrung [][]string) {
	t.Helper()

	db, fehler := FuerLaeufe("file:" + pfad + "?mode=ro")
	if fehler != nil {
		t.Fatal(fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // gelesen

	return bestandsabdruck(t, db), abfrage(t, db, "SELECT * FROM migrationen ORDER BY name")
}

func TestEinBestandAusNullFuenfBleibtUnberuehrt(t *testing.T) {
	pfad := bestandAusNullFuenf(t)
	vorher, buchVorher := abdruckDerDatei(t, pfad)

	db, fehler := Oeffnen(t.Context(), pfad)
	if fehler != nil {
		t.Fatal(fehler)
	}

	var aelter *domain.AeltereFassung
	if wanderfehler := Migriere(t.Context(), db, nil); !errors.As(wanderfehler, &aelter) {
		t.Errorf("der Fehler ist %v, erwartet eine ältere Fassung", wanderfehler)
	}

	if schliessfehler := db.Close(); schliessfehler != nil {
		t.Fatal(schliessfehler)
	}

	nachher, buchNachher := abdruckDerDatei(t, pfad)
	if nachher != vorher {
		t.Errorf("Schema, Zeilen oder Folgezähler haben sich geändert:\n%s", abdruckunterschied(vorher, nachher))
	}

	if len(buchNachher) != len(buchVorher) || len(buchVorher) != 77 {
		t.Errorf("die Buchführung nennt %d Schritte, vorher %d (erwartet 77)", len(buchNachher), len(buchVorher))
	}
}

// **Tabellen ohne Buchführung bekommen auch keine** — die Abweisung steht
// vor `buchfuehrungAnlegen` (Befund 1).
func TestTabellenOhneBuchfuehrungBekommenKeine(t *testing.T) {
	db := leererBestand(t)

	if _, fehler := db.ExecContext(t.Context(), `CREATE TABLE etwas (x)`); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := Migriere(t.Context(), db, nil); !errors.Is(fehler, ErrMigrationen) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if tabelleDa(t.Context(), t, db, "migrationen") {
		t.Error("der fremden Datei wurde eine Buchführung angelegt")
	}
}

// **`sqlite_` ist ein Präfix, kein Muster** (Befund 4): in `LIKE` steht `_`
// für ein beliebiges Zeichen, und eine Tabelle `sqliteX` galt als intern.
func TestEineTabelleMitSqliteImNamenIstKeinLeererBestand(t *testing.T) {
	db := leererBestand(t)

	if _, fehler := db.ExecContext(t.Context(), `CREATE TABLE sqliteX (x)`); fehler != nil {
		t.Fatal(fehler)
	}

	if leer, _ := bestandIstLeerAus(t, db); leer { //nolint:errcheck // gefragt ist die Leere
		t.Error("eine Datei mit der Tabelle sqliteX gilt als leer")
	}
}
