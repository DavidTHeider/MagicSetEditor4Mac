// **Aus einem Deck wird ein Tisch** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Was hier geschieht und warum es hier geschieht
//
// **Ein Deck wird eingefroren, und aus dem Abzug wird eine Bibliothek.** Das
// ist der Schritt, ohne den G1 ein Modell ohne Anwendung wäre: es gibt keinen
// Weg, eine Partie von Hand mit sechzig Objekten zu füllen.
//
// **In `service` und nicht in `domain`**, weil es zwei Entitäten verbindet —
// ein [domain.Deck] und eine [domain.Partie] —, und nicht im Speicher, weil
// es eine fachliche Entscheidung ist: **was** eingefroren wird, ist keine
// Frage der Zeilen.
//
// # Der Abzug ist mehr als die Liste
//
// **Das Issue sagt es**: „Turnier-Deckabzug allein genügt nicht." Mitgehen
// müssen die Kartennamen, der gewählte Druck, die Bereiche und — sobald es
// sie gibt — die Verweise auf Bilder und Schriften. **Was heute fehlt, steht
// als `Beigaben` da und ist leer**: G2
// (#1459)
// baut die Texturen, und erst dann gibt es Verweise, die sich einfrieren
// lassen. Eine leere Liste sagt das; ein fehlendes Feld sagte gar nichts.

package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Deckquelle ist, was der Aufbau über ein Deck braucht.
//
// **Eine eigene Schnittstelle mit einem Verfahren** (§2) — derselbe Schnitt
// wie bei [Deckabzug] daneben: der Aufbau liest ein Deck genau einmal und
// schreibt keines.
type Deckquelle interface {
	Deck(ctx context.Context, kennung int64) (domain.Deck, error)
}

// Sitzwunsch ist, wer mit welchem Deck an den Tisch will.
type Sitzwunsch struct {
	Kennung string
	Name    string

	// DeckID ist das Deck, das eingefroren wird.
	DeckID int64

	// Commander ist die Kartenkennung des Commanders — CR 903.3 (M19,
	// #1620).
	//
	// **Leer heißt: keiner**, und das ist der Normalfall — es gibt eine
	// Commandervariante und acht andere Spielweisen.
	Commander string

	// ProfilID bindet diesen Sitz an ein Profil **desselben Nutzers** (G8,
	// #1465)
	// — **freiwillig**, null heißt keines. Der Grund steht bei
	// [domain.Sitz.ProfilID].
	ProfilID int64

	// LeiheID ist eine **geliehene** Leihe statt eines eigenen Decks (G6,
	// #1463).
	//
	// **Sie schlägt das Deck.** Wer eine Leihe vorlegt, bringt einen
	// eingefrorenen Abzug mit; ein Deck daneben wäre eine zweite Antwort auf
	// dieselbe Frage. Leer heißt: das eigene Deck.
	LeiheID string

	// Eingeladen ist die Peerkennung eines Gastes, der **sein Deck
	// mitbringt** (P9b, #2377) — dann bleiben Deck und Leihe leer, und der
	// Sitz wartet, bis er sich setzt.
	Eingeladen string
}

// Blattquelle gibt her, was eine Karte zum Zeichnen braucht.
//
// # Wofür es sie gibt
//
// **Damit ein verliehener Abzug auf einem fremden Rechner zeichnet** (G6,
// #1463).
// Die Kennung einer Karte macht ihr Bild auf **diesem** Rechner erreichbar;
// wer sein Deck verleiht, verleiht es an jemanden, der diesen Bestand nicht
// hat. Erst die eingefrorenen Feldwerte und der Stil machen den Abzug
// selbsttragend.
//
// **Zwei Verfahren und keine der großen Schnittstellen** (§2): der Aufbau
// braucht eine Karte und den Stil ihres Sets, und sonst nichts von einem
// Bestand.
type Blattquelle interface {
	KarteLesen(ctx context.Context, karteID int64) (*domain.CardRow, error)

	// Setstil gibt das Stilpaket eines Sets.
	//
	// **Und nicht den ganzen Kopf**: gebraucht wird ein Name, und ein
	// Verfahren, das mehr gäbe, hinge an mehr (§2). `Setkopf` heißt an
	// `storage.SetLeser` außerdem schon etwas anderes — der Kopf einer
	// MSE-Datei.
	Setstil(ctx context.Context, setID int64) (string, error)
}

