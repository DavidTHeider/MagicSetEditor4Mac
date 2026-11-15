package service_test

// **Münzwurf und Würfel** — Werkstück M15 der Regelmaschine
// (#1617).
//
// # Wogegen hier geprüft wird
//
//   - **Zwei Seiten** (CR 705.1), und wer ruft, gewinnt oder verliert
//     (CR 705.2).
//   - **Die Seitenzahl kommt mit** (CR 706.1), und die Augen stehen einzeln
//     da (CR 706.4).
//   - **Die Quelle ist dieselbe wie beim Mischen** — das prüft der Typ:
//     `festeMischung` erfüllt beide Methoden, und es gibt nur eine
//     Schnittstelle.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Münze hat zwei Seiten, und beide kommen vor** — CR 705.1.
func TestEineMuenzeHatZweiSeiten(t *testing.T) {
	t.Parallel()

	for zahl, erwartet := range map[int]string{0: "kopf", 1: "zahl"} {
		t.Run(erwartet, func(t *testing.T) {
			t.Parallel()

			kasten := service.NeuerSandkasten(festeMischung{zahl: zahl})

			ereignisse, _, fehler := kasten.Auslegen(eineZugpartie(t),
				domain.Spielbefehl{
					PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
					Art: service.BefehlMuenzwurf,
				})
			if fehler != nil {
				t.Fatalf("münzwurf: %v", fehler)
			}

			switch {
			case len(ereignisse) != 1:
				t.Fatalf("%d Ereignisse", len(ereignisse))
			case ereignisse[0].Oeffentlich["seite"] != erwartet:
				t.Errorf("gefallen ist %v", ereignisse[0].Oeffentlich)

			// **Ohne Ruf steht kein Ausgang da** — CR 705.2: bei dieser
			// Sorte Wurf gewinnt und verliert niemand.
			case ereignisse[0].Oeffentlich["gewonnen"] != nil:
				t.Errorf("ohne Ruf steht ein Ausgang da: %v",
					ereignisse[0].Oeffentlich)
			}
		})
	}
}

// **Wer ruft, gewinnt oder verliert** — CR 705.2.
func TestWerRuftGewinntOderVerliert(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		ruf      string
		zahl     int
		gewonnen bool
	}{
		"Kopf gerufen, Kopf gefallen": {ruf: "kopf", zahl: 0, gewonnen: true},
		"Kopf gerufen, Zahl gefallen": {ruf: "kopf", zahl: 1, gewonnen: false},
		"Zahl gerufen, Zahl gefallen": {ruf: "zahl", zahl: 1, gewonnen: true},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			kasten := service.NeuerSandkasten(festeMischung{zahl: fall.zahl})

			ereignisse, _, fehler := kasten.Auslegen(eineZugpartie(t),
				domain.Spielbefehl{
					PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
					Art:      service.BefehlMuenzwurf,
					Nutzlast: map[string]any{"ruf": fall.ruf},
				})
			if fehler != nil {
				t.Fatalf("münzwurf: %v", fehler)
			}

			if ereignisse[0].Oeffentlich["gewonnen"] != fall.gewonnen {
				t.Errorf("das Ereignis sagt %v", ereignisse[0].Oeffentlich)
			}
		})
	}
}

// **Ein Ruf, den es nicht gibt, wird abgewiesen** — CR 705.1.
func TestEinRufDenEsNichtGibtWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, _, fehler := kasten.Auslegen(eineZugpartie(t), domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
		Art:      service.BefehlMuenzwurf,
		Nutzlast: map[string]any{"ruf": "kante"},
	})

	if fehler == nil || !strings.Contains(fehler.Error(), "CR 705.1") {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}

// **Ein Würfel bringt seine Seitenzahl mit** — CR 706.1.
func TestEinWuerfelBringtSeineSeitenzahlMit(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{zahl: 3})

	ereignisse, _, fehler := kasten.Auslegen(eineZugpartie(t),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art:      service.BefehlWuerfeln,
			Nutzlast: map[string]any{"seiten": float64(6)},
		})
	if fehler != nil {
		t.Fatalf("würfeln: %v", fehler)
	}

	augen, ist := ereignisse[0].Oeffentlich["augen"].([]any)

	switch {
	case !ist || len(augen) != 1:
		t.Fatalf("die Augen sind %v", ereignisse[0].Oeffentlich["augen"])

	// **Ab eins gezählt** — CR 706.2 spricht von der Zahl auf der oberen
	// Fläche, und die fängt bei eins an. Die Quelle gibt drei, der Würfel
	// zeigt vier.
	case augen[0] != float64(4):
		t.Errorf("gefallen ist %v", augen[0])
	}
}

