package app

// **Die Naht zum Retrievalindex** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Warum der Einbetter hier entsteht und nicht im Dienst
//
// **Weil der Fachkern keine Laufzeit kennen darf** (§2, und das Issue sagt es
// eigens: „keine harte Ollama-Abhängigkeit im Retrievalvertrag"). Was
// `service` verlangt, ist eine [service.Einbetterquelle] — ein Name hinein,
// ein Einbetter heraus. Dass dahinter Ollama steht, ist eine Entscheidung
// dieser Datei.
//
// # Und warum die Verwaltung gehalten wird
//
// **Weil der Riegel in ihr sitzt.** Ein je Ruf neu gebauter Aufbau hätte je
// Ruf einen eigenen Riegel — und dann bauten zwei Aufrufer zugleich in
// denselben Index. Das ist genau der Fall, den Werkstück 4 ausschließt: „Eine
// aktive Schreibinstanz je Index."

import (
	"context"
	"database/sql"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/ollama"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// einbetterquelle baut Einbetter über eine lokale Laufzeit.
//
// **Der Klient entsteht je Ruf**, weil die Adresse in den Einstellungen steht
// und sich ändert, während das Fenster offen ist — dieselbe Entscheidung wie
// bei `kilaufzeit.go`.
type einbetterquelle struct {
	bauen func(ctx context.Context) *ollama.Klient
}

// Einbetter vermisst das Modell und gibt einen Einbetter in seinem Raum.
func (e einbetterquelle) Einbetter(ctx context.Context, modell string) (
	service.Einbetter, []domain.Befund, error,
) {
	klient := e.bauen(ctx)

	raum, befunde, fehler := klient.Raum(ctx, modell)
	if fehler != nil || len(befunde) > 0 {
		return nil, befunde, fehler
	}

	return klient.Einbetter(raum), nil, nil
}

// sammeldateioeffner öffnet eine Scryfall-Sammeldatei als Korpus.
type sammeldateioeffner struct{}

// Oeffnen öffnet die Datei — der Aufrufer schließt sie über die Rückgabe.
func (sammeldateioeffner) Oeffnen(
	_ context.Context, auftrag service.Aufbauauftrag,
) (service.Korpusquelle, func(), error) {
	leser, fehler := scryfall.OeffneBulk(auftrag.Quelle)
	if fehler != nil {
		return nil, nil, fehler
	}

	return service.NeuerScryfallkorpus(leser, domain.Korpusstand{
		Quelle: domain.QuelleOracle, Revision: auftrag.Revision,
		Datum: auftrag.Datum,
	}, auftrag.Abgerufen), leser.Schliessen, nil
}

// RetrievalUeber baut die Verwaltung des Suchindex.
//
// **Die Verbindung kommt als Funktion**, weil ein Zurückspielen sie
// austauscht — dieselbe Entscheidung wie bei `ruecksetzungsnaht`.
func RetrievalUeber(
	holen func() *sql.DB, laufzeit func(ctx context.Context) *ollama.Klient,
	ort func() string,
) *service.Retrievalverwaltung {
	index := storage.NeuerRetrievalindex(holen())

	return service.NeueRetrievalverwaltung(index, index,
		einbetterquelle{bauen: laufzeit}, sammeldateioeffner{},
		// **Der Haushalt der Maschine**: die Grenzen gelten dem Rechner und
		// nicht dieser Fläche — ein Indexbau über die Kommandozeile bindet
		// dieselben Kerne wie einer aus dem Fenster.
		service.Maschinenhaushalt()).
		// **Und die Absprache über die Prozesse** (`retrievalsperre.go`): der
		// Riegel im Prozess hält das Fenster auf, nicht `mse indexaufbau`.
		MitProzesssperre(Indexsperre(ort))
}

// retrievalVerdrahten hängt den Suchindex ins Fenster.
func (a *App) retrievalVerdrahten() {
	a.MitRetrieval(RetrievalUeber(a.laufendeVerbindung, a.retrievallaufzeit,
		func() string { return a.datenbankpfad }))
}

// retrievallaufzeit baut den Klienten für die Einbettung.
//
// **Mit derselben Erlaubnisfrage wie jede andere KI-Anfrage** — ein Schalter
// ändert sich, nachdem die Naht gebaut ist, und die Frage steht deshalb am
// Transport ([App.pruefeKILaufzeit]).
func (a *App) retrievallaufzeit(ctx context.Context) *ollama.Klient {
	adresse := a.kiadresse(ctx)

	return ollama.NeuerKlient(adresse, a.netzklientMit(netz.Fristen{
		Verbindung: kiverbindungsfrist,
		Kopf:       kikopffrist,
		Stillstand: einbettungsfrist,
	}).MitErlaubnis(func(ctx context.Context) error {
		return a.pruefeKILaufzeit(ctx, adresse)
	}))
}

// einbettungsfrist ist die Stillstandsfrist einer Einbettung.
//
// **Fünf Minuten**, und die Zahl ist gewählt: ein kaltes Modell lädt seine
// Gewichte, bevor es den ersten Vektor gibt, und danach rechnet es 32 Texte.
// Wer hier die Frist einer einzelnen Frage nähme, bräche jeden ersten Stapel
// ab.
const einbettungsfrist = 5 * time.Minute

// kiadresse liest die Adresse der lokalen Laufzeit.
//
// **Leer, wo keine eingestellt ist** — dann nimmt `ollama.NeuerKlient` seine
// Vorgabeadresse auf diesem Rechner.
func (a *App) kiadresse(ctx context.Context) string {
	if a.einstellungen == nil {
		return ""
	}

	stand, fehler := a.einstellungen.Stand(ctx)
	if fehler != nil {
		return ""
	}

	return stand[domain.KIAdresse].Wirksam
}

// MitBestandsort sagt der Anwendung, wo ihr Bestand liegt.
//
// # Warum das eine eigene Naht braucht
//
// **Weil die Flächen ohne Fenster ihre Anwendung aus einem offenen
// `*sql.DB` bauen** und der Pfad dabei verlorengeht: `NutzerflaecheUeber`
// geht über `NeueApp("")`. Am Fenster steht er seit jeher, denn dort öffnet
// [App.Starten] die Datei selbst.
//
// **Und daran hängt mehr als ein Feld**: [Indexsperre] ist die Absprache
// **über Prozesse hinweg**, und ohne Pfad gibt sie „genommen" zurück, ohne
// etwas zu nehmen. Zwei `mse indexaufbau` bauten damit gleichzeitig gegen
// dieselbe Datei — während der Satz in [RetrievalUeber] zusagt, die
// Absprache gebe es gerade für diese Fläche. Gefunden vom Flächenlauf zum
// laufenden Rebuild (A14,
// #1453).
func (a *App) MitBestandsort(ort string) *App {
	a.datenbankpfad = ort

	return a
}

// Bestandsort nennt ihn — leer heißt: es gibt keine Absprache.
func (a *App) Bestandsort() string { return a.datenbankpfad }

// RetrievalAnwendungUeber baut eine Anwendung, die den Suchindex kennt.
//
// **Mit den Einstellungen**, weil die Adresse der lokalen Laufzeit dort steht
// — ohne sie zeigte der Aufbau an dieser Fläche auf einen anderen Port als am
// Fenster, und das ist dieselbe Frage mit zwei Antworten (§16).
func RetrievalAnwendungUeber(db *sql.DB) *App {
	anwendung := NutzerflaecheUeber(db).app
	anwendung.db = db

	anwendung.MitEinstellungen(einstellungsverwalterUeber(db))

	anwendung.MitRetrieval(RetrievalUeber(
		func() *sql.DB { return db }, anwendung.retrievallaufzeit,
		func() string { return anwendung.datenbankpfad }))

	return anwendung
}
