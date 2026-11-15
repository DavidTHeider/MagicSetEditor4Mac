package storage_test

// **Der Vermerk über die eigene Peerkennung** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Was hier gemessen wird
//
// **Dass ein Verlust auffällt.** Der private Schlüssel liegt im Schlüsselbund,
// die Kennung im Bestand — und erst beide zusammen sagen, ob dieses Profil
// noch dieselbe Identität hat. Ohne den Vermerk wäre „der Bund hat nichts"
// nicht von „hatte noch nie einen" zu unterscheiden.
//
// **Und dass eine Rotation die alte Kennung behält.** Ein Kontakt kennt sie;
// wer sie wegwirft, kann ihm nicht mehr sagen, welche Kennung abgelöst wurde.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// derPeerlauf ist der Zeitpunkt, an dem diese Läufe spielen.
var derPeerlauf = time.Date(2026, 9, 14, 3, 0, 0, 0, time.UTC)

// **Ohne Vermerk steht nichts da — und das ist keine Null, sondern leer.**
func TestOhneVermerkStehtNichtsDa(t *testing.T) {
	db := frischeDatenbank(t)

	vermerk, fehler := storage.NeuePeerablage(db).Lesen(probekontext(t))
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if vermerk.Kennung != "" || vermerk.Vorher != "" {
		t.Fatalf("gelesen wurde %+v", vermerk)
	}
}

// **Was gemerkt wurde, steht danach da.**
func TestEineGemerkteKennungStehtDanachDa(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeuePeerablage(db)

	if fehler := ablage.Merken(
		probekontext(t), "bkennungeins", derPeerlauf); fehler != nil {
		t.Fatalf("Merken: %v", fehler)
	}

	vermerk, fehler := ablage.Lesen(probekontext(t))
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if vermerk.Kennung != "bkennungeins" {
		t.Errorf("gelesen wurde %q", vermerk.Kennung)
	}

	if !vermerk.Seit.Equal(derPeerlauf) {
		t.Errorf("der Zeitpunkt ist %s", vermerk.Seit)
	}

	// **Ohne Rotation gibt es kein Vorher** — eine erfundene Vorgängerin
	// verlangte von jedem Kontakt eine Prüfung, die niemand veranlasst hat.
	if vermerk.Vorher != "" {
		t.Errorf("das Vorher ist %q", vermerk.Vorher)
	}
}

// **Eine Rotation hebt die alte Kennung nach vorher.**
func TestEineRotationBehaeltDieAlteKennung(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeuePeerablage(db)

	if fehler := ablage.Merken(
		probekontext(t), "balt", derPeerlauf); fehler != nil {
		t.Fatalf("Merken: %v", fehler)
	}

	spaeter := derPeerlauf.Add(time.Hour)

	if fehler := ablage.Rotieren(
		probekontext(t), "bneu", spaeter); fehler != nil {
		t.Fatalf("Rotieren: %v", fehler)
	}

	vermerk, fehler := ablage.Lesen(probekontext(t))
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if vermerk.Kennung != "bneu" || vermerk.Vorher != "balt" {
		t.Fatalf("gelesen wurde %+v", vermerk)
	}

	if !vermerk.Seit.Equal(spaeter) {
		t.Errorf("der Zeitpunkt ist %s", vermerk.Seit)
	}
}

// **Und eine zweite Rotation behält die zuletzt abgelöste.**
//
// **Die Gegenprobe** (§10): ein `vorher`, das die **erste** Kennung hielte,
// zeigte einem Kontakt eine Vorgängerin, die er nie gesehen hat.
func TestEineZweiteRotationBehaeltDieZuletztAbgeloeste(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeuePeerablage(db)

	if fehler := ablage.Merken(
		probekontext(t), "beins", derPeerlauf); fehler != nil {
		t.Fatalf("Merken: %v", fehler)
	}

	for _, kennung := range []string{"bzwei", "bdrei"} {
		if fehler := ablage.Rotieren(
			probekontext(t), kennung, derPeerlauf); fehler != nil {
			t.Fatalf("Rotieren nach %s: %v", kennung, fehler)
		}
	}

	vermerk, fehler := ablage.Lesen(probekontext(t))
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if vermerk.Kennung != "bdrei" || vermerk.Vorher != "bzwei" {
		t.Fatalf("gelesen wurde %+v", vermerk)
	}
}

