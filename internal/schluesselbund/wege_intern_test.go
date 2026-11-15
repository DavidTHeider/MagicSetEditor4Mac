//go:build darwin

package schluesselbund

// **Die Fehlerwege des Schlüsselbunds** — ein interner Lauf, weil er
// [Bund.ortsteil] und [ausHex] braucht.
//
// **Der Anmeldebund wird hier nur *gebaut*, nie beschrieben.** Ein Lauf, der
// in den echten Bund schriebe, hinterließe dort einen Eintrag und stolperte
// über den Zugriffsdialog von macOS (§3).

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Der Anmeldebund nennt keinen Ort** — dann nimmt `security` den, den macOS
// für Anmeldungen führt.
func TestDerAnmeldebundNenntKeinenOrt(t *testing.T) {
	t.Parallel()

	if teil := Anmeldebund().ortsteil(); teil != "" {
		t.Errorf("der Anmeldebund nennt %q, erwartet nichts", teil)
	}

	// **Und ein benannter Bund nennt seinen Ort** — sonst prüfte die Zeile
	// darüber nur, dass die Funktion immer nichts gibt.
	if teil := NeuerBund("/tmp/probe.keychain").ortsteil(); teil != " /tmp/probe.keychain" {
		t.Errorf("ein benannter Bund nennt %q", teil)
	}
}

// **Ein Bund, den es nicht gibt, meldet sich — beim Schreiben.**
//
// # Und beim Lesen und Löschen ist er von einem fehlenden Eintrag nicht zu
// unterscheiden
//
// **Gemessen am 2026-09-01**: `security` gibt für „der Bund fehlt" und „der
// Eintrag fehlt" **dieselbe** Meldung und denselben Code —
// `SecKeychainSearchCopyNext: The specified item could not be found` /
// `-25300`. Es gibt keinen Weg, sie an der Antwort zu trennen.
//
// **Der erste Anlauf dieses Tests erwartete einen Fehler beim Löschen** und
// war rot. Nicht der Code lag falsch, sondern die Erwartung: was sich nicht
// unterscheiden lässt, darf ein Programm nicht unterscheiden **behaupten**.
//
// Was daraus folgt, ist die Auskunft, die die Fläche gibt: „für diesen Dienst
// liegt nichts im Schlüsselbund" — und die stimmt in beiden Fällen.
func TestEinFehlenderBundMeldetSich(t *testing.T) {
	t.Parallel()

	bund := NeuerBund(filepath.Join(t.TempDir(), "gibtsnicht.keychain")).
		MitFrist(2 * time.Second)

	// **Schreiben geht nicht, und es ist ein `ErrNotFound`** — der Bund ist
	// das, was nicht gefunden wurde.
	//
	// **Hier stand das Gegenteil**, mit der Begründung, `ErrNotFound` hieße
	// „lege es an". Das war eine Aussage über den **Eintrag**; gemeint ist
	// der **Bund**, und den legt dieses Programm nicht an. Geändert am
	// 2026-09-02 mit der Prüfung, die den Systemdialog verhindert.
	if err := bund.Schreibe(t.Context(), "mse", "konto", "geheim"); !errors.Is(
		err, domain.ErrNotFound) {
		t.Errorf("Schreiben gab %v, erwartet ErrNotFound", err)
	}

	// **Lesen gibt „nichts da"** — dieselbe Auskunft wie bei einem fehlenden
	// Eintrag, weil `security` beides gleich meldet.
	if _, err := bund.Lies(t.Context(), "mse", "konto"); !errors.Is(
		err, domain.ErrNotFound) {
		t.Errorf("das Lesen gab %v, erwartet ErrNotFound", err)
	}

	// **Löschen glückt** — wer löscht, will, dass nichts da ist, und das ist
	// dann schon der Fall.
	if err := bund.Loesche(t.Context(), "mse", "konto"); err != nil {
		t.Errorf("das Löschen gab %v, erwartet nichts", err)
	}
}

// **Was kein Hex ist, wird als solches gemeldet.**
//
// Der Fall entsteht, wenn im Bund ein Eintrag liegt, den **nicht** dieses
// Programm geschrieben hat — jemand hat ihn von Hand angelegt, oder eine
// frühere Fassung legte Klartext ab. Eine stille leere Antwort wäre dort das
// Schlimmste: der Dienst bliebe „an" und schwiege.
func TestWasKeinHexIstWirdGemeldet(t *testing.T) {
	t.Parallel()

	if _, err := ausHex("kein hex"); err == nil {
		t.Error("»kein hex« kam durch, erwartet ein Fehler")
	}

	// **Und Hex kommt durch** — mit dem Zeilenumbruch, den `security` anhängt.
	wort, err := ausHex(alsHex("geheim") + "\n")
	if err != nil {
		t.Fatalf("Hex mit Umbruch: %v", err)
	}

	if wort != "geheim" {
		t.Errorf("zurück kam %q, erwartet »geheim«", wort)
	}
}

