// Der Wächter über die Kreuzübersetzung.
//
// `docs/PORTIERUNG.md` sagt zu, die Pipeline übersetze für Windows und Linux
// mit und linte über alle drei Ziele. Diese Zusage steht in einer Datei und
// wird in einer anderen eingelöst — das ist genau die Bauart, die dieses
// Repository zweimal hat auseinanderlaufen sehen (Phase 1, Punkt 4).
//
// Zu verlieren ist dabei nichts Sichtbares: nimmt jemand `windows/arm64` aus
// der Schleife, bleibt die Pipeline grün, die Datei behält ihren Satz, und der
// erste Code hinter `//go:build windows` bricht unbemerkt.
//
// # Was er nicht prüft
//
// **Ob die Übersetzung gelingt.** Er liest ein Skript und keine Ausgabe. Dass
// `go build GOOS=windows` durchläuft, sagt die Stufe `types` selbst — hier
// steht nur, dass sie es überhaupt versucht.
//
// **Ob ein Bau eine Erprobung wäre.** Er ist keine, und das steht in
// `docs/PORTIERUNG.md` gleich im ersten Abschnitt.
package werkzeuge

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// kreuzziele sind die Ziele, für die `go build` mitlaufen muss.
//
// **`windows/arm64` gehört dazu.** Es ist das Ziel, das man beim Kürzen zuerst
// wegnimmt, weil es „schon durch amd64 abgedeckt" scheint — und es ist das
// einzige, auf dem eine Annahme über die Wortbreite auffiele.
var kreuzziele = []string{"windows/amd64", "windows/arm64", "linux/amd64"}

// lintziele sind die `GOOS`, über die golangci-lint läuft.
//
// Der Grund steht in `scripts/ci.sh`: der Linter sieht nur, was für das
// aktuelle `GOOS` übersetzt wird. Ohne die Schleife entzöge sich jede Datei
// hinter einem Build-Tag der Prüfung.
var lintziele = []string{"darwin", "windows", "linux"}

// aufrufe sind die Kommandos, die in der jeweiligen Schleife stehen müssen.
//
// **Sonst prüfte der Wächter nur eine Aufzählung.** Wer den Rumpf durch ein `:`
// ersetzt, hat die Ziele stehen und übersetzt nichts mehr — und die Zusage aus
// `docs/PORTIERUNG.md` wäre wieder eine Behauptung. Gefunden vom Zweitblick
// über Phase 8: der Wächter hielt die Kopfzeile und ließ den Rumpf frei.
var aufrufe = map[string]string{
	"bau":  "go build",
	"lint": "golangci-lint run",
}

// befundeKreuzbau prüft ein `ci.sh` auf beide Schleifen.
func befundeKreuzbau(skript string) []string {
	var befunde []string

	bau, lint := schleifen(skript)

	// **Die Ziele stehen im Kopf der Schleife**, und nur dort zählen sie.
	// Seit ENG-02 nennt der Rumpf `linux` selbst — dort übersetzt er ohne die
	// Hülle (`scripts/linuxpakete.sh`) —, und ein Wächter, der den ganzen
	// Block läse, bliebe grün, wenn `linux` aus dem Kopf verschwände.
	for _, ziel := range kreuzziele {
		if !slices.Contains(kopfziele(bau), ziel) {
			befunde = append(befunde, "die Stufe types übersetzt nicht für "+ziel)
		}
	}

	for _, ziel := range lintziele {
		if !slices.Contains(kopfziele(lint), ziel) {
			befunde = append(befunde, "die Stufe lint läuft nicht über "+ziel)
		}
	}

	// **Und der Rumpf muss das Ziel auch benutzen.** `GOOS=` dazu, denn ein
	// `go build` ohne die Umgebungsvariable übersetzte dreimal für dieselbe
	// Plattform.
	for name, kommando := range aufrufe {
		rumpf := bau
		if name == "lint" {
			rumpf = lint
		}

		if !strings.Contains(rumpf, kommando) {
			befunde = append(befunde, "die Schleife "+name+" ruft kein "+kommando+" auf")
		}

		if !strings.Contains(rumpf, "GOOS=") {
			befunde = append(befunde, "die Schleife "+name+" setzt kein GOOS")
		}
	}

	return befunde
}

// schleifen schneidet die beiden `for ziel in …`-Blöcke heraus — Kopf **und**
// Rumpf bis zum `done`.
//
// **Nicht die ganze Datei**: `linux` steht auch in Kommentaren, und eine
// Prüfung auf das bloße Vorkommen wäre grün, sobald jemand das Wort irgendwo
// erwähnt. Gesucht wird der Block, der es *ausführt*.
func schleifen(skript string) (bau, lint string) {
	zeilen := strings.Split(skript, "\n")

	for i, zeile := range zeilen {
		geputzt := strings.TrimSpace(zeile)

		if !strings.HasPrefix(geputzt, "for ziel in ") {
			continue
		}

		block := strings.Join(zeilen[i:min(i+bloecktiefe, len(zeilen))], "\n")
		if kopf, _, gefunden := strings.Cut(block, "\n    done"); gefunden {
			block = kopf
		}

		// Die Bauschleife nennt Ziele mit Schrägstrich, die Lintschleife nicht.
		if strings.Contains(geputzt, "/") {
			bau += block + "\n"
		} else {
			lint += block + "\n"
		}
	}

	return bau, lint
}

