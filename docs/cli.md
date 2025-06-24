# CLI

If you run:
```sh
cargo run
```

Then the Tolliver CLI will start, which lets you interact with Tolliver programs. To start a server at `127.0.0.1:8888` do:
```tolliver
start 127.0.0.1:8888
```

Now if you run the CLI in a different terminal, you can connect to it and send a message:
```tolliver
connect 127.0.0.1:8888 0000000000000000000000000000000000000000000000000000000000000000
send proto_files/items.proto Shirt color: "Red" size: LARGE
```

If we switch back to the server, we can receive the message sent:
```tolliver
receive proto_files/items.proto Shirt
```

## Additional commands

### id_send

This is similar to send except you also specify a proto id after the message name. So if we want to use a proto id of 3 we could do:
```tolliver
id_send proto_files/items.proto Shirt 3 color: "Red" size: LARGE
```
