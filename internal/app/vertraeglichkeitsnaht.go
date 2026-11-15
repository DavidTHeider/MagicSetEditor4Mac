package app

// **Die Nähte der Verträglichkeitsprüfung** — Werkstück O6 von Phase 5.2
// (#1430).
//
// Vier Fragen, vier kleine Nähte: was der Bestand verbucht hat, wie viel auf
// seinem Volumen frei ist, wie viel gerade läuft — und ob sich eine frische
// Sicherung **zurücklesen** lässt.

import (
	"context"
	"database/sql"
	"path/filepath"
	"runtime"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// ledgernaht liest die verbuchten Wanderungen.
type ledgernaht struct{ db *sql.DB }

func (l ledgernaht) Ledger(ctx context.Context) (map[string]string, error) {
	return storage.Ledger(ctx, l.db)
}

// LedgerUeber baut die Naht über einer offenen Datenbank.
func LedgerUeber(db *sql.DB) service.Ledgerquelle { return ledgernaht{db: db} }

// platznaht sagt, wie viel neben dem Bestand frei ist.
type platznaht struct{ pfad string }

func (p platznaht) FreierPlatz() (int64, error) { return freierPlatz(filepath.Dir(p.pfad)) }

// PlatzNeben baut die Naht über dem Verzeichnis eines Bestands.
func PlatzNeben(bestandspfad string) service.Platzquelle {
	return platznaht{pfad: bestandspfad}
}

// vorgangsnaht zählt, was gerade läuft.
//
// **Aus dem Auftragsbuch** (U11): dort steht, was das Fenster angestoßen hat
// und noch nicht fertig ist. Ein Tausch mitten in einer Einfuhr ließe halbe
// Arbeit zurück.
type vorgangsnaht struct{ anwendung *App }

// **Ohne Fehlerzweig, und das ist nachgelesen**: [App.Auftraege] liest ein
// Buch im Arbeitsspeicher und setzt sein Fehlerfeld nie. Ein `if` darauf wäre
// eine Zeile, die kein Lauf betreten kann — und eine unbetretbare Zeile ist
// keine Absicherung, sondern eine Lücke in der Abdeckung mit einer Ausrede
// davor (§3).
func (v vorgangsnaht) LaufendeVorgaenge(ctx context.Context) (int, error) {
	return v.anwendung.Auftraege(ctx).Laufend, nil
}

// SicherungsprobeUeber baut die Naht über einem Bestandsweg.
//
// **Der Rumpf liegt in `internal/service`** (`SicherungsprobeUeber` dort):
// `internal/app` darf `internal/sicherung` nicht kennen (§2), und wohin die
// Probe geschrieben und wie sie zurückgelesen wird, ist ohnehin dieselbe
// Sorte Entscheidung wie bei der Bestandssicherung daneben. Hier bleibt die
// Zeile, die die zwei Schichten verbindet.
func SicherungsprobeUeber(weg *service.Bestandsweg) service.Sicherungsprobe {
	return service.SicherungsprobeUeber(weg.Probesichern)
}

// VertraeglichkeitUeber baut die Prüfung für eine Fläche **ohne Fenster**.
//
// **Nur die Fassungsfrage.** Die drei Fragen des Augenblicks — Platz,
// laufende Arbeit, eine zurückgelesene Sicherung — gehören zum Vorbereiten,
// und das tut nur das Fenster (`App.WechselVorbereiten`). Was hier
// herauskommt, ist dieselbe Antwort wie dort auf dieselbe Frage: kann die
// neueste Freigabe **diesen** Bestand öffnen?
func VertraeglichkeitUeber(db *sql.DB) *App {
	anwendung := NeueApp("")
	anwendung.MitFassungen(FassungsnahtMit(anwendung.Proxyquelle()))
	anwendung.MitVorpruefer(service.NeuerVorpruefer(
		LedgerUeber(db), runtime.GOOS+"/"+runtime.GOARCH))

	// **Und die Einstellungen dazu.** Ohne sie fragte diese Fläche nie ins
	// Netz — was sicher, aber falsch wäre: die Wahl steht im Bestand, und
	// §16 verlangt an jeder Fläche dieselbe Antwort.
	anwendung.MitEinstellungen(einstellungsverwalterUeber(db))

	return anwendung
}
