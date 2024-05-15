# tacopro-reader
European Tachograph reader

## Linux dev setup

Ensure you can run the Fyne demo:

```
go run fyne.io/fyne/v2/cmd/fyne_demo@latest
```

## Build

Run `fyne-cross` to build for other platforms from a Docker host:

```
cd tacopro-reader
go run github.com/fyne-io/fyne-cross@latest windows -app-id es.tacopro.reader .
go run github.com/fyne-io/fyne-cross@latest linux .
```

## References

- https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX%3A02016R0799-20230821&qid=1701849521342