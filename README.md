# programmfabrik - exiftool2json

Process the output of `exiftool -listx` command converting that to a valid JSON
document.

## Building it

Just invoke `go build .`

## Running it

The app will perform some self-diagnose tests to verify if the required tool is
available `exiftool`, if not, a message will be printed and the application will
abort without any further action.

After that it will try to setup a webserver at `0.0.0.0:8080`, you can change the
parameters using command line flags. Use `-h` for more details

## Closing it

The app will run until it receives a SIGTERM (Ctrl+C in any UNIX)

## Working on it

If you have `make` and `modd` installed, use `make watch` to recompile the app
everytime a `.go` file is changed. Otherwise, just call `go build` from time.
