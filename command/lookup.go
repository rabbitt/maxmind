package command

import (
	"flag"
	"fmt"
	"os"

	"github.com/pquerna/ffjson/ffjson"
	"github.com/rabbitt/maxmind/mm"
)

type LookupCommand struct {
	Ui Ui
}

var OutputTypes = map[string]bool{
	"json":  true,
	"table": true,
}

func (c *LookupCommand) outputAsJson(record *mm.GeoData) {
	j, err := ffjson.Marshal(record)
	if err != nil {
		c.Ui.Error(err)
		return
	}
	c.Ui.Output(string(j))
}

func (c *LookupCommand) outputAsTable(ipText string, record *mm.GeoData) {
	c.Ui.Output("")
	c.Ui.Outputf("[ %15s ]---------------------->\n", ipText)
	if record.Unknown() {
		c.Ui.Error("  Unable to find any valid data")
		return
	}

	if record.Continent.Name != "" {
		c.Ui.Outputf("  Continent:      [%s (%s)]\n", record.Continent.Name, record.Continent.Code)
	}

	if record.Country.Name != "" {
		c.Ui.Outputf("  Country:        [%s (%s)]\n", record.Country.Name, record.Country.IsoCode)
	}

	if len(record.Subdivisions) > 0 {
		if len(record.Subdivisions) == 1 {
			c.Ui.Outputf("  Subdivision:    [%s (%s)]\n", record.Subdivisions[0].Name, record.Subdivisions[0].IsoCode)
		} else if len(record.Subdivisions) > 1 {
			c.Ui.Output("  Subdivisions: ")
			for idx, sub := range record.Subdivisions {
				c.Ui.Outputf("    %02d:           [%s (%s)]\n", idx, sub.Name, sub.IsoCode)
			}
		}
	}

	if record.City.Name != "" {
		c.Ui.Outputf("  City:           [%s]\n", record.City.Name)
	}

	if record.Postal.Code != "" {
		c.Ui.Outputf("  Postal:         [%05s]\n", record.Postal.Code)
	}

	if record.Location.AccuracyRadius != 0 {
		c.Ui.Output("  Location:")
		c.Ui.Outputf("    Coordinates:  [%.4f, %.4f (%d)]\n", record.Location.Latitude, record.Location.Longitude, record.Location.AccuracyRadius)
		if record.Location.TimeZone != "" {
			c.Ui.Outputf("    Timezone:     [%s]\n", record.Location.TimeZone)
		}
	}
}

func (c *LookupCommand) Help() string {
	return fmt.Sprintf(`Usage: %s lookup [options] ip ... ip

Prints the details of the requested IPs, responding to requests for /ip/:ip, and /ping.

Options:
  -f, -database.file  <file>      Path to MaxMind Database
                                  (default: %s)
  -o, -output.type    <string>    Render mode (one of: json, or table)
                                  requests. (default: %s)
`, os.Args[0], DefaultDatabasePath, "table")
}

func (c *LookupCommand) Synopsis() string {
	return "Lookup on or more IPs and exit"
}

func (c *LookupCommand) Run(args []string) int {
	var dbPath *mm.Pathname
	var database *mm.Database
	var err error

	var mainParse = flag.NewFlagSet("lookup", flag.ContinueOnError)
	outType := mainParse.String("o", "table", "Output `type` for quick lookup; either 'json' or 'table'")
	mainParse.StringVar(outType, "output.type", "table", "Output `type` for quick lookup; either 'json' or 'table'")
	dbFile := mainParse.String("f", DefaultDatabasePath, "`path` to the database file that contains GeoIP information")
	mainParse.StringVar(dbFile, "database.file", DefaultDatabasePath, "`path` to the database file that contains GeoIP information")

	mainParse.Usage = func() {
		c.Ui.Output(c.Help())
		mainParse.PrintDefaults()
	}
	mainParse.Parse(args)

	if outType != nil && OutputTypes[*outType] == false {
		c.Ui.Fatalf("Invalid output type '%s'; expected one of 'json', or 'table'\n", *outType)
	}

	if *dbFile == "" {
		c.Ui.Fatal("missing required path to MasterMind DB file")
	} else {
		dbPath, err = mm.NewPathname(*dbFile).RealPath()
		if err != nil {
			c.Ui.Fatal(err)
		}
	}

	if len(mainParse.Args()) <= 0 {
		c.Ui.Fatal("no ips to look up")
	}

	if database, err = mm.GetDatabase(dbPath.Path()); err == nil {
		defer mm.CloseDatabases()
	} else {
		c.Ui.Fatal(err)
	}

	for _, ip := range mainParse.Args() {
		record, err := database.Lookup(ip)
		if err != nil {
			c.Ui.Fatal(err)
		}

		switch *outType {
		case "json":
			c.outputAsJson(record)
		case "table":
			c.outputAsTable(ip, record)
		}
	}

	return 0
}
