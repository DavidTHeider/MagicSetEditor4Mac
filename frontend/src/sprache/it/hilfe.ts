import { betrieb } from './hilfebetrieb'
import { suche } from './hilfesuche'

export const hilfe = {
  betrieb,
  suche,
  // One help entry per page — the keys come from `navigation.ts`.
  seiten: {
    start: {
      ueberschrift: 'Home',
      text: 'Il punto di partenza. In alto i numeri della libreria — set, carte, pacchetti, esemplari — e sotto ciò su cui hai lavorato per ultimo: aperto, importato, esportato, e ciò che tieni fissato in alto. Poi ciò che richiede attenzione — set con riserve, mazzi con carte mancanti, pacchetti con segnalazioni negli script — e ciò che viene scaricato in background. I primi tre passi compaiono come finestra al primo avvio; i tour nell’aiuto li mostrano di nuovo.',
    },
    sets: {
      ueberschrift: 'Set',
      text: 'Tutti i set che questa applicazione conosce. Da qui se ne apre uno, se ne crea uno nuovo o si importa un file `.mse-set`. Il set aperto determina cosa mostrano le sue schede e i cassetti «Proprietà» e «Strumenti».',
    },
    pakete: {
      ueberschrift: 'Pacchetti',
      text: 'Giochi, fogli di stile, font di simboli e modelli di esportazione. Vengono scaricati qui — a una revisione precisa e con checksum — e installati. Senza di essi una carta non ha campi né aspetto.',
    },
    vorlagen: {
      ueberschrift: 'Modelli',
      text: 'I file di testo all’interno di un pacchetto di modelli — campi, script, misure. Qui si possono modificare senza un editor di testo a parte. Una modifica vale per ogni set che usa questo pacchetto; lo stato precedente resta come backup accanto al pacchetto.',
    },
    sammlung: {
      ueberschrift: 'Collezione',
      text: 'Ciò che possiedi davvero — una riga per ogni stampa, condizione, lingua e finitura. Non ha nulla a che fare con i tuoi set: una carta progettata è una bozza, un esemplare è un pezzo di cartoncino. I prezzi vengono scaricati su richiesta e sono stime di cui nessuno risponde.',
    },
    decks: {
      ueberschrift: 'Mazzi',
      text: 'Composizioni di carte tue e stampe reali. Un mazzo rimanda alle carte invece di copiarle e appartiene a un gioco — senza gioco non esisterebbe alcun formato. Qui un formato è un record e non un elenco integrato: si crea, non viene fornito. Il controllo interroga Scryfall per le stampe reali; una carta che hai progettato tu resta non verificata, e questo non significa non legale.',
    },
    turniere: {
      ueberschrift: 'Tornei',
      text: 'Un torneo è uno svolgimento nel tempo: turni, abbinamenti, risultati. La classifica viene calcolata da essi e non è salvata da nessuna parte — se correggi un risultato, la classifica si aggiorna subito. L’abbinamento segue il sistema svizzero; i quattro spareggi e le loro cifre provengono dalle Magic Tournament Rules. Iscrivere un mazzo significa consegnare una lista: viene copiata, e le modifiche successive al mazzo non la raggiungono più.',
    },
    spielen: {
      ueberschrift: 'Gioca',
      text: 'Una partita è un tavolo con mazzi congelati: ciò che un mazzo diventa all’inizio della partita non cambia più, anche se qualcuno modifica quel mazzo in seguito. Ciò che accade al tavolo è un flusso di eventi — ognuno si aggancia al precedente e nessuno viene sovrascritto. Ciò che vedi è sempre la vista di **un** solo posto: un grimorio è coperto anche per il suo proprietario, e chi lo consulta deve poi rimescolarlo. L’host conserva lo stato completo; la pagina lo dice, perché è una cosa da sapere prima di unirsi.',
    },
    handeln: {
      ueberschrift: 'Scambi',
      text: 'Si scambia con una controparte di cui hai confrontato l’impronta — senza account e senza intermediari. Il confronto viene calcolato su **questo** computer: non ne esce nulla, e ciò che una controparte vede di te lo decidi tu in «Impostazioni › Connessioni». La registrazione avviene solo quando le carte hanno davvero cambiato proprietario; la tua collezione cambia allora, e il registro annota cosa è andato dove. **Registrato qui non significa registrato da entrambe le parti** — due collezioni su due computer non sono una registrazione condivisa, e lo stato lo dice. Quanto vale la tua collezione compare accanto come osservazione, non come promessa.',
    },
    hilfe: {
      ueberschrift: 'Aiuto',
      text: 'I termini che provengono da Magic Set Editor e che nessuno conosce già, ciascuno in una frase. Inoltre i tour, le scorciatoie da tastiera, perché alcune etichette restano in inglese e cosa questa applicazione invia all’esterno.',
    },
    nutzer: {
      ueberschrift: 'Utenti e profili',
      text: 'Qui si tiene traccia di chi possiede cosa. Un utente è la persona; un profilo è un ambiente di lavoro di quella persona — propri set, propria collezione, proprie impostazioni. Chi vuole collezionare separatamente per un torneo ha bisogno di un secondo profilo, non di un secondo account.',
    },
    einstellungen: {
      ueberschrift: 'Impostazioni',
      text: 'Lingua, aspetto e i luoghi in cui vengono cercati i pacchetti. Ciò che è qui vale per l’intero programma, non per un singolo set.',
    },
    karten: {
      ueberschrift: 'Carte',
      text: 'Le carte del set aperto: a sinistra l’elenco, al centro l’anteprima, a destra i campi della carta aperta. Una modifica a un campo si vede subito; non occorre confermare nulla.',
    },
    setangaben: {
      ueberschrift: 'Dati del set',
      text: 'Ciò che vale per l’intero set: nome, descrizione, simbolo del set, autore. Questi dati sono a disposizione di ogni carta — un foglio di stile può disegnare il simbolo su ciascuna.',
    },
    stil: {
      ueberschrift: 'Stile',
      text: 'Il foglio di stile del set e le sue opzioni. Il foglio dice come appare una carta; le opzioni lo regolano — per esempio su una cornice diversa o su un’altra stampa.',
    },
    suche: {
      ueberschrift: 'Cerca',
      text: 'Trova le carte tramite i valori dei loro campi, in tutti i set o in uno solo. La ricerca avviene sul testo completo, e un risultato porta direttamente alla carta.',
    },
    ersetzen: {
      ueberschrift: 'Sostituisci',
      text: 'Sostituisci una parola nei campi di testo di un set. Prima mostra cosa cambierebbe; un solo ⌘Z annulla l’intera operazione.',
    },
    statistik: {
      ueberschrift: 'Statistiche',
      text: 'Cosa c’è nel set aperto: quante carte, quanto spesso un campo è compilato, come si distribuisce un valore di campo. Per verificare se una serie è completa.',
    },
    bilder: {
      ueberschrift: 'Immagini',
      text: 'Scrive le carte del set come file immagine — tutte insieme, con uno schema di denominazione e una regola per i nomi duplicati. **L’assegnazione** dell’immagine di una carta non avviene qui, ma nel campo immagine della carta.',
    },
    druckbogen: {
      ueberschrift: 'Foglio di stampa',
      text: 'Le carte affiancate su carta, come PDF con segni di taglio. Pensato per la stampante, non per essere condiviso — a quello serve l’esportazione.',
    },
    schluesselwoerter: {
      ueberschrift: 'Parole chiave',
      text: 'Una parola chiave riconosce il proprio testo su ogni carta e aggiunge il testo di promemoria. Quelle del gioco sono fisse e sono mostrate qui per consultazione; le tue valgono solo in questo set e si possono aggiungere, modificare ed eliminare.',
    },
    pack: {
      ueberschrift: 'Booster casuale',
      text: 'Un booster con le carte di questo set, secondo le regole del pacchetto di gioco — quali tipi di booster esistono lo dice quello. Lo stesso seme dà la stessa estrazione; per mostrare un risultato, passa il numero.',
    },
    symbol: {
      ueberschrift: 'Simbolo',
      text: 'Il simbolo di questo set — disegnato, non caricato. Si parte da una forma di base; poi ogni punto si può spostare. Tab passa da un punto all’altro, i tasti freccia li spostano. Ciò che disegni qui compare subito nell’anteprima e finisce nel file quando salvi il set.',
    },
    konsole: {
      ueberschrift: 'Console',
      text: 'Un’espressione MSE-Script nel contesto della carta aperta — lo stesso ambiente del disegno. Non modifica nulla; calcola e restituisce. Il risultato arriva come testo sorgente, così una stringa si distingue da un nome.',
    },
    ausgeben: {
      ueberschrift: 'Esporta',
      text: 'Carte come immagine, oppure come testo tramite un modello di esportazione. Il modello proviene da un pacchetto e decide cosa esce: un elenco per un forum, una tabella, un file per un altro programma.',
    },
  },

  // **Über** (#1168).
  ueber: {
    ueberschrift: 'Informazioni su MagicSetEditor4Mac',
    fassung: 'Versione',
    lizenzName: 'Licenza',
    lizenzSatz:
      'Il codice sorgente di questa applicazione è sotto licenza GNU General Public License, versione 2 o successiva.',
    rechteName: 'Contenuti di terze parti',
    rechteSatz:
      'La licenza copre il codice sorgente, **non** le cornici delle carte, i simboli di mana, i font, i testi delle regole, i nomi e i marchi. Questi appartengono ai rispettivi titolari dei diritti — tra cui Wizards of the Coast — e non a questo progetto.',
    quelleName: 'Codice sorgente',
    quelleSatz: 'github.com/DavidTHeider/MagicSetEditor4Mac',
  },

  weiterlesen: {
    ueberschrift: 'Per approfondire',
    einleitung:
      'Due documenti fanno parte dei sorgenti di questa applicazione — nella cartella del progetto, in `docs/`, non nell’immagine disco che hai installato. Questa applicazione non può aprirli da sé: non ha una via verso l’esterno, e un collegamento che non porta da nessuna parte sarebbe peggio di nessuno.',
    handbuchName: 'Manuale',
    handbuchSatz:
      'Ciò che qui compare per parole chiave, in frasi complete: il primo avvio, ' +
      'come costruire una carta, esportare, le scorciatoie — e cosa fare quando una carta ' +
      'resta vuota. `docs/HANDBUCH.md`',
    administrationName: 'Dati e backup',
    administrationSatz:
      'Dove si trovano i dati, come fare il backup e ripristinarlo, cosa contiene il ' +
      'registro e cosa succede al cambio di versione — compresa la frase secondo cui ' +
      'un backup **senza password** è testo in chiaro. `docs/ADMINISTRATION.md`',
  },

  erststart: {
    ueberschrift: 'Al primo avvio',
    text:
      'Questa applicazione non è notarizzata. macOS la blocca una volta; ' +
      'la autorizzi in Impostazioni di Sistema, sotto «Privacy e sicurezza».',
  },

  // Tours, and the areas that have no page (#1422).
  bereiche: {
    ueberschrift: 'E un’area che non ha una pagina nella barra laterale',
    einleitung:
      'Non tutto ciò che questa applicazione sa fare si trova nella barra laterale. Ciò che arriva da fuori non è un luogo — e quindi non ha una pagina.',
    kiName: 'IA',
    kiSatz:
      'Un modello linguistico legge e modifica le carte tramite il **server MCP** — un processo separato sulla stessa libreria, che registri nel tuo strumento di IA. Cosa esce così dal computer e cosa no è spiegato in **Interfacce**.',
    kiKnopf: 'Vai alle interfacce',
  },
  fuehrungen: {
    ueberschrift: 'Tour',
    einleitung:
      'Un tour ti guida passo dopo passo attraverso l’applicazione e si può ripetere in qualsiasi momento. Accanto, una frase ciascuno su come è organizzata l’applicazione.',
    startName: 'I primi passi',
    startSatz:
      'Tre passi fino alla prima carta: aprire un set, modificare le carte, scaricare i modelli. Compaiono da soli al primo avvio; se tornano al prossimo avvio è un’impostazione.',
    startZeigen: 'Mostra ora i primi passi',
    arbeitsplatzName: 'L’area di lavoro',
    arbeitsplatzSatz:
      'Carte, modelli, collezione e mazzi condividono un unico schema: navigatore a sinistra, ispettore a destra, il centro in mezzo. **⌥⌘[** comprime il navigatore, **⌥⌘]** l’ispettore — e ciò che comprimi resta compresso per profilo. **⌥⌘S** comprime ed espande la barra laterale, così la sua larghezza va alle aree di lavoro — questo stato appartiene a questo dispositivo ed è salvato per profilo.',
    paletteName: 'La palette comandi',
    paletteSatz:
      '**⌘K** la apre sopra ogni pagina: pagine, set e comandi, trovati con poche lettere.',
    leisteName: 'La barra del contesto',
    leisteSatz:
      'Sotto il nome dell’applicazione dice chi lavora dove e con quale ruolo — persona, profilo, ruolo. Si cambia nella pagina **Utenti e profili**; la barra porta lì.',
  },

  begriffe: {
    set: {
      name: 'Set',
      satz: 'Una raccolta di carte insieme ai suoi dati — nome, descrizione, simbolo. Si trova in un file con estensione `.mse-set`.',
    },
    karte: {
      name: 'Carta',
      satz: 'L’oggetto di cui si tratta. Una carta contiene valori di campo — nome, costo, testo delle regole — e viene disegnata secondo un foglio di stile.',
    },
    spiel: {
      name: 'Gioco',
      satz: 'Dice **quali campi** ha una carta. Un `.mse-game` li dichiara: `magic` conosce costo di mana e rarità, un altro gioco ne conosce altri.',
    },
    stil: {
      name: 'Foglio di stile',
      satz: 'Dice **come** appare una carta: quale livello sta dove, quale immagine dà la cornice. Un `.mse-style` appartiene sempre a un gioco.',
    },
    stiloption: {
      name: 'Opzione di stile',
      satz: 'Un’impostazione del foglio di stile che vale per l’intero set o per una carta — per esempio la cornice di una carta draft o un’altra stampa.',
    },
    symbolschrift: {
      name: 'Font di simboli',
      satz: 'Disegna simboli di mana e simili. Non è un carattere tipografico ma una raccolta di piccole immagini con regole su quale testo mostra quale.',
    },
    paket: {
      name: 'Pacchetto',
      satz: 'Il termine generico per gioco, foglio di stile, font di simboli e modello di esportazione. I pacchetti vengono scaricati e installati; senza di essi una carta resta vuota.',
    },
    exportvorlage: {
      name: 'Modello di esportazione',
      satz: 'Un piccolo programma che trasforma le carte di un set in **testo** — un elenco per un forum, una tabella, un file per un altro programma.',
    },
    druckbogen: {
      name: 'Foglio di stampa',
      satz: 'Un PDF con le carte affiancate su carta, con segni di taglio. Per stampare e ritagliare — non per essere condiviso come file.',
    },
    vorbehalt: {
      name: 'Riga non letta',
      satz: 'Una riga di un file aperto che questo programma non ha interpretato. Non va persa: viaggia invariata e viene riscritta al salvataggio.',
    },
    befund: {
      name: 'Segnalazione',
      satz: 'Un avviso su qualcosa che non è andato a buon fine durante il disegno — un livello mancante, un carattere senza simbolo. La carta viene disegnata comunque.',
    },
  },

  sprache: {
    ueberschrift: 'Perché alcuni testi restano in inglese',
    einleitung:
      'L’etichetta di un campo della carta e la frase sotto di essa provengono dal **pacchetto di gioco**, non da questa applicazione. Compaiono come li ha scritti l’autore del pacchetto — di solito in inglese.',
    schluesselName: 'Il nome è la chiave',
    schluesselSatz:
      'Sotto di esso il tuo file `.mse-set` memorizza il valore. Una traduzione inventata sarebbe più bella e renderebbe il file illeggibile per chi lo consulta — per questo questa applicazione non ne inventa.',
    uebersetztName: 'Ciò che è tradotto viene tradotto',
    uebersetztSatz:
      'Se un pacchetto porta una traduzione, viene usata. Nel `M15-Magic-Pack` scaricato erano 25 campi della carta al 2026-08-21. Per gli altri 224 non ne portava, e allora compare la stessa parola di Magic Set Editor 2.',
    fensterName: 'La finestra stessa parla otto lingue',
    fensterSatz:
      'Tutto ciò che questa applicazione dice da sé — pulsanti, messaggi, questa pagina — segue la lingua delle impostazioni. Il tedesco è la lingua di partenza; le altre sette sono tradotte automaticamente e non ancora revisionate — lo dice anche la scelta della lingua.',
    leisteName: 'Anche la barra dei menu',
    leisteSatz:
      'Il menu dell’applicazione e i menu di modifica e della finestra usano le parole che macOS usa in questa lingua. **Taglia**, **Copia** e **Incolla** restano collegati al sistema operativo.',
  },

  // **What this application knows about you** — Werkstück 13 of Phase 4.12
  // (#1351).
  datenschutz: {
    ueberschrift: 'Cosa sa di te questa applicazione',
    einleitung:
      'Da quando esistono utenti e profili, questa applicazione registra **dati personali**. Qui è spiegato quali sono, dove si trovano, cosa di essi lascia questo computer — e cosa espressamente no.',

    wasName: 'Cosa viene registrato',
    wasSatz:
      'Tutto è **facoltativo** tranne il nome visualizzato. Sulla persona: **appellativo**, **suffisso del nome**, **nome**, **altri nomi**, **cognome**, **slogan** e due **biografie**, una breve e una lunga. Sull’account: **indirizzo**, **secondo indirizzo**, **numero di telefono** e, se ne imposti una, una **password**. Inoltre una **immagine del profilo** e un **badge** per profilo.',

    woName: 'Dove si trova',
    woSatz:
      'Nella tua libreria — un file SQLite nella tua cartella utente — e nelle immagini accanto. **Tre cose si trovano altrove**, cioè nel **portachiavi** di macOS: il segreto del tuo secondo fattore, il segreto di un accesso presso un provider e le credenziali dei servizi di carte di terze parti. La libreria si condivide, il portachiavi no — da qui la separazione.',

    formName: 'In quale forma',
    formSatz:
      'Una **password** non viene mai salvata in chiaro: è protetta con **Argon2id**, e la forma protetta porta con sé i propri parametri. Un **token di accesso** è salvato come checksum; il testo in chiaro esiste una sola volta, subito dopo la creazione. Tutto il resto è salvato come lo hai digitato — è il tuo file. **E il file stesso si può cifrare** — in **Impostazioni › Dati e backup › Database**. Allora nulla è più leggibile sul disco; la chiave va nel portachiavi di questo Mac e **non è recuperabile**. Un **backup** ottiene una propria **password** — allora è cifrato e si apre comunque su qualsiasi computer; senza password un database cifrato lo consegna solo su richiesta esplicita.',

    ausName: 'Cosa lascia questo computer',
    ausSatz:
      'Con un **accesso**, la tua scelta va al provider che hai scelto: `github.com` e `api.github.com`, `accounts.google.com` e `oauth2.googleapis.com`, `login.microsoftonline.com`, oppure `twitter.com` e `api.twitter.com`. Lì vengono richiesti solo il tuo **nome e il tuo id** — **non** il tuo indirizzo e-mail, non i tuoi contatti, non i tuoi post. Tornano indietro e vengono salvate tre cose: il provider, il tuo id presso di lui e il tuo nome visualizzato lì.',

    nichtName: 'Cosa non esce',
    nichtSatz:
      'Le tue **parti del nome, biografie, numero di telefono e indirizzo** non lasciano mai questo computer. Si trovano nella tua libreria e non vengono inviati da nessuna parte. Non c’è **nessuna telemetria**, nessun rapporto sugli arresti anomali, nessun conteggio d’uso e nessun account presso di noi — non esiste alcun «noi».',

    dienstName: 'E i servizi di carte',
    dienstSatz:
      'La ricerca di carte interroga `api.scryfall.com` e — se abilitati — `api.magicthegathering.io`, `graphql.mtgjson.com` o `api.scrydex.com`. Esce solo la **richiesta di ricerca**: nessun nome utente, nessun id, nulla del tuo profilo. **Le miniature dei risultati** le scarica la finestra stessa — da `cards.scryfall.io` e, con magicthegathering.io, da `gatherer.wizards.com` e `gatherer-static.wizards.com`; lì va solo l’indirizzo dell’immagine. I prezzi per un’intera collezione provengono da `mtgjson.com` se abilitato, le offerte del mercato da `api.cardtrader.com`; esce l’id della stampa e, per CardTrader, il tuo token di accesso — nient’altro. I pacchetti di modelli arrivano tramite `codeload.github.com`, le notifiche di versione tramite `api.github.com`. Cosa può fare qui ciascun servizio e a quali condizioni è indicato in **Impostazioni › Servizi › Servizi di carte di terze parti**; i collegamenti lì portano a `scryfall.com`, `docs.magicthegathering.io`, `mtgjson.com`, `scrydex.com` e `www.cardtrader.com` — l’applicazione stessa non chiama mai quelle pagine, si aprono nel tuo browser solo quando fai clic. **E i collegamenti di mercato su una carta** (T8): portano a `www.cardmarket.com`, `www.cardtrader.com`, `www.tcgplayer.com`, `www.cardkingdom.com`, `gatherer.wizards.com` e `www.mtggoldfish.com` — **costruiti, non chiamati**: l’indirizzo viene composto su questo computer dal nome della carta o da un identificativo del provider, ed esce solo quando fai clic, nel browser che preferisci. Non serve alcuna chiave, e nessuna è contenuta in questi indirizzi.',

    kiName: 'E l’IA',
    kiSatz:
      'Un modello **locale** gira su questo computer e non lo lascia: l’applicazione parla con `127.0.0.1` e con nient’altro. Un **modello cloud** è un’altra cosa — se attivi Anthropic, OpenAI o xAI, `api.anthropic.com`, `api.openai.com` o `api.x.ai` ricevono l’**incarico e i campi della carta** in questione, insieme alla tua chiave dal portachiavi. **Con OpenAI non resta nulla lì**: l’interfaccia salva una risposta per impostazione predefinita, e questa applicazione lo disattiva esplicitamente. **Lo stesso con xAI**: il percorso che conserva le risposte per 30 giorni non viene mai usato. Nulla del tuo profilo, nessuna seconda carta, nessun set. Entrambi sono **disattivati** finché non li attivi, e in **Impostazioni › Servizi › Lavora solo in locale** ogni modello cloud resta disattivato anche se il suo interruttore dice altro. Ciò che torna indietro viene **mostrato, non adottato** — ciò che finisce in un campo lo conferma una persona. Cosa può fare qui Anthropic e a quali condizioni è indicato in **Impostazioni › Servizi › Servizi di carte di terze parti**; i collegamenti lì puntano a `www.anthropic.com`, `platform.claude.com`, `openai.com`, `developers.openai.com`, `x.ai` e `docs.x.ai` — l’applicazione stessa non chiama mai quelle pagine, si aprono nel tuo browser al tuo clic.',

    auskunftName: 'Accesso e cancellazione',
    auskunftSatz:
      'In **Utenti e profili › Il mio account** esporti i tuoi dati come file — tutto ciò che è elencato qui, in un colpo solo. E in **Utenti e profili** elimini un utente — con i suoi profili, la sua immagine e il segreto del suo secondo fattore. **Ciò che si trova nei suoi profili resta**: set, mazzi, collezione e tornei non appartengono più ad alcun profilo e non sono più raggiungibili dall’interfaccia. Nessuno dei due richiede una rete né il consenso di qualcuno.',
  },

  schnittstellen: {
    ueberschrift: 'Cosa esce e cosa entra',
    einleitung:
      'Questa applicazione funziona offline. **Verso la rete** va solo quando qualcuno lo avvia: scaricare un modello, una ricerca di carte e un accesso richiedono la pressione di un pulsante; cerca nuove versioni solo se un’impostazione lo consente, ed è **disattivata** per impostazione predefinita. Quali indirizzi vengono contattati e cosa esce è descritto in **Cosa sa di te questa applicazione**. Inoltre ci sono due percorsi verso questo computer che anche in questo caso nessuno imbocca da solo — e uno che parte da solo: se l’avvio fallisce, mostra il motivo in una finestra.',
    bezugName: 'Scaricare i modelli',
    bezugSatz:
      'Alla pressione di un pulsante l’applicazione scarica un pacchetto di modelli tramite HTTP, a una revisione precisa e con checksum. Senza la pressione del pulsante non accade nulla.',
    scryfallName: 'Dati delle carte da Scryfall',
    scryfallSatz:
      'Alla pressione di un pulsante l’applicazione chiede a `api.scryfall.com` una carta, un set o un’immagine di carta. Ciò che arriva è **mostrato, non adottato** — ciò che finisce in un campo è una decisione umana. Senza rete rimane l’ultimo stato valido, e dice quanto è vecchio.',
    fassungName: 'Controllare le nuove versioni',
    fassungSatz:
      'L’unico punto che può interrogare senza la pressione di un pulsante — e **solo** se è attivato in **Impostazioni**. L’impostazione predefinita è **disattivato**, e allora non esce nemmeno una richiesta. Se attivato, chiede a `api.github.com` l’ultima release. Il download avviene solo alla pressione di un pulsante, e allora **firma e checksum vengono verificati prima che un file venga collocato**; nulla viene installato da solo.',
    mcpName: 'Il server MCP',
    mcpSatz:
      'Un programma a sé chiamato `mse-mcp`, con cui un modello linguistico può lavorare su un set: leggere carte e campi, impostare campi, creare, copiare, eliminare e riordinare carte, importare un set, salvare un set e disegnare una carta. La misura non è la quantità ma la reversibilità: nessuno strumento sovrascrive ciò che esiste — dove c’è già un file, rifiuta — e ogni modifica dice come annullarla. Avviato da un client come Claude Desktop, parla tramite input e output standard: nessuna porta, nessun indirizzo. **Via HTTP** ascolta solo quando lo avvii in **Impostazioni › Connessioni**, e solo dove lo indica **Dove ascoltano i servizi**.',
    mcpNichtAn:
      'Non gira insieme, ma **è incluso**: nel bundle, in `MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp`, insieme a `mse`, la riga di comando. **Da sola, questa applicazione non mette nulla fuori dal proprio bundle.** In **Impostazioni › Connessioni › Strumenti senza finestra** configuri con un pulsante ciò che vuoi — il collegamento a `mse` in `~/.local/bin` con la voce `PATH` in `~/.zprofile`, la voce per Codex, Claude Desktop o Claude Code — e lì ritiri anche ogni voce; ciò che sta fuori dal proprio blocco non lo tocca. Ciò che gli strumenti cambiano, questa finestra lo nota da sola: chiede lo stato dell’archivio ogni due secondi e ricarica la pagina interessata.',
    dienstName: 'Il servizio per script',
    dienstSatz:
      'Un programma a sé chiamato `mse-rest`, con cui uno script raggiunge le stesse capacità di `mse` — leggere set e carte, impostare un campo, importare o salvare un set, disegnare una carta. Ogni richiesta richiede un token. **Dove ascolta lo decidi tu** in **Dove ascoltano i servizi**: per impostazione predefinita solo questo Mac (`127.0.0.1`); se scegli un indirizzo che esce da questo Mac, richiede TLS e un accesso.',
    dienstNichtAn:
      'Si trova nel bundle, in `MagicSetEditor4Mac.app/Contents/MacOS/mse-rest`, e non si avvia da solo: chi non vuole un servizio non ne ha. Lo si avvia in **Impostazioni › Connessioni** — oppure come LaunchAgent in **Strumenti senza finestra**; allora macOS lo avvia a ogni accesso, finché non lo ritiri lì. Chi lo avvia riceve indirizzo e token. Le richieste da un browser vengono respinte prima ancora di controllare il token.',
    startfensterName: 'Il messaggio quando l’avvio fallisce',
    startfensterSatz:
      'L’unico percorso che l’applicazione imbocca **da sola**. Se il database non si può aprire, richiama `osascript` di macOS per mostrare il motivo in una finestra — altrimenti resterebbe su un canale che dopo un doppio clic nessuno vede. Il testo entra come valore, non come script.',
    dateienName: 'File su questo computer',
    dateienSatz:
      'Set, pacchetti di modelli, immagini scritte e PDF. Stanno dove li metti; la libreria sta nella tua cartella utente.',
    sonstSatz:
      'Nient’altro. **Nessuna telemetria**, nessun rapporto sugli arresti anomali, nessun conteggio d’uso. Un account c’è — ma presso il provider che scegli tu, non presso di noi; ciò che esce in quel caso è descritto in **Cosa sa di te questa applicazione**.',
    grenzenName: 'Limiti',
    grenzenSatz:
      'Ogni accesso automatico ha un **budget**, e ogni chiamata costa in base al suo tipo; quando si esaurisce, la risposta lo dice insieme al tempo di attesa. I lavori pesanti — rendering, importazione, esportazione, operazioni di massa — hanno per classe un numero fisso di posti che finestra, riga di comando, MCP e servizio condividono. Chi tira a indovinare blocca solo la propria credenziale.',
    grenzenZahlen:
      'Oggi: {burst} unità in blocco, {nachfuellen} al minuto; una lettura costa {lesen}, una scrittura {schreiben}, un’importazione {einfuhr}; al massimo {rendern} rendering alla volta; {reserve} unità restano alla finestra.',
    grenzenOhneKontingente:
      '**Le quote sono disattivate** (`grenzen.kontingente`) — per un ambiente controllato. I posti per tipo di lavoro, i limiti del corpo e della risposta, le scadenze, il blocco contro chi tira a indovinare, l’accesso e TLS restano in vigore.',
  },
} as const
