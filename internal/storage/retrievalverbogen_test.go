package storage_test

// **Ein verbogener Index wird gemeldet und nicht gedeutet** — Werkstück A15
// von Phase 5.3
// (#1514).
//
// # Warum diese Läufe an der Datei ansetzen
//
// **Weil der Bestand eine Datei ist, die jemand weitergibt** (§8). Der Index
// darin ist **abgeleitet**; was er trägt, hat nicht immer dieses Programm
// geschrieben, und eine fehlende Tafel oder ein `NULL` in einer Textspalte
// sind keine erfundenen Fälle, sondern das, was eine Rücksicherung oder eine
// Migration hinterlassen kann.
//
// **Jeder dieser Wege endet mit einer Meldung** — keiner mit einem halben
// Stand und keiner mit einer erfundenen Auskunft.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Weil der Bestand eine Datei ist, die jemand weitergibt** (§8). Der Index
// darin ist abgeleitet; was er trägt, hat nicht immer dieses Programm
// geschrieben.
func TestEinVerbogenerIndexWirdGemeldet(t *testing.T) {
	faelle := map[string]struct {
		verbiegen string
		suchen    bool
	}{
		"die Tafel der Generationen fehlt": {
			`DROP TABLE retrieval_generationen`, true,
		},
		"das Manifest ist kein JSON": {
			`UPDATE retrieval_generationen SET manifest = 'kein JSON'`, false,
		},
		"die Abrufzeit ist unlesbar": {
			`UPDATE retrieval_dokumente SET abgerufen = 'gestern'`, true,
		},
		"die Volltexttafel fehlt": {
			`DROP TABLE retrieval_volltext`, true,
		},
		"die Vektortafel fehlt": {
			`DROP TABLE retrieval_vektoren`, true,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			db := frischeDatenbank(t)
			index := storage.NeuerRetrievalindex(db)

			generation, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
			if fehler != nil {
				t.Fatalf("Anlegen: %v", fehler)
			}

			if geschrieben := index.Schreiben(probekontext(t), generation,
				einIndexdokument("eines", "ein Text"),
				[]float32{1, 0, 0}); geschrieben != nil {
				t.Fatalf("Schreiben: %v", geschrieben)
			}

			if aktiviert := index.Aktivieren(probekontext(t),
				generation); aktiviert != nil {
				t.Fatalf("Aktivieren: %v", aktiviert)
			}

			if _, fehler := db.ExecContext(probekontext(t),
				fall.verbiegen); fehler != nil {
				t.Fatalf("verbiegen: %v", fehler)
			}

			if !fall.suchen {
				if _, fehler := index.Manifest(probekontext(t)); fehler == nil {
					t.Fatal("der verbogene Bestand gab ein Manifest")
				}

				return
			}

			lexikalisch, lexfehler := index.Lexikalisch(probekontext(t), "text", 10)
			semantisch, semfehler := index.Semantisch(probekontext(t),
				[]float32{1, 0, 0}, 10)

			if lexfehler == nil && semfehler == nil {
				t.Fatalf("der verbogene Bestand gab Treffer: %+v und %+v",
					lexikalisch, semantisch)
			}
		})
	}
}

