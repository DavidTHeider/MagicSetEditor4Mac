package storage

// **Ein Zusammenbruch gibt die Transaktion zurück** — BK-04
// (#2129).
//
// # Der Befund
//
// Sechs Schreibwege rollten nur zurück, wenn ihr **benannter Fehler** gesetzt
// war (`if fehler != nil { … verwerfen(tx) }`). Ein `panic` setzt ihn nicht:
// die Transaktion blieb offen, hielt die **einzige** Verbindung des Pools
// (`SetMaxOpenConns(1)`) und mit `_txlock=immediate` die Schreibsperre der
// Datei. Der nächste Ruf wartete bis zu seiner Frist, und jeder andere
// Prozess bekam `busy` — bis das Programm endete.
//
// # Der Lauf
//
// Je Weg eine **echte** Datei über den Paniktreiber; zusammengebrochen wird
// vor einer Anweisung **in** der Transaktion, mit einem Elternkontext, der
// nicht endet. Danach gilt: der Pool ist frei, eine zweite Verbindung nimmt
// die Schreibsperre sofort, nichts ist halb geschrieben, und derselbe Weg
// geht wieder.

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

type panikfall struct {
	name string

	// muster ist ein Stück der Anweisung, vor der zusammengebrochen wird.
	muster string

	schreibe func(ctx context.Context, db *sql.DB) error

	// steht liest über die zweite Verbindung, was der Weg ändern wollte —
	// es muss noch der alte Stand sein.
	steht func(ctx context.Context, db *sql.DB) (string, error)
	alt   string
}

func panikfaelle() []panikfall {
	name := []domain.Value{{Field: "name", Text: "Eisball"}}
	kartenname := func(ctx context.Context, db *sql.DB) (string, error) {
		var wert string
		err := db.QueryRowContext(ctx, `SELECT wert FROM card_values WHERE card_id = 100 AND feld = 'name'`).Scan(&wert)

		return wert, err
	}

	return []panikfall{
		{
			name: "FelderSetzenMitVorwert", muster: "INSERT INTO history", alt: "Feuerball", steht: kartenname,
			schreibe: func(ctx context.Context, db *sql.DB) error {
				_, _, f := NeuerKartenschreiber(db).FelderSetzenMitVorwert(ctx, 100, name)
				return f
			},
		},
		{
			name: "FeldSetzenMitVorwert", muster: "INSERT INTO history", alt: "Feuerball", steht: kartenname,
			schreibe: func(ctx context.Context, db *sql.DB) error {
				_, _, f := NeuerKartenschreiber(db).FeldSetzenMitVorwert(ctx, 100, "name", "Eisball")
				return f
			},
		},
		{
			name: "SetwertSetzenMitVorwert", muster: "INSERT INTO history", alt: "Saat",
			schreibe: func(ctx context.Context, db *sql.DB) error {
				_, _, f := NeuerKartenschreiber(db).SetwertSetzenMitVorwert(ctx, 10, "title", "Ernte")
				return f
			},
			steht: func(ctx context.Context, db *sql.DB) (string, error) {
				var wert string
				err := db.QueryRowContext(ctx, `SELECT wert FROM set_values WHERE set_id = 10 AND feld = 'title'`).Scan(&wert)

				return wert, err
			},
		},
		{
			name: "GruppeSetzenMitHerkunft", muster: "INSERT INTO history", alt: "Feuerball", steht: kartenname,
			schreibe: func(ctx context.Context, db *sql.DB) error {
				return NeuerKartenschreiber(db).GruppeSetzenMitHerkunft(ctx, []domain.Kartenschritt{{
					KarteID: 100, Werte: name, Herkunft: "ki",
					Herkuenfte: []domain.KIFeldherkunft{{
						Feld: "name", Anbieter: "ollama", Modell: "gemma4:8b-q4",
						Aufgabe: "Schreibe den Namen.", Auftrag: "ki-2129",
						Zeitpunkt: time.Date(2026, 10, 7, 1, 0, 0, 0, time.UTC),
						Bezug:     domain.KIInhaltsbezug([]string{"Feuerball"}),
					}},
				}})
			},
		},
		{
			name: "EinstellungenUebernehmen", muster: "INSERT INTO nutzereinstellungen", alt: "",
			schreibe: func(ctx context.Context, db *sql.DB) error {
				return NeueEinstellungen(db).EinstellungenUebernehmen(ctx,
					map[string]string{domain.EinstellungSprache: "en"},
					map[string]string{domain.NutzerKontrast: domain.KontrastHoch}, "")
			},
			steht: func(ctx context.Context, db *sql.DB) (string, error) {
				var wert string
				err := db.QueryRowContext(ctx,
					`SELECT coalesce((SELECT wert FROM settings WHERE profil_id = 1 AND schluessel = ?), '')`,
					domain.EinstellungSprache).Scan(&wert)

				return wert, err
			},
		},
		{
			name: "Kontaktablage.NotizSetzen", muster: "UPDATE peerkontakte", alt: "",
			schreibe: func(ctx context.Context, db *sql.DB) error {
				return NeueKontaktablage(db).NotizSetzen(ctx, "peer-1", "tauscht fair")
			},
			steht: func(ctx context.Context, db *sql.DB) (string, error) {
				var wert string
				err := db.QueryRowContext(ctx, `SELECT coalesce(notiz, '') FROM peerkontakte`).Scan(&wert)

				return wert, err
			},
		},
	}
}

