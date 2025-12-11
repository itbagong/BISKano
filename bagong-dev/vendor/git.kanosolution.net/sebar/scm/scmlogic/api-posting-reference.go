package scmlogic

import (
	"errors"
	"sort"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
)

type JournalReference struct {
	JournalType string
	JournalID   string
}

type FindJournalRefResponse struct {
	JournalType string
	JournalID   string
	Refferences []JournalReference
}

func (pph *PostingProfileHandlerV2) FindJournalRef(ctx *kaos.Context, req *JournalReference) (*FindJournalRefResponse, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: db")
	}

	reffNos := []string{req.JournalID}
	reffs := []JournalReference{{JournalType: req.JournalType, JournalID: req.JournalID}}

	// currentJournalReffs := getReffNo(h, req.JournalType, req.JournalID)
	// if len(currentJournalReffs) > 0 {
	// 	reffs = append(reffs, currentJournalReffs...)
	// 	reffNos = append(reffNos, lo.Map(currentJournalReffs, func(ref JournalReference, i int) string {
	// 		return ref.JournalID
	// 	})...)
	// }

	sequence := []string{
		// TODO: WO
		string(scmmodel.ItemRequestType),
		string(scmmodel.PurchRequest),
		string(scmmodel.PurchOrder),
		string(scmmodel.JournalTransfer),
		string(scmmodel.InventIssuance),
		string(scmmodel.InventReceive),
	}

	_, idx, found := lo.FindIndexOf(sequence, func(seq string) bool {
		return seq == req.JournalType
	})
	if !found {
		return nil, errors.New("invalid journal type")
	}

	journalObjectMap := getJournalObjectMap(h)

	// backwards loop
	for i := idx; i >= 0; i-- {
		journalType := sequence[i]
		jo, found := journalObjectMap[journalType]
		if !found {
			return nil, errors.New("journal object configuration not found")
		}

		reffs = append(reffs, jo.GetReffByID(reffNos)...)
		reffNos = lo.Map(reffs, func(d JournalReference, i int) string {
			return d.JournalID
		})
	}

	// forward loop
	for i := idx + 1; i < len(sequence); i++ {
		journalType := sequence[i]
		jo, found := journalObjectMap[journalType]
		if !found {
			return nil, errors.New("journal object configuration not found")
		}

		reffs = append(reffs, jo.GetReffByReff(reffNos)...)
		reffNos = lo.Map(reffs, func(d JournalReference, i int) string {
			return d.JournalID
		})
	}

	reffs = lo.Filter(reffs, func(d JournalReference, i int) bool {
		return d.JournalID != req.JournalID && d.JournalType != ""
	})

	reffs = lo.UniqBy(reffs, func(d JournalReference) string {
		return d.JournalID
	})

	sort.Slice(reffs, func(i, j int) bool {
		return lo.IndexOf(sequence, reffs[i].JournalType) < lo.IndexOf(sequence, reffs[j].JournalType)
	})

	res := &FindJournalRefResponse{
		JournalType: req.JournalType,
		JournalID:   req.JournalID,
		Refferences: reffs,
	}

	return res, nil
}

type JournalObject struct {
	GetReffByID   func(ids []string) []JournalReference
	GetReffByReff func(reffs []string) []JournalReference
}

