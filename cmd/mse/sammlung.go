// **Die Sammlung an der Kommandozeile** — Werkstück 2 von Phase 3.4
// (#1165).
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort. Eine Sammlung von
// dreihundert Karten trägt niemand von Hand ein — sie kommt aus einer Datei,
// und dafür ist die Kommandozeile da.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// sammlungsbefehl zählt die Sammlung auf — ganz oder zu einem Druck.
func sammlungsbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	return mitSammlung(ctx, bestandsort, an, func(ctx context.Context, naht *app.Sammlungsnaht) int {
		exemplare, zahlen, fehler := sammlungHolen(ctx, naht, rest)
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			antwort := map[string]any{"exemplare": alsExemplarsichten(exemplare)}

			// **Ein Ausschnitt nennt seine Zahlen**, wie über REST und MCP
			// (U15, Befund 4 des Zweitblicks über Phase G) — erweiternd: wer
			// nur `exemplare` liest, liest weiter dasselbe.
			if zahlen != nil {
				antwort["gesamt"] = zahlen.Gesamt
				antwort["insgesamt"] = zahlen.Insgesamt
				antwort["stueck"] = zahlen.Stueck
			}

			return alsJSONSchreiben(an, antwort)
		}

		// **Jedes Feld, das aufgenommen werden kann, steht auch hier.**
		//
		// Die ersten sechs Spalten waren es einmal, und die vier dahinter —
		// Erwerbsdatum, Preis, Währung, Notiz — gingen über dieselbe Fläche
		// hinein und kamen nie zurück. Wer sie eintippte, sah sie nie wieder
		// und hatte keinen Grund, das zu bemerken. Angestrichen vom Zweitblick
		// über Phase 3.4; gehalten wird es jetzt vom Abnahmelauf, der die
		// Felder vergleicht statt ihre Existenz.
		for _, exemplar := range exemplare {
			an.sag("%d\t%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				exemplar.ID, exemplar.Printing, exemplar.Menge,
				exemplar.Zustand, exemplar.Sprache, exemplar.Finish,
				exemplar.Erwerbsdatum, exemplar.Erwerbspreis,
				exemplar.Erwerbswaehrung, exemplar.Notiz,

				// **Der lesbare Bezug am Ende** (#1200) — angehängt und nicht
				// eingeschoben: wer die Spalten schon einliest, zählt weiter.
				exemplar.Bezeichnung)
		}

		return an.abschluss()
	})
}

// sammlungHolen entscheidet zwischen ganzer Sammlung und einem Druck — und
// gibt bei einer Sicht ihre Zahlen mit.
func sammlungHolen(ctx context.Context, naht *app.Sammlungsnaht, rest []string) (
	[]app.ExemplarDTO, *domain.Fensterzahlen, error,
) {
	// **Ohne Angabe alles** — der Ruf, den es immer gab.
	if len(rest) == 0 {
		exemplare, fehler := naht.Sammlung(ctx)

		return exemplare, nil, fehler
	}

	// **Mit Schaltern eine Sicht** (Werkstück
	// #1500):
	// `--suche`, `--foil`, `--nonfoil`, `--etched`, `--nach`, `--ab`,
	// `--anzahl`. Ohne
	// sie bleibt das erste Wort die **Druckkennung**, wie bisher — der
	// Vertrag ändert sich erweiternd und nicht brechend.
	sicht, ab, anzahl, erkannt := sammlungssicht(rest)
	if !erkannt {
		exemplare, fehler := naht.SammlungFuer(ctx, rest[0])

		return exemplare, nil, fehler
	}

	fenster, zahlen, fehler := naht.Sammlungsfenster(ctx, sicht, ab, anzahl)

	return fenster, &zahlen, fehler
}

// sammlungssicht liest die Schalter — und sagt, ob überhaupt einer dastand.
func sammlungssicht(rest []string) (
	sicht domain.Sammlungssicht, ab, anzahl int, erkannt bool,
) {
	for stelle := 0; stelle < len(rest); stelle++ {
		wert := ""
		if stelle+1 < len(rest) {
			wert = rest[stelle+1]
		}

		switch rest[stelle] {
		case "--suche":
			sicht.Suche, erkannt, stelle = wert, true, stelle+1
		case "--nach":
			sicht.Sortierung, erkannt, stelle = domain.Sortierung(wert), true, stelle+1
		case "--ab":
			ab, erkannt, stelle = zahlOderNull(wert), true, stelle+1
		case "--anzahl":
			anzahl, erkannt, stelle = zahlOderNull(wert), true, stelle+1
		case "--foil":
			sicht.Ausfuehrung, erkannt = domain.AusfuehrungFoil, true
		case "--nonfoil":
			sicht.Ausfuehrung, erkannt = domain.AusfuehrungNonfoil, true

		// **Die dritte Ausführung** (T2, #1467) — erweiternd: wer `--foil`
		// tippt, bekommt weiter genau die Foils und nicht mehr die geätzten
		// dazu.
		case "--etched":
			sicht.Ausfuehrung, erkannt = domain.AusfuehrungEtched, true
		}
	}

	return sicht, ab, anzahl, erkannt
}

