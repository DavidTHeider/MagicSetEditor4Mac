package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine verletzte Bedingung sagt, welche** — ENG-05.AK-03
// (#1730),
// nativer Befund F08.
//
// # Was vorher geschah
//
// **Jede verletzte Eindeutigkeit hieß „ein Eintrag mit diesen Angaben liegt
// schon da"**, und das Fenster erklärte sie als fremde Änderung — „Der
// Eintrag wurde zwischenzeitlich geändert. Bitte neu laden." Eine verletzte
// `CHECK`-Bedingung, ein fehlender Pflichtwert und ein abweisender Auslöser
// gingen gar nicht durch die Einordnung und kamen als „unerwarteter Fehler"
// an.
//
// # Wie geprüft wird
//
// **Mit echten Treiberfehlern aus einem migrierten Bestand**, nicht mit
// nachgebauten: die Spalten stehen im Text des Treibers, und ein Nachbau
// prüfte nur, ob die Zerlegung ihren eigenen Nachbau versteht.
func TestEineVerletzteBedingungSagtWelche(t *testing.T) {
	t.Parallel()

	db := migrierteDatenbank(t)
	ctx := probekontext(t)

	for _, vorab := range []string{
		`INSERT INTO decks (name, spiel, profil_id) VALUES ('Probe', 'magic', 1)`,
		`INSERT INTO bestand (schluessel, wert) VALUES ('probe', '1')`,
	} {
		if _, fehler := db.ExecContext(ctx, vorab); fehler != nil {
			t.Fatalf("%s: %v", vorab, fehler)
		}
	}

	for _, fall := range []struct {
		name       string
		anweisung  string
		art        error
		schluessel domain.Schluessel
		feld       string
	}{
		{
			name:       "ein zweites Deck gleichen Namens steht am Namen",
			anweisung:  `INSERT INTO decks (name, spiel, profil_id) VALUES ('Probe', 'magic', 1)`,
			art:        domain.ErrDoppelt,
			schluessel: domain.SchluesselDoppelterName,
			feld:       "name",
		},
		{
			name:       "ein doppelter Schlüssel ohne Zuordnung ist doppelt, ohne Feld",
			anweisung:  `INSERT INTO bestand (schluessel, wert) VALUES ('probe', '2')`,
			art:        domain.ErrDoppelt,
			schluessel: domain.SchluesselDoppelt,
		},
		{
			name: "eine verletzte CHECK-Bedingung ist eine Validierung",
			anweisung: `INSERT INTO kostenbudgets (profil_id, konto, zeitraum, betrag, waehrung, geaendert_am)
			            VALUES (1, 'k', 'tag', -1, 'EUR', '')`,
			art: domain.ErrValidation,
		},
		{
			name:      "ein fehlender Pflichtwert ist eine Validierung",
			anweisung: `INSERT INTO nutzer (name) VALUES (NULL)`,
			art:       domain.ErrValidation,
		},
		{
			name:      "ein abweisender Auslöser ist eine Validierung",
			anweisung: `INSERT INTO decks (name, spiel, profil_id) VALUES ('Ohne', 'magic', 0)`,
			art:       domain.ErrValidation,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			_, treiber := db.ExecContext(ctx, fall.anweisung)
			if treiber == nil {
				t.Fatal("die Anweisung ging durch")
			}

			eingeordnet := einordnen(treiber)

			if !errors.Is(eingeordnet, fall.art) {
				t.Fatalf("%v ist nicht %v", eingeordnet, fall.art)
			}

			// **Der Treibertext bleibt im Protokoll und nicht im Satz** (§8).
			if grund := domain.GrundVon(eingeordnet); grund == nil {
				t.Error("der Treibertext ging verloren")
			}

			if strings.Contains(eingeordnet.Error(), "constraint failed") {
				t.Errorf("der Treibertext steht im Satz: %q", eingeordnet.Error())
			}

			if fall.schluessel == "" {
				return
			}

			var befund *domain.FieldError
			if !errors.As(eingeordnet, &befund) {
				t.Fatalf("%v ist kein Feldbefund", eingeordnet)
			}

			if befund.Key != fall.schluessel || befund.Field != fall.feld {
				t.Errorf("Schlüssel %q am Feld %q, erwartet %q am Feld %q",
					befund.Key, befund.Field, fall.schluessel, fall.feld)
			}
		})
	}
}

