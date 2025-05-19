# CLI

If you run:
```sh
cargo run
```

Then the Tolliver CLI will start, which lets you connect to Tolliver programs in an interactive way. To connect to `127.0.0.1:8888` for example, start with:
```tolliver
connect 127.0.0.1:8888 0000000000000000000000000000000000000000000000000000000000000000
```

and then to send a message do something like:
```tolliver
send src/proto_files/items.proto (let ((color red) (size 1)))
```
