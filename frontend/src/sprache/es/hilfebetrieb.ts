// **Operation and trust** — OPS-07
// (#1724).
export const betrieb = {
  ueberschrift: 'Funcionamiento y confianza',
  einleitung:
    'Qué está instalado, dónde está, qué se ha verificado y qué no, leído de esta misma aplicación.',
  fassungName: 'Versión',
  fassungSatz:
    '{fassung}, compilada a partir de {stand}, el estado del {gebaut}.',
  fassungVeraendert:
    'La compilación contenía cambios que no estaban confirmados en el repositorio.',
  datenName: 'Datos',
  datenSatz: '{zahl} migraciones aplicadas, la última {letzte}.',
  datenOhne: 'No hay ninguna biblioteca abierta.',
  signaturName: 'Firma',
  signaturSatz:
    'Firmada ad hoc, **no notarizada**. En el primer inicio macOS rechaza la aplicación; se puede autorizar una vez en Ajustes del Sistema › Privacidad y seguridad › **Abrir de todos modos**.',
  aktualisierungName: 'Actualizaciones',
  aktualisierung: {
    aus: 'Desactivadas: esta aplicación no busca versiones nuevas.',
    melden: 'Avisa de una versión nueva en GitHub y no descarga nada.',
    holen:
      'Descarga una versión nueva de GitHub y pregunta antes de sustituir la actual.',
  },
  orteName: 'Dónde está todo',
  orte: {
    bestand: 'Biblioteca',
    pakete: 'Paquetes de plantillas',
    bilder: 'Imágenes insertadas',
    zwischenspeicher: 'Cachés',
    protokoll: 'Registro',
    leistungsprofile: 'Perfiles de rendimiento',
  },
  lizenzenName: 'Licencias',
  lizenzenSatz:
    'El código fuente está bajo la GNU General Public License 2.0 o posterior. Las partes de terceros, con sus licencias, figuran en `DRITTANBIETER.md`, junto al código fuente; un paquete de plantillas trae sus propias condiciones, y un modelo de IA las de su proveedor.',
  empfaengerName: 'Adónde van los datos',
  empfaengerSatz:
    'Solo a los servicios que actives; cuáles son y qué sale está en **Ayuda › Lo que esta aplicación sabe de ti**. No se envían datos de uso.',
  grenzenName: 'Límites de seguridad',
  grenzenSatz:
    'El servicio y el servidor de herramientas escuchan solo en `127.0.0.1`, salvo que se configure otra cosa, y exigen un token. Si a alguno le das una dirección en la red en **Ajustes › Diagnóstico**, solo funciona con TLS, y la página avisa de que sale del equipo; por stdio, el servidor de herramientas habla por sus propios flujos. La biblioteca se puede cifrar, en **Ajustes › Datos y copia de seguridad**.',
  kontaktName: 'Cuando algo no funciona',
  kontaktSatz:
    'En **Ajustes › Diagnóstico** encontrarás el diagnóstico, el paquete de diagnóstico y las cifras de este proceso. Los informes de errores se envían a github.com/DavidTHeider/MagicSetEditor4Mac; nada sale por sí solo.',
  pruefstandName: 'Qué está verificado',
  pruefstandSatz:
    'Verificado significa que una ejecución lo ha demostrado: una prueba cuya evidencia, con su commit, está junto al código fuente, o un test que ejecuta cada pipeline; solo se publica lo que lo ha superado. Presente significa: construido, pero nunca ejecutado contra el original ni contra la plataforma.',
  spalteBereich: 'Área',
  spalteGegenstand: 'Asunto',
  spalteStand: 'Estado',
  spalteBeleg: 'Evidencia',
  belegIssue: 'incidencia n.º {nummer}',
  bereiche: {
    plattform: 'Plataforma',
    sprache: 'Idiomas',
    barrierefreiheit: 'Accesibilidad',
    regeln: 'Reglas',
    integration: 'Servicios externos',
  },
  staende: {
    geprueft: 'verificado',
    vorhanden: 'presente',
    beschrieben: 'descrito',
    blockiert: 'espera a una persona',
  },
  gegenstaende: {
    'macOS 15.3 bis 26': 'macOS 15.3 a 26',
    'macOS 15.2 auf Apple Silicon': 'macOS 15.2 en Apple Silicon',
    'macOS 27 auf Apple Silicon': 'macOS 27 en Apple Silicon',
    'Tastatur und Vorleser im WebKit': 'Teclado y lector de pantalla en WebKit',
  },
}
