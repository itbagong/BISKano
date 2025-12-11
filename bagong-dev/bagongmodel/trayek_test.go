package bagongmodel_test

import (
	"fmt"
	"testing"

	"git.kanosolution.net/sebar/bagong/bagongmodel"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetTariff(t *testing.T) {
	type Expectation struct {
		From  string
		To    string
		Tarif float64
	}

	Convey("Given Trayek Data", t, func() {
		cases := []struct {
			data      bagongmodel.Trayek
			expecteds []Expectation
		}{
			{
				data: bagongmodel.Trayek{
					ID:        "TEST001",
					Name:      "Trayek Test 001",
					Terminals: []string{"Arjosari", "Purworejo", "Pandaan", "Bungurasih"},
					Tarifs: []bagongmodel.Tarif{
						{
							From: "Arjosari",
							To:   "Purworejo",
							Rate: 10000,
						},
						{
							From: "Arjosari",
							To:   "Bungurasih",
							Rate: 20000,
						},
						{
							From: "Purworejo",
							To:   "Arjosari",
							Rate: 10000,
						},
						{
							From: "Purworejo",
							To:   "Pandaan",
							Rate: 5000,
						},
						{
							From: "Purworejo",
							To:   "Bungurasih",
							Rate: 15000,
						},
						{
							From: "Pandaan",
							To:   "Purworejo",
							Rate: 5000,
						},
						{
							From: "Pandaan",
							To:   "Bungurasih",
							Rate: 10000,
						},
						{
							From: "Bungurasih",
							To:   "Arjosari",
							Rate: 20000,
						},
						{
							From: "Bungurasih",
							To:   "Arjosari",
							Rate: 20000,
						},
						{
							From: "Bungurasih",
							To:   "Purworejo",
							Rate: 15000,
						},
						{
							From: "Bungurasih",
							To:   "Pandaan",
							Rate: 10000,
						},
					},
				},
				expecteds: []Expectation{
					{
						From:  "Arjosari",
						To:    "Pandaan",
						Tarif: 10000,
					},
					{
						From:  "Pandaan",
						To:    "Arjosari",
						Tarif: 10000,
					},
					{
						From:  "Bungurasih",
						To:    "Pandaan",
						Tarif: 10000,
					},
				},
			},
			{
				data: bagongmodel.Trayek{
					ID:        "TEST002",
					Name:      "Trayek Test 02",
					Terminals: []string{"Arjosari", "Purworejo", "Pandaan", "Bungurasih"},
					Tarifs: []bagongmodel.Tarif{
						{
							From: "Arjosari",
							To:   "Purworejo",
							Rate: 0,
						},
						{
							From: "Arjosari",
							To:   "Bungurasih",
							Rate: 20000,
						},
						{
							From: "Purworejo",
							To:   "Arjosari",
							Rate: 7000,
						},
						{
							From: "Purworejo",
							To:   "Pandaan",
							Rate: 0,
						},
						{
							From: "Purworejo",
							To:   "Bungurasih",
							Rate: 15000,
						},
						{
							From: "Pandaan",
							To:   "Purworejo",
							Rate: 0,
						},
						{
							From: "Pandaan",
							To:   "Bungurasih",
							Rate: 10000,
						},
						{
							From: "Bungurasih",
							To:   "Arjosari",
							Rate: 20000,
						},
						{
							From: "Bungurasih",
							To:   "Arjosari",
							Rate: 20000,
						},
						{
							From: "Bungurasih",
							To:   "Purworejo",
							Rate: 15000,
						},
						{
							From: "Bungurasih",
							To:   "Pandaan",
							Rate: 10000,
						},
					},
				},
				expecteds: []Expectation{
					{
						From:  "Pandaan",
						To:    "Purworejo",
						Tarif: 7000,
					},
				},
			},
		}

		for _, c := range cases {
			for _, expected := range c.expecteds {
				conv := fmt.Sprintf("Using Trayek Data with ID '%s', Get Tariff From: %s, To: %s", c.data.ID, expected.From, expected.To)
				Convey(conv, func() {
					fmt.Printf("\n\n%s\n", conv)
					tariffResult, e := c.data.GetTariff(expected.From, expected.To, true)

					fmt.Printf("Assert Tariff is OK: ")
					So(tariffResult, ShouldEqual, expected.Tarif)
					fmt.Printf("\nAssert Error is OK: ")
					So(e, ShouldBeNil)
				})
			}
		}
	})
}
