package app_test

// **Die atomare Gruppe an der Naht** — Werkstück A13.1b von Phase 5.3
// (#1452).
//
// # Was hier geprüft wird
//
// **Dass die Naht umrechnet und nichts entscheidet.** Welche Vorschläge eine
// Gruppe bilden, entscheidet `service`; hier wird aus `map[string][]string`
// der Wert des Bestands, und das Set kommt von der Karte.
//
// **Und dass jeder Fehlerweg durchreicht — mit leeren Listen.** „Nichts
// geschrieben" und „die Frage ging schief" sind zwei Auskünfte.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// dieLesungDerNaht ist der Abstand von der laufenden Lesung bis zu der, die
// [kikarten.FelderGruppeSetzen] selbst macht.
//
// **Eine benannte Zahl statt einer gestreuten**: wer sie ändert, findet den
// Grund daneben, statt in einem Lauf über ein „+3" zu stolpern.
const dieLesungDerNaht = 3

// **Ohne Bestand kommt ein Fehler und keine halbe Antwort.**
func TestEineGruppeOhneBestandGibtEinenFehler(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").KIGruppeUebernehmen(probekontext(t),
		[]app.KIGruppenwahlDTO{{Kennung: "a1", Felder: []string{"name"}}})

	if antwort.Fehler == nil {
		t.Fatal("ohne Bestand kam kein Fehler")
	}

	// **Und die Liste ist eine Liste** — für ein Sprachmodell sieht `null`
	// aus wie „konnte nicht beantwortet werden".
	if antwort.Namen == nil {
		t.Error("die Namen sind nil")
	}
}

// **Einen Auftrag, den es nicht gibt, reicht die Naht als Fehler durch.**
func TestEineGruppeMitUnbekanntemAuftragReichtDurch(t *testing.T) {
	welt := einekiwelt(t)

	antwort := welt.anwendung.KIGruppeUebernehmen(probekontext(t),
		[]app.KIGruppenwahlDTO{{Kennung: "gibtesnicht", Felder: []string{"name"}}})

	if antwort.Fehler == nil {
		t.Fatal("der erfundene Auftrag kam durch")
	}

	if antwort.Karten != 0 {
		t.Errorf("übernommen wurden %d Karten", antwort.Karten)
	}
}

// **Eine leere Gruppe wird abgewiesen und schreibt nichts.**
func TestEineLeereGruppeAnDerNahtWirdAbgewiesen(t *testing.T) {
	welt := einekiwelt(t)

	antwort := welt.anwendung.KIGruppeUebernehmen(probekontext(t), nil)
	if antwort.Fehler == nil {
		t.Fatal("die leere Gruppe kam durch")
	}
}

// **Und eine geglückte Gruppe schreibt — mit Herkunft und in einem Zug.**
//
// # Was dieser Lauf hält, den anderen nicht halten
//
// **Die Umrechnung.** Aus `map[string][]string` wird der Wert des Bestands,
// das Set kommt von der **Karte** und nicht vom Aufrufer, und die Herkunft
// trägt den Abdruck des Geschriebenen. Die Läufe darüber prüfen die Absagen;
// dieser prüft, was ankommt.
func TestEineGeglueckteGruppeSchreibtMitHerkunft(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	vorschlag := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "Schreibe den Regeltext.", "")
	if vorschlag.Fehler != nil {
		t.Fatalf("%+v", vorschlag.Fehler)
	}

	antwort := welt.anwendung.KIGruppeUebernehmen(probekontext(t),
		[]app.KIGruppenwahlDTO{{
			Kennung: vorschlag.Kennung,
			Felder:  namenDerFelder(vorschlag.Felder),
		}})
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Karten != 1 || antwort.Felder == 0 {
		t.Errorf("übernommen wurde %+v", antwort)
	}

	// **Die Namen stehen je Karte da** — nicht nur eine Zahl.
	if len(antwort.Namen) != 1 || antwort.Namen[0].KarteID != welt.karte {
		t.Errorf("die Namen sind %+v", antwort.Namen)
	}

	if werte := welt.kartenwerte(t); !strings.Contains(werte, "Blitzschlag") {
		t.Errorf("die Karte trägt %q", werte)
	}
}

