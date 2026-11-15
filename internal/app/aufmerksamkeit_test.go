package app_test

// **Aufmerksamkeit aus echten Befunden** — Werkstück U11 von Phase 5.1 (#1422).

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// befundbestand hat zwei Sets — eines mit Vorbehalten.
type befundbestand struct {
	koepfefehler error
	blockfehler  map[int64]error
}

func (b befundbestand) Setkoepfe(context.Context) ([]domain.SetHead, error) {
	return []domain.SetHead{{ID: 7, Name: "Mit Vorbehalt"}, {ID: 8, Name: "Sauber"}}, b.koepfefehler
}

func (befundbestand) Karten(context.Context, int64) ([]*domain.CardRow, error) { return nil, nil }

func (b befundbestand) Rohbloecke(_ context.Context, setID int64) ([]domain.RawBlock, error) {
	if fehler := b.blockfehler[setID]; fehler != nil {
		return nil, fehler
	}

	if setID == 7 {
		return []domain.RawBlock{{File: "set", Line: 12, Text: "unbekannt: x"}, {File: "set", Line: 40}}, nil
	}

	return nil, nil
}

func aufmerksameAnwendung(t *testing.T) (*app.App, *abgleichprobe, *paketfuehrer, *werkstattprobe) {
	t.Helper()

	abgleich := &abgleichprobe{stand: service.Habenstand{
		Zeilen:  []service.Habenzeile{{Bezeichnung: "Lightning Bolt", Gebraucht: 4, Vorhanden: 1, Fehlend: 3}},
		Fehlend: 3,
	}}
	pakete := &paketfuehrer{pakete: []domain.Paket{
		{Art: "mse-style", Name: "kaputt.mse-style"},
		{Art: "mse-game", Name: "magic.mse-game"},
	}}
	werkstatt := &werkstattprobe{texte: []service.Paketdateitext{
		{Name: "style", Inhalt: "card style:\n\tname:\n\t\tscript: to_upper(\n"},
		{Name: "extra", Inhalt: "mse version: 2.0.0\n"},
	}}

	anwendung := app.NeueApp("")
	anwendung.MitBestand(befundbestand{})
	anwendung.MitDecknaht(&deckprobe{deck: domain.Deck{ID: 3, Name: "Mono Rot"}})
	anwendung.MitDeckabgleich(abgleich)
	anwendung.MitPaketen(pakete)
	anwendung.MitPaketdateien(werkstatt)

	return anwendung, abgleich, pakete, werkstatt
}

