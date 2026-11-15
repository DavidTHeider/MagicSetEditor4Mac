package storage_test

// **Was die Reservierungsablage abweist** — Werkstück `T2` von Phase 5.6
// (#1467).
//
// # Warum die Fehlerwege eine eigene Datei haben
//
// **Weil sie eine andere Frage beantworten** (§2): nebenan steht, **was** die
// Ablage hält; hier steht, was geschieht, wenn der Kontext kein Profil trägt,
// die Datenbank zu ist, die Besitzzeile fehlt oder ein von Hand geänderter
// Bestand eine Frist enthält, die keine ist.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ohne Profil im Kontext geschieht nichts.**
func TestDieReservierungsablageBrauchtEinProfil(t *testing.T) {
	t.Parallel()

	ablage := storage.NeueReservierungsablage(frischeDatenbank(t))
	ohne := t.Context()

	if fehler := ablage.Festhalten(ohne, domain.Reservierung{Vorgang: "h1"}, nil); fehler == nil {
		t.Error("festgehalten wurde ohne Profil")
	}

	if _, fehler := ablage.Reservierungen(ohne, 1); fehler == nil {
		t.Error("gelesen wurde ohne Profil")
	}

	if fehler := ablage.Freigeben(ohne, "h1"); fehler == nil {
		t.Error("freigegeben wurde ohne Profil")
	}
}

// **Eine geschlossene Datenbank meldet sich.**
func TestDieReservierungsablageMeldetEineGeschlosseneDatenbank(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ablage, ctx := storage.NeueReservierungsablage(db), probekontext(t)

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if fehler := ablage.Festhalten(ctx, domain.Reservierung{Vorgang: "h1"}, nil); fehler == nil {
		t.Error("festgehalten wurde in eine geschlossene Datenbank")
	}

	if _, fehler := ablage.Reservierungen(ctx, 1); fehler == nil {
		t.Error("gelesen wurde aus einer geschlossenen Datenbank")
	}

	if fehler := ablage.Freigeben(ctx, "h1"); fehler == nil {
		t.Error("freigegeben wurde in einer geschlossenen Datenbank")
	}
}

// **Eine Zusage auf eine Besitzzeile, die es nicht gibt, gibt es nicht.**
func TestEineZusageBrauchtEineBesitzzeile(t *testing.T) {
	t.Parallel()

	ablage := storage.NeueReservierungsablage(frischeDatenbank(t))

	if fehler := ablage.Festhalten(probekontext(t), domain.Reservierung{
		Vorgang: "h1", Revision: 1, ExemplarID: 99, Menge: 1,
	}, nil); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("es kam %v", fehler)
	}
}

// **Eine Zusage ohne Frist gilt, bis jemand sie löst.**
//
// **Der Nullwert ist kein Zeitpunkt und wird keiner**: er geht als leere
// Angabe hinein und kommt als Nullwert zurück. Wer ihn als „1. Januar des
// Jahres eins" schriebe, hätte eine Frist, die seit zweitausend Jahren um
// ist.
func TestEineZusageOhneFristKommtOhneFristZurueck(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ctx := probekontext(t)
	exemplar := dreiStueck(t, storage.NeuerSammlungsspeicher(db))
	ablage := storage.NeueReservierungsablage(db)

	ohneFrist := eineZusage(exemplar, "h1", 1)
	ohneFrist.Frist = time.Time{}

	if fehler := ablage.Festhalten(ctx, ohneFrist, ueber(exemplar, 99)); fehler != nil {
		t.Fatalf("festhalten: %v", fehler)
	}

	offene, fehler := ablage.Reservierungen(ctx, exemplar.ID)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(offene) != 1 || !offene[0].Frist.IsZero() {
		t.Errorf("zurück kam %+v", offene)
	}

	if !offene[0].Aktiv(time.Now().Add(100 * time.Hour)) {
		t.Error("eine Zusage ohne Frist läuft ab")
	}
}

// **Eine Frist, die keine ist, ist ein Befund.**
//
// **Ein von Hand geänderter Bestand ist der Normalfall** (§8) — ein `sqlite3`
// auf der Kommandozeile genügt. Die Anwendung darf daran nicht umfallen, und
// sie darf sich auch nichts zusammenreimen.
func TestEineUnlesbareFristIstEinBefund(t *testing.T) {
	t.Parallel()

	db := frischeDatenbank(t)
	ctx := probekontext(t)
	exemplar := dreiStueck(t, storage.NeuerSammlungsspeicher(db))
	ablage := storage.NeueReservierungsablage(db)

	if fehler := ablage.Festhalten(
		ctx, eineZusage(exemplar, "h1", 1), ueber(exemplar, 99)); fehler != nil {
		t.Fatalf("festhalten: %v", fehler)
	}

	if _, fehler := db.ExecContext(ctx,
		`UPDATE reservierungen SET frist = 'irgendwann'`); fehler != nil {
		t.Fatalf("verbiegen: %v", fehler)
	}

	if _, fehler := ablage.Reservierungen(ctx, exemplar.ID); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Errorf("es kam %v", fehler)
	}
}
