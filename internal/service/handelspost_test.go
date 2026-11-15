package service_test

// **Was mit einer angekommenen Handelsnachricht geschieht** — Werkstück `T5`
// von Phase 5.6
// (#1470).
//
// # Was hier gemessen wird
//
// **Dass der Absender nicht geglaubt wird.** Wer spricht, sagt die Leitung
// (N2); was im Umschlag steht, ist eine Behauptung. Gehen beide auseinander,
// endet es vor der Naht.
//
// **Dass ein Gegenüber ohne den Zweck `handel` nichts bewirkt** — auch nicht
// mit einer tadellosen Nachricht.
//
// **Und dass zweimal dieselbe Nachricht einmal wirkt und dieselbe Antwort
// bekommt.**

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

const (
	dasGegenueber  = "bgegenueber"
	derPostvorgang = "v-1"
)

// briefkastenprobe führt den Briefkasten im Arbeitsspeicher.
type briefkastenprobe struct {
	// wirkung merkt sich, was der Dienst an den Zusagen bewirken wollte
	// (Befund 3 des Nachlaufs zum Zweitblick über Phase 5.6).
	wirkung domain.Zusagenwirkungsart

	// vorher ist der Stand, gegen den geschrieben werden sollte (Befund Z1
	// des Zweitblicks über Phase 6.8).
	vorher domain.Handelsangebot

	eintraege map[string]domain.Inboxeintrag
	angebote  []domain.Handelsangebot
	klagt     error
}

func (b *briefkastenprobe) Nachschlagen(
	_ context.Context, schluessel string,
) (*domain.Inboxeintrag, error) {
	if b.klagt != nil {
		return nil, b.klagt
	}

	eintrag, dabei := b.eintraege[schluessel]
	if !dabei {
		return nil, nil //nolint:nilnil // „nichts" ist hier kein Fehler
	}

	return &eintrag, nil
}

func (b *briefkastenprobe) Zustellen(
	_ context.Context, eintrag domain.Inboxeintrag,
	vorher, angebot domain.Handelsangebot, wirkung domain.Zusagenwirkungsart,
) error {
	b.wirkung = wirkung
	b.vorher = vorher

	if b.klagt != nil {
		return b.klagt
	}

	if b.eintraege == nil {
		b.eintraege = map[string]domain.Inboxeintrag{}
	}

	if _, dabei := b.eintraege[eintrag.Schluessel]; dabei {
		return domain.ErrConflict
	}

	b.eintraege[eintrag.Schluessel] = eintrag
	b.angebote = append(b.angebote, angebot)

	return nil
}

// standprobe gibt den heutigen Stand eines Vorgangs.
type standprobe struct {
	stand domain.Handelsangebot
	klagt error
}

func (s standprobe) Stand(
	_ context.Context, _ string,
) (domain.Handelsangebot, error) {
	if s.klagt != nil {
		return domain.Handelsangebot{}, s.klagt
	}

	return s.stand, nil
}

// derHeutigeStand ist die Revision, gegen die vorgeschlagen wird.
func derHeutigeStand() domain.Handelsangebot {
	return domain.Handelsangebot{
		Vertragsfassung: 1, Vorgang: derPostvorgang, Revision: 1,
		Art: domain.HandelTausch, Initiator: "bich",
		Parteien: []string{"bich", dasGegenueber},
		Zustand:  domain.HandelAngeboten,
		Positionen: []domain.Handelsposition{{
			Von: "bich", Printing: "a1b2", Menge: 1,
			Guete: "nearMint", Sprache: "de", Finish: "nonfoil",
		}},
	}
}

// einNachfolger baut die nächste Revision.
//
// **Mit einer anderen Position als der Stand** — sonst wäre der Vorschlag
// derselbe, und ein Vorschlag, der nichts ändert, ist keiner.
func einNachfolger() domain.Handelsangebot {
	return derHeutigeStand().Fortgeschrieben([]domain.Handelsposition{{
		Von: "bich", Printing: "a1b2", Menge: 2,
		Guete: "nearMint", Sprache: "de", Finish: "nonfoil",
	}})
}

// einePost baut, was ankommt.
func einePost(
	angebot domain.Handelsangebot, anders func(*domain.Handelsumschlag),
) service.Handelsnachricht {
	umschlag := domain.Handelsumschlag{
		Fassung: domain.Handelsprotokollfassung,
		Art:     domain.NachrichtVorschlag, NachrichtID: "n-1",
		Vorgang: derPostvorgang, Basisrevision: 1, Abdruck: angebot.Abdruck(),
		Folge: 1, Absender: dasGegenueber, Unterschrift: "sig",
	}

	if anders != nil {
		anders(&umschlag)
	}

	return service.Handelsnachricht{Umschlag: umschlag, Angebot: angebot}
}

// einePoststelle stellt die Post über die Proben.
func einePoststelle(
	briefkasten *briefkastenprobe, kontakt domain.Peerkontakt,
) *service.Handelspost {
	return service.NeueHandelspost(briefkasten,
		standprobe{stand: derHeutigeStand()},
		kontaktnachschlagprobe{kontakt: kontakt})
}

// einHandelnder ist ein bestätigtes Gegenüber, das handeln darf.
func einHandelnder() domain.Peerkontakt {
	return domain.Peerkontakt{
		Kennung: dasGegenueber, Stand: domain.KontaktBestaetigt,
		Rechte: []domain.Zweck{domain.ZweckHandel},
	}
}

