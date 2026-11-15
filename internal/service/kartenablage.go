package service

// **Karten kopieren und einfügen** — Werkstück 1 von Phase 3.7
// (#1168).
//
// # ported, und die Fundstelle ist genau
//
//	wxDataFormat CardsDataObject::format = _("application/x-mse-cards");
//
//	struct WrappedCards {
//	  Game*         expected_game;
//	  String        game_name;
//	  vector<CardP> cards;
//	};
//
//	IMPLEMENT_REFLECTION(WrappedCards) {
//	  REFLECT(game_name);
//	  if (game_name == expected_game->name()) {
//	    WITH_DYNAMIC_ARG(game_for_reading, expected_game);
//	    REFLECT(cards);
//	  }
//	}
//
// `src/data/format/clipboard.cpp`, Zeilen 42–60, Fassung f9d9356 vom
// 2021-07-06. **Drei Dinge stehen darin:**
//
//   - Das **Spiel steht vor den Karten** — `game_name` wird zuerst reflektiert.
//   - Die Karten sind **MSEs eigenes Textformat**, kein eigenes: `REFLECT(cards)`
//     schreibt sie über denselben `Writer`, der eine `.mse-set` schreibt, und
//     ein `vector<T>` bekommt dabei den **Einzahlschlüssel** (`singular_form`,
//     `src/util/string.cpp`, Zeilen 153–160) — also wiederholte `card:`-Blöcke.
//   - **Stimmt das Spiel nicht, werden die Karten gar nicht erst gelesen.**
//     Die Bedingung steht **um** `REFLECT(cards)` herum, nicht dahinter.
//
// # adapted: der Inhaltstyp ist Text
//
// MSE2 legt die Zeichenkette unter einem **eigenen** Datenformat ab
// (`SetFormat(format)`). Wails 2.13 gibt nur `ClipboardGetText` und
// `ClipboardSetText` — einen eigenen Typ auf die Pasteboard-Ebene zu bringen
// hieße, an der Bindung vorbei mit macOS zu sprechen.
//
// **Was das kostet**: eine Karte lässt sich in einen Texteditor einfügen (und
// wieder heraus). **Was es bringt**: derselbe Weg trägt zwischen zwei
// Fenstern dieser Anwendung und über jede Grenze, die Text kennt.
//
// **Erkannt wird am Inhalt und nicht am Typ**: ohne `game_name` und ohne
// `card:` ist es keine Karte, und dann entsteht auch keine.

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Ablagekarten ist, was die Ablage aus dem Bestand liest.
type Ablagekarten interface {
	Karten(ctx context.Context, setID int64) ([]*domain.CardRow, error)
}

// Kartenablage kopiert Karten in die Zwischenablage und fügt sie ein.
type Kartenablage struct {
	karten    Ablagekarten
	koepfe    Kopfliste
	anleger   Kartenanleger
	schreiber Uebernahmeschreiber

	// loescher ist nur beim Ausschneiden nötig — siehe [Kartenablage.MitLoescher].
	loescher Ablageloescher

	// bilder legt die mitgeschickten Bilddateien im Zielset ab — oder ist
	// nichts, und dann geht nur der Verweis mit (#1237).
	bilder *Bildablage

	// felder sagt, welche Felder ein Zielset **speichert** — oder ist nichts,
	// und dann geht jeder Wert unverändert an den Schreiber
	// ([Kartenablage.MitFeldern]).
	felder Kartenfeldquelle
}

// NeueKartenablage baut die Ablage.
//
// **Der Anleger und der Schreiber sind zwei** — dieselbe Trennung, die der
// Scryfall-Weg schon hat: eine Karte entsteht leer, und danach kommen ihre
// Werte in **einem** Zug hinein.
func NeueKartenablage(
	karten Ablagekarten, koepfe Kopfliste, anleger interface {
		Kartenanleger
		Uebernahmeschreiber
	},
) *Kartenablage {
	return &Kartenablage{
		karten: karten, koepfe: koepfe, anleger: anleger, schreiber: anleger,
	}
}

