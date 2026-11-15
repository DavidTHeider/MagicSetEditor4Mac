package service

import (
	"context"
	"unicode"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// feldebene entscheidet, **was** ein Feld zeichnet — und ob überhaupt.
//
// Die Reihenfolge der Fragen ist die Aussage: verborgen schlägt alles, ein
// eigenes Bild schlägt den Kartenwert, und erst danach zählt, ob die Karte für
// dieses Feld etwas hinterlegt hat.
// **Der Kasten kommt mit zurück, auch wenn nichts gezeichnet wird.** Er ist
// die Antwort auf eine andere Frage als „was male ich" — nämlich „wo liegt
// dieses Feld", und die stellt die Oberfläche ([render.Feldflaeche]). `nil`
// heißt: dieses Feld liegt nicht auf der Karte.
func feldebene(
	ctx context.Context, feld *msepaket.Feldstil, gerechnet []msescript.Eintrag,
	vorlage Vorlage, breite, hoehe float64, umgebung *msescript.Umgebung,
	symbole Symbolquelle, bilder render.Bildquelle,
) ([]render.Ebene, *render.Rechteck, []domain.Befund) {
	paket := vorlage.Stil.Paket

	// **`render style: hidden` ist keine Lücke, sondern eine Ansage.** Viermal
	// im Pack, jedes Mal an einem `card color` einer Flip-Karte.
	if feld.Darstellung == "hidden" {
		return nil, nil, nil
	}

	// **Der gerechnete Wert und nicht der gespeicherte.** Das Feld `text` von
	// `magic.mse-game` trägt `save value: false` und entsteht erst aus seinem
	// Wertskript; wer hier die rohe Karte läse, zeichnete ein leeres Feld
	// (#345).
	wert, gefuehrt := gerechneterWert(gerechnet, feld.Feld)

	// **`input` ist der Wert dieses Felds** — und steht in einer **eigenen
	// Ebene**, nicht in der gemeinsamen.
	//
	// Herkunft: MSE-Skripte kennen keine Stellungsargumente, sie lesen die
	// Kontextvariable `input` (`src/script/functions/editor.cpp` von MSE2).
	// Für das Skript eines Feldes ist der Kontext das Feld.
	//
	// Gemessen am 2026-08-16 über `M15-Magic-Pack`: 162 Feldstile scheiterten
	// am unbekannten Namen `input`
	// (#240).
	//
	// **Eine Ebene darüber**, weil sonst das nächste Feld mit dem Wert des
	// vorigen rechnete — und das fiele erst an einer Karte auf, bei der beide
	// verschieden sind. Ein Feld ohne Kartenwert bekommt den leeren Text: das
	// ist, was `nil` in einem Text ergibt, und ein Fehler ließe jedes rein
	// gerechnete Bild fallen.
	//
	// Fassung f9d9356 vom 2021-07-06.
	umgebung = umgebung.Darueber()
	umgebung.Setze("input", msescript.Text(wert))

	// **`visible:` entscheidet, ob überhaupt gezeichnet wird.**
	//
	//	bool ValueViewer::isVisible() const { return getStyle()->visible && …; }
	//
	// `src/render/value/viewer.cpp`, Zeilen 36–42, Fassung f9d9356 vom
	// 2021-07-06 — geprüft **vor** dem Zeichnen (`CardViewer::draw`, Zeile 62)
	// und nicht danach (#270).
	if !sichtbar(ctx, feld.Sichtbar, umgebung) {
		return nil, nil, nil
	}

	// **Der Kasten wird hier gerechnet — einmal, für alles.** Er ist MSEs
	// `bounding_box`, und jede Art von Ebene setzt darauf auf; vorher rechnete
	// ihn jeder Zweig für sich, und die Auskunft an die Oberfläche gäbe es gar
	// nicht, wo kein Zweig läuft.
	//
	// **Seine Befunde gehen nur mit, wo etwas gezeichnet wird.** Ein Befund
	// sagt: „hier steht nicht, was das Paket verlangt" — an einem Feld, das
	// ohnehin nichts malt, ist das keine Aussage über die Karte, sondern über
	// eine Rechnung, deren Ergebnis niemand sieht. Vorher entstand er dort gar
	// nicht, weil der Kasten dort gar nicht gerechnet wurde; ihn jetzt zu
	// melden hieße, den kanonischen Lauf mit Befunden über Unsichtbares zu
	// füllen.
	rahmen, rahmenbefunde := rahmenAus(ctx, feld, breite, hoehe, umgebung)
	flaeche := &rahmen

	// **Ein Farbfeld zeichnet eine Fläche, kein Bild und keinen Text.**
	//
	// MSE wählt den Zeichner nach dem **Typ des Feldes**
	// (`IMPLEMENT_VALUE_VIEWER(Color)`), und deshalb steht diese Frage vor
	// allen anderen: ein Farbfeldstil nennt weder Schrift noch `image:`, und
	// jede Frage danach führte am Feld vorbei (`farbebene.go`).
	if beschreibung, ist := farbfeldZu(vorlage, feld.Feld); ist {
		ebenen, befunde := farbebene(
			ctx, feld, wert, beschreibung, paket, rahmen, umgebung, symbole)

		return ebenen, flaeche, append(rahmenbefunde, befunde...)
	}

	// **Ein Bildfeld braucht keinen Kartenwert.** `indicator` in
	// `magic-m15.mse-style` rechnet sein Bild aus `has_identity()` aus, und die
	// Karte führt kein Feld dieses Namens. Wer hier auf einen Wert bestünde,
	// ließe jedes rein gerechnete Bild fallen — der Wert reist trotzdem mit,
	// weil an einem `default:`-Feld genau er das Bild der Karte ist.
	// **Und ein Bildfeld der Karte braucht keinen Feldstil, der ein Bild
	// nennt.** `default:` steht über die 190 Stile von `M15-Magic-Pack` an
	// jedem Kunstfenster — aber es ist nicht die Bedingung dafür, dass die
	// Karte ihr Bild zeigt, sondern das, was ohne es gilt
	// (#380).
	setbild := istBildfeld(vorlage, feld)

	// **Eine Mehrfachwahl als Reihe ihrer Bilder** (#2297) — vor den Wahlbildern,
	// die sonst eine einzige Wahl suchten.
	if istBildreihe(feld, vorlage) {
		ebenen, befunde := bildreihe(ctx, feld, wert, rahmen, vorlage, umgebung, symbole, bilder)

		return ebenen, flaeche, append(rahmenbefunde, befunde...)
	}

	// **Ein Bildfeld der Karte geht immer hier entlang** — auch leer, denn dann
	// zeigt MSE das Schachbrett ([bildebene]).
	if feld.Bild != "" || feld.Standardbild != "" || len(feld.Wahlbilder) > 0 || setbild {
		ebene, befunde := bildebene(ctx, feld, wert, paket, rahmen, umgebung, symbole,
			bildpassung(ctx, feld, vorlage, umgebung), setbild,
			vorlage.Setpaket, vorlage.Bilderverzeichnis)

		if ebene == nil {
			return nil, flaeche, befunde
		}

		return []render.Ebene{*ebene}, flaeche, append(rahmenbefunde, befunde...)
	}

	// **Leer heißt: nichts zu zeichnen — aber der Kasten bleibt.** Genau hier
	// ist eine frisch angelegte Karte unbedienbar geworden; MSE hält den
	// `ValueViewer` unabhängig vom Wert
	// (#562).
	if !gefuehrt || wert == "" {
		return nil, flaeche, nil
	}

	// **Die Symbolschrift schlägt die Textschrift — aber nur, wenn der ganze
	// Wert aus Symbolen besteht.**
	//
	// `casting cost` und `text` führen in `magic-m15.mse-style` **beide**
	// Blöcke. Bei den Manakosten ist `2R` vollständig in Symbolen darstellbar
	// und meint sie auch; im Regeltext steht Fließtext, in dem nur einzelne
	// Stellen Symbole sind.
	//
	// **Ein Wert, der ganz aus Symbolen besteht, wird ein Symbolzug** — er
	// misst und setzt anders als ein Textzug, und die gewöhnlichen Manakosten
	// gehen dort durch.
	//
	// **Geht es nicht auf, wird gemischt**, und nicht „alles oder nichts": MSE
	// teilt in Läufe — der Grund und die Fundstelle stehen bei [symbollaeufe]
	// (#508).
	if name := symbolschriftname(ctx, feld, umgebung); name != "" && ganzAusSymbolen(feld, wert, name, vorlage) {
		symbole, weitere := symbolebenen(ctx, feld, name, wert, rahmen, vorlage, umgebung)

		return symbole, flaeche, append(rahmenbefunde, weitere...)
	}

	if feld.Schrift.Name == "" {
		// **Ein Bildfeld ohne Bildquelle zeichnet nichts — und meldet
		// nichts.**
		//
		//	if (img.isReady()) { … }
		//	else if (viewer.nativeLook()) { margin = 18; }
		//
		// `draw_choice_viewer`, `src/render/value/choice.cpp`, Zeilen 74–92,
		// Fassung f9d9356 vom 2021-07-06. Ist kein Bild bereit, geschieht
		// nichts — kein Fehler, keine Meldung.
		//
		// **Sechs Feldstile im vollen Pack sind so, und alle heißen `card
		// color`**: sie decken die ganze Karte bei `z index: 0`, der Rahmen
		// kommt aus einem eigenen `image`-Feld, und `card color` ist ein
		// Platzhalter. Dieselbe Lage wie bei `render style: hidden` und einer
		// Breite von null — die Vorlage sagt „hier nichts"
		// (#506).
		//
		// Über alle 23.168 kanonischen Karten waren das **20** Befunde.
		if feld.Darstellung == "image" {
			return nil, flaeche, nil
		}

		return nil, flaeche, []domain.Befund{{Ort: feld.Feld, Meldung: ungedeutet(feld)}}
	}

	// **Ein gezeichneter Wahlwert beginnt groß.**
	//
	//	String text = tr(viewer.getStylePackage(), value, capitalize_sentence);
	//
	// `draw_choice_viewer`, `src/render/value/choice.cpp`, Zeile 99, und
	// wortgleich `MultipleChoiceValueViewer::draw`,
	// `src/render/value/multiple_choice.cpp`, Zeile 74. Beides Fassung f9d9356
	// vom 2021-07-06.
	//
	// **Das Sprachpaket wird dabei nicht gefragt, und das ist gemessen**
	// (#561):
	// alle dreizehn `.mse-locale` des Packs führen unter `package:` genau zwei
	// Blöcke — `magic.mse-game` und `magic-mana-*.mse-symbol-font`. Der Aufruf
	// oben schlägt unter dem **Stilpaket** nach, und darauf passt keiner der
	// beiden, auch nicht über MSEs Platzhalterregel. Es gibt im Pack keine
	// Übersetzung, die dieser Weg fände; was bleibt, ist der Rückfall.
	if istAuswahlfeld(vorlage, feld.Feld) {
		wert = satzGross(wert)
	}

	ebene, befunde := textebene(ctx, feld, wert, rahmen, umgebung, vorlage)
	befunde = append(rahmenbefunde, befunde...)

	maske, weitere := bildangabe(ctx, feld.Feld, "mask", feld.Maske, umgebung)
	befunde = append(befunde, weitere...)
	ebene.Maske = alsZeichenbaum(maske, paket, symbole)
	ebene.Satz.Zeilenmaske = zeilenmaskeAus(ebene.Maske)

	return []render.Ebene{ebene}, flaeche, befunde
}

// textebene baut die Ebene eines Textfelds.
func textebene(
	ctx context.Context, feld *msepaket.Feldstil, wert string,
	rahmen render.Rechteck, umgebung *msescript.Umgebung, vorlage Vorlage,
) (render.Ebene, []domain.Befund) {
	groesse, befunde := masszahl(ctx, feld.Feld, "font size", feld.Schrift.Groesse,
		Schriftgroessenvorgabe, umgebung)

	// **Die Schriftfarbe wird gerechnet, nicht gelesen** — 2.321 der 3.011
	// `font: color:`-Angaben des Packs sind Skripte
	// (#382).
	farbe, farbbefunde := farbangabe(ctx, feld.Feld, "font color", feld.Schrift.Farbe, umgebung)
	befunde = append(befunde, farbbefunde...)

	// **Und der Name wird gerechnet** — 664 der 3.119 `font: name:`-Angaben des
	// Packs sind Skripte, dazu 439 der 581 `italic name:`
	// (#393).
	schriftname, namensbefunde := textangabe(ctx, feld.Feld, "font name",
		feld.Schrift.Name, umgebung)
	befunde = append(befunde, namensbefunde...)

	kursivname, kursivbefunde := textangabe(ctx, feld.Feld, "italic name",
		feld.Schrift.KursivName, umgebung)
	befunde = append(befunde, kursivbefunde...)

	gewicht, gewichtsbefunde := textangabe(ctx, feld.Feld, "font weight", feld.Schrift.Gewicht, umgebung)
	lage, lagebefunde := textangabe(ctx, feld.Feld, "font style", feld.Schrift.Lage, umgebung)
	befunde = append(append(befunde, gewichtsbefunde...), lagebefunde...)

	schatten, schattenbefunde := schriftschatten(ctx, feld, umgebung)
	befunde = append(befunde, schattenbefunde...)

	roh, auszeichnungen := ohneAuszeichnungen(feld.Feld, wert)
	befunde = append(befunde, auszeichnungen...)

	// **`always symbol:` mischt Symbole und Text** — der Grund steht bei
	// [symbollaeufe]. Kommt der Wert hier an, ist er **nicht** ganz aus
	// Symbolen; sonst wäre er längst ein Symbolzug.
	if feld.ImmerSymbole {
		if schrift, da := vorlage.Symbolschriften[symbolschriftname(ctx, feld, umgebung)]; da {
			roh = aufgeteilt(roh, schrift)
		}
	}

	abschnitte, symbolbefunde := mitSymbolen(ctx, feld, rahmen, roh, umgebung, vorlage)
	befunde = append(befunde, symbolbefunde...)
	abschnitte = mitSchriftschaltern(abschnitte, gewicht == "bold", lage == "italic")

	// **Die Ausrichtung darf ein Skript sein** — und `shrink-overflow` steckt
	// dann in seinem **Ergebnis**, nicht in seinem Quelltext.
	ausrichtung := ausrichtungstext(ctx, feld, umgebung)

	return render.Ebene{
		Art:        render.ArtText,
		Rahmen:     rahmen,
		Farbe:      farbe,
		Abschnitte: abschnitte,
		Satz: render.Satz{
			Schrift:         schriftname,
			Kursivschrift:   kursivname,
			Groesse:         groesse,
			Ausrichtung:     ausrichtungAus(ausrichtung),
			Stauchen:        stauchtAus(ausrichtung),
			Einzeilig:       einzeiligesFeld(vorlage, feld.Feld),
			NurBeiUeberlauf: nurBeiUeberlauf(ausrichtung),
			Polster:         polsterAus(ctx, feld, umgebung),
			Zeilenfaktoren:  zeilenfaktorenAus(ctx, feld, umgebung),
			Absatzhoehe:     absatzhoeheAus(ctx, feld, umgebung),
			Kleinstgroesse:  kleinstgroesseAus(ctx, feld, umgebung),
			Schatten:        schatten,
		},
		Winkel:  winkelAus(ctx, feld, umgebung),
		Deckung: 1,
	}, befunde
}

// rahmenAus rechnet die Fläche eines Felds aus.
//
// **Gemeinsam für Text und Bild.** Beide stehen an derselben Stelle und messen
// dasselbe; zwei Rechnungen liefen an dem Tag auseinander, an dem eine von
// ihnen `right:` lernte und die andere nicht.
func rahmenAus(
	ctx context.Context, feld *msepaket.Feldstil,
	kartenbreite, kartenhoehe float64, umgebung *msescript.Umgebung,
) (render.Rechteck, []domain.Befund) {
	// Zwei Achsen, jede höchstens zwei Befunde.
	befunde := make([]domain.Befund, 0, 4)

	links, breite, befund := achse(ctx, feld.Feld+" → waagerecht",
		feld.Links, feld.Rechts, feld.Breite, kartenbreite, umgebung)
	befunde = append(befunde, befund...)

	oben, hoehe, befund := achse(ctx, feld.Feld+" → senkrecht",
		feld.Oben, feld.Unten, feld.Hoehe, kartenhoehe, umgebung)
	befunde = append(befunde, befund...)

	return render.Rechteck{Links: links, Oben: oben, Breite: breite, Hoehe: hoehe}, befunde
}

// satzGross macht den ersten Buchstaben groß — und sonst nichts.
//
//	void capitalize_sentence_in_place(String& s) {
//	  if (!s.empty()) s[0] = toUpper(s[0]);
//	}
//
// `src/util/string.cpp`, Zeilen 115–119, Fassung f9d9356 vom 2021-07-06.
//
// **Nicht [strings.Title] und nicht `capitalize`.** MSE hat beides und meint
// hier das eine: `capitalize` schreibt **jedes** Wort groß und lässt kurze
// Wörter aus (`is`, `the`, `of`); das gilt für Menüeinträge einer
// Symbolschrift, nicht für einen Wahlwert.
func satzGross(wert string) string {
	if wert == "" {
		return ""
	}

	zeichen := []rune(wert)

	return string(unicode.ToUpper(zeichen[0])) + string(zeichen[1:])
}
