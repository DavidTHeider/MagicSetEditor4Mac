package service

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
)

// setbildmarke kennzeichnet einen Verweis ins Set-Paket, den dieses Programm
// gebaut hat ([mitBildwerten]).
const setbildmarke = "setpaket:eigen"

// erfundenerSetort sagt, ob irgendwo im Baum ein Verweis ins Set-Paket steht,
// den **nicht** dieses Programm gebaut hat — QA-06c
// (#1976),
// Fläche 9.
//
// `setpaket:<ort>!<datei>` ist eine Form dieses Programms, gebaut aus dem
// Ort, den der Bestand kennt. Schreibt eine Vorlage sie selbst — wörtlich oder
// als Ergebnis eines Skripts —, nennt sie einen beliebigen Ort, und die
// Bildquelle öffnete ihn als Paket. MSE liest ein Kartenbild nur aus dem
// eigenen Paket (`getLocalPackage().openIn`, `src/render/value/image.cpp`,
// Zeile 35, Fassung f9d9356 vom 2021-07-06).
//
// **Nach der Herkunft und nicht nach der Gestalt.** Ein Skript sieht echte
// Verweise: [mitBildwerten] legt jedes Bildfeld der Karte als einen in die
// Umgebung, und `image: { card.image }` gibt ihn zurück. Deshalb trägt der
// eigene die Marke ([msescript.Wert.Marke]), die ein Skript weder setzen noch
// fälschen kann; was ohne sie dasteht, hat jemand geschrieben.
func erfundenerSetort(baum msescript.Bildbaum) bool {
	if _, _, ist := msepaket.ZerlegeSetpaketverweis(baum.Pfad); ist && baum.Marke != setbildmarke {
		return true
	}

	for _, kind := range baum.Kinder {
		if erfundenerSetort(kind.Baum) {
			return true
		}
	}

	return false
}
