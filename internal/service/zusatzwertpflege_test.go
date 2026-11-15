package service_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Werte der Zusatzfelder lassen sich lesen und setzen** —
// #2298.
//
// MSE2 bearbeitet sie auf der Karte selbst: der Editor zeigt die Felder der
// Karte **und** die Zusatzfelder ihres Stils
// (`setData(card->data, &card->extraDataFor(*stylesheet))`,
// `src/render/card/viewer.cpp`, Zeile 139, Fassung f9d9356 vom 2021-07-06),
// bearbeitbar, wo das Feld `editable` ist. Gespeichert wird unter
// `extra_data:` in den Block der Stilvorlage.

// zusatzkarten ist der Bestand der Vorrichtung — eine Karte je Kennung.
type zusatzkarten map[int64]*domain.CardRow

func (k zusatzkarten) KarteLesen(_ context.Context, karteID int64) (*domain.CardRow, error) {
	karte, da := k[karteID]
	if !da {
		return nil, domain.ErrNotFound
	}

	return karte, nil
}

// zusatzablage schreibt mit, was gesetzt wurde — und hält, wie der Bestand,
// den erwarteten Stand der Karte gegen den, auf dem sie steht.
type zusatzablage struct {
	geschrieben []string
	fehler      error

	// karten ist der Bestand, gegen den der erwartete Stand gilt; nil heißt:
	// jeder Stand passt.
	karten zusatzkarten

	// dazwischen läuft einmal vor dem ersten Schreiben — eine andere Fläche,
	// die dieselbe Karte gerade ändert.
	dazwischen func()

	// erwartet sind die Stände, die als Bedingung ankamen.
	erwartet []string
}

func (a *zusatzablage) FeldSetzenAufStand(
	_ context.Context, karteID int64, feld, text, erwartet string,
) (vorher string, hatte bool, fehler error) {
	werte := []domain.Value{{Field: feld, Text: text}}

	if a.fehler != nil {
		return "", false, a.fehler
	}

	if a.dazwischen != nil {
		a.dazwischen()
		a.dazwischen = nil
	}

	a.erwartet = append(a.erwartet, erwartet)

	karte, da := a.karten[karteID]
	if da && erwartet != karte.Stand() {
		return "", false, domain.ErrConflict
	}

	for _, wert := range werte {
		a.geschrieben = append(a.geschrieben, wert.Field+"="+wert.Text)

		if da {
			karte.Values = mitWert(karte.Values, wert)
		}
	}

	return "", false, nil
}

// mitWert setzt einen Feldwert in eine Werteliste.
func mitWert(werte []domain.Value, neu domain.Value) []domain.Value {
	for i := range werte {
		if werte[i].Field == neu.Field {
			werte[i].Text = neu.Text

			return werte
		}
	}

	return append(werte, neu)
}

func zusatzvorlage() service.Vorlage {
	return service.Vorlage{Stil: msepaket.Stil{
		Paket: "magic-probe.mse-style",
		Zusatzfelder: []domain.Field{
			{Name: "keyword icons", Kind: domain.KindMultipleChoice, Choices: []string{"flying", "haste"}, Editierbar: true, Gespeichert: true},
			{Name: "promo label", Kind: domain.KindText, Editierbar: true, Gespeichert: true},
			{Name: "pt box", Kind: domain.KindChoice, Skript: "card.card_color", Gespeichert: false},
			{Name: "foil layer", Kind: domain.KindChoice, Editierbar: false, Gespeichert: true},
		},
	}}
}

func zusatzpflege(karten zusatzkarten, ablage service.Kartenfeldablage) *service.Zusatzwertpflege {
	return service.NeueZusatzwertpflege(karten, festerStil{vorlage: zusatzvorlage(), da: true}, ablage)
}

