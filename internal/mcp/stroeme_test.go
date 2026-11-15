package mcp_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// sperriger Strom: er gibt beim Lesen einen Fehler.
type sperrigerLeser struct{}

func (sperrigerLeser) Read([]byte) (int, error) { return 0, errors.New("die Leitung ist tot") }

// tauberSchreiber nimmt nichts an.
type tauberSchreiber struct{}

func (tauberSchreiber) Write([]byte) (int, error) { return 0, errors.New("niemand hört zu") }

// **Ein Strom, der abbricht, wird gemeldet.**
//
// Ein Server, der still weiterläuft, wenn sein Gegenüber fort ist, hält einen
// Prozess am Leben, den niemand mehr braucht.
func TestEinAbbrechenderStromWirdGemeldet(t *testing.T) {
	server := mcp.NeuerServer(nil)

	if err := server.Bediene(t.Context(), sperrigerLeser{}, io.Discard); err == nil {
		t.Error("kein Fehler")
	}
}

// **Eine Antwort, die nicht ankommt, wird gemeldet.**
//
// Wer sie verschluckt, antwortet ins Leere und wartet danach auf die nächste
// Frage, die nie kommt.
func TestEineUnzustellbareAntwortWirdGemeldet(t *testing.T) {
	server := mcp.NeuerServer(nil)

	ein := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"ping"}` + "\n")

	if err := server.Bediene(t.Context(), ein, tauberSchreiber{}); err == nil {
		t.Error("kein Fehler")
	}
}

// Leere Zeilen werden übergangen.
//
// Manche Gegenstellen schicken sie zur Trennung; daran zu scheitern wäre eine
// Strenge ohne Gewinn.
func TestLeereZeilenWerdenUebergangen(t *testing.T) {
	zeilen := sprich(t, probenbestand(), "", `{"jsonrpc":"2.0","id":1,"method":"ping"}`, "")

	if len(zeilen) != 1 {
		t.Errorf("Antworten: %v, erwartet eine", zeilen)
	}
}
