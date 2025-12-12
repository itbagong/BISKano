package ntsllogic_test

import (
	"fmt"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/nitasalu/ntslmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/samber/lo"
)

func InjectCoreData(ctx *kaos.Context) error {
	db, _ := ctx.DefaultHub()

	users := lo.RepeatBy(1000, func(index int) *ntslmodel.Profile {
		return &ntslmodel.Profile{
			ID:    fmt.Sprintf("user-%04d", index),
			Name:  fmt.Sprintf("Name %04d", index),
			Email: fmt.Sprintf("ariefda+user%04d@hotmail.com", index),
		}
	})

	models := [][]orm.DataModel{
		sebar.ToDataModels([]*ntslmodel.City{
			{ID: "JKT", Name: "Jakarta", Depart: true},
			{ID: "SBY", Name: "Surabaya", Depart: true},
			{ID: "SAM", Name: "Samarinda", Depart: true},
			{ID: "MED", Name: "Medan", Depart: true},
			{ID: "JED", Name: "Jeddah", Trip: true},
			{ID: "MDN", Name: "Madinah", Trip: true},
			{ID: "RYD", Name: "Riyadh"},
			{ID: "TIF", Name: "Thaif", Trip: true},
			{ID: "ANK", Name: "Turki", Trip: true},
		}),

		sebar.ToDataModels([]*ntslmodel.Feature{
			{ID: "FlightGaruda", Name: "Garuda", FeatureType: ntslmodel.FeatureFlight},
			{ID: "FlightSaudia", Name: "Saudia", FeatureType: ntslmodel.FeatureFlight},
			{ID: "FlightAirAsia", Name: "AirAsia", FeatureType: ntslmodel.FeatureFlight},
			{ID: "FlightPremium", Name: "Premium Flight", FeatureType: ntslmodel.FeatureFlight},
			{ID: "FlightLowCost", Name: "Low Cost Flight", FeatureType: ntslmodel.FeatureFlight},
			{ID: "TourMecca", Name: "Tour Mecca", FeatureType: ntslmodel.FeatureTour, CityID: "MEC"},
			{ID: "TourMedina", Name: "Tour Medina", FeatureType: ntslmodel.FeatureTour, CityID: "MED"},
			{ID: "OnlyMeccaPilgrimage", Name: "Only Mecca pilgrimage", FeatureType: ntslmodel.FeatureActivity},
			{ID: "OnlyMedinaPilgrimage", Name: "Only Medina pilgrimage", FeatureType: ntslmodel.FeatureActivity},
		}),

		sebar.ToDataModels(users),
	}

	for _, ms := range models {
		for _, m := range ms {
			if err := db.Insert(m); err != nil {
				return fmt.Errorf("save %s: %s: %s", m.TableName(), modelID(m), err.Error())
			}
		}
	}

	return nil
}
