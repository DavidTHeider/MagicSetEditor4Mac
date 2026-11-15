package storage_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **`styling_values` stand seit Phase 2 im Schema und war leer.**
//
// Dieselbe unerfüllte Zusage wie bei `packages` (#54);
// #41 löst sie
// ein.
func TestDerStilwechselSchreibtSpielStilUndOptionen(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSetSpeicher(db)
	ablage := storage.NeueStilablage(db)

	setID, err := speicher.SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "probe", Style: "probe-schlicht",
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	werte := []domain.Value{
		{Field: "hintergrund", Text: "dunkel"},
		{Field: "rahmen zeigen", Text: "yes"},
	}
	if err := ablage.StilSetzen(probekontext(t), setID, "probe", "probe-bild", werte); err != nil {
		t.Fatalf("setzen: %v", err)
	}

	gelesen, fehler := ablage.Stilwerte(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(gelesen) != 2 || gelesen[0].Field != "hintergrund" || gelesen[1].Text != "yes" {
		t.Fatalf("Werte: %+v", gelesen)
	}

	koepfe, kopffehler := storage.NeuerSetLeser(db).Setkoepfe(probekontext(t))
	if kopffehler != nil {
		t.Fatalf("Köpfe: %v", kopffehler)
	}

	if koepfe[0].Style != "probe-bild" {
		t.Errorf("der Stil steht auf %q", koepfe[0].Style)
	}
}

// **Was der neue Stil nicht kennt, steht danach nicht mehr da.**
//
// Ersetzen und nicht ergänzen: ein Wert ohne Option wäre ein Wert, den kein
// Skript liest und den niemand ändern kann.
func TestEinZweiterWechselErsetztDieOptionen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueStilablage(db)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), &domain.Set{
		Name: "Probeset", Game: "probe", Style: "probe-schlicht",
	}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	erst := []domain.Value{{Field: "hintergrund", Text: "dunkel"}, {Field: "alt", Text: "ja"}}
	if fehler := ablage.StilSetzen(probekontext(t), setID, "probe", "probe-bild", erst); fehler != nil {
		t.Fatalf("erster Wechsel: %v", fehler)
	}

	bleibt := []domain.Value{{Field: "hintergrund", Text: "dunkel"}}
	if fehler := ablage.StilSetzen(probekontext(t), setID, "probe", "probe-schlicht", bleibt); fehler != nil {
		t.Fatalf("zweiter Wechsel: %v", fehler)
	}

	gelesen, fehler := ablage.Stilwerte(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(gelesen) != 1 || gelesen[0].Field != "hintergrund" {
		t.Errorf("Werte: %+v", gelesen)
	}
}

// **Ein Set, das es nicht gibt, ist ein `ErrNotFound`.**
//
// Ohne dieses Urteil ginge ein Wechsel an einem gelöschten Set als geglückt
// durch — und die Oberfläche zeigte einen Stil, den niemand trägt.
func TestEinStilwechselAmUnbekanntenSetIstNichtGefunden(t *testing.T) {
	ablage := storage.NeueStilablage(frischeDatenbank(t))

	err := ablage.StilSetzen(probekontext(t), 4711, "probe", "probe-bild", nil)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("kein ErrNotFound: %v", err)
	}
}

// **Ein Set ohne Optionen gibt eine leere Liste und kein `nil`.**
func TestOhneOptionenEineLeereListe(t *testing.T) {
	db := frischeDatenbank(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Probeset", Game: "probe"}, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	werte, fehler := storage.NeueStilablage(db).Stilwerte(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if werte == nil || len(werte) != 0 {
		t.Errorf("Werte: %+v", werte)
	}
}
