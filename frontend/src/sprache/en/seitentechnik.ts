// The texts of the **technical page** — operating state, effect counter and
// the three tools without a window.
//
// Split out of `seiten.ts` when the file grew past 400 lines with the effect
// counter (work item O9, #1433).

import { betrieb, leistungsprofil, zwischenspeicher } from './seitenbetrieb'
import { dienstzeugnis, fernwege, schnittstellen } from './seitentechnikdienst'
import { einladung } from './einladung'
import { gegenueber } from './gegenueber'

export const technik = {
  betrieb,
  dienstzeugnis,
  einladung,
  gegenueber,
  fernwege,
  schnittstellen,
  leistungsprofil,
  zwischenspeicher,
  gesundheit: 'Operating state',
  gesundheitErklaerung:
    'Whether this process is alive and whether it can work. An external service that does not answer does not make it unready — editing continues without a network.',
  lebt: 'The process answers.',
  bereit: 'It can work.',
  nichtBereit: 'It cannot work right now.',
  grundBestandZu: 'No library is open.',
  grundWanderungLaeuft: 'The library is migrating to a new schema.',
  grundBestandUnlesbar: 'The library does not answer.',
  fremdeDienste: 'External services',
  erreichbar: 'reachable',
  nichtErreichbar: 'not reachable',
  metriken: 'Effects since start',
  metrikenErklaerung:
    'What this process has done to the outside — by surface, effect class and outcome. Only checked effects are counted; plain reading does not appear here.',
  metrikenGrenze:
    'At most {grenze} rows — they do not grow with the number of cards, sets or profiles.',
  metrikenLeer: 'Nothing has affected the outside yet.',
  spalteFlaeche: 'Surface',
  spalteWirkung: 'Effect',
  spalteAusgang: 'Outcome',
  spalteAnzahl: 'Count',
  flaechen: {
    fenster: 'Window',
    kommandozeile: 'Command line',
    mcp: 'MCP',
    dienst: 'Start the service for scripts',
  },
  wirkungen: {
    lokalLesen: 'read locally',
    lokalSchreiben: 'write locally',
    externSenden: 'send outward',
    externSchreiben: 'change outside',
    verbindlich: 'binding',
  },
  ausgaenge: { gewaehrt: 'granted', abgewiesen: 'refused' },
  paket: {
    ueberschrift: 'Diagnostic bundle',
    erklaerung:
      'The same information in parts, as a file with checksums — for a report that needs more than a text. It is created where you point, and goes nowhere by itself.',
    inhalt:
      'It holds version, platform, library, packages, process counters, log lines and crash traces — no cards, images, collection or credentials.',
    zusammenstellen: 'Assemble',
    sichern: 'Save bundle',
    loeschen: 'Delete bundle',
    groesse: '{bytes} of at most {grenze}.',
    zeitraum: 'The log lines run from {von} to {bis}.',
    ausgelassen: 'Omitted — {teil}:',
    teile: {
      fassung: 'Version and platform',
      bestand: 'Library location and schema',
      pakete: 'Installed template packages',
      metriken: 'Counters of this process',
      protokoll: 'Log lines',
    },
    spalteTeil: 'Part',
    spalteDatei: 'File',
    spalteZeilen: 'Lines',
    spalteBytes: 'Size',
    spalteMitnehmen: 'Include',
    mitnehmen: 'Include {teil} in the bundle',
  },
  diagnose: {
    ueberschrift: 'Diagnostics for a bug report',
    erklaerung:
      'When something does not work, this text helps to track it down. It is shown here — passing it on is up to you.',
    inhalt:
      'It contains version, machine, database, log and template packages — no card texts, no field values, no set contents. In detail: version, machine type, the build the program comes from, the binding contract between window and program, the location of the database, the schema state, the number of orphaned rows, the location of the log, your template packages and the last 200 log lines. Your home directory appears as “~”.',
    anzeigen: 'Show diagnostics',
    kopieren: 'Copy',
    kopiert: 'Copied',
    kopierfehler:
      'Copying failed — the window has no clipboard permission. Select the text and copy it by hand.',
    sichern: 'Save as file',
    // Filter über die Zeilen (NET-02, S13.NAT-04, #1706).
    filter: 'Search the lines',
    zeilenGezeigt: '{gezeigt} of {alle} lines',
  },
  // **Die Aufschrift der Lasche — und nicht die des Abschnitts darin.**
  //
  // Sie hießen bis zum 2026-09-14 gleich, und die Lasche trug elf Kästen,
  // von denen einer „Werkzeuge ohne Fenster" war. Wer die
  // Netzeinstellungen suchte, suchte überall sonst.
  bereichUeberschrift: 'Technical and diagnostics',

  // **Das Gebiet, das dabei entstanden ist.**
  netzUeberschrift: 'Network and reachability',

  ueberschrift: 'Tools without a window',
  erklaerung:
    'This application ships three programs that work without a window. mse for the command line, mse-mcp for a language model, mse-rest for a service for scripts; they live inside the application bundle, they reach a search path only once you set them up below, and the lines here carry the full location.',
  keineShell:
    'Copied, not run — what is shown here belongs in a terminal of your choice.',
  ort: 'They live under {ort}.',
  vorhanden: 'Present: {namen}.',
  keineWerkzeuge:
    'None of the three is there. That is normal while the application runs from source — the release path puts them into the bundle.',
  kopieren: 'Copy command',
  kopiert: 'Copied.',
  kopierfehler:
    'Copying failed — the window has no clipboard permission. Select the text and copy it by hand.',
  befehle: {
    hilfe: 'What the command line can do',
    sets: 'List the sets in the library',
    verknuepfen:
      'Make mse available in the search path (asks for your password)',
    dienst: 'Start the service on the loopback interface',
    mcp: 'Start the MCP server — for a language model, over stdio',
    karten: 'List the cards of the open set',
    statistik: 'Count the open set',
  },

  // **The local search index** — work item A15 of phase 5.3
  // (#1514).
  suchindex: 'Local search index',
  suchindexErklaerung:
    'Card texts and rules passages, searchable locally — in full text and through vectors, merged into one list. Every hit names where it came from and which revision the source was in.',
  suchindexKeiner:
    'No search index has been built. Until then the search here finds nothing.',
  suchindexModell: 'Model',
  suchindexLaufzeit: 'Runtime',
  suchindexVerfahren: 'Method',
  suchindexStaende: 'Revision per source',
  suchindexSemantisch: 'Semantic search',
  suchindexJa: 'yes',
  suchindexNein: 'no',
  suchindexFrage: 'What are you looking for?',
  suchindexSuchen: 'Search',
  suchindexLeer: 'Nothing found.',
  suchindexModus: 'Searched',
  suchindexModusHybrid: 'in full text and through vectors',
  suchindexModusLexikalisch: 'in full text only',
  suchindexModusOhneIndex: 'not at all — no index is available',
  suchindexAufbau: 'Rebuild',
  suchindexAufbauErklaerung:
    'Builds a new generation beside the running one and switches over only at the end. This takes minutes and needs a configured embedding model.',
  suchindexModellfeld: 'Embedding model',
  suchindexQuellfeld: 'Bulk file',
  suchindexRevisionfeld: 'Source revision',
  suchindexDelta: 'Read only what changed',
  suchindexAufbauen: 'Build',
  suchindexAbbrechen: 'Cancel',
  suchindexLaeuft: 'Build running',
  suchindexGeschrieben: 'sections written',
  suchindexFertig: 'generation built',
  // **Configuration for agents** — A14, #1453.
  agenten: 'Ready for an agent',
  agentenErklaerung:
    'Claude Code and Codex speak MCP. The fragments below carry the full path of mse-mcp on this machine — copy and paste. This application writes to none of these files and starts no agent.',
  agentenOhneGeheimnis:
    'No access key appears in them. mse-mcp reads those itself from the keychain — anything written here would sit in plain text in a file people share.',
  agentenNamen: {
    claude: 'Claude Code',
    codex: 'Codex',
  } as Record<string, string>,
  befehlszeileTitel: 'Set up the command line',
  befehlszeileSatz:
    'Places links to mse and mse-mcp in a folder you own — without administrator rights and without touching your shell profile. Whatever is already there and did not come from here stays.',
  befehlszeileVerzeichnis: 'Folder: {verzeichnis}',
  befehlszeileNichtImPfad:
    'This folder is not on your PATH. Add this line to your shell profile yourself:',
  befehlszeileStaende: {
    frei: 'not set up yet',
    eigen: 'set up',
    veraltet: 'outdated — points to a program that is no longer there',
    fremd: 'another command is there, it stays',
  } as Record<string, string>,
  befehlszeileVerdeckt:
    'Hidden by {pfad}: that one runs when you type the name.',
  befehlszeilePruefen: 'Check with',
  befehlszeileEinrichten: 'Set up',
  befehlszeileHerausnehmen: 'Take out again',
  befehlszeileKopierenPfad: 'Copy the line for the shell profile',
  befehlszeileKopierenPruefung: 'Copy the check command',
  agentenDatei: 'Belongs in {datei}',
  agentenInhalt: 'Contents for {datei}',
  agentenBefehl: 'Or through the agent’s command line',
  agentenNachgelesen: 'Looked up: {quelle}',
  agentenKopierenDatei: 'Copy fragment',
  agentenKopierenBefehl: 'Copy the agent’s command',

  peerTitel: 'How this profile identifies itself',

  peerZeigen: 'Show identity',
  peerErklaerung:
    'A key pair that lets others recognise they are really talking to you. The private half stays in this machine’s keychain and goes nowhere; only the identifier leaves.',
  // Was diese Identität nicht ist (NET-02, S13.NAT-05, #1706).
  peerNichtKonto:
    'It belongs to this profile — it is neither your Apple Account nor the name of this computer.',
  peerStandOhne: 'This profile has no identity yet.',
  peerStandBereit: 'Ready.',
  peerStandVerloren: 'The key is missing.',
  peerStandFremd: 'Keychain and database disagree.',
  peerKennung: 'Identifier',
  peerFingerabdruck: 'Fingerprint',
  peerFingerabdruckHinweis:
    'This is what you read out to the other side — a display name is no proof.',
  peerSeit: 'Since {zeitpunkt}',
  peerVorher: 'Previously: {kennung}',
  peerAnlegen: 'Create identity',
  peerRotieren: 'Replace identity',
  peerRotierenFrage:
    'Create a new identity? Everyone who confirmed the current fingerprint has to check it again.',
  peerRotierenJa: 'Replace',
  peerRotierenNein: 'Cancel',

  // **The known counterparts** — work item N1 of the same issue.
  kontakteTitel: 'Who this profile knows',
  kontakteErklaerung:
    'Counterparts this profile has already met. Meeting alone permits nothing: only comparing the fingerprint makes someone verifiable, and only a per-purpose grant permits anything.',
  kontakteLeer:
    'Nobody met yet. Whoever gets in touch appears here afterwards — unconfirmed.',
  kontakteZeigen: 'Show contacts',
  kontakteAufnehmen: 'Add contact',
  kontakteAufnehmenErklaerung:
    'While there is no connection, identifiers are exchanged by hand. Whoever is added here is unconfirmed and may do nothing.',
  kontakteAufnehmenKennung: "The counterpart's identifier",
  kontakteAufnehmenName: 'What you call them',
  kontakteStandUnbestaetigt: 'Unconfirmed',
  kontakteStandBestaetigt: 'Confirmed',
  kontakteStandGesperrt: 'Blocked',
  kontakteAbdruck: 'Fingerprint',
  kontakteNennt: 'Calls itself: {name}',
  kontakteNenntHinweis: 'A self-chosen name is not proof of identity.',
  kontakteGesehen: 'Last seen: {zeitpunkt}',
  kontakteVergleichen: 'Enter the fingerprint you compared',
  kontakteBestaetigen: 'Confirm',
  kontakteSperren: 'Block',
  kontakteEntsperren: 'Unblock',
  kontakteRechte: 'Granted',
  kontakteKeineRechte: 'Nothing granted.',
  zweckPaarung: 'Getting acquainted',
  zweckTeilen: 'Transferring objects',
  zweckHandel: 'Trading cards',
  zweckPartie: 'Playing a game',
  kontakteAdresse: 'Reachable at',
  kontakteAdresseBeispiel: '192.168.1.7:41234',
  kontakteAdresseHinweis:
    'Either an address on the same network — or an invitation the other side sent you. Empty means none.',
  kontakteAdresseMerken: 'Remember',

  // **Die lokale Notiz und die freigegebenen Sichten** — Werkstück T3
  // von Phase 5.6 (#1468).
  kontakteNotiz: 'Note',
  kontakteNotizBeispiel: 'trades fairly, slow to reply …',
  kontakteNotizHinweis:
    'For you only. It never travels — neither into a view nor into a share.',
  kontakteNotizMerken: 'Save note',
  freigabenTitel: 'What this peer sees',
  freigabenErklaerung:
    'Without a share it sees nothing. Several scopes apply side by side — the wider one wins.',
  freigabenLeer: 'No view granted.',
  freigabenOhneZweck: 'This peer may do nothing yet. Grant it a purpose first.',
  freigabenZweck: 'For what',
  freigabenUmfang: 'What',
  freigabenErteilen: 'Grant view',
  freigabenWiderrufen: 'Revoke',
  freigabenVorschau: 'See what it sees',
  freigabenVorschauLeer: 'It would see nothing.',
  fernTitel: 'Be reachable across networks',
  fernErklaerung:
    'This lets somebody reach you even when you are not on the same network — no port forwarding, no account. You get an invitation and send it to them, by whatever means you like.',
  fernStill: 'This profile is not reachable across networks.',
  fernHinweis:
    'The connection looks for the direct path and only uses a relay when there is no other way. Only encrypted traffic passes through the relay — no cards, no names.',
  fernStarten: 'Create invitation',
  fernBeenden: 'Stop being reachable',
  fernWeitergeben:
    'This is what the other side needs. Whoever intercepts it gets a line, not an identity — the check happens on connecting, through the fingerprint.',
  fernStirbt:
    'The invitation holds as long as this profile is reachable. Stop it and the invitation becomes void — next time you get a new one.',
  fernKopieren: 'Copy invitation',
  fernKopiert: 'Copied.',
  horchenTitel: 'Be reachable for others',
  horchenErklaerung:
    'While this profile is listening, a known peer can open a connection ' +
    'to it. This never happens on its own — not even after a restart — ' +
    'and switching profiles ends it.',
  horchenAdresse: 'Address to listen on',
  horchenAdresseBeispiel: ':41234',
  horchenAdresseHinweis:
    'The port alone (:41234) listens on every interface of this machine. ' +
    'With 127.0.0.1 in front, only on this machine itself.',
  horchenStarten: 'Become reachable',
  horchenBeenden: 'Stop being reachable',
  horchenLaeuft: 'Listening on {adresse}',
  horchenNamen: 'Name (survives an address change)',
  horchenAdressen: 'Address',
  horchenNurHier:
    'With this binding nobody reaches you from outside: 127.0.0.1 listens on this machine only. Give just the port instead (:41234) so this Mac is reachable on its network.',
  horchenNurDiesesNetz:
    'Both hold on this network. A leased address can change, a name usually survives that — and an address that holds everywhere only arrives with connections across networks.',
  horchenStill: 'This profile is not listening right now.',
  horchenWeitergeben:
    'This is what the other side needs — along with this profile’s key.',
  sichtbarTitel: 'Be discovered nearby',
  sichtbarErklaerung:
    'Announcing tells everyone on the same network that this profile ' +
    'exists — including its peer ID. Being reachable and being discovered ' +
    'are two decisions, and this is the second one.',
  sichtbarName: 'Name to announce under',
  sichtbarNameHinweis:
    'A self-chosen name proves nothing. This Mac’s host name does not go out.',
  sichtbarWerden: 'Announce',
  sichtbarBeenden: 'Stop announcing',
  sichtbarLaeuft: 'This profile is announcing itself.',
  sichtbarStill: 'This profile is not announcing itself.',
  sichtbarBrauchtHorchen:
    'Become reachable first — announcing without a port would be an ' +
    'invitation to a door with nobody behind it.',
  nachbarnTitel: 'Who is nearby',
  nachbarnErklaerung:
    'Peers announcing themselves on the same network right now. What you ' +
    'see is their claim — it is only checked when you connect.',
  nachbarnSuchen: 'Take a look',
  nachbarnSucht: 'Looking …',
  nachbarnLeer: 'Nobody is announcing themselves right now.',
  nachbarnUnbekannt: 'Unknown',
  nachbarnAufnehmen: 'Add as contact',
  nachbarnGemerkt: 'Remembered as',
  nachbarnSchonBekannt: 'Already in the list',
  // Ein bekannter Peer an einer neuen Adresse (NET-02.NAT-02, #1706).
  nachbarnAdresseNeu: 'This peer is at a new address.',
  nachbarnAdresseBisher: 'Remembered so far',
  nachbarnAdresseKeine: 'none',
  nachbarnAdresseUebernehmen: 'Use new address',
  nachbarnAdresseHinweis:
    'This only changes where to connect. The peer is trusted no more than before, and its key is checked on every connection.',
  nachbarnVerworfen:
    '{zahl} replies carried no peer ID and are not in the list.',

  // **How this program goes online** — N5
  // (#1571).
  netzwegTitel: "From this process's environment",
  netzwegErklaerung:
    'These values come from the environment the program was started in. They override the fields below, and no window changes them.',
  netzwegProxyHTTP: 'Proxy for http',
  netzwegProxyHTTPS: 'Proxy for https',
  netzwegProxyAus: 'No proxy for',
  netzwegAusUmgebung: '{name}: {wert}',

  // **The two switchable services** — N6
  // (#1572).
  dienstRESTTitel: 'The interface for scripts',
  dienstRESTErklaerung:
    'A web service over the same database — for reading and changing from a script. It runs only while you let it, and ends with this window.',
  dienstMCPTitel: 'The tool server over HTTP',
  dienstMCPErklaerung:
    'The same tools as over standard input, here under an address — reachable from another device too.',
  dienstRESTAdresse: 'Address of the script interface',
  dienstRESTPort: 'Port of the script interface',
  dienstMCPAdresse: 'Address of the tool server',
  dienstMCPPort: 'Port of the tool server',
  dienstMarke: 'Access',
  dienstMarkeKopieren: 'Copy access',
  dienstFingerabdruck: 'Fingerprint',
  dienstFingerabdruckKopieren: 'Copy fingerprint',
  dienstKopiert: 'Copied',
  dienstStarten: 'Start',
  dienstBeenden: 'Stop',
  dienstErreichbar: 'Reachable at:',
  dienstMitTLS:
    'Encrypted. The certificate is self-issued and carries this profile’s ID — whoever connects compares its fingerprint.',
  dienstOhneTLS:
    'Unencrypted, and that is fine here: the address applies only on this machine.',
  dienstHinweis:
    '127.0.0.1 means: this machine only. 0.0.0.0 means: every device on the network — then encryption and access are mandatory, and both set themselves up. Port 0 means: find a free one.',
}
