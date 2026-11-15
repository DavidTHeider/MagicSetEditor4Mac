// **Die Fehlerwege der Sammlung** — Werkstück 2 und 3 von Phase 3.4
// (#1165).
//
// # Warum ein weißer Lauf
//
// Eine geschlossene Datenbank ist der einzige Weg, jeden Zweig zu betreten, an
// dem SQLite scheitert — und `geschlosseneDatenbank` steht im Paket. Was hier
// geprüft wird, ist **dass die Meldung ankommt**, nicht welche.
package storage

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// zuExemplar gibt ein Exemplar für die Fehlerwege.
func zuExemplar() *domain.Exemplar {
	return &domain.Exemplar{
		ID: 1, Printing: "x", Menge: 1,
		Zustand: domain.ZustandNearMint, Sprache: "en",
	}
}

// **Auf einer geschlossenen Datenbank meldet jeder Weg der Sammlung.**
func TestDieSammlungMeldetEineGeschlosseneDatenbank(t *testing.T) {
	speicher := NeuerSammlungsspeicher(geschlosseneDatenbank(t))

	for name, versuch := range map[string]func() error{
		"Anlegen": func() error {
			_, err := speicher.Anlegen(probekontext(t), zuExemplar())

			return err
		},
		"Aendern":  func() error { return speicher.Aendern(probekontext(t), zuExemplar()) },
		"Loeschen": func() error { return speicher.Loeschen(probekontext(t), 1, 1) },
		"Liste": func() error {
			_, err := speicher.Liste(probekontext(t))

			return err
		},
		"Fuer": func() error {
			_, err := speicher.Fuer(probekontext(t), "x")

			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := versuch(); err == nil {
				t.Error("die geschlossene Datenbank nahm den Aufruf an")
			}
		})
	}
}

// **Und jeder Weg der Preisstände ebenso.**
func TestDiePreisstaendeMeldenEineGeschlosseneDatenbank(t *testing.T) {
	speicher := NeuerPreisspeicher(geschlosseneDatenbank(t))

	stand := domain.Preisstand{
		Printing: "x", Quelle: domain.PreisquelleScryfall,
		Waehrung: "EUR", Finish: domain.FinishNonfoil,
		Wert: 1, Erhoben: time.Unix(0, 0),
	}

	for name, versuch := range map[string]func() error{
		"Eintragen": func() error { return speicher.Eintragen(probekontext(t), stand) },
		"Neueste": func() error {
			_, err := speicher.Neueste(probekontext(t), "x")

			return err
		},
		"Verlauf": func() error {
			_, err := speicher.Verlauf(probekontext(t), "x",
				reihevonScryfall("EUR", domain.FinishNonfoil))

			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := versuch(); err == nil {
				t.Error("die geschlossene Datenbank nahm den Aufruf an")
			}
		})
	}
}

// **Ein `Scan` gegen die falsche Spaltenzahl meldet sich.**
//
// # Warum dieser Weg überhaupt geprüft wird
//
// Er ist der, den niemand herbeiführen kann, solange Abfrage und `Scan`
// zusammenpassen — und genau deshalb fiele es nicht auf, wenn seine Meldung
// verloren ginge. Hier kommt eine Abfrage mit **einer** Spalte an einen
// `Scan`, der zehn erwartet.
func TestEinScanGegenDieFalscheSpaltenzahlMeldetSich(t *testing.T) {
	db := offeneDatenbank(t)

	if _, err := NeuerSammlungsspeicher(db).lesen(probekontext(t),
		`SELECT 1`); err == nil {
		t.Error("ein Scan gegen eine Spalte kam durch")
	}

	if _, err := NeuerPreisspeicher(db).lesen(probekontext(t),
		`SELECT 1`); err == nil {
		t.Error("ein Scan gegen eine Spalte kam durch")
	}

	// **Und die schmale Abfrage des Fensters ebenso** (Werkstück
	// #1500):
	// sie hat ihren eigenen `Scan` über neun Spalten.
	if _, err := NeuerSammlungsspeicher(db).ordnungLesen(probekontext(t), db,
		`SELECT 1`); err == nil {
		t.Error("ein Scan gegen eine Spalte kam durch")
	}
}

// **Ein geschlossener Bestand meldet sich auch aus den neuen Wegen.**
//
// # Was hier dazukommt
//
// `Aendern` fragt bei null getroffenen Zeilen nach, **warum** — und diese
// zweite Abfrage hat ihren eigenen Fehlerweg. Ohne einen Lauf dafür wäre die
// Unterscheidung zwischen „gibt es nicht" und „jemand anderes hat geschrieben"
// eine Behauptung über einen Weg, den niemand betreten hat.
func TestDieNachfrageMeldetEineGeschlosseneDatenbank(t *testing.T) {
	speicher := NeuerSammlungsspeicher(geschlosseneDatenbank(t))

	if err := speicher.warumNicht(probekontext(t), 1, 1); err == nil {
		t.Error("die geschlossene Datenbank nahm die Nachfrage an")
	}
}

// **Und die Preisstände melden sie auf beiden Wegen des Zugs.**
//
// Der Zug beginnt, schreibt und schließt ab — drei Stellen, an denen SQLite
// nein sagen kann. Eine geschlossene Datenbank scheitert schon am Beginn.
func TestDerPreiszugMeldetEineGeschlosseneDatenbank(t *testing.T) {
	speicher := NeuerPreisspeicher(geschlosseneDatenbank(t))

	stand := domain.Preisstand{
		Printing: "x", Quelle: domain.PreisquelleScryfall,
		Waehrung: "EUR", Finish: domain.FinishNonfoil,
		Wert: 1, Erhoben: time.Unix(0, 0),
	}

	if err := speicher.Eintragen(probekontext(t), stand); err == nil {
		t.Error("die geschlossene Datenbank nahm den Zug an")
	}
}

