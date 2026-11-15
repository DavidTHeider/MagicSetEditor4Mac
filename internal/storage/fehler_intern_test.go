// Prüfungen an `einordnen` — der Zuordnung von SQLite-Codes zu Fehlerarten.
//
// **Im Paket und nicht daneben**, weil die Funktion nicht exportiert ist. Sie
// war es bis zum Nachlauf des Zweitblicks über Phase 2, und zwar unter deutschem
// Namen — ein Verstoß gegen AGENTS.md §9. Der Ausweg war nicht, sie englisch zu
// nennen, sondern sie einzuziehen: außerhalb dieses Pakets entsteht kein
// SQLite-Fehler, den jemand deuten müsste.
package storage

import (
	"errors"
	"testing"

	sqlite "github.com/ncruces/go-sqlite3"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Fremdschlüsselverletzung ist „nicht gefunden", kein Konflikt.**
//
// Hier stand bis Phase 5 das Gegenteil, und #89 hielt fest, dass die
// Zuordnung zu breit sei — auseinandergelegt werde sie „an einem Fall, der sie
// erzeugt". Der Fall ist da: ein Feldwert für eine Karte, die es nicht gibt.
// Das heißt nicht „überholt", sondern „verfehlt", und der Unterschied
// entscheidet, ob das Frontend nachlädt oder eine Meldung zeigt.
func TestEineFremdschluesselverletzungIstNichtGefunden(t *testing.T) {
	db := offeneDatenbank(t)
	if err := Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("die Migration schlug fehl: %v", err)
	}

	_, err := db.ExecContext(probekontext(t),
		`INSERT INTO cards (set_id, reihenfolge) VALUES (999, 0)`)

	if !errors.Is(einordnen(err), domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet war ErrNotFound", err)
	}
}

// Ein Primärschlüssel bleibt ein Konflikt.
func TestEinePrimaerschluesselverletzungIstEinKonflikt(t *testing.T) {
	db := offeneDatenbank(t)
	if err := Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("die Migration schlug fehl: %v", err)
	}

	for _, anweisung := range []string{
		`INSERT INTO sets (name, spiel, geoeffnet, profil_id) VALUES ('a', 'b', '2026-01-01T00:00:00.000Z', 1)`,
		`INSERT INTO cards (set_id, reihenfolge) VALUES (1, 0)`,
		`INSERT INTO card_values (card_id, feld, wert, reihenfolge) VALUES (1, 'name', 'x', 0)`,
	} {
		if _, err := db.ExecContext(probekontext(t), anweisung); err != nil {
			t.Fatalf("vorbereiten: %v", err)
		}
	}

	_, err := db.ExecContext(probekontext(t),
		`INSERT INTO card_values (card_id, feld, wert, reihenfolge) VALUES (1, 'name', 'y', 1)`)

	if !errors.Is(einordnen(err), domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet war ein Konflikt", err)
	}
}

func TestEinordnenLaesstNilUndFremdesInRuhe(t *testing.T) {
	if einordnen(nil) != nil {
		t.Error("nil wurde zu einem Fehler")
	}

	fremd := errors.New("etwas ganz anderes")
	if !errors.Is(einordnen(fremd), fremd) {
		t.Error("ein fremder Fehler wurde umgedeutet")
	}
}

// **Eine verletzte `NOT NULL`-Bedingung ist kein Konflikt.**
//
// # Die andere Hälfte von #89
//
// Bis Phase 5 galt **jede** verletzte Bedingung als `ErrConflict`, und das
// Schema führt 38 `NOT NULL` und ein `CHECK`. Der Fremdschlüssel ist seither
// abgetrennt; die übrigen fallen **unverändert** durch, weil es dafür keinen
// Fall gibt — und eine Zuordnung ohne Fall wäre eine Behauptung darüber, was
// im Ernstfall geschieht
// (#89).
//
// **Dieser Test hält den Durchfall fest**, nicht die Zuordnung. Wer eines
// Tages ein nullbares Feld einführt oder eine `CHECK`-Bedingung ergänzt, hat
// dann den Fall, den #89 verlangt — und diese Zeile sagt ihm, dass er die
// Entscheidung neu treffen muss, statt sie irgendwo mitzunehmen.
func TestEineNichtNullBedingungIstKeinKonflikt(t *testing.T) {
	db := offeneDatenbank(t)
	if err := Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("die Migration schlug fehl: %v", err)
	}

	_, roh := db.ExecContext(probekontext(t),
		`INSERT INTO sets (name, spiel, geoeffnet, profil_id) VALUES (NULL, 'probe', '2026-01-01T00:00:00.000Z', 1)`)
	if roh == nil {
		t.Fatal("die NOT-NULL-Bedingung wurde nicht verletzt")
	}

	var sqliteFehler *sqlite.Error
	if !errors.As(roh, &sqliteFehler) ||
		sqliteFehler.ExtendedCode() != sqlite.CONSTRAINT_NOTNULL {
		t.Fatalf("Fehler %v, erwartet war SQLITE_CONSTRAINT_NOTNULL", roh)
	}

	fehler := einordnen(roh)

	if errors.Is(fehler, domain.ErrConflict) || errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("%v wurde eingeordnet — für diese Bedingung gibt es keinen Fall", fehler)
	}
}
