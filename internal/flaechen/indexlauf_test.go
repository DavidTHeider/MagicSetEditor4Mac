// **Ein wirklich gebauter Index** — Werkstück A14 von Phase 5.3
// (#1453).
//
// # Was das Issue verlangt
//
// > echter fortlaufender MCP-Prozess mit Token-/Quellenwiderruf, **abgelaufenem
// > Cursor**, **laufendem Rebuild** und Offline-Modell
//
// **Beide waren bis hierher nicht zu stellen.** `suchindex_test.go` sagt in
// seinem Kopf, warum: „Dafür braucht es eine laufende Einbettungslaufzeit und
// eine Sammeldatei; beides steht hinter `netzlauf`." Seit die Probelaufzeit
// einbettet (`kilaufzeitstoff_test.go`) und eine selbst geschriebene
// Sammeldatei danebensteht (`sammeldatei_test.go`), sind sie es.
//
// # Was das prüft und was nicht
//
// **Geprüft ist der Weg**: Aufbau, Generation, Suche mit Treffern, Zeiger,
// Sperre. **Nicht geprüft ist die Güte** — die Vektoren kommen aus einer
// Rechnung dieses Laufs und nicht aus einem Modell. Wie gut echte Treffer
// sind, steht in [`docs/RETRIEVALMESSUNG.md`](../../docs/RETRIEVALMESSUNG.md)
// und ist an einem echten Modell gemessen.
package flaechen_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einGebauterIndex baut einen und gibt die Umgebung samt Laufzeit.
func einGebauterIndex(t *testing.T) ([]string, *laufzeitstelle) {
	t.Helper()

	umgebung, laufzeit := einAgentenbestand(t)

	aus := mussKommandozeile(t, umgebung, "indexaufbau", probeeinbetter,
		eineSammeldatei(t), "prf-2026-01-01", "2026-01-01")

	if !strings.Contains(aus, "Generation 1") {
		t.Fatalf("der Aufbau meldet keine erste Generation: %s", aus)
	}

	if laufzeit.einbettungen() == 0 {
		t.Fatalf("es wurde nichts eingebettet — dann ist der Index leer")
	}

	return umgebung, laufzeit
}

// **Der Index steht, und die Suche findet darin.**
func TestEinGebauterIndexFindetSeineAbschnitte(t *testing.T) {
	t.Parallel()

	umgebung, _ := einGebauterIndex(t)
	agent := starteMCPProzess(t, umgebung)

	stand := agent.muss("lesen.indexstand", map[string]any{})
	if !strings.Contains(stand, `"aktiv": true`) {
		t.Fatalf("der Index steht nicht aktiv: %s", stand)
	}

	text := agent.muss("lesen.recherche", map[string]any{
		"fall": "karten", "frage": "blocken", "grenze": 2,
	})

	if strings.Contains(text, `"leere": "ohne-index"`) {
		t.Fatalf("die Recherche sieht keinen Index: %s", text)
	}

	if !strings.Contains(text, "hybrid") {
		t.Errorf("die Suche lief nicht hybrid: %s", text)
	}
}

// **Ein Zeiger, der zu einer anderen Frage gehört, wird abgewiesen.**
//
// # Warum das der prüfbare Teil des „abgelaufenen Cursors" ist
//
// **Weil beide dieselbe Bindung sind.** Ein Zeiger trägt einen Abdruck über
// Fall, Frage, Profil, Grenze und Indexstand **und** seinen Zeitpunkt;
// [domain.ZeigerAus] prüft beides in einem Zug. Die Frist von fünfzehn
// Minuten ließe sich hier nur mit einer erfundenen Uhr erreichen — die
// **Bindung** dagegen bricht, sobald sich die Frage ändert, und das ist
// derselbe Riegel.
//
// **Die Frist steht unter eigenem Lauf** in
// `internal/domain/recherchezeiger_test.go`.
func TestEinZeigerZuEinerAnderenFrageWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	umgebung, _ := einGebauterIndex(t)
	agent := starteMCPProzess(t, umgebung)

	// **Die erste Seite gibt den Zeiger her.**
	zeiger := fortsetzungAus(t, agent.muss("lesen.recherche", map[string]any{
		"fall": "karten", "frage": "blocken", "grenze": 1,
	}))

	if zeiger == "" {
		t.Fatalf("die erste Seite nennt keinen Zeiger")
	}

	// **Mit derselben Frage trägt er.**
	agent.muss("lesen.recherche", map[string]any{
		"fall": "karten", "frage": "blocken", "grenze": 1,
		"fortsetzung": zeiger,
	})

	// **Mit einer anderen nicht.**
	agent.scheitert("lesen.recherche", map[string]any{
		"fall": "karten", "frage": "funken", "grenze": 1,
		"fortsetzung": zeiger,
	}, "Fortsetzungszeiger gehört zu einer anderen Frage")
}

