package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// setkarten ist ein Double der Kartenquelle (§3).
type setkarten struct {
	karten []*domain.CardRow
	fehler error

	// eigene sind die Packarten, die das Set selbst mitbringt (#1081).
	eigene []*domain.PackArt

	// packfehler scheitert **nur** beim Lesen der eigenen Packarten.
	//
	// Ein eigener Wert und nicht `fehler`: wer beides zusammenwirft, kommt gar
	// nicht bis zu den Packarten — das Kartenlesen scheiterte vorher, und der
	// Weg dahinter bliebe ungeprüft.
	packfehler error
}

func (s setkarten) Karten(context.Context, int64) ([]*domain.CardRow, error) {
	return s.karten, s.fehler
}

func (s setkarten) Setpackarten(context.Context, int64) ([]*domain.PackArt, error) {
	return s.eigene, s.packfehler
}

// packkarte baut eine Karte mit einer Seltenheit.
func packkarte(id int64, seltenheit string) *domain.CardRow {
	return &domain.CardRow{
		ID: id, SetID: 1,
		Values: []domain.Value{{Field: "rarity", Text: seltenheit}},
	}
}

// packvorlage gibt eine Vorlage mit zwei Packarten und einem Filter.
func packvorlage() service.Vorlage {
	return service.Vorlage{
		Packs: []msepaket.Pack{
			{Name: "common", Auswahl: "no replace", Filter: `card.rarity == "common"`},
			{
				Name: "booster", Auswahl: "all", Waehlbar: true,
				Posten: []msepaket.Packposten{{Name: "common", Anzahl: 2, Gewicht: 1}},
			},
		},
	}
}

// packbereichMit baut den Bereich über einer Vorlage.
func packbereichMit(karten []*domain.CardRow, vorlage service.Vorlage, da bool) *service.Packbereich {
	return service.NeuerPackbereich(nil, setkarten{karten: karten},
		service.NeuerPlanbauer().MitStilen(spielwoerter{vorlage, da}))
}

// **Das Angebot nennt die Packarten und zählt die passenden Karten** —
// #1072.
//
// **Die Zahl steht dabei, weil eine leere Packart nichts zieht** und niemand
// sieht, warum: ein Set ohne seltene Karten gibt kein Booster.
func TestDasPackangebotZaehltDiePassendenKarten(t *testing.T) {
	karten := []*domain.CardRow{
		packkarte(1, "common"), packkarte(2, "common"), packkarte(3, "rare"),
	}

	angebot, befunde, err := packbereichMit(karten, packvorlage(), true).
		Angebot(t.Context(), 1)
	if err != nil {
		t.Fatalf("Angebot: %v", err)
	}

	if len(befunde) != 0 {
		t.Errorf("Befunde: %+v", befunde)
	}

	if len(angebot) != 2 {
		t.Fatalf("%d Packarten: %+v", len(angebot), angebot)
	}

	// **Der Filter lässt zwei durch.**
	if angebot[0].Name != "common" || angebot[0].Karten != 2 || angebot[0].Waehlbar {
		t.Errorf("die erste ist %+v", angebot[0])
	}

	// **Booster hat keine eigenen Karten** — es zieht aus seinem Posten. Die
	// Zahl ist deshalb die **erreichbare**: dieselben zwei.
	if !angebot[1].Waehlbar || angebot[1].Karten != 2 {
		t.Errorf("die zweite ist %+v", angebot[1])
	}
}

// **Derselbe Same gibt dieselbe Ziehung, ein anderer eine andere.**
//
// Ohne die zweite Hälfte bestünde den Lauf auch ein Zieher, der immer dasselbe
// gibt.
func TestEinPackZiehtWiederholbar(t *testing.T) {
	karten := []*domain.CardRow{
		packkarte(1, "common"), packkarte(2, "common"),
		packkarte(3, "common"), packkarte(4, "common"),
	}

	bereich := packbereichMit(karten, packvorlage(), true)

	eines, _, err := bereich.Ziehen(t.Context(), 1, "booster", 42)
	if err != nil {
		t.Fatalf("Ziehen: %v", err)
	}

	if len(eines) != 2 {
		t.Fatalf("%d Karten gezogen", len(eines))
	}

	nochmal := mussZiehen(t, bereich, 42)
	if len(nochmal) != 2 || nochmal[0].ID != eines[0].ID || nochmal[1].ID != eines[1].ID {
		t.Errorf("derselbe Same gab %+v statt %+v", nochmal, eines)
	}

	// **Über mehrere Samen**, weil zwei Ziehungen aus vier Karten zufällig
	// gleich sein können — ein einzelner Vergleich wäre ein Wackelkandidat.
	andersHerum := false

	for same := uint64(1); same < 12 && !andersHerum; same++ {
		anderes := mussZiehen(t, bereich, same)
		if len(anderes) == 2 && anderes[0].ID != eines[0].ID {
			andersHerum = true
		}
	}

	if !andersHerum {
		t.Error("über elf Samen kam immer dasselbe heraus")
	}
}

