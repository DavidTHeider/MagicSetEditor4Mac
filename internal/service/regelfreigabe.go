package service

// **Ob eine Partie so anfangen darf** — Werkstück G4 von Phase 5.5
// (#1461).
//
// # Was hier entschieden wird und wann
//
// **Vor „Bereit" und nicht beim ersten Zug.** Das Issue sagt es: „Konflikte
// werden vor Bereit abgewiesen." Wer eine Partie aufsetzt, erfährt beim
// Aufsetzen, dass sie nicht geht — nicht in der dritten Runde an einem
// Befehl, der scheitert.
//
// # Warum eine Lücke kein Fehler ist
//
// **Weil sie eine Auskunft ist.** Ein Regelbereich, den dieses Programm nicht
// durchsetzt, verhindert nur [domain.StrengeStreng]; im Sandkasten und unter
// Begleitung wird gespielt, und die Lücken stehen dabei. Das ist der
// Unterschied zwischen „geht nicht" und „geht, aber ihr macht es selbst".
//
// **Ein Konflikt dagegen ist einer.** Eine Hausregel, die das Profil nicht
// erlaubt, ist kein Hinweis, sondern eine Absage — sonst wäre die Leiter aus
// [domain.Regelrang] eine Empfehlung.
//
// # Und warum hier nichts über Karten steht
//
// **Weil das die Formatprüfung tut** (`formatpruefung.go`). Ein Deck gegen
// ein Format zu halten ist eine andere Frage als die, ob ein Regelwerk eine
// Strenge trägt — und sie hat ihre eigene Naht, ihre eigenen Urteile und
// ihre eigene Quelle bei Scryfall.

