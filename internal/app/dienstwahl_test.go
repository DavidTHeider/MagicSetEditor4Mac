package app_test

// **Die Dienstwahl über den Einstellungen** — Phase 4.8
// (#1321).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// leserprobe gibt feste Einstellungen — oder einen Fehler.
type leserprobe struct {
	werte  map[string]string
	fehler error
}

func (l leserprobe) Alle(context.Context) (map[string]string, error) {
	return l.werte, l.fehler
}

// **Ohne Eintrag gilt die Vorgabe.**
func TestOhneEintragGiltDieVorgabe(t *testing.T) {
	t.Parallel()

	wahl := app.NeueEinstellungswahl(leserprobe{werte: map[string]string{}})

	for dienst, erwartet := range map[domain.Dienst]bool{
		domain.DienstScryfall: true,
		domain.DienstMTGAPI:   false,
		domain.DienstMTGJSON:  false,
		domain.DienstScrydex:  false,
	} {
		an, err := wahl.IstAn(probekontext(t), dienst)
		if err != nil {
			t.Fatalf("%s: %v", dienst, err)
		}

		if an != erwartet {
			t.Errorf("%s ist %v, erwartet %v", dienst, an, erwartet)
		}
	}
}

// **Ein leerer Wert ist kein Wert** — er heißt „nie gesetzt", und dann gilt
// die Vorgabe. Ohne diesen Zweig läse ein leerer Eintrag als „aus", und
// Scryfall wäre nach einem Schreibfehler still abgeschaltet.
func TestEinLeererWertGiltAlsUngesetzt(t *testing.T) {
	t.Parallel()

	wahl := app.NeueEinstellungswahl(leserprobe{
		werte: map[string]string{
			domain.DienstEinstellung(domain.DienstScryfall): "",
		},
	})

	an, err := wahl.IstAn(probekontext(t), domain.DienstScryfall)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if !an {
		t.Error("ein leerer Wert schaltete Scryfall ab")
	}
}

// **Ein gesetzter Wert gilt.**
func TestEinGesetzterWertGilt(t *testing.T) {
	t.Parallel()

	wahl := app.NeueEinstellungswahl(leserprobe{
		werte: map[string]string{
			domain.DienstEinstellung(domain.DienstScryfall): domain.DienstAus,
			domain.DienstEinstellung(domain.DienstScrydex):  domain.DienstAn,
		},
	})

	aus, err := wahl.IstAn(probekontext(t), domain.DienstScryfall)
	if err != nil {
		t.Fatalf("Scryfall: %v", err)
	}

	if aus {
		t.Error("Scryfall ist an, obwohl »aus« eingetragen ist")
	}

	an, scrydexfehler := wahl.IstAn(probekontext(t), domain.DienstScrydex)
	if scrydexfehler != nil {
		t.Fatalf("Scrydex: %v", scrydexfehler)
	}

	if !an {
		t.Error("Scrydex ist aus, obwohl »an« eingetragen ist")
	}
}

// **Ein Lesefehler kommt an und heißt nicht »an«.**
func TestEinLesefehlerKommtAn(t *testing.T) {
	t.Parallel()

	wahl := app.NeueEinstellungswahl(leserprobe{fehler: domain.ErrNotFound})

	an, err := wahl.IstAn(probekontext(t), domain.DienstScryfall)

	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}

	if an {
		t.Error("nach einem Lesefehler gilt der Dienst als an")
	}
}

// **`ImmerAn` sagt immer ja** — und das ist die Stelle, an der es steht.
func TestImmerAnSagtJa(t *testing.T) {
	t.Parallel()

	an, err := app.ImmerAn{}.IstAn(probekontext(t), domain.DienstScrydex)
	if err != nil {
		t.Fatalf("ImmerAn: %v", err)
	}

	if !an {
		t.Error("ImmerAn sagt nein")
	}
}

// **`Wert` gibt die Vorgabe, den Eintrag — und die Klage** (T11, #1476).
//
// **Drei Fälle und nicht einer**: ohne Eintrag gilt die Vorgabe, mit Eintrag
// der Eintrag, und ein Leser, der klemmt, klemmt — eine Vorgabe als Antwort
// auf einen Fehler sähe aus wie ein ausgeschalteter Schalter.
func TestDerWertGibtVorgabeEintragUndKlage(t *testing.T) {
	t.Parallel()

	leer := app.NeueEinstellungswahl(leserprobe{werte: map[string]string{}})

	wert, fehler := leer.Wert(probekontext(t), domain.Kontoschau)
	if fehler != nil || wert != domain.KontoschauAus {
		t.Errorf("ohne Eintrag kam %q / %v", wert, fehler)
	}

	gesetzt := app.NeueEinstellungswahl(leserprobe{werte: map[string]string{
		domain.Kontoschau: domain.KontoschauAn,
	}})

	wert, fehler = gesetzt.Wert(probekontext(t), domain.Kontoschau)
	if fehler != nil || wert != domain.KontoschauAn {
		t.Errorf("mit Eintrag kam %q / %v", wert, fehler)
	}

	klemmend := app.NeueEinstellungswahl(leserprobe{fehler: errKlemmt})

	if _, fehler := klemmend.Wert(probekontext(t), domain.Kontoschau); fehler == nil {
		t.Error("die Klage wurde verschluckt")
	}
}

// errKlemmt ist die Klage des Lesers.
var errKlemmt = errors.New("die Einstellungen sind nicht zu lesen")
