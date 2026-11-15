package storage_test

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Ein gelesenes Set trägt seine Angaben und die Zahl seiner Karten.
func TestSetkoepfeGebenDieAngabenUndDieKartenzahl(t *testing.T) {
	db := frischeDatenbank(t)

	set := probeSet()
	set.Origin = "/pfad/zum/probeset.mse-set"

	if _, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), set, nil); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	koepfe, err := storage.NeuerSetLeser(db).Setkoepfe(probekontext(t))
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(koepfe) != 1 {
		t.Fatalf("%d Sets, erwartet war eines", len(koepfe))
	}

	kopf := koepfe[0]
	if kopf.Name != "Mein Probeset" || kopf.Game != "probe" || kopf.Style != "probe-stil" {
		t.Errorf("%+v", kopf)
	}

	if kopf.Origin != "/pfad/zum/probeset.mse-set" {
		t.Errorf("Herkunft %q", kopf.Origin)
	}

	// **Die Kartenzahl kommt mit, die Karten nicht.** Die Liste soll auch dann
	// stehen, wenn ein Set 3.000 Karten hat.
	if kopf.Cards != 2 {
		t.Errorf("%d Karten, erwartet waren 2", kopf.Cards)
	}

	if kopf.ID == 0 {
		t.Error("keine Kennung")
	}
}

func TestSetkoepfeOhneSetsGebenNichts(t *testing.T) {
	koepfe, err := storage.NeuerSetLeser(frischeDatenbank(t)).Setkoepfe(probekontext(t))
	if err != nil || len(koepfe) != 0 {
		t.Errorf("%v / %v", koepfe, err)
	}
}

// Die Karten kommen in ihrer Reihenfolge, mit ihren Feldwerten.
func TestKartenKommenInIhrerReihenfolge(t *testing.T) {
	db := frischeDatenbank(t)

	kennung, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), probeSet(), nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	karten, err := storage.NeuerSetLeser(db).Karten(probekontext(t), kennung)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(karten) != 2 {
		t.Fatalf("%d Karten, erwartet waren 2", len(karten))
	}

	if wert, _ := werteVon(karten[0]).GetName(); wert != "Feuerball" {
		t.Errorf("die erste Karte ist %q — die Reihenfolge stimmt nicht", wert)
	}

	if karten[0].Number != 0 || karten[1].Number != 1 {
		t.Errorf("Nummern %d und %d", karten[0].Number, karten[1].Number)
	}
}

// **Eine Karte ohne einen einzigen Feldwert bleibt eine Karte.**
//
// Phase 2 hat sie eigens aufgehoben, statt sie beim Import wegzuwerfen. Ein
// innerer Verbund ließe sie beim Lesen wieder verschwinden — dann wäre die
// Arbeit von damals umsonst gewesen.
func TestEineKarteOhneFeldwerteVerschwindetNicht(t *testing.T) {
	db := frischeDatenbank(t)

	set := probeSet()
	set.Cards = append(set.Cards, &domain.Card{})

	kennung, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), set, nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	karten, err := storage.NeuerSetLeser(db).Karten(probekontext(t), kennung)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(karten) != 3 {
		t.Fatalf("%d Karten, erwartet waren 3 — die leere ist verschwunden", len(karten))
	}

	if len(karten[2].Values) != 0 {
		t.Errorf("die leere Karte trägt %d Werte", len(karten[2].Values))
	}
}

// **Die Karten eines Sets, das es nicht gibt, sind „nicht gefunden"** — und
// nicht eine leere Liste.
//
// Bis #1815 stand hier das Gegenteil, und `service.SetDaSein` fing es eine
// Stufe höher ab (#1336). Seit die Profilklemme fragt, ob das Set dem
// gezeigten Profil gehört, gibt ein fehlendes dieselbe Antwort wie ein
// fremdes: dass es das eine gibt und das andere nicht, verriete eine leere
// Liste gegen einen Fehler.
func TestKartenEinesUnbekanntenSetsSindNichtGefunden(t *testing.T) {
	karten, err := storage.NeuerSetLeser(frischeDatenbank(t)).Karten(probekontext(t), 999)
	if !errors.Is(err, domain.ErrNotFound) || len(karten) != 0 {
		t.Errorf("%v / %v", karten, err)
	}
}

