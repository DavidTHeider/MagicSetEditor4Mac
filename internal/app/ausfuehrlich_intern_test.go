package app

// **Die ausführliche Stufe schreibt etwas** — Befund 5 des Zweitblicks über
// Phase 6.4 (#1665):
// im ganzen Produktcode stand keine Zeile auf ihr, und die Diagnose sagte
// trotzdem „ausführlich an".
//
// **Ohne `t.Parallel()`**: der Test setzt den Schreiber des Prozesses.

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestJederRufDesFenstersStehtAusfuehrlichImProtokoll(t *testing.T) {
	vorher := slog.Default()
	t.Cleanup(func() { slog.SetDefault(vorher) })

	puffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(&geschwaerzt{innen: slog.NewTextHandler(puffer,
		&slog.HandlerOptions{Level: slog.LevelDebug})}))

	anwendung := NeueApp("")
	anwendung.Metriken(t.Context())

	zeile := puffer.String()
	if !strings.Contains(zeile, "level=DEBUG") || !strings.Contains(zeile, "(*App).Metriken") ||
		!strings.Contains(zeile, "kennung=") {
		t.Errorf("%q", zeile)
	}

	// **Auf der üblichen Stufe steht nichts** — die Zeile kostet sonst nichts.
	puffer.Reset()
	slog.SetDefault(slog.New(slog.NewTextHandler(puffer, nil)))
	anwendung.Metriken(t.Context())

	if puffer.Len() != 0 {
		t.Errorf("auf »info«: %q", puffer.String())
	}
}

// **Was das Fenster im Takt fragt, steht nicht darin** — Nachlauf über Phase
// 6.4, N5: `Datenstand` alle zwei Sekunden gab rund 43.000 Zeilen am Tag, und
// die ausführliche Stufe hält höchstens 1 MB.
func TestEineAbfrageImTaktStehtNichtAusfuehrlichImProtokoll(t *testing.T) {
	vorher := slog.Default()
	t.Cleanup(func() { slog.SetDefault(vorher) })

	puffer := &bytes.Buffer{}
	slog.SetDefault(slog.New(&geschwaerzt{innen: slog.NewTextHandler(puffer,
		&slog.HandlerOptions{Level: slog.LevelDebug})}))

	anwendung := NeueApp("")
	anwendung.MitStand(festerStand(7))
	anwendung.Datenstand(t.Context())

	if strings.Contains(puffer.String(), "(*App).Datenstand") {
		t.Errorf("%q", puffer.String())
	}
}

type festerStand int64

func (s festerStand) Datenstand(context.Context) (int64, error) { return int64(s), nil }
