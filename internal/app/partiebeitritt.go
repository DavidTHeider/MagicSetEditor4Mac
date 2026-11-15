package app

// **An einem fremden Tisch mitspielen** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Die andere Hälfte
//
// `partiefreigabe.go` öffnet einen Tisch, `partieaufnahme.go` bedient, wer
// ankommt. **Hier wird angekommen.** Ohne diese Seite wäre die Freigabe eine
// Zusage, die dieses Programm selbst nicht einlösen kann (§10).
//
// # Warum der Gast keine Partie im Bestand hat
//
// **Weil sie dem Wirt gehört.** Er hält den Vollzustand, er legt aus, er
// vergibt Folge und Kette; der Gast bekommt das **Tischbild**, das für ihn
// gerechnet wurde, und zeichnet es. Ein zweiter Spielkern auf seiner Seite
// hätte beim ersten Unterschied eine zweite Wahrheit
// ([domain.Kontrolle]).
//
// **Was das kostet, gehört ausgesprochen**: reißt die Leitung ab, sieht der
// Gast nichts mehr — nicht einmal, was zuletzt galt. Ein letzter Stand im
// Fenster wäre ein Tisch, an dem inzwischen gespielt worden sein kann, und
// das ist schlimmer als keiner.
//
// # Warum das Tischbild gemeldet und nicht abgefragt wird
//
// **Weil der Gast nicht weiß, wann etwas geschieht.** Was vom Wirt kommt,
// erreicht ihn über dieselbe Naht, über die die Menüleiste meldet
// (`fensternaht.go`). Ein Fenster, das pollte, fragte die meiste Zeit umsonst
// und wäre trotzdem zu spät.
//
// **Ein anderer Sitz zieht, und er erfährt es** — seit dem 2026-09-19 über
// [service.Tischfunk]
// (#1627);
// was ankommt, steht in `partietischmelder.go`.

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Die Ereignisse, über die ein Gast erfährt, was am Tisch geschieht.
const (
	// EreignisPartietischbild trägt die Sicht, die der Wirt gerechnet hat.
	EreignisPartietischbild = "partie:tischbild"

	// EreignisPartiezug trägt ein einzelnes Ereignis des Stroms.
	//
	// **Neben dem Bild und nicht statt seiner**: das Bild sagt, wie es
	// **aussieht**, der Strom, was **geschehen** ist. Eine Oberfläche, die
	// nur das Bild bekäme, könnte keinen Verlauf zeigen.
	EreignisPartiezug = "partie:zug"

	// EreignisPartienebenbei trägt Chat und Präsenz.
	EreignisPartienebenbei = "partie:nebenbei"

	// EreignisEigenesTischbild und EreignisEigenerZug sind dieselben beiden
	// für den **eigenen** Tisch: was ein Gast dort tut, erfährt das Fenster
	// des Wirts (#1627). **Unter eigenem Namen** — Zweitblick über Phase 6.11
	// (#1672, zu Z1): dieselbe Anwendung kann zugleich Gast an einem anderen
	// Tisch sein, und dessen Haken nähme sonst das Bild des eigenen.
	EreignisEigenesTischbild = "partie:eigen:tischbild"
	EreignisEigenerZug       = "partie:eigen:zug"

	// EreignisPartiefreigabe sagt dem Fenster des Wirts, dass ein Gast an
	// seinem Tisch Platz genommen hat oder gegangen ist — gefunden mit zwei
	// Fenstern (`nat04`): die Freigabe zeigte „Noch sitzt niemand", während
	// Bo schon zog. Die Nutzlast ist die Partie.
	EreignisPartiefreigabe = "partie:freigabe"

	// EreignisPartiebeitritt sagt dem Fenster des Gastes, dass sein Tisch
	// endete — P9c (#2377): es fragt dann den Stand und zeigt die Absage
	// des Wirts, statt bis „Nachsehen“ an einem Tisch zu sitzen, den es
	// nicht mehr gibt.
	EreignisPartiebeitritt = "partie:beitritt"
)

// gastplatzhalter steht dort, wo der Gast seinen Sitz nicht kennt.
//
// **Er ist eine Lesehilfe und keine Angabe.** Was ein Gast als Absender oder
// Sprecher schreibt, verwirft der Wirt und setzt es aus der Sitzung; wer ihn
// für tragend hielte, suchte den Sitz an der falschen Stelle.
//
// **Der Gast könnte ihn kennen** — das Tischbild trägt ihn als
// `Partiesicht.Fuer` —, aber es wäre ein zweiter Zustand ohne zweiten Nutzen:
// gelten würde er trotzdem nicht.
const gastplatzhalter = "gast"

