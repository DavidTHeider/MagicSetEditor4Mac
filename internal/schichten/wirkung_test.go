// Der Wächter über die Zusage aus
// #51:
// **die Sprache kann keine Datei anfassen.**
//
// # Warum das hier steht und nicht an der Bibliothek
//
// Bis #227
// hing die Zusage an **Funktionsnamen**: ein Test durchsuchte
// `msescript.Bibliothek().Namen()` nach `write_`, `read_`, `open`, `file`,
// `path`, `exec`, `system`, `dir`. Eine Funktion namens `sichere` oder
// `ablegen` wäre grün geblieben — und eine bestehende Funktion, die eines Tages
// intern eine Datei öffnet, ebenfalls. Das war kein Loch, sondern ein Hinweis,
// der wie eine Grenze aussah, und genau diese Sorte streicht §10 an.
//
// **Die Grenze ist jetzt strukturell:** `internal/msescript` importiert nichts,
// womit sich eine Datei, ein Prozess oder eine Verbindung anfassen ließe. Wer
// das nicht **kann**, tut es nicht — gleich, wie seine Funktionen heißen.
//
// # Wie #216 hindurchpasst
//
// `write_image_file` schreibt absichtlich Dateien. Sie kommt **nicht** in die
// Bibliothek, sondern wird vom Aufrufer in die Umgebung gelegt
// (`Umgebung.Setze`) — von `internal/service`, das Dateien anfassen darf und
// dessen Weg dorthin ein Mensch liest. Damit bleibt die Zusage, wie sie ist:
// die Sprache **selbst** kann nichts; wer ihr eine Wirkung gibt, tut es an
// einer Stelle, die man sieht.
//
// # Was dieser Wächter nicht prüft
//
//   - **Was der Aufrufer hineinlegt.** Eine Funktion, die `service` in die
//     Umgebung setzt, darf schreiben — das ist der Zweck. Ob sie es
//     verantwortlich tut, entscheidet ihr eigener Test, nicht dieser.
//   - **Keine Testdatei.** `msescript/fremdpakete_test.go` liest echte Pakete
//     von der Platte und muss es. Ausgeliefert wird davon nichts; die Zusage
//     handelt von der Sprache, und die besteht aus dem, was gebaut wird.
//   - **Keinen Umweg über ein anderes Paket dieses Projekts.** Die
//     Schichtenregel hält `msescript` ohnehin frei von allem — es importiert
//     **nichts** aus dem Projekt, auch nicht `domain` (§2).
package schichten

import (
	"strings"
	"testing"
)

// wirkende sind die Standardbibliotheken, mit denen sich etwas außerhalb des
// Prozesses anfassen lässt.
//
// **`path` steht nicht dabei, `path/filepath` schon.** Das erste rechnet mit
// Zeichenketten und öffnet nichts; das zweite kennt das Dateisystem dieses
// Rechners. Wer beide aufnähme, machte den Wächter unbrauchbar, sobald jemand
// einen Namen zerlegen will.
var wirkende = []string{"os", "syscall", "net", "plugin", "path/filepath", "io/ioutil"}

// wirkungsfrei sind die Pakete, die nichts anfassen dürfen.
//
// **Eine Liste und keine Ausnahmeliste**, anders als beim Netzwächter: hier
// steht **eine** Zusage über **ein** Paket, und sie gilt, weil `msescript` eine
// Sprache ist. Jedes andere Paket darf Dateien lesen; die Frage ist dort eine
// andere.
var wirkungsfrei = []string{"msescript"}

func TestDieSpracheKannNichtsAnfassen(t *testing.T) {
	importe := importeLesen(t)

	for _, paket := range wirkungsfrei {
		for _, einfuhr := range importe[paket] {
			if strings.HasSuffix(einfuhr.Datei, "_test.go") {
				continue
			}

			if istWirkend(einfuhr.Pfad) {
				t.Errorf("internal/%s (%s) importiert %s — die Sprache soll keine Datei, "+
					"keinen Prozess und keine Verbindung anfassen können (#51, #227)",
					paket, einfuhr.Datei, einfuhr.Pfad)
			}
		}
	}
}

// TestDerWirkungswaechterWirdRot ist die Gegenprobe — als Testfall (§10).
//
// **Sie ist der einzige Beleg, dass er greift.** Der Wächter, den er ablöst,
// sah von außen genauso aus und hielt die halbe Zusage.
func TestDerWirkungswaechterWirdRot(t *testing.T) {
	faelle := map[string]bool{
		"os":            true,
		"os/exec":       true,
		"syscall":       true,
		"net/http":      true,
		"path/filepath": true,
		"io/ioutil":     true,
		"plugin":        true,

		// **Das darf die Sprache.** Zeichenketten, Zahlen, Zeit, Kontext.
		"path":     false,
		"strings":  false,
		"strconv":  false,
		"time":     false,
		"context":  false,
		"fmt":      false,
		"errors":   false,
		"unicode":  false,
		"osmosis":  false,
		"nettonet": false,
	}

	for pfad, erwartet := range faelle {
		t.Run(pfad, func(t *testing.T) {
			if istWirkend(pfad) != erwartet {
				t.Errorf("istWirkend(%q) = %v, erwartet %v", pfad, !erwartet, erwartet)
			}
		})
	}
}

// istWirkend sagt, ob ein Importpfad aus dem Prozess hinausführt.
//
// **Über den vollen Namen und die Unterpakete**, nicht über ein Präfix auf der
// Zeichenkette: `osmosis` beginnt mit `os` und ist nichts dergleichen.
func istWirkend(pfad string) bool {
	for _, wirkend := range wirkende {
		if pfad == wirkend || strings.HasPrefix(pfad, wirkend+"/") {
			return true
		}
	}

	return false
}
