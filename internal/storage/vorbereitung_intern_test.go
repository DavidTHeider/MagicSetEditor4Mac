// Die Vorbereitung vor der Migration —
// #939.
//
// **Eine eigene Datei, weil sie eine eigene Frage beantwortet** (§2):
// `fehlerpfade_test.go` prüft, was der Migrator meldet; hier steht, **wann** er
// den Aufrufer vorher zu Wort kommen lässt. Sie ging über die Zeilengrenze, als
// diese Fälle dazukamen.
//
// **Paketintern**, aus demselben Grund wie die Fehlerpfade: mit der
// eingebetteten Quelle sind diese Wege nicht erreichbar, weil sie richtig ist.
package storage

import (
	"context"
	"errors"
	"testing"
	"testing/fstest"
)

// **Vor einer ausstehenden Migration wird vorbereitet** —
// #939.
//
// Der Aufrufer legt dort seine Sicherung an. `storage` weiß, **ob** etwas
// aussteht; **was** dann zu tun ist, weiß die Anwendung.
func TestVorEinerAusstehendenMigrationWirdVorbereitet(t *testing.T) {
	db := offeneDatenbank(t)

	// **Erst einen Stand anlegen** — ein frischer Bestand hat nichts zu
	// sichern, und das prüft der Fall darunter.
	erste := fstest.MapFS{
		"migrationen/0001_gut.sql": {Data: []byte("CREATE TABLE eins (a TEXT);")},
	}
	if err := migriereAus(probekontext(t), db, erste, nil); err != nil {
		t.Fatalf("Grundstand: %v", err)
	}

	quelle := fstest.MapFS{
		"migrationen/0001_gut.sql": {Data: []byte("CREATE TABLE eins (a TEXT);")},
		"migrationen/0002_neu.sql": {Data: []byte("CREATE TABLE zwei (a TEXT);")},
	}

	gerufen := 0
	vorbereiten := func(context.Context) error { gerufen++; return nil }

	if err := migriereAusMitVorbereitung(probekontext(t), db, quelle, nil, vorbereiten, nil); err != nil {
		t.Fatalf("erster Lauf: %v", err)
	}

	if gerufen != 1 {
		t.Errorf("%d Aufrufe, erwartet einer", gerufen)
	}

	// **Und beim zweiten Lauf nicht** — es steht nichts mehr aus.
	if err := migriereAusMitVorbereitung(probekontext(t), db, quelle, nil, vorbereiten, nil); err != nil {
		t.Fatalf("zweiter Lauf: %v", err)
	}

	if gerufen != 1 {
		t.Errorf("%d Aufrufe — ohne ausstehende Migration darf nichts geschehen", gerufen)
	}
}

// **Scheitert die Vorbereitung, wird nicht migriert.**
//
// Eine Migration ohne Sicherung ist genau der Fall, für den es die Sicherung
// gibt. Wer hier weiterliefe, hätte sie ebenso gut weglassen können.
func TestOhneGeglueckteVorbereitungWirdNichtMigriert(t *testing.T) {
	db := offeneDatenbank(t)

	erste := fstest.MapFS{
		"migrationen/0001_gut.sql": {Data: []byte("CREATE TABLE eins (a TEXT);")},
	}
	if err := migriereAus(probekontext(t), db, erste, nil); err != nil {
		t.Fatalf("Grundstand: %v", err)
	}

	quelle := fstest.MapFS{
		"migrationen/0001_gut.sql": {Data: []byte("CREATE TABLE eins (a TEXT);")},
		"migrationen/0002_neu.sql": {Data: []byte("CREATE TABLE zwei (a TEXT);")},
	}

	scheitern := func(context.Context) error { return errGestolpert }

	if err := migriereAusMitVorbereitung(
		probekontext(t), db, quelle, nil, scheitern, nil); !errors.Is(err, errGestolpert) {
		t.Fatalf("Fehler %v, erwartet war die gescheiterte Vorbereitung", err)
	}

	verbucht, fehler := AngewendeteMigrationen(probekontext(t), db)
	if fehler != nil {
		t.Fatalf("Buchführung nicht lesbar: %v", fehler)
	}

	if len(verbucht) != 1 {
		t.Errorf("nach der gescheiterten Vorbereitung ist %v verbucht — "+
			"erwartet war nur der Grundstand", verbucht)
	}
}

// **Ein frischer Bestand wird nicht vorbereitet** —
// #939.
//
// Beim allerersten Start stehen **alle** Migrationen aus, und es gibt nichts
// zu sichern: die Datenbank ist eben erst entstanden. Eine Kopie davon wäre
// eine Datei, die niemand je braucht — und sie trüge einen Namen, der einen
// Rückweg verspricht, den es nicht gibt.
func TestEinFrischerBestandWirdNichtVorbereitet(t *testing.T) {
	db := offeneDatenbank(t)

	quelle := fstest.MapFS{
		"migrationen/0001_gut.sql": {Data: []byte("CREATE TABLE eins (a TEXT);")},
	}

	gerufen := 0
	vorbereiten := func(context.Context) error { gerufen++; return nil }

	if err := migriereAusMitVorbereitung(probekontext(t), db, quelle, nil, vorbereiten, nil); err != nil {
		t.Fatalf("erster Lauf: %v", err)
	}

	if gerufen != 0 {
		t.Errorf("%d Aufrufe — ein frischer Bestand hat nichts zu sichern", gerufen)
	}

	// **Und die zweite Migration wird vorbereitet** — jetzt gibt es einen Stand.
	zweite := fstest.MapFS{
		"migrationen/0001_gut.sql": {Data: []byte("CREATE TABLE eins (a TEXT);")},
		"migrationen/0002_neu.sql": {Data: []byte("CREATE TABLE zwei (a TEXT);")},
	}

	if err := migriereAusMitVorbereitung(probekontext(t), db, zweite, nil, vorbereiten, nil); err != nil {
		t.Fatalf("zweiter Lauf: %v", err)
	}

	if gerufen != 1 {
		t.Errorf("%d Aufrufe, erwartet einer — hier gibt es einen Stand zu sichern", gerufen)
	}
}
