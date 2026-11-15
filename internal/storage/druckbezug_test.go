package storage_test

// **Jede der fünf Nähte trägt den Druck als Zeile** — Werkstück A2 von
// Phase 5.3 (#1441).
//
// Ein `printing` beim Schreiben legt den Druck an oder findet ihn wieder;
// `druck_id` steht danach neben dem Alias, und jedes Lesen bringt es zurück.
// Eine Zeile mit Druck **ohne** Alias — die vorläufige — wird ebenso
// abgelegt und kommt ohne Alias zurück.
//
// **Kein `t.Parallel()`**, aus dem Grund, der in `sammlung_test.go` steht.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// vorlaeufigerDruck legt einen Druck ohne jede Kennung an und gibt seine Zeile.
func vorlaeufigerDruck(t *testing.T, drucke *storage.Drucke) int64 {
	t.Helper()

	druck := &domain.Druck{Name: "Eigenbau", Setcode: "eig", Nummer: "1"}
	if err := drucke.Sichern(probekontext(t), druck); err != nil {
		t.Fatalf("Druck sichern: %v", err)
	}

	return druck.ID
}

// druckMitAlias sagt, welcher Druck den Scryfall-Alias trägt — 0 für keinen.
func druckMitAlias(t *testing.T, drucke *storage.Drucke, printing string) int64 {
	t.Helper()

	druck, err := drucke.NachScryfall(probekontext(t), printing)
	if err != nil {
		t.Fatalf("NachScryfall: %v", err)
	}

	if druck == nil {
		return 0
	}

	return druck.ID
}

func TestEinExemplarMitAliasBekommtSeinenDruck(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)
	drucke := storage.NeueDrucke(db)

	erstes := probeexemplar()
	erstes.Name, erstes.Setcode, erstes.Nummer = "Lightning Bolt", "lea", "161"

	if _, err := speicher.Anlegen(probekontext(t), erstes); err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	zweites := probeexemplar()
	zweites.Zustand = domain.ZustandPlayed

	if _, err := speicher.Anlegen(probekontext(t), zweites); err != nil {
		t.Fatalf("zweites Anlegen: %v", err)
	}

	druckID := druckMitAlias(t, drucke, probedruck)
	if druckID == 0 {
		t.Fatal("der Alias hat keinen Druck bekommen")
	}

	druck, err := drucke.Fuer(probekontext(t), druckID)
	if err != nil {
		t.Fatalf("Fuer: %v", err)
	}

	if druck.Name != "Lightning Bolt" || druck.Setcode != "lea" || druck.Nummer != "161" {
		t.Errorf("der Druck trägt nicht den Rückfall der ersten Zeile: %+v", druck)
	}

	liste, err := speicher.Liste(probekontext(t))
	if err != nil {
		t.Fatalf("Liste: %v", err)
	}

	for _, exemplar := range liste {
		if exemplar.DruckID != druckID {
			t.Errorf("Exemplar %d hat DruckID %d, erwartet %d", exemplar.ID, exemplar.DruckID, druckID)
		}
	}
}

func TestEinVorlaeufigesExemplarKommtOhneAliasZurueck(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerSammlungsspeicher(db)
	druckID := vorlaeufigerDruck(t, storage.NeueDrucke(db))

	exemplar := &domain.Exemplar{DruckID: druckID, Menge: 2, Zustand: domain.ZustandNearMint, Sprache: "de"}

	id, err := speicher.Anlegen(probekontext(t), exemplar)
	if err != nil {
		t.Fatalf("Anlegen: %v", err)
	}

	nachDruck, err := speicher.NachDruck(probekontext(t), druckID)
	if err != nil {
		t.Fatalf("NachDruck: %v", err)
	}

	if len(nachDruck) != 1 || nachDruck[0].ID != id {
		t.Fatalf("NachDruck gibt %+v", nachDruck)
	}

	if nachDruck[0].Printing != "" || nachDruck[0].DruckID != druckID || !nachDruck[0].Vorlaeufig() {
		t.Errorf("die Zeile ist nicht vorläufig: %+v", nachDruck[0])
	}

	// **Ändern lässt den Druck stehen** — auch ohne Alias.
	nachDruck[0].Menge = 5
	if fehler := speicher.Aendern(probekontext(t), nachDruck[0]); fehler != nil {
		t.Fatalf("Aendern: %v", fehler)
	}

	wieder, err := speicher.NachDruck(probekontext(t), druckID)
	if err != nil {
		t.Fatalf("NachDruck: %v", err)
	}

	if len(wieder) != 1 || wieder[0].Menge != 5 || wieder[0].DruckID != druckID {
		t.Errorf("nach dem Ändern: %+v", wieder)
	}

	// **Und ein Alias beim Ändern zieht die Zeile auf den Scryfall-Druck.**
	wieder[0].Printing = probedruck
	if fehler := speicher.Aendern(probekontext(t), wieder[0]); fehler != nil {
		t.Fatalf("Aendern mit Alias: %v", fehler)
	}

	gezogen, err := speicher.Fuer(probekontext(t), probedruck)
	if err != nil {
		t.Fatalf("Fuer: %v", err)
	}

	if len(gezogen) != 1 || gezogen[0].DruckID == druckID || gezogen[0].DruckID == 0 {
		t.Errorf("der Alias hat die Zeile nicht umgehängt: %+v", gezogen)
	}
}

