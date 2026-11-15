package service_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Die Übernahme aus einer MSE2-Installation (#55).
//
// **Der Hauptweg**, weil er der einzige ist, bei dem die Rechtslage klar ist:
// die Dateien liegen bereits auf dem Rechner der Nutzenden (§11).

// mse2 baut ein Verzeichnis, wie MSE2 es unter `data/` hat.
func mse2(t *testing.T) string {
	t.Helper()

	ort := t.TempDir()
	paket(t, ort, "magic.mse-game", "mse version: 2.0.0\nshort name: Magic\nversion: 2009-07-23\n")
	paket(t, ort, "magic-m15.mse-style", "mse version: 2.0.0\ngame: magic\nversion: 2024-01-01\n")

	// Ein Set gehört nicht dazu — es ist ein Dokument, kein Vorlagenpaket.
	paket(t, ort, "meins.mse-set", "mse version: 2.0.0\ngame: magic\n")

	// Und eine lose Datei erst recht nicht.
	if err := os.WriteFile(filepath.Join(ort, "liesmich.txt"), []byte("hallo"), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	return ort
}

// **Was in der Installation liegt, kommt herüber — mit seiner Herkunft.**
func TestDieUebernahmeHoltDieVorlagenMitHerkunft(t *testing.T) {
	quelle := mse2(t)
	ziel := t.TempDir()
	katalog := &ablage{}
	verwalter := service.NeuerPaketverwalter(katalog, ziel, ziel)

	ergebnis, err := verwalter.Uebernehmen(t.Context(), quelle)
	if err != nil {
		t.Fatalf("übernehmen: %v", err)
	}

	if len(ergebnis.Uebernommen) != 2 {
		t.Fatalf("übernommen: %+v", ergebnis.Uebernommen)
	}

	if len(ergebnis.Befunde) > 0 {
		t.Fatalf("Befunde: %v", ergebnis.Befunde)
	}

	for _, paket := range ergebnis.Uebernommen {
		if paket.Herkunft != filepath.Join(quelle, paket.Name+"."+paket.Art) {
			t.Errorf("Herkunft von %q: %q", paket.Name, paket.Herkunft)
		}
	}

	// **Das Set bleibt, wo es ist.** Ein `.mse-set` ist ein Dokument der
	// Nutzenden; im Vorlagenverzeichnis fände es niemand wieder.
	if _, fehler := os.Stat(filepath.Join(ziel, "meins.mse-set")); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("das Set kam mit: %v", fehler)
	}
}

// **Was schon dasteht, wird übergangen und benannt.**
//
// Nicht überschrieben: wer eine Vorlage bearbeitet hat, verlöre sie sonst.
func TestDieUebernahmeUebergehtWasSchonDasteht(t *testing.T) {
	quelle := mse2(t)
	ziel := t.TempDir()
	paket(t, ziel, "magic.mse-game", "mse version: 2.0.0\nshort name: Meines\n")

	verwalter := service.NeuerPaketverwalter(&ablage{}, ziel, ziel)

	ergebnis, err := verwalter.Uebernehmen(t.Context(), quelle)
	if err != nil {
		t.Fatalf("übernehmen: %v", err)
	}

	if len(ergebnis.Uebergangen) != 1 || ergebnis.Uebergangen[0] != "magic.mse-game" {
		t.Errorf("übergangen: %v", ergebnis.Uebergangen)
	}

	if len(ergebnis.Uebernommen) != 1 {
		t.Errorf("übernommen: %+v", ergebnis.Uebernommen)
	}
}

// **Ein kaputtes Paket ist ein Befund, und die anderen kommen trotzdem.**
func TestEinKaputtesPaketHaeltDieUebernahmeNichtAuf(t *testing.T) {
	quelle := mse2(t)
	if err := os.MkdirAll(filepath.Join(quelle, "kaputt.mse-style"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	ziel := t.TempDir()
	verwalter := service.NeuerPaketverwalter(&ablage{}, ziel, ziel)

	ergebnis, err := verwalter.Uebernehmen(t.Context(), quelle)
	if err != nil {
		t.Fatalf("übernehmen: %v", err)
	}

	if len(ergebnis.Befunde) != 1 {
		t.Fatalf("Befunde: %v", ergebnis.Befunde)
	}

	if len(ergebnis.Uebernommen) != 2 {
		t.Errorf("übernommen: %+v", ergebnis.Uebernommen)
	}
}

// **Ein Verzeichnis, das keines ist, wird abgewiesen.**
func TestDieUebernahmeWeistEinNichtverzeichnisAb(t *testing.T) {
	verwalter := service.NeuerPaketverwalter(&ablage{}, t.TempDir(), t.TempDir())

	_, err := verwalter.Uebernehmen(t.Context(), filepath.Join(t.TempDir(), "gibtsnicht"))
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("kein Validierungsfehler: %v", err)
	}
}

// **Ein Fehler des Bestands kommt durch — und was auf der Platte liegt, bleibt.**
func TestEinBestandsfehlerHaeltDieUebernahmeNichtZurueck(t *testing.T) {
	quelle := mse2(t)
	ziel := t.TempDir()
	kaputt := errors.New("der Katalog schweigt")

	verwalter := service.NeuerPaketverwalter(&ablage{fehler: kaputt}, ziel, ziel)

	ergebnis, err := verwalter.Uebernehmen(t.Context(), quelle)
	if !errors.Is(err, kaputt) {
		t.Fatalf("Fehler: %v", err)
	}

	if _, fehler := os.Stat(filepath.Join(ziel, "magic.mse-game")); fehler != nil {
		t.Errorf("die Kopie ist verschwunden: %v", fehler)
	}

	if ergebnis == nil {
		t.Fatal("kein Ergebnis")
	}
}

// **Was sich nicht kopieren lässt, ist ein Befund.**
func TestEinUnkopierbaresPaketWirdBenannt(t *testing.T) {
	quelle := t.TempDir()
	paket(t, quelle, "sperrig.mse-style", "mse version: 2.0.0\n")

	gesperrt := filepath.Join(quelle, "sperrig.mse-style", "unlesbar")
	if err := os.WriteFile(gesperrt, []byte("geheim"), 0o000); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	ziel := t.TempDir()
	verwalter := service.NeuerPaketverwalter(&ablage{}, ziel, ziel)

	ergebnis, err := verwalter.Uebernehmen(t.Context(), quelle)
	if err != nil {
		t.Fatalf("übernehmen: %v", err)
	}

	if len(ergebnis.Befunde) != 1 {
		t.Errorf("Befunde: %v", ergebnis.Befunde)
	}
}

// **Ein Lesefehler beim Katalog lässt die Übernahme stehen.**
//
// Der Abgleich ist geglückt; was danach beim Lesen schiefgeht, macht sie nicht
// ungeschehen (§7, Regel 3). Die Liste bleibt leer, die Zahlen stehen.
func TestEinLesefehlerNachDemAbgleichLaesstDieUebernahmeStehen(t *testing.T) {
	quelle := mse2(t)
	ziel := t.TempDir()
	katalog := &ablage{lesefehler: errors.New("der Katalog schweigt jetzt")}

	verwalter := service.NeuerPaketverwalter(katalog, ziel, ziel)

	ergebnis, err := verwalter.Uebernehmen(t.Context(), quelle)
	if err != nil {
		t.Fatalf("übernehmen: %v", err)
	}

	if len(ergebnis.Uebernommen) != 0 {
		t.Errorf("übernommen: %+v", ergebnis.Uebernommen)
	}

	if _, fehler := os.Stat(filepath.Join(ziel, "magic.mse-game")); fehler != nil {
		t.Errorf("die Kopie ist verschwunden: %v", fehler)
	}
}

// **Ein Verzeichnis ohne Leserecht wird benannt.**
//
// `os.Stat` glückt, `os.ReadDir` nicht — der Fall liegt zwischen den beiden
// Prüfungen und wäre sonst ein Zweig, den kein Test betritt (§3).
func TestEinUnlesbaresVerzeichnisWirdBenannt(t *testing.T) {
	gesperrt := filepath.Join(t.TempDir(), "gesperrt")
	if err := os.MkdirAll(gesperrt, 0o000); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	t.Cleanup(func() { _ = os.Chmod(gesperrt, 0o700) }) //nolint:errcheck,gosec // nur zum Aufräumen

	verwalter := service.NeuerPaketverwalter(&ablage{}, t.TempDir(), t.TempDir())

	if _, err := verwalter.Uebernehmen(t.Context(), gesperrt); err == nil {
		t.Error("kein Fehler")
	}
}

// **Das verwaltete Verzeichnis wird gelesen, auch wenn es beim Start leer war.**
//
// # Der Befund
//
// [service.Paketverwalter] liest über die Wurzeln, die ihm beim Bauen gegeben
// wurden. In der Anwendung kommen sie aus `msepaket.Paketwurzeln`, und das
// meldet ein Verzeichnis nur, wenn **schon ein Paket darin liegt**. Auf einem
// frischen Rechner ist `~/Library/…/pakete/` leer — und damit keine Wurzel.
//
// `Installieren` und `Uebernehmen` kopieren dorthin und lesen danach über
// **diese** Wurzeln: Ergebnis leer, Meldung „geglückt". Erst der nächste
// Prozess sah das Paket, weil das Verzeichnis dann nicht mehr leer war.
//
// Gemessen am 2026-08-30 an `mse mse2-uebernehmen`
// (#1291):
// `{"uebernommen": [], "uebergangen": [], "befunde": []}` — und im Katalog
// stand das Paket.
//
// # Warum die vorhandenen Läufe es nicht zeigen
//
// Weil sie das Ziel **als Wurzel mitgeben** (`NeuerPaketverwalter(katalog,
// ziel, ziel)`). Das ist die Vorrichtung, die großzügiger ist als der Betrieb
// (§3): sie prüft eine Zusage, die der Aufrufer in der Anwendung nicht
// einlöst.
func TestDieUebernahmeLiestDasVerwalteteVerzeichnisAuchOhneWurzel(t *testing.T) {
	quelle := mse2(t)
	ziel := t.TempDir()
	katalog := &ablage{}

	// **Ohne Wurzeln** — genau das, was `Paketwurzeln` für ein leeres
	// verwaltetes Verzeichnis gibt.
	verwalter := service.NeuerPaketverwalter(katalog, ziel)

	ergebnis, err := verwalter.Uebernehmen(t.Context(), quelle)
	if err != nil {
		t.Fatalf("übernehmen: %v", err)
	}

	if len(ergebnis.Uebernommen) != 2 {
		t.Errorf("übernommen: %+v", ergebnis.Uebernommen)
	}
}

// **Und dasselbe beim Einbinden eines einzelnen Pakets.**
func TestDasEinbindenLiestDasVerwalteteVerzeichnisAuchOhneWurzel(t *testing.T) {
	quelle := mse2(t)
	ziel := t.TempDir()

	verwalter := service.NeuerPaketverwalter(&ablage{}, ziel)

	paket, err := verwalter.Installieren(t.Context(),
		filepath.Join(quelle, "magic-m15.mse-style"))
	if err != nil {
		t.Fatalf("einbinden: %v", err)
	}

	if paket.Name != "magic-m15" {
		t.Errorf("eingebunden wurde %+v", paket)
	}
}
