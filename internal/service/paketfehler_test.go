package service_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Die Fehler- und Sonderwege des Paketverwalters.
//
// Sie sind der Normalfall und nicht die Ausnahme (§8): fremde Verzeichnisse,
// fehlende Rechte, ein Set zwischen den Vorlagen.

// **Ein unlesbares Paketverzeichnis ist ein Befund, kein Abbruch.**
func TestEinUnlesbaresPaketverzeichnisWirdBenannt(t *testing.T) {
	gesperrt := filepath.Join(t.TempDir(), "gesperrt")
	if err := os.MkdirAll(gesperrt, 0o000); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	t.Cleanup(func() { _ = os.Chmod(gesperrt, 0o700) }) //nolint:errcheck,gosec // nur zum Aufräumen

	katalog := &ablage{}
	verwalter := service.NeuerPaketverwalter(katalog, t.TempDir(), gesperrt)

	befunde, err := verwalter.Auffrischen(t.Context())
	if err != nil {
		t.Fatalf("auffrischen: %v", err)
	}

	if len(befunde) != 1 {
		t.Fatalf("Befunde: %v", befunde)
	}
}

// **Eine leere Wurzel wird übergangen.**
//
// `app.Saatverzeichnis()` gibt einen leeren Pfad, solange die Anwendung ohne
// Bündel läuft; `filepath.Join("", …)` machte daraus einen **relativen** Pfad,
// und gelesen würde das Arbeitsverzeichnis des Prozesses.
func TestEineLeereWurzelWirdUebergangen(t *testing.T) {
	verwalter := service.NeuerPaketverwalter(&ablage{}, t.TempDir(), "", "  ")

	befunde, err := verwalter.Auffrischen(t.Context())
	if err != nil {
		t.Fatalf("auffrischen: %v", err)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde: %v", befunde)
	}
}