// Leihvorlage ist, was der Aufbau über eine geliehene Leihe braucht.
//
// **Ein Verfahren** (§2) — der Aufbau legt vor und schreibt keine Leihe.
type Leihvorlage interface {
	Vorlegen(ctx context.Context, kennung string, bindung domain.Leihbindung,
		recht domain.Leihrecht) (domain.Partiedeck, error)
}

// Partieaufbau macht aus Wünschen eine spielbare Partie.
type Partieaufbau struct {
	decks    Deckquelle
	blaetter Blattquelle
	leihen   Leihvorlage
}

// NeuerPartieaufbau baut ihn.
func NeuerPartieaufbau(decks Deckquelle) *Partieaufbau {
	return &Partieaufbau{decks: decks}
}

// MitBlaettern hängt die Quelle für Feldwerte und Stile ein.
//
// # Warum optional und nicht im Bauplatz
//
// **Weil eine Partie auch ohne sie eine ist.** Der Abzug bleibt dann schmal —
// Name, Menge, Bereich, Kartenkennung —, und genau so hat G1 ihn gebaut. Wer
// sie zur Pflicht machte, zwänge jeden Lauf und jede Fläche, die nur eine
// Partie anlegt, einen Kartenbestand mitzubringen.
//
// **Und warum sie überhaupt fehlen darf**: fehlt sie, fehlen die Feldwerte —
// nicht der Tisch. Ein Abzug ohne sie spielt auf diesem Rechner genauso; er
// reist nur nicht.
func (a *Partieaufbau) MitBlaettern(blaetter Blattquelle) *Partieaufbau {
	a.blaetter = blaetter

	return a
}

// MitLeihen hängt die Stelle ein, die geliehene Abzüge vorlegt.
//
// **Optional wie die Blätter daneben, und mit demselben Grund**: eine Fläche,
// die nur eine Partie aus eigenen Decks anlegt, braucht keine Leihen. **Wer
// sie ausließe und trotzdem eine Leihkennung schickte, bekommt eine
// Meldung** — und keinen stillen Rückfall auf ein Deck, das es nicht gibt.
func (a *Partieaufbau) MitLeihen(leihen Leihvorlage) *Partieaufbau {
	a.leihen = leihen

	return a
}

