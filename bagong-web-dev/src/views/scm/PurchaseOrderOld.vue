<template>
  <div class="w-full">
    <data-list
      v-show="data.appMode == 'grid'"
      id="po-data-list"
      class="card"
      ref="listControl"
      :title="data.titleForm"
      :form-hide-submit="true"
      grid-config="/scm/purchase/order/gridconfig"
      form-config="/scm/purchase/order/formconfig"
      :grid-read="`/scm/purchase/order/gets`"
      form-read="/scm/purchase/order/get"
      grid-mode="grid"
      grid-delete="/scm/purchase/order/delete"
      grid-sort-field="Created"
      grid-sort-direction="desc"
      form-keep-label
      :form-insert="data.formInsert"
      :form-update="data.formUpdate"
      :grid-fields="['Status', 'WarehouseID']"
      :form-fields="[
        '_id',
        'JournalTypeID',
        'PostingProfileID',
        'TaxType',
        'PurchaseType',
        'PIC',
        'Dimension',
        'Location',
      ]"
      :init-app-mode="data.appMode"
      :init-form-mode="data.formMode"
      :form-tabs-new="data.tabs"
      :form-tabs-edit="data.tabs"
      :form-tabs-view="data.tabs"
      :grid-hide-new="!profile.canCreate"
      :grid-hide-edit="!profile.canUpdate"
      :grid-hide-delete="!profile.canDelete"
      @formNewData="newRecord"
      @formEditData="editRecord"
      @preSave="onPreSave"
      @postSave="onPostSave"
      @controlModeChanged="onControlModeChanged"
      @form-field-change="onFormFieldChange"
      @alterFormConfig="alterFormConfig"
      @alterGridConfig="alterGridConfig"
      :stay-on-form-after-save="data.stayOnForm"
    >
      <template #grid_Status="{ item }">
        <status-text :txt="item.Status" />
      </template>
      <template #grid_WarehouseID="{ item }">
        {{ item.Location.WarehouseID }}
      </template>
      <template #form_loader>
        <loader />
      </template>
      <template #form_buttons_1="{ item }">
        <s-button
          class="bg-transparent hover:bg-blue-500 hover:text-black"
          label="Preview"
          icon="eye-outline"
          @click="onPreview"
        ></s-button>
        <s-button
          v-if="
            !['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(item.Status)
          "
          :icon="`content-save`"
          class="btn_primary submit_btn"
          label="Save"
          :disabled="data.disableButton"
          @click="onSave(item)"
        />
        <FormButtonsTrx
          :posting-profile-id="item.PostingProfileID"
          :disabled="data.disableButton"
          :status="item.Status"
          :journalId="item._id"
          journal-type-id="Purchase Order"
          moduleid="scm/new"
          @pre-submit="preSubmit"
          @post-submit="postSubmit(item)"
          @error-submit="errorSubmit"
        >
        </FormButtonsTrx>
      </template>
      <template #form_tab_Line="{ item }">
        <PurchaseLine
          ref="lineConfig"
          v-model="item.Lines"
          :general-record="item"
          purchase-type="purchase/order"
          :disable-field="data.disableField"
          :islog="true"
          :is-generate="true"
          @attachment-action="
            (attachment) => {
              item.Attachment = attachment;
            }
          "
        ></PurchaseLine>
      </template>
      <template #form_tab_Attachment="{ item }">
        <s-grid-attachment
          :journal-id="item._id"
          journal-type="Purchase Order"
          ref="gridAttachment"
          v-model="item.Attachment"
        ></s-grid-attachment>
      </template>
      <template #form_input_JournalTypeID="{ item }">
        <s-input
          :key="data.keyJournalType"
          ref="refInput"
          label="Journal Type"
          v-model="item.JournalTypeID"
          class="w-full"
          :required="true"
          :disabled="true"
          :keepErrorSection="true"
          use-list
          :lookup-url="`/scm/purchase/order/journal/type/find`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          @change="
            (field, v1, v2, old, ctlRef) => {
              getJournalType(v1, 'change');
            }
          "
        ></s-input>
      </template>
      <template #form_input_TaxType="{ item }">
        <div class="flex gap-[10px]">
          <s-input
            ref="refInput"
            label="Tax Registration"
            v-model="item.TaxRegistration"
            class="w-full"
            keepLabel
            :hide-label="false"
            :disabled="true"
          ></s-input>
          <s-input
            ref="refInput"
            label="Tax Type"
            v-model="item.TaxType"
            class="w-full"
            use-list
            :disabled="
              ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(
                item.Status
              ) || item.VendorID
            "
            :lookup-url="`/tenant/masterdata/find?MasterDataTypeID=TTY`"
            lookup-key="_id"
            :lookup-labels="['Name']"
            :lookup-searchs="['_id', 'Name']"
          ></s-input>
        </div>
      </template>
      <template #form_input_PurchaseType="{ item }">
        <s-input
          ref="refInput"
          label="Purchase Type"
          v-model="item.PurchaseType"
          class="w-full"
          :disabled="
            ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(item.Status)
          "
          use-list
          :items="
            item.Dimension &&
            item.Dimension.find((_dim) => _dim.Key === 'Site') &&
            item.Dimension.find((_dim) => _dim.Key === 'Site')['Value'] ==
              'SITE020'
              ? ['STOCK', 'VIRTUAL', 'SERVICE', 'ASSET']
              : ['STOCK', 'VIRTUAL', 'SERVICE']
          "
        ></s-input>
      </template>
      <template #form_input_PIC="{ item, config }">
        <s-input
          :key="data.keyPIC"
          ref="refInput"
          label="PIC"
          v-model="item.PIC"
          class="w-full"
          :required="true"
          :disabled="
            ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(item.Status)
          "
          use-list
          :lookup-url="`/tenant/employee/find`"
          lookup-key="_id"
          :lookup-labels="['_id', 'Name']"
          :lookup-searchs="['_id', 'Name']"
          :lookup-payload-builder="
            (search) =>
              lookupPayloadBuilder(search, ['_id', 'Name'], item.PIC, item)
          "
        ></s-input>
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor
          ref="FinancialDimension"
          :key="data.keyDimension"
          v-model="item.Dimension"
          sectionTitle="Financial Dimension"
          :default-list="profile.Dimension"
          :readOnly="
            data.lockDimension ||
            ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(item.Status)
          "
          @change="
            (field, v1, v2) => {
              onChangeDimension(field, v1, v2, item);
            }
          "
        ></dimension-editor>
      </template>
      <template #form_input_Location="{ item }">
        <dimension-invent-jurnal
          ref="InventDimControl"
          :mandatory="['WarehouseID']"
          :key="data.keyDimension"
          v-model="item.Location"
          :default-list="profile.Dimension"
          title-header="Inventory Dimension"
          :site="
            item.Dimension &&
            item.Dimension.find((_dim) => _dim.Key === 'Site') &&
            item.Dimension.find((_dim) => _dim.Key === 'Site')['Value'] != ''
              ? item.Dimension.find((_dim) => _dim.Key === 'Site')['Value']
              : undefined
          "
          :hide-field="[
            'BatchID',
            'SerialNumber',
            'SpecID',
            'InventDimID',
            'VariantID',
            'Size',
            'Grade',
          ]"
          :readOnly="
            data.lockInventDimension ||
            ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(item.Status)
          "
          @onChange="
            (field, v1, v2, old, val) => {
              onFieldChanged(val, item);
            }
          "
        ></dimension-invent-jurnal>
      </template>
      <template #form_footer_1="{ item }">
        <RejectionMessageList
          v-if="item.Status === 'REJECTED'"
          ref="listRejectionMessage"
          journalType="Purchase Order"
          :journalID="item._id"
        ></RejectionMessageList>
      </template>
      <template #grid_item_buttons_1="{ item }">
        <log-trx :id="item._id" :hide-button="item.Status == 'DRAFT'" />
      </template>
    </data-list>

    <PreviewReport
      v-if="data.appMode == 'preview'"
      class="card w-full"
      title="Preview"
      :preview="data.preview"
      @close="closePreview"
      SourceType="Purchase Order"
      :SourceJournalID="data.record._id"
    ></PreviewReport>
  </div>
