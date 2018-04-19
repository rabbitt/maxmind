package command

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/pmylund/go-cache"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/rabbitt/maxmind/mm"
)

type ServerCommand struct {
	configFile  *mm.Pathname
	database    *mm.Database
	Config      *mm.Configuration
	memCache    *cache.Cache
	serverStart time.Time
	Ui          Ui
}

func (c *ServerCommand) IpLookupHandler(writer http.ResponseWriter, req *http.Request, params httprouter.Params) {
	var ipText string
	// Prepare the response and queue sending the record.
	var cached []byte
	var record interface{} = nil
	var status = "success"
	var message = "OK"

	defer func() {
		var j []byte
		var err error
		if cached != nil {
			j = cached
		} else {
			var data *mm.GeoData = nil

			record, ok := record.(*mm.GeoData)
			if ok {
				data = record
			}

			j, err = ffjson.Marshal(&mm.JsonResponse{
				Status:  status,
				Message: message,
				Data:    data,
			})

			if err != nil {
				c.Ui.Error(err)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		if c.memCache != nil && cached == nil {
			c.memCache.Set(ipText, j, cache.DefaultExpiration)
		}

		writer.Write(j)

		if status != "success" {
			c.Ui.Errorf("failed to handle request for %s; error was: %s\n", ipText, message)
		}
	}()

	// Set headers
	writer.Header().Set("Content-Type", "application/json")
	writer.Header().Set("Last-Modified", c.serverStart.Format(http.TimeFormat))

	ipText = params.ByName("ip")

	ip := net.ParseIP(ipText)
	if ip == nil {
		status = "error"
		message = "unable to decode ip"
		writer.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if c.memCache != nil {
		v, found := c.memCache.Get(ipText)
		if found {
			cached = v.([]byte)
			return
		}
	}

	record, err := c.database.Lookup(ipText)
	if err != nil {
		message = err.Error()
		return
	}
}

func (c *ServerCommand) aliveHandler(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
	writer.Header().Set("Content-Type", "text/plain")
	writer.Header().Set("Last-Modified", c.serverStart.Format(http.TimeFormat))
	writer.Header().Set("X-Uptime", fmt.Sprintf("%s", time.Since(c.serverStart)))

	if request.Method == "GET" {
		fmt.Fprint(writer, "pong")
	}
	return
}

func (c *ServerCommand) startService() int {
	c.Ui.Info("Configuration:")
	c.Ui.Infof("    Bind Address:   [ %s:%d ]\n", c.Config.Ip, c.Config.Port)
	if c.Config.CacheTtl > 0.0 {
		c.Ui.Infof("    Cache TTL:      [ %.2f seconds ]\n", c.Config.CacheTtl)
	} else {
		c.Ui.Info("    Cache TTL:      [ disabled ]")
	}
	c.Ui.Infof("    Worker Threads: [ %d ]\n", c.Config.Threads)
	if c.configFile != nil {
		c.Ui.Infof("    Config File:    [ %s ]\n", c.configFile)
	}
	c.Ui.Infof("    Database File:  [ %s ]\n", c.Config.DbPath)

	c.serverStart = time.Now()
	runtime.GOMAXPROCS(int(c.Config.Threads))

	if c.Config.CacheTtl > 0.0 {
		c.Ui.Infof("Caching enabled; will cache requests for %0.2f seconds\n", c.Config.CacheTtl)
		c.memCache = cache.New(time.Duration(c.Config.CacheTtl)*time.Second, 1*time.Second)
	} else {
		c.Ui.Warn("Caching disabled by configuration")
	}

	var err error
	c.database, err = mm.GetDatabase(c.Config.DbPath.Path())
	if err != nil {
		c.Ui.Fatal(err)
	}

	defer c.database.Close()

	router := httprouter.New()
	router.GET("/ping", c.aliveHandler)
	router.HEAD("/ping", c.aliveHandler)
	router.GET("/ip/:ip", c.IpLookupHandler)

	address := fmt.Sprintf("%s:%d", c.Config.Ip, c.Config.Port)
	c.Ui.Infof("Listening on %s ...\n", address)
	c.Ui.Fatal(http.ListenAndServe(address, router))

	return 0
}

func (c *ServerCommand) Help() string {
	return fmt.Sprintf(`Usage: %s server [options]

Run as a caching HTTP server, responding to requests for /ip/:ip, and /ping.

Options:
  -c, -config-file      <file>         File containing configuration. Note: Command
                                       line options override config file options.
  -i, -ip               <ip address>   IP Address to bind to (default: %s)
  -p, -port             <integer>      Port to bind to (default: %d)
  -f, -database-file    <file>         Path to MaxMind Database
                                       (default: %s)
  -t, -cache-ttl        <float>        How long to cache response data before
                                       refetching it from the database.
                                       (default: %.2f)
  -T, -worker-threads   <integer>      Number of worker threads to handle incoming
                                       requests. (default: %d)

`, os.Args[0], c.Config.Ip, c.Config.Port, c.Config.DbPath, c.Config.CacheTtl, c.Config.Threads)
}

func (c *ServerCommand) Synopsis() string {
	return "Run as a server"
}

func (c *ServerCommand) Run(args []string) int {

	var err error
	var dbPath *mm.Pathname

	c.Config = mm.NewConfiguration()

	var configParse = flag.NewFlagSet("server", flag.ContinueOnError)
	configParse.SetOutput(&mm.NullWriter{})

	configPath := configParse.String("config.file", "", "`Path` to config file ")
	configParse.StringVar(configPath, "c", "", "`Path` to config file ")
	configParse.Parse(args)

	if *configPath != "" {
		if cfgPath, err := mm.NewPathname(*configPath).RealPath(); err == nil {
			c.Config.LoadFromJsonFile(cfgPath)
			c.configFile = cfgPath
		} else {
			c.Ui.Fatal(err)
		}
	}

	mainParse := flag.NewFlagSet("server", flag.ContinueOnError)
	_ = mainParse.String("config.file", "", "`path` to config file ")
	ip := mainParse.String("server.ip", c.Config.Ip, "server `IP` address; empty to bind all interfaces")
	port := mainParse.Int("server.port", int(c.Config.Port), "server `port`")
	dbFile := mainParse.String("database.file", c.Config.DbPath.Path(), "`path` to the database file that contains GeoIP information")
	cacheTtl := mainParse.Float64("cache.ttl", float64(c.Config.CacheTtl), "How many `seconds` should requests be cached. Set to 0 to disable")
	threads := mainParse.Int("worker.threads", int(c.Config.Threads), "Number of `threads` to use. Defaults to number of detected cores")

	mainParse.StringVar(configPath, "c", "", "`path` to config file ")
	mainParse.StringVar(ip, "i", c.Config.Ip, "server `IP` address; empty to bind all interfaces")
	mainParse.IntVar(port, "p", int(c.Config.Port), "server `port`")
	mainParse.StringVar(dbFile, "f", c.Config.DbPath.Path(), "`path` to the database file that contains GeoIP information")
	mainParse.Float64Var(cacheTtl, "t", float64(c.Config.CacheTtl), "How many `seconds` should requests be cached. Set to 0 to disable")
	mainParse.IntVar(threads, "T", int(c.Config.Threads), "Number of `threads` to use. Defaults to number of detected cores")

	mainParse.Usage = func() {
		c.Ui.Output(c.Help())
		mainParse.PrintDefaults()
	}
	mainParse.Parse(args)

	if *ip != c.Config.Ip {
		c.Config.Ip = *ip
	}
	if uint32(*port) != c.Config.Port {
		c.Config.Port = uint32(*port)
	}
	if float64(*cacheTtl) != c.Config.CacheTtl {
		c.Config.CacheTtl = float64(*cacheTtl)
	}
	if uint8(*threads) != c.Config.Threads {
		c.Config.Threads = uint8(*threads)
	}

	if c.Config.Threads < 1 {
		c.Ui.Fatal("Worker threads must be at least 1!")
	}

	if *dbFile == "" {
		c.Ui.Fatal("missing required path to MasterMind DB file")
	} else {
		dbPath, err = mm.NewPathname(*dbFile).RealPath()
		if err != nil {
			c.Ui.Fatal(err)
		}
	}

	if *dbFile == "" {
		c.Ui.Fatal("missing required path to MasterMind DB file")
	} else {
		// normalize the database file path
		dbPath, err = mm.NewPathname(*dbFile).RealPath()
		if err != nil {
			c.Ui.Fatal(err)
		}
		c.Config.DbPath = dbPath
	}

	return c.startService()
}
