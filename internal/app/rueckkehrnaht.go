package app

// **Die Naht zum Buch der Rückkehren** — Werkstück O7 von Phase 5.2
// (#1431).
//
// # Warum sie die Verbindung holt und nicht hält
//
// **Weil ein Zurückspielen sie wechselt.** Nach dem Tausch ist die alte
// Verbindung tot; wer sie beim Aufbau festhielte, schriebe die Epoche in eine
// Datei, die es nicht mehr gibt — und das ist ausgerechnet die Epoche, die den
// Tausch beschreibt. Derselbe Grund wie bei `chiffrenaht`.

import (
	"context"
	"database/sql"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// bestandsschaunaht liest den Stand einer **fremden** Bestandsdatei.
//
// **Eine Variable und keine Zeile in `sicherungsnaht.go`**: jene Datei ist
// keine Wurzel und darf `internal/storage` nicht kennen (§2) — dieselbe
// Gestalt wie `sicherungsnaht` und `alleinnaht` daneben.
var bestandsschaunaht = storage.BestandsstandAus

// rueckkehrnaht baut das Buch bei jedem Ruf neu.
type rueckkehrnaht struct{ holen func() *sql.DB }

func (r rueckkehrnaht) buch() *storage.Wiederherstellungsbuch {
	return storage.NeuesWiederherstellungsbuch(r.holen())
}

func (r rueckkehrnaht) Vermerken(
	ctx context.Context, epoche domain.Wiederherstellungsepoche,
) (int64, error) {
	return r.buch().Vermerken(ctx, epoche)
}

func (r rueckkehrnaht) Offene(ctx context.Context) (
	[]domain.Wiederherstellungsepoche, error,
) {
	return r.buch().Offene(ctx)
}

func (r rueckkehrnaht) Alle(ctx context.Context) (
	[]domain.Wiederherstellungsepoche, error,
) {
	return r.buch().Alle(ctx)
}

func (r rueckkehrnaht) Quittieren(
	ctx context.Context, kennung int64, wann time.Time,
) error {
	return r.buch().Quittieren(ctx, kennung, wann)
}

// eigenstandnaht liest den Stand der **laufenden** Verbindung.
type eigenstandnaht struct{ holen func() *sql.DB }

func (e eigenstandnaht) stand(ctx context.Context) domain.Bestandsstand {
	// **Ohne Verbindung ist der Stand unbekannt und nicht leer.** Ein leerer
	// hieße „nichts in der Sammlung", und daraus folgte „nichts unklar" —
	// die eine Auskunft, die niemals falsch sein darf.
	db := e.holen()
	if db == nil {
		return domain.Bestandsstand{Sets: -1, Karten: -1, Decks: -1, Exemplare: -1}
	}

	return storage.Bestandsstand(ctx, db)
}

// ruecksprungnaht führt die Fragen der Fassade an den Bestandsweg.
//
// **Über eine Funktion und nicht über einen festgehaltenen Weg**: er entsteht
// bei jedem Ruf neu, weil die Verbindung darunter nach einem Zurückspielen eine
// andere ist.
type ruecksprungnaht struct{ weg func() *service.Bestandsweg }

func (r ruecksprungnaht) Ruecksprunglage(ctx context.Context, pfad string) (
	domain.Ruecksprunglage, error,
) {
	return r.weg().Ruecksprunglage(ctx, pfad)
}

func (r ruecksprungnaht) OffeneRueckkehren(ctx context.Context) (
	[]domain.Wiederherstellungsepoche, error,
) {
	return r.weg().OffeneRueckkehren(ctx)
}

func (r ruecksprungnaht) Rueckkehren(ctx context.Context) (
	[]domain.Wiederherstellungsepoche, error,
) {
	return r.weg().Rueckkehren(ctx)
}

func (r ruecksprungnaht) RueckkehrQuittieren(ctx context.Context, kennung int64) error {
	return r.weg().RueckkehrQuittieren(ctx, kennung)
}
