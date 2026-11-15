package mtgjsonpreise

// **Der Bulkbezug von MTGJSON** — Werkstück A3 von Phase 5.3
// (#1442).
//
// # Zwei Dateien, zwei Fragen
//
// **Nachgelesen am 2026-09-11** unter <https://mtgjson.com/downloads/all-files/>:
//
//   - `AllPrices.json` — „all prices … for the past **90 days**"
//   - `AllPricesToday.json` — „all prices, for the **current day**"
//
// Beide gibt es als `.json`, `.json.gz`, `.json.bz2`, `.json.xz` und
// `.json.zip`. Geholt wird **gzip**: das kann Go von sich aus, und ein
// zusätzliches Paket für bz2 oder xz wäre eine Abhängigkeit für nichts.
//
// **Was daraus folgt und im Issue steht**: die Historie „zeichnet kein
// jährliches Datenangebot vor". Neunzig Tage sind neunzig Tage; wer länger
// zurückschauen will, braucht eigene Abrufe über die Zeit, und **Lücken sind
// sichtbar**, weil jeder Stand seinen Beobachtungstag trägt.
//
// # Was hier nicht steht
//
// **Kein Zwischenspeicher und keine Uhr.** Wann zuletzt geholt wurde und ob
// das reicht, ist eine Frage des Aufrufers; dieses Paket ist ein Blatt (§2)
// und holt, wenn es gerufen wird.

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
)

// Umfang sagt, welche der beiden Dateien gemeint ist.
type Umfang string

const (
	// UmfangHeute ist `AllPricesToday.json` — der laufende Tag.
	UmfangHeute Umfang = "heute"

	// UmfangNeunzigTage ist `AllPrices.json` — die letzten neunzig Tage.
	//
	// **Über ein Gigabyte ausgepackt.** Wer sie holt, holt einen Verlauf;
	// wer nur den heutigen Stand braucht, nimmt [UmfangHeute].
	UmfangNeunzigTage Umfang = "neunzigTage"
)

// dateinamen ordnet jedem Umfang seine Datei zu.
//
// **An einer Stelle** (§16, dritte Schicht): über den Code verteilt liefe die
// Zuordnung beim ersten neuen Umfang auseinander.
func dateinamen() map[Umfang]string {
	return map[Umfang]string{
		UmfangHeute:       "AllPricesToday.json.gz",
		UmfangNeunzigTage: "AllPrices.json.gz",
	}
}

// Vorgabeadresse ist der Ort, an dem MTGJSON seine Dateien führt.
//
// **Als Konstante und nicht fest im Code**: ein Lauf gegen einen eigenen
// Server braucht sie, und ohne sie ginge jeder Test ins Netz (§3).
const Vorgabeadresse = "https://mtgjson.com/api/v5"

// Klient holt die Preisdateien.
//
// **Adresse und Netzklient kommen von außen** (§16): sonst braucht jeder Lauf
// das Netz, und was das wert ist, sagt §3.
type Klient struct {
	adresse string
	klient  *netz.Klient
	kenner  netz.Kenner
	grenzen Grenzen
}

// NeuerKlient baut ihn.
//
// **Die Anwendung nennt sich** (§16, Satz 2): `kenner` trägt den
// `User-Agent`, und MTGJSON sieht, wer fragt.
func NeuerKlient(
	adresse string, klient *netz.Klient, kenner netz.Kenner, grenzen Grenzen,
) *Klient {
	if adresse == "" {
		adresse = Vorgabeadresse
	}

	return &Klient{
		adresse: strings.TrimRight(adresse, "/"),
		klient:  klient,
		kenner:  kenner,
		grenzen: grenzen.mitVorgaben(),
	}
}

// Familie ist der Name, unter dem alles läuft, was von hier kommt.
//
// **Ohne Anbieter**, anders als an den einzelnen Ständen: die tragen
// `mtgjson:cardkingdom` und `mtgjson:cardmarket`, weil das verschiedene
// Aussagen sind. Der **Widerruf** gilt für alles zusammen — was MTGJSON für
// einen Druck nicht mehr nennt, ist weggefallen, gleich welcher Anbieter es
// war ([domain.Preisquelle.GehoertZu]).
const Familie domain.Preisquelle = "mtgjson"

// Name sagt, wer hier antwortet.
func (k *Klient) Name() domain.Preisquelle { return Familie }

// Holen lädt die Preisdatei und gibt die Stände der genannten Kennungen.
//
// **Strömend und ohne Zwischendatei**: die Datei wiegt bis zu einem Gigabyte,
// und was davon gebraucht wird, sind die Zeilen zu den Karten im Bestand —
// bei einer Sammlung von tausend Karten ein Promille. Sie erst abzulegen und
// dann zu lesen kostete Platz für nichts.
//
// **Der Abbruch wirkt** (§16, Satz 3): `ctx` geht an die Anfrage, und ein
// abgebrochener Ruf lässt keinen halben Stand zurück — es wird nichts
// geschrieben.
// **Der Umfang kommt als Text**, nicht als [Umfang]: er stammt von einer
// Fläche — einem Befehl, einer Abfrage, einem Werkzeugargument —, und was
// davon ein gültiger Umfang ist, weiß dieses Paket. Eine Prüfung im Dienst
// wäre eine zweite Wahrheit darüber (§16, dritte Schicht).
func (k *Klient) Holen(ctx context.Context, umfang string, crosswalk map[string]string) (
	[]domain.Preisstand, []domain.Befund, error,
) {
	adresse, fehler := k.adresseFuer(Umfang(umfang))
	if fehler != nil {
		return nil, nil, fehler
	}

	antwort, holfehler := netz.Strom(ctx, k.klient, nil, k.kenner, adresse)
	if holfehler != nil {
		return nil, nil, holfehler
	}

	defer func() { _ = antwort.Rumpf.Close() }() //nolint:errcheck // gelesen ist gelesen

	// **Über `netz.StatusGut` und nicht über `net/http`** — dieselbe Zahl, und
	// der Unterschied ist die Netzgrenze (§16): dieses Paket geht über `netz`
	// wie `mtgapi` und `scrydex`, und ein `net/http` darin sähe für den
	// Wächter aus wie ein eigener Transport.
	if antwort.Status != netz.StatusGut {
		// **Der Rumpf wird weggelesen und nicht gezeigt**: eine Fehlerseite
		// kann ein Megabyte HTML sein, und in einer Meldung steht sie schlecht.
		_, _ = io.Copy(io.Discard, io.LimitReader(antwort.Rumpf, 1<<20)) //nolint:errcheck // verworfen

		return nil, nil, fmt.Errorf(
			"%w: MTGJSON antwortet auf %s mit %d",
			domain.ErrValidation, dateinamen()[Umfang(umfang)], antwort.Status)
	}

	return LiesGepackt(antwort.Rumpf, crosswalk, k.grenzen)
}

// adresseFuer baut die Adresse einer Datei.
//
// **Über `url.JoinPath` und nicht über eine Verkettung**: eine Adresse aus
// den Einstellungen kann auf `/` enden oder einen Pfad tragen, und `a + "/" +
// b` ergäbe dann `//`.
func (k *Klient) adresseFuer(umfang Umfang) (string, error) {
	name, bekannt := dateinamen()[umfang]
	if !bekannt {
		return "", fmt.Errorf("%w: den Umfang %q gibt es nicht",
			domain.ErrValidation, string(umfang))
	}

	adresse, fehler := url.JoinPath(k.adresse, name)
	if fehler != nil {
		return "", domain.MitGrund(domain.ErrValidation, fehler,
			"die Adresse für %s bauen", name)
	}

	return adresse, nil
}
