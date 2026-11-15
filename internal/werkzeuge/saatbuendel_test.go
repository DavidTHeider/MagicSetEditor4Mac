// Wächter über den Weg der Saat ins Bündel
// (#58).
//
// **Er prüft die Erklärung, nicht den Bau.** `scripts/bauen.sh` ruft nach dem
// Übersetzen `scripts/saat.sh`; hier steht, dass der Aufruf da ist, auf ein
// Skript zeigt, das es gibt, und dass **jeder** Weg zum Bündel über dieses
// Skript geht. Ob die Saat danach wirklich im Bündel liegt, sagt nur ein Bau —
// und ein Bau läuft nicht in der Pipeline (§4). Das ist die Hälfte, die dieser
// Wächter **nicht** hält, und sie steht in `docs/RELEASE.md` als Handgriff vor
// der Freigabe.
//
// # Warum ein Bauskript und nicht der Haken von Wails
//
// Unter Wails 2 stand der Aufruf als `postBuildHooks` in `wails.json` und galt
// damit für `wails build` wie für `wails dev`. Wails 3 kennt diesen Haken
// nicht mehr (ENG-02, #1727); dieselbe Eigenschaft — **ein** Aufruf für jeden
// Weg — hält jetzt `scripts/bauen.sh`, über das Freigabe, `make bau` und
// `make dev` gleichermaßen bauen. Ein Weg daran vorbei stünde wieder ohne
// Saat da, und genau den sucht der zweite Lauf hier.
package werkzeuge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// saataufruf ist der Aufruf, den `scripts/bauen.sh` tragen muss — mit dem
// Pfad des Programms, denn nur von ihm aus weiß `saat.sh`, in welches Bündel
// die Saat gehört.
const saataufruf = `"$WURZEL/scripts/saat.sh" "$INHALT/MacOS/$NAME"`

// TestDasBauskriptLegtDieSaatInsBuendel hält die Erklärung fest.
func TestDasBauskriptLegtDieSaatInsBuendel(t *testing.T) {
	if befund := buendelbefund(string(projektdatei(t, "scripts/bauen.sh")), wurzel); befund != "" {
		t.Error(befund)
	}
}

// TestJederWegZumBuendelGehtUeberDasBauskript hält die zweite Hälfte: ein
// Weg, der das Bündel selbst legte, stünde ohne Saat da.
func TestJederWegZumBuendelGehtUeberDasBauskript(t *testing.T) {
	wege := map[string]string{
		"Makefile":              "./scripts/bauen.sh freigabe",
		"scripts/release.sh":    "./scripts/bauen.sh freigabe",
		"scripts/entwickeln.sh": `./scripts/bauen.sh "$BAU"`,
	}

	for datei, aufruf := range wege {
		t.Run(datei, func(t *testing.T) {
			if !befehlszeileMit(string(projektdatei(t, datei)), aufruf) {
				t.Errorf("%s baut nicht über %q — das Bündel entstünde ohne Saat", datei, aufruf)
			}
		})
	}
}

// TestDerBuendelwaechterWirdRot ist die Gegenprobe — als Testfall und nicht als
// Handbewegung (§10).
//
// Ohne sie hielte der Wächter womöglich gar nichts: eine Prüfung, die auch für
// ein Bauskript **ohne** Aufruf grün bliebe, sähe von außen genauso aus.
func TestDerBuendelwaechterWirdRot(t *testing.T) {
	faelle := map[string]string{
		"ohne Aufruf":                 "go build .\n",
		"nur im Kommentar":            "# " + saataufruf + "\n",
		"ohne den Pfad des Programms": `"$WURZEL/scripts/saat.sh"` + "\n",
	}

	for name, skript := range faelle {
		t.Run(name, func(t *testing.T) {
			if befund := buendelbefund(skript, wurzel); befund == "" {
				t.Error("der Wächter blieb grün")
			}
		})
	}

	// **Und ein Aufruf ins Leere** — derselbe Text in einem Projekt, in dem es
	// `scripts/saat.sh` nicht gibt.
	t.Run("mit einem Skript, das es nicht gibt", func(t *testing.T) {
		if befund := buendelbefund(saataufruf+"\n", t.TempDir()); befund == "" {
			t.Error("der Wächter blieb grün")
		}
	})
}

// buendelbefund sagt, was an einem Bauskript fehlt — leer, wenn nichts fehlt.
//
// **Als Funktion über dem Text und nicht als Prüfung an Ort und Stelle**: nur
// so lässt sich die Gegenprobe fahren, ohne die echte Datei zu verfälschen
// (§10, und das Muster aus `scripts/release.test.sh`).
func buendelbefund(skript, projektwurzel string) string {
	if !befehlszeileMit(skript, saataufruf) {
		return "scripts/bauen.sh ruft " + saataufruf +
			" nicht — die gebaute Anwendung startet ohne Saat ins Leere (#58)"
	}

	if _, err := os.Stat(filepath.Join(projektwurzel, "scripts", "saat.sh")); err != nil {
		return "scripts/bauen.sh ruft scripts/saat.sh — das Skript gibt es nicht"
	}

	return ""
}

// befehlszeileMit sagt, ob eine Zeile, die kein Kommentar ist, den Aufruf
// trägt.
func befehlszeileMit(skript, aufruf string) bool {
	for zeile := range strings.SplitSeq(skript, "\n") {
		zeile = strings.TrimSpace(zeile)
		if strings.HasPrefix(zeile, "#") {
			continue
		}

		if strings.Contains(zeile, aufruf) {
			return true
		}
	}

	return false
}
