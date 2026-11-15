package app

// **Was für einen Handel festgehalten wird** —
// #1643.
//
// # Wogegen hier geprüft wird
//
// **Dass der Weg wirklich etwas festhält.** `storage.Festhalten` stand
// gebaut, geprüft und ohne Aufrufer da; dieses Programm konnte kein einziges
// Exemplar für einen Handel binden.

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// einHandelZumFesthalten legt Vorgang und Besitzzeile an.
//
// **Mit eigener Identität** — ein Vorgang trägt die Peerkennung dieses
// Profils als Partei, und ohne sie kommt er gar nicht zustande.
func einHandelZumFesthalten(
	t *testing.T,
) (anwendung *App, exemplarID int64) {
	t.Helper()

	anwendung = PeerflaecheUeber(
		bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})
	anwendung.profil.Store(1)

	ctx := riegelkontext(t)

	ausweis := anwendung.PeeridentitaetSicherstellen(ctx)
	if ausweis.Fehler != nil {
		t.Fatalf("Identität anlegen: %+v", ausweis.Fehler)
	}

	eroeffnet := anwendung.HandelEroeffnen(ctx, Handelswunsch{
		Vorgang: "v-halt", Art: string(domain.HandelTausch),
		Gegenueber: []string{"bdort"},
		Positionen: []HandelspositionDTO{{
			Von: ausweis.Kennung, Printing: "a1b2", Menge: 2,
			Guete: "nearMint", Sprache: "de", Finish: "nonfoil",
		}},
	})
	if eroeffnet.Fehler != nil {
		t.Fatalf("eröffnen: %+v", eroeffnet.Fehler)
	}

	// **Die Besitzzeile entsteht am Bestand und nicht über die Bindung** —
	// dieser Lauf misst die Zusagen und nicht die Sammlungsnaht.
	exemplarID, fehler := storage.NeuerSammlungsspeicher(anwendung.db).Anlegen(
		ctx, &domain.Exemplar{
			Printing: "a1b2", Menge: 3,
			Zustand: domain.ZustandNearMint, Sprache: "de",
		})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	return anwendung, exemplarID
}

// **Eine benannte Besitzzeile wird festgehalten und kommt zurück.**
func TestEineBesitzzeileWirdFuerEinenHandelFestgehalten(t *testing.T) {
	t.Parallel()

	anwendung, exemplarID := einHandelZumFesthalten(t)

	antwort := anwendung.HandelFesthalten(
		riegelkontext(t), "v-halt", exemplarID, 2, 1)
	if antwort.Fehler != nil {
		t.Fatalf("Festhalten: %+v", antwort.Fehler)
	}

	switch {
	case len(antwort.Zusagen) != 1:
		t.Fatalf("festgehalten wurden %d", len(antwort.Zusagen))
	case antwort.Zusagen[0].ExemplarID != exemplarID:
		t.Errorf("die Zusage nennt %d", antwort.Zusagen[0].ExemplarID)
	case antwort.Zusagen[0].Menge != 2:
		t.Errorf("festgehalten sind %d Stück", antwort.Zusagen[0].Menge)
	case antwort.Zusagen[0].Uebergeben:
		t.Error("sie gilt schon als übergeben")
	}
}

// **Mehr als dasteht, wird abgewiesen — ohne Bestandswirkung.**
func TestMehrAlsDastehtWirdNichtFestgehalten(t *testing.T) {
	t.Parallel()

	anwendung, exemplarID := einHandelZumFesthalten(t)

	antwort := anwendung.HandelFesthalten(
		riegelkontext(t), "v-halt", exemplarID, 9, 1)
	if antwort.Fehler == nil {
		t.Fatal("neun von drei Stück wurden festgehalten")
	}

	gelesen := anwendung.Handelszusagen(riegelkontext(t), "v-halt")
	if gelesen.Fehler != nil {
		t.Fatalf("Handelszusagen: %+v", gelesen.Fehler)
	}

	if len(gelesen.Zusagen) != 0 {
		t.Errorf("festgehalten sind %+v", gelesen.Zusagen)
	}
}