</template>

<script setup>
import {
  reactive,
  ref,
  inject,
  computed,
  watch,
  onMounted,
  nextTick,
} from "vue";
import { layoutStore } from "@/stores/layout.js";
import { DataList, util, SInput, SButton } from "suimjs";
import { useRoute } from "vue-router";
import helper from "@/scripts/helper.js";
import Loader from "@/components/common/Loader.vue";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionInventJurnal from "@/components/common/DimensionInventJurnal.vue";
import PurchaseLine from "./widget/PurchaseLine.vue";
import FormButtonsTrx from "@/components/common/FormButtonsTrx.vue";
import RejectionMessageList from "./widget/RejectionMessageList.vue";
import SingleAttachment from "./widget/SingleAttachment.vue";
import LogTrx from "@/components/common/LogTrx.vue";
import moment from "moment";
import lodash from "lodash";
import StatusText from "@/components/common/StatusText.vue";
import SGridAttachment from "@/components/common/SGridAttachment.vue";
import { authStore } from "@/stores/auth";
import PreviewReport from "@/components/common/PreviewReport.vue";

layoutStore().name = "tenant";
const featureID = "PurchaseOrder";
const profile = authStore().getRBAC(featureID);
const auth = authStore();
const refInput = ref(null);
const listControl = ref(null);
const lineConfig = ref(null);
const FinancialDimension = ref(null);
const InventDimControl = ref(null);
const listRejectionMessage = ref(null);
const route = useRoute();
const headOffice = layoutStore().headOfficeID;
const axios = inject("axios");
const gridAttachment = ref(SGridAttachment);
const data = reactive({
  appMode: "grid",
  formMode: "edit",
  stayOnForm: false,
  lockDimension: false,
  lockPostingProfile: false,
  lockInventDimension: false,
  compDimensionKey: "0",
  compInventDimensionKey: "0",
  keyJournalType: util.uuid(),
  keyDimension: util.uuid(),
  keyPIC: util.uuid(),
  titleForm: "Purchase Order",
  journalTypeData: {},
  formInsert: "/scm/purchase/order/save",
  formUpdate: "/scm/purchase/order/save",
  tabs: ["General"],
  disableField: [],
  record: {
    _id: "",
    TrxDate: new Date(),
    DocumentDate: new Date(),
    PRDate: new Date(),
    PODate: new Date(),
    DueDate: new Date(),
    DeliveryDate: new Date(),
    Status: "",
    TrxType: "",
    WarehouseID: "",
    DeliveryName: "",
    DeliveryAddress: "",
    PIC: "",
    BillingName: "",
    BillingAddress: "",
    Freight: 0,
    Lines: [],
    AttachmentID: "",
  },
  fileAttach: {},
  isPostFlow: false,
  disableButton: false,
  siteUser: "",
});

