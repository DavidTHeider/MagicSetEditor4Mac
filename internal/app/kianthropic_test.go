package app_test

// **Anthropic als zweiter Anbieter** — Werkstück A8 von Phase 5.3
// (#1447).

import (
	"context"
	"slices"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// zweianbieter baut eine Anwendung mit beiden Quellen.
func zweianbieter(
	t *testing.T, werte map[string]string, lokal, fremd app.KILaufzeitquelle,
) *app.App {
	t.Helper()

	db := frischerBestand(t)

	for schluessel, wert := range werte {
		if fehler := storage.NeueEinstellungen(db).Setzen(
			probekontext(t), schluessel, wert); fehler != nil {
			t.Fatalf("%s setzen: %v", schluessel, fehler)
		}
	}

	return app.KIAnwendungUeber(db).MitKIZuhause(t.TempDir()).MitKIAnbietern(
		func(string) app.KILaufzeitquelle { return lokal },
		map[string]func(adresse string) app.KILaufzeitquelle{
			"anthropic": func(string) app.KILaufzeitquelle { return fremd },
		})
}

// derSchalter ist der Einstellungsschlüssel von Anthropic.
func derSchalter() string {
	return domain.DienstEinstellung(domain.DienstAnthropic)
}

// derCloudalias ist das Modell, mit dem diese Läufe fragen.
//
// **Einer und keiner mehr**: welches Modell antwortet, ist für keinen dieser
// Läufe die Frage — sie prüfen, **wer** antwortet und **ob** er darf.
const derCloudalias = "claude-opus-5"

// einCloudmodell baut ein Modell, das nicht auf diesem Rechner rechnet.
func einCloudmodell() domain.KIModell {
	modell := einKoenner(derCloudalias)
	modell.Anbieter = "anthropic"
	modell.Ort = domain.KIOrtCloud

	return modell
}

// **Jeder Anbieter steht in der Antwort, auch wenn er aus ist.**
//
// # Warum auch der abgeschaltete
//
// **Weil „aus" eine Auskunft ist.** Wer ihn wegließe, ließe jemanden suchen,
// warum seine Modelle fehlen — und die Antwort stünde nirgends (§6).
func TestJederAnbieterStehtInDerAntwort(t *testing.T) {
	t.Parallel()

	antwort := zweianbieter(t, map[string]string{
		domain.KILaufzeit: domain.KILaufzeitAn,
	}, &laufzeitdoppel{
		fassung: "0.12.3",
		modelle: []domain.KIModell{einModell("gemma4:8b")},
	}, &laufzeitdoppel{fassung: "anthropic-version 2023-06-01"}).
		KIModelle(probekontext(t))
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	// **Vier seit A10** (#1449) — und sie werden **genannt** und nicht
	// gezählt: eine Zahl allein bliebe grün, wenn ein Anbieter gegen einen
	// anderen getauscht würde.
	namen := make([]string, 0, len(antwort.Anbieter))
	for _, einer := range antwort.Anbieter {
		namen = append(namen, einer.Name)
	}

	slices.Sort(namen)

	if !slices.Equal(namen, []string{"anthropic", "apple", "ollama", "openai", "xai"}) {
		t.Fatalf("die Anbieter sind %v", namen)
	}

	if einrichtungVon(t, antwort, "ollama") != app.KIEinrichtungBereit {
		t.Errorf("Ollama ist %q", einrichtungVon(t, antwort, "ollama"))
	}

	// **Jeder Cloudanbieter ist aus** — die Vorgabe seines Dienstes.
	for _, name := range []string{"anthropic", "openai", "xai"} {
		if einrichtungVon(t, antwort, name) != app.KIEinrichtungAus {
			t.Errorf("%s ist %q", name, einrichtungVon(t, antwort, name))
		}
	}
}

// **Der beste Stand steht über allem.**
//
// **Und nicht der schlechteste**: ein eingerichteter Anbieter macht die
// Anwendung brauchbar, auch wenn ein zweiter aus ist. Andersherum sähe eine
// funktionierende Einrichtung wie eine kaputte aus, sobald jemand einen
// zweiten Anbieter kennt und nicht benutzt.
func TestDerBesteStandStehtUeberAllem(t *testing.T) {
	t.Parallel()

	antwort := zweianbieter(t, map[string]string{
		domain.KILaufzeit: domain.KILaufzeitAn,
	}, &laufzeitdoppel{
		fassung: "0.12.3",
		modelle: []domain.KIModell{einModell("gemma4:8b")},
	}, &laufzeitdoppel{}).KIModelle(probekontext(t))

	if antwort.Einrichtung != app.KIEinrichtungBereit {
		t.Errorf("die Einrichtung ist %q", antwort.Einrichtung)
	}
}

// **Die Modelle beider Anbieter stehen in einem Katalog.**
func TestDieModelleBeiderAnbieterStehenInEinemKatalog(t *testing.T) {
	t.Parallel()

	antwort := zweianbieter(t, map[string]string{
		domain.KILaufzeit: domain.KILaufzeitAn,
		derSchalter():     domain.DienstAn,
	}, &laufzeitdoppel{
		fassung: "0.12.3",
		modelle: []domain.KIModell{einModell("gemma4:8b")},
	}, &laufzeitdoppel{
		fassung: "anthropic-version 2023-06-01",
		modelle: []domain.KIModell{einCloudmodell()},
	}).KIModelle(probekontext(t))
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Modelle) != 2 {
		t.Fatalf("%d Modelle: %+v", len(antwort.Modelle), antwort.Modelle)
	}

	gesehen := map[string]string{}
	for _, modell := range antwort.Modelle {
		gesehen[modell.Alias] = modell.Anbieter
	}

	if gesehen["gemma4:8b"] != "ollama" ||
		gesehen["claude-opus-5"] != "anthropic" {
		t.Errorf("die Modelle sind %v", gesehen)
	}
}

