package service

// **Der Gast am anderen Ende der Leitung** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Was ein Gast tut und was er ausdrücklich nicht tut
//
// **Er wünscht, und er nimmt entgegen.** Sein Befehl geht als Kontrolle
// hinaus; was daraus wird, sagt der Wirt — und was der Wirt sagt, sind
// Tatsachen. Ein Gast, der aus einem Befehl selbst Ereignisse machte, hätte
// eine zweite Wahrheit neben der des Wirts, und beim ersten Unterschied wüsste
// niemand, welche gilt.
//
// **Deshalb weist er eine Kontrolle von außen ab.** Sie ist ein **Wunsch**,
// und Wünsche nimmt nur der Wirt entgegen. Das ist die Spiegelung von
// [Partiewirt], wo ein Ereignis von außen abgewiesen wird — beide Male aus
// demselben Grund, und beide Male ist es keine Formsache.
//
// # Warum er den Strom nur weiterreicht
//
// **Weil er keinen eigenen Bestand für eine fremde Partie hat.** Wohin ein
// Ereignis geht — in eine Anzeige, in einen Spiegel, in ein Protokoll —,
// entscheidet, wer ihn einsetzt. Diese Stelle sorgt dafür, dass es **in der
// richtigen Reihenfolge** und **nur einmal** dort ankommt; das ist ihr ganzer
// Beitrag, und sie behauptet keinen zweiten.

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Ereignisnehmer bekommt, was der Wirt für wahr erklärt hat.
//
// **Zwei Wege und nicht einer**: der Strom sagt, **was** geschehen ist, das
// Tischbild, **wie es jetzt aussieht**. Wer nur den Strom bekäme, müsste
// daraus einen Stand rechnen — und hätte damit den zweiten Spielkern, den es
// nicht geben soll ([domain.Kontrolle]).
type Ereignisnehmer interface {
	Ereignis(ctx context.Context, ereignis domain.Ereignis) error

	// Tischbild ist, was der Wirt für diesen Gast gerechnet hat — die Sicht
	// und was sie erlaubt.
	Tischbild(ctx context.Context, bild domain.Tischbild) error
}

// Partiegast schickt Wünsche und nimmt Tatsachen entgegen.
type Partiegast struct {
	partie  string
	sitzung *Partiesitzung
	nehmer  Ereignisnehmer

	nebenbei Nebenstrom

	// halter merkt sich, wo dieser Gast im Strom steht — nil heißt: niemand
	// merkt es sich, und dann fängt jede Rückkehr von vorn an.
	halter Standhalter

	// sitz ist der Sitz, den der Wirt diesem Gast gegeben hat.
	//
	// **Er kommt aus dem Tischbild** (`Sicht.Fuer`) und nicht aus dem
	// Beitritt: der Gast schreibt beim Beitreten keinen Sitz, den vergibt der
	// Wirt. Gebraucht wird er für den **sichtbezogenen** Abdruck.
	sitz string

	// pruefer hält den Stand der Präsenz — je Sprecher einen.
	//
	// **Er urteilt und beglaubigt nicht.** Wer geredet hat, hat der Wirt
	// gesetzt, und der ist die Autorität am Tisch; was dieser Gast nicht
	// glauben kann, ist die **Reihenfolge**: eine Präsenz geht verwerfbar
	// über die Leitung, und eine ältere kann nach einer neueren eintreffen.
	pruefer *Nebenpruefer

	// post stellt Bildantworten des Wirts zu (P9, #2377).
	post *Bildpost

	// mitgebracht ist das Deck, mit dem sich ein eingeladener Gast setzt
	// (P9b, #2377).
	mitgebracht *domain.Mitgebracht

	// maler und sicht beantworten die Bildfragen des Wirts (P9b, #2377):
	// gezeichnet wird nur, was die letzte Sicht nennt.
	maler Kartenbildmaler
	sicht domain.Partiesicht
}

// NeuerPartiegast baut ihn für **eine** Partie.
func NeuerPartiegast(
	partie string, sitzung *Partiesitzung, nehmer Ereignisnehmer,
) *Partiegast {
	return &Partiegast{
		partie: partie, sitzung: sitzung, nehmer: nehmer,
		pruefer: NeuerNebenpruefer(), post: &Bildpost{},
	}
}

// Sage schickt einen Satz an den Tisch.
//
// **Der Sprecher darin ist eine Lesehilfe.** Der Wirt setzt ihn aus der
// Sitzung, über die der Satz kam — dieselbe Regel wie bei
// [Partiegast.Wuensche], und aus demselben Grund: was auf der Leitung steht,
// ist eine Behauptung dessen, der sie beschreibt.
func (g *Partiegast) Sage(ctx context.Context, gesagt domain.Chat) error {
	if fehler := gesagt.Pruefe(); fehler != nil {
		return fehler
	}

	return g.schicke(ctx, domain.ArtChat, gesagt)
}

