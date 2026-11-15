package app

// **Die Bindungen der Partie** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Warum ein Entwurf hereinkommt und keine Partie
//
// **Weil das Fenster keine Objekte legt.** Ein Mensch wählt ein Spiel, ein
// Regelwerk und zwei Decks; **was daraus auf dem Tisch liegt**, rechnet der
// [service.Partieaufbau] — und zwar an einer Stelle für alle vier Flächen.
// Eine Bindung, die eine fertige Partie entgegennähme, ließe das Fenster
// entscheiden, wie viele Wälder in der Bibliothek liegen.
//
// # Und warum jeder Weg eine Sitzkennung trägt
//
// **Weil es die Frage „wer sieht das?" ohne sie nicht gibt.** Das Fenster
// zeigt die Sicht **eines** Sitzes; wer sie wegließe, bekäme eine Antwort
// ohne Empfänger — und die Projektion wüsste nicht, was sie kürzen soll.

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// PartieentwurfDTO ist der Wunsch nach einer Partie.
type PartieentwurfDTO struct {
	Spiel string `json:"spiel"`

	// Regelwerk darf leer sein — dann gilt der Sandkasten in Fassung 1.
	Regelwerk RegelwerkDTO `json:"regelwerk"`

	// Wirt ist die Sitzkennung dessen, der serialisiert.
	//
	// **Leer heißt: der erste Sitz.** Ein Fenster, das allein spielt, hat
	// keinen Grund, sich für einen zu entscheiden.
	Wirt string `json:"wirt"`

	Sitze []SitzwunschDTO `json:"sitze"`
}

// SitzwunschDTO ist ein Platz am Tisch samt Deck.
type SitzwunschDTO struct {
	Kennung string `json:"kennung"`
	Name    string `json:"name"`
	DeckID  string `json:"deckID"`

	// Commander ist die Kartenkennung des Commanders — CR 903.3 (M19,
	// #1620).
	// **Leer heißt: keiner**, und in einer Partie ohne Commanderprofil
	// bewirkt er nichts.
	Commander string `json:"commander"`

	// LeiheID ist eine **geliehene** Leihe statt eines eigenen Decks (G6,
	// #1463)
	// — leer heißt: das eigene.
	LeiheID string `json:"leiheID"`

	// ProfilID bindet diesen Sitz an ein Profil **desselben Nutzers** (G8,
	// #1465)
	// — **leer heißt: keines**, und das ist der Normalfall.
	ProfilID string `json:"profilID"`

	// Eingeladen ist die Peerkennung eines Gastes, der sein Deck mitbringt
	// (P9b, #2377) — dann bleiben Deck und Leihe leer.
	Eingeladen string `json:"eingeladen"`
}

// PartieAnlegen legt eine Partie an und gibt die Sicht ihres Wirts.
//
// **Schreibend** (`docs/SCHREIBWEGE.md`).
func (b *Bindungen) PartieAnlegen(ctx context.Context, entwurf PartieentwurfDTO) PartieAntwort {
	ctx = b.ruf(ctx)

	partie, fehler := b.app(ctx).partieAus(ctx, entwurf)
	if fehler != nil {
		return leerePartie(ctx, fehler)
	}

	return b.app(ctx).PartieAnlegen(ctx, partie)
}

// Partie gibt eine Partie, wie `fuer` sie sehen darf.
//
// **Rein lesend.**
func (b *Bindungen) Partie(ctx context.Context, kennung, fuer string) PartieAntwort {
	return b.app(ctx).Partie(b.ruf(ctx), kennung, fuer)
}

// PartieVerlauf gibt den Ereignisstrom, wie `fuer` ihn sehen darf.
//
// **Werkstück G5 von Phase 5.5**
// (#1462).
// **Rein lesend** — und gefiltert, bevor es hier ankommt.
func (b *Bindungen) PartieVerlauf(ctx context.Context, kennung, fuer string) PartieverlaufAntwort {
	return b.app(ctx).PartieVerlauf(b.ruf(ctx), kennung, fuer)
}

// PartieBefehl stellt einen Spielbefehl.
//
// **Schreibend** — und die Befehlskennung kommt vom Aufrufer, denn nur er
// weiß, ob er noch einmal dasselbe meint (§7, Regel 4).
func (b *Bindungen) PartieBefehl(
	ctx context.Context,
	partieID, befehlID, absender, art string, nutzlast map[string]any,
	erwartet string,
) BefehlAntwort {
	return b.app(ctx).PartieBefehl(b.ruf(ctx), domain.Spielbefehl{
		Kennung:  befehlID,
		PartieID: partieID,
		Absender: absender,
		Art:      art,
		Nutzlast: nutzlast,
		Erwartet: fassungAus(erwartet),
	})
}

