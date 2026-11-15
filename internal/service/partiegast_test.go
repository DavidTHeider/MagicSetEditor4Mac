package service_test

// **Der Gast am anderen Ende der Leitung** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// **Dass ein Gast wünscht und entgegennimmt — und nichts erklärt.** Wer aus
// einem Befehl selbst Ereignisse machte, hätte eine zweite Wahrheit neben der
// des Wirts, und beim ersten Unterschied wüsste niemand, welche gilt.
//
// **Die Spiegelung von `partiewirt_test.go`**: dort wird ein Ereignis von
// außen abgewiesen, hier ein Wunsch. Beide Male ist es keine Formsache,
// sondern die Aufteilung selbst.

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// nehmerdoppelGast sammelt, was der Wirt für wahr erklärt hat.
type nehmerdoppelGast struct {
	bekommen []domain.Ereignis
	bilder   []domain.Tischbild
	klagt    error
}

func (n *nehmerdoppelGast) Ereignis(
	_ context.Context, ereignis domain.Ereignis,
) error {
	n.bekommen = append(n.bekommen, ereignis)

	return n.klagt
}

func (n *nehmerdoppelGast) Tischbild(
	_ context.Context, bild domain.Tischbild,
) error {
	n.bilder = append(n.bilder, bild)

	return n.klagt
}

// einGast baut ihn über einem Nahtdoppel.
func einGast(nehmer service.Ereignisnehmer) (*service.Partiegast, *nahtdoppel) {
	naht := &nahtdoppel{}
	sitzung := service.NeuePartiesitzung(
		"p1", "gast", naht, service.NeueReihenfolge(1))

	return service.NeuerPartiegast("p1", sitzung, nehmer), naht
}

// eineTatsache verpackt ein Ereignis, wie der Wirt es schickt.
func eineTatsache(t *testing.T, ereignis domain.Ereignis) domain.Partienachricht {
	t.Helper()

	inhalt, fehler := json.Marshal(ereignis)
	if fehler != nil {
		t.Fatalf("verpacken: %v", fehler)
	}

	return domain.Partienachricht{
		Art: domain.ArtEreignis, Partie: "p1", Epoche: 1, Folge: 1,
		Von: "wirt", Inhalt: inhalt, Kette: langeKette,
	}
}

// **Ein Wunsch geht als Kontrolle hinaus — und sagt, dass er einer ist.**
//
// **Die Art im Umschlag ist der Punkt**: „Kontrolle" geht in beide Richtungen
// und bedeutet Verschiedenes. Ohne sie müsste der Wirt am Inhalt raten.
func TestEinWunschGehtAlsKontrolleHinaus(t *testing.T) {
	t.Parallel()

	gast, naht := einGast(&nehmerdoppelGast{})

	fehler := gast.Wuensche(context.Background(), domain.Spielbefehl{
		PartieID: "p1", Absender: "gast", Kennung: "b1",
		Art: "objekt.tappen",
	})
	if fehler != nil {
		t.Fatalf("Wuensche: %v", fehler)
	}

	hinaus := naht.letzte(t)
	if hinaus.Art != domain.ArtKontrolle {
		t.Fatalf("hinaus ging %q statt einer Kontrolle", hinaus.Art)
	}

	var umschlag domain.Kontrolle

	if fehler := json.Unmarshal(hinaus.Inhalt, &umschlag); fehler != nil {
		t.Fatalf("auspacken: %v", fehler)
	}

	if umschlag.Art != domain.KontrolleWunsch {
		t.Errorf("der Umschlag sagt %q", umschlag.Art)
	}
}

