package main

// **Was der Dienst über KI sagt** — herausgezogen für Werkstück A13.3 von
// Phase 5.3
// (#1452),
// als `sammlung.go` an die Zeilengrenze aus §2 kam.
//
// **Der Schnitt folgt der Frage**: dort steht, was der Dienst über
// **Sammlung und Preise** sagt, hier, was er über **Modelle** sagt — den
// Katalog, den Vorschlag und was daraus ins Portfolio geht.
//
// Die Naht selbst und ihre gehaltene Anwendung stehen weiter in
// `sammlung.go`: sie ist dieselbe für beide Fragen, und zwei Verbünde
// nebeneinander hätten zwei Anwendungen ([preise.kianwendung]).

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// KIModelle nennt, was über KI-Modelle feststeht (A6, #1445).
func (p *preise) KIModelle(ctx context.Context) (rest.KIKatalogsicht, error) {
	antwort := p.kianwendung().KIModelle(ctx)
	if antwort.Fehler != nil {
		return rest.KIKatalogsicht{}, antwort.Fehler.AlsFehler()
	}

	modelle := make([]rest.KIModellsicht, 0, len(antwort.Modelle))
	for _, modell := range antwort.Modelle {
		modelle = append(modelle, rest.KIModellsicht{
			Anbieter: modell.Anbieter, Alias: modell.Alias,
			Laufzeit: modell.Laufzeit, Ort: modell.Ort,
			Faehigkeiten: belegsichten(modell.Faehigkeiten),
			Herkunft:     gliedsichten(modell.Herkunft),
			Bindung:      modell.Bindung,

			Kontextfenster: modell.Kontextfenster, KontextBeleg: modell.KontextBeleg,
		})
	}

	return rest.KIKatalogsicht{
		NurLokal: antwort.NurLokal, Einrichtung: antwort.Einrichtung,
		Anbieter: anbietersichten(antwort.Anbieter),
		Cloud:    rest.KICloudsicht(antwort.Cloud),
		Befunde:  antwort.Befunde, Modelle: modelle, Vorrang: antwort.Vorrang,
	}, nil
}

// KIVorschlag lässt ein Modell Felder einer Karte vorschlagen (A7, #1446).
//
// **In einem Zug**, weil diese Fläche die Anwendung je Aufruf baut — der
// Grund steht an [app.App.KIVorschlagInEinemZug].
func (p *preise) KIVorschlag(ctx context.Context, bezug rest.KIVorschlagsbezug) (
	rest.KIVorschlagssicht, error,
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
		return rest.KIVorschlagssicht{}, antwort.Fehler.AlsFehler()
	}

	felder := make([]rest.KIFeldwertsicht, 0, len(antwort.Felder))
	for _, feld := range antwort.Felder {
		felder = append(felder, rest.KIFeldwertsicht(feld))
	}

	return rest.KIVorschlagssicht{
		Kennung: antwort.Kennung, Modell: antwort.Modell, Felder: felder,
		Befunde: antwort.Befunde, Ablauf: antwort.Ablauf,
		Uebernommen: antwort.Uebernommen,
		Anbieter:    antwort.Anbieter, Weg: antwort.Weg,
		Ausgang: antwort.Ausgang, Rueckfaelle: antwort.Rueckfaelle,
	}, nil
}

// belegsichten macht aus den Belegen ihre Sichten.
func belegsichten(belege []app.KIBelegDTO) []rest.KIBelegsicht {
	sichten := make([]rest.KIBelegsicht, 0, len(belege))
	for _, beleg := range belege {
		sichten = append(sichten, rest.KIBelegsicht(beleg))
	}

	return sichten
}

// gliedsichten macht aus der Kette ihre Sichten.
func gliedsichten(glieder []app.KIGliedDTO) []rest.KIGliedsicht {
	sichten := make([]rest.KIGliedsicht, 0, len(glieder))
	for _, glied := range glieder {
		sichten = append(sichten, rest.KIGliedsicht(glied))
	}

	return sichten
}

func (p *preise) Portfolio(ctx context.Context) ([]rest.Portfoliozeilensicht, int, error) {
	naht, err := app.SammlungsnahtUeber(p.db)
	if err != nil {
		return nil, 0, err
	}

	zeilen, ohnePreis, rechenfehler := naht.Portfolio(ctx)
	if rechenfehler != nil {
		return nil, 0, rechenfehler
	}

	sichten := make([]rest.Portfoliozeilensicht, 0, len(zeilen))

	for _, zeile := range zeilen {
		sichten = append(sichten, rest.Portfoliozeilensicht{
			Waehrung:       zeile.Waehrung,
			Wert:           zeile.Wert,
			Exemplare:      zeile.Exemplare,
			AeltesterPreis: zeile.AeltesterPreis,
			AlterStunden:   zeile.AlterStunden,
			Quelle:         zeile.Quelle,
		})
	}

	return sichten, ohnePreis, nil
}

// alsPreissichten übersetzt die DTOs in die Sicht dieser Fläche.
func alsPreissichten(dtos []app.PreisstandDTO) []rest.Preissicht {
	sichten := make([]rest.Preissicht, 0, len(dtos))

	for _, dto := range dtos {
		sichten = append(sichten, rest.Preissicht{
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
func anbietersichten(alle []app.KIAnbieterDTO) []rest.KIAnbietersicht {
	sichten := make([]rest.KIAnbietersicht, 0, len(alle))
	for _, einer := range alle {
		sichten = append(sichten, rest.KIAnbietersicht(einer))
	}

	return sichten
}
