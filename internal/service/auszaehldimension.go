package service

// **Eine erklärte Dimension über ein Set rechnen**
// (#1338,
// Werkstück C von Phase 4.11).
//
// # Was nachgelesen wurde (§15)
//
// `StatsPanel::showCategory` (`src/gui/set/stats_panel.cpp`, Zeilen 477–507,
// Fassung f9d9356 vom 2021-07-06). Vier Sätze, und jeder steht hier als
// Regel:
//
//	String value = untag(dim->script.invoke(ctx)->toString());
//	if (value.empty() && !dim->show_empty) { show = false; break; }
//	catch (ScriptError const& e) { handle_error(…); show = false; break; }
//	if (dim->split_list) d.splitList(dim_id);
//
// # Warum das Zählen nicht hier steht
//
// Weil es keinen Interpreter braucht. Was aus den Werten wird — Reihenfolge,
// Klassen, leere Gruppen — ist eine fachliche Regel über mehrere Karten und
// steht in [domain.Auszaehlen]. Hier steht, **wie ein Wert entsteht**.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Auszaehlangabe ist eine Dimension, wie der Dienst sie braucht.
//
// **Die Gestalt der Domäne und nicht die des Pakets** (§2, §16):
// `msepaket.Auszaehldimension` trägt dazu den Pfad ihres Sinnbilds und ihren
// `position hint:` — beides Angaben über eine **Datei**.
type Auszaehlangabe struct {
	Name         string
	Beschreibung string
	Skript       string

	// Globalskript rechnet **einmal für das ganze Set**, und sein Ergebnis
	// steht jedem Kartenskript als `global_value` zur Verfügung.
	//
	// **Der Schlüssel gehört einem MSE, dessen Quelltext hier nicht vorliegt**
	// (#1355);
	// die Herkunft steht bei [msepaket.Auszaehldimension.Globalskript].
	Globalskript string

	Numerisch     bool
	Klassenbreite float64
	LeereZeigen   bool
	Zerlegen      bool

	Gruppen []string
	Farben  map[string]string
}

// Auszaehlart gibt die Angaben, die das Zählen braucht.
func (a Auszaehlangabe) Auszaehlart() domain.Auszaehlart {
	return domain.Auszaehlart{
		Numerisch:     a.Numerisch,
		Klassenbreite: a.Klassenbreite,
		Gruppen:       a.Gruppen,
		Farben:        a.Farben,
	}
}

// Auszaehlwerte rechnet den Wert der Dimension für jede Karte.
//
// **Eine Karte, ein Wert — außer `split list:` zerlegt ihn.** Was leer bleibt
// und nicht gezeigt werden soll, fällt heraus; was am Skript scheitert,
// ebenso, und dann steht ein Befund daneben.
//
// # Warum die **Vorlage** und nicht bloß die Feldnamen
//
// Weil eine Karte im Bestand nur trägt, was jemand hineingeschrieben hat —
// alles andere rechnet ihr Spiel. Gemessen am 2026-08-31 an
// `White.mse-set`: von 3.760 Karten führen **32** ein `card_color`, die
// übrigen 3.728 bekommen es aus dem `default:` des Feldes. Und dieses
// `default:` beginnt, wie alle 142 Vorgabewerte von `magic.mse-game`, mit
// `stylesheet; ` — ohne den Stil in der Umgebung scheitert es.
//
// **Ohne Vorlage gab „color category" deshalb für 3.758 von 3.760 Karten
// `Colorless`**, wo MSE2 `White` sagt. Die Umgebung, die eine Dimension
// braucht, ist dieselbe, die eine gezeichnete Karte braucht — es gibt nicht
// zwei Wahrheiten über denselben Kartenwert.
//
// # Was hier **nicht** gerechnet wird, und warum
//
// **Die Inhaltsmaße** (`card_style`, `inhaltsmasse`). Sie messen Text mit
// echten Schriftdateien und laden Bilder; eine Dimension liest sie nicht —
// die fünfzig von `magic.mse-game` rufen ausschließlich Funktionen aus
// `statistics_script`. Wer sie mitrechnete, zahlte den Preis einer
// gezeichneten Karte je Zeile einer Auszählung.
//
// # Und warum die Vorlage **je Karte** gewählt wird
//
// Weil eine Karte ihren eigenen Stil tragen darf (`card.stylesheet`), und
// weil davon Werte abhängen: `shape` rechnet `stylesheet; card_shape()`, und
// jeder Stil beantwortet das anders. Gemessen am 2026-08-31 gegen das Orakel:
// mit der Vorlage des **Sets** für jede Karte wichen **21 von 3.761** ab —
// zweiseitige Karten, deren `shape` „adventure" heißt und hier „normal".
func Auszaehlwerte(
	ctx context.Context, karten []*domain.CardRow, vorlage Vorlage,
	fuerKarte Kartenvorlagenwahl, angabe Auszaehlangabe,
) ([]string, []domain.Befund) {
	paare, befunde := AuszaehlwerteKarten(ctx, karten, vorlage, fuerKarte, angabe)

	werte := make([]string, 0, len(paare))
	for _, paar := range paare {
		werte = append(werte, paar.Wert)
	}

	return werte, befunde
}

