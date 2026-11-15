package app

// **Die lokale KI-Laufzeit an der Naht** — Werkstück A7 von Phase 5.3
// (#1446).
//
// # Was die Einrichtung unterscheidet
//
// **Fünf Stände, und sie sind nicht dasselbe.** Das Issue verlangt es
// ausdrücklich: „Einrichtungsansicht unterscheidet Modell fehlt, Fähigkeit
// fehlt, Herkunft unbekannt, Version inkompatibel und Indexaufbau nötig."
//
//	| Stand            | Heißt                                                |
//	| ---------------- | ---------------------------------------------------- |
//	| `aus`            | niemand hat eine Laufzeit eingeschaltet               |
//	| `unerreichbar`   | eingeschaltet, und es antwortet nichts                |
//	| `ohnemodell`     | sie antwortet, und es ist keines installiert          |
//	| `ohnefaehigkeit` | Modelle, aber keines mit einer **belegten** Fähigkeit |
//	| `bereit`         | mindestens eines kann etwas                           |
//
// **Herkunft unbekannt steht am Modell** und nicht hier: sie ist eine
// Eigenschaft eines Eintrags, kein Zustand der Einrichtung. Ein Modell ohne
// Kette trägt eine leere `herkunft` und einen Befund daneben.
//
// # Was hier **nicht** geprüft wird
//
// **Eine Mindestfassung von Ollama.** Sie zu prüfen hieße, eine Grenze zu
// erfinden: die Dokumentation nennt keine — „Ollama's API isn't strictly
// versioned, but the API is expected to be stable and backwards compatible"
// (**nachgelesen am 2026-09-11**). Eine Zahl, die dieses Programm sich
// ausdächte, wiese morgen eine Installation ab, die tut, was sie soll (§10,
// „auch eine Ablehnung braucht Deckung"). **Die Fassung steht trotzdem in der
// Antwort** — sie gehört zu jedem Beleg, und wer eine Unverträglichkeit
// sieht, kann sie nennen.
//
// **Und ein Indexaufbau.** Der gehört A15: hier entsteht der Fingerabdruck,
// dort wird entschieden, was aus einem Wechsel folgt.

import (
	"context"
	"fmt"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/ollama"
)

// KILaufzeitquelle ist, was diese Naht von einer Laufzeit braucht.
//
// **Zwei Methoden**, und beide lesen. Eine Naht, die mehr verlangte, könnte
// ein Blatt nicht mehr austauschen — A15 darf den Adapter über den
// A6-Vertrag ersetzen, und das Issue sagt es ausdrücklich.
type KILaufzeitquelle interface {
	Fassung(ctx context.Context) (string, error)
	Modelle(ctx context.Context) ([]domain.KIModell, []domain.Befund, error)

	// Vorschlagen fragt ein Modell und zwingt die Antwort in ein Schema.
	//
	// **Mit Schema und nicht mit einer Bitte darum**: die Laufzeit hält die
	// Form, statt sie zu erhoffen — und was zurückkommt, ist entweder
	// gültiges JSON oder ein Fehler.
	//
	// # Warum die Anweisung ein Parameter ist
	//
	// **Weil es seit A13.2 zwei Aufgaben gibt und nicht eine.** Sie stand als
	// Konstante in jeder Naht — „Du füllst Felder einer Sammelkarte" —, und
	// eine Einschätzung, die auf dieselbe Zeile trifft, bekäme eine
	// Anweisung, die von etwas anderem spricht. Eine Systemzeile, die nicht
	// zur Aufgabe passt, ist kein Formfehler: sie ist die erste Zeile, die
	// das Modell liest.
	Vorschlagen(ctx context.Context, anweisung, alias, auftrag string, schema map[string]any) (
		KIRohantwort, error)
}

// Die fünf Einrichtungsstände.
const (
	KIEinrichtungAus            = "aus"
	KIEinrichtungUnerreichbar   = "unerreichbar"
	KIEinrichtungOhneModell     = "ohnemodell"
	KIEinrichtungOhneFaehigkeit = "ohnefaehigkeit"
	KIEinrichtungBereit         = "bereit"
)

// kiVerdrahten hängt die lokale Laufzeit an die Anwendung.
//
// **Ein Bauplan und kein fertiger Klient**: die Adresse steht in den
// Einstellungen, und die ändert sich, während das Fenster offen ist. Ein
// einmal gebauter Klient zeigte nach einer Änderung weiter auf den alten
// Port, ohne dass jemand es sähe.
//
// **Gefragt wird trotzdem erst, wenn jemand einschaltet** — die Vorgabe von
// `ki.laufzeit` ist `aus`, und dann wird dieser Bauplan gar nicht gerufen
// ([App.frageAnbieter]).
func (a *App) kiVerdrahten() {
	a.kilaufzeit = a.ollamalaufzeit
	a.kicloud = a.cloudnaehte()
}

