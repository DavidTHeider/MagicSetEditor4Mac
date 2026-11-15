package service

// **Das Auftragswerk** — Werkstück A6 von Phase 5.3
// (#1445).
//
// # Drei Schritte, und der erste ändert nichts
//
// **Anlegen** friert einen Stand ein und reserviert Budget. **Vorschlagen**
// prüft gegen das Schema und legt ab. **Übernehmen** schreibt — und prüft
// vorher noch einmal alles, was sich inzwischen geändert haben kann.
//
// Das Issue sagt den Satz, um den herum das gebaut ist: „**Generieren
// verändert Zielkarte nicht.**"
//
// # Warum die Prüfung zweimal stattfindet
//
// **Weil zwischen Vorschlag und Übernahme Zeit liegt.** Die Karte kann sich
// geändert haben, die Frist kann abgelaufen sein, das Paket kann ein anderes
// sein. Wer nur beim Vorschlagen prüfte, schriebe später gegen einen Stand,
// den niemand angesehen hat — dieselbe Überlegung wie §7, Regel 2.

import (
	"context"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// KIKarten ist, was das Werk von der Kartenablage braucht.
//
// **Zwei Methoden**: den Stand lesen und Felder setzen. Mehr zu verlangen
// hieße, dass ein Auftragswerk eine Karte anlegen oder löschen könnte.
type KIKarten interface {
	// Revision gibt den Stand einer Karte — das, was sich ändert, wenn
	// jemand etwas an ihr tut.
	Revision(ctx context.Context, karte int64) (string, error)

	// FelderSetzen schreibt Werte **und ihre Herkunft in einem Zug**.
	//
	// # Warum beides in einem Ruf und nicht in zweien
	//
	// **Weil das Issue es verlangt** (A11, #1450): „Scheitert
	// Provenienz/Verlauf in derselben Transaktion, stehen keine halben
	// Inhaltsänderungen." Zwei Rufe hießen zwei Transaktionen — und
	// zwischen ihnen ein Stand, in dem die Karte einen KI-Wert trägt, der
	// nirgends als solcher steht.
	//
	// **Die Zusage einzulösen ist Sache der Ablage**, nicht dieses Werks:
	// hier steht, dass es zusammengehört, dort liegt die Transaktion.
	//
	// **`erwartet` ist der Stand, auf dem die Karte stehen muss** — und die
	// Ablage hält ihn in **derselben** Transaktion (Befund 2 des Zweitblicks
	// über Phase 5.3). [KIAuftragswerk.pruefeStand] davor bleibt: er gibt
	// den Satz, der erklärt, was schiefging, und prüft außerdem das Schema.
	// Was er nicht kann, ist die Lücke bis zum Schreiben schließen.
	FelderSetzen(ctx context.Context, karte int64, felder map[string][]string,
		herkunft []domain.KIFeldherkunft, erwartet string) error

	// FelderGruppeSetzen schreibt **mehrere Karten in einem Zug** (A13.1b,
	// #1452).
	//
	// # Warum das nicht [KIKarten.FelderSetzen] in einer Schleife ist
	//
	// **Weil eine Schleife N Transaktionen sind.** Zwischen der dritten und
	// der vierten stünde ein Bestand, in dem drei Karten den Vorschlag
	// tragen und zwei nicht — genau der Teilbestand, den das Issue
	// ausschließt („entweder vollständig … oder klarer Konflikt ohne
	// Teilbestand").
	//
	// **Die Zusage einzulösen ist Sache der Ablage**, so wie nebenan: hier
	// steht, dass es zusammengehört, dort liegt die Transaktion.
	FelderGruppeSetzen(ctx context.Context, schritte []KIGruppenschritt) error
}

// KIBudget reserviert, bevor gefragt wird.
//
// **Vor dem Ruf und nicht danach**: wer erst nach der Antwort bucht, hat bei
// zwei gleichzeitigen Aufträgen zweimal die Grenze gelesen und keinmal
// gehalten (O12, #1439).
type KIBudget interface {
	Reservieren(ctx context.Context, kennung string) (string, error)
}

// KISchemaquelle gibt das Schema einer Karte.
type KISchemaquelle interface {
	Schema(ctx context.Context, karte int64) (domain.KISchema, error)
}

// KIOrtsregel ist, was die Einstellungsebenen zusammen erlauben.
//
// # Die restriktivste gilt
//
// Das Issue: „restriktivste Local-only-/Cloud-/Profil-/Maschinenregel gilt".
// Wer sie hier zusammenrechnet, hat **eine** Frage zu beantworten — und nicht
// vier, von denen eine vergessen wird.
type KIOrtsregel struct {
	// NurLokal heißt: kein Modell, das anderswo rechnet.
	NurLokal bool
}

// KIOrtsregelAus liest die Regel aus dem Stand der Einstellungen.
//
// # Es ist die Richtlinie aus 5.2 und keine zweite
//
// **Das Issue verlangt es**: „Einstellungsebenen von 5.2 **konsumieren**" und
// „kein zweites Authsystem". `richtlinie.nurlokal` gibt es seit O-… und sie
// sagt schon, was sie hier sagen soll — „schaltet jeden fremden Dienst ab".
//
// **Eine eigene Einstellung `ki.ort` wären zwei Schalter für dieselbe
// Frage**, und einer davon bliebe beim nächsten Umbau stehen.
//
// **Die Faltung über die Ebenen macht `domain` schon** ([domain.Geltung]):
// was hier ankommt, ist der **wirksame** Wert, und „restriktivste Regel gilt"
// ist damit beantwortet, bevor diese Funktion läuft.
func KIOrtsregelAus(wirksam string) KIOrtsregel {
	return KIOrtsregel{NurLokal: wirksam == domain.NurLokalAn}
}

// Erlaubt sagt, ob ein Modell unter dieser Regel gefragt werden darf.
//
// **Entschieden wird am erklärten Ort und nicht an der Adresse.** Das Issue:
// „**Loopback allein beweist keine lokale Inferenz**." Ein Dienst auf
// `127.0.0.1` kann weiterleiten — dann steht die Adresse hier und rechnet
// jemand anderes.
func (r KIOrtsregel) Erlaubt(ort domain.KIOrt) bool {
	return !r.NurLokal || ort == domain.KIOrtLokal
}

// KIWunsch ist, was jemand von der KI will.
type KIWunsch struct {
	ProfilID   int64
	Faehigkeit domain.KIFaehigkeit
	Anbieter   string
	Alias      string
	Zielkarte  int64
}

// KIUebernahme sagt, was übernommen wurde.
type KIUebernahme struct {
	// Felder ist die Zahl der geschriebenen Felder.
	Felder int

	// Namen sind die geschriebenen Felder — sortiert (A11, #1450).
	//
	// **Neben der Zahl und nicht statt ihrer**: wer auswählt, will danach
	// lesen, **was** geschrieben wurde, und nicht nur wie viel.
	Namen []string
}

// KIFrist ist, wie lange ein Vorschlag übernehmbar bleibt.
//
// **Vierundzwanzig Stunden**: lang genug, um eine Nacht darüber zu schlafen,
// kurz genug, dass ein Stand von vorgestern nicht mehr trifft.
const KIFrist = 24 * time.Hour

// KIAuftragswerk führt die Aufträge dieses Prozesses.
type KIAuftragswerk struct {
	katalog *Modellkatalog
	karten  KIKarten
	budget  KIBudget
	schemen KISchemaquelle
	regel   KIOrtsregel
	jetzt   func() time.Time

	// buch hält Auftrag und Vorschlag, bis übernommen oder verfallen.
	//
	// **Im Prozess und nicht im Bestand.** Ein Vorschlag ist kein Besitz: er
	// gilt eine Frist lang, gehört einem Profil und verschwindet mit dem
	// Fenster. Wer ihn ablegte, hätte eine Tabelle voller abgelaufener
	// Vorschläge und eine Aufräumfrage dazu.
	//
	// **Ein Zeiger und kein Feld.** Das Buch führt sein eigenes Schloss und
	// wird von den Werken einer Naht **geteilt**; warum das Teilen nötig ist
	// und ein Kopieren nicht genügt, steht bei [kibuch].
	buch *kibuch
}

// kiLauf ist ein Auftrag samt dem, was aus ihm wurde.
type kiLauf struct {
	auftrag   domain.KIAuftrag
	vorschlag *domain.KIVorschlag
}

// NeuesKIAuftragswerk baut es.
func NeuesKIAuftragswerk(
	katalog *Modellkatalog, karten KIKarten, budget KIBudget,
	schemen KISchemaquelle, regel KIOrtsregel, jetzt func() time.Time,
) *KIAuftragswerk {
	return &KIAuftragswerk{
		katalog: katalog, karten: karten, budget: budget, schemen: schemen,
		regel: regel, jetzt: jetzt, buch: neuesBuch(),
	}
}

// Anlegen friert einen Stand ein und reserviert Budget.
//
// **Es schreibt nichts.** Was hier entsteht, ist ein Auftrag mit einer
// Kennung — und die Karte bleibt, wie sie war.
func (w *KIAuftragswerk) Anlegen(ctx context.Context, wunsch KIWunsch) (
	domain.KIAuftrag, error,
) {
	modell, fehler := w.modellFuer(wunsch)
	if fehler != nil {
		return domain.KIAuftrag{}, fehler
	}

	revision, fehler := w.karten.Revision(ctx, wunsch.Zielkarte)
	if fehler != nil {
		return domain.KIAuftrag{}, fehler
	}

	schema, fehler := w.schemen.Schema(ctx, wunsch.Zielkarte)
	if fehler != nil {
		return domain.KIAuftrag{}, fehler
	}

	jetzt := w.jetzt()
	kennung := fmt.Sprintf("ki-%d-%d", wunsch.Zielkarte, jetzt.UnixNano())

	posten, fehler := w.budget.Reservieren(ctx, kennung)
	if fehler != nil {
		return domain.KIAuftrag{}, fehler
	}

	auftrag := domain.KIAuftrag{
		Kennung: kennung, ProfilID: wunsch.ProfilID,
		Faehigkeit: wunsch.Faehigkeit,
		Anbieter:   modell.Anbieter, Alias: modell.Alias, Ort: modell.Ort,
		Zielkarte: wunsch.Zielkarte, Zielrevision: revision,
		Schemakennung: schema.Kennung(), Kostenposten: posten,
		Angelegt: jetzt, Ablauf: jetzt.Add(KIFrist),
	}

	if befunde := auftrag.Validate(); len(befunde) > 0 {
		return domain.KIAuftrag{}, befunde[0]
	}

	w.buch.anlegen(kennung, &kiLauf{auftrag: auftrag})

	return auftrag, nil
}

// Pruefen sagt, ob dieses Modell die Fähigkeit hat und hier rechnen darf —
// **ohne einen Auftrag anzulegen**.
//
// # Warum es das gibt
//
// **Weil nicht jede Frage an ein Modell eine Zielkarte hat** (A13.3, #1452).
// [KIAuftragswerk.Anlegen] friert Revision und Schema **einer Karte** ein;
// eine Bilderkennung hat keine — sie liest ein Foto und schreibt nirgends.
//
// **Was sie trotzdem braucht, ist diese Prüfung**: „vision belegt" und „darf
// hier rechnen" sind Regeln des Werks, und eine zweite Fassung daneben liefe
// beim ersten neuen Ort auseinander.
//
// **Die Buchung gehört nicht dazu.** Wer bucht, ist der Aufrufer — genauso
// wie beim Batch, der **einen** Posten für viele Ziele führt
// ([0105](../../docs/entscheidungen/0105-eine-freigabe-fuer-einen-batch.md)).
func (w *KIAuftragswerk) Pruefen(wunsch KIWunsch) (domain.KIModell, error) {
	return w.modellFuer(wunsch)
}

// modellFuer sucht das Modell und prüft Fähigkeit und Ort.
func (w *KIAuftragswerk) modellFuer(wunsch KIWunsch) (domain.KIModell, error) {
	for _, modell := range w.katalog.Alle() {
		if modell.Anbieter != wunsch.Anbieter || modell.Alias != wunsch.Alias {
			continue
		}

		if !modell.Kann(wunsch.Faehigkeit) {
			return domain.KIModell{}, fmt.Errorf(
				"%w: %q hat %q nicht belegt — eine Liste ist keine Fähigkeit",
				domain.ErrValidation, modell.Alias, string(wunsch.Faehigkeit))
		}

		if !w.regel.Erlaubt(modell.Ort) {
			return domain.KIModell{}, fmt.Errorf(
				"%w: %q rechnet %s, und hier gilt »nur lokal«",
				domain.ErrValidation, modell.Alias, string(modell.Ort))
		}

		return modell, nil
	}

	return domain.KIModell{}, fmt.Errorf(
		"%w: den Alias %q gibt es bei %q nicht",
		domain.ErrNotFound, wunsch.Alias, wunsch.Anbieter)
}

// Vorschlagen prüft einen Vorschlag gegen das Schema und legt ihn ab.
//
// **Es schreibt nichts** — auch dann nicht, wenn alles passt.
func (w *KIAuftragswerk) Vorschlagen(ctx context.Context, vorschlag domain.KIVorschlag) (
	[]domain.Befund, error,
) {
	lauf, gibts := w.lauf(vorschlag.Auftrag)
	if !gibts {
		return nil, fmt.Errorf("%w: den Auftrag %q gibt es nicht",
			domain.ErrNotFound, vorschlag.Auftrag)
	}

	// **Die Zugehörigkeit steht schon fest**: gegriffen wurde über
	// `vorschlag.Auftrag`, und was dort liegt, ist der Auftrag dieser
	// Kennung. Eine zweite Prüfung wäre ein Zweig, den nichts betritt (§10).

	schema, fehler := w.schemen.Schema(ctx, lauf.auftrag.Zielkarte)
	if fehler != nil {
		return nil, fehler
	}

	befunde := schema.Pruefe(vorschlag.Felder)

	// **Abgelegt wird, was durchgekommen ist** — und die Befunde stehen
	// daneben. Ein Vorschlag, der halb passt, ist kein Grund, die gute
	// Hälfte wegzuwerfen; ein Grund, sie stillschweigend zu übernehmen, ist
	// er auch nicht.
	gut := domain.KIVorschlag{
		Auftrag: vorschlag.Auftrag,
		Felder:  ohneBeanstandete(vorschlag.Felder, befunde),

		// **Modell und Aufgabe reisen mit** (A11, #1450) — aus ihnen
		// entsteht bei der Übernahme die Herkunft, und später weiß sie
		// niemand mehr.
		Modell: vorschlag.Modell, Aufgabe: vorschlag.Aufgabe,
	}

	w.buch.vermerke(lauf, &gut)

	return befunde, nil
}

// pruefeStand hält Karte und Schema gegen den eingefrorenen Stand.
func (w *KIAuftragswerk) pruefeStand(ctx context.Context, auftrag domain.KIAuftrag) error {
	revision, fehler := w.karten.Revision(ctx, auftrag.Zielkarte)
	if fehler != nil {
		return fehler
	}

	if revision != auftrag.Zielrevision {
		return fmt.Errorf(
			"%w: die Karte hat sich seit dem Auftrag geändert — der Vorschlag "+
				"gehört zu einem Stand, den es so nicht mehr gibt",
			domain.ErrConflict)
	}

	schema, fehler := w.schemen.Schema(ctx, auftrag.Zielkarte)
	if fehler != nil {
		return fehler
	}

	if schema.Kennung() != auftrag.Schemakennung {
		return fmt.Errorf(
			"%w: die Felder dieses Pakets haben sich seit dem Auftrag "+
				"geändert — der Vorschlag wurde gegen andere geprüft",
			domain.ErrConflict)
	}

	return nil
}
