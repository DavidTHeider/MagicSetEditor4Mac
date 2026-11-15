package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Spalten der Kartenliste und ihre Ordnung** — Werkstück 3 von
// Phase 3.7
// (#1168).
//
// # Was hier zusammenkommt
//
//	FOR_EACH(f, set->game->card_fields) {
//	  ColumnSettings& cs = settings.columnSettingsFor(*set->game, *f);
//	  if (cs.visible && f->card_list_allow) {
//	    new_column_fields[cs.position] = f;
//	  }
//	}
//
// `CardListBase::rebuild`, `src/gui/control/card_list.cpp`, Zeilen 205–253,
// Fassung f9d9356 vom 2021-07-06. Drei Quellen treffen sich: die **Felder des
// Spiels**, die **Wahl der Nutzenden** und die **Karten**. MSE hält die Wahl je
// Spiel (`GameSettings`), und deshalb überlebt sie hier das Set.
//
// # `sort script:` wird ausgewertet
//
// MSE ordnet über `getSortKey()`, und das gibt `sort_value` statt des Wertes,
// wo ein Feld ein Sortierskript hat (`src/data/field.hpp`, Zeilen 234–237).
// In `magic.mse-game` sind das **15** Felder — vier Farbfelder, vier
// Indikatoren, drei Seltenheiten und die Kartennummer.
//
// **Seit dem 2026-08-28** (#1232)
// rechnet `listenwerte.go` sie — für die Spalte, nach der geordnet wird, und
// nur für sie.

// Listenwahlablage merkt sich, was jemand gewählt hat.
type Listenwahlablage interface {
	Spalten(ctx context.Context, spiel string) ([]domain.Spaltenwahl, error)
	SpaltenSetzen(ctx context.Context, spiel string, wahl []domain.Spaltenwahl) error
	Sortierung(ctx context.Context, spiel string) (domain.Kartenordnung, error)
	SortierungSetzen(ctx context.Context, spiel string, ordnung domain.Kartenordnung) error
}

// Kartenlistenbau baut die Liste eines Sets.
type Kartenlistenbau struct {
	koepfe Kopfliste
	karten Ersetzungskarten
	felder Kartenfeldquelle

	// spielskript gibt das `init script:` des Spiels — oder nichts.
	//
	// **Eine eigene Schnittstelle neben [Kartenfeldquelle]**, weil sie eine
	// andere Frage beantwortet und weil nicht jeder Aufrufer sie hat: ohne
	// Spielskript rechnet die Liste, was sie ohne rechnen kann (#1238).
	spielskript Spielskriptquelle
	wahl        Listenwahlablage

	// marken gibt die Marken je Karte — oder ist nil, dann bleiben die Zeilen
	// ohne (#1208).
	//
	// **Nil ist erlaubt**, weil eine Liste ohne Marken eine Liste ist: die
	// Kommandozeile baut sie über denselben Weg, und ein Test, der die Spalten
	// prüft, braucht keinen Bestand voller Marken.
	marken Markenjekarte
}

// Markenjekarte gibt die Marken aller Karten eines Sets — in **einem** Ruf.
//
// **Eine eigene Schnittstelle und nicht die ganze [Markenpflege]**: die Liste
// setzt keine Marke, und §2 sagt, die Schnittstelle steht dort, wo sie
// gebraucht wird — mit dem, was gebraucht wird.
type Markenjekarte interface {
	JeKarte(ctx context.Context, setID int64) (map[int64][]string, error)
}

func NeuerKartenlistenbau(
	koepfe Kopfliste, karten Ersetzungskarten,
	felder Kartenfeldquelle, wahl Listenwahlablage, spielskript Spielskriptquelle,
) *Kartenlistenbau {
	return &Kartenlistenbau{
		koepfe: koepfe, karten: karten, felder: felder,
		wahl: wahl, spielskript: spielskript,
	}
}

// MitMarken gibt dem Listenbau die Marken der Karten (#1208).
func (k *Kartenlistenbau) MitMarken(marken Markenjekarte) *Kartenlistenbau {
	k.marken = marken

	return k
}

