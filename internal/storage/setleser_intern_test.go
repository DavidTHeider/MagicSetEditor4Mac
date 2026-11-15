// Prüfungen am `SetLeser`, die einen Treiber brauchen, der stolpert.
//
// Im Paket und nicht daneben, weil der Stolpertreiber hier steht.
package storage

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Eine Zeile, die sich nicht lesen lässt, wird gemeldet — nicht übergangen.
//
// Mit einer echten SQLite-Datenbank ist der Fall nicht herzustellen: `STRICT`
// sorgt dafür, dass alles hineinpasst, was hineinkam. Im Betrieb entsteht er an
// einem Schema, das jemand von Hand verändert hat — und dann ist eine leere
// Liste die schlechteste aller Antworten, weil sie wie ein leerer Bestand
// aussieht.
func TestDerLeserMeldetEineUnlesbareZeile(t *testing.T) {
	db, err := sql.Open("stolper-umwandeln", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	leser := NeuerSetLeser(db)

	if _, fehler := leser.Setkoepfe(probekontext(t)); fehler == nil {
		t.Error("Setkoepfe übersah eine unlesbare Zeile")
	}

	if _, fehler := leser.Karten(probekontext(t), 1); fehler == nil {
		t.Error("Karten übersah eine unlesbare Zeile")
	}

	if _, fehler := leser.Rohbloecke(probekontext(t), 1); fehler == nil {
		t.Error("Rohbloecke übersah eine unlesbare Zeile")
	}
}

// Auch `KarteLesen` meldet eine unlesbare Zeile.
func TestKarteLesenMeldetEineUnlesbareZeile(t *testing.T) {
	db, err := sql.Open("stolper-umwandeln", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	if _, fehler := NeuerSetLeser(db).KarteLesen(probekontext(t), 1); fehler == nil {
		t.Error("KarteLesen übersah eine unlesbare Zeile")
	}
}

// Und einen Lesefehler nach der Schleife.
func TestKarteLesenMeldetEinenLesefehlerNachDerSchleife(t *testing.T) {
	db, err := sql.Open("stolper-lesen", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	if _, fehler := NeuerSetLeser(db).KarteLesen(probekontext(t), 1); fehler == nil {
		t.Error("KarteLesen übersah den Lesefehler")
	}
}

// Auch die Suche meldet eine unlesbare Zeile.
func TestDieSucheMeldetEineUnlesbareZeile(t *testing.T) {
	db, err := sql.Open("stolper-umwandeln", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	leser := NeuerSetLeser(db)

	if _, fehler := leser.Suche(probekontext(t), domain.SearchQuery{Text: "a"}); fehler == nil {
		t.Error("die Suche übersah eine unlesbare Zeile")
	}

	if _, fehler := leser.Kartenwerte(probekontext(t), []int64{1}); fehler == nil {
		t.Error("Kartenwerte übersah eine unlesbare Zeile")
	}

	if _, fehler := leser.Felder(probekontext(t), 0); fehler == nil {
		t.Error("Felder übersah eine unlesbare Zeile")
	}
}

// Und einen Lesefehler **nach** der Schleife.
//
// Der Zweig ist erst durch die Kürzung entstanden: vorher gab `Suche` den
// Fehler zusammen mit den Treffern zurück, jetzt muss sie ihn vorziehen — sonst
// zählte sie Treffer, die sie gar nicht gelesen hat.
func TestDieSucheMeldetEinenLesefehlerNachDerSchleife(t *testing.T) {
	db, err := sql.Open("stolper-lesen", "")
	if err != nil {
		t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	leser := NeuerSetLeser(db)

	if _, fehler := leser.Suche(probekontext(t), domain.SearchQuery{Text: "a"}); fehler == nil {
		t.Error("die Suche übersah den Lesefehler")
	}

	if _, fehler := leser.Kartenwerte(probekontext(t), []int64{1}); fehler == nil {
		t.Error("Kartenwerte übersah den Lesefehler")
	}

	if _, fehler := leser.Felder(probekontext(t), 0); fehler == nil {
		t.Error("Felder übersah den Lesefehler")
	}
}

// Auch die Statistik meldet eine unlesbare Zeile und einen Lesefehler.
func TestDieStatistikMeldetUnlesbareZeilen(t *testing.T) {
	for _, treiber := range []string{"stolper-umwandeln", "stolper-lesen"} {
		t.Run(treiber, func(t *testing.T) {
			db, err := sql.Open(treiber, "")
			if err != nil {
				t.Fatalf("Stolpertreiber nicht öffnenbar: %v", err)
			}

			t.Cleanup(func() { _ = db.Close() })

			leser := NeuerSetLeser(db)

			if _, fehler := leser.Kennzahlen(probekontext(t), 1); fehler == nil {
				t.Error("Kennzahlen übersah die Zeile")
			}

			if _, fehler := leser.Belegung(probekontext(t), 1); fehler == nil {
				t.Error("Belegung übersah die Zeile")
			}

			if _, fehler := leser.Verteilung(probekontext(t), 1, "typ"); fehler == nil {
				t.Error("Verteilung übersah die Zeile")
			}
		})
	}
}
