package peer

// **Der signierte Handschlag** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Warum er drei Schritte hat und nicht zwei
//
// **Weil jede Seite belegen muss, dass sie *jetzt* da ist.** Eine
// unterschriebene Nachricht belegt „ich besitze diesen Schlüssel" — nicht mehr.
// Wer sie mitschneidet, hat denselben Beleg. Erst eine Unterschrift über eine
// Zahl, die die **andere** Seite gerade gewürfelt hat, sagt: ich habe sie
// gesehen, also spreche ich gerade mit dir.
//
// Das kann nur über Kreuz gehen: A würfelt, B unterschreibt A's Zahl mitsamt
// der eigenen, A unterschreibt B's Zahl. Zwei Schritte reichten für **eine**
// Richtung.
//
// # Und warum er keine eigene Unterschrift trägt
//
// **Weil die Hülle schon eine hat.** Sie unterschreibt Kopf **und** Rumpf
// ([Huelle]); eine zweite Unterschrift über denselben Inhalt wäre eine zweite
// Gestalt, die auseinanderlaufen kann. Was der Handschlag beiträgt, sind die
// Zahlen, nicht ein zweiter Umschlag.

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Handschlagfassung ist die Fassung dieses Ablaufs.
//
// **Eine Abweichung wird abgewiesen und nicht heruntergehandelt** — das
// verlangt die Abnahme des Issues wörtlich. Wer zwei Fassungen kann, bietet
// sie als zwei Protokolle an; eine Einigung im Ablauf wäre die Stelle, an der
// ein Dritter die schwächere wählt.
const Handschlagfassung = 1

// Die drei Schritte, als Nachrichtenarten.
const (
	ArtAngebot   = "handschlag/angebot"
	ArtAnnahme   = "handschlag/annahme"
	ArtAbschluss = "handschlag/abschluss"
)

// dieZwecke sagt, in welchem Protokoll ein Zweck gilt.
//
// **Die Zuordnung ist die Trennung.** Beide Angaben wären für sich gültig —
// `/mse4mac/game/1` ist ein Protokoll, `teilen` ein Zweck. Zusammen wären sie
// eine Partie, die in Wahrheit Karten überträgt.
var dieZwecke = map[domain.Zweck]string{
	domain.ZweckPaarung: ProtokollPaarung,
	domain.ZweckTeilen:  ProtokollTeilen,
	domain.ZweckHandel:  ProtokollTeilen,
	domain.ZweckPartie:  ProtokollSpiel,
}

// ProtokollZu sagt, in welchem Protokoll ein Zweck gilt.
//
// **Die Zuordnung nach draußen**, damit `service` eine Verbindung im
// **richtigen** Protokoll aufbaut. Sie bleibt hier, weil sie Protokollwissen
// ist; wer sie zweimal führte, hätte beim fünften Zweck zwei Listen.
func ProtokollZu(zweck domain.Zweck) (string, bool) {
	protokoll, gilt := dieZwecke[zweck]

	return protokoll, gilt
}

// Angebot ist der erste Schritt.
type Angebot struct {
	Fassung int          `json:"fassung"`
	Zweck   domain.Zweck `json:"zweck"`
	Rolle   string       `json:"rolle"`
	Kanal   string       `json:"kanal"`
	Nonce   string       `json:"nonce"`

	// Name ist **kein Nachweis** — siehe [Sitzung.SelbstgewaehlterName].
	Name string `json:"name"`
}

// Annahme ist der zweite Schritt.
//
// **`DeinNonce` ist die Zahl des Anbieters**, zurückgegeben; `MeinNonce` die
// eigene. Die Sicht ist die des Absenders, und das gilt in allen drei
// Schritten — sonst hieße dasselbe Feld zweimal etwas anderes.
type Annahme struct {
	Fassung   int          `json:"fassung"`
	Zweck     domain.Zweck `json:"zweck"`
	Rolle     string       `json:"rolle"`
	Kanal     string       `json:"kanal"`
	DeinNonce string       `json:"deinNonce"`
	MeinNonce string       `json:"meinNonce"`
	Name      string       `json:"name"`
}

// Abschluss ist der dritte Schritt.
type Abschluss struct {
	Fassung   int    `json:"fassung"`
	DeinNonce string `json:"deinNonce"`
	MeinNonce string `json:"meinNonce"`
}

// Wunsch ist, was die anbietende Seite öffnen will.
type Wunsch struct {
	Zweck domain.Zweck
	Rolle string
	Kanal string
	Name  string

	// Erwartet ist die Kennung, die am anderen Ende stehen **muss**.
	//
	// **Leer heißt: jeder.** Das ist der Fall einer offenen Einladung; wer
	// ein bekanntes Gegenüber anspricht, trägt es ein, und eine Antwort von
	// jemand anderem endet hier.
	Erwartet Kennung
}

// Bereitschaft ist, was die annehmende Seite beiträgt.
//
// **Ohne Zweck**: den bestimmt das Angebot. Ob er erlaubt ist, entscheidet
// N1.4 — hier wäre eine Rechteliste in einem Blatt.
type Bereitschaft struct {
	Rolle    string
	Kanal    string
	Name     string
	Erwartet Kennung
}

// Anbahnung ist der Stand der anbietenden Seite zwischen Schritt eins und drei.
type Anbahnung struct {
	ich    Identitaet
	wunsch Wunsch
	nonce  string
}

// Entgegennahme ist der Stand der annehmenden Seite.
type Entgegennahme struct {
	ich           Identitaet
	gegenueber    Kennung
	seinName      string
	zweck         domain.Zweck
	protokoll     string
	kanal         string
	meineRolle    string
	seineRolle    string
	nonceAnbieter string
	nonceAnnehmer string
}

