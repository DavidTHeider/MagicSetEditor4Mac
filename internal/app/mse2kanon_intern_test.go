//go:build fremdpakete

// Der Abstand zu MSE2 an einer Referenz mit **bekannter Herkunft**
// (#407).
//
// # Was diesen Lauf von `TestDerAbstandZuMSE2` unterscheidet
//
// Dort stehen 27 Bilder aus der Gemeinschaft, deren Vorlage **unbekannt** ist —
// und nachweislich eine andere als die, die unsere Rekonstruktionen nennen
// (Rand 17/18 gegen 14/15, Fuß, Seltenheitssymbol; #375, #376, #378). Der
// Abstand dort ist ein Verlaufsmaß und keine Aussage über Parität.
//
// **Wie gemessen wird, steht in `docs/MESSMETHODIK.md`** — das Maß, die
// Stichprobe je Schicht und der geschichtete Schätzer, jeweils mit ihren
// Gleichungen. Wer eine Zahl aus diesem Lauf zitiert, findet dort, wie sie
// zustande kommt.
//
// Hier stammen **Eingabe und Ausgabe von MSE2**: die kanonischen Sets sind
// MSE2s eigene, und die Bilder daneben hat MSE2 selbst gerastert —
// `--export-images`, gebaut auf diesem Rechner
// (`docs/entscheidungen/0026-mse2-rendert-die-referenz-selbst.md`).
//
// **Woher die Sets kommen, steht seit dem 2026-09-07 im selben Vermerk**:
// `archive.org/details/mse-canon-cards-2022-10-24`, gestellt im MSE-Forum
// (Thread 2128), Daten aus **Scryfall** mit Stand 2022-10-24. Der Ersteller
// schränkt selbst ein — „some inaccuracies and maybe missing cards", und
// silberrandige Karten seien „very off". Das gehört zur Reichweite jeder Zahl,
// die hier herauskommt (§10).
//
// **Was er trotzdem nicht ist**: ein Beweis für Parität. Er ist ihre
// **Messgröße**. Was er misst, ist der Abstand in Bildpunkten — nicht, ob ein
// Mensch die beiden unterscheiden könnte.
package app

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// **Der Abstand zu MSE2 über ein ganzes kanonisches Set.**
func TestDerAbstandZuMSE2AmKanon(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	// **Und mit den Schriften des Packs** (#1170).
	//
	// # Warum das eine eigene Angabe ist
	//
	// `MSE_FREMDPAKETE` zeigt auf das Verzeichnis, das die **Pakete** enthält
	// — bei `M15-Magic-Pack` also auf `…/data`. Die Schriften liegen
	// **daneben**, unter `Magic - Fonts/`, und damit außerhalb. Eine Wurzel
	// daraus abzuleiten wäre eine Annahme über fremde Sammlungen (§11).
	//
	// # Was ohne sie geschah
	//
	// **Jede Karte wurde ohne Text gezeichnet, und dieser Lauf maß das.** Bis
	// zum 2026-08-26 fand die Schriftsuche sie über `schriften.Paketwurzel()`
	// unter `HOME`; seither gibt `TestMain` dem Lauf ein eigenes Zuhause, und
	// dort liegt kein Paket. Der Abstand stieg von 18,853 auf 28,362, die
	// Zerlegung meldete **0 % `Text`** statt 56 % — und niemand sah es, weil
	// der Lauf zugleich absprang.
	//
	// **`MSE_SCHRIFTWURZEL` wird seit dem 2026-08-28 in `gestartetMit`
	// gelesen** und nicht mehr hier: sie stand nur an dieser Stelle, und der
	// Zeichenlauf daneben wusste nichts davon.
	anwendung := gestartetMit(t, func() string { return wurzel })

	var (
		gemessen int
		gesamt   []float64
	)

	// **Wo der Abstand sitzt, nicht nur wie groß er ist** (#521), und ob er
	// aus einem Versatz kommt (#522).
	stand := neuerMessstand()

	// **Vier Sets, so viele wie Referenzbilder dalegen**
	// (#474).
	//
	// `Archenemy` führt 70 Karten im Querformat 435 × 621, die drei anderen
	// zusammen 11.180 im Hochformat — `White` davon **3.666 über
	// `m15-altered`**: die gewöhnliche M15-Karte, an der sich entscheidet, ob
	// dieses Programm eine echte Magic-Karte zeichnet.
	//
	// **Hier standen bis zum 2026-08-18 nur zwei.** `Green` und `Red` lagen
	// gerendert daneben und wurden nicht gemessen — eine Verschlechterung, die
	// nur sie träfe, wäre grün geblieben.
	//
	// **Und bis zum 2026-08-22 waren alle vier einfarbig.** Damit war über
	// **Gold- und Hybridrahmen** nichts gemessen: `color_background` baut sie
	// aus mehreren Farbrahmen und einer Blendmaske, und dieser Weg wurde von
	// keinem einzigen Bild berührt. Aufgefallen ist es beim Ansehen zweier
	// Karten, nicht einem Test
	// (#862).
	//
	// **Und am selben Tag waren es fünf von zehn.** `Black`, `Blue`,
	// `Colorless`, `Planechase` und `Vanguard` lagen die ganze Zeit unter
	// `kanonische-sets/` — MSE2 hatte sie nur nie exportiert, und was nicht
	// gerastert danebenliegt, wird nicht verglichen
	// (#880).
	//
	// **Drei davon sind nicht mehr vom Gleichen**: `Colorless` bringt
	// Artefakte und Länder samt Artefakt-Blendmaske, `Planechase` und
	// `Vanguard` je ein **Überformat** mit eigenem Stil — wie Archenemys
	// Schemata, und die sind mit 14,5 das beste Set. `Black` und `Blue` sind
	// einfarbig wie White, Green und Red.
	//
	// Zusammen sind es jetzt **25.148 Karten** — dieselbe Zahl, die der
	// Befundlauf seit jeher nennt. Er zog von Anfang an aus allen zehn Sets;
	// **verglichen** wurde bis heute gegen vier.
	// **Die Stichprobe wird je Set gezogen, nicht geteilt** — auf Anweisung
	// des Auftraggebers vom 2026-09-07.
	//
	// Hier stand `stichprobengroesse(t, len(saetze))`: der Bedarf für alle
	// 25.148 Karten, verteilt auf zehn Sets — rund 68 je Set. Damit trug die
	// Zusage „99,9 Prozent Vertrauen für 1 Prozent" den **Topf** und nicht das
	// einzelne Set: 68 aus 3.788 Karten treffen einen Fehler in einem Prozent
	// nur etwa jedes zweite Mal.
	//
	// **Jetzt bekommt jedes Set seine eigene Schranke** ([stichprobeFuer]),
	// gerechnet aus **seiner** Kartenzahl. Der Lauf misst dadurch mehr Karten
	// und trägt dafür eine Aussage, die für **jedes** Set gilt.
	saetze := kanonsaetze

	// **Jedes Set ist eine Schicht** — auf Anweisung des Auftraggebers vom
	// 2026-09-07. Der Schätzer und seine Gleichungen stehen in
	// `kanonstratum_intern_test.go`.
	var schichten []schicht

	for _, satz := range saetze {
		abstaende, ganz := messeSatzGegenMSE2(t, anwendung, satz, stand)
		if abstaende == nil {
			continue
		}

		gemessen++
		gesamt = append(gesamt, abstaende...)
		schichten = append(schichten,
			schicht{name: satz, ganz: ganz, werte: abstaende})
	}

	if gemessen == 0 {
		t.Skip("keine MSE2-Ausgabe — scripts/mse2-referenz.sh Archenemy White")
	}

	// **Und es wurde Text gezeichnet.**
	//
	// # Warum das hier steht und nicht nur beim Scryfall-Lauf
	//
	// Weil dieselbe Klasse hier zwei Freigaben lang zugeschlagen hat, ohne
	// dass jemand es sah (#1170):
	// die Schriften der Vorlagen liegen unter der **Paketwurzel**, `TestMain`
	// gibt dem Lauf seit dem 2026-08-26 ein eigenes Zuhause, und damit wurde
	// jede Karte ohne Text gezeichnet.
	//
	// **Der Lauf war grün und die Zahl falsch**: 28,362 statt 18,853, und die
	// Zerlegung meldete 0 % `Text` statt 56 %. Wer sie las, sah einen
	// Rückschritt an der Rasterung, den es nicht gab.
	//
	// `scryfallabstand_netzlauf_test.go` hält dieselbe Zusage seit dem
	// 2026-08-23 mit [pruefeTextGemalt] — sie stand nur nie hier.
	if anteil := stand.teile.anteil(render.ArtText); anteil == 0 {
		t.Errorf("über %d Karten wurde keine einzige Textstelle gezeichnet — "+
			"fehlen die Schriften? MSE_SCHRIFTWURZEL zeigt auf die Wurzel des "+
			"Packs, nicht auf sein data-Verzeichnis (docs/RELEASE.md)", len(gesamt))
	}

	sort.Float64s(gesamt)

	t.Logf("gemessen: %d Karten aus %d Sets", len(gesamt), gemessen)

	// **Die ganze Verteilung, nicht nur ihr Mittel** — auf Anweisung des
	// Auftraggebers vom 2026-09-07.
	//
	// Ein Mittelwert über hunderte Karten verschweigt, wie sie sich verteilen:
	// „alle liegen dicht beieinander" und „die Hälfte ist perfekt, die andere
	// deutet daneben" ergeben dasselbe Mittel. Was gerechnet wird und wie,
	// steht in `kanonverteilung_intern_test.go`.
	//
	// **Diese Zeilen beschreiben die gezogenen Karten**, nicht den Kanon —
	// die Sets sind sehr verschieden stark gezogen (Archenemy 91 Prozent,
	// Black 15). Für den Kanon steht der geschichtete Schätzer darunter.
	meldeVerteilung(t, "die gezogenen Karten", gesamt)

	// **Und der Schätzer für den vollen Kanon** — geschichtet, mit
	// Endlichkeitskorrektur und gewichteten Quantilen. Die Gleichungen stehen
	// in `kanonstratum_intern_test.go`.
	geschaetzt := meldeKanon(t, schichten)

	mittel := geschaetzt.mittel

	stand.melde(t, "zusammen")
	stand.vorlagen.melde(t)
	meldeJeKarte(t, stand.karten)

	pruefeDeckel(t, mittel, gesamt[len(gesamt)-1])
}