func getJournalObjectMap(h *datahub.Hub) map[string]JournalObject {
	return map[string]JournalObject{
		string(scmmodel.ItemRequestType): {
			GetReffByID: func(ids []string) []JournalReference {
				datas := []scmmodel.ItemRequest{}
				h.GetsByFilter(new(scmmodel.ItemRequest), dbflex.In("_id", ids...), &datas)

				jfIDs := lo.Map(datas, func(d scmmodel.ItemRequest, i int) JournalReference {
					return JournalReference{
						JournalType: string(scmmodel.ItemRequestType),
						JournalID:   d.ID,
					}
				})

				jfRefs := lo.Map(datas, func(d scmmodel.ItemRequest, i int) JournalReference {
					return JournalReference{
						JournalType: "",
						JournalID:   d.WOReff,
					}
				})

				return append(jfIDs, jfRefs...)
			},
			GetReffByReff: func(reffs []string) []JournalReference {
				datas := []scmmodel.ItemRequest{}
				h.GetsByFilter(new(scmmodel.ItemRequest), dbflex.In("ReffNo", reffs...), &datas)

				return lo.Map(datas, func(d scmmodel.ItemRequest, i int) JournalReference {
					return JournalReference{
						JournalType: string(scmmodel.ItemRequestType),
						JournalID:   d.ID,
					}
				})
			},
		},

		string(scmmodel.PurchRequest): {
			GetReffByID: func(ids []string) []JournalReference {
				datas := []scmmodel.PurchaseRequestJournal{}
				h.GetsByFilter(new(scmmodel.PurchaseRequestJournal), dbflex.In("_id", ids...), &datas)

				jfIDs := lo.Map(datas, func(d scmmodel.PurchaseRequestJournal, i int) JournalReference {
					return JournalReference{
						JournalType: string(scmmodel.PurchRequest),
						JournalID:   d.ID,
					}
				})

				jfRefs := lo.FlatMap(datas, func(d scmmodel.PurchaseRequestJournal, i int) []JournalReference {
					return lo.Map(d.ReffNo, func(d string, i int) JournalReference {
						return JournalReference{
							JournalType: "",
							JournalID:   d,
						}
					})
				})

				return append(jfIDs, jfRefs...)
			},
			GetReffByReff: func(reffs []string) []JournalReference {
				datas := []scmmodel.PurchaseRequestJournal{}
				h.GetsByFilter(new(scmmodel.PurchaseRequestJournal), dbflex.In("ReffNo", reffs...), &datas)

				return lo.Map(datas, func(d scmmodel.PurchaseRequestJournal, i int) JournalReference {
					return JournalReference{
						JournalType: string(scmmodel.PurchRequest),
						JournalID:   d.ID,
					}
				})
			},
		},

		string(scmmodel.PurchOrder): {
			GetReffByID: func(ids []string) []JournalReference {
				datas := []scmmodel.PurchaseOrderJournal{}
				h.GetsByFilter(new(scmmodel.PurchaseOrderJournal), dbflex.In("_id", ids...), &datas)

				jfIDs := lo.Map(datas, func(d scmmodel.PurchaseOrderJournal, i int) JournalReference {
					return JournalReference{
						JournalType: string(scmmodel.PurchOrder),
						JournalID:   d.ID,
					}
				})

				jfRefs := lo.FlatMap(datas, func(d scmmodel.PurchaseOrderJournal, i int) []JournalReference {
					return lo.Map(d.ReffNo, func(d string, i int) JournalReference {
						return JournalReference{
							JournalType: "",
							JournalID:   d,
						}
					})
				})

				return append(jfIDs, jfRefs...)
			},
			GetReffByReff: func(reffs []string) []JournalReference {
				datas := []scmmodel.PurchaseOrderJournal{}
				h.GetsByFilter(new(scmmodel.PurchaseOrderJournal), dbflex.In("ReffNo", reffs...), &datas)

				return lo.Map(datas, func(d scmmodel.PurchaseOrderJournal, i int) JournalReference {
					return JournalReference{
						JournalType: string(scmmodel.PurchOrder),
						JournalID:   d.ID,
					}
				})
			},
		},

		string(scmmodel.JournalTransfer): {
			GetReffByID: func(ids []string) []JournalReference {
				datas := []scmmodel.InventJournal{}
				h.GetsByFilter(new(scmmodel.InventJournal), dbflex.And(dbflex.Eq("TrxType", scmmodel.JournalTransfer), dbflex.In("_id", ids...)), &datas)

				jfIDs := lo.Map(datas, func(d scmmodel.InventJournal, i int) JournalReference {
					return JournalReference{
						JournalType: string(scmmodel.JournalTransfer),
						JournalID:   d.ID,
					}
				})

				jfRefs := lo.FlatMap(datas, func(d scmmodel.InventJournal, i int) []JournalReference {
					return lo.Map(d.ReffNo, func(d string, i int) JournalReference {
						return JournalReference{
							JournalType: "",
							JournalID:   d,
						}
					})
				})

				return append(jfIDs, jfRefs...)
			},
			GetReffByReff: func(reffs []string) []JournalReference {
				datas := []scmmodel.InventJournal{}
				h.GetsByFilter(new(scmmodel.InventJournal), dbflex.And(dbflex.Eq("TrxType", scmmodel.JournalTransfer), dbflex.In("ReffNo", reffs...)), &datas)

				return lo.Map(datas, func(d scmmodel.InventJournal, i int) JournalReference {
					return JournalReference{
						JournalType: string(scmmodel.JournalTransfer),
						JournalID:   d.ID,
					}
				})
			},
		},

		string(scmmodel.InventIssuance): {
			GetReffByID: func(ids []string) []JournalReference {
				datas := []scmmodel.InventReceiveIssueJournal{}
				h.GetsByFilter(new(scmmodel.InventReceiveIssueJournal), dbflex.And(dbflex.Eq("TrxType", scmmodel.InventIssuance), dbflex.In("_id", ids...)), &datas)

				jfIDs := lo.Map(datas, func(d scmmodel.InventReceiveIssueJournal, i int) JournalReference {
					return JournalReference{
						JournalType: string(scmmodel.InventIssuance),
						JournalID:   d.ID,
					}
				})

				jfRefs := lo.FlatMap(datas, func(d scmmodel.InventReceiveIssueJournal, i int) []JournalReference {
					return lo.Map(d.ReffNo, func(d string, i int) JournalReference {
						return JournalReference{
							JournalType: "",
							JournalID:   d,
						}
					})
				})

				return append(jfIDs, jfRefs...)
			},
			GetReffByReff: func(reffs []string) []JournalReference {
				datas := []scmmodel.InventReceiveIssueJournal{}
				h.GetsByFilter(new(scmmodel.InventReceiveIssueJournal), dbflex.And(dbflex.Eq("TrxType", scmmodel.InventIssuance), dbflex.In("ReffNo", reffs...)), &datas)

				return lo.Map(datas, func(d scmmodel.InventReceiveIssueJournal, i int) JournalReference {
					return JournalReference{
						JournalType: string(scmmodel.InventIssuance),
						JournalID:   d.ID,
					}
				})
			},
		},

		string(scmmodel.InventReceive): {
			GetReffByID: func(ids []string) []JournalReference {
				datas := []scmmodel.InventReceiveIssueJournal{}
				h.GetsByFilter(new(scmmodel.InventReceiveIssueJournal), dbflex.And(dbflex.Eq("TrxType", scmmodel.InventReceive), dbflex.In("_id", ids...)), &datas)

				jfIDs := lo.Map(datas, func(d scmmodel.InventReceiveIssueJournal, i int) JournalReference {
					return JournalReference{
						JournalType: string(scmmodel.InventReceive),
						JournalID:   d.ID,
					}
				})

				jfRefs := lo.FlatMap(datas, func(d scmmodel.InventReceiveIssueJournal, i int) []JournalReference {
					return lo.Map(d.ReffNo, func(d string, i int) JournalReference {
						return JournalReference{
							JournalType: "",
							JournalID:   d,
						}
					})
				})

				return append(jfIDs, jfRefs...)
			},
			GetReffByReff: func(reffs []string) []JournalReference {
				datas := []scmmodel.InventReceiveIssueJournal{}
				h.GetsByFilter(new(scmmodel.InventReceiveIssueJournal), dbflex.And(dbflex.Eq("TrxType", scmmodel.InventReceive), dbflex.In("ReffNo", reffs...)), &datas)

				return lo.Map(datas, func(d scmmodel.InventReceiveIssueJournal, i int) JournalReference {
					return JournalReference{
						JournalType: string(scmmodel.InventReceive),
						JournalID:   d.ID,
					}
				})
			},
		},
	}
}

