package app_test

// **Was aus einer Modellantwort wird** — Werkstück A7 von Phase 5.3
// (#1446).

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Wert kann Text oder Liste sein.**
//
// # Warum beides gelesen wird
//
// **Weil ein Modell auf ein einfaches Feld mit einer Zeichenkette antwortet
// und auf ein Mehrfachfeld mit einem Array.** Wer nur eines von beiden läse,
// verlöre die Hälfte jeder Antwort — still, weil das Schema die andere
// Hälfte gar nicht erst sieht.
func TestEinWertKommtAlsTextOderAlsListeAn(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.stile([]domain.Field{
		{Name: "name", Kind: domain.KindText},
		{Name: "farben", Kind: domain.KindMultipleChoice, Choices: []string{"W", "U"}},
	})
	welt.laufzeit.antwort = app.KIRohantwort{
		Inhalt: `{"name": "Blitzschlag", "farben": ["W", "U"]}`,
		Modell: "gemma4:8b",
	}

	antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "x", "")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	gefunden := map[string][]string{}
	for _, feld := range antwort.Felder {
		gefunden[feld.Feld] = feld.Werte
	}

	if len(gefunden["name"]) != 1 || gefunden["name"][0] != "Blitzschlag" {
		t.Errorf("»name« ist %v", gefunden["name"])
	}

	if len(gefunden["farben"]) != 2 {
		t.Errorf("»farben« ist %v", gefunden["farben"])
	}
}

// **Eine Antwort, die kein JSON-Objekt ist, wird abgewiesen.**
func TestEineAntwortOhneJSONObjektWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.laufzeit.antwort = app.KIRohantwort{
		Inhalt: "Ich bin ein Sprachmodell.", Modell: "gemma4:8b",
	}

	antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "x", "")

	if antwort.Fehler == nil {
		t.Fatal("der Fließtext kam als Vorschlag durch")
	}

	// **Das Modell steht in der Meldung** — es ist die Auskunft, die
	// weiterhilft: ein anderes hielte die Form vielleicht.
	if !strings.Contains(antwort.Fehler.Meldung, "gemma4:8b") {
		t.Errorf("die Meldung ist %q", antwort.Fehler.Meldung)
	}
}

// **Und ein Wert, der weder Text noch Liste ist, ebenso.**
func TestEinWertOhneFormWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.laufzeit.antwort = app.KIRohantwort{
		Inhalt: `{"name": {"tief": "verschachtelt"}}`, Modell: "gemma4:8b",
	}

	if antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "x", ""); antwort.Fehler == nil {
		t.Fatal("der verschachtelte Wert kam durch")
	}
}

// **Was das Schema nicht durchlässt, wird nicht geschrieben.**
//
// **Und die gute Hälfte bleibt**: ein Vorschlag, der halb passt, ist kein
// Grund, sie wegzuwerfen — und keiner, sie stillschweigend zu übernehmen.
func TestWasDasSchemaNichtDurchlaesstWirdNichtGeschrieben(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.laufzeit.antwort = app.KIRohantwort{
		Inhalt: `{"name": "Blitzschlag", "rarity": "episch"}`,
		Modell: "gemma4:8b",
	}

	antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "x", "")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Befunde) == 0 {
		t.Fatal("die unbekannte Wahl kam ohne Befund durch")
	}

	for _, feld := range antwort.Felder {
		if feld.Feld == "rarity" {
			t.Errorf("»rarity« steht trotzdem im Vorschlag: %v", feld.Werte)
		}
	}
}

// **Ein Ruf ohne Profil wird abgewiesen.**
//
// **Ohne Profil lässt sich nicht sagen, wessen Daten gemeint sind** — und ein
// Auftrag, der keinem gehört, ließe sich später keinem zuordnen.
func TestEinKIVorschlagOhneProfilWirdAbgewiesen(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	if antwort := welt.anwendung.KIVorschlagen(t.Context(), welt.karte,
		"ollama", "gemma4:8b", "x", ""); antwort.Fehler == nil {
		t.Fatal("der Ruf ohne Profil kam durch")
	}
}

