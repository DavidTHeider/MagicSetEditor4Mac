package domain_test

// **Das Regelprofil und das Grundprofil** — Werkstück G4 von Phase 5.5
// (#1461).
//
// # Wogegen hier geprüft wird
//
// **Dass das Grundprofil sagt, was es kann — und was nicht.** Drei Bereiche
// durchgesetzt, acht von Hand: „streng" steht damit nicht zur Wahl, und das
// ist die Zusage, nicht ihr Mangel.
//
// **Dass jede Zusage ihre Quelle und ihre Grenzen nennt.** Ein Bereich, der
// durchgesetzt wird und keine Grenze nennt, behauptet Vollständigkeit.
//
// **Und dass ein Profil ohne Fassung, ohne Datum oder ohne Quelle nicht
// eingefroren werden kann** — es ließe sich später nicht wiederherstellen.

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Das Grundprofil kommt durch und sagt die Wahrheit über sich.**
func TestDasGrundprofilSagtWasEsKann(t *testing.T) {
	t.Parallel()

	profil := domain.Grundprofil("magic")

	if fehler := profil.Pruefe(); fehler != nil {
		t.Fatalf("das Grundprofil wird abgewiesen: %v", fehler)
	}

	// **Sandkasten und begleitet ja, streng nein.**
	switch {
	case !profil.Matrix.TraegtStrenge(domain.StrengeSandkasten):
		t.Error("der Sandkasten wird nicht getragen")
	case !profil.Matrix.TraegtStrenge(domain.StrengeBegleitet):
		t.Error("begleitet wird nicht getragen")
	case profil.Matrix.TraegtStrenge(domain.StrengeStreng):
		t.Error("streng wird getragen — acht Bereiche sind von Hand")
	}

	// **Sechs Lücken, und es sind die sechs.**
	//
	// **Bis zum 2026-09-17 waren es acht** — G9 hat `zugstruktur` und
	// `prioritaet` auf `automatisch` gestellt
	// (#1592).
	//
	// **Sechs bleiben es, und eine davon ist keine ganze mehr**: `M2`
	// (#1604)
	// hat `typen` auf `begleitet` gehoben. [Regelmatrix.Luecken] zählt
	// alles, was nicht `automatisch` ist — und das ist richtig so: wer
	// warnt, setzt nichts durch, und `streng` trägt er nicht.
	//
	// **Fünf seit `M5`** (#1607):
	// `zustandsbasiert` ist dazugekommen. Übrig waren Stapel, Kampf,
	// Spielsteine, Typen und Wechselwirkungen.
	//
	// **Vier seit `M4`** (#1606):
	// der **Stapel** ist dazugekommen — ein Zauber wird angekündigt, bezahlt
	// und obenauf gelegt. **Der Bereich `mana` kam mit `M3` als zwölfter
	// hinzu und war nie eine Lücke**: er stand von Anfang an auf
	// `automatisch`.
	//
	// **Drei seit `M6`** (#1608):
	// der **Kampf**. Übrig waren Spielsteine, Typen und Wechselwirkungen.
	//
	// **Vier seit `M19`** (#1620):
	// der **Commander** ist als dreizehnter Bereich dazugekommen.
	// **Drei seit `M9`** (#1611):
	// die **Wechselwirkungen** sind dazugekommen — ausgelöste Fähigkeiten
	// warten, werden geordnet, verrechnet und hören auf zu existieren.
	//
	// **Und zwei seit dem 2026-09-19**
	// ([0123](../../docs/entscheidungen/0123-streng-zaehlt-nur-was-zur-variante-gehoert.md),
	// #1626):
	// der Commander zählt hier **nicht** mehr mit. Eine Sandkastenpartie
	// spielt die Variante nicht, und ein Bereich, den niemand schließen
	// soll, ist keine Lücke. Übrig sind Spielsteine und Typen — und beide
	// sind Arbeit.
	luecken := profil.Matrix.Luecken()
	if len(luecken) != 2 {
		t.Errorf("%d Lücken: %v", len(luecken), luecken)
	}

	for _, offen := range luecken {
		grad := profil.Matrix.Grad(offen)

		// **Zwei sind `begleitet`** — die Typzeile seit `M2` (gelesen und
		// gewarnt) und die Handlungen mit Schlüsselwort seit `M7`
		// (#1609):
		// sie gehen, **wenn jemand sie stellt**, und von selbst geschieht
		// keine. [Regelmatrix.Luecken] zählt beide als Lücke, und das ist
		// richtig — wer nicht durchsetzt, trägt `streng` nicht.
		if offen == domain.BereichTypen ||
			offen == domain.BereichSpielsteine {
			if grad != domain.UnterstuetzungBegleitet {
				t.Errorf("%q ist %q — erwartet begleitet", offen, grad)
			}

			continue
		}

		if grad != domain.UnterstuetzungManuell {
			t.Errorf("%q ist %q — erwartet manuell, denn es ist eine "+
				"Entscheidung und kein Versäumnis", offen, grad)
		}
	}
}

