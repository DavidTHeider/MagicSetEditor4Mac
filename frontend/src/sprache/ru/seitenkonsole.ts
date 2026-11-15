// **The script console** — `seiten.karten.konsole`, split out because
// `seitenkarten.ts` reached the line limit (§2, U13 of phase 5.1, #1424).
export const konsole = {
  ueberschrift: 'Консоль скриптов',
  erklaerung:
    'Выражение скрипта MSE в контексте открытой карты. В окружении те же имена, что и при отрисовке: «card», «set», «stylesheet» и функции игрового пакета. Консоль ничего не изменяет.',
  ausdruck: 'Выражение',
  platzhalter: 'card.name',
  auswerten: 'Вычислить',
  ergebnis: 'Результат',
  art: 'Тип: {art}',
  keineKarte: 'Для этого должна быть открыта карта.',
  zuDenKarten: 'К картам',
  vorbehalte: 'Оговорки при построении окружения',
  stelle: '(строка {zeile}, столбец {spalte})',
  namen: 'Доступные имена',
  namenErklaerung:
    'Что выражение может назвать на этой карте — библиотека и всё, что добавляет шаблон. Щелчок вставляет имя.',
  beispiele: 'Примеры',
  beispieleImmer: 'Работают везде',
  beispieleImmerErklaerung:
    'Щелчок вставляет выражение. Им не нужно ничего, кроме самого языка.',
  beispieleVorlage: 'Нужен шаблон',
  beispieleVorlageErklaerung:
    '«card» и «set» помещает в окружение пакет стиля. Если для этой карты он не установлен, ответ так и скажет — примеры выше всё равно работают.',
  beispielListe: {
    rechnung: 'Арифметика',
    grossSchreiben: 'Строка заглавными буквами',
    laenge: 'Сколько в чём-то символов',
    ausschnitt: 'Часть строки',
    bedingung: 'Условие',
    name: 'Название этой карты',
    grossName: 'Её название заглавными буквами',
    regeltext: 'Длина её текста правил',
    setangabe: 'Название набора',
  },
  vervollstaendigung: 'Подсказки',
  uebernehmen: 'Принять',
  namenFiltern: 'Сузить список имён',
  namenGezeigt:
    'Показано {gezeigt} из {ganz} — введи текст, чтобы сузить список.',
  namenLeer: 'Нет подходящих имён.',
  kuerzel: '⌘↩ вычисляет.',
  abbrechen: 'Отмена',
  laeuft: 'Выполняется …',
  dauer: '{ms} мс',
  abgebrochen:
    'Отменено — выполнение завершилось до появления результата. Выражение по-прежнему в поле.',
  baum: 'Результат в виде дерева',
  verlauf: 'История этого сеанса',
  verlaufErklaerung:
    'Последние двадцать выражений, только в этом сеансе. Щелчок загружает выражение обратно; всё, что похоже на токен или пароль, здесь скрыто и обратно не загружается.',
  verlaufLeer: 'Пока ничего не вычислялось.',
  verlaufLoeschen: 'Очистить историю',
  ausgangGelungen: 'успешно',
  ausgangGescheitert: 'сбой',
  ausgangAbgebrochen: 'отменено',
}
