package service_test

// **Das Gespräch am Tisch, von der Seite des Gastes** — Werkstück G5 von
// Phase 5.5 (#1462).
//
// # Wogegen hier geprüft wird
//
// **Dass der Gast den Sprecher glaubt und den Stand nicht.** Sein einziges
// Gegenüber ist der Wirt, und der ist die Autorität am Tisch: wer geredet hat,
// hat er gesetzt. **Die Reihenfolge dagegen glaubt er nicht** — eine Präsenz
// wird verwerfbar zugestellt, und eine ältere kann nach einer neueren
// eintreffen. Ohne monotonen Empfängerstand spränge der Zeiger zurück.

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// eineWirtsnachricht baut, was der Wirt einem Gast schickt.
func eineWirtsnachricht(
	t *testing.T, art domain.Nachrichtenart, nutzlast any,
) domain.Partienachricht {
	t.Helper()

	inhalt, fehler := json.Marshal(nutzlast)
	if fehler != nil {
		t.Fatalf("die Nutzlast ließ sich nicht verpacken: %v", fehler)
	}

	nachricht := domain.Partienachricht{
		Art: art, Partie: "p1", Von: "wirt", Epoche: 1, Inhalt: inhalt,
	}

	if art.WirdAufbewahrt() {
		nachricht.Folge = 1
	}

	return nachricht
}

// **Der Gast glaubt den Sprecher, den der Wirt genannt hat.**
//
// **Der Umschlag sagt „wirt", die Nutzlast „anna", und es gilt anna.** Der
// Wirt hat den Satz **geschickt** — geredet hat jemand anderes, und beides
// gleichzusetzen hieße die Herkunft verlieren.
func TestDerGastZeigtDenSprecherUndNichtDenAbsender(t *testing.T) {
	t.Parallel()

	gast, _ := einGast(&nehmerdoppelGast{})

	neben := &nebendoppel{}
	gast.MitNebenstrom(neben)

	if fehler := gast.Angekommen(t.Context(), eineWirtsnachricht(t,
		domain.ArtChat, domain.Chat{Von: "anna", Text: "gut gezogen"},
	)); fehler != nil {
		t.Fatalf("angekommen: %v", fehler)
	}

	if len(neben.bekommen) != 1 {
		t.Fatalf("%d Nachrichten am Fenster, erwartet eine",
			len(neben.bekommen))
	}

	if gesagt := neben.bekommen[0].Chat; gesagt.Von != "anna" {
		t.Errorf("am Fenster steht %q — der Wirt sagte anna", gesagt.Von)
	}
}

// **Ein überholter Zeiger erreicht das Fenster des Gastes nicht.**
func TestEinUeberholterZeigerErreichtDasGastfensterNicht(t *testing.T) {
	t.Parallel()

	gast, _ := einGast(&nehmerdoppelGast{})

	neben := &nebendoppel{}
	gast.MitNebenstrom(neben)

	for _, stand := range []uint64{5, 4, 6} {
		if fehler := gast.Angekommen(t.Context(), eineWirtsnachricht(t,
			domain.ArtPraesenz, domain.Praesenz{
				Von: "anna", Art: domain.PraesenzZeiger, Stand: stand,
			},
		)); fehler != nil {
			t.Fatalf("Stand %d: %v", stand, fehler)
		}
	}

	if len(neben.bekommen) != 2 {
		t.Errorf("%d Zeiger am Fenster, erwartet zwei — der mittlere war "+
			"überholt", len(neben.bekommen))
	}
}

// **Jeder Sitz zählt für sich.**
//
// **Sonst verdrängte der schnellste Sprecher alle anderen**: ein Zähler über
// den ganzen Tisch machte aus annas Stand 9 und berts Stand 3 eine Reihe, in
// der bert nie wieder vorkäme.
func TestJederSitzZaehltBeimGastFuerSich(t *testing.T) {
	t.Parallel()

	gast, _ := einGast(&nehmerdoppelGast{})

	neben := &nebendoppel{}
	gast.MitNebenstrom(neben)

	for _, wer := range []struct {
		sitz  string
		stand uint64
	}{{"anna", 9}, {"bert", 3}} {
		if fehler := gast.Angekommen(t.Context(), eineWirtsnachricht(t,
			domain.ArtPraesenz, domain.Praesenz{
				Von: wer.sitz, Art: domain.PraesenzZeiger, Stand: wer.stand,
			},
		)); fehler != nil {
			t.Fatalf("%s: %v", wer.sitz, fehler)
		}
	}

	if len(neben.bekommen) != 2 {
		t.Errorf("%d Zeiger am Fenster, erwartet zwei", len(neben.bekommen))
	}
}