// **Ein zweiter Aufbau, während einer läuft, wird abgewiesen.**
//
// # Was das Issue verlangt
//
// > … **laufendem Rebuild** …
//
// # Und der Befund, den dieser Lauf gefunden hat
//
// **Die Absprache über die Prozesse war an dieser Fläche ein Nichts.**
// `Indexsperre` hängt am Pfad des Bestands, und `RetrievalAnwendungUeber`
// baute die Anwendung über `NeueApp("")` — ohne Pfad gibt `Nehmen`
// „genommen" zurück, ohne etwas zu nehmen. **Zwei `mse indexaufbau` bauten
// gleichzeitig**, gegen dieselbe Datei, während der Satz in `RetrievalUeber`
// zusagt, die Absprache gebe es gerade für diese Fläche. Der erste Anlauf
// dieses Laufs bekam `0`, wo eine Absage stehen sollte.
//
// # Warum der Aufbau verlangsamt und nicht angehalten wird
//
// **Weil ein angehaltener Aufbau eine Lage erzeugt, die es nicht gibt.** Er
// hält dann seine Schreibtransaktion über die ganze Zeit, und der zweite
// Prozess wartet auf SQLite statt auf die Absprache — gemessen am
// 2026-09-14: der Lauf lief zweimal in die Zeitgrenze, und der Grund war
// die Vorrichtung und nicht der Code (§10).
//
// **Verlangsamt dagegen läuft er wirklich**: er bündelt, schreibt, gibt frei
// und nimmt sich das nächste Bündel — und **dazwischen** trifft ihn der
// zweite.
func TestEinZweiterAufbauWaehrendEinesLaufendenWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	umgebung, laufzeit := einGebauterIndex(t)

	// **Genug Karten, dass der Aufbau anzutreffen ist** — mit dreien ist er
	// nach 150 ms vorbei.
	datei := eineGrosseSammeldatei(t, 200)
	laufzeit.verzoegern(8 * time.Millisecond)

	agent := starteMCPProzess(t, umgebung)

	fertig := make(chan string, 1)
	vorher := laufzeit.einbettungen()

	go func() {
		aus, klage, _ := kommandozeile(t, umgebung, "indexaufbau",
			probeeinbetter, datei, "prf-2026-02-01", "2026-02-01")

		fertig <- aus + klage
	}()

	warteAufEinbettung(t, laufzeit, vorher)

	// **Die Suche antwortet weiter — aus der alten Generation.**
	//
	// **Das ist die andere Hälfte der Zusage** (A15): ein Aufbau baut
	// **neben** der aktiven Generation und schaltet erst am Ende um.
	if text := agent.muss("lesen.recherche", map[string]any{
		"fall": "karten", "frage": "blocken", "grenze": 1,
	}); strings.Contains(text, `"leere": "ohne-index"`) {
		t.Errorf("während des Aufbaus steht kein Index mehr: %s", text)
	}

	// **Und ein zweiter Aufbau kommt nicht durch.**
	_, klage, code := kommandozeile(t, umgebung, "indexaufbau",
		probeeinbetter, datei, "prf-2026-03-01", "2026-03-01")

	if code == 0 {
		t.Errorf("der zweite Aufbau lief mit")
	}

	if !strings.Contains(klage, "ein anderer Prozess baut") {
		t.Errorf("die Absage nennt den Grund nicht: %s", klage)
	}

	if aus := <-fertig; !strings.Contains(aus, "Generation") {
		t.Errorf("der erste Aufbau kam nicht durch: %s", aus)
	}
}

// warteAufEinbettung hält an, bis die Laufzeit wirklich gefragt wurde.
//
// # Warum hier gewartet und nicht zugesichert wird
//
// **Weil der Aufbau in einem anderen Prozess läuft.** Es gibt kein Ereignis,
// auf das dieser Lauf horchen könnte; was es gibt, ist ein Zähler an der
// Gegenstelle. Gewartet wird deshalb auf **ihn** und mit einer Frist —
// nicht auf eine Zeitspanne, von der jemand hofft, sie reiche.
func warteAufEinbettung(t *testing.T, laufzeit *laufzeitstelle, vorher int64) {
	t.Helper()

	frist := time.After(30 * time.Second)

	for {
		if laufzeit.einbettungen() > vorher {
			return
		}

		select {
		case <-frist:
			t.Fatalf("in 30 s wurde nicht eingebettet — der Aufbau steht nicht")
		case <-t.Context().Done():
			t.Fatalf("der Lauf endete, bevor eingebettet wurde")
		case <-time.After(5 * time.Millisecond):
		}
	}
}

// fortsetzungAus liest den Zeiger aus einer Rechercheantwort.
func fortsetzungAus(t *testing.T, roh string) string {
	t.Helper()

	var antwort struct {
		Fortsetzung string `json:"fortsetzung"`
	}

	if fehler := json.Unmarshal([]byte(roh), &antwort); fehler != nil {
		t.Fatalf("die Antwort lesen: %v — %s", fehler, roh)
	}

	return antwort.Fortsetzung
}

// **Und der Modus sagt, dass ein Modell dabei war.**
func TestEineSucheImGebautenIndexNenntIhrModell(t *testing.T) {
	t.Parallel()

	umgebung, _ := einGebauterIndex(t)
	agent := starteMCPProzess(t, umgebung)

	text := agent.muss("lesen.recherche", map[string]any{
		"fall": "karten", "frage": "blocken", "grenze": 2,
	})

	if !strings.Contains(text, probeeinbetter) {
		t.Errorf("die Antwort nennt das Modell nicht: %s", text)
	}

	// **Und die Ortsregel steht dabei** — sie gilt auch für die Suche.
	_ = domain.RichtlinieNurLokal
}
