package app

// **Spieltexturen an der Naht** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Warum eine Adresse und keine Bytes
//
// **Weil 182 KB als Base64 durch die Wails-Brücke 554 ms kosten** — gemessen
// für die Vorschau (#1270),
// und ein Spieltisch zeigt nicht eine Karte, sondern sechzig. Das Issue sagt
// es ausdrücklich: „keine Bild-Base64-Rückkehr durch die Wails-Naht."
//
// # Und warum der Schlüssel in der Antwort steht
//
// **Damit das Fenster teilen kann.** Zwei Objekte mit demselben Schlüssel sind
// dieselbe Textur — sieben Wälder am Tisch sind ein Bild und sieben
// Darstellungen davon. Ohne den Schlüssel müsste das Fenster die Adresse
// vergleichen, und die ist dieselbe Auskunft in schlechterer Form.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/vorschauweg"
)

// Spielbildpfad ist der Anfang jeder Spielbildadresse — [vorschauweg.Spielpfad].
const Spielbildpfad = vorschauweg.Spielpfad

// TexturDTO ist eine Spieltextur, wie das Fenster sie bekommt.
type TexturDTO struct {
	// Schluessel ist, woran zwei gleiche Texturen sich erkennen.
	Schluessel string `json:"schluessel"`

	Adresse string `json:"adresse"`
	Guete   string `json:"guete"`
	Seite   string `json:"seite"`

	Breite int `json:"breite"`
	Hoehe  int `json:"hoehe"`

	// Befunde sagen, **warum** eine Karte behilflich aussieht.
	Befunde []string `json:"befunde"`

	// Zweiseitig sagt, ob diese Karte eine zweite öffentliche Seite führt —
	// dann lässt sie sich umdrehen und zeigt etwas anderes.
	Zweiseitig bool `json:"zweiseitig"`
}

// TexturAntwort ist der Ausgang eines Texturwegs.
type TexturAntwort struct {
	Textur TexturDTO `json:"textur"`

	Fehler *Fehler `json:"fehler"`
}

// Texturwerk ist, was diese Naht zum Zeichnen braucht.
//
// **Zwei Verfahren und nicht eines mit einem Schalter**: eine Kartenseite
// gehört einer Karte, ein Rücken einem Spiel. Wer beides in einen Ruf legte,
// hätte einen Parameter, der die Hälfte der Zeit nichts bedeutet.
type Texturwerk interface {
	Textur(ctx context.Context, karte *domain.CardRow,
		seite service.Kartenseite, guete service.Texturguete) (
		service.Textur, error)

	Ruecken(ctx context.Context, spiel string, guete service.Texturguete) (
		service.Textur, error)
}

// Spieltextur gibt die Adresse einer **Kartenseite** für den Spieltisch.
//
// **Rein lesend.** Sie zeichnet, und das kostet Rechenzeit — geschrieben wird
// dabei nichts: kein Bestand, keine Datei, keine Zeile. Was entsteht, liegt im
// Arbeitsspeicher dieses Prozesses und fällt wieder heraus.
//
// **Den Rücken gibt [App.Rueckentextur]** — er gehört einem Spiel und braucht
// keine Karte.
func (a *App) Spieltextur(
	ctx context.Context, karteID string, seite, guete string,
) TexturAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungLesen); fehler != nil {
		return leereTextur(ctx, fehler)
	}

	kennung, nutzerfehler := benannteKennungAus("karteID", karteID)
	if nutzerfehler != nil {
		return TexturAntwort{
			Textur: TexturDTO{Befunde: []string{}}, Fehler: nutzerfehler,
		}
	}

	karte, fehler := a.kartenleser.KarteLesen(ctx, kennung)
	if fehler != nil {
		return leereTextur(ctx, fehler)
	}

	textur, fehler := a.texturwerk().Textur(ctx, karte,
		service.Kartenseite(seite), service.Texturguete(guete))
	if fehler != nil {
		return leereTextur(ctx, fehler)
	}

	return alsTexturantwort(textur, seite)
}

// Rueckentextur gibt die Adresse des Kartenrückens eines Spiels.
//
// # Warum ein eigener Weg
//
// **Weil ein verdecktes Objekt keine Kartenkennung trägt.** Die Projektion
// aus G1 hält sie zurück, und zwar ausdrücklich — eine Bibliothek ist auch
// für ihren Eigner verdeckt. Ein Rücken, der eine Kennung verlangte, wäre
// genau dort nicht zu bekommen, wo er gebraucht wird.
//
// **Und weil er keine bekommen darf.** Er ist für jede Karte derselbe; wer
// ihn aus einer Karte baute, bände Maße und Schlüssel an sie — und verriete
// sie über die Metadaten (#1459: „fremde verdeckte Karten nicht über
// Texturmetadaten verraten").
//
// **Das Spiel ist kein Geheimnis**: es steht in jeder Partiesicht, für jeden
// Sitz und jeden Zuschauer.
func (a *App) Rueckentextur(
	ctx context.Context, spiel, guete string,
) TexturAntwort {
	ctx = a.imProfil(ctx)

	if fehler := a.darf(ctx, domain.WirkungLesen); fehler != nil {
		return leereTextur(ctx, fehler)
	}

	textur, fehler := a.texturwerk().Ruecken(
		ctx, spiel, service.Texturguete(guete))
	if fehler != nil {
		return leereTextur(ctx, fehler)
	}

	return alsTexturantwort(textur, service.MarkeRuecken)
}

// alsTexturantwort packt eine Textur in die Antwort der Naht.
func alsTexturantwort(textur service.Textur, seite string) TexturAntwort {
	return TexturAntwort{Textur: TexturDTO{
		Schluessel: textur.Schluessel,
		Adresse:    textur.Adresse,
		Guete:      string(textur.Guete),
		Seite:      seite,
		Breite:     textur.Breite,
		Hoehe:      textur.Hoehe,
		Befunde:    textur.Befunde,
		Zweiseitig: textur.Zweiseitig,
	}}
}

// leereTextur gibt eine Antwort mit leeren Listen und dem übersetzten Fehler.
//
// **Leere Listen und kein `nil`** — dieselbe Zusage wie auf dem Erfolgsweg.
func leereTextur(ctx context.Context, fehler error) TexturAntwort {
	return TexturAntwort{
		Textur: TexturDTO{Befunde: []string{}},
		Fehler: Uebersetze(ctx, fehler),
	}
}

// MitTexturwerk setzt es — für Läufe ohne Bestand (§3).
func (a *App) MitTexturwerk(werk Texturwerk) *App {
	a.texturen = werk

	return a
}

// texturwerk gibt das eingesetzte Werk.
func (a *App) texturwerk() Texturwerk { return a.texturen }
