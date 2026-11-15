package storage_test

// **Was eine Rücksetzung anfassen würde** — Werkstück O4 von Phase 5.2
// (#1428).

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein leeres Profil hat keine Gruppe** — und das ist die Gegenprobe zu
// allem darunter: ein Umfang, der immer etwas nennt, nennt nichts (§10).
func TestEinLeeresProfilHatKeineGruppe(t *testing.T) {
	db := frischeDatenbank(t)

	betroffen, fehler := storage.Profilumfang(probekontext(t), db, 1)
	if fehler != nil {
		t.Fatalf("der Umfang: %v", fehler)
	}

	if len(betroffen) != 0 {
		t.Errorf("ein leeres Profil nennt %d Gruppen: %+v", len(betroffen), betroffen)
	}
}

// **Ein Token wird benannt** — die Abnahme verlangt es: „Profillöschung
// benennt und widerruft seine Token."
func TestEinTokenWirdBenanntUndNichtNurGezaehlt(t *testing.T) {
	db := frischeDatenbank(t)

	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO zugriffstoken (profil_id, bezeichnung, gehuetet, scope, angelegt)
		 VALUES (1, 'Sicherungsskript', 'abc', 'lesen', ?)`,
		storage.Zeitstempel(time.Now())); fehler != nil {
		t.Fatalf("das Token legen: %v", fehler)
	}

	betroffen, fehler := storage.Profilumfang(probekontext(t), db, 1)
	if fehler != nil {
		t.Fatalf("der Umfang: %v", fehler)
	}

	nachGruppe := map[string]domain.Betroffenes{}
	for _, gruppe := range betroffen {
		nachGruppe[gruppe.Was] = gruppe
	}

	token, da := nachGruppe[domain.BetroffenToken]
	if !da {
		t.Fatalf("das Token steht in keiner Gruppe: %+v", betroffen)
	}

	if token.Anzahl != 1 || len(token.Namen) != 1 || token.Namen[0] != "Sicherungsskript" {
		t.Errorf("die Gruppe ist %+v", token)
	}
}

// **Exemplare zählen Stücke und nicht Zeilen** — dieselbe Entscheidung wie
// beim Bestandsstand (O7).
func TestExemplareZaehlenStueckeUndNichtZeilen(t *testing.T) {
	db := frischeDatenbank(t)

	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO sammlung (profil_id, printing, menge, zustand, sprache, finish, name)
		 VALUES (1, 'scryfall:abc', 4, 'nearMint', 'en', 'nonfoil', 'Lightning Bolt')`); fehler != nil {
		t.Fatalf("das Exemplar legen: %v", fehler)
	}

	betroffen, fehler := storage.Profilumfang(probekontext(t), db, 1)
	if fehler != nil {
		t.Fatalf("der Umfang: %v", fehler)
	}

	for _, gruppe := range betroffen {
		if gruppe.Was == domain.BetroffenExemplare {
			if gruppe.Anzahl != 4 {
				t.Errorf("es sind %d Stück, erwartet waren 4", gruppe.Anzahl)
			}

			if len(gruppe.Namen) != 1 {
				t.Errorf("die Namen sind %v", gruppe.Namen)
			}

			return
		}
	}

	t.Errorf("die Exemplare stehen in keiner Gruppe: %+v", betroffen)
}

// **Eine Karte hat hier keinen Namen** — die Zahl sagt trotzdem, wie viele es
// sind. Ein leerer Eintrag in der Liste wäre schlimmer als keiner.
func TestKartenWerdenGezaehltUndNichtBenannt(t *testing.T) {
	db := frischeDatenbank(t)

	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO sets (id, profil_id, name, spiel, stil, geoeffnet)
		 VALUES (1, 1, 'Probeset', 'magic', 'm15', ?)`,
		storage.Zeitstempel(time.Now())); fehler != nil {
		t.Fatalf("das Set legen: %v", fehler)
	}

	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO cards (set_id, reihenfolge) VALUES (1, 0)`); fehler != nil {
		t.Fatalf("die Karte legen: %v", fehler)
	}

	betroffen, fehler := storage.Profilumfang(probekontext(t), db, 1)
	if fehler != nil {
		t.Fatalf("der Umfang: %v", fehler)
	}

	nachGruppe := map[string]domain.Betroffenes{}
	for _, gruppe := range betroffen {
		nachGruppe[gruppe.Was] = gruppe
	}

	if nachGruppe[domain.BetroffenSets].Anzahl != 1 ||
		len(nachGruppe[domain.BetroffenSets].Namen) != 1 {
		t.Errorf("die Sets sind %+v", nachGruppe[domain.BetroffenSets])
	}

	karten := nachGruppe[domain.BetroffenKarten]
	if karten.Anzahl != 1 {
		t.Errorf("es ist %d Karte gezählt", karten.Anzahl)
	}

	if len(karten.Namen) != 0 {
		t.Errorf("die Karten sind benannt: %v", karten.Namen)
	}
}