// **Eine Rotation ohne Vorgängerin ist ein erstes Merken.**
//
// **Nicht ein Fehler**: wer einen Schlüssel verloren hat und einen neuen
// anlegt, rotiert — und es gibt nichts, was abgelöst würde.
func TestEineRotationOhneVorgaengerinIstEinErstesMerken(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeuePeerablage(db)

	if fehler := ablage.Rotieren(
		probekontext(t), "bneu", derPeerlauf); fehler != nil {
		t.Fatalf("Rotieren: %v", fehler)
	}

	vermerk, fehler := ablage.Lesen(probekontext(t))
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if vermerk.Kennung != "bneu" || vermerk.Vorher != "" {
		t.Fatalf("gelesen wurde %+v", vermerk)
	}
}

// **Jedes Profil hat seine eigene.**
//
// **Der Fund**: eine gemeinsame Kennung machte aus zwei Menschen an einem
// Rechner ein Gegenüber — und jede Freigabe, die einer gibt, gälte dem
// anderen mit.
func TestJedesProfilHatSeineEigeneKennung(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeuePeerablage(db)

	if fehler := ablage.Merken(
		probekontext(t), "berste", derPeerlauf); fehler != nil {
		t.Fatalf("Merken: %v", fehler)
	}

	zweites := storage.MitProfil(t.Context(),
		zweitesProfil(t, storage.NeueNutzerablage(db), "Turnier"))

	vermerk, fehler := ablage.Lesen(zweites)
	if fehler != nil {
		t.Fatalf("Lesen im zweiten Profil: %v", fehler)
	}

	if vermerk.Kennung != "" {
		t.Fatalf("das zweite Profil trägt %q", vermerk.Kennung)
	}
}

// **Eine leere Kennung wird nicht gemerkt.**
//
// **Sie sähe aus wie „es gab noch keine"** — und der nächste Start legte
// still eine neue Identität an, obwohl eine da war.
func TestEineLeereKennungWirdNichtGemerkt(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeuePeerablage(db)

	if fehler := ablage.Merken(
		probekontext(t), "", derPeerlauf); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if fehler := ablage.Rotieren(
		probekontext(t), "", derPeerlauf); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("beim Rotieren ist der Fehler %v", fehler)
	}
}

// **Ohne Profil im Kontext geht weder Lesen noch Schreiben.**
func TestOhneProfilGehtDerVermerkNicht(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeuePeerablage(db)

	if _, fehler := ablage.Lesen(t.Context()); fehler == nil {
		t.Error("ohne Profil kam ein Vermerk")
	}

	if fehler := ablage.Merken(
		t.Context(), "bwas", derPeerlauf); fehler == nil {
		t.Error("ohne Profil wurde gemerkt")
	}
}

// **Ein krummer Zeitpunkt im Bestand ist ein Befund und kein leerer Vermerk.**
//
// **Der Weg dorthin ist eine Hand an der Datei** — und genau deshalb steht er
// hier: wer sie anfasst, soll es erfahren, statt eine Identität ohne Datum zu
// bekommen.
func TestEinKrummerZeitpunktIstEinBefund(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeuePeerablage(db)

	if fehler := ablage.Merken(
		probekontext(t), "bwas", derPeerlauf); fehler != nil {
		t.Fatalf("Merken: %v", fehler)
	}

	if _, err := db.ExecContext(t.Context(),
		`UPDATE peeridentitaeten SET seit = 'gestern'`); err != nil {
		t.Fatalf("verbiegen: %v", err)
	}

	if _, fehler := ablage.Lesen(probekontext(t)); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Und ohne Tabelle meldet sich der Bestand, statt leer zu antworten.**
func TestOhnePeertabelleMeldetSichDerBestand(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeuePeerablage(db)

	if _, err := db.ExecContext(t.Context(),
		`DROP TABLE peeridentitaeten`); err != nil {
		t.Fatalf("fallen lassen: %v", err)
	}

	if _, fehler := ablage.Lesen(probekontext(t)); fehler == nil {
		t.Error("ohne Tabelle kam ein Vermerk")
	}

	if fehler := ablage.Merken(
		probekontext(t), "bwas", derPeerlauf); fehler == nil {
		t.Error("ohne Tabelle wurde gemerkt")
	}
}
