package app

// **Die Marke einer vorhandenen Sicherung** —
// #967.
//
// **Paketintern**, aus demselben Grund wie `sicherungvormigration_test.go`:
// die Entscheidung wird im Composition Root verdrahtet und von niemandem
// sonst gerufen.

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// marke gibt eine feste Marke — für den Fall, dass beide Seiten gleich sind.
func marke(wert string) func(context.Context, string) (string, error) {
	return func(context.Context, string) (string, error) { return wert, nil }
}

// **Eine vorhandene Sicherung vom heutigen Stand wird genommen.**
//
// Das ist der Fall, für den es die Wiederverwendung gibt: ein früherer Lauf
// hat gesichert, die Migration ist gescheitert, und weil sie eine Transaktion
// ist, steht der Bestand unverändert da.
func TestEineSicherungVomHeutigenStandWirdGenommen(t *testing.T) {
	var gesichert []string

	sichern := func(_ context.Context, ziel string) error {
		gesichert = append(gesichert, ziel)

		return konflikt()
	}

	fehler := sicherungVorMigration(sichern, taugt, marke("gleich"),
		marke("gleich"), "/ort", "0.2.0")(probekontext(t))
	if fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if len(gesichert) != 1 {
		t.Errorf("gesichert wurde nach %v — eine zweite Kopie war nicht nötig", gesichert)
	}
}

// **Eine Sicherung von einem anderen Stand wird nicht genommen.**
//
// # Der Fall, den der Nachlauf benannt hat
//
//  1. Upgrade, die Migration bricht ab. Die Sicherung liegt da.
//  2. Wer weiterarbeiten will, startet die **alte** Fassung — sie öffnet den
//     Bestand, denn die Migration ist zurückgerollt — und ändert Karten.
//  3. Zweiter Upgradeversuch.
//
// Bis zum 2026-08-25 wurde die vorhandene Sicherung in Schritt 3 angenommen,
// obwohl sie die Arbeit aus Schritt 2 **nicht** enthält. Scheitert nun die
// Migration erneut und jemand spielt zurück, ist die Arbeit fort.
//
// **Überschrieben wird sie trotzdem nicht** — eine Sicherung überschreibt
// nichts. Es entsteht eine zweite, und ihr Name trägt die Marke.
func TestEineSicherungVonGesternWirdNichtGenommen(t *testing.T) {
	var gesichert []string

	sichern := func(_ context.Context, ziel string) error {
		gesichert = append(gesichert, ziel)

		// **Nur der erste Name liegt schon da.** Der zweite trägt die Marke
		// und ist frei.
		if strings.HasSuffix(ziel, "bestand-vor-0.2.0.db") {
			return konflikt()
		}

		return nil
	}

	fehler := sicherungVorMigration(sichern, taugt, marke("heute"),
		marke("gestern"), "/ort", "0.2.0")(probekontext(t))
	if fehler != nil {
		t.Fatalf("Fehler: %v", fehler)
	}

	if len(gesichert) != 2 {
		t.Fatalf("gesichert wurde nach %v — erwartet war eine zweite Kopie", gesichert)
	}

	// **Der zweite Name trägt die Marke des heutigen Standes.** Ohne sie
	// entstünde bei jedem Versuch derselbe zweite Name, und der zweite
	// Konflikt wäre wieder eine Wiederverwendung.
	if !strings.Contains(gesichert[1], "heute") {
		t.Errorf("der zweite Name ist %q und nennt den Stand nicht", gesichert[1])
	}
}

// **Scheitert die Marke, hält die Migration auf.**
//
// Ohne Marke lässt sich nicht sagen, ob die vorhandene Sicherung die richtige
// ist — und eine Migration ohne belegte Sicherung ist genau der Fall, für den
// es sie gibt.
func TestEineUnlesbareMarkeHaeltAuf(t *testing.T) {
	kaputt := errors.New("die Datei ließ sich nicht lesen")

	scheitert := func(context.Context, string) (string, error) { return "", kaputt }

	for name, fall := range map[string][2]func(context.Context, string) (string, error){
		"die eigene":    {scheitert, marke("da")},
		"die der Kopie": {marke("da"), scheitert},
	} {
		t.Run(name, func(t *testing.T) {
			fehler := sicherungVorMigration(liegtDa, taugt, fall[0], fall[1],
				"/ort", "0.2.0")(probekontext(t))
			if !errors.Is(fehler, kaputt) {
				t.Errorf("Fehler %v, erwartet war die unlesbare Datei", fehler)
			}
		})
	}
}