// **Eine offene Rückkehr sperrt** — das ist die Stelle, an der T2, T5 und T6
// andocken (Vermerk 0088).
func TestEineOffeneRueckkehrWirdGezaehlt(t *testing.T) {
	db := frischeDatenbank(t)

	offen, fehler := storage.OffeneRueckkehrmengen(probekontext(t), db, 1)
	if fehler != nil {
		t.Fatalf("die offenen Rückkehren: %v", fehler)
	}

	if offen != 0 {
		t.Fatalf("ein frischer Bestand hat %d offene Mengen", offen)
	}

	if _, vermerkfehler := storage.NeuesWiederherstellungsbuch(db).Vermerken(
		probekontext(t), domain.Wiederherstellungsepoche{
			Zurueckgespielt: time.Now(), Datenstand: time.Now(),
			Quelle: "probe.msesicherung", Klaerung: domain.KlaerungOffen,
			Mengen: []domain.Sammlungsmenge{
				{Profil: 1, Schluessel: "scryfall:abc", Menge: 3},
			},
		}); vermerkfehler != nil {
		t.Fatalf("die Epoche vermerken: %v", vermerkfehler)
	}

	nachher, nachfehler := storage.OffeneRueckkehrmengen(probekontext(t), db, 1)
	if nachfehler != nil {
		t.Fatalf("die offenen Rückkehren: %v", nachfehler)
	}

	if nachher != 1 {
		t.Errorf("es sind %d offene Mengen, erwartet war eine", nachher)
	}

	// **Und ein anderes Profil ist nicht betroffen.**
	fremd, fremdfehler := storage.OffeneRueckkehrmengen(probekontext(t), db, 99)
	if fremdfehler != nil {
		t.Fatalf("die offenen Rückkehren: %v", fremdfehler)
	}

	if fremd != 0 {
		t.Errorf("das fremde Profil hat %d offene Mengen", fremd)
	}
}

// **Der letzte Mensch lässt sich nicht löschen** — und die Vorschau sagt es
// vorher, statt es beim Bestätigen zu melden.
func TestDieNutzerzahlSagtObEsDerLetzteIst(t *testing.T) {
	db := frischeDatenbank(t)

	anzahl, fehler := storage.Nutzerzahl(probekontext(t), db)
	if fehler != nil {
		t.Fatalf("die Nutzerzahl: %v", fehler)
	}

	// **Die Wanderung `0034` legt genau einen an** — die Eins ist keine
	// Annahme, sondern das Schema.
	if anzahl != 1 {
		t.Errorf("es sind %d Nutzer, erwartet war einer", anzahl)
	}
}

// **Auf einer Datenbank, die zu ist, meldet sich jeder Weg** — statt eines
// leeren Umfangs, der wie „da ist nichts" aussähe.
func TestAufEinerGeschlossenenDatenbankMeldetSichDerUmfang(t *testing.T) {
	db := zugemachteDatenbank(t)

	if _, fehler := storage.Profilumfang(probekontext(t), db, 1); fehler == nil {
		t.Error("der Umfang meldete nichts")
	}

	if _, fehler := storage.OffeneRueckkehrmengen(probekontext(t), db, 1); fehler == nil {
		t.Error("die offenen Rückkehren meldeten nichts")
	}

	if _, fehler := storage.Nutzerzahl(probekontext(t), db); fehler == nil {
		t.Error("die Nutzerzahl meldete nichts")
	}
}

