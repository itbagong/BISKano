<template>
  <s-grid
    class="grid_Inspection_Lines"
    ref="gridLine"
    :config="data.cfg"
    hide-search
    hide-sort
    hide-refresh-button
    hide-select
    hide-footer
    hide-new-button
    hide-action
    hide-control
    auto-commit-line
    no-confirm-delete
  >
    <template #item_IsApplicable="{ item }">
      <s-toggle
        v-model="item.IsApplicable"
        class="w-[100px]"
        @change="resetRow(item)"
      />
    </template>
    <template #item_ValueDescription="{ item }">
      <div class="text-pretty">
        {{ item.TemplateLine.Description }}
      </div>
    </template>
    <template #item_Value="{ item }">
      <s-input
        use-list
        v-model="item.Value"
        :items="formatItemsDDL(item.TemplateLine.AnswerValue)"
        @change="onChangeValue(item)"
        :disabled="item.IsApplicable"
      />
    </template>
    <template #item_Pica="{ item }">
      <div class="w-20 flex justify-center">
        <mdicon
          name="folder"
          size="16"
          @click="item.UsePica = true"
          v-if="item.Value !== item.MaxValue"
          :class="item.IsApplicable ? 'pointer-events-none' : ''"
        />
        <div v-else>-</div>
      </div>
      <s-modal
        v-if="item.UsePica"
        display
        hideButtons
        @beforeHide="item.UsePica = false"
      >
        <s-card hide-title class="min-w-[350px]">
          <pica v-model="item.Pica" />
          <s-button
            class="btn_secondary"
            @click="item.UsePica = false"
            label="Save"
          ></s-button>
        </s-card>
      </s-modal>
    </template>
    <template #item_Remark="{ item }">
      <s-input
        :multi-row="3"
        v-model="item.Remark"
        :disabled="item.IsApplicable"
      />
    </template>
    <!-- csms -->
    <template #item_Metode="{ item }">
      <div v-if="item.IsApplicable">-</div>
      <s-input
        v-model="item.Metode"
        use-list
        :items="['D', 'L']"
        multiple
        v-else
      />
    </template>
    <template #item_Result="{ item }">
      <s-input
        use-list
        v-model="item.Result"
        :disabled="item.IsApplicable"
        :items="formatItemsDDL(item.Bobot)"
        @change="onChangeResult(item)"
        v-if="kind == 'csms'"
      />
      <s-toggle
        v-model="item.Result"
        class="w-[100px]"
        v-if="kind == 'observasi'"
        :class="item.IsApplicable ? 'pointer-events-none' : ''"
      />
    </template>
    <!-- csms -->
    <template #item_Attachment="{ item }">
      <uploader
        ref="gridAttachment"
        :journalId="`SHE_${kind}_LINES_${item.Attachment}`"
        :config="{}"
        :journalType="`SHE_${kind}`"
        single-save
        :class="item.IsApplicable ? 'pointer-events-none' : ''"
      />
    </template>
    <template #item_Validation="{ item }">
      <div class="flex justify-center">
        <mdicon
          v-if="kind == 'csms' && item.IsApplicable"
          name="checkbox-marked"
          size="16"
          class="text-blue-400"
        />
        <s-input
          kind="checkbox"
          v-model="item.Validation"
          v-else-if="kind == 'csms' && !item.IsApplicable"
        />
      </div>
    </template>
    <template #item_MaxValue="{ item }">
      <div class="text-center">
        {{ item.TemplateLine.AnswerValue }}
      </div>
    </template>
    <template #item_Bobot="{ item }">
      <div class="text-center">
        {{ item.TemplateLine.AnswerValue }}
      </div>
    </template>
    <template #item_ACH="{ item }">
      <div class="text-right">
        {{ item.ACH }}
      </div>
    </template>
    <template #item_ExpDate="{ item }">
      <div class="text-right">
        <s-input kind="date" v-model="item.ExpDate" v-if="kind == 'csms'" />
      </div>
    </template>
    <!-- observasi -->
    <template #item_Deviation="{ item }">
      <div class="text-center">
        <s-input
          :multi-row="3"
          v-model="item.Deviation"
          :disabled="item.IsApplicable"
        />
      </div>
    </template>
    <template #item_HazardCode="{ item }">
      <div v-if="item.IsApplicable">-</div>
      <s-input
        v-model="item.HazardCode"
        use-list
        :items="['AA', 'A', 'B', 'C']"
        v-else
      />
    </template>
    <template #item_HarzardLevel="{ item }">
      {{ item.TemplateLine.Description }}
    </template>
    <!-- observasi -->
  </s-grid>
