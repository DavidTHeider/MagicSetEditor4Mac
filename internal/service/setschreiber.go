package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Quellenleser gibt her, was ein Set zum Zurückschreiben braucht.
type Quellenleser interface {
	SetLeser

	// Quelle gibt die Datei, aus der das Set gelesen wurde — oder nil.
	Quelle(ctx context.Context, setID int64) (*domain.Quelle, error)

	// Schluesselwoerter gibt die des **Sets** — die des Spiels gehören keinem
	// (#1062).
	Schluesselwoerter(ctx context.Context, setID int64) ([]*domain.Keyword, error)

	// Setzeile gibt Name, Spiel, Stil und Herkunft **eines** Sets.
	//
	// **Für ein Set ohne Quelldatei** (F02,
	// #1746):
	// woher `game:` und `stylesheet:` kommen, wenn keine Datei sie trägt.
	Setzeile(ctx context.Context, setID int64) (domain.SetHead, error)

	// Setkopf gibt die fünf Formatangaben aus dem Dateikopf.
	//
	// **Nicht dasselbe wie [Quellenleser.Setzeile]**: dort die Zeile im
	// Bestand, hier `mse version` und was daneben steht.
	Setkopf(ctx context.Context, setID int64) (domain.Setkopf, error)

	// Stilwerte gibt die setweiten Stiloptionen, die **hier** gesetzt wurden
	// (F05, #2247).
	// Bis dahin las der Schreiber sie nicht, und keine kam in die Datei.
	Stilwerte(ctx context.Context, setID int64) ([]domain.Value, error)
}

// Speicherung ist der Ausgang eines Speicherns.
type Speicherung struct {
	Pfad string

	// Befunde sind die Stellen, an denen etwas **nicht** mitgekommen ist.
	//
	// **Befunde und Erfolg schließen einander nicht aus** (§8): die Hauptdatei
	// steht in der Datenbank und kommt heraus; was daneben lag — Bilder,
	// Symbole —, liegt am Herkunftsort und kann fehlen.
	Befunde []domain.Befund
}

// Setschreiber schreibt ein Set als `.mse-set` zurück.
type Setschreiber struct {
	wache

	leser Quellenleser
	orte  Ortsschreiber

	// bilder sind die Bilder, die jemand **in dieser Anwendung** eingesetzt
	// hat. `nil`, solange keine Ablage eingerichtet ist — dann kommt nur
	// heraus, was im Herkunftspaket liegt.
	bilder *Bildablage

	// aufloesung sagt, unter welchem Namen das Stilpaket liegt — `nil`, wo
	// keine Paketwurzeln bekannt sind; dann bleibt der Stil, wie er im
	// Bestand steht (`setkopfstil.go`).
	aufloesung Stilaufloesung
}

// MitStilaufloesung gibt dem Schreiber die Paketwurzeln über eine Auflösung
// (F05).
func (s *Setschreiber) MitStilaufloesung(aufloesung Stilaufloesung) *Setschreiber {
	s.aufloesung = aufloesung

	return s
}

// MitBildablage sagt dem Schreiber, wo eingesetzte Bilder liegen.
//
// **Nachgereicht und nicht im Konstruktor**: der Schreiber wird an drei Stellen
// gebaut, und zwei davon speichern nur Sets ohne eigene Bilder.
func (s *Setschreiber) MitBildablage(ablage *Bildablage) *Setschreiber {
	s.bilder = ablage

	return s
}

// bilderort gibt den Ort der eingesetzten Bilder — leer ohne Ablage.
func (s *Setschreiber) bilderort(setID int64) string {
	if s.bilder == nil {
		return ""
	}

	return s.bilder.Verzeichnis(setID)
}

// Ortsschreiber setzt den Arbeitsort eines Sets — die Naht zu `storage`.
//
// **Eine eigene Schnittstelle und nicht ein Zusatz am Leser**: der Leser liest
// (§2), und ein Schreibweg an ihm wäre der erste Riss in einer Trennung, die
// seit Phase 2 hält.
type Ortsschreiber interface {
	HerkunftSetzen(ctx context.Context, setID int64, ort string) error
}

// NeuerSetschreiber baut den Schreiber über einem Leser und einem Ortsschreiber.
func NeuerSetschreiber(leser Quellenleser, orte Ortsschreiber) *Setschreiber {
	return &Setschreiber{leser: leser, orte: orte}
}

// MitPruefer setzt den Riegel, der vor dem Schreiben fragt, ob dieser Aufruf
// es darf (0078).
func (s *Setschreiber) MitPruefer(pruefer Pruefer) *Setschreiber {
	s.pruefer = pruefer

	return s
}

// ErrOhneOrt sagt, dass ein Set aus keiner Datei kam und kein Ziel genannt
// ist — FIX-F02 (#1746). **Es bleibt ein [domain.ErrNotFound]**, damit jede
// Fläche es übersetzt wie bisher; das Fenster erkennt es und fragt, wohin
// (`app.Bindungen.SetSpeichern`).
var ErrOhneOrt = fmt.Errorf("%w: das Set hat keinen Ort, an den es zurückginge", domain.ErrNotFound)

