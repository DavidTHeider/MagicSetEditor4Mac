package service

// **Die offenen Plätze eines Tisches** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum es das überhaupt braucht
//
// **Weil der Wirt bisher nur antworten konnte.** Jede Gastsitzung lief in
// ihrer eigenen Schleife, und niemand hielt sie fest; der Wirt hatte damit
// keinen Weg, von sich aus etwas zu schicken — kein Wort, keinen Zeiger, und
// erst recht nichts an **alle**. Ein Tisch, an dem nur auf Zuruf geredet wird,
// ist kein Tisch.
//
// **Und weil ein Satz von Anna bei Bert ankommen muss.** Ohne diese Stelle
// hätte eine Partie mit drei Sitzen drei Zwiegespräche mit dem Wirt und kein
// gemeinsames.
//
// # Warum der Wirt beim Weiterreichen der Absender bleibt
//
// **Weil er es wirklich geschickt hat.** `Partienachricht.Von` ist der
// Absender **dieses Sprungs**; wer geredet hat, steht in der Nutzlast und ist
// vom Wirt gesetzt ([domain.Chat]). Beides gleichzusetzen hieße entweder
// lügen oder die Herkunft verlieren.
//
// # Und warum unter dem Schloss nur abgeschrieben und nicht gesendet wird
//
// **Weil ein langsamer Gast sonst den Tisch anhält** (§7). Die Plätze werden
// unter dem Schloss abgeschrieben, gesendet wird danach — und ein Platz, der
// klagt, hält die übrigen nicht auf: gesammelt wird, wer gescheitert ist, und
// der Aufrufer bekommt alles auf einmal.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Partieplatz ist der Weg zu einem Gast — senden und hinausgeben.
//
// **Zwei Methoden und nicht eine**: [Partiesitzung.Sende] legt eine Nachricht
// in die Warteschlange, hinaus geht sie erst mit `Leere`. Wer das Zweite
// vergäße, hätte einen Tisch, an dem alles gesagt und nichts gehört wird —
// genau dieser Fehler ist in diesem Werkstück schon einmal passiert.
type Partieplatz interface {
	Partienaht

	Leere(ctx context.Context) error

	// Schliesse legt die Leitung nieder — der Gast ist entlassen
	// (Zweitblick über Phase 6.12, Z3).
	Schliesse() error

	// Stand sagt, wie es um diese Leitung steht (G5,
	// #1462).
	//
	// **Hier und nicht in einer Nebenliste der Anwendung.** Wer sitzt, steht
	// in dieser Runde; eine zweite Aufzählung „Sitz → Leitung" müsste an
	// denselben zwei Stellen gepflegt werden, und die erste, die es vergisst,
	// merkt niemand (§10).
	Stand() Sitzungsstand
}

// Gastplatz ist ein besetzter Sitz samt seiner Nummerierung.
type Gastplatz struct {
	sitz    string
	sitzung *Partiesitzung
	platz   Partieplatz

	// peer ist der Gast, der hier sitzt — für seinen Bestand wird ihm die
	// Sicht gerechnet (P9, #2377).
	peer string

	// post stellt die Bildantworten dieses Gastes zu (P9b, #2377) — sie
	// teilt sie mit dem Wirt seines Sitzes, der sie empfängt.
	post *Bildpost

	// bildfaehig sagt, dass dieser Gast Bildwünsche beantwortet — erklärt in
	// seinem Rückkehrwunsch ([Partierunde.Kann]).
	bildfaehig bool
}

// MitPeer nennt den Gast, der an diesem Platz sitzt.
func (g *Gastplatz) MitPeer(peer string) *Gastplatz {
	g.peer = peer

	return g
}

// NeuerGastplatz bindet Sitz, Nummerierung und Leitung zusammen.
func NeuerGastplatz(
	sitz string, sitzung *Partiesitzung, platz Partieplatz,
) *Gastplatz {
	return &Gastplatz{sitz: sitz, sitzung: sitzung, platz: platz}
}

// schicke legt eine Nachricht auf die Leitung und gibt sie hinaus.
func (g *Gastplatz) schicke(
	ctx context.Context, art domain.Nachrichtenart, inhalt []byte,
) error {
	if fehler := g.sitzung.Sende(ctx, art, inhalt); fehler != nil {
		return fehler
	}

	return g.platz.Leere(ctx)
}