// **Unter »nur lokal« bleibt Anthropic draußen — und sagt es.**
//
// **Als Befund und nicht als Schweigen**: wer den Schalter umgelegt hat und
// keine Modelle sieht, soll den Grund lesen können.
func TestUnterNurLokalBleibtAnthropicDraussen(t *testing.T) {
	t.Parallel()

	antwort := zweianbieter(t, map[string]string{
		derSchalter():             domain.DienstAn,
		domain.RichtlinieNurLokal: domain.NurLokalAn,
	}, &laufzeitdoppel{}, &laufzeitdoppel{
		fassung: "anthropic-version 2023-06-01",
		modelle: []domain.KIModell{einCloudmodell()},
	}).KIModelle(probekontext(t))
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Befunde) == 0 {
		t.Fatal("Anthropic fiel unter »nur lokal« ohne Befund aus")
	}
}

// **Ein Vorschlag geht an den Anbieter, den er nennt.**
func TestEinVorschlagGehtAnDenGenanntenAnbieter(t *testing.T) {
	t.Parallel()

	fremd := &laufzeitdoppel{
		fassung: "anthropic-version 2023-06-01",
		modelle: []domain.KIModell{einCloudmodell()},
		antwort: app.KIRohantwort{
			Inhalt: `{"name": "Blitzschlag"}`, Modell: "claude-opus-5",
		},
	}

	welt := einekiwelt(t)
	welt.anwendung.MitKIAnbietern(
		func(string) app.KILaufzeitquelle { return welt.laufzeit },
		map[string]func(adresse string) app.KILaufzeitquelle{
			"anthropic": func(string) app.KILaufzeitquelle { return fremd },
		})

	if fehler := storage.NeueEinstellungen(welt.db).Setzen(probekontext(t),
		derSchalter(), domain.DienstAn); fehler != nil {
		t.Fatalf("einschalten: %v", fehler)
	}

	// **Die Freigabe gehört zum Ruf** — ein Modell, das anderswo rechnet,
	// kostet Geld, und ohne den Nachweis eines Menschen geht kein solcher
	// Auftrag hinaus ([app.App.MitKIFreigabe]).
	welt.anwendung.MitKosten(stummeskijournal{})
	welt.anwendung.MitKIFreigabe(
		"der Mensch hat zugestimmt", domain.Betrag{Wert: 100, Waehrung: "EUR"})

	antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"anthropic", "claude-opus-5", "Schreibe.", "")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Modell != "claude-opus-5" {
		t.Errorf("das Modell ist %q", antwort.Modell)
	}

	// **Die lokale Laufzeit wurde nicht gefragt** — das ist die Gegenprobe:
	// ohne sie sagte der Lauf nur, dass irgendetwas geantwortet hat.
	if len(welt.laufzeit.schemen) != 0 {
		t.Errorf("die lokale Laufzeit bekam %d Fragen", len(welt.laufzeit.schemen))
	}

	if len(fremd.schemen) != 1 {
		t.Errorf("Anthropic bekam %d Fragen", len(fremd.schemen))
	}
}

