package app

// **Die Marke einer Sitzung und ihre Absturzspur** — F04
// (#2343),
// Befund A-04.
//
// # Was liegt wo
//
// Neben dem Bestand, im Ordner `Sitzungen`: je laufendem Fenster eine Marke
// `fenster-<pid>.json` und eine Spur `absturz-<zeit>-<pid>.txt`. **Die Spur
// schreibt Go selbst**: `debug.SetCrashOutput` gibt die Ausgabe einer
// unbehandelten Panik oder eines `fatal error` zusätzlich in diese Datei
// (nachgelesen in `go doc runtime/debug.SetCrashOutput`, Go 1.27.2). Ein
// SIGKILL hinterlässt keine Spur, nur die Marke.
//
// # Keine Telemetrie
//
// Nichts davon verlässt den Rechner von selbst. Was hinausgeht, geht im
// Diagnosepaket, nach der Vorschau, geschwärzt und nur für den Verwalter.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// absturzspurenHoechstens ist, wie viele Spuren mit Inhalt liegen bleiben —
// eine feste Grenze, damit eine Schleife von Abstürzen die Platte nicht füllt.
const absturzspurenHoechstens = 10

// markeHoechstens ist, wie viel einer Marke gelesen wird — sie trägt vier
// kurze Felder, und was länger ist, ist keine von uns.
const markeHoechstens = 4 << 10

// sitzungsmarke liegt auf der Platte, solange eine Sitzung läuft.
type sitzungsmarke struct {
	PID       int       `json:"pid"`
	Gestartet time.Time `json:"gestartet"`
	Fassung   string    `json:"fassung"`
	Spur      string    `json:"spur"`

	// Gesperrt sagt, dass die Sitzung ihren Anspruch neben der Marke hält
	// (`fenster-<pid>.sperre`) — ob sie lebt, entscheidet dann die Sperre und
	// nicht die Prozessnummer.
	Gesperrt bool `json:"gesperrt,omitempty"`
}

// Abbruch ist eine Sitzung des Fensters, die nicht sauber endete.
type Abbruch struct {
	// Gestartet ist der Beginn der Sitzung — die Nullzeit, wenn ihre Marke
	// sich nicht lesen ließ.
	Gestartet time.Time `json:"gestartet"`
	Fassung   string    `json:"fassung"`

	// MitSpur heißt: Go hat beim Absturz etwas geschrieben. Ohne Spur war es
	// ein Abschuss von außen — SIGKILL, ein leerer Akku, ein erzwungenes Ende.
	MitSpur bool `json:"mitSpur"`
}

// sitzungsbeginn ist, was der Beginn einer Sitzung weiß und tut.
type sitzungsbeginn struct {
	abbrueche []Abbruch

	// ordner ist der Ort von Marken und Spuren — fürs Diagnosepaket.
	ordner string

	// ende räumt Marke und leere Spur — **nur beim sauberen Schluss**.
	ende func() error
}

// sitzungBeginnen liest, welche frühere Sitzung nicht sauber endete, legt die
// eigene Marke und richtet die Spur ein. `lebt` sagt, ob ein Prozess noch
// läuft; `ausgabe` ist `debug.SetCrashOutput` und im Test ersetzt.
func sitzungBeginnen(
	ordner string, pid int, jetzt time.Time, lebt func(int) bool, ausgabe func(*os.File) error,
) (sitzungsbeginn, error) {
	return sitzungBeginnenMit(ordner, pid, jetzt, lebt, ausgabe, ordnerfrist)
}

