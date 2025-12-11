package scmlogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"github.com/ariefdarmawan/datahub"
)

type InventSplitType string

const (
	SplitByItem   InventSplitType = "ByItem"
	SplitBySource InventSplitType = "BySource"
)

type InventSplitOpts struct {
	SplitType       InventSplitType
	CompanyID       string
	SourceType      string
	SourceJournalID string
	SourceLineNo    int
	SourceStatus    string
	ItemID          string
	SpecID          string
	WhsID           string
	InventDimID     string
}

type inventSplit struct {
	db   *datahub.Hub
	opts *InventSplitOpts
}

func NewInventSplit(db *datahub.Hub) *inventSplit {
	s := new(inventSplit)
	s.db = db
	s.opts = new(InventSplitOpts)
	return s
}

func (s *inventSplit) SetOpts(opts *InventSplitOpts) ficologic.TrxSplit[*scmmodel.InventTrx, *InventSplitOpts] {
	s.opts = opts
	return s
}

func (s *inventSplit) Split(qty float64, newStatus string) (sources []*scmmodel.InventTrx, dest []*scmmodel.InventTrx, err error) {
	if s.db == nil {
		err = errors.New("db is mandatory")
		return
	}

	if sources, err = s.getSourceTrxs(); err != nil {
		return
	}

	if _, sources, dest, err = splitInventTrxs(s.db, sources, qty, scmmodel.ItemBalanceStatus(newStatus)); err != nil {
		return
	}

	newSources := []*scmmodel.InventTrx{}
	for _, source := range sources {
		if source.Qty == 0 {
			s.db.Delete(source)
			continue
		}
		if err = s.db.Save(source); err != nil {
			return
		}
		newSources = append(newSources, source)
	}
	sources = newSources

	for _, d := range dest {
		if err = s.db.Save(d); err != nil {
			return
		}

	}
	return
}

func (s *inventSplit) getSourceTrxs() (sources []*scmmodel.InventTrx, err error) {
	if s.opts.SplitType == "" {
		err = fmt.Errorf("split type is mandatory")
		return
	}

	if s.opts.SourceStatus == "" {
		err = fmt.Errorf("source status is mandatory")
		return
	}

	if s.opts.CompanyID == "" {
		err = fmt.Errorf("company id is mandatory")
		return
	}

	var wheres = []*dbflex.Filter{dbflex.Eqs("CompanyID", s.opts.CompanyID, "Status", s.opts.SourceStatus)}
	switch s.opts.SplitType {
	case SplitByItem:
		wheres = append(wheres, dbflex.Eqs("Item._id", s.opts.ItemID))
		if s.opts.InventDimID != "" {
			wheres = append(wheres, dbflex.Eq("InventDim.InventDimID", s.opts.InventDimID))
		} else if s.opts.SpecID != "" {
			wheres = append(wheres, dbflex.Eq("InventDim.SpecID", s.opts.SpecID))
		}
		if s.opts.WhsID != "" {
			wheres = append(wheres, dbflex.Eq("InventDim.WarehouseID", s.opts.WhsID))
		}

	case SplitBySource:
		wheres = append(wheres, dbflex.Eqs(
			"SourceType", s.opts.SourceType,
			"SourceJournalID", s.opts.SourceJournalID,
			"SourceLineNo", s.opts.SourceLineNo))
	}

	sources, err = datahub.Find(
		s.db, new(scmmodel.InventTrx), dbflex.NewQueryParam().
			SetWhere(dbflex.And(wheres...)).
			SetSort("TrxDate"))

	return
}

func splitInventTrx(db *datahub.Hub, orig *scmmodel.InventTrx, splitQty float64, status scmmodel.ItemBalanceStatus) (nSplit float64, splitTrx *scmmodel.InventTrx, err error) {
	if splitQty == 0 {
		err = fmt.Errorf("split qty is mandatory")
		return
	}

	nSplit = splitQty
	if orig.Qty > 0 {
		if splitQty > orig.Qty {
			nSplit = orig.Qty
		}
	} else if orig.Qty < 0 {
		if splitQty < orig.Qty {
			nSplit = orig.Qty
		}
	}

	splitTrx = new(scmmodel.InventTrx)
	*splitTrx = *orig
	splitTrx.ID = ""
	splitTrx.Status = status

	originalAdjustment := orig.AmountAdjustment
	originalPhysAmt := orig.AmountPhysical
	originalFinAmt := orig.AmountFinancial
	originalQty := orig.Qty
	originalTrxQty := orig.TrxQty

	orig.Qty -= nSplit
	ratio := orig.Qty / originalQty
	if orig.Qty == 0 {
		db.Delete(orig)
	} else {
		orig.AmountPhysical = ratio * originalPhysAmt
		orig.AmountFinancial = ratio * originalFinAmt
		orig.AmountAdjustment = ratio * originalAdjustment
		orig.TrxQty = ratio * originalTrxQty
		db.Save(orig)
	}

	splitTrx.Qty = nSplit
	ratio = splitTrx.Qty / originalQty
	splitTrx.AmountPhysical = ratio * originalPhysAmt
	splitTrx.AmountFinancial = ratio * originalFinAmt
	splitTrx.AmountAdjustment = ratio * originalAdjustment
	splitTrx.TrxQty = ratio * originalTrxQty
	db.Save(splitTrx)

	return
}

func splitInventTrxs(db *datahub.Hub, sources []*scmmodel.InventTrx, needToBeSplit float64, status scmmodel.ItemBalanceStatus) (nSplit float64,
	origTrxs []*scmmodel.InventTrx,
	splitTrxs []*scmmodel.InventTrx,
	err error) {
	for _, source := range sources {
		splitQty := needToBeSplit - nSplit
		splittedQty, dest, splitErr := splitInventTrx(db, source, splitQty, status)
		if splitErr != nil {
			err = splitErr
			return
		}

		nSplit += splittedQty
		splitTrxs = append(splitTrxs, dest)

		origTrxs = append(origTrxs, source)
		if nSplit == needToBeSplit {
			break
		}
	}

	return
}

func moreThan(qty1, qty2 float64, checkNegate bool) bool {
	if qty1 < 0 {
		if qty2 > 0 && checkNegate {
			return qty1 < -qty2
		}
		return qty1 < qty2
	}

	if qty2 < 0 && checkNegate {
		return qty1 > -qty2
	}
	return qty1 > qty2
}
