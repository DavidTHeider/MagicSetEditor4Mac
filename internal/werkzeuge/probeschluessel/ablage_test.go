package probeschluessel_test

// **Eine Vorrichtung für die Sichtprüfung** — kein Wächter.
//
// Sie legt einen Probeschlüssel in einen **umgelenkten** Bund
// (`MSE_SCHLUESSELBUND`) und läuft nur, wenn `MSE_PROBEADRESSE` gesetzt ist.
// Ohne die Variable überspringt sie — damit sie in der Pipeline nichts tut
// und den Anmeldebund dieses Rechners nie anfasst.

import (
	"context"
	"os"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

func TestLegeProbeschluessel(t *testing.T) {
	adresse := os.Getenv("MSE_PROBEADRESSE")
	if adresse == "" {
		t.Skip("ohne MSE_PROBEADRESSE nichts zu tun")
	}

	if os.Getenv(schluesselbund.Bundort) == "" {
		t.Fatal("ohne MSE_SCHLUESSELBUND ginge das in den Anmeldebund " +
			"dieses Rechners — das tut eine Vorrichtung nicht")
	}

	konto := domain.ZugangKontoFuer(domain.DienstOpenAI,
		domain.ZugangSchluessel, adresse)

	if err := schluesselbund.Anmeldebund().Schreibe(context.Background(),
		app.Schluesselbunddienst, konto, "probe-schluessel"); err != nil {
		t.Fatalf("ablegen: %v", err)
	}

	t.Logf("abgelegt unter %s", konto)
}
