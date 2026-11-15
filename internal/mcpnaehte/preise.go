// **Die Preise für den MCP-Server** —
// #1165.

package mcpnaehte

import (
	"context"
	"database/sql"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mcp"
)

// preise reicht die vier Züge an die Naht durch.
// **Und sie hält die KI-Anwendung, statt sie je Ruf zu bauen** (A13.1b,
// #1452).
//
// # Warum das nötig wurde
//
// **Weil ein offener Vorschlag im Prozess lebt.** `service.KIAuftragswerk`
// hält seine offenen Aufträge im Arbeitsspeicher; eine Anwendung je Ruf hieße,
// dass die Kennungen, die `schreiben.ki.batch` zurückgibt, beim **nächsten**
// Ruf nicht mehr existieren.
//
// **Das war schon vor A13.1b ein Befund**, nur ohne Verbraucher: der Batch
// nennt je Ziel eine Kennung samt Ablauf, und niemand konnte sie einlösen.
// Mit `schreiben.ki.gruppe` hat er einen, und der Flächenlauf hat es sofort
// gezeigt — die Gruppe fand keinen einzigen Auftrag.
//
// **Warum das über MCP geht und über die Kommandozeile nicht**: `mse-mcp` ist
// **ein** Prozess über viele Rufe (A14, #1453); `mse` baut je Aufruf neu, und
// dort bleibt es beim Weg aus A13.1 — erzeugen und übernehmen in einem Zug.
//
// **`mse-rest` steht auf der Seite von `mse-mcp`**: es ist ein **Dienst**,
// seine Nahttafel entsteht einmal beim Start und lebt so lange wie der
// Prozess. Dieselbe Korrektur steht dort in `cmd/mse-rest/sammlung.go`.
type preise struct {
	db *sql.DB

	// anwendung ist die **eine** KI-Anwendung dieses Prozesses.
	//
	// **Faul gebaut und danach gehalten**: sie erst beim ersten KI-Ruf
	// anzulegen kostet keinen Aufbau in einer Sitzung, die nie ein Modell
	// fragt.
	anwendung *app.App
}

// kianwendung gibt die Anwendung dieses Prozesses — und baut sie beim ersten
// Mal.
func (p *preise) kianwendung() *app.App {
	if p.anwendung == nil {
		p.anwendung = app.KIAnwendungUeber(p.db)
	}

	return p.anwendung
}

func (p *preise) PreiseHolen(ctx context.Context, printing string) (
	[]mcp.Preissicht, []string, error,
) {
	naht, err := app.SammlungsnahtUeber(p.db)
	if err != nil {
		return nil, nil, err
	}

	staende, befunde, holfehler := naht.PreiseHolen(ctx, printing)
	if holfehler != nil {
		return nil, befunde, holfehler
	}

	return alsPreissichten(staende), befunde, nil
}

func (p *preise) Preise(ctx context.Context, printing string) ([]mcp.Preissicht, error) {
	naht, err := app.SammlungsnahtUeber(p.db)
	if err != nil {
		return nil, err
	}

	staende, lesefehler := naht.Preise(ctx, printing)
	if lesefehler != nil {
		return nil, lesefehler
	}

	return alsPreissichten(staende), nil
}

func (p *preise) Preisverlauf(ctx context.Context, reihe mcp.Preisreihe) (
	[]mcp.Preissicht, error,
) {
	naht, err := app.SammlungsnahtUeber(p.db)
	if err != nil {
		return nil, err
	}

	staende, lesefehler := naht.Preisverlauf(ctx, app.PreisreiheDTO{
		Printing: reihe.Printing, Quelle: reihe.Quelle, Markt: reihe.Markt,
		Art: reihe.Art, Waehrung: reihe.Waehrung, Finish: reihe.Finish,
		Zustand: reihe.Zustand, Sprache: reihe.Sprache,
	})
	if lesefehler != nil {
		return nil, lesefehler
	}

	return alsPreissichten(staende), nil
}

