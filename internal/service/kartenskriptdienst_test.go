package service_test

// **Ein Kartenskript ausführen** — Werkstück B von Phase 4.11
// (#1337).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// spielprobe gibt ein Spiel mit zwei Kartenskripten.
type spielprobe struct {
	skripte []service.Kartenskriptangabe
	felder  []domain.Field
	skript  string
	fehler  error
}

func (s *spielprobe) Kartenskripte(_ context.Context, _ int64) (
	[]service.Kartenskriptangabe, []domain.Field, string, error,
) {
	return s.skripte, s.felder, s.skript, s.fehler
}

// anlegerprobe merkt sich, was angelegt wurde.
type anlegerprobe struct {
	rufe   int
	karten [][]domain.Feldstand
	fehler error
}

func (a *anlegerprobe) KartenAnlegen(
	_ context.Context, _ int64, karten [][]domain.Feldstand,
) ([]int64, error) {
	a.rufe++
	a.karten = karten

	if a.fehler != nil {
		return nil, a.fehler
	}

	kennungen := make([]int64, 0, len(karten))
	for i := range karten {
		kennungen = append(kennungen, int64(i+1))
	}

	return kennungen, nil
}

func probespiel() *spielprobe {
	return &spielprobe{
		skripte: []service.Kartenskriptangabe{
			{
				Name:    "Basic Lands",
				Kuerzel: "B",
				Skript:  `[ new_card([name: "Plains"]), new_card([name: "Island"]) ]`,
			},
			{Name: "Leerer Zug", Skript: `[]`},
		},
		felder: probefelder(),
	}
}

// **Ein Lauf legt die Karten des Skripts an — in einem Ruf.**
func TestEinKartenskriptLegtSeineKartenAn(t *testing.T) {
	t.Parallel()

	anleger := &anlegerprobe{}

	lauf, fehler := service.NeuerKartenskriptdienst(anleger, probespiel()).
		Ausfuehren(t.Context(), 1, "Basic Lands")
	if fehler != nil {
		t.Fatalf("Ausfuehren: %v", fehler)
	}

	// **Ein Ruf und nicht zwei.** Daran hängt die Zusage „ein ⌘Z holt sie
	// zusammen zurück" — zwei Rufe wären zwei Verlaufsschritte.
	if anleger.rufe != 1 {
		t.Errorf("%d Rufe an die Ablage, erwartet war einer", anleger.rufe)
	}

	if len(lauf.Karten) != 2 {
		t.Fatalf("%d Karten angelegt, erwartet waren zwei", len(lauf.Karten))
	}

	if len(anleger.karten) != 2 ||
		anleger.karten[0][0].Wert != "Plains" ||
		anleger.karten[1][0].Wert != "Island" {
		t.Errorf("angelegt wurde %+v", anleger.karten)
	}
}

// **Ein Skript, das nichts gibt, legt nichts an.**
//
// So auch MSE2: `if (!cards.empty())`. Ein leerer Zug im Verlauf wäre ein ⌘Z,
// das nichts tut.
func TestEinLeeresKartenskriptLegtNichtsAn(t *testing.T) {
	t.Parallel()

	anleger := &anlegerprobe{}

	lauf, fehler := service.NeuerKartenskriptdienst(anleger, probespiel()).
		Ausfuehren(t.Context(), 1, "Leerer Zug")
	if fehler != nil {
		t.Fatalf("Ausfuehren: %v", fehler)
	}

	if anleger.rufe != 0 {
		t.Errorf("die Ablage wurde %d-mal gerufen", anleger.rufe)
	}

	if len(lauf.Karten) != 0 {
		t.Errorf("%d Karten angelegt", len(lauf.Karten))
	}
}

// **Ein Skript, das es nicht gibt, ist ein Nicht-Gefunden.**
func TestEinUnbekanntesKartenskript(t *testing.T) {
	t.Parallel()

	_, fehler := service.NeuerKartenskriptdienst(&anlegerprobe{}, probespiel()).
		Ausfuehren(t.Context(), 1, "Gibtsnicht")

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet war ErrNotFound", fehler)
	}
}

// **Ein Skript, das scheitert, legt nichts an** (§7, Regel 3).
func TestEinGescheitertesKartenskriptLegtNichtsAn(t *testing.T) {
	t.Parallel()

	spiel := probespiel()
	spiel.skripte[0].Skript = `[ new_card([gibtsnicht: "x"]) ]`

	anleger := &anlegerprobe{}

	if _, fehler := service.NeuerKartenskriptdienst(anleger, spiel).
		Ausfuehren(t.Context(), 1, "Basic Lands"); fehler == nil {
		t.Fatal("das kaputte Skript ging durch")
	}

	if anleger.rufe != 0 {
		t.Errorf("die Ablage wurde trotzdem gerufen")
	}
}

