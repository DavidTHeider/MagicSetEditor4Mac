package app

// **Wer wandert, hält den eingerichteten Dienst vorher an — nachweislich** —
// M02 (#2346),
// Befund A-11; Zweitblick über Phase 2 von 0.9.0, Befunde 5 und 6 (#2323).
//
// Der Dienst ist hier ein echter Kindprozess, und der „Neustart" beendet ihn —
// so sieht der Wärter, ob der alte Prozess wirklich fort ist.
//
// **Nicht parallel**: die Frage nach dem Dienst und die Naht `migriere` sind
// Paketvariablen.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// alterDienst startet einen Prozess, der für den Dienst steht und `pfad`
// offen hält, und meldet ihn als den laufenden Dienst.
func alterDienst(t *testing.T, pfad string) *exec.Cmd {
	t.Helper()

	// #nosec G204 -- ein festes Programm des Laufs, der Pfad aus t.TempDir()
	prozess := exec.CommandContext(t.Context(), "/bin/sh", "-c", `exec 3<"$1"; exec /bin/sleep 30`, "sh", pfad)
	if err := prozess.Start(); err != nil {
		t.Fatal(err)
	}

	beendet := make(chan struct{})

	go func() {
		_ = prozess.Wait() //nolint:errcheck // beendet ist beendet
		close(beendet)
	}()

	t.Cleanup(func() {
		_ = prozess.Process.Kill() //nolint:errcheck // vielleicht schon fort
		<-beendet
	})

	vorher := dienstprozesse
	dienstprozesse = func(context.Context) ([]dienstprozess, error) {
		select {
		case <-beendet:
			return nil, nil
		default:
			return nurDerNeue(prozess.Process.Pid), nil
		}
	}

	t.Cleanup(func() { dienstprozesse = vorher })

	// **Gewartet wird, bis er die Datei hält** — sonst sähe `lsof` ihn nicht.
	for range 200 {
		if ausgabe, _ := lsofFragen(t.Context(), pfad); slices.Contains(splitZeilen(ausgabe), "p"+strconv.Itoa(prozess.Process.Pid)) { //nolint:errcheck // gezählt
			return prozess
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("der Dienst hält die Datei nicht")

	return nil
}

// neuereFassungStartet startet als eine Fassung, die `mehr` Wanderungen über
// die heutigen hinaus mitbringt — `befehl` steht für `launchctl` — und gibt
// die gerufenen Befehle, ob die Sicherung beim Ruf schon dalag, und den Fehler
// des Starts.
func neuereFassungStartet(
	t *testing.T, pfad, heim string, mehr int, befehl befehlsausfuehrung,
) (befehle []string, gesichertVorher bool, startfehler error) {
	t.Helper()

	if mehr > 0 {
		danach := fstest.MapFS{}
		for stelle := 1; stelle <= mehr; stelle++ {
			name := fmt.Sprintf("%04d_probe%d.sql", heuteVerbucht(t)+stelle, stelle)
			danach["migrationen/"+name] = &fstest.MapFile{Data: []byte(fmt.Sprintf("CREATE TABLE probe_m02_%d (wert INTEGER) STRICT;", stelle))}
		}

		vorher := migriere
		migriere = func(ctx context.Context, db *sql.DB, vorbereiten storage.Vorbereitung) error {
			return storage.MigriereMit(ctx, db, danach, vorbereiten)
		}

		t.Cleanup(func() { migriere = vorher })
	}

	sicherungenVorher := sicherungen(pfad)

	anwendung := NeueApp(pfad).
		MitSchnittstellenorten(func() (string, error) { return heim, nil },
			func() (string, error) { return filepath.Join(heim, "werkzeug"), nil },
			func(context.Context) string { return "" }).
		MitSchnittstellenbefehlen(func(ctx context.Context, gerufen []string) error {
			befehle = append(befehle, strings.Join(gerufen, " "))
			gesichertVorher = sicherungen(pfad) > sicherungenVorher

			return befehl(ctx, gerufen)
		})

	startfehler = anwendung.Starten(t.Context())

	if fehler := anwendung.Beenden(); fehler != nil {
		t.Fatalf("beenden: %v", fehler)
	}

	return befehle, gesichertVorher, startfehler
}

// sicherungen zählt die Sicherungen vor einer Wanderung neben dem Bestand.
func sicherungen(pfad string) int {
	gefunden, _ := filepath.Glob(filepath.Join(filepath.Dir(pfad), "bestand-vor-*.db")) //nolint:errcheck // ein festes Muster

	return len(gefunden)
}

// einBestand legt einen heutigen Bestand an und gibt seinen Ort.
func einBestand(t *testing.T) (pfad, heim string) {
	t.Helper()

	ort := t.TempDir()
	pfad, heim = filepath.Join(ort, "bestand.db"), filepath.Join(ort, "zuhause")

	if _, _, fehler := neuereFassungStartet(t, pfad, heim, 0, func(context.Context, []string) error { return nil }); fehler != nil {
		t.Fatal(fehler)
	}

	return pfad, heim
}

func TestWerWandertHaeltDenDienstVorherAn(t *testing.T) { //nolint:paralleltest // siehe darüber
	pfad, heim := einBestand(t)
	dienst := alterDienst(t, pfad)

	befehle, gesichertVorher, fehler := neuereFassungStartet(t, pfad, heim, 1, func(context.Context, []string) error {
		return dienst.Process.Kill()
	})
	if fehler != nil {
		t.Fatalf("der Start: %v", fehler)
	}

	erwartet := "launchctl kickstart -k gui/" + strconv.Itoa(os.Getuid()) + "/" + dienstlabel
	if !slices.Equal(befehle, []string{erwartet}) {
		t.Errorf("gerufen: %v, erwartet %q", befehle, erwartet)
	}

	if gesichertVorher {
		t.Error("der Dienst wurde erst nach der Sicherung angehalten — bis dahin schrieb der alte weiter")
	}

	if sicherungen(pfad) != 1 {
		t.Errorf("%d Sicherungen, erwartet eine", sicherungen(pfad))
	}
}

// **Endet der alte Prozess nicht, wird nicht gewandert** — Befund 5: bis dahin
// reichte ein Ruf, dessen Ausgang nur im Protokoll stand.
func TestEinDienstDerNichtEndetHaeltDieWanderungAuf(t *testing.T) { //nolint:paralleltest // siehe darüber
	pfad, heim := einBestand(t)
	alterDienst(t, pfad)

	vorher := dienstanhaltefrist
	dienstanhaltefrist = 200 * time.Millisecond

	t.Cleanup(func() { dienstanhaltefrist = vorher })

	for name, befehl := range map[string]befehlsausfuehrung{
		"der Ruf geht durch, der Prozess bleibt": func(context.Context, []string) error { return nil },
		"der Ruf scheitert":                      func(context.Context, []string) error { return errors.New("Could not find service") },
	} {
		_, _, fehler := neuereFassungStartet(t, pfad, heim, 1, befehl)
		if !errors.Is(fehler, domain.ErrBelegt) {
			t.Errorf("%s: der Start gab %v, erwartet belegt", name, fehler)
		}

		if sicherungen(pfad) != 0 {
			t.Errorf("%s: gesichert, obwohl nicht gewandert wird", name)
		}
	}
}

// **Angehalten wird nur ein Dienst, der diesen Bestand hält** — gefragt bei
// `lsof` (Nachlauf über Phase 2, Befunde 1, 5 und 6): ein Dienst, der einen
// anderen Bestand bedient, schreibt nicht in diesen, und ein Lauf mit eigenem
// Bestand trifft so nie den Dienst dieses Rechners. Ohne laufenden Dienst und
// für den Dienst selbst kein Ruf — **und ein Nachweis, der scheitert, ist
// kein Nein**.
func TestNurEinDienstDerDiesenBestandHaeltWirdAngehalten(t *testing.T) { //nolint:paralleltest // siehe darüber
	pfad, _ := einBestand(t)
	anderer := filepath.Join(t.TempDir(), "anderer.db")

	if err := os.WriteFile(anderer, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	keinRuf := func(context.Context, []string) error { t.Error("gerufen"); return nil }
	waerter := dienstwaerter{befehl: keinRuf}

	alterDienst(t, anderer)

	if fehler := waerter.anhalten(t.Context(), pfad); fehler != nil {
		t.Errorf("ein Dienst über einem anderen Bestand: %v", fehler)
	}

	for name, fall := range map[string]struct {
		pid    int
		fehler error
		belegt bool
	}{
		"kein laufender":         {pid: 0},
		"der Dienst ist dies":    {pid: os.Getpid()},
		"der Nachweis scheitert": {fehler: errors.New("launchctl antwortet nicht"), belegt: true},
	} {
		dienstprozesse = func(context.Context) ([]dienstprozess, error) { return nurDerNeue(fall.pid), fall.fehler }

		if fehler := waerter.anhalten(t.Context(), pfad); errors.Is(fehler, domain.ErrBelegt) != fall.belegt {
			t.Errorf("%s: %v, belegt erwartet %v", name, fehler, fall.belegt)
		}
	}

	if fehler := waerter.anhalten(t.Context(), ""); fehler != nil {
		t.Errorf("ein Bestand ohne Ort: %v", fehler)
	}
}

// **Das Warten endet mit dem Kontext.**
func TestDasWartenAufDenDienstEndetMitDemKontext(t *testing.T) { //nolint:paralleltest // siehe darüber
	pfad, _ := einBestand(t)
	alterDienst(t, pfad)

	ctx, abbrechen := context.WithCancel(t.Context())

	waerter := dienstwaerter{
		befehl: func(context.Context, []string) error { abbrechen(); return nil },
	}

	if fehler := waerter.anhalten(ctx, pfad); !errors.Is(fehler, context.Canceled) {
		t.Errorf("abgebrochen: %v", fehler)
	}
}

// **Die Frist gilt für den ganzen Vorgang** (Befund 5): ein `launchctl`, das
// nicht zurückkehrt, hält die Wanderung nicht länger auf als sie.
func TestDieFristGiltFuerDenGanzenVorgang(t *testing.T) { //nolint:paralleltest // siehe darüber
	pfad, _ := einBestand(t)
	alterDienst(t, pfad)

	vorher := dienstanhaltefrist
	dienstanhaltefrist = 200 * time.Millisecond

	t.Cleanup(func() { dienstanhaltefrist = vorher })

	waerter := dienstwaerter{
		befehl: func(ctx context.Context, _ []string) error { <-ctx.Done(); return ctx.Err() },
	}

	anfang := time.Now()
	if fehler := waerter.anhalten(t.Context(), pfad); !errors.Is(fehler, domain.ErrBelegt) || time.Since(anfang) > 2*time.Second {
		t.Errorf("nach %s: %v", time.Since(anfang), fehler)
	}
}

// **Jede Wurzel hält den Dienst über denselben Weg an** — `mse`, `mse-mcp`
// und `mse-rest` sichern über [SicherungVorMigration], und dort steht der
// Wärter (Befund 5). **Und in einem Testlauf weist der echte Ausführer den
// Neustart ab** (Befund 6), statt den Dienst dieses Rechners zu treffen —
// dann wird auch nicht gesichert. Der „Dienst" ist ein Prozess, der den
// Bestand des Laufs hält.
func TestDieWurzelnHaltenDenDienstUeberDenWaerterAn(t *testing.T) { //nolint:paralleltest // siehe darüber
	pfad := filepath.Join(t.TempDir(), "bestand.db")

	db, err := storage.Oeffnen(t.Context(), pfad)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // nur zum Aufräumen

	alterDienst(t, pfad)

	fehler := SicherungVorMigration(db, filepath.Dir(pfad), Fassung)(t.Context())
	if !errors.Is(fehler, domain.ErrBelegt) || !strings.Contains(domain.GrundVon(fehler).Error(), "Testlauf") {
		t.Errorf("im Testlauf: %v (%v)", fehler, domain.GrundVon(fehler))
	}

	if sicherungen(pfad) != 0 {
		t.Error("gesichert, obwohl der Dienst nicht angehalten war")
	}
}
