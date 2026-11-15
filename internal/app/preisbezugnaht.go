package app

// **Die Naht zum Bulkbezug** — Werkstück A3 von Phase 5.3
// (#1442).
//
// # Die Composition Root für eine Fähigkeit, die ins Netz geht
//
// Hier wird der Klient gebaut — mit Adresse, Fristen und Grenzen. Der Dienst
// dahinter kennt nur die Schnittstelle (§2), und `internal/mtgjsonpreise`
// kennt weder Bestand noch Einstellungen.

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgjsonpreise"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// BulkbilanzDTO ist, was aus einem Bezug wurde — wie das Fenster es sieht.
type BulkbilanzDTO struct {
	// Quelle ist die Familie, die gefragt wurde.
	Quelle string `json:"quelle"`

	// Umfang ist `heute` oder `neunzigTage`.
	Umfang string `json:"umfang"`

	// Drucke ist die Zahl der Drucke, nach denen gefragt wurde.
	Drucke int `json:"drucke"`

	// Staende ist die Zahl der abgelegten Preisstände.
	Staende int `json:"staende"`

	// Befunde sind die Angaben, die nicht zu deuten waren (§8).
	Befunde []string `json:"befunde"`
}

// Bulkbezugsantwort trägt die Bilanz.
type Bulkbezugsantwort struct {
	BulkbilanzDTO

	Fehler *Fehler `json:"fehler"`
}

// krosswalknaht bringt die Karte aus `storage` an die Schnittstelle.
type krosswalknaht struct{ db *sql.DB }

func (k krosswalknaht) Krosswalk(ctx context.Context) (map[string]string, error) {
	return storage.Druckkrosswalk(ctx, k.db)
}

// mtgjsonbulk baut den Klienten für den Bulkbezug.
//
// **Die Adresse kommt aus den Einstellungen, wenn eine dasteht** — sonst die
// Vorgabe des Pakets. Ein Lauf gegen einen eigenen Server braucht sie, und
// ohne sie ginge jeder Lauf ins Netz (§3).
// **Der Schalter erreicht ihn** (§16, Regel 6): `MitErlaubnis` fragt vor jedem
// Ruf, ob dieser Dienst an ist. Ohne ihn ginge ein Bezug hinaus, obwohl die
// Vorgabe für die neuen Dienste „aus" lautet
// ([0063](../../docs/entscheidungen/0063-neue-dienste-sind-aus.md)) — und
// derselbe Fehler ist am 2026-09-11 schon einmal an dieser Stelle passiert,
// beim Kompatibilitätsmanifest aus O6.
func mtgjsonbulk(wahl Dienstwahl, adresse string) *mtgjsonpreise.Klient {
	return mtgjsonpreise.NeuerKlient(bulkadresseAus(adresse),
		netz.NeuerKlientMit(netz.UeblicheFristen(), proxyquelleAus(wahl)).
			MitErlaubnis(func(ctx context.Context) error {
				return pruefeNachAussen(ctx, wahl, domain.DienstMTGJSON)
			}),
		netz.Kenner{
			Name:    "MagicSetEditor4Mac/" + Fassung,
			Annahme: "*/*",
		},
		mtgjsonpreise.Grenzen{})
}

// PreisbezugUeber baut den Anwendungsfall.
func PreisbezugUeber(wahl Dienstwahl, db *sql.DB, adresse string) *service.Preisbezug {
	return service.NeuerPreisbezug(
		storage.NeuerPreisspeicher(db),
		mtgjsonbulk(wahl, adresse),
		krosswalknaht{db: db},
		time.Now,
	)
}

// MTGJSONPreisumgebung ist die Umgebungsvariable, die die Grundadresse umlenkt.
//
// **Dieselbe Naht wie [Scryfallumgebung] und aus demselben Grund** (§3): ohne
// sie ginge jeder Flächenlauf gegen mtgjson.com — und zöge dabei ein Gigabyte.
const MTGJSONPreisumgebung = "MSE_MTGJSON_PREISADRESSE"