// **Mehrere Würfel stehen einzeln da** — CR 706.4.
//
// **Und nicht als Summe**: was mit den Augen geschieht, sagt der Kartentext;
// eine Summe wäre schon eine Auswertung (§10).
func TestMehrereWuerfelStehenEinzelnDa(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{zahl: 5})

	ereignisse, _, fehler := kasten.Auslegen(eineZugpartie(t),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art: service.BefehlWuerfeln,
			Nutzlast: map[string]any{
				"seiten": float64(20), "wuerfel": float64(3),
			},
		})
	if fehler != nil {
		t.Fatalf("würfeln: %v", fehler)
	}

	augen, ist := ereignisse[0].Oeffentlich["augen"].([]any)

	if !ist || len(augen) != 3 {
		t.Fatalf("die Augen sind %v", ereignisse[0].Oeffentlich["augen"])
	}

	for _, eines := range augen {
		if eines != float64(6) {
			t.Errorf("gefallen ist %v", eines)
		}
	}

	// **Und keine Summe** — die Gegenprobe (§10).
	if _, steht := ereignisse[0].Oeffentlich["summe"]; steht {
		t.Error("das Ereignis nennt eine Summe")
	}
}

// **Ein Würfel, den es nicht gibt, wird abgewiesen** — CR 706.1 und §8.
func TestEinWuerfelDenEsNichtGibtWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	faelle := map[string]map[string]any{
		"ohne Seitenzahl": {},
		"eine Seite":      {"seiten": float64(1)},
		"zu viele Seiten": {"seiten": float64(1000)},
		"null Würfel": {
			"seiten": float64(6), "wuerfel": float64(0),
		},
		"zu viele Würfel": {
			"seiten": float64(6), "wuerfel": float64(1000),
		},
	}

	for name, nutzlast := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if _, _, fehler := kasten.Auslegen(eineZugpartie(t),
				domain.Spielbefehl{
					PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
					Art: service.BefehlWuerfeln, Nutzlast: nutzlast,
				}); fehler == nil {
				t.Error("der Würfel kam durch")
			}
		})
	}
}

// **Eine Quelle, die klagt, lässt niemanden werfen** — §8.
func TestEineKlagendeQuelleLaesstNiemandenWerfen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{klagt: errZufallWeg})

	for _, art := range []string{
		service.BefehlMuenzwurf, service.BefehlWuerfeln,
	} {
		t.Run(art, func(t *testing.T) {
			t.Parallel()

			if _, _, fehler := kasten.Auslegen(eineZugpartie(t),
				domain.Spielbefehl{
					PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
					Art:      art,
					Nutzlast: map[string]any{"seiten": float64(6)},
				}); fehler == nil {
				t.Error("es wurde geworfen, obwohl die Quelle zu ist")
			}
		})
	}
}

// **Ein Wurf ändert nichts am Tisch** — CR 705, CR 706.
//
// **Die Gegenprobe zu allem darüber** (§10): ohne sie sagten die Läufe nur,
// dass **ein Ereignis** entsteht — ein Weg, der nebenbei Karten verschöbe,
// käme genauso durch.
func TestEinWurfAendertNichtsAmTisch(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})
	partie := eineZugpartie(t)

	_, danach, fehler := kasten.Auslegen(partie, domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
		Art: service.BefehlMuenzwurf,
	})
	if fehler != nil {
		t.Fatalf("münzwurf: %v", fehler)
	}

	switch {
	case len(danach.Objekte) != len(partie.Objekte):
		t.Errorf("%d Objekte statt %d",
			len(danach.Objekte), len(partie.Objekte))
	case danach.Zug.Abschnitt != partie.Zug.Abschnitt:
		t.Errorf("der Zug steht bei %+v", danach.Zug.Abschnitt)
	case danach.Sitze[0].Leben != partie.Sitze[0].Leben:
		t.Errorf("das Leben steht bei %d", danach.Sitze[0].Leben)
	}
}

