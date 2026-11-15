package app

// **Den Ausschnitt selbst wählen** —
// #576.
//
// MSE öffnet beim Einsetzen ein Fenster: links das Ursprungsbild mit einem
// verschiebbaren Rahmen, rechts die Vorschau in Feldgröße
// (`ImageSliceWindow`, `src/gui/image_slice_window.cpp`, Zeilen 68–360,
// Fassung f9d9356 vom 2021-07-06).
//
// **Damit zerfällt das Einsetzen in zwei Schritte** — vorbereiten und ablegen.
// Dazwischen liegt das Fenster, und es gehört nach `frontend`.

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// hoechsteVorschau ist die Kantenlänge, auf die die Quelle für die Anzeige
// gerechnet wird.
//
// **Nicht das ganze Bild über die Bindung.** Eine Digitalkamera liefert
// 4000 × 3000; als PNG mit Base64 sind das über zwanzig Megabyte durch die
// Wails-Brücke, für ein Fenster, das sie ohnehin verkleinert anzeigt. MSE
// rechnet dafür ebenso herunter (`ImageSliceSelector::createBitmap`, Zeilen
// 610–640).
//
// **Gerechnet wird trotzdem in den Punkten der Quelle**: die Auswahl, die
// zurückkommt, meint das Original. Ein Fenster, das in seinen eigenen Punkten
// rechnete, gäbe bei jedem Maßstab ein anderes Bild.
const hoechsteVorschau = 1000

// AuswahlDTO ist ein Ausschnitt in den Punkten der Quelle.
type AuswahlDTO struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Breite int `json:"breite"`
	Hoehe  int `json:"hoehe"`
}

// ZuschnittAntwort ist, was das Zuschnittfenster zum Aufmachen braucht.
//
// **Ein leerer Pfad heißt „abgebrochen"** — dieselbe Übereinkunft wie bei
// [App.SetOeffnen]: „nichts gewählt" ist eine Aussage und kein Fehler.
type ZuschnittAntwort struct {
	// Pfad ist die gewählte Datei. Der zweite Schritt braucht ihn, und nur der
	// Dialog kennt ihn.
	Pfad string `json:"pfad"`

	// Quelle ist das Ursprungsbild als PNG-Data-URI, für die Anzeige
	// verkleinert.
	Quelle string `json:"quelle"`

	QuelleBreite int `json:"quelleBreite"`
	QuelleHoehe  int `json:"quelleHoehe"`

	ZielBreite int `json:"zielBreite"`
	ZielHoehe  int `json:"zielHoehe"`

	// Auswahl ist MSEs Vorgabewahl — was das Fenster beim Öffnen zeigt.
	Auswahl AuswahlDTO `json:"auswahl"`

	Fehler *Fehler `json:"fehler,omitempty"`
}

// BildSetzen fragt nach einer Bilddatei und bereitet den Zuschnitt vor.
//
// **Es schreibt nicht mehr.** Bis #576 tat es beides; jetzt liegt das Fenster
// dazwischen, und abgelegt wird über [App.BildEinsetzen]. Das ist eine
// benannte Änderung an einer bestehenden Bindung (§15, Satz 4) — ohne sie
// gäbe es keine Stelle, an der jemand den Ausschnitt sähe.
func (a *App) BildSetzen(ctx context.Context, karteID, feld string) ZuschnittAntwort {
	ctx = a.imProfil(ctx)

	_, _, breite, hoehe, befund := a.bildziel(ctx, karteID, feld)
	if befund != nil {
		return ZuschnittAntwort{Fehler: befund}
	}

	pfad, err := a.waehleDatei(ctx, Dateiauswahl{
		Titel:  DialogtexteFuer(a.sprache).BildOeffnen,
		Filter: DialogtexteFuer(a.sprache).BildFilter,
		Muster: "*.png;*.jpg;*.jpeg;*.gif",
	})
	if err != nil {
		return ZuschnittAntwort{Fehler: Uebersetze(ctx, err)}
	}

	if pfad == "" {
		return ZuschnittAntwort{}
	}

	return a.vorbereiten(ctx, pfad, breite, hoehe)
}

