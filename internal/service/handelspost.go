package service

// **Was mit einer angekommenen Handelsnachricht geschieht** — Werkstück `T5`
// von Phase 5.6
// (#1470).
//
// # Warum der Absender nicht geglaubt wird
//
// **Weil er im Umschlag steht und der Umschlag von drüben kommt.** Wer
// wirklich spricht, sagt die Leitung: sie hat die Peeridentität geprüft, als
// die Verbindung entstand (N2, #1455).
// Gehen beide auseinander, endet es hier — und zwar **bevor** irgendetwas
// nachgeschlagen wird, denn schon ein Nachschlag verriete, dass es den
// Vorgang gibt.
//
// # Warum das nicht durch [Handelsleitung] geht
//
// **Wegen der einen Transaktion.** `Handelsleitung.Vorschlagen` schreibt die
// Revision für sich; hier müssen Vermerk und Wirkung **zusammen** stehen oder
// zusammen fallen, sonst führte die nächste Zustellung dieselbe Wirkung noch
// einmal aus. Die **Regel** ist dieselbe — beide fragen
// [domain.Handelsangebot.Nimmt]; verschieden ist nur, wer schreibt.

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Briefkasten ist, was die Post vom Bestand braucht.
type Briefkasten interface {
	// Nachschlagen gibt den Eintrag zu einem Schlüssel — **oder `nil`**.
	Nachschlagen(ctx context.Context, schluessel string) (*domain.Inboxeintrag, error)

	// Zustellen schreibt Vermerk, Revision und die Wirkung an den Zusagen
	// in **einem** Zug.
	//
	// **Die dritte gehört seit dem Nachlauf zum Zweitblick über Phase 5.6
	// dazu**: geschah sie vorher für sich, blieb nach einem gescheiterten
	// Schreiben eine Zusage als „übergeben" stehen, während der Vorgang
	// noch auf `vorbereitet` stand.
	//
	// **Und nur gegen den Stand, gegen den entschieden wurde** (Befund Z1
	// des Zweitblicks über Phase 6.8): `vorher` ist, was [Handelspost]
	// gelesen hat; sonst [domain.ErrUeberholt], und die Nachricht bleibt im
	// Ausgang drüben bis zur nächsten Nachlieferung (N4 des Nachlaufs).
	Zustellen(ctx context.Context, eintrag domain.Inboxeintrag,
		vorher, angebot domain.Handelsangebot,
		wirkung domain.Zusagenwirkungsart) error
}

// Handelsstand gibt den heutigen Stand eines Vorgangs.
//
// **Eine eigene Schnittstelle und nicht [Handelsablage]**: die Post liest den
// Stand und schreibt ihn nicht — sie schreibt über den Briefkasten. Eine
// Schnittstelle, die mehr kann, als der Aufrufer braucht, lädt dazu ein, den
// Weg daneben zu nehmen.
type Handelsstand interface {
	Stand(ctx context.Context, vorgang string) (domain.Handelsangebot, error)
}

// Handelsnachricht ist, was ankommt: ein Umschlag und sein Angebot.
type Handelsnachricht struct {
	Umschlag domain.Handelsumschlag

	// Angebot ist die Revision, die diese Nachricht bewirkt.
	Angebot domain.Handelsangebot

	// Beleg ist der unterschriebene Abschluss — **nur** bei
	// [domain.NachrichtBeleg], und auch dort wahlfrei
	// (#1650).
	//
	// **Wahlfrei, weil eine fehlende Urkunde kein Fehler ist**: die
	// Nachricht sagt den Zustand, und der steht auch ohne sie. Was dann
	// fehlt, ist die **Deckung** — und ohne Deckung bleibt `beidseitig`
	// falsch, statt behauptet zu werden.
	Beleg *domain.Belegurkunde `json:",omitempty"`

	// Sicht und Absage sind die Antwort auf eine Sichtanfrage
	// (`handelssicht.go`) — **nur** bei [domain.NachrichtSicht].
	Sicht  []domain.Sichtkarte `json:",omitempty"`
	Absage string              `json:",omitempty"`
}

// Zustellungsantwort sagt, was geschehen ist.
type Zustellungsantwort struct {
	Art domain.Zustellungsart

	// Ergebnis ist die Antwort — **bei einer Wiederholung die von vorhin**.
	Ergebnis string
}

