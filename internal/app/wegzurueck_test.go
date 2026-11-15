//go:build darwin

// **Der Weg zurück wird benannt** — Werkstück 5 von Phase 4.13
// (#1352).
//
// # Die Frage aus dem Issue
//
// *„Was passiert, wenn eine neue Fassung ein Schema mitbringt, das die alte
// nicht liest — und die Person startet danach wieder die alte?"*
//
// **Die Abweisung steht seit Werkstück 3** und trägt ihren Schlüssel. Ihr Satz
// endet mit „oder spiele eine Sicherung von vorher zurück" — und **welche**,
// sagt er nicht. Neben dem Bestand liegen bis zu dreißig Dateien (Werkstück
// 4), ihre Namen tragen Fassungsnummern, und der Ordner liegt dort, wo ihn
// niemand sucht. Ein Rat, den man nicht befolgen kann, ist eine Sackgasse mit
// Höflichkeitsformel.
package app_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Befund nennt die Datei, die diese Fassung wirklich öffnen kann.**
//
// # Warum „wirklich" das Wort ist, auf das es ankommt
//
// Die **jüngste** Sicherung ist nicht der Weg zurück. Wer von 0.3.0 über
// 0.4.0 auf 0.5.0 gegangen ist, hat neben dem Bestand
//
//	bestand-vor-0.4.0.db   den Stand von 0.3.0 — **diese Fassung liest ihn**
//	bestand-vor-0.5.0.db   den Stand von 0.4.0 — zu neu, jünger, und **kein**
//	                       Weg zurück für 0.3.0
//
// Die zweite ist die jüngere. Wer „die jüngste" nennte, nennte die falsche —
// und der Rat führte in dieselbe Abweisung, aus der er herausführen soll.
func TestDerBefundNenntDieSicherungDieDieseFassungOeffnet(t *testing.T) {
	bestand := bestandMitSicherungen(t)

	alte := app.NeueApp(bestand)

	if fehler := alte.Starten(t.Context()); fehler == nil {
		t.Fatal("die alte Fassung öffnete einen Bestand aus der Zukunft")
	}

	befund := alte.Startbefund()

	if befund.Schluessel != string(domain.SchluesselBestandNeuereFassung) {
		t.Fatalf("Schlüssel %q", befund.Schluessel)
	}

	if befund.Rueckweg == nil {
		t.Fatal("der Befund nennt keinen Weg zurück")
	}

	if befund.Rueckweg.Datei != "bestand-vor-0.2.0.db" {
		t.Errorf("genannt ist %q — das ist nicht die Datei, die aufgeht",
			befund.Rueckweg.Datei)
	}
}

// **Und er nennt keinen, wenn keiner daliegt.**
//
// **Die Gegenprobe zur Zusage** (§10): ein Befund, der immer eine Datei nennt,
// nennt irgendwann eine, die es nicht gibt — und das ist schlimmer als
// Schweigen, weil jemand danach sucht.
func TestOhneBrauchbareSicherungNenntDerBefundKeinen(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)

	if fehler := anwendung.Beenden(); fehler != nil {
		t.Fatalf("beenden: %v", fehler)
	}

	// **Eine Datei, die den Namen trägt und keine Sicherung ist.** Sie ist der
	// Fall, den ein Blick auf das Namensmuster allein durchließe.
	neben := filepath.Dir(bestand)
	if fehler := os.WriteFile(
		filepath.Join(neben, "bestand-vor-0.9.0.db"), []byte("nichts"), 0o600,
	); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	zukunftEintragen(t, bestand)

	alte := app.NeueApp(bestand)
	if fehler := alte.Starten(t.Context()); fehler == nil {
		t.Fatal("die alte Fassung öffnete einen Bestand aus der Zukunft")
	}

	if weg := alte.Startbefund().Rueckweg; weg != nil {
		t.Errorf("genannt ist %+v — es liegt keine brauchbare Sicherung da", *weg)
	}
}

