package app_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Ein halb geglückter AllSets-Import
// (#155).
//
// # Warum das hier steht und über die echte Naht geht
//
// Das Issue verlangt es ausdrücklich: „ein Test führt einen halb geglückten
// Import über die **echte** Naht, nicht nur gegen ein Double". Der Grund ist
// derselbe wie bei jedem Integrationstest in diesem Paket: jede Schicht prüft
// gegen ihre eigene Vorrichtung, und dazwischen liegt die Fehlerklasse
// (AGENTS.md §3).
//
// Hier ist sie greifbar: `service` **gibt** das Teilergebnis neben dem Fehler
// zurück, seit #154. `app` hat es weggeworfen — beide Seiten grün, und im
// Bestand lagen Sets, von denen niemand wusste.

// **Ein Abbruch mitten in einer AllSets-Datei nennt, was schon dasteht.**
func TestEinHalbGeglueckterImportNenntSeineSets(t *testing.T) {
	db := frischeAppdatenbank(t)

	anwendung := &app.App{}
	anwendung.MitImporteur(service.NeuerImporteur(storage.NeuerSetSpeicher(db)))
	anwendung.MitBestand(storage.NeuerSetLeser(db))

	antwort := anwendung.MTGJSONOeffnen(probekontext(t), halbeAllSets(t), "magic")

	// **Der Fehler bleibt ein Fehler.** Ein Import, der nicht zu Ende kam, ist
	// kein Erfolg mit Anmerkung.
	if antwort.Fehler == nil {
		t.Fatal("kein Fehler — die kaputte Datei ging durch")
	}

	if !antwort.Unvollstaendig {
		t.Error("die Antwort sagt nicht, dass die Datei nicht zu Ende gelesen wurde")
	}

	// **Und das erste Set steht da.** Jedes liegt in einer eigenen Transaktion.
	if len(antwort.Angelegt) != 1 {
		t.Fatalf("angelegt: %v", antwort.Angelegt)
	}

	// **Die Gegenprobe über den Bestand**, nicht über die Antwort: was sie
	// nennt, muss wirklich dastehen.
	koepfe, err := storage.NeuerSetLeser(db).Setkoepfe(probekontext(t))
	if err != nil {
		t.Fatalf("Köpfe: %v", err)
	}

	if len(koepfe) != 1 {
		t.Fatalf("%d Sets im Bestand", len(koepfe))
	}

	if antwort.Angelegt[0] != "1" || koepfe[0].ID != 1 {
		t.Errorf("die Antwort nennt %v, der Bestand führt %d", antwort.Angelegt, koepfe[0].ID)
	}
}

// **Ein geglückter Import nennt seine Sets ebenso.**
//
// Sonst hinge die Auskunft am Fehlerfall — und wer sie im Erfolgsfall liest,
// bekäme eine leere Liste, die etwas anderes bedeutet.
func TestEinGeglueckterImportNenntSeineSetsAuch(t *testing.T) {
	db := frischeAppdatenbank(t)

	anwendung := &app.App{}
	anwendung.MitImporteur(service.NeuerImporteur(storage.NeuerSetSpeicher(db)))
	anwendung.MitBestand(storage.NeuerSetLeser(db))

	antwort := anwendung.MTGJSONOeffnen(probekontext(t),
		filepath.Join("..", "..", "testdaten", "mtgjson", "alle.json"), "magic")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Unvollstaendig {
		t.Error("ein geglückter Import gilt als unvollständig")
	}

	if len(antwort.Angelegt) != 2 {
		t.Errorf("angelegt: %v", antwort.Angelegt)
	}
}

// halbeAllSets schreibt eine AllSets-Datei, deren **zweites** Set kaputt ist.
//
// **Aus der echten Probedatei gebaut und nicht von Hand getippt**: was der
// Leser annimmt, entscheidet er selbst — eine erfundene Datei prüfte, was wir
// für richtig halten (§3).
func halbeAllSets(t *testing.T) string {
	t.Helper()

	roh, err := os.ReadFile(filepath.Join("..", "..", "testdaten", "mtgjson", "alle.json"))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	var datei struct {
		Meta json.RawMessage            `json:"meta"`
		Data map[string]json.RawMessage `json:"data"`
	}

	if fehler := json.Unmarshal(roh, &datei); fehler != nil {
		t.Fatalf("zerlegen: %v", fehler)
	}

	if len(datei.Data) != 2 {
		t.Fatalf("die Probedatei führt %d Sets, erwartet waren 2", len(datei.Data))
	}

	// **Das zweite Set bekommt eine Karte mit einem namenlosen Schlüssel.**
	//
	// `domain.PruefeFeldname` weist ihn ab — „ein Feld braucht einen Namen" —,
	// und `Set.ValidateBasics` macht daraus den Abbruch. Eine Namenskollision
	// (`text` und `rule text`) taugt dafür **nicht**: die löst `internal/mtgjson`
	// seit Phase 9 selbst auf, als Befund. Das ist der Unterschied zwischen
	// einem Vorbehalt und einem Abbruch (§8), und er gehört hierher, weil der
	// erste Anlauf ihn verwechselt hat.
	for kuerzel := range datei.Data {
		if kuerzel == "PRB" {
			continue
		}

		datei.Data[kuerzel] = json.RawMessage(`{"code":"` + kuerzel +
			`","name":"Kaputt","cards":[{"name":"Namenlos","":"ohne Schlüssel"}]}`)
	}

	gebaut, err := json.Marshal(datei)
	if err != nil {
		t.Fatalf("bauen: %v", err)
	}

	ort := filepath.Join(t.TempDir(), "halb.json")
	if fehler := os.WriteFile(ort, gebaut, 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	return ort
}