// **Und ein Schreiben in eine abgeräumte Tafel auch.**
func TestEinSchreibenInEineAbgeraeumteTafelWirdGemeldet(t *testing.T) {
	db := frischeDatenbank(t)
	index := storage.NeuerRetrievalindex(db)

	generation, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	for name, verbiegen := range map[string]string{
		"ohne Dokumenttafel": `DROP TABLE retrieval_dokumente`,
		"ohne Volltexttafel": `DROP TABLE retrieval_volltext`,
		"ohne Vektortafel":   `DROP TABLE retrieval_vektoren`,
	} {
		t.Run(name, func(t *testing.T) {
			eigene := frischeDatenbank(t)
			eigenerIndex := storage.NeuerRetrievalindex(eigene)

			eigeneGeneration, fehler := eigenerIndex.Anlegen(probekontext(t),
				einIndexmanifest())
			if fehler != nil {
				t.Fatalf("Anlegen: %v", fehler)
			}

			if _, fehler := eigene.ExecContext(probekontext(t),
				verbiegen); fehler != nil {
				t.Fatalf("verbiegen: %v", fehler)
			}

			if geschrieben := eigenerIndex.Schreiben(probekontext(t),
				eigeneGeneration, einIndexdokument("eines", "ein Text"),
				[]float32{1, 0, 0}); geschrieben == nil {
				t.Fatal("in die abgeräumte Tafel wurde geschrieben")
			}
		})
	}

	// **Und die Sperre meldet sich ebenso.**
	if _, fehler := db.ExecContext(probekontext(t),
		`DROP TABLE retrieval_gesperrt`); fehler != nil {
		t.Fatalf("verbiegen: %v", fehler)
	}

	if fehler := index.Sperren(probekontext(t), "eines", 0); fehler == nil {
		t.Error("die Sperre kam ohne Tafel durch")
	}

	if fehler := index.Verwerfen(probekontext(t), generation); fehler != nil {
		t.Errorf("Verwerfen: %v", fehler)
	}
}

// **`STRICT` verhindert es beim Schreiben** — und das ist gut so. Der Fall
// entsteht nicht durch dieses Programm, sondern durch einen Bestand, den
// jemand von Hand bearbeitet hat (§8).
func TestEineUnlesbareIndexzeileIstEinFehler(t *testing.T) {
	db := frischeDatenbank(t)
	index := storage.NeuerRetrievalindex(db)

	generation, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if aktiviert := index.Aktivieren(probekontext(t), generation); aktiviert != nil {
		t.Fatalf("Aktivieren: %v", aktiviert)
	}

	// **Die Tafel wird ohne `STRICT` neu gebaut**, damit ein `NULL` in eine
	// Textspalte passt — dieselbe Gestalt wie in `kifeldherkunft_test.go`.
	//
	// **Eine Anweisung je Ruf**: der Treiber nimmt keine Folge entgegen.
	for _, anweisung := range []string{
		`DROP TABLE retrieval_dokumente`,
		`CREATE TABLE retrieval_dokumente (
		  id INTEGER PRIMARY KEY, generation INTEGER, kennung TEXT,
		  quelle TEXT, bezugsadresse TEXT, spiel TEXT, kennungssystem TEXT,
		  fachkennung TEXT, semantik TEXT, regelnummer TEXT, urheber TEXT,
		  sprache TEXT, titel TEXT, text TEXT, quellrevision TEXT,
		  quelldatum TEXT, abgerufen TEXT, inhaltshash TEXT, merkmale TEXT,
		  abschnitt INTEGER, von INTEGER, bis INTEGER,
		  profil_id INTEGER, sichtbarkeit TEXT)`,
		`INSERT INTO retrieval_dokumente VALUES
		  (1, 1, 'eines', 'oracle', '', 'magic', '', '', '', '', '',
		   'en', 'Titel', 'Lightning', '2026-09-01', '2026-09-01',
		   NULL, 'sha256:x', '{}', 0, 0, 9, 0, 'oeffentlich')`,
		`INSERT INTO retrieval_volltext (titel, text, dokument_id)
		  VALUES ('Titel', 'Lightning', 1)`,
		`INSERT INTO retrieval_vektoren (dokument_id, vektor) VALUES (1, x'00000000')`,
	} {
		if _, fehler := db.ExecContext(probekontext(t), anweisung); fehler != nil {
			t.Fatalf("verbiegen: %v", fehler)
		}
	}

	if _, fehler := index.Lexikalisch(probekontext(t), "lightning", 10); fehler == nil {
		t.Error("die unlesbare Zeile kam lexikalisch durch")
	}

	if _, fehler := index.Semantisch(probekontext(t),
		[]float32{1, 0, 0}, 10); fehler == nil {
		t.Error("die unlesbare Zeile kam semantisch durch")
	}
}

