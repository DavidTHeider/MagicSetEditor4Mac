package domain_test

// **Der Beleg der Gegenseite** —
// #1650.
//
// # Wogegen hier geprüft wird
//
// **Gegen einen Beleg, der zu einem anderen Handel gehört.** Eine
// Unterschrift ist echt oder nicht — das sagt `peer.Pruefe`. Ob sie zu
// **diesem** Vorgang gehört, sagt sie nicht, und genau das ist die Frage,
// die hier beantwortet wird.
//
// **Und gegen einen Beleg, der nicht spiegelt.** Was hier abgeht, kommt
// drüben an. Ein Gegenbeleg, der dieselben Zeilen in derselben Richtung
// trägt, beschreibt nicht denselben Handel — er beschreibt einen, bei dem
// beide dasselbe weggeben.

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// einBelegpaar baut zwei Belege desselben Handels — meinen und seinen.
func einBelegpaar() (meiner, seiner domain.Handelsbeleg) {
	meiner = domain.Handelsbeleg{
		Vorgang: "v-1", Abdruck: "abdruck-3",
		Parteien: []string{"anna", "bert"},
		Zeilen: []domain.Journalzeile{
			{
				Art: domain.JournalAbgang, Printing: "p-1", Menge: 2,
				Zustand: domain.ZustandNearMint, Sprache: "de",
				Finish: domain.FinishNonfoil,
			},
			{
				Art: domain.JournalZugang, Printing: "p-2", Menge: 1,
				Zustand: domain.ZustandNearMint, Sprache: "en",
				Finish: domain.FinishFoil,
			},
		},
	}

	seiner = domain.Handelsbeleg{
		Vorgang: "v-1", Abdruck: "abdruck-3",
		Parteien: []string{"anna", "bert"},
		Zeilen: []domain.Journalzeile{
			{
				Art: domain.JournalZugang, Printing: "p-1", Menge: 2,
				Zustand: domain.ZustandNearMint, Sprache: "de",
				Finish: domain.FinishNonfoil,
			},
			{
				Art: domain.JournalAbgang, Printing: "p-2", Menge: 1,
				Zustand: domain.ZustandNearMint, Sprache: "en",
				Finish: domain.FinishFoil,
			},
		},
	}

	return meiner, seiner
}

// **Ein spiegelnder Beleg passt.**
func TestEinSpiegelnderBelegPasst(t *testing.T) {
	t.Parallel()

	meiner, seiner := einBelegpaar()

	if fehler := (domain.Belegurkunde{Beleg: seiner}).PasstZu(
		meiner); fehler != nil {
		t.Errorf("der spiegelnde Beleg wurde abgewiesen: %v", fehler)
	}
}

// **Und alles andere passt nicht — mit seinem Mangel** (#2500): das Fenster
// übersetzt das Wort, der Satz geht ins Protokoll.
func TestWasNichtSpiegeltWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name      string
		verbiegen func(*domain.Handelsbeleg)
		mangel    domain.Gegenbelegmangel
	}{
		{"ein anderer Vorgang", func(b *domain.Handelsbeleg) { b.Vorgang = "v-2" }, domain.MangelVorgang},
		{"ein anderer Abdruck", func(b *domain.Handelsbeleg) { b.Abdruck = "abdruck-4" }, domain.MangelRevision},
		{"andere Parteien", func(b *domain.Handelsbeleg) { b.Parteien = []string{"anna", "clara"} }, domain.MangelParteien},
		{"eine Partei zu wenig", func(b *domain.Handelsbeleg) { b.Parteien = []string{"anna"} }, domain.MangelParteien},
		{"eine Zeile zu wenig", func(b *domain.Handelsbeleg) { b.Zeilen = b.Zeilen[:1] }, domain.MangelZeilen},
		{"dieselbe Richtung", func(b *domain.Handelsbeleg) { b.Zeilen[0].Art = domain.JournalAbgang }, domain.MangelZeilen},
		{"eine andere Menge", func(b *domain.Handelsbeleg) { b.Zeilen[0].Menge = 3 }, domain.MangelZeilen},
		{"ein anderer Druck", func(b *domain.Handelsbeleg) { b.Zeilen[0].Printing = "p-9" }, domain.MangelZeilen},
		{"eine andere Ausführung", func(b *domain.Handelsbeleg) { b.Zeilen[0].Finish = domain.FinishEtched }, domain.MangelZeilen},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			meiner, seiner := einBelegpaar()
			fall.verbiegen(&seiner)

			pruefeAbweichung(t, (domain.Belegurkunde{Beleg: seiner}).PasstZu(meiner), fall.mangel)
		})
	}
}

