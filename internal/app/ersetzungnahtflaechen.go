// **Suchen und Ersetzen für die Flächen ohne Fenster** — Werkstück 2 von
// Phase 3.7
// (#1168).
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort. **Hier gibt es keinen
// Grund für eine Ausnahme**: die Frage ist Text, die Antwort ist eine Liste,
// und beides trägt über jede Fläche.

package app

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Fremdersetzung ist das Ersetzen für eine Fläche ohne Fenster.
type Fremdersetzung struct{ quelle Ersetzungsquelle }

// Ersetzungsflaechennaht baut sie über einem Bestand.
//
// **Über [Schreibnaht] und nicht daneben** — dieselbe Begründung wie bei
// [Ablageflaechennaht]: sie baut Aufsicht, Feldweg und **Stilwähler** an einer
// Stelle, und der Stilwähler ist es, der die Feldarten kennt.
func Ersetzungsflaechennaht(db *sql.DB, urheber domain.Urheber) (*Fremdersetzung, error) {
	pakete, err := Paketverzeichnis()
	if err != nil {
		return nil, err
	}

	// **Kein Umweg über [Schreibnaht].** Der erste Anlauf rief sie, nur um
	// ihre Fehlerprüfung zu borgen, und warf das Ergebnis weg — dahinter stand
	// ein zweites `Paketverzeichnis()`, dessen Fehlerzweig damit **kein Lauf**
	// mehr erreichen konnte (§3). Was diese Naht braucht, ist der Stilwähler,
	// und der kommt aus [stilwaehlerMit]: **einer** Verdrahtung, geteilt mit
	// `Schreibnaht` und `SetangabennahtUeber`.
	return &Fremdersetzung{quelle: service.NeuerErsetzer(
		storage.NeuerSetLeser(db), stilwaehlerMit(db, pakete),
		storage.NeuerKartenschreiber(db).Als(urheber))}, nil
}

// MitErsetzungsquelle setzt die Naht — für Tests.
func (f *Fremdersetzung) MitErsetzungsquelle(quelle Ersetzungsquelle) *Fremdersetzung {
	f.quelle = quelle

	return f
}

// Vorschau sagt, was sich ändern würde.
func (f *Fremdersetzung) Vorschau(
	ctx context.Context, setID int64, frage domain.Ersetzungsfrage,
) (domain.Ersetzungsvorschau, error) {
	return f.quelle.Vorschau(ctx, setID, frage)
}

// Ausfuehren ersetzt und gibt dieselbe Übersicht zurück.
func (f *Fremdersetzung) Ausfuehren(
	ctx context.Context, setID int64, frage domain.Ersetzungsfrage,
) (domain.Ersetzungsvorschau, error) {
	return f.quelle.Ausfuehren(ctx, setID, frage)
}
