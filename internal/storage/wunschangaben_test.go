package storage_test

// **Rang und Art eines Wunsches im Bestand** — Werkstück `T4` von Phase 5.6
// (#1469).
//
// # Was hier gemessen wird
//
// **Dass beide Angaben den Weg hin und zurück überstehen.** Eine Angabe, die
// beim Schreiben verlorengeht, fällt an der Oberfläche als „hat nicht
// gespeichert" auf — und im Bestand steht dann eine Zeile, die aussieht, als
// habe niemand etwas eingetragen.
//
// **Und dass eine Zeile ohne die Angaben die neutrale Antwort gibt.** Jede
// Zeile aus einem älteren Bestand trägt sie nicht; sie muss trotzdem
// antworten, und zwar mit `normal` und `genau`.
//
// **Kein `t.Parallel()`**, aus dem Grund, der in `sammlung_test.go` steht.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// einWunschmitRang baut einen Wunsch mit beiden Angaben.
func einWunschmitRang() *domain.Exemplar {
	exemplar := probeexemplar()
	exemplar.Wunsch = true
	exemplar.Wunschrang = domain.RangSehrHoch
	exemplar.Wunschart = domain.WunschJederDruck

	return exemplar
}

func TestRangUndArtUeberstehenDenWeg(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	id, err := speicher.Anlegen(probekontext(t), einWunschmitRang())
	if err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	gelesen, err := speicher.Einzeln(probekontext(t), id)
	if err != nil {
		t.Fatalf("Einzeln: %v", err)
	}

	if gelesen.Dringlichkeit() != domain.RangSehrHoch {
		t.Errorf("der Rang ist %q", gelesen.Dringlichkeit())
	}

	if gelesen.Druckbindung() != domain.WunschJederDruck {
		t.Errorf("die Art ist %q", gelesen.Druckbindung())
	}
}

func TestEineAenderungSchreibtBeideAngabenZurueck(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	id, err := speicher.Anlegen(probekontext(t), einWunschmitRang())
	if err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	gelesen, err := speicher.Einzeln(probekontext(t), id)
	if err != nil {
		t.Fatalf("Einzeln: %v", err)
	}

	gelesen.Wunschrang = domain.RangNiedrig
	gelesen.Wunschart = domain.WunschGenau

	if fehler := speicher.Aendern(probekontext(t), gelesen); fehler != nil {
		t.Fatalf("Aendern: %v", fehler)
	}

	danach, err := speicher.Einzeln(probekontext(t), id)
	if err != nil {
		t.Fatalf("Einzeln: %v", err)
	}

	if danach.Dringlichkeit() != domain.RangNiedrig || danach.Druckbindung() != domain.WunschGenau {
		t.Errorf("es steht %q und %q", danach.Dringlichkeit(), danach.Druckbindung())
	}
}

// **Eine Zeile ohne die Angaben antwortet neutral** — das ist die Zusage der
// Migration, und sie wird an einer Zeile gemessen, die nie etwas eingetragen
// hat.
func TestEineZeileOhneAngabenAntwortetNeutral(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	schlicht := probeexemplar()
	schlicht.Wunsch = true

	id, err := speicher.Anlegen(probekontext(t), schlicht)
	if err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	gelesen, err := speicher.Einzeln(probekontext(t), id)
	if err != nil {
		t.Fatalf("Einzeln: %v", err)
	}

	if gelesen.Wunschrang != "" || gelesen.Wunschart != "" {
		t.Errorf("es steht %q und %q im Bestand", gelesen.Wunschrang, gelesen.Wunschart)
	}

	if gelesen.Dringlichkeit() != domain.RangNormal || gelesen.Druckbindung() != domain.WunschGenau {
		t.Errorf("gelesen wird %q und %q", gelesen.Dringlichkeit(), gelesen.Druckbindung())
	}
}
