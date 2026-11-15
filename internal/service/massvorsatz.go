package service

// **Die Angaben, die die Messstufe braucht — ausgerechnet, nicht roh gelesen**
// (#420).

import (
	"context"
	"math"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// zahlVorSatz wertet eine Maßangabe der ersten Stufe aus.
//
// # Warum die Befunde hier verfallen
//
// **Weil sie an der Zeichenstufe fallen, und dort einmal.** Jede dieser
// Angaben wird unten ein zweites Mal ausgerechnet — dieselbe Quelle, dieselbe
// Umgebung, dieselben Helfer. Würde die Messstufe mitreden, stünde jeder
// Befund **doppelt** in der Liste, und der erste käme aus einem Durchgang, in
// dem `card_style` noch gar nicht steht (siehe [textmass]).
//
// Was das kostet, gehört danebengeschrieben (§10): eine Angabe, die **nur**
// die Messstufe liest, meldete ihren Fehler nirgends. Es gibt sie nicht —
// `font: size:`, `font: name:`, `width:` und `height:` werden alle gezeichnet;
// `symbol font: size:` liest [symbolseiteVorSatz], und der Satz nimmt sie
// später über die Feldhöhe. Wer hier eine Angabe hinzunimmt, prüft das nach.
func zahlVorSatz(
	ctx context.Context, feld, angabe, quelle string, vorwert float64,
	umgebung *msescript.Umgebung,
) float64 {
	if strings.TrimSpace(quelle) == "" {
		return 0
	}

	zahl, _ := masszahl(ctx, feld, angabe, quelle, vorwert, umgebung)

	return zahl
}

// umbruchbreite gibt die Breite, gegen die gemessen wird.
//
// **Ausgerechnet wie beim Zeichnen.** `Style::update`
// (`src/data/field.cpp`, Zeilen 141–151, Fassung f9d9356 vom 2021-07-06)
// wertet `width` aus, bevor gesetzt wird; roh gelesen wird sie bei MSE
// nirgends.
//
// **Kein Knoten im zweistufigen Satz.** Liest die Breite selbst `card_style`,
// so bekommt sie im ersten Durchgang den Stand von vorher und ergibt null —
// dann wird ohne Grenze gesetzt, und der Zug bleibt eine Zeile. Im zweiten
// Durchgang steht der Wert. Genau diesen Weg fährt MSE.
func umbruchbreite(
	ctx context.Context, feld *msepaket.Feldstil, umgebung *msescript.Umgebung,
) float64 {
	if zahl := zahlVorSatz(ctx, feld.Feld, "breite", feld.Breite,
		Laengenvorgabe, umgebung); zahl > 0 {
		return zahl
	}

	// **Nicht null**, denn eine Breite von null bricht nach jedem Wort um.
	return math.MaxFloat64
}

// symbolseiteVorSatz gibt die Kantenlänge eines Symbols **vor** dem Satz.
//
// Erst die `size:` des Symbolschriftblocks, dann die Höhe des Feldes — beide
// ausgerechnet. Ergibt keine der beiden eine Zahl, wird nicht gemessen; eine
// geratene Zahl wäre schlechter als die deklarierte Breite, die dann dasteht.
func symbolseiteVorSatz(
	ctx context.Context, feld *msepaket.Feldstil, umgebung *msescript.Umgebung,
) float64 {
	for _, angabe := range []struct {
		name, quelle string
		vorwert      float64
	}{
		{"symbolgröße", feld.Symbolschrift.Groesse, Symbolgroessenvorgabe},
		{"höhe", feld.Hoehe, Laengenvorgabe},
	} {
		if zahl := zahlVorSatz(ctx, feld.Feld, angabe.name, angabe.quelle,
			angabe.vorwert, umgebung); zahl > 0 {
			return zahl
		}
	}

	return 0
}

// schriftVorSatz gibt Namen und Größe der Schrift, mit der gemessen wird.
//
// **Beide ausgerechnet.** `Font::update` (`src/data/font.cpp`, Zeilen 27–36,
// Fassung f9d9356 vom 2021-07-06) wertet Name und Größe aus, gerufen über
// `TextStyle::update` und `SetScriptManager::updateStyles`
// (`src/script/script_manager.cpp`, Zeilen 208–240) — vor dem Satz und vor dem
// Rückschreiben von `content_width`.
//
// Über die 190 Stile von `M15-Magic-Pack` sind **664 von 3.119** `font: name:`
// und **2.528 von 3.119** `font: size:` Skripte — vier Fünftel aller Größen;
// gezählt am Feldstil, nicht an gezeichneten Ebenen (§10). Für sie hat die
// Messstufe bis
// #420
// schweigend nichts gemessen.
func schriftVorSatz(
	ctx context.Context, feld *msepaket.Feldstil, umgebung *msescript.Umgebung,
) (name string, groesse float64) {
	gerechnet := schriftnameAus(ctx, feld.Feld, "schrift", feld.Schrift.Name, umgebung)

	return gerechnet,
		zahlVorSatz(ctx, feld.Feld, "schriftgröße", feld.Schrift.Groesse,
			Schriftgroessenvorgabe, umgebung)
}

// schriftnameAus rechnet einen Schriftnamen aus — und gibt bei einem Fehler
// den **leeren** Namen, nicht den Rohtext.
//
// **Was sich nicht ausrechnen lässt, wird nicht gesucht.** `textangabe` gibt
// dann den Rohtext zurück, und eine Schrift namens `{if …}` gibt es nicht —
// die Suche liefe bei einem frischen Werk einmal über den ganzen Paketbaum,
// um das festzustellen (F02, gefunden am Profil von E02).
//
// **Nur für die Messung.** Der erste Durchgang von `card_style` läuft vor den
// Initskripten und scheitert dort regelmäßig, und seine Befunde werden
// verworfen. Der **Plan** behält den Rohtext (`textebene`): dort soll die
// Meldung der Schriftsuche sagen, wonach sie sucht
// (`TestEinUnausrechenbarerSchriftnameWirdBenannt`).
func schriftnameAus(
	ctx context.Context, feld, angabe, quelle string, umgebung *msescript.Umgebung,
) string {
	gerechnet, befunde := textangabe(ctx, feld, angabe, quelle, umgebung)
	if len(befunde) > 0 {
		return ""
	}

	return strings.TrimSpace(gerechnet)
}
