package storage_test

// **Eine Zusage wird als übergeben vermerkt** —
// #1643.
//
// # Wogegen hier geprüft wird
//
// **Die Spalte `uebergeben` wurde nur beim Einfügen geschrieben.** Es gab
// keine Anweisung, die sie fortschrieb — und `Verbuchungsstelle.abgaengeFuer`
// zieht genau daraus die Abgangszeilen eines Belegs. Ohne sie konnte **nur
// die empfangende Seite** verbuchen: wer Karten weggibt, bekam „es fehlen N
// Stück an bestätigter Übergabe", richtig gemeldet und ohne einen Weg, es in
// Ordnung zu bringen.

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Die Zusagen eines Vorgangs gelten danach als übergeben.**
func TestDieZusagenEinesVorgangsWerdenUebergeben(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	bestand := storage.NeuerSammlungsspeicher(db)
	ablage := storage.NeueReservierungsablage(db)

	ctx := probekontext(t)
	exemplar := dreiStueck(t, bestand)

	for _, zusage := range []domain.Reservierung{
		eineZusage(exemplar, "v-1", 1),
		eineZusage(exemplar, "v-2", 1),
	} {
		if fehler := ablage.Festhalten(ctx, zusage, ueber(exemplar, 99)); fehler != nil {
			t.Fatalf("festhalten: %v", fehler)
		}
	}

	if fehler := ablage.Uebergeben(ctx, "v-1"); fehler != nil {
		t.Fatalf("Uebergeben: %v", fehler)
	}

	gefunden, fehler := ablage.FuerVorgang(ctx, "v-1")
	if fehler != nil {
		t.Fatalf("FuerVorgang: %v", fehler)
	}

	if len(gefunden) != 1 || !gefunden[0].Uebergeben {
		t.Errorf("die Zusage ist %+v", gefunden)
	}

	// **Der fremde Vorgang bleibt, wie er war** — ein Vermerk gilt einem
	// Handel und nicht einer Besitzzeile.
	fremd, fehler := ablage.FuerVorgang(ctx, "v-2")
	if fehler != nil {
		t.Fatalf("FuerVorgang: %v", fehler)
	}

	if len(fremd) != 1 || fremd[0].Uebergeben {
		t.Errorf("der fremde Vorgang ist %+v", fremd)
	}
}

// **Und danach gibt nichts mehr frei.**
//
// **Die Gegenprobe zur Sperre in `Freigeben`**: sie zählt die übergebenen
// Zusagen, und bis zu diesem Werkstück konnte diese Zahl **nie** größer als
// null werden — die Sperre war unerreichbar.
func TestNachDerUebergabeGibtNichtsMehrFrei(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	bestand := storage.NeuerSammlungsspeicher(db)
	ablage := storage.NeueReservierungsablage(db)

	ctx := probekontext(t)

	if fehler := ablage.Festhalten(
		ctx, eineZusage(dreiStueck(t, bestand), "v-1", 1), wieDreiStueck(99)); fehler != nil {
		t.Fatalf("festhalten: %v", fehler)
	}

	if fehler := ablage.Uebergeben(ctx, "v-1"); fehler != nil {
		t.Fatalf("Uebergeben: %v", fehler)
	}

	if fehler := ablage.Freigeben(ctx, "v-1"); fehler == nil {
		t.Fatal("die übergebene Zusage wurde freigegeben")
	}
}

// **Ohne Profil geht auch das nicht** (0066).
func TestDerUebergabevermerkBrauchtEinProfil(t *testing.T) {
	t.Parallel()

	ablage := storage.NeueReservierungsablage(frischeDatenbank(t))

	if fehler := ablage.Uebergeben(t.Context(), "v-1"); fehler == nil {
		t.Error("der Vermerk kam ohne Profil")
	}
}

// **Ein Bestand ohne Zusagentabelle meldet sich.**
func TestEinBestandOhneZusagenVermerktNichts(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage := storage.NeueReservierungsablage(db)

	if _, fehler := db.ExecContext(
		probekontext(t), "DROP TABLE reservierungen"); fehler != nil {
		t.Fatalf("die Zusagen wegnehmen: %v", fehler)
	}

	if fehler := ablage.Uebergeben(
		probekontext(t), "v-1"); fehler == nil {
		t.Error("der fehlende Bestand vermerkte")
	}
}

