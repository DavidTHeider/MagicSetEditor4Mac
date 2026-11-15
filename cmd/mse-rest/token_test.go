package main

// **Die Verdrahtung der Zugriffstoken im Dienst** — Werkstück 10 von
// Phase 4.12
// (#1351).
//
// # Was dieser Lauf hält
//
// Dass die Wurzel den Weg **erreicht** und die beiden Abbildungen die Felder
// nicht vertauschen. Dass ein Token am **laufenden** Dienst an der Stelle der
// Prozessmarke gilt, steht in `internal/flaechen` — das braucht zwei
// Prozesse.

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// frischerTokenbestand gibt einen migrierten Bestand im Zuhause dieses Laufs.
func frischerTokenbestand(t *testing.T) *sql.DB {
	db, _ := frischerTokenbestandMitZuhause(t)

	return db
}

// frischerTokenbestandMitZuhause gibt dazu den Ort des Zuhauses.
//
// **Er kommt von hier und nicht aus `os.Getenv`**: gosec hält eine Umgebung
// für eine fremde Eingabe und macht aus jedem Pfad darunter einen Befund
// (G703) — zu Recht, denn im Betrieb ist sie eine. Hier ist sie ein
// Verzeichnis, das dieser Lauf gerade selbst angelegt hat.
func frischerTokenbestandMitZuhause(t *testing.T) (db *sql.DB, zuhause string) {
	t.Helper()

	// **Ein eigenes Zuhause** (§3): sonst sähe der Lauf den Stand dieses
	// Rechners und wäre grün, weil er wirklich funktioniert hat.
	zuhause = t.TempDir()
	t.Setenv("HOME", zuhause)

	db, fehler := storage.Oeffnen(probekontext(t),
		filepath.Join(zuhause, "bestand.db"))
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	return db, zuhause
}

// **Der ganze Weg über die Naht.**
func TestDieTokennahtDesDienstesErreichtDieAnwendung(t *testing.T) {
	naht := fremdtoken{db: frischerTokenbestand(t)}

	angelegt, klartext, fehler := naht.TokenAnlegen(probekontext(t),
		rest.Tokenangaben{
			ProfilID: "1", Bezeichnung: "Nachtlauf", Scope: "schreiben",
			Principal: "Rechner", Ablauf: "2099-01-01T00:00:00Z",
		})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if !strings.HasPrefix(klartext, "mse_") {
		t.Errorf("der Klartext sieht nicht wie ein Token aus: %q", klartext)
	}

	faelle := map[string]struct{ ist, soll string }{
		"ProfilID":    {angelegt.ProfilID, "1"},
		"Bezeichnung": {angelegt.Bezeichnung, "Nachtlauf"},
		"Scope":       {angelegt.Scope, "schreiben"},
		"Ablauf":      {angelegt.Ablauf, "2099-01-01T00:00:00Z"},
	}

	for name, fall := range faelle {
		if fall.ist != fall.soll {
			t.Errorf("%s = %q, erwartet %q", name, fall.ist, fall.soll)
		}
	}

	zeilen, fehler := naht.Token(probekontext(t), "1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(zeilen) != 1 || zeilen[0].ID != angelegt.ID {
		t.Fatalf("gelesen = %+v", zeilen)
	}

	geaendert, fehler := naht.TokenAendern(probekontext(t), rest.Tokenangaben{
		TokenID: angelegt.ID, Bezeichnung: "Tagesbau", Scope: "lesen",
		Principal: "Rechner", Fassung: angelegt.Fassung,
	})
	if fehler != nil {
		t.Fatalf("ändern: %v", fehler)
	}

	if geaendert.Bezeichnung != "Tagesbau" || geaendert.Scope != "lesen" {
		t.Errorf("geändert = %+v", geaendert)
	}

	if fehler := naht.TokenLoeschen(probekontext(t), angelegt.ID); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}
}

// **Und jeder der vier Wege reicht sein Nein durch.**
func TestDieTokennahtDesDienstesReichtIhrNeinDurch(t *testing.T) {
	naht := fremdtoken{db: frischerTokenbestand(t)}

	if _, fehler := naht.Token(probekontext(t), "keine-zahl"); fehler == nil {
		t.Error("eine krumme Profilkennung kam durch")
	}

	if _, _, fehler := naht.TokenAnlegen(probekontext(t),
		rest.Tokenangaben{ProfilID: "4711", Bezeichnung: "Nirgends", Scope: "lesen", Principal: "Rechner"},
	); fehler == nil {
		t.Error("ein Token auf einem Profil, das es nicht gibt, kam durch")
	}

	if _, fehler := naht.TokenAendern(probekontext(t),
		rest.Tokenangaben{TokenID: "4711", Bezeichnung: "X", Scope: "lesen", Principal: "Rechner", Fassung: "1"},
	); fehler == nil {
		t.Error("ein Token, das es nicht gibt, ließ sich ändern")
	}

	if fehler := naht.TokenLoeschen(probekontext(t), "4711"); fehler == nil {
		t.Error("ein Token, das es nicht gibt, ließ sich löschen")
	}
}
