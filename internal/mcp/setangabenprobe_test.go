// **Das Double für die Setangaben** (#1107).
//
// **Ein eigenes und nicht in `bestandsprobe`**: die Setangaben kommen im
// Betrieb aus einer anderen Naht — der Stilwähler kennt die Felder, der
// Bestand die Werte —, und ein Double, das beides in einem Verbund führte,
// verdeckte genau diese Trennung.
package mcp_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// angabenprobe antwortet auf alle vier Werkzeuge.
type angabenprobe struct {
	werte  map[string]string
	fehler error
}

// neueAngabenprobe gibt eine Probe mit einem gesetzten Titel.
func neueAngabenprobe() *angabenprobe {
	return &angabenprobe{werte: map[string]string{"title": "Probeset"}}
}

func (a *angabenprobe) Angaben(context.Context, int64) (
	[]mcp.Angabenfeld, []mcp.Angabenwert, error,
) {
	if a.fehler != nil {
		return nil, nil, a.fehler
	}

	werte := make([]mcp.Angabenwert, 0, len(a.werte))
	for _, feld := range []string{"title", "set code"} {
		if wert, da := a.werte[feld]; da {
			werte = append(werte, mcp.Angabenwert{Feld: feld, Wert: wert})
		}
	}

	return []mcp.Angabenfeld{
		{Name: "title", Aufschrift: "Titel", Art: "text"},
		{Name: "set code", Aufschrift: "Setcode", Art: "text"},
	}, werte, nil
}

// Setzen schreibt und gibt den vorigen Wert.
//
// **Es schreibt wirklich**: die Idempotenzprobe ruft zweimal und vergleicht,
// und ein Double, das nichts merkt, machte jedes schreibende Werkzeug
// idempotent.
func (a *angabenprobe) Setzen(_ context.Context, _ int64, feld, wert string) (
	vorher string, hatte bool, fehler error,
) {
	if a.fehler != nil {
		return "", false, a.fehler
	}

	vorher, hatte = a.werte[feld]
	a.werte[feld] = wert

	return vorher, hatte, nil
}

func (a *angabenprobe) Vorbehalte(context.Context, int64) ([]mcp.Vorbehalt, error) {
	if a.fehler != nil {
		return nil, a.fehler
	}

	return []mcp.Vorbehalt{{Datei: "set", Zeile: 12, Inhalt: "unbekannt: x"}}, nil
}

func (a *angabenprobe) Beispielsets() ([]mcp.Beispielset, error) {
	if a.fehler != nil {
		return nil, a.fehler
	}

	return []mcp.Beispielset{{Name: "Probe", Pfad: "/bündel/saat/Probe.mse-set"}}, nil
}

// werkzeugMitAngaben sucht ein Werkzeug über einer Angabenprobe.
func werkzeugMitAngaben(t *testing.T, probe *angabenprobe, name string) mcp.Werkzeug {
	t.Helper()

	for _, werkzeug := range mcp.Werkzeuge(mcp.Naehte{
		Leser:     probenbestand(),
		Schreiber: probenbestand(),
		Ausfuhr:   ruhendeAusfuhr{},
		Auskunft:  stilleAuskunft{}, Sammeldaten: stilleSammeldaten{},
		Angaben: probe,
	}) {
		if werkzeug.Name == name {
			return werkzeug
		}
	}

	t.Fatalf("es gibt kein Werkzeug %q", name)

	return mcp.Werkzeug{}
}
