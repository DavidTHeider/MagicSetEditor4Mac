// The **Decks** page — Werkstück 7 of phase 4.6
// (#1293).
export const decks = {
  titel: 'Mazos',
  erklaerung:
    'Composiciones hechas con cartas propias e impresiones reales. Un mazo hace referencia a las cartas en lugar de copiarlas: una corrección en el set de origen le llega al instante.',

  leer: 'Todavía no hay ningún mazo. «Crear mazo», arriba, empieza uno. Un mazo pertenece a un juego; sin él no habría formato con el que comprobarlo.',

  neu: 'Crear mazo',
  name: 'Nombre',
  spiel: 'Juego',
  notiz: 'Nota',
  anlegen: 'Crear',
  ausSetTitel: 'O desde un set',
  ausSetErklaerung:
    'Cada carta del set pasa a ser una entrada del mazo principal — referenciada, no copiada: si la corriges en el set, queda corregida en el mazo.',
  ausSetSet: 'Set',
  ausSetMenge: 'Copias por carta',
  ausSetAnlegen: 'Crear a partir del set',
  ausSetKeinSet:
    'Todavía no hay ningún set con cartas — un mazo a partir de un set necesita uno.',

  hauptdeck: 'Mazo principal',
  sideboard: 'Banquillo',
  zaehlung: '{haupt} en el mazo principal, {side} en el banquillo',
  waehlen: 'Abrir',
  offenMarke: 'abierto',
  umbenennen: 'Renombrar',
  neuerName: 'Nombre nuevo',
  umbenennenTitel: 'Renombrar {name}',
  nameSichern: 'Guardar',
  umbenennenAbbrechen: 'Conservar el nombre',
  duplizieren: 'Duplicar',
  kopieTitel: 'Duplicar {name}',
  kopieName: 'Nombre de la copia',
  kopieVorschlag: '{name} (copia)',
  kopieAnlegen: 'Crear copia',
  kopieAbbrechen: 'No duplicar',
  aufWunschliste: 'Añadir lo que falta a la lista de deseos',
  wunschlisteErgebnis:
    'Nuevas en la lista de deseos: {neu} · ya estaban: {schon}',
  loeschen: 'Eliminar mazo',
  handlungenAn: 'Acciones — {name}',
  handlungenKurz: '…',
  deckmenue: 'Mazo',
  loeschenFrage: '¿Eliminar este mazo? Las cartas en sí se conservan.',
  loeschenJa: 'Eliminar',
  loeschenNein: 'Cancelar',

  plaetze: 'Huecos',
  plaetzeLeer: 'Todavía no hay ninguna carta en el mazo.',
  bezug: 'Carta o impresión',
  menge: 'Cantidad',
  bereich: 'Sección',
  setWaehlen: 'Set',
  karteWaehlen: 'Carta',
  karteSuchen: 'Buscar carta',
  karteKeine: 'Ninguna carta coincide con «{suche}».',
  setLeer:
    'No hay ningún set abierto: solo se podrán elegir cartas de tus propios sets cuando haya uno.',
  ohneSet: 'Elige un set …',
  ohneKarte: 'Elige una carta …',
  setOhneKarten: 'Este set todavía no tiene ninguna carta.',
  kurve: 'Curva de maná',
  kurveHinweis:
    'Cuántas cartas hay por valor de maná. Se cuenta todo aquel cuyo valor se conoce: en una impresión de Scryfall, el que allí figura; en las demás, el calculado a partir del coste. Las tierras no tienen coste y por eso no aparecen, como tampoco una carta propia con el campo de coste vacío.',
  farben: 'Colores',
  farbeWeiss: 'Blanco',
  farbeBlau: 'Azul',
  farbeSchwarz: 'Negro',
  farbeRot: 'Rojo',
  farbeGruen: 'Verde',
  farbenHinweis:
    'Una carta bicolor cuenta en ambos colores, por lo que la suma es mayor que el mazo.',
  karten: 'Cartas',
  gebietFormate: 'Formatos',
  formateKnopf: 'Formatos …',

  // **Taking a deck from someone else's profile** — Werkstück 12 of Phase
  // 4.12 (#1351).
  uebernahme: 'Traer un mazo de otro perfil',
  auswahlLoesen: 'Mostrar todos los mazos',
  leiste: {
    bezeichnung: 'Cifras del mazo',
    sideboard: 'Banquillo {zahl}',
    pruefung: 'Comprobación',
    ohneFormat: 'sin formato',
    nichtGeprueft: 'sin comprobar',
    fehlt: 'Falta en la colección',
    unbekannt: '{zahl} desconocidas',
    nichtAbgeglichen: 'sin comparar',
  },
  uebernahmeErklaerung:
    'Puede traerlo quien pueda cambiar a ese otro perfil. Si su propietario usa un segundo factor, hace falta allí un permiso compartido; basta con «lector». No vienen: los huecos que apuntan a una carta propia de ese perfil y el formato, porque ambos son filas de allí. Lo que se quedó atrás se enumera aquí después.',
  uebernahmeProfil: 'Identificador del otro perfil',
  uebernahmeDeck: 'Identificador del mazo',
  uebernahmeProfilwahl: '¿De qué perfil?',
  uebernahmeProfilWaehlen: 'Elige un perfil …',
  uebernahmeProfilEintrag: '{nutzer} · {profil}',
  uebernahmeKeinProfil: 'No hay ningún otro perfil en este equipo.',
  uebernahmeDeckwahl: '¿Qué mazo?',
  uebernahmeDeckWaehlen: 'Elige un mazo …',
  uebernahmeKeinDeck: 'No hay ningún mazo en este perfil.',
  uebernahmeFachleute: 'Indicar tú mismo los identificadores',
  uebernahmeFachleuteErklaerung:
    'El perfil y el mazo también se pueden indicar por sus identificadores, tal como los muestran la línea de comandos y el servicio. El permiso es el mismo.',
  uebernahmeHolen: 'Traer',
  uebernahmeGeschrieben: '«{name}» ya está entre tus mazos — {karten}.',
  uebernahmeNichtMit: 'No vinieron:',

  gebietKarten: 'Cartas',
  gebietAuswertung: 'Análisis',
  gebietPruefung: 'Comprobación',
  gebieteDeck: 'Áreas de este mazo',
  legen: 'Añadir al mazo',
  nehmen: 'Quitar',
  druck: 'Impresión (identificador de Scryfall)',
  anschlussOhneDeck:
    '«{name}» está esperando: abre un mazo o crea uno y aparecerá en su selector de cartas.',
  holenAus: 'Traer cartas de',
  holenSammlung: 'Mi colección',
  holenSet: 'Un set',
  wegeTitel: 'Traer y pasar',
  mitspielen: 'Jugar con este mazo',
  anschlussWartet:
    '«{name}» ha llegado desde tu colección: elige cantidad y sección y pasará al mazo.',

  format: 'Formato',
  formatWaehlen: 'Elegir formato',
  ohneFormat: 'Sin formato — sin comprobar',
  formatHinweis:
    'Un mazo sin formato está sin comprobar, no mal. El formato tiene que pertenecer al juego del mazo.',

  pruefen: 'Comprobar con el formato',
  zulaessig: 'Legal.',
  unzulaessig: 'No legal.',
  ungeprueftZahl: '{zahl} cartas sin comprobar.',
  ungepruefterklaerung:
    'Sin comprobar no significa ilegal: una carta que diseñaste tú no figura en ningún formato del mundo. Tampoco «restringida» va en contra del mazo: cuántas veces puede aparecer una carta restringida lo dicen las reglas de torneo, y este programa no las ha leído.',

  legal: 'legal',
  not_legal: 'no listada',
  restricted: 'restringida',
  banned: 'prohibida',
  ungeprüft: 'sin comprobar',

  bestand: 'Lo que tienes',
  vorhanden: 'Disponibles',
  gebraucht: 'Necesarias',
  fehlend: 'Faltan',
  unbekannt: 'Desconocido',
  bestandVollstaendig: 'Está todo.',
  bestandFehlt: 'Faltan {zahl} cartas.',

  // **Die Zuordnung** — Werkstück U9 von Phase 5.1 (#1420).
  gebietZuordnung: 'Asignación',
  zuordnungen: 'Necesidad y asignación',
  zuordnungenLaden: 'Cargando …',
  zuordnungenLeer:
    'Todavía no hay ninguna carta en el mazo. Lo que aparece aquí sale de sus huecos.',
  zuordnungHinweis:
    'Una lista de mazo no reserva nada: dos mazos con las mismas tres cartas están ambos completos. Asignado es lo que apartas expresamente para este mazo; libre es lo que queda después.',
  zuordnungszahlen:
    '{vorhanden} de {gebraucht} disponibles · {frei} libres · {zugeordnet} asignadas · {reserviert} reservadas',
  zuordnungPlatz: 'Sección',
  zuordnungExemplar: 'Ejemplar',
  zuordnungExemplarWaehlen: 'Elige un ejemplar',
  zuordnungMenge: 'Cantidad',
  zuordnen: 'Asignar',
  zuordnungLoesen: 'Liberar',
  abweichend: 'otra impresión',
  abweichendErklaerung:
    'Este ejemplar es una impresión distinta de la del hueco. Elegido expresamente así: el mismo nombre no es la misma carta.',
  fehlkarten: 'Cartas que faltan, como lista',
  fehlkartenliste: 'Las cartas que faltan',
  fehlkartenKeine: 'No falta nada.',

  typen: 'Tipos',
  kennzahlen: 'Cifras clave',
  laender: 'Tierras',
  durchschnitt: 'Valor de maná medio',
  keineZahl: '—',
  quellen: 'Fuentes de color',
  quellenHinweis:
    'Se cuentan las tierras con uno de los cinco tipos básicos: producen su maná sin texto de reglas (CR 305.6). Una tierra que nombra su color en el texto no aparece aquí: interpretar ese texto sería adivinar.',

  weniger: '−',
  wenigerVon: 'Una menos',
  mehr: '+',
  mehrVon: 'Una más',
  insSideboard: 'Al banquillo',
  insHauptdeck: 'Al mazo principal',
  suche: 'Buscar cartas',
  deckbereich: 'El mazo',
  analyse: 'Análisis',
  analyseLeer:
    'Aquí aparecen el análisis, la comprobación y la asignación del mazo abierto.',
  stapelTitel: 'De la colección',
  stapelWartet: '{name} de la colección, aún en ningún mazo: {liste}.',
  stapelLegen: 'Añadir todas al mazo',
  stapelTeilweise: '{gelegt} de {gesamt} están en el mazo; el resto, no:',
  stapelOhneDruck:
    '«{name}» no tiene identificador de Scryfall: un mazo solo admite impresiones que lo tengan.',
  stapelGrund: '«{name}»: {grund}',
  eintauschen: 'Intercambiar las {anzahl}',
  deckplatz: 'Taller de mazos',
  pruefungOhneFormat:
    'Sin formato no se comprueba nada. Elige un formato para el mazo; hasta entonces está sin comprobar, no mal.',
  pruefungLaeuft: 'Comprobando …',
  postenUnzulaessig: 'no legal',
  postenUnzulaessigErklaerung:
    'Esta carta no está permitida en el formato elegido. La pestaña «Comprobación» dice qué exactamente.',
  keinDeckOffen:
    'No hay ningún mazo abierto. Elige uno a la izquierda o crea uno nuevo.',

  formate: 'Formatos',
  formateLeer:
    'Todavía no hay ningún formato. Aquí un formato es un registro y no una lista incorporada: se crea, no viene de serie.',
  bauart: 'Construcción',
  bauarten: 'Construcciones documentadas',
  bauartHinweis:
    'Las cifras proceden de las Comprehensive Rules. Son una sugerencia; después se puede cambiar cada una.',
  scryfallname: 'Nombre en Scryfall',
  scryfallnameHinweis:
    'La clave en las legalities de Scryfall, por ejemplo «standard». Vacío significa que Scryfall no conoce este formato y todas las cartas quedan sin comprobar.',
  mindestgroesse: 'Como mínimo',
  hoechstgroesse: 'Como máximo',
  sideboardgroesse: 'Banquillo',
  sideboardOffen: 'Banquillo sin límite',
  hoechstzahlGleich: 'Misma carta',
  grenzenHinweis:
    'Cero significa sin límite. En el banquillo, cero significa que no hay banquillo.',
  formatNeu: 'Crear formato',
  formatSichern: 'Guardar formato',
  formatLoeschen: 'Eliminar formato',
  formatLoeschenFrage: '¿Eliminar {name}?',
  formatLoeschenJa: 'Eliminar',
  formatLoeschenHinweis:
    'Los mazos que lo usaban se conservan; después quedan sin comprobar.',
  umschlagTitel: 'Entregar como sobre',
  umschlagErklaerung:
    'Un sobre lleva un mazo a otro ordenador — firmado, con suma de comprobación y sin cuenta. Lo que llega se convierte en un mazo nuevo; no se sobrescribe nada. El nombre, el juego y cada entrada con una impresión **viajan con él**; la nota, el origen y las entradas con cartas propias **se quedan aquí** — su identificador solo vale en esta biblioteca.',
  umschlagSchnueren: 'Guardar el mazo abierto',
  umschlagAnnehmen: 'Mazo desde un sobre',
  umschlagOhneDeck: 'Abre primero un mazo.',
  umschlagGeschnuert: 'Guardado.',
  umschlagAngenommen: 'Importado',
  einfuhrTitel: 'Importar una lista de mazo',
  einfuhrMenue: 'Importar una lista de mazo …',
  einfuhrErklaerung:
    'Pega una lista de mazo, una carta por línea, «4 Lightning Bolt (M11) 149». Los encabezados la dividen: lo que sigue a «Sideboard» es el banquillo. La vista previa no escribe nada; solo al importar se añaden las líneas al mazo abierto, y no se quita nada.',
  einfuhrFeld: 'Lista de mazo',
  einfuhrVorschau: 'Vista previa',
  einfuhrUebernehmen: 'Importar lista de mazo',
  einfuhrLeer: 'Pega primero una lista.',
  einfuhrOffen: 'pendiente',
  einfuhrBilanz: 'Mazo principal / banquillo',
  einfuhrGelegt: 'Importadas',
  einfuhrStueck: 'copias',
  einfuhrNichtsOffen: 'Ninguna línea es inequívoca: no hay nada que añadir.',
} as const