// **Ein Set steht nicht im Paketkatalog**, und ein Verzeichnis ohne
// Paketendung erst recht nicht.
func TestSetsUndFremdesStehenNichtImKatalog(t *testing.T) {
	wurzel := t.TempDir()
	paket(t, wurzel, "irgendwas.mse-set", "mse version: 2.0.0\ngame: probe\n")

	if err := os.MkdirAll(filepath.Join(wurzel, "keinpaket"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	katalog := &ablage{}
	verwalter := service.NeuerPaketverwalter(katalog, t.TempDir(), wurzel)

	if _, err := verwalter.Auffrischen(t.Context()); err != nil {
		t.Fatalf("auffrischen: %v", err)
	}

	if len(katalog.pakete) != 0 {
		t.Errorf("Katalog: %+v", katalog.pakete)
	}
}

// **Dasselbe Paket in zwei Wurzeln zählt einmal — das erste gewinnt.**
//
// Die Reihenfolge der Wurzeln ist die Aussage: wer ein Paket bezogen hat, meint
// das bezogene; die mitgelieferte Saat kommt danach (#58).
func TestDasErsteVorkommenGewinnt(t *testing.T) {
	erste := t.TempDir()
	zweite := t.TempDir()
	paket(t, erste, "doppelt.mse-style", "mse version: 2.0.0\nversion: 2026-01-01\n")
	paket(t, zweite, "doppelt.mse-style", "mse version: 2.0.0\nversion: 2020-01-01\n")

	katalog := &ablage{}
	verwalter := service.NeuerPaketverwalter(katalog, t.TempDir(), erste, zweite)

	if _, err := verwalter.Auffrischen(t.Context()); err != nil {
		t.Fatalf("auffrischen: %v", err)
	}

	if len(katalog.pakete) != 1 || katalog.pakete[0].Version != "2026-01-01" {
		t.Errorf("Katalog: %+v", katalog.pakete)
	}
}

// **Ein Fehler des Bestands kommt durch** — bei jedem der drei Wege.
func TestDerBestandsfehlerKommtDurch(t *testing.T) {
	kaputt := errors.New("der Bestand schweigt")
	quelle := t.TempDir()
	paket(t, quelle, "neu.mse-style", "mse version: 2.0.0\n")

	verwalter := service.NeuerPaketverwalter(&ablage{fehler: kaputt}, t.TempDir(), quelle)

	if _, err := verwalter.Auffrischen(t.Context()); !errors.Is(err, kaputt) {
		t.Errorf("auffrischen: %v", err)
	}

	if _, err := verwalter.Installieren(t.Context(),
		filepath.Join(quelle, "neu.mse-style")); !errors.Is(err, kaputt) {
		t.Errorf("installieren: %v", err)
	}

	if err := verwalter.Entfernen(t.Context(), "mse-style", "neu", ""); !errors.Is(err, kaputt) {
		t.Errorf("entfernen: %v", err)
	}
}

// **Ein Set lässt sich nicht als Vorlage installieren.**
//
// Es trägt eine der sieben Endungen und ist trotzdem keine Vorlage — der
// Bestand führt es in `sets`.
func TestEinSetLaesstSichNichtInstallieren(t *testing.T) {
	quelle := t.TempDir()
	paket(t, quelle, "meins.mse-set", "mse version: 2.0.0\ngame: probe\n")

	verwalter := service.NeuerPaketverwalter(&ablage{}, t.TempDir(), t.TempDir())

	_, err := verwalter.Installieren(t.Context(), filepath.Join(quelle, "meins.mse-set"))
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("kein Validierungsfehler: %v", err)
	}
}

// **Ein Paket, das sich nicht öffnen lässt, landet gar nicht erst im
// Verzeichnis.**
func TestEinUnlesbaresPaketWirdNichtInstalliert(t *testing.T) {
	quelle := t.TempDir()
	if err := os.MkdirAll(filepath.Join(quelle, "leer.mse-style"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	ziel := t.TempDir()
	verwalter := service.NeuerPaketverwalter(&ablage{}, ziel, ziel)

	if _, err := verwalter.Installieren(t.Context(),
		filepath.Join(quelle, "leer.mse-style")); err == nil {
		t.Fatal("kein Fehler")
	}

	if _, fehler := os.Stat(filepath.Join(ziel, "leer.mse-style")); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("trotzdem kopiert: %v", fehler)
	}
}

// **Ein Archiv wird abgewiesen, und die Absage nennt den Grund.**
//
// [msepaket.Oeffne] läse es; [msepaket.Paketwurzeln] macht ein Archiv aber
// nicht zur Wurzel — installiert wäre es unsichtbar.
func TestEinArchivWirdAbgewiesen(t *testing.T) {
	quelle := t.TempDir()
	archiv := filepath.Join(quelle, "gepackt.mse-style")

	roh, err := os.ReadFile(filepath.Join("..", "..", "testdaten", "gepackt.mse-style"))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if fehler := os.WriteFile(archiv, roh, 0o600); fehler != nil { //nolint:gosec // der Pfad entsteht aus t.TempDir()
		t.Fatalf("schreiben: %v", fehler)
	}

	ziel := t.TempDir()
	verwalter := service.NeuerPaketverwalter(&ablage{}, ziel, ziel)

	_, fehler := verwalter.Installieren(t.Context(), archiv)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("kein Validierungsfehler: %v", fehler)
	}
}

// **Was sich nicht kopieren lässt, hinterlässt kein halbes Paket.**
func TestEinGescheitertesKopierenRaeumtHinterSich(t *testing.T) {
	quelle := t.TempDir()
	paket(t, quelle, "sperrig.mse-style", "mse version: 2.0.0\n")

	gesperrt := filepath.Join(quelle, "sperrig.mse-style", "unlesbar")
	if err := os.WriteFile(gesperrt, []byte("geheim"), 0o000); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	ziel := t.TempDir()
	verwalter := service.NeuerPaketverwalter(&ablage{}, ziel, ziel)

	if _, err := verwalter.Installieren(t.Context(),
		filepath.Join(quelle, "sperrig.mse-style")); err == nil {
		t.Fatal("kein Fehler")
	}

	if _, fehler := os.Stat(filepath.Join(ziel, "sperrig.mse-style")); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("halbes Paket geblieben: %v", fehler)
	}
}

// **Eine Quelle, die es nicht gibt, wird benannt.**
func TestEineFehlendeQuelleWirdBenannt(t *testing.T) {
	ziel := t.TempDir()
	verwalter := service.NeuerPaketverwalter(&ablage{}, ziel, ziel)

	_, err := verwalter.Installieren(t.Context(), filepath.Join(t.TempDir(), "weg.mse-style"))
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("kein Validierungsfehler: %v", err)
	}
}

