package storage_test

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// bestandMitKarte legt ein Set mit einer Karte an und gibt beide Kennungen.
func bestandMitKarte(t *testing.T) (db *sql.DB, setID, karteID int64) {
	t.Helper()

	datenbank := frischeDatenbank(t)

	nummer, err := storage.NeuerSetSpeicher(datenbank).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "probe",
		Cards: []*domain.Card{{Values: []domain.Value{{Field: "name", Text: "Feuerball"}}}},
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	karten, err := storage.NeuerSetLeser(datenbank).Karten(probekontext(t), nummer)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	return datenbank, nummer, karten[0].ID
}

// **Jede Änderung kommt in den Verlauf — in derselben Transaktion.**
//
// # Warum das der dritte Grund für SQLite ist
//
// #52: nach
// einem Neustart lässt sich zurücknehmen, was gestern geändert wurde. Ein
// Verlauf im Arbeitsspeicher könnte das nie.
//
// **In derselben Transaktion**, weil zwei Prozesse schreiben (0010): ein
// Verlaufseintrag, der nach der Änderung entsteht, fehlt genau dann, wenn
// jemand dazwischen abstürzt — und dann ist die Änderung da und die Rücknahme
// fort.
func TestJedeAenderungKommtInDenVerlauf(t *testing.T) {
	db, setID, karteID := bestandMitKarte(t)

	schreiber := storage.NeuerKartenschreiber(db)

	if _, _, err := schreiber.FeldSetzenMitVorwert(
		probekontext(t), karteID, "name", "Eisball"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	verlauf, err := storage.NeuerVerlaufsleser(db).Verlauf(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Verlauf: %v", err)
	}

	if len(verlauf) != 1 {
		t.Fatalf("%d Einträge, erwartet einer", len(verlauf))
	}

	eintrag := verlauf[0]
	if eintrag.Handlung != "feld" || eintrag.KarteID != karteID {
		t.Errorf("%+v", eintrag)
	}

	if eintrag.Vorher != "Feuerball" || !eintrag.GabVorher {
		t.Errorf("Vorwert %q / %v", eintrag.Vorher, eintrag.GabVorher)
	}

	if eintrag.Nachher != "Eisball" || !eintrag.GabNachher {
		t.Errorf("Nachwert %q / %v", eintrag.Nachher, eintrag.GabNachher)
	}
}

// **Zurücknehmen stellt den Stand von vorher her — auch nach einem Neustart.**
func TestZuruecknehmenStelltDenVorwertHer(t *testing.T) {
	db, setID, karteID := bestandMitKarte(t)

	schreiber := storage.NeuerKartenschreiber(db)

	if _, _, err := schreiber.FeldSetzenMitVorwert(
		probekontext(t), karteID, "name", "Eisball"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	// **Ein neuer Leser auf derselben Datei** — nichts steht im Arbeitsspeicher.
	_, genommen, err := storage.NeuerVerlaufsleser(db).Zuruecknehmen(probekontext(t), setID)
	if err != nil {
		t.Fatalf("zurücknehmen: %v", err)
	}

	if !genommen {
		t.Fatal("nichts zurückgenommen")
	}

	karten, err := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if wert, _ := karten[0].Value("name"); wert != "Feuerball" {
		t.Errorf("Wert %q, erwartet Feuerball", wert)
	}
}

// **Wiederherstellen macht die Rücknahme rückgängig.**
func TestWiederherstellenBringtDieAenderungZurueck(t *testing.T) {
	db, setID, karteID := bestandMitKarte(t)

	schreiber := storage.NeuerKartenschreiber(db)

	if _, _, err := schreiber.FeldSetzenMitVorwert(
		probekontext(t), karteID, "name", "Eisball"); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	leser := storage.NeuerVerlaufsleser(db)

	if _, _, err := leser.Zuruecknehmen(probekontext(t), setID); err != nil {
		t.Fatalf("zurücknehmen: %v", err)
	}

	_, gebracht, err := leser.Wiederherstellen(probekontext(t), setID)
	if err != nil {
		t.Fatalf("wiederherstellen: %v", err)
	}

	if !gebracht {
		t.Fatal("nichts wiederhergestellt")
	}

	karten, err := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if wert, _ := karten[0].Value("name"); wert != "Eisball" {
		t.Errorf("Wert %q, erwartet Eisball", wert)
	}
}

// **Ein entferntes Feld kommt beim Zurücknehmen wieder** — und zwar als Feld,
// nicht als leerer Wert.
func TestEinEntferntesFeldKommtZurueck(t *testing.T) {
	db, setID, karteID := bestandMitKarte(t)

	if _, _, _, err := storage.NeuerKartenschreiber(db).
		FeldEntfernen(probekontext(t), karteID, "name"); err != nil {
		t.Fatalf("entfernen: %v", err)
	}

	if _, _, err := storage.NeuerVerlaufsleser(db).Zuruecknehmen(probekontext(t), setID); err != nil {
		t.Fatalf("zurücknehmen: %v", err)
	}

	karten, err := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	wert, gabEs := karten[0].Value("name")
	if !gabEs || wert != "Feuerball" {
		t.Errorf("Wert %q, gab es: %v", wert, gabEs)
	}
}

// **Der Verlauf wächst nicht unbegrenzt.**
//
// `GrenzeVerlauf` ist die benannte Zahl, die #52 verlangt: was darüber
// hinausgeht, fällt vorn weg. Ein Verlauf, der mit jeder Sitzung wächst, macht
// die Datenbank zum Protokoll.
func TestDerVerlaufWaechstNichtUnbegrenzt(t *testing.T) {
	db, setID, karteID := bestandMitKarte(t)

	schreiber := storage.NeuerKartenschreiber(db)

	for lauf := range storage.GrenzeVerlauf + 5 {
		if _, _, err := schreiber.FeldSetzenMitVorwert(
			probekontext(t), karteID, "name", string(rune('a'+lauf%26))); err != nil {
			t.Fatalf("setzen: %v", err)
		}
	}

	verlauf, err := storage.NeuerVerlaufsleser(db).Verlauf(probekontext(t), setID)
	if err != nil {
		t.Fatalf("Verlauf: %v", err)
	}

	if len(verlauf) != storage.GrenzeVerlauf {
		t.Errorf("%d Einträge, erwartet %d", len(verlauf), storage.GrenzeVerlauf)
	}
}

// **Ein Feld, das es vorher nicht gab, wird beim Zurücknehmen entfernt.**
//
// Nicht geleert — entfernt. Eine leere Zeile ist nicht dasselbe wie keine
// Zeile, und wer das zusammenwirft, gibt beim Speichern ein Feld aus, das
// niemand angelegt hat (0010).
func TestEinNeuesFeldWirdBeimZuruecknehmenEntfernt(t *testing.T) {
	db, setID, karteID := bestandMitKarte(t)

	if _, _, err := storage.NeuerKartenschreiber(db).
		FeldSetzenMitVorwert(probekontext(t), karteID, "regeltext", "Fügt Schaden zu."); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if _, _, err := storage.NeuerVerlaufsleser(db).Zuruecknehmen(probekontext(t), setID); err != nil {
		t.Fatalf("zurücknehmen: %v", err)
	}

	karten, err := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if _, gabEs := karten[0].Value("regeltext"); gabEs {
		t.Error("das Feld steht noch da — geleert statt entfernt")
	}
}
