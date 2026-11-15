package storage

// **Die zwei Hälften der Suche** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Eine eigene Datei
//
// `retrievalindex.go` stand mit ihnen über der Zeilengrenze aus §2, und der
// Schnitt folgt der Frage: dort steht, **wie ein Index entsteht** — hier, wie
// er gelesen wird.
//
// # Die Sichtprüfung steht in der Abfrage und nicht dahinter
//
// **Das Issue verlangt sie zweimal**: „ACL bereits in Kandidatenauswahl und
// erneut vor Ergebnis-/Snippet-/Kontextausgabe prüfen." Der erste Ort ist
// die `WHERE`-Zeile; der zweite ist [sichtbar], durch das jede Zeile geht,
// bevor sie in die Liste kommt.
//
// **Und die Zahl zählt mit.** „Fremde Treffer dürfen auch nicht in Anzahl,
// Cache, Debugdaten oder Vorschlägen auftauchen." Deshalb wird gefiltert,
// bevor gezählt wird, und nicht danach.

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// sichtbedingung ist die Zeile, die fremde Dokumente ausschließt.
//
// **Sie steht als Konstante da und wird nicht je Abfrage gebaut.** Zwei
// Fassungen liefen auseinander, und die eine, die vergisst, fällt nicht auf.
const sichtbedingung = `
	AND (d.sichtbarkeit = 'oeffentlich' OR d.profil_id = ?)
	AND NOT EXISTS (
		SELECT 1 FROM retrieval_gesperrt g
		WHERE g.kennung = d.kennung AND g.profil_id IN (0, ?))`

// dokumentspalten sind die Spalten eines Dokuments, in fester Reihenfolge.
const dokumentspalten = `
	d.kennung, d.quelle, d.bezugsadresse, d.spiel, d.kennungssystem,
	d.fachkennung, d.semantik, d.regelnummer, d.urheber, d.sprache,
	d.titel, d.text, d.quellrevision, d.quelldatum, d.abgerufen,
	d.inhaltshash, d.merkmale, d.abschnitt, d.von, d.bis,
	d.profil_id, d.sichtbarkeit`

