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
  gesundheit: 'Estado de funcionamiento',
  gesundheitErklaerung:
    'Si este proceso está vivo y si puede trabajar. Un servicio externo que no responde no lo deja sin estar listo: la edición continúa sin red.',
  lebt: 'El proceso responde.',
  bereit: 'Puede trabajar.',
  nichtBereit: 'Ahora mismo no puede trabajar.',
  grundBestandZu: 'No hay ninguna base de datos abierta.',
  grundWanderungLaeuft: 'La base de datos se está migrando a un esquema nuevo.',
  grundBestandUnlesbar: 'La base de datos no responde.',
  fremdeDienste: 'Servicios externos',
  erreichbar: 'accesible',
  nichtErreichbar: 'no accesible',
  metriken: 'Efectos desde el inicio',
  metrikenErklaerung:
    'Lo que este proceso ha hecho hacia el exterior, por superficie, clase de efecto y resultado. Solo se cuentan los efectos comprobados; la lectura simple no aparece aquí.',
  metrikenGrenze:
    'Como máximo {grenze} filas: no crecen con el número de cartas, sets o perfiles.',
  metrikenLeer: 'Hasta ahora no se ha afectado nada del exterior.',
  spalteFlaeche: 'Superficie',
  spalteWirkung: 'Efecto',
  spalteAusgang: 'Resultado',
  spalteAnzahl: 'Cantidad',
  flaechen: {
    fenster: 'Ventana',
    kommandozeile: 'Línea de comandos',
    mcp: 'MCP',
    dienst: 'Iniciar el servicio para scripts',
  },
  wirkungen: {
    lokalLesen: 'lectura local',
    lokalSchreiben: 'escritura local',
    externSenden: 'envío al exterior',
    externSchreiben: 'cambio en el exterior',
    verbindlich: 'vinculante',
  },
  ausgaenge: { gewaehrt: 'concedido', abgewiesen: 'rechazado' },
  paket: {
    ueberschrift: 'Paquete de diagnóstico',
    erklaerung:
      'La misma información por partes, como archivo con sumas de comprobación, para un informe que necesita más que un texto. Se crea donde indiques y no va a ningún sitio por sí solo.',
    inhalt:
      'Contiene versión, plataforma, base de datos, paquetes, contadores del proceso, líneas de registro y rastros de fallos; no incluye cartas, imágenes, colección ni credenciales.',
    zusammenstellen: 'Reunir',
    sichern: 'Guardar paquete',
    loeschen: 'Eliminar paquete',
    groesse: '{bytes} de un máximo de {grenze}.',
    zeitraum: 'Las líneas de registro van de {von} a {bis}.',
    ausgelassen: 'Omitido — {teil}:',
    teile: {
      fassung: 'Versión y plataforma',
      bestand: 'Ubicación y esquema de la base de datos',
      pakete: 'Paquetes de plantillas instalados',
      metriken: 'Contadores de este proceso',
      protokoll: 'Líneas de registro',
    },
    spalteTeil: 'Parte',
    spalteDatei: 'Archivo',
    spalteZeilen: 'Líneas',
    spalteBytes: 'Tamaño',
    spalteMitnehmen: 'Incluir',
    mitnehmen: 'Incluir {teil} en el paquete',
  },
  diagnose: {
    ueberschrift: 'Diagnóstico para un informe de error',
    erklaerung:
      'Cuando algo no funciona, este texto ayuda a encontrar la causa. Se muestra aquí; pasarlo a alguien depende de ti.',
    inhalt:
      'Contiene versión, equipo, base de datos, registro y paquetes de plantillas; no incluye textos de cartas, valores de campos ni contenido de sets. En detalle: versión, tipo de equipo, la compilación de la que procede el programa, el contrato de enlaces entre ventana y programa, la ubicación de la base de datos, el estado del esquema, el número de filas huérfanas, la ubicación del registro, tus paquetes de plantillas y las últimas 200 líneas del registro. Tu carpeta personal aparece como «~».',
    anzeigen: 'Mostrar diagnóstico',
    kopieren: 'Copiar',
    kopiert: 'Copiado',
    kopierfehler:
      'No se pudo copiar: la ventana no tiene permiso para el portapapeles. Selecciona el texto y cópialo a mano.',
    sichern: 'Guardar como archivo',
    // Filter über die Zeilen (NET-02, S13.NAT-04, #1706).
    filter: 'Buscar en las líneas',
    zeilenGezeigt: '{gezeigt} de {alle} líneas',
  },
  // **Die Aufschrift der Lasche — und nicht die des Abschnitts darin.**
  //
  // Sie hießen bis zum 2026-09-14 gleich, und die Lasche trug elf Kästen,
  // von denen einer „Werkzeuge ohne Fenster" war. Wer die
  // Netzeinstellungen suchte, suchte überall sonst.
  bereichUeberschrift: 'Técnica y diagnóstico',

  // **Das Gebiet, das dabei entstanden ist.**
  netzUeberschrift: 'Red y conectividad',

  ueberschrift: 'Herramientas sin ventana',
  erklaerung:
    'Esta aplicación incluye tres programas que funcionan sin ventana. mse para la línea de comandos, mse-mcp para un modelo de lenguaje, mse-rest para un servicio para scripts; están dentro del paquete de la aplicación, llegan a una ruta de búsqueda solo cuando los configuras abajo, y las líneas de aquí llevan la ubicación completa.',
  keineShell:
    'Se copia, no se ejecuta: lo que aparece aquí va en un terminal de tu elección.',
  ort: 'Están en {ort}.',
  vorhanden: 'Presentes: {namen}.',
  keineWerkzeuge:
    'No está ninguno de los tres. Es lo normal mientras la aplicación se ejecuta desde el código fuente: el proceso de publicación los pone en el paquete.',
  kopieren: 'Copiar comando',
  kopiert: 'Copiado.',
  kopierfehler:
    'No se pudo copiar: la ventana no tiene permiso para el portapapeles. Selecciona el texto y cópialo a mano.',
  befehle: {
    hilfe: 'Qué puede hacer la línea de comandos',
    sets: 'Listar los sets de la base de datos',
    verknuepfen:
      'Hacer que mse esté disponible en la ruta de búsqueda (pide tu contraseña)',
    dienst: 'Iniciar el servicio en la interfaz de bucle local',
    mcp: 'Iniciar el servidor MCP: para un modelo de lenguaje, por stdio',
    karten: 'Listar las cartas del set abierto',
    statistik: 'Contar el set abierto',
  },

  // **The local search index** — work item A15 of phase 5.3
  // (#1514).
  suchindex: 'Índice de búsqueda local',
  suchindexErklaerung:
    'Textos de cartas y pasajes de reglas, con búsqueda local: en texto completo y mediante vectores, combinados en una sola lista. Cada resultado indica de dónde procede y en qué revisión estaba la fuente.',
  suchindexKeiner:
    'No se ha creado ningún índice de búsqueda. Hasta entonces, la búsqueda aquí no encuentra nada.',
  suchindexModell: 'Modelo',
  suchindexLaufzeit: 'Entorno de ejecución',
  suchindexVerfahren: 'Método',
  suchindexStaende: 'Revisión por fuente',
  suchindexSemantisch: 'Búsqueda semántica',
  suchindexJa: 'sí',
  suchindexNein: 'no',
  suchindexFrage: '¿Qué buscas?',
  suchindexSuchen: 'Buscar',
  suchindexLeer: 'No se encontró nada.',
  suchindexModus: 'Búsqueda',
  suchindexModusHybrid: 'en texto completo y mediante vectores',
  suchindexModusLexikalisch: 'solo en texto completo',
  suchindexModusOhneIndex: 'ninguna: no hay índice disponible',
  suchindexAufbau: 'Reconstruir',
  suchindexAufbauErklaerung:
    'Crea una generación nueva junto a la que está en uso y solo cambia a ella al final. Tarda minutos y necesita un modelo de embeddings configurado.',
  suchindexModellfeld: 'Modelo de embeddings',
  suchindexQuellfeld: 'Archivo masivo',
  suchindexRevisionfeld: 'Revisión de la fuente',
  suchindexDelta: 'Leer solo lo que cambió',
  suchindexAufbauen: 'Crear',
  suchindexAbbrechen: 'Cancelar',
  suchindexLaeuft: 'Creación en curso',
  suchindexGeschrieben: 'secciones escritas',
  suchindexFertig: 'generación creada',
  // **Configuration for agents** — A14, #1453.
  agenten: 'Preparado para un agente',
  agentenErklaerung:
    'Claude Code y Codex hablan MCP. Los fragmentos de abajo llevan la ruta completa de mse-mcp en este equipo: copia y pega. Esta aplicación no escribe en ninguno de esos archivos ni inicia ningún agente.',
  agentenOhneGeheimnis:
    'En ellos no aparece ninguna clave de acceso. mse-mcp las lee él mismo del llavero: lo que se escribiera aquí quedaría en texto plano en un archivo que la gente comparte.',
  agentenNamen: {
    claude: 'Claude Code',
    codex: 'Codex',
  } as Record<string, string>,
  befehlszeileTitel: 'Configurar la línea de comandos',
  befehlszeileSatz:
    'Coloca enlaces a mse y mse-mcp en una carpeta tuya, sin derechos de administrador y sin tocar tu perfil de shell. Lo que ya esté ahí y no venga de aquí se queda.',
  befehlszeileVerzeichnis: 'Carpeta: {verzeichnis}',
  befehlszeileNichtImPfad:
    'Esta carpeta no está en tu PATH. Añade tú mismo esta línea a tu perfil de shell:',
  befehlszeileStaende: {
    frei: 'aún no configurado',
    eigen: 'configurado',
    veraltet: 'obsoleto: apunta a un programa que ya no está',
    fremd: 'hay otro comando ahí; se queda',
  } as Record<string, string>,
  befehlszeileVerdeckt:
    'Tapado por {pfad}: ese es el que se ejecuta al escribir el nombre.',
  befehlszeilePruefen: 'Comprobar con',
  befehlszeileEinrichten: 'Configurar',
  befehlszeileHerausnehmen: 'Quitar de nuevo',
  befehlszeileKopierenPfad: 'Copiar la línea para el perfil de shell',
  befehlszeileKopierenPruefung: 'Copiar el comando de comprobación',
  agentenDatei: 'Va en {datei}',
  agentenInhalt: 'Contenido para {datei}',
  agentenBefehl: 'O mediante la línea de comandos del agente',
  agentenNachgelesen: 'Consultado: {quelle}',
  agentenKopierenDatei: 'Copiar fragmento',
  agentenKopierenBefehl: 'Copiar el comando del agente',

  peerTitel: 'Cómo se identifica este perfil',

  peerZeigen: 'Mostrar la identidad',
  peerErklaerung:
    'Un par de claves que permite a otros reconocer que realmente hablan contigo. La mitad privada se queda en el llavero de este equipo y no va a ningún sitio; solo sale el identificador.',
  // Was diese Identität nicht ist (NET-02, S13.NAT-05, #1706).
  peerNichtKonto:
    'Pertenece a este perfil: no es tu cuenta de Apple ni el nombre de este ordenador.',
  peerStandOhne: 'Este perfil todavía no tiene identidad.',
  peerStandBereit: 'Listo.',
  peerStandVerloren: 'Falta la clave.',
  peerStandFremd: 'El llavero y la base de datos no coinciden.',
  peerKennung: 'Identificador',
  peerFingerabdruck: 'Huella',
  peerFingerabdruckHinweis:
    'Esto es lo que le lees en voz alta a la otra parte: un nombre visible no es una prueba.',
  peerSeit: 'Desde {zeitpunkt}',
  peerVorher: 'Antes: {kennung}',
  peerAnlegen: 'Crear identidad',
  peerRotieren: 'Sustituir identidad',
  peerRotierenFrage:
    '¿Crear una identidad nueva? Todos los que confirmaron la huella actual tienen que comprobarla de nuevo.',
  peerRotierenJa: 'Sustituir',
  peerRotierenNein: 'Cancelar',

  // **The known counterparts** — work item N1 of the same issue.
  kontakteTitel: 'A quién conoce este perfil',
  kontakteErklaerung:
    'Interlocutores con los que este perfil ya se ha encontrado. Conocerse por sí solo no permite nada: solo comparar la huella hace verificable a alguien, y solo una concesión por finalidad permite algo.',
  kontakteLeer:
    'Todavía no has conocido a nadie. Quien se ponga en contacto aparecerá aquí después, sin confirmar.',
  kontakteZeigen: 'Mostrar contactos',
  kontakteAufnehmen: 'Añadir contacto',
  kontakteAufnehmenErklaerung:
    'Mientras no haya conexión, los identificadores se intercambian a mano. Quien se añada aquí queda sin confirmar y no puede hacer nada.',
  kontakteAufnehmenKennung: 'El identificador del interlocutor',
  kontakteAufnehmenName: 'Cómo lo llamas',
  kontakteStandUnbestaetigt: 'Sin confirmar',
  kontakteStandBestaetigt: 'Confirmado',
  kontakteStandGesperrt: 'Bloqueado',
  kontakteAbdruck: 'Huella',
  kontakteNennt: 'Se hace llamar: {name}',
  kontakteNenntHinweis:
    'Un nombre elegido por uno mismo no es una prueba de identidad.',
  kontakteGesehen: 'Visto por última vez: {zeitpunkt}',
  kontakteVergleichen: 'Introduce la huella que has comparado',
  kontakteBestaetigen: 'Confirmar',
  kontakteSperren: 'Bloquear',
  kontakteEntsperren: 'Desbloquear',
  kontakteRechte: 'Concedido',
  kontakteKeineRechte: 'No se ha concedido nada.',
  zweckPaarung: 'Conocerse',
  zweckTeilen: 'Transferir objetos',
  zweckHandel: 'Intercambiar cartas',
  zweckPartie: 'Jugar una partida',
  kontakteAdresse: 'Accesible en',
  kontakteAdresseBeispiel: '192.168.1.7:41234',
  kontakteAdresseHinweis:
    'O bien una dirección en la misma red, o una invitación que te envió la otra parte. Vacío significa ninguna.',
  kontakteAdresseMerken: 'Recordar',

  // **Die lokale Notiz und die freigegebenen Sichten** — Werkstück T3
  // von Phase 5.6 (#1468).
  kontakteNotiz: 'Nota',
  kontakteNotizBeispiel: 'intercambia con justicia, tarda en responder …',
  kontakteNotizHinweis:
    'Solo para ti. Nunca viaja: ni en una vista ni en un recurso compartido.',
  kontakteNotizMerken: 'Guardar nota',
  freigabenTitel: 'Qué ve este interlocutor',
  freigabenErklaerung:
    'Sin una concesión no ve nada. Varios ámbitos se aplican a la vez; prevalece el más amplio.',
  freigabenLeer: 'No se ha concedido ninguna vista.',
  freigabenOhneZweck:
    'Este interlocutor todavía no puede hacer nada. Concédele primero una finalidad.',
  freigabenZweck: 'Para qué',
  freigabenUmfang: 'Qué',
  freigabenErteilen: 'Conceder vista',
  freigabenWiderrufen: 'Revocar',
  freigabenVorschau: 'Ver lo que ve',
  freigabenVorschauLeer: 'No vería nada.',
  fernTitel: 'Ser accesible entre redes',
  fernErklaerung:
    'Esto permite que alguien te alcance aunque no estés en la misma red: sin redirección de puertos y sin cuenta. Recibes una invitación y se la envías por el medio que quieras.',
  fernStill: 'Este perfil no es accesible entre redes.',
  fernHinweis:
    'La conexión busca el camino directo y solo usa un relay cuando no hay otro. Por el relay solo pasa tráfico cifrado: ni cartas ni nombres.',
  fernStarten: 'Crear invitación',
  fernBeenden: 'Dejar de ser accesible',
  fernWeitergeben:
    'Esto es lo que necesita la otra parte. Quien lo intercepte obtiene una línea, no una identidad: la comprobación se hace al conectar, mediante la huella.',
  fernStirbt:
    'La invitación vale mientras este perfil sea accesible. Si lo detienes, la invitación queda anulada; la próxima vez recibirás una nueva.',
  fernKopieren: 'Copiar invitación',
  fernKopiert: 'Copiada.',
  horchenTitel: 'Ser accesible para otros',
  horchenErklaerung:
    'Mientras este perfil escucha, un interlocutor conocido puede abrirle ' +
    'una conexión. Esto nunca ocurre por sí solo, ni siquiera tras ' +
    'reiniciar, y cambiar de perfil lo termina.',
  horchenAdresse: 'Dirección en la que escuchar',
  horchenAdresseBeispiel: ':41234',
  horchenAdresseHinweis:
    'Solo el puerto (:41234) escucha en todas las interfaces de este equipo. ' +
    'Con 127.0.0.1 delante, solo en este mismo equipo.',
  horchenStarten: 'Hacerse accesible',
  horchenBeenden: 'Dejar de ser accesible',
  horchenLaeuft: 'Escuchando en {adresse}',
  horchenNamen: 'Nombre (sobrevive a un cambio de dirección)',
  horchenAdressen: 'Dirección',
  horchenNurHier:
    'Con esta configuración nadie te alcanza desde fuera: 127.0.0.1 escucha solo en este equipo. Indica solo el puerto (:41234) para que este Mac sea accesible en su red.',
  horchenNurDiesesNetz:
    'Ambas valen en esta red. Una dirección asignada temporalmente puede cambiar, un nombre normalmente sobrevive a eso, y una dirección que valga en todas partes solo llega con conexiones entre redes.',
  horchenStill: 'Este perfil no está escuchando ahora mismo.',
  horchenWeitergeben:
    'Esto es lo que necesita la otra parte, junto con la clave de este perfil.',
  sichtbarTitel: 'Ser descubierto en las cercanías',
  sichtbarErklaerung:
    'Anunciarse le dice a todos los de la misma red que este perfil ' +
    'existe, incluido su ID de peer. Ser accesible y ser descubierto son ' +
    'dos decisiones, y esta es la segunda.',
  sichtbarName: 'Nombre con el que anunciarse',
  sichtbarNameHinweis:
    'Un nombre elegido por uno mismo no prueba nada. El nombre de equipo de este Mac no sale.',
  sichtbarWerden: 'Anunciarse',
  sichtbarBeenden: 'Dejar de anunciarse',
  sichtbarLaeuft: 'Este perfil se está anunciando.',
  sichtbarStill: 'Este perfil no se está anunciando.',
  sichtbarBrauchtHorchen:
    'Hazte accesible primero: anunciarse sin un puerto sería una ' +
    'invitación a una puerta sin nadie detrás.',
  nachbarnTitel: 'Quién hay cerca',
  nachbarnErklaerung:
    'Peers que se anuncian ahora mismo en la misma red. Lo que ves es lo ' +
    'que ellos afirman; solo se comprueba al conectar.',
  nachbarnSuchen: 'Echar un vistazo',
  nachbarnSucht: 'Buscando …',
  nachbarnLeer: 'Nadie se está anunciando ahora mismo.',
  nachbarnUnbekannt: 'Desconocido',
  nachbarnAufnehmen: 'Añadir como contacto',
  nachbarnGemerkt: 'Recordado como',
  nachbarnSchonBekannt: 'Ya está en la lista',
  // Ein bekannter Peer an einer neuen Adresse (NET-02.NAT-02, #1706).
  nachbarnAdresseNeu: 'Este peer está en una dirección nueva.',
  nachbarnAdresseBisher: 'Recordada hasta ahora',
  nachbarnAdresseKeine: 'ninguna',
  nachbarnAdresseUebernehmen: 'Usar la nueva dirección',
  nachbarnAdresseHinweis:
    'Esto solo cambia a dónde se conecta. No se confía en él más que antes, y su clave se comprueba en cada conexión.',
  nachbarnVerworfen:
    '{zahl} respuestas no llevaban ID de peer y no están en la lista.',

  // **How this program goes online** — N5
  // (#1571).
  netzwegTitel: 'Del entorno de este proceso',
  netzwegErklaerung:
    'Estos valores proceden del entorno en el que se inició el programa. Prevalecen sobre los campos de abajo, y ninguna ventana los cambia.',
  netzwegProxyHTTP: 'Proxy para http',
  netzwegProxyHTTPS: 'Proxy para https',
  netzwegProxyAus: 'Sin proxy para',
  netzwegAusUmgebung: '{name}: {wert}',

  // **The two switchable services** — N6
  // (#1572).
  dienstRESTTitel: 'La interfaz para scripts',
  dienstRESTErklaerung:
    'Un servicio web sobre la misma base de datos, para leer y modificar desde un script. Solo se ejecuta mientras lo permitas y termina con esta ventana.',
  dienstMCPTitel: 'El servidor de herramientas por HTTP',
  dienstMCPErklaerung:
    'Las mismas herramientas que por la entrada estándar, aquí bajo una dirección, accesible también desde otro dispositivo.',
  dienstRESTAdresse: 'Dirección de la interfaz para scripts',
  dienstRESTPort: 'Puerto de la interfaz para scripts',
  dienstMCPAdresse: 'Dirección del servidor de herramientas',
  dienstMCPPort: 'Puerto del servidor de herramientas',
  dienstMarke: 'Acceso',
  dienstMarkeKopieren: 'Copiar acceso',
  dienstFingerabdruck: 'Huella',
  dienstFingerabdruckKopieren: 'Copiar huella',
  dienstKopiert: 'Copiado',
  dienstStarten: 'Iniciar',
  dienstBeenden: 'Detener',
  dienstErreichbar: 'Accesible en:',
  dienstMitTLS:
    'Cifrado. El certificado es autoemitido y lleva el ID de este perfil; quien se conecte compara su huella.',
  dienstOhneTLS:
    'Sin cifrar, y aquí no importa: la dirección solo vale en este equipo.',
  dienstHinweis:
    '127.0.0.1 significa: solo este equipo. 0.0.0.0 significa: todos los dispositivos de la red; entonces el cifrado y el acceso son obligatorios, y ambos se configuran solos. Puerto 0 significa: buscar uno libre.',
}
