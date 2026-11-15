//go:build darwin

package schluesselbund_test

// **Der Schlüsselbund** — Werkstück 2 von Phase 4.8
// (#1321).
//
// **Jeder Lauf hier bekommt einen eigenen Bund**, angelegt in `t.TempDir()`
// und danach gelöscht. Ein Test, der in den **echten** Anmeldebund schriebe,
// hinterließe einen Eintrag auf dem Rechner der Entwickelnden und
// stolperte über den Zugriffsdialog von macOS — dieselbe Regel wie in §3
// für einen Test, der nach draußen greift.

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schluesselbund"
)

// **Ein Geheimnis geht hinein und kommt unversehrt zurück.**
func TestEinGeheimnisLaeuftRund(t *testing.T) {
	t.Parallel()

	bund := probebund(t)

	for _, fall := range []struct {
		name     string
		konto    string
		geheimis string
	}{
		{"schlicht", "schlicht", "abc123"},
		{"mit Leerzeichen", "leerzeichen", "a b c"},

		// **Die drei, an denen der naheliegende Weg bricht.** `security` liest
		// im Zwiegespräch eine Befehlszeile und zerteilt sie selbst; ein
		// Geheimnis mit Leerzeichen kam als **leer** zurück, eines mit
		// Gegenschrägstrich verstümmelt. Gemessen am 2026-09-01.
		{"mit Gegenschrägstrich", "schraeg", `a\b`},
		{"mit Anführungszeichen", "anfuehrung", `a"b`},
		{"mit Zeilenumbruch", "umbruch", "a\nb"},

		{"mit Umlaut", "umlaut", "schlüssel-öäü"},
		{"lang", "lang", strings.Repeat("k", 1024)},
	} {
		t.Run(fall.name, func(t *testing.T) {
			if err := bund.Schreibe(t.Context(), "mse-probe", fall.konto,
				fall.geheimis); err != nil {
				t.Fatalf("schreiben: %v", err)
			}

			zurueck, err := bund.Lies(t.Context(), "mse-probe", fall.konto)
			if err != nil {
				t.Fatalf("lesen: %v", err)
			}

			if zurueck != fall.geheimis {
				t.Errorf("zurück kam %q, erwartet %q", zurueck, fall.geheimis)
			}
		})
	}
}

// **Was nicht da ist, ist nicht da** — und sagt es als fachliche Auskunft.
func TestEinFehlenderEintragIstNichtGefunden(t *testing.T) {
	t.Parallel()

	_, err := probebund(t).Lies(t.Context(), "mse-probe", "gibtsnicht")

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}
}

// **Ein zweites Schreiben ersetzt das erste.**
//
// Ohne das läge im Bund ein zweiter Eintrag mit demselben Namen, und welcher
// gelesen wird, entschiede die Reihenfolge im Bund.
func TestEinZweitesSchreibenErsetzt(t *testing.T) {
	t.Parallel()

	bund := probebund(t)

	for _, wort := range []string{"erstes", "zweites"} {
		if err := bund.Schreibe(t.Context(), "mse-probe", "ersetzt", wort); err != nil {
			t.Fatalf("schreiben %q: %v", wort, err)
		}
	}

	zurueck, err := bund.Lies(t.Context(), "mse-probe", "ersetzt")
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if zurueck != "zweites" {
		t.Errorf("zurück kam %q, erwartet »zweites«", zurueck)
	}
}

// **Ein Geheimnis lässt sich herausnehmen.**
func TestEinGeheimnisLaesstSichLoeschen(t *testing.T) {
	t.Parallel()

	bund := probebund(t)

	if err := bund.Schreibe(t.Context(), "mse-probe", "weg", "geheim"); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if err := bund.Loesche(t.Context(), "mse-probe", "weg"); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	if _, err := bund.Lies(t.Context(), "mse-probe", "weg"); !errors.Is(
		err, domain.ErrNotFound) {
		t.Errorf("nach dem Löschen ist der Fehler %v, erwartet ErrNotFound", err)
	}
}

