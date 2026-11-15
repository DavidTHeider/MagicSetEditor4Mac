package app

// **Was Platz belegt, und wie er frei wird** — Werkstück 4 von Phase 4.13
// (#1352).
//
// # Der Anlass, gemessen
//
// **Am 2026-09-07 auf dem Rechner des Auftraggebers gezählt**: neben einem
// Bestand von 127 MB lagen **dreißig** Sicherungen vor Migrationen — zusammen
// über zwei Gigabyte. Sie entstehen bei jedem Start, der migriert
// (`sicherungVorMigration`), und **niemand räumte sie weg**.
//
// **Das ist kein Fehler in der Sicherung.** Sie ist richtig, und die jüngste
// ist der Rückweg. Es fehlte der Ort, an dem steht, was daliegt.
//
// # Warum die Zählung hier steht und nicht in `service`
//
// **Weil sie Pfade braucht**, und die kennt diese Schicht: `Datenbankpfad`,
// `Bilderverzeichnis`, `Paketverzeichnis`. Dieselbe Bauform wie `Diagnose`
// nebenan, und aus demselben Grund. **Was fachlich ist — welche Posten es gibt
// und welcher wegdarf —, steht in `domain`** (§2).

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// PlatzpostenDTO sagt, wie viel ein Posten belegt.
type PlatzpostenDTO struct {
	Posten string `json:"posten"`

	// Dateien ist die Zahl der Dateien — **null ist kein Fehler**: wer nie
	// migriert hat, hat keine Vorsicherung.
	Dateien int `json:"dateien"`

	Bytes int64 `json:"bytes"`

	// Raeumbar sagt, ob dieser Posten sich von hier aus wegräumen lässt.
	//
	// **Gerechnet in `domain`**, damit die Antwort nicht an zwei Stellen
	// steht.
	Raeumbar bool `json:"raeumbar"`
}

// PlatzAntwort trägt die Posten.
type PlatzAntwort struct {
	Posten []PlatzpostenDTO `json:"posten"`

	// Nachladefehler sagt: weggeräumt **ist**, nur die Zählung danach war
	// nicht zu machen (§7, Regel 3; #1970) — dann ist `posten` leer und
	// `fehler` auch.
	Nachladefehler *Fehler `json:"nachladefehler,omitempty"`

	Fehler *Fehler `json:"fehler"`
}

// Platz sagt, wie viel Platz was belegt.
func (a *App) Platz(ctx context.Context) PlatzAntwort {
	ctx = a.imProfil(ctx)
	_ = ctx

	// **Eine leere Liste und kein `nil`** (§3).
	liste := []PlatzpostenDTO{}

	// **Was gezählt ist, wird nicht zweimal gezählt** — Werkstück 6 von Phase
	// 4.13. *Sonstiges* ist der Rest, und ein Rest lässt sich nur hier
	// bestimmen: er kennt die fünf anderen erst, wenn sie gelaufen sind.
	// Deshalb steht er in dieser Schleife und nicht in `dateienDesPostens`.
	beansprucht := map[string]bool{}

	for _, posten := range domain.PostenAlle() {
		dateien, fehler := postendateien(posten, a.datenbankpfad, beansprucht)
		if fehler != nil {
			return PlatzAntwort{Posten: []PlatzpostenDTO{}, Fehler: Uebersetze(ctx, fehler)}
		}

		eintrag := PlatzpostenDTO{
			Posten: string(posten), Raeumbar: posten.Raeumbar(),
		}

		for _, datei := range dateien {
			eintrag.Dateien++
			eintrag.Bytes += datei.wiegt
			beansprucht[datei.pfad] = true
		}

		liste = append(liste, eintrag)
	}

	return PlatzAntwort{Posten: liste}
}

// PlatzFreigeben räumt einen Posten weg — bis auf das, was bleiben muss.
//
// **Der Weg zurück bleibt stehen.** Wer nach einer Migration merkt, dass etwas
// fehlt, spielt ihn zurück; wer ihn mit wegräumte, nähme genau die Sicherung,
// für die es sie gibt.
//
// **Welche Datei das ist, steht in `wegzurueck.go`** (Werkstück 5).
func (a *App) PlatzFreigeben(ctx context.Context, posten string) PlatzAntwort {
	ctx = a.imProfil(ctx)

	// **Eine leere Liste, auch im Fehlerfall** (§3): `nil` wird zu `null` im
	// JSON, und das Fenster bekäme dann gar keine Liste statt einer leeren.
	// Der Wächter über die leeren Listen fängt genau das.
	leer := []PlatzpostenDTO{}

	if !domain.PostenGefuehrt(posten) {
		return PlatzAntwort{Posten: leer, Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: %q ist kein Posten — es gibt %v",
			domain.ErrValidation, posten, domain.PostenAlle()))}
	}

	if !domain.Platzposten(posten).Raeumbar() {
		return PlatzAntwort{Posten: leer, Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: %q lässt sich nicht wegräumen — er gehört zu den Daten, mit "+
				"denen die Anwendung arbeitet", domain.ErrValidation, posten))}
	}

	dateien, fehler := dateienDesPostens(domain.Platzposten(posten), a.datenbankpfad)
	if fehler != nil {
		return PlatzAntwort{Posten: leer, Fehler: Uebersetze(ctx, fehler)}
	}

	bleiben := a.bleibende(ctx, domain.Platzposten(posten), dateien)

	for _, datei := range dateien {
		if bleiben[datei.pfad] {
			continue
		}

		if weg := a.dateiEntfernen(datei.pfad); weg != nil {
			// **Der neue Stand geht mit, nicht der leere** — §7 Regel 3 und
			// Befund 4 des Zweitblicks über Phase 4.13. Was vorher schon
			// gelöscht wurde, **ist** gelöscht; eine Antwort mit leerer Liste
			// ließe das Fenster seine alte Zählung behalten, und die stimmt
			// dann nicht mehr.
			antwort := a.Platz(ctx)
			antwort.Fehler = Uebersetze(ctx, domain.MitGrund(
				domain.ErrValidation, weg,
				"%q ließ sich nicht wegräumen — prüfe die Rechte am Ordner",
				filepath.Base(datei.pfad)))

			return antwort
		}
	}

	// **Weggeräumt ist weggeräumt** (§7, Regel 3; #1970): die Zählung danach
	// geht durch **alle** Posten, das Wegräumen nur durch seinen — scheitert
	// sie, steht ihr Fehler daneben und nicht in `fehler`.
	antwort := a.Platz(ctx)
	antwort.Nachladefehler, antwort.Fehler = antwort.Fehler, nil

	return antwort
}

