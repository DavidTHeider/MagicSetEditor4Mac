package domain_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Einstellungen der Anwendung** —
// #245.
//
// „Eine Einstellung, die es nicht gibt, ist eine Entscheidung, die jemand
// anderes getroffen hat." Bis hierher führte die Seite die Sprache und sonst
// nichts — und die lag im Browserspeicher.
//
// **Warum die Werte hier stehen und nicht im Fenster:** ein Wert, den nur die
// Oberfläche kennt, kommt ungeprüft in den Bestand. Der MCP-Server schreibt
// auf dieselbe Datei; was er hineinschreibt, muss dieselben Regeln bestehen.
func TestJedeEinstellungKenntIhreWerte(t *testing.T) {
	t.Parallel()

	// **`erben` steht vorn, seit Werkstück 9** (#1351, Punkt 9): ein Profil
	// nimmt, was beim Menschen steht, solange es nicht selbst gewählt hat —
	// und die Vorgabe steht in dieser Anwendung oben (§6).
	faelle := map[string][]string{
		// **Acht Sprachen seit LOC-01** (#1687) — Deutsch als Ausgangssprache
		// vorn, die übrigen nach ihrem Kürzel.
		domain.EinstellungSprache: {
			domain.Erben, "de", "en", "es", "fr", "it", "ja", "ru", "zh",
		},
		domain.EinstellungErscheinungsbild: {
			domain.Erben, "system", "hell", "dunkel",
		},
		domain.EinstellungSchriftgroesse: {
			domain.Erben, "klein", "normal", "gross", "sehrgross",
		},
	}

	for schluessel, erwartet := range faelle {
		t.Run(schluessel, func(t *testing.T) {
			t.Parallel()

			werte := domain.ErlaubteEinstellungswerte(schluessel)
			if len(werte) != len(erwartet) {
				t.Fatalf("%v", werte)
			}

			for i, wert := range erwartet {
				if werte[i] != wert {
					t.Errorf("an Stelle %d steht %q, erwartet %q", i, werte[i], wert)
				}
			}
		})
	}
}

// **Ein unbekannter Schlüssel ist ein Befund, kein stilles Nichts.**
//
// Der Bestand ist eine Tabelle aus Schlüssel und Wert; ohne diese Prüfung
// schriebe ein Tippfehler eine Zeile, die nie jemand liest — und die
// Einstellung, die gemeint war, bliebe stehen.
func TestEinUnbekannterSchluesselWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	err := domain.PruefeEinstellung("erfunden", "egal")
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("%v", err)
	}
}

func TestEinUnbekannterWertWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	err := domain.PruefeEinstellung(domain.EinstellungErscheinungsbild, "grau")
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("%v", err)
	}
}

func TestEinErlaubterWertKommtDurch(t *testing.T) {
	t.Parallel()

	if err := domain.PruefeEinstellung(domain.EinstellungErscheinungsbild, "dunkel"); err != nil {
		t.Errorf("%v", err)
	}
}

// **Jede Einstellung hat einen Vorgabewert, und er ist der zurückhaltendste.**
//
// **Für die vier Erscheinungseinstellungen ist das seit Werkstück 9 `erben`**
// (#1351, Punkt 9), und das ist derselbe Gedanke eine Ebene höher: ein
// frisches Profil soll nichts über den Menschen behaupten. Was ohne Wahl
// **gilt**, steht jetzt beim Menschen — `system`, `normal`, `pergament`, und
// bei der Sprache **nichts**, denn sie hängt an `navigator.language`, und das
// weiß nur das Fenster. Der Lauf darüber steht in `nutzereinstellung_test.go`.
//
// **Für alles andere gilt hier weiter, was hier stand**: `aktualisierung` und
// `bildkonflikt` erbt niemand.
func TestJedeEinstellungHatEinenVorgabewert(t *testing.T) {
	t.Parallel()

	faelle := map[string]string{
		domain.EinstellungSprache:          domain.Erben,
		domain.EinstellungErscheinungsbild: domain.Erben,
		domain.EinstellungSchriftgroesse:   domain.Erben,
		domain.EinstellungAktualisierung:   domain.AktualisierungAus,
	}

	for schluessel, erwartet := range faelle {
		if gefunden := domain.EinstellungsVorgabe(schluessel); gefunden != erwartet {
			t.Errorf("%s: %q, erwartet %q", schluessel, gefunden, erwartet)
		}
	}
}

// **Und ein unbekannter Schlüssel hat keinen.** Der Fall entsteht nicht beim
// Lesen des Bestands — dort steht nur, was die Prüfung durchgelassen hat —,
// aber er ist der Weg, auf dem ein Tippfehler still zu einem Wert würde.
func TestEinUnbekannterSchluesselHatKeinenVorgabewert(t *testing.T) {
	t.Parallel()

	if gefunden := domain.EinstellungsVorgabe("erfunden"); gefunden != "" {
		t.Errorf("%q", gefunden)
	}
}

