// **The network settings** — work item N5 of phase 5.4
// (#1571).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge — dieselbe Gestalt wie `einstellungenzugaenge.ts`
// daneben. **Ein Katalog bleibt es trotzdem**: die drei Teile werden dort
// hineingestreut, und der Vertragstest gegen die Go-Seite sieht einen.

// netznamen sind die Beschriftungen der sechs Einstellungen.
export const netznamen = {
  'netz.proxy.http': 'Proxy per http',
  'netz.proxy.https': 'Proxy per https',
  'netz.proxy.aus': 'Raggiungere senza proxy',
  'netz.frist.verbindung': 'Timeout: connessione (s)',
  'netz.frist.kopf': 'Timeout: intestazione della risposta (s)',
  'netz.frist.stillstand': 'Timeout: nessun avanzamento (s)',
  'netz.schnittstelle': 'Interfaccia per la ricerca nella rete locale',
  'netz.lauschen': 'Dove ascoltano i servizi propri',
  'dienst.tls.zertifikat': 'Certificato proprio dei servizi',
  'dienst.tls.schluessel': 'Chiave del certificato proprio',
  'dienst.tls.kette': 'Certificati intermedi',
  'dienst.tls.anker': 'Ancore di fiducia del certificato',
  'grenzen.kontingente': 'Quote dei servizi',
}

// netzhilfe ist der erklärende Satz je Einstellung.
export const netzhilfe = {
  'netz.proxy.http':
    'Il proxy attraverso cui passano gli indirizzi `http://`. **Vuoto significa che vale `HTTP_PROXY` dell’ambiente** — quello che inserisci qui lo sostituisce. Con lo schema, ad esempio `http://proxy.azienda:3128`. **Senza password** — sarebbe salvata in chiaro sul disco; un proxy con autenticazione arriva da `HTTPS_PROXY` nell’ambiente.',
  'netz.proxy.https':
    'Lo stesso per gli indirizzi `https://`, da `HTTPS_PROXY`. **Separato**, perché in molte reti c’è un proxy solo per uno dei due.',
  'netz.proxy.aus':
    'Ciò che **non** passa dal proxy — nomi di host e intervalli di indirizzi, separati da virgole, come `NO_PROXY`. L’indirizzo di loopback resta comunque escluso; un Ollama su un **altro** computer della tua rete va inserito qui.',
  'netz.frist.verbindung':
    'Quanto può durare la connessione, TLS incluso. Vuoto significa 15 secondi.',
  'netz.frist.kopf':
    'Quanto tempo ha un server per inviare l’intestazione della risposta. Vuoto significa 30 secondi.',
  'netz.frist.stillstand':
    'Quanto può restare ferma una risposta in corso. **È questo il valore da aumentare** quando un download grande su una linea lenta si interrompe — non i due precedenti. Vuoto significa 60 secondi.',
  'netz.schnittstelle':
    'Da quale interfaccia di rete parte la ricerca nella rete locale e attraverso quale esce l’annuncio — ad esempio `en0`. **Vuoto significa automatico**: ogni interfaccia idonea con IPv4 e multicast. Un nome che al momento non esiste vale come vuoto e compare come segnalazione nel registro.',
  'netz.lauschen':
    'Indirizzi su cui ascoltano REST e MCP, separati da virgole. **Vuoto significa: solo questo Mac** (`127.0.0.1`). `0.0.0.0` significa ogni interfaccia e sta da solo. **Se un indirizzo esce da questo computer, valgono TLS e token**; è più comodo sceglierlo in «Strumenti senza finestra». Ha effetto al prossimo avvio di un servizio.',
  'dienst.tls.zertifikat':
    'Il certificato che i servizi REST e MCP presentano oltre il loopback, in formato PEM. **Vuoto significa: il certificato sull’identità di questo profilo.** Con un’autorità di certificazione tua, i client verificano nome e catena invece dell’impronta. Vale dal prossimo avvio del servizio.',
  'dienst.tls.schluessel':
    'La chiave privata del certificato, in formato PEM. **Leggibile solo da te (0600)**, altrimenti il servizio non si avvia. La sostituzione di entrambi i file vale senza riavvio al prossimo handshake.',
  'dienst.tls.kette':
    'Certificati intermedi, se non sono inclusi nel certificato. Vuoto significa: nessuno.',
  'dienst.tls.anker':
    'La radice della tua autorità, rispetto alla quale la catena viene verificata prima dell’avvio. Vuoto significa: le ancore del sistema.',
  'grenzen.kontingente':
    'Quanto può fare al minuto un accesso macchina tramite REST e MCP — per accesso e per tutti insieme. **Disattivate è per un ambiente controllato** in cui nessuno toglie nulla a nessuno, come una prova di carico. L’accesso, TLS, i posti per tipo di lavoro, i limiti del corpo e della risposta, le scadenze e il blocco contro chi tira a indovinare restano. Ha effetto al prossimo avvio di un servizio.',
}

// netzbereich ist die Überschrift des Bereichs samt Hinweis.
export const netzbereich = {
  netz: 'Come questo programma si connette all’esterno',
  netzHinweis:
    'Queste impostazioni valgono per **ogni** servizio contemporaneamente — anche per la riga di comando, il server MCP e il servizio. Vuoto significa che vale ciò che indica l’ambiente. Ciò che vale al momento, e da dove proviene, è mostrato in Rete e raggiungibilità.',
}

// netzwerte sind die Namen der Werte — QA-08.AK-03 (#1739).
export const netzwerte = {
  'grenzen.kontingente.an': 'Attivate',
  'grenzen.kontingente.aus': 'Disattivate',
}