// **Ein Anbieter, den es nicht gibt, wird abgewiesen.**
func TestEinUnbekannterKIAnbieterWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	if antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"gibtesnicht", "x", "y", ""); antwort.Fehler == nil {
		t.Fatal("der unbekannte Anbieter kam durch")
	}
}

// **Ein abgeschalteter Anbieter ebenso.**
func TestEinAbgeschalteterAnbieterWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	if antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"anthropic", "claude-opus-5", "x", ""); antwort.Fehler == nil {
		t.Fatal("der abgeschaltete Anbieter kam durch")
	}
}

// **Ohne Freigabe geht kein Auftrag an einen Anbieter, der Geld kostet.**
//
// **Das ist die Gegenprobe zur Zeile darüber** — ohne sie sagte der Lauf nur,
// dass ein Vorschlag ankommt, nicht dass die Freigabe ihn trägt.
func TestOhneFreigabeGehtKeinKostenpflichtigerAuftragHinaus(t *testing.T) {
	t.Parallel()

	fremd := &laufzeitdoppel{
		fassung: "anthropic-version 2023-06-01",
		modelle: []domain.KIModell{einCloudmodell()},
	}

	welt := einekiwelt(t)
	welt.anwendung.MitKIAnbietern(
		func(string) app.KILaufzeitquelle { return welt.laufzeit },
		map[string]func(adresse string) app.KILaufzeitquelle{
			"anthropic": func(string) app.KILaufzeitquelle { return fremd },
		})
	welt.anwendung.MitKosten(stummeskijournal{})

	if fehler := storage.NeueEinstellungen(welt.db).Setzen(probekontext(t),
		derSchalter(), domain.DienstAn); fehler != nil {
		t.Fatalf("einschalten: %v", fehler)
	}

	antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"anthropic", "claude-opus-5", "Schreibe.", "")

	if antwort.Fehler == nil {
		t.Fatal("ohne Freigabe kam ein Vorschlag")
	}

	// **Und nichts ging hinaus** — das ist die Zusage des Issues: „Unter
	// Local-only, fehlendem Scope oder Budget entstehen **null**
	// Anbieteranfragen."
	if len(fremd.schemen) != 0 {
		t.Errorf("Anthropic bekam %d Fragen", len(fremd.schemen))
	}
}

// stummeskijournal reserviert alles und bucht nichts.
type stummeskijournal struct{}

func (stummeskijournal) Reservieren(
	_ context.Context, auftrag service.Kostenauftrag, _ string,
) (domain.Kostenposten, error) {
	return domain.Kostenposten{Kennung: auftrag.Kennung}, nil
}

func (stummeskijournal) Stand(context.Context, string, domain.Zeitraum) (
	domain.Haushaltsstand, error,
) {
	return domain.Haushaltsstand{}, nil
}

func (stummeskijournal) Posten(context.Context, string, domain.Zeitraum) (
	[]domain.Kostenposten, error,
) {
	return nil, nil
}

func (stummeskijournal) BudgetSetzen(
	context.Context, string, domain.Zeitraum, domain.Betrag,
) error {
	return nil
}

func (stummeskijournal) Abschliessen(context.Context, string, domain.Betrag) (
	domain.Kostenposten, error,
) {
	return domain.Kostenposten{}, nil
}

func (stummeskijournal) Freigeben(context.Context, string) (
	domain.Kostenposten, error,
) {
	return domain.Kostenposten{}, nil
}

func (stummeskijournal) AlsUnklarMerken(context.Context, string, string) (
	domain.Kostenposten, error,
) {
	return domain.Kostenposten{}, nil
}

// **Der Anbieter steht am Vorschlag, nicht nur das Modell.**
//
// # Warum das Issue es ausdrücklich verlangt
//
// „Provider-/Modellname im Vorschlag sichtbar." Mit **einem** Anbieter war
// der Alias die ganze Auskunft; mit zweien ist er die halbe — zwei Anbieter
// dürfen denselben Namen führen, und wer den Vorschlag später ansieht, soll
// nicht raten müssen, wer geantwortet hat.
func TestDerAnbieterStehtAmVorschlag(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "Schreibe den Regeltext.", "")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.Anbieter != "ollama" {
		t.Errorf("der Anbieter am Vorschlag ist %q", antwort.Anbieter)
	}

	if antwort.Modell == "" {
		t.Error("das Modell am Vorschlag fehlt")
	}
}
