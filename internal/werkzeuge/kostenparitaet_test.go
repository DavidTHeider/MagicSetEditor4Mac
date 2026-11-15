package werkzeuge

// **Dieselbe Fähigkeit kostet an REST und MCP dasselbe** — Folgebefund F3 des
// Lastlaufs QA-08 (#2020).
//
// # Der Befund
//
// Der Dienst las die Art eines Wegs an seiner Methode, MCP am Werkzeug: `POST
// /einstellungen/vorschau` kostete 2, `lesen.einstellungen.vorschau` 1.
// Einundzwanzig solcher Paare stehen in `docs/FAEHIGKEITEN.md` — alle
// Vorschauen und Deutungen, die `POST` heißen, weil sie einen Rumpf mitnehmen,
// und nichts schreiben.
//
// # Was er hält
//
// **Jedes Paar aus der Tabelle**, gepaart nach der Stelle in der Zelle (die
// Zelle der Fläche nennt die Wege in derselben Reihenfolge wie die des
// Werkzeugs; eine Zeile, in der die Zahl abweicht, lässt sich nicht paaren und
// geht den Weg unten). Beide Seiten werden **an den echten Funktionen**
// gefragt, nicht an einer Abschrift: der Dienst über [rest.KostenklasseFuer]
// mit der Methode und dem Pfad, MCP über [mcpnaehte.KontingentFuer] mit der
// Art, die das Werkzeug **wirklich** hat — denn `schreiben.turnier.vorschlag`
// heißt schreibend und ist lesend.
//
// **Eine Zeile, die sich nicht paaren lässt** (andere Zahl von Wegen als von
// Werkzeugen: ein Weg ohne Werkzeug, ein Werkzeug für zwei Wege), hält eine
// schwächere Zusage: **was MCP für die Fähigkeit berechnet, berechnet auch
// der Dienst für mindestens einen ihrer Wege.** Sie fängt eine Zeile, in der
// ein lesender Weg fehlt, nicht eine, in der ein zweiter Weg zu viel kostet.
//
// **Eine Zeile der Kostentafel mit der Klasse Lesen nennt nur Wege, die lesen**
// — ein Werkzeug mit der Art lesend und einen Weg, der nicht `GET` ist. Sonst
// wäre sie eine zweite Wahrheit, die niemand nachliest.
//
// # Was er nicht hält
//
// **Wie viel eine Klasse kostet** — das ist `domain.Kostentafel` — und **ob ein
// Weg, der `POST` heißt, lesen darf**: die Erlaubnis eines Zugriffstokens
// folgt weiter der Methode (`rest.schreibend`), und das aus gutem Grund
// ([0163](../../docs/entscheidungen/0163-zeichnen-belegt-einen-platz-und-lesen-kostet-lesen.md)).
// Beides ist ein Satz über Schutz und nicht über Preise.

