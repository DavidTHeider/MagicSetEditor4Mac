package service

// **Beim Bearbeiten zieht MSE nach** — #2396
// (Issue (#2396)).
//
// `SetScriptManager::updateValue` rechnet das geänderte Feld und reiht die
// Felder ein, deren Skript es liest (`alsoUpdate` über `dependent_scripts`);
// `updateRecursive` rechnet sie der Reihe nach, jedes höchstens einmal je
// Runde, und reiht die Leser eines Felds nur ein, wenn es sich **geändert**
// hat (`src/script/script_manager.cpp`, Zeilen 249–262 und 309–341, Fassung
// f9d9356 vom 2021-07-06). Gespeichert wird danach, was gerechnet ist.

import (
	"context"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// nachzug fährt nachziehen über Einträge, wie die Karte sie führt.
func nachzug(
	t *testing.T, felder []domain.Field, werte map[string]string, geaendert ...string,
) (ergebnis map[string]string, gezogen []string) {
	t.Helper()

	gefuellt := make([]msescript.Eintrag, 0, len(felder))
	for _, feld := range felder {
		gefuellt = append(gefuellt, msescript.Eintrag{Name: feld.Name, Wert: msescript.Text(werte[feld.Name])})
	}

	umgebung := msescript.Bibliothek().Darueber()
	karte := func(stand []msescript.Eintrag) {
		umgebung.Setze("card", msescript.Wert{
			Art: msescript.ArtAbbildung, Eintraege: append([]msescript.Eintrag(nil), stand...),
		})
	}
	karte(gefuellt)

	stand, gerechnet, befunde := nachziehen(context.Background(), gefuellt, felder, umgebung, nil, karte, geaendert)
	if len(befunde) != 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	ergebnis = make(map[string]string, len(stand))
	for _, eintrag := range stand {
		ergebnis[eintrag.Name] = eintrag.Wert.AlsText()
	}

	return ergebnis, gerechnet
}

func TestEinGesetztesFeldWirdDurchSeinSkriptGerechnet(t *testing.T) {
	t.Parallel()

	werte, gerechnet := nachzug(t, []domain.Field{
		{Name: "eins", Skript: `"[" + value + "]"`},
	}, map[string]string{"eins": "Funke"}, "eins")

	if werte["eins"] != "[Funke]" || !slices.Equal(gerechnet, []string{"eins"}) {
		t.Errorf("eins ist %q, gerechnet %v", werte["eins"], gerechnet)
	}
}

func TestWerEinGeaendertesFeldLiestWirdNachgezogen(t *testing.T) {
	t.Parallel()

	// **`zwei` steht vorn** — nachgezogen wird nach Abhängigkeit, nicht nach
	// der Reihenfolge des Spiels.
	werte, gerechnet := nachzug(t, []domain.Field{
		{Name: "zwei", Skript: `card.eins + "!"`},
		{Name: "eins", Skript: `"[" + value + "]"`},
		{Name: "drei", Skript: `card.zwei + "?"`},
	}, map[string]string{"eins": "Funke", "zwei": "alt", "drei": "alt"}, "eins")

	if werte["zwei"] != "[Funke]!" || werte["drei"] != "[Funke]!?" {
		t.Errorf("zwei ist %q, drei %q", werte["zwei"], werte["drei"])
	}

	if !slices.Equal(gerechnet, []string{"eins", "zwei", "drei"}) {
		t.Errorf("gerechnet %v", gerechnet)
	}
}

// **Was niemand liest, bleibt, wie es gespeichert ist** — auch wenn sein
// eigenes Skript etwas anderes ergäbe. Das ist der Unterschied zum Laden.
func TestEinFeldDasNichtsLiestBleibt(t *testing.T) {
	t.Parallel()

	werte, gerechnet := nachzug(t, []domain.Field{
		{Name: "eins", Skript: `"[" + value + "]"`},
		{Name: "fremd", Skript: `"<" + value + ">"`},
	}, map[string]string{"eins": "Funke", "fremd": "roh"}, "eins")

	if werte["fremd"] != "roh" || slices.Contains(gerechnet, "fremd") {
		t.Errorf("fremd ist %q, gerechnet %v", werte["fremd"], gerechnet)
	}
}

// **Ein Leser, der sich nicht ändert, zieht nichts nach** — `updateToUpdate`
// reiht nur ein, was sich geändert hat.
func TestEinUnveraenderterLeserZiehtNichtsNach(t *testing.T) {
	t.Parallel()

	werte, gerechnet := nachzug(t, []domain.Field{
		{Name: "eins", Skript: `"[" + value + "]"`},
		{Name: "fest", Skript: `if card.eins == "" then "leer" else "voll"`},
		{Name: "danach", Skript: `card.fest + value`},
	}, map[string]string{"eins": "Funke", "fest": "voll", "danach": "x"}, "eins")

	if werte["danach"] != "x" || slices.Contains(gerechnet, "danach") || slices.Contains(gerechnet, "fest") {
		t.Errorf("danach ist %q, gerechnet %v", werte["danach"], gerechnet)
	}
}

// **Jedes Feld einmal je Runde** — ein Kreis endet, statt ewig zu laufen.
func TestEinKreisEndet(t *testing.T) {
	t.Parallel()

	werte, gerechnet := nachzug(t, []domain.Field{
		{Name: "a", Skript: `card.b + "a"`},
		{Name: "b", Skript: `card.a + "b"`},
	}, map[string]string{"a": "", "b": ""}, "a")

	if werte["a"] != "a" || werte["b"] != "ab" || !slices.Equal(gerechnet, []string{"a", "b"}) {
		t.Errorf("a %q, b %q, gerechnet %v", werte["a"], werte["b"], gerechnet)
	}
}

// **Ein Feld ohne Skript zieht seine Leser trotzdem nach** — geändert hat es
// sich durch die Eingabe.
func TestEinFeldOhneSkriptZiehtSeineLeserNach(t *testing.T) {
	t.Parallel()

	werte, gerechnet := nachzug(t, []domain.Field{
		{Name: "roh"},
		{Name: "leser", Skript: `card.roh + "!"`},
	}, map[string]string{"roh": "Funke", "leser": "alt"}, "roh")

	if werte["leser"] != "Funke!" || !slices.Equal(gerechnet, []string{"roh", "leser"}) {
		t.Errorf("leser ist %q, gerechnet %v", werte["leser"], gerechnet)
	}
}

// **Ein Skript, das scheitert, lässt den Wert stehen und sagt es** — wie
// `Value::update`, das den Fehler je Feld fängt.
func TestEinScheiterndesSkriptLaesstDenWertUndSagtEs(t *testing.T) {
	t.Parallel()

	felder := []domain.Field{{Name: "eins", Skript: `gibt_es_nicht(value)`}}
	gefuellt := []msescript.Eintrag{{Name: "eins", Wert: msescript.Text("Funke")}}
	umgebung := msescript.Bibliothek().Darueber()

	stand, gerechnet, befunde := nachziehen(context.Background(), gefuellt, felder, umgebung, nil,
		func([]msescript.Eintrag) {}, []string{"eins"})

	if stand[0].Wert.AlsText() != "Funke" || len(befunde) != 1 || befunde[0].Ort != "eins" {
		t.Errorf("Wert %q, Befunde %v", stand[0].Wert.AlsText(), befunde)
	}

	if !slices.Equal(gerechnet, []string{"eins"}) {
		t.Errorf("gerechnet %v", gerechnet)
	}
}