// **Ein gescheiterter Filter lässt die Packart leer und bekommt einen Befund.**
//
// Nicht alle Karten: wer im Zweifel alles durchließe, zöge aus einem Ausdruck,
// den niemand ausrechnen konnte (§8).
func TestEinGescheiterterPackfilterGibtEinenBefund(t *testing.T) {
	vorlage := service.Vorlage{
		Packs: []msepaket.Pack{{Name: "kaputt", Filter: "gibtsnicht("}},
	}

	angebot, befunde, err := packbereichMit(
		[]*domain.CardRow{packkarte(1, "common")}, vorlage, true).Angebot(t.Context(), 1)
	if err != nil {
		t.Fatalf("Angebot: %v", err)
	}

	if len(angebot) != 1 || angebot[0].Karten != 0 {
		t.Errorf("das Angebot ist %+v", angebot)
	}

	if len(befunde) != 1 || befunde[0].Ort != "kaputt" {
		t.Errorf("die Befunde sind %+v", befunde)
	}
}

// **Ein Filter, der sich lesen lässt und beim Rechnen scheitert, ebenso.**
//
// # Warum das ein zweiter Lauf ist
//
// Bis zum 2026-08-29 waren es **ein** Zweig: `AuswertenQuelle` zerteilte und
// rechnete in einem Zug, und ein Test über `"gibtsnicht("` deckte beides. Seit
// die Zerteilung aus der Kartenschleife heraussteht
// (#1288),
// sind es zwei — und der zweite hatte keinen Test.
//
// **Der Unterschied ist für Nutzende sichtbar**: ein Filter, der sich nicht
// **lesen** lässt, ist an der Vorlage falsch und wird einmal gemeldet; einer,
// der beim **Rechnen** scheitert, kann an einer einzelnen Karte liegen. Beide
// lassen die Packart leer (§8) — aber die Meldung ist eine andere.
func TestEinPackfilterDerBeimRechnenScheitertGibtEinenBefund(t *testing.T) {
	vorlage := service.Vorlage{
		Packs: []msepaket.Pack{{Name: "rechnet nicht", Filter: `gibtsnicht()`}},
	}

	angebot, befunde, err := packbereichMit(
		[]*domain.CardRow{packkarte(1, "common")}, vorlage, true).Angebot(t.Context(), 1)
	if err != nil {
		t.Fatalf("Angebot: %v", err)
	}

	if len(angebot) != 1 || angebot[0].Karten != 0 {
		t.Errorf("das Angebot ist %+v — eine Packart, deren Filter scheitert, "+
			"lässt keine Karte durch", angebot)
	}

	if len(befunde) != 1 || befunde[0].Ort != "rechnet nicht" {
		t.Errorf("die Befunde sind %+v", befunde)
	}
}

// **Ohne Spielpaket gibt es keine Packarten** — und keinen Fehler.
func TestOhnePackpaketGibtEsKeineArten(t *testing.T) {
	angebot, _, err := packbereichMit(nil, service.Vorlage{}, false).Angebot(t.Context(), 1)
	if err != nil || len(angebot) != 0 {
		t.Errorf("%+v / %v", angebot, err)
	}
}

// **Ein Lesefehler kommt durch, und eine unbekannte Packart auch.**
func TestDerPackbereichMeldetSeineFehler(t *testing.T) {
	kaputt := service.NeuerPackbereich(nil,
		setkarten{fehler: errors.New("die Platte ist voll")}, service.NeuerPlanbauer())

	if _, _, err := kaputt.Angebot(t.Context(), 1); err == nil {
		t.Error("das Angebot kam durch")
	}

	if _, _, err := kaputt.Ziehen(t.Context(), 1, "booster", 1); err == nil {
		t.Error("das Ziehen kam durch")
	}

	bereich := packbereichMit(nil, packvorlage(), true)
	if _, _, err := bereich.Ziehen(t.Context(), 1, "gibtsnicht", 1); err == nil {
		t.Error("die unbekannte Packart kam durch")
	}
}

// **Ein abgebrochener Kontext hält den Filter an.**
//
// `magic.mse-game` führt 30 Packarten, und der Filter läuft über **jede**
// Karte des Sets — bei `Multicolor.mse-set` sind das 1.400. Wer das Set
// wechselt, während das läuft, wartet sonst darauf, dass die alte Ziehung
// fertig gerechnet wird (§7).
func TestEinAbgebrochenerKontextHaeltDenPackfilterAn(t *testing.T) {
	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	_, _, err := packbereichMit([]*domain.CardRow{packkarte(1, "common")},
		packvorlage(), true).Angebot(ctx, 1)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("der Abbruch kam als %v durch", err)
	}
}

