package service

// **Die Handelsleitung** — Werkstück `T1` von Phase 5.6
// (#1466).
//
// # Was sie tut und was ausdrücklich nicht
//
// **Sie hält den Stand und reicht die Entscheidung an die Domäne.** Ob ein
// Vorschlag anschließt, entscheidet [domain.Handelsangebot.Nimmt]; ob jemand
// einen Schritt gehen darf, entscheidet [domain.Handelsangebot.Schreite].
// Hier steht, **wann geschrieben wird** — und vor allem, wann nicht.
//
// # Und das ist die Abnahme des Issues
//
// **„Ohne Bestandswirkung."** Jede Ablehnung geschieht, bevor irgendetwas
// geschrieben wurde; die Läufe messen es an einem Doppel, das seine Schreibe
// zählt, statt es zu beteuern.

import (
	"context"
	"errors"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Handelsablage ist, was die Leitung vom Bestand braucht.
//
// **Die Schnittstelle steht hier und nicht in `storage`** (§2): sie
// beschreibt, was dieser Anwendungsfall verlangt, nicht was eine Datenbank
// anbietet.
type Handelsablage interface {
	// Vorgang gibt die **letzte** Revision eines Vorgangs.
	//
	// **Die letzte und nicht alle**: was gilt, ist eine Frage mit einer
	// Antwort. Wer die Kette braucht — ein Protokoll, ein Streit —, holt sie
	// über T5 (#1470).
	Vorgang(ctx context.Context, kennung string) (domain.Handelsangebot, error)

	// Protokoll gibt die Zustandswechsel eines Vorgangs (TRADE-03.AK-03).
	Protokoll(ctx context.Context, vorgang string) ([]domain.Handelsprotokollzeile, error)

	// Fortschreiben legt `danach` ab — nur, wenn `vorher` noch der Stand ist
	// (TRADE-03.AK-01, #1717). Sonst `ErrUeberholt`: ein zweiter Schreiber
	// war schneller, und seine Änderung geht nicht still verloren. **Ein
	// leerer Vorstand legt nur an**, was es noch nicht gibt — der einzige
	// Weg, auf dem hier ein Vorgang entsteht (Befund N2 des Nachlaufs über
	// Phase 6.8; ein unbedingtes `Schreiben` gibt es für die Leitung nicht
	// mehr).
	Fortschreiben(ctx context.Context, vorher, danach domain.Handelsangebot) error

	// Letzte gibt je Vorgang die letzte Revision, zuletzt geänderte zuerst —
	// für die Wahl eines Handels (UX-08.AK-04, [Handelsleitung.Vorgaenge]).
	Letzte(ctx context.Context) ([]domain.Handelsangebot, error)
}

// Zusagenwirkung ist, was ein Schritt an den festgehaltenen Exemplaren
// bewirkt (#1643).
//
// **Eine eigene Schnittstelle und nicht [Zusagenablage]**: die Leitung hält
// nichts fest — das tut ein Mensch, der eine Besitzzeile benennt. Sie liest,
// vermerkt und gibt frei.
type Zusagenwirkung interface {
	FuerVorgang(ctx context.Context, vorgang string) ([]domain.Reservierung, error)
	Uebergeben(ctx context.Context, vorgang string) error
	Freigeben(ctx context.Context, vorgang string) error
}

// Handelsleitung führt die Handelsvorgänge dieses Rechners.
type Handelsleitung struct {
	ablage  Handelsablage
	zusagen Zusagenwirkung

	// kennungen vergibt die Kennung eines neuen Handels, den niemand benennt
	// — [handelskennung], im Test eingesetzt ([Handelsleitung.MitKennungen]).
	kennungen func() string

	// schranke ist die Handelsrichtlinie (#2435, `handelsschranke.go`).
	schranke Handelsschranke
}

// NeueHandelsleitung baut sie über einer Ablage und den Zusagen.
//
// **Die Zusagen sind nicht wahlweise dabei.** Ein Schritt, dessen Wirkung an
// den Exemplaren von einer gesetzten Naht abhinge, hielte die halbe Zusage
// (§10) — und die Hälfte, die fehlt, wäre die, die den Bestand bewegt.
func NeueHandelsleitung(
	ablage Handelsablage, zusagen Zusagenwirkung,
) *Handelsleitung {
	return &Handelsleitung{ablage: ablage, zusagen: zusagen, kennungen: handelskennung, schranke: ohneEinstellung{}}
}

// MitHandelsschranke setzt die Handelsrichtlinie, die gilt (#2435).
func (l *Handelsleitung) MitHandelsschranke(schranke Handelsschranke) *Handelsleitung {
	l.schranke = schranke

	return l
}

// Eroeffnen legt die erste Revision eines Vorgangs an.
//
// **Ein zweites Eröffnen ist ein Konflikt und kein Überschreiben.** Wer einen
// laufenden Vorgang noch einmal eröffnete, wäre bei Revision 1 — und alles,
// was inzwischen verhandelt wurde, hätte nie stattgefunden.
func (l *Handelsleitung) Eroeffnen(
	ctx context.Context, angebot domain.Handelsangebot,
) (domain.Handelsangebot, error) {
	// **Wer keine Kennung nennt, bekommt eine** (UX-08.AK-04, #1680) —
	// eine Transaktionskennung zu tippen ist nichts, was jemand tun muss,
	// der einen Handel beginnt.
	if strings.TrimSpace(angebot.Vorgang) == "" {
		angebot.Vorgang = l.kennungen()
	}

	if fehler := angebot.Pruefe(); fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	if fehler := angebotErlaubt(ctx, l.schranke, angebot); fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	if angebot.Revision != 1 {
		return domain.Handelsangebot{}, domain.MitGrund(domain.ErrValidation,
			nil, "ein Vorgang wird mit der ersten Revision eröffnet")
	}

	_, fehler := l.ablage.Vorgang(ctx, angebot.Vorgang)

	switch {
	case fehler == nil:
		return domain.Handelsangebot{}, domain.MitGrund(domain.ErrConflict,
			nil, "diesen Handelsvorgang gibt es schon")

	case !errors.Is(fehler, domain.ErrNotFound):
		return domain.Handelsangebot{}, fehler
	}

	// **Angelegt wird nur, was es noch nicht gibt** (Befund N2 des Nachlaufs
	// über Phase 6.8): zwei Eröffnungen, die beide „gibt es nicht" gelesen
	// hatten, bekamen sonst beide Erfolg, und die zweite ersetzte die erste.
	if fehler := l.ablage.Fortschreiben(ctx, domain.Handelsangebot{}, angebot); fehler != nil {
		if errors.Is(fehler, domain.ErrUeberholt) {
			return domain.Handelsangebot{}, domain.MitGrund(domain.ErrConflict,
				nil, "diesen Handelsvorgang gibt es schon")
		}

		return domain.Handelsangebot{}, fehler
	}

	return angebot, nil
}

// Vorschlagen trägt die nächste Revision ein — oder sagt, warum nicht.
//
// **Derselbe Vorschlag zweimal ist einmal.** Über eine Leitung geht eine
// Nachricht verloren und kommt noch einmal; wer den Stand schon **ist**,
// bekommt ihn zurück und keinen Konflikt (§7, Regel 4).
//
// **`wer` ist, wer vorschlägt** — aus `angeboten` nur die Gegenseite
// ([domain.Handelsangebot.NimmtVorschlagVon], #1975).
func (l *Handelsleitung) Vorschlagen(
	ctx context.Context, vorschlag domain.Handelsangebot, wer string,
) (domain.Handelsangebot, error) {
	heute, fehler := l.ablage.Vorgang(ctx, vorschlag.Vorgang)
	if fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	if heute.Abdruck() == vorschlag.Abdruck() {
		return heute, nil
	}

	if fehler := heute.NimmtVorschlagVon(vorschlag, wer); fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	if fehler := angebotErlaubt(ctx, l.schranke, vorschlag); fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	if fehler := l.ablage.Fortschreiben(ctx, heute, vorschlag); fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	return vorschlag, nil
}

// Schreiten geht einen Schritt an einem Vorgang.
//
// # Drei Schritte gehen hier nicht
//
// **Weil sie mehr bewegen als den Zustand** ([domain.Abschlussschritte]).
// `verbuchen` bewegt den Bestand, `bestaetigen` zeichnet den Beleg gegen,
// `abgleichen` steht zwischen beiden — sie gehören der
// [Verbuchungsstelle], und sie verlangen dort einen **Nachweis** (T10,
// #1475).
//
// **Ginge einer hier durch, wäre der Riegel umgangen** und der Zustand
// gesetzt, ohne dass geschehen wäre, wofür er steht:
// „lokalVerbucht" ohne eine Karte, die den Besitzer gewechselt hat. Die
// Absage nennt deshalb den Weg, der es kann
// (#1637).
func (l *Handelsleitung) Schreiten(
	ctx context.Context, vorgang string,
	schritt domain.Handelsschritt, wer string,
) (domain.Handelsangebot, error) {
	return l.SchreitenAls(ctx, vorgang, "", schritt, wer, wer)
}

// SchreitenAls geht einen Schritt und weiß dabei, wer **dieser Rechner** ist.
//
// **Die Deckung ist eine Frage an den eigenen Bestand** (`vorbereiten`,
// `uebergeben`). Eine Stimme des Gegenübers, die hier eingetragen wird, hat
// ihre Exemplare drüben; gegen die Zusagen hier gezählt, stünde sie immer
// bei null (Hüllenprobe `nat04handel`, #1672). [Handelsleitung.Schreiten]
// nimmt die Stimme als die eigene — so, wie es bis dahin für jede galt.
//
// **`gesehen` ist der Abdruck, den der Schritt meint** (#2496): ist der
// Stand inzwischen eine andere Revision, geht er nicht. Leer heißt: nicht
// gefragt.
func (l *Handelsleitung) SchreitenAls(
	ctx context.Context, vorgang, gesehen string,
	schritt domain.Handelsschritt, wer, ich string,
) (domain.Handelsangebot, error) {
	if domain.IstAbschlussschritt(schritt) {
		return domain.Handelsangebot{}, domain.MitGrund(domain.ErrValidation,
			nil, "%q gehört zur Verbuchung und geht nicht über die Leitung",
			schritt)
	}

	heute, fehler := l.ablage.Vorgang(ctx, vorgang)
	if fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	if ueberholt := heute.PruefeGesehen(gesehen); ueberholt != nil {
		return domain.Handelsangebot{}, ueberholt
	}

	danach, fehler := heute.Schreite(schritt, wer)
	if fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	// **Ein Schritt, der nichts ändert, schreibt nicht** — die Wiederholung
	// einer Zustimmung ist der Fall, für den es diese Zeile gibt.
	if danach.Abdruck() == heute.Abdruck() &&
		danach.Zustand == heute.Zustand &&
		len(danach.Zustimmungen) == len(heute.Zustimmungen) {
		return heute, nil
	}

	// **Nur auf dem Weg zur Einigung** — die Auswege bleiben offen
	// (`handelsschranke.go`).
	if fehler := einigungErlaubt(ctx, l.schranke, danach); fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	// **Und nur über Zeilen, die drüben zu buchen sind** (#2495) — auf
	// demselben Weg und mit derselben Ausnahme für die Auswege.
	if fehler := einigungVerbuchbar(danach); fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	// **Die Wirkung an den Exemplaren steht vor dem Schreiben** (#1643).
	// Scheitert sie, ist nichts geschehen — dieselbe Zusage wie im Kopf
	// dieser Datei.
	if fehler := l.zusagenwirkung(
		ctx, vorgang, schritt, wer == ich, wer, heute, danach); fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	if fehler := l.ablage.Fortschreiben(ctx, heute, danach); fehler != nil {
		return domain.Handelsangebot{}, fehler
	}

	return danach, nil
}

// zusagenwirkung führt aus, was dieser Schritt an den Exemplaren bedeutet.
//
// # Drei Schritte, drei Bedeutungen
//
//   - `vorbereiten` heißt „die Exemplare sind reserviert" — also wird
//     geprüft, ob festgehalten ist, was von hier weggeht.
//   - `uebergeben` heißt „sie sind über den Tisch gegangen" — also wird es
//     vermerkt, und erst dadurch findet die Verbuchung ihre Abgangszeilen.
//     **Und vorher geprüft wie beim Vorbereiten**: `vorbereiten` geht nur der
//     Initiator, und eine Gegenseite, die etwas weggibt, bestätigte sonst
//     eine Übergabe ohne Deckung — das Verbuchen scheiterte danach an einem
//     Zustand, aus dem nur die Klärung führt (Hüllenprobe `nat04handel`,
//     #1672).
//   - `ablehnen`, `abbrechen` und `ablaufen` beenden den Handel — also wird
//     freigegeben, was niemand mehr braucht.
//
// **Alles andere lässt die Zusagen in Ruhe**, und das ist die Mehrzahl.
//
// # Die Wirkung gehört zum Wechsel und nicht zur Stimme
//
// **`uebergeben` ist ein Beschluss** (`T1`): die erste Stimme trägt sich ein
// und der Vorgang steht, wo er stand. Wer schon dort vermerkte, sagte „ist
// über den Tisch gegangen", während der andere noch gar nichts getan hat —
// und die Verbuchung fände Abgangszeilen für eine Übergabe, die halb ist.
// Dasselbe gilt für das Freigeben.
func (l *Handelsleitung) zusagenwirkung(
	ctx context.Context, vorgang string, schritt domain.Handelsschritt,
	eigen bool, wer string, heute, danach domain.Handelsangebot,
) error {
	if schritt == domain.SchrittVorbereiten {
		if !eigen {
			return nil
		}

		return l.deckungPruefen(ctx, vorgang, wer, heute)
	}

	if schritt == domain.SchrittUebergeben && eigen {
		if fehler := l.deckungPruefen(ctx, vorgang, wer, heute); fehler != nil {
			return fehler
		}
	}

	// **Über dieselbe Stelle wie die Leitung von drüben** (Befund 3 des
	// Zweitblicks über Phase 5.6): derselbe Wechsel bedeutet dasselbe,
	// gleich ob ihn ein Schritt hier oder eine Nachricht dort auslöst.
	return zusagenNachWechsel(
		ctx, l.zusagen, vorgang, heute.Zustand, danach.Zustand)
}

// deckungPruefen hält fest, dass für diese Seite genug festgehalten ist.
//
// **Wer nichts weggibt, braucht nichts.** Ein Handel, bei dem alles von
// drüben kommt, hat null wegzugebende Stücke — eine Bedingung, die auch dann
// zuschlüge, machte ihn unmöglich.
func (l *Handelsleitung) deckungPruefen(
	ctx context.Context, vorgang, wer string, heute domain.Handelsangebot,
) error {
	offen := 0

	for _, stelle := range heute.Positionen {
		if stelle.Von == wer {
			offen += stelle.Menge
		}
	}

	if offen == 0 {
		return nil
	}

	zusagen, fehler := l.zusagen.FuerVorgang(ctx, vorgang)
	if fehler != nil {
		return fehler
	}

	gehalten := 0
	for _, zusage := range zusagen {
		gehalten += zusage.Menge
	}

	if gehalten < offen {
		return domain.MitGrund(domain.ErrConflict, nil,
			"festgehalten sind %d von %d Stück — wähle zuerst die Exemplare "+
				"aus, die weggehen", gehalten, offen)
	}

	return nil
}

// Stand gibt die letzte Revision eines Vorgangs.
func (l *Handelsleitung) Stand(
	ctx context.Context, vorgang string,
) (domain.Handelsangebot, error) {
	return l.ablage.Vorgang(ctx, vorgang)
}

// Protokoll gibt, wie ein Vorgang dorthin kam, wo er steht — ältester
// Wechsel zuerst (TRADE-03.AK-03, #1717).
func (l *Handelsleitung) Protokoll(
	ctx context.Context, vorgang string,
) ([]domain.Handelsprotokollzeile, error) {
	return l.ablage.Protokoll(ctx, vorgang)
}