// MitKILaufzeit setzt, wie eine Laufzeit unter einer Adresse entsteht.
//
// **Als Bauplan und nicht als fertiger Klient**: die Adresse steht in den
// Einstellungen und ändert sich, während das Fenster offen ist.
func (a *App) MitKILaufzeit(bau func(adresse string) KILaufzeitquelle) *App {
	a.kilaufzeit = bau

	return a
}

// kieinstellungen liest Schalter und Adresse aus dem wirksamen Stand.
func (a *App) kieinstellungen(ctx context.Context) (an bool, adresse string, fehler error) {
	if a.einstellungen == nil {
		return false, "", fehlenderBestand()
	}

	stand, lesefehler := a.einstellungen.Stand(ctx)
	if lesefehler != nil {
		return false, "", lesefehler
	}

	return stand[domain.KILaufzeit].Wirksam == domain.KILaufzeitAn,
		stand[domain.KIAdresse].Wirksam, nil
}

// laufzeitstand fragt die Laufzeit — oder sagt, warum nicht.
type laufzeitstand struct {
	einrichtung string
	fassung     string
	adresse     string
	modelle     []domain.KIModell
	befunde     []domain.Befund
}

// frageQuelle holt Fassung und Modelle — oder sagt, warum nicht.
//
// **Ohne Fehlerrückgabe, mit Absicht**: jeder Ausgang ist ein Stand. Ein
// `error` daneben wäre die zweite Auskunft über dieselbe Sache, und der
// Aufrufer müsste beide lesen.
func (a *App) frageQuelle(
	ctx context.Context, quelle KILaufzeitquelle, adresse string,
) laufzeitstand {
	stand := laufzeitstand{adresse: adresse}

	fassung, fassungsfehler := quelle.Fassung(ctx)
	if fassungsfehler != nil {
		stand.einrichtung = KIEinrichtungUnerreichbar
		stand.befunde = []domain.Befund{{
			Art: domain.ArtUmgebung,
			Meldung: "Unter " + angezeigteAdresse(adresse) + " antwortet " +
				"keine Laufzeit: " + fassungsfehler.Error(),
		}}

		return stand
	}

	stand.fassung = fassung

	modelle, befunde, modellfehler := quelle.Modelle(ctx)
	stand.befunde = befunde

	if modellfehler != nil {
		stand.einrichtung = KIEinrichtungUnerreichbar
		stand.befunde = append(stand.befunde, domain.Befund{
			Art:     domain.ArtUmgebung,
			Meldung: "Die Modelle waren nicht zu lesen: " + modellfehler.Error(),
		})

		return stand
	}

	stand.modelle = modelle
	stand.einrichtung = einrichtungAus(modelle)

	return stand
}

// angezeigteAdresse nennt die Adresse, auf die gefragt wurde.
func angezeigteAdresse(adresse string) string {
	if adresse == "" {
		return "der Vorgabeadresse"
	}

	return adresse
}

// einrichtungAus sagt, was aus den gefundenen Modellen folgt.
func einrichtungAus(modelle []domain.KIModell) string {
	if len(modelle) == 0 {
		return KIEinrichtungOhneModell
	}

	for _, modell := range modelle {
		if len(modell.Geprueft) > 0 {
			return KIEinrichtungBereit
		}
	}

	return KIEinrichtungOhneFaehigkeit
}

// ollamalaufzeit baut die Laufzeit für eine Adresse.
//
// **Die einzige Stelle, an der dieses Paket `internal/ollama` erzeugt** — wie
// jede andere Naht in `verdrahtung.go`. Was sie kann, steht dort; was sie
// **nicht** kann, im Kopf jenes Pakets.
func (a *App) ollamalaufzeit(adresse string) KILaufzeitquelle {
	return ollamaquelle{Klient: ollama.NeuerKlient(adresse, a.netzklientMit(
		netz.Fristen{
			Verbindung: kiverbindungsfrist,
			Kopf:       kikopffrist,
			Stillstand: kistillstandsfrist,
		}).MitErlaubnis(func(ctx context.Context) error {
		return a.pruefeKILaufzeit(ctx, adresse)
	}))}
}

