// **Der Export personenbezogener Daten** — Werkstück 8 von Phase 4.12
// (#1351),
// Punkt 10 der Kommentierung vom 2026-09-05.
//
// # Was drin ist
//
// **Alles, was dieses Programm über einen Menschen weiß** — seine Angaben,
// wie er sich anmeldet, welche zweiten Faktoren er hat, was er eingestellt
// hat, und welche Arbeitsumgebungen ihm gehören.
//
// # Was nicht drin ist, und warum
//
// **Kein Kennwort und kein Geheimnis.** Beides ist kein *Datum über einen
// Menschen*, sondern ein Mittel, ihn zu sein. Ein Export, der sie mitgäbe,
// wäre eine Datei, mit der jemand anderes sich anmeldet — der Auskunft, die
// er geben soll, hilft das nichts. **Dass** es sie gibt, steht drin.
//
// # Und was daneben liegt
//
// **Die Inhalte.** Sets, Decks und Sammlung gehören einem **Profil**, nicht
// unmittelbar dem Menschen; sie gehen den Weg, den es für sie gibt — `mse
// sichern` je Set, `mse sammlung-ausgeben` für die Sammlung. Sie hier noch
// einmal zu verpacken hieße, zwei Ausfuhrwege für dieselbe Sache zu führen.

package service

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Personendaten ist, was über einen Menschen gespeichert ist.
type Personendaten struct {
	// Erhoben ist der Zeitpunkt des Exports.
	//
	// **Weil eine Auskunft ohne Datum keine ist**: sie sagt, was am
	// Stichtag galt, und nicht, was heute gilt.
	Erhoben time.Time `json:"erhoben"`

	Nutzer   Personenangaben  `json:"nutzer"`
	Profile  []Personenprofil `json:"profile"`
	Konten   []Personenkonto  `json:"anbieterkonten"`
	Faktoren []string         `json:"zweiteFaktoren"`

	// Einstellungen sind die, die dem Menschen gehören.
	Einstellungen map[string]string `json:"einstellungen"`
}

// Personenangaben sind die Felder des Nutzers — **ohne Kennwort**.
//
// **Exportiert, obwohl es nur die Auskunft baut** — Werkstück 8 (#1351):
// `internal/app` reicht diese Auskunft an die Flächen weiter, und Wails
// erzeugt daraus TypeScript. Ein unexportierter Typ in einer Antwort ließe
// sich weder benennen noch übersetzen.
type Personenangaben struct {
	Name         string `json:"name"`
	Anrede       string `json:"anrede"`
	Namenszusatz string `json:"namenszusatz"`
	Vorname      string `json:"vorname"`
	Weiterenamen string `json:"weitereVornamen"`
	Nachname     string `json:"nachname"`
	Email        string `json:"email"`
	Zweitemail   string `json:"zweitemail"`
	Telefon      string `json:"telefon"`
	Slogan       string `json:"slogan"`
	Kurzebio     string `json:"kurzeBiografie"`
	Langebio     string `json:"langeBiografie"`
	Bild         string `json:"bild"`
	Anonym       bool   `json:"anonym"`

	// HatKennwort sagt, **dass** eines gesetzt ist — nicht welches.
	HatKennwort bool `json:"hatKennwort"`
}

// Personenprofil ist eine Arbeitsumgebung im Export.
type Personenprofil struct {
	Name string `json:"name"`
}

// Personenkonto ist eine fremde Kennung im Export.
type Personenkonto struct {
	Anbieter    string `json:"anbieter"`
	Anzeigename string `json:"anzeigename"`
}

// Personenquelle ist die Naht zu allem, was über einen Menschen dasteht.
//
// **Eine Schnittstelle und nicht die vier einzeln**: der Export ist die eine
// Stelle, an der sie zusammenkommen, und vier Nähte hier wären vier Wege,
// dieselbe Ablage hereinzureichen (§2).
type Personenquelle interface {
	Nutzer(ctx context.Context) ([]domain.Nutzer, error)
	Profile(ctx context.Context, nutzerID int64) ([]domain.Profil, error)
	Konten(ctx context.Context, nutzerID int64) ([]domain.Anbieterkonto, error)
	Faktoren(ctx context.Context, nutzerID int64) ([]domain.Faktor, error)
	Nutzereinstellungen(ctx context.Context, nutzerID int64) (map[string]string, error)
}

// Personenauskunft gibt heraus, was über einen Menschen gespeichert ist.
type Personenauskunft struct {
	quelle Personenquelle
	uhr    func() time.Time
}

// NeuePersonenauskunft baut sie über einer Quelle.
func NeuePersonenauskunft(quelle Personenquelle) *Personenauskunft {
	return &Personenauskunft{quelle: quelle, uhr: time.Now}
}

