package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Schlüsselwort im Regeltext bekommt seinen Erinnerungstext** —
// #1061.
//
// # Wie MSE2 den Ausdruck baut
//
// `Keyword::prepare`, `src/data/keyword.cpp`, Zeilen 148–208, Fassung f9d9356
// vom 2021-07-06:
//
//   - `(?i)` — Groß- und Kleinschreibung spielt keine Rolle
//     (`USE_CASE_INSENSITIVE_KEYWORDS 1`, Zeile 20).
//   - Der Text zwischen den Parametern wird **escaped**.
//   - `<atom-param>art</atom-param>` wird zur Gruppe der Parameterart.
//   - Vorn `\y`, hinten `(?=$|[^a-zA-Z0-9\(])` — **nur ganze Wörter**.
func TestEinSchluesselwortBekommtSeinenErinnerungstext(t *testing.T) {
	t.Parallel()

	liste, befunde := service.NeueWortliste(
		[]msepaket.Schluesselwort{{
			Wort: "Fliegend", Muster: "Fliegend", Modus: "core",
			Erinnerung: "Kann nur von fliegenden Kreaturen geblockt werden.",
		}}, nil)

	if len(befunde) != 0 {
		t.Fatalf("Befunde beim Vorbereiten: %v", befunde)
	}

	gesetzt := liste.Einsetzen("Fliegend, Eile.", klammern)

	const erwartet = "Fliegend (Kann nur von fliegenden Kreaturen geblockt werden.), Eile."
	if gesetzt != erwartet {
		t.Errorf("gesetzt %q, erwartet %q", gesetzt, erwartet)
	}
}

// **Nur ganze Wörter.**
//
// `Fliegend` darf in `Fliegendes` nicht treffen — sonst stünde mitten im Wort
// ein Erinnerungstext.
func TestEinSchluesselwortTrifftNurGanzeWoerter(t *testing.T) {
	t.Parallel()

	liste, _ := service.NeueWortliste([]msepaket.Schluesselwort{{
		Wort: "Fliegend", Muster: "Fliegend", Erinnerung: "erinnert",
	}}, nil)

	for _, fall := range []struct{ hinein, heraus string }{
		{"Fliegendes Ding", "Fliegendes Ding"},
		{"Nichtfliegend", "Nichtfliegend"},
		{"Fliegend", "Fliegend (erinnert)"},
		{"Fliegend.", "Fliegend (erinnert)."},
		{"fliegend", "fliegend (erinnert)"},
	} {
		if gesetzt := liste.Einsetzen(fall.hinein, klammern); gesetzt != fall.heraus {
			t.Errorf("%q wurde zu %q, erwartet %q", fall.hinein, gesetzt, fall.heraus)
		}
	}
}

// **Ein Parameter wird in den Erinnerungstext eingesetzt.**
//
// Der Erinnerungstext ist ein Skript und liest `param1`; hier wird geprüft,
// dass der **Wert** dort ankommt.
func TestEinParameterKommtImErinnerungstextAn(t *testing.T) {
	t.Parallel()

	liste, befunde := service.NeueWortliste(
		[]msepaket.Schluesselwort{{
			Wort: "Schutz vor", Muster: "Schutz vor <atom-param>name</atom-param>",
			Erinnerung: "schützt vor {param1}",
		}},
		[]msepaket.Schluesselwortparameterart{{Name: "name", Muster: "[A-Za-zÄÖÜäöü]+"}})

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	gesetzt := liste.Einsetzen("Schutz vor Rot.", klammern)

	if !strings.Contains(gesetzt, "Rot") {
		t.Errorf("gesetzt %q — der Parameter fehlt", gesetzt)
	}
}

// **Ein Schlüsselwort, dessen Ausdruck auf den leeren Text passt, wird
// abgeschaltet.**
//
//	valid = !match_re.matches("");
//	// not valid if it matches "", that would make MSE hang
//
// `src/data/keyword.cpp`, Zeilen 206–208, Fassung f9d9356 vom 2021-07-06.
func TestEinSchluesselwortDasAufLeerPasstWirdAbgeschaltet(t *testing.T) {
	t.Parallel()

	liste, befunde := service.NeueWortliste(
		[]msepaket.Schluesselwort{{
			Wort: "Leer", Muster: "<atom-param>wahlfrei</atom-param>",
			Erinnerung: "nie",
		}},
		[]msepaket.Schluesselwortparameterart{
			{Name: "wahlfrei", Muster: "[0-9]*", Wahlfrei: true},
		})

	if liste.Anzahl() != 0 {
		t.Errorf("%d Schlüsselwörter — das leere wurde nicht abgeschaltet", liste.Anzahl())
	}

	if len(befunde) != 1 {
		t.Fatalf("%d Befunde, erwartet einer: %v", len(befunde), befunde)
	}
}

