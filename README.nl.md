# ah-be-mcp (Nederlands)

Laat Claude voor je zoeken op ah.be, de bonus bekijken en je winkelmandje beheren. Bijvoorbeeld: "Zet de ingrediënten voor een high-protein maaltijd voor 2 personen in mijn winkelmandje."

> Onofficieel. Dit project heeft niets te maken met Albert Heijn en gebruikt dezelfde verborgen koppeling als de AH-app. AH kan die op elk moment wijzigen of blokkeren. Gebruik op eigen risico.

## Installeren

Het werkt met Claude Desktop op Mac en Windows. ChatGPT en Chromebook werken niet, omdat die geen programma op je eigen computer kunnen starten.

1. Download het `.mcpb`-bestand voor jouw computer van de [nieuwste release](https://github.com/FreDeRoover/ah-be-mcp/releases/latest): `macos-apple-silicon` (Mac met M1 of nieuwer), `macos-intel` (oudere Mac) of `windows`.
2. Dubbelklik erop, of sleep het in Claude Desktop, en bevestig.
3. Zeg tegen Claude: "Log in bij Albert Heijn". Er opent een tabblad van ah.be waar je inlogt zoals je gewend bent. Je wachtwoord geef je alleen aan ah.be, Claude ziet het niet.

Probeer daarna: "Wat zit er in mijn winkelmandje?" of "Welke kip zit er deze week in de bonus?"

Claude kan producten zoeken, de bonus bekijken, je winkelmandje aanpassen of leegmaken en je bestellingen en kassabonnen tonen. Er wordt nooit iets besteld of betaald. Dat doe je zelf op ah.be.

## Privacy

Alles draait op je eigen computer, er is geen server van ons. Je inlog staat lokaal en alleen jij kunt hem lezen. Zeg "Log uit bij Albert Heijn" om hem te wissen. Wat je opvraagt, zoals je winkelmandje of kassabonnen, wordt onderdeel van het gesprek met Claude.

## Problemen

- Mac of Windows waarschuwt voor een onbekende ontwikkelaar: de app is niet ondertekend. Kies op Mac "Toch openen" onder Systeeminstellingen > Privacy en beveiliging, op Windows "Meer info" en dan "Toch uitvoeren".
- Claude zegt dat je niet bent ingelogd: vraag opnieuw "Log in bij Albert Heijn".
- Iets anders: meld het via [GitHub Issues](https://github.com/FreDeRoover/ah-be-mcp/issues).

Technische uitleg staat in de [Engelse README](README.md).
