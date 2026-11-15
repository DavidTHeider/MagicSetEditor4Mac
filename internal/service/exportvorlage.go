package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// Vorlagenausfuehrer führt eine Exportvorlage aus (#51).
//
// # Was eine Exportvorlage ist
//
// Ein Programm. Der Wert ihres **letzten Ausdrucks** ist die Ausgabe — so
// schreiben es alle 19 Vorlagen von `M15-Magic-Pack`. Was sie sieht: `cards`
// als Liste, `card` je Durchgang, `options` als Abbildung ihrer
// Wahlmöglichkeiten.
//
// # Wo die Dateigrenze heute liegt
//
// `msescript.Bibliothek()` führt heute Rechnen, Text und Listen; ihre heutigen
// Implementierungen bieten keine bekannte Funktion für das Dateisystem. Eine
// Vorlage gibt **Text** zurück, und wohin er geschrieben wird, entscheidet
// allein der Aufrufer.
//
// Der Namenstest daneben meldet nur eine auffällige Erweiterung der Bibliothek.
// Er ist keine Sicherheitsgrenze: eine unauffällig benannte oder nachträglich
// geänderte Funktion könnte er nicht erkennen.
type Vorlagenausfuehrer struct {
	// symbolschriften und symbolrasterer sind gesetzt, wenn der Aufrufer
	// `symbols_to_html` erlaubt hat — siehe `exportsymbole.go`.
	//
	// **Sie stehen hier und nicht in [Ausfuhrmittel]**, weil sie an den
	// Paketwurzeln hängen und nicht an einem Ziel: die Anwendung richtet sie
	// einmal ein. Was je Ausfuhr wechselt, kommt als Wert herein (#736).
	symbolschriften Symbolschriftquelle
	symbolrasterer  Symbolrasterer
}

// NeuerVorlagenausfuehrer baut den Ausführer.
func NeuerVorlagenausfuehrer() *Vorlagenausfuehrer { return &Vorlagenausfuehrer{} }

// NamenDerBibliothek gibt die Namen, die ein Vorlagenskript vorfindet.
//
// **Für den Hinweis auf auffällige Bibliotheksnamen**, nicht für die Auswertung
// und nicht als Sicherheitsgrenze.
func NamenDerBibliothek() []string { return msescript.Bibliothek().Namen() }

