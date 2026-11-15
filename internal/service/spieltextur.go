// **Der Texturvertrag des Spieltisches** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Was hier entsteht und was ausdrücklich nicht
//
// **Hier entsteht kein zweiter Kartenrenderer.** Der Plan kommt vom
// vorhandenen Planbauer, das Bild vom vorhandenen Rasterer; was dieser Dienst
// dazutut, ist die Frage **ob** gezeichnet werden muss — und die Antwort
// darauf ist ein Schlüssel.
//
// # Warum die Kartenkennung als Schlüssel nicht genügt
//
// **Weil sie weder portabel noch inhaltsadressiert ist** (#1459 wörtlich).
// Zwei Bestände geben derselben Karte verschiedene Zeilennummern; dieselbe
// Zeilennummer trägt nach einer Änderung etwas anderes. Ein Cache über ihr
// zeigte das Bild von vorhin — und zwar still, denn ein altes Bild sieht aus
// wie ein Bild.
//
// # Woraus der Schlüssel besteht
//
//  1. **Was gezeichnet wird** — [render.Planabdruck] über die Ebenen.
//  2. **Was in den Dateien steht, die der Plan nur nennt** — Bilder und
//     Schriften, über ihre Abdrücke. Der Plan trägt Pfade und Namen; ändert
//     sich die Datei dahinter, ändert er sich nicht.
//  3. **Welche Seite** — Vorder- oder Rückseite.
//  4. **Wie groß und wie gut** — Breite in Bildpunkten und Rasterstufe.
//  5. **Welcher Rasterer** — [render.Fassung].
//
// # Und warum die Pixelgrößen Messstartwerte sind
//
// **Weil das Issue es so sagt**: „Zielgrößen des Anhangs sind Messstartwerte,
// keine feste API." Sie stehen hier als Zahlen, damit man sie messen **kann**;
// wer sie ändert, ändert jeden Schlüssel — und das ist richtig so, denn ein
// Bild in anderer Größe ist ein anderes Bild.

package service

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Ressourcenquelle sagt, was in einer Datei steht, ohne sie zu zeichnen.
//
// **Eine Schnittstelle, weil `service` nicht weiß, wo Pakete und Schriften
// liegen** — dieselbe Überlegung wie bei [render.Bildquelle]. Was hier
// gebraucht wird, ist ein **Abdruck**: zwei gleiche Abdrücke heißen, dass
// dieselben Bytes gezeichnet würden.
type Ressourcenquelle interface {
	// Bildabdruck gibt den Abdruck des Bildes unter einem Paketpfad.
	Bildabdruck(pfad string) string

	// Schriftabdruck gibt den Abdruck der Schrift zu einem Namen.
	//
	// **Zu einem Namen und nicht zu einer Datei**: welche Datei das
	// Betriebssystem für „MPlantin" findet, ist Teil der Antwort und nicht
	// der Frage.
	Schriftabdruck(name string) string

	// Rueckenbild gibt den Paketpfad des Kartenrückens eines Spiels — leer,
	// wenn das Spielpaket keinen mitbringt.
	//
	// # Warum am Dateinamen und nicht am `icon:` des Pakets
	//
	// **Weil `icon:` in fünf von sieben Spielpaketen eine Beispielkarte
	// ist** — gemessen am 2026-09-15 über `Full-Magic-Pack`:
	// `archenemy`, `planechase`, `saga`, `vanguard` und `herospath` nennen
	// dort `card-sample`, also eine **Vorderseite**. Wer das Symbol nähme,
	// legte auf jede verdeckte Archenemy-Karte eine offene.
	//
	// **`card-back.png` dagegen ist einer** — `magic.mse-game` und
	// `space.mse-game` führen ihn, und beide zeigen wirklich einen Rücken
	// (angesehen, nicht am Namen erschlossen).
	Rueckenbild(spiel string) string
}