// **Ein Zug ohne Stände tut nichts und meldet nichts.**
//
// Ein Druck ohne Marktdaten — ein Token — bringt keine Preise mit. Einen Zug
// dafür zu beginnen wäre Arbeit ohne Gegenstand.
func TestEinZugOhneStaendeTutNichts(t *testing.T) {
	if err := NeuerPreisspeicher(geschlosseneDatenbank(t)).
		Eintragen(probekontext(t)); err != nil {
		t.Errorf("der leere Zug meldete %v", err)
	}
}

// **Ein Ergebnis, das nicht zählen kann, meldet sich.**
func TestEineAenderungMitUnzaehlbaremErgebnisMeldetSich(t *testing.T) {
	speicher := NeuerSammlungsspeicher(offeneDatenbank(t))

	if err := speicher.geaendert(probekontext(t), zaehlfehler{}, 1, 1); err == nil {
		t.Error("das unzählbare Ergebnis kam durch")
	}
}

// **Und ein Zug, dessen Tabelle es nicht gibt, bricht mittendrin ab.**
//
// # Warum das der Weg ist, der zählt
//
// Der Zug beginnt, und **danach** scheitert das Schreiben — genau die Stelle,
// an der der erste Anlauf einen halben Abruf hinterließ. Ohne Tabelle ist sie
// zu erreichen, ohne eine Naht ins Dateisystem zu erfinden.
func TestEinZugOhneTabelleBrichtAb(t *testing.T) {
	db := offeneDatenbank(t)

	if err := Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	if _, err := db.ExecContext(probekontext(t), `DROP TABLE preisstaende`); err != nil {
		t.Fatalf("die Tabelle entfernen: %v", err)
	}

	stand := domain.Preisstand{
		Printing: "x", Quelle: domain.PreisquelleScryfall,
		Waehrung: "EUR", Finish: domain.FinishNonfoil,
		Wert: 1, Erhoben: time.Unix(0, 0),
	}

	if err := NeuerPreisspeicher(db).Eintragen(probekontext(t), stand); err == nil {
		t.Error("der Zug ohne Tabelle kam durch")
	}
}

// **Ein Fehler nach der ersten geschriebenen Zeile lässt keine halbe Wahrheit
// stehen.**
//
// # Warum der Lauf darüber das nicht hält
//
// Er entfernt die Tabelle und lässt damit schon den **ersten** `INSERT`
// scheitern. Ob eine zuvor geschriebene Zeile zurückgenommen wird, ist dabei
// nie zu sehen — der Wächter hätte auch eine Fassung ohne Zug grün gelassen.
// Angestrichen vom Nachlauf über Phase 3.4 als Befund 8, und der Befund ist
// die **Reichweite**: eine Zusage über Teilfortschritt braucht Teilfortschritt.
//
// **Deshalb ein Auslöser statt einer fehlenden Tabelle.** Die erste Zeile geht
// wirklich hinein, die zweite bricht ab — und danach steht die Frage, für die
// es den Zug gibt: ist die erste noch da?
func TestEinFehlerNachDerErstenZeileNimmtSieZurueck(t *testing.T) {
	db := offeneDatenbank(t)

	if err := Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	if _, err := db.ExecContext(probekontext(t), `
		CREATE TRIGGER stolperstein BEFORE INSERT ON preisstaende
		WHEN NEW.waehrung = 'STOLPER'
		BEGIN SELECT RAISE(ABORT, 'absichtlich'); END`); err != nil {
		t.Fatalf("den Auslöser anlegen: %v", err)
	}

	stand := func(waehrung string) domain.Preisstand {
		return domain.Preisstand{
			Printing: "x", Quelle: domain.PreisquelleScryfall,
			Waehrung: waehrung, Finish: domain.FinishNonfoil,
			Wert: 1, Erhoben: time.Unix(0, 0),
		}
	}

	if err := NeuerPreisspeicher(db).Eintragen(probekontext(t),
		stand("EUR"), stand("STOLPER")); err == nil {
		t.Fatal("der Stapel mit der stolpernden Zeile kam durch")
	}

	// **Und die erste Zeile ist weg.** Ohne den Zug stünde sie da, und der
	// Verlauf trüge einen Abruf, den es nie gab.
	var zeilen int
	if err := db.QueryRowContext(probekontext(t),
		`SELECT count(*) FROM preisstaende`).Scan(&zeilen); err != nil {
		t.Fatalf("zählen: %v", err)
	}

	if zeilen != 0 {
		t.Errorf("es stehen %d Zeilen da — der Zug hat nichts zurückgenommen", zeilen)
	}
}

// reihevonScryfall benennt eine Reihe dieser Quelle (A3, #1442).
//
// **Eine eigene neben der in `preisstaende_test.go`**: das dort ist
// `package storage_test`, das hier `package storage` — dieselben drei Zeilen,
// und der Weg dazwischen führte über eine Ausfuhr, die niemand braucht.
func reihevonScryfall(waehrung string, finish domain.Finish) domain.Preisstand {
	return domain.Preisstand{
		Quelle:   domain.PreisquelleScryfall,
		Markt:    domain.MarktPapier,
		Art:      domain.PreisartVerkauf,
		Waehrung: waehrung,
		Finish:   finish,
	}
}
