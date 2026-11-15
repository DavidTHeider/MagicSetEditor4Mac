package storage_test

// **Wer gerade arbeitet** — Werkstück 2 von Phase 4.12
// (#1351),
// nach Vermerk 0066.
//
// # Warum der Zeiger vor der Trennung kommt
//
// Die Trennung klemmt jede Abfrage auf ein Profil. **Solange niemand sagt,
// welches**, gibt es nichts zu klemmen — der Zeiger ist die Voraussetzung und
// nicht das Beiwerk.
//
// # Und warum er im Bestand liegt
//
// Weil ein Neustart ihn sonst vergäße. Wer sein Profil wählt und die
// Anwendung schließt, findet es beim nächsten Mal wieder — sonst wäre die
// Wahl eine Frage, die jeden Morgen neu gestellt wird.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein frischer Bestand zeigt auf das erste Profil.**
//
// Nicht auf keines: eine Anwendung, die beim ersten Start nach einem Profil
// fragt, wo vorher ein Fenster aufging, hat eine Frage eingeführt und eine
// Fähigkeit genommen.
func TestEinFrischerBestandZeigtAufDasErsteProfil(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	zeiger, err := ablage.Profilzeiger(probekontext(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if zeiger == 0 {
		t.Fatal("der Zeiger steht auf keinem Profil")
	}

	nutzer, lesefehler := ablage.Nutzer(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("die Nutzer lesen: %v", lesefehler)
	}

	profile, profilfehler := ablage.Profile(probekontext(t), nutzer[0].ID)
	if profilfehler != nil {
		t.Fatalf("die Profile lesen: %v", profilfehler)
	}

	if zeiger != profile[0].ID {
		t.Errorf("der Zeiger steht auf %d, das erste Profil ist %d", zeiger, profile[0].ID)
	}
}

// **Der Zeiger lässt sich umstellen und überlebt.**
func TestDerProfilzeigerLaesstSichUmstellen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueNutzerablage(db)

	nutzer, err := ablage.Nutzer(probekontext(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	zweites, anlegefehler := ablage.ProfilAnlegen(probekontext(t), domain.Profil{
		NutzerID: nutzer[0].ID, Name: "Turnier",
	})
	if anlegefehler != nil {
		t.Fatalf("anlegen: %v", anlegefehler)
	}

	if err := ablage.ProfilzeigerSetzen(probekontext(t), zweites.ID); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	// **Eine neue Ablage über derselben Datei** — der Zeiger liegt im
	// Bestand und nicht im Arbeitsspeicher.
	gelesen, lesefehler := storage.NeueNutzerablage(db).Profilzeiger(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if gelesen != zweites.ID {
		t.Errorf("der Zeiger steht auf %d, erwartet war %d", gelesen, zweites.ID)
	}
}

// **Ein Profil, das es nicht gibt, wird abgewiesen.**
//
// Sonst zeigte der Bestand auf nichts, und **jede** Abfrage käme leer zurück
// — eine Anwendung, die aussieht, als wären alle Daten weg.
func TestEinUnbekanntesProfilWirdAlsZeigerAbgewiesen(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))

	if err := ablage.ProfilzeigerSetzen(probekontext(t), 999); !errors.Is(
		err, domain.ErrNotFound) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}
}

// **Und ein gelöschter Nutzer nimmt den Zeiger nicht mit ins Nichts.**
//
// Die Kaskade räumt seine Profile; zeigte der Bestand auf eines davon, stünde
// er danach auf einer Kennung, die es nicht mehr gibt. Der Zeiger fällt dann
// auf das erste vorhandene zurück.
func TestEinGeloeschtesProfilLaesstDenZeigerNichtLeer(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueNutzerablage(db)

	zweiter, err := ablage.NutzerAnlegen(probekontext(t), domain.Nutzer{Name: "Zweite"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	profile, lesefehler := ablage.Profile(probekontext(t), zweiter.ID)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if err := ablage.ProfilzeigerSetzen(probekontext(t), profile[0].ID); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	if _, err := db.ExecContext(probekontext(t),
		`DELETE FROM nutzer WHERE id = ?`, zweiter.ID); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	zeiger, zeigerfehler := ablage.Profilzeiger(probekontext(t))
	if zeigerfehler != nil {
		t.Fatalf("lesen: %v", zeigerfehler)
	}

	if zeiger == 0 || zeiger == profile[0].ID {
		t.Errorf("der Zeiger steht auf %d", zeiger)
	}
}