// **Ein Tischbild kommt beim Nehmer an.**
//
// **Es ist der zweite Weg, und ohne ihn sähe der Gast nichts**: er hat keinen
// Vollzustand und keinen Regelausleger — wer ihm nur den Strom gäbe, verlangte
// von ihm einen zweiten Spielkern.
func TestEinTischbildKommtBeimNehmerAn(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppelGast{}
	gast, _ := einGast(nehmer)

	inhalt, fehler := domain.VerpackeKontrolle(domain.KontrolleTischbild,
		domain.Tischbild{
			Sicht:   domain.Partiesicht{PartieID: "p1", Fuer: "gast", Fassung: 7},
			Erlaubt: []string{"objekt.tappen"},
		})
	if fehler != nil {
		t.Fatalf("VerpackeKontrolle: %v", fehler)
	}

	if fehler := gast.Angekommen(context.Background(), domain.Partienachricht{
		Art: domain.ArtKontrolle, Partie: "p1", Epoche: 1, Folge: 1,
		Von: "wirt", Inhalt: inhalt,
	}); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	switch {
	case len(nehmer.bilder) != 1:
		t.Fatalf("%d Bilder, erwartet 1", len(nehmer.bilder))
	case nehmer.bilder[0].Sicht.Fassung != 7:
		t.Errorf("Fassung %d", nehmer.bilder[0].Sicht.Fassung)
	case len(nehmer.bilder[0].Erlaubt) != 1:
		t.Errorf("erlaubt: %v — was geht, gehört zum Bild",
			nehmer.bilder[0].Erlaubt)
	}
}

// **Ein Befehl, der sich nicht verpacken lässt, geht nicht hinaus.**
func TestEinUnverpackbarerWunschGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	gast, naht := einGast(&nehmerdoppelGast{})

	fehler := gast.Wuensche(context.Background(), domain.Spielbefehl{
		PartieID: "p1", Absender: "gast", Kennung: "b1",
		Art: "objekt.tappen", Nutzlast: map[string]any{"k": make(chan int)},
	})

	switch {
	case fehler == nil:
		t.Fatal("der Befehl kam durch")
	case len(naht.gesendet) != 0:
		t.Error("es ging trotzdem etwas hinaus")
	}
}

// **Eine Tatsache des Wirts kommt beim Nehmer an — mit der eigenen Partie.**
//
// **Die Partie kommt vom Gast und nicht aus dem Rumpf**: was auf der Leitung
// steht, ist eine Behauptung des anderen Endes.
func TestEineTatsacheKommtMitDerEigenenPartieAn(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppelGast{}
	gast, _ := einGast(nehmer)

	fehler := gast.Angekommen(context.Background(), eineTatsache(t,
		domain.Ereignis{PartieID: "eine-andere", Folge: 4, Art: "objekt.ziehen"}))
	if fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	switch {
	case len(nehmer.bekommen) != 1:
		t.Fatalf("%d Ereignisse, erwartet 1", len(nehmer.bekommen))
	case nehmer.bekommen[0].PartieID != "p1":
		t.Errorf("Partie %q — die Behauptung im Rumpf hat gewonnen",
			nehmer.bekommen[0].PartieID)
	case nehmer.bekommen[0].Folge != 4:
		t.Errorf("Folge %d", nehmer.bekommen[0].Folge)
	}
}

// **Ein Wunsch an einen Gast wird abgewiesen.**
//
// **Die Spiegelung zum Wirt**: er hat keinen Vollzustand und legt nichts aus;
// wer ihm einen Wunsch schickte, hielte ihn für den Wirt.
func TestEinWunschAnEinenGastWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppelGast{}
	gast, _ := einGast(nehmer)

	inhalt, verpackt := domain.VerpackeKontrolle(domain.KontrolleWunsch,
		domain.Spielbefehl{Kennung: "b1", Art: "objekt.tappen"})
	if verpackt != nil {
		t.Fatalf("VerpackeKontrolle: %v", verpackt)
	}

	fehler := gast.Angekommen(context.Background(), domain.Partienachricht{
		Art: domain.ArtKontrolle, Partie: "p1", Epoche: 1, Folge: 1,
		Von: "wirt", Inhalt: inhalt,
	})

	switch {
	case !errors.Is(fehler, domain.ErrValidation):
		t.Fatalf("es kam %v", fehler)
	case len(nehmer.bekommen) != 0 || len(nehmer.bilder) != 0:
		t.Error("der Wunsch wurde genommen")
	}
}

// **Eine Nachricht für eine andere Partie wird abgewiesen.**
func TestEineFremdePartieErreichtDenGastNicht(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppelGast{}
	gast, _ := einGast(nehmer)

	nachricht := eineTatsache(t, domain.Ereignis{Folge: 1, Art: "x"})
	nachricht.Partie = "eine-andere"

	if fehler := gast.Angekommen(
		context.Background(), nachricht,
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}

	if len(nehmer.bekommen) != 0 {
		t.Error("sie kam trotzdem an")
	}
}

