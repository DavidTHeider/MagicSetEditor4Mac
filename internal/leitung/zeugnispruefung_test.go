package leitung_test

// **Ein eigenes Zeugnis wird vor dem Start geprüft** — NET-03.AK-02
// (#1707):
// „Zertifikat/Schlüssel-Passung, SAN, Gültigkeit, Kettenbildung und
// Dateirechte werden vor dem Start geprüft."
//
// **Jeder Fall mit seinem Nein.** Ein Dienst, der mit einem falschen Paar
// startet, fällt erst beim ersten Klienten auf — und der sieht nur, dass der
// Handschlag scheitert, nicht warum.

import (
	"crypto/sha256"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/leitung"
)

func TestEinGueltigesZeugnisBestehtDiePruefung(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)

	_, befund, fehler := leitung.PruefeZeugnis(pki.dateien, "192.0.2.10", time.Now())
	if fehler != nil {
		t.Fatalf("PruefeZeugnis: %v", fehler)
	}

	summe := sha256.Sum256(pki.blatt.Raw)
	teile := make([]string, 0, len(summe))
	for _, byteWert := range summe {
		teile = append(teile, fmt.Sprintf("%02X", byteWert))
	}

	// **In der Form von `openssl x509 -fingerprint -sha256`** — damit ein
	// Mensch ihn gegen das halten kann, was sein Werkzeug sagt.
	if befund.Fingerabdruck != strings.Join(teile, ":") {
		t.Errorf("der Fingerabdruck ist %q", befund.Fingerabdruck)
	}

	if !befund.GueltigBis.Equal(pki.blatt.NotAfter) {
		t.Errorf("gültig bis %v, erwartet %v", befund.GueltigBis, pki.blatt.NotAfter)
	}

	if befund.Aussteller != "Haus-Zwischenstelle" {
		t.Errorf("der Aussteller ist %q", befund.Aussteller)
	}

	if strings.Join(befund.Namen, ",") != "dienst.beispiel.test,192.0.2.10" {
		t.Errorf("die Namen sind %v", befund.Namen)
	}
}

// **Eine Bindung an jede Schnittstelle** kann keinen Namen verlangen — nur,
// dass es einen gibt.
func TestJedeSchnittstelleVerlangtNurEinenNamen(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)

	for _, adresse := range []string{"", "0.0.0.0", "::"} {
		if _, _, fehler := leitung.PruefeZeugnis(pki.dateien, adresse, time.Now()); fehler != nil {
			t.Errorf("%q: %v", adresse, fehler)
		}
	}
}

// abgewiesen hält fest, dass die Prüfung mit einer Validierung und einem
// Satz scheitert, der das Nötige nennt.
func abgewiesen(t *testing.T, fehler error, erwartet string) {
	t.Helper()

	if fehler == nil {
		t.Fatalf("es kam durch; erwartet war %q", erwartet)
	}

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("keine Validierung: %v", fehler)
	}

	if !strings.Contains(fehler.Error(), erwartet) {
		t.Errorf("der Satz %q nennt %q nicht", fehler.Error(), erwartet)
	}
}

func TestEinFremderSchluesselPasstNicht(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)
	schreibe(t, pki.dateien.Schluessel, schluesselAlsPEM(t, einSchluessel(t)), 0o600)

	_, _, fehler := leitung.PruefeZeugnis(pki.dateien, "", time.Now())
	abgewiesen(t, fehler, "passt nicht")
}

func TestEinAbgelaufenesZeugnisStartetNicht(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)

	_, _, fehler := leitung.PruefeZeugnis(pki.dateien, "", pki.blatt.NotAfter.Add(time.Minute))
	abgewiesen(t, fehler, "abgelaufen")

	_, _, fehler = leitung.PruefeZeugnis(pki.dateien, "", pki.blatt.NotBefore.Add(-time.Minute))
	abgewiesen(t, fehler, "noch nicht gültig")
}

