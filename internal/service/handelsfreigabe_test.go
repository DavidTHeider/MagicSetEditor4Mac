package service_test

// **Kein verbindlicher Schritt ohne die Zustimmung eines Menschen** —
// Werkstück `T10` von Phase 5.6
// (#1475).
//
// # Was hier zugesagt wird
//
//   - **Verbuchen und Bestätigen verlangen einen Nachweis** — die zwei
//     Schritte, nach denen die Welt eine andere ist: der eine bewegt echtes
//     Papier, der andere macht aus einem einseitigen Beleg einen beidseitigen.
//   - **Abgleichen verlangt keinen.** Es sagt der Gegenseite, dass hier
//     fertig ist; es bindet nichts, und einen Menschen dafür zu fragen wäre
//     ein Dialog ohne Entscheidung.
//   - **Ein gefälschtes, abgelaufenes oder wiederverwendetes Freigabezeichen
//     bewirkt nichts** — und zwar **bevor** der Bestand sich bewegt.
//   - **Und die Bindung trägt den Vorgang samt seiner Revision**: ein
//     Nachweis für einen anderen Handel oder einen anderen Stand erlaubt
//     hier nichts.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// wachenprobe schreibt mit, wonach gefragt wurde.
type wachenprobe struct {
	wirkung domain.Wirkung
	kennung string
	bindung service.Nachweisbindung

	rufe   int
	fehler error
}

func (w *wachenprobe) PruefeMitNachweis(
	_ context.Context, wirkung domain.Wirkung, kennung string,
	bindung service.Nachweisbindung,
) error {
	w.rufe++
	w.wirkung, w.kennung, w.bindung = wirkung, kennung, bindung

	return w.fehler
}

// **Verbuchen fragt nach der verbindlichen Wirkung — mit Vorgang und
// Revision in der Bindung.**
func TestDasVerbuchenFragtNachDerVerbindlichenWirkung(t *testing.T) {
	t.Parallel()

	stelle, stand, bestand := eineVerbuchung(t)
	wache := &wachenprobe{}

	stelle.MitWache(wache)
	stelleEin(t, stand, bestand, true)

	if _, fehler := stelle.Verbuchen(
		t.Context(), derVorgang, "f-1"); fehler != nil {
		t.Fatalf("Verbuchen: %v", fehler)
	}

	if wache.wirkung != domain.WirkungVerbindlich {
		t.Errorf("gefragt wurde nach %q", wache.wirkung)
	}

	if wache.kennung != "f-1" {
		t.Errorf("vorgezeigt wurde %q", wache.kennung)
	}

	if wache.bindung.Zweck != service.HandelsverbuchungZweck ||
		wache.bindung.Ziel != derVorgang {
		t.Errorf("die Bindung lautet %+v", wache.bindung)
	}

	// **Die Revision gehört dazu** — ein Nachweis für einen Stand von vorhin
	// gilt nicht für den von jetzt (O1).
	if wache.bindung.Revision != 3 {
		t.Errorf("die Bindung nennt Revision %d", wache.bindung.Revision)
	}
}

// **Ein abgelehnter Nachweis bewegt den Bestand nicht.**
//
// **Und zwar nachweislich**: das Doppel zählt seine Abschlüsse, und ein
// gescheiterter Riegel darf keinen einzigen hinterlassen.
func TestEinAbgelehnterNachweisBewegtNichts(t *testing.T) {
	t.Parallel()

	for name, schritt := range map[string]func(
		*service.Verbuchungsstelle) error{
		"verbuchen": func(s *service.Verbuchungsstelle) error {
			_, fehler := s.Verbuchen(t.Context(), derVorgang, "f-alt")

			return fehler
		},
		"bestätigen": func(s *service.Verbuchungsstelle) error {
			_, fehler := s.Bestaetigen(t.Context(), derVorgang, "", "f-alt")

			return fehler
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stelle, stand, bestand := eineVerbuchung(t)

			stelle.MitWache(&wachenprobe{
				fehler: domain.MitGrund(domain.ErrErlaubnis, nil,
					"dieser Nachweis ist verbraucht"),
			})
			stelleEin(t, stand, bestand, true)

			if fehler := schritt(stelle); !errors.Is(
				fehler, domain.ErrErlaubnis) {
				t.Fatalf("der Schritt kam mit %v durch", fehler)
			}

			if len(bestand.gebucht) != 0 {
				t.Errorf("es wurde gebucht: %+v", bestand.gebucht)
			}
		})
	}
}