// **Ein Inhalt, der kein Ereignis ist, wird abgewiesen** (§8).
func TestEinInhaltDerKeinEreignisIstWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	gast, _ := einGast(&nehmerdoppelGast{})

	if fehler := gast.Angekommen(context.Background(), domain.Partienachricht{
		Art: domain.ArtEreignis, Partie: "p1", Epoche: 1, Folge: 1,
		Von: "wirt", Inhalt: []byte(`"kein Verbund"`), Kette: langeKette,
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Was der Nehmer klagt, kommt zurück.**
func TestWasDerEreignisnehmerKlagtKommtZurueck(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("der Spiegel ist voll")
	gast, _ := einGast(&nehmerdoppelGast{klagt: kaputt})

	if fehler := gast.Angekommen(context.Background(), eineTatsache(t,
		domain.Ereignis{Folge: 1, Art: "x"}),
	); !errors.Is(fehler, kaputt) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Chat und Präsenz gehen auch beim Gast nebenbei.**
func TestChatUndPraesenzGehenBeimGastNebenbei(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppelGast{}
	gast, _ := einGast(nehmer)

	neben := &nebendoppel{}
	gast.MitNebenstrom(neben)

	// **Beide Nutzlasten tragen seit dem Gespräch am Tisch eine Gestalt**
	// ([domain.Chat], [domain.Praesenz]), und der Gast prüft sie: ein
	// `{"text":"hallo"}` als Präsenz ging hier früher durch.
	for art, nutzlast := range map[domain.Nachrichtenart]any{
		domain.ArtChat: domain.Chat{Von: "anna", Text: "hallo"},
		domain.ArtPraesenz: domain.Praesenz{
			Von: "anna", Art: domain.PraesenzZeiger, Stand: 1,
		},
	} {
		if fehler := gast.Angekommen(context.Background(),
			eineWirtsnachricht(t, art, nutzlast)); fehler != nil {
			t.Fatalf("%s: %v", art, fehler)
		}
	}

	switch {
	case len(neben.bekommen) != 2:
		t.Errorf("%d nebenbei, erwartet 2", len(neben.bekommen))
	case len(nehmer.bekommen) != 0:
		t.Error("ein Chat wurde als Ereignis genommen")
	}
}

// **Ohne Zuhörer geschieht nebenbei nichts, und das ist kein Fehler.**
func TestOhneNebenstromHaeltEinChatDenGastNichtAn(t *testing.T) {
	t.Parallel()

	gast, _ := einGast(&nehmerdoppelGast{})

	if fehler := gast.Angekommen(context.Background(), eineWirtsnachricht(t,
		domain.ArtChat, domain.Chat{Von: "anna", Text: "hallo"},
	)); fehler != nil {
		t.Errorf("es kam %v", fehler)
	}
}

// **Eine Art, die es nicht gibt, hält an.**
func TestEineUnbekannteArtHaeltDenGastAn(t *testing.T) {
	t.Parallel()

	gast, _ := einGast(&nehmerdoppelGast{})

	if fehler := gast.Angekommen(context.Background(), domain.Partienachricht{
		Art: domain.Nachrichtenart("wunsch"), Partie: "p1", Epoche: 1,
		Folge: 1, Von: "wirt", Inhalt: []byte(`{}`),
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Ein Tischbild, das keines ist, wird abgewiesen** (§8).
//
// **Zwei Schalen, zwei Läufe**: der Umschlag kann stimmen und die Nutzlast
// darin trotzdem Unsinn sein. Ein Lauf, der nur die äußere prüfte, ließe
// offen, ob die innere überhaupt angesehen wird.
func TestEinTischbildDasKeinesIstWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppelGast{}
	gast, _ := einGast(nehmer)

	inhalt, fehler := domain.VerpackeKontrolle(
		domain.KontrolleTischbild, "kein Verbund")
	if fehler != nil {
		t.Fatalf("VerpackeKontrolle: %v", fehler)
	}

	if fehler := gast.Angekommen(context.Background(), domain.Partienachricht{
		Art: domain.ArtKontrolle, Partie: "p1", Epoche: 1, Folge: 1,
		Von: "wirt", Inhalt: inhalt,
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}

	if len(nehmer.bilder) != 0 {
		t.Error("der Unsinn kam als Tischbild an")
	}
}
