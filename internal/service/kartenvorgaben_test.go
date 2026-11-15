package service_test

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Feld, das die Karte nicht führt, hat den Vorgabewert des Spiels — nicht
// `nil`** (#276).
//
// # Der Befund
//
// `magic-m15.mse-style` legt den Kartennamen so an:
//
//	name:
//	    left: { if card.card_symbol == "none" then 32 else 50 }
//
// Eine Karte ohne Symbol führt kein Feld `card symbol`. Damit war
// `card.card_symbol` gleich `nil`, `nil == "none"` falsch, und der Name begann
// bei **50** statt bei **32** — verglichen an einer echten MSE-Ausgabe
// derselben Auflösung.
//
// **Das ist keine Feinheit an einem Feld.** Jede Stelle, die `card.<feld>` gegen
// eine Zeichenkette prüft, entscheidet ohne Vorgabewerte falsch herum, und
// `magic.mse-game` führt 249 Kartenfelder.
//
// # Woher der Wert kommt
//
// Aus dem `card field:`-Block des Spielpakets, in dieser Reihenfolge:
//
//  1. `default:`, wenn es dasteht — 142 der 249 Felder haben eines,
//  2. sonst `initial:`, der Anfangswert eines Auswahlfelds — 17-mal im Pack,
//  3. sonst die **erste** `choice:` eines Auswahlfelds — `card symbol` bekommt
//     so `none`, und genau danach fragt der Stil,
//  4. sonst der leere Text.
func TestEinFehlendesFeldBekommtDenVorgabewert(t *testing.T) {
	felder := []domain.Field{
		{Name: "card symbol", Kind: domain.KindChoice, Choices: []string{"none", "tombstone"}},

		// **`initial:` schlägt die erste Wahl.** In `magic.mse-game` steht am
		// Feld `rarity` ausdrücklich `initial: common`, und seine erste Wahl
		// heißt `basic land` — ohne diese Stufe trägt jede Karte ohne
		// Seltenheit ein Landsymbol
		// (#341).
		{
			Name: "seltenheit", Kind: domain.KindChoice, Initial: "common",
			Choices: []string{"basic land", "common", "rare"},
		},
		// **In Anführungszeichen, denn ein `default:` ist Skript.**
		//
		// Hier stand bis zum 2026-08-16 `Default: "common"` — eine Form, die in
		// `magic.mse-game` **nirgends** vorkommt: alle 142 Vorgabewerte der
		// Kartenfelder sind Skript (`stylesheet; …`), und `set_fields` schreibt
		// seine Zeichenketten in Anführungszeichen. Die Vorrichtung war
		// großzügiger als das Format
		// (#298).
		{Name: "rarity", Default: `"common"`},
		{Name: "border color", Default: `{ "rgb(0,0,0)" }`},
		{Name: "ohne alles"},
	}

	faelle := []struct {
		name    string
		quelle  string
		erwartt string
	}{
		// **Der gemessene Fall.**
		{"die erste Wahl eines Auswahlfelds", `card.card_symbol`, "none"},
		{"ein einfacher Vorgabewert", `card.rarity`, "common"},
		{"der Anfangswert schlägt die erste Wahl", `card.seltenheit`, "common"},

		// **Ein Vorgabewert darf geklammert sein.** Beide Formen kommen vor;
		// geklammert ist die seltenere.
		{"ein Vorgabewert als Skript", `card.border_color`, "rgb(0,0,0)"},

		{"ohne alles bleibt es leer", `card.ohne_alles`, ""},

		// **Der Wert der Karte schlägt die Vorgabe.** Andernfalls wäre die
		// Vorgabe keine Vorgabe, sondern eine Festlegung.
		{"die Karte gewinnt", `card.rarity_gefuellt`, "rare"},
	}

	karte := &domain.CardRow{ID: 1, Number: 1, Values: []domain.Value{
		{Field: "rarity gefuellt", Text: "rare"},
	}}

	for _, fall := range faelle {
		t.Run(fall.name, func(t *testing.T) {
			feld := msepaket.Feldstil{
				Feld: "name", Links: "0", Oben: "0", Breite: "375", Hoehe: "20",
				Bild: "{ " + fall.quelle + ` + ".png" }`,
			}

			vorlage := service.Vorlage{
				Stil:         bildstil(feld),
				Kartenfelder: append(felder, domain.Field{Name: "rarity gefuellt"}),
			}

			plan, befunde := gedeutetMitVorlage(t, karte, vorlage)
			if len(befunde) > 0 {
				t.Fatalf("Befunde: %v", befunde)
			}

			if bild := ebenen(t, plan)[0].Bild; bild.Pfad != "/probe.mse-style/"+fall.erwartt+".png" {
				t.Errorf("%q, erwartet war %q", bild.Pfad, fall.erwartt+".png")
			}
		})
	}
}

