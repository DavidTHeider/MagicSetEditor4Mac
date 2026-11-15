// **Der Bestand liegt verschlüsselt auf der Platte** — Werkstück 14 von Phase
// 4.12 (#1351),
// nach Vermerk [0072](../../docs/entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md).
package storage_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// klartextmarke ist, woran man eine unverschlüsselte SQLite-Datei erkennt.
const klartextmarke = "SQLite format 3\x00"

// bestandMit legt einen Bestand an und schreibt ein Wort hinein, das man auf
// der Platte wiederfinden könnte.
func bestandMit(t *testing.T, pfad, wort string) {
	t.Helper()

	db, err := storage.Oeffnen(probekontext(t), pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if _, fehler := db.ExecContext(probekontext(t),
		`CREATE TABLE probe (wert TEXT)`); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO probe (wert) VALUES (?)`, wort); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}
}

// **Was verschlüsselt ist, steht nicht mehr lesbar auf der Platte.**
//
// Das ist die ganze Zusage dieses Werkstücks, und sie wird hier an dem
// geprüft, was ein Mensch mit der Datei anstellen kann: sie aufmachen und
// nachsehen.
func TestEinVerschluesselterBestandTraegtKeinenKlartext(t *testing.T) {
	ort := t.TempDir()
	klar := filepath.Join(ort, "klar.db")
	chiffre := filepath.Join(ort, "chiffre.db")

	bestandMit(t, klar, "GEHEIMWORT")

	// Der Klartextbestand trägt es — sonst prüfte der zweite Teil nichts.
	// #nosec G304 -- ein Pfad aus t.TempDir()
	roh, err := os.ReadFile(klar)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if !bytes.Contains(roh, []byte("GEHEIMWORT")) {
		t.Fatal("das Wort steht nicht im Klartextbestand — die Probe misst nichts")
	}

	schluessel := storage.NeuerSchluessel()

	db, err := storage.Oeffnen(probekontext(t), klar)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Test

	if fehler := storage.Verschluesseln(
		probekontext(t), db, chiffre, schluessel); fehler != nil {
		t.Fatalf("verschlüsseln: %v", fehler)
	}

	// #nosec G304 -- ein Pfad aus t.TempDir()
	roh, err = os.ReadFile(chiffre)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if bytes.Contains(roh, []byte("GEHEIMWORT")) {
		t.Error("das Wort steht im Klartext in der verschlüsselten Datei")
	}

	if bytes.HasPrefix(roh, []byte(klartextmarke)) {
		t.Error("die verschlüsselte Datei trägt die Klartextmarke von SQLite")
	}
}

// **Und mit dem Schlüssel kommt alles zurück.** Eine Verschlüsselung, aus der
// nichts mehr herauskommt, wäre ein Datenverlust mit guter Absicht.
func TestEinVerschluesselterBestandLaesstSichWiederLesen(t *testing.T) {
	ort := t.TempDir()
	klar := filepath.Join(ort, "klar.db")
	chiffre := filepath.Join(ort, "chiffre.db")

	bestandMit(t, klar, "Feuerball")

	schluessel := storage.NeuerSchluessel()

	quelle, err := storage.Oeffnen(probekontext(t), klar)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := storage.Verschluesseln(
		probekontext(t), quelle, chiffre, schluessel); fehler != nil {
		t.Fatalf("verschlüsseln: %v", fehler)
	}

	if fehler := quelle.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	zurueck, err := storage.OeffnenMitSchluessel(probekontext(t), chiffre, schluessel)
	if err != nil {
		t.Fatalf("verschlüsselt öffnen: %v", err)
	}

	t.Cleanup(func() { _ = zurueck.Close() }) //nolint:errcheck // Test

	var wert string
	if fehler := zurueck.QueryRowContext(probekontext(t),
		`SELECT wert FROM probe`).Scan(&wert); fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if wert != "Feuerball" {
		t.Errorf("Wert %q, erwartet war „Feuerball\"", wert)
	}
}

// **Und ohne den richtigen Schlüssel kommt nichts heraus.** Ohne diesen Fall
// bliebe offen, ob überhaupt etwas verschlüsselt wurde.
func TestOhneDenRichtigenSchluesselOeffnetSichNichts(t *testing.T) {
	ort := t.TempDir()
	klar := filepath.Join(ort, "klar.db")
	chiffre := filepath.Join(ort, "chiffre.db")

	bestandMit(t, klar, "Feuerball")

	schluessel := storage.NeuerSchluessel()

	quelle, err := storage.Oeffnen(probekontext(t), klar)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := storage.Verschluesseln(
		probekontext(t), quelle, chiffre, schluessel); fehler != nil {
		t.Fatalf("verschlüsseln: %v", fehler)
	}

	_ = quelle.Close() //nolint:errcheck // Test

	anderer := storage.NeuerSchluessel()

	for name, versuch := range map[string]func() error{
		"mit dem falschen Schlüssel": func() error {
			db, fehler := storage.OeffnenMitSchluessel(probekontext(t), chiffre, anderer)
			if fehler != nil {
				return fehler
			}

			defer db.Close() //nolint:errcheck // Test

			var wert string

			return db.QueryRowContext(probekontext(t), `SELECT wert FROM probe`).Scan(&wert)
		},
		"ganz ohne Schlüssel": func() error {
			db, fehler := storage.OeffnenMitSchluessel(probekontext(t), chiffre, "")
			if fehler != nil {
				return fehler
			}

			defer db.Close() //nolint:errcheck // Test

			var wert string

			return db.QueryRowContext(probekontext(t), `SELECT wert FROM probe`).Scan(&wert)
		},
	} {
		t.Run(name, func(t *testing.T) {
			if fehler := versuch(); fehler == nil {
				t.Error("der Bestand ging auf")
			}
		})
	}
}

// **Ein gewürfelter Schlüssel ist 64 Hexziffern und jedes Mal ein anderer.**
func TestJederNeueSchluesselIstEinAnderer(t *testing.T) {
	gesehen := map[string]bool{}

	for range 32 {
		schluessel := storage.NeuerSchluessel()

		if len(schluessel) != 64 {
			t.Fatalf("Länge %d, erwartet waren 64 Hexziffern", len(schluessel))
		}

		if strings.Trim(schluessel, "0123456789abcdef") != "" {
			t.Fatalf("%q ist nicht durchweg hexadezimal", schluessel)
		}

		if gesehen[schluessel] {
			t.Fatalf("%q kam zweimal", schluessel)
		}

		gesehen[schluessel] = true
	}
}

// **Und man sieht einer Datei an, ob sie verschlüsselt ist** — daran hängt,
// ob `Oeffnen` überhaupt nach einem Schlüssel fragt.
func TestEinerDateiSiehtManAnObSieVerschluesseltIst(t *testing.T) {
	ort := t.TempDir()
	klar := filepath.Join(ort, "klar.db")
	chiffre := filepath.Join(ort, "chiffre.db")
	leer := filepath.Join(ort, "leer.db")
	fehlt := filepath.Join(ort, "gibtesnicht.db")

	bestandMit(t, klar, "x")

	schluessel := storage.NeuerSchluessel()

	db, err := storage.Oeffnen(probekontext(t), klar)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := storage.Verschluesseln(
		probekontext(t), db, chiffre, schluessel); fehler != nil {
		t.Fatalf("verschlüsseln: %v", fehler)
	}

	_ = db.Close() //nolint:errcheck // Test

	if fehler := os.WriteFile(leer, nil, 0o600); fehler != nil {
		t.Fatalf("leere Datei: %v", fehler)
	}

	for name, fall := range map[string]struct {
		pfad     string
		erwartet bool
	}{
		"ein Klartextbestand":           {klar, false},
		"ein verschlüsselter":           {chiffre, true},
		"eine leere Datei":              {leer, false},
		"eine Datei, die es nicht gibt": {fehlt, false},
	} {
		t.Run(name, func(t *testing.T) {
			ist, fehler := storage.IstVerschluesselt(fall.pfad)
			if fehler != nil {
				t.Fatalf("prüfen: %v", fehler)
			}

			if ist != fall.erwartet {
				t.Errorf("verschlüsselt=%v, erwartet war %v", ist, fall.erwartet)
			}
		})
	}
}

// **Und was entschlüsselt wird, liegt danach wirklich im Klartext.**
//
// # Warum dieser Lauf hier steht und nicht nur im Flächenlauf
//
// **Weil ihn der Flächenlauf gefunden hat und diese Schicht ihn durchließ.**
// Am 2026-09-06 erbte `VACUUM INTO` ohne `vfs=` das VFS der Hauptdatenbank —
// beim Entschlüsseln also die Chiffre. Das Ziel entstand wieder verschlüsselt,
// mit einem Schlüssel, den niemand kennt.
//
// Die Läufe darüber prüften nur die eine Richtung. Ein Weg, der hin und nicht
// zurück führt, ist an dieser Schicht zu sehen — und wird deshalb hier
// geprüft, nicht drei Schichten weiter oben.
func TestEinEntschluesselterBestandTraegtWiederDieKlartextmarke(t *testing.T) {
	ort := t.TempDir()
	klar := filepath.Join(ort, "klar.db")
	chiffre := filepath.Join(ort, "chiffre.db")
	zurueck := filepath.Join(ort, "zurueck.db")

	bestandMit(t, klar, "Feuerball")

	schluessel := storage.NeuerSchluessel()

	quelle, err := storage.Oeffnen(probekontext(t), klar)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := storage.Verschluesseln(
		probekontext(t), quelle, chiffre, schluessel); fehler != nil {
		t.Fatalf("verschlüsseln: %v", fehler)
	}

	_ = quelle.Close() //nolint:errcheck // Test

	// **Aus der verschlüsselten heraus zurück in den Klartext** — das ist der
	// Weg, den ein Mensch geht, der die Verschlüsselung wieder ausschaltet.
	verschluesselt, err := storage.OeffnenMitSchluessel(
		probekontext(t), chiffre, schluessel)
	if err != nil {
		t.Fatalf("verschlüsselt öffnen: %v", err)
	}

	if fehler := storage.Entschluesseln(
		probekontext(t), verschluesselt, zurueck); fehler != nil {
		t.Fatalf("entschlüsseln: %v", fehler)
	}

	_ = verschluesselt.Close() //nolint:errcheck // Test

	// #nosec G304 -- ein Pfad aus t.TempDir()
	roh, err := os.ReadFile(zurueck)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if !bytes.HasPrefix(roh, []byte(klartextmarke)) {
		t.Fatalf("die entschlüsselte Datei trägt die Klartextmarke nicht: %q",
			roh[:min(len(roh), 16)])
	}

	if !bytes.Contains(roh, []byte("Feuerball")) {
		t.Error("der Inhalt fehlt")
	}

	// **Und sie geht ohne jeden Schlüssel auf** — das ist der Zweck.
	offen, err := storage.Oeffnen(probekontext(t), zurueck)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = offen.Close() }) //nolint:errcheck // Test

	var wert string
	if fehler := offen.QueryRowContext(probekontext(t),
		`SELECT wert FROM probe`).Scan(&wert); fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}
}