// Ablageherkunft steht im Verlauf an einem eingefügten Wert.
//
// **Eine Herkunft und keine Adresse** — anders als bei Scryfall gibt es hier
// keine Seite zum Nachsehen. Wer den Verlauf liest, soll eine eingefügte Karte
// von einer getippten unterscheiden können.
const Ablageherkunft = "aus der Zwischenablage"

// Einfuegung ist, was beim Einfügen entstanden ist.
type Einfuegung struct {
	// Karten sind die Kennungen der neuen Karten, in der Reihenfolge des
	// Inhalts.
	Karten []int64

	// Befunde sind die Stellen, an denen etwas nicht lesbar war (§8).
	Befunde []domain.Befund
}

// Kopieren gibt die genannten Karten als Text für die Zwischenablage.
//
// **Nur Karten dieses Sets.** Eine Kennung, die nicht dazugehört, ist ein
// Fehler und kein übergangener Eintrag: wer drei Karten kopiert und zwei
// bekommt, merkt es erst beim Einfügen.
func (k *Kartenablage) Kopieren(
	ctx context.Context, setID int64, karteIDs []int64,
) (string, error) {
	if len(karteIDs) == 0 {
		return "", fmt.Errorf("%w: es wurde keine Karte genannt", domain.ErrValidation)
	}

	spiel, err := k.spielVon(ctx, setID)
	if err != nil {
		return "", err
	}

	karten, err := k.karten.Karten(ctx, setID)
	if err != nil {
		return "", fmt.Errorf("die Karten lesen: %w", err)
	}

	baum := &msepaket.Baum{Wurzel: []*msepaket.Knoten{
		{Schluessel: "game_name", Wert: spiel},
	}}

	genommen := make([]*domain.CardRow, 0, len(karteIDs))

	for _, gesucht := range karteIDs {
		stelle := slices.IndexFunc(karten, func(eine *domain.CardRow) bool {
			return eine.ID == gesucht
		})

		if stelle < 0 {
			return "", fmt.Errorf("%w: die Karte %d gehört nicht zum Set %d",
				domain.ErrNotFound, gesucht, setID)
		}

		baum.Wurzel = append(baum.Wurzel, kartenknoten(karten[stelle]))
		genommen = append(genommen, karten[stelle])
	}

	// **Und die Bilder, auf die sie zeigen** (#1237). Ein Bildfeld trägt einen
	// paketlokalen Namen, und der ist je Set aufgelöst.
	k.bilderAnhaengen(setID, genommen, baum)

	return string(msepaket.Schreibe(baum)), nil
}

// kartenknoten macht aus einer Karte einen `card:`-Block.
func kartenknoten(karte *domain.CardRow) *msepaket.Knoten {
	knoten := &msepaket.Knoten{Schluessel: "card"}

	for _, wert := range karte.Values {
		knoten.Kinder = append(knoten.Kinder, &msepaket.Knoten{
			Schluessel: wert.Field, Wert: wert.Text,
		})
	}

	return knoten
}

