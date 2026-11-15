package app

// **Ein Vorschlag entsteht und ändert nichts** — Werkstück A7 von Phase 5.3
// (#1446).
//
// # Zwei Rufe, und der erste schreibt nicht
//
// Das Issue sagt den Satz, um den herum das gebaut ist: „**Generieren
// verändert Zielkarte nicht.**" [App.KIVorschlagen] fragt das Modell, prüft
// gegen das Schema und legt den Vorschlag ab; [App.KIVorschlagUebernehmen]
// schreibt — und prüft dabei noch einmal alles, was sich inzwischen geändert
// haben kann.
//
// # Was **nicht** übernommen wird
//
// **Eine abgeschnittene Antwort.** `done_reason: "length"` heißt: das Fenster
// war voll, und was dasteht, ist eine halbe Antwort. Sie sieht aus wie eine
// ganze — deshalb wird sie hier abgewiesen und nicht dem Schema überlassen,
// das ihr halbes JSON ohnehin nicht läse.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// KIRohantwort ist, was eine Laufzeit auf eine Frage gibt.
type KIRohantwort struct {
	// Inhalt ist der Text — bei gesetztem Schema das JSON.
	Inhalt string

	// Abgeschnitten sagt, ob die Antwort am Fensterende abbrach.
	Abgeschnitten bool

	// Modell ist, was **tatsächlich** geantwortet hat.
	Modell string
}

// KIFeldwertDTO ist ein vorgeschlagener Feldwert.
type KIFeldwertDTO struct {
	Feld  string   `json:"feld"`
	Werte []string `json:"werte"`

	// Bisher ist, was **heute** im Feld steht — leer heißt: nichts.
	//
	// # Wofür
	//
	// **Damit sichtbar ist, was sich ändert** (A11, #1450): das Issue
	// verlangt „Feld-Diff" und „Nutzende sehen konkrete Änderungen". Ein
	// Vorschlag, der nur den neuen Wert zeigt, verlangt von jemandem, dass er
	// den alten im Kopf hat.
	//
	// **Gelesen beim Vorschlagen und nicht beim Anzeigen**: zwischen beidem
	// läge ein Stand, aus dem ein anderer Vergleich entstünde als der, gegen
	// den geprüft wurde.
	Bisher string `json:"bisher"`
}

// KIVorschlagDTO ist ein Vorschlag, wie das Fenster ihn sieht.
type KIVorschlagDTO struct {
	// Kennung ist, womit er übernommen wird — **einmal**.
	Kennung string `json:"kennung"`

	// Anbieter ist, wen gefragt wurde.
	//
	// **Neben dem Modell und nicht statt seiner** (A8, #1447): das Issue
	// verlangt „Provider-/Modellname im Vorschlag sichtbar". Mit einem
	// Anbieter war der Alias die ganze Auskunft; mit zweien ist er die
	// halbe, denn beide dürfen denselben Namen führen.
	Anbieter string `json:"anbieter"`

	// Modell ist, was geantwortet hat.
	Modell string `json:"modell"`

	// Felder sind die Werte, die das Schema durchgelassen hat.
	Felder []KIFeldwertDTO `json:"felder"`

	// Befunde sind die Felder, die es **nicht** durchgelassen hat, samt
	// Grund. Sie stehen neben den guten und nicht an ihrer Stelle: ein
	// Vorschlag, der halb passt, ist kein Grund, die gute Hälfte
	// wegzuwerfen.
	Befunde []string `json:"befunde"`

	// Ablauf ist, bis wann er übernommen werden darf — RFC 3339.
	Ablauf string `json:"ablauf"`

	// Uebernommen ist die Zahl der geschriebenen Felder — **null**, solange
	// niemand übernommen hat. An den Flächen ohne Fenster steht sie hinter
	// [App.KIVorschlagInEinemZug].
	Uebernommen int `json:"uebernommen"`

	// Belege sind die Stellen, die das Modell **gesehen** hat — Werkstück
	// A13.4 von Phase 5.3
	// (#1452).
	//
	// **Leer, wo keine Belegfrage gestellt wurde**, und das ist der
	// gewöhnliche Fall. Sie stehen **neben** den Feldern und nicht als deren
	// Beleg: kein Wort davon sagt, dass ein Wert daraus folgt — was das
	// Modell vorschlägt, bleibt eine Vermutung und wird als `ki` verbucht.
	Belege []BelegDTO `json:"belege"`

	// Korpusstaende sind die Stände der Quellen, aus denen die Belege
	// stammen — dieselbe Angabe wie bei der Einschätzung.
	Korpusstaende []KorpusstandDTO `json:"korpusstaende"`

	// Einschraenkungen sagen, was **nicht** gesucht wurde — etwa der eigene
	// Bestand ohne Freigabe (A13.4).
	Einschraenkungen []string `json:"einschraenkungen"`

	// Weg sagt, wie der Anbieter drankam — `gewaehlt`, `vorrang` oder
	// `rueckfall` (AI-05.AK-05, #2042).
	Weg string `json:"weg"`

	// Ausgang ist leer, oder [KIAusgangAbgelehnt]: das Modell hat
	// verweigert, und die Befunde sagen, was es sagte (AI-05.AK-06).
	Ausgang string `json:"ausgang"`

	// Rueckfaelle nennt nach einer Ablehnung die Anbieter, die es versuchen
	// dürfen — lokale, nie die Cloud.
	Rueckfaelle []string `json:"rueckfaelle"`
}

