package app_test

// **Die Balanceanalyse** — Werkstück A13.2b von Phase 5.3
// (#1452).
//
// # Was hier geprüft wird
//
// **Dass die Angaben, die das Programm macht, vom Programm kommen.** Ein
// Modell weiß nicht, wie seine Belege gefunden wurden; Strategie, Format und
// Stand stellt deshalb `internal/app` zusammen, und dieser Lauf hält es fest.
//
// **Und dass ein Vergleich ohne Unterschied auffällt.** „Ähnlich" sagt der
// Rang schon, der die Karte gefunden hat — und der bedeutet nichts (0103).

import (
	"fmt"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// balanceSagt stellt die Modellantwort für den Balanceweg ein.
func (w hinweiswelt) balanceSagt(inhalt string) {
	w.laufzeit.antwort = app.KIRohantwort{Inhalt: inhalt, Modell: "gemma4:8b"}
}

// einVollstaendigerVergleich ist eine Antwort, an der nichts fehlt.
const einVollstaendigerVergleich = `{"einschaetzung":"Sie kostet eins mehr als [B1].",` +
	`"einschraenkungen":["Nur die mitgegebenen Stellen."],` +
	`"vergleiche":[{"beleg":"B1","name":"509.1 Blocks erklären",` +
	`"unterschiede":["ein Mana teurer","trifft auch Spieler"]}],` +
	`"nichtBetrachtet":["Zusammenspiel mit Manaflut"]}`

// balance fragt nach einem Vergleich.
func (w hinweiswelt) balance(t *testing.T, format string) app.KIEinschaetzungAntwort {
	t.Helper()

	return w.anwendung.KIEinschaetzen(probekontext(t),
		app.KIEinschaetzungsbezug{
			Art: string(domain.EinschaetzungBalance), KarteID: w.karte,
			Anbieter: "ollama", Alias: "gemma4:8b",
			Frage: "Wie steht sie im Vergleich da?", Format: format,
		})
}

// **Ein vollständiger Vergleich bringt Strategie, Format, Stand und
// Unterschiede mit.**
func TestEineBalanceanalyseNenntStrategieUndUnterschiede(t *testing.T) {
	welt := einehinweiswelt(t)
	welt.balanceSagt(einVollstaendigerVergleich)
	welt.recherche.paket.Ergebnis.Modus = "hybrid"

	antwort := welt.balance(t, "Commander")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Balance == nil {
		t.Fatal("die Antwort trägt keine Balanceangaben")
	}

	if !antwort.Belegt {
		t.Errorf("nicht belegt: %+v", antwort.Befunde)
	}

	// **Die Strategie kommt vom Programm** — sie nennt den Fall, den Modus
	// und die Grenzen, und sagt, dass kein Reranker im Spiel ist.
	//
	// **„hybrid" steht hier, weil das Doppel es so gesucht hat** — die Zeile
	// darüber stellt es ein. Bis zum 2026-09-13 stand der Modus nur im Kopf
	// dieses Absatzes: eine Gegenprobe, die ihn zur Konstante machte, blieb
	// grün (§10).
	for _, teil := range []string{
		"aehnlich", "Modus hybrid", "höchstens", "kein Reranker",
	} {
		if !strings.Contains(antwort.Balance.Strategie, teil) {
			t.Errorf("die Strategie nennt %q nicht: %q", teil,
				antwort.Balance.Strategie)
		}
	}

	if antwort.Balance.Format != "Commander" {
		t.Errorf("das Format ist %q", antwort.Balance.Format)
	}

	if len(antwort.Balance.Staende) != 1 {
		t.Errorf("die Stände sind %+v", antwort.Balance.Staende)
	}

	if len(antwort.Balance.Vergleiche) != 1 ||
		len(antwort.Balance.Vergleiche[0].Unterschiede) != 2 {
		t.Errorf("die Vergleiche sind %+v", antwort.Balance.Vergleiche)
	}

	if len(antwort.Balance.NichtBetrachtet) != 1 {
		t.Errorf("die Synergien sind %+v", antwort.Balance.NichtBetrachtet)
	}

	// **Und das Modell bekam die Balanceanweisung** — nicht die des
	// Regelhinweises und nicht die der Kartenfelder.
	if len(welt.laufzeit.anweisungen) != 1 ||
		welt.laufzeit.anweisungen[0] != app.KIBalanceanweisung {
		t.Errorf("die Anweisungen sind %+v", welt.laufzeit.anweisungen)
	}

	// **Das Format steht im Auftrag und nicht nur in der Antwort.**
	//
	// Der Kasten zeigt „Format: Commander"; wenn es das Modell nie erreicht,
	// ist diese Zeile eine Behauptung über einen Vergleich, der sie nicht
	// kannte. Bis zum 2026-09-13 hielt das kein Lauf — eine Gegenprobe, die
	// `balancefrage` das Format nehmen ließ, blieb grün (§10).
	if len(welt.laufzeit.auftraege) != 1 ||
		!strings.Contains(welt.laufzeit.auftraege[0], "Format: Commander") {
		t.Errorf("der Auftrag ist %+v", welt.laufzeit.auftraege)
	}
}

// **Ein Index ohne Einbettungsmodell sucht lexikalisch — und die Strategie
// sagt es.**
//
// **Das ist der Zweig, für den `suchstrategie` aus dem Ergebnis liest.** Eine
// Konstante behauptete „hybrid", und zwar genau dort, wo jemand wissen will,
// warum gerade diese Karten dastehen.
func TestDieStrategieNenntDenGemessenenModus(t *testing.T) {
	welt := einehinweiswelt(t)
	welt.balanceSagt(einVollstaendigerVergleich)
	welt.recherche.paket.Ergebnis.Modus = "lexikalisch"

	antwort := welt.balance(t, "Commander")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if !strings.Contains(antwort.Balance.Strategie, "Modus lexikalisch") {
		t.Errorf("die Strategie ist %q", antwort.Balance.Strategie)
	}

	// **Und ohne gemessenen Modus steht „keiner" da**, nicht nichts: ein
	// Satz, in dem das Wort fehlt, liest sich wie eine vollständige Auskunft.
	welt.recherche.paket.Ergebnis.Modus = ""

	zweite := welt.balance(t, "Commander")
	if !strings.Contains(zweite.Balance.Strategie, "Modus keiner") {
		t.Errorf("die Strategie ist %q", zweite.Balance.Strategie)
	}
}

// **Ohne Format sagt die Antwort es — als Einschränkung und nicht als
// Befund.**
func TestOhneFormatStehtEsInDenEinschraenkungen(t *testing.T) {
	welt := einehinweiswelt(t)
	welt.balanceSagt(einVollstaendigerVergleich)

	antwort := welt.balance(t, "")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if !antwort.Belegt {
		t.Errorf("das fehlende Format machte sie unbelegt: %+v", antwort.Befunde)
	}

	if !strings.Contains(strings.Join(antwort.Einschraenkungen, " "),
		"kein Format") {
		t.Errorf("die Einschränkungen sind %+v", antwort.Einschraenkungen)
	}

	// **Und im Auftrag steht dann keine Formatzeile** — ein „Format: " ohne
	// Wert dahinter liest sich für ein Modell wie ein leeres Format und nicht
	// wie keines.
	if strings.Contains(welt.laufzeit.auftraege[0], "Format:") {
		t.Errorf("der Auftrag nennt ein Format: %q", welt.laufzeit.auftraege[0])
	}
}

// **Ein Vergleich ohne Unterschied steht als Befund da.**
func TestEinVergleichOhneUnterschiedStehtAlsBefund(t *testing.T) {
	welt := einehinweiswelt(t)
	welt.balanceSagt(`{"einschaetzung":"Sie ähnelt [B1].",` +
		`"vergleiche":[{"beleg":"B1","name":"Blocks","unterschiede":[]}]}`)

	antwort := welt.balance(t, "Commander")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Belegt {
		t.Fatal("der leere Vergleich kam durch")
	}

	if len(antwort.Befunde) != 1 ||
		antwort.Befunde[0].Art != domain.BefundOhneUnterschied {
		t.Errorf("die Befunde sind %+v", antwort.Befunde)
	}
}

// **Und eine Regeleinschätzung trägt keine Balanceangaben.**
//
// **`nil` und nicht ein leerer Verbund**: eine Regelfrage hat keine
// Vergleichskarten, und ein leeres Feld sähe aus, als hätte die Suche keine
// gefunden. Das sind zwei Auskünfte.
func TestEineRegeleinschaetzungHatKeineBalanceangaben(t *testing.T) {
	welt := einehinweiswelt(t)
	welt.modellSagt("Der Block ist legal [B1].")

	if antwort := welt.hinweis(t); antwort.Balance != nil {
		t.Errorf("die Regelantwort trägt Balanceangaben: %+v", antwort.Balance)
	}
}

// **Und wenn das Modell gar nicht antwortet, kommt sein Fehler durch.**
//
// **Ein eigener Lauf und kein vierter Fall nebenan**: die drei dort sind
// Fehler des **Lesens** — das Modell hat geantwortet, nur unbrauchbar. Dieser
// ist einer des **Redens**, und er liegt eine Zeile früher. Die Vorprüfung
// hat ihn als unbetreten gemeldet.
func TestEinStummesModellReichtSeinenFehlerDurch(t *testing.T) {
	welt := einehinweiswelt(t)
	welt.balanceSagt(einVollstaendigerVergleich)
	// **Mit einer Fehlerklasse und nicht als blanker Fehler**: ein
	// unklassifizierter wird zu „Es ist ein unerwarteter Fehler
	// aufgetreten", und dann prüfte der Lauf die Übersetzung statt des Wegs
	// (§8).
	welt.laufzeit.redefehler = fmt.Errorf(
		"%w: das Modell antwortet nicht", domain.ErrValidation)

	antwort := welt.balance(t, "Commander")
	if antwort.Fehler == nil {
		t.Fatal("der Fehler kam nicht durch")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "antwortet nicht") {
		t.Errorf("die Meldung ist %q", antwort.Fehler.Meldung)
	}

	// **Und die Listen sind Listen** — ein Fehler macht aus einer Antwort
	// keinen halben Stand.
	if antwort.Befunde == nil || antwort.Einschraenkungen == nil ||
		antwort.Belege == nil || antwort.Korpusstaende == nil {
		t.Errorf("eine Liste ist nil: %+v", antwort)
	}
}

// **Jeder Fehlerweg des Balancewegs reicht durch.**
func TestJederFehlerwegDerBalanceReichtDurch(t *testing.T) {
	for name, inhalt := range map[string]string{
		"kein JSON":          "kaputt",
		"ohne Einschätzung":  `{"einschaetzung":"  ","vergleiche":[]}`,
		"ohne Vergleichfeld": `{"einschaetzung":"Sie ist stark."}`,
	} {
		t.Run(name, func(t *testing.T) {
			welt := einehinweiswelt(t)
			welt.balanceSagt(inhalt)

			antwort := welt.balance(t, "Commander")

			// **Ohne Vergleichfeld ist es kein Fehler, sondern ein Befund** —
			// das Modell hat geantwortet, nur ohne Vergleich.
			if name == "ohne Vergleichfeld" {
				if antwort.Belegt {
					t.Fatal("die Antwort ohne Vergleich galt als belegt")
				}

				return
			}

			if antwort.Fehler == nil {
				t.Fatalf("der Fehler kam nicht durch: %+v", antwort)
			}
		})
	}
}

// **Und eine abgeschnittene Antwort wird ganz verworfen.**
func TestEineAbgeschnitteneBalanceantwortWirdVerworfen(t *testing.T) {
	welt := einehinweiswelt(t)
	welt.laufzeit.antwort = app.KIRohantwort{
		Inhalt: `{"einschaetzung":"halb`, Abgeschnitten: true,
	}

	if antwort := welt.balance(t, "Commander"); antwort.Fehler == nil {
		t.Fatal("die abgeschnittene Antwort kam durch")
	}
}
