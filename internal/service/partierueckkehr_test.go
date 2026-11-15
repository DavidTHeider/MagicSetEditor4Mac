package service_test

// **Was der Wirt über seinen eigenen Stand weiß** — Werkstück G5 von Phase
// 5.5 (#1462).
//
// # Wogegen hier geprüft wird
//
// **Dass der Stand aus dem Bestand kommt und sichtbezogen ist.** Die Kette
// eines Ereignisses ist für alle gleich; was einen Sitz von einem anderen
// unterscheidet, sind die privaten Ergänzungen — und genau die gehen in den
// Abdruck ein.
//
// **Und dass die Grenze zwischen klein und groß wirklich eine ist.** Sie hieß
// im ersten Entwurf „die älteste, die noch vorliegt"; im Bestand liegt alles
// vor, und damit wäre der Abzugsweg toter Code gewesen.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// mitStrom legt eine Partie samt Ereignissen in die Probe.
func mitStrom(
	t *testing.T, strom []domain.Ereignis,
) (*service.Partieleitung, *partieprobe) {
	t.Helper()

	leitung, speicher := eineLeitung(t)

	if _, fehler := leitung.Anlegen(t.Context(), einePartie()); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if speicher.ereignisse == nil {
		speicher.ereignisse = map[string][]domain.Ereignis{}
	}

	speicher.ereignisse["p1"] = strom

	return leitung, speicher
}

// **Der Stand kommt aus dem Bestand — und der Abdruck ist sichtbezogen.**
func TestDerWirtsstandIstSichtbezogen(t *testing.T) {
	t.Parallel()

	letztes := domain.Ereignis{
		PartieID: "p1", Folge: 3, Art: "karte.ziehen", Kette: "drei",
		Privat: []domain.Ergaenzung{
			{Empfaenger: "sitz-1", Inhalt: map[string]any{"karte": "k1"}},
			{Empfaenger: "sitz-2", Inhalt: map[string]any{"karte": "k2"}},
		},
	}

	leitung, _ := mitStrom(t, []domain.Ereignis{
		{PartieID: "p1", Folge: 1, Kette: "eins"},
		{PartieID: "p1", Folge: 2, Kette: "zwei"},
		letztes,
	})

	erster, fehler := leitung.Wirtsstand(t.Context(), "p1", "sitz-1")
	if fehler != nil {
		t.Fatalf("Wirtsstand: %v", fehler)
	}

	zweiter, fehler := leitung.Wirtsstand(t.Context(), "p1", "sitz-2")
	if fehler != nil {
		t.Fatalf("Wirtsstand: %v", fehler)
	}

	switch {
	case erster.Folge != 3:
		t.Errorf("die Folge ist %d statt 3", erster.Folge)
	case erster.Epoche != domain.Partieepoche:
		t.Errorf("die Epoche ist %d", erster.Epoche)
	case erster.AelteteFolge != 1:
		t.Errorf("die älteste ist %d statt 1", erster.AelteteFolge)
	case erster.Kette == zweiter.Kette:
		t.Error("beide Sitze haben denselben Abdruck — die Ergänzung geht " +
			"nicht ein")
	}

	// **Und er stimmt mit dem überein, den der Gast selbst rechnet.** Das ist
	// die eigentliche Zusage: beide Seiten müssen auf dasselbe kommen.
	beimGast, fehler := letztes.Fuer("sitz-1").Sichtabdruck("sitz-1")
	if fehler != nil {
		t.Fatalf("Sichtabdruck: %v", fehler)
	}

	if beimGast != erster.Kette {
		t.Errorf("der Gast rechnet %q, der Wirt %q", beimGast, erster.Kette)
	}
}

// **Eine Partie ohne Ereignisse steht bei null und hat keinen Abdruck.**
func TestEinePartieOhneEreignisseStehtBeiNull(t *testing.T) {
	t.Parallel()

	leitung, _ := mitStrom(t, nil)

	stand, fehler := leitung.Wirtsstand(t.Context(), "p1", "sitz-1")
	if fehler != nil {
		t.Fatalf("Wirtsstand: %v", fehler)
	}

	if stand.Folge != 0 || stand.Kette != "" {
		t.Errorf("der Stand ist %+v", stand)
	}
}

