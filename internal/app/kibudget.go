package app

// **Was ein KI-Auftrag kosten darf** — Werkstück A5 von Phase 5.3
// (#1444)
// und A8 (#1447),
// abgetrennt mit A9
// (#1448),
// weil `kiwerkzeug.go` über die Zeilengrenze aus §2 ging.
//
// # Der Schnitt folgt der Frage
//
// Dort steht, **wer** antwortet und **womit** — die Anbieternamen, die
// Quelle, das Werk. Hier steht, **was es kosten darf**.

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// KostenpostenOhneKosten steht, wo nichts zu reservieren war.
//
// # Warum überhaupt etwas dasteht
//
// **Weil A6 einen Kostenposten verlangt** ([domain.KIAuftrag]) — und das aus
// gutem Grund: ein Auftrag ohne Buchung sähe aus wie einer, dessen Buchung
// jemand vergessen hat.
//
// **Und das hier ist keine Buchung.** Es ist die ausdrückliche Aussage, dass
// keine nötig war: ein Modell auf diesem Rechner kostet kein Geld. Für ein
// Modell, das anderswo rechnet, steht hier der Posten der
// [service.Kostenwache].
const KostenpostenOhneKosten = "ohne-kosten-lokal"

// kibudget bucht, was zu buchen ist.
//
// # Zwei Wege, und der zweite kostet Geld
//
// **Lokal**: nichts zu buchen, und der Posten sagt es.
//
// **Anderswo**: das Kostenjournal aus O12 (#1439)
// reserviert **vor** dem Ruf — „wer erst nach der Antwort bucht, hat bei zwei
// gleichzeitigen Aufträgen zweimal die Grenze gelesen und keinmal gehalten".
//
// **Und es verlangt den Nachweis eines Menschen.** Das ist keine Erfindung
// dieses Werkstücks: `Kostenwache.Reservieren` weist einen Auftrag ohne
// Nachweis ab, und zwar mit den Worten „ein Auftrag, der Geld kostet, braucht
// die ausdrückliche Freigabe eines Menschen". Das Issue verlangt dasselbe von
// der anderen Seite: „Unter Local-only, fehlendem Scope oder **Budget**
// entstehen null Anbieteranfragen."
type kibudget struct {
	ort      domain.KIOrt
	anbieter string

	// wache ist das Kostenjournal — nil heißt: es gibt keines, und dann geht
	// kein kostenpflichtiger Auftrag hinaus.
	//
	// **Die Schnittstelle und nicht der Verbund**: [Kostenfuehrer] ist, was
	// das Fenster ohnehin hält, und ein Lauf kann sie einsetzen (§3).
	wache Kostenfuehrer

	// nachweis ist die Freigabe eines Menschen — leer heißt: keine.
	nachweis string

	// grenze ist der Höchstbetrag, den dieser Auftrag kosten darf.
	grenze domain.Betrag

	// gedeckt ist der Posten eines Batches, der schon gebucht hat — leer
	// heißt: dieser Auftrag bucht selbst
	// ([0105](../../docs/entscheidungen/0105-eine-freigabe-fuer-einen-batch.md)).
	//
	// **Nicht [KostenpostenOhneKosten]**, obwohl beides „bucht nicht" heißt:
	// jene Marke sagt „lokal, kostet nichts", und das ist hier das Gegenteil
	// dessen, was zutrifft. Zwei Sachverhalte unter einer Marke sind zwei
	// Auskünfte, von denen eine falsch ist.
	gedeckt string
}

// Reservieren meldet einen Auftrag an.
func (b kibudget) Reservieren(ctx context.Context, kennung string) (string, error) {
	if b.ort == domain.KIOrtLokal {
		return KostenpostenOhneKosten, nil
	}

	// **Was der Batch gebucht hat, bucht der Auftrag darin nicht noch
	// einmal** (0105). Er bekommt denselben Posten zurück — an ihm hängt
	// der Abschluss, und den macht der Batch.
	if b.gedeckt != "" {
		return b.gedeckt, nil
	}

	if b.wache == nil {
		return "", fmt.Errorf(
			"%w: für %q gibt es kein Kostenjournal — ohne eines geht kein "+
				"Auftrag hinaus, der Geld kostet", domain.ErrValidation, b.anbieter)
	}

	if b.nachweis == "" {
		return "", fmt.Errorf(
			"%w: ein Auftrag an %q kostet Geld und braucht die ausdrückliche "+
				"Freigabe eines Menschen", domain.ErrErlaubnis, b.anbieter)
	}

	posten, fehler := b.wache.Reservieren(ctx, service.Kostenauftrag{
		Kennung: kennung, Konto: b.anbieter, Aufgabe: KIAufgabe,
		Hoechstbetrag: b.grenze,
	}, b.nachweis)
	if fehler != nil {
		return "", fehler
	}

	return posten.Kennung, nil
}

