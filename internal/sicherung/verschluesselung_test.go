package sicherung_test

// **Eine Sicherung mit Kennwort** — OPS-06.AK-02
// (#1723):
// portabel, also nicht an den Schlüsselbund dieses Rechners gebunden, und mit
// einem Werkzeug außerhalb dieser Anwendung zu öffnen.

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

const probekennwort = "korrekt pferd batterie heftklammer"

// verschluesselteSicherung packt einen Teil mit Kennwort und gibt den Ort.
func verschluesselteSicherung(t *testing.T) (ziel string, inhalt []byte) {
	t.Helper()
	t.Cleanup(sicherung.ArbeitsfaktorSetzen(10))

	ort := t.TempDir()
	inhalt = []byte("ein Bestand, der niemanden etwas angeht")

	quelle := filepath.Join(ort, "bestand.db")
	if fehler := os.WriteFile(quelle, inhalt, 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	ziel = filepath.Join(ort, "s"+sicherung.Endung)
	if fehler := sicherung.PackenVerschluesselt(ziel,
		[]sicherung.Teil{{Name: "bestand.db", Quelle: quelle}},
		sicherung.Angaben{Fassung: "0.6.0"}, probekennwort); fehler != nil {
		t.Fatalf("packen: %v", fehler)
	}

	return ziel, inhalt
}

func TestEineVerschluesselteSicherungGibtMitKennwortHeraus(t *testing.T) {
	ziel, inhalt := verschluesselteSicherung(t)

	roh, fehler := os.ReadFile(ziel) // #nosec G304 -- aus t.TempDir()
	if fehler != nil {
		t.Fatal(fehler)
	}

	// **Kein Klartext auf der Platte**, und erkennbar als das, was sie ist.
	if bytes.Contains(roh, inhalt) || bytes.Contains(roh, []byte(sicherung.Manifestname)) {
		t.Error("die Sicherung trägt Klartext")
	}

	if !sicherung.IstVerschluesselt(ziel) || sicherung.IstArchiv(ziel) {
		t.Error("nicht als verschlüsselt erkannt")
	}

	platz := t.TempDir()

	beschreibung, fehler := sicherung.EntpackenVerschluesselt(ziel, platz, probekennwort)
	if fehler != nil {
		t.Fatalf("entpacken: %v", fehler)
	}

	zurueck, fehler := os.ReadFile(filepath.Join(platz, "bestand.db")) // #nosec G304 -- aus t.TempDir()
	if fehler != nil || !bytes.Equal(zurueck, inhalt) || beschreibung.Fassung != "0.6.0" {
		t.Errorf("zurück: %q, %v, %+v", zurueck, fehler, beschreibung)
	}
}

// **Portabel heißt: auch ohne diese Anwendung zu öffnen.** Das Format ist
// age (`age-encryption.org/v1`); wer das Kennwort hat, kommt mit `age -d` an
// das ZIP darin.
func TestEineVerschluesselteSicherungIstAgeMitKennwort(t *testing.T) {
	ziel, _ := verschluesselteSicherung(t)

	identitaet, fehler := age.NewScryptIdentity(probekennwort)
	if fehler != nil {
		t.Fatal(fehler)
	}

	datei, fehler := os.Open(ziel) // #nosec G304 -- aus t.TempDir()
	if fehler != nil {
		t.Fatal(fehler)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // nur gelesen

	klar, fehler := age.Decrypt(datei, identitaet)
	if fehler != nil {
		t.Fatalf("age: %v", fehler)
	}

	kopf := make([]byte, 4)
	if _, fehler := klar.Read(kopf); fehler != nil || string(kopf) != "PK\x03\x04" {
		t.Errorf("darin liegt kein ZIP: %q, %v", kopf, fehler)
	}
}

func TestEinFalschesKennwortWirdAmFeldAbgewiesen(t *testing.T) {
	ziel, _ := verschluesselteSicherung(t)
	platz := t.TempDir()

	_, fehler := sicherung.EntpackenVerschluesselt(ziel, platz, "ein anderes kennwort")

	var feld *domain.FieldError
	if !errors.As(fehler, &feld) || feld.Field != "kennwort" || !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("erwartet Feldfehler am Kennwort, bekam %v", fehler)
	}

	if eintraege, _ := os.ReadDir(platz); len(eintraege) != 0 { //nolint:errcheck // ein fehlender Ordner hat keine
		t.Errorf("trotz falschem Kennwort liegt etwas da: %v", eintraege)
	}
}

func TestEineVerschluesselteSicherungUeberschreibtNichts(t *testing.T) {
	ziel, _ := verschluesselteSicherung(t)

	fehler := sicherung.PackenVerschluesselt(ziel, nil, sicherung.Angaben{}, probekennwort)
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("erwartet Konflikt, bekam %v", fehler)
	}
}