function newRecord(record) {
  data.stayOnForm = true;
  data.formMode = "new";
  record._id = "";
  record.CompanyID = auth.companyId;
  record.TrxDate = new Date();
  record.DocumentDate = new Date();
  record.PRDate = new Date();
  record.PODate = new Date();
  record.DueDate = new Date();
  record.DeliveryDate = new Date();
  record.Status = "";
  record.TrxType = data.transactionType;
  record.Freight = 0;
  record.Location = {
    WarehouseID: "",
  };
  data.titleForm = `Create New Purchase Order`;
  data.record = record;
  data.lockDimension = false;
  data.lockPostingProfile = false;
  data.lockInventDimension = false;
  getPostingProfile(record);
  getDetailEmployee("", record);
  openForm(record);
}

function editRecord(record) {
  data.stayOnForm = true;
  data.formMode = "edit";
  data.record = record;
  data.titleForm = `Edit Purchase Order | ${record._id}`;
  if (record.JournalTypeID) {
    getJournalType(record.JournalTypeID, "init");
  }
  record.Lines.map(function (l) {
    l.ItemVarian = helper.ItemVarian(l.ItemID, l.SKU);
    return l;
  });
  openForm(record);
  nextTick(() => {
    if (["SUBMITTED", "READY", "REJECTED", "POSTED"].includes(record.Status)) {
      if (record.Status === "REJECTED") {
        listRejectionMessage.value.fetchRecord(record._id);
      }
      setFormMode("view");
    }
    if (record.ReffNo?.length > 0) {
      getPostingProfile(record);
    }
  });
}
function openForm(record) {
  util.nextTickN(2, () => {
    data.fileAttach = {};
    data.isPostFlow = false;
    listControl.value.setFormFieldAttr(
      "_id",
      "hide",
      data.formMode == "new" ? true : false
    );
    const el = document.querySelector(
      "#po-data-list .form_inputs > div.flex.section_group_container > div:nth-child(1) > div:nth-child(1) > div.flex.flex-col.gap-4 > div:nth-child(1)"
    );
    if (record._id == "") {
      el ? (el.style.display = "none") : "";
      data.tabs = ["General"];
    } else {
      el ? (el.style.display = "block") : "";
      data.tabs = ["General", "Line", "Attachment"];
      util.nextTickN(2, () => {
        let tabs = document.querySelector(".tab_container > div");
        tabs.addEventListener("click", function (event) {
          setMinDate(record);
        });
        if (record.VendorID) {
          listControl.value.setFormFieldAttr("TaxName", "readOnly", true);
          listControl.value.setFormFieldAttr("PaymentTerms", "readOnly", true);
          listControl.value.setFormFieldAttr("TaxType", "readOnly", true);
          listControl.value.setFormFieldAttr("TaxCodes", "readOnly", true);
          listControl.value.setFormFieldAttr("TaxAddress", "readOnly", true);
        } else {
          listControl.value.setFormFieldAttr("TaxName", "readOnly", false);
          listControl.value.setFormFieldAttr("PaymentTerms", "readOnly", false);
          listControl.value.setFormFieldAttr("TaxType", "readOnly", false);
          listControl.value.setFormFieldAttr("TaxCodes", "readOnly", false);
          listControl.value.setFormFieldAttr("TaxAddress", "readOnly", false);
        }
      });
    }
    setMinDate(record);
    listControl.value.setFormLoading(false);
  });
}
function setMinDate(record) {
  const el = document.querySelector('input[placeholder="PO date"]');
  if (el) {
    el.setAttribute("min", moment(record.PRDate).format("YYYY-MM-DD"));
  }
  const ell = document.querySelector('input[placeholder="Due date"]');
  if (ell) {
    ell.setAttribute("min", moment(record.PODate).format("YYYY-MM-DD"));
  }
  const elll = document.querySelector('input[placeholder="Delivery date"]');
  if (elll) {
    elll.setAttribute("min", moment(record.DueDate).format("YYYY-MM-DD"));
  }
}
function onSave(record) {
  const dim = record.Location;
  const pc = record.Dimension.find((d) => {
    return d.Key == "PC";
  }).Value;
  const cc = record.Dimension.find((d) => {
    return d.Key == "CC";
  }).Value;
  const site = record.Dimension.find((d) => {
    return d.Key == "Site";
  }).Value;

  let validate = true;
  let validDimension = true;
  if (InventDimControl.value) {
    validate = InventDimControl.value.validate();
  }
  if (FinancialDimension.value) {
    validDimension = FinancialDimension.value.validate();
  } else {
    if (!pc || !cc || !site) {
      validDimension = false;
    }
  }

  if (!record.Name) {
    return util.showError("General Name is required");
  }

  if (!record.Location?.WarehouseID) {
    return util.showError("General Warehouse ID is required");
  }

  for (let l = 0; l < record?.Lines?.length; l++) {
    if (!record.Lines[l].ItemID) {
      return util.showError("there is a line field Item required");
    }
    if (typeof record.Lines[l].Qty != "number") {
      return util.showError("there is a line field Qty required");
    }
    if (record.Lines[l].UnitCost == 0 && data.siteUser != headOffice) {
      return util.showError("UnitCost item which is 0");
    }
    if (
      record.Lines[l].PRID &&
      record.Lines[l].Qty > record.Lines[l].RemainingQty
    ) {
      return util.showError("qty more than Remaining qty");
    }
  }

  if (record?.Lines?.length > 0) {
    record.Lines.map(function (v) {
      if (dim.WarehouseID) {
        v.InventDim.WarehouseID = dim.WarehouseID;
      }
      if (dim.AisleID) {
        v.InventDim.AisleID = dim.AisleID;
      }
      if (dim.SectionID) {
        v.InventDim.SectionID = dim.SectionID;
      }
      if (dim.BoxID) {
        v.InventDim.BoxID = dim.BoxID;
      }
    });
  }
  const payload = JSON.parse(JSON.stringify(record));
  if (record.Status == "") {
    payload.Status = "DRAFT";
  }
  if (validDimension && validate && listControl.value.formValidate()) {
    listControl.value.setFormLoading(true);
    data.disableButton = true;
    listControl.value.submitForm(
      payload,
      () => {
        data.disableButton = false;
        listControl.value.setFormLoading(false);
      },
      () => {
        data.disableButton = false;
        listControl.value.setFormLoading(false);
      }
    );
  } else {
    return util.showError("field is required");
  }
}
function onFormFieldChange(name, v1, v2, old, record) {
  switch (name) {
    case "VendorID":
      // record.DeliveryName = v2;
      // record.BillingName = v2;
      listControl.value.setFormFieldAttr("TaxName", "readOnly", false);
      listControl.value.setFormFieldAttr("PaymentTerms", "readOnly", false);
      listControl.value.setFormFieldAttr("TaxType", "readOnly", false);
      listControl.value.setFormFieldAttr("TaxAddress", "readOnly", false);
      record.VendorName = "";
      record.PaymentTerms = "";
      record.TaxType = "";
      record.TaxName = "";
      record.TaxRegistration = "";
      record.TaxAddress = "";
      if (typeof v1 == "string") {
        axios.post("/tenant/vendor/get", [v1]).then((r) => {
          record.VendorName = r.data.Name;
          record.PaymentTerms = r.data.PaymentTermID;
          record.TaxType = r.data.TaxType;
          record.TaxName = r.data.TaxName;
          record.TaxRegistration = r.data.TaxRegistrationNumber;
          record.TaxAddress = r.data.TaxAddress;
          listControl.value.setFormFieldAttr("TaxName", "readOnly", true);
          listControl.value.setFormFieldAttr("PaymentTerms", "readOnly", true);
          listControl.value.setFormFieldAttr("TaxType", "readOnly", true);
          listControl.value.setFormFieldAttr("TaxAddress", "readOnly", true);
        });
      }
      break;
    case "WarehouseID":
      if (typeof v1 == "string") {
        axios.post("/tenant/warehouse/get", [v1]).then((r) => {
          record.DeliveryAddress = r.data.Address;
          record.PIC = r.data.PIC;
        });
      } else {
        record.DeliveryAddress = "";
        record.PIC = "";
      }
      break;
  }
  if (["PODate", "DueDate"].includes(name)) {
    setMinDate(record);
  }
}

