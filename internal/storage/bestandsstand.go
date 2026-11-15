package storage

// **Wie groß ein Bestand ist, und was in seiner Sammlung liegt** — Werkstück
// O7 von Phase 5.2
// (#1431).
//
// # Wozu es das neben `PruefeSicherung` gibt
//
// Jene beantwortet **„darf ich diese Datei übernehmen?"** — und zählt Sets und
// Karten als Komfort daneben. Diese hier beantwortet **„was kostet es, wenn
// ich es tue?"**: sie braucht vier Zahlen statt zwei und dazu die Sammlung
// Gegenstand für Gegenstand, denn was die Sicherung mehr hat als der laufende
// Bestand, steht danach wieder da, ohne dass jemand weiß, ob es noch jemandem
// gehört.
//
// **Zwei Leser derselben Art Datei und trotzdem zwei Wege**: die Prüfung darf
// nicht an einer Zählung scheitern, und die Zählung darf nicht durchgehen,
// wenn sie nichts zu sagen hat.

import (
	"context"
	"database/sql"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// BestandsstandAus liest den Stand einer **fremden** Bestandsdatei.
func BestandsstandAus(ctx context.Context, pfad string) (domain.Bestandsstand, error) {
	db, fehler := nurLesenOeffnen(ctx, pfad)
	if fehler != nil {
		return domain.Bestandsstand{}, fehler
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur gelesen

	return Bestandsstand(ctx, db), nil
}

// Bestandsstand liest den Stand einer offenen Verbindung.
//
// **Ohne Fehler, wie [befundLesen].** Was hier scheitert, ist eine fehlende
// Tabelle oder Spalte in einer sehr alten Datei — und das ist eine Auskunft
// (−1, `SammlungGelesen: false`) und kein Grund, den Vergleich abzubrechen.
func Bestandsstand(ctx context.Context, db *sql.DB) domain.Bestandsstand {
	stand := domain.Bestandsstand{
		Sets:      zaehlen(ctx, db, "sets"),
		Karten:    zaehlen(ctx, db, "cards"),
		Decks:     zaehlen(ctx, db, "decks"),
		Exemplare: besitzZaehlen(ctx, db),
	}

	// **Ohne Zweig, und das ist Absicht.** Fehlt die Zeile oder die ganze
	// Tabelle, bleibt `angelegt` leer; `time.Parse` gibt dann die Nullzeit
	// zurück — und genau die heißt hier „dieser Bestand sagt seinen Zeitpunkt
	// nicht". Zwei `if` darum herum wären zwei Zweige mit demselben Ergebnis,
	// und der zweite ließe sich nicht betreten: der `CHECK` in `0001` lässt
	// keinen unlesbaren Zeitpunkt in die Tabelle (§3).
	var angelegt string

	_ = db.QueryRowContext(ctx, //nolint:errcheck // ein fehlender Zeitpunkt ist die Nullzeit
		`SELECT wert FROM bestand WHERE schluessel = 'angelegt_am'`).Scan(&angelegt)

	stand.AngelegtAm, _ = time.Parse(time.RFC3339Nano, angelegt) //nolint:errcheck // dito

	stand.Sammlung, stand.SammlungGelesen = sammlungsmengen(ctx, db)

	return stand
}

// besitzZaehlen zählt die **besessenen** Stücke, nicht die Zeilen.
//
// **Stücke und nicht Zeilen**: eine Zeile mit vier Stück ist vier Karten, und
// wer nach einer Rückkehr wissen will, was ihm fehlt, meint die Karten.
//
// **Und ohne Wünsche** (0057): was man nicht hat, kann man nicht abgegeben
// haben.
func besitzZaehlen(ctx context.Context, db *sql.DB) int {
	var summe int

	if fehler := db.QueryRowContext(ctx,
		`SELECT coalesce(sum(menge), 0) FROM sammlung WHERE wunsch = 0`).Scan(&summe); fehler != nil {
		return -1
	}

	return summe
}

// sammlungsmengen liest den Besitz je Gegenstand und Profil.
//
// **Das zweite Ergebnis ist der Punkt.** Eine leere Sammlung und eine, die
// sich nicht lesen ließ, sehen gleich aus und bedeuten das Gegenteil.
func sammlungsmengen(ctx context.Context, db *sql.DB) ([]domain.Sammlungsmenge, bool) {
	zeilen, fehler := db.QueryContext(ctx,
		`SELECT profil_id, coalesce(druck_id, 0), printing, zustand, sprache, finish,
		        sum(menge), max(name), max(setcode), max(nummer)
		   FROM sammlung
		  WHERE wunsch = 0
		  GROUP BY profil_id, druck_id, printing, zustand, sprache, finish`)
	if fehler != nil {
		return nil, false
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird über Err()

	mengen := []domain.Sammlungsmenge{}

	for zeilen.Next() {
		var (
			menge   domain.Sammlungsmenge
			druckID int64
			alias   string
		)

		if lesefehler := zeilen.Scan(&menge.Profil, &druckID, &alias, &menge.Zustand,
			&menge.Sprache, &menge.Finish, &menge.Menge, &menge.Name, &menge.Setcode,
			&menge.Nummer); lesefehler != nil {
			return nil, false
		}

		menge.Schluessel = domain.Druckschluessel(druckID, alias)

		mengen = append(mengen, menge)
	}

	if zeilen.Err() != nil {
		return nil, false
	}

	return mengen, true
}
