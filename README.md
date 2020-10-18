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
every time a `.go` file is changed. Otherwise, just call `go build` from time.


## Future improvements

Currently, there is no limit to how many instances of `exiftool` are running at
any given moment, thus, the service should be executed behind a proxy to add some
form of rate-limit.

- Alternatively, one could implement a `semaphore` in `exif.Tool#WalkTags` using
buffered channels. The size of the channel would indicate the max number of
instances that should be executed at any given moment.

- It might be possible to improve the parsing code so that
`xml.Decoder#DecodeElement` could be used instead. But care must be taken to
avoid using `xml.Decoder#Token` at the root level. Otherwise, the whole XML
will be processed (as that method provides guarantees regarding who tags
are closed)

- `exif.Tool#WalkTags` doesn't use channels by design, otherwise callers
would need to deal with the added concurrency. By relying on a blocking API,
callers have the choice to use it in a sync or async manner.

- Allocation could be reduced in the JSON encoding block by implementing
MarshalJSON for some types, but this might make the code less readable.
