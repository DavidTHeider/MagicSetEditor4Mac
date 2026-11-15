package service_test

// **Wie der Wirt einen Gast wieder hereinholt** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// **Gegen die vier Ausgänge, und zwar an der Naht.** Dass `domain.Wiederaufnahme`
// richtig rechnet, hält `internal/domain`; hier steht, dass der Wirt die
// Rechnung **anwendet** — dass er den Stand aus seinem Bestand holt, die
// Antwort schickt, das Fehlende nachreicht und bei einem Abbruch die Leitung
// niederlegt statt weiterzuspielen.

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einRueckkehrwunsch verpackt einen Wunsch, wie ein Gast ihn schickt.
func einRueckkehrwunsch(
	t *testing.T, folge uint64, kette string,
) domain.Partienachricht {
	t.Helper()

	return eineKontrollnachricht(t, domain.KontrolleRueckkehr,
		domain.Wiederaufnahmewunsch{
			Partie: "p1", Epoche: domain.Partieepoche,
			Folge: folge, Kette: kette,
		})
}

// dieAntwort liest die Rückkehr aus der zuletzt gesendeten Nachricht.
func dieAntwort(t *testing.T, naht *nahtdoppel) domain.Rueckkehr {
	t.Helper()

	naht.schloss.Lock()
	defer naht.schloss.Unlock()

	for stelle := len(naht.gesendet) - 1; stelle >= 0; stelle-- {
		var nachricht domain.Partienachricht
		if fehler := json.Unmarshal(
			naht.gesendet[stelle], &nachricht); fehler != nil {
			t.Fatalf("keine Nachricht: %v", fehler)
		}

		if nachricht.Art != domain.ArtKontrolle {
			continue
		}

		var umschlag domain.Kontrolle
		if fehler := json.Unmarshal(nachricht.Inhalt, &umschlag); fehler != nil {
			t.Fatalf("keine Kontrolle: %v", fehler)
		}

		if umschlag.Art != domain.KontrolleAbzug {
			continue
		}

		var antwort domain.Rueckkehr
		if fehler := json.Unmarshal(umschlag.Inhalt, &antwort); fehler != nil {
			t.Fatalf("keine Rückkehr: %v", fehler)
		}

		return antwort
	}

	t.Fatal("es ging kein Abzug hinaus")

	return domain.Rueckkehr{}
}

// **Ein Gast auf Stand bekommt sein Bild und sonst nichts.**
func TestEinGastAufStandBekommtNurSeinBild(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppel{stand: domain.Wirtsstand{
		Epoche: domain.Partieepoche, Folge: 5, Kette: "abc", AelteteFolge: 1,
	}}

	wirt, naht := einWirt(nehmer)

	if fehler := wirt.Angekommen(
		t.Context(), einRueckkehrwunsch(t, 5, "abc")); fehler != nil {
		t.Fatalf("angekommen: %v", fehler)
	}

	antwort := dieAntwort(t, naht)

	if antwort.Art != domain.WegWeiter {
		t.Errorf("der Weg ist %q statt weiter: %s", antwort.Art, antwort.Grund)
	}

	if len(nehmer.nachgereicht) != 0 {
		t.Errorf("nachgereicht wurde ab %v — es fehlte nichts",
			nehmer.nachgereicht)
	}
}

