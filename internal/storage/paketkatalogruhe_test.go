package storage_test

// **Ein Abgleich ohne Änderung schreibt nicht** —
// #1325.
//
// # Der Befund
//
// `App.Pakete` frischt bei **jedem** Aufruf den Katalog auf, und `Abgleichen`
// bestätigte dabei immer eine Transaktion — auch wenn kein Paket sich geändert
// hatte. Jeder Commit hebt `PRAGMA data_version` für die übrigen Verbindungen;
// das Fenster liest daraus, ob ein **anderer** Prozess geschrieben hat, und lud
// daraufhin alle Seiten neu.
//
// Am 2026-09-03 im laufenden Fenster gemessen: in zehn Sekunden Leerlauf fünfmal
// `Datenstand`, **zehnmal** `Sets`, fünfmal `Pakete` und fünfmal `Sammlung` —
// die Startseite flackerte im Sekundentakt. Die Kette ist ein Kreis: `Pakete`
// schreibt → der Stand steigt → die Startseite lädt neu → sie ruft `Pakete`.
//
// # Was hier geprüft wird und was nicht
//
// **Geprüft**: dass ein zweiter Abgleich mit denselben Paketen den Datenstand
// nicht hebt. **Nicht geprüft**: das Fenster; dort steht der Kreis, hier steht
// sein Anfang.

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine Datei und ein zweiter Vorrat darauf — der Leser steht für ein
// anderes Programm.**
//
// `PRAGMA data_version` ändert sich nur, wenn eine **andere** Verbindung
// bestätigt hat. Seit #1485 hat jeder Vorrat genau eine, und das eigene
// Schreiben bewegt den eigenen Stand nie — ein Leser auf demselben Vorrat
// wäre grün, ohne etwas zu messen (§3). Was #1325 verlangt, ist die Sicht
// des **anderen** Prozesses: der MCP-Server, ein zweites Fenster. Dafür
// steht der zweite Vorrat.
func katalogAufDatei(t *testing.T) (*storage.Paketkatalog, *storage.Standleser) {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "bestand.db")
	db := migrierterBestand(t, pfad)

	beobachter, err := storage.Oeffnen(probekontext(t), "file:"+pfad)
	if err != nil {
		t.Fatalf("den Beobachter öffnen: %v", err)
	}

	t.Cleanup(func() { _ = beobachter.Close() })

	return storage.NeuerPaketkatalog(db), storage.NeuerStandleser(beobachter)
}

func TestEinAbgleichOhneAenderungHebtDenDatenstandNicht(t *testing.T) {
	katalog, leser := katalogAufDatei(t)

	pakete := []domain.Paket{
		{
			Art: "mse-game", Name: "probe", Version: "2026-09-03",
			Pfad: "/p/probe.mse-game",
			Braucht: []domain.Bedarf{
				{Paket: "grund.mse-include", Version: "2026-09-03"},
			},
		},
	}

	if err := katalog.Abgleichen(probekontext(t), pakete); err != nil {
		t.Fatalf("erster Abgleich: %v", err)
	}

	vorher, err := leser.Datenstand(probekontext(t))
	if err != nil {
		t.Fatalf("Datenstand lesen: %v", err)
	}

	// **Dreimal derselbe Abgleich** — einmal genügte nicht als Beleg: der Kreis
	// im Fenster dreht sich, solange **jeder** Aufruf schreibt.
	for i := range 3 {
		if fehler := katalog.Abgleichen(probekontext(t), pakete); fehler != nil {
			t.Fatalf("Abgleich %d: %v", i+2, fehler)
		}
	}

	nachher, err := leser.Datenstand(probekontext(t))
	if err != nil {
		t.Fatalf("Datenstand lesen: %v", err)
	}

	if nachher != vorher {
		t.Errorf("der Datenstand ist von %d auf %d gestiegen, obwohl sich nichts geändert hat",
			vorher, nachher)
	}
}

// **Eine echte Änderung hebt ihn sehr wohl** — die Gegenprobe.
//
// Ohne sie sagte der Test darüber nur, dass nichts geschrieben wird; er sagte
// nichts darüber, ob überhaupt noch etwas ankommt.
func TestEinAbgleichMitAenderungHebtDenDatenstand(t *testing.T) {
	katalog, leser := katalogAufDatei(t)

	erst := []domain.Paket{
		{Art: "mse-game", Name: "probe", Version: "2026-09-03", Pfad: "/p/probe.mse-game"},
	}

	if err := katalog.Abgleichen(probekontext(t), erst); err != nil {
		t.Fatalf("erster Abgleich: %v", err)
	}

	vorher, err := leser.Datenstand(probekontext(t))
	if err != nil {
		t.Fatalf("Datenstand lesen: %v", err)
	}

	dann := []domain.Paket{
		{Art: "mse-game", Name: "probe", Version: "2026-09-04", Pfad: "/p/probe.mse-game"},
	}

	if fehler := katalog.Abgleichen(probekontext(t), dann); fehler != nil {
		t.Fatalf("zweiter Abgleich: %v", fehler)
	}

	nachher, err := leser.Datenstand(probekontext(t))
	if err != nil {
		t.Fatalf("Datenstand lesen: %v", err)
	}

	if nachher == vorher {
		t.Errorf("der Datenstand steht bei %d, obwohl sich die Version geändert hat", vorher)
	}
}

// **Der Setkopf kommt aus dem Bestand zurück** —
// #1325.
//
// **Hier und nicht in einer eigenen Datei**: beide Läufe brauchen eine
// Datenbank auf einer **Datei**, und `katalogAufDatei` steht schon da.
func TestDerSetkopfKommtZurueck(t *testing.T) {
	db := migrierterBestand(t, filepath.Join(t.TempDir(), "bestand.db"))

	kopf := domain.Setkopf{
		MSEFassung:        "2.1.2",
		Spielfassung:      "2020-04-25",
		Stilfassung:       "2020-09-04",
		Versionskontrolle: "none",
		Apprenticecode:    "APP",
	}

	kennung, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Probe", Game: "magic", Kopf: kopf}, nil)
	if err != nil {
		t.Fatalf("das Set anlegen: %v", err)
	}

	gelesen, err := storage.NeuerSetLeser(db).Setkopf(probekontext(t), kennung)
	if err != nil {
		t.Fatalf("den Setkopf lesen: %v", err)
	}

	if gelesen != kopf {
		t.Errorf("der Kopf ist %+v, erwartet war %+v", gelesen, kopf)
	}
}

// **Ein Set, das es nicht gibt, ist ein Fehler und kein leerer Kopf.**
//
// Ein leerer Kopf sähe aus wie „stand nicht in der Datei" und wäre eine
// Falschauskunft (§8).
func TestEinFehlendesSetHatKeinenKopf(t *testing.T) {
	db := migrierterBestand(t, filepath.Join(t.TempDir(), "bestand.db"))

	_, err := storage.NeuerSetLeser(db).Setkopf(probekontext(t), 4711)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet war „nicht gefunden\"", err)
	}
}

// **Und ein Lesefehler kommt als Fehler an** — der dritte Weg.
func TestEinUnlesbarerSetkopfWirdGemeldet(t *testing.T) {
	db, err := sql.Open("stolper-fragen", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	if _, fehler := storage.NeuerSetLeser(db).Setkopf(probekontext(t), 1); fehler == nil {
		t.Error("kein Fehler, obwohl die Abfrage scheiterte")
	}
}
