package storage_test

// **Beide Ebenen in einem Zug** — Werkstück O3 von Phase 5.2
// (#1427).

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Was übernommen wird, steht danach in beiden Tafeln.**
func TestEineUebernahmeSchreibtBeideTafeln(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueEinstellungen(db)

	if fehler := ablage.EinstellungenUebernehmen(probekontext(t),
		map[string]string{domain.EinstellungSprache: "en"},
		map[string]string{domain.NutzerKontrast: domain.KontrastHoch},
		"",
	); fehler != nil {
		t.Fatalf("%v", fehler)
	}

	imProfil, fehler := ablage.Alle(probekontext(t))
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if imProfil[domain.EinstellungSprache] != "en" {
		t.Errorf("im Profil steht %v", imProfil)
	}

	beimMenschen, fehler := storage.NeueNutzerablage(db).
		NutzerwerteZumProfil(probekontext(t))
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if beimMenschen[domain.NutzerKontrast] != domain.KontrastHoch {
		t.Errorf("beim Menschen steht %v", beimMenschen)
	}
}

// **Zweimal übernommen heißt einmal gespeichert.**
func TestEineUebernahmeUeberschreibtStattZuVerdoppeln(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueEinstellungen(db)

	for _, wert := range []string{"de", "en"} {
		if fehler := ablage.EinstellungenUebernehmen(probekontext(t),
			map[string]string{domain.EinstellungSprache: wert},
			map[string]string{domain.NutzerBewegung: domain.NachSystem},
			"",
		); fehler != nil {
			t.Fatalf("%v", fehler)
		}
	}

	imProfil, fehler := ablage.Alle(probekontext(t))
	if fehler != nil {
		t.Fatalf("%v", fehler)
	}

	if imProfil[domain.EinstellungSprache] != "en" {
		t.Errorf("im Profil steht %v", imProfil)
	}
}

// **Ein fehlerhafter Wert lässt beide Tafeln unberührt.**
//
// **Die zweite Zeile ist die Probe**: der Profilwert steht **vor** dem
// Nutzerwert und wäre bei einer Prüfung erst beim Schreiben schon gesetzt.
func TestEinFehlerhafterWertHinterlaesstNichts(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueEinstellungen(db)

	fehler := ablage.EinstellungenUebernehmen(probekontext(t),
		map[string]string{domain.EinstellungSprache: "en"},
		map[string]string{domain.NutzerKontrast: "leuchtend"}, "")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("%v", fehler)
	}

	imProfil, lesefehler := ablage.Alle(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("%v", lesefehler)
	}

	if len(imProfil) != 0 {
		t.Errorf("im Profil steht %v", imProfil)
	}
}

// **Und ein unbekannter Profilschlüssel ebenso.**
func TestEinUnbekannterSchluesselWirdAbgewiesen(t *testing.T) {
	db := frischeDatenbank(t)

	fehler := storage.NeueEinstellungen(db).EinstellungenUebernehmen(
		probekontext(t), map[string]string{"gibtesnicht": "egal"}, nil, "")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("%v", fehler)
	}
}

// **Ohne Profil im Kontext ist es ein Fehler und keine leere Antwort** (§8).
func TestOhneProfilKeineUebernahme(t *testing.T) {
	db := frischeDatenbank(t)

	if fehler := storage.NeueEinstellungen(db).EinstellungenUebernehmen(
		context.Background(), nil, nil, ""); fehler == nil {
		t.Error("es kam durch")
	}
}

// **Ein Profil, das es nicht gibt, hat keinen Menschen** — und dann wird
// nichts geschrieben.
func TestOhneProfilInDerTafelKeinNutzerwert(t *testing.T) {
	db := frischeDatenbank(t)

	fehler := storage.NeueEinstellungen(db).EinstellungenUebernehmen(
		storage.MitProfil(t.Context(), 9999), nil,
		map[string]string{domain.NutzerKontrast: domain.KontrastHoch}, "")
	if fehler == nil {
		t.Error("es kam durch")
	}
}

// **Und der erste Satz meldet sich ebenso.**
//
// **Die Schleifen sind zwei**, und ein Lauf, der nur die zweite scheitern
// lässt, ließe den **ersten** Fehlerzweig unbetreten — dieselbe Lücke, die
// `TestZuruecksetzenMeldetDenErstenSatz` nebenan schließt.
//
// **Über eine entfernte Tafel und nicht über ein fremdes Profil**: `settings`
// trägt keinen Fremdschlüssel auf `profile`, und eine Zeile mit der Kennung
// 9999 kommt dort **durch** — nachgemessen am 2026-09-08, als genau dieser
// Lauf grün war, wo er rot sein sollte.
func TestEineFehlendeTafelMeldetSichBeimUebernehmen(t *testing.T) {
	db := frischeDatenbank(t)

	if _, err := db.ExecContext(probekontext(t), `DROP TABLE settings`); err != nil {
		t.Fatalf("die Tafel entfernen: %v", err)
	}

	fehler := storage.NeueEinstellungen(db).EinstellungenUebernehmen(
		probekontext(t), map[string]string{domain.EinstellungSprache: "en"}, nil, "")
	if fehler == nil {
		t.Error("es kam durch")
	}
}

