package service

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Inhaltsmass ist, wie breit und hoch der **gesetzte Inhalt** eines Feldes ist.
//
// # Der zweistufige Satz
//
// In MSE wird erst gesetzt, dann angeordnet. `card_style.<feld>.content_width`
// fragt nach der Breite des Inhalts, und die Feldangaben rechnen damit —
// `magic-m15` schreibt für die Manakosten
//
//	width: { max(30, card_style.casting_cost.content_width) + 5 }
//
// Die Breite des Feldes hängt also an seinem Inhalt. Wer dort die deklarierte
// Breite einsetzt, bekommt bei einem Skript wie oben null und damit für
// **jede** Karte dasselbe Maß.
//
// **Gemessen über die 190 Stile von `M15-Magic-Pack`:** `content_width` steht
// 500-mal da, davon 194-mal für `casting_cost`. Das ist der Grund, warum Phase
// B den Befundzähler kaum bewegt und das Goldbild sehr wohl
// (#291).
type Inhaltsmass struct {
	Breite float64
	Hoehe  float64

	// Zeilen ist die Zahl der Zeilen **nach** dem Umbruch — `content_lines`.
	//
	//	REFLECT_N("content_lines", layout->lines.size());
	//
	// `src/data/field/text.cpp`, Zeile 120, Fassung f9d9356 vom 2021-07-06.
	//
	// **221-mal in den 190 Stilen von `M15-Magic-Pack`**, 174-mal für das Feld
	// `text`: `magic-m15` entscheidet damit, ob ein kurzer Regeltext mittig
	// steht. Null heißt „nicht gemessen".
	Zeilen int

	// Satz ist das gesetzte Layout — `card_style.<feld>.layout`.
	//
	// **Das ganze Objekt und nicht drei Zahlen.** MSE gibt der Vorlage
	// `lines`, `paragraphs`, `blocks` und `separators` weiter; die
	// Planeswalker-Stile von `M15-Magic-Pack` setzen damit jede
	// Fähigkeitszeile
	// (#338).
	Satz render.Satzergebnis
}

// inhaltsmasse misst, was sich **ohne Schriftmaße** messen lässt.
//
// # Warum nur Symbolfelder
//
// Ein Symbolzug besteht aus quadratischen Symbolen bekannter Größe; seine
// Breite ist eine Multiplikation. Ein Textzug bräuchte die Schrift, ihre Größe
// und den Umbruch — und `internal/service` kennt nach §2 keine Schriften. Das
// ist die nächste Unteraufgabe und eine eigene Naht, keine Vergesslichkeit
// hier.
//
// # Gemessen wird, was gezeichnet wird
//
// **Die Felder mit ihren Vorgabewerten, und ohne Auszeichnungen.** Der erste
// Anlauf las die rohe Karte: ein Feld, das erst über seinen `default:` einen
// Wert bekommt, blieb ungemessen, und ein Text mit `<i-auto>` wurde **mit** den
// spitzen Klammern gemessen und **ohne** sie gezeichnet.
//
// Angestrichen vom Zweitblick über die Phasen B bis D — Klasse: Messung und
// Darstellung arbeiten auf verschiedenen Stufen derselben Datenaufbereitung.
//
// # Warum das nicht im Kreis läuft
//
// Gemessen wird gegen die **deklarierte** Symbolgröße, nicht gegen die
// gerechnete Feldhöhe. `magic-m15` schreibt `symbol font: size: 15` hin; wo das
// fehlt, wird die deklarierte Höhe genommen. Beides steht als Zahl in der
// Datei und hängt an keinem Skript — sonst hinge Stufe eins an Stufe zwei, und
// der zweistufige Satz wäre einer mit einem Knoten.
func inhaltsmasse(
	ctx context.Context, felder []msescript.Eintrag, vorlage Vorlage,
	schriften render.Schriftquelle, bilder render.Bildquelle,
	umgebung *msescript.Umgebung,
) map[string]Inhaltsmass {
	return inhaltsmasseVon(ctx, felder, vorlage.Stil.Felder, vorlage, schriften, bilder, umgebung)
}

