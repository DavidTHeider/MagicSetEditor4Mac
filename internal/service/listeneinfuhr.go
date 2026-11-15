package service

// **Eine Kartenliste in die Sammlung einlesen** — Werkstück 2 von Phase 4.5
// (#1292).
//
// # Der Befund
//
// Die Abnahme von #1292 verlangt, dass sich ein Bestand einlesen lässt,
// **ohne dass jemand eine Kennung von Hand kennt**. Bis hierher braucht jedes
// Exemplar eine Scryfall-UUID — die schreibt niemand ab, und für fünftausend
// Karten schreibt sie erst recht niemand fünftausendmal ab.
//
// # Warum zwei Schritte und nicht einer
//
// **§16, Regel 6: „Nichts wird still übernommen."** [Listeneinfuhr.Deuten]
// liest, ordnet zu und schreibt **nichts**; es sagt, was geschähe.
// [Listeneinfuhr.Uebernehmen] führt aus, was entschieden ist.
//
// Ein Schritt ginge auch — und schriebe dann 72 fremde Drucke in eine
// Sammlung, weil „Counterspell" 72-mal gedruckt wurde.
//
// # Was am 2026-08-30 nachgelesen wurde (§16, Regel 1)
//
// Gegen `api.scryfall.com`, mit dem Kopf aus §16 Regel 2:
//
//	!"Lightning Bolt" e:m11 cn:149   → total_cards 1
//	!"Counterspell"                  → total_cards 72
//	!"Gibtesnichtxyz"                → 404, code "not_found"
//
// **Die 72 sind der Grund für den zweiten Schritt.** Ein Name ohne Set ist
// keine Angabe über einen Druck, und `/cards/named` gäbe zwar **einen**
// zurück — für „Counterspell" den aus `dsc` von 2024 —, aber das ist
// Scryfalls Wahl und nicht die der Person. Sie still einzutragen hieße, eine
// Sammlung mit Drucken zu füllen, die niemand besitzt.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/sammelliste"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
)

// Deutungsgrenze ist die Zahl der Drucke, die zu einer Zeile gezeigt werden.
//
// **Zwanzig, und die Gesamtzahl steht daneben.** „Counterspell" hat 72; alle
// zu schicken hieße, eine Antwort um das Dreifache zu blähen für eine Liste,
// die niemand bis zum Ende liest. Was **nicht** geht, ist stilles Kappen
// (§10) — deshalb [Zeilendeutung.Drucke] neben `Treffer`.
const Deutungsgrenze = 20

// Zeilendeutung ist eine gelesene Zeile samt dem, was zu ihr gefunden wurde.
type Zeilendeutung struct {
	Zeile sammelliste.Zeile

	// Treffer sind die Drucke, die passen — höchstens [Deutungsgrenze] viele.
	//
	// Null heißt: nichts gefunden. Einer heißt: eindeutig, und er wird
	// übernommen. Mehrere heißen: **offen**, und jemand muss wählen.
	Treffer []Suchtreffer

	// Drucke ist die Zahl der passenden Drucke — auch wenn `Treffer` kürzer ist.
	Drucke int

	// Befund sagt, warum diese Zeile nicht aufgeht — leer, wenn sie aufgeht.
	Befund string

	// Weitere sagt, dass es mehr Drucke gibt, als die Quelle auf einer Seite
	// nennt — dann ist `Drucke` eine Untergrenze.
	Weitere bool

	// Fehler ist der Ausfall, an dem die Frage scheiterte — nil, wenn sie
	// eine Antwort bekam, auch die, dass es keinen Druck gibt. **Er geht mit,
	// damit eine Oberfläche ihn übersetzt** statt den deutschen `Befund` zu
	// zeigen (G01, `g01einsteiger`).
	Fehler error
}

// Eindeutig sagt, ob diese Zeile ohne Rückfrage übernommen werden kann.
func (z Zeilendeutung) Eindeutig() bool { return z.Drucke == 1 && len(z.Treffer) == 1 }

