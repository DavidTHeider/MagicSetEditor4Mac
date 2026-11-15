// **Der Tausch der Bestandsdatei, für sich** (#940).
//
// Ein Test über die ganze Anwendung sagt, dass es geht; er sagt nicht, was
// zurückbleibt, wenn es mittendrin scheitert. Genau das steht hier.
package service

import (
	"os"
	"path/filepath"
	"testing"
)

// legeAn schreibt eine Datei mit Inhalt und meldet einen Fehlschlag sofort.
func legeAn(t *testing.T, pfad, inhalt string) {
	t.Helper()

	if err := os.WriteFile(pfad, []byte(inhalt), 0o600); err != nil {
		t.Fatalf("%s anlegen: %v", pfad, err)
	}
}

// inhalt liest eine Datei oder meldet, dass es sie nicht gibt.
func inhalt(t *testing.T, pfad string) string {
	t.Helper()

	roh, err := os.ReadFile(pfad) // #nosec G304 -- Testpfad aus t.TempDir()
	if err != nil {
		return "«fehlt»"
	}

	return string(roh)
}

func TestBestandErsetzenLegtDieSicherungAnDieStelle(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	bestand := filepath.Join(ort, "bestand.db")
	sicherung := filepath.Join(ort, "sicherung.db")

	legeAn(t, bestand, "alt")
	legeAn(t, bestand+"-wal", "altes Journal")
	legeAn(t, bestand+"-shm", "alter Index")
	legeAn(t, sicherung, "neu")

	zurueck, err := bestandErsetzen(sicherung, bestand)
	if err != nil {
		t.Fatalf("der Tausch schlug fehl: %v", err)
	}

	if got := inhalt(t, bestand); got != "neu" {
		t.Errorf("am Ort des Bestands steht %q, erwartet war %q", got, "neu")
	}

	// **Die Nebendateien gehören zum alten Stand.** Bleiben sie liegen,
	// öffnet SQLite den neuen Bestand mit dem Journal des alten.
	for _, endung := range []string{"-wal", "-shm"} {
		if _, err := os.Stat(bestand + endung); !os.IsNotExist(err) {
			t.Errorf("%s liegt noch da", "bestand.db"+endung)
		}
	}

	// **Der alte Stand ist nicht fort**, sondern beiseite.
	if got := inhalt(t, filepath.Join(ort, Rueckspielsicherung)); got != "alt" {
		t.Errorf("beiseite liegt %q, erwartet war %q", got, "alt")
	}

	if zurueck == nil {
		t.Fatal("es gibt keinen Rückweg")
	}
}

func TestBestandErsetzenStelltDenAltenStandWiederHer(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	bestand := filepath.Join(ort, "bestand.db")
	sicherung := filepath.Join(ort, "sicherung.db")

	legeAn(t, bestand, "alt")
	legeAn(t, bestand+"-wal", "altes Journal")
	legeAn(t, sicherung, "neu")

	zurueck, err := bestandErsetzen(sicherung, bestand)
	if err != nil {
		t.Fatalf("der Tausch schlug fehl: %v", err)
	}

	// Der Aufrufer merkt beim Öffnen, dass die Sicherung nichts taugt.
	if err := zurueck(); err != nil {
		t.Fatalf("der Rückweg schlug fehl: %v", err)
	}

	if got := inhalt(t, bestand); got != "alt" {
		t.Errorf("nach der Rücknahme steht %q am Ort des Bestands, erwartet war %q", got, "alt")
	}

	// **Kein halber Stand.** Weder die neue Datei noch ein Journal aus dem
	// abgebrochenen Versuch darf zurückbleiben.
	if _, err := os.Stat(bestand + "-wal"); !os.IsNotExist(err) {
		t.Error("das Journal des abgebrochenen Versuchs liegt noch da")
	}
}

func TestBestandErsetzenLaesstDenBestandBeiEinemFehlerStehen(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	bestand := filepath.Join(ort, "bestand.db")

	legeAn(t, bestand, "alt")

	// Eine Sicherung, die es nicht gibt — der Tausch darf nicht auf halbem
	// Weg stehen bleiben.
	if _, err := bestandErsetzen(filepath.Join(ort, "fehlt.db"), bestand); err == nil {
		t.Fatal("erwartet war ein Fehler, bekommen: nil")
	}

	if got := inhalt(t, bestand); got != "alt" {
		t.Errorf("der Bestand steht als %q da, erwartet war unverändert %q", got, "alt")
	}
}

