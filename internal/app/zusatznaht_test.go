package app_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Zusatznaht kommt zusammen, listet, setzt und liest wieder** — gegen
// die Saat, deren Stil `vermerk` als Zusatzfeld führt (#2298).
func TestDieZusatznahtSetztUndLiestEinenZusatzwert(t *testing.T) {
	zuhause := t.TempDir()
	t.Setenv("HOME", zuhause)
	saatInsZuhause(t, zuhause)

	db := frischerBestand(t)
	setID, karteID := saatsetImBestand(t, db)

	weg, nahtfehler := app.ZusatznahtUeber(db, domain.UrheberKommandozeile)
	if nahtfehler != nil {
		t.Fatalf("Naht: %v", nahtfehler)
	}

	felder, lesefehler := weg.Felder(probekontext(t), setID, karteID)
	if lesefehler != nil || len(felder) != 1 || felder[0].Form.Name != "vermerk" || felder[0].Gespeichert {
		t.Fatalf("Felder %+v, Fehler %v", felder, lesefehler)
	}

	vorher, fehler := weg.Setzen(probekontext(t), setID, karteID, "vermerk", "Vorabdruck")
	if fehler != nil || vorher.Gespeichert {
		t.Fatalf("Setzen: Vorwert %+v, Fehler %v", vorher, fehler)
	}

	gelesen, fehler := weg.Felder(probekontext(t), setID, karteID)
	if fehler != nil || gelesen[0].Wert != "Vorabdruck" || !gelesen[0].Gespeichert {
		t.Errorf("nach dem Setzen %+v, Fehler %v", gelesen, fehler)
	}

	// **Der Vorwert kommt aus dem Stand, auf den geschrieben wurde** —
	// durch den echten Bestand und seine Bedingung am Kartenstand.
	vorher, fehler = weg.Setzen(probekontext(t), setID, karteID, "vermerk", "Nachdruck")
	if fehler != nil || !vorher.Gespeichert || vorher.Wert != "Vorabdruck" || !vorher.BlockDa {
		t.Errorf("zweites Setzen: Vorwert %+v, Fehler %v", vorher, fehler)
	}

	if _, fehler := weg.Setzen(probekontext(t), setID, karteID, "grund", "x"); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("ein gerechnetes Zusatzfeld ließ sich setzen: %v", fehler)
	}

	if _, fehler := weg.Felder(probekontext(t), setID, karteID+1000); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("eine unbekannte Karte gab Felder: %v", fehler)
	}
}

// **Ohne Zuhause kommt sie nicht zusammen.**
func TestDieZusatznahtMeldetIhrenNahtfehler(t *testing.T) {
	t.Setenv("HOME", "")

	if _, fehler := app.ZusatznahtUeber(frischerBestand(t), domain.UrheberKommandozeile); fehler == nil {
		t.Error("kein Fehler ohne Zuhause")
	}
}