// Listendeutung ist, was aus einer eingefügten Liste zu machen wäre.
type Listendeutung struct {
	Zeilen []Zeilendeutung

	// Befunde sind die Zeilen, die schon der Zerteiler nicht lesen konnte.
	Befunde []domain.Befund

	// Alter ist das Alter der ältesten Antwort, die eingegangen ist.
	//
	// **Weil sie aus dem Zwischenspeicher kommen darf** (§16, Regel 5): eine
	// Deutung ohne Netz ist möglich, solange Scryfall schon einmal gefragt
	// wurde — und dann gehört dazu, wie alt sie ist.
	Alter time.Duration
}

// Listenzeile ist eine Zeile, für die entschieden ist, welcher Druck es ist.
//
// **Sie heißt nicht `Uebernahme`**: den Namen trägt seit Phase 3.2 die
// Übernahme von MSE2-Vorlagen (`paketuebernahme.go`), und zwei Bedeutungen
// unter einem Namen im selben Paket wären eine Falle für den nächsten Leser.
type Listenzeile struct {
	Printing string
	Menge    int
	Zustand  domain.Zustand
	Sprache  string

	// Finish ist die physische Ausführung — leer heißt `nonfoil`
	// ([domain.Exemplar.Ausfuehrung]).
	Finish domain.Finish

	// Name, Setcode und Nummer sind der lesbare Bezug (§16, Regel 6: was
	// übernommen wird, trägt seine Herkunft).
	Name    string
	Setcode string
	Nummer  string
}

// Einfuhrbericht sagt, was eingelesen wurde.
type Einfuhrbericht struct {
	// Angelegt sind die Kennungen der neuen Zeilen.
	Angelegt []int64

	// Erhoeht ist die Zahl der Zeilen, deren Menge gewachsen ist.
	//
	// # Warum erhöht und nicht abgewiesen
	//
	// Der Bestand hat eine Eindeutigkeit über Druck, Zustand, Sprache und
	// Ausführung, und [Sammlungspflege.Anlegen] gibt darauf einen
	// [domain.ErrConflict] — richtig für **eine** Eingabe von Hand: wer
	// zweimal dasselbe einträgt, hat sich vertan.
	//
	// **Beim Einlesen einer Liste heißt derselbe Fall etwas anderes**: wer
	// eine zweite Liste einliest, hat mehr davon. Erhöht wird deshalb, und es
	// steht in diesem Bericht — die Eindeutigkeit bleibt gewahrt, und nichts
	// geschieht still.
	Erhoeht int

	// Ausgelassen ist die Zahl der Zeilen, die eine vorhandene trafen und
	// beim Auslassen stehen blieben (DOM-03.AK-01, #1698).
	Ausgelassen int

	// Stueck ist die Zahl der Stücke, die dazugekommen sind.
	Stueck int

	Befunde []domain.Befund
}

// Fremddrucksuche ist, was die Einfuhr von einem fremden Dienst braucht.
//
// # Warum nicht [Fremdsuche]
//
// **Weil es eine andere Frage ist.** `Suche` gibt eine Zeile je **Karte** —
// Scryfalls Vorgabe `unique=cards`, und für eine Kartensuche richtig. Die
// Einfuhr fragt: *welche Ausgaben gibt es?* Ohne `unique=prints` gibt Scryfall
// für „Counterspell" **eine** statt 72, und damit sähe jede Zeile einer
// eingefügten Liste eindeutig aus.
//
// **Gefunden hat es die laufende Brücke** und kein Test — der Grund steht bei
// [scryfall.Klient.Druckesuche]. Zwei grüne Seiten einer Grenze sind kein
// Nachweis für die Grenze.
type Fremddrucksuche interface {
	Druckesuche(ctx context.Context, frage string) (
		treffer scryfall.Treffer, alter time.Duration,
		befunde []domain.Befund, fehler error)
}