</template>
<script setup>
import { reactive, ref, watch, inject, onMounted, computed } from "vue";
import { util, SInput, SButton, SGrid, SModal, SCard } from "suimjs";
import SToggle from "@/components/common/SButtonToggle.vue";
import helper from "@/scripts/helper.js";
import Uploader from "@/components/common/Uploader.vue";
import Pica from "@/components/common/ItemPica.vue";

const axios = inject("axios");

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  templateLines: { type: Array, default: () => [] },
  kind: { type: String, default: "" },
  lineCfg: { type: Array, default: () => [] },
});

const emit = defineEmits({
  "update:modelValue": null,
});

const gridLine = ref(null);

const data = reactive({
  cfg: {},
  record: props.modelValue ?? [],
});

function getDataByTemplate(id) {
  if (!id) {
    data.record = [];
    refreshGridLines();
    return;
  }

  const url = "/she/mcuitemtemplate/get";
  axios.post(url, [id]).then(
    (r) => {
      mapFieldLine(r.data.Lines);
    },
    (e) => {
      util.showError(e);
    }
  );
}

function mapFieldLine(dt) {
  let tmp = [];
  for (let i in dt) {
    let o = dt[i];
    let x = {};
    if (props.kind == "inspection") {
      x["Pica"] = { Status: "Open" };
      x["MaxValue"] = o.AnswerValue;
      x["Value"] = 0;
    }

    if (props.kind == "csms") {
      x["Metode"] = [];
      x["Bobot"] = o.AnswerValue;
      x["Result"] = 0;
      x["ACH"] = 0;
      x["ExpDate"] = new Date();
    }

    if (props.kind == "observasi") {
      x["Result"] = false;
    }
    x["ValueDescription"] = o.Description;
    x["IsApplicable"] = false;
    x["Remark"] = "";
    x["Attachment"] = util.uuid();
    x["TemplateLine"] = o;
    tmp.push(x);
  }
  data.record = tmp;
  refreshGridLines();
}

function formatColSpan(selector, numSpan) {
  if (data.record.length == 0) return;
  const myTable = document.querySelectorAll(
    `.${selector} .suim_table [name='grid_body']`
  );

  for (let i = 0, row; (row = myTable[0].rows[i]); i++) {
    if (!data.record[i].TemplateLine.Parent) {
      row.innerHTML = `<td colspan="${numSpan}" class='p-2 font-semibold'>${data.record[i].TemplateLine.Description}</td>`;
    }
  }
  generateThead();
}

