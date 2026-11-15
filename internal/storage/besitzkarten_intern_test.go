package storage

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eine Karte oder ein Set aus einem fremden Profil trifft nichts** —
// #1815.
//
// # Der Befund
//
// Nachgestellt am 2026-09-23 an der Kommandozeile: `mse feld 1 1 name
// Übernommen` schrieb im eigenen Profil den Kartennamen eines anderen
// Nutzers um. Die Karte ist ein Abkömmling ihres Sets (0035), und keine
// dieser Funktionen fragte, ob das Set dem gezeigten Profil gehört.
//
// # Was hier steht
//
// **Jede Funktion der Karten- und Setablagen, die eine Kennung annimmt**,
// gerufen aus Profil 1 mit den Kennungen von Profil 2 — lesend und
// schreibend. Erwartet: „nicht gefunden", und Profil 2 steht danach wie
// vorher. Dass keine fehlt, hält `besitzvollstaendig_intern_test.go`.
func TestEineFremdeKarteTrifftNichts(t *testing.T) {
	t.Parallel()

	db, fremd := besitzbestand(t)

	ausFremdemProfil(t, db, kartenrufe(db, fremd, MitProfil(t.Context(), 1)))
}

// kartenrufe sind die Rufe der Karten- und Setablagen, nach `Typ.Methode`.
func kartenrufe(db *sql.DB, fremd fremdeKennungen, eigen context.Context) map[string]func() error {
	schreiber := NeuerKartenschreiber(db)
	leser := NeuerSetLeser(db)
	speicher := NeuerSetSpeicher(db)
	marken := NeueMarken(db)
	verweise := NeueKartenverweise(db)
	verlauf := NeuerVerlaufsleser(db)
	stil := NeueStilablage(db)
	worte := NeueSchluesselwortablage(db)
	name := []domain.Value{{Field: "name", Text: "X"}}
	herkunft := []domain.KIFeldherkunft{{
		Feld: "name", Anbieter: "ollama", Modell: "gemma4:8b-q4",
		Aufgabe: "Schreibe den Namen.", Auftrag: "ki-1815",
		Zeitpunkt: time.Date(2026, 9, 23, 9, 0, 0, 0, time.UTC),
		Bezug:     domain.KIInhaltsbezug([]string{"Fremde Karte"}),
	}}

	return map[string]func() error{
		"Kartenschreiber.FeldSetzen": func() error { return schreiber.FeldSetzen(eigen, fremd.karte, "name", "X") },
		"Kartenschreiber.FeldSetzenMitVorwert": func() error {
			_, _, f := schreiber.FeldSetzenMitVorwert(eigen, fremd.karte, "name", "X")
			return f
		},
		"Kartenschreiber.FeldSetzenAufStand": func() error {
			_, _, f := schreiber.FeldSetzenAufStand(eigen, fremd.karte, "name", "X", "")
			return f
		},
		"Kartenschreiber.FelderSetzenMitVorwert": func() error {
			_, _, f := schreiber.FelderSetzenMitVorwert(eigen, fremd.karte, name)
			return f
		},
		"Kartenschreiber.FeldSetzenAnStelle": func() error {
			_, _, f := schreiber.FeldSetzenAnStelle(eigen, fremd.karte, "name", "X", 0)
			return f
		},
		"Kartenschreiber.FeldEntfernen": func() error {
			_, _, _, f := schreiber.FeldEntfernen(eigen, fremd.karte, "name")
			return f
		},
		"Kartenschreiber.FelderSetzenVonAussen": func() error {
			_, _, f := schreiber.FelderSetzenVonAussen(eigen, fremd.karte, name, "x", "")
			return f
		},
		"Kartenschreiber.FeldSetzenVonAussen": func() error {
			_, _, f := schreiber.FeldSetzenVonAussen(eigen, fremd.karte, "name", "X", "x")
			return f
		},
		"Kartenschreiber.FelderSetzenMitHerkunft": func() error {
			_, _, f := schreiber.FelderSetzenMitHerkunft(eigen, fremd.karte, name, "ki", herkunft, nil, "")
			return f
		},
		"Kartenschreiber.GruppeSetzenMitHerkunft": func() error {
			return schreiber.GruppeSetzenMitHerkunft(eigen, []domain.Kartenschritt{{
				KarteID: fremd.karte, Werte: name, Herkunft: "ki", Herkuenfte: herkunft,
			}})
		},
		"Kartenschreiber.KarteKopieren":  func() error { _, f := schreiber.KarteKopieren(eigen, fremd.karte); return f },
		"Kartenschreiber.KarteLoeschen":  func() error { return schreiber.KarteLoeschen(eigen, fremd.karte) },
		"Kartenschreiber.KartenLoeschen": func() error { return schreiber.KartenLoeschen(eigen, []int64{fremd.karte}) },
		"Kartenschreiber.KarteAnlegen":   func() error { _, f := schreiber.KarteAnlegen(eigen, fremd.set); return f },
		"Kartenschreiber.KartenAnlegen": func() error {
			_, f := schreiber.KartenAnlegen(eigen, fremd.set, [][]domain.Feldstand{{{Feld: "name", Wert: "X"}}})
			return f
		},
		"Kartenschreiber.KartenTauschen": func() error { return schreiber.KartenTauschen(eigen, fremd.karte, fremd.karte+1) },
		"Kartenschreiber.WerteSetzen": func() error {
			return schreiber.WerteSetzen(eigen, []domain.Feldwechsel{{KarteID: fremd.karte, Feld: "name", Wert: "X"}})
		},
		"Kartenschreiber.FelderSetzenMitErwartung": func() error {
			_, f := schreiber.FelderSetzenMitErwartung(eigen,
				[]domain.Feldvorwert{{KarteID: fremd.karte, Wert: "Fremde Karte", GabEs: true}}, "name", "X")
			return f
		},
		"Kartenschreiber.FelderSetzen": func() error {
			_, f := schreiber.FelderSetzen(eigen, []int64{fremd.karte}, "name", "X")
			return f
		},
		"Kartenschreiber.FelderGerechnetSetzen": func() error {
			_, f := schreiber.FelderGerechnetSetzen(eigen,
				[]domain.Kartenschritt{{KarteID: fremd.karte, Werte: []domain.Value{{Field: "name", Text: "X"}}}}, nil, "name")
			return f
		},
		"Kartenschreiber.KarteAnlegenMitWerten": func() error {
			_, f := schreiber.KarteAnlegenMitWerten(eigen, fremd.set, []domain.Value{{Field: "name", Text: "X"}})
			return f
		},
		"Kartenschreiber.SetwertSetzenMitVorwert": func() error {
			_, _, f := schreiber.SetwertSetzenMitVorwert(eigen, fremd.set, "title", "X")
			return f
		},
		"SetLeser.NameZurKarte":      func() error { _, f := leser.NameZurKarte(eigen, fremd.karte); return f },
		"SetLeser.FeldZurKarte":      func() error { _, f := leser.FeldZurKarte(eigen, fremd.karte, "name"); return f },
		"SetLeser.SpielZurKarte":     func() error { _, f := leser.SpielZurKarte(eigen, fremd.karte); return f },
		"SetLeser.Kartenfeldwerte":   func() error { _, f := leser.Kartenfeldwerte(eigen, fremd.set, "name"); return f },
		"SetLeser.KarteLesen":        func() error { _, f := leser.KarteLesen(eigen, fremd.karte); return f },
		"SetLeser.Setstand":          func() error { _, f := leser.Setstand(eigen, fremd.set); return f },
		"SetLeser.Kartenzeilen":      func() error { _, f := leser.Kartenzeilen(eigen, fremd.set); return f },
		"SetLeser.Karten":            func() error { _, f := leser.Karten(eigen, fremd.set); return f },
		"SetLeser.Rohbloecke":        func() error { _, f := leser.Rohbloecke(eigen, fremd.set); return f },
		"SetLeser.Herkuenfte":        func() error { _, f := leser.Herkuenfte(eigen, fremd.karte); return f },
		"SetLeser.Setpackarten":      func() error { _, f := leser.Setpackarten(eigen, fremd.set); return f },
		"SetLeser.Schluesselwoerter": func() error { _, f := leser.Schluesselwoerter(eigen, fremd.set); return f },
		"SetLeser.Setwerte":          func() error { _, f := leser.Setwerte(eigen, fremd.set); return f },
		"SetLeser.Kartenstile":       func() error { _, f := leser.Kartenstile(eigen, fremd.set); return f },
		"SetLeser.Kennzahlen":        func() error { _, f := leser.Kennzahlen(eigen, fremd.set); return f },
		"SetLeser.Belegung":          func() error { _, f := leser.Belegung(eigen, fremd.set); return f },
		"SetLeser.Verteilung":        func() error { _, f := leser.Verteilung(eigen, fremd.set, "name"); return f },
		"SetLeser.HerkunftSetzen":    func() error { return leser.HerkunftSetzen(eigen, fremd.set, "/fremd") },
		"SetLeser.Setkopf":           func() error { _, f := leser.Setkopf(eigen, fremd.set); return f },
		"SetLeser.Setstil":           func() error { _, f := leser.Setstil(eigen, fremd.set); return f },
		"SetLeser.Setzeile":          func() error { _, f := leser.Setzeile(eigen, fremd.set); return f },
		"SetSpeicher.SetBenutzt":     func() error { return speicher.SetBenutzt(eigen, fremd.set) },
		"SetSpeicher.SetLoeschen":    func() error { return speicher.SetLoeschen(eigen, fremd.set, "Fremdes Set") },
		"Marken.Setzen":              func() error { return marken.Setzen(eigen, fremd.karte, "neu") },
		"Marken.Entfernen":           func() error { return marken.Entfernen(eigen, fremd.karte, "fremd") },
		"Marken.AnKarte":             func() error { _, f := marken.AnKarte(eigen, fremd.karte); return f },
		"Marken.ImSet":               func() error { _, f := marken.ImSet(eigen, fremd.set); return f },
		"Marken.JeKarte":             func() error { _, f := marken.JeKarte(eigen, fremd.set); return f },
		"Kartenverweise.Setzen": func() error {
			return verweise.Setzen(eigen, fremd.karte, domain.Kartenverweis{Printing: "x", Name: "X"})
		},
		"Kartenverweise.Fuer":   func() error { _, f := verweise.Fuer(eigen, fremd.karte); return f },
		"Kartenverweise.Loesen": func() error { _, f := verweise.Loesen(eigen, fremd.karte); return f },
		"Kartenverweise.JeSet":  func() error { _, f := verweise.JeSet(eigen, fremd.set); return f },
		"Verlaufsleser.Verlauf": func() error { _, f := verlauf.Verlauf(eigen, fremd.set); return f },
		"Verlaufsleser.Zuruecknehmen": func() error {
			_, _, f := verlauf.Zuruecknehmen(eigen, fremd.set)
			return f
		},
		"Verlaufsleser.Wiederherstellen": func() error {
			_, _, f := verlauf.Wiederherstellen(eigen, fremd.set)
			return f
		},
		"Verlaufsleser.VorgangZuruecknehmen": func() error {
			_, _, f := verlauf.VorgangZuruecknehmen(eigen, fremd.set, "v")
			return f
		},
		"Verlaufsleser.VorgangWiederherstellen": func() error {
			_, _, f := verlauf.VorgangWiederherstellen(eigen, fremd.set, "v")
			return f
		},
		"Stilablage.Stilwerte":           func() error { _, f := stil.Stilwerte(eigen, fremd.set); return f },
		"SetLeser.Stilwerte":             func() error { _, f := leser.Stilwerte(eigen, fremd.set); return f },
		"Stilablage.StilSetzen":          func() error { return stil.StilSetzen(eigen, fremd.set, "magic", "m15", nil) },
		"Stilablage.StiloptionSetzen":    func() error { return stil.StiloptionSetzen(eigen, fremd.set, "rahmen", "blau") },
		"Stilablage.StiloptionEntfernen": func() error { return stil.StiloptionEntfernen(eigen, fremd.set, "rahmen") },
		"Schluesselwortablage.SchluesselwortAnlegen": func() error {
			_, f := worte.SchluesselwortAnlegen(eigen, fremd.set, "ende")
			return f
		},
		"Schluesselwortablage.SchluesselwortSetzen": func() error {
			return worte.SchluesselwortSetzen(eigen, fremd.set, fremd.wort, "wort", "neu")
		},
		"Schluesselwortablage.SchluesselwortLoeschen": func() error {
			return worte.SchluesselwortLoeschen(eigen, fremd.set, fremd.wort)
		},
	}
}