// Listeneinfuhr liest eine Kartenliste in die Sammlung.
type Listeneinfuhr struct {
	suche    Fremddrucksuche
	sammlung Sammlungsablage

	// sammelsuche fragt viele Drucke auf einmal — wahlfrei, der Grund steht
	// bei [Listeneinfuhr.MitSammelsuche].
	sammelsuche Fremdsammelsuche
}

// NeueListeneinfuhr baut sie.
func NeueListeneinfuhr(suche Fremddrucksuche, sammlung Sammlungsablage) *Listeneinfuhr {
	return &Listeneinfuhr{suche: suche, sammlung: sammlung}
}

// Deuten liest die Liste und ordnet jede Zeile zu — **ohne zu schreiben**.
func (l *Listeneinfuhr) Deuten(ctx context.Context, text string) (Listendeutung, error) {
	if strings.TrimSpace(text) == "" {
		return Listendeutung{}, fmt.Errorf(
			"%w: eine leere Liste hat nichts einzulesen", domain.ErrValidation)
	}

	zeilen, befunde := sammelliste.Lies(text)

	deutung := Listendeutung{
		Zeilen:  make([]Zeilendeutung, 0, len(zeilen)),
		Befunde: befunde,
	}

	// **Gleiche Fragen werden einmal gestellt.** Eine Deckliste nennt
	// dieselbe Karte oft mehrfach; ohne dieses Gedächtnis ginge je Zeile eine
	// Anfrage hinaus, und die Drossel aus §16 machte daraus Wartezeit.
	gedaechtnis, alter, fehler := l.imSchub(ctx, zeilen)
	if fehler != nil {
		return Listendeutung{}, fehler
	}

	deutung.Alter = alter

	for _, zeile := range zeilen {
		frage := frageFuer(zeile)

		gedeutet, schon := gedaechtnis[frage]
		if !schon {
			var einzeln time.Duration

			gedeutet, einzeln = l.zeileDeuten(ctx, frage)
			gedaechtnis[frage] = gedeutet

			if einzeln > deutung.Alter {
				deutung.Alter = einzeln
			}
		}

		gedeutet.Zeile = zeile
		deutung.Zeilen = append(deutung.Zeilen, gedeutet)
	}

	return deutung, nil
}

// imSchub deutet vorab, was sich sammeln lässt — und gibt sonst nichts.
//
// **Ohne gesetzten Weg ist es eine leere Karte**, und die Schleife darüber
// fragt jede Zeile einzeln. Der Grund steht bei [Listeneinfuhr.MitSammelsuche]
// (#1348).
func (l *Listeneinfuhr) imSchub(ctx context.Context, zeilen []sammelliste.Zeile) (
	map[string]Zeilendeutung, time.Duration, error,
) {
	if l.sammelsuche == nil {
		return map[string]Zeilendeutung{}, 0, nil
	}

	// **In der Reihenfolge der Liste und ohne Wiederholung.** Eine Map allein
	// gäbe bei jedem Lauf eine andere Folge — und damit andere Schübe, andere
	// Schlüssel im Zwischenspeicher und eine Antwort, die von der Laune der
	// Laufzeit abhängt.
	var adressiert []string

	nachFrage := map[string]sammelliste.Zeile{}

	for _, zeile := range zeilen {
		if !istAdressiert(zeile) {
			continue
		}

		frage := frageFuer(zeile)
		if _, schon := nachFrage[frage]; schon {
			continue
		}

		nachFrage[frage] = zeile
		adressiert = append(adressiert, frage)
	}

	return l.sammelnd(ctx, adressiert, nachFrage)
}

