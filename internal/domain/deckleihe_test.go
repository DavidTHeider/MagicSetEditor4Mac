package domain_test

// **Die digitale Deckleihe** — Werkstück G6 von Phase 5.5
// (#1463).

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

var leihzeit = time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)

// eineLeihe gibt eine gültige Leihe mit den genannten Abweichungen.
func eineLeihe(aendere ...func(*domain.Deckleihe)) domain.Deckleihe {
	leihe := domain.Deckleihe{
		Kennung:   "l1",
		Verleiher: "12D3KooWAnna",
		Entleiher: "12D3KooWBert",
		Abdruck:   "abdruckdesabzugs",
		Bindung: domain.Leihbindung{
			Art: domain.LeiheFuerPartie, Woran: "p1",
		},
		Rechte:      []domain.Leihrecht{domain.RechtSpielen},
		Ausgestellt: leihzeit,
		Bis:         leihzeit.Add(24 * time.Hour),
		Modus:       domain.LaufZuEnde,
	}

	for _, aenderung := range aendere {
		aenderung(&leihe)
	}

	return leihe
}

func TestDeckleihePruefe(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		aendere func(*domain.Deckleihe)
		gilt    bool
	}{
		"so, wie sie ausgestellt wird": {
			aendere: func(*domain.Deckleihe) {}, gilt: true,
		},
		"ohne Kennung": {
			aendere: func(l *domain.Deckleihe) { l.Kennung = "" },
		},
		"ohne Verleiher": {
			aendere: func(l *domain.Deckleihe) { l.Verleiher = "" },
		},
		"ohne Entleiher": {
			aendere: func(l *domain.Deckleihe) { l.Entleiher = "" },
		},
		"an sich selbst": {
			aendere: func(l *domain.Deckleihe) { l.Entleiher = l.Verleiher },
		},
		"ohne Abdruck": {
			aendere: func(l *domain.Deckleihe) { l.Abdruck = "" },
		},
		"mit einer erfundenen Bindung": {
			aendere: func(l *domain.Deckleihe) { l.Bindung.Art = "sammlung" },
		},
		"ohne Bindung an eine Partie": {
			aendere: func(l *domain.Deckleihe) { l.Bindung.Woran = "" },
			gilt:    true,
		},
		"ohne Recht": {
			aendere: func(l *domain.Deckleihe) { l.Rechte = nil },
		},
		"mit einem erfundenen Recht": {
			aendere: func(l *domain.Deckleihe) {
				l.Rechte = []domain.Leihrecht{"verkaufen"}
			},
		},
		"ohne Ausstellungszeit": {
			aendere: func(l *domain.Deckleihe) { l.Ausgestellt = time.Time{} },
		},
		"ohne Ablauf": {
			aendere: func(l *domain.Deckleihe) { l.Bis = time.Time{} },
		},
		"die abläuft, bevor sie gilt": {
			aendere: func(l *domain.Deckleihe) {
				l.Bis = l.Ausgestellt.Add(-time.Hour)
			},
		},
		"mit einem erfundenen Modus": {
			aendere: func(l *domain.Deckleihe) { l.Modus = "wegnehmen" },
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := eineLeihe(fall.aendere).Pruefe()

			if fall.gilt && fehler != nil {
				t.Fatalf("erwartet gültig, bekam %v", fehler)
			}

			if !fall.gilt && !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
			}
		})
	}
}

// **Die vier Abweisungen aus der Abnahme des Issues.**
//
// „anderer Peer/falsche Session/falscher Hash/abgelaufene Signatur nicht."
func TestDeckleiheGilt(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		jetzt     time.Time
		entleiher string
		abdruck   string
		woran     string
		recht     domain.Leihrecht
		gilt      bool
	}{
		"der richtige Empfänger mit dem richtigen Abzug": {
			jetzt: leihzeit, entleiher: "12D3KooWBert",
			abdruck: "abdruckdesabzugs", woran: "p1",
			recht: domain.RechtSpielen, gilt: true,
		},
		"ein anderer Peer": {
			jetzt: leihzeit, entleiher: "12D3KooWCem",
			abdruck: "abdruckdesabzugs", woran: "p1",
			recht: domain.RechtSpielen,
		},
		"eine andere Partie": {
			jetzt: leihzeit, entleiher: "12D3KooWBert",
			abdruck: "abdruckdesabzugs", woran: "p2",
			recht: domain.RechtSpielen,
		},
		"ein anderer Abzug": {
			jetzt: leihzeit, entleiher: "12D3KooWBert",
			abdruck: "einanderer", woran: "p1",
			recht: domain.RechtSpielen,
		},
		"nach dem Ablauf": {
			jetzt: leihzeit.Add(25 * time.Hour), entleiher: "12D3KooWBert",
			abdruck: "abdruckdesabzugs", woran: "p1",
			recht: domain.RechtSpielen,
		},
		"vor der Ausstellung": {
			jetzt: leihzeit.Add(-time.Hour), entleiher: "12D3KooWBert",
			abdruck: "abdruckdesabzugs", woran: "p1",
			recht: domain.RechtSpielen,
		},
		"ein Recht, das nicht darinsteht": {
			jetzt: leihzeit, entleiher: "12D3KooWBert",
			abdruck: "abdruckdesabzugs", woran: "p1",
			recht: domain.RechtEinreichen,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fehler := eineLeihe().Gilt(fall.jetzt, fall.entleiher, fall.abdruck,
				domain.Leihbindung{Art: domain.LeiheFuerPartie, Woran: fall.woran},
				fall.recht)

			if fall.gilt && fehler != nil {
				t.Fatalf("erwartet gültig, bekam %v", fehler)
			}

			if !fall.gilt && fehler == nil {
				t.Fatal("erwartet eine Abweisung")
			}
		})
	}
}

