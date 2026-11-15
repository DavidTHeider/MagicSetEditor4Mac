package storage_test

// **Reservierungen im Bestand** — Werkstück `T2` von Phase 5.6
// (#1467).
//
// # Die erste beobachtbare Abnahme des Issues
//
// **„Zwei parallele Handelsvorgänge und ein regulärer Mengen-Schreibweg gegen
// dieselbe SQLite-Datei können insgesamt nie mehr als den Besitz
// reservieren."**
//
// Das ist keine Aussage über eine Funktion, sondern über eine **Datei unter
// gleichzeitigem Zugriff** — und deshalb steht sie hier und nicht in der
// Domäne. Dort steht, **was** hineinpasst; hier steht, dass es zwischen
// Nachsehen und Eintragen niemand ändern kann.

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// dreiStueck legt eine Besitzzeile mit drei Stücken an und gibt sie.
func dreiStueck(t *testing.T, db *storage.Sammlungsspeicher) *domain.Exemplar {
	t.Helper()

	exemplar := &domain.Exemplar{
		Printing: "abc-123", Menge: 3, Zustand: domain.ZustandNearMint,
		Sprache: "de",
	}

	kennung, fehler := db.Anlegen(probekontext(t), exemplar)
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	exemplar.ID = kennung
	exemplar.Fassung = 1

	return exemplar
}

// eineZusage gibt eine Reservierung auf dieses Exemplar.
func eineZusage(
	exemplar *domain.Exemplar, vorgang string, menge int,
) domain.Reservierung {
	return domain.Reservierung{
		Vorgang: vorgang, Revision: 1, ExemplarID: exemplar.ID,
		Menge: menge, Fassung: exemplar.Fassung,
		Frist: time.Now().Add(time.Hour),
	}
}

// **Eine Zusage wird festgehalten und kommt zurück.**
func TestEineZusageWirdFestgehalten(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ctx := probekontext(t)
	exemplar := dreiStueck(t, storage.NeuerSammlungsspeicher(db))
	ablage := storage.NeueReservierungsablage(db)

	if fehler := ablage.Festhalten(
		ctx, eineZusage(exemplar, "h1", 2), ueber(exemplar, 99)); fehler != nil {
		t.Fatalf("festhalten: %v", fehler)
	}

	offene, fehler := ablage.Reservierungen(ctx, exemplar.ID)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(offene) != 1 || offene[0].Menge != 2 || offene[0].Vorgang != "h1" {
		t.Errorf("zurück kam %+v", offene)
	}
}

// **Mehr als der Besitz geht nicht — auch nicht in zwei Zusagen.**
func TestZweiZusagenKoennenDenBesitzNichtUeberschreiten(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ctx := probekontext(t)
	exemplar := dreiStueck(t, storage.NeuerSammlungsspeicher(db))
	ablage := storage.NeueReservierungsablage(db)

	if fehler := ablage.Festhalten(
		ctx, eineZusage(exemplar, "h1", 2), ueber(exemplar, 99)); fehler != nil {
		t.Fatalf("die erste: %v", fehler)
	}

	if fehler := ablage.Festhalten(
		ctx, eineZusage(exemplar, "h2", 2), ueber(exemplar, 99)); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("die zweite ergibt %v", fehler)
	}

	// **Und die passende passt** — sonst prüfte der Lauf nur, dass überhaupt
	// etwas abgewiesen wird.
	if fehler := ablage.Festhalten(
		ctx, eineZusage(exemplar, "h3", 1), ueber(exemplar, 99)); fehler != nil {
		t.Errorf("die dritte: %v", fehler)
	}
}