func TestEineNachrichtWirktEinmal(t *testing.T) {
	t.Parallel()

	briefkasten := &briefkastenprobe{}
	post := einePoststelle(briefkasten, einHandelnder())
	nachricht := einePost(einNachfolger(), nil)

	erste, fehler := post.Empfangen(context.Background(), dasGegenueber, nachricht)
	if fehler != nil {
		t.Fatalf("Empfangen: %v", fehler)
	}

	if erste.Art != domain.ZustellungNeu {
		t.Errorf("die erste gilt als %q", erste.Art)
	}

	// **Dieselbe noch einmal** — dieselbe Antwort, keine zweite Wirkung.
	zweite, fehler := post.Empfangen(context.Background(), dasGegenueber, nachricht)
	if fehler != nil {
		t.Fatalf("das zweite Empfangen: %v", fehler)
	}

	if zweite.Art != domain.ZustellungBekannt {
		t.Errorf("die zweite gilt als %q", zweite.Art)
	}

	if zweite.Ergebnis != erste.Ergebnis {
		t.Errorf("die Antworten sind %q und %q", erste.Ergebnis, zweite.Ergebnis)
	}

	if len(briefkasten.angebote) != 1 {
		t.Errorf("es wurden %d Wirkungen geschrieben", len(briefkasten.angebote))
	}
}

// **Dieselbe Operation mit anderem Inhalt ist ein Konflikt.**
func TestDieselbeOperationMitAnderemInhaltIstEinKonflikt(t *testing.T) {
	t.Parallel()

	briefkasten := &briefkastenprobe{}
	post := einePoststelle(briefkasten, einHandelnder())

	if _, fehler := post.Empfangen(context.Background(), dasGegenueber,
		einePost(einNachfolger(), nil)); fehler != nil {
		t.Fatalf("Empfangen: %v", fehler)
	}

	anderer := einNachfolger()
	anderer.BewertungGeteilt = true

	zweite, fehler := post.Empfangen(context.Background(), dasGegenueber,
		einePost(anderer, func(u *domain.Handelsumschlag) {
			u.NachrichtID, u.Abdruck = "n-2", anderer.Abdruck()
		}))
	if fehler != nil {
		t.Fatalf("das zweite Empfangen: %v", fehler)
	}

	if zweite.Art != domain.ZustellungKonflikt {
		t.Errorf("die zweite gilt als %q", zweite.Art)
	}

	if len(briefkasten.angebote) != 1 {
		t.Errorf("es wurden %d Wirkungen geschrieben", len(briefkasten.angebote))
	}
}

func TestWasVorDerNahtEndet(t *testing.T) {
	t.Parallel()

	for name, fall := range map[string]struct {
		sender    string
		kontakt   domain.Peerkontakt
		nachricht service.Handelsnachricht
	}{
		// **Der Absender wird nicht geglaubt** — wer spricht, sagt die
		// Leitung.
		"ein fremder Absender": {
			"canderer", einHandelnder(), einePost(einNachfolger(), nil),
		},

		// **Ohne den Zweck `handel` bewirkt nichts etwas** — auch keine
		// tadellose Nachricht.
		"ohne den Zweck": {
			dasGegenueber,
			domain.Peerkontakt{
				Kennung: dasGegenueber, Stand: domain.KontaktBestaetigt,
			},
			einePost(einNachfolger(), nil),
		},

		"ein gesperrtes Gegenüber": {
			dasGegenueber,
			domain.Peerkontakt{
				Kennung: dasGegenueber, Stand: domain.KontaktGesperrt,
				Rechte: []domain.Zweck{domain.ZweckHandel},
			},
			einePost(einNachfolger(), nil),
		},

		// **Ein krummer Umschlag** — eine fremde Fassung.
		"eine fremde Fassung": {
			dasGegenueber, einHandelnder(),
			einePost(einNachfolger(), func(u *domain.Handelsumschlag) {
				u.Fassung = 99
			}),
		},

		// **Ein Umschlag, der einen anderen Vorgang nennt als sein Angebot.**
		"zwei Vorgänge in einer Nachricht": {
			dasGegenueber, einHandelnder(),
			einePost(einNachfolger(), func(u *domain.Handelsumschlag) {
				u.Vorgang = "v-9"
			}),
		},

		// **Und einer, dessen Abdruck nicht zum Angebot passt** — sonst
		// unterschriebe jemand einen Abdruck und schickte ein anderes
		// Angebot.
		"ein Abdruck, der nicht passt": {
			dasGegenueber, einHandelnder(),
			einePost(einNachfolger(), func(u *domain.Handelsumschlag) {
				u.Abdruck = "ein anderer"
			}),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			briefkasten := &briefkastenprobe{}

			if _, fehler := einePoststelle(briefkasten, fall.kontakt).
				Empfangen(context.Background(), fall.sender,
					fall.nachricht); fehler == nil {
				t.Error("die Nachricht kam durch")
			}

			if len(briefkasten.angebote) != 0 {
				t.Errorf("es wurde trotzdem geschrieben: %+v", briefkasten.angebote)
			}
		})
	}
}

// **Ein Vorschlag gegen eine überholte Revision wird abgewiesen** — und zwar
// mit [domain.ErrConflict], damit der Klient noch einmal liest statt in einer
// Schleife zu wiederholen (T1).
func TestEinUeberholterVorschlagWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	ueberholt := einNachfolger()
	ueberholt.Eltern = "ein anderer Abdruck"

	_, fehler := einePoststelle(&briefkastenprobe{}, einHandelnder()).
		Empfangen(context.Background(), dasGegenueber,
			einePost(ueberholt, func(u *domain.Handelsumschlag) {
				u.Abdruck = ueberholt.Abdruck()
			}))

	if !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("es kam %v", fehler)
	}
}