// inhaltsmasseVon misst die Felder einer Stilliste — der Kartenfelder oder
// der Zusatzfelder.
//
// **Beide über denselben Weg** (Befund 7 des Zweitblicks über E/F): MSE2
// setzt `card_style` und `extra_card_style` mit demselben `TextViewer`, und
// hier bekamen die Zusatzstile die Kastenmaße.
func inhaltsmasseVon(
	ctx context.Context, felder []msescript.Eintrag, stile []msepaket.Feldstil, vorlage Vorlage,
	schriften render.Schriftquelle, bilder render.Bildquelle,
	umgebung *msescript.Umgebung,
) map[string]Inhaltsmass {
	masse := make(map[string]Inhaltsmass, len(stile))

	// **Ohne Satz: alles null** ([ohneSatz], #2466) — wie in MSE2 an einem
	// Stil, der nie gezeichnet wurde.
	if istOhneSatz(ctx) {
		for stelle := range stile {
			masse[stile[stelle].Feld] = Inhaltsmass{}
		}

		return masse
	}

	for stelle := range stile {
		// **Auch hier wirkt der Abbruch.** Ein Stil von `M15-Magic-Pack` führt
		// bis zu 60 Feldstile, und jede Textmessung sucht eine Schrift.
		if ctx.Err() != nil {
			return masse
		}

		feld := &stile[stelle]

		roh, gefuehrt := gemessenerWert(felder, feld.Feld)

		// **Eine Wahl misst ihr Bild, nie ihren Text** ([wahlmass]) — gefüllt
		// wie leer.
		if istWahlart(feldart(vorlage, feld.Feld)) {
			masse[feld.Feld] = wahlmass(ctx, feld, klartextVon(roh), vorlage, bilder, umgebung)

			continue
		}

		if !gefuehrt || klartextVon(roh) == "" {
			if mass, da := leermass(ctx, feld, vorlage, schriften, bilder, umgebung); da {
				masse[feld.Feld] = mass
			}

			continue
		}

		wert := klartextVon(roh)

		// **Gemessen wird, was gezeichnet wird.** Ein Feld mit Symbolschrift ist
		// nur dann ein Symbolzug, wenn der Stil `always symbol:` sagt — sonst
		// ist es Text, in dem einzelne Stellen Symbole sind (#293). Ohne diese
		// Bedingung misst die erste Stufe den Regeltext von `magic-m15` als
		// Reihe von Bildern: seine Symbolschrift führt einen Auffangcode, der
		// **jedes** Zeichen deckt.
		name := strings.TrimSpace(feld.Symbolschrift.Name)
		if name == "" || !feld.ImmerSymbole {
			if mass, da := textmass(ctx, feld, roh, vorlage, schriften, bilder, umgebung); da {
				masse[feld.Feld] = mass
			}

			continue
		}

		schrift, da := vorlage.Symbolschriften[name]
		if !da {
			continue
		}

		if mass, da := symbolmass(ctx, feld, wert, name, schrift, bilder, umgebung); da {
			masse[feld.Feld] = mass
		}
	}

	return masse
}

// symbolmass misst einen Symbolzug gegen seine Bilder.
//
// **Gemessen wird, was gezeichnet wird.** Ein Symbol ist nicht quadratisch: es
// hat die Größe seines Bildes, umgerechnet über `image font size`, plus den
// Abstand der Schrift (Vermerk 0022). Ohne Bildquelle bleibt es beim Quadrat —
// dann ist [render.Symbolmasse] darauf zurückgeworfen, und das steht dort.
func symbolmass(
	ctx context.Context, feld *msepaket.Feldstil, wert, name string,
	schrift msepaket.Symbolschrift, bilder render.Bildquelle,
	umgebung *msescript.Umgebung,
) (Inhaltsmass, bool) {
	treffer, unbekannt := zerlegeInSymbole(wert, schrift)
	if unbekannt != "" || len(treffer) == 0 {
		return Inhaltsmass{}, false
	}

	seite := symbolseiteVorSatz(ctx, feld, umgebung)
	if seite <= 0 {
		return Inhaltsmass{}, false
	}

	breite, hoehe := render.Symbolzugmass(
		symbolbilder(ctx, treffer, name, umgebung),
		render.Symbolmasse{
			Groesse:           seite,
			Bildgroesse:       schrift.Bildgroesse,
			AbstandWaagerecht: schrift.AbstandWaagerecht,
			AbstandSenkrecht:  schrift.AbstandSenkrecht,
		},
		bilder)

	return Inhaltsmass{Breite: breite, Hoehe: hoehe}, true
}

