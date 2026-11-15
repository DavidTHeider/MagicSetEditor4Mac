package app_test

// **Die Turniernaht** — Werkstück 5 von Phase 4.7
// (#1294).
//
// **Was diese Schicht tut, ist Abbildung und Fehlerübersetzung** (§3): dass
// Kennungen als Text herauskommen, dass ein Bye kein `b` trägt, dass eine
// offene Paarung kein Ergebnis hat — und dass eine unbrauchbare Kennung ein
// Bedienfehler ist und kein Absturz.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// turniernahtdoppel gibt ein festes Turnier.
type turniernahtdoppel struct {
	turnier domain.Turnier
	rang    []domain.Rang
	plan    []domain.Paarung
	fehler  error

	// angelegt ist, was zuletzt angelegt werden sollte — DOM-04 (#1699).
	angelegt domain.Turnier

	// abschluss hält, was zuletzt abgeschlossen (true) oder wiedereröffnet
	// (false) werden sollte, samt Fassung — DOM-04.AK-04 (#1699).
	abschluss struct {
		turnier, fassung int64
		zu               bool
	}

	// eingelesen hält, was zuletzt eingelesen werden sollte, und worte die
	// Wörter des letzten Berichts — DOM-04.AK-03 (#1699).
	eingelesen []byte
	worte      domain.Berichttexte

	// gefragt hält den Stand, nach dessen Folgen zuletzt gefragt wurde.
	gefragt *domain.Ergebnis

	// kopiert hält, was zuletzt kopiert werden sollte.
	kopiert struct {
		quelle           int64
		name             string
		mitTeilnehmenden bool
	}
}

func (d *turniernahtdoppel) Anlegen(_ context.Context, turnier domain.Turnier) (
	domain.Turnier, error,
) {
	if d.fehler != nil {
		return domain.Turnier{}, d.fehler
	}

	d.angelegt = turnier
	turnier.ID, turnier.Fassung = 7, 1

	return turnier, nil
}

func (d *turniernahtdoppel) Turnier(_ context.Context, _ int64) (
	domain.Turnier, error,
) {
	return d.turnier, d.fehler
}

func (d *turniernahtdoppel) Turniere(_ context.Context, _ string) (
	[]domain.Turnier, error,
) {
	if d.fehler != nil {
		return nil, d.fehler
	}

	return []domain.Turnier{d.turnier}, nil
}

func (d *turniernahtdoppel) Anmelden(
	_ context.Context, turnierID int64, teilnehmer domain.Teilnehmer,
) (domain.Teilnehmer, error) {
	if d.fehler != nil {
		return domain.Teilnehmer{}, d.fehler
	}

	teilnehmer.ID, teilnehmer.TurnierID, teilnehmer.Stelle = 3, turnierID, 1

	return teilnehmer, nil
}

func (d *turniernahtdoppel) Aussteigen(_ context.Context, _, _ int64, _ bool) error {
	return d.fehler
}

func (d *turniernahtdoppel) Rangliste(_ context.Context, _ int64) ([]domain.Rang, error) {
	return d.rang, d.fehler
}

func (d *turniernahtdoppel) Paarungsvorschlag(_ context.Context, _ int64) (
	[]domain.Paarung, error,
) {
	return d.plan, d.fehler
}

func (d *turniernahtdoppel) RundeLegen(_ context.Context, _ int64) (
	domain.Runde, error,
) {
	if d.fehler != nil {
		return domain.Runde{}, d.fehler
	}

	return domain.Runde{ID: 9, Nummer: 1, Paarungen: d.plan}, nil
}

func (d *turniernahtdoppel) ErgebnisSetzen(
	_ context.Context, _, _, _ int64, _ *domain.Ergebnis,
) error {
	return d.fehler
}

func (d *turniernahtdoppel) Loeschen(_ context.Context, _ int64) error {
	return d.fehler
}

// turnieranwendung baut die Anwendung über einem Doppel.
func turnieranwendung(t *testing.T) (*app.App, *turniernahtdoppel) {
	t.Helper()

	naht := &turniernahtdoppel{}

	anwendung := app.NeueApp("")
	anwendung.MitTurniernaht(naht)

	return anwendung, naht
}

