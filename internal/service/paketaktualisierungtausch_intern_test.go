package service

// **Einsetzen, Aufheben und Tauschen, wenn etwas scheitert** — G04
// (#2251).
//
// Getrennt von `paketaktualisierung_intern_test.go` (§2, Zeilengrenze): dort
// die Fehlerwege der Vorschau und des Vermerks, hier die der Verzeichnisse —
// und die Zusage, dass **nie ein Baum ohne seinen Platz** zurückbleibt.

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// umbenennerBis scheitert beim n-ten Ruf.
func umbenennerBis(n int) func(alt, neu string) error {
	gerufen := 0

	return func(alt, neu string) error {
		gerufen++
		if gerufen == n {
			return errors.New("umbenennen scheitert")
		}

		return os.Rename(alt, neu)
	}
}

// dreiBaeume legt Ort, vorläufigen und aufgehobenen Baum an.
func dreiBaeume(t *testing.T) (ort, vorlaeufig, aufheben string) {
	t.Helper()

	wurzel := t.TempDir()
	ort, vorlaeufig, aufheben = filepath.Join(wurzel, "ort"), filepath.Join(wurzel, "neu"), filepath.Join(wurzel, "vorige", "ort")

	for pfad, inhalt := range map[string]string{ort: "bisher", vorlaeufig: "neu", aufheben: "aelter"} {
		if err := os.MkdirAll(pfad, 0o750); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(pfad, "kennung"), []byte(inhalt), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	return ort, vorlaeufig, aufheben
}

func baumkennung(t *testing.T, ort string) string {
	t.Helper()

	inhalt, err := os.ReadFile(filepath.Join(ort, "kennung")) // #nosec G304 -- aus t.TempDir()
	if err != nil {
		return ""
	}

	return string(inhalt)
}

func TestEinScheiterndesEinsetzenLaesstJedenBaumAnSeinemPlatz(t *testing.T) {
	for zug := 1; zug <= 3; zug++ {
		ort, vorlaeufig, aufheben := dreiBaeume(t)

		if err := paketEinsetzenUndAufheben(ort, vorlaeufig, aufheben, umbenennerBis(zug)); err == nil {
			t.Fatalf("Zug %d: kein Fehler", zug)
		}

		if baumkennung(t, ort) != "bisher" || baumkennung(t, aufheben) != "aelter" || baumkennung(t, vorlaeufig) != "neu" {
			t.Errorf("Zug %d: ort %q, aufgehoben %q, vorläufig %q", zug,
				baumkennung(t, ort), baumkennung(t, aufheben), baumkennung(t, vorlaeufig))
		}
	}

	ort, vorlaeufig, aufheben := dreiBaeume(t)
	if err := paketEinsetzenUndAufheben(ort, vorlaeufig, aufheben, os.Rename); err != nil ||
		baumkennung(t, ort) != "neu" || baumkennung(t, aufheben) != "bisher" {
		t.Errorf("geglückt: %v, ort %q, aufgehoben %q", err, baumkennung(t, ort), baumkennung(t, aufheben))
	}

	if _, err := os.Stat(aufheben + ".alt"); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("der ältere aufgehobene Stand blieb liegen: %v", err)
	}
}