func TestDieZusatzfelderEinerKarteStehenMitIhremWert(t *testing.T) {
	karten := zusatzkarten{1: {ID: 1, SetID: 7, Values: []domain.Value{
		{Field: "extra data", Text: "magic-probe:\n\tkeyword_icons: flying\nmagic-anders:\n\tpromo_label: fremd"},
	}}}

	felder, err := zusatzpflege(karten, &zusatzablage{}).Felder(t.Context(), 7, 1)
	if err != nil {
		t.Fatalf("Felder: %v", err)
	}

	// **Nur, was sich von Hand setzen lässt**: `pt box` speichert nichts,
	// `foil layer` ist nicht bearbeitbar.
	namen := make([]string, 0, len(felder))
	for _, feld := range felder {
		namen = append(namen, feld.Feld.Name)
	}

	if !slices.Equal(namen, []string{"keyword icons", "promo label"}) {
		t.Fatalf("Felder %v", namen)
	}

	if felder[0].Wert != "flying" || !felder[0].Gespeichert {
		t.Errorf("keyword icons: %+v", felder[0])
	}

	// **Auch aus der flachen Altform** (Befund 6).
	karten[1].Values[0].Text = "magic-probe:\nkeyword_icons: haste"

	alt, err := zusatzpflege(karten, &zusatzablage{}).Felder(t.Context(), 7, 1)
	if err != nil || alt[0].Wert != "haste" {
		t.Errorf("aus der Altform: %+v, %v", alt, err)
	}

	// **Der Wert einer anderen Stilvorlage gilt nicht.**
	if felder[1].Wert != "" || felder[1].Gespeichert {
		t.Errorf("promo label: %+v", felder[1])
	}
}

func TestEinZusatzwertWirdInDenBlockSeinerVorlageGesetzt(t *testing.T) {
	for _, fall := range []struct {
		name, vorher, feld, wert, nachher string
	}{
		{
			name: "ein neuer Block, wo die Karte keinen trägt",
			feld: "promo label", wert: "Prerelease",
			nachher: "magic-probe:\n\tpromo_label: Prerelease",
		},
		{
			name:   "ein Wert ersetzt den alten und lässt die anderen stehen",
			vorher: "magic-anders:\n\tpromo_label: fremd\nmagic-probe:\n\tkeyword_icons: flying\n\tpromo_label: alt",
			feld:   "promo label", wert: "neu",
			nachher: "magic-anders:\n\tpromo_label: fremd\nmagic-probe:\n\tkeyword_icons: flying\n\tpromo_label: neu",
		},
		{
			name:   "ein neues Feld kommt in den vorhandenen Block",
			vorher: "magic-probe:\n\tkeyword_icons: flying",
			feld:   "promo label", wert: "Prerelease",
			nachher: "magic-probe:\n\tkeyword_icons: flying\n\tpromo_label: Prerelease",
		},
		{
			name:   "ein mehrzeiliger Wert steht als Block",
			vorher: "magic-probe:\n\tkeyword_icons: flying",
			feld:   "promo label", wert: "Zeile eins\nZeile zwei",
			nachher: "magic-probe:\n\tkeyword_icons: flying\n\tpromo_label:\n\t\tZeile eins\n\t\tZeile zwei",
		},
		{
			// Befund 6: die flache Altform kommt beim Setzen in ihre Gestalt.
			name:   "die flache Altform wird zum Block",
			vorher: "magic-probe:\nkeyword_icons: flying",
			feld:   "promo label", wert: "Prerelease",
			nachher: "magic-probe:\n\tkeyword_icons: flying\n\tpromo_label: Prerelease",
		},
		{
			name:   "ein leerer Wert bleibt als leere Zeile",
			vorher: "magic-probe:\n\tpromo_label: alt",
			feld:   "promo label", wert: "",
			nachher: "magic-probe:\n\tpromo_label:",
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			karten := zusatzkarten{1: {ID: 1, SetID: 7}}
			if fall.vorher != "" {
				karten[1].Values = []domain.Value{{Field: "extra data", Text: fall.vorher}}
			}

			ablage := &zusatzablage{}

			if _, err := zusatzpflege(karten, ablage).Setzen(t.Context(), 7, 1, fall.feld, fall.wert); err != nil {
				t.Fatalf("Setzen: %v", err)
			}

			if !slices.Equal(ablage.geschrieben, []string{"extra data=" + fall.nachher}) {
				t.Errorf("geschrieben %q\nerwartet %q", ablage.geschrieben, "extra data="+fall.nachher)
			}
		})
	}
}

