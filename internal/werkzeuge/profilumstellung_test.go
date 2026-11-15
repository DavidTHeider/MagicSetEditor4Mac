// **Wer das gezeigte Profil umstellt, geht durch einen Weg** — Nachlauf zum
// Zweitblick über die Phasen 4.0 bis 4.13 (2026-09-07), Befund 1.
//
// # Der Befund
//
// **Drei Stellen stellten das gezeigte Profil um, und nur eine war gesichert:**
//
//   - `ProfilWechseln` nahm die Umschaltsperre und richtete die
//     Leserverbindung neu,
//   - `NutzerWechseln` tat dasselbe **ohne** Sperre — dieselbe Lücke, ein
//     anderer Weg hinein,
//   - `ProfilLoeschen` zog nur die Kennung nach und ließ die Verbindung
//     stehen: nach dem Löschen des aktiven Profils gehörte sie noch zur
//     bisherigen Rolle.
//
// Die erste Korrektur schloss den **Fall** und nicht die **Klasse** (§10). Seit
// dem Nachlauf gehen alle drei durch `App.profilUmstellen`, und dieser Wächter
// hält es.
//
// # Was er hält und was nicht
//
//   - **Er hält**: `a.profil.Store(` und `a.leserverbindungRichten(` stehen
//     nur dort, wo sie stehen dürfen.
//   - **Er hält nicht**, ob die Sperre lange genug gehalten wird. Sie schützt
//     die Umstellung selbst, nicht den Ruf, der schon durch `imProfil` ist —
//     der gehört zu dem Stand, aus dem er stammt (§7, Regel 2), und das steht
//     bei `profilUmstellen` ausgeschrieben.
package werkzeuge_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// umstellendeZeilen findet, wer Kennung oder Leserverbindung setzt.
var umstellendeZeilen = regexp.MustCompile(
	`a\.profil\.Store\(|a\.leserverbindungRichten\(`)

// erlaubteUmstellorte sind die Dateien, in denen das stehen darf — mit Grund.
//
// **Ein leerer Grund ist keiner**, wie bei jeder Ausnahmeliste hier.
var erlaubteUmstellorte = map[string]string{
	"profilnaht.go": "hier steht der gemeinsame Weg selbst — profilLaden beim " +
		"Start, profilUmstellen für die drei Wechselwege und " +
		"verbindungNachziehen, wenn sich die Rechte bewegt haben (ID-03)",
	"start.go": "der Start richtet die Verbindung einmal ein, bevor eine " +
		"Bindung laufen kann; eine Sperre schützte dort niemanden",
}

// **`nurlesen.go` steht bewusst nicht in der Liste.** Es enthält die
// **Definition** von `leserverbindungRichten`, keinen Ruf — und das Muster
// verlangt das `a.` davor. Der erste Anlauf nahm es trotzdem auf, und die
// Gegenprobe „führt die Liste ein Gespenst?" hat es sofort gemeldet: eine
// Ausnahme für etwas, das gar nicht vorkommt, ist eine Zusage über nichts.

func TestJedeProfilumstellungGehtDurchEinenWeg(t *testing.T) {
	t.Parallel()

	eintraege, err := os.ReadDir(filepath.Join("..", "app"))
	if err != nil {
		t.Fatalf("internal/app lesen: %v", err)
	}

	var (
		daneben  []string
		gesehen  int
		gefunden bool
	)

	for _, eintrag := range eintraege {
		name := eintrag.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		gesehen++

		inhalt := lies(t, filepath.Join("internal", "app", name))
		if !umstellendeZeilen.MatchString(inhalt) {
			continue
		}

		gefunden = true

		grund, erlaubt := erlaubteUmstellorte[name]

		switch {
		case erlaubt && strings.TrimSpace(grund) == "":
			t.Errorf("%s ist ausgenommen, ohne dass ein Grund dasteht", name)
		case erlaubt:
		default:
			daneben = append(daneben, name)
		}
	}

	// **Die Gegenprobe zur Messung**: findet sie gar nichts, ist sie grün und
	// sagt nichts (§10).
	if gesehen < 50 || !gefunden {
		t.Fatalf("%d Dateien gesehen, Fundstellen: %v — sucht dieser Wächter "+
			"noch richtig?", gesehen, gefunden)
	}

	if len(daneben) > 0 {
		sort.Strings(daneben)
		t.Errorf("diese Dateien stellen das gezeigte Profil selbst um: %s — "+
			"der Weg dafür ist App.profilUmstellen, und er nimmt die "+
			"Umschaltsperre. Ohne sie nimmt eine Bindung dazwischen das neue "+
			"Profil mit der alten Verbindung",
			strings.Join(daneben, ", "))
	}

	// **Und die Ausnahmeliste führt kein Gespenst.**
	for datei := range erlaubteUmstellorte {
		if !umstellendeZeilen.MatchString(
			lies(t, filepath.Join("internal", "app", datei))) {
			t.Errorf("%s ist ausgenommen, stellt aber nichts mehr um", datei)
		}
	}
}
