// **Kontakte aus einer vCard am Fenster** — ID-04
// (#1703),
// Vermerk [0137](../../docs/entscheidungen/0137-kontakte-kommen-als-vcard.md).
//
// Die Fachlichkeit steht in `service.Adressbuch`; hier wird übersetzt (§2).
// **Der Dateidialog ist die Berechtigung**: gelesen wird genau die Datei, die
// ein Mensch gewählt hat, und keine Freigabe von macOS für das ganze
// Adressbuch.

package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Adressbuchgrenze ist die größte vCard-Datei, die gelesen wird — ein
// Adressbuch mit Bildern ist ein paar Megabyte, keine Hunderte.
const Adressbuchgrenze = 32 << 20

// Adressbuchquelle ist, was die Kontakte am Fenster brauchen.
type Adressbuchquelle interface {
	Vorschau(ctx context.Context, daten []byte) (service.Einlesebericht, error)
	Uebernehmen(ctx context.Context, auswahl []domain.Kontakt) (angelegt, ersetzt int, fehler error)
	Liste(ctx context.Context) ([]domain.Kontakt, error)
	Loeschen(ctx context.Context, id, fassung int64) error
}

// KontaktwegDTO ist eine Adresse oder Nummer.
type KontaktwegDTO struct {
	Art  string `json:"art"`
	Wert string `json:"wert"`
}

// KontaktDTO ist ein Kontakt, wie das Fenster ihn sieht. **Das Bild reist
// als Base64** — so schreibt `encoding/json` ein `[]byte`.
type KontaktDTO struct {
	ID          int64           `json:"id"`
	Fassung     int64           `json:"fassung"`
	Name        string          `json:"name"`
	Wege        []KontaktwegDTO `json:"wege"`
	Bild        []byte          `json:"bild"`
	Bildart     string          `json:"bildart"`
	Ursprung    string          `json:"ursprung"`
	Uebernommen string          `json:"uebernommen"`
}

// KartenvorschauDTO ist eine Karte aus der Datei mit ihrem Abgleich.
type KartenvorschauDTO struct {
	Kontakt   KontaktDTO  `json:"kontakt"`
	Abgleich  string      `json:"abgleich"`
	Vorhanden *KontaktDTO `json:"vorhanden"`
}

// KontaktvorschauAntwort ist die Vorschau einer Datei.
type KontaktvorschauAntwort struct {
	Karten      []KartenvorschauDTO `json:"karten"`
	Befunde     []string            `json:"befunde"`
	Abgebrochen bool                `json:"abgebrochen"`
	Fehler      *Fehler             `json:"fehler"`
}

// KontaktuebernahmeAntwort sagt, was geschrieben wurde.
type KontaktuebernahmeAntwort struct {
	Angelegt int     `json:"angelegt"`
	Ersetzt  int     `json:"ersetzt"`
	Fehler   *Fehler `json:"fehler"`
}

// KontakteAntwort trägt die Kontakte des Profils.
type KontakteAntwort struct {
	Kontakte []KontaktDTO `json:"kontakte"`
	Fehler   *Fehler      `json:"fehler"`
}

// MitAdressbuch setzt die Naht.
func (a *App) MitAdressbuch(quelle Adressbuchquelle) { a.adressbuch = quelle }

// KontakteEinlesen fragt nach einer vCard-Datei und zeigt, was darin steht —
// **ohne zu schreiben**.
func (a *App) KontakteEinlesen(ctx context.Context) KontaktvorschauAntwort {
	ctx = a.imProfil(ctx)
	texte := DialogtexteFuer(a.sprache)

	pfad, fehler := a.waehleDatei(ctx, Dateiauswahl{
		Titel: texte.KontakteOeffnen, Filter: texte.KontakteFilter, Muster: "*.vcf",
	})
	if fehler != nil {
		return KontaktvorschauAntwort{Karten: []KartenvorschauDTO{}, Befunde: []string{}, Fehler: Uebersetze(ctx, fehler)}
	}

	if pfad == "" {
		return KontaktvorschauAntwort{Karten: []KartenvorschauDTO{}, Befunde: []string{}, Abgebrochen: true}
	}

	return a.KontakteAusDatei(ctx, pfad)
}

// KontakteAusDatei zeigt, was in einer vCard-Datei steht — der Weg der
// Kommandozeile, die keinen Dialog hat.
func (a *App) KontakteAusDatei(ctx context.Context, pfad string) KontaktvorschauAntwort {
	ctx = a.imProfil(ctx)
	leer := KontaktvorschauAntwort{Karten: []KartenvorschauDTO{}, Befunde: []string{}}

	if a.adressbuch == nil {
		leer.Fehler = Uebersetze(ctx, keinAdressbuch())

		return leer
	}

	daten, fehler := vcardLesen(pfad)
	if fehler != nil {
		leer.Fehler = Uebersetze(ctx, fehler)

		return leer
	}

	bericht, fehler := a.adressbuch.Vorschau(ctx, daten)
	if fehler != nil {
		leer.Fehler = Uebersetze(ctx, fehler)

		return leer
	}

	antwort := KontaktvorschauAntwort{Karten: []KartenvorschauDTO{}, Befunde: bericht.Befunde}

	for _, karte := range bericht.Karten {
		vorschau := KartenvorschauDTO{Kontakt: zuKontaktDTO(karte.Kontakt), Abgleich: string(karte.Abgleich)}

		if karte.Vorhanden != nil {
			vorhanden := zuKontaktDTO(*karte.Vorhanden)
			vorschau.Vorhanden = &vorhanden
		}

		antwort.Karten = append(antwort.Karten, vorschau)
	}

	return antwort
}

