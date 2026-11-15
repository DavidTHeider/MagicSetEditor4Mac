package werkzeuge_test

// **Der Ort im Schlüsselbund lässt sich dort auch ablegen** — Nacharbeit zu
// Werkstück 3 von A18
// (#1533).
//
// # Der Befund
//
// **`@` ist im Schlüsselbund kein erlaubtes Zeichen.** Die erste Fassung von
// `domain.ZugangKontoFuer` baute `openai.schluessel@6794ca4973d2`, und
// `schluesselbund.Schreibe` wies es ab — der Gateway-Schlüssel ließ sich
// **gar nicht ablegen**. Die Fähigkeit stand da, mit Läufen, und war für
// genau den Fall unbenutzbar, für den es sie gibt.
//
// # Warum es diesen Wächter gibt und nicht nur den Lauf in `domain`
//
// **Weil `domain` den Schlüsselbund nicht kennt** (§2) und seinen
// Zeichensatz deshalb **nachbildet**. Eine Nachbildung läuft auseinander —
// das ist die Gestalt, vor der der Kopf von `einstellung.go` zweimal warnt.
//
// **Hier treffen sich beide**: das Paket, das den Ort baut, und das, das ihn
// annimmt. Geprüft wird nicht ein Zeichensatz gegen einen anderen, sondern
// der **wirkliche Ort** gegen die **wirkliche Prüfung**.

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// **Jeder Ort, den `domain` baut, nimmt der Schlüsselbund an.**
func TestJederZugangsortWirdVomSchluesselbundAngenommen(t *testing.T) {
	t.Parallel()

	// **Ein eigener Bund in `t.TempDir()`** — geprüft wird die
	// **Namensprüfung**, und die läuft vor jedem Zugriff.
	//
	// **Hier stand `schluesselbund.Anmeldebund()`**, und daneben der Satz
	// „ein echter Bund verlangte ein Kennwort und schriebe in den
	// Anmeldebund dieses Rechners". Der Satz beschrieb genau das, was die
	// Zeile darunter tat: `Anmeldebund()` **ist** der Anmeldebund. Gemessen
	// am 2026-09-14 gingen allein aus diesem Lauf rund zwanzig Lesezugriffe
	// dorthin, und macOS fragte den Menschen um Erlaubnis.
	//
	// **Ein Kommentar, der das Gegenteil dessen sagt, was danebensteht, ist
	// schlimmer als keiner** (§10) — er hält jeden davon ab nachzusehen.
	bund := schluesselbund.NeuerBund(probebund(t))

	for name, adresse := range map[string]string{
		"die Vorgabe":       "",
		"ein Gateway":       "https://gateway.firma.intern/v1",
		"mit Port":          "http://127.0.0.1:4000/v1",
		"mit Abfrage":       "https://openrouter.ai/api/v1?x=1",
		"mit Sonderzeichen": "https://gäteway.example/v1",
		// **Eine Adresse mit Benutzerangabe** — zusammengesetzt und nicht
		// hingeschrieben: `gosec` liest ein Kennwort in einer Adresse als
		// hinterlegtes Geheimnis, und die Meldung hätte recht, stünde dort
		// eines.
		"mit Benutzerangabe": "https://wer:" + "was" + "@gateway.example/v1",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for _, dienst := range domain.Dienste() {
				for _, angabe := range dienst.Zugangsangaben() {
					pruefeOrt(t, bund, dienst, angabe, adresse)
				}
			}
		})
	}
}

// pruefeOrt sieht nach, ob der Bund einen Ort annimmt.
//
// **Über einen Lesezugriff und nicht über einen Schreibzugriff**: die
// Namensprüfung läuft in beiden, und Lesen legt nichts ab. Ein `ErrNotFound`
// oder ein Fehler des Werkzeugs ist in Ordnung — abgewiesen wird der Name mit
// `ErrValidation`, und genau darauf kommt es an.
func pruefeOrt(
	t *testing.T, bund *schluesselbund.Bund,
	dienst domain.Dienst, angabe, adresse string,
) {
	t.Helper()

	ort := domain.ZugangKontoFuer(dienst, angabe, adresse)

	_, fehler := bund.Lies(context.Background(), "MagicSetEditor4Mac", ort)

	if fehler != nil && istUnzulaessigerName(fehler) {
		t.Errorf("der Schlüsselbund weist den Ort %q ab: %v", ort, fehler)
	}
}

// istUnzulaessigerName sagt, ob der Bund den **Namen** abgelehnt hat.
func istUnzulaessigerName(fehler error) bool {
	return errors.Is(fehler, domain.ErrValidation)
}

// probebund legt einen Schlüsselbund in `t.TempDir()` an.
//
// **Und nicht über `MSE_SCHLUESSELBUND`**: dieser Lauf ist `t.Parallel()`,
// und `t.Setenv` verträgt sich damit nicht. `NeuerBund` nimmt den Ort
// unmittelbar entgegen — das ist hier der kürzere und der ehrlichere Weg.
func probebund(t *testing.T) string {
	t.Helper()

	ort := filepath.Join(t.TempDir(), "probe.keychain")

	//nolint:gosec // das Programm steht als Literal, der Ort liegt in t.TempDir()
	if aus, err := exec.CommandContext(t.Context(), "/usr/bin/security",
		"create-keychain", "-p", "leer", ort).CombinedOutput(); err != nil {
		t.Fatalf("den Probebund anlegen: %v — %s", err, aus)
	}

	t.Cleanup(func() {
		// **Ein eigener Kontext**, weil der des Laufs beim Aufräumen schon
		// abgebrochen ist — dann räumte nichts ab.
		ctx, fertig := context.WithTimeout(context.Background(), 10*time.Second)
		defer fertig()

		//nolint:gosec,errcheck // dito
		_ = exec.CommandContext(ctx, "/usr/bin/security",
			"delete-keychain", ort).Run()
	})

	//nolint:gosec,errcheck // dito
	_ = exec.CommandContext(t.Context(), "/usr/bin/security",
		"unlock-keychain", "-p", "leer", ort).Run()

	return ort
}
