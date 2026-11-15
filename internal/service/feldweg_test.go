package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

const verbundskript = `combined_editor(field1: card.rule_text, ` +
	`separator: "<line>\n</line>", field2: card.flavor_text)`

// typskript ist `type` aus `magic.mse-game` — **mit gerechnetem Trenner**.
const typskript = `combined_editor(
	field1: card.super_type,
	separator: language().type_separator,
	field2: card.sub_type,
	soft_before_empty: true,
	hide_when_empty:   true
)`

func m15felder() []domain.Field {
	return []domain.Field{
		{Name: "name", Gespeichert: true},
		{Name: "text", Gespeichert: false, Skript: verbundskript},
		{Name: "rule text", Gespeichert: true},
		{Name: "flavor text", Gespeichert: true},
		{Name: "card code", Gespeichert: false, Skript: `card.name + "!"`},
		{Name: "type", Gespeichert: false, Skript: typskript},
		{Name: "super type", Gespeichert: true},
		{Name: "sub type", Gespeichert: true},
	}
}

// **Wer in `Text` tippt, ändert Regeltext und Flavortext** —
// #696.
func TestFeldwegSchreibtDieTeile(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name        string
		feld, wert  string
		geschrieben []string
	}{
		{
			name: "ein gewöhnliches Feld geht seinen Weg",
			feld: "name", wert: "Blitzschlag",
			geschrieben: []string{"name=Blitzschlag"},
		},
		{
			name: "ein Verbund wird am Trenner gespalten",
			feld: "text", wert: "Füge drei Schaden zu.<sep><line>\n</line></sep>Es brennt.",
			geschrieben: []string{
				"rule text=Füge drei Schaden zu.",
				"flavor text=Es brennt.",
			},
		},
		{
			name: "ohne Trenner bekommt der erste Teil alles, der zweite nichts",
			feld: "text", wert: "nur Regeltext",
			geschrieben: []string{"rule text=nur Regeltext", "flavor text="},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			ablage := &wegablage{}
			weg := service.NeuerFeldweg(
				wegkarten{karte: &domain.CardRow{ID: 7, SetID: 3}},
				wegfelder{felder: m15felder()},
				ablage,
			)

			if _, _, err := weg.FeldSetzenMitVorwert(
				context.Background(), 7, fall.feld, fall.wert,
			); err != nil {
				t.Fatalf("schreiben: %v", err)
			}

			if strings.Join(ablage.geschrieben, " | ") !=
				strings.Join(fall.geschrieben, " | ") {
				t.Errorf("geschrieben = %q, erwartet %q",
					ablage.geschrieben, fall.geschrieben)
			}

			// **Ein Verbund geht in *einer* Transaktion hinaus**
			// (#714).
			// Zwei Aufrufe wären zwei Transaktionen — und scheiterte die
			// zweite, stünde die Karte halb geändert da.
			if len(fall.geschrieben) > 1 && ablage.reihen != 1 {
				t.Errorf("%d Reihen, erwartet eine", ablage.reihen)
			}
		})
	}
}

// **Ein Feld, das MSE nicht speichert, speichern wir auch nicht** — und wer es
// trotzdem setzt, bekommt gesagt, warum es nicht geht.
func TestFeldwegWeistUnspeicherbaresAb(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}
	weg := service.NeuerFeldweg(
		wegkarten{karte: &domain.CardRow{ID: 7, SetID: 3}},
		wegfelder{felder: m15felder()},
		ablage,
	)

	_, _, err := weg.FeldSetzenMitVorwert(context.Background(), 7, "card code", "X")
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Fehler = %v, erwartet ErrValidation", err)
	}

	if len(ablage.geschrieben) != 0 {
		t.Errorf("geschrieben wurde trotzdem: %q", ablage.geschrieben)
	}
}

// **Und wo es Teile gibt, nennt die Meldung sie** — „das geht nicht" ohne das
// „sondern" ist eine Sackgasse. `type` rechnet seinen Trenner aus und lässt
// sich nicht spalten; die Felder dahinter stehen trotzdem im Aufruf.
func TestFeldwegNenntDieTeileInDerMeldung(t *testing.T) {
	t.Parallel()

	weg := service.NeuerFeldweg(
		wegkarten{karte: &domain.CardRow{ID: 7, SetID: 3}},
		wegfelder{felder: m15felder()},
		&wegablage{},
	)

	_, _, err := weg.FeldSetzenMitVorwert(context.Background(), 7, "type", "Kreatur")
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if !strings.Contains(err.Error(), "super type") ||
		!strings.Contains(err.Error(), "sub type") {
		t.Errorf("Meldung nennt die Teile nicht: %v", err)
	}

	// **Und sie hängt am Feld** — sonst steht sie als anonymer Hinweis über
	// dem Fenster statt am verursachenden Feld (§6).
	var befund *domain.FieldError
	if !errors.As(err, &befund) || befund.Field != "type" {
		t.Errorf("kein Feldbefund an `type`: %v", err)
	}
}

