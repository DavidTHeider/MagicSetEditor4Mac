package app_test

// **Ein Mensch kann einen kostenpflichtigen Auftrag freigeben** —
// #1539.
//
// # Der Befund
//
// **Am laufenden Fenster ging kein einziger.** Gemessen an der echten
// Wails-Brücke: jeder Ruf an einen Cloudanbieter endete mit „braucht die
// ausdrückliche Freigabe eines Menschen", und **keine Bindung stellte eine
// aus**. `MitKIFreigabe` setzt sie — und wird von keinem Produktivcode
// gerufen.
//
// **Kein Lauf sah es**, weil jeder sie selbst setzte. Die Vorrichtung war
// großzügiger als der Betrieb (§3), und keiner fragte, ob es im Betrieb
// überhaupt einen Weg dorthin gibt.
//
// # Was hier geprüft wird
//
// **Der Weg von der Freigabe bis zum Auftrag** — und nicht die Freigabe für
// sich. Genau diese Naht hat gefehlt.

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// cloudwelt ist eine Anwendung, deren Anbieter Geld kostet.
type cloudwelt struct {
	anwendung *app.App
	karte     int64

	// fremd ist die Laufzeit der Wolke — ein Lauf, der eine andere Antwort
	// braucht, setzt sie hier.
	fremd *laufzeitdoppel

	// karten ist der Bestand dieser Welt — ein Lauf zählt dort nach, was
	// wirklich geschrieben wurde.
	karten *kartendoppel

	// db ist der Bestand — ein Lauf, der einen Schalter umlegt, braucht ihn.
	db *sql.DB
}

// einecloudwelt baut sie — mit Karte, eingeschaltetem Anbieter und einem
// Modell, das seine Fähigkeit belegt.
//
// **Über [einekiwelt] und nicht über [zweianbieter]**: der Vorschlag liest
// die Zielkarte, bevor er bucht, und eine Welt ohne Karte scheiterte an
// „nicht gefunden" statt an der Freigabe — ein Lauf, der das Richtige meldet
// und das Falsche prüft.
func einecloudwelt(t *testing.T) cloudwelt {
	t.Helper()

	welt := einekiwelt(t)

	// **Eine eigene Laufzeit für die Wolke.** Dieselbe zweimal einzuhängen
	// hieße, dass der lokale Anbieter das Cloudmodell führt — und dann
	// entschiede nicht der Anbietername, wer antwortet.
	fremd := &laufzeitdoppel{
		fassung: "anthropic-version 2023-06-01",
		modelle: []domain.KIModell{einCloudmodell()},
		antwort: app.KIRohantwort{
			Inhalt: `{"name": "Blitzschlag", "rarity": "rare"}`,
			Modell: derCloudalias,
		},
	}

	welt.anwendung.MitKIAnbietern(
		func(string) app.KILaufzeitquelle { return welt.laufzeit },
		map[string]func(adresse string) app.KILaufzeitquelle{
			derCloudanbieter: func(string) app.KILaufzeitquelle { return fremd },
		})

	if fehler := storage.NeueEinstellungen(welt.db).Setzen(probekontext(t),
		derSchalter(), domain.DienstAn); fehler != nil {
		t.Fatalf("den Anbieter einschalten: %v", fehler)
	}

	// **Und ein Budget, über die Bindung des Fensters.** Zwei Tore stehen
	// vor einem bezahlten Auftrag, und sie sind nicht dasselbe: das Budget
	// ist die Grenze **eines Zeitraums**, die Freigabe gilt **einem
	// Auftrag**. Wer nur eines setzte, prüfte hier das andere nicht.
	if antwort := welt.anwendung.KostenbudgetSetzen(probekontext(t),
		derCloudanbieter, string(domain.ZeitraumTag), 10_000,
		dieWaehrung); antwort.Fehler != nil {
		t.Fatalf("das Budget setzen: %+v", antwort.Fehler)
	}

	return cloudwelt{
		anwendung: welt.anwendung, karte: welt.karte, fremd: fremd,
		karten: welt.karten, db: welt.db,
	}
}

// derCloudanbieter ist, wer hier Geld kostet.
//
// **Und er ist zugleich das Ziel der Freigabe.** Ein Nachweis für einen
// anderen Anbieter trägt diesen Auftrag nicht — das hält
// [domain.Nachweis.Gilt], und deshalb steht der Name **einmal** hier.
const derCloudanbieter = "anthropic"

// dieWaehrung ist die, in der hier gerechnet wird.
//
// **Einmal und an einer Stelle**: Budget und Freigabe müssen dieselbe nennen
// — sonst sagt [domain.Kostengrenze] „die Grenze gilt in X, die Kosten fallen
// in Y an", und das wäre ein Befund über diesen Lauf und nicht über den Code.
const dieWaehrung = "EUR"