// pruefeKILaufzeit fragt **vor jeder Anfrage**, ob sie hinausgehen darf.
//
// # Warum am Transport und nicht nur an der Naht
//
// **Weil ein Schalter sich ändert, nachdem die Naht gebaut ist.** Dieselbe
// Überlegung wie bei den Dienstschaltern (`dienstsuchnaht.go`): wer
// `ki.laufzeit` auf `aus` legt, während ein Vorschlag läuft, hat den Ruf
// schon losgeschickt. Die Sperre am Transport fängt den nächsten.
//
// **Und sie fängt mehr als der Vorfilter**: [App.frageAnbieter] sieht einmal
// nach, dieser Riegel bei **jeder** Anfrage — auch bei denen, die ein
// Vorschlag Minuten später noch stellt.
func (a *App) pruefeKILaufzeit(ctx context.Context, adresse string) error {
	an, _, fehler := a.kieinstellungen(ctx)
	if fehler != nil {
		return fehler
	}

	if !an {
		return fmt.Errorf("%w: die KI-Laufzeit ist abgeschaltet",
			domain.ErrErlaubnis)
	}

	// **Und »nur lokal« hält eine fremde Adresse auf.** Die Regel steht an
	// den Modellen (`KIOrtsregel`) — dort entscheidet sie, **was** gefragt
	// werden darf; hier entscheidet sie, ob überhaupt etwas hinausgeht. Ohne
	// diese Zeile bliebe `/api/tags` an einen fremden Rechner erlaubt, und
	// die Liste der dortigen Modelle wäre schon eine Auskunft nach draußen.
	regel, fehler := a.ortsregel(ctx)
	if fehler != nil {
		return fehler
	}

	if regel.NurLokal && domain.KIOrtFuerAdresse(adresse) != domain.KIOrtLokal {
		return fmt.Errorf(
			"%w: hier gilt »nur lokal«, und %s liegt nicht auf diesem Rechner",
			domain.ErrErlaubnis, adresse)
	}

	return nil
}

// ollamaquelle bringt den Klienten auf die Gestalt, die diese Naht braucht.
//
// **Eine Hülle und keine zweite Fachlogik**: `Fassung` und `Modelle` kommen
// unverändert aus dem eingebetteten Klienten, und `Vorschlagen` übersetzt nur
// die Aufschrift — der Ruf selbst steht in `internal/ollama`.
type ollamaquelle struct{ *ollama.Klient }

// Vorschlagen fragt das Modell mit fester Form.
func (o ollamaquelle) Vorschlagen(
	ctx context.Context, anweisung, alias, auftrag string, schema map[string]any,
) (KIRohantwort, error) {
	antwort, fehler := o.Rede(ctx, ollama.Frage{
		Alias: alias, System: anweisung, Auftrag: auftrag, Schema: schema,
	})
	if fehler != nil {
		return KIRohantwort{}, fehler
	}

	return KIRohantwort{
		Inhalt:        antwort.Inhalt,
		Abgeschnitten: antwort.Abgeschnitten(),
		Modell:        antwort.Modell,
	}, nil
}

// KIKartenanweisung steht vor jedem Auftrag, der Kartenfelder füllt.
//
// **Sie sagt, was das Ergebnis ist, und nicht, was wahr ist.** Über den Inhalt
// entscheidet ein Mensch — das Schema prüft „Form/Felder, niemals
// Wahrheitsgehalt oder Spielbalance" (A6), und diese Anweisung tut dasselbe.
//
// **Auf Deutsch, weil die Ausgangssprache deutsch ist** (§9). Was sie
// erzeugt, sind Feldwerte eines Pakets — und die werden nicht übersetzt.
//
// **Sie heißt seit A13.2 nach ihrer Aufgabe** (#1452): sie hieß
// `kianweisung`, als es nur eine gab, und ein Name ohne Aufgabe verleitet
// dazu, die zweite Aufgabe mit derselben Zeile zu fahren.
const KIKartenanweisung = "Du füllst Felder einer Sammelkarte. Antworte " +
	"ausschließlich mit den Feldern, zu denen du etwas sagen kannst; lass " +
	"weg, was du nicht weißt. Erfinde keine Feldnamen."

// Die Fristen der KI-Laufzeit.
//
// # Warum die Kopffrist so lang ist
//
// **Weil das Modell geladen wird, bevor die Statuszeile kommt.** Ollama legt
// eine Datei von mehreren Gigabyte in den Speicher und antwortet erst danach;
// die üblichen dreißig Sekunden ([netz.UeblicheFristen]) träfen jeden ersten
// Ruf nach dem Start. Das Issue nennt es ausdrücklich: „Abbruch und
// Größen-/Zeit-/Parallelgrenzen bei **langsam ladendem Modell** greifen."
//
// **Und die Verbindung bleibt kurz.** Ob auf einem Port jemand lauscht, steht
// sofort fest — dafür ist Ladezeit kein Grund. Eine Laufzeit, die es nicht
// gibt, soll die Einstellungsseite nicht eine halbe Minute anhalten.
//
// **Eine Frist gibt es trotzdem** (§16, Satz 3): ein Ruf ohne sie hielte das
// Fenster für immer, und **der Abbruch kommt vom Kontext** — wer die Seite
// wechselt, bricht ab, statt zu warten.
const (
	kiverbindungsfrist = 5 * time.Second
	kikopffrist        = 5 * time.Minute
	kistillstandsfrist = 60 * time.Second
)