// Lexikalisch gibt die Treffer der Volltextsuche.
//
// **Über FTS5 und damit über BM25** — die Ordnung kommt aus `rank`, und was
// herauskommt, sind **Ränge** und keine Punkte (siehe [domain.Retrievaltreffer]).
func (r *Retrievalindex) Lexikalisch(ctx context.Context, frage string, grenze int) (
	[]domain.Retrievaltreffer, error,
) {
	generation, profil, fehler := r.suchgrundlage(ctx)
	if fehler != nil {
		return nil, fehler
	}

	if generation == 0 {
		return []domain.Retrievaltreffer{}, nil
	}

	muster := fts5Retrievalmuster(frage)
	if muster == "" {
		return []domain.Retrievaltreffer{}, nil
	}

	zeilen, fehler := r.db.QueryContext(ctx, `
		SELECT `+dokumentspalten+`,
		       snippet(retrieval_volltext, 1, '', '', '…', 12)
		FROM retrieval_volltext v
		JOIN retrieval_dokumente d ON d.id = v.dokument_id
		WHERE retrieval_volltext MATCH ?
		  AND d.generation = ?`+sichtbedingung+`
		ORDER BY rank, d.kennung
		LIMIT ?`, muster, generation, profil, profil, grenze)
	if fehler != nil {
		return nil, uebersetze(fehler, "den Suchindex lexikalisch lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird unten

	return treffernAus(zeilen, profil)
}

// Semantisch gibt die Treffer des Vektorvergleichs.
//
// # Warum ein voller Durchlauf und kein Vektorindex
//
// **Weil die Größe es hergibt.** Der Vermerk zu A15 führt die Messung; was
// hier steht, ist die Folge: ein `SELECT` über die sichtbaren Zeilen, eine
// Rechnung je Zeile, eine Ordnung. Ein Vektorindex kostete eine
// Abhängigkeit, ein Dateiformat und eine zweite Stelle, an der
// Berechtigungen hängen — für eine Antwortzeit, die schon jetzt unter dem
// liegt, was ein Mensch bemerkt.
func (r *Retrievalindex) Semantisch(ctx context.Context, frage []float32, grenze int) (
	[]domain.Retrievaltreffer, error,
) {
	generation, profil, fehler := r.suchgrundlage(ctx)
	if fehler != nil {
		return nil, fehler
	}

	if generation == 0 || len(frage) == 0 {
		return []domain.Retrievaltreffer{}, nil
	}

	zeilen, fehler := r.db.QueryContext(ctx, `
		SELECT `+dokumentspalten+`, k.vektor
		FROM retrieval_dokumente d
		JOIN retrieval_vektoren k ON k.dokument_id = d.id
		WHERE d.generation = ?`+sichtbedingung+`
		ORDER BY d.kennung`, generation, profil, profil)
	if fehler != nil {
		return nil, uebersetze(fehler, "den Suchindex semantisch lesen")
	}

	defer func() { _ = zeilen.Close() }() //nolint:errcheck // gelesen wird unten

	return naechsteAus(zeilen, frage, profil, grenze)
}

// suchgrundlage gibt Generation und Profil — Generation 0 heißt: keine.
func (r *Retrievalindex) suchgrundlage(ctx context.Context) (
	generation, profil int64, fehler error,
) {
	profil, fehler = ProfilAus(ctx)
	if fehler != nil {
		return 0, 0, fehler
	}

	generation, fehler = r.aktiveGeneration(ctx)
	if fehler != nil {
		// **Kein aktiver Index ist eine Auskunft und kein Fehler** (§6):
		// wer nie gebaut hat, findet nichts.
		if errors.Is(fehler, domain.ErrNotFound) {
			return 0, profil, nil
		}

		return 0, 0, fehler
	}

	return generation, profil, nil
}

// schreibeDokument legt Dokument, Volltext und Vektor in einer Transaktion an.
func schreibeDokument(
	ctx context.Context, tx *sql.Tx, generation int64,
	dokument domain.Retrievaldokument, vektor []float32,
) error {
	ergebnis, fehler := tx.ExecContext(ctx, `
		INSERT INTO retrieval_dokumente
			(generation, kennung, quelle, bezugsadresse, spiel, kennungssystem,
			 fachkennung, semantik, regelnummer, urheber, sprache, titel, text,
			 quellrevision, quelldatum, abgerufen, inhaltshash, merkmale,
			 abschnitt, von, bis, profil_id, sichtbarkeit)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		generation, dokument.Kennung, string(dokument.Quelle),
		dokument.Bezugsadresse, dokument.Spiel, string(dokument.Kennungssystem),
		dokument.Fachkennung, string(dokument.Semantik), dokument.Regelnummer,
		string(dokument.Urheber), dokument.Sprache, dokument.Titel, dokument.Text,
		dokument.Quellrevision, dokument.Quelldatum,
		dokument.Abgerufen.UTC().Format(time.RFC3339Nano), dokument.Inhaltshash,
		alsMerkmalsjson(dokument.Merkmale), dokument.Abschnitt, dokument.Von, dokument.Bis, dokument.ProfilID,
		string(dokument.Sichtbarkeit))
	if fehler != nil {
		return uebersetze(fehler, "das Dokument %q schreiben", dokument.Kennung)
	}

	kennung, zaehlfehler := eingefuegt(ergebnis)

	// **Alle drei Fehler werden verbunden.** Drei `if` wären drei Zweige,
	// von denen der erste nie betreten wird: `LastInsertId` scheitert bei
	// diesem Treiber nicht, und was die eine Anweisung scheitern lässt, lässt
	// auch die nächste scheitern (§10).
	_, volltext := tx.ExecContext(ctx, `
		INSERT INTO retrieval_volltext (titel, text, dokument_id)
		VALUES (?, ?, ?)`, dokument.Titel, dokument.Text, kennung)

	_, vektorfehler := tx.ExecContext(ctx, `
		INSERT INTO retrieval_vektoren (dokument_id, vektor) VALUES (?, ?)`,
		kennung, alsBlob(vektor))

	return uebersetze(errors.Join(zaehlfehler, volltext, vektorfehler),
		"das Dokument %q in den Index legen", dokument.Kennung)
}

// fts5Retrievalmuster baut das Suchmuster — dieselbe Regel wie in `suche.go`.
//
// **Eine eigene Zeile und kein zweiter Regelsatz**: sie ruft dieselbe
// Funktion, damit die Volltextsuche dieses Programms **eine** Deutung hat.
func fts5Retrievalmuster(eingabe string) string {
	return fts5Muster(strings.TrimSpace(eingabe))
}
