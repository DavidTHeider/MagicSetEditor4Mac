package flaechen_test

// **Der Neuanfang an der Kommandozeile, mit einer Freigabe vom Fenster** —
// G12 (#2282).
//
// # Was er hält
//
// **Den Weg, den es bis G12 nicht gab.** `mse zuruecksetzen anwendung` nahm
// eine Nachweiskennung entgegen, die keine Fläche ausstellte; gefahren war er
// nur über eine in die Ablage gelegte. Hier stellt das Fenster aus — über
// denselben Bestand, den der Kindprozess benutzt, und ohne Dienstmodus —, und
// die Kommandozeile löst ein. **Einmal**: der zweite Versuch mit derselben
// Kennung scheitert.
//
// # Was er nicht hält
//
// **Das Fenster selbst** — dieser Lauf ruft die Bindung im Prozess. Dass die
// Seite den Knopf und die Kennung zeigt, hält
// `frontend/src/seiten/ruecksetzungsfreigabe.test.tsx`.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestDerNeuanfangGehtMitEinerFreigabeVomFenster(t *testing.T) {
	t.Parallel()

	umgebung := deckumgebung(t)

	// **Der Bestand entsteht beim ersten Kindprozess** — davor gäbe es nichts,
	// worin das Fenster ausstellen könnte.
	mussKommandozeile(t, umgebung, "sets")

	// **Erst die Freigabe, dann die Vorschau**: der Nachweis ist eine Zeile
	// im Bestand, und die Vorschau zählt Freigabenachweise mit — ihre Marke
	// muss den Stand nach dem Ausstellen tragen.
	kennung := neuanfangsfreigabe(t, umgebung)

	aus, klage, code := kommandozeile(t, umgebung, "zuruecksetzen", "anwendung", "", kennung,
		"--revision", vorschaurevision(t, umgebung))
	if code != 0 {
		t.Fatalf("zuruecksetzen gab %d (%s / %s)", code, klage, aus)
	}

	if !strings.Contains(aus, "beiseite\t") {
		t.Errorf("die Ausgabe sagt nicht, wo der alte Bestand liegt: %q", aus)
	}

	// **Verbraucht ist verbraucht** — derselbe Nachweis geht kein zweites Mal.
	if _, _, code := kommandozeile(t, umgebung, "zuruecksetzen", "anwendung", "", kennung,
		"--revision", vorschaurevision(t, umgebung)); code == 0 {
		t.Error("derselbe Nachweis ging ein zweites Mal durch")
	}
}

// neuanfangsfreigabe stellt den Nachweis aus — **am Fenster**, wie
// [eineFreigabeFuer] den für einen Handelsschritt.
func neuanfangsfreigabe(t *testing.T, umgebung []string) string {
	t.Helper()

	ort := bestandsortAus(t, umgebung)

	db, fehler := storage.Oeffnen(t.Context(), ort)
	if fehler != nil {
		t.Fatalf("den Bestand öffnen: %v", fehler)
	}

	defer func() { _ = db.Close() }() //nolint:errcheck // nur zum Aufräumen

	ctx, profilfehler := storage.MitGezeigtemProfil(t.Context(), db)
	if profilfehler != nil {
		t.Fatalf("das Profil lesen: %v", profilfehler)
	}

	antwort := app.RuecksetzungsAnwendungUeber(db, ort).RuecksetzungFreigabeErteilen(ctx)
	if antwort.Fehler != nil || antwort.Kennung == "" {
		t.Fatalf("die Freigabe: %+v", antwort)
	}

	return antwort.Kennung
}

// vorschaurevision holt die Marke der Vorschau, an die die Ausführung
// gebunden ist.
func vorschaurevision(t *testing.T, umgebung []string) string {
	t.Helper()

	aus, _, _ := kommandozeile(t, umgebung, "ruecksetzung-vorschau", "anwendung")

	for zeile := range strings.SplitSeq(aus, "\n") {
		if marke, gefunden := strings.CutPrefix(zeile, "revision\t"); gefunden {
			return strings.TrimSpace(marke)
		}
	}

	t.Fatalf("die Vorschau nennt keine Revision:\n%s", aus)

	return ""
}
