package app_test

// **Die Datenpflege an der Fassade** — Werkstück O8 von Phase 5.2
// (#1432).

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// pflegedoppel gibt feste Antworten.
type pflegedoppel struct {
	plan     domain.Massnahmenplan
	bilanz   service.Pflegebilanz
	gewaehlt []domain.Pflegewahl
	erstellt time.Time
	fehler   error
}

func (p *pflegedoppel) Vorschau(context.Context) (domain.Massnahmenplan, error) {
	return p.plan, p.fehler
}

func (p *pflegedoppel) Pflegen(
	_ context.Context, erstellt time.Time, wahl []domain.Pflegewahl,
) (service.Pflegebilanz, error) {
	p.erstellt, p.gewaehlt = erstellt, wahl

	return p.bilanz, p.fehler
}

func mitPflege(doppel *pflegedoppel) *app.App {
	anwendung := app.NeueApp("")
	anwendung.MitPflege(doppel)

	return anwendung
}

// **Die Vorschau nennt ihre Frist und je Befund die Revision.**
func TestDieVorschauNenntFristUndRevision(t *testing.T) {
	t.Parallel()

	erstellt := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)

	antwort := mitPflege(&pflegedoppel{plan: domain.Massnahmenplan{
		Erstellt: erstellt, Gueltigbis: erstellt.Add(domain.Pflegefrist),
		Befunde: []domain.Pflegebefund{{
			Art: "kartenwerteOhneKarte", Klasse: domain.PflegeSicher,
			Anzahl: 3, Revision: "abc", Beispiele: []string{"card_values:1:name"},
		}},
	}}).Pflegevorschau(t.Context())

	if antwort.Fehler != nil {
		t.Fatalf("die Vorschau: %v", antwort.Fehler)
	}

	if antwort.Erstellt != "2026-09-10T12:00:00Z" ||
		antwort.Gueltigbis != "2026-09-10T12:15:00Z" {
		t.Errorf("die Zeiten sind %q und %q", antwort.Erstellt, antwort.Gueltigbis)
	}

	if len(antwort.Befunde) != 1 || antwort.Befunde[0].Revision != "abc" ||
		antwort.Befunde[0].Klasse != "sicher" {
		t.Errorf("der Befund ist %+v", antwort.Befunde)
	}
}

// **Ein Befund ohne Beispiele hat eine leere Liste und kein `null`** (§3):
// „keine Beispiele" ist eine Aussage, `null` ist keine.
func TestEinBefundOhneBeispieleHatEineLeereListe(t *testing.T) {
	t.Parallel()

	antwort := mitPflege(&pflegedoppel{plan: domain.Massnahmenplan{
		Befunde: []domain.Pflegebefund{{Art: "x", Klasse: domain.PflegeSicher}},
	}}).Pflegevorschau(t.Context())

	if antwort.Befunde[0].Beispiele == nil {
		t.Error("die Beispiele sind null statt leer")
	}
}

func TestEineGescheiterteVorschauMeldetSich(t *testing.T) {
	t.Parallel()

	antwort := mitPflege(&pflegedoppel{
		fehler: errors.New("die Datenbank ist zu"),
	}).Pflegevorschau(t.Context())

	if antwort.Fehler == nil {
		t.Error("der Fehler kam nicht durch")
	}

	if antwort.Befunde == nil {
		t.Error("die Liste ist null statt leer")
	}
}

