package app

// **Die Partie an der Naht** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Was hier steht und was nicht
//
// **Übersetzung und sonst nichts** (§2). Welche Sicht jemand bekommt,
// entscheidet [domain.Partie.Sicht]; ob ein Befehl durchkommt, die
// Partieleitung. Hier werden Kennungen zu Zeichenketten und Fehler zu
// Strukturen, die das Fenster lesen kann.
//
// # Warum die Sicht und nicht der Zustand über diese Naht geht
//
// **Weil ein DTO aus dem Vollzustand die fremde Hand mitnähme** — und zwar
// still: die Antwort sähe vollständiger aus, nicht falscher. Die Projektion
// ist deshalb der **einzige** Weg hierher; `Vollzustand` hat keinen DTO.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Partieleitung ist, was diese Naht vom Dienst braucht.
type Partieleitung interface {
	Anlegen(ctx context.Context, partie domain.Partie) (domain.Partie, error)
	Partie(ctx context.Context, kennung, fuer string) (domain.Partiesicht, error)

	// Partien gibt die Partien dieses Profils, zuletzt geänderte zuerst
	// (#1600).
	//
	// **Das Profil kommt mit**, denn „dran“ ist eine Frage an den Sitz, der
	// ihm gehört (UXR-F16,
	// #2031).
	Partien(ctx context.Context, profilID int64) ([]domain.Partiezeile, error)
	// Tischbild, BefehlVerkuendet und InReihe halten Bild und Verkündung in
	// einer Fassung und in einer Reihenfolge (#2481).
	Tischbild(ctx context.Context, kennung, fuer string) (
		domain.Tischbild, error)
	BefehlVerkuendet(ctx context.Context, befehl domain.Spielbefehl,
		verkuende service.Verkuendung) (service.Befehlsausgang, error)
	InReihe(tun func() error) error
	// Verlauf gibt den Ereignisstrom, **schon gefiltert** (G5, #1462).
	Verlauf(ctx context.Context, kennung, fuer string) ([]domain.Ereignis, error)

	// Wirtsstand und Nachreichen holen einen Gast zurück (G5, #1462).
	Wirtsstand(ctx context.Context, kennung, fuer string) (
		domain.Wirtsstand, error)
	Nachreichen(ctx context.Context, kennung, fuer string, abFolge int64) (
		[]domain.Ereignis, error)

	// Platznehmen setzt einen eingeladenen Gast mit seinem Deck (P9b, #2377).
	Platznehmen(ctx context.Context, kennung, sitz, peer string,
		deck domain.Partiedeck, verkuende service.Verkuendung) (
		service.Befehlsausgang, error)

	// Die vier Wege der Offenlegung und des Vergessens (G5, #1462).
	Freigaben(ctx context.Context, kennung string) (
		[]domain.Offenlegung, error)
	Offenlegen(ctx context.Context, kennung string,
		gabe domain.Offenlegung) error
	Zuruecknehmen(ctx context.Context, kennung, von string) error
	Vergessen(ctx context.Context, kennung string) error

	// Die vier Wege der Verwaltung — PLAY-03.AK-05
	// (#1711).
	//
	// **Die neue Kennung kommt von hier oben** — wie beim Anlegen, wo
	// `NeuePartiekennung()` sie stellt. Zwei Quellen für dieselbe Zusage
	// wären eine zu viel.
	Loeschen(ctx context.Context, kennung string) error
	Kopieren(ctx context.Context, kennung, neueKennung string) (
		domain.Partie, error)
	Ausfuhr(ctx context.Context, kennung string) ([]byte, error)
	Einfuhr(ctx context.Context, roh []byte, neueKennung string) (
		domain.Partie, []domain.Einfuhrbefund, error)

	Pausieren(ctx context.Context, kennung string) error
	Fortsetzen(ctx context.Context, kennung string) error
	Beenden(ctx context.Context, kennung string) error
}

