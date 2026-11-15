package service_test

// **Was neben dem Strom läuft** — Werkstück G5 von Phase 5.5
// (#1462).
//
// **Aus `partiewirt_test.go` herausgelöst**, als sie an die Zeilengrenze aus
// §2 stieß. Der Schnitt folgt der Frage: dort steht, was den **Stand** ändert
// — ein Befehl wird ausgelegt, Ereignisse gehen hinaus —, hier, was ihn
// ausdrücklich **nicht** ändert.
//
// # Und das ist der Punkt, nicht ein Nebenschauplatz
//
// **Chat und Präsenz gehen ans Fenster und nicht in den Strom.** Wer sie durch
// dieselbe Tür schickte, müsste bei jedem Chat entscheiden, ob er die Partie
// fortschreibt — und irgendwann entschiede er falsch.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Chat und Präsenz gehen nebenbei und nicht in den Strom.**
func TestChatUndPraesenzGehenNebenbei(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppel{}
	wirt, naht := einWirt(nehmer)

	neben := &nebendoppel{}
	wirt.MitNebenstrom(neben)

	// **Beide Nutzlasten tragen seit dem Gespräch am Tisch eine Gestalt**
	// ([domain.Chat], [domain.Praesenz]). Ein `{"text":"hallo"}` als Präsenz
	// ging hier früher durch und ist heute ein Befund — der Wirt beglaubigt,
	// bevor er weiterreicht.
	for art, nutzlast := range map[domain.Nachrichtenart]any{
		domain.ArtChat: domain.Chat{Von: "gast", Text: "hallo"},
		domain.ArtPraesenz: domain.Praesenz{
			Von: "gast", Art: domain.PraesenzZeiger, Stand: 1,
		},
	} {
		if fehler := wirt.Angekommen(context.Background(),
			eineGastnachricht(t, art, nutzlast)); fehler != nil {
			t.Fatalf("%s: %v", art, fehler)
		}
	}

	switch {
	case len(neben.bekommen) != 2:
		t.Errorf("%d nebenbei, erwartet 2", len(neben.bekommen))
	case len(nehmer.bekommen) != 0:
		t.Error("ein Chat wurde ausgelegt")
	case len(naht.gesendet) != 0:
		t.Error("ein Chat ging als Ereignis hinaus")
	}
}

// **Ohne Zuhörer geschieht nebenbei nichts, und das ist kein Fehler.**
//
// **Die Gegenprobe** (§10): ein Wirt an der Kommandozeile hat kein Fenster,
// und ein Chat, den niemand anzeigt, hält die Partie nicht an.
func TestOhneNebenstromHaeltEinChatNichtsAn(t *testing.T) {
	t.Parallel()

	wirt, _ := einWirt(&nehmerdoppel{})

	if fehler := wirt.Angekommen(context.Background(), eineGastnachricht(t,
		domain.ArtChat, domain.Chat{Von: "gast", Text: "hallo"},
	)); fehler != nil {
		t.Errorf("es kam %v", fehler)
	}
}

