package storage

import (
	"context"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// suchen klemmt auf das Profil über das Set der Karte.
//
// **Ohne diese Bedingung fand die Suche über alle Sets die Karten jedes
// Profils** — gefunden von der Messung zu ENG-05.AK-02 (#1730). Eine Karte
// ist ein Abkömmling (0035); sie gehört dem Profil ihres Sets, und nur über
// das Set ist zu sehen, welchem.
const suchen = `
SELECT card_suche.card_id, cards.set_id, cards.reihenfolge,
       card_suche.feld, card_suche.wert
  FROM card_suche
  JOIN cards ON cards.id = card_suche.card_id
  JOIN sets ON sets.id = cards.set_id
 WHERE card_suche MATCH ?
   AND sets.profil_id = ?
   AND (? = '' OR card_suche.feld = ?)
   AND (? = 0 OR cards.set_id = ?)
 ORDER BY rank
 LIMIT ?`

// Suche findet Kartenstellen über den Volltextindex.
//
// **Über FTS5 und nicht über einen Durchlauf** — das ist der Grund, warum
// SQLite das Arbeitsformat ist (`docs/entscheidungen/0002`).
func (l *SetLeser) Suche(ctx context.Context, frage domain.SearchQuery) (
	domain.SearchResult, error,
) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return domain.SearchResult{}, profilfehler
	}

	muster := fts5Muster(frage.Text)
	if muster == "" {
		// **Auch die Marken bleiben dann außen vor.** Ein leeres Muster heißt,
		// dass der Text nichts enthält, wonach sich suchen ließe; eine
		// Markensuche darüber gäbe alles zurück.
		return domain.SearchResult{}, nil
	}

	// **Eine Zeile mehr als die Schranke.** Daran und nur daran ist zu erkennen,
	// dass es mehr gibt — eine Abfrage mit genau der Schranke lieferte bei
	// genau so vielen Treffern dasselbe Bild wie bei tausend.
	zeilen, err := l.db.QueryContext(ctx, suchen,
		muster, profilID, frage.Field, frage.Field, frage.SetID, frage.SetID, domain.SearchLimit+1)
	if err != nil {
		return domain.SearchResult{}, uebersetze(err, "suchen")
	}

	defer func() { _ = zeilen.Close() }()

	treffer := []domain.SearchHit{}

	for zeilen.Next() {
		var fund domain.SearchHit
		if err := zeilen.Scan(&fund.CardID, &fund.SetID, &fund.Number,
			&fund.Field, &fund.Value); err != nil {
			return domain.SearchResult{}, uebersetze(err, "einen Treffer lesen")
		}

		treffer = append(treffer, fund)
	}

	if err := zeilen.Err(); err != nil {
		return domain.SearchResult{}, uebersetze(err, "suchen")
	}

	// **Und die Marken** (#1208) — dieselbe Frage, dieselbe Fläche. Sie
	// stehen **hinter** den Feldwerten: die Rangfolge von FTS5 gilt für den
	// Text, und eine Marke hat keine.
	marken, markenfehler := l.Markentreffer(ctx, frage)
	if markenfehler != nil {
		return domain.SearchResult{}, markenfehler
	}

	treffer = append(treffer, marken...)

	if len(treffer) > domain.SearchLimit {
		return domain.SearchResult{Hits: treffer[:domain.SearchLimit], Truncated: true}, nil
	}

	return domain.SearchResult{Hits: treffer}, nil
}

// fts5Muster macht aus einer Eingabe eine FTS5-Abfrage.
//
// # Warum die Eingabe nicht durchgereicht wird
//
// FTS5 hat eine eigene Sprache: `"`, `*`, `NEAR`, `AND`, `-` bedeuten dort
// etwas. Wer einen Kartennamen mit einem Anführungszeichen sucht, bekäme sonst
// einen Syntaxfehler statt eines Treffers — und ein Anführungszeichen in einem
// Kartennamen ist keine Seltenheit.
//
// Jedes Wort wird deshalb **eingefasst** und mit `*` erweitert: wer „feuer"
// tippt, meint auch „Feuerball". Mehrere Wörter gelten zusammen; das ist die
// Erwartung, die jede Suche erfüllt.
func fts5Muster(eingabe string) string {
	var teile []string

	for wort := range strings.FieldsSeq(eingabe) {
		// Ein Anführungszeichen wird in FTS5 durch Verdopplung eingefasst.
		sauber := strings.ReplaceAll(wort, `"`, `""`)
		teile = append(teile, `"`+sauber+`"*`)
	}

	return strings.Join(teile, " ")
}

