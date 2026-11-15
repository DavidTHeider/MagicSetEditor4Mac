package storage_test

// **Die Aufbewahrungsfrist für ein Protokoll** — Werkstück G5 von Phase 5.5
// (#1462).
//
// **Der zweite Halbsatz von „Aufbewahrungs-/Löschweg".** Der erste ist
// [storage.Partieablage.Vergessen] — jemand sagt, dass dieses Protokoll weg
// soll. Hier **bietet** es die Datenpflege an, und die Entscheidung bleibt
// beim Menschen: die Klasse ist `entscheidung` und nicht `sicher`.

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// eineAltePartie legt eine beendete Partie mit dem genannten Datum.
func eineAltePartie(
	t *testing.T, db *sql.DB, kennung string, geaendert time.Time,
) {
	t.Helper()

	eineAltePartieImProfil(t, db, kennung, 1, geaendert)
}

// eineAltePartieImProfil legt sie in einem bestimmten Profil.
func eineAltePartieImProfil(
	t *testing.T, db *sql.DB, kennung string, profilID int64,
	geaendert time.Time,
) {
	t.Helper()

	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO partien (id, profil_id, spiel, stand, zustand, geaendert)
		 VALUES (?, ?, 'magic', ?, '{}', ?)`,
		kennung, profilID, string(domain.StandBeendet),
		storage.Zeitstempel(geaendert)); fehler != nil {
		t.Fatalf("die Partie legen: %v", fehler)
	}
}

// **Eine Partie innerhalb der Frist wird nicht angeboten** — die Gegenprobe
// zur Zeitgrenze.
func TestEineJungePartieStehtNichtZurPflege(t *testing.T) {
	db := frischeDatenbank(t)

	eineAltePartie(t, db, "jung",
		pflegezeit.AddDate(0, 0, -(domain.AufbewahrungPartien-1)))

	if _, da := befundeNach(t, db)["alteProtokolle"]; da {
		t.Fatal("eine Partie innerhalb der Frist wurde angeboten")
	}
}

// **Jenseits der Frist steht sie da — und zwar zur Entscheidung.**
func TestEinAltesProtokollStehtZurEntscheidung(t *testing.T) {
	db := frischeDatenbank(t)

	eineAltePartie(t, db, "alt",
		pflegezeit.AddDate(0, 0, -(domain.AufbewahrungPartien+1)))

	befund, da := befundeNach(t, db)["alteProtokolle"]
	if !da {
		t.Fatal("das alte Protokoll wurde nicht gefunden")
	}

	if befund.Klasse != domain.PflegeEntscheidung {
		t.Fatalf("die Klasse ist %q", befund.Klasse)
	}

	zeilen, fehler := storage.PflegeAusfuehren(probekontext(t), db, pflegezeit,
		map[string]string{"alteProtokolle": befund.Revision})
	if fehler != nil {
		t.Fatalf("ausführen: %v", fehler)
	}

	if len(zeilen) != 1 || zeilen[0].Ausgang != domain.PflegeErledigt {
		t.Fatalf("die Bilanz ist %+v", zeilen)
	}
}

// **Eine laufende Partie wird nie angeboten**, gleich wie alt sie ist: sie ist
// kein Protokoll, sondern ein Tisch.
func TestEineLaufendePartieStehtNichtZurPflege(t *testing.T) {
	db := frischeDatenbank(t)

	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO partien (id, profil_id, spiel, stand, zustand, geaendert)
		 VALUES ('laeuft', 1, 'magic', ?, '{}', ?)`,
		string(domain.StandLaeuft), storage.Zeitstempel(
			pflegezeit.AddDate(0, 0, -10*domain.AufbewahrungPartien))); fehler != nil {
		t.Fatalf("die Partie legen: %v", fehler)
	}

	if _, da := befundeNach(t, db)["alteProtokolle"]; da {
		t.Fatal("eine laufende Partie wurde angeboten")
	}
}

// **Ein fremdes Protokoll wird nicht angeboten.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil `internal/werkzeuge/profilklausel_test.go` diese Datei ausnimmt.**
// Die Ausnahme gilt den **verwaisten** Zeilen — sie gehören zu keinem Profil
// mehr, und eine Klausel schlösse gerade sie aus. Eine beendete Partie hat
// ihr Profil aber; ohne Klausel böte die Pflege das Protokoll eines anderen
// zum Löschen an, und der Wächter dort sähe es nicht.
//
// **Deshalb steht die Zusage hier, wo sie eingelöst wird** (§10).
func TestEinFremdesProtokollStehtNichtZurPflege(t *testing.T) {
	db := frischeDatenbank(t)

	fremdes := zweitesProfil(t, storage.NeueNutzerablage(db), "Turnier")

	eineAltePartieImProfil(t, db, "fremd", fremdes,
		pflegezeit.AddDate(0, 0, -(domain.AufbewahrungPartien+1)))

	if _, da := befundeNach(t, db)["alteProtokolle"]; da {
		t.Fatal("das Protokoll eines fremden Profils wurde angeboten")
	}

	// **Und die Gegenprobe**: dasselbe Alter im eigenen Profil steht da.
	eineAltePartie(t, db, "eigen",
		pflegezeit.AddDate(0, 0, -(domain.AufbewahrungPartien+1)))

	befund, da := befundeNach(t, db)["alteProtokolle"]
	if !da {
		t.Fatal("das eigene Protokoll wurde nicht gefunden")
	}

	if len(befund.Beispiele) != 1 || befund.Beispiele[0] != "partien:eigen" {
		t.Fatalf("die Beispiele sind %v", befund.Beispiele)
	}
}

// **Ohne Profil im Kontext fährt eine profilgebundene Stelle nicht** —
// Werkstück G5 von Phase 5.5.
//
// **Und das ist die Zusage**: wer keine Sicht auf ein Profil hat, bekommt
// keine Liste, statt die aller Profile zu sehen (§2).
func TestOhneProfilStehtKeineProtokollpflege(t *testing.T) {
	db := frischeDatenbank(t)

	eineAltePartie(t, db, "alt",
		pflegezeit.AddDate(0, 0, -(domain.AufbewahrungPartien+1)))

	if _, fehler := storage.Pflegeanalyse(
		t.Context(), db, pflegezeit); fehler == nil {
		t.Fatal("ohne Profil kam kein Fehler")
	}

	// **Das Ausführen meldet es in seiner Bilanz und nicht als Fehler** —
	// dieselbe Form wie bei jedem anderen Grund, an dem eine Maßnahme
	// scheitert: die Zeile sagt, was war.
	zeilen, fehler := storage.PflegeAusfuehren(t.Context(), db, pflegezeit,
		map[string]string{domain.PflegeProtokolle: "egal"})
	if fehler != nil {
		t.Fatalf("ausführen: %v", fehler)
	}

	if len(zeilen) != 1 || zeilen[0].Ausgang != domain.PflegeGescheitert {
		t.Fatalf("die Bilanz ist %+v", zeilen)
	}
}