function onPreSave(record) {
  if (record.Status == "") {
    record.Status = "DRAFT";
  }
  data.record.Freight = parseFloat(data.record.Freight);
}

function onPostSave(record) {
  if (
    record.Status === "DRAFT" &&
    gridAttachment.value &&
    data.formMode == "edit"
  ) {
    gridAttachment.value.Save();
  }
}
function onPostPickFile(file) {
  data.fileAttach = file;
}

function submitSaveAsset(record, action) {
  axios
    .post("/scm/asset/upload", {
      JournalType: "Purchase Order",
      JournalID: record._id,
      Assets: [data.fileAttach],
    })
    .then(() => {
      if (action) {
        action();
        data.isPostFlow = false;
      }
    })
    .catch((e) => {
      util.showError(e);
      data.disableButton = false;
      listControl.value.setFormLoading(false);
    });
}
function preSubmit(status, action, doSubmit) {
  data.stayOnForm = data.record.Status == "DRAFT" ? true : false;
  const dim = data.record.Location;
  const pc = data.record.Dimension.find((d) => {
    return d.Key == "PC";
  }).Value;
  const cc = data.record.Dimension.find((d) => {
    return d.Key == "CC";
  }).Value;
  const site = data.record.Dimension.find((d) => {
    return d.Key == "Site";
  }).Value;

  data.isPostFlow = true;
  if (status == "DRAFT") {
    data.record.Freight = parseFloat(data.record.Freight);
    if (!InventDimControl.value) {
      if (
        !data.record.Name ||
        !data.record.PostingProfileID ||
        !data.record.JournalTypeID
      ) {
        return util.showError(
          "General field Name, JournalTypeID, PostingProfileID  is required"
        );
      }
      if (!dim.WarehouseID) {
        return util.showError("General Warehouse ID is required");
      }
    }

    if (data.record.Lines.length == 0) {
      return util.showError("Line items is empty");
    }

    for (let l = 0; l < data.record.Lines.length; l++) {
      if (!data.record.Lines[l].ItemID) {
        return util.showError("there is a line field Item required");
      }
      if (typeof data.record.Lines[l].Qty != "number") {
        return util.showError("there is a line field Qty required");
      }
      if (typeof data.record.Lines[l].UnitCost != "number") {
        return util.showError("there is a line field UnitCost required");
      }
      if (data.record.Lines[l].UnitCost == 0 && data.siteUser != headOffice) {
        return util.showError("UnitCost item which is 0");
      }
      if (
        data.record.Lines[l].PRID &&
        data.record.Lines[l].Qty > data.record.Lines[l].RemainingQty
      ) {
        return util.showError("qty more than Remaining qty");
      }
    }

    setRequiredAllField(true);
    util.nextTickN(2, () => {
      let valid = listControl.value.formValidate();
      let validDimension = true;
      [
        "PostingProfileID",
        "VendorID",
        "BillingName",
        "BillingAddress",
        "PIC",
      ].every((e) => {
        if (!data.record[e]) {
          valid = false;
          return false;
        }
        return true;
      });

      if (FinancialDimension.value) {
        validDimension = FinancialDimension.value.validate();
      } else {
        if (!pc || !cc || !site) {
          return util.showError("Financial Dimension is required");
        }
      }
      if (validDimension && valid) {
        listControl.value.setFormLoading(true);
        data.disableButton = true;
        const payload = JSON.parse(JSON.stringify(data.record));
        payload.Lines.map(function (v) {
          v.RemainingQty = v.RemainingQty - v.Qty;
        });
        listControl.value.submitForm(
          payload,
          () => {
            if (
              status === "DRAFT" &&
              lodash.isEmpty(data.fileAttach) === false &&
              data.isPostFlow
            ) {
              submitSaveAsset(payload, doSubmit);
            } else {
              doSubmit();
            }
          },
          () => {
            data.disableButton = false;
            listControl.value.setFormLoading(false);
          }
        );
      } else {
        return util.showError("Please check general required field");
      }
      setRequiredAllField(false);
    });
  }
}
function postSubmit(record) {
  data.disableButton = false;
  listControl.value.setFormLoading(false);
  data.appMode = "grid";
  listControl.value.refreshList();
  listControl.value.refreshForm();
  listControl.value.setControlMode("grid");
}
function errorSubmit(e) {
  data.disableButton = false;
  listControl.value.setFormLoading(false);
  return util.showError(e);
}
function setRequiredAllField(required) {
  listControl.value.getFormAllField().forEach((e) => {
    if (
      ["PostingProfileID", "BillingName", "BillingAddress", "PIC"].includes(
        e.field
      )
    ) {
      listControl.value.setFormFieldAttr(e.field, "required", required);
    }
  });
}