// **Die Werte einer fremden Karte fehlen** — wie die einer gelöschten.
//
// `Kartenwerte` nimmt eine Reihe von Kennungen und gab für eine, die es nicht
// gibt, schon immer nichts zurück statt eines Fehlers; eine fremde Karte
// verhält sich seit #1815 genauso.
func TestDieWerteEinerFremdenKarteFehlen(t *testing.T) {
	t.Parallel()

	db, fremd := besitzbestand(t)

	werte, fehler := NeuerSetLeser(db).Kartenwerte(MitProfil(t.Context(), 1), []int64{fremd.karte})
	if fehler != nil || len(werte) != 0 {
		t.Errorf("aus einem fremden Profil: %v (%v)", werte, fehler)
	}

	eigene, fehler := NeuerSetLeser(db).Kartenwerte(MitProfil(t.Context(), 2), []int64{fremd.karte})
	if fehler != nil || len(eigene[fremd.karte]) != 1 {
		t.Errorf("der Eigner liest %v (%v)", eigene, fehler)
	}

	if _, fehler := NeuerSetLeser(db).Kartenwerte(t.Context(), []int64{fremd.karte}); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("ohne Profil: %v", fehler)
	}
}

// **Die Gegenprobe: der Eigner erreicht seine Karte.** Ohne sie hieße ein
// „nicht gefunden" oben nur, dass der Ruf gar nicht ging.
func TestDerEignerErreichtSeineKarte(t *testing.T) {
	t.Parallel()

	db, fremd := besitzbestand(t)
	eigner := MitProfil(t.Context(), 2)

	if fehler := NeuerKartenschreiber(db).FeldSetzen(eigner, fremd.karte, "name", "Geändert"); fehler != nil {
		t.Fatalf("der Eigner setzt seinen Kartennamen nicht: %v", fehler)
	}

	zeilen, fehler := NeuerSetLeser(db).Kartenzeilen(eigner, fremd.set)
	if fehler != nil || len(zeilen) != 2 || zeilen[0].Name != "Geändert" {
		t.Errorf("der Eigner liest %+v (%v)", zeilen, fehler)
	}
}