// bulkadresseAus gibt die umgelenkte Adresse — oder die genannte.
func bulkadresseAus(genannt string) string {
	if ort := os.Getenv(MTGJSONPreisumgebung); ort != "" {
		return ort
	}

	return genannt
}

// PreisbulkAnwendungUeber baut eine Anwendung, die den Bulkbezug kann.
//
// **Über `NutzerflaecheUeber`**, wie die anderen Flächenwurzeln: der Akteur
// und sein Profil hängen daran, und ein Bezug ohne sie schriebe in einen
// Bestand, dessen Eigentümer nicht feststeht (§2, Composition Root).
func PreisbulkAnwendungUeber(db *sql.DB) *App {
	anwendung := NutzerflaecheUeber(db).app

	// **Die Verbindung wird gesetzt und nicht vorausgesetzt.**
	//
	// `NeueApp("")` öffnet keine — sie kommt sonst beim Start des Fensters.
	// Eine Flächenwurzel ohne sie meldete „es ist kein Bestand offen", und der
	// Satz wäre wahr und unbrauchbar: der Bestand liegt da, nur weiß diese
	// Anwendung nichts von ihm. Nachgemessen am 2026-09-11 am Flächenlauf.
	anwendung.db = db

	return anwendung
}

// MitBulkadresse setzt die Adresse, unter der die Preisdatei liegt.
//
// **Für Läufe gegen eine eigene Gegenstelle** (§3): ohne sie ginge jeder Lauf
// ins Netz. Leer heißt: die Vorgabe des Pakets.
func (a *App) MitBulkadresse(adresse string) *App {
	a.bulkadresse = adresse

	return a
}

// Preisbulk holt eine ganze Preisdatei und legt ab, was darin steht.
//
// **Die Quelle wird genannt und nicht erraten** (A3, #1442): es gibt heute
// eine, und trotzdem steht sie im Aufruf — eine Vorgabe wäre eine
// Preisstrategie, die aussieht wie keine.
func (a *App) Preisbulk(ctx context.Context, quelle, umfang string) Bulkbezugsantwort {
	ctx = a.imProfil(ctx)

	if genannt := quelle; genannt != string(mtgjsonpreise.Familie) {
		// **Der Name der Quelle gehört in die Meldung**, und er ist kein Pfad:
		// er kommt aus dem Aufruf und steht in `FAEHIGKEITEN.md`. Ohne ihn
		// hieße es „die Preisquelle kennt diese Anwendung nicht", und wer
		// `cardmarket` getippt hat, erführe nicht, dass es an seinem Wort lag.
		return Bulkbezugsantwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: die Preisquelle %q kennt diese Anwendung nicht — es gibt %q",
			domain.ErrValidation, genannt, string(mtgjsonpreise.Familie)))}
	}

	verbindung := a.laufendeVerbindung()
	if verbindung == nil {
		return Bulkbezugsantwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: es ist kein Bestand offen", domain.ErrValidation))}
	}

	// **Die Wahl kommt aus den Einstellungen dieses Bestands** — nicht aus
	// einer Vorgabe im Code.
	wahl := NeueEinstellungswahl(storage.NeueEinstellungen(verbindung))

	bilanz, fehler := PreisbezugUeber(wahl, verbindung, a.bulkadresse).Holen(ctx, umfang)
	if fehler != nil {
		return Bulkbezugsantwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return Bulkbezugsantwort{BulkbilanzDTO: zuBulkbilanz(quelle, umfang, bilanz)}
}

// zuBulkbilanz macht aus der Bilanz ihr DTO.
func zuBulkbilanz(quelle, umfang string, bilanz service.Bulkbilanz) BulkbilanzDTO {
	befunde := make([]string, 0, len(bilanz.Befunde))
	for _, befund := range bilanz.Befunde {
		befunde = append(befunde, befund.Meldung)
	}

	return BulkbilanzDTO{
		Quelle: quelle, Umfang: umfang,
		Drucke: bilanz.Drucke, Staende: bilanz.Staende, Befunde: befunde,
	}
}