// **Eine unbekannte Parameterart schaltet ihr Schlüsselwort ab, statt zu
// scheitern.**
//
//	// throwing an error can mean a set will not be loaded!
//	// instead, simply disable the keyword
//	handle_error(_("Unknown keyword parameter type: ") + type);
//	valid = false;
//
// `src/data/keyword.cpp`, Zeilen 174–181, Fassung f9d9356 vom 2021-07-06.
func TestEineUnbekannteParameterartSchaltetAb(t *testing.T) {
	t.Parallel()

	liste, befunde := service.NeueWortliste([]msepaket.Schluesselwort{{
		Wort: "Mit", Muster: "Mit <atom-param>gibtsnicht</atom-param>",
	}}, nil)

	if liste.Anzahl() != 0 {
		t.Error("das Schlüsselwort mit unbekannter Parameterart blieb stehen")
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "gibtsnicht") {
		t.Errorf("Befunde %v", befunde)
	}
}

// **Überlappende Treffer werden nicht zweimal ersetzt.**
//
// Die Reihenfolge steht in `sort_keyword_matches` (`src/data/keyword.cpp`,
// Zeilen 445–458, Fassung f9d9356 vom 2021-07-06): erst nach Stelle, dann **eigene vor festen**, dann der
// **längste** Treffer, dann nach Namen. `expand_keywords` geht sie in dieser
// Reihenfolge durch und überspringt jeden, der in einen schon gesetzten
// hineinragt.
func TestUeberlappendeTrefferWerdenNichtZweimalGesetzt(t *testing.T) {
	t.Parallel()

	liste, _ := service.NeueWortliste([]msepaket.Schluesselwort{
		{Wort: "Erste", Muster: "Erste Hilfe", Erinnerung: "lang", Fest: true},
		{Wort: "Hilfe", Muster: "Hilfe", Erinnerung: "kurz", Fest: true},
	}, nil)

	gesetzt := liste.Einsetzen("Erste Hilfe", klammern)

	// **Der längere gewinnt**, und der kürzere liegt darin — er kommt nicht
	// noch einmal.
	if gesetzt != "Erste Hilfe (lang)" {
		t.Errorf("gesetzt %q", gesetzt)
	}
}

// **Ein eigenes Schlüsselwort schlägt ein festes an derselben Stelle.**
func TestEinEigenesSchlaegtEinFestes(t *testing.T) {
	t.Parallel()

	liste, _ := service.NeueWortliste([]msepaket.Schluesselwort{
		{Wort: "Fest", Muster: "Eile", Erinnerung: "aus dem Spiel", Fest: true},
		{Wort: "Eigen", Muster: "Eile", Erinnerung: "aus dem Set", Fest: false},
	}, nil)

	if gesetzt := liste.Einsetzen("Eile", klammern); gesetzt != "Eile (aus dem Set)" {
		t.Errorf("gesetzt %q", gesetzt)
	}
}

// **Innerhalb von `<atom…>` wird nichts gesetzt.**
//
//	// don't expand keywords that are inside <atom> tags
//
// `src/data/keyword.cpp`, Zeile 526, Fassung f9d9356 vom 2021-07-06.
func TestInnerhalbEinesAtomsWirdNichtsGesetzt(t *testing.T) {
	t.Parallel()

	liste, _ := service.NeueWortliste([]msepaket.Schluesselwort{{
		Wort: "Eile", Muster: "Eile", Erinnerung: "erinnert",
	}}, nil)

	gesetzt := liste.Einsetzen("<atom-kwpph>Eile</atom-kwpph> und Eile", klammern)

	const erwartet = "<atom-kwpph>Eile</atom-kwpph> und Eile (erinnert)"
	if gesetzt != erwartet {
		t.Errorf("gesetzt %q, erwartet %q", gesetzt, erwartet)
	}
}

// klammern ist ein einfaches `combine:` für die Läufe hier.
func klammern(treffer service.Schluesselworttreffer) string {
	return treffer.Text + " (" + treffer.Erinnerung + ")"
}

