//go:build fremdpakete

// Jeder **echte** Stil einmal durch den Planbauer.
//
// # Warum das hinter einem Build-Tag steht
//
// `testdaten/` enthält keine fremden Vorlagen (§11,
// [0005](../../docs/entscheidungen/0005-keine-fremden-vorlagen.md)). Dieser Lauf
// schließt die Lücke, ohne sie ins Repository zu holen. Er läuft nicht in der
// Pipeline, zählt nicht zur Abdeckung und wird von Hand angestoßen:
//
//	MSE_FREMDPAKETE="$HOME/Library/Application Support/MagicSetEditor4Mac/pakete/M15-Magic-Pack/data" \
//	  go test -tags fremdpakete ./internal/service/ -run Fremdpakete -v
//
// # Was er prüft, und was nicht
//
// **Nicht**, ob die Karte richtig aussieht — dafür bräuchte es die Karten, und
// die liegen hier nicht. Geprüft wird die Zusage, die
// #173 macht:
// jedes Feld wird **gedeutet oder benannt**, eine dritte Antwort gibt es nicht.
//
// Und eine zweite, die schwerer wiegt: **jeder Bildpfad, der in einen Plan
// gerät, ist ein gültiger Paketpfad.** Ein Pfad, den `msepaket.Bildquelle`
// nicht zerlegen kann, wäre eine Ebene, die nie etwas zeichnet — und die
// Meldung dazu käme erst beim Rastern, mit einem Wortlaut über Dateien statt
// über die Angabe im Stil.
//
// # Was er gezählt hat, und was das heißt
//
// | Stand | Feldstile | Bildebenen | Textebenen |
// | --- | ---: | ---: | ---: |
// | vor #186 | 2016 | **10** | 1063 |
// | nach #186 | 5950 | **113** | 3119 |
//
// Zehn Bildebenen aus 356 `image:`-Angaben waren keine Deutungslücke, sondern
// eine **Umgebungslücke**: die Skripte rufen `module_card()` und
// ihresgleichen, und die stehen im `init script:` des Stils und in seinen
// `include file:`-Modulen (#186).
// Dass die Feldstile sich fast verdreifachten, hat dieselbe Ursache: die
// Module bringen eigene `card style:`-Blöcke mit.
//
// **Was danach bleibt, kommt aus dem Spielpaket** — `magic.mse-game/script`
// führt Funktionen, die jeder Stil benutzt, und diese Schicht fehlt weiter
// (#195).
//
// **Die Zahl steht hier, damit der nächste sieht, ob sie sich bewegt hat.**
package service_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// TestDerPlanbauerUeberFremdpakete baut aus jedem bezogenen Stil einen Plan.
func TestDerPlanbauerUeberFremdpakete(t *testing.T) {
	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	eintraege, err := os.ReadDir(wurzel)
	if err != nil {
		t.Fatalf("Wurzel lesen: %v", err)
	}

	var (
		stile, felder, bildebenen, textebenen, symbolzuege, farbflaechen, befunde, halbe int
		gruende                                                                          = map[string]int{}
	)

	for _, eintrag := range eintraege {
		if !strings.HasSuffix(eintrag.Name(), ".mse-style") {
			continue
		}

		stil, offen := stilAus(t, filepath.Join(wurzel, eintrag.Name()), wurzel)
		if !offen || stil.Breite <= 0 || stil.Hoehe <= 0 {
			continue
		}

		stile++
		felder += len(stil.Felder)

		// **Mit einer Bildquelle, so wie der Betrieb sie hat** (§3): ohne sie
		// misst die erste Stufe des Satzes einen Symbolzug als Reihe von
		// Quadraten, und der Zähler unten spräche über einen Weg, den niemand
		// geht.
		//
		// **Ohne Schriftsuche**, und das ist keine Nachlässigkeit: `service`
		// darf `internal/schriften` nicht kennen (§2), und eine Attrappe
		// daneben wäre großzügiger als der Betrieb. Was diesem Lauf dadurch
		// entgeht, ist die Textmessung für `content_width` — sie steht in
		// `internal/app/mse2goldbild_test.go`, wo die Schicht sie kennen darf.
		vorlage := vorlageFuer(t, stil, wurzel)

		plan, stilbefunde, fehler := service.NeuerPlanbauer().
			MitBildern(msepaket.NeueBildquelle(wurzel)).
			Stilplan(t.Context(), stilprobekarte(stil, vorlage.Kartenfelder), vorlage)
		if fehler != nil {
			t.Errorf("%s: %v", eintrag.Name(), fehler)

			continue
		}

		for _, ebene := range plan.Ebenen {
			switch ebene.Art {
			case render.ArtBild:
				bildebenen++
			case render.ArtText:
				textebenen++
			case render.ArtSymbolzug:
				symbolzuege++
			case render.ArtFlaeche:
				// **Der weiße Kartengrund gehört nicht dazu.** Er ist unsere
				// Zutat und keine Ebene der Vorlage; mitgezählt stünde in
				// dieser Spalte für jeden Stil eine Eins zu viel.
				if !grundflaeche(ebene, plan) {
					farbflaechen++
				}
			}

			// **Die eigentliche Zusage dieses Laufs — über alle Blätter.**
			//
			// Bis zum 2026-08-16 stand hier `ebene.Bild.Pfad`: der Pfad am
			// **Wurzelknoten**. Bei einem Bildbaum ist der leer, und die Zusage
			// prüfte nichts. Angestrichen vom Zweitblick über die Phasen B bis
			// D.
			verweise := append(ebene.Bild.Blaetter(), ebene.Maske.Blaetter()...)
			for _, verweis := range verweise {
				if verweis == "" {
					continue
				}

				if _, err := msepaket.ZerlegePaketpfad(verweis); err != nil {
					t.Errorf("%s: %v", eintrag.Name(), err)
				}
			}

			// **Und was nur zur Hälfte gezeichnet wird, wird gezählt.**
			//
			// Eine nicht gebaute Verknüpfung zeichnet ihr Ausgangsbild und
			// meldet das beim **Rastern** — dieser Lauf rastert nicht, und
			// damit verschwänden die 965 `recolor_image` aus dem Zähler,
			// obwohl sich an der Karte nichts geändert hat. Eine Messgröße darf
			// ihre Bedeutung nicht stillschweigend wechseln.
			// **Bild und Maske.** Der erste Anlauf zählte nur das Bild — und
			// das Pack führt sehr wohl Maskenbäume, etwa `crop(…)` in
			// `magic-m15-showcase-eternities-stellar-sights-poster`.
			// Angestrichen vom Nachlauf des Zweitblicks.
			for _, baum := range []render.Bildbaum{ebene.Bild, ebene.Maske} {
				if baum.Leer() || baum.Gebaut() {
					continue
				}

				halbe++
				gruende["halb: "+baum.Knotenart]++
			}
		}

		befunde += len(stilbefunde)

		for _, befund := range stilbefunde {
			if befund.Ort == "" || befund.Meldung == "" {
				t.Errorf("%s: ein Befund ohne Ort oder Meldung: %+v", eintrag.Name(), befund)
			}

			angabe := "sonst"
			if _, teil, geteilt := strings.Cut(befund.Ort, " → "); geteilt {
				angabe = teil
			}

			kurz := kurzgrund(befund.Meldung)
			if teil, _, _ := strings.Cut(befund.Meldung, ": Zeile "); strings.Contains(befund.Meldung, "nicht ausrechnen") {
				_ = teil
				nach := befund.Meldung
				if i := strings.LastIndex(nach, ": "); i > 0 {
					nach = nach[i+2:]
				}
				kurz = "ausrechnen → " + nach
			}

			gruende[angabe+": "+kurz]++
		}
	}

	// **Die Symbolzüge stehen daneben, nicht in den Textebenen.**
	//
	// Ein Feld mit `always symbol:` wird ein Symbolzug und keine Textebene;
	// als die Symbolschriften in diesen Lauf kamen, fiel die Zahl der
	// Textebenen um 146, ohne dass etwas verloren ging. Eine Messgröße, die
	// ihre Bedeutung wechselt, ist keine (§10).
	t.Logf("%d Stile, %d Feldstile → %d Bildebenen, %d Textebenen, %d Symbolzüge, "+
		"%d Farbflächen, %d Befunde",
		stile, felder, bildebenen, textebenen, symbolzuege, farbflaechen, befunde)

	// **Getrennt gezählt und nicht dazugerechnet.** Eine halb gezeichnete
	// Ebene ist etwas anderes als eine fehlende; sie in denselben Topf zu
	// werfen machte beide Zahlen unlesbar.
	t.Logf("davon %d Bildebenen mit einer Verknüpfung, die noch nicht gebaut ist", halbe)

	for grund, anzahl := range gruende {
		t.Logf("  %5d × %s", anzahl, grund)
	}

	if stile == 0 {
		t.Fatal("kein einziger Stil gelesen — steht MSE_FREMDPAKETE richtig?")
	}
}