// errVomTischAufgestanden ist der Grund, wenn jemand selbst aufhört.
//
// **Ein Grund und kein Fehler**: er sagt, warum eine Sitzung niedergelegt
// wird, und niemand soll ihn als Fehlschlag lesen.
var errVomTischAufgestanden = errors.New("jemand ist vom Tisch aufgestanden")

// PartieBeitreten setzt sich an den Tisch eines Gegenübers.
//
// **`externSenden` und nicht `externEmpfangen`**: dieser Rechner wird dabei
// nicht erreichbar, er **greift hinaus**. Die Klasse steht in
// `docs/AUTORISIERUNG.md` neben den anderen Wegen, die wählen.
func (a *App) PartieBeitreten(
	ctx context.Context, gegenueber, partie string, mit PartiemitbringDTO,
) PartiebeitrittAntwort {
	ctx = a.imProfil(ctx)

	// **Die Marke vor dem Start** (Nachlauf über Phase 6.12, N2).
	marke := a.laufmarke(domain.LaufPartiebeitritt)

	if fehler := a.darf(ctx, domain.WirkungExternSenden); fehler != nil {
		return PartiebeitrittAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	// **Die Partie darf fehlen** (UX-08.AK-04, #1680): dann nennt sie der
	// Wirt mit seinem ersten Wort ([App.partieErfragen]). Das Gegenüber
	// nicht — ohne es gibt es niemanden, den man fragen könnte.
	if gegenueber == "" {
		return PartiebeitrittAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: es braucht ein Gegenüber", domain.ErrValidation))}
	}

	// **Was mitkommt, steht vor dem Wählen fest** (P9b, #2377) — ein Deck,
	// das sich nicht lesen lässt, ist ein Fehler hier und keine Absage dort.
	mitgebracht, fehler := a.mitbringen(ctx, mit)
	if fehler != nil {
		return PartiebeitrittAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	griff, fehler := a.sitzungsverwaltung().Oeffne(
		ctx, peer.Kennung(gegenueber), domain.ZweckPartie)
	if fehler != nil {
		return PartiebeitrittAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	sitzung, lauf := a.beitrittMerken(griff, partie)
	a.mitgebrachtMerken(mitgebracht)

	// **Der Kontext des Rufs endet mit der Antwort** — der Tisch nicht.
	// Derselbe Grund wie beim Horchposten: `context.WithoutCancel` behält
	// das Profil und lässt das Ende fallen.
	tisch := context.WithoutCancel(ctx)
	if partie == "" {
		a.amFremdenTisch(tisch, griff, lauf, func() { a.partieErfragen(tisch, griff, lauf) })
	} else {
		a.amFremdenTisch(tisch, griff, lauf, func() { a.amTischSitzen(tisch, griff, partie, sitzung, lauf) })
	}

	// **Und dann der Schalter noch einmal** (Zweitblick über Phase 6.12,
	// Z2) — die Antwort ist der Stand danach.
	a.nachDemStart(ctx, domain.LaufPartiebeitritt, marke, a.partiebeitrittBeenden)

	return a.PartieBeitrittStand(ctx)
}

// PartieBeitrittBeenden steht vom Tisch auf.
//
// # Warum Aufstehen immer gelingt
//
// **Weil es eine Entscheidung ist und keine Bitte.** Wer aufsteht, sitzt
// danach nicht mehr — auch wenn die Leitung beim Niederlegen klagt, etwa weil
// sie ohnehin schon weg war. Ein Fehler hier zwänge das Fenster, den Menschen
// weiter an einem Tisch zu zeigen, an dem er nicht mehr ist; die Klage selbst
// erreicht niemanden, der sie brauchen könnte, und steht deshalb im Protokoll
// (§8).
func (a *App) PartieBeitrittBeenden(ctx context.Context) PartiebeitrittAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungSchreiben); fehler != nil {
		return PartiebeitrittAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	a.partiebeitrittBeenden(ctx)

	return PartiebeitrittAntwort{}
}

// partiebeitrittBeenden steht vom fremden Tisch auf — der Weg hinter
// [App.PartieBeitrittBeenden] und hinter dem Schalter (OPS-10b, #1978).
func (a *App) partiebeitrittBeenden(ctx context.Context) {
	griff, _, _ := a.beitritt()
	a.beitrittMerken(nil, "")

	// **Wer aufsteht, vergisst seinen Stand.** Er kommt nicht zurück — und
	// ein Stand, der einen fremden Tisch überdauerte, machte aus dem nächsten
	// Beitritt eine Rückkehr in eine Partie, die er nie gesehen hat.
	a.tischVergessen()

	if griff != nil {
		a.sitzungAbweisen(ctx, griff, errVomTischAufgestanden)
	}
}

// PartieBeitrittStand sagt, an welchem Tisch dieses Programm sitzt.
func (a *App) PartieBeitrittStand(ctx context.Context) PartiebeitrittAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungLesen); fehler != nil {
		return PartiebeitrittAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	griff, partie, _ := a.beitritt()
	if griff == nil {
		return PartiebeitrittAntwort{Pausiert: a.istPausiert(), Absage: a.absage()}
	}

	return PartiebeitrittAntwort{
		Laeuft: true, Partie: partie,
		Gegenueber: fingerabdruckVon(griff.Stand().Gegenueber),
	}
}

// PartieWuenschen schickt einen Spielbefehl an den Wirt.
//
// **Ein Wunsch und keine Wirkung**: was daraus wird, entscheidet der Wirt, und
// was er entscheidet, kommt als Ereignis und Tischbild zurück — nicht als
// Antwort auf diesen Ruf. Wer hier auf ein Ergebnis wartete, wartete auf
// etwas, das über eine andere Tür kommt.
func (a *App) PartieWuenschen(
	ctx context.Context, befehl domain.Spielbefehl,
) PartiebeitrittAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungExternSenden); fehler != nil {
		return PartiebeitrittAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	gast, griff, fehler := a.gastAmTisch(ctx)
	if fehler != nil {
		return PartiebeitrittAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	if fehler := gast.Wuensche(ctx, befehl); fehler != nil {
		return PartiebeitrittAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	if fehler := griff.Leere(ctx); fehler != nil {
		return PartiebeitrittAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return a.PartieBeitrittStand(ctx)
}

// errAnKeinemFremdenTisch ist der Grund, wenn dieses Programm nirgends sitzt.
var errAnKeinemFremdenTisch = fmt.Errorf(
	"%w: dieses Programm sitzt an keinem fremden Tisch", domain.ErrConflict)

// errNochOhnePartie ist der Grund, wenn der Wirt seine Partie noch nicht
// genannt hat — gesagt werden kann dann noch nichts (UX-08.AK-04, #1680).
var errNochOhnePartie = fmt.Errorf(
	"%w: der Wirt hat seine Partie noch nicht genannt", domain.ErrConflict)

// gastAmTisch baut die **Sendeseite** für den fremden Tisch.
//
// # Warum sie an einer Stelle gebaut wird
//
// **Weil es sonst zwei Wurzeln für dasselbe gäbe.** Ein Wunsch und ein Satz
// gehen denselben Weg hinaus; zwei Aufbauten liefen beim ersten Unterschied
// auseinander, und `internal/werkzeuge/nahtverdrahtung_test.go` hat genau das
// angestrichen, als der zweite entstand.
//
// # Und warum hier kein Nebenstrom hängt
//
// **Weil das die Empfangsseite ist.** Was hereinkommt, nimmt die Schleife in
// [App.amTischSitzen] entgegen — und nur sie hängt das Fenster an. Ein
// Nebenstrom an einem Gast, der ausschließlich schickt, wäre eine Zusage ohne
// Einlösung (§10).
func (a *App) gastAmTisch(ctx context.Context) (
	*service.Partiegast, service.Angenommene, error,
) {
	griff, partie, sitzung := a.beitritt()
	if griff == nil {
		return nil, nil, errAnKeinemFremdenTisch
	}

	if sitzung == nil {
		return nil, nil, errNochOhnePartie
	}

	return service.NeuerPartiegast(
		partie, sitzung, a.tischmelder(ctx, partie),
	).MitBildpost(a.bildpost()), griff, nil
}

// amTischSitzen führt die Empfangsschleife, bis die Leitung endet.
func (a *App) amTischSitzen(
	ctx context.Context, griff service.Angenommene, partie string,
	sitzung *service.Partiesitzung, lauf uint64,
) {
	gast := service.NeuerPartiegast(
		partie, sitzung, a.tischmelder(ctx, partie)).
		MitNebenstrom(a.tischmelder(ctx, partie)).
		MitStandhalter(a).
		MitBildpost(a.bildpost()).
		MitMitgebracht(a.mitgebracht()).
		MitMaler(eigenmaler{a})

	netz := service.NeuesPartienetz(
		griff, service.NeuerPartieempfang(erstePartieepoche), gast,
	).MitMeldung(func(fehler error) {
		slog.WarnContext(ctx, "eine Nachricht des Wirts wurde nicht angenommen",
			"partie", partie, "grund", fehler)
	})

	// **Der Wunsch geht nebenher hinaus, und das ist keine Bequemlichkeit.**
	//
	// Die Leitung ist ein Strom **ohne Puffer**: wer schreibt, wartet, bis
	// jemand liest. Ein Gast, der erst seinen Wunsch hinausgäbe und **danach**
	// zu lesen anfinge, hielte den Wirt beim Antworten fest — und beide
	// warteten aufeinander. Der Lauf über beide Enden ist genau so
	// stehengeblieben, mit `leitung.Sende` auf beiden Seiten des Stapels.
	//
	// **Auch beim ersten Mal** wird gefragt, mit Folge null: ein Sonderweg
	// „erstes Mal" wäre ein zweiter Weg zum selben Ziel (§16).
	a.amFremdenTisch(ctx, griff, lauf, func() { a.rueckkehrMelden(ctx, gast, griff, partie) })

	ende := netz.Bediene(ctx)
	slog.InfoContext(ctx, "der Tisch eines Wirts ist zu Ende",
		"partie", partie, "grund", ende)

	// **Die Sitzung geht mit** — sonst wiese dieses Programm die eigene
	// Rückkehr ab, für die es [PartiebeitrittAntwort.Pausiert] gibt.
	schliessePartiesitzung(ctx, griff)

	// **Der Stand bleibt, die Verbindung geht.** Wer aufsteht, räumt ihn
	// selbst weg ([App.tischVergessen]); wem die Leitung wegbricht, dem
	// gehört sein Platz weiter — und die Partie fällt in den Stand, den
	// [App.tischVerloren] nennt.
	a.tischVerloren(lauf, ende)
}

// rueckkehrMelden schickt den Wunsch und gibt ihn hinaus.
//
// **Zwei Schritte und nicht einer**: [service.Partiesitzung.Sende] legt in die
// Warteschlange, hinaus geht es erst mit `Leere`. Genau dieser zweite Schritt
// hat in diesem Werkstück schon einmal gefehlt, und der Tisch blieb still.
//
// **Und geht er schief, endet die Sitzung.** Ohne Wunsch bekäme dieser Gast
// weder Bild noch Strom; er säße vor einem Tisch, der nie kommt — und das ist
// schlimmer als eine Leitung, die ehrlich zugeht (§8).
func (a *App) rueckkehrMelden(
	ctx context.Context,
	gast *service.Partiegast, griff service.Angenommene, partie string,
) {
	// **Hier beginnt die Messung der Runde** (G5, #1462) — sie schließt in
	// [App.Merke], weil dort jeder Rückweg vorbeikommt.
	a.gefragt(time.Now())

	fehler := gast.Kehrezurueck(ctx)
	if fehler == nil {
		fehler = griff.Leere(ctx)
	}

	if fehler == nil {
		return
	}

	// **Ging der Wunsch nicht hinaus, wird auch nichts gemessen.** Eine
	// offene Messung überlebte die Sitzung und schlösse beim nächsten Tisch
	// — mit einer Zahl, die die Zeit dazwischen mitzählt.
	a.gefragt(time.Time{})

	slog.WarnContext(ctx, "die Rückkehr ließ sich nicht anmelden",
		"partie", partie, "grund", fehler)

	a.sitzungAbweisen(ctx, griff, errRueckkehrGescheitert)
}

// errRueckkehrGescheitert ist der Grund, wenn der Wunsch nicht hinausging.
var errRueckkehrGescheitert = errors.New(
	"die Rückkehr ließ sich nicht anmelden")
