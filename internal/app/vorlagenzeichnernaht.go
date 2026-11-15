// **Der Vorlagenzeichner** — Werkstück 3 von Phase 3.3
// (#1164).
//
// Er beantwortet zwei Fragen, die `App` nicht beantworten kann: **welche Karte
// zeigt diesen Stil**, und **wie sieht sie mit der geänderten Datei aus**.

package app

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// Vorlagennaht zeichnet Karten mit einer ungespeicherten Vorlage.
type Vorlagennaht struct {
	werk *Zeichenwerk

	// leser ist die Naht zum Bestand — als Schnittstelle und nicht als
	// `*storage.SetLeser`, damit ein Lauf ihre **zwei** Lesewege einzeln zum
	// Scheitern bringen kann. Ein geschlossener Bestand trifft beide auf
	// einmal, und dann bliebe der zweite unbetreten (§3).
	leser Setkartenquelle
}

// Setkartenquelle ist, was der Vorlagenzeichner aus dem Bestand braucht.
type Setkartenquelle interface {
	Setkoepfe(ctx context.Context) ([]domain.SetHead, error)
	Karten(ctx context.Context, setID int64) ([]*domain.CardRow, error)
	KarteLesen(ctx context.Context, karteID int64) (*domain.CardRow, error)
}

// MitLeser setzt die Bestandsnaht — für Läufe, die einen Lesefehler brauchen.
func (v *Vorlagennaht) MitLeser(leser Setkartenquelle) *Vorlagennaht {
	v.leser = leser

	return v
}

// VorlagennahtUeber baut sie über einer Datenbank.
func VorlagennahtUeber(db *sql.DB) (*Vorlagennaht, error) {
	werk, err := Zeichennaht(db)
	if err != nil {
		return nil, err
	}

	return &Vorlagennaht{werk: werk, leser: storage.NeuerSetLeser(db)}, nil
}

// Probekarte gibt eine Karte, die diesen Stil benutzt.
//
// # Warum die erste und nicht irgendeine
//
// Weil sie **wiederholbar** sein muss: wer eine Zeile ändert und die Vorschau
// zweimal sieht, soll dieselbe Karte sehen. Eine zufällige machte aus jedem
// Tastendruck ein anderes Bild, und niemand wüsste, was die Änderung war.
//
// **Eine genannte Karte kommt, wie sie ist** (U7, #1418) — auch wenn ihr Set
// einen anderen Stil führt. Die Werkstatt hat sie ausdrücklich gewählt, und
// eine Vorschau, die stattdessen still eine andere zeigt, wäre die Gestalt
// von Fehler, gegen die §16 „nichts wird still übernommen" steht.
func (v *Vorlagennaht) Probekarte(
	ctx context.Context, paket string, karteID int64,
) (*domain.CardRow, error) {
	if karteID > 0 {
		return v.leser.KarteLesen(ctx, karteID)
	}

	karten, err := v.Probekarten(ctx, paket)
	if err != nil {
		return nil, err
	}

	if len(karten) == 0 {
		return nil, fmt.Errorf(
			"%w %q — ohne eine Karte gibt es nichts zu zeigen",
			domain.ErrKeinSetMitStil, paketname(paket))
	}

	return karten[0], nil
}

// Probekarten gibt die Karten des ersten Sets, das diesen Stil benutzt und
// mindestens eine Karte hat — die Wahl der Werkstatt (U7, #1418).
//
// **Ein Set ohne Karten zählt nicht**: es gäbe nichts zu zeichnen, und das
// nächste mit dem Stil hat vielleicht welche. So war es schon bei der
// Probekarte; die Suche ist dieselbe, damit Vorgabe und Auswahl nie
// auseinanderlaufen.
func (v *Vorlagennaht) Probekarten(ctx context.Context, paket string) ([]*domain.CardRow, error) {
	gesucht := paketname(paket)

	koepfe, err := v.leser.Setkoepfe(ctx)
	if err != nil {
		return nil, err
	}

	for _, kopf := range koepfe {
		if !passenderStil(kopf.Style, kopf.Game, gesucht) {
			continue
		}

		karten, lesefehler := v.leser.Karten(ctx, kopf.ID)
		if lesefehler != nil {
			return nil, lesefehler
		}

		if len(karten) > 0 {
			return karten, nil
		}
	}

	return []*domain.CardRow{}, nil
}

