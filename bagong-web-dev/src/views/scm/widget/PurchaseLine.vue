<template>
  <div class="flex flex-col gap-2">
    <s-modal :display="false" ref="confirmModal" @submit="doGenerateLine">
      You will reset the existing data ! Are you sure ?<br />
      Please be noted, this can not be undone !
    </s-modal>
    <div v-if="data.gridCfg.fields.length == 0" class="loading">
      loading data from server ...
    </div>
    <data-list
      v-else
      v-show="!data.showPRLines"
      ref="listControl"
      title="Inventory Journal Line"
      class="grid-line-items"
      hide-title
      no-gap
      grid-hide-select
      grid-hide-search
      grid-hide-sort
      grid-hide-refresh
      grid-no-confirm-delete
      gridAutoCommitLine
      :grid-hide-new="props.isReff || props.isSource"
      init-app-mode="grid"
      grid-mode="grid"
      form-keep-label
      new-record-type="grid"
      :form-hide-submit="
        ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(
          data.generalRecord.Status
        )
      "
      :grid-hide-control="
        ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(
          data.generalRecord.Status
        )
      "
      :grid-hide-delete="
        ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(
          data.generalRecord.Status
        )
      "
      :grid-editor="['', 'DRAFT'].includes(data.generalRecord.Status)"
      :grid-config="`/scm/purchase/line/gridconfig`"
      :form-config="`/scm/purchase/line/formconfig`"
      :form-fields="['Dimension', 'InventDim', 'InventJournalLine']"
      :grid-fields="[
        'ItemID',
        'UnitID',
        'Dimension',
        'RemainingQty',
        'Qty',
        'UnitCost',
        'SubTotal',
        'DiscountValue',
        'DiscountAmount',
        'PPN',
        'PPH',
      ]"
      :grid-hide-edit="true"
      @grid-row-add="newRecord"
      @form-field-change="onFormFieldChanged"
      @form-edit-data="openForm"
      @grid-row-delete="onGridRowDelete"
      @grid-row-field-changed="onGridRowFieldChanged"
      @grid-row-save="onGridRowSave"
      @post-save="onFormPostSave"
      @grid-refreshed="getTaxCode"
      @alter-grid-config="alterGridConfig"
      @alter-form-config="alterFormConfig"
      form-focus
    >
      <!--  -->
      <template #grid_Dimension="{ item }">
        <DimensionText :dimension="item.Dimension" />
      </template>

      <template #grid_ItemID="{ item, idx }">
        <s-input-sku-item
          v-model="item.ItemVarian"
          :record="item"
          :lookup-url="`/tenant/item/gets-detail?_id=${helper.ItemVarian(
            item.ItemID,
            item.SKU
          )}`"
          :disabled="
            ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(
              data.generalRecord.Status
            ) || props.isReff
          "
          @beforeOnChange="onBeforeOnChange"
          @afterOnChange="onAfterOnChange"
          @cancelOnChange="onCancelOnChange"
        ></s-input-sku-item>
      </template>
      <template #grid_UnitID="{ item }">
        <s-input
          v-model="item.UnitID"
          :key="
            ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(
              data.generalRecord.Status
            ) || props.isReff
              ? item.UnitID
              : ''
          "
          :disabled="
            ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(
              data.generalRecord.Status
            ) || props.isReff
          "
          use-list
          :lookup-url="`/tenant/unit/gets-filter?ItemID=${item.ItemID}`"
          lookup-key="_id"
          :lookup-labels="['Name']"
          :lookup-searchs="['_id', 'Name']"
          class="w-full"
          @change="
            (field, v1, v2, old) => {
              onGridRowFieldChanged('UnitID', v1, v2, old, item);
            }
          "
        ></s-input>
      </template>
      <template #grid_RemainingQty="{ item }">
        <div style="text-align: right">
          {{ helper.formatNumberWithDot(item.RemainingQty) }}
        </div>
      </template>
      <template #grid_Qty="{ item }">
        <s-input
          v-if="
            !['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(
              data.generalRecord.Status
            )
          "
          v-model="item.Qty"
          kind="number"
          class="w-full"
          @change="
            (field, v1, v2, old) => {
              onGridRowFieldChanged('Qty', v1, v2, old, item);
            }
          "
        ></s-input>
        <div style="text-align: right" v-else>
          {{ helper.formatNumberWithDot(item.Qty) }}
        </div>
      </template>
      <template #grid_UnitCost="{ item }">
        <s-input
          v-if="
            !['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(
              data.generalRecord.Status
            )
          "
          v-model="item.UnitCost"
          kind="number"
          class="w-full"
          @change="
            (field, v1, v2, old) => {
              onGridRowFieldChanged('UnitCost', v1, v2, old, item);
            }
          "
        ></s-input>
        <div style="text-align: right" v-else>
          {{ helper.formatNumberWithDot(item.UnitCost) }}
        </div>
      </template>
      <template #grid_SubTotal="{ item }">
        <div style="text-align: right">
          {{ helper.formatNumberWithDot(item.SubTotal) }}
        </div>
      </template>
      <template #grid_DiscountValue="{ item }">
        <s-input
          v-if="
            !['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(
              data.generalRecord.Status
            )
          "
          v-model="item.DiscountValue"
          kind="number"
          class="w-full"
          @change="
            (field, v1, v2, old) => {
              onGridRowFieldChanged('DiscountValue', v1, v2, old, item);
            }
          "
        ></s-input>
        <div style="text-align: right" v-else>
          {{ helper.formatNumberWithDot(item.DiscountValue) }}
        </div>
      </template>

      <template #grid_DiscountAmount="{ item }">
        <div style="text-align: right">
          {{ helper.formatNumberWithDot(item.DiscountAmount) }}
        </div>
      </template>
      <template #grid_PPN="{ item }">
        <div :key="item.PPN" :val="item.PPN" style="text-align: right">
          {{ helper.formatNumberWithDot(item.PPN) }}
        </div>
      </template>
      <template #grid_PPH="{ item }">
        <div :key="item.PPH" :val="item.PPH" style="text-align: right">
          {{ helper.formatNumberWithDot(item.PPH) }}
        </div>
      </template>
      <template #form_input_InventJournalLine="{ item }">
        <s-form
          v-model="item.InventJournalLine"
          :config="data.formCfg"
          :mode="data.formMode"
          keep-label
          only-icon-top
          hide-submit
          hide-cancel
        >
        </s-form>
      </template>
      <template #form_input_Dimension="{ item }">
        <dimension-editor
          v-model="item.Dimension"
          sectionTitle="Financial Dimension"
          :default-list="props.defaultList"
          :readOnly="
            ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(
              data.generalRecord.Status
            ) || data.generalRecord.ReffNo.length > 0
          "
          @change="
            (field, v1, v2) => {
              console.log(v1);
              onChangeDimension(field, v1, v2, item);
            }
          "
        ></dimension-editor>
      </template>
      <template #form_input_InventDim="{ item }">
        <DimensionInventJurnal
          v-model="item.InventDim"
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
            ['SUBMITTED', 'READY', 'REJECTED', 'POSTED'].includes(
              data.generalRecord.Status
            ) || data.generalRecord.ReffNo.length > 0
          "
          @onChange="
            (field, v1, v2, old, val) => {
              onFieldChanged(val, item);
            }
          "
        ></DimensionInventJurnal>
      </template>
      <template #grid_header_buttons_1="{ config }">
        <s-button
          v-if="isGenerate && props.purchaseType === 'purchase/order'"
          icon="file-refresh-outline"
          class="btn_primary mr-2"
          label="Generate PR lines"
          @click="onGeneratePRLine"
        />
      </template>
      <template #grid_item_buttons_1="{ item }">
        <!-- <a href="#" @click="taxFormula(item)" class="save_action"
          ><mdicon
            name="content-save"
            width="16"
            alt="edit"
            class="cursor-pointer hover:text-primary"
        /></a> -->
        <a href="#" @click="setItemDimension(item)" class="save_action"
          ><mdicon
            name="eye"
            width="16"
            alt="edit"
            class="cursor-pointer hover:text-primary"
        /></a>
        <log-item-trx
          v-if="item.ItemID && islog && props.purchaseType === 'purchase/order'"
          :id="item.ItemID"
          :item-id="item.ItemID"
          :sku="item.SKU"
          :hide-button="false"
          :grid-config="`/scm/display/previous/log/gridconfig`"
        />
      </template>
      <template #grid_paging="{ items }"><div></div></template>
      <template #grid_footer_2="{ items }"
        ><div
          v-if="data.gridCfg.fields.length > 0"
          v-show="data.appMode === 'grid' && !data.showPRLines"
          class="gap-2 p-4 ml-[50%]"
        >
          <div class="flex flex-row justify-between">
            <span class="text-base font-semibold">Sub Total</span>
            <span class="text-base font-semibold">
              {{ helper.formatNumberWithDot(subTotal) }}
            </span>
          </div>
          <div class="flex flex-row justify-between">
            <span class="text-base font-semibold">Discount Line</span>
            <span class="text-base font-semibold">
              {{ discount ? helper.formatNumberWithDot(discount) : "-" }}
            </span>
          </div>
          <div class="flex flex-row justify-between">
            <span class="text-base font-semibold">PPN</span>
            <span class="text-base font-semibold">
              {{ PPN ? helper.formatNumberWithDot(PPN) : "-" }}
            </span>
          </div>
          <div class="flex flex-row justify-between">
            <span class="text-base font-semibold">PPh</span>
            <span class="text-base font-semibold">
              {{ PPH ? helper.formatNumberWithDot(PPH) : "-" }}
            </span>
          </div>
          <div>
            <div class="flex flex-row gap-2">
              <span class="text-base font-semibold">Other Expenses</span>
              <s-button
                v-if="['DRAFT', ''].includes(data.generalRecord.Status)"
                icon="plus"
                class="btn_primary refresh_btn"
                tooltip="refresh"
                @click="newRecordExpenses"
              />
            </div>
            <div class="ml-20">
              <s-grid
                ref="listPOExpenses"
                class="w-full grid-line-items"
                hide-sort
                hide-search
                hide-refresh-button
                hide-select
                hide-footer
                v-model="data.detailTotal.OtherExpenses"
                :editor="['DRAFT', ''].includes(data.generalRecord.Status)"
                :hide-new-button="true"
                :hide-delete-button="false"
                :hide-detail="true"
                :hide-action="false"
                :hide-select="false"
                :config="data.gridCfgExpenses"
                total-url="/scm/purchase/request/get-lines"
                auto-commit-line
                no-confirm-delete
                form-keep-label
                @delete-data="deleteRecordExpenses"
              >
                <template #item_Expenses="{ item }">
                  <s-input
                    :key="item.Expenses"
                    class="min-w-[100px]"
                    hide-label
                    use-list
                    :disabled="
                      !['DRAFT', ''].includes(data.generalRecord.Status)
                    "
                    v-model="item.Expenses"
                    lookup-url="/tenant/expensetype/find?GroupID=EXG0019"
                    lookup-key="_id"
                    :lookup-labels="['_id', 'Name']"
                    @change="handleChange"
                  />
                </template>
                <template #grid_total="{ item }">
                  <tr class="font-semibold">
                    <td class="ml-4">Sub Total</td>
                    <td class="text-right">
                      {{ helper.formatNumberWithDot(subTotalOtherExpenses) }}
                    </td>
                  </tr>
                </template>
              </s-grid>
            </div>
          </div>
          <div class="flex flex-row justify-between">
            <span class="text-base font-semibold">Discount Type</span>
            <span class="text-base font-semibold">
              <div class="flex flex-row gap-4">
                <s-input
                  ref="refInput"
                  label="Discount Type"
                  v-model="data.detailTotal.Discount.DiscountType"
                  class="w-[150px]"
                  keepLabel
                  :hide-label="true"
                  :show-clear-button="false"
                  use-list
                  :disabled="!['DRAFT', ''].includes(data.generalRecord.Status)"
                  :items="['fixed', 'percent']"
                  @change="
                    () => {
                      data.detailTotal.Discount.DiscountValue = 0;
                      data.detailTotal.Discount.DiscountAmount = 0;
                    }
                  "
                ></s-input>
                <s-input
                  v-if="data.detailTotal.Discount.DiscountType == 'percent'"
                  class="w-[200px]"
                  hide-label
                  :disabled="!['DRAFT', ''].includes(data.generalRecord.Status)"
                  v-model="data.detailTotal.Discount.DiscountValue"
                  kind="number"
                ></s-input>
              </div>
            </span>
          </div>
          <div class="flex flex-row justify-between mt-2">
            <span class="text-base font-semibold">Discount General</span>
            <span class="text-base font-semibold">
              <div
                class="flex flex-row gap-2"
                v-if="data.detailTotal.Discount.DiscountType == 'fixed'"
              >
                (-)<s-input
                  class="w-[200px]"
                  hide-label
                  v-model="data.detailTotal.Discount.DiscountAmount"
                  kind="number"
                ></s-input>
              </div>

              <div v-else>
                {{ helper.formatNumberWithDot(Math.abs(discountAmount) * -1) }}
              </div>
            </span>
          </div>
          <div class="flex flex-row justify-between">
            <span class="text-base font-semibold">Total Amount</span>
            <span class="text-base font-semibold">
              {{ helper.formatNumberWithDot(grandtotal) }}
            </span>
          </div>
        </div></template
      >
    </data-list>
  </div>

  <div v-show="data.showPRLines">
    <s-card
      :title="`Select Purchase Request Line`"
      class="w-full bg-white suim_datalist card"
      hide-footer
      :no-gap="false"
      :hide-title="false"
    >
      <s-grid
        hide-search
        class="w-full grid-line-items"
        ref="listPRControl"
        hide-sort
        :hide-new-button="true"
        :hide-delete-button="true"
        hide-refresh-button
        :hide-detail="true"
        :hide-action="true"
        auto-commit-line
        no-confirm-delete
        :config="data.prlinegridCfg"
        form-keep-label
      >
        <template #item_ItemID="{ item }">
          <div class="bg-transparent">{{ item.Text }}</div>
        </template>
        <template #item_UnitID="{ item }">
          <div class="bg-transparent">{{ item.UnitID }}</div>
        </template>
        <!-- <template #item_SKU="{ item }">
          <div class="bg-transparent">{{ item.Text }}</div>
        </template> -->
        <template #item_Dimension="{ item }">
          <DimensionText :dimension="item.Dimension" />
        </template>
        <template #header_buttons_1="{ config }">
          <s-button
            icon="refresh"
            class="btn_primary refresh_btn"
            @click="refreshData"
          />
        </template>
        <template #header_buttons="{ config }">
          <div class="flex gap-[2px] ml-2">
            <s-button
              icon="format-list-checks"
              class="btn_primary refresh_btn"
              :label="'Confirm'"
              @click="() => confirmLineSelected()"
            />
            <s-button
              icon="rewind"
              class="btn_warning back_btn"
              :label="'Back'"
              @click="data.showPRLines = false"
            />
          </div>
        </template>
      </s-grid>
    </s-card>
  </div>
</template>
<script setup>
import { reactive, ref, onMounted, inject, computed, watch } from "vue";
import {
  DataList,
  SGrid,
  SInput,
  util,
  SButton,
  SCard,
  loadGridConfig,
  loadFormConfig,
  SModal,
} from "suimjs";
import helper from "@/scripts/helper.js";
import DimensionEditor from "@/components/common/DimensionEditorVertical.vue";
import DimensionText from "@/components/common/DimensionText.vue";
import DimensionInventJurnal from "@/components/common/DimensionInventJurnal.vue";
import SInputSkuItem from "./SInputSkuItem.vue";
import LogItemTrx from "./LogItemTrx.vue";

const axios = inject("axios");
const listControl = ref(null);
const listPOExpenses = ref(null);
const refItemID = ref(null);
const listPRControl = ref(null);
const confirmModal = ref(null);

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  disableField: { type: Array, default: () => [] },
  defaultList: { type: Array, default: () => [] },
  generalRecord: { type: Object, default: () => {} },
  purchaseType: { type: String, default: () => "" },
  isReff: { type: Boolean, default: () => false },
  isSource: { type: Boolean, default: () => false },
  isGenerate: { type: Boolean, default: () => false },
  islog: { type: Boolean, default: () => false },
});
const emit = defineEmits({
  "update:modelValue": null,
  attachmentAction: null,
});
const data = reactive({
  appMode: "grid",
  formMode: "edit",
  showPRLines: false,
  generalRecord: props.generalRecord,
  records: props.modelValue.map((dt) => {
    dt.suimRecordChange = false;
    return dt;
  }),
  disableField: props.disableField,
  detailTotal: {
    TotalAmount: 0,
    TotalDiscountAmount: 0,
    OtherExpenses: [],
    Discount: {
      DiscountType: "percent",
      DiscountValue: 0,
      DiscountAmount: 0,
    },
    GrandTotalAmount: 0,
    tax: 0,
    freight: 0,
  },
  purchaseRequestLines: [],
  listTaxCode: [],
  prlinegridCfg: {},
  prlineformCfg: {},
  gridCfgExpenses: {},
  showConfirmModal: false,
  formCfg: {
    sectionGroups: [],
    setting: {},
  },
  gridCfg: {
    fields: [],
    setting: {},
  },
  searchPRNO: "",
});

function newRecord() {
  const record = {};
  record.LineNo = data.records.length + 1;
  record.ItemVarian = "";
  record.ItemID = "";
  record.SKU = "";
  record.Qty = 0;
  record.UnitID = "";
  record.PRID = "";
  record.Remarks = "";
  record.DiscountAmount = 0;
  record.DiscountValue = 0;
  record.UnitCost = 0;
  record.Taxable = true;
  if (props.generalRecord?.Dimension) {
    record.Dimension = props.generalRecord?.Dimension;
  }
  if (props.generalRecord?.Location) {
    record.InventDim = props.generalRecord?.Location;
  }
  data.records.push(record);
  listControl.value.setGridRecords(data.records);
  updateItems();
}
function openForm(record) {
  util.nextTickN(2, () => {
    if (
      ["SUBMITTED", "REJECTED", "READY", "POSTED"].includes(
        data.generalRecord.Status
      )
    ) {
      listControl.value.setFormFieldAttr("Taxable", "disabled", true);
      setFormMode("view");
      data.formMode = "view";
    }
  });
}

function newRecordExpenses() {
  const record = {};
  record.Expenses = "";
  record.Amount = 0;
  listPOExpenses.value.setRecords([
    ...listPOExpenses.value.getRecords(),
    record,
  ]);
}
function deleteRecordExpenses(record, index) {
  const newRecords = record.items.filter((dt, idx) => {
    return idx != index;
  });
  listPOExpenses.value.setRecords(newRecords);
  data.detailTotal.OtherExpenses = newRecords;
}

function setFormMode(mode) {
  listControl.value.setFormMode(mode);
}
function onFormPostSave(record, index) {
  record.suimRecordChange = false;
  if (listControl.value.getFormMode() == "new") {
    data.records.push(record);
  } else {
    data.records[index] = record;
  }
  listControl.value.setGridRecords(data.records);
  updateItems();
}
function onGridRowSave(record, index) {
  record.suimRecordChange = false;
  data.records[index] = record;
  listControl.value.setGridRecords(data.records);
  updateItems();
}
function onGridRowDelete(record, index) {
  const newRecords = data.records.filter((dt, idx) => {
    return idx != index;
  });
  data.records = newRecords;
  listControl.value.setGridRecords(data.records);
  updateItems();
}
function onFormFieldChanged(name, v1, v2, old, record) {
  console.log(name, v1, v2, old, record);
}

function updateItems() {
  util.nextTickN(2, () => {
    const committedRecords = data.records;
    emit("update:modelValue", committedRecords);
    emit("recalc");
  });
}
function gridRefreshed() {
  const records = JSON.parse(JSON.stringify(data.records));
  records.map((r) => {
    r.ItemVarian = helper.ItemVarian(r.ItemID, r.SKU);
    taxFormula(r);
    return r;
  });
  data.records = records;
  listControl.value.setGridRecords(records);
}

function onCancelOnChange(item) {
  item.UnitCost = 0;
  item.SubTotal = item.Qty * item.UnitCost;
  item.TaxCodes = [];
  item.PPNTaxCodes = [];
  item.PPHTaxCodes = [];
  item.PPN = 0;
  item.PPH = 0;
}
function onBeforeOnChange(item) {
  item.UnitCost = 0;
  item.SubTotal = item.Qty * item.UnitCost;
  item.TaxCodes = [];
  item.PPNTaxCodes = [];
  item.PPHTaxCodes = [];
  item.PPN = 0;
  item.PPH = 0;
}
function onAfterOnChange(item) {
  if (props.generalRecord.VendorID) {
    axios
      .post(
        `/scm/vendor/pricelist/find?VendorID=${props.generalRecord.VendorID}&ItemID=${item.ItemID}&SKU=${item.SKU}`
      )
      .then(
        (r) => {
          if (r.data.length > 0) {
            item.UnitCost = r.data[0].Price;
          } else {
            item.UnitCost = 0;
          }
          item.SubTotal = item.Qty * item.UnitCost;
          taxFormula(item);
        },
        (e) => {
          item.UnitCost = 0;
          item.SubTotal = item.Qty * item.UnitCost;
          util.showError(e);
        }
      );
  }
}
function getItemLine(item) {
  axios.post("/tenant/item/get", [item.ItemID]).then(
    (r) => {
      item.Item = r.data;
      formulaPPNPPh(item);
    },
    (e) => {
      item.Item = null;
    }
  );
}
function setItemDimension(item) {
  listControl.value.setFormRecord(item);
  listControl.value.setControlMode("form");
}
function onChangeDimension(field, v1, v2, item) {
  data.records.forEach((dt, index) => {
    if (dt.LineNo == item.LineNo) {
      item.InventDim.WarehouseID = "";
      data.records[index] = item;
    }
  });
  updateItems();
}
function onFieldChanged(val, item) {
  console.log(val, item);
  data.records.forEach((dt, index) => {
    if (dt.LineNo == item.LineNo) {
      data.records[index] = item;
    }
  });
  updateItems();
}

function taxFormula(item) {
  if (
    props.generalRecord.VendorID &&
    ["", "DRAFT"].includes(props.generalRecord.Status)
  ) {
    if (!item.Item) {
      getItemLine(item);
    } else {
      formulaPPNPPh(item);
    }
  }
}
function formulaPPNPPh(item) {
  let TaxCodes = [];
  let TaxCodesItemPPH = [];
  let PPNTaxCodes = [];
  let PPHTaxCodes = [];
  if (!item.Item.TaxCodes || item.Item.TaxCodes.length == 0) {
    TaxCodes = data.generalRecord.TaxCodes;
    TaxCodesItemPPH = [];
  } else {
    for (let i in data.generalRecord.TaxCodes) {
      let taxcode = item.Item.TaxCodes.find((tax) => {
        return tax == data.generalRecord.TaxCodes[i];
      });
      if (taxcode) {
        TaxCodes.push(taxcode);
      }
    }

    for (let c in item.Item.TaxCodes) {
      let taxPPH = data.listTaxCode.find((tax) => {
        return tax._id == item.Item.TaxCodes[c];
      });

      if (taxPPH && taxPPH.TaxGroup == "TG002") {
        TaxCodesItemPPH.push(item.Item.TaxCodes[c]);
      }
    }
  }

  for (let i in TaxCodes) {
    let taxcode = data.listTaxCode.find((tax) => {
      return tax._id == TaxCodes[i];
    });
    if (
      taxcode &&
      taxcode.TaxGroup == "TG001" &&
      data.generalRecord.TaxCodes.includes(TaxCodes[i])
    ) {
      PPNTaxCodes.push(TaxCodes[i]);
    } else if (
      taxcode &&
      taxcode.TaxGroup == "TG002" &&
      data.generalRecord.TaxCodes.includes(TaxCodes[i])
    ) {
      PPHTaxCodes.push(TaxCodes[i]);
    }
  }

  item.TaxCodes = TaxCodes;
  item.PPNTaxCodes = PPNTaxCodes;
  item.PPHTaxCodes = PPHTaxCodes;

  let totalPPN = 0;
  let totalPPH = 0;
  for (let i in PPNTaxCodes) {
    let taxPPN = data.listTaxCode.find((tax) => {
      return tax._id == PPNTaxCodes[i];
    });
    if (taxPPN && taxPPN.InvoiceOperation == "Increase") {
      totalPPN = totalPPN + taxPPN.Rate * (item.SubTotal - item.DiscountAmount);
    } else if (taxPPN && taxPPN.InvoiceOperation == "Decrease") {
      totalPPN = totalPPN - taxPPN.Rate * (item.SubTotal - item.DiscountAmount);
    }
  }

  if (
    !["GRP0023"].includes(item.Item.ItemGroupID) &&
    TaxCodesItemPPH.length > 0
  ) {
    for (let i in PPHTaxCodes) {
      let taxPPH = data.listTaxCode.find((tax) => {
        return tax._id == PPHTaxCodes[i];
      });
      if (taxPPH && taxPPH.InvoiceOperation == "Increase") {
        totalPPH =
          totalPPH + taxPPH.Rate * (item.SubTotal - item.DiscountAmount);
      } else if (taxPPH && taxPPH.InvoiceOperation == "Decrease") {
        totalPPH =
          totalPPH - taxPPH.Rate * (item.SubTotal - item.DiscountAmount);
      }
    }
  } else if (["GRP0023"].includes(item.Item.ItemGroupID)) {
    for (let i in PPHTaxCodes) {
      let taxPPH = data.listTaxCode.find((tax) => {
        return tax._id == PPHTaxCodes[i];
      });
      if (taxPPH && taxPPH.InvoiceOperation == "Increase") {
        totalPPH =
          totalPPH + taxPPH.Rate * (item.SubTotal - item.DiscountAmount);
      } else if (taxPPH && taxPPH.InvoiceOperation == "Decrease") {
        totalPPH =
          totalPPH - taxPPH.Rate * (item.SubTotal - item.DiscountAmount);
      }
    }
  }

  if (item.Taxable) {
    item.PPN = totalPPN;
    item.PPH = totalPPH;
  } else {
    item.PPN = 0;
    item.PPH = 0;
    item.TaxCodes = [];
    item.PPNTaxCodes = [];
    item.PPHTaxCodes = [];
  }
  console.log("formulaPPNPPh ====================", item);
  console.log("formulaPPNPPh ===================", props.generalRecord.Status);
  updateItems();
}
function getTaxCode() {
  axios.post("/fico/taxsetup/find").then(
    (r) => {
      data.listTaxCode = r.data;
      gridRefreshed();
    },
    (e) => {
      util.showError(e);
    }
  );
}

function onGridRowFieldChanged(name, v1, v2, old, record) {
  data.records = listControl.value.getGridRecords();
  if (["Qty", "UnitCost"].includes(name)) {
    util.nextTickN(2, () => {
      record.SubTotalBeforeDiscount = record.Qty * record.UnitCost;
      record.SubTotal = record.SubTotalBeforeDiscount;
      if (record.DiscountType === "fixed") {
        record.DiscountAmount = record.DiscountValue ? record.DiscountValue : 0;
      } else if (record.DiscountType === "percent") {
        record.DiscountAmount =
          record.SubTotalBeforeDiscount *
          (parseFloat(record.DiscountValue ? record.DiscountValue : 0) / 100);
      }
      taxFormula(record);
    });
  }
  switch (name) {
    case "DiscountValue":
      if (record.DiscountType === "fixed") {
        record.DiscountAmount = v1;
      } else if (record.DiscountType === "percent") {
        record.DiscountAmount =
          record.SubTotalBeforeDiscount * (parseFloat(v1) / 100);
      }
      taxFormula(record);
      break;
    case "DiscountType":
      if (v1 === "fixed") {
        record.DiscountAmount = record.DiscountValue;
      } else if (v1 === "percent") {
        record.DiscountAmount =
          record.SubTotal * (parseFloat(record.DiscountValue) / 100);
      } else {
        record.DiscountValue = 0;
        record.DiscountAmount = 0;
      }
      taxFormula(record);
      break;
    case "UnitID":
      record.UnitID = v1;
      taxFormula(record);
      break;
    case "Taxable":
      record.Taxable = v1;
      // taxFormula(record);
      getItemLine(record);
      break;
    default:
      break;
  }
  updateItems();
}
const subTotal = computed({
  get() {
    const total = data.records.reduce((a, b) => {
      let val = 0;
      if (b.SubTotalBeforeDiscount) {
        val = b.SubTotalBeforeDiscount;
      }
      return a + val;
    }, 0);
    props.generalRecord.TotalAmount = total;
    return total;
  },
});
const discount = computed({
  get() {
    const total = data.records.reduce((a, b) => {
      let val = 0;
      if (b.DiscountAmount) {
        val = b.DiscountAmount;
      }
      return a + val;
    }, 0);
    props.generalRecord.TotalDiscountAmount = Math.abs(total) * -1;
    return Math.abs(total) * -1;
  },
});
const PPN = computed({
  get() {
    const sum = data.records.reduce((accumulator, object) => {
      let val = 0;
      if (object.PPN) {
        val = object.PPN;
      }
      return accumulator + val;
    }, 0);
    props.generalRecord.PPN = sum;
    return sum;
  },
});
const PPH = computed({
  get() {
    const sum = data.records.reduce((accumulator, object) => {
      let val = 0;
      if (object.PPH) {
        val = object.PPH;
      }
      return accumulator + val;
    }, 0);
    props.generalRecord.PPH = sum;
    return sum;
  },
});
const subTotalOtherExpenses = computed({
  get() {
    let OtherExpenses = data.detailTotal.OtherExpenses
      ? data.detailTotal.OtherExpenses
      : [];
    let sum = OtherExpenses.reduce(function (acc, obj) {
      return acc + (typeof obj.Amount == "number" ? obj.Amount : 0);
    }, 0);
    return sum;
  },
});
const discountAmount = computed({
  get() {
    let total = 0;
    if (data.detailTotal.Discount.DiscountType == "percent") {
      total =
        (subTotal.value - (discount.value + PPN.value + PPH.value)) *
        (parseFloat(
          data.detailTotal.Discount.DiscountValue
            ? data.detailTotal.Discount.DiscountValue
            : 0
        ) /
          100);
    } else {
      total = data.detailTotal.Discount.DiscountValue
        ? data.detailTotal.Discount.DiscountValue
        : 0;
    }
    data.detailTotal.Discount.DiscountAmount = Math.abs(total) * -1;
    props.generalRecord.TotalTaxAmount = total * -1;
    return total;
  },
});

const grandtotal = computed({
  get() {
    let OtherExpenses = data.detailTotal.OtherExpenses
      ? data.detailTotal.OtherExpenses
      : [];
    let totalOtherExpenses = OtherExpenses.reduce(function (acc, obj) {
      return acc + (typeof obj.Amount == "number" ? obj.Amount : 0);
    }, 0);

    let total =
      subTotal.value +
      (discount.value + PPN.value + PPH.value + totalOtherExpenses);

    if (data.detailTotal.Discount.DiscountType == "percent") {
      let disc = data.detailTotal.Discount.DiscountValue
        ? data.detailTotal.Discount.DiscountValue
        : 0;
      let tot = (total * disc) / 100;
      total = total - tot; // discountAmount.value;
    } else {
      total = total - data.detailTotal.Discount.DiscountAmount;
    }
    props.generalRecord.GrandTotalAmount = total;
    return total;
  },
});

const tax = computed({
  get() {
    props.generalRecord.TotalTaxAmount = 0;
    return 0;
  },
});
function onGeneratePRLine() {
  listPRControl.value.setLoading(true);
  data.showPRLines = true;
  let Site =
    props.generalRecord.Dimension.length == 0
      ? ""
      : props.generalRecord?.Dimension?.find((o) => o.Key === "Site").Value;
  axios
    .post("/scm/purchase/request/get-lines", {
      VendorID: props.generalRecord?.VendorID ?? "",
      DeliveryTo: props.generalRecord?.Location?.WarehouseID ?? "",
      Site: Site,
    })
    .then((r) => {
      listPRControl.value.setRecords(r.data);
      listPRControl.value.setLoading(false);
    });
}

function confirmLineSelected() {
  if (data.records.length > 0) {
    confirmModal.value.show();
  } else {
    doGenerateLine();
  }
}

function onlyUnique(value, index, array) {
  return array.indexOf(value) === index;
}

async function doGenerateLine() {
  confirmModal.value.hide();

  const list = listPRControl.value.getSelected();
  let records = [];
  let listPRNo = [];
  let listPRDate = [];
  let listWarehouseIDs = [];
  const promiseBatchs = [];

  for (const value of list.value) {
    promiseBatchs.push(await GetDataAssets(value.PRID, "Purchase Request"));
  }

  emit(
    "attachmentAction",
    (await Promise.all(promiseBatchs)).flatMap((btc) => btc)
  );

  list.value.map((item) => {
    const record = {};
    record.ItemVarian = helper.ItemVarian(item.ItemID, item.SKU);
    record.LineNo = records.length + 1;
    record.ItemID = item.ItemID;
    record.SKU = item.SKU;
    record.Text = item.Text;
    record.Qty = item.Qty;
    record.RemainingQty = item.RemainingQty;
    record.PRID = item.PRID;
    record.UnitID = item.UnitID;
    record.UnitCost = item.UnitCost;
    record.OffsetAccount = item.OffsetAccount;
    record.Taxable = item.Taxable;
    record.DiscountType = item.DiscountType;
    record.DiscountValue = item.DiscountValue;
    record.DiscountAmount = item.DiscountAmount;
    record.SubTotalBeforeDiscount = item.SubTotalBeforeDiscount;
    record.SubTotal = item.SubTotal;
    record.Remarks = item.Remarks;
    record.InventDim = item.InventDim;
    records.push(record);
    listPRNo.push(item.PRID);
    listPRDate.push(item.PRDate);
    listWarehouseIDs.push(item.WarehouseID);
  });
  if (listWarehouseIDs.filter(onlyUnique).length > 1) {
    return util.showError("Delivery to must be the same destination");
  }
  let orderedDates = listPRDate.sort(function (a, b) {
    return Date.parse(a) - Date.parse(b);
  });
  let validDate = orderedDates.filter((o) => o !== null);
  if (validDate.length > 0) {
    props.generalRecord.PRDate = orderedDates[0];
  }
  props.generalRecord.ReffNo = listPRNo.filter(onlyUnique);
  axios
    .post("/scm/purchase/request/get", [props.generalRecord.ReffNo[0]])
    .then((r) => {
      props.generalRecord.WarehouseID = r.data.WarehouseID;
      props.generalRecord.DeliveryName = r.data.DeliveryName;
      props.generalRecord.DeliveryAddress = r.data.DeliveryAddress;
      props.generalRecord.PIC = r.data.PIC;
      props.generalRecord.BillingName = r.data.BillingName;
      props.generalRecord.BillingAddress = r.data.BillingAddress;
    });
  data.records = records;
  listControl.value.setGridRecords(data.records);

  updateItems();
  util.nextTickN(2, () => {
    data.showPRLines = false;
  });
}

function alterGridConfig(cfg) {
  const gridfields = data.gridCfg.fields.filter(
    (o) => !["Remarks", "Dimension"].includes(o.field)
  );
  cfg.fields = [
    ...gridfields,
    ...cfg.fields.filter(
      (o) => !["InventJournalLine", "SourceLineNo"].includes(o.field)
    ),
  ];
  let sortColum = [
    "ItemID",
    "UnitID",
    "RemainingQty",
    "ReceivedQty",
    "Qty",
    "UnitCost",
    "SubTotal",
    "DiscountType",
    "DiscountValue",
    "DiscountAmount",
    "Taxable",
    "PPN",
    "PPH",
    "PPH",
    "Total",
    "Remarks",
  ];
  cfg.fields.map(function (fields) {
    fields.idx = sortColum.indexOf(fields.field);

    if (fields.field == "ItemID") {
      fields.width = "500px";
      fields.label = "Item Varian";
    } else if (fields.field == "SKU") {
      fields.readType = "hide";
    } else if (fields.field == "DiscountType") {
      fields.width = "180px";
    } else if (fields.field == "UnitID") {
      fields.input.readOnly = props.isReff;
      fields.width = "150px";
      fields.label = "UoM";
    } else if (fields.field == "Qty") {
      fields.input.readOnly = props.isReff;
      fields.width = "180px";
    } else if (
      fields.field == "RemainingQty" &&
      props.purchaseType == "purchase/request" &&
      props.generalRecord.Status != "POSTED"
    ) {
      fields.readType = "hide";
      fields.input.readOnly = true;
    } else if (
      fields.field == "RemainingQty" &&
      props.purchaseType == "purchase/request" &&
      props.generalRecord.Status == "POSTED"
    ) {
      fields.readType = "show";
      fields.width = "180px";
      fields.input.readOnly = true;
    } else if (fields.field == "PPN") {
      fields.input.readOnly = true;
      fields.width = "180px";
      fields.label = "PPN";
    } else if (fields.field == "PPH") {
      fields.input.readOnly = true;
      fields.width = "180px";
      fields.label = "PPh";
    } else if (fields.field == "Taxable") {
      fields.width = "100px";
    } else if (
      fields.field == "RemainingQty" &&
      props.purchaseType == "purchase/order" &&
      props.generalRecord.Status == "POSTED"
    ) {
      fields.input.readOnly = true;
      fields.readType = "hide";
    } else if (
      fields.field == "RemainingQty" &&
      props.purchaseType == "purchase/order" &&
      props.generalRecord.Status != "POSTED"
    ) {
      fields.readType = "show";
      fields.width = "180px";
      fields.input.readOnly = true;
    } else if (fields.field == "Remarks") {
      fields.width = "280px";
    } else if (
      fields.field == "ReceivedQty" &&
      props.purchaseType == "purchase/order" &&
      props.generalRecord.Status == "POSTED"
    ) {
      fields.width = "200px";
      fields.readType = "show";
    } else if (
      fields.field == "ReceivedQty" &&
      props.purchaseType == "purchase/order" &&
      props.generalRecord.Status != "POSTED"
    ) {
      fields.readType = "hide";
    } else if (
      fields.field == "ReceivedQty" &&
      props.purchaseType == "purchase/request"
    ) {
      fields.readType = "hide";
    } else if (
      ["UnitCost", "DiscountValue", "DiscountAmount", "SubTotal"].includes(
        fields.field
      )
    ) {
      fields.width = "150px";
    }
    return fields;
  });
  cfg.fields.sort((a, b) => (a.idx > b.idx ? 1 : -1));
}

async function GetDataAssets(JournalID, JournalType) {
  try {
    const resp = await axios.post(`/asset/read-asset-content-by-journal`, {
      JournalType: JournalType,
      JournalID: JournalID,
    });

    return resp.data.map((resp) => ({
      OriginalFileName: resp.Asset.OriginalFileName,
      ContentType: resp.Asset.ContentType,
      FileName: resp.Asset.OriginalFileName,
      UploadDate: resp.Asset.Data.UploadDate,
      Description: resp.Asset.Data.Description,
      Content: resp.Content,
    }));
  } catch (error) {
    util.showError(error);
  }
}

function alterFormConfig(cfg) {
  const frm = JSON.parse(JSON.stringify(cfg));
  let groupCfg = [];
  for (let i = 0; i < data.formCfg.sectionGroups.length; i++) {
    if (i == 1) {
      frm.sectionGroups[0].sections[1].name = "footer";
      frm.sectionGroups[0].sections[1].title = "footer";
      let footer = {
        sections: [frm.sectionGroups[0].sections[1]],
      };
      groupCfg.push(footer);
    }
    groupCfg.push(data.formCfg.sectionGroups[i]);
  }
  groupCfg.map((sectionGroup) => {
    sectionGroup.sections.map((sections) => {
      if (["General", "footer"].includes(sections.name)) {
        sections.rows.map((row) => {
          if (["LineNo", "BatchSerials"].includes(row.inputs[0].field)) {
            row.inputs[0].hide = true;
          }
          row.inputs[0].readOnly = true;
        });
      }
    });
  });
  cfg.sectionGroups = groupCfg.slice(2, 4);
  console.log(cfg);
}
function getDataValue() {
  if (!listControl.value) {
    return [];
  }
  return listControl.value.getGridRecords();
}

function getOtherTotal() {
  return data.detailTotal;
}
function refreshData() {
  onGeneratePRLine();
}

function generateGridCfg(colum) {
  let addColm = [];
  for (let index = 0; index < colum.length; index++) {
    addColm.push({
      field: colum[index].field,
      kind: colum[index].kind,
      label: colum[index].label,
      readType: "show",
      labelField: "",
      width: colum[index].width,
      readOnly: colum[index].readOnly,
      input: {
        field: colum[index].field,
        label: colum[index].label,
        hint: "",
        hide: false,
        placeHolder: colum[index].label,
        kind: colum[index].kind,
        width: colum[index].width,
        readOnly: colum[index].readOnly,
      },
    });
  }
  return {
    fields: addColm,
    setting: {
      idField: "_id",
      keywordFields: ["_id", "Name"],
      sortable: ["_id"],
    },
  };
}

function createGridCfgExpenses() {
  const colum = [
    {
      field: "Expenses",
      kind: "text",
      label: "Expenses",
      width: "200px",
      readOnly: false,
    },
    {
      field: "Amount",
      kind: "number",
      label: "Amount",
      width: "100px",
      readOnly: false,
    },
  ];
  data.gridCfgExpenses = generateGridCfg(colum);
}
watch(
  () => subTotal.value,
  (nv) => {
    data.detailTotal.TotalAmount = nv;
  }
);
watch(
  () => discount.value,
  (nv) => {
    data.detailTotal.TotalDiscountAmount = Math.abs(nv) * -1;
  }
);
watch(
  () => grandtotal.value,
  (nv) => {
    data.detailTotal.GrandTotalAmount = nv;
  }
);
onMounted(() => {
  createGridCfgExpenses();
  loadGridConfig(axios, `/scm/purchase/line/gridconfig`).then((r) => {
    loadGridConfig(axios, `/scm/inventory/journal/line/gridconfig`).then(
      (rLine) => {
        const alterFields = [
          {
            field: "PRID",
            kind: "text",
            label: "Purchase Request No.",
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
              field: "PRID",
              label: "Purchase Request No.",
              hint: "",
              hide: false,
              placeHolder: "Line",
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
            field: "WarehouseID",
            kind: "text",
            label: "Delivery to",
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
              field: "WarehouseID",
              label: "Delivery to",
              hint: "",
              hide: false,
              placeHolder: "Delivery to",
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
              useList: true,
              allowAdd: false,
              items: [],
              useLookup: true,
              lookupUrl: "",
              lookupKey: "",
              lookupLabels: [],
              lookupSearchs: [],
              lookupFormat1: "",
              lookupFormat2: "",
              showTitle: false,
              showHint: false,
              showDetail: false,
              fixTitle: false,
              fixDetail: false,
              section: "Delivery to",
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
        ];
        // const filteredFields = r.fields.filter(o => !['InventJournalLine', 'LineNo'].includes(o.field))
        data.prlinegridCfg.fields = [
          ...rLine.fields,
          ...alterFields,
          ...r.fields,
        ].filter((o) => !["InventJournalLine", "LineNo"].includes(o.field));
        data.gridCfg = rLine;
      }
    );
  });
  loadFormConfig(axios, `/scm/inventory/journal/line/formconfig`).then(
    (rLine) => {
      data.formCfg = rLine;
    }
  );
  data.detailTotal = {
    TotalAmount: props.generalRecord.TotalAmount,
    TotalDiscountAmount: props.generalRecord.TotalDiscountAmount,
    OtherExpenses: props.generalRecord.OtherExpenses,
    Discount: {
      DiscountType: props.generalRecord.Discount.DiscountType,
      DiscountValue: props.generalRecord.Discount.DiscountValue,
      DiscountAmount: props.generalRecord.Discount.DiscountAmount,
    },
    GrandTotalAmount: props.generalRecord.GrandTotalAmount,
    tax: 0,
    freight: 0,
  };
});

defineExpose({
  getDataValue,
  getOtherTotal,
  gridRefreshed,
});
</script>
