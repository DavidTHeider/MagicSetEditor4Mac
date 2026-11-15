// **Der Symboleditor an der Kommandozeile**
// (#1391).
//
// # Wie ein Symbol durch die Befehle reist
//
// **Als JSON auf der Standardeingabe und -ausgabe.** Ein Symbol hat
// verschachtelte Teile mit Bézierpunkten; als Argumentliste wäre es
// unlesbar, und als Datei brauchte jeder Schritt einen Namen. So lässt sich
// eine Kette schreiben:
//
//	mse symbol-grundform Rechteck \
//	  | mse symbol-rechnen lesen.symbol.symmetrie --stelle 0 --weise anlegen \
//	  | mse symbol-setzen 1
//
// # Warum ein Befehl für alle Rechnungen
//
// Weil sie **eine** Fähigkeit sind: aus Teilen werden Teile. Acht Befehle
// nebeneinander hätten acht Argumentlisten, und die Hilfe wäre eine Seite
// lang für etwas, das ein Name und ein JSON-Rumpf ist.

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// symbolRechnenBefehl führt eine der Rechnungen aus.
//
// **Die Argumente kommen als `--name wert`**, die Teile über die
// Standardeingabe. Wer nichts hineinschiebt, bekommt eine Meldung, die es
// sagt — und keine leere Rechnung.
func symbolRechnenBefehl(an *schreiber, rest []string) int {
	if len(rest) < 1 {
		an.klage("symbol-rechnen braucht den Namen der Rechnung.\n%s\n%s",
			rechnungsnamen(), hilfe)

		return CodeBedienung
	}

	argumente, gut := symbolargumente(an, rest[1:])
	if !gut {
		return CodeBedienung
	}

	ergebnis, fehler := app.SymbolRechnen(rest[0], argumente)
	if fehler != nil {
		an.klageUeber(fehler)

		return CodeFachlich
	}

	return alsJSONSchreiben(an, ergebnis)
}

// symbolargumente baut den JSON-Rumpf aus den Schaltern und der Eingabe.
//
// **Die Teile aus der Eingabe stehen im selben Verbund wie die Schalter.**
// Wer sie getrennt hielte, hätte zwei Wege in dieselbe Rechnung — und der
// zweite ginge irgendwann an einem Feld vorbei.
func symbolargumente(an *schreiber, schalter []string) (json.RawMessage, bool) {
	werte, gut := ausEingabe(an)
	if !gut {
		return nil, false
	}

	for stelle := 0; stelle+1 < len(schalter); stelle += 2 {
		name, roh := strings.TrimPrefix(schalter[stelle], "--"), schalter[stelle+1]

		// **Als JSON gelesen, sonst als Text.** `--stelle 0` soll eine Zahl
		// sein und `--weise anlegen` eine Zeichenkette; ein Schalter, der
		// beides raten müsste, riete irgendwann falsch.
		var wert any
		if err := json.Unmarshal([]byte(roh), &wert); err != nil {
			wert = roh
		}

		werte[name] = wert
	}

	rumpf, err := json.Marshal(werte)
	if err != nil {
		an.klageUeber(fmt.Errorf("%w: die Argumente lassen sich nicht bauen: %w",
			domain.ErrValidation, err))

		return nil, false
	}

	return rumpf, true
}

// ausEingabe liest, was hereingeschoben wurde — **Teile oder ganzer Rumpf**.
//
// **Eine Liste sind die Teile, ein Objekt ist der ganze Rumpf.** Das erste
// ist die Kette:
//
//	mse symbol-rechnen … | mse symbol-rechnen …
//
// Das zweite ist dieselbe Gestalt, die MCP und der Dienst nehmen — damit
// derselbe Rumpf über alle drei Flächen geht und ein Flächenlauf sie
// wirklich vergleichen kann (§16). Ohne den zweiten Weg hätte die
// Kommandozeile eine eigene Sprache für dieselbe Sache.
func ausEingabe(an *schreiber) (map[string]any, bool) {
	roh := strings.TrimSpace(gelesenerText(an.eingabe))
	if roh == "" {
		return map[string]any{}, true
	}

	if strings.HasPrefix(roh, "[") {
		var gelesen []app.SymbolteilDTO
		if err := json.Unmarshal([]byte(roh), &gelesen); err != nil {
			an.klageUeber(app.SymbolteileUnlesbar(err))

			return nil, false
		}

		return map[string]any{"teile": gelesen}, true
	}

	var rumpf map[string]any
	if err := json.Unmarshal([]byte(roh), &rumpf); err != nil {
		an.klageUeber(app.SymbolteileUnlesbar(err))

		return nil, false
	}

	return rumpf, true
}

// rechnungsnamen zählt auf, was es gibt.
//
// **Aus der Tabelle und nicht daneben aufgeschrieben** — eine zweite Liste
// liefe auseinander, und die Hilfe nennte eine Rechnung, die es nicht gibt.
func rechnungsnamen() string {
	return "Rechnungen: " + strings.Join(app.Symbolrechnungsnamen(), ", ")
}

// gelesenerText liest, was hereingeschoben wurde.
func gelesenerText(eingabe io.Reader) string {
	if eingabe == nil {
		return ""
	}

	roh, err := io.ReadAll(eingabe)
	if err != nil {
		return ""
	}

	return string(roh)
}

// symbolLesenBefehl gibt das abgelegte Symbol eines Sets.
func symbolLesenBefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 1 {
		an.klage("symbol braucht die Setkennung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	name := ""
	if len(rest) > 1 {
		name = rest[1]
	}

	// **Mit dem Bestand** (#1815): wem das Set gehört, weiß nur er.
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		naht, err := app.SymbolwerkstattUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		teile, fehler := naht.Lesen(ctx, setID, name)
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		return alsJSONSchreiben(an, teile)
	})
}

// symbolSetzenBefehl legt eine Zeichnung neben das Set.
func symbolSetzenBefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 1 {
		an.klage("symbol-setzen braucht die Setkennung.\n%s", hilfe)

		return CodeBedienung
	}

	setID, gut := kennung(an, "die Setkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	var teile []app.SymbolteilDTO
	if err := json.Unmarshal([]byte(gelesenerText(an.eingabe)), &teile); err != nil {
		an.klageUeber(app.SymbolteileUnlesbar(err))

		return CodeBedienung
	}

	unter := ""
	if len(rest) > 1 {
		unter = rest[1]
	}

	// **Mit dem Bestand** (#1815): wem das Set gehört, weiß nur er.
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		naht, err := app.SymbolwerkstattUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		name, fehler := naht.Ablegen(ctx, setID, teile, unter)
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, map[string]any{"name": name})
		}

		an.sag("Abgelegt als %s.\n", name)

		return an.abschluss()
	})
}

// symbolpaletteBefehl gibt, was ein Feld zum Einfügen anbietet.
//
// **Ohne die Bildchen** — der Grund steht in
// `internal/app/symbolpalettenaht.go`.
func symbolpaletteBefehl(
	ctx context.Context, bestandsort string, an *schreiber, rest []string,
) int {
	if len(rest) < 2 {
		an.klage("symbolpalette braucht Kartenkennung und Feldname.\n%s", hilfe)

		return CodeBedienung
	}

	karteID, gut := kennung(an, "die Kartenkennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		werk, err := app.Zeichennaht(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		zeilen, fehler := werk.Palette(ctx, karteID, rest[1])
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, map[string]any{"eintraege": zeilen})
		}

		for _, zeile := range zeilen {
			an.sag("%s\t%s\t%s\n", zeile.Art, zeile.Code, zeile.Beschriftung)
		}

		return an.abschluss()
	})
}
