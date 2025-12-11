<template>
  <div>
    <div v-if="!isFilled" class="w-full items-start gap-2 grid gridCol1">
      <div class="col-auto">
        <div class="suim_input">
          <div>
            <label class="input_label"><div>Attachment</div> </label>
            <div class="mt-2">
              <input
                type="file"
                ref="fileRef"
                class="input_field p-2"
                @change="(file) => FileReadertoBase64(file)"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
    <div v-else class="flex gap-5">
      <div class="relative box-file w-28 h-28 items-center align-middle bg-gray-100 border-2 border-dashed border-gray-300 rounded">
        <!-- v-if="!readOnly" -->
        <button
          class="absolute top-[-10px] right-[-10px] z-[10]"
          @click="deleteFile()"
        >
          <mdicon name="close-circle" size="20" class="w-[20px] text-gray-700"></mdicon>
        </button>
        <mdicon name="file-check" size="100" class="w-[20px] text-gray-600"></mdicon>
      </div>
      
      <span>{{ modelValue || props.file?.AssetData?.Asset?.Title}}</span>
    </div>
  </div>
</template>
<script setup>
import {
  util,
  SButton,
  loadFormConfig,
} from "suimjs";
import { reactive, defineProps, ref, onMounted, inject, computed } from "vue";
import lodash from 'lodash';
const axios = inject("axios");
const props = defineProps({
  file: { type: Object, default: () => {} },
  modelValue: { type: String, default: "" },
  readOnly: { type: Boolean, default: false },
});
const emit = defineEmits({
	"update:modelValue": null,
  postPickFile: null,
});
const data = reactive({
  formCfg: {
    sectionGroups: [
      {
        sections: [
          {
            title: "General",
            name: "General",
            showTitle: false,
            rows: [
              [
                {
                  field: "Attachment",
                  label: "Attachment",
                  hint: "",
                  hide: false,
                  placeHolder: "Attachment",
                  kind: "file",
                  disable: false,
                  required: false,
                  multiple: false,
                  multiRow: 1,
                  minLength: 0,
                  maxLength: 999,
                  readOnly: true,
                  readOnlyOnEdit: true,
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
                  row: 1001,
                  col: 1,
                  labelField: "",
                  decimal: 0,
                  dateFormat: "DD-MMM-YYYY hh:mm:ss Z",
                  unit: "",
                  width: "",
                  spaceBefore: 0,
                  spaceAfter: 0,
                },
              ],
            ],
            autoCol: 1,
            width: "",
          },
        ],
      },
    ],
    setting: {
      idField: "_id",
      title: "Attachment Form",
      showTitle: false,
      initialMode: "edit",
      hideButtons: false,
      hideEditButton: false,
      hideSubmitButton: false,
      hideCancelButton: false,
      submitText: "Save",
      autoCol: 1,
      sectionDirection: "col",
      sectionSize: 5,
    },
  },
  record: {
    Attachment: props.modelValue,
  },
  fileDetail: {},
});
function FileReadertoBase64(eventfile, config) {
  // const value = props.modelValue;
  const file = eventfile.target.files[0];
  if (file) {
    const reader = new FileReader();
    reader.readAsDataURL(file);
    reader.onload = function () {
      const binaryString = reader.result.substr(reader.result.indexOf(",") + 1);
      const _file = {
        AssetData: {
          Asset: {
            Title: file.name,
            OriginalFileName: file.name,
            NewFileName: file.name,
          },
          Content: binaryString,
        },
        Field: "AttachmentID",
      };
      emit("postPickFile", _file);
    };
  } else {
    props.file = {};
    emit("postPickFile", {});
  }
}
function deleteFile() {
  emit("update:modelValue", '');
  emit("postPickFile", {});
} 
const isFilled = computed ({
  get () {
    const v = props.modelValue !== '' || lodash.isEmpty(props.file) === false
    return v
  }
})
onMounted(() => {
  if (props.modelValue) {
    axios.post(`/asset/view?_id=${props.modelValue}`).then(r => {
    })
  }
});
</script>