// stilAus öffnet ein Stilpaket; ein unlesbares ist hier kein Fehlschlag.
//
// **Mit eingebundenen Modulen** (#186): ohne sie trägt `Initskript` nur die
// `include file:`-Zeilen, und jedes Feldskript, das eine Modulfunktion ruft,
// bleibt unauswertbar.
func stilAus(t *testing.T, pfad, wurzel string) (msepaket.Stil, bool) {
	t.Helper()

	paket, err := msepaket.Oeffne(pfad)
	if err != nil {
		return msepaket.Stil{}, false
	}

	paket.ModuleEinbinden(msepaket.NeuerPaketleser(wurzel))

	defer func() {
		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	}()

	stil, _ := paket.Stilmodell()

	return stil, true
}

// probekarte gibt eine Karte, die für **jedes** Feld des Stils einen Wert hat.
//
// **Sonst prüfte der Lauf die Hälfte nicht**: ein Feld ohne Wert wird gar nicht
// erst gedeutet, und ein Stil mit vierzig Feldern käme mit zwei Ebenen durch.
func stilprobekarte(stil msepaket.Stil, kartenfelder []domain.Field) *domain.CardRow {
	art := make(map[string]domain.FieldKind, len(kartenfelder))
	for stelle := range kartenfelder {
		art[kartenfelder[stelle].Name] = kartenfelder[stelle].Kind
	}

	werte := make([]domain.Value, 0, len(stil.Felder))

	for stelle := range stil.Felder {
		name := stil.Felder[stelle].Feld

		// **Nur Textfelder bekommen Text.**
		//
		// Bis zum 2026-08-17 stand hier „Probe" für **jedes** Feld, auch für
		// die Auswahlfelder. Das ist die Vorrichtung, die vom Betrieb abweicht
		// (§3): `rarity` trägt in MSE `common`, nicht `Probe`, und ein Stil
		// hat für `Probe` kein Wahlbild. Rund **460** Befunde dieses Laufs
		// waren nichts als das — dazu jedes `to_int` auf eine Kartennummer.
		//
		// **Ohne Wert bekommt das Feld seine Vorgabe** aus `initial:` oder der
		// ersten Wahl, und das ist genau die Karte, die MSE beim Anlegen zeigt.
		if art[name] != domain.KindText {
			continue
		}

		werte = append(werte, domain.Value{Field: name, Text: "Probe"})
	}

	return &domain.CardRow{ID: 1, Number: 1, Values: werte}
}