// **Ein Vorschlag, den es nicht gibt, lässt sich nicht übernehmen.**
func TestEinUnbekannterVorschlagLaesstSichNichtUebernehmen(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	if antwort := welt.anwendung.KIVorschlagUebernehmen(probekontext(t),
		"gibtesnicht", []string{"name"}); antwort.Fehler == nil {
		t.Fatal("der unbekannte Vorschlag ließ sich übernehmen")
	}
}

// **Ohne Laufzeit lässt sich auch nichts übernehmen.**
func TestOhneLaufzeitLaesstSichNichtsUebernehmen(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.aus(t)

	if antwort := welt.anwendung.KIVorschlagUebernehmen(probekontext(t),
		"ki-4-1", []string{"name"}); antwort.Fehler == nil {
		t.Fatal("ohne Laufzeit ließ sich etwas übernehmen")
	}
}

// **Der Einzelzug schreibt nur auf Verlangen.**
//
// **Das ist die Fassung für die Flächen ohne Fenster** — dort sagt
// `uebernehmen` vorher, was am Fenster der zweite Klick sagt.
func TestDerEinzelzugSchreibtNurAufVerlangen(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	bezug := app.KIVorschlagsbezug{
		KarteID: welt.karte, Anbieter: "ollama", Alias: "gemma4:8b",
		Auftrag: "x",
	}

	ohne := welt.anwendung.KIVorschlagInEinemZug(probekontext(t), bezug,
		false, nil)
	if ohne.Fehler != nil {
		t.Fatalf("%+v", ohne.Fehler)
	}

	if ohne.Uebernommen != 0 {
		t.Errorf("%d Felder geschrieben, erwartet keine", ohne.Uebernommen)
	}

	mit := welt.anwendung.KIVorschlagInEinemZug(probekontext(t), bezug,
		true, nil)
	if mit.Fehler != nil {
		t.Fatalf("%+v", mit.Fehler)
	}

	if mit.Uebernommen == 0 {
		t.Error("mit »uebernehmen« wurde nichts geschrieben")
	}

	if !strings.Contains(welt.kartenwerte(t), "Blitzschlag") {
		t.Errorf("die Karte trägt %q", welt.kartenwerte(t))
	}
}

// **Und er reicht einen Fehler des ersten Schritts durch.**
func TestDerEinzelzugReichtEinenFehlerDurch(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.aus(t)

	if antwort := welt.anwendung.KIVorschlagInEinemZug(probekontext(t),
		app.KIVorschlagsbezug{
			KarteID: welt.karte, Anbieter: "ollama", Alias: "gemma4:8b",
			Auftrag: "x",
		}, true, nil); antwort.Fehler == nil {
		t.Fatal("der Fehler kam nicht durch")
	}
}

// **Und einen des zweiten ebenso — mit dem Vorschlag daneben.**
//
// **Weil beides zusammengehört**: was nicht geschrieben werden konnte, soll
// der Aufrufer trotzdem sehen.
func TestDerEinzelzugReichtEinenUebernahmefehlerDurch(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.karten.schreibfehler = errors.New("die Karte ist gesperrt")

	antwort := welt.anwendung.KIVorschlagInEinemZug(probekontext(t),
		app.KIVorschlagsbezug{
			KarteID: welt.karte, Anbieter: "ollama", Alias: "gemma4:8b",
			Auftrag: "x",
		}, true, nil)

	if antwort.Fehler == nil {
		t.Fatal("der Schreibfehler kam nicht durch")
	}

	if antwort.Kennung == "" {
		t.Error("der Vorschlag ging mit dem Fehler verloren")
	}
}