// Fuehre führt eine Vorlage über einer Kartenliste aus.
//
// `wahl` überschreibt die Vorgaben der Wahlmöglichkeiten; ein Name, den die
// Vorlage nicht führt, wird **benannt** und nicht stillschweigend übergangen —
// er ist fast immer ein Tippfehler.
//
// **Die Ausgabedatei entsteht ganz oder gar nicht.** Anders als bei einer Karte
// (§8): dort ist eine Ebene weniger brauchbar, hier wäre eine halbe Datei eine,
// die jemand für vollständig hält.
//
// **Für die Beigaben gilt das nicht** —
// #750.
// `copy_file`, `write_text_file`, `write_set_file` und die Bilder schreiben
// **während** der Auswertung; scheitert ein späterer Ausdruck, bleiben sie
// liegen.
//
// **Aufgeräumt wird nicht, gemeldet schon.** MSE schreibt ebenso inkrementell
// (`src/script/functions/export.cpp`, Zeile 368; die Hauptdatei entsteht erst
// danach, `src/gui/html_export_window.cpp`, Zeile 73, Fassung f9d9356 vom
// 2021-07-06), und §15.4 verbietet die stillschweigende Änderung. Was der
// Fehler sagen kann, ist, **wo** etwas liegt.
func (v *Vorlagenausfuehrer) Fuehre(
	ctx context.Context, mittel Ausfuhrmittel, vorlage msepaket.Exportvorlage,
	karten []*domain.CardRow, setwerte []domain.Value, wahl map[string]string,
) (string, []domain.Befund, error) {
	if strings.TrimSpace(vorlage.Skript) == "" {
		return "", nil, fmt.Errorf("%w: die Vorlage %q führt kein Skript",
			domain.ErrValidation, vorlage.KurzerName)
	}

	umgebung := msescript.Bibliothek().Darueber()

	befunde := optionenSetzen(umgebung, vorlage, wahl)

	// **Einmal je Lauf gemerkt**, für alle drei Rufer — E04 (#2241).
	stile := mittel.Stile.gemerkt()

	gerechnet, rechenbefunde := ausfuhrkarten(ctx, vorlage, karten, setwerte, stile)
	befunde = append(befunde, rechenbefunde...)

	liste, stilbefunde := mitStilen(
		mitGerechneten(kartenliste(karten, vorlage.Kartenfelder), gerechnet), stile)
	befunde = append(befunde, stilbefunde...)
	umgebung.Setze("cards", liste)

	// **`set` trägt jedes Feld, das das Spiel erklärt** (#609) — dieselbe
	// Regel wie an der Karte, eine Ebene höher:
	//
	//	Set::Set(const GameP& game) … { data.init(game->set_fields); }
	//
	// `src/data/set.cpp`, Zeilen 33–39, Fassung f9d9356 vom 2021-07-06.
	//
	// **Und eine leere Abbildung, wenn niemand etwas mitgibt** — nicht ein
	// fehlender Name: eine Vorlage, die `set.title` liest, bekommt so einen
	// Fehler über das **Glied** und keinen über `set`, und der Unterschied ist
	// die Meldung, die jemand liest.
	set, stilbefunde := mitSetstil(setabbildung(setwerte, vorlage.Setfelder, liste), stile)
	befunde = append(befunde, stilbefunde...)
	umgebung.Setze("set", set)

	// **`create directory:` entscheidet, ob die Vorlage überhaupt schreiben
	// darf** (#732).
	//
	//	} else if (export_info()->directory_relative.empty()) {
	//	  throw ScriptError(_("Can only use ") + fun +
	//	                    _(" when 'create directory' is set to true"));
	//	}
	//
	// `guard_export_info`, `src/script/functions/export.cpp`, Zeilen 26–33,
	// Fassung f9d9356 vom 2021-07-06. **Jede** schreibende Funktion hängt
	// daran; über die 19 Vorlagen des M15-Packs sagen zehn `true`, vier
	// `false` und fünf nichts — und nichts heißt `false`.
	//
	// **Die Angabe der Vorlage und der Name des Aufrufers müssen beide da
	// sein.** Die eine sagt, dass ein Verzeichnis gemeint ist; der andere,
	// wie es heißt. Fehlt eines, gibt es die Namen nicht.
	darfSchreiben := vorlage.Verzeichnis && mittel.Verzeichnis != ""

	if darfSchreiben {
		umgebung.Setze("directory", msescript.Text(mittel.Verzeichnis))
	} else {
		// **Die Namen stehen da und scheitern beim Aufruf** — so wie MSE.
		// „Der Name `copy_file` ist hier nicht bekannt" wäre irreführend: er
		// ist bekannt, die Erlaubnis fehlt. Der Unterschied ist die Stunde,
		// die jemand mit der Suche nach einer fehlenden Bibliothek verbringt.
		verweigern(umgebung, "copy_file", "write_text_file", "write_set_file",
			"write_image_file", "symbols_to_html")

		// **Und `to_html` erreicht die Wache indirekt.**
		//
		//	} else if (is_substr(str, i, _("/sym"))) {
		//	  if (!symbols.empty()) { … ret += symbols_to_html(…); }
		//
		// `export.cpp`, Zeilen 230–236 und 262–267: `to_html` mit einer
		// Symbolschrift ruft `symbols_to_html`, und dort sitzt
		// `guard_export_info`. Ohne diesen Rückruf bliebe die
		// Bibliotheksfassung stehen und meldete „steht außerhalb einer
		// Ausfuhr nicht zur Verfügung" — richtig verhindert, falsch benannt.
		// Angestrichen vom Zweitblick durch Codex am 2026-08-21.
		umgebung.Setze("to_html", msescript.NeueEingebaute(msescript.HTMLFunktion(
			func(_ context.Context, _, _ string, _ float64) (string, error) {
				return "", fmt.Errorf("%w: symbols_to_html geht nur, wenn die "+
					"Vorlage `create directory: true` sagt", domain.ErrValidation)
			})))
	}

	// **Der Lauf trägt die Merkliste**, und beide Schreiber teilen sie sich —
	// bei MSE ist es `ei.exported_images`, und die lebt je Ausfuhr
	// (#658, #660).
	lauf := &ausfuhrlauf{
		ausfuehrer: v, mittel: mittel, umgebung: umgebung,
		gelegt: map[string][2]int{},
	}

	// **Nur wenn der Aufrufer es erlaubt hat** (#216). Ohne Bildschreiber gibt
	// es den Namen nicht, und die Vorlage scheitert mit „ist hier nicht
	// bekannt" — die ehrlichere Antwort als eine Datei an einem Ort, den
	// niemand gewählt hat.
	//
	// **Oder wenn die Vorlage Bildwerte schreiben darf** (#660): dieselbe
	// Funktion, zweite Gestalt. Wer nur eine der beiden Fähigkeiten hat,
	// bekommt für die andere einen benannten Fehler statt eines Absturzes.
	if darfSchreiben && (mittel.Bilder != nil || mittel.Symbolzeichnungen != nil) {
		umgebung.Setze("write_image_file",
			msescript.NeueEingebaute(lauf.bildschreiber()))
	}

	// **Dieselbe Regel für die drei Geschwister** (#657): ohne die Fähigkeit
	// gibt es die Namen nicht.
	if darfSchreiben && mittel.Beigaben != nil {
		umgebung.Setze("copy_file", msescript.NeueEingebaute(lauf.dateikopierer()))
		umgebung.Setze("write_text_file", msescript.NeueEingebaute(lauf.textschreiber()))
		umgebung.Setze("write_set_file", msescript.NeueEingebaute(lauf.setleger()))
	}

	// **Symbolbilder brauchen beides** (#658): eine Schrift, aus der sie
	// kommen, und einen Ort, an den sie gehen. Die Merkliste entsteht **hier**
	// und nicht am Ausführer — sie gehört zu diesem Lauf und zu seinem
	// Verzeichnis.
	if darfSchreiben && v.symbolschriften != nil && v.symbolrasterer != nil &&
		mittel.Beigaben != nil {
		umgebung.Setze("symbols_to_html", msescript.NeueEingebaute(lauf.symbolfunktion()))

		// **Und `to_html` bekommt denselben Weg** (#723): eine zweite
		// Umsetzung daneben liefe auseinander, sobald eine von beiden etwas
		// lernt (§10).
		umgebung.Setze("to_html", msescript.NeueEingebaute(
			msescript.HTMLFunktion(lauf.wandler)))
	}

	// **Und das Setsymbol** (#660): `symbol_variation` gibt es beim Zeichnen
	// längst — dort steht ein Stil daneben. Hier kommen die Fassungen mit der
	// Vorlage, gelesen aus dem Spielpaket.
	if mittel.Beigaben != nil {
		lauf.setsymbolSetzen(umgebung, vorlage.Symbolfassungen)
	}

	// **Das Spielskript, nachdem die Umgebung steht** (#216).
	//
	// Eine Exportvorlage ruft die Funktionen ihres Spiels: `cmc`, `card_name`,
	// `booster_sort`, `symbol_variation` stehen in `magic.mse-game` und nicht
	// in der Bibliothek. Fünf der neunzehn Vorlagen von `M15-Magic-Pack`
	// scheitern ohne sie an einem unbekannten Namen.
	//
	// **Nach `cards` und `set` und nicht davor**, und das ist gemessen: das
	// Initskript des Spiels liest sie beim Vorbelegen. Vorangestellt bricht es
	// ab, und dann fehlen genau die Funktionen, um die es ging.
	//
	// Ein Fehler darin ist ein **Befund**, kein Abbruch: eine Vorlage, die die
	// Funktionen des Spiels gar nicht ruft, läuft ohne sie durch.
	befunde = append(befunde,
		vorbelegt(ctx, vorlage.Spielskript, "init script des Spiels", umgebung)...)

	wert, err := msescript.AuswertenQuelle(ctx, vorlage.Skript, umgebung)
	if err != nil {
		return "", append(befunde, lauf.befunde...),
			fmt.Errorf("die Vorlage %q ließ sich nicht ausführen: %w%s",
				vorlage.KurzerName, err, lauf.halberStand(mittel.Verzeichnis))
	}

	// **Was beim Rastern angefallen ist, reist mit.** Ein Symbolbild, dessen
	// Verknüpfung noch nicht gebaut ist, wird geschrieben — und gemeldet.
	return wert.AlsText(), append(befunde, lauf.befunde...), nil
}