// **Was sich nicht setzen lässt, wird benannt, und nichts wird geschrieben.**
func TestEinZusatzwertDerNichtGehtWirdBenannt(t *testing.T) {
	for _, fall := range []struct {
		name   string
		karten zusatzkarten
		set    int64
		feld   string
		art    error
	}{
		{name: "ein Feld, das der Stil nicht kennt", karten: zusatzkarten{1: {ID: 1, SetID: 7}}, set: 7, feld: "gibtsnicht", art: domain.ErrValidation},
		{name: "ein Feld, das nicht bearbeitbar ist", karten: zusatzkarten{1: {ID: 1, SetID: 7}}, set: 7, feld: "foil layer", art: domain.ErrValidation},
		{name: "ein Feld, das nichts speichert", karten: zusatzkarten{1: {ID: 1, SetID: 7}}, set: 7, feld: "pt box", art: domain.ErrValidation},
		{name: "eine Karte aus einem anderen Set", karten: zusatzkarten{1: {ID: 1, SetID: 8}}, set: 7, feld: "promo label", art: domain.ErrNotFound},
		{name: "eine Karte, die es nicht gibt", karten: zusatzkarten{}, set: 7, feld: "promo label", art: domain.ErrNotFound},
		{
			name: "Zusatzdaten, die sich nicht lesen lassen",
			karten: zusatzkarten{1: {ID: 1, SetID: 7, Values: []domain.Value{
				{Field: "extra data", Text: "magic-probe:\n\tpromo_label: \xff"},
			}}},
			set: 7, feld: "promo label", art: domain.ErrValidation,
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			ablage := &zusatzablage{}

			_, err := zusatzpflege(fall.karten, ablage).Setzen(t.Context(), fall.set, 1, fall.feld, "x")
			if !errors.Is(err, fall.art) {
				t.Errorf("Fehler %v, erwartet %v", err, fall.art)
			}

			if len(ablage.geschrieben) > 0 {
				t.Errorf("geschrieben trotz Fehler: %q", ablage.geschrieben)
			}
		})
	}
}

// **Und beim Lesen dieselben Grenzen** — eine fremde Karte ist nicht da,
// unlesbare Zusatzdaten werden benannt, und ohne Stil gibt es keine Felder.
func TestDieZusatzfelderEinerKarteHabenIhreGrenzen(t *testing.T) {
	t.Run("eine Karte aus einem anderen Set", func(t *testing.T) {
		_, err := zusatzpflege(zusatzkarten{1: {ID: 1, SetID: 8}}, &zusatzablage{}).Felder(t.Context(), 7, 1)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("Fehler %v", err)
		}
	})

	t.Run("Zusatzdaten, die sich nicht lesen lassen", func(t *testing.T) {
		karten := zusatzkarten{1: {ID: 1, SetID: 7, Values: []domain.Value{
			{Field: "extra data", Text: "magic-probe:\n\tpromo_label: \xff"},
		}}}

		_, err := zusatzpflege(karten, &zusatzablage{}).Felder(t.Context(), 7, 1)
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("Fehler %v", err)
		}
	})

	// **Ein Stil, der sich nicht lesen lässt, ist kein Stil ohne Zusatzfelder**
	// (J05, #2269, Befund 4 des Zweitblicks über die Folgepakete): bis dahin
	// kam eine leere Liste ohne Fehler, die Oberfläche blendete den Abschnitt
	// aus, und ein Setzen meldete, das Feld sei keins.
	t.Run("eine Karte ohne lesbaren Stil", func(t *testing.T) {
		stil := festerStil{befunde: []domain.Befund{{Meldung: "für den Stil \"m15\" ist kein Paketverzeichnis eingerichtet"}}}
		pflege := service.NeueZusatzwertpflege(zusatzkarten{1: {ID: 1, SetID: 7}}, stil, &zusatzablage{})

		felder, err := pflege.Felder(t.Context(), 7, 1)
		if !errors.Is(err, domain.ErrConflict) || len(felder) != 0 {
			t.Errorf("Felder %v, Fehler %v", felder, err)
		}

		if grund := domain.GrundVon(err); grund == nil || !strings.Contains(grund.Error(), "kein Paketverzeichnis") {
			t.Errorf("der Grund fehlt: %v", grund)
		}

		_, err = pflege.Setzen(t.Context(), 7, 1, "promo label", "x")
		if !errors.Is(err, domain.ErrConflict) {
			t.Errorf("Setzen: %v", err)
		}
	})

	// Ein Stil, der ohne Befund fehlt, sagt das auch — als Grund.
	t.Run("eine Karte ohne Stil und ohne Befund", func(t *testing.T) {
		pflege := service.NeueZusatzwertpflege(zusatzkarten{1: {ID: 1, SetID: 7}}, festerStil{}, &zusatzablage{})

		_, err := pflege.Felder(t.Context(), 7, 1)
		if grund := domain.GrundVon(err); !errors.Is(err, domain.ErrConflict) || grund == nil ||
			!strings.Contains(grund.Error(), "keinen Befund") {
			t.Errorf("Fehler %v, Grund %v", err, grund)
		}
	})
}

