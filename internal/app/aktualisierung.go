package app

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/aktualisierung"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Nach einer neueren Fassung sehen** —
// #1100.
//
// # Nichts geschieht ohne Einstellung
//
// §16 Regel 6: nichts wird still übernommen. Die Vorgabe ist `aus`, und dann
// **fragt dieser Weg gar nicht** — er geht nicht ins Netz, und die Antwort
// sagt, dass nicht gefragt wurde. Das ist etwas anderes als „nichts Neues",
// und die Antwort unterscheidet beides.

// Fassungsanleger sagt, welche Freigabe die neueste ist.
//
// **Als Schnittstelle hier und umgesetzt in `internal/aktualisierung`** — die
// gewohnte Richtung (§2): sie steht, wo sie gebraucht wird.
type Fassungsanleger interface {
	NeuesteFreigabe(ctx context.Context) (aktualisierung.Freigabe, error)

	// Laden holt das Abbild einer Freigabe und gibt seinen Ort.
	//
	// **Am selben Anleger und nicht an einem zweiten**: wer die neueste
	// Freigabe kennt, kennt auch ihre Beigaben — eine zweite Naht daneben
	// führte denselben Zugang zweimal.
	Laden(ctx context.Context, freigabe aktualisierung.Freigabe, verzeichnis string,
		melden aktualisierung.Fortschritt) (string, error)

	// Manifest gibt, was eine Freigabe über sich sagt — geprüft gegen die
	// **signierte** Prüfsummenliste (O6, #1430).
	//
	// **Am selben Anleger**, aus demselben Grund wie [Fassungsanleger.Laden]:
	// dieselbe Freigabe, dieselben Beigaben, derselbe Zugang.
	Manifest(ctx context.Context, freigabe aktualisierung.Freigabe,
		pruefsummen []byte) (domain.Fassungsmanifest, error)

	// Pruefsummen holt die **signierte** Liste und prüft ihre Signatur.
	Pruefsummen(ctx context.Context, freigabe aktualisierung.Freigabe) ([]byte, error)
}

// FassungsAntwort ist, was das Fenster über eine neuere Fassung erfährt.
type FassungsAntwort struct {
	// Gefragt sagt, ob überhaupt gefragt wurde.
	//
	// **Getrennt von `Neuer`**, weil „aus" und „nichts Neues" verschiedene
	// Dinge sind. Ohne diese Unterscheidung sähe eine abgeschaltete Prüfung
	// aus wie eine, die nichts gefunden hat.
	Gefragt bool `json:"gefragt"`

	// Neuer sagt, ob die gefundene Fassung neuer ist als die laufende.
	Neuer bool `json:"neuer"`

	// Fassung ist die gefundene Fassung — leer, wenn nicht gefragt wurde.
	Fassung string `json:"fassung"`

	// Seite ist die Freigabeseite, zu der die Meldung führen kann.
	Seite string `json:"seite"`

	// Holen sagt, ob die Stufe das Laden erlaubt — `holen` und nicht nur
	// `melden`. **Die Grenze bleibt hier** ([App.NeueFassungLaden] zieht sie
	// noch einmal); das Fenster zeigt den Knopf nur, wenn sie es erlaubt.
	// Unter `melden` stand er bis AP-12 (#2369) trotzdem da, und ein Druck
	// kam als Fehler zurück.
	Holen bool `json:"holen"`

	Fehler *Fehler `json:"fehler"`
}

// NachNeuerFassungSehen fragt — wenn die Einstellung es erlaubt.
//
// **Der Fehler eines fremden Dienstes hält die Anwendung nicht an** (§16,
// Satz 5): er steht in der Antwort, und weitergearbeitet wird ohnehin.
func (a *App) NachNeuerFassungSehen(ctx context.Context) FassungsAntwort {
	ctx = a.imProfil(ctx)

	werte, err := a.einstellungen.Alle(ctx)
	if err != nil {
		return FassungsAntwort{Fehler: Uebersetze(ctx, err)}
	}

	wahl := werte[domain.EinstellungAktualisierung]

	// **Leer heißt Vorgabe heißt aus.** Wer nie gewählt hat, hat „aus"
	// gewählt.
	if wahl == "" || wahl == domain.AktualisierungAus {
		return FassungsAntwort{}
	}

	if a.fassungen == nil {
		return FassungsAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: der Zugang zu den Freigaben ist nicht eingerichtet",
			domain.ErrValidation))}
	}

	freigabe, err := a.fassungen.NeuesteFreigabe(ctx)
	if err != nil {
		return FassungsAntwort{Gefragt: true, Fehler: Uebersetze(ctx, err)}
	}

	neuer, err := domain.FassungIstNeuer(freigabe.Fassung, Fassung)
	if err != nil {
		return FassungsAntwort{Gefragt: true, Fehler: Uebersetze(ctx, err)}
	}

	return FassungsAntwort{
		Gefragt: true,
		Neuer:   neuer,
		Fassung: freigabe.Fassung,
		Seite:   freigabe.Seite,
		Holen:   wahl == domain.AktualisierungHolen,
	}
}

// MitFassungen setzt den Zugang zu den Freigaben — für Tests und für den
// Composition Root.
func (a *App) MitFassungen(anleger Fassungsanleger) { a.fassungen = anleger }