// Aufbau gibt Spalten, Zeilen und die wirksame Ordnung.
func (k *Kartenlistenbau) Aufbau(ctx context.Context, setID int64) (
	domain.Kartenlistenaufbau, error,
) {
	spiel, err := k.spielVonSet(ctx, setID)
	if err != nil {
		return domain.Kartenlistenaufbau{}, err
	}

	felder, ohneVorlage := k.felderOderKeine(ctx, setID)

	gewaehlt, err := k.wahl.Spalten(ctx, spiel)
	if err != nil {
		return domain.Kartenlistenaufbau{}, fmt.Errorf("die Spaltenwahl lesen: %w", err)
	}

	ordnung, err := k.wahl.Sortierung(ctx, spiel)
	if err != nil {
		return domain.Kartenlistenaufbau{}, fmt.Errorf("die Sortierung lesen: %w", err)
	}

	karten, err := k.karten.Karten(ctx, setID)
	if err != nil {
		return domain.Kartenlistenaufbau{}, fmt.Errorf("die Karten lesen: %w", err)
	}

	if ohneVorlage {
		felder = felderAusKarten(karten)
	}

	spalten := spaltenAus(felder, gewaehlt)

	// **Einmal je Liste, nicht je Karte.** Das `init script:` von
	// `magic.mse-game` legt die Funktionen an, die die Feldskripte rufen —
	// `card_full_name()`, `cmc(…)`. Es je Karte zu fahren kostete bei 3.788
	// Karten das Vielfache (#1238).
	skript, setfelder := "", []domain.Field(nil)
	if k.spielskript != nil {
		//nolint:errcheck // ohne Skript wird gerechnet, was ohne geht
		skript, _ = k.spielskript.Spielskript(ctx, setID)
		setfelder, _ = k.spielskript.Setfelder(ctx, setID) //nolint:errcheck // dito
	}

	wirksam := wirksameOrdnung(ordnung, spalten)

	aufbau := domain.Kartenlistenaufbau{
		Spalten:     spalten,
		Zeilen:      zeilenAus(ctx, karten, felder, spalten, skript, setfelder, wirksam),
		Ordnung:     wirksam,
		OhneVorlage: ohneVorlage,
	}

	// **Abgebrochen ist nicht fertig**: [gerechneteZellen] hört beim Abbruch
	// auf, und die Zeilen stünden mit leeren Zellen als vollständig da.
	if err := ctx.Err(); err != nil {
		return domain.Kartenlistenaufbau{}, err
	}

	// **Ein Lesefehler hält die Liste nicht auf** (§8): die Karten sind der
	// Grund, warum jemand hier ist, und eine Zeile ohne Marken ist eine Zeile.
	if k.marken != nil {
		//nolint:errcheck // eine Liste ohne Marken ist eine Liste
		jeKarte, _ := k.marken.JeKarte(ctx, setID)
		for stelle := range aufbau.Zeilen {
			aufbau.Zeilen[stelle].Marken = jeKarte[aufbau.Zeilen[stelle].KarteID]
		}
	}

	ordne(aufbau.Zeilen, aufbau.Ordnung)

	return aufbau, nil
}

// SpaltenSetzen schreibt die Wahl fort — geprüft gegen die Felder des Spiels.
func (k *Kartenlistenbau) SpaltenSetzen(
	ctx context.Context, setID int64, wahl []domain.Spaltenwahl,
) error {
	spiel, felder, err := k.spielUndFelder(ctx, setID)
	if err != nil {
		return err
	}

	// **Ohne Vorlage wird nicht geprüft.** Sonst ließe sich in einem Set ohne
	// Spielpaket keine Spalte wählen, und die Duldung im Aufbau wäre nur die
	// halbe.
	for _, eine := range wahl {
		if felder != nil && !waehlbar(felder, eine.Feld) {
			return fmt.Errorf("%w: das Feld %q gibt es in diesem Spiel nicht",
				domain.ErrValidation, eine.Feld)
		}
	}

	return k.wahl.SpaltenSetzen(ctx, spiel, wahl)
}

// SortierungSetzen schreibt sie fort — ebenso geprüft.
//
// **„Nach nichts sortieren" ist erlaubt**: das ist MSEs `sort_by_column = -1`.
func (k *Kartenlistenbau) SortierungSetzen(
	ctx context.Context, setID int64, ordnung domain.Kartenordnung,
) error {
	spiel, felder, err := k.spielUndFelder(ctx, setID)
	if err != nil {
		return err
	}

	if ordnung.Feld != "" && felder != nil && !waehlbar(felder, ordnung.Feld) {
		return fmt.Errorf("%w: das Feld %q gibt es in diesem Spiel nicht",
			domain.ErrValidation, ordnung.Feld)
	}

	return k.wahl.SortierungSetzen(ctx, spiel, ordnung)
}

