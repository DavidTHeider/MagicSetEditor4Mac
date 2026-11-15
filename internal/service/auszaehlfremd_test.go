//go:build fremdpakete

package service_test

// **Die fünfzig Dimensionen von `magic.mse-game`, wirklich gerechnet** —
// #1338.
//
// **Hinter `fremdpakete`** (§3, §11): das Spielpaket liegt nicht im
// Repositorium, und dieser Lauf zählt nicht zur Abdeckung.
//
// # Warum er gefahren wird
//
// Weil eine selbst geschriebene Probe sagt, dass **ein** Skript läuft — nicht,
// dass die fünfzig laufen, die es wirklich gibt. Sie rufen ausschließlich
// Funktionen aus `statistics_script` (2.058 Zeilen), das über `init script:`
// hereinkommt; ein einziger fehlender Name nähme jede von ihnen mit.
//
//	go test -tags fremdpakete ./internal/service/ -run Auszaehl -v

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// echtemagickarte baut eine Karte mit Feldwerten, wie sie in einem echten Set
// stehen — eine weiße Kreatur mit zwei Manasymbolen.
func echtemagickarte() *domain.CardRow {
	return &domain.CardRow{
		ID: 1, Number: 0,
		Values: []domain.Value{
			{Field: "name", Text: "Prüfkarte"},

			// **`card color` und nicht `casting cost`.** `color_category_statistic`
			// liest `card.card_color` (`statistics_script`, Zeile 216) — die
			// **Rahmenfarbe**, eine Mehrfachwahl, nicht die Manakosten. Der erste
			// Anlauf setzte nur `casting cost: 1W` und bekam `Colorless`; das war
			// richtig gerechnet und falsch erwartet (§15).
			{Field: "card color", Text: "white"},
			{Field: "casting cost", Text: "1W"},
			{Field: "type", Text: "Creature — Human"},
			{Field: "rarity", Text: "common"},
			{Field: "power", Text: "2"},
			{Field: "toughness", Text: "2"},
			{Field: "rule text", Text: "Vigilance"},
		},
	}
}

func TestDieEchtenAuszaehldimensionenLaufen(t *testing.T) {
	paket, fehler := msepaket.Oeffne(
		"/Users/david/full-magic-pack/data/magic.mse-game")
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if fehler := paket.ModuleEinbinden(msepaket.NeuerPaketleser(
		"/Users/david/full-magic-pack/data")); fehler != nil {
		t.Fatalf("Module einbinden: %v", fehler)
	}

	spiel, _ := paket.Spielmodell()

	dimensionen := paket.Auszaehldimensionen()
	if len(dimensionen) != 50 {
		t.Fatalf("%d Dimensionen, erwartet waren 50", len(dimensionen))
	}

	karten := []*domain.CardRow{echtemagickarte()}

	geglueckt, gescheitert := 0, []string{}

	// **Dieselbe Abbildung wie im Betrieb**, seit dem 2026-08-31.
	//
	// Hier stand eine von Hand gebaute Angabe, und sie ließ `Globalskript`
	// weg — die Dimension `average opening hand` scheiterte deshalb weiter,
	// nachdem der Weg dafür längst stand. Eine Vorrichtung, die weniger trägt
	// als der Betrieb, prüft den Betrieb nicht (§3).
	for _, angabe := range service.Auszaehlangaben(dimensionen) {
		_, befunde := service.Auszaehlwerte(
			t.Context(), karten, magicvorlage(spiel), nil, angabe)
		if len(befunde) > 0 {
			gescheitert = append(gescheitert, angabe.Name+": "+befunde[0].Meldung)

			continue
		}

		geglueckt++
	}

	for _, meldung := range gescheitert {
		t.Logf("%s", meldung)
	}

	t.Logf("%d von %d Dimensionen gerechnet", geglueckt, len(dimensionen))

	// **50 von 50** — gemessen am 2026-08-31.
	//
	// Als **Gleichheit** und nicht als Untergrenze: eine Untergrenze bliebe
	// grün, wenn ein Fehler zehn weitere mitnähme.
	//
	// **Es waren 46 am Morgen dieses Tages**, und die vier sind an einem Tag
	// gefallen:
	//
	//   - die beiden `land drop hit %` mit
	//     #1356:
	//     `set.cards` steht jetzt in der Auszählung, und ein Name, der eine
	//     Zahl ist, meint eine Stelle;
	//   - `average opening hand` mit
	//     #1355:
	//     `global_value` ist das Ergebnis von `global_script:`, einem
	//     Schlüssel, den **MSE 2.1.2 gar nicht kennt**;
	//   - `keywords` mit
	//     #1354:
	//     `keyword_usage` liest jetzt einen Vermerk, den `expand_keywords` je
	//     Karte führt.
	//
	// **Was dieser Lauf nicht sagt, ist ob sie richtig rechnen.** Er zählt,
	// dass keine mit einem Fehler abbricht. Für drei der vier gibt es
	// daneben einen Vergleich gegen das Orakel; für `average opening hand`
	// nicht, und der Grund steht dort.
	if geglueckt != 50 {
		t.Errorf("%d von %d gerechnet, gemessen waren 50 — %d scheiterten",
			geglueckt, len(dimensionen), len(gescheitert))
	}
}