// kiPostenSchliessen schließt den Posten, den ein KI-Auftrag angelegt hat.
//
// # Warum **unklar** und nicht **abgeschlossen**
//
// **Weil `Abschliessen(…, 0)` „war kostenlos" hieße, und das ist falsch** —
// #1542.
// Der Anbieter hat die Anfrage angenommen und rechnet ab; **was** sie gekostet
// hat, sagt er an dieser Stelle nicht, und dieses Programm führt keine
// Preisliste. O12 hat den Stand genau dafür: „zählt wie ausgegeben, bis jemand
// nachgesehen hat" (#1439).
//
// # Warum überhaupt
//
// **Bis zum 2026-09-13 blieb jede Reservierung offen.** [service.KIBudget]
// hatte nur `Reservieren`, keine der drei Abschlussarten wurde je gerufen, und
// `gebundenIm` zählt einen reservierten Posten mit — eine Grenze, die nur noch
// steigt, bis kein Auftrag mehr hinausgeht. Gefunden in der Sichtprüfung am
// laufenden Fenster; kein Lauf sah es, weil jeder ein Doppel am Journal
// einsetzt (§3).
//
// # Was der Posten entscheidet und nicht der Ort
//
// **Ein lokaler Auftrag trägt [KostenpostenOhneKosten]** — er hat nichts
// gebucht, und die Wache nach einem Posten zu fragen, den es nicht gibt, gäbe
// eine Meldung über etwas, das nichts gekostet hat. Die Marke ist dafür die
// genauere Frage als der Ort: sie steht an **diesem** Auftrag.
//
// # Und warum der Fehler nur ins Protokoll geht
//
// **Weil der Auftrag geschehen ist** (§7, Regel 3). Wer die Meldung an die
// Antwort hängte, machte aus einem geglückten Vorschlag einen gescheiterten —
// wegen einer Buchung, die niemand mehr rückgängig machen kann.
func (a *App) kiPostenSchliessen(ctx context.Context, posten string) {
	if posten == "" || posten == KostenpostenOhneKosten || a.kosten == nil {
		return
	}

	if _, fehler := a.kosten.AlsUnklarMerken(
		ctx, posten, KIUnklarGrund); fehler != nil {
		// **Der Satz kurz und in einem Stück** — der Wächter über die
		// Merkmale liest ihn als Literal, und eine zusammengesetzte Meldung
		// lässt sich nicht gegen die angemeldeten Namen halten.
		slog.Default().ErrorContext(ctx, "ein Kostenposten blieb offen",
			"posten", posten, "grund", fehler.Error())
	}
}

// kiPostenFreigeben nimmt eine Reservierung zurück, für die **nichts**
// hinausgegangen ist.
//
// **Der Unterschied zu [App.kiPostenSchliessen] ist eine Aussage** (#1542):
// dort hat der Anbieter angenommen und der Preis ist unbekannt; hier steht
// fest, dass nie gefragt wurde — das Werk kam gar nicht zustande. „Unklar"
// wäre dann eine Zahl, die nie entstanden ist.
func (a *App) kiPostenFreigeben(ctx context.Context, posten string) {
	if posten == "" || posten == KostenpostenOhneKosten || a.kosten == nil {
		return
	}

	if _, fehler := a.kosten.Freigeben(ctx, posten); fehler != nil {
		slog.Default().ErrorContext(ctx, "ein Kostenposten blieb offen",
			"posten", posten, "grund", fehler.Error())
	}
}

// KIUnklarGrund steht an jedem Posten, den ein KI-Auftrag hinterlässt.
//
// **Er nennt den Grund und nicht den Vorgang**: wer ihn liest, will wissen,
// warum die Zahl fehlt — nicht, welche Funktion gelaufen ist.
const KIUnklarGrund = "der Anbieter hat die Anfrage angenommen; was sie " +
	"gekostet hat, nennt er hier nicht"