// **Die Fehlerwege des Vorbereitens und des Einsetzens.**
//
// Sie stehen zusammen, weil sie dieselbe Frage beantworten: was geschieht mit
// einer Vorlage, die nicht stimmt? Die Antwort ist überall dieselbe — ein
// Befund, kein Abbruch (§8).
func TestDieFehlerwegeDerSchluesselwoerter(t *testing.T) {
	t.Parallel()

	t.Run("ein Muster ohne schließende Marke", func(t *testing.T) {
		t.Parallel()

		liste, _ := service.NeueWortliste([]msepaket.Schluesselwort{{
			Wort: "Halb", Muster: "Halb <atom-param>name", Erinnerung: "x",
		}}, []msepaket.Schluesselwortparameterart{{Name: "name", Muster: "[a-z]+"}})

		// **Die halbe Marke bleibt Text.** MSE bricht dort ebenfalls nicht ab;
		// `skip_tag` liest bis zum Ende und macht weiter.
		if liste.Anzahl() != 1 {
			t.Errorf("%d Schlüsselwörter, erwartet eines", liste.Anzahl())
		}
	})

	t.Run("eine Parameterart mit unübersetzbarem Muster", func(t *testing.T) {
		t.Parallel()

		liste, befunde := service.NeueWortliste([]msepaket.Schluesselwort{{
			Wort: "Krumm", Muster: "Krumm <atom-param>krumm</atom-param>",
		}}, []msepaket.Schluesselwortparameterart{{Name: "krumm", Muster: "[a-"}})

		if liste.Anzahl() != 0 || len(befunde) != 1 {
			t.Errorf("%d Schlüsselwörter, %d Befunde", liste.Anzahl(), len(befunde))
		}
	})

	t.Run("eine leere Wortliste lässt den Text stehen", func(t *testing.T) {
		t.Parallel()

		var leer service.Wortliste

		if gesetzt := leer.Einsetzen("Fliegend", klammern); gesetzt != "Fliegend" {
			t.Errorf("gesetzt %q", gesetzt)
		}
	})

	t.Run("ein wahlfreier Parameter, der fehlt", func(t *testing.T) {
		t.Parallel()

		liste, _ := service.NeueWortliste([]msepaket.Schluesselwort{{
			Wort: "Zyklus", Muster: "Zyklus<atom-param>kosten</atom-param>",
			Erinnerung: "kostet {param1}",
		}}, []msepaket.Schluesselwortparameterart{
			{Name: "kosten", Muster: " [0-9]+", Wahlfrei: true},
		})

		// **Der Treffer steht, der Parameter ist leer.** MSE setzt dort einen
		// Platzhalter (`<atom-kwpph>`); den gibt es hier noch nicht, und das
		// steht am Code.
		//
		// Geprüft wird über ein eigenes `combine:`, nicht über [klammern] —
		// das setzt den Erinnerungstext roh ein und zeigte `{param1}`.
		gesetzt := liste.Einsetzen("Zyklus.", func(treffer service.Schluesselworttreffer) string {
			return treffer.Text + "[" + strings.Join(treffer.Parameter, ",") + "]"
		})

		if gesetzt != "Zyklus[]." {
			t.Errorf("gesetzt %q", gesetzt)
		}
	})
}

// **Die Reihenfolge der Treffer, in allen vier Stufen.**
//
// `sort_keyword_matches`, `src/data/keyword.cpp`, Zeilen 445–458, Fassung
// f9d9356 vom 2021-07-06: Stelle, dann eigene vor festen, dann der längste,
// dann — bei MSE — der Name. Hier entscheidet zuletzt die Stelle in der
// Eingabereihe; sie ist stabil und braucht den Namen nicht.
//
// **Die dritte und vierte Stufe brauchen einen eigenen Lauf**: sie werden nur
// erreicht, wenn Stelle **und** Festigkeit gleich sind.
func TestDieReihenfolgeDerTreffer(t *testing.T) {
	t.Parallel()

	t.Run("bei gleicher Stelle gewinnt der längere", func(t *testing.T) {
		t.Parallel()

		liste, _ := service.NeueWortliste([]msepaket.Schluesselwort{
			{Wort: "Kurz", Muster: "Erste", Erinnerung: "kurz"},
			{Wort: "Lang", Muster: "Erste Hilfe", Erinnerung: "lang"},
		}, nil)

		if gesetzt := liste.Einsetzen("Erste Hilfe", klammern); gesetzt != "Erste Hilfe (lang)" {
			t.Errorf("gesetzt %q", gesetzt)
		}
	})

	t.Run("bei gleicher Länge gewinnt der frühere Eintrag", func(t *testing.T) {
		t.Parallel()

		liste, _ := service.NeueWortliste([]msepaket.Schluesselwort{
			{Wort: "Erstes", Muster: "Eile", Erinnerung: "erstes"},
			{Wort: "Zweites", Muster: "Eile", Erinnerung: "zweites"},
		}, nil)

		if gesetzt := liste.Einsetzen("Eile", klammern); gesetzt != "Eile (erstes)" {
			t.Errorf("gesetzt %q", gesetzt)
		}
	})
}

// **Ob die Schreibung stimmt, steht am Treffer.**
//
//	ctx.setVariable("correct_case", to_script(correct_case));
//
// `src/data/keyword.cpp`, Zeile 697, Fassung f9d9356 vom 2021-07-06. Die
// M15-Vorlage entscheidet daran, ob ein Erinnerungstext überhaupt erscheint:
// `correct_case or (mode != "pseudo" and not used_placeholders)`.
//
// # Warum das sichtbar ist
//
// Ohne die Angabe stand auf einer echten Karte zweimal derselbe
// Erinnerungstext: einmal hinter „Flying" am Zeilenanfang und einmal mitten im
// Satz hinter „gains flying". Gesehen an `White.mse-set`, Karte 3.
func TestDieSchreibungStehtAmTreffer(t *testing.T) {
	t.Parallel()

	liste, _ := service.NeueWortliste([]msepaket.Schluesselwort{{
		Wort: "Fliegend", Muster: "Fliegend", Erinnerung: "erinnert",
	}}, nil)

	gesetzt := liste.Einsetzen("Fliegend und fliegend",
		func(treffer service.Schluesselworttreffer) string {
			if treffer.SchreibungStimmt {
				return treffer.Text + "!"
			}

			return treffer.Text + "?"
		})

	if gesetzt != "Fliegend! und fliegend?" {
		t.Errorf("gesetzt %q", gesetzt)
	}
}
