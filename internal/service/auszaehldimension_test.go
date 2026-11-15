package service_test

// **Eine erklärte Dimension über ein Set rechnen** — Werkstück C von Phase
// 4.11 (#1338).
//
// # Was am 2026-08-31 nachgelesen wurde (§15)
//
// `StatsPanel::showCategory` (`src/gui/set/stats_panel.cpp`, Zeilen 477–507,
// Fassung f9d9356 vom 2021-07-06):
//
//	String value = untag(dim->script.invoke(ctx)->toString());
//	e->values.push_back(value);
//	if (value.empty() && !dim->show_empty) { show = false; break; }
//	…
//	if (dim->split_list) d.splitList(dim_id);
//
// Vier Sätze, und jeder ist hier ein Fall: **ausgezeichnet wird entfernt**,
// **leer fällt raus** (außer `show empty:`), **Kommas zerlegen** (bei
// `split list:`), und ein **Skriptfehler** nimmt die Karte heraus statt den
// Lauf abzubrechen.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// auszaehlkarten baut Karten mit einem Feld `farbe`.
func auszaehlkarten(werte ...string) []*domain.CardRow {
	karten := make([]*domain.CardRow, 0, len(werte))

	for stelle, wert := range werte {
		karten = append(karten, &domain.CardRow{
			ID: int64(stelle + 1), Number: stelle,
			Values: []domain.Value{{Field: "farbe", Text: wert}},
		})
	}

	return karten
}

var auszaehlfelder = []domain.Field{{Name: "farbe", Kind: domain.KindText}}

// auszaehlprobe baut eine Vorlage mit dem einen Feld und einem Spielskript.
func auszaehlprobe(spielskript string) service.Vorlage {
	return service.Vorlage{Kartenfelder: auszaehlfelder, Spielskript: spielskript}
}

// **Der Wert kommt aus dem Skript, je Karte.**
func TestEineDimensionRechnetIhrenWertJeKarte(t *testing.T) {
	t.Parallel()

	werte, befunde := service.Auszaehlwerte(t.Context(),
		auszaehlkarten("weiss", "rot", "weiss"), auszaehlprobe(""), nil,
		service.Auszaehlangabe{Name: "farbe", Skript: "card.farbe"})

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	if len(werte) != 3 || werte[0] != "weiss" || werte[1] != "rot" {
		t.Errorf("die Werte lauten %+v", werte)
	}
}

// **Auszeichnungen fallen weg** — MSEs `untag`.
//
// Ohne sie zählte `<i>rot</i>` als eigene Gruppe neben `rot`, und dieselbe
// Farbe stünde zweimal im Diagramm.
func TestEineDimensionEntferntDieAuszeichnung(t *testing.T) {
	t.Parallel()

	werte, _ := service.Auszaehlwerte(t.Context(),
		auszaehlkarten("<i>rot</i>"), auszaehlprobe(""), nil,
		service.Auszaehlangabe{Name: "farbe", Skript: "card.farbe"})

	if len(werte) != 1 || werte[0] != "rot" {
		t.Errorf("die Werte lauten %+v", werte)
	}
}

// **Eine Karte ohne Wert fällt heraus** — außer `show empty:` steht da.
func TestEineKarteOhneWertFaelltHeraus(t *testing.T) {
	t.Parallel()

	angabe := service.Auszaehlangabe{Name: "farbe", Skript: "card.farbe"}

	werte, _ := service.Auszaehlwerte(t.Context(),
		auszaehlkarten("rot", ""), auszaehlprobe(""), nil, angabe)
	if len(werte) != 1 {
		t.Errorf("ohne `show empty:` blieben %d Werte: %+v", len(werte), werte)
	}

	angabe.LeereZeigen = true

	mitLeeren, _ := service.Auszaehlwerte(t.Context(),
		auszaehlkarten("rot", ""), auszaehlprobe(""), nil, angabe)
	if len(mitLeeren) != 2 {
		t.Errorf("mit `show empty:` blieben %d Werte: %+v", len(mitLeeren), mitLeeren)
	}
}

