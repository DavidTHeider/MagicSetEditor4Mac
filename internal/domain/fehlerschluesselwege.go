package domain

// **Die Wege zur Anmeldung der Schlüssel** — aus `fehlerschluessel.go`
// herausgelöst, als DOM-04
// (#1699)
// die Datei über die Zeilengrenze aus §2 brachte. Dort stehen die Schlüssel,
// hier ihre Anmeldung und wie man sie fragt — die Anmeldung kam mit DOM-02
// (#1697)
// nach, aus demselben Grund.

import "slices"

// Fehlerschluessel gibt die angemeldeten Schlüssel als Kopie.
func Fehlerschluessel() []Schluessel {
	angemeldet := make([]Schluessel, len(alleSchluessel))
	copy(angemeldet, alleSchluessel)

	return angemeldet
}

// IstAngemeldet sagt, ob ein Schlüssel in der Anmeldung steht.
func IstAngemeldet(schluessel Schluessel) bool {
	return slices.Contains(alleSchluessel, schluessel)
}

// alleSchluessel ist die Anmeldung.
//
// Von Hand geführt und nicht aus den Konstanten abgeleitet: Go gibt die
// Konstanten eines Pakets zur Laufzeit nicht her. Damit die Liste nicht
// zurückbleibt, liest `internal/app/fehlerschluessel_test.go` den Quelltext
// nach und weist eine Konstante ab, die hier fehlt.
var alleSchluessel = []Schluessel{
	SchluesselUnbekannt,
	SchluesselUnbekanntMitVorgang,
	SchluesselNichtGefunden,
	SchluesselKonflikt,
	SchluesselPflichtfeld,
	SchluesselValidierung,
	SchluesselValidierungsgrund,
	SchluesselSicherungNeuereFassung,
	SchluesselSicherungAeltereFassung,
	SchluesselSicherungUnbrauchbar,
	SchluesselSicherungVerworfen,
	SchluesselSicherungUnvermerkt,
	SchluesselSicherungFremdeFassung,
	SchluesselVorlageFehlt,
	SchluesselKeinSetMitStil,
	SchluesselOhnePeeridentitaet,
	SchluesselPeerNichtErreichbar,
	SchluesselPeerOhneAdresse,
	SchluesselZeileNichtVerbuchbar,
	SchluesselZusageUnpassend,
	SchluesselOhneNetzerlaubnis,
	SchluesselOhneNetzroute,
	SchluesselSchluesselbundAblage,
	SchluesselEinladungZuNeu,
	SchluesselFremderNutzer,
	SchluesselNichtFreigegeben,
	SchluesselFernhorchenLaeuft,
	SchluesselAnmeldungAbgelaufen,
	SchluesselUmschlagFassung,
	SchluesselUmschlagUnterschrift,
	SchluesselUmschlagInhalt,
	SchluesselBelegt,
	SchluesselBestandWirdGetauscht,
	SchluesselAnwendungEndet,
	SchluesselVorgangAbgebrochen,
	SchluesselErlaubnis,
	SchluesselSchluesselFehlt,
	SchluesselNurLesend,
	SchluesselBestandUnbrauchbar,
	SchluesselBestandNeuereFassung,
	SchluesselBestandAeltereFassung,
	SchluesselBestandNurLesbar,
	SchluesselBestandEA,
	SchluesselKaumPlatz,
	SchluesselStartUngeklaert,
	SchluesselUnbekannteHerkunftsart,
	SchluesselHerkunftOhneProfil,
	SchluesselSkriptfehler,
	SchluesselZweiterFaktor,
	SchluesselFaktorGesperrt,
	SchluesselKennwortFalsch,
	SchluesselKennwortLaenge,
	SchluesselSperreOhneRueckweg,
	SchluesselLokaleEntsperrungFehlt,
	SchluesselCodeFalsch,
	SchluesselNichtEditierbar,
	SchluesselBerechnetesFeld,
	SchluesselKarteInDeck,
	SchluesselOrdnerangabe,
	SchluesselOrdnerFehlt,
	SchluesselOrdnerIstDatei,
	SchluesselOrdnerGesperrt,
	SchluesselOrdnerVoll,
	SchluesselFalschesSpiel,
	SchluesselTurnierformatSpiel,
	SchluesselTurnierAbgeschlossen,
	SchluesselSitzGegangen,
	SchluesselFremderWert,
	SchluesselVorschlagNurGegenseite,
	SchluesselPaketGeaendert,
	SchluesselTurnierNichtFertig,
	SchluesselTurniernameLeer,
	SchluesselTurnierspielLeer,
	SchluesselPaketnameUngueltig,
	SchluesselBefehlBestaetigen,
	SchluesselBefehlArgumente,
	SchluesselDoppelt,
	SchluesselDoppelterName,
	SchluesselDoppeltesExemplar,
	SchluesselKonfliktgrund,
	SchluesselVoll,
	SchluesselNetzName,
	SchluesselNetzNichtErreichbar,
	SchluesselNetzKeineAntwort,
	SchluesselNetzAntwortStockt,
	SchluesselNetzAbgerissen,
	SchluesselNetzProxyAnmeldung,
	SchluesselGedrosselt,
	SchluesselGedrosseltBis,
	SchluesselKontingent,
	SchluesselProzessVeraltet,
	SchluesselBestandVomDienstGehalten,
	SchluesselGesperrt,
	SchluesselSicherungskennwortKurz,
	SchluesselSicherungskennwortFehlt,
	SchluesselSicherungskennwortFalsch,
	SchluesselSicherungImKlartext,
	SchluesselPlanVerschluesselt,
	SchluesselMTGJSONKeinJSON,
	SchluesselMTGJSONKeinSet,
	SchluesselSymbolZuVieleTeile,
	SchluesselSymbolZuVielePunkte,
	SchluesselSymbolZuTief,
	SchluesselSymbolZuGross,
	SchluesselBildZuGross,
	SchluesselSymbolGespeichertZuGross,
	SchluesselSetOhneSymbol,
	SchluesselSymbolVerwendet,
	SchluesselSymbolOhneVorige,
	SchluesselSprachkennungUngueltig,
	SchluesselSprachkennungEingebaut,
	SchluesselPlatzhalterAbweichend,
	SchluesselAuszeichnungUnsicher,
	SchluesselSprachpaketNeuer,
	SchluesselSprachpaketOhneLizenz,
	SchluesselFremdeEigeneKarten,
	SchluesselNichtGelegtePlaetze,
	SchluesselFremdeKartenUndNichtGelegte,
	SchluesselBereichAbgeschaltet,
	SchluesselDienstAus,
}
