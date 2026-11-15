package app

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Paketfuehrer führt die installierten Pakete — die Naht zu `internal/service`.
type Paketfuehrer interface {
	Auffrischen(ctx context.Context) ([]domain.Befund, error)
	Vermerken(ctx context.Context, vermerke ...service.Vermerk) ([]domain.Befund, error)
	Uebernehmen(ctx context.Context, ort string) (*service.Uebernahme, error)
	Abhaengigkeiten(ctx context.Context) ([]service.Abhaengigkeitsbefund, error)
	Liste(ctx context.Context) ([]domain.Paket, error)
	Uebersicht(ctx context.Context) ([]service.Paketzeile, error)

	// Herkunftsart nennt, woher ein Paket stammt, ohne die Übersicht zu lesen.
	Herkunftsart(paket domain.Paket) domain.Paketherkunftsart

	// Plan, Ersetzen und Zurück — DOM-01.AK-02/AK-03 (#1696).
	PlanEntfernen(ctx context.Context, art, name string) (service.Paketplan, error)
	PlanEinbinden(ctx context.Context, quelle string) (service.Paketplan, error)
	PlanZurueck(ctx context.Context, art, name string) (service.Paketplan, error)
	Ersetzen(ctx context.Context, quelle string, bestaetigt bool) (domain.Paket, error)
	Zurueck(ctx context.Context, art, name string) (domain.Paket, error)
	Installieren(ctx context.Context, quelle string) (domain.Paket, error)
	Entfernen(ctx context.Context, art, name, bestaetigterPfad string) error

	// Anlegen schreibt ein neues Paket aus einem Gerüst und trägt es ein
	// (#1350).
	Anlegen(ctx context.Context, wunsch service.Neuanlage) (domain.Paket, error)

	// Kopieren legt eine eigene Kopie an — FMT-02.AK-04 (#1691).
	Kopieren(ctx context.Context, art, name, neu string) (domain.Paket, error)

	// Weitergeben und Einlesen gehen über `.mse-installer` — FMT-02.AK-04
	// (#1691).
	Weitergeben(ctx context.Context, namen []string, ziel string) (string, error)
	Einlesen(ctx context.Context, quelle string) (*service.Uebernahme, error)

	// Setbedarf nennt, was einem Set an Paketen fehlt — DOM-01.AK-04.
	Setbedarf(ctx context.Context, setID int64) ([]service.Abhaengigkeitsbefund, error)

	// Kopfbedarf ist dasselbe für einen Kopf aus einer Datei — FMT-03.
	Kopfbedarf(ctx context.Context, kopf domain.SetHead) ([]service.Abhaengigkeitsbefund, error)
}

// PaketDTO ist ein installiertes Paket, wie das Fenster es sieht.
type PaketDTO struct {
	Art  string `json:"art"`
	Name string `json:"name"`

	// Anzeigename ist, was über dem Paket steht — der Kurzname,
	// großgeschrieben wie MSE2; ohne ihn der Name (UX-05, S07.NAT-02, #1677).
	Anzeigename string `json:"anzeigename"`

	Version  string `json:"version"`
	Pfad     string `json:"pfad"`
	Herkunft string `json:"herkunft"`

	// Pruefsumme ist, was beim Bezug herauskam — leer bei einem Paket, das
	// niemand bezogen hat.
	Pruefsumme string `json:"pruefsumme"`

	// Braucht sind die Abhängigkeiten als fertige Zeilen — „paket version".
	//
	// **Zeichenketten und keine Verbünde**: die Seite zeigt sie, sie rechnet
	// nicht mit ihnen. Wer #57 baut und mehr braucht, schneidet den Typ dann —
	// mit einem Fall, der es verlangt.
	Braucht []string `json:"braucht"`

	// Gruppe ist `installer group:` — der Pfad, unter dem die Verwaltung dieses
	// Paket zeigt, mit `/` als Trenner (#244).
	Gruppe string `json:"gruppe"`

	// Rang ist `position hint:` — die Ordnung innerhalb einer Gruppe.
	Rang int `json:"rang"`

	// Herkunftsart, Geaendert, GebrauchtVon und Benutzt sagen, woher es
	// stammt, ob es vom Stand beim Einbinden abweicht, wer es braucht und wie
	// viele Sets es benutzen — DOM-01.AK-01 (#1696), `service.Uebersicht`.
	Herkunftsart string   `json:"herkunftsart"`
	Geaendert    bool     `json:"geaendert"`
	GebrauchtVon []string `json:"gebrauchtVon"`
	Benutzt      int      `json:"benutzt"`

	// Artrang ist MSEs `PackageType` — die Reihenfolge der sieben Arten.
	//
	// **Sie kommt von hier und nicht aus dem Fenster.** Ableitbar wäre sie
	// dort aus `art`; dann stünde die portierte Reihenfolge
	// (`package_update_list.hpp`, Zeilen 57–67, Fassung f9d9356 vom
	// 2021-07-06) in TypeScript, wo kein Go-Test sie hält.
	Artrang int `json:"artrang"`

	// Hauptdatei ist die Datei, die das Paket ausmacht — `style`, `game` …
	// (FMT-02, #1691). **Aus derselben Stelle wie beim Lesen**
	// (`msepaket.Art.Hauptdatei`); die Werkstatt öffnet sie in einer Kopie.
	Hauptdatei string `json:"hauptdatei"`
}