// KIAufgabe ist der Name, unter dem KI-Aufträge im Kostenjournal stehen.
//
// **Einer für alle Anbieter.** Das Konto trennt sie (`anthropic`, später
// `openai`); die **Aufgabe** sagt, wofür das Geld ausgegeben wurde, und das
// ist bei allen dieselbe.
const KIAufgabe = "ki-vorschlag"

// budgetFuer baut das Budget eines Anbieters.
//
// **Der Nachweis und die Grenze kommen von außen** ([App.MitKIFreigabe]):
// eine Freigabe gilt **einem** Auftrag, und wer sie am Fenster gibt, gibt sie
// für das, was er gerade gelesen hat.
// **Der Nachweis kommt als Parameter und nicht aus einem Feld**
// (#1539).
// Aus einem Feld gelesen, das nur [App.MitKIFreigabe] setzt — und die ruft
// **kein Produktivcode** —, scheiterte am Fenster jeder kostenpflichtige
// Auftrag.
//
// **Leer heißt: die des Fensters**, falls eine gesetzt ist — `MitKIFreigabe`
// bleibt für Läufe, die keinen Nachweis ausstellen wollen.
//
// # Und der Höchstbetrag wird gelesen, nicht geraten
//
// **Er steht im Nachweis.** Aus einem Feld, das dieselbe ungerufene Funktion
// setzt, käme am Fenster ein leerer Betrag, und die Buchung scheiterte an
// „Eine Währung hat drei Zeichen" — eine Meldung über eine Währung, wo eine
// Freigabe fehlt.
//
// **Gelesen und nicht mitgereicht**, und der Grund steht bei
// [service.Nachweise.Grenze]: ein Aufrufer, der den Betrag nennte, nennte
// einen kleineren und unterliefe die Tagesgrenze.
func (a *App) budgetFuer(
	ctx context.Context, anbieter, adresse string, deckung kideckung,
) (kibudget, error) {
	budget := kibudget{
		ort:      ortVon(anbieter, adresse),
		anbieter: anbieter,
		wache:    a.kosten,
		nachweis: deckung.nachweis,
		grenze:   a.kigrenze,
		gedeckt:  deckung.posten,
	}

	// **Ein gedeckter Auftrag fragt nach keiner Grenze** (0105): der Batch
	// hat gebucht, und was er buchen durfte, stand in **seinem** Nachweis.
	if deckung.posten != "" {
		return budget, nil
	}

	// **Ohne Nachweis am Ruf die des Fensters — samt ihrer Grenze.**
	// `MitKIFreigabe` setzt beides zusammen, und ein Lauf, der so setzt, hat
	// keine Ablage, in der nachzusehen wäre.
	if deckung.nachweis == "" {
		budget.nachweis = a.kinachweis

		return budget, nil
	}

	if budget.ort == domain.KIOrtLokal || a.autorisierung == nil {
		return budget, nil
	}

	grenze, fehler := a.autorisierung.Nachweisgrenze(ctx, deckung.nachweis)
	if fehler != nil {
		return kibudget{}, fehler
	}

	budget.grenze = domain.Betrag{
		Wert: grenze.Betrag, Waehrung: grenze.Waehrung,
	}

	return budget, nil
}

// kideckung sagt, **wie** ein Auftrag bezahlt wird.
//
// **Zwei Felder und nicht zwei Parameter**: sie schließen einander aus —
// entweder trägt der Auftrag seine eigene Freigabe, oder er läuft unter der
// eines Batches (0105). Als Verbund steht das da; als zwei Zeichenketten
// nebeneinander müsste jeder Aufrufer es wissen.
type kideckung struct {
	// nachweis ist die Freigabe eines Menschen — leer heißt: keine.
	nachweis string

	// posten ist der Posten eines Batches, der schon gebucht hat.
	posten string
}

// MitKIFreigabe setzt Nachweis und Höchstbetrag für den nächsten Auftrag.
//
// # Warum das nicht im Auftragswerk steht
//
// **Weil eine Freigabe zu einem Ruf gehört und nicht zu einer Sitzung.** Sie
// wird gesetzt, verbraucht und fällt weg; ein Werk, das sie hielte, gäbe sie
// dem nächsten Auftrag mit, den niemand freigegeben hat.
func (a *App) MitKIFreigabe(nachweis string, grenze domain.Betrag) *App {
	a.kinachweis = nachweis
	a.kigrenze = grenze

	return a
}