// zahlOderNull liest eine Zahl — was keine ist, ist null.
//
// **Ohne Fehler**: eine krumme Zahl an `--ab` heißt „von vorn", und das ist
// die harmlose Deutung. Was **nicht** harmlos wäre, ist eine erfundene
// Ordnung — die weist `domain.Sammlungssicht.Pruefe` ab.
func zahlOderNull(wert string) int {
	zahl, fehler := strconv.Atoi(wert)
	if fehler != nil || zahl < 0 {
		return 0
	}

	return zahl
}

// exemplarbefehl nimmt ein Exemplar auf — oder sagt, welche Decks es benutzen.
//
// **Der Rückweg steht als Unterbefehl davor** (U9, #1420): das erste Wort ist
// sonst eine **Druckkennung**, und die ist eine UUID — »decks« ist keine.
func exemplarbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) > 0 && rest[0] == "decks" {
		return mitDecks(ctx, bestandsort, an, func(ctx context.Context, flaeche *app.Deckflaeche) int {
			return exemplarDecks(ctx, flaeche, an, rest[1:])
		})
	}

	if len(rest) < 2 {
		an.klage("exemplar braucht Druck und Menge.\n%s", hilfe)

		return CodeBedienung
	}

	menge, gut := kennung(an, "die Menge", rest[1])
	if !gut {
		return CodeBedienung
	}

	return mitSammlung(ctx, bestandsort, an, func(ctx context.Context, naht *app.Sammlungsnaht) int {
		id, fehler := naht.ExemplarAnlegen(ctx, exemplarAus(rest[0], int(menge), rest[2:]))
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		if an.alsJSON {
			return alsJSONSchreiben(an, map[string]any{"id": id})
		}

		an.sag("%d\n", id)

		return an.abschluss()
	})
}

// exemplarAendernbefehl ändert die Menge eines Exemplars.
//
// # Warum nur die Menge und warum vorher gelesen wird
//
// Die Bindung setzt **alle** Felder — sie ist für ein Fenster gebaut, das sie
// alle kennt. Der erste Anlauf baute hier aus zwei Argumenten ein ganzes DTO
// und schrieb damit Erwerbsdatum, Preis, Währung und Notiz **leer**: ein über
// den Dienst sorgfältig erfasstes Exemplar verlor seine Angaben, sobald jemand
// an der Kommandozeile die Menge änderte. Angestrichen vom Zweitblick über
// Phase 3.4.
//
// **Deshalb wird gelesen, geändert, geschrieben.** Und die **Fassung** kommt
// aus demselben Lesen: ändert jemand dazwischen, gibt es einen Konflikt statt
// eines stillen Überschreibens (§7, Regel 2).
func exemplarAendernbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 2 {
		an.klage("exemplar-aendern braucht Kennung und Menge.\n%s", hilfe)

		return CodeBedienung
	}

	id, gut := kennung(an, "die Kennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	menge, mengeGut := kennung(an, "die Menge", rest[1])
	if !mengeGut {
		return CodeBedienung
	}

	return mitSammlung(ctx, bestandsort, an, func(ctx context.Context, naht *app.Sammlungsnaht) int {
		vorher, fehler := exemplarMitKennung(ctx, naht, id)
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		vorher.Menge = int(menge)

		if schreibfehler := naht.ExemplarAendern(ctx, vorher); schreibfehler != nil {
			an.klageUeber(schreibfehler)

			return CodeFachlich
		}

		return an.abschluss()
	})
}

// exemplarMitKennung sucht ein Exemplar in der Sammlung.
//
// **Über die Liste und nicht über eine eigene Abfrage**: einen Weg „gib mir
// Exemplar 7" gibt es nicht, und ihn nur für die Kommandozeile zu bauen wäre
// eine fünfte Fläche mit einer eigenen Frage.
func exemplarMitKennung(
	ctx context.Context, naht *app.Sammlungsnaht, id int64,
) (app.ExemplarDTO, error) {
	alle, err := naht.Sammlung(ctx)
	if err != nil {
		return app.ExemplarDTO{}, err
	}

	for _, exemplar := range alle {
		if exemplar.ID == id {
			return exemplar, nil
		}
	}

	return app.ExemplarDTO{}, fmt.Errorf(
		"%w: Exemplar %d gibt es nicht", domain.ErrNotFound, id)
}

