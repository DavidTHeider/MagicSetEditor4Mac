package service_test

// **Derselbe Anfang, dieselbe Folge, derselbe Zustand** — Werkstück G1 von
// Phase 5.5
// (#1458).
//
// # Die erste Abnahme, und sie ist die wichtigste
//
// **Ohne sie gibt es keinen Wiedereinstieg und kein Replay** — und beides
// verlangen die Werkstücke, die auf diesem stehen. Geprüft wird sie so, wie
// sie dasteht: zwei Leitungen, derselbe Startabzug, dieselbe Befehlsfolge,
// und danach der Vergleich über **Zustand und Projektion**.

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// dieFolge ist die Befehlsfolge, die beide Läufe fahren.
func dieFolge() []domain.Spielbefehl {
	return []domain.Spielbefehl{
		{
			Kennung: "b1", PartieID: "p1", Absender: "sitz-1",
			Art:      service.BefehlSchieben,
			Nutzlast: map[string]any{"objekt": "o1", "zone": "tisch"},
		},
		{
			Kennung: "b2", PartieID: "p1", Absender: "sitz-1",
			Art:      service.BefehlTappen,
			Nutzlast: map[string]any{"objekt": "o1"},
		},
		{
			Kennung: "b3", PartieID: "p1", Absender: "sitz-1",
			Art: service.BefehlZaehler,
			Nutzlast: map[string]any{
				"objekt": "o1", "art": "+1/+1", "menge": float64(2),
			},
		},
		{
			Kennung: "b4", PartieID: "p1", Absender: "sitz-2",
			Art:      service.BefehlUmdrehen,
			Nutzlast: map[string]any{"objekt": "o1"},
		},
	}
}

// fahre legt eine Partie an und spielt die Folge.
func fahre(t *testing.T) (*partieprobe, domain.Partiesicht) {
	t.Helper()

	leitung, speicher := eineLeitung(t)
	ctx := context.Background()

	// **Derselbe Anfang heißt auch dasselbe Geheimnis** (#2480) — sonst
	// trügen zwei Läufe verschiedene Sichtkennungen, und der Vergleich
	// fände den Zufall statt der Folge.
	leitung.MitGeheimnisquelle(func() string { return "derselbe-anfang" })

	if _, fehler := leitung.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	for _, befehl := range dieFolge() {
		if _, fehler := leitung.Befehl(ctx, befehl); fehler != nil {
			t.Fatalf("%s: %v", befehl.Kennung, fehler)
		}
	}

	sicht, fehler := leitung.Partie(ctx, "p1", "sitz-1")
	if fehler != nil {
		t.Fatalf("Partie: %v", fehler)
	}

	return speicher, sicht
}

// **Zwei Läufe, ein Ergebnis.**
func TestDerselbeAnfangUndDieselbeFolgeGebenDenselbenZustand(t *testing.T) {
	t.Parallel()

	ersterSpeicher, ersteSicht := fahre(t)
	zweiterSpeicher, zweiteSicht := fahre(t)

	if !reflect.DeepEqual(ersteSicht, zweiteSicht) {
		t.Errorf("die Projektionen unterscheiden sich:\n%+v\n%+v",
			ersteSicht, zweiteSicht)
	}

	if !reflect.DeepEqual(
		ersterSpeicher.partien["p1"], zweiterSpeicher.partien["p1"]) {
		t.Error("die Zustände unterscheiden sich")
	}

	// **Und die Ereignisse tragen dieselbe Kette** — sie hängt am Inhalt
	// und an der Folge, nicht an der Uhr oder an einem Zufall.
	ersterStrom := ersterSpeicher.ereignisse["p1"]
	zweiterStrom := zweiterSpeicher.ereignisse["p1"]

	if len(ersterStrom) != len(dieFolge()) {
		t.Fatalf("der Strom hat %d Ereignisse", len(ersterStrom))
	}

	for i := range ersterStrom {
		if ersterStrom[i].Kette != zweiterStrom[i].Kette {
			t.Errorf("Ereignis %d hat zwei Ketten: %q und %q",
				i, ersterStrom[i].Kette, zweiterStrom[i].Kette)
		}
	}
}

// **Jedes Ereignis trägt den Abdruck seines Vorgängers.**
//
// **Das ist die Kette, und sie wird hier nachgerechnet** — nicht geglaubt.
// Ein Strom, dessen Glieder nicht aufeinander zeigen, ist eine Liste.
func TestDerStromIstEineKette(t *testing.T) {
	t.Parallel()

	speicher, _ := fahre(t)
	strom := speicher.ereignisse["p1"]

	vorher := ""

	for i, ereignis := range strom {
		if ereignis.Folge != int64(i+1) {
			t.Errorf("Ereignis %d trägt die Folge %d", i, ereignis.Folge)
		}

		gerechnet, fehler := ereignis.Abdruck(vorher)
		if fehler != nil {
			t.Fatalf("Abdruck: %v", fehler)
		}

		if gerechnet != ereignis.Kette {
			t.Errorf("Ereignis %d hängt nicht an seinem Vorgänger", i)
		}

		vorher = ereignis.Kette
	}

	// **Die Gegenprobe**: ein verändertes Ereignis bricht die Kette.
	verfaelscht := strom[1]
	verfaelscht.Oeffentlich = map[string]any{"objekt": "o2"}

	gerechnet, fehler := verfaelscht.Abdruck(strom[0].Kette)
	if fehler != nil {
		t.Fatalf("Abdruck: %v", fehler)
	}

	if gerechnet == strom[1].Kette {
		t.Error("ein verändertes Ereignis hat denselben Abdruck")
	}
}

