// **Die Abnahme von Phase S**, gefahren statt behauptet
// (#885).
//
// Die drei Werkstücke der Phase sind je für sich belegt: die Sicherung vor der
// Migration (#939), die Folge in **einer** Transaktion (#938), der Weg zurück
// im Fenster (#940). Was keiner von ihnen prüft, ist die **Naht** — dass ein
// abbrechendes Upgrade wirklich ohne Handarbeit zurückzuführen ist.
//
// # Wie das Abbrechen hergestellt wird
//
// **Mit einer echten Wanderung, die scheitert** — `0098_kaputt.sql`, die die
// „neue Fassung" mitbringt: sie legt eine Spalte an und schreibt dann in eine
// Tabelle, die es nicht gibt. Bis zum 2026-10-12 strich der Lauf stattdessen
// eine Zeile aus der Buchführung; seit es nur die Grundlinie gibt
// (Vermerk 0189), ist ein solcher Bestand einer von vor 0.7.0 und wird
// abgewiesen, bevor überhaupt gewandert wird.
//
// # Was der Fall nicht nachstellt
//
// **Zwei Programmfassungen.** Hier läuft eine; die neue entsteht, indem die
// Naht `migriere` eine Wanderung mehr mitbringt, und die alte, indem sie
// zurückgestellt wird. Was sich damit zeigen lässt, ist der **Weg** —
// Sicherung, Abbruch ohne halben Stand, Rückweg —, nicht das Zusammenspiel
// zweier Binärdateien. Das
// bliebe auch mit mehr Aufwand ungeprüft, solange die Pipeline eine Maschine
// prüft und keine Plattform ([0007](../../docs/entscheidungen/0007-kein-entfernter-lauf.md)).
package app

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// kaputteFassung lässt die Anwendung für die Dauer eines Laufs wie eine neue
// Fassung wandern, deren erste Wanderung scheitert — und gibt deren Namen.
//
// **Nicht parallel**: die Naht ist eine Variable des Pakets. Der Lauf, der sie
// setzt, läuft deshalb für sich; `t.Cleanup` stellt sie zurück.
func kaputteFassung(t *testing.T) string {
	t.Helper()

	// **Die nächste freie Nummer, nicht fest `0098`** — so hieß die Probe,
	// bis es die erste echte Wanderung gab (M01, #2345).
	name := fmt.Sprintf("%04d_kaputt.sql", heuteVerbucht(t)+1)

	danach := fstest.MapFS{"migrationen/" + name: {Data: []byte(
		"ALTER TABLE settings ADD COLUMN probe TEXT;\nINSERT INTO gibt_es_nicht VALUES (1);")}}

	vorher := migriere
	migriere = func(ctx context.Context, db *sql.DB, vorbereiten storage.Vorbereitung) error {
		return storage.MigriereMit(ctx, db, danach, vorbereiten)
	}

	t.Cleanup(func() { migriere = vorher })

	return name
}

// alteFassung stellt die Naht zurück — wie jemand, der nach dem Abbruch die
// Fassung von vorher startet.
func alteFassung() { migriere = storage.Migriere }

