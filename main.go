package main

import (
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"github.com/fatih/color"
)

func main() {
	configFilePath := flag.String("c", "config.yaml", "Cesta k YAML souboru konfigurace")
	flag.Parse()

	if *configFilePath == "" {
		color.Red("Je třeba zadat platný soubor konfigurace")
		return
	}

	configFile, err := ioutil.ReadFile(*configFilePath)
	if err != nil {
		color.Red("Nepodařilo se najít soubor konfigurace: %s\n", err)
		return
	}

	var config Config
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		color.Red("Chyba při čtení souboru: %v", err)
	}

	//fmt.Printf("%#v", config)

	err, invoice := config.NewInvoice()
	err = invoice.compilePdf()
	if err != nil {
		color.Red("Chyba při kompilování pdf: %v", err)
	}

	color.Green("Vytvořena faktura %s", invoice.config.GetVariableSymbol())
}

