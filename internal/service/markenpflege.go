package service

// **Marken an Karten** — Werkstück 3 von Phase 3.5
// (#1208).
//
// # Was dieser Anwendungsfall entscheidet
//
// **Dass eine Marke ihre Normalform bekommt, bevor sie in den Bestand geht.**
// Die Regel steht in [domain.PruefeMarke]; hier steht, dass **jeder** Weg sie
// durchläuft — das Fenster, die Kommandozeile, MCP und der Dienst. Wer sie an
// einer Fläche prüfte, hätte drei Flächen, an denen „Fertig" und „fertig" zwei
// Marken sind (§16).
//
// # Und dass die Karte zum genannten Set gehört
//
// Die Fremdschlüssel des Bestands halten fest, dass es die Karte **gibt**;
// dass sie zu dem Set gehört, in dem jemand gerade arbeitet, hält keiner.
// Ohne diese Prüfung markierte ein vertippter Aufruf eine fremde Karte, und
// niemand fände sie wieder.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Markenablage ist, was die Markenpflege vom Bestand braucht.
type Markenablage interface {
	Setzen(ctx context.Context, karteID int64, marke string) error
	Entfernen(ctx context.Context, karteID int64, marke string) error
	AnKarte(ctx context.Context, karteID int64) ([]string, error)
	ImSet(ctx context.Context, setID int64) ([]string, error)
	JeKarte(ctx context.Context, setID int64) (map[int64][]string, error)
}

// Markenpflege setzt, entfernt und zählt Marken auf.
type Markenpflege struct {
	ablage Markenablage
	karten Kartensicht

	// koepfe sagt, ob es ein Set gibt.
	//
	// **Weil `storage` für ein unbekanntes Set eine leere Liste gibt** und
	// keinen Fehler. Im Fenster ist das richtig — dort kann niemand ein Set
	// wählen, das es nicht gibt. An der Kommandozeile, im Dienst und über MCP
	// kann er es: `mse marken 4711` ist ein Tippfehler, und „keine Marke" wäre
	// darauf eine irreführende Antwort.
	//
	// **Hier und nicht in jedem Aufrufer** (§16): sonst hätte die Fassade den
	// Fehler und die drei anderen Flächen nicht. Der Kartenbefehl hat es
	// einzeln gelöst, und genau daran ist es aufgefallen. Angestrichen vom
	// Nachlauf über die Nacharbeit
	// (#1282,
	// Befund 8).
	koepfe Kopfquelle
}

// NeueMarkenpflege baut den Anwendungsfall.
func NeueMarkenpflege(
	ablage Markenablage, karten Kartensicht, koepfe Kopfquelle,
) *Markenpflege {
	return &Markenpflege{ablage: ablage, karten: karten, koepfe: koepfe}
}

// Setzen hängt eine Marke an eine Karte des genannten Sets.
func (m *Markenpflege) Setzen(ctx context.Context, setID, karteID int64, roh string) error {
	marke, err := m.geprueft(ctx, setID, karteID, roh)
	if err != nil {
		return err
	}

	return m.ablage.Setzen(ctx, karteID, marke)
}

// Entfernen nimmt eine Marke wieder ab.
func (m *Markenpflege) Entfernen(ctx context.Context, setID, karteID int64, roh string) error {
	marke, err := m.geprueft(ctx, setID, karteID, roh)
	if err != nil {
		return err
	}

	return m.ablage.Entfernen(ctx, karteID, marke)
}

// AnKarte gibt die Marken einer Karte.
func (m *Markenpflege) AnKarte(ctx context.Context, setID, karteID int64) ([]string, error) {
	if err := m.karteImSet(ctx, setID, karteID); err != nil {
		return nil, err
	}

	return m.ablage.AnKarte(ctx, karteID)
}

// ImSet gibt die Marken, die in einem Set vorkommen — jede einmal.
func (m *Markenpflege) ImSet(ctx context.Context, setID int64) ([]string, error) {
	if err := m.setMussGeben(ctx, setID); err != nil {
		return nil, err
	}

	return m.ablage.ImSet(ctx, setID)
}

// JeKarte gibt die Marken aller Karten eines Sets.
func (m *Markenpflege) JeKarte(ctx context.Context, setID int64) (map[int64][]string, error) {
	if err := m.setMussGeben(ctx, setID); err != nil {
		return nil, err
	}

	return m.ablage.JeKarte(ctx, setID)
}

// setMussGeben weist ein unbekanntes Set ab.
func (m *Markenpflege) setMussGeben(ctx context.Context, setID int64) error {
	koepfe, err := m.koepfe.Setkoepfe(ctx)
	if err != nil {
		return fmt.Errorf("die Sets lesen: %w", err)
	}

	for i := range koepfe {
		if koepfe[i].ID == setID {
			return nil
		}
	}

	return fmt.Errorf("%w: das Set %d gibt es nicht", domain.ErrNotFound, setID)
}

// geprueft gibt die Normalform — nachdem die Karte zum Set gehört.
//
// **In dieser Reihenfolge.** Wer die Marke zuerst prüfte, sagte bei einer
// vertippten Kartennummer „die Marke ist zu lang" — eine Meldung über das
// falsche Feld.
func (m *Markenpflege) geprueft(
	ctx context.Context, setID, karteID int64, roh string,
) (string, error) {
	if err := m.karteImSet(ctx, setID, karteID); err != nil {
		return "", err
	}

	return domain.PruefeMarke(roh)
}

// karteImSet prüft, dass es die Karte gibt und sie zu dem Set gehört.
func (m *Markenpflege) karteImSet(ctx context.Context, setID, karteID int64) error {
	karte, err := m.karten.KarteLesen(ctx, karteID)
	if err != nil {
		return err
	}

	if karte.SetID != setID {
		return fmt.Errorf("%w: die Karte %d gehört nicht zum Set %d",
			domain.ErrNotFound, karteID, setID)
	}

	return nil
}
