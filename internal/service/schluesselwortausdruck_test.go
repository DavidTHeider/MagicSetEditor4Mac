package service //nolint:testpackage // ausdruckFuer ist paketintern

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// TestDerTrennerVorEinemParameterWirdGefressen ist Befund 1 des Zweitblicks
// über Phase W, zweite Hälfte.
//
//	// remove the separator from the text to prevent duplicates
//	param->eat_separator_before(text);
//	regex += _("(") + regex_escape(text) + _(")");
//
// `src/data/keyword.cpp`, Zeilen 185–187, Fassung f9d9356 vom 2021-07-06.
// Der gesammelte feste Text wird **hinten gekürzt**, wenn er auf den Trenner
// endet — sonst steht der Trenner zweimal da: einmal als Text, einmal als
// Anfang des Parametermusters.
//
// **Gemessen am 2026-08-24** an `M15-Magic-Pack/data/magic.mse-game/keywords`:
// von 15 Parameterarten tragen **drei** einen Trenner (`cost`, `action`,
// `energy`), und **keine einzige** ein `eat separator:` — die Vorgabe `true`
// entscheidet also in allen dreien.
func TestDerTrennerVorEinemParameterWirdGefressen(t *testing.T) {
	t.Parallel()

	// **Die Angaben sind die echten**, aus
	// `M15-Magic-Pack/data/magic.mse-game/keywords`, abgerufen am 2026-08-24 —
	// samt des führenden `[ ]` im Muster, das den Trenner ein zweites Mal
	// verlangt.
	arten := map[string]msepaket.Schluesselwortparameterart{"cost": {
		Name:          "cost",
		Muster:        `[ ][HSVECTQXYZI0-9WUBRG/|]*|[-—][^(\r\n]*`,
		TrennerDavor:  "[ —-]",
		TrennerDanach: "[.]",
		// **Gesetzt, weil der Leser es setzt.** Im Paket steht kein
		// `eat separator:`; die Vorgabe `true` kommt aus
		// [msepaket.Paket.Schluesselwortparameterarten], nicht aus dem
		// Nullwert des Verbunds.
		TrennerFressen: true,
	}}

	// Und das Schlüsselwort ist eines aus derselben Datei, Zeile 228.
	ausdruck, befund := ausdruckFuer(msepaket.Schluesselwort{
		Wort:   "Cumulative upkeep",
		Muster: "Cumulative upkeep <atom-param>cost</atom-param>",
	}, arten)
	if befund != nil {
		t.Fatalf("Befund: %v", befund)
	}

	// **Ein einziges Leerzeichen.** Ohne das Fressen stünde eines im festen
	// Text und eines am Anfang des Kostenmusters — und kein Kartentext der
	// Welt trüge beide. Der Erinnerungstext blieb aus.
	if !ausdruck.MatchString("Cumulative upkeep 2") {
		t.Errorf("der gewöhnliche Kartentext passt nicht auf %s", ausdruck)
	}

	// **Und der zweite echte Fall aus derselben Datei**, Zeile 1977: dort
	// endet der feste Text auf einen Gedankenstrich mit Leerzeichen.
	//
	// **`action` gehört dazu.** Der erste Anlauf reichte nur `cost` herein —
	// `ausdruckFuer` gab dann einen Befund, und die Zusicherung dahinter stand
	// hinter `befund == nil` und wurde **still übersprungen**. Angestrichen vom
	// Nachlauf des Zweitblicks über Phase W: ein Test, der seine eigene
	// Vorrichtung nicht vollständig baut, ist grün und prüft nichts.
	arten["action"] = msepaket.Schluesselwortparameterart{
		Name: "action", Muster: `[^(:\n]+`, TrennerDanach: "[.]",
		Wahlfrei: true, TrennerFressen: true,
	}

	zweiter, befund := ausdruckFuer(msepaket.Schluesselwort{
		Wort:   "Exhaust",
		Muster: "Exhaust — <atom-param>cost</atom-param>: <atom-param>action</atom-param>",
	}, arten)
	if befund != nil {
		t.Fatalf("Exhaust: %v", befund)
	}

	if !zweiter.MatchString("Exhaust — 2: Do something") {
		t.Errorf("Exhaust passt nicht auf %s", zweiter)
	}
}

