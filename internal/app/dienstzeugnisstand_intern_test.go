package app

// **Das Fenster sieht, welches Zeugnis ein Dienst vorzeigt** — NET-03
// (#1707).
//
// Bis hierhin stand dort nur ein Fingerabdruck, und er war immer der der
// Peeridentität. Mit einer eigenen PKI nennt der Stand Aussteller und Ende —
// und wer das Zeugnis tauscht, sieht nach einem Neustart des Dienstes das
// neue.

import (
	"context"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func TestDerDienststandNenntDasEigeneZeugnis(t *testing.T) {
	t.Parallel()

	bis := time.Date(2027, 1, 2, 3, 4, 5, 0, time.UTC)
	anwendung, _, ctx := einePostenhalterin(t, "https://0.0.0.0:8443", "geheim", "AB:CD", nil)

	anwendung.MitDienststarter(func(context.Context, domain.Dienstart, []domain.Bindung) (dienstlauf, error) {
		lauf := probelauf("https://0.0.0.0:8443", "geheim", "AB:CD")
		lauf.aussteller, lauf.gueltigBis = "Haus-Stelle", bis

		return lauf, nil
	})

	if _, fehler := anwendung.dienstStarten(ctx, domain.DienstREST,
		domain.Bindung{Adresse: "0.0.0.0", Port: 8443}); fehler != nil {
		t.Fatalf("starten: %v", fehler)
	}

	stand := anwendung.dienststand(domain.DienstREST)
	if stand.ZeugnisAussteller != "Haus-Stelle" || stand.ZeugnisBis != "2027-01-02T03:04:05Z" {
		t.Errorf("der Stand ist %+v", stand)
	}

	// **Nach dem Beenden ist es fort** — und ein Zeugnis der Identität nennt
	// keines.
	anwendung.dienstBeenden(domain.DienstREST)

	if stand := anwendung.dienststand(domain.DienstREST); stand.ZeugnisAussteller != "" || stand.ZeugnisBis != "" {
		t.Errorf("nach dem Beenden: %+v", stand)
	}
}