// **Jenseits des Fensters wird nichts mehr nachgereicht.**
//
// **Das ist die Grenze zwischen klein und groß**, und sie ist eine Setzung mit
// einem Grund: ein Tischbild ist eine Nachricht, zweihundert Ereignisse sind
// zweihundert.
func TestJenseitsDesFenstersFaengtDieAeltesteSpaeterAn(t *testing.T) {
	t.Parallel()

	strom := make([]domain.Ereignis, 0, service.Nachreichfenster+50)

	for folge := 1; folge <= service.Nachreichfenster+50; folge++ {
		strom = append(strom, domain.Ereignis{
			PartieID: "p1", Folge: int64(folge), Kette: "glied",
		})
	}

	leitung, _ := mitStrom(t, strom)

	stand, fehler := leitung.Wirtsstand(t.Context(), "p1", "sitz-1")
	if fehler != nil {
		t.Fatalf("Wirtsstand: %v", fehler)
	}

	if stand.AelteteFolge != 51 {
		t.Errorf("die älteste nachgereichte ist %d, erwartet 51",
			stand.AelteteFolge)
	}
}

// **Nachgereicht wird gefiltert — was ein Sitz nicht sehen darf, kommt nicht.**
func TestNachreichenFiltertFuerDenSitz(t *testing.T) {
	t.Parallel()

	leitung, _ := mitStrom(t, []domain.Ereignis{
		{PartieID: "p1", Folge: 1, Kette: "eins"},
		{
			PartieID: "p1", Folge: 2, Kette: "zwei",
			Privat: []domain.Ergaenzung{
				{Empfaenger: "sitz-2", Inhalt: map[string]any{"karte": "k9"}},
			},
		},
	})

	strom, fehler := leitung.Nachreichen(t.Context(), "p1", "sitz-1", 2)
	if fehler != nil {
		t.Fatalf("Nachreichen: %v", fehler)
	}

	if len(strom) != 1 || strom[0].Folge != 2 {
		t.Fatalf("nachgereicht wurde %+v", strom)
	}

	if len(strom[0].Privat) != 0 {
		t.Errorf("sitz-1 bekam die Ergänzung für sitz-2: %+v", strom[0].Privat)
	}
}

// **Eine Partie, die es nicht gibt, gibt keinen Stand.**
//
// **Und zwar beide Wege** — derselbe Fund wie beim Verlauf, wo ein fehlender
// Bestandszugriff eine leere Liste statt eines Befundes gab (§8).
func TestOhnePartieGibtEsKeinenStandUndNichtsNachzureichen(t *testing.T) {
	t.Parallel()

	leitung, _ := eineLeitung(t)

	if _, fehler := leitung.Wirtsstand(
		t.Context(), "gibtsnicht", "sitz-1",
	); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Wirtsstand: es kam %v", fehler)
	}

	if _, fehler := leitung.Nachreichen(
		t.Context(), "gibtsnicht", "sitz-1", 1,
	); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Nachreichen: es kam %v", fehler)
	}
}

// **Was der Bestand klagt, kommt zurück — auf jedem der drei Wege.**
func TestWasDerBestandBeimStandKlagtKommtZurueck(t *testing.T) {
	t.Parallel()

	for weg, fahre := range map[string]func(*service.Partieleitung) error{
		"LetzteKette": func(l *service.Partieleitung) error {
			_, fehler := l.Wirtsstand(t.Context(), "p1", "sitz-1")

			return fehler
		},
		"ereignisse": func(l *service.Partieleitung) error {
			_, fehler := l.Nachreichen(t.Context(), "p1", "sitz-1", 1)

			return fehler
		},
	} {
		t.Run(weg, func(t *testing.T) {
			t.Parallel()

			kaputt := errors.New("der Bestand ist zu")

			leitung, speicher := mitStrom(t, []domain.Ereignis{
				{PartieID: "p1", Folge: 1, Kette: "eins"},
			})
			speicher.stolpert = map[string]error{weg: kaputt}

			if fehler := fahre(leitung); !errors.Is(fehler, kaputt) {
				t.Errorf("es kam %v", fehler)
			}
		})
	}
}

