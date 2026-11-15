// **Der Riegel vor jedem Weg mit Datei- oder Netzwirkung** — Werkstück O1
// von Phase 5.2 (#1425),
// nach [Vermerk 0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md).
//
// # Warum es ihn gibt
//
// **`query_only` ersetzt keine Autorisierung.** Ein Leser liest über eine
// Verbindung, die nichts annimmt (0073) — aber ein Druckbogen, ein Kartenbild,
// eine Kopie des Sets und MTG JSON schreiben **Dateien**, und die Platte
// fragt den Bestand nicht. Bis zu diesem Werkstück ging jeder dieser Wege für
// einen Leser durch.
//
// # Wer ihn setzt, und was passiert, wenn niemand es tut
//
// **Die Wurzeln setzen ihn** — `internal/app` reicht die [Autorisierung] an
// jeden Dienst, der schreibt, für das Fenster wie für die drei Flächen ohne
// Fenster. **Ohne Prüfer prüft niemand**: ein Dienst, den ein Test ohne einen
// baut, prüft seine Sache und nicht die Erlaubnis. Dass jede Wurzel ihn
// wirklich setzt, belegen die Läufe am gestarteten Fenster
// (`internal/app/leserausfuhr_test.go`) und an der gebauten Kommandozeile
// (`internal/flaechen/leserausfuhr_test.go`) — nicht dieser Riegel selbst.

package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Pruefer sagt, ob dieser Aufruf eine Wirkung haben darf — die
// [Autorisierung] ist einer.
type Pruefer interface {
	Pruefe(ctx context.Context, wirkung domain.Wirkung) error
}

// wache trägt den Prüfer eines Dienstes.
type wache struct{ pruefer Pruefer }

// darfSchreiben fragt den Prüfer nach `lokalSchreiben` — oder niemanden,
// solange keiner gesetzt ist. **Nur diese eine Wirkung**: die Dienste hier
// schreiben Dateien; wer nach draußen sendet, fragt an der Netzgrenze.
func (w wache) darfSchreiben(ctx context.Context) error {
	if w.pruefer == nil {
		return nil
	}

	return w.pruefer.Pruefe(ctx, domain.WirkungSchreiben)
}
