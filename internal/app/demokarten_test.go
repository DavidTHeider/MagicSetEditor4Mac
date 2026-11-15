//go:build fremdpakete

package app_test

import (
	"os"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Selbst geschriebene Karten, gegen echte Vorlagen gezeichnet.**
//
// # Warum es diesen Lauf gibt, obwohl es zwei andere schon gibt
//
// `mse2goldbild_test.go` misst gegen **fremde** Bilder, deren Herkunft
// unbekannt ist (#375):
// als Verlaufsmaß brauchbar, als Zusage nicht.
// `service/fremdpakete_test.go` zählt Befunde über 190 Stile — mit **einer**
// Probekarte, die jedes Textfeld auf „Probe" setzt. Beides zusammen lässt eine
// Lücke: **stimmt das Bild einer gewöhnlichen Karte?**
//
// Dieser Lauf schließt sie mit Karten, die wir selbst schreiben — Kreatur,
// Spontanzauber, mehrfarbig, Land, Planeswalker — und mit Zusagen, die **ohne**
// fremde Referenz auskommen: was der Plan führt, muss auf dem Bild ankommen.
//
// **Er prüft nicht, ob es aussieht wie MSE.** Das kann er nicht, und der Satz
// steht hier, damit niemand ihn dafür hält (§10).

// demokarte ist eine Karte samt der Zusagen, die für sie gelten.
type demokarte struct {
	name   string
	stil   string
	felder map[string]string

	// mindestens ist die Zahl der Ebenen je Art, die der Plan führen muss.
	// **Von unten und nicht genau**: eine Vorlage darf mehr zeichnen; wenn eine
	// ganze Sorte ausfällt, ist es ein Bruch.
	//
	// **Die Textzahlen stiegen mit dem Set** (#384): Setcode, Illustrator und
	// Copyright sind drei Zeilen, die jede Karte führt.
	bilder, texte int

	// doppelt sind Texte, die auf dieser Karte **berechtigt** zweimal
	// stehen — aus zwei verschiedenen Feldern.
	//
	// Eine Teilungskarte hat zwei Hälften, und beide sagen „Instant —".
	// [pruefeKeinDoppelterText] fängt sonst genau das, wofür es da ist
	// (#482).
	doppelt []string

	// kunst ist der Dateiname der Illustration im Set-Paket — leer heißt: diese
	// Karte führt keine (#380).
	kunst string

	// indikator sagt, dass der Farbpunkt vor der Typzeile stehen muss (#390).
	indikator bool

	// absatzband sagt, dass diese Karte `paragraph height` benutzt (#479).
	absatzband bool

	// mischart nennt die Mischart, die auf dieser Karte vorkommen **muss**
	// (#470).
	//
	// **Ohne eine solche Karte prüft nichts den Weg.** Die Formeln stehen
	// seit `combine_blend` in `render`, aber dass `set_combine` seine Art bis
	// an die Ebene bringt, ist eine Naht — und die trägt erst, wenn ein
	// echtes Paket sie belegt.
	mischart string

	// ohneFelder nennt die Felder, die **diese Vorlage nicht zeichnet**.
	//
	// **Mit Grund, nicht als Ausnahme** (§10): `magic-m15-textless` hat keine
	// Typzeile, und `pruefeTexte` verlangte sie von jeder Karte. Wer hier
	// etwas einträgt, schreibt dazu, warum die Vorlage es nicht führt.
	ohneFelder []string

	// buntMindestens und dunkelHoechstens sind die Schwellen für
	// [pruefeBemalung], in Prozent. Null heißt: die Vorgabe.
	//
	// # Warum sie überhaupt je Karte stehen
	//
	// Die Vorgabe — ein Fünftel farbig, höchstens sieben Zehntel schwarz —
	// stammt von der gewöhnlichen M15-Karte und war als allgemeine Zusage
	// **falsch**. Am 2026-08-18 hat sie drei Vorlagen als „tot" gemeldet, die
	// tadellos zeichnen: eine weiße Planeswalkerkarte ist in der Mitte weiß
	// (R = G = B, also nicht „farbig"), eine textlose Vorlage hat gar keinen
	// Textkasten, und bei einer Vollbildvorlage **ist** die ganze Karte das
	// Kunstfenster — ohne Illustration also schwarz
	// (#477).
	//
	// Was hier steht, ist gemessen und begründet; was es weiter fängt, ist
	// der Rahmen, der **ganz** ausfällt.
	buntMindestens, dunkelHoechstens int
}

// fussdaten sind die Zeile unter der Illustration — auf jeder echten Karte.
//
// **Eigener Inhalt** (§11): ein erfundener Name, ein erfundener Verlag.
//
// **Ohne den Setcode**, und das ist kein Vergessen: er ist kein Kartenwert
// (`editable: false`, `script: set_code_script(…)`), sondern kommt aus dem Set
// — [demoset] setzt ihn (#384).
func fussdaten() map[string]string {
	return map[string]string{
		"illustrator": "Marta Feldhaus",
		"copyright":   "™ & © 2026 Beispielverlag",
	}
}

// mitFuss hängt jeder Demokarte ihren Fuß an.
func mitFuss(karten []demokarte) []demokarte {
	// **Über den Index**: `demokarte` wiegt 160 Byte, und eine Kopie je
	// Durchgang wäre hier zusätzlich irreführend — geschrieben wird in die
	// Abbildung `felder`, und die teilt sich die Kopie mit dem Original.
	for stelle := range karten {
		for feld, wert := range fussdaten() {
			karten[stelle].felder[feld] = wert
		}
	}

	return karten
}

// TestDemokartenZeichnen zeichnet die Demokarten und prüft, was ankommt.
func TestDemokartenZeichnen(t *testing.T) {
	wurzel := fremdpaketwurzel(t)

	for _, karte := range demokarten() {
		t.Run(karte.name, func(t *testing.T) {
			pruefeDemokarte(t, wurzel, karte)
		})
	}
}

func pruefeDemokarte(t *testing.T, wurzel string, karte demokarte) {
	t.Helper()

	stil := fremdstil(t, wurzel, karte.stil)
	suche := schriften.Neue(schriften.Orte()).MitPaketen(strings.TrimSuffix(wurzel, "/data"))
	bilder := bildquelleFuer(wurzel)

	vorlage := fremdvorlage(t, stil, wurzel)

	// **Die Karten liegen in einem Set** (#384). Ohne seine Angaben bleibt
	// `set.*` im ganzen Lauf leer, und der Setcode im Fuß ist ein Leerzeichen —
	// eine Vorrichtung, die ärmer ist als der Betrieb (§3).
	vorlage.Setangaben = demoset()

	if karte.kunst != "" {
		vorlage.Setpaket = kunstsetFuer(t, karte.kunst)
	}

	plan, befunde, err := service.NeuerPlanbauer().
		MitSchriften(suche).MitBildern(bilder).
		Stilplan(probekontext(t), demokartenzeile(karte), vorlage)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	for _, befund := range befunde {
		t.Logf("  Befund %s: %s", befund.Ort, befund.Meldung)
	}

	bild, rasterbefunde, err := render.NeuerRasterer(bilder).MitSchriften(suche).
		Zeichne(probekontext(t), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	// **Ein Rasterbefund ist ein Bruch, kein Hinweis.** Der Plan steht dann
	// schon; wer ihn nicht zeichnen kann, hat ein Loch auf der Karte.
	for _, befund := range rasterbefunde {
		t.Errorf("Rasterbefund an Ebene %d: %s — die Ebene setzt %q aus %#v",
			befund.Ebene, befund.Meldung, plan.Ebenen[befund.Ebene-1].Klartext(),
			plan.Ebenen[befund.Ebene-1].Abschnitte)
	}

	bilder2, texte := ebenenarten(plan)
	t.Logf("%s über %s: %d Ebenen (%d Bild, %d Text), %d Befunde",
		karte.name, karte.stil, len(plan.Ebenen), bilder2, texte, len(befunde))

	if bilder2 < karte.bilder || texte < karte.texte {
		t.Errorf("%d Bildebenen und %d Textebenen, erwartet mindestens %d und %d",
			bilder2, texte, karte.bilder, karte.texte)
	}

	legeBildAb(t, bild, "demo-"+karte.name)

	pruefeBemalung(t, bild, karte)
	pruefeTexte(t, plan, karte)
	pruefeKeinDoppelterText(t, plan, karte)

	// **Die Illustration muss im Bild ankommen, nicht nur im Plan.** Ein
	// Verweis, den die Bildquelle nicht auflöst, stünde im Plan genauso da —
	// und die Karte sähe aus wie eine ohne.
	if karte.kunst != "" && !fuehrtIllustration(bild) {
		t.Errorf("die gemalte Ecke steht nicht auf der Karte — die Illustration "+
			"ist nicht angekommen; Befunde: %v", befunde)
	}

	// **Und der Fuß muss lesbar sein.** Er stand vor
	// #382
	// schwarz auf dem schwarzen Kartenrand: im Plan da, auf dem Bild nicht.
	//
	// **Die Schwelle ist gemessen, nicht geschätzt.** Mit gerechneter
	// Schriftfarbe stehen dort **145** helle Punkte, mit der alten (schwarzen)
	// **37** — der Rest ist nicht null, weil in dieses Band auch die
	// Illustratorpinsel und die abgerundete Ecke hineinreichen. Hundert liegt
	// zwischen beidem und weit von beidem.
	if karte.indikator {
		pruefeIndikator(t, plan)
	}

	if karte.absatzband {
		pruefeAbsatzband(t, plan, suche)
	}

	if karte.mischart != "" {
		pruefeMischart(t, plan, karte.mischart)
	}

	if karte.stil == "magic-m15" {
		pruefeSetcode(t, plan)
		pruefeSeitenrand(t, bild)

		hell := helleImFuss(bild)
		t.Logf("  helle Bildpunkte im Fußband: %d", hell)

		if hell < 100 {
			t.Errorf("nur %d helle Bildpunkte im Fußband — der Fuß der Karte ist "+
				"nicht lesbar; Befunde: %v", hell, befunde)
		}
	}
}

// demokartenzeile macht aus einer Demokarte eine Kartenzeile.
func demokartenzeile(karte demokarte) *domain.CardRow {
	zeile := &domain.CardRow{ID: 1, Number: 1}

	for feld, wert := range karte.felder {
		zeile.Values = append(zeile.Values, domain.Value{Field: feld, Text: wert})
	}

	return zeile
}

// ebenenarten zählt Bild- und Textebenen eines Plans.
func ebenenarten(plan *render.Plan) (bilder, texte int) {
	for i := range plan.Ebenen {
		switch plan.Ebenen[i].Art {
		case render.ArtBild:
			bilder++
		case render.ArtText:
			texte++
		}
	}

	return bilder, texte
}

// fremdpaketwurzel gibt das Paketverzeichnis oder überspringt den Lauf.
func fremdpaketwurzel(t *testing.T) string {
	t.Helper()

	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	return wurzel
}

// bildquelleFuer gibt die Bildquelle über dem Paketverzeichnis.
func bildquelleFuer(wurzel string) render.Bildquelle {
	return msepaket.NeueBildquelle(wurzel)
}

// pruefeKeinDoppelterText hält fest, dass **kein Text zweimal gezeichnet wird**
// (#482).
//
// # Warum es diese Zusage braucht
//
// Die Demokarte „Zeitalter" zeigte ihren Kapiteltext zweimal — einmal im
// Kapitelkasten, einmal in Stufe 1 —, und **kein einziger Lauf sagte etwas**.
// Ebenenzahlen, Bemalung und „steht der Text im Plan" gehen alle durch, wenn
// eine Zeichenkette doppelt dasteht.
//
// Die Ursache war eine Vorrichtung, die großzügiger war als der Betrieb (§3):
// die Karte speicherte einen `rule text`, den MSE bei einer Saga nie schreibt,
// und `magic.mse-game` reicht ihn dann an **zwei** Felder weiter. Die
// Erklärung steht bei der Karte in `demokartenliste_test.go`.
//
// # Was hier geprüft wird, und was nicht
//
// **Nur sichtbare Ebenen.** Eine Breite oder Höhe von null heißt „wird nicht
// gezeichnet" (`ValueViewer::isVisible`, `src/render/value/viewer.cpp`, Zeilen
// 36–42, Fassung f9d9356 vom 2021-07-06), und das Pack benutzt das: der
// Fußtext steht in `magic-m15` zweimal im Plan, einmal mit Breite null.
//
// **Und nur nennenswerte Texte.** Ein Nullbreitenzeichen ist der
// Platzhalter, mit dem `magic.mse-game` leere Stufen füllt; zwei leere Stufen
// sind kein Befund.
//
// **Eine berechtigte Doppelung wird benannt**, nicht wegdefiniert: eine
// Teilungskarte hat zwei Hälften, und beide sagen „Instant —" — aus **zwei**
// Feldern. Sie steht als `doppelt:` an der Karte, mit Grund daneben.
//
// **Was er nicht fängt**: zwei Ebenen, die *fast* denselben Text tragen, und
// eine Ebene, die an der falschen Stelle steht. Dafür ist die Sichtprüfung da.
func pruefeKeinDoppelterText(t *testing.T, plan *render.Plan, karte demokarte) {
	t.Helper()

	erlaubt := make(map[string]bool, len(karte.doppelt))
	for _, text := range karte.doppelt {
		erlaubt[text] = true
	}

	gesehen := make(map[string]int, len(plan.Ebenen))

	for i := range plan.Ebenen {
		ebene := &plan.Ebenen[i]

		if ebene.Art != render.ArtText ||
			ebene.Rahmen.Breite <= 0 || ebene.Rahmen.Hoehe <= 0 {
			continue
		}

		text := strings.TrimSpace(strings.ReplaceAll(ebene.Klartext(), "\u200b", ""))
		if len([]rune(text)) < 8 {
			continue
		}

		if erste, doppelt := gesehen[text]; doppelt && !erlaubt[text] {
			t.Errorf("Ebene %d zeichnet denselben Text wie Ebene %d: %q",
				i+1, erste, text)

			continue
		}

		gesehen[text] = i + 1
	}
}
