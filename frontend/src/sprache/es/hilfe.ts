import { betrieb } from './hilfebetrieb'
import { suche } from './hilfesuche'

export const hilfe = {
  betrieb,
  suche,
  // One help entry per page — the keys come from `navigation.ts`.
  seiten: {
    start: {
      ueberschrift: 'Inicio',
      text: 'La puerta de entrada. Arriba, las cifras de la biblioteca (sets, cartas, paquetes, ejemplares) y debajo, aquello en lo que trabajaste por última vez: abierto, importado, exportado, y lo que mantienes fijado arriba. Después, lo que necesita atención (sets con reservas, mazos a los que les faltan cartas, paquetes con hallazgos en scripts) y lo que se está descargando en segundo plano. Los tres primeros pasos aparecen como una ventana en el primer inicio; los recorridos de la ayuda los muestran de nuevo.',
    },
    sets: {
      ueberschrift: 'Sets',
      text: 'Todos los sets que conoce esta aplicación. Desde aquí se abre uno, se crea uno nuevo o se importa un archivo `.mse-set`. El set abierto decide qué muestran sus pestañas y los cajones «Propiedades» y «Herramientas».',
    },
    pakete: {
      ueberschrift: 'Paquetes',
      text: 'Juegos, hojas de estilo, fuentes de símbolos y plantillas de exportación. Se descargan aquí (en una revisión fija y con suma de comprobación) y se instalan. Sin ellos, una carta no tiene campos ni aspecto.',
    },
    vorlagen: {
      ueberschrift: 'Plantillas',
      text: 'Los archivos de texto de un paquete de plantillas: campos, scripts, medidas. Aquí se pueden modificar sin necesidad de un editor de texto aparte. Un cambio vale para todos los sets que usan este paquete; el estado anterior se conserva como copia de seguridad junto al paquete.',
    },
    sammlung: {
      ueberschrift: 'Colección',
      text: 'Lo que realmente posees: una fila por impresión, estado, idioma y acabado. No tiene nada que ver con tus sets: una carta diseñada es un borrador, un ejemplar es un trozo de cartón. Los precios se descargan bajo petición y son estimaciones de las que nadie responde.',
    },
    decks: {
      ueberschrift: 'Mazos',
      text: 'Composiciones hechas con cartas propias e impresiones reales. Un mazo hace referencia a cartas en lugar de copiarlas y pertenece a un juego; sin él no habría formato. Aquí un formato es un registro y no una lista incorporada: se crea, no viene de serie. La comprobación consulta a Scryfall las impresiones reales; una carta que diseñaste tú queda sin comprobar, y eso no significa que sea ilegal.',
    },
    turniere: {
      ueberschrift: 'Torneos',
      text: 'Un torneo es un desarrollo a lo largo del tiempo: rondas, emparejamientos, resultados. La clasificación se calcula a partir de ahí y no se guarda en ningún sitio: si corriges un resultado, la clasificación lo refleja al instante. Los emparejamientos siguen el sistema suizo; los cuatro criterios de desempate y sus cifras proceden de las Magic Tournament Rules. Inscribir un mazo entrega una lista: se copia, y los cambios posteriores del mazo ya no la alcanzan.',
    },
    spielen: {
      ueberschrift: 'Jugar',
      text: 'Una partida es una mesa con mazos congelados: lo que un mazo pasa a ser al empezar no cambia si alguien lo edita después. Lo que ocurre en la mesa es un flujo de eventos: cada uno depende del anterior y ninguno se sobrescribe. Lo que se ve es siempre la vista de **un** asiento: una biblioteca está boca abajo incluso para su propietario, y quien la consulta queda debiendo un barajado. El anfitrión guarda el estado completo; la página lo dice porque conviene saberlo antes de unirse.',
    },
    handeln: {
      ueberschrift: 'Intercambio',
      text: 'Se intercambia con una contraparte cuya huella has comparado, sin cuenta y sin intermediarios. La comparación se calcula en **este** equipo: no sale nada de él, y lo que una contraparte ve de ti lo decides tú en «Ajustes › Conexiones». Solo se contabiliza cuando las cartas han cambiado realmente de manos; entonces cambia tu biblioteca y el diario anota qué fue adónde. **Contabilizado aquí no es contabilizado en ambos lados**: dos bibliotecas en dos equipos no son un asiento contable compartido, y el estado lo dice. Lo que vale tu colección aparece al lado como observación, no como promesa.',
    },
    hilfe: {
      ueberschrift: 'Ayuda',
      text: 'Los términos que proceden de Magic Set Editor y que nadie trae sabidos, cada uno en una frase. Además, los recorridos, los atajos de teclado, por qué algunas etiquetas siguen en inglés y qué envía esta aplicación al exterior.',
    },
    nutzer: {
      ueberschrift: 'Usuarios y perfiles',
      text: 'Aquí se guarda a quién pertenece cada cosa. Un usuario es la persona; un perfil es un entorno de trabajo de esa persona, con sus propios sets, su propia colección y sus propios ajustes. Quien quiera coleccionar por separado para un torneo no necesita una segunda cuenta, sino un segundo perfil.',
    },
    einstellungen: {
      ueberschrift: 'Ajustes',
      text: 'Idioma, apariencia y los lugares donde se buscan los paquetes. Lo que figura aquí vale para todo el programa y no para un set concreto.',
    },
    karten: {
      ueberschrift: 'Cartas',
      text: 'Las cartas del set abierto: a la izquierda la lista, en el centro la vista previa, a la derecha los campos de la carta abierta. Un cambio en un campo se ve al instante; no hay que confirmar nada.',
    },
    setangaben: {
      ueberschrift: 'Datos del set',
      text: 'Lo que vale para todo el set: nombre, descripción, símbolo del set, autor. Estos datos están disponibles para todas las cartas: una hoja de estilo puede dibujar el símbolo en cada una.',
    },
    stil: {
      ueberschrift: 'Estilo',
      text: 'La hoja de estilo del set y sus opciones. La hoja dice qué aspecto tiene una carta; las opciones la ajustan, por ejemplo a otro marco o a otra edición.',
    },
    suche: {
      ueberschrift: 'Buscar',
      text: 'Encuentra cartas por los valores de sus campos, en todos los sets o en uno solo. La búsqueda se hace en el texto completo y un resultado lleva directamente a la carta.',
    },
    ersetzen: {
      ueberschrift: 'Reemplazar',
      text: 'Cambia una palabra en los campos de texto de un set. Primero muestra lo que cambiaría; un ⌘Z deshace toda la operación.',
    },
    statistik: {
      ueberschrift: 'Estadísticas',
      text: 'Lo que contiene el set abierto: cuántas cartas hay, con qué frecuencia se rellena un campo, cómo se distribuye el valor de un campo. Sirve para comprobar si una serie está completa.',
    },
    bilder: {
      ueberschrift: 'Imágenes',
      text: 'Escribe las cartas del set como archivos de imagen: todas a la vez, con un patrón de nombres y una regla para los nombres repetidos. La imagen de una carta se **asigna** en el campo de imagen de la carta, no aquí.',
    },
    druckbogen: {
      ueberschrift: 'Hoja de impresión',
      text: 'Las cartas una junto a otra sobre papel, como PDF con marcas de corte. Pensada para la impresora y no para compartir; para eso está la exportación.',
    },
    schluesselwoerter: {
      ueberschrift: 'Palabras clave',
      text: 'Una palabra clave reconoce su texto en cada carta y añade detrás el texto recordatorio. Las del juego son fijas y se muestran aquí para consulta; las propias valen solo en este set y se pueden crear, modificar y eliminar.',
    },
    pack: {
      ueberschrift: 'Sobre aleatorio',
      text: 'Un sobre con las cartas de este set, según las reglas del paquete del juego; allí se indica qué tipos de sobre existen. La misma semilla da el mismo reparto; para mostrar un resultado, pasa el número.',
    },
    symbol: {
      ueberschrift: 'Símbolo',
      text: 'El símbolo de este set: dibujado, no cargado. Una forma básica es el punto de partida; después se puede mover cada punto. El tabulador salta entre puntos y las teclas de flecha los mueven. Lo que se dibuja aquí aparece al instante en la vista previa y pasa al archivo cuando se guarda el set.',
    },
    konsole: {
      ueberschrift: 'Consola',
      text: 'Una expresión de MSE-Script en el contexto de la carta abierta: el mismo entorno que al dibujar. No cambia nada; calcula y devuelve. El resultado llega como texto fuente para poder distinguir una cadena de un nombre.',
    },
    ausgeben: {
      ueberschrift: 'Exportar',
      text: 'Cartas como imagen, o como texto mediante una plantilla de exportación. La plantilla procede de un paquete y decide lo que sale: una lista para un foro, una tabla, un archivo para otro programa.',
    },
  },

  // **Über** (#1168).
  ueber: {
    ueberschrift: 'Acerca de MagicSetEditor4Mac',
    fassung: 'Versión',
    lizenzName: 'Licencia',
    lizenzSatz:
      'El código fuente de esta aplicación se distribuye bajo la GNU General Public License, versión 2 o posterior.',
    rechteName: 'Contenido de terceros',
    rechteSatz:
      'La licencia cubre el código fuente, **no** los marcos de las cartas, los símbolos de maná, las fuentes, los textos de reglas, los nombres ni las marcas. Pertenecen a sus titulares de derechos, entre ellos Wizards of the Coast, y no a este proyecto.',
    quelleName: 'Código fuente',
    quelleSatz: 'github.com/DavidTHeider/MagicSetEditor4Mac',
  },

  weiterlesen: {
    ueberschrift: 'Para seguir leyendo',
    einleitung:
      'Dos documentos forman parte del código fuente de esta aplicación — en el directorio del proyecto, en `docs/`, no en la imagen de disco que instalaste. Esta aplicación no puede abrirlos por sí misma: no tiene salida al exterior, y un enlace que no lleva a ninguna parte sería peor que ninguno.',
    handbuchName: 'Manual',
    handbuchSatz:
      'Lo que aquí aparece en palabras clave, en frases completas: el primer ' +
      'inicio, crear una carta, exportar, los atajos, y qué hacer cuando una ' +
      'carta se queda vacía. `docs/HANDBUCH.md`',
    administrationName: 'Datos y copias de seguridad',
    administrationSatz:
      'Dónde están los datos, cómo hacer una copia de seguridad y restaurarla, ' +
      'qué contiene el registro y qué ocurre al cambiar de versión, incluida la ' +
      'frase de que una copia de seguridad **sin contraseña** es texto claro. ' +
      '`docs/ADMINISTRATION.md`',
  },

  erststart: {
    ueberschrift: 'Primer inicio',
    text:
      'Esta aplicación no está notarizada. macOS la bloquea una vez; ' +
      'la autorizas en Ajustes del Sistema, en «Privacidad y seguridad».',
  },

  // Tours, and the areas that have no page (#1422).
  bereiche: {
    ueberschrift: 'Y un área que no tiene página en la barra lateral',
    einleitung:
      'No todo lo que puede hacer esta aplicación está en la barra lateral. Lo que llega de fuera no es un lugar y, por eso, no tiene página.',
    kiName: 'IA',
    kiSatz:
      'Un modelo de lenguaje lee y modifica cartas a través del **servidor MCP**, un proceso aparte sobre la misma biblioteca, que registras en tu herramienta de IA. Lo que sale así y lo que no está en **Interfaces**.',
    kiKnopf: 'Ir a las interfaces',
  },
  fuehrungen: {
    ueberschrift: 'Recorridos',
    einleitung:
      'Un recorrido te guía paso a paso por la aplicación y se puede repetir en cualquier momento. Junto a él, en una frase cada uno, cómo está organizada la aplicación.',
    startName: 'Los primeros pasos',
    startSatz:
      'Tres pasos hasta la primera carta: abrir un set, editar cartas, descargar plantillas. Aparecen solos en el primer inicio; si vuelven a aparecer en el siguiente es un ajuste.',
    startZeigen: 'Mostrar ahora los primeros pasos',
    arbeitsplatzName: 'El espacio de trabajo',
    arbeitsplatzSatz:
      'Cartas, plantillas, colección y mazos comparten una misma organización: navegador a la izquierda, inspector a la derecha, y el centro entre ambos. **⌥⌘[** pliega el navegador y **⌥⌘]** el inspector, y lo que pliegas se queda plegado en cada perfil. **⌥⌘S** pliega y despliega la barra lateral para que su ancho pase a los espacios de trabajo; ese estado pertenece a este dispositivo y se guarda por perfil.',
    paletteName: 'La paleta de comandos',
    paletteSatz:
      '**⌘K** la abre sobre cualquier página: páginas, sets y comandos, que se encuentran con unas pocas letras.',
    leisteName: 'La barra de contexto',
    leisteSatz:
      'Bajo el nombre de la aplicación indica quién trabaja dónde y con qué rol: persona, perfil, rol. El cambio se hace en la página **Usuarios y perfiles**; la barra lleva hasta allí.',
  },

  begriffe: {
    set: {
      name: 'Set',
      satz: 'Una colección de cartas junto con sus propios datos: nombre, descripción, símbolo. Vive en un archivo con la extensión `.mse-set`.',
    },
    karte: {
      name: 'Carta',
      satz: 'Aquello de lo que trata todo esto. Una carta lleva valores de campos (nombre, coste, texto de reglas) y se dibuja según una hoja de estilo.',
    },
    spiel: {
      name: 'Juego',
      satz: 'Dice **qué campos** tiene una carta. Un `.mse-game` los declara: `magic` conoce el coste de maná y la rareza, otro juego conoce otros.',
    },
    stil: {
      name: 'Hoja de estilo',
      satz: 'Dice **cómo** es el aspecto de una carta: qué capa va dónde, qué imagen da el marco. Un `.mse-style` siempre pertenece a un juego.',
    },
    stiloption: {
      name: 'Opción de estilo',
      satz: 'Un ajuste de la hoja de estilo que vale para todo el set o para una carta, por ejemplo el marco de una carta de draft o una edición distinta.',
    },
    symbolschrift: {
      name: 'Fuente de símbolos',
      satz: 'Dibuja símbolos de maná y similares. No es un tipo de letra, sino una colección de imágenes pequeñas con reglas sobre qué texto muestra cuál.',
    },
    paket: {
      name: 'Paquete',
      satz: 'El término general para juego, hoja de estilo, fuente de símbolos y plantilla de exportación. Los paquetes se descargan y se instalan; sin ellos, una carta se queda vacía.',
    },
    exportvorlage: {
      name: 'Plantilla de exportación',
      satz: 'Un pequeño programa que convierte las cartas de un set en **texto**: una lista para un foro, una tabla, un archivo para otro programa.',
    },
    druckbogen: {
      name: 'Hoja de impresión',
      satz: 'Un PDF con las cartas una junto a otra sobre papel, con marcas de corte. Para imprimir y recortar, no para compartir como archivo.',
    },
    vorbehalt: {
      name: 'Línea sin leer',
      satz: 'Una línea de un archivo abierto que este programa no ha interpretado. No se pierde: viaja sin cambios y vuelve a aparecer al guardar.',
    },
    befund: {
      name: 'Hallazgo',
      satz: 'Un aviso sobre algo que no salió bien al dibujar: una capa que falta, un carácter sin símbolo. La carta se dibuja de todos modos.',
    },
  },

  sprache: {
    ueberschrift: 'Por qué algunos textos siguen en inglés',
    einleitung:
      'El rótulo de un campo de carta y la frase que hay debajo proceden del **paquete del juego**, no de esta aplicación. Aparecen tal como los escribió la persona autora del paquete, normalmente en inglés.',
    schluesselName: 'El nombre es la clave',
    schluesselSatz:
      'Tu archivo `.mse-set` guarda el valor bajo ese nombre. Una traducción inventada quedaría más bonita y dejaría el archivo ilegible para cualquiera que busque en él, así que esta aplicación no inventa ninguna.',
    uebersetztName: 'Lo que está traducido se traduce',
    uebersetztSatz:
      'Si un paquete trae una traducción, se usa. En el `M15-Magic-Pack` descargado eran 25 campos de carta el 2026-08-21. Para los otros 224 no traía ninguna, y entonces aparece la misma palabra que en Magic Set Editor 2.',
    fensterName: 'La propia ventana habla ocho idiomas',
    fensterSatz:
      'Todo lo que dice esta aplicación por sí misma — botones, mensajes, esta página — sigue el idioma de los ajustes. El alemán es el idioma de origen; los otros siete están traducidos automáticamente y aún sin revisión editorial; así lo indica también la elección de idioma.',
    leisteName: 'También la barra de menús',
    leisteSatz:
      'El menú de la aplicación y los menús de edición y de ventana usan las palabras que macOS emplea para ellos en este idioma. **Cortar**, **Copiar** y **Pegar** conservan su vínculo con el sistema operativo.',
  },

  // **What this application knows about you** — Werkstück 13 of Phase 4.12
  // (#1351).
  datenschutz: {
    ueberschrift: 'Lo que esta aplicación sabe de ti',
    einleitung:
      'Desde que existen usuarios y perfiles, esta aplicación registra **datos personales**. Aquí se explica cuáles son, dónde están, cuáles salen de este equipo y cuáles expresamente no.',

    wasName: 'Qué se registra',
    wasSatz:
      'Todo es **opcional** salvo el nombre visible. Sobre la persona: **tratamiento**, **título o sufijo del nombre**, **nombre**, **otros nombres**, **apellido**, **lema** y dos **biografías**, una corta y una larga. Sobre la cuenta: **dirección**, **segunda dirección**, **número de teléfono** y, si lo defines, una **contraseña**. Además, una **imagen de perfil** y una **insignia** por perfil.',

    woName: 'Dónde está',
    woSatz:
      'En tu biblioteca (un archivo SQLite en tu carpeta personal) y en imágenes junto a ella. **Tres cosas están en otro sitio**, concretamente en el **llavero** de macOS: el secreto de tu segundo factor, el secreto de un inicio de sesión en un proveedor y las credenciales de servicios de cartas de terceros. La biblioteca se comparte; el llavero, no. De ahí la separación.',

    formName: 'De qué forma',
    formSatz:
      'Una **contraseña** nunca se guarda en texto claro: se protege con **Argon2id**, y la forma protegida lleva sus propios parámetros. Un **token de acceso** se guarda como suma de comprobación; el texto claro existe una sola vez, justo después de crearlo. Todo lo demás se guarda tal como lo escribiste: es tu archivo. **Y el propio archivo se puede cifrar**, en **Ajustes › Datos y copia de seguridad › Base de datos**. Entonces nada queda legible en el disco; la clave va al llavero de este Mac y **no se puede recuperar**. Una **copia de seguridad** recibe su propia **contraseña**; así queda cifrada y aun así se abre en cualquier equipo; sin contraseña, una base de datos cifrada solo la entrega si se pide expresamente.',

    ausName: 'Qué sale de este equipo',
    ausSatz:
      'En un **inicio de sesión**, tu elección llega al proveedor que hayas escogido: `github.com` y `api.github.com`, `accounts.google.com` y `oauth2.googleapis.com`, `login.microsoftonline.com`, o `twitter.com` y `api.twitter.com`. Allí solo se pide tu **nombre y tu identificador**, **no** tu dirección de correo, ni tus contactos, ni tus publicaciones. Vuelven y se guardan tres cosas: el proveedor, tu identificador en él y tu nombre visible allí.',

    nichtName: 'Qué no sale',
    nichtSatz:
      'Las **partes de tu nombre, las biografías, el número de teléfono y la dirección** nunca salen de este equipo. Están en tu biblioteca y no se envían a ningún sitio. No hay **telemetría**, ni informes de fallos, ni recuentos de uso, ni cuenta con nosotros: no existe ningún «nosotros».',

    dienstName: 'Y los servicios de cartas',
    dienstSatz:
      'Quien busca cartas consulta `api.scryfall.com` y, si están activados, `api.magicthegathering.io`, `graphql.mtgjson.com` o `api.scrydex.com`. Solo sale la **consulta de búsqueda**: ningún nombre de usuario, ningún identificador, nada de tu perfil. **Las miniaturas de los resultados** las obtiene la propia ventana — de `cards.scryfall.io` y, con magicthegathering.io, de `gatherer.wizards.com` y `gatherer-static.wizards.com`; allí solo va la dirección de la imagen. Los precios de toda la colección los obtiene `mtgjson.com` si está activado, y las ofertas del mercado `api.cardtrader.com`; a ellos solo llega el identificador de la impresión y, en CardTrader, tu token de acceso, y nada más. Los paquetes de plantillas llegan por `codeload.github.com` y los avisos de versiones por `api.github.com`. Qué puede hacer cada servicio aquí y en qué condiciones está en **Ajustes › Servicios › Servicios de cartas de terceros**; los enlaces de allí llevan a `scryfall.com`, `docs.magicthegathering.io`, `mtgjson.com`, `scrydex.com` y `www.cardtrader.com`: la propia aplicación nunca llama a esas páginas, que solo se abren en tu navegador cuando haces clic. **Y los enlaces de mercado de una carta** (T8): llevan a `www.cardmarket.com`, `www.cardtrader.com`, `www.tcgplayer.com`, `www.cardkingdom.com`, `gatherer.wizards.com` y `www.mtggoldfish.com`. **Construidos, no llamados**: la dirección se forma en este equipo a partir del nombre de la carta o de un identificador del proveedor, y solo sale cuando haces clic, en el navegador que elijas. No hace falta ninguna clave, y ninguna va en esas direcciones.',

    kiName: 'Y la IA',
    kiSatz:
      'Un modelo **local** se ejecuta en este equipo y no sale de él: la aplicación habla con `127.0.0.1` y con nada más. Un **modelo en la nube** es otra cosa: quien activa Anthropic, OpenAI o xAI envía a `api.anthropic.com`, `api.openai.com` o `api.x.ai`, respectivamente, la **tarea y los campos de la carta** de que se trate, junto con tu clave del llavero. **Con OpenAI no queda nada allí**: la interfaz guarda una respuesta por defecto, y esta aplicación lo desactiva expresamente. **Con xAI, igual**: el endpoint que conserva las respuestas 30 días no se usa nunca. Nada de tu perfil, ninguna segunda carta, ningún set. Ambos están **desactivados** hasta que los actives, y en **Ajustes › Servicios › Trabajar solo en local** todo modelo en la nube permanece desactivado aunque su propio interruptor diga otra cosa. Lo que vuelve **se muestra, no se adopta**: lo que va a un campo lo confirma una persona. Qué puede hacer Anthropic aquí y en qué condiciones figura en **Ajustes › Servicios › Servicios de cartas de terceros**; los enlaces de allí apuntan a `www.anthropic.com`, `platform.claude.com`, `openai.com`, `developers.openai.com`, `x.ai` y `docs.x.ai`: la propia aplicación nunca llama a esas páginas, que solo se abren en tu navegador cuando haces clic.',

    auskunftName: 'Acceso y eliminación',
    auskunftSatz:
      'En **Usuarios y perfiles › Mi acceso** exportas tus datos como archivo: todo lo que aquí se enumera, de una vez. Y en **Usuarios y perfiles** eliminas a un usuario, con sus perfiles, su imagen y el secreto de su segundo factor. **Lo que hay dentro de sus perfiles se queda**: sets, mazos, colección y torneos dejan de pertenecer a un perfil y ya no se pueden alcanzar desde la interfaz. Ninguna de las dos cosas necesita red ni el consentimiento de nadie.',
  },

  schnittstellen: {
    ueberschrift: 'Lo que sale y lo que entra',
    einleitung:
      'Esta aplicación funciona sin conexión. **Hacia la red** solo sale cuando alguien lo desencadena: descargar una plantilla, buscar una carta e iniciar sesión requieren pulsar un botón; solo busca versiones nuevas si un ajuste lo permite, y ese viene **desactivado** de fábrica. Qué direcciones se contactan y qué sale al hacerlo está en **Lo que esta aplicación sabe de ti**. Además hay dos caminos hacia este equipo que tampoco nadie toma por sí solo, y uno que sí se ejecuta solo: si falla el inicio, muestra el motivo en una ventana.',
    bezugName: 'Descargar plantillas',
    bezugSatz:
      'Al pulsar un botón, la aplicación descarga un paquete de plantillas por HTTP, en una revisión fija y con suma de comprobación. Sin pulsar el botón no ocurre nada.',
    scryfallName: 'Datos de cartas de Scryfall',
    scryfallSatz:
      'Al pulsar un botón, la aplicación consulta a `api.scryfall.com` una carta, un set o una imagen de carta. Lo que llega de allí **se muestra, no se adopta**: lo que va a un campo lo elige una persona. Sin red se conserva el último estado bueno, y dice qué antigüedad tiene.',
    fassungName: 'Buscar versiones nuevas',
    fassungSatz:
      'El único lugar que puede consultar sin pulsar un botón, y **solo** si está activado en **Ajustes**. Lo predeterminado es **desactivado**, y entonces no sale ni una sola solicitud. Activado, consulta a `api.github.com` la última versión publicada. La descarga solo se hace al pulsar un botón, y entonces se **verifican la firma y la suma de comprobación antes de colocar ningún archivo**; no se instala nada por sí solo.',
    mcpName: 'El servidor MCP',
    mcpSatz:
      'Un programa propio llamado `mse-mcp` con el que un modelo de lenguaje puede trabajar en un set: leer cartas y campos, rellenar campos, crear, copiar, borrar y reordenar cartas, importar un set, guardar un set y dibujar una carta. La medida no es la cantidad sino la reversibilidad: ninguna herramienta sobrescribe nada existente —donde ya hay un archivo, se niega— y cada cambio dice cómo deshacerlo. Cuando lo inicia un cliente como Claude Desktop, habla por la entrada y salida estándar: sin puerto, sin dirección. **Por HTTP** solo escucha cuando lo inicias en **Ajustes › Conexiones**, y solo donde lo indica **Dónde escuchan los servicios**.',
    mcpNichtAn:
      'No se ejecuta a la vez, pero **viene incluido**: en el paquete, en `MagicSetEditor4Mac.app/Contents/MacOS/mse-mcp`, junto con `mse`, la línea de comandos. **Por sí sola, esta aplicación no deja nada fuera de su paquete.** En **Ajustes › Conexiones › Herramientas sin ventana** configuras con un botón lo que quieras —el enlace a `mse` en `~/.local/bin` con su entrada `PATH` en `~/.zprofile`, la entrada para Codex, Claude Desktop o Claude Code— y allí también retiras cada entrada; lo que está fuera de su propio bloque no lo toca. Lo que cambian las herramientas, esta ventana lo nota sola: pregunta cada dos segundos por el estado de la base de datos y recarga la página afectada.',
    dienstName: 'El servicio para scripts',
    dienstSatz:
      'Un programa propio llamado `mse-rest` con el que un script alcanza las mismas capacidades que `mse`: leer sets y cartas, rellenar un campo, importar o guardar un set, dibujar una carta. Cada petición necesita un token. **Dónde escucha lo decides tú** en **Dónde escuchan los servicios**: por defecto solo este Mac (`127.0.0.1`); si eliges una dirección que sale de este Mac, exige TLS y un acceso.',
    dienstNichtAn:
      'Está en el paquete, en `MagicSetEditor4Mac.app/Contents/MacOS/mse-rest`, y no se inicia solo: quien no quiere un servicio no lo tiene. Se inicia en **Ajustes › Conexiones**, o como LaunchAgent en **Herramientas sin ventana**; entonces macOS lo inicia en cada inicio de sesión hasta que lo retires allí. Quien lo inicia recibe la dirección y el token. Las peticiones desde un navegador se rechazan antes incluso de comprobar el token.',
    startfensterName: 'El mensaje cuando falla el inicio',
    startfensterSatz:
      'El único camino que la aplicación toma **por sí misma**. Si no se puede abrir la base de datos, llama a `osascript` de macOS para mostrar el motivo en una ventana; de lo contrario solo aparecería en un canal que nadie ve tras un doble clic. El texto entra como valor, no como script.',
    dateienName: 'Archivos en este equipo',
    dateienSatz:
      'Sets, paquetes de plantillas, imágenes y PDF escritos. Están donde tú los pones; la biblioteca está en tu carpeta de usuario.',
    sonstSatz:
      'Nada más. **Sin telemetría**, sin informes de fallos, sin recuentos de uso. Sí hay una cuenta, pero en el proveedor que elijas y no con nosotros; lo que sale al hacerlo se describe en **Lo que esta aplicación sabe de ti**.',
    grenzenName: 'Límites',
    grenzenSatz:
      'Cada acceso de máquina tiene un **crédito**, y cada llamada cuesta según su tipo; cuando se agota, la respuesta lo dice junto con el tiempo de espera. Los trabajos pesados (rasterizar, importar, exportar, operaciones masivas) tienen un número fijo de plazas por clase que comparten la ventana, la línea de comandos, MCP y el servicio. Quien adivina solo bloquea su propia credencial.',
    grenzenZahlen:
      'Hoy: {burst} unidades de golpe, {nachfuellen} por minuto; una lectura cuesta {lesen}, una escritura {schreiben}, una importación {einfuhr}; como máximo {rendern} rasterizaciones a la vez; {reserve} unidades quedan para la ventana.',
    grenzenOhneKontingente:
      '**Las cuotas están desactivadas** (`grenzen.kontingente`) — para un entorno controlado. Los puestos por tipo de trabajo, los límites del cuerpo y de la respuesta, los plazos, el bloqueo contra quien adivina, el acceso y TLS siguen vigentes.',
  },
} as const
