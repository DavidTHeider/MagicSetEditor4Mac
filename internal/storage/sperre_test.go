// **Zwei Prozesse an einem Bestand** — Befund Z6-02 von Phase 3.2
// (#1163).
//
// # Warum ein echter zweiter Prozess
//
// `flock` gilt für die **offene Datei**, nicht für den Prozess. Zwei Griffe im
// selben Programm ständen einander im Weg — ein Test mit zwei Griffen prüfte
// deshalb, dass sich die Anwendung gegen sich selbst aussperrt, und das ist
// das Gegenteil der Zusage.
//
// Der zweite Prozess ist der Testlauf selbst, noch einmal gestartet: `go test`
// baut ein Programm, und dieses Programm kann sich mit einem Umgebungsschalter
// als **Helfer** verhalten (`TestMain`). Das ist der übliche Weg für
// Mehrprozessläufe in Go und braucht kein zweites Ziel im Bau.
package storage_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// helferschalter sagt dem zweiten Prozess, was er tun soll.
const helferschalter = "MSE_SPERRHELFER"

// oeffnerschalter sagt dem zweiten Prozess, welchen Bestand er öffnen soll.
const oeffnerschalter = "MSE_SPERROEFFNER"

// TestMain lässt den Prozess auch Helfer sein — in vier Rollen; die dritte
// sichert und bricht dabei ab (`sicherungsabbruch_test.go`), die vierte
// sichert unter einer Grenze für die Dateigröße
// (`sicherungsschreibfehler_test.go`).
func TestMain(m *testing.M) {
	if angabe := os.Getenv(abbruchschalter); angabe != "" {
		sichereUndBrichAb(angabe)

		return
	}

	if angabe := os.Getenv(grenzschalter); angabe != "" {
		sichereUnterGrenze(angabe)

		return
	}

	if pfad := os.Getenv(helferschalter); pfad != "" {
		halteDieSperre(pfad)

		return
	}

	if pfad := os.Getenv(oeffnerschalter); pfad != "" {
		oeffneDenBestand(pfad)

		return
	}

	os.Exit(m.Run())
}

// oeffneDenBestand öffnet über den Weg, den jeder Verbraucher nimmt.
func oeffneDenBestand(pfad string) {
	db, err := storage.Oeffnen(context.Background(), pfad)
	if err != nil {
		os.Exit(2)
	}

	// **Ohne `defer`**: hier steht `os.Exit`, und danach läuft keiner mehr.
	// Der Prozess endet ohnehin, und mit ihm der Griff.
	if _, schreibfehler := os.Stdout.WriteString("offen\n"); schreibfehler != nil {
		_ = db.Close() //nolint:errcheck // gleich ist der Prozess fort

		os.Exit(3)
	}

	_ = db.Close() //nolint:errcheck // gleich ist der Prozess fort
}

// halteDieSperre nimmt den Alleinanspruch und wartet auf ein Zeichen.
//
// **Über die Standardeingabe und nicht über eine Frist**: der Aufrufer
// entscheidet, wann losgelassen wird, und ein Test, der auf eine Sekunde
// wartet, wartet auf einem langsamen Rechner zu kurz.
func halteDieSperre(pfad string) {
	zurueck, fehler := storage.Alleinanspruch(pfad)
	if fehler != nil {
		os.Exit(2)
	}

	// **Erst melden, dann warten.** Der Aufrufer darf nicht weitermachen,
	// bevor die Sperre wirklich liegt — sonst prüfte er ein Fenster, das noch
	// niemand besetzt hat (§10).
	if _, err := os.Stdout.WriteString("gefasst\n"); err != nil {
		os.Exit(3)
	}

	_, _ = os.Stdin.Read(make([]byte, 1)) //nolint:errcheck // das Zeichen zum Loslassen

	_ = zurueck() //nolint:errcheck // der Prozess endet ohnehin
}

// zweiterProzess startet den Helfer und wartet, bis er die Sperre hält.
func zweiterProzess(t *testing.T, pfad string) (loslassen func()) {
	t.Helper()

	befehl := exec.CommandContext(probekontext(t), os.Args[0]) // #nosec G204,G702 -- der Testlauf selbst
	befehl.Env = append(os.Environ(), helferschalter+"="+pfad)

	hinein, err := befehl.StdinPipe()
	if err != nil {
		t.Fatalf("Eingabe: %v", err)
	}

	heraus, err := befehl.StdoutPipe()
	if err != nil {
		t.Fatalf("Ausgabe: %v", err)
	}

	if fehler := befehl.Start(); fehler != nil {
		t.Fatalf("starten: %v", fehler)
	}

	gemeldet := make([]byte, len("gefasst\n"))
	if _, fehler := heraus.Read(gemeldet); fehler != nil {
		t.Fatalf("der Helfer meldete sich nicht: %v", fehler)
	}

	return func() {
		_, _ = hinein.Write([]byte("x")) //nolint:errcheck // der Helfer endet ohnehin
		_ = befehl.Wait()                //nolint:errcheck // dito
	}
}

