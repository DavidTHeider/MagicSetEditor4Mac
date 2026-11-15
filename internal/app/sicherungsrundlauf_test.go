package app_test

// **Der Rundlauf auf einen leeren Bestand** — OPS-06.AK-03
// (#1723):
// gesichert mit Kennwort, zurückgespielt auf einem Rechner, der nichts hat,
// ohne Netz — und danach **Tabelle für Tabelle** verglichen.
//
// # Warum mit zwei Profilen
//
// **Weil `deckleihen` und `kostenposten` das Profil im Schlüssel tragen**
// (0081, 0082). Eine Wiederherstellung, die Zeilen über die Kennung
// zusammenführte, träfe genau dort daneben — nachgeschärft nach Phase 6.1.
// Diese tauscht die ganze Datei; der Vergleich zeigt, dass danach jede Zeile
// jeder Tabelle dieselbe ist.

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

const rundlaufkennwort = "korrekt pferd batterie heftklammer"

// nachZurueckspielen sind die Tabellen, in die das Zurückspielen selbst
// schreibt — **der Vermerk über die Rückkehr**, und sonst nichts
// (O7, #1431).
var nachZurueckspielen = []string{"wiederherstellungsepochen"}

// nachDemSichern ist die Tafel, in die das Sichern selbst schreibt — **erst
// nach der zurückgelesenen Datei** (OPS01,
// #2347):
// eine Sicherung kennt den Vermerk über sich nicht. Jeder andere Vermerk
// kommt mit.
const nachDemSichern = "sicherungsvermerke"

