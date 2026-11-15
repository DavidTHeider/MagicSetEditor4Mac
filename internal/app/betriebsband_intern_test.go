package app

// **Das Betriebsband** — OPS02
// (#2348):
// je Lage genau ein Hinweis mit Ziel, kein Flattern an einer Schwelle, und
// was sich nicht fragen ließ, steht als ungemessen da.

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

const gib = uint64(1) << 30

// quellenMit baut Quellen mit festen Antworten.
func quellenMit(frei, gesamt uint64, sicherung SicherungsstandDTO, dienst domain.Skriptdienstlage, neu int) betriebsquellen {
	return betriebsquellen{
		platz:     func() (uint64, uint64, error) { return frei, gesamt, nil },
		sicherung: func(context.Context) (SicherungsstandDTO, error) { return sicherung, nil },
		dienst:    func(context.Context) (domain.Skriptdienstlage, error) { return dienst, nil },
		neustarts: func() int { return neu },
	}
}

func hinweisarten(antwort BetriebsbandAntwort) []string {
	namen := make([]string, 0, len(antwort.Hinweise))
	for _, hinweis := range antwort.Hinweise {
		namen = append(namen, hinweis.Art)
	}

	return namen
}

// **Jede der vier Lagen gibt genau einen Hinweis, und jeder führt hin.**
func TestDasBetriebsbandNenntJedeLageMitZiel(t *testing.T) {
	t.Parallel()

	var band betriebsgedaechtnis

	alt := SicherungsstandDTO{Stufe: string(domain.SicherungsstufeAlt), Tage: 20}
	beendet := domain.Skriptdienstlage{Eingerichtet: true, LetzterAusgang: 1, Laeufe: 3}

	antwort := band.abrufen(t.Context(), quellenMit(gib, 100*gib, alt, beendet, 2))

	erwartet := []string{"platzKnapp", "sicherungAlt", "dienstBeendet", "webprozessNeu"}
	if !slices.Equal(hinweisarten(antwort), erwartet) || len(antwort.Ungemessen) != 0 || antwort.Fehler != nil {
		t.Fatalf("Antwort %+v, erwartet %v", antwort, erwartet)
	}

	ziele := map[string]string{
		"platzKnapp": domain.ZielDatenUndSicherung, "sicherungAlt": domain.ZielDatenUndSicherung,
		"dienstBeendet": domain.ZielDiagnose, "webprozessNeu": domain.ZielDiagnose,
	}
	werte := map[string]string{"platzKnapp": "frei", "sicherungAlt": "tage", "dienstBeendet": "ausgang", "webprozessNeu": "anzahl"}

	for _, hinweis := range antwort.Hinweise {
		if hinweis.Ziel != ziele[hinweis.Art] || hinweis.Werte[werte[hinweis.Art]] == "" {
			t.Errorf("Hinweis %+v", hinweis)
		}
	}

	if leer := band.abrufen(t.Context(), quellenMit(100*gib, 500*gib, SicherungsstandDTO{Stufe: "frisch"}, domain.Skriptdienstlage{}, 0)); len(leer.Hinweise) != 0 {
		t.Errorf("ohne Lage %+v", leer)
	}
}

// **Ein Lesefehler der Sicherung nimmt ihren Altershinweis nicht** (Zweitblick
// über Phase 4, Befund 10): Platz und Dienst hatten ein Gedächtnis, das Alter
// der Sicherung nicht — der Hinweis verschwand ohne Beleg einer neuen.
func TestEinLesefehlerDerSicherungBehaeltIhrAlter(t *testing.T) {
	t.Parallel()

	var band betriebsgedaechtnis

	q := quellenMit(100*gib, 500*gib, SicherungsstandDTO{Stufe: string(domain.SicherungsstufeAlt), Tage: 20}, domain.Skriptdienstlage{}, 0)
	if antwort := band.abrufen(t.Context(), q); !slices.Equal(hinweisarten(antwort), []string{"sicherungAlt"}) {
		t.Fatalf("vorher %+v", antwort)
	}

	q.sicherung = func(context.Context) (SicherungsstandDTO, error) { return SicherungsstandDTO{}, errors.New("belegt") }

	if antwort := band.abrufen(t.Context(), q); !slices.Equal(hinweisarten(antwort), []string{"sicherungAlt"}) ||
		antwort.Hinweise[0].Werte["tage"] != "20" || !slices.Contains(antwort.Ungemessen, ungemessenSicherung) {
		t.Errorf("nach dem Lesefehler %+v", antwort)
	}

	q.sicherung = func(context.Context) (SicherungsstandDTO, error) { return SicherungsstandDTO{Stufe: "frisch"}, nil }

	if antwort := band.abrufen(t.Context(), q); len(antwort.Hinweise) != 0 {
		t.Errorf("nach einer neuen Sicherung %+v", antwort)
	}
}