func TestEinVerweisTraegtSeinenDruck(t *testing.T) {
	db, ablage, setID, ersteID, zweiteID := verweisstand(t)
	drucke := storage.NeueDrucke(db)

	if err := ablage.Setzen(probekontext(t), ersteID, probeverweis()); err != nil {
		t.Fatalf("Setzen: %v", err)
	}

	druckID := druckMitAlias(t, drucke, probeverweis().Printing)
	if druckID == 0 {
		t.Fatal("der Verweis hat keinen Druck angelegt")
	}

	gelesen, err := ablage.Fuer(probekontext(t), ersteID)
	if err != nil {
		t.Fatalf("Fuer: %v", err)
	}

	if gelesen.DruckID != druckID {
		t.Errorf("Fuer gibt DruckID %d, erwartet %d", gelesen.DruckID, druckID)
	}

	// **Ein vorläufiger Verweis** — Druck ohne Alias.
	vorlaeufig := vorlaeufigerDruck(t, drucke)
	if fehler := ablage.Setzen(probekontext(t), zweiteID, domain.Kartenverweis{
		DruckID: vorlaeufig, Name: "Eigenbau", Gesetzt: probeverweis().Gesetzt,
	}); fehler != nil {
		t.Fatalf("Setzen ohne Alias: %v", fehler)
	}

	jeSet, err := ablage.JeSet(probekontext(t), setID)
	if err != nil {
		t.Fatalf("JeSet: %v", err)
	}

	if jeSet[ersteID].DruckID != druckID || jeSet[zweiteID].DruckID != vorlaeufig {
		t.Errorf("JeSet: %+v", jeSet)
	}

	if !jeSet[zweiteID].Vorlaeufig() || jeSet[ersteID].Vorlaeufig() {
		t.Errorf("vorläufig ist falsch verteilt: %+v", jeSet)
	}
}

func TestEinDeckplatzTraegtSeinenDruck(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueDeckablage(db)
	drucke := storage.NeueDrucke(db)
	deck := deckAnlegen(t, ablage)

	const alias = "a1b2c3d4-0000-0000-0000-000000000000"

	if _, err := ablage.PlatzSetzen(probekontext(t), deck.ID, domain.Deckplatz{
		Printing: alias, Menge: 4, Bereich: domain.BereichHauptdeck,
	}); err != nil {
		t.Fatalf("legen: %v", err)
	}

	vorlaeufig := vorlaeufigerDruck(t, drucke)

	gelegt, err := ablage.PlatzSetzen(probekontext(t), deck.ID, domain.Deckplatz{
		DruckID: vorlaeufig, Menge: 1, Bereich: domain.BereichHauptdeck,
	})
	if err != nil {
		t.Fatalf("legen ohne Alias: %v", err)
	}

	// **Derselbe vorläufige Druck noch einmal ist eine höhere Menge.**
	nochmal, err := ablage.PlatzSetzen(probekontext(t), deck.ID, domain.Deckplatz{
		DruckID: vorlaeufig, Menge: 2, Bereich: domain.BereichHauptdeck,
	})
	if err != nil {
		t.Fatalf("nochmal legen: %v", err)
	}

	if nochmal.ID != gelegt.ID {
		t.Errorf("der zweite Platz ist ein anderer: %d gegen %d", nochmal.ID, gelegt.ID)
	}

	gefunden := deckLesen(t, ablage, deck.ID)
	if len(gefunden.Plaetze) != 2 {
		t.Fatalf("%d Plätze", len(gefunden.Plaetze))
	}

	if gefunden.Plaetze[0].DruckID != druckMitAlias(t, drucke, alias) || gefunden.Plaetze[0].DruckID == 0 {
		t.Errorf("der Platz mit Alias: %+v", gefunden.Plaetze[0])
	}

	if gefunden.Plaetze[1].DruckID != vorlaeufig || gefunden.Plaetze[1].Printing != "" ||
		gefunden.Plaetze[1].Menge != 2 {
		t.Errorf("der vorläufige Platz: %+v", gefunden.Plaetze[1])
	}
}

func TestEinPreisstandTraegtSeinenDruck(t *testing.T) {
	db := frischeDatenbank(t)
	speicher := storage.NeuerPreisspeicher(db)

	if err := speicher.Eintragen(probekontext(t), probestand(20, 417)); err != nil {
		t.Fatalf("Eintragen: %v", err)
	}

	gelesen, err := speicher.Neueste(probekontext(t), probedruck)
	if err != nil {
		t.Fatalf("Neueste: %v", err)
	}

	druckID := druckMitAlias(t, storage.NeueDrucke(db), probedruck)

	if len(gelesen) != 1 || druckID == 0 || gelesen[0].DruckID != druckID {
		t.Errorf("Neueste: %+v, Druck %d", gelesen, druckID)
	}
}

func TestEinDeckabzugTraegtSeinenDruck(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueTurnierablage(db)

	turnier, err := ablage.Anlegen(probekontext(t), domain.Turnier{Name: "Küchentisch", Spiel: "magic"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if _, fehler := ablage.Anmelden(probekontext(t), turnier.ID, domain.Teilnehmer{
		Name: "Ada",
		Deckabzug: []domain.Abzugsplatz{
			{Bezeichnung: "Lightning Bolt", Printing: "a1b2c3d4", Menge: 4, Bereich: domain.BereichHauptdeck, Stelle: 1},
			{Bezeichnung: "Eigenbau", Menge: 1, Bereich: domain.BereichHauptdeck, Stelle: 2},
		},
	}); fehler != nil {
		t.Fatalf("anmelden: %v", fehler)
	}

	gelesen, err := ablage.Turnier(probekontext(t), turnier.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	abzug := gelesen.Teilnehmer[0].Deckabzug
	druckID := druckMitAlias(t, storage.NeueDrucke(db), "a1b2c3d4")

	if len(abzug) != 2 || druckID == 0 || abzug[0].DruckID != druckID || abzug[1].DruckID != 0 {
		t.Errorf("Abzug %+v, Druck %d", abzug, druckID)
	}
}