// Aufbauen gibt die Partie, wie sie beim ersten Zug aussieht.
//
// **Alles liegt in der Bibliothek.** Kein Mischen, kein Ziehen, keine
// Eröffnungshand — das sind Befehle, und Befehle stehen im Strom. Eine
// Partie, die schon gemischt anfinge, hätte einen Zustand, den kein Ereignis
// erklärt (#1458: „derselbe Startsnapshot plus dieselbe Folge").
//
// **Die Befehle dafür gibt es seit #1839** (`spieler.starthand` und die drei
// danach), und jeder Sitz beginnt mit [domain.EroeffnungAusstehend]: nur
// dann werden sie angeboten.
func (a *Partieaufbau) Aufbauen(
	ctx context.Context, kennung, spiel string, werk domain.Regelwerk,
	wirt string, wuensche []Sitzwunsch,
) (domain.Partie, error) {
	// **Das Regelwerk wird eingefroren wie das Deck** (G4,
	// #1461).
	//
	// **Hier und nicht beim ersten Befehl**: dieselbe Partie muss nach einem
	// Neustart dieselben Entscheidungen geben, auch wenn draußen inzwischen
	// eine neue Fassung gilt. Das Profil kommt aus [domain.Grundprofil] und
	// trägt Fassung, Wirksamkeitsdatum, Quelle und die Unterstützungsmatrix.
	// **Welches Profil, sagt das gewünschte Regelwerk** (M19,
	// #1620):
	// wer eine Commanderpartie will, nennt sie beim Anlegen. Was der
	// Aufrufer sonst hineinschreibt, wird zum Sandkasten — ein Profil, das
	// niemand kennt, setzt nichts durch.
	profil := domain.ProfilZu(regelwerkskennung(werk, wuensche), spiel)

	// **Der Griff folgt dem Profil und nicht umgekehrt.** Wer eine Fassung
	// von Hand hineinschriebe, hätte zwei Angaben, die auseinanderlaufen —
	// [domain.Partie.Pruefe] weist das ab, und hier gibt es gar nicht erst
	// zwei Quellen.
	werk.Kennung = profil.Kennung
	werk.Fassung = profil.Fassung

	// **Was die Runde selbst übernimmt, wird mit eingefroren** —
	// PLAY-02.AK-03
	// (#1710).
	// Die Matrix der Partie ist damit die **wirksame**: ein abgewählter
	// Bereich steht in ihr als `manuell`, und jede Fläche, die sie zeigt,
	// zeigt, was für **diese** Runde gilt.
	profil.Matrix = domain.Regelwahl{
		Abgewaehlt: werk.Abgewaehlt,
	}.Wirksam(profil.Matrix)

	partie := domain.Partie{
		ID:        kennung,
		Spiel:     spiel,
		Regelwerk: werk,
		Profil:    profil,
		Wirt:      wirt,
		Stand:     domain.StandLaeuft,
		Sitze:     make([]domain.Sitz, 0, len(wuensche)),
		Objekte:   []domain.Spielobjekt{},
	}

	for _, wunsch := range wuensche {
		// **Wer eingeladen ist, bringt sein Deck mit** (P9b, #2377) — sein
		// Sitz entsteht leer und wartet.
		abzug := domain.Partiedeck{}

		if wunsch.Eingeladen == "" {
			gefroren, fehler := a.abzugFuer(ctx, kennung, spiel, wunsch)
			if fehler != nil {
				return domain.Partie{}, fehler
			}

			abzug = gefroren
		}

		partie.Sitze = append(partie.Sitze, domain.Sitz{
			Kennung: wunsch.Kennung, Name: wunsch.Name, Deck: abzug,
			ProfilID: wunsch.ProfilID, Commander: wunsch.Commander,
			Eroeffnung: domain.EroeffnungAusstehend, Eingeladen: wunsch.Eingeladen,
		})

		// **Ein eingeladener Sitz legt nichts** — sein Abzug ist leer, und ein
		// leerer ist kein Deck (`Partiedeck.Pruefe`).
		if wunsch.Eingeladen != "" {
			continue
		}

		gelegt, fehler := bibliothekAus(abzug, wunsch.Kennung)
		if fehler != nil {
			return domain.Partie{}, fehler
		}

		partie.Objekte = append(partie.Objekte, gelegt...)
	}

	// **Das Leben kommt aus dem eingefrorenen Profil** — CR 119.1 und
	// CR 103.4 (Fassung 2026-08-19), Werkstück G9 von Phase 5.5
	// (#1592).
	// `Regelprofil.Startleben` stand seit G4 da und wurde von nichts gelesen.
	partie = partie.MitStartleben()

	// **Und der erste Zug fängt an** — CR 103.1 und CR 500.1 (Fassung
	// 2026-09-25).
	//
	// **Ohne Zweig über die Matrix**, und das ist Absicht: der Aufbau nimmt
	// sein Profil aus [domain.Grundprofil], und das führt `zugstruktur` seit
	// G9 auf `automatisch`. Ein `if` darüber hätte einen Ast, den weder ein
	// Lauf noch der Betrieb erreicht (§10).
	//
	// **Die Zusage an die Partien aus G1 bis G8 hängt woanders**: sie werden
	// nicht aufgebaut, sondern gelesen, und ihr eingefrorenes Profil sagt
	// weiter `manuell`. Was daraus folgt, entscheidet der Sandkasten beim
	// Auslegen eines Befehls.
	//
	// **Der Wirt fängt an.** CR 103.1 lässt den Startspieler bestimmen; wer
	// das ist, steht hier fest, weil der Wirt die Partie hält. Eine Wahl
	// darüber gehört zum Spielbeginn und nicht zum Aufbau.
	partie = partie.ZugBeginnen(partie.Wirt)

	if fehler := partie.Pruefe(); fehler != nil {
		return domain.Partie{}, fehler
	}

	return partie, nil
}

