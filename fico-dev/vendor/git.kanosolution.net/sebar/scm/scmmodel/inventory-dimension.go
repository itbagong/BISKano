package scmmodel

import (
	"fmt"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"github.com/sebarcode/codekit"
)

type InventDimensionSpec struct {
	VariantID string
	Size      string
	Grade     string
	SpecID    string
}

type InventDimensionLocation struct {
	WarehouseID string `label:"Warehouse" form_lookup:"/tenant/warehouse/find|_id|_id,Name"`
	AisleID     string `label:"Aisle" form_lookup:"/tenant/aisle/find|_id|_id,Name"`
	SectionID   string `label:"Section" form_lookup:"/tenant/section/find|_id|_id,Name"`
	BoxID       string `label:"Box" form_lookup:"/tenant/box/find|_id|_id,Name"`
}

type InventDimensionProduct struct {
	BatchID      string
	SerialNumber string
}

type InventDimension struct {
	VariantID    string
	Size         string
	Grade        string
	WarehouseID  string `label:"Warehouse" form_lookup:"/tenant/warehouse/find|_id|_id,Name"`
	AisleID      string `label:"Aisle" form_lookup:"/tenant/aisle/find|_id|_id,Name"`
	SectionID    string `label:"Section" form_lookup:"/tenant/section/find|_id|_id,Name"`
	BoxID        string `label:"Box" form_lookup:"/tenant/box/find|_id|_id,Name"`
	BatchID      string
	SerialNumber string
	SpecID       string
	InventDimID  string
}

type InventDim struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Dimension         InventDimension
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *InventDim) TableName() string {
	return "InventoryDimensionss"
}

func (o *InventDim) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *InventDim) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *InventDim) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *InventDim) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *InventDim) PreSave(dbflex.IConnection) error {
	o.Dimension.Calc()
	o.ID = o.Dimension.InventDimID
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *InventDim) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *InventDimension) SpecHash() string {
	return codekit.MD5String(fmt.Sprintf("%s|%s|%s", o.VariantID, o.Size, o.Grade))
}

func (o *InventDimension) DimIdHash() string {
	fieldToHash := []string{
		o.VariantID,
		o.Size,
		o.Grade,
		o.WarehouseID,
		o.AisleID,
		o.SectionID,
		o.BoxID,
		o.BatchID,
		o.SerialNumber}

	return codekit.MD5String(strings.Join(fieldToHash, "|"))
}

func (o *InventDimension) Calc() *InventDimension {
	o.SpecID = o.SpecHash()
	o.InventDimID = o.DimIdHash()

	return o
}

func (dim *InventDimension) Where(dimField string) *dbflex.Filter {
	var fs []*dbflex.Filter
	fs = addFilters(fs, dimField+".VariantID", dim.VariantID, false)
	fs = addFilters(fs, dimField+".Size", dim.Size, false)
	fs = addFilters(fs, dimField+".Grade", dim.Grade, false)
	fs = addFilters(fs, dimField+".BatchID", dim.BatchID, false)
	fs = addFilters(fs, dimField+".SerialNumber", dim.SerialNumber, false)
	fs = addFilters(fs, dimField+".WarehouseID", dim.WarehouseID, false)
	fs = addFilters(fs, dimField+".SectionID", dim.SectionID, false)
	fs = addFilters(fs, dimField+".AisleID", dim.AisleID, false)
	fs = addFilters(fs, dimField+".BoxID", dim.BoxID, false)

	return dbflex.And(fs...)
}

// TernaryInventDimension deprecated
// use this: NewInventDimHelper(InventDimHelperOpt{DB: p.opt.Db, SKU: line.SKU}).TernaryInventDimension(&line.InventDim, &p.header.InventDim)
func TernaryInventDimension(idimSource *InventDimension, idimCopies ...*InventDimension) *InventDimension {
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

	res := &InventDimension{
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
	res.Calc()

	return res
}

func addFilters(fs []*dbflex.Filter, fieldName string, value string, addIfNilOrEmpty bool) []*dbflex.Filter {
	if fs == nil {
		fs = []*dbflex.Filter{}
	}

	if value == "" && !addIfNilOrEmpty {
		return fs
	}

	fs = append(fs, dbflex.Eq(fieldName, value))
	return fs
}