// **Ein Fehler der Ablage kommt durch.**
func TestEinFehlerBeimAnlegenKommtDurch(t *testing.T) {
	t.Parallel()

	eigener := errors.New("die Platte ist voll")

	_, fehler := service.NeuerKartenskriptdienst(
		&anlegerprobe{fehler: eigener}, probespiel()).
		Ausfuehren(t.Context(), 1, "Basic Lands")

	if !errors.Is(fehler, eigener) {
		t.Errorf("Fehler %v, erwartet war der eigene", fehler)
	}
}

// **Und ein Fehler der Spielauskunft ebenso.**
func TestEinFehlerDerSpielauskunftKommtDurch(t *testing.T) {
	t.Parallel()

	eigener := errors.New("kein Spielpaket")
	spiel := &spielprobe{fehler: eigener}

	dienst := service.NeuerKartenskriptdienst(&anlegerprobe{}, spiel)

	if _, fehler := dienst.Ausfuehren(t.Context(), 1, "Basic Lands"); !errors.Is(
		fehler, eigener) {
		t.Errorf("Ausfuehren gab %v", fehler)
	}

	if _, fehler := dienst.Angebot(t.Context(), 1); !errors.Is(fehler, eigener) {
		t.Errorf("Angebot gab %v", fehler)
	}
}

// **Das Angebot nennt die Skripte, ohne sie auszuführen.**
func TestDasAngebotNenntDieSkripte(t *testing.T) {
	t.Parallel()

	anleger := &anlegerprobe{}

	skripte, fehler := service.NeuerKartenskriptdienst(anleger, probespiel()).
		Angebot(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("Angebot: %v", fehler)
	}

	if len(skripte) != 2 || skripte[0].Name != "Basic Lands" {
		t.Errorf("das Angebot lautet %+v", skripte)
	}

	if anleger.rufe != 0 {
		t.Error("das Angebot hat etwas angelegt")
	}
}

// **Ohne Skripte ist das Angebot leer und nicht nil.**
func TestEinAngebotOhneSkripteIstLeer(t *testing.T) {
	t.Parallel()

	skripte, fehler := service.NeuerKartenskriptdienst(
		&anlegerprobe{}, &spielprobe{}).Angebot(t.Context(), 1)
	if fehler != nil {
		t.Fatalf("Angebot: %v", fehler)
	}

	if skripte == nil {
		t.Error("das Angebot ist nil")
	}
}

// **Die Auskunft findet die Skripte des Spielpakets.**
//
// Gegen die **Saat** und nicht gegen ein bezogenes Paket (§11): `probe.mse-game`
// bringt seit #1337 zwei `add cards script:`-Blöcke mit, dieselbe Gestalt wie
// `magic.mse-game`, eigener Inhalt.
func TestDieVorlagenauskunftFindetDieSkripte(t *testing.T) {
	t.Parallel()

	auskunft := service.NeueVorlagenauskunft(service.NeuerStilspeicher(
		&koepfequelle{koepfe: []domain.SetHead{
			{ID: 7, Style: "probe-schlicht", Game: "probe"},
		}}, saatwurzel()))

	skripte, felder, spielskript, fehler := auskunft.Kartenskripte(t.Context(), 7)
	if fehler != nil {
		t.Fatalf("Kartenskripte: %v", fehler)
	}

	if len(skripte) != 2 || skripte[0].Name != "Zwei Farben" ||
		skripte[0].Kuerzel != "Z" {
		t.Fatalf("die Skripte lauten %+v", skripte)
	}

	// **Die Felder kommen mit** — ohne sie fiele ein unbekannter Feldname
	// nicht auf.
	if len(felder) == 0 {
		t.Error("keine Kartenfelder")
	}

	_ = spielskript

	// **Und der Lauf legt wirklich zwei Karten an.** Ohne diese Hälfte wäre
	// der Lauf auch dann grün, wenn die Auskunft ein Skript gäbe, das
	// scheitert.
	anleger := &anlegerprobe{}

	lauf, laufFehler := service.NeuerKartenskriptdienst(anleger, auskunft).
		Ausfuehren(t.Context(), 7, "Zwei Farben")
	if laufFehler != nil {
		t.Fatalf("Ausfuehren: %v", laufFehler)
	}

	if len(lauf.Karten) != 2 {
		t.Errorf("%d Karten angelegt", len(lauf.Karten))
	}
}

// **Ein Set ohne Vorlage gibt keine Skripte — und keinen Fehler.**
//
// Ein Set, dessen Spielpaket fehlt, hat keine Kartenskripte; ein Menü mit null
// Einträgen ist die richtige Antwort darauf, ein Abbruch nicht (§8).
func TestEinSetOhneVorlageHatKeineSkripte(t *testing.T) {
	t.Parallel()

	auskunft := service.NeueVorlagenauskunft(service.NeuerStilspeicher(
		&koepfequelle{koepfe: []domain.SetHead{
			{ID: 7, Style: "gibtsnicht", Game: "gibtsnicht"},
		}}, saatwurzel()))

	skripte, _, _, fehler := auskunft.Kartenskripte(t.Context(), 7)
	if fehler != nil {
		t.Fatalf("Kartenskripte: %v", fehler)
	}

	if len(skripte) != 0 {
		t.Errorf("die Skripte lauten %+v", skripte)
	}
}