// **Die Zerlegung liest, was dieser Treiber schreibt** — und wird rot, wenn
// eine neue Fassung es anders schreibt.
//
// **Nur für den Feldbezug.** Die Art entscheidet der Fehlercode (§8); fiele
// die Zerlegung aus, bliebe es ein doppelter Eintrag — ohne Feld, aber nie
// eine falsche Art.
func TestDieEindeutigkeitNenntTafelUndSpalten(t *testing.T) {
	t.Parallel()

	db := migrierteDatenbank(t)
	ctx := probekontext(t)

	anweisung := `INSERT INTO decks (name, spiel, profil_id) VALUES ('Zweimal', 'magic', 1)`
	if _, fehler := db.ExecContext(ctx, anweisung); fehler != nil {
		t.Fatalf("das erste Deck: %v", fehler)
	}

	_, treiber := db.ExecContext(ctx, anweisung)

	tafel, spalten := verletzteEindeutigkeit(treiber)
	if tafel != "decks" || strings.Join(spalten, ",") != "profil_id,spiel,name" {
		t.Errorf("gelesen %q %v aus %q", tafel, spalten, treiber)
	}

	for _, fremd := range []error{
		nil,
		errors.New("etwas ganz anderes"),
		fmt.Errorf("UNIQUE constraint failed: "),
		fmt.Errorf("UNIQUE constraint failed: ohnePunkt"),
	} {
		if tafel, spalten := verletzteEindeutigkeit(fremd); tafel != "" || spalten != nil {
			t.Errorf("%v gab %q %v", fremd, tafel, spalten)
		}
	}
}

// **Jede Zuordnung zeigt auf eine Eindeutigkeit, die es gibt** — aus dem
// Schema gelesen und nicht aus einer Liste daneben. Wer eine Tafel umbaut
// und eine Spalte umbenennt, sieht es hier und nicht an einem Feld, an dem
// die Meldung nie mehr erscheint.
func TestJedeZuordnungZeigtAufEineEindeutigkeit(t *testing.T) {
	t.Parallel()

	db := migrierteDatenbank(t)

	vorhanden, fehler := eindeutigkeitenImSchema(probekontext(t), db)
	if fehler != nil {
		t.Fatalf("das Schema lesen: %v", fehler)
	}

	for schluessel := range eindeutigkeiten {
		if !vorhanden[schluessel] {
			t.Errorf("%s ist keine Eindeutigkeit im Schema", schluessel)
		}
	}

	if len(vorhanden) < len(eindeutigkeiten) {
		t.Errorf("%d Eindeutigkeiten im Schema, weniger als zugeordnet", len(vorhanden))
	}
}

// eindeutigkeitenImSchema liest jede Eindeutigkeit des Bestands so, wie der
// Treiber sie nennt: `tafel(spalte, spalte)`.
func eindeutigkeitenImSchema(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	zeilen, fehler := db.QueryContext(ctx,
		`SELECT m.name || '(' || group_concat(i.name, ', ' ORDER BY i.seqno) || ')'
		   FROM sqlite_master AS m,
		        pragma_index_list(m.name) AS l,
		        pragma_index_info(l.name) AS i
		  WHERE m.type = 'table' AND l."unique" = 1
		  GROUP BY m.name, l.name`)
	if fehler != nil {
		return nil, fehler
	}

	defer func() { _ = zeilen.Close() }()

	vorhanden := map[string]bool{}

	for zeilen.Next() {
		var name string
		if fehler := zeilen.Scan(&name); fehler != nil {
			return nil, fehler
		}

		vorhanden[name] = true
	}

	return vorhanden, zeilen.Err()
}