// exemplarLoeschenbefehl nimmt ein Exemplar aus der Sammlung.
//
// **Auch hier wird vorher gelesen** — aus demselben Grund wie beim Ändern: die
// Fassung sagt, welchen Stand dieser Ruf gesehen hat. Wer tippt, hat sie nicht
// zur Hand, und sie zu verlangen hieße, dem Menschen eine Buchführung
// aufzubürden, die das Programm selbst führen kann (§16 nennt das die
// Ausnahme, die begründet gehört).
func exemplarLoeschenbefehl(ctx context.Context, bestandsort string, an *schreiber, rest []string) int {
	if len(rest) < 1 {
		an.klage("exemplar-loeschen braucht die Kennung.\n%s", hilfe)

		return CodeBedienung
	}

	id, gut := kennung(an, "die Kennung", rest[0])
	if !gut {
		return CodeBedienung
	}

	return mitSammlung(ctx, bestandsort, an, func(ctx context.Context, naht *app.Sammlungsnaht) int {
		vorher, fehler := exemplarMitKennung(ctx, naht, id)
		if fehler != nil {
			an.klageUeber(fehler)

			return CodeFachlich
		}

		if schreibfehler := naht.ExemplarLoeschen(
			ctx, id, vorher.Fassung); schreibfehler != nil {
			an.klageUeber(schreibfehler)

			return CodeFachlich
		}

		return an.abschluss()
	})
}

// exemplarAus baut ein Exemplar aus den Stellungsargumenten.
//
// **Die Reihenfolge ist die der Hilfe**: Zustand, Sprache, Foil, Erwerbsdatum,
// Erwerbspreis, Währung, Notiz. Was fehlt, bekommt seine Vorgabe in der
// Domäne — `nearMint` und `en`.
//
// **Alle sieben stehen hier**, seit der Zweitblick über Phase 3.4 gezeigt hat,
// dass die Kommandozeile vier davon gar nicht ausdrücken konnte: eine Fläche,
// die weniger sagen kann als die anderen, führt nicht dieselbe Sammlung (§16).
func exemplarAus(printing string, menge int, weitere []string) app.ExemplarDTO {
	exemplar := app.ExemplarDTO{Printing: printing, Menge: menge}

	felder := []*string{
		&exemplar.Zustand, &exemplar.Sprache, nil,
		&exemplar.Erwerbsdatum, &exemplar.Erwerbspreis,
		&exemplar.Erwerbswaehrung, &exemplar.Notiz,

		// **Der lesbare Bezug** (#1200) — wahlfrei, und ohne ihn steht die
		// Kennung da.
		&exemplar.Name, &exemplar.Setcode, &exemplar.Nummer,

		// **Der Lagerort** (#1292, Vermerk 0057) — Freitext.
		&exemplar.Lagerort,

		// **Und die Absicht.** `nil` heißt wieder: kein Text, sondern ein
		// Schalter — dieselbe Stelle wie bei der Ausführung.
		nil,

		// **Rang und Art eines Wunsches** (T4, #1469) — hinten angehängt,
		// weil jede Stelle davor schon vergeben ist und eine Umstellung
		// jeden bestehenden Aufruf bräche. An einem Besitz bedeuten sie
		// nichts; die Domäne räumt sie dort weg.
		&exemplar.Wunschrang, &exemplar.Wunschart,
	}

	for stelle, wert := range weitere {
		if stelle >= len(felder) {
			break
		}

		if felder[stelle] == nil {
			// **Zwei Stellen sind Schalter und keine Texte**: die dritte ist
			// die Ausführung, die elfte die Absicht.
			if stelle == 2 {
				exemplar.Finish = wert
			} else {
				exemplar.Wunsch = wert == "wunsch"
			}

			continue
		}

		*felder[stelle] = wert
	}

	return exemplar
}

// mitSammlung öffnet den Bestand und baut die Naht.
func mitSammlung(
	ctx context.Context, bestandsort string, an *schreiber,
	arbeit func(context.Context, *app.Sammlungsnaht) int,
) int {
	return mitBestand(ctx, bestandsort, an, func(ctx context.Context, db *sql.DB) int {
		naht, err := app.SammlungsnahtUeber(db)
		if err != nil {
			an.klageUeber(err)

			return CodeFachlich
		}

		return arbeit(ctx, naht)
	})
}
