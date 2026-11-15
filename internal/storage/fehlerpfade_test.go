// Die Fehlerpfade des Migrators.
//
// **Paketintern**, und zwar aus einem Grund: mit der eingebetteten Quelle sind
// diese Pfade nicht erreichbar, weil sie richtig ist. Ein Fehlerpfad, den kein
// Test je betritt, ist eine Behauptung darüber, was im Ernstfall geschieht —
// und im Ernstfall stellt sie sich als falsch heraus, weil dort niemand
// nachgesehen hat.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"testing/fstest"
	"time"
)

func TestZeitstempelSchreibtDreiNachkommastellen(t *testing.T) {
	zeitpunkt := time.Date(2026, 8, 8, 10, 0, 0, 123456789, time.UTC)

	if gestempelt := Zeitstempel(zeitpunkt); gestempelt != "2026-08-08T10:00:00.123Z" {
		t.Errorf("Zeitstempel %q, erwartet war %q", gestempelt, "2026-08-08T10:00:00.123Z")
	}
}

// Ein Zeitpunkt in einer anderen Zone wird umgerechnet, nicht abgeschnitten.
func TestZeitstempelRechnetInUTCUm(t *testing.T) {
	berlin := time.FixedZone("CEST", 2*60*60)
	zeitpunkt := time.Date(2026, 8, 8, 0, 30, 0, 0, berlin)

	if gestempelt := Zeitstempel(zeitpunkt); gestempelt != "2026-08-07T22:30:00.000Z" {
		t.Errorf("Zeitstempel %q, erwartet war %q", gestempelt, "2026-08-07T22:30:00.000Z")
	}
}

func TestOeffnenMeldetEineUnbrauchbareQuelle(t *testing.T) {
	// Ein Verzeichnis lässt sich nicht als Datenbank öffnen.
	if _, err := Oeffnen(probekontext(t), t.TempDir()); err == nil {
		t.Error("ein Verzeichnis wurde als Datenbank angenommen")
	}
}

func TestMigriereMeldetEineGeschlosseneDatenbank(t *testing.T) {
	db := geschlosseneDatenbank(t)

	if err := Migriere(probekontext(t), db, nil); err == nil {
		t.Error("die Migration meldete auf einer geschlossenen Datenbank keinen Fehler")
	}
}

func TestMigriereMeldetEineUnlesbareQuelle(t *testing.T) {
	db := offeneDatenbank(t)

	// Eine Quelle ohne das Verzeichnis `migrationen`.
	if err := migriereAus(probekontext(t), db, fstest.MapFS{}, nil); err == nil {
		t.Error("eine Quelle ohne Migrationen wurde angenommen")
	}
}

func TestMigriereMeldetFehlerhaftesSQL(t *testing.T) {
	db := offeneDatenbank(t)

	quelle := fstest.MapFS{
		"migrationen/0001_kaputt.sql": {Data: []byte("DIES IST KEIN SQL")},
	}

	err := migriereAus(probekontext(t), db, quelle, nil)
	if err == nil {
		t.Fatal("fehlerhaftes SQL wurde angenommen")
	}

	// Und die Wirkung darf nicht halb eingetreten sein: nichts ist verbucht.
	verbucht, fehler := AngewendeteMigrationen(probekontext(t), db)
	if fehler != nil {
		t.Fatalf("Buchführung nicht lesbar: %v", fehler)
	}
	if len(verbucht) != 0 {
		t.Errorf("nach dem Fehlschlag ist %v verbucht", verbucht)
	}
}

func TestAngewendeteMigrationenMeldetEineFehlendeBuchfuehrung(t *testing.T) {
	db := offeneDatenbank(t)

	if _, err := AngewendeteMigrationen(probekontext(t), db); err == nil {
		t.Error("die Buchführung fehlt und wurde trotzdem gelesen")
	}
}

func TestPruefeBestandMeldetEinenFehlendenBestand(t *testing.T) {
	db := offeneDatenbank(t)

	if err := PruefeBestand(probekontext(t), db); err == nil {
		t.Error("ein fehlender Bestand kam durch")
	}
}

