package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

func probeSpiel() domain.Game {
	return domain.Game{
		Name: "probe",
		Fields: []domain.Field{
			{Name: "name", Kind: domain.KindText},
			{Name: "farbe", Kind: domain.KindChoice, Choices: []string{"rot", "blau"}},
			{Name: "bild", Kind: domain.KindImage},
		},
	}
}

// Die Reihenfolge der Feldwerte ist die der Datei.
//
// Deshalb eine geordnete Liste und keine Abbildung: eine Go-Map hat keine
// Reihenfolge, und die Rundlaufzusage verlangt, dass zurückkommt, was hineinging.
func TestSetValueBehaeltDieReihenfolge(t *testing.T) {
	karte := &domain.Card{}

	karte.SetValue("name", "Erste")
	karte.SetValue("farbe", "rot")
	karte.SetValue("name", "Geändert")

	if len(karte.Values) != 2 {
		t.Fatalf("%d Werte, erwartet waren 2", len(karte.Values))
	}
	if karte.Values[0].Field != "name" || karte.Values[0].Text != "Geändert" {
		t.Errorf("der erste Wert ist %+v", karte.Values[0])
	}
	if karte.Values[1].Field != "farbe" {
		t.Errorf("die Reihenfolge hat sich geändert: %+v", karte.Values)
	}
}

// „Leer" und „gibt es nicht" sind zwei verschiedene Aussagen.
func TestValueTrenntLeerVonFehlend(t *testing.T) {
	karte := &domain.Card{}
	karte.SetValue("name", "")

	text, vorhanden := karte.Value("name")
	if !vorhanden || text != "" {
		t.Errorf("ein absichtlich geleertes Feld gilt als fehlend (%q, %v)", text, vorhanden)
	}

	if _, vorhanden := karte.Value("gibtsnicht"); vorhanden {
		t.Error("ein fehlendes Feld gilt als vorhanden")
	}
}

// Ein unbekanntes Feld ist kein Fehler.
//
// Ein Set kann mit einem neueren Spielpaket entstanden sein oder mit einem, das
// jemand selbst erweitert hat. Der Wert wird mitgeführt — sonst verlöre ein
// Rundlauf ihn, und das Set käme verändert zurück.
func TestEinUnbekanntesFeldIstKeinFehler(t *testing.T) {
	karte := &domain.Card{}
	karte.SetValue("erfunden", "irgendwas")

	if err := karte.Validate(probeSpiel()); err != nil {
		t.Errorf("ein unbekanntes Feld wurde abgewiesen: %v", err)
	}
}

func TestEineAuswahlAusserhalbIhrerMoeglichkeitenWirdAbgewiesen(t *testing.T) {
	karte := &domain.Card{}
	karte.SetValue("farbe", "grün")

	err := karte.Validate(probeSpiel())
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Fehler %v, erwartet war eine Validierung", err)
	}

	var befund *domain.FieldError
	if !errors.As(err, &befund) || befund.Field != "farbe" {
		t.Errorf("der Befund hängt nicht am Feld farbe: %+v", befund)
	}
}

// Ein leerer Auswahlwert geht durch.
//
// „Noch nichts ausgewählt" ist ein gültiger Zustand — wer ihn abweist, kann
// eine Karte nicht anlegen, ohne jedes Auswahlfeld sofort zu füllen.
func TestEineLeereAuswahlGehtDurch(t *testing.T) {
	karte := &domain.Card{}
	karte.SetValue("farbe", "")

	if err := karte.Validate(probeSpiel()); err != nil {
		t.Errorf("eine leere Auswahl wurde abgewiesen: %v", err)
	}
}

func TestEinDoppeltesFeldAnEinerKarteWirdAbgewiesen(t *testing.T) {
	karte := &domain.Card{Values: []domain.Value{
		{Field: "name", Text: "Erste"},
		{Field: "name", Text: "Zweite"},
	}}

	if err := karte.Validate(probeSpiel()); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet war ein Konflikt", err)
	}
}