// KartenJeVergleich ist die Stichprobe je Set.
//
// **Eine Stichprobe, mit Grund** (§10): `White` führt 3.761 Karten, und jeder
// Vergleich ist ein Rastern plus ein Bildvergleich über 375 × 523 Bildpunkte.
// Fünfzig je Set halten den Lauf unter einer Minute und decken beide Formen ab.
const KartenJeVergleich = 50

// messeSatzGegenMSE2 vergleicht die Stichprobe eines Sets; nil heißt „keine
// Ausgabe von MSE2".
// **Zwei Rückgaben, und beide tragen einen Namen**: die gemessenen Abstände
// und `N_h` — die Karten des Sets, aus denen gezogen wurde. Das zweite ist die
// Grundgesamtheit der Schicht und ohne Namen nicht von der Stichprobengröße zu
// unterscheiden.
func messeSatzGegenMSE2(
	t *testing.T, anwendung *App, satz string, stand *messstand,
) (abstaende []float64, ganz int) {
	t.Helper()

	if _, err := os.Stat(mse2ausgabe(satz)); err != nil {
		return nil, 0
	}

	geoeffnet := anwendung.SetOeffnen(t.Context(), filepath.Join(kanonverzeichnis(),
		satz+".mse-set"))
	if geoeffnet.Fehler != nil {
		t.Errorf("%s öffnen: %+v", satz, geoeffnet.Fehler)

		return nil, 0
	}

	karten := anwendung.Karten(t.Context(), geoeffnet.SetID)
	if karten.Fehler != nil {
		t.Errorf("%s: Karten: %+v", satz, karten.Fehler)

		return nil, 0
	}

	// **Die Schranke dieses Sets, aus seiner eigenen Kartenzahl.**
	jeSatz := stichprobeFuer(t, len(karten.Karten))
	stichprobe := verteilteStichprobe(karten.Karten, jeSatz)

	satzstand := neuerMessstand()
	satzstand.setstil = gewaehlterStil(t, anwendung, geoeffnet.SetID)

	abstaende, ohneReferenz := vergleicheMitMSE2(t, anwendung, satz, stichprobe,
		satzstand)
	if ohneReferenz > 0 {
		// **Was fehlt, wird genannt.** Sonst sieht eine Stichprobe von 50
		// aus wie 50, obwohl 6 still herausgefallen sind
		// (#474).
		t.Logf("  %-12s %3d Karten der Stichprobe haben kein Referenzbild",
			satz, ohneReferenz)
	}

	if len(abstaende) == 0 {
		// **Ein Set, das ganz ausfällt, ist ein Befund über die
		// Paketwurzel** (#862). `Archenemy` verlangt
		// `magic-archenemy.mse-style`; `M15-Magic-Pack` führt es nicht, und
		// dann zeichnet dieses Programm die Schemakarten im Kartenmaß des
		// Sets — 375×523 gegen 435×621, und der Vergleich weist sie ab.
		//
		// **Das ist keine Verschlechterung, sondern eine andere Menge.**
		// Genau dieser stille Ausfall hat den Deckel rot gemacht: Archenemy
		// ist mit 14,5 das **beste** Set, und ohne es steigt das Mittel von
		// 19,7 auf 20,0.
		t.Errorf("%s: keine einzige Karte verglichen — führt diese Paketwurzel "+
			"alle Stile des Sets? Archenemy braucht magic-archenemy.mse-style, "+
			"das M15-Magic-Pack nicht hat", satz)

		return nil, 0
	}

	sort.Float64s(abstaende)

	// **Je Set dieselben Kennzahlen wie über alles** — ein Mittel je Set sagte
	// nur, welches Set schlechter abschneidet, nicht **wie** es schlechter
	// abschneidet: an wenigen Ausreißern oder auf ganzer Breite.
	je := verteilungAus(abstaende)

	t.Logf("  %-12s %3d von %3d — Mittel %7.3f  Median %7.3f  "+
		"Min %7.3f  Max %8.3f  Q1 %7.3f  Q3 %7.3f  σ %6.3f",
		satz, len(abstaende), len(stichprobe),
		je.mittel, je.median, je.kleinster, je.groesster,
		je.q1, je.q3, je.abweichung)

	satzstand.melde(t, satz)
	stand.uebernimm(satzstand)

	return abstaende, len(karten.Karten)
}