function setFormMode(mode) {
  listControl.value.setFormMode(mode);
}
function onControlModeChanged(mode) {
  if (mode === "grid") {
    data.titleForm = "Purchase Order";
  }
}

function alterGridConfig(cfg) {
  cfg.sortable = ["Created", "TrxDate", "_id"];
  cfg.setting.idField = "Created";
  cfg.setting.sortable = ["Created", "TrxDate", "_id"];
}

function alterFormConfig(cfg) {}
function getJournalType(_id, action) {
  axios
    .post("/scm/purchase/order/journal/type/find?_id=" + _id, {
      sort: ["-_id"],
    })
    .then(
      (r) => {
        data.journalTypeData = r.data[0];
        setDefaultRecordFromJournalType(r.data[0], action);
        setLockForm(r.data[0]);
      },
      (e) => util.showError(e)
    );
}
function setLockForm(JT) {
  data.lockDimension = JT.LockFinancialDimension;
  data.lockInventDimension = JT.LockInventoryDimension;
  data.lockPostingProfile = JT.LockPostingProfile;
}
function setDefaultRecordFromJournalType(JT, action) {
  if (["", "DRAFT"].includes(data.record.Status)) {
    if (action === "change") {
      // data.record.Location = JT.InventoryDimension;
      // data.record.Dimension = JT.Dimension;
      data.record.PostingProfileID = JT.PostingProfileID;
      nextTick(() => {
        data.keyDimension = util.uuid();
      });
    }
  }
}