// **Zurückgesetzt heißt gelöscht — und damit geerbt.**
//
// Die Abnahme sagt es: „Profilreset stellt Vererbung her." Ein geschriebenes
// `erben` wäre eine Wahl, die wie keine aussieht.
func TestZurueckgesetztHeisstGeloeschtUndDamitGeerbt(t *testing.T) {
	db := frischeDatenbank(t)

	einstellungen := storage.NeueEinstellungen(db)

	for schluessel, wert := range map[string]string{
		domain.EinstellungErscheinungsbild: "dunkel",
		domain.EinstellungSprache:          "en",
	} {
		if fehler := einstellungen.Setzen(probekontext(t), schluessel, wert); fehler != nil {
			t.Fatalf("%s setzen: %v", schluessel, fehler)
		}
	}

	// **Nur die Layoutschlüssel** — die Sprache bleibt stehen.
	betroffen, fehler := storage.EinstellungenZuruecksetzen(
		probekontext(t), db, domain.Layouteinstellungen())
	if fehler != nil {
		t.Fatalf("zurücksetzen: %v", fehler)
	}

	if betroffen != 1 {
		t.Errorf("es wurden %d zurückgesetzt, erwartet war eine", betroffen)
	}

	stand, lesefehler := einstellungen.Alle(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if _, steht := stand[domain.EinstellungErscheinungsbild]; steht {
		t.Error("das Erscheinungsbild steht noch als Wahl da")
	}

	if stand[domain.EinstellungSprache] != "en" {
		t.Errorf("die Sprache ist %q — sie gehört nicht zum Layout",
			stand[domain.EinstellungSprache])
	}
}

// **Ohne Schlüssel gilt es für alle** — das ist die Profilrücksetzung.
func TestOhneSchluesselWirdAllesZurueckgesetzt(t *testing.T) {
	db := frischeDatenbank(t)

	einstellungen := storage.NeueEinstellungen(db)

	for _, schluessel := range []string{
		domain.EinstellungErscheinungsbild, domain.EinstellungSprache,
	} {
		if fehler := einstellungen.Setzen(probekontext(t), schluessel, "dunkel"); fehler != nil {
			// **Ein ungültiger Wert für die Sprache ist hier egal** — geprüft
			// wird das Zurücksetzen und nicht das Setzen.
			continue
		}
	}

	betroffen, fehler := storage.EinstellungenZuruecksetzen(probekontext(t), db, nil)
	if fehler != nil {
		t.Fatalf("zurücksetzen: %v", fehler)
	}

	if betroffen == 0 {
		t.Error("es wurde nichts zurückgesetzt")
	}

	stand, lesefehler := einstellungen.Alle(probekontext(t))
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(stand) != 0 {
		t.Errorf("es steht noch etwas da: %v", stand)
	}
}

// **Die Nutzereinstellungen gehen denselben Weg.**
func TestDieNutzereinstellungenLassenSichZuruecksetzen(t *testing.T) {
	db := frischeDatenbank(t)

	ablage := storage.NeueNutzerablage(db)

	if fehler := ablage.NutzereinstellungSetzen(probekontext(t), 1,
		domain.EinstellungErscheinungsbild, "dunkel"); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	betroffen, fehler := storage.NutzereinstellungenZuruecksetzen(probekontext(t), db, 1)
	if fehler != nil {
		t.Fatalf("zurücksetzen: %v", fehler)
	}

	if betroffen != 1 {
		t.Errorf("es wurden %d zurückgesetzt", betroffen)
	}

	stand, lesefehler := ablage.Nutzereinstellungen(probekontext(t), 1)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(stand) != 0 {
		t.Errorf("es steht noch etwas da: %v", stand)
	}
}

// **Und auf einer geschlossenen Datenbank meldet sich beides.**
func TestDasZuruecksetzenMeldetEineGeschlosseneDatenbank(t *testing.T) {
	db := zugemachteDatenbank(t)

	if _, fehler := storage.EinstellungenZuruecksetzen(
		probekontext(t), db, domain.Layouteinstellungen()); fehler == nil {
		t.Error("das Zurücksetzen meldete nichts")
	}

	if _, fehler := storage.EinstellungenZuruecksetzen(probekontext(t), db, nil); fehler == nil {
		t.Error("das Zurücksetzen ohne Schlüssel meldete nichts")
	}

	if _, fehler := storage.NutzereinstellungenZuruecksetzen(
		probekontext(t), db, 1); fehler == nil {
		t.Error("die Nutzereinstellungen meldeten nichts")
	}
}

// **Ohne Profil wird nichts zurückgesetzt** — dieselbe Zusage wie überall.
func TestOhneProfilWirdNichtsZurueckgesetzt(t *testing.T) {
	db := frischeDatenbank(t)

	if _, fehler := storage.EinstellungenZuruecksetzen(
		t.Context(), db, domain.Layouteinstellungen()); fehler == nil {
		t.Error("ohne Profil wurde zurückgesetzt")
	}
}

// **Der Neuanfang sieht jedes Profil** — Befund 1 des Zweitblicks über Phase
// G (#2282): auch das eines zweiten Menschen, denn es geht mit der Datei
// beiseite.
func TestAlleProfileNenntAuchFremde(t *testing.T) {
	db := frischeDatenbank(t)

	fremder, fehler := storage.NeueNutzerablage(db).NutzerAnlegen(probekontext(t),
		domain.Nutzer{Name: "Fremde"})
	if fehler != nil {
		t.Fatalf("den zweiten Menschen anlegen: %v", fehler)
	}

	profile, lesefehler := storage.AlleProfile(probekontext(t), db)
	if lesefehler != nil {
		t.Fatalf("alle Profile: %v", lesefehler)
	}

	if len(profile) != 2 || profile[1].NutzerID != fremder.ID {
		t.Errorf("die Profile sind %+v — erwartet das eigene und das fremde", profile)
	}
}
