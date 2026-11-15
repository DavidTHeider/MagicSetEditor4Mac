package service_test

// **Der Gast setzt sich, und die Partie schreibt es auf** — P9b
// (#2377).
// Das Deck kommt in den Sitz, seine Karten in die Bibliothek, die Fassung
// steigt um eins, und im Strom steht, dass sich jemand gesetzt hat — nicht,
// mit welchen Karten.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// eineEinladung ist eine Partie, deren zweiter Sitz auf Bo wartet.
func eineEinladung() domain.Partie {
	partie := einePartie()
	partie.Sitze[1].Deck = domain.Partiedeck{}
	partie.Sitze[1].Eingeladen = "peer-bo"

	return partie
}

// bosDeck ist, was Bo mitbringt — drei Karten aus seinem Bestand.
func bosDeck() domain.Partiedeck {
	return domain.Partiedeck{
		Quelle:  "peer-bo",
		Abdruck: "von Bo behauptet",
		Karten: []domain.Abzugskarte{{
			Kennung: "k9", Name: "Moorgeist", Menge: 3,
			Seiten: []domain.Abzugsseite{{Name: "Moorgeist", KarteID: 7}},
		}},
	}
}

func eineLeitungMitEinladung(t *testing.T) (*service.Partieleitung, *partieprobe) {
	t.Helper()

	leitung, speicher := eineLeitung(t)
	if _, fehler := leitung.Anlegen(context.Background(), eineEinladung()); fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	return leitung, speicher
}

func TestDerGastSetztSichUndDiePartieSchreibtEsAuf(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineLeitungMitEinladung(t)

	ausgang, fehler := leitung.Platznehmen(context.Background(), "p1", "sitz-2", "peer-bo", bosDeck(), nil)
	if fehler != nil {
		t.Fatalf("Platznehmen: %v", fehler)
	}

	if ausgang.Fassung != 2 || len(ausgang.Ereignisse) != 1 ||
		ausgang.Ereignisse[0].Art != service.EreignisPlatznahme ||
		ausgang.Ereignisse[0].Oeffentlich["karten"] != float64(3) {
		t.Fatalf("der Ausgang: %+v", ausgang)
	}

	partie := speicher.partien["p1"]
	sitz := partie.SitzVon("sitz-2")

	if sitz.Wartet() || sitz.Deck.Quelle != "peer-bo" {
		t.Errorf("der Sitz: %+v", sitz)
	}

	// **Der Abdruck wird nachgerechnet** — er geht in die Bildkennung ein,
	// und was der Gast behauptet, ist nur eine Behauptung.
	if sitz.Deck.Abdruck == "von Bo behauptet" || sitz.Deck.Abdruck == "" {
		t.Errorf("der Abdruck: %q", sitz.Deck.Abdruck)
	}

	gelegt := 0
	for _, objekt := range partie.Objekte {
		if objekt.Eigner == "sitz-2" && objekt.Zone == domain.ZoneBibliothek {
			gelegt++
		}
	}

	if gelegt != 3 {
		t.Errorf("%d Karten in Bos Bibliothek", gelegt)
	}
}

// **Ein zweites Mal ist dieselbe Antwort und kein zweiter Zug** — eine
// Leitung, die nach dem Hinsetzen abreißt, setzt sich beim Wiederkommen
// nicht noch einmal.
func TestEinZweitesHinsetzenIstKeinZweiterZug(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineLeitungMitEinladung(t)
	ctx := context.Background()

	erst, fehler := leitung.Platznehmen(ctx, "p1", "sitz-2", "peer-bo", bosDeck(), nil)
	if fehler != nil {
		t.Fatalf("Platznehmen: %v", fehler)
	}

	noch, fehler := leitung.Platznehmen(ctx, "p1", "sitz-2", "peer-bo", bosDeck(), nil)
	if fehler != nil || noch.Fassung != erst.Fassung || speicher.partien["p1"].Fassung != 2 {
		t.Errorf("zweimal: %+v, %v, Fassung %d", noch, fehler, speicher.partien["p1"].Fassung)
	}
}

func TestWerSichNichtSetzenDarfAendertNichts(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		vorher func(*service.Partieleitung) error
		peer   string
		art    error
	}{
		"ein anderer Peer": {nil, "peer-cora", domain.ErrConflict},
		"eine angehaltene Partie": {func(l *service.Partieleitung) error {
			return l.Pausieren(context.Background(), "p1")
		}, "peer-bo", domain.ErrConflict},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			leitung, speicher := eineLeitungMitEinladung(t)
			if fall.vorher != nil {
				if fehler := fall.vorher(leitung); fehler != nil {
					t.Fatalf("vorher: %v", fehler)
				}
			}

			fassung := speicher.partien["p1"].Fassung

			if _, fehler := leitung.Platznehmen(context.Background(), "p1", "sitz-2", fall.peer, bosDeck(), nil); !errors.Is(fehler, fall.art) {
				t.Fatalf("es kam %v", fehler)
			}

			if speicher.partien["p1"].Fassung != fassung || !speicher.partien["p1"].SitzVon("sitz-2").Wartet() {
				t.Error("trotz Abweisung hat sich etwas geändert")
			}
		})
	}
}

func TestOhnePartieKeinPlatz(t *testing.T) {
	t.Parallel()

	leitung, _ := eineLeitung(t)

	if _, fehler := leitung.Platznehmen(context.Background(), "weg", "sitz-2", "peer-bo", bosDeck(), nil); !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("es kam %v", fehler)
	}
}

// **Klagt der Speicher bei der Frage nach dem ersten Mal, setzt sich
// niemand** — sonst säße ein Gast zweimal.
func TestKlagtDerSpeicherSetztSichNiemand(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineLeitungMitEinladung(t)
	kaputt := errors.New("der Speicher klemmt")
	speicher.stolpert = map[string]error{"Befehlsergebnis": kaputt}

	if _, fehler := leitung.Platznehmen(context.Background(), "p1", "sitz-2", "peer-bo", bosDeck(), nil); !errors.Is(fehler, kaputt) {
		t.Fatalf("es kam %v", fehler)
	}

	if !speicher.partien["p1"].SitzVon("sitz-2").Wartet() {
		t.Error("der Gast sitzt trotzdem")
	}
}

// **Ein übergroßer Abzug wird abgewiesen, bevor ein Objekt entsteht** —
// Zweitblick über Phase 7, Befund 2: hier entstand die Bibliothek vor der
// Prüfung, und eine Milliarde Exemplare kostete den Speicher, bevor jemand
// nein sagte.
func TestEinUebergrosserAbzugLegtNichtsAn(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineLeitungMitEinladung(t)
	riesig := bosDeck()
	riesig.Karten[0].Menge = 1 << 40

	fertig := make(chan error, 1)
	go func() {
		_, fehler := leitung.Platznehmen(context.Background(), "p1", "sitz-2", "peer-bo", riesig, nil)
		fertig <- fehler
	}()

	select {
	case fehler := <-fertig:
		if !errors.Is(fehler, domain.ErrValidation) {
			t.Fatalf("es kam %v", fehler)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("die Leitung legt noch Objekte an")
	}

	if !speicher.partien["p1"].SitzVon("sitz-2").Wartet() {
		t.Error("trotzdem gesetzt")
	}
}