func TestPruefeBestandMeldetEineFehlendeZeile(t *testing.T) {
	db := offeneDatenbank(t)

	if _, err := db.ExecContext(probekontext(t),
		`CREATE TABLE bestand (schluessel TEXT PRIMARY KEY, wert TEXT NOT NULL) STRICT`,
	); err != nil {
		t.Fatalf("Tabelle nicht anlegbar: %v", err)
	}

	err := PruefeBestand(probekontext(t), db)
	if !errors.Is(err, ErrBestand) {
		t.Errorf("Fehler %v, erwartet war ErrBestand", err)
	}
}

// ---------------------------------------------------------------- Vorrichtung

func offeneDatenbank(t *testing.T) *sql.DB {
	t.Helper()

	db, err := Oeffnen(probekontext(t), "file:"+t.TempDir()+"/probe.db")
	if err != nil {
		t.Fatalf("Datenbank nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	return db
}

func geschlosseneDatenbank(t *testing.T) *sql.DB {
	t.Helper()

	db, err := FuerLaeufe("file:" + t.TempDir() + "/zu.db")
	if err != nil {
		t.Fatalf("Datenbank nicht öffnenbar: %v", err)
	}
	if fehler := db.Close(); fehler != nil {
		t.Fatalf("Datenbank nicht schließbar: %v", fehler)
	}

	return db
}

// einfacheQuelle ist eine Migrationsquelle mit einer harmlosen Anweisung.
func einfacheQuelle() fstest.MapFS {
	return fstest.MapFS{
		"migrationen/0001_probe.sql": {Data: []byte("SELECT 1")},
	}
}

// **`BeginTx` steht seit #938
// an der Folge**, nicht an der einzelnen Migration.
func TestDieMigrationsfolgeMeldetEineGeschlosseneDatenbank(t *testing.T) {
	if err := anwendenAlle(probekontext(t), geschlosseneDatenbank(t), einfacheQuelle(),
		[]string{"0001_probe.sql"}, nil, nil, nil, nil); err == nil {
		t.Error("BeginTx auf einer geschlossenen Datenbank meldete keinen Fehler")
	}
}

func TestAnwendenMeldetEineFehlendeBuchfuehrung(t *testing.T) {
	// Die Migration selbst glückt, das Verbuchen scheitert: die Tabelle
	// `migrationen` gibt es danach nicht mehr. **Sie nimmt sie selbst weg** —
	// gelesen wird die Buchführung seit Vermerk 0189 unter der Sperre und vor
	// der ersten Wanderung, und ohne Tabelle scheiterte schon das.
	db := offeneDatenbank(t)
	if err := buchfuehrungAnlegen(probekontext(t), db); err != nil {
		t.Fatal(err)
	}

	quelle := fstest.MapFS{"migrationen/0001_probe.sql": {Data: []byte("DROP TABLE migrationen")}}

	if err := anwendenAlle(probekontext(t), db, quelle,
		[]string{"0001_probe.sql"}, nil, nil, nil, nil); err == nil {
		t.Error("das Verbuchen ohne Buchführungstabelle meldete keinen Fehler")
	}
}

// **Und ohne Buchführung scheitert schon das Lesen unter der Sperre.**
func TestAnwendenLiestDieBuchfuehrungUnterDerSperre(t *testing.T) {
	if err := anwendenAlle(probekontext(t), offeneDatenbank(t), einfacheQuelle(),
		[]string{"0001_probe.sql"}, nil, nil, nil, nil); err == nil {
		t.Error("ohne Buchführungstabelle wurde gewandert")
	}
}

// Eine Migration, die zwischen Auflisten und Lesen verschwindet.
//
// Über `anwenden` unmittelbar, weil `migriereAus` sie gar nicht erst anböte:
// seit der Namensprüfung fällt ein Verzeichnis schon dort durch. Der Lesefehler
// bleibt trotzdem möglich — eine eingebettete Quelle ist es nicht immer, und im
// Betrieb entscheidet dieser Zweig, ob die Anwendung es meldet.
func TestAnwendenMeldetEineUnlesbareMigration(t *testing.T) {
	db := offeneDatenbank(t)

	tx, err := db.BeginTx(probekontext(t), nil)
	if err != nil {
		t.Fatalf("Transaktion: %v", err)
	}

	defer func() { _ = tx.Rollback() }()

	if fehler := anwenden(probekontext(t), tx, fstest.MapFS{},
		"0001_nicht_da.sql"); fehler == nil {
		t.Error("eine fehlende Migrationsdatei wurde angenommen")
	}
}

func TestMigriereMeldetEineUnbrauchbareBuchfuehrung(t *testing.T) {
	db := offeneDatenbank(t)

	// Eine Tabelle `migrationen` in fremder Form: CREATE TABLE IF NOT EXISTS
	// lässt sie stehen, und das SELECT darauf scheitert.
	if _, err := db.ExecContext(probekontext(t),
		`CREATE TABLE migrationen (fremd TEXT)`); err != nil {
		t.Fatalf("Tabelle nicht anlegbar: %v", err)
	}

	if err := Migriere(probekontext(t), db, nil); err == nil {
		t.Error("eine Buchführung in fremder Form wurde angenommen")
	}
}

func TestAngewendeteMigrationenMeldetEinenUnlesbarenNamen(t *testing.T) {
	db := offeneDatenbank(t)

	// Ohne STRICT und ohne NOT NULL: SQLite lässt NULL in einer
	// TEXT-PRIMARY-KEY-Spalte zu, und das Scannen nach string scheitert daran.
	if _, err := db.ExecContext(probekontext(t),
		`CREATE TABLE migrationen (name TEXT PRIMARY KEY, angewendet TEXT, pruefsumme TEXT)`,
	); err != nil {
		t.Fatalf("Tabelle nicht anlegbar: %v", err)
	}
	if _, err := db.ExecContext(probekontext(t),
		`INSERT INTO migrationen VALUES (NULL, 'x', '')`); err != nil {
		t.Fatalf("Zeile nicht einfügbar: %v", err)
	}

	if _, err := AngewendeteMigrationen(probekontext(t), db); err == nil {
		t.Error("ein NULL-Name wurde als Zeichenkette gelesen")
	}
}

// **Ein Öffnen, das scheitert, wird gemeldet.**
//
// # Warum der Name der Prüfung sich geändert hat
//
// **Hier stand „ein unbekannter Treibername"** — bis Werkstück 14 von
// Phase 4.12 öffnete `Oeffnen` über `sql.Open`, und das scheiterte **nur** an
// einem Treibernamen, den niemand angemeldet hat. Seit dem Treiberwechsel
// öffnet [oeffner], und der ist eine Naht: der Lauf tauscht sie und bekommt
// denselben Weg zu prüfen, ohne einen Namen erfinden zu müssen.
//
// **Was gleich bleibt, ist der Zweck**: der Fehlerweg unter dem Öffnen ist
// sonst unerreichbar — und ein Fehlerpfad, den kein Test betritt, ist eine
// Behauptung darüber, was im Ernstfall geschieht.
func TestOeffnenMeldetEinScheiterndesOeffnen(t *testing.T) {
	ohneOeffner(t)

	if _, err := Oeffnen(probekontext(t), ImArbeitsspeicher); err == nil {
		t.Error("ein gescheitertes Öffnen wurde angenommen")
	}
}

func TestMigriereWeistEinenFalschenMigrationsnamenAb(t *testing.T) {
	quelle := fstest.MapFS{
		"migrationen/9_ohne_fuehrende_nullen.sql": {Data: []byte("SELECT 1")},
	}

	err := migriereAus(probekontext(t), offeneDatenbank(t), quelle, nil)
	if !errors.Is(err, ErrMigrationen) {
		t.Errorf("Fehler %v, erwartet war ErrMigrationen", err)
	}
}

// Was verbucht ist, muss es noch geben.
//
// So sieht eine Datenbank aus, die mit einer neueren Fassung entstanden ist:
// die Buchführung kennt eine Migration, die es hier nicht gibt. Weiter zu
// migrieren hieße, auf einem Schema aufzubauen, das man nicht kennt.
func TestMigriereWeistEineVerschwundeneMigrationAb(t *testing.T) {
	db := offeneDatenbank(t)

	for _, anweisung := range []string{
		`CREATE TABLE migrationen (name TEXT PRIMARY KEY, angewendet TEXT NOT NULL) STRICT`,
		`INSERT INTO migrationen VALUES ('0099_aus_der_zukunft.sql', '2026-08-08T10:00:00.000Z')`,
	} {
		if _, err := db.ExecContext(probekontext(t), anweisung); err != nil {
			t.Fatalf("vorbereiten: %v", err)
		}
	}

	err := Migriere(probekontext(t), db, nil)
	if !errors.Is(err, ErrMigrationen) {
		t.Errorf("Fehler %v, erwartet war ErrMigrationen", err)
	}
}

// **Die Folge ist eine Transaktion, nicht jede Migration für sich** —
// #938.
//
// # Der Anlass
//
// `anwenden` legte jede Migration in eine eigene Transaktion. Scheiterte die
// dritte von fünf, waren die ersten beiden festgeschrieben — der Bestand stand
// auf einem Stand, den es als Fassung **nie gab**.
// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md)
// hat es am 2026-08-20 angestrichen und zur Bedingung gemacht, unter der der
// Satz „rückwärts, nie" vertretbar ist.
//
// # Warum das geht
//
// Keine der zwölf Migrationen enthält `PRAGMA` oder `VACUUM` — beides wäre in
// einer Transaktion wirkungslos oder verboten. SQLite führt DDL transaktional.
//
// # Was es kostet
//
// **Den Fortschritt.** Vorher blieb erhalten, was schon lief, und der nächste
// Start setzte dort auf; jetzt beginnt er von vorn. Das ist der bessere
// Handel: „von vorn" ist ein Zustand, den es gibt, „zur Hälfte" nicht.
func TestEinFehlschlagInDerMitteLaesstNichtsZurueck(t *testing.T) {
	db := offeneDatenbank(t)

	quelle := fstest.MapFS{
		"migrationen/0001_gut.sql":    {Data: []byte("CREATE TABLE eins (a TEXT);")},
		"migrationen/0002_gut.sql":    {Data: []byte("CREATE TABLE zwei (a TEXT);")},
		"migrationen/0003_kaputt.sql": {Data: []byte("DIES IST KEIN SQL")},
		"migrationen/0004_gut.sql":    {Data: []byte("CREATE TABLE vier (a TEXT);")},
	}

	if err := migriereAus(probekontext(t), db, quelle, nil); err == nil {
		t.Fatal("die kaputte Migration wurde angenommen")
	}

	verbucht, fehler := AngewendeteMigrationen(probekontext(t), db)
	if fehler != nil {
		t.Fatalf("Buchführung nicht lesbar: %v", fehler)
	}

	if len(verbucht) != 0 {
		t.Errorf("nach dem Fehlschlag ist %v verbucht — die Folge ist keine Transaktion", verbucht)
	}

	// **Und die Wirkung ist ebenfalls fort.** Eine Buchführung ohne Tabellen
	// wäre die halbe Zusage: was zählt, ist das Schema.
	for _, tabelle := range []string{"eins", "zwei"} {
		var da int
		if err := db.QueryRowContext(probekontext(t),
			`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = ?`,
			tabelle).Scan(&da); err != nil {
			t.Fatalf("Schema nicht lesbar: %v", err)
		}

		if da != 0 {
			t.Errorf("die Tabelle %q steht noch — die Rücknahme hat sie nicht erreicht", tabelle)
		}
	}
}

// **Weist die Bestandsprüfung ab, wird die ganze Folge zurückgenommen** —
// #938.
//
// Sie lief bis zum 2026-08-23 **nach** der Migration: wies sie ab, war die
// Migration bereits festgeschrieben, und der Bestand ließ sich weder benutzen
// noch mit der alten Fassung öffnen. Hier ist die Abweisung eine Rücknahme.
func TestEineAbweisendeBestandspruefungNimmtDieFolgeZurueck(t *testing.T) {
	db := offeneDatenbank(t)

	quelle := fstest.MapFS{
		"migrationen/0001_gut.sql": {Data: []byte("CREATE TABLE eins (a TEXT);")},
	}

	abweisen := func(context.Context, Zeilenquelle) error { return errGestolpert }

	if err := migriereAus(probekontext(t), db, quelle, abweisen); !errors.Is(err, errGestolpert) {
		t.Fatalf("Fehler %v, erwartet war die Abweisung", err)
	}

	verbucht, fehler := AngewendeteMigrationen(probekontext(t), db)
	if fehler != nil {
		t.Fatalf("Buchführung nicht lesbar: %v", fehler)
	}

	if len(verbucht) != 0 {
		t.Errorf("nach der Abweisung ist %v verbucht", verbucht)
	}
}
