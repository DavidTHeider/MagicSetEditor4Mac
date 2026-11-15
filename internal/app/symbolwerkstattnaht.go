// **Der Symboleditor für die Flächen ohne Fenster**
// (#1391).
//
// # Warum es diese Naht gibt
//
// §16: dieselbe Fähigkeit, überall dieselbe Antwort. Für neun Fähigkeiten
// des Symboleditors stand in `docs/FAEHIGKEITEN.md` sinngemäß „ohne
// Zeichenfläche gibt es keinen Punkt, auf den man zeigt" — und das beschreibt
// die **Bedienung**, nicht die Fähigkeit. Nachgesehen am 2026-09-02: die
// Rechnungen dahinter sind **reine Funktionen**, Teile hinein, Teile heraus.
//
// # Wie ein Symbol in einen Aufruf kommt
//
// **Der Aufrufer trägt es.** Die Rechnungen nehmen `teile` entgegen und geben
// `teile` zurück; wer eine Kette bauen will, reicht das Ergebnis weiter.
//
// **Und nicht über einen Sitzungszustand.** Ein „gerade offenes Symbol" gäbe
// es an keiner dieser Flächen: MCP hat keine Sitzung, die Kommandozeile ist
// nach jedem Aufruf fort, und der Dienst bediente damit alle Aufrufer aus
// einem Topf — genau die Frage aus §7, Regel 4. Ein Zustand, den niemand
// besitzt, ist eine Nebenläufigkeitsfrage ohne Antwort.
//
// **Die beiden Enden hängen doch am Bestand**: `Lesen` holt das Symbol eines
// Sets, `Ablegen` schreibt es zurück. Dazwischen ist alles Rechnung.

package app

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Symbolwerkstattnaht liest und schreibt das Symbol eines Sets.
type Symbolwerkstattnaht struct {
	ablage *service.Symbolablage
	db     *sql.DB
	wache  *service.Autorisierung
}

// SymbolwerkstattUeber baut sie über dem Bilderverzeichnis — **und dem
// Bestand**, der sagt, wem das Set gehört.
//
// **Hier stand „ohne Bestand"**, mit dem Grund, ein Symbol liege neben dem Set
// auf der Platte und nicht in der Datenbank. Für die Datei stimmt das; **wem
// das Set gehört, weiß aber nur der Bestand** — und ohne ihn las und schrieb
// jeder das Symbol jedes Sets, dessen Kennung er nannte (#1815).
func SymbolwerkstattUeber(db *sql.DB) (*Symbolwerkstattnaht, error) {
	bilder, fehler := Bilderverzeichnis()
	if fehler != nil {
		return nil, fehler
	}

	return &Symbolwerkstattnaht{
		ablage: service.NeueSymbolablage(service.NeueBildablage(bilder)), db: db,
		wache: AutorisierungAus(db),
	}, nil
}

// Lesen holt das abgelegte Symbol eines Sets des gezeigten Profils.
func (n *Symbolwerkstattnaht) Lesen(ctx context.Context, setID int64, name string) ([]SymbolteilDTO, error) {
	if fehler := SetDaSein(ctx, n.db, setID); fehler != nil {
		return nil, fehler
	}

	zeichnung, err := n.ablage.Gestalt(setID, name)
	if err != nil {
		return nil, err
	}

	return teileAlsDTO(zeichnung.Teile), nil
}

// Ablegen schreibt eine Zeichnung neben ein Set des gezeigten Profils und gibt
// ihren Dateinamen.
func (n *Symbolwerkstattnaht) Ablegen(
	ctx context.Context, setID int64, teile []SymbolteilDTO, unter string,
) (string, error) {
	// **Die Rolle vor dem Set** (C04, #2229): die Zeichnung ist eine Datei,
	// und eine Verbindung, die nichts annimmt, hält sie nicht auf. Ohne diese
	// Zeile schrieb an der Kommandozeile ein Leser in ein fremdes Profil.
	if fehler := n.wache.Pruefe(ctx, domain.WirkungSchreiben); fehler != nil {
		return "", fehler
	}

	if fehler := SetDaSein(ctx, n.db, setID); fehler != nil {
		return "", fehler
	}

	return n.ablage.Ablegen(setID,
		msepaket.Symbolzeichnung{Teile: teileAusDTO(teile)}, unter)
}
