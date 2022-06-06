# Instructions

* open two terminals
* `vagrant ssh peer` and `vagrant ssh server`
* `./host-create.sh <peer|server>`
* `./container-create.sh <peer|server>`

Then ping or curl or whatever between them. Using something like echo-server would make it easier to validate.

This example disallows `container0` on `server` (10.0.10.3) from talking to `container0` on `peer` (10.0.0.3)