// **Pendelt der Platz um die Schwelle, bleibt der Hinweis stehen** — und ein
// Dienst, der immer wieder abstürzt, nimmt ihn nicht in seinen
// Millisekunden zurück.
func TestDasBetriebsbandFlattertNicht(t *testing.T) {
	t.Parallel()

	// **Je Abruf ein Takt** — erst nach der Erholfrist gilt ein Lauf als
	// durchgelaufen.
	jetzt := time.Now()
	band := betriebsgedaechtnis{uhr: func() time.Time { return jetzt }}

	frisch := SicherungsstandDTO{Stufe: "frisch"}
	abgestuerzt := domain.Skriptdienstlage{Eingerichtet: true, LetzterAusgang: 1, Laeufe: 4}
	kurzWieder := domain.Skriptdienstlage{Eingerichtet: true, Laeuft: true, LetzterAusgang: 1, Laeufe: 5}

	folge := []struct {
		frei     uint64
		dienst   domain.Skriptdienstlage
		erwartet []string
	}{
		{gib + gib/2, abgestuerzt, []string{"platzKnapp", "dienstBeendet"}},
		{2*gib + gib/4, kurzWieder, []string{"platzKnapp", "dienstBeendet"}},
		{2*gib + gib/4, kurzWieder, nil},
		{gib + gib/2, kurzWieder, []string{"platzKnapp"}},
	}

	for stelle, schritt := range folge {
		gesamt := 20 * gib
		if stelle == 2 {
			gesamt = 40 * gib // 2,25 von 40 GiB ist unter 6 % — erst 3 GiB wären gut
			schritt.frei = 3 * gib
		}

		antwort := band.abrufen(t.Context(), quellenMit(schritt.frei, gesamt, frisch, schritt.dienst, 0))
		if !slices.Equal(hinweisarten(antwort), schritt.erwartet) {
			t.Errorf("Schritt %d: %v, erwartet %v", stelle, hinweisarten(antwort), schritt.erwartet)
		}

		jetzt = jetzt.Add(domain.Erholfrist)
	}
}

// **Zwei Abrufe kurz hintereinander nehmen den Hinweis nicht zurück**
// (Zweitblick über Phase 4, Befund 11) — die Startseite verlassen und wieder
// geöffnet fragt sofort, und beide lagen im kurzen Lauf eines Dienstes, der
// danach wieder abstürzte.
func TestZweiRascheAbrufeNehmenDenHinweisNicht(t *testing.T) {
	t.Parallel()

	jetzt := time.Now()
	band := betriebsgedaechtnis{uhr: func() time.Time { return jetzt }}
	frisch := SicherungsstandDTO{Stufe: "frisch"}

	for stelle, schritt := range []struct {
		dienst   domain.Skriptdienstlage
		danach   time.Duration
		erwartet bool
	}{
		{domain.Skriptdienstlage{Eingerichtet: true, LetzterAusgang: 1, Laeufe: 4}, time.Second, true},
		{domain.Skriptdienstlage{Eingerichtet: true, Laeuft: true, LetzterAusgang: 1, Laeufe: 5}, time.Second, true},
		{domain.Skriptdienstlage{Eingerichtet: true, Laeuft: true, LetzterAusgang: 1, Laeufe: 5}, domain.Erholfrist, true},
		{domain.Skriptdienstlage{Eingerichtet: true, Laeuft: true, LetzterAusgang: 1, Laeufe: 5}, 0, false},
	} {
		antwort := band.abrufen(t.Context(), quellenMit(100*gib, 500*gib, frisch, schritt.dienst, 0))
		if beendet := slices.Contains(hinweisarten(antwort), "dienstBeendet"); beendet != schritt.erwartet {
			t.Errorf("Abruf %d: beendet %t, erwartet %t", stelle, beendet, schritt.erwartet)
		}

		jetzt = jetzt.Add(schritt.danach)
	}
}

