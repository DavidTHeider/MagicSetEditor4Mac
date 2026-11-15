package storage_test

// **Die Abnahme der Zusagen** — Werkstück `T2` von Phase 5.6
// (#1467).
//
// # Was hier steht und was nebenan
//
// **Nebenan stehen die Regeln je für sich** — `reservierungen_test.go` für das
// Festhalten, `zusagenschutz_test.go` für den Schreibweg. Hier stehen die
// **Abnahmepunkte des Issues**, und zwar als Lauf: drei Handelnde an einer
// Datei, der Fassungskonflikt, und der Neustart.
//
// # Was die Läufe nicht belegen
//
// **Zwei Prozesse.** Gemessen wird eine Datei mit einer Verbindung; dass
// `_txlock=immediate` zwei Schreiber serialisiert, steht in
// `datenbank.go` und hat dort seinen eigenen Grund. Was hier gezeigt wird,
// ist, dass die **Rechnung** stimmt, wenn drei Wege dieselbe Zeile anfassen.

import (
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Abnahmepunkt 1: zwei Handelsvorgänge und ein Schreibweg an einer Datei
// halten nie mehr fest als den Besitz.**
func TestZweiVorgaengeUndEinSchreibwegHaltenNichtMehrAlsDenBesitz(t *testing.T) {
	db := frischeDatenbank(t)
	ctx := probekontext(t)

	speicher := storage.NeuerSammlungsspeicher(db)

	id, err := speicher.Anlegen(ctx, probeexemplar())
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	ablage := storage.NeueReservierungsablage(db)

	// **Drei Handelnde auf einmal**: zwei sagen je zwei Stück zu, der dritte
	// senkt die Menge auf eines. Der Besitz sind drei.
	var warten sync.WaitGroup

	warten.Add(3)

	for _, vorgang := range []string{"h1", "h2"} {
		go func() {
			defer warten.Done()

			// **Das Ergebnis wird nicht geprüft, und das ist der Punkt**:
			// welcher der drei durchkommt, entscheidet die Sperre. Geprüft
			// wird die **Rechnung danach**.
			//nolint:errcheck // siehe darüber
			_ = ablage.Festhalten(ctx, domain.Reservierung{
				Vorgang: vorgang, Revision: 1, ExemplarID: id, Menge: 2,
				Fassung: 1,
			}, ueber(probeexemplar(), 99))
		}()
	}

	go func() {
		defer warten.Done()

		knapp := probeexemplar()
		knapp.ID, knapp.Fassung, knapp.Menge = id, 1, 1

		//nolint:errcheck // derselbe Grund wie oben
		_ = speicher.Aendern(ctx, knapp)
	}()

	warten.Wait()

	// **Die Rechnung am Ende**: festgehalten ist nie mehr, als dasteht.
	festgehalten, zaehlfehler := ablage.ReserviertJeExemplar(ctx)
	if zaehlfehler != nil {
		t.Fatalf("zählen: %v", zaehlfehler)
	}

	stand, lesefehler := speicher.Einzeln(ctx, id)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if festgehalten[id] > stand.Menge {
		t.Errorf("festgehalten sind %d Stück, im Bestand stehen %d",
			festgehalten[id], stand.Menge)
	}
}

// **Abnahmepunkt 2: drei Stück, zwei zugesagt, die freie Menge um eines
// geändert — der alte Stand schreibt danach nichts.**
func TestNachEinerMengenaenderungSchreibtDerAlteStandNichts(t *testing.T) {
	db := frischeDatenbank(t)
	ctx := probekontext(t)

	speicher := storage.NeuerSammlungsspeicher(db)

	id, err := speicher.Anlegen(ctx, probeexemplar())
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	ablage := storage.NeueReservierungsablage(db)

	if fehler := ablage.Festhalten(ctx, domain.Reservierung{
		Vorgang: "h1", Revision: 1, ExemplarID: id, Menge: 2, Fassung: 1,
	}, ueber(probeexemplar(), 99)); fehler != nil {
		t.Fatalf("festhalten: %v", fehler)
	}

	// **Die freie Menge fällt weg** — drei werden zwei, und die beiden
	// zugesagten bleiben stehen. Das darf gehen (0120, Satz 2).
	knapp := probeexemplar()
	knapp.ID, knapp.Fassung, knapp.Menge = id, 1, 2

	if fehler := speicher.Aendern(ctx, knapp); fehler != nil {
		t.Fatalf("ändern: %v", fehler)
	}

	// **Der Handel rechnet jetzt gegen einen Stand, den es nicht mehr gibt.**
	// Eine zweite Zusage gegen die alte Fassung wird abgewiesen — der
	// Vorgang muss neu prüfen und neu zustimmen lassen.
	if fehler := ablage.Festhalten(ctx, domain.Reservierung{
		Vorgang: "h2", Revision: 1, ExemplarID: id, Menge: 1, Fassung: 1,
	}, ueber(probeexemplar(), 99)); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("die Zusage gegen die alte Fassung kam mit %v durch", fehler)
	}

	// **Und der alte Commit schreibt nichts**: dieselbe Fassung ein zweites
	// Mal trifft keine Zeile.
	if fehler := speicher.Aendern(ctx, knapp); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("der alte Stand kam mit %v durch", fehler)
	}
}

// **Abnahmepunkt 3: ein Neustart erhält die Zusagen.**
//
// **Gemessen wird mit einer zweiten Verbindung auf dieselbe Datei** — das ist,
// was ein Neustart tut. Ein Lauf über dieselbe offene Verbindung prüfte den
// Arbeitsspeicher und nicht die Platte.
func TestEinNeustartErhaeltDieZusagen(t *testing.T) {
	ort := filepath.Join(t.TempDir(), "bestand.db")
	ctx := probekontext(t)

	vorher, oeffnungsfehler := storage.Oeffnen(ctx, ort)
	if oeffnungsfehler != nil {
		t.Fatalf("öffnen: %v", oeffnungsfehler)
	}

	if fehler := storage.Migriere(ctx, vorher, nil); fehler != nil {
		t.Fatalf("wandern: %v", fehler)
	}

	id, err := storage.NeuerSammlungsspeicher(vorher).Anlegen(ctx, probeexemplar())
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if fehler := storage.NeueReservierungsablage(vorher).Festhalten(ctx,
		domain.Reservierung{
			Vorgang: "h1", Revision: 1, ExemplarID: id, Menge: 2, Fassung: 1,
		}, ueber(probeexemplar(), 99)); fehler != nil {
		t.Fatalf("festhalten: %v", fehler)
	}

	if fehler := vorher.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	// **Wie ein Neustart**: dieselbe Datei, neu geöffnet und neu gewandert.
	// Die Wanderung ist Teil des Starts, und ein Lauf, der sie ausließe,
	// prüfte einen Weg, den es im Betrieb nicht gibt.
	nachher, zweitfehler := storage.Oeffnen(ctx, ort)
	if zweitfehler != nil {
		t.Fatalf("wieder öffnen: %v", zweitfehler)
	}

	if fehler := storage.Migriere(ctx, nachher, nil); fehler != nil {
		t.Fatalf("wieder wandern: %v", fehler)
	}

	t.Cleanup(func() { _ = nachher.Close() })

	festgehalten, zaehlfehler := storage.NeueReservierungsablage(nachher).
		ReserviertJeExemplar(ctx)
	if zaehlfehler != nil {
		t.Fatalf("zählen: %v", zaehlfehler)
	}

	if festgehalten[id] != 2 {
		t.Errorf("nach dem Neustart sind %d Stück festgehalten", festgehalten[id])
	}
}