// spielUndFelder gibt beides — die zwei Fragen, die jede Wahl stellt.
func (k *Kartenlistenbau) spielUndFelder(ctx context.Context, setID int64) (
	string, []domain.Field, error,
) {
	spiel, err := k.spielVonSet(ctx, setID)
	if err != nil {
		return "", nil, err
	}

	felder, _ := k.felderOderKeine(ctx, setID)

	return spiel, felder, nil
}

// felderOderKeine gibt die Kartenfelder des Spiels — oder keine.
//
// **`nil` heißt: keine Vorlage**, und das ist kein Fehler. Dieselbe Duldung
// wie beim Ersetzen ([Ersetzer.textfelder]) und beim Feldschreiben
// ([Feldweg.FelderVonAussen]), aus demselben Grund: ein Set aus MTG JSON
// bliebe sonst unbearbeitbar, solange sein Spielpaket fehlt — und ohne
// Kartenliste ist es unerreichbar.
func (k *Kartenlistenbau) felderOderKeine(ctx context.Context, setID int64) (
	[]domain.Field, bool,
) {
	felder, err := k.felder.Kartenfelder(ctx, setID)
	if err != nil || len(felder) == 0 {
		return nil, true
	}

	return felder, false
}

// felderAusKarten baut Ersatzfelder aus dem, was die Karten führen.
//
// **In der Reihenfolge ihres ersten Auftretens** — das ist die Reihenfolge des
// Pakets, wie die Karte sie mitbringt, und die einzige, die ohne Vorlage
// überhaupt zu haben ist.
//
// **Die erste ist sichtbar, die übrigen nicht.** Ohne Vorschlag des Pakets ist
// das die einzige Wahl, die eine Liste brauchbar lässt: alle sichtbar wären
// bei einem MTG-JSON-Set über zwanzig Spalten, keine sichtbar wäre eine leere
// Liste.
func felderAusKarten(karten []*domain.CardRow) []domain.Field {
	gesehen := map[string]bool{}
	felder := []domain.Field{}

	for _, karte := range karten {
		for _, wert := range karte.Values {
			if gesehen[wert.Field] {
				continue
			}

			gesehen[wert.Field] = true

			felder = append(felder, domain.Field{
				Name: wert.Field,
				Listenangaben: domain.Listenangaben{
					Listenerlaubt:  true,
					Listensichtbar: len(felder) == 0,
					Listenstelle:   len(felder) + 1,
				},
			})
		}
	}

	return felder
}

// spielVonSet gibt den Spielnamen — die Wahl gehört ihm und nicht dem Set.
func (k *Kartenlistenbau) spielVonSet(ctx context.Context, setID int64) (string, error) {
	koepfe, err := k.koepfe.Setkoepfe(ctx)
	if err != nil {
		return "", fmt.Errorf("den Setkopf lesen: %w", err)
	}

	for _, kopf := range koepfe {
		if kopf.ID == setID {
			return kopf.Game, nil
		}
	}

	return "", fmt.Errorf("%w: das Set %d führt der Bestand nicht",
		domain.ErrNotFound, setID)
}

// Spielskriptquelle gibt das `init script:` des Spiels zu einem Set.
//
// **Eine eigene Schnittstelle und keine Erweiterung von
// [Kartenfeldquelle]**: sie beantwortet eine andere Frage, und die
// Flächennaht (`listennahtflaechen.go`) hat keine Antwort darauf — dort
// rechnet die Liste, was sie ohne Spielskript rechnen kann.
type Spielskriptquelle interface {
	Spielskript(ctx context.Context, setID int64) (string, error)

	// Setfelder sind die `set field:`-Blöcke des Spiels.
	//
	// **Sie gehören dazu, und das ist gemessen**: das `init script` von
	// `magic.mse-game` liest in Zeile 10583 `set.custom_super_types`. Ein
	// fehlendes **Glied** bricht MSE-Script ab — anders als ein fehlender
	// Name lässt es sich nicht mit `or else` fangen. Ohne die 89 Setfelder
	// bricht das Skript dort, und alles danach fehlt.
	Setfelder(ctx context.Context, setID int64) ([]domain.Field, error)
}
