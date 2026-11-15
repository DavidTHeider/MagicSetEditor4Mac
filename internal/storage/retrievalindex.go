package storage

// **Der lokale Retrievalindex** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Generationen, und warum
//
// **Ein Neubau darf nicht mitten im Lesen sichtbar werden.** Eine Generation
// entsteht im Zustand `baut`, wird gefüllt, und erst [Retrievalindex.Aktivieren]
// macht sie zu der, gegen die gesucht wird — in **einer** Anweisung, die die
// vorige zugleich ablöst.
//
// # Und die Sperre steht über allem
//
// **Ein Widerruf schlägt den letzten guten Index.** Die Tombstones in
// `retrieval_gesperrt` gelten generationenübergreifend: was dort steht, kommt
// nicht heraus, auch wenn ein frischer Import es wieder hineinschreibt. Wer
// nur löschte, hätte es nach der nächsten Rücksicherung wieder.

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Retrievalindex liest und schreibt den lokalen Suchindex.
type Retrievalindex struct{ db *sql.DB }

// NeuerRetrievalindex baut ihn über einer Datenbank.
func NeuerRetrievalindex(db *sql.DB) *Retrievalindex { return &Retrievalindex{db: db} }

// Anlegen beginnt eine neue Generation und gibt ihre Nummer.
func (r *Retrievalindex) Anlegen(ctx context.Context, manifest domain.Indexmanifest) (
	int64, error,
) {
	if fehler := manifest.Pruefe(); fehler != nil {
		return 0, fehler
	}

	// **Ohne Fehlerprüfung**: `json.Marshal` scheitert an Kanälen, Funktionen
	// und Zyklen. [domain.Indexmanifest] hat nichts davon — ein Zweig dafür
	// wäre einer, den kein Lauf betreten kann (§10).
	roh, _ := json.Marshal(manifest) //nolint:errcheck,errchkjson // siehe oben

	// **Der Raumabdruck steht neben dem Fingerabdruck**, damit die Frage
	// „darf dieser Encoder hier suchen?" ohne das JSON zu beantworten ist.
	ergebnis, fehler := r.db.ExecContext(ctx, `
		INSERT INTO retrieval_generationen
			(manifest, fingerabdruck, raumabdruck, gebaut, zustand, aktiv)
		VALUES (?, ?, ?, ?, 'baut', 0)`,
		string(roh), manifest.Fingerabdruck(), raumabdruckVon(manifest),
		time.Now().UTC().Format(time.RFC3339Nano))
	if fehler != nil {
		return 0, uebersetze(fehler, "eine Indexgeneration anlegen")
	}

	return eingefuegt(ergebnis)
}

// Schreiben legt ein Dokument samt Vektor in eine Generation.
func (r *Retrievalindex) Schreiben(
	ctx context.Context, generation int64,
	dokument domain.Retrievaldokument, vektor []float32,
) error {
	if fehler := dokument.Pruefe(); fehler != nil {
		return fehler
	}

	manifest, fehler := r.manifestVon(ctx, generation)
	if fehler != nil {
		return fehler
	}

	if len(vektor) != manifest.Vertrag.Dimension {
		return fmt.Errorf(
			"%w: der Vektor hat %d Zahlen, das Manifest nennt %d",
			domain.ErrValidation, len(vektor), manifest.Vertrag.Dimension)
	}

	return imRetrievalzug(ctx, r.db, func(tx *sql.Tx) error {
		return schreibeDokument(ctx, tx, generation, dokument, vektor)
	})
}