// **Abgleichen fragt nach der schreibenden Wirkung und nach keinem
// Nachweis.**
//
// **Die Gegenprobe zu den beiden darüber**: eine Wache, die auf jeden Schritt
// dieselbe Antwort gäbe, sagte nichts über den Unterschied — und der
// Unterschied ist die ganze Zusage. Wer hier einen Menschen fragte, fragte
// ihn nach einer Meldung, die nichts bindet.
func TestDasAbgleichenFragtOhneNachweis(t *testing.T) {
	t.Parallel()

	stelle, stand, bestand := eineVerbuchung(t)
	wache := &wachenprobe{}

	stelle.MitWache(wache)
	stelleEin(t, stand, bestand, true)

	// **Aus dem Zustand, aus dem das Abgleichen führt** — sonst scheiterte es
	// am Übergang und nicht an der Wache.
	angebot := stand.staende[derVorgang]
	angebot.Zustand = domain.HandelLokalVerbucht
	stand.staende[derVorgang] = angebot

	if _, fehler := stelle.Abgleichen(t.Context(), derVorgang); fehler != nil {
		t.Fatalf("Abgleichen: %v", fehler)
	}

	if wache.wirkung != domain.WirkungSchreiben || wache.kennung != "" {
		t.Errorf("gefragt wurde nach %q mit %q", wache.wirkung, wache.kennung)
	}
}

// **Ohne Wache wird nicht verbucht.**
//
// **Der Rückfall ist das Nein und nicht das Ja.** Eine Stelle ohne Wache
// entsteht in einem Test und in einer Fläche, die vergessen hat, sie zu
// setzen; das erste ist harmlos, das zweite wäre die Lücke selbst. Wer keine
// hat, bekommt eine Absage — und der Zweig ist betreten, dieser Lauf tut es.
func TestOhneWacheWirdNichtVerbucht(t *testing.T) {
	t.Parallel()

	stand := &handelsprobe{staende: map[string]domain.Handelsangebot{}}
	bestand := &verbuchungsprobe{
		stuecke: map[int64]domain.Exemplar{},
		belege:  map[string]domain.Belegstand{},
		stand:   stand,
	}

	stelleEin(t, stand, bestand, true)

	// **Ohne `MitWache`** — genau so entsteht sie in einer Fläche, die den
	// Riegel vergessen hat.
	stelle := service.NeueVerbuchungsstelle(
		stand, bestand, festerAusweis{ich: derAusweis(t)}, bestand)

	if _, fehler := stelle.Verbuchen(
		t.Context(), derVorgang, ""); !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Errorf("ohne Wache kam der Schritt mit %v durch", fehler)
	}

	if len(bestand.gebucht) != 0 {
		t.Error("es wurde gebucht")
	}
}

// **Und jede Absage der Wache kommt an — auf jedem der drei Wege.**
//
// **Drei und nicht zwei**: das Abgleichen fragt nach einer anderen Wirkung,
// aber es fragt. Ohne diesen Lauf bliebe sein Zweig unbetreten, und ein
// Riegel, den kein Lauf betritt, sieht aus wie eine geprüfte Zusage (§10).
func TestJedeAbsageDerWacheKommtAn(t *testing.T) {
	t.Parallel()

	for name, schritt := range map[string]func(
		*service.Verbuchungsstelle) error{
		"verbuchen": func(s *service.Verbuchungsstelle) error {
			_, fehler := s.Verbuchen(t.Context(), derVorgang, derFreibrief)

			return fehler
		},
		"abgleichen": func(s *service.Verbuchungsstelle) error {
			_, fehler := s.Abgleichen(t.Context(), derVorgang)

			return fehler
		},
		"bestätigen": func(s *service.Verbuchungsstelle) error {
			_, fehler := s.Bestaetigen(
				t.Context(), derVorgang, "bgegenueber", derFreibrief)

			return fehler
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			stelle, stand, bestand := eineVerbuchung(t)

			stelle.MitWache(&wachenprobe{
				fehler: domain.MitGrund(domain.ErrErlaubnis, nil,
					"dieser Akteur darf das nicht"),
			})
			stelleEin(t, stand, bestand, true)

			if fehler := schritt(stelle); !errors.Is(
				fehler, domain.ErrErlaubnis) {
				t.Fatalf("der Schritt kam mit %v durch", fehler)
			}
		})
	}
}

// **Ein Vorgang, den es nicht gibt, wird auch nicht bestätigt.**
//
// **Und er erreicht die Wache nicht.** Der Stand wird zuerst gelesen — sonst
// verbrauchte ein Mensch seine Freigabe an einem Handel, den es nicht gibt.
func TestEinUnbekannterVorgangWirdNichtBestaetigt(t *testing.T) {
	t.Parallel()

	stelle, _, _ := eineVerbuchung(t)
	wache := &wachenprobe{}

	stelle.MitWache(wache)

	if _, fehler := stelle.Bestaetigen(
		t.Context(), "v-gibtesnicht", "bgegenueber", derFreibrief,
	); fehler == nil {
		t.Fatal("der unbekannte Vorgang kam durch")
	}

	if wache.rufe != 0 {
		t.Errorf("die Wache wurde %dmal gefragt", wache.rufe)
	}
}
