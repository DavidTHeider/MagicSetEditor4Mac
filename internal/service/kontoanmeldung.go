// **Die vier Kontoarten** — Werkstück 6 von Phase 4.12
// (#1351),
// nach der Kommentierung vom 2026-09-05.
//
// # Was hier entschieden wird
//
// **Wer sich unter einer eigenen Kennung anmelden darf** — mit einer Adresse,
// mit einer Rufnummer, anonym, oder gar nicht. Die vier stehen in
// [domain.PruefeEmail] und daneben; **was aus ihnen wird**, steht hier.
//
// # Was hier **nicht** steht
//
// **Kein Kennwortverfahren.** Was ein gehütetes Kennwort ist, entscheidet
// `internal/kennwort`; dieser Dienst ruft es und vergleicht nicht selbst.
//
// **Und keine Zustellung.** Eine Adresse ist hier eine **Kennung** und kein
// nachgewiesener Kanal: dass es das Postfach gibt, sagt nur eine Zustellung,
// und dafür bräuchte die Anwendung fremde Zugangsdaten, die man nicht
// mitliefern kann — dieselbe Lage wie bei den Anbietern in Werkstück 4. Die
// Oberfläche sagt es am Feld, statt es zu verschweigen (Werkstück 13).
//
// **Deshalb steht hier auch kein Feld „bestätigt"**: ein Zustand, den nichts
// verlassen kann, ist keine Auskunft, sondern ein Vorwurf an den Nutzer. Er
// entsteht mit der Zustellung, und die hat ihr eigenes Issue.

package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/kennwort"
)

// Kennungsablage ist die Naht zum Bestand.
type Kennungsablage interface {
	NutzerZuKennung(ctx context.Context, kennung string) (domain.Nutzer, error)
	NutzerAnlegen(ctx context.Context, neu domain.Nutzer) (domain.Nutzer, error)
}

// Kontoanmeldung trägt Konten ein und lässt sie herein.
type Kontoanmeldung struct {
	ablage Kennungsablage

	// faktor prüft den zweiten Faktor — Werkstück 7 (#1351). `nil` heißt: es
	// wird nicht gefragt.
	faktor Anmeldepruefer
}

// Anmeldepruefer ist die Naht zum zweiten Faktor **beim Anmelden**.
//
// **Eine eigene neben [Faktorpruefer]**, weil sie eine andere Frage stellt:
// jene darf der Nutzer für den Profilwechsel abschalten, diese nicht. Hier
// steht der Faktor für das ein, wofür es ihn gibt (Werkstück 8, #1351).
type Anmeldepruefer interface {
	Pruefe(ctx context.Context, nutzerID int64, nachweis string) error
}

// NeueKontoanmeldung baut sie über einer Ablage.
func NeueKontoanmeldung(ablage Kennungsablage) *Kontoanmeldung {
	return &Kontoanmeldung{ablage: ablage}
}

// MitFaktor hängt den Prüfer des zweiten Faktors an.
//
// # Warum hier und nicht bei der Anmeldung über einen Anbieter
//
// **Weil ein Kennwort allein genau das ist, was ein zweiter Faktor ergänzen
// soll.** Wer sich über GitHub anmeldet, hat dort schon einen zweiten Faktor,
// wenn er einen will — einen zweiten obendrauf zu verlangen wäre eine Hürde
// auf einen Nachweis, den dieses Programm nicht ausgestellt hat.
func (k *Kontoanmeldung) MitFaktor(pruefer Anmeldepruefer) *Kontoanmeldung {
	k.faktor = pruefer

	return k
}

// Kontowunsch ist, was jemand angibt, der ein Konto anlegen will.
//
// **Eine Kennung und nicht zwei Felder**: ob da eine Adresse oder eine Nummer
// steht, sieht man ihr an, und zwei Felder hießen für den Nutzer, sich vorher
// für eines zu entscheiden.
type Kontowunsch struct {
	// Name ist der Anzeigename — leer heißt: einen aus der Kennung.
	Name string

	// Kennung ist die Adresse oder die Rufnummer.
	Kennung string

	// Kennwort ist der Klartext. Er verlässt diesen Dienst nicht.
	Kennwort string
}

