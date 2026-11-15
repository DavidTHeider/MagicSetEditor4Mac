package render

import (
	"fmt"
	"image"
)

// mischeUeberMaske führt `masked_blend(light:, dark:, mask:)` aus.
//
// # Die Regel, nachgelesen in MSE2
//
// `src/gfx/blend_image.cpp:63` (Fassung f9d9356 vom 2021-07-06):
//
//	data1[i] = (data1[i] * dataM[i] + data2[i] * (255 - dataM[i])) / 255;
//
// `data1` ist `light`, `data2` ist `dark`, `dataM` die Maske. Wo die Maske
// **weiß** ist, gilt `light`; wo sie **schwarz** ist, `dark`.
//
// **Mit 2.734 Vorkommen die häufigste Verknüpfung des Packs**
// (#270). An
// ihr hängt der Hybridrahmen (#306).
//
// # Je Farbkanal, nicht über die Helligkeit
//
// **Der erste Anlauf nahm die mittlere Helligkeit der Maske** und begründete
// das damit, eine Maske sei kein Foto. Die Begründung war erfunden: MSE läuft
// über **Teilbildpunkte** — `dataM[i]` ist bei rot der rote Maskenkanal, bei
// grün der grüne. Bei einer grauen Maske ist beides gleich, bei einer farbigen
// nicht.
//
// Das ist der Unterschied zwischen erschlossen und nachgelesen, und er ist
// erst aufgefallen, als der Quelltext von MSE2 tatsächlich dalag.
//
// # Ohne Alpha
//
// `size = img1.GetWidth() * img1.GetHeight() * 3` — MSE mischt **drei** Kanäle.
// Das Alpha des hellen Bildes bleibt, wie es ist.
//
// # Gleiche Größe, sonst ein Befund
//
// MSE wirft `Images used for blending must have the same size`. Hier ist es ein
// Befund statt eines Abbruchs (§8): eine Karte mit einer unpassenden Maske soll
// ihre anderen sechzig Ebenen behalten.
func (r *Rasterer) mischeUeberMaske(baum Bildbaum, tiefe int) (gemischt image.Image, befund string) {
	hell, dunkel, maske, fehlt := r.dreiBilder(baum, tiefe)
	if fehlt != "" {
		return nil, fehlt
	}

	if abweichend := gleicheGroesse(hell, dunkel, maske); abweichend != "" {
		return nil, "masked_blend: " + abweichend
	}

	rahmen := hell.Bounds()
	ziel := image.NewNRGBA(image.Rect(0, 0, rahmen.Dx(), rahmen.Dy()))

	for y := range rahmen.Dy() {
		for x := range rahmen.Dx() {
			a := geradeFarbe(hell, hell.Bounds().Min.X+x, hell.Bounds().Min.Y+y)
			b := geradeFarbe(dunkel, dunkel.Bounds().Min.X+x, dunkel.Bounds().Min.Y+y)
			m := geradeFarbe(maske, maske.Bounds().Min.X+x, maske.Bounds().Min.Y+y)

			stelle := ziel.PixOffset(x, y)
			ziel.Pix[stelle] = mischeKanal(a.R, b.R, m.R)
			ziel.Pix[stelle+1] = mischeKanal(a.G, b.G, m.G)
			ziel.Pix[stelle+2] = mischeKanal(a.B, b.B, m.B)

			// **Das Alpha des hellen Bildes bleibt.** MSE mischt drei Kanäle.
			ziel.Pix[stelle+3] = a.A
		}
	}

	return ziel, ""
}