function generateThead() {
  let sumBobot = data.record.reduce(function (acc, obj) {
    return acc + (typeof obj.Bobot == "number" ? obj.Bobot : 0);
  }, 0);
  let sumResult = data.record.reduce(function (acc, obj) {
    return acc + (typeof obj.Result == "number" ? obj.Result : 0);
  }, 0);
  let sumACH = data.record.reduce(function (acc, obj) {
    return acc + (typeof obj.ACH == "number" ? obj.ACH : 0);
  }, 0);
  const dtThead = [
    { name: "Not Applicable", rowSpan: 2 },
    { name: "Work Description", rowSpan: 2 },
    { name: "Metode", rowSpan: 2, width: "130px" },
    {
      name: "Bobot",
      rowSpan: 0,
      colSpan: 1,
      child: [sumBobot],
      width: "100px",
    },
    {
      name: "Result",
      rowSpan: 0,
      colSpan: 1,
      child: [sumResult],
      width: "100px",
    },
    {
      name: "ACH (%)",
      rowSpan: 0,
      colSpan: 1,
      child: [sumACH],
      width: "100px",
    },
    { name: "Attachment", rowSpan: 2, width: "100px" },
    { name: "Validation", rowSpan: 2, width: "100px" },
    { name: "Exp Date", rowSpan: 2 },
    { name: "Remark", rowSpan: 2, width: "200px" },
  ];
  const el = document.querySelector(".grid_Inspection_Lines table thead");
  if (el.querySelectorAll("tr").length == 1) {
    let row = el.insertRow(0);
    let row2 = el.insertRow(1);
    for (let i in dtThead) {
      let c = dtThead[i];
      row.innerHTML += `<td ${c.rowSpan > 0 ? "rowspan=" + c.rowSpan : ""}
      ${
        c.colSpan > 0 ? "colspan=" + c.colSpan : ""
      } class='text-center border-r font-semibold' width='${
        c.width ? c.width : ""
      }'>${c.name}</td>`;
      if (c.child) {
        for (let j in c.child) {
          row2.innerHTML += `<td class='text-center border-r font-semibold'>${c.child[j]}</td>`;
        }
      }
    }
    el.lastElementChild.remove();
  } else {
    let td = el.querySelectorAll("tr:nth-child(2) td");
    el.querySelector(`tr:nth-child(2) td:nth-child(${1})`).textContent =
      sumBobot;
    el.querySelector(`tr:nth-child(2) td:nth-child(${2})`).textContent =
      sumResult;
    el.querySelector(`tr:nth-child(2) td:nth-child(${3})`).textContent = sumACH;
  }
}

function refreshGridLines() {
  util.nextTickN(4, () => {
    gridLine.value.setRecords(data.record);
    if (data.record.length == 0) return;
    setTimeout(() => {
      formatColSpan("grid_Inspection_Lines", 7);
    }, 500);
  });
}

function onChangeValue(item) {
  util.nextTickN(2, () => {
    if (item.TemplateLine.AnswerValue !== parseInt(item.Value)) {
      item.Pica = { Status: "Open" };
    } else {
      item.Pica = { Status: "" };
    }
  });
}

function onChangeResult(item) {
  util.nextTickN(2, () => {
    let sum = data.record.reduce(function (acc, obj) {
      return acc + (typeof obj.Bobot == "number" ? obj.Bobot : 0);
    }, 0);
    item.ACH = Math.round((item.Result / sum) * 100 * 100) / 100;
    refreshGridLines();
  });
}
function resetRow(item) {
  util.nextTickN(2, () => {
    if (props.kind == "csms") {
      item.Metode = [];
      item.Result = 0;
    }
    if (props.kind == "inspection") {
      item.Pica = { Status: "Open" };
      item.Value = 0;
    }
    if (props.kind == "observasi") {
      item.Result = false;
    }
    item.Remark = "";
  });
}

function formatItemsDDL(max) {
  let res = [];
  for (let i = 0; i <= max; i++) {
    let o = {};
    o["key"] = i;
    o["text"] = i.toString();
    res.push(o);
  }
  return res;
}

watch(
  () => props.templateLines,
  (nv) => {
    getDataByTemplate(nv);
  }
);

watch(
  () => data.record,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);

function formatCfg() {
  let cfg = [];
  for (let i in props.lineCfg) {
    let o = props.lineCfg[i];
    cfg.push(helper.gridColumnConfig(o));
  }
  data.cfg.fields = cfg;
  refreshGridLines();
}

onMounted(() => {
  formatCfg();
});
</script>
