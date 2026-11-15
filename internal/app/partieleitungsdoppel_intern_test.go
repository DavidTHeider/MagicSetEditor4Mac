package app

// **Zwei Partieleitungen für Läufe** — Werkstück G1 von Phase 5.5
// (#1458).
//
// **Aus `partierolle_intern_test.go` herausgelöst**, als sie mit der Rückschau
// über die Zeilengrenze aus §2 ging (G5,
// #1462).
// Der Schnitt folgt der Frage: dort steht, **was** geprüft wird, hier,
// **womit** — und beide Doppel werden aus mehr als einer Datei benutzt.
//
// # Warum zwei und nicht eines mit einem Schalter
//
// **Weil sie entgegengesetzte Fragen beantworten.** Das eine antwortet auf
// jedem Weg und sagt nichts — damit prüfbar ist, was die Naht aus einer
// mageren Antwort macht. Das andere klagt auf **einem** benannten Weg: eine
// Leitung, die auf jedem klagt, klagt schon beim ersten, und alles danach
// bliebe ungemessen (§10).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// schweigsameLeitung gibt eine Sicht und keine erlaubten Befehle.
type schweigsameLeitung struct {
	// freigaben merkt die Einwilligungen (G5, #1462).
	//
	// **Ein Gedächtnis und kein `nil`**: ohne es prüfte der Lauf über die
	// Fassade nur, dass ein Ruf durchging — nicht, dass sein Wert unten
	// ankommt und oben wieder heraus.
	freigaben []domain.Offenlegung
}

func (s *schweigsameLeitung) Anlegen(
	_ context.Context, partie domain.Partie,
) (domain.Partie, error) {
	return partie, nil
}

func (s *schweigsameLeitung) Partie(
	_ context.Context, kennung, fuer string,
) (domain.Partiesicht, error) {
	return domain.Partiesicht{
		PartieID: kennung, Fuer: fuer,
		Objekte: []domain.Sichtobjekt{{Kennung: "o1", Zone: domain.ZoneTisch}},
	}, nil
}

// Tischbild gibt die Sicht und **`nil`** für Erlaubnis und Handlungen —
// damit prüfbar ist, was die Naht aus einer mageren Antwort macht.
func (s *schweigsameLeitung) Tischbild(
	ctx context.Context, kennung, fuer string,
) (domain.Tischbild, error) {
	sicht, fehler := s.Partie(ctx, kennung, fuer)

	return domain.Tischbild{Sicht: sicht}, fehler
}

func (s *schweigsameLeitung) BefehlVerkuendet(
	_ context.Context, _ domain.Spielbefehl, verkuende service.Verkuendung,
) (service.Befehlsausgang, error) {
	// **Ein Ereignis ohne Inhalt** — damit die leeren Abbildungen der
	// Übersetzung betreten werden.
	ausgang := service.Befehlsausgang{
		Fassung:    2,
		Ereignisse: []domain.Ereignis{{Folge: 1, Art: "nichts"}},
	}

	if verkuende != nil {
		verkuende(ausgang)
	}

	return ausgang, nil
}

func (s *schweigsameLeitung) InReihe(tun func() error) error { return tun() }

func (s *schweigsameLeitung) Verlauf(
	context.Context, string, string,
) ([]domain.Ereignis, error) {
	return nil, nil
}

func (s *schweigsameLeitung) Wirtsstand(
	context.Context, string, string,
) (domain.Wirtsstand, error) {
	return domain.Wirtsstand{Epoche: domain.Partieepoche, AelteteFolge: 1}, nil
}

func (s *schweigsameLeitung) Nachreichen(
	context.Context, string, string, int64,
) ([]domain.Ereignis, error) {
	return nil, nil
}

func (s *schweigsameLeitung) Platznehmen(
	context.Context, string, string, string, domain.Partiedeck, service.Verkuendung,
) (service.Befehlsausgang, error) {
	return service.Befehlsausgang{}, nil
}

func (s *schweigsameLeitung) Freigaben(
	context.Context, string,
) ([]domain.Offenlegung, error) {
	return s.freigaben, nil
}

func (s *schweigsameLeitung) Offenlegen(
	ctx context.Context, kennung string, gabe domain.Offenlegung,
) error {
	if fehler := s.Zuruecknehmen(ctx, kennung, gabe.Von); fehler != nil {
		return fehler
	}

	s.freigaben = append(s.freigaben, gabe)

	return nil
}

func (s *schweigsameLeitung) Zuruecknehmen(
	_ context.Context, _, von string,
) error {
	geblieben := make([]domain.Offenlegung, 0, len(s.freigaben))

	for _, gabe := range s.freigaben {
		if gabe.Von != von {
			geblieben = append(geblieben, gabe)
		}
	}

	s.freigaben = geblieben

	return nil
}