func TestBestandErsetzenGehtAuchOhneVorhandenenBestand(t *testing.T) {
	t.Parallel()

	// **Der Fall gibt es wirklich:** wer den Bestand von Hand gelöscht hat und
	// eine Sicherung einspielt, hat keinen alten Stand, der beiseite müsste.
	ort := t.TempDir()
	bestand := filepath.Join(ort, "bestand.db")
	sicherung := filepath.Join(ort, "sicherung.db")

	legeAn(t, sicherung, "neu")

	zurueck, err := bestandErsetzen(sicherung, bestand)
	if err != nil {
		t.Fatalf("der Tausch schlug fehl: %v", err)
	}

	if got := inhalt(t, bestand); got != "neu" {
		t.Errorf("am Ort des Bestands steht %q, erwartet war %q", got, "neu")
	}

	if err := zurueck(); err != nil {
		t.Fatalf("der Rückweg schlug fehl: %v", err)
	}

	if got := inhalt(t, bestand); got != "«fehlt»" {
		t.Errorf("nach der Rücknahme steht %q da, erwartet war nichts", got)
	}
}

// **Die Fehlerwege des Tauschs** — jeder betreten, keiner behauptet (§3).
//
// Hergestellt werden sie über **Verzeichnisse**: ein nicht leeres Verzeichnis
// lässt sich weder entfernen noch überschreiben, und das gilt auf jeder
// Plattform gleich. `chmod` täte es auch — aber nicht als `root`, und ein Test,
// der unter einer Benutzerkennung grün und unter einer anderen rot ist, prüft
// die Umgebung.
func vollesVerzeichnis(t *testing.T, pfad string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(pfad, "inhalt"), 0o700); err != nil {
		t.Fatalf("%s anlegen: %v", pfad, err)
	}
}

func TestBestandErsetzenMeldetEinenFehlgeschlagenenTausch(t *testing.T) {
	t.Parallel()

	// Jeder Fall baut sich seinen Ort selbst und gibt Sicherung und Bestand.
	faelle := map[string]func(t *testing.T, ort string) (string, string){
		"der bisherige Bestand lässt sich nicht beiseitelegen": func(
			t *testing.T, ort string,
		) (string, string) {
			t.Helper()

			bestand, sicherung := filepath.Join(ort, "bestand.db"), filepath.Join(ort, "sicherung.db")
			legeAn(t, bestand, "alt")
			legeAn(t, sicherung, "neu")
			vollesVerzeichnis(t, filepath.Join(ort, Rueckspielsicherung))

			return sicherung, bestand
		},
		"das Journal lässt sich nicht entfernen": func(t *testing.T, ort string) (string, string) {
			t.Helper()

			bestand, sicherung := filepath.Join(ort, "bestand.db"), filepath.Join(ort, "sicherung.db")
			legeAn(t, bestand, "alt")
			legeAn(t, sicherung, "neu")
			vollesVerzeichnis(t, bestand+"-wal")

			return sicherung, bestand
		},
		"die eingehende Kopie ist gar nicht da": func(t *testing.T, ort string) (string, string) {
			t.Helper()

			// **Seit #964 bekommt `bestandErsetzen` eine Datei dieser
			// Anwendung** und nicht mehr den fremden Pfad. Fehlt sie, darf
			// der Bestand das nicht kosten — die Wache steht vor jeder
			// Bewegung.
			return filepath.Join(ort, "gibt-es-nicht.db"), filepath.Join(ort, "bestand.db")
		},
		"die eingehende Kopie lässt sich nicht einsetzen": func(
			t *testing.T, ort string,
		) (string, string) {
			t.Helper()

			// **Ein gesperrtes Verzeichnis, in dem die Kopie schon liegt.**
			//
			// Es liegt **kein** Bestand darin: beiseitezulegen ist nichts,
			// das Journal ist nicht da — und erst das Umbenennen scheitert,
			// weil aus einem Verzeichnis ohne Schreibrecht nichts heraus- und
			// nichts hineingeht.
			gesperrt := filepath.Join(ort, "gesperrt")
			if err := os.Mkdir(gesperrt, 0o700); err != nil {
				t.Fatalf("das gesperrte Verzeichnis anlegen: %v", err)
			}

			bestand := filepath.Join(gesperrt, "bestand.db")

			eingehend := bestand + Eingehend
			legeAn(t, eingehend, "neu")

			// #nosec G302 -- ein Verzeichnis braucht das Ausführrecht.
			if err := os.Chmod(gesperrt, 0o500); err != nil {
				t.Fatalf("das Verzeichnis sperren: %v", err)
			}

			t.Cleanup(func() {
				// #nosec G302 -- ein Verzeichnis braucht das Ausführrecht.
				_ = os.Chmod(gesperrt, 0o700) //nolint:errcheck // nur zum Aufräumen
			})

			return eingehend, bestand
		},
	}

	for name, herstellen := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			sicherung, bestand := herstellen(t, t.TempDir())

			if _, err := bestandErsetzen(sicherung, bestand); err == nil {
				t.Fatal("erwartet war ein Fehler, bekommen: nil")
			}
		})
	}
}

