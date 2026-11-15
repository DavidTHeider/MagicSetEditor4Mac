package service_test

// **Der Wirt am anderen Ende der Leitung** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// **Dass ein Gast wünscht und der Wirt Tatsachen schafft.** Die Aufteilung ist
// keine Bequemlichkeit: der Wirt hält den Vollzustand, er allein legt aus, und
// er allein vergibt Folge und Kette. Ein Gast, der Ereignisse schickte,
// schriebe am Strom mit — und niemand könnte hinterher sagen, welcher Stand
// gilt.
//
// **Und dass der Absender aus der Sitzung kommt und nicht aus dem Rumpf.** Das
// ist der Lauf, an dem am meisten hängt: wer den Rumpf glaubte, ließe einen
// Gast im Namen eines anderen ziehen.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// nehmerdoppel merkt sich den Befehl und gibt feste Ereignisse.
type nehmerdoppel struct {
	bekommen []domain.Spielbefehl
	gibt     []domain.Ereignis
	klagt    error

	// gefragt hält fest, für wen ein Tischbild gerechnet wurde.
	gefragt []string

	// ohneErlaubnis lässt Erlaubnis und Handlungen `nil` werden.
	ohneErlaubnis bool

	// reihe zählt, wie oft eine Antwort in der Reihe lief (#2481).
	reihe int

	// bildklage lässt **nur** das Tischbild scheitern — der Befehl geht
	// durch. Ein Doppel, das auf jedem Weg klagt, klagt schon beim ersten,
	// und alles danach bliebe ungemessen (§10).
	bildklage error

	// bildklageNach lässt die ersten Rufe gelingen — seit P9b fragt der Wirt
	// bei der Rückkehr zuerst, ob der Sitz wartet, und erst danach nach dem
	// Bild.
	bildklageNach int

	// stand ist, was dieser Nehmer über die Rückkehr sagt.
	stand    domain.Wirtsstand
	nachholt []domain.Ereignis

	// nachgereicht hält fest, ab welcher Folge gefragt wurde.
	nachgereicht []int64
	// nachreichklage klagt nur auf dem Weg des Nachreichens — für den Lauf,
	// der genau ihn scheitern lässt (§10).
	nachreichklage error
}

// Partie gibt das Tischbild, das der Wirt an den Gast schickt.
func (n *nehmerdoppel) Partie(
	_ context.Context, kennung, fuer string,
) (domain.Partiesicht, error) {
	n.gefragt = append(n.gefragt, fuer)

	if n.bildklage != nil && len(n.gefragt) > n.bildklageNach {
		return domain.Partiesicht{}, n.bildklage
	}

	return domain.Partiesicht{PartieID: kennung, Fuer: fuer, Fassung: 2}, nil
}

// Tischbild setzt das Bild aus der Sicht dieses Doppels zusammen — und
// klagt, wo sie klagt.
func (n *nehmerdoppel) Tischbild(
	ctx context.Context, kennung, fuer string,
) (domain.Tischbild, error) {
	sicht, fehler := n.Partie(ctx, kennung, fuer)
	if fehler != nil || n.ohneErlaubnis {
		return domain.Tischbild{Sicht: sicht}, fehler
	}

	return domain.Tischbild{
		Sicht: sicht, Erlaubt: []string{"objekt.tappen"},
		Handlungen: []domain.Objekthandlungen{{
			Objekt: "o1", Handlungen: []string{"objekt.tappen"},
			Primaer: "objekt.tappen",
		}},
	}, nil
}

func (n *nehmerdoppel) BefehlVerkuendet(
	_ context.Context, befehl domain.Spielbefehl, verkuende service.Verkuendung,
) (service.Befehlsausgang, error) {
	n.bekommen = append(n.bekommen, befehl)

	if n.klagt != nil {
		return service.Befehlsausgang{}, n.klagt
	}

	ausgang := service.Befehlsausgang{Fassung: 2, Ereignisse: n.gibt}
	verkuende(ausgang)

	return ausgang, nil
}

func (n *nehmerdoppel) InReihe(tun func() error) error {
	n.reihe++

	return tun()
}

// Wirtsstand und Nachreichen: was der Wirt für eine Rückkehr braucht.
//
// **Der Stand kommt aus dem Doppel und nicht aus einem Bestand** — diese
// Läufe prüfen, was der Wirt **tut**, nicht was der Bestand weiß.
func (n *nehmerdoppel) Wirtsstand(
	context.Context, string, string,
) (domain.Wirtsstand, error) {
	if n.klagt != nil {
		return domain.Wirtsstand{}, n.klagt
	}

	return n.stand, nil
}