// optionenSetzen legt `options` in die Umgebung.
func optionenSetzen(
	umgebung *msescript.Umgebung, vorlage msepaket.Exportvorlage, wahl map[string]string,
) []domain.Befund {
	gefuehrt := make(map[string]bool, len(vorlage.Optionen))
	eintraege := make([]msescript.Eintrag, 0, len(vorlage.Optionen))

	for _, option := range vorlage.Optionen {
		gefuehrt[option.Name] = true

		wert := option.Anfangswert()
		if gewaehlt, da := wahl[option.Name]; da {
			wert = gewaehlt
		}

		// **Leerzeichen werden zu Unterstrichen** — `options.nur_namen` für
		// `name: nur namen`. Dieselbe Regel wie bei den Kartenfeldern
		// (`docs/FORMATE.md`).
		eintraege = append(eintraege, msescript.Eintrag{
			Name: strings.ReplaceAll(option.Name, " ", "_"),
			Wert: optionswert(option.Art, wert),
		})
	}

	umgebung.Setze("options",
		msescript.Wert{Art: msescript.ArtAbbildung, Eintraege: eintraege})

	var befunde []domain.Befund

	for name := range wahl {
		if !gefuehrt[name] {
			befunde = append(befunde, domain.Befund{
				Ort:     name,
				Meldung: "die Vorlage führt keine Wahlmöglichkeit dieses Namens",
			})
		}
	}

	return befunde
}

// kartenliste macht aus den Karten eine Liste von Abbildungen.
//