// Zeige schickt einen Zeiger, einen Griff, ein Schweben oder einen Ping.
//
// **Der Stand kommt vom Aufrufer und nicht von hier.** Er muss die Sitzung
// überleben: ein Zähler in diesem Verbund finge bei jedem Ruf von vorn an,
// denn der Gast wird je Ruf neu gebaut — was ihn hält, ist die Anwendung.
func (g *Partiegast) Zeige(
	ctx context.Context, gezeigt domain.Praesenz,
) error {
	if fehler := gezeigt.Pruefe(); fehler != nil {
		return fehler
	}

	return g.schicke(ctx, domain.ArtPraesenz, gezeigt)
}

// schicke verpackt eine geprüfte Nutzlast und gibt sie auf die Leitung.
//
// **Hier steht kein Fehlerzweig fürs Verpacken, und das ist eine Aussage.**
// [domain.Chat] und [domain.Praesenz] tragen nur Zeichenketten, Zahlen und
// Wahrheitswerte; `encoding/json` klagt darüber nie. Gehalten wird die
// Voraussetzung von `internal/domain/sichtgestalt_test.go` (§10).
func (g *Partiegast) schicke(
	ctx context.Context, art domain.Nachrichtenart, nutzlast any,
) error {
	inhalt, _ := json.Marshal(nutzlast) //nolint:errcheck,errchkjson // der Grund steht darüber

	return g.sitzung.Sende(ctx, art, inhalt)
}

// MitNebenstrom setzt, wohin Chat und Präsenz gehen.
func (g *Partiegast) MitNebenstrom(nebenbei Nebenstrom) *Partiegast {
	g.nebenbei = nebenbei

	return g
}

// MitStandhalter setzt, wer sich merkt, wo dieser Gast steht.
func (g *Partiegast) MitStandhalter(halter Standhalter) *Partiegast {
	g.halter = halter

	return g
}

// Wuensche schickt einen Spielbefehl an den Wirt.
//
// **Absender und Partie schreibt der Gast, und der Wirt glaubt sie nicht** —
// er setzt beide aus seiner Sitzung. Was hier hineingeschrieben wird, ist
// deshalb eine Bequemlichkeit für den Leser und keine Angabe, auf die sich
// jemand stützt.
func (g *Partiegast) Wuensche(
	ctx context.Context, befehl domain.Spielbefehl,
) error {
	inhalt, verpackt := domain.VerpackeKontrolle(domain.KontrolleWunsch, befehl)
	if verpackt != nil {
		return verpackt
	}

	return g.sitzung.Sende(ctx, domain.ArtKontrolle, inhalt)
}

// Angekommen nimmt eine Nachricht, die schon gilt.
func (g *Partiegast) Angekommen(
	ctx context.Context, nachricht domain.Partienachricht,
) error {
	if nachricht.Partie != g.partie {
		return fmt.Errorf(
			"%w: die Nachricht gehört zur Partie %q, dieser Gast sitzt an %q",
			domain.ErrValidation, nachricht.Partie, g.partie)
	}

	switch nachricht.Art {
	case domain.ArtEreignis:
		return g.nimm(ctx, nachricht)

	case domain.ArtKontrolle:
		return g.kontrolle(ctx, nachricht)

	case domain.ArtChat, domain.ArtPraesenz:
		return g.nebenherReichen(ctx, nachricht)
	}

	// **Eine Art, die es nicht gibt, hält an** — sonst verschwände die
	// nächste neue still.
	return nachricht.Art.Pruefe()
}

// kontrolle nimmt eine Kontrollnachricht — und der Gast kennt zwei Arten.
//
// **Ein Wunsch an einen Gast geht ins Leere.** Er hat keinen Vollzustand und
// legt nichts aus; wer ihm einen schickte, hielte ihn für den Wirt. Was er
// entgegennimmt, ist das **Tischbild** und die **Antwort auf seine Rückkehr**
// — beides rechnet der Wirt, und beides kommt nur von dort.
func (g *Partiegast) kontrolle(
	ctx context.Context, nachricht domain.Partienachricht,
) error {
	art, inhalt, fehler := domain.LiesKontrolle(
		nachricht.Inhalt, nachricht.Von,
		domain.KontrolleTischbild, domain.KontrolleAbzug, domain.KontrolleBild,
		domain.KontrolleBildwunsch)
	if fehler != nil {
		return fehler
	}

	switch art {
	case domain.KontrolleAbzug:
		return g.abzug(ctx, inhalt)
	case domain.KontrolleBild:
		return g.bildNehmen(inhalt)
	case domain.KontrolleBildwunsch:
		return g.bildAntworten(ctx, inhalt)
	}

	var bild domain.Tischbild

	if fehler := json.Unmarshal(inhalt, &bild); fehler != nil {
		return fmt.Errorf("%w: der Inhalt ist kein Tischbild: %s",
			domain.ErrValidation, fehler.Error())
	}

	return g.nimmBild(ctx, bild)
}