// **Ein Vorschlag zu einer Karte, deren Schreibweg fehlt, scheitert.**
func TestEinVorschlagOhneSchreibwegScheitert(t *testing.T) {
	t.Parallel()

	db := frischerBestand(t)

	if fehler := storage.NeueEinstellungen(db).Setzen(probekontext(t),
		domain.KILaufzeit, domain.KILaufzeitAn); fehler != nil {
		t.Fatalf("einschalten: %v", fehler)
	}

	anwendung := app.KIAnwendungUeber(db).MitKIZuhause(t.TempDir()).
		MitKILaufzeit(func(string) app.KILaufzeitquelle {
			return &laufzeitdoppel{
				fassung: "0.12.3",
				modelle: []domain.KIModell{einKoenner("gemma4:8b")},
			}
		})

	anwendung.MitEditor(nil, nil, nil, nil)

	if antwort := anwendung.KIVorschlagen(probekontext(t), 4,
		"ollama", "gemma4:8b", "x", ""); antwort.Fehler == nil {
		t.Fatal("ohne Schreibweg kam ein Vorschlag")
	}
}

// **Ein Schemafehler beim Vorschlagen kommt durch.**
//
// **Und zwar der des Pakets**: das Auftragswerk liest das Schema beim Anlegen
// und die Naht noch einmal für die Form der Frage. Scheitert es dazwischen,
// gibt es keinen Vorschlag — und nicht einen ohne Form.
func TestEinSchemafehlerBeimVorschlagenKommtDurch(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.anwendung.MitStilwahl(&zweimalDannStolpernd{felder: welt.felder})

	if antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "x", ""); antwort.Fehler == nil {
		t.Fatal("der Schemafehler kam nicht durch")
	}
}

// zweimalDannStolpernd antwortet zweimal und scheitert danach.
//
// **Zweimal**, weil das Auftragswerk das Schema beim Anlegen liest und die
// Naht es danach für die Form der Frage noch einmal holt.
type zweimalDannStolpernd struct {
	app.Stilwaehler

	felder  []domain.Field
	gelesen int
}

func (z *zweimalDannStolpernd) Kartenfelder(context.Context, int64) (
	[]domain.Field, error,
) {
	z.gelesen++

	if z.gelesen > 1 {
		return nil, errors.New("das Paket ist nicht mehr zu lesen")
	}

	return z.felder, nil
}

// **Eine Laufzeit, die nicht redet, gibt keinen Vorschlag.**
func TestEineStummeLaufzeitGibtKeinenVorschlag(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.laufzeit.redefehler = errors.New("das Modell antwortet nicht")

	if antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "x", ""); antwort.Fehler == nil {
		t.Fatal("die stumme Laufzeit gab einen Vorschlag")
	}
}

// **Ein Paket, das zwischen Anlegen und Prüfen unlesbar wird, gibt keinen
// Vorschlag.**
//
// # Warum das Schema dreimal gelesen wird
//
// **Weil drei Stellen es brauchen und keine der anderen traut**: das
// Auftragswerk friert beim **Anlegen** seine Kennung ein, die Naht baut
// daraus die **Form** der Frage, und das Werk prüft den Vorschlag noch einmal
// dagegen. Fällt es dazwischen aus, gibt es keinen Vorschlag — und nicht
// einen, den niemand geprüft hat.
func TestEinSpaetUnlesbaresPaketGibtKeinenVorschlag(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.anwendung.MitStilwahl(&abDemDrittenStolpernd{felder: welt.felder})

	if antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "x", ""); antwort.Fehler == nil {
		t.Fatal("das spät unlesbare Paket kam durch")
	}
}

// abDemDrittenStolpernd antwortet zweimal und scheitert beim dritten Ruf.
type abDemDrittenStolpernd struct {
	app.Stilwaehler

	felder  []domain.Field
	gelesen int
}

func (a *abDemDrittenStolpernd) Kartenfelder(context.Context, int64) (
	[]domain.Field, error,
) {
	a.gelesen++

	if a.gelesen > 2 {
		return nil, errors.New("das Paket ist nicht mehr zu lesen")
	}

	return a.felder, nil
}
