package scmlogic

import (
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
)

type InventDimHelper struct {
	db       *datahub.Hub
	ItemSpec tenantcoremodel.ItemSpec
}

type InventDimHelperOpt struct {
	DB  *datahub.Hub
	SKU string
}

func NewInventDimHelper(opt ...InventDimHelperOpt) *InventDimHelper {
	o := new(InventDimHelper)

	if len(opt) == 0 {
		return o
	}

	o.db = opt[0].DB
	if opt[0].SKU != "" {
		itemSpecs := sebar.NewMapRecordWithORM(o.db, new(tenantcoremodel.ItemSpec))
		if spec, e := itemSpecs.Get(opt[0].SKU); e == nil {
			o.ItemSpec = *spec
		}
	}

	return o
}

func (o InventDimHelper) TernaryInventDimension(idimSource *scmmodel.InventDimension, idimCopies ...*scmmodel.InventDimension) *scmmodel.InventDimension {
	valM := map[string][]string{}
	for _, d := range idimCopies {
		valM["VariantID"] = append(valM["VariantID"], d.VariantID)
		valM["Size"] = append(valM["Size"], d.Size)
		valM["Grade"] = append(valM["Grade"], d.Grade)
		valM["WarehouseID"] = append(valM["WarehouseID"], d.WarehouseID)
		valM["AisleID"] = append(valM["AisleID"], d.AisleID)
		valM["SectionID"] = append(valM["SectionID"], d.SectionID)
		valM["BoxID"] = append(valM["BoxID"], d.BoxID)
		valM["BatchID"] = append(valM["BatchID"], d.BatchID)
		valM["SerialNumber"] = append(valM["SerialNumber"], d.SerialNumber)
	}

	res := &scmmodel.InventDimension{
		VariantID:    tenantcorelogic.TernaryString(append([]string{idimSource.VariantID}, valM["VariantID"]...)...),
		Size:         tenantcorelogic.TernaryString(append([]string{idimSource.Size}, valM["Size"]...)...),
		Grade:        tenantcorelogic.TernaryString(append([]string{idimSource.Grade}, valM["Grade"]...)...),
		WarehouseID:  tenantcorelogic.TernaryString(append([]string{idimSource.WarehouseID}, valM["WarehouseID"]...)...),
		AisleID:      tenantcorelogic.TernaryString(append([]string{idimSource.AisleID}, valM["AisleID"]...)...),
		SectionID:    tenantcorelogic.TernaryString(append([]string{idimSource.SectionID}, valM["SectionID"]...)...),
		BoxID:        tenantcorelogic.TernaryString(append([]string{idimSource.BoxID}, valM["BoxID"]...)...),
		BatchID:      tenantcorelogic.TernaryString(append([]string{idimSource.BatchID}, valM["BatchID"]...)...),
		SerialNumber: tenantcorelogic.TernaryString(append([]string{idimSource.SerialNumber}, valM["SerialNumber"]...)...),
	}

	res.VariantID = o.ItemSpec.SpecVariantID
	res.Size = o.ItemSpec.SpecSizeID
	res.Grade = o.ItemSpec.SpecGradeID
	res.Calc()

	return res
}