// **Ein Feld, das das Spiel nicht kennt, geht seinen Weg** — ein Set aus MTG
// JSON trägt Felder, für die die Vorlage keine Erklärung hat, und die zu
// verweigern hieße, sie zu verlieren.
func TestFeldwegKenntNichtJedesFeld(t *testing.T) {
	t.Parallel()

	faelle := []struct {
		name   string
		karten wegkarten
		felder wegfelder
	}{
		{
			name:   "das Spiel erklärt dieses Feld nicht",
			karten: wegkarten{karte: &domain.CardRow{ID: 7, SetID: 3}},
			felder: wegfelder{felder: m15felder()},
		},
		{
			name:   "die Karte lässt sich nicht lesen",
			karten: wegkarten{fehler: errors.New("weg")},
			felder: wegfelder{felder: m15felder()},
		},
		{
			name:   "es gibt sie nicht",
			karten: wegkarten{},
			felder: wegfelder{felder: m15felder()},
		},
		{
			name:   "das Spiel lässt sich nicht lesen",
			karten: wegkarten{karte: &domain.CardRow{ID: 7, SetID: 3}},
			felder: wegfelder{fehler: errors.New("kein Paket")},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			ablage := &wegablage{}
			weg := service.NeuerFeldweg(fall.karten, fall.felder, ablage)

			if _, _, err := weg.FeldSetzenMitVorwert(
				context.Background(), 7, "eigenes feld", "X",
			); err != nil {
				t.Fatalf("schreiben: %v", err)
			}

			if strings.Join(ablage.geschrieben, "|") != "eigenes feld=X" {
				t.Errorf("geschrieben = %q", ablage.geschrieben)
			}
		})
	}
}

// **Ein Teil ohne Feld behält seinen Namen** — dann legt der Bestand ihn so an,
// und der Wert ist wenigstens nicht verloren.
func TestFeldwegBehaeltUnbekannteTeilnamen(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{}
	weg := service.NeuerFeldweg(
		wegkarten{karte: &domain.CardRow{ID: 7, SetID: 3}},
		wegfelder{felder: []domain.Field{
			{Name: "text", Gespeichert: false, Skript: verbundskript},
		}},
		ablage,
	)

	if _, _, err := weg.FeldSetzenMitVorwert(
		context.Background(), 7, "text", "a<sep>-</sep>b",
	); err != nil {
		t.Fatalf("schreiben: %v", err)
	}

	if strings.Join(ablage.geschrieben, "|") != "rule_text=a|flavor_text=b" {
		t.Errorf("geschrieben = %q", ablage.geschrieben)
	}
}

// **Ein Schreibfehler am ersten Teil hält den zweiten auf** — sonst stünde die
// Karte halb geändert da, und niemand wüsste, welche Hälfte (§8).
func TestFeldwegHaeltBeimSchreibfehlerAn(t *testing.T) {
	t.Parallel()

	ablage := &wegablage{fehler: errors.New("Platte voll")}
	weg := service.NeuerFeldweg(
		wegkarten{karte: &domain.CardRow{ID: 7, SetID: 3}},
		wegfelder{felder: m15felder()},
		ablage,
	)

	_, _, err := weg.FeldSetzenMitVorwert(context.Background(), 7, "text", "a")
	if err == nil {
		t.Fatal("kein Fehler")
	}

	if len(ablage.geschrieben) != 0 {
		t.Errorf("geschrieben = %q", ablage.geschrieben)
	}
}

// **Welche Felder hinter einem berechneten Feld stehen** — die Oberfläche
// bietet sie an seiner Stelle an, statt eine Eingabe entgegenzunehmen, die
// nirgends ankommt.
func TestTeilfelder(t *testing.T) {
	t.Parallel()

	felder := m15felder()

	faelle := []struct {
		name  string
		form  domain.Field
		teile []string
	}{
		{name: "ein gewöhnliches Feld hat keine Teile", form: felder[0]},
		{
			name: "der Verbund nennt seine Felder, wie sie heißen",
			form: felder[1], teile: []string{"rule text", "flavor text"},
		},
		{name: "ein gerechnetes Feld hat keine Teile", form: felder[4]},
		{
			// **Ein gerechneter Trenner verschließt den Rückweg, nicht die
			// Auskunft.** `type` setzt sich aus `super type` und `sub type`
			// zusammen; **welcher** Trenner dazwischen steht, weiß erst der
			// Auswerter (`language().type_separator`). Spalten lässt es sich
			// damit nicht — aber die Oberfläche kann die beiden Felder
			// anbieten, und das ist der Weg, den jemand wirklich geht.
			name: "type nennt seine Felder trotz gerechnetem Trenner",
			form: felder[5], teile: []string{"super type", "sub type"},
		},
	}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			teile := service.Teilfelder(felder, fall.form)
			if strings.Join(teile, "|") != strings.Join(fall.teile, "|") {
				t.Errorf("Teile = %q, erwartet %q", teile, fall.teile)
			}
		})
	}
}