func (n *nehmerdoppel) Nachreichen(
	_ context.Context, _, _ string, abFolge int64,
) ([]domain.Ereignis, error) {
	if n.klagt != nil {
		return nil, n.klagt
	}

	if n.nachreichklage != nil {
		return nil, n.nachreichklage
	}

	n.nachgereicht = append(n.nachgereicht, abFolge)

	return n.nachholt, nil
}

// nebendoppel sammelt, was neben dem Strom läuft.
type nebendoppel struct {
	bekommen []service.Nebenbefund
	klagt    error
}

func (n *nebendoppel) Nebenbei(
	_ context.Context, befund service.Nebenbefund,
) error {
	n.bekommen = append(n.bekommen, befund)

	return n.klagt
}

// einWirt baut ihn über einem Nahtdoppel.
func einWirt(nehmer service.Befehlsnehmer) (*service.Partiewirt, *nahtdoppel) {
	naht := &nahtdoppel{}
	sitzung := service.NeuePartiesitzung(
		"p1", "wirt", naht, service.NeueReihenfolge(1))

	return service.NeuerPartiewirt("p1", "gast", nehmer, sitzung), naht
}

// einSpielwunsch verpackt einen Spielbefehl als Kontrollnachricht.
//
// **Der Absender steht fest und ist kein Argument.** Er ist in jedem Lauf
// derselbe — der Gast —, und der Punkt ist gerade, dass er aus der **Sitzung**
// kommt und nicht aus dem Rumpf. Ein Argument dafür lud dazu ein, ihn zu
// variieren und damit die falsche Frage zu stellen.
func einSpielwunsch(
	t *testing.T, befehl domain.Spielbefehl,
) domain.Partienachricht {
	t.Helper()

	return eineKontrollnachricht(t, domain.KontrolleWunsch, befehl)
}

// eineKontrollnachricht verpackt eine Nutzlast als Kontrolle vom Gast.
func eineKontrollnachricht(
	t *testing.T, art domain.Kontrollart, nutzlast any,
) domain.Partienachricht {
	t.Helper()

	inhalt, fehler := domain.VerpackeKontrolle(art, nutzlast)
	if fehler != nil {
		t.Fatalf("VerpackeKontrolle: %v", fehler)
	}

	return domain.Partienachricht{
		Art: domain.ArtKontrolle, Partie: "p1", Epoche: 1, Folge: 1,
		Von: "gast", Inhalt: inhalt,
	}
}

// ausgepackt liest eine hinausgegangene Nachricht zurück.
func ausgepackt(t *testing.T, roh []byte) domain.Partienachricht {
	t.Helper()

	var nachricht domain.Partienachricht

	if fehler := json.Unmarshal(roh, &nachricht); fehler != nil {
		t.Fatalf("auspacken: %v (%s)", fehler, roh)
	}

	return nachricht
}

// **Ein fremder Wunsch wird ausgelegt, und die Tatsachen gehen zurück.**
func TestEinFremderWunschWirdAusgelegtUndVerkuendet(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppel{gibt: []domain.Ereignis{
		{PartieID: "p1", Folge: 1, Art: "objekt.tappen"},
		{PartieID: "p1", Folge: 2, Art: "objekt.enttappen"},
	}}

	wirt, naht := einWirt(nehmer)

	fehler := wirt.Angekommen(context.Background(), einSpielwunsch(t,
		domain.Spielbefehl{Kennung: "b1", Art: "objekt.tappen"}))
	if fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	// **Je Ereignis eine Nachricht, und danach das Bild.** Ein Bündel hätte
	// eine Nummer für mehrere Schritte, und eine Lücke mittendrin wäre nicht
	// zu benennen.
	if len(naht.gesendet) != 3 {
		t.Fatalf("%d Nachrichten hinaus, erwartet 3", len(naht.gesendet))
	}

	erste := ausgepackt(t, naht.gesendet[0])
	if erste.Art != domain.ArtEreignis || erste.Von != "wirt" {
		t.Errorf("hinaus ging %+v — erwartet ein Ereignis vom Wirt", erste)
	}

	// **Das Bild kommt zuletzt und zeigt den Stand danach.** Wer es davor
	// schickte, gäbe dem Gast einen Tisch, der gleich veraltet ist.
	letzte := ausgepackt(t, naht.gesendet[2])
	if letzte.Art != domain.ArtKontrolle {
		t.Fatalf("zuletzt ging %q hinaus", letzte.Art)
	}

	// **Und es sagt, was geht.** Ohne diese Liste wüsste ein Gast nie, dass
	// er gerade nichts darf — die Sicht sagt es nicht.
	if !strings.Contains(string(letzte.Inhalt), "objekt.tappen") {
		t.Errorf("das Bild nennt nichts Erlaubtes: %s", letzte.Inhalt)
	}

	// **Und es ist für den Gast gerechnet** — nicht für den Wirt.
	if len(nehmer.gefragt) != 1 || nehmer.gefragt[0] != "gast" {
		t.Errorf("gerechnet wurde für %v", nehmer.gefragt)
	}
}