// **Was keine Gestalt hat, hält auch beim Gast an.**
func TestEineFormloseNebennachrichtHaeltDenGastAn(t *testing.T) {
	t.Parallel()

	gast, _ := einGast(&nehmerdoppelGast{})
	gast.MitNebenstrom(&nebendoppel{})

	if fehler := gast.Angekommen(t.Context(), eineWirtsnachricht(t,
		domain.ArtPraesenz, domain.Chat{Von: "anna", Text: "das ist kein Zeiger"},
	)); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Ein Gast sagt etwas, und es geht als Chat hinaus.**
func TestEinGastSagtEtwas(t *testing.T) {
	t.Parallel()

	gast, naht := einGast(&nehmerdoppelGast{})

	if fehler := gast.Sage(t.Context(), domain.Chat{
		Von: "gast", Text: "gleich bin ich wieder da",
	}); fehler != nil {
		t.Fatalf("sagen: %v", fehler)
	}

	nachricht := naht.letzte(t)

	if nachricht.Art != domain.ArtChat {
		t.Errorf("Art %q statt Chat", nachricht.Art)
	}

	var gesagt domain.Chat
	if fehler := json.Unmarshal(nachricht.Inhalt, &gesagt); fehler != nil {
		t.Fatalf("kein Satz: %v", fehler)
	}

	if gesagt.Text != "gleich bin ich wieder da" {
		t.Errorf("gesagt wurde %q", gesagt.Text)
	}
}

// **Ein Gast zeigt, und es geht als Präsenz hinaus — verwerfbar.**
func TestEinGastZeigtAufEtwas(t *testing.T) {
	t.Parallel()

	gast, naht := einGast(&nehmerdoppelGast{})

	if fehler := gast.Zeige(t.Context(), domain.Praesenz{
		Von: "gast", Art: domain.PraesenzSchwebt, Stand: 4,
		Objekt: "k1", X: 12, Y: 34,
	}); fehler != nil {
		t.Fatalf("zeigen: %v", fehler)
	}

	nachricht := naht.letzte(t)

	if nachricht.Art != domain.ArtPraesenz {
		t.Errorf("Art %q statt Präsenz", nachricht.Art)
	}

	if nachricht.Folge != 0 {
		t.Errorf("die Präsenz trägt die Folge %d — sie hat keine Stelle in "+
			"einer Reihenfolge", nachricht.Folge)
	}

	naht.schloss.Lock()
	defer naht.schloss.Unlock()

	if letzte := naht.arten[len(naht.arten)-1]; letzte != service.Verwerfbar {
		t.Errorf("zugestellt als %v statt verwerfbar", letzte)
	}
}

// **Was keine Gestalt hat, geht gar nicht erst hinaus.**
func TestEinGastSchicktNichtsFormloses(t *testing.T) {
	t.Parallel()

	gast, naht := einGast(&nehmerdoppelGast{})

	if fehler := gast.Sage(t.Context(), domain.Chat{Von: "gast"}); fehler == nil {
		t.Error("ein leerer Satz ging hinaus")
	}

	if fehler := gast.Zeige(t.Context(), domain.Praesenz{
		Von: "gast", Art: domain.PraesenzZeiger,
	}); fehler == nil {
		t.Error("eine Präsenz ohne Stand ging hinaus")
	}

	naht.schloss.Lock()
	defer naht.schloss.Unlock()

	if len(naht.gesendet) != 0 {
		t.Errorf("%d Nachrichten gingen hinaus, die nie hätten hinausgehen "+
			"dürfen", len(naht.gesendet))
	}
}