func TestDieAufmerksamkeitSammeltAusDreiQuellen(t *testing.T) {
	t.Parallel()

	anwendung, _, _, _ := aufmerksameAnwendung(t)

	antwort := anwendung.Aufmerksamkeit(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	// **Vier**: ein Set, ein Deck, und beide Pakete — die Probe gibt jedem
	// Paket dieselben Texte, und in `style` steht der Befund.
	if len(antwort.Hinweise) != 4 {
		t.Fatalf("%d Hinweise: %+v", len(antwort.Hinweise), antwort.Hinweise)
	}

	set, deck, paket := antwort.Hinweise[0], antwort.Hinweise[1], antwort.Hinweise[2]

	if antwort.Hinweise[3].Art != "paketBefund" || antwort.Hinweise[3].Objekt.Name != "magic.mse-game" {
		t.Errorf("viertes: %+v", antwort.Hinweise[3])
	}

	if set.Art != "setVorbehalte" || set.Objekt.Kennung != "7" || set.Anzahl != 2 || !strings.Contains(set.Grund, "Zeile 12") {
		t.Errorf("Set: %+v", set)
	}

	if deck.Art != "deckFehlend" || deck.Objekt.Name != "Mono Rot" || deck.Anzahl != 3 || deck.Grund != "Lightning Bolt" {
		t.Errorf("Deck: %+v", deck)
	}

	// **Beide Pakete tragen dieselben Texte der Probe** — beide haben den Befund
	// in `style`; der Grund nennt Datei und Meldung.
	if paket.Art != "paketBefund" || paket.Anzahl != 1 || !strings.HasPrefix(paket.Grund, "style: ") {
		t.Errorf("Paket: %+v", paket)
	}

	if antwort.Durchgesehen.Sets != 2 || antwort.Durchgesehen.Decks != 1 || antwort.Durchgesehen.Pakete != 2 {
		t.Errorf("durchgesehen: %+v", antwort.Durchgesehen)
	}

	if len(antwort.Befunde) != 0 {
		t.Errorf("Befunde: %v", antwort.Befunde)
	}
}

// **Eine Quelle, die scheitert, hält die anderen nicht an** — sie steht als Befund da.
func TestEineQuelleDieScheitertBleibtEinBefund(t *testing.T) {
	t.Parallel()

	anwendung, abgleich, pakete, werkstatt := aufmerksameAnwendung(t)
	abgleich.fehler = errors.New("die Sammlung antwortet nicht")
	pakete.listenfehler = errors.New("das Paketverzeichnis fehlt")

	antwort := anwendung.Aufmerksamkeit(t.Context())
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Hinweise) != 1 || antwort.Hinweise[0].Art != "setVorbehalte" {
		t.Errorf("Hinweise: %+v", antwort.Hinweise)
	}

	if len(antwort.Befunde) != 2 || !strings.Contains(antwort.Befunde[0], "Mono Rot") ||
		!strings.Contains(antwort.Befunde[1], "Pakete") {
		t.Errorf("Befunde: %v", antwort.Befunde)
	}

	// **Ein Set, dessen Blöcke sich nicht lesen lassen**, ist ein Befund mit Namen.
	anwendung.MitBestand(befundbestand{blockfehler: map[int64]error{8: errors.New("kaputt")}})
	pakete.listenfehler = nil
	werkstatt.fehler = errors.New("nicht lesbar")

	antwort = anwendung.Aufmerksamkeit(t.Context())
	if len(antwort.Befunde) != 4 || !strings.Contains(antwort.Befunde[0], "Sauber") {
		t.Errorf("Befunde: %v", antwort.Befunde)
	}

	// **Ein mitgeliefertes Paket ist kein Befund** — die Werkstatt weist es mit
	// einem Validierungsfehler ab, und die Übersicht übergeht es.
	werkstatt.fehler = domain.ErrValidation

	antwort = anwendung.Aufmerksamkeit(t.Context())
	if antwort.Durchgesehen.Pakete != 0 || len(antwort.Befunde) != 2 {
		t.Errorf("mitgelieferte Pakete: %+v / %v", antwort.Durchgesehen, antwort.Befunde)
	}

	// **Und eine Deckliste, die sich nicht lesen lässt**, ist ein Befund ohne Namen —
	// es gibt keinen, denn keine Zeile kam an. Er steht hinter dem Setbefund,
	// denn die Quellen werden in der Reihenfolge Sets, Decks, Pakete gelesen.
	anwendung.MitDecknaht(&deckprobe{fehler: errors.New("Decks kaputt")})

	antwort = anwendung.Aufmerksamkeit(t.Context())
	if antwort.Durchgesehen.Decks != 0 || len(antwort.Befunde) != 2 || !strings.HasPrefix(antwort.Befunde[1], "Decks:") {
		t.Errorf("Deckliste kaputt: %+v / %v", antwort.Durchgesehen, antwort.Befunde)
	}
}