// vcardLesen liest eine Datei bis zur Grenze.
func vcardLesen(pfad string) ([]byte, error) {
	datei, fehler := os.Open(pfad) // #nosec G304 -- der Pfad kommt aus dem Dialog oder vom Menschen an der Kommandozeile
	if errors.Is(fehler, fs.ErrNotExist) {
		return nil, fmt.Errorf("%w: die Datei %s gibt es nicht", domain.ErrNotFound, filepath.Base(pfad))
	}

	if fehler != nil {
		return nil, fmt.Errorf("die Datei %s öffnen: %w", filepath.Base(pfad), fehler)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // gelesen wird nur; ein Fehler beim Schließen ändert nichts

	daten, fehler := io.ReadAll(io.LimitReader(datei, Adressbuchgrenze+1))
	if fehler != nil {
		return nil, fmt.Errorf("die Datei %s lesen: %w", filepath.Base(pfad), fehler)
	}

	if len(daten) > Adressbuchgrenze {
		return nil, domain.NewFieldError("datei", domain.SchluesselValidierung,
			"Die Datei ist größer als %d MiB — wähle weniger Kontakte aus.", Adressbuchgrenze>>20)
	}

	return daten, nil
}

// KontakteUebernehmen schreibt, was gewählt wurde.
func (a *App) KontakteUebernehmen(ctx context.Context, auswahl []KontaktDTO) KontaktuebernahmeAntwort {
	ctx = a.imProfil(ctx)

	if a.adressbuch == nil {
		return KontaktuebernahmeAntwort{Fehler: Uebersetze(ctx, keinAdressbuch())}
	}

	kontakte := make([]domain.Kontakt, 0, len(auswahl))
	for _, dto := range auswahl {
		kontakte = append(kontakte, ausKontaktDTO(dto))
	}

	angelegt, ersetzt, fehler := a.adressbuch.Uebernehmen(ctx, kontakte)
	if fehler != nil {
		return KontaktuebernahmeAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return KontaktuebernahmeAntwort{Angelegt: angelegt, Ersetzt: ersetzt}
}

// Kontakte gibt die Kontakte des Profils.
func (a *App) Kontakte(ctx context.Context) KontakteAntwort {
	ctx = a.imProfil(ctx)

	if a.adressbuch == nil {
		return KontakteAntwort{Kontakte: []KontaktDTO{}, Fehler: Uebersetze(ctx, keinAdressbuch())}
	}

	kontakte, fehler := a.adressbuch.Liste(ctx)
	if fehler != nil {
		return KontakteAntwort{Kontakte: []KontaktDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	antwort := KontakteAntwort{Kontakte: make([]KontaktDTO, 0, len(kontakte))}
	for _, kontakt := range kontakte {
		antwort.Kontakte = append(antwort.Kontakte, zuKontaktDTO(kontakt))
	}

	return antwort
}

// KontaktLoeschen nimmt einen Kontakt heraus.
func (a *App) KontaktLoeschen(ctx context.Context, id, fassung int64) SchreibAntwort {
	ctx = a.imProfil(ctx)

	if a.adressbuch == nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, keinAdressbuch())}
	}

	return SchreibAntwort{Fehler: Uebersetze(ctx, a.adressbuch.Loeschen(ctx, id, fassung))}
}

func keinAdressbuch() error {
	return fmt.Errorf("%w: das Adressbuch ist nicht eingerichtet", domain.ErrValidation)
}

func zuKontaktDTO(kontakt domain.Kontakt) KontaktDTO {
	dto := KontaktDTO{
		ID: kontakt.ID, Fassung: kontakt.Fassung, Name: kontakt.Name, Wege: []KontaktwegDTO{},
		Bild: kontakt.Bild, Bildart: kontakt.Bildart, Ursprung: kontakt.Ursprung,
	}

	if !kontakt.Uebernommen.IsZero() {
		dto.Uebernommen = kontakt.Uebernommen.Format(time.RFC3339)
	}

	for _, weg := range kontakt.Wege {
		dto.Wege = append(dto.Wege, KontaktwegDTO{Art: string(weg.Art), Wert: weg.Wert})
	}

	return dto
}

func ausKontaktDTO(dto KontaktDTO) domain.Kontakt {
	kontakt := domain.Kontakt{
		ID: dto.ID, Fassung: dto.Fassung, Name: dto.Name, Bild: dto.Bild, Bildart: dto.Bildart,
		Ursprung: dto.Ursprung,
	}

	for _, weg := range dto.Wege {
		kontakt.Wege = append(kontakt.Wege, domain.Kontaktweg{Art: domain.Kontaktwegart(weg.Art), Wert: weg.Wert})
	}

	return kontakt
}