// **Der Zeitpunkt reist mit und wird gelesen.**
func TestDerZeitpunktDerVorschauReistMit(t *testing.T) {
	t.Parallel()

	doppel := &pflegedoppel{bilanz: service.Pflegebilanz{
		Gesichert: "/wo/es/liegt.db",
		Zeilen: []domain.Pflegezeile{
			{Art: "kartenwerteOhneKarte", Ausgang: domain.PflegeErledigt, Anzahl: 3},
		},
		Vorher:  []domain.Pflegebefund{{Art: "kartenwerteOhneKarte"}},
		Nachher: []domain.Pflegebefund{},
	}}

	antwort := mitPflege(doppel).Pflegen(t.Context(), "2026-09-10T12:00:00Z",
		[]app.PflegewahlDTO{{Art: "kartenwerteOhneKarte", Revision: "abc"}})
	if antwort.Fehler != nil {
		t.Fatalf("pflegen: %v", antwort.Fehler)
	}

	if !doppel.erstellt.Equal(time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)) {
		t.Errorf("der Zeitpunkt kam als %v an", doppel.erstellt)
	}

	if len(doppel.gewaehlt) != 1 || doppel.gewaehlt[0].Revision != "abc" {
		t.Errorf("die Wahl kam als %+v an", doppel.gewaehlt)
	}

	if antwort.Gesichert == "" || len(antwort.Zeilen) != 1 ||
		len(antwort.Vorher) != 1 || antwort.Nachher == nil {
		t.Errorf("die Bilanz ist %+v", antwort)
	}
}

// **Ein Zeitpunkt, der keiner ist, ist ein Fehler der Anfrage** — und keine
// frische Frist.
func TestEinKrummerZeitpunktWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	doppel := &pflegedoppel{}

	antwort := mitPflege(doppel).Pflegen(t.Context(), "irgendwann", nil)

	if antwort.Fehler == nil {
		t.Fatal("der krumme Zeitpunkt kam durch")
	}

	if doppel.gewaehlt != nil {
		t.Error("es wurde gepflegt, obwohl der Zeitpunkt unlesbar war")
	}
}

// **Die Bilanz kommt auch mit einem Fehler zurück** (§7, Regel 3): gepflegt
// ist gepflegt, und ein Fehler beim Nachsehen darf die Zeilen nicht
// verschlucken.
func TestDieBilanzKommtAuchMitEinemFehlerZurueck(t *testing.T) {
	t.Parallel()

	antwort := mitPflege(&pflegedoppel{
		bilanz: service.Pflegebilanz{
			Zeilen: []domain.Pflegezeile{
				{Art: "kartenwerteOhneKarte", Ausgang: domain.PflegeErledigt, Anzahl: 3},
			},
		},
		fehler: errors.New("danach ging die Datenbank zu"),
	}).Pflegen(t.Context(), "2026-09-10T12:00:00Z", nil)

	if antwort.Fehler == nil {
		t.Fatal("der Fehler wurde verschwiegen")
	}

	if len(antwort.Zeilen) != 1 || antwort.Zeilen[0].Ausgang != "erledigt" {
		t.Errorf("die Zeilen sind %+v", antwort.Zeilen)
	}
}

// **Der Name der Sicherung trägt den Augenblick** — es kann mehrere geben,
// und eine Sicherung überschreibt nichts.
func TestDerNameDerPflegesicherungTraegtDenAugenblick(t *testing.T) {
	t.Parallel()

	name := app.Pflegesicherungsname(time.Date(2026, 9, 10, 12, 34, 56, 0, time.UTC))

	if name != "bestand-vor-pflege-20260910-123456.db" {
		t.Errorf("der Name ist %q", name)
	}
}

// **Und die Wurzel für die Flächen ohne Fenster geht wirklich durch den
// Bestand** — über eine echte Datenbank, denn die Frage ist, ob die Naht
// verdrahtet ist (§10).
func TestDiePflegewurzelOhneFensterGehtDurchDenBestand(t *testing.T) {
	db := frischerBestand(t)
	ort := filepath.Join(t.TempDir(), "bestand.db")

	anwendung := app.PflegeAnwendungUeber(db, ort)

	vorschau := anwendung.Pflegevorschau(probekontext(t))
	if vorschau.Fehler != nil {
		t.Fatalf("die Vorschau: %v", vorschau.Fehler)
	}

	// **Ein frischer Bestand hat keinen Befund** — und eine Frist trotzdem.
	if len(vorschau.Befunde) != 0 {
		t.Errorf("ein frischer Bestand hat Befunde: %+v", vorschau.Befunde)
	}

	if vorschau.Gueltigbis == "" {
		t.Error("die Vorschau nennt keine Frist")
	}

	// **Mit leerer Wahl geschieht nichts, und es wird nicht gesichert.**
	bilanz := anwendung.Pflegen(probekontext(t), vorschau.Erstellt, nil)
	if bilanz.Fehler != nil {
		t.Fatalf("pflegen: %v", bilanz.Fehler)
	}

	if bilanz.Gesichert != "" {
		t.Errorf("es wurde gesichert, obwohl nichts zu tun war: %q", bilanz.Gesichert)
	}

	// **Und eine Maßnahme, die es nicht gibt, wird abgewiesen.**
	erfunden := anwendung.Pflegen(probekontext(t), vorschau.Erstellt,
		[]app.PflegewahlDTO{{Art: "alleKartenLoeschen", Revision: "abc"}})

	if erfunden.Fehler == nil {
		t.Error("eine erfundene Maßnahme kam durch")
	}
}

