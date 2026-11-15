package service

// **Der Wirt am anderen Ende der Leitung** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Was hier geschieht
//
// **Ein Gast schickt einen Wunsch, der Wirt schickt Tatsachen zurück.** Das
// ist die ganze Aufteilung, und sie ist keine Bequemlichkeit: der Wirt hält
// den Vollzustand, er allein legt aus, und er allein vergibt Folge und Kette.
// Ein Gast, der Ereignisse **schickte**, schriebe damit am Strom mit — und
// niemand könnte hinterher sagen, welcher Stand gilt.
//
// # Deshalb wird ein Ereignis von außen abgewiesen
//
// **Es ist kein Formfehler, sondern eine Anmaßung.** Wer sie durchließe,
// hätte zwei Quellen für dieselbe Wahrheit; der Satz „der Wirt hält den
// Vollzustand" steht seit G1 in der Oberfläche, und hier wird er eingelöst.
//
// # Und warum weder Partie noch Sitz aus der Nachricht kommen
//
// **Weil beides der Gast schreibt.** Eine Sitzung belegt, **welcher Rechner**
// am anderen Ende ist — der Handschlag prüft die Peerkennung —, und daraus
// folgt, für **welchen Sitz** dieser Wirt sie führt: er wird beim Bauen
// festgelegt und nicht je Nachricht gelesen.
//
// **Der erste Anlauf nahm `nachricht.Von`**, und der Kopf darüber sagte schon
// damals „die Sitzung sagt, wer schreibt". Das war eine Zusage, die ihr Rumpf
// nicht hielt (§10): ein Gast konnte im Namen eines anderen Sitzes ziehen —
// samt dessen Hand.
//
// **Dasselbe für die Partie.** Die Nachricht trägt eine Kennung, der Befehl
// darin trägt auch eine; es gilt die dieses Wirts, und wer eine andere nennt,
// wird abgewiesen statt umgebogen.

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Befehlsnehmer ist, was der Wirt vom Dienst braucht.
//
// **Eine Schnittstelle und nicht `*Partieleitung`**: so lässt sich der Wirt
// prüfen, ohne einen Bestand zu öffnen (§2).
//
// **Und die Sicht gehört dazu.** Der Gast rechnet nichts: er bekommt nach
// jedem Zug das **Tischbild**, das der Wirt für ihn gerechnet hat. Ohne
// diesen zweiten Weg könnte er zwar wünschen, aber nie sehen, was daraus
// wurde — der Grund steht ausführlich bei [domain.Kontrolle].
type Befehlsnehmer interface {
	// BefehlVerkuendet stellt einen Befehl und trägt ihn hinaus, bevor ein
	// späterer hinausgeht (#2481). **Ein `Befehl` ohne Verkündung steht hier
	// nicht**: wer einen Zug an einen Tisch trägt, soll keinen Weg haben, die
	// Reihe zu überspringen.
	BefehlVerkuendet(ctx context.Context, befehl domain.Spielbefehl,
		verkuende Verkuendung) (Befehlsausgang, error)

	// InReihe führt eine Antwort aus, die liest und schickt, während keine
	// Verkündung läuft (#2481).
	InReihe(tun func() error) error

	Partie(ctx context.Context, kennung, fuer string) (
		domain.Partiesicht, error)

	// Tischbild rechnet Sicht, Erlaubnis und Objekthandlungen aus **einem**
	// Stand (#2481).
	//
	// **Erlaubnis und Handlungen neben der Sicht und nicht darin**: die eine
	// beschreibt den Zustand, die anderen das Regelwerk — der Grund steht
	// bei [domain.Tischbild].
	Tischbild(ctx context.Context, kennung, fuer string) (
		domain.Tischbild, error)

	// Wirtsstand sagt, wo dieser Wirt im Strom der Partie steht.
	//
	// **Aus der Sicht eines Sitzes**, denn der Abdruck ist sichtbezogen
	// ([domain.Ereignis.Sichtabdruck]).
	Wirtsstand(ctx context.Context, kennung, fuer string) (
		domain.Wirtsstand, error)

	// Nachreichen gibt die Ereignisse ab einer Folge — für **einen** Sitz.
	Nachreichen(ctx context.Context, kennung, fuer string, abFolge int64) (
		[]domain.Ereignis, error)
}

