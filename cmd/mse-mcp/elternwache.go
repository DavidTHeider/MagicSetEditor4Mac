// **Wenn das Fenster stirbt, endet auch dieser Dienst** — Werkstück N6 von
// Phase 5.4
// (#1572).
//
// # Warum es das braucht
//
// **Weil ein Unterprozess seinen Elternprozess überlebt.** Beim gewöhnlichen
// Beenden nicht — das Fenster schickt ein Signal und wartet. Bei einem
// **Absturz** schon, und dann stünde ein Port offen, den niemand mehr
// schließt und von dem niemand mehr weiß.
//
// # Wie es gemessen wird
//
// **Am Rohr und nicht an der Prozesskennung.** Das Fenster hält die
// Standardeingabe dieses Prozesses; stirbt es, schließt das Betriebssystem
// sie, und ein Lesen darauf gibt sofort EOF. Eine Wache über `getppid()`
// müsste pollen und verpasste das Fenster zwischen zwei Blicken.
//
// # Und sie wird nur aufgezogen, wo wirklich ein Rohr hängt
//
// **Gemessen, nicht angenommen.** Der erste Anlauf zog sie über jede
// Standardeingabe auf — und unter `go test` hängt dort `/dev/null`, das
// **sofort** EOF gibt. Der Dienst beendete sich, bevor die erste Anfrage
// ankam: „connection reset by peer" in zwei Läufen, die mit der Wache nichts
// zu tun hatten.
//
// **Aufgezogen wird deshalb nur über einem Rohr** (`os.ModeNamedPipe`). Ein
// Endgerät und `/dev/null` sind Zeichengeräte, eine Umleitung aus einer Datei
// ist eine gewöhnliche Datei — in allen drei Fällen sagt ein EOF nichts über
// den Elternprozess. Wer das Programm von Hand startet, behält es also; es
// gehört ihm und keinem Fenster.

package main

import (
	"context"
	"io"
	"os"
)

// elternwache endet, sobald die Standardeingabe schließt.
//
// **Sie liest und verwirft.** Was hereinkäme, ist nichts — das Rohr ist da,
// um zu **schließen**, nicht um zu sprechen.
func elternwache(ctx context.Context, eingabe io.Reader, aufhoeren context.CancelFunc) {
	go func() {
		_, _ = io.Copy(io.Discard, eingabe) //nolint:errcheck // das Ende ist das Ereignis

		// **Nur, wenn nicht ohnehin schon Schluss ist.** Sonst meldete ein
		// gewöhnliches Beenden, das Fenster sei gestorben.
		if ctx.Err() == nil {
			aufhoeren()
		}
	}()
}

// elternwacheAufziehen ist [elternwache] über der echten Standardeingabe —
// **wenn dort ein Rohr hängt**.
//
// **Der Rückgabewert sagt, ob sie steht.** Ohne ihn wäre „die Wache greift"
// eine Behauptung, die kein Lauf prüfen kann.
func elternwacheAufziehen(ctx context.Context, aufhoeren context.CancelFunc) bool {
	angaben, fehler := os.Stdin.Stat()
	if fehler != nil || angaben.Mode()&os.ModeNamedPipe == 0 {
		return false
	}

	elternwache(ctx, os.Stdin, aufhoeren)

	return true
}