// **`split list:` zerlegt an Kommas — und überspringt das Leerzeichen.**
//
// `GraphDataPre::splitList` (`src/gui/control/graph.cpp`, Zeilen 42–60,
// Fassung f9d9356 vom 2021-07-06):
//
//	if (is_substr(v, comma, ", ")) ++comma; // skip space after it
//
// Eine zweifarbige Karte zählt damit für **beide** Farben, und `", "` gibt
// nicht ein Glied, das mit einem Leerzeichen anfängt.
func TestSplitListZerlegtAnKommas(t *testing.T) {
	t.Parallel()

	werte, _ := service.Auszaehlwerte(t.Context(),
		auszaehlkarten("weiss, blau,rot"), auszaehlprobe(""), nil,
		service.Auszaehlangabe{Name: "farbe", Skript: "card.farbe", Zerlegen: true})

	if len(werte) != 3 {
		t.Fatalf("%d Werte: %+v", len(werte), werte)
	}

	for stelle, soll := range []string{"weiss", "blau", "rot"} {
		if werte[stelle] != soll {
			t.Errorf("Wert %d ist %q, erwartet %q", stelle, werte[stelle], soll)
		}
	}
}

// **Ein Skriptfehler nimmt die Karte heraus und meldet sich.**
//
// MSE fängt ihn je Karte (`catch (ScriptError const& e)`) und setzt
// `show = false`. Ein Lauf, der an der ersten kaputten Karte abbräche, gäbe
// für ein Set mit 300 guten Karten gar nichts (§8).
func TestEinSkriptfehlerNimmtNurSeineKarteHeraus(t *testing.T) {
	t.Parallel()

	werte, befunde := service.Auszaehlwerte(t.Context(),
		auszaehlkarten("rot", "blau"), auszaehlprobe(""), nil,
		service.Auszaehlangabe{Name: "farbe", Skript: "gibtesnicht(card)"})

	if len(werte) != 0 {
		t.Errorf("es blieben %d Werte: %+v", len(werte), werte)
	}

	// **Ein Befund und nicht zwei je Karte.** Zweihundert gleiche Meldungen
	// sind keine Auskunft; was zählt, ist die Zahl der betroffenen Karten.
	if len(befunde) != 1 {
		t.Fatalf("%d Befunde: %+v", len(befunde), befunde)
	}
}

// **Ein Set ohne Karten gibt eine leere Liste und keinen Fehler.**
func TestOhneKartenGibtEsKeineWerte(t *testing.T) {
	t.Parallel()

	werte, befunde := service.Auszaehlwerte(t.Context(), nil, auszaehlprobe(""), nil,
		service.Auszaehlangabe{Name: "farbe", Skript: "card.farbe"})

	if werte == nil || len(werte) != 0 || len(befunde) != 0 {
		t.Errorf("Werte %+v, Befunde %+v", werte, befunde)
	}
}

// **Die Funktionen des Spielskripts stehen bereit.**
//
// `magic.mse-game` erklärt `color_category_statistic()` in seinem
// `init script:`; ohne es ist jede der fünfzig Dimensionen ein unbekannter
// Name.
func TestDasSpielskriptStehtDerDimensionZurVerfuegung(t *testing.T) {
	t.Parallel()

	werte, befunde := service.Auszaehlwerte(t.Context(),
		auszaehlkarten("rot"),
		auszaehlprobe(`farbgruppe_von := { "Gruppe " + card.farbe }`), nil,
		service.Auszaehlangabe{Name: "farbe", Skript: "farbgruppe_von()"})

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	if len(werte) != 1 || werte[0] != "Gruppe rot" {
		t.Errorf("die Werte lauten %+v", werte)
	}
}

// **`card.notes` steht auch hier an jeder Karte.**
//
// # Der Befund, der diesen Fall verlangt
//
// Gemessen am 2026-08-31 gegen `magic.mse-game`: drei der fünfzig Dimensionen
// — `card notes words`, `card notes clauses` und `average opening hand` —
// scheiterten mit „die Abbildung führt kein Glied `notes`". Der Grund war
// nicht die Auszählung: [listenumgebung] setzte `notes` **nie**, während
// [kartenumgebungVorbereiten] es seit [mitNotizen] tut.
//
// **Das trifft die Kartenliste genauso**: eine gerechnete Spalte, die
// `card.notes` liest, scheiterte dort ebenso still. Behoben ist deshalb die
// Klasse und nicht der Fall (§10).
func TestCardNotesStehtAnJederKarte(t *testing.T) {
	t.Parallel()

	werte, befunde := service.Auszaehlwerte(t.Context(),
		auszaehlkarten("rot"), auszaehlprobe(""), nil,
		service.Auszaehlangabe{
			Name: "notiz", Skript: `card.notes`, LeereZeigen: true,
		})

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	if len(werte) != 1 || werte[0] != "" {
		t.Errorf("die Werte lauten %+v", werte)
	}
}