import (
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Freigabewunsch ist, womit jemand anfangen will.
type Freigabewunsch struct {
	Profil domain.Regelprofil

	Strenge domain.Strenge

	// Sitze ist, wie viele mitspielen.
	Sitze int

	Hausregeln []domain.Hausregel

	// Modus ist, wie viel Hilfe das Programm geben soll — PLAY-02.AK-01
	// (#1710).
	//
	// **Leer heißt [domain.ModusAnfaenger]**, und das ist die vorsichtige
	// Vorgabe: wer nichts sagt, bekommt so viel Unterstützung, wie dieses
	// Programm tragen kann.
	Modus domain.Bedienmodus

	// Wahl ist, was die Runde **selbst** übernehmen will (PLAY-02.AK-02).
	//
	// **Eine Abwahl und keine Zuwahl** — der Grund steht bei
	// [domain.Regelwahl].
	Wahl domain.Regelwahl
}

// Freigabe ist die Antwort — und sie sagt mehr als ja oder nein.
type Freigabe struct {
	// Erlaubt sagt, ob so angefangen werden darf.
	Erlaubt bool

	// Hoechste ist die strengste Stufe, die dieses Profil trägt.
	//
	// **Sie steht auch dann da, wenn die gewünschte durchgeht** — wer
	// „begleitet" verlangt und „begleitet" bekommt, soll trotzdem sehen,
	// dass „streng" nicht zur Wahl steht.
	Hoechste domain.Strenge

	// Luecken sind die Bereiche, die [domain.StrengeStreng] nicht tragen.
	Luecken []domain.Regelbereich

	// Hindernisse sind die Gründe, warum es nicht geht — leer, wenn es geht.
	Hindernisse []string

	// Baum ordnet die Bereiche nach Ablauf und Mechanik (PLAY-02.AK-02).
	Baum []domain.Regelgruppe

	// Widersprueche sind die Abwahlen, an denen noch etwas hängt.
	//
	// **Sie machen die Freigabe nicht unmöglich, sondern erklären sie**:
	// AK-02 verlangt „unvereinbare Kombinationen werden erklärt", und eine
	// Erklärung, die abweist, statt zu erklären, ist keine.
	Widersprueche []domain.Regelwiderspruch

	// Matrix ist die **wirksame** — nach der Abwahl.
	//
	// **Und nicht die des Profils**: wer die des Profils zeigte, zeigte
	// eine Zusage, die für diese Runde nicht gilt (§10).
	Matrix domain.Regelmatrix
}

// Regelfreigabe entscheidet über den Anfang einer Partie.
//
// **Ohne Zustand und ohne Naht**: sie rechnet über das, was ihr gereicht
// wird. Damit ist sie an jeder Fläche dieselbe — und §16 verlangt genau das.
type Regelfreigabe struct{}

// NeueRegelfreigabe baut sie.
func NeueRegelfreigabe() *Regelfreigabe { return &Regelfreigabe{} }

// Pruefe gibt die Freigabe zu einem Wunsch.
//
// **Ein Fehler ist etwas anderes als ein Nein**: ein Profil, das sich selbst
// widerspricht, ist ein Fehler; ein Profil, das eine Strenge nicht trägt, ist
// eine Antwort.
func (r *Regelfreigabe) Pruefe(wunsch Freigabewunsch) (Freigabe, error) {
	if fehler := wunsch.Profil.Pruefe(); fehler != nil {
		return Freigabe{}, fehler
	}

	if !domain.IstStrenge(wunsch.Strenge) {
		return Freigabe{}, fmt.Errorf(
			"%w: %q ist keine Strenge", domain.ErrValidation, wunsch.Strenge)
	}

	if fehler := wunsch.Wahl.Pruefe(); fehler != nil {
		return Freigabe{}, fehler
	}

	modus := wunsch.Modus
	if modus == "" {
		modus = domain.ModusAnfaenger
	}

	// **Die wirksame Matrix ist die, über die gerechnet wird** — sonst
	// stünde eine Strenge zur Wahl, die nach der Abwahl niemand mehr trägt.
	wirksam := wunsch.Wahl.Wirksam(wunsch.Profil.Matrix)

	freigabe := Freigabe{
		Hoechste:      hoechsteStrenge(wirksam),
		Luecken:       wirksam.Luecken(),
		Hindernisse:   []string{},
		Baum:          domain.Regelbaum(),
		Widersprueche: wunsch.Wahl.Widersprueche(),
		Matrix:        wirksam,
	}

	if fehler := wunsch.Wahl.PasstZumModus(modus); fehler != nil {
		freigabe.Hindernisse = append(freigabe.Hindernisse, fehler.Error())
	}

	if !wirksam.TraegtStrenge(wunsch.Strenge) {
		freigabe.Hindernisse = append(freigabe.Hindernisse, fmt.Sprintf(
			"das Regelwerk %q trägt die Strenge %q nicht — %d von %d "+
				"Bereichen setzt dieses Programm nicht durch",
			wunsch.Profil.Kennung, wunsch.Strenge,
			len(freigabe.Luecken), len(domain.Regelbereiche())))
	}

	if !wunsch.Profil.TraegtSpieler(wunsch.Sitze) {
		freigabe.Hindernisse = append(freigabe.Hindernisse, fmt.Sprintf(
			"das Regelwerk %q ist für %d bis %d Spieler, gewünscht sind %d",
			wunsch.Profil.Kennung, wunsch.Profil.Spieler.Mindestens,
			wunsch.Profil.Spieler.Hoechstens, wunsch.Sitze))
	}

	for _, regel := range wunsch.Hausregeln {
		if fehler := wunsch.Profil.ErlaubtHausregel(regel); fehler != nil {
			freigabe.Hindernisse = append(freigabe.Hindernisse, fehler.Error())
		}
	}

	freigabe.Erlaubt = len(freigabe.Hindernisse) == 0

	return freigabe, nil
}

// hoechsteStrenge gibt die strengste Stufe, die eine Matrix trägt.
//
// **Von schwach nach stark gegangen** — die letzte, die trägt, ist die
// höchste.
//
// **Und nicht andersherum mit einem Rückfall am Ende.** Der erste Anlauf
// suchte von stark nach schwach und schrieb hinter die Schleife ein
// `return StrengeSandkasten` — eine Zeile, die **niemand betritt**, denn der
// Sandkasten trägt immer. Eine unerreichbare Anweisung ist keine Vorsorge,
// sondern eine Behauptung darüber, was im Ernstfall geschieht (§3).
//
// **Die Reihenfolge von [domain.Strengegrade] trägt das** — sie geht von
// schwach nach stark, und ein Lauf in `internal/domain` hält sie fest.
func hoechsteStrenge(matrix domain.Regelmatrix) domain.Strenge {
	hoechste := domain.StrengeSandkasten

	for _, grad := range domain.Strengegrade() {
		if matrix.TraegtStrenge(grad) {
			hoechste = grad
		}
	}

	return hoechste
}
