package storage_test

// **Was im Bestand herumliegt** — Werkstück O8 von Phase 5.2
// (#1432).

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

var pflegezeit = time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)

// waiseLegen trägt eine verwaiste Zeile ein — **mit abgeschalteten
// Fremdschlüsseln**, denn das Schema lässt sie sonst gar nicht zu.
//
// **Und genau das ist der Fall, um den es geht**: bis zum 2026-08-15 war
// `foreign_keys` aus, und jede Kaskade im Schema ein toter Buchstabe (#1220).
func waiseLegen(t *testing.T, db *sql.DB, anweisung string) {
	t.Helper()

	if _, fehler := db.ExecContext(context.Background(), `PRAGMA foreign_keys = OFF`); fehler != nil {
		t.Fatalf("die Fremdschlüssel abschalten: %v", fehler)
	}

	if _, fehler := db.ExecContext(context.Background(), anweisung); fehler != nil {
		t.Fatalf("die Waise legen: %v", fehler)
	}

	if _, fehler := db.ExecContext(context.Background(), `PRAGMA foreign_keys = ON`); fehler != nil {
		t.Fatalf("die Fremdschlüssel anschalten: %v", fehler)
	}
}

func befundeNach(t *testing.T, db *sql.DB) map[string]domain.Pflegebefund {
	t.Helper()

	befunde, fehler := storage.Pflegeanalyse(probekontext(t), db, pflegezeit)
	if fehler != nil {
		t.Fatalf("die Analyse: %v", fehler)
	}

	nachArt := map[string]domain.Pflegebefund{}
	for _, befund := range befunde {
		nachArt[befund.Art] = befund
	}

	return nachArt
}

