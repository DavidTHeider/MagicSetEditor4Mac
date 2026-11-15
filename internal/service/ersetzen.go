package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Suchen und Ersetzen über ein Set** — Werkstück 2 von Phase 3.7
// (#1168).
//
// # Was MSE2 hier hat und was nicht
//
// **Die Übereinstimmung ist portiert** — sie steht in [domain.Ersetzungsfrage],
// samt Fundstellen. **Der Lauf über ein Set ist es nicht**:
//
//	bool CardsPanel::doReplaceAll(wxFindReplaceData& what) {
//	  return false; // TODO
//	}
//
// `src/gui/set/cards_panel.cpp`, Zeilen 452–455, Fassung f9d9356 vom
// 2021-07-06. Gesucht wurde am 2026-08-27 über `src/` nach `ReplaceAll`,
// `replace_all` und `wxFR_REPLACEALL`: die Schaltfläche steht im Dialog, der
// Weg dahinter ist leer.
//
// # Zwei Wege, und der erste schreibt nicht
//
// [Ersetzer.Vorschau] sagt, **was sich ändern würde**; [Ersetzer.Ausfuehren]
// tut es. Das Issue verlangt beides ausdrücklich — eine Ersetzung über
// dreihundert Karten ist nicht zu überblicken, und §16 Regel 6 sagt für fremde
// Daten dasselbe: nichts wird still übernommen.
//
// **Die zweite Zusage ist die Rücknahme in einem Zug**: der Schreiber legt
// alle Werte in **eine** Gruppe (`storage.WerteSetzen`), und ⌘Z nimmt sie
// gemeinsam zurück.

// Ersetzungskarten gibt die Karten eines Sets samt ihren Werten.
type Ersetzungskarten interface {
	Karten(ctx context.Context, setID int64) ([]*domain.CardRow, error)
}

// Wertschreiber schreibt viele Feldwerte als **einen** Schritt.
type Wertschreiber interface {
	WerteSetzen(ctx context.Context, wechsel []domain.Feldwechsel) error
}

// Ersetzer sucht und ersetzt über ein Set.
type Ersetzer struct {
	karten    Ersetzungskarten
	felder    Kartenfeldquelle
	schreiber Wertschreiber
}

func NeuerErsetzer(
	karten Ersetzungskarten, felder Kartenfeldquelle, schreiber Wertschreiber,
) *Ersetzer {
	return &Ersetzer{karten: karten, felder: felder, schreiber: schreiber}
}

// Vorschau sagt, was sich ändern würde — **und schreibt nichts**.
func (e *Ersetzer) Vorschau(
	ctx context.Context, setID int64, frage domain.Ersetzungsfrage,
) (domain.Ersetzungsvorschau, error) {
	schau, _, fehler := e.sammle(ctx, setID, frage)

	return schau, fehler
}

// Ausfuehren ersetzt und gibt dieselbe Übersicht zurück.
//
// **Dieselbe Übersicht**, damit der Aufrufer nicht raten muss, ob die Vorschau
// noch galt: was hier zurückkommt, ist das, was **geschrieben wurde**.
func (e *Ersetzer) Ausfuehren(
	ctx context.Context, setID int64, frage domain.Ersetzungsfrage,
) (domain.Ersetzungsvorschau, error) {
	schau, wechsel, fehler := e.sammle(ctx, setID, frage)
	if fehler != nil {
		return domain.Ersetzungsvorschau{}, fehler
	}

	// **Die Vorschau muss noch gelten** (#1241). Zwischen ihr und diesem Zug
	// kann eine Karte durch den Editor, eine Rücknahme, den Dienst oder den
	// unabhängigen MCP-Prozess geändert worden sein — dann bestätigte der
	// Knopf **andere Treffer als die angezeigten**.
	if frage.Stand != "" && frage.Stand != schau.Stand {
		return domain.Ersetzungsvorschau{}, fmt.Errorf(
			"%w: die Vorschau gilt nicht mehr — seit sie entstand, hat sich "+
				"an den betroffenen Karten etwas geändert", domain.ErrConflict)
	}

	// **Ohne Treffer wird nicht geschrieben.** Ein leerer Zug legte sonst
	// einen Verlaufsschritt an, der nichts tut — und ⌘Z danach sähe aus, als
	// hätte es nicht gewirkt.
	if len(wechsel) == 0 {
		return schau, nil
	}

	if err := e.schreiber.WerteSetzen(ctx, wechsel); err != nil {
		return domain.Ersetzungsvorschau{}, err
	}

	return schau, nil
}

