package service

// **Sichern und Zurückspielen mit Kennwort** — OPS-06.AK-02
// (#1723).

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

const testkennwort = "korrekt pferd batterie heftklammer"

// feststehendeChiffre sagt nur, ob der Bestand verschlüsselt ist.
type feststehendeChiffre struct{ verschluesselt bool }

func (c feststehendeChiffre) IstVerschluesselt(string) (bool, error) { return c.verschluesselt, nil }
func (feststehendeChiffre) NeuerSchluessel() string                  { return "" }
func (feststehendeChiffre) Verschluesseln(context.Context, string, string) error {
	return nil
}
func (feststehendeChiffre) Entschluesseln(context.Context, string) error      { return nil }
func (feststehendeChiffre) SchluesselSchreiben(context.Context, string) error { return nil }
func (feststehendeChiffre) SchluesselLoeschen(context.Context) error          { return nil }

func feldfehlerAmKennwort(t *testing.T, fehler error) {
	t.Helper()

	var feld *domain.FieldError
	if !errors.As(fehler, &feld) || feld.Field != "kennwort" {
		t.Fatalf("erwartet Befund am Kennwort, bekam %v", fehler)
	}
}

func TestEineSicherungMitKennwortGehtHinUndZurueck(t *testing.T) {
	t.Cleanup(sicherung.ArbeitsfaktorSetzen(10))

	bau := neuerWegbau(t)
	ziel := filepath.Join(t.TempDir(), "s"+sicherung.Endung)
	mit := MitSicherungswunsch(t.Context(), Sicherungswunsch{Kennwort: testkennwort})

	if fehler := bau.weg.Sichern(mit, ziel); fehler != nil {
		t.Fatalf("sichern: %v", fehler)
	}

	if !sicherung.IstVerschluesselt(ziel) {
		t.Fatal("die Sicherung ist nicht verschlüsselt")
	}

	// **Der Klartext entsteht neben dem Bestand und nicht am Ziel** — ein
	// USB-Stick behielte eine gelöschte Datei lesbar.
	if zwischen := filepath.Dir(filepath.Dir(bau.quelle.ziele[0])); zwischen != bau.ort {
		t.Errorf("der Zwischenstand lag unter %q, nicht neben dem Bestand %q", zwischen, bau.ort)
	}

	// **Ohne Kennwort und mit falschem** — am Feld, nicht als „beschädigt".
	_, fehler := bau.weg.Beschauen(t.Context(), ziel)
	feldfehlerAmKennwort(t, fehler)

	_, fehler = bau.weg.Beschauen(MitSicherungswunsch(t.Context(),
		Sicherungswunsch{Kennwort: "ein ganz anderes kennwort"}), ziel)
	feldfehlerAmKennwort(t, fehler)

	schau, fehler := bau.weg.Beschauen(mit, ziel)
	if fehler != nil || !schau.Verschluesselt {
		t.Fatalf("beschauen: %+v, %v", schau, fehler)
	}

	if _, rueckfehler := bau.weg.Zurueckspielen(mit, ziel); rueckfehler != nil {
		t.Fatalf("zurückspielen: %v", rueckfehler)
	}

	inhalt, fehler := os.ReadFile(filepath.Join(bau.ort, "bestand.db")) // #nosec G304 -- aus t.TempDir()
	if fehler != nil || string(inhalt) != "gesicherter bestand" {
		t.Errorf("der Bestand ist %q, %v", inhalt, fehler)
	}
}

func TestEinKurzesKennwortSichertNicht(t *testing.T) {
	bau := neuerWegbau(t)
	ziel := filepath.Join(t.TempDir(), "s"+sicherung.Endung)

	fehler := bau.weg.Sichern(MitSicherungswunsch(t.Context(),
		Sicherungswunsch{Kennwort: "zu kurz"}), ziel)
	feldfehlerAmKennwort(t, fehler)

	if _, statfehler := os.Stat(ziel); statfehler == nil || len(bau.quelle.ziele) != 0 {
		t.Error("trotz Ablehnung wurde gesichert")
	}
}

// **Ein verschlüsselter Bestand gibt keinen Klartext her, ohne dass jemand es
// ausdrücklich will** — und dann einen, der auch so heißt.
func TestEinVerschluesselterBestandSichertNurAusdruecklichImKlartext(t *testing.T) {
	bau := neuerWegbau(t)
	bau.weg.MitBestandschiffre(feststehendeChiffre{verschluesselt: true})

	ziel := filepath.Join(t.TempDir(), "s"+sicherung.Endung)

	feldfehlerAmKennwort(t, bau.weg.Sichern(t.Context(), ziel))

	if _, statfehler := os.Stat(ziel); statfehler == nil {
		t.Fatal("trotz Ablehnung liegt eine Sicherung da")
	}

	if fehler := bau.weg.Sichern(MitSicherungswunsch(t.Context(),
		Sicherungswunsch{Klartext: true}), ziel); fehler != nil {
		t.Fatalf("ausdrücklich im Klartext: %v", fehler)
	}

	if !sicherung.IstArchiv(ziel) {
		t.Error("die ausdrückliche Klartextsicherung ist kein Archiv")
	}

	schau, fehler := bau.weg.Beschauen(t.Context(), ziel)
	if fehler != nil || schau.Verschluesselt {
		t.Errorf("beschauen: %+v, %v", schau, fehler)
	}
}
