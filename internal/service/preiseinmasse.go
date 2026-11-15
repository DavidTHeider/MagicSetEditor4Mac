package service

// **Preise für viele Drucke** — Werkstück 6 von Phase 4.5
// (#1292).
//
// # Der Befund
//
// `Holen` fragt **einen** Druck. Wer bei fünfhundert Drucken den Wert seiner
// Sammlung wissen will, ruft fünfhundertmal — und wartet dabei, ohne zu sehen,
// wie weit es ist.
//
// # Warum das in Schüben geht und nicht in einem Zug
//
// **Wegen der Drossel** (§16, Regel 4). Die Zahl steht in
// `internal/scryfall/klient.go`: die **Kartenwege** dürfen 2/s, und dieses
// Programm wartet 600 ms — also rund 1,7 Anfragen je Sekunde.
//
// Fünfhundert Drucke sind damit **rund fünf Minuten**, ein Schub von zwanzig
// **rund zwölf Sekunden**. Ein Aufruf, der die fünf Minuten am Stück täte,
// wäre fünf Minuten ohne Lebenszeichen.
//
// **Deshalb ein Schub und eine Zahl, wie viele bleiben.** Der Aufrufer ruft
// wieder, bis nichts mehr bleibt, und **er** zeigt den Fortschritt — an jeder
// Fläche mit ihren eigenen Mitteln. Das ist derselbe Schnitt wie bei der
// Blätterung: wer eine Schleife im Dienst versteckt, versteckt auch ihre
// Dauer.
//
// # Warum nur, was fehlt oder alt ist
//
// **Weil ein frischer Preis kein zweites Mal geholt werden muss.** Ohne diese
// Auswahl fragte jeder Lauf alles neu — und die Drossel machte daraus jedes
// Mal dieselbe Minute.

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Schubgrenze ist die Zahl der Drucke, die ein Schub höchstens **versucht**.
//
// **Zwanzig, und gezählt werden Versuche.** Bei 600 ms Abstand sind das rund
// zwölf Sekunden — lang genug, dass der Fortschritt nicht aus Einzelrufen
// besteht, und kurz genug, dass zwischen zwei Schüben etwas anderes geschehen
// kann.
//
// # Warum Versuche und nicht Erfolge
//
// Hier stand `schub.Geholt >= Schubgrenze`, und ein Fehlschlag zählte nicht
// mit. **Scheiterten alle fünfhundert Drucke, rief ein einziger „Schub" alle
// fünfhundert** — genau die Wartezeit ohne Lebenszeichen, gegen die es die
// Schübe gibt. Befund 7 des Zweitblicks über Phase 4.5.
const Schubgrenze = 20

// Preisschub sagt, was ein Schub geholt hat und was bleibt.
type Preisschub struct {
	// Geholt ist die Zahl der Drucke, für die Preise eingetragen wurden.
	Geholt int

	// Offen ist die Zahl der Drucke, die **danach** noch fehlen.
	//
	// **Sie steht da, damit der Aufrufer den Fortschritt zeigen kann** — und
	// damit er weiß, wann er aufhören darf.
	//
	// **Ein Fehlschlag vermindert sie nicht.** Er tat es, mit einem richtig
	// gesehenen Grund: „ein gescheiterter Druck bleibt sonst für immer offen,
	// und der Aufrufer riefe endlos weiter." Nur rechnet der nächste Ruf die
	// offenen Drucke aus der **unveränderten** Preisablage neu — die
	// gescheiterten stehen dann wieder da, und bei genau [Schubgrenze]
	// Fehlschlägen meldete dieser Schub `Fertig`, obwohl kein Preis abgelegt
	// wurde. Nachlauf über Phase 4.5, Befund 7.
	Offen int

	// Fehlgeschlagen ist die Zahl der Versuche, die nichts eingetragen haben.
	//
	// **Sie ist die Abhilfe für das, was das Vermindern lösen sollte**: wer
	// `Geholt: 0` neben `Fehlgeschlagen: 20` sieht, weiß, dass ein weiterer
	// Ruf dasselbe gibt — ohne dass eine Zahl dafür lügen muss.
	Fehlgeschlagen int

	Befunde []domain.Befund
}

// Fertig sagt, ob nichts mehr bleibt.
func (p Preisschub) Fertig() bool { return p.Offen == 0 }

