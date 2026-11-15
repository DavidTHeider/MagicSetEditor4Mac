package service_test

// **Die Handelsrichtlinie greift, wo gehandelt wird** — #2435.
//
// Gefunden bei der Erreichbarkeitsinventur von SEC04 (#2357): die fünf
// Schranken unter `richtlinie.handel` standen gefaltet in `domain`, die
// Einstellungsseite sagte „eine Schranke, keine Vorliebe" — und kein Dienst
// fragte sie. Hier wird jeder Weg gefahren, an dem ein Handel entsteht oder
// voranschreitet, und jeder, an dem eine Sicht oder ein Deck hinausgeht.
//
// **Was gilt, wenn keine Schranke eingesetzt ist**: die Richtlinie ohne
// Einstellung ([domain.Standardhandelsrichtlinie]) — handeln ja, Markt nein.
// Ein Dienst ohne Schranke ließe sonst den Markt zu, den das Programm ohne
// Einstellung ausschließt.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// schrankenprobe gibt eine Richtlinie, die der Test während des Laufs
// ändern kann — wie ein Mensch, der die Einstellung umlegt.
type schrankenprobe struct {
	richtlinie domain.Handelsrichtlinie
	fehler     error
}

func (s *schrankenprobe) Handelsrichtlinie(context.Context) (domain.Handelsrichtlinie, error) {
	return s.richtlinie, s.fehler
}

func ohneHandel() domain.Handelsrichtlinie {
	richtlinie := domain.Standardhandelsrichtlinie()
	richtlinie.Handeln = false

	return richtlinie
}

var errRichtlinieUnlesbar = errors.New("die Einstellungen sind nicht zu lesen")

func TestDieLeitungHaeltSichAnDieHandelsrichtlinie(t *testing.T) {
	t.Parallel()

	t.Run("eröffnet nicht, wenn das Profil nicht handelt — und schreibt nichts", func(t *testing.T) {
		t.Parallel()

		leitung, ablage := eineHandelsleitung()
		leitung = leitung.MitHandelsschranke(&schrankenprobe{richtlinie: ohneHandel()})

		if _, fehler := leitung.Eroeffnen(t.Context(), einHandelsangebot()); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("Fehler %v, erwartet ErrValidation", fehler)
		}

		if ablage.geschrieben != 0 {
			t.Errorf("geschrieben wurde %d-mal", ablage.geschrieben)
		}
	})

	t.Run("bietet einen Entwurf nicht an, wenn das Profil inzwischen nicht handelt", func(t *testing.T) {
		t.Parallel()

		schranke := &schrankenprobe{richtlinie: domain.Standardhandelsrichtlinie()}
		leitung, ablage := eineHandelsleitung()
		leitung = leitung.MitHandelsschranke(schranke)

		if _, fehler := leitung.Eroeffnen(t.Context(), einHandelsangebot()); fehler != nil {
			t.Fatal(fehler)
		}

		schranke.richtlinie = ohneHandel()

		if _, fehler := leitung.Schreiten(t.Context(), "h1", domain.SchrittAnbieten, "anna"); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("Fehler %v, erwartet ErrValidation", fehler)
		}

		if zustand := ablage.staende["h1"].Zustand; zustand != domain.HandelEntwurf {
			t.Errorf("der Handel steht auf %s", zustand)
		}
	})

	t.Run("bricht ab, auch wenn das Profil nicht handelt", func(t *testing.T) {
		t.Parallel()

		schranke := &schrankenprobe{richtlinie: domain.Standardhandelsrichtlinie()}
		leitung, _ := eineHandelsleitung()
		leitung = leitung.MitHandelsschranke(schranke)
		angebotenerHandel(t, leitung)

		schranke.richtlinie = ohneHandel()

		if _, fehler := leitung.Schreiten(t.Context(), "h1", domain.SchrittAbbrechen, "anna"); fehler != nil {
			t.Errorf("ein Ausweg wurde gesperrt: %v", fehler)
		}
	})

	t.Run("nimmt keinen Vorschlag über der Mengengrenze", func(t *testing.T) {
		t.Parallel()

		schranke := &schrankenprobe{richtlinie: domain.Standardhandelsrichtlinie()}
		leitung, _ := eineHandelsleitung()
		leitung = leitung.MitHandelsschranke(schranke)
		erste := angebotenerHandel(t, leitung)

		schranke.richtlinie.Mengengrenze = 3

		if _, fehler := leitung.Vorschlagen(t.Context(), erste.Fortgeschrieben(mitMenge(erste, 4)), "bert"); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("Fehler %v, erwartet ErrValidation", fehler)
		}

		if _, fehler := leitung.Vorschlagen(t.Context(), erste.Fortgeschrieben(mitMenge(erste, 3)), "bert"); fehler != nil {
			t.Errorf("ein Vorschlag an der Grenze wurde abgewiesen: %v", fehler)
		}
	})

	t.Run("handelt ohne eingesetzte Schranke nicht am Markt", func(t *testing.T) {
		t.Parallel()

		leitung, _ := eineHandelsleitung()
		markt := einHandelsangebot()
		markt.Art = domain.HandelMarkt

		if _, fehler := leitung.Eroeffnen(t.Context(), markt); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("Fehler %v, erwartet ErrValidation", fehler)
		}
	})

	t.Run("hält an, wenn die Richtlinie nicht zu lesen ist", func(t *testing.T) {
		t.Parallel()

		leitung, ablage := eineHandelsleitung()
		leitung = leitung.MitHandelsschranke(&schrankenprobe{fehler: errRichtlinieUnlesbar})

		if _, fehler := leitung.Eroeffnen(t.Context(), einHandelsangebot()); !errors.Is(fehler, errRichtlinieUnlesbar) {
			t.Errorf("Fehler %v, erwartet den Lesefehler", fehler)
		}

		if ablage.geschrieben != 0 {
			t.Errorf("geschrieben wurde %d-mal", ablage.geschrieben)
		}
	})
}

