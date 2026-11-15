// **Der zweite Faktor am Fenster und an den drei Flächen** — Werkstück 7 von
// Phase 4.12
// (#1351).
//
// # Was hier **nicht** geschieht
//
// **Kein Rechnen und kein Vergleich.** Diese Datei nimmt Kennungen entgegen
// und gibt Antworten zurück; **ob** ein Code stimmt, entscheidet
// `internal/service` (§2).
//
// **Und das Geheimnis geht genau einmal hinaus.** [Einrichtungsantwort] trägt
// es, weil jemand es abschreiben oder scannen muss — danach steht es in keiner
// Antwort dieser Anwendung mehr.

package app

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// FaktorDTO ist ein zweiter Faktor, wie das Fenster ihn sieht.
type FaktorDTO struct {
	Art string `json:"art"`

	// Bezeichnung sagt, welches Gerät gemeint ist — leer erlaubt.
	Bezeichnung string `json:"bezeichnung"`
}

// FaktorenAntwort trägt die Faktoren eines Nutzers samt der Frage, ob dieser
// Rechner Biometrie kann.
type FaktorenAntwort struct {
	Faktoren []FaktorDTO `json:"faktoren"`

	// Biometrie sagt, ob sich Touch ID oder Face ID hier einrichten ließe.
	Biometrie bool `json:"biometrie"`

	// Biometrieart ist „Touch ID", „Face ID" oder leer.
	Biometrieart string `json:"biometrieart"`

	// Biometriegrund steht da, wenn es nicht geht — leer, wenn es geht.
	Biometriegrund string `json:"biometriegrund"`

	Fehler *Fehler `json:"fehler"`
}

// Einrichtungsantwort trägt, was **einmal** zu sehen ist.
type Einrichtungsantwort struct {
	// Adresse ist die `otpauth://`-Adresse für den Scanner.
	Adresse string `json:"adresse"`

	// Geheimnis ist dieselbe Angabe zum Abtippen.
	Geheimnis string `json:"geheimnis"`

	// Codes sind die Wiederherstellungscodes im Klartext.
	Codes []string `json:"codes"`

	Fehler *Fehler `json:"fehler"`
}

// MitZweitemFaktor setzt die Naht.
func (a *App) MitZweitemFaktor(dienst *service.Zweiterfaktor) { a.faktor = dienst }

// Faktoren gibt die zweiten Faktoren eines Nutzers.
//
// **Samt der Auskunft über diesen Rechner**, und zwar in **einer** Antwort:
// die Seite zeigt beides nebeneinander, und zwei Rufe wären zwei Stände
// (§7, Regel 2).
func (a *App) Faktoren(ctx context.Context, nutzerID string) FaktorenAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return FaktorenAntwort{Faktoren: []FaktorDTO{}, Fehler: befund}
	}

	gelesen, fehler := a.faktor.Faktoren(ctx, kennung)
	if fehler != nil {
		return FaktorenAntwort{Faktoren: []FaktorDTO{}, Fehler: Uebersetze(ctx, fehler)}
	}

	// **Nie `nil`, sondern eine leere Liste** (§3).
	faktoren := make([]FaktorDTO, 0, len(gelesen))
	for _, einer := range gelesen {
		faktoren = append(faktoren, FaktorDTO{
			Art: string(einer.Art), Bezeichnung: einer.Bezeichnung,
		})
	}

	stand := a.biometriestand(ctx)
	stand.Faktoren = faktoren

	return stand
}

// FaktorEinrichten zieht ein TOTP-Geheimnis und gibt zurück, was einmal zu
// sehen ist.
//
// **Der Faktor gilt danach noch nicht** — erst [App.FaktorBestaetigen].
func (a *App) FaktorEinrichten(ctx context.Context, nutzerID string) Einrichtungsantwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return Einrichtungsantwort{Codes: []string{}, Fehler: befund}
	}

	if fremd := a.ueberNutzer(ctx, kennung); fremd != nil {
		return Einrichtungsantwort{Codes: []string{}, Fehler: Uebersetze(ctx, fremd)}
	}

	eingerichtet, fehler := a.faktor.TOTPEinrichten(ctx, kennung)
	if fehler != nil {
		return Einrichtungsantwort{Codes: []string{}, Fehler: Uebersetze(ctx, fehler)}
	}

	return Einrichtungsantwort{
		Adresse:   eingerichtet.Adresse,
		Geheimnis: eingerichtet.Geheimnis,
		Codes:     eingerichtet.Codes,
	}
}

// FaktorBestaetigen trägt den TOTP-Faktor ein, wenn ein Code aus ihm stimmt.
func (a *App) FaktorBestaetigen(ctx context.Context, nutzerID, code string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	if fremd := a.ueberNutzer(ctx, kennung); fremd != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fremd)}
	}

	if fehler := a.faktor.TOTPBestaetigen(ctx, kennung, code); fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return SchreibAntwort{}
}

// BiometrieEinrichten trägt den biometrischen Faktor ein.
//
// **Er verlangt einmal den Finger** — sonst trüge jemand einen Faktor ein, den
// er nicht bedienen kann.
func (a *App) BiometrieEinrichten(ctx context.Context, nutzerID string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	if fremd := a.ueberNutzer(ctx, kennung); fremd != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fremd)}
	}

	if fehler := a.faktor.BiometrieEinrichten(ctx, kennung); fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return SchreibAntwort{}
}

// FaktorEntfernen nimmt einen zweiten Faktor heraus.
func (a *App) FaktorEntfernen(ctx context.Context, nutzerID, art string) SchreibAntwort {
	ctx = a.imProfil(ctx)

	kennung, befund := benannteKennungAus("nutzerID", nutzerID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	if fremd := a.ueberNutzer(ctx, kennung); fremd != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fremd)}
	}

	if fehler := a.faktor.Entfernen(
		ctx, kennung, domain.Faktorart(art)); fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return SchreibAntwort{}
}

// biometriestand holt die Auskunft über diesen Rechner.
//
// **Über den Dienst und nicht über `internal/biometrie`**: `app` entscheidet
// nichts, und ein Ruf ins Blatt an dieser Stelle wäre eine zweite Verdrahtung
// neben der im Composition Root (§2).
func (a *App) biometriestand(ctx context.Context) FaktorenAntwort {
	if a.biometrie == nil {
		return FaktorenAntwort{Faktoren: []FaktorDTO{}}
	}

	auskunft := a.biometrie(ctx)

	return FaktorenAntwort{
		Faktoren:       []FaktorDTO{},
		Biometrie:      auskunft.Vorhanden,
		Biometrieart:   auskunft.Art,
		Biometriegrund: auskunft.Grund,
	}
}