// Marktangebote gibt die Posten eines Marktplatzes zu einem Druck (A4, #1443).
func (p *preise) Marktangebote(ctx context.Context, wunsch mcp.Marktwunsch) (
	mcp.Marktsicht, error,
) {
	antwort := app.MarktAnwendungUeber(p.db).Marktangebote(ctx,
		wunsch.Quelle, wunsch.Printing, wunsch.Sprache, wunsch.Finish, wunsch.Zustaende)
	if antwort.Fehler != nil {
		return mcp.Marktsicht{}, antwort.Fehler.AlsFehler()
	}

	angebote := make([]mcp.Angebotssicht, 0, len(antwort.Angebote))
	for _, angebot := range antwort.Angebote {
		angebote = append(angebote, mcp.Angebotssicht(angebot))
	}

	return mcp.Marktsicht{
		Quelle: antwort.Quelle, Fassung: antwort.Fassung,
		Zugeordnet: antwort.Zugeordnet, Angebote: angebote,
		Befunde: antwort.Befunde,
	}, nil
}

// Anbieterstand nennt das Kontingent an einem Anbieterkonto (A5, #1444).
func (p *preise) Anbieterstand(ctx context.Context, dienst string) (
	mcp.Anbietersicht, error,
) {
	antwort := app.AnbieterAnwendungUeber(p.db).Anbieterstand(ctx, dienst)
	if antwort.Fehler != nil {
		return mcp.Anbietersicht{}, antwort.Fehler.AlsFehler()
	}

	return mcp.Anbietersicht(antwort.AnbieterstandDTO), nil
}

// Fremdpreisverlauf holt den Verlauf eines Drucks bei einer Quelle (A5, #1444).
func (p *preise) Fremdpreisverlauf(ctx context.Context, wunsch mcp.Verlaufsbezug) (
	mcp.Verlaufssicht, error,
) {
	antwort := app.AnbieterAnwendungUeber(p.db).Fremdpreisverlauf(ctx,
		wunsch.Quelle, wunsch.Printing, wunsch.Finish, wunsch.Zustaende)
	if antwort.Fehler != nil {
		return mcp.Verlaufssicht{}, antwort.Fehler.AlsFehler()
	}

	return mcp.Verlaufssicht(antwort.FremdpreisbilanzDTO), nil
}

// KIModelle nennt, was über KI-Modelle feststeht (A6, #1445).
func (p *preise) KIModelle(ctx context.Context) (mcp.KIKatalogsicht, error) {
	antwort := p.kianwendung().KIModelle(ctx)
	if antwort.Fehler != nil {
		return mcp.KIKatalogsicht{}, antwort.Fehler.AlsFehler()
	}

	modelle := make([]mcp.KIModellsicht, 0, len(antwort.Modelle))
	for _, modell := range antwort.Modelle {
		belege := make([]mcp.KIBelegsicht, 0, len(modell.Faehigkeiten))
		for _, beleg := range modell.Faehigkeiten {
			belege = append(belege, mcp.KIBelegsicht(beleg))
		}

		glieder := make([]mcp.KIGliedsicht, 0, len(modell.Herkunft))
		for _, glied := range modell.Herkunft {
			glieder = append(glieder, mcp.KIGliedsicht(glied))
		}

		modelle = append(modelle, mcp.KIModellsicht{
			Anbieter: modell.Anbieter, Alias: modell.Alias,
			Laufzeit: modell.Laufzeit, Ort: modell.Ort,
			Faehigkeiten: belege, Herkunft: glieder, Bindung: modell.Bindung,

			Kontextfenster: modell.Kontextfenster, KontextBeleg: modell.KontextBeleg,
		})
	}

	return mcp.KIKatalogsicht{
		NurLokal: antwort.NurLokal, Einrichtung: antwort.Einrichtung,
		Anbieter: anbietersichten(antwort.Anbieter),
		Cloud:    mcp.KICloudsicht(antwort.Cloud),
		Befunde:  antwort.Befunde, Modelle: modelle, Vorrang: antwort.Vorrang,
	}, nil
}