// **Was sich nicht fragen ließ, steht als ungemessen da** — und gibt keinen
// Hinweis, denn ein gescheiterter Nachweis ist kein Befund.
func TestWasSichNichtMessenLiessWirdGenannt(t *testing.T) {
	t.Parallel()

	var band betriebsgedaechtnis

	kaputt := errors.New("kaputt")
	quellen := betriebsquellen{
		platz:     func() (uint64, uint64, error) { return 0, 0, kaputt },
		sicherung: func(context.Context) (SicherungsstandDTO, error) { return SicherungsstandDTO{}, kaputt },
		dienst:    func(context.Context) (domain.Skriptdienstlage, error) { return domain.Skriptdienstlage{}, kaputt },
		neustarts: func() int { return 0 },
	}

	antwort := band.abrufen(t.Context(), quellen)
	if len(antwort.Hinweise) != 0 || !slices.Equal(antwort.Ungemessen, []string{"platz", "sicherung", "dienst"}) {
		t.Errorf("Antwort %+v", antwort)
	}
}

// **Die Lage des Dienstes aus `launchctl print`** — Ausgaben, wie macOS 27.0.1
// sie am 2026-10-15 gab.
func TestDieDienstlageAusLaunchctl(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		ausgabe  string
		erwartet domain.Skriptdienstlage
	}{
		"läuft": {
			"\tstate = running\n\truns = 2\n\tpid = 4711\n\tlast exit code = (never exited)\n",
			domain.Skriptdienstlage{Eingerichtet: true, Laeuft: true, Laeufe: 2},
		},
		"wartet nach einem Fehler": {
			"\tstate = spawn scheduled\n\truns = 1\n\tlast exit code = 78: EX_CONFIG\n",
			domain.Skriptdienstlage{Eingerichtet: true, Laeufe: 1, LetzterAusgang: 78},
		},
		"von einem Signal beendet": {
			"\tstate = not running\n\truns = 3\n\tlast terminating signal = Killed: 9\n",
			domain.Skriptdienstlage{Eingerichtet: true, Laeufe: 3, LetzterAusgang: 128 + 9},
		},
		"sauber beendet": {
			"\tstate = not running\n\truns = 1\n\tlast exit code = 0\n",
			domain.Skriptdienstlage{Eingerichtet: true, Laeufe: 1},
		},
		"der Teil eines Dienstes zählt nicht": {
			"\tstate = running\n\truns = 1\n\tpid = 12\n\t\tsubsystem = {\n\t\t\tlast exit code = 1\n\t\t}\n",
			domain.Skriptdienstlage{Eingerichtet: true, Laeuft: true, Laeufe: 1},
		},
	} {
		if lage, err := dienstlageAus(fall.ausgabe); err != nil || lage != fall.erwartet {
			t.Errorf("%s: %+v, %v; erwartet %+v", name, lage, err, fall.erwartet)
		}
	}

	if _, err := dienstlageAus("\truns = 1\n"); err == nil {
		t.Error("eine Ausgabe ohne Zustand wurde gelesen")
	}
}

