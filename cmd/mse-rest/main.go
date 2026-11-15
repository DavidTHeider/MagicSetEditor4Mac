// Der REST-Dienst auf der Rückschleife als eigener Prozess.
//
// # Warum ein eigenes Programm
//
// **Aus denselben zwei Gründen wie beim MCP-Server**
// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)): ein
// Dienst im Fenster liefe nur, solange jemand das Fenster offen hat — und ein
// Dienst, den man einschalten muss, ohne dass es einen Schalter gibt, ist
// einer, den man vergisst.
//
// **Und weil die Netzgrenze dann eine Grenze bleibt.** Der Wächter in
// `internal/schichten` führt, wer `net/http` erreichen darf; mit einem
// eigenen Programm sind das `internal/rest` und dieses Verzeichnis. Läge der
// Dienst in `cmd/mse`, dürfte die **ganze** Kommandozeile ins Netz — auch die
// Befehle, die damit nichts zu tun haben.
//
// # Er startet nicht von selbst
//
// Kein Aufruf aus dem Fenster, kein Eintrag beim Anmelden, keine Automatik.
// Wer keinen Dienst will, hat keinen — das ist Bedingung 1 aus
// [0046](../../docs/entscheidungen/0046-doch-ein-dienst-auf-der-rueckschleife.md),
// von der anderen Seite gelesen.
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
	// **⌃C beendet den Dienst**, und beim zweiten Signal endet der Prozess:
	// `NotifyContext` gibt die Behandlung danach ans Betriebssystem zurück.
	ctx, aufraeumen := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)

	// **Geschwärzt und mit Kennung** — Befund 7 des Zweitblicks über Phase
	// 6.4.
	slog.SetDefault(app.Serverprotokoll(os.Stderr))

	err := lauf(ctx, os.Args[1:], os.Stdout)

	// **Das Gerätemodell endet mit dem Prozess** — AI-05.AK-07 (#2042):
	// `fm serve` ist ein Kind, das sonst weiterliefe.
	app.GeraetBeenden()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		aufraeumen()
		os.Exit(endcode(err))
	}

	aufraeumen()
}