func TestEineVerschluesselteSicherungKommtAufEinemLeerenBestandGanzZurueck(t *testing.T) {
	t.Cleanup(sicherung.ArbeitsfaktorSetzen(10))
	t.Setenv("HOME", t.TempDir())

	ctx := t.Context()
	quellort := filepath.Join(t.TempDir(), "bestand.db")
	ziel := filepath.Join(t.TempDir(), "rundlauf"+sicherung.Endung)

	quelle := laufendeAnwendung(t, quellort)

	neuer := quelle.NutzerAnlegen(ctx, "Zweite", "")
	if neuer.Fehler != nil {
		t.Fatalf("Nutzer: %+v", neuer.Fehler)
	}

	if profil := quelle.ProfilAnlegen(ctx, neuer.Nutzer.ID, "Zweites Profil"); profil.Fehler != nil {
		t.Fatalf("Profil: %+v", profil.Fehler)
	}

	gesichert := func(context.Context, app.Dateiauswahl) (string, error) { return ziel, nil }
	quelle.MitDateiwahl(gesichert)

	if antwort := quelle.EinstellungSetzen(ctx, "sprache", "en"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	if antwort := quelle.BestandSichern(ctx, rundlaufkennwort, false); antwort.Fehler != nil {
		t.Fatalf("sichern: %+v", antwort.Fehler)
	}

	// **Ein leerer Bestand, ein anderer Ort** — der Rechner, auf dem nichts ist.
	zielort := filepath.Join(t.TempDir(), "bestand.db")
	leer := laufendeAnwendung(t, zielort)

	schau := leer.SicherungBeschauen(ctx, ziel, rundlaufkennwort)
	if schau.Fehler != nil || !schau.Verschluesselt {
		t.Fatalf("beschauen: %+v", schau)
	}

	if antwort := leer.SicherungZurueckspielen(ctx, ziel, rundlaufkennwort); antwort.Fehler != nil {
		t.Fatalf("zurückspielen: %+v", antwort.Fehler)
	}

	if fehler := quelle.Beenden(); fehler != nil {
		t.Fatal(fehler)
	}

	if fehler := leer.Beenden(); fehler != nil {
		t.Fatal(fehler)
	}

	vorher, nachher := tabellen(t, quellort), tabellen(t, zielort)

	for name, zeilen := range vorher {
		if slices.Contains(nachZurueckspielen, name) {
			continue
		}

		if name == nachDemSichern {
			zeilen = slices.DeleteFunc(slices.Clone(zeilen), func(zeile string) bool {
				return strings.Contains(zeile, "|"+filepath.Base(ziel)+"|")
			})
		}

		if got := nachher[name]; !slices.Equal(got, zeilen) {
			t.Errorf("%s: %d Zeilen vorher, %d danach\nvorher: %.300v\ndanach: %.300v",
				name, len(zeilen), len(got), zeilen, got)
		}
	}

	if len(nachher) != len(vorher) {
		t.Errorf("%d Tabellen vorher, %d danach", len(vorher), len(nachher))
	}
}

// tabellen liest jede Tabelle als Zeilen in fester Ordnung.
func tabellen(t *testing.T, ort string) map[string][]string {
	t.Helper()

	db, fehler := storage.Oeffnen(t.Context(), ort)
	if fehler != nil {
		t.Fatal(fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur gelesen

	namen := spalte(t, db, `SELECT name FROM sqlite_master
		WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)

	inhalt := map[string][]string{}

	// **Auch leere Tabellen stehen da** — sonst fiele eine fehlende Tabelle
	// nicht auf, solange sie vorher leer war.
	for _, name := range namen {
		inhalt[name] = zeilenVon(t, db, name)
	}

	return inhalt
}

// zeilenVon liest eine Tabelle als sortierte Zeilen.
func zeilenVon(t *testing.T, db *sql.DB, name string) []string {
	t.Helper()

	zeilen, fehler := db.QueryContext(t.Context(), fmt.Sprintf(`SELECT * FROM %q`, name)) //nolint:gosec // Namen aus sqlite_master
	if fehler != nil {
		t.Fatalf("%s: %v", name, fehler)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // nur gelesen

	spalten, _ := zeilen.Columns() //nolint:errcheck // eben gelesen
	werte := make([]any, len(spalten))
	zeiger := make([]any, len(spalten))

	for i := range werte {
		zeiger[i] = &werte[i]
	}

	gelesen := []string{}

	for zeilen.Next() {
		if fehler := zeilen.Scan(zeiger...); fehler != nil {
			t.Fatal(fehler)
		}

		teile := make([]string, len(werte))
		for i, wert := range werte {
			teile[i] = fmt.Sprintf("%v", wert)
		}

		gelesen = append(gelesen, strings.Join(teile, "|"))
	}

	if fehler := zeilen.Err(); fehler != nil {
		t.Fatal(fehler)
	}

	slices.Sort(gelesen)

	return gelesen
}

func spalte(t *testing.T, db *sql.DB, abfrage string) []string {
	t.Helper()

	zeilen, fehler := db.QueryContext(t.Context(), abfrage)
	if fehler != nil {
		t.Fatal(fehler)
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // nur gelesen

	var werte []string

	for zeilen.Next() {
		var wert string
		if fehler := zeilen.Scan(&wert); fehler != nil {
			t.Fatal(fehler)
		}

		werte = append(werte, wert)
	}

	if fehler := zeilen.Err(); fehler != nil {
		t.Fatal(fehler)
	}

	return werte
}

// **Ein zu kurzes Kennwort hält vor dem Dialog an** — am Feld, und der
// Dialog geht gar nicht erst auf.
func TestEinKurzesKennwortOeffnetKeinenDialog(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	anwendung := laufendeAnwendung(t, filepath.Join(t.TempDir(), "bestand.db"))

	gefragt := false
	anwendung.MitDateiwahl(func(context.Context, app.Dateiauswahl) (string, error) {
		gefragt = true

		return "", nil
	})

	antwort := anwendung.BestandSichern(t.Context(), "kurz", false)
	if antwort.Fehler == nil || antwort.Fehler.Schluessel != domain.SchluesselSicherungskennwortKurz || gefragt {
		t.Errorf("Fehler %+v, Dialog %v", antwort.Fehler, gefragt)
	}
}

// **Die Flächen ohne Fenster sichern mit demselben Wunsch** — über die Naht,
// die Kommandozeile, Dienst und Werkzeugserver benutzen.
func TestDieNahtOhneFensterSichertMitKennwort(t *testing.T) {
	t.Cleanup(sicherung.ArbeitsfaktorSetzen(10))
	t.Setenv("HOME", t.TempDir())

	ort := filepath.Join(t.TempDir(), "bestand.db")
	db, fehler := storage.Oeffnen(t.Context(), ort)
	if fehler != nil {
		t.Fatal(fehler)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // Aufräumen

	if fehler := storage.Migriere(t.Context(), db, nil); fehler != nil {
		t.Fatal(fehler)
	}

	ziel := filepath.Join(t.TempDir(), "naht"+sicherung.Endung)
	if fehler := app.BestandsnahtUeber(&db, ort).Sichern(
		app.MitSicherungswunsch(t.Context(), domain.FlaecheCLI, rundlaufkennwort, false), ziel); fehler != nil {
		t.Fatalf("sichern: %v", fehler)
	}

	if !sicherung.IstVerschluesselt(ziel) {
		t.Error("die Sicherung ist nicht verschlüsselt")
	}
}
