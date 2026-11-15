package domain_test

// **Das Austauschformat, von außen gelesen** — Werkstück O3 von Phase 5.2
// (#1427).
//
// # Warum hier so viele Fehlerfälle stehen
//
// **Weil eine Einfuhrdatei von irgendwoher kommt.** Sie ist der einzige Weg,
// auf dem fremde Bytes in die Einstellungen dieses Programms gelangen; §8 sagt
// dazu, dass fremde Daten der Normalfall sind und nicht der Ausnahmefall. Ein
// Lauf, der nur den geglückten Weg prüft, hat die Hälfte ausgelassen (§10).

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// gutesPaket gibt einen Austausch, wie ihn diese Anwendung schreibt.
func gutesPaket(t *testing.T) []byte {
	t.Helper()

	roh, fehler := json.Marshal(domain.Austausch{
		Fassung: domain.AustauschFassung,
		Bereiche: map[string]map[string]string{
			string(domain.EbeneNutzer): {
				domain.NutzerBewegung: domain.BewegungReduziert,
			},
			string(domain.EbeneProfil): {
				domain.EinstellungSprache: "en",
			},
		},
	})
	if fehler != nil {
		t.Fatalf("das Beispiel packen: %v", fehler)
	}

	return roh
}

func TestEinGutesPaketWirdGelesen(t *testing.T) {
	t.Parallel()

	austausch, befunde, fehler := domain.Zerlege(gutesPaket(t))
	if fehler != nil {
		t.Fatalf("ein eigenes Paket wird abgewiesen: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("es gab Befunde: %+v", befunde)
	}

	if austausch.Fassung != domain.AustauschFassung {
		t.Errorf("die Fassung ist %d", austausch.Fassung)
	}
}

// **Eine neuere Fassung wird verständlich abgewiesen.**
//
// **Verständlich heißt: die Meldung nennt beide Zahlen.** „Das Format wird
// nicht unterstützt" sagt niemandem, was zu tun ist; „geschrieben in Fassung 2,
// gelesen wird bis 1" sagt es.
func TestEineNeuereFassungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	roh := []byte(`{"fassung":2,"bereiche":{}}`)

	_, _, fehler := domain.Zerlege(roh)
	if fehler == nil {
		t.Fatal("eine Fassung aus der Zukunft kam durch")
	}

	for _, teil := range []string{"2", "1"} {
		if !strings.Contains(fehler.Error(), teil) {
			t.Errorf("die Meldung nennt %q nicht: %v", teil, fehler)
		}
	}
}

func TestWasBeimLesenAbgewiesenWird(t *testing.T) {
	t.Parallel()

	faelle := map[string][]byte{
		"kein JSON": []byte("{"),

		// **Ein Bruch schon beim Zählen der Stufen.** `tiefeVon` liest die
		// Marken, bevor irgendetwas gebaut wird — eine Datei, die nicht
		// gelesen werden soll, wird nicht gelesen.
		"eine Marke, die es nicht gibt": []byte(`{"fassung": }`),

		// **Und eine Fassung, die keine Zahl ist.**
		"Fassung als Wort":    []byte(`{"fassung":"eins","bereiche":{}}`),
		"kein Objekt":         []byte(`["fassung"]`),
		"ohne Fassung":        []byte(`{"bereiche":{}}`),
		"Fassung null":        []byte(`{"fassung":0,"bereiche":{}}`),
		"Fassung negativ":     []byte(`{"fassung":-1,"bereiche":{}}`),
		"zu tief":             []byte(`{"fassung":1,"bereiche":{"profil":{"sprache":{"a":1}}}}`),
		"Bereich kein Objekt": []byte(`{"fassung":1,"bereiche":{"profil":"de"}}`),
		"Wert kein Text":      []byte(`{"fassung":1,"bereiche":{"profil":{"sprache":3}}}`),
	}

	for name, roh := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, _, fehler := domain.Zerlege(roh); fehler == nil {
				t.Errorf("%s kam durch", name)
			}
		})
	}
}

