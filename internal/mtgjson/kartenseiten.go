package mtgjson

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Was MSE über mehrseitige Karten weiß, und woher wir es wissen.
//
// # Warum, und woran es belegt ist
//
// **MSE führt eine mehrseitige Karte als eine Karte mit zwei Feldsätzen.**
// Zwei unabhängige Belege, beide am 2026-08-24:
//
//   - `magic.mse-game/card_fields` definiert 249 Kartenfelder, davon **46 mit
//     dem Zusatz ` 2`** — `name 2`, `casting cost 2`, `rule text 2`,
//     `super type 2`, `power 2` und so fort.
//
//     **41 tragen eine ` 3`**, darunter `name 3`, `casting cost 3` und
//     `rule text 3` — nachgezählt in beiden Packs. Sie werden nicht benutzt:
//     eine Karte hat **zwei** Seiten, nicht drei. Die Referenz benutzt ` 3`
//     für etwas anderes — `rule text 3` trägt dort das Schlüsselwort *Fuse*
//     (`src/mse.rs`, Zeile 293).
//   - `magic-set-generator` (MIT, §11) schreibt es genauso: sein Makro
//     `push_alt!` in `src/mse.rs` hängt bei der zweiten Seite ` 2` an jeden
//     Schlüssel und legt **beide** Seiten in dieselbe Karte.
//
// **MTGJSON macht das Gegenteil**: je Seite ein eigenes Kartenobjekt, beide
// mit demselben zusammengesetzten `name` und verbunden über `otherFaceIds`.
// Wer sie einzeln übernimmt, bekommt zwei MSE-Karten, die keine Vorlage als
// Vorder- und Rückseite zeichnen kann.
//
// **Gefunden vom Zweitblick über Phase V** — die Korrektur in
// #1018 hat
// den **Namen** richtig gestellt und die **Struktur** falsch gelassen. Das ist
// genau die Klasse, vor der AGENTS.md §10 warnt: den Fall behoben, die Klasse
// stehen gelassen.

// rohseiten liest die Angaben, die Seiten verbinden.
type rohseiten struct {
	Kennung  string   `json:"uuid"`
	Seite    string   `json:"side"`
	Andere   []string `json:"otherFaceIds"`
	Feldname string   `json:"faceName"`
}

// seitenAus liest die Verbindungsangaben einer rohen Karte.
//
// **Fehlt eine, ist die Karte einseitig.** Eine gewöhnliche Karte führt weder
// `side` noch `otherFaceIds`; sie durch dieses Verfahren zu schicken wäre
// aufwendig für ein Ergebnis, das feststeht.
func seitenAus(roh map[string]json.RawMessage) rohseiten {
	var gelesen rohseiten

	for schluessel, ziel := range map[string]any{
		"uuid":         &gelesen.Kennung,
		"side":         &gelesen.Seite,
		"otherFaceIds": &gelesen.Andere,
		"faceName":     &gelesen.Feldname,
	} {
		if wert, da := roh[schluessel]; da {
			_ = json.Unmarshal(wert, ziel) //nolint:errcheck // eine unlesbare Angabe heißt: keine Angabe
		}
	}

	return gelesen
}

// zweiteSeite hängt an einen Feldnamen den Zusatz der zweiten Seite.
//
// **Die Regel steht in `domain`** (G2,
// #1459).
// Sie ist eine Aussage über eine **Karte**, nicht über ein
// Ausfuhrformat: der Spieltisch stellt dieselbe Frage, wenn er die zweite
// Seite zeichnen soll. Zwei Fassungen liefen an dem Tag auseinander, an dem
// jemand eine ändert (§2).
func zweiteSeite(feld string) string { return domain.ZweiteSeite(feld) }

// fuehreSeitenZusammen macht aus den Karten eines Sets die Karten, die MSE
// kennt — mehrseitige zusammengelegt, einseitige unverändert.
//
// **Die Reihenfolge kommt aus `side`, nicht aus der Datei.** MTGJSON sichert
// keine Reihenfolge der Kartenobjekte zu; wer die Dateireihenfolge nähme,
// bekäme bei mancher Datei die Rückseite vorn, und die Vorlage zeichnete sie
// als Vorderseite.
func fuehreSeitenZusammen(
	karten []*Karte, rohe []map[string]json.RawMessage,
) ([]*Karte, []Hinweis) {
	gruppen, gehoertZu, hinweise := gruppiereSeiten(rohe)

	var (
		zusammen []*Karte
		erledigt = map[int]bool{}
	)

	for stelle, karte := range karten {
		if erledigt[stelle] {
			continue
		}

		kopf, da := gehoertZu[stelle]
		if !da {
			zusammen = append(zusammen, karte)

			continue
		}

		mitglieder := gruppen[kopf]
		for _, eines := range mitglieder {
			erledigt[eines] = true
		}

		zusammen = append(zusammen, legeSeitenAufeinander(karten, rohe, mitglieder))
	}

	return zusammen, hinweise
}

