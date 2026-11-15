import { dienstkatalog } from './einstellungenkatalog'
import { einstellungsbelegung } from './einstellungsbelegung'
import { einstellungsgruppennamen } from './einstellungsgruppennamen'
import {
  schalternamen,
  schalterwerte,
  seitenschalter,
} from './einstellungenschalter'

// **Die Texte der Einstellungsseite** — ausgelagert für Phase 4.8
// (#1321).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als
// einer ankommt** (§9) — dieselbe Gestalt wie `seitenkarten.ts`. Mit den
// Dienstschaltern ging `seiten.ts` über die Zeilengrenze aus §2.
import { ki } from './seitenki'
import { netzbereich, netzhilfe, netznamen } from './einstellungennetz'
import { handelshilfe, handelsnamen, handelswerte } from './einstellungenhandel'
import {
  gestaltungsbeiwerk,
  gestaltungshinweise,
  gestaltungsnamen,
  gestaltungswerte,
} from './einstellungengestaltung'
import { zugaenge } from './einstellungenzugaenge'
import { ansichtsnamen, ansichtswerte } from './einstellungenansicht'
import { vorgabennamen } from './einstellungenvorgaben'
import {
  protokollhinweise,
  protokollnamen,
  protokollwerte,
} from './einstellungenprotokoll'
import { austausch } from './einstellungenaustausch'
import { eigenuebersetzung } from './eigenuebersetzung'
import { kihinweise, kinamen, kiwerte } from './einstellungenki'

