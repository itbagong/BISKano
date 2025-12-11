const appMenu = [
  { label: "Dashboard", icon: "view-dashboard", url: "/" },
  {
    label: "Accounting",
    icon: "cash",
    submenu: [
      [{ label: "Ledger Journal", url: "/fico/ledgerjournal" }],
      [{ label: "Journal Type", url: "/fico/ledgerjournaltype" }],
      [
        // { label: 'General Journal', url: '/ledger/journal' },
        {
          label: "Payment Term",
          url: "/share/TableView?objname=/fico/paymentterm&title=Payment%20Term",
        },
        {
          label: "Chart of Accounts",
          url: "/tenant/LedgerAccount",
        },
      ],
    ],
  },
  {
    label: "Customer",
    icon: "human",
    submenu: [
      [{ label: "Customer Journal", url: "/fico/CustomerTransaction" }],
      [{ label: "Journal Type", url: "/fico/CustomerJournalType" }],
      [
        { label: "Customer", url: "/bagong/Customer" },
        {
          label: "Customer Group",
          url: "/share/TableView?objname=/tenant/customergroup&title=Customer%20Group",
        },
      ],
    ],
  },
  {
    label: "Vendor",
    icon: "basket-outline",
    submenu: [
      [
        { label: "Vendor Journal", url: "/fico/vendortransaction" },
        { label: "General Approval", url: "/admin/msgtpl" },
      ],
      [{ label: "Journal Type", url: "/fico/VendorJournalType" }],
      [
        // { label: 'Aging AP', url: '/report/ap-aging' }
        // { label: 'Vendor', url: '/share/TableView?objname=/tenant/vendor&title=Vendor' },
        { label: "Vendor", url: "/bagong/vendor" },
        {
          label: "Vendor Group",
          url: "/share/TableView?objname=/tenant/vendorgroup&title=Vendor%20Group",
        },
      ],
    ],
  },
  {
    label: "Cash & Bank",
    icon: "bank",
    submenu: [
      [
        { label: "Cash In", url: "/bagong/cashin" },
        { label: "Cash Out", url: "/bagong/cashout" },
        { label: "Apply", url: "/bagong/apply" },
      ],
      [
        {
          label: "Cash & Bank",
          url: "/fico/CashBank",
        },
        {
          label: "Cash & Bank Group",
          url: "/share/TableView?objname=/tenant/cashbankgroup&title=Cash%20Bank%20Group",
        },
        {
          label: "Cheque & Giro Book",
          url: "/fico/ChequeGiroBook",
        },
        {
          label: "Cheque & Giro",
          url: "/fico/ChequeGiro",
        },
      ],
      [
        {
          label: "Journal Type",
          url: "/fico/CashJournalType",
        },
      ],
      [
        {
          label: "Bank Reconciliation",
          url: "/fico/BankRecon",
        },
      ],
    ],
  },
  {
    label: "Fixed Asset",
    icon: "file",
    submenu: [
      [{ label: "Asset Journal", url: "" }],
      [
        { label: "Asset", url: "/bagong/asset" },
        {
          label: "Asset Group",
          url: "/share/TableView?objname=/tenant/assetgroup&title=Asset Group",
        },
        { label: "Fixed Asset Number", url: "/fico/fixedassetnumber" },
        { label: "Asset Movement", url: "/bagong/assetmovement" },
      ],
      [{ label: "Journal Type", url: "/fico/assetjournaltype" }],
    ],
  },
  {
    label: "Setting",
    icon: "remote",
    submenu: [
      [
        {
          label: "Company",
          url: "/share/TableView?objname=/tenant/company&title=Company",
        },
        {
          label: "Number Sequence",
          url: "/share/TableView?objname=/tenant/numseq&title=Number%20Sequence",
        },
        {
          label: "Number Sequence Setup",
          url: "/share/TableView?objname=/tenant/numseqsetup&title=Number%20Sequence%20Setup",
        },
        // { label: 'Currency', url: '/share/TableView?objname=/tenant/currency&title=Currency' },
        {
          label: "Charge Code",
          url: "/share/TableView?objname=/fico/chargesetup&title=Charge%20Code",
        },
        { label: "Checklist Template", url: "/tenant/checklisttemplate" },
        {
          label: "Dimension",
          url: "/share/TableView?objname=/tenant/dimension&title=Dimension",
        },
        {
          label: "Expense Type",
          url: "/share/TableView?objname=/tenant/expensetype&title=Master%20Expense%20Type",
        },
        {
          label: "Expense Type Group",
          url: "/share/TableView?objname=/tenant/expensetypegroup&title=Master%20Expense%20Type%20Group",
        },
        { label: "Reference Template", url: "/tenant/referencetemplate" },
        { label: "Posting Profile", url: "/fico/PostingProfile" },
        {
          label: "Tax Code",
          url: "/share/TableView?objname=/fico/taxsetup&title=Tax%20Code",
        },
        {
          label: "Tax Group",
          url: "/share/TableView?objname=/fico/taxsetup&title=Tax%20Group",
        },
        { label: "Fiscal Year", url: "/fico/fiscalyear" },
        { label: "Site", url: "/bagong/Site" },
        {
          label: "Master Data Type",
          url: "/tenant/masterdatatype",
        },
        {
          label: "Master Data",
          submenu: [],
        },
        {
          label: "Master Item Template",
          url: "/tenant/ItemTemplate",
        },
      ],
    ],
  },
  {
    label: "HR",
    icon: "table",
    submenu: [
      [
        // { label: "Payroll", url: "/bagong/payroll" },
        // { label: "Payroll Submission", url: "/bagong/payrollsubmission" },
        {
          label: "Employee Expense Submission",
          url: "/bagong/SubmissionEmployeeExpense",
        },
        { label: "Petty Cash Submission", url: "/bagong/SubmissionPettyCash" },
        { label: "General Submission", url: "/bagong/SubmissionVendor" },
        // { label: "Input Payroll Item", url: "" },
      ],
      [{ label: "Approval All Expense", url: "/bagong/ApprovalAllExpense" }],
      [
        { label: "Attendance", url: "/kara/history" },
        { label: "Leave", url: "/kara/leave" },
      ],
      [
        { label: "Attendance Rule", url: "/kara/rule" },
        { label: "Work Location", url: "/kara/location" },
        {
          label: "Work Holiday",
          url: "/share/TableView?objname=/kara/holiday&title=Holiday",
        },
        {
          label: "Work Holiday Item",
          url: "/share/TableView?objname=/kara/holidayitem&title=Holiday Item",
        },
        { label: "Leave Type", url: "/kara/LeaveType" },
        { label: "Profile", url: "/kara/profile" },
      ],
      [
        {
          label: "Master Benefit",
          url: "/fico/BenefitDeduction?objname=/fico/payrollbenefit&title=Benefit",
        },
        {
          label: "Master Deduction",
          url: "/fico/BenefitDeduction?objname=/fico/payrolldeduction&title=Deduction",
        },
        {
          label: "Master Loan",
          url: "/hcm/LoanSetup",
        },
        { label: "Employee", url: "/bagong/employee" },
        {
          label: "Employee Group",
          url: "/share/TableView?objname=/tenant/employeegroup&title=Employee%20Group",
        },
      ],
    ],
  },
  {
    label: "Site",
    icon: "road",
    submenu: [
      [
        // { label: "Asset Booking", url: "/bagong/AssetBooking" },
        { label: "Site Entry", url: "/bagong/SiteEntry" },
        { label: "Claim", url: "/bagong/Claim" },
        { label: "SDK", url: "/bagong/sdk" },
      ],
      [
        { label: "Terminal", url: "/bagong/Terminal" },
        { label: "Trayek", url: "/bagong/Trayek" },
      ],
      // [
      //   {
      //     label: "Rental Contract",
      //     url: "/share/TableView?objname=/bagong/rentalcontract&title=Rental%20Contract",
      //   },
      // ],
      [
        // {
        //   label: "Journal Type",
        //   url: "/fico/SiteJournalType",
        // },
        {
          label: "Mapping Journal Type",
          url: "/bagong/MappingJournalType",
        },
      ],
    ],
  },
  {
    label: "SDPM",
    icon: "link-box-outline",
    submenu: [
      [
        {
          label: "Sales Journal Type",
          url: "/sdp/salesorderjournaltype",
        },
      ],
      [
        {
          label: "Contact",
          url: "/share/TableView?objname=/tenant/contact&title=Contact",
        },
        {
          label: "Sales Opportunity",
          url: "/sdp/SalesOpportunity",
        },
        {
          label: "Sales Quotation",
          url: "/sdp/salesquotation",
        },
        {
          label: "Sales Order",
          url: "/sdp/salesorder",
        },
        {
          label: "Sales Measuring Project",
          url: "/sdp/MeasuringProject",
        },
        {
          label: "Unit Calendar",
          url: "/sdp/unitcalendar",
        },
        {
          label: "Sales Price Book",
          url: "/sdp/salespricebook",
        },
        {
          label: "Document Unit Checklist",
          url: "/sdp/documentunitchecklist",
        },
        {
          label: "Contract Checklist",
          url: "/sdp/contractchecklist",
        },
      ],
    ],
  },
  {
    label: "SCM",
    icon: "shield-check",
    submenu: [
      [
        {
          label: "Master",
          submenu: [
            [
              {
                label: "Item group",
                url: "/tenant/itemgroup",
              },
              {
                label: "Item",
                url: "/tenant/item",
              },
              {
                label: "Item serial",
                url: "/share/TableView?objname=/tenant/itemserial&title=Item%20Serial",
              },
              {
                label: "Item batch",
                url: "/share/TableView?objname=/tenant/itembatch&title=Item%20Batch",
              },
              {
                label: "Item Balance",
                url: "/scm/ItemBalance",
              },
              {
                label: "Item Min Max",
                url: "/scm/ItemMinMax",
              },
              {
                label: "Unit",
                url: "/bagong/Unit",
              },
            ],
            [
              {
                label: "Warehouse Group",
                url: "/share/TableView?objname=/tenant/warehouse/group&title=Warehouse%20Group",
              },
              {
                label: "Warehouse",
                url: "/scm/Warehouse",
                // url: "/share/TableView?objname=/tenant/warehouse&title=Warehouse",
              },
              {
                label: "Section",
                url: "/share/TableView?objname=/tenant/section&title=Section",
              },
              {
                label: "Aisle",
                url: "/share/TableView?objname=/tenant/aisle&title=Aisle",
              },
              {
                label: "Box",
                url: "/share/TableView?objname=/tenant/box&title=Box",
              },
            ],
            [
              {
                label: "Spec Variant",
                url: "/share/TableView?objname=/tenant/specvariant&title=Spec%20variant",
              },
              {
                label: "Spec Size",
                url: "/share/TableView?objname=/tenant/specsize&title=Spec%20size",
              },
              {
                label: "Spec Grade",
                url: "/share/TableView?objname=/tenant/specgrade&title=Spec%20grade",
              },
            ],
            [
              {
                label: "Inventory Journal type",
                url: "/scm/InventoryJournalType",
              },
              {
                label: "Purchase Request Journal Type",
                url: "/scm/PurchaseRequestJournalType",
              },
              {
                label: "Purchase Order Journal Type",
                url: "/scm/PurchaseOrderJournalType",
              },
              {
                label: "Item Request Journal Type",
                url: "/scm/ItemRequestJournalType",
              },
              {
                label: "Asset Acquisition Journal Type",
                url: "/scm/AssetAcquisitionJournalType",
              },
              {
                label: "Vendor pricelist",
                url: "/scm/VendorPricelist",
              },
            ],
            [
              {
                label: "MCU Condition",
                url: "/bagong/mcuCondition",
              },
            ],
          ],
        },
      ],
      [
        {
          label: "Inventory",
          submenu: [
            [
              {
                label: "Movement in",
                url: "/scm/InventoryJournal?type=Movement%20In&title=Movement%20In",
              },
              {
                label: "Movement out",
                url: "/scm/InventoryJournal?type=Movement%20Out&title=Movement%20Out",
              },
              {
                label: "Item Transfer",
                url: "/scm/InventoryJournal?type=Transfer&title=Item Transfer",
              },
              {
                label: "Stock Opname",
                url: "/scm/StockOpname",
              },
              {
                label: "Inventory Adjustment",
                url: "/scm/InventoryAdjustment",
              },
              // {
              //   label: "Inventory Journal",
              //   url: "/scm/InventoryTransaction",
              // },
              // {
              //   label: "Inventory Receive Journal",
              //   url: "/scm/InventoryReceiveJournal",
              // },
            ],
          ],
        },
        {
          label: "eProcurement",
          submenu: [
            [
              {
                label: "Purchase Request",
                url: "/scm/PurchaseRequest",
              },
              {
                label: "Purchase Order",
                url: "/scm/PurchaseOrder",
              },
            ],
          ],
        },
        {
          label: "Good Receive",
          url: "/scm/InventTrx?type=Inventory%20Receive&title=Inventory%20Receive",
        },
        {
          label: "Goods Issuance",
          url: "/scm/InventTrx?type=Inventory%20Issuance&title=Inventory%20Issuance",
        },
        {
          label: "Item Request",
          url: "/scm/ItemRequest",
        },
        {
          label: "Asset Acquisition",
          url: "/scm/AssetAcquisition",
        },
      ],
    ],
  },
  {
    label: "MFG",
    icon: "calendar",
    submenu: [
      [
        {
          label: "Master",
          submenu: [
            [
              {
                label: "Routine Template",
                url: "/mfg/RoutineTemplate",
              },
              { label: "BOM", url: "/mfg/billofmaterial" },
              {
                label: "Work Order Journal Type",
                url: "/mfg/WorkOrderJournalType",
              },
              {
                label: "Work Request Journal Type",
                url: "/mfg/WorkRequestorJournalType",
              },
            ],
          ],
        },
        {
          label: "Routine",
          url: "/mfg/routine",
        },
        {
          label: "Work Request",
          url: "/mfg/WorkRequestor",
        },
        // {
        //   label: "Work Order",
        //   url: "/mfg/WorkOrder",
        // },
        {
          label: "Work Order",
          url: "/mfg/WorkOrderPlan",
        },
        {
          label: "Physical Availability",
          url: "/mfg/PhysicalAvailability",
        },
      ],
    ],
  },
  {
    label: "flow",
    icon: "sitemap",
    submenu: [
      [
        {
          label: "Flow",
          url: "/bagong/flow",
        },
      ],
    ],
  },
  {
    label: "SHE",
    icon: "viewList",
    submenu: [
      [
        {
          label: "SHE Program",
          submenu: [
            [
              { label: "Safetycard", url: "/she/Safetycard" },
              { label: "Coaching", url: "/she/Coaching" },
              { label: "Meeting", url: "/she/Meeting" },
              { label: "Sidak", url: "/she/Sidak" },
            ],
          ],
        },
        {
          label: "PICA",
          url: "/she/Pica",
        },
        {
          label: "IBPR",
          url: "/bagong/Ibpr",
        },
        { label: "Buletin", url: "/she/Buletin" },
        { label: "Legal Register", url: "/she/LegalRegister" },
        { label: "Legal Compliance", url: "/she/LegalCompliance" },
        { label: "MCU Item Template", url: "/she/McuItemTemplate" },
        { label: "MCU Master Paket", url: "/she/McuPaket" },
        { label: "MCU Transaction", url: "/she/McuTransaction" },
        { label: "SOP", url: "/she/Sop" },
        { label: "Master IBPR", url: "/she/MasterIbpr?id=IBPR" },
        { label: "Master RSCA", url: "/she/MasterIbpr?id=RSCA" },
        { label: "Inspection", url: "/she/Inspection" },
        { label: "CSMS", url: "/she/Csms" },
      ],
    ],
  },
  //{label:'Home', icon:'home', url:import.meta.env.VITE_HOME_URL},
];

export default appMenu;