func TestEinUnbekannterSchluesselHatKeineWerte(t *testing.T) {
	t.Parallel()

	if werte := domain.ErlaubteEinstellungswerte("erfunden"); len(werte) != 0 {
		t.Errorf("%v", werte)
	}
}

// **Die ersten Schritte sind eine Einstellung, keine Seitenhälfte** —
// #1325.
//
// # Was vorher war
//
// Die Startseite zeigte *Erste Schritte* **oder** *Dein Überblick*, je
// nachdem, ob der Bestand ein Set führt. Solange die Zahlen unterwegs waren,
// galt „kein Set" — und die Seite sprang beim Eintreffen der Antwort um. Der
// Kommentar darüber sagte ausdrücklich das Gegenteil: *„eine Startseite, die
// erst »Erste Schritte« zeigt und dann etwas anderes, wäre ein Sprung vor den
// Augen dessen, der gerade liest."* Ein Kopf, der mehr verspricht als sein
// Rumpf hält (§10).
//
// **Jetzt ist die Einführung ein Pop-up**, und ob es wiederkehrt, ist eine
// Einstellung — an zwei Stellen bedienbar: im Pop-up selbst und auf der
// Einstellungsseite.
func TestDieErstenSchritteSindEineEinstellung(t *testing.T) {
	t.Parallel()

	// **Vorgabe „an"**: wer die Anwendung zum ersten Mal öffnet, ist genau
	// der, für den es die Einführung gibt.
	if vorgabe := domain.EinstellungsVorgabe(domain.EinstellungErsteSchritte); vorgabe != domain.ErsteSchritteAn {
		t.Errorf("die Vorgabe ist %q, erwartet %q", vorgabe, domain.ErsteSchritteAn)
	}

	for _, wert := range []string{domain.ErsteSchritteAn, domain.ErsteSchritteAus} {
		if err := domain.PruefeEinstellung(domain.EinstellungErsteSchritte, wert); err != nil {
			t.Errorf("%q wird abgewiesen: %v", wert, err)
		}
	}

	if err := domain.PruefeEinstellung(domain.EinstellungErsteSchritte, "vielleicht"); err == nil {
		t.Error("»vielleicht« kommt durch")
	}

	// **Und sie steht bei der Anwendung**, nicht bei der Bildausfuhr: sie
	// betrifft das Fenster als Ganzes.
	if bereich := domain.EinstellungsBereich(domain.EinstellungErsteSchritte); bereich != domain.BereichAnwendung {
		t.Errorf("der Bereich ist %q, erwartet %q", bereich, domain.BereichAnwendung)
	}
}

// **Die Anzeigeschrift ist wählbar** —
// #1325,
// Werkstück J.
//
// # Warum vier und nicht beliebig viele
//
// **Weil jede von ihnen auf dem Zielrechner liegen muss** (§16, Regel 5: die
// Anwendung arbeitet ohne Netz). Nachgesehen am 2026-09-03 auf macOS 26:
// `Hoefler Text`, `Luminari`, `Herculanum` und `Trattatello` sind alle vier
// vorhanden. Eine fünfte, die jemand hübsch fände, wäre eine, die bei
// manchen fehlt — und dann entscheidet die Rückfallkette, wie die Anwendung
// aussieht.
//
// # Und nur die Anzeigeschrift
//
// Der Fließtext bleibt die Systemschrift. Eine Kalligraphie über einem
// Absatz ist keine Gestaltung, sondern eine Zumutung; §6 verlangt
// Lesbarkeit, und die Wahl gilt deshalb den Überschriften und dem Setnamen.
func TestDieAnzeigeschriftIstWaehlbar(t *testing.T) {
	t.Parallel()

	// **Die Vorgabe steht seit Werkstück 9 beim Menschen** (#1351, Punkt 9):
	// das Profil erbt sie. Gewählt bleibt sie an beiden Stellen.
	if vorgabe := domain.NutzerVorgabe(
		domain.EinstellungSchriftart); vorgabe != domain.SchriftartPergament {
		t.Errorf("die Vorgabe ist %q, erwartet %q", vorgabe, domain.SchriftartPergament)
	}

	for _, wert := range []string{
		domain.SchriftartPergament, domain.SchriftartFantasy,
		domain.SchriftartAntike, domain.SchriftartKalligrafie,
	} {
		if err := domain.PruefeEinstellung(domain.EinstellungSchriftart, wert); err != nil {
			t.Errorf("%q wird abgewiesen: %v", wert, err)
		}
	}

	if err := domain.PruefeEinstellung(domain.EinstellungSchriftart, "comic"); err == nil {
		t.Error("»comic« kommt durch")
	}

	if bereich := domain.EinstellungsBereich(domain.EinstellungSchriftart); bereich != domain.BereichAnwendung {
		t.Errorf("der Bereich ist %q, erwartet %q", bereich, domain.BereichAnwendung)
	}
}
