package service_test

// **Wie ein Gast zurückkommt** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// **Dass der Gast sich merkt, wo er steht — und zwar sichtbezogen.** Der
// Abdruck braucht den Sitz, und den Sitz kennt er erst aus dem **Tischbild**:
// vergeben hat ihn der Wirt. Ein Gast, der ihn selbst wählte, wählte ihn
// falsch — und käme nie wieder herein.

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// standdoppel merkt sich, wo ein Gast steht.
type standdoppel struct {
	folge   int64
	abdruck string
}

func (s *standdoppel) Merke(folge int64, abdruck string) {
	s.folge, s.abdruck = folge, abdruck
}

func (s *standdoppel) Stand() (folge int64, abdruck string) {
	return s.folge, s.abdruck
}

// derWunsch liest den Rückkehrwunsch aus der letzten Nachricht.
func derWunsch(
	t *testing.T, naht *nahtdoppel,
) domain.Wiederaufnahmewunsch {
	t.Helper()

	nachricht := naht.letzte(t)

	var umschlag domain.Kontrolle
	if fehler := json.Unmarshal(nachricht.Inhalt, &umschlag); fehler != nil {
		t.Fatalf("keine Kontrolle: %v", fehler)
	}

	if umschlag.Art != domain.KontrolleRueckkehr {
		t.Fatalf("hinaus ging %q", umschlag.Art)
	}

	var wunsch domain.Wiederaufnahmewunsch
	if fehler := json.Unmarshal(umschlag.Inhalt, &wunsch); fehler != nil {
		t.Fatalf("kein Wunsch: %v", fehler)
	}

	return wunsch
}

// eineRueckkehr verpackt, was der Wirt antwortet.
func eineRueckkehr(
	t *testing.T, antwort domain.Rueckkehr,
) domain.Partienachricht {
	t.Helper()

	return eineWirtsnachricht(t, domain.ArtKontrolle, mussVerpacken(
		t, domain.KontrolleAbzug, antwort))
}

// mussVerpacken verpackt eine Kontrolle oder bricht ab.
func mussVerpacken(
	t *testing.T, art domain.Kontrollart, nutzlast any,
) json.RawMessage {
	t.Helper()

	roh, fehler := domain.VerpackeKontrolle(art, nutzlast)
	if fehler != nil {
		t.Fatalf("verpacken: %v", fehler)
	}

	return roh
}

// **Ein Gast, der noch nie dabei war, schickt Folge null.**
func TestEinNeuerGastSchicktFolgeNull(t *testing.T) {
	t.Parallel()

	gast, naht := einGast(&nehmerdoppelGast{})

	if fehler := gast.Kehrezurueck(t.Context()); fehler != nil {
		t.Fatalf("Kehrezurueck: %v", fehler)
	}

	wunsch := derWunsch(t, naht)

	if wunsch.Folge != 0 || wunsch.Kette != "" {
		t.Errorf("der Wunsch ist %+v", wunsch)
	}

	if wunsch.Partie != "p1" || wunsch.Epoche != domain.Partieepoche {
		t.Errorf("der Wunsch nennt %q in der Epoche %d",
			wunsch.Partie, wunsch.Epoche)
	}
}

// **Ein Gast merkt sich jedes Ereignis — sichtbezogen.**
//
// # Warum der Sitz aus dem Tischbild kommt
//
// **Weil der Gast ihn sonst nicht kennt.** Beim Beitreten schreibt er keinen;
// vergeben wird er vom Wirt, und er steht im Bild (`Sicht.Fuer`). Ohne ihn
// rechnete der Gast den Abdruck eines Zuschauers und käme nie wieder herein.
func TestEinGastMerktSichSeinenSichtbezogenenStand(t *testing.T) {
	t.Parallel()

	halter := &standdoppel{}
	gast, _ := einGast(&nehmerdoppelGast{})
	gast.MitStandhalter(halter)

	// **Erst das Bild** — es bringt den Sitz mit.
	bild := domain.Tischbild{
		Sicht:   domain.Partiesicht{PartieID: "p1", Fuer: "bert"},
		Erlaubt: []string{},
	}

	if fehler := gast.Angekommen(t.Context(), eineWirtsnachricht(
		t, domain.ArtKontrolle,
		mussVerpacken(t, domain.KontrolleTischbild, bild),
	)); fehler != nil {
		t.Fatalf("Tischbild: %v", fehler)
	}

	ereignis := domain.Ereignis{
		PartieID: "p1", Folge: 4, Art: "karte.ziehen", Kette: "vier",
		Privat: []domain.Ergaenzung{
			{Empfaenger: "bert", Inhalt: map[string]any{"karte": "k1"}},
		},
	}

	if fehler := gast.Angekommen(
		t.Context(), eineTatsache(t, ereignis)); fehler != nil {
		t.Fatalf("Ereignis: %v", fehler)
	}

	erwartet, fehler := ereignis.Sichtabdruck("bert")
	if fehler != nil {
		t.Fatalf("Sichtabdruck: %v", fehler)
	}

	if halter.folge != 4 || halter.abdruck != erwartet {
		t.Errorf("gemerkt ist %d/%q, erwartet 4/%q",
			halter.folge, halter.abdruck, erwartet)
	}

	// **Und beim nächsten Mal steht genau das im Wunsch.**
	gast, naht := einGast(&nehmerdoppelGast{})
	gast.MitStandhalter(halter)

	if fehler := gast.Kehrezurueck(t.Context()); fehler != nil {
		t.Fatalf("Kehrezurueck: %v", fehler)
	}

	if wunsch := derWunsch(t, naht); wunsch.Folge != 4 ||
		wunsch.Kette != erwartet {
		t.Errorf("der Wunsch ist %+v", wunsch)
	}
}