// Einfuegen legt die Karten aus einem Ablageinhalt im Zielset an.
//
// **Das Spiel wird vor den Karten geprüft** — MSEs Bedingung steht **um** das
// Lesen herum. Eine Karte eines anderen Spiels hat Felder, die es hier nicht
// gibt; sie anzulegen hieße, ein Set mit Werten zu füllen, die keine Vorlage
// zeichnet.
func (k *Kartenablage) Einfuegen(
	ctx context.Context, setID int64, inhalt string,
) (*Einfuegung, error) {
	if strings.TrimSpace(inhalt) == "" {
		return nil, fmt.Errorf(
			"%w: die Zwischenablage ist leer", domain.ErrValidation)
	}

	baum, err := msepaket.Zerteile("zwischenablage", []byte(inhalt))
	if err != nil {
		return nil, fmt.Errorf("%w: die Zwischenablage enthält keine Karten: %w",
			domain.ErrValidation, err)
	}

	spiel := ""
	if knoten := baum.Ersten("game_name"); knoten != nil {
		spiel = strings.TrimSpace(knoten.Wert)
	}

	if spiel == "" {
		return nil, fmt.Errorf(
			"%w: die Zwischenablage enthält keine Karten dieser Anwendung",
			domain.ErrValidation)
	}

	ziel, err := k.spielVon(ctx, setID)
	if err != nil {
		return nil, err
	}

	if spiel != ziel {
		return nil, fmt.Errorf(
			"%w: die kopierten Karten gehören zum Spiel %q, dieses Set zu %q",
			domain.ErrValidation, spiel, ziel)
	}

	bloecke := baum.Alle("card")
	if len(bloecke) == 0 {
		return nil, fmt.Errorf(
			"%w: die Zwischenablage enthält keine Karte", domain.ErrValidation)
	}

	// **Die Bilder zuerst.** Ein Wert, der auf ein Bild zeigt, bekommt beim
	// Anlegen den **neuen** Namen; wäre die Datei noch nicht da, zeigte er
	// einen Augenblick lang ins Leere (#1237).
	umbenannt, bildbefunde := k.bilderUebernehmen(setID, baum)
	bildbefunde = append(bildbefunde,
		verweiseOhneBild(k.formenVon(ctx, setID), bloecke, umbenannt)...)

	ergebnis, benutzt, fehler := k.anlegen(ctx, setID, bloecke, umbenannt)
	if fehler != nil {
		// **Und dann liegen sie umsonst da.** Eine Datei ohne Karte, die auf
		// sie zeigt, wandert später als Beiwerk in jede Sicherung, und
		// niemand kann sagen, wozu sie gehört. §8 und §10 fragen beide
		// danach: kein halber Stand. Angestrichen vom Zweitblick über die
		// Nacharbeit
		// (#1278,
		// Befund 2).
		//
		// **Die angelegten Karten bleiben stehen** — das ist die Zusage von
		// [Kartenablage.anlegen], und sie gilt weiter: eine Karte, die
		// entstanden ist, verschwindet nicht.
		//
		// **Und darum bleiben ihre Bilder auch stehen.** Der erste Anlauf nahm
		// **alle** neu geschriebenen Dateien zurück — auch die, auf die eine
		// geglückte Karte zeigt. Karten und Bilder haben verschiedene
		// Rücknahmegrenzen, und die eine muss der anderen folgen. Angestrichen
		// vom Nachlauf über die Nacharbeit
		// (#1282,
		// Befund 2).
		//
		// **Und ein Fehlschlag dabei wird gemeldet**, nicht verschluckt: der
		// halbe Stand, den er hinterlässt, ist genau der, den diese Rücknahme
		// verhindern soll (Befund 4).
		return ergebnis, errors.Join(fehler, k.bilderZuruecknehmen(setID, umbenannt, benutzt))
	}

	if ergebnis != nil {
		ergebnis.Befunde = append(bildbefunde, ergebnis.Befunde...)
	}

	return ergebnis, fehler
}