// vorschlagenMitFreigabe fragt das Cloudmodell — mit dieser Freigabe.
func (c cloudwelt) vorschlagenMitFreigabe(
	t *testing.T, nachweis string,
) app.KIVorschlagAntwort {
	t.Helper()

	return c.anwendung.KIVorschlagen(probekontext(t), c.karte, derCloudanbieter,
		derCloudalias, "Schreibe einen Regeltext.", nachweis)
}

// **Ohne Freigabe scheitert ein kostenpflichtiger Auftrag — mit dem Satz.**
func TestOhneFreigabeScheitertEinBezahlterAuftrag(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)

	antwort := welt.vorschlagenMitFreigabe(t, "")

	if antwort.Fehler == nil {
		t.Fatal("der Auftrag ging ohne Freigabe hinaus")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "Freigabe") {
		t.Errorf("die Meldung ist %q", antwort.Fehler.Meldung)
	}
}

// **Und mit einer, die ein Mensch am Fenster erteilt hat, geht er durch.**
//
// **Das ist die Naht, die gefehlt hat**: erteilen und einlösen, in einem Lauf.
func TestMitEinerErteiltenFreigabeGehtDerAuftragDurch(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)

	freigabe := welt.anwendung.KIFreigabeErteilen(probekontext(t),
		derCloudanbieter, 500, dieWaehrung)
	if freigabe.Fehler != nil {
		t.Fatalf("KIFreigabeErteilen: %+v", freigabe.Fehler)
	}

	if freigabe.Kennung == "" || freigabe.Laeuftab == "" {
		t.Fatalf("die Freigabe ist %+v", freigabe)
	}

	if antwort := welt.vorschlagenMitFreigabe(t,
		freigabe.Kennung); antwort.Fehler != nil {
		t.Fatalf("mit Freigabe: %+v", antwort.Fehler)
	}
}

// **Eine Freigabe gilt einem Auftrag — der zweite braucht einen zweiten.**
//
// **Sonst wäre sie eine Vollmacht** (0078). Die Zusage hält
// `Nachweise.Einloesen`; hier steht, dass sie am Fenster ankommt.
func TestEineFreigabeGiltNurEinemAuftrag(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)

	freigabe := welt.anwendung.KIFreigabeErteilen(probekontext(t),
		derCloudanbieter, 500, dieWaehrung)
	if freigabe.Fehler != nil {
		t.Fatalf("KIFreigabeErteilen: %+v", freigabe.Fehler)
	}

	if antwort := welt.vorschlagenMitFreigabe(t,
		freigabe.Kennung); antwort.Fehler != nil {
		t.Fatalf("der erste Auftrag: %+v", antwort.Fehler)
	}

	if antwort := welt.vorschlagenMitFreigabe(t,
		freigabe.Kennung); antwort.Fehler == nil {
		t.Fatal("dieselbe Freigabe trug einen zweiten Auftrag")
	}
}

// **Und eine erfundene Kennung trägt nichts.**
func TestEineErfundeneFreigabeTraegtNichts(t *testing.T) {
	t.Parallel()

	welt := einecloudwelt(t)

	if antwort := welt.vorschlagenMitFreigabe(t,
		"gibtesnicht"); antwort.Fehler == nil {
		t.Fatal("die erfundene Freigabe ging durch")
	}
}

// **Ohne Bestand gibt es keine Freigabe — und es steht dran, warum.**
//
// **Und nicht „unbekannter Fehler"**: wer ein Fenster ohne offenen Bestand
// vor sich hat, soll den Bestand öffnen und nicht nach einer Freigabe suchen
// (§8).
func TestOhneBestandGibtEsKeineFreigabe(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").KIFreigabeErteilen(probekontext(t),
		derCloudanbieter, 500, dieWaehrung)
	if antwort.Fehler == nil {
		t.Fatal("ohne Bestand kam eine Freigabe")
	}

	if antwort.Kennung != "" {
		t.Errorf("die Kennung ist %q", antwort.Kennung)
	}
}

// **Ein lokaler Auftrag fragt nicht nach der Grenze einer Freigabe.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil er sonst scheiterte.** Wer eine Kennung mitgibt und ein Modell auf
// diesem Rechner fragt, hat keinen Nachweis ausgestellt — die Grenze nachzu-
// schlagen gäbe „gibt es nicht", und ein Auftrag, der nichts kostet, bräche
// an einer Freigabe, die niemand braucht.
func TestEinLokalerAuftragFragtNichtNachDerGrenze(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	antwort := welt.anwendung.KIVorschlagen(probekontext(t), welt.karte,
		"ollama", "gemma4:8b", "Schreibe einen Regeltext.", "gibtesnicht")
	if antwort.Fehler != nil {
		t.Fatalf("der lokale Auftrag: %+v", antwort.Fehler)
	}
}