// **Gefragt wird nur ein eingerichteter Dienst** — ohne plist kein Ruf an
// launchd; ein geladener, aber nie gelaufener ist kein Befund; ein Ruf, der
// scheitert, ist unbekannt.
func TestDieDienstlageFragtNurEinenEingerichtetenDienst(t *testing.T) {
	t.Parallel()

	plist := filepath.Join(t.TempDir(), dienstlabel+".plist")
	gefragt := 0
	drucken := func(context.Context, string) ([]byte, error) {
		gefragt++

		return []byte("\tstate = running\n\truns = 1\n\tpid = 9\n"), nil
	}

	if lage, err := dienstlageUeber(t.Context(), plist, dienstlabel, drucken); err != nil || lage != (domain.Skriptdienstlage{}) || gefragt != 0 {
		t.Errorf("ohne plist: %+v, %v, %d Rufe", lage, err, gefragt)
	}

	if err := os.WriteFile(plist, []byte("<plist/>"), 0o600); err != nil {
		t.Fatal(err)
	}

	if lage, err := dienstlageUeber(t.Context(), plist, dienstlabel, drucken); err != nil || !lage.Laeuft || gefragt != 1 {
		t.Errorf("mit plist: %+v, %v", lage, err)
	}

	nichtGeladen := func(ctx context.Context, _ string) ([]byte, error) {
		return nil, exec.CommandContext(ctx, "/bin/sh", "-c", "exit 113").Run()
	}

	if lage, err := dienstlageUeber(t.Context(), plist, dienstlabel, nichtGeladen); err != nil || lage != (domain.Skriptdienstlage{Eingerichtet: true}) {
		t.Errorf("nicht geladen: %+v, %v", lage, err)
	}

	scheitert := func(context.Context, string) ([]byte, error) { return nil, errors.New("launchd schweigt") }

	if _, err := dienstlageUeber(t.Context(), plist, dienstlabel, scheitert); err == nil {
		t.Error("ein gescheiterter Ruf hieß nicht unbekannt")
	}

	unlesbar := filepath.Join(plist, "darunter")
	if _, err := dienstlageUeber(t.Context(), unlesbar, dienstlabel, drucken); err == nil {
		t.Error("eine plist, die sich nicht prüfen ließ, hieß nicht eingerichtet")
	}
}

// **Das Band sieht beide Etiketten** (Zweitblick über Phase 4, Befund 7, und
// sein Nachlauf, Befund 8): blieb der Kennungsumzug unvollständig, stehen
// zwei plists da. Eine neue, die nicht geladen ist, darf einen alten Dienst,
// der abstürzt, nicht verdecken — und ein gesunder neuer seinen Fehler nicht.
func TestDasBandSiehtBeideEtiketten(t *testing.T) {
	t.Parallel()

	beendet := "\tstate = not running\n\truns = 3\n\tlast exit code = 1\n"
	laeuft := "\tstate = running\n\truns = 1\n\tpid = 9\n"

	for name, fall := range map[string]struct {
		neu, alt        string // leer: keine plist; "113": nicht geladen
		altEigen        bool
		erwartetBeendet bool
	}{
		"nur die alte, beendet":          {"", beendet, true, true},
		"neu nicht geladen, alt beendet": {"113", beendet, true, true},
		"neu läuft, alt beendet":         {laeuft, beendet, true, true},
		"neu beendet, alt läuft":         {beendet, laeuft, true, true},
		"beide laufen":                   {laeuft, laeuft, true, false},
		"die alte ist fremd":             {laeuft, beendet, false, false},
	} {
		heim := t.TempDir()
		agenten := filepath.Join(heim, "Library", "LaunchAgents")

		if fall.neu != "" {
			legeAn(t, filepath.Join(agenten, dienstlabel+".plist"), dienstplistUnter(dienstlabel))
		}

		alt := dienstplistUnter(altdienstlabel)
		if !fall.altEigen {
			alt = strings.ReplaceAll(alt, "/mse-rest<", "/server<")
		}

		legeAn(t, filepath.Join(agenten, altdienstlabel+".plist"), alt)

		drucken := func(ctx context.Context, ziel string) ([]byte, error) {
			ausgabe := fall.neu
			if strings.HasSuffix(ziel, "/"+altdienstlabel) {
				ausgabe = fall.alt
			}

			if ausgabe == "113" {
				return nil, exec.CommandContext(ctx, "/bin/sh", "-c", "exit 113").Run()
			}

			return []byte(ausgabe), nil
		}

		lage, err := dienstlageBeider(t.Context(), heim, drucken)
		if err != nil || lage.Beendet() != fall.erwartetBeendet {
			t.Errorf("%s: %+v (%v), erwartet beendet %t", name, lage, err, fall.erwartetBeendet)
		}
	}
}

// dienstplistUnter ist die plist der Anwendung unter einem Etikett.
func dienstplistUnter(etikett string) string {
	return strings.ReplaceAll(dienstplist("/A.app/Contents/MacOS/mse-rest", "/x/dienst.txt", "/x/dienst.log"), dienstlabel, etikett)
}