// Partiewirt nimmt fremde Befehle an und schickt seine Ereignisse zurück.
//
// **Einer je Gastsitzung und nicht einer je Partie.** Der Sitz gehört zur
// Sitzung: wer über **diese** Leitung schreibt, schreibt für **diesen** Sitz.
type Partiewirt struct {
	partie string

	// gast ist der Sitz, für den diese Sitzung spricht.
	//
	// **Er steht beim Bauen fest.** Wer ihn je Nachricht läse, läse eine
	// Behauptung des Gastes.
	gast string

	nehmer  Befehlsnehmer
	sitzung *Partiesitzung

	nebenbei Nebenstrom

	// runde ist der Tisch, an den ein Satz weitergereicht wird.
	//
	// **Nil heißt: kein Tisch, und das ist kein Fehler.** Ein Wirt ohne
	// offene Gastsitzungen führt die Partie trotzdem.
	runde *Partierunde

	// pruefer beglaubigt den Sprecher und hält den Stand der Präsenz.
	//
	// **Einer je Wirt und damit je Gastsitzung.** Der Stand gehört dem Weg,
	// über den gezählt wurde; ein Gast, der neu verbindet, bekommt einen
	// frischen Wirt und darf deshalb wieder bei eins anfangen.
	pruefer *Nebenpruefer

	// fenster ist das Fenster **dieses** Rechners, fenstersitz sein Platz.
	//
	// **Damit der Wirt fremde Züge sieht** (#1627): er spielt selbst mit,
	// und bis hierher erfuhr er von dem, was ein Gast stellt, nichts.
	fenster     Ereignisnehmer
	fenstersitz string

	// gegenueber ist der Peer am Gastsitz — für seinen Bestand wird das
	// Tischbild gerechnet (P9, #2377).
	gegenueber string

	// maler zeichnet Bilder für den Gast (P9, #2377) — ohne ihn sagt der
	// Wirt, dass er keine zeichnet.
	maler Kartenbildmaler

	// platz und leihen setzen einen eingeladenen Gast mit seinem Deck
	// (P9b, #2377) — ohne Platzgeber nimmt dieser Wirt keine Decks an.
	platz  Platzgeber
	leihen Leihgeber

	// post stellt zu, was der Gast auf eine Bildfrage antwortet (P9b,
	// #2377) — ohne sie fragt dieser Wirt nie, und ein Bild ist eine
	// Anmaßung.
	post *Bildpost
}

// NeuerPartiewirt baut ihn für **eine** Partie und **einen** Gastsitz.
func NeuerPartiewirt(
	partie, gast string, nehmer Befehlsnehmer, sitzung *Partiesitzung,
) *Partiewirt {
	return &Partiewirt{
		partie: partie, gast: gast, nehmer: nehmer, sitzung: sitzung,
		pruefer: NeuerNebenpruefer(),
	}
}

// MitGegenueber nennt den Peer, der am Gastsitz sitzt.
func (w *Partiewirt) MitGegenueber(peer string) *Partiewirt {
	w.gegenueber = peer

	return w
}

// MitRunde setzt den Tisch, an den Gesagtes weitergereicht wird.
func (w *Partiewirt) MitRunde(runde *Partierunde) *Partiewirt {
	w.runde = runde

	return w
}

// MitFenster sagt, welcher Sitz an diesem Rechner sitzt und wohin er hört.
func (w *Partiewirt) MitFenster(
	sitz string, nehmer Ereignisnehmer,
) *Partiewirt {
	w.fenster = nehmer
	w.fenstersitz = sitz

	return w
}