// Anbieten eröffnet einen Handschlag.
func (i Identitaet) Anbieten(
	wunsch Wunsch, zufall io.Reader,
) (Anbahnung, []byte, error) {
	protokoll, gilt := dieZwecke[wunsch.Zweck]
	if !gilt {
		return Anbahnung{}, nil, fmt.Errorf(
			"%w: %q ist kein Zweck dieses Programms", domain.ErrValidation,
			wunsch.Zweck)
	}

	if wunsch.Rolle == "" || wunsch.Kanal == "" {
		return Anbahnung{}, nil, fmt.Errorf(
			"%w: ein Angebot braucht Rolle und Kanal", domain.ErrValidation)
	}

	nonce, fehler := eineZufallszahl(zufall)
	if fehler != nil {
		return Anbahnung{}, nil, fehler
	}

	roh, fehler := i.verpackeSchritt(protokoll, ArtAngebot, nonce, 0, Angebot{
		Fassung: Handschlagfassung,
		Zweck:   wunsch.Zweck,
		Rolle:   wunsch.Rolle,
		Kanal:   wunsch.Kanal,
		Nonce:   nonce,
		Name:    wunsch.Name,
	})
	if fehler != nil {
		return Anbahnung{}, nil, fehler
	}

	return Anbahnung{ich: i, wunsch: wunsch, nonce: nonce}, roh, nil
}

// Entgegennehmen beantwortet ein Angebot.
//
// # Die Reihenfolge der Prüfungen ist Absicht
//
// **Erst die Hülle, dann die Form, dann das Gedächtnis.** Eine Zufallszahl
// wird erst vermerkt, wenn alles andere stimmt — sonst füllte ein Absender mit
// Unsinn die Liste, und ein späteres, echtes Angebot fände seine Zahl
// „schon einmal da".
func (i Identitaet) Entgegennehmen(
	roh []byte, bereit Bereitschaft, zufall io.Reader, gedaechtnis *Nonceliste,
) (Entgegennahme, []byte, error) {
	huelle, angebot, fehler := lesAngebot(roh)
	if fehler != nil {
		return Entgegennahme{}, nil, fehler
	}

	gegenueber := Kennung(huelle.Absender)

	if befund := pruefeGegenueber(
		i.kennung, gegenueber, bereit.Erwartet); befund != nil {
		return Entgegennahme{}, nil, befund
	}

	protokoll, fehler := pruefeAngebot(huelle, angebot, bereit)
	if fehler != nil {
		return Entgegennahme{}, nil, fehler
	}

	if befund := gedaechtnis.Merke(angebot.Nonce); befund != nil {
		return Entgegennahme{}, nil, befund
	}

	meiner, fehler := eineZufallszahl(zufall)
	if fehler != nil {
		return Entgegennahme{}, nil, fehler
	}

	antwort, fehler := i.verpackeSchritt(
		protokoll, ArtAnnahme, angebot.Nonce, 1, Annahme{
			Fassung:   Handschlagfassung,
			Zweck:     angebot.Zweck,
			Rolle:     bereit.Rolle,
			Kanal:     bereit.Kanal,
			DeinNonce: angebot.Nonce,
			MeinNonce: meiner,
			Name:      bereit.Name,
		})
	if fehler != nil {
		return Entgegennahme{}, nil, fehler
	}

	return Entgegennahme{
		ich:           i,
		gegenueber:    gegenueber,
		seinName:      angebot.Name,
		zweck:         angebot.Zweck,
		protokoll:     protokoll,
		kanal:         bereit.Kanal,
		meineRolle:    bereit.Rolle,
		seineRolle:    angebot.Rolle,
		nonceAnbieter: angebot.Nonce,
		nonceAnnehmer: meiner,
	}, antwort, nil
}

// verpackeSchritt schreibt einen Schritt in seine Hülle.
//
// # Warum das Schreiben hier keinen Fehlerweg hat
//
// **Weil die drei Rümpfe aus Zeichenketten und einer Zahl bestehen** —
// `json.Marshal` scheitert daran nicht. Und **fiele es wider Erwarten doch**,
// bliebe `roh` leer, und [Identitaet.Verpacke] wiese die Hülle eine Zeile
// später ab: „ohne Rumpf ist eine Hülle keine". Der Fehler wird also nicht
// verschluckt, sondern von der nächsten Prüfung gefangen — und dafür steht
// hier kein Zweig, den kein Lauf betreten kann (§10).
func (i Identitaet) verpackeSchritt(
	protokoll, art, sitzung string, folge int, rumpf any,
) ([]byte, error) {
	roh, _ := json.Marshal(rumpf) //nolint:errcheck // siehe oben

	return i.Verpacke(Huelle{
		Protokoll: protokoll,
		Art:       art,
		ID:        sitzung + "." + strconv.Itoa(folge),
		Sitzung:   sitzung,
		Absender:  string(i.kennung),
		Folge:     folge,
		Rumpf:     roh,
	})
}

// Gegenueber ist die Kennung, die das Angebot unterschrieben hat.
//
// **Belegt und nicht behauptet**: [Identitaet.Entgegennehmen] hat die
// Unterschrift gegen genau diese Kennung gehalten.
func (e Entgegennahme) Gegenueber() Kennung { return e.gegenueber }

// Zweck ist der Zweck, den das Angebot nennt.
//
// **Er kommt von der anderen Seite** — und ob er erlaubt ist, entscheidet
// nicht dieses Paket, sondern `service` (N1.4).
func (e Entgegennahme) Zweck() domain.Zweck { return e.zweck }