// **Eine kleine Lücke wird nachgereicht — ab der richtigen Folge.**
func TestEineKleineLueckeWirdNachgereicht(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppel{
		stand: domain.Wirtsstand{
			Epoche: domain.Partieepoche, Folge: 9, Kette: "neun",
			AelteteFolge: 1,
		},
		nachholt: []domain.Ereignis{
			{Folge: 8, Art: "karte.ziehen"}, {Folge: 9, Art: "objekt.tappen"},
		},
	}

	wirt, naht := einWirt(nehmer)

	if fehler := wirt.Angekommen(
		t.Context(), einRueckkehrwunsch(t, 7, "sieben")); fehler != nil {
		t.Fatalf("angekommen: %v", fehler)
	}

	if antwort := dieAntwort(t, naht); antwort.Art != domain.WegEreignisse {
		t.Errorf("der Weg ist %q statt Ereignisse: %s",
			antwort.Art, antwort.Grund)
	}

	if len(nehmer.nachgereicht) != 1 || nehmer.nachgereicht[0] != 8 {
		t.Errorf("nachgereicht wurde ab %v, erwartet ab 8",
			nehmer.nachgereicht)
	}

	// **Und die Ereignisse gehen wirklich hinaus** — zwei davon, dazu die
	// Antwort und das Bild.
	if gezaehlt := ereignisseGezaehlt(t, naht); gezaehlt != 2 {
		t.Errorf("%d Ereignisse gingen hinaus, erwartet zwei", gezaehlt)
	}
}

// **Eine große Lücke bekommt den Abzug — mit Bild, Stand und Abdruck.**
func TestEineGrosseLueckeBekommtDenAbzug(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppel{stand: domain.Wirtsstand{
		Epoche: domain.Partieepoche, Folge: 900, Kette: "neunhundert",
		AelteteFolge: 701,
	}}

	wirt, naht := einWirt(nehmer)

	if fehler := wirt.Angekommen(
		t.Context(), einRueckkehrwunsch(t, 12, "zwoelf")); fehler != nil {
		t.Fatalf("angekommen: %v", fehler)
	}

	antwort := dieAntwort(t, naht)

	switch {
	case antwort.Art != domain.WegAbzug:
		t.Errorf("der Weg ist %q statt Abzug: %s", antwort.Art, antwort.Grund)
	case antwort.Stand != 900:
		t.Errorf("der Stand ist %d statt 900", antwort.Stand)

	// **Der Abdruck ist der Punkt.** Ohne ihn käme der Gast genau einmal
	// zurück: beim nächsten Mal schickte er eine leere Kette.
	case antwort.Abdruck != "neunhundert":
		t.Errorf("der Abdruck ist %q", antwort.Abdruck)
	case antwort.Bild.Erlaubt == nil:
		t.Error("der Abzug trägt kein Bild")
	}

	if len(nehmer.nachgereicht) != 0 {
		t.Error("beim Abzug wurde trotzdem nachgereicht")
	}
}

// **Ein Stand, der sich nicht erklären lässt, legt die Leitung nieder.**
//
// **Und die Antwort geht vorher hinaus** (§8): ein Gast, dem die Leitung ohne
// ein Wort zugeht, sucht den Fehler im Netz.
func TestEinUnerklaerlicherStandLegtDieLeitungNieder(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppel{stand: domain.Wirtsstand{
		Epoche: domain.Partieepoche, Folge: 3, Kette: "drei", AelteteFolge: 1,
	}}

	wirt, naht := einWirt(nehmer)

	// **Der Gast steht weiter als der Wirt** — er hat mit jemand anderem
	// gespielt.
	fehler := wirt.Angekommen(t.Context(), einRueckkehrwunsch(t, 9, "neun"))

	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("es kam %v", fehler)
	}

	if antwort := dieAntwort(t, naht); antwort.Art != domain.WegAbbruch {
		t.Errorf("hinaus ging %q statt eines Abbruchs", antwort.Art)
	}
}

