package service

// **Der Aufbau eines Retrievalindex** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Warum der neue Index neben dem alten entsteht
//
// **Weil ein Leser eine Generation zu Ende lesen darf.** Das Issue: „Neuer
// Index entsteht separat und wird erst nach Prüfung atomar aktiviert; Leser
// halten konsistente Generation." Wer in die aktive Generation schriebe,
// gäbe demselben Aufrufer zwei verschiedene Antworten auf dieselbe Frage.
//
// # Und warum ein Abbruch verwirft statt anzuhalten
//
// **Weil eine halbe Generation kein Index ist.** Sie enthält, was bis zum
// Abbruch geschrieben wurde, und sieht wie ein vollständiger Stand aus. Das
// Issue nennt den Fall dreimal — „Crash/Disk-full/Abbruch vor und nach
// Umschaltung erzeugen nie gemischte Vektoren".
//
// # Eine Schreibinstanz, und der zweite wartet nicht
//
// **Warten hieße, zweimal dasselbe zu bauen.** Wer aufbaut, während ein
// Aufbau läuft, will einen aktuellen Index; er bekäme ihn nach dem doppelten
// der Zeit, und der erste wäre in dem Moment veraltet. Deshalb eine
// Abweisung mit [domain.ErrConflict] und eine Auskunft über den laufenden
// Stand.