func TestEinSetPruefTSeineKartenUndNenntDieNummer(t *testing.T) {
	set := &domain.Set{
		Game: "probe",
		Cards: []*domain.Card{
			{Values: []domain.Value{{Field: "name", Text: "Erste"}}},
			{Values: []domain.Value{{Field: "farbe", Text: "grün"}}},
		},
	}

	err := set.Validate(probeSpiel())
	if err == nil {
		t.Fatal("die zweite Karte kam durch")
	}
	if !strings.Contains(err.Error(), "Karte 2") {
		t.Errorf("die Meldung %q nennt die Kartennummer nicht", err.Error())
	}

	// Und der Feldbezug geht beim Umschließen nicht verloren (AGENTS.md §6).
	var befund *domain.FieldError
	if !errors.As(err, &befund) || befund.Field != "farbe" {
		t.Errorf("der Feldbezug ging verloren: %+v", befund)
	}
}

// Auch `Validate` überlebt eine fehlende Karte.
//
// Der Schutz steht in `doppelteFelder`, das beide Wege als erstes rufen — hier
// steht der Beleg, dass der zweite Weg ihn wirklich erreicht. Ohne ihn wäre die
// Behauptung im Kommentar von `pruefe` ungeprüft.
func TestEineFehlendeKarteStuerztAuchValidateNichtAb(t *testing.T) {
	set := &domain.Set{Game: "probe", Cards: []*domain.Card{nil}}

	if err := set.Validate(probeSpiel()); err == nil {
		t.Error("eine fehlende Karte kam durch")
	}
}

func TestEinSetOhneSpielWirdAbgewiesen(t *testing.T) {
	if err := (&domain.Set{}).Validate(probeSpiel()); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("ein Set ohne Spiel kam durch: %v", err)
	}
}

func TestEinSetGegenDasFalscheSpielWirdAbgewiesen(t *testing.T) {
	set := &domain.Set{Game: "anderes"}

	if err := set.Validate(probeSpiel()); !errors.Is(err, domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet war ein Konflikt", err)
	}
}

func TestSetValueGibtSetangaben(t *testing.T) {
	set := &domain.Set{Values: []domain.Value{{Field: "title", Text: "Mein Set"}}}

	if text, vorhanden := set.Value("title"); !vorhanden || text != "Mein Set" {
		t.Errorf("Setangabe %q, %v", text, vorhanden)
	}
	if _, vorhanden := set.Value("gibtsnicht"); vorhanden {
		t.Error("eine fehlende Setangabe gilt als vorhanden")
	}
}

// ------------------------------------------------------------------ Spiel und Stil

func TestSpielFindetSeineFelder(t *testing.T) {
	spiel := probeSpiel()

	if feld, bekannt := spiel.Field("farbe"); !bekannt || feld.Kind != domain.KindChoice {
		t.Errorf("Feld %+v, %v", feld, bekannt)
	}
	if _, bekannt := spiel.Field("gibtsnicht"); bekannt {
		t.Error("ein erfundenes Feld gilt als bekannt")
	}
}

func TestSpielUndFeldValidierung(t *testing.T) {
	faelle := []struct {
		name  string
		spiel domain.Game
		gut   bool
	}{
		{"ein gültiges Spiel", probeSpiel(), true},
		{"ohne Namen", domain.Game{Fields: probeSpiel().Fields}, false},
		{
			name:  "ein Feld ohne Namen",
			spiel: domain.Game{Name: "probe", Fields: []domain.Field{{Kind: domain.KindText}}},
		},
		{
			name: "eine Auswahl ohne Möglichkeiten",
			spiel: domain.Game{Name: "probe", Fields: []domain.Field{
				{Name: "farbe", Kind: domain.KindChoice},
			}},
		},
		{
			name: "dasselbe Feld zweimal",
			spiel: domain.Game{Name: "probe", Fields: []domain.Field{
				{Name: "name", Kind: domain.KindText},
				{Name: "name", Kind: domain.KindText},
			}},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			err := fall.spiel.Validate()

			if fall.gut && err != nil {
				t.Errorf("abgewiesen: %v", err)
			}
			if !fall.gut && err == nil {
				t.Error("angenommen, erwartet war eine Abweisung")
			}
		})
	}
}

