package service_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Karten der Datei und die Karten des Bestands finden über die
// Reihenfolge zueinander, nicht übers Abzählen**
// (#621).
//
// Bis hierher lief der Schreiber die `card:`-Blöcke der Datei durch und paarte
// sie der Reihe nach mit den Karten des Bestands. Das trug, solange es beide
// gleich oft gab. Seit eine Karte gelöscht werden kann, gibt es sie nicht mehr
// gleich oft — und das Abzählen schrieb dann **die falschen Werte in die
// falschen Blöcke**: nach dem Löschen der mittleren von drei Karten stand die
// dritte zweimal in der Datei und die zweite gar nicht.
//
// `cards.reihenfolge` ist die Stelle, an der die Karte in der Datei steht. Sie
// wird beim Löschen **nicht** nachgezogen (`storage.KarteLoeschen`), und genau
// diese Lücke sagt dem Schreiber, welcher Block gehen muss.

// dreikartenquelle stellt eine Datei mit drei Kartenblöcken nach.
const dreikartenquelle = "mse version: 2.0.0\n" +
	"game: probe\n" +
	"card:\n" +
	"\tname: Feuerball\n" +
	"card:\n" +
	"\tname: Eisball\n" +
	"card:\n" +
	"\tname: Doppelt\n" +
	"unverstandener block:\n" +
	"\tirgendwas: bleibt liegen\n"

// mitKarten gibt einen Bestand über der dreikartenquelle.
func mitKarten(karten ...*domain.CardRow) quellenDouble {
	return quellenDouble{
		leserDouble: leserDouble{
			koepfe: []domain.SetHead{{ID: 7, Name: "Mein Probeset", Game: "probe"}},
			karten: karten,
		},
		quelle: &domain.Quelle{Datei: "set", Inhalt: []byte(dreikartenquelle)},
	}
}

// karteMit baut eine Bestandskarte an einer Stelle.
func karteMit(kennung int64, stelle int, name string) *domain.CardRow {
	return &domain.CardRow{
		ID: kennung, Number: stelle,
		Values: []domain.Value{{Field: "name", Text: name}},
	}
}

// geschrieben schreibt das Set und gibt die Hauptdatei zurück.
func geschrieben(t *testing.T, bestand quellenDouble) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "wieder.mse-set")

	if _, err := service.NeuerSetschreiber(bestand, &ortsschreiber{}).
		Schreibe(t.Context(), 7, ort); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	inhalt, err := os.ReadFile(filepath.Join(ort, "set")) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	return string(inhalt)
}

func TestEineGeloeschteKarteNimmtIhrenBlockMit(t *testing.T) {
	// Die mittlere Karte ist weg; die Stellen 0 und 2 sind geblieben.
	bestand := mitKarten(
		karteMit(1, 0, "Feuerball"),
		karteMit(3, 2, "Doppelt"),
	)

	erwartet := "mse version: 2.0.0\n" +
		"game: probe\n" +
		"card:\n" +
		"\tname: Feuerball\n" +
		"card:\n" +
		"\tname: Doppelt\n" +
		"unverstandener block:\n" +
		"\tirgendwas: bleibt liegen\n"

	if gelesen := geschrieben(t, bestand); gelesen != erwartet {
		t.Errorf("geschrieben:\n%q\nerwartet:\n%q", gelesen, erwartet)
	}
}

// **Und die letzte nimmt ihren Block genauso mit.** Beim Abzählen fiel dieser
// Fall zufällig richtig aus — der Block stand hinter der letzten Karte und
// blieb einfach stehen. Er tat es aus dem falschen Grund.
func TestDieLetzteGeloeschteKarteNimmtIhrenBlockMit(t *testing.T) {
	bestand := mitKarten(
		karteMit(1, 0, "Feuerball"),
		karteMit(2, 1, "Eisball"),
	)

	erwartet := "mse version: 2.0.0\n" +
		"game: probe\n" +
		"card:\n" +
		"\tname: Feuerball\n" +
		"card:\n" +
		"\tname: Eisball\n" +
		"unverstandener block:\n" +
		"\tirgendwas: bleibt liegen\n"

	if gelesen := geschrieben(t, bestand); gelesen != erwartet {
		t.Errorf("geschrieben:\n%q\nerwartet:\n%q", gelesen, erwartet)
	}
}