// verteilteStichprobe nimmt `anzahl` Karten **über das ganze Set verteilt**.
//
// **Nicht die ersten.** Ein Fehler, der nur Karten hinter der fünfzigsten
// trifft, bliebe sonst unsichtbar — und ein Set führt bis zu 3.761
// (#474).
//
// **Ohne Zufall.** Eine Stichprobe, die bei jedem Lauf eine andere ist, macht
// aus einer Verschlechterung ein Rauschen; ein fester Schritt ist
// wiederholbar.
//
// # Der Schritt ist ein Bruch, keine ganze Zahl
//
// Hier stand `schritt := len(karten) / anzahl`, und damit endete die
// Stichprobe **vor** dem Ende der Gruppe: bei 33 Karten und zwölf Plätzen ist
// der ganzzahlige Schritt 2, gezogen werden die Stellen 0 bis 22, und die
// letzten zehn Karten sind unerreichbar.
//
// Bei einem Set mit 3.761 Karten fiel das nicht auf — bei einer Vorlagengruppe
// mit 33 trägt es die halbe Aussage
// (#536).
// Gefunden vom Zweitblick durch Codex über Phase I.
//
// **Jetzt wird die Stelle gerechnet**: `i · (len−1) / (anzahl−1)` trifft die
// erste und die **letzte** Karte und verteilt die übrigen gleichmäßig
// dazwischen.
func verteilteStichprobe(karten []KarteDTO, anzahl int) []KarteDTO {
	if len(karten) <= anzahl || anzahl <= 0 {
		return karten
	}

	if anzahl == 1 {
		return karten[:1]
	}

	gewaehlt := make([]KarteDTO, 0, anzahl)
	for i := range anzahl {
		gewaehlt = append(gewaehlt, karten[i*(len(karten)-1)/(anzahl-1)])
	}

	return gewaehlt
}

// pruefeDeckel hält den heutigen Stand fest — im Mittel und im schlechtesten
// Fall.
//

func pruefeDeckel(t *testing.T, mittel, schlechtester float64) {
	t.Helper()

	if mittel > deckelMittel {
		t.Errorf("mittlerer Abstand %.3f über dem Deckel %.3f — die Karten sind "+
			"weiter von MSE2 weg als am 2026-08-18", mittel, deckelMittel)
	}

	if mittel < deckelMittel-2 {
		t.Logf("mittlerer Abstand %.3f, Deckel %.3f — der Deckel gehört "+
			"nachgezogen", mittel, deckelMittel)
	}

	if schlechtester > deckelSchlechtester {
		t.Errorf("die schlechteste Karte liegt bei %.3f, über dem Deckel %.3f",
			schlechtester, deckelSchlechtester)
	}
}