// **Auch der Rumpf des Befehls entscheidet weder Sitz noch Partie.**
//
// **Zwei Behauptungen in einer Nachricht, und beide verlieren**: der Lauf
// darunter nimmt der **Nachricht** ihren Absender weg, dieser dem **Befehl
// im Rumpf**. Es sind zwei Stellen, an denen ein Gast dasselbe versuchen
// kann, und eine Zusage, die nur eine davon hält, hält keine.
func TestDerRumpfEntscheidetWederSitzNochPartie(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppel{}
	wirt, _ := einWirt(nehmer)

	// **Der Gast behauptet, jemand anderes zu sein** — und er behauptet
	// dazu eine fremde Partie.
	gelogen := domain.Spielbefehl{
		PartieID: "eine-andere", Absender: "wirt",
		Kennung: "b1", Art: "objekt.tappen",
	}

	if fehler := wirt.Angekommen(
		context.Background(), einSpielwunsch(t, gelogen),
	); fehler != nil {
		t.Fatalf("Angekommen: %v", fehler)
	}

	if len(nehmer.bekommen) != 1 {
		t.Fatalf("%d Befehle angekommen", len(nehmer.bekommen))
	}

	switch {
	case nehmer.bekommen[0].Absender != "gast":
		t.Errorf("Absender %q — die Behauptung im Rumpf hat gewonnen",
			nehmer.bekommen[0].Absender)
	case nehmer.bekommen[0].PartieID != "p1":
		t.Errorf("Partie %q — die Behauptung im Rumpf hat gewonnen",
			nehmer.bekommen[0].PartieID)
	}
}

// **Ein Ereignis von außen wird abgewiesen.**
//
// **Es ist kein Formfehler, sondern eine Anmaßung**: wer es durchließe, hätte
// zwei Quellen für dieselbe Wahrheit.
func TestEinEreignisVonAussenWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppel{}
	wirt, naht := einWirt(nehmer)

	fehler := wirt.Angekommen(context.Background(), domain.Partienachricht{
		Art: domain.ArtEreignis, Partie: "p1", Epoche: 1, Folge: 1,
		Von: "gast", Inhalt: []byte(`{}`), Kette: langeKette,
	})

	switch {
	case !errors.Is(fehler, domain.ErrValidation):
		t.Fatalf("es kam %v", fehler)
	case len(nehmer.bekommen) != 0:
		t.Error("das Ereignis wurde ausgelegt")
	case len(naht.gesendet) != 0:
		t.Error("es ging etwas hinaus")
	}
}

// **Eine Nachricht für eine andere Partie wird abgewiesen.**
func TestEineFremdePartieWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppel{}
	wirt, _ := einWirt(nehmer)

	nachricht := einSpielwunsch(t,
		domain.Spielbefehl{Kennung: "b1", Art: "objekt.tappen"})
	nachricht.Partie = "eine-andere"

	if fehler := wirt.Angekommen(
		context.Background(), nachricht,
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("es kam %v", fehler)
	}

	if len(nehmer.bekommen) != 0 {
		t.Error("der Befehl wurde trotzdem ausgelegt")
	}
}

// **Ein Inhalt, der keine Kontrolle ist, wird abgewiesen** (§8).
func TestEinInhaltDerKeineKontrolleIstWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppel{}
	wirt, _ := einWirt(nehmer)

	if fehler := wirt.Angekommen(context.Background(), domain.Partienachricht{
		Art: domain.ArtKontrolle, Partie: "p1", Epoche: 1, Folge: 1,
		Von: "gast", Inhalt: []byte(`"kein Verbund"`),
	}); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Und ein Inhalt, der keinen Befehl trägt, ebenfalls.**
//
// **Zwei Schalen, zwei Läufe**: der Umschlag kann stimmen und die Nutzlag
// darin trotzdem Unsinn sein. Ein Lauf, der nur die äußere prüfte, ließe
// offen, ob die innere überhaupt angesehen wird.
func TestEineKontrolleOhneBefehlWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	nehmer := &nehmerdoppel{}
	wirt, _ := einWirt(nehmer)

	if fehler := wirt.Angekommen(context.Background(),
		eineKontrollnachricht(t, domain.KontrolleWunsch, "kein Verbund"),
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}

	if len(nehmer.bekommen) != 0 {
		t.Error("der Unsinn wurde ausgelegt")
	}
}
