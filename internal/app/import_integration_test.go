package app_test

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// probeset ist der Pfad des nachgebauten Testsets, vom Paket aus gesehen.
//
// Als Konstante zusammengesetzt und nicht über `filepath.Join` mit einem
// Bestandteil, der selbst Schrägstriche trägt — den meldet gocritic zu Recht:
// `Join` soll Namen verbinden, nicht Pfade aneinanderhängen.
var probeset = filepath.Join("..", "..", "testdaten", "probe.mse-set")

// Der ganze Weg, einmal durch: Paketdatei → Zerteiler → Domäne → Datenbank.
//
// **Er gehört hierher und nirgends sonst.** `internal/service` darf `storage`
// nicht kennen (AGENTS.md §2), also kann dort kein Test gegen echtes Storage
// stehen; `internal/storage` kennt `service` nicht. Der Composition Root ist die
// einzige Schicht, die beide sehen darf — und damit die einzige Stelle, an der
// sich die Naht zwischen ihnen prüfen lässt.
//
// Die Fachtests der beiden Schichten laufen gegen Doubles und sehen diese Naht
// nicht. Genau dort liegt die Fehlerklasse, vor der AGENTS.md §3 warnt: eine
// Vorrichtung, die großzügiger ist als der Betrieb.
func TestEinSetWandertVollstaendigInDenBestand(t *testing.T) {
	db := frischeAppdatenbank(t)

	ergebnis, err := service.NeuerImporteur(storage.NeuerSetSpeicher(db)).
		Set(probekontext(t), probeset)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}

	if ergebnis.Karten != 3 {
		t.Errorf("%d Karten, erwartet waren 3", ergebnis.Karten)
	}

	// Die Karten liegen in ihrer Reihenfolge …
	erwartet := []string{"Feuerball", "Eisball", "Doppelt"}
	for i, name := range erwartet {
		var gelesen string
		if fehler := db.QueryRowContext(probekontext(t),
			`SELECT wert FROM card_values
			   JOIN cards ON cards.id = card_values.card_id
			  WHERE cards.reihenfolge = ? AND card_values.feld = 'name'`, i).Scan(&gelesen); fehler != nil {
			t.Fatalf("Karte %d nicht lesbar: %v", i+1, fehler)
		}
		if gelesen != name {
			t.Errorf("Karte %d ist %q, erwartet war %q", i+1, gelesen, name)
		}
	}

	// … die Herkunft ist vermerkt …
	var herkunft string
	if fehler := db.QueryRowContext(probekontext(t),
		`SELECT herkunft FROM sets`).Scan(&herkunft); fehler != nil {
		t.Fatalf("Herkunft nicht lesbar: %v", fehler)
	}
	if !strings.HasSuffix(herkunft, "probe.mse-set") {
		t.Errorf("Herkunft %q — der Pfad der Quelldatei fehlt", herkunft)
	}

	// … und alles, was nicht abgelegt wird, ist **wörtlich** mit.
	//
	// **Einer seit #1325**: `mse version` reiste hier roh mit, weil niemand es
	// las. Seit dem 2026-09-03 steht es als Setkopf im Bestand — was übrig
	// bleibt, ist der wirklich unverstandene Block.
	roh := rohbloecke(t, db)
	if len(roh) != 1 {
		t.Fatalf("%d Rohblöcke, erwartet war einer (der unverstandene Block)", len(roh))
	}

	if erwartet := "unverstandener block:\n\tirgendwas: bleibt liegen\n"; roh[0] != erwartet {
		t.Errorf("der Rohblock ist %q, erwartet war %q", roh[0], erwartet)
	}
}

// rohbloecke liest die Rohblöcke in ihrer Reihenfolge.
func rohbloecke(t *testing.T, db *sql.DB) []string {
	t.Helper()

	zeilen, err := db.QueryContext(probekontext(t),
		`SELECT inhalt FROM raw_blocks ORDER BY reihenfolge`)
	if err != nil {
		t.Fatalf("Rohblöcke nicht lesbar: %v", err)
	}
	defer func() {
		if fehler := zeilen.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	}()

	var inhalte []string

	for zeilen.Next() {
		var inhalt string
		if fehler := zeilen.Scan(&inhalt); fehler != nil {
			t.Fatalf("Rohblöcke nicht lesbar: %v", fehler)
		}

		inhalte = append(inhalte, inhalt)
	}

	if fehler := zeilen.Err(); fehler != nil {
		t.Fatalf("Rohblöcke nicht lesbar: %v", fehler)
	}

	return inhalte
}

// Und die Suche findet die Karte, ohne dass jemand einen Index pflegt.
func TestEinImportiertesSetIstSofortDurchsuchbar(t *testing.T) {
	db := frischeAppdatenbank(t)

	if _, err := service.NeuerImporteur(storage.NeuerSetSpeicher(db)).
		Set(probekontext(t), probeset); err != nil {
		t.Fatalf("Import: %v", err)
	}

	var treffer int
	if err := db.QueryRowContext(probekontext(t),
		`SELECT count(*) FROM card_suche WHERE card_suche MATCH 'Eisball'`).Scan(&treffer); err != nil {
		t.Fatalf("Suche: %v", err)
	}
	if treffer != 1 {
		t.Errorf("%d Treffer, erwartet war einer", treffer)
	}
}
