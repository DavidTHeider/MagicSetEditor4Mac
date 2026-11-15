// **Eine seteigene Packart geht der des Spiels vor**
// (#1081).
//
// MSE sucht zuerst im Set, dann im Spiel:
//
//	FOR_EACH_CONST(type, set->pack_types)       { … }
//	FOR_EACH_CONST(type, set->game->pack_types) { … }
//
// `src/data/pack.cpp`, Zeilen 373–392, Fassung f9d9356 vom 2021-07-06. Das
// Feld heißt dort „Additional/replacement pack types" — **beides**: eine neue
// Art kommt hinzu, eine gleichnamige ersetzt.
package service_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Gleicher Name: die des Sets gewinnt.**
func TestEineGleichnamigeSetpackartErsetzt(t *testing.T) {
	t.Parallel()

	spiel := []msepaket.Pack{
		{Name: "booster", Waehlbar: true, Posten: []msepaket.Packposten{
			{Name: "common", Anzahl: 10, Gewicht: 1},
		}},
		{Name: "starter", Waehlbar: true},
	}

	eigene := []*domain.PackArt{
		{Name: "booster", Selectable: true, Items: []domain.PackItem{
			{Name: "common", Amount: 3, Weight: 1},
		}},
	}

	arten := service.PackartenMitVorrang(eigene, spiel)

	if len(arten) != 2 {
		t.Fatalf("%d Arten statt zwei: %+v", len(arten), arten)
	}

	if arten[0].Name != "booster" || len(arten[0].Posten) != 1 ||
		arten[0].Posten[0].Anzahl != 3 {
		t.Errorf("die Art des Sets hat nicht gewonnen: %+v", arten[0])
	}

	if arten[1].Name != "starter" {
		t.Errorf("die zweite Art des Spiels fehlt: %+v", arten)
	}
}

// **Neuer Name: sie kommt hinzu, und zwar vorn.**
//
// **Vorn und nicht hinten**: MSE durchläuft `set->pack_types` zuerst, und die
// Reihenfolge ist die, in der die Oberfläche sie anbietet.
func TestEineNeueSetpackartKommtVorn(t *testing.T) {
	t.Parallel()

	spiel := []msepaket.Pack{{Name: "booster", Waehlbar: true}}
	eigene := []*domain.PackArt{{Name: "jumpstart", Selectable: true}}

	arten := service.PackartenMitVorrang(eigene, spiel)

	if len(arten) != 2 {
		t.Fatalf("%d Arten statt zwei: %+v", len(arten), arten)
	}

	if arten[0].Name != "jumpstart" {
		t.Errorf("die eigene Art steht nicht vorn: %+v", arten)
	}
}

// **Ohne eigene bleibt es bei denen des Spiels — unverändert.**
//
// Die Gegenprobe: ohne sie sagte der Test darüber nur, dass eine Liste
// entsteht, nicht dass der Vorrang greift.
func TestOhneEigeneBleibenDieDesSpiels(t *testing.T) {
	t.Parallel()

	spiel := []msepaket.Pack{
		{Name: "booster", Waehlbar: true},
		{Name: "starter", Waehlbar: false},
	}

	arten := service.PackartenMitVorrang(nil, spiel)

	if len(arten) != 2 || arten[0].Name != "booster" || arten[1].Name != "starter" {
		t.Errorf("die Liste des Spiels kam verändert zurück: %+v", arten)
	}
}

// **Zwei eigene mit demselben Namen: die erste gewinnt.**
//
// MSE bricht die Schleife beim ersten Treffer ab (`return &*type;`,
// `src/data/pack.cpp`, Zeile 377, Fassung f9d9356 vom 2021-07-06).
func TestBeiZweiEigenenGewinntDieErste(t *testing.T) {
	t.Parallel()

	eigene := []*domain.PackArt{
		{Name: "booster", Filter: "erste"},
		{Name: "booster", Filter: "zweite"},
	}

	arten := service.PackartenMitVorrang(eigene, nil)

	if len(arten) != 1 {
		t.Fatalf("%d Arten statt einer: %+v", len(arten), arten)
	}

	if arten[0].Filter != "erste" {
		t.Errorf("die zweite hat gewonnen: %+v", arten[0])
	}
}

