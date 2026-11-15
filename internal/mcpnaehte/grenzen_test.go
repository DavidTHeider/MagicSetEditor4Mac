package mcpnaehte

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Was ein Werkzeug kostet** — Werkstück O5 von Phase 5.2 (#1429).

func TestJedesWerkzeugHatSeineKostenklasse(t *testing.T) {
	t.Parallel()

	// **Was die Kostentafel nennt, kostet hier dasselbe** (#2003) — eine Karte
	// zeichnen 5, ein Bogen und MTG JSON schreiben 8, an jeder Fläche.
	for _, zeile := range domain.Kostentafel() {
		for _, werkzeug := range zeile.Wege[domain.FlaecheMCP] {
			art := mcp.Lesend
			if strings.HasPrefix(werkzeug, "schreiben.") {
				art = mcp.Schreibend
			}

			if klasse := kostenklasseFuer(art, werkzeug); klasse != zeile.Klasse {
				t.Errorf("%s: %s, die Tafel sagt %s", werkzeug, klasse, zeile.Klasse)
			}
		}
	}

	// **Sonst die Art — und kein Wort im Namen.** Bis zum 2026-10-02 kostete
	// eine Vorschau der Einstellungen hier 5, weil ihr Name „vorschau" trägt.
	faelle := map[string]struct {
		art  mcp.Art
		name string
		will domain.Kostenklasse
	}{
		"lesen":                {mcp.Lesend, "lesen.sets", domain.KostenLesen},
		"schreiben":            {mcp.Schreibend, "schreiben.feld", domain.KostenSchreiben},
		"vorschau rastert nie": {mcp.Lesend, "lesen.einstellungen.vorschau", domain.KostenLesen},
		"Turnier ist kein Set": {mcp.Schreibend, "schreiben.turnier.einlesen", domain.KostenSchreiben},
		"Bild wegnehmen":       {mcp.Schreibend, "schreiben.nutzer.bild.loeschen", domain.KostenSchreiben},
	}

	for name, fall := range faelle {
		if klasse := kostenklasseFuer(fall.art, fall.name); klasse != fall.will {
			t.Errorf("%s: %s, erwartet %s", name, klasse, fall.will)
		}
	}

	// Die Naht bucht bei der Sitzung — und sagt die Wartezeit, wenn sie leer ist.
	naht := KontingentFuer(&buchprobe{})

	if ok, _ := naht(t.Context(), mcp.Lesend, "lesen.sets"); !ok {
		t.Fatal("der erste Ruf abgewiesen")
	}

	if ok, warten := naht(t.Context(), mcp.Lesend, "lesen.sets"); ok ||
		warten != time.Second {
		t.Errorf("der zweite Ruf: ok=%v warten=%v", ok, warten)
	}
}

// **Der Server nennt, was er beim Start gelesen hat** — QA-08.AK-03 (#1739):
// seine Guthaben entstehen dort, und eine Auskunft, die den Bestand von
// jetzt läse, spräche nach einem Umschalten von Grenzen, die er nicht hat.
func TestDerServerNenntDieGrenzenVomStart(t *testing.T) {
	t.Parallel()

	ohne := domain.Vorgabegrenzen()
	ohne.OhneKontingente = true

	naehte := MitGrenzenVomStart(mcp.Naehte{}, ohne)

	grenzen, err := naehte.Grenzen.Grenzen(t.Context())
	if err != nil || grenzen.Kontingente != domain.KontingenteAus || grenzen.Principal.Burst != ohne.Principal.Burst {
		t.Errorf("Grenzen: %+v, %v", grenzen, err)
	}

	// **Und Auftrags- und Verbindungsgrenze samt Kostentafel** (#2003) —
	// dieselbe Auskunft wie an den anderen Flächen.
	if grenzen.Verbindungen != 16 ||
		grenzen.Auftraege != (mcp.Auftragsgrenzen{Laufend: 4, Behalten: 16, AufbewahrungSekunden: 3600}) ||
		len(grenzen.Kostentafel) != len(domain.Kostentafel()) {
		t.Errorf("Grenzen: %+v", grenzen)
	}
}

// **Die Kommandozentrale des Fensters fragt den Bestand von jetzt** — wie
// die Hilfe daneben: sie nennt, was ein jetzt gestarteter Dienst anwendete.
func TestDieKommandozentraleNenntDieGrenzenDesBestands(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)

	db, fehler := storage.Oeffnen(probekontext(t), filepath.Join(zuhause, "bestand.db"))
	if fehler != nil {
		t.Fatalf("öffnen: %v", fehler)
	}

	t.Cleanup(func() { _ = db.Close() })

	if err := storage.Migriere(probekontext(t), db, nil); err != nil {
		t.Fatalf("migrieren: %v", err)
	}

	gezeigt, fehler := storage.MitGezeigtemProfil(probekontext(t), db)
	if fehler != nil {
		t.Fatal(fehler)
	}

	if gesetzt := app.EinstellungsnahtUeber(db).Setzen(gezeigt,
		domain.GrenzenKontingente, domain.KontingenteAus); gesetzt != nil {
		t.Fatalf("setzen: %v", gesetzt)
	}

	grenzen, err := (grenzenauskunft{db: db}).Grenzen(probekontext(t))
	if err != nil || grenzen.Kontingente != domain.KontingenteAus {
		t.Errorf("Grenzen: %+v, %v", grenzen, err)
	}

	// **Ein Bestand, der nicht antwortet, ist ein Fehler** — und keine
	// Vorgabe, die so aussähe wie eine Auskunft.
	if geschlossen := db.Close(); geschlossen != nil {
		t.Fatal(geschlossen)
	}

	if _, err := (grenzenauskunft{db: db}).Grenzen(probekontext(t)); err == nil {
		t.Error("ein geschlossener Bestand gab Grenzen")
	}
}

// buchprobe lässt genau einen Ruf durch und nennt danach eine Sekunde.
type buchprobe struct {
	gebucht    int
	principale []string
}

func (b *buchprobe) Abbuchen(
	principal string, _ domain.Kostenklasse, _ bool,
) (bool, time.Duration) {
	b.gebucht++
	b.principale = append(b.principale, principal)

	return b.gebucht == 1, time.Second
}

// **Gezählt wird nach der Maschine, die der Ruf vorzeigt.**
//
// Befund 3 des Nachlaufs zum Zweitblick über Phase 5.2: die Naht bekam keinen
// Kontext und buchte immer auf das Wort `mcp` — während am MCP-Server der
// Satz stand, der Haushalt lese die Identität aus dem Kontext.
func TestDasGuthabenZaehltNachDerMaschine(t *testing.T) {
	t.Parallel()

	buch := &buchprobe{}
	naht := KontingentFuer(buch)

	naht(service.MitTokenidentitaet(t.Context(), "7:Bauwerk"), mcp.Lesend, "lesen.sets")

	// **Und ohne Token bleibt der Altbetrieb** — die Gegenprobe: ohne sie
	// sagte der Lauf nur, dass *irgendein* Wort ankommt.
	naht(t.Context(), mcp.Lesend, "lesen.sets")

	if !slices.Equal(buch.principale, []string{"7:Bauwerk", principalMCP}) {
		t.Errorf("gebucht wurde auf %v", buch.principale)
	}
}
