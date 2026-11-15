package storage

import (
	"context"
	"database/sql"
	"slices"
	"strings"
	"testing"
)

// **Das Schema, wie der Bestand es selbst beschreibt** — die Grundlage der
// Wächter in `invarianten_intern_test.go` (ENG-05.AK-01,
// #1730).
//
// **Aus einem frisch migrierten Bestand gelesen und nicht aus den
// Wanderungen.** `sammlung` ist fünfmal umgebaut worden; was am Ende gilt,
// weiß nur der Bestand selbst — `pragma_table_info`,
// `pragma_foreign_key_list`, `pragma_index_list`.

// schemaspalte ist eine Spalte einer Tafel.
type schemaspalte struct {
	tafel, name, typ string
	pflicht          bool
}

// schemaverweis ist ein Fremdschlüssel — bei mehreren Spalten einer.
type schemaverweis struct {
	tafel, ziel, beimLoeschen string
	spalten                   []string
}

// schemabild ist alles, was die Wächter lesen.
type schemabild struct {
	spalten    []schemaspalte
	verweise   []schemaverweis
	indizes    map[string][][]string // Tafel → Spaltenfolgen ihrer Indizes und ihres Schlüssels
	definition map[string]string     // Tafel → CREATE TABLE
	ausloeser  map[string]string     // Auslöser → CREATE TRIGGER
	tafelVon   map[string]string     // Auslöser → Tafel
}

// schemaAus liest das Bild eines frisch migrierten Bestands.
func schemaAus(t *testing.T) schemabild {
	t.Helper()

	db := migrierteDatenbank(t)
	ctx := probekontext(t)

	bild := schemabild{
		indizes:    map[string][][]string{},
		definition: map[string]string{},
		ausloeser:  map[string]string{},
		tafelVon:   map[string]string{},
	}

	tafeln := eigeneTafeln(t, ctx, db)
	for _, tafel := range tafeln {
		bild.definition[tafel] = einWert(t, ctx, db,
			`SELECT sql FROM sqlite_master WHERE type = 'table' AND name = ?`, tafel)
		bild.spalten = append(bild.spalten, spaltenVon(t, ctx, db, tafel)...)
		bild.verweise = append(bild.verweise, verweiseVon(t, ctx, db, tafel)...)
		bild.indizes[tafel] = indizesVon(t, ctx, db, tafel)
	}

	for _, zeile := range zeilenAus(t, ctx, db,
		`SELECT name, tbl_name, sql FROM sqlite_master WHERE type = 'trigger'`) {
		bild.ausloeser[zeile[0]] = zeile[2]
		bild.tafelVon[zeile[0]] = zeile[1]
	}

	return bild
}

// eigeneTafeln sind die Tafeln des Bestands ohne die Schattentafeln des
// Volltextindex — die legt FTS5 selbst an, und keine Regel dieses Projekts
// gilt für sie.
func eigeneTafeln(t *testing.T, ctx context.Context, db *sql.DB) []string {
	t.Helper()

	virtuell := map[string]bool{}
	for _, zeile := range zeilenAus(t, ctx, db,
		`SELECT name FROM sqlite_master WHERE type = 'table' AND sql LIKE 'CREATE VIRTUAL%'`) {
		virtuell[zeile[0]] = true
	}

	var tafeln []string

	for _, zeile := range zeilenAus(t, ctx, db,
		`SELECT name FROM sqlite_master
		  WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name`) {
		name := zeile[0]
		if virtuell[name] || schattentafel(name, virtuell) {
			continue
		}

		tafeln = append(tafeln, name)
	}

	return tafeln
}

func schattentafel(name string, virtuell map[string]bool) bool {
	for _, endung := range []string{"_config", "_content", "_data", "_docsize", "_idx"} {
		if basis, ist := strings.CutSuffix(name, endung); ist && virtuell[basis] {
			return true
		}
	}

	return false
}

func spaltenVon(t *testing.T, ctx context.Context, db *sql.DB, tafel string) []schemaspalte {
	t.Helper()

	zeilen := zeilenAus(t, ctx, db,
		`SELECT name, type, "notnull" FROM pragma_table_info(?) ORDER BY cid`, tafel)

	spalten := make([]schemaspalte, 0, len(zeilen))
	for _, zeile := range zeilen {
		spalten = append(spalten, schemaspalte{
			tafel: tafel, name: zeile[0], typ: zeile[1], pflicht: zeile[2] == "1",
		})
	}

	return spalten
}