// **Solange ein zweiter Prozess die Nachbardatei hält, gibt es keinen
// Alleinanspruch.**
//
// # Was dieser Lauf gegenüber der SQLite-Probe hinzufügt
//
// [storage.AlleinAmBestand] fragt und **lässt wieder los**. Der Anspruch hier
// wird gehalten: zwischen „ich darf" und dem Umbenennen kann niemand mehr
// dazwischenkommen.
func TestEinZweiterProzessVerhindertDenAlleinanspruch(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "bestand.db")

	loslassen := zweiterProzess(t, pfad)
	defer loslassen()

	if _, fehler := storage.Alleinanspruch(pfad); fehler == nil {
		t.Fatal("der Alleinanspruch kam durch, obwohl ein zweiter Prozess hält")
	}
}

// **Und wenn er loslässt, geht es.**
//
// Die Gegenprobe: ohne sie sagte der Lauf darüber auch dann ja, wenn der
// Alleinanspruch **immer** scheiterte.
func TestNachDemLoslassenGehtDerAlleinanspruch(t *testing.T) {
	pfad := filepath.Join(t.TempDir(), "bestand.db")

	loslassen := zweiterProzess(t, pfad)

	if _, fehler := storage.Alleinanspruch(pfad); fehler == nil {
		t.Fatal("der Alleinanspruch kam durch, obwohl ein zweiter Prozess hält")
	}

	loslassen()

	zurueck, fehler := storage.Alleinanspruch(pfad)
	if fehler != nil {
		t.Fatalf("nach dem Loslassen: %v", fehler)
	}

	if zurueckfehler := zurueck(); zurueckfehler != nil {
		t.Errorf("zurückgeben: %v", zurueckfehler)
	}
}

// **Wer öffnet, während ein Tausch läuft, kommt nicht dazwischen.**
//
// # Der eigentliche Befund
//
// Das ist die Richtung, die vorher offenstand: die SQLite-Probe gab ihre
// Sperre frei, und **danach** kam der Tausch. Wer in dieser Lücke öffnete,
// arbeitete auf einem Inode, den es gleich nicht mehr gab.
//
// # Er wartet, er wird nicht abgewiesen
//
// Und das ist richtig: ein Tausch dauert zwei Umbenennungen. Wer in dieser
// Zeit öffnet, soll den **fertigen** Stand bekommen — nicht eine Absage und
// nicht eine Datei, die gleich verschwindet.
//
// **Deterministisch**: der Helfer kommt nicht durch, solange der Anspruch
// liegt, und kommt durch, sobald er fällt. Nicht „selten", sondern jedes Mal.
func TestWerWaehrendEinesTauschsOeffnetWartet(t *testing.T) {
	ort := t.TempDir()
	pfad := filepath.Join(ort, "bestand.db")

	zurueck, fehler := storage.Alleinanspruch(pfad)
	if fehler != nil {
		t.Fatalf("Alleinanspruch: %v", fehler)
	}

	// **Ein zweiter Prozess öffnet den Bestand** — über `storage.Oeffnen`,
	// also über den Weg, den jeder Verbraucher nimmt.
	befehl := exec.CommandContext(probekontext(t), os.Args[0]) // #nosec G204,G702 -- der Testlauf selbst
	befehl.Env = append(os.Environ(), oeffnerschalter+"="+pfad)

	heraus, err := befehl.StdoutPipe()
	if err != nil {
		t.Fatalf("Ausgabe: %v", err)
	}

	if startfehler := befehl.Start(); startfehler != nil {
		t.Fatalf("starten: %v", startfehler)
	}

	t.Cleanup(func() { _ = befehl.Wait() }) //nolint:errcheck // Aufräumen

	gemeldet := make(chan error, 1)

	go func() {
		_, lesefehler := heraus.Read(make([]byte, len("offen\n")))
		gemeldet <- lesefehler
	}()

	select {
	case lesefehler := <-gemeldet:
		t.Errorf("der zweite Prozess öffnete mitten im Tausch (%v)", lesefehler)

	case <-time.After(300 * time.Millisecond):
	}

	if zurueckfehler := zurueck(); zurueckfehler != nil {
		t.Fatalf("zurückgeben: %v", zurueckfehler)
	}

	// **Und sobald der Anspruch fällt, kommt er durch.** Ohne diese Hälfte
	// sagte der Lauf auch dann ja, wenn der Helfer aus einem ganz anderen
	// Grund schwiege (§10).
	select {
	case lesefehler := <-gemeldet:
		if lesefehler != nil {
			t.Errorf("der Helfer meldete sich nicht: %v", lesefehler)
		}

	case <-time.After(10 * time.Second):
		t.Error("der zweite Prozess kam auch nach dem Tausch nicht durch")
	}
}

