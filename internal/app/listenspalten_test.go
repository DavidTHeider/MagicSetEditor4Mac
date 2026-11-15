package app_test

// **Die Gestalt der Kartenliste über die Naht** — Werkstück 3 von Phase 3.7
// (#1168).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// listenquelle ist ein Doppel des Anwendungsfalls.
type listenquelle struct {
	fehler       error
	geschrieben  []domain.Spaltenwahl
	geordnet     *domain.Kartenordnung
	gefragtesSet int64

	// gefiltert ist, was `Filtern` gibt; gefragteAnfrage, wonach gefragt wurde.
	gefiltert       []int64
	gefragteAnfrage string
}

func (l *listenquelle) Filtern(_ context.Context, setID int64, anfrage string) ([]int64, error) {
	l.gefragtesSet, l.gefragteAnfrage = setID, anfrage

	if l.fehler != nil {
		return nil, l.fehler
	}

	return l.gefiltert, nil
}

func (l *listenquelle) Aufbau(_ context.Context, setID int64) (
	domain.Kartenlistenaufbau, error,
) {
	l.gefragtesSet = setID

	if l.fehler != nil {
		return domain.Kartenlistenaufbau{}, l.fehler
	}

	return domain.Kartenlistenaufbau{
		Spalten: []domain.Listenspalte{
			{
				Feld: "name", Aufschrift: "Name", Sichtbar: true,
				Ausrichtung: domain.ListeLinks,
			},
			{Feld: "rarity", Sichtbar: false, Ausrichtung: domain.ListeRechts},
		},
		Zeilen: []domain.Listenzeile{
			{KarteID: 7, Nummer: 2, Werte: map[string]string{"name": "Feuerball"}},
		},
		Ordnung: domain.Kartenordnung{Feld: "name", Aufsteigend: true},
	}, nil
}

func (l *listenquelle) SpaltenSetzen(
	_ context.Context, _ int64, wahl []domain.Spaltenwahl,
) error {
	if l.fehler != nil {
		return l.fehler
	}

	l.geschrieben = wahl

	return nil
}

func (l *listenquelle) SortierungSetzen(
	_ context.Context, _ int64, ordnung domain.Kartenordnung,
) error {
	if l.fehler != nil {
		return l.fehler
	}

	l.geordnet = &ordnung

	return nil
}

// mitListenbau baut eine Anwendung mit dem Doppel.
func mitListenbau(quelle app.Listenquelle) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitKartenliste(quelle)

	return anwendung
}

// **Der Aufbau kommt in der Gestalt an, die das Fenster erwartet.**
func TestDerListenaufbauKommtInSeinerGestaltAn(t *testing.T) {
	t.Parallel()

	quelle := &listenquelle{}

	antwort := mitListenbau(quelle).Kartenlistenaufbau(probekontext(t), "3")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if quelle.gefragtesSet != 3 {
		t.Errorf("gefragt wurde nach Set %d", quelle.gefragtesSet)
	}

	if len(antwort.Spalten) != 2 || antwort.Spalten[0].Feld != "name" ||
		antwort.Spalten[0].Aufschrift != "Name" || !antwort.Spalten[0].Sichtbar ||
		antwort.Spalten[1].Ausrichtung != "rechts" {
		t.Errorf("Spalten: %+v", antwort.Spalten)
	}

	// **Die Nummer ist eins-basiert**, wie die Liste zählt.
	if len(antwort.Zeilen) != 1 || antwort.Zeilen[0].KarteID != "7" ||
		antwort.Zeilen[0].Nummer != 3 {
		t.Errorf("Zeilen: %+v", antwort.Zeilen)
	}

	if antwort.Sortierfeld != "name" || !antwort.Aufsteigend {
		t.Errorf("Ordnung: %q / %v", antwort.Sortierfeld, antwort.Aufsteigend)
	}
}

// **Die Stelle kommt aus der Reihenfolge der Liste.**
//
// Eine Stellenangabe daneben wäre eine zweite Antwort auf dieselbe Frage.
func TestDieStelleKommtAusDerReihenfolgeDerListe(t *testing.T) {
	t.Parallel()

	quelle := &listenquelle{}

	antwort := mitListenbau(quelle).ListenspaltenSetzen(probekontext(t), "1",
		[]app.ListenspalteDTO{
			{Feld: "rarity", Sichtbar: true},
			{Feld: "name", Sichtbar: false},
		})
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(quelle.geschrieben) != 2 ||
		quelle.geschrieben[0].Feld != "rarity" || quelle.geschrieben[0].Stelle != 1 ||
		quelle.geschrieben[1].Feld != "name" || quelle.geschrieben[1].Stelle != 2 {
		t.Errorf("geschrieben wurde %+v", quelle.geschrieben)
	}
}