// **Weil er die Naht zwischen beiden Prüfungen misst.** Die `WHERE`-Zeile
// lässt diese Zeile durch — sie fragt „öffentlich **oder** mein Profil", und
// das Profil stimmt. Erst die zweite Prüfung wirft sie weg. Ohne diesen Lauf
// wäre nicht belegt, dass die zwei Prüfungen **verschieden** sind (§10).
func TestEineUnbekannteSichtbarkeitKommtNichtHeraus(t *testing.T) {
	db := frischeDatenbank(t)
	index := storage.NeuerRetrievalindex(db)

	generation, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if geschrieben := index.Schreiben(probekontext(t), generation,
		einEigenesDokument("meins", "Lightning in meinem Set", 1),
		[]float32{1, 0, 0}); geschrieben != nil {
		t.Fatalf("Schreiben: %v", geschrieben)
	}

	if aktiviert := index.Aktivieren(probekontext(t), generation); aktiviert != nil {
		t.Fatalf("Aktivieren: %v", aktiviert)
	}

	// **Der `CHECK` verhindert es beim Schreiben** — deshalb bekommt die
	// Spalte für diesen Lauf keinen, so wie ein Bestand ihn nicht hätte, den
	// jemand von Hand gebaut hat (§8). Dieselbe Gestalt wie oben.
	for _, anweisung := range []string{
		`ALTER TABLE retrieval_dokumente RENAME TO retrieval_dokumente_alt`,
		`CREATE TABLE retrieval_dokumente (
		  id INTEGER PRIMARY KEY, generation INTEGER, kennung TEXT,
		  quelle TEXT, bezugsadresse TEXT, spiel TEXT, kennungssystem TEXT,
		  fachkennung TEXT, semantik TEXT, regelnummer TEXT, urheber TEXT,
		  sprache TEXT, titel TEXT, text TEXT, quellrevision TEXT,
		  quelldatum TEXT, abgerufen TEXT, inhaltshash TEXT, merkmale TEXT,
		  abschnitt INTEGER, von INTEGER, bis INTEGER,
		  profil_id INTEGER, sichtbarkeit TEXT)`,
		`INSERT INTO retrieval_dokumente SELECT * FROM retrieval_dokumente_alt`,
		`UPDATE retrieval_dokumente SET sichtbarkeit = 'geheim'`,
	} {
		if _, verbogen := db.ExecContext(probekontext(t), anweisung); verbogen != nil {
			t.Fatalf("verbiegen: %v", verbogen)
		}
	}

	lexikalisch, fehler := index.Lexikalisch(probekontext(t), "lightning", 10)
	if fehler != nil {
		t.Fatalf("Lexikalisch: %v", fehler)
	}

	if len(lexikalisch) != 0 {
		t.Fatalf("die geheime Zeile kam lexikalisch heraus: %+v", lexikalisch)
	}

	semantisch, fehler := index.Semantisch(probekontext(t), []float32{1, 0, 0}, 10)
	if fehler != nil {
		t.Fatalf("Semantisch: %v", fehler)
	}

	if len(semantisch) != 0 {
		t.Fatalf("die geheime Zeile kam semantisch heraus: %+v", semantisch)
	}
}

// **Unlesbare Merkmale sind ein Fehler und kein leeres Dokument.**
//
// **Der Unterschied zählt**: ein Dokument, dessen Merkmale verloren gehen,
// fällt durch jeden Filter — still, und mit dem Anschein, es habe eben keine.
// Das Issue nennt es andersherum: „Unbekannter Wert ist kein erfundener
// Treffer"; ein **verlorener** Wert ist genauso wenig einer.
func TestUnlesbareMerkmaleSindEinFehler(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	index := storage.NeuerRetrievalindex(db)

	generation, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if geschrieben := index.Schreiben(probekontext(t), generation,
		einIndexdokument("eines", "Lightning Bolt"),
		[]float32{1, 0, 0}); geschrieben != nil {
		t.Fatalf("Schreiben: %v", geschrieben)
	}

	if aktiviert := index.Aktivieren(probekontext(t), generation); aktiviert != nil {
		t.Fatalf("Aktivieren: %v", aktiviert)
	}

	if _, verbogen := db.ExecContext(probekontext(t),
		`UPDATE retrieval_dokumente SET merkmale = 'kein JSON'`); verbogen != nil {
		t.Fatalf("verbiegen: %v", verbogen)
	}

	if _, fehler := index.Lexikalisch(probekontext(t), "lightning", 10); fehler == nil {
		t.Error("die unlesbaren Merkmale kamen lexikalisch durch")
	}

	if _, fehler := index.Semantisch(probekontext(t),
		[]float32{1, 0, 0}, 10); fehler == nil {
		t.Error("die unlesbaren Merkmale kamen semantisch durch")
	}
}