// **Der Packbereich zieht aus der Art des Sets, nicht aus der des Spiels.**
//
// Das ist der Weg **durch** die Naht — die Tests darüber prüfen die Regel für
// sich. Ohne diesen sagte keiner, dass `Packbereich` sie überhaupt ruft.
func TestDerPackbereichZiehtAusDerSetart(t *testing.T) {
	t.Parallel()

	karten := []*domain.CardRow{
		packkarte(1, "common"), packkarte(2, "common"),
		packkarte(3, "common"), packkarte(4, "common"),
	}

	// **Das Spiel zieht zwei, das Set drei.** Die Zahl unterscheidet sie —
	// gleiche Zahlen ließen den Vorrang unsichtbar.
	eigene := []*domain.PackArt{{
		Name: "booster", Select: "all", Selectable: true,
		Items: []domain.PackItem{{Name: "common", Amount: 3, Weight: 1}},
	}}

	bereich := service.NeuerPackbereich(nil,
		setkarten{karten: karten, eigene: eigene},
		service.NeuerPlanbauer().MitStilen(spielwoerter{packvorlage(), true}))

	gezogen, _, err := bereich.Ziehen(t.Context(), 1, "booster", 42)
	if err != nil {
		t.Fatalf("Ziehen: %v", err)
	}

	if len(gezogen) != 3 {
		t.Errorf("%d Karten gezogen — das Spiel zieht zwei, das Set drei", len(gezogen))
	}

	// **Die Gegenprobe**: ohne eigene Art zieht dasselbe Set zwei.
	ohne := service.NeuerPackbereich(nil, setkarten{karten: karten},
		service.NeuerPlanbauer().MitStilen(spielwoerter{packvorlage(), true}))

	zurueck, _, ohnefehler := ohne.Ziehen(t.Context(), 1, "booster", 42)
	if ohnefehler != nil {
		t.Fatalf("Ziehen ohne eigene: %v", ohnefehler)
	}

	if len(zurueck) != 2 {
		t.Errorf("%d Karten ohne eigene Art — erwartet zwei", len(zurueck))
	}
}

// **Und eine eigene Art steht im Angebot, auch ohne Stil.**
//
// Ein Set kann eine Packart nennen, ohne dass sein Stil auffindbar ist. Der
// erste Anlauf holte die eigenen erst hinter der Vorlagenprüfung — dann zog
// ein Set ohne Stil aus nichts.
func TestEineEigeneArtStehtAuchOhneStilImAngebot(t *testing.T) {
	t.Parallel()

	eigene := []*domain.PackArt{{Name: "jumpstart", Select: "all", Selectable: true}}

	bereich := service.NeuerPackbereich(nil,
		setkarten{eigene: eigene},
		service.NeuerPlanbauer().MitStilen(spielwoerter{service.Vorlage{}, false}))

	angebot, _, err := bereich.Angebot(t.Context(), 1)
	if err != nil {
		t.Fatalf("Angebot: %v", err)
	}

	if len(angebot) != 1 || angebot[0].Name != "jumpstart" {
		t.Errorf("das Angebot ist %+v", angebot)
	}
}

// **Ein Lesefehler bei den eigenen Arten kommt durch.**
func TestEinFehlerBeimLesenDerSetartenKommtDurch(t *testing.T) {
	t.Parallel()

	bereich := service.NeuerPackbereich(nil,
		setkarten{packfehler: errFehlendeSetarten}, service.NeuerPlanbauer())

	if _, _, err := bereich.Angebot(t.Context(), 1); err == nil {
		t.Error("das Angebot kam durch")
	}

	if _, _, err := bereich.Ziehen(t.Context(), 1, "booster", 1); err == nil {
		t.Error("das Ziehen kam durch")
	}
}

// errFehlendeSetarten ist der Fehler, den die Probe meldet.
var errFehlendeSetarten = errors.New("die Packarten des Sets sind nicht lesbar")
