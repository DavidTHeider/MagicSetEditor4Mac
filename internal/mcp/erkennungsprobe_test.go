package mcp_test

// **Ein Doppel mit einem Kandidaten ohne Stütze** — Werkstück A13.3 von Phase
// 5.3 (#1452).
//
// # Warum es ein eigenes gibt
//
// **Weil `nilprobe` eine Ebene zu hoch aufhört.** Sie gibt gar keinen
// Kandidaten, und die Frage „bleibt `gestuetzt` **in** einem Kandidaten eine
// Liste?" lässt sich daran nicht stellen. Genau diese Falle ist in A13.1b und
// A13.2b je einmal zugeschnappt.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// ohneStuetze gibt einen Kandidaten, den nichts stützt.
type ohneStuetze struct{ nilprobe }

func (ohneStuetze) KIErkennung(context.Context, mcp.KIErkennungsbezug) (
	mcp.KIErkennungssicht, error,
) {
	return mcp.KIErkennungssicht{
		Kandidaten: []mcp.KIBildkandidatsicht{{
			Kennung: "lea-174", Name: "Shivan Dragon",
		}},
		Drucke: 72,
	}, nil
}
