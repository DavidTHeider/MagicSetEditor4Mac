// **Der wirksame Akteur aus dem Kontext** — Werkstück O1 von Phase 5.2
// (#1425),
// nach [Vermerk 0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md).
//
// # Wozu
//
// Jede Fachoperation soll denselben Akteur fragen — nicht jede ihre eigene
// Vorstellung davon, wer arbeitet. Der Akteur entsteht hier aus vier Quellen,
// die es alle schon gibt: dem Profil im Kontext (über die Ablage, die den Kontextschlüssel kennt), dem
// Zeiger „wer arbeitet" (0074), den Freigaben (0070) und dem, was die Fläche
// beim Aufbau an den Kontext hängt — Token und Modus.
//
// # Was hier **nicht** geschieht
//
// **Keine zweite Rechtearchitektur.** Die Rollen bleiben die drei aus 0070,
// der Schreibschutz des Lesers bleibt im Treiber (0073), das Token behält
// seinen Umfang (0068). Hier wird zusammengeführt, nicht neu erfunden — und
// die Schnittmenge rechnet [domain.Akteur.Darf].
//
// # Wer ruft
//
// **Nicht jede Bindung** — das wäre die Zusage, die beim ersten vergessenen
// Ruf still bricht (0070). Gerufen wird dort, wo eine Wirkung über den
// Bestand hinausgeht und der Treiber deshalb nichts hält: beim Schreiben auf
// die Platte, beim Senden nach draußen, beim Abschließen. Welche Fähigkeit
// welche Wirkung hat, steht in `docs/AUTORISIERUNG.md`, und ein Wächter hält
// die Liste gegen den Katalog.

package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

type (
	dienstmodusschluessel struct{}
	tokenschluessel       struct{}
)

// MitDienstmodus hängt an einen Kontext die Fläche, über die ein Aufruf
// kommt — gesetzt in der Composition Root jeder Fläche, nirgends sonst.
func MitDienstmodus(ctx context.Context, modus domain.Dienstmodus) context.Context {
	return context.WithValue(ctx, dienstmodusschluessel{}, modus)
}

// DienstmodusAus liest die Fläche aus dem Kontext.
//
// **Ohne Angabe gilt das Fenster.** Es ist die Fläche, an der ein Mensch
// sitzt, und die einzige, deren Wurzel keinen Kontext baut — Wails reicht
// einen nackten herein. Die drei anderen Wurzeln sagen ausdrücklich, wer sie
// sind; ein Test hält es je Wurzel.
func DienstmodusAus(ctx context.Context) domain.Dienstmodus {
	if modus, da := ctx.Value(dienstmodusschluessel{}).(domain.Dienstmodus); da {
		return modus
	}

	return domain.ModusFenster
}

// MitTokenumfang hängt an einen Kontext den Umfang des Zugriffstokens, über
// das ein Aufruf kam — gesetzt von der Tokenwache, nirgends sonst.
func MitTokenumfang(ctx context.Context, umfang domain.Scope) context.Context {
	return context.WithValue(ctx, tokenschluessel{}, umfang)
}

// tokenidentitaet ist der Schlüssel der Maschinenidentität im Kontext.
type tokenidentitaet struct{}

// MitTokenidentitaet hängt die **Maschinenidentität** des Tokens an.
//
// **Sie ist Profil und Principal, nicht das Geheimnis** — Befund 5 des
// Zweitblicks über Phase 5.2. Der Dienst budgetierte nach dem Hash des
// vorgezeigten Klartexts: eine **Rotation** gab damit frisches Guthaben, und
// zwei Token derselben Maschine hatten zwei Töpfe. Beides widerspricht der
// Naht zwischen O2 und O5 ([0080](../../docs/entscheidungen/0080-maschinenzugaenge-und-widerrufbare-sitzungen.md),
// [0083](../../docs/entscheidungen/0083-ressourcenbudgets-ohne-aussperrung.md)):
// der Principal **ist** die Maschine.
func MitTokenidentitaet(ctx context.Context, identitaet string) context.Context {
	return context.WithValue(ctx, tokenidentitaet{}, identitaet)
}

// TokenidentitaetAus liest sie — leer, wenn kein Token im Spiel ist.
func TokenidentitaetAus(ctx context.Context) string {
	if identitaet, da := ctx.Value(tokenidentitaet{}).(string); da {
		return identitaet
	}

	return ""
}

// TokenumfangAus liest den Umfang — leer, wenn kein Token im Spiel ist:
// am Fenster, an der Kommandozeile und mit der Prozessmarke.
func TokenumfangAus(ctx context.Context) domain.Scope {
	if umfang, da := ctx.Value(tokenschluessel{}).(domain.Scope); da {
		return umfang
	}

	return ""
}