// Handelspost nimmt Handelsnachrichten entgegen.
type Handelspost struct {
	briefkasten Briefkasten
	stand       Handelsstand
	kontakte    Kontaktnachschlag

	schranke Handelsschranke // #2435, `handelsschranke.go`

	// ausgang ist wahlfrei — ohne ihn wird nur empfangen (siehe
	// [Handelspost.MitAusgang]).
	ausgang Ausgang

	// gegenbelege ist wahlfrei — ohne sie wird ein angekommener Beleg
	// nicht geprüft und nicht abgelegt
	// ([Handelspost.MitGegenbelegen], #1650).
	gegenbelege Gegenbelegstelle

	// klaerung ist wahlfrei — ohne sie bleibt ein Konflikt eine Antwort
	// ([Handelspost.MitKlaerung], TRADE-03).
	klaerung Klaerungsstelle

	ausweis Ausweisstelle // wahlfrei, [Handelspost.MitAusweis]

	sichtgeber Sichtgeber // wahlfrei, [Handelspost.MitSichtgeber] (P11)
}

// NeueHandelspost stellt sie auf.
func NeueHandelspost(
	briefkasten Briefkasten, stand Handelsstand, kontakte Kontaktnachschlag,
) *Handelspost {
	return &Handelspost{
		briefkasten: briefkasten, stand: stand, kontakte: kontakte,
		schranke: ohneEinstellung{},
	}
}

// Empfangen behandelt eine angekommene Nachricht.
//
// **`sender` ist, wer wirklich spricht** — die Leitung hat ihn geprüft. Was
// im Umschlag steht, ist eine Behauptung.
func (p *Handelspost) Empfangen(
	ctx context.Context, sender string, nachricht Handelsnachricht,
) (Zustellungsantwort, error) {
	if fehler := p.durchgelassen(ctx, sender, nachricht); fehler != nil {
		return Zustellungsantwort{}, fehler
	}

	// **Das Profil steht nicht im Schlüssel** — der Bestand klemmt jede
	// Wurzelabfrage ohnehin darauf (0066).
	schluessel := nachricht.Umschlag.Idempotenzschluessel()

	vorher, fehler := p.briefkasten.Nachschlagen(ctx, schluessel)
	if fehler != nil {
		return Zustellungsantwort{}, fehler
	}

	abdruck := nachricht.Angebot.Abdruck()

	// **Drei Antworten, drei Wege.** [domain.Zustellung] gibt genau drei:
	// `neu` führt aus, `bekannt` und `konflikt` geben die Antwort von
	// vorhin — und ein Konflikt in der Übergabe führt dazu in die Klärung
	// (TRADE-03.AK-04, `handelspostklaerung.go`).
	if vorher == nil {
		return p.ausfuehren(ctx, sender, nachricht, schluessel, abdruck)
	}

	return p.schonVermerkt(ctx, sender, nachricht, vorher, abdruck)
}

// durchgelassen prüft, was vor jedem Nachschlag stimmen muss.
//
// **In dieser Reihenfolge.** Der Absender zuerst, weil ein Nachschlag schon
// verriete, dass es den Vorgang gibt; dann der Zweck, dann die Gestalt.
func (p *Handelspost) durchgelassen(
	ctx context.Context, sender string, nachricht Handelsnachricht,
) error {
	if sender == "" || sender != nachricht.Umschlag.Absender {
		return domain.MitGrund(domain.ErrErlaubnis, nil,
			"Diese Nachricht nennt einen anderen Absender, als wer spricht.")
	}

	kontakt, fehler := p.kontakte.Lesen(ctx, sender)
	if fehler != nil {
		return domain.MitGrund(domain.ErrErlaubnis, fehler,
			"Dieses Gegenüber darf hier nicht handeln.")
	}

	if !kontakt.Darf(domain.ZweckHandel) {
		return domain.MitGrund(domain.ErrErlaubnis, nil,
			"Dieses Gegenüber darf hier nicht handeln.")
	}

	// **Handeln dürfen heißt nicht: an diesem Handel** (#1634).
	// [domain.Handelsangebot.Nimmt] hält den Kopf fest — Parteien
	// eingeschlossen —, und genau deshalb kam ein Dritter mit einem
	// unveränderten Kopf durch und änderte die **Positionen**. Und seit die
	// erste Revision hier entstehen darf, legte er damit sogar Vorgänge an,
	// die diesen Rechner nichts angehen.
	if !slices.Contains(nachricht.Angebot.Parteien, sender) {
		return domain.MitGrund(domain.ErrErlaubnis, nil,
			"Dieses Gegenüber ist an diesem Handel nicht beteiligt.")
	}

	if befunde := nachricht.Umschlag.Validate(); len(befunde) > 0 {
		return befunde[0]
	}

	return p.zusammengehoerig(nachricht)
}