// kurzgrund fasst eine Befundmeldung auf ihre Klasse zusammen.
func kurzgrund(meldung string) string {
	for _, klasse := range []string{
		"Symbolschriften sind noch nicht gedeutet",
		"Wahlbilder des Spiels",
		"weder eine Schrift noch ein Bild",
		"ein eigenes Bild",
		"ließ sich nicht ausrechnen",
		"ergab keine Zahl",
		"endet nicht darauf",
		"ließ sich nicht lesen",
	} {
		if strings.Contains(meldung, klasse) {
			return klasse
		}
	}

	return meldung
}

// spielskriptFuer liest das `init script:` eines Spielpakets — einmal je Spiel.
//
// **Gemerkt, weil es 190-mal dasselbe wäre.** `magic.mse-game` bindet über
// dreißig Module ein; es je Stil neu zu lesen ließe den Lauf spürbar warten.
var geleseneSpiele = map[string]gelesenesSpiel{}

func spielskriptFuer(t *testing.T, spiel, wurzel string) gelesenesSpiel {
	t.Helper()

	if spiel == "" {
		return gelesenesSpiel{}
	}

	if gemerkt, da := geleseneSpiele[spiel]; da {
		return gemerkt
	}

	var gelesen gelesenesSpiel

	if paket, err := msepaket.Oeffne(filepath.Join(wurzel, spiel+".mse-game")); err == nil {
		paket.ModuleEinbinden(msepaket.NeuerPaketleser(wurzel))
		gelesen.skript = paket.Initskript()
		gelesen.modell, _ = paket.Spielmodell()
		gelesen.fassungen = paket.Spielsymbolfassungen()

		if fehler := paket.Schliessen(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	}

	geleseneSpiele[spiel] = gelesen

	return gelesen
}

// gelesenesSpiel ist, was diese Messung aus einem Spielpaket braucht.
type gelesenesSpiel struct {
	skript    string
	modell    domain.Game
	fassungen []msepaket.Symbolfassung
}

// vorlageFuer baut die Vorlage **so, wie der Betrieb sie baut**.
//
// # Warum das hier steht und nicht drei Felder weniger
//
// Bis zum 2026-08-16 reichte diese Messung nur Stil und Spielskript weiter —
// keine Kartenfelder, keine Setfelder. Damit maß sie eine Umgebung, die es im
// Betrieb nicht gibt: jede Vorgabe eines Kartenfelds fehlte, und `set` war
// leer. Die gemeldeten Befunde waren deshalb **mehr**, als ein Nutzender sähe.
//
// Es ist die Klasse aus §3 „eine Vorrichtung, die großzügiger ist als der
// Betrieb", nur in die andere Richtung: eine, die strenger ist, meldet Befunde,
// die niemand hat — und ihre Zahl taugt dann nicht als Maß.
func vorlageFuer(t *testing.T, stil msepaket.Stil, wurzel string) service.Vorlage {
	t.Helper()

	spiel := spielskriptFuer(t, stil.Spiel, wurzel)

	// **Und die Symbolschriften, die der Stil nennt** — so wie
	// `Stilspeicher.Symbolschriften` sie im Betrieb lädt.
	//
	// Ohne sie meldete dieser Lauf 21-mal „die Symbolschrift magic-mana-large
	// liegt nicht vor", und das Paket **liegt** vor: eine Vorrichtung, die
	// ärmer ist als der Betrieb, zählt Mängel, die niemand hat (§3).
	schriften, _ := service.NeuerStilspeicher(nil, wurzel).Symbolschriften(stil)

	return service.Vorlage{
		Stil:            stil,
		Spielskript:     spiel.skript,
		Kartenfelder:    spiel.modell.Fields,
		Setfelder:       spiel.modell.SetFields,
		Symbolfassungen: spiel.fassungen,
		Symbolschriften: schriften,
	}
}

// grundflaeche sagt, ob diese Ebene der weiße Kartengrund des Planbauers ist.
//
// **Nach Gestalt und nicht nach Stelle.** Eine Prüfung auf „die erste Ebene"
// wäre kürzer und ginge schief, sobald der Planbauer seine Reihenfolge ändert.
func grundflaeche(ebene render.Ebene, plan *render.Plan) bool {
	return ebene.Farbe == (render.Farbe{R: 255, G: 255, B: 255, A: 255}) &&
		ebene.Rahmen == (render.Rechteck{Breite: plan.Breite, Hoehe: plan.Hoehe}) &&
		ebene.Maske.Leer()
}
