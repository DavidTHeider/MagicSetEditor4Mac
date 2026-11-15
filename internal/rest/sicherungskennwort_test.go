package rest_test

// **Das Kennwort einer Sicherung am Dienst** — OPS-06.AK-02
// (#1723):
// beim Sichern im Rumpf, beim Lesen im Kopf — **nie in der Adresse**, die in
// Protokollen und Verläufen steht.

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// kennwortprobe merkt sich, was ankam.
type kennwortprobe struct {
	williger

	ziel, kennwort string
	klartext       bool
}

func (k *kennwortprobe) BestandSichern(_ context.Context, ziel, kennwort string, klartext bool) error {
	k.ziel, k.kennwort, k.klartext = ziel, kennwort, klartext

	return nil
}

func (k *kennwortprobe) SicherungSchauen(_ context.Context, _, kennwort string) (rest.Sicherungsangaben, error) {
	k.kennwort = kennwort

	return rest.Sicherungsangaben{Verschluesselt: true}, nil
}

func (k *kennwortprobe) Ruecksprunglage(_ context.Context, _, kennwort string) (any, error) {
	k.kennwort = kennwort

	return map[string]any{}, nil
}

func TestDasKennwortKommtImRumpfUndImKopfAn(t *testing.T) {
	t.Parallel()

	probe := &kennwortprobe{}
	jetzt := func() time.Time { return time.Date(2026, time.September, 26, 12, 0, 0, 0, time.UTC) }
	dienst := mitHaushalt(t, rest.NeuerDienst(probemarke, jetzt,
		rest.Naehte{Sicherung: probe, Ruecksprung: probe}), domain.Vorgabegrenzen(), jetzt)

	antwort := ruf(t, dienst, endpunkt{
		methode: http.MethodPost, weg: "/bestand/sicherung",
		rumpf: `{"ziel":"/tmp/s.msesicherung","kennwort":"korrekt pferd batterie","klartext":true}`,
	}, gueltig)
	if antwort.Code != http.StatusCreated || probe.ziel != "/tmp/s.msesicherung" ||
		probe.kennwort != "korrekt pferd batterie" || !probe.klartext {
		t.Fatalf("sichern: %d %+v", antwort.Code, probe)
	}

	mitKopf := func(anfrage *http.Request) {
		gueltig(anfrage)
		anfrage.Header.Set(rest.Kennwortkopf, "aus dem kopf gelesen")
	}

	for _, weg := range []string{
		"/bestand/sicherung?pfad=/tmp/s.msesicherung",
		"/ruecksprung/lage?datei=/tmp/s.msesicherung",
	} {
		probe.kennwort = ""

		if antwort := ruf(t, dienst, endpunkt{methode: http.MethodGet, weg: weg}, mitKopf); antwort.Code != http.StatusOK ||
			probe.kennwort != "aus dem kopf gelesen" {
			t.Errorf("%s: %d, Kennwort %q", weg, antwort.Code, probe.kennwort)
		}
	}
}