// Texturablage hält gerasterte Bilder unter ihrem Schlüssel.
type Texturablage interface {
	// Adresse gibt die Adresse zu einem Schlüssel samt seinen Maßen — und
	// ob es ihn gibt.
	//
	// **Die Maße kommen mit und werden nicht gerechnet.** Wer im Fach
	// fündig wird, muss dieselben Zahlen nennen wie der, der gezeichnet
	// hat; eine Rechnung daneben ginge um einen Bildpunkt auseinander, und
	// ein Treffer, der andere Zahlen nennt als der Fehltreffer, ist keiner
	// (#1459: „unveränderter Inhalt trifft ihn").
	Adresse(schluessel string) (adresse string, breite, hoehe int, da bool)

	// Hinterlege legt ein Bild samt seinen Maßen ab und gibt seine Adresse.
	Hinterlege(schluessel string, bild []byte, breite, hoehe int) string

	// Vergiss wirft ein Bild weg, auf das keine Karte mehr zeigt (E06).
	Vergiss(schluessel string)
}

// Texturplanbauer macht aus einer Karte einen Zeichenplan.
type Texturplanbauer interface {
	Kartenplan(ctx context.Context, karte *domain.CardRow) (
		*render.Plan, []domain.Befund, error)
}

// Texturrasterer zeichnet einen Plan.
//
// **Mit `*image.NRGBA` und nicht mit `image.Image`** — das ist die Gestalt,
// die [render.Rasterer] gibt, und eine engere Schnittstelle als der echte
// Erfüller wäre eine, die er nicht erfüllt. Eine Umhüllung daneben hätte
// dasselbe getan und eine Stelle mehr zum Lesen gekostet.
type Texturrasterer interface {
	Zeichne(ctx context.Context, plan *render.Plan, vergroesserung float64,
		guete render.Guete) (*image.NRGBA, []render.Befund, error)
}

// Textur ist ein gerastertes Kartenbild samt seiner Adresse.
type Textur struct {
	// Schluessel ist, woran zwei gleiche Texturen sich erkennen.
	//
	// **Er steht in der Antwort**, damit das Fenster teilen kann: zwei
	// Objekte mit demselben Schlüssel sind dieselbe Textur, und das ist die
	// Frage, die eine Referenzzählung beantwortet.
	Schluessel string

	Adresse string
	Guete   Texturguete

	Breite int
	Hoehe  int

	// Befunde sind die des Plans und des Rasterers — der Grund, warum eine
	// Karte behilflich aussieht.
	Befunde []string

	// Zweiseitig sagt, ob diese Karte eine zweite öffentliche Seite führt.
	//
	// # Warum es an der Textur steht und nicht in der Projektion
	//
	// **Weil es eine Auskunft über das Bild ist.** Wer ein Bild bekommt,
	// darf die Karte sehen — und dann darf er auch wissen, dass sie sich
	// umdrehen lässt. Die Projektion müsste dafür in die Karte sehen; sie
	// kennt nur den eingefrorenen Abzug.
	//
	// **Und es verrät nichts**: für ein verdecktes Objekt gibt es kein Bild
	// und damit keine Antwort — der Rücken kennt die Karte nicht.
	Zweiseitig bool
}

// Spieltexturen liefert Kartenbilder für den Spieltisch.
type Spieltexturen struct {
	plaene     Texturplanbauer
	rasterer   Texturrasterer
	ressourcen Ressourcenquelle
	ablage     Texturablage

	// verweise zählt, wie viele Karten auf ein Bild zeigen (E06, #2278).
	verweise texturverweise
}

// NeueSpieltexturen baut den Dienst.
func NeueSpieltexturen(
	plaene Texturplanbauer, rasterer Texturrasterer,
	ressourcen Ressourcenquelle, ablage Texturablage,
) *Spieltexturen {
	return &Spieltexturen{
		plaene: plaene, rasterer: rasterer,
		ressourcen: ressourcen, ablage: ablage,
	}
}

