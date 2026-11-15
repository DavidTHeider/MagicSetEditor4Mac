package service

// **Die Fassungen des Setsymbols eines Sets** — DOM-02.AK-02
// (#1697).
//
// Dieselben zwei Quellen wie beim Zeichnen einer Karte (`stilumgebung.go`):
// das Spiel erklärt sie unter `default set style:`, der Stil ergänzt sie unter
// `set info style:` — **die des Spiels stehen vorn**, und bei gleichem Namen
// gewinnt die erste ([ErsteFassung], #548).

import (
	"context"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Symbolfassungen gibt die Fassungen, in denen das Setsymbol dieses Sets
// gezeichnet wird.
//
// **Ein Spiel oder Stil, das nicht dasteht, trägt keine bei** — und das ist
// kein Fehler: die Vorschau sagt dann, dass es keine gibt. Ein Fehler hier
// nähme die schlichte Vorschau gleich mit.
func (w *Stilwaehler) Symbolfassungen(ctx context.Context, setID int64) ([]msepaket.Symbolfassung, error) {
	kopf, err := w.kopf(ctx, setID)
	if err != nil {
		return nil, err
	}

	fassungen := w.spielsymbolfassungen(kopf.Game)

	if stil, fehler := w.stil(kopf.Style, kopf.Game); fehler == nil {
		fassungen = append(fassungen, stil.Symbolfassungen...)
	}

	return fassungen, nil
}

// spielsymbolfassungen liest die Fassungen aus dem Spielpaket — leer, wenn es
// keines gibt.
func (w *Stilwaehler) spielsymbolfassungen(spiel string) []msepaket.Symbolfassung {
	name := spiel + "." + string(msepaket.ArtSpiel)

	for _, wurzel := range w.wurzeln {
		geoeffnet, err := msepaket.Oeffne(filepath.Join(wurzel, name))
		if err != nil {
			continue
		}

		fassungen := geoeffnet.Spielsymbolfassungen()
		_ = geoeffnet.Schliessen() //nolint:errcheck // die Fassungen sind vollständig kopiert

		return fassungen
	}

	return []msepaket.Symbolfassung{}
}
