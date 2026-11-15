package app

// **Was mit einer hereinkommenden Partiesitzung geschieht** — Werkstück G5
// von Phase 5.5
// (#1462).
//
// # Die Stelle, an der alles zusammenkommt
//
// Der Horchposten nimmt an, die Freigabe sagt, wer wo sitzt, der Wirt legt
// Befehle aus, das Netz liest. **Zusammengesteckt wird hier**, und nur hier —
// das ist die Composition Root für eine Gastsitzung, und dieselbe Ausnahme
// wie bei `app.go` und `mcpnaht.go` (§2).
//
// # Warum eine fremde Sitzung ohne Freigabe geschlossen wird
//
// **Weil es sonst eine offene Leitung ohne Zweck wäre.** Der Gast wartet, und
// niemand antwortet ihm; er sähe eine Verbindung, die steht, und einen Tisch,
// der nie kommt. Eine Absage ist die ehrlichere Antwort (§8) — und sie ist
// **still**, denn ein Gegenüber, das anfragt, wenn nichts offen ist, ist kein
// Befund über diesen Rechner.
//
// # Und warum jede Sitzung ihre eigene Reihenfolge bekommt
//
// **Weil die Folgenummer dem Weg zu **einem** Gast gehört.** Zwei Gäste teilen
// sich den Ereignisstrom der Partie, aber nicht die Nummerierung ihrer
// Leitung: wer beiden aus **einem** Zähler zuteilte, ließe bei jedem Gast eine
// Lücke entstehen, sobald der andere etwas bekommt.

