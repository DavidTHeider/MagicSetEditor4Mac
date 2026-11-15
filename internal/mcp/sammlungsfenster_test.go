package mcp_test

// **Der Ausschnitt der Sammlung über MCP** — Werkstück
// #1500.
//
// # Warum das eine eigene Zusage ist
//
// Das Werkzeug **verspricht** in seinem Schema fünf Angaben — Suche,
// Ausführung, Ordnung, Anfang und Anzahl. Ein Kopf, der mehr verspricht als
// sein Rumpf hält, ist schlimmer als gar keiner (§10): dass die Felder im
// Schema stehen, sagt nichts darüber, ob sie an der Naht ankommen. Hier steht,
// dass sie es tun.

import (
	"context"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// merkendeSammlung hält fest, womit das Fenster gerufen wurde.
type merkendeSammlung struct {
	*sammlungsprobe

	gerufen                        bool
	suche, ausfuehrung, sortierung string
	ab, anzahl                     int
}

func (m *merkendeSammlung) Sammlungsfenster(
	ctx context.Context, suche, ausfuehrung, sortierung string, ab, anzahl int,
) (fenster []mcp.Exemplarsicht, zahlen domain.Fensterzahlen, fehler error) {
	m.gerufen = true
	m.suche, m.ausfuehrung, m.sortierung = suche, ausfuehrung, sortierung
	m.ab, m.anzahl = ab, anzahl

	return m.sammlungsprobe.Sammlungsfenster(ctx, suche, ausfuehrung, sortierung,
		ab, anzahl)
}

// **Jede Angabe des Schemas kommt an der Naht an.**
func TestDieAngabenDesSammlungswerkzeugsKommenAn(t *testing.T) {
	t.Parallel()

	merkend := &merkendeSammlung{sammlungsprobe: neueSammlungsprobe()}

	antwort, err := rufeSammlungswerkzeug(t, merkend, neuePreisprobe(),
		"lesen.sammlung",
		`{"suche":"blitz","ausfuehrung":"foil","sortierung":"menge","ab":2,"anzahl":5}`)
	if err != nil {
		t.Fatalf("lesen.sammlung: %v", err)
	}

	if merkend.suche != "blitz" || merkend.ausfuehrung != "foil" ||
		merkend.sortierung != "menge" || merkend.ab != 2 || merkend.anzahl != 5 {
		t.Errorf("an der Naht kamen an: %q %q %q %d %d",
			merkend.suche, merkend.ausfuehrung, merkend.sortierung,
			merkend.ab, merkend.anzahl)
	}

	// **Beide Zahlen stehen in der Antwort** — ein Modell, das einen
	// Ausschnitt bekommt, muss wissen, wovon.
	for _, nennt := range []string{`"gesamt"`, `"insgesamt"`, `"stueck"`} {
		if !strings.Contains(antwort, nennt) {
			t.Errorf("die Antwort nennt %s nicht: %s", nennt, antwort)
		}
	}
}

// **Ohne Angaben kommt alles** — der Vertrag, den es vor #1500 gab (§16).
func TestDasSammlungswerkzeugOhneAngabenGibtAlles(t *testing.T) {
	t.Parallel()

	merkend := &merkendeSammlung{sammlungsprobe: neueSammlungsprobe()}

	antwort, err := rufeSammlungswerkzeug(t, merkend, neuePreisprobe(),
		"lesen.sammlung", `{}`)
	if err != nil {
		t.Fatalf("lesen.sammlung: %v", err)
	}

	if merkend.anzahl != 0 {
		t.Errorf("ohne Angabe kam anzahl als %d an", merkend.anzahl)
	}

	if !strings.Contains(antwort, "probe-druck") {
		t.Errorf("die Antwort ist %s", antwort)
	}
}

// **Ein kaputter Rumpf kommt nicht bis zur Naht.**
//
// Vor #1500 nahm `lesen.sammlung` gar keine Angaben; seit es fünf sind, gibt
// es diesen Weg — und ein Weg, den niemand betritt, ist keiner (§3).
func TestEinKaputterRumpfAmSammlungswerkzeugWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	merkend := &merkendeSammlung{sammlungsprobe: neueSammlungsprobe()}

	_, err := rufeSammlungswerkzeug(t, merkend, neuePreisprobe(),
		"lesen.sammlung", `{"ab":`)
	if err == nil {
		t.Fatal("der kaputte Rumpf kam durch")
	}

	if merkend.gerufen {
		t.Error("die Naht wurde trotzdem gerufen")
	}
}

// **Und die Nullprobe kennt das Fenster auch.**
//
// Sie steht hier und nicht bei ihrer Schwesterschaft in `sammlung_test.go`:
// jene Datei steht an der Zeilengrenze (§2), und dieses Werkstück ist der
// Grund, warum es die Methode gibt.
func (nilprobe) Sammlungsfenster(
	context.Context, string, string, string, int, int,
) (fenster []mcp.Exemplarsicht, zahlen domain.Fensterzahlen, fehler error) {
	return nil, domain.Fensterzahlen{}, nil
}