// **Und die Ordnung geht durch — leeres Feld eingeschlossen.**
func TestDieOrdnungGehtDurch(t *testing.T) {
	t.Parallel()

	quelle := &listenquelle{}

	if antwort := mitListenbau(quelle).ListensortierungSetzen(
		probekontext(t), "1", "", true); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if quelle.geordnet == nil || quelle.geordnet.Feld != "" ||
		!quelle.geordnet.Aufsteigend {
		t.Errorf("geordnet wurde %+v", quelle.geordnet)
	}
}

// **Eine unbrauchbare Kennung erreicht die Naht nicht.**
func TestEineUnbrauchbareKennungErreichtDieListeNicht(t *testing.T) {
	t.Parallel()

	quelle := &listenquelle{}
	anwendung := mitListenbau(quelle)

	antwort := anwendung.Kartenlistenaufbau(t.Context(), "abc")
	if antwort.Fehler == nil || antwort.Fehler.Feld != "setID" {
		t.Fatalf("%+v", antwort.Fehler)
	}

	// **`[]` und nicht `null`** — eine leere Liste, die als `null` ankommt,
	// kostet jeden Leser im Fenster einen Zweig (§3).
	if antwort.Spalten == nil || antwort.Zeilen == nil {
		t.Error("die Antwort trägt keine Listen")
	}

	if befund := anwendung.ListenspaltenSetzen(
		probekontext(t), "abc", nil).Fehler; befund == nil {
		t.Error("die Spaltenwahl kam mit einer unbrauchbaren Kennung durch")
	}

	if befund := anwendung.ListensortierungSetzen(
		probekontext(t), "abc", "", true).Fehler; befund == nil {
		t.Error("die Ordnung kam mit einer unbrauchbaren Kennung durch")
	}

	if quelle.gefragtesSet != 0 {
		t.Error("die Naht wurde trotzdem gefragt")
	}
}

// **Ohne Naht sagen es alle drei.**
func TestOhneListennahtSagenEsAlleDrei(t *testing.T) {
	t.Parallel()

	anwendung := app.NeueApp("")

	if antwort := anwendung.Kartenlistenaufbau(t.Context(), "1"); antwort.Fehler == nil {
		t.Error("der Aufbau kam ohne Naht durch")
	}

	if antwort := anwendung.ListenspaltenSetzen(
		probekontext(t), "1", nil); antwort.Fehler == nil {
		t.Error("die Spaltenwahl kam ohne Naht durch")
	}

	if antwort := anwendung.ListensortierungSetzen(
		probekontext(t), "1", "", true); antwort.Fehler == nil {
		t.Error("die Ordnung kam ohne Naht durch")
	}
}

// **Ein Fehler des Dienstes wird an allen dreien übersetzt.**
func TestEinListenfehlerWirdUebersetzt(t *testing.T) {
	t.Parallel()

	anwendung := mitListenbau(&listenquelle{fehler: domain.ErrBelegt})

	if antwort := anwendung.Kartenlistenaufbau(t.Context(), "1"); antwort.Fehler == nil {
		t.Error("der Aufbau kam durch")
	}

	if antwort := anwendung.ListenspaltenSetzen(
		probekontext(t), "1", nil); antwort.Fehler == nil {
		t.Error("die Spaltenwahl kam durch")
	}

	if antwort := anwendung.ListensortierungSetzen(
		probekontext(t), "1", "", true); antwort.Fehler == nil {
		t.Error("die Ordnung kam durch")
	}
}

// **Die Naht für die Flächen ohne Fenster trägt alle drei Wege durch.**
func TestDieListennahtTraegtAlleDreiWege(t *testing.T) {
	t.Parallel()

	quelle := &listenquelle{}
	naht := (&app.Fremdliste{}).MitListenquelle(quelle)

	aufbau, fehler := naht.Aufbau(probekontext(t), 1)
	if fehler != nil || len(aufbau.Spalten) != 2 {
		t.Fatalf("Aufbau: %+v / %v", aufbau, fehler)
	}

	if err := naht.SpaltenSetzen(probekontext(t), 1,
		[]domain.Spaltenwahl{{Feld: "name", Sichtbar: true}}); err != nil {
		t.Fatalf("Spalten: %v", err)
	}

	if err := naht.SortierungSetzen(probekontext(t), 1,
		domain.Kartenordnung{Feld: "name"}); err != nil {
		t.Fatalf("Sortierung: %v", err)
	}

	if len(quelle.geschrieben) != 1 || quelle.geordnet == nil {
		t.Errorf("geschrieben %+v, geordnet %+v", quelle.geschrieben, quelle.geordnet)
	}
}