// **Eine neu angelegte Karte bekommt einen Block**
// (#623).
//
// Sie hat keinen, denn die Datei ist älter als sie. Bis hierher lief der
// Schreiber nur die vorhandenen Blöcke durch — wer eine Karte anlegte und
// speicherte, bekam eine Datei ohne sie. Das Anlegen war damit **stiller
// Datenverlust beim Speichern**, und kein Test sah es.
//
// **Sie geht hinter den letzten Kartenblock**, nicht ans Ende der Datei: die
// Karten eines Sets stehen beieinander, und der unverstandene Block gehört
// nicht zwischen sie.
func TestEineNeueKarteBekommtIhrenBlock(t *testing.T) {
	bestand := mitKarten(
		karteMit(1, 0, "Feuerball"),
		karteMit(2, 1, "Eisball"),
		karteMit(3, 2, "Doppelt"),
		karteMit(4, 3, "Neuling"),
	)

	erwartet := dreikartenquelle[:len(dreikartenquelle)-len("unverstandener block:\n\tirgendwas: bleibt liegen\n")] +
		"card:\n" +
		"\tname: Neuling\n" +
		"unverstandener block:\n" +
		"\tirgendwas: bleibt liegen\n"

	if gelesen := geschrieben(t, bestand); gelesen != erwartet {
		t.Errorf("geschrieben:\n%q\nerwartet:\n%q", gelesen, erwartet)
	}
}

// **Führt der Bestand gar keine Karte, bleibt kein Block stehen.**
//
// Hier stand bis #621
// das Gegenteil — `TestEineKarteZuVielBleibtStehen`, mit der Begründung:
//
//	Löschen gibt es noch nicht (#52). Eine Karte stillschweigend zu entfernen,
//	weil der Bestand sie nicht führt, wäre die schlimmere Antwort.
//
// **Die Voraussetzung ist weggefallen.** Löschen gibt es, und der Import wirft
// keinen Kartenblock weg: jeder wird eine Karte, auch der leere
// ([`import.go`](import.go)). Ein Block ohne Karte im Bestand hat deshalb genau
// eine Ursache — jemand hat gelöscht. Ihn stehen zu lassen machte aus dem
// Speichern eine Rücknahme.
func TestOhneKarteImBestandBleibtKeinBlockStehen(t *testing.T) {
	erwartet := "mse version: 2.0.0\n" +
		"game: probe\n" +
		"unverstandener block:\n" +
		"\tirgendwas: bleibt liegen\n"

	if gelesen := geschrieben(t, mitKarten()); gelesen != erwartet {
		t.Errorf("geschrieben:\n%q\nerwartet:\n%q", gelesen, erwartet)
	}
}

// **Und eine neue Karte findet auch dann ihren Platz** — hinter die Stelle, an
// der die Karten standen, nicht ans Ende hinter den unverstandenen Block.
func TestEineNeueKarteStehtDortWoDieKartenStanden(t *testing.T) {
	erwartet := "mse version: 2.0.0\n" +
		"game: probe\n" +
		"card:\n" +
		"\tname: Neuling\n" +
		"unverstandener block:\n" +
		"\tirgendwas: bleibt liegen\n"

	if gelesen := geschrieben(t, mitKarten(karteMit(4, 3, "Neuling"))); gelesen != erwartet {
		t.Errorf("geschrieben:\n%q\nerwartet:\n%q", gelesen, erwartet)
	}
}

// **Und trägt die Datei überhaupt keinen Kartenblock, geht die neue ans Ende.**
//
// So sieht eine `.mse-set` aus, die MSE2 nie mit Karten gesehen hat. Es gibt
// keine Stelle, hinter die die neue gehörte — also hinter alles.
func TestOhneKartenblockGehtDieNeueAnsEnde(t *testing.T) {
	bestand := quellenDouble{
		leserDouble: leserDouble{
			koepfe: []domain.SetHead{{ID: 7, Name: "Mein Probeset", Game: "probe"}},
			karten: []*domain.CardRow{karteMit(1, 0, "Neuling")},
		},
		quelle: &domain.Quelle{
			Datei:  "set",
			Inhalt: []byte("mse version: 2.0.0\ngame: probe\n"),
		},
	}

	erwartet := "mse version: 2.0.0\ngame: probe\ncard:\n\tname: Neuling\n"

	if gelesen := geschrieben(t, bestand); gelesen != erwartet {
		t.Errorf("geschrieben:\n%q\nerwartet:\n%q", gelesen, erwartet)
	}
}
