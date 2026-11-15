package storage_test

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Welche Pakete auf ein Muster passen**
// (#632).
//
// **ported.** MSE fragt für eine `package choice` seinen Paketverwalter —
// `package_manager.findMatching(field().match, choices)` — und sortiert das
// Ergebnis nach `position_hint`, bei Gleichstand nach dem Namen
// (`src/render/value/package_choice.cpp`, Zeilen 28–43, Fassung f9d9356 vom
// 2021-07-06).
//
// `match:` nennt einen Dateinamen samt Endung mit `*` als Platzhalter. Hier
// zerfällt er in Art und Namensmuster: die Endung ist die Spalte `art`, der
// Rest ein `GLOB` auf `name`.
func TestPaketeNachMusterKommenInMsesOrdnung(t *testing.T) {
	db := frischeDatenbank(t)
	katalog := storage.NeuerPaketkatalog(db)

	pakete := []domain.Paket{
		{Art: "mse-include", Name: "probe-overlay-zink", Kurzname: "zink", Pfad: "/z", Rang: 5},
		{Art: "mse-include", Name: "probe-overlay-alpha", Kurzname: "alpha", Pfad: "/a", Rang: 9},
		{Art: "mse-include", Name: "probe-overlay-beta", Kurzname: "beta", Pfad: "/b", Rang: 5},
		{Art: "mse-include", Name: "fremd-overlay", Kurzname: "fremd", Pfad: "/f", Rang: 1},
		{Art: "mse-symbol-font", Name: "probe-overlay-schrift", Pfad: "/s", Rang: 1},
	}

	if err := katalog.Abgleichen(probekontext(t), pakete); err != nil {
		t.Fatalf("abgleichen: %v", err)
	}

	gefunden, err := katalog.PaketeNachMuster(probekontext(t), "probe-overlay-*.mse-include")
	if err != nil {
		t.Fatalf("suchen: %v", err)
	}

	// **Rang zuerst, dann der Name** — `zink` und `beta` haben beide Rang 5,
	// also entscheidet der Name; `alpha` steht mit Rang 9 dahinter.
	namen := make([]string, 0, len(gefunden))
	for _, paket := range gefunden {
		namen = append(namen, paket.Name)
	}

	erwartet := []string{"probe-overlay-beta", "probe-overlay-zink", "probe-overlay-alpha"}

	if len(namen) != len(erwartet) {
		t.Fatalf("%v, erwartet %v", namen, erwartet)
	}

	for i := range erwartet {
		if namen[i] != erwartet[i] {
			t.Errorf("Stelle %d: %q, erwartet %q", i, namen[i], erwartet[i])
		}
	}

	// **Der Kurzname reist mit** — er ist es, was zur Wahl dasteht.
	if gefunden[0].Kurzname != "beta" {
		t.Errorf("Kurzname %q, erwartet `beta`", gefunden[0].Kurzname)
	}
}

// **Ein Muster ohne Endung findet nichts** — und meldet es, statt still eine
// leere Liste zu geben.
//
// `match:` nennt bei MSE immer einen Dateinamen samt Endung. Ein Muster ohne
// Punkt ist eine Angabe, die niemand beantworten kann: die Art steht dann
// nirgends, und jede Art zu nehmen hieße, `magic-mana-*` auf Spielpakete
// passen zu lassen.
func TestEinMusterOhneEndungIstEinBefund(t *testing.T) {
	db := frischeDatenbank(t)

	_, err := storage.NeuerPaketkatalog(db).PaketeNachMuster(probekontext(t), "probe-overlay-*")
	if err == nil {
		t.Error("kein Fehler auf ein Muster ohne Endung")
	}
}

// **Die Fehlerwege der Mustersuche.**
//
// Paketintern nicht erreichbar: mit einer funktionierenden Datenbank
// funktioniert sie. Über den Stolpertreiber schon.
func TestDieMustersucheMeldetIhreFehlerwege(t *testing.T) {
	faelle := map[string]string{
		"die Abfrage scheitert":       "stolper-fragen",
		"eine Zeile liest sich nicht": "stolper-lesen",
		"das Zeilenende stolpert":     "stolper-umwandeln",
	}

	for name, treiber := range faelle {
		t.Run(name, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // gelesen ist gelesen

			if _, fehler := storage.NeuerPaketkatalog(db).
				PaketeNachMuster(probekontext(t), "probe-*.mse-include"); fehler == nil {
				t.Errorf("%s: kein Fehler", name)
			}
		})
	}
}