// nimmBild merkt sich den Sitz und reicht das Bild weiter.
//
// **Der Sitz kommt hier an und nirgends sonst.** Ohne ihn ließe sich der
// sichtbezogene Abdruck nicht rechnen — und ein Gast, der ihn beim Beitreten
// selbst wählte, wählte ihn falsch: vergeben wird er vom Wirt.
func (g *Partiegast) nimmBild(
	ctx context.Context, bild domain.Tischbild,
) error {
	if bild.Sicht.Fuer != "" {
		g.sitz = bild.Sicht.Fuer
	}

	g.sicht = bild.Sicht

	return g.nehmer.Tischbild(ctx, bild)
}

// nimm packt ein Ereignis aus und reicht es weiter.
func (g *Partiegast) nimm(
	ctx context.Context, nachricht domain.Partienachricht,
) error {
	var ereignis domain.Ereignis

	if fehler := json.Unmarshal(nachricht.Inhalt, &ereignis); fehler != nil {
		return fmt.Errorf("%w: der Inhalt ist kein Ereignis: %s",
			domain.ErrValidation, fehler.Error())
	}

	// **Die Partie kommt vom Gast und nicht aus dem Rumpf** — dieselbe
	// Regel wie beim Wirt, nur andersherum: was auf der Leitung steht, ist
	// eine Behauptung des anderen Endes.
	ereignis.PartieID = g.partie

	if fehler := g.nehmer.Ereignis(ctx, ereignis); fehler != nil {
		return fehler
	}

	g.merke(ereignis)

	return nil
}

// merke hält fest, wo dieser Gast jetzt steht.
//
// **Nach dem Weiterreichen und nicht davor** (§7, Regel 3): was der Nehmer
// nicht angenommen hat, hat dieser Gast auch nicht gesehen — und käme mit
// einem Stand zurück, den sein Fenster nie gezeigt hat.
//
// # Warum hier kein Fehlerzweig steht
//
// **Weil dieses Ereignis über eine Leitung kam und damit JSON war.**
// [domain.Ereignis.Sichtabdruck] kann nur an einem Wert scheitern, den
// `encoding/json` nicht abbilden kann — und einen solchen hat
// `json.Unmarshal` nie erzeugt. Ein Zweig, den kein Lauf betreten kann, ist
// eine Zusage, die niemand hält (§10).
//
// **Und wenn es doch je geschähe**, wäre der Abdruck leer und die nächste
// Rückkehr ein **Abbruch** mit Grund — laut und nicht still. Auf der
// Wirtsseite ist derselbe Weg **erreichbar**, denn dort kommen die Ereignisse
// aus einem Speicher hinter einer Schnittstelle; dort steht der Zweig.
func (g *Partiegast) merke(ereignis domain.Ereignis) {
	if g.halter == nil {
		return
	}

	abdruck, _ := ereignis.Sichtabdruck(g.sitz) //nolint:errcheck // der Grund steht darüber

	g.halter.Merke(ereignis.Folge, abdruck)
}

// nebenherReichen prüft eine Nebennachricht und reicht sie ans Fenster.
//
// # Warum der Gast prüft, obwohl der Wirt schon geprüft hat
//
// **Weil zwischen beiden eine Leitung liegt.** Die Beglaubigung des Wirts
// sagt, **wer** geredet hat; sie sagt nichts über die Reihenfolge, in der es
// **hier** ankommt. Eine Präsenz geht verwerfbar hinaus — sie darf verloren
// gehen und sie darf überholen. Ohne Stand spränge der Zeiger zurück (#1462:
// „monotonen Empfängerstand und Ablauf").
//
// **Und weil fremde Daten der Normalfall sind** (§8): was hereinkommt, hält
// seine Gestalt oder wird benannt — auch von einem Wirt, dem man vertraut.
//
// **Ohne Fenster geschieht nichts, und das ist kein Fehler.** Geprüft wird
// trotzdem: ein Stand, der nur mitliefe, solange jemand zusieht, wäre keiner.
func (g *Partiegast) nebenherReichen(
	ctx context.Context, nachricht domain.Partienachricht,
) error {
	// **Kein Sitz von außen** — der Wirt hat den Sprecher gesetzt, und er ist
	// die Autorität am Tisch (siehe den Kopf dieser Datei).
	befund, fehler := g.pruefer.Beglaubige(nachricht, "")
	if fehler != nil {
		return fehler
	}

	if !befund.Gilt || g.nebenbei == nil {
		return nil
	}

	return g.nebenbei.Nebenbei(ctx, befund)
}