// **Eine zu große Datei wird abgewiesen, bevor sie gelesen wird.**
func TestEineZuGrosseDateiWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	roh := []byte(`{"fassung":1,"bereiche":{},"fuellung":"` +
		strings.Repeat("x", domain.AustauschHoechstens) + `"}`)

	_, _, fehler := domain.Zerlege(roh)
	if fehler == nil {
		t.Fatal("eine Datei über der Grenze kam durch")
	}
}

// **Ein unbekanntes Feld wird gemeldet und nicht verschluckt** (§16).
func TestEinUnbekanntesFeldWirdZumBefund(t *testing.T) {
	t.Parallel()

	roh := []byte(`{"fassung":1,"bereiche":{},"absender":"jemand"}`)

	_, befunde, fehler := domain.Zerlege(roh)
	if fehler != nil {
		t.Fatalf("ein unbekanntes Feld ist kein Fehler: %v", fehler)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Schluessel, "absender") {
		t.Fatalf("es kam %+v", befunde)
	}
}

// **Jeder Schlüssel steht im Abschnitt seiner Ebene — sonst ist er ein
// Befund.**
func TestWasAusEinemAbschnittAngenommenWird(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		ebene, schluessel, wert string
		angenommen              bool
	}{
		"eine Barriere beim Menschen": {
			string(domain.EbeneNutzer), domain.NutzerBewegung,
			domain.BewegungReduziert, true,
		},
		"eine Übersteuerung im Profil": {
			string(domain.EbeneProfil), domain.EinstellungSprache, "en", true,
		},
		"eine Vorgabe des Profils": {
			string(domain.EbeneProfil), domain.VorgabeZustand, domain.OhneVorgabe, true,
		},
		"das Layout im eigenen Abschnitt": {
			string(domain.EbeneArbeitsplatz), domain.ArbeitsplatzNavigator,
			domain.SpalteVerborgen, true,
		},

		// **Und was nicht durchkommt.**
		"ein Pfad dieser Platte": {
			string(domain.EbeneGeraet), domain.VorgabeHochladepfad, "/tmp", false,
		},
		"eine Kennung im Profilabschnitt": {
			string(domain.EbeneProfil),
			domain.AnbieterKennung(domain.Anbieter()[0]), "Ov23", false,
		},
		"das Layout im Profilabschnitt": {
			string(domain.EbeneProfil), domain.ArbeitsplatzNavigator,
			domain.SpalteVerborgen, false,
		},
		"ein Profilschlüssel beim Menschen": {
			string(domain.EbeneNutzer), domain.VorgabeZustand, domain.OhneVorgabe, false,
		},
		"ein erfundener Abschnitt": {
			"woanders", domain.EinstellungSprache, "en", false,
		},
		"ein erfundener Schlüssel": {
			string(domain.EbeneProfil), "gibtesnicht", "egal", false,
		},
		"ein Wert, den es nicht gibt": {
			string(domain.EbeneProfil), domain.EinstellungSprache, "klingonisch", false,
		},
		"eine Breite außerhalb der Grenzen": {
			string(domain.EbeneArbeitsplatz), domain.ArbeitsplatzNavigatorbreite,
			"9000", false,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			paare, befunde := domain.PruefeAustausch(domain.Austausch{
				Fassung: domain.AustauschFassung,
				Bereiche: map[string]map[string]string{
					fall.ebene: {fall.schluessel: fall.wert},
				},
			})

			if fall.angenommen {
				if len(paare) != 1 || len(befunde) != 0 {
					t.Fatalf("angenommen: %+v, Befunde: %+v", paare, befunde)
				}

				return
			}

			if len(paare) != 0 || len(befunde) != 1 {
				t.Fatalf("angenommen: %+v, Befunde: %+v", paare, befunde)
			}

			if befunde[0].Grund == "" {
				t.Error("der Befund nennt keinen Grund")
			}
		})
	}
}