// Angekommen nimmt eine Nachricht, die schon gilt.
//
// **Sie gilt schon** — [Partienetz] hat über Epoche, Folge und Wiederholung
// geurteilt. Was hier entschieden wird, ist etwas anderes: **darf** der
// Absender das, und was folgt daraus.
func (w *Partiewirt) Angekommen(
	ctx context.Context, nachricht domain.Partienachricht,
) error {
	if nachricht.Partie != w.partie {
		return fmt.Errorf(
			"%w: die Nachricht gehört zur Partie %q, dieser Wirt führt %q",
			domain.ErrValidation, nachricht.Partie, w.partie)
	}

	switch nachricht.Art {
	case domain.ArtKontrolle:
		return w.kontrolle(ctx, nachricht)

	case domain.ArtEreignis:
		// **Der Wirt nimmt keine Ereignisse entgegen.** Sie sind Tatsachen,
		// und Tatsachen schafft er.
		return fmt.Errorf(
			"%w: %q schickt ein Ereignis — Ereignisse macht der Wirt",
			domain.ErrValidation, nachricht.Von)

	case domain.ArtChat, domain.ArtPraesenz:
		return w.weiterreichen(ctx, nachricht)
	}

	// **Eine Art, die es nicht gibt, hält an** — sonst verschwände die
	// nächste neue still.
	return nachricht.Art.Pruefe()
}

// kontrolle nimmt eine Kontrollnachricht — und der Wirt kennt genau eine Art.
//
// **Ein Tischbild an den Wirt ist eine Anmaßung** wie ein Ereignis: es ist das
// Ergebnis seiner eigenen Rechnung, und wer es ihm schickte, behauptete, für
// ihn gerechnet zu haben.
//
// **Zwei Arten nimmt er entgegen**: einen **Wunsch** — den Zug — und eine
// **Rückkehr** — die Frage, wo der Gast stehengeblieben ist. Beide kommen
// vom Gast, und beide beantwortet er; ein Abzug dagegen ist seine eigene
// Antwort und käme von dort nie.
func (w *Partiewirt) kontrolle(
	ctx context.Context, nachricht domain.Partienachricht,
) error {
	erwartet := []domain.Kontrollart{
		domain.KontrolleWunsch, domain.KontrolleRueckkehr,
		domain.KontrolleBildwunsch,
	}

	// **Ein Bild nimmt nur, wer fragt** (P9b, #2377) — ein Wirt ohne Post
	// fragt nie, und ihm ein Bild zu schicken ist eine Anmaßung.
	if w.post != nil {
		erwartet = append(erwartet, domain.KontrolleBild)
	}

	art, inhalt, fehler := domain.LiesKontrolle(
		nachricht.Inhalt, nachricht.Von, erwartet...)
	if fehler != nil {
		return fehler
	}

	switch art {
	case domain.KontrolleRueckkehr:
		return w.rueckkehr(ctx, inhalt)
	case domain.KontrolleBildwunsch:
		return w.bildAntworten(ctx, inhalt)
	case domain.KontrolleBild:
		return w.bildNehmen(inhalt)
	}

	return w.befehlen(ctx, inhalt)
}

// befehlen legt einen fremden Befehl vor und schickt zurück, was daraus wurde.
func (w *Partiewirt) befehlen(ctx context.Context, roh []byte) error {
	var befehl domain.Spielbefehl

	if fehler := json.Unmarshal(roh, &befehl); fehler != nil {
		return fmt.Errorf("%w: der Inhalt ist kein Spielbefehl: %s",
			domain.ErrValidation, fehler.Error())
	}

	// **Beides steht fest und wird nicht gelesen.** Der Sitz kommt aus der
	// Sitzung, die Partie von diesem Wirt; was in der Nachricht steht, ist
	// eine Behauptung des Gastes. Der Grund steht im Kopf dieser Datei.
	befehl.PartieID = w.partie
	befehl.Absender = w.gast

	// **Von der Leitung** (#2480): ein Gast nennt Objekte bei ihrer
	// Sichtkennung und nie bei der Kennung.
	//
	// **Verkündet in der Reihe** (#2481): kein späterer Zug überholt diesen.
	var verkuendet error

	_, fehler := w.nehmer.BefehlVerkuendet(MitLeitungsherkunft(ctx), befehl,
		func(ausgang Befehlsausgang) { verkuendet = w.zustellen(ctx, ausgang) })
	if fehler != nil {
		return w.abgewiesen(ctx, befehl, fehler)
	}

	return verkuendet
}

