//go:build netzlauf

// **Wie weit eine gezeichnete Karte von einer gedruckten liegt** —
// #1029,
// Werkstück Ω.6.
//
// # Was er misst, das der Kanonlauf nicht misst
//
// Der Kanonlauf fragt: **rechnen wir wie MSE2?** Er vergleicht gegen eine
// MSE2-Ausgabe, und sein Ziel ist null.
//
// Dieser Lauf fragt etwas anderes: **sieht eine Karte aus wie eine
// gedruckte?** Sein Ziel ist **nicht** null, und das ist keine Schwäche der
// Messung, sondern ihre Aussage. Die Vorlage `magic-m15` ist ein Nachbau des
// echten Rahmens; sie benutzt Schriften, die jemand für frei verfügbar hält,
// und setzt Text an Stellen, die jemand abgemessen hat. Was hier
// herauskommt, ist der Abstand **dieses Nachbaus** zur Vorlage — und der
// gehört MSE2 und nicht diesem Programm.
//
// **Wozu er dann gut ist**: er sagt, **wo** der Nachbau danebenliegt. Ein
// Rahmen, der um zwei Punkte verschoben ist, und eine Typzeile, die eine
// andere Schrift trägt, sind verschiedene Befunde — und nur der zweite lohnt
// eine Portierung.
//
// # Warum er nicht in der Pipeline läuft
//
// Dieselben drei Gründe wie bei `internal/scryfall/netzlauf_test.go`: ein
// fremder Rechner, ein fremdes Kontingent, keine Wiederholbarkeit. Er zählt
// nicht zur Abdeckung.
//
// # Wie er gefahren wird
//
//	MSE_KANON=~/Library/Application\ Support/MagicSetEditor4Mac/kanonische-sets \
//	MSE_FREMDPAKETE=/…/full-magic-pack/data \
//	  go test -tags netzlauf ./internal/app/ -run TestWieWeitVonEinerGedruckten -v
//
// # Was er über die Druckfassung **nicht** sagt
//
// Die Karte wird allein über ihren **Namen** gesucht (`/cards/named?exact=`).
// Setcode und Sammlernummer reisen nicht mit; welche Druckfassung Scryfall
// daraufhin liefert, entscheidet Scryfall. Bei einer Karte, die es in mehreren
// Rahmen gibt, kann das eine andere sein als die im MSE-Set.
//
// **Für die Frage dieses Laufs trägt das trotzdem**: verglichen werden Rahmen
// und Textlage einer *gedruckten* Karte gegen unseren Nachbau, und dafür ist
// jede echte Druckfassung ein gültiger Gegenstand. Für eine Aussage über
// **eine bestimmte** Fassung wäre sie es nicht — angestrichen vom Zweitblick
// über Phase Ω.
//
// # Kein fremdes Bild bleibt liegen
//
// Geholte Bilder liegen im Zwischenspeicher des Laufs (`t.TempDir()`) und sind
// nach ihm fort — §11: nichts Fremdes kommt in den Baum.
package app