func TestEinVersperrterAufhebeortHaeltDasEinsetzenAn(t *testing.T) {
	ort, vorlaeufig, _ := dreiBaeume(t)
	datei := filepath.Join(t.TempDir(), "datei")

	if err := os.WriteFile(datei, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := paketEinsetzenUndAufheben(ort, vorlaeufig, filepath.Join(datei, "ort"), os.Rename); err == nil || baumkennung(t, ort) != "bisher" {
		t.Errorf("Fehler %v, ort %q", err, baumkennung(t, ort))
	}
}

// **Ein älterer Stand, der sich nicht räumen lässt, verwirft den neuen
// nicht** (§7, Regel 3) — er wird aufgeschrieben.
func TestEinUnraeumbarerAltstandBleibtLiegen(t *testing.T) {
	ort, vorlaeufig, aufheben := dreiBaeume(t)
	gesperrt := filepath.Join(aufheben, "gesperrt")

	if err := os.MkdirAll(gesperrt, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(gesperrt, "x"), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.Chmod(gesperrt, 0o500); err != nil { // #nosec G302 -- ein Verzeichnis ohne Schreibrecht ist der Fall
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = os.Chmod(aufheben+".alt/gesperrt", 0o700) }) //nolint:errcheck,gosec // Aufräumen

	if err := paketEinsetzenUndAufheben(ort, vorlaeufig, aufheben, os.Rename); err != nil || baumkennung(t, ort) != "neu" {
		t.Errorf("Fehler %v, ort %q", err, baumkennung(t, ort))
	}
}

func TestEinScheiternderTauschLaesstBeideAnIhremPlatz(t *testing.T) {
	// **Auch der dritte Zug** — Befund 1 des Zweitblicks über Phase G,
	// zweiter Teil: scheiterte er, lag der zurückgeholte Baum schon am Ort
	// und der andere unter `.tausch`, und gemeldet wurde ein Fehlschlag.
	for zug := 1; zug <= 3; zug++ {
		ort, _, aufheben := dreiBaeume(t)

		if err := tauschen(ort, aufheben, umbenennerBis(zug)); err == nil {
			t.Fatalf("Zug %d: kein Fehler", zug)
		}

		if baumkennung(t, ort) != "bisher" || baumkennung(t, aufheben) != "aelter" {
			t.Errorf("Zug %d: ort %q, aufgehoben %q", zug, baumkennung(t, ort), baumkennung(t, aufheben))
		}
	}
}

// **Ohne Paketordner keine Sperre** — und ohne Sperre kein Tausch.
func TestEinZurueckOhnePaketordnerTauschtNicht(t *testing.T) {
	aktualisierung, verwaltet := innererBezug(t, &zaehlenderKatalog{}, innereAuskunft{})
	vorige := filepath.Join(verwaltet+"-vorige-bezuege", "Pack")

	if err := os.MkdirAll(vorige, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := vermerkSchreiben(vorige, Bezugsvermerk{Stand: "x"}); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(verwaltet); err != nil {
		t.Fatal(err)
	}

	if _, err := aktualisierung.Zurueck(t.Context(), "Pack"); err == nil {
		t.Error("kein Fehler")
	}

	if _, err := os.Stat(vorige); err != nil {
		t.Errorf("der aufgehobene Stand ist fort: %v", err)
	}
}

// umbenennerMit scheitert bei den genannten Rufen, je mit eigenem Fehler.
func umbenennerMit(fehler map[int]error) func(alt, neu string) error {
	gerufen := 0

	return func(alt, neu string) error {
		gerufen++
		if grund, da := fehler[gerufen]; da {
			return grund
		}

		return os.Rename(alt, neu)
	}
}

// **Ein Rückweg, der selbst scheitert, wird gemeldet** — Befund 1 des
// Zweitblicks über Phase G, zweiter Teil: bis dahin verwarf das Einsetzen
// den Fehler seiner Wiederherstellung, und der Bericht sagte nicht, dass
// ein Baum nicht an seinem Platz liegt.
func TestEinScheiternderRueckwegStehtImFehler(t *testing.T) {
	einsetzen, zurueck := errors.New("einsetzen scheitert"), errors.New("zurück scheitert")

	ort, vorlaeufig, aufheben := dreiBaeume(t)
	err := paketEinsetzenUndAufheben(ort, vorlaeufig, aufheben, umbenennerMit(map[int]error{3: einsetzen, 4: zurueck}))

	if !errors.Is(err, einsetzen) || !errors.Is(err, zurueck) {
		t.Errorf("Einsetzen: %v", err)
	}

	ort, _, aufheben = dreiBaeume(t)
	err = tauschen(ort, aufheben, umbenennerMit(map[int]error{3: einsetzen, 4: zurueck}))

	if !errors.Is(err, einsetzen) || !errors.Is(err, zurueck) {
		t.Errorf("Tausch: %v", err)
	}
}

// **Beim ersten Aufheben gibt es nichts zurückzuholen** — scheitert das
// Einsetzen dann, bleibt der Baum, wo er war, und der Fehler allein steht da.
func TestEinErstesAufhebenOhneVorigenStand(t *testing.T) {
	ort, vorlaeufig, aufheben := dreiBaeume(t)
	if err := os.RemoveAll(aufheben); err != nil {
		t.Fatal(err)
	}

	einsetzen := errors.New("einsetzen scheitert")

	err := paketEinsetzenUndAufheben(ort, vorlaeufig, aufheben, umbenennerMit(map[int]error{1: einsetzen}))
	if !errors.Is(err, einsetzen) || baumkennung(t, ort) != "bisher" {
		t.Errorf("Fehler %v, ort %q", err, baumkennung(t, ort))
	}
}
