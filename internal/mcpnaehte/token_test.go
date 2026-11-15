package mcpnaehte

// **Die Verdrahtung der Zugriffstoken** — Werkstück 10 von Phase 4.12
// (#1351).
//
// # Was dieser Lauf hält
//
// Dass die Wurzel den Weg **erreicht** und die beiden Abbildungen die Felder
// nicht vertauschen — nicht, was der Dienst dahinter tut. Der Arbeitsablauf
// über alle Flächen steht in `internal/flaechen`.

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// frischerBestand gibt einen migrierten Bestand im Zuhause dieses Laufs.
func frischerBestand(t *testing.T) *sql.DB {
	db, _ := frischerBestandMitZuhause(t)

	return db
}

// frischerBestandMitZuhause gibt dazu den Ort des Zuhauses.
//
// **Er kommt von hier und nicht aus `os.Getenv`**: gosec hält eine Umgebung
// für eine fremde Eingabe und macht aus jedem Pfad darunter einen Befund
// (G703) — zu Recht, denn im Betrieb ist sie eine. Hier ist sie ein
// Verzeichnis, das dieser Lauf gerade selbst angelegt hat, und das gehört
// weitergereicht statt zurückgelesen.
func frischerBestandMitZuhause(t *testing.T) (db *sql.DB, zuhause string) {
	t.Helper()

	// **Ein eigenes Zuhause** (§3): sonst sähe der Lauf den Stand dieses
	// Rechners und wäre grün, weil er wirklich funktioniert hat.
	zuhause = t.TempDir()
	t.Setenv("HOME", zuhause)

	db, err := storage.Oeffnen(probekontext(t),
		"file:"+filepath.Join(zuhause, "bestand.db"))
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := storage.Migriere(probekontext(t), db, nil); fehler != nil {
		t.Fatalf("migrieren: %v", fehler)
	}

	return db, zuhause
}

// **Der ganze Weg über die Naht.**
func TestDieTokennahtErreichtDenDienst(t *testing.T) {
	naht := token{db: frischerBestand(t)}

	angelegt, klartext, fehler := naht.TokenAnlegen(probekontext(t),
		mcp.Tokenangaben{
			ProfilID: "1", Bezeichnung: "Nachtlauf", Principal: "Rechner", Scope: "lesen",
			Ablauf: "2099-01-01T00:00:00Z",
		})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if !strings.HasPrefix(klartext, "mse_") {
		t.Errorf("der Klartext sieht nicht wie ein Token aus: %q", klartext)
	}

	// **Jedes Feld einzeln**, weil eine Abbildung mit acht Feldern acht
	// Stellen hat, an denen ein Wert im falschen landet.
	faelle := map[string]struct{ ist, soll string }{
		"ProfilID":    {angelegt.ProfilID, "1"},
		"Bezeichnung": {angelegt.Bezeichnung, "Nachtlauf"},
		"Scope":       {angelegt.Scope, "lesen"},
		"Ablauf":      {angelegt.Ablauf, "2099-01-01T00:00:00Z"},
	}

	for name, fall := range faelle {
		if fall.ist != fall.soll {
			t.Errorf("%s = %q, erwartet %q", name, fall.ist, fall.soll)
		}
	}

	if angelegt.Abgelaufen {
		t.Error("das frische Token gilt als abgelaufen")
	}

	zeilen, fehler := naht.Token(probekontext(t), "1")
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(zeilen) != 1 || zeilen[0].ID != angelegt.ID {
		t.Fatalf("gelesen = %+v", zeilen)
	}

	geaendert, fehler := naht.TokenAendern(probekontext(t), mcp.Tokenangaben{
		TokenID: angelegt.ID, Bezeichnung: "Tagesbau", Principal: "Rechner", Scope: "schreiben",
		Fassung: angelegt.Fassung,
	})
	if fehler != nil {
		t.Fatalf("ändern: %v", fehler)
	}

	if geaendert.Bezeichnung != "Tagesbau" || geaendert.Scope != "schreiben" {
		t.Errorf("geändert = %+v", geaendert)
	}

	if loeschfehler := naht.TokenLoeschen(
		probekontext(t), angelegt.ID); loeschfehler != nil {
		t.Fatalf("löschen: %v", loeschfehler)
	}

	nachher, fehler := naht.Token(probekontext(t), "1")
	if fehler != nil {
		t.Fatalf("neu lesen: %v", fehler)
	}

	if len(nachher) != 0 {
		t.Errorf("nach dem Löschen stehen noch %d da", len(nachher))
	}
}

// **Und jeder der vier Wege reicht sein Nein durch.**
//
// **Ohne diese Fälle bliebe die Hälfte jeder Naht unbetreten** — und eine
// Naht, die nur den Erfolgsweg kennt, meldet einen Fehlschlag als Erfolg
// (§10).
func TestDieTokennahtReichtIhrNeinDurch(t *testing.T) {
	naht := token{db: frischerBestand(t)}

	if _, fehler := naht.Token(probekontext(t), "keine-zahl"); fehler == nil {
		t.Error("eine krumme Profilkennung kam durch")
	}

	if _, _, fehler := naht.TokenAnlegen(probekontext(t),
		mcp.Tokenangaben{ProfilID: "4711", Bezeichnung: "Nirgends", Principal: "Rechner", Scope: "lesen"},
	); fehler == nil {
		t.Error("ein Token auf einem Profil, das es nicht gibt, kam durch")
	}

	if _, fehler := naht.TokenAendern(probekontext(t),
		mcp.Tokenangaben{TokenID: "4711", Bezeichnung: "X", Principal: "Rechner", Scope: "lesen", Fassung: "1"},
	); fehler == nil {
		t.Error("ein Token, das es nicht gibt, ließ sich ändern")
	}

	if fehler := naht.TokenLoeschen(probekontext(t), "4711"); fehler == nil {
		t.Error("ein Token, das es nicht gibt, ließ sich löschen")
	}
}