// **Und sie räumt wirklich weg — samt Sicherung und Nachprüfung.**
func TestDiePflegewurzelRaeumtWegUndSichertVorher(t *testing.T) {
	db := frischerBestand(t)
	ort := filepath.Join(t.TempDir(), "bestand.db")

	if _, fehler := db.ExecContext(probekontext(t), `PRAGMA foreign_keys = OFF`); fehler != nil {
		t.Fatalf("die Fremdschlüssel abschalten: %v", fehler)
	}

	if _, fehler := db.ExecContext(probekontext(t),
		`INSERT INTO card_values (card_id, feld, wert, reihenfolge)
		 VALUES (4711, 'name', 'eine Waise', 0)`); fehler != nil {
		t.Fatalf("die Waise legen: %v", fehler)
	}

	anwendung := app.PflegeAnwendungUeber(db, ort)

	vorschau := anwendung.Pflegevorschau(probekontext(t))
	if vorschau.Fehler != nil {
		t.Fatalf("die Vorschau: %v", vorschau.Fehler)
	}

	// **Eine Waise im Feldwert zieht eine in der Suchzeile nach sich** — die
	// Volltexttabelle hat ihren eigenen Auslöser. Gepflegt wird deshalb, was
	// dasteht, und nicht „der eine Befund".
	if len(vorschau.Befunde) < 1 {
		t.Fatalf("die Vorschau ist %+v", vorschau)
	}

	wahl := make([]app.PflegewahlDTO, 0, len(vorschau.Befunde))
	for _, befund := range vorschau.Befunde {
		wahl = append(wahl,
			app.PflegewahlDTO{Art: befund.Art, Revision: befund.Revision})
	}

	bilanz := anwendung.Pflegen(probekontext(t), vorschau.Erstellt, wahl)
	if bilanz.Fehler != nil {
		t.Fatalf("pflegen: %v", bilanz.Fehler)
	}

	if !strings.Contains(bilanz.Gesichert, "bestand-vor-pflege-") {
		t.Errorf("gesichert wurde nach %q", bilanz.Gesichert)
	}

	// **Erledigt oder schon weg** — und das zweite ist kein Mangel: das
	// Wegräumen der Feldwerte nimmt die Suchzeilen mit (die Volltexttabelle
	// hängt an ihrem Auslöser), und die zweite Maßnahme findet nichts mehr
	// vor. Genau dafür gibt es `schonWeg` und nicht „gescheitert".
	// **Und was nur informiert, wird übersprungen** — seit Befund 2 des
	// Zweitblicks über Phase 5.2 sieht die Pflege auch nach den
	// Fremdschlüsseln, und dieselbe Waise fällt dort ein zweites Mal auf.
	// Dort wird **nichts** gelöscht: „echte Korruption nicht durch
	// kosmetisches Löschen verdecken" steht so im Issue.
	for _, zeile := range bilanz.Zeilen {
		if zeile.Ausgang == "uebersprungen" && zeile.Grund == "nurInformieren" {
			continue
		}

		if zeile.Ausgang != "erledigt" && zeile.Ausgang != "schonWeg" {
			t.Errorf("die Zeile %+v ist weder erledigt noch schon weg", zeile)
		}
	}

	// **Die Nachprüfung ist die Abnahme**: danach ist nichts mehr da.
	if len(bilanz.Vorher) == 0 || len(bilanz.Nachher) != 0 {
		t.Errorf("die Stände sind %d und %d", len(bilanz.Vorher), len(bilanz.Nachher))
	}
}