// **Die Vertragsgrenze gilt in derselben Transaktion** — Befund 1 des
// Nachlaufs zum Zweitblick über Phase 5.6.
//
// # Warum die Prüfung im Dienst nicht reichte
//
// **Weil sie vor der Transaktion lief.** `Zusagenwerk.passtInDenVertrag`
// liest die Summe, dann schreibt der Bestand — und die Transaktion prüfte
// nur Zusagen **desselben Exemplars**. Zwei Rufe für **zwei verschiedene**
// Besitzzeilen lasen beide „null von eins" und schrieben beide. Der Vertrag
// erlaubte ein Stück, die Verbuchung fand zwei Abgänge.
func TestDieVertragsgrenzeGiltInDerselbenTransaktion(t *testing.T) {
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

	// **Der Vertrag nimmt ein Stück.** Das erste geht durch.
	vertrag := ueber(erstes, 1)

	if erste := ablage.Festhalten(
		ctx, eineZusage(erstes, "v-1", 1), vertrag); erste != nil {
		t.Fatalf("die erste Zusage: %v", erste)
	}

	// **Das zweite ist eine andere Besitzzeile** — und trotzdem zu viel.
	// Seit #2492 ist sie auch eine andere Variante (die Sammlung führt je
	// Variante eine Zeile), und der Vertrag nennt sie nicht.
	if zweite := ablage.Festhalten(
		ctx, eineZusage(zweites, "v-1", 1), vertrag); zweite == nil {
		t.Fatal("die zweite Zusage überschritt den Vertrag")
	}

	gefunden, fehler := ablage.FuerVorgang(ctx, "v-1")
	if fehler != nil {
		t.Fatalf("FuerVorgang: %v", fehler)
	}

	if len(gefunden) != 1 {
		t.Errorf("festgehalten sind %d Zusagen", len(gefunden))
	}
}

// **Eine Zusagentabelle, die sich nicht zählen lässt, hält nichts fest.**
//
// **Der Fehlerweg der Vertragsgrenze**: sie zählt in derselben Transaktion,
// und was dabei schiefgeht, darf keine Zusage durchlassen.
func TestEineUnzaehlbareZusagentabelleHaeltNichtsFest(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	bestand := storage.NeuerSammlungsspeicher(db)
	ablage := storage.NeueReservierungsablage(db)

	ctx := probekontext(t)
	exemplar := dreiStueck(t, bestand)

	// **Die Tabelle verschwindet, das Exemplar bleibt** — der Weg trägt bis
	// zur Zählung und scheitert dort.
	if _, fehler := db.ExecContext(
		ctx, "DROP TABLE reservierungen"); fehler != nil {
		t.Fatalf("die Zusagen wegnehmen: %v", fehler)
	}

	if fehler := ablage.Festhalten(
		ctx, eineZusage(exemplar, "v-1", 1), ueber(exemplar, 9)); fehler == nil {
		t.Error("die unzählbare Tabelle hielt fest")
	}
}

// **Eine Zustellung führt die Wirkung an den Zusagen mit aus** — Befund 3
// des Nachlaufs zum Zweitblick über Phase 5.6.
//
// # Warum das in derselben Transaktion steht
//
// **Weil sonst ein halber Stand zurückbliebe.** Geschah die Wirkung vorher
// für sich, blieb nach einem gescheiterten Schreiben eine Zusage als
// „übergeben" stehen, während der Vorgang noch auf `vorbereitet` stand — und
// `Freigeben` gab sie nie wieder frei.
func TestEineZustellungFuehrtDieWirkungMitAus(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	bestand := storage.NeuerSammlungsspeicher(db)
	zusagen := storage.NeueReservierungsablage(db)
	briefkasten := storage.NeueBriefkastenablage(db)

	ctx := probekontext(t)

	if fehler := zusagen.Festhalten(
		ctx, eineZusage(dreiStueck(t, bestand), "v-1", 1), wieDreiStueck(9)); fehler != nil {
		t.Fatalf("festhalten: %v", fehler)
	}

	if fehler := briefkasten.Zustellen(ctx,
		domain.Inboxeintrag{Schluessel: "s-1", Ergebnis: "v-1/1"}, domain.Handelsangebot{},
		eineZugestellteRevision(), domain.ZusagenUebergeben); fehler != nil {
		t.Fatalf("Zustellen: %v", fehler)
	}

	gefunden, fehler := zusagen.FuerVorgang(ctx, "v-1")
	if fehler != nil {
		t.Fatalf("FuerVorgang: %v", fehler)
	}

	if len(gefunden) != 1 || !gefunden[0].Uebergeben {
		t.Errorf("die Zusage ist %+v", gefunden)
	}
}