// **Und was der Nebenstrom klagt, kommt zurück.**
func TestWasDerNebenstromKlagtKommtZurueck(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("das Fenster ist zu")
	wirt, _ := einWirt(&nehmerdoppel{})
	wirt.MitNebenstrom(&nebendoppel{klagt: kaputt})

	if fehler := wirt.Angekommen(context.Background(), eineGastnachricht(t,
		domain.ArtChat, domain.Chat{Von: "gast", Text: "hallo"},
	)); !errors.Is(fehler, kaputt) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Eine Art, die es nicht gibt, hält an.**
func TestEineUnbekannteArtHaeltDenWirtAn(t *testing.T) {
	t.Parallel()

	wirt, _ := einWirt(&nehmerdoppel{})

	if fehler := wirt.Angekommen(context.Background(), domain.Partienachricht{
		Art: domain.Nachrichtenart("wunsch"), Partie: "p1", Epoche: 1,
		Folge: 1, Von: "gast", Inhalt: []byte(`{}`),
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Ein Ereignis, das sich nicht verpacken lässt, geht nicht hinaus.**
//
// # Warum dieser Zweig einen Lauf bekommt
//
// **Weil er sonst der einzige ungemessene bliebe** — und weil er erreichbar
// ist: `domain.Ereignis.Oeffentlich` ist eine offene Abbildung, und was ein
// Regelausleger hineinlegt, entscheidet er. Ein Wert, den `encoding/json`
// nicht kennt, kommt von dort und nicht von einem Gegenüber; **gemeldet** wird
// er trotzdem, statt still zu verschwinden (§8).
func TestEinUnverpackbaresEreignisGehtNichtHinaus(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppel{gibt: []domain.Ereignis{{
		PartieID: "p1", Folge: 1, Art: "objekt.tappen",
		Oeffentlich: map[string]any{"kanal": make(chan int)},
	}}}

	wirt, naht := einWirt(nehmer)

	fehler := wirt.Angekommen(context.Background(), einSpielwunsch(t,
		domain.Spielbefehl{Kennung: "b1", Art: "objekt.tappen"}))

	switch {
	case fehler == nil:
		t.Fatal("das unverpackbare Ereignis kam durch")
	case len(naht.gesendet) != 0:
		t.Error("es ging trotzdem etwas hinaus")
	}
}

// **Und was die Naht klagt, hält das Verkünden an.**
//
// **Beim ersten und nicht erst am Ende**: wer weitersendete, schickte
// Ereignisse hinaus, deren Vorgänger nie angekommen ist — und der Empfänger
// bekäme eine Lücke, die es beim Wirt gar nicht gibt.
func TestEineKlagendeNahtHaeltDasVerkuendenAn(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("die Warteschlange ist voll")

	nehmer := &nehmerdoppel{gibt: []domain.Ereignis{
		{PartieID: "p1", Folge: 1, Art: "objekt.tappen"},
		{PartieID: "p1", Folge: 2, Art: "objekt.enttappen"},
	}}

	naht := &nahtdoppel{klage: kaputt}
	sitzung := service.NeuePartiesitzung(
		"p1", "wirt", naht, service.NeueReihenfolge(1))
	wirt := service.NeuerPartiewirt("p1", "gast", nehmer, sitzung)

	fehler := wirt.Angekommen(context.Background(), einSpielwunsch(t,
		domain.Spielbefehl{Kennung: "b1", Art: "objekt.tappen"}))

	if !errors.Is(fehler, kaputt) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Der Sitz kommt aus der Sitzung, nicht aus der Nachricht.**
//
// # Der Befund, für den es diesen Lauf gibt
//
// **Der erste Anlauf schrieb `befehl.Absender = nachricht.Von`** — und der
// Kopf darüber sagte „die Sitzung sagt, wer schreibt". Das war falsch:
// `Von` steht in der **Nachricht**, und die schreibt der Gast. Eine Sitzung
// belegt, **welcher Rechner** am anderen Ende ist, und nicht, welcher Sitz;
// wer die beiden gleichsetzte, ließ einen Gast im Namen eines anderen Sitzes
// ziehen — samt dessen Hand.
//
// **Ein Kopf, der mehr verspricht als sein Rumpf hält, ist schlimmer als gar
// keiner** (§10): niemand prüft ihn nach, weil er wie eine Zusage aussieht.
func TestDerSitzKommtAusDerSitzungUndNichtAusDerNachricht(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppel{}
	wirt, _ := einWirt(nehmer)

	// **Der Gast schreibt einen fremden Sitz in die Nachricht.**
	nachricht := einSpielwunsch(t,
		domain.Spielbefehl{Kennung: "b1", Art: "objekt.tappen"})
	nachricht.Von = "anna"

	if fehler := wirt.Angekommen(
		context.Background(), nachricht); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	if len(nehmer.bekommen) != 1 {
		t.Fatalf("%d Befehle angekommen", len(nehmer.bekommen))
	}

	if nehmer.bekommen[0].Absender != "gast" {
		t.Errorf("Absender %q — die Behauptung in der Nachricht hat gewonnen",
			nehmer.bekommen[0].Absender)
	}
}
