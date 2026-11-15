package app

// **Der Abzug eines Tisches geht an die Platte** — PLAY-05.AK-03 und AK-04
// (#1713).
//
// **Auf `Bindungen` und nicht auf `App`**, wie `SetSpeichernUnter`: der Weg
// fängt mit einem Dialog an, und ein Dialog braucht ein Fenster. Was danach
// geschieht, braucht keines und steht im Dienst.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Tischabzugschreiber legt einen Abzug an einen Ort — die Naht zum Dienst.
type Tischabzugschreiber interface {
	Schreibe(datenURL, pfad string) error

	// SchreibePDF legt ihn als einseitiges PDF auf ein Blatt — PLAY-05.AK-03
	// (#1713).
	SchreibePDF(datenURL, pfad, blatt string) error
}

// TischabzugAntwort ist der Ausgang eines Abzugs.
type TischabzugAntwort struct {
	Pfad string `json:"pfad"`

	// Abgebrochen sagt, dass niemand einen Ort gewählt hat.
	//
	// **Und das ist kein Fehler** — dieselbe Unterscheidung wie bei
	// `SetSpeichernUnter`: wer den Dialog wegklickt, hat sich entschieden.
	Abgebrochen bool `json:"abgebrochen"`

	Fehler *Fehler `json:"fehler"`
}

// TischabzugSpeichern fragt nach einem Ort und legt den Abzug dorthin.
//
// **Der Abzug kommt fertig aus dem Fenster.** Was die Leinwand gerade zeigt —
// mit Kamera, Drehung und gewähltem Tischbild —, weiß nur sie; ihn auf dieser
// Seite nachzubauen hieße, den ganzen Zeichenweg ein zweites Mal zu haben.
// **Ob es ein Bild ist, entscheidet trotzdem die Domäne** ([domain.DeuteAbzug]):
// was von außen kommt, wird geprüft und nicht geglaubt (§8).
func (b *Bindungen) TischabzugSpeichern(ctx context.Context, datenURL string) TischabzugAntwort {
	ctx = b.ruf(ctx)
	texte := DialogtexteFuer(b.app(ctx).sprache)

	pfad, err := b.app(ctx).waehleDatei(ctx, Dateiauswahl{
		Titel:     texte.TischabzugSpeichern,
		Filter:    texte.TischabzugFilter,
		Muster:    "*.png",
		Vorschlag: texte.TischabzugVorschlag,
		Speichern: true,
	})
	if err != nil {
		return TischabzugAntwort{Fehler: Uebersetze(ctx, err)}
	}

	if pfad == "" {
		return TischabzugAntwort{Abgebrochen: true}
	}

	return b.app(ctx).TischabzugSchreiben(ctx, datenURL, pfad)
}

// TischabzugPDFSpeichern fragt nach einem Ort und legt den Abzug als PDF
// dorthin.
//
// **Zwei Wege und nicht ein Weg mit einem Schalter.** Der Dialog trägt den
// Filter, und ein `*.png`-Dialog, der eine PDF schriebe, wäre eine
// Überraschung — dieselbe Überlegung wie bei den drei Standbindungen einer
// Partie (§16).
func (b *Bindungen) TischabzugPDFSpeichern(
	ctx context.Context,
	datenURL, blatt string,
) TischabzugAntwort {
	ctx = b.ruf(ctx)
	texte := DialogtexteFuer(b.app(ctx).sprache)

	pfad, err := b.app(ctx).waehleDatei(ctx, Dateiauswahl{
		Titel:     texte.TischabzugPDFSpeichern,
		Filter:    texte.TischabzugPDFFilter,
		Muster:    "*.pdf",
		Vorschlag: texte.TischabzugPDFVorschlag,
		Speichern: true,
	})
	if err != nil {
		return TischabzugAntwort{Fehler: Uebersetze(ctx, err)}
	}

	if pfad == "" {
		return TischabzugAntwort{Abgebrochen: true}
	}

	return b.app(ctx).TischabzugPDFSchreiben(ctx, datenURL, pfad, blatt)
}

// TischabzugPDFSchreiben legt den Abzug als PDF an einen genannten Ort.
func (a *App) TischabzugPDFSchreiben(
	ctx context.Context, datenURL, pfad, blatt string,
) TischabzugAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.tischabzug.SchreibePDF(datenURL, pfad, blatt); fehler != nil {
		return TischabzugAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return TischabzugAntwort{Pfad: pfad}
}

// TischabzugSchreiben legt den Abzug an einen genannten Ort.
func (a *App) TischabzugSchreiben(
	ctx context.Context, datenURL, pfad string,
) TischabzugAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.tischabzug.Schreibe(datenURL, pfad); fehler != nil {
		return TischabzugAntwort{Fehler: Uebersetze(ctx, fehler)}
	}

	return TischabzugAntwort{Pfad: pfad}
}

// MitTischabzugschreiber setzt die Naht zur Ausgabe.
func (a *App) MitTischabzugschreiber(schreiber Tischabzugschreiber) {
	a.tischabzug = schreiber
}

// tischabzugDienst gibt den Dienst, den die Composition Root einsetzt.
func tischabzugDienst() Tischabzugschreiber {
	return service.NeueTischbildausfuhr()
}