// PartieDTO ist eine Partie, wie einer sie sehen darf.
type PartieDTO struct {
	ID      string `json:"id"`
	Fassung string `json:"fassung"`

	Spiel string `json:"spiel"`
	Stand string `json:"stand"`

	// Fuer ist die Kennung, für die gerechnet wurde.
	//
	// **Sie steht in der Antwort**, damit ein Empfänger merkt, wenn er die
	// Sicht eines anderen bekommen hat.
	Fuer string `json:"fuer"`

	// Wirt steht offen da — die Vertrauensgrenze wird genannt (0115-Gestalt).
	Wirt string `json:"wirt"`

	Regelwerk RegelwerkDTO `json:"regelwerk"`

	Sitze     []SitzDTO      `json:"sitze"`
	Zuschauer []ZuschauerDTO `json:"zuschauer"`
	Objekte   []ObjektDTO    `json:"objekte"`

	// Mischpflichten sind die verborgenen Zonen, die jemand angesehen und
	// seitdem nicht gemischt hat.
	//
	// **Für jeden dieselben.** Sie sind die öffentliche Hälfte eines Blicks;
	// wer sie nur dem Betroffenen zeigte, verschwiege dem Gegenüber, was
	// gerade geschehen ist.
	Mischpflichten []MischpflichtDTO `json:"mischpflichten"`

	// Legendenwahlen sind die offenen Fragen aus CR 704.5j (M12, #1614).
	//
	// **Für jeden dieselben**, und aus demselben Grund wie die
	// Mischpflichten: bis jemand geantwortet hat, steht das Spiel — wer
	// den Grund nicht sähe, hielte es für einen Ausfall.
	Legendenwahlen []LegendenwahlDTO `json:"legendenwahlen"`

	// Ausgeloest sind die Fähigkeiten, die ausgelöst haben und auf den
	// Stapel warten — CR 603.2 (M9, #1611).
	//
	// **Für jeden dieselben**, und aus demselben Grund wie die
	// Legendenwahlen: solange geordnet wird, steht das Spiel, und wer den
	// Grund nicht sähe, hielte es für einen Ausfall.
	Ausgeloest []AusloeserDTO `json:"ausgeloest"`

	// Ausloeserreihung ist der Sitz, der gerade ordnen soll — CR 603.3b.
	//
	// **Leer heißt: niemand wird gefragt.**
	Ausloeserreihung string `json:"ausloeserreihung"`

	// Zug ist, wo die Partie im Zug steht (G9, #1592) — der Grund steht bei
	// [ZugDTO].
	Zug ZugDTO `json:"zug"`

	// Ausgang ist, wie die Partie ausgegangen ist (M17, #1618).
	Ausgang AusgangDTO `json:"ausgang"`

	// Vorrat sind die Kartenkennungen des **eigenen** Decks — damit die
	// Bilder da sind, bevor die Karte kommt (G2,
	// #1459).
	//
	// **Leer für jeden, der nicht an diesem Sitz sitzt.** Die Reihenfolge
	// der Bibliothek bleibt geheim; welche Karten im eigenen Deck liegen,
	// hat der Eigner selbst aufgeschrieben.
	Vorrat []string `json:"vorrat"`
}

// MischpflichtDTO nennt Sitz und Zone einer offenen Mischpflicht.
type MischpflichtDTO struct {
	Sitz string `json:"sitz"`
	Zone string `json:"zone"`
}

// LegendenwahlDTO nennt Sitz und Namen einer offenen Wahl — CR 704.5j.
type LegendenwahlDTO struct {
	Sitz string `json:"sitz"`
	Name string `json:"name"`
}

// AusloeserDTO ist eine wartende Fähigkeit — CR 603.2 (M9, #1611).
//
// **Die Kennung ist der Griff**: der Befehl `spieler.reihung` nennt sie, und
// ohne sie ließen sich zwei Fähigkeiten derselben Art nicht unterscheiden.
type AusloeserDTO struct {
	Kennung     string `json:"kennung"`
	Art         string `json:"art"`
	Beherrscher string `json:"beherrscher"`

	// Wen ist, auf wen sie wirkt — leer, wo sie den Beherrscher meint.
	Wen string `json:"wen"`
}

// RegelwerkDTO sagt, welche Regeln in welcher Strenge gelten.
type RegelwerkDTO struct {
	Kennung string `json:"kennung"`
	Fassung int    `json:"fassung"`
	Strenge string `json:"strenge"`

	// Abgewaehlt sind die Bereiche, die die Runde selbst übernimmt —
	// PLAY-02.AK-02 und AK-03
	// (#1710).
	//
	// **Leer heißt: das Programm trägt, was es tragen kann.** Sie werden
	// mit dem Regelwerk **eingefroren** und ändern sich danach nicht mehr.
	Abgewaehlt []string `json:"abgewaehlt"`
}

