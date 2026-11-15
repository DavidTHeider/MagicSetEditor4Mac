// **Die Suche über alle vier Dienste** — Werkstück 8 von Phase 4.8
// (#1321),
// nach [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md).
//
// # Warum sie an derselben Naht hängt wie die Scryfall-Suche
//
// **Damit alle vier Flächen sie erben, ohne dass eine angefasst wird.**
// Fenster, Kommandozeile, MCP und Dienst rufen `Suchnaht`; wer dort einen
// Verteiler einsetzt, hat §16 eingelöst („dieselbe Fähigkeit, überall
// dieselbe Antwort") statt es viermal zu versprechen.
//
// # Und warum die drei neuen Klienten hier zusammenkommen
//
// `internal/service` kennt keinen von ihnen — es kennt eine **Fähigkeit**
// (`Suchdienst`). Wer die Klienten baut, entscheidet über Adresse, Frist,
// Drossel und Zugang; das sind Fragen der Verdrahtung (§2, §16).

package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtgapi"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/mtggraphql"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/netz"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scrydex"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// Geheimnisleser ist der Teil des Schlüsselbunds, den die Verdrahtung braucht.
//
// **Eine Schnittstelle und kein Paket**: `internal/app` soll den
// Schlüsselbund einsetzen können, und ein Lauf ohne echten Bund soll möglich
// sein, ohne dass jemand ihn auf diesem Rechner anlegt (§3).
type Geheimnisleser interface {
	Lies(ctx context.Context, dienst, konto string) (string, error)
}

// dienstklient baut den Netzklienten eines fremden Dienstes.
//
// **Die eine Stelle, an der das geschieht** — und deshalb die eine, an der
// die Sperre hängt. Vorher entstand je Dienst ein blanker
// `netz.NeuerKlient(…)`, und der Kommentar darüber behauptete, der Transport
// halte den Schalter „ohnehin". Er hielt ihn für **Scryfall**; die drei neuen
// gingen daran vorbei.
//
// **Gemessen** (`dienstschalterwirkt_test.go`): eine Suche über
// magicthegathering.io ging hinaus und kam mit einer Karte zurück,
// **nachdem** der Dienst abgeschaltet war. Damit war die Kernzusage der Phase
// — „aus heißt: kein einziger Aufruf, und das Umlegen wirkt sofort" — für
// drei von vier Diensten falsch.
//
// # Warum vor jeder Anfrage und nicht beim Bauen
//
// **Weil ein Schalter sich ändert.** Das Fenster ruft `Suchnaht` genau
// einmal, beim Start (`scryfallstart.go`). Wer danach umlegt, ändert die
// Einstellung und nicht die Karte, die schon gebaut ist. Derselbe Grund, aus
// dem `Scryfallklient` es so macht — dort steht er ausführlich.
func dienstklient(wahl Dienstwahl, dienst domain.Dienst) *netz.Klient {
	return dienstklientMit(netz.UeblicheFristen(), wahl, dienst)
}

// dienstklientMit tut dasselbe mit benannten Fristen.
//
// **Für den Abruf, von dem der Anbieter sagt, dass er lange dauert** — der
// volle CardTrader-Bestandsexport (Befund 6 des Zweitblicks über Phase 5.6).
func dienstklientMit(
	fristen netz.Fristen, wahl Dienstwahl, dienst domain.Dienst,
) *netz.Klient {
	return netz.NeuerKlientMit(fristen, proxyquelleAus(wahl)).
		MitErlaubnis(func(ctx context.Context) error {
			return pruefeNachAussen(ctx, wahl, dienst)
		})
}

// Dienstquellen liest die Karte der bereitstehenden Dienste — **je Anfrage**.
//
// **Befund 1 des Zweitblicks über Phase 4.8.** Sie stand vorher als feste
// Karte da, gebaut beim Verdrahten; das Fenster verdrahtet einmal, beim Start.
// Wer danach einen Dienst **einschaltete** oder einen Schlüssel eintrug, sah
// bis zum Neustart keine Wirkung — die Sperre am Transport löst nur die
// umgekehrte Richtung.
//
// # Warum der Neubau nichts kostet
//
// **Weil die Drossel nicht im Klienten wohnt.** Sie kommt aus dem
// [drosselwerk], das dieser Prozess mit allen Gegenstellen teilt — und über
// `MitAbsprache` sogar mit den anderen Prozessen dieser Anwendung. Ein
// Klient, den man wegwirft und neu baut, verliert damit keinen Abstand.
//
// **Das war vorher anders**, und es war der zweite Fehler an derselben
// Stelle: die drei neuen Dienste bekamen je Klient eine **eigene** Drossel
// (`netz.NeueDrossel(Abstand, …)`). Weil MCP und REST ihre Naht schon immer je
// Aufruf bauen, sah eine solche Drossel praktisch nie einen zweiten Aufruf —
// derselbe Befund, den der Zweitblick über Phase V für Scryfall gefunden hat
// (#888).
type Dienstquellen struct {
	wahl Dienstwahl
	bund Geheimnisleser
}

