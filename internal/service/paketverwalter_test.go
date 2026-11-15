package service_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// ablage ist ein Katalog im Arbeitsspeicher — die Naht zu `storage`.
type ablage struct {
	pakete   []domain.Paket
	fehler   error
	entfernt []string

	// musterfragen sind die `match:`-Muster, nach denen gefragt wurde — damit
	// ein Test sagen kann, dass **nicht** gefragt wurde (#634).
	musterfragen []string

	// entfernfehler scheitert **nur** beim Ausbuchen.
	//
	// Ein eigener Wert und nicht `fehler`: wer beides zusammenwirft, kommt gar
	// nicht bis dorthin — die Liste scheiterte vorher, und der Weg dahinter
	// bliebe ungeprüft.
	entfernfehler error

	// lesefehler scheitert **nur** beim Lesen der Liste — für den Weg nach
	// einem geglückten Abgleich.
	lesefehler error
}

func (a *ablage) Abgleichen(ctx context.Context, pakete []domain.Paket) error {
	// **Der Abbruch wird beachtet.** Bis zum 2026-08-28 stand hier `_`, und
	// damit war jede Zusage über abgebrochene Läufe unprüfbar — angestrichen
	// vom Nachlauf über Phase 3.8
	// (#1169,
	// Befund 1).
	if fehler := ctx.Err(); fehler != nil {
		return fehler
	}

	if a.fehler != nil {
		return a.fehler
	}

	a.pakete = pakete

	return nil
}

// PaketeNachMuster beantwortet ein `match:`-Muster über der eigenen Liste.
//
// **Dasselbe Ergebnis wie im Bestand, ohne SQL**: Endung gleich, Name gegen
// das Muster, sortiert nach Rang und dann Namen (#632).
func (a *ablage) PaketeNachMuster(
	_ context.Context, muster string,
) ([]domain.Paket, error) {
	punkt := strings.LastIndex(muster, ".")
	if punkt < 0 {
		return nil, fmt.Errorf("%w: das Muster %q nennt keine Paketart",
			domain.ErrValidation, muster)
	}

	a.musterfragen = append(a.musterfragen, muster)

	namensmuster, art := muster[:punkt], muster[punkt+1:]

	getroffen := make([]domain.Paket, 0)

	for _, paket := range a.pakete {
		passt, fehler := filepath.Match(namensmuster, paket.Name)
		if fehler != nil {
			return nil, fehler
		}

		if passt && paket.Art == art {
			getroffen = append(getroffen, paket)
		}
	}

	sort.SliceStable(getroffen, func(eine, andere int) bool {
		if getroffen[eine].Rang != getroffen[andere].Rang {
			return getroffen[eine].Rang < getroffen[andere].Rang
		}

		return getroffen[eine].Name < getroffen[andere].Name
	})

	return getroffen, a.fehler
}

func (a *ablage) Pakete(_ context.Context) ([]domain.Paket, error) {
	if a.fehler != nil {
		return nil, a.fehler
	}

	if a.lesefehler != nil {
		return nil, a.lesefehler
	}

	return a.pakete, nil
}

func (a *ablage) PaketEntfernen(ctx context.Context, art, name string) error {
	if fehler := ctx.Err(); fehler != nil {
		return fehler
	}

	if a.fehler != nil {
		return a.fehler
	}

	if a.entfernfehler != nil {
		return a.entfernfehler
	}

	a.entfernt = append(a.entfernt, art+"/"+name)

	gehalten := make([]domain.Paket, 0, len(a.pakete))

	for _, paket := range a.pakete {
		if paket.Art != art || paket.Name != name {
			gehalten = append(gehalten, paket)
		}
	}

	a.pakete = gehalten

	return nil
}

