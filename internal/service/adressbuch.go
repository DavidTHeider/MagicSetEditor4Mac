// **Kontakte aus einer vCard übernehmen** — ID-04
// (#1703),
// Vermerk [0137](../../docs/entscheidungen/0137-kontakte-kommen-als-vcard.md).
//
// # Zwei Schritte, und der erste schreibt nichts
//
// **Die Vorschau** liest eine Datei und sagt je Karte, wie sie zu den
// vorhandenen Kontakten steht; **die Übernahme** schreibt nur, was ein
// Mensch daraus gewählt hat — in einem Geschäft. Eine Synchronisierung gibt
// es nicht: wer eine Karte später noch einmal übernimmt, sieht vorher, was
// sich ändert (ID-04.AK-03).

package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vcard"
)

// Adressbuchablage ist die Naht zum Bestand.
type Adressbuchablage interface {
	Kontakte(ctx context.Context) ([]domain.Kontakt, error)
	KontakteUebernehmen(ctx context.Context, kontakte []*domain.Kontakt) error
	KontaktLoeschen(ctx context.Context, id, fassung int64) error
}

// Adressbuch ist der Anwendungsfall „Kontakte aus einem Adressbuch führen".
type Adressbuch struct {
	ablage Adressbuchablage
}

// NeuesAdressbuch baut es über einer Ablage.
func NeuesAdressbuch(ablage Adressbuchablage) *Adressbuch { return &Adressbuch{ablage: ablage} }

// Kartenvorschau ist eine Karte aus der Datei, wie sie übernommen würde.
type Kartenvorschau struct {
	Kontakt   domain.Kontakt
	Abgleich  domain.Kontaktabgleich
	Vorhanden *domain.Kontakt
}

// Einlesebericht ist die Vorschau einer Datei.
type Einlesebericht struct {
	Karten []Kartenvorschau

	// Befunde nennen, was nicht gelesen wurde — mit Zeile oder Name.
	Befunde []string
}

// Vorschau liest eine vCard-Datei und gleicht jede Karte ab — **ohne zu
// schreiben**.
func (a *Adressbuch) Vorschau(ctx context.Context, daten []byte) (Einlesebericht, error) {
	vorhandene, fehler := a.ablage.Kontakte(ctx)
	if fehler != nil {
		return Einlesebericht{}, fehler
	}

	karten, gelesen := vcard.Lies(daten)
	bericht := Einlesebericht{Karten: []Kartenvorschau{}, Befunde: []string{}}

	for _, befund := range gelesen {
		bericht.Befunde = append(bericht.Befunde, befund.Satz)
	}

	// gesehen sind die Karten dieser Datei — **eine Datei verdoppelt sich
	// nicht selbst** (Zweitblick über ID-04, Befund 1).
	var gesehen []domain.Kontakt

	for _, karte := range karten {
		kontakt, befunde := ausKarte(karte)
		bericht.Befunde = append(bericht.Befunde, befunde...)

		if kontakt == nil {
			continue
		}

		// **Verworfen wird nur, was sicher dieselbe Karte ist** — dieselbe
		// UID oder dieselben Angaben. Eine Vermutung bleibt stehen (Nachlauf
		// über ID-04, Befund 2): vielleicht sind es zwei Menschen.
		if wiederholt, _ := domain.KontaktAbgleichen(gesehen, *kontakt); sicherDieselbe(wiederholt) {
			bericht.Befunde = append(bericht.Befunde, fmt.Sprintf(
				"Die Karte ab Zeile %d wiederholt %s aus derselben Datei — sie bleibt weg.", karte.Zeile, kontakt.Name))

			continue
		}

		gesehen = append(gesehen, *kontakt)

		abgleich, getroffen := domain.KontaktAbgleichen(vorhandene, *kontakt)
		bericht.Karten = append(bericht.Karten, Kartenvorschau{Kontakt: *kontakt, Abgleich: abgleich, Vorhanden: getroffen})
	}

	return bericht, nil
}

