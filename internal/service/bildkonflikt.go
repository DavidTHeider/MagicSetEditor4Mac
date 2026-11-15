package service

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Was bei einem Namenskonflikt in der Bildausgabe geschieht —
// #491.
//
// # Der Befund
//
// Die Bildausgabe schrieb eine Datei je Karte und **überschrieb**, wenn eine
// dalag. Das Muster `{card.name}` erzeugt für zwei Länder desselben Namens
// denselben Dateinamen; eines blieb übrig, und niemand erfuhr es.
//
// # Was MSE tut
//
//	case CONFLICT_KEEP_OLD:  return !fn.FileExists();
//	case CONFLICT_OVERWRITE: return true;
//	case CONFLICT_NUMBER: {
//	  int i = 0; String ext = fn.GetExt();
//	  while (fn.FileExists()) fn.SetExt(String() << ++i << _(".") << ext);
//	  return true; }
//	case CONFLICT_NUMBER_OVERWRITE: {
//	  int i = 0; String ext = fn.GetExt();
//	  while (used.find(fn.GetFullPath()) != used.end())
//	    fn.SetExt(String() << ++i << _(".") << ext);
//	  return true; }
//
// `src/util/file_utils.cpp`, Zeilen 74–100, Fassung f9d9356 vom 2021-07-06.
//
// # `SetExt` und nicht „Name (1)"
//
// **Das ist die Stelle, an der man raten würde.** `SetExt` **ersetzt** die
// Endung: aus `Feuerball.png` wird `Feuerball.1.png`, nicht `Feuerball (1).png`
// und nicht `Feuerball1.png`. Und `ext` wird **vor** der Schleife gelesen,
// bleibt also die ursprüngliche — sonst hieße der zweite `Feuerball.1.2.png`.
//
// # Zwei Quellen, ein Unterschied
//
// `nummerieren` fragt die **Platte** und weicht damit auch dem aus, was ein
// früherer Lauf hinterlassen hat. `nummerierenneu` fragt nur den **laufenden
// Durchgang** — es überschreibt Altes und stößt innerhalb eines Laufs nie
// zusammen. Das ist MSEs Vorgabe.

// Namenskonflikt gibt den Namen, unter dem geschrieben wird — und ob überhaupt.
//
// `gelaufen` sind die Namen, die dieser Durchgang schon vergeben hat; nur
// `nummerierenneu` liest sie. Ein `nil` ist erlaubt und heißt „noch keiner".
func Namenskonflikt(
	verzeichnis, datei, verhalten string, gelaufen map[string]bool,
) (string, bool) {
	switch verhalten {
	case domain.BildkonfliktAltbehalten:
		if _, err := os.Stat(filepath.Join(verzeichnis, datei)); err == nil {
			return datei, false
		}

		return datei, true

	case domain.BildkonfliktNummerieren:
		return frei(datei, func(kandidat string) bool {
			_, err := os.Stat(filepath.Join(verzeichnis, kandidat))

			return err == nil
		}), true

	case domain.BildkonfliktNummerierenNeu:
		// **Kleingeschrieben verglichen**, und das ist eine benannte
		// Abweichung: MSEs `used` ist ein `std::set<String>` und vergleicht
		// genau. macOS unterscheidet in der Vorgabe **nicht** zwischen
		// `Feuerball` und `feuerball`; ein Lauf, der beide schriebe, verlöre
		// die erste Karte auf genau der Hälfte aller Rechner.
		return frei(datei, func(kandidat string) bool {
			return gelaufen[strings.ToLower(kandidat)]
		}), true

	default:
		// **Ein unbekanntes Verhalten schreibt**, und zwar unter dem gegebenen
		// Namen. Der Fall entsteht nur aus einem Bestand, den jemand von Hand
		// geändert hat — `domain.PruefeEinstellung` lässt ihn nicht durch.
		// Nicht zu schreiben hieße, eine Ausgabe stillschweigend ausfallen zu
		// lassen; `ueberschreiben` ist die harmloseste Antwort und zugleich
		// das, was dieses Programm vor #491 immer tat.
		return datei, true
	}
}

// frei zählt hoch, solange `belegt` zutrifft.
//
// **Die Endung wird einmal gelesen**, nicht in jedem Durchgang — sonst hieße
// der zweite Versuch `da.1.2.png`.
func frei(datei string, belegt func(string) bool) string {
	endung := filepath.Ext(datei)
	rumpf := datei[:len(datei)-len(endung)]

	kandidat := datei

	for i := 1; belegt(kandidat); i++ {
		kandidat = rumpf + "." + strconv.Itoa(i) + endung
	}

	return kandidat
}