// **Ein Bestand, der eine Folge nennt und nicht hat, wird benannt.**
//
// **Er ist erreichbar**: `LetzteKette` und `Ereignisse` sind zwei Fragen an
// denselben Bestand, und zwischen ihnen kann etwas geschehen. Was dann fehlt,
// ist ein Widerspruch im eigenen Haus — und er wird benannt statt verschluckt.
func TestEineFolgeOhneEreignisWirdBenannt(t *testing.T) {
	t.Parallel()

	leitung, speicher := mitStrom(t, []domain.Ereignis{
		{PartieID: "p1", Folge: 5, Kette: "fuenf"},
	})

	// **`LetzteKette` sagt weiter fünf, die Ereignisse sind weg.**
	speicher.letzteOhneStrom = true

	if _, fehler := leitung.Wirtsstand(
		t.Context(), "p1", "sitz-1",
	); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Eine Ergänzung, die sich nicht abbilden lässt, wird benannt.**
//
// # Warum dieser Zweig auf der Wirtsseite steht und nicht beim Gast
//
// **Weil der Speicher hinter einer Schnittstelle liegt** (§2): was von dort
// kommt, ist nicht zwangsläufig durch `encoding/json` gegangen. Beim Gast
// kommt jedes Ereignis über eine Leitung und **war** JSON — dort kann dieser
// Fall nicht eintreten, und ein Zweig dafür wäre eine Zusage, die niemand
// hält (§10).
func TestEineUnabbildbareErgaenzungBeimStandWirdBenannt(t *testing.T) {
	t.Parallel()

	leitung, _ := mitStrom(t, []domain.Ereignis{{
		PartieID: "p1", Folge: 1, Kette: "eins",
		Privat: []domain.Ergaenzung{{
			Empfaenger: "sitz-1",
			Inhalt:     map[string]any{"kanal": make(chan int)},
		}},
	}})

	if _, fehler := leitung.Wirtsstand(
		t.Context(), "p1", "sitz-1"); fehler == nil {
		t.Error("eine unabbildbare Ergänzung ging durch")
	}
}

// **Was der Bestand beim Abdruck klagt, kommt zurück.**
//
// **Ein eigener Lauf**, weil `Wirtsstand` den Bestand **zweimal** fragt: erst
// nach der letzten Kette, dann nach dem Ereignis dazu. Ein Lauf, der nur die
// erste Frage scheitern ließe, ließe die zweite ungemessen (§10).
func TestWasDerBestandBeimAbdruckKlagtKommtZurueck(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("die Ereignisse sind weg")

	leitung, speicher := mitStrom(t, []domain.Ereignis{
		{PartieID: "p1", Folge: 1, Kette: "eins"},
	})
	speicher.stolpert = map[string]error{"ereignisse": kaputt}

	if _, fehler := leitung.Wirtsstand(
		t.Context(), "p1", "sitz-1"); !errors.Is(fehler, kaputt) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Eine negative Folge wird zu null.**
//
// # Warum dieser Zweig dasteht
//
// **Weil eine vorzeichenlose Umwandlung aus ihr eine sehr große Zahl machte**
// — und der Gast stünde für immer „weiter" als der Wirt, also für immer vor
// einem Abbruch. Erreichbar ist er über den [service.Standhalter]: was der
// merkt, merkt die **Anwendung**, und die ist eine Schnittstelle (§8).
func TestEineNegativeFolgeWirdZuNull(t *testing.T) {
	t.Parallel()

	gast, naht := einGast(&nehmerdoppelGast{})
	gast.MitStandhalter(&standdoppel{folge: -5, abdruck: "krumm"})

	if fehler := gast.Kehrezurueck(t.Context()); fehler != nil {
		t.Fatalf("Kehrezurueck: %v", fehler)
	}

	if wunsch := derWunsch(t, naht); wunsch.Folge != 0 {
		t.Errorf("der Wunsch nennt die Folge %d", wunsch.Folge)
	}
}
