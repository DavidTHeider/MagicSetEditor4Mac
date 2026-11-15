package service

// **Die Betriebsschau: lebt er, kann er arbeiten** — Werkstück O9 von Phase
// 5.2 (#1433).
//
// # Zwei Fragen, und sie sind nicht dieselbe
//
//   - **Lebt dieser Prozess?** Wer die Antwort liest, hat den Beweis in der
//     Hand. Das ist keine Nachlässigkeit, sondern die Definition: eine
//     Lebendprüfung, die etwas anderes prüfte, prüfte etwas anderes.
//   - **Kann er arbeiten?** Das hängt am Bestand — und **nur** am Bestand.
//
// # Warum die fremden Dienste nicht mitrechnen
//
// **Weil ein Ausfall bei Scryfall diesen Editor nicht anhält** (§16, Satz 5).
// Sie stehen in der Antwort, damit jemand sieht, warum ein Bild fehlt; in die
// Bereitschaft gehen sie nicht ein. Ein Betriebswerkzeug, das darauf hörte,
// startete den Editor neu, weil ein fremder Server gerade langsam ist.
//
// # Warum zwei schmale Nähte und keine breite
//
// Dieselbe Überlegung wie bei `Kartenspiel` im Deckbau: was die Schau vom
// Bestand braucht, sind **zwei** Auskünfte — antwortet er, und wandert er
// gerade. Eine Schnittstelle mit mehr machte jedes Doppel zu einer
// Bestandsverwaltung.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Bestandslage ist, was die Schau vom Bestand wissen muss.
//
// **`Probe` fragt etwas Triviales** — ein `SELECT 1`. Sie beantwortet „bist du
// da und antwortest du", nicht „ist alles in Ordnung": eine Prüfung, die den
// ganzen Bestand läse, wäre bei jedem Aufruf teurer als die Arbeit, die sie
// schützen soll.
type Bestandslage interface {
	Probe(ctx context.Context) error

	// WanderungLaeuft sagt, ob gerade Migrationen laufen.
	//
	// **Getrennt von der Probe**, weil ein wandernder Bestand auf ein
	// `SELECT 1` sehr wohl antwortet — und trotzdem nicht bereit ist.
	WanderungLaeuft() bool
}

// Fremdschau sagt, was über die fremden Dienste bekannt ist.
//
// **Ohne Fehlerrückgabe**, und das ist Absicht: sie beantwortet eine Frage
// über Auskünfte, die schon da sind. Ein Ruf ins Netz an dieser Stelle machte
// aus einer Zustandsabfrage einen Netzweg — und ein langsamer Dienst
// verzögerte die Antwort auf „lebst du".
type Fremdschau interface {
	Fremdzustaende(ctx context.Context) []domain.Fremdzustand
}

// Betriebsschau beantwortet beide Fragen.
type Betriebsschau struct {
	bestand Bestandslage
	fremde  Fremdschau
}

// NeueBetriebsschau baut sie — **beide Nähte dürfen fehlen**.
//
// Ohne Bestand ist er zu; ohne Fremdschau gibt es keine Auskunft über fremde
// Dienste. Beides ist ein gültiger Zustand und kein Fehler: der MCP-Server
// läuft ohne fremde Dienste, und das Fenster hat vor der Profilwahl keinen
// Bestand.
func NeueBetriebsschau(bestand Bestandslage, fremde Fremdschau) *Betriebsschau {
	return &Betriebsschau{bestand: bestand, fremde: fremde}
}

// Zustand sagt, wie es steht.
func (b *Betriebsschau) Zustand(ctx context.Context) domain.Betriebszustand {
	return domain.Betriebszustand{
		Lebt:   true,
		Grund:  b.grund(ctx),
		Fremde: b.fremdzustaende(ctx),
	}
}

// grund gibt den **eigenen** Grund, nicht bereit zu sein — leer, wenn keiner.
//
// **Die Reihenfolge ist die ihrer Genauigkeit**: kein Bestand ist etwas
// anderes als ein wandernder, und ein wandernder etwas anderes als einer, der
// schweigt.
func (b *Betriebsschau) grund(ctx context.Context) domain.Bereitschaftsgrund {
	if b.bestand == nil {
		return domain.GrundBestandZu
	}

	if b.bestand.WanderungLaeuft() {
		return domain.GrundWanderungLaeuft
	}

	if b.bestand.Probe(ctx) != nil {
		return domain.GrundBestandUnlesbar
	}

	return ""
}

// fremdzustaende gibt die Auskunft — **leer und nie `nil`** (§3).
func (b *Betriebsschau) fremdzustaende(ctx context.Context) []domain.Fremdzustand {
	if b.fremde == nil {
		return []domain.Fremdzustand{}
	}

	gelesen := b.fremde.Fremdzustaende(ctx)
	if gelesen == nil {
		return []domain.Fremdzustand{}
	}

	return gelesen
}