func TestDieBindungsadresseMussImZeugnisStehen(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)

	_, _, fehler := leitung.PruefeZeugnis(pki.dateien, "192.0.2.99", time.Now())
	abgewiesen(t, fehler, "192.0.2.99")
}

func TestEinZeugnisOhneNamenStartetNicht(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)

	blatt, schluessel := einBlatt(t, blattwunsch{
		von: time.Now().Add(-time.Hour), bis: time.Now().Add(time.Hour),
		aussteller: pki.zwischen,
	})
	schreibe(t, pki.dateien.Zertifikat, alsPEM(blatt), 0o644)
	schreibe(t, pki.dateien.Schluessel, schluesselAlsPEM(t, schluessel), 0o600)

	_, _, fehler := leitung.PruefeZeugnis(pki.dateien, "", time.Now())
	abgewiesen(t, fehler, "keinen Namen")
}

// **Die Kette reicht bis zum Anker** — und ohne Zwischenstelle nicht.
func TestDieKetteMussBisZumAnkerReichen(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)

	ohneKette := pki.dateien
	ohneKette.Kette = ""

	_, _, fehler := leitung.PruefeZeugnis(ohneKette, "", time.Now())
	abgewiesen(t, fehler, "Kette")

	// **Und die Kette darf im Zertifikat mitstehen**, wie es viele Stellen
	// ausliefern.
	schreibe(t, pki.dateien.Zertifikat, alsPEM(pki.blatt, pki.zwischen.zertifikat), 0o644)

	if _, _, fehler := leitung.PruefeZeugnis(ohneKette, "", time.Now()); fehler != nil {
		t.Errorf("mit der Kette im Zertifikat: %v", fehler)
	}
}

func TestEinFremderAnkerTraegtNicht(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)
	fremde := eineStelle(t, "Fremde Wurzel", nil)
	schreibe(t, pki.dateien.Anker, alsPEM(fremde.zertifikat), 0o644)

	_, _, fehler := leitung.PruefeZeugnis(pki.dateien, "", time.Now())
	abgewiesen(t, fehler, "Kette")
}

// **Ohne eigenen Anker gelten die des Systems** — und ein Haus-Anker steht
// dort nicht.
func TestOhneAnkerGeltenDieDesSystems(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)

	ohneAnker := pki.dateien
	ohneAnker.Anker = ""

	_, _, fehler := leitung.PruefeZeugnis(ohneAnker, "", time.Now())
	abgewiesen(t, fehler, "Kette")
}

func TestEinLesbarerSchluesselStartetNicht(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)

	if fehler := os.Chmod(pki.dateien.Schluessel, 0o644); fehler != nil { //nolint:gosec // der Fall, den der Lauf abweisen soll
		t.Fatal(fehler)
	}

	_, _, fehler := leitung.PruefeZeugnis(pki.dateien, "", time.Now())
	abgewiesen(t, fehler, "0600")
}

func TestEineFehlendeDateiWirdGenannt(t *testing.T) {
	t.Parallel()

	for name, verbiegen := range map[string]func(*leitung.Zeugnisdateien){
		"Zertifikat": func(d *leitung.Zeugnisdateien) { d.Zertifikat += ".fort" },
		"Schlüssel":  func(d *leitung.Zeugnisdateien) { d.Schluessel += ".fort" },
		"Kette":      func(d *leitung.Zeugnisdateien) { d.Kette += ".fort" },
		"Anker":      func(d *leitung.Zeugnisdateien) { d.Anker += ".fort" },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			pki := eineGueltigePKI(t)
			verbiegen(&pki.dateien)

			_, _, fehler := leitung.PruefeZeugnis(pki.dateien, "", time.Now())
			abgewiesen(t, fehler, ".fort")
		})
	}
}

// **Eine Datei ohne PEM** ist kein Zertifikat — auch wenn sie lesbar ist.
func TestEineDateiOhneZertifikatWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)
	schreibe(t, pki.dateien.Anker, []byte("kein PEM"), 0o644)

	_, _, fehler := leitung.PruefeZeugnis(pki.dateien, "", time.Now())
	abgewiesen(t, fehler, "kein Zertifikat")

	pki = eineGueltigePKI(t)
	schreibe(t, pki.dateien.Zertifikat, []byte("kein PEM"), 0o644)

	_, _, fehler = leitung.PruefeZeugnis(pki.dateien, "", time.Now())
	abgewiesen(t, fehler, "kein Zertifikat")
}

