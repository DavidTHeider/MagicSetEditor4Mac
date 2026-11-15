// Der MCP-Server als eigener Prozess.
//
// # Warum ein eigenes Programm
//
// **Weil ein Server im Fenster einen Port bräuchte**
// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)). Die
// Standardeingabe des Fensters gehört dem Fenster; über sie kann kein MCP-
// Gegenüber sprechen. Ein Port auf `localhost` machte die Anwendung zu einem
// Netzwerkdienst — und `docs/ROADMAP.md` schließt das aus.
//
// Aufgerufen wird dieses Programm vom Gegenüber selbst, das es als Unterprozess
// startet und über stdio bedient.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

func main() {
	// **⌃C und SIGTERM beenden den Server** — seit N6 (#1572), weil er über
	// HTTP lange lebt und ein Fenster ihn beenden können muss. Über stdio
	// endet er mit dem Rohr — und seit BK-07c (#2117) auch auf das Signal,
	// wenn niemand auflegt; bis dahin wartete er dort auf die nächste Zeile.
	ctx, aufraeumen := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)

	// **Geschwärzt und mit Kennung** — Befund 7 des Zweitblicks über Phase
	// 6.4. Auf die Fehlerausgabe: die Standardausgabe gehört dem Protokoll.
	slog.SetDefault(app.Serverprotokoll(os.Stderr))

	err := lauf(ctx, os.Args[1:], os.Stdin, os.Stdout)

	// **Das Gerätemodell endet mit dem Prozess** — AI-05.AK-07 (#2042):
	// `fm serve` ist ein Kind, das sonst weiterliefe.
	app.GeraetBeenden()

	if err != nil {
		// **Auf die Fehlerausgabe, nicht auf die Standardausgabe.** Dort läuft
		// das Protokoll; ein Satz Prosa dazwischen macht die Sitzung kaputt.
		fmt.Fprintln(os.Stderr, err)
		// **Aufgeräumt wird vor dem Ende und nicht im `defer`** —
		// `os.Exit` lässt keines laufen, und `gocritic` sagt es.
		aufraeumen()
		os.Exit(1)
	}

	aufraeumen()
}
