package service_test

// **Neuere Paketstände sichtbar und sicher aktualisieren** — G04
// (#2251).
//
// Ein bezogener Baum steht an einem Commit. Was hier geprüft wird: dass
// der bezogene, der geprüfte und der neueste Stand der Quelle unterscheidbar
// sind — und ein Fehler beim Fragen ein eigener Zustand ist; dass eine
// Aktualisierung vorher zeigt, welche Pakete und Sets sie berührt, und eine
// Bestätigung verlangt, wo sie etwas nehmen könnte; dass der bisherige Stand
// zum Zurückholen bleibt; und dass **kein Zweig** nachgeführt wird.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/bezug"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

const (
	m15Adresse = "https://github.com/MagicSetEditorPacks/M15-Magic-Pack"
	m15Stand   = "3f7f8ceed0c1fcc0d193d752c646aaaea7bece85"
	m15Spitze  = "c8e38c9b138c5e9c66199fd099a353bce6ff5ec7"
)

// auskunftDouble beantwortet die Fragen an die Quelle — ohne Netz.
type auskunftDouble struct {
	spitze           bezug.Spitzenstand
	spitzenfehler    error
	vergleich        bezug.Vergleich
	vergleichsfehler error
	gefragt          []string
}

func (a *auskunftDouble) Spitze(_ context.Context, adresse string) (bezug.Spitzenstand, error) {
	a.gefragt = append(a.gefragt, "spitze "+adresse)

	return a.spitze, a.spitzenfehler
}

func (a *auskunftDouble) Vergleich(_ context.Context, adresse, von, bis string) (bezug.Vergleich, error) {
	a.gefragt = append(a.gefragt, "vergleich "+adresse+" "+von+"..."+bis)

	return a.vergleich, a.vergleichsfehler
}

// setkoepfe nennt die Sets dieses Profils.
type bezugssets []domain.SetHead

func (s bezugssets) Setkoepfe(context.Context) ([]domain.SetHead, error) { return s, nil }

func aktualisierungsplatz(t *testing.T) (*service.Paketaktualisierung, *auskunftDouble, string) {
	t.Helper()

	return aktualisierungsplatzVon(t, m15Adresse)
}

// aktualisierungsplatzVon legt einen bezogenen Baum an, der von `adresse`
// stammt — für einen Bezug gegen einen Testserver.
func aktualisierungsplatzVon(t *testing.T, adresse string) (*service.Paketaktualisierung, *auskunftDouble, string) {
	t.Helper()

	verwaltet := filepath.Join(t.TempDir(), "pakete")
	baum := filepath.Join(verwaltet, "M15-Magic-Pack", "data")
	if fehler := os.MkdirAll(baum, 0o750); fehler != nil {
		t.Fatal(fehler)
	}

	herkunft := adresse + "@" + m15Stand
	katalog := &ablage{pakete: []domain.Paket{
		{Art: "mse-game", Name: "magic", Pfad: filepath.Join(baum, "magic.mse-game"), Herkunft: herkunft, Pruefsumme: "sha256:alt"},
		{Art: "mse-style", Name: "magic-m15", Pfad: filepath.Join(baum, "magic-m15.mse-style"), Herkunft: herkunft, Pruefsumme: "sha256:alt"},
		{Art: "mse-style", Name: "eigen", Pfad: filepath.Join(verwaltet, "eigen.mse-style")},
	}}

	verwalter := service.NeuerPaketverwalter(katalog, verwaltet).MitSetkoepfen(bezugssets{
		{ID: 1, Name: "Mein Set", Game: "magic", Style: "m15"},
		{ID: 2, Name: "Fremdes Spiel", Game: "vs", Style: "standard"},
	})
	auskunft := &auskunftDouble{}

	return service.NeuePaketaktualisierung(verwalter, service.NeuerPaketbezieher(verwaltet, "probe"), auskunft),
		auskunft, verwaltet
}

func TestDieBezogenenStaendeKommenOhneFrage(t *testing.T) {
	aktualisierung, auskunft, verwaltet := aktualisierungsplatz(t)

	staende, fehler := aktualisierung.Staende(t.Context(), false)
	if fehler != nil {
		t.Fatalf("Staende: %v", fehler)
	}

	if len(staende) != 1 {
		t.Fatalf("%d Stände, erwartet war einer (ein einzeln eingebundenes Paket ist kein Bezug): %+v", len(staende), staende)
	}

	stand := staende[0]
	if stand.Name != "M15-Magic-Pack" || stand.Ort != filepath.Join(verwaltet, "M15-Magic-Pack") ||
		stand.Adresse != m15Adresse || stand.Stand != m15Stand || stand.Pakete != 2 {
		t.Errorf("Stand %+v", stand)
	}

	// **Der geprüfte Stand kommt aus der Liste der Quellen** — gegen ihn ist
	// dieses Projekt gelaufen (`bezug.Quellen`).
	if stand.Geprueft != m15Stand || stand.Lage != service.BezugUngefragt || stand.Vorige {
		t.Errorf("geprüft %q, Lage %q, vorige %v", stand.Geprueft, stand.Lage, stand.Vorige)
	}

	if len(auskunft.gefragt) != 0 {
		t.Errorf("ohne Frage wurde gefragt: %v", auskunft.gefragt)
	}
}