// **Was der Bestand klagt, kommt zurück.**
func TestWasDerBestandBeiDerRueckkehrKlagtKommtZurueck(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("der Bestand ist zu")
	wirt, _ := einWirt(&nehmerdoppel{klagt: kaputt})

	if fehler := wirt.Angekommen(
		t.Context(), einRueckkehrwunsch(t, 1, "eins"),
	); !errors.Is(fehler, kaputt) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Was kein Wunsch ist, hält an.**
func TestEinKrummerRueckkehrwunschHaeltAn(t *testing.T) {
	t.Parallel()

	wirt, _ := einWirt(&nehmerdoppel{})

	nachricht := eineKontrollnachricht(t, domain.KontrolleRueckkehr, "kein Verbund")

	if fehler := wirt.Angekommen(
		t.Context(), nachricht,
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}
}

// ereignisseGezaehlt zählt, wie viele Ereignisse hinausgingen.
func ereignisseGezaehlt(t *testing.T, naht *nahtdoppel) int {
	t.Helper()

	naht.schloss.Lock()
	defer naht.schloss.Unlock()

	gezaehlt := 0

	for _, roh := range naht.gesendet {
		var nachricht domain.Partienachricht
		if fehler := json.Unmarshal(roh, &nachricht); fehler != nil {
			t.Fatalf("keine Nachricht: %v", fehler)
		}

		if nachricht.Art == domain.ArtEreignis {
			gezaehlt++
		}
	}

	return gezaehlt
}

var _ = service.Nachreichfenster

// **Was auf dem Weg hinaus klagt, kommt zurück — auf jedem der drei Wege.**
//
// **Drei und nicht einer**: die Antwort, das Nachreichen und das Bild danach.
// Ein Lauf, der nur den ersten scheitern ließe, ließe die zwei dahinter
// ungemessen (§10).
func TestWasBeiDerRueckkehrHinausKlagtKommtZurueck(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("die Leitung ist zu")

	for name, fall := range map[string]struct {
		nehmer *nehmerdoppel
		klagt  func(*nehmerdoppel, *nahtdoppel)
	}{
		"die Antwort": {
			nehmer: &nehmerdoppel{stand: domain.Wirtsstand{
				Epoche: domain.Partieepoche, AelteteFolge: 1,
			}},
			klagt: func(_ *nehmerdoppel, naht *nahtdoppel) {
				naht.klage = kaputt
			},
		},
		"das Bild beim Abzug": {
			nehmer: &nehmerdoppel{
				stand: domain.Wirtsstand{
					Epoche: domain.Partieepoche, Folge: 900,
					AelteteFolge: 701,
				},
				bildklage:     kaputt,
				bildklageNach: 1,
			},
			klagt: func(*nehmerdoppel, *nahtdoppel) {},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wirt, naht := einWirt(fall.nehmer)
			fall.klagt(fall.nehmer, naht)

			if fehler := wirt.Angekommen(
				t.Context(), einRueckkehrwunsch(t, 12, "zwoelf"),
			); !errors.Is(fehler, kaputt) {
				t.Errorf("es kam %v", fehler)
			}
		})
	}
}

// **Was das Nachreichen klagt, kommt zurück.**
func TestWasDasNachreichenKlagtKommtZurueck(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("der Strom ist weg")

	nehmer := &nehmerdoppel{
		stand: domain.Wirtsstand{
			Epoche: domain.Partieepoche, Folge: 9, Kette: "neun",
			AelteteFolge: 1,
		},
		nachreichklage: kaputt,
	}

	wirt, _ := einWirt(nehmer)

	if fehler := wirt.Angekommen(
		t.Context(), einRueckkehrwunsch(t, 7, "sieben"),
	); !errors.Is(fehler, kaputt) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Ein nachgereichtes Ereignis, das sich nicht verpacken lässt, hält an.**
//
// **Es ist erreichbar**: was nachgereicht wird, kommt aus einem Speicher
// hinter einer Schnittstelle, und `Ereignis.Oeffentlich` ist eine offene
// Abbildung — was ein Regelausleger hineinlegt, entscheidet er (§8).
func TestEinUnverpackbaresNachgereichtesEreignisHaeltAn(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppel{
		stand: domain.Wirtsstand{
			Epoche: domain.Partieepoche, Folge: 9, Kette: "neun",
			AelteteFolge: 1,
		},
		nachholt: []domain.Ereignis{{
			Folge: 8, Oeffentlich: map[string]any{"kanal": make(chan int)},
		}},
	}

	wirt, _ := einWirt(nehmer)

	if fehler := wirt.Angekommen(
		t.Context(), einRueckkehrwunsch(t, 7, "sieben")); fehler == nil {
		t.Error("ein unverpackbares Ereignis ging hinaus")
	}
}