// KIVorschlag lässt ein Modell Felder einer Karte vorschlagen (A7, #1446).
//
// **In einem Zug**, weil diese Fläche die Anwendung je Aufruf baut: ein
// zweiter Ruf fände den Vorschlag des ersten nicht mehr
// ([app.App.KIVorschlagInEinemZug]).
func (p *preise) KIVorschlag(ctx context.Context, bezug mcp.KIVorschlagsbezug) (
	mcp.KIVorschlagssicht, error,
) {
	antwort := p.kianwendung().KIVorschlagInEinemZug(ctx,
		app.KIVorschlagsbezug{
			KarteID: bezug.KarteID, Anbieter: bezug.Anbieter,
			Alias: bezug.Alias, Auftrag: bezug.Auftrag,

			// **Die Belegfrage und die Freigabe reisen mit** (A13.4, #1452)
			// — ohne sie bleibt es beim Vorschlag ohne Belege, und der
			// eigene Bestand bleibt außen.
			Belegfrage: bezug.Belegfrage, Freigegeben: bezug.Freigegeben,
			Nur: bezug.Nur,
		},
		bezug.Uebernehmen, bezug.Felder)
	if antwort.Fehler != nil {
		return mcp.KIVorschlagssicht{}, antwort.Fehler.AlsFehler()
	}

	felder := make([]mcp.KIFeldwertsicht, 0, len(antwort.Felder))
	for _, feld := range antwort.Felder {
		felder = append(felder, mcp.KIFeldwertsicht(feld))
	}

	return mcp.KIVorschlagssicht{
		Kennung: antwort.Kennung, Modell: antwort.Modell, Felder: felder,
		Befunde: antwort.Befunde, Ablauf: antwort.Ablauf,
		Uebernommen: antwort.Uebernommen,
		Anbieter:    antwort.Anbieter, Weg: antwort.Weg,
		Ausgang: antwort.Ausgang, Rueckfaelle: antwort.Rueckfaelle,
	}, nil
}

// KIModellPruefen misst, was ein Modell kann (A9, #1448).
//
// **Es kostet Geld**, wenn der Anbieter anderswo rechnet — und geht deshalb
// durch dieselbe Sperre wie ein Vorschlag.
func (p *preise) KIModellPruefen(ctx context.Context, bezug mcp.KIProbenbezug) (
	mcp.KIProbensicht, error,
) {
	antwort := p.kianwendung().KIModellPruefen(
		ctx, bezug.Anbieter, bezug.Alias, "")
	if antwort.Fehler != nil {
		return mcp.KIProbensicht{}, antwort.Fehler.AlsFehler()
	}

	proben := make([]mcp.KIProbesicht, 0, len(antwort.Proben))
	for _, probe := range antwort.Proben {
		proben = append(proben, mcp.KIProbesicht(probe))
	}

	return mcp.KIProbensicht{Proben: proben}, nil
}

// KIModellVergessen nimmt die Messungen eines Modells zurück (A9, #1448).
func (p *preise) KIModellVergessen(
	ctx context.Context, bezug mcp.KIProbenbezug,
) error {
	antwort := p.kianwendung().KIModellVergessen(
		ctx, bezug.Anbieter, bezug.Alias)
	if antwort.Fehler != nil {
		return antwort.Fehler.AlsFehler()
	}

	return nil
}

// Preisbulk holt eine ganze Preisdatei und legt ab, was darin steht.
func (p *preise) Preisbulk(ctx context.Context, quelle, umfang string) (
	mcp.Bulkbilanzsicht, error,
) {
	antwort := app.PreisbulkAnwendungUeber(p.db).Preisbulk(ctx, quelle, umfang)
	if antwort.Fehler != nil {
		return mcp.Bulkbilanzsicht{}, antwort.Fehler.AlsFehler()
	}

	return mcp.Bulkbilanzsicht{
		Quelle: antwort.Quelle, Umfang: antwort.Umfang,
		Drucke: antwort.Drucke, Staende: antwort.Staende,
		Befunde: antwort.Befunde,
	}, nil
}

