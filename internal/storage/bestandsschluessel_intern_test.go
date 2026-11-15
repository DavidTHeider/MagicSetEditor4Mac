// Prüfungen an `schluesselFuer` — der Frage, **womit** ein Bestand aufgeht.
//
// **Im Paket**, weil die Naht zum Schlüsselbund nicht exportiert ist: sie darf
// es nicht sein. Ein Aufrufer, der sie setzen könnte, könnte den Schlüssel
// auch woandershin lesen lassen.
package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// mitSchluessel legt den Schlüsselbund für die Dauer eines Laufs um.
//
// **Ohne diese Naht ginge jeder Lauf an den Schlüsselbund dieses Rechners**
// (§3) — und `security` fragt bei manchen Fehlschlägen nach einem Kennwort und
// wartet dann.
func mitSchluessel(t *testing.T, schluessel string, fehler error) {
	t.Helper()

	vorher := schluesselleser
	schluesselleser = func(context.Context) (string, error) {
		return schluessel, fehler
	}

	t.Cleanup(func() { schluesselleser = vorher })
}

// **Ein verschlüsselter Bestand geht mit dem Schlüssel aus dem Bund auf** —
// ohne dass eine Fläche ihn nennen müsste.
func TestOeffnenHoltDenSchluesselAusDemBund(t *testing.T) {
	ort := t.TempDir()
	klar := filepath.Join(ort, "klar.db")
	chiffre := filepath.Join(ort, "chiffre.db")

	db, err := Oeffnen(probekontext(t), klar)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if _, fehler := db.ExecContext(probekontext(t),
		`CREATE TABLE probe (wert TEXT); INSERT INTO probe VALUES ('Feuerball')`); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	schluessel := NeuerSchluessel()

	if fehler := Verschluesseln(probekontext(t), db, chiffre, schluessel); fehler != nil {
		t.Fatalf("verschlüsseln: %v", fehler)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	mitSchluessel(t, schluessel, nil)

	zurueck, err := Oeffnen(probekontext(t), chiffre)
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

// **Und fehlt er, sagt die Meldung das** — statt „die Datei ist kein Bestand".
//
// Das ist der Unterschied zwischen zwei Ratschlägen, von denen einer Daten
// kostet: „ersetze die Datei" gegen „dein Schlüssel ist fort". Ohne diesen
// Lauf käme der erste.
func TestOhneSchluesselImBundSagtDieMeldungWasFehlt(t *testing.T) {
	ort := t.TempDir()
	klar := filepath.Join(ort, "klar.db")
	chiffre := filepath.Join(ort, "chiffre.db")

	db, err := Oeffnen(probekontext(t), klar)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if _, fehler := db.ExecContext(probekontext(t),
		`CREATE TABLE probe (wert TEXT)`); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	schluessel := NeuerSchluessel()

	if fehler := Verschluesseln(probekontext(t), db, chiffre, schluessel); fehler != nil {
		t.Fatalf("verschlüsseln: %v", fehler)
	}

	_ = db.Close() //nolint:errcheck // Test

	mitSchluessel(t, "", errors.New("nicht gefunden"))

	_, fehler := Oeffnen(probekontext(t), chiffre)

	if !errors.Is(fehler, domain.ErrSchluesselFehlt) {
		t.Fatalf("Fehler %v, erwartet war ErrSchluesselFehlt", fehler)
	}

	// **Die Gegenprobe zum Ratschlag**: der alte Weg meldete
	// `SQLITE_NOTADB`, und `internal/app` machte daraus „die Datei ist kein
	// Bestand". Diese Zeile hält fest, dass das nicht mehr geschieht.
	if errors.Is(fehler, domain.ErrSicherung) {
		t.Error("der Fehler kommt als „keine taugliche Datei\" an")
	}
}

// **Ein Klartextbestand fragt den Schlüsselbund gar nicht erst.**
//
// Das ist keine Feinheit: `security` fragt bei manchen Fehlschlägen nach einem
// Kennwort und wartet dann (siehe `schluesselbund.UeblicheFrist`). Ein Ruf bei
// jedem Öffnen wäre ein Hänger, der auf niemandes Rechnung geht.
func TestEinKlartextbestandFragtDenBundNicht(t *testing.T) {
	gefragt := 0

	vorher := schluesselleser
	schluesselleser = func(context.Context) (string, error) {
		gefragt++

		return "", errors.New("hier soll niemand fragen")
	}

	t.Cleanup(func() { schluesselleser = vorher })

	klar := filepath.Join(t.TempDir(), "klar.db")

	db, err := Oeffnen(probekontext(t), klar)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	_ = db.Close() //nolint:errcheck // Test

	// **Der Beleg, dass der Zähler überhaupt zählt** (§10: eine Gegenprobe,
	// die nichts ändert, ist keine). Eine Datei mit fremdem Kopf gilt als
	// verschlüsselt — und dann wird gefragt.
	fremd := filepath.Join(t.TempDir(), "fremd.db")
	if fehler := os.WriteFile(fremd, []byte("kein SQLite, sondern Text"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	_, _ = Oeffnen(probekontext(t), fremd) //nolint:errcheck // der Fehler ist hier egal

	if gefragt != 1 {
		t.Errorf("der Bund wurde %dmal gefragt, erwartet war genau einmal "+
			"— nämlich für die fremde Datei", gefragt)
	}
}
