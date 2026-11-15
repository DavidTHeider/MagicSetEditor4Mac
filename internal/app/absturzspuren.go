package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// absturzspurHoechstens ist, wie viel einer Spur ins Diagnosepaket geht — der
// **Anfang**, denn dort steht, was abstürzte; der Stapel dahinter wiederholt
// sich.
const absturzspurHoechstens = 64 << 10

// absturzspuren gibt die Spuren mit Inhalt, die neueste zuerst, je mit
// einem Kopf — für den Protokollteil des Diagnosepakets (F04, #2343).
// **Ohne Startmarke keine**: dann gibt es keinen Ort, an dem sie lägen.
func (a *App) absturzspuren() string {
	ordner := a.lauf.startmarke.ordner
	if ordner == "" {
		return ""
	}

	pfade, _ := filepath.Glob(filepath.Join(ordner, "absturz-*.txt")) //nolint:errcheck // ein festes Muster

	// **Die neueste zuerst**: der Name trägt die Zeit in Nanosekunden.
	sort.Sort(sort.Reverse(sort.StringSlice(pfade)))

	var spuren strings.Builder

	for _, pfad := range pfade {
		roh, err := leseHoechstens(pfad, absturzspurHoechstens)
		stand, statfehler := os.Stat(pfad)

		if err != nil || statfehler != nil || len(roh) == 0 {
			continue
		}

		// **Der Kopf nennt die Zeit, nicht den Dateinamen**: Nanosekunden und
		// Prozessnummer nimmt der Schwärzer für eine Kennung, und übrig bliebe
		// „«geschwärzt».txt" — gemessen am nativen Fenster (wiederanlauf09).
		// **Ohne freien Paniktext** (Zweitblick über Phase 1 von 0.9.0,
		// Befund 7) — was bleibt, steht in `absturzspurgestalt.go`.
		fmt.Fprintf(&spuren, "\n--- Absturzspur, geschrieben %s ---\n%s",
			stand.ModTime().UTC().Format(time.RFC3339), spurFuerDasPaket(string(roh)))
	}

	return spuren.String()
}

// leseHoechstens liest den Anfang einer Datei, höchstens `grenze` Bytes.
func leseHoechstens(pfad string, grenze int64) ([]byte, error) {
	// #nosec G304 -- ein Pfad aus Glob im Ordner der Sitzungen
	datei, err := os.Open(pfad) //nolint:gosec // dito
	if err != nil {
		return nil, err //nolint:wrapcheck // beide Aufrufer verwerfen den Grund
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // nur gelesen

	return io.ReadAll(io.LimitReader(datei, grenze)) //nolint:wrapcheck // dito
}
