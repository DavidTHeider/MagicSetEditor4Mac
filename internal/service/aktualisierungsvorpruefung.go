package service

// **Die Vorprüfung vor einem Fassungswechsel** — Werkstück O6 von Phase 5.2
// (#1430).
//
// # Zwei Fragen, und sie sind nicht dieselbe
//
//  1. **Kann diese Fassung diesen Bestand öffnen?** Das entscheidet
//     `domain.Vertraeglichkeit` aus dem Manifest und dem echten Ledger.
//  2. **Kann dieser Rechner den Tausch jetzt überstehen?** Das entscheidet
//     sich am Augenblick: Platz, laufende Arbeit, eine Sicherung, die man
//     wirklich zurückspielen kann.
//
// Die zweite ist die, die das Issue „Vorprüfung" nennt — und ihre schärfste
// Zeile ist die letzte: **„Sicherung ist nachweislich wiederherstellbar,
// bevor ein Update als vorbereitet gilt."** Eine Sicherung, die niemand
// gelesen hat, ist eine Behauptung.
//
// # Jede Quelle ist wahlfrei — und ihr Fehlen ist ein Hindernis, kein Schweigen
//
// **Wer eine Frage nicht stellen kann, hat sie nicht beantwortet.** Fehlt die
// Sicherungsprobe, steht `keineSicherung` da; fehlt die Platzquelle, wird
// nicht so getan, als sei genug frei. Das ist dieselbe Regel wie „keine
// erfundenen Werte" beim Manifest.

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sicherung"
)

// Ledgerquelle sagt, welche Wanderungen dieser Bestand verbucht hat.
type Ledgerquelle interface {
	Ledger(ctx context.Context) (map[string]string, error)
}

// Platzquelle sagt, wie viele Bytes auf dem Volumen des Bestands frei sind.
type Platzquelle interface {
	FreierPlatz() (int64, error)
}

// Vorgangsquelle sagt, wie viel gerade läuft.
type Vorgangsquelle interface {
	LaufendeVorgaenge(ctx context.Context) (int, error)
}

// Halterquelle nennt die anderen Programme, die den Bestand offen halten.
type Halterquelle interface {
	Halter(ctx context.Context) ([]string, error)
}

// Sicherungsprobe macht eine Sicherung **und liest sie zurück**.
//
// **Beides in einem Verfahren**, weil das Zurücklesen die Aussage ist: eine
// Sicherung, deren Manifest sich nicht lesen lässt, hat keinen Wert, und
// zwei getrennte Rufe ließen die Lücke dazwischen offen.
type Sicherungsprobe interface {
	SichereUndPruefe(ctx context.Context) error
}

// Vorpruefer beantwortet beide Fragen.
type Vorpruefer struct {
	ledger    Ledgerquelle
	platz     Platzquelle
	vorgaenge Vorgangsquelle
	halter    Halterquelle
	sicherung Sicherungsprobe

	// plattform ist die dieses Rechners — eingesetzt und nicht gelesen,
	// damit ein Lauf die fremde prüfen kann.
	plattform string
}

// NeuerVorpruefer baut ihn über dem Ledger.
func NeuerVorpruefer(ledger Ledgerquelle, plattform string) *Vorpruefer {
	return &Vorpruefer{ledger: ledger, plattform: plattform}
}

// MitPlatz hängt die Frage nach dem freien Speicher ein — eine der drei des
// Augenblicks, neben [Vorpruefer.MitVorgaengen] und
// [Vorpruefer.MitSicherungsprobe].
func (v *Vorpruefer) MitPlatz(quelle Platzquelle) *Vorpruefer {
	v.platz = quelle

	return v
}

// MitVorgaengen hängt die Frage nach der laufenden Arbeit ein.
func (v *Vorpruefer) MitVorgaengen(quelle Vorgangsquelle) *Vorpruefer {
	v.vorgaenge = quelle

	return v
}