// **Die verdrahtete Naht ruft wirklich `storage`.**
//
// # Was dieser Lauf hinzufügt
//
// Die Läufe darüber prüfen die **Entscheidung** gegen gestellte Marken. Sie
// sagen nichts darüber, ob der Composition Root die richtige Funktion
// verdrahtet — eine Naht, die immer dieselbe Zeichenkette gäbe, ließe sie
// alle grün.
//
// **Hier geht es durch die echte Datenbank.** Zwei Läufe hintereinander mit
// einer Änderung dazwischen: der erste sichert, der zweite trifft die
// vorhandene Datei und muss eine **zweite** anlegen, weil der Stand ein
// anderer ist.
func TestDieVerdrahteteNahtErkenntEinenAnderenStand(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()
	pfad := filepath.Join(ort, "bestand.db")

	db, err := storage.Oeffnen(probekontext(t), "file:"+pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	if migrationsfehler := storage.Migriere(probekontext(t), db, nil); migrationsfehler != nil {
		t.Fatalf("migrieren: %v", migrationsfehler)
	}

	vorbereiten := SicherungVorMigration(db, ort, "0.2.0")

	if fehler := vorbereiten(probekontext(t)); fehler != nil {
		t.Fatalf("die erste Sicherung: %v", fehler)
	}

	// **Zwischendurch wird gearbeitet** — genau der Fall aus #967.
	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO sets (name, spiel, geoeffnet, profil_id)
		 VALUES ('Neu', 'magic', '2026-08-25T00:00:00Z', 1)`,
	); fehler != nil {
		t.Fatalf("ein Set anlegen: %v", fehler)
	}

	if fehler := vorbereiten(probekontext(t)); fehler != nil {
		t.Fatalf("die zweite Sicherung: %v", fehler)
	}

	gefunden, lesefehler := filepath.Glob(filepath.Join(ort, "bestand-vor-0.2.0*.db"))
	if lesefehler != nil {
		t.Fatalf("das Verzeichnis lesen: %v", lesefehler)
	}

	if len(gefunden) != 2 {
		t.Errorf("es liegen %d Sicherungen da: %v — erwartet waren zwei", len(gefunden), gefunden)
	}
}

// **Und ein dritter Lauf ohne Änderung legt nichts Neues an.**
//
// Ohne diesen Lauf wäre die Gegenprobe halb: eine Naht, die **immer** eine
// zweite Datei schreibt, bestünde den Lauf darüber ebenso.
func TestDieVerdrahteteNahtErkenntDenselbenStand(t *testing.T) {
	t.Parallel()

	ort := t.TempDir()

	db, err := storage.Oeffnen(probekontext(t), "file:"+filepath.Join(ort, "bestand.db"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	if migrationsfehler := storage.Migriere(probekontext(t), db, nil); migrationsfehler != nil {
		t.Fatalf("migrieren: %v", migrationsfehler)
	}

	vorbereiten := SicherungVorMigration(db, ort, "0.2.0")

	for lauf := range 3 {
		if fehler := vorbereiten(probekontext(t)); fehler != nil {
			t.Fatalf("Lauf %d: %v", lauf+1, fehler)
		}
	}

	gefunden, lesefehler := filepath.Glob(filepath.Join(ort, "bestand-vor-0.2.0*.db"))
	if lesefehler != nil {
		t.Fatalf("das Verzeichnis lesen: %v", lesefehler)
	}

	if len(gefunden) != 1 {
		t.Errorf("es liegen %d Sicherungen da: %v — erwartet war eine", len(gefunden), gefunden)
	}
}

// **Scheitert die zweite Sicherung, hält sie auf.**
//
// Eine Migration ohne Sicherung des heutigen Standes ist genau der Fall, für
// den es sie gibt — und die vorhandene trägt einen anderen.
func TestEineGescheiterteZweiteSicherungHaeltAuf(t *testing.T) {
	t.Parallel()

	voll := errors.New("die Platte ist voll")

	sichern := func(_ context.Context, ziel string) error {
		if strings.HasSuffix(ziel, "bestand-vor-0.2.0.db") {
			return konflikt()
		}

		return voll
	}

	fehler := sicherungVorMigration(sichern, taugt, marke("heute"),
		marke("gestern"), "/ort", "0.2.0")(probekontext(t))

	if !errors.Is(fehler, voll) {
		t.Fatalf("Fehler %v, erwartet war die volle Platte", fehler)
	}

	if !strings.Contains(fehler.Error(), "anderen Stand") {
		t.Errorf("die Meldung sagt nicht, warum eine zweite nötig war: %v", fehler)
	}
}