// **Ein Fehler beim Schreiben kommt beim Aufrufer an.**
func TestEinZusatzwertMitScheiterndemSchreibenMeldetDas(t *testing.T) {
	ablage := &zusatzablage{fehler: domain.ErrConflict}

	_, err := zusatzpflege(zusatzkarten{1: {ID: 1, SetID: 7}}, ablage).Setzen(t.Context(), 7, 1, "promo label", "x")
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("Fehler %v", err)
	}
}

// **Eine Änderung dazwischen geht nicht verloren** (Befund 1 des Nachlaufs
// über E/F). Gelesen wird die Karte, geschrieben ihr ganzer Block
// `extra data`; setzte eine andere Fläche dazwischen ein anderes Zusatzfeld,
// schrieb der zweite Aufruf den Block von vorhin zurück. Jetzt schreibt er
// nur auf den Stand, den er gelesen hat, und liest bei einem Widerspruch neu.
func TestEineAenderungDazwischenGehtNichtVerloren(t *testing.T) {
	karten := zusatzkarten{1: {ID: 1, SetID: 7, Values: []domain.Value{
		{Field: "extra data", Text: "magic-probe:\n\tpromo_label: alt"},
	}}}
	ablage := &zusatzablage{karten: karten}
	ablage.dazwischen = func() {
		karten[1].Values[0].Text = "magic-probe:\n\tpromo_label: alt\n\tkeyword_icons: haste"
	}

	vorwert, err := zusatzpflege(karten, ablage).Setzen(t.Context(), 7, 1, "promo label", "neu")
	if err != nil {
		t.Fatalf("Setzen: %v", err)
	}

	block := karten[1].Values[0].Text
	for _, soll := range []string{"promo_label: neu", "keyword_icons: haste"} {
		if !strings.Contains(block, soll) {
			t.Errorf("im Block fehlt %q:\n%s", soll, block)
		}
	}

	if len(ablage.erwartet) != 2 {
		t.Errorf("%d Versuche, erwartet zwei — der erste muss am Stand scheitern", len(ablage.erwartet))
	}

	// **Der Vorwert ist der des geglückten Versuchs** — der, auf den
	// geschrieben wurde, und mit ihm geht die Rücknahme.
	if !vorwert.Gespeichert || vorwert.Wert != "alt" ||
		!strings.Contains(vorwert.Block, "keyword_icons: haste") {
		t.Errorf("Vorwert %+v, erwartet »alt« im Block mit keyword_icons", vorwert)
	}
}

// **Ein Widerspruch, der nicht vergeht, wird gemeldet** — nicht endlos
// wiederholt.
func TestEinBleibenderWiderspruchWirdGemeldet(t *testing.T) {
	karten := zusatzkarten{1: {ID: 1, SetID: 7}}
	ablage := &zusatzablage{karten: karten}

	immer := func() { karten[1].Values = append(karten[1].Values, domain.Value{Field: "x", Text: "y"}) }
	ablage.dazwischen = immer

	pflege := zusatzpflege(karten, &wiederholtDazwischen{zusatzablage: ablage, dazwischen: immer})

	_, err := pflege.Setzen(t.Context(), 7, 1, "promo label", "neu")
	if !errors.Is(err, domain.ErrConflict) {
		t.Errorf("Fehler %v, erwartet ErrConflict", err)
	}
}

// wiederholtDazwischen ändert die Karte vor **jedem** Schreiben.
type wiederholtDazwischen struct {
	*zusatzablage
	dazwischen func()
}

func (w *wiederholtDazwischen) FeldSetzenAufStand(
	ctx context.Context, karteID int64, feld, wert, erwartet string,
) (vorher string, hatte bool, fehler error) {
	w.zusatzablage.dazwischen = w.dazwischen

	return w.zusatzablage.FeldSetzenAufStand(ctx, karteID, feld, wert, erwartet)
}
