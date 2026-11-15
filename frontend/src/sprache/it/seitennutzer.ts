// **The “Users and profiles” page** — phase 4.12
// (#1351).
import { faktor, konto } from './seitenkonto'
import { sitzung } from './seitensitzung'

export const nutzer = {
  titel: 'Utenti e profili',

  laschen: {
    allgemein: 'Generali',
    konto: 'Account',
    profil: 'Profili',
    persoenlich: 'Personale',
  },

  allgemeinErklaerung:
    'Un utente è la persona. Un profilo è un ambiente di lavoro di quella persona — con set, collezione e impostazioni propri. Chi sta lavorando dove in questo momento si vede nella barra laterale sotto il nome dell’applicazione; il cambio avviene qui.',

  kontoErklaerung:
    'Con che cosa accede questo utente. Finché esiste un solo utente con un solo profilo, non viene chiesto nulla.',

  profilErklaerung:
    'Chi vuole collezionare separatamente per un torneo non ha bisogno di un secondo account, ma di un secondo profilo.',

  niemand:
    'Ancora nessun altro utente. Il primo è stato creato al primo avvio.',
  keineProfile: 'Questo utente non ha ancora nessun profilo.',

  neuerNutzer: 'Crea un utente',
  neuerNutzerName: 'Nome visualizzato del nuovo utente',
  neuerNutzerNameHinweis:
    'Univoco e senza spazi ai margini — distingue due utenti',
  neuerNutzerKnopf: 'Crea utente',
  nutzerAngelegt: 'Creato: {name}',

  neuesProfil: 'Crea un altro profilo',
  neuesProfilNutzer: 'Per quale utente',
  neuesProfilName: 'Nome del profilo',
  neuesProfilNameHinweis:
    'Univoco per utente — due persone possono avere entrambe un “Torneo”',
  neuesProfilKnopf: 'Crea profilo',
  profilAnlegenOeffnen: 'Crea profilo …',
  profilAngelegt: 'Creato: {name}',

  wahl: 'In quale profilo stai lavorando',
  wahlErklaerung:
    'Set, mazzi, collezione, tornei e impostazioni appartengono al profilo scelto. Cambiare profilo non elimina nulla — ciò che c’è qui ci sarà ancora quando torni.',
  hier: 'Stai lavorando qui',
  hierhin: 'Passa qui',
  hierhinFuer: 'Passa a {nutzer} · {profil}',

  nutzerwahl: 'Chi sta lavorando',
  nutzerwahlErklaerung:
    'Guardare il profilo di un’altra persona non cambia chi sei — le tue autorizzazioni decidono dove puoi andare. Non è la stessa cosa del profilo qui sopra. Per diventare un’altra persona dimostri il suo secondo fattore; poi lavori nel suo primo profilo.',
  dasBinIch: 'Sei tu in questo momento',
  werden: 'Diventa questa persona',
  werdenFuer: 'Diventa {nutzer}',

  anmeldung: 'Accesso tramite un provider',
  anmeldungErklaerung:
    'Per accedere tramite un provider, inserisci il suo `client_id` in **Impostazioni › Utente e profilo › Accesso tramite un provider**. MagicSetEditor4Mac non include credenziali; l’applicazione OAuth la crei tu presso il provider, e un segreto, se richiesto, va nel portachiavi e non nell’archivio.',
  anmelden: 'Accedi',
  anbieterstand: '{eingerichtet}, {geprueft}',
  eingerichtet: 'configurato',
  nichtEingerichtet: 'non configurato',
  geprueft: {
    simuliert: 'testato solo con una simulazione',
    realGeprueft: 'testato con il provider reale',
  },
  anmeldenBei: 'Accedi con {anbieter}',
  anmeldungAbbrechen: 'Annulla',
  anmeldungAbbrechenBei: 'Annulla l’accesso con {anbieter}',
  verbunden: 'Collegato',
  verbundenAls: 'Collegato come {name}',
  loesen: 'Scollega',
  loesenVon: 'Scollega da {anbieter}',

  anbieternamen: {
    github: 'GitHub',
    google: 'Google',
    microsoft: 'Microsoft',
    x: 'X',
  } as Record<string, string>,

  pflege: 'Modifica utenti e profili',
  pflegeErklaerung:
    'Un nome si può cambiare. Per eliminare, il nome deve restare com’è — conferma chi è la persona interessata.',
  bearbeiten: 'Modifica {name}',
  bearbeitenKurz: 'Modifica',
  neuerNameFuer: 'Nome',
  umbenennen: 'Rinomina',
  loeschen: 'Elimina',
  loeschenBestaetigt: 'Elimina {name} definitivamente',
  loeschenFrage: 'Eliminare {name} definitivamente?',
  loeschenJa: 'Elimina definitivamente',
  abbrechen: 'Annulla',
  nutzerLoeschenHinweis:
    'Un utente eliminato porta con sé i suoi profili. L’ultimo utente non può essere eliminato.',
  profilLoeschenHinweis:
    'L’ultimo profilo di un utente non può essere eliminato. Ciò che si trova nel profilo resta nell’archivio, ma non è più raggiungibile.',

  angaben: 'Dati personali',
  angabenErklaerung:
    'Tutto tranne il nome visualizzato è facoltativo. L’**appellativo** è un campo libero: qualsiasi elenco lascerebbe fuori qualcuno, e un’applicazione che costringe a scegliere tra voci sbagliate si è rivolta alla persona in modo errato. Indirizzo, numero di telefono e password si trovano in **Account**.',
  angabenSichern: 'Salva i dati',
  felder: {
    name: 'Nome visualizzato',
    anrede: 'Appellativo',
    namenszusatz: 'Suffisso del nome',
    vorname: 'Nome',
    weiterenamen: 'Altri nomi',
    nachname: 'Cognome',
    slogan: 'Slogan',
    kurzebio: 'Biografia breve',
    langebio: 'Biografia lunga',
  },

  // **Die Profil-Einstellungen** — Werkstück 9 von Phase 4.12 (#1351).
  badge: 'Distintivo del profilo',
  badgeVon: 'Distintivo del profilo di {name}',
  badgeWaehlen: 'Scegli distintivo …',
  badgeEntfernen: 'Rimuovi distintivo',
  profilangaben: 'Dati del profilo',
  profilangabenErklaerung:
    'Appartengono a **questo profilo**, non a te — chi ne ha più di uno li distingue qui. Lo **slogan del profilo** compare sotto il tuo e non lo sostituisce. Il **distintivo** si trova in basso a destra sulla tua immagine del profilo: mostra il ruolo in cui stai lavorando ora.',
  profilangabenSichern: 'Salva profilo',
  profilfelder: {
    name: 'Nome del profilo',
    subslogan: 'Slogan del profilo',
    kurzebeschreibung: 'Descrizione breve',
    langebeschreibung: 'Descrizione lunga',
  },

  // **Technische Profile und ihre Zugriffstoken** — Werkstück 10 von
  // Phase 4.12 (#1351), Punkte 12 und 13.
  technisch: 'Profilo tecnico — dietro non c’è nessuna persona',
  technischErklaerung:
    'Un profilo tecnico tiene separato ciò che un servizio scrive da solo da ciò che fai tu. Chi fa lavorare un servizio per sé non vuole che scriva nel **suo** profilo: ciò che compare lì non è nato dalla sua mano. Resta comunque tuo, perché qualcuno deve poterlo disattivare.',
  ablauf: 'Scade il (vuoto: mai)',
  ablaufErklaerung:
    'Un momento a partire dal quale questo profilo è scaduto — **non è un ordine di eliminazione**. Nulla sparisce da solo.',

  token: 'Token di accesso',
  tokenErklaerung:
    'Così un chiamante automatico si identifica sull’endpoint REST — in **questo** profilo e con l’autorizzazione indicata qui. La finestra, la riga di comando e MCP non richiedono alcun token: lì è già stabilito chi sta lavorando.',
  tokenLeer: 'Questo profilo non ha ancora nessun token di accesso.',
  tokenBezeichnung: 'A cosa serve',
  tokenScope: 'Autorizzazione',
  tokenLesen: 'Solo lettura',
  tokenSchreiben: 'Lettura e scrittura',
  tokenAblauf: 'Scade il (vuoto: mai)',
  tokenAnlegen: 'Crea token',
  tokenAendern: 'Applica',
  tokenLoeschen: 'Revoca',
  tokenLoeschenFrage: 'Revocare il token “{name}”?',
  tokenLoeschenErklaerung:
    'Chi lo usa non potrà più entrare. Un nuovo token va poi inserito di nuovo lì.',
  tokenLoeschenJa: 'Revoca',
  tokenAbgelaufen: 'scaduto',
  tokenUnbefristet: 'senza scadenza',

  tokenEinmalig:
    'Il valore in chiaro compare **una sola volta**, subito dopo la creazione. Dopo non esiste più da nessuna parte; chi lo perde crea un nuovo token.',
  tokenKlartext: 'Copialo ora — questo valore non tornerà:',

  // **Rotation, bulk revocation, expiry presets and MCP client set-up** —
  // piece O2 of phase 5.2 (#1426).
  tokenPrincipal: 'Per chi lavora',
  tokenAblaufwahl: 'Scadenza',
  tokenAblaufwahlen: {
    tag: 'Un giorno',
    monat: 'Trenta giorni',
    jahr: 'Un anno',
    nie: 'Mai',
    eigen: 'Momento a scelta',
  },
  tokenRotieren: 'Nuovo segreto',
  tokenRotierenFrage: 'Generare un nuovo segreto?',
  tokenRotierenErklaerung:
    'Il vecchio vale ancora per un’ora, così un lavoro in corso può terminare. Il nuovo valore in chiaro compare una sola volta.',
  tokenAlleWiderrufen: 'Revoca tutti',
  tokenAlleWiderrufenFrage:
    'Revocare tutti i token di accesso di questo profilo?',
  tokenAlleWiderrufenErklaerung:
    'Ogni chiamante automatico che ne usa uno non potrà più entrare — anche a sessione in corso. Ciò che è stato creato con essi resta.',
  tokenGezogen: 'Token revocati: {zahl}.',
  tokenLetzteNutzung: 'ultimo utilizzo {zeit}',
  tokenNieBenutzt: 'mai usato',
  tokenVorgaengerBis: 'il vecchio segreto vale fino a {zeit}',
  tokenOnboarding: 'Come configurare un client MCP',
  tokenOnboardingSchritt1:
    'Crea un token e copia il valore in chiaro — compare una sola volta.',
  tokenOnboardingSchritt2:
    'Scrivilo in un file che solo tu puoi leggere (permessi 0600). Un file leggibile da altri viene rifiutato da “mse-mcp”.',
  tokenOnboardingSchritt3:
    'Avvia “mse-mcp” con MSE_TOKEN_DATEI impostata su quel file — come variabile d’ambiente, mai come argomento: comparirebbe nell’elenco dei processi.',
  tokenOnboardingSchritt4:
    'Il token viene controllato a ogni chiamata di uno strumento: una revoca, una scadenza o un’autorizzazione più ristretta ha effetto alla chiamata successiva. Senza token, “mse-mcp” funziona come prima, senza autenticazione nel profilo mostrato.',
  tokenOnboardingBeispiel:
    '{ "mcpServers": { "mse": { "command": "/Applications/MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp", "env": { "MSE_TOKEN_DATEI": "/Users/ich/.mse/mcp-token" } } } }',
  tokenOnboardingBeispielTitel: 'Esempio di configurazione',

  // **Vorgaben, die auf Zeilen zeigen** — Werkstück 15 von Phase 4.12
  // (#1351), Punkte 4 bis 7.
  vorgaben: 'Impostazioni predefinite di questo profilo',
  vorgabenErklaerung:
    'Da che cosa parte un nuovo record in **questo profilo** — un formato, un gioco, un modello dell’archivio. A differenza delle impostazioni, i valori predefiniti puntano a **righe nell’archivio**; ciò che si può scegliere dipende da ciò che c’è.',
  vorgabenamen: {
    format: 'Formato predefinito per i nuovi mazzi',
    spiel: 'Gioco predefinito per i nuovi set',
    vorlage: 'Modello predefinito per i nuovi set',
    deck: 'Mazzi preferiti',
    lagerort: 'Ubicazioni assegnate',
  } as Record<string, string>,
  vorgabenOhne: 'nessun valore predefinito',
  vorgabenLeer: 'Per questo non c’è ancora nulla nell’archivio.',
  vorgabenVerwaist: 'Non punta a nulla, perché non esiste più:',

  // **Rechte, Rollen und Sichtbarkeit** — Werkstück 11 von Phase 4.12
  // (#1351).
  freigaben: 'Chi può lavorare in questo profilo',
  freigabenErklaerung:
    'Un **lettore** vede solo che questo profilo esiste — non che cosa contiene. Un **collaboratore** vi passa e ci lavora. Un **gestore** concede lui stesso gli accessi. Questo profilo è tuo se l’hai creato tu; allora lo gestisci comunque.',
  freigabenNurLesen:
    'Gli accessi possono essere concessi solo da chi gestisce questo profilo.',
  freigabenLeer:
    'Finora nessuno tranne il proprietario può entrare in questo profilo.',
  freigabeNutzer: 'A chi',
  freigabeNiemand: 'a nessuno',
  freigabeRolle: 'Come',
  freigabeVergeben: 'Concedi',
  freigabeZurueckziehen: 'Revoca',
  rollennamen: {
    leser: 'Lettore — vede che il profilo esiste',
    mitarbeiter: 'Collaboratore — vi passa e ci lavora',
    verwalter: 'Gestore — concede lui stesso gli accessi',
  } as Record<string, string>,

  bild: 'Immagine del profilo',
  bildVon: 'Immagine del profilo di {name}',
  bildWaehlen: 'Scegli immagine …',
  bildEntfernen: 'Rimuovi immagine',
  ohneBild: '{name} non ha un’immagine del profilo',

  nochNicht:
    'I dati sono separati tra i profili — finché nessuno apre il file dell’archivio con uno strumento esterno. Contro questo aiuta la cifratura dell’archivio in “Impostazioni › Dati e backup”.',

  // **Die vier Kontoarten** — Werkstück 6 (#1351), in einer eigenen Datei.
  konto,
  sitzung,

  // **Der zweite Faktor** — Werkstück 7 (#1351).
  faktor,

  // **Was dem Menschen gehört** — Werkstück 8 von Phase 4.12 (#1351),
  // Punkte 9, 10 und 11.
  //
  // **Am 2026-09-06 aus `seiteneinstellungen` hierher gezogen**, nach dem
  // Zweitblick über die Phase: die Einstellungen einer Person lagen an zwei
  // Orten, und wer sie suchte, suchte an beiden.

  persoenlich: 'Il mio accesso',

  persoenlichErklaerung:
    'Queste impostazioni appartengono a **te**, non a questo profilo — le porti con te in ogni tuo profilo. Le due superiori hanno la precedenza sul sistema operativo, e **in una sola direzione**: attivano, mai disattivano.',

  auskunft: 'I miei dati',

  auskunftErklaerung:
    'Ciò che questo programma ha memorizzato su di te — dati personali, modalità di accesso, i **tipi** di secondo fattore, impostazioni e nomi dei profili. **Senza password e senza segreto**: nessuno dei due è un dato su di te, ma un mezzo per essere te. Set, mazzi e collezione non ne fanno parte — appartengono a un profilo e se ne vanno per vie proprie.',

  auskunftSichern: 'Salva i miei dati …',

  datenschutz: 'Che cosa sa questa applicazione di te',

  datenschutzHinweis:
    'Quali dati vengono raccolti, dove si trovano, che cosa di essi lascia questo computer — e che cosa espressamente no. Descritto in **Aiuto → Che cosa sa questa applicazione di te**.',
} as const