// **Ein sauberer Bestand hat keinen Befund** — und das ist die Gegenprobe zu
// allem darunter: eine Analyse, die immer etwas findet, findet nichts (§10).
func TestEinSauberBestandHatKeinenBefund(t *testing.T) {
	db := frischeDatenbank(t)

	befunde, fehler := storage.Pflegeanalyse(probekontext(t), db, pflegezeit)
	if fehler != nil {
		t.Fatalf("die Analyse: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("ein frischer Bestand hat %d Befunde: %+v", len(befunde), befunde)
	}
}

// **Eine verwaiste Zeile wird gefunden, benannt und eingeordnet.**
func TestEineVerwaisteZeileWirdGefundenUndEingeordnet(t *testing.T) {
	db := frischeDatenbank(t)

	waiseLegen(t, db, `INSERT INTO card_values (card_id, feld, wert, reihenfolge)
	                   VALUES (4711, 'name', 'eine Waise', 0)`)

	befund, da := befundeNach(t, db)["kartenwerteOhneKarte"]
	if !da {
		t.Fatal("die Waise wurde nicht gefunden")
	}

	if befund.Klasse != domain.PflegeSicher {
		t.Errorf("die Klasse ist %q", befund.Klasse)
	}

	if befund.Anzahl != 1 || len(befund.Beispiele) != 1 {
		t.Errorf("der Befund ist %+v", befund)
	}

	if len(befund.Revision) != 64 {
		t.Errorf("die Revision ist %q — erwartet war ein SHA-256", befund.Revision)
	}
}

// **Die Revision hängt an den Zeilen und nicht an ihrer Zahl.**
//
// Zwei Zeilen weg und zwei dazu ergäben dieselbe Zahl und einen anderen
// Bestand — eine Revision über die Zahl merkte es nicht.
func TestDieRevisionHaengtAnDenZeilenUndNichtAnDerZahl(t *testing.T) {
	db := frischeDatenbank(t)

	waiseLegen(t, db, `INSERT INTO card_values (card_id, feld, wert, reihenfolge)
	                   VALUES (4711, 'eins', 'x', 0)`)

	vorher := befundeNach(t, db)["kartenwerteOhneKarte"].Revision

	waiseLegen(t, db, `DELETE FROM card_values WHERE feld = 'eins'`)
	waiseLegen(t, db, `INSERT INTO card_values (card_id, feld, wert, reihenfolge)
	                   VALUES (4711, 'zwei', 'x', 0)`)

	nachher := befundeNach(t, db)["kartenwerteOhneKarte"]

	if nachher.Anzahl != 1 {
		t.Fatalf("die Zahl ist %d, erwartet war 1", nachher.Anzahl)
	}

	if nachher.Revision == vorher {
		t.Error("die Revision blieb gleich, obwohl es eine andere Zeile ist")
	}
}

// **Beispiele werden gedeckelt** — eine Antwort soll nicht die halbe Tabelle
// tragen.
func TestDieBeispieleWerdenGedeckelt(t *testing.T) {
	db := frischeDatenbank(t)

	for nummer := 1; nummer <= storage.Pflegebeispiele+3; nummer++ {
		waiseLegen(t, db, `INSERT INTO card_values (card_id, feld, wert, reihenfolge)
		                   VALUES (4711, 'f`+string(rune('a'+nummer))+`', 'x', 0)`)
	}

	befund := befundeNach(t, db)["kartenwerteOhneKarte"]

	if befund.Anzahl != storage.Pflegebeispiele+3 {
		t.Errorf("die Zahl ist %d", befund.Anzahl)
	}

	if len(befund.Beispiele) != storage.Pflegebeispiele {
		t.Errorf("es stehen %d Beispiele da, erwartet waren %d",
			len(befund.Beispiele), storage.Pflegebeispiele)
	}
}

// **Was `nurInformieren` heißt, wird gefunden und nicht weggeräumt.**
func TestEineOffeneRueckkehrStehtDaUndLaesstSichNichtWegraeumen(t *testing.T) {
	db := frischeDatenbank(t)

	if _, fehler := storage.NeuesWiederherstellungsbuch(db).Vermerken(probekontext(t),
		domain.Wiederherstellungsepoche{
			Zurueckgespielt: pflegezeit, Datenstand: pflegezeit,
			Quelle: "probe.msesicherung", Klaerung: domain.KlaerungOffen,
		}); fehler != nil {
		t.Fatalf("die Epoche vermerken: %v", fehler)
	}

	befund, da := befundeNach(t, db)["offeneRueckkehren"]
	if !da {
		t.Fatal("die offene Rückkehr wurde nicht gefunden")
	}

	if befund.Klasse != domain.PflegeNurInformieren {
		t.Errorf("die Klasse ist %q", befund.Klasse)
	}

	// **Und der Weg lehnt sie ab** — es gibt keine Löschabfrage dazu, und das
	// ist die Zusage und keine Vergesslichkeit.
	if _, fehler := storage.PflegeAusfuehren(probekontext(t), db, pflegezeit,
		map[string]string{"offeneRueckkehren": befund.Revision}); fehler == nil {
		t.Error("eine offene Rückkehr ließ sich wegräumen")
	}
}

// **Eine Maßnahme, die es nicht gibt, wird abgewiesen.**
func TestEineUnbekannteMassnahmeWirdAbgewiesen(t *testing.T) {
	db := frischeDatenbank(t)

	if _, fehler := storage.PflegeAusfuehren(probekontext(t), db, pflegezeit,
		map[string]string{"alleKartenLoeschen": "abc"}); fehler == nil {
		t.Error("eine erfundene Maßnahme kam durch")
	}
}

// **Ausgeführt wird, was passt — und die Bilanz nennt die Zahl.**
func TestEineMassnahmeRaeumtWegUndNenntDieZahl(t *testing.T) {
	db := frischeDatenbank(t)

	waiseLegen(t, db, `INSERT INTO card_values (card_id, feld, wert, reihenfolge)
	                   VALUES (4711, 'name', 'eine Waise', 0)`)

	befund := befundeNach(t, db)["kartenwerteOhneKarte"]

	zeilen, fehler := storage.PflegeAusfuehren(probekontext(t), db, pflegezeit,
		map[string]string{"kartenwerteOhneKarte": befund.Revision})
	if fehler != nil {
		t.Fatalf("ausführen: %v", fehler)
	}

	if len(zeilen) != 1 || zeilen[0].Ausgang != domain.PflegeErledigt || zeilen[0].Anzahl != 1 {
		t.Fatalf("die Bilanz ist %+v", zeilen)
	}

	// **Und danach ist der Befund weg** — die Nachprüfung ist die Abnahme.
	if _, da := befundeNach(t, db)["kartenwerteOhneKarte"]; da {
		t.Error("der Befund steht nach der Pflege weiter da")
	}
}

// **Eine seit der Vorschau geänderte Zeile stoppt die Maßnahme** — auch dann,
// wenn der Abgleich davor noch grün war. Das ist die zweite Prüfung, und sie
// steht in derselben Transaktion wie das Löschen.
func TestEineGeaenderteZeileStopptDieMassnahmeImLetztenAugenblick(t *testing.T) {
	db := frischeDatenbank(t)

	waiseLegen(t, db, `INSERT INTO card_values (card_id, feld, wert, reihenfolge)
	                   VALUES (4711, 'name', 'eine Waise', 0)`)

	zeilen, fehler := storage.PflegeAusfuehren(probekontext(t), db, pflegezeit,
		map[string]string{"kartenwerteOhneKarte": "eine Revision von gestern"})
	if fehler != nil {
		t.Fatalf("ausführen: %v", fehler)
	}

	if len(zeilen) != 1 || zeilen[0].Ausgang != domain.PflegeKonflikt {
		t.Fatalf("die Bilanz ist %+v", zeilen)
	}

	// **Und die Zeile steht noch da.**
	if _, da := befundeNach(t, db)["kartenwerteOhneKarte"]; !da {
		t.Error("die Zeile wurde trotz Konflikt weggeräumt")
	}
}

// **Ein Befund, der zwischendurch verschwand, ist erledigt und kein Fehler.**
func TestEinVerschwundenerBefundGiltAlsErledigt(t *testing.T) {
	db := frischeDatenbank(t)

	zeilen, fehler := storage.PflegeAusfuehren(probekontext(t), db, pflegezeit,
		map[string]string{"kartenwerteOhneKarte": "egal"})
	if fehler != nil {
		t.Fatalf("ausführen: %v", fehler)
	}

	if len(zeilen) != 1 || zeilen[0].Ausgang != domain.PflegeSchonWeg {
		t.Errorf("die Bilanz ist %+v", zeilen)
	}
}

// **Ein Konflikt an einer Stelle nimmt der anderen nichts.**
func TestEinKonfliktNimmtDerAnderenMassnahmeNichts(t *testing.T) {
	db := frischeDatenbank(t)

	waiseLegen(t, db, `INSERT INTO card_values (card_id, feld, wert, reihenfolge)
	                   VALUES (4711, 'name', 'eine Waise', 0)`)
	waiseLegen(t, db, `INSERT INTO set_values (set_id, feld, wert, reihenfolge)
	                   VALUES (4711, 'title', 'noch eine', 0)`)

	befunde := befundeNach(t, db)

	zeilen, fehler := storage.PflegeAusfuehren(probekontext(t), db, pflegezeit,
		map[string]string{
			"kartenwerteOhneKarte": "von gestern",
			"setangabenOhneSet":    befunde["setangabenOhneSet"].Revision,
		})
	if fehler != nil {
		t.Fatalf("ausführen: %v", fehler)
	}

	ausgaenge := map[string]domain.Pflegeausgang{}
	for _, zeile := range zeilen {
		ausgaenge[zeile.Art] = zeile.Ausgang
	}

	if ausgaenge["kartenwerteOhneKarte"] != domain.PflegeKonflikt {
		t.Errorf("die geänderte Maßnahme ist %q", ausgaenge["kartenwerteOhneKarte"])
	}

	if ausgaenge["setangabenOhneSet"] != domain.PflegeErledigt {
		t.Errorf("die unveränderte Maßnahme ist %q", ausgaenge["setangabenOhneSet"])
	}
}

// **Zeitgebundene Stellen richten sich nach dem übergebenen Augenblick** —
// und nicht nach der Uhr dieser Maschine.
func TestEinAbgelaufenerNachweisRichtetSichNachDemAugenblick(t *testing.T) {
	db := frischeDatenbank(t)

	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO freigabenachweise
		   (kennung, nutzer_id, profil_id, zweck, ziel, revision, ausgestellt, ablauf)
		 VALUES ('abc', 1, 1, 'probe', 'probe', 1, ?, ?)`,
		storage.Zeitstempel(pflegezeit.Add(-2*time.Hour)),
		storage.Zeitstempel(pflegezeit.Add(-time.Hour))); fehler != nil {
		t.Fatalf("den Nachweis legen: %v", fehler)
	}

	// **Vorher gibt es ihn nicht** — die Gegenprobe zur Zeitgrenze.
	frueher, fehler := storage.Pflegeanalyse(probekontext(t), db,
		pflegezeit.Add(-3*time.Hour))
	if fehler != nil {
		t.Fatalf("die frühe Analyse: %v", fehler)
	}

	for _, befund := range frueher {
		if befund.Art == "abgelaufeneNachweise" {
			t.Error("der Nachweis galt schon vor seinem Ablauf als abgelaufen")
		}
	}

	befund, da := befundeNach(t, db)["abgelaufeneNachweise"]
	if !da {
		t.Fatal("der abgelaufene Nachweis wurde nicht gefunden")
	}

	if befund.Klasse != domain.PflegeEntscheidung {
		t.Errorf("die Klasse ist %q", befund.Klasse)
	}

	zeilen, ausfuehrfehler := storage.PflegeAusfuehren(probekontext(t), db, pflegezeit,
		map[string]string{"abgelaufeneNachweise": befund.Revision})
	if ausfuehrfehler != nil {
		t.Fatalf("ausführen: %v", ausfuehrfehler)
	}

	if len(zeilen) != 1 || zeilen[0].Ausgang != domain.PflegeErledigt {
		t.Errorf("die Bilanz ist %+v", zeilen)
	}
}

// **Auf einer Datenbank, die zu ist, meldet sich jeder Weg** — statt einer
// leeren Analyse, die wie ein sauberer Bestand aussähe. Das ist die eine
// Auskunft, die niemals falsch sein darf.
func TestAufEinerGeschlossenenDatenbankMeldetSichDiePflege(t *testing.T) {
	db := zugemachteDatenbank(t)

	if _, fehler := storage.Pflegeanalyse(probekontext(t), db, pflegezeit); fehler == nil {
		t.Error("die Analyse meldete nichts")
	}

	zeilen, fehler := storage.PflegeAusfuehren(probekontext(t), db, pflegezeit,
		map[string]string{"kartenwerteOhneKarte": "abc"})
	if fehler != nil {
		t.Fatalf("ausführen: %v", fehler)
	}

	if len(zeilen) != 1 || zeilen[0].Ausgang != domain.PflegeGescheitert {
		t.Errorf("die Bilanz ist %+v", zeilen)
	}
}

// **Die zehn Stellen sind genau die zehn Arten der Domäne.**
//
// **Ein Wächter, und er hält beide Richtungen**: `domain.Pflegeabgleich` weist
// eine Art ab, die es nicht gibt — dafür braucht er die Liste. Läuft sie
// auseinander, wird entweder eine echte Maßnahme abgewiesen oder eine
// erfundene durchgelassen, und beides sähe wie ein Zufall aus (§10).
func TestDieStellenSindGenauDieArtenDerDomaene(t *testing.T) {
	db := frischeDatenbank(t)

	// **Über die Analyse und nicht über die Liste im Quelltext**: sie ist der
	// Weg, den die Anwendung geht. Damit jede Stelle **einen** Befund gibt,
	// wird zu jeder eine Zeile gelegt — das geht nicht; stattdessen prüft der
	// Lauf, dass jede Art der Domäne **ausführbar oder bekannt** ist.
	for _, art := range domain.Pflegearten() {
		// **Eine Maßnahme, die es gibt, wird nicht als erfunden abgewiesen.**
		zeilen, fehler := storage.PflegeAusfuehren(probekontext(t), db, pflegezeit,
			map[string]string{art: "egal"})

		switch {
		case art == domain.PflegeDubletten || art == domain.PflegeRueckkehren ||
			art == domain.PflegeIntegritaet || art == domain.PflegeFremdschluessel:
			// **Die vier `nurInformieren`-Arten haben keine Löschabfrage** —
			// und das ist die Zusage. Zwei davon kamen mit Befund 2 des
			// Zweitblicks über Phase 5.2 dazu: was SQLite selbst über seine
			// Datei sagt, wird gezeigt und nicht weggeräumt.
			if fehler == nil {
				t.Errorf("%q ließ sich ausführen", art)
			}
		case fehler != nil:
			t.Errorf("%q gilt als erfunden: %v", art, fehler)
		case len(zeilen) != 1:
			t.Errorf("%q gab %d Zeilen", art, len(zeilen))
		}
	}
}
