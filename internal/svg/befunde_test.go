// **Was fehlt, wird genannt** —
// #852.
//
// Bis zum 2026-08-22 gab `AusPlan` nur `(string, error)` und ließ eine
// unlesbare Ebene kommentarlos weg. Ein Test namens
// `TestOhneBildquelleBleibtEinBefund` prüfte, dass **kein** `<image>`
// vorkommt — es gab gar keinen Kanal für einen Befund, und eine SVG ohne
// Kartenrahmen entstand erfolgreich.
package svg_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/svg"
)

// **Ohne Bildquelle bleibt ein Befund — und diesmal einer, den es gibt.**
//
// Der Test dieses Namens prüfte bis zum 2026-08-22, dass **kein** `<image>`
// vorkommt, während sein Kopf behauptete, der Schreiber „sagt es über einen
// Befund". `AusPlan` gab damals nur `(string, error)`, und es gab gar keinen
// Kanal für einen Befund. Angestrichen vom Zweitblick über Phase P.
func TestOhneBildquelleBleibtEinBefund(t *testing.T) {
	t.Parallel()

	geschrieben, befunde, fehler := svg.AusPlan(t.Context(), probeplan(),
		svg.Quellen{Schriften: suchprobe{}})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if strings.Contains(geschrieben, "<image") {
		t.Error("ein Bild wurde geschrieben, obwohl es keine Quelle gibt")
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "Bildquelle") {
		t.Fatalf("kein Befund über die fehlende Bildquelle: %v", befunde)
	}

	// **Und er nennt sein Feld**, nicht nur seine Nummer — bei rund vierzig
	// Ebenen je Karte ist das der Unterschied zwischen einer Meldung und einer
	// Suche.
	if befunde[0].Feld != "rahmen" {
		t.Errorf("der Befund nennt %q statt des Feldes", befunde[0].Feld)
	}
}

// **Ein Bild, das die Quelle nicht kennt, wird genannt und ausgelassen.**
func TestEinFehlendesBildWirdGenannt(t *testing.T) {
	t.Parallel()

	geschrieben, befunde, fehler := svg.AusPlan(t.Context(), probeplan(),
		svg.Quellen{Baeume: baumprobe{fehlt: true}, Schriften: suchprobe{}})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if strings.Contains(geschrieben, "<image") {
		t.Error("ein Bild wurde geschrieben, obwohl die Quelle keines hat")
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet war einer: %v", len(befunde), befunde)
	}

	// **Und der Rest steht trotzdem da** (§8).
	if !strings.Contains(geschrieben, ">Feuerball<") {
		t.Error("der Text fehlt, obwohl nur das Bild fehlte")
	}
}

// **Eine Warnung reist mit, auch wenn ein Bild herauskommt.**
//
// Eine noch nicht gebaute Verknüpfung liefert ihr Ausgangsbild samt Meldung.
// Wer sie verwürfe, schriebe eine halb falsche Karte, ohne dass jemand etwas
// sagt — dieselbe Klasse, die am Rasterer zweimal gefunden wurde.
func TestEineWarnungReistMitDemBildMit(t *testing.T) {
	t.Parallel()

	geschrieben, befunde, fehler := svg.AusPlan(t.Context(), probeplan(),
		svg.Quellen{Baeume: baumprobe{warnt: true}, Schriften: suchprobe{}})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if !strings.Contains(geschrieben, "<image") {
		t.Error("das Bild fehlt, obwohl die Quelle eines hergab")
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "Verknüpfung") {
		t.Errorf("die Warnung ist verlorengegangen: %v", befunde)
	}
}

// **Ein Kodiererfehler wird genannt, statt verschluckt zu werden.**
func TestEinKodiererfehlerWirdGenannt(t *testing.T) {
	t.Parallel()

	_, befunde, fehler := svg.AusPlan(t.Context(), probeplan(),
		svg.Quellen{Baeume: baumprobe{ohneFlaeche: true}, Schriften: suchprobe{}})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "einbetten") {
		t.Errorf("der Kodiererfehler wurde verschluckt: %v", befunde)
	}
}

// **Eine fehlende Schriftsuche wird genannt.**
func TestOhneSchriftsucheBleibtEinBefund(t *testing.T) {
	t.Parallel()

	geschrieben, befunde, fehler := svg.AusPlan(t.Context(), probeplan(),
		svg.Quellen{Baeume: baumprobe{}})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if strings.Contains(geschrieben, "<text") {
		t.Error("Text wurde geschrieben, obwohl es keine Schriftsuche gibt")
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "Schriftsuche") {
		t.Errorf("kein Befund über die fehlende Schriftsuche: %v", befunde)
	}
}

// **Eine Schrift, die es nicht gibt, wird genannt und nicht ersetzt.**
//
// Eine Karte mit falscher Schrift sieht richtig aus und ist es nicht —
// dieselbe Haltung wie in `render.text`.
func TestEineFehlendeSchriftWirdGenannt(t *testing.T) {
	t.Parallel()

	_, befunde, fehler := svg.AusPlan(t.Context(), probeplan(),
		svg.Quellen{Baeume: baumprobe{}, Schriften: suchprobe{fehlt: true}})
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "nicht zu finden") {
		t.Errorf("kein Befund über die fehlende Schrift: %v", befunde)
	}
}

// **Eine Textebene ohne Schriftnamen wird genannt.**
func TestEineTextebeneOhneSchriftnamenWirdGenannt(t *testing.T) {
	t.Parallel()

	plan := probeplan()
	plan.Ebenen[2].Satz.Schrift = "  "

	_, befunde, fehler := svg.AusPlan(t.Context(), plan, mitAllem())
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "ohne Schriftnamen") {
		t.Errorf("kein Befund über den fehlenden Schriftnamen: %v", befunde)
	}
}
