// **The “Users and profiles” page** — phase 4.12
// (#1351).
import { faktor, konto } from './seitenkonto'
import { sitzung } from './seitensitzung'

export const nutzer = {
  titel: 'Usuarios y perfiles',

  laschen: {
    allgemein: 'General',
    konto: 'Cuenta',
    profil: 'Perfiles',
    persoenlich: 'Personal',
  },

  allgemeinErklaerung:
    'Un usuario es la persona. Un perfil es un entorno de trabajo de esa persona: sus propios sets, su propia colección, sus propios ajustes. Quién trabaja ahora y dónde aparece en la barra lateral, bajo el nombre de la aplicación; el cambio se hace aquí.',

  kontoErklaerung:
    'Con qué inicia sesión este usuario. Mientras haya exactamente un usuario con un perfil, no se pregunta nada.',

  profilErklaerung:
    'Quien quiera coleccionar por separado para un torneo necesita un segundo perfil, no una segunda cuenta.',

  niemand:
    'Aún no hay ningún otro usuario. El primero se creó en el primer inicio.',
  keineProfile: 'Este usuario todavía no tiene ningún perfil.',

  neuerNutzer: 'Crear un usuario',
  neuerNutzerName: 'Nombre visible del usuario nuevo',
  neuerNutzerNameHinweis:
    'Único y sin espacios en los extremos: distingue a dos usuarios',
  neuerNutzerKnopf: 'Crear usuario',
  nutzerAngelegt: 'Creado: {name}',

  neuesProfil: 'Crear otro perfil',
  neuesProfilNutzer: 'Para qué usuario',
  neuesProfilName: 'Nombre del perfil',
  neuesProfilNameHinweis:
    'Único por usuario: dos personas pueden tener cada una un «Torneo»',
  neuesProfilKnopf: 'Crear perfil',
  profilAnlegenOeffnen: 'Crear perfil …',
  profilAngelegt: 'Creado: {name}',

  wahl: 'En qué perfil se trabaja',
  wahlErklaerung:
    'Los sets, mazos, la colección, los torneos y los ajustes pertenecen al perfil elegido. Cambiar no borra nada: lo que hay aquí volverá al regresar.',
  hier: 'Estás trabajando aquí',
  hierhin: 'Cambiar a este',
  hierhinFuer: 'Cambiar a {nutzer} · {profil}',

  nutzerwahl: 'Quién trabaja',
  nutzerwahlErklaerung:
    'Mirar el perfil de otra persona no cambia quién eres: tus propios permisos deciden adónde puedes ir. Esto no es lo mismo que el perfil de arriba. Para convertirte en otra persona demuestras su segundo factor; después trabajas en su primer perfil.',
  dasBinIch: 'Eres tú ahora mismo',
  werden: 'Convertirme en esta persona',
  werdenFuer: 'Convertirme en {nutzer}',

  anmeldung: 'Inicio de sesión con un proveedor',
  anmeldungErklaerung:
    'Para iniciar sesión con un proveedor, introduce su `client_id` en **Ajustes › Usuario y perfil › Inicio de sesión con un proveedor**. MagicSetEditor4Mac no incluye credenciales; la aplicación OAuth propia se crea en el proveedor, y un secreto, si lo exige, va al llavero y no a la biblioteca.',
  anmelden: 'Iniciar sesión',
  anbieterstand: '{eingerichtet}, {geprueft}',
  eingerichtet: 'configurado',
  nichtEingerichtet: 'sin configurar',
  geprueft: {
    simuliert: 'probado solo contra una simulación',
    realGeprueft: 'probado contra el proveedor real',
  },
  anmeldenBei: 'Iniciar sesión con {anbieter}',
  anmeldungAbbrechen: 'Cancelar',
  anmeldungAbbrechenBei: 'Cancelar el inicio de sesión con {anbieter}',
  verbunden: 'Conectado',
  verbundenAls: 'Conectado como {name}',
  loesen: 'Desconectar',
  loesenVon: 'Desconectar de {anbieter}',

  anbieternamen: {
    github: 'GitHub',
    google: 'Google',
    microsoft: 'Microsoft',
    x: 'X',
  } as Record<string, string>,

  pflege: 'Editar usuarios y perfiles',
  pflegeErklaerung:
    'Un nombre se puede cambiar. Para borrar, el nombre debe quedarse como está: confirma a quién te refieres.',
  bearbeiten: 'Editar {name}',
  bearbeitenKurz: 'Editar',
  neuerNameFuer: 'Nombre',
  umbenennen: 'Renombrar',
  loeschen: 'Eliminar',
  loeschenBestaetigt: 'Eliminar {name} definitivamente',
  loeschenFrage: '¿Eliminar {name} definitivamente?',
  loeschenJa: 'Eliminar definitivamente',
  abbrechen: 'Cancelar',
  nutzerLoeschenHinweis:
    'Un usuario eliminado se lleva sus perfiles. El último usuario no se puede eliminar.',
  profilLoeschenHinweis:
    'El último perfil de un usuario no se puede eliminar. Lo que hay en el perfil se queda en la biblioteca, pero ya no se puede alcanzar.',

  angaben: 'Datos personales',
  angabenErklaerung:
    'Todo salvo el nombre visible es opcional. El **tratamiento** es un campo libre: cualquier lista dejaría a alguien fuera, y una aplicación que obliga a elegir una entrada equivocada se ha dirigido mal a la persona. La dirección, el número de teléfono y la contraseña están en **Cuenta**.',
  angabenSichern: 'Guardar datos',
  felder: {
    name: 'Nombre visible',
    anrede: 'Tratamiento',
    namenszusatz: 'Complemento del nombre',
    vorname: 'Nombre',
    weiterenamen: 'Otros nombres',
    nachname: 'Apellidos',
    slogan: 'Eslogan',
    kurzebio: 'Biografía breve',
    langebio: 'Biografía larga',
  },

  // **Die Profil-Einstellungen** — Werkstück 9 von Phase 4.12 (#1351).
  badge: 'Insignia del perfil',
  badgeVon: 'Insignia del perfil de {name}',
  badgeWaehlen: 'Elegir insignia …',
  badgeEntfernen: 'Quitar insignia',
  profilangaben: 'Datos del perfil',
  profilangabenErklaerung:
    'Pertenecen a **este perfil**, no a ti: quien tiene varios los distingue aquí. El **eslogan secundario** va bajo tu propio eslogan y no lo sustituye. La **insignia** está abajo a la derecha, sobre tu imagen de perfil: muestra el rol con el que trabajas ahora.',
  profilangabenSichern: 'Guardar perfil',
  profilfelder: {
    name: 'Nombre del perfil',
    subslogan: 'Eslogan secundario',
    kurzebeschreibung: 'Descripción breve',
    langebeschreibung: 'Descripción larga',
  },

  // **Technische Profile und ihre Zugriffstoken** — Werkstück 10 von
  // Phase 4.12 (#1351), Punkte 12 und 13.
  technisch: 'Perfil técnico: detrás no hay ninguna persona',
  technischErklaerung:
    'Un perfil técnico separa lo que un servicio escribe por su cuenta de lo que haces tú mismo. Quien deja que un servicio trabaje en su nombre no quiere que escriba en **su** perfil: lo que aparece allí no lo ha hecho su mano. Aun así te pertenece, porque alguien tiene que poder desactivarlo.',
  ablauf: 'Caduca el (vacío: nunca)',
  ablaufErklaerung:
    'Un momento a partir del cual este perfil ha caducado; **no es una orden de borrado**. Nada desaparece por sí solo.',

  token: 'Tokens de acceso',
  tokenErklaerung:
    'Así se identifica un proceso automático en el acceso REST, en **este** perfil y con el permiso que se indica aquí. La ventana, la línea de comandos y MCP no piden ningún token: allí ya se sabe quién trabaja.',
  tokenLeer: 'Este perfil aún no tiene ningún token de acceso.',
  tokenBezeichnung: 'Para qué sirve',
  tokenScope: 'Permiso',
  tokenLesen: 'Solo lectura',
  tokenSchreiben: 'Lectura y escritura',
  tokenAblauf: 'Caduca el (vacío: nunca)',
  tokenAnlegen: 'Crear token',
  tokenAendern: 'Aplicar',
  tokenLoeschen: 'Revocar',
  tokenLoeschenFrage: '¿Revocar el token «{name}»?',
  tokenLoeschenErklaerung:
    'Quien lo use se quedará fuera. Habrá que introducir allí un token nuevo.',
  tokenLoeschenJa: 'Revocar',
  tokenAbgelaufen: 'caducado',
  tokenUnbefristet: 'no caduca',

  tokenEinmalig:
    'El valor en texto plano aparece **exactamente una vez**, justo después de crearlo. Después no existe en ningún sitio; quien lo pierda crea un token nuevo.',
  tokenKlartext: 'Cópialo ahora: este valor no volverá a aparecer:',

  // **Rotation, bulk revocation, expiry presets and MCP client set-up** —
  // piece O2 of phase 5.2 (#1426).
  tokenPrincipal: 'Para quién trabaja',
  tokenAblaufwahl: 'Caducidad',
  tokenAblaufwahlen: {
    tag: 'Un día',
    monat: 'Treinta días',
    jahr: 'Un año',
    nie: 'Nunca',
    eigen: 'Momento personalizado',
  },
  tokenRotieren: 'Secreto nuevo',
  tokenRotierenFrage: '¿Generar un secreto nuevo?',
  tokenRotierenErklaerung:
    'El anterior sigue siendo válido una hora más para que un trabajo en curso pueda terminar. El nuevo valor en texto plano aparece exactamente una vez.',
  tokenAlleWiderrufen: 'Revocar todos',
  tokenAlleWiderrufenFrage:
    '¿Revocar todos los tokens de acceso de este perfil?',
  tokenAlleWiderrufenErklaerung:
    'Todo proceso automático que use alguno de ellos se quedará fuera, incluso en mitad de una sesión. Lo que haya creado se conserva.',
  tokenGezogen: '{zahl} tokens revocados.',
  tokenLetzteNutzung: 'usado por última vez {zeit}',
  tokenNieBenutzt: 'nunca usado',
  tokenVorgaengerBis: 'el secreto anterior es válido hasta {zeit}',
  tokenOnboarding: 'Cómo configurar un cliente MCP',
  tokenOnboardingSchritt1:
    'Crea un token y copia el texto plano: aparece una sola vez.',
  tokenOnboardingSchritt2:
    'Escríbelo en un archivo que solo tú puedas leer (permisos 0600). Un archivo que otros puedan leer lo rechaza «mse-mcp».',
  tokenOnboardingSchritt3:
    'Inicia «mse-mcp» con MSE_TOKEN_DATEI apuntando a ese archivo, como variable de entorno y nunca como argumento: aparecería en la lista de procesos.',
  tokenOnboardingSchritt4:
    'El token se comprueba en cada llamada a una herramienta: una revocación, una caducidad o un permiso menor surten efecto en la siguiente llamada. Sin token, «mse-mcp» funciona como hasta ahora, sin identificación, en el perfil mostrado.',
  tokenOnboardingBeispiel:
    '{ "mcpServers": { "mse": { "command": "/Applications/MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp", "env": { "MSE_TOKEN_DATEI": "/Users/ich/.mse/mcp-token" } } } }',
  tokenOnboardingBeispielTitel: 'Ejemplo de configuración',

  // **Vorgaben, die auf Zeilen zeigen** — Werkstück 15 von Phase 4.12
  // (#1351), Punkte 4 bis 7.
  vorgaben: 'Valores predeterminados de este perfil',
  vorgabenErklaerung:
    'Con qué empieza un registro nuevo en **este perfil**: un formato, un juego, una plantilla de la biblioteca. A diferencia de los ajustes, los valores predeterminados apuntan a **filas de la biblioteca**; lo que se puede elegir depende de lo que haya.',
  vorgabenamen: {
    format: 'Formato predeterminado para mazos nuevos',
    spiel: 'Juego predeterminado para sets nuevos',
    vorlage: 'Plantilla predeterminada para sets nuevos',
    deck: 'Mazos preferidos',
    lagerort: 'Lugares de almacenamiento asignados',
  } as Record<string, string>,
  vorgabenOhne: 'sin valor predeterminado',
  vorgabenLeer: 'Todavía no hay nada para esto en la biblioteca.',
  vorgabenVerwaist: 'No apunta a nada, porque ya no existe:',

  // **Rechte, Rollen und Sichtbarkeit** — Werkstück 11 von Phase 4.12
  // (#1351).
  freigaben: 'Quién puede trabajar en este perfil',
  freigabenErklaerung:
    'Un **lector** solo ve que este perfil existe, no lo que contiene. Un **colaborador** entra en él y trabaja allí. Un **administrador** concede permisos por su cuenta. Este perfil es tuyo si lo has creado; entonces lo administras de todos modos.',
  freigabenNurLesen:
    'Solo puede conceder permisos quien administra este perfil.',
  freigabenLeer:
    'Hasta ahora nadie salvo el propietario puede entrar en este perfil.',
  freigabeNutzer: 'A quién',
  freigabeNiemand: 'a nadie',
  freigabeRolle: 'Con qué rol',
  freigabeVergeben: 'Conceder',
  freigabeZurueckziehen: 'Retirar',
  rollennamen: {
    leser: 'Lector: ve que el perfil existe',
    mitarbeiter: 'Colaborador: entra y trabaja allí',
    verwalter: 'Administrador: concede permisos por su cuenta',
  } as Record<string, string>,

  bild: 'Imagen de perfil',
  bildVon: 'Imagen de perfil de {name}',
  bildWaehlen: 'Elegir imagen …',
  bildEntfernen: 'Quitar imagen',
  ohneBild: '{name} no tiene imagen de perfil',

  nochNicht:
    'Los datos están separados entre perfiles, siempre que nadie abra el archivo de la biblioteca con una herramienta ajena. Contra eso ayuda cifrar la biblioteca en «Ajustes › Datos y copias de seguridad».',

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

  persoenlich: 'Mi cuenta',

  persoenlichErklaerung:
    'Estos ajustes te pertenecen a **ti**, no a este perfil: te los llevas a todos tus perfiles. Los dos de arriba prevalecen sobre el sistema operativo, y **solo en un sentido**: activan, nunca desactivan.',

  auskunft: 'Mis datos',

  auskunftErklaerung:
    'Lo que este programa ha guardado sobre ti: datos personales, formas de inicio de sesión, los **tipos** de segundos factores, ajustes y nombres de perfil. **Sin contraseña y sin secreto**: ninguno de los dos es un dato sobre ti, sino un medio para ser tú. Los sets, mazos y la colección no están incluidos: pertenecen a un perfil y salen por sus propios caminos.',

  auskunftSichern: 'Guardar mis datos …',

  datenschutz: 'Lo que esta aplicación sabe de ti',

  datenschutzHinweis:
    'Qué datos se registran, dónde se guardan, qué sale de este ordenador y qué expresamente no. Se describe en **Ayuda → Lo que esta aplicación sabe de ti**.',
} as const
