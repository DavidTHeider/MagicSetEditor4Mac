package app_test

import (
	"context"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Zwei Ausfuhren dürfen sich nicht kreuzen** — Befund des Zweitblicks durch
// Codex vom 2026-08-21.
//
// # Der Befund
//
// Es gibt **einen** `Vorlagenausfuehrer` je Anwendung
// (`internal/service/vorlagenliste.go`, Zeile 235), und `schreibeVorlage`
// verändert ihn je Aufruf: Bildziel, Beigabenort, Setpaket, Verzeichnis
// (`vorlagenausfuhr.go`, Zeilen 328–360).
//
// Die Oberfläche sperrt über `useVorgang` mit dem Schlüssel
// `ausfuehren-${satz.id}` — **je Set**. Zwei Ausfuhren **verschiedener** Sets
// laufen damit ausdrücklich nebeneinander, und Wails ruft sie nebenläufig. Der
// zweite Lauf überschrieb dem ersten sein Ziel.
//
// # Was der Test zusagt, und was nicht
//
// **Er prüft die Zusage, nicht den Mechanismus.** Ein erster Anlauf hielt
// beide Läufe in `Fuehre` fest, bis der jeweils andere da war — und lief nach
// dem Einbau einer Sperre in einen Stillstand, weil genau das nicht mehr
// geschehen kann. Ein Test, der den Mechanismus festschreibt, verbietet seine
// Behebung.
//
// Geprüft wird deshalb, was von außen zu sehen ist: **jeder Lauf führt mit dem
// Verzeichnis und dem Bildziel aus, die zu seiner eigenen Zieldatei gehören.**
// Beide leiten sich aus demselben Namen ab — kreuzen sie sich, passen sie
// nicht mehr zueinander.
//
// **Was er nicht prüft**: den Rasterer. Hier stand „dazu `-race`, das den
// unabgesicherten Zugriff auch dann findet, wenn die Verschränkung diesmal
// nicht eintritt" — und das war zu viel versprochen: die Attrappe unten
// berührt den geteilten Rasterer gar nicht, und `-race` kann auf diesem Weg
// nichts finden. Angestrichen vom Zweitblick durch Codex am 2026-08-21; der
// Rasterer hat seit
// #748 eigene
// Wächter in `internal/render`.
//
// **Was er nicht prüft**: dass die Mittel *innerhalb* eines Laufs unverändert
// bleiben. Seit sie ein Wert sind (#736), kann das niemand mehr verletzen —
// eine Zusicherung darüber wäre eine über die Sprache, nicht über diesen Code.
//
// # Die Gegenprobe
//
// `schreibeVorlage` bekam den Namen über eine geteilte Variable zurück, und
// der Test wurde rot: **ein** Lauf kam an, beide mit `erste-files`. Es fällt
// dabei die Zählung, nicht der Abgleich — kreuzen sich beide Läufe, tragen sie
// denselben Schlüssel, und der zweite Eintrag überschreibt den ersten.
//
// **Ein `runtime.Gosched()` zwischen Schreiben und Lesen genügte dafür nicht**
// — der Lauf blieb grün, und das sagte nichts über den Test aus, sondern nur,
// dass der Eingriff nicht griff. Erst eine erzwungene Verschränkung zeigte es.
func TestZweiAusfuhrenKreuzenSichNicht(t *testing.T) {
	probe := &nebenlaufprobe{
		vorlagenprobe: vorlagenprobe{
			vorlagen: []msepaket.Exportvorlage{mitVerzeichnisvorlage()},
			ausgabe:  "fertig",
		},
		gesehen: map[string]string{},
	}

	anwendung := mitVorlagen(t, &probe.vorlagenprobe)
	anwendung.MitVorlagen(probe)

	ziel := t.TempDir()

	var warte sync.WaitGroup

	for _, satz := range []struct{ id, datei string }{
		{"1", "erste.txt"},
		{"2", "zweite.txt"},
	} {
		warte.Add(1)

		go func() {
			defer warte.Done()

			anwendung.VorlageAusfuehren(t.Context(), satz.id,
				"probe.mse-export-template", filepath.Join(ziel, satz.datei), nil, nil)
		}()
	}

	warte.Wait()

	// **Beide Läufe haben ihr eigenes Ziel behalten.** Der Wert unter einem
	// Verzeichnisnamen ist das Bildziel desselben Laufs; sein letzter
	// Pfadabschnitt muss derselbe Name sein.
	for verzeichnis, bildziel := range probe.gesehen {
		if filepath.Base(bildziel) != verzeichnis {
			t.Errorf("%q führte mit dem Bildziel %q aus — das gehört einem "+
				"anderen Lauf", verzeichnis, bildziel)
		}
	}

	if len(probe.gesehen) != 2 {
		t.Errorf("%d Läufe angekommen, erwartet 2: %v", len(probe.gesehen), probe.gesehen)
	}
}

// nebenlaufprobe hält fest, was ein Lauf beim Eintritt und beim Austritt sieht.
type nebenlaufprobe struct {
	vorlagenprobe

	sperre  sync.Mutex
	gesehen map[string]string
}

func (n *nebenlaufprobe) Fuehre(
	_ context.Context, mittel service.Ausfuhrmittel, _ msepaket.Exportvorlage,
	_ []*domain.CardRow, _ []domain.Value, _ map[string]string,
) (string, []domain.Befund, error) {
	// **Dem anderen Lauf die Gelegenheit geben.** Ohne den Wert reicht das, um
	// die Verschränkung herbeizuführen; mit ihm ändert es nichts.
	runtime.Gosched()

	n.sperre.Lock()
	n.gesehen[mittel.Verzeichnis] = mittel.Bildziel
	n.sperre.Unlock()

	return "fertig", nil, nil
}

// mitVerzeichnisvorlage ist eine Vorlage, die ein Beigabenverzeichnis anlegt.
func mitVerzeichnisvorlage() msepaket.Exportvorlage {
	vorlage := probevorlage()
	vorlage.Verzeichnis = true

	return vorlage
}