// zusammengehoerig prüft, dass Umschlag und Angebot dieselbe Sache meinen.
//
// **Sonst unterschriebe jemand einen Abdruck und schickte ein anderes
// Angebot.** Die Unterschrift gilt dem Umschlag; was sie bindet, muss das
// sein, was danebenliegt.
func (p *Handelspost) zusammengehoerig(nachricht Handelsnachricht) error {
	if nachricht.Umschlag.Vorgang != nachricht.Angebot.Vorgang {
		return domain.NewFieldError("vorgang", domain.SchluesselValidierung,
			"Der Umschlag nennt den Vorgang %q, das Angebot %q.",
			nachricht.Umschlag.Vorgang, nachricht.Angebot.Vorgang)
	}

	if nachricht.Umschlag.Abdruck != nachricht.Angebot.Abdruck() {
		return domain.NewFieldError("abdruck", domain.SchluesselValidierung,
			"Der Abdruck im Umschlag gehört nicht zu diesem Angebot.")
	}

	return nil
}

// ausfuehren wendet eine neue Nachricht an.
func (p *Handelspost) ausfuehren(
	ctx context.Context, sender string, nachricht Handelsnachricht,
	schluessel, abdruck string,
) (Zustellungsantwort, error) {
	// **Was geschrieben wird, ist nicht immer, was ankommt** (`T12`,
	// #1477):
	// eine Nachricht zu **derselben** Revision bringt Stimmen, und was
	// daraus wird, entscheidet [domain.Handelsangebot.NimmtStand] — hier
	// gegen den **hiesigen** Stand.
	wirkung, heute, vorher, fehler := p.anschluss(ctx, sender, nachricht.Angebot)
	if fehler != nil {
		return Zustellungsantwort{}, fehler
	}

	// **Dieselbe Schranke wie an der Leitung** (#2435).
	if fehler := einigungErlaubt(ctx, p.schranke, wirkung); fehler != nil {
		return Zustellungsantwort{}, fehler
	}

	eintrag := domain.Inboxeintrag{
		Schluessel: schluessel, Nutzlastabdruck: abdruck,
		Ergebnis: fmt.Sprintf("%s/%d", wirkung.Vorgang, wirkung.Revision),
	}

	// **Ein Wechsel über die Leitung bedeutet dasselbe wie einer von Hand**
	// (Befund 3 des Zweitblicks über Phase 5.6) — und er steht in
	// **derselben** Transaktion wie der Vermerk (Befund 3 des Nachlaufs):
	// entweder beides oder keines.
	if fehler := p.briefkasten.Zustellen(ctx, eintrag, heute, wirkung,
		domain.ZusagenwirkungFuer(vorher, wirkung.Zustand)); fehler != nil {
		return p.nachDemWettlauf(ctx, sender, nachricht, abdruck, fehler)
	}

	// **Und danach die Urkunde** (#1650). **Danach und nicht davor**: der
	// Vermerk sagt, dass diese Nachricht behandelt ist; eine Gegenzeichnung
	// davor stünde da, während die Nachricht noch als unbehandelt gälte,
	// und der nächste Anlauf zeichnete ein zweites Mal gegen.
	if fehler := p.gegenbeleg(ctx, sender, nachricht, wirkung); fehler != nil {
		return Zustellungsantwort{}, fehler
	}

	return Zustellungsantwort{
		Art: domain.ZustellungNeu, Ergebnis: eintrag.Ergebnis,
	}, nil
}