import (
	"context"
	"sync"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// StapelHoechstens ist, wie viele Texte auf einmal eingebettet werden.
//
// **Zweiunddreißig, dieselbe Zahl wie in `internal/ollama`** — und dieselbe
// Begründung: ein Stapel wird als ein Auftrag geladen. Sie steht hier
// **noch einmal**, weil der Dienst keine Laufzeit kennen darf (§2); ein
// Einbetter mit engerer Grenze weist ab, und das ist ein Fehler und kein
// stilles Kürzen.
const StapelHoechstens = 32

// Retrievalschreiber ist, was der Aufbau vom Bestand braucht.
//
// **Getrennt von [Retrievalablage]**, weil Suchen und Bauen zwei Rechte
// sind: wer sucht, schreibt nicht.
type Retrievalschreiber interface {
	Anlegen(ctx context.Context, manifest domain.Indexmanifest) (int64, error)
	Schreiben(ctx context.Context, generation int64,
		dokument domain.Retrievaldokument, vektor []float32) error
	Aktivieren(ctx context.Context, generation int64) error
	Verwerfen(ctx context.Context, generation int64) error

	// Uebernehmen kopiert die Abschnitte der **aktiven** Generation in eine
	// neue — bis auf die genannten.
	//
	// **Dafür gibt es den Deltaimport**: was sich nicht geändert hat, wird
	// kopiert statt neu eingebettet. Der teure Teil ist das Modell, nicht
	// der `INSERT`.
	Uebernehmen(ctx context.Context, nach int64, ausser []string) (int, error)
}

// Prozesssperre ist die Absprache über Prozessgrenzen hinweg.
//
// # Warum der Riegel im Prozess nicht reicht
//
// **Weil dieses Programm vier Flächen hat und drei davon eigene Prozesse
// sind.** Ein `mse indexaufbau` weiß nichts vom Fenster, und beide schreiben
// in denselben Bestand. Das Issue sagt es ausdrücklich: „Eine aktive
// Schreibinstanz je Index mit dokumentierter Konkurrenz-/Crashstrategie,
// **auch bei mehreren Prozessen**."
//
// **Sie wird eingesetzt und nicht erzeugt** (§16): welche Datei die Absprache
// trägt, weiß `internal/app`, und ein Test wartet nie auf eine Sperre, die
// ein anderer Lauf hält.
type Prozesssperre interface {
	// Nehmen versucht die alleinige Sperre, **ohne zu warten**.
	//
	// **Ohne Warten, und das ist dieselbe Entscheidung wie im Prozess**: wer
	// wartet, baut nach doppelter Zeit denselben Stand ein zweites Mal.
	Nehmen() (freigeben func(), genommen bool, fehler error)
}

// Korpusquelle liefert die Dokumente eines Aufbaus.
//
// **Stapelweise und nicht am Stück.** Ein vollständiger Oracle-Korpus sind
// Zehntausende Abschnitte; wer sie als eine Liste verlangte, verlangte sie
// alle zugleich im Speicher — und das Issue nennt „RAM-/Index-/Temp-Diskbudget"
// ausdrücklich.
type Korpusquelle interface {
	// Staende nennt, in welchem Stand jede Quelle ist.
	Staende(ctx context.Context) ([]domain.Korpusstand, error)

	// Stapel gibt den nächsten — eine leere Rückgabe beendet den Lauf.
	Stapel(ctx context.Context) ([]domain.Retrievaldokument, error)
}

// Aufbaustoff beschafft, was ein Aufbau braucht.
//
// # Warum als Funktion und nicht als zwei Parameter
//
// **Weil beides Zeit kostet und hinter den Riegel gehört.** Ein Einbetter
// entsteht aus einer Messung am echten Modell, eine Korpusquelle aus einer
// geöffneten Datei; wer beides **vor** dem Riegel beschaffte, ließe einen
// zweiten Prozess erst messen und dann abweisen.
//
// **`schliessen` gibt die Quelle zurück** — auch wenn der Aufbau scheitert.
type Aufbaustoff func(ctx context.Context) (
	einbetter Einbetter, quelle Korpusquelle, schliessen func(), fehler error)

// Aufbauwunsch nennt die Fassungen, unter denen gebaut wird.
type Aufbauwunsch struct {
	Schemafassung  int
	Zerteilfassung string
	Indexfassung   int

	// Rerankerabdruck ist leer, wo keiner im Spiel ist.
	Rerankerabdruck string

	// Fortschreiben sagt, dass die aktive Generation die Grundlage ist.
	//
	// **Dann kommt aus der Quelle nur, was sich geändert hat**, und der Rest
	// wird übernommen. Was die Quelle liefert und danach **nicht** geschrieben
	// wird, gibt es nicht — so wird gelöscht.
	Fortschreiben bool
}

// Aufbaustand ist, was ein Aufbau über sich sagt.
//
// **Er ist auch nach dem Lauf noch da.** Ein Fehlerzustand, der mit dem Lauf
// verschwände, wäre für niemanden lesbar — das Issue verlangt
// „Fortschritt/Abbruch/Fehlerzustand".
type Aufbaustand struct {
	Laeuft      bool
	Generation  int64
	Geschrieben int

	// Fehler ist der Satz des letzten Fehlschlags — leer nach einem
	// geglückten Lauf.
	Fehler string

	Begonnen time.Time
	Beendet  time.Time
}

// Aufbaubericht ist das Ergebnis eines geglückten Aufbaus.
type Aufbaubericht struct {
	Generation  int64
	Geschrieben int
	Manifest    domain.Indexmanifest
	Dauer       time.Duration
}

// Indexaufbau baut Generationen des Retrievalindex.
//
// **Der Einbetter gehört zum Lauf und nicht zum Aufbau.** Wer ihn festhielte,
// hätte je Modell einen Aufbau — und damit je Modell einen eigenen Riegel,
// während es nur **einen** Index gibt, in den geschrieben wird.
type Indexaufbau struct {
	schreiber Retrievalschreiber
	haushalt  *Haushalt

	// sperre hält die Absprache über Prozessgrenzen — `nil` heißt: keine,
	// und dann gilt nur der Riegel in diesem Prozess.
	sperre Prozesssperre

	mu    sync.Mutex
	stand Aufbaustand

	// abbrechen hält den laufenden Aufbau an — `nil`, wenn keiner läuft.
	abbrechen context.CancelFunc
}

// NeuerIndexaufbau baut ihn.
func NeuerIndexaufbau(
	schreiber Retrievalschreiber, haushalt *Haushalt,
) *Indexaufbau {
	return &Indexaufbau{schreiber: schreiber, haushalt: haushalt}
}

// MitProzesssperre setzt die Absprache über Prozessgrenzen hinweg.
func (a *Indexaufbau) MitProzesssperre(sperre Prozesssperre) *Indexaufbau {
	a.sperre = sperre

	return a
}

// Aufbauen baut eine neue Generation und schaltet auf sie um.
func (a *Indexaufbau) Aufbauen(
	ctx context.Context, wunsch Aufbauwunsch, beschaffen Aufbaustoff,
) (Aufbaubericht, error) {
	ctx, abbrechen := context.WithCancel(ctx)
	defer abbrechen()

	if fehler := a.beginnen(abbrechen); fehler != nil {
		return Aufbaubericht{}, fehler
	}

	// **Die Sperre über die Prozesse kommt vor dem Platz.** Andersherum
	// hielte ein abgewiesener zweiter Prozess kurz einen Platz, den niemand
	// nutzt — dieselbe Überlegung wie beim Riegel darüber.
	sperrefrei, fehler := a.prozesssperre()
	if fehler != nil {
		a.beenden(0, 0, fehler)

		return Aufbaubericht{}, fehler
	}

	defer sperrefrei()

	freigeben, fehler := a.haushalt.Belegen(ctx, domain.RessourceIndexbau)
	if fehler != nil {
		a.beenden(0, 0, fehler)

		return Aufbaubericht{}, fehler
	}

	defer freigeben()

	// **Erst jetzt wird gemessen und geöffnet.** Vorher hieße: ein zweiter
	// Prozess misst ein Modell, um danach abgewiesen zu werden.
	einbetter, quelle, schliessen, fehler := beschaffen(ctx)

	// **Der Schließer wird **vor** der Fehlerprüfung vorgemerkt**, und das
	// ist keine Feinheit: [Aufbaustoff] sagt zu, immer einen zu geben, und
	// wer ihn erst danach vormerkte, hätte einen Weg, den kein Lauf betreten
	// kann — er sähe aus wie eine geprüfte Zusage (§10).
	defer schliessen()

	if fehler != nil {
		a.beenden(0, 0, fehler)

		return Aufbaubericht{}, fehler
	}

	if einbetter == nil {
		fehler = domain.MitGrund(domain.ErrValidation, nil,
			"ohne eingerichtetes Einbettungsmodell entsteht kein hybrider "+
				"Index — die Volltextsuche allein gibt es schon")
		a.beenden(0, 0, fehler)

		return Aufbaubericht{}, fehler
	}

	bericht, fehler := a.lauf(ctx, einbetter, wunsch, quelle)
	a.beenden(bericht.Generation, bericht.Geschrieben, fehler)

	if fehler != nil {
		return Aufbaubericht{}, fehler
	}

	return bericht, nil
}

// lauf ist der eigentliche Aufbau.
func (a *Indexaufbau) lauf(
	ctx context.Context, einbetter Einbetter,
	wunsch Aufbauwunsch, quelle Korpusquelle,
) (Aufbaubericht, error) {
	begonnen := time.Now()

	manifest, fehler := manifestFuer(ctx, einbetter, wunsch, quelle)
	if fehler != nil {
		return Aufbaubericht{}, fehler
	}

	generation, fehler := a.schreiber.Anlegen(ctx, manifest)
	if fehler != nil {
		return Aufbaubericht{}, fehler
	}

	geschrieben, angefasst, fehler := a.fuellen(ctx, einbetter, generation, quelle)

	if fehler == nil && wunsch.Fortschreiben {
		var uebernommen int

		uebernommen, fehler = a.schreiber.Uebernehmen(ctx, generation, angefasst)
		geschrieben += uebernommen
	}

	if fehler == nil {
		fehler = a.schreiber.Aktivieren(ctx, generation)
	}

	if fehler != nil {
		// **Verworfen wird ohne den abgebrochenen Kontext.** Sonst wäre
		// gerade der Abbruch der Fall, in dem die halbe Generation stehen
		// bliebe — und das ist der Fall, für den es die Zeile gibt.
		_ = a.schreiber.Verwerfen(context.WithoutCancel(ctx), generation) //nolint:errcheck // siehe unten

		return Aufbaubericht{Generation: generation, Geschrieben: geschrieben},
			fehler
	}

	return Aufbaubericht{
		Generation: generation, Geschrieben: geschrieben,
		Manifest: manifest, Dauer: time.Since(begonnen),
	}, nil
}

// fuellen schreibt Stapel für Stapel.
//
// **`angefasst` nennt jede Kennung, die aus der Quelle kam** — auch die, deren
// Dokument gar nicht geschrieben wurde, weil der Stapel vorher scheiterte.
// Für den Deltaimport ist das die Liste, die **nicht** übernommen wird: was
// hier durchkam und nicht geschrieben wurde, ist entfernt.
func (a *Indexaufbau) fuellen(
	ctx context.Context, einbetter Einbetter,
	generation int64, quelle Korpusquelle,
) (geschrieben int, angefasst []string, fehler error) {
	angefasst = make([]string, 0)

	for {
		if fehler := ctx.Err(); fehler != nil {
			return geschrieben, angefasst, fehler
		}

		stapel, fehler := quelle.Stapel(ctx)
		if fehler != nil {
			return geschrieben, angefasst, fehler
		}

		if len(stapel) == 0 {
			return geschrieben, angefasst, nil
		}

		for _, dokument := range stapel {
			angefasst = append(angefasst, dokument.Kennung)
		}

		for teil := range teilstapel(stapel) {
			gezaehlt, fehler := a.schreibeTeil(ctx, einbetter, generation, teil)
			geschrieben += gezaehlt
			a.fortschritt(generation, geschrieben)

			if fehler != nil {
				return geschrieben, angefasst, fehler
			}
		}
	}
}

// schreibeTeil bettet einen Teilstapel ein und schreibt ihn.
func (a *Indexaufbau) schreibeTeil(
	ctx context.Context, einbetter Einbetter,
	generation int64, teil []domain.Retrievaldokument,
) (int, error) {
	vertrag := einbetter.Raum().Vertrag

	texte := make([]string, 0, len(teil))
	for _, dokument := range teil {
		// **Als Dokument und nicht als Frage.** Das Issue trennt sie
		// fachlich; wer beide gleich einbettete, legte die Fragen neben den
		// Korpus statt hinein.
		texte = append(texte, vertrag.FuerDokument(dokument.Titel+"\n"+dokument.Text))
	}

	vektoren, fehler := einbetter.Einbetten(ctx, texte)
	if fehler != nil {
		return 0, fehler
	}

	if len(vektoren) != len(teil) {
		return 0, domain.MitGrund(domain.ErrValidation, nil,
			"auf %d Abschnitte kamen %d Vektoren zurück — welcher zu welchem "+
				"gehört, entscheidet dieses Programm nicht",
			len(teil), len(vektoren))
	}

	for stelle, dokument := range teil {
		if fehler := a.schreiber.Schreiben(ctx, generation,
			dokument, vektoren[stelle]); fehler != nil {
			return stelle, fehler
		}
	}

	return len(teil), nil
}

// teilstapel zerlegt einen Stapel in Teile von höchstens [StapelHoechstens].
func teilstapel(stapel []domain.Retrievaldokument) func(
	func([]domain.Retrievaldokument) bool,
) {
	return func(weiter func([]domain.Retrievaldokument) bool) {
		for von := 0; von < len(stapel); von += StapelHoechstens {
			if !weiter(stapel[von:min(von+StapelHoechstens, len(stapel))]) {
				return
			}
		}
	}
}