func TestOhneKennwortWirdNichtVerschluesselt(t *testing.T) {
	ort := t.TempDir()

	fehler := sicherung.PackenVerschluesselt(filepath.Join(ort, "s"+sicherung.Endung),
		nil, sicherung.Angaben{}, "")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("erwartet Ablehnung, bekam %v", fehler)
	}

	if eintraege, _ := os.ReadDir(ort); len(eintraege) != 0 { //nolint:errcheck // ein fehlender Ordner hat keine
		t.Errorf("es blieb etwas liegen: %v", eintraege)
	}

	if _, fehler := sicherung.EntpackenVerschluesselt(filepath.Join(ort, "fehlt"), ort, ""); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("erwartet Ablehnung ohne Kennwort, bekam %v", fehler)
	}
}

// **Eine Datei, die keine verschlüsselte Sicherung ist**, wird genauso
// abgewiesen wie eine beschädigte — und die Bytes eines Archivs sind keine.
func TestEineBeschaedigteVerschluesselteSicherungWirdAbgewiesen(t *testing.T) {
	ziel, _ := verschluesselteSicherung(t)

	roh, fehler := os.ReadFile(ziel) // #nosec G304 -- aus t.TempDir()
	if fehler != nil {
		t.Fatal(fehler)
	}

	roh[len(roh)-5] ^= 0xff
	if fehler := os.WriteFile(ziel, roh, 0o600); fehler != nil { //nolint:gosec // G703: ein Pfad aus t.TempDir()
		t.Fatal(fehler)
	}

	if _, fehler := sicherung.EntpackenVerschluesselt(ziel, t.TempDir(), probekennwort); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("beschädigt: erwartet Ablehnung, bekam %v", fehler)
	}

	if sicherung.IstVerschluesselt(filepath.Join(t.TempDir(), "fehlt")) {
		t.Error("eine fehlende Datei gilt als verschlüsselt")
	}

	if _, fehler := sicherung.EntpackenVerschluesselt(filepath.Join(t.TempDir(), "fehlt"), t.TempDir(), probekennwort); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("fehlend: erwartet Ablehnung, bekam %v", fehler)
	}
}

// **Mit dem richtigen Kennwort und ohne Archiv darin** — etwa eine Datei,
// die jemand selbst mit age verschlüsselt hat.
func TestEinVerschluesseltesNichtarchivWirdAbgewiesen(t *testing.T) {
	ziel := filepath.Join(t.TempDir(), "fremd"+sicherung.Endung)

	empfaenger, fehler := age.NewScryptRecipient(probekennwort)
	if fehler != nil {
		t.Fatal(fehler)
	}

	empfaenger.SetWorkFactor(10)

	var roh bytes.Buffer

	huelle, fehler := age.Encrypt(&roh, empfaenger)
	if fehler != nil {
		t.Fatal(fehler)
	}

	_, schreibfehler := huelle.Write([]byte("kein Archiv"))
	if fehler := errors.Join(schreibfehler, huelle.Close(), os.WriteFile(ziel, roh.Bytes(), 0o600)); fehler != nil {
		t.Fatal(fehler)
	}

	if _, fehler := sicherung.EntpackenVerschluesselt(ziel, t.TempDir(), probekennwort); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("erwartet Ablehnung, bekam %v", fehler)
	}
}