// PaketeAntwort trägt den Katalog samt Befunden.
type PaketeAntwort struct {
	Pakete  []PaketDTO `json:"pakete"`
	Befunde []string   `json:"befunde"`
	Fehler  *Fehler    `json:"fehler"`
}

// PaketAntwort trägt ein einzelnes Paket — oder einen Abbruch.
type PaketAntwort struct {
	Paket *PaketDTO `json:"paket"`

	// Abgebrochen sagt, dass niemand etwas gewählt hat.
	//
	// Ein eigener Wert und kein leerer Pfad — derselbe Grund wie beim Öffnen
	// eines Sets: eine Oberfläche, die „nichts gewählt" von einem Fehler nicht
	// unterscheiden kann, zeigt beim Abbrechen eine Fehlermeldung.
	Abgebrochen bool    `json:"abgebrochen"`
	Fehler      *Fehler `json:"fehler"`
}

// Pakete frischt den Katalog auf und gibt ihn.
//
// **Auffrischen und Lesen in einem Zug.** Wer ein Paket bezieht oder von Hand
// ins Verzeichnis legt, sähe es sonst erst beim nächsten Start — und ein
// Katalog, der die Platte nicht kennt, ist eine Liste von Behauptungen.
func (a *App) Pakete(ctx context.Context) PaketeAntwort {
	ctx = a.imProfil(ctx)

	antwort := PaketeAntwort{Pakete: []PaketDTO{}, Befunde: []string{}}

	befunde, err := a.pakete.Auffrischen(ctx)
	for _, befund := range befunde {
		antwort.Befunde = append(antwort.Befunde, befund.String())
	}

	if err != nil {
		antwort.Fehler = Uebersetze(ctx, err)

		return antwort
	}

	zeilen, err := a.pakete.Uebersicht(ctx)
	if err != nil {
		antwort.Fehler = Uebersetze(ctx, err)

		return antwort
	}

	for _, zeile := range zeilen {
		gelesen := paketDTO(zeile.Paket)
		gelesen.Herkunftsart = string(zeile.Herkunftsart)
		gelesen.GebrauchtVon = append([]string{}, zeile.GebrauchtVon...)
		gelesen.Benutzt = zeile.Benutzt
		antwort.Pakete = append(antwort.Pakete, gelesen)
	}

	return antwort
}

// PaketInstallieren legt ein Paket von der Platte ins Paketverzeichnis.
func (a *App) PaketInstallieren(ctx context.Context, quelle string) PaketAntwort {
	ctx = a.imProfil(ctx)

	if strings.TrimSpace(quelle) == "" {
		return PaketAntwort{Abgebrochen: true}
	}

	paket, err := a.pakete.Installieren(ctx, quelle)
	if err != nil {
		return PaketAntwort{Fehler: Uebersetze(ctx, err)}
	}

	gelesen := a.mitUebersicht(ctx, paket)
	a.arbeitVermerken(ctx, domain.ArbeitsartPaket, gelesen.Name, domain.VorgangEingefuehrt)

	return PaketAntwort{Paket: &gelesen}
}

// PaketEntfernen nimmt ein Paket aus dem Paketverzeichnis.
func (a *App) PaketEntfernen(ctx context.Context, art, name string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return SchreibAntwort{Fehler: Uebersetze(ctx, a.pakete.Entfernen(ctx, art, name, ""))}
}

// MitPaketen setzt die Naht zum Paketverwalter.
func (a *App) MitPaketen(fuehrer Paketfuehrer) { a.pakete = fuehrer }