import (
	"context"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcpnaehte"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// platzhalter fängt `{name}` und `{name...}` in einem Wegmuster.
var platzhalter = regexp.MustCompile(`\{[^}]*\}`)

// buchprobe merkt, was MCP für einen Ruf bucht.
type buchprobe struct{ klasse domain.Kostenklasse }

func (b *buchprobe) Abbuchen(_ string, klasse domain.Kostenklasse, _ bool) (bool, time.Duration) {
	b.klasse = klasse

	return true, 0
}

// werkzeugarten liest je Werkzeug, ob es schreibt — aus dem echten Katalog.
func werkzeugarten(t *testing.T) map[string]mcp.Art {
	t.Helper()

	t.Setenv("HOME", t.TempDir())

	db, oeffnen := storage.Oeffnen(t.Context(), "file:"+filepath.Join(t.TempDir(), "bestand.db"))
	if oeffnen != nil {
		t.Fatalf("Bestand öffnen: %v", oeffnen)
	}

	t.Cleanup(func() { _ = db.Close() })

	if migrieren := storage.Migriere(t.Context(), db, nil); migrieren != nil {
		t.Fatalf("migrieren: %v", migrieren)
	}

	befehle, katalog := mcpnaehte.Befehle(db, t.TempDir())
	if katalog != nil {
		t.Fatalf("der Katalog der Werkzeuge: %v", katalog)
	}

	arten := make(map[string]mcp.Art, len(befehle))

	for _, befehl := range befehle {
		arten[befehl.Name] = mcp.Lesend
		if befehl.Schreibend {
			arten[befehl.Name] = mcp.Schreibend
		}
	}

	if len(arten) < 150 {
		t.Fatalf("nur %d Werkzeuge im Katalog — der Katalog ist unvollständig", len(arten))
	}

	return arten
}

// kostenAnRest fragt den Dienst, was ein Weg der Tabelle kostet.
func kostenAnRest(t *testing.T, weg string) domain.Kostenklasse {
	t.Helper()

	methode, muster, gespalten := strings.Cut(weg, " ")
	if !gespalten {
		t.Fatalf("%q ist kein Weg mit Methode", weg)
	}

	return rest.KostenklasseFuer(methode, platzhalter.ReplaceAllString(muster, "1"))
}

// kostenAnMCP fragt den Werkzeugserver, was ein Werkzeug kostet.
func kostenAnMCP(ctx context.Context, art mcp.Art, werkzeug string) domain.Kostenklasse {
	probe := &buchprobe{}
	mcpnaehte.KontingentFuer(probe)(ctx, art, werkzeug)

	return probe.klasse
}

func TestDieKostenFolgenDerFaehigkeitAnRESTUndMCP(t *testing.T) {
	arten := werkzeugarten(t)
	namen, wege := faehigkeitswege(t)

	gepaart, abgeleitet := 0, 0

	for _, name := range namen {
		je := wege[name]
		restwege, werkzeuge := je[domain.FlaecheREST], je[domain.FlaecheMCP]

		if len(restwege) == 0 || len(werkzeuge) == 0 {
			continue
		}

		// **Jedes Werkzeug muss der Katalog kennen** — ein Name, den es nicht
		// gibt, wäre eine Zeile, die diese Prüfung still ausließe.
		kosten := make([]domain.Kostenklasse, len(werkzeuge))

		for stelle, werkzeug := range werkzeuge {
			art, da := arten[werkzeug]
			if !da {
				t.Errorf("%q: das Werkzeug %q steht nicht im Katalog des Servers", name, werkzeug)

				continue
			}

			kosten[stelle] = kostenAnMCP(t.Context(), art, werkzeug)
		}

		if len(restwege) == len(werkzeuge) {
			for stelle, weg := range restwege {
				gepaart++

				if amDienst := kostenAnRest(t, weg); amDienst != kosten[stelle] {
					t.Errorf("%q: %s kostet %s, %s kostet %s", name, weg, amDienst, werkzeuge[stelle], kosten[stelle])
				}
			}

			continue
		}

		// Eine Zeile, die sich nicht paaren lässt: was MCP berechnet, berechnet
		// der Dienst für mindestens einen Weg.
		abgeleitet++

		berechnet := map[domain.Kostenklasse]bool{}
		for _, weg := range restwege {
			berechnet[kostenAnRest(t, weg)] = true
		}

		for stelle, klasse := range kosten {
			if !berechnet[klasse] {
				t.Errorf("%q: %s kostet %s, aber kein Weg des Dienstes (%v)",
					name, werkzeuge[stelle], klasse, restwege)
			}
		}
	}

	// **Beide Hälften haben Fälle**: eine Prüfung, die nichts gepaart hat,
	// sagte nichts. Die Zahlen sind eine Untergrenze und kein Sollwert — eine
	// neue Fähigkeit macht sie größer.
	if gepaart < 200 || abgeleitet < 3 {
		t.Errorf("gepaart wurden %d Wege, abgeleitet %d Zeilen — der Leser greift daneben", gepaart, abgeleitet)
	}
}

func TestEineZeileMitDerKlasseLesenNenntNurLesendeWege(t *testing.T) {
	arten := werkzeugarten(t)
	gezaehlt := 0

	for _, zeile := range domain.Kostentafel() {
		if zeile.Klasse != domain.KostenLesen {
			continue
		}

		for _, weg := range zeile.Wege[domain.FlaecheMCP] {
			gezaehlt++

			if arten[weg] != mcp.Lesend {
				t.Errorf("%q: %s schreibt — eine Zeile Lesen darf es nicht nennen", zeile.Faehigkeit, weg)
			}
		}

		for _, weg := range zeile.Wege[domain.FlaecheREST] {
			gezaehlt++

			if methode, _, _ := strings.Cut(weg, " "); methode == http.MethodGet || methode == http.MethodHead {
				t.Errorf("%q: %s kostet nach seiner Methode ohnehin Lesen — die Zeile ist überflüssig", zeile.Faehigkeit, weg)
			}
		}
	}

	if gezaehlt == 0 {
		t.Error("die Tafel hat keine Zeile der Klasse Lesen")
	}
}
