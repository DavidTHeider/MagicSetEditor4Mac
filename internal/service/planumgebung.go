package service

// Die **Skriptumgebung** eines Stilplans steht hier — drei Schichten und ihre
// Reihenfolge.
//
// Eigene Datei, weil `feldebene.go` mit ihr über 400 Zeilen ginge (§2). Der
// Schnitt folgt der Frage: dort steht, **was** eine Ebene wird, hier, **womit**
// gerechnet wird.

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// umgebungAus baut die Skriptumgebung eines **einzelnen** Stilplans.
//
// **Für viele Karten gibt es [umgebungsbauer]** — er lässt die Initskripte
// einmal laufen statt je Karte. Diese Funktion ist der Sonderfall „eine
// Karte" und baut sich ihren Bauer selbst; damit gibt es einen Weg und nicht
// zwei (#1304).
//
// # Drei Schichten, und die Reihenfolge ist die Aussage
//
//  1. Die **Standardbibliothek** — `max`, `to_int` und ihresgleichen; ein
//     Feldstil rechnet mit ihnen (`max(30, …) + 5`).
//  2. Das **`init script:` des Stils** — `template`, `mana_sort`,
//     `module_card`. Ohne diese Schicht ließen sich 354 von 356
//     `image:`-Angaben in `M15-Magic-Pack` nicht ausrechnen
//     (#186).
//  3. Die **Karte** als `card`.
//
// Die Karte kommt zuletzt, damit ein Initskript sie nicht überschreiben kann —
// eine Vorlage, die `card` neu belegte, machte jedes Feld unlesbar.
//
// **Feldnamen mit Leerzeichen werden zu Unterstrichen.** So steht es in den
// Stilen: `card.card_symbol` greift auf das Feld `card symbol` zu
// (`docs/FORMATE.md`).
func umgebungAus(
	ctx context.Context, karte *domain.CardRow, vorlage Vorlage,
	schriften render.Schriftquelle, bilder render.Bildquelle,
) (*msescript.Umgebung, []msescript.Eintrag, []domain.Befund) {
	return neuerUmgebungsbauer(ctx, vorlage, schriften, bilder, karte).fuerKarte(ctx, karte)
}

func kartenumgebungVorbereiten(
	ctx context.Context, umgebung *msescript.Umgebung, karte *domain.CardRow,
	vorlage Vorlage, schriften render.Schriftquelle, bilder render.Bildquelle,
) []msescript.Eintrag {
	felder := make([]msescript.Eintrag, 0, len(karte.Values)+1)
	for _, wert := range karte.Values {
		felder = append(felder, msescript.Eintrag{
			Name: strings.ReplaceAll(wert.Field, " ", "_"),
			Wert: msescript.Text(wert.Text),
		})
	}

	// **`notes` steht an jeder Karte, auch an einer ohne** — die Regel und
	// ihre Fundstelle stehen bei [mitNotizen]. Was die Karte selbst trägt,
	// steht schon in der Schleife darüber.
	felder = mitNotizen(felder, karte)

	// **Die Karte steht vor den Skripten, nicht dahinter**
	// (#240).
	//
	// Das Initskript von `magic.mse-game` liest `set` und `card` schon beim
	// Vorbelegen — Zeile 10580 greift auf `set` zu. Standen sie danach, brach
	// das Skript dort ab, und **alles** danach fehlte: `module_card`,
	// `card_style`, `color_background`, `module_identity`. Genau die vier
	// Namen, an denen jede `image:`-Angabe scheiterte und die Karte weiß
	// blieb.
	//
	// Dieselbe Reihenfolge hat `exportvorlage.go` schon gebraucht. Zweimal
	// dieselbe Falle heißt: es ist keine, sondern eine Regel — **die Umgebung
	// ist vollständig, bevor ein Skript sie liest.**
	// **Und `stylesheet` steht vor den Vorgabewerten**, nicht dahinter.
	//
	// Alle 142 Vorgabewerte von `magic.mse-game/card_fields` beginnen mit
	// `stylesheet; ` — einer Abhängigkeitsangabe. Stand `stylesheet` erst
	// danach in der Umgebung, war der Name beim Ausrechnen unbekannt, das
	// Skript scheiterte und **jeder einzelne** Vorgabewert wurde leer
	// (#298).
	//
	// Es ist dieselbe Regel wie eine Zeile darüber, nur eine Ebene tiefer: die
	// Umgebung ist vollständig, bevor ein Skript sie liest.
	belegeVorlage(ctx, umgebung, vorlage, bilder, karte)

	// **Und die Felder, die die Karte nicht führt** — mit dem Vorgabewert des
	// Spiels (`kartenvorgaben.go`).
	felder = mitVorgaben(ctx, felder, vorlage.Kartenfelder, umgebung)

	// **Und jede Mehrfachwahl in Normalform** — die Reihenfolge der Erklärung
	// und nicht die der Datei (`mehrfachwahl.go`). **Noch ohne Leernamen**:
	// die Vorgabeskripte kommen erst in [nachbereiten] ([ordneMehrfachwahlen]).
	felder = ordneMehrfachwahlen(felder, vorlage.Kartenfelder)

	umgebung.Setze("card", kartenabbildung(felder, vorlage.Kartenfelder, vorlage.Setpaket))

	// **`extra_card` steht schon leer da.**
	//
	// Es wird erst unten gerechnet — die Zusatzfelder lesen `card`. Aber
	// `magic.mse-game` schreibt in zwei Kartenfeldern
	// `extra_card.vorthos_box or else ""`, und `or else` fängt ein fehlendes
	// **Glied**, keinen fehlenden **Namen**. Ohne diese Zeile scheitert das
	// Feld, statt seinen Ersatzwert zu nehmen
	// (#270).
	umgebung.Setze("extra_card", msescript.Wert{Art: msescript.ArtAbbildung})

	// **`card_style` steht schon vor den Initskripten** — sie lesen es. Gemessen
	// wird dabei, was die Karte führt; nach den Wertskripten wird es unten neu
	// gesetzt. Das ist der zusätzliche Durchgang, den MSE fährt: erst
	// vorbereiten, dann die inhaltsabhängigen Stile neu rechnen
	// (`src/render/card/viewer.cpp`, Zeilen 44–46).
	//
	// Fassung f9d9356 vom 2021-07-06.
	umgebung.Setze("card_style",
		feldstile(ctx, vorlage.Kartenfelder, vorlage.Stil.Felder,
			inhaltsmasse(ctx, felder, vorlage, schriften, bilder, umgebung), umgebung))

	return felder
}

