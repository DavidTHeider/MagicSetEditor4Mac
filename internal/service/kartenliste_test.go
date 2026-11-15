// **Die Spalten der Kartenliste und ihre Ordnung** — Werkstück 3 von
// Phase 3.7
// (#1168).
package service_test

import (
	"context"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// listendoppel ist Kartenquelle, Feldquelle, Setleser und Ablage in einem.
type listendoppel struct {
	spiel         string
	spielskript   string
	setfelder     []domain.Field
	karten        []*domain.CardRow
	felder        []domain.Field
	wahl          []domain.Spaltenwahl
	ordnung       domain.Kartenordnung
	lesefehler    error
	kartenfehler  error
	feldfehler    error
	wahlfehler    error
	ordnungfehler error
	schreibfehler error

	geschrieben     []domain.Spaltenwahl
	fortgeschrieben *domain.Kartenordnung
}

func (l *listendoppel) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	return l.karten, l.kartenfehler
}

func (l *listendoppel) Kartenfelder(context.Context, int64) ([]domain.Field, error) {
	return l.felder, l.feldfehler
}

func (l *listendoppel) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	if l.lesefehler != nil {
		return nil, l.lesefehler
	}

	if l.spiel == "" {
		return nil, nil
	}

	return []domain.SetHead{{ID: 1, Name: "Probe", Game: l.spiel}}, nil
}

func (l *listendoppel) Spalten(context.Context, string) ([]domain.Spaltenwahl, error) {
	return l.wahl, l.wahlfehler
}

func (l *listendoppel) SpaltenSetzen(
	_ context.Context, _ string, wahl []domain.Spaltenwahl,
) error {
	if l.schreibfehler != nil {
		return l.schreibfehler
	}

	l.geschrieben = wahl

	return nil
}

// Spielskript gibt das Skript des Doppels — leer, wo keines gesetzt ist.
func (l *listendoppel) Spielskript(context.Context, int64) (string, error) {
	return l.spielskript, nil
}

func (l *listendoppel) Setfelder(context.Context, int64) ([]domain.Field, error) {
	return l.setfelder, nil
}

func (l *listendoppel) Sortierung(context.Context, string) (domain.Kartenordnung, error) {
	return l.ordnung, l.ordnungfehler
}

func (l *listendoppel) SortierungSetzen(
	_ context.Context, _ string, ordnung domain.Kartenordnung,
) error {
	if l.schreibfehler != nil {
		return l.schreibfehler
	}

	l.fortgeschrieben = &ordnung

	return nil
}

// dreiFelder gibt ein Doppel mit drei Karten und vier Feldern.
func dreiFelder() *listendoppel {
	return &listendoppel{
		spiel:   "magic",
		ordnung: domain.StandardKartenordnung(),
		karten: []*domain.CardRow{
			{ID: 1, Number: 0, Values: []domain.Value{
				{Field: "name", Text: "Karte 10"}, {Field: "typ", Text: "Zauber"},
			}},
			{ID: 2, Number: 1, Values: []domain.Value{
				{Field: "name", Text: "Karte 2"}, {Field: "typ", Text: "Artefakt"},
			}},
			{ID: 3, Number: 2, Values: []domain.Value{
				{Field: "name", Text: "<i>Karte 1</i>"},
			}},
		},
		felder: []domain.Field{
			listenfeld("name", "Name", 100, true, true, domain.ListeLinks),
			listenfeld("typ", "Typ", 200, false, true, domain.ListeRechts),
			listenfeld("geheim", "", 300, true, false, domain.ListeLinks),
			listenfeld("ohne stelle", "", 0, false, true, domain.ListeLinks),
		},
	}
}

// listenfeld baut ein Kartenfeld mit seinen Listenangaben.
//
// **Sechs Stellen und kein Verbund**: es ist eine Vorrichtung, sie steht
// viermal untereinander, und dort sagt die Reihenfolge mehr als sechs
// Feldnamen es täten.
func listenfeld(
	name, aufschrift string, stelle int, sichtbar, erlaubt bool,
	ausrichtung domain.Listenausrichtung,
) domain.Field {
	return domain.Field{
		Name:       name,
		Listenname: aufschrift,
		Listenangaben: domain.Listenangaben{
			Listenstelle:      stelle,
			Listensichtbar:    sichtbar,
			Listenerlaubt:     erlaubt,
			Listenausrichtung: ausrichtung,
		},
	}
}

