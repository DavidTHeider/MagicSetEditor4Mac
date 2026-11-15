package storage_test

// **Die eigenen Autoersetzungen eines Spiels** — Werkstück E von Phase 4.11
// (#1341,
// Vermerk [0060](../../docs/entscheidungen/0060-eigene-autoersetzungen-liegen-im-bestand.md)).

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Eine Regel wird angelegt und wiedergefunden.**
func TestEineAutoersetzungWirdAngelegtUndGelesen(t *testing.T) {
	ablage := storage.NeueAutoersetzungsablage(frischeDatenbank(t))

	regel := domain.Autoersetzungsregel{
		Muster: "(W)", Ersatz: "<sym>W</sym>",
		GanzesWort: true, Aktiv: true, Eigen: true,
	}

	if err := ablage.Setzen(probekontext(t), "magic", regel); err != nil {
		t.Fatalf("Setzen: %v", err)
	}

	regeln, err := ablage.Regeln(probekontext(t), "magic")
	if err != nil {
		t.Fatalf("Regeln: %v", err)
	}

	if len(regeln) != 1 || regeln[0] != regel {
		t.Errorf("die Regeln lauten %+v", regeln)
	}
}

// **Dasselbe Muster zweimal ist dieselbe Regel.**
//
// MSE vergleicht über `match` (`GameSettings::initDefaults`,
// `src/data/settings.cpp`, Zeile 93, Fassung f9d9356 vom 2021-07-06). Zwei
// Zeilen wären zwei Regeln, die
// beide greifen — und die zweite fände nichts mehr.
func TestDasselbeMusterZweimalIstDieselbeRegel(t *testing.T) {
	ablage := storage.NeueAutoersetzungsablage(frischeDatenbank(t))

	for _, ersatz := range []string{"erst", "dann"} {
		if err := ablage.Setzen(probekontext(t), "magic", domain.Autoersetzungsregel{
			Muster: "(W)", Ersatz: ersatz, GanzesWort: true, Aktiv: true, Eigen: true,
		}); err != nil {
			t.Fatalf("Setzen: %v", err)
		}
	}

	regeln, _ := ablage.Regeln(probekontext(t), "magic") //nolint:errcheck // oben geprüft
	if len(regeln) != 1 || regeln[0].Ersatz != "dann" {
		t.Errorf("die Regeln lauten %+v", regeln)
	}
}

// **Ein leeres Muster wird abgewiesen.**
func TestEineAutoersetzungOhneMusterWirdAbgewiesen(t *testing.T) {
	ablage := storage.NeueAutoersetzungsablage(frischeDatenbank(t))

	if err := ablage.Setzen(probekontext(t), "magic", domain.Autoersetzungsregel{
		Muster: "   ", Ersatz: "x",
	}); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("der Fehler lautet %v", err)
	}
}

// **Löschen nimmt sie heraus — und eine, die es nicht gibt, ist ein Befund.**
func TestEineAutoersetzungWirdGeloescht(t *testing.T) {
	ablage := storage.NeueAutoersetzungsablage(frischeDatenbank(t))

	if err := ablage.Setzen(probekontext(t), "magic", domain.Autoersetzungsregel{
		Muster: "(W)", Ersatz: "x", Eigen: true,
	}); err != nil {
		t.Fatalf("Setzen: %v", err)
	}

	if err := ablage.Loeschen(probekontext(t), "magic", "(W)"); err != nil {
		t.Fatalf("Loeschen: %v", err)
	}

	regeln, _ := ablage.Regeln(probekontext(t), "magic") //nolint:errcheck // oben geprüft
	if len(regeln) != 0 {
		t.Errorf("die Regeln lauten %+v", regeln)
	}

	if err := ablage.Loeschen(probekontext(t), "magic", "(W)"); !errors.Is(
		err, domain.ErrNotFound) {
		t.Errorf("der zweite Fehler lautet %v", err)
	}
}

// **Zurücksetzen nimmt alles heraus — Regeln und Schalter.**
func TestZuruecksetzenNimmtRegelnUndSchalter(t *testing.T) {
	ablage := storage.NeueAutoersetzungsablage(frischeDatenbank(t))

	if err := ablage.Setzen(probekontext(t), "magic", domain.Autoersetzungsregel{
		Muster: "(W)", Ersatz: "x", Eigen: true,
	}); err != nil {
		t.Fatalf("Setzen: %v", err)
	}

	if err := ablage.SchalterSetzen(probekontext(t), "magic", false); err != nil {
		t.Fatalf("SchalterSetzen: %v", err)
	}

	if err := ablage.Zuruecksetzen(probekontext(t), "magic"); err != nil {
		t.Fatalf("Zuruecksetzen: %v", err)
	}

	regeln, _ := ablage.Regeln(probekontext(t), "magic") //nolint:errcheck // oben geprüft
	if len(regeln) != 0 {
		t.Errorf("die Regeln lauten %+v", regeln)
	}

	// **Und der Schalter steht wieder auf „an"** — MSEs `onDefault` setzt ihn
	// ausdrücklich (`auto_replace_window.cpp`, Zeile 235).
	an, err := ablage.Schalter(probekontext(t), "magic")
	if err != nil || !an {
		t.Errorf("der Schalter steht auf %v (%v)", an, err)
	}
}