// Eintragen legt ein Konto unter einer Kennung an.
func (k *Kontoanmeldung) Eintragen(ctx context.Context, wunsch Kontowunsch) (
	domain.Nutzer, error,
) {
	neu, fehler := ausKennung(wunsch.Kennung)
	if fehler != nil {
		return domain.Nutzer{}, fehler
	}

	// **Erst prüfen, dann hüten.** Argon2id kostet gewollt Zeit; sie an eine
	// Eingabe zu hängen, die ohnehin abgewiesen wird, ist eine Tür, durch die
	// man das Programm mit leeren Anfragen beschäftigt.
	if wunsch.Kennwort == "" {
		return domain.Nutzer{}, domain.NewFieldError("kennwort",
			domain.SchluesselPflichtfeld, "Ein Konto braucht ein Kennwort.")
	}

	if fehler := k.kennungFrei(ctx, wunsch.Kennung); fehler != nil {
		return domain.Nutzer{}, fehler
	}

	gehuetet, huetefehler := kennwort.Huete(wunsch.Kennwort)
	if huetefehler != nil {
		return domain.Nutzer{}, domain.NewFieldError("kennwort",
			domain.SchluesselValidierung, "%s", meldungOhneMarke(huetefehler))
	}

	neu.Name = wunsch.Name
	if strings.TrimSpace(neu.Name) == "" {
		neu.Name = namenAusKennung(wunsch.Kennung)
	}

	neu.Kennwort = gehuetet

	if fehler := neu.Validate(); fehler != nil {
		return domain.Nutzer{}, fehler
	}

	return k.ablage.NutzerAnlegen(ctx, neu)
}

// Anmelden gibt den Nutzer zu dieser Kennung, wenn das Kennwort stimmt.
//
// # Warum die Meldung nicht sagt, was falsch war
//
// **Weil die Auskunft nur dem hilft, der rät.** „Diese Adresse kennen wir
// nicht" sagt einem Fremden, welche Adressen es gibt; „das Kennwort stimmt
// nicht" sagt ihm, dass er die richtige Adresse hat.
//
// **Was sie hier trotzdem nicht verbirgt**: die Nutzerliste steht in der
// Profilwahl. Der gleichlautende Satz ist deshalb keine Zusage über
// Unkenntlichkeit, sondern eine Gewohnheit, die nichts kostet — und die
// trägt, sobald ein Bestand einmal woanders liegt.
// # Der zweite Faktor kommt **nach** dem Kennwort
//
// **Weil er sonst verriete, dass es das Konto gibt.** „Nenne deinen Code" auf
// eine Kennung, die niemandem gehört, wäre die Auskunft, die [abgewiesen]
// gerade vermeidet.
//
// **Ein leerer Nachweis heißt „über den Sensor"**; wer keinen Faktor hat,
// wird nicht gefragt.
func (k *Kontoanmeldung) Anmelden(ctx context.Context, kennung, klartext, nachweis string) (
	domain.Nutzer, error,
) {
	einer, fehler := k.ablage.NutzerZuKennung(ctx, kennung)
	if fehler != nil {
		if errors.Is(fehler, domain.ErrNotFound) {
			return domain.Nutzer{}, abgewiesen()
		}

		return domain.Nutzer{}, fehler
	}

	// **Ein Nutzer ohne Kennwort meldet sich nicht mit einem an.** Anonyme und
	// der erste Nutzer haben ein leeres Feld; ohne diese Zeile hinge alles an
	// [kennwort.Stimmt] — und was das mit einer leeren Zeichenkette tut, wäre
	// eine Zusage aus einem anderen Paket.
	if einer.Kennwort == "" {
		return domain.Nutzer{}, abgewiesen()
	}

	if stimmtnicht := kennwort.Stimmt(einer.Kennwort, klartext); stimmtnicht != nil {
		if errors.Is(stimmtnicht, kennwort.ErrStimmtNicht) {
			return domain.Nutzer{}, abgewiesen()
		}

		return domain.Nutzer{}, stimmtnicht
	}

	if k.faktor != nil {
		if pruefehler := k.faktor.Pruefe(ctx, einer.ID, nachweis); pruefehler != nil {
			return domain.Nutzer{}, pruefehler
		}
	}

	return einer, nil
}

// Zufallskennungslaenge ist, wie viele Zeichen der Name eines anonymen Kontos
// hinter seinem Bindestrich trägt.
//
// **Sechzehn Zeichen aus Base32 sind achtzig Bit.** Die Zahl ist keine
// Sicherheitszusage — der Name ist keine Kennung, unter der man hereinkommt —,
// sondern die Antwort auf „zweimal derselbe Name": bei achtzig Bit trifft ein
// Bestand das nie.
const Zufallskennungslaenge = 16

