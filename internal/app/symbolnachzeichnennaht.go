// **Ein Bild als Symbol nachzeichnen — die Naht für die kopflosen Flächen** —
// Werkstück F von Phase 4.11
// (#1340).
//
// # Warum sie beides tut
//
// Am Fenster sind Nachzeichnen und Ablegen zwei Handgriffe: die Teile stehen
// erst in der Zeichenfläche, weil jemand sie noch anfassen will. Kopflos gibt
// es niemanden, der sie hält — wer eine Zeichnung zurückbekäme, ohne sie
// ablegen zu können, hätte eine halbe Fähigkeit (§16).

package app

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Symbolnachzeichnungsnaht zeichnet nach und legt ab.
type Symbolnachzeichnungsnaht struct {
	anwendung *App
	db        *sql.DB
	wache     *service.Autorisierung
}

// SymbolnachzeichnungUeber baut die Naht über einem Bestand.
//
// **Der Bestand wird gebraucht** — seit #1815. Hier stand, er werde es
// nicht, weil die Symbolablage am Bilderverzeichnis hängt. Für die Datei
// stimmt das; **wem das Set gehört, weiß aber nur der Bestand**, und ohne ihn
// zeichnete jeder in jedes Set, dessen Kennung er nannte.
func SymbolnachzeichnungUeber(db *sql.DB) (*Symbolnachzeichnungsnaht, error) {
	bilder, fehler := Bilderverzeichnis()
	if fehler != nil {
		return nil, fehler
	}

	anwendung := &App{}
	anwendung.MitSymbolablage(service.NeueSymbolablage(service.NeueBildablage(bilder)))

	return &Symbolnachzeichnungsnaht{anwendung: anwendung, db: db, wache: AutorisierungAus(db)}, nil
}

// Nachzeichnen liest das Bild und legt die Zeichnung in ein Set des
// gezeigten Profils.
func (n *Symbolnachzeichnungsnaht) Nachzeichnen(
	ctx context.Context, setID int64, bildpfad string,
) (any, error) {
	// **Davor die Rolle** (C04, #2229) — aus demselben Grund wie in der
	// Werkstatt: abgelegt wird eine Datei, an der `query_only` nichts hält.
	if fehler := n.wache.Pruefe(ctx, domain.WirkungSchreiben); fehler != nil {
		return nil, fehler
	}

	// **Zuerst, wem das Set gehört** (#1815) — vor dem Bild und vor dem
	// Ablegen, damit ein fremdes Set nicht erst als Konflikt zurückkommt.
	if fehler := SetDaSein(ctx, n.db, setID); fehler != nil {
		return nil, fehler
	}

	teile, err := Nachgezeichnet(bildpfad)
	if err != nil {
		return nil, err
	}

	abgelegt := n.anwendung.SymbolSetzen(ctx, strconv.FormatInt(setID, 10), teile, "")
	if abgelegt.Fehler != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrConflict, abgelegt.Fehler.Meldung)
	}

	punkte := 0
	for _, teil := range abgelegt.Teile {
		punkte += len(teil.Punkte)
	}

	// **Die Zahlen und nicht die Zeichnung.** Wer die Punkte sehen will, liest
	// sie mit `Symbol`; hier geht es darum, ob etwas Brauchbares entstanden ist.
	return map[string]any{
		"name":   abgelegt.Name,
		"teile":  len(abgelegt.Teile),
		"punkte": punkte,
	}, nil
}