// textmass misst einen Textzug gegen die Schrift, die der Stil nennt.
//
// # Umgebrochen wird, wenn eine Breite dasteht
//
// `content_width` und `content_lines` sind bei MSE Rückmeldungen aus dem
// **gesetzten** Text: erst wird umgebrochen, dann steht fest, wie breit und wie
// viele Zeilen es sind (`src/render/text/viewer.cpp`, Zeilen 350–385, Fassung
// f9d9356 vom 2021-07-06). Genau so wird hier gemessen — sofern die Breite des
// Feldes als **Zahl** dasteht.
//
// **Ein Skript wird ausgerechnet, nicht übergangen** ([umbruchbreite]). Ergibt
// es keine Zahl — im ersten Durchgang etwa, weil es `card_style` liest —, wird
// ohne Grenze gesetzt, und der Zug bleibt **eine** Zeile.
//
// # Woher Schriftname und Größe kommen
//
// Aus [schriftVorSatz], und damit ausgerechnet. Bis
// #420 standen
// hier der rohe Name und die rohe Größe, mit der Begründung, ein Skript hinge
// an Stufe zwei. Es hängt nicht: MSE wertet beide in `updateStyles` aus, lange
// bevor `content_width` zurückgeschrieben wird. Der Preis der alten Antwort
// waren **664 von 3.119** Schriftnamen und **2.528 von 3.119** Größen des
// Packs, für die schweigend gar nicht gemessen wurde — gezählt am Feldstil
// über die 190 Stile von `M15-Magic-Pack` (§10).
//
// # In welchem Schnitt
//
// **In dem, in dem gezeichnet wird**: fett, wo `weight: bold` oder `<b>` es
// sagen, kursiv in der Schrift aus `italic name:` oder dem Schnitt „Italic"
// der Familie ([messschnitte], [render.Kursivschnitte]). Bis F02
// (#2244)
// stand hier „aufrecht gemessen".
//
// **Und mit den Symbolen, die gezeichnet werden** ([messabschnitte]): hier
// stand bis zum 2026-10-11, `<sym>` zähle als gewöhnliche Zeichen — solange
// kursiv aufrecht gemessen wurde, glich sich das zufällig aus.
//
// # Was diese Messung nicht kann
//
// **`shrink-overflow`**: MSE verkleinert einen zu langen Zug; was
// `content_width` dann meldet, ist nicht erschlossen. Es steht hier, damit
// niemand die Zahl für genauer hält, als sie ist.
//
// **Das Polster stand hier auch**, als „noch nicht gelesen" — es wird seit
// #451 mit
// denselben Angaben gesetzt wie gezeichnet ([polsterAus] unten).
func textmass(
	ctx context.Context, feld *msepaket.Feldstil, roh []rohabschnitt, vorlage Vorlage,
	schriften render.Schriftquelle, bilder render.Bildquelle, umgebung *msescript.Umgebung,
) (Inhaltsmass, bool) {
	gesetzt, da := textsatz(ctx, feld, roh, vorlage, schriften, bilder, umgebung)
	if !da {
		return Inhaltsmass{}, false
	}

	breite := 0.0
	for _, zeile := range gesetzt.Zeilen {
		breite = max(breite, zeile.Breite)
	}

	if breite <= 0 {
		return Inhaltsmass{}, false
	}

	return Inhaltsmass{
		Breite: breite, Hoehe: gesetzt.Hoehe, Zeilen: len(gesetzt.Zeilen),
		Satz: gesetzt,
	}, true
}