// KIVorschlagAntwort trägt ihn samt Fehler.
type KIVorschlagAntwort struct {
	KIVorschlagDTO

	Fehler *Fehler `json:"fehler"`
}

// KIUebernahmeAntwort sagt, was übernommen wurde.
type KIUebernahmeAntwort struct {
	Felder int `json:"felder"`

	Fehler *Fehler `json:"fehler"`
}

// profilDiesesRufs gibt das Profil aus dem Kontext.
//
// # Warum als Methode und nicht je Stelle
//
// **Weil `internal/storage` nur an den Composition Roots bekannt sein darf**
// (§2, `internal/schichten/netz_test.go` führt die Aufzählung). Eine neue
// Naht, die das Profil braucht, bekommt es hier — und muss dafür nicht in
// die Liste der Dateien, die `storage` kennen.
func (a *App) profilDiesesRufs(ctx context.Context) (int64, error) {
	return storage.ProfilAus(ctx)
}

// KIVorschlagen fragt ein Modell und legt seinen Vorschlag ab.
//
// **Es schreibt nichts an der Karte** — auch dann nicht, wenn alles passt.
func (a *App) KIVorschlagen(
	ctx context.Context, karteID int64, anbieter, alias, auftrag string,
	nachweis string,
) KIVorschlagAntwort {
	ctx = a.imProfil(ctx)

	// **Die Rolle vor dem Modell** — ID-03.AK-04 (#1702). Die Matrix führt
	// diese Fähigkeit unter `lokalSchreiben`, Mindestrolle Mitarbeiter: ein
	// Modellruf ändert nichts an der Karte, kostet aber Zeit, Strom und bei
	// einem Anbieter Geld. Gehalten hat das bis ID-03 nur die nur lesende
	// Verbindung — und die hält einen Schritt nicht auf, der nicht schreibt.
	if verweigert := a.darf(ctx, domain.WirkungSchreiben); verweigert != nil {
		return KIVorschlagAntwort{Fehler: Uebersetze(ctx, verweigert)}
	}

	return a.kiVorschlagMitDeckung(ctx, karteID, anbieter, alias, auftrag,
		kideckung{nachweis: nachweis})
}

// kiVorschlagMitDeckung ist derselbe Weg, mit der Deckung als Verbund.
//
// **Getrennt wegen des Batches** (A13.1, #1452): dort trägt nicht jeder
// Auftrag seine eigene Freigabe, sondern alle laufen unter der des Batches
// ([0105](../../docs/entscheidungen/0105-eine-freigabe-fuer-einen-batch.md)).
//
// **Es ruft [App.imProfil] nicht noch einmal**: der Aufrufer hat es getan,
// und ein zweites gäbe eine zweite Vorgangskennung für denselben Ruf.
func (a *App) kiVorschlagMitDeckung(
	ctx context.Context, karteID int64, anbieter, alias, auftrag string,
	deckung kideckung,
) KIVorschlagAntwort {
	// **Das Profil zuerst** — es ist die billigste Prüfung, und ohne eines
	// lässt sich nicht sagen, wessen Daten gemeint sind. `imProfil` hat es
	// eben in den Kontext gelegt; ein Ruf ohne Fenster (MCP, REST, CLI) hat
	// nur ihn.
	profil, fehler := a.profilDiesesRufs(ctx)
	if fehler != nil {
		return KIVorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	werk, quelle, fehler := a.kiwerkzeug(ctx, anbieter, deckung)
	if fehler != nil {
		return KIVorschlagAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return a.vorschlagMit(ctx, kiwerkzeugsatz{werk: werk, quelle: quelle},
		kifrage{profil: profil, karte: karteID, anbieter: anbieter, alias: alias, auftrag: auftrag})
}

// kiwerkzeugsatz ist das Werk samt der Quelle, die es bedient.
//
// **Als Verbund, weil ein Batch beides einmal baut** (A13.1, #1452): jedes
// Ziel für sich zu bauen schriebe `a.kiwerk` aus mehreren Goroutinen — ein
// Wettlauf, den der Detektor am 2026-09-13 gefunden hat. Das Werk selbst ist
// nebenläufig benutzbar (`service.KIAuftragswerk` hält ein Schloss); das
// **Bauen** war es nicht.
type kiwerkzeugsatz struct {
	werk   *service.KIAuftragswerk
	quelle KILaufzeitquelle
}