func TestDieEntwuerfeHaltenSichAnDieHandelsrichtlinie(t *testing.T) {
	t.Parallel()

	entwurf := einHandelsangebot()
	verzeichnis, ablage := verzeichnis(domain.Handelsvorgangzeile{Angebot: entwurf})
	verzeichnis = verzeichnis.MitHandelsschranke(&schrankenprobe{richtlinie: ohneHandel()})

	if _, fehler := verzeichnis.EntwurfAendern(t.Context(), entwurf); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Entwurf ändern: %v, erwartet ErrValidation", fehler)
	}

	if _, fehler := verzeichnis.Kopieren(t.Context(), "anna", entwurf.Vorgang); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Kopieren: %v, erwartet ErrValidation", fehler)
	}

	if len(ablage.geschrieben)+len(ablage.ersetzt) != 0 {
		t.Errorf("geschrieben wurde %v, ersetzt %v", ablage.geschrieben, ablage.ersetzt)
	}
}

// zusagezeit ist die Uhr der Zusagen in diesen Läufen.
var zusagezeit = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)

func TestDieZusageHaeltSichAnDieHandelsrichtlinie(t *testing.T) {
	t.Parallel()

	eineStunde := domain.Handelsrichtlinie{Handeln: true, Fristgrenze: time.Hour}

	for _, fall := range []struct {
		name       string
		richtlinie domain.Handelsrichtlinie
		frist      time.Time
		erwartet   time.Time
		abgewiesen bool
	}{
		{"hält nicht fest, wenn das Profil nicht handelt", ohneHandel(), time.Time{}, time.Time{}, true},
		// **Kein Weg nennt heute eine Frist** — eine Zusage ohne hielte
		// unbegrenzt fest, also kürzt die Schranke sie auf ihre Grenze.
		{"gibt einer Zusage ohne Frist die Grenze", eineStunde, time.Time{}, zusagezeit.Add(time.Hour), false},
		{"lässt eine kürzere Frist stehen", eineStunde, zusagezeit.Add(time.Minute), zusagezeit.Add(time.Minute), false},
		{"weist eine längere Frist ab", eineStunde, zusagezeit.Add(2 * time.Hour), time.Time{}, true},
		{"lässt ohne Grenze auch keine Frist", domain.Standardhandelsrichtlinie(), time.Time{}, time.Time{}, false},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			ablage := &zusagenhaltprobe{}
			werk := service.NeuesZusagenwerk(ablage, standprobe{stand: derHeutigeStand()}).
				MitHandelsschranke(&schrankenprobe{richtlinie: fall.richtlinie}).
				MitUhr(func() time.Time { return zusagezeit })

			zusage, fehler := werk.Festhalten(t.Context(), service.Zusagenwunsch{
				Vorgang: derPostvorgang, Wer: "bich", ExemplarID: 7, Menge: 1, Fassung: 4,
				Frist: fall.frist,
			})

			if fall.abgewiesen {
				if !errors.Is(fehler, domain.ErrValidation) {
					t.Errorf("Fehler %v, erwartet ErrValidation", fehler)
				}

				if len(ablage.zusagen) != 0 {
					t.Errorf("festgehalten wurde %v", ablage.zusagen)
				}

				return
			}

			if fehler != nil {
				t.Fatal(fehler)
			}

			if !zusage.Frist.Equal(fall.erwartet) {
				t.Errorf("die Frist ist %v, erwartet %v", zusage.Frist, fall.erwartet)
			}
		})
	}
}