// MitUhr setzt die Uhr — für Läufe, die einen festen Zeitpunkt brauchen.
func (p *Personenauskunft) MitUhr(uhr func() time.Time) *Personenauskunft {
	p.uhr = uhr

	return p
}

// Personendaten sammelt alles, was über diesen Menschen dasteht.
//
// **Ein Lesefehler reißt ab und wird nicht verschluckt.** Eine halbe Auskunft
// sieht aus wie eine ganze — und genau das darf eine Auskunft über
// personenbezogene Daten nicht (§8).
func (p *Personenauskunft) Personendaten(ctx context.Context, nutzerID int64) (
	Personendaten, error,
) {
	einer, fehler := p.nutzerSuchen(ctx, nutzerID)
	if fehler != nil {
		return Personendaten{}, fehler
	}

	profile, profilfehler := p.quelle.Profile(ctx, nutzerID)
	if profilfehler != nil {
		return Personendaten{}, profilfehler
	}

	konten, kontenfehler := p.quelle.Konten(ctx, nutzerID)
	if kontenfehler != nil {
		return Personendaten{}, kontenfehler
	}

	faktoren, faktorfehler := p.quelle.Faktoren(ctx, nutzerID)
	if faktorfehler != nil {
		return Personendaten{}, faktorfehler
	}

	einstellungen, einstellungsfehler := p.quelle.Nutzereinstellungen(ctx, nutzerID)
	if einstellungsfehler != nil {
		return Personendaten{}, einstellungsfehler
	}

	return Personendaten{
		Erhoben:       p.uhr(),
		Nutzer:        angabenAus(einer),
		Profile:       profileAus(profile),
		Konten:        kontenAus(konten),
		Faktoren:      faktorartenAus(faktoren),
		Einstellungen: einstellungen,
	}, nil
}

// nutzerSuchen holt den Nutzer über die Liste.
//
// **Über die Liste und nicht über eine eigene Abfrage**: sie liest ohnehin
// jeden, und eine zweite Naht dafür wäre eine zweite Stelle, an der die
// Spalten stimmen müssen.
func (p *Personenauskunft) nutzerSuchen(ctx context.Context, nutzerID int64) (
	domain.Nutzer, error,
) {
	alle, fehler := p.quelle.Nutzer(ctx)
	if fehler != nil {
		return domain.Nutzer{}, fehler
	}

	for _, einer := range alle {
		if einer.ID == nutzerID {
			return einer, nil
		}
	}

	return domain.Nutzer{}, domain.NewFieldError("nutzerID",
		domain.SchluesselNichtGefunden, "Diesen Nutzer führt der Bestand nicht.")
}

// angabenAus übersetzt die Entität — **ohne Kennwort**.
func angabenAus(einer domain.Nutzer) Personenangaben {
	return Personenangaben{
		Name:         einer.Name,
		Anrede:       einer.Anrede,
		Namenszusatz: einer.Namenszusatz,
		Vorname:      einer.Vorname,
		Weiterenamen: einer.Weiterenamen,
		Nachname:     einer.Nachname,
		Email:        einer.Email,
		Zweitemail:   einer.Zweitemail,
		Telefon:      einer.Telefon,
		Slogan:       einer.Slogan,
		Kurzebio:     einer.Kurzebio,
		Langebio:     einer.Langebio,
		Bild:         einer.Bild,
		Anonym:       einer.Anonym,
		HatKennwort:  einer.Kennwort != "",
	}
}

// profileAus übersetzt die Arbeitsumgebungen.
func profileAus(profile []domain.Profil) []Personenprofil {
	// **Nie `nil`, sondern eine leere Liste** (§3).
	gelesen := make([]Personenprofil, 0, len(profile))
	for _, eines := range profile {
		gelesen = append(gelesen, Personenprofil{Name: eines.Name})
	}

	return gelesen
}

// kontenAus übersetzt die fremden Kennungen.
func kontenAus(konten []domain.Anbieterkonto) []Personenkonto {
	gelesen := make([]Personenkonto, 0, len(konten))
	for _, eines := range konten {
		gelesen = append(gelesen, Personenkonto{
			Anbieter: eines.Anbieter, Anzeigename: eines.Anzeigename,
		})
	}

	return gelesen
}

// faktorartenAus nennt die Arten — **nicht die Geheimnisse**.
func faktorartenAus(faktoren []domain.Faktor) []string {
	gelesen := make([]string, 0, len(faktoren))
	for _, einer := range faktoren {
		gelesen = append(gelesen, string(einer.Art))
	}

	return gelesen
}