func TestOhneBestandGibtEsKeineAufmerksamkeit(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")

	if antwort := anwendung.Aufmerksamkeit(t.Context()); antwort.Fehler == nil || antwort.Hinweise == nil {
		t.Errorf("ohne Bestand: %+v", antwort)
	}

	// **Ohne Decks und Pakete zählt nur der Bestand** — die Nähte fehlen in Tests oft.
	anwendung.MitBestand(befundbestand{koepfefehler: errors.New("kein Profil")})

	antwort := anwendung.Aufmerksamkeit(t.Context())
	if antwort.Fehler != nil || len(antwort.Befunde) != 1 || antwort.Durchgesehen.Decks != 0 {
		t.Errorf("nur Bestand, kaputt: %+v", antwort)
	}

	// **Ein Deck ohne Fehlkarten und ein Paket ohne Befund sind keine Hinweise.**
	// Und gefragt wird die Werkstatt **über den Ort** des Pakets — die Saat im
	// Bündel hieße über ihren Namen „nicht gefunden" (Sichtprüfung 2026-09-09).
	werkstatt := &werkstattprobe{texte: []service.Paketdateitext{{Name: "style", Inhalt: "mse version: 2.0.0\n"}}}
	anwendung.MitBestand(befundbestand{})
	anwendung.MitDecknaht(&deckprobe{deck: domain.Deck{ID: 3, Name: "Voll"}})
	anwendung.MitDeckabgleich(&abgleichprobe{})
	anwendung.MitPaketen(&paketfuehrer{pakete: []domain.Paket{{Name: "sauber.mse-style", Pfad: "/pakete/sauber.mse-style"}}})
	anwendung.MitPaketdateien(werkstatt)

	antwort = anwendung.Aufmerksamkeit(t.Context())
	if len(antwort.Hinweise) != 1 || antwort.Hinweise[0].Art != "setVorbehalte" {
		t.Errorf("Hinweise: %+v", antwort.Hinweise)
	}

	if werkstatt.paket != "/pakete/sauber.mse-style" {
		t.Errorf("die Werkstatt wurde über %q gefragt, nicht über den Ort", werkstatt.paket)
	}

	// Ohne Ort bleibt der Name — für ein Paket, das nur die Liste kennt.
	anwendung.MitPaketen(&paketfuehrer{pakete: []domain.Paket{{Name: "ohne-ort.mse-style"}}})
	anwendung.Aufmerksamkeit(t.Context())

	if werkstatt.paket != "ohne-ort.mse-style" {
		t.Errorf("ohne Ort: %q", werkstatt.paket)
	}

	if gebunden := app.NeueBindungen(anwendung).Aufmerksamkeit(t.Context()); gebunden.Fehler != nil {
		t.Errorf("Fensterbindung: %+v", gebunden.Fehler)
	}

	// **Fehlend ohne Zeile, die es nennt** — der Grund bleibt leer, die Zahl
	// steht. Und zwei Befunde in zwei Dateien zählen zwei, der Grund ist der erste.
	anwendung.MitDeckabgleich(&abgleichprobe{stand: service.Habenstand{Fehlend: 2, Unbekannt: 2}})
	anwendung.MitPaketdateien(&werkstattprobe{texte: []service.Paketdateitext{
		{Name: "style", Inhalt: "card style:\n\tname:\n\t\tscript: to_upper(\n"},
		{Name: "extra", Inhalt: "card style:\n\tname:\n\t\tscript: (\n"},
	}})

	antwort = anwendung.Aufmerksamkeit(t.Context())
	if len(antwort.Hinweise) != 3 || antwort.Hinweise[1].Grund != "" || antwort.Hinweise[1].Anzahl != 2 ||
		antwort.Hinweise[2].Anzahl != 2 || !strings.HasPrefix(antwort.Hinweise[2].Grund, "style: ") {
		t.Errorf("Hinweise: %+v", antwort.Hinweise)
	}
}

// **Der Grund trägt seinen Schlüssel bis ins Fenster** — LOC
// (#1963).
//
// Bis dahin kam nur der deutsche Satz an, und das Fenster zeigte ihn auch in
// englischer Sprache. Jetzt reisen Schlüssel, Werte und der Ort daneben; der
// Satz bleibt als Rückfall (§9).
func TestDerGrundTraegtSeinenSchluessel(t *testing.T) {
	t.Parallel()

	anwendung, _, _, _ := aufmerksameAnwendung(t)

	antwort := anwendung.Aufmerksamkeit(t.Context())
	if antwort.Fehler != nil || len(antwort.Hinweise) != 4 {
		t.Fatalf("%+v", antwort)
	}

	set, deck, paket := antwort.Hinweise[0], antwort.Hinweise[1], antwort.Hinweise[2]

	// **Der Set: Datei und Zeile, und „Zeile" ist ein Wort des Katalogs.**
	if set.Schluessel != domain.GrundVorbehalt || set.Ort != "" ||
		strings.Join(set.Werte, "|") != "set|12" {
		t.Errorf("Set: %+v", set)
	}

	// **Das Paket: der Befund des Zerteilers**, mit seinem Schlüssel und
	// seinen Werten — dieselben, die die Prüfung der Paketdatei gibt — und
	// die Datei als Ort davor.
	erwartet := service.PaketSkriptfehler("card style:\n\tname:\n\t\tscript: to_upper(\n")
	if erwartet == nil || erwartet.Schluessel == "" {
		t.Fatalf("die Probe trägt keinen Befund mit Schlüssel: %+v", erwartet)
	}

	if paket.Schluessel != string(erwartet.Schluessel) || paket.Ort != "style" ||
		strings.Join(paket.Werte, "|") != strings.Join(erwartet.Werte, "|") ||
		paket.Werte == nil || paket.Grund != "style: "+erwartet.Meldung {
		t.Errorf("Paket: %+v, erwartet %+v", paket, erwartet)
	}

	// **Das Deck: eine Karte ist kein Satz** — kein Schlüssel, und die Werte
	// sind eine leere Liste und nicht `null`.
	if deck.Schluessel != "" || deck.Ort != "" || deck.Werte == nil || len(deck.Werte) != 0 {
		t.Errorf("Deck: %+v", deck)
	}
}
