package storage

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Marken sind mitzufinden** —
// #1208.
//
// # Warum dieselbe Suchfläche und keine zweite
//
// Das Issue lässt beides zu: „dieselbe Suchfläche wie für Feldwerte, oder eine
// begründete zweite". Es ist **dieselbe** geworden, weil die Frage dieselbe
// ist: „wo steht das Wort?" Wer „prüfen" sucht, meint die Karten, an denen es
// steht — ob im Regeltext oder an der Karte, entscheidet **er**, nicht das
// Programm. Der Treffer sagt es ihm: er trägt das Feld [domain.Markenfeld].
//
// **Eine zweite Fläche hätte den Preis, dass niemand sie findet.** ⌘F ist die
// Suche; eine Marke, die nur über eine eigene Eingabe zu finden wäre, wäre für
// die Hälfte der Nutzenden gar nicht da.
//
// # Warum nicht über FTS5
//
// Weil der Index `card_suche` über **Feldwerte** geht und eine Marke keiner
// ist. Sie mit hineinzuschreiben hieße, den Index bei jedem Setzen einer Marke
// fortzuschreiben — und ein Kartenfeld namens `marke` überschriebe sie.
//
// Eine Marke ist ein **Wort** und höchstens [domain.Markengrenze] Zeichen
// lang; über den Index `karten_marken_marke` ist ein Teilstringvergleich hier
// die einfachere und ehrliche Antwort.
//
// # Vom Profil aus gelesen, nicht von den Marken
//
// **Bis zum 2026-09-23 las sie `karten_marken` ganz**, über jedes Profil — und
// fand die Marken fremder Karten mit (#1730, gefunden von der Messung zu
// AK-02). Ein Teilstring trifft keinen Index; eingegrenzt wird deshalb über
// die Sets des Profils (`sets_profil`), ihre Karten (`cards (set_id, …)`) und
// deren Marken (Primärschlüssel ab `card_id`). Gezählt:
// `TestDieListenLesenNichtDenGanzenBestand`.
const markensuchen = `
SELECT k.card_id, c.set_id, c.reihenfolge, k.marke
  FROM sets s
  JOIN cards c ON c.set_id = s.id
  JOIN karten_marken k ON k.card_id = c.id
 WHERE s.profil_id = ?
   AND (? = 0 OR s.id = ?)
   AND instr(k.marke, ?) > 0
 ORDER BY k.marke, c.reihenfolge
 LIMIT ?`

// Markentreffer findet Karten, deren Marke den Text enthält.
//
// **Auf der Normalform verglichen**: die Marken liegen kleingeschrieben im
// Bestand, und wer „Prüfen" sucht, meint „prüfen".
func (l *SetLeser) Markentreffer(ctx context.Context, frage domain.SearchQuery) (
	[]domain.SearchHit, error,
) {
	// **Ein Feldfilter schließt Marken aus — außer er meint sie.** Wer in
	// „regeltext" sucht, sucht nicht in Marken; wer `marke` wählt, sucht nur
	// dort.
	if frage.Field != "" && frage.Field != domain.Markenfeld {
		return nil, nil
	}

	gesucht := domain.Markennormalform(frage.Text)
	if gesucht == "" {
		return nil, nil
	}

	profilID, profilfehler := ProfilAus(ctx)
	if profilfehler != nil {
		return nil, profilfehler
	}

	return l.markenzeilen(ctx, markensuchen,
		profilID, frage.SetID, frage.SetID, gesucht, domain.SearchLimit+1)
}

// markenzeilen führt eine Abfrage aus, die Markentreffer gibt.
//
// **Die Abfrage kommt als Parameter** — dieselbe Bauform wie
// `Sammlungsspeicher.lesen`: der `Scan`-Fehlerweg lässt sich sonst von
// niemandem herbeiführen, und dann fiele es nicht auf, wenn seine Meldung
// verloren ginge.
func (l *SetLeser) markenzeilen(
	ctx context.Context, abfrage string, werte ...any,
) ([]domain.SearchHit, error) {
	zeilen, err := l.db.QueryContext(ctx, abfrage, werte...)
	if err != nil {
		return nil, uebersetze(err, "in den Marken suchen")
	}

	defer func() { _ = zeilen.Close() }()

	treffer := []domain.SearchHit{}

	for zeilen.Next() {
		fund := domain.SearchHit{Field: domain.Markenfeld}
		if err := zeilen.Scan(&fund.CardID, &fund.SetID, &fund.Number, &fund.Value); err != nil {
			return nil, uebersetze(err, "einen Markentreffer lesen")
		}

		treffer = append(treffer, fund)
	}

	return treffer, uebersetze(zeilen.Err(), "in den Marken suchen")
}