// **Der Katalog liest, was wirklich dasteht.**
//
// Aus der Saat des Repositoriums: ein Spiel und vier Stile, jeder mit Version
// und Abhängigkeit. Gegen die echten Dateien und nicht gegen nachgebaute
// Zeichenketten (§3) — dieselbe Regel wie bei `msepaket`.
func TestDerVerwalterLiestDieInstalliertenPakete(t *testing.T) {
	katalog := &ablage{}
	verwalter := service.NeuerPaketverwalter(katalog, t.TempDir(), festeSaatwurzel(t))

	befunde, err := verwalter.Auffrischen(t.Context())
	if err != nil {
		t.Fatalf("auffrischen: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	gefunden := map[string]domain.Paket{}
	for _, paket := range katalog.pakete {
		gefunden[paket.Name] = paket
	}

	spiel, da := gefunden["probe"]
	if !da {
		t.Fatalf("das Spielpaket fehlt: %v", katalog.pakete)
	}

	if spiel.Art != "mse-game" || spiel.Version != "2026-08-15" {
		t.Errorf("Spiel: %+v", spiel)
	}

	stil, da := gefunden["probe-schlicht"]
	if !da {
		t.Fatalf("der Stil fehlt: %v", katalog.pakete)
	}

	if len(stil.Braucht) != 1 || stil.Braucht[0].Paket != "probe.mse-game" {
		t.Errorf("Abhängigkeiten des Stils: %+v", stil.Braucht)
	}

	if stil.Pfad == "" {
		t.Error("der Stil führt keinen Ort")
	}
}

// **Ein kaputtes Paket ist ein Befund, kein Abbruch** (§8).
//
// Fremde Daten sind der Normalfall: ein Verzeichnis auf `.mse-style` ohne
// lesbare Hauptdatei liegt in jeder gewachsenen Sammlung. Wer daran abbräche,
// hätte für **kein** Paket einen Katalog.
func TestEinKaputtesPaketWirdBenanntUndDieAnderenBleiben(t *testing.T) {
	wurzel := t.TempDir()
	paket(t, wurzel, "gut.mse-game", "mse version: 2.0.0\nshort name: Gut\nversion: 2026-01-01\n")

	if err := os.MkdirAll(filepath.Join(wurzel, "kaputt.mse-style"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	katalog := &ablage{}
	verwalter := service.NeuerPaketverwalter(katalog, t.TempDir(), wurzel)

	befunde, err := verwalter.Auffrischen(t.Context())
	if err != nil {
		t.Fatalf("auffrischen: %v", err)
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet 1: %v", len(befunde), befunde)
	}

	if len(katalog.pakete) != 1 || katalog.pakete[0].Name != "gut" {
		t.Errorf("Katalog: %+v", katalog.pakete)
	}
}

// **Installieren und Entfernen sind umkehrbar** — die Zusage aus #54.
func TestInstallierenUndEntfernenSindUmkehrbar(t *testing.T) {
	quelle := t.TempDir()
	paket(t, quelle, "fremd.mse-style",
		"mse version: 2.0.0\ngame: probe\nshort name: Fremd\nversion: 2026-02-02\n")

	ziel := t.TempDir()
	katalog := &ablage{}
	verwalter := service.NeuerPaketverwalter(katalog, ziel, ziel)

	installiert, err := verwalter.Installieren(t.Context(), filepath.Join(quelle, "fremd.mse-style"))
	if err != nil {
		t.Fatalf("installieren: %v", err)
	}

	if installiert.Art != "mse-style" || installiert.Name != "fremd" {
		t.Fatalf("installiert: %+v", installiert)
	}

	// **Die Herkunft steht am Paket** (#56, §11): wer nicht weiß, woher eine
	// Vorlage kommt, kann ihre Rechtslage nicht einschätzen.
	if installiert.Herkunft != filepath.Join(quelle, "fremd.mse-style") {
		t.Errorf("Herkunft %q", installiert.Herkunft)
	}

	if _, fehler := os.Stat(filepath.Join(ziel, "fremd.mse-style", "style")); fehler != nil {
		t.Fatalf("nicht kopiert: %v", fehler)
	}

	if len(katalog.pakete) != 1 {
		t.Fatalf("Katalog nach dem Installieren: %+v", katalog.pakete)
	}

	if err := verwalter.Entfernen(t.Context(), "mse-style", "fremd", ""); err != nil {
		t.Fatalf("entfernen: %v", err)
	}

	if _, fehler := os.Stat(filepath.Join(ziel, "fremd.mse-style")); !errors.Is(fehler, os.ErrNotExist) {
		t.Errorf("noch auf der Platte: %v", fehler)
	}

	if len(katalog.pakete) != 0 {
		t.Errorf("Katalog nach dem Entfernen: %+v", katalog.pakete)
	}
}

// **Ein Paket, das schon dasteht, wird nicht stillschweigend überschrieben.**
//
// Wer eine Vorlage bearbeitet hat, verlöre sie sonst bei einer zweiten
// Installation — und nichts sagte es ihm. Ein Konflikt ist die Auskunft, dass
// zuerst zu entfernen ist.
func TestInstallierenMeldetEinVorhandenesPaket(t *testing.T) {
	quelle := t.TempDir()
	paket(t, quelle, "doppelt.mse-style", "mse version: 2.0.0\nshort name: Doppelt\n")

	ziel := t.TempDir()
	verwalter := service.NeuerPaketverwalter(&ablage{}, ziel, ziel)

	if _, err := verwalter.Installieren(t.Context(),
		filepath.Join(quelle, "doppelt.mse-style")); err != nil {
		t.Fatalf("erste Installation: %v", err)
	}

	_, err := verwalter.Installieren(t.Context(), filepath.Join(quelle, "doppelt.mse-style"))
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("kein Konflikt: %v", err)
	}
}

// **Was nicht im Verwaltungsverzeichnis liegt, wird nicht entfernt.**
//
// Die mitgelieferte Saat liegt im Bündel (#58). Ein „Entfernen", das dorthin
// griffe, löschte einen Teil der Anwendung — und der nächste Start stünde ohne
// Spiel und Stil da. Der Katalog führt sie trotzdem; er soll ja zeigen, was da
// ist.
func TestEntfernenGreiftNichtInsBuendel(t *testing.T) {
	// **Zwei Einträge, und der gesuchte ist der zweite.** Ein Katalog mit genau
	// einem Paket ließe die Schleife nie weitergehen — und der Fall „das ist
	// nicht das gemeinte" bliebe ungeprüft.
	katalog := &ablage{pakete: []domain.Paket{
		{Art: "mse-style", Name: "danebenliegend", Pfad: "/woanders/danebenliegend.mse-style"},
		{Art: "mse-game", Name: "probe", Pfad: filepath.Join(festeSaatwurzel(t), "probe.mse-game")},
	}}
	verwalter := service.NeuerPaketverwalter(katalog, t.TempDir(), festeSaatwurzel(t))

	err := verwalter.Entfernen(t.Context(), "mse-game", "probe", "")
	if err == nil {
		t.Fatal("die Saat ließ sich entfernen")
	}

	if _, fehler := os.Stat(filepath.Join(festeSaatwurzel(t), "probe.mse-game", "game")); fehler != nil {
		t.Fatalf("die Saat ist beschädigt: %v", fehler)
	}

	if len(katalog.entfernt) > 0 {
		t.Errorf("aus dem Katalog genommen: %v", katalog.entfernt)
	}
}

// **Ein Paket, das der Katalog nicht führt, lässt sich nicht entfernen.**
func TestEntfernenMeldetEinUnbekanntesPaket(t *testing.T) {
	verwalter := service.NeuerPaketverwalter(&ablage{}, t.TempDir(), t.TempDir())

	err := verwalter.Entfernen(t.Context(), "mse-style", "niegesehen", "")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("kein ErrNotFound: %v", err)
	}
}

// **Ein Verzeichnis, das kein Paket ist, wird abgewiesen.**
func TestInstallierenWeistEinNichtpaketAb(t *testing.T) {
	quelle := t.TempDir()
	if err := os.MkdirAll(filepath.Join(quelle, "irgendwas"), 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	verwalter := service.NeuerPaketverwalter(&ablage{}, t.TempDir(), t.TempDir())

	_, err := verwalter.Installieren(t.Context(), filepath.Join(quelle, "irgendwas"))
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("kein Validierungsfehler: %v", err)
	}
}

// **Der Katalog kommt aus dem Bestand, nicht von der Platte.**
//
// Die Oberfläche fragt ihn bei jedem Aufbau; über das Dateisystem zu gehen
// hieße, bei jeder Anzeige einen Baum abzulaufen — die Rechnung, die #187 an
// der Bildquelle aufgemacht hat.
func TestDieListeKommtAusDemBestand(t *testing.T) {
	katalog := &ablage{pakete: []domain.Paket{{Art: "mse-game", Name: "probe", Pfad: "/p"}}}
	verwalter := service.NeuerPaketverwalter(katalog, t.TempDir(), t.TempDir())

	liste, err := verwalter.Liste(t.Context())
	if err != nil {
		t.Fatalf("Liste: %v", err)
	}

	if len(liste) != 1 || liste[0].Name != "probe" {
		t.Errorf("Liste: %+v", liste)
	}
}

// paket schreibt ein Paketverzeichnis mit einer Hauptdatei.
func paket(t *testing.T, wurzel, name, inhalt string) {
	t.Helper()

	ort := filepath.Join(wurzel, name)
	if err := os.MkdirAll(ort, 0o750); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if err := os.WriteFile(filepath.Join(ort, hauptdateiname(name)), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("schreiben: %v", err)
	}
}

// hauptdateiname gibt die Hauptdatei zu einem Paketnamen — Endung ohne `mse-`.
func hauptdateiname(name string) string {
	punkt := len(name) - 1
	for punkt >= 0 && name[punkt] != '.' {
		punkt--
	}

	return name[punkt+len(".mse-"):]
}

// festeSaatwurzel gibt das Saatverzeichnis als festen Pfad.
//
// **Fest und nicht relativ**: der Verwalter vergleicht Orte (`imVerzeichnis`),
// und ein relativer Pfad verglich sich gegen einen festen nie.
func festeSaatwurzel(t *testing.T) string {
	t.Helper()

	wurzel, err := filepath.Abs(saatwurzel())
	if err != nil {
		t.Fatalf("Wurzel: %v", err)
	}

	return wurzel
}