// zeileDeuten fragt Scryfall nach den Drucken einer Zeile.
//
// **Ein Fehler wird zum Befund und hält die Liste nicht auf** (§8): wer eine
// Liste von 300 Zeilen einliest und bei zweien kein Netz hat, bekommt 298
// gedeutete Zeilen und zwei benannte.
func (l *Listeneinfuhr) zeileDeuten(ctx context.Context, frage string) (
	Zeilendeutung, time.Duration,
) {
	treffer, alter, befunde, fehler := l.suche.Druckesuche(ctx, frage)

	// **Kein Druck ist eine Antwort und kein Ausfall**: Scryfall sagt es mit
	// einem 404, und die Zeile steht dann wie eine mit null Treffern da.
	if errors.Is(fehler, domain.ErrNotFound) {
		return Zeilendeutung{Befund: "kein Druck heißt so"}, 0
	}

	if fehler != nil {
		return Zeilendeutung{Befund: fehler.Error(), Fehler: fehler}, 0
	}

	gedeutet := Zeilendeutung{
		Weitere: treffer.Weitere,
		Drucke:  len(treffer.Karten),
		Treffer: make([]Suchtreffer, 0, min(len(treffer.Karten), Deutungsgrenze)),
	}

	for _, karte := range treffer.Karten[:min(len(treffer.Karten), Deutungsgrenze)] {
		gedeutet.Treffer = append(gedeutet.Treffer, alsSuchtreffer(karte))
	}

	// **`has_more` heißt: es sind mehr, als eine Seite fasst.** Dann ist
	// `Drucke` eine Untergrenze und keine Zahl — und das gehört gesagt.
	if treffer.Weitere {
		gedeutet.Befund = fmt.Sprintf(
			"mehr als %d Drucke heißen so — nenne das Set, um die Auswahl zu "+
				"schließen", gedeutet.Drucke)
	}

	if gedeutet.Drucke == 0 {
		gedeutet.Befund = "kein Druck heißt so"
	}

	// **Die Befunde der Quelle gehen mit** — Befund 10 des Zweitblicks über
	// Phase 4.5.
	//
	// `internal/scryfall` meldet einen unbekannten Aufzählungswert bewusst als
	// **Befund** und liefert die Karte trotzdem (§16: Pflichtfelder streng,
	// Zusatzfelder tolerant).
	//
	// **Sie hängen sich an, statt zu ersetzen**: ein „kein Druck heißt so"
	// bleibt stehen, und was Scryfall daneben zu sagen hat, kommt dazu.
	gedeutet.Befund = anhaengen(gedeutet.Befund, befunde)

	return gedeutet, alter
}

// frageFuer baut die Scryfall-Frage zu einer Zeile.
//
// **Die Syntax ist Scryfalls eigene und nachgelesen** (§16, Regel 1) — der
// Kopf dieser Datei nennt die drei Abrufe vom 2026-08-30. `!"…"` ist der
// **genaue** Name; ein unscharfer Treffer machte aus „Bolt" eine andere Karte.
//
// **`e:` und nicht `set:`**, wie in `scryfall/setkarten.go`: beides geht, `e`
// ist die kurze Form.
func frageFuer(zeile sammelliste.Zeile) string {
	// **Anführungszeichen im Namen werden entfernt, nicht maskiert.** Scryfall
	// kennt keine Fluchtsequenz in `!"…"`; ein Name mit einem Zitatzeichen
	// bräche die Frage auf. Kartennamen tragen Apostrophe, keine
	// Anführungszeichen — was hier wegfällt, ist ein Tippfehler.
	frage := `!"` + strings.ReplaceAll(zeile.Name, `"`, "") + `"`

	if zeile.Setcode != "" {
		frage += " e:" + zeile.Setcode
	}

	if zeile.Nummer != "" {
		frage += " cn:" + zeile.Nummer
	}

	return frage
}

// anhaengen fügt die Befunde der Quelle an einen vorhandenen Satz.
//
// **Ohne Trenner am Anfang**: die meisten Zeilen haben keinen eigenen Befund,
// und ein führendes „ · " sähe aus wie ein fehlendes Wort.
func anhaengen(vorhanden string, befunde []domain.Befund) string {
	if len(befunde) == 0 {
		return vorhanden
	}

	saetze := make([]string, 0, len(befunde)+1)
	if vorhanden != "" {
		saetze = append(saetze, vorhanden)
	}

	for _, befund := range befunde {
		saetze = append(saetze, befund.String())
	}

	return strings.Join(saetze, " · ")
}
