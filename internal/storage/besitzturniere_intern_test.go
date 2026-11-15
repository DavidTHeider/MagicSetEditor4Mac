package storage

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Ein Turnier, eine Paarung oder eine Partie aus einem fremden Profil trifft
// nichts** — #1815.
//
// Teilnehmer, Runden, Paarungen, Matchbindungen und Meldungen sind
// Abkömmlinge ihres Turniers; Ereignisse, Befehle und Freigaben die ihrer
// Partie (0035). Gerufen aus Profil 1 mit den Kennungen von Profil 2 —
// lesend und schreibend. Erwartet: „nicht gefunden", und Profil 2 steht
// danach wie vorher.
func TestEinFremdesTurnierTrifftNichts(t *testing.T) {
	t.Parallel()

	db, fremd := besitzbestand(t)

	ausFremdemProfil(t, db, turnierrufe(db, fremd, MitProfil(t.Context(), 1)))
}

// turnierrufe sind die Rufe der Turnier- und Matchablagen, nach `Typ.Methode`.
func turnierrufe(db *sql.DB, fremd fremdeKennungen, eigen context.Context) map[string]func() error {
	turniere := NeueTurnierablage(db)
	match := NeueMatchablage(db)

	return map[string]func() error{
		"Turnierablage.Turnier":  func() error { _, f := turniere.Turnier(eigen, fremd.turnier); return f },
		"Turnierablage.Loeschen": func() error { return turniere.Loeschen(eigen, fremd.turnier) },
		"Turnierablage.Anmelden": func() error {
			_, f := turniere.Anmelden(eigen, fremd.turnier, domain.Teilnehmer{Name: "Eindringling"})
			return f
		},
		"Turnierablage.Aussteigen": func() error { return turniere.Aussteigen(eigen, fremd.teilnehmer, true) },
		"Turnierablage.RundeLegen": func() error {
			_, f := turniere.RundeLegen(eigen, fremd.turnier, 2,
				[]domain.Paarung{{A: fremd.teilnehmer, B: fremd.zweiter}})
			return f
		},
		"Turnierablage.ErgebnisSetzen": func() error {
			return turniere.ErgebnisSetzen(eigen, 1, fremd.paarung, &domain.Ergebnis{SpieleA: 2})
		},
		// **Anlegen und Einlesen nehmen über die Teilnehmenden eine
		// Deckkennung an** — DOM-04 (#1699): eine Vorlage, eine Ausgabe.
		"Turnierablage.Anlegen": func() error {
			_, f := turniere.Anlegen(eigen, domain.Turnier{
				Name: "Eindringling", Spiel: "magic",
				Teilnehmer: []domain.Teilnehmer{{Name: "E", DeckID: fremd.deck}},
			})
			return f
		},
		"Turnierablage.Einlesen": func() error {
			_, f := turniere.Einlesen(eigen, domain.Turnier{
				Name: "Eindringling", Spiel: "magic",
				Teilnehmer: []domain.Teilnehmer{{ID: 1, Name: "E", DeckID: fremd.deck}},
			})
			return f
		},
		// **Abschluss und Protokoll** — DOM-04.AK-04 (#1699).
		"Turnierablage.Abschliessen":    func() error { return turniere.Abschliessen(eigen, fremd.turnier, 1) },
		"Turnierablage.Wiedereroeffnen": func() error { return turniere.Wiedereroeffnen(eigen, fremd.turnier, 1) },
		"Turnierablage.Protokoll":       func() error { _, f := turniere.Protokoll(eigen, fremd.turnier); return f },
		"Matchablage.Bindung":           func() error { _, f := match.Bindung(eigen, fremd.paarung); return f },
		"Matchablage.Meldungen":         func() error { _, f := match.Meldungen(eigen, fremd.paarung); return f },
		"Matchablage.BindungSchreiben": func() error {
			return match.BindungSchreiben(eigen, domain.Matchbindung{
				TurnierID: fremd.turnier, PaarungID: fremd.paarung, Fassung: 2, PartieID: fremd.partie,
			})
		},
		"Matchablage.MeldungSchreiben": func() error {
			return match.MeldungSchreiben(eigen, domain.Matchmeldung{
				PaarungID: fremd.paarung, Von: "sitz2", Fassung: 1, PartieID: fremd.partie,
				Ergebnis: domain.Ergebnis{SpieleB: 2}, Kette: "gefaelscht",
			})
		},
	}
}