// AuszaehlwerteKarten rechnet wie [Auszaehlwerte] und nennt zu jedem Wert die
// Karte, aus der er kommt — UXR-F6
// (#2013).
//
// **Derselbe Durchlauf und nicht ein zweiter.** MSE2 hängt jedem Element der
// Auszählung den `original_index` seiner Karte an (`GraphElement`,
// `StatsPanel::showCategory`, `src/gui/set/stats_panel.cpp` Zeilen 477–507,
// Fassung f9d9356 vom 2021-07-06), und `split list:` kopiert ihn mit
// (`GraphDataPre::splitList`, `src/gui/control/graph.cpp` Zeilen 42–60). Eine
// Karte, deren Skript scheitert oder deren Wert leer bleibt, hat keinen Wert
// und steht deshalb in keiner Gruppe.
func AuszaehlwerteKarten(
	ctx context.Context, karten []*domain.CardRow, vorlage Vorlage,
	fuerKarte Kartenvorlagenwahl, angabe Auszaehlangabe,
) ([]domain.Dimensionswert, []domain.Befund) {
	// **Die Umgebung des Spiels steht einmal da**, nicht je Karte — dieselbe
	// Überlegung wie in [zeilenAus]: das `init script:` von `magic.mse-game`
	// hat 583.549 Zeichen, und die fünfzig Dimensionen rufen ausschließlich
	// Funktionen daraus.
	//
	// **Die Setfelder gehören dazu.** `magic.mse-game` greift in Zeile 10583
	// auf `set.custom_super_types` zu; ein fehlendes **Glied** bricht
	// MSE-Script ab, und alles danach fehlt — auch `statistics_script`.
	// Gemessen am 2026-08-31: ohne sie scheiterten **alle fünfzig**
	// Dimensionen mit „der Name `filter_from_statistic` ist hier nicht
	// bekannt".
	// **Ein Vermerk für den ganzen Lauf, geleert vor jeder Karte.**
	//
	// Er hängt an der Umgebung des **Spiels**, weil `text_filter` die
	// Schlüsselwortfunktion beim Initialisieren einfriert (Befund 5 des
	// Zweitblicks über Phase 4.11; `schluesselwortnaht.go`). Karten werden
	// nacheinander gerechnet — ein Vermerk genügt.
	vermerk := &Schluesselwortverwendung{}

	spiel := spielumgebung(ctx, vorlage.Spielskript, vorlage.Setfelder,
		vorlage.Kartenfelder, vorlage.Schluesselwoerter, vermerk)

	// **Und je Stil sein `init script:`** — einmal, nicht je Karte.
	//
	// Der Stil darf eine Funktion des Spiels überschreiben; das ist sein
	// Sinn. `magic.mse-game` erklärt `card_shape := { "normal" }`, und
	// `magic-m15-adventure` schreibt es um. Ohne diesen Schritt gab `shape`
	// für jede Karte „normal", und zweiseitige Karten zählten einfach —
	// gemessen am 2026-08-31: 21 von 3.761 Karten wichen deshalb vom Orakel
	// ab.
	stile := &stilumgebungen{basis: spiel, gemerkt: map[string]*msescript.Umgebung{}}

	// **Erst jede Karte rechnen, dann auszählen.**
	//
	// Zwei Dimensionen von `magic.mse-game` rechnen über das **ganze** Set und
	// geben ihr Ergebnis nur an der ersten Karte aus:
	//
	//	if card != set.cards.0 then "" else (
	//	  set_count := length(set.cards)-1
	//
	// MSE kann das überall: `SetScriptContext::getContext` bindet das ganze
	// Set-Objekt (`src/script/script_manager.cpp` Zeile 39, Fassung `f9d9356`
	// vom 2021-07-06), und `cards` löst auf die Kartenliste auf.
	//
	// **Und `card != set.cards.0` trägt nur, wenn es dieselbe Abbildung ist.**
	// MSE vergleicht `ScriptObject<CardP>` über den **Zeiger**
	// (`to_value.hpp` Zeilen 184–187); hier werden Abbildungen über Namen und
	// Werte verglichen. Eine zweite, aus rohen Feldern gebaute Liste wäre für
	// **keine** Karte gleich — jede gäbe `""`, und die Auszählung bliebe still
	// leer, bei grünem Lauf. Deshalb entsteht jede Kartenabbildung **einmal**
	// und steht danach an beiden Stellen.
	umgebungen := make([]*msescript.Umgebung, 0, len(karten))
	kartenliste := make([]msescript.Eintrag, 0, len(karten))

	// **Und je Karte ihr Schlüsselwortvermerk.**
	//
	// Der Vermerk hängt an der Umgebung des Spiels (Befund 5 des Zweitblicks
	// über Phase 4.11) und hielte am Ende dieser Schleife nur die **letzte**
	// Karte. Genommen wird er deshalb sofort und vor der Auswertung
	// zurückgelegt.
	vermerke := make([][]schluesselwortvermerk, 0, len(karten))

	for _, karte := range karten {
		if ctx.Err() != nil {
			break
		}

		eigene := kartenvorlage(vorlage, fuerKarte, karte)
		fertig, _, _ := auszaehlumgebung(ctx, stile.fuer(ctx, eigene), eigene, karte, vermerk, nil)

		umgebungen = append(umgebungen, fertig)
		vermerke = append(vermerke, vermerk.Stand())

		gerechnet, _ := fertig.Suche("card")
		kartenliste = append(kartenliste, msescript.Eintrag{Wert: gerechnet})
	}

	// **`set.cards` tritt neben die Setfelder, nicht an ihre Stelle.**
	//
	// `spielumgebung` setzt `set` auf die leere Gestalt der Setfelder; Zeile
	// 10583 des Spielskripts liest `set.custom_super_types`, und ein
	// fehlendes Glied bricht MSE-Script ab.
	for _, umgebung := range umgebungen {
		umgebung.Setze("set", mitKartenliste(umgebung, kartenliste))
	}

	// **Das globale Skript rechnet einmal — nach `set.cards` und vor allem
	// anderen.**
	//
	// Es läuft über das ganze Set (`average_opening_hand_global_statistic`
	// beginnt mit `if length(set.cards) < 1`), also braucht es die Liste; und
	// jede Karte liest sein Ergebnis, also muss es vorher dastehen.
	//
	// **Die Umgebung ist die der ersten Karte**, und das ist die eine Zusage,
	// die MSE selbst beantwortet: `SetScriptContext::getContext` bindet `card`
	// auf `set.cards.front()` und nennt es ausdrücklich `// dummy value`
	// (`src/script/script_manager.cpp` Zeile 43, Fassung `f9d9356` vom
	// 2021-07-06).
	if angabe.Globalskript != "" && len(umgebungen) > 0 {
		global, fehler := msescript.AuswertenQuelle(
			ctx, angabe.Globalskript, umgebungen[0])
		if fehler != nil {
			// **Ein gescheitertes globales Skript nimmt die Dimension mit.**
			//
			// Es ist die Voraussetzung jeder Karte; weiterzurechnen hieße,
			// 3.761-mal denselben Folgefehler zu melden. Der Befund nennt das
			// globale Skript, damit niemand die Ursache an der falschen Stelle
			// sucht.
			return nil, []domain.Befund{{
				Ort: angabe.Name,
				Meldung: fmt.Sprintf(
					"das globale Skript der Dimension %q ließ sich nicht rechnen: %v",
					angabe.Name, fehler),
			}}
		}

		for _, umgebung := range umgebungen {
			umgebung.Setze("global_value", global)
		}
	}

	werte := make([]domain.Dimensionswert, 0, len(karten))
	gescheitert := 0

	var ersterFehler error

	// **Über die Umgebungen und nicht über die Karten.**
	//
	// Sie sind genau die Karten, die die erste Schleife geschafft hat; ein
	// Abbruch dazwischen lässt sie kürzer werden. Über die Karten zu laufen
	// verlangte eine Schranke gegen die Länge — einen Zweig, den nur ein
	// Abbruch zwischen zwei Schleifen betritt, und der wäre ungeprüft (§3).
	// **Und ohne eigene Abbruchprüfung.**
	//
	// Ein Abbruch in der ersten Schleife lässt `umgebungen` kürzer werden;
	// was übrig ist, ist gebaut und kostet nichts mehr. Und wer **hier**
	// abbricht, wird von `AuswertenQuelle` gehört — der Auswerter prüft
	// `ctx.Err()` selbst (`auswerter.go`, Zeile 167).
	//
	// Eine zweite Prüfung wäre ein Zweig, den nur ein Abbruch **zwischen**
	// den beiden Schleifen betritt: nicht erreichbar, ohne die Zeit
	// anzuhalten, und damit ungeprüft (§3).
	for stelle, umgebung := range umgebungen {
		vermerk.Setze(vermerke[stelle])

		wert, fehler := msescript.AuswertenQuelle(ctx, angabe.Skript, umgebung)
		if fehler != nil {
			gescheitert++

			if ersterFehler == nil {
				ersterFehler = fehler
			}

			continue
		}

		roh := domain.OhneAuszeichnung(wert.AlsText())

		if roh == "" && !angabe.LeereZeigen {
			continue
		}

		// **`karten[stelle]` und `umgebungen[stelle]` sind dieselbe Karte**:
		// die erste Schleife hängt je Karte genau eine Umgebung an und bricht
		// nur ab, nie nach der Mitte einer Karte.
		kennung := karten[stelle].ID

		if angabe.Zerlegen {
			for _, glied := range zerlegteListe(roh) {
				werte = append(werte, domain.Dimensionswert{Karte: kennung, Wert: glied})
			}

			continue
		}

		werte = append(werte, domain.Dimensionswert{Karte: kennung, Wert: roh})
	}

	return werte, auszaehlbefunde(angabe.Name, gescheitert, ersterFehler)
}