func verweiseVon(t *testing.T, ctx context.Context, db *sql.DB, tafel string) []schemaverweis {
	t.Helper()

	nachKennung := map[string]*schemaverweis{}

	var folge []string

	for _, zeile := range zeilenAus(t, ctx, db,
		`SELECT id, "table", "from", on_delete FROM pragma_foreign_key_list(?) ORDER BY id, seq`, tafel) {
		verweis, gibtEs := nachKennung[zeile[0]]
		if !gibtEs {
			verweis = &schemaverweis{tafel: tafel, ziel: zeile[1], beimLoeschen: zeile[3]}
			nachKennung[zeile[0]] = verweis
			folge = append(folge, zeile[0])
		}

		verweis.spalten = append(verweis.spalten, zeile[2])
	}

	verweise := make([]schemaverweis, 0, len(folge))
	for _, kennung := range folge {
		verweise = append(verweise, *nachKennung[kennung])
	}

	return verweise
}

// indizesVon gibt die Spaltenfolgen jedes Index und des Primärschlüssels.
//
// **Der Primärschlüssel gehört dazu**, auch ohne eigenen Index: bei einer
// `INTEGER PRIMARY KEY` ist die Tafel selbst nach ihm geordnet, und bei
// `WITHOUT ROWID` ebenso.
func indizesVon(t *testing.T, ctx context.Context, db *sql.DB, tafel string) [][]string {
	t.Helper()

	var folgen [][]string

	for _, zeile := range zeilenAus(t, ctx, db,
		`SELECT l.name, group_concat(i.name, ',' ORDER BY i.seqno)
		   FROM pragma_index_list(?) AS l, pragma_index_info(l.name) AS i
		  GROUP BY l.name`, tafel) {
		folgen = append(folgen, strings.Split(zeile[1], ","))
	}

	schluesselzeilen := zeilenAus(t, ctx, db,
		`SELECT name FROM pragma_table_info(?) WHERE pk > 0 ORDER BY pk`, tafel)

	schluessel := make([]string, 0, len(schluesselzeilen))
	for _, zeile := range schluesselzeilen {
		schluessel = append(schluessel, zeile[0])
	}

	if len(schluessel) > 0 {
		folgen = append(folgen, schluessel)
	}

	return folgen
}

// beginntMit sagt, ob eine Spaltenfolge mit einer anderen anfängt.
func beginntMit(folge, anfang []string) bool {
	return len(folge) >= len(anfang) && slices.Equal(folge[:len(anfang)], anfang)
}

func einWert(t *testing.T, ctx context.Context, db *sql.DB, abfrage string, werte ...any) string {
	t.Helper()

	zeilen := zeilenAus(t, ctx, db, abfrage, werte...)
	if len(zeilen) != 1 || len(zeilen[0]) != 1 {
		t.Fatalf("%s: %d Zeilen statt einer", abfrage, len(zeilen))
	}

	return zeilen[0][0]
}

// zeilenAus liest jede Zeile als Zeichenketten.
func zeilenAus(t *testing.T, ctx context.Context, db *sql.DB, abfrage string, werte ...any) [][]string {
	t.Helper()

	zeilen, fehler := db.QueryContext(ctx, abfrage, werte...)
	if fehler != nil {
		t.Fatalf("%s: %v", abfrage, fehler)
	}

	defer func() { _ = zeilen.Close() }()

	spalten, fehler := zeilen.Columns()
	if fehler != nil {
		t.Fatalf("%s: %v", abfrage, fehler)
	}

	var alle [][]string

	for zeilen.Next() {
		werteDerZeile := make([]sql.NullString, len(spalten))
		ziele := make([]any, len(spalten))

		for i := range werteDerZeile {
			ziele[i] = &werteDerZeile[i]
		}

		if fehler := zeilen.Scan(ziele...); fehler != nil {
			t.Fatalf("%s: %v", abfrage, fehler)
		}

		zeile := make([]string, len(spalten))
		for i, wert := range werteDerZeile {
			zeile[i] = wert.String
		}

		alle = append(alle, zeile)
	}

	if fehler := zeilen.Err(); fehler != nil {
		t.Fatalf("%s: %v", abfrage, fehler)
	}

	return alle
}