// spracheImBestand liest die Einstellung unmittelbar aus der Datei.
//
// **An der Anwendung vorbei**, denn die läuft an dieser Stelle nicht.
func spracheImBestand(t *testing.T, pfad string) string {
	t.Helper()

	db, err := storage.FuerLaeufe("file:" + pfad + "?mode=ro")
	if err != nil {
		t.Fatalf("die Datenbank öffnen: %v", err)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur zum Aufräumen

	var wert string
	if fehler := db.QueryRowContext(probekontext(t),
		`SELECT wert FROM settings WHERE schluessel = 'sprache'`).Scan(&wert); fehler != nil {
		t.Fatalf("die Sprache lesen: %v", fehler)
	}

	return wert
}

func TestEinAbbrechendesUpgradeLaesstSichZurueckfuehren(t *testing.T) {
	ort := t.TempDir()
	bestand := filepath.Join(ort, "bestand.db")

	// ── 1. Der alte Stand: eine Anwendung, die läuft, mit einer Marke darin.
	alt := NeueApp(bestand)
	if err := alt.Starten(probekontext(t)); err != nil {
		t.Fatalf("starten: %v", err)
	}

	if antwort := alt.EinstellungSetzen(probekontext(t), "sprache", "de"); antwort.Fehler != nil {
		t.Fatalf("die Marke setzen: %+v", antwort.Fehler)
	}

	if err := alt.Beenden(); err != nil {
		t.Fatalf("beenden: %v", err)
	}

	// ── 2. Das Upgrade, das abbricht.
	kaputt := kaputteFassung(t)

	neu := NeueApp(bestand)

	fehler := neu.Starten(probekontext(t))
	if fehler == nil {
		t.Fatal("das Upgrade lief durch — es sollte an der fehlenden Tabelle scheitern")
	}

	if !strings.Contains(fehler.Error(), "migrieren") {
		t.Errorf("der Fehler nennt die Migration nicht: %v", fehler)
	}

	// ── 3. **Kein halber Stand** (#938): die Folge lief in einer Transaktion,
	// also ist die gescheiterte Wanderung unverbucht — und ihre Spalte nicht
	// etwa halb angelegt.
	if angewendet := verbuchte(t, bestand); slicesEnthaelt(angewendet, kaputt) {
		t.Error("die abgebrochene Migration steht als angewendet in der Buchführung")
	}

	// ── 4. **Die Sicherung vor der Migration ist da** (#939) — selbsttätig
	// geschrieben, weil etwas ausstand.
	automatisch := filepath.Join(ort, Sicherungsname(Fassung))
	if _, err := os.Stat(automatisch); err != nil {
		t.Errorf("die Sicherung vor der Migration fehlt: %v", err)
	}

	// ── 5. **Und der Weg dorthin ist erreichbar** (#962).
	//
	// Bis zum 2026-08-25 kehrte `main.lauf` bei einem Fehler aus `Starten`
	// **vor** `wails.Run` zurück: kein Fenster, keine Bindung, kein Knopf.
	// Dieser Lauf fuhr den Rückweg über das `App`-Objekt unmittelbar und
	// belegte damit, dass das Objekt wiederverwendbar ist — nicht, dass jemand
	// den Weg gehen kann.
	//
	// Jetzt geht er über die **Fassade**, so wie das Fenster: erst fragen, was
	// beim Start geschah, dann zurückspielen.
	bindungen := NeueBindungen(neu)

	befund := bindungen.Startbefund(t.Context())
	if befund.Offen {
		t.Fatal("der Startbefund sagt offen, obwohl der Start scheiterte")
	}

	if !strings.Contains(befund.Grund, "migrieren") {
		t.Errorf("der Grund nennt die Migration nicht: %q", befund.Grund)
	}

	// ── 6. **Was die selbsttätige Sicherung kann — und was nicht.**
	//
	// Der erste Anlauf prüfte, dass sie dasteht, und spielte danach die von
	// Hand erzeugte ein. Über den Weg zurück durch die automatische Datei
	// sagte er nichts — angestrichen vom Zweitblick über Phase S.
	//
	// **Mit der Fassung, an der die Wanderung scheitert, führt kein
	// Zurückspielen heraus**: jede Sicherung muss durch dieselbe Wanderung, und
	// die Anwendung startet wieder nicht. Das ist kein Mangel — die Sicherung
	// ist die Kopie **vor** dem Eingriff und nicht dessen Heilung. Ihr Zweck
	// ist, dass der Stand nicht verloren geht; heraus führt die Fassung von
	// **vorher**.
	//
	// **Und der Rückweg dreht sich zurück** (§8): scheitert der neue Stand,
	// liegt danach der alte wieder da, und der nächste Versuch findet ihn.
	if zurueck := bindungen.SicherungZurueckspielen(t.Context(), automatisch, ""); zurueck.Fehler == nil {
		t.Error("die selbsttätige Sicherung ließ die Anwendung starten — " +
			"die Wanderung, an der sie scheitert, ist noch dieselbe")
	}

	// ── 7. **Der Rückweg, ohne Handarbeit** (#940): die Fassung von vorher
	// spielt die selbsttätige Sicherung zurück — über dieselbe Fassade wie das
	// Fenster —, und der Stand von vor dem Eingriff ist wieder da.
	alteFassung()

	antwort := bindungen.SicherungZurueckspielen(t.Context(), automatisch, "")
	if antwort.Fehler != nil {
		t.Fatalf("zurückspielen: %+v", antwort.Fehler)
	}

	t.Cleanup(func() { _ = neu.Beenden() }) //nolint:errcheck // nur zum Aufräumen

	// ── 8. Und jetzt läuft sie wieder, mit dem alten Stand.
	if got := spracheImBestand(t, bestand); got != "de" {
		t.Errorf("nach dem Zurückspielen steht %q im Bestand, erwartet war %q", got, "de")
	}

	if antwort := neu.Einstellungen(probekontext(t)); antwort.Fehler != nil {
		t.Fatalf("die Anwendung läuft nicht: %+v", antwort.Fehler)
	}
}

// verbuchte gibt die Namen der angewendeten Migrationen aus der Datei.
func verbuchte(t *testing.T, pfad string) []string {
	t.Helper()

	db, err := storage.FuerLaeufe("file:" + pfad + "?mode=ro")
	if err != nil {
		t.Fatalf("die Datenbank öffnen: %v", err)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur zum Aufräumen

	namen, fehler := storage.AngewendeteMigrationen(probekontext(t), db)
	if fehler != nil {
		t.Fatalf("die Buchführung lesen: %v", fehler)
	}

	return namen
}

// slicesEnthaelt sagt, ob ein Name in der Liste steht.
func slicesEnthaelt(namen []string, gesucht string) bool {
	for _, name := range namen {
		if name == gesucht {
			return true
		}
	}

	return false
}

// heuteVerbucht zählt, was diese Fassung in einem frischen Bestand verbucht.
func heuteVerbucht(t *testing.T) int {
	t.Helper()

	db, fehler := storage.Oeffnen(t.Context(), filepath.Join(t.TempDir(), "zaehlung.db"))
	if fehler != nil {
		t.Fatal(fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur gezählt

	if wanderfehler := storage.Migriere(t.Context(), db, nil); wanderfehler != nil {
		t.Fatal(wanderfehler)
	}

	verbucht, fehler := storage.AngewendeteMigrationen(t.Context(), db)
	if fehler != nil {
		t.Fatal(fehler)
	}

	return len(verbucht)
}