// kopfziele liest die Ziele aus den Köpfen `for ziel in …;` eines Blocks.
func kopfziele(block string) []string {
	var ziele []string

	for zeile := range strings.SplitSeq(block, "\n") {
		_, hinten, gefunden := strings.Cut(strings.TrimSpace(zeile), "for ziel in ")
		if !gefunden {
			continue
		}

		liste, _, _ := strings.Cut(hinten, ";")
		ziele = append(ziele, strings.Fields(liste)...)
	}

	return ziele
}

// bloecktiefe ist, wie viele Zeilen hinter dem `for` höchstens dazugehören.
//
// Eine Schranke, weil ein `done` fehlen kann — dann liest der Wächter sonst bis
// zum Dateiende und findet `go build` irgendwo anders.
const bloecktiefe = 12

func TestDiePipelineUebersetztFuerWindowsUndLinux(t *testing.T) {
	for _, befund := range befundeKreuzbau(pipelineskript(t)) {
		t.Error(befund)
	}
}

// pipelineskript gibt `ci.sh` **und jede** ihrer ausgelagerten Stufen.
//
// **Die Zusage hängt an der Pipeline, nicht an einer Datei.** `ci.sh` steht an
// der Zeilengrenze aus §2 und gibt Stufen ab, sobald sie wächst — `lint` am
// 2026-08-25, davor `abdeckung` und `pinnung`. Wer hier nur `ci.sh` läse,
// bekäme beim nächsten Schnitt einen roten Wächter, ohne dass sich an der
// Pipeline etwas geändert hätte.
func pipelineskript(t *testing.T) string {
	t.Helper()

	orte, err := filepath.Glob(filepath.Join("..", "..", "scripts", "ci*.sh"))
	if err != nil || len(orte) == 0 {
		t.Fatalf("die Pipeline-Skripte sind nicht auffindbar: %v", err)
	}

	// **Sortiert, damit die Ausgabe eines roten Laufs dieselbe bleibt.**
	// `Glob` gibt sie zwar sortiert, aber das steht nicht im Vertrag.
	slices.Sort(orte)

	zusammen := make([]string, 0, len(orte))
	for _, ort := range orte {
		zusammen = append(zusammen, datei(t, "scripts", filepath.Base(ort)))
	}

	return strings.Join(zusammen, "\n")
}

// Die Gegenprobe: jedes weggenommene Ziel macht den Wächter rot.
func TestDerKreuzbauWaechterSchlaegtAn(t *testing.T) {
	gut := pipelineskript(t)

	if befunde := befundeKreuzbau(gut); len(befunde) != 0 {
		t.Fatalf("die echte Pipeline gilt schon als falsch: %v", befunde)
	}

	// **Ein leergeräumter Rumpf.** Der Zweitblick über Phase 8 hat genau das
	// vorgeschlagen: Ziele stehen lassen, Übersetzer entfernen. Vorher blieb
	// der Wächter grün.
	t.Run("die Schleife übersetzt nichts mehr", func(t *testing.T) {
		ohne := strings.ReplaceAll(strings.ReplaceAll(gut,
			"go build", ":"), "golangci-lint run", ":")

		if len(befundeKreuzbau(ohne)) == 0 {
			t.Error("der Wächter lässt eine Schleife ohne Übersetzer durch")
		}
	})

	// **Und ein Rumpf ohne GOOS.** Dreimal dieselbe Plattform sieht in der
	// Ausgabe aus wie drei Ziele.
	t.Run("die Schleife setzt kein GOOS", func(t *testing.T) {
		ohne := strings.ReplaceAll(gut, `GOOS="`, `NICHTS="`)

		if len(befundeKreuzbau(ohne)) == 0 {
			t.Error("der Wächter lässt eine Schleife ohne GOOS durch")
		}
	})

	for _, ziel := range append(append([]string{}, kreuzziele...), lintziele...) {
		t.Run("ohne "+ziel, func(t *testing.T) {
			ohne := entferneZiel(gut, ziel)

			if ohne == gut {
				t.Fatalf("%s steht gar nicht in einer Schleife — die Gegenprobe greift ins Leere", ziel)
			}

			if len(befundeKreuzbau(ohne)) == 0 {
				t.Error("der Wächter lässt es durch")
			}
		})
	}
}

// entferneZiel streicht ein Ziel aus den Schleifenzeilen, sonst nirgends.
//
// Nur dort, damit die Gegenprobe dasselbe trifft wie der Wächter: eine
// Ersetzung über die ganze Datei änderte auch Kommentare und bewiese nichts
// über die Schleife.
//
// **Wortweise und nicht als Zeichenkette.** Der erste Anlauf ersetzte
// `ziel+" "` und ließ das **letzte** Ziel jeder Schleife stehen — dahinter
// steht `;` und kein Leerzeichen. Die Gegenprobe war damit für `linux/amd64`
// und `linux` wirkungslos, und der Wächter sah trotzdem grün aus; gefunden hat
// es der Satz, der genau davor sitzt („die Gegenprobe greift ins Leere").
func entferneZiel(skript, ziel string) string {
	zeilen := strings.Split(skript, "\n")

	for i, zeile := range zeilen {
		if !strings.HasPrefix(strings.TrimSpace(zeile), "for ziel in ") {
			continue
		}

		vorn, hinten, gefunden := strings.Cut(zeile, "for ziel in ")
		if !gefunden {
			continue
		}

		liste, rest, gefunden := strings.Cut(hinten, ";")
		if !gefunden {
			continue
		}

		bleibt := make([]string, 0, 3)

		for _, wort := range strings.Fields(liste) {
			if wort != ziel {
				bleibt = append(bleibt, wort)
			}
		}

		zeilen[i] = vorn + "for ziel in " + strings.Join(bleibt, " ") + ";" + rest
	}

	return strings.Join(zeilen, "\n")
}