// Schreibe gibt ein Set als Paketverzeichnis an einen Ort.
//
// # Die Byte-Zusage
//
// **Ein unverändertes Set kommt Byte für Byte zurück** (#48) — nicht, weil die
// Datei gut genug nachgebaut wird, sondern weil es dieselbe ist: der Bestand
// führt sie mit ([0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)).
//
// Geändert wird darin nur, was sich geändert hat. Die Zeile eines geänderten
// Werts wird neu gesetzt; jede andere Zeile bleibt, wie sie dastand — samt
// Einrückung, Abständen, Kommentaren, Leerzeilen und dem Zeilenende, das sie
// hatte.
func (s *Setschreiber) Schreibe(
	ctx context.Context, setID int64, ziel string,
) (*Speicherung, error) {
	// **Das Nein kommt vor jeder anderen Antwort** — wer nicht schreiben
	// darf, bekommt auch keine Auskunft darüber, was er geschrieben hätte.
	if err := s.darfSchreiben(ctx); err != nil {
		return nil, err
	}

	quelle, baum, err := s.baumVon(ctx, setID)
	if err != nil {
		return nil, err
	}

	// **Ohne Ziel geht es an die Herkunft** — das ist ⌘S. Der Ort steht in der
	// Quelle: die Oberfläche soll ihn nicht führen müssen, und ein Ort, den
	// zwei Seiten führen, läuft auseinander.
	if ziel == "" {
		ziel = quelle.Herkunft
	}

	if ziel == "" {
		return nil, fmt.Errorf("%w (Set %d)", ErrOhneOrt, setID)
	}

	if befund := keinArchiv(ziel); befund != nil {
		return nil, befund
	}

	if befund := s.fuellen(ctx, setID, baum); befund != nil {
		return nil, befund
	}

	befunde, err := s.ablegen(quelle, baum, setID, ziel)
	if err != nil {
		return nil, err
	}

	// **Und der Arbeitsort wechselt mit** (#226).
	//
	// Bis hierher blieb er der alte: nach „Sichern unter" ging das folgende ⌘S
	// wieder gegen die alte Herkunft. Bei einem Archiv wurde es wieder
	// abgewiesen — der Ausweg, den die Meldung nennt, führte nicht heraus. Bei
	// einem Verzeichnis fiel es nur nicht auf, weil das alte Ziel schreibbar
	// bleibt: dann schrieb ⌘S still an eine Stelle, die niemand mehr meint.
	//
	// **Ein Feld und keine Transaktion**, und das ist der Grund: das Speichern
	// ändert die **Quelle nicht**. Sie steht seit dem Öffnen im Bestand
	// ([0011](../../docs/entscheidungen/0011-die-quelldatei-wird-mitgefuehrt.md)),
	// und was hier hinausgeht, entsteht aus ihr. Auseinanderlaufen können
	// Herkunft und Quelle deshalb nicht — es gibt nur eine Angabe, die sich
	// ändert.
	//
	// **Ein Fehler dabei nimmt das Geschriebene nicht zurück** (§7, Regel 3):
	// die Datei liegt da. Er wird gemeldet, denn der Bestand nennt danach einen
	// Ort, an dem das Set nicht mehr gepflegt wird.
	if ziel != quelle.Herkunft {
		if err := s.orte.HerkunftSetzen(ctx, setID, ziel); err != nil {
			return &Speicherung{Pfad: ziel, Befunde: befunde},
				fmt.Errorf("den Arbeitsort auf %q setzen: %w", ziel, err)
		}
	}

	return &Speicherung{Pfad: ziel, Befunde: befunde}, nil
}

// baumVon liest die Quelle eines Sets und zerteilt sie — oder baut ein
// Gerüst, wo es keine gibt.
//
// **Ohne Quelle wird zusammengesetzt** (F02,
// #1746).
// Die Byte-Zusage steht hier nicht auf dem Spiel — es gibt keine
// Ursprungsdatei, mit der sich vergleichen ließe —, und [0011] sagt seit dem
// 2026-08-15, dass in diesem Fall zusammengesetzt wird.
func (s *Setschreiber) baumVon(ctx context.Context, setID int64) (*domain.Quelle, *msepaket.Baum, error) {
	quelle, err := s.leser.Quelle(ctx, setID)
	if err != nil {
		return nil, nil, fmt.Errorf("die Quelle lesen: %w", err)
	}

	if quelle == nil {
		baum, geruestfehler := s.geruest(ctx, setID)
		if geruestfehler != nil {
			return nil, nil, geruestfehler
		}

		return &domain.Quelle{Datei: "set"}, baum, nil
	}

	baum, err := msepaket.Zerteile(quelle.Datei, quelle.Inhalt)
	if err != nil {
		return nil, nil, fmt.Errorf("die Quelle zerteilen: %w", err)
	}

	return quelle, baum, nil
}

