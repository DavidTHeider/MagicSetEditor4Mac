package storage_test

// **Die Gegenprobe zur Integritätsabfrage** — Hinweis 10 des Nachlaufs zum
// Zweitblick über Phase 5.2 (#1435).
//
// # Warum sie hier steht
//
// Die Stelle `integritaet` fragt SQLite, was es über seine eigene Datei sagt.
// Geprüft wurde sie bis hierher **nur an einem heilen Bestand**: mit einer
// mutierten Abfrage (`WHERE 0`) blieb die ganze Storage-Suite grün. Ein
// Wächter, der nur „ok" sieht, belegt nicht, dass er etwas anderes sähe.
//
// # Zwei Schäden, zwei Wege — und beide gehören dazu
//
// **Ein leichter Schaden kommt als Zeile**: verwaiste Seiten, ein fehlender
// Indexeintrag. Dafür steht das `WHERE integrity_check <> 'ok'` in der
// Abfrage.
//
// **Ein schwerer lässt die Abfrage scheitern** — `database disk image is
// malformed`. Dafür steht `klageIstBefund` daneben; ohne es endete die ganze
// Analyse mit einer rohen SQLite-Meldung (§8), und der Befund, für den es
// diese Stelle gibt, entstand nie.
//
// **Auf der Platte und nicht im Arbeitsspeicher**: `pragma_integrity_check`
// liest Seiten, und im Speicher gibt es keine, die sich verbiegen ließen.

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEinKaputterBestandWirdGemeldet(t *testing.T) {
	for name, verderben := range map[string]func(*testing.T, string){
		"eine verwaiste Seite": seitenVerwaisen,
		"eine verbogene Seite": verbiegen,
	} {
		t.Run(name, func(t *testing.T) {
			db := geoeffnet(t, kaputterOrt(t, verderben))

			befunde, fehler := storage.Pflegeanalyse(probekontext(t), db, pflegezeit)
			if fehler != nil {
				t.Fatalf("die Analyse: %v", fehler)
			}

			befund, gefunden := befundZu(befunde, "integritaet")
			if !gefunden {
				t.Fatalf("kein Befund über die Integrität: %+v", befunde)
			}

			// **Nur informieren** — „echte Korruption nicht durch kosmetisches
			// Löschen verdecken" (#1432). Ein Knopf, der das wegräumte, räumte
			// den Beleg weg.
			if befund.Klasse != domain.PflegeNurInformieren {
				t.Errorf("die Klasse ist %s", befund.Klasse)
			}

			if len(befund.Beispiele) == 0 ||
				!strings.Contains(befund.Beispiele[0], "integritaet:") {
				t.Errorf("die Beispiele sind %v", befund.Beispiele)
			}

			pruefeSchwere(t, name, befund.Beispiele)
		})
	}
}

// pruefeSchwere hält fest, **welchen** der beiden Wege ein Fall genommen hat.
//
// # Warum das hier steht
//
// **Weil der schwere Fall still zum leichten werden kann.** Bis zum
// 2026-09-12 verbog `verbiegen` 200 Byte in der Mitte der Datei; mit einer
// neuen Wanderung (0057) lag dort eine andere Seite, `integrity_check` gab
// den Schaden als **Zeile** zurück — und der Zweig, für den es diesen Lauf
// gibt (`klageIstBefund`), blieb unbetreten. Der Lauf blieb dabei **grün**;
// angestrichen hat es die Abdeckung.
//
// **Jetzt sagt der Lauf, was er meint**: ein verbogener Bestand **muss** die
// Abfrage scheitern lassen, ein verwaister **darf** es nicht.
func pruefeSchwere(t *testing.T, name string, beispiele []string) {
	t.Helper()

	// **Die Meldung steht im Beispiel**, wenn die Abfrage scheitert —
	// `klageIstBefund` setzt sie dort ein. **Seit F05 eingeordnet** (#2344):
	// aus SQLites `malformed` wird „die Datei ist beschädigt", und der rohe
	// Text bleibt im Grund.
	gescheitert := strings.Contains(beispiele[0], domain.ErrBeschaedigt.Error()) ||
		strings.Contains(beispiele[0], "SQL logic error")

	if name == "eine verbogene Seite" && !gescheitert {
		t.Errorf("der schwere Fall kam als Zeile durch: %v — dann prüft "+
			"dieser Lauf den Zweig nicht mehr, für den es ihn gibt", beispiele)
	}

	if name == "eine verwaiste Seite" && gescheitert {
		t.Errorf("der leichte Fall ließ die Abfrage scheitern: %v", beispiele)
	}
}

// **Und ein heiler Bestand meldet nichts** — die Gegenprobe zur Gegenprobe.
//
// **Ohne sie sagte der Lauf darüber nur, dass die Abfrage etwas liefert** —
// nicht, dass sie zwischen heil und kaputt unterscheidet (§10).
func TestEinHeilerBestandMeldetKeineKorruption(t *testing.T) {
	befunde, fehler := storage.Pflegeanalyse(
		probekontext(t), frischeDatenbank(t), pflegezeit)
	if fehler != nil {
		t.Fatalf("die Analyse: %v", fehler)
	}

	if befund, gefunden := befundZu(befunde, "integritaet"); gefunden {
		t.Errorf("ein heiler Bestand meldet %+v", befund)
	}
}

// befundZu sucht den Befund einer Art.
func befundZu(befunde []domain.Pflegebefund, art string) (domain.Pflegebefund, bool) {
	for _, befund := range befunde {
		if befund.Art == art {
			return befund, true
		}
	}

	return domain.Pflegebefund{}, false
}

// kaputterOrt legt einen Bestand an und macht ihn auf die genannte Art kaputt.
//
// **Mit Ballast**: eine frisch migrierte Datenbank ist zu klein, als dass eine
// Verbiegung in ihrer Mitte eine Baumseite träfe.
func kaputterOrt(t *testing.T, verderben func(*testing.T, string)) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "bestand.db")
	db := geoeffnet(t, ort)

	if err := storage.Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	for _, anweisung := range []string{
		`CREATE TABLE ballast (x INTEGER, y TEXT)`,
		`CREATE INDEX ballast_y ON ballast(y)`,
		`INSERT INTO ballast
		      SELECT value, 'Text Nummer ' || value FROM generate_series(1, 4000)`,
	} {
		if _, err := db.ExecContext(probekontext(t), anweisung); err != nil {
			t.Fatalf("%s: %v", anweisung, err)
		}
	}

	verderben(t, ort)

	return ort
}

// geoeffnet öffnet einen Bestand und schließt ihn am Ende des Laufs.
func geoeffnet(t *testing.T, ort string) *sql.DB {
	t.Helper()

	db, fehler := storage.Oeffnen(probekontext(t), ort)
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	return db
}
