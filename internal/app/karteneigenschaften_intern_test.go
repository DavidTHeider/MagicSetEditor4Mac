package app

// **Die Eigenschaften kommen durch die Naht** — Werkstück M2 der
// Regelmaschine
// (#1604).
//
// **Über den ganzen Weg und nicht über die Abbildung allein**: geprüft wird,
// dass eine Karte aus dem **Bestand** mit ihren Feldwerten in einer Partie
// ankommt und dort gedeutet wird. Eine Vorrichtung, die den Verbund
// zusammensetzte, prüfte die Vorrichtung (§3).

import (
	"database/sql"
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// einBaerImBestand legt ein Set mit **einer** Magic-Karte an und gibt ein
// Deck darauf.
//
// **Zehn Bären und nicht vier** (#1851): eine Starthand zieht sieben
// (CR 103.5), und aus vier wäre der Rest ein Griff ins Leere (CR 104.3c).
func einBaerImBestand(t *testing.T, db *sql.DB) string {
	t.Helper()

	ctx := riegelkontext(t)

	setID, fehler := storage.NeuerSetSpeicher(db).SetAnlegen(ctx, &domain.Set{
		Name: "Bärenset", Game: "magic", Style: "m15",
		Cards: []*domain.Card{{Values: []domain.Value{
			{Field: "name", Text: "Grizzlybär"},
			{Field: "casting cost", Text: "{1}{G}"},
			// **Legendär, damit auch ein Supertyp durch die Naht geht** —
			// CR 205.4a. Ohne ihn bliebe der Weg dafür unbetreten (§3).
			{Field: "super type", Text: "Legendary Creature"},
			{Field: "sub type", Text: "Bear"},
			{Field: "power", Text: "2"},
			{Field: "toughness", Text: "2"},
		}}},
	}, nil)
	if fehler != nil {
		t.Fatalf("Set anlegen: %v", fehler)
	}

	karten, lesefehler := storage.NeuerSetLeser(db).Karten(ctx, setID)
	if lesefehler != nil || len(karten) == 0 {
		t.Fatalf("Karten: %v (%d)", lesefehler, len(karten))
	}

	ablage := storage.NeueDeckablage(db)

	deck, deckfehler := ablage.Anlegen(ctx, domain.Deck{
		Name: "Bären", Spiel: "magic",
	})
	if deckfehler != nil {
		t.Fatalf("Deck anlegen: %v", deckfehler)
	}

	if _, platzfehler := ablage.PlatzSetzen(ctx, deck.ID, domain.Deckplatz{
		KarteID: karten[0].ID, Menge: 10, Bereich: domain.BereichHauptdeck,
	}); platzfehler != nil {
		t.Fatalf("Platz setzen: %v", platzfehler)
	}

	return strconv.FormatInt(deck.ID, 10)
}

// **Eine echte Karte kommt mit Farbe, Typ und Werten am Tisch an.**
func TestDieEigenschaftenKommenDurchDieNaht(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)
	anwendung := MitPartienleitung(NeueApp(""), db)
	anwendung.db = db
	anwendung.profil.Store(1)

	gebunden := NeueBindungen(anwendung)
	deckID := einBaerImBestand(t, db)

	angelegt := gebunden.PartieAnlegen(t.Context(), einEntwurf(deckID))
	if angelegt.Fehler != nil {
		t.Fatalf("PartieAnlegen: %+v", angelegt.Fehler)
	}

	beginnAbschliessen(t, ueberDieBindungen(t.Context(), gebunden), angelegt.Partie)

	// **Auf den Tisch, damit hineingesehen werden darf.** Eine Bibliothek ist
	// auch für ihren Eigner verdeckt (G1); ein Lauf über sie prüfte die
	// Geheimhaltung und nicht die Eigenschaften.
	erstes := inDerBibliothek(t, gebunden.Partie(t.Context(), angelegt.Partie.ID, "sitz-1").
		Partie.Objekte, "sitz-1")

	gestellt := gebunden.PartieBefehl(t.Context(), angelegt.Partie.ID, "b1", "sitz-1",
		"objekt.schieben", map[string]any{"objekt": erstes, "zone": "tisch"}, "")
	if gestellt.Fehler != nil {
		t.Fatalf("PartieBefehl: %+v", gestellt.Fehler)
	}

	gesehen := gebunden.Partie(t.Context(), angelegt.Partie.ID, "sitz-1")
	if gesehen.Fehler != nil {
		t.Fatalf("Partie: %+v", gesehen.Fehler)
	}

	// **Die Eigenschaften der geschobenen Karte** — seit #1851 liegen auch
	// sieben auf der Hand, und die zeigt ihr Eigner ebenfalls.
	eigen := eigenschaftenVon(t, gesehen.Partie.Objekte, erstes)

	switch {
	case len(eigen.Farben) != 1 || eigen.Farben[0] != "G":
		t.Errorf("Farben = %v", eigen.Farben)
	case eigen.Manawert != 2:
		t.Errorf("Manawert = %d", eigen.Manawert)
	case len(eigen.Typen) != 1 || eigen.Typen[0] != "creature":
		t.Errorf("Typen = %v", eigen.Typen)
	case len(eigen.Supertypen) != 1 || eigen.Supertypen[0] != "legendary":
		t.Errorf("Supertypen = %v", eigen.Supertypen)
	case len(eigen.Untertypen) != 1 || eigen.Untertypen[0] != "Bear":
		t.Errorf("Untertypen = %v", eigen.Untertypen)
	case !eigen.Staerke.Steht || eigen.Staerke.Zahl != 2:
		t.Errorf("Stärke = %+v", eigen.Staerke)
	}
}

