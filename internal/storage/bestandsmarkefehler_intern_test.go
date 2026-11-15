package storage

// **Der Fehlerweg des Zählens** —
// #967.
//
// # Warum paketintern
//
// Die Marke geht über zwei Abfragen: erst die Tabellen, dann die Zahlen.
// Scheitert die erste, kommt die zweite nie dran — über eine geschlossene
// Verbindung ist ihr Fehlerweg deshalb nicht zu erreichen.
//
// **Und er ist kein erfundener Fall.** Zwischen beiden Abfragen kann eine
// Tabelle verschwinden: eine fremde Datei, die jemand austauscht, ein
// zweiter Prozess am selben Bestand. Genau die Bedrohung, um die es in
// #964 geht.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Tabelle, die es beim Zählen nicht mehr gibt, meldet sich.**
func TestEineVerschwundeneTabelleMeldetSich(t *testing.T) {
	t.Parallel()

	db := offeneProbe(t)

	if _, fehler := db.ExecContext(probekontext(t),
		`CREATE TABLE sets (id INTEGER PRIMARY KEY)`); fehler != nil {
		t.Fatalf("Schema anlegen: %v", fehler)
	}

	_, err := zaehlungen(probekontext(t), db, []tabelle{
		{name: "sets", mitZeilennummer: true},
		{name: "diese-tabelle-gibt-es-nicht", mitZeilennummer: true},
	})

	if !errors.Is(err, domain.ErrSicherung) {
		t.Fatalf("Fehler %v, erwartet war ein Sicherungsfehler", err)
	}
}

// **Und eine Tabelle, die dasteht, aber nicht zu zählen ist.**
//
// # Warum das kein erfundener Fall ist
//
// Eine FTS5-Tabelle steht in `sqlite_master` wie jede andere; ihr Inhalt
// liegt in Schattentabellen daneben. Fehlt eine davon — eine beschädigte
// Datei, ein halb kopierter Bestand —, ist die Tabelle **aufgeführt und nicht
// lesbar**.
//
// Das ist genau die Lücke zwischen den beiden Abfragen der Marke: die erste
// findet sie, die zweite scheitert an ihr. Über die öffentliche Naht ist der
// Weg sonst nicht zu erreichen.
func TestEineUnlesbareTabelleGibtKeineMarke(t *testing.T) {
	t.Parallel()

	db := offeneProbe(t)

	if _, fehler := db.ExecContext(probekontext(t),
		`CREATE VIRTUAL TABLE suche USING fts5(text)`); fehler != nil {
		t.Fatalf("die Volltextsuche anlegen: %v", fehler)
	}

	if _, fehler := db.ExecContext(probekontext(t),
		`DROP TABLE suche_data`); fehler != nil {
		t.Fatalf("die Schattentabelle entfernen: %v", fehler)
	}

	if _, err := BestandsmarkeVon(probekontext(t), db); !errors.Is(err, domain.ErrSicherung) {
		t.Errorf("Fehler %v, erwartet war ein Sicherungsfehler", err)
	}
}