// **Ein nachträglich geänderter Wurf bricht die Kette** — G1 (#1458), und
// die Abnahme von M15 verlangt es **nachgewiesen, nicht behauptet**.
//
// # Was hier wirklich geprüft wird
//
// **Dass das Ergebnis in den Abdruck eingeht.** Die Kette ist geerbt: jedes
// Ereignis hasht den vorigen Abdruck zusammen mit seinem **öffentlichen
// Teil**. Ein Wurf, der sein Ergebnis nur nebenbei mitführte, ginge nicht
// ein — und dann ließe sich „Kopf" nachträglich zu „Zahl" machen, ohne dass
// eine Prüfung anschlüge.
//
// **Gefahren, nicht gelesen** (§10): derselbe Wurf bekommt zweimal einen
// Abdruck, einmal unverändert und einmal mit umgedrehter Seite — und die
// beiden müssen sich unterscheiden.
func TestEinGeaenderterWurfBrichtDieKette(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	ereignisse, _, fehler := kasten.Auslegen(eineZugpartie(t),
		domain.Spielbefehl{
			PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
			Art:      service.BefehlMuenzwurf,
			Nutzlast: map[string]any{"ruf": "kopf"},
		})
	if fehler != nil {
		t.Fatalf("münzwurf: %v", fehler)
	}

	wurf := ereignisse[0]
	wurf.PartieID = "p1"
	wurf.Folge = 1

	echt, fehler := wurf.Abdruck("vorher")
	if fehler != nil {
		t.Fatalf("abdruck: %v", fehler)
	}

	// **Dieselbe Stelle, ein anderes Ergebnis** — so sähe eine nachträgliche
	// Änderung aus.
	gefaelscht := wurf
	gefaelscht.Oeffentlich = map[string]any{}

	for name, wert := range wurf.Oeffentlich {
		gefaelscht.Oeffentlich[name] = wert
	}

	gefaelscht.Oeffentlich["seite"] = "zahl"
	gefaelscht.Oeffentlich["gewonnen"] = false

	anders, fehler := gefaelscht.Abdruck("vorher")
	if fehler != nil {
		t.Fatalf("abdruck der Fälschung: %v", fehler)
	}

	if echt == anders {
		t.Error("der geänderte Wurf hat denselben Abdruck — das Ergebnis " +
			"geht nicht in die Kette ein")
	}

	// **Die Gegenprobe**: derselbe Wurf gibt zweimal denselben Abdruck.
	// Ohne sie sagte der Lauf darüber nur, dass **irgendetwas** verschieden
	// ist — auch eine Kodierung mit Zufall darin käme durch (§10).
	nocheinmal, fehler := wurf.Abdruck("vorher")
	if fehler != nil {
		t.Fatalf("zweiter Abdruck: %v", fehler)
	}

	if echt != nocheinmal {
		t.Error("derselbe Wurf gibt zwei Abdrücke")
	}
}

// **Die echte Quelle gibt Zahlen im Bereich** — CR 705, CR 706.
//
// # Warum sie hier einzeln geprüft wird
//
// **Weil die Läufe darüber ein Doppel benutzen** (§3), und ein Doppel sagt
// nichts über die Quelle. Was hier trägt, ist die Zusage von
// [service.Zufallsquelle.Zahl]: eine Zahl in `[0, grenze)`, und eine Grenze
// unter eins gibt es nicht.
func TestDieEchteQuelleGibtZahlenImBereich(t *testing.T) {
	t.Parallel()

	quelle := service.NeuerZufall()

	for range 200 {
		gezogen, fehler := quelle.Zahl(6)
		if fehler != nil {
			t.Fatalf("zahl: %v", fehler)
		}

		if gezogen < 0 || gezogen > 5 {
			t.Fatalf("gezogen wurde %d", gezogen)
		}
	}

	// **Und eine Grenze unter eins gibt es nicht** — sonst teilte die
	// Rechnung durch null.
	for _, grenze := range []int{0, -1} {
		if _, fehler := quelle.Zahl(grenze); fehler == nil {
			t.Errorf("die Grenze %d kam durch", grenze)
		}
	}
}

// **Eine Würfelzahl, die keine Zahl ist, wird abgewiesen** — §8.
//
// **Der Zweig, den dieser Lauf betritt**: das Feld `wuerfel` steht da, ist
// aber keine Zahl. Der Lauf daneben lässt es **weg**, und das ist ein
// anderer Weg (§10).
func TestEineWuerfelzahlDieKeineIstWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	kasten := service.NeuerSandkasten(festeMischung{})

	_, _, fehler := kasten.Auslegen(eineZugpartie(t), domain.Spielbefehl{
		PartieID: "p1", Absender: "sitz-1", Kennung: "b1",
		Art: service.BefehlWuerfeln,
		Nutzlast: map[string]any{
			"seiten": float64(6), "wuerfel": "drei",
		},
	})

	if fehler == nil || !strings.Contains(fehler.Error(), "wuerfel") {
		t.Errorf("der Fehler lautet %v", fehler)
	}
}
