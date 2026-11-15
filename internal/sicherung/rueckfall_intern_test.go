// **Der Rückfall für Dateisysteme ohne harte Verweise** — Befund 4 des
// Nachlaufs über
// #1162.
//
// **Eine interne Datei**, weil der Verknüpfer ein Wert und kein Paketfeld ist:
// nur von innen lässt sich `veroeffentlichen` einer geben, den es auf diesem
// Rechner nicht gibt. Warum kein Paketfeld, steht dort.
package sicherung

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// packenOhneVerweise packt eine Probesicherung über den Rückfallweg.
func packenOhneVerweise(t *testing.T, ziel string) error {
	t.Helper()

	return veroeffentlichen(gepacktesArchivIntern(t), ziel,
		func(string, string) error {
			return &os.LinkError{Op: "link", Err: syscall.ENOTSUP}
		})
}

// gepacktesArchivIntern legt eine gültige Sicherung an und gibt ihren Pfad.
func gepacktesArchivIntern(t *testing.T) string {
	t.Helper()

	wurzel := t.TempDir()

	bestand := filepath.Join(wurzel, "bestand.db")
	if err := os.WriteFile(bestand, []byte("SQLite format 3\x00 tut so"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	pfad := filepath.Join(wurzel, "quelle"+Endung)
	if fehler := Packen(pfad, []Teil{{Name: "bestand.db", Quelle: bestand}},
		Angaben{Fassung: "0.3.0"}); fehler != nil {
		t.Fatalf("Packen: %v", fehler)
	}

	return pfad
}

// **Auf einem Dateisystem ohne harte Verweise entsteht die Sicherung
// trotzdem.**
//
// Befund 4 des Nachlaufs: die Veröffentlichung hing allein an `os.Link`. FAT
// und exFAT kennen keine harten Verweise, und macOS richtet externe
// Datenträger ausdrücklich so ein, wenn sie auch an Windows gehen sollen — ein
// USB-Stick ist der wahrscheinlichste Ort einer Sicherung überhaupt. Dort wäre
// alles geschrieben worden und der letzte Schritt gescheitert.
//
// **Was der Lauf nicht zeigt**: dass der Rückfall unteilbar ist. Er ist es
// nicht, und das steht bei `veroeffentlichen` — ein Dateisystem ohne harte
// Verweise hat keine andere Möglichkeit.
func TestOhneHarteVerweiseEntstehtDieSicherungTrotzdem(t *testing.T) {
	ort := t.TempDir()
	ziel := filepath.Join(ort, "aufFat"+Endung)

	if fehler := packenOhneVerweise(t, ziel); fehler != nil {
		t.Fatalf("Packen ohne harte Verweise: %v", fehler)
	}

	// **Und sie taugt** — der Rückfall kopiert vollständig, nicht halb.
	if _, fehler := Entpacken(ziel, t.TempDir()); fehler != nil {
		t.Errorf("die kopierte Sicherung ist unbrauchbar: %v", fehler)
	}

	// **Ein zweites Mal derselbe Name geht auch hier nicht.** Die Zusage „eine
	// Sicherung überschreibt nichts" hängt nicht am Weg.
	fehler := packenOhneVerweise(t, ziel)
	if fehler == nil {
		t.Fatal("der belegte Name kam durch")
	}

	if !strings.Contains(fehler.Error(), "liegt schon da") {
		t.Errorf("die Meldung nennt den Grund nicht: %v", fehler)
	}
}

// **Ein anderer Fehler beim Verknüpfen wird gemeldet, nicht umgangen.**
//
// Der Rückfall gilt für „kann das nicht" und nicht für „darf das nicht" — wer
// jeden Fehler mit einer Kopie beantwortete, verstiege sich über eine volle
// Platte hinweg.
func TestEinAndererVerknuepfungsfehlerWirdGemeldet(t *testing.T) {
	ziel := filepath.Join(t.TempDir(), "kaputt"+Endung)

	daneben := gepacktesArchivIntern(t)

	fehler := veroeffentlichen(daneben, ziel, func(string, string) error {
		return &os.LinkError{Op: "link", Err: syscall.ENOSPC}
	})
	if fehler == nil {
		t.Fatal("der Fehler beim Verknüpfen wurde übergangen")
	}

	if _, err := os.Stat(ziel); !os.IsNotExist(err) {
		t.Errorf("es liegt etwas am Ziel: %v", err)
	}
}

// **Der Rückfall meldet, was ihm im Weg steht.**
//
// Drei Wege, die nur er hat — der unteilbare `os.Link` kennt sie nicht:
// die Nachbardatei ist fort, das Zielverzeichnis nimmt nichts an, und das
// Kopieren bricht ab. In allen dreien darf am Ziel nichts liegenbleiben.
func TestDerRueckfallMeldetSeineHindernisse(t *testing.T) {
	nichtVorhanden := func(t *testing.T, ort string) (string, string) {
		t.Helper()

		return filepath.Join(ort, "gibt-es-nicht"), filepath.Join(ort, "ziel"+Endung)
	}

	zielGesperrt := func(t *testing.T, ort string) (string, string) {
		t.Helper()

		zu := filepath.Join(ort, "zu")
		if err := os.Mkdir(zu, 0o500); err != nil {
			t.Fatalf("anlegen: %v", err)
		}

		// #nosec G302 -- ein Verzeichnis braucht das Ausführrecht.
		t.Cleanup(func() { _ = os.Chmod(zu, 0o700) }) //nolint:errcheck // Aufräumen

		return gepacktesArchivIntern(t), filepath.Join(zu, "ziel"+Endung)
	}

	quelleIstVerzeichnis := func(t *testing.T, ort string) (string, string) {
		t.Helper()

		// **Ein Verzeichnis**: `os.Open` gelingt, das Kopieren nicht.
		alsOrdner := filepath.Join(ort, "ordner")
		if err := os.Mkdir(alsOrdner, 0o700); err != nil {
			t.Fatalf("anlegen: %v", err)
		}

		return alsOrdner, filepath.Join(ort, "ziel"+Endung)
	}

	for name, aufbauen := range map[string]func(*testing.T, string) (string, string){
		"die Nachbardatei ist fort":    nichtVorhanden,
		"das Zielverzeichnis ist zu":   zielGesperrt,
		"die Nachbardatei taugt nicht": quelleIstVerzeichnis,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ort := t.TempDir()
			unterwegs, ziel := aufbauen(t, ort)

			fehler := veroeffentlichen(unterwegs, ziel, func(string, string) error {
				return &os.LinkError{Op: "link", Err: syscall.ENOTSUP}
			})
			if fehler == nil {
				t.Fatal("das Hindernis wurde übergangen")
			}

			if _, err := os.Stat(ziel); !os.IsNotExist(err) {
				t.Errorf("am Ziel liegt etwas: %v", err)
			}
		})
	}
}