// **Die fünf durchgesetzten Bereiche nennen Quelle und Grenzen.**
//
// # Warum das einzeln geprüft wird
//
// **Weil `Pruefe` es verlangt und trotzdem jemand die Grenzen leeren
// könnte** — dann wäre das Profil ungültig, und ein Lauf, der nur `Pruefe`
// ruft, sagte „ungültig" statt „die Grenzen fehlen". Hier steht, welche fünf
// es sind und woran sie hängen.
func TestDieFuenfZusagenDesGrundprofils(t *testing.T) {
	t.Parallel()

	matrix := domain.Grundprofil("magic").Matrix

	erwartet := map[domain.Regelbereich]string{
		domain.BereichZonen:       "400",
		domain.BereichMarken:      "122",
		domain.BereichAnhaengsel:  "301",
		domain.BereichZugstruktur: "500",
		domain.BereichPrioritaet:  "117",
	}

	for bereich, nummer := range erwartet {
		eintrag, da := matrix.Eintrag(bereich)

		switch {
		case !da:
			t.Errorf("%q steht nicht in der Matrix", bereich)
		case eintrag.Grad != domain.UnterstuetzungAutomatisch:
			t.Errorf("%q ist %q, erwartet automatisch", bereich, eintrag.Grad)
		case len(eintrag.Grenzen) == 0:
			t.Errorf("%q nennt keine Grenze", bereich)
		case !strings.Contains(eintrag.Quelle.String(), nummer):
			t.Errorf("%q belegt sich mit %q, erwartet CR %s",
				bereich, eintrag.Quelle, nummer)
		case eintrag.Quelle.Fassung != domain.Grundprofilquelle:
			t.Errorf("%q nennt die Fassung %q", bereich, eintrag.Quelle.Fassung)
		}
	}
}

// **Ein Profil ohne Fassung, Datum, Quelle oder Spielerzahl kommt nicht
// durch.**
func TestEinProfilOhneEinfrierbarenStandWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*domain.Regelprofil){
		"ohne Kennung":     func(p *domain.Regelprofil) { p.Kennung = " " },
		"ohne Fassung":     func(p *domain.Regelprofil) { p.Fassung = 0 },
		"ohne Datum":       func(p *domain.Regelprofil) { p.Wirksam = "" },
		"ohne Quelle":      func(p *domain.Regelprofil) { p.Quelle = domain.Regelquelle{} },
		"ohne Spieler":     func(p *domain.Regelprofil) { p.Spieler.Mindestens = 0 },
		"verdrehte Zahlen": func(p *domain.Regelprofil) { p.Spieler.Hoechstens = 0 },
		"fremde Hausregel": func(p *domain.Regelprofil) {
			p.ErlaubteHausregeln = []domain.Hausregelart{"betrug"}
		},
		"fremde Zone": func(p *domain.Regelprofil) {
			p.Zusatzzonen = []domain.Zone{"keller"}
		},
		"kaputte Matrix": func(p *domain.Regelprofil) {
			p.Matrix = domain.Regelmatrix{Eintraege: []domain.Matrixeintrag{{
				Bereich: domain.BereichZonen,
				Grad:    domain.UnterstuetzungAutomatisch,
			}}}
		},
	}

	for name, verdirb := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			profil := domain.Grundprofil("magic")
			verdirb(&profil)

			fehler := profil.Pruefe()
			if fehler == nil {
				t.Fatal("kommt durch")
			}

			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("Fehler %v ist keine Validierung", fehler)
			}
		})
	}
}