// **Ohne Zeile gilt „an".**
//
// MSEs `use_auto_replace(true)` (`src/data/settings.cpp`, Zeile 74, Fassung
// f9d9356 vom 2021-07-06). Es ist
// die Vorgabe, die niemanden überrascht: die Regeln greifen schon heute.
func TestOhneZeileStehtDerSchalterAufAn(t *testing.T) {
	ablage := storage.NeueAutoersetzungsablage(frischeDatenbank(t))

	an, err := ablage.Schalter(probekontext(t), "magic")
	if err != nil || !an {
		t.Errorf("der Schalter steht auf %v (%v)", an, err)
	}
}

// **Und er lässt sich umlegen.**
func TestDerSchalterLaesstSichUmlegen(t *testing.T) {
	ablage := storage.NeueAutoersetzungsablage(frischeDatenbank(t))

	if err := ablage.SchalterSetzen(probekontext(t), "magic", false); err != nil {
		t.Fatalf("SchalterSetzen: %v", err)
	}

	an, err := ablage.Schalter(probekontext(t), "magic")
	if err != nil || an {
		t.Errorf("der Schalter steht auf %v (%v)", an, err)
	}
}

// **Ein leerer Spielname wird überall abgewiesen.**
//
// Dieselbe Wache wie bei den Listenspalten: ein Spiel ohne Namen ist kein
// Spiel, und eine Zeile darunter gehörte niemandem.
func TestEinLeeresSpielWirdBeiDenAutoersetzungenAbgewiesen(t *testing.T) {
	ablage := storage.NeueAutoersetzungsablage(frischeDatenbank(t))

	if _, err := ablage.Regeln(probekontext(t), ""); err == nil {
		t.Error("Regeln nahm den leeren Namen")
	}

	if err := ablage.Setzen(probekontext(t), "", domain.Autoersetzungsregel{
		Muster: "x",
	}); err == nil {
		t.Error("Setzen nahm den leeren Namen")
	}

	if err := ablage.Loeschen(probekontext(t), "", "x"); err == nil {
		t.Error("Loeschen nahm den leeren Namen")
	}

	if err := ablage.Zuruecksetzen(probekontext(t), ""); err == nil {
		t.Error("Zuruecksetzen nahm den leeren Namen")
	}

	if _, err := ablage.Schalter(probekontext(t), ""); err == nil {
		t.Error("Schalter nahm den leeren Namen")
	}

	if err := ablage.SchalterSetzen(probekontext(t), "", true); err == nil {
		t.Error("SchalterSetzen nahm den leeren Namen")
	}
}

// **Jeder Weg meldet einen kaputten Bestand.**
//
// # Warum das hier steht
//
// Weil §8 verlangt, dass ein Fehler **ankommt**, statt zu einer leeren Liste
// zu werden. Eine geschlossene Datenbank ist die billigste Art, jeden dieser
// Wege einmal scheitern zu sehen — dieselbe Vorrichtung wie in
// `fehlerpfade_test.go`.
func TestJederAutoersetzungswegMeldetEinenKaputtenBestand(t *testing.T) {
	ablage := storage.NeueAutoersetzungsablage(kaputterBestand(t))

	if _, err := ablage.Regeln(probekontext(t), "magic"); err == nil {
		t.Error("Regeln schwieg")
	}

	if err := ablage.Setzen(probekontext(t), "magic", domain.Autoersetzungsregel{
		Muster: "(W)",
	}); err == nil {
		t.Error("Setzen schwieg")
	}

	if err := ablage.Loeschen(probekontext(t), "magic", "(W)"); err == nil {
		t.Error("Loeschen schwieg")
	}

	if err := ablage.Zuruecksetzen(probekontext(t), "magic"); err == nil {
		t.Error("Zuruecksetzen schwieg")
	}

	if _, err := ablage.Schalter(probekontext(t), "magic"); err == nil {
		t.Error("Schalter schwieg")
	}

	if err := ablage.SchalterSetzen(probekontext(t), "magic", true); err == nil {
		t.Error("SchalterSetzen schwieg")
	}
}

