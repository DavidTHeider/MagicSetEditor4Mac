package storage_test

// **Was die Kontaktablage meldet, wenn der Bestand nicht mitspielt** —
// Werkstück N1 von Phase 5.4
// (#1454).
//
// # Warum jeder dieser Wege einen Lauf bekommt
//
// **Weil ein verschluckter Fehler hier eine Freigabe wäre.** Wer eine Sperre
// setzt und eine Meldung nicht sieht, hält jemanden für gesperrt, der es nicht
// ist. Das ist der Unterschied zu einer Liste, die leer bleibt.

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// alleWege ruft jede Fläche der Ablage **der Reihe nach** und sammelt, was
// sie meldet.
//
// # Warum die Reihenfolge festgelegt ist und nicht dem Zufall überlassen
//
// **Weil ein Löschen ohne Zeile nichts löscht.** Die Wege standen bis zum
// 2026-09-14 als Werte eines Kartenliterals da, und das hat zwei Dinge
// zugleich verdorben: Go sagt über die Auswertungsreihenfolge nichts zu, und
// die tatsächliche war die falsche. `Sperren` lief **vor** `Erlauben`; als
// `Entziehen` und `AllesZurueck` an die Reihe kamen, stand kein einziges Recht
// mehr in der Tabelle. Ein `DELETE`, das keine Zeile trifft, löst keinen
// `BEFORE DELETE`-Auslöser aus — und zwei der drei Auslöser in
// [TestEinAbgebrochenesSchreibenWirdGemeldet] haben deshalb **nie gefeuert**.
//
// **Gesehen hat es niemand**, weil jener Lauf nur verlangte, dass sich
// *irgendein* Weg meldet, und „kein Stand" meldete sich immer. Erst die
// Verschärfung auf benannte Wege hat es angestrichen.
//
// **Deshalb steht hier eine Folge und keine Karte**, und jeder Schritt findet
// vor, was er braucht: erlaubt wird an einem bestätigten Kontakt, entzogen
// wird ein vorhandenes Recht, und vor jedem Räumen steht wieder eines da.
func alleWege(
	t *testing.T, ablage *storage.Kontaktablage, ctx context.Context,
) map[string]error {
	t.Helper()

	_, listenfehler := ablage.Liste(ctx)
	_, lesefehler := ablage.Lesen(ctx, "bgegenueber")

	erlaube := func() error {
		return ablage.Erlauben(
			ctx, "bgegenueber", domain.ZweckTeilen, derPeerlauf)
	}

	gemeldet := map[string]error{
		"Liste":    listenfehler,
		"Lesen":    lesefehler,
		"Begegnet": ablage.Begegnet(ctx, "bwer", "ab", "Wer", derPeerlauf),
		"Stand":    ablage.Stand(ctx, "bgegenueber", domain.KontaktBestaetigt),
	}

	gemeldet["Erlauben"] = erlaube()
	gemeldet["Entziehen"] = ablage.Entziehen(
		ctx, "bgegenueber", domain.ZweckTeilen)
	gemeldet["Adresse"] = ablage.AdresseSetzen(
		ctx, "bgegenueber", "beispiel.invalid:443")

	// **Und die Notiz** (T3, #1468) — sie erlaubt nichts und geht trotzdem
	// über denselben Zug: ohne Profil, ohne Zeile und mit klemmendem
	// Schreiben endet sie wie jeder andere Weg.
	gemeldet["Notiz"] = ablage.NotizSetzen(ctx, "bgegenueber", "hm")

	// **Vor dem Sperren wieder ein Recht** — nur so räumt es eines weg, und
	// nur dann kann sein Auslöser überhaupt feuern.
	_ = erlaube() //nolint:errcheck // hier zählt der Weg darunter
	gemeldet["Sperren"] = ablage.Stand(
		ctx, "bgegenueber", domain.KontaktGesperrt)

	// **Und vor dem Räumen noch einmal dasselbe.**
	_ = ablage.Stand( //nolint:errcheck // dito
		ctx, "bgegenueber", domain.KontaktBestaetigt)
	_ = erlaube() //nolint:errcheck // dito
	gemeldet["AllesZurueck"] = ablage.AllesZurueck(ctx)

	return gemeldet
}