// **Eine Packart ohne Filter hat keine eigenen Karten.**
//
//	// Filter cards
//	if (pack_type.filter) {
//	  FOR_EACH(card, parent.set->cards) { … }
//	}
//
// `src/data/pack.cpp`, Zeilen 110–117, Fassung f9d9356 vom 2021-07-06: allein
// der Filter füllt die Kartenliste. Ohne ihn bleibt sie leer, `total_weight`
// wird 0, und die Packart zieht nichts.
//
// **Der Rumpf dieses Kommentars stand bis zum 2026-08-25 andersherum** und
// nannte dieselbe Datei — nur die falschen Zeilen. Angestrichen vom Nachlauf
// des Zweitblicks über Phase W: eine Korrektur, die den Kommentar daneben
// stehen lässt, hinterlässt eine Fundstelle, die das Gegenteil belegt.
func TestEinePackartOhneFilterHatKeineEigenenKarten(t *testing.T) {
	vorlage := service.Vorlage{Packs: []msepaket.Pack{{Name: "alles"}}}

	angebot, _, err := packbereichMit(
		[]*domain.CardRow{packkarte(1, "common"), packkarte(2, "rare")},
		vorlage, true).Angebot(t.Context(), 1)
	if err != nil {
		t.Fatalf("Angebot: %v", err)
	}

	// **Null, nicht alle.** Hier stand bis zum 2026-08-24 `Karten != 2` und
	// der Name „nimmt jede Karte" — erschlossen, nicht nachgelesen. Bei MSE
	// füllt allein der Filter die Kartenliste (`src/data/pack.cpp`, Zeilen
	// 110–117, Fassung f9d9356 vom 2021-07-06); ohne ihn bleibt sie leer,
	// `total_weight` wird 0, und die Packart zieht nichts.
	//
	// Von 29 Packarten im echten Pack trägt **keine** weder Filter noch
	// Posten; die Abweichung war folgenlos und stand trotzdem als geprüfte
	// Zusage da.
	if len(angebot) != 1 || angebot[0].Karten != 0 {
		t.Errorf("das Angebot ist %+v", angebot)
	}
}

// **Ein Kreis wird abgeschnitten, und ein unbekannter Posten übergangen.**
//
// `Packzieher.staende` weist einen Kreis ab; hier ist die Frage nur, wie viele
// Karten erreichbar sind, und eine endlose Rekursion wäre die falsche Antwort
// darauf.
func TestDiePackzaehlungHaeltEinemKreisStand(t *testing.T) {
	vorlage := service.Vorlage{
		Packs: []msepaket.Pack{
			{Name: "eines", Auswahl: "all", Posten: []msepaket.Packposten{
				{Name: "anderes", Anzahl: 1}, {Name: "gibtsnicht", Anzahl: 1},
			}},
			{Name: "anderes", Auswahl: "all", Posten: []msepaket.Packposten{
				{Name: "eines", Anzahl: 1},
			}},
		},
	}

	angebot, _, err := packbereichMit(
		[]*domain.CardRow{packkarte(1, "common")}, vorlage, true).Angebot(t.Context(), 1)
	if err != nil {
		t.Fatalf("Angebot: %v", err)
	}

	if len(angebot) != 2 || angebot[0].Karten != 0 {
		t.Errorf("das Angebot ist %+v", angebot)
	}
}

// **Ohne Planbauer gibt es keine Packarten** — der Stand eines frisch
// gebauten Bereichs, und er darf nicht abstürzen.
func TestOhnePlanbauerGibtEsKeinePackarten(t *testing.T) {
	angebot, _, err := service.NeuerPackbereich(nil, setkarten{}, nil).
		Angebot(t.Context(), 1)
	if err != nil || len(angebot) != 0 {
		t.Errorf("%+v / %v", angebot, err)
	}
}

// **Ein Abbruch hält auch das Ziehen an** — nicht nur das Angebot.
func TestEinAbgebrochenerKontextHaeltDasZiehenAn(t *testing.T) {
	ctx, abbrechen := context.WithCancel(t.Context())
	abbrechen()

	_, _, err := packbereichMit([]*domain.CardRow{packkarte(1, "common")},
		packvorlage(), true).Ziehen(ctx, 1, "booster", 1)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("der Abbruch kam als %v durch", err)
	}
}

// mussZiehen zieht und bricht bei einem Fehler ab.
func mussZiehen(
	t *testing.T, bereich *service.Packbereich, same uint64,
) []*domain.CardRow {
	t.Helper()

	gezogen, _, err := bereich.Ziehen(t.Context(), 1, "booster", same)
	if err != nil {
		t.Fatalf("Ziehen: %v", err)
	}

	return gezogen
}