// Die Rohblöcke sind die Grundlage der Vorbehalte aus dem Import.
func TestRohbloeckeKommenInIhrerReihenfolge(t *testing.T) {
	db := frischeDatenbank(t)

	kennung, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), probeSet(),
		[]domain.RawBlock{
			{File: "set", Line: 3, Text: "unverstanden: eins"},
			{File: "set", Line: 9, Text: "unverstanden: zwei"},
		})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	bloecke, err := storage.NeuerSetLeser(db).Rohbloecke(probekontext(t), kennung)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(bloecke) != 2 || bloecke[0].Line != 3 || bloecke[1].Line != 9 {
		t.Fatalf("%+v", bloecke)
	}

	if bloecke[0].Text != "unverstanden: eins" {
		t.Errorf("%q", bloecke[0].Text)
	}
}

func TestRohbloeckeOhneWelcheGebenNichts(t *testing.T) {
	db := frischeDatenbank(t)

	kennung, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t), probeSet(), nil)
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	bloecke, err := storage.NeuerSetLeser(db).Rohbloecke(probekontext(t), kennung)
	if err != nil || len(bloecke) != 0 {
		t.Errorf("%v / %v", bloecke, err)
	}
}

// Eine geschlossene Datenbank meldet sich, statt leer zurückzukommen.
func TestDerLeserMeldetEineGeschlosseneDatenbank(t *testing.T) {
	db, err := storage.Oeffnen(probekontext(t), "file:"+t.TempDir()+"/zu.db")
	if err != nil {
		t.Fatalf("öffnen: %v", err)
	}

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	leser := storage.NeuerSetLeser(db)

	if _, err := leser.Setkoepfe(probekontext(t)); err == nil {
		t.Error("Setkoepfe kam ohne Fehler zurück")
	}

	if _, err := leser.Karten(probekontext(t), 1); err == nil {
		t.Error("Karten kam ohne Fehler zurück")
	}

	if _, err := leser.Rohbloecke(probekontext(t), 1); err == nil {
		t.Error("Rohbloecke kam ohne Fehler zurück")
	}
}

// ---------------------------------------------------------------- Vorrichtung

type kartenwerte []domain.Value

func werteVon(karte *domain.CardRow) kartenwerte { return karte.Values }

func (w kartenwerte) GetName() (string, bool) {
	for _, wert := range w {
		if wert.Field == "name" {
			return wert.Text, true
		}
	}

	return "", false
}

// **Jede gelesene Karte trägt ihr Set** —
// #1164.
//
// # Warum das zählt
//
// Der Planbauer sucht den Stil **des Sets** (`StilFuerKarte`). Eine Karte mit
// `SetID: 0` findet keines, und gezeichnet wird der Behelfsplan — ohne dass
// irgendwo ein Fehler steht. Gefunden an der Vorlagenvorschau: sie zeigte eine
// Karte, änderte sich aber bei keiner Änderung an der Vorlage.
func TestJedeGeleseneKarteTraegtIhrSet(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)

	setID, err := storage.NeuerSetSpeicher(db).SetAnlegen(probekontext(t),
		&domain.Set{Name: "Probe", Game: "magic"}, nil)
	if err != nil {
		t.Fatalf("Set anlegen: %v", err)
	}

	if _, fehler := storage.NeuerKartenschreiber(db).KarteAnlegen(probekontext(t), setID); fehler != nil {
		t.Fatalf("Karte anlegen: %v", fehler)
	}

	karten, lesefehler := storage.NeuerSetLeser(db).Karten(probekontext(t), setID)
	if lesefehler != nil {
		t.Fatalf("Karten: %v", lesefehler)
	}

	if len(karten) != 1 {
		t.Fatalf("%d Karten", len(karten))
	}

	if karten[0].SetID != setID {
		t.Errorf("die Karte trägt das Set %d — erwartet war %d", karten[0].SetID, setID)
	}
}