// **Die Reihenfolge der Zeilen ist die der Anzeige und nicht die einer Map.**
//
// **Sonst stünde ein Diff bei jedem Aufruf anders da** — Go durchläuft eine
// `map` absichtlich in wechselnder Reihenfolge, und eine Vorschau, die sich
// zwischen zwei Blicken umsortiert, ist keine.
func TestDieZeilenStehenInDerReiheDerAnzeige(t *testing.T) {
	t.Parallel()

	austausch := domain.Austausch{
		Fassung: domain.AustauschFassung,
		Bereiche: map[string]map[string]string{
			string(domain.EbeneProfil): {
				domain.EinstellungBildkonflikt: domain.BildkonfliktAltbehalten,
				domain.EinstellungSprache:      "en",
			},
			string(domain.EbeneNutzer): {
				domain.NutzerKontrast: domain.KontrastHoch,
			},
		},
	}

	erste, _ := domain.PruefeAustausch(austausch)

	erwartet := []domain.Einfuhrpaar{
		{Ebene: domain.EbeneNutzer, Schluessel: domain.NutzerKontrast, Wert: domain.KontrastHoch},
		{Ebene: domain.EbeneProfil, Schluessel: domain.EinstellungSprache, Wert: "en"},
		{
			Ebene: domain.EbeneProfil, Schluessel: domain.EinstellungBildkonflikt,
			Wert: domain.BildkonfliktAltbehalten,
		},
	}

	if len(erste) != len(erwartet) {
		t.Fatalf("es kamen %+v", erste)
	}

	for i, paar := range erwartet {
		if erste[i] != paar {
			t.Errorf("an Stelle %d steht %+v, erwartet %+v", i, erste[i], paar)
		}
	}
}

// **Was aus dem Profil mitreist und wohin.**
func TestWohinEinProfilwertReist(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		schluessel string
		abschnitt  domain.Ebene
		reist      bool
	}{
		"eine Übersteuerung":  {domain.EinstellungSprache, domain.EbeneProfil, true},
		"eine Vorgabe":        {domain.VorgabeZustand, domain.EbeneProfil, true},
		"ein Dienstschalter":  {domain.DienstEinstellung(domain.DienstScryfall), domain.EbeneProfil, true},
		"das Layout":          {domain.ArbeitsplatzNavigator, domain.EbeneArbeitsplatz, true},
		"ein Pfad":            {domain.VorgabeHochladepfad, "", false},
		"eine Kennung":        {domain.AnbieterKennung(domain.Anbieter()[0]), "", false},
		"die Aktualisierung":  {domain.EinstellungAktualisierung, "", false},
		"ein Nutzerschlüssel": {domain.NutzerBewegung, "", false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			abschnitt, reist := domain.AbschnittImProfil(fall.schluessel)
			if reist != fall.reist {
				t.Fatalf("reist: %v — erwartet war %v", reist, fall.reist)
			}

			if reist && abschnitt != fall.abschnitt {
				t.Errorf("reist im Abschnitt %q, erwartet war %q",
					abschnitt, fall.abschnitt)
			}
		})
	}
}

// **Und was vom Menschen mitreist.**
func TestWasBeimMenschenMitreist(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		schluessel string
		reist      bool
	}{
		"eine Barriere":  {domain.NutzerKontrast, true},
		"die Sprache":    {domain.EinstellungSprache, true},
		"der Faktor":     {domain.NutzerFaktorbeimWechsel, true},
		"das Layout":     {domain.ArbeitsplatzNavigator, false},
		"eine Erfindung": {"gibtesnicht", false},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if domain.ReistBeimNutzer(fall.schluessel) != fall.reist {
				t.Errorf("%q reist: %v — erwartet war %v", fall.schluessel,
					!fall.reist, fall.reist)
			}
		})
	}
}