// fuellen gleicht den Baum an den Bestand an — Karten, Schlüsselwörter und
// die Angaben des Sets.
func (s *Setschreiber) fuellen(ctx context.Context, setID int64, baum *msepaket.Baum) error {
	karten, err := s.leser.Karten(ctx, setID)
	if err != nil {
		return fmt.Errorf("die Karten lesen: %w", err)
	}

	angleichen(baum, karten)

	// **Und die Schlüsselwörter** (#1062). Sie standen bis dahin als Rohblock
	// in der Datei und kamen unverändert heraus; jetzt führt sie der Bestand,
	// und was dort steht, geht hinaus.
	woerter, err := s.leser.Schluesselwoerter(ctx, setID)
	if err != nil {
		return fmt.Errorf("die Schlüsselwörter lesen: %w", err)
	}

	schluesselwoerterAngleichen(baum, woerter)

	// **Und die Angaben des Sets** (#890). Sie standen bis dahin nur im
	// Bestand: `SetwertSetzen` schrieb hinein, und beim Sichern blieb in der
	// Datei stehen, was vorher dort stand — Titel, Urheberrecht, Setkürzel und
	// das Symbol.
	werte, err := s.leser.Setwerte(ctx, setID)
	if err != nil {
		return fmt.Errorf("die Setangaben lesen: %w", err)
	}

	setwerteAngleichen(baum, werte)

	// **Und die setweiten Stiloptionen** (F05, #2247) — in den Block
	// `styling:`, unter das Stilpaket des Sets.
	stilwerte, err := s.leser.Stilwerte(ctx, setID)
	if err != nil {
		return fmt.Errorf("die Stiloptionen lesen: %w", err)
	}

	// **Und Spiel und Stil im Kopf** (F05; Befund 2 des Zweitblicks über
	// Phase G, zweiter Teil): ein gewechselter Stil kam bis dahin nicht in
	// die Datei, und der Name stand nicht in der Kurzform, die MSE2 liest
	// (`setkopfstil.go`).
	kopf, kopffehler := s.leser.Setzeile(ctx, setID)
	if kopffehler != nil {
		return fmt.Errorf("Spiel und Stil lesen: %w", kopffehler)
	}

	paket, kurz := stilnamen(kopf, s.aufloesung)
	kopfAngleichen(baum, kopf, kurz)
	stilblockAngleichen(baum, kopf, stilwerte, paket)

	return nil
}

// ablegen schreibt Beiwerk und Hauptdatei in ein Verzeichnis.
//
// **Das Beiwerk zuerst, die Hauptdatei zuletzt.** Wer umgekehrt vorginge,
// überschriebe beim Speichern über die Herkunft die eben geschriebene
// Hauptdatei mit der alten aus demselben Verzeichnis.
func (s *Setschreiber) ablegen(
	quelle *domain.Quelle, baum *msepaket.Baum, setID int64, ziel string,
) ([]domain.Befund, error) {
	if err := os.MkdirAll(ziel, 0o750); err != nil {
		return nil, fmt.Errorf("das Ziel anlegen: %w", err)
	}

	befunde := Beiwerk(quelle.Herkunft, s.bilderort(setID), quelle.Datei, ziel)

	// **Über eine Nachbardatei** ([schreibeAtomar], Werkstück 7 von Phase
	// 4.13). Hier stand `os.WriteFile`, und das öffnet mit `O_TRUNC`: ein ⌘Q
	// oder ein Absturz zwischen Kürzen und Fertigschreiben ließ ein halbes
	// `.mse-set` unter dem richtigen Namen zurück — die Arbeit von Wochen,
	// und im Finder sieht es aus wie ein ganzes.
	if err := schreibeAtomar(filepath.Join(ziel, quelle.Datei), msepaket.Schreibe(baum)); err != nil {
		return nil, fmt.Errorf("das Set schreiben: %w", err)
	}

	return befunde, nil
}

// geruest baut den Baum eines Sets, das aus keiner Datei stammt (F02).
//
// **Zwei Lesevorgänge und nicht einer**: die Bestandszeile sagt, welches Spiel
// und welcher Stil gemeint sind, der Formatkopf, nach welchem Stand die Datei
// aufgebaut ist. Beides steht in derselben Tabellenzeile — und zwei Abfragen
// sind hier billiger als eine Schnittstelle, die vier Angaben mit fünf
// vermischt.
func (s *Setschreiber) geruest(
	ctx context.Context, setID int64,
) (*msepaket.Baum, error) {
	zeile, err := s.leser.Setzeile(ctx, setID)
	if err != nil {
		return nil, fmt.Errorf("die Bestandszeile lesen: %w", err)
	}

	format, err := s.leser.Setkopf(ctx, setID)
	if err != nil {
		return nil, fmt.Errorf("den Setkopf lesen: %w", err)
	}

	return setgeruest(zeile, format)
}
