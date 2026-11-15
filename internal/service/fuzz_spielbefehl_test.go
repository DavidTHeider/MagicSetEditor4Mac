package service_test

// **Der Spielbefehl als Ganzes, so wie er vom Draht kommt** — QA-06c
// (#1976),
// Fläche 76 aus `docs/FUZZZIELE.md`.
//
// # Was es neben `FuzzZugfolge` (77) zu sehen gibt
//
// **Alle fünfzig Arten und die freie Nutzlast.** `FuzzZugfolge` baut zwölf
// Arten mit einer festen Nutzlast aus drei Byte. Ein Gast schickt dagegen
// rohes JSON, und `Partiewirt.befehlen` nimmt daraus jede Art und jede
// `Nutzlast` — Zahlen als Text, Listen statt Abbildungen, verschachtelte
// Werte. Hier wird es genau so gelesen: `json.Unmarshal` in einen
// [domain.Spielbefehl], Partie und Absender vom Wirt gesetzt, dann ausgelegt.
//
// **`$OBJ`, `$OBJ2` und `$SITZ` im rohen Text** werden vor dem Lesen durch
// Kennungen der Partie ersetzt — sonst landete fast jeder Lauf bei „gibt es
// nicht", und der Fuzzer käme an keinen Ausleger heran.
//
// # Was ein Fund ist — festgelegt **vor** dem ersten Lauf
//
// Dieselben Zusagen wie bei `FuzzZugfolge` (`zugschritt`): die Partie
// besteht danach `Pruefe()`, die Eingabe bleibt unberührt, zweimal derselbe
// Befehl gibt dasselbe, jede Ablehnung trägt eine der drei Klassen — dazu
// Panic und Hänger.
//
//	go test ./internal/service -run Fuzz -fuzz FuzzSpielbefehl -fuzztime 2m -fuzzminimizetime 5s

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// alleBefehlsarten sind die fünfzig Arten des Sandkastens.
var alleBefehlsarten = []string{
	service.BefehlSpielstein, service.BefehlEmblem, service.BefehlBeherrschen,
	service.BefehlUmwandeln, service.BefehlKopieren, service.BefehlHeimholen,
	service.BefehlMonarch, service.BefehlInitiative, service.BefehlStrahlung,
	service.BefehlLegende, service.BefehlReihung, service.BefehlStarthand,
	service.BefehlMulligan, service.BefehlUnterlegen, service.BefehlBehalten,
	service.BefehlZerstoeren, service.BefehlOpfern, service.BefehlExilieren,
	service.BefehlNeutralisieren, service.BefehlKaempfen, service.BefehlMahlen,
	service.BefehlHellsehen, service.BefehlUeberwachen, service.BefehlMuenzwurf,
	service.BefehlWuerfeln, service.BefehlTappen, service.BefehlEnttappen,
	service.BefehlUmdrehen, service.BefehlSchieben, service.BefehlZaehler,
	service.BefehlAnhaengen, service.BefehlLoesen, service.BefehlHaeufen,
	service.BefehlEnthaeufen, service.BefehlBlick, service.BefehlMischen,
	service.BefehlPassen, service.BefehlAbwerfen, service.BefehlLeben,
	service.BefehlVergiftung, service.BefehlSchaden, service.BefehlAbkuerzen,
	service.BefehlZugende, service.BefehlMana, service.BefehlLandmana,
	service.BefehlWirken, service.BefehlLandspielen, service.BefehlAngreifer,
	service.BefehlBlocker, service.BefehlAufgeben,
}