// **Und auch nicht, wenn beide gleichzeitig kommen.**
//
// # Warum das hier steht und nicht in der Domäne
//
// **Weil zwischen Nachsehen und Eintragen ein Fenster liegt.** Zwei Rufe, die
// beide „es sind noch zwei frei" lesen und beide zwei eintragen, hätten vier
// festgehalten — und die Domäne hätte in beiden Fällen Ja gesagt. Was das
// verhindert, ist die **Transaktion**, und die gibt es nur hier.
//
// # Was ihn hält, und was er deshalb nicht belegt
//
// **Die Reihenfolge kommt vom Treiber und nicht von diesem Code allein**:
// `_txlock=immediate` nimmt die Schreibsperre **beim Beginnen**
// (`datenbank.go`). Die Gegenprobe zeigt es deutlicher als ein grüner Lauf —
// wer die beiden Abfragen aus der Transaktion herausnimmt, bekommt keinen
// überzähligen Eintrag, sondern einen **Stillstand**: der zweite Ruf wartet
// auf eine Sperre, die der erste hält.
//
// **Dieser Lauf unterscheidet also „richtig" von „hängt", nicht „richtig" von
// „zählt falsch"** — und das gehört dazugesagt (§10). Dass die Summe nie über
// den Besitz geht, hält er; **warum** sie es nicht tut, steht zur Hälfte im
// Treiber.
func TestZweiGleichzeitigeZusagenUeberschreitenDenBesitzNicht(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ctx := probekontext(t)
	exemplar := dreiStueck(t, storage.NeuerSammlungsspeicher(db))
	ablage := storage.NeueReservierungsablage(db)

	var warten sync.WaitGroup

	for lauf := range 4 {
		warten.Add(1)

		go func() {
			defer warten.Done()

			_ = ablage.Festhalten(ctx, //nolint:errcheck // der Fehler ist hier der Normalfall
				eineZusage(exemplar, "h"+string(rune('a'+lauf)), 2), ueber(exemplar, 99))
		}()
	}

	warten.Wait()

	offene, fehler := ablage.Reservierungen(ctx, exemplar.ID)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	festgehalten := 0
	for _, zusage := range offene {
		festgehalten += zusage.Menge
	}

	if festgehalten > exemplar.Menge {
		t.Errorf("festgehalten sind %d von %d Stücken: %+v",
			festgehalten, exemplar.Menge, offene)
	}

	if festgehalten == 0 {
		t.Error("gar nichts wurde festgehalten — der Lauf misst nichts")
	}
}

// **Eine Zusage gegen eine überholte Fassung wird abgewiesen.**
//
// **Der Quellfall des Issues**: drei Stück, zwei reserviert, die freie Menge
// um eins geändert — der Fassungskonflikt erzwingt eine neue Prüfung, und der
// alte Ruf schreibt nichts.
func TestEineZusageGegenEineUeberholteFassungWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ctx := probekontext(t)
	speicher := storage.NeuerSammlungsspeicher(db)
	exemplar := dreiStueck(t, speicher)
	ablage := storage.NeueReservierungsablage(db)

	alte := eineZusage(exemplar, "h1", 2)

	// **Jemand ändert die Menge** — und die Fassung steigt.
	geaendert := *exemplar
	geaendert.Menge = 4

	if fehler := speicher.Aendern(ctx, &geaendert); fehler != nil {
		t.Fatalf("ändern: %v", fehler)
	}

	if fehler := ablage.Festhalten(
		ctx, alte, ueber(exemplar, 99)); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("es kam %v", fehler)
	}

	offene, fehler := ablage.Reservierungen(ctx, exemplar.ID)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(offene) != 0 {
		t.Errorf("der abgewiesene Ruf hat %d Zusagen hinterlassen", len(offene))
	}
}