// **Ohne Zertifikat oder Schlüssel ist es keine eigene PKI.**
func TestOhneZertifikatOderSchluesselGibtEsKeineEigene(t *testing.T) {
	t.Parallel()

	if (leitung.Zeugnisdateien{Kette: "k", Anker: "a"}).Eigene() {
		t.Error("ohne die zwei Pflichtdateien gilt es als eigene")
	}

	if !(leitung.Zeugnisdateien{Zertifikat: "z", Schluessel: "s"}).Eigene() {
		t.Error("mit beiden gilt es nicht als eigene")
	}

	_, _, fehler := leitung.PruefeZeugnis(leitung.Zeugnisdateien{Zertifikat: "z"}, "", time.Now())
	abgewiesen(t, fehler, "Schlüssel")
}

// **Was sonst noch in einer PEM-Datei stehen kann** — ein Schlüssel hinter
// dem Zertifikat, wie manche Werkzeuge beides in eine Datei legen, wird
// übergangen; ein Block, der „CERTIFICATE" heißt und keines ist, nicht.
func TestFremdeBloeckeUndKaputteZertifikate(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)
	mitSchluessel := append(alsPEM(pki.blatt), schluesselAlsPEM(t, einSchluessel(t))...)
	schreibe(t, pki.dateien.Zertifikat, mitSchluessel, 0o644)

	if _, _, fehler := leitung.PruefeZeugnis(pki.dateien, "", time.Now()); fehler != nil {
		t.Errorf("mit einem Schlüssel hinter dem Zertifikat: %v", fehler)
	}

	pki = eineGueltigePKI(t)
	kaputt := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("kein DER")})
	schreibe(t, pki.dateien.Kette, kaputt, 0o644)

	_, _, fehler := leitung.PruefeZeugnis(pki.dateien, "", time.Now())
	abgewiesen(t, fehler, "sich nicht lesen")

	pki = eineGueltigePKI(t)
	schreibe(t, pki.dateien.Kette, []byte("kein PEM"), 0o644)

	_, _, fehler = leitung.PruefeZeugnis(pki.dateien, "", time.Now())
	abgewiesen(t, fehler, "kein Zertifikat")
}

// **Ein Schlüssel, den nicht einmal sein Eigentümer lesen darf**, besteht
// die Rechteprüfung und scheitert am Lesen — mit dem Pfad im Satz.
func TestEinUnlesbarerSchluesselWirdGenannt(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)

	if fehler := os.Chmod(pki.dateien.Schluessel, 0o200); fehler != nil {
		t.Fatal(fehler)
	}

	_, _, fehler := leitung.PruefeZeugnis(pki.dateien, "", time.Now())
	abgewiesen(t, fehler, pki.dateien.Schluessel)
}

// **Eine Datei über der Grenze wird nicht gelesen** — ein Zeugnis ist ein
// paar Kilobyte, und ein Pfad, der auf etwas anderes zeigt, soll den Dienst
// nicht mit einem Gigabyte aufhalten.
func TestEineZuGrosseDateiWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	pki := eineGueltigePKI(t)
	gross := make([]byte, leitung.Zeugnisgrenze+1)
	schreibe(t, pki.dateien.Anker, gross, 0o644)

	_, _, fehler := leitung.PruefeZeugnis(pki.dateien, "", time.Now())
	abgewiesen(t, fehler, "größer als")

	// **Und ein Verzeichnis öffnet sich, lässt sich aber nicht lesen.**
	pki = eineGueltigePKI(t)
	pki.dateien.Kette = t.TempDir()

	_, _, fehler = leitung.PruefeZeugnis(pki.dateien, "", time.Now())
	abgewiesen(t, fehler, "lässt sich nicht lesen")
}