// SitzDTO ist ein Sitz ohne das, was an ihm geheim ist.
type SitzDTO struct {
	Kennung string `json:"kennung"`
	Name    string `json:"name"`

	// Deckabdruck ist der Abdruck des Abzugs — **nicht** der Abzug.
	Deckabdruck string `json:"deckabdruck"`
	Karten      int    `json:"karten"`

	// Wartet sagt, dass dieser Sitz auf das Deck seines Gastes wartet, und
	// Eingeladen, auf wen — **nur am Fenster des Wirts** (P9b, #2377).
	Wartet     bool   `json:"wartet"`
	Eingeladen string `json:"eingeladen"`

	// ProfilID ist das Profil, an das dieser Sitz gebunden ist (G8,
	// #1465)
	// — **leer heißt: keines**, und das ist der Normalfall.
	ProfilID string `json:"profilID"`

	// Leben und Vergiftung machen den Sitz zum Spieler — CR 119.1 und
	// CR 122.1f (G9, #1592).
	//
	// **Für jeden sichtbar**: ein Lebensstand ist am Tisch öffentlich, und
	// wer ihn verschwiege, machte aus einer Buchführung ein Geheimnis.
	Leben      int `json:"leben"`
	Vergiftung int `json:"vergiftung"`

	// Manavorrat ist das ungenutzte Mana als Buchstabenfolge — `RRG`,
	// **leer heißt leer** (M3,
	// #1605).
	//
	// **Ein Text und keine Tabelle**: `RRG` ist die Auskunft, die jemand
	// liest, und jede Fläche zeigt dieselbe. Der Grund steht ausführlich bei
	// [service.Vorratstext].
	//
	// **Ohne `omitempty`**, wie jede Zusage an dieser Grenze: ein leerer
	// Vorrat ist eine leere Zeichenkette und nicht `undefined`.
	Manavorrat string `json:"manavorrat"`

	// Die fünf, die bis zum 2026-09-21 in der Domäne lagen — PLAY-03
	// (#1711).
	//
	// `domain.Sitz` führt zwölf Felder, die Sicht kopierte acht, und dieses
	// DTO dieselben acht. Gerechnet, gespeichert, geprüft — und **keines der
	// fünf** war je im Fenster zu sehen. Der Entwurf der Zieloberfläche
	// verlangt zwei davon wörtlich (Artboard `TischCommander`).
	//
	// **Keines ist geheim**: CR 728.1 setzt Strahlungsmarken an den
	// **Spieler**, CR 400.2 macht die Kommandozone öffentlich, CR 903.8
	// zählt Würfe, die alle gesehen haben, CR 903.10a Kampfschaden, und der
	// Griff in eine leere Bibliothek (CR 704.5b) geschieht am Tisch.
	Strahlung       int    `json:"strahlung"`
	Commander       string `json:"commander"`
	Commandersteuer int    `json:"commandersteuer"`
	Leergezogen     bool   `json:"leergezogen"`

	// Eroeffnung, Mulligans und Unterlegen sind der Stand im Beginn der
	// Partie — CR 103.5 (#1839). **Eroeffnung ist leer bei einer Partie von
	// vorher**; `ausstehend`, `offen` und `behalten` sonst.
	Eroeffnung string `json:"eroeffnung"`
	Mulligans  int    `json:"mulligans"`
	Unterlegen int    `json:"unterlegen"`

	// Commanderschaden ist der Kampfschaden **je Commander** — CR 903.10a.
	//
	// **Ohne `omitempty`**, wie jede Zusage an dieser Grenze: eine leere
	// Tabelle ist `{}` und nicht `undefined`. Genau das ist der Fall, den §3
	// als „zwischen den Prüfebenen" führt — die Frontendläufe gehen nie
	// durch die Serialisierung, und eine verschwundene leere Tabelle bliebe
	// bei grüner Pipeline unsichtbar.
	Commanderschaden map[string]int `json:"commanderschaden"`
}

// ZuschauerDTO sieht zu und hat keinen Sitz.
type ZuschauerDTO struct {
	Kennung string `json:"kennung"`
	Name    string `json:"name"`
}

