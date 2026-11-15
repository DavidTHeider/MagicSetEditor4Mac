package storage_test

// **Was abgeschlossen ist, bleibt abgeschlossen** — TRADE-03.AK-03
// (#1717).
//
// # Was hier gemessen wird
//
// **Dass Beleg, Journal, Fremdbeleg und Protokoll sich nicht still ändern
// lassen** — auch nicht mit SQL an der Anwendung vorbei —, und dass die zwei
// Wege, die es geben muss, weiter gehen: die Gegenzeichnung setzt
// `beidseitig`, und ein gelöschtes Profil nimmt alles mit, was ihm gehörte.

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

func TestEinAbgeschlossenerHandelIstNichtZuUeberschreiben(t *testing.T) {
	db := frischeDatenbank(t)
	probe := abschlussprobe{
		bestand: storage.NeuerSammlungsspeicher(db), zusagen: storage.NeueReservierungsablage(db),
		buch: storage.NeueAbschlussablage(db),
	}
	ctx := probekontext(t)

	if fehler := probe.buch.Abschliessen(ctx, einAbschluss(probe.vierStueck(t, true))); fehler != nil {
		t.Fatalf("abschließen: %v", fehler)
	}

	for name, anweisung := range map[string]string{
		"Beleg umschreiben":      `UPDATE handelsbelege SET unterschrift = 'andere'`,
		"Beleg-Abdruck tauschen": `UPDATE handelsbelege SET beidseitig = 1, abdruck = 'x'`,
		"Beleg löschen":          `DELETE FROM handelsbelege`,
		"Journal umschreiben":    `UPDATE handelsjournal SET menge = 99`,
		"Journal löschen":        `DELETE FROM handelsjournal`,
		"Protokoll umschreiben":  `UPDATE handelsprotokoll SET nach = 'beidseitigVerbucht'`,
		"Protokoll löschen":      `DELETE FROM handelsprotokoll`,
	} {
		if _, fehler := db.ExecContext(ctx, anweisung); fehler == nil {
			t.Errorf("%s ging durch", name)
		}
	}

	// **Die Gegenzeichnung geht weiter** — sie setzt `beidseitig` und legt
	// den fremden Beleg daneben.
	if fehler := probe.buch.Gegenzeichnen(ctx, domain.Fremdbeleg{
		Vorgang: einHandel, Von: "bdu", Abdruck: "abdruck-drueben", Unterschrift: "u-drueben",
	}); fehler != nil {
		t.Fatalf("gegenzeichnen: %v", fehler)
	}

	// **Jetzt gibt es einen Fremdbeleg** — und auch er bleibt, wie er kam.
	// Vorher träfe dasselbe `UPDATE` keine Zeile, und kein Auslöser feuerte.
	if _, fehler := db.ExecContext(ctx, `UPDATE handelsfremdbelege SET unterschrift = 'andere'`); fehler == nil {
		t.Error("Fremdbeleg umschreiben ging durch")
	}

	if _, fehler := db.ExecContext(ctx, `DELETE FROM handelsfremdbelege`); fehler == nil {
		t.Error("Fremdbeleg löschen ging durch")
	}

	// **Und ein einmal beidseitiger Beleg wird es nicht wieder los.**
	if _, fehler := db.ExecContext(ctx, `UPDATE handelsbelege SET beidseitig = 0`); fehler == nil {
		t.Error("beidseitig ließ sich zurücknehmen")
	}

	// **Ein zweiter, abweichender Fremdbeleg ist ein Konflikt** — derselbe
	// ist es nicht.
	if fehler := probe.buch.Gegenzeichnen(ctx, domain.Fremdbeleg{
		Vorgang: einHandel, Von: "bdu", Abdruck: "abdruck-drueben", Unterschrift: "u-drueben",
	}); fehler != nil {
		t.Errorf("derselbe Fremdbeleg zweimal: %v", fehler)
	}

	if fehler := probe.buch.Gegenzeichnen(ctx, domain.Fremdbeleg{
		Vorgang: einHandel, Von: "bdu", Abdruck: "ein-anderer", Unterschrift: "u-anders",
	}); !errors.Is(fehler, domain.ErrConflict) {
		t.Errorf("ein abweichender Fremdbeleg: %v", fehler)
	}
}