// textsatz setzt einen Textzug mit denselben Angaben, mit denen er gezeichnet
// wird — für [textmass] und für den leeren Satz in [leermass].
func textsatz(
	ctx context.Context, feld *msepaket.Feldstil, roh []rohabschnitt, vorlage Vorlage,
	schriften render.Schriftquelle, bilder render.Bildquelle, umgebung *msescript.Umgebung,
) (render.Satzergebnis, bool) {
	if schriften == nil {
		return render.Satzergebnis{}, false
	}

	name, groesse := schriftVorSatz(ctx, feld, umgebung)
	if name == "" || groesse <= 0 {
		return render.Satzergebnis{}, false
	}

	schrift, err := schriften.Suche(name, false)
	if err != nil {
		return render.Satzergebnis{}, false
	}

	// **Gesetzt wird über Abschnitte, nicht über eine Zeichenkette.** Ein
	// `<line>` schneidet einen Block, und `card_style.<feld>.layout.blocks`
	// entsteht in dieser Stufe — eine flache Zeichenkette hätte die Grenze
	// schon verloren
	// (#338).
	// **Gesetzt wird mit denselben Angaben wie gezeichnet.** Bei MSE entstehen
	// `content_*` und `layout` aus der **tatsächlichen** Textbereitung:
	// `TextViewer::prepareLines` setzt mit dem echten Stil, und
	// `extractLayoutInfo` liest die Blöcke aus denselben Zeilen ab, die
	// gezeichnet werden (`src/render/text/viewer.cpp`, Zeilen 348–389, Fassung
	// f9d9356 vom 2021-07-06). Einen zweiten, ärmeren Satz gibt es dort nicht.
	//
	// **Hier stand ein Kasten der Höhe „Schriftgröße"**, ohne Zeilenfaktoren
	// und ohne Polster. Ohne Kastenhöhe kann [render.SetzeAbschnitte] nicht
	// dehnen, ohne Faktoren steht jede Zeile im Abstand 1,0 — und
	// `card_style.<feld>.layout` beschrieb einen Satz, den niemand sah. Auf
	// `Green / A-Tyvar Kell` standen die Treuemarken deshalb zu eng
	// (#451).
	//
	// **Der Ursprung bleibt bei null.** MSE führt die Blöcke relativ zum
	// Textlauf, nicht zur Karte; `abil_center` addiert `card_style.text.top`
	// ausdrücklich dazu.
	_, feldhoehe, _ := gerechneteKanten(ctx, feld, umgebung, feld.Oben, feld.Hoehe, feld.Unten)

	// **Und mit der Ausrichtung**, denn das Layout beschreibt den Satz, der
	// gezeichnet wird:
	//
	//	// store information about the content/layout, allow this to change alignment
	//	if (style.alignment.isScripted()) {
	//	  style.layout = extractLayoutInfo();
	//	  style.alignment.update(ctx);
	//	}
	//	// align
	//	alignLines(dc, chars, style);
	//	…
	//	// make layout available to scripts
	//	style.layout = extractLayoutInfo();
	//
	// `TextViewer::prepareLines`, `src/render/text/viewer.cpp`, Zeilen 405–425,
	// Fassung f9d9356 vom 2021-07-06. **Die zweite Zuweisung ist die, die die
	// Skripte sehen** — nach dem Ausrichten. Die erste gibt es nur, damit ein
	// Ausrichtungs**skript** überhaupt etwas zu lesen hat, und genau dieselbe
	// Reihenfolge steht hier: [ausrichtungstext] wertet aus, was die Umgebung
	// **jetzt** hergibt, und erst danach wird gesetzt.
	//
	// **Was das gekostet hat**: `magic-m15-saga` setzt die Laschen einer
	// Class-Karte auf `card_style.text.layout.separator[i]`. Ohne Ausrichtung
	// stand der Messsatz immer oben; bei `middle` und einem Kasten, der höher
	// ist als der Text, fehlte die halbe Differenz, und die Laschen landeten
	// mitten im Regeltext
	// (#494,
	// #584).
	ausrichtung := ausrichtungstext(ctx, feld, umgebung)

	rahmen := render.Rechteck{
		Breite: umbruchbreite(ctx, feld, umgebung),
		Hoehe:  max(feldhoehe, groesse),
	}

	return render.SetzeAbschnitte(
		render.Satz{
			Schrift: name, Groesse: groesse,
			Polster:         polsterAus(ctx, feld, umgebung),
			Zeilenfaktoren:  zeilenfaktorenAus(ctx, feld, umgebung),
			Absatzhoehe:     absatzhoeheAus(ctx, feld, umgebung),
			Ausrichtung:     ausrichtungAus(ausrichtung),
			Stauchen:        stauchtAus(ausrichtung),
			Einzeilig:       einzeiligesFeld(vorlage, feld.Feld),
			NurBeiUeberlauf: nurBeiUeberlauf(ausrichtung),

			// **Verkleinert wie gezeichnet** (`scale down to:`): MSE misst
			// den Satz, den es zeichnet. Ohne diese Angabe maß die Messung in
			// voller Größe, und bei den Planeswalkern rückten die Treuekästen
			// nach unten (F02, #2244).
			Kleinstgroesse: kleinstgroesseAus(ctx, feld, umgebung),

			// **Und an der Maske umbrochen wie gezeichnet** (#2299).
			Zeilenmaske: messmaske(ctx, feld, umgebung, vorlage),
		},
		rahmen,
		messabschnitte(ctx, feld, roh, rahmen, vorlage, umgebung),
		messschnitte(ctx, feld, schrift, name, schriften, umgebung), bilder), true
}

// ohneSatzSchluessel markiert einen Kontext, in dem nichts gesetzt wird.
type ohneSatzSchluessel struct{}

// ohneSatz gibt einen Kontext, in dem die Inhaltsmaße null sind —
// #2466.
//
// **Für die Auswertung eines Packfilters.** MSE2 setzt eine Karte nur, um sie
// zu zeichnen (`src/render/card/viewer.cpp`, Zeilen 42–50); einen Packfilter
// wertet es im Kontext der Karte aus, ohne Satz (`PackInstance::PackInstance`,
// `src/data/pack.cpp`, Zeilen 110–117), und die Maße eines nie gezeichneten
// Stils sind dort null (`src/data/field/text.cpp`, Zeilen 113–119, Fassung
// f9d9356 vom 2021-07-06). Hier kostete der Satz je Karte zwei Fünftel des
// ersten Aufrufs von „Zufallspack“.
func ohneSatz(ctx context.Context) context.Context {
	return context.WithValue(ctx, ohneSatzSchluessel{}, true)
}

// istOhneSatz sagt, ob ein Kontext [ohneSatz] trägt.
func istOhneSatz(ctx context.Context) bool {
	gesetzt, _ := ctx.Value(ohneSatzSchluessel{}).(bool) //nolint:errcheck // fehlt sie, wird gesetzt

	return gesetzt
}