// anlegen legt je Block eine Karte an und füllt sie.
//
// **Eine Karte entsteht, bevor ihre Werte kommen** — dieselbe Reihenfolge wie
// beim Anlegen aus Scryfall. Scheitert das Füllen, bleibt die Karte stehen und
// die Kennung kommt zurück: eine Karte, die entstanden ist, verschwindet nicht.
func (k *Kartenablage) anlegen(
	ctx context.Context, setID int64, bloecke []*msepaket.Knoten,
	umbenannt map[string]string,
) (ergebnis *Einfuegung, benutzt map[string]bool, fehler error) {
	ergebnis = &Einfuegung{}

	// **Welche neuen Bildnamen wirklich an einer Karte stehen.** Nur was hier
	// **nicht** eingetragen ist, darf eine Rücknahme entfernen (#1282,
	// Befund 2).
	benutzt = map[string]bool{}

	for _, block := range bloecke {
		karteID, err := k.anleger.KarteAnlegen(ctx, setID)
		if err != nil {
			return ergebnis, benutzt, fmt.Errorf("eine Karte anlegen: %w", err)
		}

		ergebnis.Karten = append(ergebnis.Karten, karteID)

		werte := make([]domain.Value, 0, len(block.Kinder))

		// **Jedes Kind ist ein Eintrag mit Schlüssel.** Eine Wache gegen den
		// leeren Schlüssel wäre nicht erreichbar: `Zerteile` nimmt Leerzeilen
		// und Kommentare nicht als Kinder auf, und eine Zeile **mit**
		// Doppelpunkt und **ohne** Schlüssel ist dort ein Fehler (`zerteiler.go`,
		// „der Schlüssel ist leer") — der Lauf käme gar nicht bis hierher. Ein
		// Zweig, den kein Lauf betritt, ist eine Behauptung (§3).
		for _, kind := range block.Kinder {
			// **`wertText` und nicht `Wert`** — ein mehrzeiliger Wert steht
			// unter seinem Schlüssel und nicht dahinter, und der Regeltext
			// einer gewöhnlichen Karte ist einer
			// (#398).
			// **Ein Verweis auf ein mitgeschicktes Bild bekommt seinen neuen
			// Namen** (#1237). Im Zielset kann `image1` schon belegt sein.
			text := wertText(kind)
			if neu, umgezogen := umbenannt[strings.TrimSpace(text)]; umgezogen {
				text = neu
				benutzt[neu] = true
			}

			werte = append(werte,
				domain.Value{Field: kind.Schluessel, Text: text})
		}

		werte, uebergangen := k.speicherbare(ctx, setID, werte)
		if len(uebergangen) > 0 {
			ergebnis.Befunde = append(ergebnis.Befunde, domain.Befund{
				Meldung: "diese Felder rechnet die Vorlage des Zielsets und " +
					"speichert sie nicht — sie sind nicht mitgekommen: " +
					strings.Join(uebergangen, ", "),
			})
		}

		if len(werte) == 0 {
			continue
		}

		if _, _, err := k.schreiber.FelderVonAussen(
			ctx, setID, karteID, werte, Ablageherkunft); err != nil {
			// **Die Bilder dieser Karte gelten als unbenutzt.** Ihr Wert ist
			// nicht geschrieben worden; niemand zeigt auf sie.
			for _, wert := range werte {
				delete(benutzt, wert.Text)
			}

			return ergebnis, benutzt,
				fmt.Errorf("die Werte einer Karte setzen: %w", err)
		}
	}

	return ergebnis, benutzt, nil
}

// spielVon gibt das Spiel eines Sets.
func (k *Kartenablage) spielVon(ctx context.Context, setID int64) (string, error) {
	koepfe, err := k.koepfe.Setkoepfe(ctx)
	if err != nil {
		return "", fmt.Errorf("die Sets lesen: %w", err)
	}

	for i := range koepfe {
		if koepfe[i].ID == setID {
			return koepfe[i].Game, nil
		}
	}

	return "", fmt.Errorf("%w: das Set %d", domain.ErrNotFound, setID)
}

// Ablageloescher ist, was das Ausschneiden zum Entfernen braucht.
//
// **Über die Aufsicht und nicht über den Speicher** — sie führt den Verlauf,
// und ein ausgeschnittenes Set soll sich zurücknehmen lassen.
type Ablageloescher interface {
	// KartenLoeschen entfernt alle genannten Karten **in einem Zug**.
	//
	// **Ganz oder gar nicht** — Befund 4 des Zweitblicks über Phase 3.7. Hier
	// stand `KarteLoeschen` je Karte; scheiterte die dritte, waren die ersten
	// beiden fort, und die Ablage trug alle drei. MSE2 bündelt eine Auswahl
	// ebenso in **eine** Handlung (`src/gui/control/card_list.cpp`,
	// Zeilen 170 und 191–199, Fassung `f9d9356` vom 2021-07-06).
	//
	// **Das letzte Argument bestätigt ein Löschen, das einen Deckplatz
	// mitnimmt** (#1369).
	KartenLoeschen(
		ctx context.Context, setID int64, karteIDs []int64, trotzdem bool) error
}

// MitLoescher gibt eine Ablage, die auch ausschneiden kann.
//
// **Ein zweiter und kein umgestellter**, dieselbe Gestalt wie
// `Kartenschreiber.Als`: wer nur kopiert, braucht keinen Löscher.
func (k *Kartenablage) MitLoescher(loescher Ablageloescher) *Kartenablage {
	zweite := *k
	zweite.loescher = loescher

	return &zweite
}