// anschluss prüft, ob diese Revision an den hiesigen Stand anschließt.
//
// # Warum ein unbekannter Vorgang kein Fehler ist
//
// **Weil jeder erste es ist** (#1634).
// Wem ein Handel angeboten wird, der kennt ihn noch nicht — und bis zum
// 2026-09-19 endete die Zustellung genau hier, weil der Nachschlag „gibt es
// nicht" gab. Gesehen hat das erst der Lauf über einen **echten Draht**
// zwischen zwei Prozessen; die Läufe davor stellten beide Seiten im selben
// Prozess auf, und dort hatte der Empfänger den Vorgang schon.
//
// **Niemand kommt dadurch herein**: wer spricht, ist längst geprüft, und er
// darf handeln ([Handelspost.durchgelassen]). Neu ist allein, dass eine
// **erste** Revision hier entstehen darf — eine spätere ohne Vorgeschichte
// bleibt ein Konflikt, sonst begänne ein Vorgang bei Revision sieben, deren
// sechs Vorgänger niemand gesehen hat.
//
// **`heute` geht mit hinaus** (leer ohne Vorgang): gegen ihn wird
// geschrieben (Befund Z1 des Zweitblicks über Phase 6.8).
func (p *Handelspost) anschluss(
	ctx context.Context, sender string, angebot domain.Handelsangebot,
) (wirkung, heute domain.Handelsangebot, vorher domain.Handelszustand, fehler error) {
	heute, fehler = p.stand.Stand(ctx, angebot.Vorgang)

	switch {
	case errors.Is(fehler, domain.ErrNotFound):
		if befund := angebot.Pruefe(); befund != nil {
			return domain.Handelsangebot{}, domain.Handelsangebot{}, "", befund
		}

		// **Der Grund bleibt draußen, obwohl einer dasteht.**
		// [domain.MitGrund] hängt ihn in die `Unwrap`-Kette, und ein
		// [domain.ErrNotFound] darin machte diesen Konflikt **zugleich** zu
		// einem Nichtgefunden — wer beides prüft, nähme den Zweig, der
		// zuerst steht.
		if angebot.Revision != 1 {
			return domain.Handelsangebot{}, domain.Handelsangebot{}, "", domain.MitGrund(
				domain.ErrConflict, nil,
				"diesen Handelsvorgang gibt es hier nicht")
		}

		// **Und sie behauptet nicht mehr, als ihr Absender sagen darf**
		// (TRADE-03, #1717): vom Initiator, angeboten, nur seine Stimme.
		if befund := domain.PruefeErsteRevision(angebot, sender); befund != nil {
			return domain.Handelsangebot{}, domain.Handelsangebot{}, "", befund
		}

		if befund := p.istPartei(ctx, angebot); befund != nil {
			return domain.Handelsangebot{}, domain.Handelsangebot{}, "", befund
		}

		return angebot, domain.Handelsangebot{}, angebot.Zustand, nil

	case fehler != nil:
		return domain.Handelsangebot{}, domain.Handelsangebot{}, "", fehler
	}

	// **Dieselbe Revision bringt Stimmen und keinen Vorschlag** (`T12`).
	// Bis zum 2026-09-19 endete sie hier mit „der Vorschlag ist Revision 1,
	// die nächste wäre 2" — und zwei Prozesse kamen nie über `angeboten`
	// hinaus, weil jeder nur seine eigene Stimme hatte.
	if angebot.Revision == heute.Revision {
		genommen, befund := heute.NimmtStand(angebot, sender)

		return genommen, heute, heute.Zustand, befund
	}

	// **Dieselbe Regel wie beim eigenen Vorschlag** (T1): ein Vorschlag gegen
	// eine Revision, die weitergezogen ist, wird abgewiesen — mit
	// [domain.ErrConflict], damit der Klient noch einmal liest, statt in
	// einer Schleife zu wiederholen.
	//
	// **Und ein Vorschlag behauptet keinen Zustand und keine fremde Stimme**
	// (TRADE-03, #1717) — beides entging `Nimmt`, weil es nicht im Abdruck
	// steht.
	return angebot, heute, heute.Zustand, heute.NimmtVon(angebot, sender)
}

// nachDemWettlauf liest nach, wenn eine zweite Zustellung schneller war.
//
// # Der Fund, für den es das gibt
//
// **Die Ablage fängt den zweiten Zug am Primärschlüssel**, und ihr eigener
// Kopf sagt: „der Aufrufer liest danach nach, was dasteht — statt zu raten."
// Der Aufrufer las nicht nach; er reichte den Schreibfehler durch. Zwei
// gleiche Nachrichten, die sich im Fenster zwischen Nachschlagen und
// Eintragen überholen, bekamen deshalb **verschiedene** Antworten — gerade
// im Abriss- und Wiederholungsfall, für den es die Idempotenz gibt (Befund 4
// des Zweitblicks über Phase 5.6).
//
// # Und warum ein leerer Nachschlag ein Fehler bleibt
//
// **Weil der Konflikt dann keiner war.** Steht danach nichts da, hat niemand
// etwas eingetragen; ein erfundenes „schon behandelt" ließe die Nachricht nie
// wiederkommen, und der Tausch bliebe stehen (§10: eine Zusage, die niemand
// einlöst).
func (p *Handelspost) nachDemWettlauf(
	ctx context.Context, sender string, nachricht Handelsnachricht,
	abdruck string, schreibfehler error,
) (Zustellungsantwort, error) {
	if !errors.Is(schreibfehler, domain.ErrConflict) {
		return Zustellungsantwort{}, schreibfehler
	}

	vorher, fehler := p.briefkasten.Nachschlagen(ctx, nachricht.Umschlag.Idempotenzschluessel())
	if fehler != nil {
		return Zustellungsantwort{}, fehler
	}

	if vorher == nil {
		return Zustellungsantwort{}, schreibfehler
	}

	return p.schonVermerkt(ctx, sender, nachricht, vorher, abdruck)
}