// **Und die Werte selbst — „color category" für eine weiße Karte.**
//
// Die Abnahme aus dem Issue: derselbe Wert wie MSE2. Verglichen wird gegen
// das **Orakel** (§15), nicht gegen eine erratene Zeichenkette.
func TestDieFarbgruppeEinerWeissenKarte(t *testing.T) {
	paket, fehler := msepaket.Oeffne(
		"/Users/david/full-magic-pack/data/magic.mse-game")
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if fehler := paket.ModuleEinbinden(msepaket.NeuerPaketleser(
		"/Users/david/full-magic-pack/data")); fehler != nil {
		t.Fatalf("Module einbinden: %v", fehler)
	}

	spiel, _ := paket.Spielmodell()

	var farbgruppe msepaket.Auszaehldimension

	for _, eine := range paket.Auszaehldimensionen() {
		if eine.Name == "color category" {
			farbgruppe = eine
		}
	}

	if farbgruppe.Skript == "" {
		t.Fatal("die Dimension color category fehlt")
	}

	werte, befunde := service.Auszaehlwerte(
		t.Context(), []*domain.CardRow{echtemagickarte()},
		magicvorlage(spiel), nil,
		service.Auszaehlangabe{
			Name: farbgruppe.Name, Skript: farbgruppe.Skript,
			Zerlegen: farbgruppe.Zerlegen, LeereZeigen: farbgruppe.LeereZeigen,
		})

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	t.Logf("color category einer weißen Kreatur: %+v", werte)

	if len(werte) != 1 || werte[0] != "White" {
		t.Errorf("die Farbgruppe lautet %+v, erwartet [White]", werte)
	}
}

// magicvorlage baut die Vorlage, die eine Dimension zum Rechnen braucht.
//
// **Ohne Stilpaket** — das ist der Unterschied zum Betrieb, und er ist
// benannt: `belegeVorlage` setzt `stylesheet` dann aus einem leeren
// `msepaket.Stil`. Die `default:`-Skripte laufen trotzdem, denn `stylesheet;`
// ist eine Abhängigkeitsangabe und kein Wert, der gelesen wird. Der Lauf
// gegen das Orakel prüft, ob das reicht.
func magicvorlage(spiel domain.Game) service.Vorlage {
	return service.Vorlage{
		Kartenfelder: spiel.Fields,
		Setfelder:    spiel.SetFields,
		Spielskript:  spiel.Skript,
	}
}

// mitWortliste hängt die Schlüsselwörter des Pakets an eine Vorlage.
//
// **Ohne sie trifft kein Schlüsselwort**, und die Dimension `keywords` gäbe
// für **jede** der 3.761 Karten eine leere Liste — bei grünem Lauf. Genau die
// Vorrichtung, die weniger trägt als der Betrieb (§3): `spielquelle.go` baut
// die Liste dort aus `paket.Schluesselwoerter()`, und dieser Lauf tat es
// nicht.
//
// Gefunden am 2026-08-31 daran, dass der Vergleich gegen das Orakel „0 von
// 3761 Karten tragen Schlüsselwörter" meldete
// (#1354).
func mitWortliste(
	t *testing.T, vorlage service.Vorlage, paket *msepaket.Paket,
) service.Vorlage {
	t.Helper()

	woerter, befunde := service.NeueWortliste(
		paket.Schluesselwoerter(), paket.Schluesselwortparameterarten())
	if len(befunde) > 0 {
		t.Logf("beim Bauen der Wortliste: %+v", befunde)
	}

	vorlage.Schluesselwoerter = woerter

	return vorlage
}