// BildAusDatei bereitet den Zuschnitt einer Datei vor, deren Pfad feststeht.
//
// **Für das Ziehen und Ablegen** (#246): wer eine Datei in das Fenster zieht,
// hat sie schon gewählt, und Wails gibt dabei den absoluten Pfad. Ein Dialog
// danach wäre eine Rückfrage auf eine Antwort.
func (a *App) BildAusDatei(ctx context.Context, karteID, feld, pfad string) ZuschnittAntwort {
	ctx = a.imProfil(ctx)

	_, _, breite, hoehe, befund := a.bildziel(ctx, karteID, feld)
	if befund != nil {
		return ZuschnittAntwort{Fehler: befund}
	}

	if pfad == "" {
		return ZuschnittAntwort{Fehler: Uebersetze(ctx, fmt.Errorf(
			"%w: es wurde keine Bilddatei genannt", domain.ErrValidation))}
	}

	return a.vorbereiten(ctx, pfad, breite, hoehe)
}

// BildEinsetzen schneidet auf die gewählte Fläche zu und legt das Bild ab.
func (a *App) BildEinsetzen(
	ctx context.Context, setID, karteID, feld, pfad string, auswahl AuswahlDTO,
) SchreibAntwort {
	ctx = a.imProfil(ctx)

	return a.bildAblegen(ctx, setID, karteID, feld, pfad, auswahl, "")
}

// bildAblegen schneidet zu, legt das Bild ab und schreibt die Felder.
//
// # Ein Zug und nicht zwei
//
// **`weitere` geht im selben Zug hinaus** (#1163, Befund Z7-02). Der
// Scryfall-Weg reicht hier seinen Künstler herein; zwischen zwei Aufrufen läge
// ein Stand, in dem die Karte fremde Kunst ohne ihre Nennung trägt.
// **Ohne Kennungen zurück** (#1163): sie gingen bis Phase 3.2 an den
// Aufrufer, weil `ScryfallBildEinsetzen` danach ein **zweites** Mal schrieb.
// Das tut es nicht mehr — und ein Rückgabewert, den niemand liest, ist eine
// offene Tür für genau diesen zweiten Schreibvorgang.
func (a *App) bildAblegen(
	ctx context.Context, setID, karteID, feld, pfad string, auswahl AuswahlDTO,
	herkunft string, weitere ...domain.Value,
) SchreibAntwort {
	setkennung, befund := benannteKennungAus("setID", setID)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	kennung, karte, breite, hoehe, befund := a.bildziel(ctx, karteID, feld)
	if befund != nil {
		return SchreibAntwort{Fehler: befund}
	}

	bild, err := bilddateiLesen(pfad)
	if err != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, err)}
	}

	// **Begrenzt wird auch hier, und das ist keine zweite Regel.** Die Auswahl
	// kommt aus dem Fenster, und ein Fenster ist eine Eingabe (§8): eine
	// Fläche außerhalb der Quelle ergäbe sonst schwarze Ränder statt einer
	// Absage. Es ist derselbe Aufruf, den das Fenster bei jeder Bewegung tut —
	// bei unveränderter Auswahl ändert er nichts.
	flaeche := render.Zuschnittwahl{
		Quelle:  bild.Bounds(),
		Ziel:    image.Pt(breite, hoehe),
		Auswahl: alsRechteck(auswahl),
	}.Begrenze(render.BevorzugtKeines)

	name, err := a.bildablage.Ablegen(karte.SetID, feld,
		render.ZuschnittAus(bild, flaeche, breite, hoehe))
	if err != nil {
		return SchreibAntwort{Fehler: Uebersetze(ctx, err)}
	}

	// **Das Set kommt vom Aufrufer, nicht aus `karte.SetID`** (#808). Die
	// Karte gegen ihr eigenes Set zu halten ginge immer auf — ein Wächter,
	// der nicht scheitern kann, ist keiner (§10).
	// **Auch das Bildfeld trägt seine Herkunft** (#888). Wer im Verlauf sieht,
	// dass hier ein Bildname steht, soll erkennen können, ob er von der Platte
	// kam oder von Scryfall — beides sieht danach gleich aus.
	werte := append([]domain.Value{{Field: feld, Text: name}}, weitere...)

	_, _, err = a.schreiber.FelderVonAussen(ctx, setkennung, kennung, werte, herkunft)
	if err != nil {
		// **Das Bild liegt schon und hat noch keinen Verweis** (#1163, Z7-02).
		// Bliebe es liegen, wüchse das Bilderverzeichnis bei jedem Fehlversuch
		// um eine Datei, die niemand je wieder findet — und der Stand wäre
		// weder der neue noch der alte.
		//
		// **Der Name ist frisch** (`service.freierName`), also nimmt dieses
		// Wegräumen niemandem etwas weg.
		return SchreibAntwort{Fehler: Uebersetze(ctx, errors.Join(
			err, a.bildablage.Wegnehmen(karte.SetID, name)))}
	}

	return SchreibAntwort{}
}