// **Und eine Zeile, die sich nicht lesen lässt, ebenso.**
//
// Der Weg über `zeilen.Scan`: die Spalte `aktiv` ist als `INTEGER` erklärt,
// und ein Text darin lässt sich nicht in ein `bool` lesen. Ohne diesen Fall
// bliebe der Zweig unbetreten — und er ist der, an dem eine halbe Liste
// zurückkäme.
func TestEineUnlesbareAutoersetzungszeileIstEinFehler(t *testing.T) {
	db := frischeDatenbank(t)

	// **Die Tabelle wird neu gelegt, ohne `STRICT`** — sonst weist SQLite
	// den Text schon beim Schreiben ab. Ein Bestand, den ein anderes
	// Programm angefasst hat, sieht so aus; §8 verlangt, dass er sich meldet.
	for _, satz := range []string{
		`DROP TABLE spiel_autoersetzungen`,
		`CREATE TABLE spiel_autoersetzungen (
		   spiel TEXT, muster TEXT, ersatz TEXT,
		   ganzes_wort INTEGER, aktiv TEXT, eigen INTEGER, profil_id INTEGER)`,
		`INSERT INTO spiel_autoersetzungen
		   (spiel, muster, ersatz, ganzes_wort, aktiv, eigen, profil_id)
		 VALUES ('magic', '(W)', 'x', 1, 'kein Wahrheitswert', 1, 1)`,
	} {
		if _, err := db.ExecContext(probekontext(t), satz); err != nil {
			t.Fatalf("%s: %v", satz, err)
		}
	}

	if _, err := storage.NeueAutoersetzungsablage(db).
		Regeln(probekontext(t), "magic"); err == nil {
		t.Error("die unlesbare Zeile kam durch")
	}
}

// **Und ein halb kaputter Bestand meldet sich beim zweiten Satz — und lässt
// nichts halb getan zurück.**
//
// `Zuruecksetzen` fasst **zwei** Tabellen an. Der erste Anlauf prüfte nur,
// dass ein Fehler zurückkommt, und beschrieb den halben Stand dabei
// ausdrücklich: „die Regeln wären weg und der Schalter bliebe stehen". Genau
// das ist ein Fehlerweg, der einen halben Stand zurücklässt (§10) —
// angestrichen als Befund 4 des Zweitblicks über Phase 4.11.
//
// **MSE behandelt es als eine Bedienhandlung** (`AutoReplaceWindow::onDefault`,
// Zeilen 234–238). Jetzt tut es das hier auch.
func TestZuruecksetzenLaesstNichtsHalbGetanZurueck(t *testing.T) {
	db := frischeDatenbank(t)

	ablage := storage.NeueAutoersetzungsablage(db)

	if err := ablage.Setzen(probekontext(t), "magic", domain.Autoersetzungsregel{
		Muster: "(W)", Ersatz: "weiss", Aktiv: true, Eigen: true,
	}); err != nil {
		t.Fatalf("die Regel anlegen: %v", err)
	}

	// **Die zweite Tabelle fällt weg**, und damit scheitert der zweite Satz.
	if _, err := db.ExecContext(probekontext(t),
		`DROP TABLE spiel_autoersetzung_an`); err != nil {
		t.Fatalf("die Tabelle entfernen: %v", err)
	}

	if err := ablage.Zuruecksetzen(probekontext(t), "magic"); err == nil {
		t.Fatal("der zweite Satz schwieg")
	}

	// **Und die Regel steht noch da.** Ohne die Transaktion wäre sie weg —
	// das ist der halbe Stand, um den es geht.
	var gezaehlt int

	if err := db.QueryRowContext(probekontext(t),
		`SELECT count(*) FROM spiel_autoersetzungen WHERE spiel = ?`,
		"magic").Scan(&gezaehlt); err != nil {
		t.Fatalf("nachzählen: %v", err)
	}

	if gezaehlt != 1 {
		t.Errorf("%d Regeln — die Regeln sind weg, obwohl das Zurücksetzen scheiterte", gezaehlt)
	}
}

// kaputterBestand gibt eine Datenbank, die auf jeden Ruf mit einem Fehler
// antwortet.
//
// **Eine geschlossene und keine gestörte.** Der Unterschied ist keiner für
// diesen Zweck: jeder Ruf scheitert, und das ist, was hier geprüft wird.
// `fehlerpfade_test.go` hält dieselbe Vorrichtung für `package storage`; hier
// steht sie noch einmal, weil dieser Lauf von **außen** prüft.
func kaputterBestand(t *testing.T) *sql.DB {
	t.Helper()

	db, err := storage.FuerLaeufe("file:" + t.TempDir() + "/zu.db")
	if err != nil {
		t.Fatalf("Datenbank nicht öffnenbar: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	return db
}

// **Und der erste Satz meldet sich ebenso.**
//
// `Zuruecksetzen` fasst zwei Tabellen an; ohne diesen Fall bliebe der **erste**
// Fehlerzweig unbetreten.
func TestZuruecksetzenMeldetDenErstenSatz(t *testing.T) {
	db := frischeDatenbank(t)

	if _, err := db.ExecContext(probekontext(t),
		`DROP TABLE spiel_autoersetzungen`); err != nil {
		t.Fatalf("die Tabelle entfernen: %v", err)
	}

	if err := storage.NeueAutoersetzungsablage(db).
		Zuruecksetzen(probekontext(t), "magic"); err == nil {
		t.Error("der erste Satz schwieg")
	}
}