func TestDieLageDerQuelleIstUnterscheidbar(t *testing.T) {
	faelle := []struct {
		name      string
		auskunft  auskunftDouble
		lage      service.Bezugslage
		vorsprung int
		fehler    bool
	}{
		{"die Spitze ist der bezogene Stand", auskunftDouble{spitze: bezug.Spitzenstand{Commit: m15Stand}}, service.BezugAktuell, 0, false},
		{"die Quelle ist voraus", auskunftDouble{
			spitze:    bezug.Spitzenstand{Commit: m15Spitze},
			vergleich: bezug.Vergleich{Lage: "ahead", Vorsprung: 63},
		}, service.BezugNeuer, 63, false},
		{"die Quelle ist nicht voraus", auskunftDouble{
			spitze:    bezug.Spitzenstand{Commit: m15Spitze},
			vergleich: bezug.Vergleich{Lage: "diverged", Vorsprung: 2, Rueckstand: 5},
		}, service.BezugAbweichend, 2, false},
		{"ohne Netz", auskunftDouble{spitzenfehler: domain.ErrNichtErreichbar}, service.BezugUnbekannt, 0, true},
		{"der Vergleich scheitert", auskunftDouble{
			spitze: bezug.Spitzenstand{Commit: m15Spitze}, vergleichsfehler: domain.ErrKeineAntwort,
		}, service.BezugUnbekannt, 0, true},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			aktualisierung, auskunft, _ := aktualisierungsplatz(t)
			*auskunft = fall.auskunft

			staende, fehler := aktualisierung.Staende(t.Context(), true)
			if fehler != nil {
				t.Fatalf("Staende: %v", fehler)
			}

			stand := staende[0]
			if stand.Lage != fall.lage || stand.Vorsprung != fall.vorsprung || (stand.Fehler != nil) != fall.fehler {
				t.Errorf("Lage %q, Vorsprung %d, Fehler %v", stand.Lage, stand.Vorsprung, stand.Fehler)
			}

			if fall.fehler && !errors.Is(stand.Fehler, domain.ErrGegenstelle) {
				t.Errorf("der Fehler verlor seine Klasse: %v", stand.Fehler)
			}
		})
	}
}

func TestDieVorschauNenntPaketeUndSets(t *testing.T) {
	aktualisierung, auskunft, _ := aktualisierungsplatz(t)
	auskunft.vergleich = bezug.Vergleich{Lage: "ahead", Vorsprung: 3, Dateien: []bezug.Aenderung{
		{Datei: "data/magic-m15.mse-style/style", Art: "modified"},
		{Datei: "data/magic-m15.mse-style/card.png", Art: "added"},
		{Datei: "data/magic-neu.mse-style/style", Art: "added"},
		{Datei: "README.md", Art: "modified"},
	}}

	vorschau, fehler := aktualisierung.Vorschau(t.Context(), "M15-Magic-Pack", m15Spitze)
	if fehler != nil {
		t.Fatalf("Vorschau: %v", fehler)
	}

	if vorschau.Von != m15Stand || vorschau.Zu != m15Spitze || vorschau.Vorsprung != 3 {
		t.Errorf("Vorschau %+v", vorschau)
	}

	namen := make([]string, 0, len(vorschau.Pakete))
	for _, paket := range vorschau.Pakete {
		namen = append(namen, paket.Paket)
	}

	if !slices.Equal(namen, []string{"magic-m15.mse-style", "magic-neu.mse-style"}) || vorschau.Uebrige != 1 {
		t.Errorf("Pakete %v, übrige %d", namen, vorschau.Uebrige)
	}

	if !slices.Equal(vorschau.Sets, []string{"Mein Set"}) {
		t.Errorf("betroffene Sets %v", vorschau.Sets)
	}

	if vorschau.BestaetigungNoetig || len(vorschau.Gruende) != 0 {
		t.Errorf("ohne Wegfall und vorwärts braucht es keine Bestätigung: %v", vorschau.Gruende)
	}
}

func TestDieVorschauVerlangtWoEtwasWegfallenKann(t *testing.T) {
	faelle := []struct {
		name      string
		vergleich bezug.Vergleich
		grund     string
	}{
		{"ein Paket, das ein Set nennt, verliert Dateien", bezug.Vergleich{Lage: "ahead", Dateien: []bezug.Aenderung{
			{Datei: "data/magic.mse-game/game", Art: "removed"},
		}}, service.GrundSetVerliert},
		{"die Quelle ist nicht voraus", bezug.Vergleich{Lage: "behind"}, service.GrundNichtVoraus},
		{"die Gegenstelle nennt nicht alle Dateien", bezug.Vergleich{Lage: "ahead", Gekuerzt: true}, service.GrundGekuerzt},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			aktualisierung, auskunft, _ := aktualisierungsplatz(t)
			auskunft.vergleich = fall.vergleich

			vorschau, fehler := aktualisierung.Vorschau(t.Context(), "M15-Magic-Pack", m15Spitze)
			if fehler != nil {
				t.Fatalf("Vorschau: %v", fehler)
			}

			if !vorschau.BestaetigungNoetig || !slices.Contains(vorschau.Gruende, fall.grund) {
				t.Errorf("Bestätigung %v, Gründe %v — erwartet %q", vorschau.BestaetigungNoetig, vorschau.Gruende, fall.grund)
			}
		})
	}
}

func TestDieVorschauNimmtNurEinenCommit(t *testing.T) {
	aktualisierung, auskunft, _ := aktualisierungsplatz(t)

	_, fehler := aktualisierung.Vorschau(t.Context(), "M15-Magic-Pack", "main")

	var feldfehler *domain.FieldError
	if !errors.As(fehler, &feldfehler) || feldfehler.Field != "stand" {
		t.Errorf("Fehler %v, erwartet am Feld stand", fehler)
	}

	if len(auskunft.gefragt) != 0 {
		t.Errorf("mit einem Zweig wurde gefragt: %v", auskunft.gefragt)
	}

	if _, fehler := aktualisierung.Vorschau(t.Context(), "Gibt-es-nicht", m15Spitze); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("ein unbekannter Bezug: %v", fehler)
	}
}
