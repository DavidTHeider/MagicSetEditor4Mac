// **Der Vorlagenbereich** — Werkstück 2 von Phase 3.3
// (#1164).
//
// Drei Bindungen: aufzählen, lesen, schreiben. Die Fachlichkeit steht in
// `service.Paketwerkstatt`; hier wird übersetzt (§2).

package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Paketdateiquelle ist, was der Vorlagenbereich braucht.
type Paketdateiquelle interface {
	DateienUndBeiwerk(ctx context.Context, paket string) (dateien, beiwerk []string, fehler error)
	Lesen(ctx context.Context, paket, datei string) (inhalt, stand string, fehler error)
	Schreiben(ctx context.Context, paket, datei, inhalt, erwartet string) (string, error)

	// Texte gibt jede bearbeitbare Datei samt Inhalt — für den Symbolindex
	// der Werkstatt (#1418).
	Texte(ctx context.Context, paket string) ([]service.Paketdateitext, error)

	// Der Dateibaum (FMT-02, #1691) — siehe `service/paketdateibaum.go`.
	Anlegen(ctx context.Context, paket, datei, inhalt string) (string, error)
	Einfuegen(ctx context.Context, paket, quelle, datei string) error
	Umbenennen(ctx context.Context, paket, alt, neu, erwartet string) error
	Entfernen(ctx context.Context, paket, datei, erwartet string) error
}

// Paketdateiantwort trägt die bearbeitbaren Dateien eines Pakets und sein
// Beiwerk — Bilder und Schriften, die sich nicht als Text bearbeiten lassen.
type Paketdateiantwort struct {
	Dateien []string `json:"dateien"`
	Beiwerk []string `json:"beiwerk"`
	Fehler  *Fehler  `json:"fehler"`
}

// Paketdateiinhalt trägt eine Vorlage samt ihrem Stand.
type Paketdateiinhalt struct {
	Inhalt string `json:"inhalt"`

	// Stand ist die SHA-256 des Inhalts.
	//
	// **Er geht beim Speichern zurück.** Stimmt er dann nicht mehr, hat
	// jemand anderes geschrieben — und das Fenster erfährt es, statt die
	// fremde Änderung zu überschreiben (§7, Regel 2).
	Stand string `json:"stand"`

	Fehler *Fehler `json:"fehler"`
}

// Paketdateistand trägt den Stand nach dem Schreiben.
type Paketdateistand struct {
	Stand  string  `json:"stand"`
	Fehler *Fehler `json:"fehler"`
}

// MitPaketdateien setzt die Naht.
func (a *App) MitPaketdateien(quelle Paketdateiquelle) { a.paketdateien = quelle }

// Paketdateien nennt die bearbeitbaren Dateien eines Pakets.
func (a *App) Paketdateien(ctx context.Context, paket string) Paketdateiantwort {
	ctx = a.imProfil(ctx)

	leer := Paketdateiantwort{Dateien: []string{}, Beiwerk: []string{}}

	if a.paketdateien == nil {
		leer.Fehler = Uebersetze(ctx, keinePaketwerkstatt())

		return leer
	}

	dateien, beiwerk, err := a.paketdateien.DateienUndBeiwerk(ctx, paket)
	if err != nil {
		leer.Fehler = Uebersetze(ctx, err)

		return leer
	}

	if dateien != nil {
		leer.Dateien = dateien
	}

	if beiwerk != nil {
		leer.Beiwerk = beiwerk
	}

	return leer
}

// PaketdateiLesen gibt eine Paketdatei samt ihrem Stand.
func (a *App) PaketdateiLesen(ctx context.Context, paket, datei string) Paketdateiinhalt {
	ctx = a.imProfil(ctx)

	if a.paketdateien == nil {
		return Paketdateiinhalt{Fehler: Uebersetze(ctx, keinePaketwerkstatt())}
	}

	inhalt, stand, err := a.paketdateien.Lesen(ctx, paket, datei)
	if err != nil {
		return Paketdateiinhalt{Fehler: Uebersetze(ctx, err)}
	}

	return Paketdateiinhalt{Inhalt: inhalt, Stand: stand}
}

// PaketdateiSchreiben setzt eine Paketdatei zurück, wenn der Stand noch stimmt.
func (a *App) PaketdateiSchreiben(
	ctx context.Context, paket, datei, inhalt, stand string,
) Paketdateistand {
	ctx = a.imProfil(ctx)

	if a.paketdateien == nil {
		return Paketdateistand{Fehler: Uebersetze(ctx, keinePaketwerkstatt())}
	}

	neuer, err := a.paketdateien.Schreiben(ctx, paket, datei, inhalt, stand)
	if err != nil {
		return Paketdateistand{Fehler: Uebersetze(ctx, err)}
	}

	return Paketdateistand{Stand: neuer}
}

// keinePaketwerkstatt meldet die fehlende Naht.
func keinePaketwerkstatt() error {
	return fmt.Errorf("%w: der Vorlagenbereich ist nicht eingerichtet", domain.ErrValidation)
}