// **Eine offene Bindung passt auf jede Partie** — „für die nächste".
func TestEineOffeneBindungPasstAufJedePartie(t *testing.T) {
	t.Parallel()

	offen := eineLeihe(func(l *domain.Deckleihe) { l.Bindung.Woran = "" })

	for _, partie := range []string{"p1", "p9"} {
		if fehler := offen.Gilt(leihzeit, "12D3KooWBert", "abdruckdesabzugs",
			domain.Leihbindung{Art: domain.LeiheFuerPartie, Woran: partie},
			domain.RechtSpielen); fehler != nil {
			t.Fatalf("%s: %v", partie, fehler)
		}
	}
}

// **Aber nicht auf eine andere Art.** Wer eine Partie leiht, hat kein Turnier
// geliehen.
func TestEineOffeneBindungPasstNichtAufEineAndereArt(t *testing.T) {
	t.Parallel()

	offen := eineLeihe(func(l *domain.Deckleihe) { l.Bindung.Woran = "" })

	fehler := offen.Gilt(leihzeit, "12D3KooWBert", "abdruckdesabzugs",
		domain.Leihbindung{Art: domain.LeiheFuerTurnier, Woran: "t1"},
		domain.RechtSpielen)

	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}
}

// **Die Unterschriftsform trägt jedes Feld** — sonst ließe sich eines ändern,
// ohne die Unterschrift zu brechen.
func TestDieUnterschriftsformTraegtJedesFeld(t *testing.T) {
	t.Parallel()

	aenderungen := map[string]func(*domain.Deckleihe){
		"Kennung":     func(l *domain.Deckleihe) { l.Kennung = "l2" },
		"Verleiher":   func(l *domain.Deckleihe) { l.Verleiher = "12D3KooWCem" },
		"Entleiher":   func(l *domain.Deckleihe) { l.Entleiher = "12D3KooWCem" },
		"Abdruck":     func(l *domain.Deckleihe) { l.Abdruck = "einanderer" },
		"Bindungsart": func(l *domain.Deckleihe) { l.Bindung.Art = domain.LeiheFuerTurnier },
		"Bindung":     func(l *domain.Deckleihe) { l.Bindung.Woran = "p2" },
		"Rechte": func(l *domain.Deckleihe) {
			l.Rechte = []domain.Leihrecht{domain.RechtEinreichen}
		},
		"Ausgestellt": func(l *domain.Deckleihe) { l.Ausgestellt = leihzeit.Add(time.Second) },
		"Bis":         func(l *domain.Deckleihe) { l.Bis = leihzeit.Add(time.Hour) },
		"Modus":       func(l *domain.Deckleihe) { l.Modus = domain.LaufAbbruch },
	}

	grund := string(eineLeihe().Unterschriftsform())

	for feld, aenderung := range aenderungen {
		t.Run(feld, func(t *testing.T) {
			t.Parallel()

			if string(eineLeihe(aenderung).Unterschriftsform()) == grund {
				t.Fatalf("%q geht nicht in die Unterschrift ein", feld)
			}
		})
	}
}

// **Die Unterschrift selbst geht nicht ein** — ein Feld, das sich selbst
// unterschriebe, ließe sich nie ausrechnen.
func TestDieUnterschriftGehtNichtInSichSelbstEin(t *testing.T) {
	t.Parallel()

	mit := eineLeihe(func(l *domain.Deckleihe) { l.Unterschrift = "abcdef" })

	if !bytes.Equal(mit.Unterschriftsform(), eineLeihe().Unterschriftsform()) {
		t.Fatal("die Unterschrift geht in ihre eigene Form ein")
	}
}

// **Und die Form trägt kein lokales Profil** — das Issue verlangt es
// wörtlich: „Lokale OwnerProfileID/BorrowerProfileID werden nicht über die
// Grenze übertragen."
func TestDieLeiheTraegtKeinLokalesProfil(t *testing.T) {
	t.Parallel()

	form := strings.ToLower(string(eineLeihe().Unterschriftsform()))

	for _, wort := range []string{"profil", "profile"} {
		if strings.Contains(form, wort) {
			t.Fatalf("die Unterschriftsform nennt %q: %s", wort, form)
		}
	}
}
