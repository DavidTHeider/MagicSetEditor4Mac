package app

// **Die Reste vor dem Öffnen** — OPS04
// (#2350),
// Befund A-27 des Audits 0.9.0. Was ein eigener Rest ist, entscheidet
// `service.RestePflegen`; hier steht, **wann** geräumt werden darf.
//
// # Nur, wenn niemand sonst arbeitet
//
// **Vor dem Öffnen und nur mit dem Alleinanspruch** ([alleinnaht]): hält ein
// anderer Prozess den Bestand — der Dienst, ein `mse`, ein `sqlite3` —, kann
// er gerade in einen Nachbarn schreiben, und dann bleibt alles liegen. Ein
// frischer Bestand hat noch nichts zu räumen.
//
// # Was er nicht kann
//
// **Reste außerhalb der eigenen Ordner** — eine Sicherung ohne Kennwort
// entsteht neben ihrem Ziel, und dorthin greift diese Anwendung nicht.

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// restePflegen räumt die eigenen Reste neben dem Bestand.
func (a *App) restePflegen(ctx context.Context) {
	if _, fehler := os.Stat(a.datenbankpfad); fehler != nil {
		return
	}

	freigeben, fehler := alleinnaht(ctx, a.datenbankpfad)
	if fehler != nil {
		slog.InfoContext(ctx, "Reste bleiben liegen — ein anderer Prozess hält den Bestand", "grund", fehler.Error())

		return
	}

	defer func() { _ = freigeben() }() //nolint:errcheck // der Anspruch fällt mit dem Griff

	ordner := filepath.Dir(a.datenbankpfad)
	orte := service.Restorte{Bestand: a.datenbankpfad}

	// **Beiwerk und Pakete nur am eigenen Ort** (Zweitblick über Phase 4,
	// Befund 1, und sein Nachlauf): am Vorgabeort gehört der Ordner dieser
	// Anwendung. Neben einem Bestand unter `~/Dokumente` belegt nichts, dass
	// ein Ordner `beiwerk` oder `pakete` von hier stammt — dort bleibt ein
	// Rest lieber liegen, als dass eine fremde Datei geht.
	if vorgabe, vorgabefehler := Datenbankpfad(); vorgabefehler == nil && filepath.Dir(vorgabe) == ordner {
		orte.Eigene = []string{filepath.Join(ordner, "beiwerk")}
		orte.Flach = []string{filepath.Join(ordner, "pakete")}
	}

	entfernt, fehler := service.RestePflegen(orte, time.Now())
	if len(entfernt) > 0 {
		slog.InfoContext(ctx, "liegengebliebene Reste geräumt", "anzahl", len(entfernt))
	}

	if fehler != nil {
		slog.WarnContext(ctx, "nicht jeder Rest ließ sich räumen", "grund", fehler.Error())
	}
}