// **Und ein geheimes Objekt bringt keine mit.**
//
// Die Gegenprobe: dieselbe Partie aus der Sicht des anderen Sitzes — dessen
// Bibliothek ist für ihn verdeckt, und was darin liegt, geht ihn nichts an.
func TestEinGeheimesObjektBringtKeineEigenschaftenDurchDieNaht(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)
	anwendung := MitPartienleitung(NeueApp(""), db)
	anwendung.db = db
	anwendung.profil.Store(1)

	gebunden := NeueBindungen(anwendung)
	deckID := einBaerImBestand(t, db)

	angelegt := gebunden.PartieAnlegen(t.Context(), einEntwurf(deckID))
	if angelegt.Fehler != nil {
		t.Fatalf("PartieAnlegen: %+v", angelegt.Fehler)
	}

	gesehen := gebunden.Partie(t.Context(), angelegt.Partie.ID, "sitz-1")
	if gesehen.Fehler != nil {
		t.Fatalf("Partie: %+v", gesehen.Fehler)
	}

	for _, objekt := range gesehen.Partie.Objekte {
		if objekt.Geheim && objekt.Eigenschaften != nil {
			t.Fatalf("%q ist geheim und trägt %+v",
				objekt.Kennung, objekt.Eigenschaften)
		}
	}
}

// eigenschaftenVon gibt die Eigenschaften dieses Objekts.
func eigenschaftenVon(
	t *testing.T, objekte []ObjektDTO, kennung string,
) EigenschaftenDTO {
	t.Helper()

	for _, objekt := range objekte {
		if objekt.Kennung == kennung && objekt.Eigenschaften != nil {
			return *objekt.Eigenschaften
		}
	}

	t.Fatalf("%q trägt keine Eigenschaften", kennung)

	return EigenschaftenDTO{}
}

// **Die drei Kurzformen für die Textflächen** — MCP, REST und die
// Kommandozeile lesen sie, statt einen Verbund zusammenzusetzen.
func TestDieKurzformenGebenDieKarteAlsText(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		eigen  *EigenschaftenDTO
		farben string
		typ    string
		werte  string
	}{
		"nichts ist nichts": {nil, "", "", ""},

		"eine Kreatur": {
			eigen: &EigenschaftenDTO{
				Farben: []string{"W", "U"}, Supertypen: []string{"legendary"},
				Typen: []string{"creature"}, Untertypen: []string{"Human", "Wizard"},
				Staerke:          KartenwertDTO{Steht: true, Zahl: 2},
				Widerstandskraft: KartenwertDTO{Steht: true, Zahl: 3},
			},
			farben: "WU", typ: "legendary creature — Human Wizard", werte: "2/3",
		},

		"ein Land ohne Werte": {
			eigen: &EigenschaftenDTO{
				Supertypen: []string{"basic"}, Typen: []string{"land"},
				Untertypen: []string{"Mountain"},
			},
			farben: "", typ: "basic land — Mountain", werte: "",
		},

		"ein Stern bleibt ein Stern": {
			eigen: &EigenschaftenDTO{
				Typen:            []string{"creature"},
				Staerke:          KartenwertDTO{Steht: true, Stern: true},
				Widerstandskraft: KartenwertDTO{Steht: true, Stern: true},
			},
			farben: "", typ: "creature", werte: "*/*",
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			switch {
			case fall.eigen.Farbtext() != fall.farben:
				t.Errorf("Farbtext = %q", fall.eigen.Farbtext())
			case fall.eigen.Typtext() != fall.typ:
				t.Errorf("Typtext = %q", fall.eigen.Typtext())
			case fall.eigen.Wertetext() != fall.werte:
				t.Errorf("Wertetext = %q", fall.eigen.Wertetext())
			}
		})
	}
}

// **Leere Listen und kein `null`** — dieselbe Zusage wie überall an dieser
// Grenze, und hier sind es zwei: die Untertypen und die unbekannten Wörter.
//
// **Über die Abbildung unmittelbar**, weil der Weg darüber sie nicht
// erreicht: `TypzeileAus` gibt für „Creature" eine leere Liste und nicht
// `nil`, und eine leere Liste kommt nie in diesen Zweig. Er steht trotzdem
// da — eine Karte ohne Typzeile hat beide als `nil` (§3).
func TestLeereListenKommenAlsLeereListenAn(t *testing.T) {
	t.Parallel()

	dto := eigenschaftenAus(&domain.Sichteigenschaften{})

	switch {
	case dto == nil:
		t.Fatal("aus einem Verbund wurde nichts")
	case dto.Untertypen == nil:
		t.Error("die Untertypen sind nil statt leer")
	case dto.UnbekannteTypen == nil:
		t.Error("die unbekannten Typen sind nil statt leer")
	case dto.Farben == nil || dto.Typen == nil || dto.Supertypen == nil:
		t.Error("eine der drei Listen ist nil statt leer")
	}
}

// **Und nichts bleibt nichts.**
func TestOhneEigenschaftenBleibtEsNil(t *testing.T) {
	t.Parallel()

	if dto := eigenschaftenAus(nil); dto != nil {
		t.Fatalf("aus nichts wurde %+v", dto)
	}
}