// **Und ein Abbruch gibt sie im selben Zug frei.**
func TestEineZustellungGibtDieZusagenFrei(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	bestand := storage.NeuerSammlungsspeicher(db)
	zusagen := storage.NeueReservierungsablage(db)
	briefkasten := storage.NeueBriefkastenablage(db)

	ctx := probekontext(t)

	if fehler := zusagen.Festhalten(
		ctx, eineZusage(dreiStueck(t, bestand), "v-1", 1), wieDreiStueck(9)); fehler != nil {
		t.Fatalf("festhalten: %v", fehler)
	}

	if fehler := briefkasten.Zustellen(ctx,
		domain.Inboxeintrag{Schluessel: "s-1", Ergebnis: "v-1/1"}, domain.Handelsangebot{},
		eineZugestellteRevision(), domain.ZusagenFreigeben); fehler != nil {
		t.Fatalf("Zustellen: %v", fehler)
	}

	gefunden, fehler := zusagen.FuerVorgang(ctx, "v-1")
	if fehler != nil {
		t.Fatalf("FuerVorgang: %v", fehler)
	}

	if len(gefunden) != 0 {
		t.Errorf("festgehalten sind noch %+v", gefunden)
	}
}

// **Eine übergebene Zusage hält die ganze Zustellung auf.**
//
// **Die Gegenprobe zur Transaktion**: weist das Freigeben ab, bleibt auch
// der Vermerk aus — ein halber Stand ist genau das, was diese eine
// Transaktion verhindert.
func TestEineUebergebeneZusageHaeltDieZustellungAuf(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	bestand := storage.NeuerSammlungsspeicher(db)
	zusagen := storage.NeueReservierungsablage(db)
	briefkasten := storage.NeueBriefkastenablage(db)

	ctx := probekontext(t)

	if fehler := zusagen.Festhalten(
		ctx, eineZusage(dreiStueck(t, bestand), "v-1", 1), wieDreiStueck(9)); fehler != nil {
		t.Fatalf("festhalten: %v", fehler)
	}

	if fehler := zusagen.Uebergeben(ctx, "v-1"); fehler != nil {
		t.Fatalf("Uebergeben: %v", fehler)
	}

	if fehler := briefkasten.Zustellen(ctx,
		domain.Inboxeintrag{Schluessel: "s-1", Ergebnis: "v-1/1"}, domain.Handelsangebot{},
		eineZugestellteRevision(), domain.ZusagenFreigeben); fehler == nil {
		t.Fatal("die Zustellung gab eine übergebene Zusage frei")
	}

	vermerkt, fehler := briefkasten.Nachschlagen(ctx, "s-1")
	if fehler != nil {
		t.Fatalf("Nachschlagen: %v", fehler)
	}

	if vermerkt != nil {
		t.Error("der Vermerk blieb stehen, obwohl die Wirkung scheiterte")
	}
}

// eineZugestellteRevision ist das Angebot, das eine Zustellung schreibt.
func eineZugestellteRevision() domain.Handelsangebot {
	return domain.Handelsangebot{
		Vertragsfassung: domain.Handelsvertragsfassung,
		Vorgang:         "v-1", Revision: 1,
		Art: domain.HandelTausch, Initiator: "bich",
		Parteien: []string{"bich", "bdort"},
		Zustand:  domain.HandelAngeboten,
		Positionen: []domain.Handelsposition{{
			Von: "bich", Printing: "a1b2", Menge: 1,
			Guete: "nearMint", Sprache: "de", Finish: "nonfoil",
		}},
	}
}