// **Was der Rückweg meldet — und was er absichtlich nicht mehr tut.**
//
// Bis zum Zweitblick über Phase S räumte er auch die Journaldateien weg und
// hatte dafür einen eigenen Fehlerweg. Der ist fort, und zwar bewusst: der
// Grund steht an [zuruecklegen].
func TestZuruecklegenMeldetSeineEigenenFehler(t *testing.T) {
	t.Parallel()

	t.Run("der beiseitegelegte Stand ist fort", func(t *testing.T) {
		t.Parallel()

		ort := t.TempDir()
		bestand := filepath.Join(ort, "bestand.db")

		legeAn(t, bestand, "alt")
		legeAn(t, filepath.Join(ort, "sicherung.db"), "neu")

		zurueck, err := bestandErsetzen(filepath.Join(ort, "sicherung.db"), bestand)
		if err != nil {
			t.Fatalf("tauschen: %v", err)
		}

		if fehler := zurueck(); fehler != nil {
			t.Fatalf("der erste Rückweg schlug fehl: %v", fehler)
		}

		// **Der zweite Rückweg findet nichts mehr vor.** Das ist kein
		// erfundener Fall: er entsteht, wenn ein Aufrufer ihn zweimal geht.
		if fehler := zurueck(); fehler == nil {
			t.Fatal("der zweite Rückweg meldete Erfolg, obwohl nichts mehr da war")
		}
	})

	t.Run("die zurückgespielte Datei lässt sich nicht entfernen", func(t *testing.T) {
		t.Parallel()

		ort := t.TempDir()
		bestand := filepath.Join(ort, "bestand.db")

		legeAn(t, bestand, "alt")
		legeAn(t, filepath.Join(ort, "sicherung.db"), "neu")

		zurueck, err := bestandErsetzen(filepath.Join(ort, "sicherung.db"), bestand)
		if err != nil {
			t.Fatalf("tauschen: %v", err)
		}

		if fehler := os.Remove(bestand); fehler != nil {
			t.Fatalf("aufräumen: %v", fehler)
		}

		vollesVerzeichnis(t, bestand)

		if fehler := zurueck(); fehler == nil {
			t.Fatal("ein Verzeichnis am Ort des Bestands wurde übergangen")
		}
	})
}

// **Ein gescheiterter Tausch lässt den Bestand da, wo er war** — auch wenn er
// erst nach dem Umbenennen scheitert.
//
// Angestrichen vom Zweitblick über Phase S: `beiseiteLegen` benannte um und
// räumte danach die Journaldateien weg. Scheiterte das Wegräumen, kam ein
// Fehler zurück und **kein Rückweg** — die Hauptdatei lag unter dem Namen der
// Rückspielsicherung, und der Aufrufer rief anschließend `Starten`, das eine
// **neue leere** Datenbank anlegte.
//
// Der vorhandene Fehlerwegtest verlangte nur „irgendeinen Fehler". Genau das
// ist der Wächter, der die halbe Zusage hält (§10).
func TestNachEinemGescheitertenTauschStehtDerBestandNochDa(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	bestand := filepath.Join(ort, "bestand.db")
	sicherung := filepath.Join(ort, "sicherung.db")

	legeAn(t, bestand, "alt")
	legeAn(t, sicherung, "neu")

	// Ein Journal, das sich nicht entfernen lässt — der Tausch scheitert
	// **nach** dem Umbenennen.
	vollesVerzeichnis(t, bestand+"-wal")

	if _, err := bestandErsetzen(sicherung, bestand); err == nil {
		t.Fatal("erwartet war ein Fehler, bekommen: nil")
	}

	// **Das ist die Zusage**: der alte Stand liegt wieder an seinem Platz.
	if got := inhalt(t, bestand); got != "alt" {
		t.Errorf("am Ort des Bestands steht %q, erwartet war unverändert %q", got, "alt")
	}
}

// **Die eingehende Kopie meldet ihre Fehlerwege** (#964).
//
// Sie ist die Wache vor dem Tausch: was hier scheitert, hat noch nichts
// bewegt. Bis zum 2026-08-25 lagen diese Fälle in `bestandErsetzen` — dort
// wurde kopiert, und heute wird umbenannt.