// SymbolZuschnittgrenze gibt die Auswahl, die MSEs Regeln aus einer gezogenen macht.
//
// **Sie liegt in Go, weil die Regel portiert ist.** Eine zweite Fassung in
// TypeScript wäre eine zweite Antwort auf dieselbe Frage — und sie liefe
// auseinander, sobald jemand eine davon nachzieht (§15).
//
// **Bevorzugt sagt, welche Kante gezogen wurde**: `breite`, `hoehe` oder nichts.
// Ein unbekanntes Wort heißt „keines" — die Bindung nimmt eine Zeichenkette
// entgegen, und ein Fehler dafür wäre ein Fehlerweg ohne Nutzen.
func SymbolZuschnittgrenze(
	quelleBreite, quelleHoehe, zielBreite, zielHoehe int,
	auswahl AuswahlDTO, frei bool, bevorzugt string,
) AuswahlDTO {
	flaeche := render.Zuschnittwahl{
		Quelle:  image.Rect(0, 0, quelleBreite, quelleHoehe),
		Ziel:    image.Pt(zielBreite, zielHoehe),
		Auswahl: alsRechteck(auswahl),
		Frei:    frei,
	}.Begrenze(kanteAus(bevorzugt))

	return alsDTO(flaeche)
}

// Zuschnittziehen gibt die Auswahl, nachdem der Rahmen gezogen wurde.
//
// **Die Strecke kommt in Punkten der Quelle**, nicht des Bildschirms: MSE
// teilt sie durch den Maßstab seines Fensters (`ImageSliceSelector::onMotion`,
// Zeile 611, Fassung f9d9356 vom 2021-07-06), und hier tut das Fenster es —
// nur es kennt seine Größe.
//
// **kanteX und kanteY sind −1, 0 oder 1** — MSEs `dragX`/`dragY`. Beide null
// heißt: den ganzen Rahmen verschieben.
func (a *App) Zuschnittziehen(
	quelleBreite, quelleHoehe, zielBreite, zielHoehe int,
	auswahl AuswahlDTO, frei bool, x, y float64, kanteX, kanteY int,
) AuswahlDTO {
	return alsDTO(render.Zuschnittwahl{
		Quelle:  image.Rect(0, 0, quelleBreite, quelleHoehe),
		Ziel:    image.Pt(zielBreite, zielHoehe),
		Auswahl: alsRechteck(auswahl),
		Frei:    frei,
	}.Ziehe(x, y, kanteX, kanteY))
}

// alsDTO macht aus einem Rechteck ein DTO.
func alsDTO(flaeche image.Rectangle) AuswahlDTO {
	return AuswahlDTO{
		X: flaeche.Min.X, Y: flaeche.Min.Y,
		Breite: flaeche.Dx(), Hoehe: flaeche.Dy(),
	}
}

// kanteAus deutet die gezogene Kante.
func kanteAus(bevorzugt string) render.Bevorzugt {
	switch bevorzugt {
	case "breite":
		return render.BevorzugtBreite
	case "hoehe":
		return render.BevorzugtHoehe
	default:
		return render.BevorzugtKeines
	}
}

// alsRechteck macht aus einem DTO ein Rechteck.
func alsRechteck(auswahl AuswahlDTO) image.Rectangle {
	return image.Rect(auswahl.X, auswahl.Y,
		auswahl.X+auswahl.Breite, auswahl.Y+auswahl.Hoehe)
}

// vorbereiten liest die Datei und stellt zusammen, was das Fenster zeigt.
func (a *App) vorbereiten(ctx context.Context, pfad string, breite, hoehe int) ZuschnittAntwort {
	bild, err := bilddateiLesen(pfad)
	if err != nil {
		return ZuschnittAntwort{Fehler: Uebersetze(ctx, err)}
	}

	return a.vorbereitenAus(ctx, bild, pfad, breite, hoehe)
}

