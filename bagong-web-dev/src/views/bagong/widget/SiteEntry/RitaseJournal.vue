<template>
  <table class="w-full border border-[#D1D5DC] bg-[#f7f8f9]">
    <tr class="[&>*]:px-2 [&>*]:border-l-[1px]">
      <td>
        <div class="flex gap-2 items-center grow">
          <div>Line No:</div>
          <div>{{ data.item.LineNo }}</div>
        </div>
      </td>

      <td>
        <div class="flex gap-2 items-center">
          <div>Approval Status:</div>

          <status-text :txt="data.item.ApprovalStatus" />
        </div>
      </td>

      <td>
        <div class="flex gap-2 items-center">
          <div>Journal No:</div>

          <!-- <div>{{ data.item.JournalID }}</div> -->
          <div class="bg-transparent" v-if="hasJournalID(data.item.JournalID)">
            <a
              href="#"
              class="text-blue-400 hover:text-blue-800"
              @click="redirect(data.item.JournalID)"
              >{{ data.item.JournalID }}
            </a>
          </div>
        </div>
      </td>

      <td>
        <div class="flex gap-2 items-center">
          <div>Voucher No:</div>

          <div>{{ data.item.VoucherID }}</div>
        </div>
      </td>

      <td>
        <div class="flex justify-end">
          <s-button
            v-if="
              data.item.ApprovalStatus == 'REJECTED' &&
              hasJournalID(data.item.JournalID)
            "
            class="btn_reopen submit_btn text-xs"
            @click="reOpen(item)"
            label="RE-OPEN"
          />
        </div>
      </td>
    </tr>
  </table>
</template>
<script setup>
import { reactive, watch } from "vue";
import { SButton } from "suimjs";
import StatusText from "@/components/common/StatusText.vue";
import { useRouter } from 'vue-router';

const router = useRouter()
const props = defineProps({
  modelValue: { type: Object, default: () => {} },
  journalType: { type: String, default: "" },
});

const emit = defineEmits({
  "update:modelValue": null,
  reOpen: null,
});
const data = reactive({
  item:
    props.modelValue == null || props.modelValue == undefined
      ? { ApprovalStatus: "", VoucherID: "", JournalID: "" }
      : props.modelValue,
});
function hasJournalID(journalID) {
  return journalID != "" && journalID != undefined;
}
function reOpen() {
  data.item.JournalID = "";
  data.item.ApprovalStatus = "";
  emit("reOpen");
}
function redirect(JournalID) {
  if (props.journalType === 'VENDOR') {
    const url = router.resolve({
      name: "fico-VendorTransaction",
      query: { id: JournalID },
    });
    window.open(url.href, "_blank");
  } else if (props.journalType === 'CUSTOMER') {
    const url = router.resolve({
      name: "fico-CustomerTransaction",
      query: { id: JournalID },
    });
    window.open(url.href, "_blank");
  }
}
watch(
  () => data.item,
  (nv) => {
    emit("update:modelValue", nv);
  },
  { deep: true }
);
</script>