// gruppiereSeiten findet, welche Stellen zu derselben Karte gehören.
//
// **Über `otherFaceIds`, nicht über den Namen.** Zwei verschiedene Karten
// können denselben Namen tragen — MTGJSON verbindet sie über Kennungen, und
// die sind eindeutig.
func gruppiereSeiten(
	rohe []map[string]json.RawMessage,
) (gruppen map[int][]int, gehoertZu map[int]int, hinweise []Hinweis) {
	zuKennung := map[string]int{}
	angaben := make([]rohseiten, len(rohe))

	for stelle, roh := range rohe {
		angaben[stelle] = seitenAus(roh)
		if angaben[stelle].Kennung != "" {
			zuKennung[angaben[stelle].Kennung] = stelle
		}
	}

	gruppen = map[int][]int{}
	gehoertZu = map[int]int{}

	for stelle, angabe := range angaben {
		// **Eine Vorderseite zeigt auf genau eine andere Seite.** Wer zwei
		// oder mehr nennt, ist eine **geteilte Rückseite** — gemessen, nicht
		// vermutet: siehe der Kopf dieser Datei.
		if len(angabe.Andere) != 1 {
			continue
		}

		partner, da := zuKennung[angabe.Andere[0]]
		if !da {
			// **Ein Verweis ins Leere wird gemeldet, nicht verschluckt** (§8).
			// Er kommt vor, wenn eine Datei nur einen Teil eines Satzes führt:
			// die Vorderseite ist da, die Rückseite nicht. Die Karte bleibt —
			// aber wer sie später als einseitig ansieht, soll wissen, warum.
			hinweise = append(hinweise, Hinweis{
				Feld: "otherFaceIds", JeKarte: true,
				Meldung: fmt.Sprintf(
					"die andere Seite %q liegt nicht in dieser Datei; die Karte "+
						"bleibt einseitig", angabe.Andere[0]),
			})

			continue
		}

		// **Ein gegenseitiges Paar wird einmal gruppiert, nicht zweimal.**
		// Bei `transform` zeigen beide Seiten aufeinander; wer für jede eine
		// Gruppe bildete, bekäme die Karte doppelt. Die kleinere Stelle führt.
		if len(angaben[partner].Andere) == 1 && partner < stelle {
			continue
		}

		gruppen[stelle] = []int{stelle, partner}
		gehoertZu[stelle] = stelle

		// **Die Rückseite wird keine eigene Karte.** Bei `meld` gehört sie zu
		// **beiden** Vorderseiten; welcher Gruppe sie hier zugeordnet wird,
		// ist gleich — der Eintrag verhindert nur, dass sie für sich steht.
		gehoertZu[partner] = stelle
	}

	return gruppen, gehoertZu, hinweise
}

// legeSeitenAufeinander baut aus mehreren Seiten eine Karte.
func legeSeitenAufeinander(
	karten []*Karte, rohe []map[string]json.RawMessage, mitglieder []int,
) *Karte {
	sortiert := nachSeite(rohe, mitglieder)

	gelegt := &Karte{
		Werte: append([]domain.Value{}, karten[sortiert[0]].Werte...),

		// **Kopiert und nicht geteilt.** Gleich wird angehängt; auf dem
		// Rücken einer fremden Reihe zu wachsen hieße, ihr Feld zu
		// überschreiben, sobald der Vorrat reicht.
		Rest: append([]domain.Value{}, karten[sortiert[0]].Rest...),
	}

	// **Ohne Hinweise, und das ist kein Versehen.** Eine Gruppe hat seit
	// #1038
	// **genau zwei** Mitglieder — eine Vorderseite und ihre Rückseite. Ein
	// Zweig für eine dritte Seite wäre eine Zeile, die kein Aufruf betreten
	// kann (§3); was mit einer unbekannten Verbindungsgestalt geschieht,
	// entscheidet [gruppiereSeiten]: sie wird nicht gruppiert, und keine Karte
	// geht verloren. Was sich **nicht** auflösen lässt, meldet sie dort.
	for _, eines := range sortiert[1:] {
		for _, wert := range karten[eines].Werte {
			gelegt.Werte = append(gelegt.Werte, domain.Value{
				Field: zweiteSeite(wert.Field), Text: wert.Text,
			})
		}

		// **Und ihr Rest ebenso** — mit demselben Zusatz
		// (#1039).
		//
		// `Karte.Rest` sagt zu: „roh mitgeführt, damit ein Rundlauf sie nicht
		// verliert" — für die zweite Seite genauso.
		//
		// **Der Zusatz ist derselbe wie bei den Werten, und das ist der
		// Grund, aus dem es trägt**: [TrenneKartenwerte] nimmt ihn auf dem
		// Rückweg wieder ab, und der Schlüssel landet in dem Kartenobjekt, aus
		// dem er kam. Ein eigener Mechanismus daneben liefe auseinander.
		for _, wert := range karten[eines].Rest {
			gelegt.Rest = append(gelegt.Rest, domain.Value{
				Field: zweiteSeite(wert.Field), Text: wert.Text,
			})
		}
	}

	return gelegt
}

// nachSeite sortiert die Stellen nach der Angabe `side`.
//
// **Ohne `side` bleibt die Dateireihenfolge.** Eine von Hand geschriebene
// Datei kann `otherFaceIds` führen und `side` nicht; eine erfundene
// Reihenfolge wäre schlechter als die vorgefundene.
func nachSeite(rohe []map[string]json.RawMessage, mitglieder []int) []int {
	sortiert := append([]int{}, mitglieder...)

	sort.SliceStable(sortiert, func(a, b int) bool {
		return seitenAus(rohe[sortiert[a]]).Seite < seitenAus(rohe[sortiert[b]]).Seite
	})

	return sortiert
}