// Felder gibt die Feldnamen, die im Bestand vorkommen.
//
// **Aus dem Bestand und nicht aus einer Liste bei uns.** Ein fremdes Spiel
// führt Felder, die wir nicht kennen — ein Filter über eine selbst geschriebene
// Aufzählung ließe ausgerechnet die nicht filtern. `setID` 0 heißt: über alle
// Sets.
//
// **Und nur aus dem gezeigten Profil** (#1730): über alle Sets hieß bis
// dahin über die Sets jedes Profils.
func (l *SetLeser) Felder(ctx context.Context, setID int64) ([]string, error) {
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, err := l.db.QueryContext(ctx,
		`SELECT DISTINCT card_values.feld FROM card_values
		   JOIN cards ON cards.id = card_values.card_id
		   JOIN sets ON sets.id = cards.set_id
		  WHERE sets.profil_id = ?
		    AND (? = 0 OR cards.set_id = ?)
		  ORDER BY card_values.feld`, profilID, setID, setID)
	if err != nil {
		return nil, uebersetze(err, "Feldnamen lesen")
	}

	defer func() { _ = zeilen.Close() }()

	felder := []string{}

	for zeilen.Next() {
		var feld string
		if err := zeilen.Scan(&feld); err != nil {
			return nil, uebersetze(err, "einen Feldnamen lesen")
		}

		felder = append(felder, feld)
	}

	return felder, uebersetze(zeilen.Err(), "Feldnamen lesen")
}

// Kartenwerte gibt die Feldwerte mehrerer Karten auf einmal.
//
// Für die Trefferliste: sie zeigt je Treffer den Kartennamen, und der steht in
// einem anderen Feld als dem getroffenen.
func (l *SetLeser) Kartenwerte(ctx context.Context, karten []int64) (
	map[int64][]domain.Value, error,
) {
	if len(karten) == 0 {
		return map[int64][]domain.Value{}, nil
	}

	// **Die Liste kommt als **ein** Parameter, nicht als zusammengesetzte
	// Abfrage.** Der erste Anlauf baute `IN (?,?,?)` aus so vielen Fragezeichen,
	// wie Karten da waren — sicher, weil nur Fragezeichen verkettet werden, und
	// trotzdem die Form, an der `gosec` zu Recht hängen bleibt: sie sieht aus
	// wie die unsichere und ist von ihr nur durch Nachlesen zu unterscheiden.
	//
	// `json_each` nimmt die Kennungen als JSON-Feld. Eine Abfrage, ein
	// Parameter, kein Zusammenbau **in der Abfrage**.
	//
	// Gebaut wird das Feld von Hand statt über `json.Marshal`: dessen Fehlerzweig
	// kann bei einer Liste ganzer Zahlen nicht eintreten, und ein Zweig, den kein
	// Test betreten kann, ist eine Behauptung darüber, was im Ernstfall
	// geschieht.
	//
	// **Und nur Karten des gezeigten Profils** (#1815): eine fremde Kennung
	// fehlt im Ergebnis wie eine, die es nicht gibt — dieselbe Antwort, die
	// diese Funktion für eine gelöschte Karte schon immer gab.
	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	zeilen, err := l.db.QueryContext(ctx,
		`SELECT card_values.card_id, card_values.feld, card_values.wert FROM card_values
		   JOIN cards ON cards.id = card_values.card_id
		   JOIN sets ON sets.id = cards.set_id
		  WHERE sets.profil_id = ? AND card_values.card_id IN (SELECT value FROM json_each(?))
		  ORDER BY card_values.card_id, card_values.reihenfolge`, profilID, alsJSONFeld(karten))
	if err != nil {
		return nil, uebersetze(err, "Kartenwerte lesen")
	}

	defer func() { _ = zeilen.Close() }()

	gefunden := map[int64][]domain.Value{}

	for zeilen.Next() {
		var (
			karte int64
			wert  domain.Value
		)

		if err := zeilen.Scan(&karte, &wert.Field, &wert.Text); err != nil {
			return nil, uebersetze(err, "einen Kartenwert lesen")
		}

		gefunden[karte] = append(gefunden[karte], wert)
	}

	return gefunden, uebersetze(zeilen.Err(), "Kartenwerte lesen")
}

// alsJSONFeld schreibt Kennungen als JSON-Zahlenfeld.
func alsJSONFeld(kennungen []int64) string {
	var gebaut strings.Builder

	gebaut.WriteByte('[')

	for i, kennung := range kennungen {
		if i > 0 {
			gebaut.WriteByte(',')
		}

		gebaut.WriteString(strconv.FormatInt(kennung, 10))
	}

	gebaut.WriteByte(']')

	return gebaut.String()
}
