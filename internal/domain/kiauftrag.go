package domain

// **Ein KI-Auftrag und sein Vorschlag** — Werkstück A6 von Phase 5.3
// (#1445).
//
// # Generieren ändert nichts
//
// **Das ist der Satz, um den herum dieser Typ gebaut ist.** Das Issue sagt
// ihn wörtlich: „Generieren verändert Zielkarte nicht." Ein Auftrag erzeugt
// einen **Vorschlag** mit einer Kennung; was daraus wird, entscheidet ein
// zweiter, eigener Schritt — und der prüft noch einmal.
//
// # Und er ist an einen Stand gebunden
//
// **An fünf Dinge zugleich**: Profil, Zielkarte, Zielrevision, Schemafassung
// und eine Frist. Fehlt eines, träfe ein Vorschlag irgendetwas: das falsche
// Profil, eine inzwischen geänderte Karte, ein Paket mit anderen Feldern oder
// einen Stand von vorgestern. Das Issue zählt sie deshalb einzeln auf.

import "time"

// KIOrt sagt, wo ein Modell rechnet.
type KIOrt string

// Die zwei.
const (
	// KIOrtLokal ist auf diesem Rechner.
	KIOrtLokal KIOrt = "lokal"

	// KIOrtCloud ist bei einem fremden Anbieter.
	KIOrtCloud KIOrt = "cloud"
)

var kiOrte = []KIOrt{KIOrtLokal, KIOrtCloud}

// IstKIOrt sagt, ob es den Ort gibt.
func IstKIOrt(ort KIOrt) bool { return kennt(kiOrte, ort) }

// KIAuftrag ist ein Auftrag an ein Modell — festgemacht an einem Stand.
type KIAuftrag struct {
	// Kennung ist die **konsumierbare Vorschlags-ID** aus dem Issue: sie
	// benennt diesen Auftrag und den Vorschlag, der aus ihm wird.
	Kennung string

	ProfilID int64

	Faehigkeit KIFaehigkeit

	// Anbieter und Alias sagen, wer gefragt wird — zusammen, wie im Katalog.
	Anbieter string
	Alias    string

	// Ort ist, **wo** gerechnet wird, wie das Modell es erklärt.
	//
	// **Er steht am Auftrag und wird nicht aus der Adresse geschlossen.** Das
	// Issue sagt es scharf: „**Loopback allein beweist keine lokale
	// Inferenz**." Ein Dienst auf `127.0.0.1` kann weiterleiten, und dann
	// rechnet jemand anderes.
	Ort KIOrt

	// Zielkarte ist die Karte, für die vorgeschlagen wird.
	Zielkarte int64

	// Zielrevision ist ihr Stand beim Anlegen.
	//
	// **Der eingefrorene Arbeitsstand** aus dem Ziel des Issues: hat sich die
	// Karte inzwischen geändert, gehört der Vorschlag zu einer Karte, die es
	// so nicht mehr gibt.
	Zielrevision string

	// Schemakennung bindet ihn an die Felder, gegen die geprüft wurde.
	Schemakennung string

	// Kostenposten ist die **Reservierung** aus dem Budget (O12, #1439).
	//
	// **Vor dem Ruf und nicht danach**: wer erst nach der Antwort bucht, hat
	// bei zwei gleichzeitigen Aufträgen zweimal die Grenze gelesen und
	// keinmal gehalten.
	Kostenposten string

	Angelegt time.Time

	// Ablauf ist, wann der Vorschlag nicht mehr übernommen werden darf.
	Ablauf time.Time
}

// Validate prüft, ob der Auftrag an alles gebunden ist, woran er muss.
func (a KIAuftrag) Validate() []*FieldError {
	var befunde []*FieldError

	for feld, wert := range map[string]string{
		"kennung": a.Kennung, "zielrevision": a.Zielrevision,
		"schemakennung": a.Schemakennung, "kostenposten": a.Kostenposten,
		"anbieter": a.Anbieter, "alias": a.Alias,
	} {
		if wert == "" {
			befunde = append(befunde, NewFieldError(feld, SchluesselPflichtfeld,
				"Ein KI-Auftrag nennt %s.", feld))
		}
	}

	for feld, wert := range map[string]int64{
		"profilID": a.ProfilID, "zielkarte": a.Zielkarte,
	} {
		if wert <= 0 {
			befunde = append(befunde, NewFieldError(feld, SchluesselPflichtfeld,
				"Ein KI-Auftrag nennt %s.", feld))
		}
	}

	if !IstKIFaehigkeit(a.Faehigkeit) {
		befunde = append(befunde, NewFieldError("faehigkeit", SchluesselValidierung,
			"%q ist keine bekannte KI-Fähigkeit.", string(a.Faehigkeit)))
	}

	if !IstKIOrt(a.Ort) {
		befunde = append(befunde, NewFieldError("ort", SchluesselValidierung,
			"%q ist kein bekannter Ort — es gibt »lokal« und »cloud«.",
			string(a.Ort)))
	}

	return append(befunde, a.pruefeFrist()...)
}

// pruefeFrist prüft den Ablauf gegen die Anlage.
func (a KIAuftrag) pruefeFrist() []*FieldError {
	if a.Ablauf.IsZero() {
		return []*FieldError{NewFieldError("ablauf", SchluesselPflichtfeld,
			"Ein KI-Auftrag hat eine Frist — ohne sie bliebe ein Vorschlag "+
				"für immer übernehmbar.")}
	}

	if !a.Angelegt.IsZero() && !a.Ablauf.After(a.Angelegt) {
		return []*FieldError{NewFieldError("ablauf", SchluesselValidierung,
			"Die Frist liegt vor der Anlage — dieser Auftrag wäre nie gültig.")}
	}

	return nil
}

// Abgelaufen sagt, ob die Frist vorbei ist.
//
// **Genau am Ablauf noch nicht**: eine Frist, die im selben Augenblick endet,
// in dem sie gilt, ist an einer Uhr mit Nanosekunden ein Zufall.
func (a KIAuftrag) Abgelaufen(jetzt time.Time) bool {
	return jetzt.After(a.Ablauf)
}

// KIVorschlag ist, was ein Modell vorschlägt — und **noch keine Änderung**.
type KIVorschlag struct {
	// Auftrag ist die Kennung des Auftrags, aus dem er stammt.
	Auftrag string

	// Felder sind die vorgeschlagenen Werte — leer heißt: nichts.
	//
	// **Leer ist erlaubt.** Ein Modell darf sich verweigern, und das Issue
	// nennt „Refusal" neben den Fehlern: eine Antwort ohne Vorschlag ist eine
	// Antwort.
	Felder map[string][]string

	// Modell ist, was **geantwortet** hat — Werkstück A11 (#1450).
	//
	// **Am Vorschlag und nicht am Auftrag**: der Auftrag nennt den Alias, den
	// jemand gewählt hat; hier steht, was hinter ihm lag, als geantwortet
	// wurde. Zwischen beidem darf ein Anbieter tauschen.
	Modell string

	// Aufgabe ist, was ein Mensch getippt hat — **nicht** der Rohprompt
	// ([KIFeldherkunft]).
	Aufgabe string
}