// ausKarte macht aus einer Karte einen Kontakt — **nil**, wenn er nicht
// taugt. Ein Bild, das nicht taugt, bleibt weg, der Kontakt nicht.
func ausKarte(karte vcard.Karte) (gebaut *domain.Kontakt, befunde []string) {
	kontakt := domain.Kontakt{
		Name: karte.Name, Ursprung: karte.Kennung, Bild: karte.Bild, Bildart: karte.Bildart,
		Wege: []domain.Kontaktweg{},
	}

	for _, weg := range karte.Emails {
		kontakt.Wege = append(kontakt.Wege, domain.Kontaktweg{Art: domain.KontaktwegEmail, Wert: weg.Wert})
	}

	for _, weg := range karte.Telefone {
		kontakt.Wege = append(kontakt.Wege, domain.Kontaktweg{Art: domain.KontaktwegTelefon, Wert: weg.Wert})
	}

	kontakt.Normalisieren()

	for _, befund := range kontakt.Validate() {
		if befund.Field != "bild" {
			return nil, []string{fmt.Sprintf("Die Karte ab Zeile %d: %s", karte.Zeile, befund.Message)}
		}

		kontakt.Bild, kontakt.Bildart = nil, ""
		befunde = append(befunde, fmt.Sprintf("Das Bild von %s bleibt weg: %s", kontakt.Name, befund.Message))
	}

	return &kontakt, befunde
}

// Uebernehmen schreibt, was gewählt wurde: ohne Kennung als neuen Kontakt,
// mit Kennung über den vorhandenen, gegen dessen Fassung.
//
// **Zuerst wird alles geprüft, dann alles geschrieben** — und wer eine
// Karte als neu übernehmen will, deren Ursprung es schon gibt, wird
// abgewiesen, statt eine Dublette anzulegen (ID-04.AK-02).
func (a *Adressbuch) Uebernehmen(ctx context.Context, auswahl []domain.Kontakt) (angelegt, ersetzt int, fehler error) {
	vorhandene, fehler := a.ablage.Kontakte(ctx)
	if fehler != nil {
		return 0, 0, fehler
	}

	zuSchreiben := make([]*domain.Kontakt, 0, len(auswahl))

	// neue sind die, die diese Übernahme schon anlegt.
	var neue []domain.Kontakt

	for stelle := range auswahl {
		kontakt := auswahl[stelle]
		kontakt.Normalisieren()

		if befunde := kontakt.Validate(); len(befunde) > 0 {
			return 0, 0, befunde[0]
		}

		if kontakt.ID == 0 {
			// **Gegen den Bestand und gegen das, was diese Übernahme schon
			// anlegt** — und eine Vermutung zählt mit: wer sie übernimmt,
			// ersetzt (Zweitblick über ID-04, Befund 1).
			if abgleich, getroffen := domain.KontaktAbgleichen(vorhandene, kontakt); abgleich != domain.KontaktNeu {
				return 0, 0, kontaktSchonDa(kontakt.Name, getroffen.ID)
			}

			// **In der Auswahl selbst zählt nur, was sicher dieselbe ist**
			// — zwei vermutlich gleiche Karten hat der Mensch beide gesehen.
			if abgleich, _ := domain.KontaktAbgleichen(neue, kontakt); sicherDieselbe(abgleich) {
				return 0, 0, kontaktSchonDa(kontakt.Name, 0)
			}

			neue = append(neue, kontakt)
			angelegt++
		} else {
			ersetzt++
		}

		zuSchreiben = append(zuSchreiben, &kontakt)
	}

	if fehler := a.ablage.KontakteUebernehmen(ctx, zuSchreiben); fehler != nil {
		return 0, 0, fehler
	}

	return angelegt, ersetzt, nil
}

// sicherDieselbe sagt, ob ein Abgleich dieselbe Karte meint — dieselbe UID
// oder dieselben Angaben — und nicht nur eine Vermutung.
func sicherDieselbe(abgleich domain.Kontaktabgleich) bool {
	return abgleich == domain.KontaktGleich || abgleich == domain.KontaktGeaendert
}

// kontaktSchonDa ist die Absage für eine Karte, die es schon gibt.
func kontaktSchonDa(name string, kontaktID int64) error {
	if kontaktID == 0 {
		return domain.NewFieldError("", domain.SchluesselDoppelt,
			"%s steht zweimal in der Auswahl — übernimm die Karte einmal.", name)
	}

	return domain.NewFieldError("", domain.SchluesselDoppelt,
		"%s steht schon als Kontakt %d im Adressbuch — ersetze ihn, statt ihn noch einmal anzulegen.",
		name, kontaktID)
}

// Liste gibt alle Kontakte des Profils.
func (a *Adressbuch) Liste(ctx context.Context) ([]domain.Kontakt, error) {
	return a.ablage.Kontakte(ctx)
}

// Loeschen nimmt einen Kontakt heraus — gegen seine Fassung.
func (a *Adressbuch) Loeschen(ctx context.Context, id, fassung int64) error {
	return a.ablage.KontaktLoeschen(ctx, id, fassung)
}