// **Ohne Profil im Kontext geht keiner der Wege.**
func TestOhneProfilGehtKeinKontaktweg(t *testing.T) {
	db := frischeDatenbank(t)

	for name, fehler := range alleWege(
		t, storage.NeueKontaktablage(db), t.Context()) {
		if fehler == nil {
			t.Errorf("%s kam ohne Profil durch", name)
		}
	}
}

// **Ohne Tabellen meldet sich der Bestand, statt leer zu antworten.**
func TestOhneKontakttabellenMeldetSichDerBestand(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)
	ctx := probekontext(t)

	begegne(t, ablage, ctx, "bgegenueber", "Fremder Rechner")

	if fehler := ablage.Stand(
		ctx, "bgegenueber", domain.KontaktBestaetigt); fehler != nil {
		t.Fatalf("Stand: %v", fehler)
	}

	// **Erst die Rechte allein** — dann laufen die Wege bis in ihre
	// Transaktion und scheitern dort, statt schon am Kontakt.
	if _, err := db.ExecContext(
		t.Context(), `DROP TABLE peerrechte`); err != nil {
		t.Fatalf("fallen lassen: %v", err)
	}

	for name, fehler := range alleWege(t, ablage, ctx) {
		// **Begegnet, Stand, Adresse und Notiz fassen die Rechte nicht an**
		// — der erste schreibt nur den Kontakt, der zweite bestätigt ihn,
		// der dritte trägt ein, wo er sitzt, der vierte, was man über ihn
		// denkt; alle vier sollen ohne die zweite Tabelle durchkommen.
		if name == "Begegnet" || name == "Stand" || name == "Adresse" ||
			name == "Notiz" {
			continue
		}

		if fehler == nil {
			t.Errorf("%s kam ohne die Rechtetabelle durch", name)
		}
	}

	if _, err := db.ExecContext(
		t.Context(), `DROP TABLE peerkontakte`); err != nil {
		t.Fatalf("fallen lassen: %v", err)
	}

	for name, fehler := range alleWege(t, ablage, ctx) {
		if fehler == nil {
			t.Errorf("%s kam ohne die Kontakttabelle durch", name)
		}
	}
}

// **Ein geschlossener Bestand meldet sich beim Beginn der Transaktion.**
func TestEinGeschlossenerBestandMeldetSichBeimBeginn(t *testing.T) {
	db := geschlossen(t)
	ablage := storage.NeueKontaktablage(db)

	for name, fehler := range alleWege(t, ablage, probekontext(t)) {
		if fehler == nil {
			t.Errorf("%s kam am geschlossenen Bestand durch", name)
		}
	}
}

