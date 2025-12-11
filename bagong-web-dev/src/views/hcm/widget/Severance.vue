<template>
  <div>
    <div class="title section_title">Hak Pekerja</div>
    <div>
      <s-input
        class="w-[500px] mb-5"
        label="Item Template ID"
        use-list
        lookup-url="/she/mcuitemtemplate/find"
        lookup-key="_id"
        :lookup-labels="['Name']"
        :lookupSearchs="['Name']"
        v-model="data.record.NonTaxableIncomeTemplateID"
        @change="(_, v) => renderNonTaxableIncome(v)"
      />
      <div class="suim_area_table">
        <table class="w-full table-auto suim_table">
          <thead class="grid_header">
            <tr class="border-b-[1px] border-slate-500">
              <th>No.</th>
              <th>Komponen Penghasilan Kena Pajak Tidak Final</th>
              <th>Perhitungan</th>
              <th>Nomor PP</th>
              <th>Jumlah (Rp)</th>
            </tr>
          </thead>
          <tbody v-if="data.loadingNon">
            <tr>
              <td colspan="7">
                <div class="h-[300px] flex items-center justify-center">
                  <Loader kind="circle" />
                </div>
              </td>
            </tr>
          </tbody>
          <tbody v-else>
            <tr
              v-for="(item, key) in data.record.NonTaxableIncome"
              :key="`row-${key}`"
            >
              <td class="border border-gray-300 px-4 py-1">
                {{ item.Number }}
              </td>
              <td class="border border-gray-300 px-4 py-1">
                {{ item.Name }}
              </td>
              <td class="border border-gray-300 px-4 py-1">
                <s-input
                  v-model.number="item.Calculation"
                  :read-only="readOnly"
                  kind="number"
                  class="px-2"
                />
              </td>
              <td class="border border-gray-300 px-4 py-1">
                <s-input
                  v-model="item.PPNo"
                  :read-only="readOnly"
                  kind="text"
                  class="px-2"
                />
              </td>
              <td class="border border-gray-300 px-4 py-1">
                <s-input
                  v-model.number="item.Amount"
                  :read-only="key == 0 || readOnly"
                  kind="number"
                  class="px-2"
                />
              </td>
            </tr>
          </tbody>
          <tfoot>
            <tr class="bg-gray-100">
              <td
                colspan="2"
                class="border border-gray-300 px-4 py-1 font-bold"
              >
                Jumlah Netto
              </td>
              <td
                colspan="3"
                class="border border-gray-300 px-4 py-1 font-bold text-right"
              >
                {{ helper.formatNumberWithDot(calculateNettoAmount) }}
              </td>
            </tr>
          </tfoot>
        </table>
      </div>
    </div>
    <div class="mb-5">
      <s-input
        class="w-[500px] mb-5"
        label="Item Template ID"
        use-list
        lookup-url="/she/mcuitemtemplate/find"
        lookup-key="_id"
        :lookup-labels="['Name']"
        :lookupSearchs="['Name']"
        v-model="data.record.TaxableIncomeTemplateID"
        @change="(_, v) => renderTaxableIncome(v)"
      />
      <div class="suim_area_table">
        <table class="w-full table-auto suim_table">
          <thead class="grid_header">
            <tr class="border-b-[1px] border-slate-500">
              <th>No.</th>
              <th>Komponen Penghasilan Kena Pajak Final</th>
              <th>Perhitungan</th>
              <th>Jumlah (Rp)</th>
            </tr>
          </thead>
          <tbody v-if="data.loadingNon">
            <tr>
              <td colspan="7">
                <div class="h-[300px] flex items-center justify-center">
                  <Loader kind="circle" />
                </div>
              </td>
            </tr>
          </tbody>
          <tbody v-else>
            <tr
              v-for="(item, key) in data.record.TaxableIncome"
              :key="`row-${key}`"
            >
              <td class="border border-gray-300 px-4 py-1">
                {{ item.Number }}
              </td>
              <td class="border border-gray-300 px-4 py-1">
                {{ item.Name }}
              </td>
              <td class="border border-gray-300 px-4 py-1">
                <s-input
                  v-model="item.Calculation"
                  :read-only="[0, 1].includes(key) || readOnly"
                  class="px-2"
                  @change="(_, v) => (item.Amount = parseFloat(v))"
                />
              </td>
              <td class="border border-gray-300 px-4 py-1">
                <s-input
                  v-model.number="item.Amount"
                  :read-only="[0, 1].includes(key) || readOnly"
                  kind="number"
                  class="px-2"
                />
              </td>
            </tr>
          </tbody>
          <tfoot>
            <tr class="bg-gray-100">
              <td
                colspan="2"
                class="border border-gray-300 px-4 py-1 font-bold"
              >
                Jumlah yang diterima pekerja
              </td>
              <td
                colspan="2"
                class="border border-gray-300 px-4 py-1 font-bold text-right"
              >
                {{
                  helper.formatNumberWithDot(calculateTaxableAmountWorkerAccept)
                }}
              </td>
            </tr>
          </tfoot>
        </table>
      </div>
    </div>
    <div class="title section_title">Kewajiban Pekerja</div>
    <div>
      <s-input
        class="w-[500px] mb-5"
        label="Item Template ID"
        use-list
        lookup-url="/she/mcuitemtemplate/find"
        lookup-key="_id"
        :lookup-labels="['Name']"
        :lookupSearchs="['Name']"
        v-model="data.record.MandatoryWorkTemplateID"
        @change="(_, v) => renderMandatoryWork(v)"
      />
      <div class="suim_area_table">
        <table class="w-full table-auto suim_table">
          <thead class="grid_header">
            <tr class="border-b-[1px] border-slate-500">
              <th>No.</th>
              <th>Komponen</th>
              <th>Perhitungan</th>
              <th>Nomor RV</th>
              <th>Jumlah (Rp)</th>
            </tr>
          </thead>
          <tbody v-if="data.loadingNon">
            <tr>
              <td colspan="7">
                <div class="h-[300px] flex items-center justify-center">
                  <Loader kind="circle" />
                </div>
              </td>
            </tr>
          </tbody>
          <tbody v-else>
            <tr
              v-for="(item, key) in data.record.MandatoryWork"
              :key="`row-${key}`"
            >
              <td class="border border-gray-300 px-4 py-1">
                {{ item.Number }}
              </td>
              <td class="border border-gray-300 px-4 py-1">
                {{ item.Name }}
              </td>
              <td class="border border-gray-300 px-4 py-1">
                <s-input
                  v-model.number="item.Calculation"
                  kind="number"
                  :read-only="readOnly"
                  class="px-2"
                  @change="(_, v) => (item.Amount = parseFloat(v))"
                />
              </td>
              <td class="border border-gray-300 px-4 py-1">
                <s-input
                  v-model="item.RVNo"
                  :read-only="readOnly"
                  kind="text"
                  class="px-2"
                />
              </td>
              <td class="border border-gray-300 px-4 py-1">
                <s-input
                  v-model.number="item.Amount"
                  :read-only="readOnly"
                  kind="number"
                  class="px-2"
                />
              </td>
            </tr>
          </tbody>
          <tfoot>
            <tr class="bg-gray-100">
              <td
                colspan="2"
                class="border border-gray-300 px-4 py-1 font-bold"
              >
                Jumlah Netto
              </td>
              <td
                colspan="3"
                class="border border-gray-300 px-4 py-1 font-bold text-right"
              >
                {{
                  helper.formatNumberWithDot(calculateNettoAmountMandatoryWork)
                }}
              </td>
            </tr>
            <tr class="bg-gray-100">
              <td
                colspan="2"
                class="border border-gray-300 px-4 py-1 font-bold"
              >
                Jumlah Yang Harus Diterima Pekerja
              </td>
              <td class="border border-gray-300 px-4 py-1 font-bold text-right">
                {{ helper.formatNumberWithDot(calculateAmountWorkerAccept) }}
              </td>
              <td
                colspan="2"
                class="border border-gray-300 px-4 py-1 font-bold text-right"
              >
                {{ data.record.AmountWorkerAcceptWord }}
              </td>
            </tr>
          </tfoot>
        </table>
      </div>
    </div>
    <div class="title section_title">Sebab Terjadinya PHK</div>
    <div class="content">
      <div class="relative w-full my-4 rounded-md border h-10 p-1 bg-gray-200">
        <div class="relative w-full h-full flex items-center">
          <div
            @click="handleSwitchToggle"
            class="w-full flex justify-center text-gray-400 cursor-pointer"
          >
            <button>Keluar Atas Inisiatif Perusahaan</button>
          </div>
          <div
            @click="handleSwitchToggle"
            class="w-full flex justify-center text-gray-400 cursor-pointer"
          >
            <button>Keluar Atas Inisiatif Pekerja</button>
          </div>
        </div>
        <span
          :class="
            data.isCompanyInisitive
              ? 'left-1 text-primary font-semibold'
              : 'left-1/2 -ml-1 text-primary font-semibold'
          "
          class="bg-white shadow text-sm flex items-center justify-center w-1/2 rounded h-[1.88rem] transition-all duration-150 ease-linear top-[4px] absolute"
        >
          {{
            data.isCompanyInisitive
              ? "Keluar Atas Inisiatif Perusahaan"
              : "Keluar Atas Inisiatif Pekerja"
          }}
        </span>
      </div>
      <div v-if="data.isCompanyInisitive" class="grid grid-cols-2 gap-2">
        <div>
          <div v-for="el in data.itemsCompanyReason" :key="el" class="mb-2">
            <label class="flex items-center gap-2 cursor-pointer">
              <input
                type="radio"
                :disabled="readOnly || mode == 'view'"
                :value="el"
                v-model="data.record.ResignOnCompanyInitiative"
                class="form-radio text-blue-600"
                @change="onChangeCompanyReason(el)"
              />
              <span>{{ el }}</span>
            </label>
          </div>
        </div>
        <div
          v-if="data.record.ResignOnCompanyInitiative == 'Lain-lain (sebutkan)'"
        >
          <s-input
            label="Lain-lain"
            multi-row="5"
            v-model="data.record.ResignOnCompanyInitiativeEtc"
            :read-only="readOnly"
          />
        </div>
      </div>
      <div v-else class="grid grid-cols-2 gap-2">
        <div>
          <div v-for="el in data.itemsEmployeeReason" :key="el" class="mb-2">
            <label class="flex items-center gap-2 cursor-pointer">
              <input
                type="radio"
                :disabled="readOnly || mode == 'view'"
                :value="el"
                v-model="data.record.ResignOnEmployeeInitiative"
                class="form-radio text-blue-600"
                @change="onChangeEmployeeReason(el)"
              />
              <span>{{ el }}</span>
            </label>
          </div>
        </div>
        <div
          v-if="
            data.record.ResignOnEmployeeInitiative == 'Lain-lain (sebutkan)'
          "
        >
          <s-input
            label="Lain-lain"
            multi-row="5"
            v-model="data.record.ResignOnEmployeeInitiativeEtc"
            :read-only="readOnly"
          />
        </div>
      </div>
    </div>
  </div>