func getReffNo(h *datahub.Hub, journalType, journalID string) []JournalReference {
	reffs := []JournalReference{}

	// TODO: gimana cara tau JournalType sedangkan kita nyimpannya sbg []string?
	switch journalType {
	case string(scmmodel.ItemRequestType):
		orm := sebar.NewMapRecordWithORM(h, new(scmmodel.ItemRequest))
		dt, _ := orm.Get(journalID)
		reffs = append(reffs, JournalReference{JournalType: "", JournalID: dt.WOReff})

	case string(scmmodel.PurchRequest):
		orm := sebar.NewMapRecordWithORM(h, new(scmmodel.PurchaseRequestJournal))
		dt, _ := orm.Get(journalID)
		reffs = append(reffs, lo.Map(dt.ReffNo, func(d string, i int) JournalReference {
			return JournalReference{
				JournalType: "",
				JournalID:   d,
			}
		})...)

	case string(scmmodel.PurchOrder):
		orm := sebar.NewMapRecordWithORM(h, new(scmmodel.PurchaseOrderJournal))
		dt, _ := orm.Get(journalID)
		reffs = append(reffs, lo.Map(dt.ReffNo, func(d string, i int) JournalReference {
			return JournalReference{
				JournalType: "",
				JournalID:   d,
			}
		})...)

	case string(scmmodel.JournalTransfer):
		orm := sebar.NewMapRecordWithORM(h, new(scmmodel.InventJournal))
		dt, _ := orm.Get(journalID)
		reffs = append(reffs, lo.Map(dt.ReffNo, func(d string, i int) JournalReference {
			return JournalReference{
				JournalType: "",
				JournalID:   d,
			}
		})...)

	case string(scmmodel.InventIssuance), string(scmmodel.InventReceive):
		orm := sebar.NewMapRecordWithORM(h, new(scmmodel.InventReceiveIssueJournal))
		dt, _ := orm.Get(journalID)
		reffs = append(reffs, lo.Map(dt.ReffNo, func(d string, i int) JournalReference {
			return JournalReference{
				JournalType: "",
				JournalID:   d,
			}
		})...)
	}

	return reffs
}

func getData[T orm.DataModel](h *datahub.Hub, any T, id string) (T, error) {
	if e := h.GetByID(any, id); e != nil {
		return any, e
	}

	return any, nil
}
