# mcutil
![](https://img.shields.io/github/languages/code-size/mcstatus-io/mcutil)
![](https://img.shields.io/github/issues/mcstatus-io/mcutil)
![](https://img.shields.io/github/license/mcstatus-io/mcutil)

A zero-dependency library for interacting with the Minecraft protocol in Go. Supports retrieving the status of any Minecraft server (Java or Bedrock Edition), querying a server for information, sending remote commands with RCON, and sending Votifier votes. Look at the examples in this readme or search through the documentation instead.

## Installation

```bash
go get github.com/mcstatus-io/mcutil
```

## Documentation

https://pkg.go.dev/github.com/mcstatus-io/mcutil

## Usage

### Status (1.7+)

Retrieves the status of the Java Edition Minecraft server. This method only works on netty servers, which is version 1.7 and above. An attempt to use on pre-netty servers will result in an error.

```go
import (
	"context"
	"fmt"
	"time"

	"github.com/mcstatus-io/mcutil"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	response, err := mcutil.Status(ctx, "play.hypixel.net", 25565)

	if err != nil {
		panic(err)
	}

	fmt.Println(response)
}
```

### Legacy Status (‹ 1.7)

Retrieves the status of the Java Edition Minecraft server. This is a legacy method that is supported by all servers, but only retrieves basic information. If you know the server is running version 1.7 or above, please use `Status()` instead.

```go
import (
	"context"
	"fmt"
	"time"

	"github.com/mcstatus-io/mcutil"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	response, err := mcutil.StatusLegacy(ctx, "play.hypixel.net", 25565)

	if err != nil {
		panic(err)
	}

	fmt.Println(response)
}
```

### Bedrock Status

Retrieves the status of the Bedrock Edition Minecraft server.

```go
import (
	"context"
	"fmt"
	"time"

	"github.com/mcstatus-io/mcutil"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	response, err := mcutil.StatusBedrock(ctx, "127.0.0.1", 19132)

	if err != nil {
		panic(err)
	}

	fmt.Println(response)
}
```

### Basic Query

Performs a basic query lookup on the server, retrieving most information about the server. Note that the server must explicitly enable query for this functionality to work.

```go
package a

import (
	"context"
	"fmt"
	"time"

	"github.com/mcstatus-io/mcutil"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	response, err := mcutil.BasicQuery(ctx, "play.hypixel.net", 25565)

	if err != nil {
		panic(err)
	}

	fmt.Println(response)
}

```

### Full Query

Performs a full query lookup on the server, retrieving all available information. Note that the server must explicitly enable query for this functionality to work.

```go
import (
	"context"
	"fmt"
	"time"

	"github.com/mcstatus-io/mcutil"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	response, err := mcutil.FullQuery(ctx, "play.hypixel.net", 25565)

	if err != nil {
		panic(err)
	}

	fmt.Println(response)
}
```

### RCON

Executes remote console commands on the server. You must know the connection details of the RCON server, as well as the password.

```go
import "github.com/mcstatus-io/mcutil"

func main() {
    client := mcutil.NewRCON()

    if err := client.Dial("127.0.0.1", 25575); err != nil {
        panic(err)
    }

    if err := client.Login("mypassword"); err != nil {
        panic(err)
    }

    if err := client.Run("say Hello, world!"); err != nil {
        panic(err)
    }

    fmt.Println(<- client.Messages)

    if err := client.Close(); err != nil {
        panic(err)
    }
}
```

## Send Vote

Sends a Votifier vote to the specified server, typically used by server listing websites. The host and port must be known of the Votifier server, as well as the token generated by the server. This is for use on servers running Votifier 2 and above, such as [NuVotifier](https://www.spigotmc.org/resources/nuvotifier.13449/).

```go
import (
	"context"
	"time"

	"github.com/mcstatus-io/mcutil"
	"github.com/mcstatus-io/mcutil/options"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	err := mcutil.SendVote(ctx, "127.0.0.1", 8192, options.Vote{
		ServiceName: "my-service",
		Username:    "PassTheMayo",
		Token:       "abc123", // server's Votifier token
		UUID:        "",       // recommended but not required, UUID with dashes
		Timestamp:   time.Now(),
		Timeout:     time.Second * 5,
	})

	if err != nil {
		panic(err)
	}
}
```

## Send Legacy Vote

Sends a legacy Votifier vote to the specified server, typically used by server listing websites. The host and port must be known of the Votifier server, as well as the public key generated by the server. This is for use on servers running Votifier 1, such as the original [Votifier](https://dev.bukkit.org/projects/votifier) plugin. It is impossible to tell whether the vote was successfully processed by the server because Votifier v1 protocol does not return any data.

```go
import (
	"context"
	"time"

	"github.com/mcstatus-io/mcutil"
	"github.com/mcstatus-io/mcutil/options"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	err := mcutil.SendLegacyVote(ctx, "127.0.0.1", 8192, options.LegacyVote{
		PublicKey:   "...", // the contents of the 'plugins/<Votifier>/rsa/public.key' file on the server
		ServiceName: "my-service",
		Username:    "PassTheMayo",
		IPAddress:   "127.0.0.1",
		Timestamp:   time.Now(),
		Timeout:     time.Second * 5,
	})

	if err != nil {
		panic(err)
	}
}
```

## License

[MIT License](https://github.com/mcstatus-io/mcutil/blob/main/LICENSE)