// Partierunde führt die offenen Gastsitzungen einer Partie.
//
// **Ihr Nullwert ist benutzbar** — eine Anwendung ohne offenen Tisch hat eine
// leere Runde, und das ist kein Sonderfall, sondern der Normalfall.
type Partierunde struct {
	schloss sync.Mutex
	plaetze map[string]*Gastplatz

	// staende vergibt die Zeigerstände, **je Sprecher einen**.
	//
	// # Warum sie hier liegen und nicht bei dem, der zeigt
	//
	// **Weil sie den Abriss einer Gastsitzung überleben müssen.** Zählte der
	// Gast, finge er nach einem Neuverbinden wieder bei eins an — und jeder
	// andere Gast wiese von da an alles ab, was von ihm kommt, weil sein
	// [domain.Praesenzkreis] eine höhere Zahl kennt.
	//
	// **Der Wirt nummeriert beim Weiterreichen also um, und er darf das**: er
	// ist die Autorität am Tisch, und der Stand ist eine Aussage über
	// **seinen** Ausgang, nicht über den Eingang eines Gastes.
	//
	// **Eine Präsenz verbraucht dabei keine Folge** — der Grund steht bei
	// [domain.Praesenz].
	staende map[string]uint64
}

// Setze trägt einen Platz ein — oder löst den vorigen desselben Sitzes ab.
func (r *Partierunde) Setze(platz *Gastplatz) {
	r.schloss.Lock()
	defer r.schloss.Unlock()

	if r.plaetze == nil {
		r.plaetze = map[string]*Gastplatz{}
	}

	r.plaetze[platz.sitz] = platz
}

// Geht räumt einen Sitz weg — aber nur, wenn er noch derselbe ist.
//
// **Die Prüfung ist nicht Förmlichkeit.** Eine abgerissene Sitzung endet
// **nachdem** die neue schon steht: der Gast verbindet neu, `Setze` trägt den
// frischen Platz ein, und erst danach läuft die alte Schleife aus. Wer dann
// blind löschte, nähme den lebenden Platz weg — und der Gast säße am Tisch,
// ohne etwas zu hören.
func (r *Partierunde) Geht(platz *Gastplatz) {
	r.schloss.Lock()
	defer r.schloss.Unlock()

	if r.plaetze[platz.sitz] == platz {
		delete(r.plaetze, platz.sitz)
	}
}

// Entlasse legt jede Leitung dieser Runde nieder.
//
// **Mehr als das Ende der Freigabe** (Zweitblick über Phase 6.12, Z3): die
// Freigabe zu beenden schickt niemanden vom Tisch, und so bleibt es. Wer aber
// `spielen` oder `netz` abschaltet, will auch keinen Gast mehr, der Befehle
// an einen Tisch stellt, den der Wirt nicht mehr sieht. Weggeräumt werden die
// Plätze von ihren Schleifen, sobald die Leitung endet ([Partierunde.Geht]).
// **Ein Fehler beim Niederlegen hält die übrigen nicht auf** — eine Leitung,
// die sich nicht schließen lässt, ist ohnehin am Ende, und jeder weitere
// Gast soll trotzdem gehen.
func (r *Partierunde) Entlasse() {
	r.schloss.Lock()
	plaetze := make([]*Gastplatz, 0, len(r.plaetze))

	for _, platz := range r.plaetze {
		plaetze = append(plaetze, platz)
	}
	r.schloss.Unlock()

	for _, platz := range plaetze {
		_ = platz.platz.Schliesse() //nolint:errcheck // siehe oben
	}
}

// Peer gibt den Gast an einem Sitz — leer, wenn dort keiner sitzt.
func (r *Partierunde) Peer(sitz string) string {
	r.schloss.Lock()
	defer r.schloss.Unlock()

	if platz := r.plaetze[sitz]; platz != nil {
		return platz.peer
	}

	return ""
}

// Sitze gibt die besetzten Sitze, geordnet.
//
// **Geordnet und nicht in Kartenreihenfolge**: eine Liste, die bei jedem Ruf
// anders aussieht, macht aus einer Anzeige ein Flackern.
func (r *Partierunde) Sitze() []string {
	r.schloss.Lock()
	defer r.schloss.Unlock()

	sitze := make([]string, 0, len(r.plaetze))
	for sitz := range r.plaetze {
		sitze = append(sitze, sitz)
	}

	sort.Strings(sitze)

	return sitze
}

// Leitungen sagt, wie es um jede Leitung dieses Tisches steht.
//
// **Nach Sitz geordnet**, denn die Antwort steht in einem Fenster: eine Liste,
// deren Reihenfolge bei jedem Ruf eine andere ist, lässt sich weder lesen noch
// vergleichen — derselbe Grund wie bei [Partierunde.Sitze].
func (r *Partierunde) Leitungen() []Sitzungsstand {
	r.schloss.Lock()
	defer r.schloss.Unlock()

	leitungen := make([]Sitzungsstand, 0, len(r.plaetze))

	for sitz, platz := range r.plaetze {
		stand := platz.platz.Stand()

		// **Die Kennung ist der Sitz und nicht die der Peersitzung.** Wer
		// am Tisch nachsieht, sucht einen Mitspieler; die Sitzungskennung
		// beantwortet eine Frage, die dort niemand stellt.
		stand.Kennung = sitz

		leitungen = append(leitungen, stand)
	}

	sort.Slice(leitungen, func(eins, zwei int) bool {
		return leitungen[eins].Kennung < leitungen[zwei].Kennung
	})

	return leitungen
}