</template>
<script setup>
import { reactive, inject, watch, onMounted, computed } from "vue";
import { SInput, SButton } from "suimjs";
import helper from "../../../scripts/helper";

const props = defineProps({
  modelValue: { type: Object, default: {} },
  readOnly: { type: Boolean, default: false },
});
const axios = inject("axios");
const emit = defineEmits({
  "update:modelValue": null,
});

const data = reactive({
  loadingNon: false,
  loadingTax: false,
  loadingMandatory: false,
  record: props.modelValue,
  isCompanyInisitive: props.modelValue.ResignOnCompanyInitiative !== '' ? true : false,
  itemsCompanyReason: [
    "Pelanggaran",
    "Rasionalisasi",
    "Medical Unfit",
    "Meninggal Dunia",
    "Pensiun",
    "Tidak Memenuhi Standar Prestasi",
    "Lain-lain (sebutkan)",
  ],
  itemsEmployeeReason: [
    "Salary",
    "Keluarga",
    "Melanjutkan Studi",
    "Jenjang Karir",
    "Kondisi Sosial",
    "Mengundurkan Diri",
    "Lain-lain (sebutkan)",
  ],
});

function renderNonTaxableIncome(v) {
  data.loadingNon = true;
  axios
    .post("/she/mcuitemtemplate/get", [v])
    .then((r) => {
      data.record.NonTaxableIncome = r.data.Lines.map((item) => {
        return {
          Number: item.Number,
          Name: item.Description,
          Calculation: 0.0,
          PPNo: "",
          Amount: 0,
        };
      });
    })
    .catch((e) => {
      util.showError(e);
    })
    .finally(() => {
      data.loadingNon = false;
    });
}
const calculateNettoAmount = computed(() => {
  const result = data.record?.NonTaxableIncome?.reduce(
    (total, item) =>
      total + (!isNaN(parseFloat(item.Amount)) ? parseFloat(item.Amount) : 0),
    0
  );
  data.record.NettoAmount = parseFloat(result) ?? 0;
  return data.record.NettoAmount;
});
function renderTaxableIncome(v) {
  data.loadingTax = true;
  axios
    .post("/she/mcuitemtemplate/get", [v])
    .then((r) => {
      data.record.TaxableIncome = r.data.Lines.map((item) => {
        return {
          Number: item.Number,
          Name: item.Description,
          Calculation: null,
          Amount: 0,
        };
      });
    })
    .catch((e) => {
      util.showError(e);
    })
    .finally(() => {
      data.loadingTax = false;
    });
}
const calculateTaxableAmountWorkerAccept = computed(() => {
  const result = data.record?.TaxableIncome?.reduce(
    (total, item) =>
      total + (!isNaN(parseFloat(item.Amount)) ? parseFloat(item.Amount) : 0),
    0
  );
  data.record.TaxableAmountWorkerAccept = parseFloat(result) ?? 0;
  return data.record.TaxableAmountWorkerAccept;
});
function renderMandatoryWork(v) {
  data.loadingMandatory = true;
  axios
    .post("/she/mcuitemtemplate/get", [v])
    .then((r) => {
      data.record.MandatoryWork = r.data.Lines.map((item) => {
        return {
          Number: item.Number,
          Name: item.Description,
          Calculation: 0,
          RVNo: "",
          Amount: 0,
        };
      });
    })
    .catch((e) => {
      util.showError(e);
    })
    .finally(() => {
      data.loadingMandatory = false;
    });
}
const calculateNettoAmountMandatoryWork = computed(() => {
  const result = data.record?.MandatoryWork?.reduce(
    (total, item) =>
      total + (!isNaN(parseFloat(item.Amount)) ? parseFloat(item.Amount) : 0),
    0
  );
  data.record.NettoAmountMandatoryWork = parseFloat(result) ?? 0;
  return data.record.NettoAmountMandatoryWork;
});

const calculateAmountWorkerAccept = computed(() => {
  data.record.AmountWorkerAccept =
    data.record.NettoAmount +
    data.record.TaxableAmountWorkerAccept -
    data.record.NettoAmountMandatoryWork;
  data.record.AmountWorkerAcceptWord = helper.convertToWordsIDR(
    data.record.AmountWorkerAccept
  );
  return data.record.AmountWorkerAccept.toFixed(2);
});
watch(
  () => data.record,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);

function handleSwitchToggle() {
  data.isCompanyInisitive = !data.isCompanyInisitive;
  data.record.ResignOnCompanyInitiative = "";
  data.record.ResignOnEmployeeInitiative = "";
  data.record.ResignOnCompanyInitiativeEtc = "";
  data.record.ResignOnEmployeeInitiativeEtc = "";
}
function onChangeCompanyReason(v) {
  if (v !== "Lain-lain (sebutkan)") {
    data.record.ResignOnCompanyInitiativeEtc = "";
  }
}
function onChangeEmployeeReason(v) {
  if (v !== "Lain-lain (sebutkan)") {
    data.record.ResignOnCompanyInitiativeEtc = "";
  }
}
onMounted(() => {
  // data.record = props.modelValue
});
</script>
<style></style>