// **Ein Abzug setzt den Stand — samt Abdruck — und zeigt sein Bild.**
func TestEinAbzugSetztDenStandUndZeigtSeinBild(t *testing.T) {
	t.Parallel()

	halter := &standdoppel{folge: 3, abdruck: "drei"}
	nehmer := &nehmerdoppelGast{}

	gast, _ := einGast(nehmer)
	gast.MitStandhalter(halter)

	if fehler := gast.Angekommen(t.Context(), eineRueckkehr(t, domain.Rueckkehr{
		Art: domain.WegAbzug, Grund: "zu weit zurück",
		Stand: 900, Abdruck: "neunhundert",
		Bild: domain.Tischbild{
			Sicht: domain.Partiesicht{PartieID: "p1", Fuer: "bert"},
		},
	})); fehler != nil {
		t.Fatalf("Abzug: %v", fehler)
	}

	if halter.folge != 900 || halter.abdruck != "neunhundert" {
		t.Errorf("gemerkt ist %d/%q", halter.folge, halter.abdruck)
	}

	if len(nehmer.bilder) != 1 {
		t.Errorf("%d Bilder kamen an, erwartet eines", len(nehmer.bilder))
	}
}

// **Ein Abbruch legt die Leitung nieder.**
func TestEinAbbruchLegtDieLeitungNieder(t *testing.T) {
	t.Parallel()

	gast, _ := einGast(&nehmerdoppelGast{})

	if fehler := gast.Angekommen(t.Context(), eineRueckkehr(t, domain.Rueckkehr{
		Art: domain.WegAbbruch, Grund: "das ist nicht dieselbe Partie",
	})); !errors.Is(fehler, domain.ErrConflict) || !errors.Is(fehler, service.ErrWirtLehntAb) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Die Absage trägt den Satz des Wirts allein** (P9c, #2377) — das Fenster
// des Gastes setzt ihn unter seinen eigenen, ohne die Kette davor.
func TestDieAbsageTraegtDenSatzDesWirts(t *testing.T) {
	t.Parallel()

	gast, _ := einGast(&nehmerdoppelGast{})

	fehler := gast.Angekommen(t.Context(), eineRueckkehr(t, domain.Rueckkehr{
		Art: domain.WegAbbruch, Grund: "das ist nicht dieselbe Partie",
	}))

	var absage service.Absage
	if !errors.As(fehler, &absage) || absage.Grund != "das ist nicht dieselbe Partie" {
		t.Errorf("es kam %v", fehler)
	}

	// **Fürs Protokoll trägt sie die ganze Kette.**
	if !strings.HasSuffix(fehler.Error(), ": das ist nicht dieselbe Partie") {
		t.Errorf("der Text: %q", fehler.Error())
	}
}

// **Die anderen beiden Wege tun nichts — und das ist richtig.**
//
// **Bei `weiter` fehlt nichts**, und bei `Ereignisse` kommt das Fehlende
// gleich hinterher und trägt seinen Stand selbst mit. Nur der Abzug springt.
func TestWeiterUndEreignisseTunBeimGastNichts(t *testing.T) {
	t.Parallel()

	for name, art := range map[string]domain.Wiederaufnahmeart{
		"weiter":     domain.WegWeiter,
		"Ereignisse": domain.WegEreignisse,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			halter := &standdoppel{folge: 3, abdruck: "drei"}
			nehmer := &nehmerdoppelGast{}

			gast, _ := einGast(nehmer)
			gast.MitStandhalter(halter)

			if fehler := gast.Angekommen(
				t.Context(), eineRueckkehr(t, domain.Rueckkehr{
					Art: art, Stand: 900, Abdruck: "neunhundert",
				})); fehler != nil {
				t.Fatalf("Rückkehr: %v", fehler)
			}

			if halter.folge != 3 || halter.abdruck != "drei" {
				t.Errorf("der Stand sprang auf %d/%q",
					halter.folge, halter.abdruck)
			}

			if len(nehmer.bilder) != 0 {
				t.Error("es kam ein Bild, wo keines gehört")
			}
		})
	}
}

// **Ohne Standhalter geht es auch — und es fängt jedes Mal von vorn an.**
//
// **Die Gegenprobe** (§10): ein Gast an der Kommandozeile hat niemanden, der
// sich etwas merkt, und er soll trotzdem spielen können.
func TestOhneStandhalterFaengtJedeRueckkehrVonVornAn(t *testing.T) {
	t.Parallel()

	gast, naht := einGast(&nehmerdoppelGast{})

	if fehler := gast.Angekommen(t.Context(), eineTatsache(t, domain.Ereignis{
		PartieID: "p1", Folge: 7, Kette: "sieben",
	})); fehler != nil {
		t.Fatalf("Ereignis: %v", fehler)
	}

	if fehler := gast.Kehrezurueck(t.Context()); fehler != nil {
		t.Fatalf("Kehrezurueck: %v", fehler)
	}

	if wunsch := derWunsch(t, naht); wunsch.Folge != 0 {
		t.Errorf("der Wunsch nennt die Folge %d", wunsch.Folge)
	}
}

// **Was keine Rückkehr ist, hält an.**
func TestEineKrummeRueckkehrHaeltDenGastAn(t *testing.T) {
	t.Parallel()

	gast, _ := einGast(&nehmerdoppelGast{})

	nachricht := eineWirtsnachricht(t, domain.ArtKontrolle,
		mussVerpacken(t, domain.KontrolleAbzug, "kein Verbund"))

	if fehler := gast.Angekommen(
		t.Context(), nachricht,
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}
}