// ObjektDTO ist ein Spielobjekt, so weit es gezeigt werden darf.
type ObjektDTO struct {
	Kennung string `json:"kennung"`
	Art     string `json:"art"`

	Beherrscher string `json:"beherrscher"`
	Eigner      string `json:"eigner"`
	Zone        string `json:"zone"`

	Karte string `json:"karte"`
	Name  string `json:"name"`

	// KarteID ist die Zeile im Bestand — der Weg zum **Bild** (G2, #1459).
	//
	// **Leer, wo nicht hineingesehen werden darf**, und als Zeichenkette wie
	// jede Kennung an dieser Naht.
	KarteID string `json:"karteID"`

	// Bild ist die undurchsichtige Kennung des Kartenbilds — P9 (#2377),
	// Vermerk 0198. Leer, wo nicht hineingesehen werden darf; anders als
	// `karteID` zeigt sie in keinen Bestand.
	Bild string `json:"bild"`

	// Eigenschaften sind, was diese Karte **ist** — `null`, wo nicht
	// hineingesehen werden darf (M2, #1604).
	Eigenschaften *EigenschaftenDTO `json:"eigenschaften"`

	Ordnung int `json:"ordnung"`

	Getappt bool `json:"getappt"`

	// Verdeckt heißt: es **liegt** verdeckt — eine Tatsache für alle.
	Verdeckt bool `json:"verdeckt"`

	// Geheim heißt: **dieser** Empfänger darf nicht hineinsehen.
	Geheim bool `json:"geheim"`

	HaengtAn string         `json:"haengtAn"`
	Haufen   string         `json:"haufen"`
	Zaehler  map[string]int `json:"zaehler"`
}

// EreignisDTO ist ein Ereignis, so weit es ein Empfänger bekommt.
type EreignisDTO struct {
	Folge string `json:"folge"`
	Art   string `json:"art"`

	Oeffentlich map[string]any `json:"oeffentlich"`

	// Privat ist die Ergänzung für **diesen** Empfänger — leer, wo keine
	// für ihn dabei war.
	Privat map[string]any `json:"privat"`

	Kette string `json:"kette"`
}

// ObjekthandlungDTO ist die Handlungsauskunft zu einem Objekt.
type ObjekthandlungDTO struct {
	Objekt     string   `json:"objekt"`
	Handlungen []string `json:"handlungen"`
	Primaer    string   `json:"primaer"`
	Zweit      string   `json:"zweit"`
}

// handlungenAus übersetzt die Auskunft — **eine leere Liste und kein `nil`**.
func handlungenAus(auskunft []domain.Objekthandlungen) []ObjekthandlungDTO {
	dtos := make([]ObjekthandlungDTO, 0, len(auskunft))
	for _, eine := range auskunft {
		dtos = append(dtos, ObjekthandlungDTO{
			Objekt: eine.Objekt, Handlungen: eine.Handlungen,
			Primaer: eine.Primaer, Zweit: eine.Zweit,
		})
	}

	return dtos
}

// PartieAntwort ist der Ausgang eines lesenden Partiewegs.
type PartieAntwort struct {
	Partie PartieDTO `json:"partie"`

	// Erlaubt sind die Befehlsarten, die dieser Empfänger stellen darf.
	Erlaubt []string `json:"erlaubt"`

	// Handlungen sagen je Objekt, was dieser Empfänger daran tun kann, und
	// welche Handlung Klick und Doppelklick auslösen — P3
	// (#2371).
	// **Eine leere Liste und kein `null`.**
	Handlungen []ObjekthandlungDTO `json:"handlungen"`

	// Befunde sind Auffälligkeiten, die die Partie **nicht** unbenutzbar
	// machen (G8,
	// #1465).
	//
	// **Ein Befund ist kein Fehler** (§8). Ein Sitz, dessen Profil es nicht
	// mehr gibt, bleibt ein Sitz — gespielt wird weiter, und die Bindung wird
	// **benannt** statt verschwiegen. Wer daraus einen Fehler machte, machte
	// eine Partie unbenutzbar, weil jemand ein Profil gelöscht hat.
	Befunde []string `json:"befunde"`

	// Nachladefehler sagt: angelegt oder gewechselt **ist**, nur die Sicht
	// danach war nicht zu lesen (§7, Regel 3; #1970). Dann trägt `partie`
	// nur Kennung und Sicht, damit eine Fläche selbst nachlesen kann.
	Nachladefehler *Fehler `json:"nachladefehler,omitempty"`

	// Absage sagt am fremden Tisch, dass der Wirt den letzten Wunsch abwies
	// (#2478) — mit Grund; sonst fehlt sie.
	Absage *WunschabsageDTO `json:"absage,omitempty"`

	Fehler *Fehler `json:"fehler"`
}

// BefehlAntwort ist der Ausgang eines Spielbefehls.
type BefehlAntwort struct {
	Fassung string `json:"fassung"`

	Ereignisse []EreignisDTO `json:"ereignisse"`

	Fehler *Fehler `json:"fehler"`
}