// **Und das Wegräumen nimmt ihn nicht mit.**
//
// # Warum „die jüngste bleibt" dafür nicht reicht
//
// Werkstück 4 ließ die jüngste Datei stehen, und das ist im Normalfall der
// Weg zurück. **Im Ernstfall ist es die falsche**: wer auf der alten Fassung
// steht, deren Bestand nicht aufgeht, sieht den Platz — die Fläche arbeitet
// ohne offene Datenbank, und das ist Absicht — und räumt auf. Die jüngste
// Sicherung ist dann die, die er nicht lesen kann.
//
// **Es bleiben deshalb zwei**, und die Gründe sind verschieden: die jüngste
// **lesbare** ist der Weg zurück für diese Fassung, die jüngste **überhaupt**
// der für die, die sie geschrieben hat — und die installiert vielleicht
// als Nächstes wieder jemand.
func TestDasWegraeumenLaesstDenWegZurueckStehen(t *testing.T) {
	bestand := bestandMitSicherungen(t)
	neben := filepath.Dir(bestand)

	antwort := app.PlatzflaecheUeber(bestand).PlatzFreigeben(
		t.Context(), string(domain.PostenVorsicherungen))
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	for name, sollBleiben := range map[string]bool{
		"bestand-vor-0.1.0.db": false,
		"bestand-vor-0.2.0.db": true,
		"bestand-vor-0.3.0.db": true,
	} {
		_, fehler := os.Stat(filepath.Join(neben, name))

		if geblieben := fehler == nil; geblieben != sollBleiben {
			t.Errorf("%s: geblieben %v, erwartet %v", name, geblieben, sollBleiben)
		}
	}
}

// bestandMitSicherungen baut den Ernstfall: ein Bestand aus der Zukunft, und
// daneben drei Sicherungen, von denen genau eine der Weg zurück ist.
//
// **Die Reihenfolge auf der Platte ist Teil des Falls.** `zukunft` ist die
// jüngste Datei und die falsche Antwort; `lesbar` liegt davor.
func bestandMitSicherungen(t *testing.T) string {
	t.Helper()

	anwendung, bestand := gestarteteAnwendung(t)

	if fehler := anwendung.Beenden(); fehler != nil {
		t.Fatalf("beenden: %v", fehler)
	}

	neben := filepath.Dir(bestand)
	jetzt := time.Now()

	aeltere := kopiereBestand(t, bestand, filepath.Join(neben, "bestand-vor-0.1.0.db"))
	lesbare := kopiereBestand(t, bestand, filepath.Join(neben, "bestand-vor-0.2.0.db"))
	kuenftige := kopiereBestand(t, bestand, filepath.Join(neben, "bestand-vor-0.3.0.db"))

	// **Die jüngste ist die aus der Zukunft** — sonst prüfte der Lauf nichts.
	zukunftEintragen(t, kuenftige)
	zukunftEintragen(t, bestand)

	for datei, alter := range map[string]time.Duration{
		aeltere: 3 * time.Hour, lesbare: 2 * time.Hour, kuenftige: time.Hour,
	} {
		wann := jetzt.Add(-alter)
		if fehler := os.Chtimes(datei, wann, wann); fehler != nil {
			t.Fatalf("Zeit setzen: %v", fehler)
		}
	}

	return bestand
}