// Arbeitszeiger sagt, wer arbeitet — der Ausschnitt aus [Profilzeiger], den
// die Autorisierung braucht.
type Arbeitszeiger interface {
	Nutzerzeiger(ctx context.Context) (int64, error)

	// Arbeitsprofil sagt, in welchem Profil dieser Aufruf arbeitet — was der
	// Kontext trägt (ein Token, ein Wechsel), sonst der Zeiger. **Die Ablage
	// weiß beides**; dieser Dienst kennt den Kontextschlüssel nicht (§2).
	Arbeitsprofil(ctx context.Context) (int64, error)

	// Revision ist der Stand der Rechte — sie steigt mit jedem Wechsel und
	// jeder Freigabe, **in der Ablage** und nicht im Aufrufer: die Kaskade
	// beim Löschen eines Profils ist ein Entzug, den kein Aufrufer sieht.
	Revision(ctx context.Context) (int64, error)
}

// Autorisierung baut den Akteur und prüft Wirkungen gegen ihn.
type Autorisierung struct {
	zeiger    Arbeitszeiger
	freigaben *Freigabenverwaltung

	// nachweise löst Nachweise ein — `nil` an einer Fläche, die keine
	// ausstellt; dann kommt nichts durch, was einen braucht.
	nachweise *Nachweise
}

// NeueAutorisierung baut sie über dem Zeiger und der Freigabenverwaltung —
// die ihrerseits die Profile kennt.
func NeueAutorisierung(zeiger Arbeitszeiger, freigaben *Freigabenverwaltung) *Autorisierung {
	return &Autorisierung{zeiger: zeiger, freigaben: freigaben}
}

// Akteur sagt, wer dieser Aufruf ist.
//
// **Die Reihenfolge der Fehler ist die Reihenfolge der Fragen**: kein Profil
// im Kontext ist eine Frage an den Aufruf (`ErrValidation`), kein Zeiger eine
// an den Bestand, ein unbekanntes Profil eine an die Kennung (`ErrNotFound`).
func (a *Autorisierung) Akteur(ctx context.Context) (domain.Akteur, error) {
	profilID, fehler := a.zeiger.Arbeitsprofil(ctx)
	if fehler != nil {
		return domain.Akteur{}, fehler
	}

	nutzerID, fehler := a.zeiger.Nutzerzeiger(ctx)
	if fehler != nil {
		return domain.Akteur{}, fehler
	}

	if nutzerID == 0 {
		return domain.Akteur{}, fmt.Errorf(
			"%w: der Bestand sagt nicht, wer arbeitet", domain.ErrNotFound)
	}

	profil, fehler := a.freigaben.profilSuchen(ctx, profilID)
	if fehler != nil {
		return domain.Akteur{}, fehler
	}

	rolle, fehler := a.freigaben.Rolle(ctx, nutzerID, profil)
	if fehler != nil {
		return domain.Akteur{}, fehler
	}

	revision, fehler := a.zeiger.Revision(ctx)
	if fehler != nil {
		return domain.Akteur{}, fehler
	}

	return domain.Akteur{
		NutzerID: nutzerID,
		ProfilID: profilID,
		EignerID: profil.NutzerID,
		Rolle:    rolle,

		// **Ein Leser liest über eine zweite Verbindung** (0073) — das ist
		// dieselbe Bedingung, unter der das Fenster sie aufmacht.
		NurLesend: rolle == domain.RolleLeser,
		Token:     TokenumfangAus(ctx),
		Modus:     DienstmodusAus(ctx),
		Revision:  revision,
	}, nil
}

// Pruefe weist ab, was dieser Aufruf nicht darf — ohne Nachweis.
//
// **Was einen Nachweis braucht, kommt hier nicht durch**, auch nicht für den
// Verwalter: dafür gibt es [Autorisierung.PruefeMitNachweis], und der Satz
// sagt es, damit niemand an der Rolle sucht.
func (a *Autorisierung) Pruefe(ctx context.Context, wirkung domain.Wirkung) error {
	return gezaehlt(ctx, wirkung, a.pruefe(ctx, wirkung))
}

func (a *Autorisierung) pruefe(ctx context.Context, wirkung domain.Wirkung) error {
	akteur, fehler := a.Akteur(ctx)
	if fehler != nil {
		return fehler
	}

	if fehler := akteur.Darf(wirkung); fehler != nil {
		return fehler
	}

	if wirkung.BrauchtNachweis() {
		return fmt.Errorf("%w: %q braucht einen Nachweis — die Rolle allein "+
			"reicht dafür nicht", domain.ErrErlaubnis, string(wirkung))
	}

	return nil
}