// **Dasselbe für eine Partie** — ihr Strom, ihre Befehle und die
// Einwilligungen ihrer Sitze.
func TestEineFremdePartieTrifftNichts(t *testing.T) {
	t.Parallel()

	db, fremd := besitzbestand(t)

	ausFremdemProfil(t, db, partierufe(db, fremd, MitProfil(t.Context(), 1)))
}

// partierufe sind die Rufe der Partieablage, nach `Typ.Methode`.
func partierufe(db *sql.DB, fremd fremdeKennungen, eigen context.Context) map[string]func() error {
	partien := NeuePartieablage(db)

	return map[string]func() error{
		"Partieablage.Partie": func() error { _, f := partien.Partie(eigen, fremd.partie); return f },
		"Partieablage.Befehlsergebnis": func() error {
			_, _, f := partien.Befehlsergebnis(eigen, fremd.partie, "befehl-fremd")
			return f
		},
		"Partieablage.Ereignisse":       func() error { _, f := partien.Ereignisse(eigen, fremd.partie, 0, 0); return f },
		"Partieablage.LetzteKette":      func() error { _, _, f := partien.LetzteKette(eigen, fremd.partie); return f },
		"Partieablage.LetzterBefehl":    func() error { _, f := partien.LetzterBefehl(eigen, fremd.partie); return f },
		"Partieablage.Freigaben":        func() error { _, f := partien.Freigaben(eigen, fremd.partie); return f },
		"Partieablage.Offenlegungszahl": func() error { _, f := partien.Offenlegungszahl(eigen, fremd.partie); return f },
		"Partieablage.Freigeben": func() error {
			return partien.Freigeben(eigen, fremd.partie, domain.Offenlegung{Von: "sitz1", An: []string{"sitz3"}})
		},
		"Partieablage.Freigabeloeschen": func() error { return partien.Freigabeloeschen(eigen, fremd.partie, "sitz1") },
		"Partieablage.Standsetzen": func() error {
			_, f := partien.Standsetzen(eigen, fremd.partie, domain.StandPausiert)
			return f
		},
		"Partieablage.Anwenden": func() error {
			return partien.Anwenden(eigen, 1, "befehl-neu", domain.Partie{ID: fremd.partie, Fassung: 2}, nil)
		},
		"Partieablage.Kopieren":  func() error { _, f := partien.Kopieren(eigen, fremd.partie, "partie-kopie"); return f },
		"Partieablage.Vergessen": func() error { return partien.Vergessen(eigen, fremd.partie) },
		"Partieablage.Loeschen":  func() error { return partien.Loeschen(eigen, fremd.partie) },
	}
}

// **Die Gegenprobe: der Eigner erreicht sein Turnier und seine Partie.**
func TestDerEignerErreichtSeinTurnier(t *testing.T) {
	t.Parallel()

	db, fremd := besitzbestand(t)
	eigner := MitProfil(t.Context(), 2)

	meldungen, fehler := NeueMatchablage(db).Meldungen(eigner, fremd.paarung)
	if fehler != nil || len(meldungen) != 1 || meldungen[0].Kette != "kette-fremd" {
		t.Errorf("die Meldungen des Eigners: %+v (%v)", meldungen, fehler)
	}

	kette, folge, fehler := NeuePartieablage(db).LetzteKette(eigner, fremd.partie)
	if fehler != nil || kette != "kette-fremd" || folge != 1 {
		t.Errorf("die Kette des Eigners: %q %d (%v)", kette, folge, fehler)
	}

	zahl, fehler := NeuePartieablage(db).Offenlegungszahl(eigner, fremd.partie)
	if fehler != nil || zahl != 1 {
		t.Errorf("die Offenlegungen des Eigners: %d (%v)", zahl, fehler)
	}
}