// **Ein Pfad, der sich nicht ins Verzeichnis rechnen lässt, gilt als außerhalb.**
//
// `filepath.Rel` scheitert, wenn das eine fest und das andere relativ ist. Der
// Fall entsteht, wenn jemand den Verwalter über einen relativen Ort baut — und
// dann ist „außerhalb" die sichere Antwort: nicht entfernen.
func TestEinUnvergleichbarerPfadGiltAlsAusserhalb(t *testing.T) {
	katalog := &ablage{pakete: []domain.Paket{
		{Art: "mse-style", Name: "irgendwo", Pfad: "/fest/irgendwo.mse-style"},
	}}
	verwalter := service.NeuerPaketverwalter(katalog, "relativ", "relativ")

	if err := verwalter.Entfernen(t.Context(), "mse-style", "irgendwo", ""); err == nil {
		t.Error("entfernt, obwohl der Ort nicht vergleichbar ist")
	}
}

// **Was sich nicht von der Platte nehmen lässt, bleibt im Katalog.**
func TestEinGescheitertesRaeumenBleibtImKatalog(t *testing.T) {
	verwaltet := t.TempDir()
	paket(t, verwaltet, "fest.mse-style", "mse version: 2.0.0\n")

	if err := os.Chmod(verwaltet, 0o500); err != nil { //nolint:gosec // ein schreibgeschütztes Verzeichnis ist genau der Fall
		t.Fatalf("Rechte: %v", err)
	}

	t.Cleanup(func() { _ = os.Chmod(verwaltet, 0o700) }) //nolint:errcheck,gosec // nur zum Aufräumen

	katalog := &ablage{pakete: []domain.Paket{
		{Art: "mse-style", Name: "fest", Pfad: filepath.Join(verwaltet, "fest.mse-style")},
	}}
	verwalter := service.NeuerPaketverwalter(katalog, verwaltet, verwaltet)

	if err := verwalter.Entfernen(t.Context(), "mse-style", "fest", ""); err == nil {
		t.Fatal("kein Fehler")
	}

	if len(katalog.entfernt) > 0 {
		t.Errorf("trotzdem aus dem Katalog genommen: %v", katalog.entfernt)
	}
}

// **Ein Fehler beim Ausbuchen kommt durch.**
//
// Die Platte ist geräumt, der Katalog nicht — und wer das verschwiege, zeigte
// ein Paket, das es nicht mehr gibt.
func TestEinFehlerBeimAusbuchenKommtDurch(t *testing.T) {
	verwaltet := t.TempDir()
	paket(t, verwaltet, "weg.mse-style", "mse version: 2.0.0\n")

	kaputt := errors.New("der Katalog schweigt")
	katalog := &ablage{
		pakete: []domain.Paket{
			{Art: "mse-style", Name: "weg", Pfad: filepath.Join(verwaltet, "weg.mse-style")},
		},
		entfernfehler: kaputt,
	}

	verwalter := service.NeuerPaketverwalter(katalog, verwaltet, verwaltet)

	if err := verwalter.Entfernen(t.Context(), "mse-style", "weg", ""); !errors.Is(err, kaputt) {
		t.Errorf("Fehler %v", err)
	}
}

// **Ein Verwaltungsort, der keiner ist, wird abgewiesen — auf beiden Wegen.**
//
// # Woher der Fall kommt
//
// Aus dem Zweitblick über Phase 7. `App.Starten` hebt den Fehler von
// `Paketverzeichnis()` auf und baut den Verwalter trotzdem — mit leerem Ort.
// `filepath.Join("", name)` wird daraus ein **relativer** Pfad: installiert
// würde ins Arbeitsverzeichnis des Prozesses.
//
// Die Korrektur beim Paketbezieher (Phase 9) hat den Fall behandelt, nicht die
// Klasse.
func TestEinLoserVerwaltungsortWirdAbgewiesen(t *testing.T) {
	quelle := t.TempDir()
	paket(t, quelle, "neu.mse-style", "mse version: 2.0.0\n")

	for name, ort := range map[string]string{"leer": "", "relativ": "pakete"} {
		t.Run(name, func(t *testing.T) {
			verwalter := service.NeuerPaketverwalter(&ablage{}, ort, quelle)

			if _, err := verwalter.Installieren(t.Context(),
				filepath.Join(quelle, "neu.mse-style")); !errors.Is(err, domain.ErrValidation) {
				t.Errorf("installieren: %v", err)
			}

			if _, err := verwalter.Uebernehmen(t.Context(), quelle); !errors.Is(err, domain.ErrValidation) {
				t.Errorf("übernehmen: %v", err)
			}
		})
	}

	// **Die Gegenprobe:** mit festem Ort geht beides.
	fest := t.TempDir()
	if _, err := service.NeuerPaketverwalter(&ablage{}, fest, quelle).
		Installieren(t.Context(), filepath.Join(quelle, "neu.mse-style")); err != nil {
		t.Errorf("mit festem Ort: %v", err)
	}
}