// **Ohne Sperre geht es weiter.**
//
// Sie ist eine Verbesserung des Schutzes und keine Voraussetzung des Tauschs;
// eine Anwendung, die wegen einer Nachbardatei nicht mehr zurückspielen kann,
// tauschte einen kleinen Mangel gegen einen großen (§8).
func TestOhneSperreGehtEsWeiter(t *testing.T) {
	t.Parallel()

	for name, pfad := range map[string]string{
		"kein Pfad":          "",
		"im Arbeitsspeicher": storage.ImArbeitsspeicher,
		"eine URI":           "file:probe.db?_pragma=busy_timeout(200)",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			zurueck, fehler := storage.Alleinanspruch(pfad)
			if fehler != nil {
				t.Fatalf("Alleinanspruch ohne Sperre: %v", fehler)
			}

			if zurueckfehler := zurueck(); zurueckfehler != nil {
				t.Errorf("zurückgeben ohne Sperre: %v", zurueckfehler)
			}
		})
	}
}

// **Lässt sich die Nachbardatei nicht anlegen, geht es auch weiter.**
func TestOhneNachbardateiGehtEsWeiter(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()

	// **Ein Verzeichnis am Ort der Sperrdatei**: `dateisperre.Oeffnen`
	// scheitert.
	if err := os.Mkdir(filepath.Join(ort, "bestand.db"+storage.Sperrendung), 0o700); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	zurueck, fehler := storage.Alleinanspruch(filepath.Join(ort, "bestand.db"))
	if fehler != nil {
		t.Fatalf("Alleinanspruch ohne Nachbardatei: %v", fehler)
	}

	if zurueckfehler := zurueck(); zurueckfehler != nil {
		t.Errorf("zurückgeben: %v", zurueckfehler)
	}

	// **Und das Öffnen ebenso.** Es nimmt dieselbe Nachbardatei.
	db, oeffnungsfehler := storage.Oeffnen(probekontext(t), filepath.Join(ort, "bestand.db"))
	if oeffnungsfehler != nil {
		t.Fatalf("öffnen ohne Nachbardatei: %v", oeffnungsfehler)
	}

	if schliessfehler := db.Close(); schliessfehler != nil {
		t.Errorf("schließen: %v", schliessfehler)
	}
}

// **Wer den Bestand öffnet, nimmt die Nachbardatei.**
//
// # Die andere Hälfte der Zusage
//
// Die Läufe darüber prüfen den Anspruch für sich. Was sie nicht sagen, ist, ob
// die vier Öffner — Fenster, Kommandozeile, MCP-Server, Dienst — die
// Nachbardatei überhaupt anfassen. Sie tun es nicht selbst: `Oeffnen` tut es
// für sie, denn eine Sperre, die einer von ihnen vergisst, ist keine.
func TestWerDenBestandOeffnetNimmtDieNachbardatei(t *testing.T) {
	t.Parallel()

	pfad := filepath.Join(t.TempDir(), "bestand.db")

	db, err := storage.Oeffnen(probekontext(t), pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Aufräumen

	if _, statfehler := os.Stat(pfad + storage.Sperrendung); statfehler != nil {
		t.Errorf("nach dem Öffnen liegt keine Sperrdatei: %v", statfehler)
	}
}

// **Nach der Frist wird ohne Sperre geöffnet.**
//
// Ein Tausch, der hängenbleibt, hielte sonst **jedes** Öffnen für immer auf.
// Was bleibt, ist der Stand vor Phase 3.2 — schlechter als mit, und besser als
// eine Anwendung, die nicht mehr startet.
func TestNachDerFristWirdOhneSperreGeoeffnet(t *testing.T) {
	ort := t.TempDir()
	pfad := filepath.Join(ort, "bestand.db")

	// **Ein zweiter Prozess hält den Anspruch** — im eigenen ginge es nicht:
	// `flock` gilt für die offene Datei, und ein zweiter Griff hier stünde
	// dem Öffnen ebenso im Weg, aber aus dem falschen Grund.
	loslassen := zweiterProzess(t, pfad)
	defer loslassen()

	alteFrist, alterTakt := storage.Sperrfrist, storage.Sperrtakt
	storage.Sperrfrist, storage.Sperrtakt = 60*time.Millisecond, 10*time.Millisecond

	t.Cleanup(func() { storage.Sperrfrist, storage.Sperrtakt = alteFrist, alterTakt })

	beginn := time.Now()

	db, err := storage.Oeffnen(probekontext(t), pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Aufräumen

	if gedauert := time.Since(beginn); gedauert < 60*time.Millisecond {
		t.Errorf("es dauerte nur %v — es wurde gar nicht gewartet", gedauert)
	}
}