func FuzzSpielbefehl(f *testing.F) {
	nutzlasten := []string{
		`{"objekt":"$OBJ","um":1,"anzahl":2,"zone":"friedhof","art":"G","wert":3,"ziel":"$SITZ","auf":"$OBJ2"}`,
		`{"objekt":"$OBJ","um":"eins","anzahl":-5,"zone":7}`,
		`{"angriffe":[{"kreatur":"$OBJ","ziel":"$SITZ"}],"bloecke":[{"kreatur":"$OBJ2","angreifer":"$OBJ"}]}`,
		`{"objekt":["$OBJ"],"um":1e308,"anzahl":9007199254740993}`,
		`null`,
	}

	for _, art := range alleBefehlsarten {
		for _, nutzlast := range nutzlasten {
			f.Add(byte(0), []byte(`{"Kennung":"k","Art":"`+art+`","Erwartet":0,"Nutzlast":`+nutzlast+`}`))
		}
	}

	f.Add(byte(1), []byte(`{"Kennung":"","Art":""}`))
	f.Add(byte(2), []byte(`kein JSON`))

	f.Fuzz(func(t *testing.T, wahl byte, roh []byte) {
		partie := zugpartieIn(t, wahl)
		if len(partie.Objekte) < 2 {
			t.Skip("die Partie hat zu wenig Objekte")
		}

		sitz := partie.Sitze[int(wahl)%len(partie.Sitze)].Kennung
		text := strings.NewReplacer(
			"$OBJ2", partie.Objekte[1].Kennung,
			"$OBJ", partie.Objekte[0].Kennung,
			"$SITZ", partie.Sitze[(int(wahl)+1)%len(partie.Sitze)].Kennung,
		).Replace(string(roh))

		var befehl domain.Spielbefehl
		if json.Unmarshal([]byte(text), &befehl) != nil {
			return
		}

		// **Wie beim Wirt**: Partie und Absender stehen fest, was in der
		// Nachricht steht, ist eine Behauptung.
		befehl.PartieID = partie.ID
		befehl.Absender = sitz

		if befehl.Pruefe() != nil {
			return
		}

		befehlsschritt(t, partie, befehl)
	})
}

// ordnendeMischung mischt nicht und prüft wie [service.EchterZufall] — eine
// negative Menge und eine Grenze unter eins sind abgewiesen, mit Klasse.
//
// **Nicht `festeMischung{}`**: deren leere Reihe passt zu keiner Bibliothek
// und gibt einen Fehler ohne Klasse. Der erste Lauf meldete das als Fund bei
// `spieler.starthand` — es war die Vorrichtung, nicht der Ausleger.
type ordnendeMischung struct{}

func (ordnendeMischung) Mische(anzahl int) ([]int, error) {
	if anzahl < 0 {
		return nil, fmt.Errorf("%w: keine Mischung über %d", domain.ErrValidation, anzahl)
	}

	reihe := make([]int, anzahl)
	for i := range reihe {
		reihe[i] = i
	}

	return reihe, nil
}

func (ordnendeMischung) Zahl(grenze int) (int, error) {
	if grenze <= 0 {
		return 0, fmt.Errorf("%w: keine Zahl unter %d", domain.ErrValidation, grenze)
	}

	return 0, nil
}

// befehlsschritt legt einen Befehl aus und hält dieselben Zusagen wie
// `zugschritt` — mit einer Mischung, die zu jeder Bibliothek passt.
func befehlsschritt(t *testing.T, partie domain.Partie, befehl domain.Spielbefehl) {
	t.Helper()

	vorher := alsJSON(t, partie)

	ereignisse, danach, fehler := service.NeuerSandkasten(ordnendeMischung{}).Auslegen(partie, befehl)
	nochmalEreignisse, nochmal, nochmalFehler := service.NeuerSandkasten(ordnendeMischung{}).Auslegen(partie, befehl)

	if nachher := alsJSON(t, partie); nachher != vorher {
		t.Fatalf("%s hat seine Eingabe verändert (%v)", befehl.Art, fehler)
	}

	if (fehler == nil) != (nochmalFehler == nil) || alsJSON(t, danach) != alsJSON(t, nochmal) ||
		alsJSON(t, ereignisse) != alsJSON(t, nochmalEreignisse) {
		t.Fatalf("%s zweimal ausgelegt, zwei Ergebnisse: %v / %v", befehl.Art, fehler, nochmalFehler)
	}

	if fehler != nil {
		if !errors.Is(fehler, domain.ErrValidation) && !errors.Is(fehler, domain.ErrNotFound) &&
			!errors.Is(fehler, domain.ErrConflict) {
			t.Errorf("%s abgelehnt ohne Klasse: %v", befehl.Art, fehler)
		}

		return
	}

	zugstandPruefen(t, befehl, partie, danach)
}