// Sage schickt einen Satz an alle Gäste außer dem, der ihn gesagt hat.
//
// **Außer ihm**, denn er hat ihn schon: sein eigenes Fenster hat ihn beim
// Absenden angezeigt. Ein Echo käme als zweiter Satz an und sähe aus wie ein
// doppeltes Wort.
func (r *Partierunde) Sage(ctx context.Context, gesagt domain.Chat) error {
	if fehler := gesagt.Pruefe(); fehler != nil {
		return fehler
	}

	return r.anAlle(ctx, gesagt.Von, domain.ArtChat, gesagt)
}

// Zeige schickt eine Präsenz an alle Gäste — und vergibt ihren Stand.
//
// **Der Stand kommt von hier und nicht vom Aufrufer** — der Grund steht bei
// [Partierunde.staende].
func (r *Partierunde) Zeige(
	ctx context.Context, gezeigt domain.Praesenz,
) error {
	gezeigt.Stand = r.naechsterStand(gezeigt.Von)

	if fehler := gezeigt.Pruefe(); fehler != nil {
		return fehler
	}

	return r.anAlle(ctx, gezeigt.Von, domain.ArtPraesenz, gezeigt)
}

// naechsterStand zählt für **einen** Sprecher hoch.
func (r *Partierunde) naechsterStand(sprecher string) uint64 {
	r.schloss.Lock()
	defer r.schloss.Unlock()

	if r.staende == nil {
		r.staende = map[string]uint64{}
	}

	r.staende[sprecher]++

	return r.staende[sprecher]
}

// Reiche gibt weiter, was ein Gast gesagt hat — an alle anderen.
//
// **Die Nutzlast ist schon beglaubigt** ([Nebenpruefer]); geprüft wird sie
// hier nicht noch einmal. Wer das täte, prüfte die eigene Arbeit.
//
// **Eine Präsenz wird dabei neu nummeriert.** Der Stand des Gastes gilt für
// **seinen** Sprung zum Wirt; weitergereicht bekommt sie den Stand dieser
// Runde — der Grund steht bei [Partierunde.staende].
func (r *Partierunde) Reiche(
	ctx context.Context, befund Nebenbefund,
) error {
	if befund.Nachricht.Art == domain.ArtPraesenz {
		return r.Zeige(ctx, befund.Praesenz)
	}

	return r.Sage(ctx, befund.Chat)
}

// Schicke legt eine Nachricht auf die Leitung **eines** Sitzes.
//
// **Ein Sitz, den es nicht (mehr) gibt, ist kein Fehler.** Zwischen dem
// Aufzählen der Sitze und dem Schicken kann eine Leitung zugegangen sein
// ([Partierunde.Geht]); wer daraus einen Fehler machte, ließe den Tisch an
// dem scheitern, der gerade gegangen ist.
func (r *Partierunde) Schicke(
	ctx context.Context, sitz string,
	art domain.Nachrichtenart, inhalt []byte,
) error {
	r.schloss.Lock()
	platz, sitzt := r.plaetze[sitz]
	r.schloss.Unlock()

	if !sitzt {
		return nil
	}

	return platz.schicke(ctx, art, inhalt)
}

// anAlle verpackt eine Nutzlast und schickt sie an alle außer einem.
func (r *Partierunde) anAlle(
	ctx context.Context, ausser string,
	art domain.Nachrichtenart, nutzlast any,
) error {
	// **Kein Fehlerzweig, und der Grund ist derselbe wie bei
	// [mitInhalt]**: beide Nutzlasten tragen nur Zeichenketten und Zahlen.
	inhalt, _ := json.Marshal(nutzlast) //nolint:errcheck,errchkjson // siehe mitInhalt

	return r.anJeden(ctx, ausser, art, inhalt)
}

// anJeden schickt fertige Bytes an jeden Platz außer einem.
func (r *Partierunde) anJeden(
	ctx context.Context, ausser string,
	art domain.Nachrichtenart, inhalt []byte,
) error {
	var klagen []error

	for _, platz := range r.abschrift() {
		if platz.sitz == ausser {
			continue
		}

		if fehler := platz.schicke(ctx, art, inhalt); fehler != nil {
			klagen = append(klagen, fmt.Errorf(
				"an %q: %w", platz.sitz, fehler))
		}
	}

	return errors.Join(klagen...)
}

// abschrift gibt die Plätze, ohne das Schloss über das Senden zu halten.
func (r *Partierunde) abschrift() []*Gastplatz {
	r.schloss.Lock()
	defer r.schloss.Unlock()

	plaetze := make([]*Gastplatz, 0, len(r.plaetze))
	for _, platz := range r.plaetze {
		plaetze = append(plaetze, platz)
	}

	return plaetze
}