func TestStilValidierung(t *testing.T) {
	gut := domain.Style{Name: "m15", Game: "probe", Width: 375, Height: 523, DPI: 150}

	faelle := []struct {
		name string
		stil domain.Style
		gut  bool
	}{
		{"ein gültiger Stil", gut, true},
		{"ohne Namen", domain.Style{Game: "probe", Width: 1, Height: 1, DPI: 1}, false},
		{"ohne Spiel", domain.Style{Name: "m15", Width: 1, Height: 1, DPI: 1}, false},
		{"ohne Breite", domain.Style{Name: "m15", Game: "probe", Height: 1, DPI: 1}, false},
		{"ohne Höhe", domain.Style{Name: "m15", Game: "probe", Width: 1, DPI: 1}, false},
		{"ohne Auflösung", domain.Style{Name: "m15", Game: "probe", Width: 1, Height: 1}, false},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			err := fall.stil.Validate()

			if fall.gut && err != nil {
				t.Errorf("abgewiesen: %v", err)
			}
			if !fall.gut && err == nil {
				t.Error("angenommen, erwartet war eine Abweisung")
			}
		})
	}
}

func TestEineGueltigeAuswahlGehtDurch(t *testing.T) {
	karte := &domain.Card{}
	karte.SetValue("farbe", "blau")

	if err := karte.Validate(probeSpiel()); err != nil {
		t.Errorf("eine gültige Auswahl wurde abgewiesen: %v", err)
	}
}

func TestEinFehlerfreiesSetGehtDurch(t *testing.T) {
	set := &domain.Set{
		Game:  "probe",
		Cards: []*domain.Card{{Values: []domain.Value{{Field: "name", Text: "Erste"}}}},
	}

	if err := set.Validate(probeSpiel()); err != nil {
		t.Errorf("ein fehlerfreies Set wurde abgewiesen: %v", err)
	}
}

// Die Grundprüfung sieht, was ohne Spielpaket zu sehen ist.
//
// Sie läuft beim Import — also bevor feststeht, ob das zugehörige Spiel
// überhaupt installiert ist. Was sie **nicht** sieht: ob die Feldwerte zu den
// Definitionen passen. Das kann nur `Validate`, und es braucht das Spiel.
func TestValidateBasicsOhneSpielpaket(t *testing.T) {
	faelle := []struct {
		name string
		set  *domain.Set
		gut  bool
	}{
		{
			name: "ein brauchbares Set",
			set:  &domain.Set{Game: "probe", Cards: []*domain.Card{{Values: []domain.Value{{Field: "name", Text: "Erste"}}}}},
			gut:  true,
		},
		{
			name: "ein Set ohne Spiel",
			set:  &domain.Set{Cards: []*domain.Card{}},
		},
		{
			name: "eine Karte mit doppeltem Feld",
			set: &domain.Set{Game: "probe", Cards: []*domain.Card{
				{}, {Values: []domain.Value{{Field: "name", Text: "a"}, {Field: "name", Text: "b"}}},
			}},
		},
		{
			name: "eine leere Karte ist in Ordnung",
			set:  &domain.Set{Game: "probe", Cards: []*domain.Card{{}}},
			gut:  true,
		},
		{
			// `Cards` ist eine Liste von **Zeigern**, damit eine Karte an ihrem
			// Platz geändert werden kann. Damit ist ein nil-Eintrag möglich —
			// und er stürzte ab, statt abgewiesen zu werden. Über eine
			// Wails-Bindung heißt Absturz nicht Fehlermeldung, sondern beendete
			// Anwendung. Gefunden vom Nachlauf des Zweitblicks über Phase 2.
			name: "eine fehlende Karte stürzt nicht ab",
			set:  &domain.Set{Game: "probe", Cards: []*domain.Card{nil}},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			err := fall.set.ValidateBasics()

			if fall.gut && err != nil {
				t.Errorf("abgewiesen: %v", err)
			}
			if !fall.gut && err == nil {
				t.Error("angenommen, erwartet war eine Abweisung")
			}
			if !fall.gut && err != nil && !strings.Contains(err.Error(), "Karte 2") &&
				fall.name == "eine Karte mit doppeltem Feld" {
				t.Errorf("die Meldung %q nennt die Kartennummer nicht", err.Error())
			}
		})
	}
}
