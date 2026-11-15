package netz_test

// **Ein Kennwort aus der Umgebung steht nicht im Fehler** — Befund N3 des
// Nachlaufs über Phase 6.7 (NET-Teil): `HTTPS_PROXY` darf Zugangsdaten
// tragen, und der Hinweis auf den Proxy setzte die Adresse ganz in die
// Meldung — von dort ins Protokoll und in die Diagnose.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

func TestEinProxykennwortStehtNichtImFehler(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://anna:geheim@127.0.0.1:1")
	t.Setenv("HTTPS_PROXY", "")
	t.Setenv("NO_PROXY", "")

	_, fehler := netz.Hole(t.Context(), netz.NeuerKlientMit(netz.UeblicheFristen(), nil), nil,
		netz.Kenner{Name: "probe", Annahme: "*/*"}, "http://beispiel.invalid/", 1024)
	if fehler == nil {
		t.Fatal("die Anfrage über einen toten Proxy glückte")
	}

	if strings.Contains(fehler.Error(), "geheim") || !strings.Contains(fehler.Error(), "127.0.0.1:1") {
		t.Errorf("die Meldung: %v", fehler)
	}
}