// **Das Spielskript sieht die erklärten Kartenfelder — leer, aber vorhanden.**
//
// # Der Befund, der diesen Fall verlangt
//
// `spielumgebung` setzte `card` als **leere** Abbildung. Ein Initskript, das
// beim Anlegen eines Namens `card.<feld>` liest, bricht daran ab — ein
// fehlendes **Glied** lässt sich in MSE-Script nicht mit `or else` fangen —,
// und der Name entsteht nie.
//
// **Gemessen am 2026-08-31 an `magic.mse-game`**: Zeile 13768 liest
// `card.rule_text`, und damit fiel `land_to_color` aus. Sechs von 3.761
// Karten von `White.mse-set` bekamen deshalb für ihre Rückseite `Colorless`
// statt `Land` — alles zweiseitige Karten mit einem Land hinten.
//
// **Für `set` stand die Regel längst da** (`leereAbbildung(setfelder)`), für
// `card` nicht. Es ist dieselbe Regel, und sie gilt für die Kartenliste
// genauso: sie fährt dasselbe Spielskript.
func TestDasSpielskriptSiehtDieErklaertenKartenfelder(t *testing.T) {
	t.Parallel()

	// **Dieselbe Gestalt wie im Pack**: ein Block mit `@`-Bindung, deren
	// Argument beim **Anlegen** ausgewertet wird. `magic.mse-game` schreibt
	// in Zeile 13768 genau das: `}@(rules:card.rule_text)`.
	//
	// Ein `card.farbe` in einer eigenen Zeile reichte hier **nicht**: die
	// Vorbelegung macht nach einem gescheiterten Satz weiter, und der
	// nächste Name entstünde trotzdem. Erst die Bindung nimmt den Namen mit,
	// zu dem sie gehört — und genau das ist im Pack passiert.
	skript := `farbwort := { "Wort " + text }@(text: card.farbe)`

	werte, befunde := service.Auszaehlwerte(t.Context(),
		auszaehlkarten("rot"), auszaehlprobe(skript), nil,
		service.Auszaehlangabe{Name: "x", Skript: "farbwort()"})

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	// **„Wort " und nicht „Wort rot".** Die Bindung wird beim **Anlegen**
	// ausgewertet, und da steht die leere Gestalt — bei MSE genauso: sein
	// `init script:` läuft einmal je Set, mit der Karte, die gerade
	// dasteht. Geprüft wird deshalb, dass der Name **existiert**, nicht dass
	// er den Kartenwert trägt.
	if len(werte) != 1 || werte[0] != "Wort " {
		t.Errorf("die Werte lauten %+v", werte)
	}
}

// **Ein `script:` am Kartenfeld rechnet, bevor die Dimension liest.**
//
// # Warum das hier steht
//
// MSE rechnet die Wertskripte, bevor irgendjemand `card.<feld>` liest —
// `SetScriptManager::updateAll`. `magic.mse-game` schreibt an `card color`
// `script: card_color_filter(value)`, und die Auszählung nach der Farbgruppe
// liest genau dieses Feld. Eine Umgebung, die den rohen Wert stehen ließe,
// zählte etwas anderes als MSE2.
//
// **Und jede Rechnung setzt `card` neu**: das nächste Feldskript sieht, was
// das vorige ergeben hat.
func TestEinWertskriptRechnetVorDerDimension(t *testing.T) {
	t.Parallel()

	vorlage := service.Vorlage{Kartenfelder: []domain.Field{
		{Name: "farbe", Kind: domain.KindText, Skript: `"gerechnet: " + value`},

		// **Das zweite liest das erste** — daran hängt, dass `card` zwischen
		// den Rechnungen neu gesetzt wird.
		{Name: "abgeleitet", Kind: domain.KindText, Skript: `card.farbe + "!"`},
	}}

	werte, befunde := service.Auszaehlwerte(t.Context(),
		auszaehlkarten("rot"), vorlage, nil,
		service.Auszaehlangabe{Name: "x", Skript: "card.abgeleitet"})

	if len(befunde) != 0 {
		t.Fatalf("Befunde: %+v", befunde)
	}

	if len(werte) != 1 || werte[0] != "gerechnet: rot!" {
		t.Errorf("die Werte lauten %+v", werte)
	}
}