// abzugFuer gibt den Abzug eines Sitzes — aus einem Deck oder aus einer Leihe.
func (a *Partieaufbau) abzugFuer(
	ctx context.Context, partie, spiel string, wunsch Sitzwunsch,
) (domain.Partiedeck, error) {
	if wunsch.LeiheID != "" {
		return a.geliehenerAbzug(ctx, partie, wunsch.LeiheID)
	}

	deck, fehler := a.decks.Deck(ctx, wunsch.DeckID)
	if fehler != nil {
		return domain.Partiedeck{}, fehler
	}

	if deck.Spiel != spiel {
		return domain.Partiedeck{}, fmt.Errorf(
			"%w: das Deck %q gehört zu %q, die Partie zu %q",
			domain.ErrValidation, deck.Name, deck.Spiel, spiel)
	}

	abzug := abzugAus(deck)

	// **Erst die Karten, dann der Abdruck** — er rechnet nur über die Liste
	// (siehe [abdruckDesAbzugs]), und deshalb ändert das Anreichern ihn
	// nicht. Das ist gewollt: zwei Seiten meinen dasselbe Deck, gleich ob
	// eine davon die Feldwerte hat.
	a.reichereAn(ctx, &abzug)

	return abzug, nil
}

// geliehenerAbzug legt eine Leihe vor und nimmt ihren Abzug.
//
// **Der Abzug wird nicht angereichert.** Er kommt eingefroren von drüben; wer
// ihn hier mit **seinem** Bestand ergänzte, ergänzte ihn mit anderen Karten —
// und der Abdruck stimmte danach nicht mehr mit dem überein, den das
// unterschriebene Dokument nennt.
func (a *Partieaufbau) geliehenerAbzug(
	ctx context.Context, partie, leiheID string,
) (domain.Partiedeck, error) {
	if a.leihen == nil {
		return domain.Partiedeck{}, fmt.Errorf(
			"%w: diese Fläche kann keine Leihe vorlegen", domain.ErrValidation)
	}

	return a.leihen.Vorlegen(ctx, leiheID, domain.Leihbindung{
		Art: domain.LeiheFuerPartie, Woran: partie,
	}, domain.RechtSpielen)
}

// **Wer das Regelwerk nennt, bekommt es.** Das Fenster tut es über
// `PartieentwurfDTO.Regelwerk`, und wer es dort hinschreibt, meint es.
//
// # Und wo keines genannt ist, sagen es die Decks
//
// **CR 903.3: in einer Commanderpartie benennt jedes Deck einen
// Commander.** Ein Tisch, an dem ein Sitz einen nennt, ist einer — das Feld
// gibt es nur dafür, und es steht in keinem anderen Zusammenhang.
//
// **Warum überhaupt zwei Wege**: `PartieAnlegen` heißt an den drei Flächen
// ohne Fenster `(spiel, wirt, sitze)` und trägt kein Regelwerk. Es dort
// nachzureichen wäre eine **brechende** Änderung an drei Flächen
// ([0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md));
// die Ableitung ist erweiternd und kostet niemanden etwas. **Kommt eine
// zweite Variante dazu, bekommt sie ein eigenes Feld** — dann reicht ein
// Commander als Erkennungszeichen nicht mehr.
func regelwerkskennung(werk domain.Regelwerk, wuensche []Sitzwunsch) string {
	if werk.Kennung == domain.Commanderkennung {
		return domain.Commanderkennung
	}

	for _, wunsch := range wuensche {
		if wunsch.Commander != "" {
			return domain.Commanderkennung
		}
	}

	return werk.Kennung
}

// Mitbringen gibt ein eigenes Deck so, wie ein Gast es an einen fremden Tisch
// mitbringt (P9b, #2377) — **derselbe Abzug wie am eigenen Tisch**,
// eingefroren und angereichert, und dazu sein Spiel: ein Abzug nennt es
// nicht, und der Wirt weist ein Deck eines anderen Spiels ab.
//
// **Die Quelle setzt der Wirt** — sie ist der Peer am anderen Ende seiner
// Leitung, und was hier stünde, wäre eine Behauptung.
func (a *Partieaufbau) Mitbringen(
	ctx context.Context, deckID int64,
) (domain.Mitgebracht, error) {
	deck, fehler := a.decks.Deck(ctx, deckID)
	if fehler != nil {
		return domain.Mitgebracht{}, fehler
	}

	abzug := abzugAus(deck)
	a.reichereAn(ctx, &abzug)

	return domain.Mitgebracht{Deck: &abzug, Spiel: deck.Spiel}, nil
}