// MitErsatz gibt Planbauer und Rasterer, die die ersetzte Datei sehen.
//
// **Ein eigener Planbauer über einem eigenen Stilspeicher.** Der gemeinsame
// merkt sich Stile; ein Stand, den niemand geschrieben hat, gehört dort nicht
// hinein.
//
// **Der Rasterer bleibt derselbe.** Er kennt keine Pakete — er bekommt einen
// fertigen Plan und rechnet; ein zweiter wäre eine zweite Antwort auf die
// Frage, wie breit ein Text ist (§2).
func (v *Vorlagennaht) MitErsatz(ersatz service.Paketersatz) (*service.Planbauer, Rasterer) {
	planbauer := service.NeuerPlanbauer().
		MitStilen(v.werk.Stile.MitErsatz(ersatz)).
		MitSchriften(v.werk.Schriften).
		MitBildern(v.werk.Bildquelle)

	return planbauer, v.werk.Rasterer
}

// paketname macht aus einem Paketpfad den Namen des Pakets.
//
// `…/magic-m15.mse-style` → `magic-m15`. **Das Spiel bleibt vorn stehen**: es
// gehört zum Namen des Pakets, und erst der Vergleich unten löst die Kurzform
// eines Sets dagegen auf.
func paketname(paket string) string {
	return strings.TrimSuffix(filepath.Base(paket), "."+string(msepaket.ArtStil))
}

// passenderStil sagt, ob ein Set dieses Stilpaket benutzt.
//
// # Die Rechnung geht in dieselbe Richtung wie beim Öffnen
//
// Ein Set nennt seinen Stil in der **Kurzform** — `stylesheet: m15` bei
// `game: magic` meint `magic-m15.mse-style`; die lange Form kommt daneben vor
// (unsere eigene Saat schreibt sie). Welche Paketnamen dazu passen, rechnet
// `service.stilpaketnamen` aus, und **hier wird dieselbe Rechnung gelesen**:
// `<spiel>-<stil>` oder der rohe Wert, beides **genau** verglichen.
//
// # Warum nicht über die Endung
//
// Der erste Anlauf verglich `strings.HasSuffix(genannt, "-"+kurz)`, und damit
// galt `counter-m15` als `m15`. Die Vorschau nahm dann eine Karte des fremden
// Stils; der Ersatz greift auf ihr nicht, weil er auf den bearbeiteten
// Paketpfad zielt und exakt verglichen wird. Der Weg lief durch, das Bild trug
// die Änderung nicht.
//
// **Gemessen**: 26 solcher Paare unter den 376 Stilpaketen des Packs.
// Angestrichen vom Zweitblick durch Codex über Phase 3.3.
// **Und der Wert wird nicht geglättet.** Der erste Anlauf schnitt Leerraum
// weg; `service.stilpaketnamen` tut das nicht, und ein Set mit
// `stylesheet: "m15 "` galt hier als passend, während das Öffnen sein Paket
// nicht fand — die Vorschau zeigte dann eine Karte des Behelfsplans. Dieselbe
// Rechnung heißt auch dieselbe Vorbereitung. Angestrichen vom Nachlauf des
// Zweitblicks über Phase 3.3.
func passenderStil(genannt, spiel, paketname string) bool {
	// **Ohne Spielangabe gibt es keine Kurzform aufzulösen** — dann bleibt der
	// rohe Wert, so wie beim Öffnen auch.
	if spiel != "" && spiel+"-"+genannt == paketname {
		return true
	}

	return genannt == paketname
}
