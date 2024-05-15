# tacopro-reader
European Tachograph reader

## Help needed

If you have a gen 2 card, please contact me to continue the development and improve the tool.

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

## Troubleshoting

`My card reader is not displayed in the list`: Try to insert the card and refresh.

`Fyne error: window creation error Cause: APIUnavailable: WGL: The driver does not appear to support OpenGL`: Download the opengl32.dll from https://fdossena.com/?p=mesa/index.frag and place it in the same folder as the exe file.

`Error: can't list readers: SCARD_E_SERVICE_STOPPED`: This is a Windows error handling the smart cards readers. Try to reboot and run the tool again.

## References

- https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX%3A02016R0799-20230821&qid=1701849521342