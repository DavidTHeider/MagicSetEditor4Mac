package storage

// **Die Fehlerwege des Ausgangs** — Werkstück `T5` von Phase 5.6
// (#1470).
//
// **Weil ein Ausgang, dessen Fehlerwege ungeprüft sind, genau dort schweigt,
// wo er reden müsste**: eine Zeile, die sich nicht lesen lässt, darf nicht als
// „nichts liegt an" durchgehen — dann bliebe eine Nachricht für immer liegen,
// und niemand wüsste davon.

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"strings"
	"testing"
)

// Die Stellen, an denen der Ausgang stolpern kann.
const (
	beimAusgangsscan stolperstelle = iota + 2600
	beimAusgangsende
)

func init() {
	sql.Register("stolper-ausgang-scan", stolpertreiber{stelle: beimAusgangsscan})
	sql.Register("stolper-ausgang-ende", stolpertreiber{stelle: beimAusgangsende})
}

// ausgangsabfrage beantwortet die eine Abfrage des Ausgangs.
func ausgangsabfrage(
	stelle stolperstelle, abfrage string,
) (driver.Rows, bool) {
	if !strings.HasPrefix(strings.TrimSpace(abfrage),
		"SELECT nachricht_id, vorgang, nachricht, versuche") {
		return nil, false
	}

	spalten := []string{"nachricht_id", "vorgang", "nachricht", "versuche"}

	switch stelle {
	case beimAusgangsscan:
		return &krummeZeile{spalten: spalten}, true

	case beimAusgangsende:
		return &abbrechendeZeile{spalten: spalten}, true
	}

	return nil, false
}

// **Auch der Stand kann stolpern** — und dann ist er kein leerer Stand.
func TestEinKrummerPoststandMeldetSich(t *testing.T) {
	if _, fehler := NeueAusgangsablage(stolperbestand(t, "stolper-fragen")).
		Stand(probekontext(t), "bgegenueber"); fehler == nil {
		t.Error("der krumme Stand kam durch")
	}
}

func TestEinKrummerAusgangMeldetSich(t *testing.T) {
	for name, treiber := range map[string]string{
		"eine Zeile, die sich nicht lesen lässt": "stolper-ausgang-scan",
		"ein Lauf, der mittendrin abbricht":      "stolper-ausgang-ende",
		"eine Abfrage, die gar nicht geht":       "stolper-fragen",
	} {
		t.Run(name, func(t *testing.T) {
			if _, fehler := NeueAusgangsablage(stolperbestand(t, treiber)).
				Offene(probekontext(t), "bgegenueber"); fehler == nil {
				t.Error("der krumme Ausgang kam durch")
			}
		})
	}
}

// **Die zwei zählenden Wege melden sich, wenn das Zählen scheitert.**
func TestEinZaehlenAmAusgangKannScheitern(t *testing.T) {
	db := stolperbestand(t, "stolper-zaehlen")
	ablage := NeueAusgangsablage(db)

	if fehler := ablage.Versucht(probekontext(t), "n-1"); fehler == nil {
		t.Error("der Versuch kam durch")
	}

	if fehler := ablage.Bestaetigen(probekontext(t), "n-1"); fehler == nil {
		t.Error("die Bestätigung kam durch")
	}
}

// **Und eine Anweisung, die gar nicht geht.**
func TestEineAnweisungAmAusgangKannScheitern(t *testing.T) {
	db := stolperbestand(t, "stolper-schreiben")
	ablage := NeueAusgangsablage(db)

	if fehler := ablage.Versucht(probekontext(t), "n-1"); fehler == nil {
		t.Error("der Versuch kam durch")
	}

	if fehler := ablage.Einreihen(probekontext(t), Ausgangszeile{
		NachrichtID: "n-1", Vorgang: "v-1", Empfaenger: "b", Nachricht: "{}",
	}); fehler == nil {
		t.Error("das Einreihen kam durch")
	}
}

// **Ohne Profil geht am Ausgang nichts** (0066) — auf allen drei Wegen.
func TestOhneProfilGehtAmAusgangNichts(t *testing.T) {
	ablage := NeueAusgangsablage(migrierteDatenbank(t))
	ohne := context.Background()

	if fehler := ablage.Einreihen(ohne, Ausgangszeile{}); fehler == nil {
		t.Error("das Einreihen kam ohne Profil durch")
	}

	if _, fehler := ablage.Offene(ohne, "b"); fehler == nil {
		t.Error("das Lesen kam ohne Profil durch")
	}

	if fehler := ablage.Versucht(ohne, "n-1"); fehler == nil {
		t.Error("der Versuch kam ohne Profil durch")
	}

	if _, fehler := ablage.Stand(ohne, "b"); fehler == nil {
		t.Error("der Stand kam ohne Profil durch")
	}
}
