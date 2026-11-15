package storage

// **Ein neuer Bestand gehört seinem Eigner** —
// #2508,
// gefunden beim Erststart aus dem Abbild (HAND6). SQLite legte `bestand.db`
// mit der `umask` an, `0644`; ein eingesetzter Bestand bekommt seit Vermerk
// 0174 `0600`. Derselbe Inhalt hatte damit je nach Weg zwei Rechte, und
// außerhalb von `~/Library` lasen andere Konten des Rechners mit.
//
// # Und die Begleiter
//
// **Journal, WAL und `-shm` tragen denselben Inhalt** und bekommen deshalb die
// Rechte des Bestands. Der Treiber `ncruces/go-sqlite3` legt sie in seinem
// eigenen VFS an, nachgelesen in Fassung v0.35.4: Journal und WAL mit `0666`
// und der `umask`, außer die Adresse nennt `modeof` (`vfs/file.go`); die
// `-shm` immer mit `0666` und der `umask` (`vfs/shm_bsd.go`).
//
// # Was es nicht tut
//
// **Kein `chmod` an einem Bestand, den es schon gibt** — Vermerk 0174: wer ihm
// Rechte gegeben hat, hat sie gewählt. Und **nichts an einer URI** — dem
// Arbeitsspeicher und jeder Angabe mit `file:`, die nur Läufe nehmen.

import (
	"os"
	"strings"
)

// eigenerBestand legt einen Bestand, den es noch nicht gibt, leer mit `0600`
// an, seine `-shm` mit den Rechten des Bestands, und nennt ihn SQLite als
// Vorbild für Journal und WAL.
//
// **Leer**: SQLite nimmt eine leere Datei als neue Datenbank und eine leere
// `-shm` als eine, die es erst einrichtet. **Scheitert das Anlegen**, sagt
// SQLite gleich danach mit seiner eigenen Meldung, warum — ein zweiter Satz
// dafür hier wäre derselbe Befund in anderen Worten.
func eigenerBestand(quelle, adresse string) string {
	if strings.HasPrefix(quelle, "file:") {
		return adresse
	}

	neuAnlegen(quelle, 0o600)

	if info, fehler := os.Stat(quelle); fehler == nil {
		neuAnlegen(quelle+"-shm", info.Mode().Perm())
	}

	return adresse + "&modeof=" + alsParameter(quelle)
}

// neuAnlegen legt eine leere Datei mit genau diesen Rechten an — **nur, wenn
// es sie noch nicht gibt**. Die `umask` nähme sonst Rechte weg, die der
// Bestand hat; gesetzt wird deshalb nach dem Anlegen, an der eigenen Datei.
func neuAnlegen(pfad string, rechte os.FileMode) {
	//nolint:gosec // G304: der Bestand, den der Aufrufer öffnet, und seine -shm
	// #nosec G304 -- dieselbe Begründung für den eigenständigen gosec
	datei, fehler := os.OpenFile(pfad, os.O_WRONLY|os.O_CREATE|os.O_EXCL, rechte)
	if fehler != nil {
		return
	}

	_ = datei.Chmod(rechte) //nolint:errcheck // die Datei ist eben entstanden und gehört diesem Prozess
	_ = datei.Close()       //nolint:errcheck // leer, nichts geschrieben
}

// alsParameter bringt einen Pfad in die Form eines URI-Parameters: was SQLite
// dort als Trenner liest, wird kodiert.
func alsParameter(wert string) string {
	return strings.NewReplacer("%", "%25", "&", "%26", "#", "%23", "=", "%3d").Replace(wert)
}