function onFieldChanged(val, record) {
  if (val.WarehouseID) {
    data.disableField.push("WarehouseID");
    record.DeliveryAddress = "";
    record.PIC = "";
    record.DeliveryName = "";
    axios.post("/tenant/warehouse/get", [val.WarehouseID]).then((r) => {
      if (r.data) {
        record.DeliveryAddress = r.data.Address;
        record.PIC = r.data.PIC;
        record.DeliveryName = r.data.Name;
      }
      data.keyPIC = util.uuid();
    });
  }
  if (val.AisleID) {
    data.disableField.push("AisleID");
  }
  if (val.SectionID) {
    data.disableField.push("SectionID");
  }
  if (val.BoxID) {
    data.disableField.push("BoxID");
  }
  if (!lineConfig.value) {
    return true;
  }
  const line = lineConfig.value.getDataValue();
  if (line.length > 0) {
    for (let l = 0; l < line.length; l++) {
      line[l].InventDim = val;
    }
  }
}

function getPostingProfile(record) {
  util.nextTickN(2, () => {
    listControl.value.setFormLoading(true);
    axios
      .post(`/scm/purchase/order/journal/type/find?TrxType=Purchase Order`)
      .then(
        (r) => {
          if (r.data.length > 0) {
            record.JournalTypeID = r.data[0]._id;
            record.PostingProfileID = r.data[0].PostingProfileID;
            data.record = record;
            data.keyJournalType = util.uuid();
          }
        },
        (e) => util.showError(e)
      )
      .finally(function () {
        listControl.value.setFormLoading(false);
      });
  });
}
function getDetailEmployee(_id, record) {
  let payload = [];
  if (_id) {
    payload = [_id];
  }
  data.keyDimension = util.uuid();
  axios.post("/tenant/employee/get-emp-warehouse", payload).then(
    (r) => {
      record.DeliveryAddress = r.data.Address;
      record.PIC = r.data.PIC;
      record.DeliveryName = r.data.Name;
      data.record = record;
      data.keyPIC = util.uuid();
    },
    (e) => util.showError(e)
  );
}
function onChangeDimension(field, v1, v2, item) {
  switch (field) {
    case "Site":
      item.PIC = "";
      item.DeliveryName = "";
      item.DeliveryAddress = "";
      item.Location.WarehouseID = "";
      break;
    default:
      break;
  }
}
function lookupPayloadBuilder(search, select, value, item) {
  const qp = {};
  if (search != "") data.filterTxt = search;
  qp.Take = 20;
  qp.Sort = [select[0]];
  qp.Select = select;

  //setting search
  const Site =
    item.Dimension &&
    item.Dimension.find((_dim) => _dim.Key === "Site") &&
    item.Dimension.find((_dim) => _dim.Key === "Site")["Value"] != ""
      ? item.Dimension.find((_dim) => _dim.Key === "Site")["Value"]
      : undefined;
  const querySite = [
    {
      Field: "Dimension.Key",
      Op: "$eq",
      Value: "Site",
    },
    {
      Field: "Dimension.Value",
      Op: "$eq",
      Value: Site,
    },
  ];
  if (Site) {
    qp.Where = {
      Op: "$and",
      items: querySite,
    };
  }
  if (search !== "" && search !== null) {
    let items = [
      {
        Op: "$or",
        items: [
          { Field: "_id", Op: "$contains", Value: [search] },
          { Field: "Name", Op: "$contains", Value: [search] },
        ],
      },
    ];
    if (Site) {
      items = [...items, ...querySite];
    }
    qp.Where = {
      Op: "$and",
      items: items,
    };
  }
  return qp;
}
function getByCurrentUser() {
  axios.post("/tenant/employee/get-by-current-user").then(
    (r) => {
      let Site = "";
      Site =
        r.data.Dimension &&
        r.data.Dimension.find((_dim) => _dim.Key === "Site") &&
        r.data.Dimension.find((_dim) => _dim.Key === "Site")["Value"] != ""
          ? r.data.Dimension.find((_dim) => _dim.Key === "Site")["Value"]
          : undefined;

      data.siteUser = Site;
    },
    (e) => util.showError(e)
  );
}
watch(
  () => route.query.type,
  (nv) => {
    data.transactionType = route.query.type;

    util.nextTickN(2, () => {
      data.appMode = "grid";
    });
  }
);