// sitzungBeginnenMit ist [sitzungBeginnen] mit der Frist für den Ordner.
func sitzungBeginnenMit(
	ordner string, pid int, jetzt time.Time, lebt func(int) bool, ausgabe func(*os.File) error, frist time.Duration,
) (sitzungsbeginn, error) {
	//nolint:gosec // G703: der Ordner der Sitzungen neben dem Bestand, von dieser Anwendung gewählt
	if err := os.MkdirAll(ordner, 0o700); err != nil {
		return sitzungsbeginn{}, fmt.Errorf("den Ordner der Sitzungen anlegen: %w", err)
	}

	ordnergriff, err := ordnerHalten(ordner, frist)
	if err != nil {
		return sitzungsbeginn{}, err
	}

	defer func() { _ = ordnergriff.Schliessen() }() //nolint:errcheck // der Ordner wird frei

	abbrueche := abbruecheLesen(ordner, pid, lebt)
	spurenKuerzen(ordner)

	spurname := fmt.Sprintf("absturz-%d-%d.txt", jetzt.UnixNano(), pid)
	spurpfad := filepath.Join(ordner, spurname)

	// **Ausschließlich neu**: der Name trägt Zeit und Prozesskennung, und eine
	// Spur, die es schon gibt, gehört einer anderen Sitzung.
	// #nosec G304 -- ein Name aus Zeit und Prozesskennung im Ordner der Sitzungen
	spur, err := os.OpenFile(spurpfad, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o600) //nolint:gosec // dito
	if err != nil {
		return sitzungsbeginn{}, fmt.Errorf("die Absturzspur anlegen: %w", err)
	}

	// **SetCrashOutput verdoppelt den Griff**, das Schließen hier nimmt ihm
	// nichts.
	defer func() { _ = spur.Close() }() //nolint:errcheck // nur ein Griff auf eine leere Datei

	if ausgabefehler := ausgabe(spur); ausgabefehler != nil {
		return sitzungsbeginn{}, fmt.Errorf("die Absturzspur einrichten: %w", ausgabefehler)
	}

	// **Erst der Anspruch, dann die Marke** (Zweitblick über Phase 1 von
	// 0.9.0, Befund 11): wer die Marke liest, findet die Sperre schon
	// gehalten — sonst hielte ein zweiter Start, der genau dazwischen liest,
	// eine lebende Sitzung für abgebrochen.
	anspruch, err := anspruchHalten(ordner, pid)
	if err != nil {
		_ = ausgabe(nil) //nolint:errcheck // zurück auf keine Spur

		return sitzungsbeginn{}, err
	}

	roh, _ := json.Marshal(sitzungsmarke{PID: pid, Gestartet: jetzt, Fassung: Fassung, Spur: spurname, Gesperrt: true}) //nolint:errcheck // fünf einfache Felder
	markenpfad := filepath.Join(ordner, fmt.Sprintf("fenster-%d.json", pid))

	if err := service.SchreibeAtomarVertraulich(markenpfad, roh); err != nil {
		_ = ausgabe(nil)          //nolint:errcheck // zurück auf keine Spur
		_ = anspruch.Schliessen() //nolint:errcheck // der Anspruch geht mit

		return sitzungsbeginn{}, fmt.Errorf("die Marke der Sitzung legen: %w", err)
	}

	ende := func() error {
		_ = ausgabe(nil) //nolint:errcheck // der Schluss endet so oder so

		// **Auch das Ende unter dem Ordner**, und ohne ihn trotzdem: ein
		// Schluss bleibt ein Schluss, und was dann fehlt, sagt der Fehler.
		ordnergriff, sperrfehler := ordnerHalten(ordner, frist)
		if sperrfehler == nil {
			defer func() { _ = ordnergriff.Schliessen() }() //nolint:errcheck // der Ordner wird frei
		}

		// **Der Anspruch zuletzt**: erst wenn die Marke fort ist, darf ein
		// anderer Start die Sperre frei finden.
		defer func() {
			_ = os.Remove(anspruchspfad(ordner, pid)) //nolint:errcheck,gosec // eine liegengebliebene Sperrdatei ohne Sperre ist harmlos
			_ = anspruch.Schliessen()                 //nolint:errcheck // dito
		}()

		//nolint:gosec // G703: der Ordner der Sitzungen neben dem Bestand, von dieser Anwendung gewählt
		if stand, statfehler := os.Stat(spurpfad); statfehler == nil && stand.Size() == 0 {
			_ = os.Remove(spurpfad) //nolint:errcheck,gosec // eine leere Spur zu behalten schadet nicht
		}

		//nolint:gosec // G703: der Ordner der Sitzungen neben dem Bestand, von dieser Anwendung gewählt
		if err := os.Remove(markenpfad); err != nil && !errors.Is(err, os.ErrNotExist) {
			return errors.Join(sperrfehler, fmt.Errorf("die Marke der Sitzung entfernen: %w", err))
		}

		return sperrfehler
	}

	return sitzungsbeginn{abbrueche: abbrueche, ordner: ordner, ende: ende}, nil
}

