// **Eine Sicherung, die alles trägt**
// (#1162,
// Befund Z6-01 des Audits (Entwicklungsarchiv)).
//
// # Warum es dieses Paket gibt
//
// `storage.Sichern` schreibt den **Bestand**. Die eingesetzten Bilder liegen
// daneben unter `beiwerk/`, und
// [0029](../../docs/entscheidungen/0029-bilder-liegen-neben-dem-set.md) sagt
// selbst: ein Set ist ohne dieses Verzeichnis unvollständig.
//
// **Ein Blatt** (§2): es kennt `domain` und sonst nichts. Was in eine
// Sicherung gehört, entscheidet der Aufrufer; dieses Paket weiß, wie ein
// Archiv aussieht und wie man es wieder auspackt.
package sicherung_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// probeteile legt zwei Dateien an und gibt sie als Teile.
func probeteile(t *testing.T) []sicherung.Teil {
	t.Helper()

	wurzel := t.TempDir()

	bestand := filepath.Join(wurzel, "bestand.db")
	if err := os.WriteFile(bestand, []byte("SQLite format 3\x00 tut so"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	bild := filepath.Join(wurzel, "bild.png")
	if err := os.WriteFile(bild, []byte("nicht wirklich ein PNG"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return []sicherung.Teil{
		{Name: "bestand.db", Quelle: bestand},
		{Name: "beiwerk/4/bild.png", Quelle: bild},
	}
}

// **Was hineingeht, kommt heraus.**
func TestEineSicherungGibtHerausWasHineinging(t *testing.T) {
	t.Parallel()

	teile := probeteile(t)
	ziel := filepath.Join(t.TempDir(), "probe.msesicherung")

	if err := sicherung.Packen(ziel, teile, sicherung.Angaben{
		Fassung: "0.3.0",
	}); err != nil {
		t.Fatalf("Packen: %v", err)
	}

	nach := t.TempDir()

	beschreibung, err := sicherung.Entpacken(ziel, nach)
	if err != nil {
		t.Fatalf("Entpacken: %v", err)
	}

	if beschreibung.Fassung != "0.3.0" {
		t.Errorf("die Angaben sind %+v", beschreibung)
	}

	for _, teil := range teile {
		roh, lesefehler := os.ReadFile(filepath.Join(nach, filepath.FromSlash(teil.Name))) // #nosec G304 -- aus t.TempDir()
		if lesefehler != nil {
			t.Errorf("%s fehlt: %v", teil.Name, lesefehler)

			continue
		}

		wollte, lesefehler := os.ReadFile(teil.Quelle) // #nosec G304 -- aus t.TempDir()
		if lesefehler != nil {
			t.Fatalf("die Quelle lesen: %v", lesefehler)
		}

		if !bytes.Equal(roh, wollte) {
			t.Errorf("%s kam anders heraus", teil.Name)
		}
	}
}

// **Und die Prüfsummen fangen eine veränderte Sicherung.**
//
// **Das ist der Unterschied zwischen einem Archiv und einer Sicherung**: eine
// Sicherung sagt, ob sie noch die ist, die geschrieben wurde. Ohne diese
// Zusage fällt der Fehler erst auf, wenn jemand sie braucht — und dann ist es
// zu spät.
func TestEineVeraenderteSicherungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "probe.msesicherung")

	if err := sicherung.Packen(ziel, probeteile(t), sicherung.Angaben{
		Fassung: "0.3.0",
	}); err != nil {
		t.Fatalf("Packen: %v", err)
	}

	verfaelschen(t, ziel, "beiwerk/4/bild.png", []byte("etwas ganz anderes"))

	_, err := sicherung.Entpacken(ziel, t.TempDir())

	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("der Fehler ist %v — will ErrConflict", err)
	}

	if !strings.Contains(err.Error(), "bild.png") {
		t.Errorf("die Meldung nennt den Teil nicht: %v", err)
	}
}

// **Ein fehlender Teil ebenso.**
func TestEinFehlenderTeilWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "probe.msesicherung")

	if err := sicherung.Packen(ziel, probeteile(t), sicherung.Angaben{
		Fassung: "0.3.0",
	}); err != nil {
		t.Fatalf("Packen: %v", err)
	}

	entfernen(t, ziel, "beiwerk/4/bild.png")

	if _, err := sicherung.Entpacken(ziel, t.TempDir()); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("der Fehler ist %v — will ErrConflict", err)
	}
}

