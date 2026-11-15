// **Eine Sicherung, die entsteht, während geschrieben wird** (#798).
package storage

import (
	"bytes"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// aufDerPlatte legt eine migrierte Datenbank in einer Datei an.
func aufDerPlatte(t *testing.T) (db *sql.DB, pfad string) {
	t.Helper()

	pfad = filepath.Join(t.TempDir(), "mse.db")

	db, fehler := Oeffnen(probekontext(t), pfad)
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	if fehler := Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	return db, pfad
}

// **Die Sicherung enthält, was der laufende Schreiber geschrieben hat.**
//
// # Warum hier ein Signal steht und kein `go func` genügt
//
// Der erste Anlauf startete eine Goroutine mit 200 Einfügungen und sicherte
// „dazwischen". **Belegt hat das nichts**: zwischen dem Start der Goroutine
// und `Sichern` besteht kein Happens-before — die Sicherung konnte fertig
// sein, bevor die erste Einfügung lief, und dann maß der Test einen leeren
// Bestand gegen sich selbst. Der Zweitblick über Phase N hat es angestrichen.
//
// Jetzt wartet der Sichernde auf ein **Signal**, dass mindestens eine Zeile
// steht, und der Schreiber läuft danach weiter. Damit ist die Überlappung
// keine Hoffnung mehr, sondern eine Reihenfolge — und die Zusicherung unten
// prüft, was in der Sicherung **drinsteht**, nicht nur, dass sie sich öffnen
// lässt.
func TestEineSicherungEntstehtWaehrendGeschriebenWird(t *testing.T) {
	db, _ := aufDerPlatte(t)
	ziel := filepath.Join(t.TempDir(), "sicherung.db")

	var wg sync.WaitGroup

	ersteSteht := make(chan struct{})

	wg.Add(1)

	// **Das Signal fällt in jedem Fall** — auch auf dem Fehlerweg. Bis zum
	// 2026-09-04 stand `close` nur im Erfolgszweig, und ein gescheitertes
	// `INSERT` ließ den Lesenden unten **fünf Minuten** warten, bis der
	// Testlauf abbrach. Eine Meldung, die als Zeitüberschreitung ankommt,
	// nennt ihre Ursache nicht.
	go func() {
		defer wg.Done()
		defer func() {
			select {
			case <-ersteSteht:
			default:
				close(ersteSteht)
			}
		}()

		for i := range 200 {
			if _, fehler := db.ExecContext(probekontext(t),
				`INSERT INTO sets (name, spiel, geoeffnet, profil_id) VALUES (?, 'probe', ?, 1)`,
				"Set", "2026-08-22T00:00:00.000Z"); fehler != nil {
				t.Errorf("Schreiber bei %d: %v", i, fehler)

				return
			}

			if i == 0 {
				close(ersteSteht)
			}
		}
	}()

	<-ersteSteht

	if fehler := Sichern(probekontext(t), db, ziel); fehler != nil {
		t.Fatalf("sichern: %v", fehler)
	}

	wg.Wait()

	kopie, fehler := Oeffnen(probekontext(t), ziel)
	if fehler != nil {
		t.Fatalf("die Sicherung öffnen: %v", fehler)
	}

	defer func() { _ = kopie.Close() }()

	if fehler := PruefeBestand(probekontext(t), kopie); fehler != nil {
		t.Errorf("die Sicherung besteht die eigene Prüfung nicht: %v", fehler)
	}

	if fehler := Migriere(probekontext(t), kopie, nil); fehler != nil {
		t.Errorf("die Sicherung lässt sich nicht migrieren: %v", fehler)
	}

	// **Und sie führt, was der Schreiber geschrieben hatte.**
	//
	// Ohne diese Zusicherung sagte der Test nur, dass eine öffenbare Datei
	// entstanden ist — und eine leere ist öffenbar. Zugesichert wird das
	// Belegbare: **mindestens die eine**, deren Signal oben abgewartet wurde.
	// Wie viele der 200 es sind, entscheidet der Zeitpunkt des Snapshots, und
	// darüber sollte dieser Test nichts behaupten.
	var gesichert int
	if fehler := kopie.QueryRowContext(probekontext(t),
		`SELECT count(*) FROM sets`).Scan(&gesichert); fehler != nil {
		t.Fatalf("die Sicherung zählen: %v", fehler)
	}

	if gesichert < 1 {
		t.Errorf("%d Sets in der Sicherung — der Schreiber hatte mindestens "+
			"eines geschrieben, bevor sie begann", gesichert)
	}

	t.Logf("die Sicherung führt %d von bis zu 200 Sets", gesichert)
}

// **Sie kommt ohne `-wal` und `-shm` aus** — nachgemessen, nicht behauptet.
func TestDieSicherungIstEineEinzigeDatei(t *testing.T) {
	db, _ := aufDerPlatte(t)

	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO sets (name, spiel, geoeffnet, profil_id) VALUES ('eins', 'probe', ?, 1)`,
		"2026-08-22T00:00:00.000Z"); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	ordner := t.TempDir()
	if fehler := Sichern(probekontext(t), db, filepath.Join(ordner, "sicherung.db")); fehler != nil {
		t.Fatalf("sichern: %v", fehler)
	}

	eintraege, fehler := os.ReadDir(ordner)
	if fehler != nil {
		t.Fatalf("Ordner lesen: %v", fehler)
	}

	if len(eintraege) != 1 {
		for _, eintrag := range eintraege {
			t.Logf("  %s", eintrag.Name())
		}

		t.Errorf("%d Dateien neben der Sicherung, erwartet war eine", len(eintraege))
	}
}

// **Die Gegenprobe: eine nackte Dateikopie verliert etwas.**
//
// Sie steht hier, weil der Satz „`VACUUM INTO` ist nötig" sonst eine
// Behauptung wäre. Kopiert wird allein `mse.db`, so wie jemand es täte, der
// „die Datei" sichert — und was seit dem letzten Checkpoint im `-wal` steht,
// ist danach fort.
func TestEineNackteDateikopieVerliertDenWal(t *testing.T) {
	db, pfad := aufDerPlatte(t)

	for i := range 50 {
		if _, fehler := db.ExecContext(probekontext(t),
			`INSERT INTO sets (name, spiel, geoeffnet, profil_id) VALUES (?, 'probe', ?, 1)`,
			"Set", "2026-08-22T00:00:00.000Z"); fehler != nil {
			t.Fatalf("schreiben bei %d: %v", i, fehler)
		}
	}

	roh, fehler := os.ReadFile(pfad) // #nosec G304,G703 -- Pfad aus t.TempDir()
	if fehler != nil {
		t.Fatalf("die Datei lesen: %v", fehler)
	}

	kopiepfad := filepath.Join(t.TempDir(), "kopie.db")
	if schreibfehler := os.WriteFile(kopiepfad, roh, 0o600); schreibfehler != nil { // #nosec G703 -- dito
		t.Fatalf("kopieren: %v", schreibfehler)
	}

	kopie, fehler := Oeffnen(probekontext(t), kopiepfad)
	if fehler != nil {
		t.Fatalf("die Kopie öffnen: %v", fehler)
	}

	defer func() { _ = kopie.Close() }()

	// **Es ist schlimmer als „ein paar Zeilen fehlen".**
	//
	// Der erste Anlauf zählte die Sets in der Kopie und erwartete eine Zahl
	// unter 50. Er scheiterte an etwas anderem: `no such table: sets`. Ohne
	// Checkpoint steht **das ganze Schema** im `-wal`, und die Hauptdatei ist
	// so gut wie leer. Wer sie kopiert, hat keine ältere Datenbank — er hat
	// keine.
	//
	// Zugesichert wird deshalb das Schwächere von beidem: die Kopie zeigt
	// **nicht** alle fünfzig. Ob sie dabei stolpert oder weniger zählt, hängt
	// davon ab, wann SQLite zuletzt einen Checkpoint gesetzt hat, und das ist
	// keine Zusage, die dieser Test machen sollte.
	var inDerKopie int

	fehler = kopie.QueryRowContext(probekontext(t),
		`SELECT count(*) FROM sets`).Scan(&inDerKopie)
	if fehler != nil {
		t.Logf("die nackte Dateikopie ist nicht einmal lesbar: %v", fehler)

		return
	}

	if inDerKopie == 50 {
		t.Errorf("die nackte Kopie führt alle 50 Sets — dann belegt dieser " +
			"Test nichts über das, was ohne den -wal fehlt")
	}

	t.Logf("die nackte Dateikopie führt %d von 50 Sets", inDerKopie)
}

// **Ein unmögliches Ziel meldet sich als Fehler.**
func TestEineSicherungInEinNichtVorhandenesVerzeichnisScheitert(t *testing.T) {
	db, _ := aufDerPlatte(t)

	ziel := filepath.Join(t.TempDir(), "gibtesnicht", "sicherung.db")

	if fehler := Sichern(probekontext(t), db, ziel); fehler == nil {
		t.Error("kein Fehler, obwohl das Verzeichnis fehlt")
	}
}

// **Ein Anführungszeichen im Dateinamen bleibt ein Dateiname.**
//
// Der Pfad geht als Platzhalter hinein, seit gemessen ist, dass `VACUUM INTO ?`
// ihn annimmt. Der Test bleibt trotzdem stehen: er belegt, dass die Bindung
// wirklich bindet, statt dass es jemand annimmt.
func TestEinAnfuehrungszeichenImNamenWirdMitgeschrieben(t *testing.T) {
	db, _ := aufDerPlatte(t)

	ziel := filepath.Join(t.TempDir(), "Davids' Sicherung.db")
	if fehler := Sichern(probekontext(t), db, ziel); fehler != nil {
		t.Fatalf("sichern: %v", fehler)
	}

	if _, fehler := os.Stat(ziel); fehler != nil {
		t.Errorf("die Datei liegt nicht unter ihrem Namen: %v", fehler)
	}
}

// **Ein leeres Ziel wird nicht angenommen, sondern abgewiesen.**
//
// # Der Befund dahinter
//
// SQLite weist ein vorhandenes Ziel nur ab, wenn **Inhalt** darin steht;
// nachgemessen am 2026-08-22 gibt eine Datei mit einem Byte `file is not a
// database`, eine leere wird angenommen und gefüllt. Der erste Anlauf sah
// deshalb mit `os.Stat` nach und schrieb daneben, ein Wettlauf bleibe zwar,
// aber „nichts wird überschrieben" — was für die leere Datei nicht stimmte.
//
// Der Test steht hier für **beide** Gestalten: leer und gefüllt. Ohne die
// leere prüfte er den Fall nicht, den SQLite anders behandelt als den anderen.
func TestEineSicherungWeistAuchEinLeeresZielAb(t *testing.T) {
	db, _ := aufDerPlatte(t)

	for name, inhalt := range map[string][]byte{
		"leer":       nil,
		"mit Inhalt": []byte("nicht anfassen"),
	} {
		t.Run(name, func(t *testing.T) {
			ziel := filepath.Join(t.TempDir(), "schon-da.db")
			if fehler := os.WriteFile(ziel, inhalt, 0o600); fehler != nil {
				t.Fatalf("Ziel anlegen: %v", fehler)
			}

			if fehler := Sichern(probekontext(t), db, ziel); !errors.Is(fehler, domain.ErrConflict) {
				t.Fatalf("Fehler %v, erwartet war ein Konflikt", fehler)
			}

			// **Und die Datei ist so geblieben, wie sie war.**
			gelesen, fehler := os.ReadFile(ziel) // #nosec G304 -- Pfad aus t.TempDir()
			if fehler != nil {
				t.Fatalf("Ziel wiederlesen: %v", fehler)
			}

			if !bytes.Equal(gelesen, inhalt) {
				t.Errorf("die Datei wurde angetastet: %d statt %d Bytes",
					len(gelesen), len(inhalt))
			}
		})
	}
}

// **Scheitert das Sichern, bleibt keine leere Reservierung zurück.**
//
// Der Name wird vor dem `VACUUM INTO` angelegt; ohne Aufräumen stünde nach
// einem Fehlschlag eine leere Datei da — und der zweite Versuch scheiterte an
// ihr, mit der Meldung „liegt schon da". Ein halber Stand, den §8 nicht
// zulässt.
func TestEinGescheitertesSichernRaeumtSeinenPlatzWiederWeg(t *testing.T) {
	db, _ := aufDerPlatte(t)

	// Eine geschlossene Datenbank kann nicht sichern — der Platz ist da
	// bereits angelegt.
	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	ziel := filepath.Join(t.TempDir(), "sicherung.db")

	if fehler := Sichern(probekontext(t), db, ziel); fehler == nil {
		t.Fatal("kein Fehler, obwohl die Datenbank zu ist")
	}

	if _, fehler := os.Stat(ziel); fehler == nil {
		t.Error("die Reservierung liegt noch da")
	}
}