// sammle durchsucht das Set einmal und gibt Übersicht und Änderungen.
//
// **Eine Stelle für beide Wege**, damit die Vorschau nicht etwas anderes
// verspricht, als die Ausführung tut — zwei Durchläufe liefen auseinander,
// und niemand sähe es.
func (e *Ersetzer) sammle(
	ctx context.Context, setID int64, frage domain.Ersetzungsfrage,
) (domain.Ersetzungsvorschau, []domain.Feldwechsel, error) {
	if befunde := frage.Validate(); len(befunde) > 0 {
		return domain.Ersetzungsvorschau{}, nil, befunde[0]
	}

	karten, err := e.karten.Karten(ctx, setID)
	if err != nil {
		return domain.Ersetzungsvorschau{}, nil, fmt.Errorf("die Karten lesen: %w", err)
	}

	textfelder, ohneVorlage := e.textfelder(ctx, setID)

	var (
		schau   domain.Ersetzungsvorschau
		wechsel []domain.Feldwechsel
	)

	// **Der Fingerabdruck wächst mit** (#1241) — hier, wo der **vorherige**
	// Wert noch in der Hand ist. [domain.Feldwechsel] trägt nur den neuen, und
	// [Ersetzungsvorschau.Stellen] ist abgeschnitten.
	stand := sha256.New()

	schau.OhneVorlage = ohneVorlage

	for _, karte := range karten {
		vorher := len(wechsel)

		for _, wert := range karte.Values {
			if !durchsuchbar(textfelder, frage, wert.Field) {
				continue
			}

			stellen := frage.Stellen(wert.Text)
			if len(stellen) == 0 {
				continue
			}

			neu := frage.Ersetze(wert.Text)

			wechsel = append(wechsel, domain.Feldwechsel{
				KarteID: karte.ID, Feld: wert.Field, Wert: neu, Vorher: wert.Text,
			})

			//nolint:errcheck // ein [hash.Hash] gibt nie einen Fehler (Vertrag von io.Writer dort)
			_, _ = fmt.Fprintf(stand, "%d\x00%s\x00%s\x00",
				karte.ID, wert.Field, wert.Text)

			schau.Treffer += len(stellen)

			if len(schau.Stellen) < domain.GrenzeErsetzungsvorschau {
				schau.Stellen = append(schau.Stellen, domain.Ersetzungsstelle{
					KarteID: karte.ID, Nummer: karte.Number, Feld: wert.Field,
					Vorher:  domain.OhneAuszeichnung(wert.Text),
					Nachher: domain.OhneAuszeichnung(neu),
					Treffer: len(stellen),
				})
			} else {
				schau.Weitere = true
			}
		}

		if len(wechsel) > vorher {
			schau.Karten++
		}
	}

	schau.Stand = vorschaustand(stand, len(wechsel))

	return schau, wechsel, nil
}