func (s *schweigsameLeitung) Vergessen(context.Context, string) error {
	s.freigaben = nil

	return nil
}

func (s *schweigsameLeitung) Partien(context.Context, int64) (
	[]domain.Partiezeile, error,
) {
	return nil, nil
}

func (s *schweigsameLeitung) Pausieren(context.Context, string) error  { return nil }
func (s *schweigsameLeitung) Fortsetzen(context.Context, string) error { return nil }
func (s *schweigsameLeitung) Beenden(context.Context, string) error    { return nil }

// klagendePartieleitung klagt auf dem Weg, den sie genannt bekommt.
type klagendePartieleitung struct {
	weg   string
	klagt error
}

func (k *klagendePartieleitung) stolpert(weg string) error {
	if k.weg == weg {
		return k.klagt
	}

	return nil
}

func (k *klagendePartieleitung) Anlegen(
	_ context.Context, partie domain.Partie,
) (domain.Partie, error) {
	return partie, k.stolpert("Anlegen")
}

func (k *klagendePartieleitung) Partie(
	_ context.Context, kennung, fuer string,
) (domain.Partiesicht, error) {
	return eineZuschauersicht(kennung, fuer), k.stolpert("Partie")
}

func eineZuschauersicht(kennung, fuer string) domain.Partiesicht {
	return domain.Partiesicht{
		PartieID: kennung, Fuer: fuer,
		Zuschauer: []domain.Zuschauer{{Kennung: "z1", Name: "Nachbar"}},
	}
}

func (k *klagendePartieleitung) Partien(context.Context, int64) (
	[]domain.Partiezeile, error,
) {
	return []domain.Partiezeile{{ID: "p1", Spiel: "magic"}}, k.stolpert("Partien")
}

func (k *klagendePartieleitung) Tischbild(
	_ context.Context, kennung, fuer string,
) (domain.Tischbild, error) {
	return domain.Tischbild{
		Sicht: eineZuschauersicht(kennung, fuer), Erlaubt: []string{"objekt.tappen"},
		Handlungen: []domain.Objekthandlungen{{
			Objekt: "o1", Handlungen: []string{"objekt.tappen"},
			Primaer: "objekt.tappen",
		}},
	}, k.stolpert("Tischbild")
}

func (k *klagendePartieleitung) BefehlVerkuendet(
	_ context.Context, _ domain.Spielbefehl, verkuende service.Verkuendung,
) (service.Befehlsausgang, error) {
	if fehler := k.stolpert("Befehl"); fehler != nil {
		return service.Befehlsausgang{}, fehler
	}

	if verkuende != nil {
		verkuende(service.Befehlsausgang{})
	}

	return service.Befehlsausgang{}, nil
}

func (k *klagendePartieleitung) InReihe(tun func() error) error { return tun() }

func (k *klagendePartieleitung) Verlauf(
	context.Context, string, string,
) ([]domain.Ereignis, error) {
	return nil, k.stolpert("verlauf")
}

func (k *klagendePartieleitung) Wirtsstand(
	context.Context, string, string,
) (domain.Wirtsstand, error) {
	return domain.Wirtsstand{
		Epoche: domain.Partieepoche, AelteteFolge: 1,
	}, k.stolpert("Wirtsstand")
}

func (k *klagendePartieleitung) Nachreichen(
	context.Context, string, string, int64,
) ([]domain.Ereignis, error) {
	return nil, k.stolpert("Nachreichen")
}

func (k *klagendePartieleitung) Platznehmen(
	context.Context, string, string, string, domain.Partiedeck, service.Verkuendung,
) (service.Befehlsausgang, error) {
	return service.Befehlsausgang{}, k.stolpert("Platznehmen")
}

func (k *klagendePartieleitung) Pausieren(context.Context, string) error {
	return k.stolpert("Pausieren")
}

func (k *klagendePartieleitung) Fortsetzen(context.Context, string) error {
	return k.stolpert("Fortsetzen")
}

func (k *klagendePartieleitung) Beenden(context.Context, string) error {
	return k.stolpert("Beenden")
}

func (k *klagendePartieleitung) Freigaben(
	context.Context, string,
) ([]domain.Offenlegung, error) {
	return []domain.Offenlegung{{Von: "anna", An: []string{"ben"}}},
		k.stolpert("Freigaben")
}

func (k *klagendePartieleitung) Offenlegen(
	context.Context, string, domain.Offenlegung,
) error {
	return k.stolpert("Offenlegen")
}

func (k *klagendePartieleitung) Zuruecknehmen(
	context.Context, string, string,
) error {
	return k.stolpert("Zuruecknehmen")
}

func (k *klagendePartieleitung) Vergessen(context.Context, string) error {
	return k.stolpert("Vergessen")
}