export const einstellungen = {
  ...einstellungsgruppennamen,
  ...seitenschalter,

  titel: 'Impostazioni',

  namen: {
    ...schalternamen,
    sprache: 'Lingua',
    'sprache.eigene': 'Traduzione personale',
    darstellung: 'Formato regionale',
    zeitzone: 'Fuso orario',
    'filter.gespeichert': 'Filtri salvati',
    'kuerzel.ziffern': 'Assegnazione da ⌘0 a ⌘9',
    erscheinungsbild: 'Tema',
    schriftgroesse: 'Dimensione del testo',
    aktualisierung: 'Controlla nuove versioni',
    ersteschritte: 'Introduzione all’avvio',
    seitenleiste: 'Barra laterale',
    leistenlage: 'Posizione della barra laterale',
    fensterstart: 'Finestra all’avvio',
    fensterbreite: 'Ultima larghezza memorizzata',
    fensterhoehe: 'Ultima altezza memorizzata',
    schriftart: 'Carattere di visualizzazione',
    feldtiefe: 'Modulo della carta',
    'tisch.autoanordnung': 'Disponi automaticamente le carte sul tavolo',

    'barrierefreiheit.bewegung': 'Movimento',
    'barrierefreiheit.kontrast': 'Contrasto',
    'sicherheit.faktorbeimwechsel': 'Secondo fattore al cambio di profilo',

    ...vorgabennamen,
    ...protokollnamen,
    // **Die lokale KI-Laufzeit** — A7, #1446.
    // **A cloud provider's address** — piece 3 of A18 (#1533), labelled with
    // #1540.
    'dienst.anthropic.adresse': 'Indirizzo di Anthropic',
    'dienst.openai.adresse': 'Indirizzo di OpenAI',
    'dienst.xai.adresse': 'Indirizzo di xAI',
    ...netznamen,
    ...handelsnamen,
    ...gestaltungsnamen,
    ...ansichtsnamen,
    'konto.schau': 'Leggi il proprio account di mercato',
    ...kinamen,
    // **The client id per provider** — Werkstück 4 of Phase 4.12 (#1351).
    'anbieter.github.kennung': 'ID client su GitHub',
    'anbieter.google.kennung': 'ID client su Google',
    'anbieter.microsoft.kennung': 'ID client su Microsoft',
    'anbieter.x.kennung': 'ID client su X',

    // **The workspace and the two limits** — Werkstück O3, Phase 5.2 (#1427).
    'arbeitsplatz.navigator': 'Colonna sinistra',
    'arbeitsplatz.inspektor': 'Colonna destra',
    'arbeitsplatz.navigatorbreite': 'Larghezza della colonna sinistra',
    'arbeitsplatz.inspektorbreite': 'Larghezza della colonna destra',
    'richtlinie.nurlokal': 'Lavora solo in locale',
    'richtlinie.kostengrenze':
      'Limite di costo nell’unità più piccola (per EUR: centesimi)',
  },

  // Origin, default and search — work item U11 of phase 5.1 (#1422).
  herkunftMarke: 'impostato da {herkunft}',
  wirksamKurz: 'Vale “{wert}”.',
  wirksamOffenKurz: 'Vale l’impostazione predefinita del computer.',
  wirksamGehalten: 'Vale “{wert}” — lo impone una policy.',
  herkunftZeile: 'Impostato da {herkunft}, livello {ebene}.',
  herkunftNamen: {
    profil: 'questo profilo',
    nutzer: 'te',
    geraet: 'questo dispositivo',
    vorgabe: 'nessuno (predefinito)',
  },
  ebenenNamen: {
    profil: 'profilo',
    nutzer: 'utente',
    geraet: 'dispositivo',
    arbeitsplatz: 'area di lavoro',
  },
  ebenenErklaerung:
    'Ogni impostazione vale a un livello: dispositivo, utente, profilo o area di lavoro. **Dispositivo** descrive questo computer — un percorso, una dimensione della finestra — e non viene incluso in un trasferimento. **Utente** vale per te in ciascuno dei tuoi profili. **Profilo** vale solo qui e ha la precedenza su ciò che hai impostato come utente. **Area di lavoro** è la disposizione delle colonne di questa finestra. Vale ciò che è impostato nel profilo, altrimenti ciò che hai impostato come utente, altrimenti il valore predefinito — per una policy, il valore più restrittivo. Nessuna impostazione richiede diritti di amministratore.',
  fachleute: 'Che altro possono impostare qui gli esperti',
  fachleuteSatz:
    'Timeout, raggiungibilità tramite altre reti, la via d’uscita e i servizi per altri programmi. Espandere non cambia nulla — mostra solo di più.',
  zuruecksetzen: 'Ripristina su “{vorgabe}”',
  erbeMitWert: '{erbe} ({wert})',
  zuruecksetzenVon: 'Ripristina {name} al valore predefinito',
  suche: 'Trova un’impostazione',
  suchePlatzhalter: 'Nome, chiave o valore',
  sucheTreffer: '{zahl} risultati',
  sucheKeine: 'Nessuna impostazione corrisponde a “{frage}”.',
  zumGebiet: 'Mostra nell’area “{gebiet}”',

  // **Below the appearance settings** — UX-09 (#1681).
  gestaltung: gestaltungsbeiwerk,

  hinweise: {
    ...protokollhinweise,
    'sprache.eigene':
      'Quale traduzione personale usa la finestra — vuoto significa nessuna. Se non esiste più, torna a valere la lingua scelta.',
    feldtiefe:
      'Quanto del modulo della carta viene mostrato. **Normale** è ciò che compili; **Avanzato** aggiunge ciò che calcola uno script del pacchetto di gioco; **Esperto** aggiunge ciò che il pacchetto di gioco non spiega affatto — `stylesheet` e simili. I tre sono annidati: più in profondità significa di più, mai qualcosa di diverso. Vale per profilo.',
    sprache:
      'Vale per l’intera finestra e per la barra dei menu. Ciò che proviene da un pacchetto MSE — nomi dei campi, nomi dei giochi — resta come scritto nel file. Il tedesco è la lingua di partenza; le altre lingue sono tradotte automaticamente e non ancora revisionate.',
    darstellung:
      'Come appaiono numeri, importi e date — ad esempio **de-CH** o **en-GB**. Vuoto significa: come nel sistema. La lingua dei testi non cambia.',
    zeitzone:
      'In quale fuso vengono mostrati gli orari — ad esempio **Europe/Berlin**. «Come il sistema» indica quello del computer. Viene memorizzato per profilo e non viene incluso in un trasferimento.',
    'filter.gespeichert':
      'I filtri salvati di ogni vista come elenco in JSON — nome, vista, contrassegni e se vale **e** oppure **o**. Si creano nel navigatore della rispettiva vista; appartengono al profilo e viaggiano con l’esportazione delle impostazioni.',
    'kuerzel.ziffern':
      'Quale pagina apre una cifra con ⌘. Vuoto significa: le aree della barra laterale. Si imposta nella tabella delle scorciatoie sotto “Generali”.',
    erscheinungsbild:
      'Chiaro, scuro, o ciò che dice il computer. La scelta vale solo per questa applicazione.',
    schriftgroesse:
      'Ridimensiona l’intera interfaccia, non solo il testo — spaziature e bordi crescono con essa. L’anteprima della carta resta com’è: mostra ciò che verrebbe stampato.',
    aktualisierung:
      'L’impostazione predefinita è **disattivato**: nulla esce verso la rete e non viene chiesto nulla. Con **solo notifica**, all’avvio l’applicazione controlla se esiste una versione più recente e lo dice — nient’altro. Con **notifica e scarica** offre in più un pulsante che scarica la versione nella cartella dei download e ne confronta la somma di controllo; nulla viene installato da solo.',
    ersteschritte:
      'I tre passi verso la tua prima carta, come finestra all’avvio. Lo stesso interruttore è dentro quella finestra — disattivalo lì e lo ritrovi qui.',
    seitenleiste:
      'La barra a sinistra con le aree dell’applicazione. Ripiegata, la sua larghezza va agli spazi di lavoro; ⌥⌘S, il pulsante sul suo bordo e la voce sotto **Vista** la ripiegano e la espandono. Viene memorizzata per profilo e non viene inclusa in un trasferimento.',
    leistenlage:
      'A sinistra sta in verticale accanto al contenuto, in alto come riga sopra di esso. Le stesse voci, scorciatoie e tasti freccia — solo in orizzontale. Viene memorizzata per profilo e non viene inclusa in un trasferimento.',
    fensterstart:
      'Il valore predefinito è **Riempi lo schermo** — la finestra occupa lo spazio che ha, e **non lo schermo intero**: si porterebbe via la barra dei menu. **Dimensione normale** apre alla dimensione integrata, **Come l’ultima volta** com’era quando l’hai lasciata. Viene memorizzata per profilo e non viene inclusa in un trasferimento.',
    fensterbreite:
      'Quanto è stato misurato all’ultima chiusura della finestra — **solo per “Come l’ultima volta”**. Vuoto significa che non c’è ancora stata un’ultima volta, e la finestra si apre alla dimensione integrata.',
    fensterhoehe:
      'Lo stesso per l’altezza. Entrambe vengono scritte alla chiusura della finestra; una misura fuori da 700 a 8.192 punti non viene memorizzata — una finestra che non entra nello schermo successivo non si potrebbe più afferrare.',
    schriftart:
      'Il carattere per titoli e nomi dei set. Il testo corrente resta il carattere di sistema — un carattere decorativo su un paragrafo si legge male. Tutti e quattro sono forniti con macOS; non viene scaricato nulla.',
    'barrierefreiheit.bewegung':
      'L’override va **in una sola direzione**: disattiva il movimento, non lo attiva mai. Chi ha impostato **Riduci movimento** nel sistema qui comunque non ottiene movimento — un’applicazione che lo scavalcasse lavorerebbe contro di lui.',
    'barrierefreiheit.kontrast':
      'Stessa direzione: **alto** elimina la texture a fibre e alza i colori finché ogni testo tiene 7 : 1. Chi chiede al sistema più contrasto ottiene lo stesso.',
    'sicherheit.faktorbeimwechsel':
      'L’**unico** override che allenta — e solo al **cambio** di profilo. Non cambia nulla per l’accesso: lì il secondo fattore fa ciò per cui esiste. Su un computer con una sola persona, un codice monouso a ogni cambio è un attrito senza vantaggio.',
    'vorgabe.zustand':
      'Con che cosa parte una nuova riga della collezione. **Nessun valore predefinito** significa che il campo resta vuoto, come è sempre stato.',
    'vorgabe.waehrung':
      'Con che cosa parte il campo valuta di una nuova riga, ad esempio **EUR**. Non viene convertito nulla.',
    'vorgabe.einzelspiel':
      'Se un nuovo torneo svolge incontri a **partita singola**. Decide la visualizzazione, non il calcolo — la percentuale di partite vinte e quella degli avversari vengono comunque conservate.',
    'vorgabe.hochladepfad':
      'La cartella da cui parte una finestra di apertura file — immagini, MTG JSON e tutto il resto che arriva dal disco. Un percorso completo o uno che inizia con **~/**; vuoto significa dove lo apre il sistema operativo.',
    'vorgabe.ausfuhrpfad':
      'Da dove parte una finestra di salvataggio quando l’azione non ha un luogo proprio. Se la cartella manca al momento del salvataggio — ad esempio un disco scollegato —, l’app te lo dice invece di salvare altrove.',
    'vorgabe.sicherungspfad':
      'Da dove parte la finestra per un backup del tuo archivio. Una cartella mancante interrompe il backup con un messaggio.',
    'vorgabe.diagnosepfad':
      'Dove vengono salvati testo diagnostico e pacchetti. Una cartella mancante si ferma con un messaggio.',
    'bildausfuhr.ordner':
      'Dove l’uscita delle immagini ha scritto l’ultima volta. La pagina Immagini la propone e registra qui ogni cartella scelta di nuovo — una per tutti i set, come in MSE. Vuoto significa che la finestra delle cartelle si apre sulla cartella per i file che salvi.',
    'ausgabe.ziele':
      'Quale destinazione «Esporta» ha mostrato per ultima in ogni set. Un elenco in JSON, la scelta più recente in fondo, al massimo 200; si scrive scegliendo sotto «Dove?». Vuoto significa che nulla è preselezionato.',
    'dienst.anthropic.adresse':
      'Dove vanno le richieste. Vuoto significa l’interfaccia propria di Anthropic. Un indirizzo diverso è un **gateway** — un servizio che inoltra la richiesta; la sua chiave viene memorizzata separatamente.',
    'dienst.openai.adresse':
      'Lo stesso per OpenAI. Vuoto significa `https://api.openai.com`. Tutto ciò che inserisci qui significa che non parli più con OpenAI ma con ciò che risponde a quell’indirizzo.',
    'dienst.xai.adresse': 'Lo stesso per xAI. Vuoto: `https://api.x.ai`.',
    ...netzhilfe,
    ...handelshilfe,
    ...gestaltungshinweise,
    'konto.schau':
      'Se questo programma può leggere il tuo account **personale** su un marketplace — le tue inserzioni e i tuoi ordini. **Disattivato** significa: nemmeno uno sguardo. Attivare il marketplace non basta: quello risponde a ciò che chiedono **altri** venditori, questo a ciò che sta **presso di te**. Lì non viene mai scritto nulla — nessuna inserzione, nessun carrello, nessun acquisto.',
    ...kihinweise,
    'anbieter.github.kennung':
      'Il `client_id` dell’applicazione OAuth che hai creato su GitHub. Vuoto significa: l’accesso con GitHub non viene offerto.',
    'anbieter.google.kennung':
      'Il `client_id` dell’applicazione OAuth che hai creato su Google. Di solito termina con `apps.googleusercontent.com`.',
    'anbieter.microsoft.kennung':
      'L’ID applicazione dal portale Microsoft Entra — un UUID. Microsoft **non** richiede alcun segreto per questo flusso.',
    'anbieter.x.kennung':
      'Il `client_id` dell’applicazione che hai creato nel portale sviluppatori di X.',
  },

  werte: {
    ...schalterwerte,
    ...protokollwerte,
    ...handelswerte,
    ...gestaltungswerte,
    ...ansichtswerte,
    // **Die lokale KI-Laufzeit** — A7, #1446.
    // **Die Kontoschau** — T11, #1476. Dieselben Worte wie bei der
    // KI-Laufzeit darüber: zwei Schreibweisen für denselben Zustand
    // lernt niemand gern.
    'konto.schau.aus': 'Disattivato',
    'konto.schau.an': 'Attivato',
    ...kiwerte,
    'sprache.erben': 'Come per l’utente',
    'erscheinungsbild.erben': 'Come per l’utente',
    'schriftgroesse.erben': 'Come per l’utente',
    'schriftart.erben': 'Come per l’utente',
    'sprache.de': 'Deutsch',
    'sprache.en': 'English',
    'sprache.es': 'Español',
    'sprache.fr': 'Français',
    'sprache.it': 'Italiano',
    'sprache.ja': '日本語',
    'sprache.ru': 'Русский',
    'sprache.zh': '简体中文',
    'erscheinungsbild.system': 'Come il sistema',
    'erscheinungsbild.hell': 'Chiaro',
    'erscheinungsbild.dunkel': 'Scuro',
    'schriftgroesse.klein': 'Piccola',
    'schriftgroesse.normal': 'Normale',
    'schriftgroesse.gross': 'Grande',
    'schriftgroesse.sehrgross': 'Molto grande',
    'aktualisierung.aus': 'Disattivato',
    'aktualisierung.melden': 'Solo notifica',
    'aktualisierung.holen': 'Notifica e scarica',
    'ersteschritte.an': 'Attivato',
    'ersteschritte.aus': 'Disattivato',
    'seitenleiste.sichtbar': 'Visibile',
    'seitenleiste.verborgen': 'Nascosta',
    'leistenlage.senkrecht': 'A sinistra',
    'leistenlage.waagerecht': 'In alto',
    'fensterstart.maximiert': 'Riempi lo schermo',
    'fensterstart.normal': 'Dimensione normale',
    'fensterstart.zuletzt': 'Come l’ultima volta',

    'feldtiefe.normal': 'Normale',
    'feldtiefe.erweitert': 'Avanzato',
    'feldtiefe.experte': 'Esperto',
    'tisch.autoanordnung.an': 'Sì — il tavolo le dispone',
    'tisch.autoanordnung.aus': 'No — ciò che posi resta dov’è',
    'tisch.beschriftung': 'Etichette sul tavolo',
    'tisch.beschriftung.aus': 'Disattivato — solo i conteggi',
    'tisch.beschriftung.klein': 'Piccole',
    'tisch.beschriftung.normal': 'Medie',
    'tisch.beschriftung.gross': 'Grandi',
    'schriftart.pergament': 'Pergamena — calmo e leggibile',
    'schriftart.fantasy': 'Fantasy — giocoso',
    'schriftart.antike': 'Antichità — capitali romane',
    'schriftart.kalligrafie': 'Calligrafia — scritto a mano',
    'barrierefreiheit.bewegung.system': 'Come il sistema',
    'barrierefreiheit.bewegung.reduziert': 'Sempre poco',
    'barrierefreiheit.kontrast.system': 'Come il sistema',
    'barrierefreiheit.kontrast.hoch': 'Sempre alto',
    'sicherheit.faktorbeimwechsel.verlangen': 'Anche al cambio',
    'sicherheit.faktorbeimwechsel.nurAnmelden': 'Solo all’accesso',
    'vorgabe.zustand.ohne': 'Nessun valore predefinito',
    'vorgabe.zustand.mint': 'Mint',
    'vorgabe.zustand.nearMint': 'Near Mint',
    'vorgabe.zustand.excellent': 'Excellent',
    'vorgabe.zustand.good': 'Good',
    'vorgabe.zustand.lightPlayed': 'Light Played',
    'vorgabe.zustand.played': 'Played',
    'vorgabe.zustand.poor': 'Poor',
    'vorgabe.einzelspiel.aus': 'Disattivato',
    'vorgabe.einzelspiel.an': 'Attivato',

    // **The workspace and the limit** — Werkstück O3 of Phase 5.2 (#1427).
    'arbeitsplatz.navigator.erben': 'Come per l’utente',
    'arbeitsplatz.navigatorbreite.erben': 'Come per l’utente',
    'arbeitsplatz.inspektorbreite.erben': 'Come per l’utente',
    'arbeitsplatz.navigator.sichtbar': 'Visibile',
    'arbeitsplatz.navigator.verborgen': 'Nascosta',
    'arbeitsplatz.inspektor.erben': 'Come per l’utente',
    'arbeitsplatz.inspektor.sichtbar': 'Visibile',
    'arbeitsplatz.inspektor.verborgen': 'Nascosta',
    'richtlinie.nurlokal.aus': 'Disattivato',
    'richtlinie.nurlokal.an': 'Attivato',
  },

  // **Was dem Menschen gehört** — Werkstück 8 von Phase 4.12 (#1351).
  // **Where a login comes from in the first place** — follow-up to Werkstück 4
  // of Phase 4.12 (#1351).
  anmeldung: 'Accesso tramite un provider',
  anmeldungHinweis:
    'Questa applicazione **non include credenziali**. Per accedere con GitHub, Google, Microsoft o X, crea tu stesso presso il provider un’**applicazione OAuth** e inserisci qui il suo `client_id`. Come URI di reindirizzamento indica l’indirizzo mostrato nel browser durante l’accesso — punta a questo computer ed è valido solo finché l’accesso è in corso.',

  anmeldungGeheimnis:
    'GitHub richiede inoltre un `client_secret` — Google lo lascia facoltativo, Microsoft lo vieta per programmi come questo. Va **nel portachiavi**, non qui: l’archivio viene condiviso, il portachiavi no. Crealo in **Accesso Portachiavi** come **password di applicazione**, con il nome `MagicSetEditor4Mac-Anmeldung` e l’account `github`. Da riga di comando: `security add-generic-password -s MagicSetEditor4Mac-Anmeldung -a github -w`.',

  geheimnis: 'Segreto',
  geheimnisPflicht: 'è richiesto e manca',
  geheimnisVerboten:
    'non richiede alcun segreto e ne rifiuta uno — qui basta l’ID client.',

  anbietername: {
    github: 'GitHub',
    google: 'Google',
    microsoft: 'Microsoft',
    x: 'X',
  },

  wirktNachNeustart: 'Ha effetto dopo un riavvio.',

  trefferInGruppe: '{zahl} risultati per “{frage}”',
  gesuchtMarke: 'trovato',
  sucheSprung: 'Vai al primo risultato',

  vorgaben: 'Valori predefiniti',

  // **The pointer to the help** — Werkstück 13 of Phase 4.12 (#1351).

  dienste: 'Servizi di carte di terze parti',
  diensteHinweis:
    'Disattivato significa: nemmeno una richiesta va a questo servizio. L’applicazione continua a funzionare senza alcun servizio — manca solo ciò che arriva da fuori.',

  kiAnbieter: 'Modelli linguistici',
  kiAnbieterHinweis:
    'Dove si può fare una domanda. Ogni provider calcola sui propri computer e vede ciò che invii — perciò ciascuno va attivato separatamente, e nessuno lo è da sé.',
  // See the German catalogue for why these two live here and not under
  // `werte`: they are the same for every service.
  dienstAus: 'Disattivato',
  dienstAn: 'Attivato',
  dienstUngeprueft: 'Mai eseguito con il servizio reale',
  dienstUngeprueftHinweis:
    'Per questo servizio qui non c’è alcun account. Ciò che è testato è il nostro lato del confine — la richiesta, le intestazioni, la corrispondenza dei dati. Se il servizio risponde come supponiamo, non è stato misurato.',
  fragen: 'Chi risponde a cosa',
  fragenHinweis:
    'Quando più servizi possono rispondere alla stessa domanda, qui si decide quale viene interpellato. **Non c’è alcun ripiego silenzioso**: se quello scelto tace, te lo diciamo — invece di mostrarti in silenzio una risposta da un’altra fonte.',
  'frage.suche': 'Ricerca di carte',

  // The service catalogue — work item A1 of phase 5.3 (#1440).
  ...dienstkatalog,

  dienstSchluesselFehlt: 'Nessuna credenziale memorizzata',
  dienstSchluesselHinweis:
    'Questo servizio richiede credenziali. Si trovano nel portachiavi di macOS, non nel file della collezione — quello viene condiviso.',
  zugaenge: 'Credenziali',
  zugaengeHinweis:
    'Si trovano nel portachiavi di macOS e **non** nel file della collezione — quello viene condiviso. Ciò che vedi qui è solo **se** qualcosa è memorizzato; il valore stesso non torna mai indietro.',
  // See the German catalogue: these name the headers each service asks for
  // (documented in decision 0056).

  kuerzelUeberschrift: 'Scorciatoie da tastiera',
  ...einstellungsbelegung,
  kuerzelHinweis:
    'Valgono in tutta la finestra. In un campo di testo, ⌘Z e ⌘F appartengono al campo.',
  kuerzelZweiteReihe:
    'Le otto scorciatoie con ⌥ si trovano sui tasti della riga superiore, da sinistra. ' +
    'Su una tastiera con un’altra disposizione — ad esempio AZERTY — lì è stampato un ' +
    'altro carattere; il tasto è lo stesso.',
  kuerzelTaste: 'Tasto',
  kuerzelHandlung: 'Azione',

  austausch,
  eigenuebersetzung,
  // **An account's remaining quota at a provider** — A5, #1444.
  kontingentTitel: 'Quota presso il provider',
  kontingentErklaerung:
    'Quanti crediti include il piano e quanti ne restano. La richiesta esce verso la rete e costa essa stessa un credito.',
  kontingentHolen: 'Controlla la quota',
  kontingentHolenBei: 'Controlla la quota presso {dienst}',
  kontingentLaeuft: 'Controllo in corso …',
  kontingentZahlen: 'Restano {rest} crediti su {gesamt}.',
  kontingentFrist:
    'Il provider può indicare questo valore con un ritardo fino a {minuten} minuti — non è un blocco. Ciò che qui limita davvero la spesa è il budget di questo archivio.',
  kontingentUeberzug: 'Un credito oltre il piano costa {satz}.',
  // **What is known about AI models** — A6, #1445.
  ...ki,

  // **Die Zugangsdaten** (#1534) — eigene Datei, ein Katalog (§9).
  ...zugaenge,

  ...netzbereich,
}
