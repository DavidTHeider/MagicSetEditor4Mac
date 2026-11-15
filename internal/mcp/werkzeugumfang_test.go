package mcp_test

// Der erlaubte Umfang der schreibenden Werkzeuge —
// [0040](../../docs/entscheidungen/0040-der-lebenszyklus-einer-karte-gehoert-dazu.md).
//
// **Eine eigene Datei, weil er eine eigene Frage beantwortet** (§2):
// `werkzeuge_test.go` prüft, wie ein Werkzeug sich vorstellt; hier steht,
// welche es überhaupt geben darf.

import (
	"strings"
	"testing"
)

// **Kein Werkzeug ändert mehr als die Karten, die es nennt** —
// [0040](../../docs/entscheidungen/0040-der-lebenszyklus-einer-karte-gehoert-dazu.md).
//
// # Was hier bis zum 2026-08-22 stand
//
// „Es gibt genau **ein** schreibendes Werkzeug." Das war eine Zusage, solange
// ihr Grund in [0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)
// stand: eine gelöschte Karte ließ sich damals nicht zurückholen, und deshalb
// gab es kein Werkzeug dafür.
//
// **Seit #621
// kommt sie mit derselben Kennung zurück** — fünf Tage nach 0010. Der Boden hat
// sich verschoben, und eine Zahl, deren Grund fällt, prüft nur noch sich selbst.
//
// Geblieben ist die andere Hälfte von 0010, und die steht hier jetzt als Zusage
// statt als Zahl: kein Werkzeug überschreibt ein Set, keines ändert mehrere
// Karten auf einmal.
//
// # Was dieser Wächter hält — und was der Name zu viel verspricht
//
// Er liest **Namen**, keine Wirkung. Was er wirklich hält, ist eine
// Zusage über die Buchführung: **kein schreibendes Werkzeug steht ohne eine
// aufgeschriebene Umfangsentscheidung da.** Wer eines dazuschreibt, muss sich
// festlegen, in welche der elf Gruppen es gehört — und das ist der Moment, in
// dem jemand hinsieht.
//
// **Was er nicht kann**, und was bis zum 2026-08-27 nirgends danebenstand: er
// sieht nicht nach, ob ein Werkzeug seine Gruppe **einlöst**. Die Gruppen sind
// Text, den jemand getippt hat; ein Werkzeug, das sich in die falsche einträgt,
// macht ihn nicht rot. Angestrichen vom Zweitblick über Phase 3.4, und der
// Befund lautete genau so: eine Kategorieliste, die sich als Wirkungswächter
// ausgibt.
//
// **Wo die Wirkung wirklich gemessen wird**, je Zusage:
//
//   - „eine Zeile der Sammlung" — `sammlungsumfang_test.go`, drei Zeilen
//     hinein, ein Ruf, die anderen beiden Feld für Feld unverändert
//   - „nichts, was schon da ist" — `internal/service/schonoffen_test.go` für
//     `einfahren`, `cmd/mse-mcp` für `sichern` und `zeichnen`
//   - „sagt, wie man es zurücknimmt" —
//     [TestDerLebenszyklusEinerKarteStehtUeberMCP] für die vier neuen und
//     `feldstelle_test.go` für `schreiben.feld`
//
// Die übrigen sind **nicht** gemessen, und das ist der Preis dieser Liste.
//
// Kein Löschen, kein Überschreiben eines Sets, keine Massenänderung — obwohl
// der Bestand das könnte. Ein Wächter darüber, weil ein solches Werkzeug sich
// leicht dazuschreibt und dann in derselben Reihe stünde
// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)).
//
// # Und seit dem 2026-08-23 stehen drei daneben, die mehr anfassen
//
// [0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)
// stellt den Maßstab richtig: **nicht die Menge, sondern die Umkehrbarkeit.**
// Ein Einfahren legt Hunderte Karten an — aber ein **neues** Set, und es ändert
// nichts Bestehendes. Sichern und Zeichnen fassen den Bestand gar nicht an; sie
// schreiben eine Datei, und zwar nur dorthin, wo nichts liegt.
//
// Die Liste unten führt sie deshalb **getrennt**: was eine Karte je Aufruf
// ändert, und was nur schreibt, ohne zu überschreiben. Ein Werkzeug in keiner
// der beiden Gruppen ist ein Befund.
func TestJedesSchreibendeWerkzeugStehtInEinerZusage(t *testing.T) {
	zugesagt := zusagen()

	gesehen := 0

	for _, werkzeug := range werkzeugliste(t) {
		name := zeichenkette(t, werkzeug["name"])
		if !strings.HasPrefix(name, "schreiben.") {
			continue
		}

		gesehen++

		if !zugesagt[name] {
			t.Errorf("%q schreibt und steht in keiner Zusage — 0040 erlaubt "+
				"eine Karte je Aufruf, 0044 das Schreiben ohne Überschreiben", name)
		}
	}

	// **Und die Zahl in beide Richtungen.** Ohne sie bliebe eine Zusage für
	// ein Werkzeug stehen, das es nicht mehr gibt — und niemand sähe es.
	if gesehen != len(zugesagt) {
		t.Errorf("%d schreibende Werkzeuge, erwartet %d", gesehen, len(zugesagt))
	}
}