// Aktivieren macht eine fertige Generation zu der, gegen die gesucht wird.
//
// **In einem Zug**: die vorige wird abgelöst und die neue gesetzt, oder
// keines von beidem.
func (r *Retrievalindex) Aktivieren(ctx context.Context, generation int64) error {
	return imRetrievalzug(ctx, r.db, func(tx *sql.Tx) error {
		// **Erst nachsehen, dann umschalten.** Eine Generation, die es nicht
		// gibt, soll das als Meldung bekommen — und nicht als „null Zeilen
		// geändert", was auch „schon aktiv" heißen könnte.
		var zustand string

		if fehler := tx.QueryRowContext(ctx,
			`SELECT zustand FROM retrieval_generationen WHERE generation = ?`,
			generation).Scan(&zustand); fehler != nil {
			if errors.Is(fehler, sql.ErrNoRows) {
				return fmt.Errorf("%w: die Indexgeneration %d gibt es nicht",
					domain.ErrNotFound, generation)
			}

			return uebersetze(fehler, "die Generation %d lesen", generation)
		}

		// **Beide Anweisungen laufen, und die Fehler werden verbunden.**
		//
		// Zwei `if` wären hier zwei Zweige, von denen nur der erste je
		// betreten wird: beide gehen auf dieselbe Tafel, und was den einen
		// scheitern lässt, lässt auch den anderen scheitern. Ein Zweig, den
		// kein Lauf betreten kann, sähe aus wie eine geprüfte Zusage (§10).
		_, abgeloest := tx.ExecContext(ctx,
			`UPDATE retrieval_generationen SET aktiv = 0 WHERE aktiv = 1`)

		_, gesetzt := tx.ExecContext(ctx, `
			UPDATE retrieval_generationen
			SET aktiv = 1, zustand = 'fertig'
			WHERE generation = ?`, generation)

		return uebersetze(errors.Join(abgeloest, gesetzt),
			"die Generation %d aktivieren", generation)
	})
}

// Verwerfen räumt eine begonnene Generation weg.
func (r *Retrievalindex) Verwerfen(ctx context.Context, generation int64) error {
	if _, fehler := r.db.ExecContext(ctx,
		`DELETE FROM retrieval_generationen WHERE generation = ? AND aktiv = 0`,
		generation); fehler != nil {
		return uebersetze(fehler, "die Generation %d verwerfen", generation)
	}

	return nil
}

// Sperren hält eine Kennung dauerhaft aus den Ergebnissen.
//
// **Profil 0 heißt: für alle.** Eine Sperre je Profil trifft nur dessen
// eigene Dokumente; eine für alle trifft auch die öffentlichen.
func (r *Retrievalindex) Sperren(ctx context.Context, kennung string, profil int64) error {
	if _, fehler := r.db.ExecContext(ctx, `
		INSERT INTO retrieval_gesperrt (kennung, profil_id, gesperrt)
		VALUES (?, ?, ?)
		ON CONFLICT (kennung, profil_id) DO UPDATE SET gesperrt = excluded.gesperrt`,
		kennung, profil, time.Now().UTC().Format(time.RFC3339Nano)); fehler != nil {
		return uebersetze(fehler, "die Kennung %q sperren", kennung)
	}

	return nil
}

// Manifest gibt das Manifest der aktiven Generation.
func (r *Retrievalindex) Manifest(ctx context.Context) (domain.Indexmanifest, error) {
	generation, fehler := r.aktiveGeneration(ctx)
	if fehler != nil {
		return domain.Indexmanifest{}, fehler
	}

	return r.manifestVon(ctx, generation)
}

// aktiveGeneration gibt die Nummer der aktiven — oder ErrNotFound.
func (r *Retrievalindex) aktiveGeneration(ctx context.Context) (int64, error) {
	return aktiveGenerationIn(ctx, r.db)
}

// zeilenabfrager ist, was eine einzelne Zeile lesen kann — eine Verbindung
// **oder** ein Zug.
//
// # Warum das nötig ist
//
// **Weil ein Lauf neben der eigenen Transaktion hängen bleibt.** SQLite hält
// beim Schreiben eine Verbindung; wer währenddessen über `*sql.DB` liest,
// bittet um eine zweite und wartet auf eine, die er selbst blockiert.
//
// **Gemessen am 2026-09-12**: `Uebernehmen` tat genau das, und der Paketlauf
// endete nach zehn Minuten in seiner eigenen Frist — kein Fehler, keine
// Meldung, nur Stillstand. Dieselbe Lage wie bei [zeilenfrager] nebenan, nur
// für eine Zeile statt für viele.
type zeilenabfrager interface {
	QueryRowContext(ctx context.Context, abfrage string, werte ...any) *sql.Row
}

// aktiveGenerationIn liest sie über den gegebenen Leser.
func aktiveGenerationIn(ctx context.Context, leser zeilenabfrager) (int64, error) {
	var generation int64

	fehler := leser.QueryRowContext(ctx,
		`SELECT generation FROM retrieval_generationen WHERE aktiv = 1`).
		Scan(&generation)
	if fehler != nil {
		if fehler == sql.ErrNoRows {
			return 0, fmt.Errorf(
				"%w: es ist kein Suchindex aktiv", domain.ErrNotFound)
		}

		return 0, uebersetze(fehler, "die aktive Indexgeneration lesen")
	}

	return generation, nil
}

