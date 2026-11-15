package domain_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Was eine kleine Textaufgabe ist** — AI-05.AK-05 (#2042): Name und
// Flavortext, ein Auftrag unter der gemessenen Grenze, keine Belegfrage.
func TestEineKleineTextaufgabe(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name       string
		felder     []string
		auftrag    string
		belegfrage string
		klein      bool
	}{
		{"der Name", []string{"name"}, "ein alter Hirsch", "", true},
		{"Name und Flavortext", []string{"name", "flavor text"}, "Hirsch", "", true},
		{"ohne Felder ist es die ganze Karte", nil, "Hirsch", "", false},
		{"ein Regeltext ist nicht klein", []string{"name", "rule text"}, "Hirsch", "", false},
		{"eine Belegfrage hängt Stellen an", []string{"name"}, "Hirsch", "Moor", false},
		{"genau an der Grenze", []string{"name"}, strings.Repeat("鹿", domain.KIKleineAuftragsgrenze), "", true},
		{"ein Zeichen darüber", []string{"name"}, strings.Repeat("鹿", domain.KIKleineAuftragsgrenze+1), "", false},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			if klein := domain.IstKleineTextaufgabe(fall.felder, fall.auftrag, fall.belegfrage); klein != fall.klein {
				t.Errorf("%v, erwartet %v", klein, fall.klein)
			}
		})
	}
}

// lagen ist ein Rechner mit Gerätemodell, Ollama und einem Cloudanbieter.
func lagen(geraetBereit bool) []domain.KIAnbieterlage {
	return []domain.KIAnbieterlage{
		{Name: "ollama", Ort: domain.KIOrtLokal, Bereit: true},
		{Name: "anthropic", Ort: domain.KIOrtCloud, Bereit: true},
		{Name: "apple", Ort: domain.KIOrtLokal, Bereit: geraetBereit, Geraet: true},
		{Name: "fern", Ort: domain.KIOrtCloud, Bereit: true},
	}
}

// **Wer in welcher Reihenfolge gefragt wird** — und vor allem, wer nie von
// selbst: ein Cloudanbieter bleibt hinter Journal und Freigabe.
func TestDieReihenfolgeDerAnbieter(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name     string
		gewaehlt string
		klein    bool
		bereit   bool
		erwartet []domain.KIVersuch
	}{
		{"klein: zuerst das Gerät, dann lokal", "", true, true, []domain.KIVersuch{
			{Anbieter: "apple", Weg: domain.KIWegVorrang},
			{Anbieter: "ollama", Weg: domain.KIWegRueckfall},
		}},
		{"klein, Gerät nicht bereit: lokal hat Vorrang", "", true, false, []domain.KIVersuch{
			{Anbieter: "ollama", Weg: domain.KIWegVorrang},
		}},
		{"groß: nie das Gerät", "", false, true, []domain.KIVersuch{
			{Anbieter: "ollama", Weg: domain.KIWegVorrang},
		}},
		{"eine ausdrückliche Wahl gewinnt", "anthropic", true, true, []domain.KIVersuch{
			{Anbieter: "anthropic", Weg: domain.KIWegGewaehlt},
		}},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			reihe := domain.KIReihenfolge(fall.gewaehlt, fall.klein, lagen(fall.bereit))
			if !slices.Equal(reihe, fall.erwartet) {
				t.Errorf("%+v, erwartet %+v", reihe, fall.erwartet)
			}
		})
	}
}

// **Ohne einen bereiten lokalen Anbieter ist die Reihe leer** — und nicht
// heimlich die Cloud.
func TestOhneLokalenAnbieterIstDieReiheLeer(t *testing.T) {
	t.Parallel()

	nurFern := []domain.KIAnbieterlage{
		{Name: "anthropic", Ort: domain.KIOrtCloud, Bereit: true},
		{Name: "ollama", Ort: domain.KIOrtLokal, Bereit: false},
	}

	if reihe := domain.KIReihenfolge("", true, nurFern); len(reihe) != 0 {
		t.Errorf("%+v", reihe)
	}
}

// **Was nach einer Ablehnung angeboten wird**: die bereiten lokalen Anbieter
// außer dem, der abgelehnt hat — nie das Gerät für eine große Aufgabe, nie
// die Cloud.
func TestDieRueckfaelleNachEinerAblehnung(t *testing.T) {
	t.Parallel()

	if rueck := domain.KIRueckfaelle("apple", true, lagen(true)); !slices.Equal(rueck, []string{"ollama"}) {
		t.Errorf("nach dem Gerät: %v", rueck)
	}

	if rueck := domain.KIRueckfaelle("ollama", true, lagen(true)); !slices.Equal(rueck, []string{"apple"}) {
		t.Errorf("nach Ollama, klein: %v", rueck)
	}

	if rueck := domain.KIRueckfaelle("ollama", false, lagen(true)); len(rueck) != 0 {
		t.Errorf("nach Ollama, groß: %v", rueck)
	}
}

// **Ein Modell ohne Anbieter ist nicht eindeutig** — wer den Vorrang will,
// nennt keines.
func TestDerVorrangNimmtKeinModell(t *testing.T) {
	t.Parallel()

	if fehler := domain.PruefeVorrangwunsch("system"); fehler == nil {
		t.Error("ein Modell ging durch")
	}

	if fehler := domain.PruefeVorrangwunsch(""); fehler != nil {
		t.Errorf("%v", fehler)
	}
}
