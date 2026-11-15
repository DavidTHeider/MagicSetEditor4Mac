package app

// **Vom Vorschlag in die Karte** — Werkstück A12 von Phase 5.3
// (#1451).
//
// # Derselbe Weg wie eine Datei von der Platte
//
// Das Issue verlangt: „Übernahme nutzt vorhandene Crop-/Ablage-/Renderfunktion,
// Herkunft stimmt."
//
// **Deshalb steht hier kein zweiter Zuschnitt.** Der Vorschlag bekommt seinen
// Zuschnitt über denselben [App.vorbereiten] wie jede Bilddatei, und
// eingesetzt wird über denselben Weg — nur dass er zusätzlich die Herkunft
// schreibt und den Vorschlag danach wegnimmt.
//
// # Und das Original bleibt, bis bestätigt ist
//
// „Original bis bestätigter Übernahme erhalten." Der Vorrat gibt die Bytes
// heraus und behält sie; weggenommen wird der Eintrag erst, **nachdem** das
// Feld steht.

import (
	"context"
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// KIBildZuschnitt bereitet den Zuschnitt eines wartenden Vorschlags vor.
func (a *App) KIBildZuschnitt(
	ctx context.Context, karteID, feld, kennung string,
) ZuschnittAntwort {
	ctx = a.imProfil(ctx)

	_, _, breite, hoehe, befund := a.bildziel(ctx, karteID, feld)
	if befund != nil {
		return ZuschnittAntwort{Fehler: befund}
	}

	inhalt, _, fehler := a.wartendeBytes(ctx, kennung)
	if fehler != nil {
		return ZuschnittAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	bild, _, fehler := msepaket.DekodiereBild(inhalt)
	if fehler != nil {
		return ZuschnittAntwort{Fehler: Uebersetze(ctx, msepaket.ZuGrossOder(fehler, domain.MitGrund(
			domain.ErrValidation, fehler, "das wartende Bild auspacken")))}
	}

	// **Die Kennung steht im Pfad** — das Fenster reicht sie zurück, und
	// eingesetzt wird über [App.KIBildEinsetzen] und nicht über die Datei.
	return a.vorbereitenAus(ctx, bild, kennung, breite, hoehe)
}

// KIBildEinsetzen schneidet zu, legt ab und schreibt Feld **und Herkunft**.
func (a *App) KIBildEinsetzen(
	ctx context.Context, setID, karteID, feld, kennung string, auswahl AuswahlDTO,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	profil, fehler := storage.ProfilAus(ctx)
	if fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	setkennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	zielkarte, karte, breite, hoehe, befund := a.bildziel(ctx, karteID, feld)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	inhalt, eintrag, fehler := a.wartendeBytes(ctx, kennung)
	if fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	// **Der Vorschlag gehört zu diesem Feld und zu keinem anderen.** Die
	// Kennung allein sagte nur, dass er diesem Profil gehört; ein Bild, das
	// für die Illustration erzeugt wurde, gehört nicht in das Wasserzeichen.
	if eintrag.Besitz.Karte != zielkarte || eintrag.Besitz.Feld != feld {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: dieser Vorschlag gehört zum Feld %q und nicht zu %q",
			domain.ErrValidation, eintrag.Besitz.Feld, feld))}
	}

	bild, _, fehler := msepaket.DekodiereBild(inhalt)
	if fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, msepaket.ZuGrossOder(fehler, domain.MitGrund(
			domain.ErrValidation, fehler, "das wartende Bild auspacken")))}
	}

	flaeche := render.Zuschnittwahl{
		Quelle:  bild.Bounds(),
		Ziel:    image.Pt(breite, hoehe),
		Auswahl: alsRechteck(auswahl),
	}.Begrenze(render.BevorzugtKeines)

	name, fehler := a.bildablage.Ablegen(karte.SetID, feld,
		render.ZuschnittAus(bild, flaeche, breite, hoehe))
	if fehler != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return a.bildherkunftSchreiben(ctx, bildeinsatz{
		setID: setkennung, karteID: zielkarte, setDerKarte: karte.SetID,
		feld: feld, name: name, kennung: kennung, profil: profil,
		herkunft: eintrag.Herkunft, erwartet: eintrag.Besitz.Erwartet,
	})
}

// bildeinsatz sind die Angaben, die das Schreiben braucht.
//
// **Als Verbund und nicht als acht Argumente**: eine Funktion mit acht
// gleichartigen Werten lässt sich an der Aufrufstelle nicht mehr lesen.
type bildeinsatz struct {
	setID       int64
	karteID     int64
	setDerKarte int64
	feld        string
	name        string
	kennung     string
	profil      int64
	herkunft    domain.KIBildherkunft

	// erwartet ist der Stand, den der Bildauftrag eingefroren hat — leer bei
	// einem Vorschlag aus der Zeit vor Befund 2 des Nachlaufs.
	erwartet string
}