// paketDTO übersetzt ein Paket der Domäne.
func paketDTO(paket domain.Paket) PaketDTO {
	return PaketDTO{
		Art:         paket.Art,
		Name:        paket.Name,
		Anzeigename: service.Anzeigename(paket),
		Version:     paket.Version,
		Pfad:        paket.Pfad,
		Herkunft:    paket.Herkunft,
		Pruefsumme:  paket.Pruefsumme,
		Geaendert:   paket.Geaendert,
		Braucht:     BrauchtZeilen(paket),
		Gruppe:      paket.Gruppe,
		Rang:        paket.Rang,
		Artrang:     paket.Artrang(),
		Hauptdatei:  msepaket.Art(paket.Art).Hauptdatei(),

		// **Eine Liste und kein `null`** — auch dort, wo niemand die
		// Übersicht nachliest (C01, #2226).
		GebrauchtVon: []string{},
	}
}

// mitUebersicht gibt ein eben geschriebenes Paket samt seiner Zeile der
// Übersicht zurück — Herkunftsart, wer es braucht, wie viele Sets es
// benutzen —, so wie `Pakete` es nennt. **Bis C01 (#2226)** gaben die
// Schreibwege es ohne zurück: die Herkunftsart leer, die Liste `null`.
func (a *App) mitUebersicht(ctx context.Context, paket domain.Paket) PaketDTO {
	return a.uebersichtFuer(ctx)(paket)
}

// uebersichtFuer liest die Übersicht **einmal** und gibt zurück, was ein
// Paket damit zum DTO macht — eine Übernahme bringt viele, und die Übersicht
// zählt die Nutzung über alle Sets.
//
// **Scheitert das Nachlesen, bleibt das Geschriebene geschrieben** (§7,
// Regel 3): das DTO ist dann das Paket ohne Bedarf und Nutzung — **aber mit
// seiner Herkunftsart**, denn das Fenster kennt nur vier Werte, und ein
// leerer ist keiner davon (Zweitblick über Phase B, Befund 2).
func (a *App) uebersichtFuer(ctx context.Context) func(domain.Paket) PaketDTO {
	nachName := map[string]service.Paketzeile{}

	if zeilen, err := a.pakete.Uebersicht(ctx); err == nil {
		for _, zeile := range zeilen {
			nachName[zeile.Paket.Dateiname()] = zeile
		}
	}

	return func(paket domain.Paket) PaketDTO {
		gelesen := paketDTO(paket)
		gelesen.Herkunftsart = string(a.pakete.Herkunftsart(paket))

		if zeile, da := nachName[paket.Dateiname()]; da {
			gelesen.Herkunftsart = string(zeile.Herkunftsart)
			gelesen.GebrauchtVon = append([]string{}, zeile.GebrauchtVon...)
			gelesen.Benutzt = zeile.Benutzt
		}

		return gelesen
	}
}

// AbhaengigkeitDTO ist ein Vorbehalt gegen eine Abhängigkeit.
//
// **Angaben, kein Satz** (§9): `art` ist ein Schlüssel, den die Oberfläche in
// ihrer Sprache ausformuliert — `fehlt`, `zualt`, `unlesbar`, `zyklus`.
type AbhaengigkeitDTO struct {
	Paket    string `json:"paket"`
	Braucht  string `json:"braucht"`
	Art      string `json:"art"`
	Verlangt string `json:"verlangt"`
	Da       string `json:"da"`
}

// AbhaengigkeitenAntwort trägt die Vorbehalte gegen die Abhängigkeiten.
type AbhaengigkeitenAntwort struct {
	Befunde []AbhaengigkeitDTO `json:"befunde"`
	Fehler  *Fehler            `json:"fehler"`
}

// Abhaengigkeiten prüft die Abhängigkeiten der installierten Pakete (#57).
func (a *App) Abhaengigkeiten(ctx context.Context) AbhaengigkeitenAntwort {
	ctx = a.imProfil(ctx)

	antwort := AbhaengigkeitenAntwort{Befunde: []AbhaengigkeitDTO{}}

	befunde, err := a.pakete.Abhaengigkeiten(ctx)
	if err != nil {
		antwort.Fehler = Uebersetze(ctx, err)

		return antwort
	}

	antwort.Befunde = alsAbhaengigkeiten(befunde)

	return antwort
}

// BrauchtZeilen gibt die Abhängigkeiten als fertige Zeilen — „paket version".
// **Eine Stelle für alle vier Flächen** (§16): Fenster, Kommandozeile, MCP
// und REST zeigen sie gleich.
func BrauchtZeilen(paket domain.Paket) []string {
	braucht := make([]string, 0, len(paket.Braucht))

	for _, bedarf := range paket.Braucht {
		zeile := bedarf.Paket
		if bedarf.Version != "" {
			zeile += " " + bedarf.Version
		}

		braucht = append(braucht, zeile)
	}

	return braucht
}
