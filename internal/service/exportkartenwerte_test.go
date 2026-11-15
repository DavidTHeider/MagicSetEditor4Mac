package service_test

import (
	"context"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die Karten der Ausfuhr tragen ihre gerechneten Werte** —
// #2006.
//
// `export_set` wertet die Vorlage im Kontext des Sets aus, und dessen Karten
// tragen die Werte, die `SetScriptManager::updateAll` beim Laden gerechnet hat
// (`src/gui/html_export_window.cpp`, Zeilen 55–90, und
// `src/script/script_manager.cpp`, Zeilen 265–300, Fassung f9d9356 vom
// 2021-07-06). Gemessen am Leistungsvergleich: `magic-text` und `magic-mws`
// ordnen nach `card.card_number` und schrieben `Card #:` und `Type & Class:`
// leer — beide Felder tragen `save value: false` und ein `script:` über dem
// `init script:` des Spiels.

// kartenspiel ist ein Spiel in klein: ein gespeichertes Feld, ein berechnetes
// und eines, das die Karte im Set zählt.
const kartenspielskript = `
sortname := { to_lower(input) }
ganzer_name := { card.name + " (" + card.ort + ")" }
kartennummer := {
	position(of: card, in: set, order_by: { sortname(card.name) }) + 1
}
`

func kartenspielfelder() []domain.Field {
	return []domain.Field{
		{Name: "name", Kind: domain.KindText},
		{Name: "ort", Kind: domain.KindText},
		{Name: "ganz", Kind: domain.KindText, Skript: "ganzer_name()"},
		{Name: "nummer", Kind: domain.KindText, Skript: "kartennummer()"},
		{Name: "art", Kind: domain.KindText, Default: `"Standard"`},
	}
}

func kartenspielvorlage(skript string) msepaket.Exportvorlage {
	return msepaket.Exportvorlage{
		KurzerName: "probe", Skript: skript,
		Spielskript: kartenspielskript, Kartenfelder: kartenspielfelder(),
	}
}

func dreiKartenUngeordnet() []*domain.CardRow {
	karte := func(id int64, name, ort string) *domain.CardRow {
		return &domain.CardRow{ID: id, Number: int(id), Values: []domain.Value{
			{Field: "name", Text: name}, {Field: "ort", Text: ort},
		}}
	}

	return []*domain.CardRow{
		karte(1, "Cäsar", "Rom"), karte(2, "Anton", "Gaza"), karte(3, "Berta", "Bonn"),
	}
}

func fuehreKartenspiel(t *testing.T, skript string) (string, []domain.Befund) {
	t.Helper()

	ausgabe, befunde, err := probelauf(service.Ausfuhrmittel{}).Fuehre(
		t.Context(), kartenspielvorlage(skript), dreiKartenUngeordnet(), nil, nil)
	if err != nil {
		t.Fatalf("führen: %v", err)
	}

	return ausgabe, befunde
}

// **Ein Feld mit `script:` trägt seinen gerechneten Wert**, auch wenn die Datei
// nichts davon speichert.
func TestDieAusfuhrGibtDenKartenDieWerteIhrerSkriptfelder(t *testing.T) {
	t.Parallel()

	ausgabe, _ := fuehreKartenspiel(t, `to_string(for each card in cards do card.ganz + ";")`)

	if ausgabe != "Cäsar (Rom);Anton (Gaza);Berta (Bonn);" {
		t.Errorf("Ausgabe %q", ausgabe)
	}
}

// **Ein Wertskript schlägt den gespeicherten Wert** — MSE rechnet es beim Laden
// über ihn (`field().script.invokeOn(ctx, value)`, `src/data/field/text.cpp`,
// Zeilen 178–179, Fassung f9d9356 vom 2021-07-06).
func TestEinWertskriptSchlaegtDenGespeichertenWert(t *testing.T) {
	t.Parallel()

	karten := dreiKartenUngeordnet()
	karten[0].Values = append(karten[0].Values, domain.Value{Field: "ganz", Text: "alt"})

	ausgabe, _, err := probelauf(service.Ausfuhrmittel{}).Fuehre(t.Context(),
		kartenspielvorlage(`to_string(for each card in cards do card.ganz + ";")`),
		karten, nil, nil)
	if err != nil {
		t.Fatalf("führen: %v", err)
	}

	if !strings.HasPrefix(ausgabe, "Cäsar (Rom);") {
		t.Errorf("Ausgabe %q — der gespeicherte Wert »alt« steht noch da", ausgabe)
	}
}

// **`card_number` zählt die Karte im Set**, und die Vorlage kann danach
// ordnen — das ist `magic-text`: `sort_list(cards, order_by:
// {input.card_number})`.
func TestDieKartennummerZaehltDieKarteImSetUndOrdnetDieVorlage(t *testing.T) {
	t.Parallel()

	ausgabe, _ := fuehreKartenspiel(t, `to_string(for each card in `+
		`sort_list(cards, order_by: {input.nummer}) do card.name + "=" + card.nummer + ";")`)

	if ausgabe != "Anton=1;Berta=2;Cäsar=3;" {
		t.Errorf("Ausgabe %q", ausgabe)
	}
}

// **Die Nummer ist die der gerechneten Liste.** Der Schlüssel liest
// `card.gruppe`, und das ist ein Feld, das erst ein Skript bildet — in der
// zweiten Runde steht es an jeder Karte der Liste, in der ersten an keiner.
func TestDieKartennummerOrdnetNachGerechnetenWerten(t *testing.T) {
	t.Parallel()

	vorlage := msepaket.Exportvorlage{
		KurzerName: "probe",
		Skript:     `to_string(for each card in cards do card.nummer + ";")`,
		Spielskript: `
gruppe_von := { if card.ort == "Rom" then "a" else "b" }
kartennummer := {
	position(of: card, in: set, order_by: { card.gruppe + card.name }) + 1
}
`,
		Kartenfelder: []domain.Field{
			{Name: "name", Kind: domain.KindText},
			{Name: "ort", Kind: domain.KindText},
			{Name: "gruppe", Kind: domain.KindText, Skript: "gruppe_von()"},
			{Name: "nummer", Kind: domain.KindText, Skript: "kartennummer()"},
		},
	}

	ausgabe, _, err := probelauf(service.Ausfuhrmittel{}).Fuehre(
		t.Context(), vorlage, dreiKartenUngeordnet(), nil, nil)
	if err != nil {
		t.Fatalf("führen: %v", err)
	}

	// Cäsar (aCäsar), Anton (bAnton), Berta (bBerta) — in dieser Reihenfolge.
	if ausgabe != "1;2;3;" {
		t.Errorf("Ausgabe %q", ausgabe)
	}
}

// **`set.cards` ist die Liste, und `set` durchlaufen gibt dieselben Karten.**
func TestSetCardsTraegtDieGerechnetenKarten(t *testing.T) {
	t.Parallel()

	ausgabe, _ := fuehreKartenspiel(t,
		`set.cards.1.nummer + "/" + length(set.cards) + "/" + set.cards.2.ganz`)

	if ausgabe != "1/3/Berta (Bonn)" {
		t.Errorf("Ausgabe %q", ausgabe)
	}
}

// **Ein Vorgabewert füllt, was die Karte nicht führt** — `default:` ist ein
// Skript, und `art` stand bisher leer da.
func TestDieAusfuhrFuelltVorgabewerteEin(t *testing.T) {
	t.Parallel()

	ausgabe, _ := fuehreKartenspiel(t, `to_string(for each card in cards do card.art + ";")`)

	if ausgabe != "Standard;Standard;Standard;" {
		t.Errorf("Ausgabe %q", ausgabe)
	}
}

// **Ein Wertskript, das scheitert, lässt den Wert stehen und meldet einen
// Befund** — einen, nicht einen je Karte (§8).
func TestEinGescheitertesWertskriptMeldetEinenBefund(t *testing.T) {
	t.Parallel()

	vorlage := kartenspielvorlage(`to_string(for each card in cards do card.ganz + ";")`)
	vorlage.Kartenfelder[2].Skript = "unbekannte_funktion()"

	karten := dreiKartenUngeordnet()
	karten[0].Values = append(karten[0].Values, domain.Value{Field: "ganz", Text: "alt"})

	ausgabe, befunde, err := probelauf(service.Ausfuhrmittel{}).Fuehre(
		t.Context(), vorlage, karten, nil, nil)
	if err != nil {
		t.Fatalf("führen: %v", err)
	}

	if ausgabe != "alt;;;" {
		t.Errorf("Ausgabe %q — erwartet den alten Wert und zwei leere", ausgabe)
	}

	gemeldet := 0

	for _, befund := range befunde {
		if befund.Ort == "ganz" {
			gemeldet++
		}
	}

	if gemeldet != 1 {
		t.Errorf("%d Befunde zum Feld »ganz«, erwartet 1: %v", gemeldet, befunde)
	}
}

// **Ohne erklärte Kartenfelder bleibt die Liste, wie sie war** — und `set` führt
// trotzdem die Ordnung: die Vorlage selbst kann die Karte im Set zählen.
func TestEineVorlageZaehltDieKarteImSetAuchOhneKartenfelder(t *testing.T) {
	t.Parallel()

	vorlage := msepaket.Exportvorlage{
		KurzerName: "probe",
		Skript: `to_string(for each card in cards do ` +
			`position(of: card, in: set, order_by: {card.name}) + ",")`,
	}

	ausgabe, _, err := probelauf(service.Ausfuhrmittel{}).Fuehre(
		t.Context(), vorlage, dreiKartenUngeordnet(), nil, nil)
	if err != nil {
		t.Fatalf("führen: %v", err)
	}

	if ausgabe != "2,0,1," {
		t.Errorf("Ausgabe %q", ausgabe)
	}
}

// **Ein Abbruch hält die Rechnung an.** Der Lauf endet mit dem Fehler des
// Abbruchs und nicht mit einer Ausgabe aus halb gerechneten Karten.
func TestEinAbbruchHaeltDieAusfuhrAn(t *testing.T) {
	t.Parallel()

	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	_, _, err := probelauf(service.Ausfuhrmittel{}).Fuehre(ctx,
		kartenspielvorlage(`to_string(for each card in cards do card.nummer)`),
		dreiKartenUngeordnet(), nil, nil)
	if err == nil {
		t.Error("kein Fehler nach dem Abbruch")
	}
}