// Anonym legt ein Konto ohne jede Angabe an.
//
// **Ein Vermerk und kein leeres Feld** ([domain.Nutzer.Anonym]): ein Nutzer
// ohne Adresse könnte auch der erste sein, und die Oberfläche zeigte ihn sonst
// wie einen, der seine Angaben vergessen hat.
func (k *Kontoanmeldung) Anonym(ctx context.Context) (domain.Nutzer, error) {
	return k.ablage.NutzerAnlegen(ctx, domain.Nutzer{
		Name: "anonym-" + zufallskennung(), Anonym: true,
	})
}

// kennungFrei weist ab, wenn es diese Kennung schon gibt.
//
// **Vor dem Anlegen und nicht statt der Eindeutigkeit im Bestand.** Der
// Teilindex aus Wanderung 0037 ist die Zusage; diese Abfrage ist die Meldung,
// die das Feld nennt. Wer nur eine von beiden hätte, hätte entweder einen
// rohen SQLite-Fehler an der Oberfläche (§8) oder ein Rennen zwischen zwei
// Anlagen.
func (k *Kontoanmeldung) kennungFrei(ctx context.Context, kennung string) error {
	_, fehler := k.ablage.NutzerZuKennung(ctx, kennung)

	switch {
	case fehler == nil:
		// **Doppelt, nicht überholt** (ENG-05.AK-03): unter `konflikt`
		// stand im Fenster „zwischenzeitlich geändert, lade neu".
		return domain.NewFieldError("kennung", domain.SchluesselDoppelterName,
			"Unter dieser Kennung gibt es schon ein Konto.")

	case !errors.Is(fehler, domain.ErrNotFound):
		return fehler
	}

	return nil
}

// ausKennung sagt, ob da eine Adresse oder eine Nummer steht.
//
// **Am Klammeraffen und nicht an einer Reihenfolge von Versuchen**: wer erst
// die Adresse prüft und bei einem Fehler die Nummer, meldet am Ende den
// Fehler der zweiten Prüfung für eine vertippte Adresse — „eine Rufnummer
// trägt nur Ziffern" für `wer@exampl`.
func ausKennung(kennung string) (domain.Nutzer, error) {
	if strings.Contains(kennung, "@") {
		if fehler := domain.PruefeEmail(kennung); fehler != nil {
			return domain.Nutzer{}, domain.NewFieldError("kennung",
				domain.SchluesselValidierung, "%s", meldungOhneMarke(fehler))
		}

		return domain.Nutzer{Email: kennung}, nil
	}

	if fehler := domain.PruefeTelefon(kennung); fehler != nil {
		return domain.Nutzer{}, domain.NewFieldError("kennung",
			domain.SchluesselValidierung, "%s", meldungOhneMarke(fehler))
	}

	return domain.Nutzer{Telefon: kennung}, nil
}

// namenAusKennung gibt den Namen, unter dem ein Konto ohne Namen entsteht.
func namenAusKennung(kennung string) string {
	if vorne, _, hatAt := strings.Cut(kennung, "@"); hatAt {
		return vorne
	}

	return kennung
}

// abgewiesen ist die **eine** Antwort auf jede gescheiterte Anmeldung.
func abgewiesen() error {
	return domain.NewFieldError("kennung", domain.SchluesselValidierung,
		"Kennung oder Kennwort passen nicht zusammen.")
}

// meldungOhneMarke gibt den Satz hinter dem umschließenden Fehlertyp.
//
// **Damit die Meldung nicht zweimal dasselbe sagt**: der Feldfehler trägt
// seine Marke selbst, und ohne diese Zeile stünde „ein Wert ist ungültig: ein
// Wert ist ungültig: …" an der Oberfläche.
func meldungOhneMarke(fehler error) string {
	satz := fehler.Error()

	if _, hinten, hatDoppelpunkt := strings.Cut(satz, ": "); hatDoppelpunkt {
		satz = hinten
	}

	return grossAmAnfang(satz) + "."
}

// zufallskennung gibt eine Zeichenkette, die es so noch nicht gab.
//
// **Ohne Fehlerrückgabe**, weil `crypto/rand.Read` seit Go 1.24 keinen Fehler
// mehr gibt — dieselbe Stelle wie in `internal/anmeldung`.
func zufallskennung() string {
	roh := make([]byte, Zufallskennungslaenge)
	rand.Read(roh) //nolint:errcheck // gibt seit Go 1.24 keinen Fehler

	return strings.ToLower(
		base32.StdEncoding.WithPadding(base32.NoPadding).
			EncodeToString(roh)[:Zufallskennungslaenge])
}
