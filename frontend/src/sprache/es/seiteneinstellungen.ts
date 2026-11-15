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

  titel: 'Ajustes',

  namen: {
    ...schalternamen,
    sprache: 'Idioma',
    'sprache.eigene': 'Traducción propia',
    darstellung: 'Formato regional',
    zeitzone: 'Zona horaria',
    'filter.gespeichert': 'Filtros guardados',
    'kuerzel.ziffern': 'Asignación de ⌘0 a ⌘9',
    erscheinungsbild: 'Tema',
    schriftgroesse: 'Tamaño del texto',
    aktualisierung: 'Buscar versiones nuevas',
    ersteschritte: 'Introducción al iniciar',
    seitenleiste: 'Barra lateral',
    leistenlage: 'Posición de la barra lateral',
    fensterstart: 'Ventana al iniciar',
    fensterbreite: 'Último ancho recordado',
    fensterhoehe: 'Última altura recordada',
    schriftart: 'Tipografía de títulos',
    feldtiefe: 'Formulario de carta',
    'tisch.autoanordnung': 'Ordenar las cartas de la mesa automáticamente',

    'barrierefreiheit.bewegung': 'Movimiento',
    'barrierefreiheit.kontrast': 'Contraste',
    'sicherheit.faktorbeimwechsel': 'Segundo factor al cambiar de perfil',

    ...vorgabennamen,
    ...protokollnamen,
    // **Die lokale KI-Laufzeit** — A7, #1446.
    // **A cloud provider's address** — piece 3 of A18 (#1533), labelled with
    // #1540.
    'dienst.anthropic.adresse': 'Dirección de Anthropic',
    'dienst.openai.adresse': 'Dirección de OpenAI',
    'dienst.xai.adresse': 'Dirección de xAI',
    ...netznamen,
    ...handelsnamen,
    ...gestaltungsnamen,
    ...ansichtsnamen,
    'konto.schau': 'Leer mi propia cuenta del marketplace',
    ...kinamen,
    // **The client id per provider** — Werkstück 4 of Phase 4.12 (#1351).
    'anbieter.github.kennung': 'ID de cliente en GitHub',
    'anbieter.google.kennung': 'ID de cliente en Google',
    'anbieter.microsoft.kennung': 'ID de cliente en Microsoft',
    'anbieter.x.kennung': 'ID de cliente en X',

    // **The workspace and the two limits** — Werkstück O3, Phase 5.2 (#1427).
    'arbeitsplatz.navigator': 'Columna izquierda',
    'arbeitsplatz.inspektor': 'Columna derecha',
    'arbeitsplatz.navigatorbreite': 'Ancho de la columna izquierda',
    'arbeitsplatz.inspektorbreite': 'Ancho de la columna derecha',
    'richtlinie.nurlokal': 'Trabajar solo en local',
    'richtlinie.kostengrenze':
      'Límite de coste en la unidad más pequeña (en EUR: céntimos)',
  },

  // Origin, default and search — work item U11 of phase 5.1 (#1422).
  herkunftMarke: 'establecido por {herkunft}',
  wirksamKurz: 'Se aplica «{wert}».',
  wirksamOffenKurz: 'Se aplica el valor predeterminado del ordenador.',
  wirksamGehalten: 'Se aplica «{wert}»: una política lo mantiene.',
  herkunftZeile: 'Establecido por {herkunft}, nivel {ebene}.',
  herkunftNamen: {
    profil: 'este perfil',
    nutzer: 'ti',
    geraet: 'este dispositivo',
    vorgabe: 'nadie (predeterminado)',
  },
  ebenenNamen: {
    profil: 'perfil',
    nutzer: 'usuario',
    geraet: 'dispositivo',
    arbeitsplatz: 'espacio de trabajo',
  },
  ebenenErklaerung:
    'Cada ajuste se aplica en un nivel: dispositivo, usuario, perfil o espacio de trabajo. **Dispositivo** describe este ordenador (una ruta, el tamaño de una ventana) y no viaja en un intercambio. **Usuario** se aplica a ti en cada uno de tus perfiles. **Perfil** se aplica solo aquí y prevalece sobre lo que hayas establecido como usuario. **Espacio de trabajo** es cómo están dispuestas las columnas de esta ventana. Se aplica lo que dice el perfil; si no, lo que hayas establecido como usuario; si no, el valor predeterminado; en una política, el valor más estricto. Ningún ajuste requiere derechos de administrador.',
  fachleute: 'Lo que los expertos pueden ajustar además aquí',
  fachleuteSatz:
    'Plazos, accesibilidad desde otras redes, la salida y los servicios para otros programas. Desplegar no cambia nada: solo muestra más.',
  zuruecksetzen: 'Restablecer a «{vorgabe}»',
  erbeMitWert: '{erbe} ({wert})',
  zuruecksetzenVon: 'Restablecer {name} a su valor predeterminado',
  suche: 'Buscar un ajuste',
  suchePlatzhalter: 'Nombre, clave o valor',
  sucheTreffer: '{zahl} coincidencias',
  sucheKeine: 'Ningún ajuste coincide con «{frage}».',
  zumGebiet: 'Mostrar en el área «{gebiet}»',

  // **Below the appearance settings** — UX-09 (#1681).
  gestaltung: gestaltungsbeiwerk,

  hinweise: {
    ...protokollhinweise,
    'sprache.eigene':
      'Qué traducción propia usa la ventana — vacío significa ninguna. Si ya no existe, vuelve a valer el idioma elegido.',
    feldtiefe:
      'Cuánto del formulario de carta se muestra. **Normal** es lo que rellenas; **Avanzado** añade lo que calcula un script del paquete de juego; **Experto** añade lo que el paquete de juego no explica en absoluto: `stylesheet` y similares. Los tres se anidan: más profundo significa más, nunca otra cosa. Se aplica por perfil.',
    sprache:
      'Se aplica a toda la ventana y a la barra de menús. Lo que viene de un paquete MSE (nombres de campos, nombres de juegos) se queda como está escrito en el archivo. El alemán es el idioma de origen; los demás idiomas están traducidos automáticamente y aún no han sido revisados.',
    darstellung:
      'Cómo se ven los números, importes y fechas, por ejemplo **de-CH** o **en-GB**. Vacío significa: como lo tenga el sistema. El idioma de los textos no se ve afectado.',
    zeitzone:
      'En qué zona se muestran las horas, por ejemplo **Europe/Berlin**. «Como el sistema» es la del propio ordenador. Se guarda por perfil y no viaja en un intercambio.',
    'filter.gespeichert':
      'Los filtros guardados de cada vista como lista en JSON: nombre, vista, marcas y si se aplica **y** u **o**. Se crean en el navegador de su vista; pertenecen al perfil y viajan con la exportación de ajustes.',
    'kuerzel.ziffern':
      'Qué página abre un dígito con ⌘. Vacío significa: las áreas de la barra lateral. Se establece en la tabla de atajos, en «General».',
    erscheinungsbild:
      'Claro, oscuro o lo que diga el ordenador. La elección se aplica solo a esta aplicación.',
    schriftgroesse:
      'Escala toda la interfaz, no solo el texto: los espacios y los bordes crecen con ella. La vista previa de la carta se queda como está: muestra lo que se imprimiría.',
    aktualisierung:
      'El valor predeterminado es **desactivado**: no sale nada a la red y no se pregunta nada. Con **avisar**, la aplicación comprueba al iniciar si existe una versión más reciente y lo indica, nada más. Con **descargar** ofrece además un botón que descarga la versión a la carpeta de descargas y compara su suma de comprobación; nada se instala por sí solo.',
    ersteschritte:
      'Los tres pasos hasta tu primera carta, como ventana al iniciar. El mismo interruptor está dentro de esa ventana: desactívalo allí y lo encontrarás de nuevo aquí.',
    seitenleiste:
      'La barra de la izquierda con las áreas de la aplicación. Plegada, su ancho pasa a los espacios de trabajo; ⌥⌘S, el botón de su borde y la entrada de **Visualización** la pliegan y despliegan. Se guarda por perfil y no viaja en un intercambio.',
    leistenlage:
      'A la izquierda está en vertical junto al contenido; arriba, como una fila encima de él. Las mismas entradas, atajos y teclas de flecha, solo que en horizontal. Se guarda por perfil y no viaja en un intercambio.',
    fensterstart:
      'El valor predeterminado es **Llenar la pantalla**: la ventana ocupa el espacio que tiene, y **no pantalla completa**, que se llevaría la barra de menús. **Tamaño normal** abre con el tamaño integrado, **Como la última vez** con el que la dejaste. Se guarda por perfil y no viaja en un intercambio.',
    fensterbreite:
      'Lo que se midió al cerrarse la ventana por última vez; **solo para «Como la última vez»**. Vacío significa que aún no hubo última vez y la ventana se abre con su tamaño integrado.',
    fensterhoehe:
      'Lo mismo para la altura. Ambas se escriben al cerrarse la ventana; una medida fuera de 700 a 8192 puntos no se recuerda: una ventana que no cupiera en la siguiente pantalla ya no se podría agarrar.',
    schriftart:
      'La tipografía de los títulos y de los nombres de set. El texto del cuerpo sigue siendo la tipografía del sistema: una tipografía decorativa sobre un párrafo se lee mal. Las cuatro vienen con macOS; no se descarga nada.',
    'barrierefreiheit.bewegung':
      'La anulación va **en un solo sentido**: desactiva el movimiento, nunca lo activa. Quien tenga **reducir movimiento** en el sistema no recibe movimiento aquí de todos modos: una aplicación que lo anulara iría en su contra.',
    'barrierefreiheit.kontrast':
      'En el mismo sentido: **alto** quita la textura de fibra del fondo y sube los colores hasta que todo texto alcanza 7 : 1. Quien pide más contraste al sistema obtiene lo mismo.',
    'sicherheit.faktorbeimwechsel':
      'La **única** anulación que afloja, y solo al **cambiar** de perfil. No cambia nada en el inicio de sesión: allí el segundo factor cumple su función. En un ordenador con una sola persona, un código de un solo uso en cada cambio es fricción sin beneficio.',
    'vorgabe.zustand':
      'Con qué empieza una fila nueva de la colección. **Sin valor predeterminado** significa que el campo queda vacío, como siempre.',
    'vorgabe.waehrung':
      'Con qué empieza el campo de moneda de una fila nueva, por ejemplo **EUR**. No se convierte nada.',
    'vorgabe.einzelspiel':
      'Si un torneo nuevo lleva partidas de **un solo** juego. Decide la presentación, no el cálculo: el porcentaje de victorias en partidas y el de los oponentes se calculan de todos modos.',
    'vorgabe.hochladepfad':
      'La carpeta en la que empieza un diálogo para abrir un archivo: imágenes, MTG JSON y todo lo demás que venga del disco. Una ruta completa o una que empiece por **~/**; vacío significa donde lo abra el sistema operativo.',
    'vorgabe.ausfuhrpfad':
      'Dónde empieza un diálogo de guardado cuando la acción no tiene un lugar propio. Si la carpeta no existe al guardar (un disco desconectado, por ejemplo), la app te lo dice en lugar de guardar en otro sitio.',
    'vorgabe.sicherungspfad':
      'Dónde empieza el diálogo de una copia de seguridad de tu biblioteca. Una carpeta inexistente detiene la copia con un mensaje.',
    'vorgabe.diagnosepfad':
      'Dónde se guardan el texto de diagnóstico y los paquetes. Una carpeta inexistente lo detiene con un mensaje.',
    'bildausfuhr.ordner':
      'Dónde escribió la salida de imágenes por última vez. La página Imágenes la propone y anota aquí cada carpeta elegida de nuevo — una para todos los sets, como en MSE. Vacío significa que el diálogo de carpetas se abre en la carpeta para los archivos que guardas.',
    'ausgabe.ziele':
      'Qué destino mostró «Exportar» por última vez en cada set. Una lista en JSON, la elección más reciente al final, 200 como máximo; se escribe al elegir en «¿A dónde?». Vacío significa que no hay nada preseleccionado.',
    'dienst.anthropic.adresse':
      'A dónde van las solicitudes. Vacío significa la interfaz propia de Anthropic. Una dirección distinta es una **pasarela**: un servicio que reenvía la solicitud; su clave se guarda por separado.',
    'dienst.openai.adresse':
      'Lo mismo para OpenAI. Vacío significa `https://api.openai.com`. Cualquier cosa que introduzcas aquí significa que ya no hablas con OpenAI, sino con lo que responda en esa dirección.',
    'dienst.xai.adresse':
      'Lo mismo para xAI. Vacío significa `https://api.x.ai`.',
    ...netzhilfe,
    ...handelshilfe,
    ...gestaltungshinweise,
    'konto.schau':
      'Si este programa puede leer tu **propia** cuenta en un marketplace: tus ofertas y tus pedidos. **Desactivado** significa: ni siquiera un vistazo. Activar el marketplace no basta: aquel responde lo que piden **otros** vendedores; este, lo que hay **contigo**. Nunca se escribe nada allí: ni oferta, ni carrito, ni compra.',
    ...kihinweise,
    'anbieter.github.kennung':
      'El `client_id` de la aplicación OAuth que creaste en GitHub. Vacío significa: no se ofrece iniciar sesión con GitHub.',
    'anbieter.google.kennung':
      'El `client_id` de la aplicación OAuth que creaste en Google. Suele terminar en `apps.googleusercontent.com`.',
    'anbieter.microsoft.kennung':
      'El id de aplicación del portal Microsoft Entra: un UUID. Microsoft **no** exige ningún secreto para este flujo.',
    'anbieter.x.kennung':
      'El `client_id` de la aplicación que creaste en el portal de desarrolladores de X.',
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
    'konto.schau.aus': 'Desactivado',
    'konto.schau.an': 'Activado',
    ...kiwerte,
    'sprache.erben': 'Como en el usuario',
    'erscheinungsbild.erben': 'Como en el usuario',
    'schriftgroesse.erben': 'Como en el usuario',
    'schriftart.erben': 'Como en el usuario',
    'sprache.de': 'Deutsch',
    'sprache.en': 'English',
    'sprache.es': 'Español',
    'sprache.fr': 'Français',
    'sprache.it': 'Italiano',
    'sprache.ja': '日本語',
    'sprache.ru': 'Русский',
    'sprache.zh': '简体中文',
    'erscheinungsbild.system': 'Como el sistema',
    'erscheinungsbild.hell': 'Claro',
    'erscheinungsbild.dunkel': 'Oscuro',
    'schriftgroesse.klein': 'Pequeño',
    'schriftgroesse.normal': 'Normal',
    'schriftgroesse.gross': 'Grande',
    'schriftgroesse.sehrgross': 'Muy grande',
    'aktualisierung.aus': 'Desactivado',
    'aktualisierung.melden': 'Solo avisar',
    'aktualisierung.holen': 'Avisar y descargar',
    'ersteschritte.an': 'Activado',
    'ersteschritte.aus': 'Desactivado',
    'seitenleiste.sichtbar': 'Visible',
    'seitenleiste.verborgen': 'Oculta',
    'leistenlage.senkrecht': 'Izquierda',
    'leistenlage.waagerecht': 'Arriba',
    'fensterstart.maximiert': 'Llenar la pantalla',
    'fensterstart.normal': 'Tamaño normal',
    'fensterstart.zuletzt': 'Como la última vez',

    'feldtiefe.normal': 'Normal',
    'feldtiefe.erweitert': 'Avanzado',
    'feldtiefe.experte': 'Experto',
    'tisch.autoanordnung.an': 'Sí — la mesa las coloca',
    'tisch.autoanordnung.aus': 'No — lo que colocas se queda donde está',
    'tisch.beschriftung': 'Etiquetas en la mesa',
    'tisch.beschriftung.aus': 'Desactivadas — solo recuentos',
    'tisch.beschriftung.klein': 'Pequeñas',
    'tisch.beschriftung.normal': 'Medianas',
    'tisch.beschriftung.gross': 'Grandes',
    'schriftart.pergament': 'Pergamino — tranquila y legible',
    'schriftart.fantasy': 'Fantasía — desenfadada',
    'schriftart.antike': 'Antigüedad — mayúsculas romanas',
    'schriftart.kalligrafie': 'Caligrafía — manuscrita',
    'barrierefreiheit.bewegung.system': 'Como el sistema',
    'barrierefreiheit.bewegung.reduziert': 'Siempre poco',
    'barrierefreiheit.kontrast.system': 'Como el sistema',
    'barrierefreiheit.kontrast.hoch': 'Siempre alto',
    'sicherheit.faktorbeimwechsel.verlangen': 'También al cambiar',
    'sicherheit.faktorbeimwechsel.nurAnmelden': 'Solo al iniciar sesión',
    'vorgabe.zustand.ohne': 'Sin valor predeterminado',
    'vorgabe.zustand.mint': 'Mint',
    'vorgabe.zustand.nearMint': 'Near Mint',
    'vorgabe.zustand.excellent': 'Excellent',
    'vorgabe.zustand.good': 'Good',
    'vorgabe.zustand.lightPlayed': 'Light Played',
    'vorgabe.zustand.played': 'Played',
    'vorgabe.zustand.poor': 'Poor',
    'vorgabe.einzelspiel.aus': 'Desactivado',
    'vorgabe.einzelspiel.an': 'Activado',

    // **The workspace and the limit** — Werkstück O3 of Phase 5.2 (#1427).
    'arbeitsplatz.navigator.erben': 'Como en el usuario',
    'arbeitsplatz.navigatorbreite.erben': 'Como en el usuario',
    'arbeitsplatz.inspektorbreite.erben': 'Como en el usuario',
    'arbeitsplatz.navigator.sichtbar': 'Visible',
    'arbeitsplatz.navigator.verborgen': 'Oculta',
    'arbeitsplatz.inspektor.erben': 'Como en el usuario',
    'arbeitsplatz.inspektor.sichtbar': 'Visible',
    'arbeitsplatz.inspektor.verborgen': 'Oculta',
    'richtlinie.nurlokal.aus': 'Desactivado',
    'richtlinie.nurlokal.an': 'Activado',
  },

  // **Was dem Menschen gehört** — Werkstück 8 von Phase 4.12 (#1351).
  // **Where a login comes from in the first place** — follow-up to Werkstück 4
  // of Phase 4.12 (#1351).
  anmeldung: 'Inicio de sesión con un proveedor',
  anmeldungHinweis:
    'Esta aplicación **no incluye credenciales**. Para iniciar sesión con GitHub, Google, Microsoft o X, crea tú mismo una **aplicación OAuth** en el proveedor e introduce aquí su `client_id`. Como URI de redirección, indica la dirección que se muestra en el navegador durante el inicio de sesión: apunta a este ordenador y solo es válida mientras dura el inicio de sesión.',

  anmeldungGeheimnis:
    'GitHub exige además un `client_secret`; Google lo deja opcional y Microsoft lo prohíbe para programas como este. Va **al llavero**, no aquí: la biblioteca se comparte, el llavero no. Créalo en **Acceso a Llaveros** como una **contraseña de aplicación**, con el nombre `MagicSetEditor4Mac-Anmeldung` y la cuenta `github`. En la línea de comandos: `security add-generic-password -s MagicSetEditor4Mac-Anmeldung -a github -w`.',

  geheimnis: 'Secreto',
  geheimnisPflicht: 'es obligatorio y falta',
  geheimnisVerboten:
    'no requiere ningún secreto y lo rechaza: aquí basta con el ID de cliente.',

  anbietername: {
    github: 'GitHub',
    google: 'Google',
    microsoft: 'Microsoft',
    x: 'X',
  },

  wirktNachNeustart: 'Surte efecto tras reiniciar.',

  trefferInGruppe: '{zahl} coincidencias para «{frage}»',
  gesuchtMarke: 'encontrado',
  sucheSprung: 'Ir a la primera coincidencia',

  vorgaben: 'Valores predeterminados',

  // **The pointer to the help** — Werkstück 13 of Phase 4.12 (#1351).

  dienste: 'Servicios de cartas de terceros',
  diensteHinweis:
    'Desactivado significa: ni una sola solicitud va a este servicio. La aplicación sigue funcionando sin ningún servicio; solo faltará lo que viene de fuera.',

  kiAnbieter: 'Modelos de lenguaje',
  kiAnbieterHinweis:
    'A dónde pueden ir las preguntas. Cada proveedor calcula en sus máquinas y ve lo que envías, por eso cada uno se activa por separado y ninguno lo está por sí solo.',
  // See the German catalogue for why these two live here and not under
  // `werte`: they are the same for every service.
  dienstAus: 'Desactivado',
  dienstAn: 'Activado',
  dienstUngeprueft: 'Nunca probado contra el servicio real',
  dienstUngeprueftHinweis:
    'Aquí no hay ninguna cuenta para este servicio. Lo que se prueba es nuestro lado de la frontera: la solicitud, las cabeceras, la asignación. Que el servicio responda como suponemos no está medido.',
  fragen: 'Quién responde a qué',
  fragenHinweis:
    'Cuando varios servicios pueden responder a la misma pregunta, esto decide a cuál se le pregunta. **No hay ninguna alternativa silenciosa**: si el elegido calla, se te avisa, en lugar de ver sin más una respuesta de otra fuente.',
  'frage.suche': 'Búsqueda de cartas',

  // The service catalogue — work item A1 of phase 5.3 (#1440).
  ...dienstkatalog,

  dienstSchluesselFehlt: 'No hay credenciales guardadas',
  dienstSchluesselHinweis:
    'Este servicio necesita credenciales. Viven en el llavero de macOS, no en el archivo de la biblioteca: ese se comparte.',
  zugaenge: 'Credenciales',
  zugaengeHinweis:
    'Viven en el llavero de macOS y **no** en el archivo de la biblioteca: ese se comparte. Lo que ves aquí es solo **si** hay algo guardado; el valor en sí nunca vuelve.',
  // See the German catalogue: these name the headers each service asks for
  // (documented in decision 0056).

  kuerzelUeberschrift: 'Atajos de teclado',
  ...einstellungsbelegung,
  kuerzelHinweis:
    'Valen en toda la ventana. Dentro de un campo de texto, ⌘Z y ⌘F pertenecen al campo.',
  kuerzelZweiteReihe:
    'Los ocho atajos con ⌥ están en las teclas de la fila superior, empezando por la izquierda. ' +
    'En un teclado con otra distribución (AZERTY, por ejemplo) ahí hay impreso otro carácter; ' +
    'la tecla es la misma.',
  kuerzelTaste: 'Tecla',
  kuerzelHandlung: 'Acción',

  austausch,
  eigenuebersetzung,
  // **An account's remaining quota at a provider** — A5, #1444.
  kontingentTitel: 'Cuota del proveedor',
  kontingentErklaerung:
    'Cuántos créditos incluye el plan y cuántos quedan. La consulta sale a la red y ella misma cuesta un crédito.',
  kontingentHolen: 'Comprobar cuota',
  kontingentHolenBei: 'Comprobar la cuota en {dienst}',
  kontingentLaeuft: 'Comprobando …',
  kontingentZahlen: 'Quedan {rest} de {gesamt} créditos.',
  kontingentFrist:
    'El proveedor puede informar de esta cifra con hasta {minuten} minutos de retraso; no es un tope. Lo que realmente limita aquí el gasto es el presupuesto de esta biblioteca.',
  kontingentUeberzug: 'Un crédito por encima del plan cuesta {satz}.',
  // **What is known about AI models** — A6, #1445.
  ...ki,

  // **Die Zugangsdaten** (#1534) — eigene Datei, ein Katalog (§9).
  ...zugaenge,

  ...netzbereich,
}
