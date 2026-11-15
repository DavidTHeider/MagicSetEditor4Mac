// **Eine Sicherung trägt Klartext — auch aus einem verschlüsselten Bestand** —
// Werkstück 14 von Phase 4.12
// (#1351),
// entschieden in
// [0072](../../docs/entscheidungen/0072-ein-treiber-der-verschluesseln-kann.md).
//
// # Warum das eine eigene Datei ist
//
// **Weil es eine eigene Frage ist** (§2): nebenan steht, ob die Chiffre taugt;
// hier steht, was die Chiffre mit einem **anderen** Weg auf dieselbe Datei
// macht.
//
// # Und warum es diese Läufe gibt
//
// **Weil der Fehler schon zweimal dieselbe Gestalt hatte.** `VACUUM INTO` ohne
// `vfs=` erbt das VFS der Hauptdatenbank. Gefunden hat es der Flächenlauf beim
// **Ent**schlüsseln; gesucht wurde danach nach der Klasse (§10) — und `Sichern`
// ist derselbe Weg. Gemessen am 2026-09-06: ein verschlüsselter Bestand ließ
// sich **gar nicht** sichern („unable to open database file").
//
// **Wäre es geglückt, wäre es schlimmer gewesen**: eine verschlüsselte
// Sicherung geht nur auf dem Rechner auf, auf dem sie entstand — und genau das
// soll eine Sicherung nicht sein.
package storage_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEineSicherungAusEinemVerschluesseltenBestandIstKlartext(t *testing.T) {
	ort := t.TempDir()
	klar := filepath.Join(ort, "klar.db")
	chiffre := filepath.Join(ort, "chiffre.db")
	sicherung := filepath.Join(ort, "sicherung.db")

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

	db, err := storage.OeffnenMitSchluessel(probekontext(t), chiffre, schluessel)
	if err != nil {
		t.Fatalf("verschlüsselt öffnen: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Test

	// **Erstens: es glückt überhaupt.** Vor der Behebung endete genau diese
	// Zeile mit „unable to open database file", und damit war ein
	// verschlüsselter Bestand nicht zu sichern.
	if fehler := storage.Sichern(probekontext(t), db, sicherung); fehler != nil {
		t.Fatalf("sichern: %v", fehler)
	}

	// #nosec G304 -- ein Pfad aus t.TempDir()
	roh, err := os.ReadFile(sicherung)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	// **Zweitens: sie trägt Klartext.** Ohne diese Zeile wäre auch eine
	// verschlüsselte Sicherung grün — und die ginge auf keinem zweiten
	// Rechner auf.
	if !bytes.HasPrefix(roh, []byte(klartextmarke)) {
		t.Fatalf("die Sicherung trägt die Klartextmarke nicht: %q",
			roh[:min(len(roh), 16)])
	}

	// **Drittens: sie lässt sich ohne Schlüssel öffnen** — das ist, was ein
	// Mensch auf einem anderen Rechner tut.
	zurueck, err := storage.Oeffnen(probekontext(t), sicherung)
	if err != nil {
		t.Fatalf("die Sicherung öffnen: %v", err)
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
