package service_test

// **Der Schlüssel des Packspeichers** — Befunde 5.1 und 5.3 des Nachlaufs über
// Phase 4.1 (#1288).
//
// **Eine eigene Datei**, weil `packspeicher_test.go` mit ihnen über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort steht, **dass**
// nicht zweimal gerechnet wird, hier, **woran** der Schlüssel hängt.

import (
	"context"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Ziehen nach einem Angebot rechnet nicht noch einmal.**
//
// Die Oberfläche ruft erst `Packarten` und dann beim Klick `PackZiehen`
// (`frontend/src/seiten/Zufallspack.tsx`). Ein Speicher, der nur `Angebot`
// bediente, ließe den zweiten Weg von vorn anfangen — und das ist der Weg, den
// ein Mensch geht (Nachlauf über Phase 4.1, Befund 5.3).
func TestZiehenNachEinemAngebotRechnetNichtNochEinmal(t *testing.T) {
	t.Parallel()

	stile := &zaehlendeStile{vorlage: packvorlage()}

	bereich := service.NeuerPackbereich(nil,
		setkarten{karten: []*domain.CardRow{packkarte(1, "common")}},
		service.NeuerPlanbauer().MitStilen(stile))

	if _, _, err := bereich.Angebot(t.Context(), 1); err != nil {
		t.Fatalf("Angebot: %v", err)
	}

	nachAngebot := stile.stand()
	if nachAngebot == 0 {
		t.Fatal("das Angebot hat keinen Stil nachgeschlagen — der Zähler misst " +
			"die falsche Stelle")
	}

	if _, _, err := bereich.Ziehen(t.Context(), 1, "common", 1); err != nil {
		t.Fatalf("Ziehen: %v", err)
	}

	if stile.stand() != nachAngebot {
		t.Errorf("das Ziehen hat %d Stile erneut nachgeschlagen — die Filterung "+
			"stand schon da", stile.stand()-nachAngebot)
	}
}

// **Ein geänderter eigener Stilbogen macht das Gemerkte ungültig.**
//
// Eine Karte darf einen eigenen Stil tragen, und dessen Initskript entscheidet
// über den Filter mit. Ohne ihn im Schlüssel bliebe die Filterung stehen,
// obwohl der Stil ein anderer ist (Nachlauf über Phase 4.1, Befund 5.1).
func TestEinGeaenderterEigenerStilRechnetNeu(t *testing.T) {
	t.Parallel()

	karte := karteMitEigenemStil(1, "stil-eins", "eins")
	karten := []*domain.CardRow{karte}

	// **Die Setvorlage bleibt dieselbe**, nur die der Karte wechselt. Sonst
	// änderte sich der Schlüssel ohnehin über die Setvorlage, und der Lauf
	// prüfte nicht, was in seinem Namen steht.
	//
	// **Die Packart kommt aus der Setvorlage, der Name aus dem Initskript der
	// Karte** — so hängt das Ergebnis wirklich am eigenen Stilbogen.
	stile := &wechselndeStile{
		setweit: service.Vorlage{
			Packs: []msepaket.Pack{{Name: "je Stil", Filter: `erkannt == card.name`}},
		},
		vorlage: packvorlageMitInitskript("stil-eins", "eins"),
	}

	bereich := service.NeuerPackbereich(nil, setkarten{karten: karten},
		service.NeuerPlanbauer().MitStilen(stile))

	erstes, _, err := bereich.Angebot(t.Context(), 1)
	if err != nil {
		t.Fatalf("Angebot: %v", err)
	}

	if len(erstes) != 1 || erstes[0].Karten != 1 {
		t.Fatalf("das erste Angebot ist %+v — erwartet war die eine Karte, "+
			"deren eigener Stilbogen ihren Namen setzt", erstes)
	}

	// **Derselbe Name, ein anderes Initskript** — so, wie es nach einer
	// Änderung in der Paketwerkstatt aussieht.
	stile.setze(packvorlageMitInitskript("stil-eins", "etwas anderes"))

	zweites, _, err := bereich.Angebot(t.Context(), 1)
	if err != nil {
		t.Fatalf("zweites Angebot: %v", err)
	}

	if len(zweites) != 1 || zweites[0].Karten != 0 {
		t.Errorf("das zweite Angebot ist %+v — erwartet war keine Karte: der "+
			"eigene Stilbogen setzt jetzt einen anderen Namen", zweites)
	}
}

// wechselndeStile gibt der Karte eine Vorlage, die sich zwischen zwei Läufen
// ändert — bei **unveränderter** Setvorlage.
type wechselndeStile struct {
	mu      sync.Mutex
	setweit service.Vorlage
	vorlage service.Vorlage
}

func (w *wechselndeStile) setze(vorlage service.Vorlage) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.vorlage = vorlage
}

func (w *wechselndeStile) StilFuerSet(context.Context, int64) (
	service.Vorlage, []domain.Befund, bool,
) {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.setweit, nil, true
}

func (w *wechselndeStile) StilFuerKarte(context.Context, *domain.CardRow) (
	service.Vorlage, []domain.Befund, bool,
) {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.vorlage, nil, true
}

// **Zwei Karten mit demselben eigenen Stilbogen kosten einen Nachschlag.**
//
// Der Schlüssel braucht die Vorlage je **Name** und nicht je Karte: ein Set
// mit hundert Karten desselben Sonderstils schlüge ihn sonst hundertmal nach,
// und `Stilspeicher` sagt es eine Ebene tiefer schon — „ein Set hat wenige
// Stile und viele Karten".
func TestZweiKartenMitDemselbenEigenenStilKostenEinenNachschlag(t *testing.T) {
	t.Parallel()

	vorlage := packvorlageMitInitskript("stil-eins", "eins")

	karten := []*domain.CardRow{
		karteMitEigenemStil(1, "stil-eins", "eins"),
		karteMitEigenemStil(2, "stil-eins", "eins"),

		// **Und eine ohne eigenen Stil** — sie steht gar nicht erst in der
		// Liste der nachzuschlagenden Namen.
		{ID: 3, SetID: 1, Values: []domain.Value{{Field: "name", Text: "eins"}}},
	}

	stile := &zaehlendeStile{vorlage: vorlage}

	bereich := service.NeuerPackbereich(nil, setkarten{karten: karten},
		service.NeuerPlanbauer().MitStilen(stile))

	angebot, _, err := bereich.Angebot(t.Context(), 1)
	if err != nil {
		t.Fatalf("Angebot: %v", err)
	}

	if len(angebot) != 1 || angebot[0].Karten != len(karten) {
		t.Errorf("das Angebot ist %+v — erwartet waren alle %d Karten",
			angebot, len(karten))
	}
}
