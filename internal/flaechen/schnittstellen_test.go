package flaechen_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// **Der Stand der Schnittstellen über jede Fläche** — OPS-09.AK-03
// (#1907).
//
// **Ein fremder Codex-Eintrag ist an allen dreien fremd** — und keine liest
// das echte Zuhause: `mitEigenerSaat` gibt jedem Kindprozess ein
// Wegwerfheim, und genau dort steht die Datei.
func TestDieSchnittstellenUeberJedeFlaeche(t *testing.T) {
	umgebung := mitEigenerSaat(t)
	zuhause := zuhauseAus(t, umgebung)

	codex := filepath.Join(zuhause, ".codex")
	if fehler := os.MkdirAll(codex, 0o700); fehler != nil {
		t.Fatal(fehler)
	}

	fremd := "[mcp_servers.magicseteditor4mac]\ncommand = \"/von/hand/mse-mcp\"\n"
	if fehler := os.WriteFile(filepath.Join(codex, "config.toml"), []byte(fremd), 0o600); fehler != nil {
		t.Fatal(fehler)
	}

	an := starteDienst(t, umgebung)

	fuerAlleDreiMit(t, "der fremde Codex-Eintrag",
		[3]string{"codex\tfremd", `"stand": "fremd"`, `"stand":"fremd"`},
		func() string { return mussKommandozeile(t, umgebung, "schnittstellen") },
		func() string { return mussMCP(t, umgebung, "lesen.schnittstellen", map[string]any{}) },
		func() string { return mussREST(t, an, http.MethodGet, "/schnittstellen", "") })

	profil := filepath.Join(zuhause, ".zprofile")

	fuerAlleDreiMit(t, "das Profil im Wegwerfheim",
		[3]string{"befehlszeile\toffen\t" + profil, `"datei": "` + profil + `"`, `"datei":"` + profil + `"`},
		func() string { return mussKommandozeile(t, umgebung, "schnittstellen") },
		func() string { return mussMCP(t, umgebung, "lesen.schnittstellen", map[string]any{}) },
		func() string { return mussREST(t, an, http.MethodGet, "/schnittstellen", "") })

	// **Und widerrufen an der Kommandozeile** (OPS03, #2349): der fremde
	// Eintrag bleibt und wird genannt, die übrigen sind nichts zu tun.
	aus, fehler, code := kommandozeile(t, umgebung, strings.Fields("schnittstellen widerrufen")...)
	if code != 0 || !strings.Contains(aus, "codex\tfremd") || !strings.Contains(aus, "befehle\tunveraendert") {
		t.Errorf("widerrufen: Rückgabewert %d, %q (%s)", code, aus, fehler)
	}

	// **Gelesen ist gelesen**: keine der drei hat etwas geschrieben — und der
	// Widerruf den fremden Eintrag nicht angefasst.
	if inhalt, fehler := os.ReadFile(filepath.Join(codex, "config.toml")); fehler != nil || string(inhalt) != fremd { //nolint:gosec // eine Datei des Laufs
		t.Errorf("der fremde Eintrag wurde angefasst: %q, %v", inhalt, fehler)
	}

	if _, fehler := os.Stat(profil); !os.IsNotExist(fehler) {
		t.Errorf("ein Profil entstand: %v", fehler)
	}
}