// **Und ein Fehler kommt an allen dreien durch.**
func TestDieListennahtReichtIhrenFehlerDurch(t *testing.T) {
	t.Parallel()

	naht := (&app.Fremdliste{}).MitListenquelle(&listenquelle{fehler: domain.ErrBelegt})

	if _, fehler := naht.Aufbau(probekontext(t), 1); !errors.Is(fehler, domain.ErrBelegt) {
		t.Errorf("Aufbau: %v", fehler)
	}

	if fehler := naht.SpaltenSetzen(probekontext(t), 1, nil); !errors.Is(
		fehler, domain.ErrBelegt) {
		t.Errorf("Spalten: %v", fehler)
	}

	if fehler := naht.SortierungSetzen(probekontext(t), 1,
		domain.Kartenordnung{}); !errors.Is(fehler, domain.ErrBelegt) {
		t.Errorf("Sortierung: %v", fehler)
	}
}

// **Ohne Benutzerverzeichnis wird gemeldet statt geraten.**
func TestDieListenflaechennahtMeldetEinenFehlendenOrt(t *testing.T) {
	t.Setenv("HOME", "")

	if _, fehler := app.Listenflaechennaht(nil); fehler == nil {
		t.Error("ohne Zuhause kam eine Naht")
	}
}

// **Und mit Zuhause kommt sie zusammen und trägt über einen echten Bestand.**
func TestDieListenflaechennahtTraegtUeberEinenEchtenBestand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	setID, _ := saatsetImBestand(t, db)

	naht, nahtfehler := app.Listenflaechennaht(db)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	aufbau, fehler := naht.Aufbau(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("Aufbau: %v", fehler)
	}

	if len(aufbau.Spalten) == 0 {
		t.Fatal("das Saatspiel bietet keine Spalte an")
	}

	// **Die Wahl überlebt das Set** — sie gehört dem Spiel, und ein zweiter
	// Aufbau sieht sie wieder.
	wahl := []domain.Spaltenwahl{{Feld: aufbau.Spalten[0].Feld, Sichtbar: true, Stelle: 1}}
	if err := naht.SpaltenSetzen(probekontext(t), setID, wahl); err != nil {
		t.Fatalf("Spalten: %v", err)
	}

	danach, fehler := naht.Aufbau(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("zweiter Aufbau: %v", fehler)
	}

	if !danach.Spalten[0].Sichtbar {
		t.Errorf("die Wahl kam nicht zurück: %+v", danach.Spalten)
	}
}

// **Und die Marken kommen an den Zeilen mit** (#1376).
//
// Dieselbe Klasse wie die Deckplatznamen: `Kartenlistenbau` hat mit
// `MitMarken` eine Naht, und sie hing nur an der Wurzel des Fensters
// ([start.go]). `ListenzeileDTO.Marken` gibt es an **jeder** Fläche — über
// Kommandozeile, MCP und REST kam das Feld leer zurück, gleich wie viele
// Marken an der Karte hingen. Eine Fläche, die dasselbe Feld anders
// beantwortet, ist die Lücke, gegen die §16 steht.
func TestDieFlaechenlisteTraegtDieMarkenMit(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	setID, karteID := saatsetImBestand(t, db)

	if fehler := storage.NeueMarken(db).Setzen(
		probekontext(t), karteID, "prüfen"); fehler != nil {
		t.Fatalf("die Marke setzen: %v", fehler)
	}

	naht, nahtfehler := app.Listenflaechennaht(db)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	aufbau, fehler := naht.Aufbau(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("Aufbau: %v", fehler)
	}

	if len(aufbau.Zeilen) != 1 {
		t.Fatalf("%d Zeilen, erwartet war eine", len(aufbau.Zeilen))
	}

	if len(aufbau.Zeilen[0].Marken) != 1 || aufbau.Zeilen[0].Marken[0] != "prüfen" {
		t.Errorf("die Zeile trägt %v, erwartet war [prüfen]", aufbau.Zeilen[0].Marken)
	}
}
