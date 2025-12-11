package bagonglogic

import (
	"fmt"
	"log"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongconfig"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/kmsg"
	"github.com/ariefdarmawan/kmsg/ksmsg"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson"
)

type ItemBalanceJob struct {
}

func (j *ItemBalanceJob) Run(hm *kaos.HubManager) {
	hub, err := hm.Get("Demo", "tenant")
	if err != nil {
		log.Fatalf("fail to get data hub. %s", err.Error())
	}

	// get item balance
	pipe := []bson.M{
		{
			"$group": bson.M{
				"_id": bson.M{
					"ItemID":      "$ItemID",
					"SKU":         "$SKU",
					"WarehouseID": "$InventDim.WarehouseID",
				},
				"QtyAvail": bson.M{
					"$sum": "$QtyAvail",
				},
			},
		},
		{
			"$project": bson.M{
				"ItemID":      "$_id.ItemID",
				"SKU":         "$_id.SKU",
				"WarehouseID": "$_id.WarehouseID",
				"QtyAvail":    1,
			},
		},
	}

	type itemBalance struct {
		ItemID      string  `bson:"ItemID"`
		SKU         string  `bson:"SKU"`
		WarehouseID string  `bson:"WarehouseID"`
		QtyAvail    float64 `bson:"QtyAvail"`
	}

	balances := []itemBalance{}
	cmd := dbflex.From(new(scmmodel.ItemBalance).TableName()).Command("pipe", pipe)
	if _, err := hub.Populate(cmd, &balances); err != nil {
		log.Fatalf("fail to get data item balance. %s", err.Error())
	}

	// get min max
	pipe = []bson.M{
		{
			"$group": bson.M{
				"_id": bson.M{
					"ItemID":      "$ItemID",
					"SKU":         "$SKU",
					"WarehouseID": "$InventoryDimension.WarehouseID",
				},
				"MinStock": bson.M{
					"$sum": "$MinStock",
				},
				"MaxStock": bson.M{
					"$sum": "$MaxStock",
				},
			},
		},
		{
			"$project": bson.M{
				"ItemID":      "$_id.ItemID",
				"SKU":         "$_id.SKU",
				"WarehouseID": "$_id.WarehouseID",
				"MinStock":    1,
				"MaxStock":    1,
			},
		},
	}

	type itemMinMax struct {
		ItemID      string `bson:"ItemID"`
		SKU         string `bson:"SKU"`
		WarehouseID string `bson:"WarehouseID"`
		MinStock    int    `bson:"MinStock"`
		MaxStock    int    `bson:"MaxStock"`
	}

	minMax := []itemMinMax{}
	cmd = dbflex.From(new(scmmodel.ItemMinMax).TableName()).Command("pipe", pipe)
	if _, err := hub.Populate(cmd, &minMax); err != nil {
		log.Fatalf("fail to get data item balance. %s", err.Error())
	}

	mapMinMax := map[string]itemMinMax{}
	itemID := make([]string, len(minMax))
	warehouseID := make([]string, len(minMax))
	sku := make([]string, len(minMax))
	lo.ForEach(minMax, func(item itemMinMax, i int) {
		mapMinMax[fmt.Sprintf("%s-%s-%s", item.ItemID, item.SKU, item.WarehouseID)] = item
		itemID[i] = item.ItemID
		warehouseID[i] = item.WarehouseID
		sku[i] = item.SKU
	})

	type Dim struct {
		FinancialDimension tenantcoremodel.Dimension
		ItemID             string
		SKU                string
		InventoryDimension scmmodel.InventDimension
	}

	// get dimension min max
	dimMinMax := []Dim{}
	err = hub.Gets(new(scmmodel.ItemMinMax), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("SKU", sku...),
			dbflex.In("ItemID", itemID...),
			dbflex.In("InventoryDimension.WarehouseID", warehouseID...),
		),
	).SetSelect("FinancialDimension", "ItemID", "SKU", "InventoryDimension"), &dimMinMax)
	if err != nil {
		log.Fatalf("fail to get dimension item min max: %s", err.Error())
	}

	mapDimMinMax := map[string]string{}
	siteID := make([]string, len(minMax))
	lo.ForEach(dimMinMax, func(item Dim, i int) {
		id := item.FinancialDimension.Get("Site")
		mapDimMinMax[fmt.Sprintf("%s-%s-%s", item.ItemID, item.SKU, item.InventoryDimension.WarehouseID)] = id

		siteID[i] = id
	})

	type site struct {
		ID   string `bson:"_id"`
		Name string
	}

	// get dimension
	sites := []site{}
	err = hub.Gets(new(bagongmodel.Site), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("_id", siteID...),
		),
	).SetSelect("_id", "Name"), &sites)
	if err != nil {
		log.Fatalf("fail to get site: %s", err.Error())
	}
	mapSite := lo.Associate(sites, func(item site) (string, string) {
		return item.ID, item.Name
	})

	type warehouse struct {
		ID  string `bson:"_id"`
		PIC string
	}

	// get warehouse
	warehouses := []warehouse{}
	err = hub.Gets(new(tenantcoremodel.LocationWarehouse), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("_id", warehouseID...),
		),
	).SetSelect("_id", "PIC"), &warehouses)
	if err != nil {
		log.Fatalf("fail to get warehouse: %s", err.Error())
	}

	mapWarehouse := map[string]string{}
	employeeID := make([]string, len(warehouses))
	lo.ForEach(warehouses, func(w warehouse, i int) {
		mapWarehouse[w.ID] = w.PIC
		employeeID[i] = w.PIC
	})

	type employee struct {
		ID    string `bson:"_id"`
		Email string
	}

	// get employe
	employees := []employee{}
	err = hub.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("_id", employeeID...),
		),
	).SetSelect("_id", "Email"), &employees)
	if err != nil {
		log.Fatalf("fail to get employee: %s", err.Error())
	}
	mapEmployee := lo.Associate(employees, func(emp employee) (string, string) {
		return emp.ID, emp.Email
	})

	setItem := func(bal itemBalance, st SiteItem, item Item, minMax itemMinMax) SiteItem {
		item.MinStock = minMax.MinStock
		item.MaxStock = minMax.MinStock

		if int(bal.QtyAvail) < minMax.MinStock {
			st.MinItem = append(st.MinItem, item)
		} else if int(bal.QtyAvail) > minMax.MaxStock {
			st.MaxItem = append(st.MaxItem, item)
		}

		return st
	}

	mapItemDetail, err := scmlogic.AssignItem(hub, itemID, sku)
	if err != nil {
		log.Fatalf("fail to get item detail: %s", err.Error())
	}

	// check stock
	mapSiteItems := map[string]SiteItem{}
	lo.ForEach(balances, func(bal itemBalance, _ int) {
		key := fmt.Sprintf("%s-%s-%s", bal.ItemID, bal.SKU, bal.WarehouseID)
		// check site
		if siteID, okDim := mapDimMinMax[key]; okDim {
			site := mapSite[siteID]
			item := Item{
				ItemID:      bal.ItemID,
				SKU:         bal.SKU,
				WarehouseID: bal.WarehouseID,
				Qty:         bal.QtyAvail,
				PICEmail:    mapEmployee[mapWarehouse[bal.WarehouseID]],
				Detail:      mapItemDetail[bal.ItemID+bal.SKU],
			}
			if siteItem, okSite := mapSiteItems[siteID]; okSite {
				if minMax, okMinMax := mapMinMax[key]; okMinMax {
					mapSiteItems[siteID] = setItem(bal, siteItem, item, minMax)
				}
			} else {
				if minMax, okMinMax := mapMinMax[key]; okMinMax {
					st := SiteItem{
						Name: site,
					}

					mapSiteItems[siteID] = setItem(bal, st, item, minMax)
				}
			}
		}
	})

	if len(mapSiteItems) != 0 {
		// send email notification
		j.sendEmailNotification(hub, mapSiteItems)

		// create notification
		// j.createNotification(hub, minMessage, maxMessage)
	}
}