// PartiePausieren hält eine Partie an.
func (b *Bindungen) PartiePausieren(ctx context.Context, kennung, fuer string) PartieAntwort {
	return b.app(ctx).PartiePausieren(b.ruf(ctx), kennung, fuer)
}

// PartieFortsetzen nimmt sie wieder auf.
func (b *Bindungen) PartieFortsetzen(ctx context.Context, kennung, fuer string) PartieAntwort {
	return b.app(ctx).PartieFortsetzen(b.ruf(ctx), kennung, fuer)
}

// PartieBeenden schließt sie ab.
func (b *Bindungen) PartieBeenden(ctx context.Context, kennung, fuer string) PartieAntwort {
	return b.app(ctx).PartieBeenden(b.ruf(ctx), kennung, fuer)
}

// fassungAus liest eine Fassung aus dem Text der Oberfläche.
//
// **Was keine Zahl ist, heißt null** — und null heißt „gegen die, die gerade
// gilt". Das ist der bequeme Fall, und ein Fehler dafür wäre eine Meldung
// über ein leeres Feld.
func fassungAus(text string) int64 {
	zahl, fehler := strconv.ParseInt(text, 10, 64)
	if fehler != nil {
		return 0
	}

	return zahl
}

// partieAus macht aus einem Entwurf eine spielbare Partie.
//
// # Die Rolle wird gefragt, bevor gelesen wird
//
// **Und nicht erst beim Schreiben.** Der Aufbau liest Decks aus dem Bestand
// — wer erst danach fragte, hätte einem Leser in einem fremden Profil schon
// geantwortet, bevor er ihm absagt.
//
// **Das ist die Gestalt, die N7 in Phase 5.4 aufgedeckt hat**
// (#1568):
// eine Prüfung, die es gibt und die hinter der Arbeit steht, ist keine.
// Gefunden hat es hier der Lauf, der je Weg nachzählt, **wonach** gefragt
// wurde — ein Lauf, der nur „es kam ein Fehler" geprüft hätte, wäre grün
// geblieben.
func (a *App) partieAus(
	ctx context.Context, entwurf PartieentwurfDTO,
) (domain.Partie, error) {
	sitze := make([]Sitzwunsch, 0, len(entwurf.Sitze))

	for _, sitz := range entwurf.Sitze {
		// **Eine Umwandlung und kein Literal.** Die beiden Verbünde haben
		// dieselben Felder in derselben Reihenfolge; ein Literal daneben
		// liefe beim ersten neuen Feld still auseinander, und staticcheck
		// sagt es (S1016).
		sitze = append(sitze, Sitzwunsch(sitz))
	}

	return a.partieAusWuenschenMit(
		ctx, entwurf.Spiel, entwurf.Wirt, sitze,
		regelwerkAus(entwurf.Regelwerk))
}

// partieAusWuenschenMit ist der gemeinsame Bauweg aller vier Flächen.
func (a *App) partieAusWuenschenMit(
	ctx context.Context, spiel, wirt string, sitze []Sitzwunsch,
	werk domain.Regelwerk,
) (domain.Partie, error) {
	if fehler := a.darf(ctx, domain.WirkungSchreiben); fehler != nil {
		return domain.Partie{}, fehler
	}

	wuensche := make([]service.Sitzwunsch, 0, len(sitze))

	for _, sitz := range sitze {
		// **Eine Leihe braucht keine Deckkennung** (G6,
		// #1463):
		// sie bringt ihren eingefrorenen Abzug mit. Wer sie trotzdem
		// deutete, wiese einen leeren Text als Fehler ab — und der ist hier
		// der Normalfall.
		profilID, profilfehler := profilkennungAus(sitz.ProfilID)
		if profilfehler != nil {
			return domain.Partie{}, profilfehler
		}

		// **Ein eingeladener Gast und eine Leihe bringen ihr Deck mit** —
		// eine Deckkennung gibt es für beide nicht (P9b, #2377).
		if sitz.LeiheID != "" || sitz.Eingeladen != "" {
			wuensche = append(wuensche, service.Sitzwunsch{
				Kennung: sitz.Kennung, Name: sitz.Name,
				LeiheID: sitz.LeiheID, ProfilID: profilID,
				Commander: sitz.Commander, Eingeladen: sitz.Eingeladen,
			})

			continue
		}

		deckID, fehler := strconv.ParseInt(sitz.DeckID, 10, 64)
		if fehler != nil {
			return domain.Partie{}, domain.MitGrund(
				domain.ErrValidation, fehler,
				"das Deck des Sitzes %q lesen", sitz.Kennung)
		}

		wuensche = append(wuensche, service.Sitzwunsch{
			Kennung: sitz.Kennung, Name: sitz.Name, DeckID: deckID,
			ProfilID: profilID, Commander: sitz.Commander,
		})
	}

	if wirt == "" && len(wuensche) > 0 {
		wirt = wuensche[0].Kennung
	}

	ctx = a.imProfil(ctx)

	// **Ohne Bindung sitzt das gezeigte Profil am Sitz des Wirts** (UXR-F21,
	// Vermerk 0164) — die Regel steht im Dienst. **Ohne gezeigtes Profil**
	// (ein Lauf ohne Anwendung) bindet sie nichts, und der Aufbau geht wie
	// zuvor: ein Fehler hier sagte nichts über die Partie.
	gezeigt, _ := gezeigtesProfil(ctx) //nolint:errcheck // ohne Profil keine Vorgabe

	return a.partieaufbau().Aufbauen(ctx, NeuePartiekennung(), spiel, werk, wirt,
		service.WirtImGezeigtenProfil(wuensche, wirt, gezeigt))
}