// abbruecheLesen gibt die Sitzungen, deren Marke liegt und deren Prozess
// nicht mehr läuft — und räumt ihre Marken, damit derselbe Abbruch nicht bei
// jedem Start wiederkommt. **Eine leere Spur des Abbruchs geht mit**; eine
// mit Inhalt bleibt fürs Diagnosepaket.
func abbruecheLesen(ordner string, pid int, lebt func(int) bool) []Abbruch {
	marken, _ := filepath.Glob(filepath.Join(ordner, "fenster-*.json")) //nolint:errcheck // ein festes Muster
	abbrueche := []Abbruch{}

	for _, pfad := range marken {
		var marke sitzungsmarke

		roh, lesefehler := leseHoechstens(pfad, markeHoechstens)
		if lesefehler == nil && json.Unmarshal(roh, &marke) != nil {
			lesefehler = errors.New("unlesbar")
		}

		if lesefehler == nil && sitzungLebt(ordner, marke, pid, lebt) {
			continue
		}

		abbruch := Abbruch{Gestartet: marke.Gestartet, Fassung: marke.Fassung}

		if marke.Spur != "" {
			spurpfad := filepath.Join(ordner, filepath.Base(marke.Spur))
			//nolint:gosec // G703: der Ordner der Sitzungen neben dem Bestand, von dieser Anwendung gewählt
			if stand, err := os.Stat(spurpfad); err == nil {
				abbruch.MitSpur = stand.Size() > 0
				if !abbruch.MitSpur {
					_ = os.Remove(spurpfad) //nolint:errcheck,gosec // eine leere Spur sagt nichts
				}
			}
		}

		_ = os.Remove(pfad)                             //nolint:errcheck,gosec // gelesen ist gelesen; liegt sie noch, kommt der Abbruch einmal mehr
		_ = os.Remove(anspruchspfad(ordner, marke.PID)) //nolint:errcheck,gosec // dito
		abbrueche = append(abbrueche, abbruch)
	}

	return abbrueche
}

// spurenKuerzen lässt die neuesten Spuren mit Inhalt stehen und nimmt die
// ältesten weg. **Leere Spuren bleiben**: eine davon gehört vielleicht einem
// laufenden Fenster, und ihr Griff liefe sonst ins Leere.
func spurenKuerzen(ordner string) {
	pfade, _ := filepath.Glob(filepath.Join(ordner, "absturz-*.txt")) //nolint:errcheck // ein festes Muster

	type spur struct {
		pfad string
		zeit time.Time
	}

	gefuellt := []spur{}

	for _, pfad := range pfade {
		//nolint:gosec // G703: der Ordner der Sitzungen neben dem Bestand, von dieser Anwendung gewählt
		if stand, err := os.Stat(pfad); err == nil && stand.Size() > 0 {
			gefuellt = append(gefuellt, spur{pfad: pfad, zeit: stand.ModTime()})
		}
	}

	sort.Slice(gefuellt, func(i, j int) bool { return gefuellt[i].zeit.After(gefuellt[j].zeit) })

	for _, alt := range gefuellt[min(len(gefuellt), absturzspurenHoechstens):] {
		_ = os.Remove(alt.pfad) //nolint:errcheck,gosec // bleibt sie, ist es eine mehr
	}
}