func kartenliste(doppel *listendoppel) *service.Kartenlistenbau {
	return service.NeuerKartenlistenbau(doppel, doppel, doppel, doppel, doppel)
}

// **Ohne Wahl gilt der Vorschlag des Spielpakets.**
func TestOhneWahlGiltDerVorschlagDesPakets(t *testing.T) {
	t.Parallel()

	liste, fehler := kartenliste(dreiFelder()).Aufbau(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("Aufbau: %v", fehler)
	}

	// **Vier Felder, drei zur Wahl** — `geheim` ist verboten und steht gar
	// nicht erst da.
	if len(liste.Spalten) != 3 {
		t.Fatalf("%d Spalten zur Wahl: %+v", len(liste.Spalten), liste.Spalten)
	}

	if liste.Spalten[0].Feld != "name" || !liste.Spalten[0].Sichtbar ||
		liste.Spalten[0].Aufschrift != "Name" {
		t.Errorf("erste Spalte: %+v", liste.Spalten[0])
	}

	// **Nach der Stelle geordnet, und die ohne Stelle zuletzt.**
	if liste.Spalten[1].Feld != "typ" || liste.Spalten[2].Feld != "ohne stelle" {
		t.Errorf("Reihenfolge: %+v", liste.Spalten)
	}

	if liste.Spalten[1].Sichtbar {
		t.Error("»typ« steht ohne Angabe sichtbar da")
	}

	if liste.Spalten[1].Ausrichtung != domain.ListeRechts {
		t.Errorf("Ausrichtung %q", liste.Spalten[1].Ausrichtung)
	}
}

// **Eine getroffene Wahl schlägt den Vorschlag.**
func TestEineGetroffeneWahlSchlaegtDenVorschlag(t *testing.T) {
	t.Parallel()

	doppel := dreiFelder()
	doppel.wahl = []domain.Spaltenwahl{
		{Feld: "typ", Sichtbar: true, Stelle: 10},
		{Feld: "name", Sichtbar: false, Stelle: 20},
	}

	liste, fehler := kartenliste(doppel).Aufbau(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("Aufbau: %v", fehler)
	}

	if liste.Spalten[0].Feld != "typ" || !liste.Spalten[0].Sichtbar {
		t.Errorf("erste Spalte: %+v", liste.Spalten[0])
	}

	if liste.Spalten[1].Feld != "name" || liste.Spalten[1].Sichtbar {
		t.Errorf("zweite Spalte: %+v", liste.Spalten[1])
	}

	// **Ein Feld ohne Eintrag in der Wahl behält den Vorschlag** — und steht
	// hinten, damit eine neue Spalte des Spiels nicht mitten hineinrutscht.
	if liste.Spalten[2].Feld != "ohne stelle" {
		t.Errorf("dritte Spalte: %+v", liste.Spalten[2])
	}
}

// **Eine Wahl über ein verbotenes Feld wird nicht befolgt.**
//
//	if (cs.visible && f->card_list_allow) { … }
//
// `src/gui/control/card_list.cpp`, Zeile 218, Fassung f9d9356 vom 2021-07-06.
func TestEineWahlUeberEinVerbotenesFeldWirdNichtBefolgt(t *testing.T) {
	t.Parallel()

	doppel := dreiFelder()
	doppel.wahl = []domain.Spaltenwahl{{Feld: "geheim", Sichtbar: true, Stelle: 1}}

	liste, fehler := kartenliste(doppel).Aufbau(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("Aufbau: %v", fehler)
	}

	for _, spalte := range liste.Spalten {
		if spalte.Feld == "geheim" {
			t.Errorf("das verbotene Feld steht da: %+v", spalte)
		}
	}
}