// zustellen gibt dem Gast seinen Zug und sein Bild — **und dann dem Rest des
// Tisches** (#1627). Bis dahin endete ein Zug bei dem, der ihn gestellt hat.
func (w *Partiewirt) zustellen(ctx context.Context, ausgang Befehlsausgang) error {
	if fehler := w.verkuenden(ctx, ausgang); fehler != nil {
		return fehler
	}

	if fehler := w.tischbildSchicken(ctx); fehler != nil {
		return fehler
	}

	return w.funk().Verkuende(ctx, w.gast, ausgang)
}

// tischbildSchicken gibt dem Gast die Sicht, die für ihn gerechnet wurde.
//
// # Warum nach den Ereignissen und nicht davor
//
// **Weil das Bild den Stand **nach** dem Zug zeigt.** Wer es davor schickte,
// gäbe dem Gast einen Tisch, der gleich veraltet ist — und die Ereignisse
// danach wären eine Erklärung für etwas, das er nie gesehen hat.
//
// # Und warum überhaupt ein Bild und nicht nur die Ereignisse
//
// **Weil der Gast nichts rechnet.** Er hat keinen Vollzustand und keinen
// Regelausleger; ein zweiter Spielkern auf seiner Seite hätte beim ersten
// Unterschied eine zweite Wahrheit. Der Grund steht ausführlich bei
// [domain.Kontrolle].
func (w *Partiewirt) tischbildSchicken(ctx context.Context) error {
	bild, fehler := w.tischbild(ctx)
	if fehler != nil {
		return fehler
	}

	// **Hier steht kein Fehlerzweig, und das ist eine Aussage.** Beide
	// Gründe, aus denen [domain.VerpackeKontrolle] klagen kann, sind hier
	// ausgeschlossen: die Art ist eine Konstante, und eine
	// [domain.Partiesicht] trägt ausschließlich Zeichenketten, Zahlen und
	// Wahrheitswerte — `encoding/json` klagt darüber nie.
	//
	// **Ein Zweig, den kein Lauf betreten kann, ist eine Zusage, die niemand
	// hält** (§10). Gehalten wird die Voraussetzung statt dessen von
	// `internal/domain/sichtgestalt_test.go`: es gibt kein offenes Feld im
	// Tischbild, und wer eines hinzufügt, macht ihn rot.
	inhalt, _ := domain.VerpackeKontrolle( //nolint:errcheck // der Grund steht darüber
		domain.KontrolleTischbild, bild)

	return w.sitzung.Sende(ctx, domain.ArtKontrolle, inhalt)
}

// tischbild rechnet, was dieser Sitz sehen darf.
//
// **Herausgelöst, seit die Rückkehr es auch braucht**: beim Abzug steckt es in
// der Antwort und geht nicht als eigene Nachricht hinaus. Zwei Stellen, die es
// bauen, liefen beim ersten neuen Feld auseinander (§16).
func (w *Partiewirt) tischbild(ctx context.Context) (domain.Tischbild, error) {
	return tischbildFuer(
		MitBestandssicht(ctx, w.gegenueber), w.nehmer, w.partie, w.gast)
}

// verkuenden schickt jedes Ereignis als solches hinaus.
//
// **Eines je Nachricht und nicht alle in einer.** Der Empfänger urteilt je
// Folge; ein Bündel hätte eine Nummer für mehrere Schritte, und eine Lücke
// mittendrin wäre nicht zu benennen.
func (w *Partiewirt) verkuenden(
	ctx context.Context, ausgang Befehlsausgang,
) error {
	for _, ereignis := range ausgang.Ereignisse {
		// **Projiziert und nicht roh.** Ein [domain.Ereignis] trägt die
		// privaten Ergänzungen **aller** Sitze; hier ging bis zum 2026-09-19
		// jede von ihnen an jeden Gast hinaus (#1627).
		inhalt, fehler := json.Marshal(ereignis.Fuer(w.gast))
		if fehler != nil {
			return fmt.Errorf("das Ereignis ließ sich nicht verpacken: %w",
				fehler)
		}

		if fehler := w.sitzung.Sende(
			ctx, domain.ArtEreignis, inhalt); fehler != nil {
			return fehler
		}

		// **Jedes gleich hinaus** — es können mehr sein, als die Schlange
		// fasst ([Partiesitzung.Leere]).
		if fehler := w.sitzung.Leere(ctx); fehler != nil {
			return fehler
		}
	}

	return nil
}