// Textur gibt das Bild einer Kartenseite in einer Güte — gezeichnet oder
// gefunden.
//
// **Eine öffentliche Seite dieser Karte.** Den Rücken gibt
// [Spieltexturen.Ruecken]; er gehört einem Spiel und nicht einer Karte.
func (s *Spieltexturen) Textur(
	ctx context.Context, karte *domain.CardRow,
	seite Kartenseite, guete Texturguete,
) (Textur, error) {
	if !IstKartenseite(seite) {
		return Textur{}, fmt.Errorf(
			"%w: %q ist keine Kartenseite — es gibt %s",
			domain.ErrValidation, seite, benannt(Kartenseiten()))
	}

	if fehler := pruefeGuete(guete); fehler != nil {
		return Textur{}, fehler
	}

	// **Eine zweite Seite gibt es nur, wenn die Karte sie führt.**
	//
	// **Abgewiesen und nicht behelfsweise die erste gezeigt**: wer die
	// zweite Seite eines einseitigen Waldes verlangt, hat einen Fehler im
	// Aufrufer — und ein stillschweigend ausgetauschtes Bild versteckte
	// ihn, bis jemand sich über eine Karte wundert, die sich nicht umdrehen
	// lässt.
	gesichter := 1
	if karte.HatZweiteSeite() {
		gesichter = 2
	}

	if seite == SeiteZweite && gesichter == 1 {
		return Textur{}, fmt.Errorf(
			"%w: diese Karte hat keine zweite Seite", domain.ErrValidation)
	}

	plan, planbefunde, fehler := s.plaene.Kartenplan(ctx, karte)
	if fehler != nil {
		return Textur{}, fehler
	}

	schluessel := Texturschluessel(plan, s.ressourcen, Texturwunsch{
		Seite: string(seite), Guete: guete, Gesichter: gesichter,
	})

	// **Das Bild gehört jetzt dieser Karte** — und ein voriges, auf das keine
	// mehr zeigt, geht (E06, #2278).
	if verwaist, da := s.verweise.binde(texturhalter{karte: karte.ID, seite: seite, guete: guete}, schluessel); da {
		s.ablage.Vergiss(verwaist)
	}

	// **Erst fragen, dann zeichnen.** Das ist der ganze Zweck: sieben gleiche
	// Karten am Tisch sind ein Rasterlauf.
	if adresse, breite, hoehe, da := s.ablage.Adresse(schluessel); da {
		return Textur{
			Schluessel: schluessel, Adresse: adresse, Guete: guete,
			Breite:     breite,
			Hoehe:      hoehe,
			Zweiseitig: gesichter > 1,
			// **Die Befunde des Plans stehen auch am Treffer.** Sie sagen,
			// *warum* eine Karte behilflich aussieht — und das gilt beim
			// zweiten Blick so wie beim ersten.
			Befunde: alsTexte(planbefunde),
		}, nil
	}

	textur, fehler := s.zeichne(ctx, plan, schluessel, guete, planbefunde,
		ausschnittswunsch{seite: seite, gesichter: gesichter})
	if fehler != nil {
		return Textur{}, fehler
	}

	textur.Zweiseitig = gesichter > 1

	return textur, nil
}

// Ruecken gibt den Rücken eines Spiels in einer Güte.
//
// # Warum ein eigener Weg und keine dritte Kartenseite
//
// **Weil er keine Karte hat und keine haben darf.** Ein verdecktes Objekt
// trägt keine Kartenkennung — die Projektion aus G1 hält sie zurück —, und
// wer den Rücken aus einer Karte baute, bände seine Maße und seinen
// Schlüssel an sie: zwei verdeckte Karten unterschieden sich dann in den
// Metadaten, ohne dass ein Feld sichtbar wäre (#1459).
//
// **Er gehört dem Spiel.** Zwei Spiele haben zwei Rücken; innerhalb eines
// Spiels haben alle Karten denselben. Das ist genau die Ersparnis, die ein
// Tisch braucht: vierzig verdeckte Karten sind **ein** Bild.
func (s *Spieltexturen) Ruecken(
	ctx context.Context, spiel string, guete Texturguete,
) (Textur, error) {
	if fehler := pruefeGuete(guete); fehler != nil {
		return Textur{}, fehler
	}

	plan := Rueckseitenplan(s.ressourcen.Rueckenbild(spiel))

	schluessel := Texturschluessel(&plan, s.ressourcen, Texturwunsch{
		Seite: MarkeRuecken, Guete: guete, Gesichter: 1,
	})

	if adresse, breite, hoehe, da := s.ablage.Adresse(schluessel); da {
		return Textur{
			Schluessel: schluessel, Adresse: adresse, Guete: guete,
			Breite: breite, Hoehe: hoehe, Befunde: []string{},
		}, nil
	}

	return s.zeichne(ctx, &plan, schluessel, guete, nil,
		ausschnittswunsch{seite: SeiteVorne, gesichter: 1})
}