// **Was der Sandkasten erlaubt — und wem.**
func TestWasDerSandkastenErlaubt(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineLeitung(t)
	ctx := context.Background()

	if _, fehler := leitung.Anlegen(ctx, einePartie()); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	erlaubt, fehler := erlaubtVon(leitung, ctx, "p1", "sitz-1")
	if fehler != nil {
		t.Fatalf("Erlaubnis: %v", fehler)
	}

	// **Zwei weniger als die ganze Liste**, und das ist die Zusage: die
	// Legendenwahl (CR 704.5j) und die Reihung (CR 603.3b) sind **Antworten
	// auf eine offene Frage**, und in dieser Partie steht keine. Ein Befehl,
	// den der Kern danach abwiese, wäre eine Einladung, die er
	// zurücknimmt — der Grund steht bei `service.ohneUngestellteFragen`.
	//
	// **Und ohne die vier des Beginns** (#1839): diese Partie ist von Hand
	// gebaut, ihre Sitze tragen keinen Stand — wie eine von vorher, und die
	// beginnt nicht noch einmal.
	//
	// **Und mit der Rücknahme** (P8, #2376): sie steht nicht in der Liste
	// des Sandkastens, denn sie meint den Strom, nicht den Stand.
	if len(erlaubt) != len(service.Sandkastenbefehle())-2-
		len(service.Eroeffnungsbefehle())+1 {
		t.Errorf("einem Sitz sind %d Befehle erlaubt", len(erlaubt))
	}

	// **Ein Zuschauer darf nichts** — und bekommt eine leere Liste, kein
	// `nil`: „nichts geht" ist eine Antwort.
	fuerZuschauer, fehler := erlaubtVon(leitung, ctx, "p1", "z1")
	if fehler != nil {
		t.Fatalf("Erlaubnis: %v", fehler)
	}

	if fuerZuschauer == nil || len(fuerZuschauer) != 0 {
		t.Errorf("dem Zuschauer sind %v erlaubt", fuerZuschauer)
	}

	// **Und eine pausierte Partie erlaubt niemandem etwas.**
	speicher.partien["p1"] = mitStand(speicher.partien["p1"], domain.StandPausiert)

	pausiert, fehler := erlaubtVon(leitung, ctx, "p1", "sitz-1")
	if fehler != nil {
		t.Fatalf("Erlaubnis: %v", fehler)
	}

	if pausiert == nil || len(pausiert) != 0 {
		t.Errorf("in der Pause sind %v erlaubt", pausiert)
	}
}

// **Was der Sandkasten abweist.**
func TestWasDerSandkastenAbweist(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		art      string
		nutzlast map[string]any
		ausgang  error
	}{
		"ohne Objekt": {
			service.BefehlTappen, map[string]any{}, domain.ErrValidation,
		},
		"auf ein Objekt, das es nicht gibt": {
			service.BefehlTappen,
			map[string]any{"objekt": "o9"},
			domain.ErrNotFound,
		},
		"in eine Zone, die es nicht gibt": {
			service.BefehlSchieben,
			map[string]any{"objekt": "o1", "zone": "keller"},
			domain.ErrValidation,
		},
		"mit einem namenlosen Zähler": {
			service.BefehlZaehler,
			map[string]any{"objekt": "o1", "menge": float64(1)},
			domain.ErrValidation,
		},
		"mit einem Anhängen ins Leere": {
			service.BefehlAnhaengen,
			map[string]any{"objekt": "o1"},
			domain.ErrValidation,
		},
		"mit einer erfundenen Art": {
			"objekt.verzaubern",
			map[string]any{"objekt": "o1"},
			domain.ErrValidation,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			leitung, speicher := eineLeitung(t)
			ctx := context.Background()

			if _, fehler := leitung.Anlegen(ctx, einePartie()); fehler != nil {
				t.Fatalf("Anlegen: %v", fehler)
			}

			_, fehler := leitung.Befehl(ctx, domain.Spielbefehl{
				Kennung: "b1", PartieID: "p1", Absender: "sitz-1",
				Art: fall.art, Nutzlast: fall.nutzlast,
			})

			if !errors.Is(fehler, fall.ausgang) {
				t.Fatalf("es kam %v", fehler)
			}

			// **Und geschrieben wurde nichts** — ein abgewiesener Befehl
			// hinterlässt keinen halben Stand.
			if speicher.angewendet != 0 {
				t.Error("der abgewiesene Befehl hat geschrieben")
			}
		})
	}
}