// **Ein Turnier kommt mit Kennungen als Text und gezählten Teilen.**
func TestEinTurnierKommtAlsDTO(t *testing.T) {
	t.Parallel()

	anwendung, naht := turnieranwendung(t)

	naht.turnier = domain.Turnier{
		ID: 7, Fassung: 2, Name: "Küchentisch", Spiel: "magic",
		Teilnehmer: []domain.Teilnehmer{
			{
				ID: 3, Name: "Ada", DeckID: 5, Stelle: 1,
				Deckabzug: []domain.Abzugsplatz{
					{
						ID: 11, Bezeichnung: "Lightning Bolt", Menge: 4,
						Bereich: domain.BereichHauptdeck,
					},
				},
			},
			{ID: 4, Name: "Bert", Stelle: 2, Ausgestiegen: true},
		},
		Runden: []domain.Runde{{
			ID: 9, Nummer: 1,
			Paarungen: []domain.Paarung{
				{ID: 12, A: 3, B: 4, Ergebnis: &domain.Ergebnis{SpieleA: 2}},
				{ID: 13, A: 3},
			},
		}},
	}

	antwort := anwendung.Turnier(t.Context(), "7")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	turnier := antwort.Turnier

	if turnier.ID != "7" || turnier.Fassung != "2" {
		t.Errorf("Kennung %q, Fassung %q", turnier.ID, turnier.Fassung)
	}

	// **Eine fehlende Deckkennung ist leer und nicht „0"** — die Null wäre
	// eine Kennung, und die Oberfläche versuchte, sie aufzuschlagen.
	if turnier.Teilnehmer[0].DeckID != "5" || turnier.Teilnehmer[1].DeckID != "" {
		t.Errorf("Deckkennungen: %q und %q",
			turnier.Teilnehmer[0].DeckID, turnier.Teilnehmer[1].DeckID)
	}

	if len(turnier.Teilnehmer[0].Deckabzug) != 1 ||
		turnier.Teilnehmer[0].Deckabzug[0].Bezeichnung != "Lightning Bolt" {
		t.Errorf("der Abzug ist %+v", turnier.Teilnehmer[0].Deckabzug)
	}

	// **Ein Bye trägt kein `b`** — und nicht die Null. Der Unterschied ist
	// der ganze Punkt: eine „0" sähe wie ein Gegner aus.
	paarungen := turnier.Runden[0].Paarungen
	if paarungen[0].B != "4" || paarungen[1].B != "" {
		t.Errorf("die Gegner sind %q und %q", paarungen[0].B, paarungen[1].B)
	}

	if paarungen[0].Ergebnis == nil || paarungen[1].Ergebnis != nil {
		t.Errorf("die Ergebnisse sind %+v und %+v",
			paarungen[0].Ergebnis, paarungen[1].Ergebnis)
	}
}

// **Die Quoten kommen gerundet heraus — auf vier Stellen.**
//
// **Gerechnet wird ungerundet** (Vermerk 0062, Entscheidung 6); vier Stellen,
// weil zwei Teilnehmer, die sich in der vierten unterscheiden, nicht
// gleichauf sind.
func TestDieQuotenKommenGerundetHeraus(t *testing.T) {
	t.Parallel()

	anwendung, naht := turnieranwendung(t)

	naht.rang = []domain.Rang{{
		TeilnehmerID: 3, Name: "Ada", Platz: 1,
		Bilanz:      domain.Bilanz{Siege: 2, Niederlagen: 1},
		Matchpunkte: 6,
		// 2/3 — ungerundet eine Zahl mit sechzehn Stellen.
		Matchquote:       2.0 / 3.0,
		Gegnermatchquote: 1.0 / 3.0,
	}}

	antwort := anwendung.TurnierRangliste(t.Context(), "7")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Rangliste[0].Matchquote != 0.6667 {
		t.Errorf("die Matchquote ist %v, erwartet 0,6667",
			antwort.Rangliste[0].Matchquote)
	}

	if antwort.Rangliste[0].Gegnermatchquote != 0.3333 {
		t.Errorf("die Gegnermatchquote ist %v, erwartet 0,3333",
			antwort.Rangliste[0].Gegnermatchquote)
	}
}

// **Jeder Weg weist eine unbrauchbare Kennung ab** — als Bedienfehler und
// nicht als Absturz.
func TestJederTurnierwegWeistEineUnbrauchbareKennungAb(t *testing.T) {
	t.Parallel()

	anwendung, _ := turnieranwendung(t)

	faelle := map[string]func() *app.Fehler{
		"lesen":       func() *app.Fehler { return anwendung.Turnier(t.Context(), "x").Fehler },
		"löschen":     func() *app.Fehler { return anwendung.TurnierLoeschen(t.Context(), "x").Fehler },
		"anmelden":    func() *app.Fehler { return anwendung.TurnierAnmelden(t.Context(), "x", "Ada", "").Fehler },
		"aussteigen":  func() *app.Fehler { return anwendung.TurnierAussteigen(t.Context(), "7", "x", true).Fehler },
		"vorschlagen": func() *app.Fehler { return anwendung.TurnierPaarungsvorschlag(t.Context(), "x").Fehler },
		"Runde legen": func() *app.Fehler { return anwendung.TurnierRundeLegen(t.Context(), "x").Fehler },
		"Ergebnis":    func() *app.Fehler { return anwendung.TurnierErgebnisSetzen(t.Context(), "x", "1", "1", nil).Fehler },
		"Rangliste":   func() *app.Fehler { return anwendung.TurnierRangliste(t.Context(), "x").Fehler },
	}

	for name, ruf := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if fehler := ruf(); fehler == nil {
				t.Error("angenommen, erwartet abgewiesen")
			}
		})
	}
}

