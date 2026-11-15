// **Zuletzt benutzt** —
// #891,
// Werkstück 3 von Phase Y.
//
// # Die Regel kommt aus MSE2
//
// `Settings::addRecentFile` (`src/data/settings.cpp`, Fassung f9d9356 vom
// 2021-07-06, Zeilen 186–200): absoluten Pfad bilden, **Doppel entfernen**,
// vorn einfügen, auf `max_recent_sets` = 9 kürzen.
//
// **Hier ist die Liste keine Liste**, sondern die Ordnung des Bestands:
// `sets.geoeffnet` trägt, wann ein Set zuletzt benutzt wurde, und
// `Setkoepfe` sortiert danach. Damit fallen Doppel und Kürzung von selbst
// weg — ein Set steht einmal im Bestand, und wer die ersten neun nimmt, hat
// die neun jüngsten.
//
// **Was dafür fehlte**: die Spalte wurde beim Anlegen geschrieben und nie
// wieder. Sie sagte „zuletzt **eingelesen**", nicht „zuletzt benutzt".
package storage_test

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein benutztes Set wandert nach vorn.**
func TestEinBenutztesSetWandertNachVorn(t *testing.T) {
	t.Parallel()

	db := frischeDateidatenbank(t)

	// **Eine Uhr, die weiterläuft.** `Zeitstempel` hat
	// Millisekundenauflösung; zwei Aufrufe hintereinander landen sonst im
	// selben Augenblick, und dann entscheidet die Kennung statt der Zeit —
	// der Lauf prüfte die Einfügereihenfolge statt der Benutzung.
	uhrzeit := time.Date(2026, time.August, 25, 12, 0, 0, 0, time.UTC)
	speicher := storage.NeuerSetSpeicher(db).MitUhr(func() time.Time {
		uhrzeit = uhrzeit.Add(time.Second)

		return uhrzeit
	})

	erstes := setMitNamen(t, speicher, "Zuerst")
	zweites := setMitNamen(t, speicher, "Danach")

	leser := storage.NeuerSetLeser(db)

	// **Das zuletzt angelegte steht vorn** — das galt schon vorher.
	koepfe, err := leser.Setkoepfe(probekontext(t))
	if err != nil {
		t.Fatalf("Setkoepfe: %v", err)
	}

	if koepfe[0].ID != zweites {
		t.Fatalf("vorn steht %d, erwartet war %d", koepfe[0].ID, zweites)
	}

	// **Und jetzt wird das erste benutzt.**
	if fehler := speicher.SetBenutzt(probekontext(t), erstes); fehler != nil {
		t.Fatalf("SetBenutzt: %v", fehler)
	}

	koepfe, err = leser.Setkoepfe(probekontext(t))
	if err != nil {
		t.Fatalf("Setkoepfe: %v", err)
	}

	if koepfe[0].ID != erstes {
		t.Errorf("vorn steht %d, erwartet war %d", koepfe[0].ID, erstes)
	}
}

// **Ein Set, das es nicht gibt, meldet sich.**
//
// Sonst sähe ein Aufruf auf eine gelöschte Kennung aus wie ein Erfolg, und
// die Liste bliebe, wie sie war — ohne dass jemand erführe, warum.
func TestEinUnbekanntesSetLaesstSichNichtBenutzen(t *testing.T) {
	t.Parallel()

	speicher := storage.NeuerSetSpeicher(frischeDateidatenbank(t))

	if err := speicher.SetBenutzt(probekontext(t), 4242); err == nil {
		t.Error("eine unbekannte Kennung kam durch")
	}
}

// setMitNamen legt ein Set an und gibt seine Kennung.
func setMitNamen(t *testing.T, speicher *storage.SetSpeicher, name string) int64 {
	t.Helper()

	kennung, err := speicher.SetAnlegen(probekontext(t),
		&domain.Set{Name: name, Game: "probe"}, nil)
	if err != nil {
		t.Fatalf("SetAnlegen: %v", err)
	}

	return kennung
}
