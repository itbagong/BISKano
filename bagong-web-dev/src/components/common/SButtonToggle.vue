<template>
  <div
    class="switch relative h-[22px] max-h-[22px] rounded-xl ring-inset ring-1 ring-gray-300"
    @click="change(!value)"
    :class="{ 'pointer-events-none': props.readOnly }"
  >
    <label for="no" class="switch-label switch-label-off" :class="colorText">
      <slot name="no_label">
        {{ props.noLabel }}
      </slot>
    </label>
    <label for="yes" class="switch-label switch-label-on" :class="colorText">
      <slot name="yes_label">
        {{ props.yesLabel }}
      </slot>
    </label>

    <span
      class="switch-selection"
      :class="[
        value == props.yesValue
          ? props.yesClassStyle + ' left-[50%]'
          : noClassStyle
          ? noClassStyle
          : 'bg-[#aaaaaa]',
        value == props.yesValue ? 'checked' : 'unchecked',
      ]"
    ></span>
  </div>
</template>
<script setup>
import { reactive, computed, onMounted } from "vue";
const props = defineProps({
  yesLabel: { type: String, default: "Yes" },
  noLabel: { type: String, default: "No" },
  yesValue: { type: Boolean, default: true },
  noValue: { type: Boolean, default: false },
  modelValue: { type: Boolean, default: false },
  name: { type: String, default: "radio" },
  colorLabel: { type: String, default: "white" },
  yesClassStyle: { type: String, default: "bg-secondary" },
  noClassStyle: { type: String, default: "" },
  readOnly: { type: Boolean, default: false },
});
const value = computed({
  get() {
    return props.modelValue;
  },
  set(v) {
    emit("update:modelValue", v);
  },
});
const emit = defineEmits({
  "update:modelValue": null,
  change: null,
});

const colorText = computed({
  get() {
    return "text-" + props.colorLabel;
  },
});
function change(val) {
  value.value = val;
  emit("change");
}
</script>
<style scope>
.switch-label {
  @apply relative float-left cursor-pointer leading-[1.4rem] text-[0.6rem] rounded-full  w-[50%] h-[22px] max-h-[22px]  flex justify-center items-center;
  z-index: 5 !important;
}
.switch-label:active {
  @apply font-bold;
}
.switch-input {
  @apply hidden;
}
.switch-input:checked + .switch-label {
  @apply font-bold shadow-sm transition-colors ease-out duration-150;
}

.switch-selection {
  @apply rounded-xl
        w-3/6
        block
        absolute  
        h-[22px]
         max-h-[22px]
        z-[1px];
  -webkit-transition: left 0.15s ease-out;
  -moz-transition: left 0.15s ease-out;
  -ms-transition: left 0.15s ease-out;
  -o-transition: left 0.15s ease-out;
  transition: left 0.15s ease-out;
}
</style>