// manifestVon liest das Manifest einer Generation.
func (r *Retrievalindex) manifestVon(ctx context.Context, generation int64) (
	domain.Indexmanifest, error,
) {
	var roh string

	fehler := r.db.QueryRowContext(ctx,
		`SELECT manifest FROM retrieval_generationen WHERE generation = ?`,
		generation).Scan(&roh)
	if fehler != nil {
		if fehler == sql.ErrNoRows {
			return domain.Indexmanifest{}, fmt.Errorf(
				"%w: die Indexgeneration %d gibt es nicht",
				domain.ErrNotFound, generation)
		}

		return domain.Indexmanifest{}, uebersetze(fehler,
			"das Manifest der Generation %d lesen", generation)
	}

	var manifest domain.Indexmanifest

	if fehler := json.Unmarshal([]byte(roh), &manifest); fehler != nil {
		return domain.Indexmanifest{}, domain.MitGrund(domain.ErrValidation,
			fehler, "das Manifest der Generation %d lesen", generation)
	}

	return manifest, nil
}

// raumabdruckVon gibt den Abdruck des Vektorraums allein.
//
// **Über ein Manifest ohne Suchkonfiguration und Reranker**: was den Raum
// bestimmt, steht in [domain.Indexmanifest.PasstZu], und derselbe Abdruck
// entsteht hier über denselben Weg.
func raumabdruckVon(manifest domain.Indexmanifest) string {
	nur := domain.Indexmanifest{
		Schemafassung: manifest.Schemafassung, Indexfassung: manifest.Indexfassung,
		Zerteilfassung: manifest.Zerteilfassung,
		Modell:         manifest.Modell, Modellrevision: manifest.Modellrevision,
		Artefakthash: manifest.Artefakthash, Laufzeit: manifest.Laufzeit,
		Abstandsmass: manifest.Abstandsmass, Vertrag: manifest.Vertrag,
		Suchkonfiguration: "", Rerankerabdruck: "",
	}

	return nur.Fingerabdruck()
}

// imRetrievalzug fährt eine Arbeit in einer Transaktion.
//
// **Ein eigener Helfer und kein geteilter**: die Ablagen dieses Pakets
// öffnen ihre Transaktionen je Weg, weil jede eine andere Rücknahme braucht.
// Hier ist sie einfach — alles oder nichts.
//
// **Zurückgerollt wird im `defer`**, auch nach einem Zusammenbruch der Arbeit
// (BK-04, #2129): hier stand das Zurückrollen im Fehlerzweig, und ein `panic`
// ließ die Transaktion samt Verbindung und Schreibsperre stehen.
func imRetrievalzug(ctx context.Context, db *sql.DB, arbeit func(*sql.Tx) error) (fehler error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return uebersetze(err, "den Suchindex öffnen")
	}

	defer func() { fehler = zurueckgenommen(tx, fehler) }()

	if err := arbeit(tx); err != nil {
		return err
	}

	// **Der Fehler steht im Rückgabewert und nicht in einem `if`** — dieselbe
	// Gestalt wie bei [betroffen]: ein Zweig, den kein Lauf betreten kann,
	// sähe aus wie eine geprüfte Zusage (§10).
	return uebersetze(tx.Commit(), "den Suchindex festschreiben")
}

// cosinus gibt die Ähnlichkeit zweier Vektoren — null bei Nulllänge.
//
// **Ohne Normalisierungsannahme.** Der Vertrag sagt, ob die Vektoren
// normalisiert sind; diese Rechnung gilt so oder so, und sie kostet zwei
// Multiplikationen mehr.
func cosinus(links, rechts []float32) float64 {
	var produkt, linksLaenge, rechtsLaenge float64

	for i := range links {
		produkt += float64(links[i]) * float64(rechts[i])
		linksLaenge += float64(links[i]) * float64(links[i])
		rechtsLaenge += float64(rechts[i]) * float64(rechts[i])
	}

	if linksLaenge == 0 || rechtsLaenge == 0 {
		return 0
	}

	return produkt / (math.Sqrt(linksLaenge) * math.Sqrt(rechtsLaenge))
}
