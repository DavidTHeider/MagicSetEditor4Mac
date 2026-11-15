package app

// **Ein Knopf richtet alles ein, einer nimmt es zurück** — G08
// (#2272).
//
// Eingerichtet wird, was es auf diesem Rechner gibt und was noch fehlt
// (`offen`, `veraltet`); ein **fremder** Eintrag wird nie überschrieben,
// sondern benannt. Jedes Ziel sagt, was ihm geschah, ein Teilfehler nimmt die
// übrigen nicht mit, und ein Abbruch fängt kein weiteres an.
//
// **Der Befehlsweg ist eingesetzt**: das Ziel `dienst` ruft `launchctl`, und
// den LaunchAgent dieses Rechners fasst ein Lauf nicht an (§3).

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type befehlsbuch struct {
	schloss sync.Mutex
	befehle []string
}

func (b *befehlsbuch) ausfuehren(_ context.Context, befehl []string) error {
	b.schloss.Lock()
	defer b.schloss.Unlock()

	b.befehle = append(b.befehle, strings.Join(befehl, " "))

	return nil
}

func ausgangVon(antwort SchnittstellenAlleAntwort, name string) SchnittstellenausgangDTO {
	for _, eines := range antwort.Ausgaenge {
		if eines.Name == name {
			return eines
		}
	}

	return SchnittstellenausgangDTO{}
}

func TestAllesEinrichtenUndAllesZuruecknehmen(t *testing.T) {
	t.Parallel()

	anwendung, heim := eineEinrichtende(t, "/usr/bin:/bin")
	buch := &befehlsbuch{}
	anwendung.MitSchnittstellenbefehlen(buch.ausfuehren)

	// **Codex trägt den Eintrag schon — von Hand**: das ist fremd.
	fremd := []byte("[mcp_servers.magicseteditor4mac]\ncommand = \"/eigenes/mse-mcp\"\n")
	codex := filepath.Join(heim, ".codex", "config.toml")

	if fehler := os.WriteFile(codex, fremd, 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	ctx := context.Background()

	eingerichtet := anwendung.SchnittstellenAlleEinrichten(ctx)
	if eingerichtet.Fehler != nil {
		t.Fatalf("SchnittstellenAlleEinrichten: %+v", eingerichtet.Fehler)
	}

	for name, erwartet := range map[string]string{
		"befehlszeile": "eingerichtet", "dienst": "eingerichtet", "codex": "fremd",
		"claudedesktop": "unveraendert", "claudecode": "unveraendert",
	} {
		if ausgang := ausgangVon(eingerichtet, name); ausgang.Ausgang != erwartet || ausgang.Fehler != nil {
			t.Errorf("%s: %+v, erwartet %s", name, ausgang, erwartet)
		}
	}

	if roh, _ := os.ReadFile(codex); !bytes.Equal(roh, fremd) { //nolint:errcheck,gosec // eine Datei des Laufs
		t.Errorf("der fremde Eintrag wurde angefasst: %q", roh)
	}

	if standVon(SchnittstellenAntwort{Schnittstellen: eingerichtet.Schnittstellen}, "befehlszeile").Stand != "eingerichtet" {
		t.Errorf("der Stand danach: %+v", eingerichtet.Schnittstellen)
	}

	zurueck := anwendung.SchnittstellenAlleWiderrufen(ctx)
	if zurueck.Fehler != nil {
		t.Fatalf("SchnittstellenAlleWiderrufen: %+v", zurueck.Fehler)
	}

	for name, erwartet := range map[string]string{
		"befehlszeile": "widerrufen", "dienst": "widerrufen", "codex": "fremd",
	} {
		if ausgang := ausgangVon(zurueck, name); ausgang.Ausgang != erwartet {
			t.Errorf("zurück %s: %+v, erwartet %s", name, ausgang, erwartet)
		}
	}

	if _, fehler := os.Stat(filepath.Join(heim, ".zprofile")); !os.IsNotExist(fehler) {
		t.Errorf(".zprofile steht noch: %v", fehler)
	}

	if befehle := strings.Join(buch.befehle, "\n"); !strings.Contains(befehle, "launchctl bootstrap") ||
		!strings.Contains(befehle, "launchctl bootout") {
		t.Errorf("die Befehle waren %q", befehle)
	}
}

func TestEinTeilfehlerNimmtDieAnderenNichtMit(t *testing.T) {
	t.Parallel()

	anwendung, heim := eineEinrichtende(t, "/usr/bin:/bin")
	anwendung.MitSchnittstellenbefehlen((&befehlsbuch{}).ausfuehren)

	// **In `~/.codex` lässt sich nichts schreiben** — Codex scheitert, die
	// Befehlszeile davor nicht.
	codex := filepath.Join(heim, ".codex")
	if fehler := os.Chmod(codex, 0o500); fehler != nil { //nolint:gosec // ein Ordner des Laufs, absichtlich gesperrt
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = os.Chmod(codex, 0o700) }) //nolint:errcheck,gosec // damit der Lauf aufräumen kann

	antwort := anwendung.SchnittstellenAlleEinrichten(context.Background())

	if ausgang := ausgangVon(antwort, "codex"); ausgang.Ausgang != "gescheitert" || ausgang.Fehler == nil {
		t.Errorf("codex: %+v", ausgang)
	}

	if ausgang := ausgangVon(antwort, "befehlszeile"); ausgang.Ausgang != "eingerichtet" {
		t.Errorf("befehlszeile neben dem Teilfehler: %+v", ausgang)
	}

	if ausgang := ausgangVon(antwort, "dienst"); ausgang.Ausgang != "eingerichtet" {
		t.Errorf("dienst nach dem Teilfehler: %+v", ausgang)
	}

	if antwort.Fehler == nil {
		t.Error("ein Teilfehler steht nicht als Fehler der Antwort da")
	}
}

func TestEinAbbruchFaengtKeinWeiteresZielAn(t *testing.T) {
	t.Parallel()

	anwendung, heim := eineEinrichtende(t, "/usr/bin:/bin")
	anwendung.MitSchnittstellenbefehlen((&befehlsbuch{}).ausfuehren)

	ctx, abbrechen := context.WithCancel(context.Background())
	abbrechen()

	antwort := anwendung.SchnittstellenAlleEinrichten(ctx)
	if len(antwort.Ausgaenge) == 0 {
		t.Fatal("kein Ausgang — der Lauf prüfte nichts")
	}

	for _, ausgang := range antwort.Ausgaenge {
		if ausgang.Ausgang != "abgebrochen" && ausgang.Ausgang != "unveraendert" {
			t.Errorf("nach dem Abbruch: %+v", ausgang)
		}
	}

	if _, fehler := os.Stat(filepath.Join(heim, ".zprofile")); !os.IsNotExist(fehler) {
		t.Errorf("nach dem Abbruch wurde geschrieben: %v", fehler)
	}
}
