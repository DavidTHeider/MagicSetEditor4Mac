package domain_test

// **Was gilt, und woher es kommt** — Werkstück O3 von Phase 5.2
// (#1427).

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Auflösung von gestern gilt weiter — nur sagt sie jetzt, wem sie
// gehört.**
//
// Jeder dieser Fälle steht schon in `uebersteuerung_test.go`; hier steht er
// noch einmal mit der Herkunft daneben. **Das ist keine Doppelung, sondern die
// Zusage**: die Auflösung ändert sich nicht, es kommt eine Auskunft hinzu.
func TestDerStandNenntWertUndHerkunft(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		schluessel                  string
		profil, nutzer              map[string]string
		gewaehlt, wirksam, herkunft string
	}{
		"das Profil hat gewählt": {
			schluessel: domain.EinstellungErscheinungsbild,
			profil:     map[string]string{domain.EinstellungErscheinungsbild: "dunkel"},
			nutzer:     map[string]string{domain.EinstellungErscheinungsbild: "hell"},
			gewaehlt:   "dunkel", wirksam: "dunkel", herkunft: domain.GeltungProfil,
		},
		"das Profil erbt, der Mensch hat gewählt": {
			schluessel: domain.EinstellungErscheinungsbild,
			profil:     map[string]string{domain.EinstellungErscheinungsbild: domain.Erben},
			nutzer:     map[string]string{domain.EinstellungErscheinungsbild: "hell"},
			gewaehlt:   domain.Erben, wirksam: "hell", herkunft: domain.GeltungNutzer,
		},
		"niemand hat gewählt": {
			schluessel: domain.EinstellungErscheinungsbild,
			profil:     map[string]string{},
			nutzer:     map[string]string{},
			gewaehlt:   domain.Erben, wirksam: "system",
			herkunft: domain.GeltungVorgabe,
		},
		"das Profil steht auf der Vorgabe, ohne Zeile": {
			schluessel: domain.EinstellungAktualisierung,
			profil:     map[string]string{},
			nutzer:     map[string]string{},
			gewaehlt:   domain.AktualisierungAus,
			wirksam:    domain.AktualisierungAus, herkunft: domain.GeltungVorgabe,
		},
		"dieselbe Wahl, aber ausdrücklich getroffen": {
			schluessel: domain.EinstellungAktualisierung,
			profil: map[string]string{
				domain.EinstellungAktualisierung: domain.AktualisierungAus,
			},
			nutzer:   map[string]string{},
			gewaehlt: domain.AktualisierungAus,
			wirksam:  domain.AktualisierungAus, herkunft: domain.GeltungProfil,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			geltung := domain.WirksamerStand(fall.profil, fall.nutzer)[fall.schluessel]
			if geltung.Gewaehlt != fall.gewaehlt {
				t.Errorf("gewählt ist %q, erwartet war %q",
					geltung.Gewaehlt, fall.gewaehlt)
			}

			if geltung.Wirksam != fall.wirksam {
				t.Errorf("es gilt %q, erwartet war %q", geltung.Wirksam, fall.wirksam)
			}

			if geltung.Herkunft != fall.herkunft {
				t.Errorf("die Herkunft ist %q, erwartet war %q",
					geltung.Herkunft, fall.herkunft)
			}
		})
	}
}

// **Der Stand nennt jeden Schlüssel der Reihe** — auch die, die niemand
// gesetzt hat.
//
// Eine Antwort, die einen Schlüssel wegließe, zwänge jede Fläche, die Vorgabe
// noch einmal zu kennen — und die zweite Liste liefe zurück.
func TestDerStandIstVollstaendig(t *testing.T) {
	t.Parallel()

	stand := domain.WirksamerStand(map[string]string{}, map[string]string{})

	for _, schluessel := range domain.Einstellungsreihe() {
		if _, da := stand[schluessel]; !da {
			t.Errorf("%q fehlt im Stand", schluessel)
		}
	}

	if len(stand) != len(domain.Einstellungsreihe()) {
		t.Errorf("der Stand hat %d Zeilen, die Reihe %d",
			len(stand), len(domain.Einstellungsreihe()))
	}
}

