package service

// **Ein Rückweg, der scheitert, sagt es** — A05
// (#2219),
// Restbefund des Nachlaufs über Phase A: scheiterte nach einem gescheiterten
// Tausch auch das Zurücklegen der Bilder, ging der alte Bestand trotzdem auf,
// und der Fehler sah aus wie jeder andere. **Wer die Antwort liest, muss
// unterscheiden können**, ob der bisherige Stand wieder ganz dasteht.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// bilderMitSperre legt in `bilder` ein Unterverzeichnis an, das sich nicht
// leeren lässt — das Wegräumen vor dem Zurücklegen scheitert dann.
func bilderMitSperre(t *testing.T, bilder string) {
	t.Helper()

	sperre := filepath.Join(bilder, "sperre")
	if fehler := os.MkdirAll(sperre, 0o700); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if fehler := os.WriteFile(filepath.Join(sperre, "bleibt.png"), []byte("bleibt"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	//nolint:gosec // G302: ein Verzeichnis braucht das Ausführungsrecht, um betreten zu werden
	if fehler := os.Chmod(sperre, 0o500); fehler != nil {
		t.Fatalf("Rechte nehmen: %v", fehler)
	}

	//nolint:errcheck,gosec // t.TempDir räumt sonst nicht; ein Verzeichnis braucht x
	t.Cleanup(func() { _ = os.Chmod(sperre, 0o700) })
}

// erstesOeffnenScheitert lässt das erste Öffnen scheitern — vorher läuft
// `davor` — und das zweite gelingen.
func erstesOeffnenScheitert(weg *Bestandsweg, davor, beimZweiten func()) {
	echt := weg.oeffnen
	versuche := 0
	weg.oeffnen = func(ctx context.Context) error {
		versuche++
		if versuche == 1 {
			davor()

			return errors.New("die neue Datei taugt nicht")
		}

		beimZweiten()

		return echt(ctx)
	}
}

func TestEinGelungenerRueckwegIstKeinGescheiterter(t *testing.T) {
	bau := neuerWegbau(t)
	bau.leben.oeffnungsfehler = errors.New("die frische Datei taugt nicht")

	_, fehler := bau.weg.Neuanfang(t.Context())
	if fehler == nil {
		t.Fatal("das gescheiterte Öffnen kam durch")
	}

	if errors.Is(fehler, ErrRueckwegGescheitert) {
		t.Errorf("ein gelungener Rückweg meldet sich als gescheitert: %v", fehler)
	}
}

func TestScheitertDerRueckwegDerBilderBeimZurueckspielenSagtEsDerFehler(t *testing.T) {
	weg, ort, bilder := mitBilderort(t)

	if fehler := os.MkdirAll(bilder, 0o700); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	ziel := gesichert(t, weg, filepath.Join(ort, "gut"+sicherung.Endung))

	if fehler := os.WriteFile(filepath.Join(bilder, "alt.png"), []byte("der laufende Stand"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	erstesOeffnenScheitert(weg, func() { bilderMitSperre(t, bilder) }, func() {})

	_, fehler := weg.Zurueckspielen(t.Context(), ziel)
	if !errors.Is(fehler, ErrRueckwegGescheitert) {
		t.Fatalf("das Zurückspielen gab %v, erwartet war der gescheiterte Rückweg", fehler)
	}

	// **Verloren ist nichts**: die alten Bilder liegen beiseite.
	roh, lesefehler := os.ReadFile(filepath.Join(filepath.Dir(bilder), Beiwerkbeiseite, "alt.png")) // #nosec G304 -- aus t.TempDir()
	if lesefehler != nil || string(roh) != "der laufende Stand" {
		t.Errorf("die alten Bilder liegen nicht beiseite: %q, %v", roh, lesefehler)
	}
}

func TestScheitertDerRueckwegDerBilderBeimNeuanfangSagtEsDerFehler(t *testing.T) {
	bau := neuerWegbau(t)

	erstesOeffnenScheitert(bau.weg, func() { bilderMitSperre(t, bau.bilder) }, func() {})

	if _, fehler := bau.weg.Neuanfang(t.Context()); !errors.Is(fehler, ErrRueckwegGescheitert) {
		t.Errorf("der Neuanfang gab %v, erwartet war der gescheiterte Rückweg", fehler)
	}
}

func TestScheitertDerRueckwegDesBestandsSagtEsDerFehler(t *testing.T) {
	bau := neuerWegbau(t)

	erstesOeffnenScheitert(bau.weg, func() {
		//nolint:gosec // G302: ein Verzeichnis braucht das Ausführungsrecht, um betreten zu werden
		if fehler := os.Chmod(bau.ort, 0o500); fehler != nil {
			t.Fatalf("Rechte nehmen: %v", fehler)
		}

		//nolint:errcheck,gosec // t.TempDir räumt sonst nicht
		t.Cleanup(func() { _ = os.Chmod(bau.ort, 0o700) })
	}, func() {})

	if _, fehler := bau.weg.Neuanfang(t.Context()); !errors.Is(fehler, ErrRueckwegGescheitert) {
		t.Errorf("der Neuanfang gab %v, erwartet war der gescheiterte Rückweg", fehler)
	}
}

// **Beim Neuanfang liegen die alten Bilder wieder da, bevor der alte
// Bestand aufgeht** — dieselbe Reihenfolge wie beim Zurückspielen (Befund 3
// des Zweitblicks über Phase A); hier stand sie bis zum Nachlauf umgekehrt.
func TestEinGescheiterterNeuanfangLegtDieBilderVorDemBestandZurueck(t *testing.T) {
	bau := neuerWegbau(t)

	if fehler := os.MkdirAll(bau.bilder, 0o700); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if fehler := os.WriteFile(filepath.Join(bau.bilder, "alt.png"), []byte("der laufende Stand"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	beimZweiten := ""
	erstesOeffnenScheitert(bau.weg, func() {}, func() {
		roh, _ := os.ReadFile(filepath.Join(bau.bilder, "alt.png")) //nolint:errcheck,gosec // aus t.TempDir(); fehlt sie, vergleicht der Lauf ""
		beimZweiten = string(roh)
	})

	if _, fehler := bau.weg.Neuanfang(t.Context()); fehler == nil {
		t.Fatal("das gescheiterte Öffnen kam durch")
	}

	if beimZweiten != "der laufende Stand" {
		t.Errorf("der alte Bestand ging auf, während die Bilder %q waren", beimZweiten)
	}
}
