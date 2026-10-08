# ah-be-mcp (Nederlands)

Laat Claude voor je zoeken op ah.be, de bonus bekijken en je winkelmandje beheren. Bijvoorbeeld: "Zet de ingrediënten voor een high-protein maaltijd voor 2 personen in mijn winkelmandje."

> **Onofficieel.** Dit project heeft niets te maken met Albert Heijn. Het gebruikt dezelfde verborgen koppeling als de AH-app. AH kan die op elk moment wijzigen of blokkeren. Gebruik op eigen risico.

## Installeren in Claude Desktop

1. Ga naar de [nieuwste release](https://github.com/FreDeRoover/ah-be-mcp/releases/latest) en download het `.mcpb`-bestand voor jouw computer:
   - Mac met M1, M2, M3 of nieuwer: `macos-apple-silicon`
   - oudere Mac: `macos-intel`
   - Windows: `windows`
2. Dubbelklik op het bestand (of sleep het in Claude Desktop) en bevestig de installatie.
3. Zeg tegen Claude: "Log in bij Albert Heijn". Er opent een tabblad van ah.be. Log daar in zoals je gewend bent. Je wachtwoord ga je alleen op ah.be zelf in, Claude ziet het nooit.

Klaar. Probeer daarna: "Wat zit er in mijn winkelmandje?" of "Welke kip zit er deze week in de bonus?"

## Wat kan het?

- Producten zoeken en voedingswaarden bekijken
- De bonus van deze week bekijken
- Je winkelmandje op ah.be bekijken, aanvullen, aanpassen of leegmaken
- Je bestellingen en kassabonnen bekijken

Het bestelt of betaalt nooit iets. Je rondt je bestelling altijd zelf af op ah.be.

## Privacy

- Alles draait op je eigen computer. Er is geen server van ons en er wordt niets doorgestuurd.
- Je inlog wordt lokaal bewaard, alleen leesbaar voor jou. Zeg "Log uit bij Albert Heijn" om die te wissen.
- Claude ziet wat je ermee opvraagt (producten, je winkelmandje, kassabonnen) als onderdeel van het gesprek.

## Problemen?

- **Mac zegt dat het bestand niet geopend kan worden of van een onbekende ontwikkelaar is:** de app is niet door Apple ondertekend. Ga naar Systeeminstellingen > Privacy en beveiliging en kies "Toch openen".
- **Windows toont een SmartScreen-waarschuwing:** kies "Meer info" en dan "Toch uitvoeren".
- **Claude zegt dat je niet bent ingelogd:** vraag opnieuw "Log in bij Albert Heijn".
- **Iets anders:** meld het via [GitHub Issues](https://github.com/FreDeRoover/ah-be-mcp/issues).

Meer technische uitleg (andere programma's, bouwen uit de broncode): zie de [Engelse README](README.md).
