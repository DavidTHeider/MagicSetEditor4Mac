package flaechen_test

// **Der Probeanker der gebauten Programme** — SEC03
// (#2356).
//
// `mse` und `mse-rest` prüfen eine Freigabe gegen den Anker **in ihnen**.
// Bis SEC03 signierten die Läufe deshalb mit dem echten Freigabeschlüssel
// aus dem Schlüsselbund dieses Rechners. Jetzt baut `TestMain` sie mit dem
// Tag `probeanker`, erzeugt einen Schlüssel und legt seinen öffentlichen Teil
// dorthin, wohin `MSE_PROBEANKER` zeigt — und signiert mit dem privaten.
//
// **Was das nicht prüft**: dass der eingebettete Schlüssel der richtige ist.
// Das halten `TestDerEingebetteteSchluesselIstDerRichtige` und
// `TestDieEchteFreigabesignaturPasstZumEingebettetenSchluessel` in
// `internal/aktualisierung` — der zweite gegen die veröffentlichte Signatur
// einer echten Freigabe, ohne privaten Schlüssel.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
)

// probeanker ist der Schlüssel, gegen den die gebauten Programme prüfen.
var probeanker struct {
	wer   *openpgp.Entity
	datei string
}

// probeankerAnlegen erzeugt ihn und legt den öffentlichen Teil unter `ort`.
func probeankerAnlegen(ort string) {
	wer, err := openpgp.NewEntity("Probeanker", "nur für Tests", "probeanker@beispiel.test", nil)
	if err != nil {
		panic(err)
	}

	var gepanzert strings.Builder

	schreiber, err := armor.Encode(&gepanzert, openpgp.PublicKeyType, nil)
	if err != nil {
		panic(err)
	}

	if err := wer.Serialize(schreiber); err != nil {
		panic(err)
	}

	if err := schreiber.Close(); err != nil {
		panic(err)
	}

	datei := filepath.Join(ort, "probeanker.asc")
	if err := os.WriteFile(datei, []byte(gepanzert.String()), 0o600); err != nil {
		panic(err)
	}

	probeanker.wer, probeanker.datei = wer, datei
}

// mitProbeanker gibt die Umgebung, in der ein Programm den Probeanker nimmt.
//
// **Der Name als Zeichenkette**: `aktualisierung.Probeankerumgebung` gibt es
// nur unter dem Tag, und diese Tests übersetzen ohne.
func mitProbeanker(umgebung []string) []string {
	return append(umgebung, "MSE_PROBEANKER="+probeanker.datei)
}

// probeSigniert gibt eine abgetrennte Signatur des Probeankers.
func probeSigniert(t *testing.T, text string) string {
	t.Helper()

	var gepanzert strings.Builder
	if err := openpgp.ArmoredDetachSign(&gepanzert, probeanker.wer, strings.NewReader(text), nil); err != nil {
		t.Fatalf("signieren: %v", err)
	}

	return gepanzert.String()
}
