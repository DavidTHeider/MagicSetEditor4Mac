package app

// **Ob eine Sitzung lebt, sagt ihr gehaltener Anspruch** — Befund 11 des
// Zweitblicks über Phase 1 von 0.9.0 (F04, #2343).
//
// Eine Prozessnummer wird wiederverwendet: eine Marke mit der eigenen
// verschwand ohne Hinweis, und bei einer fremden genügte irgendein lebender
// Prozess, damit ein Abbruch nie gemeldet wurde. Jetzt hält jede Sitzung eine
// Sperre (`flock`, `internal/dateisperre`) auf `fenster-<pid>.sperre`, und das
// Betriebssystem gibt sie mit dem Prozess frei — auch nach SIGKILL.
//
// **Was das nicht kann**: Windows kennt `flock` nicht (`dateisperre`), und
// dort hielte jeder Start eine Sitzung mit Anspruch für abgebrochen. Gebaut
// wird dafür nicht (§12).

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/dateisperre"
)

// ordnerfrist ist, wie lange ein Beginn oder ein Ende auf den Ordner wartet.
const ordnerfrist = 5 * time.Second

// ordnersperrpfad ist die Sperre über den ganzen Ordner der Sitzungen.
func ordnersperrpfad(ordner string) string { return filepath.Join(ordner, "sitzungen.sperre") }

// ordnerHalten nimmt die Sperre über den Ordner der Sitzungen — Nachlauf über
// Phase 1 von 0.9.0, Befund 9. **Unter ihr geschieht alles, was Marken liest
// oder legt**: das Lesen und Prüfen fremder Marken, das Räumen eines Abbruchs,
// der eigene Anspruch und die eigene Marke, und am Ende ihr Entfernen. Ohne
// sie räumte ein Start, der eine alte Marke tot fand, unter denselben Pfaden
// die **neue** eines Fensters, das mit der wiederverwendeten Nummer
// dazwischen begann; und ein Ende zwischen Lesen und Prüfen sah wie ein
// Abbruch aus. Die Datei bleibt liegen — eine Sperre auf eine Datei, die ein
// anderer gerade löscht, sperrte nichts.
func ordnerHalten(ordner string, frist time.Duration) (*dateisperre.Griff, error) {
	griff, err := dateisperre.Oeffnen(ordnersperrpfad(ordner))
	if err != nil {
		return nil, fmt.Errorf("den Ordner der Sitzungen sperren: %w", err)
	}

	if err := griff.AlleinBis(context.Background(), frist, 10*time.Millisecond); err != nil {
		_ = griff.Schliessen() //nolint:errcheck // die Sperre kam nicht zustande

		return nil, fmt.Errorf("den Ordner der Sitzungen sperren: %w", err)
	}

	return griff, nil
}

// anspruchspfad ist die Sperrdatei einer Sitzung.
func anspruchspfad(ordner string, pid int) string {
	return filepath.Join(ordner, fmt.Sprintf("fenster-%d.sperre", pid))
}

// anspruchHalten nimmt die Sperre dieser Sitzung und hält sie, bis der Griff
// zugeht.
func anspruchHalten(ordner string, pid int) (*dateisperre.Griff, error) {
	griff, err := dateisperre.Oeffnen(anspruchspfad(ordner, pid))
	if err != nil {
		return nil, fmt.Errorf("den Anspruch der Sitzung anlegen: %w", err)
	}

	// **Ein Erbe ist keine zweite Sitzung** (#2401).
	if err := griff.AlleinOhneErbe(context.Background()); err != nil {
		_ = griff.Schliessen() //nolint:errcheck // der Anspruch kam nicht zustande

		return nil, fmt.Errorf("den Anspruch der Sitzung halten: %w", err)
	}

	return griff, nil
}

// sitzungLebt sagt, ob die Sitzung einer Marke noch läuft.
//
// **Mit Anspruch entscheidet die Sperre**: frei heißt, ihr Prozess ist fort,
// gleich ob ein anderer seine Nummer trägt. **Ohne** — eine Marke aus der Zeit
// davor — entscheidet die Nummer, **nur nie die eigene**: sie kann nur einer
// früheren Sitzung gehören, denn diese legt ihre Marke erst danach.
//
// **Die Zeit davor ist kurz**: Marken gibt es seit F04, und keine Freigabe
// hat sie ohne Anspruch geschrieben — 0.8.0 kennt sie nicht, und vor 0.9.0
// kam der Anspruch dazu. Nur ein Zwischenstand der Entwicklung hinterlässt
// eine, und für sie bleibt die Nummer die Grenze, die sie war.
func sitzungLebt(ordner string, marke sitzungsmarke, pid int, lebt func(int) bool) bool {
	if !marke.Gesperrt {
		return marke.PID != pid && lebt(marke.PID)
	}

	griff, err := dateisperre.Oeffnen(anspruchspfad(ordner, marke.PID))
	if err != nil {
		return false
	}

	defer func() { _ = griff.Schliessen() }() //nolint:errcheck // nur gefragt

	// **Mit einer kurzen Frist und nicht einmalig**: ein `flock` hängt an der
	// offenen Dateibeschreibung, und ein Kind, das irgendein Prozess gerade
	// startet, hält sie zwischen `fork` und `exec` mit — gemessen am
	// 2026-10-14 in 672 von 20.000 Versuchen. Eine lebende Sitzung hält ihre
	// Sperre die ganze Zeit; ein Erbe geht nach Millisekunden.
	return griff.AlleinOhneErbe(context.Background()) != nil
}