// **Eine geschlossene Datenbank fängt gar nicht erst an.**
func TestOhneDatenbankKeineUebernahme(t *testing.T) {
	fehler := storage.NeueEinstellungen(kaputterBestand(t)).EinstellungenUebernehmen(
		probekontext(t), map[string]string{domain.EinstellungSprache: "en"}, nil, "")
	if fehler == nil {
		t.Error("es kam durch")
	}
}

// **Eine Revision, die nicht mehr gilt, schreibt nicht.**
//
// Befund 11 des Zweitblicks über Phase 5.2: `service` verglich die Revision
// **vor** der Transaktion, und zwischen Vergleich und Schreiben lag ein
// Fenster. Wer es traf, überschrieb eine fremde Änderung, ohne dass jemand
// etwas erfuhr.
func TestEineUeberholteRevisionSchreibtNicht(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueEinstellungen(db)

	if fehler := ablage.EinstellungenUebernehmen(probekontext(t),
		map[string]string{domain.EinstellungSprache: "en"}, nil,
		"eine-revision-von-vorhin",
	); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("abgewiesen wurde mit %v", fehler)
	}

	imProfil, lesefehler := ablage.Alle(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("%v", lesefehler)
	}

	if len(imProfil) != 0 {
		t.Errorf("geschrieben wurde trotzdem: %v", imProfil)
	}

	// **Die Gegenprobe**: mit der Revision des heutigen Standes geht es.
	nutzerID := nutzerkennungZumProfil(t, db)

	if fehler := ablage.EinstellungenUebernehmen(probekontext(t),
		map[string]string{domain.EinstellungSprache: "en"}, nil,
		domain.Standmarke(1, nutzerID, map[string]string{}, map[string]string{}),
	); fehler != nil {
		t.Fatalf("mit der heutigen Revision: %v", fehler)
	}
}

// nutzerkennungZumProfil liest, wem das Probeprofil gehört.
func nutzerkennungZumProfil(t *testing.T, db *sql.DB) int64 {
	t.Helper()

	var nutzerID int64

	if err := db.QueryRowContext(t.Context(),
		`SELECT nutzer_id FROM profile WHERE id = 1`).Scan(&nutzerID); err != nil {
		t.Fatalf("den Menschen zum Profil lesen: %v", err)
	}

	return nutzerID
}

// **Die Revision wird aus den Zeilen gerechnet, die dastehen.**
//
// Der Lauf darüber prüft sie an einem **leeren** Stand; hier stehen Werte auf
// beiden Ebenen, und damit läuft die Schleife, die sie einliest.
func TestDieRevisionRechnetUeberDieVorhandenenWerte(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueEinstellungen(db)

	imProfil := map[string]string{domain.EinstellungSprache: "en"}
	beimMenschen := map[string]string{domain.NutzerKontrast: domain.KontrastHoch}

	if fehler := ablage.EinstellungenUebernehmen(probekontext(t),
		imProfil, beimMenschen, "",
	); fehler != nil {
		t.Fatalf("der erste Zug: %v", fehler)
	}

	nutzerID := nutzerkennungZumProfil(t, db)
	heute := domain.Standmarke(1, nutzerID, imProfil, beimMenschen)

	// **Mit der heutigen Revision geht es.**
	if fehler := ablage.EinstellungenUebernehmen(probekontext(t),
		map[string]string{domain.EinstellungSprache: "de"}, nil, heute,
	); fehler != nil {
		t.Fatalf("mit der heutigen Revision: %v", fehler)
	}

	// **Und dieselbe Revision gilt danach nicht mehr.**
	if fehler := ablage.EinstellungenUebernehmen(probekontext(t),
		map[string]string{domain.EinstellungSprache: "en"}, nil, heute,
	); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("abgewiesen wurde mit %v", fehler)
	}
}

// **Und jede Abfrage der Revision meldet ihren Fehler.**
//
// **Über entfernte Tafeln**, eine nach der anderen: ein Lauf, der nur die
// erste scheitern lässt, ließe die beiden dahinter unbetreten (§3).
func TestDieRevisionMeldetJedeIhrerAbfragen(t *testing.T) {
	for name, tafel := range map[string]string{
		"das Profil":              "profile",
		"die Einstellungen":       "settings",
		"die Nutzereinstellungen": "nutzereinstellungen",
	} {
		t.Run(name, func(t *testing.T) {
			db := frischeDatenbank(t)

			if _, fehler := db.ExecContext(probekontext(t),
				"DROP TABLE "+tafel); fehler != nil {
				t.Fatalf("die Tafel entfernen: %v", fehler)
			}

			fehler := storage.NeueEinstellungen(db).EinstellungenUebernehmen(
				probekontext(t), map[string]string{domain.EinstellungSprache: "en"},
				nil, "eine-revision")

			if fehler == nil {
				t.Error("es kam durch")
			}
		})
	}
}