// **Ein leeres Geheimnis ist keines.**
//
// Es wird abgewiesen, nicht abgelegt. Zwei Gründe, und der zweite ist der
// harte: „kein Schlüssel eingetragen" und „ein Schlüssel, der leer ist" wären
// im Bund nicht mehr zu unterscheiden, und Vermerk 0063 baut auf genau dieser
// Unterscheidung auf. Dazu kommt, dass `security` eine leere Angabe hinter
// `-w` gar nicht als solche liest — es nimmt das **nächste** Argument dafür
// (gemessen am 2026-09-01).
//
// Wer nichts mehr hinterlegt haben will, ruft [Bund.Loesche].
func TestEinLeeresGeheimnisWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	err := probebund(t).Schreibe(t.Context(), "mse-probe", "leer", "")

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler ist %v, erwartet eine Validierung", err)
	}
}

// **Ein zu langes Geheimnis wird abgewiesen und nicht verstümmelt.**
//
// **Gemessen am 2026-09-01**: über den Zwiegesprächsweg laufen 1.024 Zeichen
// rund, 2.048 nicht — dort kam **nichts** zurück. Wo genau dazwischen die
// Grenze liegt, ist **nicht** nachgemessen; sie ist vermutlich die Länge der
// Befehlszeile und nicht die des Geheimnisses.
//
// Ein Geheimnis, das still verschwindet, wäre der schlechteste Ausgang: der
// Dienst bliebe „an", und niemand wüsste warum er nicht antwortet.
func TestEinZuLangesGeheimnisWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	err := probebund(t).Schreibe(t.Context(), "mse-probe", "lang",
		strings.Repeat("k", schluesselbund.Hoechstlaenge+1))

	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("Fehler ist %v, erwartet eine Validierung", err)
	}
}

// **Ein Aufruf, der nicht antwortet, endet trotzdem.**
//
// # Der Fall, und er ist gemessen
//
// `security` fragt bei manchen Fehlschlägen nach einem Passwort — und wartet
// dann. **Gemessen am 2026-09-02**: ein Schreiben in einen Bund, den es nicht
// gibt, bei **leerem `HOME`** kam nie zurück. Genau so läuft `make ci` (die
// Stufe `test` setzt ein leeres Zuhause, damit kein Rechnerzustand
// einsickert), und der Lauf stand nach 600 Sekunden.
//
// **Ein Unterprozess ohne Frist ist derselbe Fehler wie `http.DefaultClient`
// ohne Timeout** (§16, Regel 3): im Betrieb hinge das Fenster, und abbrechen
// könnte nur, wer die Anwendung beendet.
func TestEinHaengenderAufrufEndetTrotzdem(t *testing.T) {
	// **Kein `t.Parallel`** — `t.Setenv` und Nebenläufigkeit gehen nicht
	// zusammen, und das leere Zuhause ist der Punkt dieses Laufs.
	//
	// **Leeres Zuhause** — dasselbe wie in der Pipeline.
	t.Setenv("HOME", t.TempDir())

	bund := schluesselbund.NeuerBund(
		filepath.Join(t.TempDir(), "gibtsnicht.keychain")).
		MitFrist(2 * time.Second)

	fertig := make(chan error, 1)

	go func() {
		fertig <- bund.Schreibe(context.Background(), "mse", "konto", "geheim")
	}()

	select {
	case err := <-fertig:
		if err == nil {
			t.Error("der Aufruf glückte, erwartet ein Fehler")
		}
	case <-time.After(20 * time.Second):
		t.Fatal("der Aufruf kam nicht zurück — die Frist greift nicht")
	}
}

// **Ein Eintrag, den es nicht gibt, lässt sich löschen.**
//
// Wer löscht, will, dass nichts da ist — und das ist dann schon der Fall.
// Ein Fehler hier zwänge jeden Aufrufer, vorher nachzusehen, und zwei
// Aufrufer täten es unterschiedlich.
func TestEinFehlenderEintragLaesstSichLoeschen(t *testing.T) {
	t.Parallel()

	if err := probebund(t).Loesche(t.Context(), "mse-probe", "gibtsnicht"); err != nil {
		t.Errorf("löschen gab %v, erwartet nichts", err)
	}
}