// **Und dieselbe Abweichung in beide Richtungen** (#2500, Vermerk 0202): der
// Absender erfährt einen abgewiesenen Beleg über seine **eigene** Prüfung des
// Gegenbelegs — das trägt nur, wenn sie denselben Mangel findet.
func TestDieAbweichungIstInBeideRichtungenDieselbe(t *testing.T) {
	t.Parallel()

	for name, verbiegen := range map[string]func(*domain.Handelsbeleg){
		"Vorgang":  func(b *domain.Handelsbeleg) { b.Vorgang = "v-2" },
		"Revision": func(b *domain.Handelsbeleg) { b.Abdruck = "abdruck-4" },
		"Parteien": func(b *domain.Handelsbeleg) { b.Parteien = []string{"anna"} },
		"Zeilen":   func(b *domain.Handelsbeleg) { b.Zeilen = b.Zeilen[:1] },
		"Richtung": func(b *domain.Handelsbeleg) { b.Zeilen[0].Art = domain.JournalAbgang },
		"Menge":    func(b *domain.Handelsbeleg) { b.Zeilen[0].Menge = 3 },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			meiner, seiner := einBelegpaar()
			verbiegen(&seiner)

			hier := (domain.Belegurkunde{Beleg: seiner}).Abweichung(meiner)
			drueben := (domain.Belegurkunde{Beleg: meiner}).Abweichung(seiner)

			if hier == nil || drueben == nil || hier.Mangel != drueben.Mangel {
				t.Errorf("hier %v, drüben %v", hier, drueben)
			}
		})
	}
}

// pruefeAbweichung verlangt eine Abweichung mit Mangel und Satz — und dass
// sie eine Validierung bleibt.
func pruefeAbweichung(t *testing.T, fehler error, mangel domain.Gegenbelegmangel) {
	t.Helper()

	var abweichung *domain.Belegabweichung
	if !errors.As(fehler, &abweichung) || !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v, erwartet eine Abweichung", fehler)
	}

	if abweichung.Mangel != mangel || abweichung.Satz == "" {
		t.Errorf("Mangel %q mit Satz %q, erwartet %q", abweichung.Mangel, abweichung.Satz, mangel)
	}

	// **Als Fehler gelesen, trägt sie ihren Satz** — so steht sie in einem
	// Protokoll, das nur `Error()` kennt.
	if !strings.HasSuffix(fehler.Error(), ": "+abweichung.Satz) {
		t.Errorf("Error() gibt %q", fehler.Error())
	}
}

// **Eine Art, die es nicht gibt, hat keine Gegenrichtung.**
//
// **Und sie kommt von außen**: [domain.Journalart] ist eine Zeichenkette, und
// ein Beleg der Gegenseite ist eine Behauptung. Wer sie stillschweigend als
// Abgang läse, ließe sie durch.
func TestEineArtOhneGegenrichtungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	meiner, seiner := einBelegpaar()
	meiner.Zeilen[0].Art = domain.Journalart("korrektur")

	pruefeAbweichung(t, (domain.Belegurkunde{Beleg: seiner}).PasstZu(meiner), domain.MangelZeilen)
}

// **Was hinausgeht, trägt nichts Lokales.**
//
// **Und die Unterschrift passt trotzdem**: [domain.Handelsbeleg.Unterschriftsform]
// rechnet ohnehin ohne diese Felder — wer nur sie hätte, unterschriebe das
// Öffentliche und versendete das Private.
func TestEinBelegGehtOhneLokalesHinaus(t *testing.T) {
	t.Parallel()

	meiner, _ := einBelegpaar()
	meiner.Zeilen[0].ExemplarID = 42
	meiner.Zeilen[0].Lagerort = "Kiste 3"
	meiner.Zeilen[0].Notiz = "für Bert zurückgelegt"

	offen := meiner.Oeffentlich()

	if offen.Zeilen[0].ExemplarID != 0 ||
		offen.Zeilen[0].Lagerort != "" || offen.Zeilen[0].Notiz != "" {
		t.Errorf("hinaus ginge %+v", offen.Zeilen[0])
	}

	if !bytes.Equal(offen.Unterschriftsform(), meiner.Unterschriftsform()) {
		t.Error("die Unterschriftsform hat sich geändert")
	}
}