// mischeKanal ist die Zeile aus `blend_image.cpp:73`, in Go.
//
//	data1[i] = (data1[i] * dataM[i] + data2[i] * (255 - dataM[i])) / 255
//
// **Ganzzahlig und abschneidend, wie C++.** `alsAnteil` rundet, und das ist
// hier falsch: bei `light=0, dark=1, maske=127` gibt MSE **0** und eine
// Rundung **1**. Angestrichen vom Nachlauf des Zweitblicks — der Wächter ließ
// eine Stufe Abweichung zu und konnte den Unterschied deshalb nicht sehen.
//
// Der Wertebereich ist gezeigt: beide Summanden zusammen sind höchstens
// `255 * 255`, geteilt durch 255 also höchstens 255.
func mischeKanal(hell, dunkel, maske uint8) uint8 {
	// **Der Wertebereich steht in einem `min`, nicht in einem Kommentar.**
	//
	// Zwei Produkte aus je höchstens 255 ergeben höchstens 65.025, geteilt
	// durch 255 also höchstens 255 — die Umwandlung kann nicht überlaufen.
	// `min` sagt das dem Übersetzer und dem Leser; ein Kommentar sagte es nur
	// dem Leser, und `gosec` müsste ihn glauben.
	gemischt := (int(hell)*int(maske) + int(dunkel)*(255-int(maske))) / 255

	return uint8(min(max(gemischt, 0), 255)) //nolint:gosec // durch min und max begrenzt
}

// gleicheGroesse sagt, welches Bild aus der Reihe fällt — oder nichts.
func gleicheGroesse(bilder ...image.Image) string {
	erste := bilder[0].Bounds()

	for _, bild := range bilder[1:] {
		if bild.Bounds().Dx() != erste.Dx() || bild.Bounds().Dy() != erste.Dy() {
			return fmt.Sprintf("die Bilder sind verschieden groß (%dx%d gegen %dx%d)",
				erste.Dx(), erste.Dy(), bild.Bounds().Dx(), bild.Bounds().Dy())
		}
	}

	return ""
}

// dreiBilder holt `light`, `dark` und `mask` — oder sagt, welches fehlt.
func (r *Rasterer) dreiBilder(baum Bildbaum, tiefe int) (
	hell, dunkel, maske image.Image, fehlt string,
) {
	gefunden, fehlt := r.kinderbilder(baum, tiefe, "light", "dark", "mask")
	if fehlt != "" {
		return nil, nil, nil, fehlt
	}

	return gefunden[0], gefunden[1], gefunden[2], ""
}

// kinderbilder führt benannte Kinder aus und **behält ihre Befunde**.
//
// # Warum das Sammeln nötig ist
//
// [Rasterer.ausBaum] gibt für eine nicht gebaute Verknüpfung **Bild und
// Befund** zurück — das Ersatzbild wird gezeichnet, und die Warnung soll
// mitreisen. Der erste Anlauf übernahm den Befund nur, wenn das Bild `nil`
// war: ein `recolor_image` **innerhalb** eines `masked_blend` zeichnete damit
// sein Ersatzbild, und die Warnung verschwand.
//
// **Halbe Unterstützung sah dann aus wie vollständige, und der Befundzähler
// sank** — angestrichen vom Zweitblick über die Phasen B bis D. Ein Erfolgswert
// mit Warnung ist nicht dasselbe wie ein Wert mit Fehler, und wer beides gleich
// behandelt, verliert die Warnung.
func (r *Rasterer) kinderbilder(baum Bildbaum, tiefe int, namen ...string) (
	bilder []image.Image, fehlt string,
) {
	bilder = make([]image.Image, 0, len(namen))

	for _, name := range namen {
		kind, da := baum.Kind(name)
		if !da {
			return nil, baum.Knotenart + " fehlt " + name + ":"
		}

		// **Was das Kind beiseitelegt, bekommt seinen Namen davor.**
		// Seit [Rasterer.ausBaumTief] einen Befund neben einem Bild selbst
		// aufhebt (#473),
		// kommt er hier nicht mehr als Rückgabewert an — die Auskunft
		// „welches Kind" ginge sonst verloren, und die ist bei drei Kindern
		// die halbe Meldung.
		vorher := len(r.baumbefunde)

		bild, befund := r.ausBaumTief(kind, tiefe+1)
		if bild == nil {
			return nil, befund
		}

		for stelle := vorher; stelle < len(r.baumbefunde); stelle++ {
			r.baumbefunde[stelle] = name + ": " + r.baumbefunde[stelle]
		}

		bilder = append(bilder, bild)
	}

	return bilder, ""
}