// MitHaltern hängt die Frage ein, wer den Bestand sonst offen hält (M02).
func (v *Vorpruefer) MitHaltern(quelle Halterquelle) *Vorpruefer {
	v.halter = quelle

	return v
}

// MitSicherungsprobe hängt den Beweis ein: sichern **und** zurücklesen.
func (v *Vorpruefer) MitSicherungsprobe(probe Sicherungsprobe) *Vorpruefer {
	v.sicherung = probe

	return v
}

// Pruefe gibt den Befund für eine Freigabe.
//
// **`bedarf` ist, was das Abbild belegt** — die Größe der Beigabe. Verlangt
// wird das Doppelte: einmal für die geladene Datei und einmal für das, was
// beim Auspacken daraus wird. Wer knapp rechnet, füllt die Platte genau
// während des Tauschs.
func (v *Vorpruefer) Pruefe(
	ctx context.Context, manifest domain.Fassungsmanifest, bedarf int64,
) (domain.Aktualisierungsbefund, error) {
	befund, fehler := v.Urteil(ctx, manifest)
	if fehler != nil {
		return domain.Aktualisierungsbefund{}, fehler
	}

	urteil, hindernisse := befund.Urteil, befund.Hindernisse

	// **Bei „ungeklärt" wird der Augenblick nicht mehr geprüft.** Er sagt
	// nichts über eine Fassung, über die nichts bekannt ist — und eine Liste
	// aus einem unbekannten Urteil und vier Nebenbefunden liest niemand.
	if urteil == domain.UrteilUngeklaert {
		return domain.Bereitschaft(urteil, hindernisse), nil
	}

	augenblick, fehler := v.augenblick(ctx, bedarf)
	if fehler != nil {
		return domain.Aktualisierungsbefund{}, fehler
	}

	hindernisse = append(hindernisse, augenblick...)

	if len(hindernisse) != 0 {
		urteil = domain.UrteilUnvertraeglich
	}

	return domain.Bereitschaft(urteil, hindernisse), nil
}

// Urteil beantwortet die **erste** Frage allein: kann diese Fassung diesen
// Bestand öffnen?
//
// # Warum sie getrennt zu haben ist
//
// **Weil sie überall dieselbe Antwort hat und die zweite nicht** (§16). Ob
// eine Fassung passt, hängt am Manifest und am Ledger — das kann jede Fläche
// sagen. Ob der Tausch **jetzt** vorbereitet werden darf, hängt daran, dass
// jemand eine Sicherung macht und den Bestand dafür anhält; das tut nur das
// Fenster, und eine Kommandozeile, die es auf Zuruf täte, nähme den Bestand
// jemandem weg, der gerade arbeitet.
func (v *Vorpruefer) Urteil(
	ctx context.Context, manifest domain.Fassungsmanifest,
) (domain.Aktualisierungsbefund, error) {
	verbucht, fehler := v.ledger.Ledger(ctx)
	if fehler != nil {
		return domain.Aktualisierungsbefund{}, fehler
	}

	urteil, hindernisse := domain.Vertraeglichkeit(manifest, verbucht, v.plattform)

	return domain.Bereitschaft(urteil, hindernisse), nil
}

// augenblick prüft Platz, laufende Arbeit und die Sicherung.
func (v *Vorpruefer) augenblick(
	ctx context.Context, bedarf int64,
) ([]domain.Hindernis, error) {
	hindernisse := make([]domain.Hindernis, 0, 3)

	if v.platz == nil {
		hindernisse = append(hindernisse, domain.Hindernis{Art: domain.HindernisPlatz})
	} else {
		frei, fehler := v.platz.FreierPlatz()
		if fehler != nil {
			return nil, fehler
		}

		if frei < 2*bedarf {
			hindernisse = append(hindernisse, domain.Hindernis{Art: domain.HindernisPlatz})
		}
	}

	if v.vorgaenge != nil {
		laufen, fehler := v.vorgaenge.LaufendeVorgaenge(ctx)
		if fehler != nil {
			return nil, fehler
		}

		if laufen != 0 {
			hindernisse = append(hindernisse,
				domain.Hindernis{Art: domain.HindernisVorgaengeLaufen})
		}
	}

	if v.halter != nil {
		halter, fehler := v.halter.Halter(ctx)
		if fehler != nil {
			return nil, fehler
		}

		for _, name := range halter {
			hindernisse = append(hindernisse, domain.Hindernis{Art: domain.HindernisHalter, Stelle: name})
		}
	}

	return append(hindernisse, v.sicherungshindernis(ctx)...), nil
}