// Steht sagt, dass dieser Schub nichts bewegt hat, obwohl es etwas zu tun gab.
//
// **Der Satz, den ein Aufrufer braucht, um aufzuhören.** `Fertig` ist falsch —
// es bleibt Arbeit —, und ein weiterer Ruf gäbe dasselbe. Wer schleifenweise
// ruft, hält hier an und zeigt die Befunde.
func (p Preisschub) Steht() bool { return p.Geholt == 0 && p.Fehlgeschlagen > 0 }

// PreiseInMasse holt einen Schub Preise für die Drucke der Sammlung.
//
// **`hoechstensAlter` sagt, was noch frisch genug ist.** Ein Preis, der jünger
// ist, wird übergangen; null heißt „alles neu holen".
func (p *Preispflege) PreiseInMasse(
	ctx context.Context, hoechstensAlter time.Duration,
) (Preisschub, error) {
	freigeben, platzfehler := p.platz(ctx, domain.RessourceMasse)
	if platzfehler != nil {
		return Preisschub{}, platzfehler
	}

	defer freigeben()

	exemplare, err := p.sammlung.Liste(ctx)
	if err != nil {
		return Preisschub{}, err
	}

	alle, err := p.ablage.NeuesteAlle(ctx)
	if err != nil {
		return Preisschub{}, err
	}

	juengster := map[string]time.Time{}
	for _, stand := range alle {
		if stand.Erhoben.After(juengster[stand.Printing]) {
			juengster[stand.Printing] = stand.Erhoben
		}
	}

	offene := p.offeneDrucke(exemplare, juengster, hoechstensAlter)

	schub := Preisschub{Offen: len(offene), Befunde: []domain.Befund{}}

	// **Versuche, nicht Erfolge** (Befund 7). Ein Fehlschlag kostet dieselbe
	// Wartezeit wie ein Treffer; wer ihn nicht zählt, hat keine Grenze.
	versuche := 0

	for _, printing := range offene {
		if versuche >= Schubgrenze {
			break
		}

		versuche++

		_, _, befunde, fehler := p.Holen(ctx, printing)

		// **Ein Ausfall hält den Schub nicht auf** (§8): wer bei Druck 7 kein
		// Netz hat, soll die übrigen trotzdem bekommen.
		//
		// **Und er bleibt offen**, denn er ist es: der Preis wurde nicht
		// abgelegt, und der nächste Ruf findet ihn wieder. Gezählt wird er als
		// `Fehlgeschlagen` — das ist die Auskunft, die der Aufrufer braucht,
		// ohne dass `Offen` dafür lügt (Nachlauf über Phase 4.5, Befund 7).
		if fehler != nil {
			schub.Befunde = append(schub.Befunde, domain.Befund{
				Ort: printing, Meldung: fehler.Error(), Art: domain.ArtUmgebung,
			})
			schub.Fehlgeschlagen++

			continue
		}

		schub.Befunde = append(schub.Befunde, befunde...)
		schub.Geholt++
		schub.Offen--
	}

	return schub, nil
}

// offeneDrucke gibt die Drucke, deren Preis fehlt oder zu alt ist.
//
// **In der Reihenfolge der Sammlung und ohne Wiederholung.** Eine Map allein
// gäbe bei jedem Lauf eine andere Folge, und ein Fortschritt, der springt,
// sieht aus wie ein Fehler.
func (p *Preispflege) offeneDrucke(
	exemplare []*domain.Exemplar, juengster map[string]time.Time,
	hoechstensAlter time.Duration,
) []string {
	jetzt := p.jetzt()
	gesehen := map[string]bool{}
	offene := []string{}

	for _, exemplar := range exemplare {
		// **Ein Wunsch bekommt auch einen Preis.** Was man kaufen will, hat
		// einen Wert — anders als im Portfolio, das den Besitz zusammenrechnet
		// ([0057](../../docs/entscheidungen/0057-wie-eine-sammlung-geordnet-wird.md)).
		if gesehen[exemplar.Printing] {
			continue
		}

		gesehen[exemplar.Printing] = true

		erhoben, da := juengster[exemplar.Printing]
		if da && hoechstensAlter > 0 && jetzt.Sub(erhoben) < hoechstensAlter {
			continue
		}

		offene = append(offene, exemplar.Printing)
	}

	return offene
}