// **Eine Wunschzeile lässt sich nicht festhalten.**
func TestEineWunschzeileLaesstSichNichtFesthalten(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ctx := probekontext(t)
	speicher := storage.NeuerSammlungsspeicher(db)

	wunsch := &domain.Exemplar{
		Printing: "abc-123", Menge: 3, Zustand: domain.ZustandNearMint,
		Sprache: "de", Wunsch: true,
	}

	kennung, fehler := speicher.Anlegen(ctx, wunsch)
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	wunsch.ID, wunsch.Fassung = kennung, 1

	if fehler := storage.NeueReservierungsablage(db).Festhalten(
		ctx, eineZusage(wunsch, "h1", 1), ueber(wunsch, 99)); !errors.Is(fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Freigeben nimmt die Zusagen eines Vorgangs weg — und nur die.**
func TestFreigebenNimmtNurDieEigenenZusagen(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ctx := probekontext(t)
	exemplar := dreiStueck(t, storage.NeuerSammlungsspeicher(db))
	ablage := storage.NeueReservierungsablage(db)

	for _, vorgang := range []string{"h1", "h2"} {
		if fehler := ablage.Festhalten(
			ctx, eineZusage(exemplar, vorgang, 1), ueber(exemplar, 99)); fehler != nil {
			t.Fatalf("%s: %v", vorgang, fehler)
		}
	}

	if fehler := ablage.Freigeben(ctx, "h1"); fehler != nil {
		t.Fatalf("freigeben: %v", fehler)
	}

	offene, fehler := ablage.Reservierungen(ctx, exemplar.ID)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(offene) != 1 || offene[0].Vorgang != "h2" {
		t.Errorf("übrig sind %+v", offene)
	}
}

// **Eine bestätigte Übergabe gibt eine Zusage nicht mehr frei.**
//
// **Auch nicht durch den Ruf, der sonst freigibt** — sonst genügte ein
// verspäteter Aufräumer, um Karten wieder verfügbar zu machen, die längst
// über den Tisch gegangen sind.
func TestEineUebergebeneZusageWirdNichtFreigegeben(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ctx := probekontext(t)
	exemplar := dreiStueck(t, storage.NeuerSammlungsspeicher(db))
	ablage := storage.NeueReservierungsablage(db)

	zusage := eineZusage(exemplar, "h1", 2)
	zusage.Uebergeben = true

	if fehler := ablage.Festhalten(ctx, zusage, ueber(exemplar, 99)); fehler != nil {
		t.Fatalf("festhalten: %v", fehler)
	}

	if fehler := ablage.Freigeben(
		ctx, "h1"); !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("es kam %v", fehler)
	}

	offene, fehler := ablage.Reservierungen(ctx, exemplar.ID)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(offene) != 1 {
		t.Errorf("die übergebene Zusage ist weg: %+v", offene)
	}
}

// **Die Zusagen eines Vorgangs, nicht die eines Exemplars.**
//
// Die Verbuchung fragt „was habe ich für diesen Handel festgehalten" (T6,
// #1471).
// Eine Schleife über alle Besitzzeilen wäre dieselbe Frage von hinten — und
// sie würde mit dem Bestand teurer und nicht mit dem Handel.
func TestDieZusagenEinesVorgangsKommenZusammen(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	bestand := storage.NeuerSammlungsspeicher(db)
	ablage := storage.NeueReservierungsablage(db)
	ctx := probekontext(t)

	erstes := dreiStueck(t, bestand)

	zweites := &domain.Exemplar{
		Printing: "xyz-9", Menge: 5, Zustand: domain.ZustandNearMint,
		Sprache: "de",
	}

	kennung, fehler := bestand.Anlegen(ctx, zweites)
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	zweites.ID, zweites.Fassung = kennung, 1

	for _, zusage := range []domain.Reservierung{
		eineZusage(erstes, "v-1", 2),
		eineZusage(zweites, "v-1", 1),
		eineZusage(zweites, "v-2", 1),
	} {
		if gehalten := ablage.Festhalten(ctx, zusage, append(ueber(erstes, 99), ueber(zweites, 99)...)); gehalten != nil {
			t.Fatalf("festhalten: %v", gehalten)
		}
	}

	gefunden, fehler := ablage.FuerVorgang(ctx, "v-1")
	if fehler != nil {
		t.Fatalf("FuerVorgang: %v", fehler)
	}

	if len(gefunden) != 2 {
		t.Fatalf("es kamen %d Zusagen statt zweier: %+v", len(gefunden), gefunden)
	}

	// **Jede nennt ihre Besitzzeile** — ohne sie wüsste die Verbuchung nicht,
	// wovon sie abzieht.
	for _, zusage := range gefunden {
		if zusage.ExemplarID == 0 || zusage.Menge == 0 || zusage.Fassung == 0 {
			t.Errorf("die Zusage ist %+v", zusage)
		}
	}

	// **Und ein fremder Vorgang bekommt eine leere Liste und kein `nil`.**
	leer, fehler := ablage.FuerVorgang(ctx, "v-gibtesnicht")
	if fehler != nil {
		t.Fatalf("FuerVorgang: %v", fehler)
	}

	if leer == nil || len(leer) != 0 {
		t.Errorf("der fremde Vorgang gab %+v", leer)
	}
}

// **Ohne Profil geht auch das nicht** (0066).
func TestDieZusagenEinesVorgangsBrauchenEinProfil(t *testing.T) {
	t.Parallel()

	ablage := storage.NeueReservierungsablage(frischeDatenbank(t))

	if _, fehler := ablage.FuerVorgang(t.Context(), "v-1"); fehler == nil {
		t.Error("die Zusagen kamen ohne Profil")
	}
}
