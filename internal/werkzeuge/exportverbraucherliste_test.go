package werkzeuge

// Die Klassen, in die eine exportierte Funktion ohne Verbraucher im Betrieb
// fallen darf — **jede andere** ist ein Befund und steht mit ihrem Issue da.
const (
	// klasseAufzaehlung nennt alle Werte einer Art, damit ein Wächter oder Lauf
	// über sie gehen kann — oder gibt einem Wächter eine Auskunft, die er
	// gegen eine Datei hält.
	klasseAufzaehlung = "Aufzählung oder Auskunft für Wächter"

	// klasseNaht lenkt um oder legt frei, was der Betrieb fest setzt — **nur** für
	// Tests, und ihr Kopf sagt es.
	klasseNaht = "Naht für Tests"

	// klasseMesslauf rechnet für einen Lauf hinter einem Build-Tag.
	klasseMesslauf = "für einen Messlauf hinter einem Build-Tag"

	// klasseHuelle ist eine Hülle um einen Weg, den der Betrieb unter anderem Namen
	// nimmt — für Tests bequemer, fachlich nichts Eigenes.
	klasseHuelle = "Hülle um einen Weg des Betriebs"

	// klasseVorgebaut ist eine Fachregel, die ein **offenes Arbeitspaket**
	// anbindet — die Zeile nennt es, und seine Abnahme ist der Lauf, der sie
	// braucht. Fällt das Paket weg, ist sie ein Befund (#2436).
	klasseVorgebaut = "vorgebaut für ein offenes Arbeitspaket"
)

// nurVonTestsGenommen sind die exportierten Funktionen, die außerhalb der
// Tests niemand nimmt — je mit ihrer Klasse (SEC04,
// #2357),
// gezählt am 2026-10-16.
var nurVonTestsGenommen = map[string]string{
	"app.AdresseAus":                        klasseNaht,
	"app.Befehlsverweise":                   klasseAufzaehlung,
	"app.Bindungsvertrag":                   klasseAufzaehlung,
	"app.EndungVon":                         klasseHuelle,
	"app.ErstesMuster":                      klasseHuelle,
	"app.Menuerollen":                       klasseAufzaehlung,
	"app.MitDienstadressen":                 klasseNaht,
	"app.MitFreigabeadresse":                klasseNaht,
	"app.MitFreigabebund":                   klasseNaht,
	"app.MitSchriftwurzel":                  klasseNaht,
	"app.MitScryfalladresse":                klasseNaht,
	"app.NeueKartenskriptnaht":              klasseNaht,
	"app.NutzerflaecheOhneSensor":           klasseNaht,
	"app.Protokoll":                         klasseHuelle,
	"app.SammeldatenUeber":                  klasseNaht,
	"app.SchluesselwortpflegeUeber":         klasseHuelle,
	"app.Utf16StelleFuerTest":               klasseNaht,
	"domain.Abstandsmasse":                  klasseAufzaehlung,
	"domain.Auszaehlen":                     klasseHuelle,
	"domain.Balancefelder":                  klasseAufzaehlung,
	"domain.Bedienmodi":                     klasseAufzaehlung,
	"domain.Bilderkennungsfelder":           klasseAufzaehlung,
	"domain.Ebenenschluessel":               klasseAufzaehlung,
	"domain.Einschaetzungsarten":            klasseAufzaehlung,
	"domain.EinstellungsArt":                klasseAufzaehlung,
	"domain.Fehlerschluessel":               klasseAufzaehlung,
	"domain.Fensterseiten":                  klasseAufzaehlung,
	"domain.Fenstersprachen":                klasseAufzaehlung,
	"domain.FragenVon":                      klasseAufzaehlung,
	"domain.Handelsnachrichtenarten":        klasseAufzaehlung,
	"domain.IstDienstAus":                   klasseHuelle,
	"domain.IstInvariante":                  klasseAufzaehlung,
	"domain.IstRegelrang":                   klasseAufzaehlung,
	"domain.Journalanlaesse":                klasseAufzaehlung,
	"domain.KIBildgroessen":                 klasseAufzaehlung,
	"domain.KIBildgueten":                   klasseAufzaehlung,
	"domain.KIFaehigkeiten":                 klasseAufzaehlung,
	"domain.KIFeldherkunftsfelder":          klasseAufzaehlung,
	"domain.Kennungssysteme":                klasseAufzaehlung,
	"domain.Kostenstaende":                  klasseAufzaehlung,
	"domain.Kostentafel":                    klasseAufzaehlung,
	"domain.Laeufe":                         klasseAufzaehlung,
	"domain.Messschluessel":                 klasseAufzaehlung,
	"domain.NutzerArt":                      klasseAufzaehlung,
	"domain.Partiestaende":                  klasseAufzaehlung,
	"domain.Posten":                         klasseHuelle,
	"domain.Protokollschluessel":            klasseAufzaehlung,
	"domain.PruefeDienstordner":             klasseHuelle,
	"domain.Rechercheleeren":                klasseAufzaehlung,
	"domain.Regelregister":                  klasseAufzaehlung,
	"domain.Ressourcenklassen":              klasseAufzaehlung,
	"domain.Retrievalquellen":               klasseAufzaehlung,
	"domain.Rollen":                         klasseAufzaehlung,
	"domain.Rufausgaenge":                   klasseAufzaehlung,
	"domain.Scopes":                         klasseAufzaehlung,
	"domain.Uebersteuerbare":                klasseAufzaehlung,
	"domain.VertragFuer":                    klasseAufzaehlung,
	"domain.Wirkungsarten":                  klasseAufzaehlung,
	"domain.ZeileZuReferenz":                klasseVorgebaut, // P11 (#2379): die Sichtanfrage löst eine Sitzungsreferenz auf
	"domain.Zeugnispfade":                   klasseAufzaehlung,
	"domain.Zugfolge":                       klasseAufzaehlung,
	"service.Auszaehlwerte":                 klasseMesslauf,
	"service.Bewerte":                       klasseMesslauf,
	"service.Gueteschnitt":                  klasseMesslauf,
	"service.Kennungen":                     klasseMesslauf,
	"service.LaufzeitwerkZuruecksetzen":     klasseNaht,
	"service.MesswerkZuruecksetzen":         klasseNaht,
	"service.MitBezugsfrist":                klasseNaht,
	"service.NamenDerBibliothek":            klasseNaht, // der Namenstest in exportvorlage_test.go (#2436)
	"service.NeueTexturfachprobe":           klasseNaht,
	"service.PruefeEinbettung":              klasseMesslauf,
	"service.PruefeNeuordnung":              klasseMesslauf,
	"service.Skriptname":                    klasseAufzaehlung,
	"service.SpurUhrSetzen":                 klasseNaht,
	"service.SymbolbilderFuerTest":          klasseNaht,
	"service.SymbolschriftnamenFuerTest":    klasseNaht,
	"service.UmgebungMitSchluesselwoertern": klasseNaht,
	"service.ZaehltFuerPortfolio":           klasseNaht,
	"storage.Datenstand":                    klasseHuelle,
	"storage.FuerLaeufe":                    klasseNaht,
	"storage.MigriereMit":                   klasseNaht,
	"storage.OeffnenMitSchluessel":          klasseNaht,
}