// kopiereBestand legt eine Sicherung neben den Bestand und gibt ihren Pfad.
func kopiereBestand(t *testing.T, von, nach string) string {
	t.Helper()

	inhalt, fehler := os.ReadFile(von) //nolint:gosec // ein Pfad aus dem Test selbst
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	//nolint:gosec // beide Pfade baut der Test selbst aus t.TempDir()
	if fehler := os.WriteFile(nach, inhalt, 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	return nach
}

// **Die Flächen ohne Fenster bekommen denselben Satz — angehängt.**
//
// **Zwei Läufe in einem**, und der zweite ist die Zusage: liegt keine
// brauchbare Sicherung da, bleibt der Fehler unverändert. Ein Satz, der immer
// kommt, nennt irgendwann eine Datei, die es nicht gibt.
func TestMitWegZurueckHaengtAnUndErfindetNichts(t *testing.T) {
	bestand := bestandMitSicherungen(t)

	scheitern := errors.New("journal_mode: file is not a database")

	mit := app.MitWegZurueck(t.Context(), bestand, scheitern)

	if !errors.Is(mit, scheitern) {
		t.Error("der Grund ging beim Anhängen verloren")
	}

	if !strings.Contains(mit.Error(), "bestand-vor-0.2.0.db") {
		t.Errorf("der Weg zurück fehlt: %v", mit)
	}

	// **Und ohne Ordner gibt es nichts anzuhängen.** Der Fehler bleibt der,
	// der er war — nicht ein zweiter über das Nachsehen.
	ohne := app.MitWegZurueck(
		t.Context(), filepath.Join(t.TempDir(), "gibtsnicht", "bestand.db"),
		scheitern)

	if ohne.Error() != scheitern.Error() {
		t.Errorf("aus einem Ordner, den es nicht gibt, kam %q", ohne)
	}
}

// **Und wo nichts liegt, bleibt beim Wegräumen auch nichts liegen.**
//
// Der Zweig, den `bleibendeVorsicherungen` für die leere Liste hat: er darf
// nicht auf `dateien[0]` greifen, und die Antwort ist trotzdem eine Liste.
func TestWegraeumenOhneEineEinzigeSicherung(t *testing.T) {
	anwendung, bestand := gestarteteAnwendung(t)

	if fehler := anwendung.Beenden(); fehler != nil {
		t.Fatalf("beenden: %v", fehler)
	}

	antwort := app.PlatzflaecheUeber(bestand).PlatzFreigeben(
		t.Context(), string(domain.PostenVorsicherungen))
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	for _, posten := range antwort.Posten {
		if posten.Posten == string(domain.PostenVorsicherungen) &&
			posten.Dateien != 0 {
			t.Errorf("es blieben %d Dateien liegen", posten.Dateien)
		}
	}
}

// **Ein Begleiter verdrängt keine Sicherung — und geht mit seiner.**
//
// # Der Befund, an der laufenden Anwendung gesehen
//
// **Am 2026-09-07 in der Sichtprüfung**: nach dem Wegräumen standen
// `bestand-vor-0.4.0.db` und `bestand-vor-0.4.0.db-shm` da — zwei Dateien, aber
// **ein** Rückweg. Der Platz, den [App.bleibendeVorsicherungen] für die
// jüngste Datei überhaupt freihält, war von einem Begleiter belegt.
//
// **Und der entsteht beim Prüfen selbst.** Gemessen am selben Tag: eine
// `PruefeSicherung` auf einer Datenbank im WAL-Modus legt `-shm` und `-wal`
// daneben an, obwohl sie mit `mode=ro` öffnet. Der Begleiter ist damit
// **immer** der jüngste — er entsteht in dem Augenblick, in dem gefragt wird.
//
// **Ein Test hätte das nicht gezeigt.** Er sieht die Zahl, und zwei Dateien
// sind zwei Dateien.
//
// # Zwei Hälften, und die zweite ist die, die man vergisst
//
//  1. Ein Begleiter zählt **nicht** als Kandidat für „die jüngste".
//  2. Er geht **mit** seiner Sicherung — bleibt sie, bleibt er; geht sie, geht
//     er. Eine Sicherung ohne ihren `-wal` ist der Stand von vorgestern.
func TestEinBegleiterVerdraengtKeineSicherung(t *testing.T) {
	bestand := bestandMitSicherungen(t)
	neben := filepath.Dir(bestand)

	begleiterAnlegen(t, filepath.Join(neben, "bestand-vor-0.1.0.db"), time.Now())
	begleiterAnlegen(t, filepath.Join(neben, "bestand-vor-0.3.0.db"), time.Now())

	antwort := app.PlatzflaecheUeber(bestand).PlatzFreigeben(
		t.Context(), string(domain.PostenVorsicherungen))
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	for name, sollBleiben := range map[string]bool{
		// **Die älteste geht — samt ihren Begleitern**, obwohl die die
		// jüngsten Dateien im Ordner sind.
		"bestand-vor-0.1.0.db":     false,
		"bestand-vor-0.1.0.db-shm": false,
		"bestand-vor-0.1.0.db-wal": false,

		// **Der Weg zurück bleibt.**
		"bestand-vor-0.2.0.db": true,

		// **Und die jüngste Sicherung überhaupt, mit dem, was zu ihr gehört.**
		"bestand-vor-0.3.0.db":     true,
		"bestand-vor-0.3.0.db-shm": true,
		"bestand-vor-0.3.0.db-wal": true,
	} {
		_, fehler := os.Stat(filepath.Join(neben, name))

		if geblieben := fehler == nil; geblieben != sollBleiben {
			t.Errorf("%s: geblieben %v, erwartet %v", name, geblieben, sollBleiben)
		}
	}
}

// begleiterAnlegen legt `-shm` und `-wal` neben eine Sicherung — so jung, wie
// sie im Betrieb entstehen: **jetzt**, beim Prüfen.
func begleiterAnlegen(t *testing.T, sicherung string, wann time.Time) {
	t.Helper()

	for _, endung := range []string{"-shm", "-wal"} {
		datei := sicherung + endung

		if fehler := os.WriteFile(datei, make([]byte, 32), 0o600); fehler != nil {
			t.Fatalf("vorbereiten: %v", fehler)
		}

		if fehler := os.Chtimes(datei, wann, wann); fehler != nil {
			t.Fatalf("Zeit setzen: %v", fehler)
		}
	}
}
