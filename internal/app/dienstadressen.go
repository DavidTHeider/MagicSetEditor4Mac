// **Die Adressen der drei neuen Dienste** — Werkstück 8 von Phase 4.8.
//
// **Umlenkbar wie die von Scryfall**, und aus demselben Grund: ein Lauf, der
// gegen den echten Dienst ginge, ist keiner (§3). Die Umgebungsnamen stehen
// hier, damit sie an **einer** Stelle stehen.

package app

import (
	"os"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgapi"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtggraphql"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scrydex"
)

// Die Umgebungsnamen, mit denen sich die Adressen umlenken lassen.
const (
	MTGAPIumgebung     = "MSE_MTGAPI_ADRESSE"
	MTGGraphQLumgebung = "MSE_MTGGRAPHQL_ADRESSE"
	Scrydexumgebung    = "MSE_SCRYDEX_ADRESSE"
)

var (
	mtgapiadresse     = umgelenkt(MTGAPIumgebung, mtgapi.Adresse)
	mtggraphqladresse = umgelenkt(MTGGraphQLumgebung, mtggraphql.Adresse)
	scrydexadresse    = umgelenkt(Scrydexumgebung, scrydex.Adresse)
)

// AdresseAus liest eine Umgebungsangabe — oder gibt die echte Adresse.
//
// **Öffentlich für den Lauf, der belegt, dass sie greift.** Ohne ihn bliebe
// der Zweig unbetreten, und ein Testlauf ginge irgendwann an den echten
// Dienst, ohne dass es jemandem auffiele (§3).
func AdresseAus(umgebung, echt string) string { return umgelenkt(umgebung, echt) }

// umgelenkt liest eine Umgebungsangabe — oder gibt die echte Adresse.
func umgelenkt(umgebung, echt string) string {
	if ort := os.Getenv(umgebung); ort != "" {
		return ort
	}

	return echt
}

// MitDienstadressen lenkt alle drei um — **nur für Tests**.
//
// Sie gibt zurück, was vorher dastand, damit ein Aufrufer sie zurückstellen
// kann.
func MitDienstadressen(mtgapiOrt, graphqlOrt, scrydexOrt string) func() {
	vorher := [3]string{mtgapiadresse, mtggraphqladresse, scrydexadresse}

	mtgapiadresse, mtggraphqladresse, scrydexadresse =
		mtgapiOrt, graphqlOrt, scrydexOrt

	return func() {
		mtgapiadresse, mtggraphqladresse, scrydexadresse =
			vorher[0], vorher[1], vorher[2]
	}
}