// sicherungshindernis macht die Sicherung und liest sie zurück.
//
// **Ein Fehlschlag ist kein Fehler dieses Wegs**, sondern ein Hindernis: die
// Vorprüfung soll sagen, was im Weg steht, und nicht selbst abbrechen. Wer
// ihren Fehler durchreichte, bekäme statt einer Liste eine Meldung — und die
// anderen drei Antworten gar nicht.
func (v *Vorpruefer) sicherungshindernis(ctx context.Context) []domain.Hindernis {
	if v.sicherung == nil {
		return []domain.Hindernis{{Art: domain.HindernisKeineSicherung}}
	}

	if fehler := v.sicherung.SichereUndPruefe(ctx); fehler != nil {
		// **Der Grund geht ins Protokoll, nicht in den Befund** (§8): eine
		// Meldung aus dem Unterbau erreicht keine Oberfläche, und die
		// Vorgangskennung dieses Rufs (O9) verbindet beides.
		slog.Default().ErrorContext(ctx,
			"die Sicherung ließ sich nicht zurücklesen", "grund", fehler.Error())

		return []domain.Hindernis{{Art: domain.HindernisSicherungUnlesbar}}
	}

	return nil
}

// SicherungsprobeUeber macht aus einem Sicherungsverfahren einen Beweis.
//
// # Warum das hier steht und nicht in `internal/app`
//
// **Weil `internal/app` `internal/sicherung` nicht kennen darf** (§2), und
// der Schichtenwächter hat es angestrichen. Wohin die Probe geschrieben und
// **wie sie zurückgelesen** wird, ist dieselbe Sorte Entscheidung wie bei der
// Bestandssicherung nebenan.
//
// # Über eine Funktion und nicht über den Weg selbst
//
// Das Zurücklesen ist die Aussage dieser Naht, und ein Lauf, der es scheitern
// lassen will, muss eine Sicherung schreiben können, die sich **nicht** lesen
// lässt. Mit einem fertigen Bestandsweg ginge das nicht — er schreibt immer
// eine gute.
func SicherungsprobeUeber(sichern func(ctx context.Context, ziel string) error) Sicherungsprobe {
	return sicherungsprobe{sichern: sichern}
}

type sicherungsprobe struct {
	sichern func(ctx context.Context, ziel string) error
}

func (s sicherungsprobe) SichereUndPruefe(ctx context.Context) error {
	ordner, fehler := os.MkdirTemp("", "mse-sicherungsprobe-")
	if fehler != nil {
		return fmt.Errorf("die Probe vorbereiten: %w", fehler)
	}

	defer func() { _ = os.RemoveAll(ordner) }() //nolint:errcheck // Probeverzeichnis

	ziel := filepath.Join(ordner, "probe"+sicherung.Endung)
	if fehler := s.sichern(ctx, ziel); fehler != nil {
		return fmt.Errorf("die Probe schreiben: %w", fehler)
	}

	// **Zurückgelesen wird das Manifest**, nicht nur die Datei: es nennt
	// Größe und SHA-256 je Teil, und `Entpacken` prüft beides. Ohne diesen
	// Schritt hieße „gesichert" nur „eine Datei liegt da".
	if _, fehler := sicherung.EntpackenNur(ziel,
		filepath.Join(ordner, "aus"), sicherung.Manifestname); fehler != nil {
		return fmt.Errorf("die Probe zurücklesen: %w", fehler)
	}

	return nil
}