// **Eine Schranke des Menschen sticht die Wahl des Profils.**
func TestDieSchrankeDesMenschenGewinnt(t *testing.T) {
	t.Parallel()

	schalter := domain.DienstEinstellung(domain.DienstScryfall)

	stand := domain.WirksamerStand(
		map[string]string{schalter: domain.DienstAn},
		map[string]string{schalter: domain.DienstAus})

	if stand[schalter].Wirksam != domain.DienstAus {
		t.Errorf("es gilt %q — das Profil hat eine Schranke aufgehoben",
			stand[schalter].Wirksam)
	}

	if stand[schalter].Herkunft != domain.GeltungNutzer {
		t.Errorf("die Herkunft ist %q, erwartet war %q",
			stand[schalter].Herkunft, domain.GeltungNutzer)
	}

	// **Und die Wahl bleibt sichtbar**: das Profil hat „an" gesagt, und die
	// Auswahlliste zeigt das. Wer nur den wirksamen Wert hätte, sähe eine
	// Wahl, die niemand getroffen hat.
	if stand[schalter].Gewaehlt != domain.DienstAn {
		t.Errorf("gewählt ist %q, erwartet war %q",
			stand[schalter].Gewaehlt, domain.DienstAn)
	}
}

// **„Nur lokal" schaltet jeden Dienst ab — und sagt, wer es war.**
func TestNurLokalSchaltetJedenDienstAb(t *testing.T) {
	t.Parallel()

	fuerJedenDienst := map[string]string{}
	for _, dienst := range domain.Dienste() {
		fuerJedenDienst[domain.DienstEinstellung(dienst)] = domain.DienstAn
	}

	stand := domain.WirksamerStand(fuerJedenDienst,
		map[string]string{domain.RichtlinieNurLokal: domain.NurLokalAn})

	for _, dienst := range domain.Dienste() {
		schluessel := domain.DienstEinstellung(dienst)

		if stand[schluessel].Wirksam != domain.DienstAus {
			t.Errorf("%q steht auf %q, obwohl nur lokal gearbeitet wird",
				schluessel, stand[schluessel].Wirksam)
		}

		if stand[schluessel].Herkunft != domain.GeltungNutzer {
			t.Errorf("%q nennt die Herkunft %q — das Nein kam vom Menschen",
				schluessel, stand[schluessel].Herkunft)
		}
	}
}

// **Ein selbst abgeschalteter Dienst behält seine Herkunft.**
//
// Wer ihn im Profil ausgeschaltet hat, soll das sehen; die Schranke darüber
// ändert an seiner Entscheidung nichts.
func TestEinAbgeschalteterDienstBehaeltSeineHerkunft(t *testing.T) {
	t.Parallel()

	schalter := domain.DienstEinstellung(domain.DienstScryfall)

	stand := domain.WirksamerStand(
		map[string]string{
			schalter:                  domain.DienstAus,
			domain.RichtlinieNurLokal: domain.NurLokalAn,
		},
		map[string]string{})

	if stand[schalter].Herkunft != domain.GeltungProfil {
		t.Errorf("die Herkunft ist %q, erwartet war %q",
			stand[schalter].Herkunft, domain.GeltungProfil)
	}
}

// **Ohne „nur lokal" bleibt alles, wie es war.**
//
// Die Gegenprobe zum Lauf darüber: dieselbe Lage ohne die Schranke lässt jeden
// Dienst an. Ohne sie sagte jener Lauf nur, dass Dienste aus sind — nicht,
// dass die Schranke es war.
func TestOhneNurLokalBleibenDieDiensteAn(t *testing.T) {
	t.Parallel()

	schalter := domain.DienstEinstellung(domain.DienstScryfall)

	stand := domain.WirksamerStand(
		map[string]string{schalter: domain.DienstAn}, map[string]string{})

	if stand[schalter].Wirksam != domain.DienstAn {
		t.Errorf("es gilt %q, erwartet war %q",
			stand[schalter].Wirksam, domain.DienstAn)
	}
}

// **Ein Layoutwert erbt wie eine Schriftgröße.**
func TestDerArbeitsplatzErbtVomMenschen(t *testing.T) {
	t.Parallel()

	stand := domain.WirksamerStand(map[string]string{}, map[string]string{})

	faelle := map[string]string{
		domain.ArbeitsplatzNavigator:       domain.SpalteSichtbar,
		domain.ArbeitsplatzInspektor:       domain.SpalteSichtbar,
		domain.ArbeitsplatzNavigatorbreite: domain.BreiteNavigatorVorgabe,
		domain.ArbeitsplatzInspektorbreite: domain.BreiteInspektorVorgabe,
	}

	for schluessel, erwartet := range faelle {
		if stand[schluessel].Wirksam != erwartet {
			t.Errorf("%q gilt als %q, erwartet war %q",
				schluessel, stand[schluessel].Wirksam, erwartet)
		}
	}

	// **Und das Profil sticht.**
	eigen := domain.WirksamerStand(
		map[string]string{domain.ArbeitsplatzNavigatorbreite: "400"},
		map[string]string{})
	if eigen[domain.ArbeitsplatzNavigatorbreite].Wirksam != "400" {
		t.Errorf("die eigene Breite gilt nicht: %q",
			eigen[domain.ArbeitsplatzNavigatorbreite].Wirksam)
	}
}