func TestEinGeloeschtesProfilNimmtSeineBelegeMit(t *testing.T) {
	db := frischeDatenbank(t)
	zweit := zweitesProfil(t, storage.NeueNutzerablage(db), "Tauscher")
	ctx := imProfil(t, zweit)

	probe := abschlussprobe{
		bestand: storage.NeuerSammlungsspeicher(db), zusagen: storage.NeueReservierungsablage(db),
		buch: storage.NeueAbschlussablage(db),
	}

	exemplar := &domain.Exemplar{
		Printing: "bolt-m10", Menge: 4, Zustand: domain.ZustandNearMint, Sprache: "de",
	}

	kennung, fehler := probe.bestand.Anlegen(ctx, exemplar)
	if fehler != nil {
		t.Fatal(fehler)
	}

	exemplar.ID = kennung

	abschluss := einAbschluss(exemplar)
	abschluss.Beleg.Zeilen = abschluss.Beleg.Zeilen[1:]

	if fehler := probe.buch.Abschliessen(ctx, abschluss); fehler != nil {
		t.Fatalf("abschließen: %v", fehler)
	}

	// **Der eine Weg, auf dem ein Beleg gehen darf**: mit seinem Profil.
	if _, fehler := db.ExecContext(ctx, `DELETE FROM profile WHERE id = ?`, zweit); fehler != nil {
		t.Fatalf("das Profil ließ sich nicht löschen: %v", fehler)
	}

	for _, tafel := range []string{"handelsbelege", "handelsjournal", "handelsprotokoll", "handelsrevisionen"} {
		var zahl int
		if fehler := db.QueryRowContext(ctx,
			`SELECT count(*) FROM `+tafel+` WHERE profil_id = ?`, zweit).Scan(&zahl); fehler != nil || zahl != 0 {
			t.Errorf("%s: %d übrig (%v)", tafel, zahl, fehler)
		}
	}
}

func TestDasProtokollHaeltJedenZustandswechsel(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueHandelsablage(db)
	ctx := probekontext(t)

	angebot := domain.Handelsangebot{
		Vertragsfassung: 1, Vorgang: "v-protokoll", Revision: 1,
		Art: domain.HandelTausch, Initiator: "aich", Parteien: []string{"aich", "bdu"},
		Zustand: domain.HandelEntwurf,
	}

	schreibe := func(neu domain.Handelsangebot) {
		t.Helper()

		if fehler := ablage.Schreiben(ctx, neu); fehler != nil {
			t.Fatalf("schreiben: %v", fehler)
		}
	}

	schreibe(angebot)
	angebot.Zustand = domain.HandelAngeboten
	schreibe(angebot)
	schreibe(angebot) // **dieselbe Zeile noch einmal** — kein Wechsel, keine Zeile

	angebot.Revision, angebot.Eltern, angebot.Zustand = 2, "r1", domain.HandelVerhandlung
	schreibe(angebot)

	protokoll, fehler := ablage.Protokoll(ctx, "v-protokoll")
	if fehler != nil {
		t.Fatalf("Protokoll: %v", fehler)
	}

	erwartet := []domain.Handelsprotokollzeile{
		{Revision: 1, Von: "", Nach: domain.HandelEntwurf},
		{Revision: 1, Von: domain.HandelEntwurf, Nach: domain.HandelAngeboten},
		{Revision: 2, Von: "", Nach: domain.HandelVerhandlung},
	}

	if len(protokoll) != len(erwartet) {
		t.Fatalf("Protokoll: %+v", protokoll)
	}

	for i, zeile := range protokoll {
		if zeile.Revision != erwartet[i].Revision || zeile.Von != erwartet[i].Von ||
			zeile.Nach != erwartet[i].Nach || zeile.Zeit.IsZero() {
			t.Errorf("Zeile %d: %+v, erwartet %+v", i, zeile, erwartet[i])
		}
	}
}
