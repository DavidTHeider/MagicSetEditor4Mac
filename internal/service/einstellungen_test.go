package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einstellungsablage ist die Naht zu `storage`, im Arbeitsspeicher.
type einstellungsablage struct {
	werte  map[string]string
	fehler error
}

func (e *einstellungsablage) Alle(context.Context) (map[string]string, error) {
	if e.fehler != nil {
		return nil, e.fehler
	}

	return e.werte, nil
}

func (e *einstellungsablage) Setzen(_ context.Context, schluessel, wert string) error {
	if e.fehler != nil {
		return e.fehler
	}

	if e.werte == nil {
		e.werte = make(map[string]string)
	}

	e.werte[schluessel] = wert

	return nil
}

// **Was nicht gesetzt ist, kommt als Vorgabewert** (#245).
//
// Die Alternative wäre, die Lücke der Oberfläche zu überlassen — dann stünde
// die Regel „ohne Wahl gilt das System" dort, wo kein Go-Test sie hält, und
// der MCP-Server bekäme eine andere Antwort als das Fenster.
func TestOhneWahlGeltenDieVorgabewerte(t *testing.T) {
	t.Parallel()

	verwalter := service.NeuerEinstellungsverwalter(&einstellungsablage{})

	werte, err := verwalter.Alle(t.Context())
	if err != nil {
		t.Fatalf("%v", err)
	}

	if werte[domain.EinstellungErscheinungsbild] != "system" {
		t.Errorf("%v", werte)
	}

	if werte[domain.EinstellungSchriftgroesse] != "normal" {
		t.Errorf("%v", werte)
	}

	// **Die Sprache steht nicht darin.** Sie hat keinen Vorgabewert: was ohne
	// Wahl gilt, weiß nur das Fenster (`navigator.language`).
	if _, da := werte[domain.EinstellungSprache]; da {
		t.Errorf("die Sprache hat einen Vorgabewert bekommen: %v", werte)
	}
}

// **Und eine Wahl schlägt den Vorgabewert.**
func TestEineWahlSchlaegtDenVorgabewert(t *testing.T) {
	t.Parallel()

	ablage := &einstellungsablage{werte: map[string]string{"erscheinungsbild": "dunkel"}}

	werte, err := service.NeuerEinstellungsverwalter(ablage).Alle(t.Context())
	if err != nil {
		t.Fatalf("%v", err)
	}

	if werte[domain.EinstellungErscheinungsbild] != "dunkel" {
		t.Errorf("%v", werte)
	}
}

// **Geprüft wird vor dem Schreiben, nicht danach.** Ein `grau` im Bestand
// machte die Anwendung nicht grau, sondern unbestimmt — die Klassen dafür gibt
// es nicht.
func TestEinUnerlaubterWertErreichtDenBestandNicht(t *testing.T) {
	t.Parallel()

	ablage := &einstellungsablage{}
	verwalter := service.NeuerEinstellungsverwalter(ablage)

	err := verwalter.Setzen(t.Context(), domain.EinstellungErscheinungsbild, "grau")
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("%v", err)
	}

	if len(ablage.werte) != 0 {
		t.Errorf("geschrieben wurde trotzdem: %v", ablage.werte)
	}
}

func TestEinErlaubterWertWirdGeschrieben(t *testing.T) {
	t.Parallel()

	ablage := &einstellungsablage{}

	if err := service.NeuerEinstellungsverwalter(ablage).
		Setzen(t.Context(), domain.EinstellungSchriftgroesse, "gross"); err != nil {
		t.Fatalf("%v", err)
	}

	if ablage.werte[domain.EinstellungSchriftgroesse] != "gross" {
		t.Errorf("%v", ablage.werte)
	}
}

// **Ein Lesefehler wird durchgereicht**, nicht durch Vorgabewerte verdeckt.
// Eine Anwendung, die nach einem Plattenfehler „alles auf Vorgabe" zeigt, sagt
// nicht, dass sie gerade nichts weiß.
func TestEinLesefehlerWirdGemeldet(t *testing.T) {
	t.Parallel()

	ablage := &einstellungsablage{fehler: errors.New("die Platte schweigt")}

	if _, err := service.NeuerEinstellungsverwalter(ablage).Alle(t.Context()); err == nil {
		t.Error("kein Fehler")
	}
}

func TestEinSchreibfehlerWirdGemeldet(t *testing.T) {
	t.Parallel()

	ablage := &einstellungsablage{fehler: errors.New("die Platte schweigt")}

	if err := service.NeuerEinstellungsverwalter(ablage).
		Setzen(t.Context(), domain.EinstellungSprache, "en"); err == nil {
		t.Error("kein Fehler")
	}
}

// **Jede Einstellung kommt durch — auch die vierte** (Zweitblick über Phase G).
//
// Hier stand eine **zweite** Schlüsselliste mit drei Einträgen, während die
// Domäne vier kannte. `bildkonflikt` fiel damit heraus: die Antwort trug ihn
// mit **leerem** Wert, `service.Namenskonflikt` deutete das leere Verhalten
// wie „überschreiben", und ein ausdrückliches `altbehalten` wurde
// stillschweigend übergangen — bei zwei Karten desselben Namens ging eine
// verloren.
//
// **Der Test fragt die Domäne**, nicht eine dritte Liste: er soll rot werden,
// wenn eine fünfte Einstellung dazukommt und hier nicht ankommt.
func TestJedeEinstellungDerDomaeneKommtDurch(t *testing.T) {
	t.Parallel()

	werte, err := service.NeuerEinstellungsverwalter(&einstellungsablage{}).Alle(t.Context())
	if err != nil {
		t.Fatalf("%v", err)
	}

	for _, schluessel := range domain.Einstellungsreihe() {
		// **Erwartet wird, was ohne jede Wahl **gilt**, nicht was in der
		// Profilzeile steht** — seit Werkstück 9 (#1351, Punkt 9) sind das
		// zwei Dinge: dort steht `erben`, und was gilt, ist die Vorgabe des
		// Menschen. Die Verschmelzung fragt die Domäne, damit dieser Lauf
		// nicht seine eigene Antwort erfindet.
		gilt := domain.Verschmolzen(
			schluessel, domain.EinstellungsVorgabe(schluessel), "")
		if gilt == "" {
			// Die Sprache hat keinen — was ohne Wahl gilt, weiß nur das Fenster.
			continue
		}

		if werte[schluessel] != gilt {
			t.Errorf("%s = %q, erwartet %q", schluessel, werte[schluessel], gilt)
		}
	}
}