// regelwerkAus macht aus dem DTO ein Regelwerk — mit den Vorgaben von G1.
//
// **Der Sandkasten ist die Vorgabe**, und das ist dieselbe Entscheidung wie
// in `partieflaeche.go`: **welche** Regeln ein Format trägt, ist G4
// (#1461).
func regelwerkAus(dto RegelwerkDTO) domain.Regelwerk {
	werk := domain.Regelwerk{
		Kennung:    dto.Kennung,
		Fassung:    dto.Fassung,
		Strenge:    domain.Strenge(dto.Strenge),
		Abgewaehlt: bereicheAus(dto.Abgewaehlt),
	}

	if werk.Kennung == "" {
		werk.Kennung = "sandkasten"
	}

	if werk.Fassung < 1 {
		werk.Fassung = 1
	}

	if !domain.IstStrenge(werk.Strenge) {
		werk.Strenge = domain.StrengeSandkasten
	}

	return werk
}

// bereicheAus macht aus Zeichenketten Regelbereiche.
//
// **Ohne zu prüfen, ob es sie gibt** — das tut [domain.Regelwerk.Pruefe],
// und dort gehört es hin: eine Fläche, die es selbst täte, hätte die zweite
// Wahrheit darüber, was ein Bereich ist (§2).
//
// **Nil bleibt nil**: „nichts abgewählt" ist der Normalfall, und eine leere
// Liste daneben wäre dasselbe mit mehr Aufwand.
func bereicheAus(namen []string) []domain.Regelbereich {
	if len(namen) == 0 {
		return nil
	}

	bereiche := make([]domain.Regelbereich, 0, len(namen))

	for _, name := range namen {
		bereiche = append(bereiche, domain.Regelbereich(name))
	}

	return bereiche
}

// NeuePartiekennung gibt eine Kennung für eine neue Partie.
//
// # Warum hier kein Fehlerweg steht
//
// **Weil `crypto/rand.Read` keinen zurückgibt.** Nachgelesen am 2026-09-15
// im Quelltext von Go 1.27 (`crypto/rand/rand.go`): schlägt die Quelle des
// Betriebssystems fehl, bricht das Programm ab — `fatal(...)`, gefolgt von
// `panic("unreachable")`. Der Rückgabewert ist dann nie `nil`-verschieden,
// und ein Zweig dafür wäre einer, den kein Lauf betreten kann (§3).
//
// **Sechzehn Byte und keine UUID-Form.** Eine Kennung wird hier verglichen
// und nicht gelesen; die vier Bindestriche kosten Platz und sagen nichts.
func NeuePartiekennung() string {
	roh := make([]byte, 16)

	_, _ = rand.Read(roh) //nolint:errcheck // siehe darüber

	return hex.EncodeToString(roh)
}

// profilkennungAus deutet eine Profilkennung — **leer ist null und kein
// Fehler**.
//
// **Weil die Bindung freiwillig ist** (G8,
// #1465):
// ein Sitz ohne Profil ist der Normalfall, und ein leeres Feld als Fehler
// abzuweisen machte aus dem Normalfall den Ausnahmefall.
func profilkennungAus(wert string) (int64, error) {
	if wert == "" {
		return 0, nil
	}

	gedeutet, fehler := strconv.ParseInt(wert, 10, 64)
	if fehler != nil {
		return 0, domain.MitGrund(domain.ErrValidation, fehler,
			"das Profil eines Sitzes lesen")
	}

	return gedeutet, nil
}