// textfelder gibt die Namen der Felder, in denen gesucht werden darf.
//
// **`nil` heißt: alle**, und die zweite Rückgabe sagt, ob das eine Notlage ist.
//
// # Warum eine unlesbare Vorlage kein Fehler ist
//
// Dieselbe Duldung wie beim Feldschreiben ([Feldweg.FelderVonAussen]) und aus
// demselben Grund: ein Set aus MTG JSON bliebe sonst unbearbeitbar, solange
// sein Spielpaket fehlt. **Dann gilt jedes Feld als Text** — die einzige
// Annahme, die man ohne Vorlage treffen kann.
//
// **Und sie wird gesagt, nicht verschluckt** (§8). Ohne Vorlage kann niemand
// wissen, welches Feld ein `choice:` ist; die **Vorschau** ist es, die das
// abfängt — sie nennt jedes Feld beim Namen, bevor geschrieben wird. Der
// Hinweis daneben sagt, warum sie diesmal genauer anzusehen ist.
//
// **Hier stand bis zum 2026-08-27 ein `return nil, err`**, und die Duldung
// darüber galt nur einer **leeren** Liste. Ein fehlendes Spielpaket gibt aber
// `ErrNotFound` — der Fall, für den der Satz geschrieben war, war genau der,
// den er nicht erreichte. Gefunden von `TestDieErsetzungsflaechennahtTraegt…`.
func (e *Ersetzer) textfelder(ctx context.Context, setID int64) (
	erlaubt map[string]bool, ohneVorlage bool,
) {
	felder, err := e.felder.Kartenfelder(ctx, setID)
	if err != nil || len(felder) == 0 {
		return nil, true
	}

	erlaubt = make(map[string]bool, len(felder))

	for _, feld := range felder {
		// **Nur Text, und `KindUnknown` zählt dazu.**
		//
		//	virtual bool search(FindInfo& find, bool from_start) { return false; }
		//
		// `src/gui/value/editor.hpp`, Zeile 115, Fassung f9d9356 vom
		// 2021-07-06 — die Grundfassung findet nichts, und nur
		// `TextValueEditor` überschreibt sie. Eine Ersetzung in einem
		// `choice:`-Feld schriebe eine Wahl, die es nicht gibt.
		//
		// **`KindUnknown` ist eine Art, die wir nicht unterscheiden**, kein
		// Feld ohne Text — dieselbe Lesart wie in `domain.FieldKind`.
		erlaubt[feld.Name] = feld.Kind == domain.KindText || feld.Kind == domain.KindUnknown
	}

	return erlaubt, false
}

// durchsuchbar sagt, ob ein Feldwert an die Reihe kommt.
func durchsuchbar(textfelder map[string]bool, frage domain.Ersetzungsfrage, feld string) bool {
	if frage.Feld != "" && frage.Feld != feld {
		return false
	}

	return textfelder == nil || textfelder[feld]
}

// vorschaustand gibt den Fingerabdruck der betroffenen Stellen —
// #1241.
//
// # Was er deckt
//
// Karte, Feld und den **vorherigen** Wert jeder Stelle. Damit merkt er sowohl
// eine geänderte Karte als auch eine, die dazugekommen ist und einen Treffer
// trägt.
//
// # Über alle Treffer und nicht über die Anzeige
//
// [Ersetzungsvorschau.Stellen] ist **abgeschnitten** — sie zeigt höchstens
// eine Grenze voll. Ein Stand darüber ließe alles durch, was jenseits der
// Grenze liegt, und genau dort stehen die Treffer, die niemand gesehen hat.
//
// # Warum SHA-256 und keine Verkettung
//
// Der Stand geht über die Wails-Naht ans Fenster und wieder zurück. Eine
// Verkettung aus Kartentexten trüge Inhalte durch die Oberfläche, die dort
// nichts zu suchen haben — und sie wäre bei dreitausend Treffern länger als
// die Antwort.
//
// # Ohne Wechsel kein Stand
//
// Eine Vorschau ohne Treffer bindet nichts: es gibt nichts zu bestätigen, und
// `Ausfuehren` schreibt ohnehin nicht.
func vorschaustand(summe hash.Hash, wechsel int) string {
	if wechsel == 0 {
		return ""
	}

	return hex.EncodeToString(summe.Sum(nil))
}
