//go:build netzlauf

// **Die Aktualisierung am Original** — #2049, Entscheidung 2 in
// `docs/INTEGRATIONEN.md`: sie fragte bis hierher nur selbst geschriebene
// Antworten.
//
// # Was er prüft
//
// **Ohne Marke das fachliche Nein**: das Repositorium ist privat, und GitHub
// antwortet mit 404 — der Klient muss daraus „keine Freigabe zu sehen"
// machen und nicht einen Ausfall. **Mit Marke** (`GH_TOKEN` oder
// `GITHUB_TOKEN`, wie bei `gh`) die echte neueste Freigabe: Fassung,
// Abbild, Prüfsummenliste und Signatur — und die Liste **mit geprüfter
// Signatur** gegen den eingebetteten Schlüssel. **Kein Abbild**: es wiegt
// Hunderte Megabyte und sagt über den Weg nichts, was die signierte Liste
// nicht sagt.
//
//	go test -count=1 -tags netzlauf -run Original -v ./internal/aktualisierung/
package aktualisierung_test

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

func echterKlient(marke string) *aktualisierung.Klient {
	return aktualisierung.NeuerKlient(aktualisierung.Adresse,
		netz.NeuerKlient(netz.UeblicheFristen()),
		netz.NeuesDrosselwerk(time.Now, time.Sleep), marke)
}

func TestOhneMarkeSiehtAmOriginalNiemandEineFreigabe(t *testing.T) {
	_, fehler := echterKlient("").NeuesteFreigabe(t.Context())
	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("erwartet das fachliche Nein des privaten Repositoriums, bekam %v", fehler)
	}

	t.Logf("ohne Marke: %v", fehler)
}

func TestDieNeuesteFreigabeAmOriginal(t *testing.T) {
	marke := aktualisierung.MarkeAusUmgebung()
	if marke == "" {
		t.Fatal("BLOCKED: weder GH_TOKEN noch GITHUB_TOKEN gesetzt — ohne Lesezugriff zeigt das private Repositorium keine Freigabe")
	}

	klient := echterKlient(marke)

	freigabe, fehler := klient.NeuesteFreigabe(t.Context())
	if fehler != nil {
		t.Fatalf("neueste Freigabe: %v", fehler)
	}

	if freigabe.Fassung == "" || freigabe.Abbild.Name == "" ||
		freigabe.Pruefsummen.Name == "" || freigabe.Signatur.Name == "" {
		t.Fatalf("unvollständige Freigabe: %+v", freigabe)
	}

	liste, fehler := klient.Pruefsummen(t.Context(), freigabe)
	if fehler != nil {
		t.Fatalf("Prüfsummen mit Signatur: %v", fehler)
	}

	summe, gefunden := aktualisierung.SummeAus(string(liste), freigabe.Abbild.Name)
	if !gefunden {
		t.Fatalf("die signierte Liste nennt %s nicht", freigabe.Abbild.Name)
	}

	t.Logf("Fassung %s vom %s, Abbild %s (%d Bytes), Summe %s — Signatur geprüft",
		freigabe.Fassung, freigabe.Erschienen.Format(time.DateOnly), freigabe.Abbild.Name,
		freigabe.Abbild.Groesse, summe)
}