// **Ein Vorgabewert, dessen Skript scheitert, wird zum leeren Text.**
//
// Nicht zu einem Fehler: ein Spielpaket mit 249 Feldern, von denen eines ein
// Skript trägt, das dieses Programm nicht ausführt, muss die anderen 248
// liefern (§8). Der Befund entsteht dort, wo der Wert gebraucht wird.
func TestEinUnausfuehrbarerVorgabewertWirdLeer(t *testing.T) {
	feld := msepaket.Feldstil{
		Feld: "name", Links: "0", Oben: "0", Breite: "375", Hoehe: "20",
		Bild: `{ card.kaputt + ".png" }`,
	}

	vorlage := service.Vorlage{
		Stil:         bildstil(feld),
		Kartenfelder: []domain.Field{{Name: "kaputt", Default: "{ gibtsnicht() }"}},
	}

	plan, befunde := gedeutetMitVorlage(t, kartemit("x", "y"), vorlage)

	// **Leer, und gemeldet** — der Befund kam bis
	// #295
	// nicht, und der leere Dateiname unten war alles, was davon zu sehen war.
	if len(befunde) != 1 || !strings.Contains(befunde[0].Meldung, "gibtsnicht") {
		t.Fatalf("Befunde: %v", befunde)
	}

	if bild := ebenen(t, plan)[0].Bild; bild.Pfad != "/probe.mse-style/.png" {
		t.Errorf("%q", bild.Pfad)
	}
}

// **Ein gescheiterter Vorgabewert wird gemeldet, nicht verschwiegen.**
//
// # Was das verdeckt hat
//
// `magic.mse-game` gibt dem Feld `separator` den Vorgabewert
// `separator_default(value, margin_code: "sep1")`, und der ergibt `none` — die
// Vorlage baut daraus den Dateinamen `…/separators/none.png`. Scheiterte das
// Skript, stand dort `…/separators/.png`, und der Befund kam erst beim
// **Rastern**: „das Bild ist in keinem Paket zu finden". Die eigentliche
// Ursache — ein Name, den die Umgebung nicht führt — war nirgends zu lesen
// (#295).
//
// **Der Wert bleibt trotzdem leer.** Ein Vorgabewert, der nicht zu rechnen ist,
// hat kein Ergebnis; erfunden wird keines. Gemeldet wird, dass er fehlt.
func TestEinGescheiterterVorgabewertWirdGemeldet(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Felder: []msepaket.Feldstil{{
				Feld: "name", Links: "0", Oben: "0", Breite: "50", Hoehe: "10",
				Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "8"},
			}},
		},
		Kartenfelder: []domain.Field{
			{Name: "name", Kind: domain.KindText},
			{Name: "trenner", Kind: domain.KindChoice, Default: "nirgends_bekannt()"},
		},
	}

	karte := &domain.CardRow{ID: 1, Values: []domain.Value{{Field: "name", Text: "Feuerball"}}}

	_, befunde, err := service.NeuerPlanbauer().Stilplan(t.Context(), karte, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	gefunden := false

	for _, befund := range befunde {
		if strings.Contains(befund.Ort, "trenner") && strings.Contains(befund.Meldung, "nirgends_bekannt") {
			gefunden = true
		}
	}

	if !gefunden {
		t.Errorf("kein Befund zum Vorgabewert: %v", befunde)
	}
}

// **Ein Vorgabeskript darf die Funktionen des Spiels benutzen.**
//
// # Die Reihenfolge, nachgelesen in MSE2
//
//	value = field().default_script.invokeOnDefault(ctx, value)
//	      | field().        script.invokeOn(ctx, value);
//
// `TextValue::update`, `src/data/field/text.cpp`, Zeilen 178–179, Fassung
// f9d9356 vom 2021-07-06. **Beides läuft in `update`** — also lange nachdem
// der Skriptverwalter die `init script:` des Spiels und des Stils ausgeführt
// hat (`ScriptManager::updateAll`, `src/script/script_manager.cpp`).
//
// # Was die falsche Reihenfolge gekostet hat
//
// `magic.mse-game` legt `separator_default`, `additional_credit_default` und
// zwei Dutzend weitere im Initskript an. Liefen die Vorgabewerte davor, war
// **jeder** dieser Namen unbekannt — und weil ein gescheiterter Vorgabewert
// lautlos leer wurde, sah man davon nur den Dateinamen
// `…/separators/.png` im Rasterbefund
// (#295).
func TestEinVorgabeskriptDarfDieFunktionenDesSpielsBenutzen(t *testing.T) {
	vorlage := service.Vorlage{
		Stil: msepaket.Stil{
			Spiel: "probe", Breite: 100, Hoehe: 100,
			Felder: []msepaket.Feldstil{{
				Feld: "trenner", Links: "0", Oben: "0", Breite: "50", Hoehe: "10",
				Schrift: msepaket.Schrift{Name: "Helvetica", Groesse: "8"},
			}},
		},
		Spielskript:  `trenner_vorgabe := { "keiner" }`,
		Kartenfelder: []domain.Field{{Name: "trenner", Default: "trenner_vorgabe()"}},
	}

	plan, befunde, err := service.NeuerPlanbauer().
		Stilplan(t.Context(), &domain.CardRow{ID: 1}, vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	if len(befunde) > 0 {
		t.Fatalf("Befunde: %v", befunde)
	}

	if len(plan.Ebenen) < 2 || plan.Ebenen[1].Klartext() != "keiner" {
		t.Errorf("gezeichnet wird %q, erwartet den Vorgabewert",
			plan.Ebenen[len(plan.Ebenen)-1].Klartext())
	}
}
