package app

// **Ein Batch von Vorschlägen** — Werkstück A13.1 von Phase 5.3
// (#1452).
//
// # Ein Batch ist eine Menge und keine Schleife
//
// **Er bucht einmal** ([0105](../../docs/entscheidungen/0105-eine-freigabe-fuer-einen-batch.md)):
// der Nachweis lautet auf den Batch, und die Aufträge darin reservieren nicht
// noch einmal. Drei Buchungen wären drei Freigaben, von denen ein Mensch zwei
// nie gesehen hat.
//
// **Er schreibt nichts.** Das ist derselbe Satz wie bei A7 — „Generieren
// verändert die Zielkarte nicht" —, und er gilt hier für **jedes** Ziel.
//
// **Und ein Ziel, das scheitert, hält ihn nicht auf.** Sonst nähme eine
// Karte, die es nicht mehr gibt, elf anderen ihren Vorschlag weg — und die
// Freigabe wäre verbraucht.
//
// # Die Reihenfolge ist die der Anfrage
//
// **Und nicht die der Antworten.** Wer zwölf Karten schickt, liest die
// Antwort neben seiner Auswahl; eine Liste, die sich nach der Antwortzeit
// ordnet, wäre bei jedem Lauf eine andere.

import (
	"context"
	"fmt"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// KIBatchlage ist, was aus **einem** Ziel geworden ist.
type KIBatchlage struct {
	Zielkarte int64 `json:"zielkarte"`

	// Kennung ist die des Auftrags — leer, wo keiner entstand.
	Kennung string `json:"kennung"`

	Modell string          `json:"modell"`
	Felder []KIFeldwertDTO `json:"felder"`

	// Befunde sind die Beanstandungen des Schemas — sie stehen neben den
	// Feldern und nicht statt ihrer.
	Befunde []string `json:"befunde"`

	Ablauf string `json:"ablauf"`

	// Fehler ist der Grund, **an diesem Ziel** — nil, wo es geklappt hat.
	Fehler *Fehler `json:"fehler"`
}

// KIBatchAntwort trägt die Lagen aller Ziele.
type KIBatchAntwort struct {
	// Kennung ist die des Batches — sie steht im Nachweis als Ziel.
	Kennung string `json:"kennung"`

	Anbieter string `json:"anbieter"`

	Ziele []KIBatchlage `json:"ziele"`

	// Fehler betrifft den **ganzen** Batch — eine zu große Auswahl, eine
	// fehlende Freigabe, ein Anbieter, der nicht antwortet.
	Fehler *Fehler `json:"fehler"`
}

// KIBatchVorschlagen fragt ein Modell für mehrere Karten.
//
// **Der Nachweis gilt dem Batch** und wird **einmal** eingelöst (0105).
func (a *App) KIBatchVorschlagen(
	ctx context.Context, zielkarten []int64, anbieter, alias, auftrag,
	nachweis string,
) KIBatchAntwort {
	ctx = a.imProfil(ctx)

	leer := KIBatchAntwort{Anbieter: anbieter, Ziele: []KIBatchlage{}}

	// **Die Rolle vor dem Modell** — dieselbe Frage wie bei
	// [App.KIVorschlagen] (ID-03, #1702).
	if verweigert := a.darf(ctx, domain.WirkungSchreiben); verweigert != nil {
		leer.Fehler = Uebersetze(ctx, verweigert)

		return leer
	}

	// **Die Grenze vor dem ersten Ruf** (§16, Regel 4): eine, die erst beim
	// Buchen greift, hat schon bezahlt.
	if fehler := domain.PruefeKIBatch(zielkarten); fehler != nil {
		leer.Fehler = Uebersetze(ctx, fehler)

		return leer
	}

	// **Das Profil vor dem Buchen** — ohne eines lässt sich nicht sagen,
	// wessen Geld gemeint ist.
	profil, fehler := a.profilDiesesRufs(ctx)
	if fehler != nil {
		leer.Fehler = Uebersetze(ctx, fehler)

		return leer
	}

	kennung := fmt.Sprintf("kibatch-%d", a.uhr().UnixNano())
	leer.Kennung = kennung

	posten, buchfehler := a.batchBuchen(ctx, anbieter, kennung, nachweis)
	if buchfehler != nil {
		leer.Fehler = Uebersetze(ctx, buchfehler)

		return leer
	}

	// **Ein Werkzeugsatz für den ganzen Batch** und nicht einer je Ziel: das
	// Bauen schriebe sonst `a.kiwerk` aus mehreren Goroutinen. Das Werk
	// selbst hält ein Schloss und trägt die Aufträge aller Ziele.
	werk, quelle, werkfehler := a.kiwerkzeug(ctx, anbieter,
		kideckung{posten: posten})
	if werkfehler != nil {
		// **Hier wird freigegeben und nicht als unklar gemerkt** (#1542):
		// gebucht ist, **hinausgegangen** ist nichts — das Werk steht ja
		// nicht. Wo das belegt ist, wäre „unklar" eine Zahl, die nie
		// entstanden ist.
		a.kiPostenFreigeben(ctx, posten)

		leer.Fehler = Uebersetze(ctx, werkfehler)

		return leer
	}

	defer a.kiPostenSchliessen(ctx, posten)

	return KIBatchAntwort{
		Kennung: kennung, Anbieter: anbieter,
		Ziele: a.batchFahren(ctx, kiwerkzeugsatz{werk: werk, quelle: quelle},
			profil, zielkarten, anbieter, alias, auftrag),
	}
}

// batchBuchen reserviert **einmal** für den ganzen Batch.
//
// **Mit dem Nachweis, der auf ihn lautet** (0105): sein Ziel ist der Anbieter
// wie bei jedem Auftrag, und was ihn vom einzelnen unterscheidet, ist die
// Kennung des Postens — an ihr hängt der Abschluss.
func (a *App) batchBuchen(
	ctx context.Context, anbieter, kennung, nachweis string,
) (string, error) {
	budget, fehler := a.budgetFuer(ctx, anbieter, a.adresseVon(ctx, anbieter),
		kideckung{nachweis: nachweis})
	if fehler != nil {
		return "", fehler
	}

	return budget.Reservieren(ctx, kennung)
}

// batchFahren holt die Vorschläge — nebenläufig, aber begrenzt.
//
// **Die Schranke steht in `domain`** ([domain.KIBatchParallel]) und nicht
// hier: wie viele Anfragen gleichzeitig hinausgehen, ist eine Aussage über
// Höflichkeit gegenüber einem fremden Dienst und keine Einstellung dieser
// Naht.
func (a *App) batchFahren(
	ctx context.Context, satz kiwerkzeugsatz, profil int64,
	zielkarten []int64, anbieter, alias, auftrag string,
) []KIBatchlage {
	lagen := make([]KIBatchlage, len(zielkarten))
	schranke := make(chan struct{}, domain.KIBatchParallel)

	var warten sync.WaitGroup

	for i, karte := range zielkarten {
		warten.Add(1)

		go func() {
			defer warten.Done()

			schranke <- struct{}{}
			defer func() { <-schranke }()

			// **Ein Ziel, das zusammenbricht, ist ein gescheitertes Ziel**
			// (BK-05b-KI, #2112) — und nicht das Ende des Prozesses. Die
			// Schranke und das Warten geben die zwei Zeilen darüber frei; die
			// anderen Ziele fahren weiter wie bei jedem anderen Fehler.
			defer domain.Auffangen(func(zusammenbruch error) {
				lagen[i] = gescheitertesZiel(ctx, karte, zusammenbruch)
			})

			// **Die Stelle steht fest, bevor gefragt wird.** Wer die Antwort
			// anhängte, bekäme die Reihenfolge der Antwortzeiten.
			lagen[i] = a.einZiel(ctx, satz, profil, karte, anbieter, alias,
				auftrag)
		}()
	}

	warten.Wait()

	return lagen
}

// gescheitertesZiel ist die Lage eines Ziels, das zusammenbrach — **ohne**
// Vorschlag. Der Grund geht über [Uebersetze] ins Protokoll, unter der
// Vorgangskennung, die die Lage nennt; im Fenster steht er nicht (§8).
func gescheitertesZiel(ctx context.Context, karte int64, zusammenbruch error) KIBatchlage {
	return KIBatchlage{
		Zielkarte: karte, Felder: []KIFeldwertDTO{}, Befunde: []string{},
		Fehler: Uebersetze(ctx, zusammenbruch),
	}
}

// einZiel holt den Vorschlag für **eine** Karte.
//
// **Mit dem Posten des Batches**: dieser Auftrag bucht nicht noch einmal
// (0105).
func (a *App) einZiel(
	ctx context.Context, satz kiwerkzeugsatz, profil, karte int64,
	anbieter, alias, auftrag string,
) KIBatchlage {
	antwort := a.vorschlagMit(ctx, satz, kifrage{
		profil: profil, karte: karte, anbieter: anbieter, alias: alias, auftrag: auftrag,
	})

	lage := KIBatchlage{
		Zielkarte: karte, Kennung: antwort.Kennung, Modell: antwort.Modell,
		Felder: antwort.Felder, Befunde: antwort.Befunde,
		Ablauf: antwort.Ablauf, Fehler: antwort.Fehler,
	}

	if lage.Felder == nil {
		lage.Felder = []KIFeldwertDTO{}
	}

	if lage.Befunde == nil {
		lage.Befunde = []string{}
	}

	return lage
}

// # Und warum hier **keine** Fensterbindung steht
//
// **Weil sie eine Zusage ohne Einlösung wäre** (§2). Ein Batch am Fenster
// braucht den Kasten, der die **Auswahl** trägt — welche Karten gemeint sind,
// steht in der Kartenliste und nicht in einem Textfeld. Bis er da ist, gäbe
// eine gebundene Methode vor, das Fenster könne etwas, das niemand erreicht;
// genau das ist der Zustand, den `TestJedeBindungWirdGerufen` verhindert.
//
// **Die Fähigkeit steht deshalb an drei Flächen und nicht an vier**, und der
// Grund steht in `docs/FAEHIGKEITEN.md` — nicht als Auslassung, sondern als
// Satz.

// adresseVon nennt, wohin dieser Anbieter gefragt wird.
//
// **Aus derselben Liste, die auch das Auftragswerk baut** — eine zweite Quelle
// daneben liefe beim ersten Gateway auseinander (#1540 war genau das).
func (a *App) adresseVon(ctx context.Context, anbieter string) string {
	alle, fehler := a.kianbieter(ctx)
	if fehler != nil {
		return ""
	}

	for _, einer := range alle {
		if einer.name == anbieter {
			return einer.adresse
		}
	}

	return ""
}