import (
	"bytes"
	"context"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"sort"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// GedruckteKarten ist die Zahl der Karten, die verglichen werden.
//
// **Klein mit Absicht.** Jede kostet eine Anfrage an einen fremden Dienst;
// die Aussage hängt an der Zerlegung und nicht an der Stichprobengröße.
const GedruckteKarten = 12

// TestWieWeitVonEinerGedruckten vergleicht gezeichnete gegen gedruckte Karten.
func TestWieWeitVonEinerGedruckten(t *testing.T) {
	// # Warum `HOME` hier **nicht** umgelenkt wird
	//
	// §3 verlangt es für jeden Test, der nach draußen greift, und der erste
	// Anlauf tat es. **Das Ergebnis sah aus wie eine Messung und war keine**:
	// die Schriften der bezogenen Pakete liegen unter `HOME`, und mit einem
	// leeren Zuhause zeichnete jede Karte **ohne Text**. Gezählt über die
	// Herkunft: `map[Bild:326349 Fläche:5449 Symbolzug:530]` — keine einzige
	// Textstelle. Der Lauf meldete „Text 0,000" und war grün.
	//
	// **Was §3 verhindern will, ist hier anders gelöst**: der
	// Zwischenspeicher des Klienten ist ein `t.TempDir()` und damit
	// ausdrücklich nicht der dieses Rechners; die Adresse steht als
	// `scryfall.Adresse` da. Was bleibt, ist die Abhängigkeit von den
	// **installierten Schriften** — und die ist der Gegenstand der Messung,
	// nicht ihr Nebeneffekt.
	//
	// **Damit sie nicht wieder still ausfällt**, prüft [pruefeTextGemalt] sie
	// bei jeder Karte.

	anwendung := gestartetMit(t, func() string { return fremdpaketwurzelOderSkip(t) })

	geoeffnet := anwendung.SetOeffnen(context.Background(), kanonsatzpfad(t, "Black"))
	if geoeffnet.Fehler != nil {
		t.Fatalf("einfahren: %+v", geoeffnet.Fehler)
	}

	zeilen := anwendung.Kartenzeilen(context.Background(), geoeffnet.SetID)
	if zeilen.Fehler != nil || len(zeilen.Zeilen) < GedruckteKarten {
		t.Skipf("zu wenige Karten: %+v", zeilen.Fehler)
	}

	rasterer, ok := anwendung.rasterer.(*render.Rasterer)
	if !ok {
		t.Skip("kein Rasterer")
	}

	klient := gedrucktenklient(t)
	schritt := len(zeilen.Zeilen) / GedruckteKarten

	var (
		gesammelt []gedrucktervergleich
		verfehlt  []string
	)

	for i := range GedruckteKarten {
		zeile := zeilen.Zeilen[i*schritt]

		vergleich, grund := einegedruckte(t, anwendung, rasterer, klient, zeile.ID, zeile.Name)
		if grund != "" {
			verfehlt = append(verfehlt, zeile.Name+": "+grund)

			continue
		}

		gesammelt = append(gesammelt, vergleich)
	}

	meldeGedruckte(t, gesammelt, verfehlt)
}

// gedrucktervergleich hält den Abstand einer Karte, nach Bereichen getrennt.
type gedrucktervergleich struct {
	name string

	// gesamt ist der Abstand über alles außer der Illustration.
	gesamt float64

	// text ist der Abstand dort, wo **wir** Text gemalt haben.
	text float64

	// rahmen ist der Abstand überall sonst — ohne Illustration, ohne Text.
	rahmen float64
}

// gedrucktenklient baut einen Klienten gegen den echten Dienst.
//
// **Mit eigenem Zwischenspeicher je Lauf**, wie in
// `internal/scryfall/netzlauf_test.go`: ein geteilter machte den zweiten Lauf
// zu einer Prüfung des Speichers.
func gedrucktenklient(t *testing.T) *scryfall.Klient {
	t.Helper()

	return scryfall.NeuerKlient(scryfall.Adresse,
		netz.NeuerKlient(netz.UeblicheFristen()),
		netz.NeuerSpeicher(t.TempDir(), scryfall.Speicherfrist, time.Now),
		time.Now, "netzlauf")
}

// einegedruckte vergleicht eine Karte — oder sagt, warum nicht.
//
// **Ein Grund und kein Fehlschlag.** Eine Karte, die Scryfall nicht kennt, ist
// keine Aussage über diesen Rasterer; sie wird gezählt und benannt, damit die
// Stichprobe nicht still schrumpft (§10).
func einegedruckte(
	t *testing.T, anwendung *App, rasterer *render.Rasterer,
	klient *scryfall.Klient, karteID, name string,
) (vergleich gedrucktervergleich, grund string) {
	t.Helper()

	karte, _, _, err := klient.KarteNachName(context.Background(), name)
	if err != nil {
		return gedrucktervergleich{}, "nicht gefunden"
	}

	fund, err := klient.Bild(context.Background(), karte, scryfall.BildNormal)
	if err != nil {
		return gedrucktervergleich{}, "kein Bild"
	}

	gedruckt, _, err := image.Decode(bytes.NewReader(fund.Daten))
	if err != nil {
		return gedrucktervergleich{}, "unlesbar"
	}

	// **Ab hier ist jeder Fehler unserer.** Was Scryfall nicht liefert, ist
	// keine Aussage über den Rasterer und wird übergangen; was **dieses**
	// Programm nicht kann, ist eine — und darf nicht als Schrumpfung
	// durchgehen.
	//
	// Angestrichen vom Nachlauf des Zweitblicks über Phase Ω: bis dahin
	// konnten bis zu vier eigene Fehler den Lauf grün lassen, und die
	// Begründung daneben trug nur „Scryfall kennt die Karte nicht".
	// **Kein Kartenplan ist unser Fehler, nicht Scryfalls** — der Helfer bricht
	// dann ab und nennt den Befund.
	plan := planZurKarte(t, anwendung, karteID)

	// **Auf die Breite des gedruckten Bildes**, nicht auf eine runde Zahl: so
	// wird das fremde Bild nicht angefasst. Es zu skalieren verstieße gegen
	// Scryfalls Bedingung („Do not distort, skew, or stretch card images").
	vergroesserung := float64(gedruckt.Bounds().Dx()) / plan.Breite

	unser, herkunft, _, err := rasterer.ZeichneMitHerkunft(
		context.Background(), plan, vergroesserung)
	if err != nil {
		t.Fatalf("%s: nicht gezeichnet — das ist unser Fehler: %v", name, err)
	}

	if grund := pruefeTextGemalt(herkunft, unser.Bounds()); grund != "" {
		t.Fatalf("%s: %s", name, grund)
	}

	return vermisseGegenGedruckte(name, unser, herkunft, gedruckt,
		bildkasten(plan, vergroesserung)), ""
}

// bildkasten gibt den Rahmen der Illustration in Bildpunkten — oder leer.
//
// **Aus dem Plan und nicht aus der Herkunft.** Die Herkunft sagt „hier stand
// ein Bild"; das gilt für den Kartenrahmen genauso wie für die Illustration.
// Gefragt ist die eine Ebene, die das Kunstfeld malt.
func bildkasten(plan *render.Plan, vergroesserung float64) image.Rectangle {
	for _, ebene := range plan.Ebenen {
		if ebene.Art != render.ArtBild || ebene.Feld != "image" {
			continue
		}

		return image.Rect(
			int(ebene.Rahmen.Links*vergroesserung),
			int(ebene.Rahmen.Oben*vergroesserung),
			int((ebene.Rahmen.Links+ebene.Rahmen.Breite)*vergroesserung),
			int((ebene.Rahmen.Oben+ebene.Rahmen.Hoehe)*vergroesserung),
		)
	}

	return image.Rectangle{}
}

// meldeGedruckte schreibt das Ergebnis auf.
func meldeGedruckte(t *testing.T, gesammelt []gedrucktervergleich, verfehlt []string) {
	t.Helper()

	// **Erst sagen, was fehlt.** Eine Stichprobe, die still schrumpft, sieht
	// aus wie eine vollständige (§10).
	t.Logf("%d von %d Karten verglichen", len(gesammelt), GedruckteKarten)

	for _, grund := range verfehlt {
		t.Logf("  übergangen — %s", grund)
	}

	// **Eine Stichprobe, die schrumpft, ist eine andere Stichprobe.** Hier
	// stand „nur bei null wird es rot", und damit hätten elf von zwölf
	// ausfallen dürfen, während die Zahl weiter als „Lauf über zwölf Karten"
	// gelesen wurde. Angestrichen vom Zweitblick über Phase Ω.
	//
	// **Mehr als ein Drittel Ausfall ist zu viel** — und „mehr als", nicht
	// „ab": bei zwölf sind zwei Drittel acht, und genau vier Ausfälle ließen
	// acht Ergebnisse und damit Grün. Der Nachlauf des Zweitblicks hat den
	// Abstand zwischen dem Satz und der Bedingung angestrichen; hier steht
	// jetzt die Bedingung, und der Satz sagt dasselbe.
	//
	// **Übergangen wird nur, was Scryfall betrifft.** Eigene Fehler halten
	// den Lauf oben schon an.
	if genug := (GedruckteKarten*2)/3 + 1; len(gesammelt) < genug {
		t.Fatalf("nur %d von %d Karten verglichen, verlangt sind %d — "+
			"die Stichprobe trägt die Aussage nicht", len(gesammelt),
			GedruckteKarten, genug)
	}

	sort.Slice(gesammelt, func(i, j int) bool {
		return gesammelt[i].gesamt < gesammelt[j].gesamt
	})

	var gesamt, text, rahmen float64

	for _, eine := range gesammelt {
		gesamt += eine.gesamt
		text += eine.text
		rahmen += eine.rahmen
	}

	anzahl := float64(len(gesammelt))

	t.Logf("Abstand zur gedruckten Karte (ohne Illustration), Mittel über %d:",
		len(gesammelt))
	t.Logf("  gesamt %6.3f   Text %6.3f   Rahmen %6.3f",
		gesamt/anzahl, text/anzahl, rahmen/anzahl)

	t.Logf("  beste  %6.3f  (%s)", gesammelt[0].gesamt, gesammelt[0].name)
	t.Logf("  weiteste %6.3f  (%s)",
		gesammelt[len(gesammelt)-1].gesamt, gesammelt[len(gesammelt)-1].name)
}

// pruefeTextGemalt sagt, ob überhaupt eine Textstelle entstanden ist.
//
// **Ohne diesen Blick ist der Textabstand null und der Lauf grün.** Genau das
// war der erste Anlauf: mit umgelenktem `HOME` fand die Schriftsuche nichts,
// jede Karte kam ohne Text heraus, und die Zahl „Text 0,000" sah aus wie ein
// gutes Ergebnis.
//
// **Ein Grund und kein `t.Fatal`**: die Karte wird übergangen und benannt,
// wie eine, die Scryfall nicht kennt. Fehlt der Text bei **allen**, bleibt
// keine übrig, und [meldeGedruckte] hält den Lauf an.
func pruefeTextGemalt(herkunft *render.Herkunft, rahmen image.Rectangle) string {
	for y := range rahmen.Dy() {
		for x := range rahmen.Dx() {
			if herkunft.Art(x, y) == render.ArtText {
				return ""
			}
		}
	}

	return "keine Textstelle gezeichnet — fehlen die Schriften?"
}
