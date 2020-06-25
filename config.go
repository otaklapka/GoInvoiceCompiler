package main

import (
	"fmt"
	"github.com/pkg/errors"
	"time"
)

type Config struct {
	IssuedDate string `yaml:"datum_vystaveni"`
	DueDate string `yaml:"datum_splatnosti"`
	BankAccount string `yaml:"cislo_uctu"`
	variableSymbol string `yaml:"variabilni_symbol"`
	Iban string
	AccountingEntity Entity `yaml:"ucetni_jednotka"`
	Customer Entity `yaml:"zakaznik"`
	Items []Item `yaml:"polozky"`
}

type Entity struct {
	Name string `yaml:"nazev"`
	Address string `yaml:"adresa"`
	City string `yaml:"mesto"`
	Zip string `yaml:"psc"`
	Id string `yaml:"ic"`
	VatId string `yaml:"dic"`
	IsVatPayer bool `yaml:"je_platcem_dph"`
}

type Item struct {
	Description string `yaml:"popis"`
	UnitPrice float64 `yaml:"jednotkova_cena"`
	Quantity float64 `yaml:"mnozstvi"`
}

func (it *Item) Total() float64 {
	return it.UnitPrice * it.Quantity
}

func (config *Config) Total() float64 {
	var sum float64
	for _, item := range config.Items {
		sum += item.Total()
	}
	return sum
}

func (config *Config) GetVariableSymbol() string {
	if config.variableSymbol == "" {
		return config.Serial()
	} else {
		return config.variableSymbol
	}
}

func (config *Config) QRString() (error, string) {
	if config.Iban == "" {
		return errors.New("Pro vygenerovani QR retezce je treba IBAN"), ""
	}
	return nil, fmt.Sprintf("SPD*1.0*ACC:%s*AM:%.2f*CC:CZK*X-VS:%s", config.Iban, config.Total(), config.GetVariableSymbol())
}

func (config *Config) Serial() string {
	now := time.Now()
	t, err := time.Parse("02.01.2006", config.IssuedDate)
	if err != nil {
		return now.Format("20060201150405")
	} else {
		return time.Date(t.Year(), t.Month(), t.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location()).Format("20060201150405")
	}
}