// TestDerTrennerNachEinemParameterWirdUebersprungen ist die andere Hälfte von
// `eat_separator`.
//
//	separator_after_eat.assign(_("^") + separator_after_is);
//	…
//	i += result.length();
//
// `src/data/keyword.cpp`, Zeilen 107 und 126, Fassung f9d9356 vom 2021-07-06.
//
// **Der Fall ist gebaut, nicht gefunden.** In `M15-Magic-Pack` tragen drei
// Arten ein `separator after is`, aber in keinem der Schlüsselwörter folgt
// dem Parameter wirklich der Trenner — nach `<atom-param>cost</atom-param>`
// steht dort `:` oder `,`. Die **Regel** ist nachgelesen, der **Anlass**
// konstruiert; ohne diesen Satz sähe der Test aus wie eine Messung.
func TestDerTrennerNachEinemParameterWirdUebersprungen(t *testing.T) {
	t.Parallel()

	arten := map[string]msepaket.Schluesselwortparameterart{"action": {
		Name:           "action",
		Muster:         `[^(:\n]+`,
		TrennerDanach:  "[.]",
		Wahlfrei:       true,
		TrennerFressen: true,
	}}

	ausdruck, befund := ausdruckFuer(msepaket.Schluesselwort{
		Wort:   "Tue",
		Muster: "Tue <atom-param>action</atom-param>. Danach",
	}, arten)
	if befund != nil {
		t.Fatalf("Befund: %v", befund)
	}

	if got := ausdruck.String(); strings.Contains(got, `\.`) {
		t.Errorf("der Punkt steht noch im festen Text: %s", got)
	}
}

// TestEineParameterartIstOhneAngabeWahlfrei ist Befund 1, erste Hälfte.
//
//	KeywordParam::KeywordParam()
//	  : optional(true)
//	  , eat_separator(true)
//
// `src/data/keyword.cpp`, Zeilen 24–27, Fassung f9d9356 vom 2021-07-06.
//
// Der Kommentar an [msepaket.Schluesselwortparameterart] nannte **genau diese
// Zeilen** und behauptete das Gegenteil. Von 15 Parameterarten im echten Pack
// tragen **13** kein `optional:`; die Vorgabe entscheidet über fast alle.
func TestEineParameterartIstOhneAngabeWahlfrei(t *testing.T) {
	t.Parallel()

	baum, err := msepaket.Zerteile("keywords", []byte(
		"keyword parameter type:\n\tname: mana\n\tmatch: [0-9WUBRG]+\n"))
	if err != nil {
		t.Fatalf("zerteilen: %v", err)
	}

	arten := (&msepaket.Paket{Kopf: baum}).Schluesselwortparameterarten()
	if len(arten) != 1 {
		t.Fatalf("%d Arten", len(arten))
	}

	if !arten[0].Wahlfrei {
		t.Error("eine Parameterart ohne `optional:` ist nicht wahlfrei — MSE setzt optional(true)")
	}
}

// TestEinUnbrauchbarerTrennerLaesstDenTextStehen betritt die Fehlerwege beider
// Trennerhälften.
//
// **Ein Trennerausdruck kommt aus einem fremden Paket** (§8). Ihn zum Fehler zu
// machen nähme das ganze Schlüsselwort mit, obwohl der Rest lesbar ist — und
// MSE tut dasselbe: `Regex::assign` meldet den Fehler und lässt den Ausdruck
// leer, `eat_separator_before` kehrt bei leerem Ausdruck sofort zurück
// (`src/data/keyword.cpp`, Zeile 112, Fassung f9d9356 vom 2021-07-06).
func TestEinUnbrauchbarerTrennerLaesstDenTextStehen(t *testing.T) {
	t.Parallel()

	for name, art := range map[string]msepaket.Schluesselwortparameterart{
		"davor":  {Name: "x", Muster: "[0-9]+", TrennerDavor: "[", TrennerFressen: true},
		"danach": {Name: "x", Muster: "[0-9]+", TrennerDanach: "[", TrennerFressen: true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ausdruck, befund := ausdruckFuer(msepaket.Schluesselwort{
				Wort:   "Wort",
				Muster: "Wort <atom-param>x</atom-param>. Ende",
			}, map[string]msepaket.Schluesselwortparameterart{"x": art})
			if befund != nil {
				t.Fatalf("Befund: %v", befund)
			}

			if !ausdruck.MatchString("Wort 2. Ende") {
				t.Errorf("der Text passt nicht auf %s", ausdruck)
			}
		})
	}
}

// TestOhneTrennerBleibtAllesWieEsWar betritt die frühen Rückgaben.
func TestOhneTrennerBleibtAllesWieEsWar(t *testing.T) {
	t.Parallel()

	for name, art := range map[string]msepaket.Schluesselwortparameterart{
		"kein Trenner angegeben": {Name: "x", Muster: "[0-9]+", TrennerFressen: true},
		"Trenner, aber nicht zu fressen": {
			Name: "x", Muster: "[0-9]+", TrennerDavor: "[ ]", TrennerDanach: "[.]",
		},
		"Trenner passt nicht": {
			Name: "x", Muster: "[0-9]+", TrennerDavor: "[xy]", TrennerDanach: "[xy]",
			TrennerFressen: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ausdruck, befund := ausdruckFuer(msepaket.Schluesselwort{
				Wort:   "Wort",
				Muster: "Wort <atom-param>x</atom-param>. Ende",
			}, map[string]msepaket.Schluesselwortparameterart{"x": art})
			if befund != nil {
				t.Fatalf("Befund: %v", befund)
			}

			if !ausdruck.MatchString("Wort 2. Ende") {
				t.Errorf("der Text passt nicht auf %s", ausdruck)
			}
		})
	}
}