watch(
  () => route.query.title,
  (nv) => {
    data.titleForm = nv;
  }
);
watch(
  () => data.record.PRDate,
  (nv) => {
    setMinDate(data.record);
  }
);
onMounted(() => {
  getByCurrentUser();
});

function onPreview() {
  data.appMode = "preview";
}

function closePreview() {
  data.appMode = "grid";
}
const InitGridConfig = {
  setting: {
    idField: "",
    keywordFields: ["_id", "Name"],
    sortable: ["_id"],
  },
  fields: [
    {
      field: "PurchaseRequestID",
      kind: "text",
      label: "Purchase Request ID",
      halign: "start",
      valign: "start",
      labelField: "",
      length: 0,
      width: "",
      pos: 1000,
      readType: "show",
      decimal: 0,
      dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
      unit: "",
      input: {
        field: "PurchaseRequestID",
        label: "Purchase Request ID",
        hint: "",
        hide: true,
        placeHolder: "Purchase Request ID",
        kind: "hidden",
        disable: false,
        required: false,
        multiple: false,
        multiRow: 1,
        minLength: 0,
        maxLength: 999,
        readOnly: true,
        readOnlyOnEdit: false,
        readOnlyOnNew: false,
        useList: false,
        allowAdd: false,
        items: [],
        useLookup: false,
        lookupUrl: "",
        lookupKey: "",
        lookupLabels: null,
        lookupSearchs: null,
        lookupFormat1: "",
        lookupFormat2: "",
        showTitle: false,
        showHint: false,
        showDetail: false,
        fixTitle: false,
        fixDetail: false,
        section: "General",
        sectionWidth: "",
        row: 0,
        col: 0,
        labelField: "",
        decimal: 0,
        dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
        unit: "",
        width: "",
        spaceBefore: 0,
        spaceAfter: 0,
      },
    },
    {
      field: "FileName",
      kind: "text",
      label: "File name",
      halign: "start",
      valign: "start",
      labelField: "",
      length: 0,
      width: "",
      pos: 1000,
      readType: "show",
      decimal: 0,
      dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
      unit: "",
      input: {
        field: "FileName",
        label: "File name",
        hint: "",
        hide: false,
        placeHolder: "File name",
        kind: "text",
        disable: false,
        required: false,
        multiple: false,
        multiRow: 1,
        minLength: 0,
        maxLength: 999,
        readOnly: true,
        readOnlyOnEdit: false,
        readOnlyOnNew: false,
        useList: false,
        allowAdd: false,
        items: [],
        useLookup: false,
        lookupUrl: "",
        lookupKey: "",
        lookupLabels: null,
        lookupSearchs: null,
        lookupFormat1: "",
        lookupFormat2: "",
        showTitle: false,
        showHint: false,
        showDetail: false,
        fixTitle: false,
        fixDetail: false,
        section: "General",
        sectionWidth: "",
        row: 0,
        col: 0,
        labelField: "",
        decimal: 0,
        dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
        unit: "",
        width: "",
        spaceBefore: 0,
        spaceAfter: 0,
      },
    },
    {
      field: "Description",
      kind: "text",
      label: "Description",
      halign: "start",
      valign: "start",
      labelField: "",
      length: 0,
      width: "",
      pos: 1000,
      readType: "show",
      decimal: 0,
      dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
      unit: "",
      input: {
        field: "Description",
        label: "Description",
        hint: "",
        hide: false,
        placeHolder: "Description",
        kind: "text",
        disable: false,
        required: false,
        multiple: false,
        multiRow: 1,
        minLength: 0,
        maxLength: 999,
        readOnly: false,
        readOnlyOnEdit: false,
        readOnlyOnNew: false,
        useList: false,
        allowAdd: false,
        items: [],
        useLookup: false,
        lookupUrl: "",
        lookupKey: "",
        lookupLabels: null,
        lookupSearchs: null,
        lookupFormat1: "",
        lookupFormat2: "",
        showTitle: false,
        showHint: false,
        showDetail: false,
        fixTitle: false,
        fixDetail: false,
        section: "General",
        sectionWidth: "",
        row: 0,
        col: 0,
        labelField: "",
        decimal: 0,
        dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
        unit: "",
        width: "",
        spaceBefore: 0,
        spaceAfter: 0,
      },
    },
    {
      field: "UploadDate",
      kind: "date",
      label: "Upload date",
      halign: "start",
      valign: "start",
      labelField: "",
      length: 0,
      width: "",
      pos: 1000,
      readType: "show",
      decimal: 0,
      dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
      unit: "",
      input: {
        field: "UploadDate",
        label: "Upload date",
        hint: "",
        hide: false,
        placeHolder: "Upload date",
        kind: "date",
        disable: false,
        required: false,
        multiple: false,
        multiRow: 1,
        minLength: 0,
        maxLength: 999,
        readOnly: true,
        readOnlyOnEdit: false,
        readOnlyOnNew: false,
        useList: false,
        allowAdd: false,
        items: [],
        useLookup: false,
        lookupUrl: "",
        lookupKey: "",
        lookupLabels: null,
        lookupSearchs: null,
        lookupFormat1: "",
        lookupFormat2: "",
        showTitle: false,
        showHint: false,
        showDetail: false,
        fixTitle: false,
        fixDetail: false,
        section: "General",
        sectionWidth: "",
        row: 0,
        col: 0,
        labelField: "",
        decimal: 0,
        dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
        unit: "",
        width: "",
        spaceBefore: 0,
        spaceAfter: 0,
      },
    },
    {
      field: "Content",
      kind: "file",
      label: "File",
      halign: "start",
      valign: "start",
      labelField: "",
      length: 0,
      width: "",
      pos: 1000,
      readType: "show",
      decimal: 0,
      dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
      unit: "",
      input: {
        field: "Content",
        label: "File",
        hint: "",
        hide: false,
        placeHolder: "File",
        kind: "file",
        disable: false,
        required: false,
        multiple: false,
        multiRow: 1,
        minLength: 0,
        maxLength: 999,
        readOnly: false,
        readOnlyOnEdit: false,
        readOnlyOnNew: false,
        useList: false,
        allowAdd: false,
        items: [],
        useLookup: false,
        lookupUrl: "",
        lookupKey: "",
        lookupLabels: null,
        lookupSearchs: null,
        lookupFormat1: "",
        lookupFormat2: "",
        showTitle: false,
        showHint: false,
        showDetail: false,
        fixTitle: false,
        fixDetail: false,
        section: "General",
        sectionWidth: "",
        row: 0,
        col: 0,
        labelField: "",
        decimal: 0,
        dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
        unit: "",
        width: "",
        spaceBefore: 0,
        spaceAfter: 0,
      },
    },
  ],
};
</script>