// stilumgebungen hält je Stilpaket eine Umgebung mit seinem `init script:`.
//
// **Gemerkt und nicht je Karte gebaut.** `magic-m15-altered/style` hat allein
// 700 Zeilen; ein Set aus 3.761 Karten benutzt eine Handvoll Stile.
type stilumgebungen struct {
	basis   *msescript.Umgebung
	gemerkt map[string]*msescript.Umgebung
}

// fuer gibt die Umgebung zum Stil dieser Vorlage.
func (s *stilumgebungen) fuer(
	ctx context.Context, vorlage Vorlage,
) *msescript.Umgebung {
	if vorlage.Stil.Initskript == "" {
		return s.basis
	}

	if fertig, da := s.gemerkt[vorlage.Stil.Paket]; da {
		return fertig
	}

	eigene := s.basis.Darueber()

	// **Der Fehler fällt weg, und das ist benannt**: ein Stil, dessen
	// Initskript nicht durchläuft, gibt die Namen, die bis dahin entstanden
	// sind — dieselbe Duldung wie bei [spielumgebung] (§8). Was daraus folgt,
	// meldet die Dimension als Befund, mit der Zahl der Karten.
	_ = vorbelegt(ctx, vorlage.Stil.Initskript, "init script", eigene)

	s.gemerkt[vorlage.Stil.Paket] = eigene

	return eigene
}

// Kartenvorlagenwahl gibt die Vorlage einer einzelnen Karte.
//
// **Nil ist erlaubt und heißt: die des Sets.** Ein Set ohne eigene
// Kartenstile — und jedes Doppel in den Tests — kommt damit aus.
type Kartenvorlagenwahl func(*domain.CardRow) Vorlage

// kartenvorlage wählt die Vorlage einer Karte.
func kartenvorlage(
	vorlage Vorlage, fuerKarte Kartenvorlagenwahl, karte *domain.CardRow,
) Vorlage {
	if fuerKarte == nil {
		return vorlage
	}

	return fuerKarte(karte)
}