// **Eine Hausregel gilt nur, wo das Profil sie erlaubt.**
//
// **Und die Absage nennt beide Stufen**: eine Hausregel kann das Regelwerk
// des Spiels nicht ändern.
func TestEineHausregelGiltNurWoDasProfilSieErlaubt(t *testing.T) {
	t.Parallel()

	profil := domain.Grundprofil("magic")

	// **Der Sandkasten erlaubt alle sechs.**
	for _, art := range domain.Hausregelarten() {
		regel := domain.Hausregel{Art: art}
		if art == domain.HausregelStartleben {
			regel.Wert = "25"
		}

		if fehler := profil.ErlaubtHausregel(regel); fehler != nil {
			t.Errorf("%q wird abgewiesen: %v", art, fehler)
		}
	}

	// **Ein engeres Profil erlaubt weniger.**
	eng := profil
	eng.ErlaubteHausregeln = []domain.Hausregelart{domain.HausregelProxy}

	fehler := eng.ErlaubtHausregel(domain.Hausregel{
		Art: domain.HausregelOffeneHaende,
	})
	if fehler == nil {
		t.Fatal("offene Hände kommen durch")
	}

	satz := fehler.Error()
	if !strings.Contains(satz, "Hausregel") ||
		!strings.Contains(satz, "Regelwerk des Spiels") {
		t.Errorf("%q nennt nicht beide Stufen", satz)
	}

	// **Und etwas, das gar keine Hausregel ist, erst recht nicht.**
	if fehler := profil.ErlaubtHausregel(domain.Hausregel{Art: "betrug"}); fehler == nil {
		t.Error("eine erfundene Hausregel kommt durch")
	}

	// **Ein Startleben ohne Zahl ist keine Abweichung.**
	if fehler := profil.ErlaubtHausregel(domain.Hausregel{
		Art: domain.HausregelStartleben,
	}); fehler == nil {
		t.Error("ein Startleben ohne Zahl kommt durch")
	}
}

// **Die Spielerzahl trägt, was sie sagt.**
func TestDieSpielerzahlTraegtWasSieSagt(t *testing.T) {
	t.Parallel()

	profil := domain.Grundprofil("magic")

	for anzahl, erwartet := range map[int]bool{
		0: false, 1: true, 2: true, 4: true, 5: false,
	} {
		if profil.TraegtSpieler(anzahl) != erwartet {
			t.Errorf("%d Spieler: %v, erwartet %v",
				anzahl, profil.TraegtSpieler(anzahl), erwartet)
		}
	}
}

// **Die Hausregelarten trennen.**
func TestDieHausregelartenTrennen(t *testing.T) {
	t.Parallel()

	if len(domain.Hausregelarten()) != 6 {
		t.Errorf("%d Arten", len(domain.Hausregelarten()))
	}

	for _, art := range domain.Hausregelarten() {
		if !domain.IstHausregelart(art) {
			t.Errorf("%q gilt nicht", art)
		}
	}

	if domain.IstHausregelart("betrug") {
		t.Error("eine erfundene Art gilt")
	}
}

// **Und die Zonenprüfung trennt ebenso.**
func TestIstZoneTrennt(t *testing.T) {
	t.Parallel()

	for _, zone := range domain.Zonen() {
		if !domain.IstZone(zone) {
			t.Errorf("%q gilt nicht", zone)
		}
	}

	if domain.IstZone("keller") {
		t.Error("eine erfundene Zone gilt")
	}
}

// **Eine Partie trägt ihr Profil, und der Griff zeigt darauf.**
//
// # Die Zusage, für die es diesen Lauf gibt
//
// **Ein Griff, der auf ein anderes Profil zeigt als das eingefrorene, wäre
// eine Auskunft, die zweimal verschieden ausfällt** — das Fenster zeigte die
// eine, die Entscheidung fiele nach der anderen. Genau so entstehen Fehler,
// die niemand findet: beide Seiten sind für sich gültig.
func TestEinePartieTraegtIhrProfilUndDerGriffZeigtDarauf(t *testing.T) {
	t.Parallel()

	gut := einePartie()

	if fehler := gut.Pruefe(); fehler != nil {
		t.Fatalf("eine Partie mit Profil wird abgewiesen: %v", fehler)
	}

	faelle := map[string]func(*domain.Partie){
		"andere Kennung": func(p *domain.Partie) {
			p.Regelwerk.Kennung = "commander"
		},
		"andere Fassung": func(p *domain.Partie) {
			p.Regelwerk.Fassung = 99
		},
		"anderes Spiel": func(p *domain.Partie) {
			p.Profil.Spiel = "space"
		},
		"ohne Profil": func(p *domain.Partie) {
			p.Profil = domain.Regelprofil{}
		},

		// **Die Strenge prüft das Profil nicht** — sie gehört zur Partie und
		// nicht zum Regelwerk, das eingefroren wird. Ohne diesen Fall bliebe
		// `Regelwerk.Pruefe` unbetreten, und die Prüfung darüber sähe aus,
		// als deckte sie alles ab.
		"erfundene Strenge": func(p *domain.Partie) {
			p.Regelwerk.Strenge = "turnier"
		},
	}

	for name, verdirb := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			partie := einePartie()
			verdirb(&partie)

			fehler := partie.Pruefe()
			if fehler == nil {
				t.Fatal("kommt durch")
			}

			if !errors.Is(fehler, domain.ErrValidation) {
				t.Errorf("Fehler %v ist keine Validierung", fehler)
			}
		})
	}
}
