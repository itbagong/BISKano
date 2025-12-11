package bagonglogic

import (
	"errors"
	"sort"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
)

type TrayekEngine struct{}

type TrayekTerminal struct {
	TerminalID   string
	TerminalName string
	TerminalSort int
}

type TrayekTerminalExpense struct {
	TerminalID      string
	TerminalName    string
	ExpenseID       string
	ExpenseName     string
	ExpenseCategory string `form_items:"Per Person|Value"`
	ExpenseValue    float64
}

type TrayekTarif struct {
	From     string
	FromName string
	To       string
	ToName   string
	Rate     float64
}

type TrayekGetResponse struct {
	Detail          bagongmodel.Trayek
	Terminals       []TrayekTerminal
	Tarif           []TrayekTarif
	TerminalExpense []TrayekTerminalExpense
	FixExpense      []tenantcoremodel.ExpenseType
}

type TrayekParam struct {
	ID string `bson:"_id" json:"_id"`
}

func (engine *TrayekEngine) GetRitase(ctx *kaos.Context, req *TrayekParam) (*TrayekGetResponse, error) {
	if req.ID == "" {
		return nil, errors.New("missing: invalid request, please check your payload")
	}

	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return nil, errors.New("missing: connection")
	}

	res := new(TrayekGetResponse)

	trayek := new(bagongmodel.Trayek)
	e := hub.GetByID(trayek, req.ID)

	if e == nil {

		//get list of terminal by list of ids
		terminalIds := []interface{}{}
		for _, c := range trayek.Terminals {
			terminalIds = append(terminalIds, c)
		}
		terminals := []bagongmodel.Terminal{}
		e := hub.GetsByFilter(new(bagongmodel.Terminal), dbflex.In("_id", terminalIds...), &terminals)
		if e != nil {
			ctx.Log().Errorf("Failed populate data master terminals: %s", e.Error())
		}

		// get list tarif
		tarifs := []TrayekTarif{}
		for _, c := range trayek.Tarifs {
			tarif := TrayekTarif{}
			tarif.From = c.From
			tarif.FromName = getTerminalName(c.From, terminals)
			tarif.To = c.To
			tarif.ToName = getTerminalName(c.To, terminals)
			tarif.Rate, e = trayek.GetTariff(c.From, c.To)
			if e != nil {
				return nil, errors.New("missing: cant get tarif " + c.From + " - " + c.To)
			}
			tarifs = append(tarifs, tarif)
		}

		// get list terminal expense
		terminalExpense := []TrayekTerminalExpense{}
		for _, c := range terminals {
			for _, d := range c.Expenses {
				expense := TrayekTerminalExpense{
					TerminalID:      c.ID,
					TerminalName:    c.Name,
					ExpenseID:       d.ID,
					ExpenseName:     d.Name,
					ExpenseCategory: d.ExpenseCategory,
					ExpenseValue:    d.Value,
				}
				terminalExpense = append(terminalExpense, expense)
			}
		}

		// list terminal
		listTerminal := []TrayekTerminal{}
		for _, c := range terminals {
			t := TrayekTerminal{
				TerminalID:   c.ID,
				TerminalName: c.Name,
				TerminalSort: getTerminalIdx(c.ID, trayek.Terminals),
			}
			listTerminal = append(listTerminal, t)
		}

		sort.Slice(listTerminal, func(i, j int) bool {
			return listTerminal[i].TerminalSort < listTerminal[j].TerminalSort
		})

		res.Detail = *trayek
		res.Terminals = listTerminal
		res.Tarif = tarifs
		res.TerminalExpense = terminalExpense
		res.FixExpense = trayek.Expense
	}

	return res, nil
}

func getTerminalIdx(id string, terminals []string) int {
	for i, c := range terminals {
		if id == c {
			return i
		}
	}
	return 0
}

func getTerminalName(id string, terminals []bagongmodel.Terminal) string {
	for _, c := range terminals {
		if c.ID == id {
			return c.Name
		}
	}
	return ""
}