// Fuer gibt die Karte, die **jetzt** gilt.
func (d Dienstquellen) Fuer(ctx context.Context) (service.Quellen, error) {
	return Suchquellen(ctx, d.wahl, d.bund)
}

// Suchquellen baut die Suchdienste, die bereitstehen.
//
// # Warum ein Dienst ohne Zugang gar nicht in die Karte kommt
//
// **Weil „steht nicht bereit" und „hat nichts gefunden" zwei Auskünfte
// sind.** Wer Scrydex wählt, ohne seinen Schlüssel zu hinterlegen, soll das
// erfahren — und nicht „keine Treffer" lesen (§8). [service.Dienstsuche] sagt
// es, wenn der gewählte Dienst hier fehlt.
//
// # Und warum die Schalter hier noch einmal vorkommen
//
// Der **Transport** hält sie — seit dem 2026-09-02 für **jeden** der vier:
// die drei neuen über [dienstklient], Scryfall über seinen eigenen Bauplatz
// in `scryfallverdrahtung.go`. Hier entscheiden sie, ob ein Dienst in der Karte
// **steht**, und damit, ob die Fläche „abgeschaltet" sagt statt einer
// Fehlermeldung aus dem Netz.
func Suchquellen(
	ctx context.Context, wahl Dienstwahl, bund Geheimnisleser,
) (service.Quellen, error) {
	quellen := map[domain.Dienst]service.Suchdienst{}

	// **Warum ein Dienst fehlt, reist mit** — sonst rät die Meldung zwischen
	// „abgeschaltet" und „ohne Zugang" (siehe [service.Quellen]).
	grund := map[domain.Dienst]error{}

	// **Ein Scryfall ohne Zuhause ist kein Fehler des Ganzen.** Er braucht
	// ein Verzeichnis für seinen Zwischenspeicher; die anderen drei nicht.
	// Wer einen von ihnen gewählt hat, soll suchen können.
	if klient, err := Scryfallklient(wahl); err == nil {
		quellen[domain.DienstScryfall] = service.NeueSuche(klient)
	}

	an, err := angeschaltet(ctx, wahl)
	if err != nil {
		return service.Quellen{}, err
	}

	for _, dienst := range domain.Dienste() {
		if !an[dienst] {
			grund[dienst] = domain.DienstIstAus(dienst)
		}
	}

	if an[domain.DienstMTGAPI] {
		quellen[domain.DienstMTGAPI] = mtgapisuche{
			klient: mtgapi.NeuerKlient(mtgapiadresse,
				dienstklient(wahl, domain.DienstMTGAPI)).
				MitDrossel(drosselwerk.Fuer(mtgapiadresse, mtgapi.Abstand)),
		}
	}

	// **Ein Fehler an einem Dienst wird sein Grund, nicht der Abbruch aller.**
	//
	// **Zusatzbefund 1 des Nachlaufs** — und eine Regression, die erst mit der
	// Korrektur zu Befund 1 entstand: seit die Karte je Anfrage entsteht,
	// entsteht auch jeder Schlüsselbundzugriff je Anfrage. Ein `return` hier
	// hielt damit die Suche über **magicthegathering.io** an, weil Scrydex
	// danebenstand und sein Bund klemmte.
	//
	// Vorher fiel das nicht auf: der Fehler kam einmal, beim Verdrahten, und
	// die Naht kam gar nicht erst zustande.
	if an[domain.DienstMTGJSON] {
		quelle, fehler := mtgjsonquelle(ctx, wahl, bund)
		eintragen(quellen, grund, domain.DienstMTGJSON, quelle, fehler)
	}

	if an[domain.DienstScrydex] {
		quelle, fehler := scrydexquelle(ctx, wahl, bund)
		eintragen(quellen, grund, domain.DienstScrydex, quelle, fehler)
	}

	return service.Quellen{Bereit: quellen, Grund: grund}, nil
}

// eintragen legt eine Quelle in die Karte — oder ihren Grund daneben.
func eintragen(
	quellen map[domain.Dienst]service.Suchdienst,
	grund map[domain.Dienst]error,
	dienst domain.Dienst,
	quelle service.Suchdienst,
	fehler error,
) {
	if fehler != nil {
		grund[dienst] = fehler

		return
	}

	quellen[dienst] = quelle
}