// **Ein verbogener Deltaimport bricht ab, statt halb zu kopieren.**
//
// **Drei Fehlerwege, die es ohne einen verbogenen Bestand nicht gibt**: eine
// Tafel, die es nicht mehr gibt, eine Zeile, die sich nicht lesen lässt, und
// ein Ziel, das die Zeile nicht annimmt. Alle drei enden vor dem Umschalten —
// eine halb kopierte Generation wäre genau der gemischte Stand, den das Issue
// ausschließt.
func TestEinVerbogenerDeltaimportBrichtAb(t *testing.T) {
	t.Parallel()

	for name, verbiegen := range map[string][]string{
		"die Dokumententafel ist fort": {
			`ALTER TABLE retrieval_dokumente RENAME TO fort`,
		},
		"eine Kennung ist keine": {
			`UPDATE retrieval_dokumente SET kennung = NULL`,
		},
		"die Volltexttafel ist fort": {
			`DROP TABLE retrieval_volltext`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db := frischeDatenbank(t)
			index := storage.NeuerRetrievalindex(db)

			erste, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
			if fehler != nil {
				t.Fatalf("Anlegen: %v", fehler)
			}

			if geschrieben := index.Schreiben(probekontext(t), erste,
				einIndexdokument("eines", "Lightning Bolt"),
				[]float32{1, 0, 0}); geschrieben != nil {
				t.Fatalf("Schreiben: %v", geschrieben)
			}

			if aktiviert := index.Aktivieren(probekontext(t), erste); aktiviert != nil {
				t.Fatalf("Aktivieren: %v", aktiviert)
			}

			zweite, fehler := index.Anlegen(probekontext(t), einIndexmanifest())
			if fehler != nil {
				t.Fatalf("Anlegen: %v", fehler)
			}

			// **Die Tafel ohne `STRICT` neu bauen, wo ein `NULL` hineinsoll**
			// — dieselbe Gestalt wie in den Läufen darüber.
			if name == "eine Kennung ist keine" {
				verbiegen = append([]string{
					`ALTER TABLE retrieval_dokumente RENAME TO alt`,
					`CREATE TABLE retrieval_dokumente (
					  id INTEGER PRIMARY KEY, generation INTEGER, kennung TEXT,
					  quelle TEXT, bezugsadresse TEXT, spiel TEXT,
					  kennungssystem TEXT, fachkennung TEXT, semantik TEXT,
					  regelnummer TEXT, urheber TEXT, sprache TEXT, titel TEXT,
					  text TEXT, quellrevision TEXT, quelldatum TEXT,
					  abgerufen TEXT, inhaltshash TEXT, merkmale TEXT,
					  abschnitt INTEGER, von INTEGER, bis INTEGER,
					  profil_id INTEGER, sichtbarkeit TEXT)`,
					`INSERT INTO retrieval_dokumente SELECT * FROM alt`,
				}, verbiegen...)
			}

			for _, anweisung := range verbiegen {
				if _, verbogen := db.ExecContext(probekontext(t),
					anweisung); verbogen != nil {
					t.Fatalf("verbiegen: %v", verbogen)
				}
			}

			if _, uebernommen := index.Uebernehmen(probekontext(t), zweite,
				nil); uebernommen == nil {
				t.Error("die verbogene Übernahme kam durch")
			}
		})
	}
}