// vorbereitenAus stellt dasselbe aus einem schon gelesenen Bild zusammen.
//
// # Warum es die Trennung gibt
//
// **Weil ein erzeugtes Bild keine Datei ist** (A12, #1451). Es liegt als
// Bytes im Vorrat, und es dafür in eine Nachbardatei zu schreiben hieße:
// eine Datei anlegen, die niemand braucht, sie füllen, sie wieder wegräumen
// — drei Fehlerwege für einen Umweg.
//
// **`pfad` ist dann keine Datei, sondern die Kennung des Vorschlags.** Das
// Fenster reicht ihn unverändert zurück; welche Bindung ihn einsetzt,
// entscheidet es an anderer Stelle ([useBildfeld]).
func (a *App) vorbereitenAus(
	ctx context.Context, bild image.Image, pfad string, breite, hoehe int,
) ZuschnittAntwort {
	angezeigt, err := a.alsDatenURI(verkleinert(bild))
	if err != nil {
		return ZuschnittAntwort{Fehler: Uebersetze(ctx, err)}
	}

	vorgabe := render.Zuschnittflaeche(bild.Bounds(), breite, hoehe)

	return ZuschnittAntwort{
		Pfad:         pfad,
		Quelle:       angezeigt,
		QuelleBreite: bild.Bounds().Dx(),
		QuelleHoehe:  bild.Bounds().Dy(),
		ZielBreite:   breite,
		ZielHoehe:    hoehe,
		Auswahl: AuswahlDTO{
			X:      vorgabe.Min.X - bild.Bounds().Min.X,
			Y:      vorgabe.Min.Y - bild.Bounds().Min.Y,
			Breite: vorgabe.Dx(),
			Hoehe:  vorgabe.Dy(),
		},
	}
}

// verkleinert rechnet die Quelle auf Anzeigegröße herunter.
//
// **Über denselben Filter wie alles andere** ([render.ZuschnittAus]): ein
// zweiter daneben zeigte im Fenster ein anderes Bild als auf der Karte.
func verkleinert(bild image.Image) image.Image {
	breite, hoehe := bild.Bounds().Dx(), bild.Bounds().Dy()
	if breite <= hoechsteVorschau && hoehe <= hoechsteVorschau {
		return bild
	}

	if breite >= hoehe {
		hoehe = max(hoehe*hoechsteVorschau/breite, 1)
		breite = hoechsteVorschau
	} else {
		breite = max(breite*hoechsteVorschau/hoehe, 1)
		hoehe = hoechsteVorschau
	}

	return render.ZuschnittAus(bild, bild.Bounds(), breite, hoehe)
}

// alsDatenURI packt ein Bild als PNG in einen Data-URI.
//
// **Über die Naht [App.packen] und nicht über `png.Encode`.** Mit einem
// `bytes.Buffer` kann die Kodierung nicht scheitern; der Fehlerzweig wäre
// ohne die Naht eine Behauptung darüber, was im Ernstfall geschieht.
func (a *App) alsDatenURI(bild image.Image) (string, error) {
	var gepackt bytes.Buffer
	if err := a.packen(&gepackt, bild); err != nil {
		return "", fmt.Errorf("das Bild ließ sich nicht anzeigen: %w", err)
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(gepackt.Bytes()), nil
}

// Zuschnittgrenze hält eine Auswahl in ihren Grenzen.
//
// **Eine Methode über einer Funktion**, seit dem 2026-09-02
// (#1391):
// die Rechnung braucht kein `App`, und die Flächen ohne Fenster haben keins.
// Der Rumpf steht deshalb in [SymbolZuschnittgrenze]; hier bleibt der Weg, den die
// Wails-Bindung nimmt.
func (a *App) Zuschnittgrenze(quelleBreite, quelleHoehe, zielBreite, zielHoehe int,
	auswahl AuswahlDTO, frei bool, bevorzugt string,
) AuswahlDTO {
	return SymbolZuschnittgrenze(quelleBreite, quelleHoehe, zielBreite, zielHoehe, auswahl, frei, bevorzugt)
}
