//go:build darwin

// **Drei künstliche Fälle, und keiner erreicht das Paket** — OPS04
// (#2350):
// ein Token, ein privater Schlüssel und die persönlichen Angaben eines
// Nutzers — je im Bestand oder im `grund` einer Protokollzeile, wo ein Fehler
// aus dem Unterbau sie trüge.
//
// # Was er nicht hält
//
// **Einen Namen in freiem Text erkennt keine Gestalt.** Steht er im `grund`,
// geht er mit; die Absicherung ist hier die Auswahl — das Paket liest die
// Nutzer nicht —, und die Meldungen der Anwendung nennen keinen.
package app_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// privaterSchluessel gibt einen frischen Schlüssel im PEM-Format und seinen
// Rumpf — die Zeilen, die nicht hinausgehen dürfen.
func privaterSchluessel(t *testing.T) (pemText string, rumpf []string) {
	t.Helper()

	_, schluessel, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	roh, err := x509.MarshalPKCS8PrivateKey(schluessel)
	if err != nil {
		t.Fatal(err)
	}

	pemText = string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: roh}))

	for _, zeile := range strings.Split(pemText, "\n") {
		if zeile != "" && !strings.HasPrefix(zeile, "-----") {
			rumpf = append(rumpf, zeile)
		}
	}

	return pemText, rumpf
}

func TestDreiKuenstlicheFaelleErreichenDasDiagnosepaketNicht(t *testing.T) {
	anwendung, _ := gestarteteAnwendung(t)
	schreibeInsProtokoll(t)
	ctx := t.Context()

	// **1 — ein Token**: als Marke des Dienstes und in einer Kopfzeile.
	t.Setenv("MSE_MARKE", marker)
	slog.Default().Error("ein Ruf wurde abgewiesen", "grund", "Authorization: Bearer "+marker)

	// **2 — ein privater Schlüssel** im Grund, wie ein Fehler beim Lesen eines
	// Zeugnisses ihn trüge.
	schluessel, rumpf := privaterSchluessel(t)
	slog.Default().Error("das Zeugnis ließ sich nicht lesen", "grund", schluessel)

	// **3 — persönliche Angaben** im Bestand und eine Adresse im Grund.
	neu := anwendung.NutzerAnlegen(ctx, "Erika Mustermann", "")
	if neu.Fehler != nil || neu.Nutzer == nil {
		t.Fatalf("NutzerAnlegen: %+v", neu)
	}

	if geaendert := anwendung.NutzerAendern(ctx, app.NutzerangabenDTO{
		ID: neu.Nutzer.ID, Name: "Erika Mustermann", Vorname: "Erika", Nachname: "Mustermann",
		Kurzebio: "wohnt im Musterweg 12 in 12345 Musterstadt", Fassung: neu.Nutzer.Fassung,
	}); geaendert.Fehler != nil {
		t.Fatalf("NutzerAendern: %+v", geaendert.Fehler)
	}

	slog.Default().Error("die Anmeldung ließ sich nicht zuordnen", "grund", "erika.mustermann4711@beispiel.invalid")

	vorschau := anwendung.Diagnosepaket(ctx)
	if vorschau.Fehler != nil {
		t.Fatalf("Diagnosepaket: %+v", vorschau.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "drei"+domain.Diagnosepaketendung)
	if geschrieben := anwendung.DiagnosepaketNach(ctx, ziel); geschrieben.Fehler != nil {
		t.Fatalf("schreiben: %+v", geschrieben.Fehler)
	}

	inhalte := archivinhalte(t, ziel)

	verboten := append([]string{marker, "Mustermann", "Musterweg", "erika.mustermann4711@beispiel.invalid"}, rumpf...)
	for name, inhalt := range inhalte {
		for _, geheim := range verboten {
			if strings.Contains(inhalt, geheim) {
				t.Errorf("%s trägt %q", name, geheim)
			}
		}
	}

	// **Die Gegenprobe, dass die Zeilen ankamen** — sonst bewiese das Fehlen
	// nichts (§10).
	if protokoll := inhalte["protokoll.txt"]; !strings.Contains(protokoll, "das Zeugnis ließ sich nicht lesen") ||
		!strings.Contains(protokoll, "die Anmeldung ließ sich nicht zuordnen") {
		t.Errorf("die Zeilen stehen gar nicht im Paket:\n%s", protokoll)
	}
}
