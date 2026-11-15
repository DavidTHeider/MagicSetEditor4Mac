package service_test

import (
	"archive/zip"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **⌘S auf ein Set aus einem Archiv wird abgewiesen — mit einem Satz, der es
// erklärt.**
//
// # Der Befund
//
// Aus dem Zweitblick durch Codex über Phase 6
// (#208). Ein
// `.mse-set` darf eine ZIP-Datei sein; `msepaket.Oeffne` liest beides seit
// Phase 2. Bei einem Archiv ist `Herkunft` eine **Datei**, und `os.MkdirAll`
// traf dort auf eine reguläre Datei — die Meldung sprach von einem
// Verzeichnis und sagte niemandem, was los ist.
//
// **Die Entscheidung steht als Vermerk**
// ([0018](../../docs/entscheidungen/0018-ein-archiv-wird-nicht-zurueckgeschrieben.md)).
func TestEinSetAusEinemArchivWirdNichtZurueckgeschrieben(t *testing.T) {
	archiv := gepacktesSet(t)

	schreiber := service.NeuerSetschreiber(&archivleser{herkunft: archiv}, &ortsschreiber{})

	_, err := schreiber.Schreibe(t.Context(), 1, "")
	if err == nil {
		t.Fatal("kein Fehler — ⌘S auf ein Archiv ging durch")
	}

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet war ErrValidation", err)
	}

	// **Der Satz nennt das Archiv und den Ausweg.** Eine Meldung über ein
	// Verzeichnis war der Befund; eine über ein Archiv ohne Ausweg wäre nur
	// die halbe Verbesserung.
	meldung := err.Error()
	if !strings.Contains(meldung, "Archiv") || !strings.Contains(meldung, "Sichern unter") {
		t.Errorf("die Meldung nennt nicht Archiv und Ausweg: %q", meldung)
	}
}

// **„Sichern unter" geht weiterhin** — auch aus einem Archiv heraus.
//
// Das ist der Grund, aus dem die Abweisung tragbar ist: der Weg daneben tut,
// was gemeint ist.
func TestAusEinemArchivLaesstSichAnderswohinSichern(t *testing.T) {
	archiv := gepacktesSet(t)
	ziel := filepath.Join(t.TempDir(), "woanders.mse-set")

	schreiber := service.NeuerSetschreiber(&archivleser{herkunft: archiv}, &ortsschreiber{})

	if _, err := schreiber.Schreibe(t.Context(), 1, ziel); err != nil {
		t.Fatalf("sichern unter: %v", err)
	}

	if _, err := os.Stat(filepath.Join(ziel, "set")); err != nil {
		t.Errorf("die Setdatei fehlt: %v", err)
	}
}

// gepacktesSet legt ein `.mse-set` als ZIP-Datei an.
func gepacktesSet(t *testing.T) string {
	t.Helper()

	pfad := filepath.Join(t.TempDir(), "gepackt.mse-set")

	datei, err := os.Create(pfad) // #nosec G304 -- t.TempDir()
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	packer := zip.NewWriter(datei)

	eintrag, err := packer.Create("set")
	if err != nil {
		t.Fatalf("eintragen: %v", err)
	}

	if _, err := eintrag.Write([]byte("mse version: 2.0.0\ngame: probe\n")); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if err := packer.Close(); err != nil {
		t.Fatalf("packen: %v", err)
	}

	if err := datei.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	return pfad
}

// archivleser gibt eine Quelle, die aus einem Archiv stammt.
type archivleser struct {
	leserDouble

	herkunft string
}

func (a *archivleser) Quelle(context.Context, int64) (*domain.Quelle, error) {
	return &domain.Quelle{
		Datei:    "set",
		Inhalt:   []byte("mse version: 2.0.0\ngame: probe\n"),
		Herkunft: a.herkunft,
	}, nil
}

func (a *archivleser) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	return nil, nil
}