// postendateien gibt die Dateien eines Postens — *Sonstiges* als Rest.
//
// **Der Rest läuft denselben Baum ab wie *Bilder* und *Pakete***, und was
// denen gehört, steht dann schon in `beansprucht`. Das kostet einen zweiten
// Durchgang durch `pakete/` — **gemessen am 2026-09-07**: 60.813 Dateien im
// Anwendungsordner, 0,28 s für einen vollen Durchgang.
//
// **Der Preis ist der kleinere.** Die Ersparnis wäre gewesen, die beiden
// Verzeichnisse zu überspringen — dafür hätte der Rest ihre Orte selbst
// bestimmen müssen, mit eigenen Fehlerzweigen, die **nach** den fünf anderen
// niemand mehr betreten kann. Angestrichen von der Abdeckung: sechs
// Anweisungen ohne Lauf, und keine davon herstellbar.
func postendateien(
	posten domain.Platzposten, bestandspfad string, beansprucht map[string]bool,
) ([]platzdatei, error) {
	if posten == domain.PostenSonstiges {
		return unterhalb(filepath.Dir(bestandspfad), beansprucht)
	}

	return dateienDesPostens(posten, bestandspfad)
}

// bleibende sagt, welche Dateien eines Postens stehen bleiben.
//
// **Je Posten ein Grund, und es ist nicht derselbe**: bei den Vorsicherungen
// bleibt der Weg zurück — welche Datei das ist, rechnet `wegzurueck.go` —,
// beim Protokoll die laufende Datei, denn in sie wird gerade geschrieben.
//
// **Kein `default` über alle Posten**: gerufen wird nur mit einem räumbaren,
// und davon gibt es zwei ([domain.Platzposten.Raeumbar]).
func (a *App) bleibende(
	ctx context.Context, posten domain.Platzposten, dateien []platzdatei,
) map[string]bool {
	if posten == domain.PostenProtokoll {
		return map[string]bool{laufendesProtokoll(a.datenbankpfad): true}
	}

	return a.bleibendeVorsicherungen(ctx, dateien)
}

// **Und dieselben zwei an allen vier Flächen** (§16).
//
// **Auch ohne Fenster**, und das ist Absicht: wer über die Kommandozeile
// merkt, dass die Platte voll ist, soll nicht erst ein Fenster öffnen müssen.

// Platz sagt es dem Fenster.
func (b *Bindungen) Platz(ctx context.Context) PlatzAntwort { return b.app(ctx).Platz(b.ruf(ctx)) }

// PlatzFreigeben räumt einen Posten weg.
func (b *Bindungen) PlatzFreigeben(ctx context.Context, posten string) PlatzAntwort {
	return b.app(ctx).PlatzFreigeben(b.ruf(ctx), posten)
}

// Platzflaeche ist die Anwendung selbst, für eine Fläche ohne Fenster.
//
// **Ein eigener Typ und kein `Nutzerflaeche`**: der Platz hat mit Menschen
// nichts zu tun, und eine Fläche, die beides trüge, wäre eine Sammelstelle.
type Platzflaeche struct{ app *App }

// PlatzflaecheUeber baut sie über einem Bestandsort.
//
// **Über den Ort und nicht über eine offene Verbindung**: gezählt werden
// **Dateien**, und die liegen auch dann da, wenn der Bestand nicht aufgeht —
// genau dann will jemand wissen, was Platz belegt.
func PlatzflaecheUeber(bestandsort string) *Platzflaeche {
	return &Platzflaeche{app: NeueApp(bestandsort)}
}

// MitDateientferner setzt die Naht zum Wegräumen — **für Tests**.
//
// **Weil der halbe Weg sonst nicht herstellbar ist** (Befund 4 des
// Zweitblicks über Phase 4.13): alle Vorsicherungen liegen im selben
// Verzeichnis, und ein unschreibbares Verzeichnis lässt schon das erste
// Löschen scheitern — nie das dritte.
func (f *Platzflaeche) MitDateientferner(entfernen func(string) error) {
	f.app.dateiEntfernen = entfernen
}

// Platz sagt es auch den Flächen ohne Fenster.
func (f *Platzflaeche) Platz(ctx context.Context) PlatzAntwort {
	return f.app.Platz(ctx)
}

// PlatzFreigeben räumt einen Posten weg.
func (f *Platzflaeche) PlatzFreigeben(ctx context.Context, posten string) PlatzAntwort {
	return f.app.PlatzFreigeben(ctx, posten)
}