// **Ein unbekannter Vorgang hält nichts fest.**
func TestEinUnbekannterVorgangHaeltNichtsFest(t *testing.T) {
	t.Parallel()

	anwendung, exemplarID := einHandelZumFesthalten(t)

	if antwort := anwendung.HandelFesthalten(
		riegelkontext(t), "v-gibtesnicht", exemplarID, 1, 1,
	); antwort.Fehler == nil {
		t.Fatal("der unbekannte Vorgang hielt etwas fest")
	}
}

// **Und eine leere Liste ist eine leere Liste und kein `null`.**
func TestDieZusagenKommenAlsListeUndNichtAlsNull(t *testing.T) {
	t.Parallel()

	anwendung, _ := einHandelZumFesthalten(t)

	antwort := anwendung.Handelszusagen(riegelkontext(t), "v-leer")
	if antwort.Fehler != nil {
		t.Fatalf("Handelszusagen: %+v", antwort.Fehler)
	}

	if antwort.Zusagen == nil {
		t.Error("die Liste ist null")
	}
}

// **Ohne die Wirkung wird nichts festgehalten.**
func TestOhneDieWirkungWirdNichtsFestgehalten(t *testing.T) {
	t.Parallel()

	anwendung, exemplarID := einHandelZumFesthalten(t)

	wache := &abweisendeWache{}
	anwendung.MitWache(wache)

	antwort := anwendung.HandelFesthalten(
		riegelkontext(t), "v-halt", exemplarID, 1, 1)

	switch {
	case antwort.Fehler == nil:
		t.Fatal("die Abweisung kam nicht an")
	case wache.gefragt != domain.WirkungSchreiben:
		t.Errorf("gefragt wurde nach %q", wache.gefragt)
	case antwort.Zusagen == nil:
		t.Error("die Liste ist null statt leer")
	}
}

// **Ein Bestand, der die Zusagen nicht hergibt, meldet sich.**
//
// **Und gibt keine leere Liste zurück**: „nichts festgehalten" und „ich weiß
// nicht, was festgehalten ist" sind zwei Dinge (§8).
func TestEinUnlesbarerZusagenbestandMeldetSich(t *testing.T) {
	t.Parallel()

	anwendung, _ := einHandelZumFesthalten(t)

	if _, fehler := anwendung.db.ExecContext(
		riegelkontext(t), "DROP TABLE reservierungen"); fehler != nil {
		t.Fatalf("die Zusagen wegnehmen: %v", fehler)
	}

	if antwort := anwendung.Handelszusagen(
		riegelkontext(t), "v-halt"); antwort.Fehler == nil {
		t.Error("der fehlende Bestand kam als leere Liste")
	}
}

// **Eine Zusage mit Frist nennt sie.**
//
// **Geschrieben wird sie hier nicht** — dieses Fenster bindet an den Handel
// und nicht an eine Uhr. Gelesen wird sie trotzdem: was im Bestand steht,
// kam von irgendwoher, und eine Fläche, die es verschwiege, zeigte weniger
// als da ist (§8).
func TestEineZusageMitFristNenntSie(t *testing.T) {
	t.Parallel()

	anwendung, exemplarID := einHandelZumFesthalten(t)

	ctx := riegelkontext(t)

	if fehler := storage.NeueReservierungsablage(anwendung.db).Festhalten(
		ctx, domain.Reservierung{
			Vorgang: "v-halt", Revision: 1, ExemplarID: exemplarID,
			Menge: 1, Fassung: 1, Frist: time.Now().Add(time.Hour),
		}, domain.Zusagevertrag{{
			Printing: "a1b2", Menge: 99, Guete: "nearMint", Sprache: "de", Finish: "nonfoil",
		}}); fehler != nil {
		t.Fatalf("festhalten: %v", fehler)
	}

	antwort := anwendung.Handelszusagen(ctx, "v-halt")
	if antwort.Fehler != nil {
		t.Fatalf("Handelszusagen: %+v", antwort.Fehler)
	}

	if len(antwort.Zusagen) != 1 || antwort.Zusagen[0].Frist == "" {
		t.Errorf("die Zusage ist %+v", antwort.Zusagen)
	}
}