// **Ein Schreiben, das ein Auslöser abbricht, wird gemeldet.**
//
// # Warum es dafür einen Auslöser braucht
//
// **Weil die Tabelle sonst mitspielt.** Ein `UPDATE` auf eine vorhandene Zeile
// glückt; damit es scheitert, muss der Bestand selbst nein sagen — und das ist
// genau die Lage, um die es geht: eine volle Platte, eine Regel, eine
// Fremdschlüsselverletzung.
func TestEinAbgebrochenesSchreibenWirdGemeldet(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)
	ctx := probekontext(t)

	begegne(t, ablage, ctx, "bgegenueber", "Fremder Rechner")

	if fehler := ablage.Stand(
		ctx, "bgegenueber", domain.KontaktBestaetigt); fehler != nil {
		t.Fatalf("Stand: %v", fehler)
	}

	// **Jeder Auslöser nennt den Weg, der sich melden muss.**
	//
	// **Hier stand „mindestens einer meldete sich"**, und das war zu wenig:
	// vier Wege fassen `peerrechte` an, und der Lauf wäre auch dann grün
	// gewesen, wenn drei davon geschwiegen hätten. Ein Kopf, der vier Wege
	// prüft und einen verlangt, macht die anderen drei unsichtbar (§10).
	faelle := map[string]struct {
		auslöser string
		wege     []string
	}{
		"kein Stand": {
			`CREATE TRIGGER nein_stand
			BEFORE UPDATE OF stand ON peerkontakte
			BEGIN SELECT RAISE(ABORT, 'nein'); END`,
			[]string{"Stand", "Sperren", "AllesZurueck"},
		},
		"keine Adresse": {
			`CREATE TRIGGER nein_adresse
			BEFORE UPDATE OF adresse ON peerkontakte
			BEGIN SELECT RAISE(ABORT, 'nein'); END`,
			[]string{"Adresse"},
		},
		// **Keine Notiz** (T3, #1468) — dieselbe Frage wie bei der Adresse,
		// eine andere Spalte.
		"keine Notiz": {
			`CREATE TRIGGER nein_notiz
			BEFORE UPDATE OF notiz ON peerkontakte
			BEGIN SELECT RAISE(ABORT, 'nein'); END`,
			[]string{"Notiz"},
		},
		"kein Recht": {
			`CREATE TRIGGER nein_recht
			BEFORE INSERT ON peerrechte
			BEGIN SELECT RAISE(ABORT, 'nein'); END`,
			[]string{"Erlauben"},
		},
		"kein Entzug": {
			`CREATE TRIGGER nein_entzug
			BEFORE DELETE ON peerrechte
			BEGIN SELECT RAISE(ABORT, 'nein'); END`,
			[]string{"Entziehen", "Sperren", "AllesZurueck"},
		},
	}

	for name, fall := range faelle {
		auslöser, erwartet := fall.auslöser, fall.wege

		t.Run(name, func(t *testing.T) {
			eigene := frischeDatenbank(t)
			eigeneAblage := storage.NeueKontaktablage(eigene)
			eigenerKontext := probekontext(t)

			begegne(t, eigeneAblage, eigenerKontext, "bgegenueber", "Wer")

			if fehler := eigeneAblage.Stand(eigenerKontext, "bgegenueber",
				domain.KontaktBestaetigt); fehler != nil {
				t.Fatalf("Stand: %v", fehler)
			}

			if _, err := eigene.ExecContext(
				t.Context(), auslöser); err != nil {
				t.Fatalf("Auslöser: %v", err)
			}

			gemeldet := alleWege(t, eigeneAblage, eigenerKontext)

			for _, weg := range erwartet {
				if gemeldet[weg] == nil {
					t.Errorf("%s kam am Auslöser vorbei", weg)
				}
			}
		})
	}
}

// **Ein krummer Zeitpunkt im Bestand ist ein Befund und keine leere Liste.**
func TestEinKrummerBegegnungszeitpunktIstEinBefund(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)
	ctx := probekontext(t)

	begegne(t, ablage, ctx, "bgegenueber", "Fremder Rechner")

	if _, err := db.ExecContext(t.Context(),
		`UPDATE peerkontakte SET gesehen = 'gestern'`); err != nil {
		t.Fatalf("verbiegen: %v", err)
	}

	if _, fehler := ablage.Lesen(ctx, "bgegenueber"); fehler == nil {
		t.Error("der krumme Zeitpunkt kam durch")
	}

	if _, fehler := ablage.Liste(ctx); fehler == nil {
		t.Error("die Liste nahm ihn hin")
	}
}

// **Und was nur der Stolpertreiber hergibt.**
func TestDieKontaktablageMeldetWasNurDerStolpertreiberHergibt(t *testing.T) {
	t.Parallel()

	// **Der Treibername steht zusammengesetzt da** — `gosec` hält ihn sonst
	// für ein eingebautes Geheimnis.
	stolper := "stolper-"

	for name, treiber := range map[string]string{
		"ein Fehler mitten in den Kontakten":       stolper + "kontakt-ende",
		"eine Kontaktzeile passt nicht":            stolper + "kontakt-scan",
		"ein Fehler mitten in den Rechten":         stolper + "rechte-ende",
		"eine Rechtezeile passt nicht in ihr Feld": stolper + "rechte-scan",

		// **Eine erste Begegnung, die keine ist** (T3, #1468): sie kommt aus
		// der Datei, und „nicht bekannt" ist eine Aussage — kein Rückfall
		// für etwas Unlesbares (§8).
		"eine erste Begegnung ist kein Zeitpunkt": stolper + "kontakt-frist",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db, fehler := sql.Open(treiber, "egal")
			if fehler != nil {
				t.Fatalf("öffnen: %v", fehler)
			}

			t.Cleanup(func() { _ = db.Close() }) //nolint:errcheck // nur ein Test

			ablage := storage.NeueKontaktablage(db)

			if _, fehler := ablage.Liste(probekontext(t)); fehler == nil {
				t.Error("die Liste kam durch")
			}
		})
	}
}