// **Ohne Kartenleser scheitert die Naht, bevor sie schreibt.**
//
// **Und das ist ein anderer Weg als „ohne Bestand" oben**: dort gibt es kein
// Auftragswerk, hier gibt es eines — nur keine Karte zu lesen, und ohne sie
// weiß niemand, zu welchem Set der Schritt gehört.
func TestOhneKartenleserSchreibtDieGruppeNichts(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	vorschlag := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "Schreibe den Regeltext.", "")
	if vorschlag.Fehler != nil {
		t.Fatalf("%+v", vorschlag.Fehler)
	}

	// **Den Leser wegnehmen, nachdem der Vorschlag liegt** — sonst
	// scheiterte schon das Vorschlagen, und der Lauf sagte nichts über die
	// Gruppe.
	welt.anwendung.MitEditor(nil, nil, nil, nil)

	antwort := welt.anwendung.KIGruppeUebernehmen(probekontext(t),
		[]app.KIGruppenwahlDTO{{
			Kennung: vorschlag.Kennung,
			Felder:  namenDerFelder(vorschlag.Felder),
		}})
	if antwort.Fehler == nil {
		t.Fatal("ohne Kartenleser wurde geschrieben")
	}
}

// **Eine Kartenkennung, die keine Zahl ist, weist die Bindung ab.**
//
// **Am Fenster kommen Kennungen als Zeichenketten** — eine
// JavaScript-Zahl trägt keine 64 Bit —, und was keine Zahl ist, darf keinen
// Ruf auslösen, der Geld kostet.
func TestEineKaputteKartenkennungWeistDenBatchAb(t *testing.T) {
	t.Parallel()

	antwort := app.NeueBindungen(app.NeueApp("")).KIBatchVorschlagen(t.Context(),
		[]string{"keineZahl"}, "ollama", "gemma4:8b", "Schreibe.", "")

	if antwort.Fehler == nil {
		t.Fatal("die kaputte Kennung kam durch")
	}

	if antwort.Ziele == nil {
		t.Error("die Ziele sind nil")
	}
}

// **Und die Naht reicht die Gruppe bis zum Bestand durch.**
//
// # Warum das hier steht und nicht nur im Flächenlauf
//
// **Dort wird der Weg gefahren**, hier die Frage davor: dass die
// Durchreiche überhaupt existiert. Eine Karte gibt es nicht, die Antwort ist
// deshalb ein Fehler — und genau der belegt, dass die Naht steht, statt
// stillschweigend nichts zu tun (dieselbe Messung wie bei
// [TestDieMCPSchreibnahtEntsteht]).
func TestDieSchreibnahtReichtDieGruppeDurch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	pfad, err := app.Datenbankpfad()
	if err != nil {
		t.Fatalf("Pfad: %v", err)
	}

	if fehler := os.MkdirAll(filepath.Dir(pfad), 0o750); fehler != nil {
		t.Fatalf("Verzeichnis: %v", fehler)
	}

	db, err := storage.Oeffnen(probekontext(t), "file:"+pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() {
		if fehler := db.Close(); fehler != nil {
			t.Errorf("schließen: %v", fehler)
		}
	})

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	naht, err := app.Schreibnaht(db, domain.UrheberServer)
	if err != nil {
		t.Fatalf("Naht: %v", err)
	}

	if fehler := naht.GruppeVonAussenMitHerkunft(probekontext(t),
		[]domain.Kartenschritt{{
			KarteID: 999, SetID: 1,
			Werte:      []domain.Value{{Field: "name", Text: "Eisball"}},
			Herkunft:   "ki",
			Herkuenfte: []domain.KIFeldherkunft{{Feld: "name"}},
		}}); fehler == nil {
		t.Fatal("die Karte 999 wurde geschrieben")
	}
}