func nachbereiten(
	ctx context.Context, umgebung *msescript.Umgebung,
	felder []msescript.Eintrag, vorlage Vorlage,
	schriften render.Schriftquelle, bilder render.Bildquelle,
	vermerk *Schluesselwortverwendung,
) ([]msescript.Eintrag, []domain.Befund) {
	befunde := make([]domain.Befund, 0, 3)

	// **Das Spiel vor dem Stil.** Ein Stil darf eine Funktion des Spiels
	// überschreiben — das ist der Sinn einer Vorlage; umgekehrt machte das
	// Spiel jede Anpassung eines Stils wirkungslos.
	// **Die Wertskripte erst nach den Initskripten.**
	//
	// Das Feld `text` von `magic.mse-game` rechnet
	// `if mainframe_walker() then … else combined_editor(…)`, und
	// `mainframe_walker` legt das Initskript des Spiels an. Stünde das
	// Wertskript davor, scheiterte es an einem Namen, den es gleich danach
	// gäbe (#345).
	// **Die Vorgabeskripte vor den Wertskripten** — dieselbe Reihenfolge wie
	// bei MSE: `default_script.invokeOnDefault(…) | script.invokeOn(…)`.
	felder, vorgabebefunde := mitVorgabeskripten(ctx, felder, vorlage.Kartenfelder, umgebung)
	befunde = append(befunde, vorgabebefunde...)
	felder = normalisiereMehrfachwahlen(felder, vorlage.Kartenfelder)

	umgebung.Setze("card", kartenabbildung(felder, vorlage.Kartenfelder, vorlage.Setpaket))

	felder, wertbefunde := mitWertskripten(ctx, felder, vorlage.Kartenfelder,
		umgebung, vermerk, func(stand []msescript.Eintrag) {
			umgebung.Setze("card",
				kartenabbildung(stand, vorlage.Kartenfelder, vorlage.Setpaket))
		})
	befunde = append(befunde, wertbefunde...)
	felder = normalisiereMehrfachwahlen(felder, vorlage.Kartenfelder)
	umgebung.Setze("card", kartenabbildung(felder, vorlage.Kartenfelder, vorlage.Setpaket))

	// **Und `card_style` ein zweites Mal**: der Inhalt hat sich geändert, und
	// `content_width` misst ihn. Ein Durchgang und kein Festpunkt — der Grund
	// steht in `docs/FORMATE.md` unter „Der zweistufige Satz".
	umgebung.Setze("card_style",
		feldstile(ctx, vorlage.Kartenfelder, vorlage.Stil.Felder,
			inhaltsmasse(ctx, felder, vorlage, schriften, bilder, umgebung), umgebung))

	// **Zuletzt die Zusatzfelder des Stils** — sie lesen `card`, und `card`
	// liest sie nicht: MSE rechnet sie nach den Kartenfeldern
	// (`SetScriptManager::updateAll`, `src/script/script_manager.cpp`).
	//
	// Fassung f9d9356 vom 2021-07-06.
	gespeichert, lesebefunde := gespeicherteZusatzwerte(felder, vorlage.Stil.Paket, vorlage.Stil.Zusatzfelder)
	befunde = append(befunde, lesebefunde...)

	zusatz, zusatzbefunde := zusatzfelder(ctx, vorlage.Stil.Zusatzfelder, gespeichert, umgebung)
	befunde = append(befunde, zusatzbefunde...)

	// **Und eine Mehrfachwahl unter ihnen in Normalform** — nach dem Skript,
	// wie `MultipleChoiceValue::update` es tut
	// (`src/data/field/multiple_choice.cpp`, Zeilen 55–61, Fassung f9d9356
	// vom 2021-07-06).
	zusatz = normalisiereMehrfachwahlen(zusatz, vorlage.Stil.Zusatzfelder)

	umgebung.Setze("extra_card", msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: zusatz})
	umgebung.Setze("extra_card_style",
		feldstile(ctx, vorlage.Stil.Zusatzfelder, vorlage.Stil.Zusatzstile,
			inhaltsmasseVon(ctx, zusatz, vorlage.Stil.Zusatzstile, vorlage, schriften, bilder, umgebung),
			umgebung))

	return felder, befunde
}