func (p *preise) Portfolio(ctx context.Context) (mcp.Portfoliosicht, error) {
	naht, err := app.SammlungsnahtUeber(p.db)
	if err != nil {
		return mcp.Portfoliosicht{}, err
	}

	zeilen, ohnePreis, rechenfehler := naht.Portfolio(ctx)
	if rechenfehler != nil {
		return mcp.Portfoliosicht{}, rechenfehler
	}

	sichten := make([]mcp.Portfoliozeilensicht, 0, len(zeilen))

	for _, zeile := range zeilen {
		sichten = append(sichten, mcp.Portfoliozeilensicht{
			Waehrung:       zeile.Waehrung,
			Wert:           zeile.Wert,
			Exemplare:      zeile.Exemplare,
			AeltesterPreis: zeile.AeltesterPreis,
			AlterStunden:   zeile.AlterStunden,
			Quelle:         zeile.Quelle,
		})
	}

	return mcp.Portfoliosicht{Zeilen: sichten, OhnePreis: ohnePreis}, nil
}

// alsPreissichten übersetzt die DTOs in die Sicht dieser Fläche.
func alsPreissichten(dtos []app.PreisstandDTO) []mcp.Preissicht {
	sichten := make([]mcp.Preissicht, 0, len(dtos))

	for _, dto := range dtos {
		sichten = append(sichten, mcp.Preissicht{
			Printing:     dto.Printing,
			DruckID:      dto.DruckID,
			Quelle:       dto.Quelle,
			Waehrung:     dto.Waehrung,
			Finish:       dto.Finish,
			Wert:         dto.Wert,
			Erhoben:      dto.Erhoben,
			AlterStunden: dto.AlterStunden,
		})
	}

	return sichten
}

// anbietersichten macht aus den Anbietern ihre Sichten.
func anbietersichten(alle []app.KIAnbieterDTO) []mcp.KIAnbietersicht {
	sichten := make([]mcp.KIAnbietersicht, 0, len(alle))
	for _, einer := range alle {
		sichten = append(sichten, mcp.KIAnbietersicht(einer))
	}

	return sichten
}

// KIBatch lässt ein Modell für mehrere Karten vorschlagen (A13.1, #1452).
//
// **Ohne `uebernehmen`**: ein Batch, den ein Modell in einem Zug übernehmen
// dürfte, wäre die Vollmacht, gegen die es Nachweise gibt
// ([0105](../../docs/entscheidungen/0105-eine-freigabe-fuer-einen-batch.md)).
//
// **Und ohne Nachweis**: einen stellt nur das Fenster aus, und ein
// Cloudanbieter scheitert hier mit genau diesem Satz.
func (p *preise) KIBatch(ctx context.Context, bezug mcp.KIBatchbezug) (
	mcp.KIBatchsicht, error,
) {
	antwort := p.kianwendung().KIBatchVorschlagen(ctx,
		bezug.KarteIDs, bezug.Anbieter, bezug.Alias, bezug.Auftrag, "")
	if antwort.Fehler != nil {
		return mcp.KIBatchsicht{}, antwort.Fehler.AlsFehler()
	}

	ziele := make([]mcp.KIBatchlagensicht, 0, len(antwort.Ziele))
	for _, lage := range antwort.Ziele {
		ziele = append(ziele, alsBatchlagensicht(lage))
	}

	return mcp.KIBatchsicht{
		Kennung: antwort.Kennung, Anbieter: antwort.Anbieter, Ziele: ziele,
	}, nil
}

// alsBatchlagensicht übersetzt die Lage eines Ziels.
//
// **Der Fehler wird zum Satz und nicht zum Verbund**: ein Sprachmodell liest
// einen Satz, und die Struktur daneben verführt dazu, den Code zu deuten
// statt den Grund.
func alsBatchlagensicht(lage app.KIBatchlage) mcp.KIBatchlagensicht {
	sicht := mcp.KIBatchlagensicht{
		KarteID: lage.Zielkarte, Kennung: lage.Kennung, Modell: lage.Modell,
		Befunde: lage.Befunde, Ablauf: lage.Ablauf,
	}

	if lage.Fehler != nil {
		sicht.Fehler = lage.Fehler.Meldung
	}

	for _, feld := range lage.Felder {
		sicht.Felder = append(sicht.Felder, mcp.KIFeldwertsicht{
			Feld: feld.Feld, Werte: feld.Werte, Bisher: feld.Bisher,
		})
	}

	return sicht
}