// pruefeGuete weist ein Wort ab, das keine Stufe benennt.
func pruefeGuete(guete Texturguete) error {
	if !IstTexturguete(guete) {
		return fmt.Errorf(
			"%w: %q ist keine Texturgüte — es gibt %s",
			domain.ErrValidation, guete, benannt(Texturgueten()))
	}

	return nil
}

// zeichne rastert, legt ab und gibt die Textur.
func (s *Spieltexturen) zeichne(
	ctx context.Context, plan *render.Plan, schluessel string,
	guete Texturguete, planbefunde []domain.Befund, wunsch ausschnittswunsch,
) (Textur, error) {
	gezeichnet, rasterbefunde, fehler := s.rasterer.Zeichne(ctx, plan,
		vergroesserungFuer(plan, guete.Breite(), wunsch.gesichter),
		guete.rasterstufe())
	if fehler != nil {
		return Textur{}, fehler
	}

	bild := ausschnittAus(gezeichnet, wunsch)

	// **`BestSpeed` und nicht die Vorgabe** — derselbe Grund wie bei der
	// Vorschau (`app.schnellPacken`): das Bild bleibt Punkt für Punkt
	// dasselbe, und ein Sechstel der Zeit ginge sonst in Bytes, die niemand
	// aufhebt. Ein Spieltisch zeichnet sechzig davon.
	var gepackt bytes.Buffer

	kodierer := png.Encoder{CompressionLevel: png.BestSpeed}

	// **Ohne Fehlerweg, und das ist nachgelesen**: `png.Encode` gibt nur
	// weiter, was der Schreiber klagt, und ein `bytes.Buffer` klagt nicht
	// (`bytes.Buffer.Write`: „err is always nil"). Dieselbe Überlegung steht
	// an `app.bildnaehte`.
	_ = kodierer.Encode(&gepackt, bild) //nolint:errcheck // siehe darüber

	breite, hoehe := bild.Bounds().Dx(), bild.Bounds().Dy()

	return Textur{
		Schluessel: schluessel,
		Adresse: s.ablage.Hinterlege(
			schluessel, gepackt.Bytes(), breite, hoehe),
		Guete:   guete,
		Breite:  breite,
		Hoehe:   hoehe,
		Befunde: append(alsTexte(planbefunde), alsRastertexte(rasterbefunde)...),
	}, nil
}

// alsRastertexte macht aus Rasterbefunden Sätze.
//
// **Eine zweite Funktion und keine Vereinigung**: `render.Befund` und
// `domain.Befund` sind zwei Typen, und das ist Absicht — der Rasterer kennt
// `domain` nicht. Ein gemeinsamer Typ machte aus der Naht eine Verbindung
// (derselbe Satz steht an `render.Farbe`).
func alsRastertexte(befunde []render.Befund) []string {
	texte := make([]string, 0, len(befunde))
	for _, befund := range befunde {
		texte = append(texte, befund.String())
	}

	return texte
}

// alsTexte macht aus Befunden Sätze — leer und nicht `nil`.
func alsTexte(befunde []domain.Befund) []string {
	texte := make([]string, 0, len(befunde))
	for _, befund := range befunde {
		texte = append(texte, befund.String())
	}

	return texte
}