func TestEinZusammenbruchGibtDieTransaktionZurueck(t *testing.T) {
	for _, fall := range panikfaelle() {
		t.Run(fall.name, func(t *testing.T) {
			pfad := filepath.Join(t.TempDir(), "bestand.sqlite")
			zweite := vorbereiteterBestand(t, pfad)

			ausloeser := &panikausloeser{muster: fall.muster}
			db := panikbestand(t, pfad, ausloeser)
			ctx := MitProfil(t.Context(), 1)

			ausloeser.scharf.Store(true)

			var fehler error
			if gefangen := bricht(func() { fehler = fall.schreibe(ctx, db) }); gefangen == nil {
				t.Fatalf("kein Zusammenbruch — »%s« traf keine Anweisung (%v)", fall.muster, fehler)
			}

			if offen := db.Stats().InUse; offen != 0 {
				t.Errorf("nach dem Zusammenbruch sind %d Verbindungen belegt", offen)
			}

			// **Die Schreibsperre ist frei** — eine zweite Verbindung nimmt
			// sie, ohne zu warten.
			kurz, ende := context.WithTimeout(t.Context(), 500*time.Millisecond)
			defer ende()

			if _, err := zweite.ExecContext(kurz, `BEGIN IMMEDIATE; COMMIT`); err != nil {
				t.Errorf("eine zweite Verbindung bekommt die Schreibsperre nicht: %v", err)
			}

			if wert, err := fall.steht(t.Context(), zweite); err != nil || wert != fall.alt {
				t.Errorf("halb geschrieben: %q (%v), erwartet %q", wert, err, fall.alt)
			}

			// **Und derselbe Weg geht danach wieder.**
			wieder, aus := context.WithTimeout(MitProfil(t.Context(), 1), 2*time.Second)
			defer aus()

			if err := fall.schreibe(wieder, db); err != nil {
				t.Errorf("derselbe Weg danach: %v", err)
			}
		})
	}
}

// bricht ruft `tun` und gibt zurück, was es geworfen hat.
func bricht(tun func()) (gefangen any) {
	defer func() { gefangen = recover() }()

	tun()

	return nil
}

// vorbereiteterBestand legt an `pfad` einen migrierten Bestand mit Set 10,
// Karte 100 und einem Kontakt an und gibt eine **zweite** Verbindung zurück.
func vorbereiteterBestand(t *testing.T, pfad string) *sql.DB {
	t.Helper()

	db, err := Oeffnen(t.Context(), "file:"+pfad)
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	for _, anweisung := range []string{
		`INSERT INTO sets (id, name, spiel, geoeffnet, profil_id) VALUES (10, 'Saat', 'magic', '2026-10-07', 1)`,
		`INSERT INTO cards (id, set_id, reihenfolge) VALUES (100, 10, 0)`,
		`INSERT INTO card_values (card_id, feld, wert, reihenfolge) VALUES (100, 'name', 'Feuerball', 0)`,
		`INSERT INTO set_values (set_id, feld, wert, reihenfolge) VALUES (10, 'title', 'Saat', 0)`,
	} {
		if _, err := db.ExecContext(t.Context(), anweisung); err != nil {
			t.Fatalf("%s: %v", anweisung, err)
		}
	}

	if err := NeueKontaktablage(db).Begegnet(probekontext(t), "peer-1", "ab:cd", "Anna",
		time.Date(2026, 10, 7, 1, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("begegnet: %v", err)
	}

	return db
}

// **Dieselbe Klasse an zwei weiteren Stellen**: `imRetrievalzug` rollte nur
// im Fehlerzweig zurück, `SprachpaketAendern` nur bei gesetztem Fehler — und
// beide rufen fremde Arbeit in ihrer Transaktion. Bricht die zusammen, gilt
// dasselbe wie oben.
func TestEinZusammenbruchInFremderArbeitGibtDieTransaktionZurueck(t *testing.T) {
	faelle := map[string]func(ctx context.Context, db *sql.DB) error{
		"imRetrievalzug": func(ctx context.Context, db *sql.DB) error {
			return imRetrievalzug(ctx, db, func(tx *sql.Tx) error {
				if _, err := tx.ExecContext(ctx, `UPDATE sets SET name = 'Halb' WHERE id = 10`); err != nil {
					return err
				}

				panic("planmäßiger Zusammenbruch in der Arbeit")
			})
		},
		"SprachpaketAendern": func(ctx context.Context, db *sql.DB) error {
			ablage := NeueSprachpaketablage(db)
			if err := ablage.SprachpaketAnlegen(ctx, domain.Sprachpaket{
				Kennung: "da", Name: "Dansk", Basis: "en", Stand: domain.StandEntwurf,
			}); err != nil {
				return err
			}

			return ablage.SprachpaketAendern(ctx, "da", func(*domain.Sprachpaket) error {
				panic("planmäßiger Zusammenbruch in der Änderung")
			})
		},
	}

	for name, schreibe := range faelle {
		t.Run(name, func(t *testing.T) {
			pfad := filepath.Join(t.TempDir(), "bestand.sqlite")
			zweite := vorbereiteterBestand(t, pfad)
			db := panikbestand(t, pfad, &panikausloeser{})

			var fehler error
			if gefangen := bricht(func() { fehler = schreibe(MitProfil(t.Context(), 1), db) }); gefangen == nil {
				t.Fatalf("kein Zusammenbruch (%v)", fehler)
			}

			if offen := db.Stats().InUse; offen != 0 {
				t.Errorf("nach dem Zusammenbruch sind %d Verbindungen belegt", offen)
			}

			kurz, ende := context.WithTimeout(t.Context(), 500*time.Millisecond)
			defer ende()

			if _, err := zweite.ExecContext(kurz, `BEGIN IMMEDIATE; COMMIT`); err != nil {
				t.Errorf("eine zweite Verbindung bekommt die Schreibsperre nicht: %v", err)
			}
		})
	}
}