// **Die Zeilen tragen den Wert jeder sichtbaren Spalte — ohne Auszeichnung.**
func TestDieZeilenTragenDieSichtbarenWerte(t *testing.T) {
	t.Parallel()

	doppel := dreiFelder()
	doppel.wahl = []domain.Spaltenwahl{
		{Feld: "name", Sichtbar: true, Stelle: 10},
		{Feld: "typ", Sichtbar: true, Stelle: 20},
	}

	liste, fehler := kartenliste(doppel).Aufbau(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("Aufbau: %v", fehler)
	}

	if len(liste.Zeilen) != 3 {
		t.Fatalf("%d Zeilen", len(liste.Zeilen))
	}

	// **Ohne Auszeichnung** — ein `<i>` in einer Tabellenzelle sieht wie ein
	// Fehler aus.
	dritte := liste.Zeilen[2]
	if dritte.Werte["name"] != "Karte 1" {
		t.Errorf("dritte Zeile: %+v", dritte.Werte)
	}

	// **Ein Feld, das die Karte nicht führt, ist leer und kein Fehler.**
	if wert, da := dritte.Werte["typ"]; !da || wert != "" {
		t.Errorf("»typ« der dritten Zeile: %q / %v", wert, da)
	}
}

// **Ohne Sortierung bleibt die Reihenfolge des Sets.**
//
// Das ist die Reihenfolge, in der die Karten in der `.mse-set` stehen — und
// die, die beim Speichern wieder herauskommt.
func TestOhneSortierungBleibtDieReihenfolgeDesSets(t *testing.T) {
	t.Parallel()

	liste, fehler := kartenliste(dreiFelder()).Aufbau(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("Aufbau: %v", fehler)
	}

	if liste.Zeilen[0].KarteID != 1 || liste.Zeilen[2].KarteID != 3 {
		t.Errorf("%+v", liste.Zeilen)
	}
}

// **Und mit Sortierung zählen Zahlen als Zahlen.**
//
// »Karte 10« steht hinter »Karte 2« — das ist der Grund, warum
// [domain.Vergleiche] portiert ist.
func TestMitSortierungZaehlenZahlenAlsZahlen(t *testing.T) {
	t.Parallel()

	doppel := dreiFelder()
	doppel.ordnung = domain.Kartenordnung{Feld: "name", Aufsteigend: true}

	liste, fehler := kartenliste(doppel).Aufbau(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("Aufbau: %v", fehler)
	}

	namen := make([]string, 0, len(liste.Zeilen))
	for _, zeile := range liste.Zeilen {
		namen = append(namen, zeile.Werte["name"])
	}

	if namen[0] != "Karte 1" || namen[1] != "Karte 2" || namen[2] != "Karte 10" {
		t.Errorf("%v", namen)
	}

	// **Absteigend kehrt sie um.**
	doppel.ordnung.Aufsteigend = false

	liste, fehler = kartenliste(doppel).Aufbau(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("Aufbau: %v", fehler)
	}

	if liste.Zeilen[0].Werte["name"] != "Karte 10" {
		t.Errorf("%+v", liste.Zeilen[0])
	}
}

// **Sortiert wird über den entmarkten Wert.**
//
// Sonst stünde `<i>Karte 1</i>` unter `<`, und jede kursive Karte läge vorn.
func TestSortiertWirdUeberDenEntmarktenWert(t *testing.T) {
	t.Parallel()

	doppel := dreiFelder()
	doppel.ordnung = domain.Kartenordnung{Feld: "name", Aufsteigend: true}

	liste, fehler := kartenliste(doppel).Aufbau(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("Aufbau: %v", fehler)
	}

	if liste.Zeilen[0].KarteID != 3 {
		t.Errorf("die kursive Karte steht an Stelle %d", liste.Zeilen[0].KarteID)
	}
}

// **Nach einem Feld, das es nicht gibt, wird nicht sortiert.**
//
// Ein Spielpaket kann sich ändern; eine gemerkte Sortierung nach einem Feld,
// das verschwunden ist, wäre sonst eine Liste, die nach nichts ordnet — und
// niemand sähe, warum.
func TestNachEinemUnbekanntenFeldWirdNichtSortiert(t *testing.T) {
	t.Parallel()

	doppel := dreiFelder()
	doppel.ordnung = domain.Kartenordnung{Feld: "gibtsnicht", Aufsteigend: true}

	liste, fehler := kartenliste(doppel).Aufbau(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("Aufbau: %v", fehler)
	}

	if liste.Zeilen[0].KarteID != 1 {
		t.Errorf("es wurde sortiert: %+v", liste.Zeilen)
	}

	// **Und die Antwort sagt es** — sonst bliebe ein Pfeil in einer
	// Spaltenüberschrift, die es nicht gibt.
	if liste.Ordnung.Feld != "" {
		t.Errorf("die Ordnung nennt %q", liste.Ordnung.Feld)
	}
}