// **Eine alte Sicherung ist eine nackte Datenbank, und sie wird erkannt.**
//
// **Ohne diese Unterscheidung wäre jede Sicherung aus 0.1.0 und 0.2.0
// wertlos** — und das wäre eine schlechtere Behebung als der Fehler.
func TestEineNackteDatenbankWirdErkannt(t *testing.T) {
	t.Parallel()

	alt := filepath.Join(t.TempDir(), "bestand.db")
	if err := os.WriteFile(alt, []byte("SQLite format 3\x00 …"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if sicherung.IstArchiv(alt) {
		t.Error("eine nackte Datenbank gilt als Archiv")
	}

	neu := filepath.Join(t.TempDir(), "probe.msesicherung")
	if err := sicherung.Packen(neu, probeteile(t), sicherung.Angaben{}); err != nil {
		t.Fatalf("Packen: %v", err)
	}

	if !sicherung.IstArchiv(neu) {
		t.Error("ein Archiv gilt nicht als Archiv")
	}
}

// **Ein Archiv ohne Manifest ist keine Sicherung.**
func TestEinArchivOhneManifestWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	ziel := filepath.Join(t.TempDir(), "fremd.msesicherung")

	nurEintrag(t, ziel, "irgendwas.txt", []byte("hallo"))

	if _, err := sicherung.Entpacken(ziel, t.TempDir()); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("der Fehler ist %v — will ErrValidation", err)
	}
}

// **`EntpackenNur` holt, was genannt ist — und sonst nichts.**
func TestEntpackenNurHoltDasGenannte(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "auszug"+sicherung.Endung)
	teile := probeteile(t)

	if fehler := sicherung.Packen(pfad, teile, sicherung.Angaben{
		Fassung: "0.3.0",
	}); fehler != nil {
		t.Fatalf("Packen: %v", fehler)
	}

	ziel := t.TempDir()

	beschreibung, fehler := sicherung.EntpackenNur(pfad, ziel, teile[0].Name)
	if fehler != nil {
		t.Fatalf("EntpackenNur: %v", fehler)
	}

	// **Die Beschreibung führt trotzdem alle Teile.** Sie beschreibt das
	// Archiv und nicht den Auszug — wer sie als Inhaltsverzeichnis des
	// Zielverzeichnisses läse, läge falsch.
	if len(beschreibung.Teile) != len(teile) {
		t.Errorf("die Beschreibung führt %d Teile statt %d",
			len(beschreibung.Teile), len(teile))
	}

	if _, err := os.Stat(filepath.Join(ziel, filepath.FromSlash(teile[0].Name))); err != nil {
		t.Errorf("der genannte Teil kam nicht an: %v", err)
	}

	for _, teil := range teile[1:] {
		if _, err := os.Stat(filepath.Join(ziel, filepath.FromSlash(teil.Name))); !os.IsNotExist(err) {
			t.Errorf("%s kam mit, obwohl niemand danach fragte", teil.Name)
		}
	}
}

// **Und ein genannter Teil, den es nicht gibt, ist ein Fehler.**
//
// Ein *nicht* genannter fehlender bleibt unbemerkt — das ist die Kehrseite
// derselben Zusage und steht neben ihr, damit niemand aus einem geglückten
// Auszug auf ein vollständiges Archiv schließt.
func TestEntpackenNurWeistEinenFehlendenGenanntenAb(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "auszug"+sicherung.Endung)
	teile := probeteile(t)

	if fehler := sicherung.Packen(pfad, teile, sicherung.Angaben{
		Fassung: "0.3.0",
	}); fehler != nil {
		t.Fatalf("Packen: %v", fehler)
	}

	benennen(t, pfad, teile[0].Name, "anders")

	if _, fehler := sicherung.EntpackenNur(pfad, t.TempDir(), teile[0].Name); fehler == nil {
		t.Error("der fehlende genannte Teil kam durch")
	}

	// **Die Kehrseite**: wird er nicht genannt, fällt sein Fehlen nicht auf.
	if _, fehler := sicherung.EntpackenNur(pfad, t.TempDir(), teile[1].Name); fehler != nil {
		t.Errorf("ein Auszug ohne den fehlenden Teil scheiterte: %v", fehler)
	}
}