// **Und danach geht ⌘S** — die Folge, die #226 verlangt.
//
// # Der Befund
//
// Aus dem dritten Durchgang über Phase 6
// (#226):
// „Sichern unter" schrieb ein Verzeichnis, ließ den Arbeitsort aber den alten.
// Das folgende ⌘S ging wieder gegen das Archiv und wurde wieder abgewiesen —
// **der Ausweg, den die Meldung nennt, führte nicht heraus.**
//
// Die Klasse ist größer als der Archivfall: nach jedem geglückten „Sichern
// unter" arbeitete die Anwendung am alten Ort weiter. Bei einem Verzeichnis
// fiel es nur nicht auf, weil das alte Ziel schreibbar bleibt.
func TestNachDemSichernUnterGehtDasNaechsteSpeichern(t *testing.T) {
	archiv := gepacktesSet(t)
	ziel := filepath.Join(t.TempDir(), "woanders.mse-set")

	leser := &archivleser{herkunft: archiv}
	orte := &ortsschreiber{}
	schreiber := service.NeuerSetschreiber(leser, orte)

	// ⌘S: abgewiesen, weil die Herkunft ein Archiv ist.
	if _, err := schreiber.Schreibe(t.Context(), 1, ""); err == nil {
		t.Fatal("⌘S auf ein Archiv ging durch")
	}

	// „Sichern unter": geht — und wechselt den Arbeitsort.
	if _, err := schreiber.Schreibe(t.Context(), 1, ziel); err != nil {
		t.Fatalf("sichern unter: %v", err)
	}

	if orte.ort != ziel {
		t.Fatalf("der Arbeitsort steht auf %q, erwartet war %q", orte.ort, ziel)
	}

	// **Und jetzt ⌘S.** Der Bestand führt den neuen Ort; der Leser tut es hier
	// von Hand, weil er eine Attrappe ist.
	leser.herkunft = orte.ort

	if _, err := schreiber.Schreibe(t.Context(), 1, ""); err != nil {
		t.Errorf("⌘S nach dem Sichern unter: %v", err)
	}
}

// **Ein Speichern an denselben Ort schreibt den Arbeitsort nicht neu.**
//
// ⌘S geht an die Herkunft; ein Schreibvorgang auf dieselbe Angabe wäre eine
// Änderung, die niemand angeordnet hat — und im Verlauf stünde eine Bewegung
// ohne Bewegung.
func TestEinSpeichernAnDenselbenOrtLaesstDenArbeitsortInRuhe(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "hier.mse-set")
	orte := &ortsschreiber{}

	schreiber := service.NeuerSetschreiber(&archivleser{herkunft: ort}, orte)

	if _, err := schreiber.Schreibe(t.Context(), 1, ""); err != nil {
		t.Fatalf("speichern: %v", err)
	}

	if orte.ort != "" {
		t.Errorf("der Arbeitsort wurde auf %q gesetzt", orte.ort)
	}
}

// **Ein Fehler beim Wechsel nimmt das Geschriebene nicht zurück** (§7, Regel 3).
//
// Die Datei liegt da. Gemeldet wird er trotzdem: der Bestand nennt danach einen
// Ort, an dem das Set nicht mehr gepflegt wird.
func TestEinFehlerBeimOrtswechselMeldetUndBehaeltDieDatei(t *testing.T) {
	ziel := filepath.Join(t.TempDir(), "woanders.mse-set")
	kaputt := errors.New("der Bestand schweigt")

	schreiber := service.NeuerSetschreiber(
		&archivleser{herkunft: gepacktesSet(t)}, &ortsschreiber{fehler: kaputt})

	speicherung, err := schreiber.Schreibe(t.Context(), 1, ziel)
	if !errors.Is(err, kaputt) {
		t.Fatalf("Fehler: %v", err)
	}

	if speicherung == nil || speicherung.Pfad != ziel {
		t.Fatalf("Speicherung: %+v", speicherung)
	}

	if _, fehler := os.Stat(filepath.Join(ziel, "set")); fehler != nil {
		t.Errorf("die Datei ist verschwunden: %v", fehler)
	}
}