// bildherkunftSchreiben setzt das Feld samt Herkunft und räumt den Vorrat.
func (a *App) bildherkunftSchreiben(ctx context.Context, einsatz bildeinsatz) SchreibAntwort {
	werte := []domain.Value{{Field: einsatz.feld, Text: einsatz.name}}

	feldherkunft := einsatz.herkunft.AlsFeldherkunft(einsatz.feld,
		domain.KIInhaltsbezug([]string{einsatz.name}))

	// **Mit dem Stand, den der Auftrag eingefroren hat** (Befund 2 des
	// Nachlaufs über Phase 5.3).
	//
	// **Hier stand das Gegenteil, und die Begründung war falsch**: „der
	// Bildvorschlag friert keinen Stand ein". Er tut es — [KIAuftragswerk.Anlegen]
	// liest die Revision, auch für ein Bild, und ein Bildauftrag dauert
	// Minuten. Ohne die Bedingung schriebe das Einsetzen über eine Karte,
	// die es so nicht mehr gibt; mit ihr gilt für den Bildweg dieselbe
	// Zusage wie für Text und Gruppe.
	//
	// **Der Vorschlag bleibt bei einem Konflikt liegen** — der Vorrat wird
	// erst nach dem Schreiben geräumt, die abgelegte Datei nimmt der
	// Fehlerweg darunter wieder weg. Ein zweiter, bewusster Versuch ist
	// damit möglich.
	_, _, fehler := a.schreiber.FelderVonAussenMitHerkunft(
		ctx, einsatz.setID, einsatz.karteID, werte, kiherkunft,
		[]domain.KIFeldherkunft{feldherkunft},
		map[string][]domain.KIBildparameter{einsatz.feld: einsatz.herkunft.Parameter},
		einsatz.erwartet)
	if fehler != nil {
		// **Das Bild liegt schon und hat noch keinen Verweis** — dieselbe
		// Stelle wie in [App.bildAblegen] (#1163, Z7-02).
		return SchreibAntwort{Fehler: Uebersetze(ctx, errors.Join(
			fehler, a.bildablage.Wegnehmen(einsatz.setDerKarte, einsatz.name)))}
	}

	// **Jetzt erst fällt der Vorschlag.** Vorher wäre er weg, während das
	// Feld noch nichts trägt — und wer den Fehler sähe, hätte nichts mehr,
	// womit er es erneut versuchen könnte.
	//
	// **Und ein Fehler dabei hält nichts auf** (§7, Regel 3): geschrieben
	// ist geschrieben. Was liegen bleibt, räumt die Frist weg.
	_ = a.bildvorrat.Verwerfen(einsatz.kennung, einsatz.profil) //nolint:errcheck // siehe oben

	return SchreibAntwort{}
}

// wartendeBytes holt Inhalt und Eintrag aus dem Vorrat.
func (a *App) wartendeBytes(ctx context.Context, kennung string) (
	[]byte, service.Bildeintrag, error,
) {
	if a.bildvorrat == nil {
		return nil, service.Bildeintrag{}, fehlenderVorrat()
	}

	profil, fehler := storage.ProfilAus(ctx)
	if fehler != nil {
		return nil, service.Bildeintrag{}, fehler
	}

	return a.bildvorrat.Lesen(kennung, profil)
}

// fehlenderVorrat ist der Satz, wenn diese Anwendung keinen Bildvorrat hat.
func fehlenderVorrat() error {
	return fmt.Errorf(
		"%w: diese Anwendung hat keinen Bildvorrat — ohne einen gibt es "+
			"keine Bildvorschläge", domain.ErrValidation)
}

// bildvorlage gibt die Bytes des Bildes, das bearbeitet werden soll.
//
// # Warum das Feld gelesen wird und nicht die Vorschau
//
// **Weil bearbeitet wird, was in der Karte steht.** Die Vorschau ist
// gerastert und auf Bildschirmgröße gebracht; sie an einen Dienst zu
// schicken hieße, ein verkleinertes Bild zu bearbeiten und das Ergebnis für
// das Original zu halten.
//
// **Ohne Bearbeitungswunsch gibt es keine Vorlage** — und dann erzeugt der
// Dienst eines aus nichts als dem Prompt.
func (a *App) bildvorlage(
	ctx context.Context, karteID int64, auftrag domain.KIBildauftrag,
) ([]byte, error) {
	if !auftrag.Bearbeiten {
		return nil, nil
	}

	if a.kartenleser == nil {
		return nil, fehlenderBestand()
	}

	karte, fehler := a.kartenleser.KarteLesen(ctx, karteID)
	if fehler != nil {
		return nil, fehler
	}

	name, hat := karte.Value(auftrag.Feld)
	if !hat || name == "" {
		return nil, fmt.Errorf(
			"%w: im Feld %q steht kein Bild, das sich bearbeiten ließe",
			domain.ErrValidation, auftrag.Feld)
	}

	ort := ""
	if a.bildablage != nil {
		ort = a.bildablage.Verzeichnis(karte.SetID)
	}

	if ort == "" {
		return nil, fmt.Errorf(
			"%w: diese Anwendung hat keinen Ort für eingesetzte Bilder",
			domain.ErrValidation)
	}

	// **`filepath.Base` und nicht der rohe Name**: was im Feld steht, kommt
	// aus einem Paket, und ein Paket ist eine fremde Datei (§8). Ein
	// `../../` darin führte sonst aus dem Ablageort hinaus.
	//
	// #nosec G304 -- siehe oben
	inhalt, fehler := os.ReadFile(filepath.Join(ort, filepath.Base(name)))
	if fehler != nil {
		return nil, domain.MitGrund(domain.ErrNotFound, fehler,
			"das Bild aus %q lesen", auftrag.Feld)
	}

	return inhalt, nil
}
