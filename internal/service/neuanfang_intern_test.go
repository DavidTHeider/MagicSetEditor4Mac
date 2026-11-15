package service

// **Ganz von vorn** — Werkstück O4 von Phase 5.2
// (#1428).
//
// **Paketintern**, weil die Vorrichtung hier liegt: ein Bestandsweg über einer
// Datei, die keine Datenbank ist. Was der Neuanfang tut, ist Umbenennen — was
// darin steht, entscheidet niemand hier.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der alte Bestand liegt danach beiseite, und ein frischer Start legt einen
// neuen an.**
//
// **Das ist die Umkehrbarkeit**, auf die sich die Abnahme beruft: „Abbruch
// während Betrieb hinterlässt wiederaufnehmbaren, verständlich gemeldeten
// Zustand."
func TestDerNeuanfangLegtDenBestandBeiseite(t *testing.T) {
	bau := neuerWegbau(t)

	if fehler := os.MkdirAll(bau.bilder, 0o750); fehler != nil {
		t.Fatalf("das Bilderverzeichnis anlegen: %v", fehler)
	}

	beiseite, fehler := bau.weg.Neuanfang(t.Context())
	if fehler != nil {
		t.Fatalf("neu anfangen: %v", fehler)
	}

	// **Der alte Bestand ist da, der neue noch nicht** — die Datei entsteht
	// beim Öffnen, und das Öffnen ist in diesem Lauf eine Attrappe.
	inhalt, lesefehler := os.ReadFile(beiseite) // #nosec G304 -- ein Pfad aus dem Lauf selbst
	if lesefehler != nil {
		t.Fatalf("den alten Bestand lesen: %v", lesefehler)
	}

	if string(inhalt) != "laufender bestand" {
		t.Errorf("beiseite liegt %q", inhalt)
	}

	if _, statfehler := os.Stat(filepath.Join(bau.ort, "bestand.db")); !errors.Is(
		statfehler, os.ErrNotExist) {
		t.Errorf("die alte Bestandsdatei steht noch da: %v", statfehler)
	}

	// **Und die Bilder sind weg** — ein Neuanfang lässt keine Bilder eines
	// Bestands stehen, den es nicht mehr gibt.
	if _, statfehler := os.Stat(bau.bilder); !errors.Is(statfehler, os.ErrNotExist) {
		t.Errorf("das Bilderverzeichnis steht noch da: %v", statfehler)
	}

	// **Die Marke ist weggeräumt** — sonst holte der nächste Start eine
	// Arbeit nach, die niemand mehr will.
	if _, statfehler := os.Stat(bau.weg.markenpfad()); !errors.Is(
		statfehler, os.ErrNotExist) {
		t.Errorf("die Rückspielmarke liegt noch da: %v", statfehler)
	}

	// **Und die Verbindung ist wieder offen.**
	if bau.leben.geschlossen != 1 || bau.leben.geoeffnet != 1 {
		t.Errorf("geschlossen %d, geöffnet %d", bau.leben.geschlossen, bau.leben.geoeffnet)
	}
}

// **Wer nicht schließen kann, fängt nicht neu an.**
func TestOhneSchliessenGibtEsKeinenNeuanfang(t *testing.T) {
	bau := neuerWegbau(t)
	bau.leben.schliessfehler = errors.New("die Verbindung hängt")

	if _, fehler := bau.weg.Neuanfang(t.Context()); fehler == nil {
		t.Fatal("der Neuanfang lief ohne Schließen")
	}

	if _, statfehler := os.Stat(filepath.Join(bau.ort, "bestand.db")); statfehler != nil {
		t.Errorf("der Bestand wurde trotzdem bewegt: %v", statfehler)
	}
}

// **Und wer nicht allein ist, auch nicht** — eine Datei umzubenennen, die ein
// anderer Prozess offen hält, nennt SQLite undefiniert.
func TestOhneAlleinanspruchGibtEsKeinenNeuanfang(t *testing.T) {
	bau := neuerWegbau(t)

	bau.weg.allein = func(context.Context, string) (func() error, error) {
		return nil, errors.New("jemand hält die Datei")
	}

	if _, fehler := bau.weg.Neuanfang(t.Context()); fehler == nil {
		t.Fatal("der Neuanfang lief ohne Alleinanspruch")
	}

	// **Die Anwendung bekommt ihre Datenbank zurück.**
	if bau.leben.geoeffnet != 1 {
		t.Errorf("wieder geöffnet wurde %d-mal", bau.leben.geoeffnet)
	}
}