// **Eine Karte, die sich nicht lesen lässt, bricht die Gruppe.**
//
// **Vor dem ersten Schreiben**: die Naht liest **alle** Karten, bevor sie
// eine schreibt — dieselbe Reihenfolge wie in `service`, und aus demselben
// Grund.
func TestEineUnlesbareKarteBrichtDieGruppe(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	vorschlag := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "Schreibe den Regeltext.", "")
	if vorschlag.Fehler != nil {
		t.Fatalf("%+v", vorschlag.Fehler)
	}

	// **Und zwar erst beim Lesen der Naht**, nicht bei der Standprüfung
	// davor: `service` ruft `Revision`, die dieselbe Karte liest. Ein Doppel,
	// das **immer** klemmt, bringt die Gruppe schon dort zu Fall — und der
	// Lesefehler **in** der Naht bliebe unbetreten.
	//
	// **Der Fall kommt im Betrieb vor**: zwischen der Standprüfung und dem
	// Schreiben kann ein zweiter Prozess die Karte löschen, und 0010 sieht
	// mehrere ausdrücklich vor.
	// **Erst die dritte Lesung ist die der Naht** — davor liegen zwei, die
	// durchgehen sollen. Die Zahl ist **gemessen** und nicht überlegt: bei
	// +2 blieb der Zweig unbetreten, bei +4 ebenso, und nur +3 trifft ihn
	// (nachgezählt am 2026-09-13 über die Abdeckung).
	welt.karten.klemmeAb(welt.karten.gelesen()+dieLesungDerNaht,
		errors.New("die Karte ist fort"))

	antwort := welt.anwendung.KIGruppeUebernehmen(probekontext(t),
		[]app.KIGruppenwahlDTO{{
			Kennung: vorschlag.Kennung,
			Felder:  namenDerFelder(vorschlag.Felder),
		}})
	if antwort.Fehler == nil {
		t.Fatal("die unlesbare Karte kam durch")
	}
}

// **Und die Bindung nimmt eine Zahl entgegen.**
//
// **Der Weg dahinter scheitert** — es gibt keinen Bestand —, und das ist die
// Messung: die Kennung wurde gelesen, bevor irgendetwas anderes geschah.
func TestDieBatchbindungLiestIhreKennungen(t *testing.T) {
	t.Parallel()

	antwort := app.NeueBindungen(app.NeueApp("")).KIBatchVorschlagen(t.Context(),
		[]string{"4"}, "ollama", "gemma4:8b", "Schreibe.", "")

	if antwort.Fehler == nil {
		t.Fatal("ohne Bestand kam kein Fehler")
	}

	if antwort.Fehler.Feld == "karteIds" {
		t.Errorf("die gültige Kennung wurde abgewiesen: %+v", antwort.Fehler)
	}
}

// **Ohne Schreiber scheitert die Naht, bevor sie schreibt.**
//
// **Ein anderer Weg als der ohne Kartenleser**: `Revision` prüft den Leser
// und kommt damit vorher zum Zug; den **Schreiber** prüft sie nicht, und
// deshalb ist dieser Riegel derjenige, der wirklich greift.
func TestOhneSchreiberSchreibtDieGruppeNichts(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	vorschlag := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "Schreibe den Regeltext.", "")
	if vorschlag.Fehler != nil {
		t.Fatalf("%+v", vorschlag.Fehler)
	}

	// **Der Leser bleibt, der Schreiber geht** — genau die Lage, die der
	// Riegel meint.
	welt.anwendung.MitEditor(nil, welt.karten, nil, nil)

	antwort := welt.anwendung.KIGruppeUebernehmen(probekontext(t),
		[]app.KIGruppenwahlDTO{{
			Kennung: vorschlag.Kennung,
			Felder:  namenDerFelder(vorschlag.Felder),
		}})
	if antwort.Fehler == nil {
		t.Fatal("ohne Schreiber wurde geschrieben")
	}
}
