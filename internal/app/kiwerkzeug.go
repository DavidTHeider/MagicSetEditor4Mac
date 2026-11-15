package app

// **Woher das Auftragswerk kommt** — Werkstück A7 von Phase 5.3
// (#1446).
//
// # Eines je Fenster und nicht eines je Ruf
//
// **Weil ein Vorschlag zwei Rufe überlebt.** [service.KIAuftragswerk] hält
// seine offenen Aufträge im Prozess (A6, und dort steht warum); ein Werk je
// Ruf hätte beim Übernehmen keinen einzigen mehr. Es entsteht deshalb einmal
// und wird gemerkt.
//
// # Und der Katalog wird bei jedem Ruf neu geholt
//
// **Weil sich die Laufzeit ändert, während das Fenster offen ist.** Wer ein
// Modell löscht oder eines nachlädt, soll es beim nächsten Vorschlag merken —
// und nicht beim nächsten Start.

import (
	"context"
	"fmt"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// kinaehte hält, was die Anwendung über KI-Laufzeiten weiß.
//
// **Als eigener Verbund und nicht als fünf Felder in `anwendung.go`** — wie
// `leutenaehte` und `fremddienste` daneben, und aus demselben Grund: jene
// Datei stand an der Zeilengrenze aus §2.
type kinaehte struct {
	// geraet ist Apples Gerätemodell — `kigeraetnaht.go` (AI-05, #2042);
	// nil heißt: die Naht des Prozesses.
	geraet *geraetenaht

	// kimodelle sind KI-Modelle, die ohne Laufzeit feststehen (A6, #1445).
	//
	// **Seit A7 (#1446) steht hier im Betrieb nichts**: die Modelle kommen
	// aus der Laufzeit. Das Feld bleibt für Läufe, die einen Katalog ohne
	// Gegenstelle brauchen — und für einen künftigen zweiten Anbieter, der
	// keine Laufzeit befragt.
	kimodelle []domain.KIModell

	// kilaufzeit baut die Laufzeit unter einer Adresse — nil heißt: keine.
	//
	// **Ein Bauplan und kein fertiger Klient**: die Adresse steht in den
	// Einstellungen und ändert sich, während das Fenster offen ist.
	kilaufzeit func(adresse string) KILaufzeitquelle

	// kizuhause ist das Zuhause, unter dem `.ollama/server.json` liegt —
	// leer heißt: das echte. Warum als Feld: [App.MitKIZuhause].
	kizuhause string

	// kinachweis und kigrenze sind die Freigabe für den nächsten
	// kostenpflichtigen Auftrag — leer heißt: keine (A8, #1447).
	//
	// **Je Ruf und nicht je Sitzung**: eine Freigabe gilt dem, was ein
	// Mensch gerade gelesen hat ([App.MitKIFreigabe]).
	kinachweis string
	kigrenze   domain.Betrag

	// kicloud hält je Cloudanbieter seinen Bauplan — ein fehlender
	// Eintrag heißt: dieser Anbieter ist nicht verdrahtet.
	//
	// **Ohne Adresse**, anders als die lokale Laufzeit: die Schnittstelle
	// eines Cloudanbieters liegt, wo sie liegt. Was sich einstellen lässt,
	// ist der Schlüssel, und der kommt aus dem Schlüsselbund (A8, #1447).
	//
	// **Eine Karte und nicht je Anbieter ein Feld** (A10, #1449). Mit zweien
	// ging beides; mit dem dritten stünden drei Felder desselben Typs
	// nebeneinander, dazu je eine Bauplanfunktion und je eine Zeile in
	// `MitKIAnbietern` — sechs Stellen für eine Zeile Inhalt.
	kicloud map[string]func(adresse string) KILaufzeitquelle

	// kiwerk hält die offenen KI-Aufträge dieses Fensters — nil, bis der
	// erste entsteht. Warum eines je Fenster: [App.kiwerkzeug].
	//
	// **Hinter einem Schloss**, seit dem Zweitblick über Phase 5.3 (Befund 1):
	// Wails bedient jede Bindung in einer eigenen Goroutine, und sechs von
	// ihnen bauen das Werk neu. Lesen und Schreiben dieses Zeigers ohne
	// Schloss sind ein Datenrennen — der Verlust der offenen Aufträge ist ein
	// zweiter Fehler und in [kibuch] behoben.
	kiwerk        *service.KIAuftragswerk
	kiwerkschloss sync.Mutex

	// kinahtfehler hält, woran der Bau des KI-Schreibwegs scheiterte — nil,
	// wenn er steht. Warum gemerkt statt verschluckt: [KIAnwendungUeber].
	kinahtfehler error
}

// ollamaanbieter ist der Name, unter dem die Laufzeit im Katalog steht.
//
// **Hier und nicht aus `internal/ollama` gelesen**: die Naht kennt einen
// Anbieternamen, und ein zweiter Anbieter trüge einen anderen — die Bindung
// an genau dieses Blatt steht in `kilaufzeit.go` und nur dort.
const ollamaanbieter = "ollama"

// anthropicanbieter ist der Name, unter dem die Modelle von Anthropic im
// Katalog stehen.
//
// **Hier und nicht aus `internal/anthropic` gelesen**, aus demselben Grund
// wie bei der Zeile darüber: die Naht kennt Anbieternamen, und die Bindung
// an genau dieses Blatt steht in `kianthropicnaht.go` und nur dort.
const anthropicanbieter = "anthropic"

// openaianbieter ist der Name, unter dem die Modelle von OpenAI im Katalog
// stehen.
//
// **Dieselbe Begründung wie zwei Zeilen darüber**: die Bindung an genau
// dieses Blatt steht in `kiopenainaht.go` und nur dort.
const openaianbieter = "openai"

// xaianbieter ist der Name, unter dem die Modelle von xAI im Katalog stehen.
//
// **Dieselbe Begründung wie zwei Zeilen darüber**: die Bindung an genau
// dieses Blatt steht in `kixainaht.go` und nur dort.
const xaianbieter = "xai"

// kiwerkzeug gibt das Auftragswerk und die Quelle für einen Anbieter.
//
// **Mit dem Anbieter als Argument seit A8** (#1447): welcher gefragt wird,
// entscheidet der Alias — und die Quelle muss dieselbe sein, die das Modell
// in den Katalog gebracht hat.
func (a *App) kiwerkzeug(
	ctx context.Context, anbieter string, deckung kideckung,
) (*service.KIAuftragswerk, KILaufzeitquelle, error) {
	// **Der gemerkte Nahtfehler zuerst** — sonst käme statt seiner die
	// Meldung „kein Bestand offen", und die wäre falsch (§8).
	if a.kinahtfehler != nil {
		return nil, nil, a.kinahtfehler
	}

	staende, _, fehler := a.frageAnbieter(ctx)
	if fehler != nil {
		return nil, nil, fehler
	}

	return a.werkzeugAus(ctx, staende, anbieter, deckung)
}

// werkzeugAus baut das Werk aus schon geholten Ständen.
//
// **Getrennt, damit ein Lauf fertige Stände einsetzen kann** — die
// Nullprüfungen darunter fangen Zustände ab, die über [App.kiwerkzeug] nicht
// herzustellen sind.
func (a *App) werkzeugAus(
	ctx context.Context, staende []anbieterstand, anbieter string,
	deckung kideckung,
) (*service.KIAuftragswerk, KILaufzeitquelle, error) {
	quelle, adresse, fehler := a.quelleFuer(ctx, staende, anbieter)
	if fehler != nil {
		return nil, nil, fehler
	}

	regel, fehler := a.ortsregel(ctx)
	if fehler != nil {
		return nil, nil, fehler
	}

	katalog, fehler := service.NeuerModellkatalog(alleModelle(staende)...)
	if fehler != nil {
		return nil, nil, fehler
	}

	// **Das Budget richtet sich nach dem Ort des Anbieters** — lokal kostet
	// nichts, alles andere braucht einen Weg, den es noch nicht gibt
	// ([kibudget]).
	budget, fehler := a.budgetFuer(ctx, anbieter, adresse, deckung)
	if fehler != nil {
		return nil, nil, fehler
	}

	// **Nehmen, bauen und einsetzen unter einem Schloss** — sonst läse der
	// eine Ruf den Zeiger, während der andere ihn schreibt.
	a.kiwerkschloss.Lock()
	defer a.kiwerkschloss.Unlock()

	a.kiwerk = service.NeuesKIAuftragswerk(katalog, kikarten{app: a}, budget,
		kischemaquelle{app: a}, regel, a.uhr).MitOffenen(a.kiwerk)

	return a.kiwerk, quelle, nil
}

// kiwerkRaeumen wirft das Auftragswerk samt seinen offenen Vorschlägen weg.
//
// **Ein Wechsel des Bezugsobjekts räumt den offenen Kontext** (§7, Regel 1) —
// und ein Profilwechsel ist einer. Ohne diese Zeile stünde ein in Profil A
// angelegter Vorschlag in Profil B weiter offen; die Grenze am Auftrag wiese
// ihn zwar ab ([service.KIAuftragswerk.Uebernehmen]), aber ein Vorschlag, der
// nie wieder einzulösen ist, gehört nicht in ein Buch, sondern weg.
//
// **Zwei Zusagen und nicht eine**: die Grenze hält, was durchkommt, das
// Räumen hält, was liegen bleibt. Befund 1 des Nachlaufs über Phase 5.3
// nannte beide.
func (a *App) kiwerkRaeumen() {
	a.kiwerkschloss.Lock()
	defer a.kiwerkschloss.Unlock()

	a.kiwerk = nil
}

// quelleFuer sucht die Quelle eines Anbieters — und sagt, warum nicht.
func (a *App) quelleFuer(
	ctx context.Context, staende []anbieterstand, anbieter string,
) (KILaufzeitquelle, string, error) {
	alle, fehler := a.kianbieter(ctx)
	if fehler != nil {
		return nil, "", fehler
	}

	for _, einer := range alle {
		if einer.name != anbieter {
			continue
		}

		if !einer.eingeschaltet {
			return nil, "", abgeschaltetWeil(einer)
		}

		if einer.bauen == nil {
			return nil, "", fehlenderBestand()
		}

		return einer.bauen(), einer.adresse, nil
	}

	_ = staende

	return nil, "", fmt.Errorf("%w: den Anbieter %q gibt es nicht",
		domain.ErrNotFound, anbieter)
}

// abgeschaltetWeil sagt, **wer** einen Anbieter abgeschaltet hat.
//
// # Warum das zwei Meldungen sind und nicht eine
//
// **Weil sie an zwei verschiedene Schalter schicken** (§8). Wer selbst
// abgeschaltet hat, legt `dienst.<name>` um; unter `richtlinie.nurlokal`
// steht derselbe Schalter auf `an` und gilt trotzdem nicht — wer dort die
// Meldung „ist nicht eingeschaltet" liest, legt ihn um und kommt nicht
// weiter.
//
// # Der Satz stand schon da, eine Schicht tiefer
//
// **[App.pruefeCloudanbieter] unterscheidet die beiden Fälle seit A8** und
// sagt es im Kommentar ausdrücklich: „wer nur den Schalter läse, meldete
// »Anthropic ist abgeschaltet«. Das schickte jemanden an den falschen
// Schalter." **Nur kam der Riegel nie dran**: die Wirksamkeitsschranke aus
// [domain] liegt davor, `quelleFuer` liest den **wirksamen** Wert, und die
// Absage fiel hier. Gefunden vom Flächenlauf zur Null-Netz-Probe (A14,
// #1453).
//
// **[kianbieter.gewaehlt] gibt es genau dafür** — sein Kommentar sagt es seit
// A8, und der Katalog benutzt es in [vonDerRichtlinieAbgeschaltet]. Hier
// fehlte es.
func abgeschaltetWeil(einer kianbieter) error {
	// **Ein Grund, den der Anbieter selbst nennt, geht vor** — beim
	// Gerätemodell: was auf diesem Mac fehlt (AI-05.AK-03).
	if einer.hinweis != "" {
		return fmt.Errorf("%w: %s", domain.ErrValidation, einer.hinweis)
	}

	if einer.gewaehlt {
		return fmt.Errorf(
			"%w: hier gilt »nur lokal«, und %q rechnet nicht auf diesem "+
				"Rechner", domain.ErrErlaubnis, einer.name)
	}

	return fmt.Errorf("%w: %q ist nicht eingeschaltet",
		domain.ErrValidation, einer.name)
}

// ortVon sagt, wo ein Anbieter rechnet.
//
// **Nur die lokale Laufzeit hat eine Adresse, über die das zu entscheiden
// ist.** Das Gerätemodell rechnet immer hier — es spricht über einen Socket
// dieses Prozesses (AI-05, #2042). Jeder andere rechnet dort, wo er steht —
// und das ist nicht hier.
func ortVon(anbieter, adresse string) domain.KIOrt {
	switch anbieter {
	case geraeteanbieter:
		return domain.KIOrtLokal
	case ollamaanbieter:
		return domain.KIOrtFuerAdresse(adresse)
	default:
		return domain.KIOrtCloud
	}
}

// kiwerkOhneQuelle gibt das Auftragswerk ohne einen Anbieter zu bauen.
//
// # Wofür es das gibt
//
// **Zum Übernehmen braucht es keine Quelle.** Der Vorschlag liegt schon da;
// was noch geschieht, ist eine Prüfung gegen den Stand der Karte und ein
// Schreiben. Einen Anbieter dafür zu verlangen hieße, dass eine Übernahme
// scheitert, weil jemand inzwischen einen Schalter umgelegt hat — und was
// geprüft wurde, war schon geprüft.
func (a *App) kiwerkOhneQuelle(_ context.Context) (
	*service.KIAuftragswerk, error,
) {
	if a.kinahtfehler != nil {
		return nil, a.kinahtfehler
	}

	a.kiwerkschloss.Lock()
	werk := a.kiwerk
	a.kiwerkschloss.Unlock()

	if werk != nil {
		return werk, nil
	}

	// **Ohne offenes Werk gibt es nichts zu übernehmen** — und die Meldung
	// sagt das, statt einen Anbieter zu verlangen, der nichts damit zu tun
	// hat.
	return nil, fmt.Errorf(
		"%w: es liegt kein Vorschlag vor, der zu übernehmen wäre",
		domain.ErrNotFound)
}