// **Ein abgebrochener Kontext bricht auch den Unterprozess ab.**
//
// # Und er ist der einzige Weg zum anderen Zweig
//
// `security` meldet für **jeden** Fehlschlag beim Suchen und Löschen
// dasselbe: `SecKeychainSearchCopyNext … could not be found` / `-25300`.
// **Gemessen am 2026-09-01** an drei Fällen — Bund fehlt, Bund leer, und eine
// Datei, die gar kein Bund ist. Alle drei geben dieselbe Zeile.
//
// Damit ist der Zweig „ein anderer Fehler als *nicht gefunden*" über eine
// kaputte Datei **nicht** zu erreichen. Erreichbar ist er über einen
// abgebrochenen Aufruf: dann scheitert der Unterprozess, ohne etwas auf die
// Fehlerausgabe zu schreiben.
//
// **Und deshalb steht der Zweig da und ist kein toter Ausweg** — dieselbe
// Frage wie beim kurzsichtigen Ausweg der Paarung (Phase 4.7): ein Zweig, den
// nie jemand betritt, macht eine Zusage, die niemand einlöst.
func TestEinAbgebrochenerKontextBrichtAb(t *testing.T) {
	t.Parallel()

	// **Ein echter Bund.** Seit dem 2026-09-02 prüft [Bund.pruefeOrt] vor
	// jedem Aufruf, ob es ihn gibt — ein fehlender käme gar nicht bis zum
	// Unterprozess, und dieser Lauf prüfte dann die Prüfung statt des
	// Abbruchs.
	ort := echterBund(t)

	for _, fall := range []struct {
		name string
		ruf  func(context.Context, *Bund) error
	}{
		{"schreiben", func(ctx context.Context, b *Bund) error {
			return b.Schreibe(ctx, "mse", "konto", "geheim")
		}},
		{"lesen", func(ctx context.Context, b *Bund) error {
			_, err := b.Lies(ctx, "mse", "konto")

			return err
		}},
		{"löschen", func(ctx context.Context, b *Bund) error {
			return b.Loesche(ctx, "mse", "konto")
		}},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			ctx, abbrechen := kontextAbgebrochen(t)
			defer abbrechen()

			err := fall.ruf(ctx, NeuerBund(ort))
			if err == nil {
				t.Fatal("der Aufruf glückte, erwartet ein Fehler")
			}

			// **Und er kommt nicht als »nichts da« an.** Ein abgebrochener
			// Aufruf sagt nichts darüber, ob etwas im Bund liegt.
			if errors.Is(err, domain.ErrNotFound) {
				t.Errorf("ein Abbruch kam als ErrNotFound an: %v", err)
			}

			if !strings.Contains(err.Error(), "Geheimnis") {
				t.Errorf("die Meldung %q nennt nicht, worum es ging", err)
			}
		})
	}
}

// kontextAbgebrochen gibt einen Kontext, der schon abgebrochen ist.
func kontextAbgebrochen(t *testing.T) (ctx context.Context, abbrechen func()) {
	t.Helper()

	ctx, abbrechen = context.WithCancel(t.Context())
	abbrechen()

	return ctx, abbrechen
}

// echterBund legt einen Schlüsselbund an und gibt seinen Pfad.
//
// **Nie den Anmeldebund** (§3): er liegt in `t.TempDir()` und wird danach
// gelöscht.
func echterBund(t *testing.T) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "echt.keychain")

	//nolint:gosec // das Programm steht als Literal, der Ort liegt in t.TempDir()
	if aus, err := exec.CommandContext(t.Context(), "/usr/bin/security",
		"create-keychain", "-p", "leer", ort).CombinedOutput(); err != nil {
		t.Fatalf("den Bund anlegen: %v — %s", err, aus)
	}

	t.Cleanup(func() {
		ctx, fertig := context.WithTimeout(context.Background(), 10*time.Second)
		defer fertig()

		//nolint:gosec,errcheck // dito
		_ = exec.CommandContext(ctx, "/usr/bin/security",
			"delete-keychain", ort).Run()
	})

	return ort
}
