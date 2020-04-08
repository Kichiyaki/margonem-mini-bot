# Margonem Mini Bot

**NIE ODPOWIADAM ZA WSZELKIE BANY!**

Zbudowaną aplikację można pobrać [stąd](http://dawid-wysokinski.pl/upload/margonem-mini-bot.rar).

Przykładowy plik konfiguracyjny.

```
{
  "accounts": [
    {
      "username": "username",
      "password": "pass",
      "proxy": "proxy(opcjonalne)",
      "characters": [
        {
          "id": "świat#id_postaci",
          "map_id": "id_mapy"
        },
        {
          "id": "świat#id_postaci",
          "map_id": "id_mapy"
        }
      ]
    }
  ],
  "debug": false
}

```

## Pytania

**1. Skąd wziąć ID mapy?** - Uruchom maps_exporter.exe lub skorzystaj z wygenerowanego już [pliku](https://github.com/Kichiyaki/margonem-mini-bot/blob/master/maps.json).

**2. Skąd wziąć ID postaci?** - Wejdź do gry interesującą cię postacią w przeglądarce, włącz narzędzia dla developerów (na google znajdziecie poradnik jak), przejdź do konsoli i wpisz window.hero.id, enter.

Pojawiły Ci się w głowie jeszcze jakieś inne pytania? Proszę je kierować [tutaj](https://github.com/Kichiyaki/margonem-mini-bot/issues).

## Błędy

Znalazłeś jakiś błąd? Zgłoś go [tutaj](https://github.com/Kichiyaki/margonem-mini-bot/issues).

## Development

**UWAGA!** setup jest przygotowany tylko pod windowsa.

### Co jest potrzebne

- Golang
- windres
- Cmder/inny emulator konsoli/command prompt

### Jak uruchomić aplikację lokalnie

1. Sklonuj te repozytorium.
2. Przejdź do odpowiedniego folderu w Cmderze/command prompcie/innym emulatorze konsoli/eksploratorze plików.
3. Uruchom plik dev.bat.

### Jak zbudować plik .exe

1. Przejdź do odpowiedniego folderu w Cmderze/command prompcie/innym emulatorze konsoli/eksploratorze plików.
2. Uruchom plik build_windows.bat.

## Built with

- [colly](https://github.com/gocolly/colly) - Used to make HTTP requests to Margonem Mini API.
- [robfig/cron](https://github.com/robfig/cron) - Used to perform scheduled tasks.
