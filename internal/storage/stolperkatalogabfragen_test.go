// **Was der Stolpertreiber auf die Abfragen des Paketkatalogs antwortet.**
//
// Eine eigene Datei, weil `stolperabfragen_test.go` mit ihr über 400 Zeilen
// ginge (§2). Der Schnitt folgt der Frage: dort stehen die Abfragen der Karten
// und Sets, hier die vier, die der Katalog seit
// #1325
// stellt.
package storage

import (
	"database/sql/driver"
	"strings"
)

// katalogabfrage beantwortet eine Abfrage des Paketkatalogs.
//
// **Der dritte Rückgabewert sagt, ob sie erkannt wurde** — dieselbe Gestalt
// wie `turnierabfrage` und `fassungsabfrage` daneben: wer nichts erkennt,
// lässt alles Folgende unverändert.
func katalogabfrage(stelle stolperstelle, abfrage string) (driver.Rows, bool, error) {
	// **Der Schreibzähler des Paketkatalogs** (#1325) — eine Spalte.
	if strings.HasPrefix(strings.TrimSpace(abfrage), "SELECT total_changes()") {
		if stelle == beimSchreibzaehler {
			return nil, true, errGestolpert
		}

		// **Der zweite Zähler steht am Ende der Transaktion** — der erste
		// gleich am Anfang. Wer beim ersten stolpert, sieht den zweiten nie.
		if stelle == beimZweitenSchreibzaehler {
			zaehleraufrufe++

			if zaehleraufrufe >= 2 {
				return nil, true, errGestolpert
			}
		}

		return &zaehlerzeile{}, true, nil
	}

	// **Die Abhängigkeiten eines Pakets** (#1325) — zwei Spalten.
	if strings.HasPrefix(strings.TrimSpace(abfrage),
		"SELECT braucht, version FROM package_deps") {
		switch stelle {
		case beimBedarfslesen:
			return nil, true, errGestolpert
		case beimBedarfsscan:
			return &stolpernbedarf{}, true, nil
		case beimBedarfstyp:
			return &falscherbedarf{}, true, nil
		}

		return &bedarfszeilen{}, true, nil
	}

	// **Die Kennung einer Zeile ohne Änderung** (#1325) — eine Spalte.
	if strings.HasPrefix(strings.TrimSpace(abfrage), "SELECT id FROM packages") {
		if stelle == beimKennungslesen {
			return nil, true, errGestolpert
		}

		return &kennungszeilen{}, true, nil
	}

	// **Ein `INSERT … RETURNING`, das nichts gibt** — der Fall „an dieser
	// Zeile hat sich nichts geändert" (#1325). Nur an der einen Stelle, denn
	// überall sonst braucht der Katalog seine Kennung.
	if stelle == beimKennungslesen &&
		strings.HasPrefix(strings.TrimSpace(abfrage), "INSERT INTO packages") {
		return &leerezeilen{}, true, nil
	}

	return nil, false, nil
}