// **Geht das Wiederöffnen schief, wird zurückgedreht.**
func TestEinGescheitertesWiederoeffnenDrehtDenNeuanfangZurueck(t *testing.T) {
	bau := neuerWegbau(t)
	bau.leben.oeffnungsfehler = errors.New("die frische Datei taugt nicht")

	if _, fehler := bau.weg.Neuanfang(t.Context()); fehler == nil {
		t.Fatal("das gescheiterte Öffnen kam durch")
	}

	inhalt, lesefehler := os.ReadFile(filepath.Join(bau.ort, "bestand.db"))
	if lesefehler != nil {
		t.Fatalf("den zurückgelegten Bestand lesen: %v", lesefehler)
	}

	if string(inhalt) != "laufender bestand" {
		t.Errorf("zurückgelegt wurde %q", inhalt)
	}
}

// **Ohne Bilderort gibt es keinen Neuanfang** — und der Bestand kommt zurück.
func TestOhneBilderortDrehtDerNeuanfangZurueck(t *testing.T) {
	bau := neuerWegbau(t)

	bau.weg.orte.Bilder = func() (string, error) {
		return "", errors.New("kein Zuhause")
	}

	if _, fehler := bau.weg.Neuanfang(t.Context()); fehler == nil {
		t.Fatal("der Neuanfang lief ohne Bilderort")
	}

	inhalt, lesefehler := os.ReadFile(filepath.Join(bau.ort, "bestand.db"))
	if lesefehler != nil || string(inhalt) != "laufender bestand" {
		t.Errorf("der Bestand kam nicht zurück: %q, %v", inhalt, lesefehler)
	}
}

// **Ohne Ort für die Marke gibt es keinen Neuanfang** — sie beschreibt den
// zweiten Schritt, und ohne sie bliebe ein Abbruch dazwischen unlesbar.
func TestOhneOrtFuerDieMarkeGibtEsKeinenNeuanfang(t *testing.T) {
	bau := neuerWegbau(t)

	// **Ein Verzeichnis, das es nicht gibt** — dort lässt sich nichts
	// schreiben, und `bestandBeiseite` kommt gar nicht erst dran.
	bau.weg.orte.Datenbank = filepath.Join(bau.ort, "gibtesnicht", "bestand.db")

	if _, fehler := bau.weg.Neuanfang(t.Context()); fehler == nil {
		t.Fatal("der Neuanfang lief ohne Marke")
	}

	// **Und die Anwendung bekommt ihre Datenbank zurück.**
	if bau.leben.geoeffnet != 1 {
		t.Errorf("wieder geöffnet wurde %d-mal", bau.leben.geoeffnet)
	}
}

// **Ein Bestand, der sich nicht wegbewegen lässt, dreht zurück.**
func TestEinUnbeweglicherBestandDrehtDenNeuanfangZurueck(t *testing.T) {
	bau := neuerWegbau(t)

	// **Ein Verzeichnis unter dem Namen der Rückspielsicherung** — dorthin
	// lässt sich keine Datei umbenennen.
	if fehler := os.MkdirAll(
		filepath.Join(bau.ort, Rueckspielsicherung, "innen"), 0o750); fehler != nil {
		t.Fatalf("das Hindernis anlegen: %v", fehler)
	}

	if _, fehler := bau.weg.Neuanfang(t.Context()); fehler == nil {
		t.Fatal("der Neuanfang lief trotz des Hindernisses")
	}

	inhalt, lesefehler := os.ReadFile(filepath.Join(bau.ort, "bestand.db"))
	if lesefehler != nil || string(inhalt) != "laufender bestand" {
		t.Errorf("der Bestand kam nicht zurück: %q, %v", inhalt, lesefehler)
	}
}

// **Und eine Marke ohne Satz bekommt einen** — statt leer zu bleiben.
//
// **Paketintern geprüft**, weil `satzZuHindernis` unexportiert ist: der Weg
// erzeugt heute genau zwei Marken, und der Rückfall ist die Zusage für die
// dritte, die jemand hinzufügt, ohne einen Satz zu schreiben (§10).
func TestEineMarkeOhneSatzBekommtEinen(t *testing.T) {
	t.Parallel()

	if satz := satzZuHindernis(domain.Hindernis{Art: "zeitreise"}); satz == "" {
		t.Error("eine unbekannte Marke bleibt ohne Satz")
	}
}