func (j *ItemBalanceJob) getHOPIC(hub *datahub.Hub) []string {
	// get warehouse
	warehouses := []tenantcoremodel.LocationWarehouse{}
	err := hub.Gets(new(tenantcoremodel.LocationWarehouse), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.ElemMatch("Dimension", dbflex.Eq("Key", "Site"), dbflex.Eq("Value", "SITE020")),
		),
	), &warehouses)
	if err != nil {
		fmt.Printf("fail to get warehouse: %s", err.Error())
	}

	ids := lo.Map(warehouses, func(w tenantcoremodel.LocationWarehouse, _ int) string {
		return w.PIC
	})

	// get employe
	employees := []tenantcoremodel.Employee{}
	err = hub.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("_id", ids...),
		),
	), &employees)
	if err != nil {
		fmt.Printf("fail to get employee: %s", err.Error())
	}

	emails := make([]string, len(employees))
	for i, e := range employees {
		emails[i] = e.Email
	}

	return emails
}

type Item struct {
	ItemID      string
	SKU         string
	WarehouseID string
	Stock       int
	Qty         float64
	MinStock    int
	MaxStock    int
	PICEmail    string
	Detail      string
}

type SiteItem struct {
	Name    string
	MaxItem []Item
	MinItem []Item
}

func (j *ItemBalanceJob) sendEmailNotification(hub *datahub.Hub, mapSiteItem map[string]SiteItem) {
	sendEmailTopic := bagongconfig.Config.TopicSendMessageTemplate
	if sendEmailTopic == "" {
		return
	}

	hoEmails := j.getHOPIC(hub)
	kind := "item-stock-notification"
	style := "border: 1px solid black;"
	for _, siteItem := range mapSiteItem {
		to := ""
		tableContent := ""
		for _, item := range siteItem.MinItem {
			to = item.PICEmail
			tableContent += fmt.Sprintf(`
				<tr>
					<td style="%s">%s</td>
					<td style="%s text-align: center;">%d</td>
					<td style="%s text-align: center;">%d</td>
					<td style="%s text-align: center;">%d</td>
				</tr>
			`, style, item.Detail, style, item.MinStock, style, item.MaxStock, style, int(item.Qty))
		}

		for _, item := range siteItem.MaxItem {
			to = item.PICEmail
			tableContent += fmt.Sprintf(`
				<tr>
					<td style="%s">%s</td>
					<td style="%s text-align: center;">%d</td>
					<td style="%s text-align: center;">%d</td>
					<td style="%s text-align: center;">%d</td>
				</tr>
			`, style, item.Detail, style, item.MinStock, style, item.MaxStock, style, int(item.Qty))
		}

		if tableContent != "" {
			msg := kmsg.Message{
				Kind:   kind,
				To:     to,
				Bcc:    hoEmails,
				Method: "SMTP",
			}

			sendMessageRequest := ksmsg.SendTemplateRequest{
				TemplateName: kind,
				Message:      &msg,
				LanguageID:   "en-us",
				Data: codekit.M{
					"Site": siteItem.Name,
					"Data": tableContent,
				},
			}

			msgID := ""
			err := bagongconfig.Config.EventHub.Publish(sendEmailTopic, sendMessageRequest, &msgID, &kaos.PublishOpts{Headers: codekit.M{}})

			if err != nil {
				fmt.Printf("fail to send email to %s. %s\n", to, err.Error())
			}
		} else {
			fmt.Println("No data in ", siteItem.Name)
		}
	}
}

func (j *ItemBalanceJob) createNotification(hub *datahub.Hub, minMessage, maxMessage string) {
	employees := []tenantcoremodel.Employee{}
	err := hub.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.In("Email", []string{}...),
	), &employees)
	if err != nil {
		fmt.Printf("fail to get employees. %s\n", err.Error())
		return
	}

	lo.ForEach(employees, func(employee tenantcoremodel.Employee, _ int) {
		notification := ficomodel.Notification{
			CompanyID:   "DEMO00",
			JournalType: "Alert",
			TrxDate:     time.Now(),
			TrxType:     "MinMax",
			UserTo:      employee.ID,
			UserToEmail: employee.Email,
			Message:     fmt.Sprintf("Minimal Stok: %s,\n\nMaksimal Stok: %s", minMessage, maxMessage),
			Created:     time.Now(),
			LastUpdate:  time.Now(),
		}

		hub.Save(&notification)
	})
}
