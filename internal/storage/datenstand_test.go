package storage_test

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Der Datenstand ändert sich, wenn ein *anderer* schreibt.**
//
// # Die Naht zwischen zwei Prozessen
//
// Der MCP-Server ist ein eigener Prozess auf derselben Datei
// ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)). Was ein
// Modell dort ändert, sah das offene Fenster bisher **nicht**
// (#156).
//
// `PRAGMA data_version` ist SQLites Antwort darauf: die Zahl steigt, sobald
// eine **andere** Verbindung bestätigt hat. Für eigene Änderungen bleibt sie
// stehen — und genau das ist richtig, denn wer selbst schreibt, weiß es.
//
// # Warum das mit zwei Verbindungen geprüft wird
//
// Weil `data_version` **genau darüber** Auskunft gibt: was zählt, ist eine
// andere Verbindung, nicht ein anderer Prozess. Zwei Vorräte in einem Lauf
// sind für diese Frage dasselbe wie zwei Programme — und ein Test, der ein
// zweites Programm startet, prüfte zusätzlich das Starten.
//
// **Was das nicht zeigt**, steht im Vermerk
// ([0017](../../docs/entscheidungen/0017-die-prozesse-bemerken-einander.md)):
// dass die Abfrage im laufenden Fenster oft genug kommt.
func TestDerDatenstandSteigtWennEinAndererSchreibt(t *testing.T) {
	pfad := "file:" + t.TempDir() + "/probe.db"

	fenster, err := storage.Oeffnen(probekontext(t), pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	t.Cleanup(func() { _ = fenster.Close() })

	if fehler := storage.Migriere(probekontext(t), fenster, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	server, err := storage.Oeffnen(probekontext(t), pfad)
	if err != nil {
		t.Fatalf("zweite Verbindung: %v", err)
	}

	t.Cleanup(func() { _ = server.Close() })

	vorher, err := storage.Datenstand(probekontext(t), fenster)
	if err != nil {
		t.Fatalf("Datenstand: %v", err)
	}

	// **Die eigene Schreibung bewegt ihn nicht.**
	if _, fehler := storage.NeuerSetSpeicher(fenster).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Eigenes", Game: "probe",
	}, nil); fehler != nil {
		t.Fatalf("eigenes Anlegen: %v", fehler)
	}

	eigen, err := storage.Datenstand(probekontext(t), fenster)
	if err != nil {
		t.Fatalf("Datenstand: %v", err)
	}

	if eigen != vorher {
		t.Errorf("der eigene Schreibvorgang hat den Stand bewegt: %d → %d", vorher, eigen)
	}

	// **Die fremde schon.**
	if _, fehler := storage.NeuerSetSpeicher(server).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Fremdes", Game: "probe",
	}, nil); fehler != nil {
		t.Fatalf("fremdes Anlegen: %v", fehler)
	}

	nachher, err := storage.Datenstand(probekontext(t), fenster)
	if err != nil {
		t.Fatalf("Datenstand: %v", err)
	}

	if nachher == eigen {
		t.Errorf("der Stand blieb bei %d — die fremde Änderung kam nicht an", nachher)
	}
}

// **Das eigene Schreiben bewegt den Datenstand nicht — auch nicht neben einem
// laufenden Lesen.** Befund aus der Sichtprüfung von U5
// (#1485):
// das Fenster hielt sein eigenes `EinstellungSetzen` für eine fremde Änderung
// und lud die Kartenseite zwei Sekunden später neu, mit drei Rasterläufen.
//
// # Warum der Test oben den Fall nicht enthielt
//
// `data_version` zählt je **Verbindung**, nicht je Prozess. Der Test oben
// fragt und schreibt nacheinander über den Vorrat, und `database/sql` gibt
// beiden dieselbe untätige Verbindung — grün, ohne den Fall zu betreten. Im
// Fenster laufen Vorschau und Schreiben nebeneinander, und der Standleser
// hielt obendrein eine eigene Verbindung fest: jedes Schreiben lief auf einer
// **anderen** und hob die Zahl. Hier hält der Leser seine Verbindung, während
// geschrieben wird — so war es im Fenster.
func TestDasEigeneSchreibenBewegtDenDatenstandNicht(t *testing.T) {
	pfad := "file:" + t.TempDir() + "/probe.db"
	db, err := storage.Oeffnen(probekontext(t), pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	leser := storage.NeuerStandleser(db)

	vorher, err := leser.Datenstand(probekontext(t))
	if err != nil {
		t.Fatalf("Datenstand: %v", err)
	}

	if _, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Eigenes", Game: "probe",
	}, nil); fehler != nil {
		t.Fatalf("eigenes Anlegen: %v", fehler)
	}

	nachher, err := leser.Datenstand(probekontext(t))
	if err != nil {
		t.Fatalf("Datenstand: %v", err)
	}

	if nachher != vorher {
		t.Errorf("das eigene Schreiben gilt als fremde Änderung: %d → %d", vorher, nachher)
	}
}

// **Der Standleser ist derselbe Weg über einen Typ** — die Naht, die `app`
// halten kann, ohne `storage` außerhalb des Composition Root zu kennen (§2).
func TestDerStandleserGehtDenselbenWeg(t *testing.T) {
	db := frischeDatenbank(t)

	ueberFunktion, err := storage.Datenstand(probekontext(t), db)
	if err != nil {
		t.Fatalf("Funktion: %v", err)
	}

	ueberTyp, err := storage.NeuerStandleser(db).Datenstand(probekontext(t))
	if err != nil {
		t.Fatalf("Typ: %v", err)
	}

	if ueberTyp != ueberFunktion {
		t.Errorf("%d gegen %d", ueberTyp, ueberFunktion)
	}
}

// **Ein Fehler beim Lesen wird übersetzt** (§8) — rohe SQLite-Meldungen
// erreichen die Oberfläche nie.
func TestEinFehlerAmDatenstandWirdUebersetzt(t *testing.T) {
	db := frischeDatenbank(t)

	if err := db.Close(); err != nil {
		t.Fatalf("schließen: %v", err)
	}

	if _, err := storage.Datenstand(probekontext(t), db); err == nil {
		t.Error("kein Fehler")
	}
}

// **Eine gescheiterte Abfrage wird bei jedem Ruf gemeldet** — der Leser
// hält nichts fest, das einen zweiten Ruf anders enden ließe. Der
// Stolpertreiber führt den Fehler herbei.
func TestEineGescheiterteAbfrageWirdJedesMalGemeldet(t *testing.T) {
	db, err := sql.Open("stolper-fragen", "")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	leser := storage.NeuerStandleser(db)

	for lauf := range 2 {
		if _, fehler := leser.Datenstand(probekontext(t)); fehler == nil {
			t.Errorf("Lauf %d: kein Fehler", lauf)
		}
	}
}

// **Eine Datenbank, die keine Verbindung mehr hergibt, wird gemeldet.**
func TestOhneVerbindungGibtEsKeinenDatenstand(t *testing.T) {
	db, err := storage.FuerLaeufe("file:" + t.TempDir() + "/zu.db")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if _, fehler := storage.NeuerStandleser(db).Datenstand(probekontext(t)); fehler == nil {
		t.Error("kein Fehler")
	}
}