func TestLeiheUndFreigabeHaltenSichAnDieHandelsrichtlinie(t *testing.T) {
	t.Parallel()

	t.Run("verleiht kein Deck, wenn das Profil nicht handelt", func(t *testing.T) {
		t.Parallel()

		anna, bert := zweiPeers(t)
		stelle, ablage, _ := eineLeihstelle(t, anna)
		stelle = stelle.MitHandelsschranke(&schrankenprobe{richtlinie: ohneHandel()})

		_, _, fehler := stelle.Anbieten(t.Context(), service.Leihwunsch{
			Entleiher: string(bert.Kennung()), DeckID: 7,
			Bindung: domain.Leihbindung{Art: domain.LeiheFuerPartie, Woran: "p1"},
			Rechte:  []domain.Leihrecht{domain.RechtSpielen},
			Dauer:   24 * time.Hour, Modus: domain.LaufZuEnde,
		})
		if !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("Fehler %v, erwartet ErrValidation", fehler)
		}

		if leihen, fehler := ablage.Leihen(t.Context()); fehler != nil || len(leihen) != 0 {
			t.Errorf("ausgestellt wurden %d (%v)", len(leihen), fehler)
		}
	})

	t.Run("nimmt keine Leihe an, wenn das Profil nicht handelt", func(t *testing.T) {
		t.Parallel()

		anna, bert := zweiPeers(t)
		verleiher, _, _ := eineLeihstelle(t, anna)

		leihe, abzug, fehler := verleiher.Anbieten(t.Context(), service.Leihwunsch{
			Entleiher: string(bert.Kennung()), DeckID: 7,
			Bindung: domain.Leihbindung{Art: domain.LeiheFuerPartie, Woran: "p1"},
			Rechte:  []domain.Leihrecht{domain.RechtSpielen},
			Dauer:   24 * time.Hour, Modus: domain.LaufZuEnde,
		})
		if fehler != nil {
			t.Fatal(fehler)
		}

		entleiher, ablage, _ := eineLeihstelle(t, bert)
		if fehler := entleiher.MitHandelsschranke(&schrankenprobe{richtlinie: ohneHandel()}).
			Annehmen(t.Context(), leihe, abzug); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("Fehler %v, erwartet ErrValidation", fehler)
		}

		if leihen, fehler := ablage.Leihen(t.Context()); fehler != nil || len(leihen) != 0 {
			t.Errorf("angenommen wurden %d (%v)", len(leihen), fehler)
		}
	})

	t.Run("gibt keine Sicht frei, wenn das Profil nicht handelt", func(t *testing.T) {
		t.Parallel()

		pflege := eineFreigabepflege(handelnderKontakt(), nil, dreiFreigabezeilen(), nil).
			MitHandelsschranke(&schrankenprobe{richtlinie: ohneHandel()})

		if fehler := pflege.Erteilen(t.Context(), eineSammlungsfreigabe()); !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("Fehler %v, erwartet ErrValidation", fehler)
		}

		if freigaben, fehler := pflege.Freigaben(t.Context(), "peer-1"); fehler != nil || len(freigaben) != 0 {
			t.Errorf("erteilt wurden %d (%v)", len(freigaben), fehler)
		}
	})
}

// **Von drüben gilt dieselbe Schranke** — eine Schranke gilt unabhängig
// davon, wer fragt.
func TestDiePostHaeltSichAnDieHandelsrichtlinie(t *testing.T) {
	t.Parallel()

	briefkasten := &briefkastenprobe{}
	post := einePoststelle(briefkasten, einHandelnder()).
		MitHandelsschranke(&schrankenprobe{richtlinie: ohneHandel()})

	if _, fehler := post.Empfangen(t.Context(), dasGegenueber, einePost(einNachfolger(), nil)); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("Fehler %v, erwartet ErrValidation", fehler)
	}

	if len(briefkasten.angebote) != 0 {
		t.Errorf("zugestellt wurden %d", len(briefkasten.angebote))
	}
}

// **Eine unlesbare Richtlinie hält an**, an jedem Weg — und lässt nicht die
// Vorgabe gelten.
func TestEineUnlesbareRichtlinieHaeltLeiheUndZusageAn(t *testing.T) {
	t.Parallel()

	unlesbar := &schrankenprobe{fehler: errRichtlinieUnlesbar}

	anna, bert := zweiPeers(t)
	stelle, _, _ := eineLeihstelle(t, anna)

	if _, _, fehler := stelle.MitHandelsschranke(unlesbar).Anbieten(t.Context(), service.Leihwunsch{
		Entleiher: string(bert.Kennung()), DeckID: 7,
		Bindung: domain.Leihbindung{Art: domain.LeiheFuerPartie, Woran: "p1"},
		Rechte:  []domain.Leihrecht{domain.RechtSpielen},
		Dauer:   24 * time.Hour, Modus: domain.LaufZuEnde,
	}); !errors.Is(fehler, errRichtlinieUnlesbar) {
		t.Errorf("Leihe: %v, erwartet den Lesefehler", fehler)
	}

	werk := service.NeuesZusagenwerk(&zusagenhaltprobe{}, standprobe{stand: derHeutigeStand()}).
		MitHandelsschranke(unlesbar)

	if _, fehler := werk.Festhalten(t.Context(), service.Zusagenwunsch{
		Vorgang: derPostvorgang, Wer: "bich", ExemplarID: 7, Menge: 1, Fassung: 4,
	}); !errors.Is(fehler, errRichtlinieUnlesbar) {
		t.Errorf("Zusage: %v, erwartet den Lesefehler", fehler)
	}
}
