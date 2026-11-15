package service

// **Ein Diagnosepaket auf die Platte schreiben** — Werkstück O10 von Phase 5.2
// (#1434).
//
// # Warum das hier steht und nicht in `internal/app`
//
// **Weil `internal/app` `internal/sicherung` nicht kennen darf** (§2), und der
// Schichtenwächter hat es angestrichen. Der erste Anlauf packte das Archiv in
// der Fassade zusammen; das ist dieselbe Sorte Entscheidung wie „wohin wird
// geschrieben und was passiert bei einem Abbruch" — und die gehört in den
// Dienst, wo `SchreibeAtomar` und die Bestandssicherung schon liegen.
//
// **Was in `app` bleibt**, ist das Zusammenstellen: welche Teile es gibt, was
// darin steht und was vorher gezeigt wird. Hierher kommen fertige Bytes.

import (
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// Paketteil ist ein fertiger Teil eines Diagnosepakets.
type Paketteil struct {
	// Name ist der Weg im Archiv, Inhalt der Text.
	Name   string
	Inhalt []byte
}

// SchreibeDiagnosepaket legt die Teile samt Manifest nach `ziel`.
//
// **Unteilbar und ohne Verdrängung**: das Archiv entsteht als Nachbardatei und
// bekommt seinen Namen erst, wenn es vollständig ist — ein Abbruch lässt kein
// scheinbar fertiges Paket zurück. Und es **nimmt kein bestehendes weg**; der
// Grund steht bei [sicherung.PackenAls].
func SchreibeDiagnosepaket(
	ziel, manifest, fassung string, teile []Paketteil, jetzt time.Time,
) error {
	archiv := make([]sicherung.Teil, 0, len(teile))
	for _, teil := range teile {
		archiv = append(archiv, sicherung.Teil{Name: teil.Name, Inhalt: teil.Inhalt})
	}

	if fehler := sicherung.PackenAls(ziel, manifest, archiv,
		sicherung.Angaben{Fassung: fassung, Erstellt: jetzt}); fehler != nil {
		return fmt.Errorf("das Diagnosepaket schreiben: %w", fehler)
	}

	return nil
}