// **Und eine unbrauchbare Paarungskennung ebenso.**
func TestEineUnbrauchbarePaarungskennungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anwendung, _ := turnieranwendung(t)

	if fehler := anwendung.TurnierErgebnisSetzen(
		probekontext(t), "7", "1", "x", nil).Fehler; fehler == nil {
		t.Error("angenommen, erwartet abgewiesen")
	}

	// **Und eine unbrauchbare Fassung ebenso.**
	//
	// Sie kommt als Zeichenkette herein, weil das Fenster Kennungen so
	// schickt; was keine Zahl ist, wird am Feld gemeldet und nicht als Null
	// weitergereicht.
	if fehler := anwendung.TurnierErgebnisSetzen(
		probekontext(t), "7", "x", "12", nil).Fehler; fehler == nil {
		t.Error("angenommen, erwartet abgewiesen")
	}
}

// **Und eine unbrauchbare Deckkennung beim Anmelden.**
func TestEineUnbrauchbareDeckkennungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	anwendung, _ := turnieranwendung(t)

	if fehler := anwendung.TurnierAnmelden(
		probekontext(t), "7", "Ada", "x").Fehler; fehler == nil {
		t.Error("angenommen, erwartet abgewiesen")
	}
}

// **Jeder Weg reicht einen Fehler der Naht durch.**
func TestJederTurnierwegReichtEinenNahtfehlerDurch(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*app.App) *app.Fehler{
		"auflisten": func(a *app.App) *app.Fehler { return a.Turniere(probekontext(t), "").Fehler },
		"lesen":     func(a *app.App) *app.Fehler { return a.Turnier(probekontext(t), "7").Fehler },
		"anlegen": func(a *app.App) *app.Fehler {
			return a.TurnierAnlegen(probekontext(t), "T", "magic", "", false, "").Fehler
		},
		"löschen":     func(a *app.App) *app.Fehler { return a.TurnierLoeschen(probekontext(t), "7").Fehler },
		"anmelden":    func(a *app.App) *app.Fehler { return a.TurnierAnmelden(probekontext(t), "7", "Ada", "5").Fehler },
		"aussteigen":  func(a *app.App) *app.Fehler { return a.TurnierAussteigen(probekontext(t), "7", "3", true).Fehler },
		"vorschlagen": func(a *app.App) *app.Fehler { return a.TurnierPaarungsvorschlag(probekontext(t), "7").Fehler },
		"Runde legen": func(a *app.App) *app.Fehler { return a.TurnierRundeLegen(probekontext(t), "7").Fehler },
		"Ergebnis": func(a *app.App) *app.Fehler {
			return a.TurnierErgebnisSetzen(probekontext(t), "7", "1", "12", &app.ErgebnisDTO{SpieleA: 2}).Fehler
		},
		"Rangliste": func(a *app.App) *app.Fehler { return a.TurnierRangliste(probekontext(t), "7").Fehler },
	}

	for name, ruf := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anwendung, naht := turnieranwendung(t)
			naht.fehler = errors.New("die Naht schweigt")

			if fehler := ruf(anwendung); fehler == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Anlegen, Runde legen und Vorschlagen kommen als DTO zurück.**
func TestDieSchreibendenWegeKommenAlsDTOZurueck(t *testing.T) {
	t.Parallel()

	anwendung, naht := turnieranwendung(t)
	naht.plan = []domain.Paarung{{ID: 12, A: 3, B: 4}, {ID: 13, A: 5}}

	angelegt := anwendung.TurnierAnlegen(t.Context(), "Küchentisch", "magic", "", true, "")
	if angelegt.Fehler != nil || angelegt.Turnier.ID != "7" {
		t.Errorf("%+v", angelegt)
	}

	if !angelegt.Turnier.Einzelspiel {
		t.Error("das Einzelspiel ging verloren")
	}

	teilnehmer := anwendung.TurnierAnmelden(t.Context(), "7", "Ada", "")
	if teilnehmer.Fehler != nil || teilnehmer.Teilnehmer.ID != "3" {
		t.Errorf("%+v", teilnehmer)
	}

	vorschlag := anwendung.TurnierPaarungsvorschlag(t.Context(), "7")
	if vorschlag.Fehler != nil || len(vorschlag.Paarungen) != 2 {
		t.Errorf("%+v", vorschlag)
	}

	runde := anwendung.TurnierRundeLegen(t.Context(), "7")
	if runde.Fehler != nil || runde.Runde.Nummer != 1 {
		t.Errorf("%+v", runde)
	}

	liste := anwendung.Turniere(t.Context(), "")
	if liste.Fehler != nil || len(liste.Turniere) != 1 {
		t.Errorf("%+v", liste)
	}
}

// **Ein Ergebnis von `nil` bleibt `nil`** — es heißt „doch nicht gespielt".
func TestEinLeeresErgebnisBleibtLeer(t *testing.T) {
	t.Parallel()

	anwendung, _ := turnieranwendung(t)

	if fehler := anwendung.TurnierErgebnisSetzen(
		probekontext(t), "7", "1", "12", nil).Fehler; fehler != nil {
		t.Errorf("%+v", fehler)
	}
}