import (
	"context"
	"errors"
	"log/slog"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// erstePartieepoche ist die Epoche, in der eine frische Gastsitzung anfängt.
//
// **Eine Eins und kein gespeicherter Stand** — die Wiederaufnahme über
// Epochen hinweg steht noch aus, und eine Zahl, die nur so aussieht, als käme
// sie aus dem Bestand, wäre schlimmer als eine, die sichtbar frisch ist (§10).
const erstePartieepoche = 1

// partiesitzungAufnehmen macht aus einer angenommenen Sitzung einen Tisch.
//
// **Sie läuft in der Goroutine des Handschlags** — also je Gast eine, und
// nebenher. Die Schleife danach bekommt eine eigene, denn sie bleibt, solange
// gespielt wird.
func (a *App) partiesitzungAufnehmen(
	ctx context.Context, griff service.Angenommene,
) {
	stand := griff.Stand()

	if stand.Zweck != domain.ZweckPartie {
		return
	}

	// **Die Marke, bevor die Freigabe gelesen wird** (Nachlauf über Phase
	// 6.12, N1) — siehe [App.gastEintragen].
	marke := a.laufmarke(domain.LaufPartiefreigabe)

	offen, wirt := a.offeneFreigabe()
	if offen == nil {
		a.sitzungAbweisen(ctx, griff, errKeinePartieOffen)

		return
	}

	sitz, fehler := offen.Sitz(stand.Gegenueber)
	if fehler != nil {
		a.sitzungAbweisen(ctx, griff, fehler)

		return
	}

	partie := offen.Partie()

	sitzung := service.NeuePartiesitzung(
		partie, wirt, griff, service.NeueReihenfolge(erstePartieepoche))

	// **Der Platz wird hier eingetragen und nicht in der Schleife**, und das
	// ist eine Entscheidung mit einem Anlass.
	//
	// **Ohne die Runde könnte dieser Rechner nur antworten**: ein Satz vom
	// Wirt an seine Gäste hätte keinen Weg, und ein Satz eines Gastes
	// erreichte die anderen nie.
	//
	// **Und der erste Anlauf trug ihn in der Schleife ein** — also in einer
	// zweiten Goroutine, die erst später anlief. Zwischen „der Gast hat einen
	// Sitz" (`offen.Sitz`, hier) und „der Gast kann etwas hören" (die Runde,
	// dort) klaffte damit ein Fenster: wer in dieser Zeit etwas sagte, sagte
	// es in die Leere, **ohne Fehler**. Der Lauf über beide Enden hat fünf
	// Sekunden auf einen Zeiger gewartet, der nie kam. Beides geschieht
	// jetzt im selben Zug, und die Freigabe zeigt keinen Sitz, den noch
	// niemand bedient.
	// **Mit dem Peer, der hier sitzt** (P9, #2377): für seinen Bestand wird
	// ihm die Sicht gerechnet.
	//
	// **Und mit einer Bildpost** (P9b): bringt der Gast sein Deck mit, fragt
	// ihn dieser Rechner über den Platz nach Bildern, und die Antwort kommt
	// über den Wirt seines Sitzes.
	platz := service.NeuerGastplatz(sitz, sitzung, griff).
		MitPeer(string(stand.Gegenueber)).
		MitBildpost(&service.Bildpost{})
	if !a.gastEintragen(ctx, griff, platz, marke) {
		return
	}

	a.besetzungMelden(partie)

	// **Gezählt für den Schluss** (BK-07a): er wartet, bis diese Sitzung
	// endet, bevor der Bestand zugeht — sie liest und schreibt die Partie.
	a.lauf.kinder.Add(1)

	go func() {
		defer a.lauf.kinder.Done()

		a.partieBedienen(ctx, griff, partie, sitz, sitzung, platz)
	}()
}

// besetzungMelden sagt dem Fenster, dass sich an diesem Tisch die Besetzung
// geändert hat — gefunden mit zwei Fenstern (`nat04`, S12.NAT-04): die
// Freigabe des Wirts las ihren Stand nur auf „Nachsehen" und sagte „Noch
// sitzt niemand", während der Gast schon zog.
func (a *App) besetzungMelden(partie string) {
	if a.melden == nil {
		return
	}

	a.melden(a.fenster, EreignisPartiefreigabe, partie)
}

// gastEintragen setzt einen Gast in die Runde — und nimmt ihn wieder heraus,
// wenn `spielen` oder `netz` inzwischen aus ist oder dazwischen aus war. Der
// Rückgabewert sagt, ob er bleibt.
//
// **Nachlauf über Phase 6.12, N1**: die Aufnahme las die Freigabe, das
// Abschalten räumte sie und entließ die Runde, und danach trug die Aufnahme
// ihren Gast ein — der blieb, und sein Wirt nahm weiter Befehle an.
func (a *App) gastEintragen(
	ctx context.Context, griff service.Angenommene, platz *service.Gastplatz, marke uint64,
) bool {
	a.partie.runde.Setze(platz)

	return !a.nachDemStart(ctx, domain.LaufPartiefreigabe, marke, func(ctx context.Context) {
		a.partie.runde.Geht(platz)
		a.sitzungAbweisen(ctx, griff, errKeinePartieOffen)
	})
}

// partieBedienen führt die Schleife, bis die Leitung endet.
func (a *App) partieBedienen(
	ctx context.Context, griff service.Angenommene,
	partie, gast string,
	sitzung *service.Partiesitzung, platz *service.Gastplatz,
) {
	// **Weggeräumt wird nur der eigene Platz**, nicht der Sitz — ein Gast,
	// der neu verbindet, steht schon wieder da, wenn diese Schleife ausläuft
	// ([service.Partierunde.Geht]).
	// **Ein Zusammenbruch schließt die Sitzung** (BK-05a, #2111) — und nur
	// sie: die zwei Zeilen darunter geben Platz und Besetzung frei wie bei
	// jedem anderen Ende, und der Tisch des Wirts bleibt offen. **Der Fang
	// steht zuerst** und läuft damit zuletzt — so erreicht ihn auch ein
	// Zusammenbruch beim Aufräumen (Zweitblick über A und B, Befund 1).
	defer domain.Auffangen(func(zusammenbruch error) {
		zusammenbruchProtokollieren(ctx, "die Sitzung eines Gastes", zusammenbruch)
		schliessePartiesitzung(ctx, griff)
	})

	defer a.besetzungMelden(partie)
	defer a.partie.runde.Geht(platz)

	_, eigenerSitz := a.offeneFreigabe()

	// **Und hier hängt das Fenster dran.** Es hing bis zum 2026-09-16 nur auf
	// der **Gastseite** (`partiebeitritt.go`); auf dieser Seite fehlte der
	// Ruf, und damit verschwand jeder Satz eines Gastes still. Eine Naht, die
	// nur in eine Richtung trägt, ist eine halbe Zusage (§10).
	// **Und `MitFenster` seit dem 2026-09-19**
	// (#1627):
	// der Wirt spielt selbst mit, und was ein Gast stellt, erfuhr sein Fenster
	// bis dahin nicht — es zeigte den Tisch von vorhin, bis er selbst etwas
	// tat.
	nehmer := nebenherNehmer{Befehlsnehmer: a.partienaht(), geschrieben: a.nebenherGeschrieben}
	wirt := service.NeuerPartiewirt(partie, gast, nehmer, sitzung).
		MitGegenueber(string(griff.Stand().Gegenueber)).
		MitMaler(eigenmaler{a}).
		MitPlatzgeber(a.partienaht()).
		MitLeihgeber(a.leihstelle()).
		MitBildpost(platz.Bildpost()).
		MitNebenstrom(a.tischmelder(ctx, partie)).
		MitFenster(eigenerSitz, a.eigenerTischmelder(ctx, partie)).
		MitRunde(&a.partie.runde)

	// **Der Wirt spricht zuerst** (UX-08.AK-04, #1680): sein Tischbild nennt
	// dem Gast die Partie, die der sonst hätte abtippen müssen
	// ([service.Partieerfragen]). **Kommt es nicht hinaus, wird die Sitzung
	// niedergelegt** — ein Gast, den der Wirt nicht begrüßen kann, bekäme
	// auch auf seinen ersten Wunsch kein Bild.
	if fehler := a.begruessen(ctx, wirt, griff); fehler != nil {
		a.sitzungAbweisen(ctx, griff, fehler)

		return
	}

	netz := service.NeuesPartienetz(
		griff, service.NeuerPartieempfang(erstePartieepoche), wirt,
	).MitMeldung(func(fehler error) {
		slog.WarnContext(ctx,
			"eine Nachricht eines Gastes wurde nicht angenommen",
			"partie", partie, "grund", fehler)
	})

	// **Ein Ende ist kein Fehlschlag.** Eine Leitung, die zugeht, ist der
	// Normalfall; was daraus für die Partie folgt, entscheidet der Gast
	// (`App.tischVerloren`) — und dieser Rechner **ist** der Wirt.
	//
	// **Und es endet immer mit einem Fehler.** `Bediene` kehrt nur zurück,
	// wenn es einen gibt: der Kontext, die Leitung, eine Lücke oder der
	// Bote. Ein `if fehler != nil` wäre eine Bedingung, die immer zutrifft —
	// `staticcheck` hat es angestrichen, und der Zweig darunter war nie ein
	// Zweig.
	slog.InfoContext(ctx, "die Sitzung eines Gastes ist zu Ende",
		"partie", partie, "grund", netz.Bediene(ctx))

	// **Und dann wird sie freigegeben.** Bis zum Messlauf über Phase 6.7
	// blieb der Griff in der Übersicht, und jeder neue Versuch desselben
	// Gastes scheiterte an „läuft schon eine Sitzung" — wer aufstand und sich
	// wieder setzte, kam nie wieder an den Tisch.
	schliessePartiesitzung(ctx, griff)
}

// schliessePartiesitzung legt eine Partiesitzung nieder, deren Schleife
// endete — auf beiden Seiten des Tisches.
//
// **Ein Ende ist kein Fehlschlag**, und eine Leitung, die sich nicht
// schließen lässt, ändert daran nichts; die Klage steht im Protokoll.
// **Und zweimal schließen schadet nicht**: wer aufsteht, hat schon
// geschlossen, und der Griff weiß das.
func schliessePartiesitzung(ctx context.Context, sitzung leitungsende) {
	if fehler := sitzung.Schliesse(); fehler != nil {
		slog.WarnContext(ctx,
			"eine Partiesitzung ließ sich nicht schließen", "grund", fehler)
	}
}

// errKeinePartieOffen ist der Grund, wenn gar nichts freigegeben ist.
var errKeinePartieOffen = errors.New("keine Partie steht für Gäste offen")

// sitzungAbweisen legt eine Sitzung nieder, für die es nichts zu tun gibt.
//
// **Zwei Zeilen und nicht eine**: die Abweisung ist der Normalfall und steht
// als Auskunft da; dass das Niederlegen **selbst** schiefging, ist etwas
// anderes und wiegt schwerer.
//
// **Und kein Sitzname im Protokoll** — den tippt ein Mensch, und was er
// hineinschreibt, entscheidet er (siehe `protokollmerkmale.go`).
func (a *App) sitzungAbweisen(
	ctx context.Context, griff service.Angenommene, grund error,
) {
	slog.InfoContext(ctx, "eine Gastsitzung wurde abgewiesen",
		"grund", grund)

	if fehler := griff.Schliesse(); fehler != nil {
		slog.WarnContext(ctx,
			"eine abgewiesene Gastsitzung ließ sich nicht schließen",
			"grund", fehler)
	}
}