// mtgjsonquelle baut die MTGJSON-Suche — oder sagt, was ihr fehlt.
func mtgjsonquelle(ctx context.Context, wahl Dienstwahl, bund Geheimnisleser) (
	service.Suchdienst, error,
) {
	marke, err := geheimnis(ctx, bund,
		domain.ZugangKonto(domain.DienstMTGJSON, domain.ZugangMarke))
	if err != nil {
		return nil, err
	}

	if marke == "" {
		return nil, domain.ZugangFehlt(domain.DienstMTGJSON, domain.ZugangMarke)
	}

	return mtgjsonsuche{
		klient: mtggraphql.NeuerKlient(mtggraphqladresse,
			dienstklient(wahl, domain.DienstMTGJSON), marke).
			MitDrossel(drosselwerk.Fuer(mtggraphqladresse, mtggraphql.Abstand)),
	}, nil
}

// scrydexquelle baut die Scrydex-Suche — oder sagt, was ihr fehlt.
func scrydexquelle(ctx context.Context, wahl Dienstwahl, bund Geheimnisleser) (
	service.Suchdienst, error,
) {
	zugang, err := scrydexzugang(ctx, bund)
	if err != nil {
		return nil, err
	}

	if fehlt := zugang.Fehlt(); fehlt != "" {
		return nil, domain.ZugangFehlt(domain.DienstScrydex, fehlt)
	}

	return scrydexsuche{
		klient: scrydex.NeuerKlient(scrydexadresse,
			dienstklient(wahl, domain.DienstScrydex), zugang).
			MitDrossel(drosselwerk.Fuer(scrydexadresse, scrydex.Abstand)),
	}, nil
}

// angeschaltet fragt für jeden Dienst, ob er gefragt werden darf.
//
// **Ein Lesefehler kommt an und heißt nicht »aus«.** Wer den Schalter nicht
// lesen kann, weiß nicht, was erlaubt ist — und eine Karte ohne diesen Dienst
// sähe aus wie eine bewusste Abschaltung. Dieselbe Überlegung wie bei
// [pruefeDienst], nur andersherum.
func angeschaltet(ctx context.Context, wahl Dienstwahl) (
	map[domain.Dienst]bool, error,
) {
	an := make(map[domain.Dienst]bool, len(domain.Dienste()))

	for _, dienst := range domain.Dienste() {
		gesetzt, err := wahl.IstAn(ctx, dienst)
		if err != nil {
			return nil, err
		}

		an[dienst] = gesetzt
	}

	return an, nil
}

// geheimnis holt eine Angabe aus dem Schlüsselbund.
//
// # »Steht nicht drin« ist kein Fehler
//
// **Ein fehlender Eintrag heißt: es gibt keinen Zugang** — und das ist der
// Normalfall für einen Dienst, den niemand eingerichtet hat. Ihn als Fehler
// zu melden hieße, dass die Suche stehen bleibt, weil jemand Scrydex nicht
// benutzt.
//
// **Jeder andere Fehler kommt an.** „Der Schlüsselbund antwortet nicht" ist
// etwas anderes als „da steht nichts", und wer beides zusammenwürfe,
// verschluckte den Fall, für den es die Meldung gibt (§8).
func geheimnis(ctx context.Context, bund Geheimnisleser, konto string) (
	string, error,
) {
	wert, err := bund.Lies(ctx, Schluesselbunddienst, konto)

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return "", nil
	case err != nil:
		return "", fmt.Errorf("den Zugang aus dem Schlüsselbund lesen: %w", err)
	}

	return wert, nil
}

// scrydexzugang holt beide Angaben aus dem Schlüsselbund.
//
// **Beide oder keine.** Scrydex verlangt zwei Kopfzeilen; eine allein reicht
// nicht, und ein halber Zugang wäre ein Klient, der sicher scheitert.
func scrydexzugang(ctx context.Context, bund Geheimnisleser) (
	scrydex.Zugang, error,
) {
	schluessel, schluesselfehler := geheimnis(ctx, bund,
		domain.ZugangKonto(domain.DienstScrydex, domain.ZugangSchluessel))
	if schluesselfehler != nil {
		return scrydex.Zugang{}, schluesselfehler
	}

	team, teamfehler := geheimnis(ctx, bund,
		domain.ZugangKonto(domain.DienstScrydex, domain.ZugangTeam))
	if teamfehler != nil {
		return scrydex.Zugang{}, teamfehler
	}

	return scrydex.Zugang{Schluessel: schluessel, Team: team}, nil
}

// Schluesselbunddienst ist der Name, unter dem dieses Programm im
// Schlüsselbund steht.
//
// **Ein Name für alle Dienste**, und das Konto unterscheidet sie
// ([domain.ZugangKonto]). Wer in der Schlüsselbundverwaltung nachsieht,
// findet alles unter einem Eintragstyp beisammen.
const Schluesselbunddienst = "MagicSetEditor4Mac"
